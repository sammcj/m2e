package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sammcj/murican-to-english/pkg/converter"
)

// Helper function to create a temporary file with content
func createTempFile(t *testing.T, content string, suffix string) (string, func()) {
	tmpFile, err := os.CreateTemp("", "mcp_test_*"+suffix)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name())
		t.Fatalf("Failed to close temp file: %v", err)
	}

	cleanup := func() {
		_ = os.Remove(tmpFile.Name())
	}

	return tmpFile.Name(), cleanup
}

// Helper function to read file content
func readFileContent(t *testing.T, filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}
	return string(content)
}

// ConvertFileResult represents the result of file conversion
type ConvertFileResult struct {
	Success      bool
	Message      string
	ChangesCount int
	Error        error
}

// Helper function to simulate MCP convert_file tool logic
func simulateConvertFileTool(t *testing.T, filePath string) ConvertFileResult {
	conv, err := converter.NewConverter()
	if err != nil {
		return ConvertFileResult{
			Success: false,
			Error:   fmt.Errorf("failed to create converter: %v", err),
		}
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ConvertFileResult{
			Success: false,
			Message: fmt.Sprintf("File does not exist: %s", filePath),
		}
	}

	// Read the original file content
	originalContent, err := os.ReadFile(filePath)
	if err != nil {
		return ConvertFileResult{
			Success: false,
			Message: fmt.Sprintf("Error reading file %s: %v", filePath, err),
		}
	}

	// Convert the content based on file type (using the same logic as MCP tool)
	convertedContent := convertFileContent(conv, string(originalContent), filePath)

	// Check if there were any changes
	if string(originalContent) == convertedContent {
		return ConvertFileResult{
			Success:      true,
			Message:      fmt.Sprintf("File %s processed but no changes were needed - already in British English", filePath),
			ChangesCount: 0,
		}
	}

	// Write the converted content back to the file
	err = os.WriteFile(filePath, []byte(convertedContent), 0644)
	if err != nil {
		return ConvertFileResult{
			Success: false,
			Message: fmt.Sprintf("Error writing to file %s: %v", filePath, err),
		}
	}

	return ConvertFileResult{
		Success:      true,
		Message:      fmt.Sprintf("File %s completed processing to international/British English, the file has been updated.", filePath),
		ChangesCount: 1, // Simplified - just indicates changes were made
	}
}

// Helper functions from MCP server (duplicated for testing)
func isPlainTextFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	plainTextExtensions := []string{
		".txt", ".md", ".markdown", ".rst", ".text", ".doc", ".rtf",
		".tex", ".latex", ".org", ".adoc", ".asciidoc",
	}

	for _, plainExt := range plainTextExtensions {
		if ext == plainExt {
			return true
		}
	}
	return false
}

func convertFileContent(conv *converter.Converter, content, filePath string) string {
	if isPlainTextFile(filePath) {
		// For plain text files, use code-aware processing
		return conv.ProcessCodeAware(content, true)
	} else {
		// For code/config files, only convert comments
		return convertOnlyComments(conv, content)
	}
}

func convertOnlyComments(conv *converter.Converter, code string) string {
	comments := conv.ExtractComments(code, "")

	if len(comments) == 0 {
		return code
	}

	// Work backwards through comments so positions don't shift
	result := code
	for i := len(comments) - 1; i >= 0; i-- {
		comment := comments[i]

		// Get the original comment text
		originalComment := code[comment.Start:comment.End]

		// Convert only the comment content
		convertedComment := conv.ConvertToBritish(comment.Content, true)

		// Preserve the comment structure
		if len(originalComment) > len(comment.Content) {
			prefix := ""
			suffix := ""

			contentStart := strings.Index(originalComment, strings.TrimSpace(comment.Content))
			if contentStart >= 0 {
				prefix = originalComment[:contentStart]
				suffix = originalComment[contentStart+len(strings.TrimSpace(comment.Content)):]
				convertedComment = prefix + convertedComment + suffix
			} else {
				convertedComment = originalComment[:len(originalComment)-len(comment.Content)] + convertedComment
			}
		}

		// Replace this comment in the code
		result = result[:comment.Start] + convertedComment + result[comment.End:]
	}

	return result
}

func TestMCPConvertFileWithTestText(t *testing.T) {
	// Read the original test_text.txt file
	originalContent, err := os.ReadFile("../test_text.txt")
	if err != nil {
		t.Fatalf("Failed to read test_text.txt: %v", err)
	}

	// Create a temporary file with the content
	tempFile, cleanup := createTempFile(t, string(originalContent), ".txt")
	defer cleanup()

	// Get absolute path
	absPath, err := filepath.Abs(tempFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Simulate the MCP convert_file tool
	result := simulateConvertFileTool(t, absPath)
	if result.Error != nil {
		t.Fatalf("MCP tool simulation failed: %v", result.Error)
	}

	// Check that the result indicates success
	if !result.Success {
		t.Fatalf("MCP tool returned error: %s", result.Message)
	}

	expectedMessage := fmt.Sprintf("File %s completed processing to international/British English, the file has been updated.", absPath)
	if result.Message != expectedMessage {
		t.Errorf("Expected success message, got: %s", result.Message)
	}

	// Read the converted content
	convertedContent := readFileContent(t, tempFile)

	// Verify specific conversions happened
	expectedConversions := map[string]string{
		"organize":        "organise",
		"favorite":        "favourite",
		"colors":          "colours",
		"analyze":         "analyse",
		"behavior":        "behaviour",
		"center":          "centre",
		"theater":         "theatre",
		"modernized":      "modernised",
		"prioritize":      "prioritise",
		"optimization":    "optimisation",
		"recognize":       "recognise",
		"standardization": "standardisation",
		"optimizes":       "optimises", // In comment
		"aluminum":        "aluminium",
		"realize":         "realise",
		"pediatric":       "paediatric",
		"flavor":          "flavour",
		"humor":           "humour",
		"neighbor":        "neighbour",
	}

	for american, british := range expectedConversions {
		if !strings.Contains(convertedContent, british) {
			t.Errorf("Expected '%s' to be converted to '%s' in converted content", american, british)
		}
	}

	// Verify that code within markdown blocks is preserved
	if !strings.Contains(convertedContent, `const mom = "mother" // But change this mum`) {
		t.Error("Expected comment in code block to be converted while preserving code")
	}

	// Verify that regular code outside blocks is converted
	if !strings.Contains(convertedContent, `const mum = "mother"`) {
		t.Error("Expected regular code outside blocks to be converted")
	}
}

func TestMCPConvertFileWithCodeFile(t *testing.T) {
	// Create a Go code file with American spellings in comments and code
	codeContent := `package main

import "fmt"

// This function optimizes the color algorithm
func TestFunction() {
	// We need to analyze the behavior here
	mom := "mother"
	color := "red"
	organization := "company"
	fmt.Println("Hello, world!")

	/* This is a block comment that should be converted
	   We want to organize our favorite colors */
}
`

	// Create a temporary Go file
	tempFile, cleanup := createTempFile(t, codeContent, ".go")
	defer cleanup()

	// Get absolute path
	absPath, err := filepath.Abs(tempFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Simulate the MCP convert_file tool
	result := simulateConvertFileTool(t, absPath)
	if result.Error != nil {
		t.Fatalf("MCP tool simulation failed: %v", result.Error)
	}

	// Check that the result indicates success
	if !result.Success {
		t.Fatalf("MCP tool returned error: %s", result.Message)
	}

	// Read the converted content
	convertedContent := readFileContent(t, tempFile)

	// Verify that comments were converted
	if !strings.Contains(convertedContent, "// This function optimises the colour algorithm") {
		t.Error("Expected comment to be converted: 'optimizes' -> 'optimises', 'color' -> 'colour'")
	}

	if !strings.Contains(convertedContent, "// We need to analyse the behaviour here") {
		t.Error("Expected comment to be converted: 'analyze' -> 'analyse', 'behavior' -> 'behaviour'")
	}

	if !strings.Contains(convertedContent, "We want to organise our favourite colours") {
		t.Error("Expected block comment to be converted")
	}

	// Verify that code variables were NOT converted
	if !strings.Contains(convertedContent, `mom := "mother"`) {
		t.Error("Expected variable 'mom' to be preserved in code")
	}

	if !strings.Contains(convertedContent, `color := "red"`) {
		t.Error("Expected variable 'color' to be preserved in code")
	}

	if !strings.Contains(convertedContent, `organization := "company"`) {
		t.Error("Expected variable 'organization' to be preserved in code")
	}
}

func TestMCPConvertFileNonExistentFile(t *testing.T) {
	// Test with a non-existent file
	nonExistentFile := "/tmp/non_existent_file_12345.txt"

	result := simulateConvertFileTool(t, nonExistentFile)

	// Should return a failure
	if result.Success {
		t.Error("Expected failure for non-existent file")
	}

	expectedError := fmt.Sprintf("File does not exist: %s", nonExistentFile)
	if result.Message != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, result.Message)
	}
}

func TestMCPConvertFileNoChangesNeeded(t *testing.T) {
	// Create a file that's already in British English
	britishContent := `This is a test file with British spellings.

I need to organise my favourite colours and analyse the behaviour of the programme.
The centre of the theatre has been modernised with new technology.
We should prioritise optimisation and recognise the importance of standardisation.
`

	// Create a temporary file
	tempFile, cleanup := createTempFile(t, britishContent, ".txt")
	defer cleanup()

	// Get absolute path
	absPath, err := filepath.Abs(tempFile)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Simulate the MCP convert_file tool
	result := simulateConvertFileTool(t, absPath)
	if result.Error != nil {
		t.Fatalf("MCP tool simulation failed: %v", result.Error)
	}

	// Should indicate no changes were needed
	if !result.Success {
		t.Fatalf("MCP tool returned unexpected error: %s", result.Message)
	}

	expectedMessage := fmt.Sprintf("File %s processed but no changes were needed - already in British English", absPath)
	if result.Message != expectedMessage {
		t.Errorf("Expected 'no changes needed' message, got: %s", result.Message)
	}

	// Verify content is unchanged
	finalContent := readFileContent(t, tempFile)
	if finalContent != britishContent {
		t.Error("File content should be unchanged when no conversions are needed")
	}
}
