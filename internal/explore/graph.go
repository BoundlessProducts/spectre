package explore

import (
	"sort"

	"github.com/akkeshavan/spectre/internal/state"
)

// NewTransitionGraph creates a new empty transition graph
func NewTransitionGraph() *TransitionGraph {
	return &TransitionGraph{
		States:      make(map[string]*StateNode),
		Transitions: []*Transition{},
		Outgoing:    make(map[string][]*Transition),
		Incoming:    make(map[string][]*Transition),
	}
}

// AddState adds a state to the graph
func (tg *TransitionGraph) AddState(s *state.State, hash string) *StateNode {
	if node, exists := tg.States[hash]; exists {
		return node
	}
	
	node := &StateNode{
		State:    s,
		Hash:     hash,
		Outgoing: []*Transition{},
		Incoming: []*Transition{},
		InCycle:  false,
		Cycles:   []*CycleInfo{},
	}
	tg.States[hash] = node
	return node
}

// AddTransition adds a transition to the graph
func (tg *TransitionGraph) AddTransition(trans *Transition, fromHash, toHash string) {
	// Ensure both states are in the graph
	fromNode := tg.AddState(trans.FromState, fromHash)
	toNode := tg.AddState(trans.ToState, toHash)
	
	// Add transition
	tg.Transitions = append(tg.Transitions, trans)
	tg.Outgoing[fromHash] = append(tg.Outgoing[fromHash], trans)
	tg.Incoming[toHash] = append(tg.Incoming[toHash], trans)
	
	// Update node connections
	fromNode.Outgoing = append(fromNode.Outgoing, trans)
	toNode.Incoming = append(toNode.Incoming, trans)
}

// DetectCycles finds all cycles in the transition graph
func (tg *TransitionGraph) DetectCycles(hasher *StateHasher) []*CycleInfo {
	var cycles []*CycleInfo
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []*StateNode{}
	
	// DFS to find cycles
	var dfs func(node *StateNode)
	dfs = func(node *StateNode) {
		if recStack[node.Hash] {
			// Found a cycle - build cycle info
			cycleStart := -1
			for i, n := range path {
				if n.Hash == node.Hash {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycleStates := make([]*state.State, 0)
				cycleTransitions := make([]*Transition, 0)
				for i := cycleStart; i < len(path); i++ {
					cycleStates = append(cycleStates, path[i].State)
					// Find transition from path[i] to path[i+1]
					if i < len(path)-1 {
						for _, trans := range path[i].Outgoing {
							if hasher.HashState(trans.ToState) == path[i+1].Hash {
								cycleTransitions = append(cycleTransitions, trans)
								break
							}
						}
					}
				}
				// Add transition from last to first (or self-loop transition)
				lastNode := path[len(path)-1]
				// Check if this is a self-loop (path has only one node)
				if len(path) == cycleStart+1 && len(path) == 1 {
					// Self-loop: find the self-loop transition
					for _, trans := range lastNode.Outgoing {
						if hasher.HashState(trans.ToState) == node.Hash {
							// Make sure we haven't already added it
							alreadyAdded := false
							for _, t := range cycleTransitions {
								if t == trans {
									alreadyAdded = true
									break
								}
							}
							if !alreadyAdded {
								cycleTransitions = append(cycleTransitions, trans)
							}
							break
						}
					}
				} else if len(path) > cycleStart+1 {
					// Multi-state cycle - add transition from last to first
					for _, trans := range lastNode.Outgoing {
						if hasher.HashState(trans.ToState) == node.Hash {
							cycleTransitions = append(cycleTransitions, trans)
							break
						}
					}
				}
				
				cycleInfo := &CycleInfo{
					States:      cycleStates,
					Transitions: cycleTransitions,
					StartState:  cycleStates[0],
				}
				cycles = append(cycles, cycleInfo)
				
				// Mark all states in cycle
				for _, n := range path[cycleStart:] {
					n.InCycle = true
					n.Cycles = append(n.Cycles, cycleInfo)
				}
			}
			return
		}
		
		if visited[node.Hash] {
			return
		}
		
		visited[node.Hash] = true
		recStack[node.Hash] = true
		path = append(path, node)
		
		// Visit all neighbors
		for _, trans := range node.Outgoing {
			toHash := hasher.HashState(trans.ToState)
			if toNode, exists := tg.States[toHash]; exists {
				dfs(toNode)
			}
		}
		
		// Backtrack
		recStack[node.Hash] = false
		path = path[:len(path)-1]
	}
	
	// Start DFS from each unvisited state
	// Sort state hashes for deterministic iteration order
	stateHashes := make([]string, 0, len(tg.States))
	for hash := range tg.States {
		stateHashes = append(stateHashes, hash)
	}
	sort.Strings(stateHashes)
	
	for _, hash := range stateHashes {
		node := tg.States[hash]
		if !visited[node.Hash] {
			path = []*StateNode{}
			dfs(node)
		}
	}
	
	return cycles
}

// GetStateNode returns the node for a given state hash
func (tg *TransitionGraph) GetStateNode(hash string) *StateNode {
	return tg.States[hash]
}

// GetAllStates returns all states in the graph
func (tg *TransitionGraph) GetAllStates() []*state.State {
	states := make([]*state.State, 0, len(tg.States))
	// Sort state hashes for deterministic iteration order
	stateHashes := make([]string, 0, len(tg.States))
	for hash := range tg.States {
		stateHashes = append(stateHashes, hash)
	}
	sort.Strings(stateHashes)
	
	for _, hash := range stateHashes {
		states = append(states, tg.States[hash].State)
	}
	return states
}

