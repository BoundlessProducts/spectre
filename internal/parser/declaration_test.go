package parser

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestVariableDeclaration(t *testing.T) {
	input := `var counter: int`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseVariableDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	varDecl, ok := decl.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl not *ast.VariableDecl. got=%T", decl)
	}

	if varDecl.Name != "counter" {
		t.Errorf("varDecl.Name not 'counter'. got=%s", varDecl.Name)
	}

	primType, ok := varDecl.Type.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("varDecl.Type not *ast.PrimitiveType. got=%T", varDecl.Type)
	}

	if primType.Name != "int" {
		t.Errorf("varDecl.Type.Name not 'int'. got=%s", primType.Name)
	}
}

func TestVariableDeclarationWithDescription(t *testing.T) {
	input := `description "Tracks a numeric counter value"
var counter: int`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseVariableDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	varDecl, ok := decl.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl not *ast.VariableDecl. got=%T", decl)
	}

	if varDecl.Description != "Tracks a numeric counter value" {
		t.Errorf("varDecl.Description not 'Tracks a numeric counter value'. got=%s", varDecl.Description)
	}

	if varDecl.Name != "counter" {
		t.Errorf("varDecl.Name not 'counter'. got=%s", varDecl.Name)
	}
}

func TestVariableDeclarationWithComplexType(t *testing.T) {
	tests := []struct {
		input    string
		typeTest func(ast.Type) bool
	}{
		{"var users: Set<str>", func(t ast.Type) bool {
			_, ok := t.(*ast.SetType)
			return ok
		}},
		{"var accounts: Map<str, int>", func(t ast.Type) bool {
			_, ok := t.(*ast.MapType)
			return ok
		}},
		{"var items: List<int>", func(t ast.Type) bool {
			_, ok := t.(*ast.ListType)
			return ok
		}},
		{"var user: { name: str, age: int }", func(t ast.Type) bool {
			_, ok := t.(*ast.RecordType)
			return ok
		}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		decl := p.parseVariableDecl()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors for %s: %v", len(p.Errors()), tt.input, p.Errors())
		}

		varDecl, ok := decl.(*ast.VariableDecl)
		if !ok {
			t.Fatalf("decl not *ast.VariableDecl for %s. got=%T", tt.input, decl)
		}

		if !tt.typeTest(varDecl.Type) {
			t.Errorf("type test failed for: %s", tt.input)
		}
	}
}

func TestMultipleVariableDeclarations(t *testing.T) {
	input := `var counter: int
var status: str
var active: bool`

	l := lexer.New(input)
	p := New(l)

	decl1 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	varDecl1, ok := decl1.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl1 not *ast.VariableDecl. got=%T", decl1)
	}
	if varDecl1.Name != "counter" {
		t.Errorf("decl1.Name not 'counter'. got=%s", varDecl1.Name)
	}

	decl2 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors after second decl: %v", len(p.Errors()), p.Errors())
	}

	varDecl2, ok := decl2.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl2 not *ast.VariableDecl. got=%T", decl2)
	}
	if varDecl2.Name != "status" {
		t.Errorf("decl2.Name not 'status'. got=%s", varDecl2.Name)
	}

	decl3 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors after third decl: %v", len(p.Errors()), p.Errors())
	}

	varDecl3, ok := decl3.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl3 not *ast.VariableDecl. got=%T", decl3)
	}
	if varDecl3.Name != "active" {
		t.Errorf("decl3.Name not 'active'. got=%s", varDecl3.Name)
	}
}

