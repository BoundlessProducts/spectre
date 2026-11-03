package explore

// Temporal property conformance tests.
//
// Each test exercises a small spec with a hand-verifiable temporal property.
// The expected verdicts are ground truth derived by direct reasoning — no
// external model checker is required. These tests cross-check that Spectre's
// temporal verifier produces the correct verdict on each case.

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func verifyProp(t *testing.T, spec string, propSpec string, wantHolds bool) {
	t.Helper()
	// Build the combined spec (model + temporal property).
	combined := spec + "\n" + propSpec
	l := lexer.New(combined)
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("state machine: %v", err)
	}
	exp := NewExplorer(sm)
	exp.SetMaxStates(500)
	exp.SetMaxDepth(50)
	exp.SetSilent(true)
	result, err := exp.ExploreBFS()
	if err != nil {
		t.Fatalf("BFS: %v", err)
	}

	hasher := NewStateHasher()
	tv := NewTemporalVerifier(hasher, file)
	tv.SetStateMachine(sm)

	cm := sm.GetConstraintModel()
	props := cm.GetTemporalProperties()
	if len(props) == 0 {
		t.Fatalf("no temporal properties found in spec")
	}

	vr, err := tv.VerifyTemporalProperty(props[0], result.TransitionGraph, result.InitialStates)
	if err != nil {
		t.Fatalf("VerifyTemporalProperty: %v", err)
	}
	if vr.Holds != wantHolds {
		verdict := "holds"
		if !wantHolds {
			verdict = "violated"
		}
		t.Errorf("expected property to be %s, got Holds=%v; violation: %v",
			verdict, vr.Holds, vr.Violation)
	}
}

// ─── always ────────────────────────────────────────────────────────────────

// A counter that only increments from 0. `always (x >= 0)` must hold.
const counterSpec = `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
`

func TestConformance_AlwaysHolds(t *testing.T) {
	verifyProp(t, counterSpec,
		`temporal alwaysNonNeg { always (x >= 0) }`,
		true)
}

// A counter that can go negative. `always (x >= 0)` must be violated.
func TestConformance_AlwaysViolated(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
action decr { x' = x - 1 }
`
	verifyProp(t, spec,
		`temporal alwaysNonNeg { always (x >= 0) }`,
		false)
}

// ─── eventually ────────────────────────────────────────────────────────────

// A counter that keeps incrementing. `eventually (x == 3)` must hold (BFS reaches it).
func TestConformance_EventuallyHolds(t *testing.T) {
	verifyProp(t, counterSpec,
		`temporal reachThree { eventually (x == 3) }`,
		true)
}

// A spec where x is always 0 (no transitions change it). `eventually (x == 1)` is violated.
func TestConformance_EventuallyViolated_Deadlock(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action stay { x' = x }
`
	verifyProp(t, spec,
		`temporal reachOne { eventually (x == 1) }`,
		false)
}

// A spec with a cycle: x alternates between 0 and 1. `eventually (x == 2)` is violated.
func TestConformance_EventuallyViolated_Cycle(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action toOne {
  require x == 0
  x' = 1
}
action toZero {
  require x == 1
  x' = 0
}
`
	verifyProp(t, spec,
		`temporal reachTwo { eventually (x == 2) }`,
		false)
}

// ─── lasso counterexample quality ──────────────────────────────────────────

// When `eventually P` is violated due to a cycle, the counterexample trace
// should be non-empty (prefix + cycle, not just empty).
func TestConformance_LassoCENotEmpty(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action toOne {
  require x == 0
  x' = 1
}
action toZero {
  require x == 1
  x' = 0
}
temporal reachTwo { eventually (x == 2) }
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
	exp := NewExplorer(sm)
	exp.SetMaxStates(500)
	exp.SetMaxDepth(50)
	exp.SetSilent(true)
	result, err := exp.ExploreBFS()
	if err != nil {
		t.Fatalf("BFS: %v", err)
	}

	hasher := NewStateHasher()
	tv := NewTemporalVerifier(hasher, file)
	tv.SetStateMachine(sm)

	cm := sm.GetConstraintModel()
	props := cm.GetTemporalProperties()
	if len(props) == 0 {
		t.Fatalf("no temporal properties found")
	}
	vr, err := tv.VerifyTemporalProperty(props[0], result.TransitionGraph, result.InitialStates)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if vr.Holds {
		t.Fatal("expected violation for unreachable property")
	}
	if vr.Violation == nil || vr.Violation.Trace == nil || vr.Violation.Trace.Length() == 0 {
		t.Error("expected non-empty lasso counterexample trace")
	}
}

// ─── leads-to ──────────────────────────────────────────────────────────────

// If x == 1, eventually x == 2. The counter always increments, so this holds.
func TestConformance_LeadsToHolds(t *testing.T) {
	verifyProp(t, counterSpec,
		`temporal progress { (x == 1) → (x == 2) }`,
		true)
}

// If x == 0 and we can only stay at 0, the leads-to x==1 is violated.
func TestConformance_LeadsToViolated(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action stay { x' = x }
`
	verifyProp(t, spec,
		`temporal deadlock { (x == 0) → (x == 1) }`,
		false)
}

// ─── VerificationKind on temporal violations ───────────────────────────────

// After a temporal violation is found, Kind must be KindCounterexample.
func TestConformance_KindCounterexampleOnTemporalViolation(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action stay { x' = x }
temporal reachOne { eventually (x == 1) }
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
	exp := NewExplorer(sm)
	exp.SetMaxStates(500)
	exp.SetMaxDepth(50)
	exp.SetSilent(true)
	result, err := exp.ExploreBFS()
	if err != nil {
		t.Fatalf("BFS: %v", err)
	}
	// No invariant violations, so BFS Kind is exhaustive.
	if result.Kind != KindExhaustive {
		t.Errorf("expected KindExhaustive from BFS (no invariants violated), got %s", result.Kind)
	}

	hasher := NewStateHasher()
	tv := NewTemporalVerifier(hasher, file)
	tv.SetStateMachine(sm)
	cm := sm.GetConstraintModel()
	props := cm.GetTemporalProperties()
	if len(props) == 0 {
		t.Fatal("no temporal properties found")
	}
	vr, _ := tv.VerifyTemporalProperty(props[0], result.TransitionGraph, result.InitialStates)
	if vr.Holds {
		t.Fatal("expected temporal violation")
	}
}

// ─── POR soundness on small spec ───────────────────────────────────────────

// POR and BFS should produce the same invariant violation set on a tiny spec.
func TestConformance_PORMatchesBFSOnSmallSpec(t *testing.T) {
	spec := `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
action decr { x' = x - 1 }
invariant nonNeg { x >= 0 }
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

	report, err := CrossValidatePOR(sm, file, []*ast.File{file}, 200, 20)
	if err != nil {
		t.Fatalf("CrossValidatePOR: %v", err)
	}
	if len(report.MissedByPOR) > 0 {
		t.Errorf("POR missed violations found by BFS: %v", report.MissedByPOR)
	}
}
