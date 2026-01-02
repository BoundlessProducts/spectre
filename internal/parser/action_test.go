package parser

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestActionDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "Simple action without parameters",
			input: `action increment {
  counter = counter + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if actionDecl.Name != "increment" {
					t.Errorf("action name not 'increment'. got=%s", actionDecl.Name)
				}
				if len(actionDecl.Parameters) != 0 {
					t.Errorf("expected 0 parameters, got %d", len(actionDecl.Parameters))
				}
				if actionDecl.Body == nil {
					t.Fatal("actionDecl.Body is nil")
				}
			},
		},
		{
			name: "Action with parameters",
			input: `action addUser(user: User) {
  users = users
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
					t.Fatalf("expected 1 parameter, got %d", len(actionDecl.Parameters))
				}
				if actionDecl.Parameters[0].Name != "user" {
					t.Errorf("parameter name not 'user'. got=%s", actionDecl.Parameters[0].Name)
				}
			},
		},
		{
			name: "Action with multiple parameters",
			input: `action transfer(fromId: int, toId: int, amount: int) {
  accounts = accounts
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if len(actionDecl.Parameters) != 3 {
					t.Fatalf("expected 3 parameters, got %d", len(actionDecl.Parameters))
				}
			},
		},
		{
			name: "Action with description",
			input: `description "Increments the counter by one"
action increment {
  counter = counter + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if actionDecl.Description != "Increments the counter by one" {
					t.Errorf("description not set correctly. got=%q", actionDecl.Description)
				}
			},
		},
		{
			name: "Action with multiple statements",
			input: `action addUser(user: User) {
  users = users
  counter = counter + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if len(actionDecl.Body.Statements) < 2 {
					t.Errorf("expected at least 2 statements, got %d", len(actionDecl.Body.Statements))
				}
			},
		},
		{
			name: "Action with empty body",
			input: `action noop {
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				actionDecl, ok := decl.(*ast.ActionDecl)
				if !ok {
					t.Fatalf("not *ast.ActionDecl. got=%T", decl)
				}
				if actionDecl.Body == nil {
					t.Fatal("actionDecl.Body is nil")
				}
				if len(actionDecl.Body.Statements) != 0 {
					t.Errorf("expected 0 statements, got %d", len(actionDecl.Body.Statements))
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

			tt.validate(t, decl)
		})
	}
}

