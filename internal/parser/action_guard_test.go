package parser

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestActionWithGuard(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name:  "Action with simple guard",
			input: `action increment when counter < 100 {
  counter' = counter + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if actionDecl.Name != "increment" {
					t.Errorf("action name not 'increment'. got=%s", actionDecl.Name)
				}
				if actionDecl.Guard == nil {
					t.Fatal("guard should not be nil")
				}
				binaryExpr, ok := actionDecl.Guard.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("guard not *ast.BinaryExpr. got=%T", actionDecl.Guard)
				}
				if binaryExpr.Op != ast.Lt {
					t.Errorf("expected < operator in guard, got=%v", binaryExpr.Op)
				}
			},
		},
		{
			name:  "Action with guard and parameters",
			input: `action addUser(user: int) when user < 100 {
  user' = user + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if actionDecl.Name != "addUser" {
					t.Errorf("action name not 'addUser'. got=%s", actionDecl.Name)
				}
				if len(actionDecl.Parameters) != 1 {
					t.Fatalf("expected 1 parameter, got=%d", len(actionDecl.Parameters))
				}
				if actionDecl.Guard == nil {
					t.Fatal("guard should not be nil")
				}
			},
		},
		{
			name:  "Action with complex guard expression",
			input: `action increment when counter < 100 && counter >= 0 {
  counter' = counter + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if actionDecl.Guard == nil {
					t.Fatal("guard should not be nil")
				}
				binaryExpr, ok := actionDecl.Guard.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("guard not *ast.BinaryExpr. got=%T", actionDecl.Guard)
				}
				if binaryExpr.Op != ast.And {
					t.Errorf("expected && operator in guard, got=%v", binaryExpr.Op)
				}
			},
		},
		{
			name:  "Action without guard",
			input: `action increment {
  counter' = counter + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if actionDecl.Guard != nil {
					t.Error("guard should be nil for action without guard")
				}
			},
		},
		{
			name:  "Action with guard and description",
			input: `description "Increments counter when below limit"
action increment when counter < 100 {
  counter' = counter + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if actionDecl.Description != "Increments counter when below limit" {
					t.Errorf("description not correct. got=%s", actionDecl.Description)
				}
				if actionDecl.Guard == nil {
					t.Fatal("guard should not be nil")
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
				t.Fatalf("parseActionDecl returned nil")
			}
			tt.validate(t, decl)
		})
	}
}

func TestActionGuardWithPrimedIdentifier(t *testing.T) {
	input := `action updateCounter when counter' > counter {
  counter' = counter + 1
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
	if actionDecl.Guard == nil {
		t.Fatal("guard should not be nil")
	}

	binaryExpr, ok := actionDecl.Guard.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("guard not *ast.BinaryExpr. got=%T", actionDecl.Guard)
	}

	leftIdent, ok := binaryExpr.Left.(*ast.Ident)
	if !ok {
		t.Fatalf("left not *ast.Ident. got=%T", binaryExpr.Left)
	}
	if !leftIdent.Prime {
		t.Error("left identifier in guard should be primed")
	}
}

