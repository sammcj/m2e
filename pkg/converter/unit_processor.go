// Package converter provides unit processing functionality for imperial-to-metric conversion
package converter

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// UnitProcessor handles unit detection and conversion
type UnitProcessor struct {
	detector  UnitDetector
	converter UnitConverter
	config    *UnitConfig
}

// NewUnitProcessor creates a new UnitProcessor with default components
func NewUnitProcessor() *UnitProcessor {
	// Load configuration with defaults
	config, err := LoadConfigWithDefaults()
	if err != nil {
		// Fall back to default config if loading fails
		fmt.Fprintf(os.Stderr, "Warning: Failed to load unit configuration: %v\n", err)
		config = GetDefaultUnitConfig()
	}

	processor := &UnitProcessor{
		detector:  NewContextualUnitDetector(),
		converter: NewBasicUnitConverter(),
		config:    config,
	}

	// Apply configuration to components
	processor.applyConfigToComponents()

	return processor
}

// NewUnitProcessorWithConfig creates a new UnitProcessor with a specific configuration
func NewUnitProcessorWithConfig(config *UnitConfig) *UnitProcessor {
	if config == nil {
		config = GetDefaultUnitConfig()
	}

	processor := &UnitProcessor{
		detector:  NewContextualUnitDetector(),
		converter: NewBasicUnitConverter(),
		config:    config,
	}

	// Apply configuration to components
	processor.applyConfigToComponents()

	return processor
}

// SetEnabled enables or disables unit processing
func (p *UnitProcessor) SetEnabled(enabled bool) {
	if p.config != nil {
		p.config.Enabled = enabled
	}
}

// IsEnabled returns whether unit processing is enabled
func (p *UnitProcessor) IsEnabled() bool {
	return p.config != nil && p.config.Enabled
}

// GetConfig returns the current configuration
func (p *UnitProcessor) GetConfig() *UnitConfig {
	return p.config
}

// SetConfig sets a new configuration
func (p *UnitProcessor) SetConfig(config *UnitConfig) {
	if config != nil {
		p.config = config
		// Apply configuration to detector and converter
		p.applyConfigToComponents()
	}
}

// applyConfigToComponents applies the current configuration to detector and converter
func (p *UnitProcessor) applyConfigToComponents() {
	if p.config == nil {
		return
	}

	// Apply configuration to detector
	if detector, ok := p.detector.(*ContextualUnitDetector); ok {
		detector.SetMinConfidence(p.config.Detection.MinConfidence)
		detector.SetMaxNumberDistance(p.config.Detection.MaxNumberDistance)
	}

	// Apply configuration to converter
	if converter, ok := p.converter.(*BasicUnitConverter); ok {
		// Set precision for each unit type
		for unitType := range p.config.Precision {
			switch unitType {
			case "length":
				converter.SetPrecision(Length, p.config.GetPrecisionForUnitType(Length))
			case "mass":
				converter.SetPrecision(Mass, p.config.GetPrecisionForUnitType(Mass))
			case "volume":
				converter.SetPrecision(Volume, p.config.GetPrecisionForUnitType(Volume))
			case "temperature":
				converter.SetPrecision(Temperature, p.config.GetPrecisionForUnitType(Temperature))
			case "area":
				converter.SetPrecision(Area, p.config.GetPrecisionForUnitType(Area))
			}
		}

		// Set conversion preferences
		converter.SetPreferences(p.config.Preferences)
	}
}

// ProcessText processes text for unit conversion
func (p *UnitProcessor) ProcessText(text string, isCode bool, language string) string {
	if !p.IsEnabled() {
		return text
	}

	// If this is code, don't process units directly - only in comments
	if isCode {
		return p.ProcessComments(text, language)
	}

	// For regular text, detect and convert all units
	return p.convertUnitsInText(text)
}

// ProcessComments processes only comments within code for unit conversion
func (p *UnitProcessor) ProcessComments(code string, language string) string {
	if !p.IsEnabled() {
		return code
	}

	// Extract comments from the code using the same patterns as extractCommentsManually
	comments := p.extractCommentsFromCode(code)

	if len(comments) == 0 {
		return code
	}

	// Process comments in reverse order to maintain positions
	result := code
	for i := len(comments) - 1; i >= 0; i-- {
		comment := comments[i]

		// Convert units in the comment content
		convertedContent := p.convertUnitsInText(comment.Content)

		// If the original comment had a trailing newline, preserve it
		originalBlock := code[comment.Start:comment.End]
		if strings.HasSuffix(originalBlock, "\n") && !strings.HasSuffix(convertedContent, "\n") {
			convertedContent += "\n"
		}

		// Replace this comment in the code
		before := result[:comment.Start]
		after := result[comment.End:]
		result = before + convertedContent + after
	}

	return result
}

// extractCommentsFromCode extracts comments from code using the same patterns as extractCommentsManually
func (p *UnitProcessor) extractCommentsFromCode(code string) []CommentBlock {
	var comments []CommentBlock

	// Line comment patterns that should include newlines
	lineCommentPatterns := []*regexp.Regexp{
		regexp.MustCompile(`//.*?(?:\n|$)`), // Line comments: // comment with newline
		regexp.MustCompile(`#.*?(?:\n|$)`),  // Hash comments: # comment with newline
	}

	// Block comment patterns (already include their boundaries)
	blockCommentPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?s)/\*.*?\*/`), // Block comments: /* comment */
		regexp.MustCompile(`(?s)""".*?"""`), // Python docstrings: """comment"""
		regexp.MustCompile(`(?s)'''.*?'''`), // Python docstrings: '''comment'''
		regexp.MustCompile(`<!--.*?-->`),    // HTML comments: <!-- comment -->
	}

	// Find line comments (include newline if present)
	for _, pattern := range lineCommentPatterns {
		matches := pattern.FindAllStringIndex(code, -1)
		for _, match := range matches {
			start := match[0]
			end := match[1]
			content := code[start:end]

			// Remove trailing newline from content for processing, but keep the position
			content = strings.TrimSuffix(content, "\n")

			comments = append(comments, CommentBlock{
				Start:   start,
				End:     end,
				Content: content,
			})
		}
	}

	// Find block comments
	for _, pattern := range blockCommentPatterns {
		matches := pattern.FindAllStringIndex(code, -1)
		for _, match := range matches {
			start := match[0]
			end := match[1]
			content := code[start:end]

			comments = append(comments, CommentBlock{
				Start:   start,
				End:     end,
				Content: content,
			})
		}
	}

	return comments
}

// convertUnitsInText performs the actual unit detection and conversion
func (p *UnitProcessor) convertUnitsInText(text string) string {
	// Detect units in the text
	matches := p.detector.DetectUnits(text)

	if len(matches) == 0 {
		return text
	}

	// Filter matches based on configuration
	var filteredMatches []UnitMatch
	for _, match := range matches {
		// Check if this unit type is enabled
		if !p.config.IsUnitTypeEnabled(match.UnitType) {
			continue
		}

		// Check if this match should be excluded based on custom patterns
		if p.shouldExcludeMatch(match, text) {
			continue
		}

		filteredMatches = append(filteredMatches, match)
	}

	if len(filteredMatches) == 0 {
		return text
	}

	// Process matches in reverse order to maintain positions
	result := text
	for i := len(filteredMatches) - 1; i >= 0; i-- {
		match := filteredMatches[i]

		// Convert the unit
		conversion, err := p.converter.Convert(match)
		if err != nil {
			// Log error but continue processing other units
			fmt.Fprintf(os.Stderr, "Warning: Unit conversion failed for %s: %v\n", match.Unit, err)
			continue
		}

		// Handle compound units specially to preserve hyphen structure
		var replacement string
		if match.IsCompound {
			// For compound units like "9-foot", format as "2.7-metre"
			replacement = fmt.Sprintf("%.1f-%s", conversion.MetricValue, conversion.MetricUnit)
		} else {
			replacement = conversion.Formatted
		}

		// Replace the original unit with the converted one
		before := result[:match.Start]
		after := result[match.End:]
		result = before + replacement + after
	}

	return result
}

// shouldExcludeMatch checks if a match should be excluded based on custom exclude patterns
func (p *UnitProcessor) shouldExcludeMatch(match UnitMatch, text string) bool {
	if p.config == nil || len(p.config.ExcludePatterns) == 0 {
		return false
	}

	// Get the context around the match for pattern matching
	contextStart := match.Start - 50
	if contextStart < 0 {
		contextStart = 0
	}
	contextEnd := match.End + 50
	if contextEnd > len(text) {
		contextEnd = len(text)
	}
	context := text[contextStart:contextEnd]

	// Check each exclude pattern
	for _, pattern := range p.config.ExcludePatterns {
		if matched, err := regexp.MatchString(pattern, context); err == nil && matched {
			return true
		}
	}

	return false
}
