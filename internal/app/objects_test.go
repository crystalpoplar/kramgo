package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteJSONFileAndGetJSONFile(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	path := filepath.Join(t.TempDir(), "kramgo.json")
	input := payload{Name: "david", Age: 30}

	if err := WriteJSONFile(path, input); err != nil {
		t.Fatalf("WriteJSONFile returned error: %v", err)
	}

	raw, err := GetJSONFile(path)
	if err != nil {
		t.Fatalf("GetJSONFile returned error: %v", err)
	}

	var output payload
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if output != input {
		t.Fatalf("unexpected JSON content: got %+v, want %+v", output, input)
	}
}

func TestWriteJSONFileReturnsCreateError(t *testing.T) {
	t.Parallel()

	// Creating inside a missing parent directory should fail.
	path := filepath.Join(t.TempDir(), "missing", "kramgo.json")
	if err := WriteJSONFile(path, map[string]string{"ok": "no"}); err == nil {
		t.Fatal("WriteJSONFile returned nil error for invalid path")
	}
}

func TestGetJSONFileReturnsReadError(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected missing file before test, got err=%v", err)
	}

	if _, err := GetJSONFile(path); err == nil {
		t.Fatal("GetJSONFile returned nil error for missing file")
	}
}
