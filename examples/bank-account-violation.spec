// Bank Account Example with Invariant Violation
// Demonstrates how missing preconditions can violate invariants

description "Balance of Alice's account"
var aliceBalance: int

description "Balance of Bob's account"
var bobBalance: int

description "At the initial state, all balances are zero"
init {
  aliceBalance = 0
  bobBalance = 0
}

description "Deposit 50 to Alice's account"
action depositAlice50 {
  aliceBalance' = aliceBalance + 50
}

description "Withdraw 30 from Alice's account"
description "PROBLEM: Missing precondition to check if funds are sufficient"
action withdrawAlice30 {
  // Missing: require aliceBalance >= 30
  aliceBalance' = aliceBalance - 30
}

description "Withdraw 100 from Alice's account"
description "PROBLEM: Missing precondition - can withdraw more than balance"
action withdrawAlice100 {
  // Missing: require aliceBalance >= 100
  aliceBalance' = aliceBalance - 100
}

description "Transfer 50 from Alice to Bob"
description "PROBLEM: Missing precondition to ensure sufficient funds"
action transfer50ToBob {
  // Missing: require aliceBalance >= 50
  aliceBalance' = aliceBalance - 50
  bobBalance' = bobBalance + 50
}

description "Invariant: Account balances should never be negative"
invariant no_negatives {
  aliceBalance >= 0 && bobBalance >= 0
}
