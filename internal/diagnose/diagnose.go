package diagnose

import (
	"fmt"
	"sort"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/eval"
	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// StutterRepair holds diagnosis and repair suggestions for one distinct stuttering pattern.
type StutterRepair struct {
	ActionName   string
	ActionArgs   []state.Value
	StutterState *state.State
	NoOpSummary  string
	Options      []RepairOption
}

// RepairOption is one concrete suggestion for eliminating the stutter.
type RepairOption struct {
	Kind        string // "guard", "fairness", or "redesign"
	Title       string
	Explanation string
	SpectreCode string
}

// AnalyzeStuttering analyzes stuttering steps from state-space exploration and produces
// repair suggestions, deduplicating across equivalent (action + state) patterns.
func AnalyzeStuttering(
	stuttering []*explore.Stuttering,
	file *ast.File,
	sm *exec.StateMachine,
) []StutterRepair {
	seen := make(map[string]bool)
	repairs := make([]StutterRepair, 0)
	for _, s := range stuttering {
		key := stutKey(s)
		if seen[key] {
			continue
		}
		seen[key] = true
		repairs = append(repairs, analyzeOne(s, file, sm))
	}
	return repairs
}

func stutKey(s *explore.Stuttering) string {
	parts := []string{s.Action}
	for _, arg := range s.Args {
		parts = append(parts, arg.String())
	}
	varKeys := make([]string, 0, len(s.State.Variables))
	for k, v := range s.State.Variables {
		varKeys = append(varKeys, k+"="+v.String())
	}
	sort.Strings(varKeys)
	parts = append(parts, varKeys...)
	return strings.Join(parts, ":")
}

func analyzeOne(s *explore.Stuttering, file *ast.File, sm *exec.StateMachine) StutterRepair {
	decl := findActionDecl(file, s.Action)
	noOps := findNoOpAssignments(s, decl, file)
	return StutterRepair{
		ActionName:   s.Action,
		ActionArgs:   s.Args,
		StutterState: s.State,
		NoOpSummary:  buildNoOpSummary(s, noOps),
		Options:      buildRepairOptions(s, noOps, sm),
	}
}

// findActionDecl finds an action declaration by name, including inside modules.
func findActionDecl(file *ast.File, name string) *ast.ActionDecl {
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok && a.Name == name {
			return a
		}
		if m, ok := d.(*ast.ModuleDecl); ok {
			for _, md := range m.Decls {
				if a, ok := md.(*ast.ActionDecl); ok && a.Name == name {
					return a
				}
			}
		}
	}
	return nil
}

type noOpInfo struct {
	varName    string
	currentVal state.Value
	guardHint  string // empty string means identity assignment — needs redesign
}

// findNoOpAssignments evaluates each primed assignment in the action body against the
// stutter state and returns those that produce no change.
func findNoOpAssignments(s *explore.Stuttering, decl *ast.ActionDecl, file *ast.File) []noOpInfo {
	if decl == nil || decl.Body == nil {
		return nil
	}

	env := eval.NewEnvironment()
	eval.RegisterEnumTypes(env, file)
	for varName, val := range s.State.Variables {
		env.SetVariable(varName, val)
	}
	for i, param := range decl.Parameters {
		if i < len(s.Args) {
			env.SetVariable(param.Name, s.Args[i])
		}
	}
	evaluator := eval.NewEvaluator(env)

	var result []noOpInfo
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

		currentVal, exists := s.State.Variables[varName]
		if !exists {
			continue
		}

		rhsVal, err := evaluator.Eval(assign.Right)
		if err != nil || rhsVal.String() != currentVal.String() {
			continue
		}

		hint := guardHintFor(varName, currentVal, assign.Right, decl.Parameters, s.Args)
		result = append(result, noOpInfo{
			varName:    varName,
			currentVal: currentVal,
			guardHint:  hint,
		})
	}
	return result
}

// guardHintFor returns a require expression that would prevent the no-op.
// Returns "" for identity assignments (var' = var) which need redesign instead of a guard.
func guardHintFor(
	varName string,
	currentVal state.Value,
	rhs ast.Expr,
	params []ast.Parameter,
	args []state.Value,
) string {
	// Identity assignment (var' = var) — no guard can help.
	if ident, ok := rhs.(*ast.Ident); ok && !ident.Prime && ident.Name == varName {
		return ""
	}

	// For non-literal RHS, check if a zero-valued parameter is causing the no-op.
	if _, isLit := rhs.(*ast.BasicLit); !isLit {
		for i, param := range params {
			if i >= len(args) {
				break
			}
			if hint := zeroParamGuard(param.Name, args[i]); hint != "" {
				return hint
			}
		}
	}

	return negateValue(varName, currentVal)
}

func zeroParamGuard(paramName string, argVal state.Value) string {
	prim, ok := argVal.(*state.PrimitiveValue)
	if !ok {
		return ""
	}
	switch prim.TypeName {
	case "int":
		if prim.IntValue != nil && *prim.IntValue == 0 {
			return fmt.Sprintf("%s > 0", paramName)
		}
	case "bool":
		if prim.BoolValue != nil && !*prim.BoolValue {
			return paramName
		}
	case "str":
		if prim.StringValue != nil && *prim.StringValue == "" {
			return fmt.Sprintf(`%s != ""`, paramName)
		}
	}
	return ""
}

func negateValue(varName string, val state.Value) string {
	if prim, ok := val.(*state.PrimitiveValue); ok {
		switch prim.TypeName {
		case "int":
			if prim.IntValue != nil && *prim.IntValue == 0 {
				return fmt.Sprintf("%s > 0", varName)
			}
			return fmt.Sprintf("%s != %s", varName, val.String())
		case "bool":
			if prim.BoolValue != nil && *prim.BoolValue {
				return fmt.Sprintf("!%s", varName)
			}
			return varName
		}
	}
	return fmt.Sprintf("%s != %s", varName, val.String())
}

func buildNoOpSummary(s *explore.Stuttering, noOps []noOpInfo) string {
	stateDesc := stateCompact(s.State)
	if len(noOps) == 0 {
		return fmt.Sprintf("%s makes no state change in %s", actionCall(s), stateDesc)
	}
	parts := make([]string, len(noOps))
	for i, n := range noOps {
		parts[i] = fmt.Sprintf("%s' = %s (already %s)", n.varName, n.currentVal.String(), n.currentVal.String())
	}
	return fmt.Sprintf("in %s, %s", stateDesc, strings.Join(parts, "; "))
}

func actionCall(s *explore.Stuttering) string {
	if len(s.Args) == 0 {
		return fmt.Sprintf("action `%s`", s.Action)
	}
	argStr := make([]string, len(s.Args))
	for i, a := range s.Args {
		argStr[i] = a.String()
	}
	return fmt.Sprintf("action `%s(%s)`", s.Action, strings.Join(argStr, ", "))
}

func stateCompact(s *state.State) string {
	pairs := make([]string, 0, len(s.Variables))
	for k, v := range s.Variables {
		pairs = append(pairs, fmt.Sprintf("%s = %s", k, v.String()))
	}
	sort.Strings(pairs)
	return "{" + strings.Join(pairs, ", ") + "}"
}

func buildRepairOptions(s *explore.Stuttering, noOps []noOpInfo, sm *exec.StateMachine) []RepairOption {
	var opts []RepairOption

	for _, noOp := range noOps {
		if noOp.guardHint == "" {
			opts = append(opts, RepairOption{
				Kind:  "redesign",
				Title: fmt.Sprintf("Redesign action `%s` (identity assignment)", s.Action),
				Explanation: fmt.Sprintf(
					"`%s' = %s` always produces the same value — this action can never make progress.",
					noOp.varName, noOp.varName,
				),
				SpectreCode: fmt.Sprintf(
					"// In action %s:\n// Replace `%s' = %s` with an assignment that changes the state.",
					s.Action, noOp.varName, noOp.varName,
				),
			})
		} else {
			opts = append(opts, RepairOption{
				Kind:  "guard",
				Title: fmt.Sprintf("Add a guard to action `%s`", s.Action),
				Explanation: fmt.Sprintf(
					"Prevent `%s` from firing when it would be a no-op (%s = %s).",
					s.Action, noOp.varName, noOp.currentVal.String(),
				),
				SpectreCode: fmt.Sprintf("// In action %s — add:\nrequire %s", s.Action, noOp.guardHint),
			})
		}
	}

	progressActions := findProgressActions(s, sm)
	if len(progressActions) > 3 {
		progressActions = progressActions[:3]
	}
	for _, progressAction := range progressActions {
		opts = append(opts, RepairOption{
			Kind:  "fairness",
			Title: fmt.Sprintf("Add weak fairness on `%s`", progressAction),
			Explanation: fmt.Sprintf(
				"Guarantees `%s` is eventually taken whenever continuously enabled, "+
					"breaking the stutter cycle and ensuring liveness.",
				progressAction,
			),
			SpectreCode: fmt.Sprintf("temporal stutter_fix {\n  WF(%s)\n}", progressAction),
		})
	}

	if len(opts) == 0 {
		opts = append(opts, RepairOption{
			Kind:        "guard",
			Title:       fmt.Sprintf("Add a precondition to action `%s`", s.Action),
			Explanation: "The action fires but produces no change. Add a `require` to prevent this.",
			SpectreCode: fmt.Sprintf("// In action %s — add:\nrequire <condition that ensures state changes>", s.Action),
		})
	}
	return opts
}

// findProgressActions returns actions (other than the stuttering action) that are
// enabled in the stutter state and would produce a different next state.
func findProgressActions(s *explore.Stuttering, sm *exec.StateMachine) []string {
	if sm == nil {
		return nil
	}
	available, err := sm.GetAvailableActions(s.State)
	if err != nil {
		return nil
	}
	seen := make(map[string]bool)
	var progress []string
	for _, name := range available {
		if name == s.Action || seen[name] {
			continue
		}
		seen[name] = true
		nextState, err := sm.ExecuteAction(name, s.State, nil)
		if err != nil {
			continue // parameterized actions with nil args may fail — skip
		}
		if !nextState.Equals(s.State) {
			progress = append(progress, name)
		}
	}
	return progress
}
