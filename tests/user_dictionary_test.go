package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestUserDictionary(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "m2e_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Set HOME to our temp directory for this test
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	t.Run("Creates user dictionary with example", func(t *testing.T) {
		// Create a new converter (this should create the user dictionary)
		conv, err := converter.NewConverter()
		if err != nil {
			t.Fatalf("Failed to create converter: %v", err)
		}

		// Check that the user dictionary file was created
		configDir := filepath.Join(tempDir, ".config", "m2e")
		dictPath := filepath.Join(configDir, "american_spellings.json")

		if _, err := os.Stat(dictPath); os.IsNotExist(err) {
			t.Error("User dictionary file was not created")
		}

		// Check that it contains the example entry
		data, err := os.ReadFile(dictPath)
		if err != nil {
			t.Fatalf("Failed to read user dictionary: %v", err)
		}

		var userDict map[string]string
		if err := json.Unmarshal(data, &userDict); err != nil {
			t.Fatalf("Failed to parse user dictionary: %v", err)
		}

		if userDict["customize"] != "customise" {
			t.Error("User dictionary does not contain expected example entry")
		}

		// Test that the conversion works
		result := conv.ConvertToBritish("I need to customize this", false)
		expected := "I need to customise this"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("Merges user dictionary with built-in dictionary", func(t *testing.T) {
		// Create a custom user dictionary
		configDir := filepath.Join(tempDir, ".config", "m2e")
		dictPath := filepath.Join(configDir, "american_spellings.json")

		customDict := map[string]string{
			"testword":  "testword-british",
			"customize": "customise",     // Override built-in
			"color":     "colour-custom", // Override built-in with custom value
		}

		data, err := json.MarshalIndent(customDict, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal custom dictionary: %v", err)
		}

		if err := os.WriteFile(dictPath, data, 0644); err != nil {
			t.Fatalf("Failed to write custom dictionary: %v", err)
		}

		// Create a new converter
		conv, err := converter.NewConverter()
		if err != nil {
			t.Fatalf("Failed to create converter: %v", err)
		}

		// Test custom word
		result := conv.ConvertToBritish("This is a testword", false)
		expected := "This is a testword-british"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}

		// Test overridden built-in word
		result = conv.ConvertToBritish("I like this color", false)
		expected = "I like this colour-custom"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}

		// Test that built-in words still work
		result = conv.ConvertToBritish("I need to organize this", false)
		expected = "I need to organise this"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("Handles invalid JSON gracefully", func(t *testing.T) {
		// Create an invalid JSON file
		configDir := filepath.Join(tempDir, ".config", "m2e")
		dictPath := filepath.Join(configDir, "american_spellings.json")

		invalidJSON := `{"invalid": json}`
		if err := os.WriteFile(dictPath, []byte(invalidJSON), 0644); err != nil {
			t.Fatalf("Failed to write invalid JSON: %v", err)
		}

		// Create a new converter (should handle the error gracefully)
		conv, err := converter.NewConverter()
		if err != nil {
			t.Fatalf("Converter creation should not fail with invalid user dictionary: %v", err)
		}

		// Should still work with built-in dictionary
		result := conv.ConvertToBritish("I need to organize this", false)
		expected := "I need to organise this"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("Handles missing config directory", func(t *testing.T) {
		// Remove the entire config directory
		configDir := filepath.Join(tempDir, ".config")
		_ = os.RemoveAll(configDir)

		// Create a new converter (should create the directory and file)
		conv, err := converter.NewConverter()
		if err != nil {
			t.Fatalf("Failed to create converter: %v", err)
		}

		// Check that the directory and file were created
		dictPath := filepath.Join(configDir, "m2e", "american_spellings.json")
		if _, err := os.Stat(dictPath); os.IsNotExist(err) {
			t.Error("User dictionary file was not created when config directory was missing")
		}

		// Should work normally
		result := conv.ConvertToBritish("I need to customize this", false)
		expected := "I need to customise this"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}
