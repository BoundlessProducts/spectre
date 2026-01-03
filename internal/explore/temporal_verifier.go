package explore

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/eval"
	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/internal/temporal"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// TemporalVerifier verifies temporal properties over the state space
type TemporalVerifier struct {
	temporalEval *temporal.TemporalEvaluator
	hasher       *StateHasher
	file         *ast.File // For enum type registration
	stateMachine *exec.StateMachine // For checking action enabledness (fairness)
}

// NewTemporalVerifier creates a new temporal verifier
func NewTemporalVerifier(hasher *StateHasher, file *ast.File) *TemporalVerifier {
	return &TemporalVerifier{
		temporalEval: temporal.NewTemporalEvaluator(),
		hasher:       hasher,
		file:         file,
		stateMachine: nil, // Will be set if needed for fairness checking
	}
}

// SetStateMachine sets the state machine for fairness checking
func (tv *TemporalVerifier) SetStateMachine(sm *exec.StateMachine) {
	tv.stateMachine = sm
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
	// Handle nested temporal operators specially
	switch innerExpr := expr.Expr.(type) {
	case *ast.LeadsToExpr:
		// For "always (P → Q)", verify that in every state where P holds, Q eventually holds
		return tv.verifyAlwaysLeadsTo(prop, innerExpr, graph, initialStates)
	case *ast.EventuallyExpr:
		// For "always eventually P", verify that from every state, eventually P holds
		return tv.verifyAlwaysEventually(prop, innerExpr, graph, initialStates)
	default:
		// For state predicates, check all states
		return tv.verifyAlwaysStatePredicate(prop, expr.Expr, graph, initialStates)
	}
}

// verifyAlwaysStatePredicate verifies "always P" where P is a state predicate
func (tv *TemporalVerifier) verifyAlwaysStatePredicate(prop *state.TemporalPropertyInfo, predicate ast.Expr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
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
		
		holds, err := tv.evaluateProperty(predicate, s)
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

// verifyAlwaysLeadsTo verifies "always (P → Q)" - in every state where P holds, Q eventually holds
func (tv *TemporalVerifier) verifyAlwaysLeadsTo(prop *state.TemporalPropertyInfo, leadsToExpr *ast.LeadsToExpr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
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
		
		// Check if P holds in this state
		pHolds, _ := tv.evaluateProperty(leadsToExpr.Left, s)
		if pHolds {
			// P holds - verify that Q eventually holds from this state
			if !tv.canEventuallyReachP(s, leadsToExpr.Right, graph, make(map[string]bool)) {
				if violatingPState == nil {
					violatingPState = s
					violatingTrace = trace.Copy()
				}
				return
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

// verifyAlwaysEventually verifies "always eventually P" - from every state, eventually P holds
func (tv *TemporalVerifier) verifyAlwaysEventually(prop *state.TemporalPropertyInfo, eventuallyExpr *ast.EventuallyExpr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
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
		
		// Check if P eventually holds from this state
		if !tv.canEventuallyReachP(s, eventuallyExpr.Expr, graph, make(map[string]bool)) {
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
				Description:  fmt.Sprintf("Property '%s' violated: eventually P does not hold from some state", prop.Name),
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
// If P is a fairness condition (WF or SF), it filters paths to only consider fair executions
func (tv *TemporalVerifier) verifyLeadsToWithInitial(prop *state.TemporalPropertyInfo, expr *ast.LeadsToExpr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
	// Check if the left side is a fairness condition (WF or SF)
	var fairGraph *TransitionGraph = graph
	var actionName string
	var fairnessType string // "WF" or "SF"
	var err error
	
	switch leftExpr := expr.Left.(type) {
	case *ast.WFExpr:
		// Extract action name from WF expression
		actionName, err = tv.extractActionNameFromExpr(leftExpr.Target)
		if err != nil {
			return nil, fmt.Errorf("error extracting action name from WF expression: %w", err)
		}
		fairnessType = "WF"
		// Filter graph to only include fair paths
		fairGraph = tv.filterFairPaths(graph, actionName, "WF")
	case *ast.SFExpr:
		// Extract action name from SF expression
		actionName, err = tv.extractActionNameFromExpr(leftExpr.Target)
		if err != nil {
			return nil, fmt.Errorf("error extracting action name from SF expression: %w", err)
		}
		fairnessType = "SF"
		// Filter graph to only include fair paths
		fairGraph = tv.filterFairPaths(graph, actionName, "SF")
	default:
		// Not a fairness condition - use original graph and standard verification
		// (existing logic for regular leads-to)
	}
	
	// If we have a fairness condition, verify the property over the fair graph
	if fairGraph != graph {
		return tv.verifyPropertyWithFairGraph(prop, expr.Right, fairGraph, initialStates, fairnessType, actionName)
	}
	
	// Standard leads-to verification (no fairness constraint)
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
	
	// Handle temporal expressions specially (before checking visited, to allow recursion)
	switch expr := propExpr.(type) {
	case *ast.EventuallyExpr:
		// For "eventually P", check if P eventually holds from this state
		// Use a fresh visited map to allow exploring cycles
		return tv.canEventuallyReachP(from, expr.Expr, graph, make(map[string]bool))
	case *ast.LeadsToExpr:
		// For "P → Q", if P holds, Q must eventually hold
		pHolds, _ := tv.evaluateProperty(expr.Left, from)
		if pHolds {
			// P holds, must check if Q eventually holds
			return tv.canEventuallyReachP(from, expr.Right, graph, make(map[string]bool))
		}
		// P doesn't hold, so P → Q is trivially true
		return true
	}
	
	// For non-temporal expressions, check if property holds in current state
	holds, err := tv.evaluateProperty(propExpr, from)
	if err == nil && holds {
		return true
	}
	
	// Check if we've already visited this state in this path
	// If yes, we're in a cycle - check if any state in the cycle satisfies the property
	if visited[hash] {
		// We're in a cycle - need to check if property can be satisfied by continuing the cycle
		// For simple properties, if we've revisited and property doesn't hold, it might never hold
		// But for "eventually", we need to explore cycles
		// For now, return false to avoid infinite loops
		return false
	}
	visited[hash] = true
	
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

// extractActionNameFromExpr extracts the action name from an expression (usually an Ident)
func (tv *TemporalVerifier) extractActionNameFromExpr(expr ast.Expr) (string, error) {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name, nil
	default:
		return "", fmt.Errorf("expected identifier for action name, got %T", expr)
	}
}

// filterFairPaths filters the transition graph to only include paths that satisfy fairness
// For WF: removes cycles where action is continuously enabled but never executes
// For SF: removes cycles where action is enabled at least once but never executes
func (tv *TemporalVerifier) filterFairPaths(graph *TransitionGraph, actionName string, fairnessType string) *TransitionGraph {
	if tv.stateMachine == nil {
		// Cannot filter without state machine - return original graph
		return graph
	}
	
	// Create a new graph with only fair transitions
	fairGraph := NewTransitionGraph()
	
	// Copy all states
	for hash, node := range graph.States {
		fairGraph.AddState(node.State, hash)
	}
	
	// Detect cycles and check fairness
	cycles := graph.DetectCycles(tv.hasher)
	unfairCycles := make(map[string]bool) // Cycle identifier -> is unfair
	
	for _, cycle := range cycles {
		isUnfair := tv.isCycleUnfair(cycle, actionName, fairnessType)
		if isUnfair {
			// Mark all transitions in this cycle as unfair
			for _, trans := range cycle.Transitions {
				transHash := fmt.Sprintf("%s->%s", tv.hasher.HashState(trans.FromState), tv.hasher.HashState(trans.ToState))
				unfairCycles[transHash] = true
			}
		}
	}
	
	// Add only fair transitions (not in unfair cycles)
	for _, trans := range graph.Transitions {
		transHash := fmt.Sprintf("%s->%s", tv.hasher.HashState(trans.FromState), tv.hasher.HashState(trans.ToState))
		if !unfairCycles[transHash] {
			fromHash := tv.hasher.HashState(trans.FromState)
			toHash := tv.hasher.HashState(trans.ToState)
			fairGraph.AddTransition(trans, fromHash, toHash)
		}
	}
	
	return fairGraph
}

// isCycleUnfair checks if a cycle violates fairness constraints
func (tv *TemporalVerifier) isCycleUnfair(cycle *CycleInfo, actionName string, fairnessType string) bool {
	if tv.stateMachine == nil {
		return false
	}
	
	// Check if action executes in the cycle
	actionExecutes := false
	for _, trans := range cycle.Transitions {
		if trans.Action == actionName {
			actionExecutes = true
			break
		}
	}
	
	if actionExecutes {
		// Action executes - cycle is fair
		return false
	}
	
	// Action doesn't execute - check if it should have (based on fairness type)
	if fairnessType == "WF" {
		// Weak fairness: cycle is unfair if action is continuously enabled in all states
		continuouslyEnabled := true
		for _, cycleState := range cycle.States {
			enabled, err := tv.stateMachine.GetAvailableActions(cycleState)
			if err != nil {
				continuouslyEnabled = false
				break
			}
			actionEnabled := false
			for _, name := range enabled {
				if name == actionName {
					actionEnabled = true
					break
				}
			}
			if !actionEnabled {
				continuouslyEnabled = false
				break
			}
		}
		return continuouslyEnabled // Unfair if continuously enabled but never executes
	} else if fairnessType == "SF" {
		// Strong fairness: cycle is unfair if action is enabled in at least one state
		enabledAtLeastOnce := false
		for _, cycleState := range cycle.States {
			enabled, err := tv.stateMachine.GetAvailableActions(cycleState)
			if err != nil {
				continue
			}
			for _, name := range enabled {
				if name == actionName {
					enabledAtLeastOnce = true
					break
				}
			}
			if enabledAtLeastOnce {
				break
			}
		}
		return enabledAtLeastOnce // Unfair if enabled at least once but never executes
	}
	
	return false
}

// verifyPropertyWithFairGraph verifies a property over a fairness-filtered graph
func (tv *TemporalVerifier) verifyPropertyWithFairGraph(prop *state.TemporalPropertyInfo, propertyExpr ast.Expr, fairGraph *TransitionGraph, initialStates []*state.State, fairnessType string, actionName string) (*TemporalVerificationResult, error) {
	// The property expression can be various types - handle them appropriately
	switch expr := propertyExpr.(type) {
	case *ast.EventuallyExpr:
		return tv.verifyEventuallyWithFairGraph(prop, expr, fairGraph, initialStates, fairnessType, actionName)
	case *ast.AlwaysExpr:
		return tv.verifyAlwaysWithFairGraph(prop, expr, fairGraph, initialStates, fairnessType, actionName)
	default:
		// For other types, use a generic approach: verify property holds in all fair paths
		return tv.verifyGenericWithFairGraph(prop, propertyExpr, fairGraph, initialStates, fairnessType, actionName)
	}
}

// verifyEventuallyWithFairGraph verifies "eventually P" over fair paths
func (tv *TemporalVerifier) verifyEventuallyWithFairGraph(prop *state.TemporalPropertyInfo, expr *ast.EventuallyExpr, fairGraph *TransitionGraph, initialStates []*state.State, fairnessType string, actionName string) (*TemporalVerificationResult, error) {
	// Use the same logic as verifyEventuallyWithInitial but on the fair graph
	return tv.verifyEventuallyWithInitial(prop, expr, fairGraph, initialStates)
}

// verifyAlwaysWithFairGraph verifies "always P" over fair paths
func (tv *TemporalVerifier) verifyAlwaysWithFairGraph(prop *state.TemporalPropertyInfo, expr *ast.AlwaysExpr, fairGraph *TransitionGraph, initialStates []*state.State, fairnessType string, actionName string) (*TemporalVerificationResult, error) {
	// Use the same logic as verifyAlwaysWithInitial but on the fair graph
	return tv.verifyAlwaysWithInitial(prop, expr, fairGraph, initialStates)
}

// verifyGenericWithFairGraph verifies a generic property expression over fair paths
func (tv *TemporalVerifier) verifyGenericWithFairGraph(prop *state.TemporalPropertyInfo, propertyExpr ast.Expr, fairGraph *TransitionGraph, initialStates []*state.State, fairnessType string, actionName string) (*TemporalVerificationResult, error) {
	// For generic expressions, check if property holds in all states reachable via fair paths
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
		
		holds, err := tv.evaluateProperty(propertyExpr, s)
		if err != nil || !holds {
			if violatingState == nil {
				violatingState = s
				violatingTrace = trace.Copy()
			}
			return
		}
		
		// Continue exploring fair paths
		if node := fairGraph.GetStateNode(hash); node != nil {
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
				Description:  fmt.Sprintf("Property '%s' violated under %s(%s) fairness constraint", prop.Name, fairnessType, actionName),
				Trace:        violatingTrace,
			},
		}, nil
	}
	
	return &TemporalVerificationResult{
		PropertyName: prop.Name,
		Holds:        true,
	}, nil
}

