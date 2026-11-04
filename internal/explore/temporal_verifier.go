package explore

import (
	"fmt"
	"sort"

	"github.com/BoundlessProducts/spectre/internal/eval"
	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/internal/temporal"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// ViolationCallback is called immediately when a temporal violation is detected
type ViolationCallback func(result *TemporalVerificationResult)

// TemporalVerifier verifies temporal properties over the state space
type TemporalVerifier struct {
	temporalEval     *temporal.TemporalEvaluator
	hasher           *StateHasher
	file             *ast.File // For enum type registration
	stateMachine     *exec.StateMachine // For checking action enabledness (fairness)
	violationCallback ViolationCallback // Optional callback for immediate violation reporting
}

// NewTemporalVerifier creates a new temporal verifier
func NewTemporalVerifier(hasher *StateHasher, file *ast.File) *TemporalVerifier {
	return &TemporalVerifier{
		temporalEval: temporal.NewTemporalEvaluator(),
		hasher:       hasher,
		file:         file,
		stateMachine: nil, // Will be set if needed for fairness checking
		violationCallback: nil, // Will be set if needed for immediate reporting
	}
}

// SetStateMachine sets the state machine for fairness checking
func (tv *TemporalVerifier) SetStateMachine(sm *exec.StateMachine) {
	tv.stateMachine = sm
}

// SetViolationCallback sets a callback that will be called immediately when a violation is detected
func (tv *TemporalVerifier) SetViolationCallback(callback ViolationCallback) {
	tv.violationCallback = callback
}

// reportViolationImmediately calls the violation callback if set, with the violation details
func (tv *TemporalVerifier) reportViolationImmediately(prop *state.TemporalPropertyInfo, desc string, violatingTrace *temporal.Trace) {
	if tv.violationCallback != nil {
		result := &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  desc,
				Trace:        violatingTrace,
			},
		}
		tv.violationCallback(result)
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
	
	// Property doesn't hold — build a lasso counterexample showing a reachable cycle
	// where P never becomes true.
	trace := tv.buildLassoCE(initialStates, expr.Expr, graph)

	// Identify blocking cycles for diagnostic info.
	var blockingCycles []*CycleInfo
	for _, cycle := range graph.DetectCycles(tv.hasher) {
		anyHolds := false
		for _, cycleState := range cycle.States {
			if holds, err := tv.evaluateProperty(expr.Expr, cycleState); err == nil && holds {
				anyHolds = true
				break
			}
		}
		if !anyHolds {
			blockingCycles = append(blockingCycles, cycle)
		}
	}

	desc := fmt.Sprintf("Property '%s' never becomes true in any reachable state", prop.Name)
	if prop.Description != "" {
		desc = fmt.Sprintf("Property '%s' violated: (%s) property never becomes true in any reachable state", prop.Name, prop.Description)
	}
	return &TemporalVerificationResult{
		PropertyName: prop.Name,
		Holds:        false,
		Violation: &TemporalViolation{
			PropertyName: prop.Name,
			Description:  desc,
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
				// Report violation immediately
				desc := fmt.Sprintf("Property '%s' violated in reachable state", prop.Name)
				if prop.Description != "" {
					desc = fmt.Sprintf("Property '%s' violated: (%s) property does not hold in some reachable state", prop.Name, prop.Description)
				}
				tv.reportViolationImmediately(prop, desc, violatingTrace)
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
		desc := fmt.Sprintf("Property '%s' violated in reachable state", prop.Name)
		if prop.Description != "" {
			desc = fmt.Sprintf("Property '%s' violated: (%s) property does not hold in some reachable state", prop.Name, prop.Description)
		}
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  desc,
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
			// P holds - verify that Q eventually holds from this state on ALL paths
			// For "always (P → eventually Q)", we need to check that Q eventually holds
			// on all paths from states where P holds.
			// This is violated if there exists a path where Q never holds.
			// 
			// We check this by seeing if Q can be reached from ALL successors, or
			// if there exists a successor from which Q cannot be reached.
			canReachQ := tv.canEventuallyReachP(s, leadsToExpr.Right, graph, make(map[string]bool))
			if !canReachQ {
				// Cannot reach Q from this state - violation
				if violatingPState == nil {
					violatingPState = s
					violatingTrace = trace.Copy()
					// Report violation immediately
					desc := fmt.Sprintf("Property '%s' violated: P holds but Q never becomes true", prop.Name)
					if prop.Description != "" {
						desc = fmt.Sprintf("Property '%s' violated: (%s) P holds but Q never becomes true", prop.Name, prop.Description)
					}
					tv.reportViolationImmediately(prop, desc, violatingTrace)
				}
				return
			}
			
			// Can reach Q, but need to check if there's a path where we get stuck
			// in a cycle before reaching Q. Check if there's a blocking cycle.
			// This is called for unfiltered graphs, so isFairGraph = false
			if tv.hasBlockingCycle(s, leadsToExpr.Right, graph, false) {
				if violatingPState == nil {
					violatingPState = s
					violatingTrace = trace.Copy()
					// Report violation immediately
					desc := fmt.Sprintf("Property '%s' violated: P holds but Q never becomes true", prop.Name)
					if prop.Description != "" {
						desc = fmt.Sprintf("Property '%s' violated: (%s) P holds but Q never becomes true", prop.Name, prop.Description)
					}
					tv.reportViolationImmediately(prop, desc, violatingTrace)
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
		desc := fmt.Sprintf("Property '%s' violated: P holds but Q never becomes true", prop.Name)
		if prop.Description != "" {
			desc = fmt.Sprintf("Property '%s' violated: (%s) P holds but Q never becomes true", prop.Name, prop.Description)
		}
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  desc,
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
				// Report violation immediately
				desc := fmt.Sprintf("Property '%s' violated: eventually P does not hold from some state", prop.Name)
				if prop.Description != "" {
					desc = fmt.Sprintf("Property '%s' violated: (%s) eventually P does not hold from some state", prop.Name, prop.Description)
				}
				tv.reportViolationImmediately(prop, desc, violatingTrace)
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
		desc := fmt.Sprintf("Property '%s' violated: eventually P does not hold from some state", prop.Name)
		if prop.Description != "" {
			desc = fmt.Sprintf("Property '%s' violated: (%s) eventually P does not hold from some state", prop.Name, prop.Description)
		}
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  desc,
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
				// Report violation immediately
				desc := fmt.Sprintf("Property '%s' violated: P doesn't hold before Q", prop.Name)
				if prop.Description != "" {
					desc = fmt.Sprintf("Property '%s' violated: (%s) P doesn't hold before Q", prop.Name, prop.Description)
				}
				tv.reportViolationImmediately(prop, desc, violatingTrace)
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
		desc := fmt.Sprintf("Property '%s' violated: P doesn't hold before Q", prop.Name)
		if prop.Description != "" {
			desc = fmt.Sprintf("Property '%s' violated: (%s) P doesn't hold before Q", prop.Name, prop.Description)
		}
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  desc,
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
					// Report violation immediately
					desc := fmt.Sprintf("Property '%s' violated: P holds but Q never becomes true", prop.Name)
					if prop.Description != "" {
						desc = fmt.Sprintf("Property '%s' violated: (%s) P holds but Q never becomes true", prop.Name, prop.Description)
					}
					tv.reportViolationImmediately(prop, desc, violatingTrace)
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
	eval.RegisterEnumTypes(env, tv.file)
	eval.RegisterConstants(env, tv.file)

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
	
	// Sort state hashes for deterministic iteration order
	stateHashes := make([]string, 0, len(graph.States))
	for hash := range graph.States {
		stateHashes = append(stateHashes, hash)
	}
	sort.Strings(stateHashes)
	
	for _, hash := range stateHashes {
		node := graph.States[hash]
		hasIncoming[node.Hash] = len(node.Incoming) > 0
	}
	
	for _, hash := range stateHashes {
		node := graph.States[hash]
		if !hasIncoming[node.Hash] {
			initialStates = append(initialStates, node.State)
		}
	}
	
	// If no states found without incoming, use all states (fallback)
	// This shouldn't happen in practice, but handle it gracefully
	if len(initialStates) == 0 && len(stateHashes) > 0 {
		// Return first state as fallback (should use actual initial states from exploration result)
		initialStates = append(initialStates, graph.States[stateHashes[0]].State)
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

// hasBlockingCycle checks if there exists an infinite path from the given state
// where the property never holds. For "always (P → eventually Q)", if such a path
// exists, the property is violated because Q never holds on that path.
// 
// For "always (P → eventually Q)", this checks if there exists a path where Q never holds.
// This is true if there exists a reachable cycle where Q never holds, even if there
// also exists another path where Q does hold.
func (tv *TemporalVerifier) hasBlockingCycle(from *state.State, propExpr ast.Expr, graph *TransitionGraph, isFairGraph bool) bool {
	fromHash := tv.hasher.HashState(from)
	
	// First, check for self-loops explicitly (they might not be detected by DetectCycles)
	// A self-loop where the property never holds is a blocking cycle if reachable
	for _, trans := range graph.Transitions {
		fromTransHash := tv.hasher.HashState(trans.FromState)
		toTransHash := tv.hasher.HashState(trans.ToState)
		if fromTransHash == toTransHash {
			// Self-loop found - check if property can ever be reached.
			// Use canEventuallyReachP so temporal exprs (EventuallyExpr) are handled correctly.
			if !tv.canEventuallyReachP(trans.FromState, propExpr, graph, make(map[string]bool)) {
				// Property doesn't hold in the self-loop state
				// Check if this state is reachable from 'from'
				visited := make(map[string]bool)
				queue := []*state.State{from}
				visited[fromHash] = true
				
				stateReachable := false
				for len(queue) > 0 && !stateReachable {
					current := queue[0]
					queue = queue[1:]
					
					currentHash := tv.hasher.HashState(current)
					if currentHash == fromTransHash {
						stateReachable = true
						break
					}
					
					// Explore successors
					if node := graph.GetStateNode(currentHash); node != nil {
						for _, t := range node.Outgoing {
							toHash := tv.hasher.HashState(t.ToState)
							if !visited[toHash] {
								visited[toHash] = true
								queue = append(queue, t.ToState)
							}
						}
					}
				}
				
				if stateReachable {
					// Self-loop is reachable - check if it can be exited (for fair graphs only)
					if isFairGraph {
						// For fair graphs, check if the self-loop can be exited to reach the property
						canExitToProperty := false
						if node := graph.GetStateNode(fromTransHash); node != nil {
							for _, exitTrans := range node.Outgoing {
								toHash := tv.hasher.HashState(exitTrans.ToState)
								if toHash != fromTransHash {
									// Exit transition found - check if we can reach property from exit state
									if tv.canEventuallyReachP(exitTrans.ToState, propExpr, graph, make(map[string]bool)) {
										canExitToProperty = true
										break
									}
								}
							}
						}
						if !canExitToProperty {
							return true
						}
					} else {
						// For unfiltered graphs, any reachable self-loop where property never holds is blocking
						return true
					}
				}
			}
		}
	}
	
	// Find all cycles in the graph (multi-state cycles)
	cycles := graph.DetectCycles(tv.hasher)
	
	// Check each cycle to see if:
	// 1. Property never holds in any state in the cycle
	// 2. Cycle is reachable from 'from'
	// If such a cycle exists, there's a path where property never holds (violation)
	for _, cycle := range cycles {
		// Check if property holds in any state in the cycle.
		// For temporal expressions (e.g. eventually(!pending)), use canEventuallyReachP
		// so the inner expression is unwrapped correctly.
		propertyHoldsInCycle := false
		for _, cycleState := range cycle.States {
			if tv.canEventuallyReachP(cycleState, propExpr, graph, make(map[string]bool)) {
				propertyHoldsInCycle = true
				break
			}
		}
		
		// If property never holds in this cycle, check if cycle is reachable from 'from'
		if !propertyHoldsInCycle {
			// Use BFS to check if cycle is reachable from 'from'
			visited := make(map[string]bool)
			queue := []*state.State{from}
			visited[fromHash] = true
			
			cycleReachable := false
			for len(queue) > 0 && !cycleReachable {
				current := queue[0]
				queue = queue[1:]
				
				currentHash := tv.hasher.HashState(current)
				
				// Check if current state is in the cycle
				for _, cycleState := range cycle.States {
					if tv.hasher.HashState(cycleState) == currentHash {
						cycleReachable = true
						break
					}
				}
				
				if cycleReachable {
					break
				}
				
				// Explore successors
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
			
			if cycleReachable {
				// Cycle is reachable - for fair graphs, check if it can be exited
				if isFairGraph {
					// For fair graphs, check if the cycle can be exited to reach the property
					canExitToProperty := false
					for _, cycleState := range cycle.States {
						cycleStateHash := tv.hasher.HashState(cycleState)
						if node := graph.GetStateNode(cycleStateHash); node != nil {
							for _, trans := range node.Outgoing {
								toHash := tv.hasher.HashState(trans.ToState)
								// Check if this transition exits the cycle
								isExitTransition := true
								for _, cs := range cycle.States {
									if tv.hasher.HashState(cs) == toHash {
										isExitTransition = false
										break
									}
								}
								if isExitTransition {
									// Exit transition found - check if we can reach property from exit state
									if tv.canEventuallyReachP(trans.ToState, propExpr, graph, make(map[string]bool)) {
										canExitToProperty = true
										break
									}
								}
							}
							if canExitToProperty {
								break
							}
						}
					}
					if !canExitToProperty {
						return true
					}
				} else {
					// For unfiltered graphs, any reachable cycle where property never holds is blocking
					return true
				}
			}
		}
	}
	
	return false
}

// buildLassoCE builds a lasso counterexample for a liveness violation:
// a finite prefix from an initial state to a cycle entry, followed by the cycle itself.
// The cycle contains no state where violatedProp holds (if provided).
// Falls back to a simple linear trace when no suitable cycle is found.
func (tv *TemporalVerifier) buildLassoCE(initialStates []*state.State, violatedProp ast.Expr, graph *TransitionGraph) *temporal.Trace {
	if len(initialStates) == 0 {
		return temporal.NewTrace()
	}

	// BFS from initial states to find shortest prefix to any cycle entry.
	// Collect all states that are part of a blocking cycle (no P-satisfying state).
	cycles := graph.DetectCycles(tv.hasher)

	// Identify blocking cycle states (where violatedProp never holds in the cycle).
	blockingCycleStates := make(map[string]*CycleInfo)
	for _, cycle := range cycles {
		if violatedProp != nil {
			anyHolds := false
			for _, s := range cycle.States {
				if holds, err := tv.evaluateProperty(violatedProp, s); err == nil && holds {
					anyHolds = true
					break
				}
			}
			if anyHolds {
				continue // This cycle contains a P-state — not a blocking cycle.
			}
		}
		for _, s := range cycle.States {
			h := tv.hasher.HashState(s)
			if _, exists := blockingCycleStates[h]; !exists {
				blockingCycleStates[h] = cycle
			}
		}
	}

	// BFS to find shortest path from any initial state to any blocking cycle state.
	type bfsEntry struct {
		state  *state.State
		parent *bfsEntry
		action string
		args   []state.Value
	}
	visited := make(map[string]*bfsEntry)
	queue := make([]*bfsEntry, 0, len(initialStates))
	for _, s := range initialStates {
		h := tv.hasher.HashState(s)
		if _, seen := visited[h]; !seen {
			e := &bfsEntry{state: s}
			visited[h] = e
			queue = append(queue, e)
		}
	}

	var cycleEntry *bfsEntry
	var blockingCycle *CycleInfo
	for len(queue) > 0 && cycleEntry == nil {
		cur := queue[0]
		queue = queue[1:]
		h := tv.hasher.HashState(cur.state)
		if cycle, ok := blockingCycleStates[h]; ok {
			cycleEntry = cur
			blockingCycle = cycle
			break
		}
		if node := graph.GetStateNode(h); node != nil {
			for _, trans := range node.Outgoing {
				toH := tv.hasher.HashState(trans.ToState)
				if _, seen := visited[toH]; !seen {
					child := &bfsEntry{state: trans.ToState, parent: cur, action: trans.Action, args: trans.Args}
					visited[toH] = child
					queue = append(queue, child)
				}
			}
		}
	}

	trace := temporal.NewTrace()

	if cycleEntry == nil {
		// No blocking cycle reachable — fall back to a simple linear trace.
		trace.AddState(initialStates[0], "", nil)
		visited2 := make(map[string]bool)
		h := tv.hasher.HashState(initialStates[0])
		visited2[h] = true
		for steps := 0; steps < 10; steps++ {
			if node := graph.GetStateNode(h); node != nil && len(node.Outgoing) > 0 {
				trans := node.Outgoing[0]
				toH := tv.hasher.HashState(trans.ToState)
				if visited2[toH] {
					break
				}
				visited2[toH] = true
				trace.AddState(trans.ToState, trans.Action, trans.Args)
				h = toH
			} else {
				break
			}
		}
		return trace
	}

	// Reconstruct the prefix path from initial state to cycle entry.
	prefix := make([]*bfsEntry, 0)
	for e := cycleEntry; e != nil; e = e.parent {
		prefix = append(prefix, e)
	}
	// Reverse to get root → cycle-entry order.
	for i, j := 0, len(prefix)-1; i < j; i, j = i+1, j-1 {
		prefix[i], prefix[j] = prefix[j], prefix[i]
	}
	trace.AddState(prefix[0].state, "", nil)
	for _, e := range prefix[1:] {
		trace.AddState(e.state, e.action, e.args)
	}

	// Append the cycle itself (lasso loop).
	if blockingCycle != nil {
		for _, trans := range blockingCycle.Transitions {
			trace.AddState(trans.ToState, trans.Action, trans.Args)
		}
	}

	return trace
}

// buildCounterexampleTrace builds a counterexample trace for a liveness violation.
// It delegates to buildLassoCE with no property filter (caller already knows no state satisfies the prop).
func (tv *TemporalVerifier) buildCounterexampleTrace(initialStates []*state.State, graph *TransitionGraph) *temporal.Trace {
	return tv.buildLassoCE(initialStates, nil, graph)
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
// 
// The filtering works by:
// 1. Detecting all cycles (including self-loops)
// 2. Identifying cycles that violate fairness (unfair cycles)
// 3. Removing all transitions that are part of unfair cycles
// 4. This ensures that fair executions cannot get stuck in unfair cycles
func (tv *TemporalVerifier) filterFairPaths(graph *TransitionGraph, actionName string, fairnessType string) *TransitionGraph {
	if tv.stateMachine == nil {
		// Cannot filter without state machine - return original graph
		return graph
	}
	
	// Create a new graph with only fair transitions
	fairGraph := NewTransitionGraph()
	
	// Copy all states (sort for deterministic order)
	stateHashes := make([]string, 0, len(graph.States))
	for hash := range graph.States {
		stateHashes = append(stateHashes, hash)
	}
	sort.Strings(stateHashes)
	
	for _, hash := range stateHashes {
		node := graph.States[hash]
		fairGraph.AddState(node.State, hash)
	}
	
	// Build a map of all transitions for efficient lookup
	allTransitions := make(map[string]*Transition)
	for _, trans := range graph.Transitions {
		fromHash := tv.hasher.HashState(trans.FromState)
		toHash := tv.hasher.HashState(trans.ToState)
		transHash := fmt.Sprintf("%s->%s", fromHash, toHash)
		allTransitions[transHash] = trans
	}
	
	// Helper function to create a unique transition hash including the action
	getTransitionHash := func(trans *Transition) string {
		fromHash := tv.hasher.HashState(trans.FromState)
		toHash := tv.hasher.HashState(trans.ToState)
		return fmt.Sprintf("%s-[%s]->%s", fromHash, trans.Action, toHash)
	}
	
	// Detect cycles and check fairness
	cycles := graph.DetectCycles(tv.hasher)
	unfairTransitions := make(map[string]bool) // Transition hash -> is unfair
	
	// Check all detected cycles
	for _, cycle := range cycles {
		isUnfair := tv.isCycleUnfair(cycle, actionName, fairnessType)
		if isUnfair {
			// Mark all transitions in this cycle as unfair
			for _, trans := range cycle.Transitions {
				transHash := getTransitionHash(trans)
				unfairTransitions[transHash] = true
			}
		}
	}
	
	// Also check for self-loops explicitly (they might not be detected as cycles by DFS)
	// Self-loops are cycles of length 1 - they need special handling
	// IMPORTANT: We must check self-loops BEFORE adding transitions to fairGraph
	// because self-loops might not be detected by DetectCycles if DFS doesn't explore them
	for _, trans := range graph.Transitions {
		fromHash := tv.hasher.HashState(trans.FromState)
		toHash := tv.hasher.HashState(trans.ToState)
		if fromHash == toHash {
			// Self-loop detected - check if it's unfair
			// Create a temporary cycle info for the self-loop
			cycle := &CycleInfo{
				States:      []*state.State{trans.FromState},
				Transitions: []*Transition{trans},
				StartState:  trans.FromState,
			}
			// For self-loops, we need to check if the action is enabled in the state
			// If the action is continuously enabled but never executes, the self-loop is unfair
			if tv.isCycleUnfair(cycle, actionName, fairnessType) {
				transHash := getTransitionHash(trans)
				unfairTransitions[transHash] = true
			}
		}
	}
	
	// Add only fair transitions (not in unfair cycles)
	for _, trans := range graph.Transitions {
		transHash := getTransitionHash(trans)
		
		if !unfairTransitions[transHash] {
			fromHash := tv.hasher.HashState(trans.FromState)
			toHash := tv.hasher.HashState(trans.ToState)
			fairGraph.AddTransition(trans, fromHash, toHash)
		}
	}
	
	// Iteratively remove unfair cycles until the fair graph reaches a fixed point.
	// Each iteration either removes at least one transition (strictly shrinking the graph)
	// or terminates — so convergence is guaranteed.
	for {
		additionalUnfairTransitions := make(map[string]bool)
		
		// Check all cycles in the current fair graph
		fairCycles := fairGraph.DetectCycles(tv.hasher)
		for _, cycle := range fairCycles {
			if tv.isCycleUnfair(cycle, actionName, fairnessType) {
				// Found an unfair cycle - mark all its transitions as unfair
				for _, trans := range cycle.Transitions {
					transHash := getTransitionHash(trans)
					additionalUnfairTransitions[transHash] = true
				}
			}
		}
		
		// Also check for self-loops in the fair graph explicitly
		for _, trans := range fairGraph.Transitions {
			fromHash := tv.hasher.HashState(trans.FromState)
			toHash := tv.hasher.HashState(trans.ToState)
			if fromHash == toHash {
				// Self-loop in fair graph - check if it's unfair
				cycle := &CycleInfo{
					States:      []*state.State{trans.FromState},
					Transitions: []*Transition{trans},
					StartState:  trans.FromState,
				}
				if tv.isCycleUnfair(cycle, actionName, fairnessType) {
					transHash := getTransitionHash(trans)
					additionalUnfairTransitions[transHash] = true
				}
			}
		}
		
		// If we found additional unfair transitions, merge them and rebuild
		if len(additionalUnfairTransitions) > 0 {
			// Merge with existing unfair transitions
			for transHash := range additionalUnfairTransitions {
				unfairTransitions[transHash] = true
			}
			// Rebuild the fair graph from the original graph
			fairGraph = NewTransitionGraph()
			// Copy all states (sort for deterministic order)
			rebuildStateHashes := make([]string, 0, len(graph.States))
			for hash := range graph.States {
				rebuildStateHashes = append(rebuildStateHashes, hash)
			}
			sort.Strings(rebuildStateHashes)
			for _, hash := range rebuildStateHashes {
				node := graph.States[hash]
				fairGraph.AddState(node.State, hash)
			}
			// Add only fair transitions
			for _, trans := range graph.Transitions {
				transHash := getTransitionHash(trans)
				
				if !unfairTransitions[transHash] {
					fromHash := tv.hasher.HashState(trans.FromState)
					toHash := tv.hasher.HashState(trans.ToState)
					fairGraph.AddTransition(trans, fromHash, toHash)
				}
			}
		} else {
			// No more unfair cycles found - we're done
			break
		}
	}
	
	return fairGraph
}

// isCycleUnfair checks if a cycle violates fairness constraints
// A cycle is unfair if:
// - For WF: the action is continuously enabled in all states of the cycle but never executes
// - For SF: the action is enabled in at least one state of the cycle but never executes
func (tv *TemporalVerifier) isCycleUnfair(cycle *CycleInfo, actionName string, fairnessType string) bool {
	if tv.stateMachine == nil {
		return false
	}
	
	// Ensure cycle has states and transitions
	if len(cycle.States) == 0 || len(cycle.Transitions) == 0 {
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
		// This means the action must be enabled in EVERY state of the cycle
		continuouslyEnabled := true
		for _, cycleState := range cycle.States {
			if cycleState == nil {
				continuouslyEnabled = false
				break
			}
			enabled, err := tv.stateMachine.GetAvailableActions(cycleState)
			if err != nil {
				// If we can't determine enabled actions, assume not continuously enabled
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
				// Action not enabled in this state - not continuously enabled
				continuouslyEnabled = false
				break
			}
		}
		// Unfair if continuously enabled but never executes
		return continuouslyEnabled
	} else if fairnessType == "SF" {
		// Strong fairness: cycle is unfair if action is enabled in at least one state
		// but never executes in the cycle
		enabledAtLeastOnce := false
		for _, cycleState := range cycle.States {
			enabled, err := tv.stateMachine.GetAvailableActions(cycleState)
			if err != nil {
				// Skip states where we can't determine enabled actions
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
		// Unfair if enabled at least once but never executes
		return enabledAtLeastOnce
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
	// We need to pass a flag indicating this is a fair graph
	// For now, we'll modify verifyAlwaysLeadsTo to accept a fairGraph flag
	// But to avoid changing the signature, we'll use a different approach:
	// Check if the graph passed is the same as the original (not filtered)
	// This is a workaround - in the future we could add a field to TransitionGraph
	return tv.verifyAlwaysWithInitialOnFairGraph(prop, expr, fairGraph, initialStates)
}

// verifyAlwaysWithInitialOnFairGraph is like verifyAlwaysWithInitial but for fair graphs
func (tv *TemporalVerifier) verifyAlwaysWithInitialOnFairGraph(prop *state.TemporalPropertyInfo, expr *ast.AlwaysExpr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
	// Handle nested temporal operators specially
	switch innerExpr := expr.Expr.(type) {
	case *ast.LeadsToExpr:
		// For "always (P → Q)", verify that in every state where P holds, Q eventually holds
		return tv.verifyAlwaysLeadsToOnFairGraph(prop, innerExpr, graph, initialStates)
	case *ast.EventuallyExpr:
		// For "always eventually P", verify that from every state, eventually P holds
		return tv.verifyAlwaysEventually(prop, innerExpr, graph, initialStates)
	default:
		// For state predicates, check all states
		return tv.verifyAlwaysStatePredicate(prop, expr.Expr, graph, initialStates)
	}
}

// verifyAlwaysLeadsToOnFairGraph is like verifyAlwaysLeadsTo but for fair graphs
func (tv *TemporalVerifier) verifyAlwaysLeadsToOnFairGraph(prop *state.TemporalPropertyInfo, leadsToExpr *ast.LeadsToExpr, graph *TransitionGraph, initialStates []*state.State) (*TemporalVerificationResult, error) {
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
			// P holds - verify that Q eventually holds from this state on ALL paths
			canReachQ := tv.canEventuallyReachP(s, leadsToExpr.Right, graph, make(map[string]bool))
			if !canReachQ {
				if violatingPState == nil {
					violatingPState = s
					violatingTrace = trace.Copy()
					// Report violation immediately
					desc := fmt.Sprintf("Property '%s' violated: P holds but Q never becomes true", prop.Name)
					if prop.Description != "" {
						desc = fmt.Sprintf("Property '%s' violated: (%s) P holds but Q never becomes true", prop.Name, prop.Description)
					}
					tv.reportViolationImmediately(prop, desc, violatingTrace)
				}
				return
			}
			
			// Can reach Q, but need to check if there's a path where we get stuck
			// in a cycle before reaching Q. Check if there's a blocking cycle.
			// This is a fair graph, so cycles that can be exited are not blocking
			if tv.hasBlockingCycle(s, leadsToExpr.Right, graph, true) {
				if violatingPState == nil {
					violatingPState = s
					violatingTrace = trace.Copy()
					// Report violation immediately
					desc := fmt.Sprintf("Property '%s' violated: P holds but Q never becomes true", prop.Name)
					if prop.Description != "" {
						desc = fmt.Sprintf("Property '%s' violated: (%s) P holds but Q never becomes true", prop.Name, prop.Description)
					}
					tv.reportViolationImmediately(prop, desc, violatingTrace)
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
		desc := fmt.Sprintf("Property '%s' violated: P holds but Q never becomes true", prop.Name)
		if prop.Description != "" {
			desc = fmt.Sprintf("Property '%s' violated: (%s) P holds but Q never becomes true", prop.Name, prop.Description)
		}
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  desc,
				Trace:        violatingTrace,
			},
		}, nil
	}
	
	return &TemporalVerificationResult{
		PropertyName: prop.Name,
		Holds:        true,
	}, nil
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
				// Report violation immediately
				desc := fmt.Sprintf("Property '%s' violated under %s(%s) fairness constraint", prop.Name, fairnessType, actionName)
				if prop.Description != "" {
					desc = fmt.Sprintf("Property '%s' violated: (%s) property does not hold under %s(%s) fairness constraint", prop.Name, prop.Description, fairnessType, actionName)
				}
				tv.reportViolationImmediately(prop, desc, violatingTrace)
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
		desc := fmt.Sprintf("Property '%s' violated under %s(%s) fairness constraint", prop.Name, fairnessType, actionName)
		if prop.Description != "" {
			desc = fmt.Sprintf("Property '%s' violated: (%s) property does not hold under %s(%s) fairness constraint", prop.Name, prop.Description, fairnessType, actionName)
		}
		return &TemporalVerificationResult{
			PropertyName: prop.Name,
			Holds:        false,
			Violation: &TemporalViolation{
				PropertyName: prop.Name,
				Description:  desc,
				Trace:        violatingTrace,
			},
		}, nil
	}
	
	return &TemporalVerificationResult{
		PropertyName: prop.Name,
		Holds:        true,
	}, nil
}

