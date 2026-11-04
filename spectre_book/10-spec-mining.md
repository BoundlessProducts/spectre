# Chapter 10: Mining specs from Rust

### What mining does

```
src/account.rs ──► spectre mine ──► account.spec skeleton
                                         │
                                    add invariants
                                         │
                                    spectre verify
```

The miner reads your Rust source and extracts struct fields, assert guards, and field assignments — giving you a runnable spec in seconds.

What it extracts:
- `struct` fields → `var` declarations
- `assert!` / `debug_assert!` → `require` guards
- `if cond { return Err(..) }` → negated `require` guards
- `self.field = expr` / `self.field += n` → primed assignments
- `fn new(...)` constructor → `init` values
- Rust enums → Spectre enums

---

### Your first mine

```rust
// src/counter.rs
pub struct Counter {
    pub value: i64,
    pub max:   i64,
}

impl Counter {
    pub fn new(max: i64) -> Self {
        Counter { value: 0, max }
    }
    pub fn increment(&mut self) {
        assert!(self.value < self.max);
        self.value += 1;
    }
    pub fn decrement(&mut self) {
        assert!(self.value > 0);
        self.value -= 1;
    }
    pub fn reset(&mut self) {
        self.value = 0;
    }
}
```

```bash
spectre mine --lang rust src/counter.rs
```

Output:

```spectre
// Spectre spec mined from: src/counter.rs

var value: int
var max:   int

init {
  value = 0
  max   = 0
}

action increment {
  require value < max
  value' = value + 1
}

action decrement {
  require value > 0
  value' = value - 1
}

action reset {
  value' = 0
}
```

Save it and verify immediately:

```bash
spectre mine --lang rust src/counter.rs -o counter.spec
spectre verify counter.spec
# Exhaustive verification: traversed 1 states (full reachable space)
# Found no violations.
# (max=0 by default — set max to a real value before verifying)
```

---

### What gets extracted vs. what gets skipped

| Extracted ✓ | Skipped / approximated ✗ |
|-------------|--------------------------|
| Primitive fields (`i64`, `bool`, `String`) | `new()`, `default()`, `clone()`, `fmt()` |
| Rust enums — unit variants only | Tuple/struct enum variants (data dropped) |
| `assert!` / `debug_assert!` guards | `panic!`, `unwrap()`, `expect()` |
| `self.x = …`, `self.x += …`, `self.x -= …` | `checked_add`, `wrapping_add`, `saturating_sub` |
| `if cond { return Err(..) }` patterns | `Result<T,E>` error paths |
| `Arc<T>`, `Mutex<T>`, `Box<T>` wrappers | (inner type extracted, wrapper dropped) |

---

### Seeing what was skipped: `--show-omissions`

```bash
spectre mine --lang rust src/account.rs --show-omissions
```

```
Extraction omissions (4):
  SKIPPED_METHOD in new: constructor/utility method — not translated to action
  RESULT_RETURN in deposit: Result<(),TxError> return — error path abstracted
  OVERFLOW_OP in deposit: checked_add — saturating arithmetic abstracted
  TUPLE_VARIANT in TxError: InsufficientFunds(i64) — data dropped, unit variant kept

Mined: 3 fields, 1 enum, 4 actions, 4 omissions
```

Four omission kinds to act on:

- **RESULT_RETURN** — error paths not modelled. Add invariants for the error conditions manually if they matter.
- **OVERFLOW_OP** — arithmetic is unbounded in the spec. Add `require` guards if overflow-safe bounds are important.
- **TUPLE_VARIANT** — enum variant data was dropped. Add a `var` for the data if you need to reason about it.
- **SKIPPED_METHOD** — constructors, `fmt`, `drop`. Almost always safe to ignore.

The mine summary line shows how many omissions occurred:

```
Mined: 3 fields, 1 enum, 4 actions, 4 omissions (run --show-omissions for details)
```

---

### Evaluate a whole directory

```bash
spectre eval-rust --dir src/
```

```
# Rust Extraction Evaluation Report

## Per-File Summary

| File              | Struct      | Fields | Enums | Methods | Omissions | Success |
|-------------------|-------------|--------|-------|---------|-----------|---------|
| account.rs        | BankAccount |   3    |   1   |    4    |     4     |   50%   |
| counter.rs        | Counter     |   2    |   0   |    3    |     1     |  100%   |
| state_machine.rs  | Machine     |   4    |   2   |    2    |     4     |   50%   |

**Totals** — files: 3 | fields: 9 | methods: 9 | omissions: 9 | mean success: 67%

## Omissions by Kind

| Kind            | Count |
|-----------------|-------|
| SKIPPED_METHOD  |   3   |
| RESULT_RETURN   |   2   |
| OVERFLOW_OP     |   2   |
| TUPLE_VARIANT   |   2   |
```

**Success fraction** = methods with no body-level omissions (overflow / panic / Result / complex assign). Files at 100% can be verified with minimal manual work. Files at 50% need review.

Use this output to prioritise enrichment effort — start with files nearest 100%.

---

### Enrich the mined spec

Three things to always add after mining:

1. **Invariants** — the miner extracts no invariants (only guards). Add at least one per resource, one per ordering constraint.
2. **Temporal properties** — if your system makes progress guarantees, add `WF(action)` or `eventually(goal)`.
3. **Realistic init values** — `max = 0` won't verify anything useful. Set values that match production defaults.

Example enrichment for the counter spec:

```spectre
// counter.spec — after enrichment
var value: int
var max:   int

init {
  value = 0
  max   = 10    // ← set a realistic default
}

action increment {
  require value < max
  value' = value + 1
}

action decrement {
  require value > 0
  value' = value - 1
}

action reset {
  require value > 0    // guard prevents stutter when value is already 0
  value' = 0
}

// ← these three lines are the additions
invariant non_negative  { value >= 0 }
invariant within_bounds { value <= max }
temporal  progress      { WF(increment) }
```

---

### LLM-enhanced mining

```bash
ANTHROPIC_API_KEY=sk-... spectre mine --lang rust src/account.rs --ai
```

The LLM adds (always commented out — never applied automatically):
- One-sentence `description` for each action
- Up to three `invariant` candidates
- Up to three `temporal` property candidates

```spectre
// LLM-suggested invariants (review before enabling):
// invariant balanced_books { total_deposits >= total_withdrawals }
// invariant no_freeze_with_balance { frozen → balance = 0 }
```

Always review LLM suggestions against your actual requirements before uncommenting.

---

### The full mining workflow

```
1. spectre eval-rust --dir src/
   └── identify which files mine cleanly (Success = 100%)

2. spectre mine --lang rust src/target.rs -o target.spec
   └── get the skeleton

3. spectre mine ... --show-omissions
   └── understand what was skipped

4. Edit target.spec
   └── add invariants, fix init values, handle omissions

5. spectre verify target.spec
   └── safe → check in both files
       violation → fix spec, re-verify

6. spectre sync src/target.rs --spec target.spec  (ongoing)
   └── catch drift as code evolves
```
