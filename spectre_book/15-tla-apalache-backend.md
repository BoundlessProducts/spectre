# TLA+ and Apalache backend

BFS is the learning backend — great for writing and exploring specs. When you need validated fairness semantics or your state space is too large for BFS, `--backend tlc` and `--backend apalache` hand off to proven model checkers built on TLA+.

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

Spectre translates your `.spec` file into a standard TLA+ module plus a `.cfg` file. Here is the counter example:

**`counter.spec`**
```
var x: int
param MAX: int
init { x = 0 }
action incr { require x < MAX  x' = x + 1 }
invariant non_negative { x >= 0 }
temporal wf_progress { WF(incr) }
```

**Generated `SpectreSpec.tla`**
```tla
---- MODULE SpectreSpec ----
EXTENDS Integers, Sequences, FiniteSets, TLC

CONSTANTS MAX

VARIABLES
  \* @type: Int;
  x

\* @type: <<Int>>;
vars == <<x>>

TypeOK ==
  /\ x \in Int

Init ==
  /\ x = 0

Incr ==
  /\ x < MAX
  /\ x' = (x + 1)

Next == Incr

non_negative == (x >= 0)

Spec == Init /\ [][Next]_vars
  /\ WF_vars(Incr)

====
```

**Generated `SpectreSpec.cfg`**
```
INIT Init
NEXT Next
INVARIANT non_negative
PROPERTY wf_progress
CONSTANTS
  MAX = 0
```

Spectre runs `tlc -config SpectreSpec.cfg SpectreSpec.tla` and parses the output.

The `\* @type:` comment annotations are for Apalache's Snowcat type checker — TLC ignores them. The module name is always `SpectreSpec` (not the spec filename) because SANY, the TLA+ parser used by both TLC and Apalache, requires the filename to match the `MODULE` name.

---

## Expression mapping reference

| Spectre | TLA+ |
|---------|------|
| `x != y` | `x # y` |
| `x <= y` | `x =< y` |
| `x && y` | `x /\ y` |
| `x \|\| y` | `x \/ y` |
| `!x` | `~x` |
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

BFS explores traces and flags cases where an enabled action is never taken. This is exploratory — useful for learning, but not an authoritative liveness result.

TLC's approach is exact:

```
Liveness check (4 steps):

1. Decompose Spec into safety + fairness constraints
   Spec == Init /\ [][Next]_vars /\ WF_vars(incr)
                                   +-- fairness constraint

2. Find all SCCs in the state graph
   (strongly connected components = potential infinite loops)

3. For each SCC, check: does incr get to fire?
   If incr is enabled in some state of an SCC but never fires
   -> genuine fairness violation

4. Report: the specific cycle that violates WF(incr)
```

This SCC-based approach is why `--backend tlc` gives you authoritative liveness results.

---

## TLC vs Apalache: when to use which

| | TLC | Apalache |
|---|-----|----------|
| **Approach** | Explicit-state enumeration | Symbolic (SMT-based) |
| **Best for** | Liveness, fairness, small-medium spaces | Large state spaces, safety |
| **Fairness (WF/SF)** | Validated (exact SCC-based) | Limited |
| **Temporal / liveness** | Validated | Limited |
| **Safety invariants** | Exhaustive | Exhaustive (symbolic) |
| **State space limit** | ~10M states practical | Effectively unbounded |
| **Speed on small specs** | Fast | Slower (SMT overhead) |
| **Speed on large specs** | Slow (memory) | Fast (symbolic) |
| **Counterexample** | Full trace | Symbolic witness |
| **Requires Java** | Yes | Yes |

Rule of thumb:
- Use **TLC** when you have `temporal` or `WF`/`SF` properties — it is the only backend that validates fairness correctly
- Use **Apalache** when BFS times out on safety properties — it handles large spaces symbolically, but fairness support is limited

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
mutex.spec: violation found
  Violation 1: invariant "mutual_exclusion" triggered by action "enter_cs"
    Step 0: [pc_a = idle,    pc_b = idle,    flag_a = false, flag_b = false]
    Step 1: [pc_a = wanting, pc_b = idle,    flag_a = true,  flag_b = false]
    Step 2: [pc_a = wanting, pc_b = wanting, flag_a = true,  flag_b = true ]
    Step 3: [pc_a = critical,pc_b = wanting, flag_a = true,  flag_b = true ]
    Step 4: [pc_a = critical,pc_b = critical,flag_a = true,  flag_b = true ]
                                              ^ both in critical section -- violated
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

Apalache is not in Homebrew. Download from GitHub releases and install manually:

```bash
# 1. Download from https://github.com/informalsystems/apalache/releases
#    (choose the .zip for your platform)
unzip apalache-*.zip
sudo ln -sf "$PWD/apalache/bin/apalache-mc" /usr/local/bin/apalache-mc

# 2. Requires Java 21+ (Apalache v0.62+ uses class file version 65)
brew install openjdk          # macOS — installs Java 25; Java 21+ works

# 3. Verify
apalache-mc version           # should print e.g. 0.62.4
```

**Java setup:** Spectre auto-detects the Homebrew OpenJDK at `/opt/homebrew/opt/openjdk` and sets `JAVA_HOME` for the Apalache subprocess — you do not need to set `JAVA_HOME` manually on macOS if you installed via `brew install openjdk`.

**Bounded check:** Apalache performs bounded symbolic model checking — it checks up to N steps (default 10). Unlike BFS, it is not a full state-space enumeration.

```bash
# Default (depth 10)
spectre verify --backend apalache account.spec

# Custom binary path
spectre verify --backend apalache \
    --apalache-bin ~/tools/apalache/bin/apalache-mc account.spec
```
