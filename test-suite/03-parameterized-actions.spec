// Test 3: Parameterized Actions
// Action takes int parameter from 0-2
// Expected: Multiple states based on parameter combinations

var value: int

init {
  value = 0
}

action setValue(v: int) {
  require v >= 0 && v <= 2
  value' = v
}

invariant validValue {
  value >= 0 && value <= 2
}

