// Test 6: Multiple Initial States
// Uses oneOf to create 3 initial states
// Expected: At least 3 initial states explored

var mode: int
var counter: int

init oneOf {
  { mode = 0, counter = 0 },
  { mode = 1, counter = 10 },
  { mode = 2, counter = 20 }
}

action increment {
  counter' = counter + 1
}

invariant validMode {
  mode >= 0 && mode <= 2
}

