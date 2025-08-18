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

// WordConfig represents the configuration for a contextual word pair
type WordConfig struct {
	Noun    string `json:"noun"`    // British spelling when used as noun
	Verb    string `json:"verb"`    // British spelling when used as verb
	Enabled bool   `json:"enabled"` // Whether this word pair is enabled
}

// GeneralPattern represents a reusable pattern template
type GeneralPattern struct {
	Name       string   // Pattern identifier
	Template   string   // Pattern template with {WORD} placeholder
	TargetType WordType // The grammatical role this pattern detects
	Confidence float64  // Base confidence for this pattern (0.0-1.0)
}

// ContextualWordPatterns holds all the patterns and configuration for contextual word detection
type ContextualWordPatterns struct {
	// Word configurations by base word
	WordConfigs map[string]WordConfig

	// Generated patterns by base word
	GeneratedPatterns map[string][]ContextualWordPattern

	// Exclusion patterns for ambiguous or problematic contexts
	ExclusionPatterns []*regexp.Regexp

	// General pattern templates
	GeneralPatterns []GeneralPattern
}

// NewContextualWordPatterns creates and initialises the contextual word detection system
func NewContextualWordPatterns() *ContextualWordPatterns {
	patterns := &ContextualWordPatterns{
		WordConfigs:       make(map[string]WordConfig),
		GeneratedPatterns: make(map[string][]ContextualWordPattern),
	}

	patterns.initialiseDefaultWordConfigs()
	patterns.initialiseGeneralPatterns()
	patterns.initialiseExclusionPatterns()
	patterns.generateAllPatterns()

	return patterns
}

// initialiseDefaultWordConfigs sets up the default word configurations
func (p *ContextualWordPatterns) initialiseDefaultWordConfigs() {
	p.WordConfigs = map[string]WordConfig{
		"license": {
			Noun:    "licence",
			Verb:    "license",
			Enabled: true,
		},
		"practice": {
			Noun:    "practice",
			Verb:    "practise",
			Enabled: true,
		},
		"advice": {
			Noun:    "advice",
			Verb:    "advise",
			Enabled: true,
		},
	}
}

// initialiseGeneralPatterns sets up the reusable pattern templates
func (p *ContextualWordPatterns) initialiseGeneralPatterns() {
	p.GeneralPatterns = []GeneralPattern{
		// NOUN PATTERNS
		{
			Name: "determiner_noun",
			// Matches: "a|an|the|this|that|my|your|his|her|our|their|each|every|any|some" + optional words + target word
			// Examples: "a device", "the licence", "my advice", "some practice sessions"
			Template:   `(?i)\b(?:a|an|the|this|that|my|your|his|her|our|their|each|every|any|some)\s+(?:\w+\s+)*?['"]?({WORD})['"]?\b`,
			TargetType: Noun,
			Confidence: 0.8,
		},
		{
			Name: "preposition_object",
			// Matches: preposition + optional words + target word as object
			// Examples: "with a licence", "for advice", "about the device", "regarding practice"
			Template:   `(?i)\b(?:with|without|by|under|for|against|on|in|of|from|about|regarding|concerning)\s+(?:\w+\s+)*?['"]?({WORD})['"]?(?:\s|$)`,
			TargetType: Noun,
			Confidence: 0.85,
		},
		{
			Name: "possessive",
			// Matches: target word + possessive 's
			// Examples: "device's features", "licence's terms", "advice's value"
			Template:   `(?i)\b['"]?({WORD})['"]?'?s\b`,
			TargetType: Noun,
			Confidence: 0.95,
		},
		{
			Name: "compound_noun",
			// Matches: target word + common noun compounds
			// Examples: "licence holder", "device number", "practice sessions"
			Template:   `(?i)\b['"]?({WORD})['"]?\s+(?:holder|number|plate|renewal|application|fee|requirement|agreement|terms|expiration|document|copy|file)\b`,
			TargetType: Noun,
			Confidence: 0.9,
		},
		{
			Name: "sentence_end_noun",
			// Matches: target word at the end of sentence or before punctuation
			// Examples: "I need a licence.", "Get some advice!", "Buy the device,"
			Template:   `(?i)\b['"]?({WORD})['"]?(?:\s*)(?:[.!?;,]|$)`,
			TargetType: Noun,
			Confidence: 0.7,
		},

		// VERB PATTERNS
		{
			Name: "infinitive",
			// Matches: "to" + target word (infinitive form)
			// Examples: "to license", "to devise", "to practise", "to advise"
			Template:   `(?i)\bto\s+['"]?({WORD})['"]?\b`,
			TargetType: Verb,
			Confidence: 0.98,
		},
		{
			Name: "modal_verb",
			// Matches: modal verb + optional words + target word
			// Examples: "will license", "can devise", "should practise", "might advise"
			Template:   `(?i)\b(?:will|shall|must|can|could|should|would|may|might)\s+(?:\w+\s+)*?['"]?({WORD})['"]?\b`,
			TargetType: Verb,
			Confidence: 0.95,
		},
		{
			Name: "subject_verb",
			// Matches: subject pronoun + optional adverbs + target word
			// Examples: "I license", "they devise", "we practise", "you advise"
			Template:   `(?i)\b(?:I|you|we|they|he|she|it|who)\s+(?:also\s+|often\s+|always\s+|never\s+|sometimes\s+|usually\s+)?['"]?({WORD})['"]?\b`,
			TargetType: Verb,
			Confidence: 0.8,
		},
		{
			Name: "direct_object",
			// Matches: target word + direct object (technology/software terms)
			// Examples: "license software", "devise technology", "practise skills"
			Template:   `(?i)\b['"]?({WORD})['"]?\s+(?:software|technology|content|users|products|materials|code|applications|services|data|information)\b`,
			TargetType: Verb,
			Confidence: 0.9,
		},
	}
}

// initialiseExclusionPatterns creates patterns for excluding ambiguous or problematic contexts
func (p *ContextualWordPatterns) initialiseExclusionPatterns() {
	// Contexts where conversion should be avoided
	exclusions := []string{
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

	for _, pattern := range exclusions {
		compiled := regexp.MustCompile(pattern)
		p.ExclusionPatterns = append(p.ExclusionPatterns, compiled)
	}
}

// generateAllPatterns generates contextual patterns for all enabled words
func (p *ContextualWordPatterns) generateAllPatterns() {
	for baseWord, config := range p.WordConfigs {
		if config.Enabled {
			patterns := p.generatePatternsForWord(baseWord, config)
			p.GeneratedPatterns[baseWord] = patterns
		}
	}
}

// generatePatternsForWord generates contextual patterns for a specific word
func (p *ContextualWordPatterns) generatePatternsForWord(word string, config WordConfig) []ContextualWordPattern {
	var patterns []ContextualWordPattern

	for _, generalPattern := range p.GeneralPatterns {
		// Replace {WORD} placeholder with actual word
		patternText := strings.ReplaceAll(generalPattern.Template, "{WORD}", word)
		compiled, err := regexp.Compile(patternText)
		if err != nil {
			continue // Skip invalid patterns
		}

		var replacement string
		if generalPattern.TargetType == Noun {
			replacement = config.Noun
		} else {
			replacement = config.Verb
		}

		patterns = append(patterns, ContextualWordPattern{
			Pattern:     compiled,
			WordType:    generalPattern.TargetType,
			BaseWord:    word,
			Replacement: replacement,
			Confidence:  generalPattern.Confidence,
			Description: generalPattern.Name + " pattern for " + word,
		})
	}

	return patterns
}

// GetPatternsForWord returns all patterns for a specific base word
func (p *ContextualWordPatterns) GetPatternsForWord(baseWord string) []ContextualWordPattern {
	return p.GeneratedPatterns[strings.ToLower(baseWord)]
}

// GetAllPatterns returns all contextual word patterns grouped by base word
func (p *ContextualWordPatterns) GetAllPatterns() map[string][]ContextualWordPattern {
	return p.GeneratedPatterns
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
	var supportedWords []string
	for word, config := range p.WordConfigs {
		if config.Enabled {
			supportedWords = append(supportedWords, word)
		}
	}
	return supportedWords
}

// ExtractMatchedWord extracts the actual word from a regex match, handling different capture group scenarios
func ExtractMatchedWord(match []string, baseWord string) string {
	if len(match) == 0 {
		return ""
	}

	// If we have capture groups, use the first capture group (which should be the word)
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

// AddWordConfig adds or updates a word configuration
func (p *ContextualWordPatterns) AddWordConfig(word string, config WordConfig) {
	p.WordConfigs[strings.ToLower(word)] = config
	if config.Enabled {
		patterns := p.generatePatternsForWord(word, config)
		p.GeneratedPatterns[word] = patterns
	} else {
		delete(p.GeneratedPatterns, word)
	}
}

// GetWordConfig returns the configuration for a specific word
func (p *ContextualWordPatterns) GetWordConfig(word string) (WordConfig, bool) {
	config, exists := p.WordConfigs[strings.ToLower(word)]
	return config, exists
}
