# Model-based testing

## The idea

```
account.spec
      │
      ▼
spectre verify --emit-traces trace.itf.json
      │
      ▼
spectre generate-driver --lang rust account.spec -o src/driver.rs
      │
      ▼
fill in apply_action() + current_state() against your BankAccount impl
      │
      ▼
cargo test
      │
      ├── ✓ 6/6 steps match ── implementation agrees with spec
      └── ✗ step 3 mismatch  ── spec found a bug in your Rust code
```

The spec is the oracle. The driver replays spec-generated execution traces against your actual Rust implementation and reports any disagreement.

---

## Generate a trace

```bash
spectre verify account.spec \
  --emit-traces trace.itf.json \
  --coverage-mode property
```

Coverage modes — pick based on what you want to exercise:

| Mode | Strategy | Best for |
|------|----------|----------|
| `action` | Prefer actions not yet seen in trace | Breadth of action coverage (default) |
| `property` | Steer toward invariant boundary values | Catching off-by-one bugs |
| `transition-pair` | Cover all (prev, next) action pairs | Transition sequence coverage |
| `boundary` | Maximise integer variable values | Overflow and limit bugs |
| `rare-action` | Prefer globally least-executed actions | Obscure code paths |

`property` mode achieves 52% boundary-state coverage vs 5–25% for the others — use it when chasing subtle arithmetic bugs.

---

## Generate the driver

```bash
spectre generate-driver --lang rust account.spec -o src/driver.rs
```

`driver.rs` gives you two traits to implement:

```rust
pub trait AccountDriver {
    /// Apply one action from the spec trace to your implementation.
    fn apply_action(&mut self, action: &str, args: &[Value]) -> Result<(), String>;

    /// Return the current implementation state in spec terms.
    fn current_state(&self) -> SpecState;
}
```

`SpecState` mirrors the spec's `var` declarations:

```rust
pub struct SpecState {
    pub balance: i64,
    pub frozen:  bool,
}
```

Wire these to your `BankAccount` struct and you're done.

---

## CI-ready driver with `--ci`

```bash
spectre generate-driver --lang rust account.spec --ci -o src/driver.rs
```

Three files are written:

| File | What to do with it |
|------|-------------------|
| `src/driver.rs` | Implement the two trait methods |
| `src/driver.cargo.toml` | Paste the `[dev-dependencies]` block into `Cargo.toml` |
| `src/driver.ci.yml` | Copy to `.github/workflows/` — runs `cargo test --test driver` on every push |

The Cargo snippet:

```toml
# paste into your [dev-dependencies]
spectre-connect = { git = "https://github.com/BoundlessProducts/spectre-connect" }
```

The CI workflow:

```yaml
name: Spectre conformance
on: [push, pull_request]
jobs:
  conformance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: rust-toolchain@stable
      - run: cargo test --test driver
```

No extra configuration — the workflow picks up `trace.itf.json` from the repo.

---

## Run the replay

```bash
cargo run -- trace.itf.json
```

Passing run:

```
Replaying 6 steps...
  step 1  deposit(100):   balance  0 → 100  ✓
  step 2  deposit(100):   balance 100 → 200  ✓
  step 3  withdraw(50):   balance 200 → 150  ✓
  step 4  freeze:         frozen false → true  ✓
  step 5  deposit(100):   skipped (frozen)  ✓
  step 6  unfreeze:       frozen true → false  ✓
✓ 6/6 steps match
```

Failing run:

```
✗ 1 mismatch:
  step 3, action 'withdraw', var 'balance':
    spec expects: 150
    impl returned: 200
    hint: check withdraw() arithmetic — possible sign error
```

---

## Violation traces as regression tests

When `spectre verify` finds a violation, `--emit-traces` captures the exact counterexample path:

```bash
spectre verify account.spec --emit-traces violation.itf.json
```

Check this file into your repo. It should:
- **Fail** against the unfixed implementation
- **Pass** after you apply the fix

```bash
# Before fix:
cargo run -- violation.itf.json  # → ✗ mismatch at step 4

# After fix:
cargo run -- violation.itf.json  # → ✓ 4/4 steps match
```

This turns every spec violation into a permanent, reproducible regression test.

---

## Repair suggestions before fixing

When verification fails, get candidate patches before editing anything:

```bash
spectre verify account.spec --suggest-repairs
```

```
Repair candidates (CANDIDATE — review before applying):

Violation 1: action 'withdraw' violates 'solvent'
  [1] CANDIDATE strengthen_require: add `require balance >= amount` to withdraw
      Validation: violation fixed · no deadlock · reachability unchanged — SAFE
  [2] CANDIDATE add_invariant_guard: add `when balance >= amount` to withdraw
      Validation: violation fixed · deadlock at balance=0 — review required
```

How to pick:

| Candidate type | Pick when | Risk |
|---------------|-----------|------|
| `strengthen_require` | The action should simply not fire when condition fails | Minimal |
| `add_invariant_guard` | You want the action blocked at the invariant boundary | Can introduce deadlock |

Apply the chosen patch to the spec, re-verify, then propagate the change to the Rust source.

---

## Checking refinement

When you have traces from an external source (a different tool, a recorded production run), check whether they conform to your spec:

```bash
spectre check-refinement account.spec
```

This reads `driver.json` (the refinement mapping) and validates that each external trace is a valid execution of the spec.

---

## Complete MBT workflow

```
1.  spectre mine --lang rust src/account.rs -o account.spec
          │
2.  Add invariants + temporal properties to account.spec
          │
3.  spectre verify account.spec
          ├── violation? → --suggest-repairs → fix → repeat
          └── safe
                │
4.  spectre verify account.spec \
      --emit-traces trace.itf.json \
      --coverage-mode property
          │
5.  spectre generate-driver --lang rust account.spec --ci -o src/driver.rs
          │
6.  Implement apply_action() and current_state() in driver.rs
          │
7.  cargo test --test driver
          ├── mismatch? → read step output → fix Rust impl → repeat from 7
          └── ✓ all pass
                │
8.  Check in: account.spec, trace.itf.json, driver.rs, driver.ci.yml
          │
9.  On every code change:
      spectre sync src/account.rs --spec account.spec
          ├── [✗] found? → re-verify → update trace → repeat from 7
          └── [=]/[≡] → nothing to do
```
