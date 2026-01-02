package eval

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/state"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestNewEnvironment(t *testing.T) {
	env := NewEnvironment()
	if env == nil {
		t.Fatal("NewEnvironment returned nil")
	}
	if env.variables == nil {
		t.Error("variables map should not be nil")
	}
	if env.functions == nil {
		t.Error("functions map should not be nil")
	}
	if env.constants == nil {
		t.Error("constants map should not be nil")
	}
	if env.parent != nil {
		t.Error("root environment should not have a parent")
	}
}

func TestEnvironmentSetAndGetVariable(t *testing.T) {
	env := NewEnvironment()

	// Set a variable
	intVal := state.NewIntValue(42)
	env.SetVariable("x", intVal)

	// Get the variable
	value, exists := env.GetVariable("x")
	if !exists {
		t.Error("variable 'x' should exist")
	}
	if value == nil {
		t.Error("value should not be nil")
	}

	// Verify it's the correct value
	if pv, ok := value.(*state.PrimitiveValue); ok {
		if pv.IntValue == nil || *pv.IntValue != 42 {
			t.Errorf("expected value 42, got %v", pv.IntValue)
		}
	} else {
		t.Error("value should be *PrimitiveValue")
	}
}

func TestEnvironmentNestedScopes(t *testing.T) {
	parent := NewEnvironment()
	parent.SetVariable("x", state.NewIntValue(10))

	child := NewChildEnvironment(parent)
	child.SetVariable("y", state.NewIntValue(20))

	// Child should see parent's variable
	value, exists := child.GetVariable("x")
	if !exists {
		t.Error("child should see parent's variable 'x'")
	}
	if pv, ok := value.(*state.PrimitiveValue); ok {
		if pv.IntValue == nil || *pv.IntValue != 10 {
			t.Errorf("expected value 10, got %v", pv.IntValue)
		}
	}

	// Child should have its own variable
	value, exists = child.GetVariable("y")
	if !exists {
		t.Error("child should have its own variable 'y'")
	}
	if pv, ok := value.(*state.PrimitiveValue); ok {
		if pv.IntValue == nil || *pv.IntValue != 20 {
			t.Errorf("expected value 20, got %v", pv.IntValue)
		}
	}

	// Parent should not see child's variable
	_, exists = parent.GetVariable("y")
	if exists {
		t.Error("parent should not see child's variable 'y'")
	}
}

func TestEnvironmentShadowing(t *testing.T) {
	parent := NewEnvironment()
	parent.SetVariable("x", state.NewIntValue(10))

	child := NewChildEnvironment(parent)
	child.SetVariable("x", state.NewIntValue(20)) // Shadow parent's x

	// Child should see its own x, not parent's
	value, exists := child.GetVariable("x")
	if !exists {
		t.Error("child should have variable 'x'")
	}
	if pv, ok := value.(*state.PrimitiveValue); ok {
		if pv.IntValue == nil || *pv.IntValue != 20 {
			t.Errorf("expected shadowed value 20, got %v", pv.IntValue)
		}
	}

	// Parent's x should be unchanged
	value, exists = parent.GetVariable("x")
	if !exists {
		t.Error("parent should still have variable 'x'")
	}
	if pv, ok := value.(*state.PrimitiveValue); ok {
		if pv.IntValue == nil || *pv.IntValue != 10 {
			t.Errorf("expected parent value 10, got %v", pv.IntValue)
		}
	}
}

func TestEnvironmentDefineFunction(t *testing.T) {
	env := NewEnvironment()

	// Define a function
	fnDef := &FunctionDef{
		Params: []ast.Parameter{},
		Body:   nil,
	}

	err := env.DefineFunction("add", fnDef)
	if err != nil {
		t.Errorf("unexpected error defining function: %v", err)
	}

	// Get the function
	fn, exists := env.GetFunction("add")
	if !exists {
		t.Error("function 'add' should exist")
	}
	if fn == nil {
		t.Error("function definition should not be nil")
	}

	// Try to define duplicate function
	err = env.DefineFunction("add", fnDef)
	if err == nil {
		t.Error("expected error when defining duplicate function")
	}
}

func TestEnvironmentNestedFunctionScopes(t *testing.T) {
	parent := NewEnvironment()
	parentFn := &FunctionDef{
		Params: []ast.Parameter{},
		Body:   nil,
	}
	parent.DefineFunction("parentFn", parentFn)

	child := NewChildEnvironment(parent)
	childFn := &FunctionDef{
		Params: []ast.Parameter{},
		Body:   nil,
	}
	child.DefineFunction("childFn", childFn)

	// Child should see parent's function
	fn, exists := child.GetFunction("parentFn")
	if !exists {
		t.Error("child should see parent's function")
	}
	if fn != parentFn {
		t.Error("child should get parent's function definition")
	}

	// Child should have its own function
	fn, exists = child.GetFunction("childFn")
	if !exists {
		t.Error("child should have its own function")
	}
	if fn != childFn {
		t.Error("child should get its own function definition")
	}

	// Parent should not see child's function
	_, exists = parent.GetFunction("childFn")
	if exists {
		t.Error("parent should not see child's function")
	}
}

func TestEnvironmentSetAndGetConstant(t *testing.T) {
	env := NewEnvironment()

	// Set a constant
	intVal := state.NewIntValue(100)
	env.SetConstant("MAX_VALUE", intVal)

	// Get the constant
	value, exists := env.GetConstant("MAX_VALUE")
	if !exists {
		t.Error("constant 'MAX_VALUE' should exist")
	}
	if value == nil {
		t.Error("value should not be nil")
	}

	// Verify it's the correct value
	if pv, ok := value.(*state.PrimitiveValue); ok {
		if pv.IntValue == nil || *pv.IntValue != 100 {
			t.Errorf("expected value 100, got %v", pv.IntValue)
		}
	}
}

func TestEnvironmentEnterExitScope(t *testing.T) {
	env := NewEnvironment()
	env.SetVariable("x", state.NewIntValue(10))

	// Enter a new scope
	child := env.EnterScope()
	child.SetVariable("y", state.NewIntValue(20))

	// Verify variables
	if !env.HasVariable("x") {
		t.Error("parent should have variable 'x'")
	}
	if env.HasVariable("y") {
		t.Error("parent should not have variable 'y'")
	}

	if !child.HasVariable("x") {
		t.Error("child should see parent's variable 'x'")
	}
	if !child.HasVariable("y") {
		t.Error("child should have variable 'y'")
	}

	// Exit scope
	parent := child.ExitScope()
	if parent != env {
		t.Error("ExitScope should return parent environment")
	}
}

func TestEnvironmentHasMethods(t *testing.T) {
	env := NewEnvironment()

	// Test HasVariable
	if env.HasVariable("x") {
		t.Error("should not have variable 'x' initially")
	}
	env.SetVariable("x", state.NewIntValue(10))
	if !env.HasVariable("x") {
		t.Error("should have variable 'x' after setting")
	}

	// Test HasFunction
	if env.HasFunction("add") {
		t.Error("should not have function 'add' initially")
	}
	fnDef := &FunctionDef{Params: []ast.Parameter{}, Body: nil}
	env.DefineFunction("add", fnDef)
	if !env.HasFunction("add") {
		t.Error("should have function 'add' after defining")
	}

	// Test HasConstant
	if env.HasConstant("MAX") {
		t.Error("should not have constant 'MAX' initially")
	}
	env.SetConstant("MAX", state.NewIntValue(100))
	if !env.HasConstant("MAX") {
		t.Error("should have constant 'MAX' after setting")
	}
}

