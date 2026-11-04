# Benchmarking and performance

Before you change a spec, you want to know how big the state space is and how fast Spectre explores it. That is what the built-in benchmark suite is for.

---

## Run the benchmarks

```bash
spectre benchmark
```

```
Benchmark results (BFS backend)
================================
counter.spec        10 states   0.12 ms    depth 5
account.spec        84 states   0.41 ms    depth 8
mutex.spec         430 states   2.1 ms     depth 14
dist_queue.spec   3210 states   18 ms      depth 24
ring_leader.spec  9980 states   94 ms      depth 40

Total: 5 specs, 13714 states, 115 ms
```

The benchmark runs your spec files in `benchmarks/` against the BFS backend and reports state count, wall time, and max depth.

Add your own specs:

```bash
# drop any .spec file into benchmarks/
cp account.spec benchmarks/

spectre benchmark
```

---

## Understand state-space growth

State spaces grow exponentially with the number of variables and their ranges. Here is what to expect:

| Pattern | Typical state count | Why |
|---------|-------------------|-----|
| Counter (1 var, range 0-10) | ~11 | Linear |
| 2 independent counters (0-10 each) | ~121 | Product |
| Mutex (2 threads, 3 states each) | ~9 | Small product |
| Mutex + shared counter (0-10) | ~90 | 3-way product |
| Producer-consumer (queue depth 5) | ~500 | Queue contents combinatorial |
| Ring leader election (5 nodes) | ~100K+ | Permutation space |

The rule: **every new variable multiplies the state space by its domain size**.

---

## State-space growth diagram

```
Variables:  1     2     3     4     5
            |     |     |     |     |
Range 10:  10   100  1000  10K  100K
Range 5:    5    25   125  625   3K
Range 2:    2     4     8   16   32
            |
            +-- BFS exhausts this in < 1 ms
                              |
                              +-- starts taking seconds
                                          |
                                          +-- needs --backend tlc or z3
```

---

## Bounded exploration options

When a spec is too large for exhaustive BFS, cap the search:

```bash
# Cap at 50,000 states
spectre verify --max-states 50000 ring_leader.spec

# Cap at depth 30
spectre verify --max-depth 30 ring_leader.spec

# Both
spectre verify --max-states 50000 --max-depth 30 ring_leader.spec
```

The result will say `(bounded)` rather than `(exhaustive)`.

---

## Partial-order reduction

For concurrent systems, many interleavings produce identical states. `--por` (partial-order reduction) prunes these by exploring only one representative ordering:

```bash
spectre verify --por mutex.spec
```

```
Before --por: 430 states, 2.1 ms
After  --por: 127 states, 0.7 ms
```

Savings are largest when:
- Many threads share little state
- Actions are mostly independent

`--por` is sound for safety properties. Avoid it for liveness properties unless you also use `--backend tlc`.

---

## Property-guided exploration

```bash
spectre verify --property-guided mutex.spec
```

Property-guided mode biases BFS toward states that are close to violating an invariant. It finds bugs faster on large specs but may miss violations in rarely-reached corners.

Use it for:
- Interactive debugging — "show me a violation fast"
- Large specs where exhaustive BFS times out

Do not use it for production sign-off — use exhaustive BFS or TLC for that.

---

## When to switch backends

| Situation | Recommendation |
|-----------|----------------|
| Learning or iterating on a spec | BFS (default, zero setup) |
| Safety only, <= 100K states | BFS |
| Fairness / liveness properties | `--backend tlc` (validated SCC-based) |
| > 1M states, safety only | `--backend apalache` (symbolic) |
| Fast shallow bug hunting (< 20 steps) | `--backend z3 --bmc-depth 20` |
| Exhaustive + fairness + large space | `--backend tlc` |

---

## Progressive verification strategy

Start cheap, escalate only if needed:

```
spectre verify --max-states 10000 my.spec   (fast exploration)
        |
   safe in 10K states --> probably fine for small specs
        |
   violation or exhausted
        |
        v
spectre verify my.spec                         (exhaustive BFS, no limit)
        |
   exhaustive safe --> ship
        |
   too large / times out
        |
        v
spectre verify --backend z3 --bmc-depth 50 my.spec   (fast shallow bugs)
        |
   sat -> violation in <=50 steps --> fix, restart
        |
   unsat (no shallow bugs found)
        |
        v
spectre verify --backend tlc my.spec                 (full model check)
```

Each step is significantly more expensive. Stop as soon as you get a definitive answer.

---

## Benchmark in CI

```yaml
# .github/workflows/ci.yml (excerpt)
- name: Benchmark smoke test
  run: |
    spectre benchmark --max-states 1000 --ci
    # Fails if any spec exceeds 1000 states (catches accidental bloat)
```

The `--ci` flag exits non-zero if any spec exceeds `--max-states`. This catches spec bloat before it becomes a slow-CI problem.
