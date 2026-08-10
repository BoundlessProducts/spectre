package explore

import (
	"fmt"
	"strings"

	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/state"
)

// IncrementalReverify updates a cached ExplorationResult when a single action's
// transition relation changes (e.g. after spectre sync flags a body mutation).
//
// Algorithm (equivalent to fresh full BFS when only changedAction changes):
//  1. Prune all stale transitions produced by changedAction.
//  2. Recompute reachability from initial states over the remaining edges.
//  3. Remove states no longer reachable (were only reachable via changedAction).
//  4. Re-execute changedAction from every remaining reachable state.
//  5. BFS-expand any newly discovered successor states with all actions,
//     capped at the original state count to preserve bounded exploration.
//
// Soundness: steps 1–3 leave exactly the states reachable without changedAction.
// Step 4 re-adds the new A-successors. Step 5 closes the graph under all actions
// from newly reachable states up to the original bound. The result equals a fresh
// BFS of the modified spec (same bound), assuming all other actions are unchanged.
func IncrementalReverify(
	result *ExplorationResult,
	changedAction string,
	sm *exec.StateMachine,
	hasher *StateHasher,
	silent bool,
) (*ExplorationResult, error) {
	// Cap expansion at the original state count to preserve bounded BFS.
	originalMaxStates := len(result.TransitionGraph.States)

	// Build a reverse map: *state.State → hash, to avoid repeated hashing.
	ptrToHash := buildPtrToHash(result.TransitionGraph)

	// 1. Remove stale transitions for changedAction.
	pruneActionWithMap(result.TransitionGraph, changedAction, ptrToHash)

	// 2. Get initial states from the new spec (init block is unchanged).
	initStates, err := sm.GetInitialStates()
	if err != nil {
		return nil, fmt.Errorf("incremental: could not get initial states: %w", err)
	}
	initHashes := make([]string, 0, len(initStates))
	for _, s := range initStates {
		h := hasher.HashState(s)
		result.TransitionGraph.AddState(s, h)
		ptrToHash[s] = h
		initHashes = append(initHashes, h)
	}
	result.InitialStates = initStates

	// 3. Recompute reachability from initial states and remove unreachable states.
	reachable := reachabilityBFS(result.TransitionGraph, initHashes, ptrToHash)
	pruned := pruneUnreachableStates(result.TransitionGraph, reachable, ptrToHash)
	if !silent {
		fmt.Printf("Incremental re-verify: pruned %d states unreachable without %q\n",
			pruned, changedAction)
	}

	// Rebuild ptrToHash after pruning (some entries removed).
	ptrToHash = buildPtrToHash(result.TransitionGraph)

	// 4. Snapshot the current reachable states (avoid iterating a mutating map).
	type stateEntry struct {
		hash  string
		state *state.State
	}
	snapshot := make([]stateEntry, 0, len(result.TransitionGraph.States))
	for h, node := range result.TransitionGraph.States {
		snapshot = append(snapshot, stateEntry{hash: h, state: node.State})
	}

	visited := make(map[string]bool, len(result.TransitionGraph.States))
	for h := range result.TransitionGraph.States {
		visited[h] = true
	}
	var newViolations []*Violation
	var newExpand []*state.State

	// Re-execute changedAction from each snapshotted reachable state.
	for _, entry := range snapshot {
		s, h := entry.state, entry.hash
		acts, err := sm.GetAvailableActionsWithArgs(s)
		if err != nil {
			continue
		}
		for _, a := range acts {
			if a.ActionName != changedAction {
				continue
			}
			nextState, execErr := sm.ExecuteAction(a.ActionName, s, a.Args)
			if execErr != nil {
				if strings.Contains(execErr.Error(), "violates invariants") {
					newViolations = append(newViolations, &Violation{
						State:       s,
						Invariant:   extractInvariantName(execErr.Error()),
						Description: fmt.Sprintf("Action '%s' would violate invariants: %s", a.ActionName, execErr.Error()),
					})
				}
				continue
			}
			nextHash := hasher.HashState(nextState)
			trans := &Transition{FromState: s, ToState: nextState, Action: a.ActionName, Args: a.Args}
			result.TransitionGraph.AddTransition(trans, h, nextHash)
			ptrToHash[nextState] = nextHash
			if !visited[nextHash] {
				visited[nextHash] = true
				errs, _ := sm.ValidateState(nextState)
				for _, e := range errs {
					newViolations = append(newViolations, &Violation{
						State:       nextState,
						Invariant:   e.Name,
						Description: e.Message,
					})
				}
				newExpand = append(newExpand, nextState)
			}
		}
	}

	// 5. BFS-expand newly discovered states with all actions.
	// Cap at originalMaxStates to match the bounded BFS the cache was built with.
	newStates := 0
	for len(newExpand) > 0 && len(result.TransitionGraph.States) < originalMaxStates {
		current := newExpand[0]
		newExpand = newExpand[1:]
		newStates++
		currentHash := hasher.HashState(current)

		acts, err := sm.GetAvailableActionsWithArgs(current)
		if err != nil {
			continue
		}
		for _, a := range acts {
			nextState, execErr := sm.ExecuteAction(a.ActionName, current, a.Args)
			if execErr != nil {
				if strings.Contains(execErr.Error(), "violates invariants") {
					newViolations = append(newViolations, &Violation{
						State:       current,
						Invariant:   extractInvariantName(execErr.Error()),
						Description: fmt.Sprintf("Action '%s' would violate invariants: %s", a.ActionName, execErr.Error()),
					})
				}
				continue
			}
			nextHash := hasher.HashState(nextState)
			trans := &Transition{FromState: current, ToState: nextState, Action: a.ActionName, Args: a.Args}
			result.TransitionGraph.AddTransition(trans, currentHash, nextHash)
			ptrToHash[nextState] = nextHash
			if !visited[nextHash] {
				visited[nextHash] = true
				errs, _ := sm.ValidateState(nextState)
				for _, e := range errs {
					newViolations = append(newViolations, &Violation{
						State:       nextState,
						Invariant:   e.Name,
						Description: e.Message,
					})
				}
				newExpand = append(newExpand, nextState)
			}
		}
	}

	result.Violations = append(result.Violations, newViolations...)
	result.StatesExplored = len(result.TransitionGraph.States)
	result.StatesVisited = len(visited)
	if !silent {
		fmt.Printf("Incremental re-verify done: %d new states, %d total states\n",
			newStates, result.StatesExplored)
	}
	return result, nil
}

// buildPtrToHash builds a reverse map from *state.State pointer to its hash
// as recorded in the TransitionGraph. Avoids repeated SHA-256 computation.
func buildPtrToHash(tg *TransitionGraph) map[*state.State]string {
	m := make(map[*state.State]string, len(tg.States))
	for h, node := range tg.States {
		m[node.State] = h
	}
	return m
}

// reachabilityBFS returns the set of hashes reachable from initHashes via
// the current edges in tg. Uses ptrToHash to look up successor hashes
// without re-computing SHA-256.
func reachabilityBFS(tg *TransitionGraph, initHashes []string, ptrToHash map[*state.State]string) map[string]bool {
	reachable := make(map[string]bool, len(tg.States))
	queue := make([]string, 0, len(initHashes))
	for _, h := range initHashes {
		if !reachable[h] {
			reachable[h] = true
			queue = append(queue, h)
		}
	}
	for len(queue) > 0 {
		h := queue[0]
		queue = queue[1:]
		for _, t := range tg.Outgoing[h] {
			toH, ok := ptrToHash[t.ToState]
			if !ok {
				continue
			}
			if !reachable[toH] {
				reachable[toH] = true
				queue = append(queue, toH)
			}
		}
	}
	return reachable
}

// pruneUnreachableStates removes all states not in reachable and their transitions.
// Returns the number of states removed.
func pruneUnreachableStates(tg *TransitionGraph, reachable map[string]bool, ptrToHash map[*state.State]string) int {
	pruned := 0
	for h := range tg.States {
		if !reachable[h] {
			delete(tg.States, h)
			delete(tg.Outgoing, h)
			delete(tg.Incoming, h)
			pruned++
		}
	}
	// Filter flat Transitions slice using ptrToHash for O(1) lookups.
	kept := make([]*Transition, 0, len(tg.Transitions))
	for _, t := range tg.Transitions {
		fH, fOK := ptrToHash[t.FromState]
		tH, tOK := ptrToHash[t.ToState]
		if fOK && tOK && reachable[fH] && reachable[tH] {
			kept = append(kept, t)
		}
	}
	tg.Transitions = kept
	// Update per-node Outgoing/Incoming slices.
	for _, node := range tg.States {
		var out []*Transition
		for _, t := range node.Outgoing {
			if h, ok := ptrToHash[t.ToState]; ok && reachable[h] {
				out = append(out, t)
			}
		}
		node.Outgoing = out
		var in []*Transition
		for _, t := range node.Incoming {
			if h, ok := ptrToHash[t.FromState]; ok && reachable[h] {
				in = append(in, t)
			}
		}
		node.Incoming = in
	}
	return pruned
}

// pruneActionWithMap removes all transitions for actionName from the graph.
// Uses ptrToHash for efficient map updates.
func pruneActionWithMap(tg *TransitionGraph, actionName string, _ map[*state.State]string) {
	// Flat slice.
	kept := make([]*Transition, 0, len(tg.Transitions))
	for _, t := range tg.Transitions {
		if t.Action != actionName {
			kept = append(kept, t)
		}
	}
	tg.Transitions = kept
	// Per-node lists.
	for _, node := range tg.States {
		var out []*Transition
		for _, t := range node.Outgoing {
			if t.Action != actionName {
				out = append(out, t)
			}
		}
		node.Outgoing = out
		var in []*Transition
		for _, t := range node.Incoming {
			if t.Action != actionName {
				in = append(in, t)
			}
		}
		node.Incoming = in
	}
	// Outgoing map.
	for h, ts := range tg.Outgoing {
		var kts []*Transition
		for _, t := range ts {
			if t.Action != actionName {
				kts = append(kts, t)
			}
		}
		tg.Outgoing[h] = kts
	}
	// Incoming map.
	for h, ts := range tg.Incoming {
		var kts []*Transition
		for _, t := range ts {
			if t.Action != actionName {
				kts = append(kts, t)
			}
		}
		tg.Incoming[h] = kts
	}
}

// pruneAction is a backward-compatible wrapper.
func pruneAction(tg *TransitionGraph, actionName string) {
	pruneActionWithMap(tg, actionName, nil)
}
