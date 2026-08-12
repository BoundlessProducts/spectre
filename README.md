> **⚠️ Under Construction** — This is a research project. Not intended for commercial use. No PRs encouraged at this point.
>
> Some parts of the code were generated using LLMs.

---

# Spectre

A formal specification language and Rust verification toolchain inspired by TLA+ and Quint.

Model your system as a state machine, write invariants and temporal properties, and let Spectre exhaustively explore every reachable state. A refinement map connects your Rust implementation to the formal model, covering verification, counterexample-guided repair, conformance testing, drift detection, incremental re-verification, and runtime monitoring.

```spectre
var balance: int
var frozen: bool

init { balance = 0  frozen = false }

action deposit(amount: int) {
  require amount > 0
  balance' = balance + amount
}
action withdraw(amount: int) {
  require amount > 0
  require balance >= amount
  require frozen == false
  balance' = balance - amount
}
action freeze   { require frozen == false  frozen' = true  }
action unfreeze { require frozen == true   frozen' = false }

invariant solvency { balance >= 0 }
temporal eventuallyUnfreeze { always (frozen == true -> eventually frozen == false) }
```

---

## Table of Contents

1. [What is Spectre?](#what-is-spectre)
2. [Installation — macOS](#installation--macos)
3. [Installation — Linux](#installation--linux)
4. [Installation — Windows](#installation--windows)
5. [Running the Examples](#running-the-examples)
6. [CLI Reference](#cli-reference)
7. [Advanced Workflows](#advanced-workflows)
8. [Documentation](#documentation)

---

## What is Spectre?

Spectre is a formal specification language for modelling systems as state machines and verifying their correctness properties. It explores your system's entire reachable state space and reports exactly which sequence of actions leads to a bug.

**Key capabilities:**

| Feature | Description |
|---------|-------------|
| State machines | Declare state variables, an initial state, and actions that transition between states |
| Invariant checking | Assert properties that must hold in *every* reachable state |
| Temporal logic | Express liveness (`eventually`), safety (`always`), and response (`→` leads-to) properties |
| Fairness | Weak (`WF`) and strong (`SF`) fairness conditions for concurrent systems |
| CEGIS repair | Automatic weakest-precondition guard suggestions when an invariant is violated |
| Partial-order reduction | `--por` heuristic skips redundant interleavings — 3-node Raft: 22,432 → 6,818 states (7.6× speedup) |
| Property-directed exploration | `--property-guided` best-first BFS toward invariant boundaries — finds all 29 violation traces in 204 states |
| Spec mining | Extract a Spectre skeleton from existing Rust source — fields, actions, guards, and init values |
| SMT drift detection | `spectre sync` uses Z3/QF_LIA to classify each change as `[=]` unchanged, `[≡]` SMT-proved-equivalent, or `[✗]` semantically changed |
| Incremental re-verification | When one action changes, `--incremental` re-verifies only that action — 4.3–5.2× faster than a cold BFS |
| State caching | `--use-cache` restores a previous BFS graph in milliseconds (47× faster for Raft) |
| Model-based testing | Generate Rust test drivers and ITF trace replay; five coverage modes including property-directed |
| Rust monitor generation | Emit a self-contained `monitor.rs` that checks spec invariants in production |
| Simulation | Direct path sampling via `spectre simulate` — covers 5-/7-node Raft beyond exhaustive BFS scale |

---

## Installation — macOS

### Prerequisites

| Dependency | Required for | Install |
|------------|-------------|---------|
| **Go ≥ 1.24** | Core CLI (`spectre verify`, `sync`, etc.) | `brew install go` or [go.dev/dl](https://go.dev/dl/) |
| **Z3 ≥ 4.8** | `spectre sync` (SMT equivalence checking) | `brew install z3` |
| **Rust stable** | `spectre mine --lang rust` (spec mining) | [rustup.rs](https://rustup.rs) |

Z3 and Rust are optional if you only use `verify`, `simulate`, `generate-monitor`, and `generate-driver`.

### Step 1 — Install Go

```bash
brew install go
go version
# Expected: go version go1.24.x darwin/arm64  (or amd64)
```

Or download the `.pkg` installer from [https://go.dev/dl/](https://go.dev/dl/).

### Step 2 — Install Z3 (for `spectre sync`)

```bash
brew install z3
z3 --version
```

### Step 3 — Clone and build the CLI

```bash
git clone https://github.com/BoundlessProducts/spectre.git
cd spectre
go build -o spectre ./cmd/spectre
./spectre
```

### Step 4 — Build `spectre-mine-rs` (for `spectre mine --lang rust`)

```bash
# Install Rust if needed
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"

# Build the Rust AST miner
cargo build --release --manifest-path rust/spectre-mine-rs/Cargo.toml
cp rust/spectre-mine-rs/target/release/spectre-mine-rs .
```

The `spectre-mine-rs` binary must be in the same directory as `spectre` (or on `PATH`).

### Step 5 — (Optional) Install globally

```bash
go install ./cmd/spectre
cp spectre-mine-rs "$(go env GOPATH)/bin/"
export PATH="$PATH:$(go env GOPATH)/bin"
# Add to ~/.zshrc or ~/.bash_profile to persist
```

### Step 6 — Run the test suite (optional)

```bash
go test ./...
```

---

## Installation — Linux

### Prerequisites

| Dependency | Required for | Install |
|------------|-------------|---------|
| **Go ≥ 1.24** | Core CLI | see below |
| **Z3 ≥ 4.8** | `spectre sync` | `sudo apt install z3` / `sudo dnf install z3` |
| **Rust stable** | `spectre mine --lang rust` | [rustup.rs](https://rustup.rs) |

### Step 1 — Install Go

**Debian / Ubuntu**

```bash
wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

> For ARM64 (e.g., Raspberry Pi, AWS Graviton): replace `amd64` with `arm64` in the filename.

**Fedora / RHEL / CentOS**

```bash
sudo dnf install -y golang
go version
```

**Arch Linux**

```bash
sudo pacman -S go
go version
```

### Step 2 — Install Z3 (for `spectre sync`)

```bash
# Debian / Ubuntu
sudo apt install -y z3

# Fedora / RHEL
sudo dnf install -y z3

# Arch
sudo pacman -S z3
```

### Step 3 — Clone and build

```bash
git clone https://github.com/BoundlessProducts/spectre.git
cd spectre
go build -o spectre ./cmd/spectre
go test ./...              # optional: run test suite
go install ./cmd/spectre   # optional: install globally
```

### Step 4 — Build `spectre-mine-rs` (for `spectre mine --lang rust`)

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"
cargo build --release --manifest-path rust/spectre-mine-rs/Cargo.toml
cp rust/spectre-mine-rs/target/release/spectre-mine-rs .
```

---

## Installation — Windows

### Step 1 — Install Go

1. Go to [https://go.dev/dl/](https://go.dev/dl/) and download the latest **Windows installer** (`.msi`).
2. Run the installer. It adds Go to your `PATH` automatically.
3. Open a new **Command Prompt** or **PowerShell** and verify:

```powershell
go version
```

**Alternative — winget**

```powershell
winget install GoLang.Go
```

### Step 2 — Install Z3 (for `spectre sync`)

Download the Z3 Windows binary from [https://github.com/Z3Prover/z3/releases](https://github.com/Z3Prover/z3/releases) and add the `bin/` folder to your `PATH`.

### Step 3 — Clone and build

```powershell
git clone https://github.com/BoundlessProducts/spectre.git
cd spectre
go build -o spectre.exe ./cmd/spectre
.\spectre.exe
go install ./cmd/spectre   # optional: install globally
```

### Step 4 — Build `spectre-mine-rs` (for `spectre mine --lang rust`)

Install Rust from [https://rustup.rs](https://rustup.rs), then:

```powershell
cargo build --release --manifest-path rust\spectre-mine-rs\Cargo.toml
copy rust\spectre-mine-rs\target\release\spectre-mine-rs.exe .
```

> **Windows note:** All commands below use `./spectre` (macOS/Linux). On Windows substitute `.\spectre.exe` (PowerShell) or `./spectre.exe` (Git Bash).

---

## Running the Examples

All examples are in the `examples/` directory.

---

### Example 1 — Hello World: Counter

```bash
./spectre verify examples/counter.spec
```

The counter's `reset` action prevents it from ever reaching 10, so the temporal property is violated. Spectre shows the counterexample trace.

```bash
./spectre verify examples/counter-corrected.spec   # should pass
```

---

### Example 2 — Missing Guard: Bank Account

```bash
./spectre verify examples/bank-account-violation.spec
# Finds invariant violation in one step: withdrawing before depositing
```

```bash
./spectre verify examples/bank-account-corrected.spec
# Passes: no violations found within the exploration bounds
```

**With parameterized actions (used for MBT and monitoring):**

```bash
./spectre verify examples/bank-account-parameterized.spec
# 3,745 states exhaustively explored; no violations
```

---

### Example 3 — Mutual Exclusion

```bash
./spectre verify examples/concurrent-lock-violation.spec
./spectre verify examples/concurrent-lock-corrected.spec
# Corrected version: 1,928 states, no violations
```

---

### Example 4 — Raft Election Safety (3-node cluster)

Spectre mines the Raft per-node state machine from Rust source and composes a 3-node cluster spec.

```bash
./spectre verify examples/raft-election-safety.spec --max-states 30000
# Exhaustively explores 22,432 states; no violations of
# electionSafety, leaderMajority, or candidateSelfVote
```

**Cache restore (unchanged spec, 33× faster):**

```bash
./spectre verify examples/raft-election-safety.spec --max-states 30000 --use-cache
# Restored state graph from cache (22432 states)
```

**Incremental re-verification when one action changes:**

```bash
./spectre verify examples/raft-election-safety.spec --max-states 30000 \
    --use-cache --incremental --changed-action vote2for1
# Re-verifies only vote2for1; ~3× faster than cold BFS
```

---

### Example 5 — Stuttering and Fairness

```bash
./spectre verify examples/stuttering-counter.spec
# Reports stuttering warnings for actions that self-loop

./spectre verify examples/stuttering-counter-with-fairness.spec
# Fairness constraints resolve the liveness issue
```

---

### Example 6 — Coverage-Guided Model-Based Testing

```bash
# Property-directed: steers traces toward invariant boundaries
./spectre verify examples/bank-account-parameterized.spec \
    --emit-traces trace.itf.json --coverage-mode property

# Other modes: action | transition-pair | boundary | rare-action
./spectre verify examples/bank-account-parameterized.spec \
    --emit-traces trace.itf.json --coverage-mode boundary
```

---

### Example 7 — Exploration Limits

```bash
# Explore up to 50,000 states
./spectre verify examples/bank-account-corrected.spec --max-states 50000

# No state limit (for small specs only)
./spectre verify examples/counter.spec --max-states unlimited

# Parameterised model (e.g., 5-node Raft)
./spectre verify examples/raft-election-safety-5node.spec --max-states 50000

# Direct simulation (no full graph): 1,000 random paths on 7-node Raft
./spectre simulate examples/raft-election-safety-7node.spec --traces 1000
```

---

## CLI Reference

```
spectre <command> [flags] <file>
```

| Command | Description |
|---------|-------------|
| `parse <file>` | Check syntax — report parse errors |
| `typecheck <file>` | Check types — report type mismatches |
| `verify <file>` | BFS state exploration, invariant and temporal property checking |
| `simulate <file>` | Random path sampling without building the full state graph |
| `mine --lang rust <source.rs>` | Mine a Spectre spec skeleton from Rust source |
| `sync <source.rs> --spec <spec>` | SMT-backed drift detection; classifies each action as `[=]` / `[≡]` / `[✗]` |
| `drift <spec> <source.rs>` | Detect bidirectional staleness between a spec and its Rust implementation |
| `generate-driver --lang rust <file>` | Generate a Rust MBT driver skeleton |
| `generate-monitor --lang rust <file>` | Generate an embedded Rust runtime monitor |
| `check-refinement <file>` | Check that ITF traces conform to the spec via the refinement mapping |
| `diff-conformance <file>` | Replay an ITF trace against two spec versions and report divergences |
| `impact <spec>` | Given a git diff, report affected actions/invariants and incremental verification status |

### `verify` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--verbose`, `-v` | off | Show detailed state dumps in traces |
| `--max-states <n>` | 5000 | Stop after exploring `n` states. Use `unlimited` for no limit |
| `--max-depth <n>` | 100 | Stop BFS at depth `n`. Use `unlimited` for no limit |
| `--param Name=Value` | — | Bind a spec parameter (e.g., `--param N=3`). Repeatable |
| `--por` | off | Partial-order reduction: skip redundant action interleavings (ample-set heuristic) |
| `--property-guided` | off | Best-first BFS toward invariant boundaries — finds violations with far fewer states explored |
| `--emit-traces <file>` | — | Write an ITF execution trace to `<file>` for MBT replay |
| `--coverage-mode <mode>` | `action` | Trace coverage strategy: `action`, `transition-pair`, `boundary`, `rare-action`, `property` |
| `--use-cache` | off | Restore a previously cached BFS graph; run fresh BFS and save cache if absent |
| `--incremental` | off | Re-verify only the changed action (requires `--use-cache` and `--changed-action`) |
| `--changed-action <name>` | — | Name of the action whose semantics changed |

### `simulate` flags

| Flag | Default | Description |
|------|---------|-------------|
| `--traces <n>` | 100 | Number of random paths to sample |
| `--max-steps <n>` | 100 | Maximum steps per path (also accepted as `--max-depth`) |
| `--seed <n>` | — | Random seed for reproducible traces |
| `--output-dir <dir>` | — | Write each trace as a separate ITF JSON file to `<dir>` |
| `--use-bfs` | off | BFS walk instead of random walk |

### `mine` flags

| Flag | Description |
|------|-------------|
| `--lang rust` | Source language (only `rust` supported) |
| `--output`, `-o <file>` | Write mined spec to file instead of stdout |
| `--ai` | Enhance with LLM (requires `ANTHROPIC_API_KEY` env var) |

---

## Advanced Workflows

### Spec Mining + Drift Detection

Extract a spec from Rust source, then keep it in sync as the code evolves:

```bash
# 1. Mine the spec
./spectre mine --lang rust src/account.rs -o account.spec

# 2. Verify it
./spectre verify account.spec

# 3. After code changes, check for drift (SMT-backed)
./spectre sync src/account.rs --spec account.spec
# [=]  deposit     (structurally unchanged)
# [≡]  withdraw    (SMT-proved equivalent rewrite — no re-verification needed)
# [✗]  freeze      (semantically changed — witness: frozen=false; re-verify required)

# 4. For bidirectional drift (spec also may have changed)
./spectre drift account.spec src/account.rs
```

`spectre sync` uses Z3/QF_LIA to classify each change. `[✗]` includes a counterexample witness. In evaluation: 8/8 semantic mutations detected with 1 false positive (vs. 4 FP without SMT).

---

### Incremental Re-verification

After `spectre sync` flags an action as **possibly-unsafe**, re-verify only that action without re-running the full BFS:

```bash
# First run: build and cache the state graph
./spectre verify raft.spec --max-states 30000 --use-cache

# After modifying vote2for1's guard:
./spectre verify raft.spec --max-states 30000 \
    --use-cache --incremental --changed-action vote2for1
```

The algorithm: prune stale transitions → recompute reachability → remove now-unreachable states → re-execute the changed action → BFS-expand new states up to the original bound. On Raft (22,432 states), incremental is **4.3–5.2× faster** than cold BFS; cache restore for an unchanged spec is **47× faster**. (Timings on Apple M3 Pro; ratios are hardware-independent.)

---

### Partial-Order Reduction and Property-Directed Exploration

Two flags that can dramatically reduce the number of states explored:

```bash
# POR: skip redundant action interleavings
# 3-node Raft: 22,432 → 6,818 states (7.6× reduction, M3 Pro)
./spectre verify examples/raft-election-safety.spec --max-states 30000 --por

# Property-directed: best-first BFS toward invariant boundaries
# Finds all 29 violation traces in 204 states; random walk finds 0 in 150,000 steps
./spectre verify examples/queue.spec --property-guided
```

`--por` and `--property-guided` can be combined. Use `--por` for safety verification of large models; use `--property-guided` when you expect violations and want them found fast.

---

### Coverage-Guided Model-Based Testing

```bash
# 1. Generate a property-directed trace (steers toward invariant boundaries)
./spectre verify examples/bank-account-parameterized.spec \
    --emit-traces trace.itf.json --coverage-mode property

# 2. Generate a Rust driver
./spectre generate-driver --lang rust examples/bank-account-parameterized.spec \
    --output src/spec_driver.rs

# 3. Fill in the action implementations in spec_driver.rs, then replay:
./spectre check-refinement examples/bank-account-parameterized.spec \
    --traces trace.itf.json
```

**Coverage modes:**

| Mode | Strategy |
|------|---------|
| `action` | Prefers actions not yet covered in the current trace |
| `transition-pair` | Prefers (prev-action, next-action) pairs not yet seen |
| `boundary` | Steers toward the maximum boundary of integer variables |
| `rare-action` | Prefers globally least-executed actions |
| `property` | Steers toward states where the minimum integer variable is lowest (closest to invariant boundary) |

In evaluation on bank-account, `property` mode achieved **52% boundary-state coverage** vs 5–25% for other modes.

---

### Embedded Runtime Monitor

```bash
# 1. Generate the monitor
./spectre generate-monitor --lang rust myspec.spec -o src/monitor.rs

# 2. In your Rust code, call after every transition:
#    monitor.step("action_name", new_state);

# 3. At shutdown, check liveness:
#    monitor.unmet_liveness_properties()
```

The generated monitor checks all invariants in O(k) time (k = number of invariants) with no external dependencies.

---

## Documentation

| Document | Description |
|----------|-------------|
| [docs/spec.md](./docs/spec.md) | Full language specification and API reference |
| [docs/language_definition.md](./docs/language_definition.md) | Formal grammar (EBNF) and operational semantics |
| [docs/spectre_book.md](./docs/spectre_book.md) | The Spectre Language Book — tutorials and worked examples |
| [spectre_book/09-spec-mining-from-rust.md](./spectre_book/09-spec-mining-from-rust.md) | Spec mining from Rust and drift detection |
| [spectre_book/10-model-based-testing.md](./spectre_book/10-model-based-testing.md) | Model-based testing and ITF trace replay |
| [spectre_book/11-runtime-monitoring.md](./spectre_book/11-runtime-monitoring.md) | Embedded runtime monitoring |

---

## License

To be determined.
