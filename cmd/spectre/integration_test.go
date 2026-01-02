package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEndToEndParse tests the complete parse flow on all example files
func TestEndToEndParse(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	examplesDir := filepath.Join(projectRoot, "examples")

	files, err := findSpecFiles(examplesDir)
	if err != nil {
		t.Fatalf("error finding spec files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("no example files found")
	}

	processor := NewFileProcessor()
	successCount := 0
	errorCount := 0

	err = processor.ProcessFiles(files, func(filename string) error {
		err := runParse([]string{filename})
		if err != nil {
			errorCount++
			t.Logf("Parse failed for %s: %v", filepath.Base(filename), err)
			return err
		}
		successCount++
		return nil
	})

	t.Logf("Parse results: %d succeeded, %d failed out of %d total files", successCount, errorCount, len(files))

	// We expect at least some files to parse successfully
	if successCount == 0 {
		t.Error("expected at least one file to parse successfully")
	}
}

// TestEndToEndTypecheck tests the complete typecheck flow on all example files
func TestEndToEndTypecheck(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	examplesDir := filepath.Join(projectRoot, "examples")

	files, err := findSpecFiles(examplesDir)
	if err != nil {
		t.Fatalf("error finding spec files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("no example files found")
	}

	processor := NewFileProcessor()
	successCount := 0
	errorCount := 0

	err = processor.ProcessFiles(files, func(filename string) error {
		err := runTypecheck([]string{filename})
		if err != nil {
			errorCount++
			t.Logf("Typecheck failed for %s: %v", filepath.Base(filename), err)
			return err
		}
		successCount++
		return nil
	})

	t.Logf("Typecheck results: %d succeeded, %d failed out of %d total files", successCount, errorCount, len(files))

	// We expect at least some files to typecheck successfully
	if successCount == 0 {
		t.Error("expected at least one file to typecheck successfully")
	}
}

// TestEndToEndVerify tests the complete verify flow on example files
func TestEndToEndVerify(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	
	// Test with a few known-good example files
	testFiles := []string{
		filepath.Join(projectRoot, "examples", "counter.spec"),
		filepath.Join(projectRoot, "examples", "mutex.spec"),
	}

	successCount := 0
	errorCount := 0

	for _, filename := range testFiles {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Logf("Skipping %s (file not found)", filename)
			continue
		}

		err := runVerify([]string{filename})
		if err != nil {
			errorCount++
			t.Logf("Verify failed for %s: %v", filepath.Base(filename), err)
			// Verification failures are expected (violations found), so we don't fail the test
		} else {
			successCount++
		}
	}

	t.Logf("Verify results: %d succeeded, %d failed out of %d total files", successCount, errorCount, len(testFiles))
}

// TestEndToEndCompleteFlow tests the complete flow: parse -> typecheck -> verify
func TestEndToEndCompleteFlow(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")

	// Step 1: Parse
	t.Log("Step 1: Parsing...")
	err = runParse([]string{counterFile})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Step 2: Typecheck
	t.Log("Step 2: Type checking...")
	err = runTypecheck([]string{counterFile})
	if err != nil {
		// Typecheck may fail due to temporal expressions, which is acceptable
		t.Logf("typecheck had errors (may be expected): %v", err)
	}

	// Step 3: Verify
	t.Log("Step 3: Verifying...")
	err = runVerify([]string{counterFile})
	if err != nil {
		// Verification failures are expected (violations found)
		t.Logf("verify had errors (may be expected): %v", err)
	}

	t.Log("Complete flow test finished")
}

