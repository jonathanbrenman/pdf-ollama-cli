package pdf

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractor_ExtractText(t *testing.T) {
	// Create a temporary script to act as the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "mock_pdftotext")
	
	// The script just prints its arguments to help us verify it was called
	// or just prints some fixed text if we want to simulate output.
	scriptContent := "#!/bin/sh\necho \"extracted text\""
	if err := os.WriteFile(binaryPath, []byte(scriptContent), 0755); err != nil {
		t.Fatal(err)
	}

	extractor := NewExtractor(binaryPath)

	t.Run("successful extraction", func(t *testing.T) {
		got, err := extractor.ExtractText(context.Background(), "dummy.pdf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "extracted text\n"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("binary not found", func(t *testing.T) {
		badExtractor := NewExtractor("non_existent_binary")
		_, err := badExtractor.ExtractText(context.Background(), "dummy.pdf")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestNewExtractor(t *testing.T) {
	t.Run("default binary", func(t *testing.T) {
		e := NewExtractor("")
		if e.binary != "pdftotext" {
			t.Errorf("expected pdftotext, got %s", e.binary)
		}
	})

	t.Run("custom binary", func(t *testing.T) {
		e := NewExtractor("custom")
		if e.binary != "custom" {
			t.Errorf("expected custom, got %s", e.binary)
		}
	})
}
