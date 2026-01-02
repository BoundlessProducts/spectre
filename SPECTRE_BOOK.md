# The Spectre Language Book

A comprehensive guide to writing specifications in Spectre.

---

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Basic Concepts](#basic-concepts)
4. [Types](#types)
5. [State Variables](#state-variables)
6. [Initial States](#initial-states)
7. [Actions and Transitions](#actions-and-transitions)
8. [Pure Functions](#pure-functions)
9. [Invariants](#invariants)
10. [Temporal Properties](#temporal-properties)
11. [Fairness Conditions](#fairness-conditions)
12. [Modules](#modules)
13. [Constants](#constants)
14. [Error Messages and Debugging](#error-messages-and-debugging)
15. [Verification](#verification)
16. [Debugging Failures: Invariants and Temporal Properties](#debugging-failures-invariants-and-temporal-properties)
17. [Advanced Examples](#advanced-examples)

---

## Introduction

### What is Spectre?

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

## Getting Started

### Installation

Install Spectre using your preferred method:

**macOS (Homebrew):**

**Prerequisites**: Go 1.19 or later must be installed (Homebrew will build Spectre from source).

If Go is not installed, install it first:
```bash
brew install go
```

Since the Homebrew formula is in the main repository (not a separate `homebrew-spectre` repo), use the full GitHub URL:

```bash
brew tap akkeshavan/spectre https://github.com/akkeshavan/spectre.git
brew install spectre
```

**Note**: The installation builds from source, so Go is required. Homebrew will automatically install Go as a dependency if it's not already installed, but it's recommended to install Go first to ensure you have the correct version.

**Note**: If you later create a separate `homebrew-spectre` repository, you can simplify to:
```bash
brew tap akkeshavan/spectre
brew install spectre
```

**Linux:**

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

**Windows:**  
```powershell
powershell -ExecutionPolicy Bypass -File scripts/install.ps1
```

### Finding Example Files After Installation

When you install Spectre using Homebrew (macOS) or the installation script (Linux), **example files are automatically included** with the installation and placed in standard system directories.

**macOS (Homebrew Installation):**

The examples are installed to the Homebrew share directory:
- **Location**: `/opt/homebrew/share/spectre/examples/` (Apple Silicon) or `/usr/local/share/spectre/examples/` (Intel)
- **Using Homebrew prefix**: `$(brew --prefix)/share/spectre/examples/`

```bash
# List all available examples:
ls $(brew --prefix)/share/spectre/examples/

# Copy examples to a working directory (recommended):
mkdir -p ~/my-spectre-examples
cp $(brew --prefix)/share/spectre/examples/*.spec ~/my-spectre-examples/
cd ~/my-spectre-examples

# Now test the examples:
spectre parse counter.spec
spectre typecheck counter.spec
spectre verify counter.spec
```

**Linux (Installation Script):**

The examples are installed to the system share directory:
- **Location**: `/usr/local/share/spectre/examples/`

```bash
# List all available examples:
ls /usr/local/share/spectre/examples/

# Copy examples to a working directory (recommended):
mkdir -p ~/my-spectre-examples
cp /usr/local/share/spectre/examples/*.spec ~/my-spectre-examples/
cd ~/my-spectre-examples

# Now test the examples:
spectre parse counter.spec
spectre typecheck counter.spec
spectre verify counter.spec
```

> **⚠️ Important**: The examples are installed in system directories. **Before testing or modifying examples, copy them to a new directory** (e.g., `~/my-spectre-examples/` or `./test-examples/`). This prevents accidental modification of system files and allows you to experiment freely.
>
> **Example:**
> ```bash
> # macOS
> mkdir -p ~/my-spectre-examples
> cp $(brew --prefix)/share/spectre/examples/*.spec ~/my-spectre-examples/
> cd ~/my-spectre-examples
> spectre parse counter.spec
>
> # Linux
> mkdir -p ~/my-spectre-examples
> cp /usr/local/share/spectre/examples/*.spec ~/my-spectre-examples/
> cd ~/my-spectre-examples
> spectre parse counter.spec
> ```

**Note**: 
- If you installed to a custom location using `INSTALL_DIR` or `SHARE_DIR` environment variables, adjust the paths accordingly.
- The examples directory contains all the example `.spec` files from the repository, so you don't need to clone the repository just to access examples.

### Your First Specification

Let's start with a simple counter example:

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
```

Save this as `counter.spec` and verify it:

```bash
spectre parse counter.spec
spectre typecheck counter.spec
spectre verify counter.spec
```

**Or use the installed example files:**

The examples are automatically installed with Spectre and available in the appropriate directories:

```bash
# macOS (Homebrew installation)
spectre parse $(brew --prefix)/share/spectre/examples/counter.spec

# Linux (installation script)
spectre parse /usr/local/share/spectre/examples/counter.spec
```

> **⚠️ Important**: Before testing or modifying examples, **copy them to a new directory** first. The examples are in system directories and should not be modified directly. See the [Finding Example Files After Installation](#finding-example-files-after-installation) section above for instructions on copying examples to a working directory.

### Understanding the Example

- **`var counter: int`**: Declares a state variable named `counter` of type `int`
- **`init { ... }`**: Defines the initial state of the system
- **`action increment { ... }`**: Defines a transition that can change the state
- **`counter' = counter + 1`**: The prime (`'`) denotes the next state value
- **`invariant nonNegative { ... }`**: A property that must always hold

---

## Basic Concepts

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

This is called **prime notation** and comes from TLA+.

### Example: Simple Counter

Let's examine the counter example from `examples/counter.spec`:

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
```

**Key Points:**
- `counter` is the current value
- `counter'` is the value after the action
- `require` adds a precondition (guard) on the action
- The action only executes if the precondition is true

---

## Types

Spectre provides a rich type system with both primitive and compound types.

### Primitive Types

- **`int`**: Integers (e.g., `42`, `-10`, `0`)
- **`bool`**: Booleans (`true`, `false`)
- **`str`**: Strings (e.g., `"hello"`, `"world"`)
- **`float`**: Floating-point numbers (e.g., `3.14`, `-0.5`)

### Compound Types

#### Records

Records group related data together:

```spectre
type User = {
  id: int,
  name: str,
  age: int,
  active: bool
}

var user: User
```

Access fields using dot notation: `user.name`, `user.age`

#### Sets

Sets represent unordered collections of unique elements:

```spectre
var users: Set<User>
var numbers: Set<int>
```

#### Maps

Maps represent key-value pairs:

```spectre
var userMap: Map<int, User>  // Maps user IDs to User records
var scores: Map<str, int>     // Maps names to scores
```

#### Lists

Lists represent ordered sequences:

```spectre
var items: List<int>
var names: List<str>
```

#### Options

Options represent values that may or may not exist:

```spectre
var maybeUser: Option<User>
```

#### Enums

Enums represent a fixed set of values:

```spectre
enum ProcessState {
  Idle,
  Waiting,
  Critical
}

var state: ProcessState
```

### Type Examples

See `examples/pure-functions.spec` for examples of record types:

```spectre
type User = {
  id: int,
  name: str,
  age: int,
  active: bool
}

var users: Set<User>
```

---

## State Variables

State variables represent the data that changes over time in your system.

### Declaration Syntax

```spectre
description "Optional description for better error messages"
var variableName: Type
```

### Examples

**Simple counter:**
```spectre
description "Tracks a numeric counter value"
var counter: int
```

**Multiple variables:**
```spectre
description "State of the first process"
var process1: ProcessState

description "State of the second process"
var process2: ProcessState

description "Lock flag indicating if a process is in critical section"
var lock: bool
```

**Collections:**
```spectre
description "Collection of all users in the system"
var users: Set<User>

description "Counter tracking number of operations"
var counter: int
```

### Best Practices

1. **Use descriptive names**: `counter` is better than `c`
2. **Add descriptions**: They appear in error messages
3. **Group related variables**: Keep related state together
4. **Choose appropriate types**: Use `Set` for unique collections, `List` for ordered sequences

---

## Initial States

The initial state defines how your system starts.

### Deterministic Initial State

Use `init` for a single, deterministic starting state:

```spectre
description "System starts with counter initialized to zero"
init {
  counter = 0
}
```

### Multiple Initial States with `oneOf`

Use `oneOf` when your system can start from multiple configurations:

```spectre
description "System can start from multiple configurations"
init oneOf {
  {
    counter = 0
    mode = "start"
    initialized = false
  },
  {
    counter = 10
    mode = "resume"
    initialized = true
  },
  {
    counter = 20
    mode = "restart"
    initialized = true
  }
}
```

See `examples/oneof-example.spec` for a complete example.

### Why Use `oneOf`?

`oneOf` is useful when:
- Your system can start in different modes
- You want to verify properties hold for all possible starting states
- You're modeling non-deterministic initialization

**Example:** A system that can start fresh, resume from a checkpoint, or restart from a saved state.

---

## Actions and Transitions

Actions define how the system can transition from one state to another.

### Basic Action Syntax

```spectre
description "Optional description"
action actionName {
  variable' = expression
}
```

### Prime Notation

The prime (`'`) denotes the next state:

```spectre
action increment {
  counter' = counter + 1  // Next state = current state + 1
}
```

### Actions with Parameters

Actions can take parameters:

```spectre
description "Changes the system mode to a new value"
action setMode(newMode: str) {
  require newMode = "start" || newMode = "resume" || newMode = "restart"
  counter' = counter
  mode' = newMode
  initialized' = initialized
}
```

### Preconditions with `require`

Use `require` to add guards (preconditions) to actions:

```spectre
description "Decrements the counter by one, only when counter is positive"
action decrement {
  require counter > 0
  counter' = counter - 1
}
```

The action only executes if `counter > 0` is true.

### Action Guards with `when`

You can also use `when` in the action declaration:

```spectre
action decrement when counter > 0 {
  counter' = counter - 1
}
```

### Postconditions with `ensure`

Use `ensure` to specify what must be true after the action:

```spectre
action increment {
  counter' = counter + 1
  ensure counter' > counter  // Postcondition: counter must increase
}
```

### Multiple Assignments

Actions can update multiple variables:

```spectre
description "Process 1 requests and acquires the lock"
action process1Request {
  require process1 = ProcessState.Idle && !lock
  process1' = ProcessState.Critical
  lock' = true
}
```

### Complete Example

From `examples/counter.spec`:

```spectre
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
```

---

## Pure Functions

Pure functions are computational helpers that don't modify state. They're useful for reusable logic.

### Function Syntax

```spectre
description "Optional description"
fun functionName(param1: Type1, param2: Type2): ReturnType {
  return expression
}
```

### Simple Functions

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

### Recursive Functions

Functions can be recursive:

```spectre
description "Calculates the factorial of a number recursively"
fun factorial(n: int): int {
  if (n <= 1) {
    return 1
  } else {
    return n * factorial(n - 1)
  }
}
```

### Functions with Collections

Functions can work with collections:

```spectre
description "Counts the number of active users in a set"
fun countActiveUsers(userSet: Set<User>): int {
  return userSet.filter(u => u.active).size()
}

description "Calculates the average age of users in a set"
fun averageAge(userSet: Set<User>): float {
  if (userSet.size() = 0) {
    return 0.0
  } else {
    let totalAge = userSet.map(u => u.age).reduce(0, (acc, age) => acc + age)
    return totalAge / userSet.size()
  }
}
```

### Using Functions in Actions

Pure functions can be called from actions:

```spectre
action addUser(id: int, name: str, age: int) {
  require isValidUserId(id)
  require isValidUserName(name)
  
  let newUser = { id: id, name: name, age: age, active: true }
  users' = users.union(Set.of(newUser))
  counter' = add(counter, 1)  // Using pure function
}
```

### Purity Rules

Pure functions must:
- ✅ Only read their parameters
- ✅ Not access state variables
- ✅ Not modify state variables
- ✅ Return a value based only on inputs

See `examples/pure-functions.spec` for comprehensive examples.

---

## Invariants

Invariants are properties that must **always** hold in every reachable state.

### Invariant Syntax

```spectre
description "Optional description"
invariant invariantName {
  condition
}
```

### Simple Invariants

```spectre
description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}

description "Keeps counter within reasonable bounds"
invariant bounded {
  counter <= 100
}
```

### Complex Invariants

Invariants can use logical operators:

```spectre
description "CRITICAL: Ensures only one process can be in critical section at a time"
invariant mutualExclusion {
  !(process1 = ProcessState.Critical && process2 = ProcessState.Critical)
}

description "Ensures lock flag accurately reflects critical section state"
invariant lockConsistency {
  (process1 = ProcessState.Critical || process2 = ProcessState.Critical) = lock
}
```

### Invariants with Collections

```spectre
description "Validates that all users have valid IDs and names"
invariant allUsersValid {
  users.forall(u => isValidUserId(u.id) && isValidUserName(u.name))
}

description "Ensures eligible users are always active"
invariant eligibleUsersActive {
  users.forall(u => isEligible(u) → u.active)
}
```

### When Invariants Are Checked

Invariants are checked:
- After the initial state is set
- After every action execution
- During state space exploration

If an invariant fails, the verifier reports:
- Which invariant failed
- The state that violated it
- The execution trace leading to the violation

### Example from `examples/counter.spec`

```spectre
description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}

description "Keeps counter within reasonable bounds"
invariant bounded {
  counter <= 100
}
```

---

## Temporal Properties

Temporal properties describe how the system behaves **over time**. Unlike invariants (which must always hold), temporal properties can express liveness conditions like "eventually something happens."

### Temporal Operators

#### `always`

The `always` operator means "at every point in the execution":

```spectre
description "Ensures counter remains non-negative throughout execution"
temporal alwaysNonNegative {
  always (counter >= 0)
}
```

This is similar to an invariant but expressed as a temporal property.

#### `eventually`

The `eventually` operator means "at some point in the future":

```spectre
description "Verifies that counter eventually reaches value 10"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}
```

#### `until`

The `until` operator means "condition1 holds until condition2 becomes true":

```spectre
temporal safeUntilGoal {
  counter < 100 until counter = 100
}
```

#### `leads-to` (→)

The `leads-to` operator means "if condition1 becomes true, then condition2 will eventually become true":

```spectre
description "Guarantees progress: if counter is below 10, it will eventually reach 10"
temporal progress {
  always (counter < 10 → eventually counter = 10)
}
```

### Combining Temporal Operators

You can combine temporal operators:

```spectre
description "Guarantees progress: if counter is below 10, it will eventually reach 10"
temporal progress {
  always (counter < 10 → eventually counter = 10)
}
```

### Examples from `examples/counter.spec`

```spectre
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

### When Temporal Properties Are Checked

Temporal properties are evaluated over **execution traces** (sequences of states). The verifier explores the state space and checks if the temporal property holds for all possible execution paths.

---

## Fairness Conditions

Fairness conditions ensure that actions execute "fairly" in concurrent systems. They're essential for verifying liveness properties.

### Weak Fairness (WF)

**Weak Fairness** means: "If an action is **continuously enabled**, it will eventually execute."

```spectre
description "Weak fairness ensures process1 gets fair access when continuously waiting"
temporal process1WeakFairness {
  WF(process1Request)
}
```

**Meaning**: If `process1Request` is enabled and stays enabled, it will eventually execute.

### Strong Fairness (SF)

**Strong Fairness** means: "If an action is enabled **infinitely often**, it will eventually execute."

```spectre
description "Strong fairness ensures process1 executes even if intermittently enabled"
temporal process1StrongFairness {
  SF(process1Request)
}
```

**Meaning**: Even if `process1Request` is enabled and disabled repeatedly, it will eventually execute if it's enabled infinitely often.

### When to Use Weak vs Strong Fairness

- **Weak Fairness (WF)**: Use when an action should execute if it's continuously available
  - Example: A process waiting for a lock should eventually get it if the lock is continuously free

- **Strong Fairness (SF)**: Use when an action should execute even if it's intermittently enabled
  - Example: A message queue processor should eventually process messages even if the queue is sometimes empty

### Fairness Examples

From `examples/fairness-example.spec`:

```spectre
description "Weak fairness ensures process1 gets fair access when continuously waiting"
temporal process1WeakFairness {
  WF(process1Request)
}

description "Strong fairness ensures process1 executes even if intermittently enabled"
temporal process1StrongFairness {
  SF(process1Request)
}

description "With weak fairness, process1 will eventually enter critical section"
temporal eventuallyProcess1Critical {
  WF(process1Request) → eventually (process1 = ProcessState.Critical)
}
```

### Mutex Example

See `examples/mutex.spec` for a complete mutex implementation with fairness:

```spectre
description "Fairness guarantee: if process1 is idle and lock is free, it will eventually get the lock"
temporal fairnessProcess1 {
  always (process1 = ProcessState.Idle && !lock → eventually process1 = ProcessState.Critical)
}
```

---

## Modules

Modules help organize your specifications into reusable components.

### Module Syntax

```spectre
module ModuleName {
  // Declarations: vars, actions, invariants, etc.
}
```

### Basic Module

```spectre
module Counter {
  description "Tracks a numeric counter value"
  var counter: int

  description "System starts with counter initialized to zero"
  init {
    counter = 0
  }

  description "Increments the counter by one"
  public action increment {
    counter' = counter + 1
  }

  description "Ensures counter never becomes negative"
  public invariant nonNegative {
    counter >= 0
  }
}
```

### Importing Modules

```spectre
import Counter

// Now you can use Counter's public members
```

### Module Extension

Modules can extend other modules:

```spectre
module BoundedCounter extends Counter {
  description "Maximum allowed counter value"
  const MAX_VALUE: int = 100

  description "Increments counter but enforces maximum bound"
  public action increment {
    require counter < MAX_VALUE
    super.increment()  // Call parent's increment
  }

  description "Ensures counter stays within bounds"
  public invariant bounded {
    counter <= MAX_VALUE
  }
}
```

### Visibility: `public` and `private`

- **`public`**: Can be accessed from outside the module
- **`private`**: Only accessible within the module (default)

```spectre
module Counter {
  public action increment { ... }      // Can be called from outside
  private action internalHelper { ... } // Only within module
}
```

### Super Calls

Use `super` to call parent module methods:

```spectre
module BoundedCounter extends Counter {
  public action increment {
    require counter < MAX_VALUE
    super.increment()  // Calls Counter.increment()
  }
}
```

### Complete Example

See `examples/modules-example.spec` for a complete module example:

```spectre
// Base counter module
module Counter {
  var counter: int
  init { counter = 0 }
  public action increment { counter' = counter + 1 }
  public invariant nonNegative { counter >= 0 }
}

// Bounded counter extends base counter
module BoundedCounter extends Counter {
  const MAX_VALUE: int = 100
  public action increment {
    require counter < MAX_VALUE
    super.increment()
  }
  public invariant bounded { counter <= MAX_VALUE }
}
```

---

## Constants

Constants are fixed values that don't change during execution. They're useful for parameterizing specifications.

### Constant Syntax

```spectre
description "Optional description"
const CONSTANT_NAME: Type = value
```

### Simple Constants

```spectre
description "Number of processes in the system"
const NUM_PROCESSES: int = 2

description "Maximum allowed counter value"
const MAX_VALUE: int = 100

description "Timeout duration in seconds"
const TIMEOUT: int = 30
```

### Using Constants

Constants can be used in actions, invariants, and temporal properties:

```spectre
const MAX_USERS: int = 100

action addUser {
  require users.size() < MAX_USERS
  // ...
}

invariant userLimit {
  users.size() <= MAX_USERS
}
```

### Constants in Modules

Constants can be defined in modules:

```spectre
module BoundedCounter extends Counter {
  const MAX_VALUE: int = 100
  
  public action increment {
    require counter < MAX_VALUE
    super.increment()
  }
}
```

### Example from `examples/fairness-example.spec`

```spectre
description "Number of processes in the system"
const NUM_PROCESSES: int = 2
```

---

## Error Messages and Debugging

Spectre provides excellent error messages with descriptions and stack traces to help you debug your specifications.

### Descriptions

Add `description` fields to all language elements for better error messages:

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

### Error Message Format

When an invariant fails, you get:

```
21:2: Invariant 'nonNegative' violated: (Ensures counter never becomes negative) condition evaluated to false
```

The error includes:
- **Position**: `21:2` (line and column)
- **Element name**: `nonNegative`
- **Description**: `(Ensures counter never becomes negative)`
- **Error details**: `condition evaluated to false`

### Stack Traces

When a violation occurs, Spectre provides a stack trace showing the execution path:

```
Stack trace:
  1. 5:3 initial state: System starts with counter initialized to zero
  2. 13:2 action 'increment' (Increments the counter by one): State 1
  3. 18:2 action 'decrement' (Decrements the counter by one): State 2
  4. 29:2 invariant 'nonNegative' violated: (Ensures counter never becomes negative) counter < 0
```

### Example: Error Trace

See `examples/error-trace-example.spec` for an example that demonstrates error traces:

```spectre
description "CRITICAL: Ensures only one process can be in critical section at a time"
invariant mutualExclusion {
  !(process1 = ProcessState.Critical && process2 = ProcessState.Critical)
}
```

If this invariant fails, you'll see a trace showing how both processes ended up in the critical section.

### Best Practices for Debugging

1. **Add descriptions**: They make error messages much more readable
2. **Use meaningful names**: `mutualExclusion` is better than `inv1`
3. **Check preconditions**: Ensure `require` conditions are correct
4. **Verify initial states**: Make sure `init` sets up valid states
5. **Review execution traces**: The stack trace shows exactly what happened

---

## Verification

The Spectre verifier checks your specification for correctness.

### What Gets Verified

1. **Syntax**: Parse errors are caught first
2. **Types**: Type errors are reported with positions
3. **Invariants**: Checked at every state
4. **Temporal Properties**: Evaluated over execution traces
5. **Fairness**: Verified for concurrent systems

### Running Verification

```bash
# Parse only (check syntax)
spectre parse examples/counter.spec

# Type check (verify types)
spectre typecheck examples/counter.spec

# Full verification (check invariants and temporal properties)
spectre verify examples/counter.spec
```

### Verification Output

**Success:**
```
✓ Verification passed for examples/counter.spec
  Explored 10 states
```

**Failure:**
```
Verification failed: 1 violation(s) found

Violation 1:
  Invariant 'nonNegative' violated: counter < 0
  Path:
    1. increment
    2. decrement
    3. decrement
```

### State Space Exploration

The verifier explores the state space using BFS (Breadth-First Search) or DFS (Depth-First Search):

- **BFS**: Explores all states at depth 1, then depth 2, etc.
- **DFS**: Explores one path deeply before backtracking

The verifier:
- Tracks visited states (using hashing)
- Detects cycles
- Checks invariants at each state
- Evaluates temporal properties over traces
- Generates counterexamples for violations

### Verification Limits

To prevent infinite exploration:
- **Max depth**: Limits how deep to explore
- **Max states**: Limits total number of states to explore

These can be configured in the verifier.

---

## Advanced Examples

### Example 1: Mutex Lock

The mutex example (`examples/mutex.spec`) demonstrates:
- Multiple processes competing for a shared resource
- Mutual exclusion invariant
- Fairness properties
- Process state management

```spectre
enum ProcessState {
  Idle,
  Waiting,
  Critical
}

var process1: ProcessState
var process2: ProcessState
var lock: bool

init {
  process1 = ProcessState.Idle
  process2 = ProcessState.Idle
  lock = false
}

action process1Request {
  require process1 = ProcessState.Idle && !lock
  process1' = ProcessState.Critical
  lock' = true
}

invariant mutualExclusion {
  !(process1 = ProcessState.Critical && process2 = ProcessState.Critical)
}
```

### Example 2: Multiple Initial States

The `oneOf` example (`examples/oneof-example.spec`) shows:
- Non-deterministic initialization
- Multiple starting configurations
- Verification across all initial states

```spectre
init oneOf {
  { counter = 0, mode = "start", initialized = false },
  { counter = 10, mode = "resume", initialized = true },
  { counter = 20, mode = "restart", initialized = true }
}
```

### Example 3: Pure Functions

The pure functions example (`examples/pure-functions.spec`) demonstrates:
- Reusable computational helpers
- Recursive functions
- Functions working with collections
- Using functions in actions

```spectre
fun factorial(n: int): int {
  if (n <= 1) {
    return 1
  } else {
    return n * factorial(n - 1)
  }
}

fun countActiveUsers(userSet: Set<User>): int {
  return userSet.filter(u => u.active).size()
}
```

### Example 4: Fairness

The fairness example (`examples/fairness-example.spec`) shows:
- Weak Fairness (WF)
- Strong Fairness (SF)
- Fairness guarantees for concurrent systems

```spectre
temporal process1WeakFairness {
  WF(process1Request)
}

temporal process1StrongFairness {
  SF(process1Request)
}
```

### Example 5: Modules

The modules example (`examples/modules-example.spec`) demonstrates:
- Module organization
- Module extension
- Super calls
- Public/private visibility

```spectre
module Counter {
  var counter: int
  public action increment { counter' = counter + 1 }
}

module BoundedCounter extends Counter {
  const MAX_VALUE: int = 100
  public action increment {
    require counter < MAX_VALUE
    super.increment()
  }
}
```

---

## Common Patterns

### Pattern 1: Bounded Counter

A counter that stays within bounds:

```spectre
var counter: int
const MAX: int = 100

init { counter = 0 }

action increment {
  require counter < MAX
  counter' = counter + 1
}

invariant bounded { counter >= 0 && counter <= MAX }
```

### Pattern 2: Resource Management

Managing a shared resource:

```spectre
var available: int
const TOTAL: int = 10

init { available = TOTAL }

action acquire {
  require available > 0
  available' = available - 1
}

action release {
  require available < TOTAL
  available' = available + 1
}

invariant resourceConservation {
  available >= 0 && available <= TOTAL
}
```

### Pattern 3: State Machine

Modeling a state machine:

```spectre
enum State {
  Idle,
  Running,
  Stopped
}

var currentState: State

init { currentState = State.Idle }

action start {
  require currentState = State.Idle
  currentState' = State.Running
}

action stop {
  require currentState = State.Running
  currentState' = State.Stopped
}

invariant validState {
  currentState = State.Idle || 
  currentState = State.Running || 
  currentState = State.Stopped
}
```

---

## Tips and Best Practices

### 1. Start Simple

Begin with a simple specification and gradually add complexity:
1. Define state variables
2. Set initial state
3. Add one action
4. Add an invariant
5. Verify and iterate

### 2. Use Descriptions

Always add descriptions to improve error messages:

```spectre
description "Tracks the number of active sessions"
var sessionCount: int
```

### 3. Name Things Clearly

Use descriptive names:
- ✅ `sessionCount` not `sc`
- ✅ `processState` not `ps`
- ✅ `mutualExclusion` not `inv1`

### 4. Verify Incrementally

Verify after each change:
```bash
spectre parse yourfile.spec
spectre typecheck yourfile.spec
spectre verify yourfile.spec
```

### 5. Use Preconditions

Add `require` statements to prevent invalid states:

```spectre
action decrement {
  require counter > 0  // Prevents negative counter
  counter' = counter - 1
}
```

### 6. Test Edge Cases

Verify your specification handles:
- Empty collections
- Boundary values
- All initial states (if using `oneOf`)
- Concurrent actions

### 7. Organize with Modules

Use modules to organize large specifications:

```spectre
module UserManagement {
  // User-related state and actions
}

module Authentication {
  // Auth-related state and actions
}
```

---

## Debugging Failures: Invariants and Temporal Properties

This section shows real examples of invariant and temporal property failures, their error messages, and how to fix them.

### Invariant Failures

#### Example 1: Missing Precondition

**Problem**: An action can create an invalid state.

**Broken Specification:**

```spectre
var counter: int

init {
  counter = 0
}

action decrement {
  counter' = counter - 1  // ❌ No precondition!
}

invariant nonNegative {
  counter >= 0  // This will fail!
}
```

**Error Message:**

```
Verification failed: 1 violation(s) found

Violation 1:
  Invariant 'nonNegative' violated: (Ensures counter never becomes negative)
  Condition: counter >= 0 evaluated to false
  State: counter = -1
  
  Execution Trace:
    Step 0: Initial state: "System starts with counter initialized to zero"
      counter = 0
    
    Step 1: Action: "Decrements the counter by one" (decrement)
      counter = -1
      ❌ Invariant violated here
```

**Fix**: Add a precondition to prevent invalid states:

```spectre
action decrement {
  require counter > 0  // ✅ Precondition prevents negative counter
  counter' = counter - 1
}
```

**Key Lesson**: Always add `require` statements to prevent actions from creating invalid states.

---

#### Example 2: Incorrect Invariant Logic

**Problem**: The invariant doesn't match the intended property.

**Broken Specification:**

```spectre
var balance: int

init {
  balance = 100
}

action withdraw(amount: int) {
  require amount > 0
  balance' = balance - amount
}

invariant sufficientBalance {
  balance > 0  // ❌ Should be >= 0, not > 0
}
```

**Error Message:**

```
Violation 1:
  Invariant 'sufficientBalance' violated: (Ensures balance stays positive)
  Condition: balance > 0 evaluated to false
  State: balance = 0
  
  Execution Trace:
    Step 0: Initial state
      balance = 100
    Step 1: Action: "Withdraws money" (withdraw(100))
      balance = 0
      ❌ Invariant violated here
```

**Fix**: Correct the invariant logic:

```spectre
invariant sufficientBalance {
  balance >= 0  // ✅ Allow zero balance
}
```

**Key Lesson**: Ensure invariants match your actual requirements. Zero might be a valid state.

---

#### Example 3: Missing State Update

**Problem**: An action doesn't update all relevant state variables.

**Broken Specification:**

```spectre
var counter: int
var initialized: bool

init {
  counter = 0
  initialized = false
}

action increment {
  counter' = counter + 1
  // ❌ Forgot to update initialized!
}

invariant initializationConsistency {
  initialized → counter >= 0  // If initialized, counter should be non-negative
}
```

**Error Message:**

```
Violation 1:
  Invariant 'initializationConsistency' violated
  Condition: initialized → counter >= 0 evaluated to false
  State: counter = 1, initialized = false
  
  Note: This might seem odd, but the invariant can fail if initialized becomes true
  through another action without updating counter properly.
```

**Fix**: Update all relevant state variables:

```spectre
action increment {
  counter' = counter + 1
  initialized' = true  // ✅ Update initialized flag
}
```

**Key Lesson**: When an action changes system state, update all related state variables.

---

#### Example 4: Race Condition in Concurrent System

**Problem**: Multiple processes can violate mutual exclusion.

**Broken Specification:**

```spectre
enum ProcessState {
  Idle,
  Critical
}

var process1: ProcessState
var process2: ProcessState
var lock: bool

init {
  process1 = ProcessState.Idle
  process2 = ProcessState.Idle
  lock = false
}

action process1Request {
  require process1 = ProcessState.Idle && !lock
  process1' = ProcessState.Critical
  lock' = true
}

action process2Request {
  require process2 = ProcessState.Idle && !lock
  process2' = ProcessState.Critical
  lock' = true
}

invariant mutualExclusion {
  !(process1 = ProcessState.Critical && process2 = ProcessState.Critical)
}
```

**Error Message:**

```
Violation 1:
  Invariant 'mutualExclusion' violated: (CRITICAL: Ensures only one process can be in critical section)
  Condition: !(process1 = ProcessState.Critical && process2 = ProcessState.Critical) evaluated to false
  State: process1 = Critical, process2 = Critical, lock = true
  
  Execution Trace:
    Step 0: Initial state
      process1 = Idle, process2 = Idle, lock = false
    Step 1: Action: "Process 1 requests lock" (process1Request)
      process1 = Critical, process2 = Idle, lock = true
    Step 2: Action: "Process 2 requests lock" (process2Request)
      process1 = Critical, process2 = Critical, lock = true
      ❌ Invariant violated here
```

**Analysis**: The problem is that both actions check `!lock` and then set `lock = true`, but they can both execute if they check the condition simultaneously. However, in Spectre, actions execute atomically, so this shouldn't happen unless there's a logic error.

**Fix**: The issue is that the precondition doesn't prevent both processes from entering. We need to ensure atomicity:

```spectre
action process1Request {
  require process1 = ProcessState.Idle && !lock && process2 != ProcessState.Critical
  process1' = ProcessState.Critical
  lock' = true
}

action process2Request {
  require process2 = ProcessState.Idle && !lock && process1 != ProcessState.Critical
  process2' = ProcessState.Critical
  lock' = true
}
```

**Key Lesson**: In concurrent systems, ensure preconditions check all relevant state to prevent race conditions.

---

### Temporal Property Failures

#### Example 1: Unreachable Goal

**Problem**: The temporal property requires something that can never happen.

**Broken Specification:**

```spectre
var counter: int

init {
  counter = 0
}

action increment {
  require counter < 10
  counter' = counter + 1
}

temporal eventuallyReachesTwenty {
  eventually (counter = 20)  // ❌ Counter can never reach 20!
}
```

**Error Message:**

```
Verification failed: Temporal property violation

Violation 1:
  Temporal property 'eventuallyReachesTwenty' violated: (Verifies that counter eventually reaches value 20)
  Property: eventually (counter = 20)
  
  Analysis: The counter is bounded by the precondition (counter < 10) and can only increment.
  Maximum reachable value is 9. The property requires counter = 20, which is unreachable.
  
  Execution Trace:
    Step 0: counter = 0
    Step 1: increment → counter = 1
    Step 2: increment → counter = 2
    ...
    Step 9: increment → counter = 9
    Step 10: increment → ❌ Precondition fails (counter < 10)
    No further progress possible.
```

**Fix**: Either remove the bound or adjust the property:

**Option 1: Remove the bound**

```spectre
action increment {
  counter' = counter + 1  // ✅ No upper bound
}

temporal eventuallyReachesTwenty {
  eventually (counter = 20)  // ✅ Now achievable
}
```

**Option 2: Adjust the property**

```spectre
temporal eventuallyReachesHigh {
  eventually (counter >= 9)  // ✅ Achievable within bounds
}
```

**Key Lesson**: Ensure temporal properties are actually achievable given the system's constraints.

---

#### Example 2: Missing Fairness Condition

**Problem**: A temporal property requires fairness but doesn't specify it.

**Broken Specification:**

```spectre
enum ProcessState {
  Idle,
  Critical
}

var process1: ProcessState
var lock: bool

init {
  process1 = ProcessState.Idle
  lock = false
}

action process1Request {
  require process1 = ProcessState.Idle && !lock
  process1' = ProcessState.Critical
  lock' = true
}

temporal eventuallyCritical {
  eventually (process1 = ProcessState.Critical)  // ❌ May never happen without fairness
}
```

**Error Message:**

```
Verification failed: Temporal property violation

Violation 1:
  Temporal property 'eventuallyCritical' violated: (Verifies that process1 eventually enters critical section)
  Property: eventually (process1 = ProcessState.Critical)
  
  Analysis: The action 'process1Request' is enabled when process1 = Idle && !lock = true.
  However, without fairness, the system could remain in a state where the action is enabled
  but never executes. The property requires eventual execution, which needs fairness.
  
  Counterexample Trace:
    Step 0: process1 = Idle, lock = false
    Step 1: (no action taken, system stutters)
    Step 2: (no action taken, system stutters)
    ... (infinite stuttering)
    ❌ Property never satisfied
```

**Fix**: Add a fairness condition:

```spectre
temporal eventuallyCritical {
  WF(process1Request) → eventually (process1 = ProcessState.Critical)
}

// Or use weak fairness directly
temporal process1Fairness {
  WF(process1Request)
}
```

**Key Lesson**: Liveness properties (like `eventually`) often require fairness conditions to guarantee progress.

---

#### Example 3: Incorrect Leads-To Condition

**Problem**: The `leads-to` property has incorrect logic.

**Broken Specification:**

```spectre
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action reset {
  counter' = 0
}

temporal progress {
  always (counter < 5 → eventually counter = 10)  // ❌ Logic error!
}
```

**Error Message:**

```
Violation 1:
  Temporal property 'progress' violated: (Guarantees progress)
  Property: always (counter < 5 → eventually counter = 10)
  
  Analysis: The property states "if counter < 5, then eventually counter = 10".
  However, if counter is reset to 0, it can stay below 5 indefinitely.
  The property requires reaching 10, but reset prevents this.
  
  Counterexample Trace:
    Step 0: counter = 0
    Step 1: increment → counter = 1
    Step 2: increment → counter = 2
    Step 3: reset → counter = 0
    Step 4: increment → counter = 1
    ... (cycle repeats, never reaches 10)
    ❌ Property violated
```

**Fix**: Correct the property logic:

**Option 1: Require no resets**

```spectre
temporal progress {
  always (counter < 5 && counter' >= counter → eventually counter = 10)
}
```

**Option 2: Adjust the goal**

```spectre
temporal progress {
  always (counter < 5 → eventually counter >= 5)  // ✅ More achievable
}
```

**Key Lesson**: Ensure `leads-to` properties account for all possible system behaviors, including resets and cycles.

---

#### Example 4: Always Property That Can Be Violated

**Problem**: An `always` property can be violated by an action.

**Broken Specification:**

```spectre
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action decrement {
  counter' = counter - 1  // ❌ Can make counter negative
}

temporal alwaysNonNegative {
  always (counter >= 0)  // ❌ Will be violated
}
```

**Error Message:**

```
Violation 1:
  Temporal property 'alwaysNonNegative' violated: (Ensures counter remains non-negative)
  Property: always (counter >= 0)
  
  Execution Trace:
    Step 0: counter = 0
    Step 1: decrement → counter = -1
    ❌ Property violated
```

**Fix**: Add a precondition to prevent violation:

```spectre
action decrement {
  require counter > 0  // ✅ Prevent negative values
  counter' = counter - 1
}

temporal alwaysNonNegative {
  always (counter >= 0)  // ✅ Now holds
}
```

**Key Lesson**: `always` properties must be protected by preconditions in actions.

---

### Summary: Common Failure Patterns

#### Invariant Failures

1. **Missing Preconditions**: Add `require` statements to prevent invalid states
2. **Incorrect Logic**: Review invariant conditions to match requirements
3. **Missing Updates**: Ensure all related state variables are updated
4. **Race Conditions**: Check all relevant state in preconditions

#### Temporal Property Failures

1. **Unreachable Goals**: Ensure properties are achievable given constraints
2. **Missing Fairness**: Add `WF()` or `SF()` for liveness properties
3. **Incorrect Logic**: Review `leads-to` and `until` conditions
4. **Unprotected Always**: Protect `always` properties with preconditions

### Debugging Workflow

When a property fails:

1. **Read the error message**: It shows the violating state and trace
2. **Examine the trace**: See what actions led to the violation
3. **Identify the root cause**: Missing precondition? Incorrect logic? Missing fairness?
4. **Fix systematically**: Add preconditions, correct logic, or add fairness
5. **Re-verify**: Run `spectre verify` again to confirm the fix

---

## Troubleshooting

### Common Issues

#### Issue: "Parse error: unexpected token"

**Solution**: Check syntax, ensure keywords are spelled correctly, verify brackets match.

#### Issue: "Type error: cannot assign int to string"

**Solution**: Check variable types match. Use type annotations if needed.

#### Issue: "Invariant violated"

**Solution**: 
1. Check the execution trace to see what happened
2. Verify your invariants are correct
3. Check if preconditions (`require`) are too weak
4. See [Debugging Failures: Invariants](#debugging-failures-invariants-and-temporal-properties) section above

#### Issue: "Temporal property does not hold"

**Solution**:
1. Review the temporal property logic
2. Check if fairness conditions are needed
3. Verify the property is actually achievable
4. See [Debugging Failures: Temporal Properties](#debugging-failures-invariants-and-temporal-properties) section above

### Getting Help

1. **Check error messages**: They include positions and descriptions
2. **Review execution traces**: They show the path to the error
3. **Simplify**: Try a smaller example first
4. **Verify syntax**: Use `spectre parse` to check syntax first

---

## Conclusion

Spectre provides a powerful yet accessible way to write formal specifications. With its familiar syntax, strong typing, and excellent error messages, you can:

- Model complex systems as state machines
- Verify safety properties with invariants
- Verify liveness properties with temporal logic
- Express fairness conditions for concurrent systems
- Organize code with modules
- Debug easily with descriptions and stack traces

### Next Steps

1. **Try the examples**: Run `spectre verify` on the example files
2. **Write your own**: Start with a simple counter and expand
3. **Read the spec**: See `SPEC.md` for complete language details
4. **Explore advanced features**: Try modules, fairness, and complex types

### Resources

- **Language Specification**: `SPEC.md`
- **CLI Usage**: `USAGE.md`
- **Installation Guide**: `INSTALL.md`
- **Development Guide**: `README_DEV.md`

Happy specifying! 🎉

---

*Last Updated: December 2024*

