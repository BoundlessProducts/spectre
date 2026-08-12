package explore

import (
	"fmt"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/state"
)

// Note: strings package is used for string building

// Counterexample represents a counterexample trace showing how a violation was reached
type Counterexample struct {
	Violation   *Violation
	Trace       []*Transition
	Description string
}

// FormatCounterexample formats a counterexample as a readable string
func FormatCounterexample(ce *Counterexample) string {
	var sb strings.Builder

	sb.WriteString("Counterexample:\n")
	sb.WriteString(fmt.Sprintf("  Violation: %s\n", ce.Violation.Invariant))
	sb.WriteString(fmt.Sprintf("  Description: %s\n", ce.Violation.Description))
	sb.WriteString("\n  Trace:\n")

	if len(ce.Trace) == 0 {
		sb.WriteString("    (initial state)\n")
	} else {
		for i, transition := range ce.Trace {
			sb.WriteString(fmt.Sprintf("    Step %d: %s\n", i+1, formatTransition(transition)))
		}
	}

	sb.WriteString(fmt.Sprintf("\n  Final State:\n"))
	sb.WriteString(formatState(ce.Violation.State))

	return sb.String()
}

// formatTransition formats a transition for display
func formatTransition(t *Transition) string {
	argsStr := ""
	if len(t.Args) > 0 {
		argStrs := make([]string, len(t.Args))
		for i, arg := range t.Args {
			argStrs[i] = arg.String()
		}
		argsStr = fmt.Sprintf("(%s)", strings.Join(argStrs, ", "))
	}
	return fmt.Sprintf("%s%s", t.Action, argsStr)
}

// formatState formats a state for display
func formatState(s *state.State) string {
	var sb strings.Builder
	sb.WriteString("    {\n")
	for varName, varValue := range s.Variables {
		sb.WriteString(fmt.Sprintf("      %s = %s\n", varName, varValue.String()))
	}
	sb.WriteString("    }\n")
	return sb.String()
}

// BuildCounterexample builds a counterexample from a violation
func BuildCounterexample(violation *Violation) *Counterexample {
	return &Counterexample{
		Violation:   violation,
		Trace:       violation.Path,
		Description: violation.Description,
	}
}

// FormatViolations formats all violations as counterexamples
func FormatViolations(violations []*Violation) []string {
	formatted := make([]string, len(violations))
	for i, violation := range violations {
		ce := BuildCounterexample(violation)
		formatted[i] = FormatCounterexample(ce)
	}
	return formatted
}

