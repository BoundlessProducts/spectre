package parser

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestSuperExpression(t *testing.T) {
	t.Run("Super call with no arguments", func(t *testing.T) {
		input := `super.increment()`
		l := lexer.New(input)
		p := New(l)

		expr := p.parseSuperExpression()
		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
		}
		if expr == nil {
			t.Fatalf("parseSuperExpression returned nil. Errors: %v", p.Errors())
		}

		callExpr, ok := expr.(*ast.CallExpr)
		if !ok {
			t.Fatalf("not *ast.CallExpr. got=%T", expr)
		}
		superExpr, ok := callExpr.Fun.(*ast.SuperExpr)
		if !ok {
			t.Fatalf("function not *ast.SuperExpr. got=%T", callExpr.Fun)
		}
		if superExpr.Method != "increment" {
			t.Errorf("method name not 'increment'. got=%s", superExpr.Method)
		}
		if len(callExpr.Args) != 0 {
			t.Errorf("expected 0 arguments, got %d", len(callExpr.Args))
		}
	})

	t.Run("Super call in action", func(t *testing.T) {
		input := `action increment {
  super.increment()
}`
		l := lexer.New(input)
		p := New(l)

		result := p.parseActionDecl()
		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
		}
		if result == nil {
			t.Fatal("parseActionDecl returned nil")
		}

		actionDecl, ok := result.(*ast.ActionDecl)
		if !ok {
			t.Fatalf("not *ast.ActionDecl. got=%T", result)
		}
		if len(actionDecl.Body.Statements) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(actionDecl.Body.Statements))
		}
		exprStmt, ok := actionDecl.Body.Statements[0].(*ast.ExprStmt)
		if !ok {
			t.Fatalf("statement not *ast.ExprStmt. got=%T", actionDecl.Body.Statements[0])
		}
		callExpr, ok := exprStmt.Expr.(*ast.CallExpr)
		if !ok {
			t.Fatalf("expression not *ast.CallExpr. got=%T", exprStmt.Expr)
		}
		_, ok = callExpr.Fun.(*ast.SuperExpr)
		if !ok {
			t.Fatalf("function not *ast.SuperExpr. got=%T", callExpr.Fun)
		}
	})
}

