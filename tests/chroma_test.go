package tests

import (
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// TestChromaBasicIntegration verifies that Chroma can detect code and identify comments
func TestChromaBasicIntegration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		language string
		expected []string // Expected comment tokens
	}{
		{
			name:     "Go line comment",
			input:    "// This is a color test\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			language: "go",
			expected: []string{"// This is a color test"},
		},
		{
			name:     "JavaScript block comment",
			input:    "/* This has color in it */\nfunction test() {\n\treturn 'hello';\n}",
			language: "javascript",
			expected: []string{"/* This has color in it */"},
		},
		{
			name:     "Python comment",
			input:    "# This color should change\ndef hello():\n\tprint('world')",
			language: "python",
			expected: []string{"# This color should change"},
		},
		{
			name:     "Bash comment",
			input:    "#!/bin/bash\n# This color needs fixing\necho 'hello'",
			language: "bash",
			expected: []string{"# This color needs fixing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get lexer for the language
			lexer := lexers.Get(tt.language)
			if lexer == nil {
				t.Fatalf("Could not get lexer for language: %s", tt.language)
			}

			// Tokenize the input
			iterator, err := lexer.Tokenise(nil, tt.input)
			if err != nil {
				t.Fatalf("Failed to tokenize input: %v", err)
			}

			// Extract comment tokens
			var comments []string
			tokens := iterator.Tokens()
			for _, token := range tokens {
				if token.Type.InCategory(chroma.Comment) {
					comments = append(comments, token.Value)
				}
			}

			// Verify we found the expected comments
			if len(comments) == 0 {
				t.Errorf("No comments found in input")
			}

			// Check if our expected comment text is found
			found := false
			for _, comment := range comments {
				for _, expected := range tt.expected {
					if comment == expected {
						found = true
						break
					}
				}
			}

			if !found {
				t.Errorf("Expected comment not found. Got comments: %v, Expected: %v", comments, tt.expected)
			}
		})
	}
}

// TestChromaLanguageDetection verifies that Chroma can detect different languages
func TestChromaLanguageDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		filename string
		expected string
	}{
		{
			name:     "Go file",
			input:    "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			filename: "test.go",
			expected: "Go",
		},
		{
			name:     "JavaScript file",
			input:    "function test() {\n\treturn 'hello';\n}",
			filename: "test.js",
			expected: "JavaScript",
		},
		{
			name:     "Python file",
			input:    "def hello():\n\tprint('world')",
			filename: "test.py",
			expected: "Python",
		},
		{
			name:     "Bash file",
			input:    "#!/bin/bash\necho 'hello'",
			filename: "test.sh",
			expected: "Bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Try to detect language by filename
			lexer := lexers.Match(tt.filename)
			if lexer == nil {
				// Fallback to content-based detection
				lexer = lexers.Analyse(tt.input)
			}

			if lexer == nil {
				t.Fatalf("Could not detect language for input")
			}

			config := lexer.Config()
			if config == nil || config.Name != tt.expected {
				actualName := "unknown"
				if config != nil {
					actualName = config.Name
				}
				t.Errorf("Expected language %s, got %s", tt.expected, actualName)
			}
		})
	}
}

// TestChromaMarkdownCodeBlocks tests detection of code blocks in markdown
func TestChromaMarkdownCodeBlocks(t *testing.T) {
	input := `# Title

Some text here with color and flavor.

` + "```go" + `
// This comment has color in it
func main() {
	fmt.Println("Hello")
}
` + "```" + `

More text with color here.

` + "```javascript" + `
/* This comment has flavor */
function test() {
	return 'hello';
}
` + "```" + `
`

	// Get markdown lexer
	lexer := lexers.Get("markdown")
	if lexer == nil {
		t.Fatal("Could not get markdown lexer")
	}

	// Tokenize the input
	iterator, err := lexer.Tokenise(nil, input)
	if err != nil {
		t.Fatalf("Failed to tokenize markdown: %v", err)
	}

	// Look for code block tokens
	tokens := iterator.Tokens()
	foundCodeBlock := false
	for _, token := range tokens {
		if token.Type == chroma.LiteralStringBacktick ||
			token.Type.InCategory(chroma.LiteralString) {
			if len(token.Value) > 10 { // Code blocks are typically longer
				foundCodeBlock = true
				preview := token.Value
				if len(preview) > 50 {
					preview = preview[:50] + "..."
				}
				t.Logf("Found code block token: %s (type: %s)", preview, token.Type)
			}
		}
	}

	if !foundCodeBlock {
		t.Error("No code blocks found in markdown")
	}
}
