package converter

import (
	"testing"
)

func TestConvertToBritish(t *testing.T) {
	converter, err := NewConverter()
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertToBritish(tt.input, false)
			if result != tt.expected {
				t.Errorf("ConvertToBritish(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormaliseSmartQuotes(t *testing.T) {
	converter, err := NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Smart double quotes",
			input:    "He said \u201Chello\u201D to me",
			expected: "He said \"hello\" to me",
		},
		{
			name:     "Smart single quotes",
			input:    "It's \u2018quoted\u2019 text",
			expected: "It's 'quoted' text",
		},
		{
			name:     "Em-dash",
			input:    "This is a sentence\u2014with an em-dash",
			expected: "This is a sentence-with an em-dash",
		},
		{
			name:     "En-dash",
			input:    "Pages 125\u2013130",
			expected: "Pages 125-130",
		},
		{
			name:     "Mixed quotes and dashes",
			input:    "\u201CHello\u201D\u2014he said\u2014\u2018how are you?\u2019",
			expected: "\"Hello\"-he said-'how are you?'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.normaliseSmartQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("normaliseSmartQuotes(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertToAmerican(t *testing.T) {
	converter, err := NewConverter()
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
			input:    "colour",
			expected: "color",
		},
		{
			name:     "Capitalized word",
			input:    "Colour",
			expected: "Color",
		},
		{
			name:     "Word with punctuation",
			input:    "colour.",
			expected: "color.",
		},
		{
			name:     "Multiple words",
			input:    "The colour of the centre is grey",
			expected: "The color of the center is gray",
		},
		{
			name:     "Sentence with mixed words",
			input:    "I favour the colour grey for the centre of my flat.",
			expected: "I favor the color gray for the center of my flat.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertToAmerican(tt.input, false)
			if result != tt.expected {
				t.Errorf("ConvertToAmerican(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
