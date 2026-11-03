package monitor_test

import (
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/monitor"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// buildSM parses a Spectre spec and returns a StateMachine.
func buildSM(t *testing.T, src string) (*exec.StateMachine, *ast.File) {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("NewStateMachine: %v", err)
	}
	return sm, file
}

// runBFS runs BFS on the spec and returns violations (up to maxStates).
func runBFS(t *testing.T, sm *exec.StateMachine, maxStates int) []*explore.Violation {
	t.Helper()
	e := explore.NewExplorer(sm)
	e.SetMaxStates(maxStates)
	e.SetSilent(true)
	result, err := e.ExploreBFS()
	if err != nil {
		t.Fatalf("BFS error: %v", err)
	}
	return result.Violations
}

const counterSpec = `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant small { x < 3 }
`

// ── ValidateTrace ─────────────────────────────────────────────────────────────

func TestValidateTrace_ConsistentViolation(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("expected at least one violation")
	}
	result := monitor.ValidateTrace(sm, violations[0])
	if !result.Consistent {
		t.Errorf("expected consistent result; note: %s", result.Note)
	}
	if !result.FinalViolated {
		t.Error("expected final step to violate the invariant")
	}
	if result.EarlyViolation {
		t.Error("unexpected early violation in trace")
	}
}

func TestValidateTrace_FinalViolated(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}
	r := monitor.ValidateTrace(sm, violations[0])
	if !r.FinalViolated {
		t.Error("expected FinalViolated = true")
	}
}

func TestValidateTrace_StepCount(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}
	v := violations[0]
	r := monitor.ValidateTrace(sm, v)
	if r.Steps != len(v.Path) {
		t.Errorf("expected Steps=%d, got %d", len(v.Path), r.Steps)
	}
}

func TestValidateTrace_TargetInvariant(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}
	r := monitor.ValidateTrace(sm, violations[0])
	if r.TargetInvariant != "small" {
		t.Errorf("expected TargetInvariant='small', got %q", r.TargetInvariant)
	}
}

func TestValidateTrace_EmptyPath_InitialViolation(t *testing.T) {
	src := `
var x: int
init { x = 5 }
invariant small { x < 3 }
`
	sm, _ := buildSM(t, src)
	violations := runBFS(t, sm, 50)
	if len(violations) == 0 {
		t.Fatal("expected violation at initial state")
	}
	r := monitor.ValidateTrace(sm, violations[0])
	// May have empty path if violation is at init state
	if !r.FinalViolated {
		t.Errorf("expected FinalViolated for initial violation; note: %s", r.Note)
	}
}

func TestValidateTrace_StepChecks_NotEmpty(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}
	r := monitor.ValidateTrace(sm, violations[0])
	if len(r.StepChecks) == 0 {
		t.Error("expected non-empty StepChecks")
	}
}

func TestValidateTrace_LastStepHasFailingInvariant(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}
	r := monitor.ValidateTrace(sm, violations[0])
	last := r.StepChecks[len(r.StepChecks)-1]
	if len(last.Failing) == 0 {
		t.Error("expected failing invariants in last step check")
	}
}

func TestValidateTrace_IntermediateStepsPass(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	if len(violations) == 0 || len(violations[0].Path) < 2 {
		t.Skip("need a violation with at least 2 steps")
	}
	r := monitor.ValidateTrace(sm, violations[0])
	for i, sc := range r.StepChecks[:len(r.StepChecks)-1] {
		if len(sc.Failing) > 0 {
			t.Errorf("step %d has unexpected invariant failure: %v", i, sc.Failing)
		}
	}
}

func TestValidateTrace_NoteIsNonEmpty(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}
	r := monitor.ValidateTrace(sm, violations[0])
	if r.Note == "" {
		t.Error("expected non-empty Note")
	}
}

// ── ValidateAll ───────────────────────────────────────────────────────────────

func TestValidateAll_ReturnsOnePerViolation(t *testing.T) {
	src := `
var x: int
var y: int
init { x = 0  y = 0 }
action incr_x { x' = x + 1 }
action incr_y { y' = y + 1 }
invariant x_small { x < 2 }
invariant y_small { y < 2 }
`
	sm, _ := buildSM(t, src)
	violations := runBFS(t, sm, 500)
	if len(violations) == 0 {
		t.Fatal("expected violations")
	}
	results := monitor.ValidateAll(sm, violations)
	if len(results) != len(violations) {
		t.Errorf("expected %d results, got %d", len(violations), len(results))
	}
}

func TestValidateAll_AllConsistent(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}
	results := monitor.ValidateAll(sm, violations)
	for i, r := range results {
		if !r.Consistent {
			t.Errorf("result %d not consistent: %s", i, r.Note)
		}
	}
}

// ── Summary ───────────────────────────────────────────────────────────────────

func TestSummary_ContainsCount(t *testing.T) {
	sm, _ := buildSM(t, counterSpec)
	violations := runBFS(t, sm, 200)
	results := monitor.ValidateAll(sm, violations)
	s := monitor.Summary(results)
	if !strings.Contains(s, "consistent") {
		t.Errorf("expected 'consistent' in summary: %q", s)
	}
}

func TestSummary_Empty(t *testing.T) {
	s := monitor.Summary(nil)
	if !strings.Contains(s, "0/0") {
		t.Errorf("expected '0/0' in empty summary: %q", s)
	}
}
