package main

import (
	"fmt"
	"os"
	"strings"
)

// FileProcessor handles processing of single or multiple files
type FileProcessor struct {
	verbose bool
}

// NewFileProcessor creates a new file processor
func NewFileProcessor() *FileProcessor {
	return &FileProcessor{
		verbose: false,
	}
}

// SetVerbose sets verbose output mode
func (fp *FileProcessor) SetVerbose(verbose bool) {
	fp.verbose = verbose
}

// ProcessFiles processes one or more files
func (fp *FileProcessor) ProcessFiles(filenames []string, processor func(string) error) error {
	if len(filenames) == 0 {
		return fmt.Errorf("no files specified")
	}

	var errors []error
	successCount := 0

	for _, filename := range filenames {
		// Expand if it's a directory
		files, err := fp.expandFiles(filename)
		if err != nil {
			errors = append(errors, fmt.Errorf("error expanding %s: %w", filename, err))
			continue
		}

		for _, file := range files {
			if fp.verbose {
				fmt.Printf("Processing %s...\n", file)
			}

			err := processor(file)
			if err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", file, err))
			} else {
				successCount++
			}
		}
	}

	// Report summary
	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "\nSummary: %d file(s) succeeded, %d file(s) failed\n", successCount, len(errors))
		return fmt.Errorf("processing failed for %d file(s)", len(errors))
	}

	if len(filenames) > 1 || successCount > 1 {
		fmt.Printf("\nSummary: Successfully processed %d file(s)\n", successCount)
	}

	return nil
}

// expandFiles expands a filename pattern to actual files
// If filename is a directory, finds all .spec files in it
// If filename contains wildcards, expands them
// Otherwise, returns the single file
func (fp *FileProcessor) expandFiles(filename string) ([]string, error) {
	// Check if it's a directory
	info, err := os.Stat(filename)
	if err == nil && info.IsDir() {
		return findSpecFiles(filename)
	}

	// Check if file exists
	if _, err := os.Stat(filename); err != nil {
		return nil, fmt.Errorf("file not found: %s", filename)
	}

	// Check if it's a .spec file
	if !strings.HasSuffix(filename, ".spec") {
		return nil, fmt.Errorf("not a .spec file: %s", filename)
	}

	return []string{filename}, nil
}

// ProcessDirectory processes all .spec files in a directory
func (fp *FileProcessor) ProcessDirectory(dir string, processor func(string) error) error {
	files, err := findSpecFiles(dir)
	if err != nil {
		return fmt.Errorf("error finding files in %s: %w", dir, err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no .spec files found in %s", dir)
	}

	return fp.ProcessFiles(files, processor)
}

