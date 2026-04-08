# pdf-ollama-cli

A Go CLI that extracts text from a PDF, summarizes it in parallel chunks with Ollama, and then builds a consolidated final summary.

## Overview

The tool runs a two-stage summarization pipeline:

1. Extract raw text from a PDF using `pdftotext`.
2. Split text into chunks and summarize each chunk concurrently.
3. Merge partial summaries and request a final executive summary.

This design improves throughput for large documents while keeping the summarization logic isolated from infrastructure concerns.

## Architecture

The project follows a layered architecture:

- `main.go`: composition root and CLI entrypoint.
- `config/`: configuration loading from environment variables with defaults.
- `internal/app/summarizer/`: application service and business flow orchestration.
- `internal/infra/pdf/`: PDF extraction adapter (`pdftotext`).
- `internal/infra/ollama/`: HTTP adapter for Ollama API.

### Dependency direction

- `main` depends on `config`, `internal/app`, and `internal/infra`.
- `internal/app` depends only on interfaces (ports), not concrete infrastructure.
- `internal/infra` implements those interfaces.

## Requirements

- Go `1.23+`
- Ollama running locally or remotely with an available model
- `pdftotext` installed and available in PATH

### Install pdftotext (Linux)

```bash
sudo apt-get update && sudo apt-get install -y poppler-utils
```

## Configuration

The application uses environment variables with safe defaults:

| Variable | Default | Description |
| --- | --- | --- |
| `CHUNK_SIZE` | `2000` | Approximate number of characters per chunk |
| `NUM_WORKERS` | `4` | Number of concurrent chunk workers |
| `OLLAMA_MODEL` | `gemma4:e4b` | Ollama model used for generation |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama base URL |
| `PDFTOTEXT_BIN` | `pdftotext` | Path or command name for the PDF extractor binary |
| `HTTP_TIMEOUT_SECONDS` | `120` | HTTP timeout used by the Ollama client |

## Quick Start

1. Ensure Ollama is running and the model is available.
2. Export custom environment variables only if needed.
3. Run the CLI with a PDF path.

```bash
go run . /path/to/file.pdf
```

Expected output:

- A final summary printed to stdout.
- Any execution errors printed to stderr with a non-zero exit code.

## Build

```bash
go build ./...
```

## Example with custom settings

```bash
export OLLAMA_MODEL="gemma4:e4b"
export NUM_WORKERS=6
export CHUNK_SIZE=3000
go run . ./sample.pdf
```

## Troubleshooting

- `run pdftotext: ...`: install `poppler-utils` or set `PDFTOTEXT_BIN` to the correct binary.
- `send request: ...`: verify Ollama is running and reachable from `OLLAMA_URL`.
- `ollama returned status ...`: check model availability and request limits in Ollama.
- Empty or weak summaries: tune `CHUNK_SIZE`, `NUM_WORKERS`, and the selected model.

## Notes

- The current CLI accepts exactly one positional argument: the PDF path.
- Prompts are defined in the summarizer service and can be evolved without changing infrastructure adapters.
