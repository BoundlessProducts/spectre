# Chapter 12: Evaluating Rust Coverage and Benchmarking

This chapter covers two meta-tooling commands: `spectre eval-rust`, which measures how completely the miner can translate a Rust codebase, and `spectre benchmark`, which measures how quickly Spectre explores the state space of a set of specs.

---

## Evaluating Rust Extraction with `eval-rust`

Before committing to a Spectre workflow for a Rust project, it helps to know up-front which parts of the codebase the miner can translate cleanly and which require manual follow-up.  `spectre eval-rust` runs the miner over one or more Rust source files and prints a Markdown report summarising extraction quality.

```bash
spectre eval-rust src/account.rs src/order.rs src/payment.rs
```

Or over an entire directory:

```bash
spectre eval-rust --dir src/
```

### Reading the Report

```
# Rust Extraction Evaluation Report

## Per-File Summary

| File           | Struct       | Fields | Enums | Methods | Omissions | SuccessFraction |
|----------------|--------------|--------|-------|---------|-----------|----------------|
| bank_account.rs | BankAccount  |   3    |   1   |    4    |     4     |      50%       |
| counter.rs      | Counter      |   2    |   0   |    3    |     1     |     100%       |
| state_machine.rs | Machine     |   4    |   2   |    2    |     4     |      50%       |

**Totals** — files: 3 | fields: 9 | enums: 3 | methods: 9 | omissions: 9 | mean success: 67%

## Omissions by Kind

| Kind            | Count |
|-----------------|-------|
| SKIPPED_METHOD  |   3   |
| TUPLE_VARIANT   |   2   |
| OVERFLOW_OP     |   2   |
| RESULT_RETURN   |   2   |
```

**SuccessFraction** is the fraction of methods with no body-level omission (overflow, panic, Result, or complex assign).  `SKIPPED_METHOD` (constructors) does not count against the fraction — skipping `new()` is expected and benign.

A file at 100% means every extracted method translates cleanly to Spectre actions.  A file at 50% has at least one method whose body required abstraction; those methods should be reviewed before treating the mined spec as authoritative.

### Interpreting the Omissions-by-Kind Table

Focus attention in order:

1. **`RESULT_RETURN`** — methods that return `Result<T,E>` are mined as happy-path only.  If error paths are semantically important for your invariants, model them manually.
2. **`OVERFLOW_OP`** — `checked_add`, `wrapping_add`, etc. are abstracted to plain addition.  If your invariants involve overflow-safe bounds, add explicit `require` guards.
3. **`TUPLE_VARIANT`** — enum variants with data are collapsed to unit variants.  Add back any fields you need as separate `var` declarations.
4. **`SKIPPED_METHOD`** — expected.  Review if a skipped method has semantically important side-effects.

### Using `eval-rust` in CI

Add an `eval-rust` step to your CI pipeline to track extraction coverage over time:

```yaml
- name: Check Spectre extraction coverage
  run: |
    spectre eval-rust --dir src/ > eval-report.md
    # Fail if mean success drops below 80%
    grep "mean success" eval-report.md | grep -v " [0-7][0-9]%" || true
```

You can also commit the Markdown report as an artifact alongside the generated spec, providing a permanent record of which Rust constructs required manual work.

---

## Benchmarking with `spectre benchmark`

`spectre benchmark` runs BFS exploration on a set of spec files and prints a timing table.  It is useful for tracking the performance impact of spec changes, comparing exploration depth across models, and establishing a baseline before enabling optimisations such as `--por` or `--property-guided`.

```bash
spectre benchmark
```

By default it scans the `benchmarks/` directory for `*.spec` files.  To use a different directory or explicit files:

```bash
spectre benchmark --dir my-specs/
spectre benchmark examples/raft.spec examples/bank.spec
```

### Reading the Output

```
Spec                            States      Elapsed     Depth
------------------------------------------------------------------
counter.spec                        11         0ms          10
bank.spec                          150         4ms          30
mutex.spec                           9         0ms           8
resource_alloc.spec                480         1ms          20
cache_coherence.spec              2400        12ms          40
------------------------------------------------------------------
5 spec(s) benchmarked
```

**States** is the number of distinct reachable states (up to `--max-states`).  **Elapsed** is wall-clock time for the full BFS.  **Depth** is the deepest state reached.

### Flags

| Flag | Default | Effect |
|------|---------|--------|
| `--max-states` | 50000 | Stop BFS after this many states |
| `--dir` | `benchmarks/` | Directory to scan for `*.spec` files |

### The Built-in Benchmark Suite

Spectre ships with five benchmark specs in `benchmarks/`:

| Spec | States | What it models |
|------|--------|---------------|
| `counter.spec` | ~11 | Simple integer counter with a bound |
| `bank.spec` | ~150 | Parameterised bank account (`param LIMIT`) |
| `mutex.spec` | ~9 | Two-process mutex with `ProcState` enum |
| `resource_alloc.spec` | ~480 | N-slot resource pool (`param N`) |
| `cache_coherence.spec` | ~2400 | MSI cache coherence protocol |

These cover the range from trivial (counter) to moderately complex (cache coherence), providing a quick health-check of the verifier engine across representative models.

### Adding Your Own Benchmarks

Any `.spec` file dropped into `benchmarks/` is automatically picked up.  For parameterised specs, use `param` declarations so the benchmark suite covers multiple configurations:

```spectre
param N: int = 4        // default used by benchmark; override with --param N=8
var slots: int
```

### Benchmark vs Verify

`spectre benchmark` intentionally does not check invariants — it measures raw exploration throughput.  To measure verified throughput (invariant checking included), use `spectre verify` directly:

```bash
time spectre verify benchmarks/cache_coherence.spec --max-states 50000
```

The difference between benchmark and verify timing gives you the cost of invariant evaluation for your spec.

---

## Combined Workflow

```
eval-rust → identify gaps → mine + fix → verify → benchmark
```

1. **`eval-rust --dir src/`** — measure extraction coverage before writing any specs manually
2. **`mine + review`** — focus manual effort on the files and methods flagged by `eval-rust`
3. **`verify`** — confirm the enriched spec is safe
4. **`benchmark`** — establish a timing baseline; re-run after adding new actions or invariants to catch state-space explosions early
