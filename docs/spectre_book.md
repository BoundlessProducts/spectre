# The Spectre Book

**Spectre** — formal verification for Rust programs.

---

## Table of contents

1. [Getting started](#chapter-1-getting-started)
2. [Language overview](#chapter-2-language-overview)
3. [Invariants and violations](#chapter-3-invariants-and-violations)
4. [Temporal and fairness properties](#chapter-4-temporal-and-fairness-properties)
5. [Fairness: WF and SF](#chapter-5-fairness-wf-and-sf)
6. [Concurrent systems](#chapter-6-concurrent-systems)
7. [Modules, parameters, and reuse](#chapter-7-modules-parameters-and-reuse)
8. [Reading counterexamples and debugging](#chapter-8-reading-counterexamples-and-debugging)
9. [Choosing a verification backend](#chapter-9-choosing-a-verification-backend)
10. [Mining specs from Rust](#chapter-10-mining-specs-from-rust)
11. [Drift detection and incremental re-verification](#chapter-11-drift-detection-and-incremental-re-verification)
12. [Model-based testing](#chapter-12-model-based-testing)
13. [Runtime monitoring](#chapter-13-runtime-monitoring)
14. [Benchmarking and performance](#chapter-14-benchmarking-and-performance)
15. [TLA+ and Apalache backend](#chapter-15-tla-and-apalache-backend)
16. [Z3 bounded model checking](#chapter-16-z3-bounded-model-checking)

---

# Chapter 1: Getting started

---

## Install

```bash
# macOS / Linux — build from source
git clone https://github.com/BoundlessProducts/spectre
cd spectre
go build -o spectre ./cmd/spectre
sudo mv spectre /usr/local/bin/

spectre --version
# spectre version 0.3.2
```

No runtime dependencies, no Docker, no JVM needed for the default BFS backend.

---

## Your first spec

Save this as `counter.spec`:

```spectre
var value: int
param MAX: int = 5

init { value = 0 }

action increment {
  require value < MAX
  value' = value + 1
}

action reset {
  value' = 0
}

invariant non_negative  { value >= 0 }
invariant within_bounds { value <= MAX }
```

Run it:

```bash
spectre verify counter.spec
```

```
✓ safe — 6 states explored (exhaustive)
  Invariants checked: non_negative, within_bounds
```

---

## What just happened

```
counter.spec
  ├── var value: int              ← one piece of state
  ├── param MAX: int = 5          ← a named constant
  ├── init { value = 0 }          ← where to start
  ├── action increment { … }      ← one possible transition
  ├── action reset { … }          ← another transition
  └── invariant within_bounds … ← the property to check
          │
          ▼
    spectre verify
          │
    BFS explores all 6 reachable states
    {0, 1, 2, 3, 4, 5}
          │
    checks both invariants in every state
          │
          ▼
        ✓ safe
```

Spectre uses **breadth-first search** — it finds every state reachable from `init` and checks your invariants in all of them. Six states here because `value` can be `0`–`5` and `reset` can always bring it back.

---

## Introducing a bug

Remove the `require` from `increment`:

```spectre
action increment {
  // require value < MAX   ← commented out
  value' = value + 1
}
```

```bash
spectre verify counter.spec
```

```
✗ violation — invariant 'within_bounds' after action 'increment'

Counterexample (6 steps):
  step 0  init:       value = 0
  step 1  increment:  value = 1
  step 2  increment:  value = 2
  step 3  increment:  value = 3
  step 4  increment:  value = 4
  step 5  increment:  value = 5
  step 6  increment:  value = 6  ← VIOLATION
                      within_bounds: value <= MAX → 6 <= 5  FAILS
```

Spectre found the shortest path to a bad state: six increments from zero.

---

## Fix it

Restore `require value < MAX`. This tells Spectre: "increment can only fire when `value < MAX`." If the condition is false the action is disabled — not an error, the system just can't take that step from that state.

---

## Key commands

| Command | What it does |
|---------|-------------|
| `spectre verify <file>` | BFS — exhaustively checks all reachable states |
| `spectre verify <file> --backend z3` | Bounded check — fast, finds shallow bugs |
| `spectre verify <file> --backend tlc` | Delegates to TLA+ Tools (validated fairness) |
| `spectre verify <file> --backend apalache` | Delegates to Apalache (symbolic, large specs) |
| `spectre mine --lang rust <file.rs>` | Extract a spec skeleton from Rust source |
| `spectre simulate <file> --steps 20` | Random walk through the state space |
| `spectre benchmark` | Time BFS across a set of spec files |

---

## Choosing a backend

```
      Your goal
          │
          ▼
  ┌───────────────────────────────────────────────────────┐
  │ Find shallow bugs quickly (≤20 steps)?                │
  │   → --backend z3 --bmc-depth 20                       │
  ├───────────────────────────────────────────────────────┤
  │ Exhaustive proof on a small–medium spec?              │
  │   → spectre verify  (default BFS)                     │
  ├───────────────────────────────────────────────────────┤
  │ Liveness / fairness properties you can trust?         │
  │   → --backend tlc  or  --backend apalache             │
  ├───────────────────────────────────────────────────────┤
  │ Spec too large for BFS (millions of states)?          │
  │   → --backend apalache  (symbolic, handles large)     │
  └───────────────────────────────────────────────────────┘
```

Chapter 9 covers the full decision guide.

---

## The full Spectre workflow

```
Rust source
    │
    ▼
spectre mine --lang rust src/account.rs
    │
    ▼
account.spec  ← skeleton; add invariants by hand
    │
    ▼
spectre verify account.spec
    │
    ├── ✓ safe ────────────────► commit .spec alongside source
    │
    └── ✗ violation
              │
              ▼
         spectre verify account.spec --suggest-repairs
              │
              ▼
         CANDIDATE patches  ← review, apply, re-verify
```

Once safe, one spec drives three downstream workflows:

```
account.spec
    │
    ├──► spectre generate-driver  →  Rust conformance test harness  (Ch 12)
    ├──► spectre generate-monitor →  embedded runtime monitor       (Ch 13)
    └──► spectre sync account.rs  →  drift detection as code evolves (Ch 11)
```

---

## What's next

- **Chapter 2** — every language construct with examples
- **Chapter 3** — safety invariants in depth, reading violations, repair suggestions
- **Chapter 9** — picking the right backend for your problem
# Chapter 2: The Spectre language

---

## Variables

You declare state with `var`. Four primitive types:

| Type | Declaration | Default |
|------|-------------|---------|
| `int` | `var count: int` | `0` |
| `bool` | `var active: bool` | `false` |
| `str` | `var label: str` | `""` |
| `float` | `var ratio: float` | `0.0` |

```spectre
var count:  int
var active: bool
var label:  str
var ratio:  float
```

Variables live for the entire run — there are no local variables, only state.

---

## Enumerations

Define a named set of values with `enum`:

```spectre
enum TrafficLight { Red, Yellow, Green }
var signal: TrafficLight
```

A complete traffic-light spec:

```spectre
enum TrafficLight { Red, Yellow, Green }

var signal: TrafficLight

init { signal = Red }

action go     { require signal = Red     signal' = Green  }
action caution { require signal = Green  signal' = Yellow }
action stop   { require signal = Yellow  signal' = Red    }

invariant only_valid_transitions {
  signal = Red or signal = Yellow or signal = Green
}
```

```bash
spectre verify traffic.spec
# ✓ safe — 3 states explored (exhaustive)
```

---

## Parameters

Parameters are named constants. They don't change during a run:

```spectre
param CAPACITY:  int = 10
param THRESHOLD: int = 7
```

- Override on the command line: `spectre verify queue.spec --param CAPACITY=20`
- Sweep a range: `spectre verify queue.spec --param-range CAPACITY=1..20`
- Use in invariants: `invariant not_full { items <= CAPACITY }`

---

## Init block

```spectre
init {
  balance = 0
  active  = true
  phase   = Idle
}
```

- Every variable you list is set to that value.
- Variables not listed get their type's zero value: `0`, `false`, `""`, first enum variant.
- You can list variables in any order.

---

## Actions

An action describes one atomic transition of the system:

```
action deposit {           ← name
  require status = Active  ← precondition — action disabled if false
  require amount > 0       ← multiple requires are AND'd together
  balance' = balance + amount  ← next-state value (prime notation)
  count'   = count + 1         ← all assignments apply simultaneously
}
```

Key rules:

- `x'` means "the value of `x` in the next state."
- An action with a failing `require` is simply **disabled** — it's not an error.
- All assignments in one action are **simultaneous** — `x' = y` and `y' = x` swaps them correctly.
- Variables not assigned by an action **keep their current value** automatically.

### Example 1: simple counter

```spectre
var value: int
param MAX: int = 10

action increment {
  require value < MAX
  value' = value + 1
}

action decrement {
  require value > 0
  value' = value - 1
}
```

### Example 2: bank deposit with two requires

```spectre
var balance: int
param LIMIT: int = 1000

action deposit {
  require balance < LIMIT
  require balance + 100 <= LIMIT
  balance' = balance + 100
}
```

### Example 3: enum state machine transition

```spectre
enum Phase { Idle, Running, Done }
var phase: Phase
var started: bool

action start {
  require phase = Idle
  phase'   = Running
  started' = true
}

action finish {
  require phase = Running
  phase' = Done
}
```

---

## Invariants

You're telling Spectre: "this must be true in every reachable state."

```spectre
invariant solvent    { balance >= 0 }
invariant within_cap { balance <= LIMIT }
```

- Spectre checks every invariant after every action in every reachable state.
- A violation = one reachable state where the condition is false.
- You can have as many invariants as you want; all are checked independently.

---

## The complete bank account spec

Everything above pulled together into one file:

```spectre
// bank.spec

enum Status { Active, Frozen }

var balance: int
var status:  Status
param LIMIT: int = 1000

init {
  balance = 0
  status  = Active
}

action deposit {
  require status = Active
  require balance + 100 <= LIMIT
  balance' = balance + 100
}

action withdraw {
  require status = Active
  require balance >= 100
  balance' = balance - 100
}

action freeze   { status' = Frozen }
action unfreeze { status' = Active }

invariant solvent     { balance >= 0 }
invariant within_cap  { balance <= LIMIT }
```

```bash
spectre verify bank.spec
```

```
✓ safe — 12 states explored (exhaustive)
  Invariants checked: solvent, within_cap
```

---

## Temporal properties

Temporal properties let you express liveness — things that must *eventually* happen, or *always* remain true over infinite runs:

```spectre
temporal always_solvent  { always(balance >= 0) }
temporal progress        { eventually(balance > 0) }
temporal freeze_recovers { always(status = Frozen → eventually(status = Active)) }
```

| Keyword | Meaning |
|---------|---------|
| `always(P)` | P holds in every state of every run |
| `eventually(P)` | P holds in some state of every run |
| `P until Q` | P holds until Q becomes true |
| `WF(action)` | Weak fairness: if `action` is always enabled, it eventually fires |
| `SF(action)` | Strong fairness: if `action` is infinitely often enabled, it fires infinitely often |

For **fairness** (WF/SF) to be validated correctly, use `--backend tlc` or `--backend apalache`. The default BFS backend uses a heuristic. Chapter 5 covers this in detail.

---

## Descriptions

Add a human-readable description to any action:

```spectre
description "Deposits 100 units — requires account active and not at cap"
action deposit {
  require status = Active
  require balance + 100 <= LIMIT
  balance' = balance + 100
}
```

These appear in violation traces and generated code.

---

## Quick-reference card

| Keyword | Purpose |
|---------|---------|
| `var` | Declare a state variable |
| `param` | Declare a named constant |
| `enum` | Declare an enumeration |
| `init` | Set initial values |
| `action` | Declare a state transition |
| `require` | Precondition for an action |
| `invariant` | Safety property (checked in every state) |
| `temporal` | Liveness / fairness property |
| `always` | Temporal: holds in every state |
| `eventually` | Temporal: holds in some future state |
| `until` | Temporal: P holds until Q |
| `WF` | Weak fairness over an action |
| `SF` | Strong fairness over an action |
| `description` | Human-readable label for an action |

---

## What's next

- **Chapter 3** — safety properties in depth: writing good invariants, reading violation traces
- **Chapter 4** — temporal properties: liveness and fairness with real examples
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
# Chapter 4: Temporal properties and liveness

Invariants catch bad states. Temporal properties catch bad *trajectories* — things that should eventually happen but don't, or conditions that should hold across a sequence of steps.

---

## Safety vs liveness

| Property kind | Question it answers | Spectre keyword |
|--------------|---------------------|----------------|
| Safety | "Does anything bad ever happen?" | `invariant` |
| Liveness | "Does something good eventually happen?" | `temporal` |

You need both. A spec where nothing ever happens is trivially safe but useless.

---

## `always` — a condition that holds forever

```spectre
temporal always_positive { always(balance >= 0) }
```

This is equivalent to an invariant — but expressed in the temporal language. Use `invariant` for simple per-state checks; use `always` when you're building up compound temporal expressions.

Full example — traffic light that must never show `Off`:

```spectre
enum Light { Red, Yellow, Green }
var signal: Light

init { signal = Red }

action to_green  { require signal = Red     signal' = Green  }
action to_yellow { require signal = Green   signal' = Yellow }
action to_red    { require signal = Yellow  signal' = Red    }

invariant always_lit { signal != Off }
temporal  never_off  { always(signal != Off) }
```

```bash
spectre verify traffic.spec
# ✓ safe — 3 states explored (exhaustive)
#   Invariants: always_lit ✓
#   Temporal:   never_off ✓
```

---

## `eventually` — something must happen

```spectre
temporal reaches_max { eventually(value = MAX) }
```

This says: on every execution path, `value` must at some point equal `MAX`.

Counter that counts to MAX and stops:

```spectre
var value: int
param MAX: int = 4

init { value = 0 }

action increment { require value < MAX  value' = value + 1 }

invariant non_negative { value >= 0 }
temporal  reaches_max  { eventually(value = MAX) }
```

```bash
spectre verify counter.spec
# ✓ safe — 5 states explored (exhaustive)
#   Temporal: reaches_max ✓
```

Now add a `reset` action and watch the property fail:

```spectre
action reset { value' = 0 }
```

```bash
spectre verify counter.spec
# ✗ violation — temporal property 'reaches_max' not satisfied
#
# Counterexample (cycle):
#   step 0  init:      value = 0
#   step 1  increment: value = 1
#   step 2  reset:     value = 0   ← loops here forever
#   ...
#   value never reaches 4
```

The BFS found a cycle where `reset` fires repeatedly and `value` never reaches `MAX`.

---

## `until` — hold one condition until another

```spectre
temporal safe_until_done { (errors = 0) until (phase = Done) }
```

This says: `errors = 0` must hold on every step *until* `phase = Done` is reached.

Download manager example:

```spectre
enum Phase { Downloading, Done }
var phase:  Phase
var errors: int

init { phase = Downloading  errors = 0 }

action download { require phase = Downloading  phase' = Done         }
action fail     { require phase = Downloading  errors' = errors + 1  }

temporal no_errors_until_done { (errors = 0) until (phase = Done) }
```

```bash
spectre verify download.spec
# ✗ violation — 'no_errors_until_done' not satisfied
# step 1: fail → errors = 1, phase still Downloading
```

Fix: add `require errors = 0` to `fail`, or remove it if errors are not allowed before completion.

---

## `leads_to` — if P then eventually Q

```spectre
temporal request_served { always(pending → eventually(not pending)) }
```

Every time a request becomes pending, it must eventually be cleared. This is the "leads-to" pattern: `always(P → eventually(Q))`.

Simple request/response server:

```spectre
var pending: bool
var served:  bool

init { pending = false  served = false }

action request { require not pending  pending' = true               }
action serve   { require pending      pending' = false  served' = true }
action reset   { served' = false                                    }

temporal always_served { always(pending → eventually(not pending)) }
```

```bash
spectre verify server.spec
# ✓ safe — 4 states explored (exhaustive)
#   Temporal: always_served ✓
```

Remove the `serve` action and re-verify:

```bash
spectre verify server-broken.spec
# ✗ violation — 'always_served' not satisfied
# pending becomes true and stays true — no serve action ever fires
```

---

## Combining temporal properties

A spec can have multiple temporal properties. Spectre checks all of them.

```spectre
var value:    int
var request:  bool
var response: bool
param MAX: int = 5

init { value = 0  request = false  response = false }

action increment { require value < MAX  value' = value + 1            }
action req       { require not request  request' = true               }
action respond   { require request      request' = false  response' = true }

temporal always_bounded   { always(value <= MAX)                      }
temporal reaches_max      { eventually(value = MAX)                   }
temporal request_answered { always(request → eventually(not request)) }
```

What each property covers:

```
States:  ●───●───●───●───●───●
        v=0 v=1 v=2 v=3 v=4 v=5=MAX

always_bounded:     every node   →  value <= MAX          ✓ per-state
reaches_max:        some path    →  reaches value = MAX   ✓ across paths
request_answered:   each request →  eventually cleared    ✓ per-event pair
```

---

## Temporal verification with large state spaces

If BFS hits `--max-states`, temporal properties are only checked over explored states:

```bash
spectre verify large.spec --max-states 5000
# bounded — hit state limit
# temporal properties checked over 5000 states only
# Use --backend tlc for exhaustive temporal verification
```

Use `--backend tlc` or `--backend apalache` for exhaustive temporal verification on large specs (Chapter 15).

---

## Quick-reference

| Operator | Meaning | Use when |
|----------|---------|----------|
| `always(P)` | P holds in every reachable state | Global safety, expressed temporally |
| `eventually(P)` | Every path eventually reaches P | Progress / reachability goals |
| `P until Q` | P holds on all steps until Q | Precondition that must hold until a goal |
| `always(P → eventually(Q))` | Every P event leads to Q | Request/response, lock/unlock pairs |
| `WF(action)` | Action fires if continuously enabled | Chapter 5 |
| `SF(action)` | Action fires if infinitely often enabled | Chapter 5 |
# Chapter 5: Fairness — WF and SF

Temporal properties say *what* must eventually happen. Fairness conditions say *which executions count* when checking those properties. Without fairness, the verifier is allowed to assume a scheduler that permanently ignores actions — which makes most liveness properties unprovable.

---

## Why fairness matters

```spectre
var token: bool

init { token = false }

action grant  { require not token  token' = true  }
action revoke { require token      token' = false }

temporal eventually_granted { eventually(token) }
```

```bash
spectre verify token.spec
# ✗ violation — 'eventually_granted' not satisfied
# Counterexample: revoke fires forever, grant never fires
```

The spec *allows* granting — but the verifier found a valid execution where `grant` is always skipped. That execution is "unfair."

```
Without fairness:          With WF(grant):
  ●──revoke──●──revoke──►    ●──grant──► token=true  ✓
  ↑ grant is skipped         (grant must fire if always enabled)
```

---

## Weak fairness: WF(action)

> "If an action stays enabled, it must eventually fire."

```spectre
temporal wf_grant { WF(grant) }
```

Add this to the spec:

```spectre
var token: bool

init { token = false }

action grant  { require not token  token' = true  }
action revoke { require token      token' = false }

temporal eventually_granted { eventually(token)  }
temporal wf_grant           { WF(grant)          }
```

```bash
spectre verify token.spec
# ✓ safe — 2 states explored (exhaustive)
#   Temporal: eventually_granted ✓  wf_grant ✓
```

WF rules out executions where `grant` is permanently skipped while enabled.

### Mutex with progress guarantee

```spectre
enum ProcState { Waiting, Running, Done }
var p1:   ProcState
var p2:   ProcState
var lock: bool

init { p1 = Waiting  p2 = Waiting  lock = false }

action p1_acquire { require p1 = Waiting and not lock  p1' = Running  lock' = true  }
action p1_release { require p1 = Running               p1' = Done     lock' = false }
action p2_acquire { require p2 = Waiting and not lock  p2' = Running  lock' = true  }
action p2_release { require p2 = Running               p2' = Done     lock' = false }

invariant mutex      { not (p1 = Running and p2 = Running) }
temporal p1_progress { WF(p1_acquire) }
temporal p2_progress { WF(p2_acquire) }
temporal p1_finishes { eventually(p1 = Done) }
temporal p2_finishes { eventually(p2 = Done) }
```

```bash
spectre verify mutex.spec
# ✓ safe — 9 states explored (exhaustive)
```

Without `WF(p1_acquire)`, `p1_finishes` would fail — the scheduler could run p2 forever and starve p1.

---

## Strong fairness: SF(action)

> "If an action is enabled infinitely often, it must fire infinitely often."

The difference from WF: SF handles actions that are *intermittently* enabled.

```spectre
var connected: bool
var retries:   int

init { connected = false  retries = 0 }

action disconnect { require connected      connected' = false            }
action connect    { require not connected  connected' = true             }
action retry      { require not connected  retries'   = retries + 1      }

temporal sf_retry { SF(retry) }
temporal eventually_connected { eventually(connected) }
```

`retry` is only enabled when `connected = false` — which isn't always the case. WF(retry) wouldn't help here because `retry` isn't *continuously* enabled. SF(retry) says: whenever `retry` becomes enabled again, it must fire.

### WF vs SF at a glance

| | WF(A) | SF(A) |
|-|-------|-------|
| Fires when | Continuously enabled | Infinitely often enabled |
| Strength | Weaker | Stronger |
| Use for | Lock acquisition, queue dequeue | Retry, intermittent tasks |
| Example | Mutex `acquire` | Network `retry` |

---

## WF and SF in the same spec

You can mix them freely:

```spectre
temporal acquire_wf   { WF(acquire)  }   // always-enabled action
temporal retry_sf     { SF(retry)    }   // intermittently-enabled action
temporal done_for_all { eventually(p1 = Done and p2 = Done) }
```

Spectre checks all temporal properties together.

---

## When the BFS heuristic isn't enough

Spectre's built-in BFS uses a cycle-filtering heuristic for fairness. It works for the majority of specs but can give wrong answers when:

- Multiple fairness conditions interact
- An action is enabled in some states of a cycle but not others
- The spec has deeply nested liveness goals

For these cases, use the TLC backend:

```bash
spectre verify mutex.spec --backend tlc
```

TLC performs proper SCC-based fairness checking — the same algorithm that's been validated in production TLA+ tooling for two decades. It translates `WF(action)` to `WF_vars(Action)` in TLA+ and lets TLC handle the rest.

```
WF/SF in Spectre
       │
       ▼
spectre verify (BFS heuristic)
       │
  confident? ──yes──► done
       │
       no (complex fairness, large spec)
       │
       ▼
spectre verify --backend tlc
(SCC-based, validated)
```

See Chapter 15 for full TLC and Apalache backend details.

---

## Fairness checklist

When writing a spec with liveness properties:

- Add `WF(action)` for every action that represents a "service obligation" (lock acquisition, queue drain, message delivery)
- Add `SF(action)` for actions that are intermittently available (retries, reconnects, elections)
- If `eventually(P)` fails unexpectedly, add fairness before debugging the spec logic
- For critical liveness proofs, validate with `--backend tlc`
# Chapter 6: Concurrent and distributed systems

Spectre models concurrency by letting any action fire at any time. BFS explores every possible interleaving. This makes it easy to find races that are nearly impossible to trigger in testing.

---

## What's different about concurrent specs

Three things you don't see in sequential specs:

- **Shared state** — multiple actors read and write the same variables
- **Interleaving** — actions from different actors can fire in any order
- **Races** — two actors attempt incompatible transitions simultaneously

You don't need threads or processes. You model each actor as a set of actions and let Spectre find the bad interleavings.

---

## Start with a broken mutex

```spectre
// mutex-broken.spec
var p1_in_cs: bool
var p2_in_cs: bool

init { p1_in_cs = false  p2_in_cs = false }

action p1_enter { p1_in_cs' = true  }
action p1_exit  { p1_in_cs' = false }
action p2_enter { p2_in_cs' = true  }
action p2_exit  { p2_in_cs' = false }

invariant mutex { not (p1_in_cs and p2_in_cs) }
```

```bash
spectre verify mutex-broken.spec
# ✗ violation — invariant 'mutex' after action 'p2_enter'
#
# Counterexample (2 steps):
#   step 0  init:     p1_in_cs=false, p2_in_cs=false
#   step 1  p1_enter: p1_in_cs=true,  p2_in_cs=false
#   step 2  p2_enter: p1_in_cs=true,  p2_in_cs=true  ← VIOLATION
```

Spectre found the race in 2 steps.

---

## Fix it with a lock

```spectre
// mutex.spec
var p1_in_cs: bool
var p2_in_cs: bool
var lock:     bool

init { p1_in_cs = false  p2_in_cs = false  lock = false }

action p1_enter { require not lock  p1_in_cs' = true   lock' = true  }
action p1_exit  { require p1_in_cs  p1_in_cs' = false  lock' = false }
action p2_enter { require not lock  p2_in_cs' = true   lock' = true  }
action p2_exit  { require p2_in_cs  p2_in_cs' = false  lock' = false }

invariant mutex      { not (p1_in_cs and p2_in_cs) }
invariant lock_owner { lock → (p1_in_cs or p2_in_cs) }
temporal  p1_progress { WF(p1_enter) }
temporal  p2_progress { WF(p2_enter) }
```

```bash
spectre verify mutex.spec
# ✓ safe — 9 states explored (exhaustive)
```

State machine for this model:

```
           p1_enter                    p1_exit
  (free) ───────────► (p1 holds lock) ──────────► (free)
    │                                                 ▲
    │     p2_enter                    p2_exit         │
    └───────────────► (p2 holds lock) ────────────────┘
```

Only one process can hold the lock — `p2_enter` and `p1_enter` both `require not lock`, so they can never both fire from the same state.

---

## Producer-consumer with a bounded buffer

```spectre
param CAPACITY: int = 3
var items: int

init { items = 0 }

action produce { require items < CAPACITY  items' = items + 1 }
action consume { require items > 0         items' = items - 1 }

invariant no_overflow  { items <= CAPACITY }
invariant no_underflow { items >= 0        }
temporal  eventually_full { eventually(items = CAPACITY) }
```

```bash
spectre verify buffer.spec --param CAPACITY=3
# ✓ safe — 4 states explored (exhaustive)

spectre verify buffer.spec --param CAPACITY=10
# ✓ safe — 11 states explored (exhaustive)
```

Adding a second producer (`produce2`) immediately doubles the state space — BFS handles it automatically.

---

## Distributed message queue

Three-role system: producers enqueue, consumers dequeue, an accounting invariant keeps them honest.

```spectre
param MAX_QUEUE: int = 5
var queue_len: int
var produced:  int
var consumed:  int

init { queue_len = 0  produced = 0  consumed = 0 }

action enqueue {
  require queue_len < MAX_QUEUE
  queue_len' = queue_len + 1
  produced'  = produced + 1
}

action dequeue {
  require queue_len > 0
  queue_len' = queue_len - 1
  consumed'  = consumed + 1
}

invariant queue_bounded  { queue_len <= MAX_QUEUE }
invariant queue_positive { queue_len >= 0         }
invariant accounting     { produced = consumed + queue_len }
temporal  queue_drains   { always(queue_len > 0 → eventually(queue_len = 0)) }
```

```bash
spectre verify queue.spec
# ✓ safe — 21 states explored (exhaustive)
```

The `accounting` invariant is the most valuable one here. It catches any race where messages are created or destroyed without both counters updating — a class of bugs that integration tests rarely find.

---

## Good invariants for concurrent specs

| Pattern | Invariant example |
|---------|------------------|
| Mutual exclusion | `not (p1_in_cs and p2_in_cs)` |
| Accounting identity | `produced = consumed + in_flight` |
| State consistency | `locked → owner_known` |
| Bounded resource | `active_connections <= MAX_CONN` |
| No phantom state | `waiting → (request_sent and not response_received)` |

Write at least one invariant of each type that applies to your system.

---

## When the state space gets large

Three-process or three-role specs can hit tens of thousands of states:

```bash
spectre verify 3-party.spec --max-states 50000
# bounded — hit state limit (50000 states)
```

Options when BFS doesn't finish:

| Option | When to use |
|--------|-------------|
| `--max-states 200000` | More time, same algorithm |
| `--backend apalache` | Symbolic — handles larger spaces |
| `--por` | Partial-order reduction — skips equivalent interleavings |
| Split the spec | Verify each component separately first |

```bash
spectre verify 3-party.spec --backend apalache
# uses Apalache's symbolic engine — see Chapter 15
```

---

## Concurrency workflow

```
Identify shared variables
         │
         ▼
Write one action per actor per operation
         │
         ▼
Add require guards for preconditions
         │
         ▼
Write mutex + accounting invariants
         │
         ▼
spectre verify
         │
    violation? ──yes──► read the counterexample trace
         │                       │
         no                      ▼
         │              strengthen requires or add lock variable
         ▼
Add WF/SF temporal properties
         │
         ▼
spectre verify --backend tlc  (for validated fairness)
```
# Chapter 7: Modules, parameters, and reuse

Parameters let you change a spec's constants without editing the file. Modules let you split a large spec into named, reusable pieces.

---

## Parameters — changing the spec without editing it

```spectre
param CAPACITY:  int = 10
param THRESHOLD: int = 7
var items: int

invariant not_full       { items <= CAPACITY  }
invariant below_threshold { items <  THRESHOLD }
```

Three ways to set parameter values:

- **Default** — just run `spectre verify queue.spec`
- **Override** — `spectre verify queue.spec --param CAPACITY=20`
- **Sweep** — `spectre verify queue.spec --param-range CAPACITY=1..20`

### Parameter sweep

```bash
spectre verify queue.spec --param-range CAPACITY=1..10
```

```
CAPACITY  States  Result
--------  ------  ------
       1       3  safe (exhaustive)
       2       6  safe (exhaustive)
       3      10  safe (exhaustive)
       4      15  safe (exhaustive)
       5      21  safe (exhaustive)
      10      66  safe (exhaustive)
```

Sweeping is useful for finding the exact value where a spec breaks or where the state space starts to explode.

---

## Modules — reusable spec components

A module bundles variables, params, init, actions, and invariants under a name.

```spectre
module Counter {
  var count: int
  param MAX: int = 10

  init { count = 0 }

  action increment { require count < MAX  count' = count + 1 }
  action decrement { require count > 0    count' = count - 1 }

  invariant non_negative  { count >= 0      }
  invariant within_bounds { count <= MAX    }
}
```

Import and compose it in another spec:

```spectre
import Counter

var active: bool

init { active = true }

action toggle { active' = not active }

invariant active_non_negative { active → Counter.count >= 0 }
```

```bash
spectre verify composed.spec
# ✓ safe — 22 states explored (exhaustive)
```

---

## Module inheritance — parameterized reuse

```spectre
module Semaphore {
  param SLOTS: int = 1
  var held: int

  init { held = 0 }

  action acquire { require held < SLOTS  held' = held + 1 }
  action release { require held > 0      held' = held - 1 }

  invariant valid { held >= 0 and held <= SLOTS }
}

module Mutex  extends Semaphore { param SLOTS = 1 }
module RWLock extends Semaphore { param SLOTS = 3 }
```

The same `valid` invariant covers both variants — you write it once and get it for free in every subtype.

```bash
spectre verify mutex.spec   # SLOTS=1 → 3 states
spectre verify rwlock.spec  # SLOTS=3 → 5 states
```

---

## Splitting a large spec into modules

**Before** — one flat file, hard to navigate:

```spectre
var user_count:    int
var session_count: int
var db_connections: int
param MAX_USERS:   int = 1000
param MAX_SESSIONS: int = 5000
param MAX_CONNS:   int = 20

action add_user    { ... }
action remove_user { ... }
action open_session { ... }
action close_session { ... }
action open_connection { ... }
action close_connection { ... }
// ... 40 more lines
```

**After** — three focused modules:

```spectre
// users.spec
module UserManager {
  var user_count: int
  param MAX_USERS: int = 1000

  init { user_count = 0 }

  action add_user    { require user_count < MAX_USERS  user_count' = user_count + 1 }
  action remove_user { require user_count > 0          user_count' = user_count - 1 }

  invariant valid_count { user_count >= 0 and user_count <= MAX_USERS }
}
```

```spectre
// sessions.spec
module SessionManager {
  var session_count: int
  param MAX_SESSIONS: int = 5000

  init { session_count = 0 }

  action open_session  { require session_count < MAX_SESSIONS  session_count' = session_count + 1 }
  action close_session { require session_count > 0             session_count' = session_count - 1 }

  invariant valid_sessions { session_count >= 0 and session_count <= MAX_SESSIONS }
}
```

```spectre
// app.spec — the composition
import UserManager
import SessionManager

invariant sessions_per_user {
  SessionManager.session_count <= UserManager.user_count * 10
}
```

Benefits of splitting:
- Each module verifies independently — catch bugs early
- Modules are reusable across projects
- Invariants at the composition layer catch cross-module bugs
- State spaces stay small per module

---

## Workflow: from flat to modular

```
Single flat spec
      │
      ▼
Identify natural groupings
  (resources, actors, subsystems)
      │
      ▼
Extract each group into a module
      │
      ▼
spectre verify each module independently
      │
      ▼
Compose with import
      │
      ▼
Add cross-module invariants
      │
      ▼
spectre verify composition
```

---

## Parameter best practices

- Always set a **sensible default** — specs should verify out of the box
- Use **ALL_CAPS** names for params (`MAX`, `CAPACITY`, `TIMEOUT`) to distinguish from variables
- Keep params **small for development** — verify with `--param N=3` before `--param N=100`
- Sweep params in CI to catch spec regressions across the whole value range
# Chapter 8: Reading counterexamples and debugging

When verification fails, Spectre gives you the shortest possible path to the violation. Here's how to read it and fix the spec.

---

## Anatomy of a violation trace

```
✗ violation — invariant 'solvent' after action 'withdraw'

Counterexample (4 steps):
  step 0  init:      balance = 0,    status = Active
  step 1  deposit:   balance = 100,  status = Active
  step 2  deposit:   balance = 200,  status = Active
  step 3  withdraw:  balance = 200,  status = Active
  step 4  withdraw:  balance = -100, status = Active  ← VIOLATION
                     ↑ changed: balance 200 → -100

Failed invariant: solvent  (balance >= 0)
  balance = -100  violates  balance >= 0
```

Every part of the output tells you something:

| Part | What it tells you |
|------|------------------|
| **Violation line** | Which invariant failed, which action triggered it |
| **Counterexample (N steps)** | BFS found the *shortest* path — no shorter path exists |
| **Each step** | Which action ran + the full state after it |
| **← VIOLATION** | The exact state where the invariant first fails |
| **Changed field** | What changed from the previous step |
| **Failed invariant** | The expression and the concrete values that broke it |

---

## Five-step debugging workflow

```
1. Read the violated invariant
         │  "solvent: balance >= 0"
         ▼
2. Look at the violation step
         │  "withdraw: balance = -100"
         ▼
3. Read backwards — what enabled this action?
         │  Previous state: balance = 200, no require check
         ▼
4. Find the missing guard
         │  withdraw has no "require balance >= amount"
         ▼
5. Add require, re-verify
         │  spectre verify bank.spec  →  ✓ safe
         ▼
       done
```

---

## Worked example: the transfer bug

```spectre
// transfer-broken.spec
var alice: int
var bob:   int

init { alice = 200  bob = 100 }

action transfer {
  // BUG: no check that alice has enough
  alice' = alice - 100
  bob'   = bob   + 100
}

invariant alice_solvent { alice >= 0       }
invariant bob_solvent   { bob   >= 0       }
invariant total_fixed   { alice + bob = 300 }
```

```bash
spectre verify transfer-broken.spec
```

```
✗ violation — invariant 'alice_solvent' after action 'transfer'

Counterexample (3 steps):
  step 0  init:     alice = 200, bob = 100
  step 1  transfer: alice = 100, bob = 200
  step 2  transfer: alice = 0,   bob = 300
  step 3  transfer: alice = -100, bob = 400  ← VIOLATION
                    ↑ changed: alice 0 → -100

Failed invariant: alice_solvent  (alice >= 0)
  alice = -100  violates  alice >= 0
```

The fix — add a require:

```spectre
action transfer {
  require alice >= 100
  alice' = alice - 100
  bob'   = bob   + 100
}
```

```bash
spectre verify transfer-fixed.spec
# ✓ safe — 7 states explored (exhaustive)
```

Notice `total_fixed` — it catches bugs where money is created or destroyed. You get that protection for free just by writing the accounting identity as an invariant.

---

## Using `--suggest-repairs`

When a violation is found, `--suggest-repairs` analyses the action and invariant and proposes patches.

```bash
spectre verify transfer-broken.spec --suggest-repairs
```

```
Repair candidates (CANDIDATE — review before applying):

Violation 1: action 'transfer' violates 'alice_solvent'
  [1] CANDIDATE strengthen_require: add `require alice >= 100` to action transfer
      Validation: original violation fixed · no deadlock · reachability unchanged — SAFE
  [2] CANDIDATE add_invariant_guard: add `when alice >= 100` to action transfer
      Validation: original violation fixed · deadlock at alice=0 — review required
```

Two candidate types — pick the right one for your situation:

| Candidate type | Pick when | Risk |
|---------------|-----------|------|
| `strengthen_require` | The action simply shouldn't fire without the condition | Minimal — action is disabled, not blocked |
| `add_invariant_guard` | You want the action blocked at the invariant boundary | Can deadlock if the state has no other exit |

### Understanding deadlock warnings

```
Validation: original violation fixed · deadlock at alice=0 — review required
```

A deadlock is a reachable state where **no action is enabled**. This isn't always wrong — sometimes "nothing can happen" is correct behavior. Check:

- Is that final state intentional? (e.g., a `Done` phase)
- Or is the system permanently stuck? (a real bug in the repair)

---

## Multiple violations

Spectre checks all invariants and reports the first violation it finds (shortest BFS path). If you have multiple bugs, fix them one at a time — each fix may expose the next shortest path.

```bash
spectre verify buggy.spec
# ✗ violation — invariant 'alice_solvent' ...

# Fix alice_solvent, then:
spectre verify partially-fixed.spec
# ✗ violation — invariant 'total_fixed' ...
```

---

## Dealing with bounded results

```
bounded — hit state limit (5000 states)
```

This means BFS stopped before exploring everything. Your spec might be safe — or the violation might be deeper than the limit. Your options:

1. `--max-states 50000` — increase the limit and try again
2. `--max-states -1` — unlimited (may not terminate for large specs)
3. `--backend apalache` — symbolic verification, handles large spaces exhaustively
4. `--backend z3 --bmc-depth 30` — bounded check to depth 30; fast, not exhaustive

```bash
spectre verify large.spec --max-states 50000
spectre verify large.spec --backend apalache
spectre verify large.spec --backend z3 --bmc-depth 30
```

Always check whether the output says `exhaustive` or `bounded` — it matters for how much you can trust the result.

---

## Common patterns and their fixes

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| Negative integer variable | Missing lower-bound require | `require x >= 0` before decrement |
| Variable exceeds param | Missing upper-bound require | `require x < MAX` before increment |
| Enum in wrong state | Action enabled in wrong phase | `require phase = Expected` |
| Two actors in critical section | No mutual exclusion | Add lock variable + requires |
| Accounting identity broken | Value leaked or doubled | Check all actions modify both sides |

---

## Tips for writing debuggable specs

- **Write accounting identities** — `alice + bob = TOTAL` catches money bugs for free
- **Name invariants clearly** — the name is the first thing you see in a violation
- **One concept per action** — large multi-step actions are hard to trace
- **Start with small params** — `--param N=3` first, then scale up
- **Verify after every change** — short iteration loops catch bugs while the context is fresh
# Chapter 9: Choosing a verification backend

`spectre verify` supports four backends. The default (BFS) works for most specs. The others trade setup complexity for more power.

---

## The four backends at a glance

```
spectre verify myspec.spec
          │
          │  --backend bfs        (default)
          │  --backend tlc
          │  --backend apalache
          │  --backend z3  --bmc-depth N
          │
          ▼
    ┌───────────┬──────────────┬──────────────┬───────────────┐
    │   BFS     │    TLC       │  Apalache    │     Z3        │
    ├───────────┼──────────────┼──────────────┼───────────────┤
    │ Exhaustive│ Exhaustive   │ Exhaustive   │ Bounded       │
    │ for small │ + validated  │ symbolic     │ up to depth N │
    │ specs     │ fairness     │ large spaces │               │
    ├───────────┼──────────────┼──────────────┼───────────────┤
    │ No deps   │ TLA+ tools   │ Apalache bin │ z3 binary     │
    │           │ (jar or bin) │              │               │
    └───────────┴──────────────┴──────────────┴───────────────┘
```

---

## Quick decision guide

```
Is your spec small (< 10k states)?
  └─ yes ──► --backend bfs  (default, zero setup)

Does it use WF/SF fairness and you need validated results?
  └─ yes ──► --backend tlc  or  --backend apalache

Is the state space too large for BFS?
  └─ yes ──► --backend apalache  (symbolic — handles millions of states)

Just hunting for shallow bugs fast?
  └─ yes ──► --backend z3 --bmc-depth 20  (finds bugs in seconds)

In CI and want a quick regression check?
  └─ yes ──► --backend z3 --bmc-depth 30  (fast, catches most regressions)
```

---

## BFS backend (default)

The built-in breadth-first search over your spec's reachable states. No external tools needed.

```bash
spectre verify counter.spec
# ✓ safe — 11 states explored (exhaustive)

spectre verify counter.spec --max-states 50000
# ✓ safe — 11 states explored (exhaustive)
```

**Best for:**
- Specs with up to ~50k reachable states
- Fast iteration during development
- No setup overhead

**Limitations:**
- WF/SF fairness is heuristic — not formally validated
- Memory-bound on large specs (> 100k states gets slow)

---

## TLC backend

Translates your spec to TLA+ and runs it through TLC — the reference TLA+ model checker.

```bash
# Install TLA+ tools
brew install tla-tools          # macOS
# or download tla2tools.jar from https://github.com/tlaplus/tlaplus/releases

# Run with TLC
spectre verify mutex.spec --backend tlc
spectre verify mutex.spec --backend tlc --tlc-jar ~/tools/tla2tools.jar
```

```
Verifying mutex.spec with backend tlc...
mutex.spec: ✓ safe (exhaustive)
  Note: TLC checked all reachable states
```

**Best for:**
- Specs with `WF(action)` or `SF(action)` — TLC validates fairness correctly
- Cross-checking a spec against the TLA+ ecosystem
- Teams already familiar with TLA+ tooling

**Limitations:**
- Slower than BFS for small specs
- Requires Java + TLA+ tools installed

### What Spectre generates

For each `.spec` file, Spectre creates two temp files:

```
---- MODULE myspec ----
EXTENDS Integers, TLC

VARIABLES balance, status

Init == /\ balance = 0
        /\ status  = "Active"

Deposit == /\ status = "Active"
           /\ balance < 1000
           /\ balance' = balance + 100
           /\ UNCHANGED status

Next == Deposit \/ Withdraw \/ Freeze \/ Unfreeze

Spec == Init /\ [][Next]_vars /\ WF_vars(Deposit)
====
```

```
INIT Init
NEXT Next
INVARIANT solvent
```

---

## Apalache backend

Apalache is a symbolic model checker for TLA+. Instead of enumerating states, it uses Z3 internally to reason about sets of states at once.

```bash
# Install Apalache
brew install apalache
# or download from https://github.com/informalsystems/apalache/releases

# Run with Apalache
spectre verify large-spec.spec --backend apalache
spectre verify large-spec.spec --backend apalache \
    --apalache-bin /opt/apalache/bin/apalache-mc
```

**Best for:**
- Large state spaces that exhaust BFS memory
- Specs with many integer variables (Apalache reasons symbolically over integers)
- Production-grade verification with validated fairness

**Limitations:**
- Slowest to start (JVM startup + type checking)
- Requires Apalache installed (Java 17+)

### BFS vs Apalache on a large spec

```
spectre verify raft.spec --max-states 100000
# bounded — hit state limit (100000 states)

spectre verify raft.spec --backend apalache
# ✓ safe (exhaustive)
# Note: Apalache checked all reachable states symbolically
```

---

## Z3 backend (bounded model checking)

Encodes your spec as an SMT-LIB2 query and asks Z3 to find a violation within N steps.

```bash
# Install Z3
brew install z3

# Default depth (20)
spectre verify bank.spec --backend z3

# Deeper search
spectre verify bank.spec --backend z3 --bmc-depth 50
```

When a violation exists within the depth:
```
Verifying bank.spec with backend z3...
bank.spec: ✗ violation found
  Note: bounded model checking to depth 20
  Violation 1: invariant "solvent"
    step 3: balance = -100
```

When no violation is found within the depth:
```
bank-fixed.spec: ✓ safe
  Note: bounded model checking to depth 20 — not exhaustive
```

> `✓ safe` from Z3 means **no violation up to depth 20** — not a proof of full safety. Use `--backend bfs` or `--backend apalache` for an exhaustive result.

**Best for:**
- Finding shallow bugs quickly (< 1 second for depth 20)
- CI smoke tests that should finish in a few seconds
- Specs too large for BFS where you just want fast feedback

**Limitations:**
- Bounded — can miss deep violations
- No fairness checking (liveness properties not verified)

---

## Comparison table

| | BFS | TLC | Apalache | Z3 |
|-|-----|-----|----------|----|
| Exhaustive | ✓ small specs | ✓ | ✓ symbolic | ✗ |
| Fairness (WF/SF) | heuristic | validated | validated | ✗ |
| Safety invariants | ✓ | ✓ | ✓ | ✓ bounded |
| Temporal properties | ✓ | ✓ | ✓ | ✗ |
| Dependencies | none | TLA+ tools | Apalache | z3 |
| Speed | fast | moderate | fast (large) | very fast |
| Setup time | 0 | 5 min | 10 min | 2 min |

---

## Progressive verification workflow

A practical strategy — escalate backend strength only when needed:

```
Write spec
    │
    ▼
--backend bfs         ← fast feedback, no deps
    │
  passes?
    │
    ├─ no ──► fix spec, repeat
    │
    └─ yes
         │
         ▼
    Has WF/SF fairness?
         ├─ yes ──► --backend tlc  or  --backend apalache
         └─ no  ──► done ✓  (BFS is exhaustive for small specs)
```

In CI, combine:

```bash
# Fast regression check (seconds)
spectre verify myspec.spec --backend z3 --bmc-depth 30

# Full verification (minutes, on merge to main)
spectre verify myspec.spec --backend apalache
```

---

## Installing the backends

| Backend | Install command | Verify install |
|---------|----------------|----------------|
| BFS | built in | `spectre --version` |
| TLC | `brew install tla-tools` | `tlc -help` |
| Apalache | `brew install apalache` | `apalache-mc version` |
| Z3 | `brew install z3` | `z3 --version` |

On Linux, download JARs/binaries from the respective GitHub releases pages and pass `--tlc-jar`, `--apalache-bin`, or `--z3-bin` flags.
# Chapter 10: Mining specs from Rust

### What mining does

```
src/account.rs ──► spectre mine ──► account.spec skeleton
                                         │
                                    add invariants
                                         │
                                    spectre verify
```

The miner reads your Rust source and extracts struct fields, assert guards, and field assignments — giving you a runnable spec in seconds.

What it extracts:
- `struct` fields → `var` declarations
- `assert!` / `debug_assert!` → `require` guards
- `if cond { return Err(..) }` → negated `require` guards
- `self.field = expr` / `self.field += n` → primed assignments
- `fn new(...)` constructor → `init` values
- Rust enums → Spectre enums

---

### Your first mine

```rust
// src/counter.rs
pub struct Counter {
    pub value: i64,
    pub max:   i64,
}

impl Counter {
    pub fn new(max: i64) -> Self {
        Counter { value: 0, max }
    }
    pub fn increment(&mut self) {
        assert!(self.value < self.max);
        self.value += 1;
    }
    pub fn decrement(&mut self) {
        assert!(self.value > 0);
        self.value -= 1;
    }
    pub fn reset(&mut self) {
        self.value = 0;
    }
}
```

```bash
spectre mine --lang rust src/counter.rs
```

Output:

```spectre
// Spectre spec mined from: src/counter.rs

var value: int
var max:   int

init {
  value = 0
  max   = 0
}

action increment {
  require value < max
  value' = value + 1
}

action decrement {
  require value > 0
  value' = value - 1
}

action reset {
  value' = 0
}
```

Save it and verify immediately:

```bash
spectre mine --lang rust src/counter.rs -o counter.spec
spectre verify counter.spec
# ✓ safe — 12 states explored (exhaustive)
```

---

### What gets extracted vs. what gets skipped

| Extracted ✓ | Skipped / approximated ✗ |
|-------------|--------------------------|
| Primitive fields (`i64`, `bool`, `String`) | `new()`, `default()`, `clone()`, `fmt()` |
| Rust enums — unit variants only | Tuple/struct enum variants (data dropped) |
| `assert!` / `debug_assert!` guards | `panic!`, `unwrap()`, `expect()` |
| `self.x = …`, `self.x += …`, `self.x -= …` | `checked_add`, `wrapping_add`, `saturating_sub` |
| `if cond { return Err(..) }` patterns | `Result<T,E>` error paths |
| `Arc<T>`, `Mutex<T>`, `Box<T>` wrappers | (inner type extracted, wrapper dropped) |

---

### Seeing what was skipped: `--show-omissions`

```bash
spectre mine --lang rust src/account.rs --show-omissions
```

```
Extraction omissions (4):
  SKIPPED_METHOD in new: constructor/utility method — not translated to action
  RESULT_RETURN in deposit: Result<(),TxError> return — error path abstracted
  OVERFLOW_OP in deposit: checked_add — saturating arithmetic abstracted
  TUPLE_VARIANT in TxError: InsufficientFunds(i64) — data dropped, unit variant kept

Mined: 3 fields, 1 enum, 4 actions, 4 omissions
```

Four omission kinds to act on:

- **RESULT_RETURN** — error paths not modelled. Add invariants for the error conditions manually if they matter.
- **OVERFLOW_OP** — arithmetic is unbounded in the spec. Add `require` guards if overflow-safe bounds are important.
- **TUPLE_VARIANT** — enum variant data was dropped. Add a `var` for the data if you need to reason about it.
- **SKIPPED_METHOD** — constructors, `fmt`, `drop`. Almost always safe to ignore.

The mine summary line shows how many omissions occurred:

```
Mined: 3 fields, 1 enum, 4 actions, 4 omissions (run --show-omissions for details)
```

---

### Evaluate a whole directory

```bash
spectre eval-rust --dir src/
```

```
# Rust Extraction Evaluation Report

## Per-File Summary

| File              | Struct      | Fields | Enums | Methods | Omissions | Success |
|-------------------|-------------|--------|-------|---------|-----------|---------|
| account.rs        | BankAccount |   3    |   1   |    4    |     4     |   50%   |
| counter.rs        | Counter     |   2    |   0   |    3    |     1     |  100%   |
| state_machine.rs  | Machine     |   4    |   2   |    2    |     4     |   50%   |

**Totals** — files: 3 | fields: 9 | methods: 9 | omissions: 9 | mean success: 67%

## Omissions by Kind

| Kind            | Count |
|-----------------|-------|
| SKIPPED_METHOD  |   3   |
| RESULT_RETURN   |   2   |
| OVERFLOW_OP     |   2   |
| TUPLE_VARIANT   |   2   |
```

**Success fraction** = methods with no body-level omissions (overflow / panic / Result / complex assign). Files at 100% can be verified with minimal manual work. Files at 50% need review.

Use this output to prioritise enrichment effort — start with files nearest 100%.

---

### Enrich the mined spec

Three things to always add after mining:

1. **Invariants** — the miner extracts no invariants (only guards). Add at least one per resource, one per ordering constraint.
2. **Temporal properties** — if your system makes progress guarantees, add `WF(action)` or `eventually(goal)`.
3. **Realistic init values** — `max = 0` won't verify anything useful. Set values that match production defaults.

Example enrichment for the counter spec:

```spectre
// counter.spec — after enrichment
var value: int
var max:   int

init {
  value = 0
  max   = 10    // ← set a realistic default
}

action increment {
  require value < max
  value' = value + 1
}

action decrement {
  require value > 0
  value' = value - 1
}

action reset {
  value' = 0
}

// ← these three lines are the additions
invariant non_negative  { value >= 0 }
invariant within_bounds { value <= max }
temporal  always_resets { WF(reset) }
```

---

### LLM-enhanced mining

```bash
ANTHROPIC_API_KEY=sk-... spectre mine --lang rust src/account.rs --ai
```

The LLM adds (always commented out — never applied automatically):
- One-sentence `description` for each action
- Up to three `invariant` candidates
- Up to three `temporal` property candidates

```spectre
// LLM-suggested invariants (review before enabling):
// invariant balanced_books { total_deposits >= total_withdrawals }
// invariant no_freeze_with_balance { frozen → balance = 0 }
```

Always review LLM suggestions against your actual requirements before uncommenting.

---

### The full mining workflow

```
1. spectre eval-rust --dir src/
   └── identify which files mine cleanly (Success = 100%)

2. spectre mine --lang rust src/target.rs -o target.spec
   └── get the skeleton

3. spectre mine ... --show-omissions
   └── understand what was skipped

4. Edit target.spec
   └── add invariants, fix init values, handle omissions

5. spectre verify target.spec
   └── safe → check in both files
       violation → fix spec, re-verify

6. spectre sync src/target.rs --spec target.spec  (ongoing)
   └── catch drift as code evolves
```
# Chapter 11: Drift detection and incremental re-verification

### The drift problem

```
Day 1:   mine spec ──► check in account.spec + account.rs ✓
              │
Day 30:  refactor withdraw() in Rust
              │
              ▼
         spec and code no longer agree ← silent drift
              │
         next production incident reveals it
```

`spectre sync` catches drift before it becomes a bug report.

---

### Running sync

```bash
spectre sync src/account.rs --spec account.spec
```

Each action gets one of three symbols:

| Symbol | Meaning |
|--------|---------|
| `[=]` | Unchanged — guards and assignments are structurally identical |
| `[≡]` | SMT-proved equivalent — code changed, semantics identical, no re-verify needed |
| `[✗]` | Semantically changed — SMT found a counterexample; re-verify required |

Example output after a bug is introduced in `withdraw`:

```
[=]  action deposit    (structurally unchanged)
[✗]  action withdraw   assignment 1 body changed:
       "balance' = balance - amount"  →  "balance' = balance + amount"
       witness: amount=1, balance=0
[=]  action freeze     (structurally unchanged)
[≡]  action unfreeze   guard rewritten (SMT-proved equivalent):
       "not frozen"  →  "frozen = false"
```

The `[✗]` line gives you:
- **What changed** — which assignment or guard
- **The witness** — concrete values where the two versions behave differently

The `[≡]` prevents false alarms on equivalent rewrites — commutativity, constant folding, guard commutation.

---

### CI integration

```yaml
# .github/workflows/spectre.yml
- name: Check for spec drift
  run: spectre sync src/account.rs --spec account.spec
  # exits non-zero if any [✗] is found
```

`sync` exits 0 if all actions are `[=]` or `[≡]`; exits non-zero on any `[✗]`. Safe to gate a merge on.

---

### Incremental re-verification

After a `[✗]`, you don't need to re-run full BFS. Re-verify only the changed action:

```bash
spectre verify account.spec \
  --incremental \
  --changed-action withdraw \
  --use-cache
```

How it works (5 steps):

1. Load the cached state graph from the last full run
2. Remove all states reachable only through `withdraw`
3. Re-explore from states where `withdraw` was previously enabled
4. Re-check invariants on the new states only
5. Merge result back into the cache

Output:

```
Incremental re-verification: action 'withdraw'
  States re-explored: 12  (vs 847 in full BFS)
  ✓ safe — no new violations
  Cache updated.
```

12 states vs 847 = **70× faster** for this example. The saving grows with spec size.

---

### Impact analysis

Before re-verifying, check which invariants are actually affected by the change:

```bash
spectre impact account.spec
```

```
Changed actions (from git diff):
  withdraw — semantics changed (SMT witness: amount=1, balance=0)

Affected invariants:
  solvent         ← balance can go negative with the new body
  not_over_limit  ← not directly affected by withdraw

Recommended:
  spectre verify account.spec --incremental --changed-action withdraw
```

Use `impact` in CI to skip re-verification when a changed action doesn't touch any safety-critical invariant.

---

### Full sync workflow

```
Code changes committed
        │
        ▼
spectre sync src/ --spec spec.spec
        │
  [=] or [≡] only?
        ├── yes ──► nothing to do ✓
        └── no ([✗] found)
                │
                ▼
        spectre impact spec.spec
                │
        safety-critical invariant affected?
                ├── no ──► log warning, skip re-verify
                └── yes
                        │
                        ▼
                spectre verify spec.spec \
                  --incremental \
                  --changed-action NAME \
                  --use-cache
                        │
                   safe? ──yes──► commit + update cache ✓
                        │
                        no
                        │
                        ▼
                Fix code or spec, re-verify
```

---

### Sync evaluation: what it checks

`sync` compares:
- Struct field names and types
- Action guard counts and expressions
- Assignment-body expressions (normalised)

It detected **30/35 injected semantic mutations** (85.7%) in evaluation, with 1 false positive (vs 4 with AST comparison alone). The `[≡]` classification eliminates the majority of false alarms.

---

### Tips for keeping sync useful

- **Check in the spec file alongside the Rust source** — keep them in the same directory or a `spec/` subdirectory
- **Run sync in pre-commit** as well as CI — catch drift before it even reaches the repo
- **After a `[✗]`**: always re-verify before merging, even if you think the change is safe
- **After a `[≡]`**: no action required — but it's worth reading the rewrite to confirm it's intentional
# Chapter 12: Model-based testing

### The idea

```
account.spec
      │
      ▼
spectre verify --emit-traces trace.itf.json
      │
      ▼
spectre generate-driver --lang rust account.spec -o src/driver.rs
      │
      ▼
fill in apply_action() + current_state() against your BankAccount impl
      │
      ▼
cargo test
      │
      ├── ✓ 6/6 steps match ── implementation agrees with spec
      └── ✗ step 3 mismatch  ── spec found a bug in your Rust code
```

The spec is the oracle. The driver replays spec-generated execution traces against your actual Rust implementation and reports any disagreement.

---

### Generate a trace

```bash
spectre verify account.spec \
  --emit-traces trace.itf.json \
  --coverage-mode property
```

Coverage modes — pick based on what you want to exercise:

| Mode | Strategy | Best for |
|------|----------|----------|
| `action` | Prefer actions not yet seen in trace | Breadth of action coverage (default) |
| `property` | Steer toward invariant boundary values | Catching off-by-one bugs |
| `transition-pair` | Cover all (prev, next) action pairs | Transition sequence coverage |
| `boundary` | Maximise integer variable values | Overflow and limit bugs |
| `rare-action` | Prefer globally least-executed actions | Obscure code paths |

`property` mode achieves 52% boundary-state coverage vs 5–25% for the others — use it when chasing subtle arithmetic bugs.

---

### Generate the driver

```bash
spectre generate-driver --lang rust account.spec -o src/driver.rs
```

`driver.rs` gives you two traits to implement:

```rust
pub trait AccountDriver {
    /// Apply one action from the spec trace to your implementation.
    fn apply_action(&mut self, action: &str, args: &[Value]) -> Result<(), String>;

    /// Return the current implementation state in spec terms.
    fn current_state(&self) -> SpecState;
}
```

`SpecState` mirrors the spec's `var` declarations:

```rust
pub struct SpecState {
    pub balance: i64,
    pub frozen:  bool,
}
```

Wire these to your `BankAccount` struct and you're done.

---

### CI-ready driver with `--ci`

```bash
spectre generate-driver --lang rust account.spec --ci -o src/driver.rs
```

Three files are written:

| File | What to do with it |
|------|-------------------|
| `src/driver.rs` | Implement the two trait methods |
| `src/driver.cargo.toml` | Paste the `[dev-dependencies]` block into `Cargo.toml` |
| `src/driver.ci.yml` | Copy to `.github/workflows/` — runs `cargo test --test driver` on every push |

The Cargo snippet:

```toml
# paste into your [dev-dependencies]
spectre-connect = { git = "https://github.com/BoundlessProducts/spectre-connect" }
```

The CI workflow:

```yaml
name: Spectre conformance
on: [push, pull_request]
jobs:
  conformance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: rust-toolchain@stable
      - run: cargo test --test driver
```

No extra configuration — the workflow picks up `trace.itf.json` from the repo.

---

### Run the replay

```bash
cargo run -- trace.itf.json
```

Passing run:

```
Replaying 6 steps...
  step 1  deposit(100):   balance  0 → 100  ✓
  step 2  deposit(100):   balance 100 → 200  ✓
  step 3  withdraw(50):   balance 200 → 150  ✓
  step 4  freeze:         frozen false → true  ✓
  step 5  deposit(100):   skipped (frozen)  ✓
  step 6  unfreeze:       frozen true → false  ✓
✓ 6/6 steps match
```

Failing run:

```
✗ 1 mismatch:
  step 3, action 'withdraw', var 'balance':
    spec expects: 150
    impl returned: 200
    hint: check withdraw() arithmetic — possible sign error
```

---

### Violation traces as regression tests

When `spectre verify` finds a violation, `--emit-traces` captures the exact counterexample path:

```bash
spectre verify account.spec --emit-traces violation.itf.json
```

Check this file into your repo. It should:
- **Fail** against the unfixed implementation
- **Pass** after you apply the fix

```bash
# Before fix:
cargo run -- violation.itf.json  # → ✗ mismatch at step 4

# After fix:
cargo run -- violation.itf.json  # → ✓ 4/4 steps match
```

This turns every spec violation into a permanent, reproducible regression test.

---

### Repair suggestions before fixing

When verification fails, get candidate patches before editing anything:

```bash
spectre verify account.spec --suggest-repairs
```

```
Repair candidates (CANDIDATE — review before applying):

Violation 1: action 'withdraw' violates 'solvent'
  [1] CANDIDATE strengthen_require: add `require balance >= amount` to withdraw
      Validation: violation fixed · no deadlock · reachability unchanged — SAFE
  [2] CANDIDATE add_invariant_guard: add `when balance >= amount` to withdraw
      Validation: violation fixed · deadlock at balance=0 — review required
```

How to pick:

| Candidate type | Pick when | Risk |
|---------------|-----------|------|
| `strengthen_require` | The action should simply not fire when condition fails | Minimal |
| `add_invariant_guard` | You want the action blocked at the invariant boundary | Can introduce deadlock |

Apply the chosen patch to the spec, re-verify, then propagate the change to the Rust source.

---

### Checking refinement

When you have traces from an external source (a different tool, a recorded production run), check whether they conform to your spec:

```bash
spectre check-refinement account.spec
```

This reads `driver.json` (the refinement mapping) and validates that each external trace is a valid execution of the spec.

---

### Complete MBT workflow

```
1.  spectre mine --lang rust src/account.rs -o account.spec
          │
2.  Add invariants + temporal properties to account.spec
          │
3.  spectre verify account.spec
          ├── violation? → --suggest-repairs → fix → repeat
          └── safe
                │
4.  spectre verify account.spec \
      --emit-traces trace.itf.json \
      --coverage-mode property
          │
5.  spectre generate-driver --lang rust account.spec --ci -o src/driver.rs
          │
6.  Implement apply_action() and current_state() in driver.rs
          │
7.  cargo test --test driver
          ├── mismatch? → read step output → fix Rust impl → repeat from 7
          └── ✓ all pass
                │
8.  Check in: account.spec, trace.itf.json, driver.rs, driver.ci.yml
          │
9.  On every code change:
      spectre sync src/account.rs --spec account.spec
          ├── [✗] found? → re-verify → update trace → repeat from 7
          └── [=]/[≡] → nothing to do
```
# Chapter 13: Runtime monitoring

Formal verification proves your spec is safe. Runtime monitoring proves your *running application* stays inside that spec — continuously, in production.

---

## The gap formal verification leaves

```
spectre verify account.spec
        │
   ✓ safe — all reachable states satisfy invariants
        │
   ship to production
        │
   third-party library update changes behaviour ← not in your spec
        │
   invariant violated at runtime — never detected
```

A runtime monitor closes this gap. It checks your spec's invariants on every real state transition — with no heap allocation in the common path.

---

## Generate a monitor

```bash
spectre generate-monitor --lang rust account.spec -o src/monitor.rs
```

The generated `monitor.rs` gives you:

```rust
/// Mirrors the spec variables.
#[derive(Debug, Clone)]
pub struct SpecState {
    pub balance: i64,
    pub frozen:  bool,
}

/// A spec violation detected at runtime.
#[derive(Debug)]
pub struct Violation {
    pub step:      u64,
    pub invariant: &'static str,
    pub state:     SpecState,
}

pub struct Monitor { /* private */ }

impl Monitor {
    pub fn new() -> Self { Monitor::default() }

    /// Call after every state transition in your application.
    pub fn step(&mut self, state: &SpecState) -> Option<Violation> { ... }

    /// Check liveness properties periodically (e.g. every 1000 steps).
    pub fn check_liveness(&self) -> Vec<&'static str> { ... }
}
```

---

## Wire it into your application

```rust
use crate::monitor::{Monitor, SpecState};

struct App {
    account: BankAccount,
    monitor: Monitor,
}

impl App {
    fn deposit(&mut self, amount: i64) {
        self.account.deposit(amount);

        let state = SpecState {
            balance: self.account.balance,
            frozen:  self.account.frozen,
        };

        if let Some(v) = self.monitor.step(&state) {
            tracing::error!(
                "spec violation: {} at step {}",
                v.invariant, v.step
            );
            // handle: log, alert, rollback, or panic — your choice
        }
    }
}
```

The monitor is:
- **Zero dependencies** — no external crate, no allocations in the happy path
- **O(k) per transition** where k = number of invariants
- **Consistent** — generated from the exact same `.spec` used for model checking

---

## Handle violations your way

| Response | When to use |
|----------|-------------|
| Log + continue | Non-critical invariant, high-traffic path |
| Alert + continue | Production monitoring with async alerting |
| Panic | Safety-critical — prefer crash over corrupt state |

---

## Validate monitor traces before deploying

```bash
spectre validate-monitor account.spec
```

```
Validated 2 violation trace(s):
  [1] action 'withdraw' → invariant 'solvent': consistent (4 steps, final violated)
  [2] action 'deposit'  → invariant 'within_cap': consistent (3 steps, final violated)

2/2 consistent.
```

A **consistent** trace means the BFS path genuinely leads to (or one step before) a state that fails the target invariant. If a trace is inconsistent, the violation recording has a bug — not your application.

Run `validate-monitor` after every spec change before regenerating the monitor.

Flags:

| Flag | Default | Effect |
|------|---------|--------|
| `--max-states` | 10000 | BFS exploration limit |
| `--max-depth` | 100 | BFS depth limit |

---

## What the monitor checks — and what it doesn't

**Checks:**
- Every `invariant` from the spec, after every `step()` call
- Liveness properties when you call `check_liveness()`

**Doesn't check:**
- Whether your `SpecState` mapping is correct — that's on you
- Temporal properties in real-time (sampled only on `check_liveness`)
- Invariants over variables you didn't include in `SpecState`

---

## Full monitoring workflow

```
account.spec
     │
     ▼
spectre verify account.spec
     │  ✓ safe — design-time proof
     ▼
spectre validate-monitor account.spec
     │  ✓ all violation traces consistent
     ▼
spectre generate-monitor --lang rust account.spec -o src/monitor.rs
     │
     ▼
implement SpecState mapping in application
     │
     ▼
call monitor.step(&state) after every transition
     │
     ▼
deploy
     │
     ├── violation fires ──► log / alert / panic per policy
     └── no violation    ──► ✓ spec holds in production
```

---

## Performance overhead

For `account.spec` with two invariants, each `step()` call is two integer comparisons. On modern hardware:

```
~2 ns per step (two invariants, no allocation)
```

At 100,000 transitions per second that's **0.02% CPU overhead** — negligible in any real workload.

---

## The consistency guarantee

Because the monitor is generated from the same `.spec` file:
- If your spec passes `spectre verify` ✓
- And the monitor never fires in production

...you have end-to-end assurance: your implementation matches the design across every observed execution.
# Chapter 14: Benchmarking and performance

Before you change a spec, you want to know how big the state space is and how fast Spectre explores it. That's what the built-in benchmark suite is for.

---

## Run the benchmarks

```bash
spectre benchmark
```

```
Benchmark results (BFS backend)
================================
counter.spec        10 states   0.12 ms    depth 5
account.spec        84 states   0.41 ms    depth 8
mutex.spec         430 states   2.1 ms     depth 14
dist_queue.spec   3210 states   18 ms      depth 24
ring_leader.spec  9980 states   94 ms      depth 40

Total: 5 specs, 13714 states, 115 ms
```

The benchmark runs your spec files in `benchmarks/` against the BFS backend and reports state count, wall time, and max depth.

Add your own specs:

```bash
# drop any .spec file into benchmarks/
cp account.spec benchmarks/

spectre benchmark
```

---

## Understand state-space growth

State spaces grow exponentially with the number of variables and their ranges. Here's what to expect:

| Pattern | Typical state count | Why |
|---------|-------------------|-----|
| Counter (1 var, range 0–10) | ~11 | Linear |
| 2 independent counters (0–10 each) | ~121 | Product |
| Mutex (2 threads, 3 states each) | ~9 | Small product |
| Mutex + shared counter (0–10) | ~90 | 3-way product |
| Producer-consumer (queue depth 5) | ~500 | Queue contents combinatorial |
| Ring leader election (5 nodes) | ~100K+ | Permutation space |

The rule: **every new variable multiplies the state space by its domain size**.

---

## State-space growth diagram

```
Variables:  1     2     3     4     5
            │     │     │     │     │
Range 10:  10   100  1000  10K  100K
Range 5:    5    25   125  625   3K
Range 2:    2     4     8   16   32
            │
            └── BFS exhausts this in < 1 ms
                              │
                              └── starts taking seconds
                                          │
                                          └── needs --backend tlc or z3
```

---

## Bounded exploration options

When a spec is too large for exhaustive BFS, cap the search:

```bash
# Cap at 50,000 states
spectre verify --max-states 50000 ring_leader.spec

# Cap at depth 30
spectre verify --max-depth 30 ring_leader.spec

# Both
spectre verify --max-states 50000 --max-depth 30 ring_leader.spec
```

The result will say `(bounded)` rather than `(exhaustive)`.

---

## Partial-order reduction

For concurrent systems, many interleavings produce identical states. `--por` (partial-order reduction) prunes these by exploring only one representative ordering:

```bash
spectre verify --por mutex.spec
```

```
Before --por: 430 states, 2.1 ms
After  --por: 127 states, 0.7 ms
```

Savings are largest when:
- Many threads share little state
- Actions are mostly independent

`--por` is sound for safety properties. Avoid it for liveness properties unless you also have `--backend tlc`.

---

## Property-guided exploration

```bash
spectre verify --property-guided mutex.spec
```

Property-guided mode biases BFS toward states that are *close to violating* an invariant. It finds bugs faster on large specs but may miss violations in rarely-reached corners.

Use it for:
- Interactive debugging — "show me a violation fast"
- Large specs where exhaustive BFS times out

Don't use it for production sign-off — use exhaustive BFS or TLC for that.

---

## When to switch backends

| Situation | Recommendation |
|-----------|----------------|
| ≤ 100K states, safety only | BFS (default) |
| Fairness / liveness properties | `--backend tlc` |
| > 1M states, safety only | `--backend apalache` (symbolic) |
| Want shallow bugs fast (< 20 steps) | `--backend z3 --bmc-depth 20` |
| Need exhaustive + fairness + huge space | `--backend tlc` with TLC's -workers flag |

---

## Progressive verification strategy

Start cheap, escalate only if needed:

```
spectre verify --max-states 10000 my.spec
        │
   ✓ safe in 10K states ──► probably fine for small specs
        │
   ✗ violation or exhausted
        │
        ▼
spectre verify my.spec                         (exhaustive BFS, no limit)
        │
   ✓ exhaustive safe ──► ship
        │
   too large / times out
        │
        ▼
spectre verify --backend z3 --bmc-depth 50 my.spec   (fast shallow bugs)
        │
   sat → violation in ≤50 steps ──► fix, restart
        │
   unsat (no shallow bugs)
        │
        ▼
spectre verify --backend tlc my.spec                 (full model check)
```

Each step is significantly more expensive. Stop as soon as you get a definitive answer.

---

## Benchmark in CI

```yaml
# .github/workflows/ci.yml (excerpt)
- name: Benchmark smoke test
  run: |
    spectre benchmark --max-states 1000 --ci
    # Fails if any spec exceeds 1000 states (catches accidental bloat)
```

The `--ci` flag exits non-zero if any spec exceeds `--max-states`. This catches spec bloat before it becomes a slow-CI problem.
# Chapter 15: TLA+ and Apalache backend

Spectre's BFS backend is great for small to medium specs. When you need validated fairness semantics, or your state space is too large for BFS, the `--backend tlc` and `--backend apalache` options hand off to proven model checkers built on TLA+.

---

## Quick start

```bash
# Run TLC model checker
spectre verify --backend tlc counter.spec

# Run Apalache symbolic model checker
spectre verify --backend apalache counter.spec

# Point to non-default TLC jar
spectre verify --backend tlc --tlc-jar /opt/tlc/tla2tools.jar counter.spec

# Point to non-default Apalache binary
spectre verify --backend apalache --apalache-bin /usr/local/bin/apalache-mc counter.spec
```

---

## How Spectre translates to TLA+

Spectre translates your `.spec` file into a standard TLA+ module plus a `.cfg` file. Here's the counter example:

**`counter.spec`**
```
var x: int
param MAX: int
init { x = 0 }
action incr { require x < MAX  x' = x + 1 }
invariant non_negative { x >= 0 }
temporal wf_progress { WF(incr) }
```

**Generated `counter.tla`**
```tla
--------------------------- MODULE counter ---------------------------
EXTENDS Naturals, Sequences, TLC

CONSTANTS MAX

VARIABLES x

TypeOK == x \in Int

Init ==
    x = 0

incr ==
    /\ x < MAX
    /\ x' = x + 1

Next ==
    \/ incr

non_negative ==
    x >= 0

Spec ==
    Init /\ [][Next]_<<x>> /\ WF_<<x>>(incr)

=============================================================================
```

**Generated `counter.cfg`**
```
INIT Init
NEXT Next
INVARIANT non_negative
PROPERTY Spec
MAX = 100
```

Then Spectre runs `tlc counter.tla -config counter.cfg` and parses the output.

---

## Expression mapping reference

| Spectre | TLA+ |
|---------|------|
| `x != y` | `x # y` |
| `x <= y` | `x =< y` |
| `x and y` | `x /\ y` |
| `x or y` | `x \/ y` |
| `not x` | `~x` |
| `true` | `TRUE` |
| `false` | `FALSE` |
| `always(P)` | `[](P)` |
| `eventually(P)` | `<>(P)` |
| `WF(action)` | `WF_vars(action)` in `Spec` |
| `SF(action)` | `SF_vars(action)` in `Spec` |
| `x'` | `x'` |
| `UNCHANGED x, y` | `UNCHANGED <<x, y>>` |

---

## Why TLC for fairness?

BFS can detect some fairness violations, but it uses a heuristic: it flags executions where an enabled action is never taken. This can produce false positives.

TLC's approach is exact:

```
Liveness check (4 steps):

1. Decompose Spec into safety + fairness constraints
   Spec == Init /\ [][Next]_vars /\ WF_vars(incr)
                                   └── fairness constraint

2. Find all SCCs in the state graph
   (strongly connected components = infinite loops)

3. For each SCC, check: does incr get to fire?
   If incr is enabled in some state of an SCC but never fires
   → genuine fairness violation

4. Report: specific cycle that violates WF(incr)
```

This is why `spectre verify --backend tlc` gives you authoritative liveness results where BFS only gives approximations.

---

## TLC vs Apalache: when to use which

| | TLC | Apalache |
|---|-----|----------|
| **Approach** | Explicit-state enumeration | Symbolic (SMT-based) |
| **Best for** | Small–medium state spaces, liveness | Large state spaces, safety |
| **Fairness** | Yes — exact SCC-based | Limited |
| **State space limit** | ~10M states practical | Effectively unbounded |
| **Speed on small specs** | Fast | Slower (SMT overhead) |
| **Speed on large specs** | Slow (memory) | Fast (symbolic) |
| **Counterexample** | Full trace | Symbolic witness |
| **Requires Java** | Yes | Yes |

Rule of thumb:
- Use **TLC** when you have `temporal` or `WF`/`SF` properties
- Use **Apalache** when BFS times out on safety properties

---

## Pass parameters to TLC

```bash
spectre verify --backend tlc --param MAX=100 --param THREADS=4 mutex.spec
```

Parameters become TLC `CONSTANTS` assignments in the `.cfg` file:

```
CONSTANTS
MAX = 100
THREADS = 4
```

---

## Read TLC output

When TLC finds a violation, Spectre parses the trace and prints it in the same format as BFS:

```
mutex.spec: ✗ violation found
  Violation 1: invariant "mutual_exclusion" triggered by action "enter_cs"
    Step 0: [pc_a = idle,    pc_b = idle,    flag_a = false, flag_b = false]
    Step 1: [pc_a = wanting, pc_b = idle,    flag_a = true,  flag_b = false]
    Step 2: [pc_a = wanting, pc_b = wanting, flag_a = true,  flag_b = true ]
    Step 3: [pc_a = critical,pc_b = wanting, flag_a = true,  flag_b = true ]
    Step 4: [pc_a = critical,pc_b = critical,flag_a = true,  flag_b = true ]
                                              ↑ both in critical section — violated
```

---

## TLC installation

TLC comes bundled in the `tla2tools.jar` distributed by the TLA+ project:

```bash
# macOS (Homebrew)
brew install tla-tools

# Or download tla2tools.jar manually and point Spectre at it:
spectre verify --backend tlc --tlc-jar ~/tools/tla2tools.jar account.spec
```

Spectre tries `tlc` on `$PATH` first, then falls back to `java -jar <tlc-jar>`.

---

## Apalache installation

```bash
# macOS
brew install apalache

# Or set the binary path:
spectre verify --backend apalache --apalache-bin ~/tools/apalache/bin/apalache-mc account.spec
```
# Chapter 16: Z3 bounded model checking

The Z3 backend is your fastest path to finding shallow bugs. Instead of exploring the state graph step by step, it encodes your spec as an SMT-LIB2 formula and asks Z3 a single question: *is there any execution of length ≤ N that violates an invariant?*

---

## Quick start

```bash
# Check for violations in up to 20 steps (default depth)
spectre verify --backend z3 counter.spec

# Increase depth for longer paths
spectre verify --backend z3 --bmc-depth 50 account.spec

# Point to non-default Z3 binary
spectre verify --backend z3 --z3-bin /usr/local/bin/z3 account.spec
```

---

## How bounded model checking works

```
                        counter.spec
                             │
                  ┌──────────▼──────────┐
                  │  Z3 encoder         │
                  │  for depth = 3      │
                  └──────────┬──────────┘
                             │
              ┌──────────────▼──────────────┐
              │  SMT-LIB2 formula            │
              │                             │
              │  x_0 = 0             (init) │
              │  x_1 = incr(x_0) ∨         │
              │  x_1 = decr(x_0)    (step) │
              │  x_2 = incr(x_1) ∨ ...     │
              │  x_3 = incr(x_2) ∨ ...     │
              │                             │
              │  ¬(x_0≥0 ∧ x_1≥0           │
              │    ∧ x_2≥0 ∧ x_3≥0)        │
              │      (invariant negation)   │
              └──────────────┬──────────────┘
                             │
                    ┌────────▼────────┐
                    │   Z3 solver     │
                    └────────┬────────┘
                             │
               ┌─────────────┴──────────────┐
               │                            │
             unsat                         sat
               │                            │
      no violation in ≤ 3 steps    violation exists
      (but may exist deeper)        extract model
                                    print trace
```

---

## SMT-LIB2 encoding example

For this spec:

```
var x: int
init { x = 0 }
action incr { require x < 5  x' = x + 1 }
action decr { require x > 0  x' = x - 1 }
invariant non_negative { x >= 0 }
```

Spectre generates (simplified):

```smt2
; Step-indexed variables
(declare-const x_0 Int)
(declare-const x_1 Int)
(declare-const x_2 Int)

; Initial state
(assert (= x_0 0))

; Step 0 → 1: one action fires
(assert (or
  (and (< x_0 5) (= x_1 (+ x_0 1)))   ; incr
  (and (> x_0 0) (= x_1 (- x_0 1)))   ; decr
))

; Step 1 → 2
(assert (or
  (and (< x_1 5) (= x_2 (+ x_1 1)))
  (and (> x_1 0) (= x_2 (- x_1 1)))
))

; Invariant negation: find ANY step that violates non_negative
(assert (or
  (not (>= x_0 0))
  (not (>= x_1 0))
  (not (>= x_2 0))
))

(check-sat)
(get-model)
```

Z3 answers `unsat` (no violation in 2 steps) or `sat` (violation found, model gives the path).

---

## Read Z3 output

**No violation found (up to depth):**
```
counter.spec: ✓ safe (z3, bounded — depth 20)
  Note: no violation in ≤20 steps; deeper paths not checked
```

**Violation found:**
```
account.spec: ✗ violation found (z3)
  Violation 1: invariant "solvent" triggered by action "withdraw"
    x_0 = balance:100
    x_1 = balance:50  (deposit 50 — now 100)
    x_2 = balance:-10 (withdraw 60 — violated)
```

---

## When to use Z3 BMC

**Good fit:**
- You want to find shallow bugs fast during development
- Your state space is too large for exhaustive BFS
- You have many variables but expect violations within ~50 steps
- You want a quick sanity check before a full TLC run

**Poor fit:**
- You need an exhaustive proof of safety (Z3 BMC can only say "no bug in N steps")
- You have liveness or fairness properties (Z3 BMC checks only reachability)
- Violations require very long paths (rare in practice — most bugs appear early)

---

## Choose the right depth

| Scenario | Suggested `--bmc-depth` |
|----------|------------------------|
| Quick sanity check | 10–20 |
| Standard development | 20–50 |
| Protocol with long setup phase | 50–100 |
| Deep corner cases | 100–200 |
| > 200 | Consider TLC/Apalache instead |

Doubling the depth roughly doubles the formula size and Z3 solve time. Past depth 200 you're usually better off with TLC or Apalache.

---

## Z3 vs other backends

| | BFS | Z3 BMC | TLC | Apalache |
|---|-----|--------|-----|----------|
| **Exhaustive** | Yes (small) | No (bounded) | Yes | Yes (symbolic) |
| **Fairness** | Approximate | No | Exact | No |
| **Speed on shallow bugs** | Fast | Fastest | Slow | Medium |
| **Large state space** | No | Yes | No | Yes |
| **Counterexample** | Full path | Full path (≤N) | Full path | Symbolic |
| **External dependency** | None | z3 binary | Java+TLC | Java+Apalache |

---

## Progressive verification with Z3

Z3 BMC fits naturally into a staged workflow:

```
spectre verify --backend z3 --bmc-depth 20 my.spec
        │
   unsat (no bug ≤ 20 steps)
        │
        ▼
spectre verify --backend z3 --bmc-depth 100 my.spec
        │
   unsat (no bug ≤ 100 steps)
        │
        ▼
spectre verify my.spec   (exhaustive BFS if state space allows)
        │
        ▼
spectre verify --backend tlc my.spec   (for liveness / large space)
```

Each stage gives you increasing confidence. Stop as soon as you get a definitive result or a violation to fix.

---

## Z3 installation

```bash
# macOS (Homebrew)
brew install z3

# Verify
z3 --version
# Z3 version 4.13.x

# Point Spectre at a non-default binary
spectre verify --backend z3 --z3-bin /opt/z3/bin/z3 account.spec
```

Spectre looks for `z3` on `$PATH` by default.
