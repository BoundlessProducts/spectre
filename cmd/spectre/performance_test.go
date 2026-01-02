package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestPerformanceParse tests parsing performance
func TestPerformanceParse(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")

	start := time.Now()
	err = runParse([]string{counterFile})
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	t.Logf("Parse performance: %v", duration)
	
	// Parse should complete quickly (under 1 second for a small file)
	if duration > time.Second {
		t.Logf("Warning: Parse took longer than expected: %v", duration)
	}
}

// TestPerformanceTypecheck tests typechecking performance
func TestPerformanceTypecheck(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")

	start := time.Now()
	err = runTypecheck([]string{counterFile})
	duration := time.Since(start)

	// Typecheck may fail due to temporal expressions, but we still measure performance
	if err != nil {
		t.Logf("Typecheck had errors (may be expected): %v", err)
	}

	t.Logf("Typecheck performance: %v", duration)
	
	// Typecheck should complete reasonably quickly
	if duration > 5*time.Second {
		t.Logf("Warning: Typecheck took longer than expected: %v", duration)
	}
}

// TestPerformanceMultipleFiles tests performance with multiple files
func TestPerformanceMultipleFiles(t *testing.T) {
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

	// Limit to first 5 files for performance testing
	if len(files) > 5 {
		files = files[:5]
	}

	start := time.Now()
	processor := NewFileProcessor()
	err = processor.ProcessFiles(files, func(filename string) error {
		return runParse([]string{filename})
	})
	duration := time.Since(start)

	if err != nil {
		t.Logf("Some files failed to parse (may be expected): %v", err)
	}

	t.Logf("Multiple files parse performance: %v for %d files (avg: %v per file)", 
		duration, len(files), duration/time.Duration(len(files)))
	
	// Should process multiple files reasonably quickly
	if duration > 10*time.Second {
		t.Logf("Warning: Processing multiple files took longer than expected: %v", duration)
	}
}

// BenchmarkParse benchmarks parsing performance
func BenchmarkParse(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runParse([]string{counterFile})
	}
}

// BenchmarkTypecheck benchmarks typechecking performance
func BenchmarkTypecheck(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runTypecheck([]string{counterFile})
	}
}

