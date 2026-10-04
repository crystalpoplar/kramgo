package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOllamaBaseURL = "http://127.0.0.1:11434"
	defaultListenAddr    = "0.0.0.0:8080"
)

// Service exposes a local network HTTP API that proxies requests to a local Ollama instance.
type Service struct {
	OllamaBaseURL string
	ListenAddr    string
	Client        *http.Client
}

func NewService(ollamaBaseURL, listenAddr string) *Service {
	if ollamaBaseURL == "" {
		ollamaBaseURL = defaultOllamaBaseURL
	}
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}
	return &Service{
		OllamaBaseURL: ollamaBaseURL,
		ListenAddr:    listenAddr,
		Client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/generate", s.handleGenerate)
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/", s.handleRoot)
	return mux
}

func (s *Service) ListenAndServe() error {
	server := &http.Server{
		Addr:         s.ListenAddr,
		Handler:      s.Handler(),
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  10 * time.Minute,
	}
	return server.ListenAndServe()
}

func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Service) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"service":"kramgo","ollama":"` + s.OllamaBaseURL + `"}`))
}

func (s *Service) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := s.proxyOllamaRequest(w, r, "/api/generate"); err != nil {
		log.Printf("proxy error endpoint=%s remote=%s err=%v", "/api/generate", r.RemoteAddr, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
}

func (s *Service) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := s.proxyOllamaRequest(w, r, "/api/chat"); err != nil {
		log.Printf("proxy error endpoint=%s remote=%s err=%v", "/api/chat", r.RemoteAddr, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
}

func (s *Service) proxyOllamaRequest(w http.ResponseWriter, r *http.Request, path string) error {
	started := time.Now()
	requestID := fmt.Sprintf("%d", started.UnixNano())
	log.Printf("proxy start id=%s endpoint=%s remote=%s", requestID, path, r.RemoteAddr)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	defer r.Body.Close()

	body, streamRequested, err := normalizeOllamaRequestBody(body)
	if err != nil {
		return fmt.Errorf("normalize request body: %w", err)
	}
	log.Printf("proxy body id=%s endpoint=%s bytes=%d", requestID, path, len(body))

	url := strings.TrimRight(s.OllamaBaseURL, "/") + path
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build ollama request: %w", err)
	}
	req.Header = r.Header.Clone()
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		log.Printf("proxy upstream fail id=%s endpoint=%s elapsed=%s err=%v", requestID, path, time.Since(started), err)
		return fmt.Errorf("call ollama: %w", err)
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(resp.StatusCode)

	if streamRequested {
		bytesWritten, err := streamResponseBody(w, resp.Body)
		if err != nil {
			log.Printf("proxy stream fail id=%s endpoint=%s status=%d elapsed=%s err=%v", requestID, path, resp.StatusCode, time.Since(started), err)
			return fmt.Errorf("stream response: %w", err)
		}
		log.Printf("proxy success id=%s endpoint=%s status=%d response_bytes=%d streamed=true elapsed=%s", requestID, path, resp.StatusCode, bytesWritten, time.Since(started))
		return nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read ollama response: %w", err)
	}

	_, err = w.Write(respBody)
	if err != nil {
		log.Printf("proxy write fail id=%s endpoint=%s status=%d elapsed=%s err=%v", requestID, path, resp.StatusCode, time.Since(started), err)
		return fmt.Errorf("write response: %w", err)
	}

	log.Printf("proxy success id=%s endpoint=%s status=%d response_bytes=%d elapsed=%s", requestID, path, resp.StatusCode, len(respBody), time.Since(started))
	return nil
}

func normalizeOllamaRequestBody(body []byte) ([]byte, bool, error) {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return body, false, nil
	}

	streamRequested := false
	if streamVal, exists := payload["stream"]; exists {
		if streamBool, ok := streamVal.(bool); ok {
			streamRequested = streamBool
		}
	} else {
		payload["stream"] = false
	}

	updated, err := json.Marshal(payload)
	if err != nil {
		return nil, false, err
	}
	return updated, streamRequested, nil
}

func streamResponseBody(w http.ResponseWriter, body io.Reader) (int, error) {
	flusher, _ := w.(http.Flusher)
	buffer := make([]byte, 32*1024)
	total := 0

	for {
		readBytes, readErr := body.Read(buffer)
		if readBytes > 0 {
			written, writeErr := w.Write(buffer[:readBytes])
			total += written
			if writeErr != nil {
				return total, writeErr
			}
			if flusher != nil {
				flusher.Flush()
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				return total, nil
			}
			return total, readErr
		}
	}
}

func (s *Service) proxyJSONRequest(w http.ResponseWriter, r *http.Request, path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := s.Client.Post(strings.TrimRight(s.OllamaBaseURL, "/")+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("call ollama: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read ollama response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, err = w.Write(respBody)
	if err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}
