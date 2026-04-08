package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"pdf-ollama-cli/config"
	"pdf-ollama-cli/internal/app/summarizer"
	"pdf-ollama-cli/internal/infra/ollama"
	"pdf-ollama-cli/internal/infra/pdf"
	"pdf-ollama-cli/internal/infra/terminal"
)

type cliArgs struct {
	pdfPath  string
	language string
}

func main() {
	args, err := parseCLIArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Usage: cli [--language <language>] <file.pdf>")
		os.Exit(1)
	}

	cfg := config.Load()

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

	final, err := service.SummarizePDF(context.Background(), args.pdfPath, args.language)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	rendered := final
	renderer, err := terminal.NewMarkdownRenderer(100)
	if err == nil {
		styled, renderErr := renderer.Render(final)
		if renderErr == nil {
			rendered = styled
		}
	}

	fmt.Println("Final result:")
	fmt.Println(rendered)
}

func parseCLIArgs(args []string) (cliArgs, error) {
	flagSet := flag.NewFlagSet("pdf-ollama-cli", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	language := flagSet.String("language", "Spanish", "Language used for the generated summary")
	flagSet.StringVar(language, "lang", "Spanish", "Language used for the generated summary")

	if err := flagSet.Parse(args); err != nil {
		return cliArgs{}, err
	}

	remainingArgs := flagSet.Args()
	if len(remainingArgs) != 1 {
		return cliArgs{}, fmt.Errorf("expected exactly one PDF file path")
	}

	return cliArgs{
		pdfPath:  remainingArgs[0],
		language: *language,
	}, nil
}
