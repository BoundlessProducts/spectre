package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/monitor"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

func init() {
	Commands["validate-monitor"] = &Command{
		Name:        "validate-monitor",
		Description: "Validate that BFS violation traces are consistent with generated-monitor semantics",
		Run:         runValidateMonitor,
	}
}

// runValidateMonitor runs BFS on a spec and validates that every found
// violation trace is consistent with what a generated monitor would observe.
func runValidateMonitor(args []string) error {
	maxStates := 10000
	maxDepth := 0
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
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}

	if len(filteredArgs) == 0 {
		return fmt.Errorf("no spec file specified")
	}

	filename := filteredArgs[0]
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filename, err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()
	if errs := p.Errors(); len(errs) > 0 {
		return fmt.Errorf("parse errors in %s: %s", filename, strings.Join(errs, "; "))
	}

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		return fmt.Errorf("error building state machine: %w", err)
	}

	e := explore.NewExplorer(sm)
	e.SetMaxStates(maxStates)
	if maxDepth > 0 {
		e.SetMaxDepth(maxDepth)
	}
	e.SetSilent(true)

	result, err := e.ExploreBFS()
	if err != nil {
		return fmt.Errorf("BFS error: %w", err)
	}

	if len(result.Violations) == 0 {
		fmt.Printf("No violations found (%d states explored) — nothing to validate.\n", result.StatesExplored)
		return nil
	}

	fmt.Printf("Validating %d violation trace(s) from %d states explored...\n",
		len(result.Violations), result.StatesExplored)

	results := monitor.ValidateAll(sm, result.Violations)
	fmt.Print(monitor.Summary(results))

	inconsistent := 0
	for _, r := range results {
		if !r.Consistent {
			inconsistent++
		}
	}
	if inconsistent > 0 {
		return fmt.Errorf("%d/%d traces inconsistent with monitor semantics",
			inconsistent, len(results))
	}
	return nil
}
