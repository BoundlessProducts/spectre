package types

import (
	"testing"

	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestRecordFieldAccess(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
			"name": &Primitive{Kind: Str},
		},
	}
	env.DeclareVariable("point", recordType)

	tests := []struct {
		name     string
		field    string
		expected string
		hasError bool
	}{
		{"access x field", "x", "int", false},
		{"access y field", "y", "int", false},
		{"access name field", "name", "str", false},
		{"access non-existent field", "z", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewChecker(env)
			sel := &ast.SelectorExpr{
				X:   &ast.Ident{Name: "point"},
				Sel: tt.field,
			}

			typ := c.CheckExpression(sel)
			if tt.hasError {
				if typ != nil || len(c.Errors()) == 0 {
					t.Error("expected type error")
				}
			} else {
				if typ == nil {
					t.Fatal("expected type, got nil")
				}
				if typ.String() != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, typ.String())
				}
			}
		})
	}
}

func TestRecordFieldUpdate(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	env.DeclareVariable("point", recordType)

	checker := NewChecker(env)

	// Valid: point.x = 10
	assign := &ast.AssignStmt{
		Left: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "point"},
			Sel: "x",
		},
		Right: &ast.BasicLit{Kind: ast.IntLit, Value: "10"},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment to record field")
	}

	// Invalid: point.x = "string"
	checker2 := NewChecker(env)
	assign2 := &ast.AssignStmt{
		Left: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "point"},
			Sel: "x",
		},
		Right: &ast.BasicLit{Kind: ast.StringLit},
	}

	if checker2.CheckAssignment(assign2) {
		t.Error("expected invalid assignment (type mismatch)")
	}
}

func TestNestedRecordAccess(t *testing.T) {
	env := NewEnvironment()
	
	// Create nested record: { position: { x: int, y: int }, name: str }
	innerRecord := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	outerRecord := &Record{
		Fields: map[string]Type{
			"position": innerRecord,
			"name": &Primitive{Kind: Str},
		},
	}
	env.DeclareVariable("entity", outerRecord)

	checker := NewChecker(env)

	// Access nested field: entity.position.x
	nestedSel := &ast.SelectorExpr{
		X: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "entity"},
			Sel: "position",
		},
		Sel: "x",
	}

	typ := checker.CheckExpression(nestedSel)
	if typ == nil {
		t.Fatal("expected type for nested field access")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestNestedRecordUpdate(t *testing.T) {
	env := NewEnvironment()
	
	innerRecord := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	outerRecord := &Record{
		Fields: map[string]Type{
			"position": innerRecord,
		},
	}
	env.DeclareVariable("entity", outerRecord)

	checker := NewChecker(env)

	// Update nested field: entity.position.x = 10
	assign := &ast.AssignStmt{
		Left: &ast.SelectorExpr{
			X: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "entity"},
				Sel: "position",
			},
			Sel: "x",
		},
		Right: &ast.BasicLit{Kind: ast.IntLit, Value: "10"},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment to nested record field")
	}
}

func TestRecordWithComplexFields(t *testing.T) {
	env := NewEnvironment()
	
	// Record with various field types
	recordType := &Record{
		Fields: map[string]Type{
			"id": &Primitive{Kind: Int},
			"name": &Primitive{Kind: Str},
			"active": &Primitive{Kind: Bool},
			"scores": &List{Element: &Primitive{Kind: Int}},
			"metadata": &Map{
				Key:   &Primitive{Kind: Str},
				Value: &Primitive{Kind: Str},
			},
		},
	}
	env.DeclareVariable("user", recordType)

	checker := NewChecker(env)

	// Access list field
	listSel := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "user"},
		Sel: "scores",
	}
	typ := checker.CheckExpression(listSel)
	if typ == nil {
		t.Fatal("expected type for list field")
	}
	if typ.String() != "List<int>" {
		t.Errorf("expected List<int>, got %s", typ.String())
	}

	// Access map field
	mapSel := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "user"},
		Sel: "metadata",
	}
	typ = checker.CheckExpression(mapSel)
	if typ == nil {
		t.Fatal("expected type for map field")
	}
	if typ.String() != "Map<str, str>" {
		t.Errorf("expected Map<str, str>, got %s", typ.String())
	}
}

func TestRecordFieldInExpression(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	env.DeclareVariable("p1", recordType)
	env.DeclareVariable("p2", recordType)

	checker := NewChecker(env)

	// Expression: p1.x + p2.y
	expr := &ast.BinaryExpr{
		Op: ast.Add,
		Left: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "p1"},
			Sel: "x",
		},
		Right: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "p2"},
			Sel: "y",
		},
	}

	typ := checker.CheckExpression(expr)
	if typ == nil {
		t.Fatal("expected type for expression with record fields")
	}
	if typ.String() != "int" {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestRecordFieldComparison(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
		},
	}
	env.DeclareVariable("p1", recordType)
	env.DeclareVariable("p2", recordType)

	checker := NewChecker(env)

	// Comparison: p1.x == p2.x
	expr := &ast.BinaryExpr{
		Op: ast.Eq,
		Left: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "p1"},
			Sel: "x",
		},
		Right: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "p2"},
			Sel: "x",
		},
	}

	typ := checker.CheckExpression(expr)
	if typ == nil {
		t.Fatal("expected type for comparison expression")
	}
	if typ.String() != "bool" {
		t.Errorf("expected bool, got %s", typ.String())
	}
}

func TestRecordFieldAssignmentWithExpression(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	env.DeclareVariable("point", recordType)
	env.DeclareVariable("offset", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Assignment: point.x = point.x + offset
	assign := &ast.AssignStmt{
		Left: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "point"},
			Sel: "x",
		},
		Right: &ast.BinaryExpr{
			Op: ast.Add,
			Left: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "point"},
				Sel: "x",
			},
			Right: &ast.Ident{Name: "offset"},
		},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment with expression")
	}
	if len(checker.Errors()) != 0 {
		t.Errorf("unexpected errors: %v", checker.Errors())
	}
}

func TestRecordFieldAccessOnNonRecord(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Try to access field on primitive type
	sel := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "x"},
		Sel: "field",
	}

	typ := checker.CheckExpression(sel)
	if typ != nil {
		t.Error("expected nil for field access on non-record")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for field access on non-record")
	}
}

