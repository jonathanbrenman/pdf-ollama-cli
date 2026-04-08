package summarizer

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockExtractor struct {
	text string
	err  error
}

func (m *mockExtractor) ExtractText(ctx context.Context, path string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return m.text, m.err
	}
}

type mockLLM struct {
	responses map[string]string
	err       error
	finalErr  error
	prompts   []string
}

func (m *mockLLM) Generate(ctx context.Context, prompt string) (string, error) {
	m.prompts = append(m.prompts, prompt)
	if m.err != nil {
		return "", m.err
	}
	if strings.Contains(prompt, "You have partial") && m.finalErr != nil {
		return "", m.finalErr
	}
	for k, v := range m.responses {
		if strings.Contains(prompt, k) {
			return v, nil
		}
	}
	return "generic response", nil
}

type syncMockLLM struct {
	mockLLM
	startWait chan struct{}
	endWait   chan struct{}
}

func (m *syncMockLLM) Generate(ctx context.Context, prompt string) (string, error) {
	if m.startWait != nil {
		m.startWait <- struct{}{}
	}
	if m.endWait != nil {
		<-m.endWait
	}
	return m.mockLLM.Generate(ctx, prompt)
}

func TestWorker(t *testing.T) {
	t.Run("worker context cancelled during send", func(t *testing.T) {
		start := make(chan struct{})
		end := make(chan struct{})
		llm := &syncMockLLM{
			startWait: start,
			endWait:   end,
		}

		ctx, cancel := context.WithCancel(context.Background())
		s := &Service{llm: llm, numWorkers: 1}
		jobs := make(chan job, 1)
		results := make(chan result)

		go s.worker(ctx, jobs, results, "Spanish")

		jobs <- job{id: 1, chunk: "test"}
		<-start    // Wait for LLM to be called
		cancel()   // Cancel context while LLM is "working"
		close(end) // Let LLM finish

		// Worker should exit without sending to results
	})

	t.Run("worker jobs closed", func(t *testing.T) {
		ctx := context.Background()
		s := &Service{}
		jobs := make(chan job)
		results := make(chan result)
		close(jobs)
		s.worker(ctx, jobs, results, "Spanish")
	})

	t.Run("worker context cancelled initially", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s := &Service{}
		jobs := make(chan job)
		results := make(chan result)
		s.worker(ctx, jobs, results, "Spanish")
	})
}

func TestSplitText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		size     int
		expected []string
	}{
		{"empty text", "", 10, nil},
		{"smaller than chunk", "hello", 10, []string{"hello"}},
		{"exact chunk", "hello", 5, []string{"hello"}},
		{"larger than chunk", "hello world", 5, []string{"hello", " worl", "d"}},
		{"size 0", "hello", 0, []string{"hello"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitText(tt.text, tt.size)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(got))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("expected %q, got %q at index %d", tt.expected[i], got[i], i)
				}
			}
		})
	}
}

func TestSummarizePDF(t *testing.T) {
	t.Run("successful summary", func(t *testing.T) {
		extractor := &mockExtractor{text: "This is a long document text"}
		llm := &mockLLM{
			responses: map[string]string{
				"Summarize this fragment": "fragment summary",
				"You have partial":        "final summary",
			},
		}

		svc := NewService(extractor, llm, Options{ChunkSize: 10, NumWorkers: 1})
		summary, err := svc.SummarizePDF(context.Background(), "test.pdf", "Spanish")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary != "final summary" {
			t.Errorf("expected %q, got %q", "final summary", summary)
		}
	})

	t.Run("extractor error", func(t *testing.T) {
		extractor := &mockExtractor{err: errors.New("extract error")}
		llm := &mockLLM{}

		svc := NewService(extractor, llm, Options{})
		_, err := svc.SummarizePDF(context.Background(), "test.pdf", "Spanish")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("llm error", func(t *testing.T) {
		extractor := &mockExtractor{text: "some text"}
		llm := &mockLLM{err: errors.New("llm error")}

		svc := NewService(extractor, llm, Options{})
		_, err := svc.SummarizePDF(context.Background(), "test.pdf", "Spanish")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("final summary error", func(t *testing.T) {
		extractor := &mockExtractor{text: "some text"}
		llm := &mockLLM{finalErr: errors.New("final error")}

		svc := NewService(extractor, llm, Options{})
		_, err := svc.SummarizePDF(context.Background(), "test.pdf", "Spanish")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("empty text error", func(t *testing.T) {
		extractor := &mockExtractor{text: ""}
		llm := &mockLLM{}

		svc := NewService(extractor, llm, Options{})
		_, err := svc.SummarizePDF(context.Background(), "test.pdf", "Spanish")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		extractor := &mockExtractor{text: "some text"}
		llm := &mockLLM{}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		svc := NewService(extractor, llm, Options{})
		_, err := svc.SummarizePDF(ctx, "test.pdf", "Spanish")

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("final summary uses requested language", func(t *testing.T) {
		extractor := &mockExtractor{text: "some text that will be summarized"}
		llm := &mockLLM{
			responses: map[string]string{
				"Summarize this fragment": "fragment summary",
				"You have partial":        "final summary",
			},
		}

		svc := NewService(extractor, llm, Options{ChunkSize: 100, NumWorkers: 1})
		_, err := svc.SummarizePDF(context.Background(), "test.pdf", "English")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var finalPrompt string
		for _, prompt := range llm.prompts {
			if strings.Contains(prompt, "You have partial summaries for a document") {
				finalPrompt = prompt
				break
			}
		}

		if finalPrompt == "" {
			t.Fatal("expected final prompt to be captured")
		}

		if !strings.Contains(finalPrompt, "The output must be in English.") {
			t.Fatalf("expected final prompt to contain selected language, got %q", finalPrompt)
		}
	})
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "uses explicit language", input: "English", expected: "English"},
		{name: "trims whitespace", input: "  Portuguese  ", expected: "Portuguese"},
		{name: "defaults when empty", input: "", expected: "Spanish"},
		{name: "defaults when blank", input: "   ", expected: "Spanish"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeLanguage(tt.input)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
