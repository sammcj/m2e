// Package converter provides dictionary management functionality for loading and managing spelling dictionaries
package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// getUserDictionaryPath returns the path to the user's custom dictionary file
func getUserDictionaryPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "m2e")
	dictPath := filepath.Join(configDir, "american_spellings.json")

	return dictPath, nil
}

// createUserDictionary creates the user dictionary file with an example entry if it doesn't exist
func createUserDictionary(dictPath string) error {
	// Create the directory if it doesn't exist
	configDir := filepath.Dir(dictPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	// Check if file already exists
	if _, err := os.Stat(dictPath); err == nil {
		return nil // File already exists
	}

	// Create the file with example entries and a note
	exampleDict := map[string]string{
		"customize":    "customise",
		"example_note": "For context-aware conversions like license/licence based on noun vs verb usage, see ~/.config/m2e/contextual_word_config.json",
	}

	data, err := json.MarshalIndent(exampleDict, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal example dictionary: %w", err)
	}

	if err := os.WriteFile(dictPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create user dictionary file %s: %w", dictPath, err)
	}

	return nil
}

// loadUserDictionary loads the user's custom dictionary if it exists
func loadUserDictionary() (map[string]string, error) {
	dictPath, err := getUserDictionaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get user dictionary path: %w", err)
	}

	// Try to create the user dictionary if it doesn't exist
	if err := createUserDictionary(dictPath); err != nil {
		return nil, fmt.Errorf("failed to create user dictionary: %w", err)
	}

	// Read the user dictionary file
	data, err := os.ReadFile(dictPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty dictionary
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to read user dictionary file %s: %w", dictPath, err)
	}

	// Parse the user dictionary
	userDict := make(map[string]string)
	if err := json.Unmarshal(data, &userDict); err != nil {
		return nil, fmt.Errorf("failed to parse user dictionary file %s (please check JSON format): %w", dictPath, err)
	}

	return userDict, nil
}

// LoadDictionaries loads the American to British spelling dictionary from the embedded JSON file
// and merges it with the user's custom dictionary
func LoadDictionaries() (*Dictionaries, error) {
	// Load built-in American to British dictionary
	amToBrData, err := dictFS.ReadFile("data/american_spellings.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read built-in American spellings dictionary: %w", err)
	}

	// Parse the built-in dictionary
	amToBr := make(map[string]string)
	if err := json.Unmarshal(amToBrData, &amToBr); err != nil {
		return nil, fmt.Errorf("failed to parse built-in American spellings dictionary: %w", err)
	}

	// Load user dictionary
	userDict, err := loadUserDictionary()
	if err != nil {
		// Log the error but don't fail completely - just use the built-in dictionary
		fmt.Fprintf(os.Stderr, "Warning: Failed to load user dictionary: %v\n", err)
		userDict = make(map[string]string)
	}

	// Merge user dictionary into built-in dictionary (user entries override built-in ones)
	for american, british := range userDict {
		amToBr[american] = british
	}

	return &Dictionaries{
		AmericanToBritish: amToBr,
	}, nil
}
