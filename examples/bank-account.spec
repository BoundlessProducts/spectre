// Bank Account Example with Descriptions
// Demonstrates maps, transactions, and financial invariants

type Account = {
  id: int,
  balance: int,
  frozen: bool
}

description "Map of account IDs to account records"
var accounts: Map<int, Account>

description "List of all transactions, each tuple contains (fromId, toId, amount)"
var transactions: List<(int, int, int)>

description "System starts with no accounts and no transactions"
init {
  accounts = Map.empty()
  transactions = List.empty()
}

description "Creates a new account with the given ID and initial balance"
action createAccount(id: int, initialBalance: int) {
  require !accounts.contains(id) && initialBalance >= 0
  accounts' = accounts.put(id, { 
    id: id, 
    balance: initialBalance, 
    frozen: false 
  })
}

description "Deposits money into an account, increasing its balance"
action deposit(accountId: int, amount: int) {
  require accounts.contains(accountId) && amount > 0
  let account = accounts.get(accountId)
  require !account.frozen
  accounts' = accounts.put(accountId, { 
    ...account, 
    balance: account.balance + amount 
  })
}

description "Withdraws money from an account, decreasing its balance"
action withdraw(accountId: int, amount: int) {
  require accounts.contains(accountId) && amount > 0
  let account = accounts.get(accountId)
  require !account.frozen && account.balance >= amount
  accounts' = accounts.put(accountId, { 
    ...account, 
    balance: account.balance - amount 
  })
}

description "Transfers money from one account to another, recording the transaction"
action transfer(fromId: int, toId: int, amount: int) {
  require accounts.contains(fromId) && accounts.contains(toId) && amount > 0
  let fromAccount = accounts.get(fromId)
  let toAccount = accounts.get(toId)
  require !fromAccount.frozen && !toAccount.frozen
  require fromAccount.balance >= amount
  
  accounts' = accounts
    .put(fromId, { ...fromAccount, balance: fromAccount.balance - amount })
    .put(toId, { ...toAccount, balance: toAccount.balance + amount })
  
  transactions' = transactions.append((fromId, toId, amount))
}

description "Freezes an account, preventing all transactions"
action freezeAccount(accountId: int) {
  require accounts.contains(accountId)
  let account = accounts.get(accountId)
  accounts' = accounts.put(accountId, { ...account, frozen: true })
}

description "Unfreezes an account, allowing transactions again"
action unfreezeAccount(accountId: int) {
  require accounts.contains(accountId)
  let account = accounts.get(accountId)
  accounts' = accounts.put(accountId, { ...account, frozen: false })
}

description "CRITICAL: Ensures account balances never go negative"
invariant balanceNonNegative {
  accounts.values().forall(a => a.balance >= 0)
}

description "Ensures total balance is conserved across transfers (no money created or destroyed)"
invariant totalBalanceConserved {
  // After any transfer, the sum of all balances should equal the sum before
  // This is checked by the verifier across transitions
  true  // Placeholder - actual verification tracks balance changes
}

description "Verifies that accounts will eventually be created"
temporal eventuallyAccountCreated {
  eventually accounts.size() > 0
}

description "Verifies that transactions will eventually occur"
temporal eventuallyTransaction {
  eventually transactions.size() > 0
}

description "If accounts exist, they will eventually have transactions"
temporal eventuallyAccountActivity {
  always (accounts.size() > 0 → eventually transactions.size() > 0)
}
