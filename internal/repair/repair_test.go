package repair_test

import (
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/internal/repair"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

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

func getViolations(t *testing.T, sm *exec.StateMachine, maxStates int) []*explore.Violation {
	t.Helper()
	e := explore.NewExplorer(sm)
	e.SetMaxStates(maxStates)
	e.SetSilent(true)
	result, err := e.ExploreBFS()
	if err != nil {
		t.Fatalf("BFS: %v", err)
	}
	return result.Violations
}

const counterSrc = `
var x: int
init { x = 0 }
action incr { x' = x + 1 }
invariant small { x < 3 }
`

// ── RepairKind.String() ───────────────────────────────────────────────────────

func TestRepairKindString(t *testing.T) {
	if repair.RepairStrengthenRequire.String() != "STRENGTHEN_REQUIRE" {
		t.Error("wrong string for RepairStrengthenRequire")
	}
	if repair.RepairAddInvariantGuard.String() != "ADD_INVARIANT_GUARD" {
		t.Error("wrong string for RepairAddInvariantGuard")
	}
}

// ── RepairCandidate.String() ──────────────────────────────────────────────────

func TestRepairCandidateString_ContainsCandidateLabel(t *testing.T) {
	c := repair.RepairCandidate{
		Kind:        repair.RepairStrengthenRequire,
		ActionName:  "incr",
		RequireExpr: "x < 3",
		Confidence:  "low",
		Description: "test",
	}
	s := c.String()
	if !strings.Contains(s, "CANDIDATE") {
		t.Errorf("expected CANDIDATE label in: %q", s)
	}
}

// ── SuggestRepairs ────────────────────────────────────────────────────────────

func TestSuggestRepairs_NotEmpty(t *testing.T) {
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}
	candidates := repair.SuggestRepairs(violations[0], file)
	if len(candidates) == 0 {
		t.Error("expected at least one repair candidate")
	}
}

func TestSuggestRepairs_HasLowConfidenceGuard(t *testing.T) {
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	found := false
	for _, c := range repair.SuggestRepairs(violations[0], file) {
		if c.Confidence == "low" {
			found = true
		}
	}
	if !found {
		t.Error("expected at least one low-confidence candidate")
	}
}

func TestSuggestRepairs_HasMediumConfidenceTighterBound(t *testing.T) {
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	found := false
	for _, c := range repair.SuggestRepairs(violations[0], file) {
		if c.Confidence == "medium" {
			found = true
		}
	}
	if !found {
		t.Error("expected at least one medium-confidence (tighter-bound) candidate")
	}
}

func TestSuggestRepairs_ActionNameMatchesTrigger(t *testing.T) {
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	for _, c := range repair.SuggestRepairs(violations[0], file) {
		if c.ActionName != "incr" {
			t.Errorf("expected ActionName='incr', got %q", c.ActionName)
		}
	}
}

func TestSuggestRepairs_RequireExprContainsInvariantVar(t *testing.T) {
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	for _, c := range repair.SuggestRepairs(violations[0], file) {
		if !strings.Contains(c.RequireExpr, "x") {
			t.Errorf("expected 'x' in RequireExpr: %q", c.RequireExpr)
		}
	}
}

func TestSuggestRepairs_EmptyPath_NoSuggestions(t *testing.T) {
	_, file := buildSM(t, counterSrc)
	// Violation with empty path and unknown invariant → no suggestions
	v := &explore.Violation{Invariant: "nonexistent"}
	if candidates := repair.SuggestRepairs(v, file); len(candidates) != 0 {
		t.Errorf("expected no candidates for unknown invariant, got %d", len(candidates))
	}
}

// ── tighterBound (via SuggestRepairs medium candidate) ────────────────────────

func TestTighterBound_LtProducesTighterVersion(t *testing.T) {
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	var mediumExpr string
	for _, c := range repair.SuggestRepairs(violations[0], file) {
		if c.Confidence == "medium" {
			mediumExpr = c.RequireExpr
		}
	}
	if mediumExpr == "" {
		t.Fatal("no medium confidence candidate")
	}
	// x < 3 should produce x < 3 - 1
	if !strings.Contains(mediumExpr, "x") {
		t.Errorf("tighter bound should contain 'x': %q", mediumExpr)
	}
}

// ── PatchSpec ─────────────────────────────────────────────────────────────────

func TestPatchSpec_InsertsRequire(t *testing.T) {
	src := "var x: int\ninit { x = 0 }\naction incr { x' = x + 1 }\n"
	patched := repair.PatchSpec(src, "incr", "x < 2")
	if !strings.Contains(patched, "require x < 2") {
		t.Errorf("expected require in patched spec: %q", patched)
	}
}

func TestPatchSpec_ActionNotFound_ReturnsEmpty(t *testing.T) {
	src := "var x: int\ninit { x = 0 }\n"
	if patched := repair.PatchSpec(src, "nonexistent", "x < 2"); patched != "" {
		t.Errorf("expected empty string for missing action")
	}
}

func TestPatchSpec_OriginalUnmodified(t *testing.T) {
	src := "var x: int\ninit { x = 0 }\naction incr { x' = x + 1 }\n"
	patched := repair.PatchSpec(src, "incr", "x < 5")
	if !strings.Contains(patched, "x' = x + 1") {
		t.Errorf("original action body should still be present: %q", patched)
	}
}

// ── ValidateRepair ────────────────────────────────────────────────────────────

func TestValidateRepair_MediumCandidateFixesViolation(t *testing.T) {
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}

	var medium *repair.RepairCandidate
	for _, c := range repair.SuggestRepairs(violations[0], file) {
		if c.Confidence == "medium" {
			cc := c
			medium = &cc
			break
		}
	}
	if medium == nil {
		t.Fatal("no medium confidence candidate")
	}

	rv := repair.ValidateRepair(medium, counterSrc, 10, 500)
	if !rv.OriginalFixed {
		t.Errorf("expected original violation fixed; note: %s", rv.Note)
	}
}

func TestValidateRepair_LowCandidateMayNotFixViolation(t *testing.T) {
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}

	var low *repair.RepairCandidate
	for _, c := range repair.SuggestRepairs(violations[0], file) {
		if c.Confidence == "low" {
			cc := c
			low = &cc
			break
		}
	}
	if low == nil {
		t.Fatal("no low confidence candidate")
	}

	rv := repair.ValidateRepair(low, counterSrc, 10, 500)
	// Low-confidence repair adds "require x < 3", which doesn't prevent the
	// violation (x=2→3 still fires because x=2 satisfies x<3).
	// We just check the validation ran without error.
	if rv.Note == "" {
		t.Error("expected non-empty note in validation result")
	}
}

func TestValidateRepair_SafeFlag(t *testing.T) {
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}

	var medium *repair.RepairCandidate
	for _, c := range repair.SuggestRepairs(violations[0], file) {
		if c.Confidence == "medium" {
			cc := c
			medium = &cc
			break
		}
	}
	if medium == nil {
		t.Fatal("no medium confidence candidate")
	}

	rv := repair.ValidateRepair(medium, counterSrc, 10, 500)
	// Safe means OriginalFixed && !HasDeadlock && ReachabilityDrop < 0.5
	if rv.OriginalFixed && !rv.HasDeadlock {
		// ReachabilityDrop may be small since we just add a tighter bound
		if rv.ReachabilityDrop < 0.5 && !rv.Safe {
			t.Error("expected Safe=true when all conditions met")
		}
	}
}

func TestValidateRepair_TighterBoundDetectsDeadlock(t *testing.T) {
	// The tighter bound "x < 3 - 1" stops incr at x>=2, leaving x=2 as a
	// deadlock state. ValidateRepair should detect this.
	sm, file := buildSM(t, counterSrc)
	violations := getViolations(t, sm, 200)
	if len(violations) == 0 {
		t.Fatal("no violations")
	}
	for _, c := range repair.SuggestRepairs(violations[0], file) {
		if c.Confidence == "medium" {
			rv := repair.ValidateRepair(&c, counterSrc, 10, 500)
			// Tighter bound fixes the invariant but introduces a deadlock.
			if !rv.OriginalFixed {
				t.Error("expected original violation to be fixed")
			}
			if !rv.HasDeadlock {
				t.Error("expected deadlock to be detected: incr is blocked at x=2 with no other actions")
			}
			// Safe should be false because of the deadlock.
			if rv.Safe {
				t.Error("expected Safe=false due to deadlock")
			}
			return
		}
	}
	t.Fatal("no medium candidate")
}
