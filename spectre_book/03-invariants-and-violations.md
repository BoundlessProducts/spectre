# Chapter 3: Safety properties and invariants

---

## What a safety property says

> "Nothing bad ever happens."

- An `invariant` is a boolean expression that must hold in **every reachable state**.
- A violation = one state where it doesn't hold.
- Spectre finds the **shortest path** to that state (BFS guarantee).

---

## Three patterns for invariants

### 1. Resource bound

```spectre
var items:  int
param CAPACITY: int = 10

invariant not_full    { items <= CAPACITY }
invariant not_empty   { items >= 0 }
```

Use this for: buffers, queues, pools, counters.

### 2. State consistency

```spectre
enum Phase { Idle, Running, Done }
var phase:   Phase
var started: bool

invariant consistent { phase = Running → started }
```

"If the machine is running, it must have been started." Use `→` (or write `not P or Q`) for implication.

### 3. Mutual exclusion

```spectre
var lock_a: bool
var lock_b: bool

invariant mutual_exclusion { not (lock_a and lock_b) }
```

Use this for: locks, tokens, single-owner resources.

---

## A spec with a bug

```spectre
// bank-buggy.spec

var balance: int
param LIMIT: int = 1000

init { balance = 0 }

action deposit  { balance' = balance + 100 }
action withdraw { balance' = balance - 100 }   // ← no require!

invariant solvent    { balance >= 0 }
invariant within_cap { balance <= LIMIT }
```

```bash
spectre verify bank-buggy.spec
```

```
✗ violation — invariant 'solvent' after action 'withdraw'

Counterexample (2 steps):
  step 0  init:     balance = 0
  step 1  withdraw: balance = -100  ← VIOLATION
                    solvent: balance >= 0  →  -100 >= 0  FAILS
```

---

## Reading a violation trace

```
✗ violation — invariant 'solvent' after action 'withdraw'
  ↑ which invariant failed    ↑ which action triggered it

Counterexample (2 steps):              ← shortest path to the bad state
  step 0  init:     balance = 0        ← initial state
  step 1  withdraw: balance = -100     ← state after the action

                    solvent: balance >= 0  →  -100 >= 0  FAILS
                    ↑ invariant name      ↑ expression   ↑ why it fails
```

Every line in the trace shows:
- The **step number**
- The **action** that fired (or `init`)
- The **full state** after the action (all variables)
- On the violation line: the invariant expression and what value made it fail

---

## Longer traces

Here's a more interesting bug — withdraw only fails after a freeze/unfreeze cycle:

```spectre
// bank-freeze-bug.spec

enum Status { Active, Frozen }
var balance: int
var status:  Status
param LIMIT: int = 1000

init { balance = 0  status = Active }

action deposit  {
  require status = Active
  require balance + 100 <= LIMIT
  balance' = balance + 100
}

action withdraw {
  require status = Active
  // missing: require balance >= 100
  balance' = balance - 100
}

action freeze   { status' = Frozen  }
action unfreeze { status' = Active  }

invariant solvent { balance >= 0 }
```

```bash
spectre verify bank-freeze-bug.spec
```

```
✗ violation — invariant 'solvent' after action 'withdraw'

Counterexample (2 steps):
  step 0  init:     balance = 0, status = Active
  step 1  withdraw: balance = -100, status = Active  ← VIOLATION
```

Spectre still finds the shortest path (2 steps), ignoring the freeze/unfreeze actions since they're not needed to trigger the bug.

---

## Using `--suggest-repairs`

```bash
spectre verify bank-freeze-bug.spec --suggest-repairs
```

```
Repair candidates (CANDIDATE — review before applying):

Violation 1: action 'withdraw' violates 'solvent'
  [1] CANDIDATE strengthen_require: add `require balance >= 100` to action withdraw
      Validation: original violation fixed · no deadlock · reachability unchanged — SAFE

  [2] CANDIDATE add_invariant_guard: add `when balance >= 100` to action withdraw
      Validation: original violation fixed · deadlock at balance=0 — review required
```

- **SAFE** — the patch fixes the violation and doesn't introduce new problems.
- **review required** — the patch fixes the violation but causes a deadlock or significantly restricts reachability. May be intentional, but check.
- Repairs are **always labelled CANDIDATE** — Spectre never applies them automatically.

Applying repair [1]: add `require balance >= 100` to `withdraw`. Re-verify:

```bash
spectre verify bank.spec
# ✓ safe — 12 states explored (exhaustive)
```

---

## Checking multiple invariants

```spectre
invariant solvent     { balance >= 0 }
invariant within_cap  { balance <= LIMIT }
invariant frozen_safe { status = Frozen → balance = 0 }
```

Spectre checks all invariants independently. If multiple are violated, each is reported with its own trace.

---

## Parameterized safety

```spectre
// producer-consumer.spec
param BUFFER_SIZE: int = 5

var produced: int
var consumed: int

init { produced = 0  consumed = 0 }

action produce {
  require produced - consumed < BUFFER_SIZE
  produced' = produced + 1
}

action consume {
  require consumed < produced
  consumed' = consumed + 1
}

invariant buffer_bounded { produced - consumed <= BUFFER_SIZE }
invariant no_underflow   { consumed <= produced }
```

Verify at different sizes:

```bash
spectre verify producer.spec --param BUFFER_SIZE=3
# ✓ safe — 10 states explored (exhaustive)

spectre verify producer.spec --param BUFFER_SIZE=10
# ✓ safe — 66 states explored (exhaustive)
```

Sweep the whole range:

```bash
spectre verify producer.spec --param-range BUFFER_SIZE=1..20
```

---

## State space size and `--max-states`

| Output line | What it means |
|------------|---------------|
| `6 states explored (exhaustive)` | Full proof — every reachable state checked |
| `5000 states explored (bounded — hit state limit)` | Incomplete — violation may still exist beyond the limit |

Use `--max-states` to control the limit:

```bash
spectre verify large.spec --max-states 50000
spectre verify large.spec --max-states -1   # unlimited (may not terminate)
```

Always check whether your result was **exhaustive**. A bounded result is not a proof.

---

## Invariant-writing checklist

For each resource or constraint in your system, ask:

- **Bounds**: is there a min or max? → `x >= 0`, `x <= LIMIT`
- **Ordering**: must A happen before B? → `phase = B → started_A`
- **Exclusion**: can only one hold at a time? → `not (P and Q)`
- **Consistency**: do related variables agree? → `frozen → balance = 0`
- **Monotonicity**: can this only go up/down? → `count >= prev_count`

---

## Workflow summary

```
write invariants
      │
      ▼
spectre verify
      │
   violation?
   ┌──┴──┐
  yes    no
   │      └──► ✓ invariant holds in all reachable states  ← exhaustive result
   ▼
--suggest-repairs
   │
   ▼
SAFE candidates  ──► apply · re-verify
   │
review-required  ──► inspect manually first
```
