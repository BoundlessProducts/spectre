package parser

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestImportDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "Simple import",
			input: `import BoundedCounter`,
			validate: func(t *testing.T, decl ast.Decl) {
				importDecl, ok := decl.(*ast.ImportDecl)
				if !ok {
					t.Fatalf("not *ast.ImportDecl. got=%T", decl)
				}
				if importDecl.Module != "BoundedCounter" {
					t.Errorf("module name not 'BoundedCounter'. got=%s", importDecl.Module)
				}
			},
		},
		{
			name: "Import in module",
			input: `module App {
  import BoundedCounter
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				moduleDecl, ok := decl.(*ast.ModuleDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
				}
				if len(moduleDecl.Decls) != 1 {
					t.Fatalf("expected 1 declaration, got %d", len(moduleDecl.Decls))
				}
				importDecl, ok := moduleDecl.Decls[0].(*ast.ImportDecl)
				if !ok {
					t.Fatalf("declaration not *ast.ImportDecl. got=%T", moduleDecl.Decls[0])
				}
				if importDecl.Module != "BoundedCounter" {
					t.Errorf("module name not 'BoundedCounter'. got=%s", importDecl.Module)
				}
			},
		},
		{
			name: "Multiple imports",
			input: `module App {
  import Counter
  import BoundedCounter
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				moduleDecl, ok := decl.(*ast.ModuleDecl)
				if !ok {
					t.Fatalf("not *ast.ModuleDecl. got=%T", decl)
				}
				if len(moduleDecl.Decls) != 2 {
					t.Fatalf("expected 2 declarations, got %d", len(moduleDecl.Decls))
				}
				import1, ok := moduleDecl.Decls[0].(*ast.ImportDecl)
				if !ok {
					t.Fatalf("first declaration not *ast.ImportDecl. got=%T", moduleDecl.Decls[0])
				}
				if import1.Module != "Counter" {
					t.Errorf("first import not 'Counter'. got=%s", import1.Module)
				}
				import2, ok := moduleDecl.Decls[1].(*ast.ImportDecl)
				if !ok {
					t.Fatalf("second declaration not *ast.ImportDecl. got=%T", moduleDecl.Decls[1])
				}
				if import2.Module != "BoundedCounter" {
					t.Errorf("second import not 'BoundedCounter'. got=%s", import2.Module)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			
			var decl ast.Decl
			if p.curTokenIs(lexer.MODULE) {
				decl = p.parseModuleDecl()
			} else if p.curTokenIs(lexer.IMPORT) {
				decl = p.parseImportDecl()
			} else {
				t.Fatalf("unexpected token: %s", p.curToken.Type)
			}

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatal("parse returned nil")
			}

			tt.validate(t, decl)
		})
	}
}

