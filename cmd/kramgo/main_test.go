package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSuccess(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "kramgo.json")
	if err := os.WriteFile(path, []byte("{\"hello\":\"world\"}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	var buf bytes.Buffer
	if err := run(&buf, path); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "kramgo is ready\n") {
		t.Fatalf("missing ready line in output: %q", got)
	}
	if !strings.Contains(got, "{\"hello\":\"world\"}") {
		t.Fatalf("missing JSON content in output: %q", got)
	}
}

func TestRunReadError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := run(&buf, filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("run returned nil error for missing json file")
	}
	if !strings.Contains(err.Error(), "read json file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMainSuccess(t *testing.T) {
	tempDir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	if err := os.WriteFile("kramgo.json", []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	main()
}
