package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

func init() {
	Commands["benchmark"] = &Command{
		Name:        "benchmark",
		Description: "Run BFS exploration benchmarks on spec files and report timing",
		Run:         runBenchmark,
	}
}

type benchResult struct {
	file    string
	states  int
	elapsed time.Duration
	depth   int
	err     error
}

// runBenchmark runs BFS on each provided spec file (or all *.spec files in the
// benchmarks/ directory) and prints a timing table.
func runBenchmark(args []string) error {
	maxStates := 50000
	dir := "benchmarks"
	filteredArgs := []string{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--max-states":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxStates)
				i++
			}
		case "--dir":
			if i+1 < len(args) {
				dir = args[i+1]
				i++
			}
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}

	var files []string
	if len(filteredArgs) > 0 {
		files = filteredArgs
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("cannot read benchmark directory %q: %w", dir, err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".spec") {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
	}
	if len(files) == 0 {
		return fmt.Errorf("no spec files found (use --dir or pass files explicitly)")
	}

	fmt.Printf("%-30s  %8s  %12s  %8s\n", "Spec", "States", "Elapsed", "Depth")
	fmt.Println(strings.Repeat("-", 66))

	var results []benchResult
	for _, f := range files {
		r := runOneBenchmark(f, maxStates)
		results = append(results, r)
		if r.err != nil {
			fmt.Printf("%-30s  %8s  %12s  %8s  ERROR: %v\n",
				filepath.Base(f), "-", "-", "-", r.err)
		} else {
			fmt.Printf("%-30s  %8d  %12s  %8d\n",
				filepath.Base(f), r.states, r.elapsed.Round(time.Millisecond), r.depth)
		}
	}

	// Summary
	total := len(results)
	errs := 0
	for _, r := range results {
		if r.err != nil {
			errs++
		}
	}
	fmt.Println(strings.Repeat("-", 66))
	fmt.Printf("%d spec(s) benchmarked", total)
	if errs > 0 {
		fmt.Printf(", %d error(s)", errs)
	}
	fmt.Println()
	return nil
}

func runOneBenchmark(filename string, maxStates int) benchResult {
	r := benchResult{file: filename}

	content, err := os.ReadFile(filename)
	if err != nil {
		r.err = err
		return r
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		r.err = fmt.Errorf("parse errors: %s", strings.Join(p.Errors(), "; "))
		return r
	}

	sm, err := exec.NewStateMachine(file)
	if err != nil {
		r.err = fmt.Errorf("build error: %w", err)
		return r
	}

	e := explore.NewExplorer(sm)
	e.SetMaxStates(maxStates)
	e.SetSilent(true)

	start := time.Now()
	result, err := e.ExploreBFS()
	r.elapsed = time.Since(start)
	if err != nil {
		r.err = err
		return r
	}

	r.states = result.StatesExplored
	r.depth = result.MaxDepth
	return r
}
