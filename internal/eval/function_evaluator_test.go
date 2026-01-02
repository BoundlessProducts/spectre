package eval

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
	"github.com/spectre-lang/spectre/internal/state"
)

func TestFunctionEvaluatorSimpleFunctions(t *testing.T) {
	// Parse a simple spec with pure functions
	spec := `
var counter: int

fun add(x: int, y: int): int {
  x + y
}

fun multiply(x: int, y: int): int {
  x * y
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	fe, err := NewFunctionEvaluator(file)
	if err != nil {
		t.Fatalf("failed to create function evaluator: %v", err)
	}

	// Test add function
	result, err := fe.CallFunction("add", []state.Value{
		state.NewIntValue(5),
		state.NewIntValue(3),
	})
	if err != nil {
		t.Fatalf("error calling add: %v", err)
	}

	expected := state.NewIntValue(8)
	if !valuesEqual(result, expected) {
		t.Errorf("add(5, 3): expected %s, got %s", expected.String(), result.String())
	}

	// Test multiply function
	result, err = fe.CallFunction("multiply", []state.Value{
		state.NewIntValue(4),
		state.NewIntValue(7),
	})
	if err != nil {
		t.Fatalf("error calling multiply: %v", err)
	}

	expected = state.NewIntValue(28)
	if !valuesEqual(result, expected) {
		t.Errorf("multiply(4, 7): expected %s, got %s", expected.String(), result.String())
	}
}

func TestFunctionEvaluatorRecursiveFunction(t *testing.T) {
	spec := `
var counter: int

fun factorial(n: int): int {
  if (n <= 1) {
    1
  } else {
    n * factorial(n - 1)
  }
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	fe, err := NewFunctionEvaluator(file)
	if err != nil {
		t.Fatalf("failed to create function evaluator: %v", err)
	}

	// Test factorial(5) = 120
	result, err := fe.CallFunction("factorial", []state.Value{
		state.NewIntValue(5),
	})
	if err != nil {
		t.Fatalf("error calling factorial: %v", err)
	}

	expected := state.NewIntValue(120)
	if !valuesEqual(result, expected) {
		t.Errorf("factorial(5): expected %s, got %s", expected.String(), result.String())
	}
}

func TestFunctionEvaluatorPurityViolation(t *testing.T) {
	spec := `
var counter: int

fun getCounter(): int {
  return counter
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	_, err := NewFunctionEvaluator(file)
	if err == nil {
		t.Error("expected purity violation error for accessing state variable")
	}
}

func TestFunctionEvaluatorMaxFunction(t *testing.T) {
	spec := `
var counter: int

fun max(a: int, b: int): int {
  if (a > b) {
    a
  } else {
    b
  }
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	fe, err := NewFunctionEvaluator(file)
	if err != nil {
		t.Fatalf("failed to create function evaluator: %v", err)
	}

	// Test max(10, 20) = 20
	result, err := fe.CallFunction("max", []state.Value{
		state.NewIntValue(10),
		state.NewIntValue(20),
	})
	if err != nil {
		t.Fatalf("error calling max: %v", err)
	}

	expected := state.NewIntValue(20)
	if !valuesEqual(result, expected) {
		t.Errorf("max(10, 20): expected %s, got %s", expected.String(), result.String())
	}

	// Test max(30, 15) = 30
	result, err = fe.CallFunction("max", []state.Value{
		state.NewIntValue(30),
		state.NewIntValue(15),
	})
	if err != nil {
		t.Fatalf("error calling max: %v", err)
	}

	expected = state.NewIntValue(30)
	if !valuesEqual(result, expected) {
		t.Errorf("max(30, 15): expected %s, got %s", expected.String(), result.String())
	}
}

func TestFunctionEvaluatorMinFunction(t *testing.T) {
	spec := `
var counter: int

fun min(a: int, b: int): int {
  if (a < b) {
    a
  } else {
    b
  }
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	fe, err := NewFunctionEvaluator(file)
	if err != nil {
		t.Fatalf("failed to create function evaluator: %v", err)
	}

	// Test min(10, 20) = 10
	result, err := fe.CallFunction("min", []state.Value{
		state.NewIntValue(10),
		state.NewIntValue(20),
	})
	if err != nil {
		t.Fatalf("error calling min: %v", err)
	}

	expected := state.NewIntValue(10)
	if !valuesEqual(result, expected) {
		t.Errorf("min(10, 20): expected %s, got %s", expected.String(), result.String())
	}
}

