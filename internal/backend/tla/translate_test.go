package tla

import (
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

const counterSpec = `
var x: int
var active: bool
param MAX: int
init { x = 0  active = true }
action incr { require x < MAX  x' = x + 1 }
action deactivate { require active  active' = false }
invariant non_negative { x >= 0 }
`

const temporalSpec = `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant non_negative { x >= 0 }
temporal liveness { always(x >= 0) }
temporal wf_progress { WF(incr) }
`

func parseSpec(t *testing.T, src string) *parser.Parser {
	t.Helper()
	l := lexer.New(src)
	return parser.New(l)
}

func translate(t *testing.T, src string) (tla string, cfg string) {
	t.Helper()
	p := parseSpec(t, src)
	file := p.ParseFile()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	mod, cfgOut, err := Translate(file, nil)
	if err != nil {
		t.Fatalf("Translate error: %v", err)
	}
	return mod, cfgOut
}

func TestTranslate_SimpleCounter_ContainsVARIABLES(t *testing.T) {
	tla, _ := translate(t, counterSpec)
	if !strings.Contains(tla, "VARIABLES") {
		t.Error("TLA+ output missing VARIABLES declaration")
	}
}

func TestTranslate_SimpleCounter_ContainsInit(t *testing.T) {
	tla, _ := translate(t, counterSpec)
	if !strings.Contains(tla, "Init ==") {
		t.Error("TLA+ output missing Init == definition")
	}
}

func TestTranslate_SimpleCounter_ContainsNext(t *testing.T) {
	tla, _ := translate(t, counterSpec)
	if !strings.Contains(tla, "Next ==") {
		t.Error("TLA+ output missing Next == definition")
	}
}

func TestTranslate_SimpleCounter_ContainsInvariant(t *testing.T) {
	tla, _ := translate(t, counterSpec)
	if !strings.Contains(tla, "non_negative") {
		t.Error("TLA+ output missing invariant name 'non_negative'")
	}
}

func TestTranslate_BoolExpr_TrueUppercase(t *testing.T) {
	tla, _ := translate(t, counterSpec)
	if !strings.Contains(tla, "TRUE") {
		t.Error("TLA+ output should contain TRUE (from 'active = true' in init)")
	}
}

func TestTranslate_Neq_UsesHash(t *testing.T) {
	src := `
var x: int
init { x = 0 }
action a { require x != 5  x' = x + 1 }
invariant inv { x >= 0 }
`
	tla, _ := translate(t, src)
	if !strings.Contains(tla, "#") {
		t.Error("TLA+ output should use '#' for != but did not")
	}
}

func TestTranslate_Leq_UsesEqLt(t *testing.T) {
	tla, _ := translate(t, counterSpec)
	// x < MAX uses Lt, but let's add a spec with <=
	src := `
var x: int
init { x = 0 }
action a { require x <= 10  x' = x + 1 }
invariant inv { x >= 0 }
`
	tla, _ = translate(t, src)
	if !strings.Contains(tla, "=<") {
		t.Error("TLA+ output should use '=<' for <= but did not")
	}
}

func TestTranslate_And_UsesSlashBackslash(t *testing.T) {
	src := `
var x: int
var y: bool
init { x = 0  y = true }
action a { require x < 10 and y  x' = x + 1 }
invariant inv { x >= 0 }
`
	tla, _ := translate(t, src)
	if !strings.Contains(tla, "/\\") {
		t.Error("TLA+ output should use '/\\' for 'and' but did not")
	}
}

func TestTranslate_PrimedVar_HasPrime(t *testing.T) {
	tla, _ := translate(t, counterSpec)
	if !strings.Contains(tla, "x'") {
		t.Error("TLA+ output should contain primed variable x' but did not")
	}
}

func TestTranslate_UNCHANGED_ListsOtherVars(t *testing.T) {
	tla, _ := translate(t, counterSpec)
	// incr only changes x, so active should be in UNCHANGED
	if !strings.Contains(tla, "UNCHANGED") {
		t.Error("TLA+ output should contain UNCHANGED clause for non-modified vars")
	}
	if !strings.Contains(tla, "active") {
		t.Error("TLA+ output should mention 'active' in UNCHANGED for incr action")
	}
}

func TestTranslate_Param_InCONSTANTS(t *testing.T) {
	tla, _ := translate(t, counterSpec)
	if !strings.Contains(tla, "CONSTANTS") {
		t.Error("TLA+ output should contain CONSTANTS block for param MAX")
	}
	if !strings.Contains(tla, "MAX") {
		t.Error("TLA+ output should contain param name MAX in CONSTANTS")
	}
}

func TestTranslate_Config_HasINIT(t *testing.T) {
	_, cfg := translate(t, counterSpec)
	if !strings.Contains(cfg, "INIT Init") {
		t.Errorf("config missing 'INIT Init', got:\n%s", cfg)
	}
}

func TestTranslate_Config_HasINVARIANT(t *testing.T) {
	_, cfg := translate(t, counterSpec)
	if !strings.Contains(cfg, "INVARIANT non_negative") {
		t.Errorf("config missing 'INVARIANT non_negative', got:\n%s", cfg)
	}
}

func TestTranslate_WF_InSpec(t *testing.T) {
	tla, _ := translate(t, temporalSpec)
	if !strings.Contains(tla, "WF_vars") {
		t.Error("TLA+ output should contain WF_vars for WF(incr) temporal property")
	}
}

func TestTranslate_Always_SquareBrackets(t *testing.T) {
	tla, _ := translate(t, temporalSpec)
	if !strings.Contains(tla, "[](") {
		t.Error("TLA+ output should use '[](' for always() but did not")
	}
}

func TestTranslate_Eventually_AngleBrackets(t *testing.T) {
	src := `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant inv { x >= 0 }
temporal ev { eventually(x > 5) }
`
	tla, _ := translate(t, src)
	if !strings.Contains(tla, "<>(") {
		t.Error("TLA+ output should use '<>(' for eventually() but did not")
	}
}

func TestTranslate_ParamOverride_InConfig(t *testing.T) {
	l := lexer.New(counterSpec)
	p := parser.New(l)
	file := p.ParseFile()
	_, cfg, err := Translate(file, map[string]string{"MAX": "42"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cfg, "MAX = 42") {
		t.Errorf("config should contain 'MAX = 42' with override, got:\n%s", cfg)
	}
}
