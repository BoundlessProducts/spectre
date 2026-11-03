package z3

import (
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

const counterSpec = `
var x: int
var active: bool
param MAX: int = 10
init { x = 0  active = true }
action incr { require x < MAX  x' = x + 1 }
action deactivate { require active  active' = false }
invariant non_negative { x >= 0 }
`

const enumSpec = `
enum Color { Red, Green, Blue }
var c: Color
var n: int
init { c = Red  n = 0 }
action advance { c' = Green }
invariant non_neg { n >= 0 }
`

func parseSpec(t *testing.T, src string) *parser.Parser {
	t.Helper()
	l := lexer.New(src)
	return parser.New(l)
}

func encode(t *testing.T, src string, depth int) string {
	t.Helper()
	p := parseSpec(t, src)
	file := p.ParseFile()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	out, err := Encode(file, map[string]string{"MAX": "10"}, depth)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	return out
}

func TestEncode_ContainsDeclareConst(t *testing.T) {
	out := encode(t, counterSpec, 1)
	if !strings.Contains(out, "declare-const") {
		t.Error("expected declare-const in output")
	}
}

func TestEncode_IntVarDeclaredAsInt(t *testing.T) {
	out := encode(t, counterSpec, 1)
	if !strings.Contains(out, "(declare-const x_0 Int)") {
		t.Error("expected x_0 declared as Int")
	}
}

func TestEncode_BoolVarDeclaredAsBool(t *testing.T) {
	out := encode(t, counterSpec, 1)
	if !strings.Contains(out, "(declare-const active_0 Bool)") {
		t.Error("expected active_0 declared as Bool")
	}
}

func TestEncode_InitAsserted(t *testing.T) {
	out := encode(t, counterSpec, 1)
	if !strings.Contains(out, "(assert (= x_0 0))") {
		t.Error("expected init assertion for x_0")
	}
}

func TestEncode_CheckSatPresent(t *testing.T) {
	out := encode(t, counterSpec, 1)
	if !strings.Contains(out, "(check-sat)") {
		t.Error("expected (check-sat)")
	}
}

func TestEncode_GetModelPresent(t *testing.T) {
	out := encode(t, counterSpec, 1)
	if !strings.Contains(out, "(get-model)") {
		t.Error("expected (get-model)")
	}
}

func TestEncode_TransitionDisjunction(t *testing.T) {
	out := encode(t, counterSpec, 2)
	// Two actions → transition must use (or ...)
	if !strings.Contains(out, "(assert (or") {
		t.Error("expected (assert (or ...) for transition disjunction")
	}
}

func TestEncode_RequireInAction(t *testing.T) {
	out := encode(t, counterSpec, 1)
	// require x < MAX (MAX=10) → (< x_0 10)
	if !strings.Contains(out, "(< x_0 10)") {
		t.Errorf("expected require condition (< x_0 10) in output\n%s", out)
	}
}

func TestEncode_AssignmentInAction(t *testing.T) {
	out := encode(t, counterSpec, 1)
	// x' = x + 1 at step 0 → (= x_1 (+ x_0 1))
	if !strings.Contains(out, "x_1") {
		t.Error("expected x_1 in action assignment encoding")
	}
}

func TestEncode_UnchangedVar(t *testing.T) {
	out := encode(t, counterSpec, 1)
	// incr action assigns x but not active → active_1 = active_0
	if !strings.Contains(out, "(= active_1 active_0)") {
		t.Errorf("expected unchanged active: (= active_1 active_0)\n%s", out)
	}
}

func TestEncode_InvariantNegated(t *testing.T) {
	out := encode(t, counterSpec, 1)
	if !strings.Contains(out, "(not") {
		t.Error("expected invariant negation (not ...) in output")
	}
}

func TestEncode_DepthZero_OnlyStep0(t *testing.T) {
	out := encode(t, counterSpec, 0)
	if !strings.Contains(out, "x_0") {
		t.Error("expected x_0 at depth 0")
	}
	if strings.Contains(out, "x_1") {
		t.Error("depth 0: should not contain x_1")
	}
}

func TestEncode_DepthTwo_ThreeSteps(t *testing.T) {
	out := encode(t, counterSpec, 2)
	if !strings.Contains(out, "x_2") {
		t.Error("expected x_2 at depth 2")
	}
}

func TestEncode_ParamSubstituted(t *testing.T) {
	out := encode(t, counterSpec, 1)
	// MAX should not appear as a declare-const — substituted inline as 10
	if strings.Contains(out, "(declare-const MAX") {
		t.Error("param MAX should be substituted inline, not declared as SMT const")
	}
	if !strings.Contains(out, "10") {
		t.Error("param value 10 should appear in encoding")
	}
}

func TestEncode_EnumDeclaredAsDatatype(t *testing.T) {
	p := parseSpec(t, enumSpec)
	file := p.ParseFile()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	out, err := Encode(file, nil, 1)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !strings.Contains(out, "declare-datatypes") {
		t.Errorf("expected declare-datatypes for enum\n%s", out)
	}
	if !strings.Contains(out, "Color") {
		t.Error("expected enum name Color in datatypes declaration")
	}
}

func TestEncode_BinaryAdd_PlusSign(t *testing.T) {
	out := encode(t, counterSpec, 1)
	if !strings.Contains(out, "(+ x_0 1)") {
		t.Errorf("expected (+ x_0 1) for addition\n%s", out)
	}
}

func TestEncode_Neq_NotEquals(t *testing.T) {
	src := `
var x: int
init { x = 0 }
action tick { x' = x + 1 }
invariant not_five { x != 5 }
`
	p := parseSpec(t, src)
	file := p.ParseFile()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	out, err := Encode(file, nil, 1)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !strings.Contains(out, "(not (= x_") {
		t.Errorf("expected (not (= x_... for != operator\n%s", out)
	}
}

func TestEncode_BoolInit_TrueLowercase(t *testing.T) {
	out := encode(t, counterSpec, 1)
	// active initialised to true — SMT uses lowercase
	if !strings.Contains(out, "(assert (= active_0 true))") {
		t.Errorf("expected active_0 init to true (lowercase)\n%s", out)
	}
}

func TestEncode_InvariantAtStep0(t *testing.T) {
	out := encode(t, counterSpec, 0)
	// At depth 0 only step 0; invariant checked at step 0 → (not (>= x_0 0))
	if !strings.Contains(out, "x_0") || !strings.Contains(out, "not") {
		t.Errorf("expected invariant check at step 0\n%s", out)
	}
}
