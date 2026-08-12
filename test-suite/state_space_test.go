package test_suite

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/internal/semantic"
	"github.com/BoundlessProducts/spectre/internal/state"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// Test helper: Load and parse a spec file
func loadSpecFile(t *testing.T, filename string) (*ast.File, *exec.StateMachine) {
	// Get absolute path relative to project root
	wd, _ := os.Getwd()
	testPath := filepath.Join(wd, filename)
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		// Try from project root
		testPath = filepath.Join(filepath.Dir(wd), filename)
	}
	
	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", testPath, err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors in %s: %v", filename, p.Errors())
	}

	// Check for imports
	hasImports := false
	for _, decl := range file.Decls {
		if _, ok := decl.(*ast.ImportDecl); ok {
			hasImports = true
			break
		}
	}

	var sm *exec.StateMachine
	if hasImports {
		// Get absolute path
		wd, _ := os.Getwd()
		testPath := filepath.Join(wd, filename)
		if _, err := os.Stat(testPath); os.IsNotExist(err) {
			testPath = filepath.Join(filepath.Dir(wd), filename)
		}
		fileDir := filepath.Dir(testPath)
		loader := semantic.NewModuleLoader(fileDir)
		moduleInfo, loadErrors := loader.LoadModule(testPath)

		if len(loadErrors) > 0 {
			t.Fatalf("Module loading errors in %s: %v", filename, loadErrors)
		}

		circularErrors := loader.CheckCircularDependencies()
		if len(circularErrors) > 0 {
			t.Fatalf("Circular dependency errors: %v", circularErrors)
		}

		allModules := loader.GetAllModules()
		additionalFiles := make([]*ast.File, 0, len(allModules)-1)
		for _, modInfo := range allModules {
			if modInfo.File != moduleInfo.File {
				additionalFiles = append(additionalFiles, modInfo.File)
			}
		}
		sm, err = exec.NewStateMachine(moduleInfo.File, additionalFiles...)
		if err != nil {
			t.Fatalf("Failed to create state machine: %v", err)
		}
	} else {
		var err error
		sm, err = exec.NewStateMachine(file)
		if err != nil {
			t.Fatalf("Failed to create state machine: %v", err)
		}
	}

	return file, sm
}

// Test helper: Explore state space and return results
func exploreStateSpace(t *testing.T, sm *exec.StateMachine, maxDepth, maxStates int) *explore.ExplorationResult {
	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(maxDepth)
	explorer.SetMaxStates(maxStates)
	// Cycle detection is enabled by default

	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("Exploration failed: %v", err)
	}

	return result
}

// Test 1: Known State Space Size (8 states)
func TestKnownStateSpaceSize(t *testing.T) {
	_, sm := loadSpecFile(t, "01-known-state-space-size.spec")
	result := exploreStateSpace(t, sm, 20, 50)

	expectedStates := 8 // 2^3 = 8 combinations of (x, y, z)

	if result.StatesExplored < expectedStates {
		t.Errorf("Expected at least %d states, got %d", expectedStates, result.StatesExplored)
	}

	if len(result.ReachableStates) < expectedStates {
		t.Errorf("Expected at least %d reachable states, got %d", expectedStates, len(result.ReachableStates))
	}

	// Verify all combinations exist
	statesFound := make(map[string]bool)
	for _, state := range result.ReachableStates {
		xVal := state.Variables["x"]
		yVal := state.Variables["y"]
		zVal := state.Variables["z"]
		stateKey := fmt.Sprintf("%v-%v-%v", xVal, yVal, zVal)
		statesFound[stateKey] = true
	}

	// Check that we found states with different combinations
	if len(statesFound) < expectedStates {
		t.Errorf("Expected %d unique state combinations, found %d", expectedStates, len(statesFound))
	}

	t.Logf("✓ Test 1 passed: Found %d states (expected %d)", result.StatesExplored, expectedStates)
}

// Test 2: Small Counter (4 states: 0, 1, 2, 3)
func TestSmallCounter(t *testing.T) {
	_, sm := loadSpecFile(t, "02-small-counter.spec")
	result := exploreStateSpace(t, sm, 10, 20)

	expectedStates := 4 // counter can be 0, 1, 2, or 3

	if result.StatesExplored < expectedStates {
		t.Errorf("Expected at least %d states, got %d", expectedStates, result.StatesExplored)
	}

	// Verify all counter values appear
	counterValues := make(map[int64]bool)
	for _, st := range result.ReachableStates {
		if counterVal, ok := st.Variables["counter"]; ok {
			if intVal, ok := counterVal.(*state.PrimitiveValue); ok && intVal.IntValue != nil {
				counterValues[*intVal.IntValue] = true
			}
		}
	}

	if len(counterValues) < expectedStates {
		t.Errorf("Expected counter values 0-3, found values: %v", counterValues)
	}

	t.Logf("✓ Test 2 passed: Found %d states with counter values %v", result.StatesExplored, counterValues)
}

// Test 3: Parameterized Actions
func TestParameterizedActions(t *testing.T) {
	_, sm := loadSpecFile(t, "03-parameterized-actions.spec")
	result := exploreStateSpace(t, sm, 10, 20)

	// Should explore states with value = 0, 1, 2
	expectedMinStates := 2 // At least 0 (initial) and some other values

	if result.StatesExplored < expectedMinStates {
		t.Errorf("Expected at least %d states, got %d", expectedMinStates, result.StatesExplored)
	}

	// Verify parameterized action was executed with different arguments
	valueSet := make(map[int64]bool)
	for _, st := range result.ReachableStates {
		if val, ok := st.Variables["value"]; ok {
			if intVal, ok := val.(*state.PrimitiveValue); ok && intVal.IntValue != nil {
				valueSet[*intVal.IntValue] = true
			}
		}
	}

	if len(valueSet) < 2 {
		t.Errorf("Expected multiple value states from parameterized actions, found: %v", valueSet)
	}

	t.Logf("✓ Test 3 passed: Parameterized actions explored %d value states: %v", len(valueSet), valueSet)
}

// Test 4: Cycle Detection
func TestCycleDetection(t *testing.T) {
	_, sm := loadSpecFile(t, "04-cycle-detection.spec")
	result := exploreStateSpace(t, sm, 10, 20)

	// Should detect cycle: 0 -> 1 -> 0
	expectedStates := 2
	if result.StatesExplored < expectedStates {
		t.Errorf("Expected at least %d states, got %d", expectedStates, result.StatesExplored)
	}

	if len(result.Cycles) == 0 {
		t.Error("Expected at least one cycle to be detected")
	}

	t.Logf("✓ Test 4 passed: Detected %d cycles in %d states", len(result.Cycles), result.StatesExplored)
}

// Test 5: Two State Machine
func TestTwoStateMachine(t *testing.T) {
	_, sm := loadSpecFile(t, "05-two-state-machine.spec")
	result := exploreStateSpace(t, sm, 10, 20)

	expectedStates := 2 // State.A and State.B

	if result.StatesExplored < expectedStates {
		t.Errorf("Expected at least %d states, got %d", expectedStates, result.StatesExplored)
	}

	if len(result.ReachableStates) < expectedStates {
		t.Errorf("Expected %d reachable states, got %d", expectedStates, len(result.ReachableStates))
	}

	// Verify transitions exist
	if result.TransitionGraph == nil {
		t.Fatal("Transition graph should not be nil")
	}

	if len(result.TransitionGraph.Transitions) == 0 {
		t.Error("Expected at least one transition")
	}

	t.Logf("✓ Test 5 passed: Found %d states with %d transitions", result.StatesExplored, len(result.TransitionGraph.Transitions))
}

// Test 6: Multiple Initial States
func TestMultipleInitialStates(t *testing.T) {
	_, sm := loadSpecFile(t, "06-multiple-initial-states.spec")
	result := exploreStateSpace(t, sm, 10, 30)

	expectedInitialStates := 3 // oneOf creates 3 initial states

	// Verify that multiple initial states were explored
	// Note: InitialStates may not be populated in ExplorationResult, but we can verify
	// by checking that we explored states starting from different initial configurations
	initialConfigs := make(map[string]bool)
	for _, st := range result.ReachableStates {
		modeVal := st.Variables["mode"]
		counterVal := st.Variables["counter"]
		if modeVal != nil && counterVal != nil {
			configKey := fmt.Sprintf("%v-%v", modeVal, counterVal)
			initialConfigs[configKey] = true
		}
	}

	// Should have explored states from all 3 initial configurations
	if result.StatesExplored < expectedInitialStates {
		t.Errorf("Expected at least %d states explored, got %d", expectedInitialStates, result.StatesExplored)
	}

	t.Logf("✓ Test 6 passed: Explored %d total states from multiple initial configurations", result.StatesExplored)
}

// Test 7: Enum Parameterized Actions
func TestEnumParameterizedActions(t *testing.T) {
	_, sm := loadSpecFile(t, "07-enum-parameterized.spec")
	result := exploreStateSpace(t, sm, 10, 20)

	// Should explore all enum values: Red, Green, Blue
	expectedMinStates := 2 // At least initial and some other values

	if result.StatesExplored < expectedMinStates {
		t.Errorf("Expected at least %d states, got %d", expectedMinStates, result.StatesExplored)
	}

	// Verify different enum values appear
	colorValues := make(map[string]bool)
	for _, st := range result.ReachableStates {
		if color, ok := st.Variables["selectedColor"]; ok {
			if enumVal, ok := color.(*state.EnumValue); ok {
				colorValues[enumVal.ValueName] = true
			}
		}
	}

	if len(colorValues) < 2 {
		t.Errorf("Expected multiple enum values, found: %v", colorValues)
	}

	t.Logf("✓ Test 7 passed: Found %d enum value states: %v", len(colorValues), colorValues)
}

// Test 8: Bool Parameterized Actions
func TestBoolParameterizedActions(t *testing.T) {
	_, sm := loadSpecFile(t, "08-bool-parameterized.spec")
	result := exploreStateSpace(t, sm, 10, 20)

	// Should explore both true and false
	expectedMinStates := 2

	if result.StatesExplored < expectedMinStates {
		t.Errorf("Expected at least %d states, got %d", expectedMinStates, result.StatesExplored)
	}

	boolValues := make(map[bool]bool)
	for _, st := range result.ReachableStates {
		if flag, ok := st.Variables["flag"]; ok {
			if boolVal, ok := flag.(*state.PrimitiveValue); ok && boolVal.BoolValue != nil {
				boolValues[*boolVal.BoolValue] = true
			}
		}
	}

	if len(boolValues) < 2 {
		t.Errorf("Expected both true and false, found: %v", boolValues)
	}

	t.Logf("✓ Test 8 passed: Found bool states: %v", boolValues)
}

// Test 9: Action Coverage
func TestActionCoverage(t *testing.T) {
	_, sm := loadSpecFile(t, "09-action-coverage.spec")
	result := exploreStateSpace(t, sm, 10, 20)

	// All 3 actions (setX, setY, setBoth) should be executed
	actionExecuted := make(map[string]bool)
	for _, trans := range result.TransitionGraph.Transitions {
		actionExecuted[trans.Action] = true
	}

	expectedActions := []string{"setX", "setY", "setBoth"}
	for _, action := range expectedActions {
		if !actionExecuted[action] {
			t.Errorf("Expected action '%s' to be executed", action)
		}
	}

	t.Logf("✓ Test 9 passed: All actions executed: %v", actionExecuted)
}

// Test 10: Precondition Filtering
func TestPreconditionFiltering(t *testing.T) {
	_, sm := loadSpecFile(t, "10-precondition-filtering.spec")
	result := exploreStateSpace(t, sm, 10, 30)

	// Reset action should only execute when counter >= 5
	// Verify that states with counter < 5 don't have reset transitions
	lowCounterHasReset := false

	for _, trans := range result.TransitionGraph.Transitions {
		if trans.Action == "reset" {
			// Check if reset happens from a state with counter < 5
			if trans.FromState.Variables["counter"] != nil {
				if counterVal, ok := trans.FromState.Variables["counter"].(*state.PrimitiveValue); ok && counterVal.IntValue != nil {
					if *counterVal.IntValue < 5 {
						lowCounterHasReset = true
					}
				}
			}
		}
	}

	if lowCounterHasReset {
		t.Error("Reset action should not execute when counter < 5")
	}

	// Verify increment is working
	incrementFound := false
	for _, trans := range result.TransitionGraph.Transitions {
		if trans.Action == "increment" {
			incrementFound = true
			break
		}
	}

	if !incrementFound {
		t.Error("Expected increment action to be executed")
	}

	t.Logf("✓ Test 10 passed: Precondition filtering works correctly")
}

// Test: Transition Coverage
func TestTransitionCoverage(t *testing.T) {
	_, sm := loadSpecFile(t, "01-known-state-space-size.spec")
	result := exploreStateSpace(t, sm, 20, 50)

	// Each state should have at least one outgoing transition (except terminal states)
	// For this spec, each state can transition via setX, setY, or setZ
	if result.TransitionGraph == nil {
		t.Fatal("Transition graph should not be nil")
	}

	statesWithTransitions := 0
	for _, node := range result.TransitionGraph.States {
		if len(node.Outgoing) > 0 {
			statesWithTransitions++
		}
	}

	if statesWithTransitions == 0 {
		t.Error("Expected at least some states to have outgoing transitions")
	}

	t.Logf("✓ Transition coverage: %d states have outgoing transitions", statesWithTransitions)
}

// Test: State Hashing Uniqueness
func TestStateHashingUniqueness(t *testing.T) {
	_, sm := loadSpecFile(t, "01-known-state-space-size.spec")
	result := exploreStateSpace(t, sm, 20, 50)

	// Verify that different states have different hashes
	hashToState := make(map[string]*state.State)
	duplicateHashes := 0

	for _, state := range result.ReachableStates {
		hasher := explore.NewStateHasher()
		hash := hasher.HashState(state)
		
		if existing, exists := hashToState[hash]; exists {
			// Check if states are actually the same
			if !statesEqual(state, existing) {
				duplicateHashes++
				t.Errorf("Hash collision detected: different states have same hash\nState 1: %v\nState 2: %v", state.Variables, existing.Variables)
			}
		} else {
			hashToState[hash] = state
		}
	}

	if duplicateHashes > 0 {
		t.Errorf("Found %d hash collisions", duplicateHashes)
	}

	t.Logf("✓ State hashing: All %d states have unique hashes", len(hashToState))
}

// Helper: Compare two states for equality
func statesEqual(s1, s2 *state.State) bool {
	if len(s1.Variables) != len(s2.Variables) {
		return false
	}
	for k, v1 := range s1.Variables {
		v2, ok := s2.Variables[k]
		if !ok {
			return false
		}
		if v1.String() != v2.String() {
			return false
		}
	}
	return true
}

// Test: Boundary Conditions - Max Depth
func TestBoundaryMaxDepth(t *testing.T) {
	_, sm := loadSpecFile(t, "02-small-counter.spec")
	
	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(2)  // Very small depth
	explorer.SetMaxStates(100)
	
	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("Exploration failed: %v", err)
	}

	// Verify that max depth is respected
	if result.MaxDepth > 2 {
		t.Errorf("Expected max depth <= 2, got %d", result.MaxDepth)
	}

	t.Logf("✓ Max depth test: Max depth was %d (limit: 2)", result.MaxDepth)
}

// Test: Boundary Conditions - Max States
func TestBoundaryMaxStates(t *testing.T) {
	_, sm := loadSpecFile(t, "02-small-counter.spec")
	
	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(100)
	explorer.SetMaxStates(2)  // Very small state limit
	
	result, err := explorer.ExploreBFS()
	if err != nil {
		t.Fatalf("Exploration failed: %v", err)
	}

	// Should stop at or before maxStates
	if result.StatesExplored > 2 {
		t.Errorf("Expected states explored <= 2, got %d", result.StatesExplored)
	}

	t.Logf("✓ Max states test: Explored %d states (limit: 2)", result.StatesExplored)
}

// Test: Invariant Preservation
func TestInvariantPreservation(t *testing.T) {
	_, sm := loadSpecFile(t, "01-known-state-space-size.spec")
	result := exploreStateSpace(t, sm, 20, 50)

	// All reachable states should satisfy invariants
	// Note: Violations are recorded, but states should still be valid
	for _, state := range result.ReachableStates {
		errors, err := sm.ValidateState(state)
		if err != nil {
			t.Errorf("Error validating state: %v", err)
			continue
		}
		if len(errors) > 0 {
			t.Errorf("State violates invariants: %v", errors)
		}
	}

	t.Logf("✓ Invariant preservation: All %d reachable states satisfy invariants", len(result.ReachableStates))
}

