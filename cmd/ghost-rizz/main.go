package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ThothandSon/ghost-rizz/internal/generator"
	"github.com/ThothandSon/ghost-rizz/internal/processor"
)

// Allow mocking output in tests
var stdout io.Writer = os.Stdout
var stderr io.Writer = os.Stderr
var osExit = os.Exit

// Set by the release build.
var version = "dev"

const usage = `ghost-rizz: strip, fuzz or audit image and video metadata, in parallel.

Usage:
  ghost-rizz report   -in <dir> [-out <dir>]   list every file's metadata in a CSV
  ghost-rizz clean    -in <dir> -out <dir>     remove metadata
  ghost-rizz fuzz     -in <dir> -out <dir>     replace metadata with random values
  ghost-rizz generate -count <n> -out <dir>    create test images with metadata
  ghost-rizz version

Originals are never modified: results go to -out with a _clean or _fuzz suffix.
Run ghost-rizz <command> -h for a command's options.

Examples:
  ghost-rizz report -in ~/Pictures/trip
  ghost-rizz clean  -in ~/Pictures/trip -out ~/Pictures/trip-clean
`

func main() {
	osExit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < 1 {
		_, _ = fmt.Fprint(stderr, usage)
		return 1
	}

	switch args[0] {
	case "help", "-h", "-help", "--help":
		_, _ = fmt.Fprint(stdout, usage)
		return 0
	case "version", "-version", "--version":
		_, _ = fmt.Fprintf(stdout, "ghost-rizz %s\n", version)
		return 0
	case "generate":
		return generate(args[1:])
	case "clean":
		return process("clean", args[1:])
	case "fuzz":
		return process("fuzz", args[1:])
	case "report":
		return report(args[1:])
	default:
		_, _ = fmt.Fprintf(stderr, "ghost-rizz: unknown command %q\n\n%s", args[0], usage)
		return 1
	}
}

func newFlags(name, summary, example string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "%s\n\nUsage:\n  %s\n\nOptions:\n", summary, example)
		fs.PrintDefaults()
	}
	return fs
}

// parse returns the exit code to use when parsing ends the run (help or a bad flag).
func parse(fs *flag.FlagSet, args []string) (int, bool) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0, false
		}
		return 1, false
	}
	return 0, true
}

func generate(args []string) int {
	fs := newFlags("generate", "Create test JPEGs with realistic metadata (camera, software, GPS, dates).",
		"ghost-rizz generate -count 100 -out ./samples")
	count := fs.Int("count", 10, "number of images to generate")
	out := fs.String("out", "./input_photos", "output directory")
	if code, ok := parse(fs, args); !ok {
		return code
	}

	start := time.Now()
	if err := generator.GenerateImages(*count, *out); err != nil {
		_, _ = fmt.Fprintf(stderr, "ghost-rizz: Error generating images: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "✓ Generation complete. %d images in %s → %s\n", *count, elapsed(start), *out)
	return 0
}

func process(command string, args []string) int {
	summaries := map[string]string{
		"clean": "Remove all metadata from every supported image and video in a folder.",
		"fuzz":  "Replace metadata with random values, keeping the file structure intact.",
	}
	fs := newFlags(command, summaries[command], "ghost-rizz "+command+" -in ./photos -out ./photos-"+command)
	in := fs.String("in", "./input_photos", "folder to read (never modified)")
	out := fs.String("out", "./output_photos", "folder to write results to")
	mode := command
	if command == "fuzz" {
		// Before 1.0, clean was spelled "fuzz -mode clean"; keep accepting it.
		fs.StringVar(&mode, "mode", "fuzz", "deprecated: use the clean command instead of -mode clean")
	}
	if code, ok := parse(fs, args); !ok {
		return code
	}

	start := time.Now()
	summary, err := processor.ProcessDir(*in, *out, mode)
	if summary.Processed == 0 && err != nil {
		_, _ = fmt.Fprintf(stderr, "ghost-rizz: Error processing images: %v\n", err)
		return 1
	}

	verb := map[string]string{"clean": "Cleaned", "fuzz": "Fuzzed"}[mode]
	_, _ = fmt.Fprintf(stdout, "✓ Processing complete. %s %d %s in %s → %s\n", verb, summary.Processed, plural(summary.Processed, "file"), elapsed(start), *out)
	if summary.Skipped > 0 {
		_, _ = fmt.Fprintf(stdout, "  %d unsupported %s skipped\n", summary.Skipped, plural(summary.Skipped, "entry"))
	}
	if err != nil {
		reportFailures(summary.Failed, err)
		return 1
	}
	return 0
}

func report(args []string) int {
	fs := newFlags("report", "Write every file's metadata to report.csv without changing anything.",
		"ghost-rizz report -in ./photos [-out ./reports]")
	in := fs.String("in", "./input_photos", "folder to scan")
	out := fs.String("out", "", "folder for report.csv (defaults to the scanned folder)")
	if code, ok := parse(fs, args); !ok {
		return code
	}

	dir := *out
	if dir == "" {
		dir = *in
	}
	csvPath := filepath.Join(dir, "report.csv")
	start := time.Now()
	summary, err := processor.ReportDir(*in, csvPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "ghost-rizz: Error generating report: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "✓ Report complete. %d %s scanned in %s → %s\n", summary.Files, plural(summary.Files, "file"), elapsed(start), csvPath)
	_, _ = fmt.Fprintf(stdout, "  %d with metadata", summary.WithMetadata)
	if summary.Failed > 0 {
		_, _ = fmt.Fprintf(stdout, " · %d could not be read (see the Error column)", summary.Failed)
	}
	_, _ = fmt.Fprintln(stdout)
	return 0
}

func reportFailures(count int, err error) {
	_, _ = fmt.Fprintf(stderr, "✗ %d %s failed:\n", count, plural(count, "file"))
	lines := strings.Split(err.Error(), "\n")
	for i, line := range lines {
		if i == 10 {
			_, _ = fmt.Fprintf(stderr, "  … and %d more\n", len(lines)-10)
			break
		}
		_, _ = fmt.Fprintf(stderr, "  %s\n", line)
	}
}

func elapsed(start time.Time) string {
	d := time.Since(start)
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	if strings.HasSuffix(word, "y") {
		return strings.TrimSuffix(word, "y") + "ies"
	}
	return word + "s"
}
