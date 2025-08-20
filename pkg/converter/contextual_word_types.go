package converter

import "regexp"

// WordType represents the grammatical role of a word
type WordType int

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

	// For semantic variants (different meanings, not grammatical roles)
	SemanticVariants map[string]string `json:"semanticVariants,omitempty"` // Context pattern -> correct word
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

// ContextualWordConfig holds all configuration options for contextual word conversion
type ContextualWordConfig struct {
	// Global enable/disable flag
	Enabled bool `json:"enabled"`

	// Word configurations by base word
	WordConfigs map[string]WordConfig `json:"wordConfigs"`

	// Minimum confidence threshold for contextual detection (0.0 - 1.0)
	MinConfidence float64 `json:"minConfidence"`

	// Custom exclusion patterns (regex patterns to avoid conversion)
	ExcludePatterns []string `json:"excludePatterns"`

	// Conversion preferences
	Preferences ContextualWordPreferences `json:"preferences"`

	// Backward compatibility fields
	SupportedWords []string                     `json:"-"` // Populated dynamically
	CustomMappings map[string]ContextualMapping `json:"-"` // Populated dynamically
}

// ContextualMapping represents a word that has different spellings based on context
type ContextualMapping struct {
	BaseWord        string             `json:"baseWord"`        // The base American word (e.g., "license")
	NounReplacement string             `json:"nounReplacement"` // British spelling when used as noun (e.g., "licence")
	VerbReplacement string             `json:"verbReplacement"` // British spelling when used as verb (e.g., "license")
	Confidence      map[string]float64 `json:"confidence"`      // Confidence overrides for different contexts
}

// ContextualWordPreferences holds user preferences for contextual word conversion
type ContextualWordPreferences struct {
	// Whether to prefer noun conversion when context is ambiguous
	PreferNounOnAmbiguity bool `json:"preferNounOnAmbiguity"`

	// Whether to fall back to regular dictionary when contextual conversion fails
	FallbackToDictionary bool `json:"fallbackToDictionary"`

	// Whether to show warnings for ambiguous contexts
	ShowAmbiguityWarnings bool `json:"showAmbiguityWarnings"`

	// Case sensitivity for pattern matching
	CaseSensitive bool `json:"caseSensitive"`

	// Whether to convert within quoted strings
	ConvertQuotedText bool `json:"convertQuotedText"`
}
