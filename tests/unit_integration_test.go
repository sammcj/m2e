package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

// TestUnitConversion_EndToEnd tests the complete unit conversion pipeline
func TestUnitConversion_EndToEnd(t *testing.T) {
	// Create detector and converter
	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple_measurement",
			input:    "The room is 12 feet wide",
			expected: "The room is 3.7 metres wide",
		},
		{
			name:     "multiple_units",
			input:    "Package weighs 5 pounds and is 10 inches long",
			expected: "Package weighs 2.3 kg and is 25.4 cm long",
		},
		{
			name:     "temperature_conversion",
			input:    "Set oven to 350°F",
			expected: "Set oven to 177°C",
		},
		{
			name:     "mixed_unit_types",
			input:    "Drive 5 miles to buy 2 gallons of gas at 75°F",
			expected: "Drive 8 km to buy 7.6 litres of gas at 24°C",
		},
		{
			name:     "compound_units",
			input:    "Install a 6-foot fence",
			expected: "Install a 1.8 metre fence",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Detect units
			matches := detector.DetectUnits(tt.input)
			if len(matches) == 0 {
				t.Errorf("No units detected in: %s", tt.input)
				return
			}

			// Step 2: Convert each unit
			result := tt.input
			offset := 0

			for _, match := range matches {
				conversion, err := conv.Convert(match)
				if err != nil {
					t.Errorf("Conversion failed: %v", err)
					continue
				}

				// Replace the original unit with converted unit
				start := match.Start + offset
				end := match.End + offset

				if start < 0 || end > len(result) {
					continue // Skip invalid positions
				}

				originalText := result[start:end]
				result = result[:start] + conversion.Formatted + result[end:]
				offset += len(conversion.Formatted) - len(originalText)
			}

			if result != tt.expected {
				t.Errorf("End-to-end conversion failed:\nInput:    %s\nExpected: %s\nGot:      %s",
					tt.input, tt.expected, result)
			}
		})
	}
}

// TestUnitConversion_RealWorldScenarios tests with real-world text samples
func TestUnitConversion_RealWorldScenarios(t *testing.T) {
	// Load real world examples
	data, err := os.ReadFile("testdata/real_world_examples.txt")
	if err != nil {
		t.Fatalf("Failed to load real world examples: %v", err)
	}

	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	examples := strings.Split(string(data), "\n\n")

	tests := []struct {
		name            string
		input           string
		expectedMatches int
		shouldConvert   bool
	}{
		{
			name:            "construction_specification",
			input:           examples[1], // Construction spec
			expectedMatches: 5,           // Should find multiple units
			shouldConvert:   true,
		},
		{
			name:            "recipe_measurements",
			input:           examples[2], // Recipe
			expectedMatches: 4,           // Should find units in recipe
			shouldConvert:   true,
		},
		{
			name:            "travel_description",
			input:           examples[3], // Travel
			expectedMatches: 3,           // Should find distance and temperature
			shouldConvert:   true,
		},
		{
			name:            "mixed_with_idioms",
			input:           examples[4], // Mixed with idioms
			expectedMatches: 2,           // Should exclude idioms but find real measurements
			shouldConvert:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == "" {
				t.Skip("Example text not found")
				return
			}

			// Test detection
			matches := detector.DetectUnits(tt.input)
			if len(matches) != tt.expectedMatches {
				t.Errorf("Expected %d matches, got %d in text: %s",
					tt.expectedMatches, len(matches), tt.input)
				for i, match := range matches {
					t.Logf("Match %d: %s = %.2f (confidence: %.2f)",
						i, match.Unit, match.Value, match.Confidence)
				}
			}

			// Test conversion
			if tt.shouldConvert && len(matches) > 0 {
				for _, match := range matches {
					result, err := conv.Convert(match)
					if err != nil {
						t.Errorf("Conversion failed for %s: %v", match.Unit, err)
						continue
					}

					// Basic sanity checks
					if result.MetricUnit == "" {
						t.Errorf("Empty metric unit for %s", match.Unit)
					}
					if result.Formatted == "" {
						t.Errorf("Empty formatted result for %s", match.Unit)
					}
					if result.MetricValue < 0 && match.Value >= 0 {
						t.Errorf("Unexpected negative conversion: %f -> %f",
							match.Value, result.MetricValue)
					}
				}
			}
		})
	}
}

// TestUnitConversion_IntegrationEdgeCases tests edge cases and error conditions
func TestUnitConversion_IntegrationEdgeCases(t *testing.T) {
	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	tests := []struct {
		name        string
		input       string
		shouldWork  bool
		description string
	}{
		{
			name:        "empty_string",
			input:       "",
			shouldWork:  true, // Should work but find no units
			description: "Empty string should not crash",
		},
		{
			name:        "no_units",
			input:       "This text has no measurements at all",
			shouldWork:  true, // Should work but find no units
			description: "Text without units should work fine",
		},
		{
			name:        "very_large_numbers",
			input:       "Distance is 1000000 miles",
			shouldWork:  true,
			description: "Very large numbers should be handled",
		},
		{
			name:        "very_small_numbers",
			input:       "Gap is 0.001 inches",
			shouldWork:  true,
			description: "Very small numbers should be handled",
		},
		{
			name:        "zero_values",
			input:       "Height is 0 feet",
			shouldWork:  true,
			description: "Zero values should be handled",
		},
		{
			name:        "negative_temperature",
			input:       "Temperature dropped to -10°F",
			shouldWork:  true,
			description: "Negative temperatures should work",
		},
		{
			name:        "mixed_valid_invalid",
			input:       "Room is 12 feet wide and has nice color",
			shouldWork:  true,
			description: "Mix of valid units and regular text",
		},
		{
			name:        "special_characters",
			input:       "Measure 5.5 feet (exactly) for the 10-inch gap!",
			shouldWork:  true,
			description: "Special characters and punctuation",
		},
		{
			name:        "unicode_text",
			input:       "The café is 20 feet from the résumé shop",
			shouldWork:  true,
			description: "Unicode characters should not interfere",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test detection
			matches := detector.DetectUnits(tt.input)

			if !tt.shouldWork {
				// If we expect it to fail, we can't test much more
				return
			}

			// Test conversion for any detected units
			for _, match := range matches {
				result, err := conv.Convert(match)
				if err != nil {
					t.Errorf("Conversion failed for %s in '%s': %v",
						match.Unit, tt.input, err)
					continue
				}

				// Basic validation
				if result.MetricUnit == "" {
					t.Errorf("Empty metric unit for %s", match.Unit)
				}
				if result.Formatted == "" {
					t.Errorf("Empty formatted result for %s", match.Unit)
				}
			}
		})
	}
}

// TestUnitConversion_PerformanceBaseline provides basic performance testing
func TestUnitConversion_PerformanceBaseline(t *testing.T) {
	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	// Test with moderately large text
	largeText := strings.Repeat("The room is 12 feet wide and 8 feet tall. Temperature is 72°F. ", 100)

	t.Run("large_text_detection", func(t *testing.T) {
		matches := detector.DetectUnits(largeText)
		if len(matches) == 0 {
			t.Error("Expected to find units in large text")
		}

		// Should find units without taking too long
		// This is just a baseline - no strict timing requirements
		t.Logf("Detected %d units in text of %d characters", len(matches), len(largeText))
	})

	t.Run("large_text_conversion", func(t *testing.T) {
		matches := detector.DetectUnits(largeText)

		convertedCount := 0
		for _, match := range matches {
			_, err := conv.Convert(match)
			if err == nil {
				convertedCount++
			}
		}

		if convertedCount == 0 {
			t.Error("Expected to convert some units")
		}

		t.Logf("Successfully converted %d out of %d detected units", convertedCount, len(matches))
	})
}

// TestUnitConversion_ConfigurationIntegration tests configuration integration
func TestUnitConversion_ConfigurationIntegration(t *testing.T) {
	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	// Test with different configuration settings
	tests := []struct {
		name        string
		preferences converter.ConversionPreferences
		input       string
		expected    string
	}{
		{
			name: "no_spaces_preference",
			preferences: converter.ConversionPreferences{
				PreferWholeNumbers:          true,
				MaxDecimalPlaces:            2,
				UseLocalizedUnits:           true,
				TemperatureFormat:           "°C",
				UseSpaceBetweenValueAndUnit: false,
				RoundingThreshold:           0.05,
			},
			input:    "Temperature is 32°F",
			expected: "0°C",
		},
		{
			name: "custom_temperature_format",
			preferences: converter.ConversionPreferences{
				PreferWholeNumbers:          true,
				MaxDecimalPlaces:            2,
				UseLocalizedUnits:           true,
				TemperatureFormat:           "degrees Celsius",
				UseSpaceBetweenValueAndUnit: false,
				RoundingThreshold:           0.05,
			},
			input:    "Heat to 212°F",
			expected: "100 degrees Celsius",
		},
		{
			name: "high_precision",
			preferences: converter.ConversionPreferences{
				PreferWholeNumbers:          false,
				MaxDecimalPlaces:            3,
				UseLocalizedUnits:           true,
				TemperatureFormat:           "°C",
				UseSpaceBetweenValueAndUnit: true,
				RoundingThreshold:           0.01,
			},
			input:    "Length is 1 foot",
			expected: "30.5 cm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv.SetPreferences(tt.preferences)

			matches := detector.DetectUnits(tt.input)
			if len(matches) != 1 {
				t.Fatalf("Expected 1 match, got %d", len(matches))
			}

			result, err := conv.Convert(matches[0])
			if err != nil {
				t.Errorf("Conversion failed: %v", err)
				return
			}

			if result.Formatted != tt.expected {
				t.Errorf("Expected formatted result '%s', got '%s'",
					tt.expected, result.Formatted)
			}
		})
	}
}

// TestUnitConversion_ConsistencyCheck tests mathematical consistency
func TestUnitConversion_ConsistencyCheck(t *testing.T) {
	conv := converter.NewBasicUnitConverter()

	// Test that related conversions are mathematically consistent
	tests := []struct {
		name        string
		conversions []converter.UnitMatch
		description string
	}{
		{
			name: "12_inches_equals_1_foot",
			conversions: []converter.UnitMatch{
				{Value: 12.0, Unit: "inches", UnitType: converter.Length, Confidence: 0.9},
				{Value: 1.0, Unit: "feet", UnitType: converter.Length, Confidence: 0.9},
			},
			description: "12 inches and 1 foot should convert to same metric value",
		},
		{
			name: "16_ounces_equals_1_pound",
			conversions: []converter.UnitMatch{
				{Value: 16.0, Unit: "ounces", UnitType: converter.Mass, Confidence: 0.9},
				{Value: 1.0, Unit: "pounds", UnitType: converter.Mass, Confidence: 0.9},
			},
			description: "16 ounces and 1 pound should convert to same metric value",
		},
		{
			name: "4_quarts_equals_1_gallon",
			conversions: []converter.UnitMatch{
				{Value: 4.0, Unit: "quarts", UnitType: converter.Volume, Confidence: 0.9},
				{Value: 1.0, Unit: "gallons", UnitType: converter.Volume, Confidence: 0.9},
			},
			description: "4 quarts and 1 gallon should convert to same metric value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.conversions) != 2 {
				t.Fatalf("Test requires exactly 2 conversions")
			}

			result1, err1 := conv.Convert(tt.conversions[0])
			if err1 != nil {
				t.Errorf("First conversion error = %v", err1)
				return
			}

			result2, err2 := conv.Convert(tt.conversions[1])
			if err2 != nil {
				t.Errorf("Second conversion error = %v", err2)
				return
			}

			// Convert both to the same base unit for comparison
			var value1, value2 float64

			// Normalize to base metric units
			switch tt.conversions[0].UnitType {
			case converter.Length:
				// Convert to metres
				value1 = normalizeToMetres(result1.MetricValue, result1.MetricUnit)
				value2 = normalizeToMetres(result2.MetricValue, result2.MetricUnit)
			case converter.Mass:
				// Convert to kilograms
				value1 = normalizeToKilograms(result1.MetricValue, result1.MetricUnit)
				value2 = normalizeToKilograms(result2.MetricValue, result2.MetricUnit)
			case converter.Volume:
				// Convert to litres
				value1 = normalizeToLitres(result1.MetricValue, result1.MetricUnit)
				value2 = normalizeToLitres(result2.MetricValue, result2.MetricUnit)
			default:
				value1 = result1.MetricValue
				value2 = result2.MetricValue
			}

			// Check that the values are equal (within tolerance)
			tolerance := 0.001
			diff := value1 - value2
			if diff < 0 {
				diff = -diff
			}

			if diff > tolerance {
				t.Errorf("Consistency check failed: %.6f vs %.6f (tolerance %.6f) - %s",
					value1, value2, tolerance, tt.description)
			}
		})
	}
}

// Helper functions for consistency checking
func normalizeToMetres(value float64, unit string) float64 {
	switch unit {
	case "mm":
		return value / 1000
	case "cm":
		return value / 100
	case "metres":
		return value
	case "km":
		return value * 1000
	default:
		return value
	}
}

func normalizeToKilograms(value float64, unit string) float64 {
	switch unit {
	case "mg":
		return value / 1000000
	case "g":
		return value / 1000
	case "kg":
		return value
	case "tonnes":
		return value * 1000
	default:
		return value
	}
}

func normalizeToLitres(value float64, unit string) float64 {
	switch unit {
	case "ml":
		return value / 1000
	case "litres":
		return value
	default:
		return value
	}
}
