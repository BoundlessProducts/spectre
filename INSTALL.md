# Installation

Two paths: Docker (recommended for artifact evaluation) or local build.

---

## Docker — recommended

Requires Docker 20.10 or later and an internet connection for the first build.

```bash
# Build the image (~3-5 min first time; cached thereafter)
make docker-build

# Verify the tool works
docker run --rm spectre-vmcai --help

# Run the full benchmark suite (Tables 2, 3, 4 and §5 coverage comparison)
make docker-test
```

Expected final output: `Results: 30 passed, 0 failed`

---

## Local build

Requires Go 1.21+ and Rust stable. Python 3 is required for the coverage-mode
benchmark section.

```bash
# Install Go — https://go.dev/dl/
# Install Rust — curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
# Install Python 3 — standard on macOS/Linux

# Build and install both binaries to $GOPATH/bin
make install

# Add Go's bin directory to PATH if not already present
export PATH="$(go env GOPATH)/bin:$PATH"

# Verify
spectre --help
spectre-mine-rs --help   # optional: only needed for spec mining

# Run the test suite
go test ./...

# Run the artifact benchmark suite from the repository root
sh artifact/run-benchmarks.sh
```

Expected final output: `Results: 30 passed, 0 failed`

---

## Quick smoke test (~10 seconds)

```bash
sh artifact/smoke-test.sh
```

Runs three deterministic checks without the slow Raft BFS: concurrent-lock state
count, spec mining guard extraction, and bank-account invariant violation detection.

---

## Supported platforms

| Platform | Status |
|----------|--------|
| macOS arm64 (Apple Silicon) | Primary development platform |
| macOS amd64 (Intel) | Supported via `go build` |
| Linux amd64 | Supported; Docker image uses `debian:bookworm-slim` |
| Linux arm64 | Supported via `go build` |
| Windows | `go build` supported; Docker recommended |

All benchmark numbers in the paper were measured on **Apple M3 Pro, 18 GB RAM,
macOS 15.6, Go 1.24** (`go install`, default optimization). State counts are
hardware-independent; wall times are not.
