package parser

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestInvariantDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "Simple invariant",
			input: `invariant nonNegative {
  counter >= 0
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				invariantDecl, ok := decl.(*ast.InvariantDecl)
				if !ok {
					t.Fatalf("not *ast.InvariantDecl. got=%T", decl)
				}
				if invariantDecl.Name != "nonNegative" {
					t.Errorf("invariant name not 'nonNegative'. got=%s", invariantDecl.Name)
				}
				if invariantDecl.Condition == nil {
					t.Fatal("invariant condition is nil")
				}
			},
		},
		{
			name: "Invariant with description",
			input: `description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				invariantDecl, ok := decl.(*ast.InvariantDecl)
				if !ok {
					t.Fatalf("not *ast.InvariantDecl. got=%T", decl)
				}
				if invariantDecl.Description != "Ensures counter never becomes negative" {
					t.Errorf("description not set correctly. got=%q", invariantDecl.Description)
				}
				if invariantDecl.Name != "nonNegative" {
					t.Errorf("invariant name not 'nonNegative'. got=%s", invariantDecl.Name)
				}
			},
		},
		{
			name: "Invariant with complex expression",
			input: `invariant bounded {
  counter >= 0 && counter <= 100
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				invariantDecl, ok := decl.(*ast.InvariantDecl)
				if !ok {
					t.Fatalf("not *ast.InvariantDecl. got=%T", decl)
				}
				if invariantDecl.Name != "bounded" {
					t.Errorf("invariant name not 'bounded'. got=%s", invariantDecl.Name)
				}
				binaryExpr, ok := invariantDecl.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("condition not *ast.BinaryExpr. got=%T", invariantDecl.Condition)
				}
				if binaryExpr.Op != ast.And {
					t.Errorf("expected && operator, got=%v", binaryExpr.Op)
				}
			},
		},
		{
			name: "Invariant with comparison",
			input: `invariant usersNotEmpty {
  counter > 0
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				invariantDecl, ok := decl.(*ast.InvariantDecl)
				if !ok {
					t.Fatalf("not *ast.InvariantDecl. got=%T", decl)
				}
				if invariantDecl.Name != "usersNotEmpty" {
					t.Errorf("invariant name not 'usersNotEmpty'. got=%s", invariantDecl.Name)
				}
				if invariantDecl.Condition == nil {
					t.Fatal("invariant condition is nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseInvariantDecl()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatal("parseInvariantDecl returned nil")
			}

			tt.validate(t, decl)
		})
	}
}

