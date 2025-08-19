// Package converter provides contextual word conversion configuration functionality
package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ContextualWordConfig holds all configuration options for contextual word conversion
type ContextualWordConfig struct {
	// Global enable/disable flag
	Enabled bool `json:"enabled"`

	// Word configurations by base word
	WordConfigs map[string]WordConfig `json:"wordConfigs"`

	// Minimum confidence threshold for contextual detection (0.0 - 1.0)
	MinConfidence float64 `json:"minConfidence"`

	// Custom exclusion patterns (regex patterns to avoid conversion)
	ExcludePatterns []string `json:"excludePatterns"`

	// Conversion preferences
	Preferences ContextualWordPreferences `json:"preferences"`

	// Backward compatibility fields
	SupportedWords []string                     `json:"-"` // Populated dynamically
	CustomMappings map[string]ContextualMapping `json:"-"` // Populated dynamically
}

// ContextualMapping represents a word that has different spellings based on context
type ContextualMapping struct {
	BaseWord        string             `json:"baseWord"`        // The base American word (e.g., "license")
	NounReplacement string             `json:"nounReplacement"` // British spelling when used as noun (e.g., "licence")
	VerbReplacement string             `json:"verbReplacement"` // British spelling when used as verb (e.g., "license")
	Confidence      map[string]float64 `json:"confidence"`      // Confidence overrides for different contexts
}

// ContextualWordPreferences holds user preferences for contextual word conversion
type ContextualWordPreferences struct {
	// Whether to prefer noun conversion when context is ambiguous
	PreferNounOnAmbiguity bool `json:"preferNounOnAmbiguity"`

	// Whether to fall back to regular dictionary when contextual conversion fails
	FallbackToDictionary bool `json:"fallbackToDictionary"`

	// Whether to show warnings for ambiguous contexts
	ShowAmbiguityWarnings bool `json:"showAmbiguityWarnings"`

	// Case sensitivity for pattern matching
	CaseSensitive bool `json:"caseSensitive"`

	// Whether to convert within quoted strings
	ConvertQuotedText bool `json:"convertQuotedText"`
}

// GetDefaultContextualWordConfig returns the default configuration with sensible defaults
func GetDefaultContextualWordConfig() *ContextualWordConfig {
	config := &ContextualWordConfig{
		Enabled: true,
		WordConfigs: map[string]WordConfig{
			"license": {
				Noun:    "licence",
				Verb:    "license",
				Enabled: true,
			},
			"practice": {
				Noun:    "practice",
				Verb:    "practise",
				Enabled: true,
			},
			"advice": {
				Noun:    "advice",
				Verb:    "advise",
				Enabled: true,
			},
		},
		MinConfidence: 0.7,
		ExcludePatterns: []string{
			// Software license names
			`(?i)(?:MIT|BSD|GPL|Apache|Creative\s+Commons|GNU|Mozilla)\s+license`,
			`(?i)software\s+license\s+(?:agreement|terms)`,

			// License filenames
			`(?i)LICENSE\s*\.(?:txt|md|doc|pdf|html)`,
			`(?i)the\s+LICENSE\s*\.(?:txt|md|doc|pdf|html)\s+file`,

			// URLs and file paths
			`(?i)(?:https?://|www\.)\S*license\S*`,
			`(?i)(?:/|\\)\S*license\S*(?:/|\\|\.)`,

			// Code contexts
			`(?i)(?:var|const|let|def|function|class|interface|struct|type)\s+\w*\b(?:license|practice|advice)\w*`,
			`(?i)\w*\b(?:license|practice|advice)\w*\s*(?:=|:=|==|!=|<|>|\+|\-|\*|/)`,

			// Quoted strings in code contexts
			`(?i)(?:=|:)\s*["']\s*\w*\b(?:license|practice|advice)\w*\s*["']`,
			`(?i)["']\s*\w*\b(?:license|practice|advice)\w*\s*["']\s*(?:=|:|\)|;|,)`,

			// Dialog in code contexts (HTML, JavaScript, CSS, etc.)
			`(?i)<dialog\b`,           // HTML dialog element
			`(?i)</dialog>`,           // HTML dialog closing tag
			`(?i)\bdialog\s*\.\s*\w+`, // dialog.method() calls
			`(?i)\w*\.dialog\b`,       // object.dialog properties
			`(?i)\.dialog\b`,          // CSS .dialog classes
			`(?i)#dialog\b`,           // CSS #dialog IDs
			`(?i)data-dialog`,         // data-dialog attributes
			`(?i)dialog-\w+`,          // dialog-* attributes/classes
			`(?i)\b(?:show|open|close|hide|modal)Dialog\b`,                                    // showDialog, openDialog functions
			`(?i)\bdialog(?:Box|Modal|Window|Panel)\b`,                                        // dialogBox, dialogModal compound words
			`(?i)(?:var|const|let|def|function|class|interface|struct|type)\s+\w*\bdialog\w*`, // variable/function names
			`(?i)\w*\bdialog\w*\s*(?:=|:=|==|!=|<|>|\+|\-|\*|/)`,                              // dialog in expressions
			`(?i)(?:=|:)\s*["']\s*\w*\bdialog\w*\s*["']`,                                      // quoted dialog strings in code
			`(?i)["']\s*\w*\bdialog\w*\s*["']\s*(?:=|:|\)|;|,)`,                               // dialog in quoted assignments
		},
		Preferences: ContextualWordPreferences{
			PreferNounOnAmbiguity: true,  // Default to noun when uncertain
			FallbackToDictionary:  false, // Don't use dictionary for contextual words
			ShowAmbiguityWarnings: false,
			CaseSensitive:         false,
			ConvertQuotedText:     false, // Skip quoted text by default
		},
	}

	// Populate backward compatibility fields
	config.populateBackwardCompatibilityFields()

	return config
}

// getContextualWordConfigPath returns the path to the contextual word configuration file
func getContextualWordConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "m2e")
	configPath := filepath.Join(configDir, "contextual_word_config.json")

	return configPath, nil
}

// createDefaultContextualWordConfig creates the default configuration file if it doesn't exist
func createDefaultContextualWordConfig(configPath string) error {
	// Create the directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // File already exists
	}

	// Create the file with default configuration
	defaultConfig := GetDefaultContextualWordConfig()

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default contextual word configuration: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create contextual word configuration file %s: %w", configPath, err)
	}

	return nil
}

// LoadContextualWordConfig loads the contextual word configuration from file
func LoadContextualWordConfig() (*ContextualWordConfig, error) {
	configPath, err := getContextualWordConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get contextual word config path: %w", err)
	}

	// Try to create the default config if it doesn't exist
	if err := createDefaultContextualWordConfig(configPath); err != nil {
		return nil, fmt.Errorf("failed to create default contextual word config: %w", err)
	}

	// Read the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return default configuration
			return GetDefaultContextualWordConfig(), nil
		}
		return nil, fmt.Errorf("failed to read contextual word configuration file %s: %w", configPath, err)
	}

	// Parse the configuration
	config := &ContextualWordConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse contextual word configuration file %s (please check JSON format): %w", configPath, err)
	}

	// Validate and apply defaults for missing values
	if config.MinConfidence <= 0 || config.MinConfidence > 1 {
		config.MinConfidence = 0.7
	}

	if config.WordConfigs == nil {
		config.WordConfigs = GetDefaultContextualWordConfig().WordConfigs
	}

	// Populate backward compatibility fields
	config.populateBackwardCompatibilityFields()

	return config, nil
}

// LoadContextualWordConfigWithDefaults loads configuration with fallback to defaults
func LoadContextualWordConfigWithDefaults() (*ContextualWordConfig, error) {
	config, err := LoadContextualWordConfig()
	if err != nil {
		// Log error but return default config
		fmt.Fprintf(os.Stderr, "Warning: Failed to load contextual word configuration: %v\n", err)
		return GetDefaultContextualWordConfig(), nil
	}
	return config, nil
}

// SaveContextualWordConfig saves the configuration to file
func SaveContextualWordConfig(config *ContextualWordConfig) error {
	configPath, err := getContextualWordConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get contextual word config path: %w", err)
	}

	// Create the directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	// Marshal the configuration
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal contextual word configuration: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write contextual word configuration file %s: %w", configPath, err)
	}

	return nil
}

// IsWordSupported checks if a word is enabled for contextual conversion
func (c *ContextualWordConfig) IsWordSupported(word string) bool {
	config, exists := c.WordConfigs[word]
	return exists && config.Enabled
}

// GetWordConfig returns the configuration for a specific word
func (c *ContextualWordConfig) GetWordConfig(word string) (WordConfig, bool) {
	config, exists := c.WordConfigs[word]
	return config, exists && config.Enabled
}

// AddExclusionPattern adds a new exclusion pattern to the configuration
func (c *ContextualWordConfig) AddExclusionPattern(pattern string) {
	c.ExcludePatterns = append(c.ExcludePatterns, pattern)
}

// RemoveExclusionPattern removes an exclusion pattern from the configuration
func (c *ContextualWordConfig) RemoveExclusionPattern(pattern string) {
	for i, existing := range c.ExcludePatterns {
		if existing == pattern {
			c.ExcludePatterns = append(c.ExcludePatterns[:i], c.ExcludePatterns[i+1:]...)
			break
		}
	}
	// Update compatibility fields after modification
	c.populateBackwardCompatibilityFields()
}

// AddCustomWord adds a new word with contextual mappings to the configuration
func (c *ContextualWordConfig) AddCustomWord(baseWord, nounForm, verbForm string) {
	if c.WordConfigs == nil {
		c.WordConfigs = make(map[string]WordConfig)
	}

	c.WordConfigs[baseWord] = WordConfig{
		Noun:    nounForm,
		Verb:    verbForm,
		Enabled: true,
	}
}

// RemoveCustomWord removes a word from contextual conversion
func (c *ContextualWordConfig) RemoveCustomWord(baseWord string) {
	if c.WordConfigs != nil {
		delete(c.WordConfigs, baseWord)
	}
}

// EnableWord enables contextual conversion for a specific word
func (c *ContextualWordConfig) EnableWord(baseWord string) {
	if config, exists := c.WordConfigs[baseWord]; exists {
		config.Enabled = true
		c.WordConfigs[baseWord] = config
	}
}

// DisableWord disables contextual conversion for a specific word
func (c *ContextualWordConfig) DisableWord(baseWord string) {
	if config, exists := c.WordConfigs[baseWord]; exists {
		config.Enabled = false
		c.WordConfigs[baseWord] = config
	}
}

// GetSupportedWords returns a list of all enabled words for contextual conversion
func (c *ContextualWordConfig) GetSupportedWords() []string {
	var supportedWords []string
	for word, config := range c.WordConfigs {
		if config.Enabled {
			supportedWords = append(supportedWords, word)
		}
	}
	return supportedWords
}

// GetMappingForWord returns the contextual mapping for a specific word in old format
func (c *ContextualWordConfig) GetMappingForWord(word string) (ContextualMapping, bool) {
	config, exists := c.WordConfigs[word]
	if !exists || !config.Enabled {
		return ContextualMapping{}, false
	}

	mapping := ContextualMapping{
		BaseWord:        word,
		NounReplacement: config.Noun,
		VerbReplacement: config.Verb,
		Confidence: map[string]float64{
			"noun": 0.9,
			"verb": 0.9,
		},
	}
	return mapping, true
}

// populateBackwardCompatibilityFields populates the compatibility fields from WordConfigs
func (c *ContextualWordConfig) populateBackwardCompatibilityFields() {
	// Populate SupportedWords
	c.SupportedWords = c.GetSupportedWords()

	// Populate CustomMappings
	c.CustomMappings = make(map[string]ContextualMapping)
	for word, config := range c.WordConfigs {
		if config.Enabled {
			c.CustomMappings[word] = ContextualMapping{
				BaseWord:        word,
				NounReplacement: config.Noun,
				VerbReplacement: config.Verb,
				Confidence: map[string]float64{
					"noun": 0.9,
					"verb": 0.9,
				},
			}
		}
	}
}

// GetUserConfigurationExample returns an example configuration for users
func GetUserConfigurationExample() *ContextualWordConfig {
	config := GetDefaultContextualWordConfig()

	// Add example custom exclusion
	config.AddExclusionPattern(`(?i)my\s+custom\s+pattern`)

	// Repopulate compatibility fields
	config.populateBackwardCompatibilityFields()

	return config
}

// CreateUserConfigurationTemplate creates a template configuration file with examples
func CreateUserConfigurationTemplate() error {
	configPath, err := getContextualWordConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get contextual word config path: %w", err)
	}

	// Don't overwrite existing config
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", configPath)
	}

	// Create template with examples
	template := GetUserConfigurationExample()

	return SaveContextualWordConfig(template)
}
