// Test 9: Action Coverage
// Multiple actions, all should execute at least once
// Expected: All 3 actions should be executed

var x: int
var y: int

init {
  x = 0
  y = 0
}

action setX {
  x' = 1
}

action setY {
  y' = 1
}

action setBoth {
  x' = 1
  y' = 1
}

invariant bounded {
  (x = 0 || x = 1) && (y = 0 || y = 1)
}

