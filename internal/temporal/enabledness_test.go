package temporal

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/internal/state"
)

func TestActionEnablednessCheckerBasic(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action decrement when counter > 0 {
  counter' = counter - 1
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

	checker := NewActionEnablednessChecker(sm)

	// Test state where counter = 0
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))

	// increment should be enabled
	enabled, err := checker.IsActionEnabled("increment", s0)
	if err != nil {
		t.Fatalf("error checking enabledness: %v", err)
	}
	if !enabled {
		t.Error("expected increment to be enabled when counter = 0")
	}

	// decrement should not be enabled (counter == 0, not > 0)
	enabled, err = checker.IsActionEnabled("decrement", s0)
	if err != nil {
		t.Fatalf("error checking enabledness: %v", err)
	}
	if enabled {
		t.Error("expected decrement to not be enabled when counter = 0")
	}

	// Test state where counter = 5
	s5 := state.NewState()
	s5.SetVariable("counter", state.NewIntValue(5))

	// Both actions should be enabled
	enabled, err = checker.IsActionEnabled("increment", s5)
	if err != nil {
		t.Fatalf("error checking enabledness: %v", err)
	}
	if !enabled {
		t.Error("expected increment to be enabled when counter = 5")
	}

	enabled, err = checker.IsActionEnabled("decrement", s5)
	if err != nil {
		t.Fatalf("error checking enabledness: %v", err)
	}
	if !enabled {
		t.Error("expected decrement to be enabled when counter = 5")
	}
}

func TestActionEnablednessCheckerInTrace(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action decrement when counter > 0 {
  counter' = counter - 1
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

	checker := NewActionEnablednessChecker(sm)

	// Create trace
	trace := NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(1))
	trace.AddState(s1, "increment", nil)

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(2))
	trace.AddState(s2, "increment", nil)

	// Check enabledness at position 0 (counter = 0)
	enabled, err := checker.IsActionEnabledInTrace("increment", trace, 0)
	if err != nil {
		t.Fatalf("error checking enabledness: %v", err)
	}
	if !enabled {
		t.Error("expected increment to be enabled at position 0")
	}

	enabled, err = checker.IsActionEnabledInTrace("decrement", trace, 0)
	if err != nil {
		t.Fatalf("error checking enabledness: %v", err)
	}
	if enabled {
		t.Error("expected decrement to not be enabled at position 0")
	}

	// Check enabledness at position 1 (counter = 1)
	enabled, err = checker.IsActionEnabledInTrace("decrement", trace, 1)
	if err != nil {
		t.Fatalf("error checking enabledness: %v", err)
	}
	if !enabled {
		t.Error("expected decrement to be enabled at position 1")
	}
}

func TestActionEnablednessCheckerContinuouslyEnabled(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action decrement when counter > 0 {
  counter' = counter - 1
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

	checker := NewActionEnablednessChecker(sm)

	// Create trace where increment is continuously enabled
	trace := NewTrace()
	for i := 0; i <= 5; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		if i == 0 {
			trace.AddState(s, "", nil)
		} else {
			trace.AddState(s, "increment", nil)
		}
	}

	// increment should be continuously enabled from 0 to 5
	continuouslyEnabled, err := checker.IsContinuouslyEnabled("increment", trace, 0, 6)
	if err != nil {
		t.Fatalf("error checking continuous enabledness: %v", err)
	}
	if !continuouslyEnabled {
		t.Error("expected increment to be continuously enabled")
	}

	// decrement should not be continuously enabled (disabled at position 0)
	continuouslyEnabled, err = checker.IsContinuouslyEnabled("decrement", trace, 0, 6)
	if err != nil {
		t.Fatalf("error checking continuous enabledness: %v", err)
	}
	if continuouslyEnabled {
		t.Error("expected decrement to not be continuously enabled")
	}

	// decrement should be continuously enabled from position 1 onwards
	continuouslyEnabled, err = checker.IsContinuouslyEnabled("decrement", trace, 1, 6)
	if err != nil {
		t.Fatalf("error checking continuous enabledness: %v", err)
	}
	if !continuouslyEnabled {
		t.Error("expected decrement to be continuously enabled from position 1")
	}
}

func TestActionEnablednessCheckerEnabledPositions(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action decrement when counter > 0 {
  counter' = counter - 1
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

	checker := NewActionEnablednessChecker(sm)

	// Create trace
	trace := NewTrace()
	for i := 0; i <= 3; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		if i == 0 {
			trace.AddState(s, "", nil)
		} else {
			trace.AddState(s, "increment", nil)
		}
	}

	// increment should be enabled at all positions
	positions, err := checker.GetEnabledPositions("increment", trace)
	if err != nil {
		t.Fatalf("error getting enabled positions: %v", err)
	}
	if len(positions) != 4 {
		t.Errorf("expected 4 enabled positions for increment, got %d", len(positions))
	}

	// decrement should be enabled at positions 1, 2, 3 (not 0)
	positions, err = checker.GetEnabledPositions("decrement", trace)
	if err != nil {
		t.Fatalf("error getting enabled positions: %v", err)
	}
	if len(positions) != 3 {
		t.Errorf("expected 3 enabled positions for decrement, got %d", len(positions))
	}
	if len(positions) > 0 && positions[0] != 1 {
		t.Errorf("expected first enabled position to be 1, got %d", positions[0])
	}
}

func TestActionEnablednessCheckerExecutionPositions(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
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

	checker := NewActionEnablednessChecker(sm)

	// Create trace with increment executing
	// trace.States[0] = s0 (initial, no action)
	// trace.Actions[0] = "increment" (action leading to state 1)
	// trace.States[1] = s1
	// trace.Actions[1] = "increment" (action leading to state 2)
	// trace.States[2] = s2
	trace := NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(1))
	trace.AddState(s1, "increment", nil)

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(2))
	trace.AddState(s2, "increment", nil)

	// Check execution positions
	// trace has: States[0]=s0, Actions[0]="", States[1]=s1, Actions[1]="increment", States[2]=s2, Actions[2]="increment"
	// So "increment" executes at positions 1 and 2 in the Actions array
	positions := checker.GetExecutionPositions("increment", trace)
	if len(positions) != 2 {
		t.Errorf("expected 2 execution positions, got %d", len(positions))
	}
	if len(positions) > 0 && positions[0] != 1 {
		t.Errorf("expected first execution position to be 1, got %d", positions[0])
	}
	if len(positions) > 1 && positions[1] != 2 {
		t.Errorf("expected second execution position to be 2, got %d", positions[1])
	}

	// Check execution count
	count := checker.CountExecutionOccurrences("increment", trace)
	if count != 2 {
		t.Errorf("expected 2 execution occurrences, got %d", count)
	}
}

