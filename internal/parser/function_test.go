package parser

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestFunctionDeclaration(t *testing.T) {
	input := `fun add(x: int, y: int): int {
  return x + y
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseFunctionDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	funcDecl, ok := decl.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl not *ast.FunctionDecl. got=%T", decl)
	}

	if funcDecl.Name != "add" {
		t.Errorf("funcDecl.Name not 'add'. got=%s", funcDecl.Name)
	}

	if len(funcDecl.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(funcDecl.Parameters))
	}

	if funcDecl.Parameters[0].Name != "x" {
		t.Errorf("first parameter name not 'x'. got=%s", funcDecl.Parameters[0].Name)
	}

	if funcDecl.Parameters[1].Name != "y" {
		t.Errorf("second parameter name not 'y'. got=%s", funcDecl.Parameters[1].Name)
	}

	returnType, ok := funcDecl.ReturnType.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("returnType not *ast.PrimitiveType. got=%T", funcDecl.ReturnType)
	}

	if returnType.Name != "int" {
		t.Errorf("returnType.Name not 'int'. got=%s", returnType.Name)
	}

	if funcDecl.Body == nil {
		t.Fatal("function body is nil")
	}

	if len(funcDecl.Body.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(funcDecl.Body.Statements))
	}

	returnStmt, ok := funcDecl.Body.Statements[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("first statement not *ast.ReturnStmt. got=%T", funcDecl.Body.Statements[0])
	}

	binExpr, ok := returnStmt.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("return value not *ast.BinaryExpr. got=%T", returnStmt.Value)
	}

	if binExpr.Op != ast.Add {
		t.Errorf("return expression operator not Add. got=%v", binExpr.Op)
	}
}

func TestFunctionDeclarationWithDescription(t *testing.T) {
	input := `description "Adds two integers together"
fun add(x: int, y: int): int {
  return x + y
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseFunctionDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	funcDecl, ok := decl.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl not *ast.FunctionDecl. got=%T", decl)
	}

	if funcDecl.Description != "Adds two integers together" {
		t.Errorf("funcDecl.Description not 'Adds two integers together'. got=%s", funcDecl.Description)
	}
}

func TestFunctionDeclarationNoParameters(t *testing.T) {
	input := `fun getValue(): int {
  return 42
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseFunctionDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	funcDecl, ok := decl.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl not *ast.FunctionDecl. got=%T", decl)
	}

	if len(funcDecl.Parameters) != 0 {
		t.Errorf("expected 0 parameters, got %d", len(funcDecl.Parameters))
	}
}

func TestFunctionDeclarationNoReturnType(t *testing.T) {
	input := `fun multiply(x: int, y: int) {
  return x * y
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseFunctionDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	funcDecl, ok := decl.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl not *ast.FunctionDecl. got=%T", decl)
	}

	if funcDecl.ReturnType != nil {
		t.Errorf("expected nil return type, got %T", funcDecl.ReturnType)
	}
}

func TestFunctionDeclarationWithIfStatement(t *testing.T) {
	// Note: This test currently expects if-else as expressions in function bodies
	// For now, we'll test a simpler case that works with our current implementation
	input := `fun max(a: int, b: int): int {
  return if (a > b) { a } else { b }
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseFunctionDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	funcDecl, ok := decl.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl not *ast.FunctionDecl. got=%T", decl)
	}

	if len(funcDecl.Body.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(funcDecl.Body.Statements))
	}

	returnStmt, ok := funcDecl.Body.Statements[0].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("first statement not *ast.ReturnStmt. got=%T", funcDecl.Body.Statements[0])
	}

	_, ok = returnStmt.Value.(*ast.IfExpr)
	if !ok {
		t.Fatalf("return value not *ast.IfExpr. got=%T", returnStmt.Value)
	}
}

func TestFunctionDeclarationWithLetBinding(t *testing.T) {
	input := `fun calculate(x: int): int {
  let doubled = x * 2
  return doubled
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseFunctionDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	funcDecl, ok := decl.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl not *ast.FunctionDecl. got=%T", decl)
	}

	if len(funcDecl.Body.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(funcDecl.Body.Statements))
	}

	// First statement should be a let expression
	exprStmt1, ok := funcDecl.Body.Statements[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("first statement not *ast.ExprStmt. got=%T", funcDecl.Body.Statements[0])
	}

	_, ok = exprStmt1.Expr.(*ast.LetExpr)
	if !ok {
		t.Fatalf("first expression not *ast.LetExpr. got=%T", exprStmt1.Expr)
	}

	// Second statement should be a return
	_, ok = funcDecl.Body.Statements[1].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("second statement not *ast.ReturnStmt. got=%T", funcDecl.Body.Statements[1])
	}
}

