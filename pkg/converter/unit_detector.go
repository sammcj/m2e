// Package converter provides unit detection functionality
package converter

import (
	"regexp"
	"strconv"
	"strings"
)

// UnitDetector interface defines the contract for unit detection
type UnitDetector interface {
	DetectUnits(text string) []UnitMatch
	SupportedUnits() []UnitType
}

// ContextualUnitDetector implements contextual unit detection with confidence scoring
type ContextualUnitDetector struct {
	patterns *UnitPatterns

	// Configuration for contextual detection
	maxNumberDistance int     // Maximum words between number and unit
	minConfidence     float64 // Minimum confidence threshold for matches
}

// NewContextualUnitDetector creates a new contextual unit detector
func NewContextualUnitDetector() *ContextualUnitDetector {
	return &ContextualUnitDetector{
		patterns:          NewUnitPatterns(),
		maxNumberDistance: 3,   // Allow up to 3 words between number and unit
		minConfidence:     0.5, // Minimum confidence threshold
	}
}

// DetectUnits detects units in text using contextual analysis and confidence scoring
func (d *ContextualUnitDetector) DetectUnits(text string) []UnitMatch {
	var matches []UnitMatch

	// Get all pattern types
	allPatterns := d.patterns.GetAllPatterns()

	// Process each unit type
	for unitType, patterns := range allPatterns {
		for _, pattern := range patterns {
			// Find all matches for this pattern
			regexMatches := pattern.Pattern.FindAllStringSubmatch(text, -1)
			regexIndices := pattern.Pattern.FindAllStringSubmatchIndex(text, -1)

			for i, match := range regexMatches {
				if len(match) < 2 || len(regexIndices) <= i {
					continue
				}

				// Extract the numeric value
				var value float64
				var err error
				if len(match) > 1 && match[1] != "" {
					valueStr := match[1]
					value, err = d.parseNumericValue(valueStr)
					if err != nil {
						continue // Skip if we can't parse the number
					}
				} else {
					// Handle patterns without explicit numbers (e.g., "several tons")
					value = d.estimateQuantityFromContext(match[0])
				}

				// Get match positions
				start := regexIndices[i][0]
				end := regexIndices[i][1]

				// Extract the unit name from the full match
				unitName := ExtractUnitFromMatch(match, pattern.UnitNames)
				if unitName == "" {
					unitName = pattern.UnitNames[0] // Fallback to first unit name
				}

				// Get context around the match for analysis
				context := d.extractContext(text, start, end)

				// Check if this specific match should be excluded due to idiomatic usage
				if d.patterns.IsExcluded(match[0]) {
					continue // Skip this match if it's idiomatic
				}

				// Calculate confidence score
				confidence := d.calculateConfidence(match[0], context, pattern, value)

				// Check if this is a compound unit (contains hyphen)
				isCompound := strings.Contains(match[0], "-")

				// Only include matches above minimum confidence threshold
				if confidence >= d.minConfidence {
					unitMatch := UnitMatch{
						Start:      start,
						End:        end,
						Value:      value,
						Unit:       unitName,
						UnitType:   unitType,
						Context:    context,
						Confidence: confidence,
						IsCompound: isCompound,
					}

					matches = append(matches, unitMatch)
				}
			}
		}
	}

	// Sort matches by position and filter overlapping matches
	matches = d.filterOverlappingMatches(matches)

	return matches
}

// SupportedUnits returns the list of supported unit types
func (d *ContextualUnitDetector) SupportedUnits() []UnitType {
	return []UnitType{Length, Mass, Volume, Temperature, Area}
}

// parseNumericValue parses various numeric formats including decimals, fractions, and written numbers
func (d *ContextualUnitDetector) parseNumericValue(valueStr string) (float64, error) {
	valueStr = strings.TrimSpace(valueStr)

	// Handle written numbers
	writtenNumbers := map[string]float64{
		"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
		"eleven": 11, "twelve": 12, "twenty": 20, "thirty": 30,
		"forty": 40, "fifty": 50, "sixty": 60, "seventy": 70,
		"eighty": 80, "ninety": 90, "hundred": 100,
	}

	if val, exists := writtenNumbers[strings.ToLower(valueStr)]; exists {
		return val, nil
	}

	// Handle fractions (e.g., "2 1/2" or "1/2")
	if strings.Contains(valueStr, "/") {
		return d.parseFraction(valueStr)
	}

	// Handle regular decimals and integers
	return strconv.ParseFloat(valueStr, 64)
}

// parseFraction parses fraction formats like "2 1/2" or "1/2"
func (d *ContextualUnitDetector) parseFraction(valueStr string) (float64, error) {
	// Handle mixed fractions like "2 1/2"
	parts := strings.Fields(valueStr)
	if len(parts) == 2 {
		// Mixed fraction: whole number + fraction
		whole, err1 := strconv.ParseFloat(parts[0], 64)
		frac, err2 := d.parseSimpleFraction(parts[1])
		if err1 == nil && err2 == nil {
			return whole + frac, nil
		}
	}

	// Handle simple fractions like "1/2"
	return d.parseSimpleFraction(valueStr)
}

// parseSimpleFraction parses simple fractions like "1/2"
func (d *ContextualUnitDetector) parseSimpleFraction(fracStr string) (float64, error) {
	parts := strings.Split(fracStr, "/")
	if len(parts) != 2 {
		return 0, strconv.ErrSyntax
	}

	numerator, err1 := strconv.ParseFloat(parts[0], 64)
	denominator, err2 := strconv.ParseFloat(parts[1], 64)

	if err1 != nil || err2 != nil || denominator == 0 {
		return 0, strconv.ErrSyntax
	}

	return numerator / denominator, nil
}

// extractContext extracts surrounding text for contextual analysis
func (d *ContextualUnitDetector) extractContext(text string, start, end int) string {
	// Extract a reasonable amount of context (about 50 characters before and after)
	contextStart := start - 50
	if contextStart < 0 {
		contextStart = 0
	}

	contextEnd := end + 50
	if contextEnd > len(text) {
		contextEnd = len(text)
	}

	return text[contextStart:contextEnd]
}

// calculateConfidence calculates confidence score based on various factors
func (d *ContextualUnitDetector) calculateConfidence(match, context string, pattern UnitPattern, value float64) float64 {
	confidence := pattern.Confidence // Start with base pattern confidence

	// Boost confidence for explicit measurement contexts
	measurementContexts := []string{
		"tall", "high", "long", "wide", "deep", "thick", "heavy", "weighs", "weight",
		"distance", "length", "width", "height", "depth", "size", "area", "volume",
		"temperature", "temp", "degrees", "capacity", "holds", "contains",
	}

	lowerContext := strings.ToLower(context)
	for _, ctx := range measurementContexts {
		if strings.Contains(lowerContext, ctx) {
			confidence += 0.1
			break
		}
	}

	// Boost confidence for direct number-unit adjacency
	if d.isDirectAdjacency(match) {
		confidence += 0.1
	}

	// Reduce confidence for very large or very small values that might be errors
	if value > 10000 || value < 0.001 {
		confidence -= 0.2
	}

	// Boost confidence for common measurement ranges
	confidence += d.getValueRangeBoost(pattern.UnitType, value)

	// Reduce confidence if context suggests idiomatic usage
	if d.hasIdiomaticContext(lowerContext, pattern.UnitType) {
		confidence -= 0.3
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

// isDirectAdjacency checks if number and unit are directly adjacent
func (d *ContextualUnitDetector) isDirectAdjacency(match string) bool {
	// Look for patterns like "5ft", "12in", "100lbs" (no space)
	noSpacePattern := regexp.MustCompile(`\d+[a-zA-Z]`)
	return noSpacePattern.MatchString(match)
}

// getValueRangeBoost provides confidence boost for realistic measurement ranges
func (d *ContextualUnitDetector) getValueRangeBoost(unitType UnitType, value float64) float64 {
	switch unitType {
	case Length:
		// Common ranges: 1-20 feet, 1-100 inches, 1-1000 miles
		if (value >= 1 && value <= 20) || (value >= 1 && value <= 100) || (value >= 1 && value <= 1000) {
			return 0.05
		}
	case Mass:
		// Common ranges: 1-500 pounds, 1-32 ounces, 1-100 tons
		if (value >= 1 && value <= 500) || (value >= 1 && value <= 32) || (value >= 1 && value <= 100) {
			return 0.05
		}
	case Volume:
		// Common ranges: 1-50 gallons, 1-100 fluid ounces
		if (value >= 1 && value <= 50) || (value >= 1 && value <= 100) {
			return 0.05
		}
	case Temperature:
		// Common ranges: -20 to 120 Fahrenheit
		if value >= -20 && value <= 120 {
			return 0.05
		}
	case Area:
		// Common ranges: 100-10000 square feet, 1-1000 acres
		if (value >= 100 && value <= 10000) || (value >= 1 && value <= 1000) {
			return 0.05
		}
	}
	return 0.0
}

// hasIdiomaticContext checks for idiomatic usage patterns in context
func (d *ContextualUnitDetector) hasIdiomaticContext(context string, unitType UnitType) bool {
	switch unitType {
	case Length:
		idioms := []string{
			"miles away from", "miles apart", "inch by inch", "every inch",
			"cold feet", "foot in the door", "foot the bill",
		}
		for _, idiom := range idioms {
			if strings.Contains(context, idiom) {
				return true
			}
		}
	case Mass:
		idioms := []string{
			"tons of fun", "tons of work", "pounds of pressure",
			"pound the pavement", "pound the table",
		}
		for _, idiom := range idioms {
			if strings.Contains(context, idiom) {
				return true
			}
		}
	}
	return false
}

// filterOverlappingMatches removes overlapping matches, keeping the highest confidence ones
func (d *ContextualUnitDetector) filterOverlappingMatches(matches []UnitMatch) []UnitMatch {
	if len(matches) <= 1 {
		return matches
	}

	// Sort by start position
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].Start > matches[j].Start {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Remove overlapping matches
	var filtered []UnitMatch
	for _, match := range matches {
		isOverlapping := false

		// Check if this match overlaps with any already accepted match
		for _, accepted := range filtered {
			if d.isOverlapping(match, accepted) {
				// Keep the one with higher confidence
				if match.Confidence <= accepted.Confidence {
					isOverlapping = true
					break
				} else {
					// Remove the lower confidence match and add this one
					for j, existing := range filtered {
						if existing.Start == accepted.Start && existing.End == accepted.End {
							filtered = append(filtered[:j], filtered[j+1:]...)
							break
						}
					}
					break
				}
			}
		}

		if !isOverlapping {
			filtered = append(filtered, match)
		}
	}

	return filtered
}

// isOverlapping checks if two matches overlap
func (d *ContextualUnitDetector) isOverlapping(match1, match2 UnitMatch) bool {
	return match1.End > match2.Start && match2.End > match1.Start
}

// SetMinConfidence sets the minimum confidence threshold
func (d *ContextualUnitDetector) SetMinConfidence(confidence float64) {
	d.minConfidence = confidence
}

// SetMaxNumberDistance sets the maximum allowed distance between numbers and units
func (d *ContextualUnitDetector) SetMaxNumberDistance(distance int) {
	d.maxNumberDistance = distance
}

// estimateQuantityFromContext estimates a numeric value from contextual words
func (d *ContextualUnitDetector) estimateQuantityFromContext(match string) float64 {
	lowerMatch := strings.ToLower(match)

	// Estimate quantities from contextual words
	if strings.Contains(lowerMatch, "several") {
		return 3.0 // Estimate "several" as 3
	}
	if strings.Contains(lowerMatch, "many") {
		return 5.0 // Estimate "many" as 5
	}
	if strings.Contains(lowerMatch, "few") {
		return 2.0 // Estimate "few" as 2
	}
	if strings.Contains(lowerMatch, "some") {
		return 2.0 // Estimate "some" as 2
	}

	// Default fallback
	return 1.0
}
