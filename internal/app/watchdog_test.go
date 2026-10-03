package app

import "testing"

func TestResponseLooksCorrupted(t *testing.T) {
	corrupted := " of⠀⠀.,⠀-s-⠀⠀⠀⠀,,⠀,-.-,.\no in⠀⠀�⠀⠀⠀⠀..."
	if !responseLooksCorrupted(corrupted) {
		t.Fatalf("expected corrupted output to be detected")
	}

	if responseLooksCorrupted("OK") {
		t.Fatalf("expected simple OK response to be healthy")
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
