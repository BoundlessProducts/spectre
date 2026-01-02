package parser

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestConstantDeclaration(t *testing.T) {
	input := `const MAX_VALUE: int = 100`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseConstantDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	constDecl, ok := decl.(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("decl not *ast.ConstantDecl. got=%T", decl)
	}

	if constDecl.Name != "MAX_VALUE" {
		t.Errorf("constDecl.Name not 'MAX_VALUE'. got=%s", constDecl.Name)
	}

	primType, ok := constDecl.Type.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("constDecl.Type not *ast.PrimitiveType. got=%T", constDecl.Type)
	}

	if primType.Name != "int" {
		t.Errorf("constDecl.Type.Name not 'int'. got=%s", primType.Name)
	}

	lit, ok := constDecl.Value.(*ast.BasicLit)
	if !ok {
		t.Fatalf("constDecl.Value not *ast.BasicLit. got=%T", constDecl.Value)
	}

	if lit.Value != "100" {
		t.Errorf("constDecl.Value not '100'. got=%s", lit.Value)
	}
}

func TestConstantDeclarationWithDescription(t *testing.T) {
	input := `description "Maximum allowed value"
const MAX_VALUE: int = 100`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseConstantDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	constDecl, ok := decl.(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("decl not *ast.ConstantDecl. got=%T", decl)
	}

	if constDecl.Description != "Maximum allowed value" {
		t.Errorf("constDecl.Description not 'Maximum allowed value'. got=%s", constDecl.Description)
	}

	if constDecl.Name != "MAX_VALUE" {
		t.Errorf("constDecl.Name not 'MAX_VALUE'. got=%s", constDecl.Name)
	}
}

func TestConstantDeclarationWithDifferentTypes(t *testing.T) {
	tests := []struct {
		input    string
		typeName string
		value    string
		valueKind ast.LitKind
	}{
		{"const MAX: int = 100", "int", "100", ast.IntLit},
		{"const PI: float = 3.14", "float", "3.14", ast.FloatLit},
		{"const NAME: str = \"test\"", "str", "test", ast.StringLit},
		{"const ACTIVE: bool = true", "bool", "true", ast.BoolLit},
		{"const INACTIVE: bool = false", "bool", "false", ast.BoolLit},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		decl := p.parseConstantDecl()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors for %s: %v", len(p.Errors()), tt.input, p.Errors())
		}

		constDecl, ok := decl.(*ast.ConstantDecl)
		if !ok {
			t.Fatalf("decl not *ast.ConstantDecl for %s. got=%T", tt.input, decl)
		}

		primType, ok := constDecl.Type.(*ast.PrimitiveType)
		if !ok {
			t.Fatalf("constDecl.Type not *ast.PrimitiveType for %s. got=%T", tt.input, constDecl.Type)
		}

		if primType.Name != tt.typeName {
			t.Errorf("type name not %s for %s. got=%s", tt.typeName, tt.input, primType.Name)
		}

		lit, ok := constDecl.Value.(*ast.BasicLit)
		if !ok {
			t.Fatalf("constDecl.Value not *ast.BasicLit for %s. got=%T", tt.input, constDecl.Value)
		}

		if lit.Value != tt.value {
			t.Errorf("value not %s for %s. got=%s", tt.value, tt.input, lit.Value)
		}

		if lit.Kind != tt.valueKind {
			t.Errorf("value kind not %v for %s. got=%v", tt.valueKind, tt.input, lit.Kind)
		}
	}
}

func TestConstantDeclarationWithExpression(t *testing.T) {
	tests := []struct {
		input string
		test  func(ast.Expr) bool
	}{
		{"const SUM: int = 10 + 20", func(e ast.Expr) bool {
			_, ok := e.(*ast.BinaryExpr)
			return ok
		}},
		{"const NEG: int = -5", func(e ast.Expr) bool {
			_, ok := e.(*ast.UnaryExpr)
			return ok
		}},
		{"const PAREN: int = (10)", func(e ast.Expr) bool {
			_, ok := e.(*ast.ParenExpr)
			return ok
		}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		decl := p.parseConstantDecl()

		if len(p.Errors()) != 0 {
			t.Fatalf("parser has %d errors for %s: %v", len(p.Errors()), tt.input, p.Errors())
		}

		constDecl, ok := decl.(*ast.ConstantDecl)
		if !ok {
			t.Fatalf("decl not *ast.ConstantDecl for %s. got=%T", tt.input, decl)
		}

		if !tt.test(constDecl.Value) {
			t.Errorf("expression test failed for: %s", tt.input)
		}
	}
}

func TestMultipleConstantDeclarations(t *testing.T) {
	input := `const MAX: int = 100
const MIN: int = 0
const DEFAULT: str = "test"`

	l := lexer.New(input)
	p := New(l)

	decl1 := p.parseConstantDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	constDecl1, ok := decl1.(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("decl1 not *ast.ConstantDecl. got=%T", decl1)
	}
	if constDecl1.Name != "MAX" {
		t.Errorf("decl1.Name not 'MAX'. got=%s", constDecl1.Name)
	}

	decl2 := p.parseConstantDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors after second decl: %v", len(p.Errors()), p.Errors())
	}

	constDecl2, ok := decl2.(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("decl2 not *ast.ConstantDecl. got=%T", decl2)
	}
	if constDecl2.Name != "MIN" {
		t.Errorf("decl2.Name not 'MIN'. got=%s", constDecl2.Name)
	}

	decl3 := p.parseConstantDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors after third decl: %v", len(p.Errors()), p.Errors())
	}

	constDecl3, ok := decl3.(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("decl3 not *ast.ConstantDecl. got=%T", decl3)
	}
	if constDecl3.Name != "DEFAULT" {
		t.Errorf("decl3.Name not 'DEFAULT'. got=%s", constDecl3.Name)
	}
}

