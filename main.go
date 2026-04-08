package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"pdf-ollama-cli/config"
	"pdf-ollama-cli/internal/app/summarizer"
	"pdf-ollama-cli/internal/infra/ollama"
	"pdf-ollama-cli/internal/infra/pdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: cli <file.pdf>")
		return
	}

	cfg := config.Load()
	pdfPath := os.Args[1]

	pdfExtractor := pdf.NewExtractor(cfg.PDFToTextBinary)
	ollamaClient := ollama.NewClient(
		cfg.OllamaURL,
		cfg.OllamaModel,
		&http.Client{Timeout: cfg.HTTPTimeout},
	)
	service := summarizer.NewService(pdfExtractor, ollamaClient, summarizer.Options{
		ChunkSize:  cfg.ChunkSize,
		NumWorkers: cfg.NumWorkers,
	})

	final, err := service.SummarizePDF(context.Background(), pdfPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Final result:\n")
	fmt.Println(final)
}
