//! Example: replay a Spectre counter trace against a Rust counter.
//!
//! Usage:
//!   spectre verify examples/counter.spec --emit-traces /tmp/counter.itf.json
//!   cargo run --example counter -- /tmp/counter.itf.json

use spectre_connect::{ItfValue, StepHandler, TraceReplayer};
use std::collections::HashMap;

struct Counter {
    value: i64,
}

impl StepHandler for Counter {
    type Error = Box<dyn std::error::Error>;

    fn apply_action(&mut self, action: &str, _args: &[ItfValue]) -> Result<(), Self::Error> {
        match action {
            "increment" => self.value += 1,
            "decrement" => {
                if self.value > 0 {
                    self.value -= 1;
                }
            }
            "reset" => self.value = 0,
            _ => return Err(format!("unknown action: {action}").into()),
        }
        Ok(())
    }

    fn current_state(&self) -> HashMap<String, ItfValue> {
        let mut s = HashMap::new();
        s.insert("counter".to_string(), ItfValue::Int(self.value));
        s
    }
}

fn main() {
    let path = std::env::args()
        .nth(1)
        .unwrap_or_else(|| "trace.itf.json".to_string());

    println!("Replaying trace: {path}");

    let counter = Counter { value: 0 };
    let mut replayer = TraceReplayer::new(counter);

    let result = replayer
        .replay_file(&path)
        .expect("failed to load/replay trace");

    if result.passed() {
        println!("✓ All {} steps matched the spec", result.steps_executed);
    } else {
        eprintln!("✗ {} mismatch(es) found:", result.mismatches.len());
        for m in &result.mismatches {
            eprintln!(
                "  step {}, action '{}', var '{}': expected {:?}, got {:?}",
                m.step, m.action, m.variable, m.expected, m.actual
            );
        }
        std::process::exit(1);
    }
}
