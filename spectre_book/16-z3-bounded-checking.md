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
