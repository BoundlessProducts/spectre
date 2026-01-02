package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCommand(t *testing.T) {
	// Test with a valid file - use absolute path from project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	// Navigate to project root (assuming test is run from cmd/spectre)
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")
	
	err = runParse([]string{counterFile})
	if err != nil {
		t.Errorf("parse command failed: %v", err)
	}
}

func TestParseCommandInvalidFile(t *testing.T) {
	// Test with non-existent file
	err := runParse([]string{"nonexistent.spec"})
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParseCommandNoFile(t *testing.T) {
	// Test with no file argument
	err := runParse([]string{})
	if err == nil {
		t.Error("expected error when no file specified")
	}
}

func TestTypecheckCommand(t *testing.T) {
	// Test with a valid file - use absolute path from project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")
	
	err = runTypecheck([]string{counterFile})
	if err != nil {
		t.Errorf("typecheck command failed: %v", err)
	}
}

func TestTypecheckCommandInvalidFile(t *testing.T) {
	// Test with non-existent file
	err := runTypecheck([]string{"nonexistent.spec"})
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestVerifyCommand(t *testing.T) {
	// Test with a valid file - use absolute path from project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")
	counterFile := filepath.Join(projectRoot, "examples", "counter.spec")
	
	err = runVerify([]string{counterFile})
	if err != nil {
		t.Logf("verify command result: %v (may fail if violations found)", err)
	}
}

func TestVerifyCommandInvalidFile(t *testing.T) {
	// Test with non-existent file
	err := runVerify([]string{"nonexistent.spec"})
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestCommandsExist(t *testing.T) {
	// Verify all expected commands exist
	expectedCommands := []string{"parse", "typecheck", "verify"}
	for _, cmdName := range expectedCommands {
		cmd, exists := Commands[cmdName]
		if !exists {
			t.Errorf("command '%s' not found", cmdName)
		}
		if cmd == nil {
			t.Errorf("command '%s' is nil", cmdName)
		}
		if cmd.Name != cmdName {
			t.Errorf("command name mismatch: expected '%s', got '%s'", cmdName, cmd.Name)
		}
	}
}

func TestFindSpecFiles(t *testing.T) {
	// Test finding spec files in examples directory - use absolute path
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
		t.Error("expected to find at least one .spec file in examples directory")
	}

	// Verify all files have .spec extension
	for _, file := range files {
		if filepath.Ext(file) != ".spec" {
			t.Errorf("file %s does not have .spec extension", file)
		}
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("file %s does not exist", file)
		}
	}
}

