package eval

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// Environment represents the evaluation environment for pure functions
// It manages variable bindings, function definitions, and constants
type Environment struct {
	parent     *Environment
	variables  map[string]state.Value // Variable name -> value
	functions  map[string]*FunctionDef // Function name -> function definition
	constants  map[string]state.Value // Constant name -> value
}

// FunctionDef represents a function definition in the evaluation environment
type FunctionDef struct {
	Decl     *ast.FunctionDecl
	Params   []ast.Parameter
	Body     *ast.BlockStmt
	ReturnType ast.Type
}

// NewEnvironment creates a new evaluation environment
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]state.Value),
		functions: make(map[string]*FunctionDef),
		constants: make(map[string]state.Value),
	}
}

// NewChildEnvironment creates a new environment with the given parent
func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:    parent,
		variables: make(map[string]state.Value),
		functions: make(map[string]*FunctionDef),
		constants: make(map[string]state.Value),
	}
}

// SetVariable sets a variable value in the current environment
func (e *Environment) SetVariable(name string, value state.Value) {
	e.variables[name] = value
}

// GetVariable gets a variable value, searching parent environments if not found
func (e *Environment) GetVariable(name string) (state.Value, bool) {
	if value, exists := e.variables[name]; exists {
		return value, true
	}
	if e.parent != nil {
		return e.parent.GetVariable(name)
	}
	return nil, false
}

// DefineFunction defines a function in the environment
func (e *Environment) DefineFunction(name string, def *FunctionDef) error {
	if _, exists := e.functions[name]; exists {
		return fmt.Errorf("function %s already defined", name)
	}
	e.functions[name] = def
	return nil
}

// GetFunction gets a function definition, searching parent environments if not found
func (e *Environment) GetFunction(name string) (*FunctionDef, bool) {
	if fn, exists := e.functions[name]; exists {
		return fn, true
	}
	if e.parent != nil {
		return e.parent.GetFunction(name)
	}
	return nil, false
}

// SetConstant sets a constant value in the environment
func (e *Environment) SetConstant(name string, value state.Value) {
	e.constants[name] = value
}

// GetConstant gets a constant value, searching parent environments if not found
func (e *Environment) GetConstant(name string) (state.Value, bool) {
	if value, exists := e.constants[name]; exists {
		return value, true
	}
	if e.parent != nil {
		return e.parent.GetConstant(name)
	}
	return nil, false
}

// EnterScope creates a new child environment (enters a new scope)
func (e *Environment) EnterScope() *Environment {
	return NewChildEnvironment(e)
}

// ExitScope returns the parent environment (exits current scope)
func (e *Environment) ExitScope() *Environment {
	if e.parent == nil {
		return e // Already at root
	}
	return e.parent
}

// HasVariable checks if a variable exists in the environment
func (e *Environment) HasVariable(name string) bool {
	_, exists := e.GetVariable(name)
	return exists
}

// HasFunction checks if a function exists in the environment
func (e *Environment) HasFunction(name string) bool {
	_, exists := e.GetFunction(name)
	return exists
}

// HasConstant checks if a constant exists in the environment
func (e *Environment) HasConstant(name string) bool {
	_, exists := e.GetConstant(name)
	return exists
}

