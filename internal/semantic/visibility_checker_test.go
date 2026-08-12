package semantic

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestVisibilityChecking(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  bool
		errorContains string
	}{
		{
			name: "Public variable",
			input: `
module Counter {
    public var count: int
}
`,
			expectErrors: false,
		},
		{
			name: "Private variable",
			input: `
module Counter {
    private var count: int
}
`,
			expectErrors: false,
		},
		{
			name: "Public action",
			input: `
module Counter {
    var count: int
    
    public action increment {
        count' = count + 1
    }
}
`,
			expectErrors: false,
		},
		{
			name: "Private action",
			input: `
module Counter {
    var count: int
    
    private action increment {
        count' = count + 1
    }
}
`,
			expectErrors: false,
		},
		{
			name: "Public function",
			input: `
module Counter {
    public fun getValue(): int {
        return 10
    }
}
`,
			expectErrors: false,
		},
		{
			name: "Public invariant",
			input: `
module Counter {
    var count: int
    
    public invariant nonNegative {
        count >= 0
    }
}
`,
			expectErrors: false,
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
			if len(buildErrors) > 0 {
				t.Fatalf("build errors: %v", buildErrors)
			}

			moduleResolver := NewModuleResolver(symbolTable)
			moduleErrors := moduleResolver.ResolveModules(file)
			if len(moduleErrors) > 0 {
				t.Fatalf("module resolution errors: %v", moduleErrors)
			}

			visibilityChecker := NewVisibilityChecker(symbolTable, moduleResolver)
			errors := visibilityChecker.CheckVisibility(file)

			if tt.expectErrors {
				if len(errors) == 0 {
					t.Error("expected visibility errors but got none")
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
					t.Errorf("unexpected visibility errors: %v", errors)
				}
			}
		})
	}
}

func TestVisibilityAccessControl(t *testing.T) {
	tests := []struct {
		name           string
		symbolModule   string
		accessorModule string
		visibility     ast.Visibility
		shouldAllow    bool
	}{
		{
			name:           "Same module access",
			symbolModule:   "Counter",
			accessorModule: "Counter",
			visibility:     ast.Private,
			shouldAllow:    true,
		},
		{
			name:           "Cross-module public access",
			symbolModule:   "Counter",
			accessorModule: "App",
			visibility:     ast.Public,
			shouldAllow:    true,
		},
		{
			name:           "Cross-module private access",
			symbolModule:   "Counter",
			accessorModule: "App",
			visibility:     ast.Private,
			shouldAllow:    false,
		},
		{
			name:           "Top-level access to public",
			symbolModule:   "Counter",
			accessorModule: "",
			visibility:     ast.Public,
			shouldAllow:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock symbol
			varDecl := &ast.VariableDecl{
				Position:   ast.Position{Line: 1, Column: 1},
				Name:       "count",
				Type:       &ast.PrimitiveType{Name: "int"},
				Visibility: tt.visibility,
			}

			symbolTable := NewSymbolTable()
			moduleResolver := NewModuleResolver(symbolTable)
			visibilityChecker := NewVisibilityChecker(symbolTable, moduleResolver)

			symbol := &Symbol{
				Name:        "count",
				Kind:        SymbolVariable,
				Decl:        varDecl,
				Scope:       symbolTable.GlobalScope,
				Position:    ast.Position{Line: 1, Column: 1},
				Description: "",
			}

			allowed := visibilityChecker.CheckAccess(symbol, tt.accessorModule, tt.symbolModule)
			if allowed != tt.shouldAllow {
				t.Errorf("expected access %v, got %v", tt.shouldAllow, allowed)
			}
		})
	}
}

