// Package repair suggests and validates candidate repairs for Spectre spec
// invariant violations.
//
// A repair is always labelled "CANDIDATE" — it must be reviewed before being
// applied, because automatically derived repairs can be over-restrictive,
// introduce deadlocks, or break liveness properties.
package repair

import (
	"fmt"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// RepairKind classifies how a candidate repair strengthens the spec.
type RepairKind int

const (
	// RepairStrengthenRequire adds a require predicate to the triggering action.
	RepairStrengthenRequire RepairKind = iota
	// RepairAddInvariantGuard is a stricter version derived from the invariant bound.
	RepairAddInvariantGuard
)

func (k RepairKind) String() string {
	switch k {
	case RepairStrengthenRequire:
		return "STRENGTHEN_REQUIRE"
	case RepairAddInvariantGuard:
		return "ADD_INVARIANT_GUARD"
	default:
		return "UNKNOWN"
	}
}

// RepairCandidate is a single suggested repair.  Always labelled CANDIDATE —
// the caller must validate it before applying.
type RepairCandidate struct {
	Kind            RepairKind
	ActionName      string // action to which the repair applies
	ViolatedInvariant string // invariant that triggered this suggestion
	RequireExpr     string // Spectre expression to add as a require statement
	Description     string
	Confidence      string // "low", "medium", or "high"
}

func (r *RepairCandidate) String() string {
	return fmt.Sprintf("[CANDIDATE/%s] action %s: require %s (%s confidence) — %s",
		r.Kind, r.ActionName, r.RequireExpr, r.Confidence, r.Description)
}

// RepairValidation records whether a candidate repair is safe to apply.
type RepairValidation struct {
	Candidate        *RepairCandidate
	OriginalFixed    bool    // original violation no longer appears
	HasDeadlock      bool    // repair introduces unreachable-action states
	ReachabilityDrop float64 // fraction of original reachable states lost (0–1)
	Safe             bool    // OriginalFixed && !HasDeadlock && ReachabilityDrop < 0.5
	Note             string
}

// SuggestRepairs analyses a violation and returns candidate repairs for the
// triggering action.  Every candidate is labelled CANDIDATE and carries a
// confidence rating.
func SuggestRepairs(violation *explore.Violation, file *ast.File) []RepairCandidate {
	var candidates []RepairCandidate

	invExpr := invariantExpr(file, violation.Invariant)
	if invExpr == "" {
		return nil
	}

	triggerAction := lastActionInPath(violation.Path)
	if triggerAction == "" {
		return nil
	}

	// Candidate 1 (low): add require <invariant_expr> to prevent the action
	// from firing when the invariant already fails.
	candidates = append(candidates, RepairCandidate{
		Kind:              RepairStrengthenRequire,
		ActionName:        triggerAction,
		ViolatedInvariant: violation.Invariant,
		RequireExpr:       invExpr,
		Confidence:        "low",
		Description: fmt.Sprintf(
			"prevent %q from firing when invariant %q already fails; "+
				"does not prevent the invariant from being violated for the first time",
			triggerAction, violation.Invariant),
	})

	// Candidate 2 (medium): if the invariant is a simple comparison, generate
	// a tighter pre-condition that prevents the action from crossing the bound.
	if tighter := tighterBound(invExpr); tighter != "" {
		candidates = append(candidates, RepairCandidate{
			Kind:              RepairAddInvariantGuard,
			ActionName:        triggerAction,
			ViolatedInvariant: violation.Invariant,
			RequireExpr:       tighter,
			Confidence:        "medium",
			Description: fmt.Sprintf(
				"prevent %q from producing a state that violates %q by "+
					"requiring a stricter pre-condition; review arithmetic carefully",
				triggerAction, violation.Invariant),
		})
	}

	return candidates
}

// ValidateRepair re-runs BFS on a patched version of specSrc (with the
// candidate repair inserted) and checks whether the original violation
// disappears and whether new problems are introduced.
//
// originalReachable is the state count from the pre-repair BFS; it is used to
// compute ReachabilityDrop.  maxStates bounds the validation BFS (0 = 5000).
func ValidateRepair(candidate *RepairCandidate, specSrc string, originalReachable, maxStates int) *RepairValidation {
	rv := &RepairValidation{Candidate: candidate}
	if maxStates <= 0 {
		maxStates = 5000
	}

	patched := PatchSpec(specSrc, candidate.ActionName, candidate.RequireExpr)
	if patched == "" {
		rv.Note = "could not patch spec: action not found in source"
		return rv
	}

	sm, err := parseAndBuild(patched)
	if err != nil {
		rv.Note = fmt.Sprintf("patched spec has errors: %v", err)
		return rv
	}

	e := explore.NewExplorer(sm)
	e.SetMaxStates(maxStates)
	e.SetSilent(true)
	result, bfsErr := e.ExploreBFS()
	if bfsErr != nil {
		rv.Note = fmt.Sprintf("BFS on patched spec failed: %v", bfsErr)
		return rv
	}

	rv.OriginalFixed = !violationsContainInvariant(result.Violations, candidate.ViolatedInvariant)

	if originalReachable > 0 {
		drop := float64(originalReachable-result.StatesExplored) / float64(originalReachable)
		if drop < 0 {
			drop = 0
		}
		rv.ReachabilityDrop = drop
	}

	// Deadlock check: any reachable state from which no action is enabled.
	for _, s := range result.ReachableStates {
		actions, aErr := sm.GetAvailableActionsWithArgs(s)
		if aErr == nil && len(actions) == 0 {
			rv.HasDeadlock = true
			break
		}
	}

	rv.Safe = rv.OriginalFixed && !rv.HasDeadlock && rv.ReachabilityDrop < 0.5
	rv.Note = formatNote(rv)
	return rv
}

// PatchSpec inserts `require <expr>` as the first statement of actionName in
// the spec source.  Returns "" if the action is not found.
func PatchSpec(src, actionName, requireExpr string) string {
	search := "action " + actionName
	idx := strings.Index(src, search)
	if idx < 0 {
		return ""
	}
	brace := strings.Index(src[idx:], "{")
	if brace < 0 {
		return ""
	}
	insertAt := idx + brace + 1
	return src[:insertAt] + "\n  require " + requireExpr + src[insertAt:]
}

// ── helpers ───────────────────────────────────────────────────────────────────

func invariantExpr(file *ast.File, name string) string {
	for _, decl := range file.Decls {
		if inv, ok := decl.(*ast.InvariantDecl); ok && inv.Name == name {
			return exprToString(inv.Condition)
		}
	}
	return ""
}

func lastActionInPath(path []*explore.Transition) string {
	if len(path) == 0 {
		return ""
	}
	return path[len(path)-1].Action
}

// tighterBound generates a stricter precondition from a simple comparison.
// "x < N" → "x < N - 1"; "x <= N" → "x < N"; "x >= N" → "x > N".
func tighterBound(expr string) string {
	expr = strings.TrimSpace(expr)
	for _, op := range []string{" <= ", " >= ", " < ", " > "} {
		idx := strings.Index(expr, op)
		if idx < 0 {
			continue
		}
		lhs := strings.TrimSpace(expr[:idx])
		rhs := strings.TrimSpace(expr[idx+len(op):])
		if strings.ContainsAny(lhs, " ()+*/") || strings.ContainsAny(rhs, " ()+*/") {
			continue
		}
		switch op {
		case " < ":
			return lhs + " < " + rhs + " - 1"
		case " <= ":
			return lhs + " < " + rhs
		case " >= ":
			return lhs + " > " + rhs
		case " > ":
			return lhs + " > " + rhs + " + 1"
		}
	}
	return ""
}

func exprToString(e ast.Expr) string {
	if e == nil {
		return ""
	}
	switch n := e.(type) {
	case *ast.Ident:
		if n.Prime {
			return n.Name + "'"
		}
		return n.Name
	case *ast.BasicLit:
		return n.Value
	case *ast.BinaryExpr:
		return exprToString(n.Left) + " " + binOpStr(n.Op) + " " + exprToString(n.Right)
	case *ast.UnaryExpr:
		if n.Op == ast.Not {
			return "!" + exprToString(n.Expr)
		}
		return "-" + exprToString(n.Expr)
	case *ast.CallExpr:
		var args []string
		for _, a := range n.Args {
			args = append(args, exprToString(a))
		}
		return exprToString(n.Fun) + "(" + strings.Join(args, ", ") + ")"
	case *ast.SelectorExpr:
		return exprToString(n.X) + "." + n.Sel
	case *ast.ParenExpr:
		return "(" + exprToString(n.X) + ")"
	default:
		return "<expr>"
	}
}

func binOpStr(op ast.BinaryOp) string {
	switch op {
	case ast.Add:
		return "+"
	case ast.Sub:
		return "-"
	case ast.Mul:
		return "*"
	case ast.Div:
		return "/"
	case ast.Eq:
		return "=="
	case ast.Neq:
		return "!="
	case ast.Lt:
		return "<"
	case ast.Leq:
		return "<="
	case ast.Gt:
		return ">"
	case ast.Geq:
		return ">="
	case ast.And:
		return "&&"
	case ast.Or:
		return "||"
	default:
		return "?"
	}
}

func parseAndBuild(src string) (*exec.StateMachine, error) {
	l := lexer.New(src)
	p := parser.New(l)
	file := p.ParseFile()
	if errs := p.Errors(); len(errs) > 0 {
		return nil, fmt.Errorf("parse errors: %s", strings.Join(errs, "; "))
	}
	return exec.NewStateMachine(file)
}

func violationsContainInvariant(violations []*explore.Violation, name string) bool {
	for _, v := range violations {
		if v.Invariant == name {
			return true
		}
	}
	return false
}

func formatNote(rv *RepairValidation) string {
	var parts []string
	if rv.OriginalFixed {
		parts = append(parts, "original violation fixed")
	} else {
		parts = append(parts, "original violation persists")
	}
	if rv.HasDeadlock {
		parts = append(parts, "introduces deadlock")
	}
	if rv.ReachabilityDrop > 0.01 {
		parts = append(parts, fmt.Sprintf("%.0f%% of states unreachable", rv.ReachabilityDrop*100))
	}
	if rv.Safe {
		parts = append(parts, "SAFE TO APPLY (review before committing)")
	} else {
		parts = append(parts, "NOT SAFE — review carefully")
	}
	return strings.Join(parts, "; ")
}
