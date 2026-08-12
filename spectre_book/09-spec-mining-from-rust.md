# Chapter 9: Spec Mining, Drift Detection, and Incremental Re-verification

This chapter covers three related commands: `spectre mine`, `spectre sync`, and `spectre verify --incremental`.

---

## How Spec Mining Works

The `spectre mine --lang rust` command invokes a companion binary (`spectre-mine-rs`) that parses your Rust source using the `syn` AST library — the same parser used by `serde`, `tokio`, and most of the Rust ecosystem.  It traverses:

- **`struct` definitions** → Spectre `var` declarations with mapped types
- **`impl` blocks** → Spectre `action` blocks
- **`assert!` / `debug_assert!` / `assert_eq!` / `assert_ne!` calls** → `require` guards
- **`if cond { return Err(...) }` early-return patterns** → negated `require` guards
- **Self-field assignments** (`self.x = expr`, `self.x += expr`, …) → primed assignments (`x' = x + expr`)
- **`fn new(...)` constructor literal** → `init` values

The result is a specification that you can verify immediately.

---

## Example: Mining a Bank Account

Consider the following Rust implementation (`examples/rust/bank_account.rs`):

```rust
pub struct BankAccount {
    pub balance: i64,
    pub owner:   String,
    pub frozen:  bool,
}

impl BankAccount {
    pub fn new(owner: String) -> Self {
        Self { balance: 0, owner, frozen: false }
    }

    pub fn deposit(&mut self, amount: i64) {
        assert!(amount > 0);
        assert!(self.balance + amount <= 1_000_000);
        self.balance += amount;
    }

    pub fn withdraw(&mut self, amount: i64) {
        assert!(amount > 0);
        assert!(self.balance >= amount);
        assert!(!self.frozen);
        self.balance -= amount;
    }

    pub fn freeze(&mut self)   { assert!(!self.frozen); self.frozen = true;  }
    pub fn unfreeze(&mut self) { assert!(self.frozen);  self.frozen = false; }
}
```

Run the miner:

```bash
spectre mine --lang rust examples/rust/bank_account.rs
```

Output:

```spectre
var balance: int
var owner: str
var frozen: bool

init {
  balance = 0
  owner = ""
  frozen = false
}

action deposit(amount: int) {
  require amount > 0
  require balance + amount <= 1000000
  balance' = balance + amount
}

action withdraw(amount: int) {
  require amount > 0
  require balance >= amount
  require !frozen
  balance' = balance - amount
}

action freeze {
  require !frozen
  frozen' = true
}

action unfreeze {
  require frozen
  frozen' = false
}
```

All seven `assert!` guards were extracted, including the compound condition `balance + amount <= 1_000_000` (numeric underscore stripped).  Constructor init values (`balance = 0`, `frozen = false`) came from the `new` method's struct literal.

---

## Saving and Verifying the Mined Spec

```bash
spectre mine --lang rust examples/rust/bank_account.rs \
    --output examples/bank-account-mined.spec

spectre verify examples/bank-account-mined.spec
# Traversed 3745 states
# Found no violations.
```

The spec as mined has no `invariant` blocks — those are your design intent, not extractable from assertions alone. Add them before the final verify:

```spectre
invariant non_negative {
  balance >= 0
}

invariant within_cap {
  balance <= 1000000
}
```

---

## What the Miner Extracts — and What It Does Not

| Construct | Extracted? | Notes |
|-----------|-----------|-------|
| `struct` fields with primitive types | ✅ | `i64`→`int`, `bool`→`bool`, `String`→`str` |
| `Vec<T>` fields | ✅ | mapped to `List[T]` |
| `HashMap<K,V>` fields | ✅ | mapped to `Map[K,V]` |
| `Arc<Mutex<T>>` fields | ✅ | inner type unwrapped |
| `Option<T>` fields | ✅ | inner type used |
| `assert!` / `debug_assert!` guards | ✅ | all positions |
| `assert_eq!(a, b)` | ✅ | extracted as `require a = b` |
| `assert_ne!(a, b)` | ✅ | extracted as `require a != b` |
| `if cond { return Err(..) }` guards | ✅ | extracted as `require cond` |
| Constructor init values from `fn new` | ✅ | struct literal fields |
| Self-field assignments (`self.x = …`) | ✅ | primed assignments |
| Compound assignments (`+=`, `-=`, `*=`) | ✅ | expanded to `x' = x + …` |
| Tuple / struct enum variants | ❌ | flagged with `// TODO` comment |
| Unknown types | ❌ | fall back to `int` |
| Complex control flow (nested match, iterators) | ❌ | partial extraction; needs manual completion |

---

## Drift Detection with `spectre sync`

Once the spec exists, the code and spec can diverge silently as the codebase evolves. `spectre sync` compares the current Rust source against the existing spec using **SMT-backed equivalence checking** (Z3, QF_LIA theory):

```bash
spectre sync examples/rust/bank_account.rs --spec examples/bank-account-mined.spec
```

Each action is classified with one of three symbols:

| Symbol | Meaning |
|--------|---------|
| `[=]` | Structurally unchanged — guards and assignments are identical |
| `[≡]` | SMT-proved equivalent — the code changed but Z3 proves the semantics are identical (safe rewrite, no re-verification needed) |
| `[✗]` | Semantically changed — SMT found a counterexample witness; re-verification required |

**Example output (semantic mutation: `balance - amount` → `balance + amount`):**

```
[=] action deposit (structurally unchanged)
[✗] action withdraw assignment 1 body changed: "balance'=balance - amount" → "balance'=balance + amount" (witness: amount=1, balance=0)
[=] action freeze (structurally unchanged)
[=] action unfreeze (structurally unchanged)
```

**Example output (equivalent rewrite: guard commuted):**

```
[≡] action deposit guard 2 rewritten (SMT-proved equivalent): "balance + amount <= 1000000" → "amount + balance <= 1000000"
[=] action withdraw (structurally unchanged)
```

The `[≡]` classification prevents false alarms on logically-equivalent refactors (e.g., commutativity, constant folding). In evaluation: **8/8 semantic mutations detected** with **1 false positive** (vs. 4 FP using AST comparison alone).

---

## Incremental Re-verification

When `spectre sync` flags an action as **possibly-unsafe**, the safe path is to re-verify the entire specification. For large models this can be expensive. `spectre verify --incremental` re-verifies only the changed action:

```bash
# Step 1: build and cache the state graph (first run only)
spectre verify examples/raft-election-safety.spec --max-states 30000 --use-cache

# Step 2: after code changes, sync to find what changed
spectre sync examples/rust/raft.rs
# Output: vote2for1  possibly-unsafe

# Step 3: re-verify only vote2for1
spectre verify examples/raft-election-safety.spec --max-states 30000 \
    --use-cache --incremental --changed-action vote2for1
```

### The Algorithm (5 steps)

1. **Prune** — remove all transitions produced by the changed action from the cached graph
2. **Reachability BFS** — recompute which states are reachable from initial states over the remaining edges
3. **Remove unreachable states** — discard states that were only reachable via the changed action
4. **Re-execute** — run the changed action from every remaining reachable state; collect new successor states
5. **BFS-expand** — BFS-expand new successor states with all actions, capped at the original state count

The result is provably equivalent to a fresh BFS of the modified spec under the same exploration bound (Proposition: Incremental Equivalence).

### Benchmark

All timings on Apple M3 Pro (18 GB, macOS 15.6, Go 1.24). Absolute times are hardware-dependent; speedup ratios are the primary claims.

| Spec | Changed action | Cold BFS | Incremental | Speedup | States pruned / new |
|------|---------------|----------|-------------|---------|---------------------|
| Raft (22,432) | `stepDown1_sees2` | 12.4 s | 2.4 s | **5.2×** | 1,649 / 1,645 |
| Raft (22,432) | `vote2for1` | 12.4 s | 2.9 s | **4.3×** | 6,501 / 6,495 |

Cache restore (unchanged spec): 0.265 s for Raft (47× faster than cold BFS).

Speedup depends on the fraction of states reachable only via the changed action. When that fraction is high (e.g., `freeze` owns 1,882 of 3,745 states), most of the graph can be reused without re-expansion.

---

## Impact Analysis with `spectre impact`

`spectre impact` reads the current `git diff` and maps changed Rust functions to the actions they implement, then reports which spec actions and invariants are affected and whether incremental re-verification is needed:

```bash
spectre impact examples/raft-election-safety.spec
```

Example output after modifying `vote2for1` in Rust source:

```
Changed actions (from git diff):
  vote2for1  — semantics changed (SMT witness available)

Affected invariants:
  electionSafety
  leaderMajority

Recommended:
  spectre verify examples/raft-election-safety.spec --use-cache --incremental --changed-action vote2for1
```

`impact` is designed for CI: it exits non-zero if any changed action touches a safety-critical invariant, so pipelines can gate on it.

---

## The Full Workflow

```
mine → verify → check-in spec → evolving code → sync → incremental re-verify
```

1. **Mine** — extract the spec skeleton from Rust (`spectre mine`)
2. **Enrich** — add invariants, temporal properties, verify and iterate
3. **Check in** — commit the `.spec` file alongside the Rust source
4. **Evolve** — modify the Rust implementation as needed
5. **Sync** — run `spectre sync` in CI to detect drift; exit non-zero on `possibly-unsafe`
6. **Re-verify** — run `spectre verify --incremental --changed-action <name>` for fast re-check

For bidirectional drift detection (when the spec itself may also have changed independently):

```bash
spectre drift account.spec examples/rust/bank_account.rs
```

---

## LLM Enhancement

With the optional `--ai` flag, the mined spec is sent to an LLM together with the Rust source for semantic enrichment. The LLM proposes:

- One-sentence `description` for each action
- Up to three additional `invariant` candidates (commented-out suggestions)
- Up to three `temporal` property suggestions

```bash
ANTHROPIC_API_KEY=sk-... spectre mine --lang rust src/wallet.rs --ai
```

The LLM additions are always commented out and require explicit human review before taking effect.
