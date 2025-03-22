package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"murican-to-english/pkg/converter"
)

// App struct
type App struct {
	ctx       context.Context
	converter *converter.Converter
	filePath  string // Store the path of the file being processed
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize the converter
	var err error
	a.converter, err = converter.NewConverter()
	if err != nil {
		fmt.Printf("Error initializing converter: %v\n", err)
	}

	// Check if the app was launched with a file path argument
	args := os.Args
	if len(args) > 1 {
		filePath := args[1]
		// Check if the file exists
		if _, err := os.Stat(filePath); err == nil {
			a.filePath = filePath
			// Process the file and exit
			err := a.ConvertFileToEnglish(filePath)
			if err != nil {
				fmt.Printf("Error converting file: %v\n", err)
			}
			// Exit the application after processing the file
			os.Exit(0)
		}
	}
}

// ConvertToBritish converts American English text to British English
func (a *App) ConvertToBritish(text string, normaliseSmartQuotes bool) string {
	if a.converter == nil {
		return "Error: Converter not initialized"
	}
	return a.converter.ConvertToBritish(text, normaliseSmartQuotes)
}

// ConvertToAmerican converts British English text to American English
func (a *App) ConvertToAmerican(text string, normaliseSmartQuotes bool) string {
	if a.converter == nil {
		return "Error: Converter not initialized"
	}
	return a.converter.ConvertToAmerican(text, normaliseSmartQuotes)
}

// ConvertFileToEnglish converts a file's content from American to British English and saves it back
func (a *App) ConvertFileToEnglish(filePath string) error {
	// Read the file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Convert the content
	convertedContent := a.ConvertToBritish(string(content), true)

	// Write the converted content back to the file
	err = ioutil.WriteFile(filePath, []byte(convertedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// HandleDroppedFile processes a file that was dropped onto the application
func (a *App) HandleDroppedFile(filePath string) (string, error) {
	// Store the file path
	a.filePath = filePath

	// Read the file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// Return the content for display in the UI
	return string(content), nil
}

// SaveConvertedFile saves the converted content back to the original file
func (a *App) SaveConvertedFile(content string) error {
	if a.filePath == "" {
		return fmt.Errorf("no file path set")
	}

	// Write the content back to the file
	err := ioutil.WriteFile(a.filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	// Clear the file path after saving
	a.filePath = ""

	return nil
}

// GetCurrentFilePath returns the path of the currently loaded file
func (a *App) GetCurrentFilePath() string {
	return a.filePath
}

// ClearCurrentFile clears the current file path
func (a *App) ClearCurrentFile() {
	a.filePath = ""
}
