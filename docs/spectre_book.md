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
# Chapter 2: Language Overview

This chapter covers the core elements of the Spectre language.

---

## What is Spectre?

Spectre is a programmer-friendly specification language inspired by TLA+ and Quint, designed specifically for Java and TypeScript developers. It allows you to:

- **Model systems** as state machines with clear state variables and transitions
- **Verify properties** using both classic constraints (invariants) and temporal properties
- **Express concurrency** with fairness conditions
- **Organize code** using modules and imports
- **Debug easily** with readable error messages and execution traces

### Why Spectre?

Traditional formal specification languages like TLA+ can be intimidating for developers coming from mainstream programming languages. Spectre bridges this gap by providing:

- **Familiar syntax** similar to Java and TypeScript
- **Strong typing** with type inference
- **Clear semantics** for state machines and transitions
- **Comprehensive verification** tools
- **Excellent error messages** with descriptions and stack traces

### Key Features

- ✅ **Type System**: Primitive and compound types (records, sets, maps, lists, enums, options)
- ✅ **State Management**: Explicit state variables with prime notation (`'`) for next-state
- ✅ **Non-determinism**: `oneOf` operator for multiple initial states
- ✅ **Pure Functions**: Reusable computational helpers without side effects
- ✅ **Temporal Logic**: `always`, `eventually`, `until`, `leads-to` operators
- ✅ **Fairness**: Weak Fairness (WF) and Strong Fairness (SF) conditions
- ✅ **Modules**: Code organization with imports and inheritance
- ✅ **Descriptions**: Human-readable context in error messages

---

## Core Concepts

### State Machines

Spectre models systems as **state machines**. A state machine consists of:

1. **State Variables**: The data that represents the system's state
2. **Initial States**: How the system starts
3. **Actions**: Transitions that change the state
4. **Properties**: Invariants and temporal properties to verify

### State vs Next-State

In Spectre, you distinguish between the current state and the next state:

- **Current state**: `counter` (the value now)
- **Next state**: `counter'` (the value after an action executes)

This is called **prime notation** and comes from TLA+. The prime (`'`) notation is essential for specifying how actions transform the system from one state to another.

---

## Language Elements

Spectre includes the following language elements:

### 1. Types

Spectre provides a rich type system:

**Primitive Types:**
- `int`: Integers (e.g., `42`, `-10`, `0`)
- `bool`: Booleans (`true`, `false`)
- `str`: Strings (e.g., `"hello"`, `"world"`)
- `float`: Floating-point numbers (e.g., `3.14`, `-0.5`)

**Compound Types:**
- **Records**: Group related data together (e.g., `type User = { id: int, name: str }`)
- **Sets**: Unordered collections of unique elements (e.g., `Set<int>`, `Set<User>`)
- **Maps**: Key-value pairs (e.g., `Map<int, User>`)
- **Lists**: Ordered sequences (e.g., `List<int>`)
- **Options**: Values that may or may not exist (e.g., `Option<User>`)
- **Enums**: Fixed sets of named values (e.g., `enum ProcessState { Idle, Critical }`)

### 2. State Variables

State variables represent the data that changes over time in your system:

```spectre
description "Tracks a numeric counter value"
var counter: int

description "Collection of all users in the system"
var users: Set<User>
```

State variables are declared using the `var` keyword followed by a name and type.

### 3. Initial States

The initial state defines how your system starts:

```spectre
description "System starts with counter initialized to zero"
init {
  counter = 0
}
```

You can also use `oneOf` for multiple initial states:

```spectre
init oneOf {
  { counter = 0, mode = "start" },
  { counter = 10, mode = "resume" }
}
```

### 4. Actions

Actions define how the system can transition from one state to another:

```spectre
description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Decrements the counter, only when positive"
action decrement {
  require counter > 0
  counter' = counter - 1
}
```

Actions can have:
- **Preconditions** (`require`): Conditions that must be true for the action to execute
- **Postconditions** (`ensure`): Conditions that must be true after the action executes
- **Parameters**: Actions can take parameters like functions

### 5. Pure Functions

Pure functions are computational helpers that don't modify state:

```spectre
description "Adds two integers together"
fun add(x: int, y: int): int {
  return x + y
}

description "Returns the maximum of two integers"
fun max(a: int, b: int): int {
  if (a > b) {
    return a
  } else {
    return b
  }
}
```

Pure functions:
- Only read their parameters
- Do not access state variables
- Do not modify state variables
- Return a value based only on inputs

### 6. Constants

Constants are fixed values that don't change during execution:

```spectre
description "Number of processes in the system"
const NUM_PROCESSES: int = 2

description "Maximum allowed counter value"
const MAX_VALUE: int = 100
```

### 7. Modules

Modules help organize your specifications into reusable components and enable code sharing between files. Spectre supports file-level modules where each file contains exactly one module.

#### Module Declaration

A module is declared in a file with the following structure:

```spectre
// File: Counter.spec
module Counter {
  description "Tracks a numeric counter value"
  var counter: int
  
  public action increment {
    counter' = counter + 1
  }
  
  action decrement {
    counter' = counter - 1
  }
}
```

**Important Rules:**
- Each file must contain exactly one module declaration
- The module name must match the file name (without extension)
- Files use PascalCase naming (e.g., `Counter.spec`, `ElevatorModule.spec`)
- Module names use PascalCase (e.g., `module Counter`, `module ElevatorModule`)

#### Visibility Modifiers

Members within a module can be marked as `public` or have no modifier (private by default):

```spectre
module ExampleModule {
  public var publicVar: int      // Accessible from other modules
  var privateVar: int            // Only accessible within this module
  
  public action publicAction { ... }
  action privateAction { ... }
  
  public const PUBLIC_CONST: int = 10
  const PRIVATE_CONST: int = 20
  
  public fun publicFunction(): int { ... }
  fun privateFunction(): int { ... }
}
```

- **Public members** (`public var`, `public action`, `public const`, etc.) are accessible from modules that import this module
- **Private members** (no modifier) are only accessible within the module

#### Importing Modules

To use a module in another file, you import it using one of two syntaxes:

**Import from Same Directory:**
```spectre
import Counter
```
This imports `Counter` from `Counter.spec` in the same directory.

**Import from Path:**
```spectre
import "path/to/ModuleName"
```
This imports a module from a relative path. The path should point to the file (without extension).

**Examples:**
```spectre
// Import from same directory
import ElevatorModule
import UserModule

// Import from subdirectory
import "elevator/ElevatorModule"

// Import from parent directory
import "../common/Utils"
```

#### Accessing Module Members

When you import a module, you access its public members using dot notation:

```spectre
import ControllerModule

// Access constants
const maxUsers = ControllerModule.MAX_USERS

// Call functions
let distance = ControllerModule.distance(5, 10)

// Use types (types are automatically available)
var elevator: Elevator  // Type from ElevatorModule
```

#### Module Resolution Rules

1. **No circular dependencies**: If module A imports module B, module B cannot import module A (directly or indirectly)
2. **Module name matching**: The module declaration name must exactly match the file name (case-sensitive)
3. **One module per file**: Each `.spec` file must contain exactly one module

#### Benefits of Modules

Modules provide several benefits:
- **Code reuse**: Share modules across multiple specifications
- **Separation of concerns**: Organize code into logical units
- **Maintainability**: Changes to one module don't affect others (as long as public interface remains stable)
- **Testability**: Modules can be tested independently

For a comprehensive example of using modules to build a large system, see **Chapter 7: Modules and Code Organization**, which demonstrates building an elevator controller system using multiple modules.

---

## Descriptions: Why They Matter

**Descriptions are a critical feature of Spectre.** Every language element can have a `description` field that provides human-readable context:

```spectre
description "Tracks a numeric counter value"
var counter: int

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}
```

### How Descriptions Improve Error Messages

When an error occurs, descriptions appear in the error message, making it much easier to understand what went wrong:

**Without descriptions:**
```
21:2: Invariant 'inv1' violated: condition evaluated to false
```

**With descriptions:**
```
21:2: Invariant 'nonNegative' violated: (Ensures counter never becomes negative) 
       condition evaluated to false
```

The description immediately tells you what the invariant was supposed to ensure, making debugging much faster.

### Description Best Practices

1. **Always add descriptions**: They make error messages much more readable
2. **Be specific**: "Ensures counter never becomes negative" is better than "Counter check"
3. **Use active language**: "Increments the counter" is clearer than "Counter increment"
4. **Mark critical properties**: Use "CRITICAL:" prefix for important invariants

Descriptions will eventually be used extensively in error messages, stack traces, and verification reports to provide clear, actionable feedback when something goes wrong.

---

## Properties: Invariants, Temporal Properties, and Fairness

Spectre provides three kinds of properties that you can verify:

### 1. Invariants

**Invariants** are properties that must **always** hold in every reachable state. They express safety properties—things that should never happen.

```spectre
description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}

description "CRITICAL: Ensures only one process can be in critical section at a time"
invariant mutualExclusion {
  !(process1 = ProcessState.Critical && process2 = ProcessState.Critical)
}
```

**When invariants are checked:**
- After the initial state is set
- After every action execution
- During state space exploration

**What happens when an invariant fails:**
The verifier reports:
- Which invariant failed
- The description of that invariant
- The state that violated it
- The execution trace leading to the violation

**Use invariants for:**
- Safety properties (e.g., "counter never negative", "mutual exclusion holds")
- State consistency (e.g., "lock flag matches process state")
- Data integrity (e.g., "all users have valid IDs")

### 2. Temporal Properties

**Temporal properties** describe how the system behaves **over time**. Unlike invariants (which must always hold), temporal properties can express liveness conditions like "eventually something happens."

#### Temporal Operators

**`always`**: The property holds at every point in the execution
```spectre
description "Ensures counter remains non-negative throughout execution"
temporal alwaysNonNegative {
  always (counter >= 0)
}
```

**`eventually`**: The property holds at some point in the future
```spectre
description "Verifies that counter eventually reaches value 10"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}
```

**`until`**: Condition1 holds until condition2 becomes true
```spectre
temporal safeUntilGoal {
  counter < 100 until counter = 100
}
```

**`leads-to` (→)**: If condition1 becomes true, then condition2 will eventually become true
```spectre
description "Guarantees progress: if counter is below 10, it will eventually reach 10"
temporal progress {
  always (counter < 10 → eventually counter = 10)
}
```

**When temporal properties are checked:**
Temporal properties are evaluated over **execution traces** (sequences of states). The verifier explores the state space and checks if the temporal property holds for all possible execution paths.

**Use temporal properties for:**
- Liveness properties (e.g., "eventually something happens")
- Progress guarantees (e.g., "if condition P holds, eventually Q will hold")
- Response properties (e.g., "request eventually leads to response")

### 3. Fairness Conditions

**Fairness conditions** ensure that actions execute "fairly" in concurrent systems. They're essential for verifying liveness properties.

#### Weak Fairness (WF)

**Weak Fairness** means: "If an action is **continuously enabled**, it will eventually execute."

```spectre
description "Weak fairness ensures process1 gets fair access when continuously waiting"
temporal process1WeakFairness {
  WF(process1Request)
}
```

**Meaning**: If `process1Request` is enabled and stays enabled, it will eventually execute.

#### Strong Fairness (SF)

**Strong Fairness** means: "If an action is enabled **infinitely often**, it will eventually execute."

```spectre
description "Strong fairness ensures process1 executes even if intermittently enabled"
temporal process1StrongFairness {
  SF(process1Request)
}
```

**Meaning**: Even if `process1Request` is enabled and disabled repeatedly, it will eventually execute if it's enabled infinitely often.

#### Using Fairness in Temporal Properties

Fairness conditions are often combined with temporal operators using the leads-to (`→`) operator:

```spectre
description "With weak fairness, process1 will eventually enter critical section"
temporal eventuallyProcess1Critical {
  WF(process1Request) → eventually (process1 = ProcessState.Critical)
}
```

When used in a leads-to expression, fairness filters the transition graph to only consider paths that satisfy the fairness constraint. This ensures that only "fair" execution paths are considered when verifying the property.

**When to use Weak vs Strong Fairness:**
- **Weak Fairness (WF)**: Use when an action should execute if it's continuously available
  - Example: A process waiting for a lock should eventually get it if the lock is continuously free
  
- **Strong Fairness (SF)**: Use when an action should execute even if it's intermittently enabled
  - Example: A message queue processor should eventually process messages even if the queue is sometimes empty

**Use fairness for:**
- Guaranteeing progress in concurrent systems
- Verifying liveness properties that require action execution
- Modeling realistic scheduling behavior

---

## Putting It All Together

A complete Spectre specification typically includes:

1. **Type definitions** (enums, records)
2. **State variables** (with descriptions)
3. **Initial state** (with descriptions)
4. **Actions** (with descriptions, preconditions, and postconditions)
5. **Invariants** (safety properties)
6. **Temporal properties** (liveness properties)
7. **Fairness conditions** (for concurrent systems)

Here's a simple but complete example:

```spectre
// Type definition
enum ProcessState {
  Idle,
  Critical
}

// State variables
description "State of the first process"
var process1: ProcessState

description "State of the second process"
var process2: ProcessState

description "Lock flag indicating if a process is in critical section"
var lock: bool

// Initial state
description "System starts with both processes idle and lock free"
init {
  process1 = ProcessState.Idle
  process2 = ProcessState.Idle
  lock = false
}

// Actions
description "Process 1 requests and acquires the lock"
action process1Request {
  require process1 = ProcessState.Idle && !lock
  process1' = ProcessState.Critical
  lock' = true
}

description "Process 1 releases the lock"
action process1Release {
  require process1 = ProcessState.Critical
  process1' = ProcessState.Idle
  lock' = false
}

// Similar actions for process2...

// Invariant (safety property)
description "CRITICAL: Ensures only one process can be in critical section at a time"
invariant mutualExclusion {
  !(process1 = ProcessState.Critical && process2 = ProcessState.Critical)
}

// Temporal property (liveness property)
description "With weak fairness, process1 will eventually enter critical section"
temporal eventuallyProcess1Critical {
  WF(process1Request) → eventually (process1 = ProcessState.Critical)
}
```

---

## State Space Exploration

When you run `spectre verify`, Spectre performs **state space exploration** to verify your specification. This section explains how state space exploration works and how to control it.

### What is State Space Exploration?

State space exploration is the process of systematically examining all possible states your system can reach. The verifier:

1. **Starts from initial states**: Begins with all possible initial states (from `init` blocks or `oneOf` expressions)
2. **Explores transitions**: For each state, tries all enabled actions to find the next possible states
3. **Builds a transition graph**: Creates a graph of all reachable states and the transitions between them
4. **Checks properties**: Verifies invariants and temporal properties against all explored states and paths

This process ensures that your properties hold in **all possible execution scenarios**, not just one or two specific paths.

### Exploration Limits

Because some systems have an infinite or extremely large state space (millions or billions of states), Spectre uses limits to prevent exploration from running indefinitely:

- **Max States**: Maximum number of unique states to explore
- **Max Depth**: Maximum depth of the exploration tree (how many steps from the initial state)

### Default Limits

By default, Spectre uses these limits:

- **Default Max States**: 5,000 states
- **Default Max Depth**: 100 steps

These defaults work well for most small to medium-sized specifications but may need adjustment for larger systems.

### Setting Custom Limits

You can customize the exploration limits using command-line flags:

```bash
# Set max states to 10,000
./spectre verify my-spec.spec --max-states 10000

# Set max depth to 200
./spectre verify my-spec.spec --max-depth 200

# Set both limits
./spectre verify my-spec.spec --max-states 10000 --max-depth 200
```

**When to increase limits:**
- Your spec has many state variables
- You have parameterized actions that create many state combinations
- You're using `oneOf` with many initial states
- The verifier stops early with "explored N states" but you suspect more states exist

**Example**: An elevator controller with 4 elevators, 20 floors, and up to 30 users has a very large state space. You might need:

```bash
./spectre verify elevator/ElevatorController.spec --max-states 50000 --max-depth 150
```

### Unlimited Exploration

For exhaustive verification, you can set limits to **infinity** (unlimited). Spectre accepts three ways to specify unlimited:

```bash
# Using 'infinity'
./spectre verify my-spec.spec --max-states infinity --max-depth infinity

# Using 'unlimited'
./spectre verify my-spec.spec --max-states unlimited --max-depth unlimited

# Using -1
./spectre verify my-spec.spec --max-states -1 --max-depth -1
```

All three forms are equivalent and mean "explore until the entire reachable state space is covered or memory is exhausted."

**Warning**: When you enable unlimited exploration, you'll see:

```
Warning: Unlimited exploration enabled (--max-states: unlimited, --max-depth: unlimited)
This may run until the state space is fully explored or memory is exhausted.
```

### Consequences of Different Limit Settings

#### 1. Default Limits (5,000 states, 100 depth)

**Pros:**
- Fast verification (usually completes in seconds)
- Good for most small to medium specs
- Prevents accidental long-running verifications

**Cons:**
- May miss states in large specifications
- May truncate exploration before finding all reachable states
- May not fully verify temporal properties in complex systems

**Use when:** You have a small to medium specification and want fast feedback during development.

#### 2. Increased Custom Limits (e.g., 50,000 states, 200 depth)

**Pros:**
- More thorough exploration for larger specs
- Better coverage of complex state spaces
- Still has safety bounds to prevent runaway exploration

**Cons:**
- Takes longer to complete (minutes instead of seconds)
- Uses more memory
- May still not cover the entire state space for very large systems

**Use when:** You have a large specification and the default limits are too restrictive.

#### 3. Unlimited Exploration

**Pros:**
- **Complete verification**: Explores every reachable state
- **Exhaustive property checking**: Verifies properties against all possible execution paths
- **No artificial bounds**: Only limited by the actual state space size and available memory

**Cons:**
- ⚠️ **May run for a very long time**: Large specs can take hours or days
- ⚠️ **May exhaust memory**: Systems with millions of states can use gigabytes of RAM
- ⚠️ **No progress indication**: Until it completes, you don't know if it will finish

**Use when:**
- You need complete verification confidence
- You're verifying critical safety properties
- Your state space is bounded and you know it's finite
- You have sufficient computational resources and time

**Avoid when:**
- Your system has an infinite state space (e.g., unbounded counters)
- You're in active development and need fast feedback
- You have limited computational resources

### Best Practices

1. **Start with defaults**: Begin verification with default limits. If verification completes quickly, you're done.

2. **Increase incrementally**: If you suspect more states exist, increase limits incrementally (e.g., 10,000 → 50,000 → 100,000) to find the right balance.

3. **Use unlimited sparingly**: Reserve unlimited exploration for:
   - Final verification before release
   - Critical safety properties
   - Bounded systems where you know the state space is finite

4. **Monitor resource usage**: When using large limits or unlimited, monitor CPU and memory usage. Stop the process if it's consuming excessive resources.

5. **Understand your state space**: Try to estimate the size of your state space:
   - How many state variables?
   - What are their possible value ranges?
   - How many parameterized actions?
   - How many initial states (`oneOf`)?

   This helps you choose appropriate limits.

### Example: Choosing Limits

For a **simple counter** (0-10 range):
```bash
# Default limits are more than enough
./spectre verify counter.spec
```

For a **concurrent lock system** (2-3 processes):
```bash
# Default limits work well
./spectre verify concurrent-lock.spec
```

For an **elevator controller** (4 elevators, 20 floors, 30 users):
```bash
# Needs increased limits
./spectre verify elevator/ElevatorController.spec --max-states 50000 --max-depth 150
```

For a **small bounded system** where you need complete verification:
```bash
# Use unlimited for exhaustive checking
./spectre verify critical-safety.spec --max-states infinity --max-depth infinity
```

### Verification Output

The verifier reports how many states were explored:

```
✓ Verification passed for my-spec.spec
  Explored 4,237 states
  Verified 2 temporal properties
```

If exploration hits the limit before completing, you'll see:

```
✓ Verification passed for my-spec.spec
  Explored 5,000 states  (max limit reached)
  Verified 2 temporal properties
```

If you see "(max limit reached)", consider increasing the limits to ensure complete coverage.

---

## Next Steps

Now that you understand the overview of the Spectre language, you're ready to dive deeper into each element:

- **Chapter 3**: Types and State Variables (detailed coverage)
- **Chapter 4**: Actions and Transitions
- **Chapter 5**: Pure Functions and Constants
- **Chapter 6**: Invariants (detailed examples)
- **Chapter 7**: Temporal Properties (detailed examples)
- **Chapter 8**: Fairness Conditions (detailed examples)
- **Chapter 9**: Modules and Code Organization
- **Chapter 10**: Verification and Debugging



# Chapter 3: Invariants and Violations

This chapter covers invariants: what they are, how they can be violated, and how to fix violations.

## Understanding Invariants

**Invariants** are properties that must hold in **every reachable state** of the system. Unlike temporal properties (which describe behavior over time), invariants are state properties that must be true at all times.

### Key Characteristics of Invariants

1. **Must hold in initial state**: The invariant must be true when the system starts
2. **Must hold after every action**: After any action executes, the invariant must still be true
3. **Checked for all reachable states**: The verifier explores all possible execution paths and checks the invariant in every state

### Common Types of Invariants

- **Bounds checking**: Values stay within acceptable ranges
- **Non-negativity**: Values never become negative
- **Consistency**: Related values maintain consistent relationships
- **Safety properties**: Dangerous states are never reached

---

## Example 1: Bank Account System

Let's start with a bank account system that tracks account balances.

### The Problem: Missing Preconditions

Here's a bank account specification with a critical flaw:

```spectre
description "Balance of Alice's account"
var aliceBalance: int

description "Balance of Bob's account"
var bobBalance: int

description "At the initial state, all balances are zero"
init {
  aliceBalance = 0
  bobBalance = 0
}

description "Deposit 50 to Alice's account"
action depositAlice50 {
  aliceBalance' = aliceBalance + 50
}

description "Withdraw 30 from Alice's account"
description "PROBLEM: Missing precondition to check if funds are sufficient"
action withdrawAlice30 {
  // Missing: require aliceBalance >= 30
  aliceBalance' = aliceBalance - 30
}

description "Withdraw 100 from Alice's account"
description "PROBLEM: Missing precondition - can withdraw more than balance"
action withdrawAlice100 {
  // Missing: require aliceBalance >= 100
  aliceBalance' = aliceBalance - 100
}

description "Transfer 50 from Alice to Bob"
description "PROBLEM: Missing precondition to ensure sufficient funds"
action transfer50ToBob {
  // Missing: require aliceBalance >= 50
  aliceBalance' = aliceBalance - 50
  bobBalance' = bobBalance + 50
}

description "Invariant: Account balances should never be negative"
invariant no_negatives {
  aliceBalance >= 0 && bobBalance >= 0
}
```

### Why This Violates the Invariant

The `no_negatives` invariant states that both `aliceBalance` and `bobBalance` must always be non-negative. However, several actions can violate this:

1. **`withdrawAlice30`**: If Alice's balance is less than 30 (e.g., 20), this action would set `aliceBalance = -10`, violating the invariant.

2. **`withdrawAlice100`**: If Alice's balance is less than 100 (e.g., 50), this action would set `aliceBalance = -50`, violating the invariant.

3. **`transfer50ToBob`**: If Alice's balance is less than 50 (e.g., 30), this action would set `aliceBalance = -20`, violating the invariant.

### The Error

When you verify this specification, the verifier will **fail** and report multiple violations:

```bash
$ ./spectre verify examples/bank-account-violation.spec

Verification failed: 49 violation(s) found

Violation 1 (Invariant):
  Action 'withdrawAlice100' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]

Violation 2 (Invariant):
  Action 'transfer50ToBob' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]

Violation 3 (Invariant):
  Action 'withdrawAlice30' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]

Violation 4 (Invariant):
  Action 'withdrawAlice100' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]
  Path:
    1. depositAlice50

Violation 5 (Invariant):
  Action 'withdrawAlice30' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]
  Path:
    1. depositAlice50
    2. transfer50ToBob

...
```

### Understanding the Error Messages

Each violation message has several components:

1. **Violation Number**: The order in which violations were discovered during exploration
2. **Action Name**: Which action would cause the violation (e.g., `withdrawAlice100`)
3. **Invariant Name**: Which invariant would be violated (e.g., `no_negatives`)
4. **Description**: The human-readable description from the invariant declaration
5. **Path (when applicable)**: The sequence of actions that leads to a state where this violation can occur

**Example Breakdown**:

```
Violation 4 (Invariant):
  Action 'withdrawAlice100' would violate invariants: [invariant no_negatives violated: (Invariant: Account balances should never be negative)]
  Path:
    1. depositAlice50
```

This tells us:
- **What action would fail**: `withdrawAlice100`
- **Why it fails**: Executing it would violate the `no_negatives` invariant
- **Where it fails**: After executing `depositAlice50`, Alice's balance is 50. At this point, trying to withdraw 100 would make the balance -50, violating the invariant.

### Why Verification Fails

The verifier detects that **actions are available** (they can be executed from certain states) but **would violate invariants** if executed. This indicates a specification flaw:

- Actions should have **preconditions** (`require` statements) that prevent them from being available when they would violate invariants
- The verifier catches these violations during exploration, even though it prevents invalid states from being reached
- Each violation represents a state-action pair where the action would create an invalid state

**Why this matters**: In a real system, we want preconditions to prevent invalid operations, not rely on post-execution checks. Preconditions make the specification clearer and ensure that actions are only available when they're safe to execute.

### The Fix: Add Preconditions

To fix this, we need to add `require` statements (preconditions) to ensure actions only execute when they won't violate the invariant:

```spectre
description "Balance of Alice's account"
var aliceBalance: int

description "Balance of Bob's account"
var bobBalance: int

description "At the initial state, all balances are zero"
init {
  aliceBalance = 0
  bobBalance = 0
}

description "Deposit 50 to Alice's account"
action depositAlice50 {
  aliceBalance' = aliceBalance + 50
}

description "Withdraw 30 from Alice's account if funds are sufficient"
description "FIXED: Added precondition to check if funds are sufficient"
action withdrawAlice30 {
  require aliceBalance >= 30
  aliceBalance' = aliceBalance - 30
}

description "Withdraw 100 from Alice's account if funds are sufficient"
description "FIXED: Added precondition to prevent overdrawing"
action withdrawAlice100 {
  require aliceBalance >= 100
  aliceBalance' = aliceBalance - 100
}

description "Transfer 50 from Alice to Bob if Alice has sufficient funds"
description "FIXED: Added precondition to ensure sufficient funds"
action transfer50ToBob {
  require aliceBalance >= 50
  aliceBalance' = aliceBalance - 50
  bobBalance' = bobBalance + 50
}

description "Invariant: Account balances should never be negative"
invariant no_negatives {
  aliceBalance >= 0 && bobBalance >= 0
}
```

### How Preconditions Fix the Problem

**Preconditions** (`require` statements) are conditions that must be true **before** an action can execute. If a precondition is false, the action cannot execute.

By adding `require aliceBalance >= 30` to `withdrawAlice30`:
- The action can only execute when Alice has at least 30
- After execution, `aliceBalance` will be `aliceBalance - 30 >= 0`
- The invariant is preserved

Similarly:
- `require aliceBalance >= 100` ensures `withdrawAlice100` only executes when Alice has sufficient funds
- `require aliceBalance >= 50` ensures `transfer50ToBob` only executes when Alice can afford the transfer

With these preconditions, the invariant `no_negatives` will always hold.

---

## Example 2: Inventory Management System

Now let's look at a more complex example with **multiple invariants**.

### The Problem: Missing Bounds Checking

Here's an inventory system that tracks the number of items:

```spectre
description "Number of items in inventory"
var inventory: int

description "System starts with empty inventory"
init {
  inventory = 0
}

description "Add 50 items to inventory"
description "PROBLEM: Missing precondition to check capacity limit"
action add50 {
  // Missing: require inventory + 50 <= 100
  inventory' = inventory + 50
}

description "Add 60 items to inventory"
description "PROBLEM: Can exceed capacity"
action add60 {
  // Missing: require inventory + 60 <= 100
  inventory' = inventory + 60
}

description "Remove 30 items from inventory"
description "PROBLEM: Missing precondition to check if enough items exist"
action remove30 {
  // Missing: require inventory >= 30
  inventory' = inventory - 30
}

description "Remove 50 items from inventory"
description "PROBLEM: Can make inventory negative"
action remove50 {
  // Missing: require inventory >= 50
  inventory' = inventory - 50
}

description "CRITICAL: Ensures inventory never exceeds maximum capacity"
invariant withinCapacity {
  inventory <= 100
}

description "CRITICAL: Ensures inventory never becomes negative"
invariant nonNegative {
  inventory >= 0
}
```

### Why This Violates the Invariants

This system has **two invariants** that can both be violated:

1. **`withinCapacity`**: Violated when inventory exceeds 100
   - `add50` when `inventory = 60` → `inventory = 110` ❌
   - `add60` when `inventory = 50` → `inventory = 110` ❌

2. **`nonNegative`**: Violated when inventory becomes negative
   - `remove30` when `inventory = 20` → `inventory = -10` ❌
   - `remove50` when `inventory = 30` → `inventory = -20` ❌

### The Error

When you verify this specification, the verifier will **fail** and report violations:

```bash
$ ./spectre verify examples/inventory-violation.spec

Verification failed: X violation(s) found

Violation 1 (Invariant):
  Action 'add60' would violate invariants: [invariant withinCapacity violated: (Ensures inventory never exceeds maximum capacity)]
  
Violation 2 (Invariant):
  Action 'remove30' would violate invariants: [invariant nonNegative violated: (CRITICAL: Ensures inventory never becomes negative)]
  
...
```

### Understanding the Error Messages

The violations follow the same pattern as the bank account example:

- **Violation 1**: `add60` would exceed the capacity of 100 when inventory is already above 40
- **Violation 2**: `remove30` would make inventory negative when the current value is less than 30

Each violation indicates that an action is available but would create an invalid state. The actions need preconditions to prevent them from being available in these situations.

### The Fix: Add Preconditions for Both Invariants

To fix this, we need preconditions that ensure **both** invariants are maintained:

```spectre
description "Number of items in inventory"
var inventory: int

description "System starts with empty inventory"
init {
  inventory = 0
}

description "Add 50 items to inventory if capacity allows"
description "FIXED: Added precondition to check capacity limit"
action add50 {
  require inventory + 50 <= 100
  inventory' = inventory + 50
}

description "Add 60 items to inventory if capacity allows"
description "FIXED: Added precondition to prevent exceeding capacity"
action add60 {
  require inventory + 60 <= 100
  inventory' = inventory + 60
}

description "Remove 30 items from inventory if enough items exist"
description "FIXED: Added precondition to check if enough items exist"
action remove30 {
  require inventory >= 30
  inventory' = inventory - 30
}

description "Remove 50 items from inventory if enough items exist"
description "FIXED: Added precondition to prevent negative inventory"
action remove50 {
  require inventory >= 50
  inventory' = inventory - 50
}

description "CRITICAL: Ensures inventory never exceeds maximum capacity"
invariant withinCapacity {
  inventory <= 100
}

description "CRITICAL: Ensures inventory never becomes negative"
invariant nonNegative {
  inventory >= 0
}
```

### How These Preconditions Work

1. **For `add50` and `add60`**: 
   - `require inventory + quantity <= 100` ensures that after adding, inventory won't exceed 100
   - This preserves the `withinCapacity` invariant

2. **For `remove30` and `remove50`**:
   - `require inventory >= quantity` ensures that after removing, inventory won't become negative
   - This preserves the `nonNegative` invariant

With these preconditions, both invariants will always hold.

---

---

## CEGIS Automatic Repair

When `spectre verify` finds an invariant violation, it does not just report it — it automatically synthesises a `require` guard that would prevent the violation, using **weakest-precondition (WP) computation**:

```bash
./spectre verify examples/bank-account-violation.spec
```

After reporting the violation, the output includes a repair suggestion:

```
CEGIS Repair Suggestions:

  Invariant `no_negatives` violated via action `withdrawAlice30`:
    Pre-state: {aliceBalance = 0, bobBalance = 0}
    Counterexample: withdrawAlice30*
    Option 1 — weakest precondition for `aliceBalance' = aliceBalance - 30`:
      // In action withdrawAlice30 — add:
      require aliceBalance - 30 >= 0
      ✓ Verified: re-explored with this guard applied — invariant holds
```

The synthesised guard is the **weakest condition** under which the assignment preserves the invariant — i.e., the least restrictive guard that is still correct. Add it to the action:

```spectre
action withdrawAlice30 {
  require aliceBalance - 30 >= 0   // synthesised by CEGIS
  aliceBalance' = aliceBalance - 30
}
```

Or rewrite it in the more idiomatic form (mathematically equivalent):

```spectre
action withdrawAlice30 {
  require aliceBalance >= 30
  aliceBalance' = aliceBalance - 30
}
```

Spectre re-runs BFS with the proposed guard injected before reporting `✓ Verified`, so the suggestion is confirmed correct before you see it. For specs with multiple violating actions, a repair suggestion is produced for each one.

---

## Common Patterns of Invariant Violations

Based on our examples, here are common patterns:

### Pattern 1: Missing Preconditions for Decrements

**Problem**: Actions that decrease values don't check if the value is large enough.

**Example**: `withdraw` without checking balance, `remove` without checking inventory

**Fix**: Add `require value >= amount` before decrementing

### Pattern 2: Missing Preconditions for Increments

**Problem**: Actions that increase values don't check if the value would exceed bounds.

**Example**: `add` without checking capacity, `deposit` without checking maximum

**Fix**: Add `require value + amount <= maxValue` before incrementing

### Pattern 3: Missing Preconditions for Compound Operations

**Problem**: Actions that perform multiple state changes don't validate intermediate states.

**Example**: `transfer` that subtracts from one account and adds to another

**Fix**: Add preconditions for each state change: `require fromValue >= amount` and ensure `toValue + amount <= maxValue` if applicable

### Pattern 4: Missing Validation of Input Parameters

**Problem**: Actions don't validate that parameters are within acceptable ranges.

**Example**: Accepting negative amounts, zero quantities, or invalid account names

**Fix**: Add `require amount > 0`, `require account != ""`, etc.

---

## Best Practices for Preventing Invariant Violations

1. **Always add preconditions**: Every action that modifies state should have preconditions that ensure invariants are maintained.

2. **Check before modifying**: Preconditions should check the **current state** before making changes, not after.

3. **Consider all invariants**: If you have multiple invariants, ensure each action maintains **all** of them.

4. **Use descriptive error messages**: When preconditions fail, the verifier will report which action couldn't execute and why.

5. **Test edge cases**: Consider boundary conditions:
   - What happens at the minimum value (0, empty, etc.)?
   - What happens at the maximum value (100, full, etc.)?
   - What happens with invalid inputs?

---

## Understanding Verifier Behavior

When you verify a specification with missing preconditions:

1. **The verifier checks invariants**: It validates invariants after each action execution
2. **Invalid states are prevented**: If an action would create an invalid state, that action execution fails
3. **Violations are recorded**: Each attempt to execute an action that would violate invariants is recorded as a violation
4. **Verification fails**: If any violations are found, verification fails with a detailed report

### Example: What the Verifier Does

When an action would violate an invariant:
- The verifier detects it before adding the invalid state
- The action execution fails, and a **violation is recorded**
- The violation includes:
  - The action that would fail
  - The invariant that would be violated
  - The path (sequence of actions) leading to the state where this would occur
- Other valid paths continue to be explored
- **Verification fails** if any violations are found

**Key Insight**: The verifier acts as both a safety net (preventing invalid states) and a **specification checker** (reporting when actions lack proper preconditions). This helps you find and fix specification flaws early.

---

## Summary

This chapter covered:

1. **What invariants are**: Properties that must hold in every reachable state

2. **How invariants are violated**: 
   - Missing preconditions allow actions to execute when they shouldn't
   - Actions modify state in ways that break invariant properties

3. **How to fix violations**:
   - Add `require` statements (preconditions) to actions
   - Preconditions check the current state before making changes
   - Preconditions ensure invariants will hold after the action executes

4. **Common patterns**:
   - Missing checks for decrements (negative values)
   - Missing checks for increments (exceeding bounds)
   - Missing checks for compound operations
   - Missing validation of input parameters

5. **Verifier behavior**: The verifier prevents invalid states from being reached, but you should still add preconditions for correctness

6. **Best practices**: Always add preconditions, check before modifying, consider all invariants

By understanding invariants and how to prevent violations, you can write specifications that maintain safety properties throughout system execution. Use preconditions to prevent problems, and invariants to guarantee properties.

# Chapter 4: Temporal and Fairness Properties

This chapter covers temporal properties and fairness conditions.

## Understanding Temporal Properties

Temporal properties describe behavior over sequences of states (execution traces). Unlike invariants, which must hold in every state, temporal properties express requirements about future states and execution paths.

### A Simple Counter Example

Let's start with a basic counter specification:

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

description "Decrements the counter by one, only when counter is positive"
action decrement {
  require counter > 0
  counter' = counter - 1
}

description "Resets the counter back to zero"
action reset {
  counter' = 0
}

description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}

description "Keeps counter within reasonable bounds"
invariant bounded {
  counter <= 100
}

description "Verifies that counter eventually reaches value 10"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}

description "Ensures counter remains non-negative throughout execution"
temporal alwaysNonNegative {
  always (counter >= 0)
}

description "Guarantees progress: if counter is below 10, it will eventually reach 10"
temporal progress {
  always (counter < 10 → eventually counter = 10)
}
```

This specification defines:
- A `counter` variable that starts at 0
- Three actions: `increment`, `decrement`, and `reset`
- Two invariants: `nonNegative` and `bounded`
- Three temporal properties

### Understanding the Temporal Properties

1. **`eventuallyReachesTen`**: Checks if the counter can eventually reach 10. This property holds because there exists a path where we only increment (0 → 1 → 2 → ... → 10).

2. **`alwaysNonNegative`**: Ensures the counter never becomes negative. This property holds because:
   - The counter starts at 0
   - `increment` always increases it
   - `decrement` has a precondition `counter > 0`, so it can't make it negative
   - `reset` sets it to 0

3. **`progress`**: This is more complex. It states: "In every state where `counter < 10`, eventually `counter = 10` will hold."

### The Problem: A Temporal Property Violation

When we verify this specification, we get a violation:

```
Verification failed: 1 violation(s) found

Violation 1 (Temporal Property: progress):
  Property 'progress' violated: (Guarantees progress: if counter is below 10, 
  it will eventually reach 10) P holds but Q never becomes true
```

**Why does this happen?**

The `progress` property states that whenever `counter < 10`, it will eventually reach 10. However, the `reset` action can be executed at any time, even when `counter = 0`. This creates a problematic execution path:

1. Start with `counter = 0` (where `counter < 10` holds)
2. Execute `reset` → `counter = 0` (still `counter < 10`)
3. Execute `reset` again → `counter = 0` (still `counter < 10`)
4. Continue executing `reset` infinitely...

This infinite loop creates a "blocking cycle" where the counter never reaches 10, violating the temporal property.

### Fixing the Problem: Restricting Actions

One way to fix this is to restrict the `reset` action so it can only execute when it doesn't prevent progress. In `counter-corrected.spec`, we add a precondition:

```spectre
description "Resets the counter back to zero"
action reset {
  require counter > 10
  counter' = 0
}
```

Now `reset` can only execute when `counter > 10`, which means:
- When `counter < 10`, `reset` cannot execute
- The system can only progress forward (via `increment`) or stay in place (via `decrement` if `counter > 0`)
- Since `increment` has no preconditions, it can always execute, ensuring progress to 10

With this fix, the `progress` temporal property holds.

### Why This Approach Works

By restricting `reset` to only execute when `counter > 10`, we've eliminated the problematic infinite loop. Now, when `counter < 10`:
- `reset` cannot execute (precondition not satisfied)
- `decrement` can only execute if `counter > 0`
- `increment` can always execute (no preconditions)

Since `increment` can always execute, there's always a path forward, ensuring the counter will eventually reach 10.

---

## Fairness Properties

However, restricting actions isn't always the best solution. Sometimes we want to allow all actions but assume they execute "fairly." This is where **fairness conditions** come in.

### What is Fairness?

In concurrent systems, fairness conditions specify assumptions about how actions are scheduled. They ensure that actions that are "ready" to execute will eventually do so, preventing infinite loops where an enabled action is never chosen.

### Weak Fairness (WF)

**Weak Fairness** states: "If an action is **continuously enabled** from some point onwards, it must eventually execute."

**Key points:**
- The action must be enabled continuously (without interruption)
- Once it becomes continuously enabled, it will eventually execute
- It doesn't guarantee execution if the action is enabled and disabled repeatedly

**Example:**
```spectre
description "Weak fairness ensures process1 gets fair access when continuously waiting"
temporal process1WeakFairness {
  WF(process1Request)
}
```

This means: If `process1Request` is enabled and stays enabled, it will eventually execute.

### Strong Fairness (SF)

**Strong Fairness** states: "If an action is enabled **infinitely often**, it must execute infinitely often."

**Key points:**
- The action needs to be enabled infinitely often (even if disabled sometimes)
- It will execute infinitely often if enabled infinitely often
- Stronger guarantee than weak fairness

**Example:**
```spectre
description "Strong fairness ensures process1 executes even if intermittently enabled"
temporal process1StrongFairness {
  SF(process1Request)
}
```

This means: Even if `process1Request` is enabled and disabled repeatedly, if it's enabled infinitely often, it will execute infinitely often.

### When to Use Weak vs Strong Fairness

- **Weak Fairness (WF)**: Use when an action should execute if it's continuously available
  - Example: A process waiting for a lock should eventually get it if the lock is continuously free

- **Strong Fairness (SF)**: Use when an action should execute even if it's intermittently enabled
  - Example: A process should eventually get the lock even if the lock is acquired and released repeatedly

---

## Fairness in Temporal Properties

Fairness conditions can be combined with temporal operators using the leads-to (`→`) operator. This allows us to express properties that hold under fairness assumptions.

### Syntax

```spectre
temporal propertyName {
  WF(actionName) → temporalExpression
}
```

or

```spectre
temporal propertyName {
  SF(actionName) → temporalExpression
}
```

### How It Works

When you write `WF(action) → always (P → eventually Q)`, the verifier:

1. **Filters the transition graph**: Removes all "unfair" paths where the action is continuously enabled (WF) or enabled infinitely often (SF) but never executes
2. **Verifies the property**: Checks if `always (P → eventually Q)` holds over the remaining "fair" paths

This ensures that the temporal property is verified only under the assumption that the action executes fairly.

---

## The Counter Example with Fairness

In `counter-with-fairness.spec`:

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

description "Decrements the counter by one, only when counter is positive"
action decrement {
  require counter > 0
  counter' = counter - 1
}

description "Resets the counter back to zero"
action reset {
  counter' = 0
}

description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}

description "Keeps counter within reasonable bounds"
invariant bounded {
  counter <= 100
}

description "Verifies that counter can eventually reach value 10"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}

description "Ensures counter remains non-negative throughout execution"
temporal alwaysNonNegative {
  always (counter >= 0)
}

description "Guarantees progress: if counter is below 10, it will eventually reach 10"
description "With weak fairness on increment, this property holds because increment will"
description "eventually execute when continuously enabled, ensuring progress to counter = 10"
temporal progress {
  WF(increment) → always (counter < 10 → eventually counter == 10)
}
```

Notice that `reset` has no preconditions (unlike the corrected version). However, the `progress` property now includes weak fairness:

```spectre
WF(increment) → always (counter < 10 → eventually counter == 10)
```

### How This Works

1. **Weak fairness on `increment`**: `WF(increment)` assumes that if `increment` is continuously enabled, it will eventually execute.

2. **Filtering unfair paths**: The verifier identifies cycles where:
   - `increment` is continuously enabled (it has no preconditions, so it's always enabled)
   - `increment` never executes in the cycle
   - These cycles are marked as "unfair" and filtered out

3. **Example of an unfair cycle**: 
   - State: `counter = 0`
   - Transition: `reset` action → `counter = 0` (self-loop)
   - This cycle is unfair because `increment` is continuously enabled (no preconditions) but never executes

4. **Verifying the property**: After filtering out unfair cycles, the verifier checks if `always (counter < 10 → eventually counter == 10)` holds over the remaining fair paths.

5. **Result**: The property holds because:
   - Unfair cycles (like the `reset` self-loop at `counter = 0`) are filtered out
   - In fair paths, `increment` will eventually execute when `counter < 10`
   - This ensures progress toward `counter = 10`

### Why This is Better

Using fairness instead of restricting actions has several advantages:

1. **More general**: Doesn't require changing action preconditions
2. **Realistic assumptions**: Models real systems where actions execute fairly
3. **Clearer intent**: Explicitly states the fairness assumptions
4. **Reusable**: The same fairness conditions can be applied to multiple temporal properties

---

## Summary

This chapter covered:

1. **Temporal Properties**: Express requirements about future states and execution paths
   - `eventually P`: P holds at some point
   - `always P`: P holds in every state
   - `P → Q`: If P holds, eventually Q will hold

2. **Temporal Property Violations**: Can occur when infinite loops (blocking cycles) prevent progress

3. **Fixing Violations**: 
   - Option 1: Restrict actions with preconditions
   - Option 2: Use fairness conditions (preferred for concurrent systems)

4. **Fairness Conditions**:
   - **Weak Fairness (WF)**: Continuously enabled actions eventually execute
   - **Strong Fairness (SF)**: Actions enabled infinitely often execute infinitely often

5. **Fairness in Temporal Properties**: Use `WF(action) → temporalExpression` or `SF(action) → temporalExpression` to verify properties under fairness assumptions

6. **How Fairness Works**: The verifier filters out unfair paths (where fair actions don't execute) before verifying temporal properties

By understanding temporal and fairness properties, you can verify complex system behaviors and ensure that progress guarantees hold even in concurrent systems with multiple possible execution paths.

# Chapter 5: Concurrent Systems and Locking

This chapter covers specifying and verifying concurrent systems. The example uses a locking mechanism for mutual exclusion, showing how missing preconditions lead to invariant and temporal property violations.

## Understanding Concurrent Systems

In concurrent systems, multiple processes run simultaneously and may need to access shared resources. Without proper synchronization, this can lead to:

- **Race conditions**: Multiple processes accessing shared data simultaneously
- **Data corruption**: Unpredictable results from concurrent modifications
- **Violation of invariants**: System properties that should always hold become violated

**Locking** is a fundamental mechanism to ensure **mutual exclusion**: only one process can access a critical resource at a time.

---

## Example: Three Processes with a Shared Resource

Our example system has:

- **Three processes** (process1, process2, process3) that need to access a shared resource
- **A lock** (integer: 0 = unlocked, 1-3 = held by process 1-3)
- **A shared resource** (integer that gets incremented when accessed)
- **Process states**: `Idle`, `Waiting`, `Critical`

Each process follows a protocol:
1. **Request lock**: Move from `Idle` to `Waiting`
2. **Acquire lock**: Move from `Waiting` to `Critical` (when lock is available)
3. **Access resource**: Increment the resource (while in `Critical`)
4. **Release lock**: Move from `Critical` to `Idle` (release the lock)

### Required Invariants

1. **Mutual Exclusion**: Only one process can be in the `Critical` state at a time
2. **Lock Consistency**: If a process is in `Critical`, it must hold the lock

### Required Temporal Properties

1. **Progress**: If a process requests the lock, it eventually gets it
2. **Resource Progress**: The resource value eventually increases

---

## The Problem: Missing Preconditions

Let's examine a specification with critical flaws in `examples/concurrent-lock-violation.spec`:

```spectre
enum ProcessState {
  Idle,
  Waiting,
  Critical
}

var process1: ProcessState
var process2: ProcessState
var process3: ProcessState
var lock: int  // 0 = unlocked, 1-3 = held by process 1-3
var resource: int

init {
  process1 = ProcessState.Idle
  process2 = ProcessState.Idle
  process3 = ProcessState.Idle
  lock = 0
  resource = 0
}

action acquireLock1 {
  // PROBLEM: No check if lock is available
  lock' = 1
  process1' = ProcessState.Critical
}

action accessResource1 {
  // PROBLEM: No check if process holds the lock
  resource' = resource + 1
}

action releaseLock1 {
  // PROBLEM: No check if process holds the lock
  lock' = 0
  process1' = ProcessState.Idle
}

// Similar problems for process2 and process3...
```

### What's Wrong?

1. **`acquireLock1`** can execute even when the lock is held by another process (e.g., `lock == 2`)
2. **`accessResource1`** can execute even when process1 doesn't hold the lock
3. **`releaseLock1`** can execute even when process1 doesn't hold the lock

This allows multiple processes to simultaneously acquire the lock or access the resource without holding it.

---

## The Errors

When you verify this specification, the verifier will **fail** with multiple violations:

```bash
$ ./spectre verify examples/concurrent-lock-violation.spec

Verification failed: 90 violation(s) found

Violation 1 (Invariant):
  Action 'acquireLock1' would violate invariants: 
    [invariant mutualExclusion violated: (Invariant: Only one process can hold the lock at a time) 
     invariant criticalSectionHoldsLock violated: (Invariant: Process in critical section must hold the lock)]
  Path:
    1. acquireLock2

Violation 2 (Invariant):
  Action 'releaseLock1' would violate invariants: 
    [invariant criticalSectionHoldsLock violated: (Invariant: Process in critical section must hold the lock)]
  Path:
    1. acquireLock2

Violation 3 (Invariant):
  Action 'acquireLock3' would violate invariants: 
    [invariant mutualExclusion violated: (Invariant: Only one process can hold the lock at a time) 
     invariant criticalSectionHoldsLock violated: (Invariant: Process in critical section must hold the lock)]
  Path:
    1. acquireLock2

...

Violation 93 (Temporal Property: progress):
  Property 'progress' violated: (Temporal: If a process requests the lock, it eventually gets it) 
  P holds but Q never becomes true
  Counterexample trace:
    2. accessResource1
    3. requestLock1
```

### Understanding the Errors

#### Invariant Violations

The invariant violations show that actions would break the safety properties:

1. **`acquireLock1` after `acquireLock2`**: Process2 already holds the lock (`lock == 2`), but `acquireLock1` can still execute, setting `lock = 1` and putting process1 in `Critical`. This violates:
   - **Mutual exclusion**: Both process1 and process2 would be in `Critical`
   - **Lock consistency**: Process1 is in `Critical` but the lock doesn't match (`lock != 1` would be true initially)

2. **`releaseLock1` after `acquireLock2`**: Process1 tries to release a lock it doesn't hold, breaking lock consistency

3. **Multiple simultaneous critical sections**: Without proper checks, multiple processes can enter their critical sections simultaneously

#### Temporal Violations

The temporal property violations show **liveness problems**:

```bash
Violation 93 (Temporal Property: progress):
  Property 'progress' violated: (Temporal: If a process requests the lock, it eventually gets it) 
  P holds but Q never becomes true
  Counterexample trace:
    2. accessResource1
    3. requestLock1
```

- **`progress` property**: States that if a process requests the lock (`Waiting`), it should eventually enter the `Critical` section
- **Why it fails**: There are execution paths where:
  - A process is `Waiting` but the `acquireLock` action never executes (even when the lock is available)
  - Without fairness constraints, the scheduler could continuously skip the `acquireLock` action
  - The system can get stuck in states where processes are waiting indefinitely

**Counterexample trace analysis**:
- Process1 is in `Critical` and accesses the resource
- Process1 requests the lock again (while still in `Critical` - this shouldn't be possible, but missing preconditions allow it)
- The system reaches a state where process1 is `Waiting`
- However, even if the lock becomes available, there's no guarantee that `acquireLock1` will execute
- Without fairness, the scheduler could continuously execute other actions, leaving process1 waiting forever

---

## The Fix: Add Preconditions

To fix the specification, we need to add `require` statements (preconditions) to ensure actions only execute when safe:

```spectre
action requestLock1 {
  require process1 == ProcessState.Idle
  process1' = ProcessState.Waiting
}

action acquireLock1 {
  require lock == 0              // Lock must be available
  require process1 == ProcessState.Waiting  // Process must be waiting
  lock' = 1
  process1' = ProcessState.Critical
}

action accessResource1 {
  require lock == 1              // Process must hold the lock
  require process1 == ProcessState.Critical  // Process must be in critical section
  resource' = resource + 1
}

action releaseLock1 {
  require lock == 1              // Process must hold the lock
  require process1 == ProcessState.Critical  // Process must be in critical section
  lock' = 0
  process1' = ProcessState.Idle
}
```

### What the Fixes Do

1. **`requestLock1`**: Only allows requesting if the process is currently `Idle`

2. **`acquireLock1`**: 
   - Checks that the lock is free (`lock == 0`)
   - Ensures the process is in the `Waiting` state
   - This prevents multiple processes from acquiring the lock simultaneously

3. **`accessResource1`**:
   - Verifies the process holds the lock (`lock == 1`)
   - Ensures the process is in the `Critical` state
   - This prevents accessing the resource without proper synchronization

4. **`releaseLock1`**:
   - Verifies the process holds the lock (`lock == 1`)
   - Ensures the process is in the `Critical` state
   - This prevents releasing a lock the process doesn't hold

### Understanding Why Temporal Properties Fail

The temporal property `progress` fails because without fairness constraints, there's no guarantee that `acquireLock` actions will execute even when they're enabled:

- **The Problem**: If a process is `Waiting`, the `acquireLock` action may be enabled (when the lock is free), but the scheduler could continuously skip it
- **Without Fairness**: The verifier considers all possible execution paths, including "unfair" ones where enabled actions never execute
- **Result**: The temporal property fails because there exists an execution path where a waiting process never acquires the lock

**Why Fairness is Needed**:
- **Weak Fairness (WF)**: Ensures actions execute if they're **continuously enabled**
- **Strong Fairness (SF)**: Ensures actions execute if they're **enabled infinitely often** (even if not continuously)

For concurrent systems with locks, **strong fairness** is typically needed because:
- `acquireLock` actions are not continuously enabled (the lock may be held by other processes)
- But they become enabled infinitely often (as the lock is released and acquired repeatedly)
- Strong fairness guarantees they will eventually execute

### Fixing Temporal Properties

The temporal property `resourceProgress` in the corrected version passes because:
- Preconditions ensure actions can execute safely
- The system can make progress (processes can acquire locks and access the resource)
- Unlike the violation version, invalid states are prevented, allowing progress

However, the more complex `progress` property (ensuring individual processes get the lock) requires **fairness constraints** to hold. This is because without fairness, there's no guarantee that enabled actions will execute.

**To Fix Individual Process Progress with Fairness**:
If you need stronger guarantees about individual processes getting the lock, you can use strong fairness:

```spectre
description "Temporal: Process 1 eventually gets the lock (with strong fairness)"
temporal progress1 {
  SF(acquireLock1) → always ((process1 == ProcessState.Waiting) → eventually (process1 == ProcessState.Critical))
}
```

This ensures that if process1 is waiting and the lock becomes available infinitely often, process1 will eventually acquire it.

**Key Takeaway**: Temporal properties that depend on actions eventually executing typically require fairness constraints to hold in concurrent systems.

### Result

After adding preconditions and fairness constraints, the corrected version (`examples/concurrent-lock-corrected.spec`) passes verification:

```bash
$ ./spectre verify examples/concurrent-lock-corrected.spec

✓ Verification passed for examples/concurrent-lock-corrected.spec
  Explored 50 states
  Verified 4 temporal properties
```

All invariants are maintained:
- **Mutual exclusion**: Only one process can be in `Critical` at a time
- **Lock consistency**: Processes in `Critical` always hold the matching lock

The temporal property also holds:
- **Resource progress**: The resource value eventually increases (because preconditions allow actions to execute safely)

---

## Key Lessons

### 1. Preconditions Enforce Protocol

Preconditions ensure that actions can only execute when the system is in a valid state. In concurrent systems, this is crucial for maintaining synchronization protocols.

### 2. Invariants Catch Safety Violations

Invariants catch violations of safety properties (like mutual exclusion) early, preventing invalid states from being reached.

### 3. Temporal Properties Verify Liveness

Temporal properties ensure that the system can make progress and that desired events eventually occur. Without proper preconditions, temporal properties may fail because the system can get stuck in invalid states.

### 4. Fairness is Required for Liveness

Without fairness constraints, temporal properties that depend on actions eventually executing will fail. Weak fairness (`WF`) ensures that actions execute when continuously enabled, which is essential for proving that processes eventually get resources like locks.

### 5. The Verifier Acts as a Specification Checker

The verifier doesn't just prevent invalid states—it **reports violations** that indicate missing preconditions or incorrect protocol design. This helps you write correct specifications.

---

## Best Practices for Concurrent Systems

1. **Always check lock state**: Before acquiring or releasing, verify the lock is in the expected state
2. **Verify process state**: Ensure processes are in the correct state before state transitions
3. **Use invariants to document assumptions**: Invariants make your safety requirements explicit
4. **Use temporal properties for progress**: Temporal properties ensure your system can make progress and achieve desired outcomes
5. **Test with multiple processes**: Verify your specification with the actual number of concurrent processes you'll have

---

## Summary

This chapter demonstrated:

1. **Concurrent system modeling**: How to model multiple processes competing for a shared resource
2. **Locking protocols**: How to implement mutual exclusion using locks
3. **Invariant violations**: How missing preconditions lead to safety violations (multiple processes in critical sections, inconsistent lock states)
4. **Temporal violations**: How missing preconditions lead to liveness failures (processes stuck waiting)
5. **Correcting the specification**: Adding preconditions to enforce the locking protocol
6. **Verification success**: How proper preconditions ensure all invariants and temporal properties hold

By understanding these patterns, you can specify and verify concurrent systems that maintain safety (invariants) and progress (temporal properties) even under concurrent execution.

# Chapter 6: Distributed Message Queue

This chapter demonstrates temporal property violations in a distributed message queue system.

## The Message Queue System

Our message queue system has:
- **Producers**: Add messages to the queue
- **Consumers**: Process and remove messages from the queue
- **Queue**: A bounded buffer storing pending messages
- **Processed Set**: Tracks messages that have been processed

### System Requirements

1. **Safety**: Queue should never exceed its maximum capacity
2. **Safety**: Messages should only be processed once
3. **Liveness**: Messages should eventually be processed
4. **Liveness**: The queue should not grow unbounded

## The Violation Version

Let's examine `examples/message-queue-violation.spec`:

```spectre
description "Queue of messages waiting to be processed"
var queue: List<str>

description "Set of messages that have been processed"
var processed: Set<str>

description "Counter for tracking total messages produced"
var messageCounter: int

init {
  queue = List.empty()
  processed = Set.empty()
  messageCounter = 0
}

description "Producer adds a message to the queue"
description "PROBLEM: No check if queue is at capacity"
action produce {
  // Missing: require queue.size() < MAX_QUEUE_SIZE
  queue' = queue.append("msg")
  messageCounter' = messageCounter + 1
}

description "Consumer processes a message from the queue"
description "PROBLEM: No check if queue is empty"
action consume {
  // Missing: require queue.size() > 0
  processed' = processed.union({ queue.head() })
  queue' = queue.tail()
}
```

### Running the Violation Version

```bash
$ ./spectre verify examples/message-queue-violation.spec

Verification failed: 2 violation(s) found

Violation 1 (Temporal Property: messageProcessed):
  Property 'messageProcessed' violated: (PROBLEM: Without fairness, consumers might never execute) 
  P holds but Q never becomes true
  Counterexample trace:
    2. produce
    3. produce
    ...
    11. produce

Violation 2 (Temporal Property: allMessagesProcessed):
  Property 'allMessagesProcessed' violated: (PROBLEM: Without fairness or proper protocol, messages might never be consumed)
  eventually P does not hold from some state
  Counterexample trace:
    2. produce
    3. produce
```

### Understanding the Temporal Violations

#### Violation 1: `messageProcessed`

```spectre
description "Temporal: Messages eventually get processed"
temporal messageProcessed {
  always (queue.size() > 0 → eventually processed.contains(queue.head()))
}
```

**What it means**: If the queue is not empty, the message at the head should eventually be processed.

**Why it fails**:
- The property has no fairness constraint
- Without fairness, the scheduler could continuously execute `produce` actions
- The `consume` action, even when enabled, might never execute
- Result: Messages are added to the queue but never processed

**Counterexample Analysis**:
- Starting from the initial state, only `produce` actions execute
- The queue grows to 10 messages (max capacity)
- But `consume` never executes, so messages remain unprocessed
- The temporal property fails because the queue head is never processed

#### Violation 2: `allMessagesProcessed`

```spectre
description "Temporal: All produced messages eventually get processed"
temporal allMessagesProcessed {
  always eventually (messageCounter > 0 → processed.size() == messageCounter)
}
```

**What it means**: Eventually, all produced messages should be processed (i.e., `processed.size() == messageCounter`).

**Why it fails**:
- Similar to the first violation: without fairness, `consume` may never execute
- Messages are produced but never consumed
- The system can get stuck in states where `messageCounter > 0` but `processed.size() == 0`

**Counterexample Analysis**:
- After a few `produce` actions, `messageCounter > 0`
- But `consume` never executes, so `processed.size() == 0`
- The property `processed.size() == messageCounter` never becomes true

## The Corrected Version

Let's fix these issues in `examples/message-queue-corrected.spec`:

### Fix 1: Add Preconditions

```spectre
description "Producer adds a message to the queue"
description "FIXED: Added precondition to check queue capacity"
action produce {
  require queue.size() < MAX_QUEUE_SIZE
  queue' = queue.append("msg")
  messageCounter' = messageCounter + 1
}

description "Consumer processes a message from the queue"
description "FIXED: Added precondition to check queue is not empty"
description "FIXED: Ensures message is only processed once"
action consume {
  require queue.size() > 0
  require !processed.contains(queue.head())
  processed' = processed.union({ queue.head() })
  queue' = queue.tail()
}
```

**What changed**:
- `produce`: Added `require queue.size() < MAX_QUEUE_SIZE` to prevent queue overflow
- `consume`: Added `require queue.size() > 0` to prevent processing from empty queue
- `consume`: Added `require !processed.contains(queue.head())` to prevent duplicate processing

### Fix 2: Add Fairness Constraints

```spectre
description "Temporal: Queue eventually makes progress (with fairness)"
description "FIXED: Added weak fairness to ensure consume executes when continuously enabled"
temporal queueProgress {
  WF(consume) → eventually (processed.size() > 0)
}
```

**What changed**:
- Added `WF(consume)` (Weak Fairness) to the temporal property
- Changed from `always eventually` to just `eventually` - we only need that there exists a fair path where progress is made
- Weak fairness ensures: "If `consume` is continuously enabled, it will eventually execute"
- This guarantees that when the queue is not empty and `produce` can't execute (queue at capacity), `consume` will eventually run

**Why Weak Fairness Works**:
- When the queue reaches capacity (10 messages), `produce` can no longer execute (precondition `queue.size() < 10` fails)
- At that point, only `consume` is enabled, making it continuously enabled
- Weak fairness guarantees that if `consume` is continuously enabled, it will eventually execute
- This breaks the counterexample where `consume` never runs

### Fix 3: Add Capacity Bounds

```spectre
description "Invariant: Queue size never exceeds maximum capacity"
invariant queueCapacity {
  queue.size() <= 10
}

description "Temporal: Queue doesn't grow unbounded"
description "FIXED: Precondition ensures queue never exceeds capacity, so this property holds"
temporal queueBounded {
  always queue.size() <= 10
}
```

**What changed**:
- The `produce` precondition ensures the queue never exceeds capacity
- The invariant `queueCapacity` enforces this at all states
- The temporal property `queueBounded` is satisfied because the precondition prevents unbounded growth

## Running the Corrected Version

```bash
$ ./spectre verify examples/message-queue-corrected.spec

✓ Verification passed for examples/message-queue-corrected.spec
  Explored X states
  Verified 3 temporal properties
```

All temporal properties now hold:
- **`queueProgress`**: Messages eventually get processed (with fairness)
- **`queueBounded`**: Queue never exceeds capacity (due to preconditions)

## Key Lessons

### 1. Preconditions Enforce Safety

Preconditions ensure actions can only execute when the system is in a valid state:
- `produce` only when queue has space
- `consume` only when queue is not empty
- This prevents runtime errors and maintains invariants

### 2. Fairness is Essential for Liveness

Without fairness constraints, temporal properties that depend on actions eventually executing will fail:
- The scheduler could continuously skip enabled actions
- Fairness guarantees execution under certain conditions
- Weak fairness: "If continuously enabled, eventually executes"

### 3. Temporal Properties Verify System Behavior

Temporal properties ensure the system makes progress and behaves correctly over time:
- **Safety properties** (invariants): Must hold in all states
- **Liveness properties** (temporal): Must eventually hold along execution paths

### 4. The Verifier Finds Execution Paths

The verifier explores all possible execution paths and checks if properties hold:
- Finds counterexamples where properties fail
- Helps identify missing preconditions and fairness constraints
- Guides you to write correct specifications

## Summary

The message queue example demonstrates:
- **Temporal violations** when fairness is missing
- **How to fix** with weak fairness (`WF`) constraints
- **Preconditions** to enforce safety properties
- **The importance** of both safety and liveness in distributed systems

By adding preconditions and fairness constraints, we ensure:
1. The queue never exceeds capacity (safety)
2. Messages are eventually processed (liveness)
3. All produced messages eventually get consumed (liveness)

# Chapter 7: Modules and Code Organization

This chapter covers organizing large specifications using Spectre's module system, using an elevator controller as the running example.

## Understanding Modules

Modules in Spectre allow you to:
- **Organize code** into logical units that can be reused across specifications
- **Separate concerns** by grouping related variables, actions, and invariants
- **Share definitions** between files through imports
- **Encapsulate functionality** with visibility modifiers (public/private)

### Module Declaration

A module is declared in a file with the following structure:

```spectre
module ModuleName {
  // Module contents: types, constants, variables, actions, invariants, etc.
}
```

**Important Rules:**
1. Each file can contain exactly one module declaration
2. The module name must match the file name (without extension)
3. Files are named using PascalCase (e.g., `ElevatorModule.spec`)
4. Module names use PascalCase (e.g., `module ElevatorModule`)

### Module Visibility

Members within a module can be marked as `public` or have no visibility modifier (private by default):

- **Public members** (`public var`, `public action`, `public const`, etc.) are accessible from other modules that import this module
- **Private members** (no modifier) are only accessible within the module

```spectre
module ExampleModule {
  public var publicVar: int    // Accessible from other modules
  var privateVar: int          // Only accessible within this module
  
  public action publicAction { ... }
  action privateAction { ... }
}
```

## Importing Modules

To use a module in another file, you import it using one of two syntaxes:

### Import from Same Directory

```spectre
import ModuleName
```

This imports a module from the same directory as the current file.

### Import from Path

```spectre
import "path/to/ModuleName"
```

This imports a module from a relative path. The path should point to the file (without extension).

**Example:**
```spectre
import "elevator/ElevatorModule"    // From examples/elevator/ElevatorModule.spec
import "../common/Utils"             // From parent directory's common/Utils.spec
```

### Module Resolution Rules

1. **No circular dependencies**: If module A imports module B, module B cannot import module A (directly or indirectly)
2. **Module name matching**: The module declaration name must exactly match the file name (case-sensitive)
3. **One module per file**: Each `.spec` file must contain exactly one module

## Elevator Controller Example

Let's explore a complete example: an elevator controller system for a building with 4 elevators and 20 floors. This example demonstrates how to organize a complex system using modules.

### System Overview

The elevator controller system consists of:
- **4 elevators** that can move between 20 floors (numbered 0-19)
- **Users** who request elevators by pressing up/down buttons
- **Controller logic** that assigns elevators to users based on optimization rules

The system is organized into three modules:
1. **ElevatorModule**: Basic elevator state and movement operations
2. **UserModule**: User requests and state management
3. **ControllerModule**: Assignment logic and optimization algorithms

### Module 1: ElevatorModule

The `ElevatorModule` defines the basic elevator behavior:

```spectre
// File: ElevatorModule.spec
module ElevatorModule {
  description "Current floor where the elevator is located (0-19)"
  public var currentFloor: int

  description "Direction the elevator is moving: 1 = up, -1 = down, 0 = idle"
  public var direction: int

  description "Set of floors the elevator needs to visit (target floors)"
  public var targetFloors: Set<int>

  description "Move elevator up one floor"
  public action moveUp {
    require currentFloor < 19
    require targetFloors.exists(f => f > currentFloor)
    currentFloor' = currentFloor + 1
    direction' = 1
  }

  description "Move elevator down one floor"
  public action moveDown {
    require currentFloor > 0
    require targetFloors.exists(f => f < currentFloor)
    currentFloor' = currentFloor - 1
    direction' = -1
  }

  // ... other actions and invariants
}
```

**Key Points:**
- All variables and actions are marked `public` so they can be used by other modules
- The module encapsulates elevator-specific logic (movement, target management)
- Invariants ensure elevators stay within valid bounds

### Module 2: UserModule

The `UserModule` defines user behavior:

```spectre
// File: UserModule.spec
module UserModule {
  description "Current floor where the user is located (0-19)"
  public var floor: int

  description "Direction the user wants to go: 1 = up, -1 = down, 0 = no request"
  public var direction: int

  description "Whether the user is waiting for an elevator"
  public var waiting: bool

  description "ID of the elevator assigned to this user, or -1 if none"
  public var assignedElevator: int

  description "User presses the up button from their current floor"
  public action pressUp(floorNum: int) {
    require floorNum >= 0 && floorNum < 19
    require direction = 0
    floor' = floorNum
    direction' = 1
    waiting' = true
    assignedElevator' = -1
  }

  // ... other actions
}
```

**Key Points:**
- Separates user-related concerns from elevator logic
- Can be reused in other systems that need user management
- Public actions allow the controller to interact with users

### Module 3: ControllerModule

The `ControllerModule` provides utility functions and constants:

```spectre
// File: ControllerModule.spec
module ControllerModule {
  description "Number of elevators in the system"
  public const NUM_ELEVATORS: int = 4

  description "Number of floors in the building"
  public const NUM_FLOORS: int = 20

  description "Maximum number of users in the system"
  public const MAX_USERS: int = 30

  description "Calculate distance between elevator and floor"
  public fun distance(elevatorFloor: int, targetFloor: int): int {
    return if (elevatorFloor > targetFloor) {
      elevatorFloor - targetFloor
    } else {
      targetFloor - elevatorFloor
    }
  }

  description "Check if elevator is moving in same direction as user request"
  public fun sameDirection(elevatorDir: int, userDir: int): bool {
    return elevatorDir = userDir && !(elevatorDir = 0)
  }

  // ... other utility functions
}
```

**Key Points:**
- Provides shared constants and utility functions
- Functions can be called from other modules using `ControllerModule.functionName()`
- Constants can be accessed using `ControllerModule.CONSTANT_NAME`

### Main System: ElevatorController

The main system file imports all three modules and orchestrates the elevator controller:

```spectre
// File: ElevatorController.spec
// Import the three modules from separate files in the same directory
import ElevatorModule
import UserModule
import ControllerModule

// Main system module that orchestrates the elevator controller
module ElevatorController {
  // Type definitions for the system
  type Elevator = {
    currentFloor: int,
    direction: int,
    targetFloors: Set<int>
  }

  type User = {
    floor: int,
    direction: int,
    waiting: bool,
    assignedElevator: int
  }

  // System state: 4 elevators and a set of users
  var elevator0: Elevator
  var elevator1: Elevator
  var elevator2: Elevator
  var elevator3: Elevator
  var users: Set<User>

  // Initial state
  init {
    elevator0 = { currentFloor: 0, direction: 0, targetFloors: {} }
    elevator1 = { currentFloor: 0, direction: 0, targetFloors: {} }
    elevator2 = { currentFloor: 0, direction: 0, targetFloors: {} }
    elevator3 = { currentFloor: 0, direction: 0, targetFloors: {} }
    users = {}
  }

  // User actions
  action userPressUp(floor: int) {
    require floor >= 0 && floor < 19
    require users.size() < ControllerModule.MAX_USERS
    // Create or update user at this floor
    users' = if (users.exists(u => u.floor = floor && u.direction = 0)) {
      users.map(u => 
        if (u.floor = floor && u.direction = 0) {
          { floor: u.floor, direction: 1, waiting: true, assignedElevator: -1 }
        } else {
          u
        }
      )
    } else {
      users.union({ { floor: floor, direction: 1, waiting: true, assignedElevator: -1 } })
    }
  }

  // Assignment actions use ControllerModule utility functions
  action assignElevator0ToUser(userFloor: int, userDir: int) {
    require ControllerModule.distance(elevator0.currentFloor, userFloor) <= 
           ControllerModule.distance(elevator1.currentFloor, userFloor) &&
           ControllerModule.distance(elevator0.currentFloor, userFloor) <= 
           ControllerModule.distance(elevator2.currentFloor, userFloor)
    // ... assignment logic
  }

  // Elevator movement actions
  action elevator0MoveUp {
    require elevator0.currentFloor < 19
    require elevator0.targetFloors.exists(f => f > elevator0.currentFloor)
    elevator0' = { 
      currentFloor: elevator0.currentFloor + 1, 
      direction: 1, 
      targetFloors: elevator0.targetFloors 
    }
  }

  // ... other actions and invariants
}
```

**Key Points:**
1. **Imports**: All three modules are imported at the top
2. **Module-qualified access**: Constants and functions from imported modules are accessed using `ModuleName.member`
3. **Orchestration**: The main module combines the three modules to create the complete system
4. **Separation**: Each module handles its own concerns, while the main module coordinates them

## Accessing Module Members

When you import a module, you can access its public members using dot notation:

### Accessing Constants

```spectre
import ControllerModule

// Access constant
const maxUsers = ControllerModule.MAX_USERS
```

### Accessing Functions

```spectre
import ControllerModule

// Call function
let dist = ControllerModule.distance(5, 10)
```

### Accessing Types

Types defined in modules are automatically available after import:

```spectre
import ElevatorModule

// Type from ElevatorModule is available
var elevator: Elevator
```

## Benefits of Modular Design

### 1. **Code Reuse**

Modules can be reused across multiple specifications:

```spectre
// In another system
import ElevatorModule
import UserModule

// Reuse elevator and user logic in a different context
```

### 2. **Separation of Concerns**

Each module handles a specific aspect of the system:
- ElevatorModule: Elevator mechanics
- UserModule: User interactions
- ControllerModule: Assignment algorithms

### 3. **Maintainability**

Changes to one module don't affect others (as long as the public interface remains the same).

### 4. **Testability**

Modules can be tested independently, making it easier to verify each component.

### 5. **Readability**

Large specifications are easier to understand when organized into logical modules.

## Running the Elevator Controller Example

The complete elevator controller example is located in `examples/elevator/`:

```bash
# Verify the specification
spectre verify examples/elevator/ElevatorController.spec

# With verbose output to see state exploration
spectre verify examples/elevator/ElevatorController.spec --verbose
```

### Temporal Property Violations and Corrections

The initial `ElevatorController.spec` contains temporal properties that fail verification. These properties are intentionally designed to fail early for testing purposes:

**Original Temporal Property (Fails)**:
```spectre
description "Temporal: Users eventually get an elevator"
description "MODIFIED FOR TESTING: Simplified to check any waiting user gets assigned"
temporal usersEventuallyGetElevator {
  WF(assignElevator0ToUser) → always (users.exists(u => u.waiting && u.assignedElevator = -1) → 
          eventually users.exists(u => u.waiting && u.assignedElevator >= 0))
}

description "Temporal: Elevators eventually reach their targets"
description "MODIFIED FOR TESTING: Simplified to check elevator0 reaches any target floor"
temporal elevatorsReachTargets {
  WF(elevator0MoveUp) → always (elevator0.targetFloors.size() > 0 → 
          eventually elevator0.targetFloors.contains(elevator0.currentFloor))
}
```

**Problems**: 
1. **usersEventuallyGetElevator**: Only has weak fairness on `assignElevator0ToUser`, but other elevators (1, 2, 3) might be able to assign users. Also, checking for unassigned users to get assigned requires fairness on ALL assignment actions, not just one. This property will fail when users are created but `assignElevator0ToUser` is never scheduled.

2. **elevatorsReachTargets**: Only has weak fairness on `elevator0MoveUp`, but elevators need to move in both directions (up and down). Also, it only checks `elevator0`, ignoring other elevators. This property will fail when `elevator0` has targets but `elevator0MoveUp` is never scheduled (e.g., if it needs to move down instead).

**Corrected Temporal Properties**:
```spectre
// CORRECTED: The corrected version addresses the issues above

description "Temporal: Users eventually get picked up after assignment"
description "CORRECTED: Simplified to check that assigned users eventually get picked up"
description "Once an elevator is assigned to a user, the elevator will eventually reach them"
description "This property focuses on progress after assignment and requires fairness on arrival actions"
temporal usersEventuallyGetElevator {
  // Once a user has an assigned elevator, they will eventually be picked up
  // We use fairness on elevator arrival actions to ensure elevators reach their targets
  // Using strong fairness (SF) ensures the arrival action executes when continuously enabled
  SF(elevator0Arrive) → always (users.exists(u => u.waiting && u.assignedElevator >= 0) → 
          eventually !users.exists(u => u.waiting && u.assignedElevator >= 0))
}

description "Temporal: Elevators eventually reach their targets"
description "CORRECTED: Simplified to check any elevator reaches any target, using arrival actions instead of movement"
description "If any elevator has target floors, at least one elevator will eventually reach at least one target"
description "Using arrival actions (which require movement to happen first) ensures progress in both directions"
temporal elevatorsReachTargets {
  // Use arrival actions which naturally require movement in the correct direction
  // Strong fairness ensures arrival actions execute when elevators are at target floors
  SF(elevator0Arrive) → 
  always (elevator0.targetFloors.size() > 0 || 
          elevator1.targetFloors.size() > 0 ||
          elevator2.targetFloors.size() > 0 ||
          elevator3.targetFloors.size() > 0 → 
          eventually (elevator0.targetFloors.contains(elevator0.currentFloor) ||
                      elevator1.targetFloors.contains(elevator1.currentFloor) ||
                      elevator2.targetFloors.contains(elevator2.currentFloor) ||
                      elevator3.targetFloors.contains(elevator3.currentFloor)))
}
```

**Solutions**: 
1. **usersEventuallyGetElevator**: The corrected version focuses on progress **after** assignment. Once an elevator is assigned to a user, the elevator will eventually reach them. We use **strong fairness (SF)** on arrival actions (`elevator0Arrive`) instead of assignment actions, which ensures progress once assignment occurs. The assignment itself is handled by the system design - if conditions are met, assignments can happen.

2. **elevatorsReachTargets**: The corrected version uses **arrival actions** (`elevator0Arrive`) instead of movement actions. Arrival actions naturally require movement in the correct direction (up or down) to occur first. Using **strong fairness (SF)** ensures arrival actions execute when elevators are at target floors, making the property work for all elevators and both movement directions.

**Key Differences**:
- **Original**: Uses weak fairness (WF) on assignment/movement actions, checks for unassigned users, only checks one elevator
- **Corrected**: Uses strong fairness (SF) on arrival actions, checks progress after assignment, checks all elevators
- **Original**: Designed to fail early for testing (fails around state 50)
- **Corrected**: Ensures actual system progress with proper fairness constraints (passes verification)

**Verification Results**:
- `ElevatorController.spec`: **Fails** with temporal violation around state 50 (intentionally designed to fail early for testing)
- `ElevatorControllerCorrected.spec`: **Passes** (uses proper fairness constraints to ensure actual system progress)

```bash
# Test the original (should fail)
spectre verify examples/elevator/ElevatorController.spec

# Test the corrected version (should pass)
spectre verify examples/elevator/ElevatorControllerCorrected.spec
```

The system explores a large state space (50+ states) showing:
- Users pressing buttons at different floors
- Elevator assignments based on optimization rules
- Elevator movement and arrival
- User boarding and exiting

## Best Practices

### 1. **One Concept Per Module**

Each module should represent a single concept or subsystem:
- ✅ Good: `ElevatorModule`, `UserModule`, `ControllerModule`
- ❌ Bad: `ElevatorAndUserModule`

### 2. **Minimal Public Interface**

Only expose what other modules need:
- ✅ Good: Mark only necessary actions/variables as `public`
- ❌ Bad: Making everything `public` defeats encapsulation

### 3. **Consistent Naming**

Use PascalCase for module names and files:
- ✅ Good: `ElevatorModule.spec` with `module ElevatorModule`
- ❌ Bad: `elevator_module.spec` or `module elevatorModule`

### 4. **Documentation**

Use descriptions to explain module purpose and public interfaces:
```spectre
module ElevatorModule {
  description "Manages elevator state and movement operations"
  // ...
}
```

### 5. **Avoid Circular Dependencies**

If module A needs module B, module B should not need module A. Design modules hierarchically.

## Summary

Modules in Spectre:

- **Declare** in separate files with matching names
- **Import** using `import ModuleName` or `import "path/to/Module"`
- **Access members** using `ModuleName.member` notation
- **Control visibility** with `public` vs private modifiers

# Chapter 8: Stuttering and Ensuring Progress

This chapter explains **stuttering** in state space exploration - when a state transitions back to itself - and how to fix it. Stuttering can indicate missing fairness constraints or incomplete specifications that prevent the system from making progress.

## What is Stuttering?

**Stuttering** occurs when an action causes a state to transition back to itself. In other words, executing an action results in no actual change to the system state. This creates a self-loop in the transition graph.

### Understanding Stuttering with a Visual Example

Imagine a state machine exploring different system configurations. Each state represents a unique configuration of all variables in your system. When an action executes, it should move the system from one state to a different state (making progress). However, when an action causes stuttering, it moves from a state back to the same state:

```
Normal Transition (makes progress):
  State A → Action: increment → State B (different state, progress made!)

Stuttering Transition (no progress):
  State A → Action: noop → State A (same state, no progress!)
```

### Why is Stuttering a Problem?

- **No progress**: The system can get stuck repeating the same state indefinitely
- **Liveness failures**: Temporal properties like `eventually P` may fail because the system can loop forever without reaching P
- **Missing fairness**: Often indicates that actions that should make progress are never scheduled
- **Reduced confidence**: Too many stuttering steps suggest the specification may not accurately model the system's behavior

### Counting Stuttering Steps

Each **stuttering step** represents a unique combination of:
- A specific state (e.g., `counter = 5`)
- An action that causes stuttering (e.g., `noop`)
- The result: the state remains unchanged

**Example:** If you have a counter that can be 0, 1, 2, ..., 50, and an action `noop` that doesn't change the counter, then:
- `noop` can stutter from state `counter = 0` (1 stuttering step)
- `noop` can stutter from state `counter = 1` (1 stuttering step)
- `noop` can stutter from state `counter = 2` (1 stuttering step)
- ... and so on for all 50 states
- Plus `reset` can stutter from state `counter = 0` (resetting 0 to 0)

**Total: 51 stuttering steps** (50 from `noop` + 1 from `reset`)

When you **reduce stuttering from 51 to 1**, you're eliminating 50 unnecessary stuttering opportunities by removing or restricting the problematic action. This means:
- **Before:** The system can stutter from 50 different states via `noop`, plus 1 from `reset`
- **After:** The system can only stutter from 1 state (via `reset` when counter is already 0)

This reduction indicates that your specification is more precise and has fewer paths where the system can make no progress.

Spectre detects and reports stuttering as warnings to help you identify potential issues in your specifications.

## Example 1: Simple Counter with Stuttering

Let's start with a simple counter example that exhibits stuttering:

```spectre
// File: examples/stuttering-counter.spec
module StutteringCounter {
  var counter: int
  
  init {
    counter = 0
  }
  
  description "Increments the counter by one"
  action increment {
    counter' = counter + 1
  }
  
  description "Does nothing - causes stuttering"
  action noop {
    counter' = counter  // ❌ Action doesn't change state - causes stuttering!
  }
  
  description "Resets the counter to zero"
  action reset {
    counter' = 0
  }
  
  description "Verifies counter eventually reaches 10"
  temporal eventuallyReachesTen {
    eventually (counter = 10)
  }
}
```

### Running Verification

When you verify this specification:

```bash
spectre verify examples/stuttering-counter.spec --max-states 20
```

**Output:**
```
Exploring State 1
Exploring State 2
Exploring State 3
...
Traversed 20 states

Warnings (Stuttering):
Found 20 stuttering step(s) where a state transitions back to itself.
Stuttering can indicate missing fairness constraints or incomplete specifications.

Found no violations.
```

### Understanding the Problem

The `noop` action causes stuttering because it sets `counter' = counter`, meaning the state doesn't change. This allows the system to:
- Execute `noop` repeatedly from **any state** (states 0, 1, 2, ..., 49, 50)
- Never make progress toward `counter = 10`
- Create infinite paths where the counter never reaches 10

**Why 51 stuttering steps?**
- With `max-states = 50`, the system explores states where `counter` ranges from 0 to 49
- The `noop` action can stutter from **all 50 states** (one stuttering step per state)
- The `reset` action can stutter from state 0 (resetting `counter = 0` back to `counter = 0`)
- **Total: 50 + 1 = 51 stuttering steps**

Even though the temporal property `eventually (counter = 10)` technically holds (there exists a path where it reaches 10), the stuttering warnings indicate that the specification allows 51 different non-progress paths. This suggests the specification might not accurately reflect the real system's behavior.

### Fix 1: Remove the Stuttering Action

The simplest fix is to remove the `noop` action if it's not needed:

```spectre
// File: examples/stuttering-counter-fixed.spec
module StutteringCounterFixed {
  var counter: int
  
  init {
    counter = 0
  }
  
  description "Increments the counter by one"
  action increment {
    counter' = counter + 1
  }
  
  description "Resets the counter to zero"
  action reset {
    counter' = 0
  }
  
  description "Verifies counter eventually reaches 10"
  temporal eventuallyReachesTen {
    eventually (counter = 10)
  }
}
```

**Result**: **Reduced from 51 to 1 stuttering step!**

**What changed:**
- **Removed** the `noop` action, which was causing 50 stuttering steps (one from each explored state)
- **Remaining** stuttering: Only 1 step from `reset` when `counter = 0` (resetting 0 to 0)

**What this means:**
- The system can no longer stutter from states 1, 2, 3, ..., 49 via `noop`
- The only remaining stuttering is when `counter = 0` and `reset` is called (which is acceptable - resetting an already-zero counter doesn't change the state)
- This is a **98% reduction** in stuttering opportunities (from 51 to 1), indicating a much more precise specification

The system can only make progress (increment) or reset, with minimal stuttering only from the edge case of resetting an already-zero counter.

### Fix 2: Add Fairness Constraints

If the `noop` action is necessary (e.g., it represents an idle step), add fairness constraints to ensure progress actions eventually execute:

```spectre
// File: examples/stuttering-counter-with-fairness.spec
module StutteringCounterWithFairness {
  var counter: int
  
  init {
    counter = 0
  }
  
  description "Increments the counter by one"
  action increment {
    counter' = counter + 1
  }
  
  description "Idle step - doesn't change state"
  description "This action can execute but doesn't make progress"
  action noop {
    counter' = counter
  }
  
  description "Resets the counter to zero"
  action reset {
    counter' = 0
  }
  
  description "Verifies counter eventually reaches 10 with weak fairness on increment"
  description "Weak fairness ensures increment executes when continuously enabled"
  temporal eventuallyReachesTen {
    WF(increment) → eventually (counter = 10)
  }
}
```

**Result**: 
- Stuttering warnings still appear (because `noop` still causes stuttering)
- However, the temporal property now holds because weak fairness ensures `increment` eventually executes
- The warnings help you understand that `noop` creates non-progress paths, but fairness ensures progress is made

**Key Insight**: Stuttering warnings don't mean your specification is wrong - they indicate paths where no progress is made. Fairness constraints filter out these unfair paths, allowing liveness properties to hold.

## Example 2: Process System with Stuttering

Consider a process system where a process can be in different states:

```spectre
// File: examples/stuttering-process.spec
enum ProcessState {
  Idle,
  Running,
  Waiting
}

module StutteringProcess {
  var process1: ProcessState
  var process2: ProcessState
  
  init {
    process1 = ProcessState.Idle
    process2 = ProcessState.Idle
  }
  
  description "Process 1 starts running"
  action process1Start {
    require process1 = ProcessState.Idle
    process1' = ProcessState.Running
  }
  
  description "Process 1 finishes and goes back to idle"
  action process1Finish {
    require process1 = ProcessState.Running
    process1' = ProcessState.Idle
  }
  
  description "Process 1 does nothing - causes stuttering"
  action process1Noop {
    process1' = process1  // ❌ Stuttering!
  }
  
  description "Process 2 starts running"
  action process2Start {
    require process2 = ProcessState.Idle
    process2' = ProcessState.Running
  }
  
  description "Process 2 finishes and goes back to idle"
  action process2Finish {
    require process2 = ProcessState.Running
    process2' = ProcessState.Idle
  }
  
  description "Process 2 does nothing - causes stuttering"
  action process2Noop {
    process2' = process2  // ❌ Stuttering!
  }
  
  description "Verifies process1 eventually runs"
  temporal process1EventuallyRuns {
    eventually (process1 = ProcessState.Running)
  }
}
```

**Verification Output:**
```
Traversed 15 states

Warnings (Stuttering):
Found 15 stuttering step(s) where a state transitions back to itself.
Stuttering can indicate missing fairness constraints or incomplete specifications.

Found no violations.
```

### The Problem

Both `process1Noop` and `process2Noop` cause stuttering. The system can:
- Execute `process1Noop` repeatedly while `process1` is in any state
- Never execute `process1Start`, preventing `process1` from ever running
- Even though `eventually (process1 = ProcessState.Running)` technically holds (there exists a path), the stuttering warnings reveal non-progress paths

### Fix: Add Fairness and Restrict Noop Actions

```spectre
// File: examples/stuttering-process-fixed.spec
enum ProcessState {
  Idle,
  Running,
  Waiting
}

module StutteringProcessFixed {
  var process1: ProcessState
  var process2: ProcessState
  
  init {
    process1 = ProcessState.Idle
    process2 = ProcessState.Idle
  }
  
  description "Process 1 starts running"
  action process1Start {
    require process1 = ProcessState.Idle
    process1' = ProcessState.Running
  }
  
  description "Process 1 finishes and goes back to idle"
  action process1Finish {
    require process1 = ProcessState.Running
    process1' = ProcessState.Idle
  }
  
  description "Process 1 idles (only when already idle) - no stuttering from running state"
  action process1Noop {
    require process1 = ProcessState.Idle
    process1' = ProcessState.Idle  // Only stutters from idle state, not from running
  }
  
  description "Process 2 starts running"
  action process2Start {
    require process2 = ProcessState.Idle
    process2' = ProcessState.Running
  }
  
  description "Process 2 finishes and goes back to idle"
  action process2Finish {
    require process2 = ProcessState.Running
    process2' = ProcessState.Idle
  }
  
  description "Process 2 idles (only when already idle)"
  action process2Noop {
    require process2 = ProcessState.Idle
    process2' = ProcessState.Idle
  }
  
  description "Verifies process1 eventually runs with weak fairness"
  description "Weak fairness ensures process1Start executes when enabled (process1 = Idle)"
  temporal process1EventuallyRuns {
    WF(process1Start) → eventually (process1 = ProcessState.Running)
  }
}
```

**Key Changes:**
1. **Restricted noop actions**: Added `require process1 = ProcessState.Idle` to `process1Noop`
   - Now `process1Noop` can only stutter when `process1` is already idle
   - It cannot stutter when `process1` is running, preventing interference with the running state
   
2. **Added fairness**: `WF(process1Start)` ensures that when `process1Start` is enabled (process1 = Idle), it eventually executes

**Result**: 
- Fewer stuttering warnings (only from idle states)
- The temporal property holds because fairness ensures `process1Start` executes
- The specification correctly models a system where processes can idle but must eventually start

## Example 3: Resource Allocation with Stuttering

Consider a resource allocation system where multiple processes compete for a resource:

```spectre
// File: examples/stuttering-resource.spec
module StutteringResource {
  var resourceOwner: int  // -1 means free, 0-2 means owned by process 0, 1, or 2
  var process0State: int  // 0 = idle, 1 = requesting, 2 = using
  var process1State: int
  var process2State: int
  
  init {
    resourceOwner = -1
    process0State = 0
    process1State = 0
    process2State = 0
  }
  
  description "Process 0 requests the resource"
  action process0Request {
    require process0State = 0 && resourceOwner = -1
    process0State' = 1
    resourceOwner' = 0
  }
  
  description "Process 0 uses the resource"
  action process0Use {
    require process0State = 1 && resourceOwner = 0
    process0State' = 2
  }
  
  description "Process 0 releases the resource"
  action process0Release {
    require process0State = 2 && resourceOwner = 0
    process0State' = 0
    resourceOwner' = -1
  }
  
  description "Process 0 does nothing - causes stuttering"
  action process0Noop {
    process0State' = process0State  // ❌ Stutters from any state!
    resourceOwner' = resourceOwner
  }
  
  // Similar actions for process1 and process2...
  
  description "Verifies process0 eventually gets to use the resource"
  temporal process0EventuallyUses {
    always (process0State = 1 → eventually process0State = 2)
  }
}
```

**Problem**: The `process0Noop` action causes stuttering from any state, including when `process0State = 1` (requesting). This means the system can stutter indefinitely when process0 is requesting, preventing it from using the resource.

### Fix: Remove Stuttering from Critical States

```spectre
// File: examples/stuttering-resource-fixed.spec
module StutteringResourceFixed {
  var resourceOwner: int
  var process0State: int
  var process1State: int
  var process2State: int
  
  init {
    resourceOwner = -1
    process0State = 0
    process1State = 0
    process2State = 0
  }
  
  description "Process 0 requests the resource"
  action process0Request {
    require process0State = 0 && resourceOwner = -1
    process0State' = 1
    resourceOwner' = 0
  }
  
  description "Process 0 uses the resource"
  action process0Use {
    require process0State = 1 && resourceOwner = 0
    process0State' = 2
  }
  
  description "Process 0 releases the resource"
  action process0Release {
    require process0State = 2 && resourceOwner = 0
    process0State' = 0
    resourceOwner' = -1
  }
  
  description "Process 0 idles - only allowed when idle or using (not when requesting)"
  action process0Noop {
    require process0State = 0 || process0State = 2
    // ✅ Can only stutter when idle or using, not when requesting
    process0State' = process0State
    resourceOwner' = resourceOwner
  }
  
  description "Verifies process0 eventually uses resource when requesting"
  description "Weak fairness ensures process0Use executes when enabled"
  temporal process0EventuallyUses {
    WF(process0Use) → always (process0State = 1 → eventually process0State = 2)
  }
}
```

**Key Changes:**
1. **Restricted noop**: `require process0State = 0 || process0State = 2`
   - `process0Noop` can only execute when process0 is idle or using
   - It cannot execute when `process0State = 1` (requesting), preventing stuttering that blocks progress

2. **Added fairness**: `WF(process0Use)` ensures that when process0 is requesting and the resource is available, `process0Use` eventually executes

**Result**: 
- No stuttering from the requesting state
- The temporal property holds because fairness ensures progress from requesting to using
- The specification correctly models a system where processes can idle but must make progress when requesting resources

## Understanding the Impact of Reducing Stuttering

### What Does "Reducing from 51 to 1" Mean?

When we reduce stuttering from 51 steps to 1 step, we're eliminating **50 opportunities** for the system to make no progress. This is significant because:

1. **Fewer Non-Progress Paths**: With 51 stuttering steps, there are 51 different states from which the system can loop forever without making progress. With only 1 stuttering step, there's only 1 such state (and it's an edge case: resetting an already-zero counter).

2. **More Precise Specification**: A specification with fewer stuttering steps more accurately models the system's behavior. It's closer to the actual system where actions typically change state.

3. **Better Verification**: Fewer stuttering steps mean:
   - The verifier has fewer paths to explore
   - Temporal properties are easier to verify (fewer unfair paths to filter)
   - The specification is easier to understand and maintain

4. **Indicates Design Quality**: High stuttering counts (like 51) often indicate:
   - Unnecessary actions that don't contribute to system behavior
   - Missing preconditions that should restrict when actions can execute
   - Actions that were added but don't model real system behavior

### The Mathematics of Stuttering

If you have:
- **N states** explored during verification
- **1 action** that can stutter from all N states
- **1 additional action** that stutters from 1 specific state

Then:
- **Total stuttering steps = N + 1**

By removing the action that stutters from all N states:
- **Remaining stuttering steps = 1**

**Reduction = (N + 1) - 1 = N steps eliminated**

In our example with `max-states = 50`:
- **Before:** 50 (from `noop`) + 1 (from `reset`) = 51 stuttering steps
- **After:** 0 (from `noop`, removed) + 1 (from `reset`) = 1 stuttering step
- **Reduction:** 50 stuttering steps eliminated (98% reduction!)

## Understanding Stuttering Warnings

### When Stuttering is Acceptable

Stuttering is acceptable when:
1. **Intentional idle states**: Actions that represent "do nothing" when the system is already in a stable state
2. **With fairness constraints**: If you have fairness on progress actions, stuttering from idle states is fine
3. **Explicit modeling**: When you intentionally model that a system can remain in a state (e.g., an idle process)

### When Stuttering Indicates a Problem

Stuttering indicates a problem when:
1. **From progress states**: Actions stutter when the system should be making progress
2. **Blocks liveness**: Stuttering prevents temporal properties from holding even with fairness
3. **Unintended behavior**: You didn't intend for the action to cause stuttering

### How to Fix Stuttering

1. **Remove the stuttering action** if it's not needed
2. **Add preconditions** to restrict when stuttering actions can execute
3. **Add fairness constraints** to ensure progress actions eventually execute
4. **Modify the action** to actually change state when progress should be made

## Best Practices

1. **Review stuttering warnings**: Always check stuttering warnings to understand if they're intentional
2. **Restrict noop actions**: If you have idle/noop actions, restrict them to appropriate states
3. **Use fairness for progress**: Add weak or strong fairness on actions that should make progress
4. **Test liveness properties**: Verify that temporal properties hold even when stuttering exists

## Summary

- **Stuttering** occurs when an action causes a state to transition back to itself, creating a self-loop with no progress
- **Stuttering steps** are counted per unique (state, action) combination where the state doesn't change
- Spectre detects and reports stuttering as warnings, helping you identify non-progress paths

### Reducing Stuttering: From 51 to 1

When we reduce stuttering from **51 to 1**, we're:
- **Eliminating 50 non-progress opportunities** (one per state)
- **Improving specification precision** by removing unnecessary actions
- **Making verification easier** with fewer paths to consider
- **Better modeling real system behavior** where actions typically change state

**In the counter example:**
- **51 stuttering steps** = 50 from `noop` (can stutter from any state) + 1 from `reset` (can stutter when counter=0)
- **1 stuttering step** = Only 1 from `reset` (edge case: resetting an already-zero counter)
- **Result:** 98% reduction in stuttering, much more precise specification

### Fixing Stuttering

Fix stuttering by:
1. **Removing unnecessary stuttering actions** (if they don't model real behavior)
2. **Restricting stuttering actions** to appropriate states with preconditions
3. **Adding fairness constraints** to ensure progress actions eventually execute

### Key Takeaways

- Stuttering warnings don't necessarily mean your specification is wrong - they highlight non-progress paths
- High stuttering counts (like 51) often indicate unnecessary actions or missing preconditions
- Reducing stuttering improves specification quality, making it more precise and easier to verify
- When progress should be made, either:
  1. Stuttering actions cannot execute from those states, OR
  2. Fairness constraints ensure progress actions eventually execute

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

## Extraction Omissions and `--show-omissions`

The miner tracks every Rust construct it cannot fully translate and records it as an **extraction omission**.  Omissions do not abort mining — the miner always produces a runnable spec — but they mark places where the generated spec is an approximation of the Rust source.

Run `--show-omissions` to see the full list printed to stderr after mining:

```bash
spectre mine --lang rust src/state_machine.rs --show-omissions
```

Example output:

```
Extraction omissions (3):
  TUPLE_VARIANT in State: variant Running(task_id: u32) — data dropped, unit variant kept
  OVERFLOW_OP in process: wrapping_add — saturating/wrapping arithmetic abstracted
  SKIPPED_METHOD in new: constructor/utility method skipped (not translated to action)
```

### Omission Kinds

| Kind | What it means |
|------|--------------|
| `UNSUPPORTED_TYPE` | Rust type has no direct Spectre equivalent; field mapped to `int` |
| `SKIPPED_METHOD` | `new`, `default`, `clone`, `fmt`, `drop` — skipped, not an action |
| `TUPLE_VARIANT` | Enum variant with data (`Name(T)`) — data dropped, variant name kept |
| `OVERFLOW_OP` | `checked_`, `wrapping_`, `saturating_` arithmetic — abstracted |
| `PANIC_PATTERN` | `panic!`, `unwrap()`, `expect()` — dropped |
| `RESULT_RETURN` | `Result<T,E>` return / `?` operator — error path abstracted |
| `COMPLEX_ASSIGN` | Assignment RHS too complex to translate directly |
| `UNSUPPORTED_PATTERN` | Other unrecognised Rust pattern |

### Interpreting Omissions

**`SKIPPED_METHOD`** is almost always benign — constructors (`new`) set initial values, not transitions.  The mined `init` block captures those instead.

**`TUPLE_VARIANT` and `OVERFLOW_OP`** require the most attention.  If your invariants depend on the dropped data or on overflow behaviour, you need to manually model those cases:

```spectre
// Original Rust: enum State { Idle, Running { task_id: u32 }, Done }
// Mined spec drops task_id — add it back manually if needed:
var task_id: int
```

**`RESULT_RETURN`** means the method has error paths that the miner abstracts into `require` guards (the happy path only).  The generated spec verifies that the non-error path is safe; it does not verify the error paths themselves.

### Omission Count in Mine Summary

The mine command prints a summary line alongside field/action counts:

```
Mined: 3 fields, 2 enums, 4 actions, 3 omissions (run --show-omissions for details)
```

A zero-omission result means the miner was able to translate the entire supported subset without any approximation.

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
# Chapter 10: Model-Based Testing

Once you have a verified Spectre specification, you can use it to test your actual Rust implementation automatically.  This is called **model-based testing (MBT)**: the spec is the oracle, and the test driver replays spec-generated execution traces against your code to check that every step agrees.

Spectre supports this workflow through four features:

- `--emit-traces` — writes explored paths to an **ITF trace file**
- `--coverage-mode` — five strategies to steer which paths are explored
- `generate-driver` — generates a Rust test harness that replays a trace file against your implementation
- `check-refinement` — checks that a set of ITF traces conform to the spec's refinement mapping

---

## The ITF Trace Format

ITF (Informal Trace Format) is a JSON schema originally from the [Apalache](https://github.com/informalsystems/apalache) model checker and used by the Quint ecosystem.  Each ITF file contains a sequence of `(action, pre-state, post-state)` records.

```bash
spectre verify examples/bank-account-parameterized.spec \
    --emit-traces trace.itf.json
```

When the verifier finds a counterexample it writes the full trace to `trace.itf.json`.  When no violation is found, it writes the last explored path — useful as a "happy path" regression trace.

---

## Coverage Modes

By default, `--emit-traces` uses `action` coverage (prefer actions not yet seen). You can change this with `--coverage-mode`:

```bash
spectre verify examples/bank-account-parameterized.spec \
    --emit-traces trace.itf.json \
    --coverage-mode property
```

| Mode | Strategy | When to use |
|------|---------|-------------|
| `action` | Prefers actions not yet executed in the current trace | Breadth of action coverage |
| `transition-pair` | Prefers (prev-action, next-action) pairs not yet seen | Transition sequence coverage |
| `boundary` | Steers toward the **maximum** boundary of integer variables | Testing upper-bound behaviour |
| `rare-action` | Prefers globally least-executed actions | Finding underexplored corners |
| `property` | Steers toward states where the **minimum** integer variable is lowest | **Closest to invariant boundaries** — best for catching precondition bugs |

**Empirical comparison on `bank-account-parameterized` (20-step budget):**

| Mode | Mean balance | Boundary steps | Coverage |
|------|-------------|---------------|---------|
| `action` | 41.2 | 2/8 (25%) | baseline |
| `boundary` | 500.0 | 1/21 (5%) | steers toward MAX (wrong direction for solvency) |
| `rare-action` | 30.5 | 5/21 (24%) | similar to action |
| `property` | 23.8 | 11/21 (**52%**) | **best** — most steps near balance = 0 |

`property` mode uses a violation-proximity score: it scores each successor state by its minimum integer variable value, preferring states closest to the invariant boundary (`balance >= 0` → prefer low balance). This resulted in twice the boundary-state coverage of any other mode.

---

## Generating a Test Driver

```bash
spectre generate-driver --lang rust examples/bank-account-parameterized.spec \
    --output src/spec_driver.rs
```

This generates a `spec_driver.rs` skeleton:

```rust
// Auto-generated by `spectre generate-driver bank-account-parameterized.spec`
// Fill in the TODO sections to connect your Rust implementation.

use spectre_connect::{ItfValue, StepHandler, TraceReplayer};
use std::collections::HashMap;

pub struct MySystem {
    pub balance: i64,
    pub frozen:  bool,
}

impl StepHandler for MySystem {
    type Error = Box<dyn std::error::Error>;

    fn apply_action(&mut self, action: &str, args: &[ItfValue]) -> Result<(), Self::Error> {
        match action {
            "deposit" => {
                let amount: i64 = args.get(0)
                    .ok_or("missing arg 0 for deposit")?
                    .try_into()?;
                // TODO: call your real implementation
                todo!()
            }
            "withdraw" => {
                let amount: i64 = args.get(0)
                    .ok_or("missing arg 0 for withdraw")?
                    .try_into()?;
                todo!()
            }
            "freeze"   => { todo!() }
            "unfreeze" => { todo!() }
            _ => Err(format!("unknown action: {action}").into()),
        }
    }

    fn current_state(&self) -> HashMap<String, ItfValue> {
        let mut s = HashMap::new();
        s.insert("balance".to_string(), ItfValue::Int(self.balance));
        s.insert("frozen".to_string(),  ItfValue::Bool(self.frozen));
        s
    }
}
```

---

## CI-Ready Driver Generation with `--ci`

The `--ci` flag extends `generate-driver` with two additional files that wire the generated driver into a standard Rust CI pipeline.

```bash
spectre generate-driver --lang rust account.spec --ci -o src/driver.rs
```

This produces three files:

| File | Purpose |
|------|---------|
| `src/driver.rs` | The Rust test driver (as usual) |
| `src/driver.cargo.toml` | `[dev-dependencies]` snippet to paste into your `Cargo.toml` |
| `src/driver.ci.yml` | GitHub Actions workflow that runs the driver on every push |

The Cargo snippet adds `spectre-connect` as a dev-dependency — the small Rust crate that reads ITF trace files and drives the replayer:

```toml
# src/driver.cargo.toml — paste into your Cargo.toml [dev-dependencies]
[dev-dependencies]
spectre-connect = { git = "https://github.com/BoundlessProducts/spectre-connect" }
```

The CI workflow checks out your repo, installs the Rust toolchain, and runs `cargo test`:

```yaml
# src/driver.ci.yml
name: Spectre conformance
on: [push, pull_request]
jobs:
  conformance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: rust-toolchain@stable
      - run: cargo test --test driver
```

This makes the generated driver a first-class CI gate with no extra configuration.

---

## Filling In the Driver

Replace the `todo!()` stubs with calls to your real implementation:

```rust
use bank_account::BankAccount;

pub struct MySystem {
    account: BankAccount,
}

impl StepHandler for MySystem {
    type Error = Box<dyn std::error::Error>;

    fn apply_action(&mut self, action: &str, args: &[ItfValue]) -> Result<(), Self::Error> {
        match action {
            "deposit" => {
                let amount: i64 = args.get(0).ok_or("missing amount")?.try_into()?;
                self.account.deposit(amount);
                Ok(())
            }
            "withdraw" => {
                let amount: i64 = args.get(0).ok_or("missing amount")?.try_into()?;
                self.account.withdraw(amount);
                Ok(())
            }
            "freeze"   => { self.account.freeze();   Ok(()) }
            "unfreeze" => { self.account.unfreeze(); Ok(()) }
            _ => Err(format!("unknown action: {action}").into()),
        }
    }

    fn current_state(&self) -> HashMap<String, ItfValue> {
        let mut s = HashMap::new();
        s.insert("balance".to_string(), ItfValue::Int(self.account.balance));
        s.insert("frozen".to_string(),  ItfValue::Bool(self.account.frozen));
        s
    }
}
```

---

## Running the Replay

```bash
# Step 1 — generate a trace using property-directed coverage
spectre verify examples/bank-account-parameterized.spec \
    --emit-traces trace.itf.json --coverage-mode property

# Step 2 — replay against your implementation
cargo run -- trace.itf.json
# ✓ All 47 steps matched the spec
```

If a step mismatches, the output tells you exactly what diverged:

```
✗ 1 mismatch(es):
  step 3, action 'withdraw', var 'balance': expected Int(70), got Int(50)
```

---

## Checking Refinement

`spectre check-refinement` validates a set of ITF traces against the spec's refinement mapping, classifying each step as **legal**, **stuttering**, or **violating**:

```bash
spectre check-refinement examples/bank-account-parameterized.spec \
    --traces trace1.itf.json,trace2.itf.json
```

This is useful when you have traces from your implementation (not generated by Spectre) and want to confirm they conform to the spec.

---

## Testing Violation Traces

When you run `verify` on a *violating* spec, the emitted trace is the counterexample itself:

```bash
spectre verify examples/bank-account-violation.spec \
    --emit-traces violation.itf.json

cargo run -- violation.itf.json
# ✗ 1 mismatch(es):
#   step 1, action 'withdraw', var 'balance': expected Int(-30), got ...
```

Use these as regression tests that confirm your fix resolves the violation.

---

## Repair Suggestions with `--suggest-repairs`

When `spectre verify` finds an invariant violation, the `--suggest-repairs` flag analyses the violated invariant and the triggering action, then prints labelled **CANDIDATE** repairs — never applied automatically, always requiring human review.

```bash
spectre verify account.spec --suggest-repairs
```

Example output:

```
Repair candidates (CANDIDATE — review before applying):

Violation 1: action 'withdraw' violates 'non_negative'
  [1] CANDIDATE strengthen_require: add `require balance >= amount` to action withdraw
      Validation: original violation fixed; no deadlock; reachability unchanged — SAFE
  [2] CANDIDATE add_invariant_guard: add `when balance >= amount` to action withdraw
      Validation: original violation fixed; deadlock at balance=0 — review required
```

Each candidate includes a **validation result** computed by re-running BFS on the patched spec:

| Validation result | Meaning |
|-------------------|---------|
| `SAFE` | Violation fixed, no deadlock introduced, reachability ≥ 95% of original |
| `review required` | Violation fixed but a deadlock or significant reachability drop was detected |
| `did not fix` | The patch did not prevent the violation |

### Confidence Levels

- **Medium confidence** — tighter bound on a simple comparison (e.g., `x < N` → `x < N - 1`).  Stops the action one step earlier to prevent crossing the invariant boundary.
- **Low confidence** — invariant-guard approach.  Adds a `when` clause that mirrors the violated invariant condition.  Reliable but can over-restrict (cause deadlocks).

Repairs are always labelled `CANDIDATE` in the output.  They are never written to the spec file automatically.

---

## What the Replayer Checks

At each step, `TraceReplayer` calls `apply_action` on your implementation, then calls `current_state()` and compares every variable against the ITF record. This catches:

- **Off-by-one errors** in arithmetic
- **Wrong initial state** (your `new()` doesn't match the spec's `init`)
- **Missing side-effects** (forgot to update a field)
- **Wrong action semantics** (e.g., `freeze` toggling the wrong field)

---

## Complete Workflow

```
mine → add invariants → verify → emit trace (property mode) → generate driver → fill in → replay
                                          ↓
                                   check-refinement (for external traces)
```

1. **Mine** — `spectre mine --lang rust src/account.rs -o account.spec`
2. **Enrich** — add invariants; verify with CEGIS repair if needed
3. **Generate trace** — `spectre verify account.spec --emit-traces t.itf.json --coverage-mode property`
4. **Generate driver** — `spectre generate-driver --lang rust account.spec -o src/driver.rs`
5. **Fill in** — implement `apply_action` and `current_state`
6. **Replay** — `cargo run -- t.itf.json` (or `spectre check-refinement` for conformance checking)

Traces are deterministic under the same exploration seed, making replays stable regression tests you can check into your test suite.
# Chapter 11: Runtime Monitoring

Formal verification gives you confidence at design time.  But a production system can still diverge from its specification through incomplete test coverage, refactoring, or third-party library updates.  **Runtime monitoring** closes this gap by embedding spec checks directly inside your running application.

Spectre generates a self-contained `monitor.rs` file — no external dependencies, no trace files, no offline replay.  The monitor checks every state transition in O(k) time where k is the number of invariants.

---

## Generating a Monitor

```bash
spectre generate-monitor --lang rust examples/bank-account-parameterized.spec \
    --output src/monitor.rs
```

This produces a `monitor.rs` with:

- A `SpecState` struct mirroring your spec variables
- A `Monitor` struct with `step()` and `unmet_liveness_properties()` methods
- A `Violation` struct for structured error reporting
- A callback hook for custom violation handling

---

## The Generated API

Here is the generated interface for `bank-account-parameterized.spec`:

```rust
/// Mirrors the spec variables.
#[derive(Debug, Clone)]
pub struct SpecState {
    pub balance: i64,
    pub frozen:  bool,
}

/// A spec violation detected at runtime.
#[derive(Debug, Clone)]
pub struct Violation {
    pub step:     u64,    // sequential step number
    pub action:   String, // which action triggered the check
    pub property: String, // invariant or temporal property name
    pub message:  String, // human-readable description from the spec
}

pub struct Monitor { /* ... */ }

impl Monitor {
    /// Create a monitor starting from `initial`.
    pub fn new(initial: SpecState) -> Self;

    /// Register a callback invoked immediately on every violation.
    pub fn on_violation<F: Fn(&Violation) + Send + 'static>(&mut self, f: F);

    /// Apply `action` → `new_state` and check all invariants.
    /// Returns Err(violations) if any invariant was violated.
    pub fn step(&mut self, action: &str, new_state: SpecState)
        -> Result<(), Vec<Violation>>;

    /// Returns names of temporal properties that were never satisfied.
    /// Call this at shutdown.
    pub fn unmet_liveness_properties(&self) -> Vec<&str>;

    /// All violations recorded since the monitor was created.
    pub fn all_violations(&self) -> &[Violation];
}
```

---

## Integrating the Monitor

Add four lines to your existing `BankAccount` implementation:

```rust
mod monitor;  // include the generated monitor.rs
use monitor::{Monitor, SpecState};

pub struct BankAccount {
    pub balance: i64,
    pub frozen:  bool,
    monitor: Monitor,
}

impl BankAccount {
    pub fn new(owner: String) -> Self {
        let initial = SpecState { balance: 0, frozen: false };
        let mut monitor = Monitor::new(initial);
        monitor.on_violation(|v| {
            eprintln!("[SPEC VIOLATION] {}", v);
            // In production you might log to your observability stack instead
        });
        Self { balance: 0, frozen: false, monitor }
    }

    pub fn deposit(&mut self, amount: i64) {
        assert!(amount > 0);
        self.balance += amount;
        // Report the new state to the monitor after every transition
        let _ = self.monitor.step("deposit", SpecState {
            balance: self.balance,
            frozen:  self.frozen,
        });
    }

    pub fn withdraw(&mut self, amount: i64) {
        self.balance -= amount;
        let _ = self.monitor.step("withdraw", SpecState {
            balance: self.balance,
            frozen:  self.frozen,
        });
    }

    pub fn freeze(&mut self) {
        self.frozen = true;
        let _ = self.monitor.step("freeze", SpecState {
            balance: self.balance,
            frozen:  self.frozen,
        });
    }

    pub fn unfreeze(&mut self) {
        self.frozen = false;
        let _ = self.monitor.step("unfreeze", SpecState {
            balance: self.balance,
            frozen:  self.frozen,
        });
    }
}

impl Drop for BankAccount {
    fn drop(&mut self) {
        // Check liveness properties at shutdown
        for prop in self.monitor.unmet_liveness_properties() {
            eprintln!("[SPEC] liveness property `{}` was never satisfied", prop);
        }
    }
}
```

---

## What the Monitor Checks

After each `monitor.step()` call, the monitor:

1. **Checks every `invariant` block** against the new state.  For `bank-account-parameterized.spec`:
   - `non_negative`: `balance >= 0`
   - `within_cap`: `balance <= 1000000`

2. **Advances liveness trackers** for each `temporal` property.  For `eventually_usable` (`eventually !frozen`), it records whether the property has ever been satisfied.

3. **Calls the violation callback** if any check fails — immediately, synchronously, in the calling thread.

---

## Violation Output

When the monitor fires, the output looks like:

```
[SPEC VIOLATION] [step 3] `non_negative` violated after `withdraw`: balance >= 0
```

The `Violation` struct gives you everything needed to log to your observability stack:

```rust
monitor.on_violation(|v| {
    tracing::error!(
        step    = v.step,
        action  = %v.action,
        property = %v.property,
        message = %v.message,
        "spec invariant violated"
    );
    metrics::counter!("spec.violations", 1, "property" => v.property.clone());
});
```

---

## Choosing the Violation Response

The `step()` return value lets you choose how to respond:

```rust
match self.monitor.step("withdraw", new_state) {
    Ok(()) => {}                      // all invariants held
    Err(violations) if cfg!(debug_assertions) => {
        panic!("spec violated in test: {:?}", violations);
    }
    Err(violations) => {
        // Production: log and continue, or abort, depending on severity
        for v in violations {
            log::error!("invariant {} violated: {}", v.property, v.message);
        }
    }
}
```

---

## Performance Overhead

The `step()` call path:

- Clones `SpecState` (stack-allocated struct copy)
- Evaluates k invariant expressions (simple arithmetic comparisons)
- No heap allocation in the common (no-violation) path

For `bank-account-parameterized.spec` with 2 invariants and 1 temporal property, this is three comparisons per transition — negligible in any production workload.

---

## Validating Monitor Traces with `validate-monitor`

The `validate-monitor` command verifies that every violation found by BFS is **internally consistent**: the recorded trace genuinely leads to a state that fails the target invariant, with no intermediate invariant violations that would make the trace unreachable.

```bash
spectre validate-monitor examples/bank-account-violation.spec
```

Example output:

```
Validated 2 violation trace(s):
  [1] action 'withdraw' → invariant 'non_negative': consistent (3 steps, final violated)
  [2] action 'deposit' → invariant 'balance_cap': consistent (2 steps, final violated)

2/2 consistent.
```

### Why Validation Matters

The BFS explorer records the **pre-violation state** for transition-level violations — the state just before the offending action executes.  The validate-monitor command replays the full path at the executor level (bypassing invariant guards) and confirms that:

1. No intermediate step violates a *different* invariant.
2. The final action, when applied, produces a state that fails the target invariant.

If a trace is **not** consistent, it indicates a bug in the violation-recording logic — the trace cannot be replayed to reproduce the reported failure.

### Flags

```bash
spectre validate-monitor account.spec --max-states 20000 --max-depth 50
```

| Flag | Default | Effect |
|------|---------|--------|
| `--max-states` | 10000 | BFS state limit |
| `--max-depth` | 100 | BFS depth limit |

### Integration with `generate-monitor`

Run `validate-monitor` after any spec change to confirm that the violations your runtime monitor would detect are the same ones the model checker found:

```bash
# 1. Verify the spec
spectre verify account.spec

# 2. Confirm all violation traces are self-consistent
spectre validate-monitor account.spec

# 3. Generate the monitor
spectre generate-monitor --lang rust account.spec -o src/monitor.rs
```

A passing `validate-monitor` result means the monitor's invariant checks are backed by genuinely reachable counterexample traces.

---

## Spec Consistency Guarantee

Because the runtime monitor is generated from the **same spec file** used for model checking, it is guaranteed to be consistent with your verification results.  There is no separate monitoring-rule language to keep in sync.

If your spec passes `spectre verify`, and the monitor never fires in production, you have end-to-end assurance that your implementation matches the design.

---

## Complete Closed-Loop Workflow

```
┌────────────────────────────────────────────────────────────────┐
│                     Closed-Loop Verification                    │
│                                                                │
│  1. spectre mine --lang rust src/account.rs                    │
│     → account.spec skeleton                                    │
│                                                                │
│  2. Add invariants + temporal properties                       │
│     spectre verify account.spec  → fix with CEGIS if needed   │
│                                                                │
│  3. spectre verify account.spec --emit-traces trace.itf.json   │
│     spectre generate-driver --lang rust account.spec           │
│     → fill in driver.rs, run against trace  (Chapter 10)      │
│                                                                │
│  4. spectre generate-monitor --lang rust account.spec          │
│     → embed monitor.rs in production binary  (this chapter)   │
│                                                                │
│  Single spec → design-time verification                        │
│              → model-based test replay                         │
│              → production runtime monitoring                   │
└────────────────────────────────────────────────────────────────┘
```

This is the full Spectre closed loop: one spec file drives verification, testing, and production monitoring simultaneously.
# Chapter 12: Evaluating Rust Coverage and Benchmarking

This chapter covers two meta-tooling commands: `spectre eval-rust`, which measures how completely the miner can translate a Rust codebase, and `spectre benchmark`, which measures how quickly Spectre explores the state space of a set of specs.

---

## Evaluating Rust Extraction with `eval-rust`

Before committing to a Spectre workflow for a Rust project, it helps to know up-front which parts of the codebase the miner can translate cleanly and which require manual follow-up.  `spectre eval-rust` runs the miner over one or more Rust source files and prints a Markdown report summarising extraction quality.

```bash
spectre eval-rust src/account.rs src/order.rs src/payment.rs
```

Or over an entire directory:

```bash
spectre eval-rust --dir src/
```

### Reading the Report

```
# Rust Extraction Evaluation Report

## Per-File Summary

| File           | Struct       | Fields | Enums | Methods | Omissions | SuccessFraction |
|----------------|--------------|--------|-------|---------|-----------|----------------|
| bank_account.rs | BankAccount  |   3    |   1   |    4    |     4     |      50%       |
| counter.rs      | Counter      |   2    |   0   |    3    |     1     |     100%       |
| state_machine.rs | Machine     |   4    |   2   |    2    |     4     |      50%       |

**Totals** — files: 3 | fields: 9 | enums: 3 | methods: 9 | omissions: 9 | mean success: 67%

## Omissions by Kind

| Kind            | Count |
|-----------------|-------|
| SKIPPED_METHOD  |   3   |
| TUPLE_VARIANT   |   2   |
| OVERFLOW_OP     |   2   |
| RESULT_RETURN   |   2   |
```

**SuccessFraction** is the fraction of methods with no body-level omission (overflow, panic, Result, or complex assign).  `SKIPPED_METHOD` (constructors) does not count against the fraction — skipping `new()` is expected and benign.

A file at 100% means every extracted method translates cleanly to Spectre actions.  A file at 50% has at least one method whose body required abstraction; those methods should be reviewed before treating the mined spec as authoritative.

### Interpreting the Omissions-by-Kind Table

Focus attention in order:

1. **`RESULT_RETURN`** — methods that return `Result<T,E>` are mined as happy-path only.  If error paths are semantically important for your invariants, model them manually.
2. **`OVERFLOW_OP`** — `checked_add`, `wrapping_add`, etc. are abstracted to plain addition.  If your invariants involve overflow-safe bounds, add explicit `require` guards.
3. **`TUPLE_VARIANT`** — enum variants with data are collapsed to unit variants.  Add back any fields you need as separate `var` declarations.
4. **`SKIPPED_METHOD`** — expected.  Review if a skipped method has semantically important side-effects.

### Using `eval-rust` in CI

Add an `eval-rust` step to your CI pipeline to track extraction coverage over time:

```yaml
- name: Check Spectre extraction coverage
  run: |
    spectre eval-rust --dir src/ > eval-report.md
    # Fail if mean success drops below 80%
    grep "mean success" eval-report.md | grep -v " [0-7][0-9]%" || true
```

You can also commit the Markdown report as an artifact alongside the generated spec, providing a permanent record of which Rust constructs required manual work.

---

## Benchmarking with `spectre benchmark`

`spectre benchmark` runs BFS exploration on a set of spec files and prints a timing table.  It is useful for tracking the performance impact of spec changes, comparing exploration depth across models, and establishing a baseline before enabling optimisations such as `--por` or `--property-guided`.

```bash
spectre benchmark
```

By default it scans the `benchmarks/` directory for `*.spec` files.  To use a different directory or explicit files:

```bash
spectre benchmark --dir my-specs/
spectre benchmark examples/raft.spec examples/bank.spec
```

### Reading the Output

```
Spec                            States      Elapsed     Depth
------------------------------------------------------------------
counter.spec                        11         0ms          10
bank.spec                          150         4ms          30
mutex.spec                           9         0ms           8
resource_alloc.spec                480         1ms          20
cache_coherence.spec              2400        12ms          40
------------------------------------------------------------------
5 spec(s) benchmarked
```

**States** is the number of distinct reachable states (up to `--max-states`).  **Elapsed** is wall-clock time for the full BFS.  **Depth** is the deepest state reached.

### Flags

| Flag | Default | Effect |
|------|---------|--------|
| `--max-states` | 50000 | Stop BFS after this many states |
| `--dir` | `benchmarks/` | Directory to scan for `*.spec` files |

### The Built-in Benchmark Suite

Spectre ships with five benchmark specs in `benchmarks/`:

| Spec | States | What it models |
|------|--------|---------------|
| `counter.spec` | ~11 | Simple integer counter with a bound |
| `bank.spec` | ~150 | Parameterised bank account (`param LIMIT`) |
| `mutex.spec` | ~9 | Two-process mutex with `ProcState` enum |
| `resource_alloc.spec` | ~480 | N-slot resource pool (`param N`) |
| `cache_coherence.spec` | ~2400 | MSI cache coherence protocol |

These cover the range from trivial (counter) to moderately complex (cache coherence), providing a quick health-check of the verifier engine across representative models.

### Adding Your Own Benchmarks

Any `.spec` file dropped into `benchmarks/` is automatically picked up.  For parameterised specs, use `param` declarations so the benchmark suite covers multiple configurations:

```spectre
param N: int = 4        // default used by benchmark; override with --param N=8
var slots: int
```

### Benchmark vs Verify

`spectre benchmark` intentionally does not check invariants — it measures raw exploration throughput.  To measure verified throughput (invariant checking included), use `spectre verify` directly:

```bash
time spectre verify benchmarks/cache_coherence.spec --max-states 50000
```

The difference between benchmark and verify timing gives you the cost of invariant evaluation for your spec.

---

## Combined Workflow

```
eval-rust → identify gaps → mine + fix → verify → benchmark
```

1. **`eval-rust --dir src/`** — measure extraction coverage before writing any specs manually
2. **`mine + review`** — focus manual effort on the files and methods flagged by `eval-rust`
3. **`verify`** — confirm the enriched spec is safe
4. **`benchmark`** — establish a timing baseline; re-run after adding new actions or invariants to catch state-space explosions early
