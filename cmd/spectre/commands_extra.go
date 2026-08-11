package main

// commands_extra.go adds the sync, drift, diff-conformance, and simulate commands.

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	osexec "os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/akkeshavan/spectre/internal/codegen"
	"github.com/akkeshavan/spectre/internal/diagnose"
	"github.com/akkeshavan/spectre/internal/eval"
	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/explore"
	"github.com/akkeshavan/spectre/internal/itf"
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/mine"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

func init() {
	Commands["sync"] = &Command{
		Name:        "sync",
		Description: "Diff Rust source against an existing spec and classify changes",
		Run:         runSync,
	}
	Commands["drift"] = &Command{
		Name:        "drift",
		Description: "Detect bidirectional drift between a spec and its Rust implementation sidecar",
		Run:         runDrift,
	}
	Commands["diff-conformance"] = &Command{
		Name:        "diff-conformance",
		Description: "Replay an ITF trace against two spec versions and report divergences",
		Run:         runDiffConformance,
	}
	Commands["simulate"] = &Command{
		Name:        "simulate",
		Description: "Generate random execution traces by random-walking the explored state graph",
		Run:         runSimulate,
	}
	Commands["impact"] = &Command{
		Name:        "impact",
		Description: "Given a git diff, report affected spec actions/invariants and incremental verification status",
		Run:         runImpact,
	}
}

// ── helpers shared across new commands ───────────────────────────────────────

// parseSpecFile reads and parses a .spec file, returning the AST.
func parseSpecFile(filename string) (*ast.File, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}
	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors in %s: %s", filename, strings.Join(p.Errors(), "; "))
	}
	return file, nil
}

// specVars extracts variable names and their spec-type strings from a parsed spec.
func specVars(file *ast.File) map[string]string {
	m := make(map[string]string)
	for _, d := range file.Decls {
		if v, ok := d.(*ast.VariableDecl); ok {
			m[v.Name] = fmt.Sprintf("%v", v.Type)
		}
	}
	return m
}

// specActions extracts action names → require-clause count from a parsed spec.
func specActions(file *ast.File) map[string]int {
	m := make(map[string]int)
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok {
			count := 0
			if a.Body != nil {
				for _, s := range a.Body.Statements {
					if _, ok := s.(*ast.RequireStmt); ok {
						count++
					}
				}
			}
			m[a.Name] = count
		}
	}
	return m
}

// specActionAssignCount extracts action names → assignment count from a parsed spec.
func specActionAssignCount(file *ast.File) map[string]int {
	m := make(map[string]int)
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok {
			count := 0
			if a.Body != nil {
				for _, s := range a.Body.Statements {
					if _, ok := s.(*ast.AssignStmt); ok {
						count++
					}
				}
			}
			m[a.Name] = count
		}
	}
	return m
}

// specActionGuards extracts action names → sorted slice of require-expression strings.
func specActionGuards(file *ast.File) map[string][]string {
	m := make(map[string][]string)
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok {
			var guards []string
			if a.Body != nil {
				for _, s := range a.Body.Statements {
					if req, ok := s.(*ast.RequireStmt); ok {
						guards = append(guards, diagnose.PrintExpr(req.Condition))
					}
				}
			}
			m[a.Name] = guards
		}
	}
	return m
}

// specActionGuardExprs extracts action names → slice of guard AST expressions.
// Used by runSync for SMT-backed equivalence checking.
func specActionGuardExprs(file *ast.File) map[string][]ast.Expr {
	m := make(map[string][]ast.Expr)
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok {
			var guards []ast.Expr
			if a.Body != nil {
				for _, s := range a.Body.Statements {
					if req, ok := s.(*ast.RequireStmt); ok {
						guards = append(guards, req.Condition)
					}
				}
			}
			m[a.Name] = guards
		}
	}
	return m
}

// specActionAssignRHSExprs extracts action names → slice of assignment RHS AST expressions.
func specActionAssignRHSExprs(file *ast.File) map[string][]ast.Expr {
	m := make(map[string][]ast.Expr)
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok {
			var exprs []ast.Expr
			if a.Body != nil {
				for _, s := range a.Body.Statements {
					if assign, ok := s.(*ast.AssignStmt); ok {
						exprs = append(exprs, assign.Right)
					}
				}
			}
			m[a.Name] = exprs
		}
	}
	return m
}

// parseSpecString parses a spec from an in-memory string.
func parseSpecString(src string) (*ast.File, error) {
	l := lexer.New(src)
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors: %s", strings.Join(p.Errors(), "; "))
	}
	return file, nil
}

// specActionAssignExprs extracts action names → slice of "lhs=rhs" canonical strings.
// Used by runSync to detect semantic changes in assignment bodies (e.g. wrong-arith mutations).
func specActionAssignExprs(file *ast.File) map[string][]string {
	m := make(map[string][]string)
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok {
			var exprs []string
			if a.Body != nil {
				for _, s := range a.Body.Statements {
					if assign, ok := s.(*ast.AssignStmt); ok {
						lhs := diagnose.PrintExpr(assign.Left)
						rhs := diagnose.PrintExpr(assign.Right)
						exprs = append(exprs, lhs+"="+rhs)
					}
				}
			}
			m[a.Name] = exprs
		}
	}
	return m
}

// specInvariants returns invariant names from a parsed spec.
func specInvariants(file *ast.File) []string {
	var names []string
	for _, d := range file.Decls {
		if inv, ok := d.(*ast.InvariantDecl); ok {
			names = append(names, inv.Name)
		}
	}
	return names
}

// ── Feature 2: spectre sync ───────────────────────────────────────────────────

func runSync(args []string) error {
	specFile := ""
	filteredArgs := []string{}
	i := 0
	for i < len(args) {
		switch args[i] {
		case "--spec":
			if i+1 < len(args) {
				specFile = args[i+1]
				i++
			}
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}
	if len(filteredArgs) == 0 {
		return fmt.Errorf("usage: spectre sync <rust-source.rs> [--spec <existing.spec>]")
	}
	sourceFile := filteredArgs[0]

	// Determine existing spec path.
	if specFile == "" {
		base := strings.TrimSuffix(filepath.Base(sourceFile), ".rs")
		specFile = filepath.Join(filepath.Dir(sourceFile), base+".spec")
	}
	if _, err := os.Stat(specFile); os.IsNotExist(err) {
		fmt.Printf("No existing spec found at %s — run `spectre mine` first.\n", specFile)
		return nil
	}

	// Mine fresh spec.
	specName := strings.TrimSuffix(filepath.Base(specFile), ".spec")
	mined, err := mine.MineFromRustFile(sourceFile, specName)
	if err != nil {
		return fmt.Errorf("mining failed: %w", err)
	}

	// Parse existing spec.
	oldFile, err := parseSpecFile(specFile)
	if err != nil {
		return fmt.Errorf("could not parse existing spec: %w", err)
	}
	oldVars := specVars(oldFile)
	oldActions := specActions(oldFile)
	oldGuards := specActionGuards(oldFile)

	// Build new maps from mined result.
	newVars := make(map[string]string)
	for _, f := range mined.Fields {
		newVars[f.Name] = f.SpectreType
	}
	newActions := make(map[string]int)
	newGuards := make(map[string][]string)
	newAssignCount := make(map[string]int) // assignment count per action
	for _, m := range mined.Methods {
		if m.MutSelf {
			newActions[m.Name] = len(m.Requires)
			newGuards[m.Name] = m.Requires
			newAssignCount[m.Name] = len(m.Assigns)
		}
	}
	oldAssignCount := specActionAssignCount(oldFile)

	// Generate fresh spec text and parse it for expression-level comparison.
	freshText := mined.Generate(nil)
	newFile, freshErr := parseSpecString(freshText)
	var oldAssignExprs, newAssignExprs map[string][]string
	var oldGuardExprs, newGuardExprs map[string][]ast.Expr
	var oldAssignRHSExprs, newAssignRHSExprs map[string][]ast.Expr
	if freshErr == nil && newFile != nil {
		oldAssignExprs = specActionAssignExprs(oldFile)
		newAssignExprs = specActionAssignExprs(newFile)
		// AST expressions for SMT-backed semantic equivalence checking.
		oldGuardExprs = specActionGuardExprs(oldFile)
		newGuardExprs = specActionGuardExprs(newFile)
		oldAssignRHSExprs = specActionAssignRHSExprs(oldFile)
		newAssignRHSExprs = specActionAssignRHSExprs(newFile)
	}

	fmt.Printf("Syncing %s against %s\n\n", sourceFile, specFile)

	nPreserving, nUpdate, nUnsafe := 0, 0, 0

	// Var comparison.
	fmt.Println("Variables:")
	for name, typ := range newVars {
		if _, exists := oldVars[name]; !exists {
			fmt.Printf("  [+] var %s: %s (new field — spec update required)\n", name, typ)
			nUpdate++
		}
	}
	for name := range oldVars {
		if _, exists := newVars[name]; !exists {
			fmt.Printf("  [-] var %s (removed field — possibly unsafe)\n", name)
			nUnsafe++
		}
	}

	// Action comparison: check structure (count) then semantics (expressions).
	fmt.Println("Actions:")
	for name, newCount := range newActions {
		oldCount, exists := oldActions[name]
		if !exists {
			fmt.Printf("  [+] action %s (new action — spec update required)\n", name)
			nUpdate++
			continue
		}
		if oldCount != newCount {
			fmt.Printf("  [~] action %s (guard count changed: was %d, now %d — spec update required)\n", name, oldCount, newCount)
			nUpdate++
			continue
		}
		// Same guard count — compare guard expressions, using SMT for semantic precision.
		oldG, newG := oldGuards[name], newGuards[name]
		semanticChange := false
		for i := 0; i < len(oldG) && i < len(newG); i++ {
			if oldG[i] == newG[i] {
				continue
			}
			// Strings differ — attempt SMT equivalence check.
			smtResult, witness := smtCheckGuards(oldGuardExprs, newGuardExprs, name, i)
			switch smtResult {
			case diagnose.SMTEquivalent:
				fmt.Printf("  [≡] action %s guard %d rewritten (SMT-proved equivalent): %q → %q\n",
					name, i+1, oldG[i], newG[i])
			case diagnose.SMTChanged:
				msg := fmt.Sprintf("  [✗] action %s guard %d changed semantically: %q → %q", name, i+1, oldG[i], newG[i])
				if witness != "" {
					msg += fmt.Sprintf(" (witness: %s)", witness)
				}
				fmt.Println(msg)
				semanticChange = true
			default:
				fmt.Printf("  [?] action %s guard %d changed: %q → %q (possibly unsafe — SMT inconclusive)\n",
					name, i+1, oldG[i], newG[i])
				semanticChange = true
			}
		}
		// Assignment count change (new/removed assignments) is structurally significant.
		if oldAssignCount[name] != newAssignCount[name] {
			fmt.Printf("  [?] action %s assignment count changed: was %d, now %d (possibly unsafe)\n",
				name, oldAssignCount[name], newAssignCount[name])
			semanticChange = true
		}
		// Compare assignment RHS expressions with SMT when count is unchanged.
		if oldAssignExprs != nil && oldAssignCount[name] == newAssignCount[name] {
			for i, newExpr := range newAssignExprs[name] {
				if i >= len(oldAssignExprs[name]) || oldAssignExprs[name][i] == newExpr {
					continue
				}
				smtResult, witness := smtCheckAssigns(oldAssignRHSExprs, newAssignRHSExprs, name, i)
				switch smtResult {
				case diagnose.SMTEquivalent:
					fmt.Printf("  [≡] action %s assignment %d rewritten (SMT-proved equivalent)\n", name, i+1)
				case diagnose.SMTChanged:
					msg := fmt.Sprintf("  [✗] action %s assignment %d body changed: %q → %q", name, i+1, oldAssignExprs[name][i], newExpr)
					if witness != "" {
						msg += fmt.Sprintf(" (witness: %s)", witness)
					}
					fmt.Println(msg)
					semanticChange = true
				default:
					fmt.Printf("  [?] action %s assignment %d body changed: %q → %q (possibly unsafe)\n",
						name, i+1, oldAssignExprs[name][i], newExpr)
					semanticChange = true
				}
			}
		}
		if semanticChange {
			nUnsafe++
		} else {
			fmt.Printf("  [=] action %s (structurally unchanged)\n", name)
			nPreserving++
		}
	}
	for name := range oldActions {
		if _, exists := newActions[name]; !exists {
			fmt.Printf("  [-] action %s (removed action — possibly unsafe)\n", name)
			nUnsafe++
		}
	}

	fmt.Printf("\nSummary: %d model-preserving, %d spec-update-required, %d possibly-unsafe\n",
		nPreserving, nUpdate, nUnsafe)

	if nUpdate > 0 || nUnsafe > 0 {
		return fmt.Errorf("sync detected %d change(s) requiring attention", nUpdate+nUnsafe)
	}
	return nil
}

// ── Feature 6: spectre drift ──────────────────────────────────────────────────

func runDrift(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: spectre drift <spec-file.spec>")
	}
	input := args[0]

	// For simplicity, require a .spec file; load its .driver.json sidecar.
	if !strings.HasSuffix(input, ".spec") {
		return fmt.Errorf("drift expects a .spec file (got %s)", input)
	}

	metaPath := codegen.MetaPath(input)
	meta, err := codegen.LoadDriverMeta(metaPath)
	if err != nil {
		return fmt.Errorf("error reading sidecar: %w", err)
	}
	if meta == nil {
		fmt.Printf("No .driver.json sidecar found at %s.\n", metaPath)
		fmt.Println("Run `spectre mine -o <spec>` to create one.")
		return nil
	}

	// Parse the spec.
	specFile, err := parseSpecFile(input)
	if err != nil {
		return err
	}
	specVarMap := specVars(specFile)
	specActionMap := specActions(specFile)
	invNames := specInvariants(specFile)

	fmt.Printf("Drift report for %s  ↔  %s\n\n", input, metaPath)

	nRustStale, nImplStale, nInvImpact := 0, 0, 0

	// Rust → Spec direction.
	fmt.Println("Rust → Spec (sidecar fields/actions missing from spec):")
	for _, f := range meta.Fields {
		if _, ok := specVarMap[f.SpecVar]; !ok {
			fmt.Printf("  [!] Rust field '%s' has no matching spec variable '%s' — spec is stale\n",
				f.RustField, f.SpecVar)
			nRustStale++
		}
	}
	for _, a := range meta.Actions {
		if _, ok := specActionMap[a.SpecAction]; !ok {
			fmt.Printf("  [!] Rust method '%s' has no matching spec action '%s' — spec is stale\n",
				a.RustMethod, a.SpecAction)
			nRustStale++
		}
	}
	if nRustStale == 0 {
		fmt.Println("  (none)")
	}

	// Spec → Rust direction.
	fmt.Println("\nSpec → Rust (spec vars/actions missing from sidecar):")
	sidecarVars := make(map[string]bool)
	for _, f := range meta.Fields {
		sidecarVars[f.SpecVar] = true
	}
	for varName := range specVarMap {
		if !sidecarVars[varName] {
			fmt.Printf("  [?] Spec variable '%s' has no Rust implementation mapping — implementation may be stale\n", varName)
			nImplStale++
		}
	}
	sidecarActions := make(map[string]bool)
	for _, a := range meta.Actions {
		sidecarActions[a.SpecAction] = true
	}
	for actionName := range specActionMap {
		if !sidecarActions[actionName] {
			fmt.Printf("  [?] Spec action '%s' has no Rust method mapping — implementation may be stale\n", actionName)
			nImplStale++
		}
	}
	if nImplStale == 0 {
		fmt.Println("  (none)")
	}

	// Invariant impact.
	fmt.Println("\nInvariant monitor impact:")
	for _, name := range invNames {
		fmt.Printf("  [i] Invariant '%s' may need monitor regeneration if spec changed\n", name)
		nInvImpact++
	}
	if nInvImpact == 0 {
		fmt.Println("  (no invariants found)")
	}

	fmt.Printf("\nSummary: %d spec-stale, %d impl-stale, %d invariants to re-check\n",
		nRustStale, nImplStale, nInvImpact)
	if nRustStale > 0 || nImplStale > 0 {
		fmt.Println("Next steps:")
		if nRustStale > 0 {
			fmt.Println("  • Run `spectre mine -o <spec>` to refresh the spec and sidecar from source.")
		}
		if nImplStale > 0 {
			fmt.Println("  • Run `spectre generate-driver <spec>` to regenerate the MBT adapter.")
		}
		if nInvImpact > 0 {
			fmt.Println("  • Run `spectre generate-monitor <spec>` to rebuild the runtime monitor.")
		}
	}
	return nil
}

// ── Feature 8: spectre diff-conformance ──────────────────────────────────────

// itfTraceJSON is the top-level structure of an ITF JSON file.
type itfTraceJSON struct {
	Vars   []string                 `json:"vars"`
	States []map[string]interface{} `json:"states"`
}

// loadITFTrace reads an ITF JSON file and returns its states.
func loadITFTrace(path string) ([]map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading trace %s: %w", path, err)
	}
	var t itfTraceJSON
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("error parsing trace JSON: %w", err)
	}
	return t.States, nil
}

// itfStateToSpectre converts a single ITF state map to a state.State.
func itfStateToSpectre(raw map[string]interface{}) *state.State {
	s := &state.State{Variables: make(map[string]state.Value)}
	for k, v := range raw {
		if k == "#meta" {
			continue
		}
		var sv state.Value
		switch val := v.(type) {
		case bool:
			b := val
			sv = &state.PrimitiveValue{TypeName: "bool", BoolValue: &b}
		case float64:
			n := int64(val)
			sv = &state.PrimitiveValue{TypeName: "int", IntValue: &n}
		case string:
			sv = &state.PrimitiveValue{TypeName: "str", StringValue: &val}
		case map[string]interface{}:
			// ITF bigint: {"#bigint": "123"}
			if bi, ok := val["#bigint"]; ok {
				if s, ok := bi.(string); ok {
					var n int64
					fmt.Sscanf(s, "%d", &n)
					sv = &state.PrimitiveValue{TypeName: "int", IntValue: &n}
				}
			}
		default:
			fmt.Fprintf(os.Stderr, "Warning: ITF field %q has unsupported type %T — skipped; conformance check may be incomplete\n", k, v)
		}
		if sv != nil {
			s.Variables[k] = sv
		}
	}
	return s
}

// itfActionName extracts the action name from an ITF state's #meta.
func itfActionName(raw map[string]interface{}) string {
	if meta, ok := raw["#meta"].(map[string]interface{}); ok {
		if act, ok := meta["action"].(string); ok {
			return act
		}
	}
	return ""
}

// checkActionInSpec returns true if the action exists in the spec.
func checkActionInSpec(file *ast.File, actionName string) bool {
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok && a.Name == actionName {
			return true
		}
	}
	return false
}

// checkGuardsSatisfied checks all require stmts of an action against a state.
// Returns (ok bool, firstFailedGuard string).
func checkGuardsSatisfied(file *ast.File, actionName string, s *state.State) (bool, string) {
	for _, d := range file.Decls {
		a, ok := d.(*ast.ActionDecl)
		if !ok || a.Name != actionName || a.Body == nil {
			continue
		}
		env := eval.NewEnvironment()
		eval.RegisterEnumTypes(env, file)
		for k, v := range s.Variables {
			env.SetVariable(k, v)
		}
		evaluator := eval.NewEvaluator(env)
		for _, stmt := range a.Body.Statements {
			req, ok := stmt.(*ast.RequireStmt)
			if !ok {
				continue
			}
			result, err := evaluator.Eval(req.Condition)
			if err != nil {
				continue
			}
			prim, ok := result.(*state.PrimitiveValue)
			if ok && prim.TypeName == "bool" && prim.BoolValue != nil && !*prim.BoolValue {
				return false, fmt.Sprintf("require %v", req.Condition)
			}
		}
		return true, ""
	}
	// Action not found → report as missing.
	return false, "(action not defined in spec)"
}

func runDiffConformance(args []string) error {
	traceFile, specV1, specV2 := "", "", ""
	i := 0
	for i < len(args) {
		switch args[i] {
		case "--trace":
			if i+1 < len(args) {
				traceFile = args[i+1]
				i++
			}
		case "--spec-v1":
			if i+1 < len(args) {
				specV1 = args[i+1]
				i++
			}
		case "--spec-v2":
			if i+1 < len(args) {
				specV2 = args[i+1]
				i++
			}
		}
		i++
	}
	if traceFile == "" || specV1 == "" || specV2 == "" {
		return fmt.Errorf("usage: spectre diff-conformance --spec-v1 <s1.spec> --spec-v2 <s2.spec> --trace <trace.itf.json>")
	}

	file1, err := parseSpecFile(specV1)
	if err != nil {
		return fmt.Errorf("spec-v1: %w", err)
	}
	file2, err := parseSpecFile(specV2)
	if err != nil {
		return fmt.Errorf("spec-v2: %w", err)
	}

	states, err := loadITFTrace(traceFile)
	if err != nil {
		return err
	}

	fmt.Printf("Diff-conformance: %s  vs  %s\n", specV1, specV2)
	fmt.Printf("Trace: %s  (%d states)\n\n", traceFile, len(states))

	divergences := 0
	for idx, rawState := range states {
		action := itfActionName(rawState)
		if action == "" {
			continue // initial state, no action
		}
		s := itfStateToSpectre(rawState)

		ok1, fail1 := checkGuardsSatisfied(file1, action, s)
		ok2, fail2 := checkGuardsSatisfied(file2, action, s)

		v1str := "OK"
		if !ok1 {
			v1str = fmt.Sprintf("REJECTED (%s)", fail1)
		}
		v2str := "OK"
		if !ok2 {
			v2str = fmt.Sprintf("REJECTED (%s)", fail2)
		}

		if ok1 != ok2 {
			divergences++
			fmt.Printf("  [DIVERGE] Step %d: action '%s'\n", idx, action)
			fmt.Printf("            v1: %s\n", v1str)
			fmt.Printf("            v2: %s\n", v2str)
		} else {
			fmt.Printf("  Step %d: action '%s' — v1: %s, v2: %s\n", idx, action, v1str, v2str)
		}
	}

	fmt.Printf("\n%d step(s) replayed. %d divergence(s) found between v1 and v2.\n",
		len(states)-1, divergences)
	if divergences > 0 {
		return fmt.Errorf("%d conformance divergence(s) detected", divergences)
	}
	return nil
}

// ── Feature 11: spectre simulate ─────────────────────────────────────────────

func runSimulate(args []string) error {
	nTraces := 10
	maxSteps := 100
	seed := int64(42)
	outputDir := ""
	useBFS := false
	filteredArgs := []string{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--traces":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &nTraces)
				i++
			}
		case "--max-steps", "--max-depth":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxSteps)
				i++
			}
		case "--seed":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &seed)
				i++
			}
		case "--output-dir":
			if i+1 < len(args) {
				outputDir = args[i+1]
				i++
			}
		case "--use-bfs":
			useBFS = true
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}
	if len(filteredArgs) == 0 {
		return fmt.Errorf("usage: spectre simulate <spec.spec> [--traces N] [--max-steps M] [--seed S] [--output-dir dir] [--use-bfs]")
	}
	filename := filteredArgs[0]

	file, err := parseSpecFile(filename)
	if err != nil {
		return err
	}
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		return fmt.Errorf("error building state machine: %w", err)
	}

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("error creating output directory: %w", err)
		}
	}

	if useBFS {
		return runSimulateBFS(filename, sm, nTraces, maxSteps, seed, outputDir)
	}
	return runSimulateDirect(filename, file, sm, nTraces, maxSteps, seed, outputDir)
}

// runSimulateDirect samples traces directly from the transition semantics without
// constructing a reachable-state graph first. At each step it evaluates enabled
// actions in the current state and picks one at random, making it usable even when
// the full state space is too large for exhaustive BFS.
func runSimulateDirect(filename string, _ interface{}, sm *exec.StateMachine,
	nTraces, maxSteps int, seed int64, outputDir string) error {

	initStates, err := sm.GetInitialStates()
	if err != nil || len(initStates) == 0 {
		return fmt.Errorf("could not obtain initial states: %w", err)
	}

	fmt.Printf("Simulating %s (%d traces, max %d steps/trace — direct mode)\n",
		filename, nTraces, maxSteps)

	violations := 0
	for t := 0; t < nTraces; t++ {
		rng := rand.New(rand.NewSource(seed + int64(t)))
		cur := initStates[rng.Intn(len(initStates))]

		desc := fmt.Sprintf("Direct simulation trace %d", t+1)
		trace := itf.NewTrace(filename, desc)
		itf.AppendStep(trace, cur, 0, "", nil)

		for step := 1; step <= maxSteps; step++ {
			avail, aErr := sm.GetAvailableActionsWithArgs(cur)
			if aErr != nil || len(avail) == 0 {
				break
			}
			chosen := avail[rng.Intn(len(avail))]
			next, eErr := sm.ExecuteAction(chosen.ActionName, cur, chosen.Args)
			if eErr != nil {
				break
			}
			itf.AppendStep(trace, next, step, chosen.ActionName, chosen.Args)
			// Check invariants on new state.
			errs, vErr := sm.ValidateState(next)
			if vErr == nil && len(errs) > 0 {
				violations++
				break
			}
			cur = next
		}

		data, jsonErr := json.MarshalIndent(trace, "", "  ")
		if jsonErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not marshal trace %d: %v\n", t+1, jsonErr)
			continue
		}
		if outputDir != "" {
			outPath := filepath.Join(outputDir, fmt.Sprintf("sim_trace_%03d.itf.json", t+1))
			if writeErr := os.WriteFile(outPath, data, 0644); writeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not write %s: %v\n", outPath, writeErr)
			} else {
				fmt.Printf("  Wrote %s\n", outPath)
			}
		} else {
			fmt.Printf("--- Trace %d ---\n%s\n", t+1, string(data))
		}
	}

	fmt.Printf("\nGenerated %d trace(s), %d invariant violation(s) found.\n", nTraces, violations)
	if violations > 0 {
		fmt.Println("Run `spectre verify` for full counterexample analysis.")
	}
	return nil
}

// runSimulateBFS is the legacy BFS-then-random-walk mode, kept for users who want
// to verify exhaustively before sampling. Only suitable when state space fits in memory.
func runSimulateBFS(filename string, sm *exec.StateMachine,
	nTraces, maxDepth int, seed int64, outputDir string) error {

	maxStates := 5000
	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(maxDepth)
	explorer.SetMaxStates(maxStates)
	explorer.SetSilent(true)

	fmt.Printf("Simulating %s (BFS mode: max %d states, %d traces)…\n", filename, maxStates, nTraces)
	result, err := explorer.ExploreBFS()
	if err != nil {
		return fmt.Errorf("exploration error: %w", err)
	}
	fmt.Printf("Explored %d states.\n", result.StatesExplored)

	hasher := explore.NewStateHasher()
	for t := 0; t < nTraces; t++ {
		rng := rand.New(rand.NewSource(seed + int64(t)))
		trace := itf.RandomWalk(filename, result.TransitionGraph, result.InitialStates, maxDepth, rng, hasher)
		if trace == nil {
			continue
		}
		data, jsonErr := json.MarshalIndent(trace, "", "  ")
		if jsonErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not marshal trace %d: %v\n", t+1, jsonErr)
			continue
		}
		if outputDir != "" {
			outPath := filepath.Join(outputDir, fmt.Sprintf("sim_trace_%03d.itf.json", t+1))
			if writeErr := os.WriteFile(outPath, data, 0644); writeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not write %s: %v\n", outPath, writeErr)
			} else {
				fmt.Printf("  Wrote %s\n", outPath)
			}
		} else {
			fmt.Printf("--- Trace %d ---\n%s\n", t+1, string(data))
		}
	}

	violations := len(result.Violations)
	fmt.Printf("\nGenerated %d trace(s), %d state(s) explored, %d violation(s) found during BFS.\n",
		nTraces, result.StatesExplored, violations)
	if violations > 0 {
		fmt.Println("Run `spectre verify` for full counterexample analysis.")
	}
	return nil
}

// ── Feature: spectre impact ───────────────────────────────────────────────────

// runImpact analyses a git diff (or a provided diff file) to determine which
// spec actions and invariants are affected by the Rust source changes, whether
// the refinement map changed, and whether incremental re-verification is
// required.
//
// Usage:
//
//	spectre impact <spec-file.spec> [--diff <file.diff>] [--git-diff]
func runImpact(args []string) error {
	specFile := ""
	diffFile := ""
	useGitDiff := false
	filteredArgs := []string{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--diff":
			if i+1 < len(args) {
				diffFile = args[i+1]
				i++
			}
		case "--git-diff":
			useGitDiff = true
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}

	if len(filteredArgs) > 0 {
		specFile = filteredArgs[0]
	}
	if specFile == "" {
		return fmt.Errorf("usage: spectre impact <spec-file.spec> [--diff <file.diff>] [--git-diff]")
	}

	// Load the spec.
	parsedSpec, err := parseSpecFile(specFile)
	if err != nil {
		return fmt.Errorf("cannot parse spec: %w", err)
	}

	// Obtain the diff text.
	var diffText string
	if useGitDiff {
		diffText = runGitDiff()
	} else if diffFile != "" {
		data, readErr := os.ReadFile(diffFile)
		if readErr != nil {
			return fmt.Errorf("cannot read diff file: %w", readErr)
		}
		diffText = string(data)
	} else {
		// Try git diff by default.
		diffText = runGitDiff()
		if diffText == "" {
			fmt.Println("No diff provided and no uncommitted git changes detected.")
			fmt.Println("Usage: spectre impact <spec.spec> [--diff <file.diff>] [--git-diff]")
			return nil
		}
	}

	// Extract changed Rust method names from the diff.
	changedMethods := extractChangedMethods(diffText)

	// Load the driver sidecar for the spec.
	meta, _ := codegen.LoadDriverMeta(codegen.MetaPath(specFile))

	// Determine which spec actions are affected.
	affectedActions := findAffectedActions(changedMethods, parsedSpec, meta)

	// Determine which invariants reference affected variables.
	affectedVars := findWrittenVars(affectedActions, parsedSpec)
	affectedInvariants := findAffectedInvariants(affectedVars, parsedSpec)

	// Check if refinement map changed (method signatures changed for mapped actions).
	refinementMapChanged := refinementChanged(changedMethods, meta)

	fmt.Printf("Impact Analysis: %s\n\n", specFile)

	if len(changedMethods) == 0 {
		fmt.Println("No Rust method changes detected in diff.")
	} else {
		fmt.Printf("Changed Rust methods: %s\n", strings.Join(sortedKeys(changedMethods), ", "))
	}

	fmt.Printf("\nAffected spec actions:\n")
	if len(affectedActions) == 0 {
		fmt.Println("  (none)")
	}
	for _, a := range affectedActions {
		fmt.Printf("  [!] %s\n", a)
	}

	fmt.Printf("\nAffected invariants (reference written variables):\n")
	if len(affectedInvariants) == 0 {
		fmt.Println("  (none)")
	}
	for _, inv := range affectedInvariants {
		fmt.Printf("  [!] %s\n", inv)
	}

	fmt.Printf("\nRefinement map changed: %v\n", refinementMapChanged)

	fmt.Printf("\nRecommendation:\n")
	if len(affectedActions) > 0 {
		fmt.Printf("  Run: spectre verify %s --incremental", specFile)
		for _, a := range affectedActions {
			fmt.Printf(" --changed-action %s", a)
			break // show only first; user can add more
		}
		fmt.Println()
	}
	if len(affectedInvariants) > 0 {
		fmt.Println("  Re-generate tests: spectre verify --emit-rust-tests tests.rs", specFile)
	}
	if refinementMapChanged {
		fmt.Println("  Re-run: spectre mine to update the driver sidecar")
	}
	if len(affectedActions) == 0 && len(affectedInvariants) == 0 && !refinementMapChanged {
		fmt.Println("  No re-verification required (changes appear spec-neutral).")
	}

	return nil
}

// runGitDiff runs `git diff HEAD` and returns the output, or "" on failure.
func runGitDiff() string {
	// Use os/exec to avoid import cycle — commands_extra is already in main.
	cmd := osexec.Command("git", "diff", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// extractChangedMethods parses a unified diff and returns a set of Rust
// method names that have changed (added/removed/modified lines in fn bodies).
func extractChangedMethods(diff string) map[string]bool {
	changed := make(map[string]bool)
	// Heuristic: look for diff hunks that touch lines matching `fn <name>(`.
	for _, line := range strings.Split(diff, "\n") {
		line = strings.TrimLeft(line, " \t")
		if (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-")) &&
			strings.Contains(line, "fn ") {
			// Skip unified-diff file headers (--- a/file, +++ b/file).
			if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
				continue
			}
			// Extract function name.
			rest := line[1:]
			idx := strings.Index(rest, "fn ")
			if idx < 0 {
				continue
			}
			nameStart := idx + 3
			nameEnd := strings.IndexAny(rest[nameStart:], "(<")
			if nameEnd < 0 {
				continue
			}
			name := strings.TrimSpace(rest[nameStart : nameStart+nameEnd])
			if name != "" {
				changed[name] = true
			}
		}
	}
	return changed
}

// findAffectedActions returns spec action names that correspond to changed Rust methods.
func findAffectedActions(changedMethods map[string]bool, file *ast.File, meta *codegen.DriverMeta) []string {
	var affected []string
	seen := make(map[string]bool)

	// Via driver sidecar: direct method→action mapping.
	if meta != nil {
		for _, a := range meta.Actions {
			if changedMethods[a.RustMethod] {
				if !seen[a.SpecAction] {
					seen[a.SpecAction] = true
					affected = append(affected, a.SpecAction)
				}
			}
		}
	}

	// Fallback: match by name (action name == method name).
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok {
			if changedMethods[a.Name] && !seen[a.Name] {
				seen[a.Name] = true
				affected = append(affected, a.Name)
			}
		}
	}
	return affected
}

// findWrittenVars returns the set of variables written by the affected actions.
func findWrittenVars(actions []string, file *ast.File) map[string]bool {
	actionSet := make(map[string]bool)
	for _, a := range actions {
		actionSet[a] = true
	}
	written := make(map[string]bool)
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok && actionSet[a.Name] && a.Body != nil {
			for _, s := range a.Body.Statements {
				if assign, ok := s.(*ast.AssignStmt); ok {
					if id, ok := assign.Left.(*ast.Ident); ok {
						written[id.Name] = true
					}
				}
			}
		}
	}
	return written
}

// findAffectedInvariants returns invariant names that reference any of the given variables.
func findAffectedInvariants(vars map[string]bool, file *ast.File) []string {
	var affected []string
	for _, d := range file.Decls {
		if inv, ok := d.(*ast.InvariantDecl); ok {
			if exprReferencesAny(inv.Condition, vars) {
				affected = append(affected, inv.Name)
			}
		}
	}
	return affected
}

func exprReferencesAny(e ast.Expr, vars map[string]bool) bool {
	switch expr := e.(type) {
	case *ast.Ident:
		return vars[expr.Name]
	case *ast.BinaryExpr:
		return exprReferencesAny(expr.Left, vars) || exprReferencesAny(expr.Right, vars)
	case *ast.UnaryExpr:
		return exprReferencesAny(expr.Expr, vars)
	case *ast.ParenExpr:
		return exprReferencesAny(expr.X, vars)
	}
	return false
}

// refinementChanged returns true if any field or method in meta matches a changed Rust method.
func refinementChanged(changedMethods map[string]bool, meta *codegen.DriverMeta) bool {
	if meta == nil {
		return false
	}
	for _, a := range meta.Actions {
		if changedMethods[a.RustMethod] {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ── SMT helpers for runSync ───────────────────────────────────────────────────

// smtCheckGuards runs an SMT equivalence check on the i-th guard of the given action.
// Returns SMTUnknown when AST expressions are unavailable.
func smtCheckGuards(
	oldExprs, newExprs map[string][]ast.Expr,
	action string, idx int,
) (diagnose.SMTResult, string) {
	if oldExprs == nil || newExprs == nil {
		return diagnose.SMTUnknown, ""
	}
	oldList, newList := oldExprs[action], newExprs[action]
	if idx >= len(oldList) || idx >= len(newList) {
		return diagnose.SMTUnknown, ""
	}
	return diagnose.CheckExprEquivalence(oldList[idx], newList[idx])
}

// smtCheckAssigns runs an SMT equivalence check on the i-th assignment RHS of
// the given action.  Returns SMTUnknown when AST expressions are unavailable.
func smtCheckAssigns(
	oldExprs, newExprs map[string][]ast.Expr,
	action string, idx int,
) (diagnose.SMTResult, string) {
	if oldExprs == nil || newExprs == nil {
		return diagnose.SMTUnknown, ""
	}
	oldList, newList := oldExprs[action], newExprs[action]
	if idx >= len(oldList) || idx >= len(newList) {
		return diagnose.SMTUnknown, ""
	}
	return diagnose.CheckExprEquivalence(oldList[idx], newList[idx])
}
