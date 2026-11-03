package diagnose

// counterexample_display.go — Item 16: rich counterexample formatting with
// state diffs, failed-predicate annotation, and replay-script generation.

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/state"
)

// VarChange records a single variable that changed between two states.
type VarChange struct {
	Name   string
	Before state.Value
	After  state.Value
}

// StateDiff returns the sorted list of variables that differ between before and after.
// Variables present in one state but not the other are included with a nil value on
// the missing side.
func StateDiff(before, after *state.State) []VarChange {
	seen := map[string]bool{}
	var changes []VarChange

	for k, bv := range before.Variables {
		seen[k] = true
		av, ok := after.Variables[k]
		if !ok || av.String() != bv.String() {
			after_val := av
			if !ok {
				after_val = nil
			}
			changes = append(changes, VarChange{Name: k, Before: bv, After: after_val})
		}
	}
	for k, av := range after.Variables {
		if !seen[k] {
			changes = append(changes, VarChange{Name: k, Before: nil, After: av})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Name < changes[j].Name })
	return changes
}

// DisplayOptions controls the detail level of FormatViolation.
type DisplayOptions struct {
	// ShowFullStates prints every variable at every step (default: only diffs).
	ShowFullStates bool
	// Indent string prepended to every output line.
	Indent string
	// ShowReplayScript appends a "spectre repl" replay script at the end.
	ShowReplayScript bool
	// SpecFile is the spec filename used in the replay script.
	SpecFile string
}

// FormatViolation writes a human-readable counterexample trace for v to w.
// Each step shows the action taken, the variables that changed (diff), and
// the failed predicate at the violating state.
func FormatViolation(v *explore.Violation, opts DisplayOptions, w io.Writer) {
	ind := opts.Indent

	fmt.Fprintf(w, "%sInvariant violated: %s\n", ind, v.Invariant)
	if v.Description != "" && v.Description != v.Invariant {
		fmt.Fprintf(w, "%s%s\n", ind, v.Description)
	}

	if len(v.Path) == 0 {
		fmt.Fprintf(w, "%s(violation in initial state — no steps needed)\n", ind)
		if v.State != nil {
			printStateVars(w, ind+"  ", v.State)
		}
		return
	}

	fmt.Fprintf(w, "%sCounterexample trace (%d step(s)):\n", ind, len(v.Path))

	// Print initial state (FromState of first transition).
	if v.Path[0].FromState != nil {
		fmt.Fprintf(w, "%s  [init]\n", ind)
		if opts.ShowFullStates {
			printStateVars(w, ind+"    ", v.Path[0].FromState)
		}
	}

	for i, tr := range v.Path {
		actionLabel := tr.Action
		if len(tr.Args) > 0 {
			argStrs := make([]string, len(tr.Args))
			for j, a := range tr.Args {
				argStrs[j] = a.String()
			}
			actionLabel = fmt.Sprintf("%s(%s)", tr.Action, strings.Join(argStrs, ", "))
		}
		fmt.Fprintf(w, "%s  [step %d] %s\n", ind, i+1, actionLabel)

		if tr.FromState != nil && tr.ToState != nil {
			diffs := StateDiff(tr.FromState, tr.ToState)
			if len(diffs) == 0 {
				fmt.Fprintf(w, "%s    (no state change)\n", ind)
			} else {
				for _, d := range diffs {
					before := "<absent>"
					after := "<absent>"
					if d.Before != nil {
						before = d.Before.String()
					}
					if d.After != nil {
						after = d.After.String()
					}
					fmt.Fprintf(w, "%s    %s: %s → %s\n", ind, d.Name, before, after)
				}
			}
			if opts.ShowFullStates {
				fmt.Fprintf(w, "%s    Full state:\n", ind)
				printStateVars(w, ind+"      ", tr.ToState)
			}
		}
	}

	// Final (violating) state.
	fmt.Fprintf(w, "%s  [VIOLATION] invariant %q fails in this state\n", ind, v.Invariant)
	if v.State != nil {
		printStateVars(w, ind+"    ", v.State)
	}

	if opts.ShowReplayScript && opts.SpecFile != "" {
		fmt.Fprintf(w, "\n%sReplay script (paste into 'spectre repl %s'):\n", ind, opts.SpecFile)
		for _, tr := range v.Path {
			fmt.Fprintf(w, "%s  %s\n", ind, tr.Action)
		}
	}
}

// FormatViolationString is a convenience wrapper around FormatViolation that
// returns the formatted string.
func FormatViolationString(v *explore.Violation, opts DisplayOptions) string {
	var sb strings.Builder
	FormatViolation(v, opts, &sb)
	return sb.String()
}

// PrintCounterexampleSummary writes a one-line summary of each violation.
func PrintCounterexampleSummary(violations []*explore.Violation, w io.Writer) {
	for i, v := range violations {
		steps := len(v.Path)
		fmt.Fprintf(w, "  %d. %s (%d step(s))\n", i+1, v.Invariant, steps)
	}
}

// FormatStateDiff formats a StateDiff result as a compact string for inline use.
func FormatStateDiff(diffs []VarChange) string {
	if len(diffs) == 0 {
		return "(unchanged)"
	}
	parts := make([]string, len(diffs))
	for i, d := range diffs {
		before, after := "<absent>", "<absent>"
		if d.Before != nil {
			before = d.Before.String()
		}
		if d.After != nil {
			after = d.After.String()
		}
		parts[i] = fmt.Sprintf("%s: %s→%s", d.Name, before, after)
	}
	return strings.Join(parts, ", ")
}

// ── helpers ───────────────────────────────────────────────────────────────────

func printStateVars(w io.Writer, indent string, s *state.State) {
	names := make([]string, 0, len(s.Variables))
	for k := range s.Variables {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		fmt.Fprintf(w, "%s%s = %s\n", indent, k, s.Variables[k].String())
	}
}
