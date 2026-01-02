package semantic

import (
	"testing"

	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestNewSymbolTable(t *testing.T) {
	st := NewSymbolTable()
	if st == nil {
		t.Fatal("NewSymbolTable returned nil")
	}
	if st.GlobalScope == nil {
		t.Fatal("GlobalScope is nil")
	}
	if st.GlobalScope.Kind != ScopeGlobal {
		t.Error("GlobalScope should be ScopeGlobal")
	}
	if st.GlobalScope.Name != "global" {
		t.Error("GlobalScope name should be 'global'")
	}
}

func TestDefineSymbol(t *testing.T) {
	st := NewSymbolTable()
	
	varDecl := &ast.VariableDecl{
		Position: ast.Position{Line: 1, Column: 1},
		Name:     "x",
		Type:     &ast.PrimitiveType{Name: "int"},
	}

	err := st.DefineSymbol(st.GlobalScope, "x", SymbolVariable, varDecl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to define again - should fail
	err = st.DefineSymbol(st.GlobalScope, "x", SymbolVariable, varDecl)
	if err == nil {
		t.Error("expected error when defining duplicate symbol")
	}
}

func TestLookupSymbol(t *testing.T) {
	st := NewSymbolTable()
	
	varDecl := &ast.VariableDecl{
		Position: ast.Position{Line: 1, Column: 1},
		Name:     "x",
		Type:     &ast.PrimitiveType{Name: "int"},
	}

	st.DefineSymbol(st.GlobalScope, "x", SymbolVariable, varDecl)

	// Lookup in global scope
	symbol, found := st.LookupSymbol(st.GlobalScope, "x")
	if !found {
		t.Error("symbol x should be found")
	}
	if symbol == nil {
		t.Fatal("symbol should not be nil")
	}
	if symbol.Name != "x" {
		t.Errorf("expected name 'x', got %s", symbol.Name)
	}
	if symbol.Kind != SymbolVariable {
		t.Errorf("expected SymbolVariable, got %d", symbol.Kind)
	}

	// Lookup non-existent symbol
	_, found = st.LookupSymbol(st.GlobalScope, "y")
	if found {
		t.Error("symbol y should not be found")
	}
}

func TestNestedScopes(t *testing.T) {
	st := NewSymbolTable()
	
	// Define symbol in global scope
	varDecl1 := &ast.VariableDecl{
		Position: ast.Position{Line: 1, Column: 1},
		Name:     "x",
		Type:     &ast.PrimitiveType{Name: "int"},
	}
	st.DefineSymbol(st.GlobalScope, "x", SymbolVariable, varDecl1)

	// Create function scope
	funcScope := st.NewScope(st.GlobalScope, ScopeFunction, "testFunc")
	
	// Define symbol in function scope
	varDecl2 := &ast.VariableDecl{
		Position: ast.Position{Line: 2, Column: 1},
		Name:     "y",
		Type:     &ast.PrimitiveType{Name: "int"},
	}
	st.DefineSymbol(funcScope, "y", SymbolVariable, varDecl2)

	// Function scope should see its own symbol
	symbol, found := st.LookupSymbol(funcScope, "y")
	if !found {
		t.Error("symbol y should be found in function scope")
	}
	if symbol.Name != "y" {
		t.Errorf("expected name 'y', got %s", symbol.Name)
	}

	// Function scope should see parent's symbol
	symbol, found = st.LookupSymbol(funcScope, "x")
	if !found {
		t.Error("symbol x should be found in function scope (from parent)")
	}
	if symbol.Name != "x" {
		t.Errorf("expected name 'x', got %s", symbol.Name)
	}

	// Global scope should not see function's symbol
	_, found = st.LookupSymbol(st.GlobalScope, "y")
	if found {
		t.Error("symbol y should not be found in global scope")
	}
}

func TestShadowing(t *testing.T) {
	st := NewSymbolTable()
	
	// Define x in global scope
	varDecl1 := &ast.VariableDecl{
		Position: ast.Position{Line: 1, Column: 1},
		Name:     "x",
		Type:     &ast.PrimitiveType{Name: "int"},
	}
	st.DefineSymbol(st.GlobalScope, "x", SymbolVariable, varDecl1)

	// Create function scope and shadow x
	funcScope := st.NewScope(st.GlobalScope, ScopeFunction, "testFunc")
	varDecl2 := &ast.VariableDecl{
		Position: ast.Position{Line: 2, Column: 1},
		Name:     "x",
		Type:     &ast.PrimitiveType{Name: "bool"},
	}
	st.DefineSymbol(funcScope, "x", SymbolVariable, varDecl2)

	// Function scope should see its own x (shadowed)
	symbol, found := st.LookupSymbol(funcScope, "x")
	if !found {
		t.Error("symbol x should be found in function scope")
	}
	if symbol.Name != "x" {
		t.Errorf("expected name 'x', got %s", symbol.Name)
	}
	// Verify it's the shadowed one (bool type)
	if varDecl, ok := symbol.Decl.(*ast.VariableDecl); ok {
		if primType, ok := varDecl.Type.(*ast.PrimitiveType); ok {
			if primType.Name != "bool" {
				t.Errorf("expected shadowed x to be bool, got %s", primType.Name)
			}
		}
	}
}

func TestDifferentSymbolKinds(t *testing.T) {
	st := NewSymbolTable()
	
	// Define variable
	varDecl := &ast.VariableDecl{
		Position: ast.Position{Line: 1, Column: 1},
		Name:     "x",
		Type:     &ast.PrimitiveType{Name: "int"},
	}
	st.DefineSymbol(st.GlobalScope, "x", SymbolVariable, varDecl)

	// Define constant
	constDecl := &ast.ConstantDecl{
		Position: ast.Position{Line: 2, Column: 1},
		Name:     "MAX",
		Type:     &ast.PrimitiveType{Name: "int"},
		Value:    &ast.BasicLit{Kind: ast.IntLit, Value: "100"},
	}
	st.DefineSymbol(st.GlobalScope, "MAX", SymbolConstant, constDecl)

	// Define function
	funcDecl := &ast.FunctionDecl{
		Position: ast.Position{Line: 3, Column: 1},
		Name:     "add",
		ReturnType: &ast.PrimitiveType{Name: "int"},
	}
	st.DefineSymbol(st.GlobalScope, "add", SymbolFunction, funcDecl)

	// Verify all symbols exist
	symbols := []string{"x", "MAX", "add"}
	kinds := []SymbolKind{SymbolVariable, SymbolConstant, SymbolFunction}

	for i, name := range symbols {
		symbol, found := st.LookupSymbol(st.GlobalScope, name)
		if !found {
			t.Errorf("symbol %s should be found", name)
			continue
		}
		if symbol.Kind != kinds[i] {
			t.Errorf("symbol %s should have kind %d, got %d", name, kinds[i], symbol.Kind)
		}
	}
}

func TestLookupSymbolInScope(t *testing.T) {
	st := NewSymbolTable()
	
	// Define symbol in global scope
	varDecl := &ast.VariableDecl{
		Position: ast.Position{Line: 1, Column: 1},
		Name:     "x",
		Type:     &ast.PrimitiveType{Name: "int"},
	}
	st.DefineSymbol(st.GlobalScope, "x", SymbolVariable, varDecl)

	// Create function scope
	funcScope := st.NewScope(st.GlobalScope, ScopeFunction, "testFunc")

	// LookupInScope should not find x (only searches current scope)
	_, found := st.LookupSymbolInScope(funcScope, "x")
	if found {
		t.Error("LookupSymbolInScope should not find x in function scope")
	}

	// LookupSymbol should find x (searches parent scopes)
	_, found = st.LookupSymbol(funcScope, "x")
	if !found {
		t.Error("LookupSymbol should find x in function scope (from parent)")
	}
}

func TestNewScope(t *testing.T) {
	st := NewSymbolTable()
	
	funcScope := st.NewScope(st.GlobalScope, ScopeFunction, "testFunc")
	if funcScope == nil {
		t.Fatal("NewScope returned nil")
	}
	if funcScope.Parent != st.GlobalScope {
		t.Error("function scope parent should be global scope")
	}
	if funcScope.Kind != ScopeFunction {
		t.Error("function scope kind should be ScopeFunction")
	}
	if funcScope.Name != "testFunc" {
		t.Errorf("expected name 'testFunc', got %s", funcScope.Name)
	}

	if st.ScopeCount() != 2 {
		t.Errorf("expected 2 scopes, got %d", st.ScopeCount())
	}
}

