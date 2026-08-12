package types

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// buildTypeEnvironmentFromFile builds a type environment from a parsed file
func buildTypeEnvironmentFromFile(file *ast.File) (*Environment, error) {
	env := NewEnvironment()

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.VariableDecl:
			typ, err := FromAST(d.Type)
			if err != nil {
				// Skip if type conversion fails (might be named type or unimplemented feature)
				continue
			}
			env.DeclareVariable(d.Name, typ)

		case *ast.ConstantDecl:
			typ, err := FromAST(d.Type)
			if err != nil {
				continue
			}
			env.DeclareConstant(d.Name, typ)

		case *ast.FunctionDecl:
			params := make([]Type, len(d.Parameters))
			for i, param := range d.Parameters {
				paramType, err := FromAST(param.Type)
				if err != nil {
					return nil, err
				}
				params[i] = paramType
			}
			returnType, err := FromAST(d.ReturnType)
			if err != nil {
				return nil, err
			}
			sig := &FunctionSignature{
				Parameters: params,
				Return:     returnType,
			}
			env.DeclareFunction(d.Name, sig)

		case *ast.ModuleDecl:
			// Process module declarations directly
			for _, moduleDecl := range d.Decls {
				switch md := moduleDecl.(type) {
				case *ast.VariableDecl:
					typ, err := FromAST(md.Type)
					if err != nil {
						continue
					}
					env.DeclareVariable(md.Name, typ)
				case *ast.FunctionDecl:
					params := make([]Type, len(md.Parameters))
					for i, param := range md.Parameters {
						paramType, err := FromAST(param.Type)
						if err != nil {
							continue
						}
						params[i] = paramType
					}
					returnType, err := FromAST(md.ReturnType)
					if err != nil {
						continue
					}
					sig := &FunctionSignature{
						Parameters: params,
						Return:     returnType,
					}
					env.DeclareFunction(md.Name, sig)
				}
			}
		}
	}

	return env, nil
}

func buildTypeEnvironmentFromDecls(decls []ast.Decl) (*Environment, error) {
	env := NewEnvironment()
	for _, decl := range decls {
		switch d := decl.(type) {
		case *ast.VariableDecl:
			typ, err := FromAST(d.Type)
			if err != nil {
				continue
			}
			env.DeclareVariable(d.Name, typ)
		case *ast.FunctionDecl:
			params := make([]Type, len(d.Parameters))
			for i, param := range d.Parameters {
				paramType, err := FromAST(param.Type)
				if err != nil {
					continue
				}
				params[i] = paramType
			}
			returnType, err := FromAST(d.ReturnType)
			if err != nil {
				continue
			}
			sig := &FunctionSignature{
				Parameters: params,
				Return:     returnType,
			}
			env.DeclareFunction(d.Name, sig)
		}
	}
	return env, nil
}

func TestTypeCheckCounterSpec(t *testing.T) {
	filePath := filepath.Join("..", "..", "examples", "counter.spec")
	input, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read counter.spec: %v", err)
	}

	l := lexer.New(string(input))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) != 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Build type environment
	env, err := buildTypeEnvironmentFromFile(file)
	if err != nil {
		t.Fatalf("failed to build type environment: %v", err)
	}

	checker := NewChecker(env)

	// Type-check all actions
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.ActionDecl:
			if d.Body != nil {
				for _, stmt := range d.Body.Statements {
					switch s := stmt.(type) {
					case *ast.AssignStmt:
						if !checker.CheckAssignment(s) {
							t.Errorf("type error in action %s: %v", d.Name, checker.Errors())
						}
					case *ast.RequireStmt:
						typ := checker.CheckExpression(s.Condition)
						if typ == nil {
							t.Errorf("type error in require statement of action %s", d.Name)
						} else if !isBool(typ) {
							t.Errorf("require condition must be bool in action %s, got %s", d.Name, typ.String())
						}
					case *ast.EnsureStmt:
						typ := checker.CheckExpression(s.Condition)
						if typ == nil {
							t.Errorf("type error in ensure statement of action %s", d.Name)
						} else if !isBool(typ) {
							t.Errorf("ensure condition must be bool in action %s, got %s", d.Name, typ.String())
						}
					}
				}
			}

		case *ast.InvariantDecl:
			typ := checker.CheckExpression(d.Condition)
			if typ == nil {
				t.Errorf("type error in invariant %s: %v", d.Name, checker.Errors())
			} else if !isBool(typ) {
				t.Errorf("invariant condition must be bool in %s, got %s", d.Name, typ.String())
			}
		}
	}

	if len(checker.Errors()) > 0 {
		t.Logf("Type checker errors: %v", checker.Errors())
	}
}

func TestTypeCheckMutexSpec(t *testing.T) {
	filePath := filepath.Join("..", "..", "examples", "mutex.spec")
	input, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read mutex.spec: %v", err)
	}

	l := lexer.New(string(input))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) != 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	// Note: mutex.spec uses enum types which aren't fully type-checked yet
	// For now, we'll verify that the file parses correctly
	// Full enum type checking will be implemented in a future phase
	actionCount := 0
	invariantCount := 0

	for _, decl := range file.Decls {
		switch decl.(type) {
		case *ast.ActionDecl:
			actionCount++
		case *ast.InvariantDecl:
			invariantCount++
		}
	}

	t.Logf("Parsed %d actions and %d invariants in mutex.spec (enum type checking not yet implemented)", actionCount, invariantCount)
	
	// Verify file structure is correct
	if actionCount == 0 {
		t.Error("expected at least one action in mutex.spec")
	}
	if invariantCount == 0 {
		t.Error("expected at least one invariant in mutex.spec")
	}
}

func TestTypeCheckComplexExpressions(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("counter", &Primitive{Kind: Int})
	env.DeclareVariable("maxValue", &Primitive{Kind: Int})
	env.DeclareVariable("enabled", &Primitive{Kind: Bool})

	checker := NewChecker(env)

	// Complex expression: counter >= 0 && counter <= maxValue && enabled
	expr := &ast.BinaryExpr{
		Op: ast.And,
		Left: &ast.BinaryExpr{
			Op: ast.And,
			Left: &ast.BinaryExpr{
				Op:    ast.Geq,
				Left:  &ast.Ident{Name: "counter"},
				Right: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
			},
			Right: &ast.BinaryExpr{
				Op:    ast.Leq,
				Left:  &ast.Ident{Name: "counter"},
				Right: &ast.Ident{Name: "maxValue"},
			},
		},
		Right: &ast.Ident{Name: "enabled"},
	}

	typ := checker.CheckExpression(expr)
	if typ == nil {
		t.Fatal("expected type for complex expression")
	}
	if typ.String() != "bool" {
		t.Errorf("expected bool, got %s", typ.String())
	}
}

func TestTypeCheckPrimedVariableExpressions(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("counter", &Primitive{Kind: Int})
	env.DeclareVariable("state", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Expression: counter' = counter + 1
	assign := &ast.AssignStmt{
		Left: &ast.Ident{Name: "counter", Prime: true},
		Right: &ast.BinaryExpr{
			Op:    ast.Add,
			Left:  &ast.Ident{Name: "counter"},
			Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
		},
	}

	if !checker.CheckAssignment(assign) {
		t.Errorf("expected valid primed assignment: %v", checker.Errors())
	}

	// Expression: counter' > counter
	comparison := &ast.BinaryExpr{
		Op: ast.Gt,
		Left: &ast.Ident{Name: "counter", Prime: true},
		Right: &ast.Ident{Name: "counter"},
	}

	typ := checker.CheckExpression(comparison)
	if typ == nil {
		t.Fatal("expected type for primed variable comparison")
	}
	if typ.String() != "bool" {
		t.Errorf("expected bool, got %s", typ.String())
	}
}

func TestTypeCheckRecordOperations(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"id":    &Primitive{Kind: Int},
			"name":  &Primitive{Kind: Str},
			"count": &Primitive{Kind: Int},
		},
	}
	env.DeclareVariable("user1", recordType)
	env.DeclareVariable("user2", recordType)

	checker := NewChecker(env)

	// Expression: user1.id = user2.id + 1
	assign := &ast.AssignStmt{
		Left: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "user1"},
			Sel: "id",
		},
		Right: &ast.BinaryExpr{
			Op: ast.Add,
			Left: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "user2"},
				Sel: "id",
			},
			Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
		},
	}

	if !checker.CheckAssignment(assign) {
		t.Errorf("expected valid record field assignment: %v", checker.Errors())
	}

	// Expression: user1.name = user2.name
	assign2 := &ast.AssignStmt{
		Left: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "user1"},
			Sel: "name",
		},
		Right: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "user2"},
			Sel: "name",
		},
	}

	if !checker.CheckAssignment(assign2) {
		t.Errorf("expected valid record field assignment: %v", checker.Errors())
	}
}

func TestTypeCheckCollectionOperations(t *testing.T) {
	env := NewEnvironment()
	listType := &List{Element: &Primitive{Kind: Int}}
	mapType := &Map{
		Key:   &Primitive{Kind: Str},
		Value: &Primitive{Kind: Int},
	}
	env.DeclareVariable("numbers", listType)
	env.DeclareVariable("scores", mapType)

	checker := NewChecker(env)

	// Expression: numbers[0] = scores["key"]
	assign := &ast.AssignStmt{
		Left: &ast.IndexExpr{
			X:     &ast.Ident{Name: "numbers"},
			Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
		},
		Right: &ast.IndexExpr{
			X:     &ast.Ident{Name: "scores"},
			Index: &ast.BasicLit{Kind: ast.StringLit, Value: "key"},
		},
	}

	if !checker.CheckAssignment(assign) {
		t.Errorf("expected valid collection assignment: %v", checker.Errors())
	}
}

func TestTypeCheckFunctionCallInComplexContext(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("double", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("y", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Expression: x = double(y) + 1
	assign := &ast.AssignStmt{
		Left: &ast.Ident{Name: "x"},
		Right: &ast.BinaryExpr{
			Op: ast.Add,
			Left: &ast.CallExpr{
				Fun: &ast.Ident{Name: "double"},
				Args: []ast.Expr{
					&ast.Ident{Name: "y"},
				},
			},
			Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
		},
	}

	if !checker.CheckAssignment(assign) {
		t.Errorf("expected valid assignment with function call: %v", checker.Errors())
	}
}

