# Chapter 13: Runtime monitoring

Formal verification proves your spec is safe. Runtime monitoring proves your *running application* stays inside that spec — continuously, in production.

---

## The gap formal verification leaves

```
spectre verify account.spec
        │
   ✓ safe — all reachable states satisfy invariants
        │
   ship to production
        │
   third-party library update changes behaviour ← not in your spec
        │
   invariant violated at runtime — never detected
```

A runtime monitor closes this gap. It checks your spec's invariants on every real state transition — with no heap allocation in the common path.

---

## Generate a monitor

```bash
spectre generate-monitor --lang rust account.spec -o src/monitor.rs
```

The generated `monitor.rs` gives you:

```rust
/// Mirrors the spec variables.
#[derive(Debug, Clone)]
pub struct SpecState {
    pub balance: i64,
    pub frozen:  bool,
}

/// A spec violation detected at runtime.
#[derive(Debug)]
pub struct Violation {
    pub step:      u64,
    pub invariant: &'static str,
    pub state:     SpecState,
}

pub struct Monitor { /* private */ }

impl Monitor {
    pub fn new() -> Self { Monitor::default() }

    /// Call after every state transition in your application.
    pub fn step(&mut self, state: &SpecState) -> Option<Violation> { ... }

    /// Check liveness properties periodically (e.g. every 1000 steps).
    pub fn check_liveness(&self) -> Vec<&'static str> { ... }
}
```

---

## Wire it into your application

```rust
use crate::monitor::{Monitor, SpecState};

struct App {
    account: BankAccount,
    monitor: Monitor,
}

impl App {
    fn deposit(&mut self, amount: i64) {
        self.account.deposit(amount);

        let state = SpecState {
            balance: self.account.balance,
            frozen:  self.account.frozen,
        };

        if let Some(v) = self.monitor.step(&state) {
            tracing::error!(
                "spec violation: {} at step {}",
                v.invariant, v.step
            );
            // handle: log, alert, rollback, or panic — your choice
        }
    }
}
```

The monitor is:
- **Zero dependencies** — no external crate, no allocations in the happy path
- **O(k) per transition** where k = number of invariants
- **Consistent** — generated from the exact same `.spec` used for model checking

---

## Handle violations your way

| Response | When to use |
|----------|-------------|
| Log + continue | Non-critical invariant, high-traffic path |
| Alert + continue | Production monitoring with async alerting |
| Panic | Safety-critical — prefer crash over corrupt state |

---

## Validate monitor traces before deploying

```bash
spectre validate-monitor account.spec
```

```
Validated 2 violation trace(s):
  [1] action 'withdraw' → invariant 'solvent': consistent (4 steps, final violated)
  [2] action 'deposit'  → invariant 'within_cap': consistent (3 steps, final violated)

2/2 consistent.
```

A **consistent** trace means the BFS path genuinely leads to (or one step before) a state that fails the target invariant. If a trace is inconsistent, the violation recording has a bug — not your application.

Run `validate-monitor` after every spec change before regenerating the monitor.

Flags:

| Flag | Default | Effect |
|------|---------|--------|
| `--max-states` | 10000 | BFS exploration limit |
| `--max-depth` | 100 | BFS depth limit |

---

## What the monitor checks — and what it doesn't

**Checks:**
- Every `invariant` from the spec, after every `step()` call
- Liveness properties when you call `check_liveness()`

**Doesn't check:**
- Whether your `SpecState` mapping is correct — that's on you
- Temporal properties in real-time (sampled only on `check_liveness`)
- Invariants over variables you didn't include in `SpecState`

---

## Full monitoring workflow

```
account.spec
     │
     ▼
spectre verify account.spec
     │  ✓ safe — design-time proof
     ▼
spectre validate-monitor account.spec
     │  ✓ all violation traces consistent
     ▼
spectre generate-monitor --lang rust account.spec -o src/monitor.rs
     │
     ▼
implement SpecState mapping in application
     │
     ▼
call monitor.step(&state) after every transition
     │
     ▼
deploy
     │
     ├── violation fires ──► log / alert / panic per policy
     └── no violation    ──► ✓ spec holds in production
```

---

## Performance overhead

For `account.spec` with two invariants, each `step()` call is two integer comparisons. On modern hardware:

```
~2 ns per step (two invariants, no allocation)
```

At 100,000 transitions per second that's **0.02% CPU overhead** — negligible in any real workload.

---

## The consistency guarantee

Because the monitor is generated from the same `.spec` file:
- If your spec passes `spectre verify` ✓
- And the monitor never fires in production

...you have end-to-end assurance: your implementation matches the design across every observed execution.
