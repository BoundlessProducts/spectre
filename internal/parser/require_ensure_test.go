package parser

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestRequireStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Stmt)
	}{
		{
			name:  "Simple require",
			input: "require counter > 0",
			validate: func(t *testing.T, stmt ast.Stmt) {
				requireStmt, ok := stmt.(*ast.RequireStmt)
				if !ok {
					t.Fatalf("not *ast.RequireStmt. got=%T", stmt)
				}
				binaryExpr, ok := requireStmt.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("condition not *ast.BinaryExpr. got=%T", requireStmt.Condition)
				}
				if binaryExpr.Op != ast.Gt {
					t.Errorf("expected > operator, got=%v", binaryExpr.Op)
				}
			},
		},
		{
			name:  "Require with complex expression",
			input: "require counter > 0 && counter < 100",
			validate: func(t *testing.T, stmt ast.Stmt) {
				requireStmt, ok := stmt.(*ast.RequireStmt)
				if !ok {
					t.Fatalf("not *ast.RequireStmt. got=%T", stmt)
				}
				binaryExpr, ok := requireStmt.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("condition not *ast.BinaryExpr. got=%T", requireStmt.Condition)
				}
				if binaryExpr.Op != ast.And {
					t.Errorf("expected && operator, got=%v", binaryExpr.Op)
				}
			},
		},
		{
			name:  "Require with primed identifier",
			input: "require counter' > counter",
			validate: func(t *testing.T, stmt ast.Stmt) {
				requireStmt, ok := stmt.(*ast.RequireStmt)
				if !ok {
					t.Fatalf("not *ast.RequireStmt. got=%T", stmt)
				}
				binaryExpr, ok := requireStmt.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("condition not *ast.BinaryExpr. got=%T", requireStmt.Condition)
				}
				leftIdent, ok := binaryExpr.Left.(*ast.Ident)
				if !ok {
					t.Fatalf("left not *ast.Ident. got=%T", binaryExpr.Left)
				}
				if !leftIdent.Prime {
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
				t.Fatalf("parseStatement returned nil")
			}
			tt.validate(t, stmt)
		})
	}
}

func TestEnsureStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Stmt)
	}{
		{
			name:  "Simple ensure",
			input: "ensure counter > 0",
			validate: func(t *testing.T, stmt ast.Stmt) {
				ensureStmt, ok := stmt.(*ast.EnsureStmt)
				if !ok {
					t.Fatalf("not *ast.EnsureStmt. got=%T", stmt)
				}
				binaryExpr, ok := ensureStmt.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("condition not *ast.BinaryExpr. got=%T", ensureStmt.Condition)
				}
				if binaryExpr.Op != ast.Gt {
					t.Errorf("expected > operator, got=%v", binaryExpr.Op)
				}
			},
		},
		{
			name:  "Ensure with primed identifier",
			input: "ensure counter' > counter",
			validate: func(t *testing.T, stmt ast.Stmt) {
				ensureStmt, ok := stmt.(*ast.EnsureStmt)
				if !ok {
					t.Fatalf("not *ast.EnsureStmt. got=%T", stmt)
				}
				binaryExpr, ok := ensureStmt.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("condition not *ast.BinaryExpr. got=%T", ensureStmt.Condition)
				}
				leftIdent, ok := binaryExpr.Left.(*ast.Ident)
				if !ok {
					t.Fatalf("left not *ast.Ident. got=%T", binaryExpr.Left)
				}
				if !leftIdent.Prime {
					t.Error("left identifier should be primed")
				}
			},
		},
		{
			name:  "Ensure with complex expression",
			input: "ensure counter' = counter + 1",
			validate: func(t *testing.T, stmt ast.Stmt) {
				ensureStmt, ok := stmt.(*ast.EnsureStmt)
				if !ok {
					t.Fatalf("not *ast.EnsureStmt. got=%T", stmt)
				}
				binaryExpr, ok := ensureStmt.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("condition not *ast.BinaryExpr. got=%T", ensureStmt.Condition)
				}
				if binaryExpr.Op != ast.Eq {
					t.Errorf("expected = operator, got=%v", binaryExpr.Op)
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
				t.Fatalf("parseStatement returned nil")
			}
			tt.validate(t, stmt)
		})
	}
}

func TestRequireEnsureInAction(t *testing.T) {
	input := `action increment {
  require counter < 100
  counter' = counter + 1
  ensure counter' > counter
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseActionDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
	if decl == nil {
		t.Fatalf("parseActionDecl returned nil")
	}

	actionDecl, ok := decl.(*ast.ActionDecl)
	if !ok {
		t.Fatalf("not *ast.ActionDecl. got=%T", decl)
	}

	if len(actionDecl.Body.Statements) != 3 {
		t.Fatalf("expected 3 statements, got=%d", len(actionDecl.Body.Statements))
	}

	// First statement should be require
	requireStmt, ok := actionDecl.Body.Statements[0].(*ast.RequireStmt)
	if !ok {
		t.Fatalf("first statement not *ast.RequireStmt. got=%T", actionDecl.Body.Statements[0])
	}
	if requireStmt == nil {
		t.Fatal("require statement is nil")
	}

	// Second statement should be assignment
	assignStmt, ok := actionDecl.Body.Statements[1].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("second statement not *ast.AssignStmt. got=%T", actionDecl.Body.Statements[1])
	}
	if assignStmt == nil {
		t.Fatal("assignment statement is nil")
	}

	// Third statement should be ensure
	ensureStmt, ok := actionDecl.Body.Statements[2].(*ast.EnsureStmt)
	if !ok {
		t.Fatalf("third statement not *ast.EnsureStmt. got=%T", actionDecl.Body.Statements[2])
	}
	if ensureStmt == nil {
		t.Fatal("ensure statement is nil")
	}
}

