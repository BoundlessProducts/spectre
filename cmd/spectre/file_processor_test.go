package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFileProcessorSingleFile(t *testing.T) {
	processor := NewFileProcessor()
	
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")

	count := 0
	err = processor.ProcessFiles([]string{counterFile}, func(filename string) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("ProcessFiles failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 file processed, got %d", count)
	}
}

func TestFileProcessorMultipleFiles(t *testing.T) {
	processor := NewFileProcessor()
	
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")
	mutexFile := filepath.Join(projectRoot, "examples", "mutex.spec")

	count := 0
	err = processor.ProcessFiles([]string{counterFile, mutexFile}, func(filename string) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("ProcessFiles failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 files processed, got %d", count)
	}
}

func TestFileProcessorDirectory(t *testing.T) {
	processor := NewFileProcessor()
	
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	examplesDir := filepath.Join(projectRoot, "examples")

	count := 0
	err = processor.ProcessDirectory(examplesDir, func(filename string) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("ProcessDirectory failed: %v", err)
	}

	if count == 0 {
		t.Error("expected at least one file processed")
	}
}

func TestFileProcessorInvalidFile(t *testing.T) {
	processor := NewFileProcessor()

	err := processor.ProcessFiles([]string{"nonexistent.spec"}, func(filename string) error {
		return nil
	})

	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestFileProcessorWithErrors(t *testing.T) {
	processor := NewFileProcessor()
	
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")

	processError := fmt.Errorf("processing error")
	err = processor.ProcessFiles([]string{counterFile}, func(filename string) error {
		return processError
	})

	if err == nil {
		t.Error("expected error when processor returns error")
	}
}

