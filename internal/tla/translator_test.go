package tla

import (
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// parseSpec is a test helper that parses a Spectre spec string.
func parseSpec(t *testing.T, src string) *ast.File {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	return file
}

// assertContains checks that got contains sub (case-sensitive).
func assertContains(t *testing.T, label, got, sub string) {
	t.Helper()
	if !strings.Contains(got, sub) {
		t.Errorf("%s: expected output to contain %q\nGot:\n%s", label, sub, got)
	}
}

// assertNotContains checks that got does NOT contain sub.
func assertNotContains(t *testing.T, label, got, sub string) {
	t.Helper()
	if strings.Contains(got, sub) {
		t.Errorf("%s: expected output NOT to contain %q\nGot:\n%s", label, sub, got)
	}
}

// ── module header ─────────────────────────────────────────────────────────────

func TestTranslate_ModuleHeader(t *testing.T) {
	file := parseSpec(t, `var x: int init { x = 0 }`)
	r := Translate(file, "counter.spec")
	assertContains(t, "module header", r.TLA, "MODULE Counter")
	assertContains(t, "module name", r.ModuleName, "Counter")
}

func TestTranslate_ModuleNameFromFilePath(t *testing.T) {
	file := parseSpec(t, `var x: int init { x = 0 }`)
	r := Translate(file, "/some/deep/path/my_spec.spec")
	assertContains(t, "module name", r.TLA, "MODULE My_spec")
}

func TestTranslate_ExtendsLine(t *testing.T) {
	file := parseSpec(t, `var x: int init { x = 0 }`)
	r := Translate(file, "t.spec")
	assertContains(t, "EXTENDS", r.TLA, "EXTENDS Integers")
}

// ── variables ─────────────────────────────────────────────────────────────────

func TestTranslate_Variables(t *testing.T) {
	file := parseSpec(t, `
var counter: int
var flag: bool
init { counter = 0 flag = false }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "VARIABLES", r.TLA, "VARIABLES counter, flag")
}

func TestTranslate_TypeInvariant(t *testing.T) {
	file := parseSpec(t, `
var x: int
init { x = 0 }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "TypeInvariant", r.TLA, "TypeInvariant ==")
	assertContains(t, "x in Int", r.TLA, "x \\in Int")
}

func TestTranslate_BoolType(t *testing.T) {
	file := parseSpec(t, `
var flag: bool
init { flag = true }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "bool type", r.TLA, "\\in BOOLEAN")
}

// ── init ──────────────────────────────────────────────────────────────────────

func TestTranslate_DeterministicInit(t *testing.T) {
	file := parseSpec(t, `
var x: int
init { x = 0 }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "Init def", r.TLA, "Init ==")
	assertContains(t, "x = 0", r.TLA, "/\\ x = 0")
}

func TestTranslate_OneOfInit(t *testing.T) {
	file := parseSpec(t, `
var x: int
init oneOf {
  { x = 0 },
  { x = 1 }
}
`)
	r := Translate(file, "t.spec")
	assertContains(t, "oneOf init", r.TLA, "Init ==")
	// Both branches should appear.
	assertContains(t, "branch 0", r.TLA, "Init_0")
	assertContains(t, "branch 1", r.TLA, "Init_1")
}

// ── actions ───────────────────────────────────────────────────────────────────

func TestTranslate_SimpleAction(t *testing.T) {
	file := parseSpec(t, `
var x: int
init { x = 0 }
action incr {
  require x < 10
  x' = x + 1
}
`)
	r := Translate(file, "t.spec")
	assertContains(t, "action def", r.TLA, "incr ==")
	assertContains(t, "precondition", r.TLA, "/\\ (x < 10)")
	assertContains(t, "next-state", r.TLA, "x' = (x + 1)")
}

func TestTranslate_NextOperator(t *testing.T) {
	file := parseSpec(t, `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "Next def", r.TLA, "Next ==")
	assertContains(t, "action in Next", r.TLA, "\\/ incr")
}

func TestTranslate_SpecDef(t *testing.T) {
	file := parseSpec(t, `
var x: int
init { x = 0 }
action noop { x' = x }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "Spec def", r.TLA, "Spec ==")
	assertContains(t, "Init in Spec", r.TLA, "Init")
	assertContains(t, "Next in Spec", r.TLA, "Next")
}

// ── invariants ────────────────────────────────────────────────────────────────

func TestTranslate_Invariant(t *testing.T) {
	file := parseSpec(t, `
var x: int
init { x = 0 }
action noop { x' = x }
invariant nonNeg { x >= 0 }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "invariant def", r.TLA, "nonNeg ==")
	assertContains(t, "invariant condition", r.TLA, "(x >= 0)")
	assertContains(t, "THEOREM", r.TLA, "THEOREM Spec => []nonNeg")
}

// ── enums ─────────────────────────────────────────────────────────────────────

func TestTranslate_EnumDecl(t *testing.T) {
	file := parseSpec(t, `
enum Color { Red, Green, Blue }
var c: Color
init { c = Color.Red }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "CONSTANTS Color", r.TLA, "Color_VALUES")
	assertContains(t, "ASSUME enum", r.TLA, "ASSUME Color_VALUES")
}

// ── params ────────────────────────────────────────────────────────────────────

func TestTranslate_Params(t *testing.T) {
	file := parseSpec(t, `
param N: int
var x: int
init { x = 0 }
action incr { require x < N  x' = x + 1 }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "param as CONSTANT", r.TLA, "CONSTANTS N")
}

// ── functions ─────────────────────────────────────────────────────────────────

func TestTranslate_Function(t *testing.T) {
	file := parseSpec(t, `
fun double(n: int): int {
  return n + n
}
var x: int
init { x = 1 }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "function def", r.TLA, "double(n) ==")
}

// ── modules ───────────────────────────────────────────────────────────────────

func TestTranslate_Module(t *testing.T) {
	file := parseSpec(t, `
module Math {
  public fun square(n: int): int {
    return n * n
  }
}
var x: int
init { x = 4 }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "module function", r.TLA, "Math_square(n) ==")
}

// ── diagnostics ───────────────────────────────────────────────────────────────

func TestTranslate_StringTypeDiagnostic(t *testing.T) {
	file := parseSpec(t, `
var name: str
init { name = "hello" }
`)
	r := Translate(file, "t.spec")
	found := false
	for _, d := range r.Diagnostics {
		if d.Kind == LimitPartial && strings.Contains(d.Message, "string") {
			found = true
		}
	}
	if !found {
		t.Error("expected a PARTIAL diagnostic about string type")
	}
}

func TestTranslate_NoDiagnosticsForCleanSpec(t *testing.T) {
	file := parseSpec(t, `
var x: int
init { x = 0 }
action incr { require x < 5  x' = x + 1 }
invariant small { x <= 5 }
`)
	r := Translate(file, "t.spec")
	for _, d := range r.Diagnostics {
		if d.Kind == LimitUnsupported {
			t.Errorf("unexpected UNSUPPORTED diagnostic: %s", d)
		}
	}
}

// ── binary operators ──────────────────────────────────────────────────────────

func TestTranslate_BinaryOps(t *testing.T) {
	file := parseSpec(t, `
var x: int
var y: bool
init { x = 0  y = true }
invariant check { x + 1 > 0 && y == true }
`)
	r := Translate(file, "t.spec")
	assertContains(t, "add op", r.TLA, "+")
	assertContains(t, "and op", r.TLA, "/\\")
}

// ── sanitize ─────────────────────────────────────────────────────────────────

func TestSanitize_Basic(t *testing.T) {
	if sanitize("counter") != "counter" {
		t.Errorf("expected 'counter'")
	}
}

func TestSanitize_StartsWithDigit(t *testing.T) {
	if sanitize("1bad") != "v1bad" {
		t.Errorf("expected 'v1bad', got %q", sanitize("1bad"))
	}
}

func TestSanitize_SpecialChars(t *testing.T) {
	s := sanitize("my-var")
	if strings.Contains(s, "-") {
		t.Errorf("dash should be sanitized, got %q", s)
	}
}

// ── diagnostic String() ───────────────────────────────────────────────────────

func TestDiagnostic_StringWithPos(t *testing.T) {
	d := Diagnostic{Kind: LimitUnsupported, Message: "oops", Position: ast.Position{Line: 3, Column: 5}}
	s := d.String()
	if !strings.Contains(s, "3:5") {
		t.Errorf("expected position in string, got %q", s)
	}
}

func TestDiagnostic_StringNoPos(t *testing.T) {
	d := Diagnostic{Kind: LimitPartial, Message: "warning"}
	s := d.String()
	if !strings.Contains(s, "PARTIAL") {
		t.Errorf("expected kind in string, got %q", s)
	}
}
