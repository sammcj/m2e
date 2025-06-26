package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sammcj/murican-to-english/pkg/converter"
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
  -h, -help
        Show this help message

Examples:
  m2e-cli "color and flavor"
  m2e-cli -input input.txt -output output.txt
  echo "American text" | m2e-cli
`)
}

func main() {
	inputFile := flag.String("input", "", "Input file to convert. If not specified, reads from stdin.")
	outputFile := flag.String("output", "", "Output file to write to. If not specified, writes to stdout.")
	help := flag.Bool("help", false, "Show help message")
	helpShort := flag.Bool("h", false, "Show help message")
	flag.Parse()

	if *help || *helpShort {
		printUsage()
		os.Exit(0)
	}

	var inputReader io.Reader
	var err error
	var inputText string

	// Check if there are non-flag arguments (direct text input)
	if flag.NArg() > 0 {
		inputText = flag.Arg(0)
		for i := 1; i < flag.NArg(); i++ {
			inputText += " " + flag.Arg(i)
		}
	} else if *inputFile != "" {
		inputReader, err = os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
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

	convertedText := conv.ConvertToBritish(inputTextFinal, true)

	var outputWriter io.Writer
	if *outputFile != "" {
		outputWriter, err = os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
	} else {
		outputWriter = os.Stdout
	}

	_, err = fmt.Fprint(outputWriter, convertedText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
}
