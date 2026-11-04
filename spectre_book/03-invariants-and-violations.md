# Safety properties and invariants

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
const CAPACITY: int = 10

invariant not_full    { items <= CAPACITY }
invariant not_empty   { items >= 0 }
```

Use this for: buffers, queues, pools, counters.

### 2. State consistency

```spectre
enum Phase { Idle, Running, Done }
var phase:   Phase
var started: bool

invariant consistent { phase != Phase.Running || started }
```

"If the machine is running, it must have been started." Use `P != Q || R` or `!P || Q` for implication in invariants.

### 3. Mutual exclusion

```spectre
var lock_a: bool
var lock_b: bool

invariant mutual_exclusion { !(lock_a && lock_b) }
```

Use this for: locks, tokens, single-owner resources.

---

## A spec with a bug

```spectre
// bank-buggy.spec

var balance: int
const LIMIT: int = 1000

init { balance = 0 }

action deposit  { balance' = balance + 100 }
action withdraw { balance' = balance - 100 }   // <- no require!

invariant solvent    { balance >= 0 }
invariant within_cap { balance <= LIMIT }
```

```bash
spectre verify bank-buggy.spec
```

```
Violations:

1.   Invariant violated: solvent
  Counterexample trace (1 step(s)):
    [init]
    [step 1] withdraw: balance: 0 -> -100
    [VIOLATION] invariant "solvent" fails in this state
      balance = -100
```

---

## Reading a violation trace

```
Violations:

1.   Invariant violated: solvent     <- which invariant failed
  Counterexample trace (1 step(s)):
    [init]                            <- initial state
    [step 1] withdraw: balance: 0 -> -100   <- which action fired + what changed
    [VIOLATION] invariant "solvent" fails in this state
      balance = -100                  <- concrete value that broke the invariant
```

Every line in the trace shows:
- The **step number**
- The **action** that fired (or `init`)
- The **changed variable** and its before/after values
- The **violated invariant** name and the concrete state that broke it

---

## Longer traces

Here's a more interesting bug — withdraw only fails after a freeze/unfreeze cycle:

```spectre
// bank-freeze-bug.spec

enum Status { Active, Frozen }
var balance: int
var status:  Status
const LIMIT: int = 1000

init { balance = 0  status = Status.Active }

action deposit  {
  require status == Status.Active
  require balance + 100 <= LIMIT
  balance' = balance + 100
}

action withdraw {
  require status == Status.Active
  // missing: require balance >= 100
  balance' = balance - 100
}

action freeze   { status' = Status.Frozen  }
action unfreeze { status' = Status.Active  }

invariant solvent { balance >= 0 }
```

```bash
spectre verify bank-freeze-bug.spec
```

```
Violations:

1.   Invariant violated: solvent
  Counterexample trace (1 step(s)):
    [init]
    [step 1] withdraw: balance: 0 -> -100
    [VIOLATION] invariant "solvent" fails in this state
      balance = -100, status = Active
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
# Exhaustive verification: traversed 22 states (full reachable space)
```

---

## Checking multiple invariants

```spectre
invariant solvent     { balance >= 0 }
invariant within_cap  { balance <= LIMIT }
invariant frozen_safe { status != Status.Frozen || balance == 0 }
```

Spectre checks all invariants independently. If multiple are violated, each is reported with its own trace.

---

## Parameterized safety

```spectre
// producer-consumer.spec
const BUFFER_SIZE: int = 5

var items: int

init { items = 0 }

action produce { require items < BUFFER_SIZE  items' = items + 1 }
action consume { require items > 0            items' = items - 1 }

invariant buffer_bounded { items <= BUFFER_SIZE }
invariant no_underflow   { items >= 0 }
```

Verify at different sizes by overriding the constant at the command line:

```bash
spectre verify producer.spec --param BUFFER_SIZE=3
# Exhaustive verification: traversed 4 states (full reachable space)

spectre verify producer.spec --param BUFFER_SIZE=10
# Exhaustive verification: traversed 11 states (full reachable space)
```

Sweep the whole range:

```bash
spectre verify producer.spec --param-range BUFFER_SIZE=1..20
```

---

## State space size and `--max-states`

| Output line | What it means |
|------------|---------------|
| `traversed 6 states (full reachable space)` | Full proof — every reachable state checked |
| `traversed 5000 states (state or depth limit reached)` | Incomplete — violation may still exist beyond the limit |

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
- **Ordering**: must A happen before B? → `phase != Phase.B || startedA`
- **Exclusion**: can only one hold at a time? → `!(P && Q)`
- **Consistency**: do related variables agree? → `status != Status.Frozen || balance == 0`
- **Monotonicity**: can this only go up/down? → `count >= prev_count`

---

## Workflow summary

```
write invariants
      │
      v
spectre verify
      │
   violation?
   +--+--+
  yes    no
   │      +--> (ok) invariant holds in all reachable states  <- exhaustive result
   v
--suggest-repairs
   │
   v
SAFE candidates  --> apply · re-verify
   │
review-required  --> inspect manually first
```
