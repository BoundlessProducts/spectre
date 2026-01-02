package eval

import (
	"fmt"

	"github.com/spectre-lang/spectre/internal/state"
	"github.com/spectre-lang/spectre/pkg/ast"
)

// Evaluator evaluates expressions in a function context
type Evaluator struct {
	env *Environment
}

// NewEvaluator creates a new evaluator with the given environment
func NewEvaluator(env *Environment) *Evaluator {
	return &Evaluator{
		env: env,
	}
}

// Eval evaluates an expression and returns its value
func (e *Evaluator) Eval(expr ast.Expr) (state.Value, error) {
	if expr == nil {
		return nil, fmt.Errorf("cannot evaluate nil expression")
	}

	switch ex := expr.(type) {
	case *ast.BasicLit:
		return e.evalBasicLit(ex)
	case *ast.Ident:
		return e.evalIdent(ex)
	case *ast.BinaryExpr:
		return e.evalBinaryExpr(ex)
	case *ast.UnaryExpr:
		return e.evalUnaryExpr(ex)
	case *ast.ParenExpr:
		return e.evalParenExpr(ex)
	case *ast.CallExpr:
		return e.evalCallExpr(ex)
	case *ast.IfExpr:
		return e.evalIfExpr(ex)
	case *ast.LetExpr:
		return e.evalLetExpr(ex)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// evalBasicLit evaluates a basic literal
func (e *Evaluator) evalBasicLit(lit *ast.BasicLit) (state.Value, error) {
	switch lit.Kind {
	case ast.IntLit:
		// Parse integer value
		var val int64
		_, err := fmt.Sscanf(lit.Value, "%d", &val)
		if err != nil {
			return nil, fmt.Errorf("invalid integer literal: %s", lit.Value)
		}
		return state.NewIntValue(val), nil
	case ast.FloatLit:
		// Parse float value
		var val float64
		_, err := fmt.Sscanf(lit.Value, "%f", &val)
		if err != nil {
			return nil, fmt.Errorf("invalid float literal: %s", lit.Value)
		}
		return state.NewFloatValue(val), nil
	case ast.StringLit:
		// Remove quotes from string literal
		str := lit.Value
		if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
			str = str[1 : len(str)-1]
		}
		return state.NewStringValue(str), nil
	case ast.BoolLit:
		// Parse boolean value
		if lit.Value == "true" {
			return state.NewBoolValue(true), nil
		} else if lit.Value == "false" {
			return state.NewBoolValue(false), nil
		}
		return nil, fmt.Errorf("invalid boolean literal: %s", lit.Value)
	default:
		return nil, fmt.Errorf("unsupported literal type: %v", lit.Kind)
	}
}

// evalIdent evaluates an identifier (variable, constant, or function name)
func (e *Evaluator) evalIdent(ident *ast.Ident) (state.Value, error) {
	name := ident.Name

	// Check if it's a variable
	if value, exists := e.env.GetVariable(name); exists {
		return value, nil
	}

	// Check if it's a constant
	if value, exists := e.env.GetConstant(name); exists {
		return value, nil
	}

	return nil, fmt.Errorf("undefined identifier: %s", name)
}

// evalBinaryExpr evaluates a binary expression
func (e *Evaluator) evalBinaryExpr(expr *ast.BinaryExpr) (state.Value, error) {
	left, err := e.Eval(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := e.Eval(expr.Right)
	if err != nil {
		return nil, err
	}

	return e.evalBinaryOp(expr.Op, left, right)
}

// evalBinaryOp evaluates a binary operation
func (e *Evaluator) evalBinaryOp(op ast.BinaryOp, left, right state.Value) (state.Value, error) {
	// For now, support primitive operations only
	leftPrim, leftOk := left.(*state.PrimitiveValue)
	rightPrim, rightOk := right.(*state.PrimitiveValue)

	if !leftOk || !rightOk {
		return nil, fmt.Errorf("binary operations only supported for primitive types")
	}

	switch op {
	case ast.Add:
		return e.evalAdd(leftPrim, rightPrim)
	case ast.Sub:
		return e.evalSub(leftPrim, rightPrim)
	case ast.Mul:
		return e.evalMul(leftPrim, rightPrim)
	case ast.Div:
		return e.evalDiv(leftPrim, rightPrim)
	case ast.Eq:
		return e.evalEq(leftPrim, rightPrim)
	case ast.Neq:
		return e.evalNe(leftPrim, rightPrim)
	case ast.Lt:
		return e.evalLt(leftPrim, rightPrim)
	case ast.Leq:
		return e.evalLe(leftPrim, rightPrim)
	case ast.Gt:
		return e.evalGt(leftPrim, rightPrim)
	case ast.Geq:
		return e.evalGe(leftPrim, rightPrim)
	case ast.And:
		return e.evalAnd(leftPrim, rightPrim)
	case ast.Or:
		return e.evalOr(leftPrim, rightPrim)
	default:
		return nil, fmt.Errorf("unsupported binary operator: %v", op)
	}
}

// evalAdd evaluates addition
func (e *Evaluator) evalAdd(left, right *state.PrimitiveValue) (state.Value, error) {
	if left.TypeName == "int" && right.TypeName == "int" {
		leftVal := int64(0)
		rightVal := int64(0)
		if left.IntValue != nil {
			leftVal = *left.IntValue
		}
		if right.IntValue != nil {
			rightVal = *right.IntValue
		}
		return state.NewIntValue(leftVal + rightVal), nil
	}
	if left.TypeName == "float" && right.TypeName == "float" {
		leftVal := float64(0)
		rightVal := float64(0)
		if left.FloatValue != nil {
			leftVal = *left.FloatValue
		}
		if right.FloatValue != nil {
			rightVal = *right.FloatValue
		}
		return state.NewFloatValue(leftVal + rightVal), nil
	}
	if left.TypeName == "str" && right.TypeName == "str" {
		leftVal := ""
		rightVal := ""
		if left.StringValue != nil {
			leftVal = *left.StringValue
		}
		if right.StringValue != nil {
			rightVal = *right.StringValue
		}
		return state.NewStringValue(leftVal + rightVal), nil
	}
	return nil, fmt.Errorf("addition not supported for types %s and %s", left.TypeName, right.TypeName)
}

// evalSub evaluates subtraction
func (e *Evaluator) evalSub(left, right *state.PrimitiveValue) (state.Value, error) {
	if left.TypeName == "int" && right.TypeName == "int" {
		leftVal := int64(0)
		rightVal := int64(0)
		if left.IntValue != nil {
			leftVal = *left.IntValue
		}
		if right.IntValue != nil {
			rightVal = *right.IntValue
		}
		return state.NewIntValue(leftVal - rightVal), nil
	}
	if left.TypeName == "float" && right.TypeName == "float" {
		leftVal := float64(0)
		rightVal := float64(0)
		if left.FloatValue != nil {
			leftVal = *left.FloatValue
		}
		if right.FloatValue != nil {
			rightVal = *right.FloatValue
		}
		return state.NewFloatValue(leftVal - rightVal), nil
	}
	return nil, fmt.Errorf("subtraction not supported for types %s and %s", left.TypeName, right.TypeName)
}

// evalMul evaluates multiplication
func (e *Evaluator) evalMul(left, right *state.PrimitiveValue) (state.Value, error) {
	if left.TypeName == "int" && right.TypeName == "int" {
		leftVal := int64(0)
		rightVal := int64(0)
		if left.IntValue != nil {
			leftVal = *left.IntValue
		}
		if right.IntValue != nil {
			rightVal = *right.IntValue
		}
		return state.NewIntValue(leftVal * rightVal), nil
	}
	if left.TypeName == "float" && right.TypeName == "float" {
		leftVal := float64(0)
		rightVal := float64(0)
		if left.FloatValue != nil {
			leftVal = *left.FloatValue
		}
		if right.FloatValue != nil {
			rightVal = *right.FloatValue
		}
		return state.NewFloatValue(leftVal * rightVal), nil
	}
	return nil, fmt.Errorf("multiplication not supported for types %s and %s", left.TypeName, right.TypeName)
}

// evalDiv evaluates division
func (e *Evaluator) evalDiv(left, right *state.PrimitiveValue) (state.Value, error) {
	if left.TypeName == "int" && right.TypeName == "int" {
		leftVal := int64(0)
		rightVal := int64(0)
		if left.IntValue != nil {
			leftVal = *left.IntValue
		}
		if right.IntValue != nil {
			rightVal = *right.IntValue
		}
		if rightVal == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return state.NewIntValue(leftVal / rightVal), nil
	}
	if left.TypeName == "float" && right.TypeName == "float" {
		leftVal := float64(0)
		rightVal := float64(0)
		if left.FloatValue != nil {
			leftVal = *left.FloatValue
		}
		if right.FloatValue != nil {
			rightVal = *right.FloatValue
		}
		if rightVal == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return state.NewFloatValue(leftVal / rightVal), nil
	}
	return nil, fmt.Errorf("division not supported for types %s and %s", left.TypeName, right.TypeName)
}

// evalEq evaluates equality
func (e *Evaluator) evalEq(left, right *state.PrimitiveValue) (state.Value, error) {
	if left.TypeName != right.TypeName {
		return state.NewBoolValue(false), nil
	}
	equal := e.primitiveValuesEqual(left, right)
	return state.NewBoolValue(equal), nil
}

// primitiveValuesEqual checks if two primitive values are equal
func (e *Evaluator) primitiveValuesEqual(v1, v2 *state.PrimitiveValue) bool {
	if v1.TypeName != v2.TypeName {
		return false
	}

	switch v1.TypeName {
	case "int":
		val1 := int64(0)
		val2 := int64(0)
		if v1.IntValue != nil {
			val1 = *v1.IntValue
		}
		if v2.IntValue != nil {
			val2 = *v2.IntValue
		}
		return val1 == val2
	case "float":
		val1 := float64(0)
		val2 := float64(0)
		if v1.FloatValue != nil {
			val1 = *v1.FloatValue
		}
		if v2.FloatValue != nil {
			val2 = *v2.FloatValue
		}
		return val1 == val2
	case "str":
		val1 := ""
		val2 := ""
		if v1.StringValue != nil {
			val1 = *v1.StringValue
		}
		if v2.StringValue != nil {
			val2 = *v2.StringValue
		}
		return val1 == val2
	case "bool":
		val1 := false
		val2 := false
		if v1.BoolValue != nil {
			val1 = *v1.BoolValue
		}
		if v2.BoolValue != nil {
			val2 = *v2.BoolValue
		}
		return val1 == val2
	default:
		return false
	}
}

// evalNe evaluates inequality
func (e *Evaluator) evalNe(left, right *state.PrimitiveValue) (state.Value, error) {
	result, err := e.evalEq(left, right)
	if err != nil {
		return nil, err
	}
	if boolVal, ok := result.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil {
		return state.NewBoolValue(!*boolVal.BoolValue), nil
	}
	return state.NewBoolValue(true), nil
}

// evalLt evaluates less than
func (e *Evaluator) evalLt(left, right *state.PrimitiveValue) (state.Value, error) {
	if left.TypeName == "int" && right.TypeName == "int" {
		leftVal := int64(0)
		rightVal := int64(0)
		if left.IntValue != nil {
			leftVal = *left.IntValue
		}
		if right.IntValue != nil {
			rightVal = *right.IntValue
		}
		return state.NewBoolValue(leftVal < rightVal), nil
	}
	if left.TypeName == "float" && right.TypeName == "float" {
		leftVal := float64(0)
		rightVal := float64(0)
		if left.FloatValue != nil {
			leftVal = *left.FloatValue
		}
		if right.FloatValue != nil {
			rightVal = *right.FloatValue
		}
		return state.NewBoolValue(leftVal < rightVal), nil
	}
	return nil, fmt.Errorf("less than not supported for types %s and %s", left.TypeName, right.TypeName)
}

// evalLe evaluates less than or equal
func (e *Evaluator) evalLe(left, right *state.PrimitiveValue) (state.Value, error) {
	lt, err := e.evalLt(left, right)
	if err != nil {
		return nil, err
	}
	eq, err := e.evalEq(left, right)
	if err != nil {
		return nil, err
	}
	if ltBool, ok := lt.(*state.PrimitiveValue); ok && ltBool.BoolValue != nil {
		if eqBool, ok := eq.(*state.PrimitiveValue); ok && eqBool.BoolValue != nil {
			return state.NewBoolValue(*ltBool.BoolValue || *eqBool.BoolValue), nil
		}
	}
	return nil, fmt.Errorf("error evaluating <= operator")
}

// evalGt evaluates greater than
func (e *Evaluator) evalGt(left, right *state.PrimitiveValue) (state.Value, error) {
	le, err := e.evalLe(left, right)
	if err != nil {
		return nil, err
	}
	if boolVal, ok := le.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil {
		return state.NewBoolValue(!*boolVal.BoolValue), nil
	}
	return nil, fmt.Errorf("error evaluating > operator")
}

// evalGe evaluates greater than or equal
func (e *Evaluator) evalGe(left, right *state.PrimitiveValue) (state.Value, error) {
	lt, err := e.evalLt(left, right)
	if err != nil {
		return nil, err
	}
	if boolVal, ok := lt.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil {
		return state.NewBoolValue(!*boolVal.BoolValue), nil
	}
	return nil, fmt.Errorf("error evaluating >= operator")
}

// evalAnd evaluates logical AND
func (e *Evaluator) evalAnd(left, right *state.PrimitiveValue) (state.Value, error) {
	if left.TypeName != "bool" || right.TypeName != "bool" {
		return nil, fmt.Errorf("AND operator requires boolean operands")
	}
	leftVal := false
	rightVal := false
	if left.BoolValue != nil {
		leftVal = *left.BoolValue
	}
	if right.BoolValue != nil {
		rightVal = *right.BoolValue
	}
	return state.NewBoolValue(leftVal && rightVal), nil
}

// evalOr evaluates logical OR
func (e *Evaluator) evalOr(left, right *state.PrimitiveValue) (state.Value, error) {
	if left.TypeName != "bool" || right.TypeName != "bool" {
		return nil, fmt.Errorf("OR operator requires boolean operands")
	}
	leftVal := false
	rightVal := false
	if left.BoolValue != nil {
		leftVal = *left.BoolValue
	}
	if right.BoolValue != nil {
		rightVal = *right.BoolValue
	}
	return state.NewBoolValue(leftVal || rightVal), nil
}

// evalUnaryExpr evaluates a unary expression
func (e *Evaluator) evalUnaryExpr(expr *ast.UnaryExpr) (state.Value, error) {
	operand, err := e.Eval(expr.Expr)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case ast.Not:
		return e.evalNot(operand)
	case ast.Neg:
		return e.evalMinus(operand)
	default:
		return nil, fmt.Errorf("unsupported unary operator: %v", expr.Op)
	}
}

// evalNot evaluates logical NOT
func (e *Evaluator) evalNot(operand state.Value) (state.Value, error) {
	if prim, ok := operand.(*state.PrimitiveValue); ok && prim.TypeName == "bool" {
		val := false
		if prim.BoolValue != nil {
			val = *prim.BoolValue
		}
		return state.NewBoolValue(!val), nil
	}
	return nil, fmt.Errorf("NOT operator requires boolean operand")
}

// evalMinus evaluates unary minus
func (e *Evaluator) evalMinus(operand state.Value) (state.Value, error) {
	if prim, ok := operand.(*state.PrimitiveValue); ok {
		if prim.TypeName == "int" {
			val := int64(0)
			if prim.IntValue != nil {
				val = *prim.IntValue
			}
			return state.NewIntValue(-val), nil
		}
		if prim.TypeName == "float" {
			val := float64(0)
			if prim.FloatValue != nil {
				val = *prim.FloatValue
			}
			return state.NewFloatValue(-val), nil
		}
	}
	return nil, fmt.Errorf("unary minus requires numeric operand")
}

// evalParenExpr evaluates a parenthesized expression
func (e *Evaluator) evalParenExpr(expr *ast.ParenExpr) (state.Value, error) {
	return e.Eval(expr.X)
}

// evalCallExpr evaluates a function call
func (e *Evaluator) evalCallExpr(expr *ast.CallExpr) (state.Value, error) {
	// Get function name
	var funcName string
	switch fun := expr.Fun.(type) {
	case *ast.Ident:
		funcName = fun.Name
	case *ast.SelectorExpr:
		// For now, only support simple function calls
		// Method calls (selector expressions) will be handled later
		return nil, fmt.Errorf("method calls not yet supported")
	default:
		return nil, fmt.Errorf("unsupported function call expression type: %T", expr.Fun)
	}

	// Get function definition
	fnDef, exists := e.env.GetFunction(funcName)
	if !exists {
		return nil, fmt.Errorf("undefined function: %s", funcName)
	}

	// Evaluate arguments
	args := make([]state.Value, len(expr.Args))
	for i, arg := range expr.Args {
		val, err := e.Eval(arg)
		if err != nil {
			return nil, fmt.Errorf("error evaluating argument %d: %w", i, err)
		}
		args[i] = val
	}

	// Check argument count
	if len(args) != len(fnDef.Params) {
		return nil, fmt.Errorf("function %s expects %d arguments, got %d", funcName, len(fnDef.Params), len(args))
	}

	// Create new scope for function execution
	funcEnv := e.env.EnterScope()

	// Bind parameters to argument values
	for i, param := range fnDef.Params {
		funcEnv.SetVariable(param.Name, args[i])
	}

	// Evaluate function body
	evaluator := NewEvaluator(funcEnv)
	return evaluator.evalFunctionBody(fnDef.Body)
}

// evalFunctionBody evaluates a function body and returns the return value
func (e *Evaluator) evalFunctionBody(body *ast.BlockStmt) (state.Value, error) {
	if body == nil {
		return nil, fmt.Errorf("function body is nil")
	}

	if len(body.Statements) == 0 {
		return nil, fmt.Errorf("function body is empty")
	}

	// In Spectre, function bodies can have:
	// 1. A return statement
	// 2. An expression statement (the expression is the return value)
	// 3. An if expression (the result is the return value)
	
	// Get the last statement (or only statement)
	lastStmt := body.Statements[len(body.Statements)-1]

	switch s := lastStmt.(type) {
	case *ast.ReturnStmt:
		if s.Value == nil {
			return nil, fmt.Errorf("return statement must have a value")
		}
		return e.Eval(s.Value)
	case *ast.ExprStmt:
		// Expression statement - evaluate and return the value
		return e.Eval(s.Expr)
	default:
		return nil, fmt.Errorf("function body must end with an expression or return statement, got %T", lastStmt)
	}
}

// evalIfExpr evaluates an if expression
func (e *Evaluator) evalIfExpr(expr *ast.IfExpr) (state.Value, error) {
	// Evaluate condition
	cond, err := e.Eval(expr.Condition)
	if err != nil {
		return nil, err
	}

	// Check if condition is true
	isTrue := false
	if prim, ok := cond.(*state.PrimitiveValue); ok && prim.TypeName == "bool" {
		if prim.BoolValue != nil {
			isTrue = *prim.BoolValue
		}
	} else {
		return nil, fmt.Errorf("if condition must evaluate to boolean")
	}

	// Evaluate then or else branch
	if isTrue {
		return e.Eval(expr.Then)
	} else {
		if expr.Else == nil {
			return nil, fmt.Errorf("if expression without else branch")
		}
		return e.Eval(expr.Else)
	}
}

// evalLetExpr evaluates a let expression
func (e *Evaluator) evalLetExpr(expr *ast.LetExpr) (state.Value, error) {
	// Evaluate the value expression
	value, err := e.Eval(expr.Value)
	if err != nil {
		return nil, err
	}

	// Enter new scope
	letEnv := e.env.EnterScope()

	// Bind the variable
	letEnv.SetVariable(expr.Name, value)

	// Evaluate body in new scope
	evaluator := NewEvaluator(letEnv)
	return evaluator.Eval(expr.Body)
}

