package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/backend"
	"github.com/BoundlessProducts/spectre/internal/backend/tla"
	"github.com/BoundlessProducts/spectre/internal/backend/z3"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

// verifyExternalOpts bundles the flags that are only relevant for non-BFS backends.
type verifyExternalOpts struct {
	backendName string
	bmcDepth    int
	tlcJar      string
	apalacheBin string
	z3Bin       string
	params      map[string]int64
	maxStates   int
	maxDepth    int
	silent      bool
}

// isExternalBackend returns true when the named backend is not the built-in BFS.
func isExternalBackend(name string) bool {
	switch backend.Kind(name) {
	case backend.KindTLC, backend.KindApalache, backend.KindZ3:
		return true
	}
	return false
}

// makeBackend constructs the Backend implementation for the given name.
func makeBackend(name string) (backend.Backend, error) {
	switch backend.Kind(name) {
	case backend.KindTLC:
		return tla.NewTLCBackend(), nil
	case backend.KindApalache:
		return tla.NewApalacheBackend(), nil
	case backend.KindZ3:
		return z3.NewZ3Backend(), nil
	default:
		return nil, fmt.Errorf("unknown backend %q — choose bfs, tlc, apalache, or z3", name)
	}
}

// runVerifyExternal handles verification for non-BFS backends.
// It parses the file, builds backend.Options, and calls Backend.Verify.
func runVerifyExternal(filenames []string, o verifyExternalOpts) error {
	b, err := makeBackend(o.backendName)
	if err != nil {
		return err
	}

	// Convert int64 params to string map for backend.Options.
	strParams := make(map[string]string, len(o.params))
	for k, v := range o.params {
		strParams[k] = strconv.FormatInt(v, 10)
	}

	opts := backend.Options{
		MaxStates:   o.maxStates,
		MaxDepth:    o.maxDepth,
		BMCDepth:    o.bmcDepth,
		TLCJar:      o.tlcJar,
		ApalacheBin: o.apalacheBin,
		Z3Bin:       o.z3Bin,
		Params:      strParams,
		Silent:      o.silent,
	}

	anyViolation := false

	for _, filename := range filenames {
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("read %s: %w", filename, err)
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		file := p.ParseFile()
		if len(p.Errors()) > 0 {
			return fmt.Errorf("parse errors in %s: %s", filename, strings.Join(p.Errors(), "; "))
		}

		fmt.Printf("Verifying %s with backend %s...\n", filename, b.Name())

		result, err := b.Verify(file, opts)
		if err != nil {
			return fmt.Errorf("backend %s: %w", b.Name(), err)
		}

		printBackendResult(filename, result)
		if !result.Safe {
			anyViolation = true
		}
	}

	if anyViolation {
		return fmt.Errorf("verification failed")
	}
	return nil
}

// printBackendResult prints the result from an external backend.
func printBackendResult(filename string, r *backend.Result) {
	label := "✓ safe"
	if !r.Safe {
		label = "✗ violation found"
	}
	exhaustiveStr := ""
	if r.Exhaustive {
		exhaustiveStr = " (exhaustive)"
	}

	fmt.Printf("%s: %s%s", filename, label, exhaustiveStr)
	if r.StatesExplored > 0 {
		fmt.Printf(" — %d states", r.StatesExplored)
	}
	fmt.Println()

	if r.Note != "" {
		fmt.Printf("  Note: %s\n", r.Note)
	}

	for i, v := range r.Violations {
		fmt.Printf("  Violation %d: invariant %q", i+1, v.Invariant)
		if len(v.Path) > 0 {
			last := v.Path[len(v.Path)-1]
			fmt.Printf(" triggered by action %q", last.Action)
		}
		fmt.Println()
	}
}
