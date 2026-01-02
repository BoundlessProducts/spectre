package temporal

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/explore"
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestTemporalPropertyVerification(t *testing.T) {
	// Test spec with temporal property
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

temporal eventuallyReachesFive {
  eventually (counter == 5)
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Extract temporal declaration
	var temporalDecl *ast.TemporalDecl
	for _, decl := range file.Decls {
		if td, ok := decl.(*ast.TemporalDecl); ok {
			temporalDecl = td
			break
		}
	}

	if temporalDecl == nil {
		t.Fatal("expected temporal declaration")
	}

	// Create state machine
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	// Explore state space to generate trace
	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(10)
	explorer.SetMaxStates(20)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	// Find a trace that reaches counter == 5
	var trace *Trace
	for _, reachableState := range result.ReachableStates {
		counterVal, exists := reachableState.GetVariable("counter")
		if exists && counterVal != nil {
			if pv, ok := counterVal.(*state.PrimitiveValue); ok && pv.IntValue != nil {
				if *pv.IntValue == 5 {
					// Build trace to this state
					trace = NewTrace()
					// Add initial state
					initialStates, _ := sm.GetInitialStates()
					if len(initialStates) > 0 {
						trace.AddState(initialStates[0], "", nil)
					}
					// Add intermediate states
					for i := 1; i <= 5; i++ {
						s := state.NewState()
						s.SetVariable("counter", state.NewIntValue(int64(i)))
						trace.AddState(s, "increment", nil)
					}
					break
				}
			}
		}
	}

	if trace == nil {
		// Create a simple trace manually
		trace = NewTrace()
		s0 := state.NewState()
		s0.SetVariable("counter", state.NewIntValue(0))
		trace.AddState(s0, "", nil)

		for i := 1; i <= 5; i++ {
			s := state.NewState()
			s.SetVariable("counter", state.NewIntValue(int64(i)))
			trace.AddState(s, "increment", nil)
		}
	}

	// Evaluate temporal property
	evaluator := NewTemporalEvaluator()
	holds, err := evaluator.EvaluateTemporalProperty(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}

	if !holds {
		t.Error("expected temporal property 'eventually (counter == 5)' to hold")
	}
}

func TestTemporalPropertyAlwaysViolation(t *testing.T) {
	// Test spec with temporal property that should be violated
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

temporal alwaysLessThanFive {
  always (counter < 5)
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Extract temporal declaration
	var temporalDecl *ast.TemporalDecl
	for _, decl := range file.Decls {
		if td, ok := decl.(*ast.TemporalDecl); ok {
			temporalDecl = td
			break
		}
	}

	if temporalDecl == nil {
		t.Fatal("expected temporal declaration")
	}

	// Create trace where counter eventually becomes 5
	trace := NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)

	for i := 1; i <= 5; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		trace.AddState(s, "increment", nil)
	}

	// Evaluate temporal property
	evaluator := NewTemporalEvaluator()
	holds, err := evaluator.EvaluateTemporalProperty(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}

	if holds {
		t.Error("expected temporal property 'always (counter < 5)' to be violated")
	}
}

func TestTemporalPropertyUntil(t *testing.T) {
	// Test spec with until property
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

temporal untilReachesFive {
  (counter < 5) until (counter == 5)
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Extract temporal declaration
	var temporalDecl *ast.TemporalDecl
	for _, decl := range file.Decls {
		if td, ok := decl.(*ast.TemporalDecl); ok {
			temporalDecl = td
			break
		}
	}

	if temporalDecl == nil {
		t.Fatal("expected temporal declaration")
	}

	// Create trace where counter < 5 until counter == 5
	trace := NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)

	for i := 1; i <= 5; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		trace.AddState(s, "increment", nil)
	}

	// Evaluate temporal property
	evaluator := NewTemporalEvaluator()
	holds, err := evaluator.EvaluateTemporalProperty(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}

	if !holds {
		t.Error("expected temporal property '(counter < 5) until (counter == 5)' to hold")
	}
}

func TestTemporalPropertyLeadsTo(t *testing.T) {
	// Test spec with leads-to property
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

temporal leadsToFive {
  (counter == 0) -> (counter == 5)
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Extract temporal declaration
	var temporalDecl *ast.TemporalDecl
	for _, decl := range file.Decls {
		if td, ok := decl.(*ast.TemporalDecl); ok {
			temporalDecl = td
			break
		}
	}

	if temporalDecl == nil {
		t.Fatal("expected temporal declaration")
	}

	// Create trace where counter == 0 leads to counter == 5
	trace := NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)

	for i := 1; i <= 5; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		trace.AddState(s, "increment", nil)
	}

	// Evaluate temporal property
	evaluator := NewTemporalEvaluator()
	holds, err := evaluator.EvaluateTemporalProperty(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}

	if !holds {
		t.Error("expected temporal property '(counter == 0) -> (counter == 5)' to hold")
	}
}

func TestTemporalPropertyWithFairness(t *testing.T) {
	// Test spec with fairness property
	spec := `
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

temporal weakFairness {
  WF(increment)
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Extract temporal declaration
	var temporalDecl *ast.TemporalDecl
	for _, decl := range file.Decls {
		if td, ok := decl.(*ast.TemporalDecl); ok {
			temporalDecl = td
			break
		}
	}

	if temporalDecl == nil {
		t.Fatal("expected temporal declaration")
	}

	// Create state machine
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	// Create trace where increment is continuously enabled and executes
	trace := NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)

	for i := 1; i <= 3; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		trace.AddState(s, "increment", nil)
	}

	// Evaluate temporal property with fairness checker
	checker := NewFairnessChecker(sm)
	holds, err := checker.EvaluateFairness(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating fairness property: %v", err)
	}

	if !holds {
		t.Error("expected temporal property 'WF(increment)' to hold")
	}
}

