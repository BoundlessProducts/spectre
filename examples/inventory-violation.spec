// Inventory Management Example with Invariant Violation
// Demonstrates how missing bounds checking can violate invariants

description "Number of items in inventory"
var inventory: int

description "System starts with empty inventory"
init {
  inventory = 0
}

description "Add 50 items to inventory"
description "PROBLEM: Missing precondition to check capacity limit"
action add50 {
  // Missing: require inventory + 50 <= 100
  inventory' = inventory + 50
}

description "Add 60 items to inventory"
description "PROBLEM: Can exceed capacity"
action add60 {
  // Missing: require inventory + 60 <= 100
  inventory' = inventory + 60
}

description "Remove 30 items from inventory"
description "PROBLEM: Missing precondition to check if enough items exist"
action remove30 {
  // Missing: require inventory >= 30
  inventory' = inventory - 30
}

description "Remove 50 items from inventory"
description "PROBLEM: Can make inventory negative"
action remove50 {
  // Missing: require inventory >= 50
  inventory' = inventory - 50
}

description "CRITICAL: Ensures inventory never exceeds maximum capacity"
invariant withinCapacity {
  inventory <= 100
}

description "CRITICAL: Ensures inventory never becomes negative"
invariant nonNegative {
  inventory >= 0
}
