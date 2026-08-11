package codegen

import (
	"fmt"
	"strings"

	"github.com/akkeshavan/spectre/internal/diagnose"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// GenerateRustPatch translates a WP-derived Spectre guard into a suggested Rust
// source patch.  The refinement map in meta is used to convert spec variable names
// to Rust field names (self.<field>).  Action parameters are kept as-is because
// they already match Rust local-variable names in well-named code.
//
// The function does NOT auto-apply the patch; it emits a human-readable
// suggestion the developer must inspect and apply manually.
//
// Fallback behavior:
//   - If meta is nil or the action is not found, field mapping is skipped.
//   - If the expression uses constructs that cannot be translated, a comment
//     recommends manual translation and the Spectre expression is shown verbatim.
func GenerateRustPatch(guard diagnose.SynthGuard, actionName string, meta *DriverMeta) string {
	fieldMap := buildFieldMap(meta, actionName)

	rustExpr, err := exprToRust(guard.WPExpr, fieldMap)
	if err != nil {
		// Fallback: show the Spectre guard; developer translates manually.
		return fmt.Sprintf(
			"// Suggested Rust patch for action `%s` (translate manually):\n"+
				"// Spectre guard: %s\n"+
				"// if !(<translated condition>) { return Err(...); }",
			actionName, guard.WPString,
		)
	}

	// Determine whether the Rust method has a visible error-return path by
	// looking up the action in meta.  When unknown, default to assert!().
	hasErrReturn := methodHasErrReturn(meta, actionName)

	var sb strings.Builder
	fmt.Fprintf(&sb, "// Suggested Rust patch for `%s` (apply manually):\n", actionName)
	fmt.Fprintf(&sb, "// Spectre WP guard: %s\n", guard.WPString)
	if hasErrReturn {
		fmt.Fprintf(&sb, "if !(%s) {\n    return Err(/* appropriate error variant */);\n}", rustExpr)
	} else {
		fmt.Fprintf(&sb, "assert!(%s, \"invariant guard: %s\");", rustExpr, guard.WPString)
	}
	return sb.String()
}

// buildFieldMap constructs a mapping from spec variable name → Rust field accessor.
// For fields that exist in the refinement map, the accessor is "self.<rust_field>".
// For action parameters, they map to their own name (local variables in Rust).
func buildFieldMap(meta *DriverMeta, actionName string) map[string]string {
	m := make(map[string]string)
	if meta == nil {
		return m
	}
	for _, f := range meta.Fields {
		m[f.SpecVar] = "self." + f.RustField
	}
	// Action parameters shadow field names with their own local names.
	for _, a := range meta.Actions {
		if a.SpecAction == actionName {
			for _, p := range a.Params {
				m[p.Name] = p.Name // local variable
			}
		}
	}
	return m
}

// methodHasErrReturn returns true only when the driver sidecar explicitly marks
// the Rust method as returning Result<_, _> via the "returns_result" field.
// Defaults to false (emit assert! instead of return Err) because assert! is
// valid in any Rust method while return Err requires a Result return type.
func methodHasErrReturn(meta *DriverMeta, actionName string) bool {
	if meta == nil {
		return false
	}
	for _, a := range meta.Actions {
		if a.SpecAction == actionName {
			return a.ReturnsResult
		}
	}
	return false
}

// exprToRust translates a Spectre AST expression to a Rust expression string.
// fieldMap maps spec variable names to Rust accessor forms.
func exprToRust(e ast.Expr, fieldMap map[string]string) (string, error) {
	switch expr := e.(type) {
	case *ast.Ident:
		if expr.Prime {
			return "", fmt.Errorf("primed variable %s cannot appear in Rust patch", expr.Name)
		}
		if rust, ok := fieldMap[expr.Name]; ok {
			return rust, nil
		}
		return expr.Name, nil // unknown: use spec name verbatim

	case *ast.BasicLit:
		switch expr.Value {
		case "true":
			return "true", nil
		case "false":
			return "false", nil
		}
		return expr.Value, nil

	case *ast.UnaryExpr:
		inner, err := exprToRust(expr.Expr, fieldMap)
		if err != nil {
			return "", err
		}
		switch expr.Op {
		case ast.Not:
			return fmt.Sprintf("!(%s)", inner), nil
		case ast.Neg:
			return fmt.Sprintf("-%s", inner), nil
		default:
			return "", fmt.Errorf("unsupported unary op %v", expr.Op)
		}

	case *ast.BinaryExpr:
		left, err := exprToRust(expr.Left, fieldMap)
		if err != nil {
			return "", err
		}
		right, err := exprToRust(expr.Right, fieldMap)
		if err != nil {
			return "", err
		}
		// Wrap sub-expressions in parentheses when they involve lower-precedence ops
		// to avoid Rust precedence surprises.
		if needsRustParens(expr.Left, expr.Op) {
			left = "(" + left + ")"
		}
		if needsRustParens(expr.Right, expr.Op) {
			right = "(" + right + ")"
		}
		op, err := rustBinOp(expr.Op)
		if err != nil {
			return "", err
		}
		return left + " " + op + " " + right, nil

	case *ast.ParenExpr:
		inner, err := exprToRust(expr.X, fieldMap)
		if err != nil {
			return "", err
		}
		return "(" + inner + ")", nil

	default:
		return "", fmt.Errorf("unsupported expression type %T", e)
	}
}

func rustBinOp(op ast.BinaryOp) (string, error) {
	switch op {
	case ast.Add:
		return "+", nil
	case ast.Sub:
		return "-", nil
	case ast.Mul:
		return "*", nil
	case ast.Div:
		return "/", nil
	case ast.Eq:
		return "==", nil // Spectre = is equality; Rust uses ==
	case ast.Neq:
		return "!=", nil
	case ast.Lt:
		return "<", nil
	case ast.Gt:
		return ">", nil
	case ast.Leq:
		return "<=", nil
	case ast.Geq:
		return ">=", nil
	case ast.And:
		return "&&", nil
	case ast.Or:
		return "||", nil
	default:
		return "", fmt.Errorf("unsupported binary op %v", op)
	}
}

// needsRustParens reports whether a child expression requires explicit parentheses
// when it appears as an operand of parentOp.  Uses operator precedence and
// right-operand associativity rules to avoid producing incorrect Rust expressions
// (e.g. `a - b + c` instead of `a - (b + c)`).
func needsRustParens(sub ast.Expr, parentOp ast.BinaryOp) bool {
	bin, ok := sub.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	childPrec := rustPrec(bin.Op)
	parentPrec := rustPrec(parentOp)
	if childPrec < parentPrec {
		return true
	}
	// For Sub and Div, right operands at equal precedence are not associative:
	// a - (b - c) ≠ a - b - c, a / (b / c) ≠ a / b / c.
	if childPrec == parentPrec {
		switch parentOp {
		case ast.Sub, ast.Div:
			return true
		}
	}
	return false
}

// rustPrec returns a numeric precedence level for a binary op (higher = binds tighter).
func rustPrec(op ast.BinaryOp) int {
	switch op {
	case ast.Or:
		return 1
	case ast.And:
		return 2
	case ast.Eq, ast.Neq, ast.Lt, ast.Gt, ast.Leq, ast.Geq:
		return 3
	case ast.Add, ast.Sub:
		return 4
	case ast.Mul, ast.Div:
		return 5
	}
	return 0
}
