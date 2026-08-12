package semantic

import (
	"fmt"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// Validator performs semantic validation on an AST
type Validator struct {
	symbolTable      *SymbolTable
	resolver         *Resolver
	functionValidator *FunctionValidator
	errors           []string
	warnings         []string
}

// NewValidator creates a new validator
func NewValidator(symbolTable *SymbolTable, resolver *Resolver) *Validator {
	return &Validator{
		symbolTable:       symbolTable,
		resolver:          resolver,
		functionValidator: NewFunctionValidator(symbolTable),
		errors:            []string{},
		warnings:          []string{},
	}
}

// ValidateFile validates a file AST
func (v *Validator) ValidateFile(file *ast.File) []string {
	v.errors = []string{}
	v.warnings = []string{}

	// First, resolve all names (this catches undefined variables)
	resolutionErrors := v.resolver.ResolveFile(file)
	v.errors = append(v.errors, resolutionErrors...)

	// Then perform additional validations
	for _, decl := range file.Decls {
		v.validateDecl(decl)
	}

	// Validate function calls match declarations
	functionCallErrors := v.functionValidator.ValidateFunctionCalls(file)
	v.errors = append(v.errors, functionCallErrors...)

	return v.errors
}

// GetWarnings returns validation warnings (non-fatal issues)
func (v *Validator) GetWarnings() []string {
	return v.warnings
}

// validateDecl validates a declaration
func (v *Validator) validateDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.VariableDecl:
		v.validateVariableDecl(d)
	case *ast.ConstantDecl:
		v.validateConstantDecl(d)
	case *ast.FunctionDecl:
		v.validateFunctionDecl(d)
	case *ast.ActionDecl:
		v.validateActionDecl(d)
	case *ast.InitDecl:
		v.validateInitDecl(d)
	case *ast.OneOfInitDecl:
		v.validateOneOfInitDecl(d)
	case *ast.InvariantDecl:
		v.validateInvariantDecl(d)
	case *ast.TemporalDecl:
		v.validateTemporalDecl(d)
	case *ast.ModuleDecl:
		v.validateModuleDecl(d)
	default:
		return
	}
}

// validateVariableDecl validates a variable declaration
func (v *Validator) validateVariableDecl(decl *ast.VariableDecl) {
	// Variables must have a type
	if decl.Type == nil {
		v.addError(decl.Pos(), "variable %s must have a type", decl.Name)
	}

	// Check if variable name is valid (not a keyword, etc.)
	if !isValidIdentifier(decl.Name) {
		v.addError(decl.Pos(), "invalid variable name: %s", decl.Name)
	}
}

// validateConstantDecl validates a constant declaration
func (v *Validator) validateConstantDecl(decl *ast.ConstantDecl) {
	// Constants must have a type
	if decl.Type == nil {
		v.addError(decl.Pos(), "constant %s must have a type", decl.Name)
	}

	// Constants must have a value
	if decl.Value == nil {
		v.addError(decl.Pos(), "constant %s must have a value", decl.Name)
	}

	// Check if constant name is valid
	if !isValidIdentifier(decl.Name) {
		v.addError(decl.Pos(), "invalid constant name: %s", decl.Name)
	}
}

// validateFunctionDecl validates a function declaration
func (v *Validator) validateFunctionDecl(decl *ast.FunctionDecl) {
	// Function must have a return type
	if decl.ReturnType == nil {
		v.addError(decl.Pos(), "function %s must have a return type", decl.Name)
	}

	// Check parameter names for duplicates
	paramNames := make(map[string]bool)
	for _, param := range decl.Parameters {
		if paramNames[param.Name] {
			v.addError(param.Position, "duplicate parameter name: %s", param.Name)
		}
		paramNames[param.Name] = true

		// Parameters must have types
		if param.Type == nil {
			v.addError(param.Position, "parameter %s must have a type", param.Name)
		}
	}

	// Validate function body (check for state mutations)
	if decl.Body != nil {
		v.validateFunctionBody(decl.Body, decl.Name)
	}
}

// validateActionDecl validates an action declaration
func (v *Validator) validateActionDecl(decl *ast.ActionDecl) {
	// Check parameter names for duplicates
	paramNames := make(map[string]bool)
	for _, param := range decl.Parameters {
		if paramNames[param.Name] {
			v.addError(param.Position, "duplicate parameter name: %s", param.Name)
		}
		paramNames[param.Name] = true

		// Parameters must have types
		if param.Type == nil {
			v.addError(param.Position, "parameter %s must have a type", param.Name)
		}
	}

	// Validate action body
	if decl.Body != nil {
		v.validateActionBody(decl.Body)
	}
}

// validateInitDecl validates an init declaration
func (v *Validator) validateInitDecl(decl *ast.InitDecl) {
	// Init must have either a body or an expression
	if decl.Body == nil && decl.Expression == nil {
		v.addError(decl.Pos(), "init declaration must have a body or expression")
	}

	// Validate that all variables referenced in init are declared
	if decl.Body != nil {
		v.validateBlock(decl.Body, v.symbolTable.GlobalScope)
	}
	if decl.Expression != nil {
		v.validateExpression(decl.Expression, v.symbolTable.GlobalScope)
	}
}

// validateOneOfInitDecl validates a oneOf init declaration
func (v *Validator) validateOneOfInitDecl(decl *ast.OneOfInitDecl) {
	if len(decl.Options) == 0 {
		v.addError(decl.Pos(), "oneOf init declaration must have at least one option")
	}

	for i, option := range decl.Options {
		if option == nil {
			v.addError(decl.Pos(), "oneOf option %d is nil", i)
			continue
		}
		v.validateBlock(option, v.symbolTable.GlobalScope)
	}
}

// validateInvariantDecl validates an invariant declaration
func (v *Validator) validateInvariantDecl(decl *ast.InvariantDecl) {
	if decl.Condition == nil {
		v.addError(decl.Pos(), "invariant %s must have a condition", decl.Name)
		return
	}

	// Validate that all variables in condition are declared
	v.validateExpression(decl.Condition, v.symbolTable.GlobalScope)
}

// validateTemporalDecl validates a temporal declaration
func (v *Validator) validateTemporalDecl(decl *ast.TemporalDecl) {
	if decl.Expression == nil {
		v.addError(decl.Pos(), "temporal property %s must have an expression", decl.Name)
		return
	}

	// Validate that all variables in expression are declared
	v.validateExpression(decl.Expression, v.symbolTable.GlobalScope)
}

// validateModuleDecl validates a module declaration
func (v *Validator) validateModuleDecl(decl *ast.ModuleDecl) {
	if len(decl.Decls) == 0 {
		v.addWarning(decl.Pos(), "module %s has no declarations", decl.Name)
	}

	// Validate all declarations within the module
	for _, innerDecl := range decl.Decls {
		v.validateDecl(innerDecl)
	}
}

// validateFunctionBody validates a function body (ensures no state mutations)
func (v *Validator) validateFunctionBody(body *ast.BlockStmt, funcName string) {
	for _, stmt := range body.Statements {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			// Functions cannot mutate state variables
			if v.isStateVariable(s.Left) {
				v.addError(s.Pos(), "function %s cannot mutate state variable", funcName)
			}
		case *ast.ReturnStmt:
			// Return statements are allowed
			if s.Value != nil {
				v.validateExpression(s.Value, v.findFunctionScope(funcName))
			}
		case *ast.ExprStmt:
			v.validateExpression(s.Expr, v.findFunctionScope(funcName))
		case *ast.RequireStmt:
			// Require statements are allowed in functions
			if s.Condition != nil {
				v.validateExpression(s.Condition, v.findFunctionScope(funcName))
			}
		case *ast.EnsureStmt:
			// Ensure statements are allowed in functions
			if s.Condition != nil {
				v.validateExpression(s.Condition, v.findFunctionScope(funcName))
			}
		default:
			// Other statements are allowed
		}
	}
}

// validateActionBody validates an action body
func (v *Validator) validateActionBody(body *ast.BlockStmt) {
	for _, stmt := range body.Statements {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			// Actions can mutate state variables
			v.validateExpression(s.Left, v.symbolTable.GlobalScope)
			v.validateExpression(s.Right, v.symbolTable.GlobalScope)
		case *ast.ExprStmt:
			v.validateExpression(s.Expr, v.symbolTable.GlobalScope)
		case *ast.RequireStmt:
			if s.Condition != nil {
				v.validateExpression(s.Condition, v.symbolTable.GlobalScope)
			}
		case *ast.EnsureStmt:
			if s.Condition != nil {
				v.validateExpression(s.Condition, v.symbolTable.GlobalScope)
			}
		default:
			// Other statements
		}
	}
}

// validateBlock validates a block statement
func (v *Validator) validateBlock(block *ast.BlockStmt, scope *Scope) {
	if block == nil {
		return
	}

	for _, stmt := range block.Statements {
		v.validateStatement(stmt, scope)
	}
}

// validateStatement validates a statement
func (v *Validator) validateStatement(stmt ast.Stmt, scope *Scope) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		v.validateExpression(s.Expr, scope)
	case *ast.ReturnStmt:
		if s.Value != nil {
			v.validateExpression(s.Value, scope)
		}
	case *ast.AssignStmt:
		v.validateExpression(s.Left, scope)
		v.validateExpression(s.Right, scope)
	case *ast.RequireStmt:
		if s.Condition != nil {
			v.validateExpression(s.Condition, scope)
		}
	case *ast.EnsureStmt:
		if s.Condition != nil {
			v.validateExpression(s.Condition, scope)
		}
	default:
		return
	}
}

// validateExpression validates an expression (ensures all identifiers are declared)
func (v *Validator) validateExpression(expr ast.Expr, scope *Scope) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.BasicLit:
		// Literals don't need validation
		return
	case *ast.Ident:
		// Check if identifier is declared
		_, found := v.symbolTable.LookupSymbol(scope, e.Name)
		if !found {
			v.addError(e.Pos(), "undefined identifier: %s", e.Name)
		}
	case *ast.UnaryExpr:
		v.validateExpression(e.Expr, scope)
	case *ast.BinaryExpr:
		v.validateExpression(e.Left, scope)
		v.validateExpression(e.Right, scope)
	case *ast.ParenExpr:
		v.validateExpression(e.X, scope)
	case *ast.CallExpr:
		v.validateCallExpression(e, scope)
	case *ast.IndexExpr:
		v.validateExpression(e.X, scope)
		v.validateExpression(e.Index, scope)
	case *ast.SelectorExpr:
		v.validateExpression(e.X, scope)
	case *ast.IfExpr:
		v.validateExpression(e.Condition, scope)
		v.validateExpression(e.Then, scope)
		if e.Else != nil {
			v.validateExpression(e.Else, scope)
		}
	case *ast.LetExpr:
		v.validateLetExpression(e, scope)
	case *ast.AlwaysExpr:
		v.validateExpression(e.Expr, scope)
	case *ast.EventuallyExpr:
		v.validateExpression(e.Expr, scope)
	case *ast.UntilExpr:
		v.validateExpression(e.Left, scope)
		v.validateExpression(e.Right, scope)
	case *ast.LeadsToExpr:
		v.validateExpression(e.Left, scope)
		v.validateExpression(e.Right, scope)
	case *ast.WFExpr:
		v.validateExpression(e.Target, scope)
	case *ast.SFExpr:
		v.validateExpression(e.Target, scope)
	default:
		return
	}
}

// validateCallExpression validates a function call
func (v *Validator) validateCallExpression(expr *ast.CallExpr, scope *Scope) {
	// Validate function name
	switch fun := expr.Fun.(type) {
	case *ast.Ident:
		symbol, found := v.symbolTable.LookupSymbol(scope, fun.Name)
		if !found {
			v.addError(fun.Pos(), "undefined function: %s", fun.Name)
		} else if symbol.Kind != SymbolFunction {
			v.addError(fun.Pos(), "cannot call non-function: %s", fun.Name)
		}
	case *ast.SelectorExpr:
		// Method calls - validate the object
		v.validateExpression(fun.X, scope)
	default:
		v.addError(expr.Pos(), "invalid function call expression")
	}

	// Validate arguments
	for _, arg := range expr.Args {
		v.validateExpression(arg, scope)
	}
}

// validateLetExpression validates a let expression
func (v *Validator) validateLetExpression(expr *ast.LetExpr, scope *Scope) {
	// Validate the binding value
	v.validateExpression(expr.Value, scope)

	// Find the let scope (it should have been created during building)
	letScope := v.findLetScope(expr.Name, scope)
	if letScope == nil {
		// Create a temporary scope for validation
		letScope = v.symbolTable.NewScope(scope, ScopeBlock, "let")
	}

	// Validate the body expression (can reference the binding)
	v.validateExpression(expr.Body, letScope)
}

// isStateVariable checks if an expression is a state variable assignment
func (v *Validator) isStateVariable(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		// Check if it's a primed variable (next-state assignment)
		if e.Prime {
			return true
		}
		// Check if it's a state variable (not a parameter or local)
		symbol, found := v.symbolTable.LookupSymbol(v.symbolTable.GlobalScope, e.Name)
		if found && symbol.Kind == SymbolVariable {
			// It's a state variable if it's in the global scope
			return symbol.Scope == v.symbolTable.GlobalScope
		}
		return false
	case *ast.SelectorExpr:
		// Record field access - check if the record is a state variable
		return v.isStateVariable(e.X)
	case *ast.IndexExpr:
		// Collection element access - check if the collection is a state variable
		return v.isStateVariable(e.X)
	default:
		return false
	}
}

// findFunctionScope finds a function scope by name
func (v *Validator) findFunctionScope(funcName string) *Scope {
	for _, scope := range v.symbolTable.scopes {
		if scope.Kind == ScopeFunction && scope.Name == funcName {
			return scope
		}
	}
	return v.symbolTable.GlobalScope
}

// findLetScope finds a let scope (helper for validation)
func (v *Validator) findLetScope(bindingName string, parentScope *Scope) *Scope {
	// Search for a block scope that might contain this binding
	for _, scope := range v.symbolTable.scopes {
		if scope.Kind == ScopeBlock && scope.Parent == parentScope {
			// Check if this scope has the binding
			if _, found := scope.Symbols[bindingName]; found {
				return scope
			}
		}
	}
	return nil
}

// addError adds an error message
func (v *Validator) addError(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if pos.Line > 0 {
		v.errors = append(v.errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		v.errors = append(v.errors, msg)
	}
}

// addWarning adds a warning message
func (v *Validator) addWarning(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if pos.Line > 0 {
		v.warnings = append(v.warnings, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		v.warnings = append(v.warnings, msg)
	}
}

// isValidIdentifier checks if an identifier name is valid
func isValidIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Check if it starts with a letter or underscore
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	// Check remaining characters
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

