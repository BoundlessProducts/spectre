package explore

import (
	"fmt"

	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/state"
)

// Explorer explores the state space of a Spectre specification
type Explorer struct {
	stateMachine *exec.StateMachine
	hasher       *StateHasher
	cycleDetector *CycleDetector
	visited      map[string]bool // State hash -> visited
	queue        []*StateNode    // Queue for BFS
	stack        []*StateNode    // Stack for DFS
	maxDepth     int             // Maximum exploration depth
	maxStates    int             // Maximum number of states to explore
	detectCycles bool            // Whether to detect cycles
}

// StateNode represents a state in the exploration tree
type StateNode struct {
	State       *state.State
	Depth       int
	Parent      *StateNode
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
	StatesExplored int
	StatesVisited  int
	Violations     []*Violation
	ReachableStates []*state.State
	MaxDepth       int
	Cycles         []*Cycle
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

// NewExplorer creates a new state space explorer
func NewExplorer(stateMachine *exec.StateMachine) *Explorer {
	hasher := NewStateHasher()
	return &Explorer{
		stateMachine: stateMachine,
		hasher:       hasher,
		cycleDetector: NewCycleDetector(hasher),
		visited:      make(map[string]bool),
		queue:        []*StateNode{},
		stack:        []*StateNode{},
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

// ExploreBFS explores the state space using Breadth-First Search
func (e *Explorer) ExploreBFS() (*ExplorationResult, error) {
	e.visited = make(map[string]bool)
	e.queue = []*StateNode{}

	// Get initial states
	initialStates, err := e.stateMachine.GetInitialStates()
	if err != nil {
		return nil, fmt.Errorf("error getting initial states: %w", err)
	}

	result := &ExplorationResult{
		Violations:      []*Violation{},
		ReachableStates: []*state.State{},
		Cycles:          []*Cycle{},
	}

	// Add initial states to queue
	for _, initState := range initialStates {
		hash := e.hasher.HashState(initState)
		if !e.visited[hash] {
			e.visited[hash] = true
			result.ReachableStates = append(result.ReachableStates, initState)
			e.queue = append(e.queue, &StateNode{
				State: initState,
				Depth: 0,
				Path:  []*Transition{},
			})
		}
	}

	// BFS exploration
	for len(e.queue) > 0 && result.StatesExplored < e.maxStates {
		current := e.queue[0]
		e.queue = e.queue[1:]

		if current.Depth >= e.maxDepth {
			continue
		}

		result.StatesExplored++

		// Validate current state
		errors, err := e.stateMachine.ValidateState(current.State)
		if err != nil {
			return nil, fmt.Errorf("error validating state: %w", err)
		}

		if len(errors) > 0 {
			// Found invariant violation
			for _, validationError := range errors {
				result.Violations = append(result.Violations, &Violation{
					State:       current.State,
					Invariant:   validationError.Name,
					Path:        current.Path,
					Description: validationError.Message,
				})
			}
		}

		// Get available actions
		availableActions, err := e.stateMachine.GetAvailableActions(current.State)
		if err != nil {
			return nil, fmt.Errorf("error getting available actions: %w", err)
		}

		// Explore each action
		for _, actionName := range availableActions {
			nextState, err := e.stateMachine.ExecuteAction(actionName, current.State, nil)
			if err != nil {
				// Action execution failed (e.g., postcondition violation)
				// Continue with other actions
				continue
			}

			hash := e.hasher.HashState(nextState)
			
			// Check for cycles if enabled
			if e.detectCycles {
				if e.cycleDetector.HasCycle(nextState) {
					// Found a cycle
					cyclePath := e.cycleDetector.GetCyclePath(nextState)
					if len(cyclePath) > 0 {
						// Build transition path for the cycle
						transitionPath := make([]*Transition, 0)
						for i := 0; i < len(current.Path); i++ {
							transitionPath = append(transitionPath, current.Path[i])
						}
						transitionPath = append(transitionPath, &Transition{
							FromState: current.State,
							ToState:   nextState,
							Action:    actionName,
							Args:      nil,
						})
						
						result.Cycles = append(result.Cycles, &Cycle{
							Path:        transitionPath,
							Description: fmt.Sprintf("Cycle detected: %d states", len(cyclePath)),
						})
					}
					continue // Don't explore cycles
				}
			}
			
			if !e.visited[hash] {
				e.visited[hash] = true
				result.ReachableStates = append(result.ReachableStates, nextState)

				// Create transition
				transition := &Transition{
					FromState: current.State,
					ToState:   nextState,
					Action:    actionName,
					Args:      nil,
				}

				// Create new path
				newPath := make([]*Transition, len(current.Path))
				copy(newPath, current.Path)
				newPath = append(newPath, transition)

				// Update cycle detector path
				if e.detectCycles {
					e.cycleDetector.PushState(nextState)
				}

				// Add to queue
				e.queue = append(e.queue, &StateNode{
					State:      nextState,
					Depth:      current.Depth + 1,
					Parent:     current,
					Action:     actionName,
					ActionArgs: nil,
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

// ExploreDFS explores the state space using Depth-First Search
func (e *Explorer) ExploreDFS() (*ExplorationResult, error) {
	e.visited = make(map[string]bool)
	e.stack = []*StateNode{}

	// Get initial states
	initialStates, err := e.stateMachine.GetInitialStates()
	if err != nil {
		return nil, fmt.Errorf("error getting initial states: %w", err)
	}

	result := &ExplorationResult{
		Violations:      []*Violation{},
		ReachableStates: []*state.State{},
		Cycles:          []*Cycle{},
	}

	// Add initial states to stack
	for _, initState := range initialStates {
		hash := e.hasher.HashState(initState)
		if !e.visited[hash] {
			e.visited[hash] = true
			result.ReachableStates = append(result.ReachableStates, initState)
			e.stack = append(e.stack, &StateNode{
				State: initState,
				Depth: 0,
				Path:  []*Transition{},
			})
		}
	}

	// DFS exploration
	for len(e.stack) > 0 && result.StatesExplored < e.maxStates {
		// Pop from stack
		current := e.stack[len(e.stack)-1]
		e.stack = e.stack[:len(e.stack)-1]

		if current.Depth >= e.maxDepth {
			continue
		}

		result.StatesExplored++

		// Validate current state
		errors, err := e.stateMachine.ValidateState(current.State)
		if err != nil {
			return nil, fmt.Errorf("error validating state: %w", err)
		}

		if len(errors) > 0 {
			// Found invariant violation
			for _, validationError := range errors {
				result.Violations = append(result.Violations, &Violation{
					State:       current.State,
					Invariant:   validationError.Name,
					Path:        current.Path,
					Description: validationError.Message,
				})
			}
		}

		// Get available actions
		availableActions, err := e.stateMachine.GetAvailableActions(current.State)
		if err != nil {
			return nil, fmt.Errorf("error getting available actions: %w", err)
		}

		// Explore each action (push in reverse order for DFS)
		for i := len(availableActions) - 1; i >= 0; i-- {
			actionName := availableActions[i]
			nextState, err := e.stateMachine.ExecuteAction(actionName, current.State, nil)
			if err != nil {
				// Action execution failed
				continue
			}

			hash := e.hasher.HashState(nextState)
			
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
						transitionPath = append(transitionPath, &Transition{
							FromState: current.State,
							ToState:   nextState,
							Action:    actionName,
							Args:      nil,
						})
						
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
					Action:    actionName,
					Args:      nil,
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
				e.stack = append(e.stack, &StateNode{
					State:      nextState,
					Depth:      current.Depth + 1,
					Parent:     current,
					Action:     actionName,
					ActionArgs: nil,
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


