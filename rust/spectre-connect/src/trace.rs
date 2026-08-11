use std::collections::HashMap;

use serde::Deserialize;
use serde_json::Value as JsonValue;

use crate::value::{decode_state_vars, ItfValue};

/// Top-level ITF trace document.
#[derive(Debug, Deserialize)]
pub struct Trace {
    #[serde(rename = "#meta")]
    pub meta: TraceMeta,
    pub vars: Vec<String>,
    /// Raw JSON states — each is a flat object with `#meta` plus variable values.
    pub states: Vec<serde_json::Map<String, JsonValue>>,
}

/// Metadata attached to the whole trace.
#[derive(Debug, Deserialize)]
pub struct TraceMeta {
    pub format: String,
    #[serde(rename = "format-description", default)]
    pub format_description: String,
    #[serde(default)]
    pub source: String,
    #[serde(default)]
    pub description: String,
}

/// One decoded step in a trace.
#[derive(Debug, Clone)]
pub struct TraceStep {
    /// 0-based index of this step.
    pub index: usize,
    /// Name of the action that produced this state (empty for the initial state).
    pub action: String,
    /// Positional arguments that were passed to the action (may be empty).
    pub action_args: Vec<ItfValue>,
    /// Variable values in this state.
    pub variables: HashMap<String, ItfValue>,
}

impl Trace {
    /// Load a trace from an ITF JSON string.
    pub fn from_str(json: &str) -> Result<Self, serde_json::Error> {
        serde_json::from_str(json)
    }

    /// Load a trace from an ITF JSON file.
    pub fn from_file(path: &str) -> Result<Self, Box<dyn std::error::Error>> {
        let content = std::fs::read_to_string(path)?;
        Ok(Self::from_str(&content)?)
    }

    /// Decode all raw states into [`TraceStep`] values.
    pub fn steps(&self) -> Vec<TraceStep> {
        self.states
            .iter()
            .map(|raw| decode_step(raw))
            .collect()
    }
}

fn decode_step(raw: &serde_json::Map<String, JsonValue>) -> TraceStep {
    let mut index = 0usize;
    let mut action = String::new();
    let mut action_args = Vec::new();

    if let Some(JsonValue::Object(meta)) = raw.get("#meta") {
        if let Some(JsonValue::Number(n)) = meta.get("index") {
            index = n.as_u64().unwrap_or(0) as usize;
        }
        if let Some(JsonValue::String(a)) = meta.get("action") {
            action = a.clone();
        }
        if let Some(JsonValue::Array(args)) = meta.get("actionArgs") {
            action_args = args.iter().map(ItfValue::from_json).collect();
        }
    }

    TraceStep {
        index,
        action,
        action_args,
        variables: decode_state_vars(raw),
    }
}
