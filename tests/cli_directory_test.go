package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIDirectoryProcessing(t *testing.T) {
	// Build the CLI first
	cmd := exec.Command("go", "build", "-o", "../build/bin/m2e-test", "../cmd/m2e-cli")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() { _ = os.Remove("../build/bin/m2e-test") }()

	// Helper function to create test files with American English
	createTestFiles := func() string {
		tempDir, err := os.MkdirTemp("", "m2e-dir-test-")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}

		testFiles := map[string]string{
			"README.md": `# Test Project

This project uses color schemes and has a flavor profile.
The width is 12 feet and temperature is 75°F.`,
			"docs/guide.txt": `User Guide

Configure the color settings for optimal flavor.
Room dimensions: 6 inches by 8 inches.`,
			"config.json": `{
  "color": "red",
  "flavor": "vanilla",
  "temp": "68°F"
}`,
			"notes.txt":  `Notes about color and flavor preferences.`,
			"binary.exe": "fake binary content", // Should be ignored
			".hidden":    "hidden file",         // Should be ignored
		}

		for relPath, content := range testFiles {
			fullPath := filepath.Join(tempDir, relPath)

			// Create directory if needed
			dir := filepath.Dir(fullPath)
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}

			err = os.WriteFile(fullPath, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", relPath, err)
			}
		}
		return tempDir
	}

	testCases := []struct {
		name     string
		args     func(dir string) []string
		contains []string
		exitCode int
	}{
		{
			name: "Directory stats mode",
			args: func(dir string) []string { return []string{"-stats", dir} },
			contains: []string{
				"📊 **Words processed:**",
				"🔤 **Spelling changes:**",
			},
			exitCode: 0,
		},
		{
			name: "Directory stats mode with units",
			args: func(dir string) []string { return []string{"-stats", "-units", dir} },
			contains: []string{
				"📊 **Words processed:**",
				"🔤 **Spelling changes:**",
				"📏 **Unit conversions:**",
			},
			exitCode: 0,
		},
		{
			name: "Directory with exit-on-change",
			args: func(dir string) []string { return []string{"-exit-on-change", dir} },
			contains: []string{
				"Found",
				"text file(s)",
				"Processing:",
			},
			exitCode: 1, // Should exit with 1 because changes are detected
		},
		{
			name: "Directory default mode (in-place editing)",
			args: func(dir string) []string { return []string{dir} },
			contains: []string{
				"Found",
				"text file(s)",
				"Processing:",
				"Updated:",
			},
			exitCode: 0,
		},
		{
			name: "Directory diff mode",
			args: func(dir string) []string { return []string{"-diff", dir} },
			contains: []string{
				"Found",
				"text file(s)",
				"Processing:",
				"=== README.md ===",
				"--- README.md.orig",
			},
			exitCode: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh test directory for each test case
			tempDir := createTestFiles()
			defer func() { _ = os.RemoveAll(tempDir) }()

			cmd := exec.Command("../build/bin/m2e-test", tc.args(tempDir)...)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check exit code
			if tc.exitCode == 0 {
				if err != nil {
					t.Errorf("Expected success but got error: %v\nStderr: %s", err, stderr.String())
				}
			} else {
				if err == nil {
					t.Errorf("Expected exit code %d but command succeeded", tc.exitCode)
				} else if exitError, ok := err.(*exec.ExitError); ok {
					if exitError.ExitCode() != tc.exitCode {
						t.Errorf("Expected exit code %d, got %d", tc.exitCode, exitError.ExitCode())
					}
				}
			}

			output := stdout.String()

			// Check that expected strings are present
			for _, expected := range tc.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain '%s'\nActual output:\n%s", expected, output)
				}
			}
		})
	}
}

func TestCLIDirectoryWithNoTextFiles(t *testing.T) {
	// Build the CLI first
	cmd := exec.Command("go", "build", "-o", "../build/bin/m2e-test", "../cmd/m2e-cli")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() { _ = os.Remove("../build/bin/m2e-test") }()

	// Create temporary directory with only binary files
	tempDir, err := os.MkdirTemp("", "m2e-empty-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create only binary files
	binaryFiles := map[string]string{
		"image.png":  "fake png content",
		"video.mp4":  "fake video content",
		"binary.exe": "fake binary content",
	}

	for filename, content := range binaryFiles {
		fullPath := filepath.Join(tempDir, filename)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create binary file %s: %v", filename, err)
		}
	}

	// Test directory processing with no text files
	cmd = exec.Command("../build/bin/m2e-test", tempDir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Errorf("Command should succeed even with no text files: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "No text files found") {
		t.Errorf("Expected 'No text files found' message, got: %s", output)
	}
}

func TestCLISingleFileVsDirectory(t *testing.T) {
	// Build the CLI first
	cmd := exec.Command("go", "build", "-o", "../build/bin/m2e-test", "../cmd/m2e-cli")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() { _ = os.Remove("../build/bin/m2e-test") }()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "m2e-single-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a single test file
	testFile := filepath.Join(tempDir, "test.md")
	content := "# Test\n\nThis file contains color text and flavor information."
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test single file processing in stats mode
	cmd = exec.Command("../build/bin/m2e-test", "-stats", testFile)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		t.Errorf("Single file processing failed: %v", err)
	}

	output := stdout.String()

	// Should contain stats elements
	expectedElements := []string{
		"📊 **Words processed:**",
		"🔤 **Spelling changes:**",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Single file stats should contain '%s'\nActual output:\n%s", expected, output)
		}
	}
}

func TestCLIDirectoryInPlaceEditing(t *testing.T) {
	// Build the CLI first
	cmd := exec.Command("go", "build", "-o", "../build/bin/m2e-test", "../cmd/m2e-cli")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() { _ = os.Remove("../build/bin/m2e-test") }()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "m2e-inplace-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test file with American English
	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := "This file contains color text and flavor information."
	err = os.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Process directory in normal mode (should edit files in-place)
	cmd = exec.Command("../build/bin/m2e-test", tempDir)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		t.Errorf("Directory processing failed: %v", err)
	}

	// Check that file was modified
	modifiedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modifiedString := string(modifiedContent)

	// Should contain British spellings
	if !strings.Contains(modifiedString, "colour") {
		t.Error("File should contain 'colour' after processing")
	}

	if !strings.Contains(modifiedString, "flavour") {
		t.Error("File should contain 'flavour' after processing")
	}

	// Should not contain American spellings
	if strings.Contains(modifiedString, "color ") { // Space to avoid matching "colour"
		t.Error("File should not contain 'color' after processing")
	}

	if strings.Contains(modifiedString, "flavor ") { // Space to avoid matching "flavour"
		t.Error("File should not contain 'flavor' after processing")
	}

	// Check output indicates file was updated
	output := stdout.String()
	if !strings.Contains(output, "Updated: test.txt") {
		t.Errorf("Output should indicate file was updated, got: %s", output)
	}
}
