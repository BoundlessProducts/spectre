# State Space Exploration Test Suite

This directory contains a comprehensive test suite for verifying that Spectre's state space exploration is complete and correct.

## Test Strategy

The test suite implements the comprehensive testing strategy outlined in the conversation, covering:

1. **Known State Space Size Tests** - Specifications with hand-countable states
2. **Parameterized Action Tests** - Verifying all argument combinations are explored
3. **Cycle Detection Tests** - Ensuring cycles are detected correctly
4. **Transition Coverage Tests** - Verifying all actions execute when enabled
5. **State Hashing Tests** - Ensuring unique state identification
6. **Boundary Condition Tests** - Max depth and max states limits
7. **Invariant Preservation Tests** - All reachable states satisfy invariants

## Running the Tests

From the project root:

```bash
# Run all tests
go test -v ./test-suite

# Run a specific test
go test -v ./test-suite -run TestKnownStateSpaceSize

# Run tests without verbose output
go test ./test-suite
```

## Test Specifications

### 01-known-state-space-size.spec
- **Purpose**: Verify exact state count (8 states: 2^3 combinations)
- **Expected**: 8 states explored

### 02-small-counter.spec
- **Purpose**: Verify bounded counter exploration (0-3)
- **Expected**: 4 states, all counter values appear

### 03-parameterized-actions.spec
- **Purpose**: Verify parameterized actions are explored with all argument values
- **Expected**: Multiple states with different parameter values

### 04-cycle-detection.spec
- **Purpose**: Verify cycles are detected
- **Expected**: At least one cycle detected

### 05-two-state-machine.spec
- **Purpose**: Verify simple state machine (2 states)
- **Expected**: Exactly 2 states, transitions between them

### 06-multiple-initial-states.spec
- **Purpose**: Verify `oneOf` initial states are all explored
- **Expected**: States from all 3 initial configurations explored

### 07-enum-parameterized.spec
- **Purpose**: Verify enum parameterized actions explore all enum values
- **Expected**: All enum values (Red, Green, Blue) appear in states

### 08-bool-parameterized.spec
- **Purpose**: Verify bool parameterized actions explore both true and false
- **Expected**: Both true and false values appear

### 09-action-coverage.spec
- **Purpose**: Verify all actions execute at least once
- **Expected**: All 3 actions (setX, setY, setBoth) appear in transitions

### 10-precondition-filtering.spec
- **Purpose**: Verify preconditions correctly filter invalid transitions
- **Expected**: Reset action only executes when counter >= 5

## Test Results Interpretation

### Passing Tests
- ✓ All expected states are explored
- ✓ Parameterized actions use all valid argument combinations
- ✓ Cycles are detected correctly
- ✓ All actions execute when enabled
- ✓ State hashes are unique
- ✓ Boundary conditions are respected
- ✓ All reachable states satisfy invariants

### Failing Tests
If a test fails, it indicates a potential issue with:
- **State space coverage**: Not all reachable states are being explored
- **Parameter generation**: Not all valid argument combinations are being tried
- **Action execution**: Actions that should execute are being skipped
- **State identification**: Hash collisions or incorrect state comparison
- **Cycle detection**: Cycles not being detected or incorrectly reported

## Adding New Tests

To add a new test:

1. Create a new `.spec` file with a descriptive name (e.g., `11-new-feature.spec`)
2. Add a corresponding test function in `state_space_test.go`:
   ```go
   func TestNewFeature(t *testing.T) {
       _, sm := loadSpecFile(t, "11-new-feature.spec")
       result := exploreStateSpace(t, sm, maxDepth, maxStates)
       
       // Add assertions
       if result.StatesExplored < expectedStates {
           t.Errorf("Expected at least %d states, got %d", expectedStates, result.StatesExplored)
       }
   }
   ```
3. Document the test's purpose and expected results in this README

## Metrics Tracked

The test suite verifies:
- **State Coverage**: Number of unique states explored
- **Transition Coverage**: Number of unique transitions explored
- **Action Coverage**: Percentage of actions that executed at least once
- **Parameter Coverage**: For each parameterized action, all valid argument combinations tried
- **Depth Distribution**: Maximum depth reached
- **Cycle Count**: Number of cycles detected

