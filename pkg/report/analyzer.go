package report

import (
	"regexp"
	"strings"
	"unicode"
)

// Analyzer provides functionality to analyze text changes and generate statistics
type Analyzer struct {
	americanWords map[string]string
	unitPatterns  []*regexp.Regexp
}

// NewAnalyzer creates a new text change analyzer
func NewAnalyzer(americanWords map[string]string) *Analyzer {
	analyzer := &Analyzer{
		americanWords: americanWords,
		unitPatterns:  make([]*regexp.Regexp, 0),
	}

	// Initialize unit conversion patterns
	analyzer.initUnitPatterns()

	return analyzer
}

// initUnitPatterns sets up regex patterns for detecting unit conversions
func (a *Analyzer) initUnitPatterns() {
	unitPatterns := []string{
		`\b\d+(?:\.\d+)?\s*(?:feet|foot|ft)\b`,
		`\b\d+(?:\.\d+)?\s*(?:inches?|in)\b`,
		`\b\d+(?:\.\d+)?\s*(?:yards?|yds?)\b`,
		`\b\d+(?:\.\d+)?\s*(?:miles?|mi)\b`,
		`\b\d+(?:\.\d+)?\s*(?:pounds?|lbs?|lb)\b`,
		`\b\d+(?:\.\d+)?\s*(?:ounces?|oz)\b`,
		`\b\d+(?:\.\d+)?\s*(?:tons?)\b`,
		`\b\d+(?:\.\d+)?\s*(?:gallons?|gal)\b`,
		`\b\d+(?:\.\d+)?\s*(?:quarts?|qt)\b`,
		`\b\d+(?:\.\d+)?\s*(?:pints?|pt)\b`,
		`\b\d+(?:\.\d+)?\s*(?:fluid\s+ounces?|fl\s+oz)\b`,
		`\b\d+(?:\.\d+)?\s*°F\b`,
		`\b\d+(?:\.\d+)?\s*(?:square\s+feet|sq\s+ft)\b`,
		`\b\d+(?:\.\d+)?\s*(?:acres?)\b`,
	}

	for _, pattern := range unitPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			a.unitPatterns = append(a.unitPatterns, regex)
		}
	}
}

// AnalyzeChanges compares original and converted text to generate detailed statistics
func (a *Analyzer) AnalyzeChanges(original, converted string) ChangeStats {
	stats := ChangeStats{
		ChangedWords: make([]WordChange, 0),
		ChangedUnits: make([]UnitChange, 0),
	}

	// Count total words
	stats.TotalWords = a.countWords(original)

	// Analyze spelling changes
	a.analyzeSpellingChanges(original, converted, &stats)

	// Analyze unit conversions
	a.analyzeUnitConversions(original, converted, &stats)

	// Analyze quote changes
	a.analyzeQuoteChanges(original, converted, &stats)

	return stats
}

// countWords counts the number of words in the text
func (a *Analyzer) countWords(text string) int {
	words := strings.FieldsFunc(text, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '\'' && c != '-'
	})
	return len(words)
}

// analyzeSpellingChanges detects American to British spelling changes
func (a *Analyzer) analyzeSpellingChanges(original, converted string, stats *ChangeStats) {
	// Simple word-by-word comparison
	originalWords := a.extractWords(original)
	convertedWords := a.extractWords(converted)

	if len(originalWords) != len(convertedWords) {
		return // Can't do reliable word-by-word comparison
	}

	position := 0
	for i, origWord := range originalWords {
		if i < len(convertedWords) && origWord != convertedWords[i] {
			// Check if this is a known American->British conversion
			if _, isAmericanWord := a.americanWords[strings.ToLower(origWord)]; isAmericanWord {
				stats.ChangedWords = append(stats.ChangedWords, WordChange{
					Original: origWord,
					Changed:  convertedWords[i],
					Position: position,
				})
				stats.SpellingChanges++
			}
		}
		position += len(origWord) + 1 // +1 for space
	}
}

// analyzeUnitConversions detects unit conversions
func (a *Analyzer) analyzeUnitConversions(original, converted string, stats *ChangeStats) {
	// Find all unit patterns in original text
	for _, pattern := range a.unitPatterns {
		matches := pattern.FindAllStringSubmatch(original, -1)
		for _, match := range matches {
			if len(match) > 0 {
				originalUnit := match[0]
				// Look for the corresponding conversion in the converted text
				if convertedUnit := a.findCorrespondingConversion(originalUnit, original, converted); convertedUnit != "" {
					unitType := a.determineUnitType(originalUnit)
					stats.ChangedUnits = append(stats.ChangedUnits, UnitChange{
						Original: originalUnit,
						Changed:  convertedUnit,
						Position: strings.Index(original, originalUnit),
						UnitType: unitType,
					})
					stats.UnitConversions++
				}
			}
		}
	}
}

// analyzeQuoteChanges detects smart quote normalizations
func (a *Analyzer) analyzeQuoteChanges(original, converted string, stats *ChangeStats) {
	smartQuotes := []string{"\u201c", "\u201d", "\u2018", "\u2019", "\u2013", "\u2014"}

	for _, quote := range smartQuotes {
		originalCount := strings.Count(original, quote)
		convertedCount := strings.Count(converted, quote)
		if originalCount > convertedCount {
			stats.QuoteChanges += originalCount - convertedCount
		}
	}
}

// extractWords extracts words from text for comparison
func (a *Analyzer) extractWords(text string) []string {
	return strings.FieldsFunc(text, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '\'' && c != '-'
	})
}

// findCorrespondingConversion finds the metric equivalent of an imperial unit
func (a *Analyzer) findCorrespondingConversion(originalUnit, original, converted string) string {
	// Find the position of the original unit
	pos := strings.Index(original, originalUnit)
	if pos == -1 {
		return ""
	}

	// Look for a metric unit in the same approximate position in converted text
	metricPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:metres?|meters?|m)\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:centimetres?|centimeters?|cm)\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:kilometres?|kilometers?|km)\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:kilograms?|kg)\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:grams?|g)\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:tonnes?|t)\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:litres?|liters?|l)\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:millilitres?|milliliters?|ml)\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*°C\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:square\s+metres?|square\s+meters?|m²)\b`),
		regexp.MustCompile(`\b\d+(?:\.\d+)?\s*(?:hectares?|ha)\b`),
	}

	// Look in a reasonable range around the original position
	start := max(0, pos-50)
	end := min(len(converted), pos+len(originalUnit)+100)
	searchArea := converted[start:end]

	for _, pattern := range metricPatterns {
		if match := pattern.FindString(searchArea); match != "" {
			return match
		}
	}

	return ""
}

// determineUnitType determines the type of unit being converted
func (a *Analyzer) determineUnitType(unit string) string {
	lowerUnit := strings.ToLower(unit)

	if strings.Contains(lowerUnit, "feet") || strings.Contains(lowerUnit, "foot") ||
		strings.Contains(lowerUnit, "ft") || strings.Contains(lowerUnit, "inch") ||
		strings.Contains(lowerUnit, "yard") || strings.Contains(lowerUnit, "mile") {
		return "length"
	}

	if strings.Contains(lowerUnit, "pound") || strings.Contains(lowerUnit, "lb") ||
		strings.Contains(lowerUnit, "ounce") || strings.Contains(lowerUnit, "oz") ||
		strings.Contains(lowerUnit, "ton") {
		return "mass"
	}

	if strings.Contains(lowerUnit, "gallon") || strings.Contains(lowerUnit, "gal") ||
		strings.Contains(lowerUnit, "quart") || strings.Contains(lowerUnit, "pint") ||
		strings.Contains(lowerUnit, "fluid") {
		return "volume"
	}

	if strings.Contains(lowerUnit, "°f") || strings.Contains(lowerUnit, "fahrenheit") {
		return "temperature"
	}

	if strings.Contains(lowerUnit, "square") || strings.Contains(lowerUnit, "acre") {
		return "area"
	}

	return "unknown"
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
