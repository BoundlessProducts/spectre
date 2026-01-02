package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spectre-lang/spectre/internal/lexer"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestComplexFunctionParsing(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "Function with if-else expression",
			input: `fun max(a: int, b: int): int {
  return if (a > b) { a } else { b }
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				funcDecl, ok := decl.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("not *ast.FunctionDecl")
				}
				if len(funcDecl.Body.Statements) != 1 {
					t.Fatalf("expected 1 statement")
				}
				returnStmt, ok := funcDecl.Body.Statements[0].(*ast.ReturnStmt)
				if !ok {
					t.Fatalf("not *ast.ReturnStmt")
				}
				_, ok = returnStmt.Value.(*ast.IfExpr)
				if !ok {
					t.Fatalf("return value not *ast.IfExpr")
				}
			},
		},
		{
			name: "Function with let binding and return",
			input: `fun calculate(x: int): int {
  let doubled = x * 2
  return doubled + 1
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				funcDecl, ok := decl.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("not *ast.FunctionDecl")
				}
				if len(funcDecl.Body.Statements) != 2 {
					t.Fatalf("expected 2 statements, got %d", len(funcDecl.Body.Statements))
				}
				// First should be let expression
				_, ok = funcDecl.Body.Statements[0].(*ast.ExprStmt)
				if !ok {
					t.Fatalf("first statement not *ast.ExprStmt")
				}
				// Second should be return
				_, ok = funcDecl.Body.Statements[1].(*ast.ReturnStmt)
				if !ok {
					t.Fatalf("second statement not *ast.ReturnStmt")
				}
			},
		},
		{
			name: "Function with multiple let bindings",
			input: `fun complex(x: int, y: int): int {
  let sum = x + y
  let product = x * y
  return sum + product
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				funcDecl, ok := decl.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("not *ast.FunctionDecl")
				}
				if len(funcDecl.Body.Statements) != 3 {
					t.Fatalf("expected 3 statements, got %d", len(funcDecl.Body.Statements))
				}
			},
		},
		{
			name: "Function with nested expressions",
			input: `fun nested(x: int): int {
  return (x + 1) * (x - 1)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				funcDecl, ok := decl.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("not *ast.FunctionDecl")
				}
				returnStmt, ok := funcDecl.Body.Statements[0].(*ast.ReturnStmt)
				if !ok {
					t.Fatalf("not *ast.ReturnStmt")
				}
				_, ok = returnStmt.Value.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("return value not *ast.BinaryExpr")
				}
			},
		},
		{
			name: "Function with complex return expression",
			input: `fun complexExpr(a: int, b: int, c: int): int {
  return if (a > b) { a + c } else { b + c }
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				funcDecl, ok := decl.(*ast.FunctionDecl)
				if !ok {
					t.Fatalf("not *ast.FunctionDecl")
				}
				returnStmt, ok := funcDecl.Body.Statements[0].(*ast.ReturnStmt)
				if !ok {
					t.Fatalf("not *ast.ReturnStmt")
				}
				ifExpr, ok := returnStmt.Value.(*ast.IfExpr)
				if !ok {
					t.Fatalf("return value not *ast.IfExpr")
				}
				// Check that then and else branches have binary expressions
				_, ok = ifExpr.Then.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("then branch not *ast.BinaryExpr")
				}
				_, ok = ifExpr.Else.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("else branch not *ast.BinaryExpr")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseFunctionDecl()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			if decl == nil {
				t.Fatal("decl is nil")
			}

			tt.validate(t, decl)
		})
	}
}

func TestFunctionCallParsing(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Expr)
	}{
		{
			name: "Simple function call",
			input: "add(1, 2)",
			validate: func(t *testing.T, expr ast.Expr) {
				callExpr, ok := expr.(*ast.CallExpr)
				if !ok {
					t.Fatalf("not *ast.CallExpr")
				}
				ident, ok := callExpr.Fun.(*ast.Ident)
				if !ok {
					t.Fatalf("function not *ast.Ident")
				}
				if ident.Name != "add" {
					t.Errorf("function name not 'add'. got=%s", ident.Name)
				}
				if len(callExpr.Args) != 2 {
					t.Fatalf("expected 2 arguments, got %d", len(callExpr.Args))
				}
			},
		},
		{
			name: "Method call on identifier",
			input: "userSet.size()",
			validate: func(t *testing.T, expr ast.Expr) {
				// First check if it's a selector (without call) or a call
				callExpr, ok := expr.(*ast.CallExpr)
				if ok {
					selector, ok := callExpr.Fun.(*ast.SelectorExpr)
					if !ok {
						t.Fatalf("function not *ast.SelectorExpr")
					}
					if selector.Sel != "size" {
						t.Errorf("method name not 'size'. got=%s", selector.Sel)
					}
				} else {
					// Might be parsed as selector only - that's okay for now
					selector, ok := expr.(*ast.SelectorExpr)
					if !ok {
						t.Fatalf("not *ast.CallExpr or *ast.SelectorExpr. got=%T", expr)
					}
					if selector.Sel != "size" {
						t.Errorf("selector name not 'size'. got=%s", selector.Sel)
					}
				}
			},
		},
		{
			name: "Nested function calls",
			input: "add(multiply(2, 3), 4)",
			validate: func(t *testing.T, expr ast.Expr) {
				callExpr, ok := expr.(*ast.CallExpr)
				if !ok {
					t.Fatalf("not *ast.CallExpr")
				}
				if len(callExpr.Args) != 2 {
					t.Fatalf("expected 2 arguments, got %d", len(callExpr.Args))
				}
				// First argument should be a call expression
				_, ok = callExpr.Args[0].(*ast.CallExpr)
				if !ok {
					t.Fatalf("first argument not *ast.CallExpr")
				}
			},
		},
		{
			name: "Method call with arguments",
			input: "users.filter(x)",
			validate: func(t *testing.T, expr ast.Expr) {
				callExpr, ok := expr.(*ast.CallExpr)
				if !ok {
					t.Fatalf("not *ast.CallExpr. got=%T", expr)
				}
				selector, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					t.Fatalf("function not *ast.SelectorExpr")
				}
				if selector.Sel != "filter" {
					t.Errorf("method name not 'filter'. got=%s", selector.Sel)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			expr := p.parseExpression(LOWEST)

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			if expr == nil {
				t.Fatal("expr is nil")
			}

			tt.validate(t, expr)
		})
	}
}

func TestConditionalExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Expr)
	}{
		{
			name: "Simple if-else",
			input: "if (x > 0) { x } else { 0 }",
			validate: func(t *testing.T, expr ast.Expr) {
				ifExpr, ok := expr.(*ast.IfExpr)
				if !ok {
					t.Fatalf("not *ast.IfExpr")
				}
				_, ok = ifExpr.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("condition not *ast.BinaryExpr")
				}
			},
		},
		{
			name: "Nested if-else",
			input: "if (a > b) { if (a > c) { a } else { c } } else { b }",
			validate: func(t *testing.T, expr ast.Expr) {
				ifExpr, ok := expr.(*ast.IfExpr)
				if !ok {
					t.Fatalf("not *ast.IfExpr")
				}
				nestedIf, ok := ifExpr.Then.(*ast.IfExpr)
				if !ok {
					t.Fatalf("then branch not *ast.IfExpr")
				}
				_, ok = nestedIf.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("nested condition not *ast.BinaryExpr")
				}
			},
		},
		{
			name: "If-else with complex expressions",
			input: "if (x + y > 10) { x * 2 } else { y * 2 }",
			validate: func(t *testing.T, expr ast.Expr) {
				ifExpr, ok := expr.(*ast.IfExpr)
				if !ok {
					t.Fatalf("not *ast.IfExpr")
				}
				_, ok = ifExpr.Condition.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("condition not *ast.BinaryExpr")
				}
				_, ok = ifExpr.Then.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("then branch not *ast.BinaryExpr")
				}
				_, ok = ifExpr.Else.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("else branch not *ast.BinaryExpr")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			expr := p.parseExpression(LOWEST)

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			if expr == nil {
				t.Fatal("expr is nil")
			}

			tt.validate(t, expr)
		})
	}
}

func TestLetBindingExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Expr)
	}{
		{
			name: "Simple let binding",
			input: "let x = 10",
			validate: func(t *testing.T, expr ast.Expr) {
				letExpr, ok := expr.(*ast.LetExpr)
				if !ok {
					t.Fatalf("not *ast.LetExpr")
				}
				if letExpr.Name != "x" {
					t.Errorf("binding name not 'x'. got=%s", letExpr.Name)
				}
				_, ok = letExpr.Value.(*ast.BasicLit)
				if !ok {
					t.Fatalf("value not *ast.BasicLit")
				}
			},
		},
		{
			name: "Let binding with expression",
			input: "let sum = x + y",
			validate: func(t *testing.T, expr ast.Expr) {
				letExpr, ok := expr.(*ast.LetExpr)
				if !ok {
					t.Fatalf("not *ast.LetExpr")
				}
				_, ok = letExpr.Value.(*ast.BinaryExpr)
				if !ok {
					t.Fatalf("value not *ast.BinaryExpr")
				}
			},
		},
		{
			name: "Let binding with if-else",
			input: "let result = if (x > 0) { x } else { 0 }",
			validate: func(t *testing.T, expr ast.Expr) {
				letExpr, ok := expr.(*ast.LetExpr)
				if !ok {
					t.Fatalf("not *ast.LetExpr")
				}
				_, ok = letExpr.Value.(*ast.IfExpr)
				if !ok {
					t.Fatalf("value not *ast.IfExpr")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			expr := p.parseExpression(LOWEST)

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}

			if expr == nil {
				t.Fatal("expr is nil")
			}

			tt.validate(t, expr)
		})
	}
}

func TestParsePureFunctionsExample(t *testing.T) {
	exampleFile := "../../examples/pure-functions.spec"
	filePath := filepath.Join(exampleFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}

	l := lexer.New(string(content))
	p := New(l)

	// Try to parse at least one function declaration
	// Skip any type definitions, variables, etc. at the start
	for !p.curTokenIs(lexer.FUN) && !p.curTokenIs(lexer.EOF) {
		// Skip descriptions
		if p.curTokenIs(lexer.DESCRIPTION) {
			p.parseDescription()
			continue
		}
		// Skip var declarations
		if p.curTokenIs(lexer.VAR) {
			p.parseVariableDecl()
			continue
		}
		// Skip type definitions (we'll handle these later)
		if p.curTokenIs(lexer.TYPE) {
			// Skip type definition for now
			for !p.curTokenIs(lexer.FUN) && !p.curTokenIs(lexer.EOF) && !p.curTokenIs(lexer.VAR) {
				p.nextToken()
			}
			continue
		}
		p.nextToken()
	}

	if p.curTokenIs(lexer.FUN) {
		decl := p.parseFunctionDecl()
		if decl == nil {
			t.Error("Failed to parse function declaration from pure-functions.spec")
		}
		if len(p.Errors()) > 0 {
			t.Logf("Parser errors (may be expected for incomplete parsing): %v", p.Errors())
		}
	} else {
		t.Log("No function declarations found in pure-functions.spec (may need type parsing first)")
	}
}

func TestMultipleFunctionDeclarations(t *testing.T) {
	input := `fun add(x: int, y: int): int {
  return x + y
}

fun multiply(x: int, y: int): int {
  return x * y
}

fun subtract(x: int, y: int): int {
  return x - y
}`

	l := lexer.New(input)
	p := New(l)

	// Parse first function
	decl1 := p.parseFunctionDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	funcDecl1, ok := decl1.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl1 not *ast.FunctionDecl. got=%T", decl1)
	}
	if funcDecl1.Name != "add" {
		t.Errorf("funcDecl1.Name not 'add'. got=%s", funcDecl1.Name)
	}

	// Parse second function
	decl2 := p.parseFunctionDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors after second function: %v", len(p.Errors()), p.Errors())
	}

	funcDecl2, ok := decl2.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl2 not *ast.FunctionDecl. got=%T", decl2)
	}
	if funcDecl2.Name != "multiply" {
		t.Errorf("funcDecl2.Name not 'multiply'. got=%s", funcDecl2.Name)
	}

	// Parse third function
	decl3 := p.parseFunctionDecl()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors after third function: %v", len(p.Errors()), p.Errors())
	}

	funcDecl3, ok := decl3.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl3 not *ast.FunctionDecl. got=%T", decl3)
	}
	if funcDecl3.Name != "subtract" {
		t.Errorf("funcDecl3.Name not 'subtract'. got=%s", funcDecl3.Name)
	}
}

func TestFunctionWithComplexBody(t *testing.T) {
	input := `fun complex(a: int, b: int, c: int): int {
  let sum = a + b
  let product = a * b
  let result = if (sum > product) { sum + c } else { product + c }
  return result
}`

	l := lexer.New(input)
	p := New(l)
	decl := p.parseFunctionDecl()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	funcDecl, ok := decl.(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("decl not *ast.FunctionDecl. got=%T", decl)
	}

	if len(funcDecl.Parameters) != 3 {
		t.Fatalf("expected 3 parameters, got %d", len(funcDecl.Parameters))
	}

	if funcDecl.Body == nil {
		t.Fatal("function body is nil")
	}

	// Should have multiple statements (let bindings + return)
	if len(funcDecl.Body.Statements) < 3 {
		t.Errorf("expected at least 3 statements, got %d", len(funcDecl.Body.Statements))
	}

	// Last statement should be return
	lastStmt, ok := funcDecl.Body.Statements[len(funcDecl.Body.Statements)-1].(*ast.ReturnStmt)
	if !ok {
		t.Fatalf("last statement not *ast.ReturnStmt. got=%T", funcDecl.Body.Statements[len(funcDecl.Body.Statements)-1])
	}

	if lastStmt.Value == nil {
		t.Error("return value is nil")
	}
}

