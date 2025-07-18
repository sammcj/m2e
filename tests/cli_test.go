package tests

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIUsage(t *testing.T) {
	// Use the existing built CLI
	cliPath := filepath.Join("..", "build", "bin", "m2e-cli")

	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
	}{
		{
			name:     "No arguments shows usage",
			args:     []string{},
			expected: "m2e-cli - Convert American English to British English",
			wantErr:  true, // Usage is printed to stderr and exits with code 1
		},
		{
			name:     "Help flag shows usage",
			args:     []string{"-h"},
			expected: "m2e-cli - Convert American English to British English",
			wantErr:  false, // Help exits with code 0
		},
		{
			name:     "Direct text conversion",
			args:     []string{"color", "and", "flavor"},
			expected: "colour and flavour",
			wantErr:  false,
		},
		{
			name:     "Single quoted argument",
			args:     []string{"American color and flavor"},
			expected: "American colour and flavour",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(cliPath, tt.args...)

			output, err := cmd.CombinedOutput()

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v\nOutput: %s", err, string(output))
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, outputStr)
			}
		})
	}
}

func TestCLIUnitConversion(t *testing.T) {
	// Use the existing built CLI
	cliPath := filepath.Join("..", "build", "bin", "m2e-cli")

	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
	}{
		{
			name:     "Unit conversion enabled",
			args:     []string{"-units", "The room is 12 feet wide and weighs 100 pounds."},
			expected: "The room is 3.7 metres wide and weighs 45.4 kg.",
			wantErr:  false,
		},
		{
			name:     "Unit conversion disabled",
			args:     []string{"The room is 12 feet wide and weighs 100 pounds."},
			expected: "The room is 12 feet wide and weighs 100 pounds.",
			wantErr:  false,
		},
		{
			name:     "Spelling and unit conversion",
			args:     []string{"-units", "The color of the 5-foot fence is gray."},
			expected: "The colour of the 1.5-metre fence is grey.",
			wantErr:  false,
		},
		{
			name:     "Help includes unit options",
			args:     []string{"-help"},
			expected: "-units",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(cliPath, tt.args...)

			output, err := cmd.CombinedOutput()

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v\nOutput: %s", err, string(output))
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, outputStr)
			}
		})
	}
}
