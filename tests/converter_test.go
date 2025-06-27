package tests

import (
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestConvertToBritish(t *testing.T) {
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
			name:     "Simple word",
			input:    "color",
			expected: "colour",
		},
		{
			name:     "Capitalized word",
			input:    "Color",
			expected: "Colour",
		},
		{
			name:     "Word with punctuation",
			input:    "color.",
			expected: "colour.",
		},
		{
			name:     "Multiple words",
			input:    "The color of the center is gray",
			expected: "The colour of the centre is grey",
		},
		{
			name:     "Sentence with mixed words",
			input:    "I favor the color gray for the center of my apartment.",
			expected: "I favour the colour grey for the centre of my apartment.",
		},
		{
			name:     "Text with newlines",
			input:    "The color is gray.\nThe center is beautiful.",
			expected: "The colour is grey.\nThe centre is beautiful.",
		},
		{
			name:     "Text with multiple newlines",
			input:    "The color is gray.\n\nThe center is beautiful.",
			expected: "The colour is grey.\n\nThe centre is beautiful.",
		},
		{
			name:     "Possessive word",
			input:    "The organization's color is gray.",
			expected: "The organisation's colour is grey.",
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

func TestNormaliseSmartQuotes(t *testing.T) {
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
			name:     "Smart double quotes with American word",
			input:    "He said \u201chello color\u201d to me",
			expected: "He said \"hello colour\" to me",
		},
		{
			name:     "Smart single quotes with American word",
			input:    "It\u2019s \u2018color\u2019 quoted text",
			expected: "It's 'colour' quoted text",
		},
		{
			name:     "Em-dash with American word",
			input:    "This is a sentence\u2014with color\u2014an em-dash",
			expected: "This is a sentence-with colour-an em-dash",
		},
		{
			name:     "En-dash with American word",
			input:    "Pages 125\u2013130 about color",
			expected: "Pages 125-130 about colour",
		},
		{
			name:     "Mixed quotes and dashes with American words",
			input:    "\u201cHello color\u201d\u2014he said\u2014\u2018how are you color?\u2019",
			expected: "\"Hello colour\"-he said-'how are you colour?'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that smart quotes are normalized AND American words are converted
			result := conv.ConvertToBritish(tt.input, true)
			if result != tt.expected {
				t.Errorf("ConvertToBritish(%q, true) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
