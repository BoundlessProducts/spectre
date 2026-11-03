package main

// commands_repl_test.go — tests for Items 14 (REPL helpers) and 15 (spectre test).

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ── runOneSpecTest (Item 15) ───────────────────────────────────────────────────

func TestRunOneSpecTest_Passes(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action incr {
  require x < 3
  x' = x + 1
}
invariant bounded { x <= 3 }
`
	f := writeTempSpec(t, spec)
	r := runOneSpecTest(f, 100, 20, false)
	if r.err != nil {
		t.Fatalf("unexpected error: %v", r.err)
	}
	if !r.passed {
		t.Errorf("expected spec to pass, got %d violation(s)", r.violations)
	}
	if r.states == 0 {
		t.Error("expected non-zero states explored")
	}
	if r.elapsed < 0 {
		t.Error("elapsed should be non-negative")
	}
}

func TestRunOneSpecTest_FailsOnViolation(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant small { x < 2 }
`
	f := writeTempSpec(t, spec)
	r := runOneSpecTest(f, 100, 20, false)
	if r.err != nil {
		t.Fatalf("unexpected error: %v", r.err)
	}
	if r.passed {
		t.Error("expected invariant violation, but spec passed")
	}
	if r.violations == 0 {
		t.Error("expected at least one violation")
	}
}

func TestRunOneSpecTest_ParseError(t *testing.T) {
	f := writeTempSpec(t, "this is not valid spectre syntax @@@@")
	r := runOneSpecTest(f, 100, 20, false)
	// A parse error means either r.err != nil or a state machine error.
	// Either way the spec should not "pass".
	if r.passed {
		t.Error("a malformed spec should not pass")
	}
}

func TestRunOneSpecTest_MissingFile(t *testing.T) {
	r := runOneSpecTest("/nonexistent/path/test.spec", 100, 20, false)
	if r.err == nil {
		t.Error("expected error for non-existent file")
	}
	if r.passed {
		t.Error("missing file should not pass")
	}
}

func TestRunOneSpecTest_Elapsed(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action noop { x' = x }
`
	f := writeTempSpec(t, spec)
	before := time.Now()
	r := runOneSpecTest(f, 10, 5, false)
	after := time.Now()
	if r.elapsed < 0 || r.elapsed > after.Sub(before)+time.Second {
		t.Errorf("elapsed %s looks wrong", r.elapsed)
	}
}

// ── runTest integration (Item 15) ─────────────────────────────────────────────

func TestRunTest_SinglePassingSpec(t *testing.T) {
	spec := `
var c: int
init { c = 0 }
action step {
  require c < 5
  c' = c + 1
}
invariant ok { c <= 5 }
`
	f := writeTempSpec(t, spec)
	err := runTest([]string{f})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestRunTest_SingleFailingSpec(t *testing.T) {
	spec := `
var c: int
init { c = 0 }
action step { c' = c + 1 }
invariant small { c < 3 }
`
	f := writeTempSpec(t, spec)
	err := runTest([]string{f})
	if err == nil {
		t.Error("expected error from failing spec")
	}
}

func TestRunTest_MixedPassFail(t *testing.T) {
	pass := writeTempSpec(t, `
var x: int
init { x = 0 }
action noop { x' = x }
invariant zero { x == 0 }
`)
	fail := writeTempSpec(t, `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant forever_zero { x == 0 }
`)
	err := runTest([]string{pass, fail})
	if err == nil {
		t.Error("expected error because one spec fails")
	}
}

func TestRunTest_NoArgs(t *testing.T) {
	err := runTest([]string{})
	if err == nil {
		t.Error("expected usage error with no args")
	}
}

func TestRunTest_MaxStatesFlag(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
`
	f := writeTempSpec(t, spec)
	// With max-states = 2, exploration terminates early.
	err := runTest([]string{"--max-states", "2", f})
	// No invariant, so it should pass (no violations found).
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunTest_FailFast(t *testing.T) {
	fail1 := writeTempSpec(t, `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant small { x < 1 }
`)
	fail2 := writeTempSpec(t, `
var y: int
init { y = 0 }
action decr { y' = y - 1 }
invariant pos { y >= 0 }
`)
	err := runTest([]string{"--fail-fast", fail1, fail2})
	if err == nil {
		t.Error("expected error with --fail-fast and two failing specs")
	}
}

func TestRunTest_DirectoryOfSpecs(t *testing.T) {
	dir := t.TempDir()
	writeSpecAt(t, dir, "a.spec", `
var x: int
init { x = 0 }
action noop { x' = x }
`)
	writeSpecAt(t, dir, "b.spec", `
var y: int
init { y = 1 }
action noop { y' = y }
invariant pos { y > 0 }
`)
	err := runTest([]string{dir})
	if err != nil {
		t.Errorf("expected all passing, got: %v", err)
	}
}

// ── REPL unit: evalExpr ────────────────────────────────────────────────────────

func TestReplEvalExpr_BasicArithmetic(t *testing.T) {
	sess := makeCounterSession(t)
	// Just verify evalExpr runs without panic.
	sess.evalExpr("1 + 2")
}

func TestReplEvalExpr_StateVariable(t *testing.T) {
	sess := makeCounterSession(t)
	// "x" is a state variable in the counter spec; it should evaluate to 0.
	sess.evalExpr("x")
}

func TestReplEvalExpr_ParseError(t *testing.T) {
	sess := makeCounterSession(t)
	// Malformed expression should not panic.
	sess.evalExpr("@@@")
}

// ── REPL unit: printState / printActions ──────────────────────────────────────

func TestReplPrintState_NoPanic(t *testing.T) {
	sess := makeCounterSession(t)
	sess.printState() // just must not panic
}

func TestReplPrintActions_NoPanic(t *testing.T) {
	sess := makeCounterSession(t)
	sess.printActions()
}

// ── REPL unit: history / back ─────────────────────────────────────────────────

func TestReplHistory_Empty(t *testing.T) {
	sess := makeCounterSession(t)
	// handleLine("history") should not panic when history is empty.
	_ = sess.handleLine("history")
}

func TestReplBack_Empty(t *testing.T) {
	sess := makeCounterSession(t)
	_ = sess.handleLine("back")
}

func TestReplReset(t *testing.T) {
	sess := makeCounterSession(t)
	initial := sess.current
	// Step then reset.
	sess.tryAction("incr")
	sess.handleLine("reset")
	if sess.current != initial {
		t.Error("reset should restore initial state")
	}
	if len(sess.history) != 0 {
		t.Error("reset should clear history")
	}
}

func TestReplBack_AfterStep(t *testing.T) {
	sess := makeCounterSession(t)
	initial := sess.current
	sess.tryAction("incr")
	if sess.current == initial {
		t.Skip("action had no effect; cannot test back")
	}
	sess.handleLine("back")
	if sess.current != initial {
		t.Error("back should restore previous state")
	}
}

func TestReplQuit(t *testing.T) {
	sess := makeCounterSession(t)
	done := sess.handleLine("quit")
	if !done {
		t.Error("quit should signal done=true")
	}
}

func TestReplExit(t *testing.T) {
	sess := makeCounterSession(t)
	done := sess.handleLine("exit")
	if !done {
		t.Error("exit should signal done=true")
	}
}

func TestReplHelp_NoPanic(t *testing.T) {
	sess := makeCounterSession(t)
	_ = sess.handleLine("help")
}

func TestReplUnknownCommand(t *testing.T) {
	sess := makeCounterSession(t)
	// Should not panic.
	_ = sess.handleLine("this_is_not_a_command")
}

func TestReplCheck_NoPanic(t *testing.T) {
	sess := makeCounterSession(t)
	_ = sess.handleLine("check")
}

func TestReplInit_Idx0(t *testing.T) {
	sess := makeCounterSession(t)
	_ = sess.handleLine("init 0")
	if len(sess.history) != 0 {
		t.Error("init should reset history")
	}
}

func TestReplInit_OutOfBounds(t *testing.T) {
	sess := makeCounterSession(t)
	_ = sess.handleLine("init 999")
	// should not panic
}

// ── helpers ───────────────────────────────────────────────────────────────────

// writeTempSpec writes content to a temp .spec file and returns the path.
func writeTempSpec(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "spec_*.spec")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	f.Close()
	return f.Name()
}

// writeSpecAt writes a spec to a named file inside dir.
func writeSpecAt(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write spec %s: %v", name, err)
	}
}

// counterSpec is a minimal spec with a bounded counter.
const counterSpec = `
var x: int
init { x = 0 }
action incr {
  require x < 10
  x' = x + 1
}
invariant positive { x >= 0 }
`

// makeCounterSession builds a replSession from counterSpec.
func makeCounterSession(t *testing.T) *replSession {
	t.Helper()
	f := writeTempSpec(t, counterSpec)
	content, err := os.ReadFile(f)
	if err != nil {
		t.Fatalf("read counter spec: %v", err)
	}

	// Use the same startup sequence as runREPL.
	sess, err := buildReplSession(f, string(content))
	if err != nil {
		t.Fatalf("build repl session: %v", err)
	}
	return sess
}
