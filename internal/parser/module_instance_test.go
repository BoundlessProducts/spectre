package parser

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestModuleInstanceDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "Simple module instance",
			input: `module MyCounter = Counter with {
  counter = myCounter
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				instanceDecl, ok := decl.(*ast.ModuleInstanceDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleInstanceDecl. got=%T", decl)
				}
				if instanceDecl.Name != "MyCounter" {
					t.Errorf("instance name not 'MyCounter'. got=%s", instanceDecl.Name)
				}
				if instanceDecl.Module != "Counter" {
					t.Errorf("module name not 'Counter'. got=%s", instanceDecl.Module)
				}
				if len(instanceDecl.Substitutions) != 1 {
					t.Errorf("expected 1 substitution, got %d", len(instanceDecl.Substitutions))
				}
				if instanceDecl.Substitutions["counter"] != "myCounter" {
					t.Errorf("substitution not correct. expected 'myCounter', got %q", instanceDecl.Substitutions["counter"])
				}
			},
		},
		{
			name: "Module instance with multiple substitutions",
			input: `module MyCounter = Counter with {
  counter = myCounter,
  maxValue = max
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				instanceDecl, ok := decl.(*ast.ModuleInstanceDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleInstanceDecl. got=%T", decl)
				}
				if len(instanceDecl.Substitutions) != 2 {
					t.Errorf("expected 2 substitutions, got %d", len(instanceDecl.Substitutions))
				}
				if instanceDecl.Substitutions["counter"] != "myCounter" {
					t.Errorf("substitution 'counter' not correct. expected 'myCounter', got %q", instanceDecl.Substitutions["counter"])
				}
				if instanceDecl.Substitutions["maxValue"] != "max" {
					t.Errorf("substitution 'maxValue' not correct. expected 'max', got %q", instanceDecl.Substitutions["maxValue"])
				}
			},
		},
		{
			name: "Module instance with description",
			input: `description "Create an instance with different variable name"
module MyCounter = Counter with {
  counter = myCounter
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				instanceDecl, ok := decl.(*ast.ModuleInstanceDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleInstanceDecl. got=%T", decl)
				}
				// Note: descriptions are not stored in ModuleInstanceDecl currently
				// This test just verifies parsing doesn't fail
				if instanceDecl.Name != "MyCounter" {
					t.Errorf("instance name not 'MyCounter'. got=%s", instanceDecl.Name)
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

