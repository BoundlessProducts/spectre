package errors

import (
	"fmt"
	"strings"

	"github.com/spectre-lang/spectre/internal/explore"
	"github.com/spectre-lang/spectre/internal/temporal"
	"github.com/spectre-lang/spectre/pkg/ast"
)

// StackTraceEntry represents a single entry in a stack trace
type StackTraceEntry struct {
	Position    ast.Position
	Description string
	ElementName string
	ElementType string
	Message     string
}

// StackTrace represents a complete stack trace showing the path to an error
type StackTrace struct {
	Entries []*StackTraceEntry
}

// NewStackTrace creates a new empty stack trace
func NewStackTrace() *StackTrace {
	return &StackTrace{
		Entries: []*StackTraceEntry{},
	}
}

// AddEntry adds an entry to the stack trace
func (st *StackTrace) AddEntry(context *ErrorContext, message string) {
	st.Entries = append(st.Entries, &StackTraceEntry{
		Position:    context.Position,
		Description: context.Description,
		ElementName: context.ElementName,
		ElementType: context.ElementType,
		Message:     message,
	})
}

// Format formats the stack trace as a readable string
func (st *StackTrace) Format() string {
	if len(st.Entries) == 0 {
		return "No stack trace available"
	}

	var lines []string
	lines = append(lines, "Stack trace:")

	for i, entry := range st.Entries {
		var parts []string

		// Add entry number
		parts = append(parts, fmt.Sprintf("  %d.", i+1))

		// Add position
		if entry.Position.Line > 0 {
			parts = append(parts, fmt.Sprintf("%d:%d", entry.Position.Line, entry.Position.Column))
		}

		// Add element information
		if entry.ElementName != "" || entry.ElementType != "" {
			if entry.ElementName != "" {
				parts = append(parts, fmt.Sprintf("%s '%s'", entry.ElementType, entry.ElementName))
			} else {
				parts = append(parts, entry.ElementType)
			}
		}

		// Add description if available
		if entry.Description != "" {
			parts = append(parts, fmt.Sprintf("(%s)", entry.Description))
		}

		// Add message
		if entry.Message != "" {
			parts = append(parts, ":", entry.Message)
		}

		lines = append(lines, strings.Join(parts, " "))
	}

	return strings.Join(lines, "\n")
}

// BuildStackTraceFromTrace builds a stack trace from an execution trace
func BuildStackTraceFromTrace(trace *temporal.Trace, violationContext *ErrorContext, violationMessage string) *StackTrace {
	st := NewStackTrace()

	// Add entries for each state transition in the trace
	for i := 0; i < trace.Length(); i++ {
		state := trace.GetState(i)
		action := trace.GetAction(i)

		// Create context for this trace entry
		context := NewErrorContext(
			ast.Position{}, // Position not available from trace
			"",
			action,
			"action",
		)

		message := fmt.Sprintf("State %d: %s", i, formatStateBrief(state))
		st.AddEntry(context, message)
	}

	// Add the violation entry at the end
	st.AddEntry(violationContext, violationMessage)

	return st
}

// formatStateBrief formats a brief description of a state
func formatStateBrief(state interface{}) string {
	// This is a placeholder - in a real implementation, we'd format the state values
	return "state values"
}

// BuildStackTraceFromViolationPath builds a stack trace from a violation path
func BuildStackTraceFromViolationPath(path []*explore.Transition, violationContext *ErrorContext, violationMessage string) *StackTrace {
	st := NewStackTrace()

	// Add entries for each transition in the path
	for i, transition := range path {
		context := NewErrorContext(
			ast.Position{}, // Position not available from transition
			"",
			transition.Action,
			"action",
		)

		message := fmt.Sprintf("Transition %d: %s", i+1, transition.Action)
		if len(transition.Args) > 0 {
			message += fmt.Sprintf(" with args")
		}
		st.AddEntry(context, message)
	}

	// Add the violation entry at the end
	st.AddEntry(violationContext, violationMessage)

	return st
}

