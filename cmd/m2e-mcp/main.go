package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sammcj/m2e/pkg/converter"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// isPlainTextFile checks if a file extension indicates it's a plain text file
// that can be safely converted entirely (not just comments)
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

// convertFileContent converts file content based on file type
func convertFileContent(conv *converter.Converter, content, filePath string) string {
	if isPlainTextFile(filePath) {
		// For plain text files, use code-aware processing which:
		// - Converts all regular text
		// - Only converts comments within code blocks (preserving code)
		return conv.ProcessCodeAware(content, true)
	} else {
		// For code/config files, only convert comments to preserve functionality
		return convertOnlyComments(conv, content)
	}
}

// convertOnlyComments converts only comments in code, preserving all other content
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

	convertTool := mcp.NewTool("convert_text",
		mcp.WithDescription("Convert American English text to British English"),
		mcp.WithString("text", mcp.Required(), mcp.Description("The text to convert")),
	)
	s.AddTool(convertTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		text, err := req.RequireString("text")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		convertedText := conv.ConvertToBritish(text, true)
		return mcp.NewToolResultText(convertedText), nil
	})

	convertFileTool := mcp.NewTool("convert_file",
		mcp.WithDescription("Convert a file from American English to International / British English and save it back. Uses intelligent processing: for plain text files (.txt, .md, etc.), converts all text but preserves code within markdown blocks. For code/config files (.go, .js, .py, etc.), only converts comments to preserve functionality."),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("The fully qualified path to the file to convert")),
	)
	s.AddTool(convertFileTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := req.RequireString("file_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("File does not exist: %s", filePath)), nil
		}

		// Read the original file content
		originalContent, err := os.ReadFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error reading file %s: %v", filePath, err)), nil
		}

		// Convert the content based on file type
		convertedContent := convertFileContent(conv, string(originalContent), filePath)

		// Check if there were any changes
		if string(originalContent) == convertedContent {
			return mcp.NewToolResultText(fmt.Sprintf("File %s processed but no changes were needed - already in British English", filePath)), nil
		}

		// Write the converted content back to the file
		err = os.WriteFile(filePath, []byte(convertedContent), 0644)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error writing to file %s: %v", filePath, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("File %s completed processing to international / British English, the file has been updated.", filePath)), nil
	})

	dictionaryResource := mcp.NewResource("dictionary://american-to-british", "American to British Dictionary")
	s.AddResource(dictionaryResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		dict := conv.GetAmericanToBritishDictionary()
		var dictString string
		for k, v := range dict {
			dictString += fmt.Sprintf("%s: %s\n", k, v)
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "dictionary://american-to-british",
				MIMEType: "text/plain",
				Text:     dictString,
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
