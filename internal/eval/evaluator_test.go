package eval

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// parseExpressionForTest is a helper to parse expressions for testing
func parseExpressionForTest(p *parser.Parser, input string) ast.Expr {
	// Parse expression by wrapping it in a function and extracting the return value
	wrapped := "fun test() { return " + input + " }"
	l := lexer.New(wrapped)
	pp := parser.New(l)
	file := pp.ParseFile()
	if len(pp.Errors()) > 0 {
		return nil
	}
	// Find the function and extract the return expression
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FunctionDecl); ok && fn.Name == "test" {
			if fn.Body != nil && len(fn.Body.Statements) > 0 {
				if ret, ok := fn.Body.Statements[0].(*ast.ReturnStmt); ok {
					return ret.Value
				}
			}
		}
	}
	return nil
}

func TestEvaluatorBasicLiterals(t *testing.T) {
	env := NewEnvironment()
	eval := NewEvaluator(env)

	tests := []struct {
		name     string
		input    string
		expected state.Value
	}{
		{
			name:     "integer literal",
			input:    "42",
			expected: state.NewIntValue(42),
		},
		{
			name:     "float literal",
			input:    "3.14",
			expected: state.NewFloatValue(3.14),
		},
		{
			name:     "string literal",
			input:    `"hello"`,
			expected: state.NewStringValue("hello"),
		},
		{
			name:     "boolean true",
			input:    "true",
			expected: state.NewBoolValue(true),
		},
		{
			name:     "boolean false",
			input:    "false",
			expected: state.NewBoolValue(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parseExpressionForTest(nil, tt.input)
			if expr == nil {
				t.Fatalf("failed to parse expression: %s", tt.input)
			}

			result, err := eval.Eval(expr)
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestEvaluatorVariables(t *testing.T) {
	env := NewEnvironment()
	env.SetVariable("x", state.NewIntValue(10))
	env.SetVariable("y", state.NewIntValue(20))

	eval := NewEvaluator(env)

	// Test variable lookup
	expr := parseExpressionForTest(nil, "x")
	if expr == nil {
		t.Fatalf("failed to parse expression: x")
	}

	result, err := eval.Eval(expr)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}

	expected := state.NewIntValue(10)
	if !valuesEqual(result, expected) {
		t.Errorf("expected %s, got %s", expected.String(), result.String())
	}
}

func TestEvaluatorBinaryExpressions(t *testing.T) {
	env := NewEnvironment()
	eval := NewEvaluator(env)

	tests := []struct {
		name     string
		input    string
		expected state.Value
	}{
		{
			name:     "addition",
			input:    "10 + 20",
			expected: state.NewIntValue(30),
		},
		{
			name:     "subtraction",
			input:    "20 - 10",
			expected: state.NewIntValue(10),
		},
		{
			name:     "multiplication",
			input:    "5 * 6",
			expected: state.NewIntValue(30),
		},
		{
			name:     "division",
			input:    "20 / 4",
			expected: state.NewIntValue(5),
		},
		{
			name:     "equality true",
			input:    "10 = 10",
			expected: state.NewBoolValue(true),
		},
		{
			name:     "equality false",
			input:    "10 = 20",
			expected: state.NewBoolValue(false),
		},
		{
			name:     "less than true",
			input:    "5 < 10",
			expected: state.NewBoolValue(true),
		},
		{
			name:     "less than false",
			input:    "10 < 5",
			expected: state.NewBoolValue(false),
		},
		{
			name:     "greater than true",
			input:    "10 > 5",
			expected: state.NewBoolValue(true),
		},
		{
			name:     "logical AND true",
			input:    "true && true",
			expected: state.NewBoolValue(true),
		},
		{
			name:     "logical AND false",
			input:    "true && false",
			expected: state.NewBoolValue(false),
		},
		{
			name:     "logical OR true",
			input:    "true || false",
			expected: state.NewBoolValue(true),
		},
		{
			name:     "logical OR false",
			input:    "false || false",
			expected: state.NewBoolValue(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parseExpressionForTest(nil, tt.input)
			if expr == nil {
				t.Fatalf("failed to parse expression: %s", tt.input)
			}

			result, err := eval.Eval(expr)
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestEvaluatorUnaryExpressions(t *testing.T) {
	env := NewEnvironment()
	eval := NewEvaluator(env)

	tests := []struct {
		name     string
		input    string
		expected state.Value
	}{
		{
			name:     "unary minus",
			input:    "-10",
			expected: state.NewIntValue(-10),
		},
		{
			name:     "logical NOT true",
			input:    "!true",
			expected: state.NewBoolValue(false),
		},
		{
			name:     "logical NOT false",
			input:    "!false",
			expected: state.NewBoolValue(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parseExpressionForTest(nil, tt.input)
			if expr == nil {
				t.Fatalf("failed to parse expression: %s", tt.input)
			}

			result, err := eval.Eval(expr)
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestEvaluatorIfExpression(t *testing.T) {
	env := NewEnvironment()
	eval := NewEvaluator(env)

	tests := []struct {
		name     string
		input    string
		expected state.Value
	}{
		{
			name:     "if true",
			input:    "if (true) { 10 } else { 20 }",
			expected: state.NewIntValue(10),
		},
		{
			name:     "if false",
			input:    "if (false) { 10 } else { 20 }",
			expected: state.NewIntValue(20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parseExpressionForTest(nil, tt.input)
			if expr == nil {
				t.Fatalf("failed to parse expression: %s", tt.input)
			}

			result, err := eval.Eval(expr)
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

func TestEvaluatorLetExpression(t *testing.T) {
	env := NewEnvironment()
	eval := NewEvaluator(env)

	// Create a simple let expression manually for testing
	letExpr := &ast.LetExpr{
		Name:  "x",
		Value: &ast.BasicLit{Kind: ast.IntLit, Value: "10"},
		Body:  &ast.Ident{Name: "x"},
	}

	result, err := eval.Eval(letExpr)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}

	expected := state.NewIntValue(10)
	if !valuesEqual(result, expected) {
		t.Errorf("expected %s, got %s", expected.String(), result.String())
	}
}

func TestEvaluatorFunctionCall(t *testing.T) {
	env := NewEnvironment()

	// Define a simple function: fun add(x: int, y: int): int { return x + y }
	fnDecl := &ast.FunctionDecl{
		Name: "add",
		Parameters: []ast.Parameter{
			{Name: "x", Type: &ast.PrimitiveType{Name: "int"}},
			{Name: "y", Type: &ast.PrimitiveType{Name: "int"}},
		},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.BinaryExpr{
						Op:    ast.Add,
						Left:  &ast.Ident{Name: "x"},
						Right: &ast.Ident{Name: "y"},
					},
				},
			},
		},
	}

	fnDef := &FunctionDef{
		Decl:   fnDecl,
		Params: fnDecl.Parameters,
		Body:   fnDecl.Body,
	}

	env.DefineFunction("add", fnDef)

	eval := NewEvaluator(env)

	// Test function call: add(5, 3)
	callExpr := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit, Value: "5"},
			&ast.BasicLit{Kind: ast.IntLit, Value: "3"},
		},
	}

	result, err := eval.Eval(callExpr)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}

	expected := state.NewIntValue(8)
	if !valuesEqual(result, expected) {
		t.Errorf("expected %s, got %s", expected.String(), result.String())
	}
}

func TestEvaluatorNestedExpressions(t *testing.T) {
	env := NewEnvironment()
	eval := NewEvaluator(env)

	tests := []struct {
		name     string
		input    string
		expected state.Value
	}{
		{
			name:     "nested addition",
			input:    "(10 + 20) + 30",
			expected: state.NewIntValue(60),
		},
		{
			name:     "complex expression",
			input:    "(10 + 5) * 2",
			expected: state.NewIntValue(30),
		},
		{
			name:     "nested comparison",
			input:    "(5 + 5) = 10",
			expected: state.NewBoolValue(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parseExpressionForTest(nil, tt.input)
			if expr == nil {
				t.Fatalf("failed to parse expression: %s", tt.input)
			}

			result, err := eval.Eval(expr)
			if err != nil {
				t.Fatalf("evaluation error: %v", err)
			}

			if !valuesEqual(result, tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected.String(), result.String())
			}
		})
	}
}

// valuesEqual is a helper to compare values
func valuesEqual(v1, v2 state.Value) bool {
	if v1.Type() != v2.Type() {
		return false
	}
	return v1.String() == v2.String()
}

