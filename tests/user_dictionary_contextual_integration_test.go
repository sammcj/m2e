package tests

import (
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestUserDictionaryContextualIntegration(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	// Enable contextual word detection
	conv.SetContextualWordDetectionEnabled(true)

	tests := []struct {
		name     string
		input    string
		expected string
		note     string
	}{
		{
			name:     "Contextual conversion takes precedence over dictionary for 'license'",
			input:    "I need a license to drive",
			expected: "I need a licence to drive",
			note:     "Should use contextual conversion (noun->licence) not dictionary conversion",
		},
		{
			name:     "Verb form stays as 'license'",
			input:    "We license our software",
			expected: "We license our software",
			note:     "Should use contextual conversion (verb->license) not dictionary conversion",
		},
		{
			name:     "Regular dictionary words still work",
			input:    "I like the color gray in the center",
			expected: "I like the colour grey in the centre",
			note:     "Non-contextual words should use regular dictionary conversion",
		},
		{
			name:     "Mixed contextual and dictionary words",
			input:    "The license application uses gray color schemes",
			expected: "The licence application uses grey colour schemes",
			note:     "Both contextual (license->licence) and dictionary (gray->grey, color->colour) conversions",
		},
		{
			name:     "Contextual disabled falls back to no conversion for contextual words",
			input:    "I need a license to license software",
			expected: "I need a license to license software", // When disabled, license should remain unchanged
			note:     "When contextual is disabled, contextual words are not converted",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Special handling for the disabled test case
			if test.name == "Contextual disabled falls back to no conversion for contextual words" {
				conv.SetContextualWordDetectionEnabled(false)
				defer conv.SetContextualWordDetectionEnabled(true) // Reset for other tests
			}

			result := conv.ConvertToBritishSimple(test.input, false)
			if result != test.expected {
				t.Errorf("Test %s failed:\nInput: %q\nExpected: %q\nActual: %q\nNote: %s",
					test.name, test.input, test.expected, result, test.note)
			}
		})
	}
}

func TestContextualWordConfigUserCustomization(t *testing.T) {
	config := converter.GetDefaultContextualWordConfig()

	// Test adding custom word
	config.AddCustomWord("practice", "practice", "practise")

	if !config.IsWordSupported("practice") {
		t.Error("Should support 'practice' after adding custom word")
	}

	mapping, exists := config.GetMappingForWord("practice")
	if !exists {
		t.Fatal("Should have mapping for 'practice'")
	}

	if mapping.NounReplacement != "practice" {
		t.Errorf("Expected noun form 'practice', got '%s'", mapping.NounReplacement)
	}

	if mapping.VerbReplacement != "practise" {
		t.Errorf("Expected verb form 'practise', got '%s'", mapping.VerbReplacement)
	}

	// Test removing custom word
	config.RemoveCustomWord("practice")

	if config.IsWordSupported("practice") {
		t.Error("Should not support 'practice' after removal")
	}

	_, exists = config.GetMappingForWord("practice")
	if exists {
		t.Error("Should not have mapping for 'practice' after removal")
	}
}

func TestUserConfigurationTemplate(t *testing.T) {
	template := converter.GetUserConfigurationExample()

	// Should have the default license mapping
	if !template.IsWordSupported("license") {
		t.Error("Template should support 'license'")
	}

	// Should have the example custom words
	if !template.IsWordSupported("practice") {
		t.Error("Template should support 'practice' as example")
	}

	if !template.IsWordSupported("advice") {
		t.Error("Template should support 'advice' as example")
	}

	// Should have example exclusion pattern
	hasCustomPattern := false
	for _, pattern := range template.ExcludePatterns {
		if pattern == `(?i)my\s+custom\s+pattern` {
			hasCustomPattern = true
			break
		}
	}
	if !hasCustomPattern {
		t.Error("Template should have example custom exclusion pattern")
	}

	// Test the practice mapping
	mapping, exists := template.GetMappingForWord("practice")
	if !exists {
		t.Fatal("Template should have practice mapping")
	}

	if mapping.NounReplacement != "practice" || mapping.VerbReplacement != "practise" {
		t.Errorf("Practice mapping should be noun='practice', verb='practise', got noun='%s', verb='%s'",
			mapping.NounReplacement, mapping.VerbReplacement)
	}

	// Test the advice mapping
	mapping, exists = template.GetMappingForWord("advice")
	if !exists {
		t.Fatal("Template should have advice mapping")
	}

	if mapping.NounReplacement != "advice" || mapping.VerbReplacement != "advise" {
		t.Errorf("Advice mapping should be noun='advice', verb='advise', got noun='%s', verb='%s'",
			mapping.NounReplacement, mapping.VerbReplacement)
	}
}

func TestUserDictionaryExampleCreation(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	// The user dictionary should be loaded
	dict := conv.GetAmericanToBritishDictionary()

	// Should have the example entry (this is always present)
	customize, hasCustomize := dict["customize"]
	if !hasCustomize {
		t.Error("User dictionary should contain example 'customize' entry")
	}

	if customize != "customise" {
		t.Errorf("Expected 'customize' -> 'customise', got 'customize' -> '%s'", customize)
	}

	// Check if the note is present (might not be if dictionary already existed)
	note, hasNote := dict["example_note"]
	if hasNote {
		if !strings.Contains(note, "contextual_word_config.json") {
			t.Error("Note should mention the contextual word config file")
		}
		t.Logf("Found note in user dictionary: %s", note)
	} else {
		t.Log("Note not found in user dictionary (dictionary may have already existed)")
	}
}

func TestContextualWordConfigurationWithUserPatterns(t *testing.T) {
	config := converter.GetDefaultContextualWordConfig()

	// Add a custom exclusion pattern
	customPattern := `(?i)test\s+custom\s+license`
	config.AddExclusionPattern(customPattern)

	// Create detector with custom config
	detector := converter.NewContextAwareWordDetectorWithConfig(config)

	// Test text that matches the custom exclusion pattern
	matches := detector.DetectWords("This is a test custom license that should be excluded")

	if len(matches) > 0 {
		t.Errorf("Expected no matches due to custom exclusion pattern, got %d matches", len(matches))
	}

	// Test text that doesn't match the exclusion pattern
	matches = detector.DetectWords("I need a regular license for driving")

	if len(matches) == 0 {
		t.Error("Expected matches for text that doesn't match exclusion pattern")
	}

	// Remove the custom pattern
	config.RemoveExclusionPattern(customPattern)
	detector.SetConfig(config)

	// Should now match
	matches = detector.DetectWords("This is a test custom license that should now be detected")

	if len(matches) == 0 {
		t.Error("Expected matches after removing custom exclusion pattern")
	}
}
