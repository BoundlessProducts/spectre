# IMPORTANT: This is still under construction. It has been made public only to test installations etc. No PRs  WILL BE ENTERTAINED. 



# Spectre Language

A programmer-friendly specification language inspired by TLA+ and Quint, designed for Java and TypeScript developers.

## Features

- **Familiar Syntax**: Java/TypeScript-like syntax for easy adoption
- **Strong Typing**: Support for primitive and compound types
- **Modularity**: Modules for code organization and reuse
- **Constants**: Parameterized specifications with constants
- **State Management**: Clear state variable declarations and transitions
- **Verification**: Support for both classic constraints (invariants) and temporal properties
- **Fairness Conditions**: Weak and strong fairness for concurrent systems
- **Readable Error Messages**: Descriptions provide context in error traces
- **Pure Functions**: Reusable computational helpers without side effects
- **Non-deterministic Initialization**: `oneOf` operator for multiple starting states

## Quick Start

### Installation

#### macOS (Homebrew)

**Prerequisites**: Go 1.19 or later must be installed (Homebrew will build Spectre from source).

If Go is not installed, install it first:
```bash
brew install go
```

Since the Homebrew formula is in the main repository (not a separate `homebrew-spectre` repo), use the full GitHub URL:

```bash
# Add the tap (using full GitHub URL since formula is in main repo)
brew tap akkeshavan/spectre https://github.com/akkeshavan/spectre.git

# Install (this will build from source, so Go is required)
brew install spectre
```

**Note**: Homebrew will automatically install Go as a dependency if it's not already installed, but it's recommended to install Go first to ensure you have the correct version.

**Note**: If you later create a separate `homebrew-spectre` repository, you can simplify to:
```bash
brew tap akkeshavan/spectre
brew install spectre
```

**Alternative: Build from source**
```bash
git clone https://github.com/akkeshavan/spectre.git
cd spectre
go build -o spectre ./cmd/spectre
sudo mv spectre /usr/local/bin/
```

#### Linux

**Using the install script (builds from source):**

**Prerequisites**: Go 1.19 or later must be installed (the script builds Spectre from source).

The install script downloads the source code and builds Spectre locally:

```bash
curl -fsSL https://raw.githubusercontent.com/akkeshavan/spectre/main/scripts/install.sh | bash
```

The script will:
- Check for Go 1.19+ installation (with helpful error messages if not found)
- Check for Git installation
- Clone the repository from GitHub
- Build Spectre from source using Go
- Install to `/usr/local/bin` (or custom `INSTALL_DIR`)

**If Go is not installed**, the script will provide installation instructions. You can install Go with:

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install golang-go

# Fedora/RHEL/CentOS
sudo dnf install golang
```

Or download from: https://golang.org/dl/

**Note**: The script automatically detects if it's running as root (e.g., in Docker containers) and skips `sudo` commands. On regular Linux systems, it will use `sudo` for installation.

**Or build from source:**
```bash
git clone https://github.com/akkeshavan/spectre.git
cd spectre
go build -o spectre ./cmd/spectre
sudo mv spectre /usr/local/bin/
```

**Using package managers:**

Debian/Ubuntu (.deb):
```bash
# Download from releases page, then:
sudo dpkg -i spectre_*.deb
```

RHEL/CentOS (.rpm):
```bash
# Download from releases page, then:
sudo rpm -i spectre_*.rpm
```

Snap:
```bash
sudo snap install spectre
```

#### Windows

**Using PowerShell script:**
```powershell
powershell -ExecutionPolicy Bypass -File scripts/install.ps1
```

**Or build from source:**
```powershell
git clone https://github.com/akkeshavan/spectre.git
cd spectre
go build -o spectre.exe ./cmd/spectre
# Add to PATH manually or use install.ps1
```

**Using Chocolatey (coming soon):**
```powershell
choco install spectre
```

#### Verify Installation

```bash
spectre --version
```

Should output: `spectre version 0.1.0`

### Usage

The Spectre CLI provides three main commands:

**Parse** - Check syntax:
```bash
spectre parse examples/counter.spec
```

**Typecheck** - Verify types:
```bash
spectre typecheck examples/counter.spec
```

**Verify** - Check invariants and temporal properties:
```bash
spectre verify examples/counter.spec
```

You can also process multiple files:
```bash
spectre parse examples/*.spec
```

Or process all files in a directory:
```bash
spectre parse examples/
```

### Basic Example

```spectre
description "Tracks a numeric counter value"
var counter: int

description "System starts with counter initialized to zero"
init {
  counter = 0
}

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}

description "Verifies that counter eventually reaches value 10"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}
```

## Documentation

- **[SPEC.md](./SPEC.md)** - Complete language specification
- **[USAGE.md](./USAGE.md)** - CLI usage guide and examples
- **[INSTALL.md](./INSTALL.md)** - Detailed installation instructions
- **[PACKAGING.md](./PACKAGING.md)** - Packaging and distribution guide
- **[README_DEV.md](./README_DEV.md)** - Development setup and workflow

## Examples

The `examples/` directory contains several example specifications:

- **counter.spec**: Simple counter with basic transitions and properties
- **mutex.spec**: Mutual exclusion lock with fairness properties
- **user-management.spec**: User management system with sets and complex types
- **bank-account.spec**: Bank account system with transactions
- **message-queue.spec**: Priority queue with processing guarantees
- **pure-functions.spec**: Demonstrates pure functions for computations
- **oneof-example.spec**: Multiple initial states using `oneOf` operator
- **error-trace-example.spec**: Shows how descriptions improve error messages
- **modules-example.spec**: Module organization, imports, and extension
- **constants-example.spec**: Parameterized specifications with constants
- **fairness-example.spec**: Weak and strong fairness conditions

## Language Highlights

### Types

- Primitives: `int`, `bool`, `str`, `float`
- Records: `type User = { id: int, name: str }`
- Sets: `Set<User>`
- Maps: `Map<int, User>`
- Lists: `List<User>`
- Enums: `enum Status { Pending, Completed }`

### State and Transitions

```spectre
description "Tracks a numeric counter value"
var counter: int

description "System starts with counter initialized to zero"
init {
  counter = 0
}

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}
```

### Pure Functions

```spectre
description "Adds two integers together"
fun add(x: int, y: int): int {
  return x + y
}
```

### Constraints

**Classic (Invariants):**
```spectre
description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}
```

**Temporal:**
```spectre
description "Verifies that counter eventually reaches value 10"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}
```

### Multiple Initial States

```spectre
description "System can start from multiple configurations"
init oneOf {
  { counter = 0, mode = "start" },
  { counter = 10, mode = "resume" },
  { counter = 20, mode = "restart" }
}
```

### Modules

```spectre
module Counter {
  var counter: int
  action increment { counter' = counter + 1 }
}

import Counter
```

### Constants

```spectre
const MAX_USERS: int = 100
const TIMEOUT: int = 30

action addUser {
  require users.size() < MAX_USERS
  // ...
}
```

### Fairness Conditions

```spectre
description "Weak fairness ensures action executes when continuously enabled"
temporal weakFairness {
  WF(increment)
}

description "Strong fairness ensures action executes if enabled infinitely often"
temporal strongFairness {
  SF(decrement)
}
```

## Verification

The Spectre verifier checks:
- Type safety
- Invariants (classic constraints)
- Temporal properties (liveness)
- Preconditions and postconditions
- All initial states when using `oneOf`

Error messages include full execution traces with descriptions, making debugging easier.

### Example Output

**Successful verification:**
```
✓ Verification passed for examples/counter.spec
  Explored 10 states
```

**Violation found:**
```
Verification failed: 1 violation(s) found

Violation 1:
  Invariant 'nonNegative' violated: counter < 0
  Path:
    1. increment
    2. decrement
```

**Type error:**
```
Type errors in examples/counter.spec:
  15:8: Type error: cannot assign int to string
```

## Contributing

We welcome contributions! Please see [README_DEV.md](./README_DEV.md) for development setup and workflow.

### Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/akkeshavan/spectre.git
   cd spectre
   ```

2. Build from source:
   ```bash
   go build -o spectre ./cmd/spectre
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

4. Test with examples:
   ```bash
   ./spectre parse examples/counter.spec
   ```

See [README_DEV.md](./README_DEV.md) for detailed development instructions.

## License

(To be determined)

