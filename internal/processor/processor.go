package processor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func isSupportedFormat(filename string) bool {
	// Skip macOS resource fork files (._*)
	if strings.HasPrefix(filename, "._") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".heic" || ext == ".heif" ||
		ext == ".mp4" || ext == ".mov" || ext == ".mkv" || ext == ".avi" || ext == ".webm" || ext == ".m4v" || ext == ".3gp"
}

func isVideoFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".mp4" || ext == ".mov" || ext == ".mkv" || ext == ".avi" || ext == ".webm" || ext == ".m4v" || ext == ".3gp"
}

// ProcessImages reads supported image files from inDir, processes them using
// the specified mode, and writes the results to outDir.
func ProcessImages(inDir, outDir, mode string) error {
	absIn, _ := filepath.Abs(inDir)
	absOut, _ := filepath.Abs(outDir)
	if absIn == absOut {
		return errors.New("input and output directories must be different")
	}

	files, err := os.ReadDir(inDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(files))

	workers := runtime.GOMAXPROCS(0)
	sem := make(chan struct{}, workers)

	for _, file := range files {
		if file.IsDir() || !isSupportedFormat(file.Name()) {
			continue
		}

		// Only check for exiftool if we have an HEIC/HEIF file
		if strings.HasSuffix(strings.ToLower(file.Name()), ".heic") || strings.HasSuffix(strings.ToLower(file.Name()), ".heif") {
			if err := CheckExifTool(); err != nil {
				// We fail early here if a required tool for this file type is missing
				return fmt.Errorf("exiftool check failed for %s: %w", file.Name(), err)
			}
		}

		inPath := filepath.Join(inDir, file.Name())
		ext := filepath.Ext(file.Name())
		base := strings.TrimSuffix(file.Name(), ext)
		outPath := filepath.Join(outDir, fmt.Sprintf("%s_%s%s", base, mode, ext))

		wg.Add(1)
		sem <- struct{}{}
		go func(in, out string) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := processSingleImage(in, out, mode); err != nil {
				errCh <- fmt.Errorf("failed to process %s: %w", in, err)
			}
		}(inPath, outPath)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
