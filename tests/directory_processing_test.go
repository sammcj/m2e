package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/fileutil"
)

func TestIsTextFile(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "m2e-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	testCases := []struct {
		name         string
		filename     string
		content      string
		expectedText bool
	}{
		{
			name:         "Markdown file",
			filename:     "test.md",
			content:      "# Heading\n\nThis is a markdown file with color text.",
			expectedText: true,
		},
		{
			name:         "Text file",
			filename:     "test.txt",
			content:      "This is a plain text file with color words.",
			expectedText: true,
		},
		{
			name:         "JSON file",
			filename:     "config.json",
			content:      `{"color": "red", "flavor": "vanilla"}`,
			expectedText: true,
		},
		{
			name:         "File without extension",
			filename:     "README",
			content:      "This is a README file with color information.",
			expectedText: true,
		},
		{
			name:         "Binary file (simulated)",
			filename:     "test.exe",
			content:      "fake binary", // Will be treated as binary due to extension
			expectedText: false,
		},
		{
			name:         "Image file",
			filename:     "image.png",
			content:      "fake png content",
			expectedText: false,
		},
		{
			name:         "Binary content with null bytes",
			filename:     "binary.dat",
			content:      "text\x00binary\x00content",
			expectedText: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tc.filename)

			err := os.WriteFile(filePath, []byte(tc.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			isText, err := fileutil.IsTextFile(filePath)
			if err != nil {
				t.Fatalf("IsTextFile failed: %v", err)
			}

			if isText != tc.expectedText {
				t.Errorf("Expected IsTextFile(%s) = %v, got %v", tc.filename, tc.expectedText, isText)
			}
		})
	}
}

func TestFindTextFiles(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "m2e-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test directory structure
	testFiles := map[string]string{
		"README.md":        "# Main README with color text",
		"docs/guide.txt":   "User guide with color information",
		"docs/api.md":      "API documentation with color examples",
		"config.json":      `{"color": "blue"}`,
		"image.png":        "fake png content", // Should be ignored
		"binary.exe":       "fake binary",      // Should be ignored
		".hidden.txt":      "hidden file",      // Should be ignored
		"subdir/notes.txt": "Notes with color words",
		"subdir/data.csv":  "name,color\nJohn,red",
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

	// Test finding text files
	files, err := fileutil.FindTextFiles(tempDir)
	if err != nil {
		t.Fatalf("FindTextFiles failed: %v", err)
	}

	// Check expected files are found
	expectedFiles := []string{
		"README.md",
		"docs/guide.txt",
		"docs/api.md",
		"config.json",
		"subdir/notes.txt",
		"subdir/data.csv",
	}

	if len(files) != len(expectedFiles) {
		t.Errorf("Expected %d files, got %d", len(expectedFiles), len(files))
	}

	foundFiles := make(map[string]bool)
	for _, file := range files {
		foundFiles[file.RelativePath] = true
	}

	for _, expected := range expectedFiles {
		if !foundFiles[expected] {
			t.Errorf("Expected file %s not found", expected)
		}
	}

	// Check that binary and hidden files are not included
	unexpectedFiles := []string{
		"image.png",
		"binary.exe",
		".hidden.txt",
	}

	for _, unexpected := range unexpectedFiles {
		if foundFiles[unexpected] {
			t.Errorf("Unexpected file %s was found", unexpected)
		}
	}
}

func TestFindTextFilesSingleFile(t *testing.T) {
	// Create temporary file
	tempDir, err := os.MkdirTemp("", "m2e-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	testFile := filepath.Join(tempDir, "test.md")
	err = os.WriteFile(testFile, []byte("# Test\n\nColor text here."), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with single file
	files, err := fileutil.FindTextFiles(testFile)
	if err != nil {
		t.Fatalf("FindTextFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if files[0].RelativePath != "test.md" {
		t.Errorf("Expected relative path 'test.md', got '%s'", files[0].RelativePath)
	}

	if files[0].Path != testFile {
		t.Errorf("Expected path '%s', got '%s'", testFile, files[0].Path)
	}
}

func TestReadWriteFileContent(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "m2e-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := "This file contains color text and flavor information."

	// Test write
	err = fileutil.WriteFileContent(testFile, originalContent)
	if err != nil {
		t.Fatalf("WriteFileContent failed: %v", err)
	}

	// Test read
	readContent, err := fileutil.ReadFileContent(testFile)
	if err != nil {
		t.Fatalf("ReadFileContent failed: %v", err)
	}

	if readContent != originalContent {
		t.Errorf("Content mismatch. Expected: %s, Got: %s", originalContent, readContent)
	}
}

func TestReadFileContentLargeFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "m2e-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	testFile := filepath.Join(tempDir, "large.txt")

	// Create a file larger than 10MB
	largeContent := strings.Repeat("This is color text. ", 1024*1024) // ~20MB
	err = os.WriteFile(testFile, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	// Test that reading large file fails
	_, err = fileutil.ReadFileContent(testFile)
	if err == nil {
		t.Error("Expected error when reading large file, but got none")
	}

	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("Expected 'too large' error, got: %v", err)
	}
}

func TestGetFileStats(t *testing.T) {
	files := []fileutil.FileInfo{
		{Path: "test1.txt", RelativePath: "test1.txt", Size: 1024},
		{Path: "test2.md", RelativePath: "test2.md", Size: 2048},
		{Path: "data.json", RelativePath: "data.json", Size: 512},
		{Path: "notes", RelativePath: "notes", Size: 256},
	}

	stats := fileutil.GetFileStats(files)

	if stats["total_files"] != 4 {
		t.Errorf("Expected total_files = 4, got %v", stats["total_files"])
	}

	if stats["total_size"] != int64(3840) {
		t.Errorf("Expected total_size = 3840, got %v", stats["total_size"])
	}

	extensions, ok := stats["extensions"].(map[string]int)
	if !ok {
		t.Fatalf("Expected extensions to be map[string]int")
	}

	expectedExtensions := map[string]int{
		".txt":           1,
		".md":            1,
		".json":          1,
		"(no extension)": 1,
	}

	for ext, count := range expectedExtensions {
		if extensions[ext] != count {
			t.Errorf("Expected extension %s count = %d, got %d", ext, count, extensions[ext])
		}
	}
}

func TestDirectoryWalkWithPermissionIssues(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "m2e-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a readable file
	readableFile := filepath.Join(tempDir, "readable.txt")
	err = os.WriteFile(readableFile, []byte("readable content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create readable file: %v", err)
	}

	// Create an unreadable directory
	unreadableDir := filepath.Join(tempDir, "unreadable")
	err = os.Mkdir(unreadableDir, 0000)
	if err != nil {
		t.Fatalf("Failed to create unreadable directory: %v", err)
	}
	defer func() { _ = os.Chmod(unreadableDir, 0755) }() // Cleanup

	// Create a file in the unreadable directory
	unreadableFile := filepath.Join(unreadableDir, "hidden.txt")
	_ = os.Chmod(unreadableDir, 0755) // Temporarily make it writable
	err = os.WriteFile(unreadableFile, []byte("hidden content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file in unreadable directory: %v", err)
	}
	_ = os.Chmod(unreadableDir, 0000) // Make it unreadable again

	// Test that FindTextFiles handles permission errors gracefully
	files, err := fileutil.FindTextFiles(tempDir)
	if err != nil {
		t.Fatalf("FindTextFiles should handle permission errors gracefully: %v", err)
	}

	// Should find at least the readable file
	found := false
	for _, file := range files {
		if file.RelativePath == "readable.txt" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Should have found the readable file despite permission errors")
	}
}
