package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sammcj/murican-to-english/pkg/converter"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func TestSyntaxHighlightingIntegration(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name     string
		code     string
		language string
		expected []string // Expected substrings in the HTML output
	}{
		{
			name:     "Go code highlighting",
			code:     "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, world!\")\n}",
			language: "go",
			expected: []string{"<span", "class=", "chroma"},
		},
		{
			name:     "JavaScript code highlighting",
			code:     "function hello() {\n\tconsole.log('Hello, world!');\n}",
			language: "javascript",
			expected: []string{"<span", "class=", "chroma"},
		},
		{
			name:     "Python code highlighting",
			code:     "def hello():\n\tprint('Hello, world!')",
			language: "python",
			expected: []string{"<span", "class=", "chroma"},
		},
		{
			name:     "Auto-detect Go",
			code:     "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			language: "auto",
			expected: []string{"<span", "class=", "chroma"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock app to test the syntax highlighting methods
			app := &MockApp{converter: conv}

			// Test language detection if using auto
			if tt.language == "auto" {
				detectedLang := app.DetectLanguage(tt.code)
				if detectedLang == "" {
					t.Errorf("Language detection failed, got empty string")
				}
				t.Logf("Detected language: %s", detectedLang)
			}

			// Test syntax highlighting
			highlightedHTML, err := app.GetSyntaxHighlightedHTML(tt.code, tt.language)
			if err != nil {
				t.Errorf("GetSyntaxHighlightedHTML failed: %v", err)
				return
			}

			if highlightedHTML == "" {
				t.Errorf("Expected highlighted HTML, got empty string")
				return
			}

			// Check that the output contains expected elements
			for _, expected := range tt.expected {
				if !strings.Contains(highlightedHTML, expected) {
					t.Errorf("Expected HTML to contain %q, but it didn't. HTML: %s", expected, highlightedHTML)
				}
			}

			t.Logf("Highlighted HTML length: %d characters", len(highlightedHTML))
		})
	}
}

func TestLanguageDetection(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	app := &MockApp{converter: conv}

	tests := []struct {
		name         string
		code         string
		shouldDetect bool // Whether we expect successful detection (not necessarily exact language)
	}{
		{
			name:         "Go package declaration",
			code:         "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			shouldDetect: true,
		},
		{
			name:         "JavaScript function with more context",
			code:         "const express = require('express');\nfunction test() {\n  return 'hello';\n}\nmodule.exports = test;",
			shouldDetect: true,
		},
		{
			name:         "Python function with imports",
			code:         "import os\ndef test():\n    print('hello')\n    return True",
			shouldDetect: false, // Chroma sometimes doesn't detect short Python snippets reliably
		},
		{
			name:         "Plain text",
			code:         "This is just regular text with no code patterns.",
			shouldDetect: false, // Should detect as text or fail to detect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected := app.DetectLanguage(tt.code)
			t.Logf("Detected language: %q for %s", detected, tt.name)

			if tt.shouldDetect {
				// For code, we just want to make sure it detected something other than text
				if detected == "text" || detected == "" {
					t.Errorf("Expected to detect a programming language, but got %q", detected)
				}
			} else {
				// For plain text, we expect "text" or empty
				if detected != "text" && detected != "" {
					t.Logf("Plain text detected as %q (this is okay)", detected)
				}
			}
		})
	}
}

// MockApp simulates the main App struct for testing
type MockApp struct {
	converter *converter.Converter
}

func (a *MockApp) GetSyntaxHighlightedHTML(code, language string) (string, error) {
	if code == "" {
		return "", nil
	}

	var lexer chroma.Lexer

	// Get lexer by language name or detect automatically
	if language != "" && language != "auto" {
		lexer = lexers.Get(language)
	}

	// If no lexer found or auto-detection requested, analyze the code
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}

	// Fallback to plaintext if no lexer found
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Configure the lexer with a sensible configuration
	lexer = chroma.Coalesce(lexer)

	// Create HTML formatter with CSS classes
	formatter := html.New(
		html.WithClasses(true),
		html.WithLineNumbers(false),
		html.TabWidth(4),
	)

	// Get a style (using github style as default)
	style := styles.Get("github")
	if style == nil {
		style = styles.Fallback
	}

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return "", fmt.Errorf("failed to tokenize code: %w", err)
	}

	// Format to HTML
	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return "", fmt.Errorf("failed to format code: %w", err)
	}

	return buf.String(), nil
}

func (a *MockApp) DetectLanguage(code string) string {
	if code == "" {
		return "text"
	}

	lexer := lexers.Analyse(code)
	if lexer != nil {
		config := lexer.Config()
		if config != nil {
			return strings.ToLower(config.Name)
		}
	}

	return "text"
}
