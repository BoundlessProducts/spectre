//! Embedded runtime monitor generated from `bank-account-corrected`.
//!
//! Drop this file into your project to check spec invariants on every state
//! transition in production or staging — no trace files, no offline replay.
//!
//! # Quick start
//! ```rust,ignore
//! let mut monitor = Monitor::new(SpecState { /* initial values */ });
//! monitor.on_violation(|v| eprintln!("[SPEC VIOLATION] {}", v));
//!
//! // After every state transition in your application:
//! if let Err(violations) = monitor.step("action_name", SpecState { /* new values */ }) {
//!     // violations is Vec<Violation> — handle or log them
//! }
//!
//! // At shutdown, check liveness obligations:
//! for prop in monitor.unmet_liveness_properties() {
//!     eprintln!("[SPEC] liveness property `{}` was never satisfied", prop);
//! }
//! ```

#![allow(dead_code)]

// ================================================================
// State
// ================================================================

/// Mirrors the spec variables. Fill in values from your application state
/// before calling `Monitor::step`.
#[derive(Debug, Clone)]
pub struct SpecState {
    pub aliceBalance: i64,
    pub bobBalance: i64,
}

// ================================================================
// Monitor
// ================================================================

/// A spec violation detected at runtime.
#[derive(Debug, Clone)]
pub struct Violation {
    /// Sequential step number when this violation was detected.
    pub step: u64,
    /// Name of the action that triggered the check.
    pub action: String,
    /// Name of the invariant or temporal property that was violated.
    pub property: String,
    /// Human-readable description from the spec.
    pub message: String,
}

impl std::fmt::Display for Violation {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "[step {}] `{}` violated after `{}`: {}",
            self.step, self.property, self.action, self.message
        )
    }
}

struct LivenessTracker {
    property: &'static str,
    satisfied: bool,
}

/// Embedded runtime monitor for the `bank-account-corrected` specification.
///
/// Checks all invariants after every `Monitor::step` call.
pub struct Monitor {
    state: SpecState,
    step_count: u64,
    violations: Vec<Violation>,
    liveness: Vec<LivenessTracker>,
    on_violation: Option<Box<dyn Fn(&Violation) + Send>>,
}

impl Monitor {
    /// Create a monitor starting from `initial`.
    pub fn new(initial: SpecState) -> Self {
        Self {
            state: initial,
            step_count: 0,
            violations: Vec::new(),
            liveness: Vec::new(),
            on_violation: None,
        }
    }

    /// Register a callback invoked immediately on every violation.
    /// Only one callback is supported; calling again replaces the previous one.
    pub fn on_violation<F: Fn(&Violation) + Send + 'static>(&mut self, f: F) {
        self.on_violation = Some(Box::new(f));
    }

    /// Apply `action` -> `new_state` and check all invariants.
    ///
    /// Returns `Err` with new violations; they are also accumulated in
    /// `Monitor::all_violations` regardless of the return value.
    pub fn step(&mut self, action: &str, new_state: SpecState) -> Result<(), Vec<Violation>> {
        self.step_count += 1;
        self.state = new_state;
        let new_violations = self.check_invariants(action);
        for v in &new_violations {
            self.violations.push(v.clone());
            if let Some(cb) = &self.on_violation {
                cb(v);
            }
        }
        if new_violations.is_empty() { Ok(()) } else { Err(new_violations) }
    }

    /// Current spec state.
    pub fn state(&self) -> &SpecState { &self.state }

    /// All violations recorded since the monitor was created.
    pub fn all_violations(&self) -> &[Violation] { &self.violations }

    /// Returns `true` if any invariant violation has been recorded.
    pub fn has_violations(&self) -> bool { !self.violations.is_empty() }

    /// Current step count (number of `Monitor::step` calls).
    pub fn step_count(&self) -> u64 { self.step_count }

    fn check_invariants(&self, action: &str) -> Vec<Violation> {
        let mut out = Vec::new();

        // invariant `no_negatives`
        if !(((self.state.aliceBalance >= 0) && (self.state.bobBalance >= 0))) {
            out.push(Violation {
                step: self.step_count,
                action: action.to_string(),
                property: "no_negatives".to_string(),
                message: "Invariant: Account balances should never be negative".to_string(),
            });
        }
        out
    }
}
