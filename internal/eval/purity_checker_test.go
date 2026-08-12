package eval

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestPurityCheckerValidFunction(t *testing.T) {
	// Create a variable model with a state variable
	vm := state.NewVariableModel(&ast.File{
		Decls: []ast.Decl{
			&ast.VariableDecl{
				Name: "counter",
				Type: &ast.PrimitiveType{Name: "int"},
			},
		},
	})

	checker := NewPurityChecker(vm)

	// Create a pure function that doesn't access state variables
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

	err := checker.CheckFunction(fnDecl)
	if err != nil {
		t.Errorf("expected no purity violations, got: %v", err)
	}
}

func TestPurityCheckerStateVariableAccess(t *testing.T) {
	// Create a variable model with a state variable
	vm := state.NewVariableModel(&ast.File{
		Decls: []ast.Decl{
			&ast.VariableDecl{
				Name: "counter",
				Type: &ast.PrimitiveType{Name: "int"},
			},
		},
	})

	checker := NewPurityChecker(vm)

	// Create a function that accesses a state variable (should fail)
	fnDecl := &ast.FunctionDecl{
		Name: "getCounter",
		Parameters: []ast.Parameter{},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.Ident{Name: "counter"}, // Accessing state variable
				},
			},
		},
	}

	err := checker.CheckFunction(fnDecl)
	if err == nil {
		t.Error("expected purity violation for accessing state variable 'counter'")
	}
}

func TestPurityCheckerNestedStateVariableAccess(t *testing.T) {
	vm := state.NewVariableModel(&ast.File{
		Decls: []ast.Decl{
			&ast.VariableDecl{
				Name: "counter",
				Type: &ast.PrimitiveType{Name: "int"},
			},
		},
	})

	checker := NewPurityChecker(vm)

	// Create a function with nested expression accessing state variable
	fnDecl := &ast.FunctionDecl{
		Name: "addToCounter",
		Parameters: []ast.Parameter{
			{Name: "x", Type: &ast.PrimitiveType{Name: "int"}},
		},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.BinaryExpr{
						Op:    ast.Add,
						Left:  &ast.Ident{Name: "counter"}, // State variable access
						Right: &ast.Ident{Name: "x"},
					},
				},
			},
		},
	}

	err := checker.CheckFunction(fnDecl)
	if err == nil {
		t.Error("expected purity violation for accessing state variable 'counter'")
	}
}

func TestPurityCheckerFunctionCallWithStateVariable(t *testing.T) {
	vm := state.NewVariableModel(&ast.File{
		Decls: []ast.Decl{
			&ast.VariableDecl{
				Name: "counter",
				Type: &ast.PrimitiveType{Name: "int"},
			},
		},
	})

	checker := NewPurityChecker(vm)

	// Create a function that calls another function with a state variable as argument
	fnDecl := &ast.FunctionDecl{
		Name: "test",
		Parameters: []ast.Parameter{},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.CallExpr{
						Fun: &ast.Ident{Name: "add"},
						Args: []ast.Expr{
							&ast.Ident{Name: "counter"}, // State variable as argument
							&ast.BasicLit{Kind: ast.IntLit, Value: "1"},
						},
					},
				},
			},
		},
	}

	err := checker.CheckFunction(fnDecl)
	if err == nil {
		t.Error("expected purity violation for passing state variable as argument")
	}
}

func TestPurityCheckerLetExpressionWithStateVariable(t *testing.T) {
	vm := state.NewVariableModel(&ast.File{
		Decls: []ast.Decl{
			&ast.VariableDecl{
				Name: "counter",
				Type: &ast.PrimitiveType{Name: "int"},
			},
		},
	})

	checker := NewPurityChecker(vm)

	// Create a function with let expression accessing state variable
	fnDecl := &ast.FunctionDecl{
		Name: "test",
		Parameters: []ast.Parameter{},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.LetExpr{
						Name:  "x",
						Value: &ast.Ident{Name: "counter"}, // State variable in let binding
						Body:  &ast.Ident{Name: "x"},
					},
				},
			},
		},
	}

	err := checker.CheckFunction(fnDecl)
	if err == nil {
		t.Error("expected purity violation for accessing state variable in let expression")
	}
}

func TestPurityCheckerAssignmentStatement(t *testing.T) {
	vm := state.NewVariableModel(&ast.File{
		Decls: []ast.Decl{
			&ast.VariableDecl{
				Name: "counter",
				Type: &ast.PrimitiveType{Name: "int"},
			},
		},
	})

	checker := NewPurityChecker(vm)

	// Create an assignment statement manually (assignment statements are only in actions)
	assignStmt := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "counter"},
		Right: &ast.BasicLit{Kind: ast.IntLit, Value: "10"},
	}

	fnDecl := &ast.FunctionDecl{
		Name:       "test",
		Parameters: []ast.Parameter{},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				assignStmt,
				&ast.ReturnStmt{
					Value: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
				},
			},
		},
	}

	err := checker.CheckFunction(fnDecl)
	if err == nil {
		t.Error("expected purity violation for assignment statement in pure function")
	}
}

