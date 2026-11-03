package types

import (
	"strings"
	"testing"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// helperChecker builds a Checker with one state variable "x: int" and
// one function parameter "p: int" in the environment.
func helperChecker() *Checker {
	env := NewEnvironment()
	env.DeclareVariable("x", &Primitive{Kind: Int})
	env.DeclareVariable("p", &Primitive{Kind: Int}) // function parameter, NOT a state var
	checker := NewChecker(env)
	checker.SetStateVariables(map[string]bool{"x": true})
	return checker
}

func primedIdent(name string) *ast.Ident {
	return &ast.Ident{Name: name, Prime: true}
}

func unprimedIdent(name string) *ast.Ident {
	return &ast.Ident{Name: name}
}

// ─── primed variable context checks ─────────────────────────────────────────

func TestContext_PrimedVarAcceptedInStateUpdate(t *testing.T) {
	c := helperChecker()
	typ := c.CheckExpressionInContext(primedIdent("x"), CtxStateUpdate)
	if typ == nil {
		t.Fatal("expected type for x' in CtxStateUpdate, got nil")
	}
	if len(c.Errors()) != 0 {
		t.Errorf("unexpected errors for x' in CtxStateUpdate: %v", c.Errors())
	}
}

func TestContext_PrimedVarRejectedInStateRead(t *testing.T) {
	c := helperChecker()
	c.CheckExpressionInContext(primedIdent("x"), CtxStateRead)
	if len(c.Errors()) == 0 {
		t.Error("expected error for x' in CtxStateRead (invariant context), got none")
	}
	if !strings.Contains(c.Errors()[0].Message, "primed variable") {
		t.Errorf("unexpected error message: %s", c.Errors()[0].Message)
	}
}

func TestContext_PrimedVarRejectedInPure(t *testing.T) {
	c := helperChecker()
	c.CheckExpressionInContext(primedIdent("x"), CtxPure)
	if len(c.Errors()) == 0 {
		t.Error("expected error for x' in CtxPure (function body), got none")
	}
	if !strings.Contains(c.Errors()[0].Message, "primed variable") {
		t.Errorf("unexpected error message: %s", c.Errors()[0].Message)
	}
}

func TestContext_PrimedVarAllowedInCtxUnknown(t *testing.T) {
	// CtxUnknown is the default (backward-compat): no extra checks.
	c := helperChecker()
	c.CheckExpressionInContext(primedIdent("x"), CtxUnknown)
	if len(c.Errors()) != 0 {
		t.Errorf("unexpected errors for x' in CtxUnknown: %v", c.Errors())
	}
}

// ─── state-variable read in pure context ─────────────────────────────────────

func TestContext_StateVarRejectedInPure(t *testing.T) {
	c := helperChecker()
	c.CheckExpressionInContext(unprimedIdent("x"), CtxPure)
	if len(c.Errors()) == 0 {
		t.Error("expected error for state var 'x' in CtxPure, got none")
	}
	if !strings.Contains(c.Errors()[0].Message, "state variable") {
		t.Errorf("unexpected error message: %s", c.Errors()[0].Message)
	}
}

func TestContext_StateVarAllowedInStateRead(t *testing.T) {
	c := helperChecker()
	typ := c.CheckExpressionInContext(unprimedIdent("x"), CtxStateRead)
	if typ == nil {
		t.Fatal("expected type for 'x' in CtxStateRead, got nil")
	}
	if len(c.Errors()) != 0 {
		t.Errorf("unexpected errors for 'x' in CtxStateRead: %v", c.Errors())
	}
}

func TestContext_StateVarAllowedInStateUpdate(t *testing.T) {
	c := helperChecker()
	typ := c.CheckExpressionInContext(unprimedIdent("x"), CtxStateUpdate)
	if typ == nil {
		t.Fatal("expected type for 'x' in CtxStateUpdate, got nil")
	}
	if len(c.Errors()) != 0 {
		t.Errorf("unexpected errors for 'x' in CtxStateUpdate: %v", c.Errors())
	}
}

// A function parameter is NOT a state variable: it must be readable in pure context.
func TestContext_FuncParamAllowedInPure(t *testing.T) {
	c := helperChecker()
	typ := c.CheckExpressionInContext(unprimedIdent("p"), CtxPure)
	if typ == nil {
		t.Fatal("expected type for param 'p' in CtxPure, got nil")
	}
	if len(c.Errors()) != 0 {
		t.Errorf("unexpected errors for param 'p' in CtxPure: %v", c.Errors())
	}
}

// ─── context propagates into sub-expressions ────────────────────────────────

// Primed var inside a binary expr in CtxStateRead should still be rejected.
func TestContext_PrimedVarInBinaryExprStateRead(t *testing.T) {
	c := helperChecker()
	// x' + x  — the primed part is invalid in CtxStateRead
	expr := &ast.BinaryExpr{
		Op:    ast.Add,
		Left:  primedIdent("x"),
		Right: unprimedIdent("x"),
	}
	c.CheckExpressionInContext(expr, CtxStateRead)
	if len(c.Errors()) == 0 {
		t.Error("expected error for x' in binary expr under CtxStateRead")
	}
}

// State var in binary expr in CtxPure should be rejected.
func TestContext_StateVarInBinaryExprPure(t *testing.T) {
	c := helperChecker()
	// x + 1  — reading state var 'x' is invalid in CtxPure
	expr := &ast.BinaryExpr{
		Op:    ast.Add,
		Left:  unprimedIdent("x"),
		Right: &ast.BasicLit{Kind: ast.IntLit, Value: "1"},
	}
	c.CheckExpressionInContext(expr, CtxPure)
	if len(c.Errors()) == 0 {
		t.Error("expected error for state var 'x' in binary expr under CtxPure")
	}
}

// ─── ExprContext.String() ─────────────────────────────────────────────────────

func TestExprContextString(t *testing.T) {
	cases := []struct {
		ctx  ExprContext
		want string
	}{
		{CtxPure, "pure function"},
		{CtxStateRead, "invariant/temporal/guard"},
		{CtxStateUpdate, "action body"},
		{CtxUnknown, "unknown"},
	}
	for _, tc := range cases {
		if got := tc.ctx.String(); got != tc.want {
			t.Errorf("ExprContext(%d).String() = %q, want %q", tc.ctx, got, tc.want)
		}
	}
}
