# Reading counterexamples and debugging

When verification fails, Spectre gives you the shortest possible path to the violation. Here's how to read it and fix the spec.

---

## Anatomy of a violation trace

```
Violations:

1.   Invariant violated: solvent              <- which invariant failed
  Counterexample trace (4 step(s)):           <- BFS shortest path
    [init]                                    <- starting state
    [step 1] deposit:  balance: 0 -> 100       <- action + what changed
    [step 2] deposit:  balance: 100 -> 200
    [step 3] withdraw: balance: 200 -> 200     <- no change (wrong action)
    [step 4] withdraw: balance: 200 -> -100
    [VIOLATION] invariant "solvent" fails in this state
      balance = -100                          <- concrete value that broke it
      status = Active
```

Every part of the output tells you something:

| Part | What it tells you |
|------|------------------|
| **Invariant violated** | Which invariant failed |
| **Counterexample (N step(s))** | BFS found the *shortest* path — no shorter path exists |
| **Each step** | Which action ran + which variable changed and how |
| **[VIOLATION]** | The exact state where the invariant first fails |
| **Variable list** | The full state that broke the invariant |

---

## Five-step debugging workflow

```
1. Read the violated invariant
         │  "solvent: balance >= 0"
         v
2. Look at the violation step
         │  "withdraw: balance = -100"
         v
3. Read backwards — what enabled this action?
         │  Previous state: balance = 200, no require check
         v
4. Find the missing guard
         │  withdraw has no "require balance >= amount"
         v
5. Add require, re-verify
         │  spectre verify bank.spec  ->  (ok) safe
         v
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
invariant total_fixed   { alice + bob == 300 }
```

```bash
spectre verify transfer-broken.spec
```

```
Violations:

1.   Invariant violated: alice_solvent
  Action 'transfer' would violate invariants: [invariant alice_solvent violated]
  Counterexample trace (2 step(s)):
    [init]
    [step 1] transfer: alice: 200 -> 100, bob: 100 -> 200
    [step 2] transfer: alice: 100 -> 0,   bob: 200 -> 300
    [VIOLATION] invariant "alice_solvent" fails in this state
      alice = 0
      bob = 300
```

Spectre reports the state where the *next* `transfer` would violate `alice_solvent` (alice=0 -> -100). The fix — add a require:

```spectre
action transfer {
  require alice >= 100
  alice' = alice - 100
  bob'   = bob   + 100
}
```

```bash
spectre verify transfer-fixed.spec
# Exhaustive verification: traversed 3 states (full reachable space)
# Found no violations.
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
# (x) violation -- invariant 'alice_solvent' ...

# Fix alice_solvent, then:
spectre verify partially-fixed.spec
# (x) violation -- invariant 'total_fixed' ...
```

---

## Dealing with bounded results

```
bounded -- hit state limit (5000 states)
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
| Enum in wrong state | Action enabled in wrong phase | `require phase == Phase.Expected` |
| Two actors in critical section | No mutual exclusion | Add lock variable + requires |
| Accounting identity broken | Value leaked or doubled | Check all actions modify both sides |

---

## Tips for writing debuggable specs

- **Write accounting identities** — `alice + bob == TOTAL` catches money bugs for free
- **Name invariants clearly** — the name is the first thing you see in a violation
- **One concept per action** — large multi-step actions are hard to trace
- **Start with small params** — `--param N=3` first, then scale up
- **Verify after every change** — short iteration loops catch bugs while the context is fresh
