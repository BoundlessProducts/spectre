package ast

import "testing"

func TestASTNodeTypes(t *testing.T) {
	// Test that all node types implement their interfaces correctly
	
	// Test VariableDecl
	varDecl := &VariableDecl{
		Position: Position{Line: 1, Column: 1, Offset: 0},
		Name:     "counter",
		Type:     &PrimitiveType{Name: "int"},
	}
	if varDecl.Pos().Line != 1 {
		t.Error("VariableDecl.Pos() failed")
	}
	
	// Test ConstantDecl
	constDecl := &ConstantDecl{
		Position: Position{Line: 2, Column: 1, Offset: 10},
		Name:     "MAX",
		Type:     &PrimitiveType{Name: "int"},
		Value:    &BasicLit{Kind: IntLit, Value: "100"},
	}
	if constDecl.Pos().Line != 2 {
		t.Error("ConstantDecl.Pos() failed")
	}
	
	// Test PrimitiveType
	intType := &PrimitiveType{Name: "int"}
	if intType.Name != "int" {
		t.Error("PrimitiveType.Name failed")
	}
	
	// Test BasicLit
	intLit := &BasicLit{Kind: IntLit, Value: "42"}
	if intLit.Kind != IntLit {
		t.Error("BasicLit.Kind failed")
	}
	
	// Test BinaryExpr
	binExpr := &BinaryExpr{
		Op:    Add,
		Left:  &Ident{Name: "x"},
		Right: &Ident{Name: "y"},
	}
	if binExpr.Op != Add {
		t.Error("BinaryExpr.Op failed")
	}
}

func TestASTInterfaces(t *testing.T) {
	// Verify that all types implement their interfaces
	
	var _ Decl = (*VariableDecl)(nil)
	var _ Decl = (*ConstantDecl)(nil)
	
	var _ Type = (*PrimitiveType)(nil)
	var _ Type = (*RecordType)(nil)
	var _ Type = (*SetType)(nil)
	var _ Type = (*MapType)(nil)
	var _ Type = (*ListType)(nil)
	var _ Type = (*EnumType)(nil)
	var _ Type = (*OptionType)(nil)
	var _ Type = (*NamedType)(nil)
	
	var _ Expr = (*Ident)(nil)
	var _ Expr = (*BasicLit)(nil)
	var _ Expr = (*BinaryExpr)(nil)
	var _ Expr = (*UnaryExpr)(nil)
	var _ Expr = (*CallExpr)(nil)
	var _ Expr = (*SelectorExpr)(nil)
	var _ Expr = (*IndexExpr)(nil)
	var _ Expr = (*ParenExpr)(nil)
}

