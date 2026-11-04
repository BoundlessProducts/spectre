# Getting started

---

## Install

```bash
# macOS / Linux — build from source
git clone https://github.com/BoundlessProducts/spectre
cd spectre
go build -o spectre ./cmd/spectre
sudo mv spectre /usr/local/bin/

spectre --version
# spectre version dev
```

No runtime dependencies, no Docker, no JVM needed for the default BFS backend.

---

## Your first spec

Save this as `counter.spec`:

```spectre
var value: int
const MAX: int = 5

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
Exhaustive verification: traversed 6 states (full reachable space)
Found no violations.
```

---

## What just happened

```
counter.spec
  ├── var value: int              ← one piece of state
  ├── const MAX: int = 5          ← a named constant
  ├── init { value = 0 }          ← where to start
  ├── action increment { ... }    ← one possible transition
  ├── action reset { ... }        ← another transition
  └── invariant within_bounds ... ← the property to check
          │
          v
    spectre verify
          │
    BFS explores all 6 reachable states
    {0, 1, 2, 3, 4, 5}
          │
    checks both invariants in every state
          │
          v
        (ok) safe
```

Spectre uses **breadth-first search (BFS)** as its default learning and exploration backend — zero setup, instant feedback, great for writing and debugging specs. It finds every state reachable from `init` and checks your invariants in all of them. Six states here because `value` can be `0`–`5` and `reset` can always bring it back.

BFS is perfect for learning Spectre and iterating quickly. For production liveness verification or large state spaces, use `--backend tlc` or `--backend apalache` (see the Backend Selection chapter).

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
Violations:

1.   Invariant violated: within_bounds
  Action 'increment' would violate invariants: [invariant within_bounds violated]
  Counterexample trace (5 step(s)):
    [init]
    [step 1] increment
      value: 0 -> 1
    [step 2] increment
      value: 1 -> 2
    [step 3] increment
      value: 2 -> 3
    [step 4] increment
      value: 3 -> 4
    [step 5] increment
      value: 4 -> 5
    [VIOLATION] invariant "within_bounds" fails in this state
      value = 5
```

Spectre shows the state where the *next* `increment` would violate `within_bounds` (value 5 + 1 = 6 > MAX). Add back `require value < MAX` to fix it.

Spectre found the shortest path to a bad state: six increments from zero.

---

## Fix it

Restore `require value < MAX`. This tells Spectre: "increment can only fire when `value < MAX`." If the condition is false the action is disabled — not an error, the system just can't take that step from that state.

---

## Key commands

| Command | What it does |
|---------|-------------|
| `spectre verify <file>` | BFS — exhaustively checks all reachable states (learning/exploration mode) |
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
          v
  +-------------------------------------------------------+
  | Learning / exploring a new spec?                      |
  |   -> spectre verify  (default BFS, zero setup)        |
  +-------------------------------------------------------+
  | Find shallow bugs quickly (<=20 steps)?               |
  |   -> --backend z3 --bmc-depth 20                      |
  +-------------------------------------------------------+
  | Exhaustive proof on a small-medium spec?              |
  |   -> spectre verify  (default BFS)                    |
  +-------------------------------------------------------+
  | Liveness / fairness properties you can trust?         |
  |   -> --backend tlc  or  --backend apalache            |
  +-------------------------------------------------------+
  | Spec too large for BFS (millions of states)?          |
  |   -> --backend apalache  (symbolic, handles large)    |
  +-------------------------------------------------------+
```

The Backend Selection chapter covers the full decision guide.

---

## The full Spectre workflow

```
Rust source
    │
    v
spectre mine --lang rust src/account.rs
    │
    v
account.spec  <- skeleton; add invariants by hand
    │
    v
spectre verify account.spec
    │
    +-- (ok) safe -----------------> commit .spec alongside source
    │
    +-- (x) violation
              │
              v
         spectre verify account.spec --suggest-repairs
              │
              v
         CANDIDATE patches  <- review, apply, re-verify
```

Once safe, one spec drives three downstream workflows:

```
account.spec
    │
    +---> spectre generate-driver  ->  Rust conformance test harness  (Model-Based Testing chapter)
    +---> spectre generate-monitor ->  embedded runtime monitor       (Runtime Monitoring chapter)
    +---> spectre sync account.rs  ->  drift detection as code evolves (Drift and Sync chapter)
```

---

## What's next

- **The Spectre Language chapter** — every language construct with examples
- **Safety Properties chapter** — safety invariants in depth, reading violations, repair suggestions
- **Backend Selection chapter** — picking the right backend for your problem
