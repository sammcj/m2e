// Package converter provides contextual word detection functionality
package converter

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// NewContextAwareWordDetector creates a new contextual word detector
func NewContextAwareWordDetector() *ContextAwareWordDetector {
	// Load configuration with defaults
	config, err := LoadContextualWordConfigWithDefaults()
	if err != nil {
		// Fall back to default config if loading fails
		fmt.Fprintf(os.Stderr, "Warning: Failed to load contextual word configuration: %v\n", err)
		config = GetDefaultContextualWordConfig()
	}

	// Create patterns using configuration
	patterns := NewContextualWordPatterns()

	// Update patterns to use configuration
	patterns.WordConfigs = config.WordConfigs
	patterns.generateAllPatterns()

	// Add custom exclusion patterns from config
	for _, pattern := range config.ExcludePatterns {
		compiled, err := regexp.Compile(pattern)
		if err == nil {
			patterns.ExclusionPatterns = append(patterns.ExclusionPatterns, compiled)
		}
	}

	detector := &ContextAwareWordDetector{
		patterns:      patterns,
		config:        config,
		minConfidence: config.MinConfidence,
		enabled:       config.Enabled,
	}
	detector.buildQuickCheckWords()

	return detector
}

// NewContextAwareWordDetectorWithConfig creates a new contextual word detector with specific configuration
func NewContextAwareWordDetectorWithConfig(config *ContextualWordConfig) *ContextAwareWordDetector {
	// Create patterns using configuration
	patterns := NewContextualWordPatterns()

	// Update patterns to use configuration
	patterns.WordConfigs = config.WordConfigs
	patterns.generateAllPatterns()

	// Add custom exclusion patterns from config
	for _, pattern := range config.ExcludePatterns {
		compiled, err := regexp.Compile(pattern)
		if err == nil {
			patterns.ExclusionPatterns = append(patterns.ExclusionPatterns, compiled)
		}
	}

	detector := &ContextAwareWordDetector{
		patterns:      patterns,
		config:        config,
		minConfidence: config.MinConfidence,
		enabled:       config.Enabled,
	}
	detector.buildQuickCheckWords()

	return detector
}

// buildQuickCheckWords pre-computes the list of lowercase base words for fast pre-screening.
func (d *ContextAwareWordDetector) buildQuickCheckWords() {
	seen := make(map[string]bool)
	for baseWord, wordConfig := range d.config.WordConfigs {
		if !wordConfig.Enabled {
			continue
		}
		lower := strings.ToLower(baseWord)
		if !seen[lower] {
			d.quickCheckWords = append(d.quickCheckWords, lower)
			seen[lower] = true
		}
	}
}

// DetectWords finds contextual words in the given text and returns matches with confidence scores
func (d *ContextAwareWordDetector) DetectWords(text string) []ContextualWordMatch {
	if !d.enabled {
		return nil
	}

	// Fast pre-check: skip all regex work if text contains none of the base words.
	// This eliminates the vast majority of lines from expensive regex processing.
	textLower := strings.ToLower(text)
	hasAny := false
	for _, w := range d.quickCheckWords {
		if strings.Contains(textLower, w) {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return nil
	}

	// Check full-text exclusion once instead of per-pattern.
	if d.patterns.IsExcluded(text) {
		return nil
	}

	var matches []ContextualWordMatch

	// Process only words that are actually present in the text
	for baseWord, wordConfig := range d.config.WordConfigs {
		if !wordConfig.Enabled {
			continue
		}

		// Skip words not present in the text (avoids running patterns for absent words)
		if !strings.Contains(textLower, strings.ToLower(baseWord)) {
			continue
		}

		// Get patterns for this word
		patterns := d.patterns.GetPatternsForWord(baseWord)

		// Find matches for each pattern
		for _, pattern := range patterns {
			patternMatches := d.findPatternMatches(text, pattern)
			matches = append(matches, patternMatches...)
		}
	}

	// Filter matches by confidence and remove duplicates
	matches = d.filterAndDeduplicateMatches(matches)

	return matches
}

// findPatternMatches finds all matches for a specific pattern in the text.
// The caller should check full-text exclusion via IsExcluded before calling this.
func (d *ContextAwareWordDetector) findPatternMatches(text string, pattern ContextualWordPattern) []ContextualWordMatch {
	var matches []ContextualWordMatch

	// Find all matches for this pattern
	allMatches := pattern.Pattern.FindAllStringSubmatchIndex(text, -1)

	for _, match := range allMatches {
		var start, end int
		var originalWord string

		// Handle semantic patterns differently (they match the entire phrase but replace captured group)
		if pattern.WordType == Unknown { // Semantic patterns
			if len(match) < 4 { // Need at least full match and first capture group
				continue
			}
			start = match[2] // Start of first capture group (the word to replace)
			end = match[3]   // End of first capture group
			if start == -1 || end == -1 {
				continue
			}
			originalWord = text[start:end]
		} else { // Traditional grammatical patterns
			if len(match) < 4 { // Need at least start, end, group start, group end
				continue
			}
			start = match[2] // Start of first capture group
			end = match[3]   // End of first capture group

			if start == -1 || end == -1 {
				continue
			}
			originalWord = text[start:end]
		}

		if originalWord == "" {
			continue
		}

		// Extract surrounding context for analysis
		contextStart := maxInt(0, start-50)
		contextEnd := minInt(len(text), end+50)
		context := text[contextStart:contextEnd]

		// Check if this specific context should be excluded
		if d.patterns.IsExcluded(context) {
			continue
		}

		// Calculate confidence for this match
		confidence := d.calculateConfidence(pattern, context, originalWord)

		if confidence >= d.minConfidence {
			// Get the appropriate replacement word
			replacement := d.getReplacementWord(originalWord, pattern)

			matches = append(matches, ContextualWordMatch{
				Start:        start,
				End:          end,
				OriginalWord: originalWord,
				WordType:     pattern.WordType,
				Replacement:  replacement,
				Confidence:   confidence,
				Context:      context,
				BaseWord:     pattern.BaseWord,
			})
		}
	}

	return matches
}

// calculateConfidence determines the confidence score for a match
func (d *ContextAwareWordDetector) calculateConfidence(pattern ContextualWordPattern, context, originalWord string) float64 {
	confidence := pattern.Confidence

	// Adjust confidence based on context analysis
	contextLower := strings.ToLower(context)

	// Boost confidence for very strong indicators
	if pattern.WordType == Verb {
		if strings.Contains(contextLower, "to "+strings.ToLower(originalWord)) {
			confidence = minFloat(confidence+0.1, 1.0) // Infinitive is very strong verb indicator
		}
	}

	if pattern.WordType == Noun {
		if strings.Contains(contextLower, "the "+strings.ToLower(originalWord)) {
			confidence = minFloat(confidence+0.05, 1.0) // Definite article is strong noun indicator
		}
	}

	// Reduce confidence for specific technical contexts
	if strings.Contains(contextLower, "software license") {
		confidence = maxFloat(confidence-0.2, 0.0) // Software license agreements are often technical terms
	}

	return confidence
}

// getReplacementWord returns the appropriate replacement for the detected word
func (d *ContextAwareWordDetector) getReplacementWord(originalWord string, pattern ContextualWordPattern) string {
	if originalWord == "" {
		return pattern.Replacement
	}

	// For semantic patterns, handle smart replacement with singular/plural preservation
	if pattern.WordType == Unknown {
		replacement := pattern.Replacement

		// Handle singular/plural for semantic patterns
		isOriginalPlural := strings.HasSuffix(strings.ToLower(originalWord), "s")
		isReplacementPlural := strings.HasSuffix(strings.ToLower(replacement), "s")

		// If original is singular but replacement is plural, make replacement singular
		if !isOriginalPlural && isReplacementPlural {
			// Remove the 's' from the end
			replacement = replacement[:len(replacement)-1]
		}
		// If original is plural but replacement is singular, make replacement plural
		if isOriginalPlural && !isReplacementPlural {
			replacement = replacement + "s"
		}

		return d.preserveCaseInPhrase(replacement, originalWord)
	}

	// For grammatical patterns, preserve case of the single word
	return d.preserveCase(pattern.Replacement, originalWord)
}

// preserveCaseInPhrase preserves case patterns in phrase replacements
func (d *ContextAwareWordDetector) preserveCaseInPhrase(replacement, original string) string {
	if len(original) == 0 || len(replacement) == 0 {
		return replacement
	}

	// For phrases, we try to preserve the case pattern of key words
	// This is a simple approach that handles the most common cases
	if original == strings.ToUpper(original) && original != strings.ToLower(original) {
		return strings.ToUpper(replacement)
	}

	// Check if first character is capitalised
	if len(original) > 0 && len(replacement) > 0 {
		if original[0:1] == strings.ToUpper(original[0:1]) {
			return strings.ToUpper(replacement[0:1]) + replacement[1:]
		}
	}

	// Default: keep replacement as-is (lowercase)
	return replacement
}

// preserveCase preserves the case pattern of the original word when applying replacement
func (d *ContextAwareWordDetector) preserveCase(replacement, original string) string {
	if len(original) == 0 || len(replacement) == 0 {
		return replacement
	}

	// Handle different case patterns
	if original == strings.ToUpper(original) && original != strings.ToLower(original) {
		return strings.ToUpper(replacement)
	}

	if len(original) > 0 && len(replacement) > 0 {
		first := original[0:1]
		if len(original) > 1 {
			rest := original[1:]
			if first == strings.ToUpper(first) && rest == strings.ToLower(rest) {
				// Capitalised
				return strings.ToUpper(replacement[0:1]) + strings.ToLower(replacement[1:])
			}
		} else if first == strings.ToUpper(first) {
			// Single character, capitalised
			return strings.ToUpper(replacement)
		}
	}

	// Default: keep replacement as-is (lowercase)
	return strings.ToLower(replacement)
}

// filterAndDeduplicateMatches removes duplicates and filters by confidence
func (d *ContextAwareWordDetector) filterAndDeduplicateMatches(matches []ContextualWordMatch) []ContextualWordMatch {
	if len(matches) == 0 {
		return matches
	}

	// Sort matches by position to handle overlaps
	// Use a simple bubble sort since we typically have few matches
	for i := 0; i < len(matches)-1; i++ {
		for j := 0; j < len(matches)-i-1; j++ {
			if matches[j].Start > matches[j+1].Start {
				matches[j], matches[j+1] = matches[j+1], matches[j]
			}
		}
	}

	// Remove overlapping matches, keeping the one with higher confidence
	filtered := []ContextualWordMatch{}
	for _, match := range matches {
		if match.Confidence < d.minConfidence {
			continue
		}

		// Check for overlap with previous match
		if len(filtered) > 0 {
			lastMatch := &filtered[len(filtered)-1]
			if match.Start < lastMatch.End {
				// Overlapping matches - keep the one with higher confidence
				if match.Confidence > lastMatch.Confidence {
					// Replace the last match with current match
					*lastMatch = match
				}
				// Skip current match if previous has higher confidence
				continue
			}
		}

		filtered = append(filtered, match)
	}

	return filtered
}

// SupportedWords returns a list of words that support contextual conversion
func (d *ContextAwareWordDetector) SupportedWords() []string {
	return d.config.GetSupportedWords()
}

// SetMinConfidence sets the minimum confidence threshold for matches
func (d *ContextAwareWordDetector) SetMinConfidence(confidence float64) {
	if confidence >= 0.0 && confidence <= 1.0 {
		d.minConfidence = confidence
	}
}

// SetEnabled enables or disables contextual word detection
func (d *ContextAwareWordDetector) SetEnabled(enabled bool) {
	d.enabled = enabled
}

// IsEnabled returns whether contextual word detection is currently enabled
func (d *ContextAwareWordDetector) IsEnabled() bool {
	return d.enabled
}

// GetConfiguration returns the current configuration
func (d *ContextAwareWordDetector) GetConfiguration() *ContextualWordConfig {
	return d.config
}

// UpdateConfiguration updates the detector with new configuration
func (d *ContextAwareWordDetector) UpdateConfiguration(config *ContextualWordConfig) {
	d.config = config
	d.minConfidence = config.MinConfidence
	d.enabled = config.Enabled

	// Regenerate patterns with new configuration
	d.patterns.WordConfigs = config.WordConfigs
	d.patterns.generateAllPatterns()

	// Clear and regenerate exclusion patterns from scratch
	d.patterns.ExclusionPatterns = nil
	d.patterns.initialiseExclusionPatterns() // Add default exclusions

	// Add custom exclusion patterns from config that aren't already included
	defaultPatterns := GetDefaultExclusionPatterns()
	for _, pattern := range config.ExcludePatterns {
		// Skip default patterns that are already added
		isDefault := false
		for _, defaultPattern := range defaultPatterns {
			if pattern == defaultPattern {
				isDefault = true
				break
			}
		}

		if !isDefault {
			compiled, err := regexp.Compile(pattern)
			if err == nil {
				d.patterns.ExclusionPatterns = append(d.patterns.ExclusionPatterns, compiled)
			}
		}
	}

	// Rebuild the quick check word list
	d.buildQuickCheckWords()
}

// GetConfig returns the current configuration (backward compatibility)
func (d *ContextAwareWordDetector) GetConfig() *ContextualWordConfig {
	return d.config
}

// SetConfig updates the detector with new configuration (backward compatibility)
func (d *ContextAwareWordDetector) SetConfig(config *ContextualWordConfig) {
	d.UpdateConfiguration(config)
}

// Helper functions

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
