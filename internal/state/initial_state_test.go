package state

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
)

func TestNewInitialStateModelDeterministic(t *testing.T) {
	input := `
var counter: int

init {
    counter = 0
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model == nil {
		t.Fatal("model should not be nil")
	}

	if !model.IsDeterministic() {
		t.Error("model should be deterministic")
	}

	if model.IsOneOf() {
		t.Error("model should not be oneOf")
	}

	if model.Count() != 1 {
		t.Errorf("expected 1 initial state, got %d", model.Count())
	}

	initStates := model.GetInitialStates()
	if len(initStates) != 1 {
		t.Errorf("expected 1 initial state config, got %d", len(initStates))
	}

	if initStates[0].InitDecl == nil {
		t.Error("InitDecl should not be nil")
	}
}

func TestNewInitialStateModelOneOf(t *testing.T) {
	input := `
var counter: int

init oneOf {
    { counter = 0 },
    { counter = 5 },
    { counter = 10 }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model == nil {
		t.Fatal("model should not be nil")
	}

	if model.IsDeterministic() {
		t.Error("model should not be deterministic")
	}

	if !model.IsOneOf() {
		t.Error("model should be oneOf")
	}

	if model.Count() != 3 {
		t.Errorf("expected 3 initial states, got %d", model.Count())
	}

	initStates := model.GetInitialStates()
	if len(initStates) != 3 {
		t.Errorf("expected 3 initial state configs, got %d", len(initStates))
	}

	for i, config := range initStates {
		if config.OneOfOption == nil {
			t.Errorf("OneOfOption should not be nil for option %d", i)
		}
	}
}

func TestNewInitialStateModelNoInit(t *testing.T) {
	input := `
var counter: int
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model, err := NewInitialStateModel(file)
	if err == nil {
		t.Error("expected error for missing init declaration")
	}
	if model != nil {
		t.Error("model should be nil when error occurs")
	}
}

func TestNewInitialStateModelMultipleInit(t *testing.T) {
	input := `
var counter: int

init {
    counter = 0
}

init {
    counter = 1
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model, err := NewInitialStateModel(file)
	if err == nil {
		t.Error("expected error for multiple init declarations")
	}
	if model != nil {
		t.Error("model should be nil when error occurs")
	}
}

func TestNewInitialStateModelBothInitAndOneOf(t *testing.T) {
	input := `
var counter: int

init {
    counter = 0
}

init oneOf {
    { counter = 5 }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model, err := NewInitialStateModel(file)
	if err == nil {
		t.Error("expected error for both init and oneOf")
	}
	if model != nil {
		t.Error("model should be nil when error occurs")
	}
}

func TestInitialStateModelFromModule(t *testing.T) {
	input := `
module Counter {
    var counter: int
    
    init {
        counter = 0
    }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model == nil {
		t.Fatal("model should not be nil")
	}

	if !model.IsDeterministic() {
		t.Error("model should be deterministic")
	}

	if model.Count() != 1 {
		t.Errorf("expected 1 initial state, got %d", model.Count())
	}
}

func TestInitialStateModelOneOfMultipleVariables(t *testing.T) {
	input := `
var counter: int
var flag: bool

init oneOf {
    { counter = 0, flag = false },
    { counter = 10, flag = true }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model == nil {
		t.Fatal("model should not be nil")
	}

	if !model.IsOneOf() {
		t.Error("model should be oneOf")
	}

	if model.Count() != 2 {
		t.Errorf("expected 2 initial states, got %d", model.Count())
	}

	options := model.GetOneOfOptions()
	if len(options) != 2 {
		t.Errorf("expected 2 oneOf options, got %d", len(options))
	}
}

func TestInitialStateModelGetDeterministicInit(t *testing.T) {
	input := `
var counter: int

init {
    counter = 0
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	initDecl := model.GetDeterministicInit()
	if initDecl == nil {
		t.Error("GetDeterministicInit should not return nil")
	}

	// For oneOf, should return nil
	input2 := `
var counter: int

init oneOf {
    { counter = 0 }
}
`

	p2 := parser.New(lexer.New(input2))
	file2 := p2.ParseFile()
	if len(p2.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p2.Errors())
	}

	model2, err := NewInitialStateModel(file2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	initDecl2 := model2.GetDeterministicInit()
	if initDecl2 != nil {
		t.Error("GetDeterministicInit should return nil for oneOf")
	}
}

func TestInitialStateModelGetOneOfOptions(t *testing.T) {
	input := `
var counter: int

init oneOf {
    { counter = 0 },
    { counter = 5 },
    { counter = 10 }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model, err := NewInitialStateModel(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	options := model.GetOneOfOptions()
	if len(options) != 3 {
		t.Errorf("expected 3 options, got %d", len(options))
	}

	// For deterministic, should return nil
	input2 := `
var counter: int

init {
    counter = 0
}
`

	p2 := parser.New(lexer.New(input2))
	file2 := p2.ParseFile()
	if len(p2.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p2.Errors())
	}

	model2, err := NewInitialStateModel(file2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	options2 := model2.GetOneOfOptions()
	if options2 != nil {
		t.Error("GetOneOfOptions should return nil for deterministic")
	}
}

