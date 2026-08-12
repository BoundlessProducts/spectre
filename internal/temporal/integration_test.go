package temporal_test

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/internal/temporal"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestTemporalPropertyVerification(t *testing.T) {
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

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(10)
	explorer.SetMaxStates(20)

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("error exploring state space: %v", err)
	}

	var trace *temporal.Trace
	for _, reachableState := range result.ReachableStates {
		counterVal, exists := reachableState.GetVariable("counter")
		if exists && counterVal != nil {
			if pv, ok := counterVal.(*state.PrimitiveValue); ok && pv.IntValue != nil {
				if *pv.IntValue == 5 {
					trace = temporal.NewTrace()
					initialStates, _ := sm.GetInitialStates()
					if len(initialStates) > 0 {
						trace.AddState(initialStates[0], "", nil)
					}
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
		trace = temporal.NewTrace()
		s0 := state.NewState()
		s0.SetVariable("counter", state.NewIntValue(0))
		trace.AddState(s0, "", nil)
		for i := 1; i <= 5; i++ {
			s := state.NewState()
			s.SetVariable("counter", state.NewIntValue(int64(i)))
			trace.AddState(s, "increment", nil)
		}
	}

	evaluator := temporal.NewTemporalEvaluator()
	holds, err := evaluator.EvaluateTemporalProperty(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}

	if !holds {
		t.Error("expected temporal property 'eventually (counter == 5)' to hold")
	}
}

func TestTemporalPropertyAlwaysViolation(t *testing.T) {
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

	trace := temporal.NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)
	for i := 1; i <= 5; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		trace.AddState(s, "increment", nil)
	}

	evaluator := temporal.NewTemporalEvaluator()
	holds, err := evaluator.EvaluateTemporalProperty(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}

	if holds {
		t.Error("expected temporal property 'always (counter < 5)' to be violated")
	}
}

func TestTemporalPropertyUntil(t *testing.T) {
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

	trace := temporal.NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)
	for i := 1; i <= 5; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		trace.AddState(s, "increment", nil)
	}

	evaluator := temporal.NewTemporalEvaluator()
	holds, err := evaluator.EvaluateTemporalProperty(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}

	if !holds {
		t.Error("expected temporal property '(counter < 5) until (counter == 5)' to hold")
	}
}

func TestTemporalPropertyLeadsTo(t *testing.T) {
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

	trace := temporal.NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)
	for i := 1; i <= 5; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		trace.AddState(s, "increment", nil)
	}

	evaluator := temporal.NewTemporalEvaluator()
	holds, err := evaluator.EvaluateTemporalProperty(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}

	if !holds {
		t.Error("expected temporal property '(counter == 0) -> (counter == 5)' to hold")
	}
}

func TestTemporalPropertyWithFairness(t *testing.T) {
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

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		t.Fatalf("error creating state machine: %v", err)
	}

	trace := temporal.NewTrace()
	s0 := state.NewState()
	s0.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s0, "", nil)
	for i := 1; i <= 3; i++ {
		s := state.NewState()
		s.SetVariable("counter", state.NewIntValue(int64(i)))
		trace.AddState(s, "increment", nil)
	}

	checker := temporal.NewFairnessChecker(sm)
	holds, err := checker.EvaluateFairness(temporalDecl.Expression, trace)

	if err != nil {
		t.Fatalf("error evaluating fairness property: %v", err)
	}

	if !holds {
		t.Error("expected temporal property 'WF(increment)' to hold")
	}
}
