package explore

import (
	"github.com/BoundlessProducts/spectre/internal/state"
)

// CycleDetector detects cycles in the state space
type CycleDetector struct {
	hasher *StateHasher
	path   []string // Path of state hashes from root to current node
}

// NewCycleDetector creates a new cycle detector
func NewCycleDetector(hasher *StateHasher) *CycleDetector {
	return &CycleDetector{
		hasher: hasher,
		path:   []string{},
	}
}

// HasCycle checks if adding a state would create a cycle
func (cd *CycleDetector) HasCycle(s *state.State) bool {
	hash := cd.hasher.HashState(s)
	
	// Check if this state hash appears in the current path
	for _, pathHash := range cd.path {
		if pathHash == hash {
			return true
		}
	}
	
	return false
}

// PushState adds a state to the current path
func (cd *CycleDetector) PushState(s *state.State) {
	hash := cd.hasher.HashState(s)
	cd.path = append(cd.path, hash)
}

// PopState removes the last state from the path
func (cd *CycleDetector) PopState() {
	if len(cd.path) > 0 {
		cd.path = cd.path[:len(cd.path)-1]
	}
}

// GetCyclePath returns the path that forms the cycle (if detected)
func (cd *CycleDetector) GetCyclePath(s *state.State) []string {
	hash := cd.hasher.HashState(s)
	
	// Find where the cycle starts
	cycleStart := -1
	for i, pathHash := range cd.path {
		if pathHash == hash {
			cycleStart = i
			break
		}
	}
	
	if cycleStart == -1 {
		return nil
	}
	
	// Return the cycle path
	cyclePath := make([]string, len(cd.path)-cycleStart+1)
	copy(cyclePath, cd.path[cycleStart:])
	cyclePath[len(cyclePath)-1] = hash
	
	return cyclePath
}

// Reset clears the path
func (cd *CycleDetector) Reset() {
	cd.path = []string{}
}

