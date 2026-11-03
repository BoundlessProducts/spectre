// Benchmark: simplified two-cache MSI coherence protocol.
// Tests enum-heavy state spaces with multiple interacting state machines.
enum CacheState { Modified, Shared, Invalid }

var cache1: CacheState
var cache2: CacheState
var memory_clean: bool

init { cache1 = CacheState.Invalid  cache2 = CacheState.Invalid  memory_clean = true }

// Cache 1 acquires shared read
action c1_get_shared {
  require cache1 == CacheState.Invalid && cache2 != CacheState.Modified
  cache1' = CacheState.Shared
  cache2' = cache2
  memory_clean' = memory_clean
}

// Cache 1 acquires exclusive write (invalidates cache2)
action c1_get_modified {
  require cache1 != CacheState.Modified && cache2 != CacheState.Modified
  cache1' = CacheState.Modified
  cache2' = CacheState.Invalid
  memory_clean' = false
}

// Cache 1 writes back
action c1_writeback {
  require cache1 == CacheState.Modified
  cache1' = CacheState.Shared
  cache2' = cache2
  memory_clean' = true
}

// Cache 1 evicts
action c1_evict {
  require cache1 != CacheState.Invalid
  cache1' = CacheState.Invalid
  cache2' = cache2
  memory_clean' = cache1 != CacheState.Modified
}

// Cache 2 acquires shared read
action c2_get_shared {
  require cache2 == CacheState.Invalid && cache1 != CacheState.Modified
  cache1' = cache1
  cache2' = CacheState.Shared
  memory_clean' = memory_clean
}

// Cache 2 acquires exclusive write
action c2_get_modified {
  require cache2 != CacheState.Modified && cache1 != CacheState.Modified
  cache1' = CacheState.Invalid
  cache2' = CacheState.Modified
  memory_clean' = false
}

// Cache 2 evicts
action c2_evict {
  require cache2 != CacheState.Invalid
  cache1' = cache1
  cache2' = CacheState.Invalid
  memory_clean' = cache2 != CacheState.Modified
}

invariant single_writer {
  !(cache1 == CacheState.Modified && cache2 == CacheState.Modified)
}
invariant no_stale_shared {
  !(cache1 == CacheState.Modified && cache2 == CacheState.Shared)
  && !(cache2 == CacheState.Modified && cache1 == CacheState.Shared)
}
