package parser

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func TestTemporalDeclaration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		validate func(*testing.T, ast.Decl)
	}{
		{
			name: "Simple always temporal",
			input: `temporal alwaysPositive {
  always (counter >= 0)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				if temporalDecl.Name != "alwaysPositive" {
					t.Errorf("temporal name not 'alwaysPositive'. got=%s", temporalDecl.Name)
				}
				alwaysExpr, ok := temporalDecl.Expression.(*ast.AlwaysExpr)
				if !ok {
					t.Fatalf("expression not *ast.AlwaysExpr. got=%T", temporalDecl.Expression)
				}
				if alwaysExpr.Expr == nil {
					t.Fatal("always expression is nil")
				}
			},
		},
		{
			name: "Simple eventually temporal",
			input: `temporal eventuallyReachesTen {
  eventually (counter = 10)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				if temporalDecl.Name != "eventuallyReachesTen" {
					t.Errorf("temporal name not 'eventuallyReachesTen'. got=%s", temporalDecl.Name)
				}
				eventuallyExpr, ok := temporalDecl.Expression.(*ast.EventuallyExpr)
				if !ok {
					t.Fatalf("expression not *ast.EventuallyExpr. got=%T", temporalDecl.Expression)
				}
				if eventuallyExpr.Expr == nil {
					t.Fatal("eventually expression is nil")
				}
			},
		},
		{
			name: "Until temporal",
			input: `temporal counterUntilTen {
  counter < 10 until counter = 10
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				if temporalDecl.Name != "counterUntilTen" {
					t.Errorf("temporal name not 'counterUntilTen'. got=%s", temporalDecl.Name)
				}
				untilExpr, ok := temporalDecl.Expression.(*ast.UntilExpr)
				if !ok {
					t.Fatalf("expression not *ast.UntilExpr. got=%T", temporalDecl.Expression)
				}
				if untilExpr.Left == nil || untilExpr.Right == nil {
					t.Fatal("until expression has nil left or right")
				}
			},
		},
		{
			name: "Temporal with description",
			input: `description "Ensures counter remains non-negative"
temporal alwaysNonNegative {
  always (counter >= 0)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				if temporalDecl.Description != "Ensures counter remains non-negative" {
					t.Errorf("description not set correctly. got=%q", temporalDecl.Description)
				}
			},
		},
		{
			name: "Nested always eventually",
			input: `temporal alwaysEventually {
  always eventually (counter = 0)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				alwaysExpr, ok := temporalDecl.Expression.(*ast.AlwaysExpr)
				if !ok {
					t.Fatalf("outer expression not *ast.AlwaysExpr. got=%T", temporalDecl.Expression)
				}
				_, ok = alwaysExpr.Expr.(*ast.EventuallyExpr)
				if !ok {
					t.Fatalf("inner expression not *ast.EventuallyExpr. got=%T", alwaysExpr.Expr)
				}
			},
		},
		{
			name: "Nested eventually always",
			input: `temporal eventuallyAlways {
  eventually always (counter = 0)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				eventuallyExpr, ok := temporalDecl.Expression.(*ast.EventuallyExpr)
				if !ok {
					t.Fatalf("outer expression not *ast.EventuallyExpr. got=%T", temporalDecl.Expression)
				}
				_, ok = eventuallyExpr.Expr.(*ast.AlwaysExpr)
				if !ok {
					t.Fatalf("inner expression not *ast.AlwaysExpr. got=%T", eventuallyExpr.Expr)
				}
			},
		},
		{
			name: "Weak fairness on action",
			input: `temporal incrementFairness {
  WF(increment)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				wfExpr, ok := temporalDecl.Expression.(*ast.WFExpr)
				if !ok {
					t.Fatalf("expression not *ast.WFExpr. got=%T", temporalDecl.Expression)
				}
				if wfExpr.Target == nil {
					t.Fatal("WF target is nil")
				}
				ident, ok := wfExpr.Target.(*ast.Ident)
				if !ok {
					t.Fatalf("WF target not *ast.Ident. got=%T", wfExpr.Target)
				}
				if ident.Name != "increment" {
					t.Errorf("WF target name not 'increment'. got=%s", ident.Name)
				}
			},
		},
		{
			name: "Strong fairness on action",
			input: `temporal decrementFairness {
  SF(decrement)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				sfExpr, ok := temporalDecl.Expression.(*ast.SFExpr)
				if !ok {
					t.Fatalf("expression not *ast.SFExpr. got=%T", temporalDecl.Expression)
				}
				if sfExpr.Target == nil {
					t.Fatal("SF target is nil")
				}
				ident, ok := sfExpr.Target.(*ast.Ident)
				if !ok {
					t.Fatalf("SF target not *ast.Ident. got=%T", sfExpr.Target)
				}
				if ident.Name != "decrement" {
					t.Errorf("SF target name not 'decrement'. got=%s", ident.Name)
				}
			},
		},
		{
			name: "Weak fairness on variable",
			input: `temporal counterFairness {
  WF(counter)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				wfExpr, ok := temporalDecl.Expression.(*ast.WFExpr)
				if !ok {
					t.Fatalf("expression not *ast.WFExpr. got=%T", temporalDecl.Expression)
				}
				ident, ok := wfExpr.Target.(*ast.Ident)
				if !ok {
					t.Fatalf("WF target not *ast.Ident. got=%T", wfExpr.Target)
				}
				if ident.Name != "counter" {
					t.Errorf("WF target name not 'counter'. got=%s", ident.Name)
				}
			},
		},
		{
			name: "Fairness with description",
			input: `description "Ensures increment action executes fairly"
temporal incrementFairness {
  WF(increment)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				if temporalDecl.Description != "Ensures increment action executes fairly" {
					t.Errorf("description not set correctly. got=%q", temporalDecl.Description)
				}
			},
		},
		{
			name: "Leads-to operator",
			input: `temporal requestLeadsToResponse {
  requestSent → responseReceived
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				leadsToExpr, ok := temporalDecl.Expression.(*ast.LeadsToExpr)
				if !ok {
					t.Fatalf("expression not *ast.LeadsToExpr. got=%T", temporalDecl.Expression)
				}
				if leadsToExpr.Left == nil || leadsToExpr.Right == nil {
					t.Fatal("leads-to expression has nil left or right")
				}
			},
		},
		{
			name: "Always with leads-to",
			input: `temporal progress {
  always (counter < 10 → eventually counter = 10)
}`,
			validate: func(t *testing.T, decl ast.Decl) {
				temporalDecl, ok := decl.(*ast.TemporalDecl)
				if !ok {
					t.Fatalf("not *ast.TemporalDecl. got=%T", decl)
				}
				alwaysExpr, ok := temporalDecl.Expression.(*ast.AlwaysExpr)
				if !ok {
					t.Fatalf("outer expression not *ast.AlwaysExpr. got=%T", temporalDecl.Expression)
				}
				leadsToExpr, ok := alwaysExpr.Expr.(*ast.LeadsToExpr)
				if !ok {
					t.Fatalf("inner expression not *ast.LeadsToExpr. got=%T", alwaysExpr.Expr)
				}
				if leadsToExpr.Left == nil || leadsToExpr.Right == nil {
					t.Fatal("leads-to expression has nil left or right")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			decl := p.parseTemporalDecl()

			if len(p.Errors()) != 0 {
				t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
			}
			if decl == nil {
				t.Fatalf("parseTemporalDecl returned nil. Errors: %v", p.Errors())
			}

			tt.validate(t, decl)
		})
	}
}

