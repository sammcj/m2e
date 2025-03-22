// Package converter provides functionality to convert between American and British English spellings
package converter

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

//go:embed data/*.json
var dictFS embed.FS

// Dictionaries holds the mapping between American and British English spellings
type Dictionaries struct {
	AmericanToBritish map[string]string
	BritishToAmerican map[string]string
}

// LoadDictionaries loads the spelling dictionaries from the embedded JSON files
func LoadDictionaries() (*Dictionaries, error) {
	// Load American to British dictionary
	amToBrData, err := dictFS.ReadFile("data/american_spellings.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read American spellings dictionary: %w", err)
	}

	// Load British to American dictionary
	brToAmData, err := dictFS.ReadFile("data/british_spellings.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read British spellings dictionary: %w", err)
	}

	// Parse the dictionaries
	amToBr := make(map[string]string)
	brToAm := make(map[string]string)

	if err := json.Unmarshal(amToBrData, &amToBr); err != nil {
		return nil, fmt.Errorf("failed to parse American spellings dictionary: %w", err)
	}

	if err := json.Unmarshal(brToAmData, &brToAm); err != nil {
		return nil, fmt.Errorf("failed to parse British spellings dictionary: %w", err)
	}

	return &Dictionaries{
		AmericanToBritish: amToBr,
		BritishToAmerican: brToAm,
	}, nil
}

// Converter provides methods to convert between American and British English
type Converter struct {
	dict *Dictionaries
}

// SmartQuotesMap holds mappings for smart quotes and em-dashes to their normal equivalents
var SmartQuotesMap = map[string]string{
	"\u201C": "\"", // Left double quote to normal double quote
	"\u201D": "\"", // Right double quote to normal double quote
	"\u2018": "'",  // Left single quote to normal single quote
	"\u2019": "'",  // Right single quote to normal single quote
	"\u2013": "-",  // En-dash to hyphen
	"\u2014": "--", // Em-dash to double hyphen
}

// NewConverter creates a new Converter instance
func NewConverter() (*Converter, error) {
	dict, err := LoadDictionaries()
	if err != nil {
		return nil, err
	}

	return &Converter{
		dict: dict,
	}, nil
}

// ConvertToBritish converts American English text to British English
func (c *Converter) ConvertToBritish(text string, normaliseSmartQuotes bool) string {
	result := c.convert(text, c.dict.AmericanToBritish)
	if normaliseSmartQuotes {
		result = c.normaliseSmartQuotes(result)
	}
	return result
}

// ConvertToAmerican converts British English text to American English
func (c *Converter) ConvertToAmerican(text string, normaliseSmartQuotes bool) string {
	result := c.convert(text, c.dict.BritishToAmerican)
	if normaliseSmartQuotes {
		result = c.normaliseSmartQuotes(result)
	}
	return result
}

// NormaliseSmartQuotes converts smart quotes and em-dashes to their normal equivalents
func (c *Converter) normaliseSmartQuotes(text string) string {
	result := text
	for smart, normal := range SmartQuotesMap {
		result = strings.ReplaceAll(result, smart, normal)
	}
	return result
}

// convert performs the actual conversion using the provided dictionary
func (c *Converter) convert(text string, dict map[string]string) string {
	// We need to preserve newlines and other whitespace
	// First, split the text into lines
	lines := strings.Split(text, "\n")
	resultLines := make([]string, len(lines))

	for lineIdx, line := range lines {
		if line == "" {
			resultLines[lineIdx] = ""
			continue
		}

		// Process each line separately
		// Split the line into words and whitespace
		var tokens []string
		var isWhitespace []bool

		// Tokenize the line preserving whitespace
		currentToken := ""
		currentIsWhitespace := false

		for _, char := range line {
			isCurrentCharWhitespace := unicode.IsSpace(char)

			// If we're switching between whitespace and non-whitespace, store the current token
			if currentToken != "" && currentIsWhitespace != isCurrentCharWhitespace {
				tokens = append(tokens, currentToken)
				isWhitespace = append(isWhitespace, currentIsWhitespace)
				currentToken = ""
			}

			currentToken += string(char)
			currentIsWhitespace = isCurrentCharWhitespace
		}

		// Add the last token if there is one
		if currentToken != "" {
			tokens = append(tokens, currentToken)
			isWhitespace = append(isWhitespace, currentIsWhitespace)
		}

		// Process each token
		for i := 0; i < len(tokens); i++ {
			// Skip whitespace tokens
			if isWhitespace[i] {
				continue
			}

			word := tokens[i]
			// Check if the word is in the dictionary
			if replacement, ok := dict[strings.ToLower(word)]; ok {
				// Preserve the original case
				if isCapitalized(word) {
					replacement = capitalize(replacement)
				} else if isAllCaps(word) {
					replacement = strings.ToUpper(replacement)
				}
				tokens[i] = replacement
			} else {
				// Check if the word with punctuation is in the dictionary
				cleanWord, punctuation := splitPunctuation(word)
				if replacement, ok := dict[strings.ToLower(cleanWord)]; ok {
					// Preserve the original case
					if isCapitalized(cleanWord) {
						replacement = capitalize(replacement)
					} else if isAllCaps(cleanWord) {
						replacement = strings.ToUpper(replacement)
					}
					tokens[i] = replacement + punctuation
				}
			}
		}

		// Reconstruct the line
		resultLines[lineIdx] = strings.Join(tokens, "")
	}

	// Join the lines back together with newlines
	return strings.Join(resultLines, "\n")
}

// Helper functions for case preservation
func isCapitalized(s string) bool {
	if len(s) == 0 {
		return false
	}
	firstChar := s[0]
	return 'A' <= firstChar && firstChar <= 'Z'
}

func isAllCaps(s string) bool {
	for _, c := range s {
		if 'a' <= c && c <= 'z' {
			return false
		}
	}
	return true
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// splitPunctuation separates a word from its trailing punctuation
func splitPunctuation(word string) (string, string) {
	for i := len(word) - 1; i >= 0; i-- {
		if isLetter(word[i]) || isDigit(word[i]) {
			if i == len(word)-1 {
				return word, ""
			}
			return word[:i+1], word[i+1:]
		}
	}
	return word, ""
}

func isLetter(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
