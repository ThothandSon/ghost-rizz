package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ThothandSon/ghost-rizz/internal/generator"
	"github.com/ThothandSon/ghost-rizz/internal/processor"
)

// Allow mocking output in tests
var stdout io.Writer = os.Stdout
var stderr io.Writer = os.Stderr
var osExit = os.Exit

func main() {
	osExit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < 1 {
		_, _ = fmt.Fprintln(stderr, "expected 'generate', 'clean', 'fuzz' or 'report' subcommands")
		return 1
	}

	generateCmd := flag.NewFlagSet("generate", flag.ContinueOnError)
	generateCmd.SetOutput(stderr)
	genCount := generateCmd.Int("count", 10, "number of images to generate")
	genOut := generateCmd.String("out", "./input_photos", "output directory for generated images")

	cleanCmd := flag.NewFlagSet("clean", flag.ContinueOnError)
	cleanCmd.SetOutput(stderr)
	cleanIn := cleanCmd.String("in", "./input_photos", "input directory of images")
	cleanOut := cleanCmd.String("out", "./output_photos", "output directory for processed images")

	fuzzCmd := flag.NewFlagSet("fuzz", flag.ContinueOnError)
	fuzzCmd.SetOutput(stderr)
	fuzzIn := fuzzCmd.String("in", "./input_photos", "input directory of images")
	fuzzOut := fuzzCmd.String("out", "./output_photos", "output directory for processed images")
	fuzzMode := fuzzCmd.String("mode", "fuzz", "mode of operation: 'clean' or 'fuzz'")

	reportCmd := flag.NewFlagSet("report", flag.ContinueOnError)
	reportCmd.SetOutput(stderr)
	reportIn := reportCmd.String("in", "./input_photos", "input directory of images")
	reportOut := reportCmd.String("out", "", "output directory for the report (defaults to input dir)")

	switch args[0] {
	case "generate":
		if err := generateCmd.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 1
		}
		_, _ = fmt.Fprintf(stdout, "Generating %d images to %s\n", *genCount, *genOut)
		err := generator.GenerateImages(*genCount, *genOut)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error generating images: %v\n", err)
			return 1
		}
		_, _ = fmt.Fprintln(stdout, "Generation complete.")
	case "clean":
		if err := cleanCmd.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 1
		}
		_, _ = fmt.Fprintf(stdout, "Cleaning images from %s to %s\n", *cleanIn, *cleanOut)
		err := processor.ProcessImages(*cleanIn, *cleanOut, "clean")
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error processing images: %v\n", err)
			return 1
		}
		_, _ = fmt.Fprintln(stdout, "Processing complete.")
	case "fuzz":
		if err := fuzzCmd.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 1
		}
		_, _ = fmt.Fprintf(stdout, "Processing images from %s to %s (mode: %s)\n", *fuzzIn, *fuzzOut, *fuzzMode)
		err := processor.ProcessImages(*fuzzIn, *fuzzOut, *fuzzMode)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error processing images: %v\n", err)
			return 1
		}
		_, _ = fmt.Fprintln(stdout, "Processing complete.")
	case "report":
		if err := reportCmd.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return 0
			}
			return 1
		}
		outDir := *reportOut
		if outDir == "" {
			outDir = *reportIn
		}
		outCSV := filepath.Join(outDir, "report.csv")
		_, _ = fmt.Fprintf(stdout, "Generating report from %s to %s\n", *reportIn, outCSV)
		err := processor.GenerateReport(*reportIn, outCSV)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error generating report: %v\n", err)
			return 1
		}
		_, _ = fmt.Fprintln(stdout, "Report complete.")
	default:
		_, _ = fmt.Fprintln(stderr, "expected 'generate', 'clean', 'fuzz' or 'report' subcommands")
		return 1
	}
	return 0
}
