package tests

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCLIOutputModes(t *testing.T) {
	// Build the CLI first
	cmd := exec.Command("go", "build", "-o", "../build/bin/m2e-test", "../cmd/m2e")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() { _ = os.Remove("../build/bin/m2e-test") }()

	testCases := []struct {
		name        string
		input       string
		args        []string
		expectError bool
		expectExit1 bool
		contains    []string
		notContains []string
	}{
		{
			name:  "Default mode with changes",
			input: "I love color and humor.",
			args:  []string{},
			contains: []string{
				"--- stdin.orig",
				"+++ stdin",
				"I love colour and humour",
				"📊 **Words processed:** 5",
				"🔤 **Spelling changes needed:** 2",
			},
		},
		{
			name:  "Diff mode only",
			input: "I love color.",
			args:  []string{"-diff"},
			contains: []string{
				"--- stdin.orig",
				"+++ stdin",
				"@@", // Unified diff hunk header
				"-I love color.",
				"+I love colour.",
			},
			notContains: []string{
				"📊 **Words processed:**",
			},
		},
		{
			name:  "Raw mode only",
			input: "I love color.",
			args:  []string{"-raw"},
			contains: []string{
				"I love colour.",
			},
			notContains: []string{
				"--- stdin.orig",
				"📊 **Words processed:**",
			},
		},
		{
			name:  "Stats mode only",
			input: "I love color.",
			args:  []string{"-stats"},
			contains: []string{
				"📊 **Words processed:** 3",
				"🔤 **Spelling changes needed:** 1",
			},
			notContains: []string{
				"--- stdin.orig",
				"I love colour.",
			},
		},
		{
			name:        "Exit on change",
			input:       "I love color.",
			args:        []string{"-exit-on-change"},
			expectExit1: true,
			contains: []string{
				"📊 **Words processed:** 3",
				"🔤 **Spelling changes needed:** 1",
			},
		},
		{
			name:  "No changes",
			input: "I love colour and humour.",
			args:  []string{},
			contains: []string{
				"I love colour and humour",
				"📊 **Words processed:** 5",
				"🔤 **Spelling changes needed:** 0",
			},
			notContains: []string{
				"--- stdin.orig", // No diff shown when no changes
			},
		},
		{
			name:  "Units conversion",
			input: "The room is 12 feet wide.",
			args:  []string{"-units"},
			contains: []string{
				"3.7 metres",
				"📏 **Unit conversions needed:** 1",
			},
		},
		{
			name:  "Diff inline mode",
			input: "I love color.",
			args:  []string{"-diff-inline"},
			contains: []string{
				"I love colo", // Should contain the base text
				"u",           // Should contain the added 'u' character
				"r.",          // Should contain the ending
			},
			notContains: []string{
				"📊 **Words processed:**",
				"--- stdin.orig", // Should not contain unified diff headers
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("../build/bin/m2e-test", tc.args...)
			cmd.Stdin = strings.NewReader(tc.input)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check exit code expectations
			if tc.expectExit1 {
				if err == nil {
					t.Errorf("Expected exit code 1, but command succeeded")
				} else if exitError, ok := err.(*exec.ExitError); ok {
					if exitError.ExitCode() != 1 {
						t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
					}
				} else {
					t.Errorf("Expected ExitError, got %v", err)
				}
			} else if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, but command succeeded")
				}
			} else {
				if err != nil {
					t.Errorf("Expected success, but got error: %v\nStderr: %s", err, stderr.String())
				}
			}

			output := stdout.String()

			// Check that expected strings are present
			for _, expected := range tc.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain '%s'\nActual output:\n%s", expected, output)
				}
			}

			// Check that unexpected strings are not present
			for _, unexpected := range tc.notContains {
				if strings.Contains(output, unexpected) {
					t.Errorf("Output should not contain '%s'\nActual output:\n%s", unexpected, output)
				}
			}
		})
	}
}

func TestCLIOutputModeErrors(t *testing.T) {
	// Build the CLI first
	cmd := exec.Command("go", "build", "-o", "../build/bin/m2e-test", "../cmd/m2e")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() { _ = os.Remove("../build/bin/m2e-test") }()

	t.Run("Multiple output modes should error", func(t *testing.T) {
		cmd := exec.Command("../build/bin/m2e-test", "-diff", "-raw")
		cmd.Stdin = strings.NewReader("test")

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Error("Expected error when using multiple output mode flags")
		}

		if !strings.Contains(stderr.String(), "Only one output mode flag can be specified at a time") {
			t.Errorf("Expected specific error message, got: %s", stderr.String())
		}
	})

	t.Run("Output file with output mode should error", func(t *testing.T) {
		cmd := exec.Command("../build/bin/m2e-test", "-diff", "-o", "test.txt")
		cmd.Stdin = strings.NewReader("test")

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Error("Expected error when using -o with output mode flags")
		}

		if !strings.Contains(stderr.String(), "Output file (-o) cannot be used with output mode flags") {
			t.Errorf("Expected specific error message, got: %s", stderr.String())
		}
	})

	t.Run("Multiple diff modes should error", func(t *testing.T) {
		cmd := exec.Command("../build/bin/m2e-test", "-diff", "-diff-inline")
		cmd.Stdin = strings.NewReader("test")

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Error("Expected error when using multiple diff mode flags")
		}

		if !strings.Contains(stderr.String(), "Only one output mode flag can be specified at a time") {
			t.Errorf("Expected specific error message, got: %s", stderr.String())
		}
	})
}

func TestCLILegacyCompatibility(t *testing.T) {
	// Build the CLI first
	cmd := exec.Command("go", "build", "-o", "../build/bin/m2e-test", "../cmd/m2e")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() { _ = os.Remove("../build/bin/m2e-test") }()

	t.Run("Raw output mode works like old normal mode", func(t *testing.T) {
		cmd := exec.Command("../build/bin/m2e-test", "-raw")
		cmd.Stdin = strings.NewReader("I love color.")

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		err := cmd.Run()
		if err != nil {
			t.Errorf("Raw mode should work: %v", err)
		}

		output := strings.TrimSpace(stdout.String())
		expected := "I love colour."
		if output != expected {
			t.Errorf("Expected '%s', got '%s'", expected, output)
		}
	})
}
