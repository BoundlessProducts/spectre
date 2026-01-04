// Test 1: Known State Space Size
// This spec has exactly 8 reachable states: all combinations of (x, y, z) where each is 0 or 1
// Expected: 2^3 = 8 states

var x: int  // 0 or 1
var y: int  // 0 or 1
var z: int  // 0 or 1

init {
  x = 0
  y = 0
  z = 0
}

action setX {
  x' = 1 - x
}

action setY {
  y' = 1 - y
}

action setZ {
  z' = 1 - z
}

invariant validValues {
  (x = 0 || x = 1) && (y = 0 || y = 1) && (z = 0 || z = 1)
}

