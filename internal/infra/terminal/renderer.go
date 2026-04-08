package terminal

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/glamour"
)

type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
}

func NewMarkdownRenderer(width int) (*MarkdownRenderer, error) {
	if width <= 0 {
		width = 100
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}

	return &MarkdownRenderer{renderer: renderer}, nil
}

func (r *MarkdownRenderer) Render(markdown string) (string, error) {
	return r.renderer.Render(sanitizeTerminalInput(markdown))
}

func sanitizeTerminalInput(input string) string {
	return strings.Map(func(value rune) rune {
		switch value {
		case '\n', '\r', '\t':
			return value
		}

		if unicode.IsControl(value) {
			return -1
		}

		return value
	}, input)
}
