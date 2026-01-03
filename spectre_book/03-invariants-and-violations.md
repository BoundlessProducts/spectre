# Chapter 3: Invariants and Violations

This chapter explores invariants in Spectre: what they are, how they can be violated, and how to fix violations. We'll use two examples to demonstrate common patterns of invariant violations and their solutions.

## Understanding Invariants

**Invariants** are properties that must hold in **every reachable state** of the system. Unlike temporal properties (which describe behavior over time), invariants are state properties that must be true at all times.

### Key Characteristics of Invariants

1. **Must hold in initial state**: The invariant must be true when the system starts
2. **Must hold after every action**: After any action executes, the invariant must still be true
3. **Checked for all reachable states**: The verifier explores all possible execution paths and checks the invariant in every state

### Common Types of Invariants

- **Bounds checking**: Values stay within acceptable ranges
- **Non-negativity**: Values never become negative
- **Consistency**: Related values maintain consistent relationships
- **Safety properties**: Dangerous states are never reached

---

## Example 1: Bank Account System

Let's start with a bank account system that tracks account balances.

### The Problem: Missing Preconditions

Here's a bank account specification with a critical flaw:

```spectre
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
```

### Why This Violates the Invariant

The `no_negatives` invariant states that both `aliceBalance` and `bobBalance` must always be non-negative. However, several actions can violate this:

1. **`withdrawAlice30`**: If Alice's balance is less than 30 (e.g., 20), this action would set `aliceBalance = -10`, violating the invariant.

2. **`withdrawAlice100`**: If Alice's balance is less than 100 (e.g., 50), this action would set `aliceBalance = -50`, violating the invariant.

3. **`transfer50ToBob`**: If Alice's balance is less than 50 (e.g., 30), this action would set `aliceBalance = -20`, violating the invariant.

### The Error

When you verify this specification, the verifier will **fail** and report multiple violations:

```bash
$ ./spectre verify examples/bank-account-violation.spec

Verification failed: 49 violation(s) found

Violation 1 (Invariant):
  Action 'withdrawAlice100' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]

Violation 2 (Invariant):
  Action 'transfer50ToBob' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]

Violation 3 (Invariant):
  Action 'withdrawAlice30' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]

Violation 4 (Invariant):
  Action 'withdrawAlice100' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]
  Path:
    1. depositAlice50

Violation 5 (Invariant):
  Action 'withdrawAlice30' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]
  Path:
    1. depositAlice50
    2. transfer50ToBob

...
```

### Understanding the Error Messages

Each violation message has several components:

1. **Violation Number**: The order in which violations were discovered during exploration
2. **Action Name**: Which action would cause the violation (e.g., `withdrawAlice100`)
3. **Invariant Name**: Which invariant would be violated (e.g., `no_negatives`)
4. **Description**: The human-readable description from the invariant declaration
5. **Path (when applicable)**: The sequence of actions that leads to a state where this violation can occur

**Example Breakdown**:

```
Violation 4 (Invariant):
  Action 'withdrawAlice100' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]
  Path:
    1. depositAlice50
```

This tells us:
- **What action would fail**: `withdrawAlice100`
- **Why it fails**: Executing it would violate the `no_negatives` invariant
- **Where it fails**: After executing `depositAlice50`, Alice's balance is 50. At this point, trying to withdraw 100 would make the balance -50, violating the invariant.

### Why Verification Fails

The verifier detects that **actions are available** (they can be executed from certain states) but **would violate invariants** if executed. This indicates a specification flaw:

- Actions should have **preconditions** (`require` statements) that prevent them from being available when they would violate invariants
- The verifier catches these violations during exploration, even though it prevents invalid states from being reached
- Each violation represents a state-action pair where the action would create an invalid state

**Why this matters**: In a real system, we want preconditions to prevent invalid operations, not rely on post-execution checks. Preconditions make the specification clearer and ensure that actions are only available when they're safe to execute.

### The Fix: Add Preconditions

To fix this, we need to add `require` statements (preconditions) to ensure actions only execute when they won't violate the invariant:

```spectre
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

description "Withdraw 30 from Alice's account if funds are sufficient"
description "FIXED: Added precondition to check if funds are sufficient"
action withdrawAlice30 {
  require aliceBalance >= 30
  aliceBalance' = aliceBalance - 30
}

description "Withdraw 100 from Alice's account if funds are sufficient"
description "FIXED: Added precondition to prevent overdrawing"
action withdrawAlice100 {
  require aliceBalance >= 100
  aliceBalance' = aliceBalance - 100
}

description "Transfer 50 from Alice to Bob if Alice has sufficient funds"
description "FIXED: Added precondition to ensure sufficient funds"
action transfer50ToBob {
  require aliceBalance >= 50
  aliceBalance' = aliceBalance - 50
  bobBalance' = bobBalance + 50
}

description "Invariant: Account balances should never be negative"
invariant no_negatives {
  aliceBalance >= 0 && bobBalance >= 0
}
```

### How Preconditions Fix the Problem

**Preconditions** (`require` statements) are conditions that must be true **before** an action can execute. If a precondition is false, the action cannot execute.

By adding `require aliceBalance >= 30` to `withdrawAlice30`:
- The action can only execute when Alice has at least 30
- After execution, `aliceBalance` will be `aliceBalance - 30 >= 0`
- The invariant is preserved

Similarly:
- `require aliceBalance >= 100` ensures `withdrawAlice100` only executes when Alice has sufficient funds
- `require aliceBalance >= 50` ensures `transfer50ToBob` only executes when Alice can afford the transfer

With these preconditions, the invariant `no_negatives` will always hold.

---

## Example 2: Inventory Management System

Now let's look at a more complex example with **multiple invariants**.

### The Problem: Missing Bounds Checking

Here's an inventory system that tracks the number of items:

```spectre
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
```

### Why This Violates the Invariants

This system has **two invariants** that can both be violated:

1. **`withinCapacity`**: Violated when inventory exceeds 100
   - `add50` when `inventory = 60` → `inventory = 110` ❌
   - `add60` when `inventory = 50` → `inventory = 110` ❌

2. **`nonNegative`**: Violated when inventory becomes negative
   - `remove30` when `inventory = 20` → `inventory = -10` ❌
   - `remove50` when `inventory = 30` → `inventory = -20` ❌

### The Error

When you verify this specification, the verifier will **fail** and report violations:

```bash
$ ./spectre verify examples/inventory-violation.spec

Verification failed: X violation(s) found

Violation 1 (Invariant):
  Action 'add60' would violate invariants: [invariant withinCapacity violated: (Ensures inventory never exceeds maximum capacity)]
  
Violation 2 (Invariant):
  Action 'remove30' would violate invariants: [invariant nonNegative violated: (CRITICAL: Ensures inventory never becomes negative)]
  
...
```

### Understanding the Error Messages

The violations follow the same pattern as the bank account example:

- **Violation 1**: `add60` would exceed the capacity of 100 when inventory is already above 40
- **Violation 2**: `remove30` would make inventory negative when the current value is less than 30

Each violation indicates that an action is available but would create an invalid state. The actions need preconditions to prevent them from being available in these situations.

### The Fix: Add Preconditions for Both Invariants

To fix this, we need preconditions that ensure **both** invariants are maintained:

```spectre
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
```

### How These Preconditions Work

1. **For `add50` and `add60`**: 
   - `require inventory + quantity <= 100` ensures that after adding, inventory won't exceed 100
   - This preserves the `withinCapacity` invariant

2. **For `remove30` and `remove50`**:
   - `require inventory >= quantity` ensures that after removing, inventory won't become negative
   - This preserves the `nonNegative` invariant

With these preconditions, both invariants will always hold.

---

## Common Patterns of Invariant Violations

Based on our examples, here are common patterns:

### Pattern 1: Missing Preconditions for Decrements

**Problem**: Actions that decrease values don't check if the value is large enough.

**Example**: `withdraw` without checking balance, `remove` without checking inventory

**Fix**: Add `require value >= amount` before decrementing

### Pattern 2: Missing Preconditions for Increments

**Problem**: Actions that increase values don't check if the value would exceed bounds.

**Example**: `add` without checking capacity, `deposit` without checking maximum

**Fix**: Add `require value + amount <= maxValue` before incrementing

### Pattern 3: Missing Preconditions for Compound Operations

**Problem**: Actions that perform multiple state changes don't validate intermediate states.

**Example**: `transfer` that subtracts from one account and adds to another

**Fix**: Add preconditions for each state change: `require fromValue >= amount` and ensure `toValue + amount <= maxValue` if applicable

### Pattern 4: Missing Validation of Input Parameters

**Problem**: Actions don't validate that parameters are within acceptable ranges.

**Example**: Accepting negative amounts, zero quantities, or invalid account names

**Fix**: Add `require amount > 0`, `require account != ""`, etc.

---

## Best Practices for Preventing Invariant Violations

1. **Always add preconditions**: Every action that modifies state should have preconditions that ensure invariants are maintained.

2. **Check before modifying**: Preconditions should check the **current state** before making changes, not after.

3. **Consider all invariants**: If you have multiple invariants, ensure each action maintains **all** of them.

4. **Use descriptive error messages**: When preconditions fail, the verifier will report which action couldn't execute and why.

5. **Test edge cases**: Consider boundary conditions:
   - What happens at the minimum value (0, empty, etc.)?
   - What happens at the maximum value (100, full, etc.)?
   - What happens with invalid inputs?

---

## Understanding Verifier Behavior

When you verify a specification with missing preconditions:

1. **The verifier checks invariants**: It validates invariants after each action execution
2. **Invalid states are prevented**: If an action would create an invalid state, that action execution fails
3. **Violations are recorded**: Each attempt to execute an action that would violate invariants is recorded as a violation
4. **Verification fails**: If any violations are found, verification fails with a detailed report

### Example: What the Verifier Does

When an action would violate an invariant:
- The verifier detects it before adding the invalid state
- The action execution fails, and a **violation is recorded**
- The violation includes:
  - The action that would fail
  - The invariant that would be violated
  - The path (sequence of actions) leading to the state where this would occur
- Other valid paths continue to be explored
- **Verification fails** if any violations are found

**Key Insight**: The verifier acts as both a safety net (preventing invalid states) and a **specification checker** (reporting when actions lack proper preconditions). This helps you find and fix specification flaws early.

---

## Summary

This chapter covered:

1. **What invariants are**: Properties that must hold in every reachable state

2. **How invariants are violated**: 
   - Missing preconditions allow actions to execute when they shouldn't
   - Actions modify state in ways that break invariant properties

3. **How to fix violations**:
   - Add `require` statements (preconditions) to actions
   - Preconditions check the current state before making changes
   - Preconditions ensure invariants will hold after the action executes

4. **Common patterns**:
   - Missing checks for decrements (negative values)
   - Missing checks for increments (exceeding bounds)
   - Missing checks for compound operations
   - Missing validation of input parameters

5. **Verifier behavior**: The verifier prevents invalid states from being reached, but you should still add preconditions for correctness

6. **Best practices**: Always add preconditions, check before modifying, consider all invariants

By understanding invariants and how to prevent violations, you can write specifications that maintain safety properties throughout system execution. Use preconditions to prevent problems, and invariants to guarantee properties.

