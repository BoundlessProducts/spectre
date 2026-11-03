package lsp

import (
	"strings"
	"testing"
)

const counterSpec = `
var counter: int
init { counter = 0 }
action incr {
  require counter < 10
  counter' = counter + 1
}
invariant positive { counter >= 0 }
`

const enumSpec = `
enum Color { Red, Green, Blue }
var c: Color
init { c = Color.Red }
action setColor(v: Color) { c' = v }
`

const moduleSpec = `
module Math {
  public fun double(n: int): int {
    return n + n
  }
}
var x: int
init { x = 1 }
action step { x' = Math.double(x) }
`

// ── ParseDocument ─────────────────────────────────────────────────────────────

func TestParseDocument_Valid(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	if len(r.Errors) != 0 {
		t.Errorf("expected no errors, got %v", r.Errors)
	}
	if r.File == nil {
		t.Error("expected non-nil file")
	}
}

func TestParseDocument_SyntaxError(t *testing.T) {
	// Unclosed block produces a parse error.
	r := ParseDocument("file:///t.spec", "action incr {\n")
	if len(r.Errors) == 0 {
		t.Error("expected at least one error for unclosed block")
	}
	for _, d := range r.Errors {
		if d.Severity != SeverityError {
			t.Errorf("expected SeverityError, got %d", d.Severity)
		}
		if d.Source != "spectre" {
			t.Errorf("expected source 'spectre', got %q", d.Source)
		}
	}
}

func TestParseDocument_ErrorHasPosition(t *testing.T) {
	src := "var x: int\naction incr { require }\n"
	r := ParseDocument("file:///t.spec", src)
	if len(r.Errors) == 0 {
		t.Skip("no parse errors produced")
	}
	if r.Errors[0].Range.Start.Line < 0 {
		t.Error("expected non-negative line number")
	}
}

// ── CompletionItems ───────────────────────────────────────────────────────────

func TestCompletionItems_IncludesVars(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	items := CompletionItems(r.File)
	labels := labelSet(items)
	if !labels["counter"] {
		t.Error("expected 'counter' in completions")
	}
}

func TestCompletionItems_IncludesActions(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	items := CompletionItems(r.File)
	labels := labelSet(items)
	if !labels["incr"] {
		t.Error("expected 'incr' in completions")
	}
}

func TestCompletionItems_IncludesInvariants(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	items := CompletionItems(r.File)
	labels := labelSet(items)
	if !labels["positive"] {
		t.Error("expected 'positive' invariant in completions")
	}
}

func TestCompletionItems_IncludesEnumValues(t *testing.T) {
	r := ParseDocument("file:///t.spec", enumSpec)
	items := CompletionItems(r.File)
	labels := labelSet(items)
	if !labels["Color"] {
		t.Error("expected 'Color' enum in completions")
	}
	if !labels["Color.Red"] {
		t.Error("expected 'Color.Red' in completions")
	}
}

func TestCompletionItems_IncludesModuleFunctions(t *testing.T) {
	r := ParseDocument("file:///t.spec", moduleSpec)
	items := CompletionItems(r.File)
	labels := labelSet(items)
	if !labels["Math"] {
		t.Error("expected 'Math' module in completions")
	}
	if !labels["Math.double"] {
		t.Error("expected 'Math.double' in completions")
	}
}

func TestCompletionItems_Sorted(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	items := CompletionItems(r.File)
	for i := 1; i < len(items); i++ {
		if items[i-1].Label > items[i].Label {
			t.Errorf("completions not sorted: %q > %q", items[i-1].Label, items[i].Label)
		}
	}
}

func TestCompletionItems_NoDuplicates(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	items := CompletionItems(r.File)
	seen := map[string]bool{}
	for _, it := range items {
		if seen[it.Label] {
			t.Errorf("duplicate completion label: %q", it.Label)
		}
		seen[it.Label] = true
	}
}

// ── HoverInfo ─────────────────────────────────────────────────────────────────

func TestHoverInfo_Variable(t *testing.T) {
	// "counter" starts on line 1 (0-indexed), col 4 in counterSpec.
	// Let's find the exact offset dynamically.
	r := ParseDocument("file:///t.spec", counterSpec)
	pos := findWordPos(counterSpec, "counter")
	h := HoverInfo(r.File, counterSpec, pos)
	if h == nil {
		t.Fatal("expected hover info for 'counter'")
	}
	if !strings.Contains(h.Contents.Value, "counter") {
		t.Errorf("expected 'counter' in hover: %q", h.Contents.Value)
	}
}

func TestHoverInfo_Action(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	pos := findWordPos(counterSpec, "incr")
	h := HoverInfo(r.File, counterSpec, pos)
	if h == nil {
		t.Fatal("expected hover info for 'incr'")
	}
	if !strings.Contains(h.Contents.Value, "action") {
		t.Errorf("expected 'action' in hover: %q", h.Contents.Value)
	}
}

func TestHoverInfo_Enum(t *testing.T) {
	r := ParseDocument("file:///t.spec", enumSpec)
	pos := findWordPos(enumSpec, "Color")
	h := HoverInfo(r.File, enumSpec, pos)
	if h == nil {
		t.Fatal("expected hover info for 'Color'")
	}
	if !strings.Contains(h.Contents.Value, "enum") {
		t.Errorf("expected 'enum' in hover: %q", h.Contents.Value)
	}
}

func TestHoverInfo_NoWord(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	h := HoverInfo(r.File, counterSpec, Position{Line: 0, Character: 0})
	// Blank or newline position — may or may not return nil depending on content.
	_ = h // just must not panic
}

func TestHoverInfo_UnknownWord(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	// Position that lands on "var" (keyword, not a declared name).
	pos := findWordPos(counterSpec, "var")
	h := HoverInfo(r.File, counterSpec, pos)
	// "var" is not a declared name, so nil is expected.
	_ = h // just must not panic
}

func TestHoverInfo_Module(t *testing.T) {
	r := ParseDocument("file:///t.spec", moduleSpec)
	pos := findWordPos(moduleSpec, "Math")
	h := HoverInfo(r.File, moduleSpec, pos)
	if h == nil {
		t.Fatal("expected hover info for 'Math'")
	}
	if !strings.Contains(h.Contents.Value, "module") {
		t.Errorf("expected 'module' in hover: %q", h.Contents.Value)
	}
}

// ── DefinitionLocation ────────────────────────────────────────────────────────

func TestDefinitionLocation_Variable(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	pos := findWordPos(counterSpec, "counter")
	loc := DefinitionLocation(r.File, "file:///t.spec", counterSpec, pos)
	if loc == nil {
		t.Fatal("expected definition location for 'counter'")
	}
	if loc.URI != "file:///t.spec" {
		t.Errorf("unexpected URI: %q", loc.URI)
	}
	if loc.Range.Start.Line < 0 {
		t.Error("expected valid line")
	}
}

func TestDefinitionLocation_Action(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	pos := findWordPos(counterSpec, "incr")
	loc := DefinitionLocation(r.File, "file:///t.spec", counterSpec, pos)
	if loc == nil {
		t.Fatal("expected definition location for 'incr'")
	}
}

func TestDefinitionLocation_Unknown(t *testing.T) {
	r := ParseDocument("file:///t.spec", counterSpec)
	loc := DefinitionLocation(r.File, "file:///t.spec", counterSpec, Position{Line: 0, Character: 0})
	// Blank position — may return nil.
	_ = loc
}

func TestDefinitionLocation_Enum(t *testing.T) {
	r := ParseDocument("file:///t.spec", enumSpec)
	pos := findWordPos(enumSpec, "Color")
	loc := DefinitionLocation(r.File, "file:///t.spec", enumSpec, pos)
	if loc == nil {
		t.Fatal("expected definition for 'Color'")
	}
}

// ── wordAtPosition ────────────────────────────────────────────────────────────

func TestWordAtPosition_Basic(t *testing.T) {
	content := "var x: int"
	// Position 4 = 'x'
	w := wordAtPosition(content, Position{Line: 0, Character: 4})
	if w != "x" {
		t.Errorf("expected 'x', got %q", w)
	}
}

func TestWordAtPosition_OutOfBounds(t *testing.T) {
	w := wordAtPosition("var x: int", Position{Line: 99, Character: 0})
	if w != "" {
		t.Errorf("expected empty string for out-of-bounds, got %q", w)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func labelSet(items []CompletionItem) map[string]bool {
	m := map[string]bool{}
	for _, it := range items {
		m[it.Label] = true
	}
	return m
}

// findWordPos returns the LSP position of the first occurrence of word in content.
func findWordPos(content, word string) Position {
	lines := strings.Split(content, "\n")
	for l, line := range lines {
		if idx := strings.Index(line, word); idx >= 0 {
			return Position{Line: l, Character: idx}
		}
	}
	return Position{}
}
