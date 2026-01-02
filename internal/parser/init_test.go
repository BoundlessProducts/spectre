package parser

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestInitDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "Simple init block",
			input: `init {
  counter = 0
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				initDecl, ok := decl.(*ast.InitDecl)
				if !ok {
					t.Fatalf("not *ast.InitDecl. got=%T", decl)
				}
				if initDecl.Body == nil {
					t.Fatal("initDecl.Body is nil")
				}
				if len(initDecl.Body.Statements) != 1 {
					t.Fatalf("expected 1 statement, got %d", len(initDecl.Body.Statements))
				}
				assignStmt, ok := initDecl.Body.Statements[0].(*ast.AssignStmt)
				if !ok {
					t.Fatalf("first statement not *ast.AssignStmt. got=%T", initDecl.Body.Statements[0])
				}
				ident, ok := assignStmt.Left.(*ast.Ident)
				if !ok {
					t.Fatalf("assignStmt.Left not *ast.Ident. got=%T", assignStmt.Left)
				}
				if ident.Name != "counter" {
					t.Errorf("identifier name not 'counter'. got=%s", ident.Name)
				}
			},
		},
		{
			name: "Init block with multiple assignments",
			input: `init {
  counter = 0
  users = Set.empty()
  status = Status.Pending
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				initDecl, ok := decl.(*ast.InitDecl)
				if !ok {
					t.Fatalf("not *ast.InitDecl. got=%T", decl)
				}
				if len(initDecl.Body.Statements) != 3 {
					t.Fatalf("expected 3 statements, got %d", len(initDecl.Body.Statements))
				}
			},
		},
		{
			name: "Init with single expression",
			input: `init counter = 0 && users = Set.empty()`,
			validate: func(t *testing.T, decl ast.Decl) {
				initDecl, ok := decl.(*ast.InitDecl)
				if !ok {
					t.Fatalf("not *ast.InitDecl. got=%T", decl)
				}
				if initDecl.Expression == nil {
					t.Fatal("initDecl.Expression is nil")
				}
				if initDecl.Body != nil {
					t.Error("initDecl.Body should be nil for expression form")
				}
			},
		},
		{
			name: "Init with description",
			input: `description "System starts with counter initialized to zero"
init {
  counter = 0
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				initDecl, ok := decl.(*ast.InitDecl)
				if !ok {
					t.Fatalf("not *ast.InitDecl. got=%T", decl)
				}
				if initDecl.Description != "System starts with counter initialized to zero" {
					t.Errorf("description not set correctly. got=%q", initDecl.Description)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseInitDecl()

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

func TestOneOfInitDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "oneOf with single variable",
			input: `init oneOf {
  counter = 0,
  counter = 5,
  counter = 10
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				oneOfDecl, ok := decl.(*ast.OneOfInitDecl)
				if !ok {
					t.Fatalf("not *ast.OneOfInitDecl. got=%T", decl)
				}
				if len(oneOfDecl.Options) != 3 {
					t.Fatalf("expected 3 options, got %d", len(oneOfDecl.Options))
				}
			},
		},
		{
			name: "oneOf with multiple variables (block syntax)",
			input: `init oneOf {
  {
    counter = 0
    mode = "start"
    initialized = false
  },
  {
    counter = 10
    mode = "resume"
    initialized = true
  }
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				oneOfDecl, ok := decl.(*ast.OneOfInitDecl)
				if !ok {
					t.Fatalf("not *ast.OneOfInitDecl. got=%T", decl)
				}
				if len(oneOfDecl.Options) != 2 {
					t.Fatalf("expected 2 options, got %d", len(oneOfDecl.Options))
				}
				if len(oneOfDecl.Options[0].Statements) != 3 {
					t.Fatalf("expected 3 statements in first option, got %d", len(oneOfDecl.Options[0].Statements))
				}
			},
		},
		{
			name: "oneOf with description",
			input: `description "System can start from multiple configurations"
init oneOf {
  { counter = 0 },
  { counter = 10 }
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				oneOfDecl, ok := decl.(*ast.OneOfInitDecl)
				if !ok {
					t.Fatalf("not *ast.OneOfInitDecl. got=%T", decl)
				}
				if oneOfDecl.Description != "System can start from multiple configurations" {
					t.Errorf("description not set correctly. got=%q", oneOfDecl.Description)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseInitDecl()

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

func TestAssignStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Stmt)
	}{
		{
			name: "Simple assignment",
			input: "counter = 0",
			validate: func(t *testing.T, stmt ast.Stmt) {
				assignStmt, ok := stmt.(*ast.AssignStmt)
				if !ok {
					t.Fatalf("not *ast.AssignStmt. got=%T", stmt)
				}
				ident, ok := assignStmt.Left.(*ast.Ident)
				if !ok {
					t.Fatalf("assignStmt.Left not *ast.Ident. got=%T", assignStmt.Left)
				}
				if ident.Name != "counter" {
					t.Errorf("identifier name not 'counter'. got=%s", ident.Name)
				}
				lit, ok := assignStmt.Right.(*ast.BasicLit)
				if !ok {
					t.Fatalf("assignStmt.Right not *ast.BasicLit. got=%T", assignStmt.Right)
				}
				if lit.Value != "0" {
					t.Errorf("literal value not '0'. got=%s", lit.Value)
				}
			},
		},
		{
			name: "Assignment with expression",
			input: "counter = counter + 1",
			validate: func(t *testing.T, stmt ast.Stmt) {
				assignStmt, ok := stmt.(*ast.AssignStmt)
				if !ok {
					t.Fatalf("not *ast.AssignStmt. got=%T", stmt)
				}
				_, ok = assignStmt.Right.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("assignStmt.Right not *ast.BinaryExpr. got=%T", assignStmt.Right)
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

			tt.validate(t, stmt)
		})
	}
}

