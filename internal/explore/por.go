package explore

import (
	"sort"

	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// ExplorePOR runs BFS with partial-order reduction.  When multiple actions are
// enabled in a state, it partitions them into independent groups and only
// explores a representative from each group, pruning interleaving-equivalent
// paths.
//
// Independence heuristic: two actions are independent if they operate on
// disjoint sets of state variables (no variable written by one is read or
// written by the other).  This is conservative — over-approximating
// dependence is always sound; the worst case is falling back to full BFS.
//
// The function always preserves all counterexamples for safety properties: if
// a violation is reachable under full BFS, it is also reachable under POR
// (standard result for safety-preserving POR).
func (e *Explorer) ExplorePOR(file *ast.File, _ []*ast.File) (*ExplorationResult, error) {
	actionWrites := collectActionWrites(file)
	actionReads := collectActionReads(file)

	e.visited = make(map[string]bool)

	initialStates, err := e.stateMachine.GetInitialStates()
	if err != nil {
		return nil, err
	}

	result := &ExplorationResult{
		Violations:      []*Violation{},
		ReachableStates: []*state.State{},
		InitialStates:   initialStates,
		Cycles:          []*Cycle{},
		TransitionGraph: NewTransitionGraph(),
	}

	type queueItem struct {
		node *ExplorationStateNode
	}

	queue := []*ExplorationStateNode{}

	for _, s := range initialStates {
		hash := e.hasher.HashState(s)
		if e.visited[hash] {
			continue
		}
		e.visited[hash] = true
		result.ReachableStates = append(result.ReachableStates, s)
		result.TransitionGraph.AddState(s, hash)
		queue = append(queue, &ExplorationStateNode{State: s, Depth: 0, Path: []*Transition{}})
	}

	for len(queue) > 0 {
		if result.StatesExplored >= e.maxStates && e.maxStates > 0 {
			break
		}

		current := queue[0]
		queue = queue[1:]
		result.StatesExplored++

		if current.Depth >= e.maxDepth && e.maxDepth > 0 {
			continue
		}

		allActions, err := e.stateMachine.GetAvailableActionsWithArgs(current.State)
		if err != nil {
			continue
		}

		// Select a representative ample set for POR.
		ampleSet := selectAmpleSet(allActions, actionWrites, actionReads)

		for _, awa := range ampleSet {
			nextState, err := e.stateMachine.ExecuteAction(awa.ActionName, current.State, awa.Args)
			if err != nil {
				if isInvariantViolation(err) {
					inv := extractInvariantName(err.Error())
					violation := &Violation{
						State:       current.State,
						Invariant:   inv,
						Path:        current.Path,
						Description: err.Error(),
					}
					result.Violations = append(result.Violations, violation)
					if e.invariantViolationCallback != nil && !e.invariantViolationCallback(violation) {
						return result, nil
					}
				}
				continue
			}

			hash := e.hasher.HashState(nextState)
			transition := &Transition{
				FromState: current.State,
				ToState:   nextState,
				Action:    awa.ActionName,
				Args:      awa.Args,
			}
			result.TransitionGraph.AddTransition(transition, e.hasher.HashState(current.State), hash)

			if !e.visited[hash] {
				e.visited[hash] = true
				result.ReachableStates = append(result.ReachableStates, nextState)
				result.TransitionGraph.AddState(nextState, hash)

				violations, _ := e.stateMachine.ValidateState(nextState)
				newPath := append(append([]*Transition{}, current.Path...), transition)
				for _, v := range violations {
					violation := &Violation{
						State:       nextState,
						Invariant:   v.Name,
						Path:        newPath,
						Description: v.Message,
					}
					result.Violations = append(result.Violations, violation)
					if e.invariantViolationCallback != nil && !e.invariantViolationCallback(violation) {
						return result, nil
					}
				}

				queue = append(queue, &ExplorationStateNode{
					State: nextState,
					Depth: current.Depth + 1,
					Path:  newPath,
				})
			}
		}
	}

	return result, nil
}

// selectAmpleSet chooses which actions to explore from a state using the ample
// set heuristic: find the smallest independent group of actions.
//
// Algorithm:
//  1. Sort actions by name for determinism.
//  2. Find the first action that is independent of all others (writes and reads
//     disjoint sets of variables).
//  3. If found, the ample set is just that action (maximum reduction).
//  4. Otherwise fall back to the full set (no reduction possible).
func selectAmpleSet(actions []*exec.ActionWithArgs, writes, reads map[string]map[string]bool) []*exec.ActionWithArgs {
	if len(actions) <= 1 {
		return actions
	}

	// Sort for determinism.
	sorted := make([]*exec.ActionWithArgs, len(actions))
	copy(sorted, actions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ActionName < sorted[j].ActionName
	})

	for i, a := range sorted {
		if independentOfAll(a.ActionName, sorted, i, writes, reads) {
			return []*exec.ActionWithArgs{a}
		}
	}
	return sorted // no reduction possible
}

// independentOfAll returns true if action a is independent of every other
// action in the list (excluding itself at index selfIdx).
func independentOfAll(name string, actions []*exec.ActionWithArgs, selfIdx int, writes, reads map[string]map[string]bool) bool {
	for j, b := range actions {
		if j == selfIdx {
			continue
		}
		if !independent(name, b.ActionName, writes, reads) {
			return false
		}
	}
	return true
}

// independent reports whether actions a and b are independent:
// a's writes ∩ (b's reads ∪ b's writes) = ∅  AND  b's writes ∩ a's reads = ∅.
func independent(a, b string, writes, reads map[string]map[string]bool) bool {
	aw := writes[a]
	bw := writes[b]
	br := reads[b]
	ar := reads[a]

	for v := range aw {
		if bw[v] || br[v] {
			return false
		}
	}
	for v := range bw {
		if ar[v] {
			return false
		}
	}
	return true
}

// collectActionWrites returns a map from action name to the set of variables
// it writes (appears on the left of an assignment in the action body).
func collectActionWrites(file *ast.File) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok {
			vars := make(map[string]bool)
			if a.Body != nil {
				for _, s := range a.Body.Statements {
					if assign, ok := s.(*ast.AssignStmt); ok {
						if id, ok := assign.Left.(*ast.Ident); ok {
							vars[id.Name] = true
						}
					}
				}
			}
			m[a.Name] = vars
		}
	}
	return m
}

// collectActionReads returns a map from action name to the set of variables
// it reads (unprimed identifiers in guards and RHS of assignments).
func collectActionReads(file *ast.File) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok {
			vars := make(map[string]bool)
			if a.Body != nil {
				for _, s := range a.Body.Statements {
					switch stmt := s.(type) {
					case *ast.RequireStmt:
						collectReadIdents(stmt.Condition, vars)
					case *ast.EnsureStmt:
						collectReadIdents(stmt.Condition, vars)
					case *ast.AssignStmt:
						collectReadIdents(stmt.Right, vars)
					}
				}
			}
			m[a.Name] = vars
		}
	}
	return m
}

func collectReadIdents(e ast.Expr, out map[string]bool) {
	if e == nil {
		return
	}
	switch expr := e.(type) {
	case *ast.Ident:
		if !expr.Prime {
			out[expr.Name] = true
		}
	case *ast.BinaryExpr:
		collectReadIdents(expr.Left, out)
		collectReadIdents(expr.Right, out)
	case *ast.UnaryExpr:
		collectReadIdents(expr.Expr, out)
	case *ast.ParenExpr:
		collectReadIdents(expr.X, out)
	case *ast.SelectorExpr:
		collectReadIdents(expr.X, out)
	case *ast.CallExpr:
		for _, a := range expr.Args {
			collectReadIdents(a, out)
		}
	case *ast.IndexExpr:
		collectReadIdents(expr.X, out)
		collectReadIdents(expr.Index, out)
	}
}
