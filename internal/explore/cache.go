package explore

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/akkeshavan/spectre/internal/state"
)

// GraphCache persists the explored state graph to disk so subsequent verify
// runs can skip re-exploring unchanged portions of the spec.
type GraphCache struct {
	SpecHash    string                 `json:"spec_hash"`
	ConfigHash  string                 `json:"config_hash"` // hash of params + exploration limits
	ExploreTime time.Time              `json:"explore_time"`
	States      map[string]CachedState `json:"states"`
	Transitions []CachedTransition     `json:"transitions"`
	Violations  []string               `json:"violations"`
}

// CacheConfig holds the exploration configuration that must match for a cache hit.
type CacheConfig struct {
	Params     map[string]int64 // spec parameter bindings (e.g. N=3)
	MaxStates  int
	MaxDepth   int
	ParamBound int
}

// CachedState is a serializable snapshot of one explored state.
type CachedState struct {
	Hash      string                 `json:"hash"`
	Variables map[string]interface{} `json:"variables"`
	InCycle   bool                   `json:"in_cycle"`
}

// CachedTransition records one edge in the cached graph.
type CachedTransition struct {
	FromHash string `json:"from"`
	ToHash   string `json:"to"`
	Action   string `json:"action"`
}

// CachePath returns the filesystem path for the cache file of a given spec.
// e.g. "examples/bank.spec" → "examples/.spectre-cache/bank.spec.cache.json"
func CachePath(specFile string) string {
	dir := filepath.Join(filepath.Dir(specFile), ".spectre-cache")
	base := filepath.Base(specFile) + ".cache.json"
	return filepath.Join(dir, base)
}

// HashSpec returns the hex SHA-256 of content, used for cache invalidation.
func HashSpec(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("%x", sum)
}

// HashConfig returns a stable hex SHA-256 of the exploration configuration,
// so caches are invalidated when --param, --max-states, --max-depth, or
// --param-bound change even if the spec file is identical.
func HashConfig(cfg CacheConfig) string {
	// Stable serialisation: sorted param keys, then fixed fields.
	keys := make([]string, 0, len(cfg.Params))
	for k := range cfg.Params {
		keys = append(keys, k)
	}
	// sort inline without importing sort (keep imports minimal)
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	var b []byte
	for _, k := range keys {
		b = append(b, []byte(fmt.Sprintf("%s=%d;", k, cfg.Params[k]))...)
	}
	b = append(b, []byte(fmt.Sprintf("ms=%d;md=%d;pb=%d", cfg.MaxStates, cfg.MaxDepth, cfg.ParamBound))...)
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum)
}

// LoadCache reads a previously saved cache for specFile.
// Returns nil (without error) if the cache doesn't exist or is stale.
func LoadCache(specFile string, cfg CacheConfig) (*GraphCache, error) {
	cachePath := CachePath(specFile)
	data, err := os.ReadFile(cachePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var cache GraphCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, nil // treat corrupt cache as absent
	}

	// Invalidate if spec content changed.
	specData, err := os.ReadFile(specFile)
	if err != nil {
		return nil, nil
	}
	if HashSpec(specData) != cache.SpecHash {
		return nil, nil // stale
	}
	// Invalidate if exploration configuration changed (params, limits).
	if HashConfig(cfg) != cache.ConfigHash {
		return nil, nil // different configuration
	}
	return &cache, nil
}

// SaveCache persists an ExplorationResult so the next run can reuse it.
func SaveCache(specFile string, result *ExplorationResult, specContent []byte, cfg CacheConfig) error {
	dir := filepath.Join(filepath.Dir(specFile), ".spectre-cache")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	hasher := NewStateHasher()

	cachedStates := make(map[string]CachedState, len(result.TransitionGraph.States))
	for hash, node := range result.TransitionGraph.States {
		vars := stateToMap(node.State)
		cachedStates[hash] = CachedState{
			Hash:      hash,
			Variables: vars,
			InCycle:   node.InCycle,
		}
	}

	cachedTransitions := make([]CachedTransition, 0, len(result.TransitionGraph.Transitions))
	for _, t := range result.TransitionGraph.Transitions {
		cachedTransitions = append(cachedTransitions, CachedTransition{
			FromHash: hasher.HashState(t.FromState),
			ToHash:   hasher.HashState(t.ToState),
			Action:   t.Action,
		})
	}

	violations := make([]string, 0, len(result.Violations))
	for _, v := range result.Violations {
		violations = append(violations, v.Description)
	}

	cache := GraphCache{
		SpecHash:    HashSpec(specContent),
		ConfigHash:  HashConfig(cfg),
		ExploreTime: time.Now(),
		States:      cachedStates,
		Transitions: cachedTransitions,
		Violations:  violations,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(CachePath(specFile), data, 0644)
}

// RestoreResult reconstructs a minimal ExplorationResult from a GraphCache.
// The restored result has StatesExplored populated and the transition graph
// rebuilt, but State pointers use placeholder values (variables only, no AST).
func RestoreResult(cache *GraphCache) *ExplorationResult {
	tg := NewTransitionGraph()

	// Rebuild state pool.
	statePool := make(map[string]*state.State, len(cache.States))
	for hash, cs := range cache.States {
		s := mapToState(cs.Variables)
		statePool[hash] = s
		node := tg.AddState(s, hash)
		node.InCycle = cs.InCycle
	}

	// Rebuild transitions.
	for _, ct := range cache.Transitions {
		from, fok := statePool[ct.FromHash]
		to, tok := statePool[ct.ToHash]
		if !fok || !tok {
			continue
		}
		trans := &Transition{FromState: from, ToState: to, Action: ct.Action}
		tg.AddTransition(trans, ct.FromHash, ct.ToHash)
	}

	return &ExplorationResult{
		StatesExplored:  len(cache.States),
		StatesVisited:   len(cache.States),
		Violations:      []*Violation{},
		Stuttering:      []*Stuttering{},
		ReachableStates: tg.GetAllStates(),
		InitialStates:   []*state.State{},
		MaxDepth:        0,
		Cycles:          []*Cycle{},
		TransitionGraph: tg,
	}
}

// ---- helpers ----------------------------------------------------------------

func stateToMap(s *state.State) map[string]interface{} {
	m := make(map[string]interface{}, len(s.Variables))
	for k, v := range s.Variables {
		switch pv := v.(type) {
		case *state.PrimitiveValue:
			switch pv.TypeName {
			case "int":
				if pv.IntValue != nil {
					m[k] = *pv.IntValue
				}
			case "bool":
				if pv.BoolValue != nil {
					m[k] = *pv.BoolValue
				}
			case "float":
				if pv.FloatValue != nil {
					m[k] = *pv.FloatValue
				}
			case "str":
				if pv.StringValue != nil {
					m[k] = *pv.StringValue
				}
			}
		default:
			m[k] = v.String()
		}
	}
	return m
}

func mapToState(m map[string]interface{}) *state.State {
	vars := make(map[string]state.Value, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case float64:
			i := int64(val)
			vars[k] = &state.PrimitiveValue{TypeName: "int", IntValue: &i}
		case bool:
			b := val
			vars[k] = &state.PrimitiveValue{TypeName: "bool", BoolValue: &b}
		case string:
			// Detect enum values serialised as "EnumName.ValueName".
			if dot := strings.Index(val, "."); dot > 0 {
				enumName := val[:dot]
				valueName := val[dot+1:]
				vars[k] = state.NewEnumValue(enumName, valueName)
			} else {
				s := val
				vars[k] = &state.PrimitiveValue{TypeName: "str", StringValue: &s}
			}
		case int64:
			i := val
			vars[k] = &state.PrimitiveValue{TypeName: "int", IntValue: &i}
		}
	}
	return &state.State{Variables: vars}
}
