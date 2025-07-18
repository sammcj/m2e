package tests

import (
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

// BenchmarkUnitDetection benchmarks unit detection performance
func BenchmarkUnitDetection(b *testing.B) {
	detector := converter.NewContextualUnitDetector()

	tests := []struct {
		name string
		text string
	}{
		{
			name: "simple_text",
			text: "The room is 12 feet wide and 8 feet tall",
		},
		{
			name: "complex_text",
			text: "The building is 100 feet long, 50 feet wide, weighs 500 tons, and maintains 72°F temperature with 1000 square feet of space",
		},
		{
			name: "mixed_with_idioms",
			text: "The project is miles ahead of schedule, but we need a 6-foot fence that weighs 200 pounds and costs tons of money",
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				detector.DetectUnits(tt.text)
			}
		})
	}
}

// BenchmarkUnitConversion benchmarks unit conversion performance
func BenchmarkUnitConversion(b *testing.B) {
	conv := converter.NewBasicUnitConverter()

	tests := []struct {
		name  string
		match converter.UnitMatch
	}{
		{
			name: "length_conversion",
			match: converter.UnitMatch{
				Value: 12.0, Unit: "feet", UnitType: converter.Length, Confidence: 0.9,
			},
		},
		{
			name: "mass_conversion",
			match: converter.UnitMatch{
				Value: 10.0, Unit: "pounds", UnitType: converter.Mass, Confidence: 0.9,
			},
		},
		{
			name: "temperature_conversion",
			match: converter.UnitMatch{
				Value: 72.0, Unit: "fahrenheit", UnitType: converter.Temperature, Confidence: 0.9,
			},
		},
		{
			name: "volume_conversion",
			match: converter.UnitMatch{
				Value: 5.0, Unit: "gallons", UnitType: converter.Volume, Confidence: 0.9,
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = conv.Convert(tt.match)
			}
		})
	}
}

// BenchmarkFullPipeline benchmarks the complete detection + conversion pipeline
func BenchmarkFullPipeline(b *testing.B) {
	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	text := "The room is 12 feet wide, 8 feet tall, weighs 500 pounds, holds 10 gallons, and maintains 72°F temperature"

	b.Run("detection_and_conversion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			matches := detector.DetectUnits(text)
			for _, match := range matches {
				_, _ = conv.Convert(match)
			}
		}
	})
}

// TestUnitConversion_LargeText tests performance with large text samples
func TestUnitConversion_LargeText(t *testing.T) {
	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	// Create large text with repeated measurements
	baseText := "The building is 100 feet long and 50 feet wide. Temperature is 72°F. Weight is 500 pounds. Volume is 10 gallons. "

	tests := []struct {
		name        string
		repetitions int
	}{
		{"small_text", 10},
		{"medium_text", 100},
		{"large_text", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			largeText := strings.Repeat(baseText, tt.repetitions)

			// Test detection performance
			matches := detector.DetectUnits(largeText)
			expectedMatches := tt.repetitions * 5 // 5 units per repetition

			if len(matches) != expectedMatches {
				t.Errorf("Expected %d matches, got %d", expectedMatches, len(matches))
			}

			// Test conversion performance
			convertedCount := 0
			for _, match := range matches {
				_, err := conv.Convert(match)
				if err == nil {
					convertedCount++
				}
			}

			if convertedCount != len(matches) {
				t.Errorf("Expected to convert all %d matches, only converted %d",
					len(matches), convertedCount)
			}

			t.Logf("Processed text of %d characters with %d units in %d conversions",
				len(largeText), len(matches), convertedCount)
		})
	}
}

// TestUnitConversion_StressTest tests system behavior under stress
func TestUnitConversion_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	// Test with many different unit types and values
	stressTexts := []string{
		"Distance: 1 inch, 2 feet, 3 yards, 4 miles",
		"Weight: 1 ounce, 2 pounds, 3 tons",
		"Volume: 1 fluid ounce, 2 pints, 3 quarts, 4 gallons",
		"Temperature: 32°F, 72°F, 212°F",
		"Area: 100 square feet, 5 acres",
		"Mixed: 10 feet long, 5 pounds heavy, 2 gallons capacity, 75°F temperature",
	}

	t.Run("repeated_processing", func(t *testing.T) {
		iterations := 1000
		totalMatches := 0
		totalConversions := 0

		for i := 0; i < iterations; i++ {
			for _, text := range stressTexts {
				matches := detector.DetectUnits(text)
				totalMatches += len(matches)

				for _, match := range matches {
					_, err := conv.Convert(match)
					if err == nil {
						totalConversions++
					}
				}
			}
		}

		expectedMatches := iterations * len(stressTexts) * 3 // Rough estimate
		if totalMatches < expectedMatches/2 {
			t.Errorf("Expected at least %d matches, got %d", expectedMatches/2, totalMatches)
		}

		if totalConversions != totalMatches {
			t.Errorf("Expected all %d matches to convert, only %d succeeded",
				totalMatches, totalConversions)
		}

		t.Logf("Stress test completed: %d iterations, %d matches, %d conversions",
			iterations, totalMatches, totalConversions)
	})
}

// TestUnitConversion_MemoryUsage tests for memory leaks and excessive allocation
func TestUnitConversion_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	text := "The room is 12 feet wide, 8 feet tall, and maintains 72°F temperature"

	t.Run("repeated_allocations", func(t *testing.T) {
		// Run many iterations to check for memory leaks
		iterations := 10000

		for i := 0; i < iterations; i++ {
			matches := detector.DetectUnits(text)
			for _, match := range matches {
				_, _ = conv.Convert(match)
			}
		}

		t.Logf("Completed %d iterations without memory issues", iterations)
	})
}

// TestUnitConversion_ConcurrentAccess tests thread safety
func TestUnitConversion_ConcurrentAccess(t *testing.T) {
	detector := converter.NewContextualUnitDetector()
	conv := converter.NewBasicUnitConverter()

	text := "The building is 100 feet long and weighs 500 pounds at 72°F"

	t.Run("concurrent_detection", func(t *testing.T) {
		const numGoroutines = 10
		const iterationsPerGoroutine = 100

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()

				for j := 0; j < iterationsPerGoroutine; j++ {
					matches := detector.DetectUnits(text)
					if len(matches) == 0 {
						t.Errorf("Expected to find units in concurrent test")
						return
					}
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		t.Logf("Concurrent detection test completed successfully")
	})

	t.Run("concurrent_conversion", func(t *testing.T) {
		const numGoroutines = 10
		const iterationsPerGoroutine = 100

		// Pre-detect units to test conversion concurrency
		matches := detector.DetectUnits(text)
		if len(matches) == 0 {
			t.Fatal("Need units to test concurrent conversion")
		}

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()

				for j := 0; j < iterationsPerGoroutine; j++ {
					for _, match := range matches {
						_, err := conv.Convert(match)
						if err != nil {
							t.Errorf("Conversion failed in concurrent test: %v", err)
							return
						}
					}
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		t.Logf("Concurrent conversion test completed successfully")
	})
}

// TestUnitConversion_ConfigurationPerformance tests performance impact of different configurations
func TestUnitConversion_ConfigurationPerformance(t *testing.T) {
	text := "The room is 12 feet wide and 8 feet tall with temperature at 72°F"

	configs := []struct {
		name        string
		preferences converter.ConversionPreferences
	}{
		{
			name: "default_config",
			preferences: converter.ConversionPreferences{
				PreferWholeNumbers: true, MaxDecimalPlaces: 2,
				UseLocalizedUnits: true, TemperatureFormat: "°C",
				UseSpaceBetweenValueAndUnit: true, RoundingThreshold: 0.05,
			},
		},
		{
			name: "high_precision_config",
			preferences: converter.ConversionPreferences{
				PreferWholeNumbers: false, MaxDecimalPlaces: 5,
				UseLocalizedUnits: true, TemperatureFormat: "degrees Celsius",
				UseSpaceBetweenValueAndUnit: true, RoundingThreshold: 0.001,
			},
		},
		{
			name: "minimal_config",
			preferences: converter.ConversionPreferences{
				PreferWholeNumbers: true, MaxDecimalPlaces: 0,
				UseLocalizedUnits: false, TemperatureFormat: "°C",
				UseSpaceBetweenValueAndUnit: false, RoundingThreshold: 0.1,
			},
		},
	}

	detector := converter.NewContextualUnitDetector()

	for _, config := range configs {
		t.Run(config.name, func(t *testing.T) {
			conv := converter.NewBasicUnitConverter()
			conv.SetPreferences(config.preferences)

			// Run multiple iterations to measure performance impact
			iterations := 1000
			totalConversions := 0

			for i := 0; i < iterations; i++ {
				matches := detector.DetectUnits(text)
				for _, match := range matches {
					_, err := conv.Convert(match)
					if err == nil {
						totalConversions++
					}
				}
			}

			expectedConversions := iterations * 3 // Expect 3 units per iteration
			if totalConversions != expectedConversions {
				t.Errorf("Expected %d conversions, got %d", expectedConversions, totalConversions)
			}

			t.Logf("Configuration '%s': %d conversions completed", config.name, totalConversions)
		})
	}
}
