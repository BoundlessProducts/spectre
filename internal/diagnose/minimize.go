package diagnose

import (
	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/state"
)

// MinimizeCounterexample applies delta-debugging to reduce the action sequence in
// v.Path to a 1-minimal sub-sequence that still triggers the same invariant
// violation when replayed from each initial state.
//
// "1-minimal" means: no single action can be removed from the result while
// preserving the violation.  The returned Violation has a shortened Path; all
// other fields are copied from v unchanged.
//
// If the violation cannot be reproduced by any sub-sequence (e.g., the model
// uses non-determinism that differs per run), the original is returned unchanged.
func MinimizeCounterexample(v *explore.Violation, sm *exec.StateMachine) *explore.Violation {
	if len(v.Path) <= 1 {
		return v
	}

	steps := pathToSteps(v.Path)
	initialStates, err := sm.GetInitialStates()
	if err != nil || len(initialStates) == 0 {
		return v
	}

	targetInvariant := v.Invariant

	// Prefer the exact initial state the violation was found from, so that
	// delta-debugging doesn't confirm a sub-sequence via a different execution path.
	var replayInit *state.State
	if len(v.Path) > 0 && v.Path[0].FromState != nil {
		replayInit = v.Path[0].FromState
	}

	// check returns true if the given sub-sequence still violates targetInvariant.
	check := func(sub []exec.ActionStep) bool {
		if replayInit != nil {
			violated, _ := sm.ReplayForViolation(replayInit, sub, targetInvariant)
			return violated
		}
		for _, init := range initialStates {
			if violated, _ := sm.ReplayForViolation(init, sub, targetInvariant); violated {
				return true
			}
		}
		return false
	}

	// Only proceed if the original trace is reproducible.
	if !check(steps) {
		return v
	}

	minimal := deltaDebug(steps, check)

	if len(minimal) >= len(steps) {
		return v
	}

	// Build a new Violation with the minimized path.
	min := *v
	min.Path = stepsToPath(minimal, v.Path)
	return &min
}

// deltaDebug implements Zeller's delta-debugging algorithm (ddmin) adapted for
// action traces, yielding a 1-minimal sub-sequence.
func deltaDebug(steps []exec.ActionStep, check func([]exec.ActionStep) bool) []exec.ActionStep {
	n := 2
	for len(steps) >= 2 {
		chunkSize := len(steps) / n
		if chunkSize == 0 {
			break
		}

		reduced := false
		for i := 0; i < n; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if i == n-1 {
				end = len(steps)
			}
			complement := exclude(steps, start, end)
			if len(complement) > 0 && check(complement) {
				steps = complement
				n = max2(n-1, 2)
				reduced = true
				break
			}
		}

		if !reduced {
			if n == len(steps) {
				break
			}
			n = min2(n*2, len(steps))
		}
	}
	return steps
}

// exclude returns steps with indices [start, end) removed.
func exclude(steps []exec.ActionStep, start, end int) []exec.ActionStep {
	result := make([]exec.ActionStep, 0, len(steps)-(end-start))
	result = append(result, steps[:start]...)
	result = append(result, steps[end:]...)
	return result
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// pathToSteps converts an explore.Transition path to a flat exec.ActionStep slice.
func pathToSteps(path []*explore.Transition) []exec.ActionStep {
	steps := make([]exec.ActionStep, 0, len(path))
	for _, t := range path {
		if t.Action != "" {
			steps = append(steps, exec.ActionStep{Name: t.Action, Args: t.Args})
		}
	}
	return steps
}

// stepsToPath reconstructs a minimal Transition slice from exec.ActionStep.
// It matches steps to original transitions by scanning the original path in order,
// which correctly handles repeated actions after delta-debugging removes some occurrences.
func stepsToPath(steps []exec.ActionStep, orig []*explore.Transition) []*explore.Transition {
	result := make([]*explore.Transition, 0, len(steps))
	origIdx := 0
	for _, s := range steps {
		// Find the next transition in the original path with a matching action name.
		for origIdx < len(orig) && orig[origIdx].Action != s.Name {
			origIdx++
		}
		if origIdx < len(orig) {
			result = append(result, orig[origIdx])
			origIdx++
		} else {
			// No matching original transition left; emit a stub.
			result = append(result, &explore.Transition{
				Action: s.Name,
				Args:   s.Args,
			})
		}
	}
	return result
}
