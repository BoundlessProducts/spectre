use std::collections::HashMap;

use crate::error::ReplayError;
use crate::trace::{Trace, TraceStep};
use crate::value::ItfValue;

/// A mismatch between the spec's expected state and the implementation's actual state.
#[derive(Debug, Clone)]
pub struct StateMismatch {
    /// 0-based step number where the mismatch occurred.
    pub step: usize,
    /// Action that was applied at this step.
    pub action: String,
    /// Name of the variable that differed.
    pub variable: String,
    /// Value the spec said the variable should have.
    pub expected: ItfValue,
    /// Value the implementation returned.
    pub actual: ItfValue,
}

/// Summary of a completed replay.
#[derive(Debug)]
pub struct ReplayResult {
    /// Number of steps that were executed (including the initial state check).
    pub steps_executed: usize,
    /// All mismatches found. Empty means the implementation matched the spec at every step.
    pub mismatches: Vec<StateMismatch>,
}

impl ReplayResult {
    /// Returns `true` if the implementation matched the spec at every step.
    pub fn passed(&self) -> bool {
        self.mismatches.is_empty()
    }
}

/// Implement this trait to connect your Rust system to Spectre traces.
///
/// # Example (counter)
/// ```rust,ignore
/// struct Counter { value: i64 }
///
/// impl StepHandler for Counter {
///     type Error = Box<dyn std::error::Error>;
///
///     fn apply_action(&mut self, action: &str, args: &[ItfValue]) -> Result<(), Self::Error> {
///         match action {
///             "increment" => { self.value += 1; Ok(()) }
///             "decrement" => { self.value -= 1; Ok(()) }
///             "reset"     => { self.value = 0;  Ok(()) }
///             _ => Err(format!("unknown action: {action}").into()),
///         }
///     }
///
///     fn current_state(&self) -> HashMap<String, ItfValue> {
///         let mut s = HashMap::new();
///         s.insert("counter".into(), ItfValue::Int(self.value));
///         s
///     }
/// }
/// ```
pub trait StepHandler {
    type Error: std::fmt::Display;

    /// Apply `action` with `args` to mutate the system's state.
    ///
    /// The replayer calls this for every step in the trace (skipping index 0,
    /// which is the initial state). Errors are returned as [`ReplayError::StepFailed`].
    fn apply_action(&mut self, action: &str, args: &[ItfValue]) -> Result<(), Self::Error>;

    /// Return the system's current state as a map from variable name to value.
    ///
    /// The replayer calls this after every `apply_action` and compares it to
    /// the corresponding state in the ITF trace.
    fn current_state(&self) -> HashMap<String, ItfValue>;
}

/// Replays an ITF trace against a [`StepHandler`] and reports mismatches.
pub struct TraceReplayer<H: StepHandler> {
    handler: H,
}

impl<H: StepHandler> TraceReplayer<H> {
    pub fn new(handler: H) -> Self {
        Self { handler }
    }

    /// Replay a trace loaded from an ITF JSON file path.
    pub fn replay_file(&mut self, path: &str) -> Result<ReplayResult, ReplayError> {
        let trace = Trace::from_file(path)
            .map_err(|e| ReplayError::TraceLoad(e.to_string()))?;
        self.replay(&trace)
    }

    /// Replay a trace loaded from an ITF JSON string.
    pub fn replay_str(&mut self, json: &str) -> Result<ReplayResult, ReplayError> {
        let trace = Trace::from_str(json)
            .map_err(|e| ReplayError::TraceLoad(e.to_string()))?;
        self.replay(&trace)
    }

    /// Replay a pre-decoded [`Trace`].
    pub fn replay(&mut self, trace: &Trace) -> Result<ReplayResult, ReplayError> {
        let steps = trace.steps();
        let mut mismatches = Vec::new();

        // Step 0 is the initial state — check it before taking any action.
        if let Some(first) = steps.first() {
            self.check_state(first, &mut mismatches);
        }

        for step in steps.iter().skip(1) {
            if step.action.is_empty() {
                continue;
            }
            self.handler
                .apply_action(&step.action, &step.action_args)
                .map_err(|e| ReplayError::StepFailed {
                    step: step.index,
                    action: step.action.clone(),
                    source: e.to_string(),
                })?;

            self.check_state(step, &mut mismatches);
        }

        Ok(ReplayResult {
            steps_executed: steps.len(),
            mismatches,
        })
    }

    fn check_state(&self, step: &TraceStep, mismatches: &mut Vec<StateMismatch>) {
        let actual = self.handler.current_state();
        for (var, expected) in &step.variables {
            let got = actual.get(var).cloned().unwrap_or(ItfValue::Null);
            if &got != expected {
                mismatches.push(StateMismatch {
                    step: step.index,
                    action: step.action.clone(),
                    variable: var.clone(),
                    expected: expected.clone(),
                    actual: got,
                });
            }
        }
    }
}
