package types

import (
	"fmt"

	"github.com/BoundlessProducts/spectre/pkg/ast"
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
	modules map[string]*ast.ModuleDecl // Module name -> Module declaration (for module-qualified access)
}

// NewChecker creates a new type checker with the given environment
func NewChecker(env *Environment) *Checker {
	return &Checker{
		env:     env,
		errors:  []*TypeError{},
		modules: make(map[string]*ast.ModuleDecl),
	}
}

// SetModules sets the modules map for module-qualified name resolution
func (c *Checker) SetModules(modules map[string]*ast.ModuleDecl) {
	c.modules = modules
}

// Errors returns all type errors found during checking
func (c *Checker) Errors() []*TypeError {
	return c.errors
}

// MergeErrors merges errors from another checker into this checker
func (c *Checker) MergeErrors(other *Checker) {
	c.errors = append(c.errors, other.errors...)
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
	case *ast.LambdaExpr:
		return c.checkLambdaExpr(e)
	case *ast.RecordLiteral:
		return c.checkRecordLiteral(e)
	case *ast.SetLiteral:
		return c.checkSetLiteral(e)
	case *ast.ListLiteral:
		return c.checkListLiteral(e)
	case *ast.AlwaysExpr:
		// Type-check the inner expression - temporal expressions evaluate to bool
		c.CheckExpression(e.Expr)
		return &Primitive{Kind: Bool}
	case *ast.EventuallyExpr:
		// Type-check the inner expression - temporal expressions evaluate to bool
		c.CheckExpression(e.Expr)
		return &Primitive{Kind: Bool}
	case *ast.UntilExpr:
		// Type-check both expressions - temporal expressions evaluate to bool
		c.CheckExpression(e.Left)
		c.CheckExpression(e.Right)
		return &Primitive{Kind: Bool}
	case *ast.WFExpr, *ast.SFExpr, *ast.LeadsToExpr:
		// These are handled during verification, return bool as placeholder
		return &Primitive{Kind: Bool}
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
		// Resolve any named types in the type (e.g., Set<User> -> Set<Record{...}>)
		// Multiple passes to ensure deep resolution
		resolved := typ
		for i := 0; i < 5; i++ {
			prevResolved := resolved
			resolved = c.resolveNamedTypesInType(resolved)
			if resolved == prevResolved {
				break
			}
		}
		return resolved
	}

	// Check if it's a constant
	if typ, found := c.env.LookupConstant(name); found {
		return typ
	}

	// Check if it's a type name (e.g., enum ProcessState)
	// This allows enum value access like ProcessState.Idle
	if typ, found := c.env.LookupType(name); found {
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
	// Try to find function signature
	// If Fun is a SelectorExpr, extract the method name
	var funcName string
	var receiverType Type
	needsLambdaInference := false
	var lambda *ast.LambdaExpr
	
	if sel, ok := expr.Fun.(*ast.SelectorExpr); ok {
		funcName = sel.Sel
		
		// Check if it's a static method call (Set.empty(), Set.of(), etc.)
		if ident, ok := sel.X.(*ast.Ident); ok {
			if ident.Name == "Set" || ident.Name == "List" || ident.Name == "Map" || ident.Name == "Option" {
				// Static method call on collection type - check arguments normally
				argTypes := make([]Type, len(expr.Args))
				for i, arg := range expr.Args {
					argTypes[i] = c.CheckExpression(arg)
					if argTypes[i] == nil {
						return nil // Error already reported
					}
				}
				return c.checkStaticMethodCall(ident.Name, funcName, argTypes, expr.Pos())
			}
			
			// Check if it's a module-qualified function call (e.g., ControllerModule.distance(...))
			if moduleDecl, found := c.modules[ident.Name]; found {
				// Look for the function in the module's declarations
				for _, decl := range moduleDecl.Decls {
					if funcDecl, ok := decl.(*ast.FunctionDecl); ok {
						if funcDecl.Name == funcName && funcDecl.Visibility == ast.Public {
							// Build function signature
							params := make([]Type, len(funcDecl.Parameters))
							for i, param := range funcDecl.Parameters {
								paramType, err := c.resolveType(param.Type)
								if err == nil {
									params[i] = paramType
								}
							}
							returnType, err := c.resolveType(funcDecl.ReturnType)
							if err == nil {
								// Check arguments
								argTypes := make([]Type, len(expr.Args))
								for i, arg := range expr.Args {
									argTypes[i] = c.CheckExpression(arg)
									if argTypes[i] == nil {
										return nil
									}
								}
								
								// Check argument count and types
								if len(argTypes) != len(params) {
									c.addError(expr.Pos(), "function %s.%s expects %d arguments, got %d",
										ident.Name, funcName, len(params), len(argTypes))
									return nil
								}
								
								for i, argType := range argTypes {
									if !IsAssignable(argType, params[i]) {
										c.addError(expr.Pos(), "argument %d: cannot assign %s to %s",
											i+1, argType.String(), params[i].String())
										return nil
									}
								}
								
								return returnType
							}
						}
					}
				}
				c.addError(expr.Pos(), "function %s not found in module %s or not public",
					funcName, ident.Name)
				return nil
			}
		}
		
		// Instance method call - check the receiver type first
		receiverType = c.CheckExpression(sel.X)
		if receiverType == nil {
			return nil
		}
		
		// Resolve any named types in the receiver type (e.g., Set<User> -> Set<Record{...}>)
		// Multiple passes to ensure deep resolution
		for i := 0; i < 3; i++ {
			resolved := c.resolveNamedTypesInType(receiverType)
			if resolved == receiverType {
				break
			}
			receiverType = resolved
		}
		
		// Check if first argument is a lambda that needs type inference
		if len(expr.Args) > 0 {
			if lambdaExpr, ok := expr.Args[0].(*ast.LambdaExpr); ok {
				needsLambdaInference = true
				lambda = lambdaExpr
			}
		}
	}
	
	// Check arguments - but if we need lambda inference, do it first
	argTypes := make([]Type, len(expr.Args))
	if needsLambdaInference && lambda != nil {
		// Infer lambda parameter types from method context before checking arguments
		inferredLambdaType := c.inferLambdaTypeFromMethod(receiverType, funcName, lambda, nil)
		if inferredLambdaType != nil {
			// Check the lambda with inferred types
			lambdaChecker := &Checker{env: c.env, errors: []*TypeError{}}
			recheckedLambdaType := lambdaChecker.checkLambdaExprWithExpected(lambda, inferredLambdaType)
			// Merge errors from lambda re-checking
			c.errors = append(c.errors, lambdaChecker.errors...)
			// Use the properly type-checked lambda
			argTypes[0] = recheckedLambdaType
		} else {
			// Fallback: check lambda without inference
			argTypes[0] = c.CheckExpression(expr.Args[0])
			if argTypes[0] == nil {
				return nil
			}
		}
		// Check remaining arguments normally
		for i := 1; i < len(expr.Args); i++ {
			argTypes[i] = c.CheckExpression(expr.Args[i])
			if argTypes[i] == nil {
				return nil // Error already reported
			}
		}
	} else {
		// No lambda inference needed - check all arguments normally
		for i, arg := range expr.Args {
			argTypes[i] = c.CheckExpression(arg)
			if argTypes[i] == nil {
				return nil // Error already reported
			}
		}
	}
	
	if _, ok := expr.Fun.(*ast.SelectorExpr); ok {
		// Handle known collection methods
		return c.checkMethodCall(receiverType, funcName, argTypes, expr.Pos())
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
				if argType == nil {
					c.addError(expr.Pos(), "argument %d: cannot determine type", i+1)
					return nil
				}
				paramType := sig.Parameters[i]
				if paramType == nil {
					c.addError(expr.Pos(), "argument %d: function parameter type is nil", i+1)
					return nil
				}
				if !IsAssignable(argType, paramType) {
					c.addError(expr.Pos(), "argument %d: cannot assign %s to %s",
						i+1, argType.String(), paramType.String())
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
	// First check if it's a module-qualified name (e.g., ControllerModule.sameDirection)
	if ident, ok := expr.X.(*ast.Ident); ok {
		if moduleDecl, found := c.modules[ident.Name]; found {
			// Look for the member (constant or function) in the module's declarations
			for _, decl := range moduleDecl.Decls {
				switch d := decl.(type) {
				case *ast.ConstantDecl:
					if d.Name == expr.Sel && d.Visibility == ast.Public {
						// Resolve the constant's type
						resolvedType, err := c.resolveType(d.Type)
						if err == nil {
							return resolvedType
						}
					}
				case *ast.FunctionDecl:
					if d.Name == expr.Sel && d.Visibility == ast.Public {
						// Build function signature
						params := make([]Type, len(d.Parameters))
						for i, param := range d.Parameters {
							paramType, err := c.resolveType(param.Type)
							if err == nil {
								params[i] = paramType
							}
						}
						returnType, err := c.resolveType(d.ReturnType)
						if err == nil {
							return &Function{
								Params:     params,
								ReturnType: returnType,
							}
						}
					}
				}
			}
			c.addError(expr.Pos(), "member %s not found in module %s or not public",
				expr.Sel, ident.Name)
			return nil
		}
	}

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

	// Check if it's an enum type - enum values like ProcessState.Idle
	if enum, ok := xType.(*Enum); ok {
		// Check if the selector is a valid enum value
		found := false
		for _, val := range enum.Values {
			if val == expr.Sel {
				found = true
				break
			}
		}
		if !found {
			c.addError(expr.Pos(), "enum value %s not found in enum %s",
				expr.Sel, enum.Name)
			return nil
		}
		// Enum value access returns the enum type itself
		return enum
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

// checkLambdaExpr checks a lambda expression
// If expectedType is provided, it's used to infer parameter types
func (c *Checker) checkLambdaExpr(expr *ast.LambdaExpr) Type {
	return c.checkLambdaExprWithExpected(expr, nil)
}

// checkLambdaExprWithExpected checks a lambda expression with an expected function type
func (c *Checker) checkLambdaExprWithExpected(expr *ast.LambdaExpr, expectedType *Function) Type {
	// Create a new scope for lambda parameters
	lambdaEnv := NewChildEnvironment(c.env)
	
	// Add parameters to the lambda environment
	var paramTypes []Type
	for i, param := range expr.Params {
		var paramType Type
		if param.Type != nil {
			// Type is explicitly specified
			var err error
			paramType, err = c.resolveType(param.Type)
			if err != nil {
				c.addError(expr.Pos(), "invalid parameter type in lambda: %v", err)
				return nil
			}
		} else if expectedType != nil && i < len(expectedType.Params) {
			// Infer from expected type
			paramType = expectedType.Params[i]
			// Resolve any named types in the parameter type (multiple passes)
			for j := 0; j < 5; j++ {
				resolved := c.resolveNamedTypesInType(paramType)
				if resolved == paramType {
					break
				}
				paramType = resolved
			}
		} else {
			// Cannot infer - this will cause an error later
			// For now, use a placeholder
			paramType = &Primitive{Kind: Int} // Placeholder
		}
		
		paramTypes = append(paramTypes, paramType)
		lambdaEnv.DeclareVariable(param.Name, paramType)
	}
	
	// Check lambda body in the new environment
	lambdaChecker := &Checker{env: lambdaEnv, errors: c.errors}
	bodyType := lambdaChecker.CheckExpression(expr.Body)
	
	// Merge errors
	c.errors = lambdaChecker.errors
	
	if bodyType == nil {
		return nil
	}
	
	// Return a function type representing the lambda
	return &Function{
		Params:     paramTypes,
		ReturnType: bodyType,
	}
}

// inferLambdaTypeFromMethod infers lambda parameter types from method call context
func (c *Checker) inferLambdaTypeFromMethod(receiverType Type, methodName string, lambda *ast.LambdaExpr, currentArgType Type) *Function {
	// Infer based on method name and receiver type
	switch methodName {
	case "filter", "map", "forall", "exists":
		// These methods take a lambda: element => bool (filter/forall/exists) or element => newElement (map)
		if setType, ok := receiverType.(*Set); ok {
			// Set<T>.filter/map/etc. expects lambda: T => ...
			if len(lambda.Params) == 1 && lambda.Params[0].Type == nil {
				// Resolve the element type (in case it's a named type)
				// Multiple passes to ensure deep resolution
				elementType := setType.Element
				for i := 0; i < 3; i++ {
					resolved := c.resolveNamedTypesInType(elementType)
					if resolved == elementType {
						break
					}
					elementType = resolved
				}
				// Infer parameter type from set element type
				expectedFunc := &Function{
					Params:     []Type{elementType},
					ReturnType: &Primitive{Kind: Bool}, // Default return type
				}
				if methodName == "map" {
					// Map returns same type as lambda return type
					expectedFunc.ReturnType = nil // Will be inferred from lambda body
				}
				return expectedFunc
			}
		}
		if listType, ok := receiverType.(*List); ok {
			// List<T>.filter/map/etc. expects lambda: T => ...
			if len(lambda.Params) == 1 && lambda.Params[0].Type == nil {
				// Resolve the element type (in case it's a named type)
				// Multiple passes to ensure deep resolution
				elementType := listType.Element
				for i := 0; i < 3; i++ {
					resolved := c.resolveNamedTypesInType(elementType)
					if resolved == elementType {
						break
					}
					elementType = resolved
				}
				expectedFunc := &Function{
					Params:     []Type{elementType},
					ReturnType: &Primitive{Kind: Bool},
				}
				if methodName == "map" {
					expectedFunc.ReturnType = nil
				}
				return expectedFunc
			}
		}
	case "reduce":
		// reduce(initial, (acc, element) => acc)
		if setType, ok := receiverType.(*Set); ok {
			if len(lambda.Params) == 2 && lambda.Params[0].Type == nil && lambda.Params[1].Type == nil {
				// Infer from first argument type (initial value)
				if len(lambda.Params) >= 1 {
					// Resolve the element type (in case it's a named type)
					elementType := c.resolveNamedTypesInType(setType.Element)
					// First param is accumulator (same type as initial)
					// Second param is element (set element type)
					return &Function{
						Params:     []Type{currentArgType, elementType}, // Will be refined
						ReturnType: currentArgType,
					}
				}
			}
		}
	}
	return nil
}

// checkMethodCall checks a method call on a receiver type
func (c *Checker) checkMethodCall(receiverType Type, methodName string, argTypes []Type, pos ast.Position) Type {
	switch methodName {
	case "filter":
		if len(argTypes) != 1 {
			c.addError(pos, "filter expects 1 argument (predicate), got %d", len(argTypes))
			return nil
		}
		// Check that argument is a function: T => bool
		if funcType, ok := argTypes[0].(*Function); ok {
			if len(funcType.Params) != 1 {
				c.addError(pos, "filter predicate must have 1 parameter")
				return nil
			}
			if !isBool(funcType.ReturnType) {
				c.addError(pos, "filter predicate must return bool, got %s", funcType.ReturnType.String())
				return nil
			}
			// Return same collection type
			return receiverType
		}
		c.addError(pos, "filter argument must be a function")
		return nil
	case "map":
		if len(argTypes) != 1 {
			c.addError(pos, "map expects 1 argument (function), got %d", len(argTypes))
			return nil
		}
		if funcType, ok := argTypes[0].(*Function); ok {
			if len(funcType.Params) != 1 {
				c.addError(pos, "map function must have 1 parameter")
				return nil
			}
			// Return collection with mapped element type
			if _, ok := receiverType.(*Set); ok {
				return &Set{Element: funcType.ReturnType}
			}
			if _, ok := receiverType.(*List); ok {
				return &List{Element: funcType.ReturnType}
			}
		}
		c.addError(pos, "map argument must be a function")
		return nil
	case "reduce":
		if len(argTypes) != 2 {
			c.addError(pos, "reduce expects 2 arguments (initial, function), got %d", len(argTypes))
			return nil
		}
		if funcType, ok := argTypes[1].(*Function); ok {
			if len(funcType.Params) != 2 {
				c.addError(pos, "reduce function must have 2 parameters")
				return nil
			}
			// Return type is the accumulator type (first argument type)
			return argTypes[0]
		}
		c.addError(pos, "reduce second argument must be a function")
		return nil
	case "forall", "exists":
		if len(argTypes) != 1 {
			c.addError(pos, "%s expects 1 argument (predicate), got %d", methodName, len(argTypes))
			return nil
		}
		if funcType, ok := argTypes[0].(*Function); ok {
			if len(funcType.Params) != 1 {
				c.addError(pos, "%s predicate must have 1 parameter", methodName)
				return nil
			}
			if !isBool(funcType.ReturnType) {
				c.addError(pos, "%s predicate must return bool, got %s", methodName, funcType.ReturnType.String())
				return nil
			}
			return &Primitive{Kind: Bool}
		}
		c.addError(pos, "%s argument must be a function", methodName)
		return nil
	case "size":
		if len(argTypes) != 0 {
			c.addError(pos, "size expects 0 arguments, got %d", len(argTypes))
			return nil
		}
		return &Primitive{Kind: Int}
	case "union", "intersection":
		if len(argTypes) != 1 {
			c.addError(pos, "%s expects 1 argument, got %d", methodName, len(argTypes))
			return nil
		}
		// Argument must be the same collection type
		if !IsAssignable(argTypes[0], receiverType) && !IsAssignable(receiverType, argTypes[0]) {
			c.addError(pos, "%s argument must be compatible with receiver type %s, got %s",
				methodName, receiverType.String(), argTypes[0].String())
			return nil
		}
		return receiverType
	case "contains":
		if len(argTypes) != 1 {
			c.addError(pos, "contains expects 1 argument, got %d", len(argTypes))
			return nil
		}
		// Argument should be compatible with element type
		if setType, ok := receiverType.(*Set); ok {
			if !IsAssignable(argTypes[0], setType.Element) && !IsAssignable(setType.Element, argTypes[0]) {
				c.addError(pos, "contains argument must be compatible with element type %s, got %s",
					setType.Element.String(), argTypes[0].String())
				return nil
			}
		} else if listType, ok := receiverType.(*List); ok {
			if !IsAssignable(argTypes[0], listType.Element) && !IsAssignable(listType.Element, argTypes[0]) {
				c.addError(pos, "contains argument must be compatible with element type %s, got %s",
					listType.Element.String(), argTypes[0].String())
				return nil
			}
		}
		return &Primitive{Kind: Bool}
	case "put":
		// Map.put(key, value) - returns a new map with the key-value pair added/updated
		if mapType, ok := receiverType.(*Map); ok {
			if len(argTypes) != 2 {
				c.addError(pos, "put expects 2 arguments (key, value), got %d", len(argTypes))
				return nil
			}
			// Check key type
			if !IsAssignable(argTypes[0], mapType.Key) && !IsAssignable(mapType.Key, argTypes[0]) {
				c.addError(pos, "put key argument must be compatible with key type %s, got %s",
					mapType.Key.String(), argTypes[0].String())
				return nil
			}
			// Check value type
			if !IsAssignable(argTypes[1], mapType.Value) && !IsAssignable(mapType.Value, argTypes[1]) {
				c.addError(pos, "put value argument must be compatible with value type %s, got %s",
					mapType.Value.String(), argTypes[1].String())
				return nil
			}
			// Return the same map type
			return receiverType
		}
		c.addError(pos, "put can only be called on Map, got %s", receiverType.String())
		return nil
	case "get":
		// Map.get(key) - returns the value for the key (or Option<Value> if key might not exist)
		// For now, we'll return the value type directly (assuming key exists)
		// In the future, this could return Option<Value>
		if mapType, ok := receiverType.(*Map); ok {
			if len(argTypes) != 1 {
				c.addError(pos, "get expects 1 argument (key), got %d", len(argTypes))
				return nil
			}
			// Check key type
			if !IsAssignable(argTypes[0], mapType.Key) && !IsAssignable(mapType.Key, argTypes[0]) {
				c.addError(pos, "get key argument must be compatible with key type %s, got %s",
					mapType.Key.String(), argTypes[0].String())
				return nil
			}
			// Return the value type
			return mapType.Value
		}
		c.addError(pos, "get can only be called on Map, got %s", receiverType.String())
		return nil
	case "append":
		if len(argTypes) != 1 {
			c.addError(pos, "append expects 1 argument, got %d", len(argTypes))
			return nil
		}
		// Argument must be compatible with element type
		if listType, ok := receiverType.(*List); ok {
			if !IsAssignable(argTypes[0], listType.Element) && !IsAssignable(listType.Element, argTypes[0]) {
				c.addError(pos, "append argument must be compatible with element type %s, got %s",
					listType.Element.String(), argTypes[0].String())
				return nil
			}
			return receiverType
		}
		c.addError(pos, "append can only be called on List, got %s", receiverType.String())
		return nil
	case "head", "tail", "toList", "toSet":
		// These methods are handled in the evaluator, but we need to type-check them
		if methodName == "head" {
			if len(argTypes) != 0 {
				c.addError(pos, "head expects 0 arguments, got %d", len(argTypes))
				return nil
			}
			if listType, ok := receiverType.(*List); ok {
				return listType.Element
			}
		} else if methodName == "tail" {
			if len(argTypes) != 0 {
				c.addError(pos, "tail expects 0 arguments, got %d", len(argTypes))
				return nil
			}
			return receiverType
		} else if methodName == "toList" {
			if len(argTypes) != 0 {
				c.addError(pos, "toList expects 0 arguments, got %d", len(argTypes))
				return nil
			}
			if setType, ok := receiverType.(*Set); ok {
				return &List{Element: setType.Element}
			}
		} else if methodName == "toSet" {
			if len(argTypes) != 0 {
				c.addError(pos, "toSet expects 0 arguments, got %d", len(argTypes))
				return nil
			}
			if listType, ok := receiverType.(*List); ok {
				return &Set{Element: listType.Element}
			}
		}
		return receiverType
	default:
		c.addError(pos, "unknown method: %s", methodName)
		return nil
	}
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

// resolveType resolves a type from AST, handling named types by looking them up in the environment
func (c *Checker) resolveType(astType ast.Type) (Type, error) {
	return FromASTWithResolver(astType, func(name string) (Type, bool) {
		return c.env.LookupType(name)
	})
}

// ResolveNamedTypesInType recursively resolves any named types within a type
// For example, if we have Set<User> where User is a type alias, this will resolve it to Set<Record{...}>
func (c *Checker) ResolveNamedTypesInType(typ Type) Type {
	return c.resolveNamedTypesInType(typ)
}

// resolveNamedTypesInType recursively resolves any named types within a type
// For example, if we have Set<User> where User is a type alias, this will resolve it to Set<Record{...}>
func (c *Checker) resolveNamedTypesInType(typ Type) Type {
	if typ == nil {
		return nil
	}
	
	switch t := typ.(type) {
	case *Set:
		// Resolve the element type
		resolvedElement := c.resolveNamedTypesInType(t.Element)
		if resolvedElement != t.Element {
			return &Set{Element: resolvedElement}
		}
		return typ
	case *List:
		// Resolve the element type
		resolvedElement := c.resolveNamedTypesInType(t.Element)
		if resolvedElement != t.Element {
			return &List{Element: resolvedElement}
		}
		return typ
	case *Map:
		// Resolve key and value types
		resolvedKey := c.resolveNamedTypesInType(t.Key)
		resolvedValue := c.resolveNamedTypesInType(t.Value)
		if resolvedKey != t.Key || resolvedValue != t.Value {
			return &Map{Key: resolvedKey, Value: resolvedValue}
		}
		return typ
	case *Option:
		// Resolve the element type
		resolvedElement := c.resolveNamedTypesInType(t.Element)
		if resolvedElement != t.Element {
			return &Option{Element: resolvedElement}
		}
		return typ
	case *Named:
		// If it's a named type, try to resolve it
		if resolved, found := c.env.LookupType(t.Name); found {
			// Resolve recursively to handle nested named types
			// Use multiple passes to ensure deep resolution
			resolvedType := resolved
			for i := 0; i < 5; i++ {
				prevResolved := resolvedType
				resolvedType = c.resolveNamedTypesInType(resolvedType)
				if resolvedType == prevResolved {
					break
				}
			}
			return resolvedType
		}
		// If it has a Base type, try resolving that
		if t.Base != nil {
			resolvedBase := c.resolveNamedTypesInType(t.Base)
			if resolvedBase != t.Base {
				return resolvedBase
			}
		}
		// Don't try to resolve as primitive - if it's not in the environment,
		// it should be reported as an error, not silently converted to int
		// Return the original Named type if we can't resolve it
		// This allows the error to be reported later during type checking
		return typ
	default:
		// Primitive types, records, etc. don't need resolution
		return typ
	}
}

// checkStaticMethodCall checks static method calls like Set.empty(), Set.of(x), etc.
func (c *Checker) checkStaticMethodCall(collectionType string, methodName string, argTypes []Type, pos ast.Position) Type {
	switch methodName {
	case "empty":
		if len(argTypes) != 0 {
			c.addError(pos, "%s.empty() expects 0 arguments, got %d", collectionType, len(argTypes))
			return nil
		}
		switch collectionType {
		case "Set":
			return &Set{Element: &Primitive{Kind: Int}} // Element type will be inferred from usage
		case "List":
			return &List{Element: &Primitive{Kind: Int}} // Element type will be inferred from usage
		case "Map":
			return &Map{Key: &Primitive{Kind: Int}, Value: &Primitive{Kind: Int}} // Types will be inferred
		case "Option":
			return &Option{Element: &Primitive{Kind: Int}} // Element type will be inferred
		default:
			c.addError(pos, "unknown collection type: %s", collectionType)
			return nil
		}
	case "of":
		if len(argTypes) != 1 {
			c.addError(pos, "%s.of() expects 1 argument, got %d", collectionType, len(argTypes))
			return nil
		}
		elemType := argTypes[0]
		if elemType == nil {
			c.addError(pos, "cannot determine element type for %s.of()", collectionType)
			return nil
		}
		switch collectionType {
		case "Set":
			return &Set{Element: elemType}
		case "List":
			return &List{Element: elemType}
		default:
			c.addError(pos, "%s.of() not supported", collectionType)
			return nil
		}
	default:
		c.addError(pos, "unknown static method %s.%s", collectionType, methodName)
		return nil
	}
}

// checkRecordLiteral type-checks a record literal expression
// It infers the record type from the field types
func (c *Checker) checkRecordLiteral(expr *ast.RecordLiteral) Type {
	fields := make(map[string]Type)
	
	// Check each field value and infer its type
	for _, field := range expr.Fields {
		if field.Spread {
			// Handle spread operator: ...identifier
			spreadType := c.CheckExpression(field.Value)
			if spreadType == nil {
				return nil
			}
			// Spread should be a record type
			if rec, ok := spreadType.(*Record); ok {
				// Merge fields from spread record
				for name, typ := range rec.Fields {
					fields[name] = typ
				}
			} else {
				c.addError(expr.Pos(), "spread operator can only be used with record types, got %s", spreadType.String())
				return nil
			}
		} else {
			// Regular field: name: value
			fieldType := c.CheckExpression(field.Value)
			if fieldType == nil {
				return nil
			}
			fields[field.Name] = fieldType
		}
	}
	
	return &Record{Fields: fields}
}

// checkSetLiteral type-checks a set literal expression
// It infers the element type from the element types
func (c *Checker) checkSetLiteral(expr *ast.SetLiteral) Type {
	if len(expr.Elements) == 0 {
		// Empty set literal, cannot infer type - will need type annotation
		// For now, return a generic Set<int> as a placeholder
		// In practice, this will be inferred from context
		return &Set{Element: &Primitive{Kind: Int}}
	}
	
	// Check all elements and find common type
	var elementType Type
	for i, elem := range expr.Elements {
		elemType := c.CheckExpression(elem)
		if elemType == nil {
			return nil
		}
		
		if i == 0 {
			elementType = elemType
		} else {
			// Try to find a common type
			// For now, if types are assignable, use the first type
			// In the future, we might want more sophisticated type unification
			if !IsAssignable(elemType, elementType) && !IsAssignable(elementType, elemType) {
				c.addError(expr.Pos(), "set elements have incompatible types: %s and %s",
					elementType.String(), elemType.String())
				return nil
			}
		}
	}
	
	return &Set{Element: elementType}
}

// checkListLiteral type-checks a list literal expression
// It infers the element type from the element types
func (c *Checker) checkListLiteral(expr *ast.ListLiteral) Type {
	if len(expr.Elements) == 0 {
		// Empty list literal, cannot infer type - will need type annotation
		// For now, return a generic List<int> as a placeholder
		// In practice, this will be inferred from context
		return &List{Element: &Primitive{Kind: Int}}
	}
	
	// Check all elements and find common type
	var elementType Type
	for i, elem := range expr.Elements {
		elemType := c.CheckExpression(elem)
		if elemType == nil {
			return nil
		}
		
		if i == 0 {
			elementType = elemType
		} else {
			// Try to find a common type
			// For now, if types are assignable, use the first type
			// In the future, we might want more sophisticated type unification
			if !IsAssignable(elemType, elementType) && !IsAssignable(elementType, elemType) {
				c.addError(expr.Pos(), "list elements have incompatible types: %s and %s",
					elementType.String(), elemType.String())
				return nil
			}
		}
	}
	
	return &List{Element: elementType}
}

