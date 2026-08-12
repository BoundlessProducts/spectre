package diagnose

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// SMTResult classifies the outcome of an SMT equivalence check.
type SMTResult int

const (
	// SMTEquivalent means the solver proved the two expressions identical for all inputs.
	SMTEquivalent SMTResult = iota
	// SMTChanged means the solver found a valuation where the expressions differ.
	SMTChanged
	// SMTUnknown means the check timed out, used unsupported constructs, or z3 was unavailable.
	SMTUnknown
)

func (r SMTResult) String() string {
	switch r {
	case SMTEquivalent:
		return "equivalent"
	case SMTChanged:
		return "changed"
	default:
		return "unknown"
	}
}

// CheckExprEquivalence uses Z3 to decide whether oldExpr and newExpr are
// semantically equivalent over all integer/Boolean valuations of their free
// variables.
//
// Returns:
//   - (SMTEquivalent, "")         – UNSAT: expressions are equal for all inputs
//   - (SMTChanged,   witness)     – SAT:   found a distinguishing valuation
//   - (SMTUnknown,   "")          – timeout / unsupported construct / z3 absent
func CheckExprEquivalence(oldExpr, newExpr ast.Expr) (SMTResult, string) {
	vars := collectAllVars(oldExpr, newExpr)
	boolVars := inferBoolVars(oldExpr, newExpr)

	var sb strings.Builder
	// Bool-sorted variables are not part of the strict QF_LIA signature.
	// Omit set-logic when Bool vars are present so Z3 uses its default (ALL).
	if len(boolVars) == 0 {
		sb.WriteString("(set-logic QF_LIA)\n")
	}
	for _, v := range vars {
		s := "Int"
		if boolVars[v] {
			s = "Bool"
		}
		fmt.Fprintf(&sb, "(declare-const %s %s)\n", smtID(v), s)
	}

	oldSMT, err1 := toSMT(oldExpr, boolVars)
	newSMT, err2 := toSMT(newExpr, boolVars)
	if err1 != nil || err2 != nil {
		return SMTUnknown, ""
	}

	// Ask: ∃x⃗. e_old(x⃗) ≠ e_new(x⃗)?
	if exprResultIsBool(oldExpr, boolVars) {
		fmt.Fprintf(&sb, "(assert (xor %s %s))\n", oldSMT, newSMT)
	} else {
		fmt.Fprintf(&sb, "(assert (not (= %s %s)))\n", oldSMT, newSMT)
	}
	sb.WriteString("(check-sat)\n(get-model)\n")

	out, err := callZ3(sb.String(), 5*time.Second)
	if err != nil {
		return SMTUnknown, ""
	}

	lines := strings.SplitN(strings.TrimSpace(out), "\n", 2)
	verdict := strings.TrimSpace(lines[0])
	switch verdict {
	case "unsat":
		return SMTEquivalent, ""
	case "sat":
		witness := ""
		if len(lines) > 1 {
			witness = extractWitness(lines[1], vars)
		}
		return SMTChanged, witness
	}
	return SMTUnknown, ""
}

// toSMT converts a Spectre expression to SMT-LIB2 notation (QF_LIA fragment).
// Returns an error for constructs outside the supported fragment.
func toSMT(e ast.Expr, boolVars map[string]bool) (string, error) {
	switch expr := e.(type) {
	case *ast.Ident:
		if expr.Prime {
			return "", fmt.Errorf("primed variable %s not supported in SMT query", expr.Name)
		}
		return smtID(expr.Name), nil

	case *ast.BasicLit:
		switch expr.Value {
		case "true":
			return "true", nil
		case "false":
			return "false", nil
		}
		if strings.HasPrefix(expr.Value, "-") {
			return fmt.Sprintf("(- %s)", expr.Value[1:]), nil
		}
		return expr.Value, nil

	case *ast.UnaryExpr:
		inner, err := toSMT(expr.Expr, boolVars)
		if err != nil {
			return "", err
		}
		switch expr.Op {
		case ast.Not:
			return fmt.Sprintf("(not %s)", inner), nil
		case ast.Neg:
			return fmt.Sprintf("(- %s)", inner), nil
		default:
			return "", fmt.Errorf("unsupported unary op %v", expr.Op)
		}

	case *ast.BinaryExpr:
		left, err := toSMT(expr.Left, boolVars)
		if err != nil {
			return "", err
		}
		right, err := toSMT(expr.Right, boolVars)
		if err != nil {
			return "", err
		}
		switch expr.Op {
		case ast.Add:
			return fmt.Sprintf("(+ %s %s)", left, right), nil
		case ast.Sub:
			return fmt.Sprintf("(- %s %s)", left, right), nil
		case ast.Mul:
			return fmt.Sprintf("(* %s %s)", left, right), nil
		case ast.Div:
			return fmt.Sprintf("(div %s %s)", left, right), nil
		case ast.Eq:
			return fmt.Sprintf("(= %s %s)", left, right), nil
		case ast.Neq:
			return fmt.Sprintf("(not (= %s %s))", left, right), nil
		case ast.Lt:
			return fmt.Sprintf("(< %s %s)", left, right), nil
		case ast.Gt:
			return fmt.Sprintf("(> %s %s)", left, right), nil
		case ast.Leq:
			return fmt.Sprintf("(<= %s %s)", left, right), nil
		case ast.Geq:
			return fmt.Sprintf("(>= %s %s)", left, right), nil
		case ast.And:
			return fmt.Sprintf("(and %s %s)", left, right), nil
		case ast.Or:
			return fmt.Sprintf("(or %s %s)", left, right), nil
		default:
			return "", fmt.Errorf("unsupported binary op %v", expr.Op)
		}

	case *ast.ParenExpr:
		return toSMT(expr.X, boolVars)

	default:
		return "", fmt.Errorf("unsupported expression type %T", e)
	}
}

// exprResultIsBool returns true if the top-level expression produces a Bool value.
func exprResultIsBool(e ast.Expr, boolVars map[string]bool) bool {
	switch expr := e.(type) {
	case *ast.Ident:
		return boolVars[expr.Name]
	case *ast.BasicLit:
		return expr.Value == "true" || expr.Value == "false"
	case *ast.BinaryExpr:
		switch expr.Op {
		case ast.And, ast.Or, ast.Eq, ast.Neq, ast.Lt, ast.Gt, ast.Leq, ast.Geq:
			return true
		}
		return false
	case *ast.UnaryExpr:
		return expr.Op == ast.Not
	case *ast.ParenExpr:
		return exprResultIsBool(expr.X, boolVars)
	}
	return false
}

// inferBoolVars returns names of variables that appear directly in boolean-typed
// positions (operands of &&, ||, !) as opposed to arithmetic/comparison positions.
func inferBoolVars(exprs ...ast.Expr) map[string]bool {
	boolVars := make(map[string]bool)
	for _, e := range exprs {
		markBoolCtx(e, false, boolVars)
	}
	return boolVars
}

func markBoolCtx(e ast.Expr, inBoolCtx bool, out map[string]bool) {
	if e == nil {
		return
	}
	switch expr := e.(type) {
	case *ast.Ident:
		if inBoolCtx && !expr.Prime {
			out[expr.Name] = true
		}
	case *ast.BinaryExpr:
		switch expr.Op {
		case ast.And, ast.Or:
			markBoolCtx(expr.Left, true, out)
			markBoolCtx(expr.Right, true, out)
		case ast.Eq, ast.Neq:
			// If one side is a Bool literal (true/false), the other is also Bool.
			if isBoolLit(expr.Left) || isBoolLit(expr.Right) {
				markBoolCtx(expr.Left, true, out)
				markBoolCtx(expr.Right, true, out)
			} else {
				markBoolCtx(expr.Left, false, out)
				markBoolCtx(expr.Right, false, out)
			}
		default:
			// comparison or arithmetic: operands are Int context
			markBoolCtx(expr.Left, false, out)
			markBoolCtx(expr.Right, false, out)
		}
	case *ast.UnaryExpr:
		markBoolCtx(expr.Expr, expr.Op == ast.Not, out)
	case *ast.ParenExpr:
		markBoolCtx(expr.X, inBoolCtx, out)
	}
}

// collectAllVars returns sorted distinct free-variable names across all expressions.
func collectAllVars(exprs ...ast.Expr) []string {
	seen := make(map[string]bool)
	for _, e := range exprs {
		walkIdents(e, func(id *ast.Ident) {
			if !id.Prime {
				seen[id.Name] = true
			}
		})
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func walkIdents(e ast.Expr, fn func(*ast.Ident)) {
	if e == nil {
		return
	}
	switch expr := e.(type) {
	case *ast.Ident:
		fn(expr)
	case *ast.BinaryExpr:
		walkIdents(expr.Left, fn)
		walkIdents(expr.Right, fn)
	case *ast.UnaryExpr:
		walkIdents(expr.Expr, fn)
	case *ast.ParenExpr:
		walkIdents(expr.X, fn)
	case *ast.SelectorExpr:
		walkIdents(expr.X, fn)
	case *ast.CallExpr:
		walkIdents(expr.Fun, fn)
		for _, a := range expr.Args {
			walkIdents(a, fn)
		}
	case *ast.IndexExpr:
		walkIdents(expr.X, fn)
		walkIdents(expr.Index, fn)
	}
}

// smtID returns an SMT-LIB2-safe identifier for a Spectre variable.
// ASCII-only names pass through unchanged. Names with non-ASCII characters
// are wrapped in SMT-LIB2 quoted-symbol notation (| ... |) which accepts
// any printable character except '|' and '\'.
func smtID(name string) string {
	for _, r := range name {
		if r > 127 {
			return "|" + strings.ReplaceAll(strings.ReplaceAll(name, "|", "_"), "\\", "_") + "|"
		}
	}
	return name
}

// callZ3 invokes Z3 with the given SMT-LIB2 input and a timeout.
func callZ3(input string, timeout time.Duration) (string, error) {
	z3 := findZ3()
	if z3 == "" {
		return "", fmt.Errorf("z3 not found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, z3, "-in", "-smt2")
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return "", fmt.Errorf("z3 timed out")
	}
	if err != nil {
		// z3 exits non-zero on "unsat" in some builds; treat output as valid if it has a verdict.
		s := string(out)
		if strings.Contains(s, "unsat") || strings.Contains(s, "sat") {
			return s, nil
		}
		return "", fmt.Errorf("z3 error: %w", err)
	}
	return string(out), nil
}

// findZ3 searches for the z3 binary in PATH then known Homebrew locations.
func findZ3() string {
	if path, err := exec.LookPath("z3"); err == nil {
		return path
	}
	for _, p := range []string{"/opt/homebrew/bin/z3", "/usr/local/bin/z3", "/usr/bin/z3"} {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	return ""
}

// extractWitness builds a short human-readable witness string from Z3's model output.
// Z3's (get-model) format places the value on the line *after* the define-fun header:
//
//	(define-fun x () Int
//	  42)
func extractWitness(model string, vars []string) string {
	var parts []string
	lines := strings.Split(model, "\n")
	for _, v := range vars {
		prefix := "(define-fun " + v + " "
		for i, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), prefix) {
				continue
			}
			// Value is on the next non-empty line.
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if next == "" {
					continue
				}
				val := strings.TrimRight(next, ")")
				val = strings.TrimSpace(val)
				if val != "" {
					parts = append(parts, v+"="+val)
				}
				break
			}
			break
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

// isBoolLit reports whether e is the Boolean literal true or false.
func isBoolLit(e ast.Expr) bool {
	lit, ok := e.(*ast.BasicLit)
	return ok && (lit.Value == "true" || lit.Value == "false")
}
