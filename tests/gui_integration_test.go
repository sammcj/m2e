package tests

import (
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

// MockGUIApp simulates the GUI App struct for testing
type MockGUIApp struct {
	converter *converter.Converter
}

func NewMockGUIApp() *MockGUIApp {
	conv, _ := converter.NewConverter()
	return &MockGUIApp{
		converter: conv,
	}
}

// ConvertToBritish simulates the GUI method
func (a *MockGUIApp) ConvertToBritish(text string, normaliseSmartQuotes bool) string {
	if a.converter == nil {
		return "Error: Converter not initialized"
	}
	return a.converter.ConvertToBritish(text, normaliseSmartQuotes)
}

// ConvertToBritishWithUnits simulates the new GUI method with unit conversion
func (a *MockGUIApp) ConvertToBritishWithUnits(text string, normaliseSmartQuotes bool, convertUnits bool) string {
	if a.converter == nil {
		return "Error: Converter not initialized"
	}

	// Set unit processing enabled/disabled
	a.converter.SetUnitProcessingEnabled(convertUnits)

	return a.converter.ConvertToBritish(text, normaliseSmartQuotes)
}

// GetUnitProcessingStatus simulates the GUI method
func (a *MockGUIApp) GetUnitProcessingStatus() bool {
	if a.converter == nil {
		return false
	}
	return a.converter.GetUnitProcessor().IsEnabled()
}

// SetUnitProcessingEnabled simulates the GUI method
func (a *MockGUIApp) SetUnitProcessingEnabled(enabled bool) {
	if a.converter != nil {
		a.converter.SetUnitProcessingEnabled(enabled)
	}
}

func TestGUIUnitConversionIntegration(t *testing.T) {
	app := NewMockGUIApp()

	tests := []struct {
		name         string
		input        string
		convertUnits bool
		expected     string
		description  string
	}{
		{
			name:         "Unit conversion enabled",
			input:        "The room is 12 feet wide and weighs 100 pounds.",
			convertUnits: true,
			expected:     "The room is 3.7 metres wide and weighs 45.4 kg.",
			description:  "Should convert units when enabled",
		},
		{
			name:         "Unit conversion disabled",
			input:        "The room is 12 feet wide and weighs 100 pounds.",
			convertUnits: false,
			expected:     "The room is 12 feet wide and weighs 100 pounds.",
			description:  "Should not convert units when disabled",
		},
		{
			name:         "Spelling conversion with units enabled",
			input:        "The color of the 5-foot fence is gray.",
			convertUnits: true,
			expected:     "The colour of the 1.5-metre fence is grey.",
			description:  "Should convert both spelling and units when enabled",
		},
		{
			name:         "Spelling conversion with units disabled",
			input:        "The color of the 5-foot fence is gray.",
			convertUnits: false,
			expected:     "The colour of the 5-foot fence is grey.",
			description:  "Should only convert spelling when units disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the ConvertToBritishWithUnits method
			result := app.ConvertToBritishWithUnits(tt.input, true, tt.convertUnits)

			if result != tt.expected {
				t.Errorf("ConvertToBritishWithUnits() = %q, expected %q", result, tt.expected)
			}

			// Test that the unit processing status is correctly set
			app.SetUnitProcessingEnabled(tt.convertUnits)
			status := app.GetUnitProcessingStatus()
			if status != tt.convertUnits {
				t.Errorf("GetUnitProcessingStatus() = %v, expected %v", status, tt.convertUnits)
			}
		})
	}
}

func TestGUIUnitConversionToggle(t *testing.T) {
	app := NewMockGUIApp()

	// Test initial state
	initialStatus := app.GetUnitProcessingStatus()

	// Toggle unit processing
	app.SetUnitProcessingEnabled(!initialStatus)
	newStatus := app.GetUnitProcessingStatus()

	if newStatus == initialStatus {
		t.Errorf("Unit processing status should have changed from %v to %v", initialStatus, !initialStatus)
	}

	// Toggle back
	app.SetUnitProcessingEnabled(initialStatus)
	finalStatus := app.GetUnitProcessingStatus()

	if finalStatus != initialStatus {
		t.Errorf("Unit processing status should have returned to %v, got %v", initialStatus, finalStatus)
	}
}

func TestGUIFileProcessingWithUnits(t *testing.T) {
	app := NewMockGUIApp()

	// Test processing file content with units enabled
	fileContent := `// The buffer should be 1024 bytes in size
// Set the width to 100 inches for display
const ROOM_WIDTH_FEET = 12
func convertFeetToMeters() {
    // Width is 100 inches
    width := 100
}`

	expectedWithUnits := `// The buffer should be 1024 bytes in size
// Set the width to 254 cm for display
const ROOM_WIDTH_FEET = 12
func convertFeetToMeters() {
    // Width is 254 cm
    width := 100
}`

	expectedWithoutUnits := `// The buffer should be 1024 bytes in size
// Set the width to 100 inches for display
const ROOM_WIDTH_FEET = 12
func convertFeetToMeters() {
    // Width is 100 inches
    width := 100
}`

	// Test with units enabled
	resultWithUnits := app.ConvertToBritishWithUnits(fileContent, true, true)
	if resultWithUnits != expectedWithUnits {
		t.Errorf("File processing with units enabled failed.\nExpected:\n%s\nGot:\n%s", expectedWithUnits, resultWithUnits)
	}

	// Test with units disabled
	resultWithoutUnits := app.ConvertToBritishWithUnits(fileContent, true, false)
	if resultWithoutUnits != expectedWithoutUnits {
		t.Errorf("File processing with units disabled failed.\nExpected:\n%s\nGot:\n%s", expectedWithoutUnits, resultWithoutUnits)
	}
}
