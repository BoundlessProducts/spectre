package semantic

import (
	"fmt"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// InheritanceAnalyzer analyzes module extension and super calls
type InheritanceAnalyzer struct {
	symbolTable    *SymbolTable
	moduleResolver *ModuleResolver
	errors         []string
}

// NewInheritanceAnalyzer creates a new inheritance analyzer
func NewInheritanceAnalyzer(symbolTable *SymbolTable, moduleResolver *ModuleResolver) *InheritanceAnalyzer {
	return &InheritanceAnalyzer{
		symbolTable:    symbolTable,
		moduleResolver: moduleResolver,
		errors:         []string{},
	}
}

// AnalyzeInheritance analyzes module extensions and super calls
func (ia *InheritanceAnalyzer) AnalyzeInheritance(file *ast.File) []string {
	ia.errors = []string{}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.ModuleDecl:
			ia.analyzeModule(d)
		}
	}

	return ia.errors
}

// analyzeModule analyzes a module for inheritance and super calls
func (ia *InheritanceAnalyzer) analyzeModule(decl *ast.ModuleDecl) {
	// Check if module extends another module
	if decl.Extends == "" {
		// No extension, just analyze declarations within
		for _, innerDecl := range decl.Decls {
			ia.analyzeDeclForSuperCalls(innerDecl, decl.Name, "")
		}
		return
	}

	// Module extends another module - verify parent exists
	parentModule, exists := ia.moduleResolver.GetModule(decl.Extends)
	if !exists {
		ia.addError(decl.Pos(), "module %s extends undefined module: %s", decl.Name, decl.Extends)
		return
	}

	// Analyze all declarations within this module
	for _, innerDecl := range decl.Decls {
		ia.analyzeDeclForSuperCalls(innerDecl, decl.Name, decl.Extends)
	}

	// Check for method overrides (actions/functions with same name as parent)
	ia.checkMethodOverrides(decl, parentModule)
}

// analyzeDeclForSuperCalls analyzes a declaration for super calls
func (ia *InheritanceAnalyzer) analyzeDeclForSuperCalls(decl ast.Decl, currentModule string, parentModule string) {
	switch d := decl.(type) {
	case *ast.ActionDecl:
		ia.analyzeActionForSuperCalls(d, currentModule, parentModule)
	case *ast.FunctionDecl:
		ia.analyzeFunctionForSuperCalls(d, currentModule, parentModule)
	case *ast.ModuleDecl:
		// Nested modules - analyze recursively
		ia.analyzeModule(d)
	default:
		return
	}
}

// analyzeActionForSuperCalls analyzes an action for super calls
func (ia *InheritanceAnalyzer) analyzeActionForSuperCalls(decl *ast.ActionDecl, currentModule string, parentModule string) {
	if decl.Body == nil {
		return
	}

	// Find the action scope
	actionScope := ia.findActionScope(decl.Name)
	if actionScope == nil {
		actionScope = ia.symbolTable.GlobalScope
	}

	// Analyze statements in the action body
	for _, stmt := range decl.Body.Statements {
		ia.analyzeStatementForSuperCalls(stmt, currentModule, parentModule, actionScope)
	}
}

// analyzeFunctionForSuperCalls analyzes a function for super calls
func (ia *InheritanceAnalyzer) analyzeFunctionForSuperCalls(decl *ast.FunctionDecl, currentModule string, parentModule string) {
	if decl.Body == nil {
		return
	}

	// Find the function scope
	funcScope := ia.findFunctionScope(decl.Name)
	if funcScope == nil {
		funcScope = ia.symbolTable.GlobalScope
	}

	// Analyze statements in the function body
	for _, stmt := range decl.Body.Statements {
		ia.analyzeStatementForSuperCalls(stmt, currentModule, parentModule, funcScope)
	}
}

// analyzeStatementForSuperCalls analyzes a statement for super calls
func (ia *InheritanceAnalyzer) analyzeStatementForSuperCalls(stmt ast.Stmt, currentModule string, parentModule string, scope *Scope) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		ia.analyzeExpressionForSuperCalls(s.Expr, currentModule, parentModule, scope)
	case *ast.ReturnStmt:
		if s.Value != nil {
			ia.analyzeExpressionForSuperCalls(s.Value, currentModule, parentModule, scope)
		}
	case *ast.AssignStmt:
		ia.analyzeExpressionForSuperCalls(s.Left, currentModule, parentModule, scope)
		ia.analyzeExpressionForSuperCalls(s.Right, currentModule, parentModule, scope)
	case *ast.RequireStmt:
		if s.Condition != nil {
			ia.analyzeExpressionForSuperCalls(s.Condition, currentModule, parentModule, scope)
		}
	case *ast.EnsureStmt:
		if s.Condition != nil {
			ia.analyzeExpressionForSuperCalls(s.Condition, currentModule, parentModule, scope)
		}
	default:
		return
	}
}

// analyzeExpressionForSuperCalls analyzes an expression for super calls
func (ia *InheritanceAnalyzer) analyzeExpressionForSuperCalls(expr ast.Expr, currentModule string, parentModule string, scope *Scope) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.SuperExpr:
		ia.analyzeSuperCall(e, currentModule, parentModule)
	case *ast.CallExpr:
		// Check if this is a super call (super.method())
		// The parser creates CallExpr with SuperExpr as Fun
		if super, ok := e.Fun.(*ast.SuperExpr); ok {
			ia.analyzeSuperCall(super, currentModule, parentModule)
			// Verify the method exists in parent module
			ia.verifySuperMethod(super.Method, parentModule, e.Pos())
		} else if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
			// Check if selector's X is a SuperExpr
			if super, ok := sel.X.(*ast.SuperExpr); ok {
				ia.analyzeSuperCall(super, currentModule, parentModule)
				// Verify the method exists in parent module
				ia.verifySuperMethod(sel.Sel, parentModule, e.Pos())
			}
		}
		// Analyze arguments
		for _, arg := range e.Args {
			ia.analyzeExpressionForSuperCalls(arg, currentModule, parentModule, scope)
		}
	case *ast.BinaryExpr:
		ia.analyzeExpressionForSuperCalls(e.Left, currentModule, parentModule, scope)
		ia.analyzeExpressionForSuperCalls(e.Right, currentModule, parentModule, scope)
	case *ast.UnaryExpr:
		ia.analyzeExpressionForSuperCalls(e.Expr, currentModule, parentModule, scope)
	case *ast.IfExpr:
		ia.analyzeExpressionForSuperCalls(e.Condition, currentModule, parentModule, scope)
		ia.analyzeExpressionForSuperCalls(e.Then, currentModule, parentModule, scope)
		if e.Else != nil {
			ia.analyzeExpressionForSuperCalls(e.Else, currentModule, parentModule, scope)
		}
	case *ast.LetExpr:
		ia.analyzeExpressionForSuperCalls(e.Value, currentModule, parentModule, scope)
		letScope := ia.findLetScope(e.Name, scope)
		if letScope == nil {
			letScope = scope
		}
		ia.analyzeExpressionForSuperCalls(e.Body, currentModule, parentModule, letScope)
	default:
		return
	}
}

// analyzeSuperCall analyzes a super call expression
func (ia *InheritanceAnalyzer) analyzeSuperCall(expr *ast.SuperExpr, currentModule string, parentModule string) {
	if parentModule == "" {
		ia.addError(expr.Pos(), "super call in module %s but module does not extend another module", currentModule)
		return
	}

	// Verify parent module exists
	_, exists := ia.moduleResolver.GetModule(parentModule)
	if !exists {
		ia.addError(expr.Pos(), "super call references undefined parent module: %s", parentModule)
		return
	}
}

// verifySuperMethod verifies that a super method exists in the parent module or any ancestor
func (ia *InheritanceAnalyzer) verifySuperMethod(methodName string, parentModule string, pos ast.Position) {
	if parentModule == "" {
		return
	}

	// Search through the inheritance chain
	visited := make(map[string]bool)
	currentModule := parentModule

	for currentModule != "" && !visited[currentModule] {
		visited[currentModule] = true

		moduleDecl, exists := ia.moduleResolver.GetModule(currentModule)
		if !exists {
			return // Already reported error
		}

		// Search for the method in current module
		found := false
		for _, decl := range moduleDecl.Decls {
			switch d := decl.(type) {
			case *ast.ActionDecl:
				if d.Name == methodName {
					found = true
					break
				}
			case *ast.FunctionDecl:
				if d.Name == methodName {
					found = true
					break
				}
			}
		}

		if found {
			return // Method found in this module or ancestor
		}

		// Continue up the inheritance chain
		currentModule = moduleDecl.Extends
	}

	// Method not found in any ancestor
	ia.addError(pos, "super method %s not found in parent module %s or its ancestors", methodName, parentModule)
}

// checkMethodOverrides checks if methods override parent methods correctly
func (ia *InheritanceAnalyzer) checkMethodOverrides(childModule *ast.ModuleDecl, parentModule *ast.ModuleDecl) {
	// Build a map of parent methods
	parentMethods := make(map[string]bool)
	for _, decl := range parentModule.Decls {
		switch d := decl.(type) {
		case *ast.ActionDecl:
			parentMethods[d.Name] = true
		case *ast.FunctionDecl:
			parentMethods[d.Name] = true
		}
	}

	// Check child methods
	for _, decl := range childModule.Decls {
		switch d := decl.(type) {
		case *ast.ActionDecl:
			if parentMethods[d.Name] {
				// Method override - verify signature compatibility
				ia.checkMethodSignature(d, parentModule, d.Name)
			}
		case *ast.FunctionDecl:
			if parentMethods[d.Name] {
				// Method override - verify signature compatibility
				ia.checkFunctionSignature(d, parentModule, d.Name)
			}
		}
	}
}

// checkMethodSignature checks if an action override has compatible signature
func (ia *InheritanceAnalyzer) checkMethodSignature(childAction *ast.ActionDecl, parentModule *ast.ModuleDecl, methodName string) {
	// Find parent action
	var parentAction *ast.ActionDecl
	for _, decl := range parentModule.Decls {
		if action, ok := decl.(*ast.ActionDecl); ok && action.Name == methodName {
			parentAction = action
			break
		}
	}

	if parentAction == nil {
		return
	}

	// Check parameter count
	if len(childAction.Parameters) != len(parentAction.Parameters) {
		ia.addError(childAction.Pos(), "action %s override has %d parameters, parent has %d",
			methodName, len(childAction.Parameters), len(parentAction.Parameters))
		return
	}

	// Check parameter types (basic check - full type checking done elsewhere)
	for i, childParam := range childAction.Parameters {
		if i < len(parentAction.Parameters) {
			parentParam := parentAction.Parameters[i]
			// Type checking will be done by type checker
			// Here we just verify the structure is correct
			_ = childParam
			_ = parentParam
		}
	}
}

// checkFunctionSignature checks if a function override has compatible signature
func (ia *InheritanceAnalyzer) checkFunctionSignature(childFunc *ast.FunctionDecl, parentModule *ast.ModuleDecl, methodName string) {
	// Find parent function
	var parentFunc *ast.FunctionDecl
	for _, decl := range parentModule.Decls {
		if fn, ok := decl.(*ast.FunctionDecl); ok && fn.Name == methodName {
			parentFunc = fn
			break
		}
	}

	if parentFunc == nil {
		return
	}

	// Check parameter count
	if len(childFunc.Parameters) != len(parentFunc.Parameters) {
		ia.addError(childFunc.Pos(), "function %s override has %d parameters, parent has %d",
			methodName, len(childFunc.Parameters), len(parentFunc.Parameters))
		return
	}

	// Check return type (basic check - full type checking done elsewhere)
	// Type checking will be done by type checker
}

// Helper methods to find scopes
func (ia *InheritanceAnalyzer) findActionScope(actionName string) *Scope {
	for _, scope := range ia.symbolTable.scopes {
		if scope.Kind == ScopeAction && scope.Name == actionName {
			return scope
		}
	}
	return nil
}

func (ia *InheritanceAnalyzer) findFunctionScope(funcName string) *Scope {
	for _, scope := range ia.symbolTable.scopes {
		if scope.Kind == ScopeFunction && scope.Name == funcName {
			return scope
		}
	}
	return nil
}

func (ia *InheritanceAnalyzer) findLetScope(bindingName string, parentScope *Scope) *Scope {
	for _, scope := range ia.symbolTable.scopes {
		if scope.Kind == ScopeBlock && scope.Parent == parentScope {
			if _, found := scope.Symbols[bindingName]; found {
				return scope
			}
		}
	}
	return nil
}

// addError adds an error message
func (ia *InheritanceAnalyzer) addError(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if pos.Line > 0 {
		ia.errors = append(ia.errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
	} else {
		ia.errors = append(ia.errors, msg)
	}
}

