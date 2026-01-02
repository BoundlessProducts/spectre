package semantic

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/internal/parser"
)

func TestValidateFunctionCallArgumentCount(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  bool
		errorContains string
	}{
		{
			name: "Correct argument count",
			input: `
fun add(x: int, y: int) : int {
    return x + y
}

var result : int
init {
    result = add(5, 10)
}
`,
			expectErrors: false,
		},
		{
			name: "Too few arguments",
			input: `
fun add(x: int, y: int) : int {
    return x + y
}

var result : int
init {
    result = add(5)
}
`,
			expectErrors: true,
			errorContains: "expects 2 arguments, got 1",
		},
		{
			name: "Too many arguments",
			input: `
fun add(x: int, y: int) : int {
    return x + y
}

var result : int
init {
    result = add(5, 10, 15)
}
`,
			expectErrors: true,
			errorContains: "expects 2 arguments, got 3",
		},
		{
			name: "Function with no parameters",
			input: `
fun getValue() : int {
    return 42
}

var result : int
init {
    result = getValue()
}
`,
			expectErrors: false,
		},
		{
			name: "Function with no parameters called with arguments",
			input: `
fun getValue() : int {
    return 42
}

var result : int
init {
    result = getValue(10)
}
`,
			expectErrors: true,
			errorContains: "expects 0 arguments, got 1",
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

			resolver := NewResolver(symbolTable)
			validator := NewValidator(symbolTable, resolver)
			errors := validator.ValidateFile(file)

			if tt.expectErrors {
				if len(errors) == 0 {
					t.Error("expected validation errors but got none")
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
					t.Errorf("unexpected validation errors: %v", errors)
				}
			}
		})
	}
}

func TestValidateFunctionCallUndefined(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  bool
		errorContains string
	}{
		{
			name: "Call to undefined function",
			input: `
var result : int
init {
    result = undefinedFunc(5)
}
`,
			expectErrors: true,
			errorContains: "undefined function",
		},
		{
			name: "Call to variable as function",
			input: `
var x : int
var result : int
init {
    result = x(5)
}
`,
			expectErrors: true,
			errorContains: "cannot call non-function",
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

			resolver := NewResolver(symbolTable)
			validator := NewValidator(symbolTable, resolver)
			errors := validator.ValidateFile(file)

			if tt.expectErrors {
				if len(errors) == 0 {
					t.Error("expected validation errors but got none")
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
					t.Errorf("unexpected validation errors: %v", errors)
				}
			}
		})
	}
}

func TestValidateNestedFunctionCalls(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  bool
	}{
		{
			name: "Nested function calls",
			input: `
fun add(x: int, y: int) : int {
    return x + y
}

fun multiply(x: int, y: int) : int {
    return x * y
}

var result : int
init {
    result = add(multiply(2, 3), multiply(4, 5))
}
`,
			expectErrors: false,
		},
		{
			name: "Nested function calls with wrong argument count",
			input: `
fun add(x: int, y: int) : int {
    return x + y
}

var result : int
init {
    result = add(add(5), 10)
}
`,
			expectErrors: true,
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

			resolver := NewResolver(symbolTable)
			validator := NewValidator(symbolTable, resolver)
			errors := validator.ValidateFile(file)

			if tt.expectErrors {
				if len(errors) == 0 {
					t.Error("expected validation errors but got none")
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("unexpected validation errors: %v", errors)
				}
			}
		})
	}
}

func TestValidateFunctionCallsInExpressions(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectErrors  bool
	}{
		{
			name: "Function call in assignment",
			input: `
fun getValue() : int {
    return 10
}

var result : int
init {
    result = getValue()
}
`,
			expectErrors: false,
		},
		{
			name: "Function call in if expression",
			input: `
fun isPositive(x: int) : bool {
    return x > 0
}

var result : bool
init {
    result = if (isPositive(5)) { true } else { false }
}
`,
			expectErrors: false,
		},
		{
			name: "Function call in assignment with parameter",
			input: `
fun double(x: int) : int {
    return x * 2
}

var result : int
init {
    result = double(5)
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

			resolver := NewResolver(symbolTable)
			validator := NewValidator(symbolTable, resolver)
			errors := validator.ValidateFile(file)

			if tt.expectErrors {
				if len(errors) == 0 {
					t.Error("expected validation errors but got none")
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("unexpected validation errors: %v", errors)
				}
			}
		})
	}
}

