use std::fmt;

/// Errors produced by the trace replayer.
#[derive(Debug)]
pub enum ReplayError {
    /// Failed to load or parse the ITF trace.
    TraceLoad(String),
    /// The step handler returned an error when applying an action.
    StepFailed { step: usize, action: String, source: String },
}

impl fmt::Display for ReplayError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ReplayError::TraceLoad(msg) => write!(f, "trace load error: {msg}"),
            ReplayError::StepFailed { step, action, source } => {
                write!(f, "step {step} (action '{action}') failed: {source}")
            }
        }
    }
}

impl std::error::Error for ReplayError {}
