package eval

import (
	"testing"

	"github.com/spectre-lang/spectre/internal/state"
	"github.com/spectre-lang/spectre/pkg/ast"
)

func TestEvaluatorRecursiveFunction(t *testing.T) {
	env := NewEnvironment()

	// Define factorial function: fun factorial(n: int): int { if (n <= 1) { return 1 } else { return n * factorial(n - 1) } }
	factorialBody := &ast.BlockStmt{
		Statements: []ast.Stmt{
			&ast.ReturnStmt{
				Value: &ast.IfExpr{
					Condition: &ast.BinaryExpr{
						Op:    ast.Leq,
						Left:  &ast.Ident{Name: "n"},
						Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
					},
					Then: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
					Else: &ast.BinaryExpr{
						Op: ast.Mul,
						Left: &ast.Ident{Name: "n"},
						Right: &ast.CallExpr{
							Fun: &ast.Ident{Name: "factorial"},
							Args: []ast.Expr{
								&ast.BinaryExpr{
									Op:    ast.Sub,
									Left:  &ast.Ident{Name: "n"},
									Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
								},
							},
						},
					},
				},
			},
		},
	}

	fnDecl := &ast.FunctionDecl{
		Name: "factorial",
		Parameters: []ast.Parameter{
			{Name: "n", Type: &ast.PrimitiveType{Name: "int"}},
		},
		Body: factorialBody,
	}

	fnDef := &FunctionDef{
		Decl:   fnDecl,
		Params: fnDecl.Parameters,
		Body:   fnDecl.Body,
	}

	env.DefineFunction("factorial", fnDef)
	eval := NewEvaluator(env)

	// Test factorial(0) = 1
	callExpr0 := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "factorial"},
		Args: []ast.Expr{&ast.BasicLit{Kind: ast.IntLit, Value: "0"}},
	}

	result, err := eval.Eval(callExpr0)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}

	expected := state.NewIntValue(1)
	if !valuesEqual(result, expected) {
		t.Errorf("factorial(0): expected %s, got %s", expected.String(), result.String())
	}

	// Test factorial(1) = 1
	callExpr1 := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "factorial"},
		Args: []ast.Expr{&ast.BasicLit{Kind: ast.IntLit, Value: "1"}},
	}

	result, err = eval.Eval(callExpr1)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}

	if !valuesEqual(result, expected) {
		t.Errorf("factorial(1): expected %s, got %s", expected.String(), result.String())
	}

	// Test factorial(5) = 120
	callExpr5 := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "factorial"},
		Args: []ast.Expr{&ast.BasicLit{Kind: ast.IntLit, Value: "5"}},
	}

	result, err = eval.Eval(callExpr5)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}

	expected = state.NewIntValue(120)
	if !valuesEqual(result, expected) {
		t.Errorf("factorial(5): expected %s, got %s", expected.String(), result.String())
	}
}

func TestEvaluatorMutualRecursion(t *testing.T) {
	env := NewEnvironment()

	// Define isEven and isOdd functions that call each other
	// isEven(n) = if (n = 0) then true else isOdd(n - 1)
	// isOdd(n) = if (n = 0) then false else isEven(n - 1)

	isEvenBody := &ast.BlockStmt{
		Statements: []ast.Stmt{
			&ast.ReturnStmt{
				Value: &ast.IfExpr{
					Condition: &ast.BinaryExpr{
						Op:    ast.Eq,
						Left:  &ast.Ident{Name: "n"},
						Right: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
					},
					Then: &ast.BasicLit{Kind: ast.BoolLit, Value: "true"},
					Else: &ast.CallExpr{
						Fun: &ast.Ident{Name: "isOdd"},
						Args: []ast.Expr{
							&ast.BinaryExpr{
								Op:    ast.Sub,
								Left:  &ast.Ident{Name: "n"},
								Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
							},
						},
					},
				},
			},
		},
	}

	isOddBody := &ast.BlockStmt{
		Statements: []ast.Stmt{
			&ast.ReturnStmt{
				Value: &ast.IfExpr{
					Condition: &ast.BinaryExpr{
						Op:    ast.Eq,
						Left:  &ast.Ident{Name: "n"},
						Right: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
					},
					Then: &ast.BasicLit{Kind: ast.BoolLit, Value: "false"},
					Else: &ast.CallExpr{
						Fun: &ast.Ident{Name: "isEven"},
						Args: []ast.Expr{
							&ast.BinaryExpr{
								Op:    ast.Sub,
								Left:  &ast.Ident{Name: "n"},
								Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
							},
						},
					},
				},
			},
		},
	}

	isEvenDecl := &ast.FunctionDecl{
		Name: "isEven",
		Parameters: []ast.Parameter{
			{Name: "n", Type: &ast.PrimitiveType{Name: "int"}},
		},
		Body: isEvenBody,
	}

	isOddDecl := &ast.FunctionDecl{
		Name: "isOdd",
		Parameters: []ast.Parameter{
			{Name: "n", Type: &ast.PrimitiveType{Name: "int"}},
		},
		Body: isOddBody,
	}

	env.DefineFunction("isEven", &FunctionDef{Decl: isEvenDecl, Params: isEvenDecl.Parameters, Body: isEvenDecl.Body})
	env.DefineFunction("isOdd", &FunctionDef{Decl: isOddDecl, Params: isOddDecl.Parameters, Body: isOddDecl.Body})

	eval := NewEvaluator(env)

	// Test isEven(4) = true
	callExpr := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "isEven"},
		Args: []ast.Expr{&ast.BasicLit{Kind: ast.IntLit, Value: "4"}},
	}

	result, err := eval.Eval(callExpr)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}

	expected := state.NewBoolValue(true)
	if !valuesEqual(result, expected) {
		t.Errorf("isEven(4): expected %s, got %s", expected.String(), result.String())
	}

	// Test isEven(5) = false
	callExpr2 := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "isEven"},
		Args: []ast.Expr{&ast.BasicLit{Kind: ast.IntLit, Value: "5"}},
	}

	result, err = eval.Eval(callExpr2)
	if err != nil {
		t.Fatalf("evaluation error: %v", err)
	}

	expected = state.NewBoolValue(false)
	if !valuesEqual(result, expected) {
		t.Errorf("isEven(5): expected %s, got %s", expected.String(), result.String())
	}
}

