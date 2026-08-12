package types

import (
	"testing"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// buildEnvironmentFromAST builds a type environment from AST declarations
func buildEnvironmentFromAST(file *ast.File) (*Environment, error) {
	env := NewEnvironment()

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.VariableDecl:
			typ, err := FromAST(d.Type)
			if err != nil {
				return nil, err
			}
			if err := env.DeclareVariable(d.Name, typ); err != nil {
				return nil, err
			}

		case *ast.ConstantDecl:
			typ, err := FromAST(d.Type)
			if err != nil {
				return nil, err
			}
			if err := env.DeclareConstant(d.Name, typ); err != nil {
				return nil, err
			}

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
			if err := env.DeclareFunction(d.Name, sig); err != nil {
				return nil, err
			}
		}
	}

	return env, nil
}

func TestTypeCheckSimpleExpressions(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("y", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Add integers: x + y
	addExpr := &ast.BinaryExpr{
		Op:    ast.Add,
		Left:  &ast.Ident{Name: "x"},
		Right: &ast.Ident{Name: "y"},
	}
	typ := checker.CheckExpression(addExpr)
	if typ == nil {
		t.Fatal("expected type for x + y")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Compare integers: x < y
	compareExpr := &ast.BinaryExpr{
		Op:    ast.Lt,
		Left:  &ast.Ident{Name: "x"},
		Right: &ast.Ident{Name: "y"},
	}
	typ = checker.CheckExpression(compareExpr)
	if typ == nil {
		t.Fatal("expected type for x < y")
	}
	if typ.String() != "bool" {
		t.Errorf("expected bool, got %s", typ.String())
	}

	// Invalid operation: x + true
	checker2 := NewChecker(env)
	invalidExpr := &ast.BinaryExpr{
		Op:    ast.Add,
		Left:  &ast.Ident{Name: "x"},
		Right: &ast.BasicLit{Kind: ast.BoolLit},
	}
	typ = checker2.CheckExpression(invalidExpr)
	if typ != nil {
		t.Error("expected nil for invalid operation")
	}
	if len(checker2.Errors()) == 0 {
		t.Error("expected type error")
	}
}

func TestTypeCheckAssignmentsIntegration(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("counter", &Primitive{Kind: Int})
	env.DeclareVariable("flag", &Primitive{Kind: Bool})

	// Valid assignment: counter = 10
	checker := NewChecker(env)
	assign := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "counter"},
		Right: &ast.BasicLit{Kind: ast.IntLit, Value: "10"},
	}
	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment")
	}

	// Type mismatch: counter = true
	checker2 := NewChecker(env)
	assign2 := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "counter"},
		Right: &ast.BasicLit{Kind: ast.BoolLit},
	}
	if checker2.CheckAssignment(assign2) {
		t.Error("expected invalid assignment")
	}
	if len(checker2.Errors()) == 0 {
		t.Error("expected type error")
	}

	// Primed assignment: counter' = counter + 1
	checker3 := NewChecker(env)
	assign3 := &ast.AssignStmt{
		Left: &ast.Ident{Name: "counter", Prime: true},
		Right: &ast.BinaryExpr{
			Op:    ast.Add,
			Left:  &ast.Ident{Name: "counter"},
			Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
		},
	}
	if !checker3.CheckAssignment(assign3) {
		t.Error("expected valid primed assignment")
	}
}

func TestTypeCheckComplexTypes(t *testing.T) {
	env := NewEnvironment()

	// Declare a list variable
	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("numbers", listType)

	// Declare a map variable
	mapType := &Map{
		Key:   &Primitive{Kind: Str},
		Value: &Primitive{Kind: Int},
	}
	env.DeclareVariable("scores", mapType)

	// Declare a record variable
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	env.DeclareVariable("point", recordType)

	checker := NewChecker(env)

	// Test list indexing
	listIndex := &ast.IndexExpr{
		X:     &ast.Ident{Name: "numbers"},
		Index: &ast.BasicLit{Kind: ast.IntLit},
	}
	typ := checker.CheckExpression(listIndex)
	if typ == nil {
		t.Fatal("expected type for list index")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Test map indexing
	mapIndex := &ast.IndexExpr{
		X:     &ast.Ident{Name: "scores"},
		Index: &ast.BasicLit{Kind: ast.StringLit},
	}
	typ = checker.CheckExpression(mapIndex)
	if typ == nil {
		t.Fatal("expected type for map index")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Test record field access
	fieldSel := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "point"},
		Sel: "x",
	}
	typ = checker.CheckExpression(fieldSel)
	if typ == nil {
		t.Fatal("expected type for record field")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestTypeCheckFunctionCalls(t *testing.T) {
	env := NewEnvironment()
	env.DeclareFunction("add", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}, &Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	})

	env.DeclareFunction("isPositive", &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Bool},
	})

	checker := NewChecker(env)

	// Valid function call
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit},
			&ast.BasicLit{Kind: ast.IntLit},
		},
	}

	typ := checker.CheckExpression(call)
	if typ == nil {
		t.Fatal("expected type for function call")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Invalid: wrong argument count
	checker2 := NewChecker(env)
	call2 := &ast.CallExpr{
		Fun: &ast.Ident{Name: "add"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.IntLit},
		},
	}

	typ = checker2.CheckExpression(call2)
	if typ != nil {
		t.Error("expected nil for invalid function call")
	}
	if len(checker2.Errors()) == 0 {
		t.Error("expected error for wrong argument count")
	}

	// Invalid: wrong argument type
	checker3 := NewChecker(env)
	call3 := &ast.CallExpr{
		Fun: &ast.Ident{Name: "isPositive"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: ast.BoolLit},
		},
	}

	typ = checker3.CheckExpression(call3)
	if typ != nil {
		t.Error("expected nil for invalid function call")
	}
	if len(checker3.Errors()) == 0 {
		t.Error("expected error for wrong argument type")
	}
}

func TestTypeCheckNestedExpressions(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("y", &Primitive{Kind: Int})
	env.DeclareVariable("z", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Complex nested expression: (x + y) * z
	nested := &ast.BinaryExpr{
		Op: ast.Mul,
		Left: &ast.ParenExpr{
			X: &ast.BinaryExpr{
				Op:    ast.Add,
				Left:  &ast.Ident{Name: "x"},
				Right: &ast.Ident{Name: "y"},
			},
		},
		Right: &ast.Ident{Name: "z"},
	}

	typ := checker.CheckExpression(nested)
	if typ == nil {
		t.Fatal("expected type for nested expression")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestTypeCheckIfExpression(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Valid if-else expression
	ifExpr := &ast.IfExpr{
		Condition: &ast.BinaryExpr{
			Op:    ast.Gt,
			Left:  &ast.Ident{Name: "x"},
			Right: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
		},
		Then: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
		Else: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
	}

	typ := checker.CheckExpression(ifExpr)
	if typ == nil {
		t.Fatal("expected type for if expression")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Invalid: condition not boolean
	checker2 := NewChecker(env)
	ifExpr2 := &ast.IfExpr{
		Condition: &ast.Ident{Name: "x"}, // int, not bool
		Then:      &ast.BasicLit{Kind: ast.IntLit},
		Else:      &ast.BasicLit{Kind: ast.IntLit},
	}

	typ = checker2.CheckExpression(ifExpr2)
	if typ != nil {
		t.Error("expected nil for invalid if expression")
	}
	if len(checker2.Errors()) == 0 {
		t.Error("expected error for non-boolean condition")
	}
}

func TestTypeCheckLetExpression(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Valid let expression: let y = x + 1 in y * 2
	letExpr := &ast.LetExpr{
		Name:  "y",
		Value: &ast.BinaryExpr{
			Op:    ast.Add,
			Left:  &ast.Ident{Name: "x"},
			Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
		},
		Body: &ast.BinaryExpr{
			Op:    ast.Mul,
			Left:  &ast.Ident{Name: "y"},
			Right: &ast.BasicLit{Kind: ast.IntLit, Value: "2"},
		},
	}

	typ := checker.CheckExpression(letExpr)
	if typ == nil {
		t.Fatal("expected type for let expression")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

