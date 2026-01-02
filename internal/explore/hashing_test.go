package explore

import (
	"testing"

	"github.com/akkeshavan/spectre/internal/state"
)

func TestStateHasherSameState(t *testing.T) {
	hasher := NewStateHasher()

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(5))
	s1.SetVariable("flag", state.NewBoolValue(true))

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(5))
	s2.SetVariable("flag", state.NewBoolValue(true))

	hash1 := hasher.HashState(s1)
	hash2 := hasher.HashState(s2)

	if hash1 != hash2 {
		t.Error("same states should have same hash")
	}

	if !hasher.StatesEqual(s1, s2) {
		t.Error("StatesEqual should return true for identical states")
	}
}

func TestStateHasherDifferentState(t *testing.T) {
	hasher := NewStateHasher()

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(5))

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(10))

	hash1 := hasher.HashState(s1)
	hash2 := hasher.HashState(s2)

	if hash1 == hash2 {
		t.Error("different states should have different hashes")
	}

	if hasher.StatesEqual(s1, s2) {
		t.Error("StatesEqual should return false for different states")
	}
}

func TestStateHasherOrderIndependent(t *testing.T) {
	hasher := NewStateHasher()

	s1 := state.NewState()
	s1.SetVariable("x", state.NewIntValue(1))
	s1.SetVariable("y", state.NewIntValue(2))

	s2 := state.NewState()
	s2.SetVariable("y", state.NewIntValue(2))
	s2.SetVariable("x", state.NewIntValue(1))

	hash1 := hasher.HashState(s1)
	hash2 := hasher.HashState(s2)

	if hash1 != hash2 {
		t.Error("states with same variables in different order should have same hash")
	}
}

func TestStateHasherCache(t *testing.T) {
	hasher := NewStateHasher()

	s := state.NewState()
	s.SetVariable("counter", state.NewIntValue(5))

	hash1 := hasher.HashState(s)
	hash2 := hasher.HashState(s)

	if hash1 != hash2 {
		t.Error("cached hash should match")
	}

	// Hash should be retrieved from cache (same pointer)
	if len(hasher.hashCache) != 1 {
		t.Errorf("expected 1 cached hash, got %d", len(hasher.hashCache))
	}
}

func TestStateHasherDifferentTypes(t *testing.T) {
	hasher := NewStateHasher()

	s1 := state.NewState()
	s1.SetVariable("counter", state.NewIntValue(5))
	s1.SetVariable("name", state.NewStringValue("test"))

	s2 := state.NewState()
	s2.SetVariable("counter", state.NewIntValue(5))
	s2.SetVariable("name", state.NewStringValue("test"))

	hash1 := hasher.HashState(s1)
	hash2 := hasher.HashState(s2)

	if hash1 != hash2 {
		t.Error("states with same values of different types should have same hash")
	}
}

