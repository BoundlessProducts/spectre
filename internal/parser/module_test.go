package parser

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestModuleDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "Simple module",
			input: `module Counter {
  var counter: int
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				moduleDecl, ok := decl.(*ast.ModuleDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
				}
				if moduleDecl.Name != "Counter" {
					t.Errorf("module name not 'Counter'. got=%s", moduleDecl.Name)
				}
				if len(moduleDecl.Decls) != 1 {
					t.Errorf("expected 1 declaration, got %d", len(moduleDecl.Decls))
				}
			},
		},
		{
			name: "Module with extends",
			input: `module BoundedCounter extends Counter {
  const MAX_VALUE: int = 100
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				moduleDecl, ok := decl.(*ast.ModuleDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
				}
				if moduleDecl.Name != "BoundedCounter" {
					t.Errorf("module name not 'BoundedCounter'. got=%s", moduleDecl.Name)
				}
				if moduleDecl.Extends != "Counter" {
					t.Errorf("extends not 'Counter'. got=%s", moduleDecl.Extends)
				}
			},
		},
		{
			name: "Module with description",
			input: `description "Base counter module"
module Counter {
  var counter: int
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				moduleDecl, ok := decl.(*ast.ModuleDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
				}
				if moduleDecl.Description != "Base counter module" {
					t.Errorf("description not set correctly. got=%q", moduleDecl.Description)
				}
			},
		},
		{
			name: "Module with public action",
			input: `module Counter {
  public action increment {
    counter' = counter + 1
  }
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				moduleDecl, ok := decl.(*ast.ModuleDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
				}
				if len(moduleDecl.Decls) != 1 {
					t.Fatalf("expected 1 declaration, got %d", len(moduleDecl.Decls))
				}
				actionDecl, ok := moduleDecl.Decls[0].(*ast.ActionDecl)
				if !ok {
					t.Fatalf("declaration not *ast.ActionDecl. got=%T", moduleDecl.Decls[0])
				}
				if actionDecl.Visibility != ast.Public {
					t.Errorf("action visibility not Public. got=%v", actionDecl.Visibility)
				}
			},
		},
		{
			name: "Module with private variable",
			input: `module Counter {
  private var internal: int
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				moduleDecl, ok := decl.(*ast.ModuleDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
				}
				varDecl, ok := moduleDecl.Decls[0].(*ast.VariableDecl)
				if !ok {
					t.Fatalf("declaration not *ast.VariableDecl. got=%T", moduleDecl.Decls[0])
				}
				if varDecl.Visibility != ast.Private {
					t.Errorf("variable visibility not Private. got=%v", varDecl.Visibility)
				}
			},
		},
		{
			name: "Module with multiple declarations",
			input: `module Counter {
  var counter: int
  public action increment {
    counter' = counter + 1
  }
  public invariant nonNegative {
    counter >= 0
  }
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				moduleDecl, ok := decl.(*ast.ModuleDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
				}
				if len(moduleDecl.Decls) != 3 {
					t.Errorf("expected 3 declarations, got %d", len(moduleDecl.Decls))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseModuleDecl()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatal("parseModuleDecl returned nil")
			}

			tt.validate(t, decl)
		})
	}
}

