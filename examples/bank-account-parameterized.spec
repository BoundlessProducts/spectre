// Bank Account — parameterized actions
// Used in Chapter 10 (Model-Based Testing) and Chapter 11 (Runtime Monitoring).
//
// This spec uses parameterized actions (deposit/withdraw with an amount argument)
// rather than hard-coded amounts, making it better suited for driver generation
// and runtime monitoring.  Mine it with:
//
//   spectre mine --lang rust examples/rust/bank_account.rs
//
// Verify it:
//   spectre verify examples/bank-account-parameterized.spec
//
// Generate an ITF trace for model-based testing:
//   spectre verify examples/bank-account-parameterized.spec --emit-traces trace.itf.json
//
// Generate a Rust test driver:
//   spectre generate-driver --lang rust examples/bank-account-parameterized.spec
//
// Generate an embedded runtime monitor:
//   spectre generate-monitor --lang rust examples/bank-account-parameterized.spec

description "Current account balance"
var balance: int

description "Whether the account has been frozen by the owner"
var frozen: bool

description "Accounts start with zero balance and unfrozen"
init {
  balance = 0
  frozen = false
}

description "Deposit a positive amount into the account"
action deposit(amount: int) {
  require amount > 0
  require balance + amount <= 1000000
  balance' = balance + amount
}

description "Withdraw amount from the account; requires sufficient unfrozen balance"
action withdraw(amount: int) {
  require amount > 0
  require balance >= amount
  require !frozen
  balance' = balance - amount
}

description "Freeze the account, blocking all withdrawals"
action freeze {
  require !frozen
  frozen' = true
}

description "Unfreeze the account, re-enabling withdrawals"
action unfreeze {
  require frozen
  frozen' = false
}

description "Balance must never go negative"
invariant non_negative {
  balance >= 0
}

description "Balance must not exceed the deposit cap"
invariant within_cap {
  balance <= 1000000
}

description "The account must eventually be usable (not permanently frozen)"
temporal eventually_usable {
  eventually !frozen
}
