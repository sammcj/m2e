// Package converter provides contextual word detection functionality for grammatically-aware text conversion
package converter

import (
	"regexp"
	"strings"
)

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
		"program": {
			Noun:    "programme", // For non-computer contexts (TV programme, training programme)
			Verb:    "program",   // Less common as verb, but kept consistent
			Enabled: true,
		},
		"check": {
			Noun:    "cheque", // Financial instrument only
			Verb:    "check",  // Verification/examination
			Enabled: true,
		},
		"story": {
			Noun:    "storey", // Building floor context
			Verb:    "story",  // Rarely used as verb
			Enabled: true,
		},
		"inquiry": {
			Noun:    "enquiry", // General questions in British
			Verb:    "enquire", // To ask/question
			Enabled: true,
		},
		"disk": {
			Noun:    "disc", // Optical media, brake discs
			Verb:    "disc", // Rarely used as verb
			Enabled: true,
		},
		"tire": {
			Noun:    "tyre", // Automotive wheel component
			Verb:    "tire", // To become weary/fatigued
			Enabled: true,
		},
		"metre": {
			Noun:    "metre", // Unit of measurement (100 metres, square metre)
			Verb:    "metre", // Rarely used as verb
			Enabled: true,
		},
		"meter": {
			Noun:    "meter", // Measuring device (gas meter, parking meter)
			Verb:    "meter", // Rarely used as verb
			Enabled: true,
		},
		"curb": {
			Noun:    "kerb", // Pavement edge
			Verb:    "curb", // To restrain/control
			Enabled: true,
		},
		"draught": {
			Noun:    "draught", // Air current/beer context
			Verb:    "draught", // Rarely used as verb
			Enabled: true,
		},
		"draft": {
			Noun:    "draft", // Document/conscription context
			Verb:    "draft", // To conscript/create preliminary version
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
			// Matches: "a|an|the|this|that|my|your|his|her|its|our|their|each|every|any|some|no" + optional words + target word
			// Examples: "a device", "the licence", "my advice", "some practice sessions"
			Template:   `(?i)\b(?:a|an|the|this|that|my|your|his|her|its|our|their|each|every|any|some|no)\s+(?:\w+\s+)*?['"]?({WORD})['"]?\b`,
			TargetType: Noun,
			Confidence: 0.8,
		},
		{
			Name: "plural_noun",
			// Matches: plural forms with s/es ending
			// Examples: "practices", "licences", "stories"
			Template:   `(?i)\b['"]?({WORD})(?:s|es)(?!\w)['"]?\b`,
			TargetType: Noun,
			Confidence: 0.85,
		},
		{
			Name: "preposition_object",
			// Matches: preposition + optional words + target word as object
			// Examples: "with a licence", "for advice", "about the device", "regarding practice"
			Template:   `(?i)\b(?:with|without|by|under|for|against|on|in|of|from|about|regarding|concerning|during|through|across|between|among)\s+(?:\w+\s+)*?['"]?({WORD})['"]?(?:\s|$)`,
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
			// Examples: "licence holder", "practice sessions", "cheque book"
			Template:   `(?i)\b['"]?({WORD})['"]?\s+(?:holder|number|plate|renewal|application|fee|requirement|agreement|terms|expiration|document|copy|file|sessions?|book|account|building|floor)\b`,
			TargetType: Noun,
			Confidence: 0.9,
		},
		{
			Name: "high_confidence_noun_phrases",
			// Matches: high-confidence noun phrases like "best practice", "common practice"
			// Examples: "best practice", "standard practice", "driving licence"
			Template:   `(?i)\b(?:best|common|standard|good|bad|usual|normal|general|medical|legal|professional|driving|software|fishing|hunting)\s+['"]?({WORD})['"]?\b`,
			TargetType: Noun,
			Confidence: 0.95,
		},
		{
			Name: "automotive_context",
			// Matches: automotive contexts for tire → tyre
			// Examples: "car tire", "bike tire", "tire pressure", "spare tire"
			Template:   `(?i)\b(?:car|bike|bicycle|motorcycle|truck|vehicle|auto|wheel|spare|flat|front|rear|left|right)\s+['"]?({WORD})['"]?\b`,
			TargetType: Noun,
			Confidence: 0.95,
		},
		{
			Name: "measurement_unit_context",
			// Matches: measurement contexts for meter → metre
			// Examples: "100 meters", "square meters", "cubic meters"
			Template:   `(?i)\b(?:\d+(?:\.\d+)?|one|two|three|four|five|six|seven|eight|nine|ten|hundred|thousand|million|square|cubic|linear)\s+['"]?({WORD})(?:s|es)?['"]?\b`,
			TargetType: Noun,
			Confidence: 0.95,
		},
		{
			Name: "device_context",
			// Matches: device contexts for metre → meter
			// Examples: "gas meter", "parking meter", "electricity meter"
			Template:   `(?i)\b(?:gas|electric|electricity|water|parking|speed|flow|pressure|taxi|postage|postal)\s+['"]?({WORD})['"]?\b`,
			TargetType: Noun,
			Confidence: 0.95,
		},
		{
			Name: "pavement_context",
			// Matches: pavement/street contexts for curb → kerb
			// Examples: "hit the curb", "stepped off curb", "curb appeal"
			Template:   `(?i)\b(?:hit|step|stepped|off|onto|along|beside|near|against|the)\s+(?:the\s+)?['"]?({WORD})['"]?\b`,
			TargetType: Noun,
			Confidence: 0.9,
		},
		{
			Name: "air_beer_context",
			// Matches: air current/beer contexts for draft → draught
			// Examples: "cold draft", "draft beer", "feel a draft"
			Template:   `(?i)\b(?:cold|warm|cool|icy|feel|felt|beer|ale|bitter|pint|glass|bottle|tap)\s+(?:a\s+|the\s+)?['"]?({WORD})['"]?\b`,
			TargetType: Noun,
			Confidence: 0.9,
		},
		{
			Name: "document_context",
			// Matches: document/conscription contexts for draught → draft
			// Examples: "rough draft", "first draft", "draft document", "military draft"
			Template:   `(?i)\b(?:rough|first|final|initial|preliminary|military|army|navy|write|review|edit|revise)\s+(?:a\s+|the\s+)?['"]?({WORD})['"]?\b`,
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
			Template:   `(?i)\b(?:will|shall|must|can|could|should|would|may|might|ought\s+to)\s+(?:\w+\s+)*?['"]?({WORD})['"]?\b`,
			TargetType: Verb,
			Confidence: 0.95,
		},
		{
			Name: "auxiliary_verb",
			// Matches: auxiliary verbs (have/has/had/been) + target word
			// Examples: "have practised", "has licensed", "been practising"
			Template:   `(?i)\b(?:have|has|had|having|been|being)\s+(?:\w+\s+)*?['"]?({WORD})(?:d|ed)?['"]?\b`,
			TargetType: Verb,
			Confidence: 0.9,
		},
		{
			Name: "gerund_participle",
			// Matches: -ing forms in verb contexts
			// Examples: "is practising", "was advising", "been licensing"
			Template:   `(?i)\b(?:is|are|was|were|am|be|been|being|keep|keeps|kept|start|started|stop|stopped|continue|continued|finish|finished)\s+['"]?({WORD})(?:ing)['"]?\b`,
			TargetType: Verb,
			Confidence: 0.85,
		},
		{
			Name: "question_verb",
			// Matches: question words + target word
			// Examples: "Do you practice?", "Can I advise?", "Should we license?"
			Template:   `(?i)\b(?:do|does|did|can|could|should|would|will|shall|may|might)\s+(?:\w+\s+)*?['"]?({WORD})['"]?\b\?`,
			TargetType: Verb,
			Confidence: 0.9,
		},
		{
			Name: "negative_verb",
			// Matches: negative constructions with verbs
			// Examples: "don't practice", "won't advise", "can't license"
			Template:   `(?i)\b(?:don't|doesn't|didn't|won't|wouldn't|can't|couldn't|shouldn't|mustn't|mightn't|mayn't|shan't|haven't|hasn't|hadn't)\s+['"]?({WORD})['"]?\b`,
			TargetType: Verb,
			Confidence: 0.92,
		},
		{
			Name: "imperative_start",
			// Matches: imperative at sentence start
			// Examples: "Practice daily.", "License the software.", "Check your work."
			Template:   `(?i)^['"]?({WORD})['"]?\s+(?:\w+)`,
			TargetType: Verb,
			Confidence: 0.75,
		},
		{
			Name: "subject_verb",
			// Matches: subject pronoun + optional adverbs + target word
			// Examples: "I license", "they devise", "we practise", "you advise"
			Template:   `(?i)\b(?:I|you|we|they|he|she|it|who)\s+(?:also\s+|often\s+|always\s+|never\s+|sometimes\s+|usually\s+|regularly\s+|frequently\s+)?['"]?({WORD})['"]?\b`,
			TargetType: Verb,
			Confidence: 0.8,
		},
		{
			Name: "direct_object",
			// Matches: target word + direct object
			// Examples: "license software", "practise skills", "check accounts"
			Template:   `(?i)\b['"]?({WORD})['"]?\s+(?:the\s+)?(?:software|technology|content|users|products|materials|code|applications|services|data|information|skills|medicine|law|accounts|results|work)\b`,
			TargetType: Verb,
			Confidence: 0.9,
		},
		{
			Name: "professional_verb_context",
			// Matches: professional contexts where word is likely a verb
			// Examples: "practice medicine", "practice law"
			Template:   `(?i)\b['"]?({WORD})['"]?\s+(?:medicine|law|dentistry|nursing|accounting|engineering)\b`,
			TargetType: Verb,
			Confidence: 0.95,
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
		`(?i)license\s+(?:file|txt|md|mdx|doc)`,
		// Software license agreements - avoid converting in legal contexts
		`(?i)software\s+license\s+(?:agreement|terms)`,
		// License plate - avoid converting vehicle license plates
		`(?i)license\s+plate`,

		// License filenames - avoid converting literal filename references
		`(?i)LICENSE\s*\.(?:txt|md|mdx|doc|pdf|html)`,
		// License file references with "the" article
		`(?i)the\s+LICENSE\s*\.(?:txt|md|mdx|doc|pdf|html)\s+file`,

		// Computer program contexts - keep "program" for software
		`(?i)(?:computer|software|application|executable|binary)\s+program`,
		`(?i)program\s+(?:file|files|code|source|binary|executable)`,
		`(?i)(?:C|Java|Python|Go|Rust|JavaScript|TypeScript)\s+program`,

		// Financial check contexts that should NOT convert to cheque
		`(?i)(?:spell|grammar|syntax|error|bounds|null|type|security|health|status)\s+check`,
		`(?i)check\s+(?:box|boxes|mark|list|point|up|out|in|off|over)`,
		`(?i)(?:background|reference|credit|fact)\s+check`,

		// Story contexts that should NOT convert to storey
		`(?i)(?:news|short|long|love|horror|fairy|folk|bed\s*time)\s+story`,
		`(?i)story\s+(?:teller|telling|book|books|line|lines|arc|board)`,
		`(?i)(?:tell|telling|told|write|writing|wrote|read|reading)\s+(?:a\s+|the\s+)?story`,

		// Disk contexts for computer storage
		`(?i)(?:hard|floppy|solid\s+state|SSD|HDD|magnetic)\s+disk`,
		`(?i)disk\s+(?:drive|drives|space|usage|storage|partition|format|image)`,

		// Tire contexts that should NOT convert to tyre (fatigue usage)
		`(?i)(?:I|you|we|they|he|she|it|don't|doesn't|didn't|won't|wouldn't|will|would|can|could|should|might|may)\s+(?:easily\s+|quickly\s+|never\s+|often\s+|sometimes\s+)?tire`,
		`(?i)tire\s+(?:easily|quickly|of|from|out)`,

		// Meter contexts that should be metre (measurement units)
		`(?i)(?:\d+(?:\.\d+)?|square|cubic|linear)\s+meter`,

		// Curb contexts that should NOT convert to kerb (restraint usage)
		`(?i)curb\s+(?:your|his|her|their|our|my|the|this|that)\s+(?:enthusiasm|appetite|spending|desire|impulse|habit)`,
		`(?i)(?:must|should|need\s+to|have\s+to|ought\s+to)\s+curb`,

		// Draft contexts that are ambiguous or should stay as draft
		`(?i)(?:rough|first|final|initial|preliminary)\s+draft`,
		`(?i)draft\s+(?:document|paper|letter|email|version|copy)`,
		`(?i)(?:military|army|navy|war)\s+draft`,

		// URLs and file paths - avoid converting in web addresses and paths
		`(?i)(?:https?://|www\.)\S*(?:license|program|check|story|disk|inquiry|tire|meter|metre|curb|kerb|draft|draught)\S*`,
		// File system paths containing these words
		`(?i)(?:/|\\)\S*(?:license|program|check|story|disk|inquiry|tire|meter|metre|curb|kerb|draft|draught)\S*(?:/|\\|\.)`,

		// Code variable names and identifiers - avoid converting programming constructs
		`(?i)(?:var|const|let|def|function|class|interface|struct|type)\s+\w*\b(?:license|practice|advice|program|check|story|disk|inquiry|tire|meter|metre|curb|kerb|draft|draught)\w*`,
		// Variable assignments and operators - avoid converting in code assignments
		`(?i)\w*\b(?:license|practice|advice|program|check|story|disk|inquiry|tire|meter|metre|curb|kerb|draft|draught)\w*\s*(?:=|:=|==|!=|<|>|\+|\-|\*|/)`,

		// Quoted strings in code contexts - avoid converting in string literals
		`(?i)(?:=|:)\s*["']\s*\w*\b(?:license|practice|advice|program|check|story|disk|inquiry|tire|meter|metre|curb|kerb|draft|draught)\w*\s*["']`,
		// String literals with trailing operators
		`(?i)["']\s*\w*\b(?:license|practice|advice|program|check|story|disk|inquiry|tire|meter|metre|curb|kerb|draft|draught)\w*\s*["']\s*(?:=|:|\))`,
	}

	for _, pattern := range exclusions {
		compiled := regexp.MustCompile(pattern)
		p.ExclusionPatterns = append(p.ExclusionPatterns, compiled)
	}
}

// GetDefaultContextualWordConfig returns the default configuration with sensible defaults
func GetDefaultContextualWordConfig() *ContextualWordConfig {
	config := &ContextualWordConfig{
		Enabled: true,
		WordConfigs: map[string]WordConfig{
			"license": {
				Noun:    "licence",
				Verb:    "license",
				Enabled: true,
			},
			"practice": {
				Noun:    "practice",
				Verb:    "practise",
				Enabled: true,
				// Semantic variants: ensure plural nouns stay as "practices"
				SemanticVariants: map[string]string{
					// Ensure "best practices" stays as noun (plural)
					`(?i)best\s+(practices)\b`:        "practices",
					`(?i)good\s+(practices)\b`:        "practices",
					`(?i)common\s+(practices)\b`:      "practices",
					`(?i)standard\s+(practices)\b`:    "practices",
					`(?i)coding\s+(practices)\b`:      "practices",
					`(?i)security\s+(practices)\b`:    "practices",
					`(?i)business\s+(practices)\b`:    "practices",
					`(?i)development\s+(practices)\b`: "practices",
					`(?i)engineering\s+(practices)\b`: "practices",
					`(?i)programming\s+(practices)\b`: "practices",
					`(?i)software\s+(practices)\b`:    "practices",
					`(?i)testing\s+(practices)\b`:     "practices",
					`(?i)deployment\s+(practices)\b`:  "practices",
					`(?i)industry\s+(practices)\b`:    "practices",
					`(?i)recommended\s+(practices)\b`: "practices",
				},
			},
			"practices": {
				// Semantic variants for plural form: ensure they stay as "practices" (noun plural)
				SemanticVariants: map[string]string{
					// These patterns ensure plural stays plural
					`(?i)best\s+(practices)\b`:        "practices",
					`(?i)good\s+(practices)\b`:        "practices",
					`(?i)common\s+(practices)\b`:      "practices",
					`(?i)standard\s+(practices)\b`:    "practices",
					`(?i)coding\s+(practices)\b`:      "practices",
					`(?i)security\s+(practices)\b`:    "practices",
					`(?i)business\s+(practices)\b`:    "practices",
					`(?i)development\s+(practices)\b`: "practices",
					`(?i)engineering\s+(practices)\b`: "practices",
					`(?i)programming\s+(practices)\b`: "practices",
					`(?i)software\s+(practices)\b`:    "practices",
					`(?i)testing\s+(practices)\b`:     "practices",
					`(?i)deployment\s+(practices)\b`:  "practices",
					`(?i)industry\s+(practices)\b`:    "practices",
					`(?i)recommended\s+(practices)\b`: "practices",
				},
				Enabled: true,
			},
			"advice": {
				Noun:    "advice",
				Verb:    "advise",
				Enabled: true,
			},
			"principal": {
				// Semantic variants: correct specific technical misuses
				SemanticVariants: map[string]string{
					// Security/design contexts should be "principle"
					`(?i)(principal)\s+of\s+least\s+privile?ge?d?`: "principle",
					`(?i)security\s+(principals?)\b`:               "principle", // when referring to guidelines
					`(?i)design\s+(principals?)\b`:                 "principle",
					`(?i)fundamental\s+(principals?)\b`:            "principle",
					`(?i)core\s+(principals?)\b`:                   "principle",
					`(?i)guiding\s+(principals?)\b`:                "principle",
					`(?i)basic\s+(principals?)\b`:                  "principle",
					`(?i)engineering\s+(principals?)\b`:            "principle",
					`(?i)architectural\s+(principals?)\b`:          "principle",
					`(?i)programming\s+(principals?)\b`:            "principle",
					`(?i)DRY\s+(principals?)\b`:                    "principle",
					`(?i)SOLID\s+(principals?)\b`:                  "principle",
				},
				Enabled: true,
			},
			"principle": {
				// Semantic variants: correct specific technical misuses
				SemanticVariants: map[string]string{
					// AWS/IAM contexts should be "principal"
					`(?i)AWS\s+IAM\s+(principles?)\b`:      "principal",
					`(?i)service\s+(principles?)\b`:        "principal",
					`(?i)user\s+(principles?)\b`:           "principal",
					`(?i)(principle)\s+ARN\b`:              "principal",
					`(?i)authentication\s+(principles?)\b`: "principal",
					`(?i)Kerberos\s+(principles?)\b`:       "principal",
					`(?i)OAuth\s+(principles?)\b`:          "principal",
					`(?i)database\s+(principles?)\b`:       "principal",
					`(?i)login\s+(principles?)\b`:          "principal",
					`(?i)(principle)\s+name\b`:             "principal",
					`(?i)(principle)\s+ID\b`:               "principal",
					// Finance contexts
					`(?i)loan\s+(principles?)\b`:  "principal",
					`(?i)(principle)\s+amount\b`:  "principal",
					`(?i)(principle)\s+payment\b`: "principal",
				},
				Enabled: true,
			},
		},
		MinConfidence: 0.7,
		ExcludePatterns: []string{
			// Software license names
			`(?i)(?:MIT|BSD|GPL|Apache|Creative\s+Commons|GNU|Mozilla)\s+license`,
			`(?i)software\s+license\s+(?:agreement|terms)`,

			// License filenames
			`(?i)LICENSE\s*\.(?:txt|md|doc|pdf|html)`,
			`(?i)the\s+LICENSE\s*\.(?:txt|md|doc|pdf|html)\s+file`,

			// URLs and file paths
			`(?i)(?:https?://|www\.)\S*license\S*`,
			`(?i)(?:/|\\)\S*license\S*(?:/|\\|\.)`,

			// Code contexts
			`(?i)(?:var|const|let|def|function|class|interface|struct|type)\s+\w*\b(?:license|practice|advice)\w*`,
			`(?i)\w*\b(?:license|practice|advice)\w*\s*(?:=|:=|==|!=|<|>|\+|\-|\*|/)`,

			// Quoted strings in code contexts
			`(?i)(?:=|:)\s*["']\s*\w*\b(?:license|practice|advice)\w*\s*["']`,
			`(?i)["']\s*\w*\b(?:license|practice|advice)\w*\s*["']\s*(?:=|:|\)|;|,)`,

			// Dialog in code contexts (HTML, JavaScript, CSS, etc.)
			`(?i)<dialog\b`,           // HTML dialog element
			`(?i)</dialog>`,           // HTML dialog closing tag
			`(?i)\bdialog\s*\.\s*\w+`, // dialog.method() calls
			`(?i)\w*\.dialog\b`,       // object.dialog properties
			`(?i)\.dialog\b`,          // CSS .dialog classes
			`(?i)#dialog\b`,           // CSS #dialog IDs
			`(?i)data-dialog`,         // data-dialog attributes
			`(?i)dialog-\w+`,          // dialog-* attributes/classes
			`(?i)\b(?:show|open|close|hide|modal)Dialog\b`,                                    // showDialog, openDialog functions
			`(?i)\bdialog(?:Box|Modal|Window|Panel)\b`,                                        // dialogBox, dialogModal compound words
			`(?i)(?:var|const|let|def|function|class|interface|struct|type)\s+\w*\bdialog\w*`, // variable/function names
			`(?i)\w*\bdialog\w*\s*(?:=|:=|==|!=|<|>|\+|\-|\*|/)`,                              // dialog in expressions
			`(?i)(?:=|:)\s*["']\s*\w*\bdialog\w*\s*["']`,                                      // quoted dialog strings in code
			`(?i)["']\s*\w*\bdialog\w*\s*["']\s*(?:=|:|\)|;|,)`,                               // dialog in quoted assignments
		},
		Preferences: ContextualWordPreferences{
			PreferNounOnAmbiguity: true,  // Default to noun when uncertain
			FallbackToDictionary:  false, // Don't use dictionary for contextual words
			ShowAmbiguityWarnings: false,
			CaseSensitive:         false,
			ConvertQuotedText:     false, // Skip quoted text by default
		},
	}

	// Populate backward compatibility fields
	config.populateBackwardCompatibilityFields()

	return config
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

	// Generate semantic variant patterns FIRST (higher priority)
	if config.SemanticVariants != nil {
		for patternText, replacement := range config.SemanticVariants {
			compiled, err := regexp.Compile(patternText)
			if err != nil {
				continue // Skip invalid patterns
			}

			patterns = append(patterns, ContextualWordPattern{
				Pattern:     compiled,
				WordType:    Unknown, // Semantic variants don't have grammatical types
				BaseWord:    word,
				Replacement: replacement,
				Confidence:  0.99, // Very high confidence for specific semantic patterns
				Description: "semantic pattern for " + word + " -> " + replacement,
			})
		}
	}

	// Generate traditional grammatical patterns if Noun/Verb are defined
	if config.Noun != "" || config.Verb != "" {
		for _, generalPattern := range p.GeneralPatterns {
			// Replace {WORD} placeholder with actual word
			patternText := strings.ReplaceAll(generalPattern.Template, "{WORD}", word)
			compiled, err := regexp.Compile(patternText)
			if err != nil {
				continue // Skip invalid patterns
			}

			var replacement string
			if generalPattern.TargetType == Noun && config.Noun != "" {
				replacement = config.Noun
			} else if generalPattern.TargetType == Verb && config.Verb != "" {
				replacement = config.Verb
			} else {
				continue // Skip if no replacement defined
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

// GetDefaultExclusionPatterns returns the default exclusion patterns
func GetDefaultExclusionPatterns() []string {
	return []string{
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
}
