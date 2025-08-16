// Package converter provides contextual word detection functionality
package converter

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ContextualWordDetector interface defines the contract for contextual word detection
type ContextualWordDetector interface {
	DetectWords(text string) []ContextualWordMatch
	SupportedWords() []string
	SetMinConfidence(confidence float64)
	SetEnabled(enabled bool)
	IsEnabled() bool
}

// ContextAwareWordDetector implements contextual word detection with confidence scoring
type ContextAwareWordDetector struct {
	patterns *ContextualWordPatterns
	config   *ContextualWordConfig

	// Configuration for contextual detection
	minConfidence float64 // Minimum confidence threshold for matches
	enabled       bool    // Whether contextual detection is enabled
}

// NewContextAwareWordDetector creates a new contextual word detector
func NewContextAwareWordDetector() *ContextAwareWordDetector {
	// Load configuration with defaults
	config, err := LoadContextualWordConfigWithDefaults()
	if err != nil {
		// Fall back to default config if loading fails
		fmt.Fprintf(os.Stderr, "Warning: Failed to load contextual word configuration: %v\n", err)
		config = GetDefaultContextualWordConfig()
	}

	detector := &ContextAwareWordDetector{
		patterns:      NewContextualWordPatterns(),
		config:        config,
		minConfidence: config.MinConfidence,
		enabled:       config.Enabled,
	}

	return detector
}

// NewContextAwareWordDetectorWithConfig creates a new contextual word detector with specific configuration
func NewContextAwareWordDetectorWithConfig(config *ContextualWordConfig) *ContextAwareWordDetector {
	if config == nil {
		config = GetDefaultContextualWordConfig()
	}

	return &ContextAwareWordDetector{
		patterns:      NewContextualWordPatterns(),
		config:        config,
		minConfidence: config.MinConfidence,
		enabled:       config.Enabled,
	}
}

// DetectWords detects words that need contextual conversion using pattern analysis and confidence scoring
func (d *ContextAwareWordDetector) DetectWords(text string) []ContextualWordMatch {
	if !d.enabled {
		return nil
	}

	var matches []ContextualWordMatch

	// Process each supported word
	for _, baseWord := range d.patterns.GetSupportedWords() {
		wordMatches := d.detectWordInContext(text, baseWord)
		matches = append(matches, wordMatches...)
	}

	// Filter overlapping matches and sort by confidence
	matches = d.filterAndSortMatches(matches)

	return matches
}

// detectWordInContext detects a specific word in various grammatical contexts
func (d *ContextAwareWordDetector) detectWordInContext(text string, baseWord string) []ContextualWordMatch {
	var matches []ContextualWordMatch

	patterns := d.patterns.GetPatternsForWord(baseWord)
	if len(patterns) == 0 {
		return matches
	}

	// Apply each pattern for this word
	for _, pattern := range patterns {
		// Find all matches for this pattern
		regexMatches := pattern.Pattern.FindAllStringSubmatch(text, -1)
		regexIndices := pattern.Pattern.FindAllStringSubmatchIndex(text, -1)

		for i, match := range regexMatches {
			if len(regexIndices) <= i {
				continue
			}

			// Get match positions - use capture group positions if available
			var start, end int
			if len(regexIndices[i]) >= 4 { // Full match + 1 capture group = 4 indices
				// Use the capture group positions (indices 2 and 3)
				start = regexIndices[i][2]
				end = regexIndices[i][3]
			} else {
				// Fallback to full match positions
				start = regexIndices[i][0]
				end = regexIndices[i][1]
			}

			// Get context around the match for analysis
			context := d.extractContext(text, start, end, 100)

			// Check if this specific match should be excluded
			if d.isContextExcluded(context) {
				continue
			}

			// Extract the actual matched word from the text
			matchedWord := ExtractMatchedWord(match, baseWord)
			if matchedWord == "" {
				continue
			}

			// Calculate confidence score for this match
			confidence := d.calculateConfidence(match[0], context, pattern)

			// Only include matches above minimum confidence threshold
			if confidence >= d.minConfidence {
				wordMatch := ContextualWordMatch{
					Start:        start,
					End:          end,
					OriginalWord: matchedWord,
					WordType:     pattern.WordType,
					Replacement:  d.getReplacementWord(matchedWord, pattern),
					Confidence:   confidence,
					Context:      context,
					BaseWord:     baseWord,
				}

				matches = append(matches, wordMatch)
			}
		}
	}

	return matches
}

// extractContext extracts surrounding text for contextual analysis
func (d *ContextAwareWordDetector) extractContext(text string, start, end int, contextSize int) string {
	contextStart := start - contextSize
	if contextStart < 0 {
		contextStart = 0
	}

	contextEnd := end + contextSize
	if contextEnd > len(text) {
		contextEnd = len(text)
	}

	return text[contextStart:contextEnd]
}

// calculateConfidence calculates confidence score based on pattern strength and contextual clues
func (d *ContextAwareWordDetector) calculateConfidence(match, context string, pattern ContextualWordPattern) float64 {
	confidence := pattern.Confidence // Start with base pattern confidence

	// Boost confidence for multiple contextual indicators
	confidence += d.getContextualBoosts(context, pattern.WordType)

	// Boost confidence for clear grammatical markers
	confidence += d.getGrammaticalMarkerBoost(context, pattern.WordType)

	// Reduce confidence for ambiguous contexts
	confidence -= d.getAmbiguityPenalty(context, pattern.BaseWord)

	// Boost confidence for direct adjacency to strong indicators
	if d.hasStrongAdjacentIndicators(match, context, pattern.WordType) {
		confidence += 0.1
	}

	// Ensure confidence stays within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// getContextualBoosts provides confidence boosts based on surrounding context
func (d *ContextAwareWordDetector) getContextualBoosts(context string, wordType WordType) float64 {
	boost := 0.0
	lowerContext := strings.ToLower(context)

	switch wordType {
	case Noun:
		// Boost for noun contexts
		nounIndicators := []string{
			"valid", "expired", "driver", "driving", "commercial", "professional",
			"temporary", "permanent", "annual", "renewed", "suspended", "revoked",
			"check", "verification", "copy", "photo", "scan", "document",
		}
		for _, indicator := range nounIndicators {
			if strings.Contains(lowerContext, indicator) {
				boost += 0.05
				break // Only boost once per category
			}
		}

		// Additional boost for possessive indicators
		if strings.Contains(lowerContext, "'s") || strings.Contains(lowerContext, "of") {
			boost += 0.05
		}

	case Verb:
		// Boost for verb contexts
		verbIndicators := []string{
			"agree to", "able to", "need to", "want to", "plan to", "intend to",
			"refuse to", "decide to", "choose to", "required to", "permitted to",
			"authorised to", "authorized to", "allowed to", "qualified to",
		}
		for _, indicator := range verbIndicators {
			if strings.Contains(lowerContext, indicator) {
				boost += 0.05
				break
			}
		}

		// Boost for business/legal verb contexts
		businessContexts := []string{
			"software", "technology", "intellectual property", "patent", "trademark",
			"copyright", "content", "material", "product", "service", "platform",
		}
		for _, ctx := range businessContexts {
			if strings.Contains(lowerContext, ctx) {
				boost += 0.05
				break
			}
		}
	}

	return boost
}

// getGrammaticalMarkerBoost provides boosts for clear grammatical markers
func (d *ContextAwareWordDetector) getGrammaticalMarkerBoost(context string, wordType WordType) float64 {
	boost := 0.0
	lowerContext := strings.ToLower(context)

	switch wordType {
	case Noun:
		// Strong noun markers
		if strings.Contains(lowerContext, "the ") || strings.Contains(lowerContext, "a ") || strings.Contains(lowerContext, "an ") {
			boost += 0.05
		}
		// Preposition indicators
		prepositions := []string{"with ", "without ", "for ", "under ", "by ", "of ", "from "}
		for _, prep := range prepositions {
			if strings.Contains(lowerContext, prep) {
				boost += 0.03
				break
			}
		}

	case Verb:
		// Strong verb markers
		if strings.Contains(lowerContext, "to ") {
			boost += 0.1 // Strong indicator for infinitive
		}
		// Modal verb indicators
		modals := []string{"will ", "can ", "must ", "should ", "would ", "could ", "may ", "might "}
		for _, modal := range modals {
			if strings.Contains(lowerContext, modal) {
				boost += 0.08
				break
			}
		}
	}

	return boost
}

// getAmbiguityPenalty reduces confidence for ambiguous contexts
func (d *ContextAwareWordDetector) getAmbiguityPenalty(context string, baseWord string) float64 {
	penalty := 0.0
	lowerContext := strings.ToLower(context)

	// Penalty for very short contexts (less reliable)
	if len(context) < 20 {
		penalty += 0.1
	}

	// Penalty for potential code/technical contexts
	if strings.Contains(lowerContext, "=") || strings.Contains(lowerContext, "{") ||
		strings.Contains(lowerContext, "}") || strings.Contains(lowerContext, ";") {
		penalty += 0.2
	}

	// Penalty for URLs or file paths
	if strings.Contains(lowerContext, "http") || strings.Contains(lowerContext, "www") ||
		strings.Contains(lowerContext, "/") || strings.Contains(lowerContext, "\\") {
		penalty += 0.3
	}

	return penalty
}

// hasStrongAdjacentIndicators checks for strong grammatical indicators immediately adjacent to the word
func (d *ContextAwareWordDetector) hasStrongAdjacentIndicators(match, context string, wordType WordType) bool {
	lowerContext := strings.ToLower(context)
	baseWord := strings.ToLower(match)

	// Find the position of the word in context
	wordIndex := strings.Index(lowerContext, baseWord)
	if wordIndex == -1 {
		return false
	}

	// Check words immediately before and after
	beforeWord := ""
	afterWord := ""

	// Extract word before
	if wordIndex > 0 {
		beforeText := lowerContext[:wordIndex]
		words := strings.Fields(beforeText)
		if len(words) > 0 {
			beforeWord = words[len(words)-1]
		}
	}

	// Extract word after
	afterIndex := wordIndex + len(baseWord)
	if afterIndex < len(lowerContext) {
		afterText := lowerContext[afterIndex:]
		words := strings.Fields(afterText)
		if len(words) > 0 {
			afterWord = words[0]
		}
	}

	switch wordType {
	case Noun:
		// Strong noun indicators before
		strongNounBefore := []string{"the", "a", "an", "my", "your", "his", "her", "our", "their", "this", "that"}
		for _, indicator := range strongNounBefore {
			if beforeWord == indicator {
				return true
			}
		}

		// Strong noun indicators after
		strongNounAfter := []string{"holder", "number", "plate", "renewal", "fee", "application", "agreement"}
		for _, indicator := range strongNounAfter {
			if afterWord == indicator {
				return true
			}
		}

	case Verb:
		// Strong verb indicators before
		strongVerbBefore := []string{"to", "will", "can", "must", "should", "would", "could", "may", "might"}
		for _, indicator := range strongVerbBefore {
			if beforeWord == indicator {
				return true
			}
		}

		// Strong verb indicators after (direct objects)
		strongVerbAfter := []string{"software", "technology", "content", "products", "materials", "intellectual"}
		for _, indicator := range strongVerbAfter {
			if afterWord == indicator {
				return true
			}
		}
	}

	return false
}

// getReplacementWord determines the appropriate replacement word maintaining case
func (d *ContextAwareWordDetector) getReplacementWord(originalWord string, pattern ContextualWordPattern) string {
	if originalWord == "" {
		return pattern.Replacement
	}

	// Handle inflected verb forms that should remain unchanged
	lowerOriginal := strings.ToLower(originalWord)
	if pattern.WordType == Verb {
		// These verb forms don't change in British English
		if lowerOriginal == "licensed" || lowerOriginal == "licenses" || lowerOriginal == "licensing" {
			return originalWord // Keep the original word unchanged
		}
	}

	// Handle possessive forms
	if strings.HasSuffix(originalWord, "'s") {
		baseReplacement := pattern.Replacement
		return d.preserveCase(baseReplacement, originalWord[:len(originalWord)-2]) + "'s"
	}
	if strings.HasSuffix(originalWord, "'") {
		baseReplacement := pattern.Replacement
		return d.preserveCase(baseReplacement, originalWord[:len(originalWord)-1]) + "'"
	}

	// Normal case preservation
	return d.preserveCase(pattern.Replacement, originalWord)
}

// preserveCase maintains the case pattern of the original word
func (d *ContextAwareWordDetector) preserveCase(replacement, original string) string {
	if len(original) == 0 {
		return replacement
	}

	// Check if original is all caps
	if strings.ToUpper(original) == original {
		return strings.ToUpper(replacement)
	}

	// Check if original is title case (first letter uppercase)
	if len(original) > 0 && strings.ToUpper(original[:1]) == original[:1] {
		if len(replacement) > 0 {
			return strings.ToUpper(replacement[:1]) + strings.ToLower(replacement[1:])
		}
	}

	// Default to lowercase
	return strings.ToLower(replacement)
}

// filterAndSortMatches removes overlapping matches and sorts by confidence
func (d *ContextAwareWordDetector) filterAndSortMatches(matches []ContextualWordMatch) []ContextualWordMatch {
	if len(matches) <= 1 {
		return matches
	}

	// Sort by start position first
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].Start > matches[j].Start {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Remove overlapping matches, keeping the highest confidence ones
	var filtered []ContextualWordMatch
	for _, match := range matches {
		isOverlapping := false

		// Check if this match overlaps with any already accepted match
		for j, accepted := range filtered {
			if d.isOverlapping(match, accepted) {
				// Keep the one with higher confidence
				if match.Confidence > accepted.Confidence {
					// Replace the lower confidence match
					filtered[j] = match
				}
				isOverlapping = true
				break
			}
		}

		if !isOverlapping {
			filtered = append(filtered, match)
		}
	}

	return filtered
}

// isOverlapping checks if two matches overlap
func (d *ContextAwareWordDetector) isOverlapping(match1, match2 ContextualWordMatch) bool {
	return match1.End > match2.Start && match2.End > match1.Start
}

// SupportedWords returns the list of words that support contextual conversion
func (d *ContextAwareWordDetector) SupportedWords() []string {
	if d.config != nil {
		return d.config.SupportedWords
	}
	return d.patterns.GetSupportedWords()
}

// SetMinConfidence sets the minimum confidence threshold
func (d *ContextAwareWordDetector) SetMinConfidence(confidence float64) {
	if confidence >= 0.0 && confidence <= 1.0 {
		d.minConfidence = confidence
	}
}

// SetEnabled enables or disables contextual word detection
func (d *ContextAwareWordDetector) SetEnabled(enabled bool) {
	d.enabled = enabled
}

// IsEnabled returns whether contextual word detection is enabled
func (d *ContextAwareWordDetector) IsEnabled() bool {
	return d.enabled
}

// GetConfig returns the current configuration
func (d *ContextAwareWordDetector) GetConfig() *ContextualWordConfig {
	return d.config
}

// SetConfig sets a new configuration and applies it
func (d *ContextAwareWordDetector) SetConfig(config *ContextualWordConfig) {
	if config != nil {
		d.config = config
		d.minConfidence = config.MinConfidence
		d.enabled = config.Enabled
	}
}

// ReloadConfig reloads the configuration from file
func (d *ContextAwareWordDetector) ReloadConfig() error {
	config, err := LoadContextualWordConfig()
	if err != nil {
		return err
	}
	d.SetConfig(config)
	return nil
}

// isContextExcluded checks if a context should be excluded from conversion
func (d *ContextAwareWordDetector) isContextExcluded(context string) bool {
	// Check built-in pattern exclusions
	if d.patterns.IsExcluded(context) {
		return true
	}

	// Check configuration exclusions
	if d.config != nil {
		for _, pattern := range d.config.ExcludePatterns {
			if matched, err := regexp.MatchString(pattern, context); err == nil && matched {
				return true
			}
		}
	}

	return false
}
