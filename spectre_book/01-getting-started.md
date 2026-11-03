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
