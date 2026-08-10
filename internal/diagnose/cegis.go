package diagnose

import (
	"fmt"
	"strings"

	"github.com/akkeshavan/spectre/internal/eval"
	"github.com/akkeshavan/spectre/internal/explore"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// CEGISRepair holds synthesized repair candidates for one invariant violation.
type CEGISRepair struct {
	InvariantName  string
	ActionName     string
	ActionArgs     []state.Value
	PreState       *state.State // state before the violating action
	ViolatingState *state.State // state that violated the invariant (nil for type-1)
	CounterexPath  []string     // action sequence; last entry marked with * if type-1
	Guards         []SynthGuard
}

// SynthGuard is one synthesized guard expression derived by weakest-precondition calculus.
type SynthGuard struct {
	VarName     string   // state variable whose assignment was responsible
	AssignRHS   string   // printed form of the assignment RHS (e.g. "balance - amount")
	WPExpr      ast.Expr // weakest-precondition expression ready to inject as RequireStmt
	WPString    string   // printed form of the wp (e.g. "balance - amount >= 0")
	Verified    bool     // set by commands.go after re-verification
	SpectreCode string   // the require line to add
}

// SynthesizeCEGIS analyzes invariant violations and derives guard candidates using
// weakest-precondition calculus.  Handles both violation kinds the explorer emits:
//
//   - Type 1: action execution failed ("would violate") — v.State is the pre-state
//   - Type 2: state reached then validated — v.Path[last] carries the action/pre-state
func SynthesizeCEGIS(violations []*explore.Violation, file *ast.File) []CEGISRepair {
	seen := make(map[string]bool)
	var repairs []CEGISRepair
	for _, v := range violations {
		r := synthesizeOne(v, file)
		if r == nil {
			continue
		}
		key := r.InvariantName + ":" + r.ActionName
		if seen[key] {
			continue
		}
		seen[key] = true
		repairs = append(repairs, *r)
	}
	return repairs
}

func synthesizeOne(v *explore.Violation, file *ast.File) *CEGISRepair {
	var actionName, invariantName string
	var preState, violatingState *state.State
	var pathActions []string
	var actionArgs []state.Value

	if strings.Contains(v.Description, "would violate") {
		// Type 1: action would have caused the violation; v.State is the pre-state.
		actionName = extractField(v.Description, "Action '", "'")
		invariantName = extractField(v.Description, "invariant ", " violated")
		preState = v.State
		violatingState = nil
		// Build path from existing path + the pending violating action.
		for _, t := range v.Path {
			if t.Action != "" {
				pathActions = append(pathActions, t.Action)
			}
		}
		if actionName != "" {
			pathActions = append(pathActions, actionName+"*")
		}
	} else {
		// Type 2: state reached and then invariant was checked.
		if len(v.Path) == 0 {
			return nil
		}
		last := v.Path[len(v.Path)-1]
		actionName = last.Action
		invariantName = v.Invariant
		preState = last.FromState
		violatingState = v.State
		actionArgs = last.Args
		for _, t := range v.Path {
			if t.Action != "" {
				pathActions = append(pathActions, t.Action)
			}
		}
	}

	if actionName == "" || invariantName == "" {
		return nil
	}

	actionDecl := findActionDecl(file, actionName)
	invDecl := findInvariantDecl(file, invariantName)
	if actionDecl == nil || invDecl == nil {
		return nil
	}

	guards := synthesizeGuards(invDecl.Condition, actionDecl, preState, actionArgs, file)

	return &CEGISRepair{
		InvariantName:  invariantName,
		ActionName:     actionName,
		ActionArgs:     actionArgs,
		PreState:       preState,
		ViolatingState: violatingState,
		CounterexPath:  pathActions,
		Guards:         guards,
	}
}

// synthesizeGuards computes a minimal wp-based guard for each assignment in the action
// body that affects a variable referenced by the invariant.
func synthesizeGuards(
	invariant ast.Expr,
	decl *ast.ActionDecl,
	preState *state.State,
	args []state.Value,
	file *ast.File,
) []SynthGuard {
	if decl.Body == nil {
		return nil
	}

	clauses := andClauses(invariant)
	seen := make(map[string]bool)
	var guards []SynthGuard

	for _, stmt := range decl.Body.Statements {
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok {
			continue
		}
		ident, ok := assign.Left.(*ast.Ident)
		if !ok || !ident.Prime {
			continue
		}
		varName := ident.Name

		// Compute wp(var' := rhs, I) = I[var←rhs] restricted to clauses referencing var.
		// This handles arithmetic, boolean, and enum assignments uniformly via substitution.
		wpExpr := computeWP(varName, assign.Right, clauses)
		if wpExpr == nil {
			continue // invariant doesn't mention this variable
		}

		// If the substituted expression has no state-variable references it is a
		// constant (the assignment unconditionally preserves or violates the invariant).
		// A constant-true wp needs no guard; a constant-false wp means no precondition
		// can repair this action — either way, skip rather than emit `require true/false`.
		if !exprHasStateVar(wpExpr) {
			continue
		}

		// Skip guards that are already satisfied in the pre-state: they wouldn't have
		// blocked the action, so they're not the root cause.
		if preState != nil && evalBoolInState(wpExpr, preState, file, decl, args) {
			continue
		}

		wpStr := PrintExpr(wpExpr)
		if seen[wpStr] {
			continue
		}
		seen[wpStr] = true

		rhsStr := PrintExpr(assign.Right)
		guards = append(guards, SynthGuard{
			VarName:     varName,
			AssignRHS:   rhsStr,
			WPExpr:      wpExpr,
			WPString:    wpStr,
			SpectreCode: fmt.Sprintf("// In action %s — add:\nrequire %s", decl.Name, wpStr),
		})
	}
	return guards
}

// computeWP returns the weakest precondition for the assignment var' = rhs with
// respect to the invariant clauses that reference var. Returns nil if no clause
// references the variable.
func computeWP(varName string, rhs ast.Expr, clauses []ast.Expr) ast.Expr {
	var relevant []ast.Expr
	for _, c := range clauses {
		if exprReferences(c, varName) {
			relevant = append(relevant, SubstituteExpr(c, varName, rhs))
		}
	}
	if len(relevant) == 0 {
		return nil
	}
	result := relevant[0]
	for _, c := range relevant[1:] {
		result = &ast.BinaryExpr{Op: ast.And, Left: result, Right: c}
	}
	return result
}

// andClauses decomposes a conjunction into its constituent clauses.
func andClauses(expr ast.Expr) []ast.Expr {
	if bin, ok := expr.(*ast.BinaryExpr); ok && bin.Op == ast.And {
		return append(andClauses(bin.Left), andClauses(bin.Right)...)
	}
	return []ast.Expr{expr}
}

// exprHasStateVar reports whether expr contains any unprimed identifier,
// i.e., references at least one state variable (as opposed to being a constant).
func exprHasStateVar(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return !e.Prime
	case *ast.BinaryExpr:
		return exprHasStateVar(e.Left) || exprHasStateVar(e.Right)
	case *ast.UnaryExpr:
		return exprHasStateVar(e.Expr)
	case *ast.ParenExpr:
		return exprHasStateVar(e.X)
	}
	return false
}

// exprReferences reports whether expr contains an unprimed reference to varName.
func exprReferences(expr ast.Expr, varName string) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return !e.Prime && e.Name == varName
	case *ast.BinaryExpr:
		return exprReferences(e.Left, varName) || exprReferences(e.Right, varName)
	case *ast.UnaryExpr:
		return exprReferences(e.Expr, varName)
	case *ast.ParenExpr:
		return exprReferences(e.X, varName)
	}
	return false
}

// evalBoolInState evaluates expr in s and returns true if the result is boolean true.
// Returns false on any error (safe default: keep the guard).
func evalBoolInState(expr ast.Expr, s *state.State, file *ast.File, decl *ast.ActionDecl, args []state.Value) bool {
	env := eval.NewEnvironment()
	eval.RegisterEnumTypes(env, file)
	for k, v := range s.Variables {
		env.SetVariable(k, v)
	}
	// Set action parameters so WP expressions referencing them can be evaluated.
	if decl != nil {
		for i, param := range decl.Parameters {
			if i < len(args) {
				env.SetVariable(param.Name, args[i])
			}
		}
	}
	result, err := eval.NewEvaluator(env).Eval(expr)
	if err != nil {
		return false
	}
	prim, ok := result.(*state.PrimitiveValue)
	if !ok || prim.TypeName != "bool" || prim.BoolValue == nil {
		return false
	}
	return *prim.BoolValue
}

// findInvariantDecl finds an invariant declaration by name, including inside modules.
func findInvariantDecl(file *ast.File, name string) *ast.InvariantDecl {
	for _, d := range file.Decls {
		if inv, ok := d.(*ast.InvariantDecl); ok && inv.Name == name {
			return inv
		}
		if m, ok := d.(*ast.ModuleDecl); ok {
			for _, md := range m.Decls {
				if inv, ok := md.(*ast.InvariantDecl); ok && inv.Name == name {
					return inv
				}
			}
		}
	}
	return nil
}

// SubstituteExpr replaces all unprimed occurrences of varName with replacement in expr.
// Exported so commands.go can use it when building RequireStmt for re-verification.
func SubstituteExpr(expr ast.Expr, varName string, replacement ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case *ast.Ident:
		if !e.Prime && e.Name == varName {
			return replacement
		}
		return e
	case *ast.BinaryExpr:
		return &ast.BinaryExpr{
			Position: e.Position,
			Op:       e.Op,
			Left:     SubstituteExpr(e.Left, varName, replacement),
			Right:    SubstituteExpr(e.Right, varName, replacement),
		}
	case *ast.UnaryExpr:
		return &ast.UnaryExpr{
			Position: e.Position,
			Op:       e.Op,
			Expr:     SubstituteExpr(e.Expr, varName, replacement),
		}
	case *ast.ParenExpr:
		return &ast.ParenExpr{
			Position: e.Position,
			X:        SubstituteExpr(e.X, varName, replacement),
		}
	case *ast.SelectorExpr:
		return &ast.SelectorExpr{
			Position: e.Position,
			X:        SubstituteExpr(e.X, varName, replacement),
			Sel:      e.Sel,
		}
	case *ast.CallExpr:
		args := make([]ast.Expr, len(e.Args))
		for i, a := range e.Args {
			args[i] = SubstituteExpr(a, varName, replacement)
		}
		return &ast.CallExpr{
			Position: e.Position,
			Fun:      SubstituteExpr(e.Fun, varName, replacement),
			Args:     args,
		}
	case *ast.IndexExpr:
		return &ast.IndexExpr{
			Position: e.Position,
			X:        SubstituteExpr(e.X, varName, replacement),
			Index:    SubstituteExpr(e.Index, varName, replacement),
		}
	default:
		return expr
	}
}

// PrintExpr renders an AST expression back to Spectre source.
// Exported so commands.go can display the synthesized guard.
func PrintExpr(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Prime {
			return e.Name + "'"
		}
		return e.Name
	case *ast.BasicLit:
		return e.Value
	case *ast.BinaryExpr:
		left := PrintExpr(e.Left)
		right := PrintExpr(e.Right)
		if needsParens(e.Left, e.Op) {
			left = "(" + left + ")"
		}
		if needsParens(e.Right, e.Op) {
			right = "(" + right + ")"
		}
		return left + " " + binOpStr(e.Op) + " " + right
	case *ast.UnaryExpr:
		inner := PrintExpr(e.Expr)
		if _, isBin := e.Expr.(*ast.BinaryExpr); isBin {
			inner = "(" + inner + ")"
		}
		if e.Op == ast.Not {
			return "!" + inner
		}
		return "-" + inner
	case *ast.ParenExpr:
		return "(" + PrintExpr(e.X) + ")"
	case *ast.SelectorExpr:
		return PrintExpr(e.X) + "." + e.Sel
	case *ast.CallExpr:
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = PrintExpr(a)
		}
		return PrintExpr(e.Fun) + "(" + strings.Join(args, ", ") + ")"
	case *ast.IndexExpr:
		return PrintExpr(e.X) + "[" + PrintExpr(e.Index) + "]"
	default:
		return "<expr>"
	}
}

// needsParens reports whether a sub-expression needs parentheses when used under op.
// Precedence (low to high): || < && < comparisons < arithmetic.
func needsParens(sub ast.Expr, op ast.BinaryOp) bool {
	bin, ok := sub.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	switch op {
	case ast.And:
		// || has lower precedence than &&, so (a || b) && c needs parens.
		return bin.Op == ast.Or
	case ast.Or:
		// && has higher precedence than ||; no parens needed.
		return false
	case ast.Mul, ast.Div:
		// +/- has lower precedence than */÷.
		return bin.Op == ast.Add || bin.Op == ast.Sub
	}
	return false
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
		return "="
	case ast.Neq:
		return "!="
	case ast.Lt:
		return "<"
	case ast.Gt:
		return ">"
	case ast.Leq:
		return "<="
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

// RootCauseDesc builds a human-readable description of the root cause.
func RootCauseDesc(guard SynthGuard, preState *state.State, args []state.Value) string {
	if preState == nil {
		return fmt.Sprintf("`%s' = %s` violated the invariant", guard.VarName, guard.AssignRHS)
	}

	preVal := "<unknown>"
	if v, ok := preState.Variables[guard.VarName]; ok {
		preVal = v.String()
	}

	argDesc := ""
	if len(args) > 0 {
		argStrs := make([]string, len(args))
		for i, a := range args {
			argStrs[i] = a.String()
		}
		argDesc = " (args: " + strings.Join(argStrs, ", ") + ")"
	}

	return fmt.Sprintf(
		"`%s' = %s`%s — with %s = %s, the result violated the invariant",
		guard.VarName, guard.AssignRHS, argDesc, guard.VarName, preVal,
	)
}

// AccumulateGuards merges guards for the same (InvariantName, ActionName) pair
// across multiple CEGISRepair results, producing conjunctions of all distinct
// WP conditions found. This supports multi-counterexample guard strengthening.
func AccumulateGuards(repairs []CEGISRepair) []CEGISRepair {
	type key struct{ inv, action string }
	merged := make(map[key]*CEGISRepair)
	order := []key{}

	for i := range repairs {
		r := &repairs[i]
		k := key{r.InvariantName, r.ActionName}
		if existing, ok := merged[k]; ok {
			// Merge guards: add any guard string not already present.
			seenGuards := make(map[string]bool)
			for _, g := range existing.Guards {
				seenGuards[g.WPString] = true
			}
			for _, g := range r.Guards {
				if !seenGuards[g.WPString] {
					seenGuards[g.WPString] = true
					existing.Guards = append(existing.Guards, g)
				}
			}
		} else {
			cp := *r
			merged[k] = &cp
			order = append(order, k)
		}
	}

	result := make([]CEGISRepair, 0, len(order))
	for _, k := range order {
		result = append(result, *merged[k])
	}
	return result
}

// extractField extracts the text between prefix and suffix in s, returning "" if not found.
func extractField(s, prefix, suffix string) string {
	i := strings.Index(s, prefix)
	if i < 0 {
		return ""
	}
	rest := s[i+len(prefix):]
	j := strings.Index(rest, suffix)
	if j <= 0 {
		return ""
	}
	return rest[:j]
}


// CheckGuardMinimality verifies that a synthesized guard does not block any
// transition in the provided positive ITF trace (encoded as a raw JSON state array).
// Returns a list of step indices (1-based) that would be blocked by the guard.
func CheckGuardMinimality(guard SynthGuard, actionName string, trace []map[string]interface{}, file *ast.File) []int {
	var blocked []int
	for stepIdx := 1; stepIdx < len(trace); stepIdx++ {
		meta, ok := trace[stepIdx]["#meta"].(map[string]interface{})
		if !ok {
			continue
		}
		act, _ := meta["action"].(string)
		if act != actionName {
			continue
		}
		// Evaluate the guard in the pre-state (step before this one).
		preState := rawMapToState(trace[stepIdx-1])
		if !evalBoolInState(guard.WPExpr, preState, file, nil, nil) {
			blocked = append(blocked, stepIdx)
		}
	}
	return blocked
}

// rawMapToState converts a raw ITF state map to a state.State for evaluation.
func rawMapToState(m map[string]interface{}) *state.State {
	vars := make(map[string]state.Value)
	for k, v := range m {
		if k == "#meta" {
			continue
		}
		switch val := v.(type) {
		case float64:
			i := int64(val)
			vars[k] = &state.PrimitiveValue{TypeName: "int", IntValue: &i}
		case bool:
			b := val
			vars[k] = &state.PrimitiveValue{TypeName: "bool", BoolValue: &b}
		case string:
			s := val
			vars[k] = &state.PrimitiveValue{TypeName: "str", StringValue: &s}
		case map[string]interface{}:
			// ITF encodes ints as {"#bigint": "N"}
			if bigint, ok2 := val["#bigint"].(string); ok2 {
				n := int64(0)
				fmt.Sscanf(bigint, "%d", &n)
				vars[k] = &state.PrimitiveValue{TypeName: "int", IntValue: &n}
			}
		}
	}
	return &state.State{Variables: vars}
}
