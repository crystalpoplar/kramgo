package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/crystalpoplar/kramgo/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.Provider != "openai" {
		t.Errorf("expected provider 'openai', got %q", cfg.Provider)
	}
	if cfg.Model == "" {
		t.Error("expected non-empty default model")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.json")

	// Point HOME to temp dir so ConfigFilePath resolves inside it.
	orig := os.Getenv("HOME")
	t.Setenv("HOME", dir)
	defer os.Setenv("HOME", orig)

	// Write a known config.
	cfg := &config.Config{
		Provider: "anthropic",
		Model:    "claude-3-5-sonnet-20241022",
		APIKey:   "sk-test",
	}

	// Manually write to the expected path so we don't rely on OS detection.
	_ = cfgFile // keep linter happy
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Provider != "anthropic" {
		t.Errorf("expected provider 'anthropic', got %q", loaded.Provider)
	}
	if loaded.APIKey != "sk-test" {
		t.Errorf("expected api_key 'sk-test', got %q", loaded.APIKey)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load on missing file should not error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}
