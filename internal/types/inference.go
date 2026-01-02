package types

import (
	"github.com/spectre-lang/spectre/pkg/ast"
)

// InferType attempts to infer the type of an expression
// Returns the inferred type, or nil if inference is not possible
func InferType(expr ast.Expr, env *Environment) Type {
	if expr == nil {
		return nil
	}

	checker := NewChecker(env)
	return checker.CheckExpression(expr)
}

// InferLetBindingType infers the type of a let binding from its value expression
func InferLetBindingType(letExpr *ast.LetExpr, env *Environment) Type {
	if letExpr == nil {
		return nil
	}

	// Infer type from the value expression
	return InferType(letExpr.Value, env)
}

// InferFunctionReturnType attempts to infer the return type of a function from its body
// This analyzes the function body to determine what type is returned
func InferFunctionReturnType(funDecl *ast.FunctionDecl, env *Environment) Type {
	if funDecl == nil || funDecl.Body == nil {
		return nil
	}

	// Create a new environment for the function body
	// Include function parameters in the environment
	funcEnv := NewChildEnvironment(env)
	for _, param := range funDecl.Parameters {
		paramType, err := FromAST(param.Type)
		if err != nil {
			return nil // Can't infer if parameter types are invalid
		}
		funcEnv.DeclareVariable(param.Name, paramType)
	}

	checker := NewChecker(funcEnv)

	// Check all return statements in the function body
	var returnTypes []Type
	for _, stmt := range funDecl.Body.Statements {
		if retStmt, ok := stmt.(*ast.ReturnStmt); ok {
			retType := checker.CheckExpression(retStmt.Value)
			if retType != nil {
				returnTypes = append(returnTypes, retType)
			}
		}
	}

	// If we have return statements, try to find a common type
	if len(returnTypes) > 0 {
		// For now, return the first return type
		// In a more sophisticated implementation, we'd find the least common supertype
		return returnTypes[0]
	}

	// If no return statements, function returns void (or unit type)
	// For now, return nil to indicate no return type
	return nil
}

// InferExpressionType infers the type of an expression in a given context
// This is a convenience function that wraps InferType
func InferExpressionType(expr ast.Expr, env *Environment) Type {
	return InferType(expr, env)
}

