package types

import (
	"fmt"
)

// Environment represents a type environment for tracking variable and function types
// It supports nested scopes (e.g., for modules, functions, blocks)
type Environment struct {
	parent   *Environment
	variables map[string]Type
	functions map[string]*FunctionSignature
	constants map[string]Type
}

// FunctionSignature represents the type signature of a function
type FunctionSignature struct {
	Parameters []Type
	Return    Type
}

// NewEnvironment creates a new type environment
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Type),
		functions: make(map[string]*FunctionSignature),
		constants: make(map[string]Type),
	}
}

// NewChildEnvironment creates a new environment with the given parent
func NewChildEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:    parent,
		variables: make(map[string]Type),
		functions: make(map[string]*FunctionSignature),
		constants: make(map[string]Type),
	}
}

// DeclareVariable declares a variable with the given name and type
// Returns an error if the variable is already declared in the current scope
func (e *Environment) DeclareVariable(name string, typ Type) error {
	if _, exists := e.variables[name]; exists {
		return fmt.Errorf("variable %s already declared in this scope", name)
	}
	e.variables[name] = typ
	return nil
}

// LookupVariable looks up a variable's type, searching parent scopes if not found
func (e *Environment) LookupVariable(name string) (Type, bool) {
	if typ, exists := e.variables[name]; exists {
		return typ, true
	}
	if e.parent != nil {
		return e.parent.LookupVariable(name)
	}
	return nil, false
}

// DeclareFunction declares a function with the given name and signature
func (e *Environment) DeclareFunction(name string, sig *FunctionSignature) error {
	if _, exists := e.functions[name]; exists {
		return fmt.Errorf("function %s already declared in this scope", name)
	}
	e.functions[name] = sig
	return nil
}

// LookupFunction looks up a function's signature, searching parent scopes if not found
func (e *Environment) LookupFunction(name string) (*FunctionSignature, bool) {
	if sig, exists := e.functions[name]; exists {
		return sig, true
	}
	if e.parent != nil {
		return e.parent.LookupFunction(name)
	}
	return nil, false
}

// DeclareConstant declares a constant with the given name and type
func (e *Environment) DeclareConstant(name string, typ Type) error {
	if _, exists := e.constants[name]; exists {
		return fmt.Errorf("constant %s already declared in this scope", name)
	}
	e.constants[name] = typ
	return nil
}

// LookupConstant looks up a constant's type, searching parent scopes if not found
func (e *Environment) LookupConstant(name string) (Type, bool) {
	if typ, exists := e.constants[name]; exists {
		return typ, true
	}
	if e.parent != nil {
		return e.parent.LookupConstant(name)
	}
	return nil, false
}

// SetVariable sets a variable's type (for updates, allows overwriting)
func (e *Environment) SetVariable(name string, typ Type) {
	e.variables[name] = typ
}

// HasVariable checks if a variable exists in the current scope (not parent scopes)
func (e *Environment) HasVariable(name string) bool {
	_, exists := e.variables[name]
	return exists
}

// HasFunction checks if a function exists in the current scope (not parent scopes)
func (e *Environment) HasFunction(name string) bool {
	_, exists := e.functions[name]
	return exists
}

// HasConstant checks if a constant exists in the current scope (not parent scopes)
func (e *Environment) HasConstant(name string) bool {
	_, exists := e.constants[name]
	return exists
}

// GetParent returns the parent environment, or nil if this is the root
func (e *Environment) GetParent() *Environment {
	return e.parent
}

// IsRoot returns true if this is the root environment (no parent)
func (e *Environment) IsRoot() bool {
	return e.parent == nil
}

