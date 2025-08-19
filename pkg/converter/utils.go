// Package converter provides utility functions for text processing and case preservation
package converter

import (
	"strings"
)

// Helper functions for case preservation

// isCapitalized checks if a string starts with a capital letter
func isCapitalized(s string) bool {
	if len(s) == 0 {
		return false
	}
	firstChar := s[0]
	return 'A' <= firstChar && firstChar <= 'Z'
}

// isAllCaps checks if a string is entirely in uppercase
func isAllCaps(s string) bool {
	for _, c := range s {
		if 'a' <= c && c <= 'z' {
			return false
		}
	}
	return true
}

// capitalize capitalizes the first letter of a string
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// splitPunctuation separates a word from its trailing punctuation
func splitPunctuation(word string) (string, string) {
	for i := len(word) - 1; i >= 0; i-- {
		if isLetter(word[i]) || isDigit(word[i]) {
			if i == len(word)-1 {
				return word, ""
			}
			return word[:i+1], word[i+1:]
		}
	}
	return word, ""
}

// isLetter checks if a byte represents a letter
func isLetter(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

// isDigit checks if a byte represents a digit
func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
