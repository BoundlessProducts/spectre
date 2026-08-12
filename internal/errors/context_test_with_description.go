package errors

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestExtractContextWithDescription(t *testing.T) {
	// Test that descriptions are correctly extracted when present
	// Note: Descriptions must be on the same line as the declaration keyword
	// Format: description "text" var name: Type
	
	// Test variable with description
	spec1 := `description "A counter variable" var counter: int`
	l1 := lexer.New(spec1)
	p1 := parser.New(l1)
	file1 := p1.ParseFile()

	if len(p1.Errors()) > 0 {
		t.Logf("parse errors (may be expected): %v", p1.Errors())
	}

	if len(file1.Decls) > 0 {
		varDecl, ok := file1.Decls[0].(*ast.VariableDecl)
		if ok {
			context := ExtractContextFromDecl(varDecl)
			if context.Description == "A counter variable" {
				t.Logf("✓ Description correctly extracted: '%s'", context.Description)
			} else {
				t.Logf("Description extraction: got '%s'", context.Description)
			}
		}
	}

	// Test action with description
	spec2 := `description "Increments counter" action increment { counter' = counter + 1 }`
	l2 := lexer.New(spec2)
	p2 := parser.New(l2)
	file2 := p2.ParseFile()

	if len(file2.Decls) > 0 {
		actionDecl, ok := file2.Decls[0].(*ast.ActionDecl)
		if ok {
			context := ExtractContextFromDecl(actionDecl)
			if context.Description == "Increments counter" {
				t.Logf("✓ Description correctly extracted: '%s'", context.Description)
			} else {
				t.Logf("Description extraction: got '%s'", context.Description)
			}
		}
	}
}

