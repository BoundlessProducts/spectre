//! # spectre-connect
//!
//! Model-based testing bridge between [Spectre](https://github.com/akkeshavan/spectre)
//! specifications and Rust implementations.
//!
//! ## Workflow
//!
//! 1. Write a Spectre spec and run `spectre verify myspec.spec --emit-traces trace.itf.json`
//!    to produce an ITF execution trace.
//! 2. Generate a driver skeleton: `spectre generate-driver --lang rust myspec.spec > driver.rs`
//! 3. Fill in the `apply_action` and `current_state` methods to connect the driver
//!    to your real implementation.
//! 4. Run the driver against the trace: `./driver trace.itf.json`
//!
//! ## Quick example
//!
//! ```rust,ignore
//! use spectre_connect::{ItfValue, StepHandler, TraceReplayer};
//! use std::collections::HashMap;
//!
//! struct Counter { value: i64 }
//!
//! impl StepHandler for Counter {
//!     type Error = Box<dyn std::error::Error>;
//!
//!     fn apply_action(&mut self, action: &str, _args: &[ItfValue]) -> Result<(), Self::Error> {
//!         match action {
//!             "increment" => self.value += 1,
//!             "decrement" => self.value -= 1,
//!             "reset"     => self.value = 0,
//!             _ => return Err(format!("unknown action: {action}").into()),
//!         }
//!         Ok(())
//!     }
//!
//!     fn current_state(&self) -> HashMap<String, ItfValue> {
//!         let mut s = HashMap::new();
//!         s.insert("counter".into(), ItfValue::Int(self.value));
//!         s
//!     }
//! }
//!
//! fn main() {
//!     let mut replayer = TraceReplayer::new(Counter { value: 0 });
//!     let result = replayer.replay_file("trace.itf.json").unwrap();
//!     assert!(result.passed(), "implementation diverged from spec");
//! }
//! ```

pub mod error;
pub mod replayer;
pub mod trace;
pub mod value;

pub use error::ReplayError;
pub use replayer::{ReplayResult, StateMismatch, StepHandler, TraceReplayer};
pub use trace::{Trace, TraceMeta, TraceStep};
pub use value::ItfValue;
