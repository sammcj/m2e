// Package converter provides unit conversion configuration functionality
package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// UnitConfig holds all configuration options for unit conversion
type UnitConfig struct {
	// Global enable/disable flag
	Enabled bool `json:"enabled"`

	// Unit type specific settings
	EnabledUnitTypes []UnitType `json:"enabledUnitTypes"`

	// Precision settings for each unit type
	Precision map[string]int `json:"precision"`

	// Custom unit mappings (American -> British)
	CustomMappings map[string]string `json:"customMappings"`

	// Patterns to exclude from conversion (regex patterns)
	ExcludePatterns []string `json:"excludePatterns"`

	// Conversion preferences
	Preferences ConversionPreferences `json:"preferences"`

	// Detection settings
	Detection DetectionConfig `json:"detection"`
}

// DetectionConfig holds configuration for unit detection
type DetectionConfig struct {
	// Minimum confidence threshold for unit detection (0.0 - 1.0)
	MinConfidence float64 `json:"minConfidence"`

	// Maximum distance between number and unit (in words)
	MaxNumberDistance int `json:"maxNumberDistance"`

	// Whether to detect compound units (e.g., "6-foot")
	DetectCompoundUnits bool `json:"detectCompoundUnits"`

	// Whether to detect written numbers (e.g., "five feet")
	DetectWrittenNumbers bool `json:"detectWrittenNumbers"`
}

// GetDefaultUnitConfig returns the default configuration with sensible defaults
func GetDefaultUnitConfig() *UnitConfig {
	return &UnitConfig{
		Enabled: true,
		EnabledUnitTypes: []UnitType{
			Length,
			Mass,
			Volume,
			Temperature,
			Area,
		},
		Precision: map[string]int{
			"length":      1,
			"mass":        1,
			"volume":      1,
			"temperature": 0,
			"area":        1,
		},
		CustomMappings: make(map[string]string),
		ExcludePatterns: []string{
			// Common idiomatic expressions to exclude
			`miles?\s+(?:away|apart|from\s+home|ahead)`,
			`inch\s+by\s+inch`,
			`every\s+inch`,
			`tons?\s+of\s+(?:fun|work|stuff|things)`,
			`pounds?\s+of\s+(?:pressure|force)\b(?!\s*\d)`, // "pounds of pressure" without numbers
			`cold\s+feet`,
			`foot\s+(?:in\s+the\s+door|the\s+bill)`,
			`pound\s+(?:the\s+pavement|the\s+table)`,
		},
		Preferences: ConversionPreferences{
			PreferWholeNumbers:          true,
			MaxDecimalPlaces:            2,
			UseLocalizedUnits:           true,
			TemperatureFormat:           "°C",
			UseSpaceBetweenValueAndUnit: true,
			RoundingThreshold:           0.1,
		},
		Detection: DetectionConfig{
			MinConfidence:        0.5,
			MaxNumberDistance:    3,
			DetectCompoundUnits:  true,
			DetectWrittenNumbers: true,
		},
	}
}

// ValidateConfig validates the configuration and returns any errors
func ValidateConfig(config *UnitConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate enabled unit types
	validUnitTypes := map[UnitType]bool{
		Length:      true,
		Mass:        true,
		Volume:      true,
		Temperature: true,
		Area:        true,
	}

	for _, unitType := range config.EnabledUnitTypes {
		if !validUnitTypes[unitType] {
			return fmt.Errorf("invalid unit type: %v", unitType)
		}
	}

	// Validate precision values
	for unitTypeStr, precision := range config.Precision {
		if precision < 0 || precision > 10 {
			return fmt.Errorf("precision for %s must be between 0 and 10, got %d", unitTypeStr, precision)
		}
	}

	// Validate detection config
	if config.Detection.MinConfidence < 0.0 || config.Detection.MinConfidence > 1.0 {
		return fmt.Errorf("minConfidence must be between 0.0 and 1.0, got %f", config.Detection.MinConfidence)
	}

	if config.Detection.MaxNumberDistance < 1 || config.Detection.MaxNumberDistance > 10 {
		return fmt.Errorf("maxNumberDistance must be between 1 and 10, got %d", config.Detection.MaxNumberDistance)
	}

	// Validate preferences
	if config.Preferences.MaxDecimalPlaces < 0 || config.Preferences.MaxDecimalPlaces > 10 {
		return fmt.Errorf("maxDecimalPlaces must be between 0 and 10, got %d", config.Preferences.MaxDecimalPlaces)
	}

	if config.Preferences.RoundingThreshold < 0.0 || config.Preferences.RoundingThreshold > 1.0 {
		return fmt.Errorf("roundingThreshold must be between 0.0 and 1.0, got %f", config.Preferences.RoundingThreshold)
	}

	// Validate temperature format
	validTempFormats := map[string]bool{
		"°C":              true,
		"degrees Celsius": true,
		"C":               true,
		"celsius":         true,
	}
	if !validTempFormats[config.Preferences.TemperatureFormat] {
		return fmt.Errorf("invalid temperature format: %s", config.Preferences.TemperatureFormat)
	}

	return nil
}

// IsUnitTypeEnabled checks if a specific unit type is enabled in the configuration
func (c *UnitConfig) IsUnitTypeEnabled(unitType UnitType) bool {
	if !c.Enabled {
		return false
	}

	for _, enabledType := range c.EnabledUnitTypes {
		if enabledType == unitType {
			return true
		}
	}
	return false
}

// GetPrecisionForUnitType returns the precision setting for a unit type
func (c *UnitConfig) GetPrecisionForUnitType(unitType UnitType) int {
	unitTypeStr := c.unitTypeToString(unitType)
	if precision, exists := c.Precision[unitTypeStr]; exists {
		return precision
	}

	// Return default precision if not configured
	defaults := GetDefaultUnitConfig()
	if precision, exists := defaults.Precision[unitTypeStr]; exists {
		return precision
	}

	return 1 // Fallback default
}

// SetPrecisionForUnitType sets the precision for a specific unit type
func (c *UnitConfig) SetPrecisionForUnitType(unitType UnitType, precision int) {
	if c.Precision == nil {
		c.Precision = make(map[string]int)
	}
	c.Precision[c.unitTypeToString(unitType)] = precision
}

// unitTypeToString converts UnitType to string for JSON serialization
func (c *UnitConfig) unitTypeToString(unitType UnitType) string {
	switch unitType {
	case Length:
		return "length"
	case Mass:
		return "mass"
	case Volume:
		return "volume"
	case Temperature:
		return "temperature"
	case Area:
		return "area"
	default:
		return "unknown"
	}
}

// stringToUnitType converts string to UnitType for JSON deserialization
func stringToUnitType(s string) UnitType {
	switch s {
	case "length":
		return Length
	case "mass":
		return Mass
	case "volume":
		return Volume
	case "temperature":
		return Temperature
	case "area":
		return Area
	default:
		return Length // Default fallback
	}
}

// MarshalJSON implements custom JSON marshaling for UnitConfig
func (c *UnitConfig) MarshalJSON() ([]byte, error) {
	// Convert UnitType slice to string slice for JSON
	enabledTypes := make([]string, len(c.EnabledUnitTypes))
	for i, unitType := range c.EnabledUnitTypes {
		enabledTypes[i] = c.unitTypeToString(unitType)
	}

	// Create a temporary struct for JSON marshaling
	temp := struct {
		Enabled          bool                  `json:"enabled"`
		EnabledUnitTypes []string              `json:"enabledUnitTypes"`
		Precision        map[string]int        `json:"precision"`
		CustomMappings   map[string]string     `json:"customMappings"`
		ExcludePatterns  []string              `json:"excludePatterns"`
		Preferences      ConversionPreferences `json:"preferences"`
		Detection        DetectionConfig       `json:"detection"`
	}{
		Enabled:          c.Enabled,
		EnabledUnitTypes: enabledTypes,
		Precision:        c.Precision,
		CustomMappings:   c.CustomMappings,
		ExcludePatterns:  c.ExcludePatterns,
		Preferences:      c.Preferences,
		Detection:        c.Detection,
	}

	return json.Marshal(temp)
}

// UnmarshalJSON implements custom JSON unmarshaling for UnitConfig
func (c *UnitConfig) UnmarshalJSON(data []byte) error {
	// Create a temporary struct for JSON unmarshaling
	temp := struct {
		Enabled          bool                  `json:"enabled"`
		EnabledUnitTypes []string              `json:"enabledUnitTypes"`
		Precision        map[string]int        `json:"precision"`
		CustomMappings   map[string]string     `json:"customMappings"`
		ExcludePatterns  []string              `json:"excludePatterns"`
		Preferences      ConversionPreferences `json:"preferences"`
		Detection        DetectionConfig       `json:"detection"`
	}{}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Convert string slice back to UnitType slice
	enabledTypes := make([]UnitType, len(temp.EnabledUnitTypes))
	for i, typeStr := range temp.EnabledUnitTypes {
		enabledTypes[i] = stringToUnitType(typeStr)
	}

	// Assign values to the config
	c.Enabled = temp.Enabled
	c.EnabledUnitTypes = enabledTypes
	c.Precision = temp.Precision
	c.CustomMappings = temp.CustomMappings
	c.ExcludePatterns = temp.ExcludePatterns
	c.Preferences = temp.Preferences
	c.Detection = temp.Detection

	return nil
}

// Clone creates a deep copy of the configuration
func (c *UnitConfig) Clone() *UnitConfig {
	clone := &UnitConfig{
		Enabled:          c.Enabled,
		EnabledUnitTypes: make([]UnitType, len(c.EnabledUnitTypes)),
		Precision:        make(map[string]int),
		CustomMappings:   make(map[string]string),
		ExcludePatterns:  make([]string, len(c.ExcludePatterns)),
		Preferences:      c.Preferences, // ConversionPreferences is a value type, so this is fine
		Detection:        c.Detection,   // DetectionConfig is a value type, so this is fine
	}

	// Deep copy slices and maps
	copy(clone.EnabledUnitTypes, c.EnabledUnitTypes)
	copy(clone.ExcludePatterns, c.ExcludePatterns)

	for k, v := range c.Precision {
		clone.Precision[k] = v
	}

	for k, v := range c.CustomMappings {
		clone.CustomMappings[k] = v
	}

	return clone
}

// Merge merges another configuration into this one, with the other config taking precedence
func (c *UnitConfig) Merge(other *UnitConfig) {
	if other == nil {
		return
	}

	// Merge simple fields (other takes precedence)
	c.Enabled = other.Enabled

	// Merge enabled unit types (replace entirely)
	if len(other.EnabledUnitTypes) > 0 {
		c.EnabledUnitTypes = make([]UnitType, len(other.EnabledUnitTypes))
		copy(c.EnabledUnitTypes, other.EnabledUnitTypes)
	}

	// Merge precision settings (other overrides)
	if c.Precision == nil {
		c.Precision = make(map[string]int)
	}
	for k, v := range other.Precision {
		c.Precision[k] = v
	}

	// Merge custom mappings (other overrides)
	if c.CustomMappings == nil {
		c.CustomMappings = make(map[string]string)
	}
	for k, v := range other.CustomMappings {
		c.CustomMappings[k] = v
	}

	// Merge exclude patterns (replace entirely)
	if len(other.ExcludePatterns) > 0 {
		c.ExcludePatterns = make([]string, len(other.ExcludePatterns))
		copy(c.ExcludePatterns, other.ExcludePatterns)
	}

	// Merge preferences and detection config (replace entirely)
	c.Preferences = other.Preferences
	c.Detection = other.Detection
}

// GetUserConfigPath returns the path to the user's unit configuration file
func GetUserConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "m2e")
	configPath := filepath.Join(configDir, "unit_config.json")

	return configPath, nil
}

// CreateUserConfigDirectory creates the user configuration directory if it doesn't exist
func CreateUserConfigDirectory() error {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}

// LoadUserConfig loads the user's unit configuration file
// Returns the default configuration if the file doesn't exist
func LoadUserConfig() (*UnitConfig, error) {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config path: %w", err)
	}

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist, return default configuration
		return GetDefaultUnitConfig(), nil
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse the configuration
	var config UnitConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s (please check JSON format): %w", configPath, err)
	}

	// Validate the loaded configuration
	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration in %s: %w", configPath, err)
	}

	return &config, nil
}

// SaveUserConfig saves the configuration to the user's config file
func SaveUserConfig(config *UnitConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate the configuration before saving
	if err := ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	configPath, err := GetUserConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get user config path: %w", err)
	}

	// Create the directory if it doesn't exist
	if err := CreateUserConfigDirectory(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal the configuration to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write the configuration file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configPath, err)
	}

	return nil
}

// CreateExampleUserConfig creates an example user configuration file with comments
func CreateExampleUserConfig() error {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get user config path: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Create the directory if it doesn't exist
	if err := CreateUserConfigDirectory(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create an example configuration with some customizations
	exampleConfig := GetDefaultUnitConfig()

	// Add some example customizations
	exampleConfig.CustomMappings = map[string]string{
		"customize": "customise",
		"color":     "colour",
	}

	// Add an example exclude pattern
	exampleConfig.ExcludePatterns = append(exampleConfig.ExcludePatterns,
		`miles?\s+(?:better|worse)`) // Exclude "miles better/worse"

	// Adjust some preferences as examples
	exampleConfig.Preferences.MaxDecimalPlaces = 1
	exampleConfig.Detection.MinConfidence = 0.6

	// Create JSON with example comments (we'll add them as a separate comment block)
	configJSON, err := json.MarshalIndent(exampleConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal example configuration: %w", err)
	}

	// Create the full content with comments
	content := `{
  "_comment": "M2E Unit Conversion Configuration",
  "_description": "This file controls how imperial units are converted to metric units",
  "_examples": {
    "enabled": "Set to false to disable all unit conversion",
    "enabledUnitTypes": "Array of unit types to convert: length, mass, volume, temperature, area",
    "precision": "Decimal places for each unit type",
    "customMappings": "Custom unit mappings (American -> British)",
    "excludePatterns": "Regex patterns to exclude from conversion (for idiomatic expressions)",
    "preferences": {
      "preferWholeNumbers": "Round to whole numbers when close (e.g., 2.98 -> 3)",
      "maxDecimalPlaces": "Maximum decimal places to show",
      "temperatureFormat": "Format for temperature: '°C' or 'degrees Celsius'",
      "useSpaceBetweenValueAndUnit": "Add space between number and unit: '5 kg' vs '5kg'",
      "roundingThreshold": "How close to whole number before rounding (0.1 = within 10%)"
    },
    "detection": {
      "minConfidence": "Minimum confidence (0.0-1.0) to convert a detected unit",
      "maxNumberDistance": "Maximum words between number and unit (1-10)",
      "detectCompoundUnits": "Detect compound units like '6-foot fence'",
      "detectWrittenNumbers": "Detect written numbers like 'five feet'"
    }
  },
` + string(configJSON)[1:] // Remove the opening brace since we added our own

	// Write the configuration file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write example config file %s: %w", configPath, err)
	}

	return nil
}

// LoadConfigWithDefaults loads user configuration and merges it with defaults
// This is the main function that should be used to get the effective configuration
func LoadConfigWithDefaults() (*UnitConfig, error) {
	// Start with default configuration
	defaultConfig := GetDefaultUnitConfig()

	// Try to load user configuration
	userConfig, err := LoadUserConfig()
	if err != nil {
		// If there's an error loading user config, log it but continue with defaults
		fmt.Fprintf(os.Stderr, "Warning: Failed to load user configuration: %v\n", err)
		fmt.Fprintf(os.Stderr, "Using default configuration. You can create an example config with: m2e --create-unit-config\n")
		return defaultConfig, nil
	}

	// If user config is just the default (file didn't exist), return it as-is
	if userConfig != nil {
		// Merge user configuration with defaults (user config takes precedence)
		defaultConfig.Merge(userConfig)
	}

	return defaultConfig, nil
}

// GetConfigStatus returns information about the current configuration status
func GetConfigStatus() (map[string]interface{}, error) {
	configPath, err := GetUserConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config path: %w", err)
	}

	status := map[string]interface{}{
		"configPath": configPath,
		"exists":     false,
		"valid":      false,
		"error":      nil,
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); err == nil {
		status["exists"] = true

		// Try to load and validate the config
		_, err := LoadUserConfig()
		if err != nil {
			status["error"] = err.Error()
		} else {
			status["valid"] = true
		}
	}

	return status, nil
}
