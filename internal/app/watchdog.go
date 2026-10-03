package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

const defaultWatchdogPrompt = "Reply with OK only."

type WatchdogConfig struct {
	OllamaBaseURL string
	Model         string
	ServiceName   string
	Interval      time.Duration
	MaxFailures   int
	HealthPrompt  string
}

func DefaultWatchdogConfig() WatchdogConfig {
	return WatchdogConfig{
		OllamaBaseURL: "http://127.0.0.1:11434",
		Model:         "crystal:latest",
		ServiceName:   "ollama",
		Interval:      30 * time.Second,
		MaxFailures:   3,
		HealthPrompt:  defaultWatchdogPrompt,
	}
}

func (c WatchdogConfig) Validate() error {
	if strings.TrimSpace(c.OllamaBaseURL) == "" {
		return fmt.Errorf("ollama base url is required")
	}
	if strings.TrimSpace(c.Model) == "" {
		return fmt.Errorf("model is required")
	}
	if strings.TrimSpace(c.ServiceName) == "" {
		return fmt.Errorf("service name is required")
	}
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	if c.MaxFailures <= 0 {
		return fmt.Errorf("max failures must be positive")
	}
	return nil
}

func (c WatchdogConfig) Run(ctx context.Context) error {
	if err := c.Validate(); err != nil {
		return err
	}

	failures := 0
	for {
		if err := c.Probe(ctx); err != nil {
			failures++
			log.Printf("watchdog probe failed: %v (attempt %d/%d)", err, failures, c.MaxFailures)
			if failures >= c.MaxFailures {
				log.Printf("watchdog restarting %s after repeated failures", c.ServiceName)
				if restartErr := RestartService(c.ServiceName); restartErr != nil {
					log.Printf("watchdog restart failed: %v", restartErr)
				}
				failures = 0
			}
		} else {
			failures = 0
			log.Printf("watchdog health check passed for model %s", c.Model)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.Interval):
		}
	}
}

func (c WatchdogConfig) Probe(ctx context.Context) error {
	payload := map[string]any{
		"model":  c.Model,
		"prompt": c.HealthPrompt,
		"stream": false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal watchdog payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.OllamaBaseURL, "/")+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build watchdog request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("call ollama health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama health check returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read health check response: %w", err)
	}

	text := firstResponseText(data)
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("ollama health check returned empty response")
	}
	if responseLooksCorrupted(text) {
		return fmt.Errorf("ollama health check response looks corrupted: %q", text)
	}
	if !strings.Contains(strings.ToLower(text), "ok") {
		return fmt.Errorf("ollama health check response did not include expected confirmation: %q", text)
	}
	return nil
}

func RestartService(serviceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemctl", "restart", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl restart %s failed: %w: %s", serviceName, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func responseLooksCorrupted(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return true
	}
	if strings.Contains(strings.ToUpper(trimmed), "OK") && !strings.ContainsAny(trimmed, "⠀�") {
		return false
	}

	weirdCount := 0
	runes := []rune(trimmed)
	if len(runes) == 0 {
		return true
	}

	for _, r := range runes {
		switch {
		case r == '\u2800', r == '\uFFFD':
			weirdCount++
		case unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t':
			weirdCount++
		case r > 127 && !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsSpace(r) && !unicode.IsPunct(r):
			weirdCount++
		}
	}

	pct := (weirdCount * 100) / len(runes)
	return pct >= 15 || strings.Contains(trimmed, "⠀") || strings.ContainsRune(trimmed, '\uFFFD')
}

func firstResponseText(data []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return strings.TrimSpace(string(data))
	}

	if text, ok := payload["response"].(string); ok && strings.TrimSpace(text) != "" {
		return text
	}
	if text, ok := payload["content"].(string); ok && strings.TrimSpace(text) != "" {
		return text
	}
	if message, ok := payload["message"].(map[string]any); ok {
		if text, ok := message["content"].(string); ok && strings.TrimSpace(text) != "" {
			return text
		}
	}
	if text, ok := payload["text"].(string); ok && strings.TrimSpace(text) != "" {
		return text
	}
	return strings.TrimSpace(string(data))
}
