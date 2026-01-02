package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestParseVariableAndConstantDeclarations(t *testing.T) {
	input := `description "Counter variable"
var counter: int

description "Maximum value"
const MAX: int = 100

var status: str
const ACTIVE: bool = true`

	l := lexer.New(input)
	p := New(l)

	// Parse first variable
	decl1 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	varDecl1, ok := decl1.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl1 not *ast.VariableDecl. got=%T", decl1)
	}
	if varDecl1.Description != "Counter variable" {
		t.Errorf("varDecl1.Description not 'Counter variable'. got=%s", varDecl1.Description)
	}
	if varDecl1.Name != "counter" {
		t.Errorf("varDecl1.Name not 'counter'. got=%s", varDecl1.Name)
	}

	// Parse first constant
	decl2 := p.parseConstantDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	constDecl1, ok := decl2.(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("decl2 not *ast.ConstantDecl. got=%T", decl2)
	}
	if constDecl1.Description != "Maximum value" {
		t.Errorf("constDecl1.Description not 'Maximum value'. got=%s", constDecl1.Description)
	}
	if constDecl1.Name != "MAX" {
		t.Errorf("constDecl1.Name not 'MAX'. got=%s", constDecl1.Name)
	}

	// Parse second variable (no description)
	decl3 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	varDecl2, ok := decl3.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl3 not *ast.VariableDecl. got=%T", decl3)
	}
	if varDecl2.Description != "" {
		t.Errorf("varDecl2.Description should be empty. got=%s", varDecl2.Description)
	}
	if varDecl2.Name != "status" {
		t.Errorf("varDecl2.Name not 'status'. got=%s", varDecl2.Name)
	}

	// Parse second constant
	decl4 := p.parseConstantDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	constDecl2, ok := decl4.(*ast.ConstantDecl)
	if !ok {
		t.Fatalf("decl4 not *ast.ConstantDecl. got=%T", decl4)
	}
	if constDecl2.Name != "ACTIVE" {
		t.Errorf("constDecl2.Name not 'ACTIVE'. got=%s", constDecl2.Name)
	}
}

func TestParseComplexTypes(t *testing.T) {
	input := `var users: Set<str>
var accounts: Map<str, int>
var items: List<int>
var user: { name: str, age: int }
var status: Option<str>`

	l := lexer.New(input)
	p := New(l)

	// Parse Set type
	decl1 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
	varDecl1, ok := decl1.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl1 not *ast.VariableDecl. got=%T", decl1)
	}
	_, ok = varDecl1.Type.(*ast.SetType)
	if !ok {
		t.Errorf("varDecl1.Type not *ast.SetType. got=%T", varDecl1.Type)
	}

	// Parse Map type
	decl2 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
	varDecl2, ok := decl2.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl2 not *ast.VariableDecl. got=%T", decl2)
	}
	_, ok = varDecl2.Type.(*ast.MapType)
	if !ok {
		t.Errorf("varDecl2.Type not *ast.MapType. got=%T", varDecl2.Type)
	}

	// Parse List type
	decl3 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
	varDecl3, ok := decl3.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl3 not *ast.VariableDecl. got=%T", decl3)
	}
	_, ok = varDecl3.Type.(*ast.ListType)
	if !ok {
		t.Errorf("varDecl3.Type not *ast.ListType. got=%T", varDecl3.Type)
	}

	// Parse Record type
	decl4 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
	varDecl4, ok := decl4.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl4 not *ast.VariableDecl. got=%T", decl4)
	}
	_, ok = varDecl4.Type.(*ast.RecordType)
	if !ok {
		t.Errorf("varDecl4.Type not *ast.RecordType. got=%T", varDecl4.Type)
	}

	// Parse Option type
	decl5 := p.parseVariableDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}
	varDecl5, ok := decl5.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("decl5 not *ast.VariableDecl. got=%T", decl5)
	}
	_, ok = varDecl5.Type.(*ast.OptionType)
	if !ok {
		t.Errorf("varDecl5.Type not *ast.OptionType. got=%T", varDecl5.Type)
	}
}

func TestParseFromExampleFiles(t *testing.T) {
	exampleDir := "../../examples"
	
	// Test files that should have variable/constant declarations
	testFiles := []string{
		"counter.spec",
		"constants-example.spec",
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			filePath := filepath.Join(exampleDir, filename)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", filePath, err)
			}

			l := lexer.New(string(content))
			p := New(l)

			// Try to parse at least one variable or constant declaration
			// We'll just verify the parser doesn't crash and can parse descriptions
			desc, found := p.parseDescription()
			if found {
				if desc == "" {
					t.Errorf("Found description but it's empty")
				}
			}

			// Check if we can parse a variable or constant
			if p.curTokenIs(lexer.VAR) {
				decl := p.parseVariableDecl()
				if decl == nil {
					t.Errorf("Failed to parse variable declaration")
				}
				if len(p.Errors()) > 0 {
					t.Logf("Parser errors: %v", p.Errors())
				}
			} else if p.curTokenIs(lexer.CONST) {
				decl := p.parseConstantDecl()
				if decl == nil {
					t.Errorf("Failed to parse constant declaration")
				}
				if len(p.Errors()) > 0 {
					t.Logf("Parser errors: %v", p.Errors())
				}
			}
		})
	}
}

func TestASTStructureValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		parseFn  func() ast.Decl
		validate func(*testing.T, ast.Decl)
	}{
		{
			name:  "Variable with primitive type",
			input: "var counter: int",
			parseFn: func() ast.Decl {
				l := lexer.New("var counter: int")
				p := New(l)
				return p.parseVariableDecl()
			},
			validate: func(t *testing.T, decl ast.Decl) {
				varDecl, ok := decl.(*ast.VariableDecl)
				if !ok {
					t.Fatalf("not *ast.VariableDecl")
				}
				if varDecl.Name != "counter" {
					t.Errorf("name not 'counter'")
				}
				primType, ok := varDecl.Type.(*ast.PrimitiveType)
				if !ok {
					t.Fatalf("type not *ast.PrimitiveType")
				}
				if primType.Name != "int" {
					t.Errorf("type name not 'int'")
				}
			},
		},
		{
			name:  "Constant with expression",
			input: "const SUM: int = 10 + 20",
			parseFn: func() ast.Decl {
				l := lexer.New("const SUM: int = 10 + 20")
				p := New(l)
				return p.parseConstantDecl()
			},
			validate: func(t *testing.T, decl ast.Decl) {
				constDecl, ok := decl.(*ast.ConstantDecl)
				if !ok {
					t.Fatalf("not *ast.ConstantDecl")
				}
				if constDecl.Name != "SUM" {
					t.Errorf("name not 'SUM'")
				}
				binExpr, ok := constDecl.Value.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("value not *ast.BinaryExpr")
				}
				if binExpr.Op != ast.Add {
					t.Errorf("operator not Add")
				}
			},
		},
		{
			name:  "Variable with Set type",
			input: "var users: Set<str>",
			parseFn: func() ast.Decl {
				l := lexer.New("var users: Set<str>")
				p := New(l)
				return p.parseVariableDecl()
			},
			validate: func(t *testing.T, decl ast.Decl) {
				varDecl, ok := decl.(*ast.VariableDecl)
				if !ok {
					t.Fatalf("not *ast.VariableDecl")
				}
				setType, ok := varDecl.Type.(*ast.SetType)
				if !ok {
					t.Fatalf("type not *ast.SetType")
				}
				primType, ok := setType.Element.(*ast.PrimitiveType)
				if !ok {
					t.Fatalf("element type not *ast.PrimitiveType")
				}
				if primType.Name != "str" {
					t.Errorf("element type not 'str'")
				}
			},
		},
		{
			name:  "Variable with Record type",
			input: "var user: { name: str, age: int }",
			parseFn: func() ast.Decl {
				l := lexer.New("var user: { name: str, age: int }")
				p := New(l)
				return p.parseVariableDecl()
			},
			validate: func(t *testing.T, decl ast.Decl) {
				varDecl, ok := decl.(*ast.VariableDecl)
				if !ok {
					t.Fatalf("not *ast.VariableDecl")
				}
				recordType, ok := varDecl.Type.(*ast.RecordType)
				if !ok {
					t.Fatalf("type not *ast.RecordType")
				}
				if len(recordType.Fields) != 2 {
					t.Errorf("expected 2 fields, got %d", len(recordType.Fields))
				}
				if recordType.Fields[0].Name != "name" {
					t.Errorf("first field name not 'name'")
				}
				if recordType.Fields[1].Name != "age" {
					t.Errorf("second field name not 'age'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decl := tt.parseFn()
			if decl == nil {
				t.Fatal("decl is nil")
			}
			tt.validate(t, decl)
		})
	}
}

func TestParserErrorRecovery(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"Missing type", "var counter:", true},
		{"Missing variable name", "var : int", true},
		{"Missing colon", "var counter int", true},
		{"Invalid constant value", "const MAX: int =", true},
		{"Missing equals", "const MAX: int 100", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)

			if p.curTokenIs(lexer.VAR) {
				p.parseVariableDecl()
			} else if p.curTokenIs(lexer.CONST) {
				p.parseConstantDecl()
			}

			hasErrors := len(p.Errors()) > 0
			if tt.expectError && !hasErrors {
				t.Errorf("expected parser errors but got none")
			}
			if !tt.expectError && hasErrors {
				t.Errorf("unexpected parser errors: %v", p.Errors())
			}
		})
	}
}

