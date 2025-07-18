// Package converter provides unit detection patterns and regex
package converter

import (
	"regexp"
	"strings"
)

// UnitPattern represents a regex pattern for detecting units
type UnitPattern struct {
	Pattern    *regexp.Regexp
	UnitType   UnitType
	UnitNames  []string // Possible unit names this pattern can match
	Confidence float64  // Base confidence for this pattern
}

// UnitPatterns holds all the regex patterns for unit detection
type UnitPatterns struct {
	// Positive patterns for detecting measurements
	LengthPatterns      []UnitPattern
	MassPatterns        []UnitPattern
	VolumePatterns      []UnitPattern
	TemperaturePatterns []UnitPattern
	AreaPatterns        []UnitPattern

	// Negative patterns for excluding idiomatic usage
	ExclusionPatterns []*regexp.Regexp
}

// NewUnitPatterns creates and initializes all unit detection patterns
func NewUnitPatterns() *UnitPatterns {
	patterns := &UnitPatterns{}
	patterns.initializeLengthPatterns()
	patterns.initializeMassPatterns()
	patterns.initializeVolumePatterns()
	patterns.initializeTemperaturePatterns()
	patterns.initializeAreaPatterns()
	patterns.initializeExclusionPatterns()
	return patterns
}

// initializeLengthPatterns creates regex patterns for length units (feet, inches, yards, miles)
func (p *UnitPatterns) initializeLengthPatterns() {
	// Feet patterns - capture only number and unit
	p.LengthPatterns = append(p.LengthPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\s+\d+/\d+)?|\d+\.\d+|\d+/\d+)\s*(feet|foot|ft)\b`),
		UnitType:   Length,
		UnitNames:  []string{"feet", "foot", "ft"},
		Confidence: 0.9,
	})

	// Compound feet patterns (e.g., "6-foot", "six-foot")
	p.LengthPatterns = append(p.LengthPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+|one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve)-(feet|foot|ft)\b`),
		UnitType:   Length,
		UnitNames:  []string{"feet", "foot", "ft"},
		Confidence: 0.85,
	})

	// Written numbers with feet
	p.LengthPatterns = append(p.LengthPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|twenty|thirty|forty|fifty)\s+(feet|foot)\b`),
		UnitType:   Length,
		UnitNames:  []string{"feet", "foot"},
		Confidence: 0.8,
	})

	// Inches patterns - capture only number and unit
	p.LengthPatterns = append(p.LengthPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:/\d+)?)\s*(inches?|inch|in)\b`),
		UnitType:   Length,
		UnitNames:  []string{"inches", "inch", "in"},
		Confidence: 0.9,
	})

	// Compound inches patterns
	p.LengthPatterns = append(p.LengthPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+|one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve)-(inches?|inch|in)\b`),
		UnitType:   Length,
		UnitNames:  []string{"inches", "inch", "in"},
		Confidence: 0.85,
	})

	// Yards patterns - capture only number and unit
	p.LengthPatterns = append(p.LengthPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:/\d+)?)\s*(yards?|yd)\b`),
		UnitType:   Length,
		UnitNames:  []string{"yards", "yard", "yd"},
		Confidence: 0.9,
	})

	// Miles patterns - capture only number and unit
	p.LengthPatterns = append(p.LengthPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\s+\d+/\d+)?|\d+\.\d+|\d+/\d+)\s*(miles?|mi)\b`),
		UnitType:   Length,
		UnitNames:  []string{"miles", "mile", "mi"},
		Confidence: 0.9,
	})

	// Contextual miles patterns (a few miles, several miles) - capture the whole phrase
	p.LengthPatterns = append(p.LengthPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(?:a\s+few|several|many|about|around|roughly|approximately)\s+(\d+(?:\.\d+)?)\s*(miles?|mi)\b`),
		UnitType:   Length,
		UnitNames:  []string{"miles", "mile", "mi"},
		Confidence: 0.75,
	})
}

// initializeMassPatterns creates regex patterns for mass units (pounds, ounces, tons)
func (p *UnitPatterns) initializeMassPatterns() {
	// Pounds patterns - capture only number and unit
	p.MassPatterns = append(p.MassPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:/\d+)?)\s*(pounds?|lbs?|lb)\b`),
		UnitType:   Mass,
		UnitNames:  []string{"pounds", "pound", "lbs", "lb"},
		Confidence: 0.9,
	})

	// Ounces patterns - capture only number and unit
	p.MassPatterns = append(p.MassPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:/\d+)?)\s*(ounces?|oz)\b`),
		UnitType:   Mass,
		UnitNames:  []string{"ounces", "ounce", "oz"},
		Confidence: 0.9,
	})

	// Tons patterns (US short ton) - capture only number and unit
	p.MassPatterns = append(p.MassPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:/\d+)?)\s*(tons?|ton)\b`),
		UnitType:   Mass,
		UnitNames:  []string{"tons", "ton"},
		Confidence: 0.85, // Lower confidence due to potential idiomatic usage
	})

	// Contextual tons patterns (several tons, many tons)
	p.MassPatterns = append(p.MassPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(several|many|few|some)\s+(tons?|ton)\b`),
		UnitType:   Mass,
		UnitNames:  []string{"tons", "ton"},
		Confidence: 0.6, // Lower confidence for ambiguous quantities
	})
}

// initializeVolumePatterns creates regex patterns for volume units (gallons, quarts, pints, fluid ounces)
func (p *UnitPatterns) initializeVolumePatterns() {
	// Gallons patterns - capture only number and unit
	p.VolumePatterns = append(p.VolumePatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:/\d+)?)\s*(gallons?|gal)\b`),
		UnitType:   Volume,
		UnitNames:  []string{"gallons", "gallon", "gal"},
		Confidence: 0.9,
	})

	// Quarts patterns - capture only number and unit
	p.VolumePatterns = append(p.VolumePatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:/\d+)?)\s*(quarts?|qt)\b`),
		UnitType:   Volume,
		UnitNames:  []string{"quarts", "quart", "qt"},
		Confidence: 0.9,
	})

	// Pints patterns - capture only number and unit
	p.VolumePatterns = append(p.VolumePatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:/\d+)?)\s*(pints?|pt)\b`),
		UnitType:   Volume,
		UnitNames:  []string{"pints", "pint", "pt"},
		Confidence: 0.9,
	})

	// Fluid ounces patterns - capture only number and unit
	p.VolumePatterns = append(p.VolumePatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:/\d+)?)\s*(fluid\s+ounces?|fl\s*oz|floz)\b`),
		UnitType:   Volume,
		UnitNames:  []string{"fluid ounces", "fluid ounce", "fl oz", "floz"},
		Confidence: 0.9,
	})
}

// initializeTemperaturePatterns creates regex patterns for temperature units (Fahrenheit)
func (p *UnitPatterns) initializeTemperaturePatterns() {
	// Fahrenheit with degree symbol - capture only number and unit
	p.TemperaturePatterns = append(p.TemperaturePatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?)\s*(°F)\b`),
		UnitType:   Temperature,
		UnitNames:  []string{"°F"},
		Confidence: 0.95,
	})

	// Fahrenheit without degree symbol - capture only number and unit
	p.TemperaturePatterns = append(p.TemperaturePatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?)\s*((?:degrees?\s*)?fahrenheit)\b`),
		UnitType:   Temperature,
		UnitNames:  []string{"fahrenheit", "degrees fahrenheit"},
		Confidence: 0.9,
	})

	// F (standalone, context-dependent) - capture only number and unit
	p.TemperaturePatterns = append(p.TemperaturePatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)(?:temperature|temp|heat|cold|warm|hot)\s+(?:of|is|was|reached)\s+(\d+(?:\.\d+)?)\s*(F)\b`),
		UnitType:   Temperature,
		UnitNames:  []string{"F"},
		Confidence: 0.8,
	})
}

// initializeAreaPatterns creates regex patterns for area units (square feet, acres)
func (p *UnitPatterns) initializeAreaPatterns() {
	// Square feet patterns - capture only number and unit
	p.AreaPatterns = append(p.AreaPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:,\d{3})*)\s*(square\s+feet|sq\s*ft|ft²|ft2)(?:\s|$|[.,;!?])`),
		UnitType:   Area,
		UnitNames:  []string{"square feet", "sq ft", "ft²"},
		Confidence: 0.95,
	})

	// Acres patterns - capture only number and unit
	p.AreaPatterns = append(p.AreaPatterns, UnitPattern{
		Pattern:    regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)?(?:,\d{3})*)\s*(acres?|acre)\b`),
		UnitType:   Area,
		UnitNames:  []string{"acres", "acre"},
		Confidence: 0.9,
	})
}

// initializeExclusionPatterns creates patterns for excluding idiomatic usage
func (p *UnitPatterns) initializeExclusionPatterns() {
	// Idiomatic expressions that should NOT be converted
	exclusions := []string{
		// Miles idioms
		`(?i)miles?\s+(?:away\s+from\s+home|apart|ahead\s+of|behind)`,
		`(?i)go\s+the\s+extra\s+mile`,
		`(?i)miles?\s+from\s+nowhere`,
		`(?i)miles?\s+and\s+miles?`,

		// Inches idioms
		`(?i)inch\s+by\s+inch`,
		`(?i)every\s+inch`,
		`(?i)give\s+(?:an\s+)?inch`,
		`(?i)inch\s+(?:closer|further|away)`,

		// Feet idioms
		`(?i)(?:get|getting|have|having)\s+cold\s+feet`,
		`(?i)(?:put|set)\s+foot\s+(?:in|on)`,
		`(?i)foot\s+in\s+(?:the\s+)?door`,
		`(?i)foot\s+the\s+bill`,

		// Pounds idioms (without numbers)
		`(?i)pounds?\s+of\s+(?:fun|pressure|force|flesh)`,
		`(?i)pound\s+(?:the\s+pavement|sand|table)`,

		// Tons idioms (without numbers)
		`(?i)tons?\s+of\s+(?:fun|work|stuff|things|people)`,

		// Compound words that aren't measurements
		`(?i)\b(?:milestone|footprint|yardstick|inchworm|footstep|foothold|footpath)\b`,

		// Temperature context exclusions
		`(?i)fahrenheit\s+(?:scale|thermometer)`,
	}

	for _, pattern := range exclusions {
		compiled := regexp.MustCompile(pattern)
		p.ExclusionPatterns = append(p.ExclusionPatterns, compiled)
	}
}

// GetAllPatterns returns all unit patterns grouped by type
func (p *UnitPatterns) GetAllPatterns() map[UnitType][]UnitPattern {
	return map[UnitType][]UnitPattern{
		Length:      p.LengthPatterns,
		Mass:        p.MassPatterns,
		Volume:      p.VolumePatterns,
		Temperature: p.TemperaturePatterns,
		Area:        p.AreaPatterns,
	}
}

// IsExcluded checks if the given text matches any exclusion pattern
func (p *UnitPatterns) IsExcluded(text string) bool {
	for _, pattern := range p.ExclusionPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// ExtractUnitFromMatch extracts the unit name from a regex match
func ExtractUnitFromMatch(match []string, unitNames []string) string {
	if len(match) < 1 {
		return ""
	}

	// Try to find the unit in the capture groups first (match[2] if available)
	var unitMatch string
	if len(match) >= 3 && match[2] != "" {
		unitMatch = strings.ToLower(match[2])
	} else {
		// Fallback to searching in the full match (match[0])
		unitMatch = strings.ToLower(match[0])
	}

	// Find the longest matching unit name
	var bestMatch string
	for _, unitName := range unitNames {
		if strings.Contains(unitMatch, strings.ToLower(unitName)) {
			if len(unitName) > len(bestMatch) {
				bestMatch = unitName
			}
		}
	}

	return bestMatch
}
