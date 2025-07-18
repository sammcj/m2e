package tests

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

// TestCase structures for loading from JSON
type DetectionTestCase struct {
	Name     string `json:"name"`
	Input    string `json:"input"`
	Expected []struct {
		Value      float64 `json:"value"`
		Unit       string  `json:"unit"`
		Type       string  `json:"type"`
		Confidence float64 `json:"confidence"`
	} `json:"expected"`
}

type ConversionTestCase struct {
	Name  string `json:"name"`
	Input struct {
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
		Type  string  `json:"type"`
	} `json:"input"`
	Expected struct {
		MetricValue float64 `json:"metric_value"`
		MetricUnit  string  `json:"metric_unit"`
		Formatted   string  `json:"formatted"`
	} `json:"expected"`
}

type IdiomTestCase struct {
	Name        string `json:"name"`
	Input       string `json:"input"`
	ShouldMatch bool   `json:"should_match"`
}

type EdgeCaseTestCase struct {
	Name  string `json:"name"`
	Input struct {
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
		Type  string  `json:"type"`
	} `json:"input"`
	ShouldWork bool `json:"should_work"`
}

type TestData struct {
	DetectionTests      []DetectionTestCase  `json:"detection_tests"`
	ConversionTests     []ConversionTestCase `json:"conversion_tests"`
	IdiomExclusionTests []IdiomTestCase      `json:"idiom_exclusion_tests"`
	EdgeCases           []EdgeCaseTestCase   `json:"edge_cases"`
}

// Helper function to load test data
func loadTestData(t *testing.T) TestData {
	data, err := os.ReadFile("testdata/unit_test_cases.json")
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	var testData TestData
	if err := json.Unmarshal(data, &testData); err != nil {
		t.Fatalf("Failed to parse test data: %v", err)
	}

	return testData
}

// Helper function to convert string type to UnitType
func stringToUnitType(s string) converter.UnitType {
	switch s {
	case "Length":
		return converter.Length
	case "Mass":
		return converter.Mass
	case "Volume":
		return converter.Volume
	case "Temperature":
		return converter.Temperature
	case "Area":
		return converter.Area
	default:
		return converter.Length // Default fallback
	}
}

// TestUnitDetection_BasicPatterns tests core unit detection functionality
func TestUnitDetection_BasicPatterns(t *testing.T) {
	detector := converter.NewContextualUnitDetector()
	testData := loadTestData(t)

	for _, tt := range testData.DetectionTests {
		t.Run(tt.Name, func(t *testing.T) {
			matches := detector.DetectUnits(tt.Input)

			if len(matches) != len(tt.Expected) {
				t.Errorf("Expected %d matches, got %d for text: '%s'",
					len(tt.Expected), len(matches), tt.Input)
				for i, match := range matches {
					t.Logf("Match %d: %s = %.2f (confidence: %.2f)",
						i, match.Unit, match.Value, match.Confidence)
				}
				return
			}

			for i, expected := range tt.Expected {
				if i >= len(matches) {
					break
				}
				match := matches[i]

				if match.Value != expected.Value {
					t.Errorf("Match %d: expected value %.2f, got %.2f",
						i, expected.Value, match.Value)
				}
				if match.Unit != expected.Unit {
					t.Errorf("Match %d: expected unit '%s', got '%s'",
						i, expected.Unit, match.Unit)
				}
				if match.UnitType != stringToUnitType(expected.Type) {
					t.Errorf("Match %d: expected type %s, got %v",
						i, expected.Type, match.UnitType)
				}
			}
		})
	}
}

// TestUnitDetection_IdiomExclusion tests that idiomatic expressions are excluded
func TestUnitDetection_IdiomExclusion(t *testing.T) {
	detector := converter.NewContextualUnitDetector()
	testData := loadTestData(t)

	for _, tt := range testData.IdiomExclusionTests {
		t.Run(tt.Name, func(t *testing.T) {
			matches := detector.DetectUnits(tt.Input)
			hasMatch := len(matches) > 0

			if hasMatch != tt.ShouldMatch {
				t.Errorf("Text: '%s'\nExpected match: %v, Got match: %v",
					tt.Input, tt.ShouldMatch, hasMatch)
				if hasMatch {
					for i, match := range matches {
						t.Logf("Unexpected match %d: %s = %.2f",
							i, match.Unit, match.Value)
					}
				}
			}
		})
	}
}

// TestUnitConversion_MathematicalAccuracy tests conversion accuracy
func TestUnitConversion_MathematicalAccuracy(t *testing.T) {
	conv := converter.NewBasicUnitConverter()
	testData := loadTestData(t)

	for _, tt := range testData.ConversionTests {
		t.Run(tt.Name, func(t *testing.T) {
			match := converter.UnitMatch{
				Value:      tt.Input.Value,
				Unit:       tt.Input.Unit,
				UnitType:   stringToUnitType(tt.Input.Type),
				Confidence: 0.9,
			}

			result, err := conv.Convert(match)
			if err != nil {
				t.Errorf("Convert() error = %v", err)
				return
			}

			// Check metric unit
			if result.MetricUnit != tt.Expected.MetricUnit {
				t.Errorf("Expected metric unit %s, got %s",
					tt.Expected.MetricUnit, result.MetricUnit)
			}

			// Check formatted output
			if result.Formatted != tt.Expected.Formatted {
				t.Errorf("Expected formatted '%s', got '%s'",
					tt.Expected.Formatted, result.Formatted)
			}

			// Check metric value with tolerance
			tolerance := 0.1
			if math.Abs(result.MetricValue-tt.Expected.MetricValue) > tolerance {
				t.Errorf("Expected metric value %.3f, got %.3f (tolerance %.3f)",
					tt.Expected.MetricValue, result.MetricValue, tolerance)
			}
		})
	}
}

// TestUnitConversion_UnitSelection tests appropriate metric unit selection
func TestUnitConversion_UnitSelection(t *testing.T) {
	conv := converter.NewBasicUnitConverter()

	tests := []struct {
		name         string
		match        converter.UnitMatch
		expectedUnit string
	}{
		{
			name: "very_small_inches_to_mm",
			match: converter.UnitMatch{
				Value: 0.01, Unit: "inches", UnitType: converter.Length, Confidence: 0.9,
			},
			expectedUnit: "mm",
		},
		{
			name: "small_inches_to_cm",
			match: converter.UnitMatch{
				Value: 6.0, Unit: "inches", UnitType: converter.Length, Confidence: 0.9,
			},
			expectedUnit: "cm",
		},
		{
			name: "medium_feet_to_metres",
			match: converter.UnitMatch{
				Value: 20.0, Unit: "feet", UnitType: converter.Length, Confidence: 0.9,
			},
			expectedUnit: "metres",
		},
		{
			name: "large_miles_to_km",
			match: converter.UnitMatch{
				Value: 5.0, Unit: "miles", UnitType: converter.Length, Confidence: 0.9,
			},
			expectedUnit: "km",
		},
		{
			name: "small_ounces_to_g",
			match: converter.UnitMatch{
				Value: 2.0, Unit: "ounces", UnitType: converter.Mass, Confidence: 0.9,
			},
			expectedUnit: "g",
		},
		{
			name: "medium_pounds_to_kg",
			match: converter.UnitMatch{
				Value: 50.0, Unit: "pounds", UnitType: converter.Mass, Confidence: 0.9,
			},
			expectedUnit: "kg",
		},
		{
			name: "small_fl_oz_to_ml",
			match: converter.UnitMatch{
				Value: 8.0, Unit: "fluid ounces", UnitType: converter.Volume, Confidence: 0.9,
			},
			expectedUnit: "ml",
		},
		{
			name: "large_gallons_to_litres",
			match: converter.UnitMatch{
				Value: 10.0, Unit: "gallons", UnitType: converter.Volume, Confidence: 0.9,
			},
			expectedUnit: "litres",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.match)
			if err != nil {
				t.Errorf("Convert() error = %v", err)
				return
			}

			if result.MetricUnit != tt.expectedUnit {
				t.Errorf("Expected unit %s, got %s", tt.expectedUnit, result.MetricUnit)
			}
		})
	}
}

// TestUnitConversion_Formatting tests formatting preferences and edge cases
func TestUnitConversion_Formatting(t *testing.T) {
	tests := []struct {
		name        string
		preferences converter.ConversionPreferences
		match       converter.UnitMatch
		expected    string
	}{
		{
			name: "whole_number_rounding",
			preferences: converter.ConversionPreferences{
				PreferWholeNumbers: true, MaxDecimalPlaces: 2,
				UseLocalizedUnits: true, TemperatureFormat: "°C",
				UseSpaceBetweenValueAndUnit: true, RoundingThreshold: 0.05,
			},
			match: converter.UnitMatch{
				Value: 3.2808398950131, Unit: "feet", UnitType: converter.Length, Confidence: 0.9,
			},
			expected: "100 cm",
		},
		{
			name: "no_space_formatting",
			preferences: converter.ConversionPreferences{
				PreferWholeNumbers: true, MaxDecimalPlaces: 2,
				UseLocalizedUnits: true, TemperatureFormat: "°C",
				UseSpaceBetweenValueAndUnit: false, RoundingThreshold: 0.05,
			},
			match: converter.UnitMatch{
				Value: 10.0, Unit: "feet", UnitType: converter.Length, Confidence: 0.9,
			},
			expected: "3metres",
		},
		{
			name: "temperature_formatting",
			preferences: converter.ConversionPreferences{
				PreferWholeNumbers: true, MaxDecimalPlaces: 2,
				UseLocalizedUnits: true, TemperatureFormat: "°C",
				UseSpaceBetweenValueAndUnit: true, RoundingThreshold: 0.05,
			},
			match: converter.UnitMatch{
				Value: 32.0, Unit: "fahrenheit", UnitType: converter.Temperature, Confidence: 0.9,
			},
			expected: "0°C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.NewBasicUnitConverter()
			conv.SetPreferences(tt.preferences)

			result, err := conv.Convert(tt.match)
			if err != nil {
				t.Errorf("Convert() error = %v", err)
				return
			}

			if result.Formatted != tt.expected {
				t.Errorf("Expected formatted '%s', got '%s'", tt.expected, result.Formatted)
			}
		})
	}
}

// TestUnitConversion_EdgeCases tests edge cases and error conditions
func TestUnitConversion_EdgeCases(t *testing.T) {
	conv := converter.NewBasicUnitConverter()
	testData := loadTestData(t)

	for _, tt := range testData.EdgeCases {
		t.Run(tt.Name, func(t *testing.T) {
			match := converter.UnitMatch{
				Value:      tt.Input.Value,
				Unit:       tt.Input.Unit,
				UnitType:   stringToUnitType(tt.Input.Type),
				Confidence: 0.9,
			}

			result, err := conv.Convert(match)

			if tt.ShouldWork {
				if err != nil {
					t.Errorf("Expected conversion to work but got error: %v", err)
				}
				// Basic sanity check - result should have some reasonable values
				if result.MetricUnit == "" || result.Formatted == "" {
					t.Errorf("Expected valid conversion result, got empty values")
				}
			} else {
				if err == nil {
					t.Errorf("Expected conversion to fail but it succeeded")
				}
			}
		})
	}
}

// TestUnitConfig_LoadAndValidate tests configuration loading and validation
func TestUnitConfig_LoadAndValidate(t *testing.T) {
	// Test default configuration
	t.Run("default_config", func(t *testing.T) {
		conv := converter.NewBasicUnitConverter()
		prefs := conv.GetPreferences()

		// Verify default preferences
		if !prefs.PreferWholeNumbers {
			t.Error("Expected PreferWholeNumbers to be true by default")
		}
		if prefs.MaxDecimalPlaces != 2 {
			t.Errorf("Expected MaxDecimalPlaces to be 2, got %d", prefs.MaxDecimalPlaces)
		}
		if prefs.TemperatureFormat != "°C" {
			t.Errorf("Expected TemperatureFormat to be '°C', got '%s'", prefs.TemperatureFormat)
		}
	})

	// Test custom configuration
	t.Run("custom_config", func(t *testing.T) {
		conv := converter.NewBasicUnitConverter()
		customPrefs := converter.ConversionPreferences{
			PreferWholeNumbers:          false,
			MaxDecimalPlaces:            3,
			UseLocalizedUnits:           false,
			TemperatureFormat:           "degrees Celsius",
			UseSpaceBetweenValueAndUnit: false,
			RoundingThreshold:           0.01,
		}

		conv.SetPreferences(customPrefs)
		retrievedPrefs := conv.GetPreferences()

		if retrievedPrefs.PreferWholeNumbers != customPrefs.PreferWholeNumbers {
			t.Error("Custom preferences not applied correctly")
		}
		if retrievedPrefs.MaxDecimalPlaces != customPrefs.MaxDecimalPlaces {
			t.Error("Custom MaxDecimalPlaces not applied correctly")
		}
	})

	// Test precision settings
	t.Run("precision_settings", func(t *testing.T) {
		conv := converter.NewBasicUnitConverter()
		conv.SetPrecision(converter.Length, 3)

		// Test that precision is applied (indirectly through conversion)
		match := converter.UnitMatch{
			Value: 3.14159, Unit: "feet", UnitType: converter.Length, Confidence: 0.9,
		}

		result, err := conv.Convert(match)
		if err != nil {
			t.Errorf("Convert() error = %v", err)
			return
		}

		// Should have reasonable precision in output
		if !strings.Contains(result.Formatted, ".") && result.MetricValue != math.Round(result.MetricValue) {
			t.Error("Expected precision to be applied to conversion result")
		}
	})
}

// TestUnitDetection_RealWorldScenarios tests detection with real-world text samples
func TestUnitDetection_RealWorldScenarios(t *testing.T) {
	detector := converter.NewContextualUnitDetector()

	// Load real world examples
	data, err := os.ReadFile("testdata/real_world_examples.txt")
	if err != nil {
		t.Fatalf("Failed to load real world examples: %v", err)
	}

	examples := strings.Split(string(data), "\n\n")

	tests := []struct {
		name            string
		text            string
		expectedMatches int
	}{
		{"construction_spec", examples[1], 5}, // 12 feet, 10 feet, 9-foot, 120 square feet, 2 tons
		{"recipe", examples[2], 4},            // 2 pounds, 16 ounces, 350°F, 9-inch
		{"travel", examples[3], 3},            // 150 miles, 50 miles, 85°F
		{"mixed_idioms", examples[4], 2},      // 6-foot, 200 pounds (miles ahead is idiomatic)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.text == "" {
				t.Skip("Example text not found")
				return
			}

			matches := detector.DetectUnits(tt.text)
			if len(matches) != tt.expectedMatches {
				t.Errorf("Expected %d matches, got %d for text: '%s'",
					tt.expectedMatches, len(matches), tt.text)
				for i, match := range matches {
					t.Logf("Match %d: %s = %.2f %s (confidence: %.2f)",
						i, match.Unit, match.Value, match.Unit, match.Confidence)
				}
			}
		})
	}
}
