package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/sammcj/m2e/pkg/converter"
)

type ConvertRequest struct {
	Text                 string `json:"text"`
	ConvertUnits         *bool  `json:"convert_units,omitempty"`
	NormaliseSmartQuotes *bool  `json:"normalise_smart_quotes,omitempty"`
}

type ConvertResponse struct {
	Text    string       `json:"text"`
	Changes []ChangeInfo `json:"changes,omitempty"`
}

type ChangeInfo struct {
	Position     int    `json:"position"`
	Original     string `json:"original"`
	Converted    string `json:"converted"`
	Type         string `json:"type"` // "spelling" or "unit"
	IsContextual bool   `json:"is_contextual,omitempty"`
}

func main() {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	conv, err := converter.NewConverter()
	if err != nil {
		log.Fatalf("Failed to create converter: %v", err)
	}

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}

	http.HandleFunc("/api/v1/health", withCORS(healthHandler, corsOrigin))
	http.HandleFunc("/api/v1/convert", withCORS(makeConvertHandler(conv), corsOrigin))

	log.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// withCORS wraps a handler with CORS headers.
func withCORS(next http.HandlerFunc, origin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "OK"); err != nil {
		log.Printf("Error writing health response: %v", err)
	}
}

// generateChanges analyzes the differences between original and converted text
func generateChanges(originalText, convertedText string, conv *converter.Converter) []ChangeInfo {
	var changes []ChangeInfo

	if originalText == convertedText {
		return changes
	}

	// Get the list of contextual words for comparison
	contextualWords := conv.GetContextualWordDetector().SupportedWords()
	contextualWordSet := make(map[string]bool)
	for _, word := range contextualWords {
		contextualWordSet[strings.ToLower(word)] = true
	}

	// Simple word-by-word comparison to find changes
	originalWords := strings.FieldsFunc(originalText, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '\''
	})
	convertedWords := strings.FieldsFunc(convertedText, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '\''
	})

	// Find position in original text and match words
	originalPos := 0

	for i := 0; i < len(originalWords) && i < len(convertedWords); i++ {
		originalWord := originalWords[i]
		convertedWord := convertedWords[i]

		// Find the actual position in the original text
		wordStart := strings.Index(originalText[originalPos:], originalWord)
		if wordStart == -1 {
			originalPos += len(originalWord) + 1 // Approximate advance
			continue
		}
		actualPos := originalPos + wordStart

		if originalWord != convertedWord {
			// Determine if this is a contextual word change
			isContextual := contextualWordSet[strings.ToLower(originalWord)] ||
				contextualWordSet[strings.ToLower(convertedWord)]

			// Simple heuristic: if contains numbers, likely unit conversion
			changeType := "spelling"
			if strings.ContainsAny(originalWord, "0123456789") || strings.ContainsAny(convertedWord, "0123456789") {
				changeType = "unit"
				isContextual = false // Unit changes are not contextual spelling
			}

			changes = append(changes, ChangeInfo{
				Position:     actualPos,
				Original:     originalWord,
				Converted:    convertedWord,
				Type:         changeType,
				IsContextual: isContextual,
			})
		}

		originalPos = actualPos + len(originalWord)
	}

	return changes
}

func makeConvertHandler(conv *converter.Converter) http.HandlerFunc {
	var mu sync.Mutex
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Validate Content-Type
		ct := r.Header.Get("Content-Type")
		if ct != "" && !strings.HasPrefix(ct, "application/json") {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		// Limit request body to 10 MB to prevent abuse
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		defer func() { _ = r.Body.Close() }()

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

		// Mutex protects shared converter state from concurrent requests
		mu.Lock()
		conv.SetUnitProcessingEnabled(convertUnits)
		originalText := req.Text
		convertedText := conv.ConvertToBritish(req.Text, normaliseSmartQuotes)
		mu.Unlock()

		// Generate change information
		changes := generateChanges(originalText, convertedText, conv)

		resp := ConvertResponse{
			Text:    convertedText,
			Changes: changes,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}
