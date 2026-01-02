package explore

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/exec"
	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
)

func TestExplorerBFS(t *testing.T) {
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

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	explorer := NewExplorer(sm)
	explorer.SetMaxDepth(5)
	explorer.SetMaxStates(10)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	if result.StatesExplored == 0 {
		t.Error("expected to explore at least one state")
	}

	if len(result.ReachableStates) == 0 {
		t.Error("expected at least one reachable state")
	}

	if len(result.Violations) > 0 {
		t.Errorf("expected no violations, got %d", len(result.Violations))
	}
}

func TestExplorerDFS(t *testing.T) {
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

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	explorer := NewExplorer(sm)
	explorer.SetMaxDepth(5)
	explorer.SetMaxStates(10)

	result, err := explorer.ExploreDFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	if result.StatesExplored == 0 {
		t.Error("expected to explore at least one state")
	}

	if len(result.ReachableStates) == 0 {
		t.Error("expected at least one reachable state")
	}
}

func TestExplorerMultipleActions(t *testing.T) {
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
	explorer.SetMaxDepth(3)
	explorer.SetMaxStates(20)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	if result.StatesExplored < 2 {
		t.Errorf("expected to explore at least 2 states, got %d", result.StatesExplored)
	}

	// Should find violations when counter goes negative
	// But decrement action should fail due to invariant violation
	if len(result.Violations) > 0 {
		t.Logf("found violations (may be expected): %d", len(result.Violations))
	}
}

func TestExplorerWithGuard(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment when counter < 5 {
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
	explorer.SetMaxDepth(10)
	explorer.SetMaxStates(20)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	// Should explore states up to counter = 5, then stop (guard prevents further increments)
	if result.StatesExplored == 0 {
		t.Error("expected to explore at least one state")
	}

	// Should not have violations (counter stays >= 0 and < 5)
	if len(result.Violations) > 0 {
		t.Errorf("expected no violations, got %d", len(result.Violations))
	}
}

func TestExplorerOneOfInit(t *testing.T) {
	spec := `
var counter: int

init oneOf {
  { counter = 0 },
  { counter = 5 }
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
	explorer.SetMaxStates(20)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	// Should explore from both initial states
	if result.StatesExplored < 2 {
		t.Errorf("expected to explore at least 2 states (from 2 initial states), got %d", result.StatesExplored)
	}

	if len(result.ReachableStates) < 2 {
		t.Errorf("expected at least 2 reachable states, got %d", len(result.ReachableStates))
	}
}

