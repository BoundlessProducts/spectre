package codegen

import (
	"fmt"
	"strconv"

	"github.com/akkeshavan/spectre/internal/eval"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// RefinementReport describes how well an ITF trace set conforms to the spec
// via the driver.json mapping.
type RefinementReport struct {
	SpecFile           string
	StructName         string
	CheckedTransitions int
	LegalSteps         int
	StutterSteps       int
	Violations         []RefinementViolation
}

// RefinementViolation is one transition that couldn't be classified as legal or stutter.
type RefinementViolation struct {
	TraceIndex int
	StepIndex  int
	Action     string
	PreState   map[string]interface{}
	PostState  map[string]interface{}
	Reason     string
}

// CheckRefinement evaluates traces against the spec + driver.json mapping.
//
// Each transition (preState, action, postState) is classified as:
//   - LegalStep    — action maps to a spec action and all require guards hold
//   - StutterStep  — action not in spec but all spec-relevant fields are unchanged
//   - Violation    — neither of the above
//
// traces is a list of raw ITF state arrays (one per trace file); each element
// is the decoded JSON array from the "states" key.
func CheckRefinement(file *ast.File, meta *DriverMeta, traces [][]map[string]interface{}) *RefinementReport {
	report := &RefinementReport{
		SpecFile:   "",
		StructName: meta.StructName,
	}

	// Build action → require-clause map from the spec AST.
	requireMap := buildRequireMap(file)

	// Build a set of spec action names for quick lookup.
	specActions := make(map[string]bool)
	for _, ma := range meta.Actions {
		specActions[ma.SpecAction] = true
	}

	// Build set of spec-variable names for stutter detection.
	specVarNames := make(map[string]bool)
	for _, mf := range meta.Fields {
		specVarNames[mf.SpecVar] = true
	}

	for traceIdx, trace := range traces {
		for stepIdx := 1; stepIdx < len(trace); stepIdx++ {
			preRaw := trace[stepIdx-1]
			postRaw := trace[stepIdx]

			action := extractAction(postRaw)
			if action == "" {
				continue // init step, no action
			}

			report.CheckedTransitions++

			preState := rawToState(preRaw)
			postState := rawToState(postRaw)

			if specActions[action] {
				// Check all require clauses.
				clauses := requireMap[action]
				violated := ""
				for _, clause := range clauses {
					if !evalBoolState(clause, preState, file) {
						violated = fmt.Sprintf("require clause not satisfied in pre-state")
						break
					}
				}
				if violated == "" {
					report.LegalSteps++
				} else {
					report.Violations = append(report.Violations, RefinementViolation{
						TraceIndex: traceIdx,
						StepIndex:  stepIdx,
						Action:     action,
						PreState:   stateToRawMap(preState),
						PostState:  stateToRawMap(postState),
						Reason:     violated,
					})
				}
			} else {
				// Check stutter: all spec variables unchanged.
				isStutter := true
				for varName := range specVarNames {
					preVal, preOk := preState.Variables[varName]
					postVal, postOk := postState.Variables[varName]
					if preOk != postOk {
						isStutter = false
						break
					}
					if preOk && postOk && preVal.String() != postVal.String() {
						isStutter = false
						break
					}
				}
				if isStutter {
					report.StutterSteps++
				} else {
					report.Violations = append(report.Violations, RefinementViolation{
						TraceIndex: traceIdx,
						StepIndex:  stepIdx,
						Action:     action,
						PreState:   stateToRawMap(preState),
						PostState:  stateToRawMap(postState),
						Reason:     fmt.Sprintf("action %q not in spec and not a stutter step", action),
					})
				}
			}
		}
	}

	return report
}

// ---- helpers ----------------------------------------------------------------

func buildRequireMap(file *ast.File) map[string][]ast.Expr {
	m := make(map[string][]ast.Expr)
	var collect func(decl ast.Decl)
	collect = func(decl ast.Decl) {
		switch d := decl.(type) {
		case *ast.ActionDecl:
			if d.Body != nil {
				for _, stmt := range d.Body.Statements {
					if req, ok := stmt.(*ast.RequireStmt); ok {
						m[d.Name] = append(m[d.Name], req.Condition)
					}
				}
			}
			if d.Guard != nil {
				m[d.Name] = append(m[d.Name], d.Guard)
			}
		case *ast.ModuleDecl:
			for _, md := range d.Decls {
				collect(md)
			}
		}
	}
	for _, decl := range file.Decls {
		collect(decl)
	}
	return m
}

func extractAction(stateMap map[string]interface{}) string {
	meta, ok := stateMap["#meta"].(map[string]interface{})
	if !ok {
		return ""
	}
	action, _ := meta["action"].(string)
	return action
}

func rawToState(m map[string]interface{}) *state.State {
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
			if bigint, ok := val["#bigint"].(string); ok {
				n, err := strconv.ParseInt(bigint, 10, 64)
				if err == nil {
					vars[k] = &state.PrimitiveValue{TypeName: "int", IntValue: &n}
				}
			}
		}
	}
	return &state.State{Variables: vars}
}

func stateToRawMap(s *state.State) map[string]interface{} {
	m := make(map[string]interface{}, len(s.Variables))
	for k, v := range s.Variables {
		m[k] = v.String()
	}
	return m
}

func evalBoolState(expr ast.Expr, s *state.State, file *ast.File) bool {
	env := eval.NewEnvironment()
	eval.RegisterEnumTypes(env, file)
	for k, v := range s.Variables {
		env.SetVariable(k, v)
	}
	result, err := eval.NewEvaluator(env).Eval(expr)
	if err != nil {
		return false
	}
	pv, ok := result.(*state.PrimitiveValue)
	if !ok || pv.TypeName != "bool" || pv.BoolValue == nil {
		return false
	}
	return *pv.BoolValue
}
