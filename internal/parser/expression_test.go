package parser

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5"

	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpression(LOWEST)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors", len(p.Errors()))
	}

	lit, ok := expr.(*ast.BasicLit)
	if !ok {
		t.Fatalf("expr not *ast.BasicLit. got=%T", expr)
	}

	if lit.Kind != ast.IntLit {
		t.Errorf("lit.Kind not IntLit. got=%v", lit.Kind)
	}

	if lit.Value != "5" {
		t.Errorf("lit.Value not %s. got=%s", "5", lit.Value)
	}
}

func TestFloatLiteralExpression(t *testing.T) {
	input := "3.14"

	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpression(LOWEST)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors", len(p.Errors()))
	}

	lit, ok := expr.(*ast.BasicLit)
	if !ok {
		t.Fatalf("expr not *ast.BasicLit. got=%T", expr)
	}

	if lit.Kind != ast.FloatLit {
		t.Errorf("lit.Kind not FloatLit. got=%v", lit.Kind)
	}

	if lit.Value != "3.14" {
		t.Errorf("lit.Value not %s. got=%s", "3.14", lit.Value)
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world"`

	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpression(LOWEST)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors", len(p.Errors()))
	}

	lit, ok := expr.(*ast.BasicLit)
	if !ok {
		t.Fatalf("expr not *ast.BasicLit. got=%T", expr)
	}

	if lit.Kind != ast.StringLit {
		t.Errorf("lit.Kind not StringLit. got=%v", lit.Kind)
	}

	if lit.Value != "hello world" {
		t.Errorf("lit.Value not %s. got=%s", "hello world", lit.Value)
	}
}

func TestBooleanLiteralExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		expr := p.parseExpression(LOWEST)

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors", len(p.Errors()))
		}

		lit, ok := expr.(*ast.BasicLit)
		if !ok {
			t.Fatalf("expr not *ast.BasicLit. got=%T", expr)
		}

		if lit.Kind != ast.BoolLit {
			t.Errorf("lit.Kind not BoolLit. got=%v", lit.Kind)
		}

		if lit.Value != tt.input {
			t.Errorf("lit.Value not %s. got=%s", tt.input, lit.Value)
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar"

	l := lexer.New(input)
	p := New(l)
	expr := p.parseExpression(LOWEST)

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors", len(p.Errors()))
	}

	ident, ok := expr.(*ast.Ident)
	if !ok {
		t.Fatalf("expr not *ast.Ident. got=%T", expr)
	}

	if ident.Name != "foobar" {
		t.Errorf("ident.Name not %s. got=%s", "foobar", ident.Name)
	}
}

func TestPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator ast.UnaryOp
		value    interface{}
	}{
		{"!5", ast.Not, "5"},
		{"-15", ast.Neg, "15"},
		{"!true", ast.Not, "true"},
		{"!false", ast.Not, "false"},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		expr := p.parseExpression(LOWEST)

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors", len(p.Errors()))
		}

		unary, ok := expr.(*ast.UnaryExpr)
		if !ok {
			t.Fatalf("expr not *ast.UnaryExpr. got=%T", expr)
		}

		if unary.Op != tt.operator {
			t.Fatalf("unary.Op is not %v. got=%v", tt.operator, unary.Op)
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  string
		operator   ast.BinaryOp
		rightValue string
	}{
		{"5 + 5", "5", ast.Add, "5"},
		{"5 - 5", "5", ast.Sub, "5"},
		{"5 * 5", "5", ast.Mul, "5"},
		{"5 / 5", "5", ast.Div, "5"},
		{"5 == 5", "5", ast.Eq, "5"},
		{"5 != 5", "5", ast.Neq, "5"},
		{"5 < 5", "5", ast.Lt, "5"},
		{"5 > 5", "5", ast.Gt, "5"},
		{"5 <= 5", "5", ast.Leq, "5"},
		{"5 >= 5", "5", ast.Geq, "5"},
		{"true && false", "true", ast.And, "false"},
		{"true || false", "true", ast.Or, "false"},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		expr := p.parseExpression(LOWEST)

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
		}

		bin, ok := expr.(*ast.BinaryExpr)
		if !ok {
			t.Fatalf("expr not *ast.BinaryExpr. got=%T", expr)
		}

		if bin.Op != tt.operator {
			t.Fatalf("bin.Op is not %v. got=%v", tt.operator, bin.Op)
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"-a * b"},
		{"!-a"},
		{"a + b + c"},
		{"a + b - c"},
		{"a * b * c"},
		{"a * b / c"},
		{"a + b / c"},
		{"a + b * c + d / e - f"},
		{"5 > 4 == 3 < 4"},
		{"5 < 4 != 3 > 4"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5"},
		{"true"},
		{"false"},
		{"3 > 5 == false"},
		{"3 < 5 == true"},
		{"1 + (2 + 3) + 4"},
		{"(5 + 5) * 2"},
		{"2 / (5 + 5)"},
		{"-(5 + 5)"},
		{"!(true == true)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		expr := p.parseExpression(LOWEST)

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
		}

		if expr == nil {
			t.Errorf("expression is nil for input: %s", tt.input)
		}
	}
}

