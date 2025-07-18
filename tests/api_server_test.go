package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

// ConvertRequest matches the API server request structure
type ConvertRequest struct {
	Text                 string `json:"text"`
	ConvertUnits         *bool  `json:"convert_units,omitempty"`
	NormaliseSmartQuotes *bool  `json:"normalise_smart_quotes,omitempty"`
}

// ConvertResponse matches the API server response structure
type ConvertResponse struct {
	Text string `json:"text"`
}

// MockAPIServer simulates the HTTP API server for testing
type MockAPIServer struct {
	converter *converter.Converter
}

func NewMockAPIServer() *MockAPIServer {
	conv, _ := converter.NewConverter()
	return &MockAPIServer{
		converter: conv,
	}
}

// convertHandler simulates the API server convert handler
func (s *MockAPIServer) convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// Get optional parameters with defaults
	convertUnits := false
	if req.ConvertUnits != nil {
		convertUnits = *req.ConvertUnits
	}

	normaliseSmartQuotes := true
	if req.NormaliseSmartQuotes != nil {
		normaliseSmartQuotes = *req.NormaliseSmartQuotes
	}

	// Set unit processing based on parameter
	s.converter.SetUnitProcessingEnabled(convertUnits)

	convertedText := s.converter.ConvertToBritish(req.Text, normaliseSmartQuotes)

	resp := ConvertResponse{Text: convertedText}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func TestAPIServerUnitConversion(t *testing.T) {
	server := NewMockAPIServer()

	tests := []struct {
		name        string
		request     ConvertRequest
		expected    string
		description string
	}{
		{
			name: "Unit conversion enabled",
			request: ConvertRequest{
				Text:         "The room is 12 feet wide and weighs 100 pounds.",
				ConvertUnits: boolPtr(true),
			},
			expected:    "The room is 3.7 metres wide and weighs 45.4 kg.",
			description: "Should convert units when enabled",
		},
		{
			name: "Unit conversion disabled",
			request: ConvertRequest{
				Text:         "The room is 12 feet wide and weighs 100 pounds.",
				ConvertUnits: boolPtr(false),
			},
			expected:    "The room is 12 feet wide and weighs 100 pounds.",
			description: "Should not convert units when disabled",
		},
		{
			name: "Spelling and unit conversion",
			request: ConvertRequest{
				Text:         "The color of the 5-foot fence is gray.",
				ConvertUnits: boolPtr(true),
			},
			expected:    "The colour of the 1.5-metre fence is grey.",
			description: "Should convert both spelling and units when enabled",
		},
		{
			name: "Default parameters (no unit conversion)",
			request: ConvertRequest{
				Text: "The color of the 5-foot fence is gray.",
			},
			expected:    "The colour of the 5-foot fence is grey.",
			description: "Should only convert spelling when units not specified (default false)",
		},
		{
			name: "Smart quotes disabled",
			request: ConvertRequest{
				Text:                 "The \u201croom\u201d is 10 feet wide.",
				ConvertUnits:         boolPtr(true),
				NormaliseSmartQuotes: boolPtr(false),
			},
			expected:    "The \u201croom\u201d is 3 metres wide.",
			description: "Should preserve smart quotes when disabled",
		},
		{
			name: "Smart quotes enabled (default)",
			request: ConvertRequest{
				Text:         "The \u201croom\u201d is 10 feet wide.",
				ConvertUnits: boolPtr(true),
			},
			expected:    "The \"room\" is 3 metres wide.",
			description: "Should normalise smart quotes when enabled (default)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			reqBody, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/convert", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler
			server.convertHandler(w, req)

			// Check status code
			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
				return
			}

			// Parse response
			var resp ConvertResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Check result
			if resp.Text != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, resp.Text)
			}
		})
	}
}

func TestAPIServerErrorHandling(t *testing.T) {
	server := NewMockAPIServer()

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			body:           `{"text": "test"}`,
			expectedStatus: http.StatusMethodNotAllowed,
			description:    "Should reject non-POST requests",
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           `{"text": "test"`,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject malformed JSON",
		},
		{
			name:           "Valid request",
			method:         http.MethodPost,
			body:           `{"text": "color"}`,
			expectedStatus: http.StatusOK,
			description:    "Should accept valid requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/convert", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.convertHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestAPIServerJSONStructure(t *testing.T) {
	server := NewMockAPIServer()

	// Test that the API accepts and returns proper JSON structure
	request := ConvertRequest{
		Text:                 "The color is gray and the room is 12 feet wide.",
		ConvertUnits:         boolPtr(true),
		NormaliseSmartQuotes: boolPtr(true),
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/convert", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.convertHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify response structure
	var resp ConvertResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Verify response contains converted text
	expected := "The colour is grey and the room is 3.7 metres wide."
	if resp.Text != expected {
		t.Errorf("Expected %q, got %q", expected, resp.Text)
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}

func TestAPIServerIntegration(t *testing.T) {
	// This test simulates a full integration test with the actual server
	server := NewMockAPIServer()

	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(server.convertHandler))
	defer ts.Close()

	// Test data
	testCases := []struct {
		name     string
		request  ConvertRequest
		expected string
	}{
		{
			name: "Full integration test",
			request: ConvertRequest{
				Text:                 "The color of the 10-foot fence is gray.",
				ConvertUnits:         boolPtr(true),
				NormaliseSmartQuotes: boolPtr(true),
			},
			expected: "The colour of the 3.0-metre fence is grey.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(reqBody))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}

			var response ConvertResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Text != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, response.Text)
			}
		})
	}
}
