// Test 2: Small Counter with Known States
// Counter can be 0, 1, 2, 3 (4 states total)
// Expected: 4 states

var counter: int

init {
  counter = 0
}

action increment {
  require counter < 3
  counter' = counter + 1
}

action decrement {
  require counter > 0
  counter' = counter - 1
}

invariant bounded {
  counter >= 0 && counter <= 3
}

