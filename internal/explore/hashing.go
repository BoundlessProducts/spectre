package explore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/spectre-lang/spectre/internal/state"
)

// StateHasher provides efficient state hashing for comparison and deduplication
type StateHasher struct {
	// Cache for computed hashes
	hashCache map[*state.State]string
}

// NewStateHasher creates a new state hasher
func NewStateHasher() *StateHasher {
	return &StateHasher{
		hashCache: make(map[*state.State]string),
	}
}

// HashState computes a hash for a state
// Uses SHA-256 for collision resistance
func (sh *StateHasher) HashState(s *state.State) string {
	// Check cache first
	if hash, exists := sh.hashCache[s]; exists {
		return hash
	}

	// Compute hash
	hash := sh.computeHash(s)
	sh.hashCache[s] = hash
	return hash
}

// computeHash computes the hash of a state
func (sh *StateHasher) computeHash(s *state.State) string {
	// Get all variable names and sort them for consistent hashing
	varNames := make([]string, 0, len(s.Variables))
	for name := range s.Variables {
		varNames = append(varNames, name)
	}
	sort.Strings(varNames)

	// Build hash string from sorted variables
	hashStr := ""
	for _, name := range varNames {
		value := s.Variables[name]
		hashStr += fmt.Sprintf("%s:%s;", name, value.String())
	}

	// Compute SHA-256 hash
	hasher := sha256.New()
	hasher.Write([]byte(hashStr))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

// StatesEqual checks if two states are equal using their hashes
func (sh *StateHasher) StatesEqual(s1, s2 *state.State) bool {
	return sh.HashState(s1) == sh.HashState(s2)
}

// ClearCache clears the hash cache
func (sh *StateHasher) ClearCache() {
	sh.hashCache = make(map[*state.State]string)
}

