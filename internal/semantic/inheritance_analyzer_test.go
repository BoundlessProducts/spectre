package semantic

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
)

func TestInheritanceAnalysis(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  bool
		errorContains string
	}{
		{
			name: "Module with extension",
			input: `
module Base {
    action increment {
        counter' = counter + 1
    }
}

module Extended extends Base {
    var counter: int
}
`,
			expectErrors: false,
		},
		{
			name: "Super call in extended module",
			input: `
module Base {
    action increment {
        counter' = counter + 1
    }
}

module Extended extends Base {
    var counter: int
    
    action increment {
        super.increment()
    }
}
`,
			expectErrors: false,
		},
		{
			name: "Super call without extension",
			input: `
module Counter {
    var counter: int
    
    action increment {
        super.increment()
    }
}
`,
			expectErrors: true,
			errorContains: "does not extend another module",
		},
		{
			name: "Super call to undefined method",
			input: `
module Base {
    action increment {
        counter' = counter + 1
    }
}

module Extended extends Base {
    var counter: int
    
    action increment {
        super.undefinedMethod()
    }
}
`,
			expectErrors: true,
			errorContains: "super method",
		},
		{
			name: "Method override with compatible signature",
			input: `
module Base {
    action increment {
        counter' = counter + 1
    }
}

module Extended extends Base {
    var counter: int
    
    action increment {
        super.increment()
    }
}
`,
			expectErrors: false,
		},
		{
			name: "Method override with incompatible parameter count",
			input: `
module Base {
    action increment {
        counter' = counter + 1
    }
}

module Extended extends Base {
    var counter: int
    
    action increment(x: int) {
        counter' = counter + x
    }
}
`,
			expectErrors: true,
			errorContains: "has 1 parameters, parent has 0",
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

			inheritanceAnalyzer := NewInheritanceAnalyzer(symbolTable, moduleResolver)
			errors := inheritanceAnalyzer.AnalyzeInheritance(file)

			if tt.expectErrors {
				if len(errors) == 0 {
					t.Error("expected inheritance analysis errors but got none")
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
					t.Errorf("unexpected inheritance analysis errors: %v", errors)
				}
			}
		})
	}
}

func TestInheritanceChain(t *testing.T) {
	input := `
module Base {
    action increment {
        counter' = counter + 1
    }
}

module Middle extends Base {
    var counter: int
}

module Top extends Middle {
    action increment {
        super.increment()
    }
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
	moduleErrors := moduleResolver.ResolveModules(file)
	if len(moduleErrors) > 0 {
		t.Fatalf("module resolution errors: %v", moduleErrors)
	}

	inheritanceAnalyzer := NewInheritanceAnalyzer(symbolTable, moduleResolver)
	errors := inheritanceAnalyzer.AnalyzeInheritance(file)
	if len(errors) > 0 {
		t.Errorf("unexpected inheritance analysis errors: %v", errors)
	}
}

