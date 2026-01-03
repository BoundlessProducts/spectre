// Bank Account Example - Converted from Quint
// Demonstrates maps, invariants, and non-deterministic actions
//
// STATE MACHINE DEFINITION:
// - State Variable: balances (Map<str, int>) - maps account names to balances
// - Initial State: init block - all balances start at 0
// - Transitions: deposit, withdraw, stepDeposit, stepWithdraw actions
//   The verifier automatically explores all possible actions from each state
//   (no explicit "step" action needed - this is implicit in Spectre)

module bank {
  description "A state variable to store the balance of each account"
  var balances: Map<str, int>

  description "Set of valid account addresses"
  const ADDRESSES: Set<str> = Set.of("alice").union(Set.of("bob")).union(Set.of("charlie"))

  description "At the initial state, all balances are zero"
  init {
    // Initialize map with all addresses having balance 0
    balances = Map.empty()
    balances = balances.put("alice", 0)
    balances = balances.put("bob", 0)
    balances = balances.put("charlie", 0)
  }

  description "Increment balance of account by amount"
  action deposit(account: str, amount: int) {
    balances' = balances.put(account, balances[account] + amount)
  }

  description "Decrement balance of account by amount if funds are sufficient"
  action withdraw(account: str, amount: int) {
    require balances[account] >= amount
    balances' = balances.put(account, balances[account] - amount)
  }

  description "Non-deterministic step: deposit to a random account with a random amount"
  // The verifier will explore all possible (account, amount) combinations
  // that satisfy the require conditions
  action stepDeposit(account: str, amount: int) {
    require ADDRESSES.contains(account)
    require amount >= 1 && amount <= 100
    deposit(account, amount)
  }

  description "Non-deterministic step: withdraw from a random account with a random amount"
  // The verifier will explore all possible (account, amount) combinations
  action stepWithdraw(account: str, amount: int) {
    require ADDRESSES.contains(account)
    require amount >= 1 && amount <= 100
    withdraw(account, amount)
  }

  description "Invariant: Account balances should never be negative"
  invariant no_negatives {
    ADDRESSES.forall(addr => balances[addr] >= 0)
  }
}

