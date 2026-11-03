// Package monitor validates that BFS violation traces are consistent with what a
// generated monitor would observe at runtime.
//
// The BFS explorer records violations in two forms:
//
//  1. State-level: the current state itself fails an invariant.
//     violation.State is the failing state; path leads to it.
//
//  2. Transition-level: from the current state, executing action A produces a
//     state that fails an invariant.  violation.State is the *pre-violation*
//     state; the actual failing state is one step further.
//
// ValidateTrace handles both cases.  It uses sm.ReplayForViolation (which
// bypasses the state machine's invariant guard so it can produce violating
// states) and, when needed, tries one extra step from the final state to
// reproduce case-2 violations.
package monitor

import (
	"fmt"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
)

// StepCheck records which invariants passed or failed after a single step.
type StepCheck struct {
	StepIndex int
	Action    string
	Passing   []string // invariant names that held after this step (informational)
	Failing   []string // invariant names that failed after this step
}

// ValidationResult summarises whether a violation trace is consistent.
type ValidationResult struct {
	// Consistent is true when the trace faithfully reproduces the violation:
	// all intermediate steps pass every invariant and the invariant fails at
	// the final (or extended) step.
	Consistent      bool
	Steps           int
	TargetInvariant string
	FinalViolated   bool   // target invariant failed at the last step (expected)
	EarlyViolation  bool   // an invariant failed before the last step (unexpected)
	StepChecks      []StepCheck
	Note            string
}

// ValidateTrace replays violation.Path through sm and checks invariants at
// every step.  It returns a ValidationResult describing how well the trace
// matches what a generated monitor would observe.
func ValidateTrace(sm *exec.StateMachine, violation *explore.Violation) *ValidationResult {
	r := &ValidationResult{
		TargetInvariant: violation.Invariant,
	}

	path := violation.Path
	r.Steps = len(path)

	if r.Steps == 0 {
		// Violation at the initial state itself.
		errs, _ := sm.ValidateState(violation.State)
		sc := StepCheck{StepIndex: 0, Action: "(initial)"}
		for _, e := range errs {
			sc.Failing = append(sc.Failing, e.Name)
		}
		r.StepChecks = append(r.StepChecks, sc)
		r.FinalViolated = containsName(sc.Failing, violation.Invariant)
		r.Consistent = r.FinalViolated
		if r.Consistent {
			r.Note = fmt.Sprintf("initial state already violates %q", violation.Invariant)
		} else {
			r.Note = "empty path but invariant not violated in the reported state"
		}
		return r
	}

	// Build StepChecks from the recorded path states so we can report
	// per-step invariant status.
	for i, tr := range path {
		errs, _ := sm.ValidateState(tr.ToState)
		sc := StepCheck{StepIndex: i, Action: tr.Action}
		for _, e := range errs {
			sc.Failing = append(sc.Failing, e.Name)
		}
		isLast := i == len(path)-1
		if !isLast && len(sc.Failing) > 0 {
			r.EarlyViolation = true
		}
		r.StepChecks = append(r.StepChecks, sc)
	}

	// Use ReplayForViolation to confirm the violation is reproducible.
	// This uses the executor directly (bypasses the state machine's invariant
	// guard) so it can produce invariant-violating next states.
	initState := path[0].FromState
	steps := make([]exec.ActionStep, len(path))
	for i, tr := range path {
		steps[i] = exec.ActionStep{Name: tr.Action, Args: tr.Args}
	}

	found, finalState := sm.ReplayForViolation(initState, steps, violation.Invariant)
	if found {
		// Case 1: violation appears within the given path steps.
		r.FinalViolated = true
		// Update the last StepCheck to reflect the failing state if it
		// wasn't already populated (ReplayForViolation may have found it
		// via the executor's state, not tr.ToState).
		if len(r.StepChecks) > 0 {
			last := &r.StepChecks[len(r.StepChecks)-1]
			if !containsName(last.Failing, violation.Invariant) {
				last.Failing = append(last.Failing, violation.Invariant)
			}
		}
	} else {
		// Case 2: violation.State is the pre-violation state.  Try one extra
		// action from the final state to reproduce the violation.
		available, _ := sm.GetAvailableActionsWithArgs(finalState)
		for _, a := range available {
			extra := []exec.ActionStep{{Name: a.ActionName, Args: a.Args}}
			if ok, _ := sm.ReplayForViolation(finalState, extra, violation.Invariant); ok {
				r.FinalViolated = true
				// Append an extra StepCheck for the extending step.
				r.StepChecks = append(r.StepChecks, StepCheck{
					StepIndex: len(path),
					Action:    a.ActionName,
					Failing:   []string{violation.Invariant},
				})
				break
			}
		}
	}

	r.Consistent = r.FinalViolated && !r.EarlyViolation
	switch {
	case r.Consistent:
		r.Note = fmt.Sprintf("trace consistent: %q violated as expected", violation.Invariant)
	case r.EarlyViolation:
		r.Note = fmt.Sprintf("early invariant failure before expected step — trace may be stale")
	case !r.FinalViolated:
		r.Note = fmt.Sprintf("%q not reproducible from given path — spec or trace may have diverged", violation.Invariant)
	}
	return r
}

// ValidateAll validates every violation in a slice and returns one result each.
func ValidateAll(sm *exec.StateMachine, violations []*explore.Violation) []*ValidationResult {
	results := make([]*ValidationResult, len(violations))
	for i, v := range violations {
		results[i] = ValidateTrace(sm, v)
	}
	return results
}

// Summary returns a multi-line human-readable report for a slice of results.
func Summary(results []*ValidationResult) string {
	var b strings.Builder
	total := len(results)
	consistent := 0
	for _, r := range results {
		if r.Consistent {
			consistent++
		}
	}
	fmt.Fprintf(&b, "Monitor trace validation: %d/%d consistent\n", consistent, total)
	for i, r := range results {
		mark := "✓"
		if !r.Consistent {
			mark = "✗"
		}
		fmt.Fprintf(&b, "  [%s] violation %d (%s, %d steps): %s\n",
			mark, i+1, r.TargetInvariant, r.Steps, r.Note)
	}
	return b.String()
}

func containsName(names []string, target string) bool {
	for _, n := range names {
		if n == target {
			return true
		}
	}
	return false
}
