package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
  -save, -s
        Overwrite the input file with converted content
  (default: show diff + processed output + stats)

Additional Options:
  -width int
        Set output width for formatting (default: 80)
  -exit-on-change
        Exit with code 1 if changes are detected
  -rename
        Rename files that have American spellings in their filename
  -size-max-kb int
        Maximum file size to process in KB (default: 10240 KB = 10 MB)

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
  m2e -save document.txt                    # Overwrite file with converted content
  m2e -s document.txt                       # Same as -save (shorthand)
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
	saveInPlace := flag.Bool("save", false, "Overwrite the input file with converted content (cannot be used with other output modes)")
	saveInPlaceShort := flag.Bool("s", false, "Shorthand for -save")

	// Additional flags
	width := flag.Int("width", 80, "Set output width for formatting")
	exitOnChange := flag.Bool("exit-on-change", false, "Exit with code 1 if changes are detected")
	renameFiles := flag.Bool("rename", false, "Rename files that have American spellings in their filename")
	maxFileSize := flag.Int("size-max-kb", 10240, "Maximum file size to process in KB (default: 10240)") // 10MB default

	help := flag.Bool("help", false, "Show help message")
	helpShort := flag.Bool("h", false, "Show help message")

	// Custom argument parsing to handle flags after positional arguments
	var nonFlagArgs []string
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			// Handle flags with values
			switch arg {
			case "-o", "-output":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					if arg == "-o" {
						outputFile = args[i+1]
					} else {
						outputFileLong = args[i+1]
					}
					i++ // Skip the value
				}
			case "-width":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					// Parse width manually
					i++ // Skip the value for now, flag.Parse() will handle it
				}
			case "-size-max-kb":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					// Parse size-max-kb manually
					i++ // Skip the value for now, flag.Parse() will handle it
				}
			case "-s":
				*saveInPlaceShort = true
			case "-units":
				*convertUnits = true
			case "-no-smart-quotes":
				*noSmartQuotes = true
			case "-save":
				*saveInPlace = true
			case "-diff":
				*showDiff = true
			case "-diff-inline":
				*showDiffInline = true
			case "-raw":
				*showRaw = true
			case "-stats":
				*showStats = true
			case "-exit-on-change":
				*exitOnChange = true
			case "-rename":
				*renameFiles = true
			case "-help", "--help":
				*help = true
			case "-h":
				*helpShort = true
			}
		} else {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}

	// Set flag.Args() to our non-flag arguments
	os.Args = append([]string{os.Args[0]}, nonFlagArgs...)
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
		// Handle multiple file arguments or single input
		if flag.NArg() == 1 {
			// Single argument - could be direct text input or a file/directory path
			potentialPath := flag.Args()[0]

			// Check if it's a file or directory path
			if _, err := os.Stat(potentialPath); err == nil {
				inputPath = potentialPath
			} else {
				// Treat as direct text input
				inputText = potentialPath
				isDirectText = true
			}
		} else {
			// Multiple arguments - check if they're all valid files
			allFilesValid := true
			for _, arg := range flag.Args() {
				if _, err := os.Stat(arg); err != nil {
					allFilesValid = false
					break
				}
			}

			if allFilesValid {
				// All arguments are valid files - process them as multiple files
				err = handleMultipleFiles(flag.Args(), conv, normaliseSmartQuotes, finalOutputFile,
					*showDiff, *showDiffInline, *showRaw, *showStats, (*saveInPlace || *saveInPlaceShort), *exitOnChange, *width, *maxFileSize)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
					os.Exit(1)
				}
				return // Exit early after processing multiple files
			} else {
				// Not all arguments are valid files - treat as direct text input
				inputText = strings.Join(flag.Args(), " ")
				isDirectText = true
			}
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
	if *saveInPlace {
		outputModeCount++
	}

	if *saveInPlaceShort {
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

	// Check if save flag is used with text input (not allowed)
	if (*saveInPlace || *saveInPlaceShort) && isDirectText {
		fmt.Fprintf(os.Stderr, "Error: -save flag can only be used with file input, not text input or stdin\n")
		os.Exit(1)
	}

	// Handle different input types
	if isDirectText {
		// Handle direct text input (single string or stdin)
		err = handleSingleText(inputText, conv, normaliseSmartQuotes, finalOutputFile,
			*showDiff, *showDiffInline, *showRaw, *showStats, (*saveInPlace || *saveInPlaceShort), *exitOnChange, *width)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing text: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Handle file or directory input
		// Use max file size flag
		finalMaxFileSize := *maxFileSize
		err = handleFileOrDirectory(inputPath, conv, normaliseSmartQuotes, finalOutputFile,
			*showDiff, *showDiffInline, *showRaw, *showStats, (*saveInPlace || *saveInPlaceShort), *exitOnChange, *renameFiles, *width, finalMaxFileSize)
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
	outputFile string, showDiff, showDiffInline, showRaw, showStats, saveInPlace, exitOnChange bool, width int) error {

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
	return showStatsOutputWithMode(stats, false)
}

// showStatsOutputWithMode displays conversion statistics with context-aware wording
func showStatsOutputWithMode(stats report.ChangeStats, savedChanges bool) error {
	if savedChanges {
		fmt.Println("----- Changes Applied -----")
		fmt.Printf("📊 **Words processed:** %d\n", stats.TotalWords)
		fmt.Printf("🔤 **Spelling changes applied:** %d\n", stats.SpellingChanges)
		if stats.UnitConversions > 0 {
			fmt.Printf("📏 **Unit conversions applied:** %d\n", stats.UnitConversions)
		}
		if stats.QuoteChanges > 0 {
			fmt.Printf("📝 **Quote changes applied:** %d\n", stats.QuoteChanges)
		}
	} else {
		fmt.Println("----- Changes Detected -----")
		fmt.Printf("📊 **Words processed:** %d\n", stats.TotalWords)
		fmt.Printf("🔤 **Spelling changes needed:** %d\n", stats.SpellingChanges)
		if stats.UnitConversions > 0 {
			fmt.Printf("📏 **Unit conversions needed:** %d\n", stats.UnitConversions)
		}
		if stats.QuoteChanges > 0 {
			fmt.Printf("📝 **Quote changes needed:** %d\n", stats.QuoteChanges)
		}
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

// createLineBasedUnifiedDiff creates a simple line-based diff showing only lines with actual changes
func createLineBasedUnifiedDiff(original, converted, filename string) string {
	originalLines := strings.Split(original, "\n")
	convertedLines := strings.Split(converted, "\n")

	var result strings.Builder
	fmt.Fprintf(&result, "--- %s\n", filename+".orig")
	fmt.Fprintf(&result, "+++ %s\n", filename)

	hasAnyChanges := false

	// Compare original and converted lines directly
	lineCount := max(len(originalLines), len(convertedLines))

	for i := 0; i < lineCount; i++ {
		var origLine, convLine string
		if i < len(originalLines) {
			origLine = originalLines[i]
		}
		if i < len(convertedLines) {
			convLine = convertedLines[i]
		}

		if origLine != convLine {
			hasAnyChanges = true
			lineNum := i + 1
			fmt.Fprintf(&result, "@@ -%d,1 +%d,1 @@\n", lineNum, lineNum)
			fmt.Fprintf(&result, "-%s\n", origLine)
			fmt.Fprintf(&result, "+%s\n", convLine)
		}
	}

	// If no changes found, return empty string
	if !hasAnyChanges {
		return ""
	}

	return result.String()
}

// handleFileOrDirectory processes file or directory input
func handleFileOrDirectory(inputPath string, conv *converter.Converter, normaliseSmartQuotes bool,
	outputFile string, showDiff, showDiffInline, showRaw, showStats, saveInPlace, exitOnChange, renameFiles bool, width, maxFileSize int) error {

	// Check if input is a directory or file
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat input path: %w", err)
	}

	if info.IsDir() {
		// Directory processing
		return handleDirectory(inputPath, conv, normaliseSmartQuotes, outputFile,
			showDiff, showDiffInline, showRaw, showStats, saveInPlace, exitOnChange, renameFiles, width, maxFileSize)
	} else {
		// Single file processing
		return handleSingleFile(inputPath, conv, normaliseSmartQuotes, outputFile,
			showDiff, showDiffInline, showRaw, showStats, saveInPlace, exitOnChange, width, maxFileSize)
	}
}

// handleSingleFile processes a single file
func handleSingleFile(filePath string, conv *converter.Converter, normaliseSmartQuotes bool,
	outputFile string, showDiff, showDiffInline, showRaw, showStats, saveInPlace, exitOnChange bool, width, maxFileSize int) error {

	// Read file content
	content, err := fileutil.ReadFileContentWithMaxSize(filePath, maxFileSize)
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

	// If save flag is specified, overwrite the original file
	if saveInPlace {
		if hasChanges {
			err := os.WriteFile(filePath, []byte(convertedContent), 0644)
			if err != nil {
				return fmt.Errorf("failed to save changes to file %s: %w", filePath, err)
			}
			fmt.Printf("Saved changes to: %s\n", filePath)
		} else {
			fmt.Printf("No changes needed: %s\n", filePath)
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
	outputFile string, showDiff, showDiffInline, showRaw, showStats, saveInPlace, exitOnChange, renameFiles bool, width, maxFileSize int) error {

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
	var changedFiles []string
	var fileStats []report.ChangeStats
	var filenameChanges []string // Track files that need renaming
	analyser := report.NewAnalyser(conv.GetAmericanToBritishDictionary())

	for _, file := range files {
		fmt.Printf("Processing: %s\n", file.RelativePath)

		// Read file content
		content, err := fileutil.ReadFileContentWithMaxSize(file.Path, maxFileSize)
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

		// Handle filename renaming if requested
		var newFilePath string
		var filenameChanged bool
		if renameFiles {
			newFilePath, filenameChanged = convertFilename(file.Path, conv)
			if filenameChanged {
				anyChanges = true
			}
		}

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
		} else if saveInPlace {
			// Save mode: overwrite files with changes
			if hasChanges {
				err = os.WriteFile(file.Path, []byte(convertedContent), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to save changes to file %s: %v\n", file.Path, err)
				} else {
					fmt.Printf("Saved changes to: %s\n", file.RelativePath)
				}
			} else if !filenameChanged {
				fmt.Printf("No changes needed: %s\n", file.RelativePath)
			}

			// Handle file renaming if requested and filename needs changing
			if renameFiles && filenameChanged {
				err = os.Rename(file.Path, newFilePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to rename file %s to %s: %v\n", file.Path, newFilePath, err)
				} else {
					// Calculate relative path for display
					var newRelativePath string
					if filepath.Dir(newFilePath) == dirPath {
						newRelativePath = filepath.Base(newFilePath)
					} else {
						rel, err := filepath.Rel(dirPath, newFilePath)
						if err != nil {
							newRelativePath = filepath.Base(newFilePath)
						} else {
							newRelativePath = rel
						}
					}
					fmt.Printf("Renamed file: %s → %s\n", file.RelativePath, newRelativePath)
				}
			}
		} else if !showStats {
			// Default mode: collect file changes for summary report
			if hasChanges || filenameChanged {
				changedFiles = append(changedFiles, file.RelativePath)
				fileStats = append(fileStats, stats)
				if renameFiles && filenameChanged {
					// Calculate relative path for the new filename
					var newRelativePath string
					if filepath.Dir(newFilePath) == dirPath {
						newRelativePath = filepath.Base(newFilePath)
					} else {
						rel, err := filepath.Rel(dirPath, newFilePath)
						if err != nil {
							newRelativePath = filepath.Base(newFilePath)
						} else {
							newRelativePath = rel
						}
					}
					filenameChanges = append(filenameChanges, fmt.Sprintf("%s → %s", file.RelativePath, newRelativePath))
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
	} else if saveInPlace {
		// Save mode: show summary of applied changes
		if totalStats.SpellingChanges > 0 || totalStats.UnitConversions > 0 || totalStats.QuoteChanges > 0 {
			fmt.Println()
			err := showStatsOutputWithMode(totalStats, true)
			if err != nil {
				return err
			}
		}
	} else {
		// Default mode: show summary report
		fmt.Println()
		if len(changedFiles) > 0 {
			fmt.Printf("Files requiring changes (%d):\n", len(changedFiles))
			for i, filePath := range changedFiles {
				stats := fileStats[i]
				fmt.Printf("  %s: %d spelling change(s) needed", filePath, stats.SpellingChanges)
				if stats.UnitConversions > 0 {
					fmt.Printf(", %d unit conversion(s) needed", stats.UnitConversions)
				}
				if stats.QuoteChanges > 0 {
					fmt.Printf(", %d quote change(s) needed", stats.QuoteChanges)
				}
				fmt.Println()
			}

			// Show filename changes if any
			if len(filenameChanges) > 0 {
				fmt.Printf("\nFiles requiring filename changes (%d):\n", len(filenameChanges))
				for _, change := range filenameChanges {
					fmt.Printf("  %s\n", change)
				}
			}

			var flagSuggestion string
			if len(filenameChanges) > 0 {
				flagSuggestion = "\nTo apply these changes, use the -save -rename flags."
			} else {
				flagSuggestion = "\nTo apply these changes, use the -save flag."
			}
			fmt.Println(flagSuggestion)
		} else {
			fmt.Println("No files require changes.")
		}

		fmt.Println()
		err := showStatsOutput(totalStats)
		if err != nil {
			return err
		}

		// Default mode exits with status 1 if changes are required
		if len(changedFiles) > 0 {
			os.Exit(1)
		}
	}

	// Handle exitOnChange
	if exitOnChange && anyChanges {
		os.Exit(1)
	}

	return nil
}

// handleMultipleFiles processes multiple individual files
func handleMultipleFiles(filePaths []string, conv *converter.Converter, normaliseSmartQuotes bool,
	outputFile string, showDiff, showDiffInline, showRaw, showStats, saveInPlace, exitOnChange bool, width, maxFileSize int) error {

	if outputFile != "" {
		return fmt.Errorf("output file not supported when processing multiple files")
	}

	// Track changes and files for summary
	anyChanges := false
	var totalStats report.ChangeStats
	var changedFiles []string
	var unchangedFiles []string
	analyser := report.NewAnalyser(conv.GetAmericanToBritishDictionary())

	fmt.Printf("Processing %d file(s)...\n", len(filePaths))

	for _, filePath := range filePaths {
		// Read and process file content
		originalContent, err := fileutil.ReadFileContentWithMaxSize(filePath, maxFileSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to read file %s: %v\n", filePath, err)
			continue
		}

		// Convert content
		convertedContent := conv.ConvertToBritish(originalContent, normaliseSmartQuotes)
		hasChanges := originalContent != convertedContent

		if hasChanges {
			anyChanges = true
			changedFiles = append(changedFiles, filePath)

			// Save file if requested
			if saveInPlace {
				err = os.WriteFile(filePath, []byte(convertedContent), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to save changes to file %s: %v\n", filePath, err)
					continue
				}
			}

			// Handle diff output modes
			if showDiff {
				fmt.Printf("=== %s ===\n", filePath)
				err := showDiffOutput(originalContent, convertedContent, filePath, false)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to show diff for %s: %v\n", filePath, err)
				}
				fmt.Println()
			} else if showDiffInline {
				fmt.Printf("=== %s ===\n", filePath)
				err := showDiffOutput(originalContent, convertedContent, filePath, true)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to show diff for %s: %v\n", filePath, err)
				}
				fmt.Println()
			} else if showRaw {
				fmt.Printf("=== %s ===\n%s\n", filePath, convertedContent)
			}
		} else {
			unchangedFiles = append(unchangedFiles, filePath)
		}

		// Calculate stats
		stats := analyser.AnalyseChanges(originalContent, convertedContent)
		totalStats.TotalWords += stats.TotalWords
		totalStats.SpellingChanges += stats.SpellingChanges
		totalStats.UnitConversions += stats.UnitConversions
		totalStats.QuoteChanges += stats.QuoteChanges
	}

	// Show summary
	if len(changedFiles) > 0 {
		if saveInPlace {
			fmt.Printf("Saved changes to %d file(s):\n", len(changedFiles))
		} else {
			fmt.Printf("Found changes in %d file(s):\n", len(changedFiles))
		}
		for _, file := range changedFiles {
			fmt.Printf("  %s\n", file)
		}
	}

	if len(unchangedFiles) > 0 && !showDiff && !showDiffInline && !showRaw {
		fmt.Printf("No changes needed for %d file(s)\n", len(unchangedFiles))
	}

	// Show aggregate stats if changes were made or specifically requested
	if (totalStats.SpellingChanges > 0 || totalStats.UnitConversions > 0 || totalStats.QuoteChanges > 0) || showStats {
		fmt.Println()
		if saveInPlace {
			err := showStatsOutputWithMode(totalStats, true)
			if err != nil {
				return err
			}
		} else {
			err := showStatsOutputWithMode(totalStats, false)
			if err != nil {
				return err
			}
		}
	}

	// Handle exitOnChange
	if exitOnChange && anyChanges {
		os.Exit(1)
	}

	return nil
}

// convertFilename converts American spellings to British spellings in filenames
func convertFilename(filename string, converter *converter.Converter) (string, bool) {
	// Split filename into directory, basename, and extension
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Convert the name part (without extension)
	convertedName := converter.ConvertToBritish(nameWithoutExt, false)

	// Check if there were any changes
	hasChanges := nameWithoutExt != convertedName

	if !hasChanges {
		return filename, false
	}

	// Reconstruct the full path
	newBase := convertedName + ext
	var newFilename string
	if dir == "." {
		newFilename = newBase
	} else {
		newFilename = filepath.Join(dir, newBase)
	}

	return newFilename, true
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
