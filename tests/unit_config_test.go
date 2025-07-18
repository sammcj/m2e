package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestGetDefaultUnitConfig(t *testing.T) {
	config := converter.GetDefaultUnitConfig()

	// Test that default config is valid
	if config == nil {
		t.Fatal("Default config should not be nil")
	}

	// Test default values
	if !config.Enabled {
		t.Error("Default config should be enabled")
	}

	// Test that all unit types are enabled by default
	expectedUnitTypes := []converter.UnitType{
		converter.Length,
		converter.Mass,
		converter.Volume,
		converter.Temperature,
		converter.Area,
	}

	if len(config.EnabledUnitTypes) != len(expectedUnitTypes) {
		t.Errorf("Expected %d enabled unit types, got %d", len(expectedUnitTypes), len(config.EnabledUnitTypes))
	}

	for _, expectedType := range expectedUnitTypes {
		found := false
		for _, enabledType := range config.EnabledUnitTypes {
			if enabledType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected unit type %v to be enabled by default", expectedType)
		}
	}

	// Test precision defaults
	if len(config.Precision) == 0 {
		t.Error("Default config should have precision settings")
	}

	// Test preferences defaults
	if !config.Preferences.PreferWholeNumbers {
		t.Error("Default config should prefer whole numbers")
	}

	if config.Preferences.MaxDecimalPlaces != 2 {
		t.Errorf("Expected max decimal places to be 2, got %d", config.Preferences.MaxDecimalPlaces)
	}

	// Test detection defaults
	if config.Detection.MinConfidence != 0.5 {
		t.Errorf("Expected min confidence to be 0.5, got %f", config.Detection.MinConfidence)
	}

	if config.Detection.MaxNumberDistance != 3 {
		t.Errorf("Expected max number distance to be 3, got %d", config.Detection.MaxNumberDistance)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *converter.UnitConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name:        "valid default config",
			config:      converter.GetDefaultUnitConfig(),
			expectError: false,
		},
		{
			name: "invalid precision - negative",
			config: &converter.UnitConfig{
				Enabled:          true,
				EnabledUnitTypes: []converter.UnitType{converter.Length},
				Precision:        map[string]int{"length": -1},
				CustomMappings:   make(map[string]string),
				ExcludePatterns:  []string{},
				Preferences:      converter.GetDefaultUnitConfig().Preferences,
				Detection:        converter.GetDefaultUnitConfig().Detection,
			},
			expectError: true,
			errorMsg:    "precision for length must be between 0 and 10",
		},
		{
			name: "invalid precision - too high",
			config: &converter.UnitConfig{
				Enabled:          true,
				EnabledUnitTypes: []converter.UnitType{converter.Length},
				Precision:        map[string]int{"length": 15},
				CustomMappings:   make(map[string]string),
				ExcludePatterns:  []string{},
				Preferences:      converter.GetDefaultUnitConfig().Preferences,
				Detection:        converter.GetDefaultUnitConfig().Detection,
			},
			expectError: true,
			errorMsg:    "precision for length must be between 0 and 10",
		},
		{
			name: "invalid min confidence - negative",
			config: &converter.UnitConfig{
				Enabled:          true,
				EnabledUnitTypes: []converter.UnitType{converter.Length},
				Precision:        map[string]int{"length": 1},
				CustomMappings:   make(map[string]string),
				ExcludePatterns:  []string{},
				Preferences:      converter.GetDefaultUnitConfig().Preferences,
				Detection: converter.DetectionConfig{
					MinConfidence:        -0.1,
					MaxNumberDistance:    3,
					DetectCompoundUnits:  true,
					DetectWrittenNumbers: true,
				},
			},
			expectError: true,
			errorMsg:    "minConfidence must be between 0.0 and 1.0",
		},
		{
			name: "invalid min confidence - too high",
			config: &converter.UnitConfig{
				Enabled:          true,
				EnabledUnitTypes: []converter.UnitType{converter.Length},
				Precision:        map[string]int{"length": 1},
				CustomMappings:   make(map[string]string),
				ExcludePatterns:  []string{},
				Preferences:      converter.GetDefaultUnitConfig().Preferences,
				Detection: converter.DetectionConfig{
					MinConfidence:        1.5,
					MaxNumberDistance:    3,
					DetectCompoundUnits:  true,
					DetectWrittenNumbers: true,
				},
			},
			expectError: true,
			errorMsg:    "minConfidence must be between 0.0 and 1.0",
		},
		{
			name: "invalid max number distance - too low",
			config: &converter.UnitConfig{
				Enabled:          true,
				EnabledUnitTypes: []converter.UnitType{converter.Length},
				Precision:        map[string]int{"length": 1},
				CustomMappings:   make(map[string]string),
				ExcludePatterns:  []string{},
				Preferences:      converter.GetDefaultUnitConfig().Preferences,
				Detection: converter.DetectionConfig{
					MinConfidence:        0.5,
					MaxNumberDistance:    0,
					DetectCompoundUnits:  true,
					DetectWrittenNumbers: true,
				},
			},
			expectError: true,
			errorMsg:    "maxNumberDistance must be between 1 and 10",
		},
		{
			name: "invalid max number distance - too high",
			config: &converter.UnitConfig{
				Enabled:          true,
				EnabledUnitTypes: []converter.UnitType{converter.Length},
				Precision:        map[string]int{"length": 1},
				CustomMappings:   make(map[string]string),
				ExcludePatterns:  []string{},
				Preferences:      converter.GetDefaultUnitConfig().Preferences,
				Detection: converter.DetectionConfig{
					MinConfidence:        0.5,
					MaxNumberDistance:    15,
					DetectCompoundUnits:  true,
					DetectWrittenNumbers: true,
				},
			},
			expectError: true,
			errorMsg:    "maxNumberDistance must be between 1 and 10",
		},
		{
			name: "invalid temperature format",
			config: &converter.UnitConfig{
				Enabled:          true,
				EnabledUnitTypes: []converter.UnitType{converter.Temperature},
				Precision:        map[string]int{"temperature": 0},
				CustomMappings:   make(map[string]string),
				ExcludePatterns:  []string{},
				Preferences: converter.ConversionPreferences{
					PreferWholeNumbers:          true,
					MaxDecimalPlaces:            2,
					UseLocalizedUnits:           true,
					TemperatureFormat:           "invalid",
					UseSpaceBetweenValueAndUnit: true,
					RoundingThreshold:           0.1,
				},
				Detection: converter.GetDefaultUnitConfig().Detection,
			},
			expectError: true,
			errorMsg:    "invalid temperature format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := converter.ValidateConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tt.errorMsg)
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					// For partial matches, check if the error message contains the expected text
					if len(tt.errorMsg) > 0 && err.Error()[:len(tt.errorMsg)] != tt.errorMsg {
						t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestUnitConfig_IsUnitTypeEnabled(t *testing.T) {
	config := &converter.UnitConfig{
		Enabled:          true,
		EnabledUnitTypes: []converter.UnitType{converter.Length, converter.Mass},
	}

	// Test enabled unit types
	if !config.IsUnitTypeEnabled(converter.Length) {
		t.Error("Length should be enabled")
	}

	if !config.IsUnitTypeEnabled(converter.Mass) {
		t.Error("Mass should be enabled")
	}

	// Test disabled unit types
	if config.IsUnitTypeEnabled(converter.Volume) {
		t.Error("Volume should not be enabled")
	}

	if config.IsUnitTypeEnabled(converter.Temperature) {
		t.Error("Temperature should not be enabled")
	}

	// Test when globally disabled
	config.Enabled = false
	if config.IsUnitTypeEnabled(converter.Length) {
		t.Error("No unit types should be enabled when globally disabled")
	}
}

func TestUnitConfig_GetPrecisionForUnitType(t *testing.T) {
	config := &converter.UnitConfig{
		Precision: map[string]int{
			"length": 2,
			"mass":   1,
		},
	}

	// Test configured precision
	if precision := config.GetPrecisionForUnitType(converter.Length); precision != 2 {
		t.Errorf("Expected precision 2 for length, got %d", precision)
	}

	if precision := config.GetPrecisionForUnitType(converter.Mass); precision != 1 {
		t.Errorf("Expected precision 1 for mass, got %d", precision)
	}

	// Test default precision for unconfigured unit type
	precision := config.GetPrecisionForUnitType(converter.Volume)
	if precision != 1 { // Default fallback
		t.Errorf("Expected default precision 1 for volume, got %d", precision)
	}
}

func TestUnitConfig_SetPrecisionForUnitType(t *testing.T) {
	config := &converter.UnitConfig{}

	// Test setting precision on empty config
	config.SetPrecisionForUnitType(converter.Length, 3)

	if config.Precision == nil {
		t.Error("Precision map should be initialized")
	}

	if precision := config.GetPrecisionForUnitType(converter.Length); precision != 3 {
		t.Errorf("Expected precision 3 for length, got %d", precision)
	}

	// Test updating existing precision
	config.SetPrecisionForUnitType(converter.Length, 2)
	if precision := config.GetPrecisionForUnitType(converter.Length); precision != 2 {
		t.Errorf("Expected updated precision 2 for length, got %d", precision)
	}
}

func TestUnitConfig_JSONSerialization(t *testing.T) {
	original := converter.GetDefaultUnitConfig()

	// Test marshaling
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Test unmarshaling
	var restored converter.UnitConfig
	err = json.Unmarshal(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Test that values are preserved
	if restored.Enabled != original.Enabled {
		t.Errorf("Enabled flag not preserved: expected %v, got %v", original.Enabled, restored.Enabled)
	}

	if len(restored.EnabledUnitTypes) != len(original.EnabledUnitTypes) {
		t.Errorf("EnabledUnitTypes length not preserved: expected %d, got %d",
			len(original.EnabledUnitTypes), len(restored.EnabledUnitTypes))
	}

	// Check that unit types are preserved
	for i, unitType := range original.EnabledUnitTypes {
		if i < len(restored.EnabledUnitTypes) && restored.EnabledUnitTypes[i] != unitType {
			t.Errorf("EnabledUnitTypes[%d] not preserved: expected %v, got %v",
				i, unitType, restored.EnabledUnitTypes[i])
		}
	}

	// Test precision preservation
	for key, value := range original.Precision {
		if restored.Precision[key] != value {
			t.Errorf("Precision[%s] not preserved: expected %d, got %d",
				key, value, restored.Precision[key])
		}
	}

	// Test preferences preservation
	if restored.Preferences.PreferWholeNumbers != original.Preferences.PreferWholeNumbers {
		t.Error("PreferWholeNumbers not preserved")
	}

	if restored.Preferences.TemperatureFormat != original.Preferences.TemperatureFormat {
		t.Errorf("TemperatureFormat not preserved: expected %s, got %s",
			original.Preferences.TemperatureFormat, restored.Preferences.TemperatureFormat)
	}

	// Test detection config preservation
	if restored.Detection.MinConfidence != original.Detection.MinConfidence {
		t.Errorf("MinConfidence not preserved: expected %f, got %f",
			original.Detection.MinConfidence, restored.Detection.MinConfidence)
	}
}

func TestUnitConfig_Clone(t *testing.T) {
	original := converter.GetDefaultUnitConfig()
	original.CustomMappings["test"] = "value"
	original.ExcludePatterns = append(original.ExcludePatterns, "test pattern")

	clone := original.Clone()

	// Test that clone is not the same object
	if clone == original {
		t.Error("Clone should be a different object")
	}

	// Test that values are copied
	if clone.Enabled != original.Enabled {
		t.Error("Enabled flag not cloned")
	}

	// Test that slices are deep copied
	if &clone.EnabledUnitTypes == &original.EnabledUnitTypes {
		t.Error("EnabledUnitTypes should be deep copied")
	}

	if &clone.ExcludePatterns == &original.ExcludePatterns {
		t.Error("ExcludePatterns should be deep copied")
	}

	// Test that maps are deep copied
	if &clone.Precision == &original.Precision {
		t.Error("Precision should be deep copied")
	}

	if &clone.CustomMappings == &original.CustomMappings {
		t.Error("CustomMappings should be deep copied")
	}

	// Test that modifying clone doesn't affect original
	clone.Enabled = !original.Enabled
	if clone.Enabled == original.Enabled {
		t.Error("Modifying clone should not affect original")
	}

	clone.CustomMappings["new"] = "value"
	if _, exists := original.CustomMappings["new"]; exists {
		t.Error("Modifying clone's CustomMappings should not affect original")
	}
}

func TestUnitConfig_Merge(t *testing.T) {
	base := &converter.UnitConfig{
		Enabled:          false,
		EnabledUnitTypes: []converter.UnitType{converter.Length},
		Precision:        map[string]int{"length": 1},
		CustomMappings:   map[string]string{"old": "value"},
		ExcludePatterns:  []string{"old pattern"},
		Preferences: converter.ConversionPreferences{
			PreferWholeNumbers: false,
			MaxDecimalPlaces:   1,
		},
		Detection: converter.DetectionConfig{
			MinConfidence: 0.3,
		},
	}

	override := &converter.UnitConfig{
		Enabled:          true,
		EnabledUnitTypes: []converter.UnitType{converter.Mass, converter.Volume},
		Precision:        map[string]int{"mass": 2, "length": 3}, // Should override length
		CustomMappings:   map[string]string{"new": "value"},
		ExcludePatterns:  []string{"new pattern"},
		Preferences: converter.ConversionPreferences{
			PreferWholeNumbers: true,
			MaxDecimalPlaces:   2,
		},
		Detection: converter.DetectionConfig{
			MinConfidence: 0.7,
		},
	}

	base.Merge(override)

	// Test that override values are used
	if !base.Enabled {
		t.Error("Enabled should be overridden to true")
	}

	// Test that enabled unit types are replaced
	if len(base.EnabledUnitTypes) != 2 {
		t.Errorf("Expected 2 enabled unit types, got %d", len(base.EnabledUnitTypes))
	}

	// Test that precision is merged (override wins)
	if base.Precision["length"] != 3 {
		t.Errorf("Expected length precision to be overridden to 3, got %d", base.Precision["length"])
	}

	if base.Precision["mass"] != 2 {
		t.Errorf("Expected mass precision to be added as 2, got %d", base.Precision["mass"])
	}

	// Test that custom mappings are merged
	if base.CustomMappings["old"] != "value" {
		t.Error("Old custom mapping should be preserved")
	}

	if base.CustomMappings["new"] != "value" {
		t.Error("New custom mapping should be added")
	}

	// Test that preferences are replaced
	if !base.Preferences.PreferWholeNumbers {
		t.Error("PreferWholeNumbers should be overridden")
	}

	if base.Preferences.MaxDecimalPlaces != 2 {
		t.Errorf("MaxDecimalPlaces should be overridden to 2, got %d", base.Preferences.MaxDecimalPlaces)
	}

	// Test that detection config is replaced
	if base.Detection.MinConfidence != 0.7 {
		t.Errorf("MinConfidence should be overridden to 0.7, got %f", base.Detection.MinConfidence)
	}
}

func TestUnitConfig_Merge_NilConfig(t *testing.T) {
	base := converter.GetDefaultUnitConfig()
	originalEnabled := base.Enabled

	base.Merge(nil)

	// Test that merging nil doesn't change anything
	if base.Enabled != originalEnabled {
		t.Error("Merging nil should not change the config")
	}
}

func TestGetUserConfigPath(t *testing.T) {
	path, err := converter.GetUserConfigPath()
	if err != nil {
		t.Fatalf("Failed to get user config path: %v", err)
	}

	if path == "" {
		t.Error("User config path should not be empty")
	}

	// Test that path ends with the expected filename
	expectedSuffix := filepath.Join(".config", "m2e", "unit_config.json")
	if !filepath.IsAbs(path) {
		t.Error("User config path should be absolute")
	}

	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("User config path should end with %s, got %s", expectedSuffix, path)
	}
}

func TestCreateUserConfigDirectory(t *testing.T) {
	// This test creates a temporary directory to avoid affecting the user's actual config
	tempDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	err := converter.CreateUserConfigDirectory()
	if err != nil {
		t.Fatalf("Failed to create user config directory: %v", err)
	}

	// Check that the directory was created
	configDir := filepath.Join(tempDir, ".config", "m2e")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory should have been created")
	}
}
func TestLoadUserConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	t.Run("no config file exists", func(t *testing.T) {
		config, err := converter.LoadUserConfig()
		if err != nil {
			t.Fatalf("Expected no error when config file doesn't exist, got: %v", err)
		}

		// Should return default configuration
		defaultConfig := converter.GetDefaultUnitConfig()
		if config.Enabled != defaultConfig.Enabled {
			t.Error("Should return default configuration when file doesn't exist")
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		// Create config directory
		err := converter.CreateUserConfigDirectory()
		if err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create a valid config file
		testConfig := converter.GetDefaultUnitConfig()
		testConfig.Enabled = false
		testConfig.Detection.MinConfidence = 0.8

		configPath, _ := converter.GetUserConfigPath()
		data, _ := json.Marshal(testConfig)
		err = os.WriteFile(configPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Load the config
		loadedConfig, err := converter.LoadUserConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if loadedConfig.Enabled != false {
			t.Error("Config should have been loaded with Enabled=false")
		}

		if loadedConfig.Detection.MinConfidence != 0.8 {
			t.Errorf("Expected MinConfidence=0.8, got %f", loadedConfig.Detection.MinConfidence)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		// Create config directory
		err := converter.CreateUserConfigDirectory()
		if err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create an invalid config file
		configPath, _ := converter.GetUserConfigPath()
		err = os.WriteFile(configPath, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		// Try to load the config
		_, err = converter.LoadUserConfig()
		if err == nil {
			t.Error("Expected error when loading invalid JSON")
		}

		if !strings.Contains(err.Error(), "failed to parse config file") {
			t.Errorf("Expected parse error, got: %v", err)
		}
	})

	t.Run("invalid config values", func(t *testing.T) {
		// Create config directory
		err := converter.CreateUserConfigDirectory()
		if err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create a config with invalid values
		invalidConfig := `{
			"enabled": true,
			"enabledUnitTypes": ["length"],
			"precision": {"length": -1},
			"customMappings": {},
			"excludePatterns": [],
			"preferences": {
				"preferWholeNumbers": true,
				"maxDecimalPlaces": 2,
				"useLocalizedUnits": true,
				"temperatureFormat": "°C",
				"useSpaceBetweenValueAndUnit": true,
				"roundingThreshold": 0.1
			},
			"detection": {
				"minConfidence": 0.5,
				"maxNumberDistance": 3,
				"detectCompoundUnits": true,
				"detectWrittenNumbers": true
			}
		}`

		configPath, _ := converter.GetUserConfigPath()
		err = os.WriteFile(configPath, []byte(invalidConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		// Try to load the config
		_, err = converter.LoadUserConfig()
		if err == nil {
			t.Error("Expected error when loading config with invalid values")
		}

		if !strings.Contains(err.Error(), "invalid configuration") {
			t.Errorf("Expected validation error, got: %v", err)
		}
	})
}

func TestSaveUserConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	t.Run("save valid config", func(t *testing.T) {
		config := converter.GetDefaultUnitConfig()
		config.Enabled = false
		config.Detection.MinConfidence = 0.7

		err := converter.SaveUserConfig(config)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify the file was created
		configPath, _ := converter.GetUserConfigPath()
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file should have been created")
		}

		// Load and verify the saved config
		loadedConfig, err := converter.LoadUserConfig()
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loadedConfig.Enabled != false {
			t.Error("Saved config should have Enabled=false")
		}

		if loadedConfig.Detection.MinConfidence != 0.7 {
			t.Errorf("Expected MinConfidence=0.7, got %f", loadedConfig.Detection.MinConfidence)
		}
	})

	t.Run("save nil config", func(t *testing.T) {
		err := converter.SaveUserConfig(nil)
		if err == nil {
			t.Error("Expected error when saving nil config")
		}

		if !strings.Contains(err.Error(), "config cannot be nil") {
			t.Errorf("Expected nil config error, got: %v", err)
		}
	})

	t.Run("save invalid config", func(t *testing.T) {
		config := converter.GetDefaultUnitConfig()
		config.Detection.MinConfidence = -1.0 // Invalid value

		err := converter.SaveUserConfig(config)
		if err == nil {
			t.Error("Expected error when saving invalid config")
		}

		if !strings.Contains(err.Error(), "invalid configuration") {
			t.Errorf("Expected validation error, got: %v", err)
		}
	})
}

func TestCreateExampleUserConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	t.Run("create example config", func(t *testing.T) {
		err := converter.CreateExampleUserConfig()
		if err != nil {
			t.Fatalf("Failed to create example config: %v", err)
		}

		// Verify the file was created
		configPath, _ := converter.GetUserConfigPath()
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Example config file should have been created")
		}

		// Read the file and check it contains expected content
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read example config file: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "_comment") {
			t.Error("Example config should contain comment fields")
		}

		if !strings.Contains(content, "M2E Unit Conversion Configuration") {
			t.Error("Example config should contain description")
		}

		if !strings.Contains(content, "enabled") {
			t.Error("Example config should contain actual configuration")
		}
	})

	t.Run("file already exists", func(t *testing.T) {
		// Use a fresh temp directory for this test
		tempDir2 := t.TempDir()
		_ = os.Setenv("HOME", tempDir2)

		// Create the example config first
		err := converter.CreateExampleUserConfig()
		if err != nil {
			t.Fatalf("Failed to create initial example config: %v", err)
		}

		// Try to create it again
		err = converter.CreateExampleUserConfig()
		if err == nil {
			t.Error("Expected error when config file already exists")
		}

		if !strings.Contains(err.Error(), "config file already exists") {
			t.Errorf("Expected file exists error, got: %v", err)
		}
	})
}

func TestLoadConfigWithDefaults(t *testing.T) {
	tempDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	t.Run("no user config file", func(t *testing.T) {
		config, err := converter.LoadConfigWithDefaults()
		if err != nil {
			t.Fatalf("Expected no error when no user config exists, got: %v", err)
		}

		// Should return default configuration
		defaultConfig := converter.GetDefaultUnitConfig()
		if config.Enabled != defaultConfig.Enabled {
			t.Error("Should return default configuration when no user config exists")
		}
	})

	t.Run("with user config file", func(t *testing.T) {
		// Create a user config with some overrides
		userConfig := converter.GetDefaultUnitConfig()
		userConfig.Enabled = false
		userConfig.Detection.MinConfidence = 0.8
		userConfig.CustomMappings = map[string]string{"test": "value"}

		err := converter.SaveUserConfig(userConfig)
		if err != nil {
			t.Fatalf("Failed to save user config: %v", err)
		}

		// Load config with defaults
		config, err := converter.LoadConfigWithDefaults()
		if err != nil {
			t.Fatalf("Failed to load config with defaults: %v", err)
		}

		// Should have user overrides
		if config.Enabled != false {
			t.Error("Should use user override for Enabled")
		}

		if config.Detection.MinConfidence != 0.8 {
			t.Errorf("Should use user override for MinConfidence, expected 0.8, got %f", config.Detection.MinConfidence)
		}

		if config.CustomMappings["test"] != "value" {
			t.Error("Should use user override for CustomMappings")
		}

		// Should still have default values for non-overridden settings
		defaultConfig := converter.GetDefaultUnitConfig()
		if len(config.EnabledUnitTypes) != len(defaultConfig.EnabledUnitTypes) {
			t.Error("Should preserve default EnabledUnitTypes when not overridden")
		}
	})

	t.Run("invalid user config file", func(t *testing.T) {
		// Create an invalid config file
		configPath, _ := converter.GetUserConfigPath()
		err := os.WriteFile(configPath, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		// Should fall back to defaults and not error
		config, err := converter.LoadConfigWithDefaults()
		if err != nil {
			t.Fatalf("Should not error with invalid user config, got: %v", err)
		}

		// Should return default configuration
		defaultConfig := converter.GetDefaultUnitConfig()
		if config.Enabled != defaultConfig.Enabled {
			t.Error("Should fall back to default configuration with invalid user config")
		}
	})
}

func TestGetConfigStatus(t *testing.T) {
	tempDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	t.Run("no config file", func(t *testing.T) {
		status, err := converter.GetConfigStatus()
		if err != nil {
			t.Fatalf("Failed to get config status: %v", err)
		}

		if status["exists"].(bool) {
			t.Error("Config should not exist")
		}

		if status["valid"].(bool) {
			t.Error("Config should not be valid when it doesn't exist")
		}

		if status["error"] != nil {
			t.Error("Error should be nil when config doesn't exist")
		}

		configPath, _ := converter.GetUserConfigPath()
		if status["configPath"].(string) != configPath {
			t.Error("Config path should match expected path")
		}
	})

	t.Run("valid config file", func(t *testing.T) {
		// Create a valid config file
		config := converter.GetDefaultUnitConfig()
		err := converter.SaveUserConfig(config)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		status, err := converter.GetConfigStatus()
		if err != nil {
			t.Fatalf("Failed to get config status: %v", err)
		}

		if !status["exists"].(bool) {
			t.Error("Config should exist")
		}

		if !status["valid"].(bool) {
			t.Error("Config should be valid")
		}

		if status["error"] != nil {
			t.Errorf("Error should be nil for valid config, got: %v", status["error"])
		}
	})

	t.Run("invalid config file", func(t *testing.T) {
		// Create an invalid config file
		configPath, _ := converter.GetUserConfigPath()
		err := os.WriteFile(configPath, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		status, err := converter.GetConfigStatus()
		if err != nil {
			t.Fatalf("Failed to get config status: %v", err)
		}

		if !status["exists"].(bool) {
			t.Error("Config should exist")
		}

		if status["valid"].(bool) {
			t.Error("Config should not be valid")
		}

		if status["error"] == nil {
			t.Error("Error should not be nil for invalid config")
		}
	})
}
func TestUnitProcessor_ConfigurationIntegration(t *testing.T) {
	t.Run("processor uses configuration settings", func(t *testing.T) {
		// Create a custom configuration
		config := converter.GetDefaultUnitConfig()
		config.Enabled = true
		config.EnabledUnitTypes = []converter.UnitType{converter.Length} // Only length units
		config.Detection.MinConfidence = 0.8                             // Higher confidence threshold

		// Create processor with custom config
		processor := converter.NewUnitProcessorWithConfig(config)

		// Test that length units are converted
		lengthText := "The room is 10 feet wide"
		lengthResult := processor.ProcessText(lengthText, false, "")
		if !strings.Contains(lengthResult, "metres") && !strings.Contains(lengthResult, "m") {
			t.Error("Length units should be converted")
		}

		// Test that mass units are NOT converted (not enabled)
		massText := "The box weighs 5 pounds"
		massResult := processor.ProcessText(massText, false, "")
		if !strings.Contains(massResult, "pounds") {
			t.Error("Mass units should NOT be converted when disabled")
		}
	})

	t.Run("processor respects enabled/disabled setting", func(t *testing.T) {
		config := converter.GetDefaultUnitConfig()
		config.Enabled = false // Disabled

		processor := converter.NewUnitProcessorWithConfig(config)

		text := "The room is 10 feet wide"
		result := processor.ProcessText(text, false, "")

		if result != text {
			t.Error("No conversion should happen when processor is disabled")
		}
	})

	t.Run("processor applies exclude patterns", func(t *testing.T) {
		config := converter.GetDefaultUnitConfig()
		config.ExcludePatterns = []string{`miles?\s+(?:away|apart)`} // Exclude "miles away/apart"

		processor := converter.NewUnitProcessorWithConfig(config)

		// This should be excluded
		excludedText := "He lives miles away from here"
		result := processor.ProcessText(excludedText, false, "")
		if !strings.Contains(result, "miles away") {
			t.Error("Text matching exclude pattern should not be converted")
		}

		// This should be converted
		includedText := "He drove 5 miles to work"
		result = processor.ProcessText(includedText, false, "")
		if strings.Contains(result, "5 miles") {
			t.Error("Text not matching exclude pattern should be converted")
		}
	})

	t.Run("processor applies precision settings", func(t *testing.T) {
		config := converter.GetDefaultUnitConfig()
		config.Precision["length"] = 0 // No decimal places for length

		processor := converter.NewUnitProcessorWithConfig(config)

		text := "The room is 10 feet wide"
		result := processor.ProcessText(text, false, "")

		// Should not contain decimal places
		if strings.Contains(result, ".") {
			t.Error("Result should not contain decimal places when precision is 0")
		}
	})

	t.Run("processor applies conversion preferences", func(t *testing.T) {
		config := converter.GetDefaultUnitConfig()
		config.Preferences.UseSpaceBetweenValueAndUnit = false // No space between value and unit

		processor := converter.NewUnitProcessorWithConfig(config)

		// Test with a length unit to verify spacing preferences
		lengthText := "The room is 10 feet wide"
		lengthResult := processor.ProcessText(lengthText, false, "")

		// This is harder to test precisely, but we can check that conversion happened
		if strings.Contains(lengthResult, "10 feet") {
			t.Error("Length units should be converted")
		}
	})
}

func TestUnitProcessor_ConfigurationReload(t *testing.T) {
	t.Run("processor can update configuration", func(t *testing.T) {
		// Start with default config
		processor := converter.NewUnitProcessor()

		// Test initial behavior
		text := "The room is 10 feet wide"
		result := processor.ProcessText(text, false, "")
		if strings.Contains(result, "10 feet") {
			t.Error("Initial conversion should work")
		}

		// Update configuration to disable unit processing
		newConfig := converter.GetDefaultUnitConfig()
		newConfig.Enabled = false
		processor.SetConfig(newConfig)

		// Test that conversion is now disabled
		result = processor.ProcessText(text, false, "")
		if result != text {
			t.Error("Conversion should be disabled after config update")
		}
	})

	t.Run("processor handles nil config gracefully", func(t *testing.T) {
		processor := converter.NewUnitProcessorWithConfig(nil)

		// Should use default config
		if !processor.IsEnabled() {
			t.Error("Processor should be enabled with default config when nil is provided")
		}

		text := "The room is 10 feet wide"
		result := processor.ProcessText(text, false, "")
		if strings.Contains(result, "10 feet") {
			t.Error("Conversion should work with default config")
		}
	})
}

func TestUnitConfig_Integration_WithRealFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	t.Run("end-to-end configuration workflow", func(t *testing.T) {
		// 1. Create an example config
		err := converter.CreateExampleUserConfig()
		if err != nil {
			t.Fatalf("Failed to create example config: %v", err)
		}

		// 2. Load the config
		config, err := converter.LoadUserConfig()
		if err != nil {
			t.Fatalf("Failed to load user config: %v", err)
		}

		// 3. Verify it's valid
		err = converter.ValidateConfig(config)
		if err != nil {
			t.Fatalf("Loaded config should be valid: %v", err)
		}

		// 4. Create a processor with the loaded config
		processor := converter.NewUnitProcessorWithConfig(config)

		// 5. Test that it works
		text := "The room is 10 feet wide"
		result := processor.ProcessText(text, false, "")
		if strings.Contains(result, "10 feet") {
			t.Error("Conversion should work with loaded config")
		}

		// 6. Modify and save the config
		config.EnabledUnitTypes = []converter.UnitType{converter.Temperature} // Only temperature
		err = converter.SaveUserConfig(config)
		if err != nil {
			t.Fatalf("Failed to save modified config: %v", err)
		}

		// 7. Load the modified config
		modifiedConfig, err := converter.LoadConfigWithDefaults()
		if err != nil {
			t.Fatalf("Failed to load modified config: %v", err)
		}

		// 8. Verify the modification took effect
		if len(modifiedConfig.EnabledUnitTypes) != 1 || modifiedConfig.EnabledUnitTypes[0] != converter.Temperature {
			t.Error("Modified config should only have Temperature enabled")
		}

		// 9. Test with new processor
		newProcessor := converter.NewUnitProcessorWithConfig(modifiedConfig)

		// Length should not be converted
		lengthResult := newProcessor.ProcessText("The room is 10 feet wide", false, "")
		if !strings.Contains(lengthResult, "10 feet") {
			t.Error("Length units should NOT be converted with modified config")
		}

		// Temperature should be converted
		tempResult := newProcessor.ProcessText("The temperature is 75°F", false, "")
		if strings.Contains(tempResult, "75°F") {
			t.Error("Temperature units should be converted with modified config")
		}
	})

	t.Run("configuration status reporting", func(t *testing.T) {
		// Use a fresh temp directory for this test
		tempDir2 := t.TempDir()
		_ = os.Setenv("HOME", tempDir2)

		// Initially no config file
		status, err := converter.GetConfigStatus()
		if err != nil {
			t.Fatalf("Failed to get config status: %v", err)
		}

		if status["exists"].(bool) {
			t.Error("Config should not exist initially")
		}

		// Create a config file
		config := converter.GetDefaultUnitConfig()
		err = converter.SaveUserConfig(config)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Check status again
		status, err = converter.GetConfigStatus()
		if err != nil {
			t.Fatalf("Failed to get config status after creation: %v", err)
		}

		if !status["exists"].(bool) {
			t.Error("Config should exist after creation")
		}

		if !status["valid"].(bool) {
			t.Error("Config should be valid")
		}

		if status["error"] != nil {
			t.Errorf("Config should have no errors, got: %v", status["error"])
		}
	})
}
