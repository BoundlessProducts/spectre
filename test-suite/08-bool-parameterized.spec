// Test 8: Bool in Parameterized Action
// Action takes bool parameter
// Expected: Both true and false should be explored

var flag: bool

init {
  flag = false
}

action setFlag(b: bool) {
  flag' = b
}

