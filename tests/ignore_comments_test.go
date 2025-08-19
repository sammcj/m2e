package tests

import (
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/converter"
)

func TestIgnoreCommentDetection(t *testing.T) {
	processor := converter.NewCommentIgnoreProcessor()

	testCases := []struct {
		name          string
		text          string
		expectedCount int
		expectedType  converter.IgnoreDirective
		description   string
	}{
		{
			name:          "Single line ignore with //",
			text:          "// m2e-ignore\nThis is some american text with color and flavor.",
			expectedCount: 1,
			expectedType:  converter.IgnoreLine,
			description:   "Should detect m2e-ignore in C-style comment",
		},
		{
			name:          "Single line ignore with #",
			text:          "# m2e-ignore\nThis has american spelling like color and flavor.",
			expectedCount: 1,
			expectedType:  converter.IgnoreLine,
			description:   "Should detect m2e-ignore in hash comment",
		},
		{
			name:          "File ignore with //",
			text:          "// m2e-ignore-file\nThis entire file should be ignored with color and flavor.",
			expectedCount: 1,
			expectedType:  converter.IgnoreFile,
			description:   "Should detect file-level ignore",
		},
		{
			name:          "Next line ignore",
			text:          "// m2e-ignore-next\nThis line should be ignored with color and flavor.\nThis line should be processed.",
			expectedCount: 1,
			expectedType:  converter.IgnoreNext,
			description:   "Should detect next-line ignore",
		},
		{
			name:          "Case insensitive",
			text:          "// M2E-IGNORE\nThis should be ignored despite uppercase.",
			expectedCount: 1,
			expectedType:  converter.IgnoreLine,
			description:   "Should be case insensitive",
		},
		{
			name:          "Multiple comment types",
			text:          "// m2e-ignore\nSome text\n# m2e-ignore-next\nMore text\n<!-- m2e-ignore-file -->",
			expectedCount: 3,
			expectedType:  converter.IgnoreFile, // Last one found
			description:   "Should detect multiple ignore types",
		},
		{
			name:          "No ignore comments",
			text:          "This is normal text with color and flavor.\n// This is just a regular comment",
			expectedCount: 0,
			expectedType:  converter.IgnoreNone,
			description:   "Should find no ignore directives",
		},
		{
			name:          "HTML comment ignore",
			text:          "<!-- m2e-ignore -->\n<p>This has american spelling like color.</p>",
			expectedCount: 1,
			expectedType:  converter.IgnoreLine,
			description:   "Should detect ignore in HTML comment",
		},
		{
			name:          "SQL comment ignore",
			text:          "-- m2e-ignore\nSELECT * FROM color_table;",
			expectedCount: 1,
			expectedType:  converter.IgnoreLine,
			description:   "Should detect ignore in SQL comment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches := processor.ProcessIgnoreComments(tc.text)

			if len(matches) != tc.expectedCount {
				t.Errorf("Expected %d ignore matches, got %d", tc.expectedCount, len(matches))
			}

			if tc.expectedCount > 0 {
				// Check that we found the expected type (check the last/most restrictive one)
				hasExpectedType := false
				for _, match := range matches {
					if match.Directive == tc.expectedType {
						hasExpectedType = true
						break
					}
				}
				if !hasExpectedType {
					t.Errorf("Expected to find ignore type %v, but didn't find it in matches", tc.expectedType)
				}
			}
		})
	}
}

func TestIgnoreCommentIntegration(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	testCases := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name: "Line ignore with //",
			input: `// m2e-ignore
This line has color and flavor and should not be converted.
This line has color and flavor and should be converted.`,
			expected: `// m2e-ignore
This line has colour and flavour and should not be converted.
This line has colour and flavour and should be converted.`,
			description: "Comment line should be ignored, other lines should be converted",
		},
		{
			name: "Next line ignore",
			input: `// m2e-ignore-next
This line has color and flavor and should not be converted.
This line has color and flavor and should be converted.`,
			expected: `// m2e-ignore-next
This line has color and flavor and should not be converted.
This line has colour and flavour and should be converted.`,
			description: "Line after ignore-next should be ignored",
		},
		{
			name: "File ignore",
			input: `// m2e-ignore-file
This file has color and flavor and should not be converted.
This line also has color and flavor and should not be converted.`,
			expected: `// m2e-ignore-file
This file has color and flavor and should not be converted.
This line also has color and flavor and should not be converted.`,
			description: "Entire file should be ignored",
		},
		{
			name: "Hash comment ignore",
			input: `# m2e-ignore
This line has color and flavor.
# Regular comment
This line has color and flavor.`,
			expected: `# m2e-ignore
This line has colour and flavour.
# Regular comment
This line has colour and flavour.`,
			description: "Hash comment ignore should work",
		},
		{
			name: "Multiple ignores",
			input: `// m2e-ignore
First ignored line with color.
// m2e-ignore-next
Second ignored line with flavor.
Normal line with color and flavor.`,
			expected: `// m2e-ignore
First ignored line with colour.
// m2e-ignore-next
Second ignored line with flavor.
Normal line with colour and flavour.`,
			description: "Multiple ignore directives should work",
		},
		{
			name: "Mixed with contextual words",
			input: `// m2e-ignore
I need a license to practice medicine.
I need a license to practice medicine.`,
			expected: `// m2e-ignore
I need a licence to practise medicine.
I need a licence to practise medicine.`,
			description: "Ignore should work with contextual word detection",
		},
		{
			name: "HTML comment ignore",
			input: `<!-- m2e-ignore -->
<p>This has color and flavor.</p>
<p>This has color and flavor.</p>`,
			expected: `<!-- m2e-ignore -->
<p>This has colour and flavor.</p>
<p>This has colour and flavor.</p>`,
			description: "HTML comment ignore should work",
		},
		{
			name: "No ignore comments",
			input: `// Just a regular comment
This line has color and flavor.
This line also has color and flavor.`,
			expected: `// Just a regular comment
This line has colour and flavour.
This line also has colour and flavour.`,
			description: "Without ignore comments, everything should be converted",
		},
		{
			name: "Case insensitive ignore",
			input: `// M2E-IGNORE
This line has COLOR and FLAVOR.
This line has color and flavor.`,
			expected: `// M2E-IGNORE
This line has Colour and Flavour.
This line has colour and flavour.`,
			description: "Case insensitive ignore should work",
		},
		{
			name: "SQL comment ignore",
			input: `-- m2e-ignore
SELECT color, flavor FROM american_table;
SELECT color, flavor FROM british_table;`,
			expected: `-- m2e-ignore
SELECT colour, flavour FROM american_table;
SELECT colour, flavour FROM british_table;`,
			description: "SQL comment ignore should work",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tc.input, true)

			// Normalise line endings for comparison
			result = strings.ReplaceAll(result, "\r\n", "\n")
			expected := strings.ReplaceAll(tc.expected, "\r\n", "\n")

			if result != expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
			}
		})
	}
}

func TestIgnoreStatsAndDirectives(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	testText := `// m2e-ignore-file
Some text with color.
// m2e-ignore
More text with flavor.
# m2e-ignore-next
Even more text.`

	// Test getting ignore directives
	directives := conv.GetIgnoreDirectives(testText)
	if len(directives) != 3 {
		t.Errorf("Expected 3 ignore directives, got %d", len(directives))
	}

	// Test getting ignore stats
	stats := conv.GetIgnoreStats(testText)
	expectedStats := map[string]int{
		"ignore-file": 1,
		"ignore-line": 1,
		"ignore-next": 1,
	}

	for statName, expectedCount := range expectedStats {
		if stats[statName] != expectedCount {
			t.Errorf("Expected %d %s directives, got %d", expectedCount, statName, stats[statName])
		}
	}
}

func TestIgnoreWithoutIgnoreComments(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	input := `// m2e-ignore
This line has color and flavor and should not be converted normally.
This line has color and flavor.`

	// Test normal conversion (with ignore comments)
	normalResult := conv.ConvertToBritish(input, true)
	expectedNormal := `// m2e-ignore
This line has colour and flavour and should not be converted normally.
This line has colour and flavour.`

	if normalResult != expectedNormal {
		t.Errorf("Normal conversion failed.\nExpected:\n%s\n\nGot:\n%s", expectedNormal, normalResult)
	}

	// Test bypassing ignore comments
	bypassResult := conv.ConvertToBritishWithoutIgnores(input, true)
	expectedBypass := `// m2e-ignore
This line has colour and flavour and should not be converted normally.
This line has colour and flavour.`

	if bypassResult != expectedBypass {
		t.Errorf("Bypass ignore conversion failed.\nExpected:\n%s\n\nGot:\n%s", expectedBypass, bypassResult)
	}
}

func TestIgnoreCommentsWithCodeAwareness(t *testing.T) {
	conv, err := converter.NewConverter()
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	testCases := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name: "Ignore in Go code",
			input: `package main

// m2e-ignore
// This comment talks about color and flavor
func main() {
    // This comment about color should be converted
    fmt.Println("color")  // But not the string
}`,
			expected: `package main

// m2e-ignore
// This comment talks about colour and flavour
func main() {
    // This comment about colour should be converted
    fmt.Println("color")  // But not the string
}`,
			description: "Should ignore specific comment but convert others in code",
		},
		{
			name: "Ignore next in JavaScript",
			input: `// m2e-ignore-next
// This comment about color should be ignored
function colorFunc() {
    // This comment about color should be converted
    return "color"; // String stays unchanged
}`,
			expected: `// m2e-ignore-next
// This comment about color should be ignored
function colorFunc() {
    // This comment about colour should be converted
    return "color"; // String stays unchanged
}`,
			description: "Should work with code-aware processing",
		},
		{
			name: "File ignore in Python",
			input: `# m2e-ignore-file
# This entire Python file should be ignored
def color_function():
    """This docstring about color should not be converted"""
    # This comment about flavor should not be converted
    return "color"  # This should not be converted`,
			expected: `# m2e-ignore-file
# This entire Python file should be ignored
def color_function():
    """This docstring about color should not be converted"""
    # This comment about flavor should not be converted
    return "color"  # This should not be converted`,
			description: "File ignore should override code-awareness",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := conv.ConvertToBritish(tc.input, true)

			// Normalise line endings for comparison
			result = strings.ReplaceAll(result, "\r\n", "\n")
			expected := strings.ReplaceAll(tc.expected, "\r\n", "\n")

			if result != expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
			}
		})
	}
}
