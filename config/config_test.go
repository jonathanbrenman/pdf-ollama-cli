package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		t.Setenv("CHUNK_SIZE", "")
		t.Setenv("NUM_WORKERS", "")
		t.Setenv("OLLAMA_MODEL", "")
		t.Setenv("OLLAMA_URL", "")
		t.Setenv("PDFTOTEXT_BIN", "")

		cfg := Load()

		if cfg.ChunkSize != 2000 {
			t.Errorf("expected 2000, got %d", cfg.ChunkSize)
		}
		if cfg.OllamaModel != "gemma4:e4b" {
			t.Errorf("expected gemma4:e4b, got %s", cfg.OllamaModel)
		}
	})

	t.Run("environment overrides", func(t *testing.T) {
		t.Setenv("CHUNK_SIZE", "5000")
		t.Setenv("OLLAMA_MODEL", "llama2")

		cfg := Load()

		if cfg.ChunkSize != 5000 {
			t.Errorf("expected 5000, got %d", cfg.ChunkSize)
		}
		if cfg.OllamaModel != "llama2" {
			t.Errorf("expected llama2, got %s", cfg.OllamaModel)
		}
	})

	t.Run("invalid int value", func(t *testing.T) {
		t.Setenv("CHUNK_SIZE", "not-a-number")

		cfg := Load()

		if cfg.ChunkSize != 2000 {
			t.Errorf("expected fallback 2000, got %d", cfg.ChunkSize)
		}
	})
}
