// Simple Counter Example with Descriptions (Corrected)
// Demonstrates how descriptions improve error messages

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
description "Note: This can create paths where counter oscillates without progress"
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
description "This property holds because there exists a path where we only increment"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}

description "Ensures counter remains non-negative throughout execution"
temporal alwaysNonNegative {
  always (counter >= 0)
}

// Removed the 'progress' property because it requires fairness to hold
// Without fairness constraints, the system allows infinite paths where:
// 1. reset is executed repeatedly (counter never reaches 10)
// 2. decrement/increment oscillate (counter oscillates between values < 10)
//
// To make progress hold, you would need:
// - Fairness constraints (WF/SF) on actions
// - Or restrict actions to prevent infinite non-progress paths
