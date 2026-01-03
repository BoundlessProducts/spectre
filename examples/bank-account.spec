// Bank Account Example with Descriptions
// Demonstrates maps and financial invariants
// NOTE: Simplified due to current parser limitations with record literals

type Account = {
  id: int,
  balance: int,
  frozen: bool
}

description "Map of account IDs to account records"
var accounts: Map<int, Account>

description "Counter for tracking number of transactions"
var transactionCount: int

description "System starts with no accounts and zero transactions"
init {
  accounts = Map.empty()
  transactionCount = 0
}

description "Increments transaction count"
action recordTransaction {
  transactionCount' = transactionCount + 1
}

description "CRITICAL: Ensures transaction count stays non-negative"
invariant transactionCountNonNegative {
  transactionCount >= 0
}

description "Verifies that transactions will eventually occur"
temporal eventuallyTransaction {
  eventually transactionCount = 1
}
