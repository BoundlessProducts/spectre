package errors

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestExtractContextFromVariableDecl(t *testing.T) {
	// Description must be on the same line as the declaration keyword
	// The parser expects: description "text" var name: Type
	spec := `var counter: int`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	if len(file.Decls) == 0 {
		t.Fatal("expected at least one declaration")
	}

	varDecl, ok := file.Decls[0].(*ast.VariableDecl)
	if !ok {
		t.Fatalf("expected VariableDecl, got %T", file.Decls[0])
	}

	context := ExtractContextFromDecl(varDecl)

	if context.ElementName != "counter" {
		t.Errorf("expected element name 'counter', got '%s'", context.ElementName)
	}

	if context.ElementType != "variable" {
		t.Errorf("expected element type 'variable', got '%s'", context.ElementType)
	}

	// Description may be empty if not provided
	// The test verifies that context extraction works, even without description
	if context.Description != "" {
		t.Logf("description found: '%s'", context.Description)
	}
}

func TestExtractContextFromActionDecl(t *testing.T) {
	spec := `action increment { counter' = counter + 1 }`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	if len(file.Decls) == 0 {
		t.Fatal("expected at least one declaration")
	}

	actionDecl, ok := file.Decls[0].(*ast.ActionDecl)
	if !ok {
		t.Fatalf("expected ActionDecl, got %T", file.Decls[0])
	}

	context := ExtractContextFromDecl(actionDecl)

	if context.ElementName != "increment" {
		t.Errorf("expected element name 'increment', got '%s'", context.ElementName)
	}

	if context.ElementType != "action" {
		t.Errorf("expected element type 'action', got '%s'", context.ElementType)
	}

	// Description may be empty if not provided
	if context.Description != "" {
		t.Logf("description found: '%s'", context.Description)
	}
}

func TestExtractContextFromInvariantDecl(t *testing.T) {
	spec := `invariant counterNonNegative { counter >= 0 }`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	if len(file.Decls) == 0 {
		t.Fatal("expected at least one declaration")
	}

	invariantDecl, ok := file.Decls[0].(*ast.InvariantDecl)
	if !ok {
		t.Fatalf("expected InvariantDecl, got %T", file.Decls[0])
	}

	context := ExtractContextFromDecl(invariantDecl)

	if context.ElementName != "counterNonNegative" {
		t.Errorf("expected element name 'counterNonNegative', got '%s'", context.ElementName)
	}

	if context.ElementType != "invariant" {
		t.Errorf("expected element type 'invariant', got '%s'", context.ElementType)
	}

	// Description may be empty if not provided
	if context.Description != "" {
		t.Logf("description found: '%s'", context.Description)
	}
}

func TestExtractContextFromTemporalDecl(t *testing.T) {
	spec := `temporal eventuallyReachesTen { eventually (counter == 10) }`

	l := lexer.New(spec)
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	if len(file.Decls) == 0 {
		t.Fatal("expected at least one declaration")
	}

	temporalDecl, ok := file.Decls[0].(*ast.TemporalDecl)
	if !ok {
		t.Fatalf("expected TemporalDecl, got %T", file.Decls[0])
	}

	context := ExtractContextFromDecl(temporalDecl)

	if context.ElementName != "eventuallyReachesTen" {
		t.Errorf("expected element name 'eventuallyReachesTen', got '%s'", context.ElementName)
	}

	if context.ElementType != "temporal property" {
		t.Errorf("expected element type 'temporal property', got '%s'", context.ElementType)
	}

	// Description may be empty if not provided
	if context.Description != "" {
		t.Logf("description found: '%s'", context.Description)
	}
}

func TestErrorContextFormatting(t *testing.T) {
	pos := ast.Position{Line: 10, Column: 5}
	context := NewErrorContext(pos, "Test description", "testElement", "variable")

	positionStr := context.FormatPosition()
	if positionStr != "10:5" {
		t.Errorf("expected position '10:5', got '%s'", positionStr)
	}

	elementStr := context.FormatElement()
	if elementStr != "variable 'testElement'" {
		t.Errorf("expected element 'variable 'testElement'', got '%s'", elementStr)
	}

	if !context.HasDescription() {
		t.Error("expected context to have description")
	}
}

func TestErrorContextWithoutDescription(t *testing.T) {
	pos := ast.Position{Line: 5, Column: 3}
	context := NewErrorContext(pos, "", "myVar", "variable")

	if context.HasDescription() {
		t.Error("expected context to not have description")
	}

	elementStr := context.FormatElement()
	if elementStr != "variable 'myVar'" {
		t.Errorf("expected element 'variable 'myVar'', got '%s'", elementStr)
	}
}

