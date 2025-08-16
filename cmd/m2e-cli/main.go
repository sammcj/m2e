package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/sammcj/m2e/pkg/converter"
	"github.com/sammcj/m2e/pkg/fileutil"
	"github.com/sammcj/m2e/pkg/report"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// ANSI colour codes for diff output
const (
	ColourReset  = "\033[0m"
	ColourRed    = "\033[31m"
	ColourGreen  = "\033[32m"
	ColourYellow = "\033[33m"
	ColourCyan   = "\033[36m"
	ColourBold   = "\033[1m"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `m2e - Convert American English to British English

Usage:
  m2e [options] [text]                      # Convert text to stdout
  m2e [options] [file]                      # Convert file to stdout
  m2e [options] -o [output] [file]          # Convert file to output file
  m2e [options] [directory]                 # Convert all text files in directory (in-place)
  echo "text" | m2e [options]               # Convert stdin to stdout

Conversion Options:
  -o, -output string
        Output file to write to. If not specified, writes to stdout.
        (Not supported when processing directories or with output mode flags)
  -units
        Freedom Unit Conversion (default: false)
  -no-smart-quotes
        Disable smart quote normalisation (default: false)

Output Mode (mutually exclusive):
  -diff
        Show only git-style unified diff of changes (patch compatible)
  -diff-inline
        Show only character-level inline diff with colours
  -raw
        Show only the processed plain text
  -stats
        Show only conversion statistics
  (default: show diff + processed output + stats)

Additional Options:
  -width int
        Set output width for formatting (default: 80)
  -exit-on-change
        Exit with code 1 if changes are detected

Legacy Options (for backwards compatibility):
  -input string
        Input file or directory (use positional argument instead)

  -h, -help
        Show this help message

Examples:
  m2e document.txt                          # Show diff + processed text + stats
  m2e -diff document.txt                    # Show only unified diff (patch compatible)
  m2e -diff-inline document.txt             # Show only character-level diff with colours
  m2e -raw document.txt                     # Show only processed text
  m2e -stats document.txt                   # Show only conversion statistics
  m2e -o converted.txt document.txt         # Convert file to output file
  m2e -units document.txt                   # Convert with unit conversion
  m2e /path/to/project                      # Process all text files in directory
  echo "American text" | m2e -units        # Convert stdin with units

CI/CD Examples:
  m2e -exit-on-change /docs/               # Exit with code 1 if changes needed
  m2e -diff -exit-on-change README.md      # Show diff and exit 1 if changes
`)
}

func main() {
	// Modern flags
	var outputFile, outputFileLong string
	flag.StringVar(&outputFile, "o", "", "Output file to write to. If not specified, writes to stdout.")
	flag.StringVar(&outputFileLong, "output", "", "Output file to write to (same as -o)")
	convertUnits := flag.Bool("units", false, "Freedom Unit Conversion")
	noSmartQuotes := flag.Bool("no-smart-quotes", false, "Disable smart quote normalisation")

	// Legacy flags for backwards compatibility
	inputFile := flag.String("input", "", "Input file to convert (legacy, use positional argument instead)")

	// Output mode flags (mutually exclusive)
	showDiff := flag.Bool("diff", false, "Show only git-style unified diff of changes (patch compatible)")
	showDiffInline := flag.Bool("diff-inline", false, "Show only character-level inline diff with colours")
	showRaw := flag.Bool("raw", false, "Show only the processed plain text")
	showStats := flag.Bool("stats", false, "Show only conversion statistics")

	// Additional flags
	width := flag.Int("width", 80, "Set output width for formatting")
	exitOnChange := flag.Bool("exit-on-change", false, "Exit with code 1 if changes are detected")

	help := flag.Bool("help", false, "Show help message")
	helpShort := flag.Bool("h", false, "Show help message")
	flag.Parse()

	if *help || *helpShort {
		printUsage()
		os.Exit(0)
	}

	if os.Getenv("M2E_CLIPBOARD") == "1" || os.Getenv("M2E_CLIPBOARD") == "true" {
		if runtime.GOOS == "darwin" {
			// Determine smart quotes setting (default is true, disable if flag is set)
			normaliseSmartQuotes := !*noSmartQuotes
			handleClipboard(*convertUnits, normaliseSmartQuotes)
			return
		}
		fmt.Fprintf(os.Stderr, "Clipboard functionality is only supported on macOS.\n")
		os.Exit(1)
	}

	// Initialize converter
	conv, err := converter.NewConverter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing converter: %v\n", err)
		os.Exit(1)
	}

	// Set unit processing based on flag
	conv.SetUnitProcessingEnabled(*convertUnits)

	// Determine smart quotes setting (default is true, disable if flag is set)
	normaliseSmartQuotes := !*noSmartQuotes

	// Handle output file flag precedence (-o takes precedence over -output)
	finalOutputFile := ""
	if outputFile != "" {
		finalOutputFile = outputFile
	} else if outputFileLong != "" {
		finalOutputFile = outputFileLong
	}

	// Determine input source with improved logic
	var inputPath string
	var isDirectText bool
	var inputText string

	// Check if there are non-flag arguments (direct text input or file/directory path)
	if flag.NArg() > 0 {
		// Could be direct text input or a file/directory path
		potentialPath := flag.Args()[0]

		// Check if it's a file or directory path
		if _, err := os.Stat(potentialPath); err == nil {
			inputPath = potentialPath
		} else {
			// Treat as direct text input (join all arguments to handle quoted strings)
			inputText = strings.Join(flag.Args(), " ")
			isDirectText = true
		}
	} else if *inputFile != "" {
		// Legacy support for -input flag
		inputPath = *inputFile
	} else {
		// Check if stdin has data available
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			// No piped input and no arguments - show usage
			printUsage()
			os.Exit(1)
		}

		// Read from stdin
		inputBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
		inputText = string(inputBytes)
		isDirectText = true
	}

	// Determine output mode
	outputModeCount := 0
	if *showDiff {
		outputModeCount++
	}
	if *showDiffInline {
		outputModeCount++
	}
	if *showRaw {
		outputModeCount++
	}
	if *showStats {
		outputModeCount++
	}

	if outputModeCount > 1 {
		fmt.Fprintf(os.Stderr, "Error: Only one output mode flag can be specified at a time\n")
		os.Exit(1)
	}

	// Check for incompatible combinations
	if finalOutputFile != "" && outputModeCount > 0 {
		fmt.Fprintf(os.Stderr, "Error: Output file (-o) cannot be used with output mode flags\n")
		os.Exit(1)
	}

	// Handle different input types
	if isDirectText {
		// Handle direct text input (single string or stdin)
		err = handleSingleText(inputText, conv, normaliseSmartQuotes, finalOutputFile,
			*showDiff, *showDiffInline, *showRaw, *showStats, *exitOnChange, *width)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing text: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Handle file or directory input
		err = handleFileOrDirectory(inputPath, conv, normaliseSmartQuotes, finalOutputFile,
			*showDiff, *showDiffInline, *showRaw, *showStats, *exitOnChange, *width)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
			if *exitOnChange {
				os.Exit(1)
			} else {
				os.Exit(2)
			}
		}
	}
}

// handleSingleText processes a single text input (direct text or stdin)
func handleSingleText(inputText string, conv *converter.Converter, normaliseSmartQuotes bool,
	outputFile string, showDiff, showDiffInline, showRaw, showStats, exitOnChange bool, width int) error {

	convertedText := conv.ConvertToBritish(inputText, normaliseSmartQuotes)

	// Check if any changes were made
	hasChanges := inputText != convertedText

	// Exit early if exitOnChange is set and changes were detected
	if exitOnChange && hasChanges {
		defer os.Exit(1)
	}

	// If output file is specified, write converted text and exit
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(convertedText), 0644)
		if err != nil {
			return fmt.Errorf("failed to write to output file %s: %w", outputFile, err)
		}
		return nil
	}

	// Create analyser for statistics
	analyser := report.NewAnalyser(conv.GetAmericanToBritishDictionary())
	stats := analyser.AnalyseChanges(inputText, convertedText)

	// Handle specific output modes
	if showDiff {
		return showDiffOutput(inputText, convertedText, "stdin", false)
	}

	if showDiffInline {
		return showDiffOutput(inputText, convertedText, "stdin", true)
	}

	if showRaw {
		fmt.Print(convertedText)
		return nil
	}

	if showStats {
		return showStatsOutput(stats)
	}

	// Default mode: show diff + processed output + stats
	if hasChanges {
		// Show diff
		err := showDiffOutput(inputText, convertedText, "stdin", false)
		if err != nil {
			return err
		}
		fmt.Println() // Add separator
	}

	// Show processed output
	fmt.Print(convertedText)
	if !strings.HasSuffix(convertedText, "\n") {
		fmt.Println() // Ensure newline
	}
	fmt.Println() // Add separator

	// Show stats
	return showStatsOutput(stats)
}

// showDiffOutput displays diff of changes
func showDiffOutput(original, converted, filename string, inline bool) error {
	if original == converted {
		return nil // No changes to show
	}

	// Use unified diff format
	diff := createUnifiedDiff(original, converted, filename, inline)
	fmt.Print(diff)
	return nil
}

// showStatsOutput displays conversion statistics
func showStatsOutput(stats report.ChangeStats) error {
	fmt.Println("----- Conversion Statistics -----")
	fmt.Printf("📊 **Words processed:** %d\n", stats.TotalWords)
	fmt.Printf("🔤 **Spelling changes:** %d\n", stats.SpellingChanges)
	if stats.UnitConversions > 0 {
		fmt.Printf("📏 **Unit conversions:** %d\n", stats.UnitConversions)
	}
	if stats.QuoteChanges > 0 {
		fmt.Printf("📝 **Quote changes:** %d\n", stats.QuoteChanges)
	}
	return nil
}

// createUnifiedDiff creates a proper unified diff using the diffmatchpatch library
func createUnifiedDiff(original, converted, filename string, inline bool) string {
	dmp := diffmatchpatch.New()

	// Create a proper unified diff
	diffs := dmp.DiffMain(original, converted, false)

	if inline {
		// Character-level inline diff with colours
		return dmp.DiffPrettyText(diffs)
	} else {
		// Line-based unified diff format (patch compatible)
		return createLineBasedUnifiedDiff(original, converted, filename)
	}
}

// createLineBasedUnifiedDiff creates a traditional line-based unified diff
func createLineBasedUnifiedDiff(original, converted, filename string) string {
	originalLines := strings.Split(original, "\n")
	convertedLines := strings.Split(converted, "\n")

	// Remove trailing empty line if it exists (to handle files with/without trailing newlines consistently)
	if len(originalLines) > 0 && originalLines[len(originalLines)-1] == "" {
		originalLines = originalLines[:len(originalLines)-1]
	}
	if len(convertedLines) > 0 && convertedLines[len(convertedLines)-1] == "" {
		convertedLines = convertedLines[:len(convertedLines)-1]
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("--- %s\n", filename+".orig"))
	result.WriteString(fmt.Sprintf("+++ %s\n", filename))

	// Simple line-by-line comparison to find changed lines
	maxLines := len(originalLines)
	if len(convertedLines) > maxLines {
		maxLines = len(convertedLines)
	}

	// Find contiguous blocks of changes
	i := 0
	for i < maxLines {
		// Skip matching lines
		for i < maxLines {
			var origLine, convLine string
			if i < len(originalLines) {
				origLine = originalLines[i]
			}
			if i < len(convertedLines) {
				convLine = convertedLines[i]
			}

			if origLine != convLine {
				break // Found a difference
			}
			i++
		}

		if i >= maxLines {
			break // No more changes
		}

		// Find the end of this block of changes
		start := i
		for i < maxLines {
			var origLine, convLine string
			if i < len(originalLines) {
				origLine = originalLines[i]
			}
			if i < len(convertedLines) {
				convLine = convertedLines[i]
			}

			if origLine == convLine {
				break // Found matching line, end of this change block
			}
			i++
		}

		// Generate hunk for this block of changes
		oldStart := start + 1
		newStart := start + 1
		oldCount := 0
		newCount := 0

		// Count lines in this hunk
		for j := start; j < i; j++ {
			if j < len(originalLines) {
				oldCount++
			}
			if j < len(convertedLines) {
				newCount++
			}
		}

		// Add hunk header
		result.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", oldStart, oldCount, newStart, newCount))

		// Add the actual line changes
		for j := start; j < i; j++ {
			if j < len(originalLines) && j < len(convertedLines) {
				// Both lines exist, show as replacement
				if originalLines[j] != convertedLines[j] {
					result.WriteString(fmt.Sprintf("-%s\n", originalLines[j]))
					result.WriteString(fmt.Sprintf("+%s\n", convertedLines[j]))
				}
			} else if j < len(originalLines) {
				// Line removed
				result.WriteString(fmt.Sprintf("-%s\n", originalLines[j]))
			} else if j < len(convertedLines) {
				// Line added
				result.WriteString(fmt.Sprintf("+%s\n", convertedLines[j]))
			}
		}
	}

	return result.String()
}

// shouldUseColour determines if we should output ANSI colour codes
func shouldUseColour() bool {
	// Check if NO_COLOR environment variable is set (standard)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if FORCE_COLOR is set
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check if stdout is a terminal
	// This is a simple check - in production code you might want to use a library like isatty
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	// Check if we're in a CI environment (usually no colour support)
	if os.Getenv("CI") != "" {
		return false
	}

	return true
}

// handleFileOrDirectory processes file or directory input
func handleFileOrDirectory(inputPath string, conv *converter.Converter, normaliseSmartQuotes bool,
	outputFile string, showDiff, showDiffInline, showRaw, showStats, exitOnChange bool, width int) error {

	// Check if input is a directory or file
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat input path: %w", err)
	}

	if info.IsDir() {
		// Directory processing
		return handleDirectory(inputPath, conv, normaliseSmartQuotes, outputFile,
			showDiff, showDiffInline, showRaw, showStats, exitOnChange, width)
	} else {
		// Single file processing
		return handleSingleFile(inputPath, conv, normaliseSmartQuotes, outputFile,
			showDiff, showDiffInline, showRaw, showStats, exitOnChange, width)
	}
}

// handleSingleFile processes a single file
func handleSingleFile(filePath string, conv *converter.Converter, normaliseSmartQuotes bool,
	outputFile string, showDiff, showDiffInline, showRaw, showStats, exitOnChange bool, width int) error {

	// Read file content
	content, err := fileutil.ReadFileContent(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Convert content
	convertedContent := conv.ConvertToBritish(content, normaliseSmartQuotes)

	// Check if any changes were made
	hasChanges := content != convertedContent

	// Exit early if exitOnChange is set and changes were detected
	if exitOnChange && hasChanges {
		defer os.Exit(1)
	}

	// If output file is specified, write converted text and exit
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(convertedContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to write to output file %s: %w", outputFile, err)
		}
		return nil
	}

	// Create analyser for statistics
	analyser := report.NewAnalyser(conv.GetAmericanToBritishDictionary())
	stats := analyser.AnalyseChanges(content, convertedContent)

	// Handle specific output modes
	if showDiff {
		return showDiffOutput(content, convertedContent, filePath, false)
	}

	if showDiffInline {
		return showDiffOutput(content, convertedContent, filePath, true)
	}

	if showRaw {
		fmt.Print(convertedContent)
		return nil
	}

	if showStats {
		return showStatsOutput(stats)
	}

	// Default mode: show diff + processed output + stats
	if hasChanges {
		// Show diff
		err := showDiffOutput(content, convertedContent, filePath, false)
		if err != nil {
			return err
		}
		fmt.Println() // Add separator
	}

	// Show processed output
	fmt.Print(convertedContent)
	if !strings.HasSuffix(convertedContent, "\n") {
		fmt.Println() // Ensure newline
	}
	fmt.Println() // Add separator

	// Show stats
	return showStatsOutput(stats)
}

// handleDirectory processes all text files in a directory recursively
func handleDirectory(dirPath string, conv *converter.Converter, normaliseSmartQuotes bool,
	outputFile string, showDiff, showDiffInline, showRaw, showStats, exitOnChange bool, width int) error {

	if outputFile != "" {
		return fmt.Errorf("output file not supported when processing directories")
	}

	// Find all text files in directory
	files, err := fileutil.FindTextFiles(dirPath)
	if err != nil {
		return fmt.Errorf("failed to find text files in directory %s: %w", dirPath, err)
	}

	if len(files) == 0 {
		fmt.Printf("No text files found in directory: %s\n", dirPath)
		return nil
	}

	fmt.Printf("Found %d text file(s) in directory: %s\n", len(files), dirPath)

	// Track overall changes for exitOnChange
	anyChanges := false

	// For output modes, collect all results
	var allResults []string
	var totalStats report.ChangeStats
	analyser := report.NewAnalyser(conv.GetAmericanToBritishDictionary())

	for _, file := range files {
		fmt.Printf("Processing: %s\n", file.RelativePath)

		// Read file content
		content, err := fileutil.ReadFileContent(file.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read file %s: %v\n", file.Path, err)
			continue
		}

		// Convert content
		convertedContent := conv.ConvertToBritish(content, normaliseSmartQuotes)
		hasChanges := content != convertedContent

		if hasChanges {
			anyChanges = true
		}

		// Generate statistics for this file
		stats := analyser.AnalyseChanges(content, convertedContent)

		// Accumulate total stats
		totalStats.TotalWords += stats.TotalWords
		totalStats.SpellingChanges += stats.SpellingChanges
		totalStats.UnitConversions += stats.UnitConversions
		totalStats.QuoteChanges += stats.QuoteChanges

		// Handle specific output modes
		if showDiff && hasChanges {
			diff := createUnifiedDiff(content, convertedContent, file.RelativePath, false)
			allResults = append(allResults, fmt.Sprintf("=== %s ===\n%s", file.RelativePath, diff))
		} else if showDiffInline && hasChanges {
			diff := createUnifiedDiff(content, convertedContent, file.RelativePath, true)
			allResults = append(allResults, fmt.Sprintf("=== %s ===\n%s", file.RelativePath, diff))
		} else if showRaw && hasChanges {
			allResults = append(allResults, fmt.Sprintf("=== %s ===\n%s", file.RelativePath, convertedContent))
		} else if !showStats {
			// Default mode or no specific output mode: process files in-place and show results
			if hasChanges {
				err = os.WriteFile(file.Path, []byte(convertedContent), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to write changes to file %s: %v\n", file.Path, err)
				} else {
					fmt.Printf("Updated: %s\n", file.RelativePath)
				}
			}
		}
	}

	// Handle output modes
	if showDiff || showDiffInline || showRaw {
		for _, result := range allResults {
			fmt.Print(result)
			fmt.Println()
		}
	} else if showStats {
		err := showStatsOutput(totalStats)
		if err != nil {
			return err
		}
	}

	// Handle exitOnChange
	if exitOnChange && anyChanges {
		os.Exit(1)
	}

	return nil
}

func handleClipboard(convertUnits bool, normaliseSmartQuotes bool) {
	// Get text from clipboard
	pasteCmd := exec.Command("pbpaste")
	var pasteOut bytes.Buffer
	pasteCmd.Stdout = &pasteOut
	err := pasteCmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from clipboard: %v\n", err)
		os.Exit(1)
	}

	clipboardText := pasteOut.String()

	// Convert the text
	conv, err := converter.NewConverter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing converter: %v\n", err)
		os.Exit(1)
	}

	// Set unit processing based on flag
	conv.SetUnitProcessingEnabled(convertUnits)

	convertedText := conv.ConvertToBritish(clipboardText, normaliseSmartQuotes)

	// Copy text to clipboard
	copyCmd := exec.Command("pbcopy")
	copyCmd.Stdin = strings.NewReader(convertedText)
	err = copyCmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to clipboard: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Clipboard content converted and updated.")
}
