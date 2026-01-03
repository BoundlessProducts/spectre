// Counter example with fairness constraints
// This demonstrates Option 3 from SPECTRE_BOOK.md: Add Fairness Constraints

description "Tracks a numeric counter value"
var counter: int

description "System starts with counter initialized to zero"
init {
  counter = 0
}

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Decrements the counter by one, only when counter is positive"
action decrement {
  require counter > 0
  counter' = counter - 1
}

description "Resets the counter back to zero"
action reset {
  counter' = 0
}

description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}

description "Keeps counter within reasonable bounds"
invariant bounded {
  counter <= 100
}

description "Verifies that counter can eventually reach value 10"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}

description "Ensures counter remains non-negative throughout execution"
temporal alwaysNonNegative {
  always (counter >= 0)
}

description "Guarantees progress: if counter is below 10, it will eventually reach 10"
description "With weak fairness on increment, this property holds because increment will"
description "eventually execute when continuously enabled, ensuring progress to counter = 10"
temporal progress {
  WF(increment) → always (counter < 10 → eventually counter == 10)
}

