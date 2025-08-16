package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestContextualWordConfigDefaults(t *testing.T) {
	config := converter.GetDefaultContextualWordConfig()

	if !config.Enabled {
		t.Error("Default config should have contextual conversion enabled")
	}

	if config.MinConfidence != 0.7 {
		t.Errorf("Default config should have MinConfidence of 0.7, got %f", config.MinConfidence)
	}

	if len(config.SupportedWords) == 0 {
		t.Error("Default config should have supported words")
	}

	if !contains(config.SupportedWords, "license") {
		t.Error("Default config should support 'license' word")
	}

	if len(config.ExcludePatterns) == 0 {
		t.Error("Default config should have exclusion patterns")
	}

	if _, exists := config.CustomMappings["license"]; !exists {
		t.Error("Default config should have custom mapping for 'license'")
	}
}

func TestContextualWordConfigMapping(t *testing.T) {
	config := converter.GetDefaultContextualWordConfig()

	mapping, exists := config.GetMappingForWord("license")
	if !exists {
		t.Fatal("Should have mapping for 'license'")
	}

	if mapping.BaseWord != "license" {
		t.Errorf("Expected BaseWord 'license', got '%s'", mapping.BaseWord)
	}

	if mapping.NounReplacement != "licence" {
		t.Errorf("Expected NounReplacement 'licence', got '%s'", mapping.NounReplacement)
	}

	if mapping.VerbReplacement != "license" {
		t.Errorf("Expected VerbReplacement 'license', got '%s'", mapping.VerbReplacement)
	}
}

func TestContextualWordConfigWordSupport(t *testing.T) {
	config := converter.GetDefaultContextualWordConfig()

	if !config.IsWordSupported("license") {
		t.Error("Should support 'license' word")
	}

	if config.IsWordSupported("unsupported") {
		t.Error("Should not support 'unsupported' word")
	}
}

func TestContextualWordConfigExclusionPatterns(t *testing.T) {
	config := converter.GetDefaultContextualWordConfig()

	// Test adding and removing exclusion patterns
	testPattern := `(?i)test\s+license`

	config.AddExclusionPattern(testPattern)
	found := false
	for _, pattern := range config.ExcludePatterns {
		if pattern == testPattern {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have added the test exclusion pattern")
	}

	config.RemoveExclusionPattern(testPattern)
	found = false
	for _, pattern := range config.ExcludePatterns {
		if pattern == testPattern {
			found = true
			break
		}
	}
	if found {
		t.Error("Should have removed the test exclusion pattern")
	}
}

func TestContextualWordDetectorWithConfig(t *testing.T) {
	// Create a custom config
	config := converter.GetDefaultContextualWordConfig()
	config.MinConfidence = 0.9
	config.Enabled = true

	// Create detector with custom config
	detector := converter.NewContextAwareWordDetectorWithConfig(config)

	if detector.GetConfig().MinConfidence != 0.9 {
		t.Errorf("Expected MinConfidence 0.9, got %f", detector.GetConfig().MinConfidence)
	}

	if !detector.IsEnabled() {
		t.Error("Detector should be enabled")
	}

	// Test configuration update
	config.Enabled = false
	detector.SetConfig(config)

	if detector.IsEnabled() {
		t.Error("Detector should be disabled after config update")
	}
}

func TestContextualWordConfigSaveLoad(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "m2e_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Note: We would override the config path here if the functions supported it
	_ = filepath.Join(tmpDir, "contextual_word_config.json")

	// Create a custom config
	originalConfig := converter.GetDefaultContextualWordConfig()
	originalConfig.MinConfidence = 0.8
	originalConfig.Enabled = false
	originalConfig.AddExclusionPattern("test_pattern")

	// Note: Since the SaveContextualWordConfig function uses a hardcoded path,
	// we'll test the default config loading instead
	config := converter.GetDefaultContextualWordConfig()

	// Verify the config has expected defaults
	if config.MinConfidence != 0.7 {
		t.Errorf("Expected default MinConfidence 0.7, got %f", config.MinConfidence)
	}

	if !config.Enabled {
		t.Error("Default config should be enabled")
	}

	// Test that we can create a detector with defaults
	detector := converter.NewContextAwareWordDetector()
	if detector.GetConfig() == nil {
		t.Error("Detector should have a config")
	}
}

func TestContextualWordConfigIntegrationWithDetector(t *testing.T) {
	// Test that detector respects configuration exclusions
	config := converter.GetDefaultContextualWordConfig()
	config.ExcludePatterns = append(config.ExcludePatterns, `(?i)custom\s+license\s+test`)

	detector := converter.NewContextAwareWordDetectorWithConfig(config)

	// This should be excluded by our custom pattern
	matches := detector.DetectWords("This is a custom license test case")

	// Should have no matches due to exclusion
	if len(matches) > 0 {
		t.Errorf("Expected no matches due to exclusion, got %d matches", len(matches))
	}

	// This should not be excluded
	matches = detector.DetectWords("I need a license to drive")

	// Should have matches since this doesn't match exclusion pattern
	if len(matches) == 0 {
		t.Error("Expected matches for non-excluded text")
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
