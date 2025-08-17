package tests

import (
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestContextualWordConversion(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	// Ensure contextual word detection is enabled
	conv.SetContextualWordDetectionEnabled(true)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// NOUN PATTERNS - should convert to "licence"
		{
			name:     "Determiner + license (noun)",
			input:    "I need a license to drive",
			expected: "I need a licence to drive",
		},
		{
			name:     "Definite article + license (noun)",
			input:    "The license is valid for 5 years",
			expected: "The licence is valid for 5 years",
		},
		{
			name:     "Possessive + license (noun)",
			input:    "My license expires next month",
			expected: "My licence expires next month",
		},
		{
			name:     "License + noun compound",
			input:    "The license holder must renew annually",
			expected: "The licence holder must renew annually",
		},
		{
			name:     "License number (compound noun)",
			input:    "Please provide your license number",
			expected: "Please provide your licence number",
		},
		{
			name:     "License renewal (compound noun)",
			input:    "The license renewal fee is £50",
			expected: "The licence renewal fee is £50",
		},
		{
			name:     "Preposition + license (object)",
			input:    "He was arrested for driving without a license",
			expected: "He was arrested for driving without a licence",
		},
		{
			name:     "Preposition with license (object)",
			input:    "She came with her license in hand",
			expected: "She came with her licence in hand",
		},
		{
			name:     "Possessive license's",
			input:    "The license's terms are strict",
			expected: "The licence's terms are strict",
		},
		{
			name:     "License at sentence end",
			input:    "She forgot her driving license.",
			expected: "She forgot her driving licence.",
		},

		// VERB PATTERNS - should keep "license"
		{
			name:     "Infinitive to license (verb)",
			input:    "The company plans to license the technology",
			expected: "The company plans to license the technology",
		},
		{
			name:     "Modal + license (verb)",
			input:    "We will license our software to partners",
			expected: "We will license our software to partners",
		},
		{
			name:     "Subject pronoun + license (verb)",
			input:    "They license their products globally",
			expected: "They license their products globally",
		},
		{
			name:     "License + direct object (verb)",
			input:    "The university will license the technology to startups",
			expected: "The university will license the technology to startups",
		},
		{
			name:     "License software (verb)",
			input:    "Companies license software to reduce costs",
			expected: "Companies license software to reduce costs",
		},
		{
			name:     "License content (verb)",
			input:    "We license content from various providers",
			expected: "We license content from various providers",
		},
		{
			name:     "Modal can license (verb)",
			input:    "You can license this under MIT terms",
			expected: "You can license this under MIT terms",
		},
		{
			name:     "Modal should license (verb)",
			input:    "Companies should license their intellectual property",
			expected: "Companies should license their intellectual property",
		},

		// INFLECTED FORMS - now handled by dictionary (licenced/licences/licencing)
		{
			name:     "Licensed (past tense)",
			input:    "The software is licensed under GPL",
			expected: "The software is licenced under GPL",
		},
		{
			name:     "Licenses (present tense)",
			input:    "The company licenses its technology widely",
			expected: "The company licences its technology widely",
		},
		{
			name:     "Licensing (present participle)",
			input:    "They are licensing their patents to competitors",
			expected: "They are licencing their patents to competitors",
		},

		// MIXED CONTEXTS
		{
			name:     "Mixed noun and verb in same sentence",
			input:    "The license allows us to license the software to others",
			expected: "The licence allows us to license the software to others",
		},
		{
			name:     "Multiple instances of different types",
			input:    "His driving license expired, so he cannot license vehicles to customers",
			expected: "His driving licence expired, so he cannot license vehicles to customers",
		},
		{
			name:     "Complex sentence with multiple license types",
			input:    "To license software, you need a valid license from the authority",
			expected: "To license software, you need a valid licence from the authority",
		},

		// CASE PRESERVATION
		{
			name:     "Capitalized noun",
			input:    "The License was issued last week",
			expected: "The Licence was issued last week",
		},
		{
			name:     "ALL CAPS noun",
			input:    "SHOW YOUR LICENSE",
			expected: "SHOW YOUR LICENCE",
		},
		{
			name:     "Capitalized verb",
			input:    "We License our technology globally",
			expected: "We License our technology globally",
		},

		// PUNCTUATION HANDLING
		{
			name:     "License with comma",
			input:    "The license, which expires soon, needs renewal",
			expected: "The licence, which expires soon, needs renewal",
		},
		{
			name:     "License with period",
			input:    "I lost my driving license.",
			expected: "I lost my driving licence.",
		},
		{
			name:     "License with quotes",
			input:    "The 'license' document is here",
			expected: "The 'licence' document is here",
		},

		// EDGE CASES AND COMPLEX CONTEXTS
		{
			name:     "License in business context (noun)",
			input:    "The business license fee increased this year",
			expected: "The business licence fee increased this year",
		},
		{
			name:     "Professional license (noun)",
			input:    "Her medical license allows her to practice",
			expected: "Her medical licence allows her to practise",
		},
		{
			name:     "License agreement context",
			input:    "Read the license agreement before proceeding",
			expected: "Read the licence agreement before proceeding",
		},
		{
			name:     "Multiple adjectives + license (noun)",
			input:    "The temporary professional license is valid",
			expected: "The temporary professional licence is valid",
		},
		{
			name:     "Complex verb phrase",
			input:    "The company decided to license their technology",
			expected: "The company decided to license their technology",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := conv.ConvertToBritishSimple(test.input, false)
			if result != test.expected {
				t.Errorf("ConvertToBritishSimple(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

func TestContextualWordDetectionDisabled(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	// Disable contextual word detection
	conv.SetContextualWordDetectionEnabled(false)

	tests := []struct {
		name     string
		input    string
		expected string // Should remain unchanged when contextual detection is disabled
	}{
		{
			name:     "Noun license (disabled)",
			input:    "I need a license to drive",
			expected: "I need a license to drive", // No conversion when disabled
		},
		{
			name:     "Verb license (disabled)",
			input:    "We license software to partners",
			expected: "We license software to partners", // No conversion when disabled
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := conv.ConvertToBritishSimple(test.input, false)
			if result != test.expected {
				t.Errorf("ConvertToBritishSimple(%q) with disabled contextual detection = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

func TestContextualWordDetector(t *testing.T) {
	detector := converter.NewContextAwareWordDetector()

	tests := []struct {
		name             string
		input            string
		expectedMatches  int
		expectedWordType converter.WordType
	}{
		{
			name:             "Detect noun license",
			input:            "I need a license to drive",
			expectedMatches:  1,
			expectedWordType: converter.Noun,
		},
		{
			name:             "Detect verb license",
			input:            "We will license our software",
			expectedMatches:  1,
			expectedWordType: converter.Verb,
		},
		{
			name:            "Multiple license instances",
			input:           "The license allows us to license software",
			expectedMatches: 2, // Should detect both noun and verb
		},
		{
			name:            "No license instances",
			input:           "This text has no problematic words",
			expectedMatches: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matches := detector.DetectWords(test.input)
			if len(matches) != test.expectedMatches {
				t.Errorf("DetectWords(%q) found %d matches, want %d", test.input, len(matches), test.expectedMatches)
			}

			if len(matches) > 0 && test.expectedMatches == 1 {
				if matches[0].WordType != test.expectedWordType {
					t.Errorf("DetectWords(%q) found word type %v, want %v", test.input, matches[0].WordType, test.expectedWordType)
				}
			}
		})
	}
}

func TestContextualWordPatternConfidence(t *testing.T) {
	detector := converter.NewContextAwareWordDetector()

	tests := []struct {
		name                  string
		input                 string
		expectedMinConfidence float64
	}{
		{
			name:                  "High confidence noun (determiner)",
			input:                 "The license is valid",
			expectedMinConfidence: 0.8,
		},
		{
			name:                  "High confidence verb (infinitive)",
			input:                 "We plan to license software",
			expectedMinConfidence: 0.8,
		},
		{
			name:                  "Medium confidence (ambiguous context)",
			input:                 "License terms apply",
			expectedMinConfidence: 0.5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matches := detector.DetectWords(test.input)
			if len(matches) == 0 {
				t.Errorf("DetectWords(%q) found no matches, expected at least one", test.input)
				return
			}

			confidence := matches[0].Confidence
			if confidence < test.expectedMinConfidence {
				t.Errorf("DetectWords(%q) confidence %.2f is below expected minimum %.2f", test.input, confidence, test.expectedMinConfidence)
			}
		})
	}
}

func TestContextualWordExclusions(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	conv.SetContextualWordDetectionEnabled(true)

	tests := []struct {
		name     string
		input    string
		expected string // Should remain unchanged due to exclusion patterns
	}{
		{
			name:     "MIT license (excluded)",
			input:    "This is under MIT license",
			expected: "This is under MIT license",
		},
		{
			name:     "Software license agreement (excluded)",
			input:    "Read the software license agreement",
			expected: "Read the software license agreement",
		},
		{
			name:     "License file (excluded)",
			input:    "Check the LICENSE.txt file",
			expected: "Check the LICENSE.txt file",
		},
		{
			name:     "License plate (excluded)",
			input:    "The license plate was stolen",
			expected: "The license plate was stolen",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := conv.ConvertToBritishSimple(test.input, false)
			if result != test.expected {
				t.Errorf("ConvertToBritishSimple(%q) = %q, want %q (should be excluded)", test.input, result, test.expected)
			}
		})
	}
}

func TestContextualWordIntegrationWithRegularDictionary(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	conv.SetContextualWordDetectionEnabled(true)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Contextual license + regular dictionary words",
			input:    "The license color is gray and the center is beautiful",
			expected: "The licence colour is grey and the centre is beautiful",
		},
		{
			name:     "Verb license + regular dictionary words",
			input:    "We license software with gray color schemes",
			expected: "We license software with grey colour schemes",
		},
		{
			name:     "Mixed contextual and regular conversions",
			input:    "To license this software, check the license in the center of the gray folder",
			expected: "To license this software, check the licence in the centre of the grey folder",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := conv.ConvertToBritishSimple(test.input, false)
			if result != test.expected {
				t.Errorf("ConvertToBritishSimple(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}
