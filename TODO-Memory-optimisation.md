# Memory Optimization for State Space Exploration

## Overview

State space exploration can consume significant memory, especially for large or unbounded state spaces. This document outlines memory optimization strategies and verification approaches for the Spectre state space explorer.

## Memory Optimization Strategies

### 1. State Compression

**Problem**: Storing full state objects (with all variable values) consumes significant memory, especially for large states with many variables or complex data structures.

**Solution**: Compress states before storing them in the visited set.

**Approach**:
- Use a compact serialization format (e.g., binary encoding, delta compression)
- Store only variable deltas for similar states
- Use canonical representations (e.g., sort set/list elements before hashing)
- Compress state hashes (use shorter hash representations when collision probability is acceptable)

**Implementation Considerations**:
- Compression/decompression overhead must be balanced against memory savings
- Need to ensure hash collisions are handled correctly
- May need to store full states for violation reporting (can decompress on-demand)

**Memory Savings**: Potentially 50-80% reduction in memory for state storage, depending on state complexity.

### 2. Hash-Only Visited Set

**Problem**: The visited set stores full state objects, which can be large.

**Solution**: Store only state hashes in the visited set, not full state objects.

**Approach**:
- Use a hash set (e.g., `map[string]bool`) instead of storing full states
- Only store full states when needed (e.g., for violation traces, cycle detection)
- Use bloom filters for very large state spaces (with fallback to hash set for verification)

**Trade-offs**:
- **Pros**: Massive memory savings (hash strings are much smaller than full states)
- **Cons**: Cannot retrieve full state from visited set (must recompute if needed)
- **Cons**: Hash collisions could cause missed states (mitigate with collision detection)

**Memory Savings**: 90-95% reduction in visited set memory usage.

**Verification**: Need to ensure hash collisions don't cause incorrect results.

### 3. Selective Path Storage

**Problem**: Storing complete paths for all states (for violation traces) consumes significant memory.

**Solution**: Store paths only when necessary (e.g., for states that violate invariants or are part of cycles).

**Approach**:
- Don't store paths for all states during exploration
- Only store paths when:
  - A violation is detected
  - A state is part of a cycle (for temporal verification)
  - A state is explicitly requested (e.g., for debugging)
- Reconstruct paths on-demand when needed (using parent pointers or BFS/DFS backtracking)

**Implementation**:
- Store parent pointers in exploration nodes (already done)
- Reconstruct paths by following parent chain when violation detected
- Clear path information for states that are no longer needed

**Memory Savings**: 60-80% reduction in path storage memory, depending on exploration depth.

## Verification Strategy

To ensure memory optimizations don't introduce correctness issues, we need comprehensive verification:

### 1. State Compression Verification

**Test Cases**:
1. **Compression/Decompression Round-trip**: Verify that compressing and decompressing a state produces an identical state.
2. **Hash Consistency**: Verify that compressed and uncompressed states produce the same hash.
3. **Collision Detection**: Test with known collision cases to ensure they're handled correctly.
4. **State Comparison**: Verify that state comparison works correctly with compressed states.

**Test Approach**:
- Create a test suite with various state types (primitives, collections, records, nested structures)
- For each state type, test compression/decompression
- Verify hash computation is consistent
- Test edge cases (empty states, large states, states with special values)

### 2. Hash-Only Visited Set Verification

**Test Cases**:
1. **Correctness**: Verify that the same states are detected as visited with hash-only approach vs. full state storage.
2. **Hash Collision Handling**: Test with states that have hash collisions (if using shorter hashes).
3. **State Retrieval**: Verify that states can be correctly retrieved when needed (e.g., for violation traces).
4. **Exploration Completeness**: Ensure all states are still explored correctly.

**Test Approach**:
- Create a reference implementation using full state storage
- Create an optimized implementation using hash-only storage
- Run both on the same specs and compare:
  - Number of states explored
  - States visited
  - Violations found
  - Cycles detected
- Use specs with known state counts to verify completeness
- Test with specs that have hash collisions (if applicable)

**Verification Metrics**:
- States explored should be identical
- Violations found should be identical
- Cycles detected should be identical
- Memory usage should be significantly lower

### 3. Selective Path Storage Verification

**Test Cases**:
1. **Path Reconstruction**: Verify that paths can be correctly reconstructed from parent pointers.
2. **Violation Trace Accuracy**: Verify that violation traces are correct when paths are reconstructed.
3. **Cycle Path Accuracy**: Verify that cycle paths are correct for temporal verification.
4. **Path Completeness**: Ensure all necessary paths are available when needed.

**Test Approach**:
- Create test cases with known violation paths
- Verify that reconstructed paths match expected paths
- Test with various path lengths and complexities
- Verify that paths are available when violations are detected
- Test cycle detection with reconstructed paths

**Verification Metrics**:
- Violation traces should be identical to full path storage
- Cycle paths should be correct for temporal verification
- Memory usage should be lower than full path storage

## Implementation Phases

### Phase 1: Hash-Only Visited Set (Easiest, Highest Impact)

**Priority**: High
**Complexity**: Low
**Memory Savings**: 90-95%

**Steps**:
1. Modify `Explorer` to use `map[string]bool` for visited set instead of storing full states
2. Keep full states in exploration nodes (for current exploration)
3. Store full states only when violations detected
4. Add verification tests

### Phase 2: Selective Path Storage (Medium Complexity, Good Savings)

**Priority**: Medium
**Complexity**: Medium
**Memory Savings**: 60-80%

**Steps**:
1. Remove path storage from exploration nodes (keep parent pointers)
2. Implement path reconstruction function
3. Reconstruct paths only when violations detected
4. Update violation reporting to use reconstructed paths
5. Add verification tests

### Phase 3: State Compression (Most Complex, Moderate Savings)

**Priority**: Low
**Complexity**: High
**Memory Savings**: 50-80%

**Steps**:
1. Implement state compression/decompression
2. Integrate with visited set
3. Handle hash computation with compressed states
4. Add on-demand decompression for violation reporting
5. Add comprehensive verification tests

## Testing Strategy

### Unit Tests

1. **State Compression**:
   - Test compression/decompression for all state types
   - Test hash consistency
   - Test collision handling

2. **Hash-Only Visited Set**:
   - Test visited detection correctness
   - Test hash collision scenarios
   - Test state retrieval

3. **Path Reconstruction**:
   - Test path reconstruction from parent chain
   - Test path correctness for violations
   - Test path completeness

### Integration Tests

1. **End-to-End Verification**:
   - Run reference implementation (full storage)
   - Run optimized implementation
   - Compare results (states explored, violations, cycles)
   - Verify memory usage reduction

2. **Spec-Based Tests**:
   - Use existing test suite specs
   - Verify same results with optimizations
   - Measure memory usage improvements

### Performance Tests

1. **Memory Profiling**:
   - Profile memory usage before and after optimizations
   - Measure peak memory usage
   - Measure memory per state

2. **Speed Impact**:
   - Measure exploration speed with optimizations
   - Ensure optimizations don't significantly slow down exploration
   - Optimize hot paths if needed

## Success Criteria

1. **Correctness**: All optimizations must produce identical results to the reference implementation
2. **Memory Reduction**: Achieve at least 50% memory reduction overall
3. **Performance**: Exploration speed should not degrade by more than 10%
4. **Test Coverage**: 100% test coverage for optimization code paths

## Notes

- Start with Phase 1 (Hash-Only Visited Set) as it provides the highest memory savings with lowest complexity
- Each phase should be implemented and verified independently
- Keep the reference implementation available for comparison
- Document any trade-offs or limitations of each optimization
