package explore

import (
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/state"
)

func TestFormatCounterexample(t *testing.T) {
	// Create a violation with a trace
	violation := &Violation{
		State:     state.NewState(),
		Invariant: "counterNonNegative",
		Path: []*Transition{
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
		},
		Description: "counter < 0",
	}

	violation.State.SetVariable("counter", state.NewIntValue(-1))

	ce := BuildCounterexample(violation)
	formatted := FormatCounterexample(ce)

	if !strings.Contains(formatted, "counterNonNegative") {
		t.Error("formatted counterexample should contain invariant name")
	}

	if !strings.Contains(formatted, "increment") {
		t.Error("formatted counterexample should contain action names")
	}

	if !strings.Contains(formatted, "decrement") {
		t.Error("formatted counterexample should contain all actions")
	}
}

func TestFormatCounterexampleNoTrace(t *testing.T) {
	violation := &Violation{
		State:       state.NewState(),
		Invariant:   "counterNonNegative",
		Path:        []*Transition{},
		Description: "counter < 0",
	}

	violation.State.SetVariable("counter", state.NewIntValue(-1))

	ce := BuildCounterexample(violation)
	formatted := FormatCounterexample(ce)

	if !strings.Contains(formatted, "(initial state)") {
		t.Error("formatted counterexample should indicate initial state when no trace")
	}
}

func TestFormatViolations(t *testing.T) {
	violations := []*Violation{
		{
			State:       state.NewState(),
			Invariant:   "invariant1",
			Path:        []*Transition{},
			Description: "violation 1",
		},
		{
			State:       state.NewState(),
			Invariant:   "invariant2",
			Path:        []*Transition{},
			Description: "violation 2",
		},
	}

	formatted := FormatViolations(violations)

	if len(formatted) != 2 {
		t.Errorf("expected 2 formatted violations, got %d", len(formatted))
	}

	if !strings.Contains(formatted[0], "invariant1") {
		t.Error("first violation should contain invariant1")
	}

	if !strings.Contains(formatted[1], "invariant2") {
		t.Error("second violation should contain invariant2")
	}
}

