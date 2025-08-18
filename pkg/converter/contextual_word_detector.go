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

	return detector
}

// DetectWords finds contextual words in the given text and returns matches with confidence scores
func (d *ContextAwareWordDetector) DetectWords(text string) []ContextualWordMatch {
	if !d.enabled {
		return nil
	}

	var matches []ContextualWordMatch

	// Process each configured word
	for baseWord, wordConfig := range d.config.WordConfigs {
		if !wordConfig.Enabled {
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

// findPatternMatches finds all matches for a specific pattern in the text
func (d *ContextAwareWordDetector) findPatternMatches(text string, pattern ContextualWordPattern) []ContextualWordMatch {
	var matches []ContextualWordMatch

	// Check if text should be excluded
	if d.patterns.IsExcluded(text) {
		return matches
	}

	// Find all matches for this pattern
	allMatches := pattern.Pattern.FindAllStringSubmatchIndex(text, -1)

	for _, match := range allMatches {
		if len(match) < 4 { // Need at least start, end, group start, group end
			continue
		}

		start := match[2] // Start of first capture group
		end := match[3]   // End of first capture group

		if start == -1 || end == -1 {
			continue
		}

		originalWord := text[start:end]
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

	// Preserve case of original word
	return d.preserveCase(pattern.Replacement, originalWord)
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
	for _, pattern := range config.ExcludePatterns {
		// Skip default patterns that are already added
		isDefault := false
		defaultPatterns := []string{
			// Software license names and technical terms - avoid converting in legal/technical contexts
			`(?i)(?:MIT|BSD|GPL|Apache|Creative\s+Commons|GNU|Mozilla)\s+license`,
			// License files - avoid converting when referring to license documents
			`(?i)license\s+(?:file|txt|md|doc)`,
			// Software license agreements - avoid converting in legal contexts
			`(?i)software\s+license\s+(?:agreement|terms)`,
			// License plate - avoid converting vehicle license plates
			`(?i)license\s+plate`,
			// License filenames - avoid converting literal filename references
			`(?i)LICENSE\s*\.(?:txt|md|doc|pdf|html)`,
			// License file references with "the" article
			`(?i)the\s+LICENSE\s*\.(?:txt|md|doc|pdf|html)\s+file`,
			// URLs and file paths - avoid converting in web addresses and paths
			`(?i)(?:https?://|www\.)\S*license\S*`,
			// File system paths containing license
			`(?i)(?:/|\\)\S*license\S*(?:/|\\|\.)`,
			// Code variable names and identifiers - avoid converting programming constructs
			`(?i)(?:var|const|let|def|function|class|interface|struct|type)\s+\w*\b(?:license|practice|advice)\w*`,
			// Variable assignments and operators - avoid converting in code assignments
			`(?i)\w*\b(?:license|practice|advice)\w*\s*(?:=|:=|==|!=|<|>|\+|\-|\*|/)`,
			// Quoted strings in code contexts - avoid converting in string literals
			`(?i)(?:=|:)\s*["']\s*\w*\b(?:license|practice|advice)\w*\s*["']`,
			// String literals with trailing operators
			`(?i)["']\s*\w*\b(?:license|practice|advice)\w*\s*["']\s*(?:=|:|\))`,
		}

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
