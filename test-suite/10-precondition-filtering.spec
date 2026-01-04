// Test 10: Precondition Filtering
// Action with strict precondition limits reachable states
// Expected: Only valid states explored

var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action reset {
  require counter >= 5
  counter' = 0
}

invariant nonNegative {
  counter >= 0
}

