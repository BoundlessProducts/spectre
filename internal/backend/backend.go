// Package backend defines the pluggable verification backend interface.
// The default backend is BFS (the existing explorer); optional backends
// delegate to external tools (TLC, Apalache, Z3).
package backend

import (
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// Kind names a verification backend.
type Kind string

const (
	KindBFS      Kind = "bfs"
	KindTLC      Kind = "tlc"
	KindApalache Kind = "apalache"
	KindZ3       Kind = "z3"
)

// Options holds parameters shared across all backends.
type Options struct {
	MaxStates int
	MaxDepth  int

	// Z3 / BMC only.
	BMCDepth int

	// TLC: path to tla2tools.jar (or "tlc" if on PATH).
	TLCJar string
	// Apalache: path to apalache-mc binary (or "apalache-mc" if on PATH).
	ApalacheBin string
	// Z3: path to z3 binary (or "z3" if on PATH).
	Z3Bin string

	// Params are constant overrides (name → value string), e.g. {"MAX":"10"}.
	Params map[string]string

	Silent bool
}

// Result is the unified result returned by every backend.
type Result struct {
	// Safe is true when no violations were found within the explored bound.
	Safe bool
	// Exhaustive is true when the backend explored the full reachable state space.
	Exhaustive bool
	// Violations are the violations found (may be empty even when !Safe for
	// backends that only report sat/unsat without a witness).
	Violations []*explore.Violation
	// StatesExplored is the number of distinct states visited (0 if unknown).
	StatesExplored int
	// MaxDepth is the deepest path explored (0 if unknown).
	MaxDepth int
	// BackendName is the name of the backend that produced this result.
	BackendName string
	// Note carries extra human-readable information (e.g. "bounded to depth 20").
	Note string
}

// Backend is the interface every verification backend must implement.
type Backend interface {
	// Name returns the backend identifier.
	Name() Kind
	// Verify runs verification on the parsed spec file.
	Verify(file *ast.File, opts Options) (*Result, error)
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		MaxStates:   10000,
		MaxDepth:    200,
		BMCDepth:    20,
		TLCJar:      "tlc",
		ApalacheBin: "apalache-mc",
		Z3Bin:       "z3",
		Params:      make(map[string]string),
	}
}
