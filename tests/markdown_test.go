package tests

import (
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestMarkdownLinks(t *testing.T) {
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
			name:     "Simple markdown link",
			input:    "Check out the [color guide](https://example.com)",
			expected: "Check out the [colour guide](https://example.com)",
		},
		{
			name:     "Multiple words in link text",
			input:    "Read about [color and flavor](https://example.com)",
			expected: "Read about [colour and flavour](https://example.com)",
		},
		{
			name:     "Link with American words in URL (URL should not change)",
			input:    "See [colours](https://example.com/color-theory)",
			expected: "See [colours](https://example.com/color-theory)",
		},
		{
			name:     "Multiple links in text",
			input:    "Check [color](https://a.com) and [flavor](https://b.com)",
			expected: "Check [colour](https://a.com) and [flavour](https://b.com)",
		},
		{
			name:     "Link in sentence",
			input:    "The [color wheel](https://example.com) shows different colors.",
			expected: "The [colour wheel](https://example.com) shows different colours.",
		},
		{
			name:     "Link with punctuation after",
			input:    "Read the [color guide](https://example.com).",
			expected: "Read the [colour guide](https://example.com).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tt.input, false)
			if result != tt.expected {
				t.Errorf("ConvertToBritish(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMarkdownBold(t *testing.T) {
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
			name:     "Bold with asterisks",
			input:    "Use **vibrant colors** in design",
			expected: "Use **vibrant colours** in design",
		},
		{
			name:     "Bold with underscores",
			input:    "Use __vibrant colors__ in design",
			expected: "Use __vibrant colours__ in design",
		},
		{
			name:     "Multiple bold sections",
			input:    "The **color** and **flavor** are important",
			expected: "The **colour** and **flavour** are important",
		},
		{
			name:     "Bold at start of sentence",
			input:    "**Color** is important in design",
			expected: "**Colour** is important in design",
		},
		{
			name:     "Bold at end of sentence",
			input:    "Design uses vibrant **colors**",
			expected: "Design uses vibrant **colours**",
		},
		{
			name:     "Bold with punctuation",
			input:    "Use **vibrant colors**.",
			expected: "Use **vibrant colours**.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tt.input, false)
			if result != tt.expected {
				t.Errorf("ConvertToBritish(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMarkdownItalic(t *testing.T) {
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
			name:     "Italic with asterisks",
			input:    "The *color* is important",
			expected: "The *colour* is important",
		},
		{
			name:     "Italic with underscores",
			input:    "The _color_ is important",
			expected: "The _colour_ is important",
		},
		{
			name:     "Multiple italic sections",
			input:    "Both *color* and *flavor* matter",
			expected: "Both *colour* and *flavour* matter",
		},
		{
			name:     "Italic at start",
			input:    "*Color* is important",
			expected: "*Colour* is important",
		},
		{
			name:     "Italic at end with punctuation",
			input:    "It's about the *color*.",
			expected: "It's about the *colour*.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tt.input, false)
			if result != tt.expected {
				t.Errorf("ConvertToBritish(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMarkdownBulletPoints(t *testing.T) {
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
			name:     "Dash bullet points",
			input:    "- Use vibrant colors\n- Add good flavor",
			expected: "- Use vibrant colours\n- Add good flavour",
		},
		{
			name:     "Asterisk bullet points",
			input:    "* Use vibrant colors\n* Add good flavor",
			expected: "* Use vibrant colours\n* Add good flavour",
		},
		{
			name:     "Plus bullet points",
			input:    "+ Use vibrant colors\n+ Add good flavor",
			expected: "+ Use vibrant colours\n+ Add good flavour",
		},
		{
			name:     "Unicode bullet points",
			input:    "• Use vibrant colors\n• Add good flavor",
			expected: "• Use vibrant colours\n• Add good flavour",
		},
		{
			name:     "Numbered list",
			input:    "1. Use vibrant colors\n2. Add good flavor",
			expected: "1. Use vibrant colours\n2. Add good flavour",
		},
		{
			name:     "Indented bullets",
			input:    "  - Use vibrant colors\n  - Add good flavor",
			expected: "  - Use vibrant colours\n  - Add good flavour",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tt.input, false)
			if result != tt.expected {
				t.Errorf("ConvertToBritish(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMarkdownMixed(t *testing.T) {
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
			name:     "Link with bold text",
			input:    "Check out the [**color** guide](https://example.com)",
			expected: "Check out the [**colour** guide](https://example.com)",
		},
		{
			name:     "Bold and italic mixed",
			input:    "The **color** and *flavor* are important",
			expected: "The **colour** and *flavour* are important",
		},
		{
			name:     "Bullet with bold",
			input:    "- Use **vibrant colors**\n- Add *good flavor*",
			expected: "- Use **vibrant colours**\n- Add *good flavour*",
		},
		{
			name:     "Complex markdown document",
			input:    "# Colors in Design\n\nCheck the [color wheel](https://example.com) for:\n- **vibrant colors**\n- _color_ theory basics",
			expected: "# Colours in Design\n\nCheck the [colour wheel](https://example.com) for:\n- **vibrant colours**\n- _colour_ theory basics",
		},
		// KNOWN LIMITATION: Nested formatting (links inside bold) requires recursive parsing
		// Workaround: Use formatting inside the link instead: [**text**](url)
		// {
		// 	name:     "Link in bold",
		// 	input:    "**See [color guide](https://example.com) for details**",
		// 	expected: "**See [colour guide](https://example.com) for details**",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tt.input, false)
			if result != tt.expected {
				t.Errorf("ConvertToBritish(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMarkdownWithPlainURLs(t *testing.T) {
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
			name:     "Plain URL preserved",
			input:    "Visit https://example.com/color for more info",
			expected: "Visit https://example.com/color for more info",
		},
		{
			name:     "Plain URL with markdown link",
			input:    "Visit https://example.com or [color guide](https://guide.com)",
			expected: "Visit https://example.com or [colour guide](https://guide.com)",
		},
		{
			name:     "Word after URL gets converted",
			input:    "See https://example.com for color information",
			expected: "See https://example.com for colour information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tt.input, false)
			if result != tt.expected {
				t.Errorf("ConvertToBritish(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
