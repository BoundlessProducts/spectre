package state

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
)

func TestNewConstraintModel(t *testing.T) {
	input := `
var counter: int

invariant nonNegative {
    counter >= 0
}

invariant bounded {
    counter <= 100
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	actionModel := NewActionModel(file)
	model := NewConstraintModel(file, actionModel)
	if model == nil {
		t.Fatal("NewConstraintModel returned nil")
	}

	if len(model.Invariants) != 2 {
		t.Errorf("expected 2 invariants, got %d", len(model.Invariants))
	}

	if !model.HasInvariants() {
		t.Error("model should have invariants")
	}

	if model.GetInvariantCount() != 2 {
		t.Errorf("expected 2 invariants, got %d", model.GetInvariantCount())
	}
}

func TestConstraintModelGetInvariant(t *testing.T) {
	input := `
var counter: int

invariant nonNegative {
    counter >= 0
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	actionModel := NewActionModel(file)
	model := NewConstraintModel(file, actionModel)

	// Get existing invariant
	inv := model.GetInvariant("nonNegative")
	if inv == nil {
		t.Error("invariant 'nonNegative' should be found")
	}
	if inv.Name != "nonNegative" {
		t.Errorf("expected name 'nonNegative', got '%s'", inv.Name)
	}

	// Get non-existent invariant
	inv = model.GetInvariant("nonexistent")
	if inv != nil {
		t.Error("invariant 'nonexistent' should not be found")
	}
}

func TestConstraintModelFromModule(t *testing.T) {
	input := `
module Counter {
    var counter: int
    
    invariant nonNegative {
        counter >= 0
    }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	actionModel := NewActionModel(file)
	model := NewConstraintModel(file, actionModel)
	if len(model.Invariants) != 1 {
		t.Errorf("expected 1 invariant, got %d", len(model.Invariants))
	}

	inv := model.GetInvariant("nonNegative")
	if inv == nil {
		t.Error("invariant 'nonNegative' not found")
	}
}

func TestConstraintModelPreconditions(t *testing.T) {
	input := `
var counter: int

action decrement {
    require counter > 0
    counter' = counter - 1
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	actionModel := NewActionModel(file)
	model := NewConstraintModel(file, actionModel)

	preconds := model.GetPreconditions("decrement")
	if len(preconds) != 1 {
		t.Errorf("expected 1 precondition, got %d", len(preconds))
	}

	if !model.HasPreconditions("decrement") {
		t.Error("action 'decrement' should have preconditions")
	}

	if preconds[0].Action != "decrement" {
		t.Errorf("expected action name 'decrement', got '%s'", preconds[0].Action)
	}

	if preconds[0].Condition == nil {
		t.Error("precondition condition should not be nil")
	}
}

func TestConstraintModelPostconditions(t *testing.T) {
	input := `
var counter: int

action increment {
    counter' = counter + 1
    ensure counter > 0
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	actionModel := NewActionModel(file)
	model := NewConstraintModel(file, actionModel)

	postconds := model.GetPostconditions("increment")
	if len(postconds) != 1 {
		t.Errorf("expected 1 postcondition, got %d", len(postconds))
	}

	if !model.HasPostconditions("increment") {
		t.Error("action 'increment' should have postconditions")
	}

	if postconds[0].Action != "increment" {
		t.Errorf("expected action name 'increment', got '%s'", postconds[0].Action)
	}

	if postconds[0].Condition == nil {
		t.Error("postcondition condition should not be nil")
	}
}

func TestConstraintModelMultiplePreconditions(t *testing.T) {
	input := `
var counter: int
var maxValue: int

action increment {
    require counter >= 0
    require counter < maxValue
    counter' = counter + 1
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	actionModel := NewActionModel(file)
	model := NewConstraintModel(file, actionModel)

	preconds := model.GetPreconditions("increment")
	if len(preconds) != 2 {
		t.Errorf("expected 2 preconditions, got %d", len(preconds))
	}
}

func TestConstraintModelMultiplePostconditions(t *testing.T) {
	input := `
var counter: int

action increment {
    counter' = counter + 1
    ensure counter > 0
    ensure counter <= 100
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	actionModel := NewActionModel(file)
	model := NewConstraintModel(file, actionModel)

	postconds := model.GetPostconditions("increment")
	if len(postconds) != 2 {
		t.Errorf("expected 2 postconditions, got %d", len(postconds))
	}
}

func TestConstraintModelNoConstraints(t *testing.T) {
	input := `
var counter: int

action increment {
    counter' = counter + 1
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	actionModel := NewActionModel(file)
	model := NewConstraintModel(file, actionModel)

	if model.HasInvariants() {
		t.Error("model should not have invariants")
	}

	if model.HasPreconditions("increment") {
		t.Error("action 'increment' should not have preconditions")
	}

	if model.HasPostconditions("increment") {
		t.Error("action 'increment' should not have postconditions")
	}
}

func TestConstraintModelComplete(t *testing.T) {
	input := `
var counter: int

invariant nonNegative {
    counter >= 0
}

action increment {
    require counter < 100
    counter' = counter + 1
    ensure counter > 0
}

action decrement {
    require counter > 0
    counter' = counter - 1
    ensure counter >= 0
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	actionModel := NewActionModel(file)
	model := NewConstraintModel(file, actionModel)

	// Check invariants
	if model.GetInvariantCount() != 1 {
		t.Errorf("expected 1 invariant, got %d", model.GetInvariantCount())
	}

	// Check preconditions
	if !model.HasPreconditions("increment") {
		t.Error("action 'increment' should have preconditions")
	}
	if !model.HasPreconditions("decrement") {
		t.Error("action 'decrement' should have preconditions")
	}

	// Check postconditions
	if !model.HasPostconditions("increment") {
		t.Error("action 'increment' should have postconditions")
	}
	if !model.HasPostconditions("decrement") {
		t.Error("action 'decrement' should have postconditions")
	}

	// Verify counts
	incPreconds := model.GetPreconditions("increment")
	if len(incPreconds) != 1 {
		t.Errorf("expected 1 precondition for increment, got %d", len(incPreconds))
	}

	decPreconds := model.GetPreconditions("decrement")
	if len(decPreconds) != 1 {
		t.Errorf("expected 1 precondition for decrement, got %d", len(decPreconds))
	}

	incPostconds := model.GetPostconditions("increment")
	if len(incPostconds) != 1 {
		t.Errorf("expected 1 postcondition for increment, got %d", len(incPostconds))
	}

	decPostconds := model.GetPostconditions("decrement")
	if len(decPostconds) != 1 {
		t.Errorf("expected 1 postcondition for decrement, got %d", len(decPostconds))
	}
}

