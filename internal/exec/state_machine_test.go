package exec

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/state"
)

func TestStateMachineCompleteFlow(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

invariant counterNonNegative {
  counter >= 0
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	sm, err := NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	// Get initial states
	initialStates, err := sm.GetInitialStates()
	if err != nil {
		t.Fatalf("error getting initial states: %v", err)
	}

	if len(initialStates) != 1 {
		t.Fatalf("expected 1 initial state, got %d", len(initialStates))
	}

	currentState := initialStates[0]

	// Validate initial state
	errors, err := sm.ValidateState(currentState)
	if err != nil {
		t.Fatalf("error validating initial state: %v", err)
	}

	if len(errors) > 0 {
		t.Errorf("initial state should satisfy invariants, got errors: %v", errors)
	}

	// Execute increment action
	nextState, err := sm.ExecuteAction("increment", currentState, nil)
	if err != nil {
		t.Fatalf("error executing action: %v", err)
	}

	// Check that counter was incremented
	counterVal, _ := nextState.GetVariable("counter")
	if pv, ok := counterVal.(*state.PrimitiveValue); ok && pv.IntValue != nil {
		if *pv.IntValue != 1 {
			t.Errorf("expected counter = 1, got %d", *pv.IntValue)
		}
	}

	// Get available actions
	available, err := sm.GetAvailableActions(nextState)
	if err != nil {
		t.Fatalf("error getting available actions: %v", err)
	}

	if len(available) == 0 {
		t.Error("expected at least one available action")
	}
}

func TestStateMachineWithGuard(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment when counter < 10 {
  counter' = counter + 1
}

invariant counterNonNegative {
  counter >= 0
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	sm, err := NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	initialStates, _ := sm.GetInitialStates()
	currentState := initialStates[0]

	// Execute increment multiple times until guard fails
	for i := 0; i < 12; i++ {
		available, _ := sm.GetAvailableActions(currentState)
		if len(available) == 0 {
			if i < 10 {
				t.Errorf("action should be available when counter < 10, counter = %d", i)
			}
			break
		}

		nextState, err := sm.ExecuteAction("increment", currentState, nil)
		if err != nil {
			if i >= 10 {
				// Expected - guard should prevent execution
				break
			}
			t.Fatalf("unexpected error executing action: %v", err)
		}

		currentState = nextState
	}
}

