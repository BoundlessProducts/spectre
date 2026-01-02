package temporal

import (
	"github.com/spectre-lang/spectre/internal/state"
)

// Trace represents an execution trace (sequence of states)
type Trace struct {
	States   []*state.State
	Actions  []string // Action names that led to each transition
	Args     [][]state.Value // Arguments for each action (if any)
	Position int // Current position in trace
}

// NewTrace creates a new empty trace
func NewTrace() *Trace {
	return &Trace{
		States:   []*state.State{},
		Actions:  []string{},
		Args:     [][]state.Value{},
		Position: 0,
	}
}

// AddState adds a state to the trace
func (t *Trace) AddState(s *state.State, action string, args []state.Value) {
	t.States = append(t.States, s)
	t.Actions = append(t.Actions, action)
	t.Args = append(t.Args, args)
}

// Length returns the length of the trace
func (t *Trace) Length() int {
	return len(t.States)
}

// GetState returns the state at the given index
func (t *Trace) GetState(index int) *state.State {
	if index < 0 || index >= len(t.States) {
		return nil
	}
	return t.States[index]
}

// GetAction returns the action at the given index
func (t *Trace) GetAction(index int) string {
	if index < 0 || index >= len(t.Actions) {
		return ""
	}
	return t.Actions[index]
}

// GetArgs returns the arguments at the given index
func (t *Trace) GetArgs(index int) []state.Value {
	if index < 0 || index >= len(t.Args) {
		return nil
	}
	return t.Args[index]
}

// CurrentState returns the current state (at Position)
func (t *Trace) CurrentState() *state.State {
	return t.GetState(t.Position)
}

// NextState advances to the next state and returns it
func (t *Trace) NextState() *state.State {
	if t.Position+1 < len(t.States) {
		t.Position++
		return t.CurrentState()
	}
	return nil
}

// PreviousState goes back to the previous state and returns it
func (t *Trace) PreviousState() *state.State {
	if t.Position > 0 {
		t.Position--
		return t.CurrentState()
	}
	return nil
}

// Reset resets the position to the beginning
func (t *Trace) Reset() {
	t.Position = 0
}

// IsComplete returns true if the trace is complete (position at end)
func (t *Trace) IsComplete() bool {
	return t.Position >= len(t.States)-1
}

// Copy creates a copy of the trace
func (t *Trace) Copy() *Trace {
	newTrace := &Trace{
		States:   make([]*state.State, len(t.States)),
		Actions:  make([]string, len(t.Actions)),
		Args:     make([][]state.Value, len(t.Args)),
		Position: t.Position,
	}
	
	copy(newTrace.States, t.States)
	copy(newTrace.Actions, t.Actions)
	for i, args := range t.Args {
		newTrace.Args[i] = make([]state.Value, len(args))
		copy(newTrace.Args[i], args)
	}
	
	return newTrace
}

// Slice returns a slice of the trace from start to end
func (t *Trace) Slice(start, end int) *Trace {
	if start < 0 {
		start = 0
	}
	if end > len(t.States) {
		end = len(t.States)
	}
	if start >= end {
		return NewTrace()
	}
	
	newTrace := &Trace{
		States:   make([]*state.State, end-start),
		Actions:  make([]string, end-start-1),
		Args:     make([][]state.Value, end-start-1),
		Position: 0,
	}
	
	copy(newTrace.States, t.States[start:end])
	if end-start > 1 {
		copy(newTrace.Actions, t.Actions[start:end-1])
		for i := start; i < end-1; i++ {
			newTrace.Args[i-start] = make([]state.Value, len(t.Args[i]))
			copy(newTrace.Args[i-start], t.Args[i])
		}
	}
	
	return newTrace
}

