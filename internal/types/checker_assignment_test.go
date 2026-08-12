package types

import (
	"testing"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func TestCheckAssignment(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("y", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Valid assignment: int = int
	assign := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "x"},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment")
	}
	if len(checker.Errors()) != 0 {
		t.Errorf("unexpected errors: %v", checker.Errors())
	}
}

func TestCheckAssignmentTypeMismatch(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Invalid assignment: int = bool
	assign := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "x"},
		Right: &ast.BasicLit{Kind: ast.BoolLit},
	}

	if checker.CheckAssignment(assign) {
		t.Error("expected invalid assignment")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected type error")
	}
}

func TestCheckAssignmentToConstant(t *testing.T) {
	env := NewEnvironment()
	env.DeclareConstant("MAX", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Try to assign to constant
	assign := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "MAX"},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	if checker.CheckAssignment(assign) {
		t.Error("expected invalid assignment to constant")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for assigning to constant")
	}
}

func TestCheckAssignmentUndefinedVariable(t *testing.T) {
	env := NewEnvironment()
	checker := NewChecker(env)

	// Try to assign to undefined variable
	assign := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "undefined"},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	if checker.CheckAssignment(assign) {
		t.Error("expected invalid assignment to undefined variable")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected error for undefined variable")
	}
}

func TestCheckAssignmentNumericConversion(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Float})

	checker := NewChecker(env)

	// Valid: float = int (int can be assigned to float)
	assign := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "x"},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment (int to float)")
	}
}

func TestCheckAssignmentToRecordField(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
			"y": &Primitive{Kind: Int},
		},
	}
	env.DeclareVariable("point", recordType)

	checker := NewChecker(env)

	// Valid: point.x = int
	assign := &ast.AssignStmt{
		Left: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "point"},
			Sel: "x",
		},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment to record field")
	}
	if len(checker.Errors()) != 0 {
		t.Errorf("unexpected errors: %v", checker.Errors())
	}
}

func TestCheckAssignmentToRecordFieldTypeMismatch(t *testing.T) {
	env := NewEnvironment()
	recordType := &Record{
		Fields: map[string]Type{
			"x": &Primitive{Kind: Int},
		},
	}
	env.DeclareVariable("point", recordType)

	checker := NewChecker(env)

	// Invalid: point.x = bool
	assign := &ast.AssignStmt{
		Left: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "point"},
			Sel: "x",
		},
		Right: &ast.BasicLit{Kind: ast.BoolLit},
	}

	if checker.CheckAssignment(assign) {
		t.Error("expected invalid assignment to record field")
	}
	if len(checker.Errors()) == 0 {
		t.Error("expected type error")
	}
}

func TestCheckAssignmentToListElement(t *testing.T) {
	env := NewEnvironment()
	listType := &List{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("list", listType)

	checker := NewChecker(env)

	// Valid: list[0] = int
	assign := &ast.AssignStmt{
		Left: &ast.IndexExpr{
			X:     &ast.Ident{Name: "list"},
			Index: &ast.BasicLit{Kind: ast.IntLit},
		},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment to list element")
	}
	if len(checker.Errors()) != 0 {
		t.Errorf("unexpected errors: %v", checker.Errors())
	}
}

func TestCheckAssignmentToMapValue(t *testing.T) {
	env := NewEnvironment()
	mapType := &Map{
		Key:   &Primitive{Kind: Str},
		Value: &Primitive{Kind: Int},
	}
	env.DeclareVariable("map", mapType)

	checker := NewChecker(env)

	// Valid: map["key"] = int
	assign := &ast.AssignStmt{
		Left: &ast.IndexExpr{
			X:     &ast.Ident{Name: "map"},
			Index: &ast.BasicLit{Kind: ast.StringLit},
		},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment to map value")
	}
	if len(checker.Errors()) != 0 {
		t.Errorf("unexpected errors: %v", checker.Errors())
	}
}

func TestCheckAssignmentToOption(t *testing.T) {
	env := NewEnvironment()
	optType := &Option{Element: &Primitive{Kind: Int}}
	env.DeclareVariable("opt", optType)

	checker := NewChecker(env)

	// Valid: opt = int (int can be assigned to Option<int>)
	assign := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "opt"},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment (int to Option<int>)")
	}
	if len(checker.Errors()) != 0 {
		t.Errorf("unexpected errors: %v", checker.Errors())
	}
}

func TestCheckAssignmentPrimeNotation(t *testing.T) {
	env := NewEnvironment()
	env.DeclareVariable("counter", &Primitive{Kind: Int})

	checker := NewChecker(env)

	// Prime notation: counter' = expression
	// Primed variables refer to the next state of a variable
	// They should have the same type as the unprimed variable
	assign := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "counter", Prime: true},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	// This should succeed because "counter" exists and counter' has the same type
	if !checker.CheckAssignment(assign) {
		t.Error("expected valid assignment (primed variable references existing variable)")
	}
	if len(checker.Errors()) != 0 {
		t.Errorf("unexpected errors: %v", checker.Errors())
	}

	// Test with undefined base variable
	checker2 := NewChecker(env)
	assign2 := &ast.AssignStmt{
		Left:  &ast.Ident{Name: "undefined", Prime: true},
		Right: &ast.BasicLit{Kind: ast.IntLit},
	}

	if checker2.CheckAssignment(assign2) {
		t.Error("expected invalid assignment (primed variable with undefined base)")
	}
	if len(checker2.Errors()) == 0 {
		t.Error("expected error for undefined primed variable")
	}
}

