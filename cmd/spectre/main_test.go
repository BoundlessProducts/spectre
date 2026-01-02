package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectSetup(t *testing.T) {
	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	// Navigate to project root (assuming test is run from cmd/spectre)
	projectRoot := filepath.Join(wd, "..", "..")
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); os.IsNotExist(err) {
		// Try current directory
		projectRoot = wd
	}
	
	// Test that the binary can be built
	testBinary := filepath.Join(projectRoot, "spectre_test")
	cmd := exec.Command("go", "build", "-o", testBinary, filepath.Join(projectRoot, "cmd", "spectre"))
	cmd.Dir = projectRoot
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build spectre: %v", err)
	}
	defer os.Remove(testBinary)

	// Test that the binary runs and shows usage
	cmd = exec.Command(testBinary)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected error when running without arguments")
	}
	
	expected := "Usage: spectre <command>"
	if !strings.Contains(string(output), expected) {
		t.Errorf("Expected output to contain '%s', got: %s", expected, string(output))
	}
}

