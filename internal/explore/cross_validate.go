package explore

import (
	"fmt"
	"sort"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// PORValidationReport summarises the result of running BFS and POR on the same spec
// and comparing their violation and state-space findings.
type PORValidationReport struct {
	BFSStates       int
	PORStates       int
	BFSViolations   []string // invariant names found by BFS
	PORViolations   []string // invariant names found by POR
	// MissedByPOR lists invariant violations found by BFS that POR did not report.
	// A non-empty list means POR was unsound on this spec.
	MissedByPOR     []string
	// ExtraByPOR lists violations POR reported that BFS did not (should be empty).
	ExtraByPOR      []string
	// StateDiscrepancy is the fraction of BFS states that POR skipped.
	StateDiscrepancy float64
}

// FormatPORReport returns a human-readable summary of a PORValidationReport.
func FormatPORReport(r *PORValidationReport) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("BFS: %d states, %d violation(s)", r.BFSStates, len(r.BFSViolations)))
	lines = append(lines, fmt.Sprintf("POR: %d states, %d violation(s) (%.0f%% of BFS state space)",
		r.PORStates, len(r.PORViolations), (1-r.StateDiscrepancy)*100))

	if len(r.MissedByPOR) == 0 {
		lines = append(lines, "Sound: POR reported all violations found by BFS.")
	} else {
		lines = append(lines, fmt.Sprintf("UNSOUND: POR missed %d violation(s) found by BFS:", len(r.MissedByPOR)))
		for _, v := range r.MissedByPOR {
			lines = append(lines, "  - "+v)
		}
	}
	if len(r.ExtraByPOR) > 0 {
		lines = append(lines, fmt.Sprintf("WARNING: POR reported %d extra violation(s) not in BFS (unexpected):", len(r.ExtraByPOR)))
		for _, v := range r.ExtraByPOR {
			lines = append(lines, "  - "+v)
		}
	}

	result := ""
	for i, l := range lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

// CrossValidatePOR runs BFS and POR on the same spec with the same bounds and
// compares their violation sets. BFS is the ground truth: if POR misses a
// violation that BFS found, POR is unsound on this spec.
func CrossValidatePOR(sm *exec.StateMachine, file *ast.File, allFiles []*ast.File, maxStates, maxDepth int) (*PORValidationReport, error) {
	hasher := NewStateHasher()
	_ = hasher

	// BFS run (ground truth).
	bfsExplorer := NewExplorer(sm)
	bfsExplorer.SetMaxStates(maxStates)
	bfsExplorer.SetMaxDepth(maxDepth)
	bfsExplorer.SetSilent(true)
	bfsResult, err := bfsExplorer.ExploreBFS()
	if err != nil {
		return nil, fmt.Errorf("BFS failed: %w", err)
	}

	// POR run.
	porExplorer := NewExplorer(sm)
	porExplorer.SetMaxStates(maxStates)
	porExplorer.SetMaxDepth(maxDepth)
	porExplorer.SetSilent(true)
	porResult, err := porExplorer.ExplorePOR(file, allFiles)
	if err != nil {
		return nil, fmt.Errorf("POR failed: %w", err)
	}

	// Collect violation invariant names (deduplicated).
	bfsViolSet := make(map[string]bool)
	for _, v := range bfsResult.Violations {
		bfsViolSet[v.Invariant] = true
	}
	porViolSet := make(map[string]bool)
	for _, v := range porResult.Violations {
		porViolSet[v.Invariant] = true
	}

	bfsViolations := sortedKeys(bfsViolSet)
	porViolations := sortedKeys(porViolSet)

	// Violations in BFS but not POR → POR is unsound.
	var missedByPOR []string
	for inv := range bfsViolSet {
		if !porViolSet[inv] {
			missedByPOR = append(missedByPOR, inv)
		}
	}
	sort.Strings(missedByPOR)

	// Violations in POR but not BFS → unexpected (shouldn't happen).
	var extraByPOR []string
	for inv := range porViolSet {
		if !bfsViolSet[inv] {
			extraByPOR = append(extraByPOR, inv)
		}
	}
	sort.Strings(extraByPOR)

	stateDiscrepancy := 0.0
	if bfsResult.StatesExplored > 0 {
		stateDiscrepancy = float64(bfsResult.StatesExplored-porResult.StatesExplored) / float64(bfsResult.StatesExplored)
		if stateDiscrepancy < 0 {
			stateDiscrepancy = 0
		}
	}

	return &PORValidationReport{
		BFSStates:        bfsResult.StatesExplored,
		PORStates:        porResult.StatesExplored,
		BFSViolations:    bfsViolations,
		PORViolations:    porViolations,
		MissedByPOR:      missedByPOR,
		ExtraByPOR:       extraByPOR,
		StateDiscrepancy: stateDiscrepancy,
	}, nil
}

// IncrementalValidationReport summarises a comparison between incremental and fresh BFS results.
type IncrementalValidationReport struct {
	FreshStates          int
	IncrementalStates    int
	FreshViolations      []string
	IncrementalViolations []string
	// MissedByIncremental lists violations found fresh but not incrementally.
	MissedByIncremental  []string
	// ExtraByIncremental lists violations incremental found that fresh did not.
	ExtraByIncremental   []string
}

// FormatIncrementalReport returns a human-readable summary.
func FormatIncrementalReport(r *IncrementalValidationReport) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Fresh BFS: %d states, %d violation(s)", r.FreshStates, len(r.FreshViolations)))
	lines = append(lines, fmt.Sprintf("Incremental: %d states, %d violation(s)", r.IncrementalStates, len(r.IncrementalViolations)))
	if len(r.MissedByIncremental) == 0 && len(r.ExtraByIncremental) == 0 {
		lines = append(lines, "Sound: incremental result matches fresh BFS.")
	}
	if len(r.MissedByIncremental) > 0 {
		lines = append(lines, fmt.Sprintf("UNSOUND: incremental missed %d violation(s):", len(r.MissedByIncremental)))
		for _, v := range r.MissedByIncremental {
			lines = append(lines, "  - "+v)
		}
	}
	if len(r.ExtraByIncremental) > 0 {
		lines = append(lines, fmt.Sprintf("WARNING: incremental reported %d extra violation(s):", len(r.ExtraByIncremental)))
		for _, v := range r.ExtraByIncremental {
			lines = append(lines, "  - "+v)
		}
	}
	result := ""
	for i, l := range lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

// CrossValidateIncremental runs a fresh BFS and compares it against an already-computed
// incremental result. Violations in incrResult that don't appear in freshBFS (or vice
// versa) indicate an incremental unsoundness.
func CrossValidateIncremental(incrResult *ExplorationResult, sm *exec.StateMachine, maxStates, maxDepth int) (*IncrementalValidationReport, error) {
	freshExplorer := NewExplorer(sm)
	freshExplorer.SetMaxStates(maxStates)
	freshExplorer.SetMaxDepth(maxDepth)
	freshExplorer.SetSilent(true)
	freshResult, err := freshExplorer.ExploreBFS()
	if err != nil {
		return nil, fmt.Errorf("fresh BFS failed: %w", err)
	}

	freshViolSet := make(map[string]bool)
	for _, v := range freshResult.Violations {
		freshViolSet[v.Invariant] = true
	}
	incrViolSet := make(map[string]bool)
	for _, v := range incrResult.Violations {
		incrViolSet[v.Invariant] = true
	}

	var missed []string
	for inv := range freshViolSet {
		if !incrViolSet[inv] {
			missed = append(missed, inv)
		}
	}
	sort.Strings(missed)

	var extra []string
	for inv := range incrViolSet {
		if !freshViolSet[inv] {
			extra = append(extra, inv)
		}
	}
	sort.Strings(extra)

	return &IncrementalValidationReport{
		FreshStates:           freshResult.StatesExplored,
		IncrementalStates:     incrResult.StatesExplored,
		FreshViolations:       sortedKeys(freshViolSet),
		IncrementalViolations: sortedKeys(incrViolSet),
		MissedByIncremental:   missed,
		ExtraByIncremental:    extra,
	}, nil
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
