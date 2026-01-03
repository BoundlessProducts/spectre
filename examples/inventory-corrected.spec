// Inventory Management Example - Corrected
// Demonstrates how to fix invariant violations with proper preconditions

description "Number of items in inventory"
var inventory: int

description "System starts with empty inventory"
init {
  inventory = 0
}

description "Add 50 items to inventory if capacity allows"
description "FIXED: Added precondition to check capacity limit"
action add50 {
  require inventory + 50 <= 100
  inventory' = inventory + 50
}

description "Add 60 items to inventory if capacity allows"
description "FIXED: Added precondition to prevent exceeding capacity"
action add60 {
  require inventory + 60 <= 100
  inventory' = inventory + 60
}

description "Remove 30 items from inventory if enough items exist"
description "FIXED: Added precondition to check if enough items exist"
action remove30 {
  require inventory >= 30
  inventory' = inventory - 30
}

description "Remove 50 items from inventory if enough items exist"
description "FIXED: Added precondition to prevent negative inventory"
action remove50 {
  require inventory >= 50
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
