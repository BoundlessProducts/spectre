// Test 5: Two State Machine
// Exactly 2 states: A and B
// Expected: 2 states, transitions between them

enum State {
  A,
  B
}

var state: State

init {
  state = State.A
}

action switchToB {
  require state = State.A
  state' = State.B
}

action switchToA {
  require state = State.B
  state' = State.A
}

