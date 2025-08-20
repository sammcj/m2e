// Package tests provides comprehensive testing for principal/principle contextual conversion
package tests

import (
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestPrincipalPrincipleConversion(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	// Enable contextual word detection
	conv.SetContextualWordDetectionEnabled(true)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Test cases where "principal" should be corrected to "principle"
		{
			name:     "least privilege principle",
			input:    "Follow the principal of least privilege.",
			expected: "Follow the principle of least privilege.",
		},
		{
			name:     "least privileged principle",
			input:    "Follow the principal of least privileged security.",
			expected: "Follow the principle of least privileged security.",
		},
		{
			name:     "security principles (guidelines)",
			input:    "Our security principals include encryption and access control.",
			expected: "Our security principles include encryption and access control.",
		},
		{
			name:     "design principles",
			input:    "These design principals guide our architecture.",
			expected: "These design principles guide our architecture.",
		},
		{
			name:     "fundamental principles",
			input:    "We follow fundamental principals of software engineering.",
			expected: "We follow fundamental principles of software engineering.",
		},
		{
			name:     "core principles",
			input:    "The core principals of our team are collaboration and quality.",
			expected: "The core principles of our team are collaboration and quality.",
		},
		{
			name:     "engineering principles",
			input:    "Engineering principals should be well-defined.",
			expected: "Engineering principles should be well-defined.",
		},
		{
			name:     "SOLID principles",
			input:    "Follow SOLID principals in your code.",
			expected: "Follow SOLID principles in your code.",
		},

		// Test cases where "principle" should be corrected to "principal"
		{
			name:     "AWS IAM principal",
			input:    "The AWS IAM principle has S3 access.",
			expected: "The AWS IAM principal has S3 access.",
		},
		{
			name:     "service principal",
			input:    "Create a service principle for the application.",
			expected: "Create a service principal for the application.",
		},
		{
			name:     "user principal",
			input:    "The user principle authentication failed.",
			expected: "The user principal authentication failed.",
		},
		{
			name:     "principal ARN",
			input:    "Check the principle ARN in the policy.",
			expected: "Check the principal ARN in the policy.",
		},
		{
			name:     "authentication principal",
			input:    "Authentication principles must be configured.",
			expected: "Authentication principals must be configured.",
		},
		{
			name:     "database principal",
			input:    "The database principles need permissions.",
			expected: "The database principals need permissions.",
		},
		{
			name:     "principal name",
			input:    "Set the principle name in the configuration.",
			expected: "Set the principal name in the configuration.",
		},
		{
			name:     "principal ID",
			input:    "The principle ID is used for identification.",
			expected: "The principal ID is used for identification.",
		},
		{
			name:     "loan principal",
			input:    "Pay down the loan principles first.",
			expected: "Pay down the loan principals first.",
		},
		{
			name:     "principal amount",
			input:    "The principle amount is £10,000.",
			expected: "The principal amount is £10,000.",
		},

		// Test cases that should NOT be converted (ambiguous contexts)
		{
			name:     "school principal - no change",
			input:    "The school principal called a meeting.",
			expected: "The school principal called a meeting.",
		},
		{
			name:     "principal concern - no change",
			input:    "My principal concern is safety.",
			expected: "My principal concern is safety.",
		},
		{
			name:     "general principle - no change",
			input:    "This is a general principle of physics.",
			expected: "This is a general principle of physics.",
		},
		{
			name:     "principle vs principal - no change when ambiguous",
			input:    "The principle decided to change the policy.",
			expected: "The principle decided to change the policy.",
		},

		// Test case sensitivity
		{
			name:     "capitalised least privilege",
			input:    "Follow the Principal of Least Privilege.",
			expected: "Follow the Principle of Least Privilege.",
		},
		{
			name:     "uppercase AWS IAM",
			input:    "THE AWS IAM PRINCIPLE HAS ACCESS.",
			expected: "THE AWS IAM PRINCIPAL HAS ACCESS.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := conv.ConvertToBritish(test.input, false)
			if result != test.expected {
				t.Errorf("Expected: %q, Got: %q", test.expected, result)
			}
		})
	}
}

func TestPrincipalPrincipleContextualDetection(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	detector := conv.GetContextualWordDetector()
	if detector == nil {
		t.Fatal("Contextual word detector is nil")
	}

	// Test that principal/principle are in supported words
	supportedWords := detector.SupportedWords()
	hasPrincipal := false
	hasPrinciple := false

	for _, word := range supportedWords {
		if strings.ToLower(word) == "principal" {
			hasPrincipal = true
		}
		if strings.ToLower(word) == "principle" {
			hasPrinciple = true
		}
	}

	if !hasPrincipal {
		t.Error("'principal' should be in supported words")
	}
	if !hasPrinciple {
		t.Error("'principle' should be in supported words")
	}

	// Test detection of specific contexts
	testTexts := []struct {
		text            string
		expectedMatches int
		description     string
	}{
		{
			text:            "Follow the principal of least privilege in your security design.",
			expectedMatches: 1,
			description:     "Should detect 'principal of least privilege' pattern",
		},
		{
			text:            "The AWS IAM principle should have minimal permissions.",
			expectedMatches: 1,
			description:     "Should detect 'AWS IAM principle' pattern",
		},
		{
			text:            "The school principal is a good person with strong principles.",
			expectedMatches: 0,
			description:     "Should not detect ambiguous contexts",
		},
	}

	for _, test := range testTexts {
		t.Run(test.description, func(t *testing.T) {
			matches := detector.DetectWords(test.text)
			if len(matches) != test.expectedMatches {
				t.Errorf("Expected %d matches, got %d for text: %q",
					test.expectedMatches, len(matches), test.text)
				for i, match := range matches {
					t.Logf("Match %d: %q -> %q (confidence: %.2f)",
						i, match.OriginalWord, match.Replacement, match.Confidence)
				}
			}
		})
	}
}

func TestPrincipalPrincipleConfiguration(t *testing.T) {
	config := converter.GetDefaultContextualWordConfig()

	// Test that principal and principle are configured
	principalConfig, hasPrincipal := config.WordConfigs["principal"]
	if !hasPrincipal {
		t.Error("'principal' should be in word configs")
	}

	principleConfig, hasPrinciple := config.WordConfigs["principle"]
	if !hasPrinciple {
		t.Error("'principle' should be in word configs")
	}

	// Test that semantic variants are defined
	if len(principalConfig.SemanticVariants) == 0 {
		t.Error("'principal' should have semantic variants defined")
	}

	if len(principleConfig.SemanticVariants) == 0 {
		t.Error("'principle' should have semantic variants defined")
	}

	// Test specific patterns exist (with capture groups)
	expectedPrincipalPatterns := []string{
		`(?i)(principal)\s+of\s+least\s+privile?ge?d?`,
		`(?i)security\s+(principals?)\b`,
		`(?i)design\s+(principals?)\b`,
	}

	for _, pattern := range expectedPrincipalPatterns {
		if _, exists := principalConfig.SemanticVariants[pattern]; !exists {
			t.Errorf("Expected pattern not found in principal config: %s", pattern)
		}
	}

	expectedPrinciplePatterns := []string{
		`(?i)AWS\s+IAM\s+(principles?)\b`,
		`(?i)service\s+(principles?)\b`,
		`(?i)(principle)\s+ARN\b`,
	}

	for _, pattern := range expectedPrinciplePatterns {
		if _, exists := principleConfig.SemanticVariants[pattern]; !exists {
			t.Errorf("Expected pattern not found in principle config: %s", pattern)
		}
	}
}

func TestPrincipalPrincipleRealWorldExamples(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	conv.SetContextualWordDetectionEnabled(true)

	realWorldTests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "security documentation",
			input: `Security Guidelines:
1. Follow the principal of least privilege
2. Ensure service principals have minimal permissions
3. Review security principals regularly`,
			expected: `Security Guidelines:
1. Follow the principle of least privilege
2. Ensure service principals have minimal permissions
3. Review security principles regularly`,
		},
		{
			name: "AWS IAM policy documentation",
			input: `IAM Policy Configuration:
- The principle ARN must be specified
- Each service principle needs S3 access
- AWS IAM principles should be reviewed monthly`,
			expected: `IAM Policy Configuration:
- The principal ARN must be specified
- Each service principal needs S3 access
- AWS IAM principals should be reviewed monthly`,
		},
		{
			name: "software engineering principles",
			input: `Our engineering principals include:
- DRY principals (Don't Repeat Yourself)
- SOLID principals for object-oriented design
- These fundamental principals guide our development`,
			expected: `Our engineering principles include:
- DRY principles (Don't Repeat Yourself)
- SOLID principles for object-oriented design
- These fundamental principles guide our development`,
		},
	}

	for _, test := range realWorldTests {
		t.Run(test.name, func(t *testing.T) {
			result := conv.ConvertToBritish(test.input, false)
			if result != test.expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", test.expected, result)
			}
		})
	}
}
