package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

// MockMCPRequest simulates an MCP CallToolRequest for testing
type MockMCPRequest struct {
	params map[string]string
}

func NewMockMCPRequest() *MockMCPRequest {
	return &MockMCPRequest{
		params: make(map[string]string),
	}
}

func (r *MockMCPRequest) SetString(key, value string) {
	r.params[key] = value
}

func (r *MockMCPRequest) RequireString(key string) (string, error) {
	if val, ok := r.params[key]; ok {
		return val, nil
	}
	return "", fmt.Errorf("parameter %s not found", key)
}

// MockMCPServer simulates the MCP server functionality for testing
type MockMCPServer struct {
	converter *converter.Converter
}

func NewMockMCPServer() *MockMCPServer {
	conv, _ := converter.NewConverter()
	return &MockMCPServer{
		converter: conv,
	}
}

// ConvertText simulates the MCP convert_text tool
func (s *MockMCPServer) ConvertText(req *MockMCPRequest) (string, error) {
	text, err := req.RequireString("text")
	if err != nil {
		return "", err
	}

	// Get optional parameters with defaults
	convertUnits := false
	if val, err := req.RequireString("convert_units"); err == nil {
		convertUnits = strings.ToLower(val) == "true"
	}

	normaliseSmartQuotes := true
	if val, err := req.RequireString("normalise_smart_quotes"); err == nil {
		normaliseSmartQuotes = strings.ToLower(val) != "false"
	}

	// Set unit processing based on parameter
	s.converter.SetUnitProcessingEnabled(convertUnits)

	convertedText := s.converter.ConvertToBritish(text, normaliseSmartQuotes)
	return convertedText, nil
}

// ConvertFile simulates the MCP convert_file tool
func (s *MockMCPServer) ConvertFile(req *MockMCPRequest) (string, error) {
	filePath, err := req.RequireString("file_path")
	if err != nil {
		return "", err
	}

	// Get optional parameters with defaults
	convertUnits := false
	if val, err := req.RequireString("convert_units"); err == nil {
		convertUnits = strings.ToLower(val) == "true"
	}

	normaliseSmartQuotes := true
	if val, err := req.RequireString("normalise_smart_quotes"); err == nil {
		normaliseSmartQuotes = strings.ToLower(val) != "false"
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read the original file content
	originalContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	// Set unit processing based on parameter
	s.converter.SetUnitProcessingEnabled(convertUnits)

	// Convert the content based on file type (simplified for testing)
	var convertedContent string
	if strings.HasSuffix(strings.ToLower(filePath), ".txt") || strings.HasSuffix(strings.ToLower(filePath), ".md") {
		// For plain text files, use code-aware processing
		convertedContent = s.converter.ProcessCodeAware(string(originalContent), normaliseSmartQuotes)
	} else {
		// For code files, only convert comments
		convertedContent = s.converter.ConvertToBritish(string(originalContent), normaliseSmartQuotes)
	}

	// Check if there were any changes
	if string(originalContent) == convertedContent {
		return fmt.Sprintf("File %s processed but no changes were needed - already in British English", filePath), nil
	}

	// Write the converted content back to the file
	err = os.WriteFile(filePath, []byte(convertedContent), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing to file %s: %v", filePath, err)
	}

	return fmt.Sprintf("File %s completed processing to international / British English, the file has been updated.", filePath), nil
}

func TestMCPConvertTextWithUnits(t *testing.T) {
	server := NewMockMCPServer()

	tests := []struct {
		name         string
		text         string
		convertUnits string
		expected     string
		description  string
	}{
		{
			name:         "Unit conversion enabled",
			text:         "The room is 12 feet wide and weighs 100 pounds.",
			convertUnits: "true",
			expected:     "The room is 3.7 metres wide and weighs 45.4 kg.",
			description:  "Should convert units when enabled",
		},
		{
			name:         "Unit conversion disabled",
			text:         "The room is 12 feet wide and weighs 100 pounds.",
			convertUnits: "false",
			expected:     "The room is 12 feet wide and weighs 100 pounds.",
			description:  "Should not convert units when disabled",
		},
		{
			name:         "Spelling conversion with units enabled",
			text:         "The color of the 5-foot fence is gray.",
			convertUnits: "true",
			expected:     "The colour of the 1.5-metre fence is grey.",
			description:  "Should convert both spelling and units when enabled",
		},
		{
			name:         "Spelling conversion with units disabled",
			text:         "The color of the 5-foot fence is gray.",
			convertUnits: "false",
			expected:     "The colour of the 5-foot fence is grey.",
			description:  "Should only convert spelling when units disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewMockMCPRequest()
			req.SetString("text", tt.text)
			req.SetString("convert_units", tt.convertUnits)

			result, err := server.ConvertText(req)
			if err != nil {
				t.Fatalf("ConvertText failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("ConvertText() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMCPConvertFileWithUnits(t *testing.T) {
	server := NewMockMCPServer()

	// Create a temporary test file
	testContent := `// The color buffer should be 1024 bytes in size
// Set the width to 100 inches for display
const ROOM_WIDTH_FEET = 12
func convertFeetToMeters() {
    // Width is 100 inches - this is the color
    width := 100
}`

	tempFile, err := os.CreateTemp("", "test_mcp_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }()

	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tempFile.Close()

	tests := []struct {
		name         string
		convertUnits string
		expected     string
		description  string
	}{
		{
			name:         "File conversion with units enabled",
			convertUnits: "true",
			expected:     "completed processing to international / British English",
			description:  "Should convert units in file when enabled",
		},
		{
			name:         "File conversion with units disabled",
			convertUnits: "false",
			expected:     "completed processing to international / British English",
			description:  "Should not convert units in file when disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset file content for each test
			if err := os.WriteFile(tempFile.Name(), []byte(testContent), 0644); err != nil {
				t.Fatalf("Failed to reset file content: %v", err)
			}

			req := NewMockMCPRequest()
			req.SetString("file_path", tempFile.Name())
			req.SetString("convert_units", tt.convertUnits)

			result, err := server.ConvertFile(req)
			if err != nil {
				t.Fatalf("ConvertFile failed: %v", err)
			}

			if !strings.Contains(result, tt.expected) {
				t.Errorf("ConvertFile() = %q, expected to contain %q", result, tt.expected)
			}

			// Verify file was actually modified
			modifiedContent, err := os.ReadFile(tempFile.Name())
			if err != nil {
				t.Fatalf("Failed to read modified file: %v", err)
			}

			// Check that units were converted or not based on the flag
			modifiedStr := string(modifiedContent)
			if tt.convertUnits == "true" {
				// Should contain metric units
				if !strings.Contains(modifiedStr, "254 cm") {
					t.Errorf("File should contain converted units (254 cm) when units enabled, got: %s", modifiedStr)
				}
			} else {
				// Should still contain imperial units
				if !strings.Contains(modifiedStr, "100 inches") {
					t.Errorf("File should contain original units (100 inches) when units disabled, got: %s", modifiedStr)
				}
			}
		})
	}
}

func TestMCPSmartQuotesHandling(t *testing.T) {
	server := NewMockMCPServer()

	tests := []struct {
		name                 string
		text                 string
		normaliseSmartQuotes string
		expected             string
		description          string
	}{
		{
			name:                 "Smart quotes enabled",
			text:                 "The \u201croom\u201d is 10 feet wide.",
			normaliseSmartQuotes: "true",
			expected:             "The \"room\" is 10 feet wide.",
			description:          "Should normalise smart quotes when enabled",
		},
		{
			name:                 "Smart quotes disabled",
			text:                 "The \u201croom\u201d is 10 feet wide.",
			normaliseSmartQuotes: "false",
			expected:             "The \u201croom\u201d is 10 feet wide.",
			description:          "Should preserve smart quotes when disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewMockMCPRequest()
			req.SetString("text", tt.text)
			req.SetString("normalise_smart_quotes", tt.normaliseSmartQuotes)

			result, err := server.ConvertText(req)
			if err != nil {
				t.Fatalf("ConvertText failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("ConvertText() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMCPParameterDefaults(t *testing.T) {
	server := NewMockMCPServer()

	// Test with no optional parameters (should use defaults)
	req := NewMockMCPRequest()
	req.SetString("text", "The color of the 5-foot fence is gray.")

	result, err := server.ConvertText(req)
	if err != nil {
		t.Fatalf("ConvertText failed: %v", err)
	}

	// Should only convert spelling (units disabled by default)
	expected := "The colour of the 5-foot fence is grey."
	if result != expected {
		t.Errorf("ConvertText() with defaults = %q, expected %q", result, expected)
	}
}
