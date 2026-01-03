package exec

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/state"
)

func TestActionExecutorSimpleAction(t *testing.T) {
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

	vm := state.NewVariableModel(file)
	am := state.NewActionModel(file)

	cm := state.NewConstraintModel(file, am)
	executor := NewActionExecutor(vm, am, cm, file)

	// Create initial state
	ism, _ := state.NewInitialStateModel(file)
	initializer := NewStateInitializer(vm, ism, &ast.File{Decls: []ast.Decl{}})
	initialStates, _ := initializer.GenerateInitialStates()
	currentState := initialStates[0]

	// Execute increment action
	nextState, err := executor.ExecuteAction("increment", currentState, nil)
	if err != nil {
		t.Fatalf("error executing action: %v", err)
	}

	// Check that counter was incremented
	counterVal, exists := nextState.GetVariable("counter")
	if !exists {
		t.Error("counter variable not found in next state")
	}

	if pv, ok := counterVal.(*state.PrimitiveValue); ok {
		if pv.IntValue == nil {
			t.Error("counter value is nil")
		} else if *pv.IntValue != 1 {
			t.Errorf("expected counter = 1, got %d", *pv.IntValue)
		}
	} else {
		t.Error("counter is not a primitive value")
	}
}

func TestActionExecutorWithParameters(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action add(amount: int) {
  counter' = counter + amount
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := state.NewVariableModel(file)
	am := state.NewActionModel(file)

	cm := state.NewConstraintModel(file, am)
	executor := NewActionExecutor(vm, am, cm, file)

	ism, _ := state.NewInitialStateModel(file)
	initializer := NewStateInitializer(vm, ism, &ast.File{Decls: []ast.Decl{}})
	initialStates, _ := initializer.GenerateInitialStates()
	currentState := initialStates[0]

	// Execute add action with argument 5
	nextState, err := executor.ExecuteAction("add", currentState, []state.Value{
		state.NewIntValue(5),
	})
	if err != nil {
		t.Fatalf("error executing action: %v", err)
	}

	// Check that counter was incremented by 5
	counterVal, _ := nextState.GetVariable("counter")
	if pv, ok := counterVal.(*state.PrimitiveValue); ok && pv.IntValue != nil {
		if *pv.IntValue != 5 {
			t.Errorf("expected counter = 5, got %d", *pv.IntValue)
		}
	}
}

func TestActionExecutorMultipleAssignments(t *testing.T) {
	spec := `
var x: int
var y: int

init {
  x = 0
  y = 0
}

action setBoth(a: int, b: int) {
  x' = a
  y' = b
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := state.NewVariableModel(file)
	am := state.NewActionModel(file)

	cm := state.NewConstraintModel(file, am)
	executor := NewActionExecutor(vm, am, cm, file)

	ism, _ := state.NewInitialStateModel(file)
	initializer := NewStateInitializer(vm, ism, &ast.File{Decls: []ast.Decl{}})
	initialStates, _ := initializer.GenerateInitialStates()
	currentState := initialStates[0]

	// Execute setBoth action
	nextState, err := executor.ExecuteAction("setBoth", currentState, []state.Value{
		state.NewIntValue(10),
		state.NewIntValue(20),
	})
	if err != nil {
		t.Fatalf("error executing action: %v", err)
	}

	// Check both variables
	xVal, _ := nextState.GetVariable("x")
	if pv, ok := xVal.(*state.PrimitiveValue); ok && pv.IntValue != nil {
		if *pv.IntValue != 10 {
			t.Errorf("expected x = 10, got %d", *pv.IntValue)
		}
	}

	yVal, _ := nextState.GetVariable("y")
	if pv, ok := yVal.(*state.PrimitiveValue); ok && pv.IntValue != nil {
		if *pv.IntValue != 20 {
			t.Errorf("expected y = 20, got %d", *pv.IntValue)
		}
	}
}

func TestActionExecutorCanExecute(t *testing.T) {
	spec := `
var counter: int

init {
  counter = 0
}

action increment when counter < 10 {
  counter' = counter + 1
}
`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := state.NewVariableModel(file)
	am := state.NewActionModel(file)

	cm := state.NewConstraintModel(file, am)
	executor := NewActionExecutor(vm, am, cm, file)

	ism, _ := state.NewInitialStateModel(file)
	initializer := NewStateInitializer(vm, ism, &ast.File{Decls: []ast.Decl{}})
	initialStates, _ := initializer.GenerateInitialStates()
	currentState := initialStates[0]

	// Check if action can be executed (counter = 0 < 10, so should be true)
	canExecute, err := executor.CanExecute("increment", currentState, nil)
	if err != nil {
		t.Fatalf("error checking if action can execute: %v", err)
	}

	if !canExecute {
		t.Error("expected action to be executable when counter < 10")
	}

	// Set counter to 10 and check again
	currentState.SetVariable("counter", state.NewIntValue(10))
	canExecute, err = executor.CanExecute("increment", currentState, nil)
	if err != nil {
		t.Fatalf("error checking if action can execute: %v", err)
	}

	if canExecute {
		t.Error("expected action to not be executable when counter >= 10")
	}
}

