package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

// ── FormatSpec ────────────────────────────────────────────────────────────────

func parseForFmt(t *testing.T, src string) string {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	return FormatSpec(file)
}

func TestFmt_VarDecl(t *testing.T) {
	out := parseForFmt(t, "var x: int\n")
	if !strings.Contains(out, "var x: int") {
		t.Errorf("expected 'var x: int' in output: %q", out)
	}
}

func TestFmt_MultipleVars(t *testing.T) {
	out := parseForFmt(t, "var x: int\nvar y: bool\n")
	if !strings.Contains(out, "var x: int") || !strings.Contains(out, "var y: bool") {
		t.Errorf("expected both vars: %q", out)
	}
}

func TestFmt_EnumDecl(t *testing.T) {
	out := parseForFmt(t, "enum Color { Red, Green, Blue }\n")
	if !strings.Contains(out, "enum Color") {
		t.Errorf("expected enum: %q", out)
	}
	if !strings.Contains(out, "Red") || !strings.Contains(out, "Green") || !strings.Contains(out, "Blue") {
		t.Errorf("expected enum values: %q", out)
	}
}

func TestFmt_ActionDecl(t *testing.T) {
	out := parseForFmt(t, "var x: int\ninit { x = 0 }\naction incr {\n  require x < 10\n  x' = x + 1\n}\n")
	if !strings.Contains(out, "action incr {") {
		t.Errorf("expected action block: %q", out)
	}
	if !strings.Contains(out, "require x < 10") {
		t.Errorf("expected require: %q", out)
	}
	if !strings.Contains(out, "x' = x + 1") {
		t.Errorf("expected assignment: %q", out)
	}
}

func TestFmt_InitDecl(t *testing.T) {
	out := parseForFmt(t, "var x: int\ninit { x = 0 }\n")
	if !strings.Contains(out, "init {") {
		t.Errorf("expected init block: %q", out)
	}
	if !strings.Contains(out, "x = 0") {
		t.Errorf("expected assignment in init: %q", out)
	}
}

func TestFmt_InvariantDecl(t *testing.T) {
	out := parseForFmt(t, "var x: int\ninit { x = 0 }\ninvariant positive { x >= 0 }\n")
	if !strings.Contains(out, "invariant positive") {
		t.Errorf("expected invariant: %q", out)
	}
}

func TestFmt_FunctionDecl(t *testing.T) {
	out := parseForFmt(t, "fun double(n: int): int { return n + n }\n")
	if !strings.Contains(out, "fun double") {
		t.Errorf("expected function: %q", out)
	}
	if !strings.Contains(out, "return n + n") {
		t.Errorf("expected return: %q", out)
	}
}

func TestFmt_ModuleDecl(t *testing.T) {
	out := parseForFmt(t, `
module Math {
  public fun square(n: int): int {
    return n * n
  }
}
var x: int
init { x = 0 }
`)
	if !strings.Contains(out, "module Math {") {
		t.Errorf("expected module: %q", out)
	}
	if !strings.Contains(out, "fun square") {
		t.Errorf("expected module function: %q", out)
	}
}

func TestFmt_ParamDecl(t *testing.T) {
	out := parseForFmt(t, "param N: int\nvar x: int\ninit { x = 0 }\n")
	if !strings.Contains(out, "param N: int") {
		t.Errorf("expected param: %q", out)
	}
}

func TestFmt_OneOfInit(t *testing.T) {
	out := parseForFmt(t, `
var x: int
init oneOf {
  { x = 0 },
  { x = 1 }
}
`)
	if !strings.Contains(out, "init oneOf {") {
		t.Errorf("expected oneOf init: %q", out)
	}
}

func TestFmt_Idempotent(t *testing.T) {
	// Format twice and check the output is stable.
	src := "var x: int\nvar y: bool\ninit { x = 0  y = true }\naction step { x' = x + 1 }\ninvariant ok { x >= 0 }\n"
	out1 := parseForFmt(t, src)
	out2 := parseForFmt(t, out1)
	if out1 != out2 {
		t.Errorf("formatter is not idempotent:\nfirst: %q\nsecond: %q", out1, out2)
	}
}

func TestFmt_SetType(t *testing.T) {
	// Set type uses Set<int> syntax in the Spectre parser.
	out := parseForFmt(t, "var s: Set<int>\ninit { s = {} }\n")
	// The formatter emits the type using the AST type node; just check it includes "Set".
	if !strings.Contains(out, "Set") {
		t.Errorf("expected Set type in output: %q", out)
	}
}

// ── runFmt ────────────────────────────────────────────────────────────────────

func TestRunFmt_PrintsToStdout(t *testing.T) {
	f := writeTempSpec(t, "var x: int\ninit { x = 0 }\n")
	// Just check it doesn't error.
	err := runFmt([]string{f})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunFmt_ParseError(t *testing.T) {
	f := writeTempSpec(t, "action incr {\n") // unclosed block
	err := runFmt([]string{f})
	if err == nil {
		t.Error("expected error for parse failure")
	}
}

func TestRunFmt_MissingFile(t *testing.T) {
	err := runFmt([]string{"/nonexistent/path.spec"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRunFmt_WriteBack(t *testing.T) {
	src := "var x: int\ninit { x = 0 }\n"
	f := writeTempSpec(t, src)
	err := runFmt([]string{"-w", f})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// File should still be valid Spectre.
	content, _ := os.ReadFile(f)
	if !strings.Contains(string(content), "var x") {
		t.Errorf("expected var x in written file: %q", string(content))
	}
}

// ── check-json ────────────────────────────────────────────────────────────────

func TestCheckJSON_PassingSpec(t *testing.T) {
	f := writeTempSpec(t, `
var x: int
init { x = 0 }
action incr { require x < 5  x' = x + 1 }
invariant ok { x <= 5 }
`)
	out := captureCheckJSON(t, f)
	var jr JSONResult
	if err := json.Unmarshal([]byte(out), &jr); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, out)
	}
	if jr.Version != 1 {
		t.Errorf("expected version 1, got %d", jr.Version)
	}
	if len(jr.Violations) != 0 {
		t.Errorf("expected no violations, got %v", jr.Violations)
	}
	if jr.StatesExplored == 0 {
		t.Error("expected non-zero states explored")
	}
}

func TestCheckJSON_FailingSpec(t *testing.T) {
	f := writeTempSpec(t, `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant small { x < 3 }
`)
	out := captureCheckJSONIgnoreErr(t, f)
	var jr JSONResult
	if err := json.Unmarshal([]byte(out), &jr); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, out)
	}
	if len(jr.Violations) == 0 {
		t.Error("expected at least one violation")
	}
	if jr.Violations[0].Invariant != "small" {
		t.Errorf("expected 'small' violation, got %q", jr.Violations[0].Invariant)
	}
}

func TestCheckJSON_ViolationHasTrace(t *testing.T) {
	f := writeTempSpec(t, `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant small { x < 2 }
`)
	out := captureCheckJSONIgnoreErr(t, f)
	var jr JSONResult
	if err := json.Unmarshal([]byte(out), &jr); err != nil {
		t.Fatalf("parse JSON: %v", err)
	}
	if len(jr.Violations) == 0 {
		t.Fatal("no violations")
	}
	v := jr.Violations[0]
	if v.TraceLength == 0 {
		t.Error("expected non-empty trace")
	}
	if len(v.Trace) != v.TraceLength {
		t.Errorf("trace length mismatch: %d vs %d", len(v.Trace), v.TraceLength)
	}
}

func TestCheckJSON_ViolationHasFinalState(t *testing.T) {
	f := writeTempSpec(t, `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant small { x < 1 }
`)
	out := captureCheckJSONIgnoreErr(t, f)
	var jr JSONResult
	if err := json.Unmarshal([]byte(out), &jr); err != nil {
		t.Fatalf("parse JSON: %v", err)
	}
	if len(jr.Violations) == 0 {
		t.Fatal("no violations")
	}
	v := jr.Violations[0]
	if len(v.FinalState) == 0 {
		t.Error("expected non-empty final state")
	}
}

func TestCheckJSON_ParseError(t *testing.T) {
	f := writeTempSpec(t, "action incr {\n") // unclosed block
	out := captureCheckJSONIgnoreErr(t, f)
	var jr JSONResult
	if err := json.Unmarshal([]byte(out), &jr); err != nil {
		t.Fatalf("parse JSON: %v\noutput: %s", err, out)
	}
	if len(jr.Errors) == 0 {
		t.Error("expected errors for parse failure")
	}
	if jr.Kind != "error" {
		t.Errorf("expected kind 'error', got %q", jr.Kind)
	}
}

func TestCheckJSON_OutputFile(t *testing.T) {
	f := writeTempSpec(t, "var x: int\ninit { x = 0 }\naction noop { x' = x }\n")
	outFile := f + ".json"
	defer os.Remove(outFile)
	err := runCheckJSON([]string{"--output", outFile, f})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	content, readErr := os.ReadFile(outFile)
	if readErr != nil {
		t.Fatalf("reading output file: %v", readErr)
	}
	var jr JSONResult
	if err := json.Unmarshal(content, &jr); err != nil {
		t.Fatalf("parse JSON: %v", err)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// captureCheckJSON runs check-json and captures stdout by redirecting.
// It returns the JSON output as a string.
func captureCheckJSON(t *testing.T, specFile string) string {
	t.Helper()
	// Redirect stdout via a pipe.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runCheckJSON([]string{specFile})
	w.Close()
	os.Stdout = old
	var buf strings.Builder
	io := make([]byte, 4096)
	for {
		n, _ := r.Read(io)
		if n == 0 {
			break
		}
		buf.Write(io[:n])
	}
	if err != nil {
		t.Fatalf("check-json error: %v", err)
	}
	return buf.String()
}

func captureCheckJSONIgnoreErr(t *testing.T, specFile string) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	runCheckJSON([]string{specFile}) //nolint
	w.Close()
	os.Stdout = old
	var buf strings.Builder
	io := make([]byte, 4096)
	for {
		n, _ := r.Read(io)
		if n == 0 {
			break
		}
		buf.Write(io[:n])
	}
	return buf.String()
}
