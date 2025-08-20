package tests

import (
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestIgnoreCommentValidation(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name         string
		input        string
		shouldDetect bool
		description  string
	}{
		// Valid comments - should detect
		{
			name:         "Valid hash comment",
			input:        "# m2e-ignore",
			shouldDetect: true,
			description:  "Simple hash comment should be detected",
		},
		{
			name:         "Valid hash comment with space",
			input:        "    # m2e-ignore",
			shouldDetect: true,
			description:  "Indented hash comment should be detected",
		},
		{
			name:         "Valid double slash comment",
			input:        "// m2e-ignore",
			shouldDetect: true,
			description:  "Double slash comment should be detected",
		},
		{
			name:         "Valid SQL comment",
			input:        "-- m2e-ignore",
			shouldDetect: true,
			description:  "SQL comment should be detected",
		},
		{
			name:         "Valid shebang",
			input:        "#!/bin/bash m2e-ignore",
			shouldDetect: true,
			description:  "Shebang with ignore should be detected",
		},

		// Invalid comments - should NOT detect
		{
			name:         "Hex color code",
			input:        "color: #ffffff;",
			shouldDetect: false,
			description:  "Hex color codes should not be detected as comments",
		},
		{
			name:         "Short hex color",
			input:        "color: #fff;",
			shouldDetect: false,
			description:  "Short hex color codes should not be detected as comments",
		},
		{
			name:         "Hash with hex-like content",
			input:        "hash = #abc123def",
			shouldDetect: false,
			description:  "Hash with hex-like content should not be detected",
		},
		{
			name:         "Decrement operator",
			input:        "--count;",
			shouldDetect: false,
			description:  "Decrement operators should not be detected as comments",
		},
		{
			name:         "CSS custom property",
			input:        "--main-color: blue;",
			shouldDetect: false,
			description:  "CSS custom properties should not be detected as comments",
		},
		{
			name:         "Modulo operation",
			input:        "result = value % 10;",
			shouldDetect: false,
			description:  "Modulo operations should not be detected as comments",
		},
		{
			name:         "Printf format specifier",
			input:        "printf(\"%d\", number);",
			shouldDetect: false,
			description:  "Printf format specifiers should not be detected as comments",
		},
		{
			name:         "String literal",
			input:        "message = 'hello world';",
			shouldDetect: false,
			description:  "String literals should not be detected as comments",
		},
		{
			name:         "URL with double slash",
			input:        "url = \"https://example.com\";",
			shouldDetect: false,
			description:  "URLs with // should not be detected as comments",
		},
		{
			name:         "Python docstring",
			input:        "\"\"\"This is a docstring with m2e-ignore\"\"\"",
			shouldDetect: true,
			description:  "Python docstrings should be detected as comments",
		},
		{
			name:         "Python single quote docstring",
			input:        "'''This is a docstring with m2e-ignore'''",
			shouldDetect: true,
			description:  "Python single quote docstrings should be detected",
		},

		// Edge cases
		{
			name:         "Hash followed by space and ignore",
			input:        "# This is a comment with m2e-ignore",
			shouldDetect: true,
			description:  "Hash comments with spaces should be detected",
		},
		{
			name:         "SQL comment with ignore",
			input:        "-- SQL comment m2e-ignore",
			shouldDetect: true,
			description:  "SQL comments should be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get ignore directives to see if any were found
			directives := conv.GetIgnoreDirectives(tt.input)

			hasIgnoreDirective := len(directives) > 0

			if hasIgnoreDirective != tt.shouldDetect {
				t.Errorf("%s: Expected shouldDetect=%v, but got %v. Found %d directives in: %q",
					tt.description, tt.shouldDetect, hasIgnoreDirective, len(directives), tt.input)

				if len(directives) > 0 {
					for i, directive := range directives {
						t.Logf("  Directive %d: %s at line %d", i, directive.Directive.String(), directive.LineNumber)
					}
				}
			}
		})
	}
}

func TestIgnoreCommentValidationIntegration(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name: "Code with hex colors should convert normally",
			input: `CSS example with color and flavor:
.button {
	color: #ffffff;
	background-color: #abcdef;
}
The button has a nice color and great flavor.`,
			expected: `CSS example with colour and flavour:
.button {
	colour: #ffffff;
	background-colour: #abcdef;
}
The button has a nice colour and great flavour.`,
			description: "Hex colors should not prevent conversion of surrounding text",
		},
		{
			name: "Code with decrements should convert normally",
			input: `JavaScript with decrements:
--count;
--variable;
The color and flavor should still convert.`,
			expected: `JavaScript with decrements:
--count;
--variable;
The colour and flavour should still convert.`,
			description: "Decrement operators should not prevent conversion",
		},
		{
			name: "Valid comment should prevent conversion",
			input: `This line has color and flavor.
# m2e-ignore-next  
This line also has color and flavor but should be ignored.
This line has color and flavor and should convert.`,
			expected: `This line has colour and flavour.
# m2e-ignore-next  
This line also has color and flavor but should be ignored.
This line has colour and flavour and should convert.`,
			description: "Valid ignore comments should still work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tt.input, true)

			if result != tt.expected {
				t.Errorf("%s\nExpected:\n%s\n\nGot:\n%s", tt.description, tt.expected, result)
			}
		})
	}
}
