package tests

import (
	"strings"
	"testing"

	"github.com/sammcj/murican-to-english/pkg/converter"
)

// TestDetectMarkdownCodeBlocks tests detection of fenced code blocks
func TestDetectMarkdownCodeBlocks(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected int // Number of code blocks expected
	}{
		{
			name:     "Go code block",
			input:    "Some text\n\n```go\n// This has color in it\nfunc main() {}\n```\n\nMore text",
			expected: 1,
		},
		{
			name:     "JavaScript code block",
			input:    "```javascript\n/* This has flavor */\nfunction test() {}\n```",
			expected: 1,
		},
		{
			name:     "Multiple code blocks",
			input:    "```go\npackage main\n```\n\nSome text\n\n```js\nconsole.log();\n```",
			expected: 2,
		},
		{
			name:     "No code blocks",
			input:    "Just regular text with color and flavor words",
			expected: 0,
		},
		{
			name:     "Tilde fenced blocks",
			input:    "~~~python\n# This color should change\nprint('hello')\n~~~",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := conv.DetectCodeBlocks(tt.input)
			// Count only the fenced code blocks
			fencedBlocks := 0
			for _, block := range blocks {
				if block.IsCode && (strings.Contains(tt.input, "```") || strings.Contains(tt.input, "~~~")) {
					fencedBlocks++
				}
			}
			if fencedBlocks != tt.expected {
				t.Errorf("Expected %d code blocks, got %d", tt.expected, fencedBlocks)
			}

			// Verify all detected code blocks are marked as code
			for _, block := range blocks {
				if block.IsCode && !block.IsCode {
					t.Errorf("Code block should be marked as code")
				}
			}
		})
	}
}

// TestDetectInlineCode tests detection of inline code
func TestDetectInlineCode(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Single inline code",
			input:    "Use `color` instead of `colour` in code",
			expected: 2,
		},
		{
			name:     "No inline code",
			input:    "Just regular text",
			expected: 0,
		},
		{
			name:     "Mixed with newlines",
			input:    "Code `variable` and more\ntext with `function()`",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := conv.DetectCodeBlocks(tt.input)
			// Count only inline code blocks
			inlineBlocks := 0
			for _, block := range blocks {
				if block.IsCode && strings.Contains(tt.input, "`") && !strings.Contains(tt.input, "```") {
					inlineBlocks++
				}
			}
			if inlineBlocks != tt.expected {
				t.Errorf("Expected %d inline code blocks, got %d", tt.expected, inlineBlocks)
			}
		})
	}
}

// TestExtractComments tests comment extraction from code
func TestExtractComments(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name     string
		code     string
		language string
		expected int // Number of comments expected
	}{
		{
			name:     "Go line comment",
			code:     "// This color should change\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
			language: "go",
			expected: 1,
		},
		{
			name:     "JavaScript block comment",
			code:     "/* This has color and flavor */\nfunction test() {\n\treturn 'hello';\n}",
			language: "javascript",
			expected: 1,
		},
		{
			name:     "Python comment",
			code:     "# Change color to colour\ndef hello():\n\tprint('world')",
			language: "python",
			expected: 1,
		},
		{
			name:     "Multiple comments",
			code:     "// First color comment\nvar x = 1;\n// Second flavor comment\nvar y = 2;",
			language: "javascript",
			expected: 2,
		},
		{
			name:     "No comments",
			code:     "function test() {\n\treturn 'hello';\n}",
			language: "javascript",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments := conv.ExtractComments(tt.code, tt.language)
			if len(comments) != tt.expected {
				t.Errorf("Expected %d comments, got %d", tt.expected, len(comments))
				for i, comment := range comments {
					t.Logf("Comment %d: %q", i, comment.Content)
				}
			}
		})
	}
}

// TestProcessCodeAware tests the main code-aware processing function
func TestProcessCodeAware(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Regular text only",
			input:    "This has American color and flavor words that should change.",
			expected: "This has American colour and flavour words that should change.",
		},
		{
			name:     "Code block with comment",
			input:    "```go\n// This color should change to colour\nfunc main() {\n\tfmt.Println(\"color\") // keep this color\n}\n```",
			expected: "```go\n// This colour should change to colour\nfunc main() {\n\tfmt.Println(\"color\") // keep this colour\n}\n```",
		},
		{
			name:     "Mixed text and code",
			input:    "This color should change.\n\n```go\nfmt.Println(\"color\") // but this color should change\n```\n\nThis flavor should also change.",
			expected: "This colour should change.\n\n```go\nfmt.Println(\"color\") // but this colour should change\n```\n\nThis flavour should also change.",
		},
		{
			name:     "Inline code preservation",
			input:    "Use `color` variable but this color should change.",
			expected: "Use `color` variable but this colour should change.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.ProcessCodeAware(tt.input, true)
			if result != tt.expected {
				t.Errorf("ProcessCodeAware failed\nInput:    %q\nExpected: %q\nGot:      %q", tt.input, tt.expected, result)
			}
		})
	}
}

// TestCodeBlockDetection tests the overall code block detection
func TestCodeBlockDetection(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	input := `# Document Title

This text has color and flavor that should change.

` + "```go" + `
// This comment has color that should change
func main() {
    fmt.Println("color") // This color should also change
    var color = "red"    // Don't change this variable name
}
` + "```" + `

More text with color here.

` + "```javascript" + `
/* This comment has flavor */
function getColor() {
    return "color"; // keep the string literal
}
` + "```" + `

Final text with flavor.`

	blocks := conv.DetectCodeBlocks(input)

	// Should detect: text, code, text, code, text (5 blocks total)
	expectedBlocks := 5
	if len(blocks) != expectedBlocks {
		t.Errorf("Expected %d blocks, got %d", expectedBlocks, len(blocks))
		for i, block := range blocks {
			t.Logf("Block %d: IsCode=%t, Language=%s, Content=%.50s...", i, block.IsCode, block.Language, strings.ReplaceAll(block.Content, "\n", "\\n"))
		}
	}

	// Check that code blocks are identified correctly
	codeBlockCount := 0
	for _, block := range blocks {
		if block.IsCode {
			codeBlockCount++
		}
	}

	expectedCodeBlocks := 2
	if codeBlockCount != expectedCodeBlocks {
		t.Errorf("Expected %d code blocks, got %d", expectedCodeBlocks, codeBlockCount)
	}
}

// TestManualCommentExtraction tests the fallback comment extraction
func TestManualCommentExtraction(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name:     "C++ style comments",
			code:     "// This color comment\nint x = 1;\n/* Block color comment */",
			expected: 2,
		},
		{
			name:     "Python style comments",
			code:     "# This color comment\nx = 1\n# Another flavor comment",
			expected: 2,
		},
		{
			name:     "Mixed comment styles",
			code:     "// Line comment\n/* Block comment */\n# Hash comment",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments := conv.ExtractComments(tt.code, "")
			if len(comments) != tt.expected {
				t.Errorf("Expected %d comments, got %d", tt.expected, len(comments))
				for i, comment := range comments {
					t.Logf("Comment %d: %q", i, comment.Content)
				}
			}
		})
	}
}
