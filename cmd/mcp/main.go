package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"murican-to-english/pkg/converter"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"'Murican to English Converter",
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
