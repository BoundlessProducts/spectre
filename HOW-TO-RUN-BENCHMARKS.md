# How to Reproduce the Results

This document explains how to reproduce every result in the paper
*"Spectre: Closed-Loop Refinement, Repair, and Conformance for Rust"* (VMCAI 2027).

The benchmark script (`artifact/run-benchmarks.sh`) covers:

| Paper location | What is checked |
|----------------|----------------|
| **Table 2** (§6 RQ1) | State counts + violation/no-violation for all 9 verification benchmarks |
| **Table 3** (§6 RQ3) | Guard extraction count + init values from `bank_account.rs` |
| **Table 4** (§6 RQ4) | Incremental re-verification: pruned counts (deterministic) and total state count |
| **§5 coverage table** | `property` mode achieves the highest boundary-step coverage across all 4 modes |

**What the script does not automate:**  
WP repair timing (§6 RQ2) is exercised by Table 2 row 1 (`bank-account-violation`) — the script checks that CEGIS output is present but does not measure synthesis time across all five repair benchmarks.

---

## Prerequisites

Choose **one** path:

### Path A — Docker (recommended for reviewers)

- Docker 20.10 or later
- Internet access for the initial build (fetches Go + Rust base images and crate deps; ~800 MB, ~3–5 min; subsequent builds are cached)

### Path B — Local build

- Go 1.21 or later — <https://go.dev/dl/>
- Rust stable toolchain — `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`
- GNU Make
- `$GOPATH/bin` on your `$PATH`
- `python3` (for the coverage-mode section of the script)

Verify:

```bash
go version      # go1.21.x or later
cargo --version # cargo 1.70.x or later
python3 --version
```

---

## Path A — Docker

### Step 1 — Build the image

```bash
make docker-build
# or: docker build -t spectre-vmcai .
```

### Step 2 — Run all benchmarks

```bash
docker run --rm --entrypoint sh spectre-vmcai /artifact/run-benchmarks.sh
```

Expected final line: `Results: 30 passed, 0 failed`

Expected total runtime: ~40 s on an amd64 machine (Raft BFS dominates at ~8–10 s).

---

## Path B — Local build

### Step 1 — Build and install

```bash
make install
# Builds spectre (Go) and spectre-mine-rs (Rust) and installs both to $GOPATH/bin
```

### Step 2 — Run all benchmarks from the repository root

```bash
sh artifact/run-benchmarks.sh
```

The script detects `./spectre` if present (local build mode) and falls back to the
`spectre` binary on `$PATH` (Docker mode).  Run from the repository root so that
relative paths to `examples/` resolve correctly.

Expected final line: `Results: 30 passed, 0 failed`

---

## What the Output Tells You

### Deterministic values (must match exactly)

| Check | Expected value |
|-------|---------------|
| `concurrent-lock-corrected` states | 1928 |
| `message-queue-corrected` states | 22 |
| `inventory-corrected` states | 11 |
| Raft cold BFS states | 22432 |
| bank-account cold BFS states | 3745 |
| freeze: pruned states | 1882 |
| stepDown1\_sees2: pruned states | 1649 |
| vote2for1: pruned states | 6501 |
| stepDown1\_sees2/vote2for1: total after incremental | 22432 |
| Guard count from `bank_account.rs` | ≥ 7 |

### Non-deterministic values (informational only)

- **Wall times** — vary with hardware. Paper figures are from an Apple M3 Pro (18 GB, macOS 15.6, Go 1.24, `go install` default optimization). The ratios (33× cache restore, 7.7×/3.6×/3.0× incremental) compare same-hardware runs and are the meaningful metric.
- **Incremental "new BFS" counts** — vary by ±2 across runs due to Go map iteration order during BFS deduplication. The pruned count and total count are stable.

---

## Running Individual Commands (Table 4 details)

To reproduce the incremental re-verification workflow manually:

```bash
# Step 1 — cold BFS, build and save cache
./spectre verify examples/raft-election-safety.spec --max-states 30000 --use-cache

# Step 2 — cache restore (unchanged spec, ~33× faster)
./spectre verify examples/raft-election-safety.spec --max-states 30000 --use-cache

# Step 3 — incremental: re-verify only stepDown1_sees2 (~3.6× faster)
./spectre verify examples/raft-election-safety.spec --max-states 30000 \
    --use-cache --incremental --changed-action stepDown1_sees2

# Step 4 — incremental: re-verify only vote2for1 (~3.0× faster)
./spectre verify examples/raft-election-safety.spec --max-states 30000 \
    --use-cache --incremental --changed-action vote2for1
```

---

## Correspondence: Paper Claims → Benchmark Checks

| Paper claim | Script check |
|-------------|-------------|
| "Invariant violation found in one BFS step" (§6) | `bank-account-violation`: VIOLATION DETECTED present |
| "CEGIS synthesises `require aliceBalance - 50 >= 0`" (§6) | CEGIS Repair output present |
| "1928 states for concurrent-lock-corrected" (Table 2) | State count = 1928 |
| "22432 states for 3-node Raft" (Table 2, §6) | Raft state count = 22432 |
| "7 assert! guards extracted" (Table 3) | `grep -c require` ≥ 7 |
| "Constructor init extracted" (Table 3) | `balance = 0`, `frozen = false` in init |
| "3745 states, cache restore 11×" (Table 4) | bank-account: 3745 states + Restored message |
| "freeze: 1882 pruned, 0 new BFS, 7.7×" (Table 4) | `pruned 1882 states`, `0 new states` |
| "stepDown1_sees2: 1649 pruned, 3.6×" (Table 4) | `pruned 1649 states`, total = 22432 |
| "vote2for1: 6501 pruned, 3.0×" (Table 4) | `pruned 6501 states`, total = 22432 |
| "property mode: 52% boundary coverage" (§5) | BOUNDARY=11 out of 21 steps |
| "property mode dominates all other modes" (§5) | 11 > 5 > 2 > 1 boundary steps |

---

## Interactive Docker Commands

```bash
# Full CEGIS trace for bank-account-violation
docker run --rm spectre-vmcai verify /examples/bank-account-violation.spec

# Coverage-mode comparison (runs all four modes)
docker run --rm --entrypoint sh spectre-vmcai -c '
  for mode in action boundary rare-action property; do
    spectre verify /examples/bank-account-corrected.spec \
        --emit-traces /tmp/t.itf.json --coverage-mode $mode --max-depth 20
  done
'

# Generate a Rust test driver
docker run --rm spectre-vmcai generate-driver --lang rust \
    /examples/bank-account-parameterized.spec

# Generate an embedded runtime monitor
docker run --rm spectre-vmcai generate-monitor --lang rust \
    /examples/bank-account-parameterized.spec
```

---

## Troubleshooting

**`spectre: command not found`**  
Add `$GOPATH/bin` to your PATH: `export PATH="$(go env GOPATH)/bin:$PATH"`

**`make install` fails with "command not found: cargo"`**  
Install Rust: `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`, then open a new shell.

**`spectre mine` prints "using regex miner"**  
`spectre-mine-rs` was not installed. Re-run `make install` and verify `which spectre-mine-rs`.

**Docker build fails with "network timeout"**  
The build downloads Rust crates from crates.io. Retry once; if it persists, check that Docker has internet access.

**State counts differ for bounded specs (rows 1, 2, 3, 5, 8)**  
These hit a configurable `--max-states` limit. Counts match only if you use the same limit as the paper (defaults shown in Table 2).

**State counts differ for finite specs (rows 4, 6, 7)**  
These are exhaustive and hardware-independent. If they differ, please open an issue.
