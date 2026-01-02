package semantic

import (
	"fmt"

	"github.com/spectre-lang/spectre/pkg/ast"
)

// Builder builds the symbol table from an AST
type Builder struct {
	symbolTable *SymbolTable
	currentScope *Scope
	errors      []string
}

// NewBuilder creates a new symbol table builder
func NewBuilder() *Builder {
	st := NewSymbolTable()
	return &Builder{
		symbolTable:  st,
		currentScope: st.GlobalScope,
		errors:       []string{},
	}
}

// BuildSymbolTable builds the symbol table from a file AST
func (b *Builder) BuildSymbolTable(file *ast.File) (*SymbolTable, []string) {
	b.errors = []string{}
	b.currentScope = b.symbolTable.GlobalScope

	for _, decl := range file.Decls {
		b.buildDecl(decl)
	}

	return b.symbolTable, b.errors
}

// buildDecl builds symbols for a declaration
func (b *Builder) buildDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.VariableDecl:
		b.buildVariableDecl(d)
	case *ast.ConstantDecl:
		b.buildConstantDecl(d)
	case *ast.FunctionDecl:
		b.buildFunctionDecl(d)
	case *ast.ActionDecl:
		b.buildActionDecl(d)
	case *ast.InitDecl:
		b.buildInitDecl(d)
	case *ast.OneOfInitDecl:
		b.buildOneOfInitDecl(d)
	case *ast.InvariantDecl:
		b.buildInvariantDecl(d)
	case *ast.TemporalDecl:
		b.buildTemporalDecl(d)
	case *ast.ModuleDecl:
		b.buildModuleDecl(d)
	case *ast.ImportDecl:
		// Imports are handled separately
		return
	case *ast.ModuleInstanceDecl:
		// Module instances are handled separately
		return
	default:
		return
	}
}

// buildVariableDecl builds a variable declaration symbol
func (b *Builder) buildVariableDecl(decl *ast.VariableDecl) {
	err := b.symbolTable.DefineSymbol(b.currentScope, decl.Name, SymbolVariable, decl)
	if err != nil {
		b.addError(decl.Pos(), "%s", err.Error())
	}
}

// buildConstantDecl builds a constant declaration symbol
func (b *Builder) buildConstantDecl(decl *ast.ConstantDecl) {
	err := b.symbolTable.DefineSymbol(b.currentScope, decl.Name, SymbolConstant, decl)
	if err != nil {
		b.addError(decl.Pos(), "%s", err.Error())
	}
}

// buildFunctionDecl builds a function declaration symbol
func (b *Builder) buildFunctionDecl(decl *ast.FunctionDecl) {
	// Define function symbol in current scope
	err := b.symbolTable.DefineSymbol(b.currentScope, decl.Name, SymbolFunction, decl)
	if err != nil {
		b.addError(decl.Pos(), "%s", err.Error())
		return
	}

	// Create function scope and define parameters
	funcScope := b.symbolTable.NewScope(b.currentScope, ScopeFunction, decl.Name)
	oldScope := b.currentScope
	b.currentScope = funcScope

	for _, param := range decl.Parameters {
		// Parameters are variables in the function scope
		paramVar := &ast.VariableDecl{
			Position: param.Position,
			Name:     param.Name,
			Type:     param.Type,
		}
		err := b.symbolTable.DefineSymbol(funcScope, param.Name, SymbolVariable, paramVar)
		if err != nil {
			b.addError(param.Position, "%s", err.Error())
		}
	}

	b.currentScope = oldScope
}

// buildActionDecl builds an action declaration symbol
func (b *Builder) buildActionDecl(decl *ast.ActionDecl) {
	// Define action symbol in current scope
	err := b.symbolTable.DefineSymbol(b.currentScope, decl.Name, SymbolAction, decl)
	if err != nil {
		b.addError(decl.Pos(), "%s", err.Error())
		return
	}

	// Create action scope and define parameters
	actionScope := b.symbolTable.NewScope(b.currentScope, ScopeAction, decl.Name)
	oldScope := b.currentScope
	b.currentScope = actionScope

	for _, param := range decl.Parameters {
		// Parameters are variables in the action scope
		paramVar := &ast.VariableDecl{
			Position: param.Position,
			Name:     param.Name,
			Type:     param.Type,
		}
		err := b.symbolTable.DefineSymbol(actionScope, param.Name, SymbolVariable, paramVar)
		if err != nil {
			b.addError(param.Position, "%s", err.Error())
		}
	}

	b.currentScope = oldScope
}

// buildInitDecl builds an init declaration (no symbol needed, just scope)
func (b *Builder) buildInitDecl(decl *ast.InitDecl) {
	// Init declarations don't create symbols, but they may use variables
	// Variables are already in the global scope
}

// buildOneOfInitDecl builds a oneOf init declaration
func (b *Builder) buildOneOfInitDecl(decl *ast.OneOfInitDecl) {
	// OneOf init declarations don't create symbols
}

// buildInvariantDecl builds an invariant declaration symbol
func (b *Builder) buildInvariantDecl(decl *ast.InvariantDecl) {
	// Invariants don't create symbols, but they reference variables
}

// buildTemporalDecl builds a temporal declaration symbol
func (b *Builder) buildTemporalDecl(decl *ast.TemporalDecl) {
	// Temporal declarations don't create symbols, but they reference variables
}

// buildModuleDecl builds a module declaration symbol
func (b *Builder) buildModuleDecl(decl *ast.ModuleDecl) {
	// Define module symbol in current scope
	err := b.symbolTable.DefineSymbol(b.currentScope, decl.Name, SymbolModule, decl)
	if err != nil {
		b.addError(decl.Pos(), "%s", err.Error())
		return
	}

	// Create module scope
	moduleScope := b.symbolTable.NewScope(b.currentScope, ScopeModule, decl.Name)
	oldScope := b.currentScope
	b.currentScope = moduleScope

	// Build all declarations within the module
	for _, innerDecl := range decl.Decls {
		b.buildDecl(innerDecl)
	}

	b.currentScope = oldScope
}

// addError adds an error message
func (b *Builder) addError(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if pos.Line > 0 {
		b.errors = append(b.errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		b.errors = append(b.errors, msg)
	}
}

