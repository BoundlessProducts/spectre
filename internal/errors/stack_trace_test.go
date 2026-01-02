package errors

import (
	"strings"
	"testing"

	"github.com/akkeshavan/spectre/internal/explore"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/internal/temporal"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestStackTraceBasic(t *testing.T) {
	st := NewStackTrace()

	pos1 := ast.Position{Line: 5, Column: 3}
	context1 := NewErrorContext(pos1, "Initial state", "", "initial state")
	st.AddEntry(context1, "Starting execution")

	pos2 := ast.Position{Line: 10, Column: 8}
	context2 := NewErrorContext(pos2, "Increments counter", "increment", "action")
	st.AddEntry(context2, "Executing action")

	pos3 := ast.Position{Line: 15, Column: 12}
	context3 := NewErrorContext(pos3, "Counter must be non-negative", "counterNonNegative", "invariant")
	st.AddEntry(context3, "Invariant violated")

	formatted := st.Format()

	if !strings.Contains(formatted, "Stack trace:") {
		t.Error("formatted stack trace should contain 'Stack trace:'")
	}

	if !strings.Contains(formatted, "increment") {
		t.Error("formatted stack trace should contain action name")
	}

	if !strings.Contains(formatted, "Counter must be non-negative") {
		t.Error("formatted stack trace should contain description")
	}

	if !strings.Contains(formatted, "Invariant violated") {
		t.Error("formatted stack trace should contain violation message")
	}
}

func TestStackTraceFromTrace(t *testing.T) {
	// Create a trace
	trace := temporal.NewTrace()
	
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(1))
	trace.AddState(s1, "increment", nil)

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(-1))
	trace.AddState(s2, "decrement", nil)

	// Create violation context
	pos := ast.Position{Line: 20, Column: 10}
	context := NewErrorContext(pos, "Counter must be non-negative", "counterNonNegative", "invariant")

	// Build stack trace
	st := BuildStackTraceFromTrace(trace, context, "Invariant violated: counter < 0")

	formatted := st.Format()

	if !strings.Contains(formatted, "Stack trace:") {
		t.Error("formatted stack trace should contain 'Stack trace:'")
	}

	if !strings.Contains(formatted, "increment") {
		t.Error("formatted stack trace should contain increment action")
	}

	if !strings.Contains(formatted, "decrement") {
		t.Error("formatted stack trace should contain decrement action")
	}

	if !strings.Contains(formatted, "Invariant violated") {
		t.Error("formatted stack trace should contain violation message")
	}
}

func TestStackTraceFromViolationPath(t *testing.T) {
	// Create violation path
	path := []*explore.Transition{
		{
			FromState: state.NewState(),
			ToState:   state.NewState(),
			Action:    "increment",
			Args:      nil,
		},
		{
			FromState: state.NewState(),
			ToState:   state.NewState(),
			Action:    "decrement",
			Args:      nil,
		},
	}

	// Create violation context
	pos := ast.Position{Line: 25, Column: 15}
	context := NewErrorContext(pos, "Counter must be non-negative", "counterNonNegative", "invariant")

	// Build stack trace
	st := BuildStackTraceFromViolationPath(path, context, "Invariant violated: counter < 0")

	formatted := st.Format()

	if !strings.Contains(formatted, "Stack trace:") {
		t.Error("formatted stack trace should contain 'Stack trace:'")
	}

	if !strings.Contains(formatted, "increment") {
		t.Error("formatted stack trace should contain increment action")
	}

	if !strings.Contains(formatted, "decrement") {
		t.Error("formatted stack trace should contain decrement action")
	}

	if !strings.Contains(formatted, "Transition 1") {
		t.Error("formatted stack trace should contain transition numbers")
	}
}

func TestStackTraceEmpty(t *testing.T) {
	st := NewStackTrace()
	formatted := st.Format()

	if formatted != "No stack trace available" {
		t.Errorf("expected 'No stack trace available', got '%s'", formatted)
	}
}

