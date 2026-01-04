package eval

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
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
	case *ast.SelectorExpr:
		return e.evalSelectorExpr(ex)
	case *ast.IfExpr:
		return e.evalIfExpr(ex)
	case *ast.LetExpr:
		return e.evalLetExpr(ex)
	case *ast.LambdaExpr:
		return e.evalLambdaExpr(ex)
	case *ast.IndexExpr:
		return e.evalIndexExpr(ex)
	case *ast.SetLiteral:
		return e.evalSetLiteral(ex)
	case *ast.ListLiteral:
		return e.evalListLiteral(ex)
	case *ast.RecordLiteral:
		return e.evalRecordLiteral(ex)
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
	// Handle equality/inequality for enum values
	if op == ast.Eq || op == ast.Neq {
		leftEnum, leftEnumOk := left.(*state.EnumValue)
		rightEnum, rightEnumOk := right.(*state.EnumValue)
		
		if leftEnumOk && rightEnumOk {
			// Both are enum values - compare enum name and value name
			equal := leftEnum.EnumName == rightEnum.EnumName && leftEnum.ValueName == rightEnum.ValueName
			if op == ast.Eq {
				return state.NewBoolValue(equal), nil
			} else {
				return state.NewBoolValue(!equal), nil
			}
		}
	}

	// For arithmetic and other operations, support primitive operations
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
		// Check if it's a static method call (Set.empty(), Set.of(), etc.)
		if ident, ok := fun.X.(*ast.Ident); ok {
			if ident.Name == "Set" || ident.Name == "List" || ident.Name == "Map" || ident.Name == "Option" {
				// Static method call on collection type
				return e.evalStaticMethodCall(ident.Name, fun.Sel, expr.Args)
			}
		}
		
		// Instance method call (e.g., users.filter(...))
		obj, err := e.Eval(fun.X)
		if err != nil {
			return nil, fmt.Errorf("error evaluating object in method call: %w", err)
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
		
		// Evaluate method call with the object
		return e.evalMethodCall(obj, fun.Sel, args)
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

// evalSelectorExpr evaluates a selector expression (e.g., obj.field or obj.method or Enum.Value)
func (e *Evaluator) evalSelectorExpr(expr *ast.SelectorExpr) (state.Value, error) {
	// Check if it's an enum value access (e.g., ProcessState.Idle)
	if ident, ok := expr.X.(*ast.Ident); ok {
		// Check if it's an enum type name
		enumDef, exists := e.env.GetEnumType(ident.Name)
		if exists {
			// Verify that the selector is a valid enum value
			valid := false
			for _, val := range enumDef.Values {
				if val == expr.Sel {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("enum value %s.%s not found in enum %s", ident.Name, expr.Sel, ident.Name)
			}
			// Return the enum value
			return state.NewEnumValue(ident.Name, expr.Sel), nil
		}
	}

	// Evaluate the object (for method calls or field access)
	obj, err := e.Eval(expr.X)
	if err != nil {
		return nil, fmt.Errorf("error evaluating object: %w", err)
	}
	
	// Check if this is record field access (record.field)
	// Records are stored as MapValue with string keys
	if mapVal, ok := obj.(*state.MapValue); ok {
		// This is a record - extract the field value
		fieldKey := state.NewStringValue(expr.Sel)
		fieldValue, exists := mapVal.Get(fieldKey)
		if !exists {
			return nil, fmt.Errorf("field %s not found in record", expr.Sel)
		}
		return fieldValue, nil
	}
	
	// Otherwise, this is a method call - return the selector expression wrapped so it can be called
	return &SelectorValue{
		Object: obj,
		Method:  expr.Sel,
	}, nil
}

// SelectorValue represents a method selector that can be called
type SelectorValue struct {
	Object state.Value
	Method string
}

func (v *SelectorValue) Type() string {
	return "selector"
}

func (v *SelectorValue) String() string {
	return fmt.Sprintf("%s.%s", v.Object.String(), v.Method)
}

// evalMethodCall evaluates a method call on an object
func (e *Evaluator) evalMethodCall(obj state.Value, methodName string, args []state.Value) (state.Value, error) {
	// Handle collection methods
	switch methodName {
	case "filter":
		return e.evalFilter(obj, args)
	case "map":
		return e.evalMap(obj, args)
	case "reduce":
		return e.evalReduce(obj, args)
	case "forall":
		return e.evalForall(obj, args)
	case "exists":
		return e.evalExists(obj, args)
	case "size":
		return e.evalSize(obj)
	case "contains":
		return e.evalContains(obj, args)
	case "union":
		return e.evalUnion(obj, args)
	case "intersection":
		return e.evalIntersection(obj, args)
	case "head":
		return e.evalHead(obj)
	case "tail":
		return e.evalTail(obj)
	case "toList":
		return e.evalToList(obj)
	case "toSet":
		return e.evalToSet(obj)
	case "append":
		return e.evalAppend(obj, args)
	case "put":
		return e.evalPut(obj, args)
	case "get":
		return e.evalGet(obj, args)
	default:
		return nil, fmt.Errorf("unknown method: %s", methodName)
	}
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

// LambdaValue represents a lambda function value
// This is used to pass lambdas as values (e.g., to filter, map, etc.)
type LambdaValue struct {
	Params []ast.Parameter // Lambda parameters
	Body   ast.Expr        // Lambda body expression
	Env    *Environment    // Captured environment (closure)
}

func (v *LambdaValue) Type() string {
	return "lambda"
}

func (v *LambdaValue) String() string {
	if len(v.Params) == 1 {
		return fmt.Sprintf("(%s => ...)", v.Params[0].Name)
	}
	paramNames := make([]string, len(v.Params))
	for i, p := range v.Params {
		paramNames[i] = p.Name
	}
	return fmt.Sprintf("((%s) => ...)", fmt.Sprint(paramNames))
}

// Call executes the lambda with the given arguments
func (v *LambdaValue) Call(args []state.Value) (state.Value, error) {
	if len(args) != len(v.Params) {
		return nil, fmt.Errorf("lambda expects %d arguments, got %d", len(v.Params), len(args))
	}
	
	// Create new scope for lambda execution
	lambdaEnv := NewChildEnvironment(v.Env)
	
	// Bind parameters to argument values
	for i, param := range v.Params {
		lambdaEnv.SetVariable(param.Name, args[i])
	}
	
	// Evaluate lambda body in new scope
	evaluator := NewEvaluator(lambdaEnv)
	return evaluator.Eval(v.Body)
}

// evalLambdaExpr evaluates a lambda expression
// Returns a LambdaValue that can be called later
func (e *Evaluator) evalLambdaExpr(expr *ast.LambdaExpr) (state.Value, error) {
	// Create a lambda value that captures the current environment
	// and the lambda AST node
	return &LambdaValue{
		Params: expr.Params,
		Body:   expr.Body,
		Env:    e.env, // Capture current environment (closure)
	}, nil
}

// evalStaticMethodCall evaluates static method calls like Set.empty(), Set.of(x), etc.
func (e *Evaluator) evalStaticMethodCall(collectionType string, methodName string, args []ast.Expr) (state.Value, error) {
	// Evaluate arguments
	argValues := make([]state.Value, len(args))
	for i, arg := range args {
		val, err := e.Eval(arg)
		if err != nil {
			return nil, fmt.Errorf("error evaluating argument %d: %w", i, err)
		}
		argValues[i] = val
	}
	
	switch methodName {
	case "empty":
		return e.evalEmptyConstructor(collectionType, argValues)
	case "of":
		return e.evalOfConstructor(collectionType, argValues)
	default:
		return nil, fmt.Errorf("unknown static method %s.%s", collectionType, methodName)
	}
}

// evalEmptyConstructor evaluates Set.empty(), List.empty(), Map.empty()
func (e *Evaluator) evalEmptyConstructor(collectionType string, args []state.Value) (state.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("%s.empty() expects 0 arguments, got %d", collectionType, len(args))
	}
	
	switch collectionType {
	case "Set":
		return state.NewSetValue(), nil
	case "List":
		return state.NewListValue(), nil
	case "Map":
		return state.NewMapValue(), nil
	default:
		return nil, fmt.Errorf("unknown collection type: %s", collectionType)
	}
}

// evalOfConstructor evaluates Set.of(x), List.of(x)
func (e *Evaluator) evalOfConstructor(collectionType string, args []state.Value) (state.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%s.of() expects 1 argument, got %d", collectionType, len(args))
	}
	
	elem := args[0]
	
	switch collectionType {
	case "Set":
		result := state.NewSetValue()
		result.Add(elem)
		return result, nil
	case "List":
		result := state.NewListValue()
		result.Append(elem)
		return result, nil
	default:
		return nil, fmt.Errorf("%s.of() not supported", collectionType)
	}
}

