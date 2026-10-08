package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResponseLooksCorrupted(t *testing.T) {
	corrupted := " of⠀⠀.,⠀-s-⠀⠀⠀⠀,,⠀,-.-,.\no in⠀⠀�⠀⠀⠀⠀..."
	if !responseLooksCorrupted(corrupted) {
		t.Fatalf("expected corrupted output to be detected")
	}

	if responseLooksCorrupted("OK") {
		t.Fatalf("expected simple OK response to be healthy")
	}
}

func TestHealthResponseValidation(t *testing.T) {
	for _, s := range []string{"OK", "ok", "OK.", " OK\n", "okay"} {
		if !looksLikeOKResponse(s) {
			t.Fatalf("expected %q to pass as a healthy response", s)
		}
	}

	for _, s := range []string{".", "PRESIDIO", "The system is okay", "blah blah"} {
		if looksLikeOKResponse(s) {
			t.Fatalf("expected %q to fail as a healthy response", s)
		}
	}
}

func TestWatchdogConfigValidatesRequiredFields(t *testing.T) {
	cfg := WatchdogConfig{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for empty watchdog config")
	}

	cfg = WatchdogConfig{
		OllamaBaseURL: "http://127.0.0.1:11434",
		Model:         "crystal:latest",
		ServiceName:   "ollama",
		Interval:      30 * 1000000000,
		MaxFailures:   3,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected config to validate, got error: %v", err)
	}
}

func TestWatchdogServiceUsesTenMinuteInterval(t *testing.T) {
	paths := []string{
		filepath.Join("..", "..", "kramgo-watchdog.service"),
		filepath.Join("..", "..", "dist", "linux-amd64", "kramgo-watchdog.service"),
	}
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if !strings.Contains(string(content), "--interval 10m") {
			t.Fatalf("%s should contain --interval 10m, got: %s", p, string(content))
		}
	}
}
