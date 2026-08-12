# Chapter 4: Temporal and Fairness Properties

This chapter covers temporal properties and fairness conditions.

## Understanding Temporal Properties

Temporal properties describe behavior over sequences of states (execution traces). Unlike invariants, which must hold in every state, temporal properties express requirements about future states and execution paths.

### A Simple Counter Example

Let's start with a basic counter specification:

```spectre
description "Tracks a numeric counter value"
var counter: int

description "System starts with counter initialized to zero"
init {
  counter = 0
}

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Decrements the counter by one, only when counter is positive"
action decrement {
  require counter > 0
  counter' = counter - 1
}

description "Resets the counter back to zero"
action reset {
  counter' = 0
}

description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}

description "Keeps counter within reasonable bounds"
invariant bounded {
  counter <= 100
}

description "Verifies that counter eventually reaches value 10"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}

description "Ensures counter remains non-negative throughout execution"
temporal alwaysNonNegative {
  always (counter >= 0)
}

description "Guarantees progress: if counter is below 10, it will eventually reach 10"
temporal progress {
  always (counter < 10 → eventually counter = 10)
}
```

This specification defines:
- A `counter` variable that starts at 0
- Three actions: `increment`, `decrement`, and `reset`
- Two invariants: `nonNegative` and `bounded`
- Three temporal properties

### Understanding the Temporal Properties

1. **`eventuallyReachesTen`**: Checks if the counter can eventually reach 10. This property holds because there exists a path where we only increment (0 → 1 → 2 → ... → 10).

2. **`alwaysNonNegative`**: Ensures the counter never becomes negative. This property holds because:
   - The counter starts at 0
   - `increment` always increases it
   - `decrement` has a precondition `counter > 0`, so it can't make it negative
   - `reset` sets it to 0

3. **`progress`**: This is more complex. It states: "In every state where `counter < 10`, eventually `counter = 10` will hold."

### The Problem: A Temporal Property Violation

When we verify this specification, we get a violation:

```
Verification failed: 1 violation(s) found

Violation 1 (Temporal Property: progress):
  Property 'progress' violated: (Guarantees progress: if counter is below 10, 
  it will eventually reach 10) P holds but Q never becomes true
```

**Why does this happen?**

The `progress` property states that whenever `counter < 10`, it will eventually reach 10. However, the `reset` action can be executed at any time, even when `counter = 0`. This creates a problematic execution path:

1. Start with `counter = 0` (where `counter < 10` holds)
2. Execute `reset` → `counter = 0` (still `counter < 10`)
3. Execute `reset` again → `counter = 0` (still `counter < 10`)
4. Continue executing `reset` infinitely...

This infinite loop creates a "blocking cycle" where the counter never reaches 10, violating the temporal property.

### Fixing the Problem: Restricting Actions

One way to fix this is to restrict the `reset` action so it can only execute when it doesn't prevent progress. In `counter-corrected.spec`, we add a precondition:

```spectre
description "Resets the counter back to zero"
action reset {
  require counter > 10
  counter' = 0
}
```

Now `reset` can only execute when `counter > 10`, which means:
- When `counter < 10`, `reset` cannot execute
- The system can only progress forward (via `increment`) or stay in place (via `decrement` if `counter > 0`)
- Since `increment` has no preconditions, it can always execute, ensuring progress to 10

With this fix, the `progress` temporal property holds.

### Why This Approach Works

By restricting `reset` to only execute when `counter > 10`, we've eliminated the problematic infinite loop. Now, when `counter < 10`:
- `reset` cannot execute (precondition not satisfied)
- `decrement` can only execute if `counter > 0`
- `increment` can always execute (no preconditions)

Since `increment` can always execute, there's always a path forward, ensuring the counter will eventually reach 10.

---

## Fairness Properties

However, restricting actions isn't always the best solution. Sometimes we want to allow all actions but assume they execute "fairly." This is where **fairness conditions** come in.

### What is Fairness?

In concurrent systems, fairness conditions specify assumptions about how actions are scheduled. They ensure that actions that are "ready" to execute will eventually do so, preventing infinite loops where an enabled action is never chosen.

### Weak Fairness (WF)

**Weak Fairness** states: "If an action is **continuously enabled** from some point onwards, it must eventually execute."

**Key points:**
- The action must be enabled continuously (without interruption)
- Once it becomes continuously enabled, it will eventually execute
- It doesn't guarantee execution if the action is enabled and disabled repeatedly

**Example:**
```spectre
description "Weak fairness ensures process1 gets fair access when continuously waiting"
temporal process1WeakFairness {
  WF(process1Request)
}
```

This means: If `process1Request` is enabled and stays enabled, it will eventually execute.

### Strong Fairness (SF)

**Strong Fairness** states: "If an action is enabled **infinitely often**, it must execute infinitely often."

**Key points:**
- The action needs to be enabled infinitely often (even if disabled sometimes)
- It will execute infinitely often if enabled infinitely often
- Stronger guarantee than weak fairness

**Example:**
```spectre
description "Strong fairness ensures process1 executes even if intermittently enabled"
temporal process1StrongFairness {
  SF(process1Request)
}
```

This means: Even if `process1Request` is enabled and disabled repeatedly, if it's enabled infinitely often, it will execute infinitely often.

### When to Use Weak vs Strong Fairness

- **Weak Fairness (WF)**: Use when an action should execute if it's continuously available
  - Example: A process waiting for a lock should eventually get it if the lock is continuously free

- **Strong Fairness (SF)**: Use when an action should execute even if it's intermittently enabled
  - Example: A process should eventually get the lock even if the lock is acquired and released repeatedly

---

## Fairness in Temporal Properties

Fairness conditions can be combined with temporal operators using the leads-to (`→`) operator. This allows us to express properties that hold under fairness assumptions.

### Syntax

```spectre
temporal propertyName {
  WF(actionName) → temporalExpression
}
```

or

```spectre
temporal propertyName {
  SF(actionName) → temporalExpression
}
```

### How It Works

When you write `WF(action) → always (P → eventually Q)`, the verifier:

1. **Filters the transition graph**: Removes all "unfair" paths where the action is continuously enabled (WF) or enabled infinitely often (SF) but never executes
2. **Verifies the property**: Checks if `always (P → eventually Q)` holds over the remaining "fair" paths

This ensures that the temporal property is verified only under the assumption that the action executes fairly.

---

## The Counter Example with Fairness

In `counter-with-fairness.spec`:

```spectre
description "Tracks a numeric counter value"
var counter: int

description "System starts with counter initialized to zero"
init {
  counter = 0
}

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Decrements the counter by one, only when counter is positive"
action decrement {
  require counter > 0
  counter' = counter - 1
}

description "Resets the counter back to zero"
action reset {
  counter' = 0
}

description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}

description "Keeps counter within reasonable bounds"
invariant bounded {
  counter <= 100
}

description "Verifies that counter can eventually reach value 10"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}

description "Ensures counter remains non-negative throughout execution"
temporal alwaysNonNegative {
  always (counter >= 0)
}

description "Guarantees progress: if counter is below 10, it will eventually reach 10"
description "With weak fairness on increment, this property holds because increment will"
description "eventually execute when continuously enabled, ensuring progress to counter = 10"
temporal progress {
  WF(increment) → always (counter < 10 → eventually counter == 10)
}
```

Notice that `reset` has no preconditions (unlike the corrected version). However, the `progress` property now includes weak fairness:

```spectre
WF(increment) → always (counter < 10 → eventually counter == 10)
```

### How This Works

1. **Weak fairness on `increment`**: `WF(increment)` assumes that if `increment` is continuously enabled, it will eventually execute.

2. **Filtering unfair paths**: The verifier identifies cycles where:
   - `increment` is continuously enabled (it has no preconditions, so it's always enabled)
   - `increment` never executes in the cycle
   - These cycles are marked as "unfair" and filtered out

3. **Example of an unfair cycle**: 
   - State: `counter = 0`
   - Transition: `reset` action → `counter = 0` (self-loop)
   - This cycle is unfair because `increment` is continuously enabled (no preconditions) but never executes

4. **Verifying the property**: After filtering out unfair cycles, the verifier checks if `always (counter < 10 → eventually counter == 10)` holds over the remaining fair paths.

5. **Result**: The property holds because:
   - Unfair cycles (like the `reset` self-loop at `counter = 0`) are filtered out
   - In fair paths, `increment` will eventually execute when `counter < 10`
   - This ensures progress toward `counter = 10`

### Why This is Better

Using fairness instead of restricting actions has several advantages:

1. **More general**: Doesn't require changing action preconditions
2. **Realistic assumptions**: Models real systems where actions execute fairly
3. **Clearer intent**: Explicitly states the fairness assumptions
4. **Reusable**: The same fairness conditions can be applied to multiple temporal properties

---

## Summary

This chapter covered:

1. **Temporal Properties**: Express requirements about future states and execution paths
   - `eventually P`: P holds at some point
   - `always P`: P holds in every state
   - `P → Q`: If P holds, eventually Q will hold

2. **Temporal Property Violations**: Can occur when infinite loops (blocking cycles) prevent progress

3. **Fixing Violations**: 
   - Option 1: Restrict actions with preconditions
   - Option 2: Use fairness conditions (preferred for concurrent systems)

4. **Fairness Conditions**:
   - **Weak Fairness (WF)**: Continuously enabled actions eventually execute
   - **Strong Fairness (SF)**: Actions enabled infinitely often execute infinitely often

5. **Fairness in Temporal Properties**: Use `WF(action) → temporalExpression` or `SF(action) → temporalExpression` to verify properties under fairness assumptions

6. **How Fairness Works**: The verifier filters out unfair paths (where fair actions don't execute) before verifying temporal properties

By understanding temporal and fairness properties, you can verify complex system behaviors and ensure that progress guarantees hold even in concurrent systems with multiple possible execution paths.

