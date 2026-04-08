package terminal

import (
	"strings"
	"testing"
)

func TestSanitizeTerminalInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal text", "hello world", "hello world"},
		{"escape character", "hello\x1b[31mworld", "hello[31mworld"},
		{"bell character", "hello\x07world", "helloworld"},
		{"newlines are kept", "hello\nworld", "hello\nworld"},
		{"tabs are kept", "hello\tworld", "hello\tworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeTerminalInput(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestMarkdownRenderer(t *testing.T) {
	t.Run("successful rendering", func(t *testing.T) {
		renderer, err := NewMarkdownRenderer(80)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out, err := renderer.Render("# Title\n\n- item")
		if err != nil {
			t.Fatalf("unexpected rendering error: %v", err)
		}

		if strings.TrimSpace(out) == "" {
			t.Fatal("expected non-empty output")
		}
	})

	t.Run("default width", func(t *testing.T) {
		renderer, err := NewMarkdownRenderer(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if renderer == nil {
			t.Fatal("expected non-nil renderer")
		}
	})
}
