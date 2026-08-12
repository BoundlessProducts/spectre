package diagnose

import (
	"fmt"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// Explanation is a structured analysis of a single counterexample violation.
type Explanation struct {
	InvariantName string
	ActionName    string
	// FirstDivergence is the path index where a relevant variable first changed
	// in a way that contributed to the violation.
	FirstDivergence int
	// ViolatedGuard is the WP condition that was not satisfied (from CEGIS).
	ViolatedGuard string
	// MinimalVars is the subset of state variables actually referenced by the invariant
	// and that changed along the counterexample path.
	MinimalVars []string
	// ResponsiblePath lists only the actions that modified at least one MinimalVar.
	ResponsiblePath []string
	// CandidateFix is the suggested require guard in Spectre syntax.
	CandidateFix string
	// Summary is a human-readable one-paragraph description of the violation.
	Summary string
}

// Explain produces a structured explanation for a counterexample violation.
// repair may be nil (CEGIS could not synthesize a fix).
func Explain(v *explore.Violation, repair *CEGISRepair, file *ast.File) *Explanation {
	invName := v.Invariant
	actionName := ""
	if len(v.Path) > 0 {
		actionName = v.Path[len(v.Path)-1].Action
	}

	// Collect variable names referenced by the invariant expression.
	invDecl := findInvariantDecl(file, invName)
	invVars := map[string]bool{}
	if invDecl != nil {
		collectVarRefs(invDecl.Condition, invVars)
	}

	// Find variables that actually changed along the path.
	changedVars := changedAlongPath(v.Path)

	// MinimalVars = invariant vars ∩ changed vars.
	minimal := []string{}
	for name := range invVars {
		if changedVars[name] {
			minimal = append(minimal, name)
		}
	}
	if len(minimal) == 0 {
		// Fall back to all changed vars if intersection is empty.
		for name := range changedVars {
			minimal = append(minimal, name)
		}
	}

	// ResponsiblePath: transitions that modified at least one minimal var.
	responsible := []string{}
	for _, t := range v.Path {
		if t.Action == "" {
			continue
		}
		if transitionModifies(t, minimal) {
			responsible = append(responsible, t.Action)
		}
	}

	// FirstDivergence: earliest step where a minimal var changed.
	firstDiv := 0
	for i, t := range v.Path {
		if transitionModifies(t, minimal) {
			firstDiv = i
			break
		}
	}

	// CandidateFix and ViolatedGuard from CEGIS.
	candidateFix := "no fix synthesized"
	violatedGuard := ""
	if repair != nil && len(repair.Guards) > 0 {
		g := repair.Guards[0]
		candidateFix = g.SpectreCode
		violatedGuard = g.WPString
	}

	// Build summary.
	preDesc := ""
	if repair != nil && repair.PreState != nil {
		pairs := []string{}
		for k, val := range repair.PreState.Variables {
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, val.String()))
		}
		if len(pairs) > 0 {
			preDesc = " with pre-state {" + strings.Join(pairs, ", ") + "}"
		}
	}
	summary := fmt.Sprintf(
		"Invariant `%s` was violated when action `%s` executed%s. "+
			"The relevant variable(s) are: %s. "+
			"The suggested fix is: %s",
		invName, actionName, preDesc,
		strings.Join(minimal, ", "),
		candidateFix,
	)

	return &Explanation{
		InvariantName:   invName,
		ActionName:      actionName,
		FirstDivergence: firstDiv,
		ViolatedGuard:   violatedGuard,
		MinimalVars:     minimal,
		ResponsiblePath: responsible,
		CandidateFix:    candidateFix,
		Summary:         summary,
	}
}

// PrintExplanation formats an Explanation for CLI output.
func PrintExplanation(e *Explanation) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  Explanation for `%s` via `%s`:\n", e.InvariantName, e.ActionName))
	b.WriteString(fmt.Sprintf("    Relevant vars    : %s\n", strings.Join(e.MinimalVars, ", ")))
	b.WriteString(fmt.Sprintf("    First divergence : step %d\n", e.FirstDivergence+1))
	if e.ViolatedGuard != "" {
		b.WriteString(fmt.Sprintf("    Violated guard   : %s\n", e.ViolatedGuard))
	}
	b.WriteString(fmt.Sprintf("    Responsible path : %s\n", strings.Join(e.ResponsiblePath, " → ")))
	b.WriteString(fmt.Sprintf("    Candidate fix    : %s\n", e.CandidateFix))
	return b.String()
}

// collectVarRefs populates set with all unprimed variable names referenced in expr.
func collectVarRefs(expr ast.Expr, set map[string]bool) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.Ident:
		if !e.Prime {
			set[e.Name] = true
		}
	case *ast.BinaryExpr:
		collectVarRefs(e.Left, set)
		collectVarRefs(e.Right, set)
	case *ast.UnaryExpr:
		collectVarRefs(e.Expr, set)
	case *ast.ParenExpr:
		collectVarRefs(e.X, set)
	case *ast.CallExpr:
		for _, a := range e.Args {
			collectVarRefs(a, set)
		}
	}
}

// changedAlongPath returns the set of variable names that changed value in any transition.
func changedAlongPath(path []*explore.Transition) map[string]bool {
	changed := map[string]bool{}
	for _, t := range path {
		if t.FromState == nil || t.ToState == nil {
			continue
		}
		for name, after := range t.ToState.Variables {
			before, ok := t.FromState.Variables[name]
			if !ok || before.String() != after.String() {
				changed[name] = true
			}
		}
	}
	return changed
}

// transitionModifies reports whether transition t changed any of the named variables.
func transitionModifies(t *explore.Transition, vars []string) bool {
	if t.FromState == nil || t.ToState == nil {
		return false
	}
	for _, name := range vars {
		before, bOk := t.FromState.Variables[name]
		after, aOk := t.ToState.Variables[name]
		if bOk && aOk && before.String() != after.String() {
			return true
		}
		if bOk != aOk {
			return true
		}
	}
	return false
}
