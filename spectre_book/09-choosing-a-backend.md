# Choosing a verification backend

`spectre verify` supports four backends. BFS is the learning backend — zero setup, instant traces, great while you are writing and exploring a spec. The others are production backends for when you need exhaustive or validated results.

---

## Two tracks: learning vs. production

```
Learning track
  └─ --backend bfs  (default)
       No setup. Instant traces. Great for writing specs and debugging.
       Liveness results are exploratory — use TLC to validate them.

Production track
  ├─ --backend tlc        Validated liveness + fairness (WF/SF)
  ├─ --backend apalache   Exhaustive safety on large state spaces
  └─ --backend z3         Fast shallow bug hunting in CI
```

Pick BFS while you are learning or iterating. Graduate to a production backend before sign-off.

---

## The four backends at a glance

```
spectre verify myspec.spec
          |
          |  --backend bfs        (default, learning)
          |  --backend tlc
          |  --backend apalache
          |  --backend z3  --bmc-depth N
          |
          v
    +-----------+--------------+--------------+---------------+
    |   BFS     |    TLC       |  Apalache    |     Z3        |
    +-----------+--------------+--------------+---------------+
    | Learning  | Exhaustive   | Exhaustive   | Bounded       |
    | & explore | + validated  | symbolic     | up to depth N |
    |           | fairness     | large spaces |               |
    +-----------+--------------+--------------+---------------+
    | No deps   | TLA+ tools   | Apalache bin | z3 binary     |
    |           | (jar or bin) |              |               |
    +-----------+--------------+--------------+---------------+
```

---

## Quick decision guide

```
Writing a new spec or debugging?
  +-- yes --> --backend bfs  (default, zero setup, instant traces)

Need validated liveness or WF/SF fairness?
  +-- yes --> --backend tlc  (exact SCC-based cycle detection)

State space too large for BFS (> 100k states)?
  +-- yes --> --backend apalache  (symbolic, handles millions of states)

Just hunting for shallow bugs fast?
  +-- yes --> --backend z3 --bmc-depth 20  (finds bugs in seconds)

CI regression check?
  +-- yes --> --backend z3 --bmc-depth 30  (fast, catches most regressions)
```

---

## BFS backend (default — learning mode)

The built-in breadth-first search over your spec's reachable states. No external tools needed. This is where you start.

```bash
spectre verify counter.spec
# safe -- 11 states explored (exhaustive)

spectre verify counter.spec --max-states 50000
# safe -- 11 states explored (exhaustive)
```

**Best for:**
- Writing and iterating on specs — zero setup, immediate feedback
- Specs with up to ~50k reachable states
- Learning how your spec behaves

**Keep in mind:**
- BFS explores traces; liveness results are exploratory. For validated liveness, use `--backend tlc`.
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
mutex.spec: safe (exhaustive)
  Note: TLC checked all reachable states
```

**Best for:**
- Specs with `WF(action)` or `SF(action)` — TLC validates fairness correctly using SCC-based cycle detection
- Final sign-off on liveness properties
- Cross-checking a spec against the TLA+ ecosystem

**Limitations:**
- Slower than BFS for small specs
- Requires Java + TLA+ tools installed

### What Spectre generates

For each `.spec` file, Spectre creates two temp files:

```
---- MODULE myspec ----
EXTENDS Integers, TLC

VARIABLES balance, status

vars == <<balance, status>>

Init == /\ balance = 0
        /\ status  = "Active"

Deposit == /\ status = "Active"
           /\ balance < 1000
           /\ balance' = balance + 100
           /\ UNCHANGED <<status>>

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
- Production-grade safety verification

**Limitations:**
- Fairness (WF/SF): limited — Apalache has partial temporal support but does not perform full SCC-based liveness checking. Use TLC for validated fairness.
- Slowest to start (JVM startup + type checking)
- Requires Apalache installed (Java 17+)

### BFS vs Apalache on a large spec

```
spectre verify raft.spec --max-states 100000
# bounded -- hit state limit (100000 states)

spectre verify raft.spec --backend apalache
# safe (exhaustive)
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
bank.spec: violation found
  Note: bounded model checking to depth 20
  Violation 1: invariant "solvent"
    step 3: balance = -100
```

When no violation is found within the depth:
```
bank-fixed.spec: safe
  Note: bounded model checking to depth 20 -- not exhaustive
```

> `safe` from Z3 means **no violation up to depth 20** — not a proof of full safety. Use `--backend bfs` or `--backend apalache` for an exhaustive result.

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
| Role | learning | production | production | production |
| Exhaustive | small specs | yes | yes | no (bounded) |
| Fairness (WF/SF) | exploratory | validated | limited | none |
| Liveness | exploratory | validated | limited | none |
| Safety | exhaustive | exhaustive | exhaustive | bounded |
| Dependencies | none | TLA+ tools | Apalache bin | z3 binary |
| Speed | fast | moderate | fast | very fast |
| Setup | 0 min | 5 min | 10 min | 2 min |

---

## Progressive verification workflow

Start with BFS while writing. Graduate to production backends for sign-off.

```
Write spec
    |
    v
--backend bfs         <- instant feedback, no deps, great for learning
    |
  passes?
    |
    +- no --> fix spec, repeat
    |
    +- yes
         |
         v
    Has WF/SF fairness or liveness?
         +- yes --> --backend tlc   (validated SCC-based checking)
         |
         +- no, state space large?
              +- yes --> --backend apalache  (symbolic exhaustive)
              +- no  --> done  (BFS exhaustive for small specs)
```

In CI, combine:

```bash
# Fast regression check (seconds)
spectre verify myspec.spec --backend z3 --bmc-depth 30

# Full verification (minutes, on merge to main)
spectre verify myspec.spec --backend tlc
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
