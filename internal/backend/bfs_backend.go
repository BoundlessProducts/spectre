package backend

import (
	"fmt"
	"strconv"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// BFSBackend wraps the existing BFS explorer as a Backend.
type BFSBackend struct{}

func NewBFSBackend() *BFSBackend { return &BFSBackend{} }

func (b *BFSBackend) Name() Kind { return KindBFS }

func (b *BFSBackend) Verify(file *ast.File, opts Options) (*Result, error) {
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		return nil, fmt.Errorf("bfs: build state machine: %w", err)
	}

	// Apply param overrides.
	for k, v := range opts.Params {
		n, parseErr := strconv.ParseInt(v, 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("bfs: param %s: %q is not a valid integer: %w", k, v, parseErr)
		}
		sm.SetParam(k, n)
	}

	e := explore.NewExplorer(sm)
	if opts.MaxStates > 0 {
		e.SetMaxStates(opts.MaxStates)
	}
	if opts.MaxDepth > 0 {
		e.SetMaxDepth(opts.MaxDepth)
	}
	e.SetSilent(opts.Silent)

	er, err := e.ExploreBFS()
	if err != nil {
		return nil, fmt.Errorf("bfs: explore: %w", err)
	}

	exhaustive := er.Kind == explore.KindExhaustive
	note := ""
	if !exhaustive {
		note = fmt.Sprintf("bounded to %d states / depth %d", opts.MaxStates, opts.MaxDepth)
	}

	return &Result{
		Safe:           len(er.Violations) == 0,
		Exhaustive:     exhaustive,
		Violations:     er.Violations,
		StatesExplored: er.StatesExplored,
		MaxDepth:       er.MaxDepth,
		BackendName:    string(KindBFS),
		Note:           note,
	}, nil
}
