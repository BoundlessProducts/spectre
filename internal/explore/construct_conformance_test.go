package explore

// Construct conformance tests (Item 12 — Cluster 3).
//
// These tests validate that collections, nondeterministic init choices,
// inline modules, and parameterized models are correctly executed during
// state-space exploration.  Each test uses a hand-verifiable spec whose
// expected state count or violation outcome is determined by direct reasoning.

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

// helper: build a StateMachine from an inline spec string.
func buildSMFromSpec(t *testing.T, spec string) *exec.StateMachine {
	t.Helper()
	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("state machine: %v", err)
	}
	return sm
}

// helper: run BFS and return the result.
func bfsResult(t *testing.T, sm *exec.StateMachine, maxStates, maxDepth int) *ExplorationResult {
	t.Helper()
	exp := NewExplorer(sm)
	exp.SetMaxStates(maxStates)
	exp.SetMaxDepth(maxDepth)
	exp.SetSilent(true)
	result, err := exp.ExploreBFS()
	if err != nil {
		t.Fatalf("BFS: %v", err)
	}
	return result
}

// ─── oneOf nondeterministic init ──────────────────────────────────────────────

// A spec with two oneOf initial states should yield ≥2 reachable states.
func TestConstruct_OneOfInitProducesMultipleInitialStates(t *testing.T) {
	spec := `
var x: int
init oneOf {
  { x = 0 },
  { x = 1 }
}
action noop { x' = x }
`
	sm := buildSMFromSpec(t, spec)
	result := bfsResult(t, sm, 100, 10)

	if result.StatesExplored < 2 {
		t.Errorf("expected ≥2 states from oneOf init, got %d", result.StatesExplored)
	}
	// Both branches must be reachable — exactly 2 distinct states (x=0, x=1).
	if result.StatesExplored != 2 {
		t.Errorf("expected exactly 2 states, got %d", result.StatesExplored)
	}
}

// An invariant that holds in one branch but not the other must be violated.
func TestConstruct_OneOfInitViolatesInvariantInOneBranch(t *testing.T) {
	spec := `
var x: int
init oneOf {
  { x = 0 },
  { x = 5 }
}
action noop { x' = x }
invariant xSmall { x < 3 }
`
	sm := buildSMFromSpec(t, spec)
	result := bfsResult(t, sm, 100, 10)

	if len(result.Violations) == 0 {
		t.Error("expected invariant violation in oneOf branch where x=5, got none")
	}
}

// ─── parameterized models ─────────────────────────────────────────────────────

// With N=3 the counter can reach 3; the invariant x <= N should hold when N=3
// and x only increments up to N (via a guard).
func TestConstruct_ParamBindingAffectsReachability(t *testing.T) {
	spec := `
param N: int
var x: int
init { x = 0 }
action incr {
  require x < N
  x' = x + 1
}
invariant bounded { x <= N }
`
	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("state machine: %v", err)
	}
	sm.SetParam("N", 3)

	result := bfsResult(t, sm, 200, 20)

	// States: x=0, x=1, x=2, x=3 (guard blocks x>3) → 4 states.
	if result.StatesExplored != 4 {
		t.Errorf("expected 4 states with N=3, got %d", result.StatesExplored)
	}
	if len(result.Violations) != 0 {
		t.Errorf("expected no violations with N=3, got %v", result.Violations)
	}
}

// Sweeping N: with N=1 there are 2 states; with N=2 there are 3 states.
func TestConstruct_ParamChangesStateSpace(t *testing.T) {
	spec := `
param N: int
var x: int
init { x = 0 }
action incr {
  require x < N
  x' = x + 1
}
`
	for _, tc := range []struct {
		n      int64
		states int
	}{
		{1, 2},
		{2, 3},
		{3, 4},
	} {
		l := lexer.New(spec)
		p := parser.New(l)
		file := p.ParseFile()
		if len(p.Errors()) > 0 {
			t.Fatalf("parse errors: %v", p.Errors())
		}
		sm, err := exec.NewStateMachine(file)
		if err != nil {
			t.Fatalf("state machine: %v", err)
		}
		sm.SetParam("N", tc.n)

		result := bfsResult(t, sm, 200, 20)
		if result.StatesExplored != tc.states {
			t.Errorf("N=%d: expected %d states, got %d", tc.n, tc.states, result.StatesExplored)
		}
	}
}

// ─── inline module declarations ───────────────────────────────────────────────

// An inline module with a pure function should be callable from an action.
// The function doubles a value; the invariant checks the result.
func TestConstruct_InlineModuleFunctionUsedInAction(t *testing.T) {
	spec := `
module Math {
  public fun double(n: int): int {
    return n + n
  }
}
var x: int
init { x = 1 }
action step {
  require x <= 4
  x' = Math.double(x)
}
invariant positive { x > 0 }
`
	sm := buildSMFromSpec(t, spec)
	result := bfsResult(t, sm, 200, 20)

	if len(result.Violations) != 0 {
		t.Errorf("expected no violations, got %v", result.Violations)
	}
	// States reachable: x=1 (init), x=2 (double(1)), x=4 (double(2)), x=8 (double(4) but
	// guard blocks when x>4 — wait, double(4)=8 which fails require x<=4; so states: 1,2,4).
	if result.StatesExplored < 3 {
		t.Errorf("expected ≥3 states, got %d", result.StatesExplored)
	}
}

// ─── enum parameterized actions ───────────────────────────────────────────────

// An enum with 3 values used as action parameter should explore all values.
func TestConstruct_EnumParamActionCoversAllValues(t *testing.T) {
	spec := `
enum Color { Red, Green, Blue }
var c: Color
init { c = Color.Red }
action setColor(v: Color) {
  c' = v
}
`
	sm := buildSMFromSpec(t, spec)
	result := bfsResult(t, sm, 200, 10)

	// All 3 enum values must be reachable.
	if result.StatesExplored != 3 {
		t.Errorf("expected 3 states (one per enum value), got %d", result.StatesExplored)
	}
}

// ─── invariant violation detection ───────────────────────────────────────────

// A spec where an action can violate an invariant must produce a violation report.
func TestConstruct_InvariantViolationDetected(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant nonNeg { x >= 0 }
action decr { x' = x - 1 }
invariant smallNeg { x >= -1 }
`
	sm := buildSMFromSpec(t, spec)
	result := bfsResult(t, sm, 500, 50)

	// x can go to -2, -3, … violating both nonNeg and smallNeg.
	if len(result.Violations) == 0 {
		t.Error("expected at least one invariant violation")
	}
}

// A spec with no reachable violation must not produce one.
func TestConstruct_NoSpuriousViolations(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action incr {
  require x < 5
  x' = x + 1
}
invariant bounded { x <= 5 }
`
	sm := buildSMFromSpec(t, spec)
	result := bfsResult(t, sm, 200, 20)

	if len(result.Violations) != 0 {
		t.Errorf("expected no violations, got %v", result.Violations)
	}
}
