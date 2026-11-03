package main

import (
	"os"
	"path/filepath"
	"testing"
)

// ── runEvalRust error paths ───────────────────────────────────────────────────

func TestEvalRust_NoArgs_ReturnsError(t *testing.T) {
	err := runEvalRust(nil)
	if err == nil {
		t.Error("expected error when no args given, got nil")
	}
}

func TestEvalRust_EmptyArgs_ReturnsError(t *testing.T) {
	err := runEvalRust([]string{})
	if err == nil {
		t.Error("expected error for empty args")
	}
}

func TestEvalRust_BadDir_ReturnsError(t *testing.T) {
	err := runEvalRust([]string{"--dir", "/nonexistent_spectre_eval_dir"})
	if err == nil {
		t.Error("expected error for bad --dir, got nil")
	}
}

func TestEvalRust_SingleRsFile_Succeeds(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "rust_eval", "counter.rs")
	err := runEvalRust([]string{path})
	if err != nil {
		t.Fatalf("runEvalRust(counter.rs): %v", err)
	}
}

func TestEvalRust_MultipleFiles_Succeeds(t *testing.T) {
	dir := filepath.Join("..", "..", "testdata", "rust_eval")
	err := runEvalRust([]string{
		filepath.Join(dir, "counter.rs"),
		filepath.Join(dir, "bank_account.rs"),
	})
	if err != nil {
		t.Fatalf("runEvalRust with multiple files: %v", err)
	}
}

func TestEvalRust_DirFlag_Succeeds(t *testing.T) {
	dir := filepath.Join("..", "..", "testdata", "rust_eval")
	err := runEvalRust([]string{"--dir", dir})
	if err != nil {
		t.Fatalf("runEvalRust --dir: %v", err)
	}
}

func TestEvalRust_DirWithNoRsFiles_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	// Write a non-.rs file so the dir exists but has no .rs files.
	_ = os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hi"), 0644)
	err := runEvalRust([]string{"--dir", dir})
	if err == nil {
		t.Error("expected error for directory with no .rs files")
	}
}

func TestEvalRust_BadRsFile_ReturnsError(t *testing.T) {
	err := runEvalRust([]string{"/nonexistent/path/foo.rs"})
	if err == nil {
		t.Error("expected error for nonexistent .rs file")
	}
}

func TestEvalRust_CommandRegistered(t *testing.T) {
	if _, ok := Commands["eval-rust"]; !ok {
		t.Error("eval-rust command not registered in Commands map")
	}
}
