package types

import (
	"testing"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestInferLetBindingType(t *testing.T) {
	env := NewEnvironment()

	tests := []struct {
		name     string
		value    ast.Expr
		expected string
	}{
		{"int literal", &ast.BasicLit{Kind: ast.IntLit, Value: "42"}, "int"},
		{"float literal", &ast.BasicLit{Kind: ast.FloatLit, Value: "3.14"}, "float"},
		{"string literal", &ast.BasicLit{Kind: ast.StringLit, Value: "hello"}, "str"},
		{"bool literal", &ast.BasicLit{Kind: ast.BoolLit, Value: "true"}, "bool"},
		{"binary expression", &ast.BinaryExpr{
			Op:    ast.Add,
			Left:  &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
			Right: &ast.BasicLit{Kind: ast.IntLit, Value: "2"},
		}, "int"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			letExpr := &ast.LetExpr{
				Name:  "x",
				Value: tt.value,
				Body:  &ast.Ident{Name: "x"},
			}

			typ := InferLetBindingType(letExpr, env)
			if typ == nil {
				t.Fatal("expected inferred type")
			}
			if typ.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, typ.String())
			}
		})
	}
}

func TestInferLetBindingWithVariable(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})

	letExpr := &ast.LetExpr{
		Name:  "y",
		Value: &ast.Ident{Name: "x"},
		Body:  &ast.Ident{Name: "y"},
	}

	typ := InferLetBindingType(letExpr, env)
	if typ == nil {
		t.Fatal("expected inferred type")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestInferLetBindingWithExpression(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("y", &Primitive{Kind: Int})

	letExpr := &ast.LetExpr{
		Name: "sum",
		Value: &ast.BinaryExpr{
			Op:    ast.Add,
			Left:  &ast.Ident{Name: "x"},
			Right: &ast.Ident{Name: "y"},
		},
		Body: &ast.Ident{Name: "sum"},
	}

	typ := InferLetBindingType(letExpr, env)
	if typ == nil {
		t.Fatal("expected inferred type")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestInferLetBindingWithComplexType(t *testing.T) {
	env := NewEnvironment()
	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("numbers", listType)

	letExpr := &ast.LetExpr{
		Name:  "list",
		Value: &ast.Ident{Name: "numbers"},
		Body:  &ast.Ident{Name: "list"},
	}

	typ := InferLetBindingType(letExpr, env)
	if typ == nil {
		t.Fatal("expected inferred type")
	}
	if typ.String() != "List<int>" {
		t.Errorf("expected List<int>, got %s", typ.String())
	}
}

func TestInferFunctionReturnType(t *testing.T) {
	env := NewEnvironment()

	// Function with explicit return statement
	funDecl := &ast.FunctionDecl{
		Name: "add",
		Parameters: []ast.Parameter{
			{Name: "x", Type: &ast.PrimitiveType{Name: "int"}},
			{Name: "y", Type: &ast.PrimitiveType{Name: "int"}},
		},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.BinaryExpr{
						Op:    ast.Add,
						Left:  &ast.Ident{Name: "x"},
						Right: &ast.Ident{Name: "y"},
					},
				},
			},
		},
	}

	typ := InferFunctionReturnType(funDecl, env)
	if typ == nil {
		t.Fatal("expected inferred return type")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestInferFunctionReturnTypeWithConditional(t *testing.T) {
	env := NewEnvironment()

	// Function with conditional return
	funDecl := &ast.FunctionDecl{
		Name: "max",
		Parameters: []ast.Parameter{
			{Name: "a", Type: &ast.PrimitiveType{Name: "int"}},
			{Name: "b", Type: &ast.PrimitiveType{Name: "int"}},
		},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.IfExpr{
						Condition: &ast.BinaryExpr{
							Op:    ast.Gt,
							Left:  &ast.Ident{Name: "a"},
							Right: &ast.Ident{Name: "b"},
						},
						Then: &ast.Ident{Name: "a"},
						Else: &ast.Ident{Name: "b"},
					},
				},
			},
		},
	}

	typ := InferFunctionReturnType(funDecl, env)
	if typ == nil {
		t.Fatal("expected inferred return type")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestInferExpressionType(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("y", &Primitive{Kind: Int})

	// Infer type of binary expression
	expr := &ast.BinaryExpr{
		Op:    ast.Add,
		Left:  &ast.Ident{Name: "x"},
		Right: &ast.Ident{Name: "y"},
	}

	typ := InferExpressionType(expr, env)
	if typ == nil {
		t.Fatal("expected inferred type")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestInferTypeWithComplexExpression(t *testing.T) {
	env := NewEnvironment()
	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("numbers", listType)

	// Infer type of index expression
	expr := &ast.IndexExpr{
		X:     &ast.Ident{Name: "numbers"},
		Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
	}

	typ := InferExpressionType(expr, env)
	if typ == nil {
		t.Fatal("expected inferred type")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestInferTypeWithFunctionCall(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("double", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	// Infer type of function call
	expr := &ast.CallExpr{
		Fun: &ast.Ident{Name: "double"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "5"},
		},
	}

	typ := InferExpressionType(expr, env)
	if typ == nil {
		t.Fatal("expected inferred type")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestInferTypeWithRecordField(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	env.DeclareVariable("point", recordType)

	// Infer type of record field access
	expr := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "point"},
		Sel: "x",
	}

	typ := InferExpressionType(expr, env)
	if typ == nil {
		t.Fatal("expected inferred type")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestInferTypeWithNestedExpressions(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("y", &Primitive{Kind: Int})
	env.DeclareVariable("z", &Primitive{Kind: Int})

	// Infer type of nested expression: (x + y) * z
	expr := &ast.BinaryExpr{
		Op: ast.Mul,
		Left: &ast.ParenExpr{
			X: &ast.BinaryExpr{
				Op:    ast.Add,
				Left:  &ast.Ident{Name: "x"},
				Right: &ast.Ident{Name: "y"},
			},
		},
		Right: &ast.Ident{Name: "z"},
	}

	typ := InferExpressionType(expr, env)
	if typ == nil {
		t.Fatal("expected inferred type")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

