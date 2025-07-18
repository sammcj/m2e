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
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `m2e-cli - Convert American English to British English

Usage:
  m2e-cli [options] [text]
  echo "text" | m2e-cli [options]

Options:
  -input string
        Input file to convert. If not specified, reads from stdin.
  -output string
        Output file to write to. If not specified, writes to stdout.
  -units
        Freedom Unit Conversion (default: false)
  -no-smart-quotes
        Disable smart quote normalisation (default: false)
  -h, -help
        Show this help message

Examples:
  m2e-cli "color and flavor"
  m2e-cli -units "The room is 12 feet wide"
  m2e-cli -input input.txt -output output.txt -units
  echo "American text with 5 pounds" | m2e-cli -units
`)
}

func main() {
	inputFile := flag.String("input", "", "Input file to convert. If not specified, reads from stdin.")
	outputFile := flag.String("output", "", "Output file to write to. If not specified, writes to stdout.")
	convertUnits := flag.Bool("units", false, "Freedom Unit Conversion")
	noSmartQuotes := flag.Bool("no-smart-quotes", false, "Disable smart quote normalisation")
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

	var inputReader io.Reader
	var err error
	var inputText string

	// Check if there are non-flag arguments (direct text input)
	if flag.NArg() > 0 {
		inputText = strings.Join(flag.Args(), " ")
	} else if *inputFile != "" {
		file, err := os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing input file: %v\n", err)
			}
		}()
		inputReader = file
	} else {
		// Check if stdin has data available
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			// No piped input and no arguments - show usage
			printUsage()
			os.Exit(1)
		}
		inputReader = os.Stdin
	}

	var inputTextFinal string

	if inputText != "" {
		// Use direct text input
		inputTextFinal = inputText
	} else {
		// Read from file or stdin
		inputBytes, err := io.ReadAll(inputReader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		inputTextFinal = string(inputBytes)
	}

	conv, err := converter.NewConverter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing converter: %v\n", err)
		os.Exit(1)
	}

	// Set unit processing based on flag
	conv.SetUnitProcessingEnabled(*convertUnits)

	// Determine smart quotes setting (default is true, disable if flag is set)
	normaliseSmartQuotes := !*noSmartQuotes

	convertedText := conv.ConvertToBritish(inputTextFinal, normaliseSmartQuotes)

	var outputWriter io.Writer
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing output file: %v\n", err)
			}
		}()
		outputWriter = file
	} else {
		outputWriter = os.Stdout
	}

	_, err = fmt.Fprint(outputWriter, convertedText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
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
