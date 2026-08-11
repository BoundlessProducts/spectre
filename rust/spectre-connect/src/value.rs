use std::collections::HashMap;

use serde_json::Value as JsonValue;

/// A Spectre runtime value decoded from an ITF trace.
///
/// This mirrors the Spectre runtime value types. The `Int` variant uses `i64`
/// for convenience; the ITF `#bigint` tag is decoded transparently.
#[derive(Debug, Clone, PartialEq)]
pub enum ItfValue {
    Int(i64),
    Float(f64),
    Bool(bool),
    Str(String),
    /// Set members — order is not significant.
    Set(Vec<ItfValue>),
    /// Ordered list / tuple.
    List(Vec<ItfValue>),
    /// Map entries as (key, value) pairs.
    Map(Vec<(ItfValue, ItfValue)>),
    /// Null / missing value.
    Null,
}

impl ItfValue {
    /// Decode a `serde_json::Value` produced by the ITF serializer into an [`ItfValue`].
    pub fn from_json(v: &JsonValue) -> Self {
        match v {
            JsonValue::Null => ItfValue::Null,
            JsonValue::Bool(b) => ItfValue::Bool(*b),
            JsonValue::Number(n) => {
                if let Some(i) = n.as_i64() {
                    ItfValue::Int(i)
                } else if let Some(f) = n.as_f64() {
                    ItfValue::Float(f)
                } else {
                    ItfValue::Null
                }
            }
            JsonValue::String(s) => ItfValue::Str(s.clone()),
            JsonValue::Array(arr) => {
                // Bare JSON arrays are not used in ITF — tagged objects are.
                // Treat as a list as a fallback.
                ItfValue::List(arr.iter().map(ItfValue::from_json).collect())
            }
            JsonValue::Object(obj) => decode_tagged_object(obj),
        }
    }

    /// Return the contained `i64`, or `None`.
    pub fn as_int(&self) -> Option<i64> {
        if let ItfValue::Int(n) = self { Some(*n) } else { None }
    }

    /// Return the contained `bool`, or `None`.
    pub fn as_bool(&self) -> Option<bool> {
        if let ItfValue::Bool(b) = self { Some(*b) } else { None }
    }

    /// Return the contained `&str`, or `None`.
    pub fn as_str(&self) -> Option<&str> {
        if let ItfValue::Str(s) = self { Some(s.as_str()) } else { None }
    }
}

// Conversion helpers for use in generated drivers --------------------------------

impl TryFrom<&ItfValue> for i64 {
    type Error = String;
    fn try_from(v: &ItfValue) -> Result<Self, Self::Error> {
        v.as_int().ok_or_else(|| format!("expected Int, got {v:?}"))
    }
}

impl TryFrom<&ItfValue> for bool {
    type Error = String;
    fn try_from(v: &ItfValue) -> Result<Self, Self::Error> {
        v.as_bool().ok_or_else(|| format!("expected Bool, got {v:?}"))
    }
}

impl TryFrom<&ItfValue> for String {
    type Error = String;
    fn try_from(v: &ItfValue) -> Result<Self, Self::Error> {
        v.as_str()
            .map(|s| s.to_string())
            .ok_or_else(|| format!("expected Str, got {v:?}"))
    }
}

impl TryFrom<&ItfValue> for f64 {
    type Error = String;
    fn try_from(v: &ItfValue) -> Result<Self, Self::Error> {
        match v {
            ItfValue::Float(f) => Ok(*f),
            ItfValue::Int(i) => Ok(*i as f64),
            _ => Err(format!("expected Float, got {v:?}")),
        }
    }
}

// ---- internal decoding -------------------------------------------------------

fn decode_tagged_object(obj: &serde_json::Map<String, JsonValue>) -> ItfValue {
    // ITF encodes integers as {"#bigint": "<decimal>"}
    if let Some(JsonValue::String(s)) = obj.get("#bigint") {
        return s.parse::<i64>()
            .map(ItfValue::Int)
            .unwrap_or(ItfValue::Str(s.clone()));
    }

    // Sets: {"#set": [...]}
    if let Some(JsonValue::Array(elems)) = obj.get("#set") {
        return ItfValue::Set(elems.iter().map(ItfValue::from_json).collect());
    }

    // Lists / tuples: {"#tup": [...]}
    if let Some(JsonValue::Array(elems)) = obj.get("#tup") {
        return ItfValue::List(elems.iter().map(ItfValue::from_json).collect());
    }

    // Maps: {"#map": [[key, val], ...]}
    if let Some(JsonValue::Array(entries)) = obj.get("#map") {
        let pairs = entries
            .iter()
            .filter_map(|entry| {
                if let JsonValue::Array(pair) = entry {
                    if pair.len() == 2 {
                        return Some((ItfValue::from_json(&pair[0]), ItfValue::from_json(&pair[1])));
                    }
                }
                None
            })
            .collect();
        return ItfValue::Map(pairs);
    }

    // Unknown tagged object — fall back to a map of string keys.
    let pairs = obj
        .iter()
        .map(|(k, v)| (ItfValue::Str(k.clone()), ItfValue::from_json(v)))
        .collect();
    ItfValue::Map(pairs)
}

/// Decode a flat variable map (one state entry) into a `HashMap<String, ItfValue>`,
/// skipping the reserved `#meta` key.
pub fn decode_state_vars(obj: &serde_json::Map<String, JsonValue>) -> HashMap<String, ItfValue> {
    obj.iter()
        .filter(|(k, _)| !k.starts_with('#'))
        .map(|(k, v)| (k.clone(), ItfValue::from_json(v)))
        .collect()
}
