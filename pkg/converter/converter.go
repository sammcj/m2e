// Package converter provides functionality to convert between American and British English spellings
package converter

import (
	"embed"
	"maps"
	"runtime"
	"strings"
	"sync"
)

//go:embed data/*.json
var dictFS embed.FS

// isURL checks if a token looks like a URL using fast string prefix checks
// instead of running a regex on every token.
func isURL(s string) bool {
	n := len(s)
	if n < 4 {
		return false
	}
	// Case-insensitive prefix check via OR-with-0x20 to lowercase ASCII letters.
	c0 := s[0] | 0x20
	if c0 == 'h' && n > 7 {
		pre := strings.ToLower(s[:8])
		return strings.HasPrefix(pre, "http://") || strings.HasPrefix(pre, "https://")
	}
	if c0 == 'w' && n > 4 {
		return strings.EqualFold(s[:4], "www.")
	}
	return false
}

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

// isASCIISpace checks if a byte is ASCII whitespace (space, tab, CR, LF, VT, FF).
func isASCIISpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}

// tokeniseLine splits a line into tokens preserving whitespace boundaries.
// Optimised for ASCII-dominant text by operating on bytes directly.
func tokeniseLine(line string) (tokens []string, wsFlags []bool) {
	if len(line) == 0 {
		return nil, nil
	}

	// Pre-allocate: estimate ~1 token per 5 chars as a rough heuristic
	estTokens := len(line)/5 + 1
	tokens = make([]string, 0, estTokens)
	wsFlags = make([]bool, 0, estTokens)

	start := 0
	currentIsWS := isASCIISpace(line[0])

	for i := 1; i < len(line); i++ {
		charIsWS := isASCIISpace(line[i])
		if currentIsWS != charIsWS {
			tokens = append(tokens, line[start:i])
			wsFlags = append(wsFlags, currentIsWS)
			start = i
			currentIsWS = charIsWS
		}
	}
	// Append the final token
	tokens = append(tokens, line[start:])
	wsFlags = append(wsFlags, currentIsWS)
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

// hasSpecialChars checks whether a word contains quotes, hyphens, or trailing punctuation
// that would require the more expensive conversion strategies.
func hasSpecialChars(word string) bool {
	for i := 0; i < len(word); i++ {
		c := word[i]
		if c == '\'' || c == '"' || c == '-' {
			return true
		}
		// Check for trailing punctuation (non-letter, non-digit at the end)
		if i == len(word)-1 && !isLetter(c) && !isDigit(c) {
			return true
		}
	}
	return false
}

// convertToken applies all conversion strategies to a single token.
func convertToken(word string, dict map[string]string) string {
	// Direct dictionary match (most common hit path)
	if repl, ok := lookupWithCase(word, dict); ok {
		return repl
	}

	// Fast path: if the word has no special characters (quotes, hyphens, trailing
	// punctuation), none of the fallback strategies can possibly match, so skip them.
	if !hasSpecialChars(word) {
		return word
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

// parallelLineThreshold is the minimum number of lines before we use parallel processing.
const parallelLineThreshold = 500

// convertLine processes a single line through tokenisation and dictionary lookup.
func convertLine(line string, dict map[string]string) string {
	if line == "" {
		return ""
	}

	tokens, wsFlags := tokeniseLine(line)

	for i := range tokens {
		if wsFlags[i] {
			continue
		}
		if isURL(tokens[i]) {
			continue
		}
		tokens[i] = convertToken(tokens[i], dict)
	}

	return strings.Join(tokens, "")
}

// convert performs the actual conversion using the provided dictionary.
// For large texts, lines are processed in parallel across available CPU cores.
func (c *Converter) convert(text string, dict map[string]string) string {
	lines := strings.Split(text, "\n")
	resultLines := make([]string, len(lines))

	if len(lines) < parallelLineThreshold {
		// Sequential path for small/medium texts
		for lineIdx, line := range lines {
			resultLines[lineIdx] = convertLine(line, dict)
		}
	} else {
		// Parallel path for large texts
		numWorkers := runtime.GOMAXPROCS(0)
		var wg sync.WaitGroup
		chunkSize := (len(lines) + numWorkers - 1) / numWorkers

		for w := range numWorkers {
			start := w * chunkSize
			if start >= len(lines) {
				break
			}
			end := start + chunkSize
			if end > len(lines) {
				end = len(lines)
			}

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				for i := start; i < end; i++ {
					resultLines[i] = convertLine(lines[i], dict)
				}
			}(start, end)
		}
		wg.Wait()
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
