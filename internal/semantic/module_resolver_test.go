package semantic

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
)

func TestModuleResolution(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  bool
		errorContains string
	}{
		{
			name: "Simple module declaration",
			input: `
module Counter {
    var count: int
}
`,
			expectErrors: false,
		},
		{
			name: "Module with extension",
			input: `
module Base {
    var x: int
}

module Extended extends Base {
    var y: int
}
`,
			expectErrors: false,
		},
		{
			name: "Module extends undefined module",
			input: `
module Extended extends Undefined {
    var y: int
}
`,
			expectErrors: true,
			errorContains: "extends undefined module",
		},
		{
			name: "Duplicate module declaration",
			input: `
module Counter {
    var count: int
}

module Counter {
    var other: int
}
`,
			expectErrors: true,
			errorContains: "duplicate module declaration",
		},
		{
			name: "Circular dependency",
			input: `
module A extends B {
    var x: int
}

module B extends A {
    var y: int
}
`,
			expectErrors: true,
			errorContains: "circular module dependency",
		},
		{
			name: "Import existing module",
			input: `
module Counter {
    var count: int
}

import Counter
`,
			expectErrors: false,
		},
		{
			name: "Import undefined module",
			input: `
import UndefinedModule
`,
			expectErrors: true,
			errorContains: "imported module not found",
		},
		{
			name: "Duplicate import",
			input: `
module Counter {
    var count: int
}

import Counter
import Counter
`,
			expectErrors: true,
			errorContains: "duplicate import",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.input))
			file := p.ParseFile()
			if len(p.Errors()) > 0 {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			builder := NewBuilder()
			symbolTable, buildErrors := builder.BuildSymbolTable(file)
			// For duplicate module tests, build errors are expected (builder catches them)
			if len(buildErrors) > 0 {
				if tt.expectErrors && tt.errorContains == "duplicate module declaration" {
					// Builder correctly caught duplicate modules - this is expected
					// Check that the error mentions the duplicate
					found := false
					for _, err := range buildErrors {
						if containsSubstring([]string{err}, "already defined") {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected duplicate module error, got: %v", buildErrors)
					}
					// Don't continue validation if build failed
					return
				} else if !tt.expectErrors {
					t.Fatalf("unexpected build errors: %v", buildErrors)
				}
			}

			moduleResolver := NewModuleResolver(symbolTable)
			errors := moduleResolver.ResolveModules(file)

			if tt.expectErrors {
				if len(errors) == 0 {
					t.Error("expected module resolution errors but got none")
				} else if tt.errorContains != "" {
					found := false
					for _, err := range errors {
						if containsSubstring([]string{err}, tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got: %v", tt.errorContains, errors)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("unexpected module resolution errors: %v", errors)
				}
			}
		})
	}
}

func TestModuleExtensionChain(t *testing.T) {
	input := `
module Base {
    var x: int
}

module Middle extends Base {
    var y: int
}

module Top extends Middle {
    var z: int
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

	moduleResolver := NewModuleResolver(symbolTable)
	errors := moduleResolver.ResolveModules(file)
	if len(errors) > 0 {
		t.Errorf("unexpected module resolution errors: %v", errors)
	}

	// Verify all modules exist
	modules := moduleResolver.GetModules()
	if len(modules) != 3 {
		t.Errorf("expected 3 modules, got %d", len(modules))
	}

	// Verify extension chain
	top, exists := modules["Top"]
	if !exists {
		t.Error("module Top not found")
	} else if top.Extends != "Middle" {
		t.Errorf("expected Top to extend Middle, got %s", top.Extends)
	}

	middle, exists := modules["Middle"]
	if !exists {
		t.Error("module Middle not found")
	} else if middle.Extends != "Base" {
		t.Errorf("expected Middle to extend Base, got %s", middle.Extends)
	}

	base, exists := modules["Base"]
	if !exists {
		t.Error("module Base not found")
	} else if base.Extends != "" {
		t.Errorf("expected Base to have no extension, got %s", base.Extends)
	}
}

func TestModuleResolverGetModule(t *testing.T) {
	input := `
module Counter {
    var count: int
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

	moduleResolver := NewModuleResolver(symbolTable)
	errors := moduleResolver.ResolveModules(file)
	if len(errors) > 0 {
		t.Errorf("unexpected module resolution errors: %v", errors)
	}

	// Test GetModule
	module, found := moduleResolver.GetModule("Counter")
	if !found {
		t.Error("module Counter not found")
	} else if module.Name != "Counter" {
		t.Errorf("expected module name Counter, got %s", module.Name)
	}

	// Test GetModule for non-existent module
	_, found = moduleResolver.GetModule("NonExistent")
	if found {
		t.Error("expected module NonExistent not to be found")
	}
}

