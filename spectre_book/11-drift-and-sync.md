# Chapter 11: Drift detection and incremental re-verification

### The drift problem

```
Day 1:   mine spec ──► check in account.spec + account.rs ✓
              │
Day 30:  refactor withdraw() in Rust
              │
              ▼
         spec and code no longer agree ← silent drift
              │
         next production incident reveals it
```

`spectre sync` catches drift before it becomes a bug report.

---

### Running sync

```bash
spectre sync src/account.rs --spec account.spec
```

Each action gets one of three symbols:

| Symbol | Meaning |
|--------|---------|
| `[=]` | Unchanged — guards and assignments are structurally identical |
| `[≡]` | SMT-proved equivalent — code changed, semantics identical, no re-verify needed |
| `[✗]` | Semantically changed — SMT found a counterexample; re-verify required |

Example output after a bug is introduced in `withdraw`:

```
[=]  action deposit    (structurally unchanged)
[✗]  action withdraw   assignment 1 body changed:
       "balance' = balance - amount"  →  "balance' = balance + amount"
       witness: amount=1, balance=0
[=]  action freeze     (structurally unchanged)
[≡]  action unfreeze   guard rewritten (SMT-proved equivalent):
       "not frozen"  →  "frozen = false"
```

The `[✗]` line gives you:
- **What changed** — which assignment or guard
- **The witness** — concrete values where the two versions behave differently

The `[≡]` prevents false alarms on equivalent rewrites — commutativity, constant folding, guard commutation.

---

### CI integration

```yaml
# .github/workflows/spectre.yml
- name: Check for spec drift
  run: spectre sync src/account.rs --spec account.spec
  # exits non-zero if any [✗] is found
```

`sync` exits 0 if all actions are `[=]` or `[≡]`; exits non-zero on any `[✗]`. Safe to gate a merge on.

---

### Incremental re-verification

After a `[✗]`, you don't need to re-run full BFS. Re-verify only the changed action:

```bash
spectre verify account.spec \
  --incremental \
  --changed-action withdraw \
  --use-cache
```

How it works (5 steps):

1. Load the cached state graph from the last full run
2. Remove all states reachable only through `withdraw`
3. Re-explore from states where `withdraw` was previously enabled
4. Re-check invariants on the new states only
5. Merge result back into the cache

Output:

```
Incremental re-verification: action 'withdraw'
  States re-explored: 12  (vs 847 in full BFS)
  ✓ safe — no new violations
  Cache updated.
```

12 states vs 847 = **70× faster** for this example. The saving grows with spec size.

---

### Impact analysis

Before re-verifying, check which invariants are actually affected by the change:

```bash
spectre impact account.spec
```

```
Changed actions (from git diff):
  withdraw — semantics changed (SMT witness: amount=1, balance=0)

Affected invariants:
  solvent         ← balance can go negative with the new body
  not_over_limit  ← not directly affected by withdraw

Recommended:
  spectre verify account.spec --incremental --changed-action withdraw
```

Use `impact` in CI to skip re-verification when a changed action doesn't touch any safety-critical invariant.

---

### Full sync workflow

```
Code changes committed
        │
        ▼
spectre sync src/ --spec spec.spec
        │
  [=] or [≡] only?
        ├── yes ──► nothing to do ✓
        └── no ([✗] found)
                │
                ▼
        spectre impact spec.spec
                │
        safety-critical invariant affected?
                ├── no ──► log warning, skip re-verify
                └── yes
                        │
                        ▼
                spectre verify spec.spec \
                  --incremental \
                  --changed-action NAME \
                  --use-cache
                        │
                   safe? ──yes──► commit + update cache ✓
                        │
                        no
                        │
                        ▼
                Fix code or spec, re-verify
```

---

### Sync evaluation: what it checks

`sync` compares:
- Struct field names and types
- Action guard counts and expressions
- Assignment-body expressions (normalised)

It detected **30/35 injected semantic mutations** (85.7%) in evaluation, with 1 false positive (vs 4 with AST comparison alone). The `[≡]` classification eliminates the majority of false alarms.

---

### Tips for keeping sync useful

- **Check in the spec file alongside the Rust source** — keep them in the same directory or a `spec/` subdirectory
- **Run sync in pre-commit** as well as CI — catch drift before it even reaches the repo
- **After a `[✗]`**: always re-verify before merging, even if you think the change is safe
- **After a `[≡]`**: no action required — but it's worth reading the rewrite to confirm it's intentional
