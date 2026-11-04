# Runtime monitoring

Formal verification proves your spec is safe. A runtime monitor checks your *running application* against those same spec invariants — catching violations that formal verification cannot see (third-party library updates, misconfigured deployments, out-of-spec inputs).

---

## The gap formal verification leaves

```
spectre verify account.spec
        |
   safe -- all reachable states satisfy invariants
        |
   ship to production
        |
   third-party library update changes behaviour  <- not in your spec
        |
   invariant violated at runtime -- never detected
```

A runtime monitor closes this gap by re-checking the spec's invariants on every real state transition.

---

## What the generated monitor actually is

`spectre generate-monitor` produces a Rust struct with one inline boolean assertion per spec invariant. It is a **generated assertion library** — not a formal runtime verifier. Its value is that the assertions are derived from the exact same predicates your model checker verified, so they are guaranteed to be consistent with the spec.

The `SpecState` mapping between your application and the monitor is **entirely on you**. If you map the wrong fields, the monitor checks the wrong things.

---

## Generate a monitor

```bash
spectre generate-monitor --lang rust account.spec -o src/monitor.rs
```

The generated `monitor.rs` gives you:

```rust
/// Mirrors the spec variables. You fill this in from your application state.
#[derive(Debug, Clone)]
pub struct SpecState {
    pub balance: i64,
    pub frozen:  bool,
}

/// A spec invariant violation detected at runtime.
#[derive(Debug)]
pub struct Violation {
    pub step:      u64,
    pub invariant: &'static str,
    pub state:     SpecState,
}

pub struct Monitor { /* private */ }

impl Monitor {
    pub fn new() -> Self { Monitor::default() }

    /// Call after every state transition. Checks all spec invariants.
    /// Returns Some(Violation) if any invariant fails.
    pub fn step(&mut self, state: &SpecState) -> Option<Violation> { ... }

    /// Sample liveness properties periodically (e.g. every 1000 steps).
    pub fn check_liveness(&self) -> Vec<&'static str> { ... }
}
```

Each `step()` call is O(k) where k is the number of invariants — typically two to five integer comparisons, with no heap allocation.

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

        // You are responsible for this mapping being correct.
        let state = SpecState {
            balance: self.account.balance,
            frozen:  self.account.frozen,
        };

        if let Some(v) = self.monitor.step(&state) {
            tracing::error!(
                "spec violation: {} at step {}",
                v.invariant, v.step
            );
            // handle: log, alert, rollback, or panic -- your choice
        }
    }
}
```

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
  [1] action 'withdraw' -> invariant 'solvent': consistent (4 steps, final violated)
  [2] action 'deposit'  -> invariant 'within_cap': consistent (3 steps, final violated)

2/2 consistent.
```

A **consistent** trace means the BFS path genuinely leads to a state that fails the target invariant. Run `validate-monitor` after every spec change before regenerating the monitor.

---

## What the monitor checks — and what it doesn't

**Checks:**
- Every `invariant` from the spec, after every `step()` call
- Liveness properties when you call `check_liveness()`

**Does not check:**
- Whether your `SpecState` mapping is correct — that is entirely on you
- Temporal properties in real-time (sampled only on `check_liveness`)
- Invariants over variables you did not include in `SpecState`

---

## Full monitoring workflow

```
account.spec
     |
     v
spectre verify account.spec
     |  safe -- design-time exhaustive check
     v
spectre validate-monitor account.spec
     |  all violation traces consistent
     v
spectre generate-monitor --lang rust account.spec -o src/monitor.rs
     |
     v
implement SpecState mapping in application
     |
     v
call monitor.step(&state) after every transition
     |
     v
deploy
     |
     +-- violation fires --> log / alert / panic per policy
     +-- no violation    --> spec holds in every observed execution
```

---

## Performance overhead

For `account.spec` with two invariants, each `step()` call is two integer comparisons:

```
~2 ns per step (two invariants, no allocation)
```

At 100,000 transitions per second that is 0.02% CPU overhead — negligible in any real workload.
