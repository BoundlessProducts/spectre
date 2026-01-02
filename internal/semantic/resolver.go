package semantic

import (
	"fmt"

	"github.com/spectre-lang/spectre/pkg/ast"
)

// Resolver performs name resolution on an AST
type Resolver struct {
	symbolTable *SymbolTable
	currentScope *Scope
	errors      []string
}

// NewResolver creates a new resolver
func NewResolver(symbolTable *SymbolTable) *Resolver {
	return &Resolver{
		symbolTable:  symbolTable,
		currentScope: symbolTable.GlobalScope,
		errors:       []string{},
	}
}

// ResolveFile resolves all names in a file
func (r *Resolver) ResolveFile(file *ast.File) []string {
	r.errors = []string{}
	r.currentScope = r.symbolTable.GlobalScope

	for _, decl := range file.Decls {
		r.resolveDecl(decl)
	}

	return r.errors
}

// resolveDecl resolves a declaration
func (r *Resolver) resolveDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.VariableDecl:
		r.resolveVariableDecl(d)
	case *ast.ConstantDecl:
		r.resolveConstantDecl(d)
	case *ast.FunctionDecl:
		r.resolveFunctionDecl(d)
	case *ast.ActionDecl:
		r.resolveActionDecl(d)
	case *ast.InitDecl:
		r.resolveInitDecl(d)
	case *ast.OneOfInitDecl:
		r.resolveOneOfInitDecl(d)
	case *ast.InvariantDecl:
		r.resolveInvariantDecl(d)
	case *ast.TemporalDecl:
		r.resolveTemporalDecl(d)
	case *ast.ModuleDecl:
		r.resolveModuleDecl(d)
	case *ast.ImportDecl:
		// Imports are handled separately in module resolution
		return
	case *ast.ModuleInstanceDecl:
		// Module instances are handled separately
		return
	default:
		// Unknown declaration type
		return
	}
}

// resolveVariableDecl resolves a variable declaration
func (r *Resolver) resolveVariableDecl(decl *ast.VariableDecl) {
	// Variable is already in symbol table from building phase
	// Just resolve its type if it's a named type
	r.resolveType(decl.Type)
}

// resolveConstantDecl resolves a constant declaration
func (r *Resolver) resolveConstantDecl(decl *ast.ConstantDecl) {
	// Resolve type
	r.resolveType(decl.Type)
	// Resolve value expression
	r.resolveExpression(decl.Value, r.currentScope)
}

// resolveFunctionDecl resolves a function declaration
func (r *Resolver) resolveFunctionDecl(decl *ast.FunctionDecl) {
	// Resolve return type
	r.resolveType(decl.ReturnType)

	// Look up the function scope (it was created during building)
	// We need to find the scope that matches this function
	funcScope := r.findFunctionScope(decl.Name)
	if funcScope == nil {
		// Fallback: create a new scope
		funcScope = r.symbolTable.NewScope(r.currentScope, ScopeFunction, decl.Name)
	}

	oldScope := r.currentScope
	r.currentScope = funcScope

	// Resolve parameters (they're already in symbol table from building phase)
	for _, param := range decl.Parameters {
		r.resolveType(param.Type)
	}

	// Resolve function body
	if decl.Body != nil {
		r.resolveBlock(decl.Body, funcScope)
	}

	r.currentScope = oldScope
}

// findFunctionScope finds the scope for a function by name
func (r *Resolver) findFunctionScope(funcName string) *Scope {
	// Search through all scopes to find the function scope
	for _, scope := range r.symbolTable.scopes {
		if scope.Kind == ScopeFunction && scope.Name == funcName {
			return scope
		}
	}
	return nil
}

// resolveActionDecl resolves an action declaration
func (r *Resolver) resolveActionDecl(decl *ast.ActionDecl) {
	// Look up the action scope (it was created during building)
	actionScope := r.findActionScope(decl.Name)
	if actionScope == nil {
		// Fallback: create a new scope
		actionScope = r.symbolTable.NewScope(r.currentScope, ScopeAction, decl.Name)
	}

	oldScope := r.currentScope
	r.currentScope = actionScope

	// Resolve parameters (they're already in symbol table from building phase)
	for _, param := range decl.Parameters {
		r.resolveType(param.Type)
	}

	// Resolve guard condition
	if decl.Guard != nil {
		r.resolveExpression(decl.Guard, actionScope)
	}

	// Resolve action body
	if decl.Body != nil {
		r.resolveBlock(decl.Body, actionScope)
	}

	r.currentScope = oldScope
}

// findActionScope finds the scope for an action by name
func (r *Resolver) findActionScope(actionName string) *Scope {
	// Search through all scopes to find the action scope
	for _, scope := range r.symbolTable.scopes {
		if scope.Kind == ScopeAction && scope.Name == actionName {
			return scope
		}
	}
	return nil
}

// resolveInitDecl resolves an init declaration
func (r *Resolver) resolveInitDecl(decl *ast.InitDecl) {
	if decl.Body != nil {
		// Use current scope (could be module scope or global scope)
		r.resolveBlock(decl.Body, r.currentScope)
	}
}

// resolveOneOfInitDecl resolves a oneOf init declaration
func (r *Resolver) resolveOneOfInitDecl(decl *ast.OneOfInitDecl) {
	for _, option := range decl.Options {
		if option != nil {
			r.resolveBlock(option, r.currentScope)
		}
	}
}

// resolveInvariantDecl resolves an invariant declaration
func (r *Resolver) resolveInvariantDecl(decl *ast.InvariantDecl) {
	if decl.Condition != nil {
		r.resolveExpression(decl.Condition, r.currentScope)
	}
}

// resolveTemporalDecl resolves a temporal declaration
func (r *Resolver) resolveTemporalDecl(decl *ast.TemporalDecl) {
	if decl.Expression != nil {
		r.resolveExpression(decl.Expression, r.currentScope)
	}
}

// resolveModuleDecl resolves a module declaration
func (r *Resolver) resolveModuleDecl(decl *ast.ModuleDecl) {
	// Find the existing module scope (created during building)
	moduleScope := r.findModuleScope(decl.Name)
	if moduleScope == nil {
		// Fallback: create a new scope
		moduleScope = r.symbolTable.NewScope(r.currentScope, ScopeModule, decl.Name)
	}

	oldScope := r.currentScope
	r.currentScope = moduleScope

	// Resolve all declarations within the module
	for _, innerDecl := range decl.Decls {
		r.resolveDecl(innerDecl)
	}

	r.currentScope = oldScope
}

// findModuleScope finds the scope for a module by name
func (r *Resolver) findModuleScope(moduleName string) *Scope {
	for _, scope := range r.symbolTable.scopes {
		if scope.Kind == ScopeModule && scope.Name == moduleName {
			return scope
		}
	}
	return nil
}

// resolveBlock resolves a block statement
func (r *Resolver) resolveBlock(block *ast.BlockStmt, scope *Scope) {
	if block == nil {
		return
	}

	oldScope := r.currentScope
	r.currentScope = scope

	for _, stmt := range block.Statements {
		r.resolveStatement(stmt, scope)
	}

	r.currentScope = oldScope
}

// resolveStatement resolves a statement
func (r *Resolver) resolveStatement(stmt ast.Stmt, scope *Scope) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		r.resolveExpression(s.Expr, scope)
	case *ast.ReturnStmt:
		if s.Value != nil {
			r.resolveExpression(s.Value, scope)
		}
	case *ast.AssignStmt:
		r.resolveAssignStatement(s, scope)
	case *ast.RequireStmt:
		if s.Condition != nil {
			r.resolveExpression(s.Condition, scope)
		}
	case *ast.EnsureStmt:
		if s.Condition != nil {
			r.resolveExpression(s.Condition, scope)
		}
	default:
		// Unknown statement type
		return
	}
}

// resolveAssignStatement resolves an assignment statement
func (r *Resolver) resolveAssignStatement(stmt *ast.AssignStmt, scope *Scope) {
	// Resolve left-hand side (L-value)
	r.resolveLValue(stmt.Left, scope)
	// Resolve right-hand side (R-value)
	r.resolveExpression(stmt.Right, scope)
}

// resolveLValue resolves an L-value (assignment target)
func (r *Resolver) resolveLValue(expr ast.Expr, scope *Scope) {
	switch e := expr.(type) {
	case *ast.Ident:
		r.resolveIdentifier(e, scope, true)
	case *ast.SelectorExpr:
		// Resolve the record/object
		r.resolveExpression(e.X, scope)
		// Field name doesn't need resolution (it's a literal)
	case *ast.IndexExpr:
		// Resolve the collection
		r.resolveExpression(e.X, scope)
		// Resolve the index
		r.resolveExpression(e.Index, scope)
	default:
		r.addError(expr.Pos(), "invalid assignment target")
	}
}

// resolveExpression resolves an expression
func (r *Resolver) resolveExpression(expr ast.Expr, scope *Scope) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.BasicLit:
		// Literals don't need resolution
		return
	case *ast.Ident:
		r.resolveIdentifier(e, scope, false)
	case *ast.UnaryExpr:
		r.resolveExpression(e.Expr, scope)
	case *ast.BinaryExpr:
		r.resolveExpression(e.Left, scope)
		r.resolveExpression(e.Right, scope)
	case *ast.ParenExpr:
		r.resolveExpression(e.X, scope)
	case *ast.CallExpr:
		r.resolveCallExpression(e, scope)
	case *ast.IndexExpr:
		r.resolveExpression(e.X, scope)
		r.resolveExpression(e.Index, scope)
	case *ast.SelectorExpr:
		r.resolveSelectorExpression(e, scope)
	case *ast.IfExpr:
		r.resolveIfExpression(e, scope)
	case *ast.LetExpr:
		r.resolveLetExpression(e, scope)
	case *ast.AlwaysExpr:
		r.resolveExpression(e.Expr, scope)
	case *ast.EventuallyExpr:
		r.resolveExpression(e.Expr, scope)
	case *ast.UntilExpr:
		r.resolveExpression(e.Left, scope)
		r.resolveExpression(e.Right, scope)
	case *ast.LeadsToExpr:
		r.resolveExpression(e.Left, scope)
		r.resolveExpression(e.Right, scope)
	case *ast.WFExpr:
		r.resolveExpression(e.Target, scope)
	case *ast.SFExpr:
		r.resolveExpression(e.Target, scope)
	case *ast.SuperExpr:
		// Super expressions are resolved in module context
		// For now, just verify the method exists
		return
	default:
		// Unknown expression type
		return
	}
}

// resolveIdentifier resolves an identifier
func (r *Resolver) resolveIdentifier(ident *ast.Ident, scope *Scope, isLValue bool) {
	name := ident.Name

	// Handle primed variables (next-state variables)
	if ident.Prime {
		// Primed variable refers to the next state of a variable
		// Look up the base variable name
		symbol, found := r.symbolTable.LookupSymbol(scope, name)
		if !found {
			r.addError(ident.Pos(), "undefined variable: %s (primed variable must reference existing variable)", name)
			return
		}
		if symbol.Kind != SymbolVariable {
			r.addError(ident.Pos(), "cannot prime non-variable: %s", name)
			return
		}
		return
	}

	// Regular identifier resolution
	symbol, found := r.symbolTable.LookupSymbol(scope, name)
	if !found {
		r.addError(ident.Pos(), "undefined identifier: %s", name)
		return
	}

	// Check if it's an L-value and if assignment is allowed
	if isLValue {
		if symbol.Kind == SymbolConstant {
			r.addError(ident.Pos(), "cannot assign to constant: %s", name)
			return
		}
		if symbol.Kind != SymbolVariable {
			r.addError(ident.Pos(), "cannot assign to %s: %s", symbolKindString(symbol.Kind), name)
			return
		}
	}
}

// resolveCallExpression resolves a function call expression
func (r *Resolver) resolveCallExpression(expr *ast.CallExpr, scope *Scope) {
	// Resolve function name
	switch fun := expr.Fun.(type) {
	case *ast.Ident:
		// Direct function call: funcName(args)
		r.resolveIdentifier(fun, scope, false)
		// Verify it's a function
		symbol, found := r.symbolTable.LookupSymbol(scope, fun.Name)
		if found && symbol.Kind != SymbolFunction {
			r.addError(fun.Pos(), "cannot call non-function: %s", fun.Name)
		}
	case *ast.SelectorExpr:
		// Method call: obj.method(args)
		r.resolveSelectorExpression(fun, scope)
		// For now, we'll handle method calls later
	default:
		r.addError(expr.Pos(), "invalid function call expression")
	}

	// Resolve arguments
	for _, arg := range expr.Args {
		r.resolveExpression(arg, scope)
	}
}

// resolveSelectorExpression resolves a selector expression (field/method access)
func (r *Resolver) resolveSelectorExpression(expr *ast.SelectorExpr, scope *Scope) {
	// Resolve the object/record
	r.resolveExpression(expr.X, scope)
	// Field/method name doesn't need resolution (it's a literal)
}

// resolveIfExpression resolves an if expression
func (r *Resolver) resolveIfExpression(expr *ast.IfExpr, scope *Scope) {
	r.resolveExpression(expr.Condition, scope)
	// IfExpr.Then and IfExpr.Else are Expr, not BlockStmt
	// They can be expressions or blocks, so we resolve them as expressions
	r.resolveExpression(expr.Then, scope)
	if expr.Else != nil {
		r.resolveExpression(expr.Else, scope)
	}
}

// resolveLetExpression resolves a let expression
func (r *Resolver) resolveLetExpression(expr *ast.LetExpr, scope *Scope) {
	// Create a new scope for the let binding
	letScope := r.symbolTable.NewScope(scope, ScopeBlock, "let")
	oldScope := r.currentScope
	r.currentScope = letScope

	// Resolve the binding expression (in parent scope)
	r.resolveExpression(expr.Value, scope)

	// The binding variable should be in the symbol table from building phase
	// Resolve the body expression in the new scope (can reference the binding)
	r.resolveExpression(expr.Body, letScope)

	r.currentScope = oldScope
}

// resolveType resolves a type (for named types)
func (r *Resolver) resolveType(typ ast.Type) {
	switch t := typ.(type) {
	case *ast.NamedType:
		// Named types need to be resolved
		// For now, we'll handle this in a later phase
		return
	case *ast.RecordType:
		// Resolve field types
		for _, field := range t.Fields {
			r.resolveType(field.Type)
		}
	case *ast.SetType:
		r.resolveType(t.Element)
	case *ast.MapType:
		r.resolveType(t.Key)
		r.resolveType(t.Value)
	case *ast.ListType:
		r.resolveType(t.Element)
	case *ast.OptionType:
		r.resolveType(t.Element)
	case *ast.EnumType:
		// Enum types don't need resolution
		return
	case *ast.PrimitiveType:
		// Primitive types don't need resolution
		return
	default:
		// Unknown type
		return
	}
}

// addError adds an error message
func (r *Resolver) addError(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if pos.Line > 0 {
		r.errors = append(r.errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		r.errors = append(r.errors, msg)
	}
}

// symbolKindString returns a string representation of a symbol kind
func symbolKindString(kind SymbolKind) string {
	switch kind {
	case SymbolVariable:
		return "variable"
	case SymbolConstant:
		return "constant"
	case SymbolFunction:
		return "function"
	case SymbolAction:
		return "action"
	case SymbolModule:
		return "module"
	case SymbolType:
		return "type"
	default:
		return "unknown"
	}
}

