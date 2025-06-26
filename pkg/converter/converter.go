// Package converter provides functionality to convert between American and British English spellings
package converter

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

//go:embed data/*.json
var dictFS embed.FS

// Dictionaries holds the mapping for American to British English spellings
type Dictionaries struct {
	AmericanToBritish map[string]string
}

// getUserDictionaryPath returns the path to the user's custom dictionary file
func getUserDictionaryPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "m2e")
	dictPath := filepath.Join(configDir, "american_spellings.json")

	return dictPath, nil
}

// createUserDictionary creates the user dictionary file with an example entry if it doesn't exist
func createUserDictionary(dictPath string) error {
	// Create the directory if it doesn't exist
	configDir := filepath.Dir(dictPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	// Check if file already exists
	if _, err := os.Stat(dictPath); err == nil {
		return nil // File already exists
	}

	// Create the file with an example entry
	exampleDict := map[string]string{
		"customize": "customise",
	}

	data, err := json.MarshalIndent(exampleDict, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal example dictionary: %w", err)
	}

	if err := os.WriteFile(dictPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create user dictionary file %s: %w", dictPath, err)
	}

	return nil
}

// loadUserDictionary loads the user's custom dictionary if it exists
func loadUserDictionary() (map[string]string, error) {
	dictPath, err := getUserDictionaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get user dictionary path: %w", err)
	}

	// Try to create the user dictionary if it doesn't exist
	if err := createUserDictionary(dictPath); err != nil {
		return nil, fmt.Errorf("failed to create user dictionary: %w", err)
	}

	// Read the user dictionary file
	data, err := os.ReadFile(dictPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty dictionary
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to read user dictionary file %s: %w", dictPath, err)
	}

	// Parse the user dictionary
	userDict := make(map[string]string)
	if err := json.Unmarshal(data, &userDict); err != nil {
		return nil, fmt.Errorf("failed to parse user dictionary file %s (please check JSON format): %w", dictPath, err)
	}

	return userDict, nil
}

// LoadDictionaries loads the American to British spelling dictionary from the embedded JSON file
// and merges it with the user's custom dictionary
func LoadDictionaries() (*Dictionaries, error) {
	// Load built-in American to British dictionary
	amToBrData, err := dictFS.ReadFile("data/american_spellings.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read built-in American spellings dictionary: %w", err)
	}

	// Parse the built-in dictionary
	amToBr := make(map[string]string)
	if err := json.Unmarshal(amToBrData, &amToBr); err != nil {
		return nil, fmt.Errorf("failed to parse built-in American spellings dictionary: %w", err)
	}

	// Load user dictionary
	userDict, err := loadUserDictionary()
	if err != nil {
		// Log the error but don't fail completely - just use the built-in dictionary
		fmt.Fprintf(os.Stderr, "Warning: Failed to load user dictionary: %v\n", err)
		userDict = make(map[string]string)
	}

	// Merge user dictionary into built-in dictionary (user entries override built-in ones)
	for american, british := range userDict {
		amToBr[american] = british
	}

	return &Dictionaries{
		AmericanToBritish: amToBr,
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
	"\u2014": "-",  // Em-dash to hyphen
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
	// Use code-aware processing for all text
	return c.ProcessCodeAware(text, normaliseSmartQuotes)
}

// ConvertToBritishSimple converts text without code-awareness (for internal use)
func (c *Converter) ConvertToBritishSimple(text string, normaliseSmartQuotes bool) string {
	// First normalise smart quotes if needed
	processedText := text
	if normaliseSmartQuotes {
		processedText = c.normaliseSmartQuotes(text)
	}

	// Then convert the text
	result := c.convert(processedText, c.dict.AmericanToBritish)
	return result
}

// GetAmericanToBritishDictionary returns the American to British dictionary
func (c *Converter) GetAmericanToBritishDictionary() map[string]string {
	if c.dict == nil {
		return map[string]string{}
	}
	return c.dict.AmericanToBritish
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

			// First, try to match the word as-is
			if replacement, ok := dict[strings.ToLower(word)]; ok {
				// Preserve the original case
				if isCapitalized(word) {
					replacement = capitalize(replacement)
				} else if isAllCaps(word) {
					replacement = strings.ToUpper(replacement)
				}
				tokens[i] = replacement
				continue
			}

			// If that didn't work, try to handle words with quotes
			// First, check for words with double quotes
			if len(word) >= 2 && word[0] == '"' && word[len(word)-1] == '"' {
				innerWord := word[1 : len(word)-1]
				if replacement, ok := dict[strings.ToLower(innerWord)]; ok {
					// Preserve the original case
					if isCapitalized(innerWord) {
						replacement = capitalize(replacement)
					} else if isAllCaps(innerWord) {
						replacement = strings.ToUpper(replacement)
					}
					tokens[i] = "\"" + replacement + "\""
					continue
				}
			}

			// Then, check for words with single quotes
			if len(word) >= 2 && word[0] == '\'' && word[len(word)-1] == '\'' {
				innerWord := word[1 : len(word)-1]
				if replacement, ok := dict[strings.ToLower(innerWord)]; ok {
					// Preserve the original case
					if isCapitalized(innerWord) {
						replacement = capitalize(replacement)
					} else if isAllCaps(innerWord) {
						replacement = strings.ToUpper(replacement)
					}
					tokens[i] = "'" + replacement + "'"
					continue
				}
			}

			// Also check for words with smart quotes
			// This is a more general approach that doesn't rely on specific Unicode characters
			if len(word) >= 2 {
				// Check if the first character is a quote (any kind)
				firstChar := word[0]
				lastChar := word[len(word)-1]

				// Check if the first and last characters are quotes
				// We'll just check for the most common quote characters
				isFirstQuote := firstChar == '\'' || firstChar == '"'
				isLastQuote := lastChar == '\'' || lastChar == '"'

				if isFirstQuote && isLastQuote {
					innerWord := word[1 : len(word)-1]
					if replacement, ok := dict[strings.ToLower(innerWord)]; ok {
						// Preserve the original case
						if isCapitalized(innerWord) {
							replacement = capitalize(replacement)
						} else if isAllCaps(innerWord) {
							replacement = strings.ToUpper(replacement)
						}
						// Preserve the original quote characters
						tokens[i] = string(firstChar) + replacement + string(lastChar)
						continue
					}
				}
			}

			// Check for words with a single quote at the beginning
			if len(word) >= 2 && word[0] == '\'' {
				innerWord := word[1:]
				if replacement, ok := dict[strings.ToLower(innerWord)]; ok {
					// Preserve the original case
					if isCapitalized(innerWord) {
						replacement = capitalize(replacement)
					} else if isAllCaps(innerWord) {
						replacement = strings.ToUpper(replacement)
					}
					tokens[i] = "'" + replacement
					continue
				}
			}

			// Check for words with a single quote at the end
			if len(word) >= 2 && word[len(word)-1] == '\'' {
				innerWord := word[:len(word)-1]
				if replacement, ok := dict[strings.ToLower(innerWord)]; ok {
					// Preserve the original case
					if isCapitalized(innerWord) {
						replacement = capitalize(replacement)
					} else if isAllCaps(innerWord) {
						replacement = strings.ToUpper(replacement)
					}
					tokens[i] = replacement + "'"
					continue
				}
			}

			// Special handling for words with single quotes
			// This is a more direct approach that doesn't rely on the previous checks
			if len(word) >= 3 {
				// Try to extract a word surrounded by single quotes
				// First, find all possible substrings that could be words
				for start := 0; start < len(word)-1; start++ {
					if word[start] == '\'' {
						for end := start + 2; end <= len(word); end++ {
							if end < len(word) && word[end] == '\'' {
								// Found a potential word surrounded by single quotes
								innerWord := word[start+1 : end]
								if replacement, ok := dict[strings.ToLower(innerWord)]; ok {
									// Preserve the original case
									if isCapitalized(innerWord) {
										replacement = capitalize(replacement)
									} else if isAllCaps(innerWord) {
										replacement = strings.ToUpper(replacement)
									}
									// Replace just the inner word, keeping everything else
									tokens[i] = word[:start+1] + replacement + word[end:]
									continue
								}
							}
						}
					}
				}
			}

			// Try a more aggressive approach for single quotes
			// This will handle cases where the tokenization might be incorrect
			if strings.Contains(word, "'") {
				// Extract all possible words from the token
				var possibleWords []string
				var startIndices []int
				var endIndices []int

				// Find all possible words between single quotes
				inQuote := false
				startIndex := 0
				for i := 0; i < len(word); i++ {
					if word[i] == '\'' {
						if inQuote {
							// End of a quoted section
							// Extract the word, but also handle the case where there's a comma inside the quotes
							quotedText := word[startIndex+1 : i]

							// Check if the quoted text has a comma
							commaIndex := strings.LastIndex(quotedText, ",")
							if commaIndex >= 0 {
								// There's a comma, so we need to check both the word with and without the comma
								possibleWords = append(possibleWords, quotedText[:commaIndex])
								startIndices = append(startIndices, startIndex)
								endIndices = append(endIndices, startIndex+1+commaIndex)

								// Also add the full quoted text
								possibleWords = append(possibleWords, quotedText)
								startIndices = append(startIndices, startIndex)
								endIndices = append(endIndices, i)
							} else {
								// No comma, just add the quoted text
								possibleWords = append(possibleWords, quotedText)
								startIndices = append(startIndices, startIndex)
								endIndices = append(endIndices, i)
							}

							inQuote = false
						} else {
							// Start of a quoted section
							startIndex = i
							inQuote = true
						}
					}
				}

				// Check each possible word
				for j, possibleWord := range possibleWords {
					if replacement, ok := dict[strings.ToLower(possibleWord)]; ok {
						// Preserve the original case
						if isCapitalized(possibleWord) {
							replacement = capitalize(replacement)
						} else if isAllCaps(possibleWord) {
							replacement = strings.ToUpper(replacement)
						}

						// Replace the word in the original token
						startIdx := startIndices[j]
						endIdx := endIndices[j]
						tokens[i] = word[:startIdx+1] + replacement + word[endIdx:]
						word = tokens[i] // Update the word for subsequent replacements
					}
				}
			}

			// Check for words with a comma at the end
			if len(word) >= 2 && word[len(word)-1] == ',' {
				innerWord := word[:len(word)-1]
				if replacement, ok := dict[strings.ToLower(innerWord)]; ok {
					// Preserve the original case
					if isCapitalized(innerWord) {
						replacement = capitalize(replacement)
					} else if isAllCaps(innerWord) {
						replacement = strings.ToUpper(replacement)
					}
					tokens[i] = replacement + ","
					continue
				}
			}

			// If that didn't work, try to strip punctuation
			cleanWord, punctuation := splitPunctuation(word)
			if cleanWord != word {
				if replacement, ok := dict[strings.ToLower(cleanWord)]; ok {
					// Preserve the original case
					if isCapitalized(cleanWord) {
						replacement = capitalize(replacement)
					} else if isAllCaps(cleanWord) {
						replacement = strings.ToUpper(replacement)
					}

					// Add the punctuation back
					if len(word) > 0 && (word[0] == '.' || word[0] == ',' || word[0] == ';' || word[0] == ':' ||
						word[0] == '!' || word[0] == '?' || word[0] == '(' || word[0] == '[' || word[0] == '{') {
						// Leading punctuation
						tokens[i] = string(word[0]) + replacement + word[1+len(cleanWord):]
					} else {
						// Trailing punctuation
						tokens[i] = replacement + punctuation
					}
					continue
				}
			}

			// Finally, check for hyphenated words
			parts := strings.Split(word, "-")
			if len(parts) > 1 {
				changed := false
				// Check each part of the hyphenated word
				for j, part := range parts {
					// Try to match the part as-is
					if replacement, ok := dict[strings.ToLower(part)]; ok {
						// Preserve the original case
						if isCapitalized(part) {
							replacement = capitalize(replacement)
						} else if isAllCaps(part) {
							replacement = strings.ToUpper(replacement)
						}
						parts[j] = replacement
						changed = true
						continue
					}

					// Try with punctuation stripped
					cleanPart, partPunct := splitPunctuation(part)
					if cleanPart != part {
						if replacement, ok := dict[strings.ToLower(cleanPart)]; ok {
							// Preserve the original case
							if isCapitalized(cleanPart) {
								replacement = capitalize(replacement)
							} else if isAllCaps(cleanPart) {
								replacement = strings.ToUpper(replacement)
							}

							// Add the punctuation back
							if len(part) > 0 && !isLetter(part[0]) && !isDigit(part[0]) {
								parts[j] = string(part[0]) + replacement + part[1+len(cleanPart):]
							} else {
								parts[j] = replacement + partPunct
							}
							changed = true
						}
					}
				}

				// Only update the token if we made changes
				if changed {
					tokens[i] = strings.Join(parts, "-")
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
