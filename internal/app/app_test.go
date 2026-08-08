package app

import (
	"bytes"
	"errors"
	"testing"
)

type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestRunWritesReadyMessage(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if got, want := buf.String(), "kramgo is ready\n"; got != want {
		t.Fatalf("unexpected output: got %q, want %q", got, want)
	}
}

func TestRunReturnsWriterError(t *testing.T) {
	err := Run(errWriter{})
	if err == nil {
		t.Fatal("Run returned nil error, want write error")
	}
}
