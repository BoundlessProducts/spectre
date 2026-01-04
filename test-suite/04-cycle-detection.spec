// Test 4: Cycle Detection
// Simple toggle that creates a cycle: 0 -> 1 -> 0 -> ...
// Expected: Cycle detected, 2 states

var toggle: int

init {
  toggle = 0
}

action flip {
  toggle' = 1 - toggle
}

invariant validToggle {
  toggle = 0 || toggle = 1
}

