package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sammcj/m2e/pkg/converter"
)

type ConvertRequest struct {
	Text                 string `json:"text"`
	ConvertUnits         *bool  `json:"convert_units,omitempty"`
	NormaliseSmartQuotes *bool  `json:"normalise_smart_quotes,omitempty"`
}

type ConvertResponse struct {
	Text string `json:"text"`
}

func main() {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/v1/health", healthHandler)
	http.HandleFunc("/api/v1/convert", convertHandler)

	log.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, "OK"); err != nil {
		log.Printf("Error writing health response: %v", err)
	}
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	conv, err := converter.NewConverter()
	if err != nil {
		http.Error(w, "Error initializing converter", http.StatusInternalServerError)
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
	conv.SetUnitProcessingEnabled(convertUnits)

	convertedText := conv.ConvertToBritish(req.Text, normaliseSmartQuotes)

	resp := ConvertResponse{Text: convertedText}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
