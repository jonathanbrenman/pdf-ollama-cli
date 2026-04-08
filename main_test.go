package main

import "testing"

func TestParseCLIArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedPath string
		expectedLang string
		expectError  bool
	}{
		{
			name:         "uses default language",
			args:         []string{"document.pdf"},
			expectedPath: "document.pdf",
			expectedLang: "Spanish",
		},
		{
			name:         "uses long language flag",
			args:         []string{"--language", "English", "document.pdf"},
			expectedPath: "document.pdf",
			expectedLang: "English",
		},
		{
			name:         "uses lang alias",
			args:         []string{"--lang", "French", "document.pdf"},
			expectedPath: "document.pdf",
			expectedLang: "French",
		},
		{
			name:        "fails without path",
			args:        []string{"--language", "German"},
			expectError: true,
		},
		{
			name:        "fails with extra positional args",
			args:        []string{"first.pdf", "second.pdf"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseCLIArgs(tt.args)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if parsed.pdfPath != tt.expectedPath {
				t.Fatalf("expected path %q, got %q", tt.expectedPath, parsed.pdfPath)
			}

			if parsed.language != tt.expectedLang {
				t.Fatalf("expected language %q, got %q", tt.expectedLang, parsed.language)
			}
		})
	}
}
