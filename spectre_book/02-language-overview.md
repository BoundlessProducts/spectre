# The Spectre language

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

## Collections

Spectre supports first-class collection types for modelling sets of nodes, queues, and mappings:

| Type | Declaration | Example |
|------|-------------|---------|
| `Set<T>` | `var nodes: Set<int>` | `{1, 2, 3}` |
| `Map<K,V>` | `var votes: Map<int, bool>` | `{}` |
| `List<T>` | `var queue: List<int>` | `{}` |

Common collection methods:

| Method | Works on | Returns |
|--------|----------|---------|
| `.union(other)` | Set | Set — combined elements |
| `.intersection(other)` | Set | Set — common elements |
| `.contains(x)` | Set, Map | bool |
| `.size()` | Set, Map, List | int |
| `.put(k, v)` | Map | Map — with key set |
| `.get(k)` | Map | value |
| `.append(x)` | List | List — with element added |
| `.filter(fn)` | Set, List | filtered collection |
| `.forall(fn)` | Set, List | bool — all match |
| `.exists(fn)` | Set, List | bool — any match |

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

init { signal = TrafficLight.Red }

action go     { require signal == TrafficLight.Red     signal' = TrafficLight.Green  }
action caution { require signal == TrafficLight.Green  signal' = TrafficLight.Yellow }
action stop   { require signal == TrafficLight.Yellow  signal' = TrafficLight.Red    }

invariant only_valid_transitions {
  signal == TrafficLight.Red || signal == TrafficLight.Yellow || signal == TrafficLight.Green
}
```

```bash
spectre verify traffic.spec
# Exhaustive verification: traversed 3 states (full reachable space)
# Found no violations.
```

---

## Parameters

Constants are named values that don't change during a run:

```spectre
const CAPACITY:  int = 10
const THRESHOLD: int = 7
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
  phase   = Phase.Idle
}
```

- Every variable you list is set to that value.
- Variables not listed get their type's zero value: `0`, `false`, `""`, first enum variant.
- You can list variables in any order.

---

## Actions

An action describes one atomic transition of the system:

```
action deposit {           <- name
  require status == Status.Active  <- precondition (action disabled if false)
  require amount > 0               <- multiple requires are AND'd together
  balance' = balance + amount      <- next-state value (prime notation)
  count'   = count + 1             <- all assignments apply simultaneously
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
const MAX: int = 10

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
const LIMIT: int = 1000

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
  require phase == Phase.Idle
  phase'   = Phase.Running
  started' = true
}

action finish {
  require phase == Phase.Running
  phase' = Phase.Done
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
const LIMIT: int = 1000

init {
  balance = 0
  status  = Status.Active
}

action deposit {
  require status == Status.Active
  require balance + 100 <= LIMIT
  balance' = balance + 100
}

action withdraw {
  require status == Status.Active
  require balance >= 100
  balance' = balance - 100
}

action freeze   { status' = Status.Frozen }
action unfreeze { status' = Status.Active }

invariant solvent     { balance >= 0 }
invariant within_cap  { balance <= LIMIT }
```

```bash
spectre verify bank.spec
```

```
Exhaustive verification: traversed 22 states (full reachable space)
Found no violations.
```

---

## Temporal properties

Temporal properties let you express liveness — things that must *eventually* happen, or *always* remain true over infinite runs:

```spectre
temporal always_solvent  { always(balance >= 0) }
temporal progress        { eventually(balance > 0) }
temporal freeze_recovers { always(status == Status.Frozen -> eventually(status == Status.Active)) }
```

| Keyword | Meaning |
|---------|---------|
| `always(P)` | P holds in every state of every run |
| `eventually(P)` | P holds in some state of every run |
| `P until Q` | P holds until Q becomes true |
| `WF(action)` | Weak fairness: if `action` is always enabled, it eventually fires |
| `SF(action)` | Strong fairness: if `action` is infinitely often enabled, it fires infinitely often |

BFS is the learning backend — great for exploring what your spec does. For **fairness** (WF/SF) and liveness properties you need to trust, use `--backend tlc` or `--backend apalache`. The Fairness chapter covers this in detail.

---

## Descriptions

Add a human-readable description to any action:

```spectre
description "Deposits 100 units — requires account active and not at cap"
action deposit {
  require status == Status.Active
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
| `const` | Declare a named constant (override with `--param`) |
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

- **Safety Properties chapter** — safety properties in depth: writing good invariants, reading violation traces
- **Temporal Properties chapter** — temporal properties: liveness and fairness with real examples
