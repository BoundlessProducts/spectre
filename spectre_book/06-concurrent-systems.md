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
