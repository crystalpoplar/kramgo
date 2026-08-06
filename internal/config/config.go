// Package config handles application configuration for kramgo.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds the application configuration.
type Config struct {
	// Provider is the AI provider to use (e.g. "openai", "anthropic").
	Provider string `json:"provider"`
	// APIKey is the API key for the chosen provider.
	APIKey string `json:"api_key"`
	// Model is the model identifier to use for completions.
	Model string `json:"model"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Provider: "openai",
		Model:    "gpt-4o",
	}
}

// configDir returns the platform-appropriate directory for kramgo config files.
func configDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", errors.New("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "kramgo"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config", "kramgo"), nil
	}
}

// ConfigFilePath returns the path to the kramgo config file.
func ConfigFilePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config file and returns a Config. If the file does not exist,
// the default config is returned.
func Load() (*Config, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save writes cfg to the config file, creating directories as needed.
func Save(cfg *Config) error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
