# Chapter 2: Language Overview

This chapter provides a comprehensive overview of the Spectre language, its design philosophy, and all its core elements.

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

Modules help organize your specifications into reusable components:

```spectre
module Counter {
  description "Tracks a numeric counter value"
  var counter: int
  
  public action increment {
    counter' = counter + 1
  }
}

// Later, import and use:
import Counter
```

Modules support:
- **Inheritance**: Modules can extend other modules
- **Visibility**: `public` and `private` modifiers
- **Super calls**: Call parent module methods

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

Each chapter will provide detailed syntax, examples, and best practices for using that language element effectively.


