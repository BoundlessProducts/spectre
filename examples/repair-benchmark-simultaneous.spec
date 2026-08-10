// Repair benchmark: simultaneous assignment — withdraw atomically updates
// balance and records transaction, needs balance sufficiency guard.

var balance:   int
var txCount:   int

init {
  balance = 100
  txCount = 0
}

description "Withdraw: deduct 30 and increment transaction count simultaneously"
action withdraw {
  balance' = balance - 30
  txCount' = txCount + 1
}

description "Deposit: add 50"
action deposit {
  require balance + 50 <= 200
  balance' = balance + 50
}

description "Balance must never go negative"
invariant nonNegativeBalance {
  balance >= 0
}

description "Transaction count must stay non-negative"
invariant txCountNonNeg {
  txCount >= 0
}
