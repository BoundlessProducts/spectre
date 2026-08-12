// Package itf implements the Informal Trace Format (ITF) for Spectre.
// Spec: https://apalache-mc.org/docs/adr/015adr-trace.html
//
// ITF is a JSON-based exchange format for execution traces. Spectre uses it
// as the wire format between the verifier and the spectre-connect Rust crate,
// which replays traces against real implementations for model-based testing.
package itf

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/state"
)

// Trace is a single ITF execution trace.
type Trace struct {
	meta        traceMeta
	vars        []string
	rawStates   []map[string]interface{}
}

type traceMeta struct {
	Format            string `json:"format"`
	FormatDescription string `json:"format-description"`
	Source            string `json:"source,omitempty"`
	Description       string `json:"description,omitempty"`
}

// RawStates returns the raw ITF state maps suitable for minimality checking.
func (t *Trace) RawStates() []map[string]interface{} { return t.rawStates }

// MarshalJSON produces ITF-compliant JSON where #meta, vars, and states are
// all top-level keys in the same object.
func (t *Trace) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"#meta":  t.meta,
		"vars":   t.vars,
		"states": t.rawStates,
	})
}

// FromViolation builds an ITF trace from an invariant violation counterexample.
func FromViolation(specSource string, v *explore.Violation) *Trace {
	desc := fmt.Sprintf("Counterexample for invariant: %s", v.Invariant)
	if len(v.Path) == 0 {
		vars := collectVarNames(v.State)
		return &Trace{
			meta:      newMeta(specSource, desc),
			vars:      vars,
			rawStates: []map[string]interface{}{encodeState(v.State, 0, "", nil)},
		}
	}

	vars := collectVarNames(v.Path[0].FromState)
	states := make([]map[string]interface{}, 0, len(v.Path)+1)
	states = append(states, encodeState(v.Path[0].FromState, 0, "", nil))
	for i, trans := range v.Path {
		states = append(states, encodeState(trans.ToState, i+1, trans.Action, trans.Args))
	}
	return &Trace{meta: newMeta(specSource, desc), vars: vars, rawStates: states}
}

// RandomWalk generates a valid ITF execution trace by performing a
// coverage-directed walk through the transition graph.  It prefers
// transitions to actions not yet seen, falling back to random choices when
// all reachable actions have been covered.  This produces a positive MBT
// witness: a trace that exercises as many distinct actions as possible.
func RandomWalk(
	specSource string,
	graph *explore.TransitionGraph,
	initialStates []*state.State,
	maxLen int,
	rng *rand.Rand,
	hasher *explore.StateHasher,
) *Trace {
	if len(initialStates) == 0 {
		return nil
	}

	start := initialStates[rng.Intn(len(initialStates))]
	vars := collectVarNames(start)
	states := make([]map[string]interface{}, 0, maxLen+1)
	states = append(states, encodeState(start, 0, "", nil))

	seenActions := make(map[string]bool)
	visited := make(map[string]bool)
	current := start

	for i := 0; i < maxLen; i++ {
		h := hasher.HashState(current)
		if visited[h] {
			break
		}
		visited[h] = true

		node := graph.GetStateNode(h)
		if node == nil || len(node.Outgoing) == 0 {
			break
		}

		// Prefer a transition to an unseen action; fall back to any.
		var chosen *explore.Transition
		for _, t := range node.Outgoing {
			if !seenActions[t.Action] {
				chosen = t
				break
			}
		}
		if chosen == nil {
			chosen = node.Outgoing[rng.Intn(len(node.Outgoing))]
		}

		seenActions[chosen.Action] = true
		states = append(states, encodeState(chosen.ToState, i+1, chosen.Action, chosen.Args))
		current = chosen.ToState
	}

	return &Trace{
		meta:      newMeta(specSource, fmt.Sprintf("Valid execution witness (%d actions covered)", len(seenActions))),
		vars:      vars,
		rawStates: states,
	}
}

// PropertyDirectedWalk generates a trace that biases toward states near invariant
// violation boundaries. At each step it scores candidate successor states by
// their violation-proximity: the minimum integer variable value across all
// variables. States with lower minima are preferred, driving the trace toward
// the boundary conditions where integer invariants (e.g. balance >= 0) are
// hardest to satisfy. Falls back to action-coverage when all successors score
// the same.
func PropertyDirectedWalk(
	specSource string,
	graph *explore.TransitionGraph,
	initialStates []*state.State,
	maxLen int,
	rng *rand.Rand,
	hasher *explore.StateHasher,
) *Trace {
	if len(initialStates) == 0 {
		return nil
	}
	start := initialStates[rng.Intn(len(initialStates))]
	vars := collectVarNames(start)
	states := make([]map[string]interface{}, 0, maxLen+1)
	states = append(states, encodeState(start, 0, "", nil))

	current := start
	seenActions := make(map[string]bool)
	stepsAtBoundary := 0

	for i := 0; i < maxLen; i++ {
		h := hasher.HashState(current)
		node := graph.GetStateNode(h)
		if node == nil || len(node.Outgoing) == 0 {
			break
		}

		// Score each candidate by violation proximity (lower score = closer to violation).
		type candidate struct {
			trans *explore.Transition
			score float64
		}
		candidates := make([]candidate, len(node.Outgoing))
		for j, t := range node.Outgoing {
			candidates[j] = candidate{trans: t, score: violationProximityScore(t.ToState)}
		}

		// Find the best (lowest) score.
		best := math.MaxFloat64
		for _, c := range candidates {
			if c.score < best {
				best = c.score
			}
		}

		// Among ties at best score, prefer an unseen action.
		var chosen *explore.Transition
		for _, c := range candidates {
			if c.score == best && !seenActions[c.trans.Action] {
				chosen = c.trans
				break
			}
		}
		if chosen == nil {
			for _, c := range candidates {
				if c.score == best {
					chosen = c.trans
					break
				}
			}
		}

		if violationProximityScore(chosen.ToState) <= 0 {
			stepsAtBoundary++
		}
		seenActions[chosen.Action] = true
		states = append(states, encodeState(chosen.ToState, i+1, chosen.Action, chosen.Args))
		current = chosen.ToState
	}

	return &Trace{
		meta: newMeta(specSource, fmt.Sprintf(
			"Property-directed walk (%d steps at invariant boundary)", stepsAtBoundary)),
		vars:      vars,
		rawStates: states,
	}
}

// violationProximityScore returns the minimum integer variable value in a state.
// Lower values indicate proximity to violating non-negativity invariants.
func violationProximityScore(s *state.State) float64 {
	minVal := math.MaxFloat64
	for _, v := range s.Variables {
		pv, ok := v.(*state.PrimitiveValue)
		if !ok || pv.TypeName != "int" || pv.IntValue == nil {
			continue
		}
		fv := float64(*pv.IntValue)
		if fv < minVal {
			minVal = fv
		}
	}
	if minVal == math.MaxFloat64 {
		return 0
	}
	return minVal
}

// SerializeValue converts a Spectre runtime value to its ITF JSON representation.
// Callers outside this package (e.g. tests) can use this directly.
func SerializeValue(v state.Value) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case *state.PrimitiveValue:
		return serializePrimitive(val)
	case *state.EnumValue:
		// ITF encodes enum variants as plain strings (just the variant name).
		return val.ValueName
	case *state.SetValue:
		elems := make([]interface{}, len(val.Values))
		for i, e := range val.Values {
			elems[i] = SerializeValue(e)
		}
		return map[string]interface{}{"#set": elems}
	case *state.ListValue:
		elems := make([]interface{}, len(val.Elements))
		for i, e := range val.Elements {
			elems[i] = SerializeValue(e)
		}
		return map[string]interface{}{"#tup": elems}
	case *state.MapValue:
		// Both records ({field: val}) and maps are stored as MapValue.
		// Serialize as ITF #map to preserve non-string keys.
		keys := make([]string, 0, len(val.Entries))
		for k := range val.Entries {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		entries := make([]interface{}, 0, len(keys))
		for _, k := range keys {
			entries = append(entries, []interface{}{k, SerializeValue(val.Entries[k])})
		}
		return map[string]interface{}{"#map": entries}
	}
	return v.String()
}

// ---- helpers ----------------------------------------------------------------

func serializePrimitive(v *state.PrimitiveValue) interface{} {
	switch v.TypeName {
	case "int":
		n := int64(0)
		if v.IntValue != nil {
			n = *v.IntValue
		}
		// ITF encodes integers as {"#bigint": "<decimal>"} to avoid JSON number limits.
		return map[string]string{"#bigint": fmt.Sprintf("%d", n)}
	case "bool":
		if v.BoolValue != nil {
			return *v.BoolValue
		}
		return false
	case "str":
		if v.StringValue != nil {
			return *v.StringValue
		}
		return ""
	case "float":
		if v.FloatValue != nil {
			return *v.FloatValue
		}
		return 0.0
	}
	return nil
}

func encodeState(s *state.State, index int, action string, args []state.Value) map[string]interface{} {
	meta := map[string]interface{}{
		"index":  index,
		"action": action,
	}
	if len(args) > 0 {
		encoded := make([]interface{}, len(args))
		for i, a := range args {
			encoded[i] = SerializeValue(a)
		}
		meta["actionArgs"] = encoded
	}

	m := map[string]interface{}{"#meta": meta}

	// Sort variable names for deterministic output.
	names := make([]string, 0, len(s.Variables))
	for name := range s.Variables {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		m[name] = SerializeValue(s.Variables[name])
	}
	return m
}

// NewTrace creates an empty ITF trace with the given source and description.
func NewTrace(specSource, description string) *Trace {
	return &Trace{
		meta:      newMeta(specSource, description),
		rawStates: nil,
	}
}

// AppendStep appends one state to a trace. stepIdx is 0 for the initial state.
func AppendStep(t *Trace, s *state.State, stepIdx int, action string, args []state.Value) {
	if len(t.vars) == 0 && s != nil {
		t.vars = collectVarNames(s)
	}
	t.rawStates = append(t.rawStates, encodeState(s, stepIdx, action, args))
}

func collectVarNames(s *state.State) []string {
	names := make([]string, 0, len(s.Variables))
	for name := range s.Variables {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func newMeta(source, description string) traceMeta {
	return traceMeta{
		Format:            "ITF",
		FormatDescription: "https://apalache-mc.org/docs/adr/015adr-trace.html",
		Source:            source,
		Description:       description,
	}
}

// TransitionPairWalk generates a trace that maximises coverage of (prev_action, curr_action)
// transition pairs. At each step it prefers the action that creates a new pair.
func TransitionPairWalk(
	specSource string,
	graph *explore.TransitionGraph,
	initialStates []*state.State,
	maxLen int,
	rng *rand.Rand,
	hasher *explore.StateHasher,
) *Trace {
	if len(initialStates) == 0 {
		return nil
	}
	start := initialStates[rng.Intn(len(initialStates))]
	vars := collectVarNames(start)
	states := make([]map[string]interface{}, 0, maxLen+1)
	states = append(states, encodeState(start, 0, "", nil))

	seenPairs := make(map[[2]string]bool)
	current := start
	prevAction := ""

	for i := 0; i < maxLen; i++ {
		h := hasher.HashState(current)
		node := graph.GetStateNode(h)
		if node == nil || len(node.Outgoing) == 0 {
			break
		}
		// Prefer a transition that produces a new (prevAction, currAction) pair.
		var chosen *explore.Transition
		for _, t := range node.Outgoing {
			pair := [2]string{prevAction, t.Action}
			if !seenPairs[pair] {
				chosen = t
				break
			}
		}
		if chosen == nil {
			chosen = node.Outgoing[rng.Intn(len(node.Outgoing))]
		}
		seenPairs[[2]string{prevAction, chosen.Action}] = true
		prevAction = chosen.Action
		states = append(states, encodeState(chosen.ToState, i+1, chosen.Action, chosen.Args))
		current = chosen.ToState
	}

	return &Trace{
		meta:      newMeta(specSource, fmt.Sprintf("Transition-pair coverage walk (%d pairs covered)", len(seenPairs))),
		vars:      vars,
		rawStates: states,
	}
}

// BoundaryCoverageWalk generates a trace that attempts to visit states where integer
// variables hit their observed min or max values across the entire state graph.
func BoundaryCoverageWalk(
	specSource string,
	graph *explore.TransitionGraph,
	initialStates []*state.State,
	maxLen int,
	rng *rand.Rand,
	hasher *explore.StateHasher,
) *Trace {
	if len(initialStates) == 0 {
		return nil
	}

	// Scan all graph states to find observed min/max per integer variable.
	type minmax struct{ min, max int64 }
	bounds := map[string]*minmax{}
	for _, node := range graph.States {
		if node == nil || node.State == nil {
			continue
		}
		for name, val := range node.State.Variables {
			pv, ok := val.(*state.PrimitiveValue)
			if !ok || pv.TypeName != "int" || pv.IntValue == nil {
				continue
			}
			v := *pv.IntValue
			if b, exists := bounds[name]; exists {
				if v < b.min {
					b.min = v
				}
				if v > b.max {
					b.max = v
				}
			} else {
				bounds[name] = &minmax{min: v, max: v}
			}
		}
	}

	// Build a set of "target" state hashes that represent boundary states.
	targets := map[string]bool{}
	for _, node := range graph.States {
		if node == nil || node.State == nil {
			continue
		}
		for name, val := range node.State.Variables {
			pv, ok := val.(*state.PrimitiveValue)
			if !ok || pv.TypeName != "int" || pv.IntValue == nil {
				continue
			}
			b, exists := bounds[name]
			if !exists {
				continue
			}
			if *pv.IntValue == b.min || *pv.IntValue == b.max {
				targets[node.Hash] = true
				break
			}
		}
	}

	start := initialStates[rng.Intn(len(initialStates))]
	vars := collectVarNames(start)
	states := make([]map[string]interface{}, 0, maxLen+1)
	states = append(states, encodeState(start, 0, "", nil))

	visitedBoundaries := 0
	current := start
	seenActions := make(map[string]bool)

	for i := 0; i < maxLen; i++ {
		h := hasher.HashState(current)
		node := graph.GetStateNode(h)
		if node == nil || len(node.Outgoing) == 0 {
			break
		}
		// Prefer a transition leading toward an unvisited boundary state.
		var chosen *explore.Transition
		for _, t := range node.Outgoing {
			toH := hasher.HashState(t.ToState)
			if targets[toH] {
				chosen = t
				break
			}
		}
		if chosen == nil {
			// Fall back to action-coverage walk.
			for _, t := range node.Outgoing {
				if !seenActions[t.Action] {
					chosen = t
					break
				}
			}
		}
		if chosen == nil {
			chosen = node.Outgoing[rng.Intn(len(node.Outgoing))]
		}
		toH := hasher.HashState(chosen.ToState)
		if targets[toH] {
			visitedBoundaries++
		}
		seenActions[chosen.Action] = true
		states = append(states, encodeState(chosen.ToState, i+1, chosen.Action, chosen.Args))
		current = chosen.ToState
	}

	return &Trace{
		meta:      newMeta(specSource, fmt.Sprintf("Boundary coverage walk (%d boundary states visited)", visitedBoundaries)),
		vars:      vars,
		rawStates: states,
	}
}

// RareActionWalk generates a trace that preferentially exercises the least-frequently
// taken actions, driving the walk toward rarely-reachable transitions.
func RareActionWalk(
	specSource string,
	graph *explore.TransitionGraph,
	initialStates []*state.State,
	maxLen int,
	rng *rand.Rand,
	hasher *explore.StateHasher,
) *Trace {
	if len(initialStates) == 0 {
		return nil
	}
	start := initialStates[rng.Intn(len(initialStates))]
	vars := collectVarNames(start)
	states := make([]map[string]interface{}, 0, maxLen+1)
	states = append(states, encodeState(start, 0, "", nil))

	counts := map[string]int{} // action name → times taken
	current := start

	for i := 0; i < maxLen; i++ {
		h := hasher.HashState(current)
		node := graph.GetStateNode(h)
		if node == nil || len(node.Outgoing) == 0 {
			break
		}
		// Find the minimum count among available actions.
		minCount := -1
		for _, t := range node.Outgoing {
			c := counts[t.Action]
			if minCount < 0 || c < minCount {
				minCount = c
			}
		}
		// Collect all transitions tied at the minimum count and pick one randomly.
		var candidates []*explore.Transition
		for _, t := range node.Outgoing {
			if counts[t.Action] == minCount {
				candidates = append(candidates, t)
			}
		}
		chosen := candidates[rng.Intn(len(candidates))]
		counts[chosen.Action]++
		states = append(states, encodeState(chosen.ToState, i+1, chosen.Action, chosen.Args))
		current = chosen.ToState
	}

	return &Trace{
		meta:      newMeta(specSource, "Rare-action coverage walk (least-frequently-taken transitions preferred)"),
		vars:      vars,
		rawStates: states,
	}
}
