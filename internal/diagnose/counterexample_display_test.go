package diagnose

import (
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/state"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func makeState(vars map[string]int64) *state.State {
	s := state.NewState()
	for k, v := range vars {
		iv := v
		s.SetVariable(k, &state.PrimitiveValue{TypeName: "int", IntValue: &iv})
	}
	return s
}

func makeTransition(from, to *state.State, action string) *explore.Transition {
	return &explore.Transition{
		FromState: from,
		ToState:   to,
		Action:    action,
	}
}

// ── StateDiff ─────────────────────────────────────────────────────────────────

func TestStateDiff_NoChange(t *testing.T) {
	s1 := makeState(map[string]int64{"x": 0, "y": 1})
	s2 := makeState(map[string]int64{"x": 0, "y": 1})
	diffs := StateDiff(s1, s2)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs, got %v", diffs)
	}
}

func TestStateDiff_OneChange(t *testing.T) {
	s1 := makeState(map[string]int64{"x": 0, "y": 1})
	s2 := makeState(map[string]int64{"x": 5, "y": 1})
	diffs := StateDiff(s1, s2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Name != "x" {
		t.Errorf("expected diff on x, got %s", diffs[0].Name)
	}
}

func TestStateDiff_MultipleChanges(t *testing.T) {
	s1 := makeState(map[string]int64{"x": 0, "y": 0})
	s2 := makeState(map[string]int64{"x": 1, "y": 2})
	diffs := StateDiff(s1, s2)
	if len(diffs) != 2 {
		t.Errorf("expected 2 diffs, got %d", len(diffs))
	}
}

func TestStateDiff_Sorted(t *testing.T) {
	s1 := makeState(map[string]int64{"z": 0, "a": 0, "m": 0})
	s2 := makeState(map[string]int64{"z": 1, "a": 1, "m": 1})
	diffs := StateDiff(s1, s2)
	names := make([]string, len(diffs))
	for i, d := range diffs {
		names[i] = d.Name
	}
	if names[0] != "a" || names[1] != "m" || names[2] != "z" {
		t.Errorf("expected sorted names a,m,z got %v", names)
	}
}

func TestStateDiff_NewVariable(t *testing.T) {
	s1 := makeState(map[string]int64{"x": 0})
	s2 := makeState(map[string]int64{"x": 0, "y": 1})
	diffs := StateDiff(s1, s2)
	if len(diffs) != 1 || diffs[0].Name != "y" {
		t.Errorf("expected diff on new var y, got %v", diffs)
	}
	if diffs[0].Before != nil {
		t.Error("Before should be nil for new variable")
	}
}

func TestStateDiff_RemovedVariable(t *testing.T) {
	s1 := makeState(map[string]int64{"x": 0, "y": 1})
	s2 := makeState(map[string]int64{"x": 0})
	diffs := StateDiff(s1, s2)
	if len(diffs) != 1 || diffs[0].Name != "y" {
		t.Errorf("expected diff on removed var y, got %v", diffs)
	}
	if diffs[0].After != nil {
		t.Error("After should be nil for removed variable")
	}
}

// ── FormatViolation ───────────────────────────────────────────────────────────

func makeViolation(inv string, path []*explore.Transition, finalState *state.State) *explore.Violation {
	return &explore.Violation{
		Invariant:   inv,
		Description: inv,
		Path:        path,
		State:       finalState,
	}
}

func TestFormatViolation_EmptyPath(t *testing.T) {
	s := makeState(map[string]int64{"x": 5})
	v := makeViolation("small", nil, s)
	out := FormatViolationString(v, DisplayOptions{Indent: "  "})
	if !strings.Contains(out, "small") {
		t.Error("expected invariant name in output")
	}
	if !strings.Contains(out, "initial state") {
		t.Error("expected 'initial state' note for empty path")
	}
}

func TestFormatViolation_OneStep(t *testing.T) {
	s0 := makeState(map[string]int64{"x": 0})
	s1 := makeState(map[string]int64{"x": 1})
	tr := makeTransition(s0, s1, "incr")
	v := makeViolation("small", []*explore.Transition{tr}, s1)

	out := FormatViolationString(v, DisplayOptions{Indent: ""})
	if !strings.Contains(out, "incr") {
		t.Error("expected action name 'incr'")
	}
	if !strings.Contains(out, "x: 0 → 1") {
		t.Errorf("expected state diff 'x: 0 → 1', got:\n%s", out)
	}
}

func TestFormatViolation_MultiStep(t *testing.T) {
	s0 := makeState(map[string]int64{"x": 0})
	s1 := makeState(map[string]int64{"x": 1})
	s2 := makeState(map[string]int64{"x": 2})
	path := []*explore.Transition{
		makeTransition(s0, s1, "incr"),
		makeTransition(s1, s2, "incr"),
	}
	v := makeViolation("small", path, s2)
	out := FormatViolationString(v, DisplayOptions{Indent: ""})
	if !strings.Contains(out, "step 1") || !strings.Contains(out, "step 2") {
		t.Errorf("expected step numbers in output:\n%s", out)
	}
}

func TestFormatViolation_NoChangeDiff(t *testing.T) {
	s0 := makeState(map[string]int64{"x": 0})
	s1 := makeState(map[string]int64{"x": 0})
	tr := makeTransition(s0, s1, "noop")
	v := makeViolation("inv", []*explore.Transition{tr}, s1)
	out := FormatViolationString(v, DisplayOptions{})
	if !strings.Contains(out, "no state change") {
		t.Errorf("expected 'no state change' for noop:\n%s", out)
	}
}

func TestFormatViolation_ReplayScript(t *testing.T) {
	s0 := makeState(map[string]int64{"x": 0})
	s1 := makeState(map[string]int64{"x": 1})
	tr := makeTransition(s0, s1, "incr")
	v := makeViolation("small", []*explore.Transition{tr}, s1)
	opts := DisplayOptions{ShowReplayScript: true, SpecFile: "counter.spec"}
	out := FormatViolationString(v, opts)
	if !strings.Contains(out, "spectre repl counter.spec") {
		t.Errorf("expected replay script header:\n%s", out)
	}
	if !strings.Contains(out, "incr") {
		t.Errorf("expected action in replay script:\n%s", out)
	}
}

func TestFormatViolation_ShowFullStates(t *testing.T) {
	s0 := makeState(map[string]int64{"x": 0, "y": 7})
	s1 := makeState(map[string]int64{"x": 1, "y": 7})
	tr := makeTransition(s0, s1, "incr")
	v := makeViolation("small", []*explore.Transition{tr}, s1)
	opts := DisplayOptions{ShowFullStates: true}
	out := FormatViolationString(v, opts)
	if !strings.Contains(out, "y = 7") {
		t.Errorf("expected full state with y=7:\n%s", out)
	}
}

func TestFormatViolation_ViolationMarker(t *testing.T) {
	s0 := makeState(map[string]int64{"x": 0})
	s1 := makeState(map[string]int64{"x": 10})
	tr := makeTransition(s0, s1, "big_jump")
	v := makeViolation("bounded", []*explore.Transition{tr}, s1)
	out := FormatViolationString(v, DisplayOptions{})
	if !strings.Contains(out, "VIOLATION") {
		t.Errorf("expected VIOLATION marker:\n%s", out)
	}
}

// ── FormatStateDiff ───────────────────────────────────────────────────────────

func TestFormatStateDiff_Empty(t *testing.T) {
	s := FormatStateDiff(nil)
	if s != "(unchanged)" {
		t.Errorf("expected '(unchanged)', got %q", s)
	}
}

func TestFormatStateDiff_OneChange(t *testing.T) {
	s1 := makeState(map[string]int64{"x": 3})
	s2 := makeState(map[string]int64{"x": 7})
	diffs := StateDiff(s1, s2)
	s := FormatStateDiff(diffs)
	if !strings.Contains(s, "x") {
		t.Errorf("expected x in diff string: %q", s)
	}
}

// ── PrintCounterexampleSummary ────────────────────────────────────────────────

func TestPrintCounterexampleSummary(t *testing.T) {
	s0 := makeState(map[string]int64{"x": 0})
	s1 := makeState(map[string]int64{"x": 1})
	v := makeViolation("inv", []*explore.Transition{makeTransition(s0, s1, "step")}, s1)
	var sb strings.Builder
	PrintCounterexampleSummary([]*explore.Violation{v}, &sb)
	if !strings.Contains(sb.String(), "inv") {
		t.Errorf("expected invariant name in summary: %q", sb.String())
	}
}
