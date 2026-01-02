package types

import (
	"testing"

	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestFunctionCallWithPrimitiveTypes(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("add", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	checker := NewChecker(env)

	// Valid call: add(1, 2)
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "1"},
			&ast.BasicLit{Kind: ast.IntLit, Value: "2"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestFunctionCallWithWrongArgumentCount(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("add", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	checker := NewChecker(env)

	// Too few arguments
	call1 := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "1"},
		},
	}

	typ := checker.CheckExpression(call1)
	if typ != nil {
		t.Error("expected nil for wrong argument count")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for wrong argument count")
	}

	// Too many arguments
	checker2 := NewChecker(env)
	call2 := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "1"},
			&ast.BasicLit{Kind: ast.IntLit, Value: "2"},
			&ast.BasicLit{Kind: ast.IntLit, Value: "3"},
		},
	}

	typ = checker2.CheckExpression(call2)
	if typ != nil {
		t.Error("expected nil for too many arguments")
	}
	if len(checker2.Errors()) == 0 {
		t.Error("expected error for too many arguments")
	}
}

func TestFunctionCallWithWrongArgumentTypes(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("add", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	checker := NewChecker(env)

	// Wrong type: add(1, "string")
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "1"},
			&ast.BasicLit{Kind: ast.StringLit, Value: "string"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ != nil {
		t.Error("expected nil for wrong argument type")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for wrong argument type")
	}
}

func TestFunctionCallWithComplexParameterTypes(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("processList", &FunctionSignature{
		Parameters: []Type{&List{Element: &Primitive{Kind: Int}}},
		Return:     &Primitive{Kind: Int},
	})

	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("numbers", listType)

	checker := NewChecker(env)

	// Valid call: processList(numbers)
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "processList"},
		Args: []ast.Expr{
			&ast.Ident{Name: "numbers"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestFunctionCallWithRecordParameter(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	env.DeclareFunction("getX", &FunctionSignature{
		Parameters: []Type{recordType},
		Return:     &Primitive{Kind: Int},
	})

	env.DeclareVariable("point", recordType)

	checker := NewChecker(env)

	// Valid call: getX(point)
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "getX"},
		Args: []ast.Expr{
			&ast.Ident{Name: "point"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestFunctionCallWithMapParameter(t *testing.T) {
	env := NewEnvironment()
	mapType := &Map{
		Key:   &Primitive{Kind: Str},
		Value: &Primitive{Kind: Int},
	}
	env.DeclareFunction("getValue", &FunctionSignature{
		Parameters: []Type{mapType},
		Return:     &Primitive{Kind: Int},
	})

	env.DeclareVariable("scores", mapType)

	checker := NewChecker(env)

	// Valid call: getValue(scores)
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "getValue"},
		Args: []ast.Expr{
			&ast.Ident{Name: "scores"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestFunctionCallWithMultipleComplexParameters(t *testing.T) {
	env := NewEnvironment()
	listType := &List{Element: &Primitive{Kind: Int}}
	mapType := &Map{
		Key:   &Primitive{Kind: Str},
		Value: &Primitive{Kind: Int},
	}
	env.DeclareFunction("process", &FunctionSignature{
		Parameters: []Type{
			listType,
			mapType,
			&Primitive{Kind: Int},
		},
		Return: &Primitive{Kind: Bool},
	})

	env.DeclareVariable("list", listType)
	env.DeclareVariable("map", mapType)

	checker := NewChecker(env)

	// Valid call: process(list, map, 10)
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "process"},
		Args: []ast.Expr{
			&ast.Ident{Name: "list"},
			&ast.Ident{Name: "map"},
			&ast.BasicLit{Kind: ast.IntLit, Value: "10"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call")
	}
	if typ.String() != "bool" {
		t.Errorf("expected bool, got %s", typ.String())
	}
}

func TestFunctionCallReturnTypeUsage(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("getValue", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	env.DeclareVariable("x", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Use return value in assignment: x = getValue(10)
	assign := &ast.AssignStmt{
		Left: &ast.Ident{Name: "x"},
		Right: &ast.CallExpr{
			Fun: &ast.Ident{Name: "getValue"},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: ast.IntLit, Value: "10"},
			},
		},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment with function return value")
	}
}

func TestFunctionCallReturnTypeInExpression(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("double", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	checker := NewChecker(env)

	// Use return value in expression: double(5) + double(3)
	expr := &ast.BinaryExpr{
		Op: ast.Add,
		Left: &ast.CallExpr{
			Fun: &ast.Ident{Name: "double"},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: ast.IntLit, Value: "5"},
			},
		},
		Right: &ast.CallExpr{
			Fun: &ast.Ident{Name: "double"},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: ast.IntLit, Value: "3"},
			},
		},
	}

	typ := checker.CheckExpression(expr)
	if typ == nil {
		t.Fatal("expected type for expression with function calls")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestFunctionCallWithNestedCalls(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("add", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})
	env.DeclareFunction("multiply", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	checker := NewChecker(env)

	// Nested calls: multiply(add(1, 2), add(3, 4))
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "multiply"},
		Args: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{Name: "add"},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: ast.IntLit, Value: "1"},
					&ast.BasicLit{Kind: ast.IntLit, Value: "2"},
				},
			},
			&ast.CallExpr{
				Fun: &ast.Ident{Name: "add"},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: ast.IntLit, Value: "3"},
					&ast.BasicLit{Kind: ast.IntLit, Value: "4"},
				},
			},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for nested function calls")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestFunctionCallWithOptionReturnType(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("find", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}},
		Return:     &Option{Element: &Primitive{Kind: Int}},
	})

	checker := NewChecker(env)

	// Call function returning Option<int>
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "find"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "10"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call")
	}
	if typ.String() != "Option<int>" {
		t.Errorf("expected Option<int>, got %s", typ.String())
	}
}

func TestFunctionCallWithListReturnType(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("getNumbers", &FunctionSignature{
		Parameters: []Type{},
		Return:     &List{Element: &Primitive{Kind: Int}},
	})

	checker := NewChecker(env)

	// Call function returning List<int>
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "getNumbers"},
		Args: []ast.Expr{},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call")
	}
	if typ.String() != "List<int>" {
		t.Errorf("expected List<int>, got %s", typ.String())
	}
}

func TestFunctionCallWithRecordReturnType(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	env.DeclareFunction("createPoint", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     recordType,
	})

	checker := NewChecker(env)

	// Call function returning record
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "createPoint"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "10"},
			&ast.BasicLit{Kind: ast.IntLit, Value: "20"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call")
	}
	if _, ok := typ.(*Record); !ok {
		t.Errorf("expected Record type, got %T", typ)
	}
}

func TestFunctionCallWithTypePromotion(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("processFloat", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Float}},
		Return:     &Primitive{Kind: Float},
	})

	checker := NewChecker(env)

	// Call with int (should promote to float): processFloat(10)
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "processFloat"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "10"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call with type promotion")
	}
	if typ.String() != "float" {
		t.Errorf("expected float, got %s", typ.String())
	}
}

func TestUndefinedFunctionCall(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	// Call undefined function
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "undefined"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "10"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ != nil {
		t.Error("expected nil for undefined function")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for undefined function")
	}
}

func TestFunctionCallWithVariableArguments(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("y", &Primitive{Kind: Int})
	env.DeclareFunction("add", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	checker := NewChecker(env)

	// Call with variables: add(x, y)
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.Ident{Name: "x"},
			&ast.Ident{Name: "y"},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call with variables")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestFunctionCallWithExpressionArguments(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("y", &Primitive{Kind: Int})
	env.DeclareFunction("add", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	checker := NewChecker(env)

	// Call with expressions: add(x + 1, y * 2)
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BinaryExpr{
				Op:    ast.Add,
				Left:  &ast.Ident{Name: "x"},
				Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
			},
			&ast.BinaryExpr{
				Op:    ast.Mul,
				Left:  &ast.Ident{Name: "y"},
				Right: &ast.BasicLit{Kind: ast.IntLit, Value: "2"},
			},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call with expressions")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

