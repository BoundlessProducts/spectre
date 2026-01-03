package explore

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/eval"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/internal/temporal"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// TemporalVerifier verifies temporal properties over the state space
type TemporalVerifier struct {
	temporalEval *temporal.TemporalEvaluator
	hasher       *StateHasher
	file         *ast.File // For enum type registration
}

// NewTemporalVerifier creates a new temporal verifier
func NewTemporalVerifier(hasher *StateHasher, file *ast.File) *TemporalVerifier {
	return &TemporalVerifier{
		temporalEval: temporal.NewTemporalEvaluator(),
		hasher:       hasher,
		file:         file,
	}
}

// TemporalVerificationResult represents the result of verifying a temporal property
type TemporalVerificationResult struct {
	PropertyName string
	Holds        bool
	Violation    *TemporalViolation // nil if property holds
}

// TemporalViolation represents a temporal property violation
type TemporalViolation struct {
	PropertyName string
	Description  string
	Trace        *temporal.Trace // Counterexample trace
	Cycles       []*CycleInfo    // Cycles that prevent the property
}

// VerifyTemporalProperty verifies a temporal property against the transition graph
// It accepts initialStates from the exploration result since the graph may not preserve that information
func (tv *TemporalVerifier) VerifyTemporalProperty(prop *state.TemporalPropertyInfo, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
	// Use provided initial states, or try to find them in graph
	if len(initialStates) == 0 {
		initialStates = tv.findInitialStates(graph)
	}
	
	switch expr := prop.Expression.(type) {
	case *ast.EventuallyExpr:
		return tv.verifyEventuallyWithInitial(prop, expr, graph, initialStates)
	case *ast.AlwaysExpr:
		return tv.verifyAlwaysWithInitial(prop, expr, graph, initialStates)
	case *ast.UntilExpr:
		return tv.verifyUntilWithInitial(prop, expr, graph, initialStates)
	case *ast.LeadsToExpr:
		return tv.verifyLeadsToWithInitial(prop, expr, graph, initialStates)
	default:
		return nil, fmt.Errorf("unsupported temporal expression type: %T", expr)
	}
}

// verifyEventuallyWithInitial verifies "eventually P" starting from given initial states
func (tv *TemporalVerifier) verifyEventuallyWithInitial(prop *state.TemporalPropertyInfo, expr *ast.EventuallyExpr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
	if len(initialStates) == 0 {
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  "No initial states provided for verification",
			},
		}, nil
	}
	
	// Check if property holds in any reachable state using BFS from initial states
	visited := make(map[string]bool)
	queue := make([]*state.State, 0)
	
	// Start BFS from all initial states
	for _, initState := range initialStates {
		hash := tv.hasher.HashState(initState)
		visited[hash] = true
		queue = append(queue, initState)
	}
	
	// BFS to find if property eventually holds
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		// Check if property holds in current state
		holds, err := tv.evaluateProperty(expr.Expr, current)
		if err != nil {
			return nil, fmt.Errorf("error evaluating property: %w", err)
		}
		if holds {
			// Property holds - temporal property satisfied
			return &TemporalVerificationResult{
				PropertyName: prop.Name,
				Holds:        true,
			}, nil
		}
		
		// Explore successors
		currentHash := tv.hasher.HashState(current)
		if node := graph.GetStateNode(currentHash); node != nil {
			for _, trans := range node.Outgoing {
				toHash := tv.hasher.HashState(trans.ToState)
				if !visited[toHash] {
					visited[toHash] = true
					queue = append(queue, trans.ToState)
				}
			}
		}
	}
	
	// Property doesn't hold - find counterexample
	// Check if we're stuck in cycles that prevent reaching states where P holds
	var blockingCycles []*CycleInfo
	for _, cycle := range graph.DetectCycles(tv.hasher) {
		// Check if cycle prevents reaching P
		canReachP := false
		for _, cycleState := range cycle.States {
			holds, err := tv.evaluateProperty(expr.Expr, cycleState)
			if err == nil && holds {
				canReachP = true
				break
			}
			// Check if we can exit cycle to reach P
			cycleHash := tv.hasher.HashState(cycleState)
			if node := graph.GetStateNode(cycleHash); node != nil {
				for _, trans := range node.Outgoing {
					toHash := tv.hasher.HashState(trans.ToState)
					if graph.GetStateNode(toHash) != nil {
						if tv.canEventuallyReachP(trans.ToState, expr.Expr, graph, make(map[string]bool)) {
							canReachP = true
							break
						}
					}
				}
			}
		}
		if !canReachP {
			blockingCycles = append(blockingCycles, cycle)
		}
	}
	
	// Build counterexample trace (simplified - just show one path)
	trace := tv.buildCounterexampleTrace(initialStates, graph)
	
	return &TemporalVerificationResult{
		PropertyName: prop.Name,
		Holds:        false,
		Violation: &TemporalViolation{
			PropertyName: prop.Name,
			Description:  fmt.Sprintf("Property '%s' never becomes true in any reachable state", prop.Name),
			Trace:        trace,
			Cycles:       blockingCycles,
		},
	}, nil
}

// verifyAlwaysWithInitial verifies "always P" starting from given initial states
func (tv *TemporalVerifier) verifyAlwaysWithInitial(prop *state.TemporalPropertyInfo, expr *ast.AlwaysExpr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
	
	// Check all states reachable from initial states
	visited := make(map[string]bool)
	var violatingState *state.State
	var violatingTrace *temporal.Trace
	
	var dfs func(s *state.State, trace *temporal.Trace)
	dfs = func(s *state.State, trace *temporal.Trace) {
		hash := tv.hasher.HashState(s)
		if visited[hash] {
			return
		}
		visited[hash] = true
		
		holds, err := tv.evaluateProperty(expr.Expr, s)
		if err != nil || !holds {
			if violatingState == nil {
				violatingState = s
				violatingTrace = trace.Copy()
			}
			return
		}
		
		// Continue exploring
		if node := graph.GetStateNode(hash); node != nil {
			for _, trans := range node.Outgoing {
				newTrace := trace.Copy()
				newTrace.AddState(trans.ToState, trans.Action, trans.Args)
				dfs(trans.ToState, newTrace)
			}
		}
	}
	
	// Start from all initial states
	for _, initState := range initialStates {
		trace := temporal.NewTrace()
		trace.AddState(initState, "", nil)
		dfs(initState, trace)
	}
	
	if violatingState != nil {
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  fmt.Sprintf("Property '%s' violated in reachable state", prop.Name),
				Trace:        violatingTrace,
			},
		}, nil
	}
	
	return &TemporalVerificationResult{
		PropertyName: prop.Name,
		Holds:        true,
	}, nil
}

// verifyUntilWithInitial verifies "P until Q" starting from given initial states
func (tv *TemporalVerifier) verifyUntilWithInitial(prop *state.TemporalPropertyInfo, expr *ast.UntilExpr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
	
	visited := make(map[string]bool)
	var violatingState *state.State
	var violatingTrace *temporal.Trace
	
	var dfs func(s *state.State, trace *temporal.Trace)
	dfs = func(s *state.State, trace *temporal.Trace) {
		hash := tv.hasher.HashState(s)
		if visited[hash] {
			return
		}
		visited[hash] = true
		
		qHolds, _ := tv.evaluateProperty(expr.Right, s)
		if qHolds {
			// Q holds - property satisfied
			return
		}
		
		pHolds, err := tv.evaluateProperty(expr.Left, s)
		if err != nil || !pHolds {
			if violatingState == nil {
				violatingState = s
				violatingTrace = trace.Copy()
			}
			return
		}
		
		// Continue exploring
		if node := graph.GetStateNode(hash); node != nil {
			for _, trans := range node.Outgoing {
				newTrace := trace.Copy()
				newTrace.AddState(trans.ToState, trans.Action, trans.Args)
				dfs(trans.ToState, newTrace)
			}
		}
	}
	
	for _, initState := range initialStates {
		trace := temporal.NewTrace()
		trace.AddState(initState, "", nil)
		dfs(initState, trace)
	}
	
	if violatingState != nil {
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  fmt.Sprintf("Property '%s' violated: P doesn't hold before Q", prop.Name),
				Trace:        violatingTrace,
			},
		}, nil
	}
	
	return &TemporalVerificationResult{
		PropertyName: prop.Name,
		Holds:        true,
	}, nil
}

// verifyLeadsToWithInitial verifies "P → Q" (P leads to Q) starting from given initial states
func (tv *TemporalVerifier) verifyLeadsToWithInitial(prop *state.TemporalPropertyInfo, expr *ast.LeadsToExpr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
	
	visited := make(map[string]bool)
	var violatingPState *state.State
	var violatingTrace *temporal.Trace
	
	var dfs func(s *state.State, trace *temporal.Trace)
	dfs = func(s *state.State, trace *temporal.Trace) {
		hash := tv.hasher.HashState(s)
		if visited[hash] {
			return
		}
		visited[hash] = true
		
		pHolds, _ := tv.evaluateProperty(expr.Left, s)
		if pHolds {
			// P holds - check if Q eventually holds
			if !tv.canEventuallyReachP(s, expr.Right, graph, make(map[string]bool)) {
				if violatingPState == nil {
					violatingPState = s
					violatingTrace = trace.Copy()
				}
			}
		}
		
		// Continue exploring
		if node := graph.GetStateNode(hash); node != nil {
			for _, trans := range node.Outgoing {
				newTrace := trace.Copy()
				newTrace.AddState(trans.ToState, trans.Action, trans.Args)
				dfs(trans.ToState, newTrace)
			}
		}
	}
	
	for _, initState := range initialStates {
		trace := temporal.NewTrace()
		trace.AddState(initState, "", nil)
		dfs(initState, trace)
	}
	
	if violatingPState != nil {
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  fmt.Sprintf("Property '%s' violated: P holds but Q never becomes true", prop.Name),
				Trace:        violatingTrace,
			},
		}, nil
	}
	
	return &TemporalVerificationResult{
		PropertyName: prop.Name,
		Holds:        true,
	}, nil
}

// Helper methods

func (tv *TemporalVerifier) evaluateProperty(expr ast.Expr, s *state.State) (bool, error) {
	env := eval.NewEnvironment()
	// Register enum types
	eval.RegisterEnumTypes(env, tv.file)
	
	for varName, varValue := range s.Variables {
		env.SetVariable(varName, varValue)
	}
	
	evaluator := eval.NewEvaluator(env)
	result, err := evaluator.Eval(expr)
	if err != nil {
		return false, err
	}
	
	if pv, ok := result.(*state.PrimitiveValue); ok && pv.TypeName == "bool" && pv.BoolValue != nil {
		return *pv.BoolValue, nil
	}
	
	return false, fmt.Errorf("property did not evaluate to boolean")
}

func (tv *TemporalVerifier) findInitialStates(graph *TransitionGraph) []*state.State {
	initialStates := []*state.State{}
	
	// Find states with no incoming transitions - these are initial states
	// But if all states have incoming transitions (e.g., due to cycles), 
	// we need to track which ones were actually initial from exploration
	hasIncoming := make(map[string]bool)
	for _, node := range graph.States {
		hasIncoming[node.Hash] = len(node.Incoming) > 0
	}
	
	for _, node := range graph.States {
		if !hasIncoming[node.Hash] {
			initialStates = append(initialStates, node.State)
		}
	}
	
	// If no states found without incoming, use all states (fallback)
	// This shouldn't happen in practice, but handle it gracefully
	if len(initialStates) == 0 && len(graph.States) > 0 {
		// Return first state as fallback (should use actual initial states from exploration result)
		for _, node := range graph.States {
			initialStates = append(initialStates, node.State)
			break
		}
	}
	
	return initialStates
}

func (tv *TemporalVerifier) isReachable(fromStates []*state.State, target *state.State, graph *TransitionGraph) bool {
	targetHash := tv.hasher.HashState(target)
	visited := make(map[string]bool)
	
	var dfs func(hash string) bool
	dfs = func(hash string) bool {
		if hash == targetHash {
			return true
		}
		if visited[hash] {
			return false
		}
		visited[hash] = true
		
		if node := graph.GetStateNode(hash); node != nil {
			for _, trans := range node.Outgoing {
				toHash := tv.hasher.HashState(trans.ToState)
				if dfs(toHash) {
					return true
				}
			}
		}
		return false
	}
	
	for _, fromState := range fromStates {
		fromHash := tv.hasher.HashState(fromState)
		if dfs(fromHash) {
			return true
		}
	}
	return false
}

func (tv *TemporalVerifier) canEventuallyReachP(from *state.State, propExpr ast.Expr, graph *TransitionGraph, visited map[string]bool) bool {
	hash := tv.hasher.HashState(from)
	if visited[hash] {
		return false
	}
	visited[hash] = true
	
	// Check if property holds in current state
	holds, err := tv.evaluateProperty(propExpr, from)
	if err == nil && holds {
		return true
	}
	
	// Check successors
	if node := graph.GetStateNode(hash); node != nil {
		for _, trans := range node.Outgoing {
			if tv.canEventuallyReachP(trans.ToState, propExpr, graph, visited) {
				return true
			}
		}
	}
	
	return false
}

func (tv *TemporalVerifier) buildCounterexampleTrace(initialStates []*state.State, graph *TransitionGraph) *temporal.Trace {
	if len(initialStates) == 0 {
		return temporal.NewTrace()
	}
	
	trace := temporal.NewTrace()
	trace.AddState(initialStates[0], "", nil)
	
	// Build a simple trace by following transitions
	visited := make(map[string]bool)
	hash := tv.hasher.HashState(initialStates[0])
	visited[hash] = true
	
	maxSteps := 10 // Limit trace length
	steps := 0
	
	for steps < maxSteps {
		if node := graph.GetStateNode(hash); node != nil && len(node.Outgoing) > 0 {
			trans := node.Outgoing[0] // Take first transition
			toHash := tv.hasher.HashState(trans.ToState)
			if visited[toHash] {
				break // Hit cycle
			}
			visited[toHash] = true
			trace.AddState(trans.ToState, trans.Action, trans.Args)
			hash = toHash
			steps++
		} else {
			break
		}
	}
	
	return trace
}

