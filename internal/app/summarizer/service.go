package summarizer

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type TextExtractor interface {
	ExtractText(ctx context.Context, path string) (string, error)
}

type LLMClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type Options struct {
	ChunkSize  int
	NumWorkers int
}

type Service struct {
	extractor  TextExtractor
	llm        LLMClient
	chunkSize  int
	numWorkers int
}

type job struct {
	id    int
	chunk string
}

type result struct {
	id      int
	content string
	err     error
}

func NewService(extractor TextExtractor, llm LLMClient, opts Options) *Service {
	chunkSize := opts.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 2000
	}

	numWorkers := opts.NumWorkers
	if numWorkers <= 0 {
		numWorkers = 4
	}

	return &Service{
		extractor:  extractor,
		llm:        llm,
		chunkSize:  chunkSize,
		numWorkers: numWorkers,
	}
}

func (s *Service) SummarizePDF(ctx context.Context, pdfPath string) (string, error) {
	text, err := s.extractor.ExtractText(ctx, pdfPath)
	if err != nil {
		return "", fmt.Errorf("extract text: %w", err)
	}

	partial, err := s.processText(ctx, text)
	if err != nil {
		return "", fmt.Errorf("process text: %w", err)
	}

	final, err := s.finalSummary(ctx, partial)
	if err != nil {
		return "", fmt.Errorf("final summary: %w", err)
	}

	return final, nil
}

func (s *Service) processText(ctx context.Context, text string) (string, error) {
	chunks := splitText(text, s.chunkSize)
	if len(chunks) == 0 {
		return "", fmt.Errorf("no text extracted from pdf")
	}

	jobs := make(chan job, len(chunks))
	results := make(chan result, len(chunks))

	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < s.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.worker(workerCtx, jobs, results)
		}()
	}

	for i, chunk := range chunks {
		jobs <- job{id: i, chunk: chunk}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	outputs := make([]string, len(chunks))
	var firstErr error

	for res := range results {
		if res.err != nil {
			if firstErr == nil {
				firstErr = res.err
				cancel()
			}
			continue
		}
		outputs[res.id] = res.content
	}

	if firstErr != nil {
		return "", firstErr
	}

	return strings.Join(outputs, "\n\n"), nil
}

func (s *Service) worker(ctx context.Context, jobs <-chan job, results chan<- result) {
	for {
		select {
		case <-ctx.Done():
			return
		case currentJob, ok := <-jobs:
			if !ok {
				return
			}

			prompt := fmt.Sprintf(`Act as a technical analyst.
Summarize this fragment with concise bullet points:

%s`, currentJob.chunk)

			resp, err := s.llm.Generate(ctx, prompt)

			select {
			case results <- result{id: currentJob.id, content: resp, err: err}:
			case <-ctx.Done():
				return
			}
		}
	}
}

func splitText(text string, size int) []string {
	if len(text) == 0 {
		return nil
	}

	if size <= 0 {
		size = len(text)
	}

	chunks := make([]string, 0, (len(text)/size)+1)

	for start := 0; start < len(text); start += size {
		end := start + size
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
	}

	return chunks
}

func (s *Service) finalSummary(ctx context.Context, intermediate string) (string, error) {
	prompt := fmt.Sprintf(`You have partial summaries for a document.
Unify everything into:
- executive summary
- key points
- conclusions

%s`, intermediate)

	return s.llm.Generate(ctx, prompt)
}
