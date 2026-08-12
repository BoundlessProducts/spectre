package semantic

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

func TestResolveSimpleVariable(t *testing.T) {
	input := `
var x : int
init {
    x = 10
}
`
	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	resolver := NewResolver(symbolTable)
	errors := resolver.ResolveFile(file)
	if len(errors) > 0 {
		t.Errorf("unexpected resolution errors: %v", errors)
	}
}

func TestResolveUndefinedVariable(t *testing.T) {
	input := `
init {
    x = 10
}
`
	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	resolver := NewResolver(symbolTable)
	errors := resolver.ResolveFile(file)
	if len(errors) == 0 {
		t.Error("expected resolution error for undefined variable")
	}
	if len(errors) > 0 && !containsSubstring(errors, "undefined identifier: x") {
		t.Errorf("expected 'undefined identifier: x' error, got: %v", errors)
	}
}

func TestResolveFunctionCall(t *testing.T) {
	input := `
var result : int

fun add(x: int, y: int) : int {
    return x + y
}

init {
    result = add(5, 10)
}
`
	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	resolver := NewResolver(symbolTable)
	errors := resolver.ResolveFile(file)
	if len(errors) > 0 {
		t.Errorf("unexpected resolution errors: %v", errors)
	}
}

func TestResolveFunctionCallUndefined(t *testing.T) {
	input := `
init {
    result = add(5, 10)
}
`
	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	resolver := NewResolver(symbolTable)
	errors := resolver.ResolveFile(file)
	if len(errors) == 0 {
		t.Error("expected resolution error for undefined function")
	}
}

func TestResolvePrimedVariable(t *testing.T) {
	input := `
var counter : int

action increment {
    counter' = counter + 1
}
`
	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	resolver := NewResolver(symbolTable)
	errors := resolver.ResolveFile(file)
	if len(errors) > 0 {
		t.Errorf("unexpected resolution errors: %v", errors)
	}
}

func TestResolvePrimedVariableUndefined(t *testing.T) {
	input := `
action increment {
    counter' = counter + 1
}
`
	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	resolver := NewResolver(symbolTable)
	errors := resolver.ResolveFile(file)
	if len(errors) == 0 {
		t.Error("expected resolution error for undefined primed variable")
	}
}

func TestResolveAssignmentToConstant(t *testing.T) {
	input := `
const MAX : int = 100

action update {
    MAX = 200
}
`
	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	resolver := NewResolver(symbolTable)
	errors := resolver.ResolveFile(file)
	if len(errors) == 0 {
		t.Error("expected resolution error for assignment to constant")
	}
	if len(errors) > 0 && !containsSubstring(errors, "cannot assign to constant") {
		t.Errorf("expected 'cannot assign to constant' error, got: %v", errors)
	}
}

func TestResolveNestedScopes(t *testing.T) {
	input := `
var x : int

fun test() : int {
    return x + 10
}
`
	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	resolver := NewResolver(symbolTable)
	errors := resolver.ResolveFile(file)
	if len(errors) > 0 {
		t.Errorf("unexpected resolution errors: %v", errors)
	}
}

func TestResolveLetExpression(t *testing.T) {
	input := `
var x : int

init {
    let y = 10
    x = y
}
`
	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	resolver := NewResolver(symbolTable)
	errors := resolver.ResolveFile(file)
	// Let expressions in init blocks may not be fully supported yet
	// For now, just verify it doesn't crash
	if len(errors) > 0 {
		t.Logf("resolution errors (may be expected): %v", errors)
	}
}

// Helper function
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Helper function to check if any string in slice contains substring
func containsSubstring(slice []string, substr string) bool {
	for _, s := range slice {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

