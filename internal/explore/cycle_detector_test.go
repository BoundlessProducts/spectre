package explore

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/state"
)

func TestCycleDetectorNoCycle(t *testing.T) {
	hasher := NewStateHasher()
	detector := NewCycleDetector(hasher)

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(1))

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(2))

	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(3))

	detector.PushState(s1)
	detector.PushState(s2)
	detector.PushState(s3)

	// Check a different state (s4) that's not in the path - should not create a cycle
	s4 := state.NewState()
	s4.SetVariable("counter", state.NewIntValue(4))
	
	if detector.HasCycle(s4) {
		t.Error("should not detect cycle for different state")
	}
}

func TestCycleDetectorCycle(t *testing.T) {
	hasher := NewStateHasher()
	detector := NewCycleDetector(hasher)

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(1))

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(2))

	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(3))

	detector.PushState(s1)
	detector.PushState(s2)
	detector.PushState(s3)

	// Push s1 again - should detect cycle
	if !detector.HasCycle(s1) {
		t.Error("should detect cycle when state repeats")
	}
}

func TestCycleDetectorCyclePath(t *testing.T) {
	hasher := NewStateHasher()
	detector := NewCycleDetector(hasher)

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(1))

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(2))

	s3 := state.NewState()
	s3.SetVariable("counter", state.NewIntValue(3))

	detector.PushState(s1)
	detector.PushState(s2)
	detector.PushState(s3)

	// Check cycle path
	cyclePath := detector.GetCyclePath(s1)
	if len(cyclePath) == 0 {
		t.Error("should return cycle path")
	}

	if len(cyclePath) != 4 {
		t.Errorf("expected cycle path length 4, got %d", len(cyclePath))
	}
}

func TestCycleDetectorPushPop(t *testing.T) {
	hasher := NewStateHasher()
	detector := NewCycleDetector(hasher)

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(1))

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(2))

	detector.PushState(s1)
	detector.PushState(s2)

	// s1 is in the path but not at the end, so HasCycle(s1) should return true
	// because adding s1 would create a cycle back to an earlier state
	if !detector.HasCycle(s1) {
		t.Error("should detect cycle when state appears earlier in path")
	}

	detector.PopState()

	// After popping s2, s1 is at the end of the path
	// Adding s1 again would create a self-loop, so HasCycle should return true
	if !detector.HasCycle(s1) {
		t.Error("should detect cycle when adding same state again")
	}
}

func TestCycleDetectorReset(t *testing.T) {
	hasher := NewStateHasher()
	detector := NewCycleDetector(hasher)

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(1))

	detector.PushState(s1)
	detector.Reset()

	if detector.HasCycle(s1) {
		t.Error("should not detect cycle after reset")
	}
}

