// Package converter provides functionality to convert between American and British English spellings
package converter

import (
	"embed"
	"maps"
	"regexp"
	"strings"
	"unicode"
)

//go:embed data/*.json
var dictFS embed.FS

// Regex to detect URLs and skip them during conversion
var urlRegex = regexp.MustCompile(`(?i)(https?://|www\.)\S+`)

// Dictionaries holds the mapping for American to British English spellings
type Dictionaries struct {
	AmericanToBritish map[string]string
}

// Converter provides methods to convert between American and British English
type Converter struct {
	dict                   *Dictionaries
	filteredDict           map[string]string // dictionary with contextual words removed
	unitProcessor          *UnitProcessor
	contextualWordDetector ContextualWordDetector
	ignoreProcessor        *CommentIgnoreProcessor
	markdownProcessor      *MarkdownProcessor
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

// smartQuoteReplacer performs all smart quote replacements in a single pass.
var smartQuoteReplacer = strings.NewReplacer(
	"\u201C", "\"",
	"\u201D", "\"",
	"\u2018", "'",
	"\u2019", "'",
	"\u2013", "-",
	"\u2014", "-",
)

// NewConverter creates a new Converter instance
func NewConverter() (*Converter, error) {
	dict, err := LoadDictionaries()
	if err != nil {
		return nil, err
	}

	contextualWordDetector := NewContextAwareWordDetector()

	// Pre-compute filtered dictionary with contextual words removed
	filtered := make(map[string]string, len(dict.AmericanToBritish))
	maps.Copy(filtered, dict.AmericanToBritish)
	if contextualWordDetector != nil {
		for _, word := range contextualWordDetector.SupportedWords() {
			delete(filtered, strings.ToLower(word))
		}
	}

	return &Converter{
		dict:                   dict,
		filteredDict:           filtered,
		unitProcessor:          NewUnitProcessor(),
		contextualWordDetector: contextualWordDetector,
		ignoreProcessor:        NewCommentIgnoreProcessor(),
		markdownProcessor:      NewMarkdownProcessor(),
	}, nil
}

// ConvertToBritish converts American English text to British English
func (c *Converter) ConvertToBritish(text string, normaliseSmartQuotes bool) string {
	// Process ignore comments first
	return c.ConvertToBritishWithIgnoreComments(text, normaliseSmartQuotes)
}

// ConvertToBritishWithIgnoreComments handles ignore comments and selective conversion
func (c *Converter) ConvertToBritishWithIgnoreComments(text string, normaliseSmartQuotes bool) string {
	// Find all ignore directives in the text
	ignoreMatches := c.ignoreProcessor.ProcessIgnoreComments(text)

	// If the entire file should be ignored, return original text
	if c.ignoreProcessor.ShouldIgnoreFile(ignoreMatches) {
		return text
	}

	// Apply selective ignore using the ignore processor
	return c.ignoreProcessor.ApplySelectiveIgnore(text, ignoreMatches, func(lineText string) string {
		// Use code-aware processing for each non-ignored line
		return c.ProcessCodeAware(lineText, normaliseSmartQuotes)
	})
}

// ConvertToBritishSimple converts text without code-awareness (for internal use)
func (c *Converter) ConvertToBritishSimple(text string, normaliseSmartQuotes bool) string {
	// Wrap the entire conversion in markdown processing to preserve formatting
	if c.markdownProcessor != nil {
		return c.markdownProcessor.ProcessWithMarkdown(text, func(innerText string) string {
			return c.convertWithoutMarkdown(innerText, normaliseSmartQuotes)
		})
	}

	// Fallback if markdown processor is not available
	return c.convertWithoutMarkdown(text, normaliseSmartQuotes)
}

// convertWithoutMarkdown performs conversion without markdown processing
func (c *Converter) convertWithoutMarkdown(text string, normaliseSmartQuotes bool) string {
	// First normalise smart quotes if needed
	processedText := text
	if normaliseSmartQuotes {
		processedText = c.normaliseSmartQuotes(text)
	}

	// Apply contextual word conversion if enabled
	if c.contextualWordDetector != nil && c.contextualWordDetector.IsEnabled() {
		processedText = c.applyContextualWordConversion(processedText)
	}

	// Apply standard dictionary conversion using pre-computed filtered dictionary
	return c.convert(processedText, c.filteredDict)
}

// GetAmericanToBritishDictionary returns the American to British dictionary
func (c *Converter) GetAmericanToBritishDictionary() map[string]string {
	if c.dict == nil {
		return map[string]string{}
	}
	return c.dict.AmericanToBritish
}

// GetUnitProcessor returns the unit processor instance
func (c *Converter) GetUnitProcessor() *UnitProcessor {
	return c.unitProcessor
}

// SetUnitProcessingEnabled enables or disables unit processing
func (c *Converter) SetUnitProcessingEnabled(enabled bool) {
	if c.unitProcessor != nil {
		c.unitProcessor.SetEnabled(enabled)
	}
}

// GetContextualWordDetector returns the contextual word detector instance
func (c *Converter) GetContextualWordDetector() ContextualWordDetector {
	return c.contextualWordDetector
}

// SetContextualWordDetectionEnabled enables or disables contextual word detection
func (c *Converter) SetContextualWordDetectionEnabled(enabled bool) {
	if c.contextualWordDetector != nil {
		c.contextualWordDetector.SetEnabled(enabled)
	}
}

// IsContextualWordDetectionEnabled returns whether contextual word detection is enabled
func (c *Converter) IsContextualWordDetectionEnabled() bool {
	return c.contextualWordDetector != nil && c.contextualWordDetector.IsEnabled()
}

// GetIgnoreDirectives analyses text and returns ignore directives found
func (c *Converter) GetIgnoreDirectives(text string) []IgnoreMatch {
	if c.ignoreProcessor == nil {
		return nil
	}
	return c.ignoreProcessor.ProcessIgnoreComments(text)
}

// GetIgnoreStats returns statistics about ignore directives in the text
func (c *Converter) GetIgnoreStats(text string) map[string]int {
	if c.ignoreProcessor == nil {
		return make(map[string]int)
	}
	ignoreMatches := c.ignoreProcessor.ProcessIgnoreComments(text)
	return c.ignoreProcessor.ExtractIgnoreStats(ignoreMatches)
}

// ConvertToBritishWithoutIgnores bypasses ignore comments and processes all text
func (c *Converter) ConvertToBritishWithoutIgnores(text string, normaliseSmartQuotes bool) string {
	// Use code-aware processing for all text, bypassing ignore comments
	return c.ProcessCodeAware(text, normaliseSmartQuotes)
}

// normaliseSmartQuotes converts smart quotes and em-dashes to their normal equivalents
func (c *Converter) normaliseSmartQuotes(text string) string {
	return smartQuoteReplacer.Replace(text)
}

// lookupWithCase looks up a word in the dictionary and preserves the original casing.
func lookupWithCase(word string, dict map[string]string) (string, bool) {
	replacement, ok := dict[strings.ToLower(word)]
	if !ok {
		return "", false
	}
	if isCapitalized(word) {
		replacement = capitalize(replacement)
	} else if isAllCaps(word) {
		replacement = strings.ToUpper(replacement)
	}
	return replacement, true
}

// tokeniseLine splits a line into tokens preserving whitespace boundaries.
func tokeniseLine(line string) (tokens []string, wsFlags []bool) {
	var b strings.Builder
	currentIsWS := false

	for _, char := range line {
		charIsWS := unicode.IsSpace(char)
		if b.Len() > 0 && currentIsWS != charIsWS {
			tokens = append(tokens, b.String())
			wsFlags = append(wsFlags, currentIsWS)
			b.Reset()
		}
		b.WriteRune(char)
		currentIsWS = charIsWS
	}
	if b.Len() > 0 {
		tokens = append(tokens, b.String())
		wsFlags = append(wsFlags, currentIsWS)
	}
	return tokens, wsFlags
}

// convertQuotedWord tries to convert a word surrounded by or containing quotes.
func convertQuotedWord(word string, dict map[string]string) (string, bool) {
	// Words ending in 's (possessive)
	if strings.HasSuffix(strings.ToLower(word), "'s") {
		baseWord := word[:len(word)-2]
		if repl, ok := lookupWithCase(baseWord, dict); ok {
			return repl + "'s", true
		}
	}

	// Words wrapped in double quotes
	if len(word) >= 2 && word[0] == '"' && word[len(word)-1] == '"' {
		if repl, ok := lookupWithCase(word[1:len(word)-1], dict); ok {
			return "\"" + repl + "\"", true
		}
	}

	// Words wrapped in single quotes
	if len(word) >= 2 && word[0] == '\'' && word[len(word)-1] == '\'' {
		if repl, ok := lookupWithCase(word[1:len(word)-1], dict); ok {
			return "'" + repl + "'", true
		}
	}

	// General quote wrapping (any common quote chars)
	if len(word) >= 2 {
		firstChar := word[0]
		lastChar := word[len(word)-1]
		isFirstQuote := firstChar == '\'' || firstChar == '"'
		isLastQuote := lastChar == '\'' || lastChar == '"'
		if isFirstQuote && isLastQuote {
			if repl, ok := lookupWithCase(word[1:len(word)-1], dict); ok {
				return string(firstChar) + repl + string(lastChar), true
			}
		}
	}

	// Leading single quote only
	if len(word) >= 2 && word[0] == '\'' {
		if repl, ok := lookupWithCase(word[1:], dict); ok {
			return "'" + repl, true
		}
	}

	// Trailing single quote only
	if len(word) >= 2 && word[len(word)-1] == '\'' {
		if repl, ok := lookupWithCase(word[:len(word)-1], dict); ok {
			return repl + "'", true
		}
	}

	return "", false
}

// convertEmbeddedQuotedWords handles words with embedded single-quote pairs.
func convertEmbeddedQuotedWords(word string, dict map[string]string) (string, bool) {
	// Try to find and replace words surrounded by single quotes within the token
	if len(word) >= 3 {
		for start := 0; start < len(word)-1; start++ {
			if word[start] == '\'' {
				for end := start + 2; end <= len(word); end++ {
					if end < len(word) && word[end] == '\'' {
						innerWord := word[start+1 : end]
						if repl, ok := lookupWithCase(innerWord, dict); ok {
							return word[:start+1] + repl + word[end:], true
						}
					}
				}
			}
		}
	}

	// More aggressive: find quoted sections and replace individually
	if strings.Contains(word, "'") {
		type quotedMatch struct {
			word     string
			startIdx int
			endIdx   int
		}
		var matches []quotedMatch

		inQuote := false
		startIndex := 0
		for ci := 0; ci < len(word); ci++ {
			if word[ci] != '\'' {
				continue
			}
			if inQuote {
				quotedText := word[startIndex+1 : ci]
				if commaIdx := strings.LastIndex(quotedText, ","); commaIdx >= 0 {
					matches = append(matches, quotedMatch{quotedText[:commaIdx], startIndex, startIndex + 1 + commaIdx})
					matches = append(matches, quotedMatch{quotedText, startIndex, ci})
				} else {
					matches = append(matches, quotedMatch{quotedText, startIndex, ci})
				}
				inQuote = false
			} else {
				startIndex = ci
				inQuote = true
			}
		}

		changed := false
		result := word
		for _, m := range matches {
			if repl, ok := lookupWithCase(m.word, dict); ok {
				result = result[:m.startIdx+1] + repl + result[m.endIdx:]
				changed = true
			}
		}
		if changed {
			return result, true
		}
	}

	return "", false
}

// convertPunctuatedWord handles words with trailing/leading punctuation or commas.
func convertPunctuatedWord(word string, dict map[string]string) (string, bool) {
	// Trailing comma
	if len(word) >= 2 && word[len(word)-1] == ',' {
		if repl, ok := lookupWithCase(word[:len(word)-1], dict); ok {
			return repl + ",", true
		}
	}

	// General punctuation stripping
	cleanWord, punctuation := splitPunctuation(word)
	if cleanWord != word {
		if repl, ok := lookupWithCase(cleanWord, dict); ok {
			if len(word) > 0 && (word[0] == '.' || word[0] == ',' || word[0] == ';' || word[0] == ':' ||
				word[0] == '!' || word[0] == '?' || word[0] == '(' || word[0] == '[' || word[0] == '{') {
				return string(word[0]) + repl + word[1+len(cleanWord):], true
			}
			return repl + punctuation, true
		}
	}

	return "", false
}

// convertHyphenatedWord handles hyphenated words by converting each part.
func convertHyphenatedWord(word string, dict map[string]string) (string, bool) {
	parts := strings.Split(word, "-")
	if len(parts) <= 1 {
		return "", false
	}

	changed := false
	for j, part := range parts {
		if repl, ok := lookupWithCase(part, dict); ok {
			parts[j] = repl
			changed = true
			continue
		}
		cleanPart, partPunct := splitPunctuation(part)
		if cleanPart != part {
			if repl, ok := lookupWithCase(cleanPart, dict); ok {
				if len(part) > 0 && !isLetter(part[0]) && !isDigit(part[0]) {
					parts[j] = string(part[0]) + repl + part[1+len(cleanPart):]
				} else {
					parts[j] = repl + partPunct
				}
				changed = true
			}
		}
	}
	if changed {
		return strings.Join(parts, "-"), true
	}
	return "", false
}

// convertToken applies all conversion strategies to a single token.
func convertToken(word string, dict map[string]string) string {
	// Direct dictionary match
	if repl, ok := lookupWithCase(word, dict); ok {
		return repl
	}

	// Quoted word variations
	if repl, ok := convertQuotedWord(word, dict); ok {
		return repl
	}

	// Embedded quoted words
	if repl, ok := convertEmbeddedQuotedWords(word, dict); ok {
		return repl
	}

	// Punctuated words (comma, trailing/leading punctuation)
	if repl, ok := convertPunctuatedWord(word, dict); ok {
		return repl
	}

	// Hyphenated words
	if repl, ok := convertHyphenatedWord(word, dict); ok {
		return repl
	}

	return word
}

// convert performs the actual conversion using the provided dictionary
func (c *Converter) convert(text string, dict map[string]string) string {
	lines := strings.Split(text, "\n")
	resultLines := make([]string, len(lines))

	for lineIdx, line := range lines {
		if line == "" {
			resultLines[lineIdx] = ""
			continue
		}

		tokens, wsFlags := tokeniseLine(line)

		for i := range tokens {
			if wsFlags[i] {
				continue
			}
			if urlRegex.MatchString(tokens[i]) {
				continue
			}
			tokens[i] = convertToken(tokens[i], dict)
		}

		resultLines[lineIdx] = strings.Join(tokens, "")
	}

	return strings.Join(resultLines, "\n")
}

// applyContextualWordConversion applies contextual word detection and conversion to text
func (c *Converter) applyContextualWordConversion(text string) string {
	if c.contextualWordDetector == nil || !c.contextualWordDetector.IsEnabled() {
		return text
	}

	// Detect contextual word matches
	matches := c.contextualWordDetector.DetectWords(text)
	if len(matches) == 0 {
		return text
	}

	// Process matches in reverse order to maintain positions
	result := text
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]

		// Skip if the replacement would be the same as the original
		if match.OriginalWord == match.Replacement {
			continue
		}

		// Skip words that would result in no change
		// This prevents unnecessary processing

		// Apply the contextual replacement
		before := result[:match.Start]
		after := result[match.End:]
		result = before + match.Replacement + after
	}

	return result
}
