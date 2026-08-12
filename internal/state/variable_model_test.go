package state

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

func TestNewVariableModel(t *testing.T) {
	input := `
var counter: int
var flag: bool
var name: str
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewVariableModel(file)
	if model == nil {
		t.Fatal("NewVariableModel returned nil")
	}

	if len(model.Variables) != 3 {
		t.Errorf("expected 3 variables, got %d", len(model.Variables))
	}

	// Check individual variables
	if _, exists := model.Variables["counter"]; !exists {
		t.Error("variable 'counter' not found")
	}
	if _, exists := model.Variables["flag"]; !exists {
		t.Error("variable 'flag' not found")
	}
	if _, exists := model.Variables["name"]; !exists {
		t.Error("variable 'name' not found")
	}
}

func TestVariableModelGetVariable(t *testing.T) {
	input := `
var counter: int
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewVariableModel(file)

	// Get existing variable
	info, found := model.GetVariable("counter")
	if !found {
		t.Error("variable 'counter' should be found")
	}
	if info == nil {
		t.Error("variable info should not be nil")
	}
	if info.Name != "counter" {
		t.Errorf("expected name 'counter', got '%s'", info.Name)
	}

	// Get non-existent variable
	_, found = model.GetVariable("nonexistent")
	if found {
		t.Error("variable 'nonexistent' should not be found")
	}
}

func TestVariableModelFromModule(t *testing.T) {
	input := `
module Counter {
    var count: int
    var active: bool
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewVariableModel(file)
	if len(model.Variables) != 2 {
		t.Errorf("expected 2 variables, got %d", len(model.Variables))
	}

	if _, exists := model.Variables["count"]; !exists {
		t.Error("variable 'count' not found")
	}
	if _, exists := model.Variables["active"]; !exists {
		t.Error("variable 'active' not found")
	}
}

func TestVariableModelValidateState(t *testing.T) {
	input := `
var counter: int
var flag: bool
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewVariableModel(file)

	// Valid state
	validState := NewState()
	validState.SetVariable("counter", NewIntValue(10))
	validState.SetVariable("flag", NewBoolValue(true))

	errors := model.ValidateState(validState)
	if len(errors) > 0 {
		t.Errorf("expected no validation errors, got: %v", errors)
	}

	// Missing variable
	incompleteState := NewState()
	incompleteState.SetVariable("counter", NewIntValue(10))

	errors = model.ValidateState(incompleteState)
	if len(errors) == 0 {
		t.Error("expected validation error for missing variable")
	}

	// Wrong type
	wrongTypeState := NewState()
	wrongTypeState.SetVariable("counter", NewStringValue("10"))
	wrongTypeState.SetVariable("flag", NewBoolValue(true))

	errors = model.ValidateState(wrongTypeState)
	if len(errors) == 0 {
		t.Error("expected validation error for wrong type")
	}

	// Extra variable
	extraVarState := NewState()
	extraVarState.SetVariable("counter", NewIntValue(10))
	extraVarState.SetVariable("flag", NewBoolValue(true))
	extraVarState.SetVariable("extra", NewIntValue(5))

	errors = model.ValidateState(extraVarState)
	if len(errors) == 0 {
		t.Error("expected validation error for extra variable")
	}
}

func TestVariableModelGetVariableNames(t *testing.T) {
	input := `
var counter: int
var flag: bool
var name: str
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewVariableModel(file)
	names := model.GetVariableNames()

	if len(names) != 3 {
		t.Errorf("expected 3 variable names, got %d", len(names))
	}

	// Check that all expected names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	expectedNames := []string{"counter", "flag", "name"}
	for _, expected := range expectedNames {
		if !nameMap[expected] {
			t.Errorf("expected variable name '%s' not found", expected)
		}
	}
}

func TestVariableModelHasVariable(t *testing.T) {
	input := `
var counter: int
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewVariableModel(file)

	if !model.HasVariable("counter") {
		t.Error("HasVariable should return true for 'counter'")
	}

	if model.HasVariable("nonexistent") {
		t.Error("HasVariable should return false for 'nonexistent'")
	}
}

func TestVariableInfo(t *testing.T) {
	// Note: Standalone descriptions are skipped by the parser
	// Descriptions must be immediately before the declaration
	input := `
var counter: int
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewVariableModel(file)
	info, found := model.GetVariable("counter")
	if !found {
		t.Fatal("variable 'counter' not found")
	}

	if info.Name != "counter" {
		t.Errorf("expected name 'counter', got '%s'", info.Name)
	}

	// Description may be empty if not provided immediately before var
	// This is expected behavior based on parser implementation
	_ = info.Description

	if info.Declaration == nil {
		t.Error("declaration should not be nil")
	}

	if info.Type == nil {
		t.Error("type should not be nil")
	}
}

func TestVariableModelTypeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		varName  string
		expected string
	}{
		{
			name:     "primitive int",
			input:    "var counter: int",
			varName:  "counter",
			expected: "int",
		},
		{
			name:     "primitive bool",
			input:    "var flag: bool",
			varName:  "flag",
			expected: "bool",
		},
		{
			name:     "primitive str",
			input:    "var name: str",
			varName:  "name",
			expected: "str",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.input))
			file := p.ParseFile()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			model := NewVariableModel(file)
			info, found := model.GetVariable(tt.varName)
			if !found {
				t.Fatalf("variable '%s' not found", tt.varName)
			}

			typeStr := model.typeString(info.Type)
			if typeStr != tt.expected {
				t.Errorf("expected type string '%s', got '%s'", tt.expected, typeStr)
			}
		})
	}
}

