package converter

import (
	"github.com/neurosnap/sentences"
	"github.com/neurosnap/sentences/english"
	"strings"
)

// SentenceAwareConverter enhances text conversion by processing text sentence by sentence
// This provides better context boundaries for contextual word detection
type SentenceAwareConverter struct {
	converter *Converter
	tokenizer sentences.SentenceTokenizer
}

// NewSentenceAwareConverter creates a new sentence-aware converter
func NewSentenceAwareConverter() (*SentenceAwareConverter, error) {
	// Create base converter
	conv, err := NewConverter()
	if err != nil {
		return nil, err
	}

	// Create sentence tokenizer
	tokenizer, err := english.NewSentenceTokenizer(nil)
	if err != nil {
		return nil, err
	}

	return &SentenceAwareConverter{
		converter: conv,
		tokenizer: tokenizer,
	}, nil
}

// ConvertToBritishWithSentenceAwareness converts text using sentence-level contextual analysis
// This provides better accuracy for longer texts with multiple contextual words
func (sac *SentenceAwareConverter) ConvertToBritishWithSentenceAwareness(text string, normaliseSmartQuotes bool) string {
	// For short texts, use the regular converter
	if len(strings.TrimSpace(text)) < 100 {
		return sac.converter.ConvertToBritishSimple(text, normaliseSmartQuotes)
	}

	// Tokenize into sentences
	sentences := sac.tokenizer.Tokenize(text)

	// If only one sentence, use regular converter
	if len(sentences) <= 1 {
		return sac.converter.ConvertToBritishSimple(text, normaliseSmartQuotes)
	}

	// Process each sentence separately
	var result strings.Builder

	for i, sentence := range sentences {
		sentenceText := sentence.Text

		// Convert the sentence
		convertedSentence := sac.converter.ConvertToBritishSimple(sentenceText, normaliseSmartQuotes)

		// Append to result
		result.WriteString(convertedSentence)

		// Add proper spacing between sentences (if not the last sentence)
		if i < len(sentences)-1 {
			// Check if the next sentence starts with a space
			if i+1 < len(sentences) && !strings.HasPrefix(sentences[i+1].Text, " ") {
				// Only add a space if the current sentence doesn't end with whitespace
				// and the next sentence doesn't start with whitespace
				if !strings.HasSuffix(convertedSentence, " ") {
					result.WriteString(" ")
				}
			}
		}
	}

	return result.String()
}

// ConvertToBritishSimple delegates to the underlying converter for backwards compatibility
func (sac *SentenceAwareConverter) ConvertToBritishSimple(text string, normaliseSmartQuotes bool) string {
	return sac.converter.ConvertToBritishSimple(text, normaliseSmartQuotes)
}

// GetConverter returns the underlying converter for advanced usage
func (sac *SentenceAwareConverter) GetConverter() *Converter {
	return sac.converter
}

// IsContextualWordDetectionEnabled checks if contextual word detection is enabled
func (sac *SentenceAwareConverter) IsContextualWordDetectionEnabled() bool {
	return sac.converter.IsContextualWordDetectionEnabled()
}

// SetContextualWordDetectionEnabled enables or disables contextual word detection
func (sac *SentenceAwareConverter) SetContextualWordDetectionEnabled(enabled bool) {
	sac.converter.SetContextualWordDetectionEnabled(enabled)
}

// GetContextualWordDetector returns the contextual word detector
func (sac *SentenceAwareConverter) GetContextualWordDetector() ContextualWordDetector {
	return sac.converter.GetContextualWordDetector()
}
