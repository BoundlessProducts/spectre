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
18. [Chapter 14: Incremental Verification and Drift Detection](#chapter-14-incremental-verification-and-drift-detection)
19. [Chapter 15: Coverage-Guided Model-Based Testing](#chapter-15-coverage-guided-model-based-testing)

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

Build from source (requires **Go ≥ 1.24**, **Z3 ≥ 4.8** for `spectre sync`, **Rust stable** for `spectre mine --lang rust`):

**macOS:**
```bash
brew install go z3
git clone https://github.com/BoundlessProducts/spectre.git && cd spectre
go build -o spectre ./cmd/spectre
# Optional: build the Rust AST miner
cargo build --release --manifest-path rust/spectre-mine-rs/Cargo.toml
cp rust/spectre-mine-rs/target/release/spectre-mine-rs .
```

**Linux (Debian/Ubuntu):**
```bash
sudo apt install -y golang z3
git clone https://github.com/BoundlessProducts/spectre.git && cd spectre
go build -o spectre ./cmd/spectre
```

**Docker (no local toolchain required):**
```bash
docker build -t spectre-vmcai .
docker run --rm -it spectre-vmcai sh
```

All examples are in the `examples/` directory of the cloned repository. Full installation instructions for all platforms are in [Chapter 1](../spectre_book/01-getting-started.md) and the [README](../README.md).

---

##  Understanding Temporal Property Violations

Before diving into the language details, let's examine a real example that demonstrates how Spectre finds and reports temporal property violations. This will help you understand what temporal verification does and how to fix common issues.

### The Counter Example

The `counter.spec` file demonstrates a common pattern in specifications: a counter that can be incremented, decremented, or reset. Here's the specification:

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

### Running Verification

When you run verification on this file:

```bash
spectre verify counter.spec --verbose
```

The verifier explores the state space and finds that the `progress` temporal property is **violated**.

### Understanding the Error

**What the property means:**

The `progress` property states:
```spectre
always (counter < 10 → eventually counter = 10)
```

This means: "In every state, if counter is less than 10, then eventually counter must become 10."

**Why it fails:**

The property requires that whenever `counter < 10`, there must be a path where `counter` eventually reaches 10. However, the system allows **infinite paths** where this never happens:

1. **Infinite reset loop**: Starting from `counter = 4`, the system can execute:
   - `reset` → `counter = 0`
   - `reset` → `counter = 0` (can repeat forever)
   - On this path, counter never reaches 10

2. **Oscillation loop**: The system can execute:
   - `increment` → `counter = 1`
   - `decrement` → `counter = 0`
   - `increment` → `counter = 1`
   - `decrement` → `counter = 0`
   - (repeats forever, counter stays below 10)

Since the `always` operator requires the property to hold on **all possible execution paths**, and these infinite paths violate the property, verification fails.

**Verification Output:**

```
Verification failed: 1 violation(s) found

Violation 1 (Temporal Property: progress):
  Property 'progress' violated in reachable state
  Counterexample trace:
    counter = 4
```

### How to Fix It

There are several ways to fix this specification, depending on your requirements:

#### Option 1: Remove the Problematic Property (Simplest)

If you don't need the strict progress guarantee, simply remove or comment out the `progress` property:

```spectre
// Removed the 'progress' property because it requires fairness to hold
// Without fairness constraints, the system allows infinite paths where:
// 1. reset is executed repeatedly (counter never reaches 10)
// 2. decrement/increment oscillate (counter oscillates between values < 10)
```

#### Option 2: Restrict Actions to Prevent Non-Progress Paths

Add conditions to prevent infinite loops. For example, only allow `reset` when counter is above a threshold:

```spectre
description "Resets the counter back to zero, but only when counter is greater than 10"
action reset {
  require counter > 10  // ✅ Only reset after reaching 10
  counter' = 0
}
```

This ensures that once counter reaches 10, it can be reset, but the reset can't prevent reaching 10 in the first place.

#### Option 3: Add Fairness Constraints

You can add fairness constraints to ensure that certain actions execute fairly:

```spectre
temporal progress {
  WF(increment) → always (counter < 10 → eventually counter = 10)
}
```

This means: "If `increment` has weak fairness (is continuously enabled and eventually executes), then progress holds."

The fairness constraint (`WF(increment)`) filters out execution paths where `increment` is continuously enabled but never executes, ensuring that only "fair" execution paths are considered when verifying the property.

### Key Lessons

1. **Temporal properties check all paths**: The `always` operator requires the property to hold on every possible execution path, not just one.

2. **Infinite paths matter**: Systems without fairness constraints can have infinite execution paths where progress never occurs.

3. **Fix strategies**: You can fix violations by:
   - Removing or relaxing the property
   - Restricting actions to prevent non-progress paths
   - Adding fairness conditions (when supported)

4. **Verification output is helpful**: The verifier tells you exactly which property failed and provides a counterexample trace showing where the violation occurs.

### The Corrected Specification

Here's a corrected version that removes the problematic property:

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
description "This property holds because there exists a path where we only increment"
temporal eventuallyReachesTen {
  eventually (counter = 10)
}

description "Ensures counter remains non-negative throughout execution"
temporal alwaysNonNegative {
  always (counter >= 0)
}
```

This corrected version verifies successfully because:
- The `eventuallyReachesTen` property only requires that there **exists** a path where counter reaches 10 (which is true: we can just increment repeatedly).
- The `alwaysNonNegative` property holds because the `decrement` action has a precondition preventing negative values.

---

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
3. **Read the spec**: See `docs/spec.md` for complete language details
4. **Explore advanced features**: Try modules, fairness, and complex types

### Resources

- **Language Specification**: `docs/spec.md`
- **Language Definition (formal grammar + semantics)**: `docs/language_definition.md`

Happy specifying! 🎉

---

*Last Updated: August 2026*

---

---

# Part II — Advanced Features

The following chapters cover the four research-grade capabilities that distinguish Spectre from standard model checkers. Each chapter is self-contained: it introduces the problem, shows a runnable example, explains the output, and describes how to act on the results.

---

## Chapter 9: Stuttering Detection and Repair

### What Is Stuttering?

A **stuttering step** is a transition in which an action executes but the system state does not change. Formally, a step `(σ, A, σ')` stutters when `σ = σ'`. Stuttering is not always wrong — an idle action when the system is quiescent is fine — but it becomes a problem when it prevents liveness properties from holding, because the verifier can construct infinite paths that loop on the stutter forever and never reach the goal.

### A Minimal Example

```spectre
// examples/stuttering-counter.spec

var counter: int

init { counter = 0 }

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Does nothing — causes stuttering"
action noop {
  counter' = counter   // identical to pre-state: stutter!
}

description "Resets the counter to zero"
action reset {
  counter' = 0
}

temporal eventuallyReachesTen {
  eventually (counter = 10)
}
```

Run it:

```bash
./spectre verify examples/stuttering-counter.spec --max-states 20
```

**Output (abridged):**

```
Traversed 20 states

Warnings (Stuttering):
Found 21 stuttering step(s) where a state transitions back to itself.
Stuttering can indicate missing fairness constraints or incomplete specifications.
  [1] in {counter = 0}, counter' = 0 (already 0)
      Option 1 — Redesign action `noop` (identity assignment)
        `counter' = counter` always produces the same value — this action can never make progress.
        // In action noop:
        // Replace `counter' = counter` with an assignment that changes the state.
      Option 2 — Add weak fairness on `increment`
        Guarantees `increment` is eventually taken whenever continuously enabled.
        temporal stutter_fix {
          WF(increment)
        }
  ...
Found no violations.
```

Spectre detects stuttering and immediately suggests two repair strategies for each stuttering site.

### Understanding the Output

Each reported stuttering entry shows:

- **The stuttering state** — `{counter = 0}` — where the self-loop occurs.
- **The identity assignment** — `counter' = counter` — that caused it.
- **Option 1 — Redesign**: the action is a pure no-op; redesign it to actually change something, or remove it.
- **Option 2 — Fairness**: keep the action but add `WF(increment)` so the verifier is required to also pick `increment` whenever it is enabled. This rules out the infinite `noop, noop, noop, …` path.

### Fix Option A — Remove the No-Op Action

If `noop` serves no modelling purpose, simply delete it:

```spectre
// examples/stuttering-counter-fixed.spec

var counter: int
init { counter = 0 }

action increment { counter' = counter + 1 }
action reset     { counter' = 0 }

temporal eventuallyReachesTen {
  eventually (counter = 10)
}
```

```bash
./spectre verify examples/stuttering-counter-fixed.spec --max-states 20
```

Still reports a temporal violation (`reset` prevents ever reaching 10), but **no stuttering warnings**.

### Fix Option B — Add Weak Fairness

If the idle action models a real system behaviour (e.g., a process polling), keep it and add fairness:

```spectre
// examples/stuttering-counter-with-fairness.spec

var counter: int
init { counter = 0 }

action increment { counter' = counter + 1 }
action noop      { counter' = counter }
action reset     { counter' = 0 }

temporal eventuallyReachesTen {
  WF(increment) → eventually (counter = 10)
}
```

`WF(increment)` says: "On any path where `increment` is continuously enabled, it will eventually execute." Since `increment` has no `require` guard it is always enabled, so weak fairness guarantees it fires infinitely often — and the counter will eventually reach 10 even if `noop` or `reset` also fire.

```bash
./spectre verify examples/stuttering-counter-with-fairness.spec
```

Output: no violations, no stuttering warnings about `increment`.

### Concurrent System Example

Stuttering also appears in concurrent models when a process takes an "idle step":

```spectre
// examples/stuttering-process.spec

enum ProcessState { Idle, Running }

var process1: ProcessState
var process2: ProcessState

init {
  process1 = ProcessState.Idle
  process2 = ProcessState.Idle
}

action process1Start  { require process1 = ProcessState.Idle;    process1' = ProcessState.Running }
action process1Finish { require process1 = ProcessState.Running; process1' = ProcessState.Idle   }
action process1Noop   { process1' = process1 }  // <-- stutter

action process2Start  { require process2 = ProcessState.Idle;    process2' = ProcessState.Running }
action process2Finish { require process2 = ProcessState.Running; process2' = ProcessState.Idle   }
action process2Noop   { process2' = process2 }  // <-- stutter

temporal process1EventuallyRuns {
  eventually (process1 = ProcessState.Running)
}
```

The two `Noop` actions create a state space where both processes stutter forever without ever running, causing the temporal property to fail. The repair suggestion is `WF(process1Start)`.

### When Stuttering Is Acceptable

Stuttering is not always a bug:

| Situation | Verdict |
|-----------|---------|
| Modelling a quiescent system that can stay idle | Acceptable — add `WF` if liveness is needed |
| Modelling a polling loop that waits for input | Acceptable |
| An action that forgot to change the state | Bug — redesign the action |
| An action that changes nothing under a specific `require` | Acceptable — but check whether that `require` is correctly guarded |

---

## Chapter 10: CEGIS Automatic Spec Repair

### What Is CEGIS?

**CEGIS** (Counterexample-Guided Inductive Synthesis) is a technique for automatically synthesising a repair — in Spectre's case, a `require` guard — that prevents an invariant from being violated. When verification finds an invariant violation, Spectre computes the **weakest precondition** (WP) of the violating assignment with respect to the invariant and proposes it as a `require` statement. It then re-runs verification with the proposed guard to confirm it actually fixes the problem.

### Step-by-Step Example

**The broken spec:**

```spectre
var balance: int

init { balance = 10 }

action withdraw(amount: int) {
  require amount > 0
  balance' = balance - amount   // ← no lower bound on result
}

invariant nonNegative {
  balance >= 0
}
```

Run verification:

```bash
./spectre verify myspec.spec
```

**Output:**

```
Exploring State 1

[VIOLATION DETECTED] Invariant: unknown
  Action 'withdraw' would violate invariants: [invariant nonNegative violated]

Continue exploration? (y/n): n
Traversed 1 states
Violations:
1. Invariant: Action 'withdraw' would violate invariants: [invariant nonNegative violated]

CEGIS Repair Suggestions:

  Invariant `nonNegative` violated via action `withdraw`:
    Pre-state: {balance = 10}
    Counterexample: withdraw*
    Option 1 — weakest precondition for `balance' = balance - amount`:
      `balance' = balance - amount` — with balance = 10, the result violated the invariant
      // In action withdraw — add:
      require balance - amount >= 0
      ✓ Verified: re-explored with this guard applied — invariant holds
```

### How to Read the Output

| Field | Meaning |
|-------|---------|
| `Invariant \`nonNegative\` violated via action \`withdraw\`` | Which invariant, which action |
| `Pre-state: {balance = 10}` | The state the system was in when the action attempted to fire |
| `Counterexample: withdraw*` | The action sequence leading to the violation; `*` marks the violating action |
| `Option 1 — weakest precondition for \`balance' = balance - amount\`` | The assignment that drove the invariant false |
| `require balance - amount >= 0` | The synthesised guard — the weakest condition under which the assignment is safe |
| `✓ Verified: …` | Spectre re-ran BFS with this guard injected and found no violations |

### Applying the Fix

Add the suggested `require` to the action:

```spectre
action withdraw(amount: int) {
  require amount > 0
  require balance - amount >= 0   // ← synthesised by CEGIS
  balance' = balance - amount
}
```

Or equivalently (and more idiomatically):

```spectre
action withdraw(amount: int) {
  require amount > 0
  require balance >= amount        // ← same condition, cleaner form
  balance' = balance - amount
}
```

Re-run to confirm:

```bash
./spectre verify myspec.spec
```

```
Traversed N states
Found no violations.
```

### Multiple Violation Example

When multiple actions violate the same invariant, CEGIS produces one repair suggestion per `(invariant, action)` pair — de-duplicated. For example:

```spectre
var balance: int
init { balance = 100 }

action withdraw(amount: int) {
  require amount > 0
  balance' = balance - amount
}

action fee {
  balance' = balance - 10
}

invariant nonNegative { balance >= 0 }
```

Running verification will produce:

```
CEGIS Repair Suggestions:

  Invariant `nonNegative` violated via action `withdraw`:
    Option 1 — weakest precondition for `balance' = balance - amount`:
      require balance - amount >= 0
      ✓ Verified: …

  Invariant `nonNegative` violated via action `fee`:
    Option 1 — weakest precondition for `balance' = balance - 10`:
      require balance - 10 >= 0
      ✓ Verified: …
```

### Two Violation Kinds

Spectre handles two structural forms of invariant violation:

| Kind | Description | When it occurs |
|------|-------------|----------------|
| **Type 1** — "would violate" | The action is about to produce a violating next state. Reported before the transition executes. | Common: invariant check on the proposed next state fails |
| **Type 2** — "state reached" | A state was already reached that violates the invariant (post-transition check). The last action in `Path` is the responsible one. | Occurs when the invariant evaluates the post-state rather than pre-testing the action |

CEGIS handles both: for Type 1 it uses the pre-state directly; for Type 2 it reconstructs the pre-state from the last path transition.

### Limitations

- CEGIS synthesises **`require` guards only** — it does not propose changes to the invariant or to the assignment itself.
- If the invariant references a variable that is not touched by any assignment in the action, no guard can be synthesised and the output says `(Could not synthesize an automatic repair for this violation)`.
- The verified re-exploration uses the same `--max-states` and `--max-depth` limits as the original run. If those limits are low, the `✓ Verified` claim is bounded.

---

## Chapter 11: Embedded Rust Runtime Monitor

### What Is a Runtime Monitor?

A **runtime monitor** is a small piece of code you embed in your production binary that checks your Spectre invariants on every state transition while the application runs. Instead of replaying traces offline, the monitor watches the live system and reports violations the instant they occur — no test harness needed.

Spectre generates a self-contained, dependency-free `monitor.rs` file from any `.spec` file.

### Generating the Monitor

```bash
spectre generate-monitor --lang rust examples/bank-account-corrected.spec -o src/monitor.rs
```

If `-o` is omitted, the file is printed to stdout.

**Generated file (abridged):**

```rust
//! Embedded runtime monitor generated from `bank-account-corrected`.
//!
//! # Quick start
//! ```rust,ignore
//! let mut monitor = Monitor::new(SpecState { /* initial values */ });
//! monitor.on_violation(|v| eprintln!("[SPEC VIOLATION] {}", v));
//!
//! // After every state transition in your application:
//! if let Err(violations) = monitor.step("action_name", SpecState { /* new values */ }) {
//!     // violations is Vec<Violation> — handle or log them
//! }
//!
//! // At shutdown, check liveness obligations:
//! for prop in monitor.unmet_liveness_properties() {
//!     eprintln!("[SPEC] liveness property `{}` was never satisfied", prop);
//! }
//! ```

#[derive(Debug, Clone)]
pub struct SpecState {
    pub aliceBalance: i64,
    pub bobBalance: i64,
}

#[derive(Debug, Clone)]
pub struct Violation {
    pub step: u64,
    pub action: String,
    pub property: String,
    pub message: String,
}

pub struct Monitor { /* ... */ }

impl Monitor {
    pub fn new(initial: SpecState) -> Self { /* ... */ }
    pub fn on_violation<F: Fn(&Violation) + Send + 'static>(&mut self, f: F) { /* ... */ }
    pub fn step(&mut self, action: &str, new_state: SpecState) -> Result<(), Vec<Violation>> { /* ... */ }
    pub fn unmet_liveness_properties(&self) -> Vec<&'static str> { /* ... */ }
}
```

### Integrating the Monitor in Your Application

**Step 1.** Copy `monitor.rs` into your crate (e.g. `src/monitor.rs`) and add `mod monitor;` to `main.rs` or `lib.rs`.

**Step 2.** Instantiate the monitor with your initial state:

```rust
use crate::monitor::{Monitor, SpecState};

let mut monitor = Monitor::new(SpecState {
    aliceBalance: 0,
    bobBalance: 0,
});

// Register a callback — called synchronously on every violation
monitor.on_violation(|v| {
    eprintln!("[SPEC VIOLATION] {}", v);
    // In production you might: log to alerting, increment a metric, panic, etc.
});
```

**Step 3.** After every state-changing operation, call `monitor.step`:

```rust
fn deposit(account: &mut Account, amount: i64, monitor: &mut Monitor) {
    account.balance += amount;

    // Snapshot your real state into SpecState and hand it to the monitor
    if let Err(violations) = monitor.step("depositAlice50", SpecState {
        aliceBalance: account.balance,
        bobBalance: other_account.balance,
    }) {
        for v in &violations {
            tracing::error!(violation = %v, "spec invariant violated in production");
        }
    }
}
```

**Step 4.** At application shutdown, check liveness obligations:

```rust
for prop in monitor.unmet_liveness_properties() {
    tracing::warn!("liveness property `{}` was never satisfied during this run", prop);
}
```

### What Gets Checked

The monitor translates each Spectre declaration into a Rust check:

| Spec construct | Monitor behaviour |
|----------------|-------------------|
| `invariant I { φ }` | Evaluated after every `step()`. Violation added to result if `φ` is false in the new state. |
| `temporal always φ` | Treated as an invariant — checked on every step. |
| `temporal eventually φ` | Tracked as a liveness obligation. `unmet_liveness_properties()` returns it if `φ` was never true. |
| Enum types | Generated as Rust enums matching the spec variants. |

### Example: Bank Account Monitor

The spec:

```spectre
var aliceBalance: int
var bobBalance: int

init { aliceBalance = 0; bobBalance = 0 }

action depositAlice50  { aliceBalance' = aliceBalance + 50 }
action withdrawAlice30 { aliceBalance' = aliceBalance - 30 }
action transfer50ToBob {
  aliceBalance' = aliceBalance - 50
  bobBalance'   = bobBalance   + 50
}

invariant no_negatives {
  aliceBalance >= 0 && bobBalance >= 0
}
```

Generated invariant check inside `monitor.rs`:

```rust
fn check_invariants(&self, action: &str) -> Vec<Violation> {
    let mut vs = Vec::new();
    // invariant: no_negatives
    if !(self.state.aliceBalance >= 0 && self.state.bobBalance >= 0) {
        vs.push(Violation {
            step: self.step_count,
            action: action.to_string(),
            property: "no_negatives".to_string(),
            message: "".to_string(),
        });
    }
    vs
}
```

No external dependencies, no runtime overhead beyond one function call per transition.

### Monitoring a System with Enums

When your spec uses enum types, the monitor generates matching Rust enums:

```spectre
enum AccountStatus { Active, Frozen, Closed }

var status: AccountStatus
var balance: int

invariant frozenMeansNoBalance {
  status = AccountStatus.Frozen → balance = 0
}
```

Generated:

```rust
#[derive(Debug, Clone, PartialEq)]
pub enum AccountStatus { Active, Frozen, Closed }

pub struct SpecState {
    pub status: AccountStatus,
    pub balance: i64,
}
```

The invariant `status = AccountStatus.Frozen → balance = 0` compiles to:

```rust
if !(!(self.state.status == AccountStatus::Frozen) || self.state.balance == 0) {
    // violation
}
```

### Online vs Offline Monitoring

| Mode | When | How |
|------|------|-----|
| **Offline (trace replay)** | Testing / CI | `spectre verify --emit-traces trace.itf.json` + `spectre generate-driver` |
| **Online (runtime monitor)** | Staging / production | `spectre generate-monitor` + embed in binary |

Use offline replay to find bugs during development. Use the runtime monitor as a safety net in staging or production where the exact production workload is needed to exercise the spec.

---

## Chapter 12: Spec Mining from Rust Source

### What Is Spec Mining?

**Spec mining** extracts a Spectre specification skeleton from existing Rust source code. Instead of writing a spec from scratch, you point `spectre mine` at a `.rs` file and get a working `.spec` that captures the structs, enums, method signatures, assert conditions, and assignment patterns it finds. The result is not a finished spec — it is a starting point that you refine.

### Running the Miner

```bash
spectre mine --lang rust src/account.rs -o account.spec
```

Optional flags:

| Flag | Description |
|------|-------------|
| `--lang rust` | Source language (only `rust` supported) |
| `-o <file>` | Write to file instead of stdout |
| `--spec-name <name>` | Override the inferred spec name |
| `--ai` | Enhance with an LLM (requires `ANTHROPIC_API_KEY`) |

### Worked Example

**Input Rust file (`bank_account.rs`):**

```rust
pub struct BankAccount {
    pub balance: i64,
    pub owner: String,
    pub frozen: bool,
}

impl BankAccount {
    pub fn deposit(&mut self, amount: i64) {
        assert!(amount > 0, "amount must be positive");
        assert!(self.balance + amount <= 1_000_000, "exceeds limit");
        self.balance += amount;
    }

    pub fn withdraw(&mut self, amount: i64) {
        assert!(amount > 0);
        assert!(self.balance >= amount);
        assert!(!self.frozen);
        self.balance -= amount;
    }

    pub fn freeze(&mut self) {
        assert!(!self.frozen);
        self.frozen = true;
    }

    pub fn unfreeze(&mut self) {
        assert!(self.frozen);
        self.frozen = false;
    }
}
```

**Run the miner:**

```bash
spectre mine --lang rust bank_account.rs
```

**Output:**

```spectre
// Spectre spec mined from: bank_account.rs
// Generated by: spectre mine --lang rust
// Review and refine before use in verification.

// ── State ──────────────────────────────────────────────
// Mined from: struct BankAccount

var balance: int
var owner: str
var frozen: bool

init {
  balance = 0
  owner = ""
  frozen = false
}

// ── Actions ────────────────────────────────────────────
// Mined from: impl BankAccount

description "Mined from: func deposit"
action deposit(amount: int) {
  require amount > 0
  require balance + amount <= 1000000
  balance' = balance + amount
}

description "Mined from: func withdraw"
action withdraw(amount: int) {
  require amount > 0
  require balance >= amount
  require !frozen
  balance' = balance - amount
}

description "Mined from: func freeze"
action freeze {
  require !frozen
  frozen' = true
}

description "Mined from: func unfreeze"
action unfreeze {
  require frozen
  frozen' = false
}
```

### What the Miner Extracts

| Rust construct | Spectre output |
|----------------|----------------|
| `struct` fields | `var` declarations with inferred Spectre types |
| `enum` variants (unit only) | `enum` declaration |
| `impl` methods (`&mut self`) | `action` with parameters |
| `assert!(cond)` / `assert!(cond, "msg")` | `require cond` |
| `if cond { return Err(...) }` | `require ¬cond` (negated guard) |
| `self.field = expr` or `self.field += expr` | `field' = expr` (next-state assignment) |
| Methods named `new`, `default`, `clone`, `fmt`, `drop` | Skipped |

### Type Mapping

| Rust type | Spectre type | Default value |
|-----------|--------------|---------------|
| `i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32`, `u64`, `usize`, `isize` | `int` | `0` |
| `f32`, `f64` | `float` | `0.0` |
| `bool` | `bool` | `false` |
| `String`, `&str`, `str` | `str` | `""` |
| `Vec<T>` | `List[T]` | `[]` |
| `HashSet<T>`, `BTreeSet<T>` | `Set[T]` | `{}` |
| `HashMap<K,V>`, `BTreeMap<K,V>` | `Map[K,V]` | `Map.empty()` |
| `Option<T>` | `T` (unwrapped) | type default |
| `Arc<T>`, `Mutex<T>`, `Rc<T>`, `RefCell<T>`, `Box<T>` | `T` (unwrapped) | type default |
| `&T`, `&mut T` | `T` (reference stripped) | type default |
| Enum named in file | Enum type | `EnumName.Default` |
| Unknown | `int` | `0` |

### LLM Enhancement (`--ai`)

With `--ai`, Spectre sends the mined spec to Claude (claude-haiku) and asks it to:

1. Suggest **natural-language descriptions** for each action based on the code context.
2. Propose **additional invariants** (e.g., "balance is always non-negative after any operation").
3. Suggest **temporal properties** (e.g., "a frozen account eventually becomes unfrozen").

**Setup:**

```bash
export ANTHROPIC_API_KEY=sk-ant-...
spectre mine --lang rust --ai bank_account.rs -o account.spec
```

The LLM suggestions are injected as commented-out blocks so you can review them before uncommenting:

```spectre
// ── Invariants (LLM-suggested) ─────────────────────────
// description "LLM: balance should never go negative"
// invariant llm_balance_non_negative {
//   balance >= 0
// }

// ── Temporal properties (LLM-suggested) ──────────────
// temporal llm_eventually_unfrozen {
//   eventually (!frozen)
// }
```

The `--ai` flag requires `ANTHROPIC_API_KEY` to be set. If the key is absent or the API call fails, the miner continues with static analysis only.

### The Full Workflow

```
Rust source  ──mine──>  .spec skeleton  ──refine──>  .spec  ──verify──>  violations
                                                                               │
                                                                          CEGIS suggests fix
                                                                               │
                                                                          apply & re-verify
```

1. **Mine**: `spectre mine --lang rust src/account.rs -o account.spec`
2. **Review**: open `account.spec` and check that the mined state variables, actions, and preconditions are correct. Add missing invariants.
3. **Verify**: `spectre verify account.spec`
4. **Repair**: if violations are found, read the CEGIS suggestions and add the recommended `require` guards.
5. **Iterate**: keep refining until `Found no violations.`

### Example: Refining a Mined Spec

After mining, you might notice the spec is missing an invariant that the code assumes globally. Add it manually:

```spectre
// After mining, add:
description "Balance can never be negative"
invariant balanceNonNegative {
  balance >= 0
}

description "Frozen account cannot be deposited into"
// The miner didn't add this because there's no assert for it in deposit:
// Un-comment after reviewing:
// invariant frozenMeansNoDeposit {
//   frozen → balance = balance  // placeholder: strengthen as needed
// }
```

Then verify:

```bash
spectre verify account.spec
```

If the invariant is violated, CEGIS will suggest the exact `require` to add.

### Limitations

- **Only unit enum variants are mined.** Tuple variants (`Some(T)`) and struct variants (`Foo { x: i32 }`) are skipped to keep the mined spec clean.
- **Complex expressions are approximated.** Bitwise operations, method chains, and closure bodies in assert conditions are simplified or skipped.
- **No control-flow analysis.** The miner uses regex-based extraction, not a full AST parse. Conditions inside `match` arms or nested `if` blocks may be missed.
- **The mined spec is a starting point, not a proof.** Always review the output before running verification and treat the result as a draft that needs human refinement.

---

## Chapter 13: Model-Based Testing with Rust

### What Is Model-Based Testing?

Model-Based Testing (MBT) uses a formal spec as the *oracle* for testing a real implementation. Spectre's MBT workflow has three steps:

1. **Verify** the spec and emit a counterexample or witness trace as an ITF JSON file.
2. **Generate** a Rust driver skeleton that knows how to replay that trace.
3. **Fill in** the driver with calls to your real implementation and run it.

If the implementation behaves differently from the spec, the driver reports a mismatch.

### Step 1 — Emit a Trace

```bash
spectre verify myspec.spec --emit-traces trace.itf.json
```

This writes the first counterexample (or if there are no violations, the longest explored path) to `trace.itf.json` in the [ITF (Interactive Traces Format)](https://apalache.informal.systems/docs/adr/015adr-trace.html) used by Apalache and Quint.

### Step 2 — Generate a Driver

```bash
spectre generate-driver --lang rust myspec.spec -o src/driver.rs
```

**Generated `driver.rs` (abridged):**

```rust
// Auto-generated by `spectre generate-driver myspec.spec`
// Fill in the TODO sections to connect your Rust implementation.

use spectre_connect::{ItfValue, StepHandler, TraceReplayer};
use std::collections::HashMap;

/// Holds your system's state. Replace field types with your real types.
pub struct MySystem {
    pub aliceBalance: i64,
    pub bobBalance: i64,
}

impl StepHandler for MySystem {
    type Error = Box<dyn std::error::Error>;

    fn apply_action(&mut self, action: &str, args: &[ItfValue]) -> Result<(), Self::Error> {
        match action {
            "depositAlice50" => {
                // TODO: call your real implementation
                todo!()
            }
            "withdrawAlice30" => {
                // TODO: call your real implementation
                todo!()
            }
            _ => Err(format!("unknown action: {action}").into()),
        }
    }

    fn current_state(&self) -> HashMap<String, ItfValue> {
        let mut s = HashMap::new();
        s.insert("aliceBalance".to_string(), ItfValue::Int(self.aliceBalance));
        s.insert("bobBalance".to_string(), ItfValue::Int(self.bobBalance));
        s
    }
}

fn main() {
    let trace_path = std::env::args().nth(1)
        .expect("Usage: driver <trace.itf.json>");
    let system = MySystem::new();
    TraceReplayer::new(system)
        .run(&trace_path)
        .expect("trace replay failed");
}
```

### Step 3 — Fill In and Run

Replace each `todo!()` with a call to your real Rust implementation:

```rust
"depositAlice50" => {
    self.account.deposit(50)?;
    self.aliceBalance = self.account.balance();
    Ok(())
}
"withdrawAlice30" => {
    self.account.withdraw(30)?;
    self.aliceBalance = self.account.balance();
    Ok(())
}
```

Add `spectre-connect` to your `Cargo.toml`:

```toml
[dependencies]
spectre-connect = "0.1"
```

Then run:

```bash
cargo run -- trace.itf.json
```

The `TraceReplayer` calls `apply_action` for each step in the trace, then compares `current_state()` with the expected state from the ITF file. A mismatch means your implementation diverges from the spec.

### Comparing the Three Testing Modes

| Mode | When to use | What it tests |
|------|-------------|---------------|
| `spectre verify` | Spec development | The spec itself — finds logical gaps |
| `spectre verify --emit-traces` + driver | CI / integration testing | Implementation vs spec on specific paths |
| `spectre generate-monitor` | Staging / production | Implementation vs spec on live traffic |

The three modes are complementary: use verification to prove the spec is correct, MBT to test the implementation against specific traces, and the runtime monitor to catch deviations in production.

---

## Chapter 14: Incremental Verification and Drift Detection

As your Rust codebase evolves, re-running a full BFS from scratch on every change is expensive.  Spectre provides three mechanisms to keep verification fast and continuous: **state caching**, **incremental re-verification**, and **semantic sync**.

---

### State Caching

The `--use-cache` flag saves the explored BFS graph to disk and restores it on subsequent runs.  When the spec has not changed at all, cache restore skips the entire traversal:

```bash
# First run: explore 22,432 states and write cache
spectre verify examples/raft-election-safety.spec --max-states 30000 --use-cache
# Explored 22432 states. Cache saved.

# Every subsequent run on an unchanged spec restores from disk
spectre verify examples/raft-election-safety.spec --max-states 30000 --use-cache
# Restored state graph from cache (22432 states). Elapsed: 0.26 s
```

Cache restore is **47× faster** than a cold BFS for the 22,432-state Raft model (0.265 s vs 12.4 s on Apple M3 Pro, macOS 15.6, Go 1.24). Absolute times are hardware-dependent; the speedup ratio is the primary claim.

---

### Incremental Re-verification

When `spectre sync` reports that one action has changed (`[✗]`), the safe path is to re-verify — but only that action needs to change:

```bash
# After modifying vote2for1's guard in the Rust source:
spectre verify examples/raft-election-safety.spec --max-states 30000 \
    --use-cache --incremental --changed-action vote2for1
```

The algorithm:
1. **Prune** — remove all transitions produced by the changed action from the cached graph
2. **Reachability BFS** — recompute reachable states over remaining edges
3. **Remove unreachable states** — discard states only reachable via the changed action
4. **Re-execute** — run the changed action from every remaining reachable state
5. **BFS-expand** — expand new states with all actions, capped at the original bound

**Measured speedups (Apple M3 Pro; ratios are hardware-independent):**

| Spec | Changed action | Cold BFS | Incremental | Speedup | States pruned / new |
|------|---------------|----------|-------------|---------|---------------------|
| Raft (22,432) | `stepDown1_sees2` | 12.4 s | 2.4 s | **5.2×** | 1,649 / 1,645 |
| Raft (22,432) | `vote2for1` | 12.4 s | 2.9 s | **4.3×** | 6,501 / 6,495 |

Speedup depends on what fraction of states are reachable only via the changed action.

---

### Semantic Sync with `spectre sync`

As you modify Rust source, `spectre sync` compares the current implementation against the `.spec` file and classifies every change:

```bash
spectre sync examples/rust/bank_account.rs --spec examples/bank-account-parameterized.spec
```

Example output (semantic mutation: `balance - amount` → `balance + amount`):

```
[=] action deposit (structurally unchanged)
[✗] action withdraw assignment 1 body changed: "balance'=balance - amount" → "balance'=balance + amount" (witness: amount=1, balance=0)
[=] action freeze (structurally unchanged)
[=] action unfreeze (structurally unchanged)
```

Example output (equivalent rewrite — guard commuted):

```
[≡] action deposit guard 2 rewritten (SMT-proved equivalent): "balance + amount <= 1000000" → "amount + balance <= 1000000"
[=] action withdraw (structurally unchanged)
```

Each symbol means:

| Symbol | Meaning |
|--------|---------|
| `[=]` | Structurally unchanged — no re-verification needed |
| `[≡]` | SMT-proved equivalent rewrite — safe, no re-verification needed |
| `[✗]` | Semantically changed — SMT witness provided; run `--incremental` or full BFS |

`spectre sync` uses Z3/QF_LIA for equivalence checking. In evaluation: **8/8 semantic mutations detected** with **1 false positive** (vs. 4 FP using AST comparison alone).

---

### Drift Detection with `spectre drift`

`spectre drift` checks bidirectional staleness between a spec and its `.driver.json` sidecar (produced by `spectre mine`).  It reports: fields in Rust with no spec variable, spec variables with no Rust mapping, and invariants that may need monitor regeneration.

```bash
spectre drift account.spec
```

---

### The Full CI Workflow

```
mine → verify → commit spec ──→ evolving code → sync → incremental re-verify
                                      ↑                         ↓
                                      └──── drift → regenerate driver / monitor ──┘
```

In CI, add these two checks:

```bash
# Fail the build if any action is semantically changed ([✗])
spectre sync src/account.rs --spec spec/account.spec

# Re-verify the flagged action (fast path)
spectre verify spec/account.spec --max-states 30000 \
    --use-cache --incremental --changed-action <flagged-action>
```

---

## Chapter 15: Coverage-Guided Model-Based Testing

When you run `spectre verify --emit-traces`, the verifier writes ITF trace files that the `generate-driver` command turns into Rust test harnesses.  The `--coverage-mode` flag controls which execution paths are prioritised when generating traces.

---

### Coverage Modes

```bash
# Property-directed: steers toward states near invariant boundaries
spectre verify examples/bank-account-parameterized.spec \
    --emit-traces trace.itf.json --coverage-mode property

# Transition-pair: exercises every (prev-action, next-action) pair
spectre verify examples/bank-account-parameterized.spec \
    --emit-traces trace.itf.json --coverage-mode transition-pair
```

| Mode | Strategy | Best for |
|------|---------|----------|
| `action` | Cover each action at least once | Quick smoke test |
| `transition-pair` | Cover every (A → B) action sequence | State-machine edge coverage |
| `boundary` | Steer toward maximum integer values | Overflow / capacity bugs |
| `rare-action` | Prefer least-executed actions globally | Dead-action detection |
| `property` | Steer toward invariant boundary states | Pre-condition / guard testing |

In evaluation on bank-account, `property` mode achieved **52% boundary-state coverage** — states where `balance` was within 1 of 0 or the cap — compared to 5–25% for other modes.

---

### Selecting a Mode

- Use `property` when you want to stress-test invariant guards: it steers traces toward the states most likely to violate an invariant.
- Use `transition-pair` when you want systematic edge coverage of the action graph.
- Use `boundary` when you have integer variables with known caps (e.g. `balance <= 1_000_000`).
- Use `rare-action` when the spec has actions that are rarely reached in practice but matter for correctness.

---

### Full MBT Workflow

```bash
# 1. Generate a property-directed trace
spectre verify examples/bank-account-parameterized.spec \
    --emit-traces traces/bank.itf.json --coverage-mode property

# 2. Generate a Rust driver
spectre generate-driver --lang rust examples/bank-account-parameterized.spec \
    --output src/spec_driver.rs

# 3. Fill in the TODO sections in spec_driver.rs

# 4. Replay the trace
spectre check-refinement examples/bank-account-parameterized.spec \
    --traces traces/bank.itf.json
```

