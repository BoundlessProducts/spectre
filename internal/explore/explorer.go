package explore

import (
	"fmt"
	"sort"
	"strings"

	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/state"
)

// InvariantViolationCallback is called immediately when an invariant violation is detected
// Returns true if exploration should continue, false if it should stop
type InvariantViolationCallback func(violation *Violation) bool

// TemporalVerificationCallback is called periodically during exploration to check temporal properties incrementally
// It receives the current transition graph and initial states, and should return true if exploration should continue
type TemporalVerificationCallback func(graph *TransitionGraph, initialStates []*state.State) bool

// Explorer explores the state space of a Spectre specification
type Explorer struct {
	stateMachine *exec.StateMachine
	hasher       *StateHasher
	cycleDetector *CycleDetector
	visited      map[string]bool // State hash -> visited
	queue        []*ExplorationStateNode    // Queue for BFS
	stack        []*ExplorationStateNode    // Stack for DFS
	maxDepth     int             // Maximum exploration depth
	maxStates    int             // Maximum number of states to explore
	detectCycles bool            // Whether to detect cycles
	verbose      bool            // Whether to print verbose exploration details
	invariantViolationCallback InvariantViolationCallback // Optional callback for immediate violation reporting
	temporalVerificationCallback TemporalVerificationCallback // Optional callback for incremental temporal verification
	shouldStopExploration bool // Flag to stop exploration if user wants to stop after violation
	checkTemporalEvery int // Check temporal properties every N states (0 = never, 1 = every state)
}

// ExplorationStateNode represents a state in the exploration tree (during BFS/DFS)
type ExplorationStateNode struct {
	State       *state.State
	Depth       int
	Parent      *ExplorationStateNode
	Action      string // Action that led to this state
	ActionArgs  []state.Value
	Path        []*Transition // Path from initial state to this state
}

// Transition represents a state transition
type Transition struct {
	FromState *state.State
	ToState   *state.State
	Action    string
	Args      []state.Value
}

// ExplorationResult contains the results of state space exploration
type ExplorationResult struct {
	StatesExplored  int
	StatesVisited   int
	Violations      []*Violation
	ReachableStates []*state.State
	InitialStates   []*state.State // Initial states from which exploration started
	MaxDepth        int
	Cycles          []*Cycle
	TransitionGraph *TransitionGraph // Complete transition graph including cycles
}

// TransitionGraph represents the complete state transition graph
type TransitionGraph struct {
	States      map[string]*StateNode      // State hash -> StateNode
	Transitions []*Transition              // All transitions (including cycles)
	Outgoing    map[string][]*Transition   // State hash -> outgoing transitions
	Incoming    map[string][]*Transition   // State hash -> incoming transitions
}

// StateNode represents a node in the transition graph
type StateNode struct {
	State     *state.State
	Hash      string
	Outgoing  []*Transition
	Incoming  []*Transition
	InCycle   bool              // Whether this state is part of a cycle
	Cycles    []*CycleInfo      // Cycles this state participates in
}

// CycleInfo represents detailed information about a cycle
type CycleInfo struct {
	States      []*state.State // States in the cycle (in order)
	Transitions []*Transition  // Transitions forming the cycle
	StartState  *state.State   // Starting state of the cycle
}

// Cycle represents a cycle found in the state space
type Cycle struct {
	Path        []*Transition
	Description string
}

// Violation represents an invariant violation found during exploration
type Violation struct {
	State       *state.State
	Invariant   string
	Path        []*Transition
	Description string
}

// SetInvariantViolationCallback sets a callback that will be called immediately when an invariant violation is detected
// The callback should return true if exploration should continue, false if it should stop
func (e *Explorer) SetInvariantViolationCallback(callback InvariantViolationCallback) {
	e.invariantViolationCallback = callback
}

// SetTemporalVerificationCallback sets a callback that will be called periodically during exploration
// to check temporal properties incrementally. The checkInterval parameter specifies how often to check
// (e.g., 1 = every state, 5 = every 5 states, 0 = disabled)
func (e *Explorer) SetTemporalVerificationCallback(callback TemporalVerificationCallback, checkInterval int) {
	e.temporalVerificationCallback = callback
	e.checkTemporalEvery = checkInterval
}

// NewExplorer creates a new state space explorer
func NewExplorer(stateMachine *exec.StateMachine) *Explorer {
	hasher := NewStateHasher()
	return &Explorer{
		stateMachine: stateMachine,
		hasher:       hasher,
		cycleDetector: NewCycleDetector(hasher),
		visited:      make(map[string]bool),
		queue:        []*ExplorationStateNode{},
		stack:        []*ExplorationStateNode{},
		shouldStopExploration: false,
		maxDepth:     100,  // Default max depth
		maxStates:    1000, // Default max states
		detectCycles: true, // Enable cycle detection by default
	}
}

// SetMaxDepth sets the maximum exploration depth
func (e *Explorer) SetMaxDepth(depth int) {
	e.maxDepth = depth
}

// SetMaxStates sets the maximum number of states to explore
func (e *Explorer) SetMaxStates(max int) {
	e.maxStates = max
}

// SetVerbose sets verbose logging mode
func (e *Explorer) SetVerbose(verbose bool) {
	e.verbose = verbose
}

// ExploreBFS explores the state space using Breadth-First Search
func (e *Explorer) ExploreBFS() (*ExplorationResult, error) {
	e.visited = make(map[string]bool)
	e.queue = []*ExplorationStateNode{}

	// Get initial states
	initialStates, err := e.stateMachine.GetInitialStates()
	if err != nil {
		return nil, fmt.Errorf("error getting initial states: %w", err)
	}

	result := &ExplorationResult{
		Violations:      []*Violation{},
		ReachableStates: []*state.State{},
		InitialStates:   initialStates, // Store initial states for temporal verification
		Cycles:          []*Cycle{},
		TransitionGraph: NewTransitionGraph(),
	}

	// Add initial states to queue - sort by hash for deterministic ordering (like TLC)
	// First, collect initial states with their hashes
	type initStateWithHash struct {
		state *state.State
		hash  string
	}
	initStatesWithHashes := make([]initStateWithHash, 0, len(initialStates))
	for _, initState := range initialStates {
		hash := e.hasher.HashState(initState)
		if !e.visited[hash] {
			initStatesWithHashes = append(initStatesWithHashes, initStateWithHash{
				state: initState,
				hash:  hash,
			})
		}
	}
	
	// Sort initial states by hash for deterministic ordering
	sort.Slice(initStatesWithHashes, func(i, j int) bool {
		return initStatesWithHashes[i].hash < initStatesWithHashes[j].hash
	})
	
	// Add sorted initial states to queue
	for _, item := range initStatesWithHashes {
		e.visited[item.hash] = true
		result.ReachableStates = append(result.ReachableStates, item.state)
		
		// Add initial state to transition graph
		result.TransitionGraph.AddState(item.state, item.hash)
		
		// Update cycle detector with initial state
		if e.detectCycles {
			e.cycleDetector.PushState(item.state)
		}
		
		e.queue = append(e.queue, &ExplorationStateNode{
			State: item.state,
			Depth: 0,
			Path:  []*Transition{},
		})
	}

	// BFS exploration
	// -1 means unlimited
	for len(e.queue) > 0 && (e.maxStates == -1 || result.StatesExplored < e.maxStates) && !e.shouldStopExploration {
		current := e.queue[0]
		e.queue = e.queue[1:]

		if e.maxDepth != -1 && current.Depth >= e.maxDepth {
			if e.verbose {
				fmt.Printf("[DEPTH LIMIT] Skipping state at depth %d (max: %d)\n", current.Depth, e.maxDepth)
			}
			continue
		}

		result.StatesExplored++

		// Print current state info
		if e.verbose {
			fmt.Printf("\n[STATE %d] Exploring state (depth: %d):\n", result.StatesExplored, current.Depth)
			for name, value := range current.State.Variables {
				fmt.Printf("  %s = %s\n", name, value.String())
			}
		} else {
			// Non-verbose mode: show simple progress
			fmt.Printf("Exploring State %d\n", result.StatesExplored)
		}

		// Check temporal properties incrementally if callback is set
		if e.temporalVerificationCallback != nil && e.checkTemporalEvery > 0 && result.StatesExplored % e.checkTemporalEvery == 0 {
			// Use initial states from result (set during initialization)
			if !e.temporalVerificationCallback(result.TransitionGraph, result.InitialStates) {
				e.shouldStopExploration = true
				break
			}
		}

		// Validate current state
		errors, err := e.stateMachine.ValidateState(current.State)
		if err != nil {
			return nil, fmt.Errorf("error validating state: %w", err)
		}

		if len(errors) > 0 {
			// Found invariant violation
			if e.verbose {
				fmt.Printf("  [VIOLATION] Found %d invariant violation(s)\n", len(errors))
			}
			for _, validationError := range errors {
				result.Violations = append(result.Violations, &Violation{
					State:       current.State,
					Invariant:   validationError.Name,
					Path:        current.Path,
					Description: validationError.Message,
				})
			}
		}

		// Get available actions with their argument combinations
		availableActions, err := e.stateMachine.GetAvailableActionsWithArgs(current.State)
		if err != nil {
			return nil, fmt.Errorf("error getting available actions: %w", err)
		}

		if e.verbose {
			actionNames := make([]string, len(availableActions))
			for i, a := range availableActions {
				if len(a.Args) > 0 {
					argsStr := make([]string, len(a.Args))
					for j, arg := range a.Args {
						argsStr[j] = arg.String()
					}
					actionNames[i] = fmt.Sprintf("%s(%s)", a.ActionName, strings.Join(argsStr, ", "))
				} else {
					actionNames[i] = a.ActionName
				}
			}
			fmt.Printf("  [ACTIONS] Available actions: %v\n", actionNames)
		}

		// Explore each action with its arguments
		for _, actionWithArgs := range availableActions {
			if e.verbose {
				if len(actionWithArgs.Args) > 0 {
					argsStr := make([]string, len(actionWithArgs.Args))
					for i, arg := range actionWithArgs.Args {
						argsStr[i] = arg.String()
					}
					fmt.Printf("    [TRY] Executing action: %s(%s)\n", actionWithArgs.ActionName, strings.Join(argsStr, ", "))
				} else {
					fmt.Printf("    [TRY] Executing action: %s\n", actionWithArgs.ActionName)
				}
			}
			// Check if we should stop before executing action
			if e.shouldStopExploration {
				break
			}
			
			nextState, err := e.stateMachine.ExecuteAction(actionWithArgs.ActionName, current.State, actionWithArgs.Args)
			if err != nil {
				// Action execution failed - check if it's due to invariant violations
				errMsg := err.Error()
				if strings.Contains(errMsg, "violates invariants") {
					// Extract invariant violation information from the error
					violationDesc := errMsg
					// Try to extract just the violation part
					if idx := strings.Index(errMsg, ": "); idx > 0 {
						violationDesc = errMsg[idx+2:]
					}
					
					violation := &Violation{
						State:       current.State,
						Invariant:   "unknown", // We'll extract this from the error message if possible
						Path:        current.Path,
						Description: fmt.Sprintf("Action '%s' would violate invariants: %s", actionWithArgs.ActionName, violationDesc),
					}
					result.Violations = append(result.Violations, violation)
					
					// Call callback if set, and stop exploration if callback returns false
					if e.invariantViolationCallback != nil {
						if !e.invariantViolationCallback(violation) {
							e.shouldStopExploration = true
							break
						}
					}
					
					if e.verbose {
						fmt.Printf("      [FAIL] Action %s failed: %s\n", actionWithArgs.ActionName, errMsg)
					}
				} else if e.verbose {
					fmt.Printf("      [FAIL] Action %s failed: %v\n", actionWithArgs.ActionName, err)
				}
				// This could be due to preconditions not being met (shouldn't happen since
				// GetAvailableActions filters them), postcondition violations, or invariant violations
				// Continue with other actions
				continue
			}

			hash := e.hasher.HashState(nextState)
			
			// Create transition (even if already visited, to build complete graph)
			transition := &Transition{
				FromState: current.State,
				ToState:   nextState,
				Action:    actionWithArgs.ActionName,
				Args:      actionWithArgs.Args,
			}
			
			currentHash := e.hasher.HashState(current.State)
			
			// Add to transition graph (even for cycles - we want complete graph)
			result.TransitionGraph.AddTransition(transition, currentHash, hash)
			
			// Only skip exploring if we've already visited this state
			if e.visited[hash] {
				if e.verbose {
					fmt.Printf("      [SKIP] Next state already visited (hash: %s)\n", hash)
					if e.detectCycles && e.cycleDetector.HasCycle(nextState) {
						fmt.Printf("        [CYCLE] This would create a cycle (but transition recorded in graph)\n")
					}
				}
				// If cycle detection is enabled and this creates a cycle in the current path, record it
				if e.detectCycles && e.cycleDetector.HasCycle(nextState) {
					cyclePath := e.cycleDetector.GetCyclePath(nextState)
					if len(cyclePath) > 0 {
						// Build transition path for the cycle
						transitionPath := make([]*Transition, 0)
						for i := 0; i < len(current.Path); i++ {
							transitionPath = append(transitionPath, current.Path[i])
						}
						transitionPath = append(transitionPath, transition)
						
						result.Cycles = append(result.Cycles, &Cycle{
							Path:        transitionPath,
							Description: fmt.Sprintf("Cycle detected: %d states", len(cyclePath)),
						})
					}
				}
				continue // Already visited, skip exploring further
			}
			
			// New state - mark as visited and explore it
			e.visited[hash] = true
			result.ReachableStates = append(result.ReachableStates, nextState)
			
			if e.verbose {
				if len(actionWithArgs.Args) > 0 {
					argsStr := make([]string, len(actionWithArgs.Args))
					for i, arg := range actionWithArgs.Args {
						argsStr[i] = arg.String()
					}
					fmt.Printf("      [SUCCESS] Action %s(%s) succeeded, next state:\n", actionWithArgs.ActionName, strings.Join(argsStr, ", "))
				} else {
					fmt.Printf("      [SUCCESS] Action %s succeeded, next state:\n", actionWithArgs.ActionName)
				}
				for name, value := range nextState.Variables {
					fmt.Printf("        %s = %s\n", name, value.String())
				}
				fmt.Printf("        Adding to queue for exploration\n")
			}

			// Create new path
			newPath := make([]*Transition, len(current.Path))
			copy(newPath, current.Path)
			newPath = append(newPath, transition)

			// Update cycle detector path
			if e.detectCycles {
				e.cycleDetector.PushState(nextState)
			}

			// Add to queue in sorted position by state hash for deterministic ordering (like TLC)
			// This ensures states are always explored in the same order regardless of discovery order
			// Queue is maintained sorted: first by depth (BFS level), then by state hash within same depth
			newNode := &ExplorationStateNode{
				State:      nextState,
				Depth:      current.Depth + 1,
				Parent:     current,
				Action:     actionWithArgs.ActionName,
				ActionArgs: actionWithArgs.Args,
				Path:       newPath,
			}
			
			// Insert in sorted position: maintain queue sorted by depth, then by state hash
			// This ensures deterministic exploration order like TLC
			insertPos := len(e.queue)
			nextStateHash := hash
			nextDepth := current.Depth + 1
			
			for i, node := range e.queue {
				nodeHash := e.hasher.HashState(node.State)
				// Compare first by depth, then by hash within same depth
				if node.Depth > nextDepth {
					// Insert before nodes at deeper depth
					insertPos = i
					break
				} else if node.Depth == nextDepth {
					// Same depth: compare by hash
					if nextStateHash < nodeHash {
						insertPos = i
						break
					}
				}
				// If node.Depth < nextDepth, continue (insert after shallower states)
			}
			
			// Insert at the determined position
			e.queue = append(e.queue, nil) // Extend slice
			copy(e.queue[insertPos+1:], e.queue[insertPos:])
			e.queue[insertPos] = newNode

			if current.Depth+1 > result.MaxDepth {
				result.MaxDepth = current.Depth + 1
			}
		}
	}

	result.StatesVisited = len(e.visited)
	return result, nil
}

// ExploreDFS explores the state space using Depth-First Search
func (e *Explorer) ExploreDFS() (*ExplorationResult, error) {
	e.visited = make(map[string]bool)
	e.stack = []*ExplorationStateNode{}

	// Get initial states
	initialStates, err := e.stateMachine.GetInitialStates()
	if err != nil {
		return nil, fmt.Errorf("error getting initial states: %w", err)
	}

	result := &ExplorationResult{
		Violations:      []*Violation{},
		ReachableStates: []*state.State{},
		InitialStates:   initialStates, // Store initial states
		Cycles:          []*Cycle{},
		TransitionGraph: NewTransitionGraph(),
	}

	// Add initial states to stack
	for _, initState := range initialStates {
		hash := e.hasher.HashState(initState)
		if !e.visited[hash] {
			e.visited[hash] = true
			result.ReachableStates = append(result.ReachableStates, initState)
			
			// Add initial state to transition graph
			result.TransitionGraph.AddState(initState, hash)
			
			// Update cycle detector with initial state
			if e.detectCycles {
				e.cycleDetector.PushState(initState)
			}
			
			e.stack = append(e.stack, &ExplorationStateNode{
				State: initState,
				Depth: 0,
				Path:  []*Transition{},
			})
		}
	}

	// DFS exploration
	// -1 means unlimited
	for len(e.stack) > 0 && (e.maxStates == -1 || result.StatesExplored < e.maxStates) && !e.shouldStopExploration {
		// Pop from stack
		current := e.stack[len(e.stack)-1]
		e.stack = e.stack[:len(e.stack)-1]

		if e.maxDepth != -1 && current.Depth >= e.maxDepth {
			continue
		}

		result.StatesExplored++

		// Print current state info
		if e.verbose {
			fmt.Printf("\n[STATE %d] Exploring state (depth: %d):\n", result.StatesExplored, current.Depth)
			for name, value := range current.State.Variables {
				fmt.Printf("  %s = %s\n", name, value.String())
			}
		} else {
			// Non-verbose mode: show simple progress
			fmt.Printf("Exploring State %d\n", result.StatesExplored)
		}

		// Validate current state
		errors, err := e.stateMachine.ValidateState(current.State)
		if err != nil {
			return nil, fmt.Errorf("error validating state: %w", err)
		}

		if len(errors) > 0 {
			// Found invariant violation
			if e.verbose {
				fmt.Printf("  [VIOLATION] Found %d invariant violation(s)\n", len(errors))
			}
			for _, validationError := range errors {
				violation := &Violation{
					State:       current.State,
					Invariant:   validationError.Name,
					Path:        current.Path,
					Description: validationError.Message,
				}
				result.Violations = append(result.Violations, violation)
				
				// Call callback if set, and stop exploration if callback returns false
				if e.invariantViolationCallback != nil {
					if !e.invariantViolationCallback(violation) {
						e.shouldStopExploration = true
						break
					}
				}
			}
			
			// Check if we should stop after processing violations
			if e.shouldStopExploration {
				break
			}
		}

		// Get available actions with their argument combinations
		availableActions, err := e.stateMachine.GetAvailableActionsWithArgs(current.State)
		if err != nil {
			return nil, fmt.Errorf("error getting available actions: %w", err)
		}

		// Explore each action (push in reverse order for DFS)
		for i := len(availableActions) - 1; i >= 0; i-- {
			// Check if we should stop before executing action
			if e.shouldStopExploration {
				break
			}
			
			actionWithArgs := availableActions[i]
			nextState, err := e.stateMachine.ExecuteAction(actionWithArgs.ActionName, current.State, actionWithArgs.Args)
			if err != nil {
				// Action execution failed - check if it's due to invariant violations
				errMsg := err.Error()
				if strings.Contains(errMsg, "violates invariants") {
					// Extract invariant violation information from the error
					// The error format is: "next state violates invariants: [validation error messages]"
					// We need to parse this and create violation entries
					violationDesc := errMsg
					// Try to extract just the violation part
					if idx := strings.Index(errMsg, ": "); idx > 0 {
						violationDesc = errMsg[idx+2:]
					}
					
					violation := &Violation{
						State:       current.State,
						Invariant:   "unknown", // We'll extract this from the error message if possible
						Path:        current.Path,
						Description: fmt.Sprintf("Action '%s' would violate invariants: %s", actionWithArgs.ActionName, violationDesc),
					}
					result.Violations = append(result.Violations, violation)
					
					// Call callback if set, and stop exploration if callback returns false
					if e.invariantViolationCallback != nil {
						if !e.invariantViolationCallback(violation) {
							e.shouldStopExploration = true
							break
						}
					}
					
					if e.verbose {
						fmt.Printf("      [FAIL] Action %s failed: %s\n", actionWithArgs.ActionName, errMsg)
					}
				} else if e.verbose {
					fmt.Printf("      [FAIL] Action %s failed: %s\n", actionWithArgs.ActionName, errMsg)
				}
				continue
			}

		hash := e.hasher.HashState(nextState)
			
			// Create transition
			transition := &Transition{
				FromState: current.State,
				ToState:   nextState,
				Action:    actionWithArgs.ActionName,
				Args:      actionWithArgs.Args,
			}
			
			currentHash := e.hasher.HashState(current.State)
			result.TransitionGraph.AddTransition(transition, currentHash, hash)
			
			// Check for cycles if enabled
			if e.detectCycles {
				if e.cycleDetector.HasCycle(nextState) {
					// Found a cycle
					cyclePath := e.cycleDetector.GetCyclePath(nextState)
					if len(cyclePath) > 0 {
						transitionPath := make([]*Transition, 0)
						for i := 0; i < len(current.Path); i++ {
							transitionPath = append(transitionPath, current.Path[i])
						}
						transitionPath = append(transitionPath, transition)
						
						result.Cycles = append(result.Cycles, &Cycle{
							Path:        transitionPath,
							Description: fmt.Sprintf("Cycle detected: %d states", len(cyclePath)),
						})
					}
					continue
				}
			}
			
			if !e.visited[hash] {
				e.visited[hash] = true
				result.ReachableStates = append(result.ReachableStates, nextState)

				// Create transition
				transition := &Transition{
					FromState: current.State,
					ToState:   nextState,
					Action:    actionWithArgs.ActionName,
					Args:      actionWithArgs.Args,
				}

				// Create new path
				newPath := make([]*Transition, len(current.Path))
				copy(newPath, current.Path)
				newPath = append(newPath, transition)

				// Update cycle detector path
				if e.detectCycles {
					e.cycleDetector.PushState(nextState)
				}

				// Push to stack
				e.stack = append(e.stack, &ExplorationStateNode{
					State:      nextState,
					Depth:      current.Depth + 1,
					Parent:     current,
					Action:     actionWithArgs.ActionName,
					ActionArgs: actionWithArgs.Args,
					Path:       newPath,
				})

				if current.Depth+1 > result.MaxDepth {
					result.MaxDepth = current.Depth + 1
				}
			}
		}
	}

	result.StatesVisited = len(e.visited)
	return result, nil
}


