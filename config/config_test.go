package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		// Clear env vars to ensure we use fallbacks
		os.Unsetenv("CHUNK_SIZE")
		os.Unsetenv("NUM_WORKERS")
		os.Unsetenv("OLLAMA_MODEL")
		os.Unsetenv("OLLAMA_URL")
		os.Unsetenv("PDFTOTEXT_BIN")

		cfg := Load()

		if cfg.ChunkSize != 2000 {
			t.Errorf("expected 2000, got %d", cfg.ChunkSize)
		}
		if cfg.OllamaModel != "gemma4:e4b" {
			t.Errorf("expected gemma4:e4b, got %s", cfg.OllamaModel)
		}
	})

	t.Run("environment overrides", func(t *testing.T) {
		os.Setenv("CHUNK_SIZE", "5000")
		os.Setenv("OLLAMA_MODEL", "llama2")
		defer os.Unsetenv("CHUNK_SIZE")
		defer os.Unsetenv("OLLAMA_MODEL")

		cfg := Load()

		if cfg.ChunkSize != 5000 {
			t.Errorf("expected 5000, got %d", cfg.ChunkSize)
		}
		if cfg.OllamaModel != "llama2" {
			t.Errorf("expected llama2, got %s", cfg.OllamaModel)
		}
	})

	t.Run("invalid int value", func(t *testing.T) {
		os.Setenv("CHUNK_SIZE", "not-a-number")
		defer os.Unsetenv("CHUNK_SIZE")

		cfg := Load()

		if cfg.ChunkSize != 2000 {
			t.Errorf("expected fallback 2000, got %d", cfg.ChunkSize)
		}
	})
}
