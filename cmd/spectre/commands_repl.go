package main

// commands_repl.go — Item 14: interactive REPL and Item 15: spec test runner.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BoundlessProducts/spectre/internal/eval"
	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/internal/semantic"
	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

func init() {
	Commands["repl"] = &Command{
		Name:        "repl",
		Description: "Interactive REPL: inspect states, step actions, and evaluate expressions",
		Run:         runREPL,
	}
	Commands["test"] = &Command{
		Name:        "test",
		Description: "Run verification on one or more spec files and report a summary table",
		Run:         runTest,
	}
}

// ── Item 14: spectre repl ──────────────────────────────────────────────────────

// replSession holds the REPL's mutable state.
type replSession struct {
	sm          *exec.StateMachine
	file        *ast.File
	allFiles    []*ast.File
	current     *state.State
	initStates  []*state.State
	history     []replStep
}

type replStep struct {
	action string
	before *state.State
	after  *state.State
}

func runREPL(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: spectre repl <spec.spec>")
	}
	filename := args[0]

	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filename, err)
	}

	sess, err := buildReplSession(filename, string(content))
	if err != nil {
		return err
	}

	fmt.Printf("Spectre REPL — %s\n", filename)
	fmt.Println("Type 'help' for available commands.")
	if len(sess.initStates) > 1 {
		fmt.Printf("(%d initial states; showing first)\n", len(sess.initStates))
	}
	fmt.Println()
	sess.printState()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if done := sess.handleLine(line); done {
			break
		}
	}
	fmt.Println("Goodbye.")
	return nil
}

// buildReplSession constructs a replSession from a filename and its content.
// Extracted for testability.
func buildReplSession(filename, content string) (*replSession, error) {
	l := lexer.New(content)
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors in %s:\n  %s", filename, strings.Join(p.Errors(), "\n  "))
	}

	// Load additional files for import'd modules.
	var allFiles []*ast.File
	allFiles = append(allFiles, file)
	for _, d := range file.Decls {
		if _, ok := d.(*ast.ImportDecl); ok {
			dir := filepath.Dir(filename)
			loader := semantic.NewModuleLoader(dir)
			_, loadErrs := loader.LoadModule(filename)
			if len(loadErrs) == 0 {
				for _, mi := range loader.GetAllModules() {
					if mi.File != file {
						allFiles = append(allFiles, mi.File)
					}
				}
			}
			break
		}
	}

	additionalFiles := allFiles[1:]
	sm, err := exec.NewStateMachine(file, additionalFiles...)
	if err != nil {
		return nil, fmt.Errorf("error building state machine: %w", err)
	}

	initStates, err := sm.GetInitialStates()
	if err != nil || len(initStates) == 0 {
		return nil, fmt.Errorf("cannot obtain initial states: %w", err)
	}

	return &replSession{
		sm:         sm,
		file:       file,
		allFiles:   allFiles,
		current:    initStates[0],
		initStates: initStates,
	}, nil
}

func (s *replSession) handleLine(line string) (quit bool) {
	parts := strings.Fields(line)
	cmd := parts[0]

	switch cmd {
	case "quit", "exit", "q":
		return true

	case "help", "h", "?":
		fmt.Println("Commands:")
		fmt.Println("  state               — show current state")
		fmt.Println("  actions             — list enabled actions")
		fmt.Println("  <action>            — execute action (e.g. incr)")
		fmt.Println("  eval <expr>         — evaluate expression in current state")
		fmt.Println("  reset               — return to initial state")
		fmt.Println("  back                — undo last step")
		fmt.Println("  history             — show step history")
		fmt.Println("  init [N]            — switch to initial state N (default 0)")
		fmt.Println("  check               — check invariants on current state")
		fmt.Println("  quit                — exit REPL")

	case "state", "s":
		s.printState()

	case "actions", "a":
		s.printActions()

	case "eval", "e":
		if len(parts) < 2 {
			fmt.Println("Usage: eval <expression>")
			return false
		}
		exprStr := strings.Join(parts[1:], " ")
		s.evalExpr(exprStr)

	case "reset", "r":
		s.current = s.initStates[0]
		s.history = nil
		fmt.Println("Reset to initial state.")
		s.printState()

	case "back", "b":
		if len(s.history) == 0 {
			fmt.Println("Nothing to undo.")
			return false
		}
		last := s.history[len(s.history)-1]
		s.current = last.before
		s.history = s.history[:len(s.history)-1]
		fmt.Printf("Undid '%s'.\n", last.action)
		s.printState()

	case "history":
		if len(s.history) == 0 {
			fmt.Println("No steps taken yet.")
			return false
		}
		for i, step := range s.history {
			fmt.Printf("  [%d] %s\n", i+1, step.action)
		}

	case "init":
		idx := 0
		if len(parts) >= 2 {
			fmt.Sscanf(parts[1], "%d", &idx)
		}
		if idx < 0 || idx >= len(s.initStates) {
			fmt.Printf("Invalid initial state index %d (0..%d)\n", idx, len(s.initStates)-1)
			return false
		}
		s.current = s.initStates[idx]
		s.history = nil
		fmt.Printf("Switched to initial state %d.\n", idx)
		s.printState()

	case "check", "c":
		errs, err := s.sm.ValidateState(s.current)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return false
		}
		if len(errs) == 0 {
			fmt.Println("All invariants hold.")
		} else {
			for _, e := range errs {
				fmt.Printf("  VIOLATION: %s\n", e.Message)
			}
		}

	default:
		// Try as action name.
		s.tryAction(line)
	}
	return false
}

func (s *replSession) tryAction(line string) {
	// Accept "actionName" or "actionName()" — no argument parsing for parameterized actions.
	name := strings.TrimSuffix(strings.TrimSpace(line), "()")
	name = strings.TrimRight(name, " \t")

	// Check if action exists.
	if !s.sm.HasAction(name) {
		fmt.Printf("Unknown command or action: %q\n", line)
		fmt.Println("Type 'help' for available commands or 'actions' to list enabled actions.")
		return
	}

	next, err := s.sm.ExecuteAction(name, s.current, nil)
	if err != nil {
		fmt.Printf("Action '%s' failed: %v\n", name, err)
		return
	}

	s.history = append(s.history, replStep{
		action: name,
		before: s.current,
		after:  next,
	})
	s.current = next
	fmt.Printf("Executed '%s'.\n", name)
	s.printState()

	// Automatically check invariants.
	errs, _ := s.sm.ValidateState(s.current)
	for _, e := range errs {
		fmt.Printf("  *** INVARIANT VIOLATION: %s ***\n", e.Message)
	}
}

func (s *replSession) evalExpr(exprStr string) {
	l := lexer.New(exprStr)
	p := parser.New(l)
	expr := p.ParseExpression()
	if len(p.Errors()) > 0 {
		fmt.Printf("Parse error: %s\n", strings.Join(p.Errors(), "; "))
		return
	}
	if expr == nil {
		fmt.Println("Empty expression.")
		return
	}

	// Build evaluation environment from current state.
	env := eval.NewEnvironment()
	for _, f := range s.allFiles {
		eval.RegisterEnumTypes(env, f)
		eval.RegisterConstants(env, f, s.allFiles...)
	}
	eval.RegisterFunctions(env, s.file, s.allFiles[1:]...)
	for k, v := range s.current.Variables {
		env.SetVariable(k, v)
	}

	evaluator := eval.NewEvaluator(env)
	result, err := evaluator.Eval(expr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("= %s\n", result.String())
}

func (s *replSession) printState() {
	if s.current == nil {
		fmt.Println("(no state)")
		return
	}
	vars := make([]string, 0, len(s.current.Variables))
	for k := range s.current.Variables {
		vars = append(vars, k)
	}
	sort.Strings(vars)
	fmt.Println("State:")
	for _, k := range vars {
		fmt.Printf("  %s = %s\n", k, s.current.Variables[k].String())
	}
}

func (s *replSession) printActions() {
	avail, err := s.sm.GetAvailableActionsWithArgs(s.current)
	if err != nil {
		fmt.Printf("Error listing actions: %v\n", err)
		return
	}
	if len(avail) == 0 {
		fmt.Println("No actions enabled in current state (deadlock).")
		return
	}
	fmt.Println("Enabled actions:")
	for _, a := range avail {
		if len(a.Args) == 0 {
			fmt.Printf("  %s\n", a.ActionName)
		} else {
			argStrs := make([]string, len(a.Args))
			for i, arg := range a.Args {
				argStrs[i] = arg.String()
			}
			fmt.Printf("  %s(%s)\n", a.ActionName, strings.Join(argStrs, ", "))
		}
	}
}

// ── Item 15: spectre test ─────────────────────────────────────────────────────

// testResult records the outcome of running verification on one spec file.
type testResult struct {
	file       string
	passed     bool
	violations int
	states     int
	elapsed    time.Duration
	err        error
	kind       explore.VerificationKind
}

func runTest(args []string) error {
	maxStates := 10000
	maxDepth := 100
	failFast := false
	verbose := false
	filteredArgs := []string{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--max-states":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxStates)
				i++
			}
		case "--max-depth":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxDepth)
				i++
			}
		case "--fail-fast":
			failFast = true
		case "--verbose", "-v":
			verbose = true
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}

	if len(filteredArgs) == 0 {
		return fmt.Errorf("usage: spectre test <spec.spec> [<spec2.spec>...] [--max-states N] [--max-depth N] [--fail-fast] [--verbose]")
	}

	// Expand directories into spec files.
	var specFiles []string
	for _, arg := range filteredArgs {
		info, statErr := os.Stat(arg)
		if statErr != nil {
			specFiles = append(specFiles, arg)
			continue
		}
		if info.IsDir() {
			found, findErr := findSpecFiles(arg)
			if findErr == nil {
				specFiles = append(specFiles, found...)
			}
		} else {
			specFiles = append(specFiles, arg)
		}
	}
	if len(specFiles) == 0 {
		return fmt.Errorf("no .spec files found in %v", filteredArgs)
	}

	fmt.Printf("Running %d spec test(s)...\n\n", len(specFiles))

	results := make([]testResult, 0, len(specFiles))
	for _, sf := range specFiles {
		r := runOneSpecTest(sf, maxStates, maxDepth, verbose)
		results = append(results, r)
		status := "PASS"
		if !r.passed {
			if r.err != nil {
				status = "ERROR"
			} else {
				status = "FAIL"
			}
		}
		kindTag := ""
		if r.err == nil {
			kindTag = fmt.Sprintf(" [%s]", r.kind.String())
		}
		if verbose || !r.passed {
			fmt.Printf("  %-8s  %s%s\n", status, sf, kindTag)
			if r.err != nil {
				fmt.Printf("           error: %v\n", r.err)
			} else if r.violations > 0 {
				fmt.Printf("           violations: %d  states: %d  elapsed: %s\n",
					r.violations, r.states, r.elapsed.Round(time.Millisecond))
			} else {
				fmt.Printf("           states: %d  elapsed: %s\n",
					r.states, r.elapsed.Round(time.Millisecond))
			}
		} else {
			marker := "."
			if status != "PASS" {
				marker = "F"
			}
			fmt.Print(marker)
		}
		if failFast && !r.passed {
			fmt.Println()
			break
		}
	}

	// Summary table.
	nPass, nFail, nErr := 0, 0, 0
	for _, r := range results {
		if r.err != nil {
			nErr++
		} else if r.passed {
			nPass++
		} else {
			nFail++
		}
	}

	fmt.Printf("\n\n--- Test Summary ---\n")
	fmt.Printf("  Passed:  %d\n", nPass)
	fmt.Printf("  Failed:  %d\n", nFail)
	fmt.Printf("  Errors:  %d\n", nErr)
	fmt.Printf("  Total:   %d\n", len(results))

	if nFail > 0 || nErr > 0 {
		fmt.Println("\nFailed specs:")
		for _, r := range results {
			if !r.passed {
				if r.err != nil {
					fmt.Printf("  %s  (error: %v)\n", r.file, r.err)
				} else {
					fmt.Printf("  %s  (%d violation(s), %d states)\n", r.file, r.violations, r.states)
				}
			}
		}
		return fmt.Errorf("%d test(s) failed", nFail+nErr)
	}
	fmt.Println("\nAll tests passed.")
	return nil
}

func runOneSpecTest(filename string, maxStates, maxDepth int, verbose bool) testResult {
	start := time.Now()

	content, err := os.ReadFile(filename)
	if err != nil {
		return testResult{file: filename, err: err, elapsed: time.Since(start)}
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		return testResult{
			file:    filename,
			err:     fmt.Errorf("parse errors: %s", strings.Join(p.Errors(), "; ")),
			elapsed: time.Since(start),
		}
	}

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		return testResult{file: filename, err: err, elapsed: time.Since(start)}
	}

	exp := explore.NewExplorer(sm)
	exp.SetMaxStates(maxStates)
	exp.SetMaxDepth(maxDepth)
	exp.SetSilent(true)

	result, err := exp.ExploreBFS()
	if err != nil {
		return testResult{file: filename, err: err, elapsed: time.Since(start)}
	}

	violations := len(result.Violations)
	return testResult{
		file:       filename,
		passed:     violations == 0,
		violations: violations,
		states:     result.StatesExplored,
		elapsed:    time.Since(start),
		kind:       result.Kind,
	}
}
