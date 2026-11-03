package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTempSpecDriver writes a spec file to a temp dir and returns the path.
func writeTempSpecDriver(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.spec")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempSpecDriver: %v", err)
	}
	return path
}

const simpleDriverSpec = `
var x: int
var active: bool
init { x = 0  active = true }
action incr { require x < 10  x' = x + 1 }
action deactivate { require active  active' = false }
invariant non_negative { x >= 0 }
`

// ── basic driver generation ───────────────────────────────────────────────────

func TestGenerateDriver_PrintsToStdout(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	err := runGenerateDriver([]string{f})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateDriver_ParseError(t *testing.T) {
	f := writeTempSpecDriver(t, "action incr {\n") // unclosed block
	err := runGenerateDriver([]string{f})
	if err == nil {
		t.Error("expected error for parse failure")
	}
}

func TestGenerateDriver_MissingFile(t *testing.T) {
	err := runGenerateDriver([]string{"/nonexistent/path.spec"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestGenerateDriver_NoArgs(t *testing.T) {
	err := runGenerateDriver([]string{})
	if err == nil {
		t.Error("expected error for no arguments")
	}
}

func TestGenerateDriver_WrongLang(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	err := runGenerateDriver([]string{"--lang", "python", f})
	if err == nil {
		t.Error("expected error for unsupported language")
	}
}

func TestGenerateDriver_OutputFile(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	outFile := filepath.Join(t.TempDir(), "driver.rs")
	err := runGenerateDriver([]string{"--output", outFile, f})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if len(content) == 0 {
		t.Error("expected non-empty driver output")
	}
}

func TestGenerateDriver_OutputContainsActions(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	outFile := filepath.Join(t.TempDir(), "driver.rs")
	if err := runGenerateDriver([]string{"--output", outFile, f}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, _ := os.ReadFile(outFile)
	src := string(content)
	if !strings.Contains(src, "incr") {
		t.Errorf("expected 'incr' action in driver; got %q", src[:min(200, len(src))])
	}
	if !strings.Contains(src, "deactivate") {
		t.Errorf("expected 'deactivate' action in driver")
	}
}

func TestGenerateDriver_OutputContainsVars(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	outFile := filepath.Join(t.TempDir(), "driver.rs")
	if err := runGenerateDriver([]string{"--output", outFile, f}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content, _ := os.ReadFile(outFile)
	src := string(content)
	if !strings.Contains(src, "x") {
		t.Errorf("expected variable 'x' in driver")
	}
}

// ── --ci flag ─────────────────────────────────────────────────────────────────

func TestGenerateDriver_CI_CreatesCargoSnippet(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	outFile := filepath.Join(t.TempDir(), "driver.rs")
	err := runGenerateDriver([]string{"--ci", "--output", outFile, f})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cargoFile := strings.TrimSuffix(outFile, ".rs") + ".cargo.toml"
	content, err := os.ReadFile(cargoFile)
	if err != nil {
		t.Fatalf("Cargo snippet not created: %v", err)
	}
	if !strings.Contains(string(content), "spectre-connect") {
		t.Errorf("expected spectre-connect in Cargo snippet; got %q", string(content))
	}
}

func TestGenerateDriver_CI_CargoSnippetHasDevDeps(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	outFile := filepath.Join(t.TempDir(), "driver.rs")
	if err := runGenerateDriver([]string{"--ci", "--output", outFile, f}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cargoFile := strings.TrimSuffix(outFile, ".rs") + ".cargo.toml"
	content, _ := os.ReadFile(cargoFile)
	if !strings.Contains(string(content), "[dev-dependencies]") {
		t.Errorf("expected [dev-dependencies] in Cargo snippet")
	}
}

func TestGenerateDriver_CI_CreatesCIWorkflow(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	outFile := filepath.Join(t.TempDir(), "driver.rs")
	err := runGenerateDriver([]string{"--ci", "--output", outFile, f})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ciFile := strings.TrimSuffix(outFile, ".rs") + ".ci.yml"
	content, err := os.ReadFile(ciFile)
	if err != nil {
		t.Fatalf("CI workflow not created: %v", err)
	}
	if !strings.Contains(string(content), "cargo test") {
		t.Errorf("expected 'cargo test' in CI workflow; got %q", string(content))
	}
}

func TestGenerateDriver_CI_WorkflowHasCheckout(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	outFile := filepath.Join(t.TempDir(), "driver.rs")
	if err := runGenerateDriver([]string{"--ci", "--output", outFile, f}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ciFile := strings.TrimSuffix(outFile, ".rs") + ".ci.yml"
	content, _ := os.ReadFile(ciFile)
	if !strings.Contains(string(content), "actions/checkout") {
		t.Errorf("expected 'actions/checkout' in CI workflow")
	}
}

func TestGenerateDriver_CI_WorkflowHasRustToolchain(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	outFile := filepath.Join(t.TempDir(), "driver.rs")
	if err := runGenerateDriver([]string{"--ci", "--output", outFile, f}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ciFile := strings.TrimSuffix(outFile, ".rs") + ".ci.yml"
	content, _ := os.ReadFile(ciFile)
	if !strings.Contains(string(content), "rust-toolchain") {
		t.Errorf("expected rust-toolchain step in CI workflow")
	}
}

func TestGenerateDriver_CI_StdoutModeIncludesSnippets(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	// stdout mode: capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runGenerateDriver([]string{"--ci", f})
	w.Close()
	os.Stdout = old
	var buf strings.Builder
	tmp := make([]byte, 8192)
	for {
		n, _ := r.Read(tmp)
		if n == 0 {
			break
		}
		buf.Write(tmp[:n])
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "spectre-connect") {
		t.Errorf("expected Cargo snippet in stdout output")
	}
	if !strings.Contains(out, "cargo test") {
		t.Errorf("expected CI workflow in stdout output")
	}
}

func TestGenerateDriver_NoCIFlag_NoArtifacts(t *testing.T) {
	f := writeTempSpecDriver(t, simpleDriverSpec)
	outFile := filepath.Join(t.TempDir(), "driver.rs")
	if err := runGenerateDriver([]string{"--output", outFile, f}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cargoFile := strings.TrimSuffix(outFile, ".rs") + ".cargo.toml"
	if _, err := os.ReadFile(cargoFile); err == nil {
		t.Error("cargo snippet should not be created without --ci flag")
	}
	ciFile := strings.TrimSuffix(outFile, ".rs") + ".ci.yml"
	if _, err := os.ReadFile(ciFile); err == nil {
		t.Error("CI workflow should not be created without --ci flag")
	}
}

// ── ciCargoSnippet / ciWorkflowYAML helpers ───────────────────────────────────

func TestCICargoSnippet_ContainsSpectreCrate(t *testing.T) {
	s := ciCargoSnippet("mydriver")
	if !strings.Contains(s, "spectre-connect") {
		t.Errorf("expected spectre-connect in Cargo snippet")
	}
}

func TestCICargoSnippet_ContainsTestDecl(t *testing.T) {
	s := ciCargoSnippet("mydriver")
	if !strings.Contains(s, "[[test]]") {
		t.Errorf("expected [[test]] in Cargo snippet")
	}
}

func TestCIWorkflowYAML_ValidYAMLLike(t *testing.T) {
	s := ciWorkflowYAML("mydriver")
	if !strings.Contains(s, "name:") || !strings.Contains(s, "jobs:") {
		t.Errorf("YAML missing name or jobs: %q", s)
	}
}

func TestCIWorkflowYAML_HasOnTriggers(t *testing.T) {
	s := ciWorkflowYAML("mydriver")
	if !strings.Contains(s, "push") || !strings.Contains(s, "pull_request") {
		t.Errorf("expected push/pull_request triggers in CI workflow")
	}
}

// min is a local helper (Go 1.21 min() not guaranteed in all build envs).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
