# Concurrent and distributed systems

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

invariant mutex { !(p1_in_cs && p2_in_cs) }
```

```bash
spectre verify mutex-broken.spec
# Violations:
# 1. Invariant violated: mutex
#    Action 'p1_enter' would violate invariants: [invariant mutex violated]
#    Counterexample trace (1 step(s)):
#      [init]
#      [step 1] p2_enter: p2_in_cs: false -> true
#      [VIOLATION] invariant "mutex" fails in this state
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

action p1_enter { require !lock    p1_in_cs' = true   lock' = true  }
action p1_exit  { require p1_in_cs  p1_in_cs' = false  lock' = false }
action p2_enter { require !lock    p2_in_cs' = true   lock' = true  }
action p2_exit  { require p2_in_cs  p2_in_cs' = false  lock' = false }

invariant mutex      { !(p1_in_cs && p2_in_cs) }
invariant lock_owner { !lock || p1_in_cs || p2_in_cs }
temporal  p1_progress { WF(p1_enter) }
temporal  p2_progress { WF(p2_enter) }
```

```bash
spectre verify mutex.spec
# Exhaustive verification: traversed 3 states (full reachable space)
# Found no violations.
```

State machine for this model:

```
           p1_enter                    p1_exit
  (free) ------------> (p1 holds lock) ----------> (free)
    │                                                 ^
    │     p2_enter                    p2_exit         │
    +---------------- > (p2 holds lock) --------------+
```

Only one process can hold the lock — `p2_enter` and `p1_enter` both `require !lock`, so they can never both fire from the same state.

---

## Producer-consumer with a bounded buffer

```spectre
const CAPACITY: int = 3
var items: int

init { items = 0 }

action produce { require items < CAPACITY  items' = items + 1 }
action consume { require items > 0         items' = items - 1 }

invariant no_overflow  { items <= CAPACITY }
invariant no_underflow { items >= 0        }
temporal  eventually_full { eventually(items == CAPACITY) }
```

```bash
spectre verify buffer.spec
# Exhaustive verification: traversed 4 states (full reachable space)
# Found no violations.

spectre verify buffer.spec --param CAPACITY=10
# Exhaustive verification: traversed 11 states (full reachable space)
```

Adding a second producer (`produce2`) immediately doubles the state space — BFS handles it automatically.

---

## Distributed message queue

Three-role system: producers enqueue, consumers dequeue, an accounting invariant keeps them honest.

```spectre
const MAX_QUEUE: int = 5
var queue_len: int
var produced:  int
var consumed:  int

init { queue_len = 0  produced = 0  consumed = 0 }

action enqueue {
  require queue_len < MAX_QUEUE
  require produced < MAX_QUEUE
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
invariant accounting     { produced == consumed + queue_len }
```

```bash
spectre verify queue.spec
# Exhaustive verification: traversed 21 states (full reachable space)
# Found no violations.
```

The `accounting` invariant is the most valuable one here. It catches any race where messages are created or destroyed without both counters updating — a class of bugs that integration tests rarely find.

To add a liveness property (queue must eventually drain), use `--backend tlc` since `always(P -> eventually(Q))` requires SCC-based checking:

```bash
spectre verify queue.spec --backend tlc
# -> verifies always(queue_len > 0 -> eventually(queue_len == 0)) with WF(dequeue)
```

---

## Good invariants for concurrent specs

| Pattern | Invariant example |
|---------|------------------|
| Mutual exclusion | `!(p1_in_cs && p2_in_cs)` |
| Accounting identity | `produced == consumed + in_flight` |
| State consistency | `!locked || owner_known` |
| Bounded resource | `active_connections <= MAX_CONN` |
| No phantom state | `!waiting || (request_sent && !response_received)` |

Write at least one invariant of each type that applies to your system.

---

## When the state space gets large

Three-process or three-role specs can hit tens of thousands of states:

```bash
spectre verify 3-party.spec --max-states 50000
# bounded -- hit state limit (50000 states)
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
# uses Apalache's symbolic engine — see the TLA+ and Apalache chapter
```

---

## Concurrency workflow

```
Identify shared variables
         │
         v
Write one action per actor per operation
         │
         v
Add require guards for preconditions
         │
         v
Write mutex + accounting invariants
         │
         v
spectre verify  (BFS — learning mode)
         │
    violation? --yes--> read the counterexample trace
         │                       │
         no                      v
         │              strengthen requires or add lock variable
         v
Add WF/SF temporal properties
         │
         v
spectre verify --backend tlc  (for validated fairness)
```
