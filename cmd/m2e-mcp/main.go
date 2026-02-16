package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/sammcj/m2e/pkg/converter"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// sensitivePathPrefixes lists path prefixes that should be rejected for file conversion.
var sensitivePathPrefixes = []string{
	"/etc/",
	"/var/",
	"/usr/",
	"/sys/",
	"/proc/",
	"/dev/",
}

// sensitiveFilenames lists filenames that should never be overwritten.
var sensitiveFilenames = []string{
	".bashrc", ".bash_profile", ".zshrc", ".zprofile", ".profile",
	".ssh", ".gnupg", ".env", ".netrc", ".npmrc",
	"authorized_keys", "known_hosts", "id_rsa", "id_ed25519",
	"shadow", "passwd", "sudoers",
}

// validateFilePath checks that a file path is safe to read/write.
func validateFilePath(filePath string) error {
	cleaned := filepath.Clean(filePath)
	absPath, err := filepath.Abs(cleaned)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Reject paths containing .. after cleaning
	if strings.Contains(absPath, "..") {
		return fmt.Errorf("path traversal not allowed: %s", filePath)
	}

	// Reject sensitive system paths
	for _, prefix := range sensitivePathPrefixes {
		if strings.HasPrefix(absPath, prefix) {
			return fmt.Errorf("access to system path not allowed: %s", absPath)
		}
	}

	// Reject sensitive filenames
	base := filepath.Base(absPath)
	if slices.Contains(sensitiveFilenames, base) {
		return fmt.Errorf("access to sensitive file not allowed: %s", base)
	}

	return nil
}

// isPlainTextFile checks if a file extension indicates it's a plain text file
// that can be safely converted entirely (not just comments)
func isPlainTextFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	plainTextExtensions := []string{
		".txt", ".md", ".markdown", ".rst", ".text", ".doc", ".rtf",
		".tex", ".latex", ".org", ".adoc", ".asciidoc",
	}
	return slices.Contains(plainTextExtensions, ext)
}

// convertFileContentWithOptions converts file content based on file type with custom options
func convertFileContentWithOptions(conv *converter.Converter, content, filePath string, normaliseSmartQuotes bool) string {
	if isPlainTextFile(filePath) {
		// For plain text files, use code-aware processing which:
		// - Converts all regular text
		// - Only converts comments within code blocks (preserving code)
		return conv.ProcessCodeAware(content, normaliseSmartQuotes)
	} else {
		// For code/config files, only convert comments to preserve functionality
		return convertOnlyCommentsWithOptions(conv, content, normaliseSmartQuotes)
	}
}

// convertOnlyCommentsWithOptions converts only comments in code with custom options
func convertOnlyCommentsWithOptions(conv *converter.Converter, code string, normaliseSmartQuotes bool) string {
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
		convertedComment := conv.ConvertToBritish(comment.Content, normaliseSmartQuotes)

		// Preserve the comment structure (e.g., //, /* */, #, etc.)
		// by replacing just the content part
		if len(originalComment) > len(comment.Content) {
			// This handles cases where the comment has prefix/suffix (like /* */)
			prefix := ""
			suffix := ""

			// Find where the actual content starts and ends
			contentStart := strings.Index(originalComment, strings.TrimSpace(comment.Content))
			if contentStart >= 0 {
				prefix = originalComment[:contentStart]
				suffix = originalComment[contentStart+len(strings.TrimSpace(comment.Content)):]
				convertedComment = prefix + convertedComment + suffix
			} else {
				// Fallback: just use the converted comment
				convertedComment = originalComment[:len(originalComment)-len(comment.Content)] + convertedComment
			}
		}

		// Replace this comment in the code
		result = result[:comment.Start] + convertedComment + result[comment.End:]
	}

	return result
}

func main() {
	s := server.NewMCPServer(
		"M2E - 'Murican to English Converter",
		"1.0.0",
	)

	conv, err := converter.NewConverter()
	if err != nil {
		log.Fatalf("Failed to create converter: %v", err)
	}
	var convMu sync.Mutex // protects mutable converter state during concurrent requests

	convertTool := mcp.NewTool("convert_text",
		mcp.WithDescription("Convert American English text to British English with optional unit conversion"),
		mcp.WithString("text", mcp.Required(), mcp.Description("The text to convert")),
		mcp.WithString("convert_units", mcp.Description("Freedom Unit Conversion (true/false, default: false)")),
		mcp.WithString("normalise_smart_quotes", mcp.Description("Normalise smart quotes to regular quotes (true/false, default: true)")),
	)
	s.AddTool(convertTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		text, err := req.RequireString("text")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get optional parameters with defaults
		convertUnits := false
		if val, err := req.RequireString("convert_units"); err == nil {
			convertUnits = strings.ToLower(val) == "true"
		}

		normaliseSmartQuotes := true
		if val, err := req.RequireString("normalise_smart_quotes"); err == nil {
			normaliseSmartQuotes = strings.ToLower(val) != "false"
		}

		// Lock around mutable state mutation + conversion for concurrent safety
		convMu.Lock()
		conv.SetUnitProcessingEnabled(convertUnits)
		convertedText := conv.ConvertToBritish(text, normaliseSmartQuotes)
		convMu.Unlock()

		return mcp.NewToolResultText(convertedText), nil
	})

	convertFileTool := mcp.NewTool("convert_file",
		mcp.WithDescription("Convert a file from American English to International / British English and save it back. Uses intelligent processing: for plain text files (.txt, .md, etc.), converts all text but preserves code within markdown blocks. For code/config files (.go, .js, .py, etc.), only converts comments to preserve functionality. Supports optional unit conversion from imperial to metric."),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("The fully qualified path to the file to convert")),
		mcp.WithString("convert_units", mcp.Description("Freedom Unit Conversion (true/false, default: false)")),
		mcp.WithString("normalise_smart_quotes", mcp.Description("Normalise smart quotes to regular quotes (true/false, default: true)")),
	)
	s.AddTool(convertFileTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := req.RequireString("file_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Validate the file path for security
		if err := validateFilePath(filePath); err != nil {
			log.Printf("Rejected file path %q: %v", filePath, err)
			return mcp.NewToolResultError(fmt.Sprintf("File path rejected: %v", err)), nil
		}

		// Get optional parameters with defaults
		convertUnits := false
		if val, err := req.RequireString("convert_units"); err == nil {
			convertUnits = strings.ToLower(val) == "true"
		}

		normaliseSmartQuotes := true
		if val, err := req.RequireString("normalise_smart_quotes"); err == nil {
			normaliseSmartQuotes = strings.ToLower(val) != "false"
		}

		// Check if file exists and get its permissions
		fileInfo, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("File does not exist: %s", filePath)), nil
		}
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error accessing file %s: %v", filePath, err)), nil
		}
		originalMode := fileInfo.Mode()

		// Read the original file content
		originalContent, err := os.ReadFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error reading file %s: %v", filePath, err)), nil
		}

		// Lock around mutable state mutation + conversion for concurrent safety
		convMu.Lock()
		conv.SetUnitProcessingEnabled(convertUnits)
		convertedContent := convertFileContentWithOptions(conv, string(originalContent), filePath, normaliseSmartQuotes)
		convMu.Unlock()

		// Check if there were any changes
		if string(originalContent) == convertedContent {
			return mcp.NewToolResultText(fmt.Sprintf("File %s processed but no changes were needed - already in British English", filePath)), nil
		}

		// Write the converted content back to the file, preserving original permissions
		err = os.WriteFile(filePath, []byte(convertedContent), originalMode.Perm())
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error writing to file %s: %v", filePath, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("File %s completed processing to international / British English, the file has been updated.", filePath)), nil
	})

	dictionaryResource := mcp.NewResource("dictionary://american-to-british", "American to British Dictionary")
	s.AddResource(dictionaryResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		dict := conv.GetAmericanToBritishDictionary()
		var b strings.Builder
		b.Grow(len(dict) * 30)
		for k, v := range dict {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "dictionary://american-to-british",
				MIMEType: "text/plain",
				Text:     b.String(),
			},
		}, nil
	})

	transport := os.Getenv("MCP_TRANSPORT")
	if transport == "stdio" {
		// In stdio mode, we should not log to stdout/stderr.
		// This will be implemented properly with file logging later.
		log.SetOutput(io.Discard)
		if err := server.ServeStdio(s); err != nil {
			// Since we can't log, we can't do much here.
			os.Exit(1)
		}
	} else {
		port := os.Getenv("MCP_PORT")
		if port == "" {
			port = "8081"
		}
		log.Printf("MCP server starting on port %s\n", port)
		httpServer := server.NewStreamableHTTPServer(s)
		if err := httpServer.Start(":" + port); err != nil {
			log.Fatalf("MCP server failed to start: %v", err)
		}
	}
}
