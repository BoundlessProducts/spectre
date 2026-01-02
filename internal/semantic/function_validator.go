package semantic

import (
	"fmt"

	"github.com/akkeshavan/spectre/pkg/ast"
)

// FunctionValidator performs function-specific validation
type FunctionValidator struct {
	symbolTable *SymbolTable
	errors      []string
}

// NewFunctionValidator creates a new function validator
func NewFunctionValidator(symbolTable *SymbolTable) *FunctionValidator {
	return &FunctionValidator{
		symbolTable: symbolTable,
		errors:      []string{},
	}
}

// ValidateFunctionCalls validates all function calls in a file
func (fv *FunctionValidator) ValidateFunctionCalls(file *ast.File) []string {
	fv.errors = []string{}

	for _, decl := range file.Decls {
		fv.validateDeclForFunctionCalls(decl)
	}

	return fv.errors
}

// validateDeclForFunctionCalls validates function calls within a declaration
func (fv *FunctionValidator) validateDeclForFunctionCalls(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.FunctionDecl:
		fv.validateFunctionBodyCalls(d.Body, fv.findFunctionScope(d.Name))
	case *ast.ActionDecl:
		fv.validateActionBodyCalls(d)
	case *ast.InitDecl:
		if d.Body != nil {
			fv.validateBlockCalls(d.Body, fv.symbolTable.GlobalScope)
		}
		if d.Expression != nil {
			fv.validateExpressionCalls(d.Expression, fv.symbolTable.GlobalScope)
		}
	case *ast.OneOfInitDecl:
		for _, option := range d.Options {
			if option != nil {
				fv.validateBlockCalls(option, fv.symbolTable.GlobalScope)
			}
		}
	case *ast.InvariantDecl:
		if d.Condition != nil {
			fv.validateExpressionCalls(d.Condition, fv.symbolTable.GlobalScope)
		}
	case *ast.TemporalDecl:
		if d.Expression != nil {
			fv.validateExpressionCalls(d.Expression, fv.symbolTable.GlobalScope)
		}
	case *ast.ModuleDecl:
		for _, innerDecl := range d.Decls {
			fv.validateDeclForFunctionCalls(innerDecl)
		}
	default:
		return
	}
}

// validateFunctionBodyCalls validates function calls in a function body
func (fv *FunctionValidator) validateFunctionBodyCalls(body *ast.BlockStmt, scope *Scope) {
	if body == nil {
		return
	}

	for _, stmt := range body.Statements {
		fv.validateStatementCalls(stmt, scope)
	}
}

// validateActionBodyCalls validates function calls in an action body
func (fv *FunctionValidator) validateActionBodyCalls(action *ast.ActionDecl) {
	scope := fv.findActionScope(action.Name)
	if scope == nil {
		scope = fv.symbolTable.GlobalScope
	}

	if action.Guard != nil {
		fv.validateExpressionCalls(action.Guard, scope)
	}

	if action.Body != nil {
		fv.validateBlockCalls(action.Body, scope)
	}
}

// validateBlockCalls validates function calls in a block
func (fv *FunctionValidator) validateBlockCalls(block *ast.BlockStmt, scope *Scope) {
	if block == nil {
		return
	}

	for _, stmt := range block.Statements {
		fv.validateStatementCalls(stmt, scope)
	}
}

// validateStatementCalls validates function calls in a statement
func (fv *FunctionValidator) validateStatementCalls(stmt ast.Stmt, scope *Scope) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		fv.validateExpressionCalls(s.Expr, scope)
	case *ast.ReturnStmt:
		if s.Value != nil {
			fv.validateExpressionCalls(s.Value, scope)
		}
	case *ast.AssignStmt:
		fv.validateExpressionCalls(s.Left, scope)
		fv.validateExpressionCalls(s.Right, scope)
	case *ast.RequireStmt:
		if s.Condition != nil {
			fv.validateExpressionCalls(s.Condition, scope)
		}
	case *ast.EnsureStmt:
		if s.Condition != nil {
			fv.validateExpressionCalls(s.Condition, scope)
		}
	default:
		return
	}
}

// validateExpressionCalls validates function calls in an expression
func (fv *FunctionValidator) validateExpressionCalls(expr ast.Expr, scope *Scope) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.BasicLit:
		// Literals don't contain function calls
		return
	case *ast.Ident:
		// Identifiers don't contain function calls
		return
	case *ast.UnaryExpr:
		fv.validateExpressionCalls(e.Expr, scope)
	case *ast.BinaryExpr:
		fv.validateExpressionCalls(e.Left, scope)
		fv.validateExpressionCalls(e.Right, scope)
	case *ast.ParenExpr:
		fv.validateExpressionCalls(e.X, scope)
	case *ast.CallExpr:
		fv.validateCallExpression(e, scope)
	case *ast.IndexExpr:
		fv.validateExpressionCalls(e.X, scope)
		fv.validateExpressionCalls(e.Index, scope)
	case *ast.SelectorExpr:
		fv.validateExpressionCalls(e.X, scope)
		// Method calls are handled as CallExpr with SelectorExpr as Fun
	case *ast.IfExpr:
		fv.validateExpressionCalls(e.Condition, scope)
		fv.validateExpressionCalls(e.Then, scope)
		if e.Else != nil {
			fv.validateExpressionCalls(e.Else, scope)
		}
	case *ast.LetExpr:
		fv.validateExpressionCalls(e.Value, scope)
		letScope := fv.findLetScope(e.Name, scope)
		if letScope == nil {
			letScope = scope
		}
		fv.validateExpressionCalls(e.Body, letScope)
	case *ast.AlwaysExpr:
		fv.validateExpressionCalls(e.Expr, scope)
	case *ast.EventuallyExpr:
		fv.validateExpressionCalls(e.Expr, scope)
	case *ast.UntilExpr:
		fv.validateExpressionCalls(e.Left, scope)
		fv.validateExpressionCalls(e.Right, scope)
	case *ast.LeadsToExpr:
		fv.validateExpressionCalls(e.Left, scope)
		fv.validateExpressionCalls(e.Right, scope)
	case *ast.WFExpr:
		fv.validateExpressionCalls(e.Target, scope)
	case *ast.SFExpr:
		fv.validateExpressionCalls(e.Target, scope)
	default:
		return
	}
}

// validateCallExpression validates a function call expression
func (fv *FunctionValidator) validateCallExpression(expr *ast.CallExpr, scope *Scope) {
	// Determine function name and get its declaration
	var funcName string
	var funcDecl *ast.FunctionDecl

	switch fun := expr.Fun.(type) {
	case *ast.Ident:
		funcName = fun.Name
		symbol, found := fv.symbolTable.LookupSymbol(scope, funcName)
		if !found {
			fv.addError(fun.Pos(), "undefined function: %s", funcName)
			return
		}
		if symbol.Kind != SymbolFunction {
			fv.addError(fun.Pos(), "cannot call non-function: %s", funcName)
			return
		}
		// Get the function declaration
		if funcDeclAST, ok := symbol.Decl.(*ast.FunctionDecl); ok {
			funcDecl = funcDeclAST
		} else {
			fv.addError(fun.Pos(), "invalid function declaration for: %s", funcName)
			return
		}
	case *ast.SelectorExpr:
		// Method calls - for now, we'll validate the object
		fv.validateExpressionCalls(fun.X, scope)
		// Method call validation will be enhanced in later phases
		return
	default:
		fv.addError(expr.Pos(), "invalid function call expression")
		return
	}

	if funcDecl == nil {
		return
	}

	// Validate argument count
	expectedArgCount := len(funcDecl.Parameters)
	actualArgCount := len(expr.Args)
	if actualArgCount != expectedArgCount {
		fv.addError(expr.Pos(), "function %s expects %d arguments, got %d",
			funcName, expectedArgCount, actualArgCount)
		return
	}

	// Validate argument types (basic check - full type checking is done in type checker)
	// For now, we just ensure the arguments are valid expressions
	for _, arg := range expr.Args {
		fv.validateExpressionCalls(arg, scope)
		// Type checking will be done by the type checker
		// Here we just ensure the call structure is correct
	}

	// Validate that function calls are pure (no side effects in arguments)
	// This is already enforced by the parser/type system, but we can add checks here
}

// validateFunctionPurity validates that a function is pure (no state mutations)
func (fv *FunctionValidator) validateFunctionPurity(funcDecl *ast.FunctionDecl) {
	if funcDecl.Body == nil {
		return
	}

	scope := fv.findFunctionScope(funcDecl.Name)
	if scope == nil {
		scope = fv.symbolTable.GlobalScope
	}

	for _, stmt := range funcDecl.Body.Statements {
		fv.checkStatementPurity(stmt, funcDecl.Name, scope)
	}
}

// checkStatementPurity checks if a statement violates function purity
func (fv *FunctionValidator) checkStatementPurity(stmt ast.Stmt, funcName string, scope *Scope) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		// Check if assigning to a state variable
		if fv.isStateVariable(s.Left) {
			fv.addError(s.Pos(), "function %s cannot mutate state variable", funcName)
		}
	case *ast.ReturnStmt:
		// Return statements are allowed
		if s.Value != nil {
			fv.validateExpressionCalls(s.Value, scope)
		}
	case *ast.ExprStmt:
		// Expression statements are allowed (as long as they don't mutate state)
		fv.validateExpressionCalls(s.Expr, scope)
	case *ast.RequireStmt:
		// Require statements are allowed
		if s.Condition != nil {
			fv.validateExpressionCalls(s.Condition, scope)
		}
	case *ast.EnsureStmt:
		// Ensure statements are allowed
		if s.Condition != nil {
			fv.validateExpressionCalls(s.Condition, scope)
		}
	default:
		return
	}
}

// isStateVariable checks if an expression is a state variable assignment
func (fv *FunctionValidator) isStateVariable(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		// Check if it's a primed variable (next-state assignment)
		if e.Prime {
			return true
		}
		// Check if it's a state variable (not a parameter or local)
		symbol, found := fv.symbolTable.LookupSymbol(fv.symbolTable.GlobalScope, e.Name)
		if found && symbol.Kind == SymbolVariable {
			// It's a state variable if it's in the global scope
			return symbol.Scope == fv.symbolTable.GlobalScope
		}
		return false
	case *ast.SelectorExpr:
		// Record field access - check if the record is a state variable
		return fv.isStateVariable(e.X)
	case *ast.IndexExpr:
		// Collection element access - check if the collection is a state variable
		return fv.isStateVariable(e.X)
	default:
		return false
	}
}

// Helper methods to find scopes
func (fv *FunctionValidator) findFunctionScope(funcName string) *Scope {
	for _, scope := range fv.symbolTable.scopes {
		if scope.Kind == ScopeFunction && scope.Name == funcName {
			return scope
		}
	}
	return nil
}

func (fv *FunctionValidator) findActionScope(actionName string) *Scope {
	for _, scope := range fv.symbolTable.scopes {
		if scope.Kind == ScopeAction && scope.Name == actionName {
			return scope
		}
	}
	return nil
}

func (fv *FunctionValidator) findLetScope(bindingName string, parentScope *Scope) *Scope {
	for _, scope := range fv.symbolTable.scopes {
		if scope.Kind == ScopeBlock && scope.Parent == parentScope {
			if _, found := scope.Symbols[bindingName]; found {
				return scope
			}
		}
	}
	return nil
}

// addError adds an error message
func (fv *FunctionValidator) addError(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if pos.Line > 0 {
		fv.errors = append(fv.errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		fv.errors = append(fv.errors, msg)
	}
}

