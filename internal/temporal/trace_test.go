package temporal

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/state"
)

func TestTraceBasic(t *testing.T) {
	trace := NewTrace()
	
	if trace.Length() != 0 {
		t.Errorf("expected empty trace, got length %d", trace.Length())
	}
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	
	trace.AddState(s1, "", nil)
	
	if trace.Length() != 1 {
		t.Errorf("expected trace length 1, got %d", trace.Length())
	}
	
	if trace.GetState(0) != s1 {
		t.Error("expected state s1 at index 0")
	}
}

func TestTraceMultipleStates(t *testing.T) {
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(1))
	
	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(2))
	
	trace.AddState(s1, "", nil)
	trace.AddState(s2, "increment", nil)
	trace.AddState(s3, "increment", nil)
	
	if trace.Length() != 3 {
		t.Errorf("expected trace length 3, got %d", trace.Length())
	}
	
	if trace.GetAction(0) != "" {
		t.Error("expected empty action for initial state")
	}
	
	if trace.GetAction(1) != "increment" {
		t.Errorf("expected action 'increment', got '%s'", trace.GetAction(1))
	}
}

func TestTracePosition(t *testing.T) {
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(1))
	
	trace.AddState(s1, "", nil)
	trace.AddState(s2, "increment", nil)
	
	if trace.CurrentState() != s1 {
		t.Error("expected current state to be s1 initially")
	}
	
	next := trace.NextState()
	if next != s2 {
		t.Error("expected next state to be s2")
	}
	
	if trace.CurrentState() != s2 {
		t.Error("expected current state to be s2 after NextState")
	}
	
	prev := trace.PreviousState()
	if prev != s1 {
		t.Error("expected previous state to be s1")
	}
	
	if trace.CurrentState() != s1 {
		t.Error("expected current state to be s1 after PreviousState")
	}
}

func TestTraceReset(t *testing.T) {
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(1))
	
	trace.AddState(s1, "", nil)
	trace.AddState(s2, "increment", nil)
	
	trace.NextState()
	trace.Reset()
	
	if trace.CurrentState() != s1 {
		t.Error("expected current state to be s1 after reset")
	}
}

func TestTraceIsComplete(t *testing.T) {
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(1))
	
	trace.AddState(s1, "", nil)
	trace.AddState(s2, "increment", nil)
	
	if trace.IsComplete() {
		t.Error("trace should not be complete at position 0")
	}
	
	trace.NextState()
	if !trace.IsComplete() {
		t.Error("trace should be complete at last position")
	}
}

func TestTraceCopy(t *testing.T) {
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(1))
	
	trace.AddState(s1, "", nil)
	trace.AddState(s2, "increment", nil)
	trace.NextState()
	
	copy := trace.Copy()
	
	if copy.Length() != trace.Length() {
		t.Error("copy should have same length")
	}
	
	if copy.Position != trace.Position {
		t.Error("copy should have same position")
	}
	
	// Modify original
	trace.Reset()
	
	if copy.Position == trace.Position {
		t.Error("copy should be independent")
	}
}

func TestTraceSlice(t *testing.T) {
	trace := NewTrace()
	
	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(0))
	
	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(1))
	
	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(2))
	
	trace.AddState(s1, "", nil)
	trace.AddState(s2, "increment", nil)
	trace.AddState(s3, "increment", nil)
	
	slice := trace.Slice(1, 3)
	
	if slice.Length() != 2 {
		t.Errorf("expected slice length 2, got %d", slice.Length())
	}
	
	if slice.GetState(0) != s2 {
		t.Error("expected slice to start with s2")
	}
}

