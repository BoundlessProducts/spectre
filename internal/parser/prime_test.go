package parser

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestPrimeNotationParsing(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Expr)
	}{
		{
			name: "Primed identifier",
			input: "counter'",
			validate: func(t *testing.T, expr ast.Expr) {
				ident, ok := expr.(*ast.Ident)
				if !ok {
					t.Fatalf("not *ast.Ident. got=%T", expr)
				}
				if ident.Name != "counter" {
					t.Errorf("identifier name not 'counter'. got=%s", ident.Name)
				}
				if !ident.Prime {
					t.Error("identifier should be primed")
				}
			},
		},
		{
			name: "Non-primed identifier",
			input: "counter",
			validate: func(t *testing.T, expr ast.Expr) {
				ident, ok := expr.(*ast.Ident)
				if !ok {
					t.Fatalf("not *ast.Ident. got=%T", expr)
				}
				if ident.Prime {
					t.Error("identifier should not be primed")
				}
			},
		},
		{
			name: "Primed identifier in expression",
			input: "counter' = counter + 1",
			validate: func(t *testing.T, expr ast.Expr) {
				// This should parse as an assignment statement
				// But we're parsing as expression, so it might fail
				// Let's test the assignment parser instead
			},
		},
		{
			name: "Primed identifier comparison",
			input: "counter' > counter",
			validate: func(t *testing.T, expr ast.Expr) {
				binaryExpr, ok := expr.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("not *ast.BinaryExpr. got=%T", expr)
				}
				leftIdent, ok := binaryExpr.Left.(*ast.Ident)
				if !ok {
					t.Fatalf("left not *ast.Ident. got=%T", binaryExpr.Left)
				}
				if !leftIdent.Prime {
					t.Error("left identifier should be primed")
				}
				rightIdent, ok := binaryExpr.Right.(*ast.Ident)
				if !ok {
					t.Fatalf("right not *ast.Ident. got=%T", binaryExpr.Right)
				}
				if rightIdent.Prime {
					t.Error("right identifier should not be primed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			expr := p.parseExpression(LOWEST)

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			if expr == nil {
				t.Fatal("expr is nil")
			}

			if tt.validate != nil {
				tt.validate(t, expr)
			}
		})
	}
}

func TestPrimeNotationInAssignments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Stmt)
	}{
		{
			name: "Assignment with primed identifier",
			input: "counter' = counter + 1",
			validate: func(t *testing.T, stmt ast.Stmt) {
				assignStmt, ok := stmt.(*ast.AssignStmt)
				if !ok {
					t.Fatalf("not *ast.AssignStmt. got=%T", stmt)
				}
				ident, ok := assignStmt.Left.(*ast.Ident)
				if !ok {
					t.Fatalf("left not *ast.Ident. got=%T", assignStmt.Left)
				}
				if !ident.Prime {
					t.Error("left identifier should be primed")
				}
				if ident.Name != "counter" {
					t.Errorf("identifier name not 'counter'. got=%s", ident.Name)
				}
			},
		},
		{
			name: "Multiple primed assignments",
			input: `counter' = counter + 1
mode' = mode`,
			validate: func(t *testing.T, stmt ast.Stmt) {
				// This will parse the first statement
				assignStmt, ok := stmt.(*ast.AssignStmt)
				if !ok {
					t.Fatalf("not *ast.AssignStmt. got=%T", stmt)
				}
				ident, ok := assignStmt.Left.(*ast.Ident)
				if !ok {
					t.Fatalf("left not *ast.Ident. got=%T", assignStmt.Left)
				}
				if !ident.Prime {
					t.Error("left identifier should be primed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			stmt := p.parseStatement()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			if stmt == nil {
				t.Fatal("stmt is nil")
			}

			if tt.validate != nil {
				tt.validate(t, stmt)
			}
		})
	}
}

func TestPrimeNotationInActions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "Action with primed assignment",
			input: `action increment {
  counter' = counter + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if len(actionDecl.Body.Statements) != 1 {
					t.Fatalf("expected 1 statement, got %d", len(actionDecl.Body.Statements))
				}
				assignStmt, ok := actionDecl.Body.Statements[0].(*ast.AssignStmt)
				if !ok {
					t.Fatalf("first statement not *ast.AssignStmt. got=%T", actionDecl.Body.Statements[0])
				}
				ident, ok := assignStmt.Left.(*ast.Ident)
				if !ok {
					t.Fatalf("left not *ast.Ident. got=%T", assignStmt.Left)
				}
				if !ident.Prime {
					t.Error("left identifier should be primed")
				}
			},
		},
		{
			name: "Action with multiple primed assignments",
			input: `action update {
  counter' = counter + 1
  mode' = "active"
  initialized' = true
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if len(actionDecl.Body.Statements) != 3 {
					t.Fatalf("expected 3 statements, got %d", len(actionDecl.Body.Statements))
				}
				// Check first assignment
				assignStmt1, ok := actionDecl.Body.Statements[0].(*ast.AssignStmt)
				if !ok {
					t.Fatalf("first statement not *ast.AssignStmt")
				}
				ident1, ok := assignStmt1.Left.(*ast.Ident)
				if !ok || !ident1.Prime {
					t.Error("first assignment should have primed identifier")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseActionDecl()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			if decl == nil {
				t.Fatal("decl is nil")
			}

			if tt.validate != nil {
				tt.validate(t, decl)
			}
		})
	}
}

