package types

import (
	"testing"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestListTypeChecking(t *testing.T) {
	env := NewEnvironment()
	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("numbers", listType)

	checker := NewChecker(env)

	// Test list indexing: numbers[0]
	indexExpr := &ast.IndexExpr{
		X:     &ast.Ident{Name: "numbers"},
		Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
	}

	typ := checker.CheckExpression(indexExpr)
	if typ == nil {
		t.Fatal("expected type for list index")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Test invalid index type: numbers["0"]
	checker2 := NewChecker(env)
	invalidIndex := &ast.IndexExpr{
		X:     &ast.Ident{Name: "numbers"},
		Index: &ast.BasicLit{Kind: ast.StringLit, Value: "0"},
	}

	typ = checker2.CheckExpression(invalidIndex)
	if typ != nil {
		t.Error("expected nil for invalid list index type")
	}
	if len(checker2.Errors()) == 0 {
		t.Error("expected error for invalid index type")
	}
}

func TestListElementAssignment(t *testing.T) {
	env := NewEnvironment()
	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("numbers", listType)

	checker := NewChecker(env)

	// Valid: numbers[0] = 10
	assign := &ast.AssignStmt{
		Left: &ast.IndexExpr{
			X:     &ast.Ident{Name: "numbers"},
			Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
		},
		Right: &ast.BasicLit{Kind: ast.IntLit, Value: "10"},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment to list element")
	}

	// Invalid: numbers[0] = "string"
	checker2 := NewChecker(env)
	assign2 := &ast.AssignStmt{
		Left: &ast.IndexExpr{
			X:     &ast.Ident{Name: "numbers"},
			Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
		},
		Right: &ast.BasicLit{Kind: ast.StringLit},
	}

	if checker2.CheckAssignment(assign2) {
		t.Error("expected invalid assignment (type mismatch)")
	}
}

func TestMapTypeChecking(t *testing.T) {
	env := NewEnvironment()
	mapType := &Map{
		Key:   &Primitive{Kind: Str},
		Value: &Primitive{Kind: Int},
	}
	env.DeclareVariable("scores", mapType)

	checker := NewChecker(env)

	// Test map indexing: scores["key"]
	indexExpr := &ast.IndexExpr{
		X:     &ast.Ident{Name: "scores"},
		Index: &ast.BasicLit{Kind: ast.StringLit, Value: "key"},
	}

	typ := checker.CheckExpression(indexExpr)
	if typ == nil {
		t.Fatal("expected type for map index")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Test invalid key type: scores[0]
	checker2 := NewChecker(env)
	invalidKey := &ast.IndexExpr{
		X:     &ast.Ident{Name: "scores"},
		Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
	}

	typ = checker2.CheckExpression(invalidKey)
	if typ != nil {
		t.Error("expected nil for invalid map key type")
	}
	if len(checker2.Errors()) == 0 {
		t.Error("expected error for invalid key type")
	}
}

func TestMapValueAssignment(t *testing.T) {
	env := NewEnvironment()
	mapType := &Map{
		Key:   &Primitive{Kind: Str},
		Value: &Primitive{Kind: Int},
	}
	env.DeclareVariable("scores", mapType)

	checker := NewChecker(env)

	// Valid: scores["key"] = 100
	assign := &ast.AssignStmt{
		Left: &ast.IndexExpr{
			X:     &ast.Ident{Name: "scores"},
			Index: &ast.BasicLit{Kind: ast.StringLit, Value: "key"},
		},
		Right: &ast.BasicLit{Kind: ast.IntLit, Value: "100"},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment to map value")
	}

	// Invalid: scores["key"] = "string"
	checker2 := NewChecker(env)
	assign2 := &ast.AssignStmt{
		Left: &ast.IndexExpr{
			X:     &ast.Ident{Name: "scores"},
			Index: &ast.BasicLit{Kind: ast.StringLit, Value: "key"},
		},
		Right: &ast.BasicLit{Kind: ast.StringLit},
	}

	if checker2.CheckAssignment(assign2) {
		t.Error("expected invalid assignment (type mismatch)")
	}
}

func TestSetTypeChecking(t *testing.T) {
	env := NewEnvironment()
	setType := &Set{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("numbers", setType)

	checker := NewChecker(env)

	// Sets don't support indexing, but we can check the type
	// For now, we'll verify that accessing a set as an index expression fails
	indexExpr := &ast.IndexExpr{
		X:     &ast.Ident{Name: "numbers"},
		Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
	}

	typ := checker.CheckExpression(indexExpr)
	if typ != nil {
		t.Error("expected nil for set indexing (sets don't support indexing)")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for set indexing")
	}
}

func TestNestedCollectionTypes(t *testing.T) {
	env := NewEnvironment()
	
	// List of lists: List<List<int>>
	nestedListType := &List{
		Element: &List{Element: &Primitive{Kind: Int}},
	}
	env.DeclareVariable("matrix", nestedListType)

	checker := NewChecker(env)

	// Access first level: matrix[0]
	firstIndex := &ast.IndexExpr{
		X:     &ast.Ident{Name: "matrix"},
		Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
	}

	typ := checker.CheckExpression(firstIndex)
	if typ == nil {
		t.Fatal("expected type for first-level index")
	}
	if typ.String() != "List<int>" {
		t.Errorf("expected List<int>, got %s", typ.String())
	}

	// Nested access: matrix[0][1]
	nestedIndex := &ast.IndexExpr{
		X: firstIndex,
		Index: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
	}

	typ = checker.CheckExpression(nestedIndex)
	if typ == nil {
		t.Fatal("expected type for nested index")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestMapWithComplexValueType(t *testing.T) {
	env := NewEnvironment()
	
	// Map with list values: Map<str, List<int>>
	mapType := &Map{
		Key:   &Primitive{Kind: Str},
		Value: &List{Element: &Primitive{Kind: Int}},
	}
	env.DeclareVariable("groups", mapType)

	checker := NewChecker(env)

	// Access map value: groups["key"]
	indexExpr := &ast.IndexExpr{
		X:     &ast.Ident{Name: "groups"},
		Index: &ast.BasicLit{Kind: ast.StringLit, Value: "key"},
	}

	typ := checker.CheckExpression(indexExpr)
	if typ == nil {
		t.Fatal("expected type for map access")
	}
	if typ.String() != "List<int>" {
		t.Errorf("expected List<int>, got %s", typ.String())
	}

	// Nested access: groups["key"][0]
	nestedIndex := &ast.IndexExpr{
		X:     indexExpr,
		Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
	}

	typ = checker.CheckExpression(nestedIndex)
	if typ == nil {
		t.Fatal("expected type for nested access")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestCollectionInExpressions(t *testing.T) {
	env := NewEnvironment()
	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("list1", listType)
	env.DeclareVariable("list2", listType)

	checker := NewChecker(env)

	// Expression: list1[0] + list2[0]
	expr := &ast.BinaryExpr{
		Op: ast.Add,
		Left: &ast.IndexExpr{
			X:     &ast.Ident{Name: "list1"},
			Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
		},
		Right: &ast.IndexExpr{
			X:     &ast.Ident{Name: "list2"},
			Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
		},
	}

	typ := checker.CheckExpression(expr)
	if typ == nil {
		t.Fatal("expected type for expression with collection access")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestCollectionAssignmentWithExpression(t *testing.T) {
	env := NewEnvironment()
	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("numbers", listType)
	env.DeclareVariable("value", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Assignment: numbers[0] = numbers[0] + value
	assign := &ast.AssignStmt{
		Left: &ast.IndexExpr{
			X:     &ast.Ident{Name: "numbers"},
			Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
		},
		Right: &ast.BinaryExpr{
			Op: ast.Add,
			Left: &ast.IndexExpr{
				X:     &ast.Ident{Name: "numbers"},
				Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
			},
			Right: &ast.Ident{Name: "value"},
		},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment with expression")
	}
	if len(checker.Errors()) != 0 {
		t.Errorf("unexpected errors: %v", checker.Errors())
	}
}

func TestIndexOnNonCollectionType(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Try to index a primitive type
	indexExpr := &ast.IndexExpr{
		X:     &ast.Ident{Name: "x"},
		Index: &ast.BasicLit{Kind: ast.IntLit, Value: "0"},
	}

	typ := checker.CheckExpression(indexExpr)
	if typ != nil {
		t.Error("expected nil for indexing non-collection type")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for indexing non-collection type")
	}
}

