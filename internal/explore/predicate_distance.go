package explore

import (
	"container/heap"
	"math"

	"github.com/BoundlessProducts/spectre/internal/eval"
	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// LinearPredicateDistance computes a real-valued distance metric that measures
// how far a state is from violating the conjunctive invariant condition.
//
// Supported linear predicate shapes (over integer-valued expressions):
//   - e >= c  →  distance = max(0, e-c)   (0 when at boundary, positive when safe)
//   - e <= c  →  distance = max(0, c-e)
//   - e > c   →  distance = max(0, e-c-1)
//   - e < c   →  distance = max(0, c-e-1)
//   - e = c   →  distance = |e-c|
//   - e != c  →  distance = 0 if already violated, 1 if satisfied
//   - e && f  →  sum of sub-distances
//   - e || f  →  min of sub-distances
//
// Unsupported sub-expressions return math.MaxFloat64 (conservative).
func LinearPredicateDistance(s *state.State, inv *ast.InvariantDecl, file *ast.File, allFiles []*ast.File) float64 {
	env := eval.NewEnvironment()
	for _, f := range allFiles {
		eval.RegisterEnumTypes(env, f)
		eval.RegisterConstants(env, f, allFiles...)
	}
	for k, v := range s.Variables {
		env.SetVariable(k, v)
	}
	evaluator := eval.NewEvaluator(env)
	return exprDistance(inv.Condition, evaluator)
}

// exprDistance recursively computes the distance for a single expression clause.
func exprDistance(e ast.Expr, ev *eval.Evaluator) float64 {
	bin, ok := e.(*ast.BinaryExpr)
	if !ok {
		if paren, ok2 := e.(*ast.ParenExpr); ok2 {
			return exprDistance(paren.X, ev)
		}
		return math.MaxFloat64
	}

	switch bin.Op {
	case ast.And:
		// Conjunction: sum of sub-distances (both must be satisfied).
		d := exprDistance(bin.Left, ev) + exprDistance(bin.Right, ev)
		if d < 0 {
			return 0
		}
		return d

	case ast.Or:
		// Disjunction: minimum sub-distance (only one needs to hold).
		dl := exprDistance(bin.Left, ev)
		dr := exprDistance(bin.Right, ev)
		if dl < dr {
			return dl
		}
		return dr

	case ast.Geq, ast.Gt, ast.Leq, ast.Lt, ast.Eq, ast.Neq:
		// Linear comparison: evaluate both sides and compute signed difference.
		lv, err := ev.Eval(bin.Left)
		if err != nil {
			return math.MaxFloat64
		}
		rv, err := ev.Eval(bin.Right)
		if err != nil {
			return math.MaxFloat64
		}
		lInt, lok := primInt(lv)
		rInt, rok := primInt(rv)
		if !lok || !rok {
			return math.MaxFloat64
		}
		diff := float64(lInt - rInt)
		switch bin.Op {
		case ast.Geq:
			// Violated when e < c (diff < 0); distance = max(0, e-c).
			return math.Max(0, diff)
		case ast.Gt:
			// Violated when e <= c (diff <= 0); distance = max(0, e-c-1).
			return math.Max(0, diff-1)
		case ast.Leq:
			// Violated when e > c (diff > 0); distance = max(0, c-e).
			return math.Max(0, -diff)
		case ast.Lt:
			// Violated when e >= c (diff >= 0); distance = max(0, c-e-1).
			return math.Max(0, -diff-1)
		case ast.Eq:
			if diff == 0 {
				return 0
			}
			return math.Abs(diff)
		case ast.Neq:
			if diff != 0 {
				return 0
			}
			return 1
		}
	}
	return math.MaxFloat64
}

func primInt(v state.Value) (int64, bool) {
	pv, ok := v.(*state.PrimitiveValue)
	if !ok || pv.TypeName != "int" || pv.IntValue == nil {
		return 0, false
	}
	return *pv.IntValue, true
}

// ── Priority-queue BFS ────────────────────────────────────────────────────────

// priorityNode wraps an exploration node with a distance score.
type priorityNode struct {
	node     *ExplorationStateNode
	distance float64
	index    int // for heap.Interface
}

type nodeHeap []*priorityNode

func (h nodeHeap) Len() int            { return len(h) }
func (h nodeHeap) Less(i, j int) bool  { return h[i].distance < h[j].distance }
func (h nodeHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *nodeHeap) Push(x interface{}) { *h = append(*h, x.(*priorityNode)) }
func (h *nodeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// ExplorePropertyGuided runs best-first BFS prioritising states that are closest
// to violating the spec's invariants according to the linear predicate distance.
// Falls back to standard BFS when no invariant declarations are found in file.
func (e *Explorer) ExplorePropertyGuided(file *ast.File, allFiles []*ast.File) (*ExplorationResult, error) {
	invDecls := collectInvariantDecls(file)
	if len(invDecls) == 0 {
		return e.ExploreBFS()
	}

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

	pq := &nodeHeap{}
	heap.Init(pq)

	for _, s := range initialStates {
		hash := e.hasher.HashState(s)
		if e.visited[hash] {
			continue
		}
		e.visited[hash] = true
		result.ReachableStates = append(result.ReachableStates, s)
		result.TransitionGraph.AddState(s, hash)

		d := minDistance(s, invDecls, file, allFiles)
		heap.Push(pq, &priorityNode{
			node: &ExplorationStateNode{State: s, Depth: 0, Path: []*Transition{}},
			distance: d,
		})
	}

	for pq.Len() > 0 {
		if e.maxStates != -1 && result.StatesExplored >= e.maxStates {
			break
		}

		current := heap.Pop(pq).(*priorityNode).node
		result.StatesExplored++

		if e.maxDepth != -1 && current.Depth >= e.maxDepth {
			continue
		}

		actions, err := e.stateMachine.GetAvailableActionsWithArgs(current.State)
		if err != nil {
			continue
		}

		for _, awa := range actions {
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

				newPath := make([]*Transition, len(current.Path)+1)
				copy(newPath, current.Path)
				newPath[len(current.Path)] = transition

				// Validate invariants — if violated, record and keep exploring.
				violations, _ := e.stateMachine.ValidateState(nextState)
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

				d := minDistance(nextState, invDecls, file, allFiles)
				heap.Push(pq, &priorityNode{
					node: &ExplorationStateNode{
						State:  nextState,
						Depth:  current.Depth + 1,
						Path:   newPath,
					},
					distance: d,
				})
			}
		}
	}

	return result, nil
}

// minDistance returns the minimum invariant distance across all invariant declarations.
func minDistance(s *state.State, invDecls []*ast.InvariantDecl, file *ast.File, allFiles []*ast.File) float64 {
	d := math.MaxFloat64
	for _, inv := range invDecls {
		dist := LinearPredicateDistance(s, inv, file, allFiles)
		if dist < d {
			d = dist
		}
	}
	return d
}

// collectInvariantDecls returns all invariant declarations in the file.
func collectInvariantDecls(file *ast.File) []*ast.InvariantDecl {
	var invs []*ast.InvariantDecl
	for _, d := range file.Decls {
		if inv, ok := d.(*ast.InvariantDecl); ok {
			invs = append(invs, inv)
		}
		if m, ok := d.(*ast.ModuleDecl); ok {
			for _, md := range m.Decls {
				if inv, ok := md.(*ast.InvariantDecl); ok {
					invs = append(invs, inv)
				}
			}
		}
	}
	return invs
}

// isInvariantViolation reports whether an error from ExecuteAction signals an invariant violation.
func isInvariantViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return len(msg) > 0 && (containsStr(msg, "violates invariants") || containsStr(msg, "invariant") && containsStr(msg, "violated"))
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && findStr(s, sub)
}

func findStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
