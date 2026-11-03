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
