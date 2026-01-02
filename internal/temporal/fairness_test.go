package temporal

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/exec"
	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
	"github.com/spectre-lang/spectre/internal/state"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestFairnessCheckerWeakFairness(t *testing.T) {
	// Create a spec with an action that can be enabled
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

	// Parse WF expression - need to parse as temporal expression
	wfSpec := `temporal test { WF(increment) }`
	l2 := lexer.New(wfSpec)
	p2 := parser.New(l2)
	file2 := p2.ParseFile()
	
	if len(p2.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p2.Errors())
	}
	
	temporalDecl, ok := file2.Decls[0].(*ast.TemporalDecl)
	if !ok {
		t.Fatalf("expected TemporalDecl, got %T", file2.Decls[0])
	}
	
	expr := temporalDecl.Expression
	
	if len(p2.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p2.Errors())
	}

	wfExpr, ok := expr.(*ast.WFExpr)
	if !ok {
		t.Fatalf("expected WFExpr, got %T", expr)
	}

	// Create trace where increment is continuously enabled and executes
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s1, "", nil)
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(1))
	trace.AddState(s2, "increment", nil)
	
	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(2))
	trace.AddState(s3, "increment", nil)

	checker := NewFairnessChecker(sm)
	result, err := checker.EvaluateFairness(wfExpr, trace)
	
	if err != nil {
		t.Fatalf("error evaluating fairness: %v", err)
	}
	
	if !result {
		t.Error("expected WF(increment) to hold")
	}
}

func TestFairnessCheckerStrongFairness(t *testing.T) {
	// Create a spec with an action
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

	// Parse SF expression - need to parse as temporal expression
	sfSpec := `temporal test { SF(increment) }`
	l2 := lexer.New(sfSpec)
	p2 := parser.New(l2)
	file2 := p2.ParseFile()
	
	if len(p2.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p2.Errors())
	}
	
	temporalDecl, ok := file2.Decls[0].(*ast.TemporalDecl)
	if !ok {
		t.Fatalf("expected TemporalDecl, got %T", file2.Decls[0])
	}
	
	expr := temporalDecl.Expression
	
	sfExpr, ok := expr.(*ast.SFExpr)
	if !ok {
		t.Fatalf("expected SFExpr, got %T", expr)
	}

	// Create trace where increment is enabled multiple times and executes
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s1, "", nil)
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(1))
	trace.AddState(s2, "increment", nil)
	
	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(2))
	trace.AddState(s3, "increment", nil)

	checker := NewFairnessChecker(sm)
	result, err := checker.EvaluateFairness(sfExpr, trace)
	
	if err != nil {
		t.Fatalf("error evaluating fairness: %v", err)
	}
	
	if !result {
		t.Error("expected SF(increment) to hold")
	}
}

