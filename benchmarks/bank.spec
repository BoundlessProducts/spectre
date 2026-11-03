// Benchmark: bank account with balance invariant.
// ~100 states at limit 100.
param LIMIT: int
var balance: int
var total_deposits: int
var total_withdrawals: int
init { balance = 0  total_deposits = 0  total_withdrawals = 0 }
action deposit {
  require balance + 10 <= LIMIT
  balance' = balance + 10
  total_deposits' = total_deposits + 10
  total_withdrawals' = total_withdrawals
}
action withdraw {
  require balance >= 10
  balance' = balance - 10
  total_deposits' = total_deposits
  total_withdrawals' = total_withdrawals + 10
}
invariant non_negative { balance >= 0 }
invariant within_limit { balance <= LIMIT }
invariant accounting   { balance == total_deposits - total_withdrawals }
