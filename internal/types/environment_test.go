package types

import (
	"testing"
)

func TestNewEnvironment(t *testing.T) {
	env := NewEnvironment()
	if env == nil {
		t.Fatal("NewEnvironment returned nil")
	}
	if env.parent != nil {
		t.Error("new environment should not have a parent")
	}
	if len(env.variables) != 0 {
		t.Error("new environment should have no variables")
	}
}

func TestDeclareVariable(t *testing.T) {
	env := NewEnvironment()
	intType := &Primitive{Kind: Int}

	err := env.DeclareVariable("x", intType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to declare again - should fail
	err = env.DeclareVariable("x", intType)
	if err == nil {
		t.Error("expected error when declaring variable twice")
	}
}

func TestLookupVariable(t *testing.T) {
	env := NewEnvironment()
	intType := &Primitive{Kind: Int}
	boolType := &Primitive{Kind: Bool}

	env.DeclareVariable("x", intType)
	env.DeclareVariable("y", boolType)

	typ, found := env.LookupVariable("x")
	if !found {
		t.Error("variable x should be found")
	}
	if !typ.Equals(intType) {
		t.Errorf("expected int, got %s", typ.String())
	}

	typ, found = env.LookupVariable("y")
	if !found {
		t.Error("variable y should be found")
	}
	if !typ.Equals(boolType) {
		t.Errorf("expected bool, got %s", typ.String())
	}

	_, found = env.LookupVariable("z")
	if found {
		t.Error("variable z should not be found")
	}
}

func TestNestedScopes(t *testing.T) {
	parent := NewEnvironment()
	intType := &Primitive{Kind: Int}
	boolType := &Primitive{Kind: Bool}

	parent.DeclareVariable("x", intType)

	child := NewChildEnvironment(parent)
	child.DeclareVariable("y", boolType)

	// Child should see parent's variable
	typ, found := child.LookupVariable("x")
	if !found {
		t.Error("child should see parent's variable x")
	}
	if !typ.Equals(intType) {
		t.Errorf("expected int, got %s", typ.String())
	}

	// Child should see its own variable
	typ, found = child.LookupVariable("y")
	if !found {
		t.Error("child should see its own variable y")
	}
	if !typ.Equals(boolType) {
		t.Errorf("expected bool, got %s", typ.String())
	}

	// Parent should not see child's variable
	_, found = parent.LookupVariable("y")
	if found {
		t.Error("parent should not see child's variable y")
	}
}

func TestShadowing(t *testing.T) {
	parent := NewEnvironment()
	intType := &Primitive{Kind: Int}
	boolType := &Primitive{Kind: Bool}

	parent.DeclareVariable("x", intType)

	child := NewChildEnvironment(parent)
	child.DeclareVariable("x", boolType) // Shadow parent's x

	// Child should see its own x (shadowed)
	typ, found := child.LookupVariable("x")
	if !found {
		t.Error("child should see its own variable x")
	}
	if !typ.Equals(boolType) {
		t.Errorf("expected bool (shadowed), got %s", typ.String())
	}

	// Parent should still see its own x
	typ, found = parent.LookupVariable("x")
	if !found {
		t.Error("parent should still see its own variable x")
	}
	if !typ.Equals(intType) {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestDeclareFunction(t *testing.T) {
	env := NewEnvironment()
	sig := &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Bool},
	}

	err := env.DeclareFunction("f", sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to declare again - should fail
	err = env.DeclareFunction("f", sig)
	if err == nil {
		t.Error("expected error when declaring function twice")
	}
}

func TestLookupFunction(t *testing.T) {
	env := NewEnvironment()
	sig := &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Bool},
	}

	env.DeclareFunction("f", sig)

	foundSig, found := env.LookupFunction("f")
	if !found {
		t.Error("function f should be found")
	}
	if foundSig.Return.String() != "bool" {
		t.Errorf("expected bool return type, got %s", foundSig.Return.String())
	}
	if len(foundSig.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(foundSig.Parameters))
	}

	_, found = env.LookupFunction("g")
	if found {
		t.Error("function g should not be found")
	}
}

func TestNestedFunctionScopes(t *testing.T) {
	parent := NewEnvironment()
	parentSig := &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Int}},
		Return:     &Primitive{Kind: Int},
	}
	parent.DeclareFunction("f", parentSig)

	child := NewChildEnvironment(parent)
	childSig := &FunctionSignature{
		Parameters: []Type{&Primitive{Kind: Bool}},
		Return:     &Primitive{Kind: Bool},
	}
	child.DeclareFunction("g", childSig)

	// Child should see parent's function
	sig, found := child.LookupFunction("f")
	if !found {
		t.Error("child should see parent's function f")
	}
	if !sig.Return.Equals(&Primitive{Kind: Int}) {
		t.Error("child should see parent's function signature")
	}

	// Child should see its own function
	sig, found = child.LookupFunction("g")
	if !found {
		t.Error("child should see its own function g")
	}
}

func TestDeclareConstant(t *testing.T) {
	env := NewEnvironment()
	intType := &Primitive{Kind: Int}

	err := env.DeclareConstant("MAX", intType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typ, found := env.LookupConstant("MAX")
	if !found {
		t.Error("constant MAX should be found")
	}
	if !typ.Equals(intType) {
		t.Errorf("expected int, got %s", typ.String())
	}
}

func TestSetVariable(t *testing.T) {
	env := NewEnvironment()
	intType := &Primitive{Kind: Int}
	boolType := &Primitive{Kind: Bool}

	env.SetVariable("x", intType)
	typ, found := env.LookupVariable("x")
	if !found || !typ.Equals(intType) {
		t.Error("SetVariable should set the variable")
	}

	// SetVariable allows overwriting
	env.SetVariable("x", boolType)
	typ, found = env.LookupVariable("x")
	if !found || !typ.Equals(boolType) {
		t.Error("SetVariable should allow overwriting")
	}
}

func TestHasVariable(t *testing.T) {
	parent := NewEnvironment()
	parent.DeclareVariable("x", &Primitive{Kind: Int})

	child := NewChildEnvironment(parent)
	child.DeclareVariable("y", &Primitive{Kind: Bool})

	// HasVariable only checks current scope
	if !child.HasVariable("y") {
		t.Error("child should have variable y")
	}
	if child.HasVariable("x") {
		t.Error("child should not have variable x (only in parent)")
	}
	if !parent.HasVariable("x") {
		t.Error("parent should have variable x")
	}
}

func TestIsRoot(t *testing.T) {
	root := NewEnvironment()
	if !root.IsRoot() {
		t.Error("root environment should be root")
	}

	child := NewChildEnvironment(root)
	if child.IsRoot() {
		t.Error("child environment should not be root")
	}
}

func TestGetParent(t *testing.T) {
	parent := NewEnvironment()
	child := NewChildEnvironment(parent)

	if child.GetParent() != parent {
		t.Error("GetParent should return the parent")
	}

	if parent.GetParent() != nil {
		t.Error("root environment should have no parent")
	}
}

