package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ChunkSize       int
	NumWorkers      int
	OllamaModel     string
	OllamaURL       string
	PDFToTextBinary string
	HTTPTimeout     time.Duration
}

func Load() Config {
	return Config{
		ChunkSize:       getInt("CHUNK_SIZE", 2000),
		NumWorkers:      getInt("NUM_WORKERS", 4),
		OllamaModel:     getString("OLLAMA_MODEL", "gemma4:e4b"),
		OllamaURL:       getString("OLLAMA_URL", "http://localhost:11434"),
		PDFToTextBinary: getString("PDFTOTEXT_BIN", "pdftotext"),
		HTTPTimeout:     time.Duration(getInt("HTTP_TIMEOUT_SECONDS", 120)) * time.Second,
	}
}

func getString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return parsed
}
