# Chapter 1: Getting Started

This chapter covers installing Spectre and running your first specification.

---

## Prerequisites

| Dependency | Required for | Notes |
|------------|-------------|-------|
| **Go ≥ 1.24** | Core CLI — `verify`, `sync`, `simulate`, `generate-*` | Required |
| **Z3 ≥ 4.8** | `spectre sync` (SMT equivalence checking) | Optional — only needed for drift detection |
| **Rust stable** | `spectre mine --lang rust` (spec mining) | Optional — only needed for spec mining |
| **Git** | Cloning the repository | Required |

Z3 and Rust are only needed if you use `spectre sync` and `spectre mine --lang rust` respectively.
All other commands (`verify`, `simulate`, `generate-monitor`, `generate-driver`) work with Go alone.

---

## Installation — macOS

### Step 1 — Install Go

```bash
brew install go
go version
# Expected: go version go1.24.x darwin/arm64  (or amd64)
```

Or download the `.pkg` installer from [https://go.dev/dl/](https://go.dev/dl/).

### Step 2 — Install Z3

Required for `spectre sync`. Skip if you are not using drift detection.

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

### Step 4 — Build `spectre-mine-rs`

Required for `spectre mine --lang rust`. Skip if you are not mining specs from Rust source.

```bash
# Install Rust if needed
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"

# Build the Rust AST miner (inside the repo)
cargo build --release --manifest-path rust/spectre-mine-rs/Cargo.toml
cp rust/spectre-mine-rs/target/release/spectre-mine-rs .
```

The `spectre-mine-rs` binary must sit next to the `spectre` binary (or be on `PATH`).

### Step 5 — (Optional) Install globally

```bash
go install ./cmd/spectre
cp spectre-mine-rs "$(go env GOPATH)/bin/"

# Add Go's bin to PATH if not already there
export PATH="$PATH:$(go env GOPATH)/bin"
# Persist: add the export line to ~/.zshrc or ~/.bash_profile
```

### Step 6 — Verify the installation

```bash
./spectre verify examples/counter.spec
# Expected: Traversed N states. Found no violations.
```

---

## Installation — Linux

### Step 1 — Install Go

**Debian / Ubuntu:**

```bash
wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

> For ARM64 (Raspberry Pi, AWS Graviton): replace `amd64` with `arm64`.

**Fedora / RHEL / CentOS:**

```bash
sudo dnf install -y golang
go version
```

**Arch Linux:**

```bash
sudo pacman -S go
```

### Step 2 — Install Z3

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
go test ./...   # optional: run test suite
```

### Step 4 — Build `spectre-mine-rs`

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"
cargo build --release --manifest-path rust/spectre-mine-rs/Cargo.toml
cp rust/spectre-mine-rs/target/release/spectre-mine-rs .
```

---

## Installation — Windows

### Step 1 — Install Go

Download the `.msi` installer from [https://go.dev/dl/](https://go.dev/dl/) and run it. Or:

```powershell
winget install GoLang.Go
```

### Step 2 — Install Z3

Download the Windows binary from [https://github.com/Z3Prover/z3/releases](https://github.com/Z3Prover/z3/releases).
Extract the archive and add the `bin\` folder to your system `PATH`.

### Step 3 — Clone and build

```powershell
git clone https://github.com/BoundlessProducts/spectre.git
cd spectre
go build -o spectre.exe ./cmd/spectre
.\spectre.exe
```

### Step 4 — Build `spectre-mine-rs`

Install Rust from [https://rustup.rs](https://rustup.rs), then:

```powershell
cargo build --release --manifest-path rust\spectre-mine-rs\Cargo.toml
copy rust\spectre-mine-rs\target\release\spectre-mine-rs.exe .
```

---

## Docker (Recommended for Reviewers)

The Docker image includes all dependencies (Go, Z3, Rust, pre-built binaries).

```bash
docker build -t spectre-vmcai .
docker run --rm spectre-vmcai sh /artifact/reproduce.sh
```

To run interactively:

```bash
docker run --rm -it spectre-vmcai sh
```

---

## Your First Specification

Save the following as `counter.spec`:

```spectre
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

invariant nonNegative {
  counter >= 0
}
```

Run it:

```bash
./spectre parse counter.spec
./spectre typecheck counter.spec
./spectre verify counter.spec
```

### What each line means

- `var counter: int` — a state variable named `counter` of type integer
- `init { counter = 0 }` — the system starts with `counter = 0`
- `action increment { counter' = counter + 1 }` — one step: counter increases by 1. The prime (`'`) denotes the next-state value.
- `invariant nonNegative { counter >= 0 }` — a property checked on every reachable state

Since `increment` can only increase the counter and it starts at 0, the invariant holds for all reachable states.

---

## Running the Bundled Examples

All examples are in the `examples/` directory of the repository:

```bash
# Basic counter
./spectre verify examples/counter.spec

# Raft election safety (3-node, 22,432 states)
./spectre verify examples/raft-election-safety.spec --max-states 30000

# Spec mining from Rust
./spectre mine --lang rust examples/rust/bank_account.rs
```

---

## Next Steps

Continue to [Chapter 2: Language Overview](02-language-overview.md) to learn all the elements of the Spectre language.
