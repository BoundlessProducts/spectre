package state

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

func TestNewActionModel(t *testing.T) {
	input := `
var counter: int

action increment {
    counter' = counter + 1
}

action decrement {
    counter' = counter - 1
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewActionModel(file)
	if model == nil {
		t.Fatal("NewActionModel returned nil")
	}

	if len(model.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(model.Actions))
	}

	// Check individual actions
	if _, exists := model.Actions["increment"]; !exists {
		t.Error("action 'increment' not found")
	}
	if _, exists := model.Actions["decrement"]; !exists {
		t.Error("action 'decrement' not found")
	}
}

func TestActionModelGetAction(t *testing.T) {
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

	model := NewActionModel(file)

	// Get existing action
	info, found := model.GetAction("increment")
	if !found {
		t.Error("action 'increment' should be found")
	}
	if info == nil {
		t.Error("action info should not be nil")
	}
	if info.Name != "increment" {
		t.Errorf("expected name 'increment', got '%s'", info.Name)
	}

	// Get non-existent action
	_, found = model.GetAction("nonexistent")
	if found {
		t.Error("action 'nonexistent' should not be found")
	}
}

func TestActionModelFromModule(t *testing.T) {
	input := `
module Counter {
    var counter: int
    
    action increment {
        counter' = counter + 1
    }
    
    action decrement {
        counter' = counter - 1
    }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewActionModel(file)
	if len(model.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(model.Actions))
	}

	if _, exists := model.Actions["increment"]; !exists {
		t.Error("action 'increment' not found")
	}
	if _, exists := model.Actions["decrement"]; !exists {
		t.Error("action 'decrement' not found")
	}
}

func TestActionModelWithGuard(t *testing.T) {
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

	model := NewActionModel(file)
	info, found := model.GetAction("decrement")
	if !found {
		t.Fatal("action 'decrement' not found")
	}

	// Note: Guards are parsed as 'require' statements in the body, not as Guard field
	// The Guard field is for 'when' clauses
	if info.Body == nil {
		t.Error("action body should not be nil")
	}
}

func TestActionModelWithWhenGuard(t *testing.T) {
	input := `
var counter: int

action decrement when counter > 0 {
    counter' = counter - 1
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewActionModel(file)
	info, found := model.GetAction("decrement")
	if !found {
		t.Fatal("action 'decrement' not found")
	}

	if !info.HasGuard() {
		t.Error("action should have a guard (when clause)")
	}

	if info.Guard == nil {
		t.Error("guard expression should not be nil")
	}
}

func TestActionModelWithParameters(t *testing.T) {
	input := `
var counter: int

action add(x: int) {
    counter' = counter + x
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewActionModel(file)
	info, found := model.GetAction("add")
	if !found {
		t.Fatal("action 'add' not found")
	}

	if info.GetParameterCount() != 1 {
		t.Errorf("expected 1 parameter, got %d", info.GetParameterCount())
	}

	paramNames := info.GetParameterNames()
	if len(paramNames) != 1 {
		t.Errorf("expected 1 parameter name, got %d", len(paramNames))
	}
	if paramNames[0] != "x" {
		t.Errorf("expected parameter name 'x', got '%s'", paramNames[0])
	}
}

func TestActionModelGetActionNames(t *testing.T) {
	input := `
var counter: int

action increment {
    counter' = counter + 1
}

action decrement {
    counter' = counter - 1
}

action reset {
    counter' = 0
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewActionModel(file)
	names := model.GetActionNames()

	if len(names) != 3 {
		t.Errorf("expected 3 action names, got %d", len(names))
	}

	// Check that all expected names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	expectedNames := []string{"increment", "decrement", "reset"}
	for _, expected := range expectedNames {
		if !nameMap[expected] {
			t.Errorf("expected action name '%s' not found", expected)
		}
	}
}

func TestActionModelHasAction(t *testing.T) {
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

	model := NewActionModel(file)

	if !model.HasAction("increment") {
		t.Error("HasAction should return true for 'increment'")
	}

	if model.HasAction("nonexistent") {
		t.Error("HasAction should return false for 'nonexistent'")
	}
}

func TestActionInfoMethods(t *testing.T) {
	input := `
var counter: int

action increment {
    counter' = counter + 1
}

action decrement when counter > 0 {
    counter' = counter - 1
}

action add(x: int, y: int) {
    counter' = counter + x + y
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewActionModel(file)

	// Test increment (no guard, no parameters)
	incInfo, _ := model.GetAction("increment")
	if !incInfo.HasBody() {
		t.Error("increment should have a body")
	}
	if incInfo.HasGuard() {
		t.Error("increment should not have a guard")
	}
	if incInfo.GetParameterCount() != 0 {
		t.Errorf("increment should have 0 parameters, got %d", incInfo.GetParameterCount())
	}

	// Test decrement (has guard, no parameters)
	decInfo, _ := model.GetAction("decrement")
	if !decInfo.HasBody() {
		t.Error("decrement should have a body")
	}
	if !decInfo.HasGuard() {
		t.Error("decrement should have a guard")
	}
	if decInfo.GetParameterCount() != 0 {
		t.Errorf("decrement should have 0 parameters, got %d", decInfo.GetParameterCount())
	}

	// Test add (no guard, has parameters)
	addInfo, _ := model.GetAction("add")
	if !addInfo.HasBody() {
		t.Error("add should have a body")
	}
	if addInfo.HasGuard() {
		t.Error("add should not have a guard")
	}
	if addInfo.GetParameterCount() != 2 {
		t.Errorf("add should have 2 parameters, got %d", addInfo.GetParameterCount())
	}

	paramNames := addInfo.GetParameterNames()
	if len(paramNames) != 2 {
		t.Errorf("expected 2 parameter names, got %d", len(paramNames))
	}
	if paramNames[0] != "x" || paramNames[1] != "y" {
		t.Errorf("expected parameter names ['x', 'y'], got %v", paramNames)
	}
}

func TestActionInfoString(t *testing.T) {
	input := `
var counter: int

action increment {
    counter' = counter + 1
}

action add(x: int) {
    counter' = counter + x
}

action decrement when counter > 0 {
    counter' = counter - 1
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	model := NewActionModel(file)

	// Test string representation
	incInfo, _ := model.GetAction("increment")
	str := incInfo.String()
	if str == "" {
		t.Error("String() should not return empty string")
	}
	if len(str) < len("action increment") {
		t.Error("String() should contain action name")
	}

	addInfo, _ := model.GetAction("add")
	str = addInfo.String()
	if str == "" {
		t.Error("String() should not return empty string")
	}

	decInfo, _ := model.GetAction("decrement")
	str = decInfo.String()
	if str == "" {
		t.Error("String() should not return empty string")
	}
}

