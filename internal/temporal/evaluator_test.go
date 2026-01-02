package temporal

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestTemporalEvaluatorAlways(t *testing.T) {
	// Parse: always (counter >= 0)
	spec := `temporal test { always (counter >= 0) }`
	
	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()
	
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	
	temporalDecl, ok := file.Decls[0].(*ast.TemporalDecl)
	if !ok {
		t.Fatalf("expected TemporalDecl, got %T", file.Decls[0])
	}
	
	alwaysExpr, ok := temporalDecl.Expression.(*ast.AlwaysExpr)
	if !ok {
		t.Fatalf("expected AlwaysExpr, got %T", temporalDecl.Expression)
	}
	
	// Create trace with states where counter >= 0
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
	
	evaluator := NewTemporalEvaluator()
	result, err := evaluator.EvaluateTemporalProperty(alwaysExpr, trace)
	
	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}
	
	if !result {
		t.Error("expected always (counter >= 0) to hold")
	}
}

func TestTemporalEvaluatorAlwaysViolation(t *testing.T) {
	// Parse: always (counter >= 0)
	spec := `temporal test { always (counter >= 0) }`
	
	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()
	
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	
	temporalDecl, ok := file.Decls[0].(*ast.TemporalDecl)
	if !ok {
		t.Fatalf("expected TemporalDecl, got %T", file.Decls[0])
	}
	
	alwaysExpr, ok := temporalDecl.Expression.(*ast.AlwaysExpr)
	if !ok {
		t.Fatalf("expected AlwaysExpr, got %T", temporalDecl.Expression)
	}
	
	// Create trace with a state where counter < 0
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s1, "", nil)
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(-1))
	trace.AddState(s2, "decrement", nil)
	
	evaluator := NewTemporalEvaluator()
	result, err := evaluator.EvaluateTemporalProperty(alwaysExpr, trace)
	
	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}
	
	if result {
		t.Error("expected always (counter >= 0) to be violated")
	}
}

func TestTemporalEvaluatorEventually(t *testing.T) {
	// Parse: eventually (counter == 5)
	spec := `temporal test { eventually (counter == 5) }`
	
	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()
	
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	
	temporalDecl, ok := file.Decls[0].(*ast.TemporalDecl)
	if !ok {
		t.Fatalf("expected TemporalDecl, got %T", file.Decls[0])
	}
	
	eventuallyExpr, ok := temporalDecl.Expression.(*ast.EventuallyExpr)
	if !ok {
		t.Fatalf("expected EventuallyExpr, got %T", temporalDecl.Expression)
	}
	
	// Create trace where counter eventually becomes 5
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s1, "", nil)
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(3))
	trace.AddState(s2, "increment", nil)
	
	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(5))
	trace.AddState(s3, "increment", nil)
	
	evaluator := NewTemporalEvaluator()
	result, err := evaluator.EvaluateTemporalProperty(eventuallyExpr, trace)
	
	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}
	
	if !result {
		t.Error("expected eventually (counter == 5) to hold")
	}
}

func TestTemporalEvaluatorUntil(t *testing.T) {
	// Parse: (counter < 5) until (counter == 5)
	spec := `temporal test { counter < 5 until counter == 5 }`
	
	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()
	
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	
	temporalDecl, ok := file.Decls[0].(*ast.TemporalDecl)
	if !ok {
		t.Fatalf("expected TemporalDecl, got %T", file.Decls[0])
	}
	
	untilExpr, ok := temporalDecl.Expression.(*ast.UntilExpr)
	if !ok {
		t.Fatalf("expected UntilExpr, got %T", temporalDecl.Expression)
	}
	
	// Create trace where counter < 5 until counter == 5
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s1, "", nil)
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(3))
	trace.AddState(s2, "increment", nil)
	
	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(5))
	trace.AddState(s3, "increment", nil)
	
	evaluator := NewTemporalEvaluator()
	result, err := evaluator.EvaluateTemporalProperty(untilExpr, trace)
	
	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}
	
	if !result {
		t.Error("expected (counter < 5) until (counter == 5) to hold")
	}
}

func TestTemporalEvaluatorLeadsTo(t *testing.T) {
	// Parse: (counter == 0) → (counter == 5)
	spec := `temporal test { counter == 0 -> counter == 5 }`
	
	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()
	
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	
	temporalDecl, ok := file.Decls[0].(*ast.TemporalDecl)
	if !ok {
		t.Fatalf("expected TemporalDecl, got %T", file.Decls[0])
	}
	
	leadsToExpr, ok := temporalDecl.Expression.(*ast.LeadsToExpr)
	if !ok {
		t.Fatalf("expected LeadsToExpr, got %T", temporalDecl.Expression)
	}
	
	// Create trace where counter == 0 leads to counter == 5
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	trace.AddState(s1, "", nil)
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(3))
	trace.AddState(s2, "increment", nil)
	
	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(5))
	trace.AddState(s3, "increment", nil)
	
	evaluator := NewTemporalEvaluator()
	result, err := evaluator.EvaluateTemporalProperty(leadsToExpr, trace)
	
	if err != nil {
		t.Fatalf("error evaluating temporal property: %v", err)
	}
	
	if !result {
		t.Error("expected (counter == 0) → (counter == 5) to hold")
	}
}

