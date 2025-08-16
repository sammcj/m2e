// Package converter provides contextual word detection functionality for grammatically-aware text conversion
package converter

import (
	"regexp"
	"strings"
)

// WordType represents the grammatical role of a word
type WordType int

const (
	Noun WordType = iota
	Verb
	Adjective
	Unknown
)

// String returns the string representation of WordType
func (wt WordType) String() string {
	switch wt {
	case Noun:
		return "noun"
	case Verb:
		return "verb"
	case Adjective:
		return "adjective"
	default:
		return "unknown"
	}
}

// ContextualWordPattern represents a regex pattern for detecting words in specific grammatical contexts
type ContextualWordPattern struct {
	Pattern     *regexp.Regexp // Regex pattern to match the word in context
	WordType    WordType       // The grammatical role this pattern detects
	BaseWord    string         // The base word this pattern applies to (e.g., "license")
	Replacement string         // The appropriate spelling for this context (e.g., "licence" for noun)
	Confidence  float64        // Base confidence for this pattern (0.0-1.0)
	Description string         // Human-readable description of this pattern
}

// ContextualWordMatch represents a detected word that needs contextual conversion
type ContextualWordMatch struct {
	Start        int      // Start position in text
	End          int      // End position in text
	OriginalWord string   // The original word found
	WordType     WordType // Detected grammatical role
	Replacement  string   // The contextually appropriate replacement
	Confidence   float64  // Confidence score for this match (0.0-1.0)
	Context      string   // Surrounding context used for detection
	BaseWord     string   // The base word this match relates to
}

// ContextualWordPatterns holds all the regex patterns for contextual word detection
type ContextualWordPatterns struct {
	// Pattern groups by base word
	LicensePatterns []ContextualWordPattern

	// Exclusion patterns for ambiguous or problematic contexts
	ExclusionPatterns []*regexp.Regexp

	// Base words that support contextual conversion
	SupportedWords []string
}

// NewContextualWordPatterns creates and initialises all contextual word detection patterns
func NewContextualWordPatterns() *ContextualWordPatterns {
	patterns := &ContextualWordPatterns{
		SupportedWords: []string{"license", "licensed", "licenses", "licensing"},
	}
	patterns.initialiseLicensePatterns()
	patterns.initialiseExclusionPatterns()
	return patterns
}

// initialiseLicensePatterns creates regex patterns for license/licence detection
func (p *ContextualWordPatterns) initialiseLicensePatterns() {
	// NOUN PATTERNS - should use "licence" in British English
	// Note: Inflected forms (licensed, licenses, licensing) are handled by the dictionary

	// Pattern 1a: Determiners + adjectives + license (noun indicators) - handles quotes
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(?:a|an|the|this|that|my|your|his|her|our|their|each|every|any|some)\s+(?:(?:valid|expired|driving|temporary|permanent|professional|medical|business|commercial|new|old|current|previous|annual|renewed|suspended|revoked|original|duplicate|replacement|authentic|official|special|required|necessary|mandatory|proper|regular|standard|basic|advanced|full|partial|test|custom|demo|sample|trial)\s+)+['"]?(license)['"]?\b`),
		WordType:    Noun,
		BaseWord:    "license",
		Replacement: "licence",
		Confidence:  0.9,
		Description: "Determiner + specific adjective + license (noun usage)",
	})

	// Pattern 1b: Simple determiners + license (noun) - covers most noun contexts but excludes modal verbs
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(?:a|an|the|this|that|my|your|his|her|our|their|each|every|any|some)\s+['"]?(license)['"]?\b`),
		WordType:    Noun,
		BaseWord:    "license",
		Replacement: "licence",
		Confidence:  0.8,
		Description: "Determiner + license (noun usage)",
	})

	// Pattern 2: License + noun (license as modifier/part of compound noun) - no quotes for true compounds
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(license)\s+(?:holder|number|plate|renewal|application|fee|requirement|agreement|terms|expiration|suspension|revocation|validation|check|verification|registration|authority|bureau|office|department|system|database|record|document|copy|photo|picture|scan|file)\b`),
		WordType:    Noun,
		BaseWord:    "license",
		Replacement: "licence",
		Confidence:  0.95,
		Description: "License + noun compound (licence as modifier)",
	})

	// Pattern 3: Preposition + license (object of preposition = noun) - handles quotes
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(?:with|without|by|under|for|against|on|in|of|from|about|regarding|concerning)\s+(?:\w+\s+)*?['"]?(license)['"]?(?:\s|$)`),
		WordType:    Noun,
		BaseWord:    "license",
		Replacement: "licence",
		Confidence:  0.85,
		Description: "Preposition + license (object of preposition)",
	})

	// Pattern 4: Possessive forms (inherently noun)
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(license)'?s\b`),
		WordType:    Noun,
		BaseWord:    "license",
		Replacement: "licence",
		Confidence:  0.95,
		Description: "Possessive license's (noun usage)",
	})

	// Pattern 5: License at end of sentence or before punctuation (often noun) - handles quotes
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b['"]?(license)['"]?(?:\s*)(?:[.!?;,]|$)`),
		WordType:    Noun,
		BaseWord:    "license",
		Replacement: "licence",
		Confidence:  0.7,
		Description: "License at sentence end (likely noun)",
	})

	// VERB PATTERNS - should use "license" in British English

	// Pattern 6: Infinitive "to license" (clear verb indicator)
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\bto\s+(license)\b`),
		WordType:    Verb,
		BaseWord:    "license",
		Replacement: "license",
		Confidence:  0.98, // Very high confidence for infinitive
		Description: "Infinitive 'to license' (verb usage)",
	})

	// Pattern 7: Modal verbs + license (verb indicator)
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(?:will|shall|must|can|could|should|would|may|might)\s+(?:\w+\s+)*?(license)\b`),
		WordType:    Verb,
		BaseWord:    "license",
		Replacement: "license",
		Confidence:  0.95, // Increased confidence
		Description: "Modal verb + license (verb usage)",
	})

	// Pattern 8: Subject pronouns + license (verb pattern) - more restrictive
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(?:I|you|we|they|he|she|it|who)\s+(?:also\s+|often\s+|always\s+|never\s+|sometimes\s+|usually\s+|currently\s+|actively\s+)?(license)\b`),
		WordType:    Verb,
		BaseWord:    "license",
		Replacement: "license",
		Confidence:  0.85,
		Description: "Subject pronoun + license (verb usage)",
	})

	// Pattern 9: License + direct object (verb taking object)
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(license)\s+(?:software|technology|content|users|products|materials|intellectual|property|code|applications|services|platforms|tools|systems|data|information|assets|rights|patents|trademarks|copyrights)\b`),
		WordType:    Verb,
		BaseWord:    "license",
		Replacement: "license",
		Confidence:  0.9,
		Description: "License + direct object (verb taking object)",
	})

	// INFLECTED FORM PATTERNS - handle past tense, plural, and participle forms

	// Pattern 10: Licensed (past tense/adjective) - primarily verb form, should stay "licensed"
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(licensed)\b`),
		WordType:    Verb,
		BaseWord:    "license",
		Replacement: "licensed", // Keep American spelling for verb forms
		Confidence:  0.8,
		Description: "Licensed (past tense/adjective - verb form)",
	})

	// Pattern 11: Licenses (present tense verb or plural noun) - context dependent
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(?:he|she|it|the\s+company|the\s+organization|the\s+authority)\s+(?:\w+\s+)*?(licenses)\b`),
		WordType:    Verb,
		BaseWord:    "license",
		Replacement: "licenses", // Keep American spelling for verb forms
		Confidence:  0.8,
		Description: "Licenses (third person singular verb)",
	})

	// Pattern 12: Licenses (plural noun)
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(?:the|their|our|his|her|my|your|some|many|few|several|multiple|valid|expired|temporary|permanent)\s+(?:\w+\s+)*?(licenses)\b`),
		WordType:    Noun,
		BaseWord:    "license",
		Replacement: "licences", // British spelling for noun forms
		Confidence:  0.8,
		Description: "Licenses (plural noun)",
	})

	// Pattern 13: Licensing (present participle) - primarily verb form
	p.LicensePatterns = append(p.LicensePatterns, ContextualWordPattern{
		Pattern:     regexp.MustCompile(`(?i)\b(licensing)\b`),
		WordType:    Verb,
		BaseWord:    "license",
		Replacement: "licensing", // Keep American spelling for verb forms
		Confidence:  0.8,
		Description: "Licensing (present participle - verb form)",
	})

}

// initialiseExclusionPatterns creates patterns for excluding ambiguous or problematic contexts
func (p *ContextualWordPatterns) initialiseExclusionPatterns() {
	// Contexts where conversion should be avoided
	exclusions := []string{
		// Software license names and technical terms where American spelling is standard
		`(?i)(?:MIT|BSD|GPL|Apache|Creative\s+Commons|GNU|Mozilla)\s+license`,
		`(?i)license\s+(?:file|txt|md|doc)`,
		`(?i)software\s+license\s+(?:agreement|terms)`,

		// License filenames (like LICENSE.txt, LICENSE.md, etc.)
		`(?i)LICENSE\s*\.(?:txt|md|doc|pdf|html)`,
		`(?i)the\s+LICENSE\s*\.(?:txt|md|doc|pdf|html)\s+file`,

		// URLs and file paths
		`(?i)(?:https?://|www\.)\S*license\S*`,
		`(?i)(?:/|\\)\S*license\S*(?:/|\\|\.)`,

		// Code variable names and identifiers
		`(?i)(?:var|const|let|def|function|class|interface|struct|type)\s+\w*license\w*`,
		`(?i)\w*license\w*\s*(?:=|:=|==|!=|<|>|\+|\-|\*|/)`,

		// Quoted strings in code contexts (variables, identifiers) - more specific
		`(?i)(?:=|:)\s*["']\s*\w*license\w*\s*["']`,                          // assignment contexts like var = "license"
		`(?i)["']\s*\w*license\w*\s*["']\s*(?:=|:|\))`,                       // value contexts like "license" = or "license")
		`(?i)(?:var|const|let|function|class)\s+["']\s*\w*license\w*\s*["']`, // declaration contexts

		// License plate contexts (should remain "license" as it's a compound noun from American usage)
		`(?i)license\s+plate`,
		`(?i)number\s+plate`, // British equivalent, but if "license plate" appears, keep it
	}

	for _, pattern := range exclusions {
		compiled := regexp.MustCompile(pattern)
		p.ExclusionPatterns = append(p.ExclusionPatterns, compiled)
	}
}

// GetPatternsForWord returns all patterns for a specific base word
func (p *ContextualWordPatterns) GetPatternsForWord(baseWord string) []ContextualWordPattern {
	baseWord = strings.ToLower(baseWord)
	switch baseWord {
	case "license":
		return p.LicensePatterns
	default:
		return nil
	}
}

// GetAllPatterns returns all contextual word patterns grouped by base word
func (p *ContextualWordPatterns) GetAllPatterns() map[string][]ContextualWordPattern {
	return map[string][]ContextualWordPattern{
		"license": p.LicensePatterns,
	}
}

// IsExcluded checks if the given text matches any exclusion pattern
func (p *ContextualWordPatterns) IsExcluded(text string) bool {
	for _, pattern := range p.ExclusionPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// GetSupportedWords returns the list of words that support contextual conversion
func (p *ContextualWordPatterns) GetSupportedWords() []string {
	return p.SupportedWords
}

// ExtractMatchedWord extracts the actual word from a regex match, handling different capture group scenarios
func ExtractMatchedWord(match []string, baseWord string) string {
	if len(match) == 0 {
		return ""
	}

	// If we have capture groups, use the first capture group (which should be the license word)
	if len(match) > 1 && match[1] != "" {
		return match[1]
	}

	// Fallback to searching in the full match (for patterns without capture groups)
	fullMatch := match[0]
	baseWordLower := strings.ToLower(baseWord)

	// Find the base word in the full match
	lowerMatch := strings.ToLower(fullMatch)
	index := strings.Index(lowerMatch, baseWordLower)
	if index == -1 {
		return ""
	}

	// Extract the word preserving original case
	wordStart := index
	wordEnd := index + len(baseWordLower)

	// Handle possessive forms
	if wordEnd < len(fullMatch) && fullMatch[wordEnd:wordEnd+1] == "'" {
		if wordEnd+1 < len(fullMatch) && fullMatch[wordEnd+1:wordEnd+2] == "s" {
			wordEnd += 2
		} else {
			wordEnd += 1
		}
	}

	return fullMatch[wordStart:wordEnd]
}
