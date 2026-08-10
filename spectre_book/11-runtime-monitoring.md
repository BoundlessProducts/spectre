# Chapter 11: Runtime Monitoring

Formal verification gives you confidence at design time.  But a production system can still diverge from its specification through incomplete test coverage, refactoring, or third-party library updates.  **Runtime monitoring** closes this gap by embedding spec checks directly inside your running application.

Spectre generates a self-contained `monitor.rs` file — no external dependencies, no trace files, no offline replay.  The monitor checks every state transition in O(k) time where k is the number of invariants.

---

## Generating a Monitor

```bash
spectre generate-monitor --lang rust examples/bank-account-parameterized.spec \
    --output src/monitor.rs
```

This produces a `monitor.rs` with:

- A `SpecState` struct mirroring your spec variables
- A `Monitor` struct with `step()` and `unmet_liveness_properties()` methods
- A `Violation` struct for structured error reporting
- A callback hook for custom violation handling

---

## The Generated API

Here is the generated interface for `bank-account-parameterized.spec`:

```rust
/// Mirrors the spec variables.
#[derive(Debug, Clone)]
pub struct SpecState {
    pub balance: i64,
    pub frozen:  bool,
}

/// A spec violation detected at runtime.
#[derive(Debug, Clone)]
pub struct Violation {
    pub step:     u64,    // sequential step number
    pub action:   String, // which action triggered the check
    pub property: String, // invariant or temporal property name
    pub message:  String, // human-readable description from the spec
}

pub struct Monitor { /* ... */ }

impl Monitor {
    /// Create a monitor starting from `initial`.
    pub fn new(initial: SpecState) -> Self;

    /// Register a callback invoked immediately on every violation.
    pub fn on_violation<F: Fn(&Violation) + Send + 'static>(&mut self, f: F);

    /// Apply `action` → `new_state` and check all invariants.
    /// Returns Err(violations) if any invariant was violated.
    pub fn step(&mut self, action: &str, new_state: SpecState)
        -> Result<(), Vec<Violation>>;

    /// Returns names of temporal properties that were never satisfied.
    /// Call this at shutdown.
    pub fn unmet_liveness_properties(&self) -> Vec<&str>;

    /// All violations recorded since the monitor was created.
    pub fn all_violations(&self) -> &[Violation];
}
```

---

## Integrating the Monitor

Add four lines to your existing `BankAccount` implementation:

```rust
mod monitor;  // include the generated monitor.rs
use monitor::{Monitor, SpecState};

pub struct BankAccount {
    pub balance: i64,
    pub frozen:  bool,
    monitor: Monitor,
}

impl BankAccount {
    pub fn new(owner: String) -> Self {
        let initial = SpecState { balance: 0, frozen: false };
        let mut monitor = Monitor::new(initial);
        monitor.on_violation(|v| {
            eprintln!("[SPEC VIOLATION] {}", v);
            // In production you might log to your observability stack instead
        });
        Self { balance: 0, frozen: false, monitor }
    }

    pub fn deposit(&mut self, amount: i64) {
        assert!(amount > 0);
        self.balance += amount;
        // Report the new state to the monitor after every transition
        let _ = self.monitor.step("deposit", SpecState {
            balance: self.balance,
            frozen:  self.frozen,
        });
    }

    pub fn withdraw(&mut self, amount: i64) {
        self.balance -= amount;
        let _ = self.monitor.step("withdraw", SpecState {
            balance: self.balance,
            frozen:  self.frozen,
        });
    }

    pub fn freeze(&mut self) {
        self.frozen = true;
        let _ = self.monitor.step("freeze", SpecState {
            balance: self.balance,
            frozen:  self.frozen,
        });
    }

    pub fn unfreeze(&mut self) {
        self.frozen = false;
        let _ = self.monitor.step("unfreeze", SpecState {
            balance: self.balance,
            frozen:  self.frozen,
        });
    }
}

impl Drop for BankAccount {
    fn drop(&mut self) {
        // Check liveness properties at shutdown
        for prop in self.monitor.unmet_liveness_properties() {
            eprintln!("[SPEC] liveness property `{}` was never satisfied", prop);
        }
    }
}
```

---

## What the Monitor Checks

After each `monitor.step()` call, the monitor:

1. **Checks every `invariant` block** against the new state.  For `bank-account-parameterized.spec`:
   - `non_negative`: `balance >= 0`
   - `within_cap`: `balance <= 1000000`

2. **Advances liveness trackers** for each `temporal` property.  For `eventually_usable` (`eventually !frozen`), it records whether the property has ever been satisfied.

3. **Calls the violation callback** if any check fails — immediately, synchronously, in the calling thread.

---

## Violation Output

When the monitor fires, the output looks like:

```
[SPEC VIOLATION] [step 3] `non_negative` violated after `withdraw`: balance >= 0
```

The `Violation` struct gives you everything needed to log to your observability stack:

```rust
monitor.on_violation(|v| {
    tracing::error!(
        step    = v.step,
        action  = %v.action,
        property = %v.property,
        message = %v.message,
        "spec invariant violated"
    );
    metrics::counter!("spec.violations", 1, "property" => v.property.clone());
});
```

---

## Choosing the Violation Response

The `step()` return value lets you choose how to respond:

```rust
match self.monitor.step("withdraw", new_state) {
    Ok(()) => {}                      // all invariants held
    Err(violations) if cfg!(debug_assertions) => {
        panic!("spec violated in test: {:?}", violations);
    }
    Err(violations) => {
        // Production: log and continue, or abort, depending on severity
        for v in violations {
            log::error!("invariant {} violated: {}", v.property, v.message);
        }
    }
}
```

---

## Performance Overhead

The `step()` call path:

- Clones `SpecState` (stack-allocated struct copy)
- Evaluates k invariant expressions (simple arithmetic comparisons)
- No heap allocation in the common (no-violation) path

For `bank-account-parameterized.spec` with 2 invariants and 1 temporal property, this is three comparisons per transition — negligible in any production workload.

---

## Spec Consistency Guarantee

Because the runtime monitor is generated from the **same spec file** used for model checking, it is guaranteed to be consistent with your verification results.  There is no separate monitoring-rule language to keep in sync.

If your spec passes `spectre verify`, and the monitor never fires in production, you have end-to-end assurance that your implementation matches the design.

---

## Complete Closed-Loop Workflow

```
┌────────────────────────────────────────────────────────────────┐
│                     Closed-Loop Verification                    │
│                                                                │
│  1. spectre mine --lang rust src/account.rs                    │
│     → account.spec skeleton                                    │
│                                                                │
│  2. Add invariants + temporal properties                       │
│     spectre verify account.spec  → fix with CEGIS if needed   │
│                                                                │
│  3. spectre verify account.spec --emit-traces trace.itf.json   │
│     spectre generate-driver --lang rust account.spec           │
│     → fill in driver.rs, run against trace  (Chapter 10)      │
│                                                                │
│  4. spectre generate-monitor --lang rust account.spec          │
│     → embed monitor.rs in production binary  (this chapter)   │
│                                                                │
│  Single spec → design-time verification                        │
│              → model-based test replay                         │
│              → production runtime monitoring                   │
└────────────────────────────────────────────────────────────────┘
```

This is the full Spectre closed loop: one spec file drives verification, testing, and production monitoring simultaneously.
