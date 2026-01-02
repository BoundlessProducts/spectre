package parser

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestIfExpression(t *testing.T) {
	input := `if (x > 0) { x } else { 0 }`

	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpression(LOWEST)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	ifExpr, ok := expr.(*ast.IfExpr)
	if !ok {
		t.Fatalf("expr not *ast.IfExpr. got=%T", expr)
	}

	// Check condition
	binExpr, ok := ifExpr.Condition.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("condition not *ast.BinaryExpr. got=%T", ifExpr.Condition)
	}
	if binExpr.Op != ast.Gt {
		t.Errorf("condition operator not Gt. got=%v", binExpr.Op)
	}

	// Check then expression
	thenIdent, ok := ifExpr.Then.(*ast.Ident)
	if !ok {
		t.Fatalf("then expression not *ast.Ident. got=%T", ifExpr.Then)
	}
	if thenIdent.Name != "x" {
		t.Errorf("then expression name not 'x'. got=%s", thenIdent.Name)
	}

	// Check else expression
	elseLit, ok := ifExpr.Else.(*ast.BasicLit)
	if !ok {
		t.Fatalf("else expression not *ast.BasicLit. got=%T", ifExpr.Else)
	}
	if elseLit.Value != "0" {
		t.Errorf("else expression value not '0'. got=%s", elseLit.Value)
	}
}

func TestIfExpressionWithComplexConditions(t *testing.T) {
	tests := []struct {
		input string
		test  func(*testing.T, *ast.IfExpr)
	}{
		{
			"if (a > b) { a } else { b }",
			func(t *testing.T, expr *ast.IfExpr) {
				binExpr, ok := expr.Condition.(*ast.BinaryExpr)
				if !ok || binExpr.Op != ast.Gt {
					t.Error("condition should be a > b")
				}
			},
		},
		{
			"if (x == 0) { 1 } else { x }",
			func(t *testing.T, expr *ast.IfExpr) {
				binExpr, ok := expr.Condition.(*ast.BinaryExpr)
				if !ok || binExpr.Op != ast.Eq {
					t.Error("condition should be x == 0")
				}
			},
		},
		{
			"if (true) { 1 } else { 0 }",
			func(t *testing.T, expr *ast.IfExpr) {
				lit, ok := expr.Condition.(*ast.BasicLit)
				if !ok || lit.Value != "true" {
					t.Error("condition should be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			expr := p.parseExpression(LOWEST)

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			ifExpr, ok := expr.(*ast.IfExpr)
			if !ok {
				t.Fatalf("expr not *ast.IfExpr. got=%T", expr)
			}

			tt.test(t, ifExpr)
		})
	}
}

func TestLetExpression(t *testing.T) {
	input := `let x = 10`

	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpression(LOWEST)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	letExpr, ok := expr.(*ast.LetExpr)
	if !ok {
		t.Fatalf("expr not *ast.LetExpr. got=%T", expr)
	}

	if letExpr.Name != "x" {
		t.Errorf("letExpr.Name not 'x'. got=%s", letExpr.Name)
	}

	lit, ok := letExpr.Value.(*ast.BasicLit)
	if !ok {
		t.Fatalf("letExpr.Value not *ast.BasicLit. got=%T", letExpr.Value)
	}

	if lit.Value != "10" {
		t.Errorf("letExpr.Value not '10'. got=%s", lit.Value)
	}
}

func TestLetExpressionWithComplexValue(t *testing.T) {
	tests := []struct {
		input string
		name  string
		test  func(*testing.T, ast.Expr)
	}{
		{
			"let sum = 10 + 20",
			"sum",
			func(t *testing.T, value ast.Expr) {
				_, ok := value.(*ast.BinaryExpr)
				if !ok {
					t.Error("value should be a BinaryExpr")
				}
			},
		},
		{
			"let result = if (x > 0) { x } else { 0 }",
			"result",
			func(t *testing.T, value ast.Expr) {
				_, ok := value.(*ast.IfExpr)
				if !ok {
					t.Error("value should be an IfExpr")
				}
			},
		},
		{
			"let name = \"test\"",
			"name",
			func(t *testing.T, value ast.Expr) {
				lit, ok := value.(*ast.BasicLit)
				if !ok || lit.Value != "test" {
					t.Error("value should be string 'test'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			expr := p.parseExpression(LOWEST)

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			letExpr, ok := expr.(*ast.LetExpr)
			if !ok {
				t.Fatalf("expr not *ast.LetExpr. got=%T", expr)
			}

			if letExpr.Name != tt.name {
				t.Errorf("letExpr.Name not '%s'. got=%s", tt.name, letExpr.Name)
			}

			tt.test(t, letExpr.Value)
		})
	}
}

