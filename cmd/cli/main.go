package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"murican-to-english/pkg/converter"
)

func main() {
	inputFile := flag.String("input", "", "Input file to convert. If not specified, reads from stdin.")
	outputFile := flag.String("output", "", "Output file to write to. If not specified, writes to stdout.")
	flag.Parse()

	var inputReader io.Reader
	var err error

	if *inputFile != "" {
		inputReader, err = os.Open(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
	} else {
		inputReader = os.Stdin
	}

	inputBytes, err := io.ReadAll(inputReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	conv, err := converter.NewConverter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing converter: %v\n", err)
		os.Exit(1)
	}

	convertedText := conv.ConvertToBritish(string(inputBytes), true)

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
