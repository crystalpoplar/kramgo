package app

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServiceUsesDefaults(t *testing.T) {
	service := NewService("", "")
	if service == nil {
		t.Fatal("NewService returned nil")
	}
	if service.OllamaBaseURL != "http://127.0.0.1:11434" {
		t.Fatalf("unexpected OllamaBaseURL: %q", service.OllamaBaseURL)
	}
	if service.ListenAddr != "0.0.0.0:8080" {
		t.Fatalf("unexpected ListenAddr: %q", service.ListenAddr)
	}
}

func TestGenerateEndpointForwardsToOllama(t *testing.T) {
	ollama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Fatalf("unexpected path: %q", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}
		payload := string(body)
		if !strings.Contains(payload, "\"prompt\":\"hello from the network\"") {
			t.Fatalf("request body missing prompt: %q", payload)
		}
		if !strings.Contains(payload, "\"stream\":false") {
			t.Fatalf("request body should disable streaming: %q", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"llama3.2","response":"hello back","done":true}`))
	}))
	defer ollama.Close()

	service := NewService(ollama.URL, "127.0.0.1:0")
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader(`{"model":"llama3.2","prompt":"hello from the network"}`))
	res := httptest.NewRecorder()

	service.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "hello back") {
		t.Fatalf("unexpected body: %s", res.Body.String())
	}
}

func TestChatEndpointPreservesStreamTrue(t *testing.T) {
	ollama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path: %q", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("unmarshal request failed: %v", err)
		}

		stream, ok := payload["stream"].(bool)
		if !ok {
			t.Fatalf("stream should be a bool: %v", payload["stream"])
		}
		if !stream {
			t.Fatalf("stream=true should be preserved, payload=%s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"llama3.2","message":{"role":"assistant","content":"hello"},"done":true}`))
	}))
	defer ollama.Close()

	service := NewService(ollama.URL, "127.0.0.1:0")
	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"model":"llama3.2","messages":[{"role":"user","content":"hello"}],"stream":true}`))
	res := httptest.NewRecorder()

	service.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "hello") {
		t.Fatalf("unexpected body: %s", res.Body.String())
	}
}
