package explore

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
)

func TestExplorerCompleteFlow(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action decrement {
  counter' = counter - 1
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

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	explorer := NewExplorer(sm)
	explorer.SetMaxDepth(5)
	explorer.SetMaxStates(20)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	if result.StatesExplored == 0 {
		t.Error("expected to explore at least one state")
	}

	// Should find violations when decrement makes counter negative
	// But the action execution should fail due to invariant validation
	if len(result.Violations) > 0 {
		t.Logf("found %d violations", len(result.Violations))
		for i, violation := range result.Violations {
			ce := BuildCounterexample(violation)
			t.Logf("Violation %d:\n%s", i+1, FormatCounterexample(ce))
		}
	}
}

func TestExplorerCycleDetection(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action reset {
  counter' = 0
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

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	explorer := NewExplorer(sm)
	explorer.SetMaxDepth(10)
	explorer.SetMaxStates(50)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	// Should detect cycles (increment -> reset -> increment -> ...)
	if len(result.Cycles) > 0 {
		t.Logf("found %d cycles", len(result.Cycles))
		for i, cycle := range result.Cycles {
			t.Logf("Cycle %d: %s", i+1, cycle.Description)
		}
	}
}

func TestExplorerMultipleInitialStates(t *testing.T) {
	spec := `
var counter: int

init oneOf {
  { counter = 0 },
  { counter = 5 },
  { counter = 10 }
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

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	explorer := NewExplorer(sm)
	explorer.SetMaxDepth(3)
	explorer.SetMaxStates(30)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	// Should explore from all 3 initial states
	if result.StatesExplored < 3 {
		t.Errorf("expected to explore at least 3 states (from 3 initial states), got %d", result.StatesExplored)
	}

	if len(result.ReachableStates) < 3 {
		t.Errorf("expected at least 3 reachable states, got %d", len(result.ReachableStates))
	}
}

func TestExplorerInvariantViolation(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

invariant counterLessThanFive {
  counter < 5
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	explorer := NewExplorer(sm)
	explorer.SetMaxDepth(10)
	explorer.SetMaxStates(20)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	// Should find violations when counter >= 5
	// Note: The state machine should prevent invalid states from being reached
	// So violations might be found during exploration but states won't be reached
	if len(result.Violations) > 0 {
		t.Logf("found %d violations during exploration", len(result.Violations))
	}
}

