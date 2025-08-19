// Package converter provides ignore comment functionality for selective text conversion exclusion
package converter

import (
	"regexp"
	"strings"
)

// IgnoreDirective represents different types of ignore directives
type IgnoreDirective int

const (
	IgnoreNone  IgnoreDirective = iota
	IgnoreLine                  // Ignore the following line
	IgnoreNext                  // Ignore the next line (alternative syntax)
	IgnoreFile                  // Ignore the entire file
	IgnoreBlock                 // Ignore until end comment (future enhancement)
)

// String returns the string representation of IgnoreDirective
func (id IgnoreDirective) String() string {
	switch id {
	case IgnoreLine:
		return "ignore-line"
	case IgnoreNext:
		return "ignore-next"
	case IgnoreFile:
		return "ignore-file"
	case IgnoreBlock:
		return "ignore-block"
	default:
		return "none"
	}
}

// CommentIgnoreProcessor handles detection and processing of ignore comments
type CommentIgnoreProcessor struct {
	// Patterns for different comment syntaxes
	commentPatterns []*regexp.Regexp

	// Patterns for ignore directives
	ignorePatterns map[IgnoreDirective]*regexp.Regexp
}

// IgnoreMatch represents a detected ignore directive
type IgnoreMatch struct {
	LineNumber int             // Line number where the ignore comment was found
	Directive  IgnoreDirective // Type of ignore directive
	StartPos   int             // Start position in the text
	EndPos     int             // End position in the text
	Comment    string          // The full comment text
}

// NewCommentIgnoreProcessor creates a new ignore comment processor
func NewCommentIgnoreProcessor() *CommentIgnoreProcessor {
	processor := &CommentIgnoreProcessor{
		commentPatterns: make([]*regexp.Regexp, 0),
		ignorePatterns:  make(map[IgnoreDirective]*regexp.Regexp),
	}

	processor.initialiseCommentPatterns()
	processor.initialiseIgnorePatterns()

	return processor
}

// initialiseCommentPatterns sets up regex patterns for various comment syntaxes
func (cip *CommentIgnoreProcessor) initialiseCommentPatterns() {
	commentSyntaxes := []string{
		// C-style comments
		`//.*?$`,    // Single line: // comment
		`/\*.*?\*/`, // Multi-line: /* comment */

		// Hash-style comments
		`#.*?$`, // Shell, Python, Ruby: # comment

		// HTML/XML comments
		`<!--.*?-->`, // HTML/XML: <!-- comment -->

		// SQL comments
		`--.*?$`, // SQL: -- comment

		// LaTeX/MATLAB comments
		`%.*?$`, // LaTeX/MATLAB: % comment

		// Assembly comments
		`;.*?$`, // Assembly: ; comment

		// VB.NET comments
		`'.*?$`, // VB.NET: ' comment

		// Batch file comments
		`REM\s+.*?$`, // Batch: REM comment
	}

	for _, syntax := range commentSyntaxes {
		compiled, err := regexp.Compile(`(?m)` + syntax)
		if err == nil {
			cip.commentPatterns = append(cip.commentPatterns, compiled)
		}
	}
}

// initialiseIgnorePatterns sets up regex patterns for ignore directives
func (cip *CommentIgnoreProcessor) initialiseIgnorePatterns() {
	// Common ignore directive patterns - order matters for precedence
	patterns := map[IgnoreDirective]string{
		// More specific patterns first to avoid conflicts
		IgnoreFile:  `(?i)\bm2e-ignore-file\b`,
		IgnoreNext:  `(?i)\bm2e-ignore-next\b`,
		IgnoreBlock: `(?i)\bm2e-ignore-start\b`,

		// General ignore pattern last (catches m2e-ignore-line and m2e-ignore)
		IgnoreLine: `(?i)\bm2e-ignore(?:-line)?\b`,
	}

	for directive, pattern := range patterns {
		compiled, err := regexp.Compile(pattern)
		if err == nil {
			cip.ignorePatterns[directive] = compiled
		}
	}
}

// ProcessIgnoreComments analyses text and returns ignore directives found
func (cip *CommentIgnoreProcessor) ProcessIgnoreComments(text string) []IgnoreMatch {
	var ignoreMatches []IgnoreMatch

	lines := strings.Split(text, "\n")

	for lineNum, line := range lines {
		// Check if this line contains any ignore comments
		matches := cip.findIgnoreDirectivesInLine(line, lineNum)
		ignoreMatches = append(ignoreMatches, matches...)
	}

	return ignoreMatches
}

// findIgnoreDirectivesInLine finds ignore directives in a specific line
func (cip *CommentIgnoreProcessor) findIgnoreDirectivesInLine(line string, lineNum int) []IgnoreMatch {
	var matches []IgnoreMatch

	// First, find all comments in this line
	commentMatches := cip.findCommentsInLine(line)

	// Then check each comment for ignore directives - check most specific first
	for _, commentMatch := range commentMatches {
		commentText := commentMatch.text

		// Check for specific directives first to avoid duplicates
		found := false

		// Check in order of specificity
		checkOrder := []IgnoreDirective{IgnoreFile, IgnoreNext, IgnoreBlock, IgnoreLine}

		for _, directive := range checkOrder {
			if pattern, exists := cip.ignorePatterns[directive]; exists && pattern.MatchString(commentText) {
				matches = append(matches, IgnoreMatch{
					LineNumber: lineNum,
					Directive:  directive,
					StartPos:   commentMatch.start,
					EndPos:     commentMatch.end,
					Comment:    commentText,
				})
				found = true
				break // Only match the most specific directive
			}
		}

		// If we found a match, don't check other patterns for this comment
		if found {
			continue
		}
	}

	return matches
}

// commentMatch represents a comment found in a line
type commentMatch struct {
	start int
	end   int
	text  string
}

// findCommentsInLine finds all comments in a line using a simplified approach
func (cip *CommentIgnoreProcessor) findCommentsInLine(line string) []commentMatch {
	var comments []commentMatch

	// Simplified approach: check for common comment patterns manually
	line = strings.TrimSpace(line)

	// Check each comment type
	if strings.HasPrefix(line, "//") {
		comments = append(comments, commentMatch{
			start: 0,
			end:   len(line),
			text:  line,
		})
	} else if strings.HasPrefix(line, "#") {
		comments = append(comments, commentMatch{
			start: 0,
			end:   len(line),
			text:  line,
		})
	} else if strings.HasPrefix(line, "--") {
		comments = append(comments, commentMatch{
			start: 0,
			end:   len(line),
			text:  line,
		})
	} else if strings.HasPrefix(line, "%") {
		comments = append(comments, commentMatch{
			start: 0,
			end:   len(line),
			text:  line,
		})
	} else if strings.HasPrefix(line, "<!--") && strings.HasSuffix(line, "-->") {
		comments = append(comments, commentMatch{
			start: 0,
			end:   len(line),
			text:  line,
		})
	} else if strings.HasPrefix(line, ";") {
		comments = append(comments, commentMatch{
			start: 0,
			end:   len(line),
			text:  line,
		})
	} else if strings.HasPrefix(line, "'") {
		comments = append(comments, commentMatch{
			start: 0,
			end:   len(line),
			text:  line,
		})
	} else if strings.HasPrefix(strings.ToUpper(line), "REM ") {
		comments = append(comments, commentMatch{
			start: 0,
			end:   len(line),
			text:  line,
		})
	}

	return comments
}

// ShouldIgnoreLine checks if a specific line should be ignored based on ignore directives
func (cip *CommentIgnoreProcessor) ShouldIgnoreLine(lineNum int, ignoreMatches []IgnoreMatch) bool {
	for _, match := range ignoreMatches {
		switch match.Directive {
		case IgnoreFile:
			// If any file-level ignore is found, ignore everything
			return true
		case IgnoreLine:
			// m2e-ignore: ignore the same line where the comment appears
			if match.LineNumber == lineNum {
				return true
			}
		case IgnoreNext:
			// m2e-ignore-next: ignore the next line after the comment
			if match.LineNumber+1 == lineNum {
				return true
			}
		}
	}
	return false
}

// ShouldIgnoreFile checks if the entire file should be ignored
func (cip *CommentIgnoreProcessor) ShouldIgnoreFile(ignoreMatches []IgnoreMatch) bool {
	for _, match := range ignoreMatches {
		if match.Directive == IgnoreFile {
			return true
		}
	}
	return false
}

// RemoveIgnoredLines removes lines that should be ignored and returns the filtered text
func (cip *CommentIgnoreProcessor) RemoveIgnoredLines(text string, ignoreMatches []IgnoreMatch) string {
	// If file should be ignored entirely, return original text
	if cip.ShouldIgnoreFile(ignoreMatches) {
		return text
	}

	lines := strings.Split(text, "\n")
	var filteredLines []string

	for i, line := range lines {
		if !cip.ShouldIgnoreLine(i, ignoreMatches) {
			filteredLines = append(filteredLines, line)
		} else {
			// Keep the line but mark it for later restoration
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}

// ApplySelectiveIgnore applies conversion to text while respecting ignore directives
func (cip *CommentIgnoreProcessor) ApplySelectiveIgnore(text string, ignoreMatches []IgnoreMatch, convertFunc func(string) string) string {
	// If entire file should be ignored, return original text
	if cip.ShouldIgnoreFile(ignoreMatches) {
		return text
	}

	lines := strings.Split(text, "\n")
	var processedLines []string

	for i, line := range lines {
		if cip.ShouldIgnoreLine(i, ignoreMatches) {
			// Keep the original line without conversion
			processedLines = append(processedLines, line)
		} else {
			// Apply conversion to this line
			processedLines = append(processedLines, convertFunc(line))
		}
	}

	return strings.Join(processedLines, "\n")
}

// ExtractIgnoreStats returns statistics about ignore directives found
func (cip *CommentIgnoreProcessor) ExtractIgnoreStats(ignoreMatches []IgnoreMatch) map[string]int {
	stats := make(map[string]int)

	for _, match := range ignoreMatches {
		stats[match.Directive.String()]++
	}

	return stats
}
