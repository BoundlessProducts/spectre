package parser

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestErrorRecovery(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrors int
		expectedDecls  int
		description    string
	}{
		{
			name: "Recover from missing colon",
			input: `var a: int
var b int
var c: int`,
			expectedErrors: 1, // Missing colon
			expectedDecls:  2, // Should parse a and c
			description:    "Parser should recover from syntax errors",
		},
		{
			name: "Recover from missing identifier",
			input: `var a: int
var : int
var c: int`,
			expectedErrors: 1, // Missing identifier
			expectedDecls:  2, // Should parse a and c
			description:    "Parser should recover from missing identifier",
		},
		{
			name: "Recover from invalid action syntax",
			input: `action valid {
  x' = x + 1
}
action invalid {
  x' = x + 1
}
var y: int`,
			expectedErrors: 0, // All syntax is valid
			expectedDecls:  3, // Should parse both actions and y variable
			description:    "Parser should parse all valid declarations",
		},
		{
			name: "Recover from multiple syntax errors",
			input: `var a: int
var b int
var c: int
var d int
var e: int`,
			expectedErrors: 2, // Two missing colons
			expectedDecls:  3, // Should parse a, c, and e
			description:    "Parser should recover from multiple errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			file := p.ParseFile()

			if file == nil {
				t.Fatal("ParseFile returned nil")
			}

			errorCount := len(p.Errors())
			if errorCount != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d. Errors: %v", tt.expectedErrors, errorCount, p.Errors())
			}

			declCount := len(file.Decls)
			if declCount != tt.expectedDecls {
				t.Errorf("expected %d declarations after recovery, got %d", tt.expectedDecls, declCount)
			}

			t.Logf("%s: Found %d errors, parsed %d declarations", tt.description, errorCount, declCount)
		})
	}
}

func TestSyncToDeclaration(t *testing.T) {
	input := `var x: int
some invalid tokens here
var y: int`

	l := lexer.New(input)
	p := New(l)

	// Parse first declaration
	decl1 := p.parseVariableDecl()
	if decl1 == nil {
		t.Fatal("failed to parse first variable")
	}

	// Manually sync to next declaration
	p.syncToDeclaration()

	// Should now be on VAR token for y
	if !p.curTokenIs(lexer.VAR) {
		t.Errorf("expected VAR token after sync, got %s", p.curToken.Type)
	}

	// Parse second declaration
	decl2 := p.parseVariableDecl()
	if decl2 == nil {
		t.Fatal("failed to parse second variable")
	}

	varDecl, ok := decl2.(*ast.VariableDecl)
	if !ok {
		t.Fatalf("not *ast.VariableDecl. got=%T", decl2)
	}
	if varDecl.Name != "y" {
		t.Errorf("expected variable name 'y', got %s", varDecl.Name)
	}
}

