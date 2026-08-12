package types

import (
	"testing"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestCheckBasicLit(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	tests := []struct {
		name     string
		lit      *ast.BasicLit
		expected string
	}{
		{"int", &ast.BasicLit{Kind: ast.IntLit, Value: "42"}, "int"},
		{"float", &ast.BasicLit{Kind: ast.FloatLit, Value: "3.14"}, "float"},
		{"string", &ast.BasicLit{Kind: ast.StringLit, Value: "hello"}, "str"},
		{"bool", &ast.BasicLit{Kind: ast.BoolLit, Value: "true"}, "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := checker.CheckExpression(tt.lit)
			if typ == nil {
				t.Fatal("expected type, got nil")
			}
			if typ.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, typ.String())
			}
			if len(checker.Errors()) != 0 {
				t.Errorf("unexpected errors: %v", checker.Errors())
			}
		})
	}
}

func TestCheckIdent(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareConstant("MAX", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Check variable
	ident := &ast.Ident{Name: "x"}
	typ := checker.CheckExpression(ident)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Check constant
	ident = &ast.Ident{Name: "MAX"}
	typ = checker.CheckExpression(ident)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Check undefined
	ident = &ast.Ident{Name: "undefined"}
	typ = checker.CheckExpression(ident)
	if typ != nil {
		t.Error("expected nil for undefined identifier")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for undefined identifier")
	}
}

func TestCheckBinaryExprArithmetic(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	tests := []struct {
		name     string
		left     ast.Expr
		right    ast.Expr
		op       ast.BinaryOp
		expected string
	}{
		{"int + int", &ast.BasicLit{Kind: ast.IntLit}, &ast.BasicLit{Kind: ast.IntLit}, ast.Add, "int"},
		{"int + float", &ast.BasicLit{Kind: ast.IntLit}, &ast.BasicLit{Kind: ast.FloatLit}, ast.Add, "float"},
		{"float + int", &ast.BasicLit{Kind: ast.FloatLit}, &ast.BasicLit{Kind: ast.IntLit}, ast.Add, "float"},
		{"float + float", &ast.BasicLit{Kind: ast.FloatLit}, &ast.BasicLit{Kind: ast.FloatLit}, ast.Add, "float"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.BinaryExpr{
				Op:    tt.op,
				Left:  tt.left,
				Right: tt.right,
			}
			typ := checker.CheckExpression(expr)
			if typ == nil {
				t.Fatal("expected type, got nil")
			}
			if typ.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, typ.String())
			}
		})
	}
}

func TestCheckBinaryExprComparison(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	expr := &ast.BinaryExpr{
		Op:    ast.Lt,
		Left:  &ast.BasicLit{Kind: ast.IntLit},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	typ := checker.CheckExpression(expr)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "bool" {
		t.Errorf("expected bool, got %s", typ.String())
	}
}

func TestCheckBinaryExprLogical(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	expr := &ast.BinaryExpr{
		Op:    ast.And,
		Left:  &ast.BasicLit{Kind: ast.BoolLit},
		Right: &ast.BasicLit{Kind: ast.BoolLit},
	}

	typ := checker.CheckExpression(expr)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "bool" {
		t.Errorf("expected bool, got %s", typ.String())
	}
}

func TestCheckBinaryExprError(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	// int + bool should error
	expr := &ast.BinaryExpr{
		Op:    ast.Add,
		Left:  &ast.BasicLit{Kind: ast.IntLit},
		Right: &ast.BasicLit{Kind: ast.BoolLit},
	}

	typ := checker.CheckExpression(expr)
	if typ != nil {
		t.Error("expected nil for invalid operation")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for invalid operation")
	}
}

func TestCheckUnaryExpr(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	// Test negation
	negExpr := &ast.UnaryExpr{
		Op:   ast.Neg,
		Expr: &ast.BasicLit{Kind: ast.IntLit},
	}
	typ := checker.CheckExpression(negExpr)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Test logical not
	notExpr := &ast.UnaryExpr{
		Op:   ast.Not,
		Expr: &ast.BasicLit{Kind: ast.BoolLit},
	}
	typ = checker.CheckExpression(notExpr)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "bool" {
		t.Errorf("expected bool, got %s", typ.String())
	}
}

func TestCheckCallExpr(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("add", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	checker := NewChecker(env)

	// Valid call
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit},
			&ast.BasicLit{Kind: ast.IntLit},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Wrong argument count
	call = &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit},
		},
	}

	typ = checker.CheckExpression(call)
	if typ != nil {
		t.Error("expected nil for wrong argument count")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for wrong argument count")
	}
}

func TestCheckSelectorExpr(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	// Create a record type
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}

	// We need to create an expression that evaluates to this record type
	// For testing, we'll use a variable
	env.DeclareVariable("point", recordType)

	// Check field access
	sel := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "point"},
		Sel: "x",
	}

	typ := checker.CheckExpression(sel)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestCheckIndexExpr(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	// Test list indexing
	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("list", listType)

	indexExpr := &ast.IndexExpr{
		X:     &ast.Ident{Name: "list"},
		Index: &ast.BasicLit{Kind: ast.IntLit},
	}

	typ := checker.CheckExpression(indexExpr)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Test map indexing
	mapType := &Map{
		Key:   &Primitive{Kind: Str},
		Value: &Primitive{Kind: Int},
	}
	env.DeclareVariable("map", mapType)

	indexExpr = &ast.IndexExpr{
		X:     &ast.Ident{Name: "map"},
		Index: &ast.BasicLit{Kind: ast.StringLit},
	}

	typ = checker.CheckExpression(indexExpr)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestCheckIfExpr(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	// Valid if-else
	ifExpr := &ast.IfExpr{
		Condition: &ast.BasicLit{Kind: ast.BoolLit},
		Then:      &ast.BasicLit{Kind: ast.IntLit},
		Else:      &ast.BasicLit{Kind: ast.IntLit},
	}

	typ := checker.CheckExpression(ifExpr)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestCheckLetExpr(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	letExpr := &ast.LetExpr{
		Name:  "x",
		Value: &ast.BasicLit{Kind: ast.IntLit},
		Body:  &ast.Ident{Name: "x"},
	}

	typ := checker.CheckExpression(letExpr)
	if typ == nil {
		t.Fatal("expected type, got nil")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

