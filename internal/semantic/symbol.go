package semantic

import (
	"fmt"

	"github.com/spectre-lang/spectre/pkg/ast"
)

// SymbolKind represents the kind of symbol
type SymbolKind int

const (
	SymbolVariable SymbolKind = iota
	SymbolConstant
	SymbolFunction
	SymbolAction
	SymbolModule
	SymbolType
)

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name        string
	Kind        SymbolKind
	Decl        ast.Decl // The declaration that defines this symbol
	Scope       *Scope   // The scope where this symbol is defined
	Position    ast.Position
	Description string // Description from the declaration
}

// Scope represents a scope in the symbol table
type Scope struct {
	Parent      *Scope
	Symbols     map[string]*Symbol // Symbol name -> Symbol
	Kind        ScopeKind
	Name        string // Scope name (module name, function name, etc.)
	Position    ast.Position
}

// ScopeKind represents the kind of scope
type ScopeKind int

const (
	ScopeGlobal ScopeKind = iota
	ScopeModule
	ScopeFunction
	ScopeAction
	ScopeBlock
)

// SymbolTable represents the symbol table for a Spectre program
type SymbolTable struct {
	GlobalScope *Scope
	scopes      []*Scope // All scopes for cleanup/reference
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	globalScope := &Scope{
		Parent:  nil,
		Symbols: make(map[string]*Symbol),
		Kind:    ScopeGlobal,
		Name:    "global",
	}
	return &SymbolTable{
		GlobalScope: globalScope,
		scopes:      []*Scope{globalScope},
	}
}

// NewScope creates a new scope with the given parent
func (st *SymbolTable) NewScope(parent *Scope, kind ScopeKind, name string) *Scope {
	scope := &Scope{
		Parent:   parent,
		Symbols:  make(map[string]*Symbol),
		Kind:     kind,
		Name:     name,
		Position: ast.Position{},
	}
	st.scopes = append(st.scopes, scope)
	return scope
}

// DefineSymbol defines a symbol in the given scope
func (st *SymbolTable) DefineSymbol(scope *Scope, name string, kind SymbolKind, decl ast.Decl) error {
	if scope == nil {
		scope = st.GlobalScope
	}

	// Check if symbol already exists in this scope
	if _, exists := scope.Symbols[name]; exists {
		return fmt.Errorf("symbol %s already defined in scope %s", name, scope.Name)
	}

	// Extract description from declaration
	description := ""
	switch d := decl.(type) {
	case *ast.VariableDecl:
		description = d.Description
	case *ast.ConstantDecl:
		description = d.Description
	case *ast.FunctionDecl:
		description = d.Description
	case *ast.ActionDecl:
		description = d.Description
	case *ast.InvariantDecl:
		description = d.Description
	case *ast.TemporalDecl:
		description = d.Description
	case *ast.ModuleDecl:
		description = d.Description
	}

	symbol := &Symbol{
		Name:        name,
		Kind:        kind,
		Decl:        decl,
		Scope:       scope,
		Position:    decl.Pos(),
		Description: description,
	}

	scope.Symbols[name] = symbol
	return nil
}

// LookupSymbol looks up a symbol starting from the given scope and searching parent scopes
func (st *SymbolTable) LookupSymbol(scope *Scope, name string) (*Symbol, bool) {
	current := scope
	if current == nil {
		current = st.GlobalScope
	}

	for current != nil {
		if symbol, exists := current.Symbols[name]; exists {
			return symbol, true
		}
		current = current.Parent
	}

	return nil, false
}

// LookupSymbolInScope looks up a symbol only in the given scope (not parent scopes)
func (st *SymbolTable) LookupSymbolInScope(scope *Scope, name string) (*Symbol, bool) {
	if scope == nil {
		scope = st.GlobalScope
	}
	symbol, exists := scope.Symbols[name]
	return symbol, exists
}

// GetScope returns the scope at the given index (for testing/debugging)
func (st *SymbolTable) GetScope(index int) *Scope {
	if index < 0 || index >= len(st.scopes) {
		return nil
	}
	return st.scopes[index]
}

// ScopeCount returns the total number of scopes
func (st *SymbolTable) ScopeCount() int {
	return len(st.scopes)
}

