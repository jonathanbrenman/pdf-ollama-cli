package pdf

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type Extractor struct {
	binary string
}

func NewExtractor(binary string) *Extractor {
	if binary == "" {
		binary = "pdftotext"
	}

	return &Extractor{binary: binary}
}

func (e *Extractor) ExtractText(ctx context.Context, pdfPath string) (string, error) {
	cmd := exec.CommandContext(ctx, e.binary, pdfPath, "-")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run %s: %w", e.binary, err)
	}

	return out.String(), nil
}
