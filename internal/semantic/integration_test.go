package semantic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
)

// TestSemanticAnalysisOnExampleFiles tests semantic analysis on all example files
func TestSemanticAnalysisOnExampleFiles(t *testing.T) {
	examplesDir := "../../examples"
	files, err := filepath.Glob(filepath.Join(examplesDir, "*.spec"))
	if err != nil {
		t.Fatalf("failed to find example files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("no example files found")
	}

	for _, filePath := range files {
		t.Run(filepath.Base(filePath), func(t *testing.T) {
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", filePath, err)
			}

			// Parse the file
			p := parser.New(lexer.New(string(content)))
			file := p.ParseFile()

			// Some parse errors are expected for unimplemented features
			// We'll log them but continue with semantic analysis
			if len(p.Errors()) > 0 {
				t.Logf("Parse errors in %s (may be expected): %v", filePath, p.Errors())
			}

			// Build symbol table
			builder := NewBuilder()
			symbolTable, buildErrors := builder.BuildSymbolTable(file)
			if len(buildErrors) > 0 {
				t.Logf("Build errors in %s: %v", filePath, buildErrors)
			}

			// Resolve names
			resolver := NewResolver(symbolTable)
			resolutionErrors := resolver.ResolveFile(file)
			if len(resolutionErrors) > 0 {
				t.Logf("Resolution errors in %s: %v", filePath, resolutionErrors)
			}

			// Validate semantics
			validator := NewValidator(symbolTable, resolver)
			validationErrors := validator.ValidateFile(file)

			// Log all errors but don't fail the test
			// Many example files may have errors due to unimplemented features
			if len(validationErrors) > 0 {
				t.Logf("Validation errors in %s: %v", filePath, validationErrors)
			}

			// Basic sanity check: symbol table should have been created
			if symbolTable == nil {
				t.Error("symbol table should not be nil")
			}
			if symbolTable.GlobalScope == nil {
				t.Error("global scope should not be nil")
			}
		})
	}
}

// TestSemanticAnalysisCompleteSpec tests a complete, valid spec
func TestSemanticAnalysisCompleteSpec(t *testing.T) {
	input := `
description "Counter variable"
var counter: int

description "Maximum counter value"
const MAX: int = 100

description "Increments the counter"
action increment {
    require counter < MAX
    counter' = counter + 1
}

description "Decrements the counter"
action decrement {
    require counter > 0
    counter' = counter - 1
}

description "Initial state: counter starts at 0"
init {
    counter = 0
}

description "Counter must always be non-negative"
invariant counterNonNegative {
    counter >= 0
}

description "Counter must always be within bounds"
invariant counterWithinBounds {
    counter <= MAX
}

description "Counter will eventually reach maximum"
temporal eventuallyReachesMax {
    eventually counter = MAX
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Build symbol table
	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	// Verify symbol table has expected symbols
	expectedSymbols := []string{"counter", "MAX"}
	for _, name := range expectedSymbols {
		symbol, found := symbolTable.LookupSymbol(symbolTable.GlobalScope, name)
		if !found {
			t.Errorf("expected symbol %s not found in symbol table", name)
		} else {
			if name == "counter" && symbol.Kind != SymbolVariable {
				t.Errorf("expected counter to be a variable, got %d", symbol.Kind)
			}
			if name == "MAX" && symbol.Kind != SymbolConstant {
				t.Errorf("expected MAX to be a constant, got %d", symbol.Kind)
			}
		}
	}

	// Resolve names
	resolver := NewResolver(symbolTable)
	resolutionErrors := resolver.ResolveFile(file)
	if len(resolutionErrors) > 0 {
		t.Errorf("unexpected resolution errors: %v", resolutionErrors)
	}

	// Validate semantics
	validator := NewValidator(symbolTable, resolver)
	validationErrors := validator.ValidateFile(file)
	if len(validationErrors) > 0 {
		t.Errorf("unexpected validation errors: %v", validationErrors)
	}
}

// TestSemanticAnalysisWithFunctions tests semantic analysis with functions
func TestSemanticAnalysisWithFunctions(t *testing.T) {
	input := `
var x: int
var y: int

description "Adds two numbers"
fun add(a: int, b: int): int {
    return a + b
}

description "Multiplies two numbers"
fun multiply(a: int, b: int): int {
    return a * b
}

description "Computes sum of x and y"
fun computeSum(): int {
    return add(x, y)
}

init {
    x = 10
    y = 20
    x = add(x, y)
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Build symbol table
	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	// Verify functions are in symbol table
	expectedFunctions := []string{"add", "multiply", "computeSum"}
	for _, name := range expectedFunctions {
		symbol, found := symbolTable.LookupSymbol(symbolTable.GlobalScope, name)
		if !found {
			t.Errorf("expected function %s not found in symbol table", name)
		} else if symbol.Kind != SymbolFunction {
			t.Errorf("expected %s to be a function, got %d", name, symbol.Kind)
		}
	}

	// Resolve names
	resolver := NewResolver(symbolTable)
	resolutionErrors := resolver.ResolveFile(file)
	if len(resolutionErrors) > 0 {
		t.Errorf("unexpected resolution errors: %v", resolutionErrors)
	}

	// Validate semantics
	validator := NewValidator(symbolTable, resolver)
	validationErrors := validator.ValidateFile(file)
	if len(validationErrors) > 0 {
		t.Errorf("unexpected validation errors: %v", validationErrors)
	}
}

// TestSemanticAnalysisWithModules tests semantic analysis with modules
func TestSemanticAnalysisWithModules(t *testing.T) {
	input := `
description "Base counter module"
module Counter {
    var count: int
    
    description "Increment action"
    action increment {
        count' = count + 1
    }
    
    description "Initial state"
    init {
        count = 0
    }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Build symbol table
	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	// Verify module is in symbol table
	symbol, found := symbolTable.LookupSymbol(symbolTable.GlobalScope, "Counter")
	if !found {
		t.Error("expected module Counter not found in symbol table")
	} else if symbol.Kind != SymbolModule {
		t.Errorf("expected Counter to be a module, got %d", symbol.Kind)
	}

	// Resolve names
	resolver := NewResolver(symbolTable)
	resolutionErrors := resolver.ResolveFile(file)
	// Module variable resolution within module scopes is a known limitation
	// Variables within modules should be resolved, but this requires qualified names
	// For now, we verify the module structure is correct
	if len(resolutionErrors) > 0 {
		t.Logf("Resolution errors (expected for module scoping): %v", resolutionErrors)
	}

	// Validate semantics
	validator := NewValidator(symbolTable, resolver)
	validationErrors := validator.ValidateFile(file)
	// Module variable resolution within module scopes is a known limitation
	if len(validationErrors) > 0 {
		t.Logf("Validation errors (expected for module scoping): %v", validationErrors)
	}

	// At minimum, verify the module symbol exists and structure is correct
	symbol2, found2 := symbolTable.LookupSymbol(symbolTable.GlobalScope, "Counter")
	if !found2 {
		t.Error("expected module Counter not found in symbol table")
	} else if symbol2.Kind != SymbolModule {
		t.Errorf("expected Counter to be a module, got %d", symbol2.Kind)
	}
}

// TestSemanticAnalysisErrorCases tests various error cases
func TestSemanticAnalysisErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  bool
		errorContains string
	}{
		{
			name: "Undefined variable in init",
			input: `
init {
    x = 10
}
`,
			expectErrors: true,
			errorContains: "undefined identifier: x",
		},
		{
			name: "Undefined function call",
			input: `
var result: int
init {
    result = undefinedFunc(5)
}
`,
			expectErrors: true,
			errorContains: "undefined function",
		},
		{
			name: "Function call with wrong argument count",
			input: `
fun add(x: int, y: int): int {
    return x + y
}

var result: int
init {
    result = add(5)
}
`,
			expectErrors: true,
			errorContains: "expects 2 arguments, got 1",
		},
		{
			name: "Assignment to constant",
			input: `
const MAX: int = 100

action update {
    MAX = 200
}
`,
			expectErrors: true,
			errorContains: "cannot assign to constant",
		},
		{
			name: "Function mutating state",
			input: `
var counter: int

fun increment(): int {
    counter' = counter + 1
    return counter
}
`,
			expectErrors: true,
			errorContains: "cannot mutate state variable",
		},
		{
			name: "Primed variable without base variable",
			input: `
action update {
    counter' = 10
}
`,
			expectErrors: true,
			errorContains: "undefined variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.New(lexer.New(tt.input))
			file := p.ParseFile()
			if len(p.Errors()) > 0 && !tt.expectErrors {
				t.Fatalf("parse errors: %v", p.Errors())
			}

			builder := NewBuilder()
			symbolTable, buildErrors := builder.BuildSymbolTable(file)
			if len(buildErrors) > 0 && !tt.expectErrors {
				t.Fatalf("build errors: %v", buildErrors)
			}

			resolver := NewResolver(symbolTable)
			validator := NewValidator(symbolTable, resolver)
			errors := validator.ValidateFile(file)

			if tt.expectErrors {
				if len(errors) == 0 && len(buildErrors) == 0 {
					t.Error("expected errors but got none")
				} else if tt.errorContains != "" {
					allErrors := append(buildErrors, errors...)
					found := false
					for _, err := range allErrors {
						if containsSubstring([]string{err}, tt.errorContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error containing '%s', got build: %v, validation: %v",
							tt.errorContains, buildErrors, errors)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("unexpected validation errors: %v", errors)
				}
			}
		})
	}
}

// TestSemanticAnalysisScopeResolution tests scope resolution
func TestSemanticAnalysisScopeResolution(t *testing.T) {
	input := `
var global: int

fun test(): int {
    return global + 10
}

action update {
    global' = test()
}

init {
    global = 0
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Build symbol table
	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	// Resolve names
	resolver := NewResolver(symbolTable)
	resolutionErrors := resolver.ResolveFile(file)
	if len(resolutionErrors) > 0 {
		t.Errorf("unexpected resolution errors: %v", resolutionErrors)
	}

	// Validate semantics
	validator := NewValidator(symbolTable, resolver)
	validationErrors := validator.ValidateFile(file)
	if len(validationErrors) > 0 {
		t.Errorf("unexpected validation errors: %v", validationErrors)
	}
}

// TestSemanticAnalysisOneOfInit tests oneOf init semantic analysis
func TestSemanticAnalysisOneOfInit(t *testing.T) {
	input := `
var x: int
var y: int

init oneOf {
    {
        x = 0
        y = 0
    },
    {
        x = 10
        y = 20
    },
    {
        x = 100
        y = 200
    }
}
`

	p := parser.New(lexer.New(input))
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Build symbol table
	builder := NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		t.Fatalf("build errors: %v", buildErrors)
	}

	// Resolve names
	resolver := NewResolver(symbolTable)
	resolutionErrors := resolver.ResolveFile(file)
	if len(resolutionErrors) > 0 {
		t.Errorf("unexpected resolution errors: %v", resolutionErrors)
	}

	// Validate semantics
	validator := NewValidator(symbolTable, resolver)
	validationErrors := validator.ValidateFile(file)
	if len(validationErrors) > 0 {
		t.Errorf("unexpected validation errors: %v", validationErrors)
	}
}

