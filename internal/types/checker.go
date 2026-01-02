package types

import (
	"fmt"

	"github.com/akkeshavan/spectre/pkg/ast"
)

// TypeError represents a type checking error
type TypeError struct {
	Position ast.Position
	Message  string
}

func (e *TypeError) Error() string {
	if e.Position.Line > 0 {
		return fmt.Sprintf("%d:%d: %s", e.Position.Line, e.Position.Column, e.Message)
	}
	return e.Message
}

// Checker performs type checking on AST nodes
type Checker struct {
	env    *Environment
	errors []*TypeError
}

// NewChecker creates a new type checker with the given environment
func NewChecker(env *Environment) *Checker {
	return &Checker{
		env:    env,
		errors: []*TypeError{},
	}
}

// Errors returns all type errors found during checking
func (c *Checker) Errors() []*TypeError {
	return c.errors
}

// addError adds a type error
func (c *Checker) addError(pos ast.Position, format string, args ...interface{}) {
	c.errors = append(c.errors, &TypeError{
		Position: pos,
		Message:  fmt.Sprintf(format, args...),
	})
}

// CheckExpression type-checks an expression and returns its type
func (c *Checker) CheckExpression(expr ast.Expr) Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.Ident:
		return c.checkIdent(e)
	case *ast.BasicLit:
		return c.checkBasicLit(e)
	case *ast.BinaryExpr:
		return c.checkBinaryExpr(e)
	case *ast.UnaryExpr:
		return c.checkUnaryExpr(e)
	case *ast.CallExpr:
		return c.checkCallExpr(e)
	case *ast.SelectorExpr:
		return c.checkSelectorExpr(e)
	case *ast.IndexExpr:
		return c.checkIndexExpr(e)
	case *ast.ParenExpr:
		return c.CheckExpression(e.X)
	case *ast.IfExpr:
		return c.checkIfExpr(e)
	case *ast.LetExpr:
		return c.checkLetExpr(e)
	default:
		c.addError(expr.Pos(), "unsupported expression type: %T", expr)
		return nil
	}
}

// CheckAssignment type-checks an assignment statement
// Returns true if the assignment is valid, false otherwise
func (c *Checker) CheckAssignment(stmt *ast.AssignStmt) bool {
	if stmt == nil {
		return false
	}

	// Check the right-hand side expression
	rightType := c.CheckExpression(stmt.Right)
	if rightType == nil {
		return false // Error already reported
	}

	// Check the left-hand side and get its type
	leftType := c.checkLValue(stmt.Left)
	if leftType == nil {
		return false // Error already reported
	}

	// Check if right type is assignable to left type
	if !IsAssignable(rightType, leftType) {
		c.addError(stmt.Pos(), "cannot assign %s to %s",
			rightType.String(), leftType.String())
		return false
	}

	return true
}

// checkLValue checks a left-hand side expression (lvalue) and returns its type
// An lvalue can be an identifier or a selector/index expression
func (c *Checker) checkLValue(expr ast.Expr) Type {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.Ident:
		// Handle primed variables (next-state variables)
		if e.Prime {
			// Primed variables refer to the next state of a variable
			// They should have the same type as the unprimed variable
			if typ, found := c.env.LookupVariable(e.Name); found {
				return typ
			}
			c.addError(e.Pos(), "undefined variable: %s (primed variable must reference existing variable)", e.Name)
			return nil
		}

		// Check if it's a variable (not a constant)
		if typ, found := c.env.LookupVariable(e.Name); found {
			return typ
		}
		// Check if it's a constant (constants cannot be assigned to)
		if _, found := c.env.LookupConstant(e.Name); found {
			c.addError(e.Pos(), "cannot assign to constant %s", e.Name)
			return nil
		}
		c.addError(e.Pos(), "undefined variable: %s", e.Name)
		return nil

	case *ast.SelectorExpr:
		// Field selection: record.field
		recordType := c.CheckExpression(e.X)
		if recordType == nil {
			return nil
		}

		if rec, ok := recordType.(*Record); ok {
			fieldType, exists := rec.Fields[e.Sel]
			if !exists {
				c.addError(e.Pos(), "field %s not found in record", e.Sel)
				return nil
			}
			return fieldType
		}

		c.addError(e.Pos(), "cannot assign to field %s of non-record type %s",
			e.Sel, recordType.String())
		return nil

	case *ast.IndexExpr:
		// Index access: array[index] or map[key]
		containerType := c.CheckExpression(e.X)
		if containerType == nil {
			return nil
		}

		// Check index type
		indexType := c.CheckExpression(e.Index)
		if indexType == nil {
			return nil
		}

		// Check if it's a list
		if list, ok := containerType.(*List); ok {
			if !isInt(indexType) {
				c.addError(e.Pos(), "list index must be int, got %s", indexType.String())
				return nil
			}
			return list.Element
		}

		// Check if it's a map
		if m, ok := containerType.(*Map); ok {
			if !IsAssignable(indexType, m.Key) {
				c.addError(e.Pos(), "map key type mismatch: cannot use %s as %s",
					indexType.String(), m.Key.String())
				return nil
			}
			return m.Value
		}

		c.addError(e.Pos(), "cannot index type %s", containerType.String())
		return nil

	default:
		c.addError(expr.Pos(), "invalid left-hand side of assignment: %T", expr)
		return nil
	}
}

// checkIdent checks an identifier expression
func (c *Checker) checkIdent(ident *ast.Ident) Type {
	name := ident.Name

	// Check if it's a variable
	if typ, found := c.env.LookupVariable(name); found {
		return typ
	}

	// Check if it's a constant
	if typ, found := c.env.LookupConstant(name); found {
		return typ
	}

	c.addError(ident.Pos(), "undefined identifier: %s", name)
	return nil
}

// checkBasicLit checks a basic literal expression
func (c *Checker) checkBasicLit(lit *ast.BasicLit) Type {
	switch lit.Kind {
	case ast.IntLit:
		return &Primitive{Kind: Int}
	case ast.FloatLit:
		return &Primitive{Kind: Float}
	case ast.StringLit:
		return &Primitive{Kind: Str}
	case ast.BoolLit:
		return &Primitive{Kind: Bool}
	default:
		c.addError(lit.Pos(), "unknown literal kind: %d", lit.Kind)
		return nil
	}
}

// checkBinaryExpr checks a binary expression
func (c *Checker) checkBinaryExpr(expr *ast.BinaryExpr) Type {
	leftType := c.CheckExpression(expr.Left)
	rightType := c.CheckExpression(expr.Right)

	if leftType == nil || rightType == nil {
		return nil // Error already reported
	}

	switch expr.Op {
	case ast.Add, ast.Sub, ast.Mul, ast.Div:
		// Arithmetic operations require numeric types
		if !isNumeric(leftType) || !isNumeric(rightType) {
			c.addError(expr.Pos(), "arithmetic operation requires numeric types, got %s and %s",
				leftType.String(), rightType.String())
			return nil
		}
		// Result type is the "wider" type (float if either is float, else int)
		if isFloat(leftType) || isFloat(rightType) {
			return &Primitive{Kind: Float}
		}
		return &Primitive{Kind: Int}

	case ast.Eq, ast.Neq:
		// Equality operations work on any types (if compatible)
		if !IsAssignable(leftType, rightType) && !IsAssignable(rightType, leftType) {
			c.addError(expr.Pos(), "cannot compare %s and %s",
				leftType.String(), rightType.String())
			return nil
		}
		return &Primitive{Kind: Bool}

	case ast.Lt, ast.Gt, ast.Leq, ast.Geq:
		// Comparison operations require numeric types
		if !isNumeric(leftType) || !isNumeric(rightType) {
			c.addError(expr.Pos(), "comparison operation requires numeric types, got %s and %s",
				leftType.String(), rightType.String())
			return nil
		}
		return &Primitive{Kind: Bool}

	case ast.And, ast.Or:
		// Logical operations require boolean types
		if !isBool(leftType) || !isBool(rightType) {
			c.addError(expr.Pos(), "logical operation requires boolean types, got %s and %s",
				leftType.String(), rightType.String())
			return nil
		}
		return &Primitive{Kind: Bool}

	default:
		c.addError(expr.Pos(), "unknown binary operator: %d", expr.Op)
		return nil
	}
}

// checkUnaryExpr checks a unary expression
func (c *Checker) checkUnaryExpr(expr *ast.UnaryExpr) Type {
	operandType := c.CheckExpression(expr.Expr)

	if operandType == nil {
		return nil // Error already reported
	}

	switch expr.Op {
	case ast.Not:
		// Logical negation requires boolean
		if !isBool(operandType) {
			c.addError(expr.Pos(), "logical negation requires boolean type, got %s",
				operandType.String())
			return nil
		}
		return &Primitive{Kind: Bool}

	case ast.Neg:
		// Arithmetic negation requires numeric
		if !isNumeric(operandType) {
			c.addError(expr.Pos(), "arithmetic negation requires numeric type, got %s",
				operandType.String())
			return nil
		}
		return operandType // Same type as operand

	default:
		c.addError(expr.Pos(), "unknown unary operator: %d", expr.Op)
		return nil
	}
}

// checkCallExpr checks a function call expression
func (c *Checker) checkCallExpr(expr *ast.CallExpr) Type {
	// Check arguments first (before checking function, to get better error messages)
	argTypes := make([]Type, len(expr.Args))
	for i, arg := range expr.Args {
		argTypes[i] = c.CheckExpression(arg)
		if argTypes[i] == nil {
			return nil // Error already reported
		}
	}

	// Try to find function signature
	// If Fun is a SelectorExpr, extract the method name
	var funcName string
	if sel, ok := expr.Fun.(*ast.SelectorExpr); ok {
		funcName = sel.Sel
		// Check the receiver type
		receiverType := c.CheckExpression(sel.X)
		if receiverType == nil {
			return nil
		}
		// For now, we'll handle method calls later
		// This is a placeholder
	} else if ident, ok := expr.Fun.(*ast.Ident); ok {
		funcName = ident.Name
		// Don't check the identifier as an expression - it's a function name
		// Just look it up directly
	} else {
		// For other function expressions, check them
		funType := c.CheckExpression(expr.Fun)
		if funType == nil {
			return nil
		}
		c.addError(expr.Pos(), "cannot call non-function type: %s", funType.String())
		return nil
	}

	if funcName != "" {
		sig, found := c.env.LookupFunction(funcName)
		if found {
			// Check argument count
			if len(argTypes) != len(sig.Parameters) {
				c.addError(expr.Pos(), "function %s expects %d arguments, got %d",
					funcName, len(sig.Parameters), len(argTypes))
				return nil
			}

			// Check argument types
			for i, argType := range argTypes {
				if !IsAssignable(argType, sig.Parameters[i]) {
					c.addError(expr.Pos(), "argument %d: cannot assign %s to %s",
						i+1, argType.String(), sig.Parameters[i].String())
					return nil
				}
			}

			return sig.Return
		}

		// Function not found
		c.addError(expr.Pos(), "undefined function: %s", funcName)
		return nil
	}

	// Should not reach here
	c.addError(expr.Pos(), "invalid function call")
	return nil
}

// checkSelectorExpr checks a field/method selection expression
func (c *Checker) checkSelectorExpr(expr *ast.SelectorExpr) Type {
	xType := c.CheckExpression(expr.X)

	if xType == nil {
		return nil // Error already reported
	}

	// Check if it's a record type
	if rec, ok := xType.(*Record); ok {
		fieldType, exists := rec.Fields[expr.Sel]
		if !exists {
			c.addError(expr.Pos(), "field %s not found in record type %s",
				expr.Sel, xType.String())
			return nil
		}
		return fieldType
	}

	// For other types (like method calls), we'll handle them later
	c.addError(expr.Pos(), "cannot select field %s from type %s",
		expr.Sel, xType.String())
	return nil
}

// checkIndexExpr checks an index expression (array/list/map access)
func (c *Checker) checkIndexExpr(expr *ast.IndexExpr) Type {
	xType := c.CheckExpression(expr.X)
	indexType := c.CheckExpression(expr.Index)

	if xType == nil || indexType == nil {
		return nil // Error already reported
	}

	// Check if it's a list
	if list, ok := xType.(*List); ok {
		if !isInt(indexType) {
			c.addError(expr.Pos(), "list index must be int, got %s", indexType.String())
			return nil
		}
		return list.Element
	}

	// Check if it's a map
	if m, ok := xType.(*Map); ok {
		if !IsAssignable(indexType, m.Key) {
			c.addError(expr.Pos(), "map key type mismatch: cannot use %s as %s",
				indexType.String(), m.Key.String())
			return nil
		}
		return m.Value
	}

	c.addError(expr.Pos(), "cannot index type %s", xType.String())
	return nil
}

// checkIfExpr checks an if-else expression
func (c *Checker) checkIfExpr(expr *ast.IfExpr) Type {
	condType := c.CheckExpression(expr.Condition)

	if condType == nil {
		return nil // Error already reported
	}

	if !isBool(condType) {
		c.addError(expr.Pos(), "if condition must be boolean, got %s", condType.String())
		return nil
	}

	// Create a new scope for the then branch
	thenEnv := NewChildEnvironment(c.env)
	thenChecker := &Checker{env: thenEnv, errors: c.errors}
	thenType := thenChecker.CheckExpression(expr.Then)

	// Create a new scope for the else branch
	elseEnv := NewChildEnvironment(c.env)
	elseChecker := &Checker{env: elseEnv, errors: c.errors}
	elseType := elseChecker.CheckExpression(expr.Else)

	// Merge errors
	c.errors = thenChecker.errors
	c.errors = append(c.errors, elseChecker.errors...)

	if thenType == nil || elseType == nil {
		return nil
	}

	// Both branches should return compatible types
	if !IsAssignable(thenType, elseType) && !IsAssignable(elseType, thenType) {
		c.addError(expr.Pos(), "if-else branches have incompatible types: %s and %s",
			thenType.String(), elseType.String())
		return nil
	}

	// Return the "wider" type
	if IsAssignable(thenType, elseType) {
		return elseType
	}
	return thenType
}

// checkLetExpr checks a let expression
func (c *Checker) checkLetExpr(expr *ast.LetExpr) Type {
	// Check the binding value
	valueType := c.CheckExpression(expr.Value)

	if valueType == nil {
		return nil // Error already reported
	}

	// Create a new scope for the let binding
	letEnv := NewChildEnvironment(c.env)
	letEnv.DeclareVariable(expr.Name, valueType)

	// Check the body in the new scope
	letChecker := &Checker{env: letEnv, errors: c.errors}
	bodyType := letChecker.CheckExpression(expr.Body)

	// Merge errors
	c.errors = letChecker.errors

	return bodyType
}

// Helper functions

func isNumeric(typ Type) bool {
	prim, ok := typ.(*Primitive)
	return ok && (prim.Kind == Int || prim.Kind == Float)
}

func isInt(typ Type) bool {
	prim, ok := typ.(*Primitive)
	return ok && prim.Kind == Int
}

func isFloat(typ Type) bool {
	prim, ok := typ.(*Primitive)
	return ok && prim.Kind == Float
}

func isBool(typ Type) bool {
	prim, ok := typ.(*Primitive)
	return ok && prim.Kind == Bool
}

