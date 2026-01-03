# Spectre Language Specification

## Overview

Spectre is a programmer-friendly specification language inspired by TLA+ and Quint, designed to be accessible to Java and TypeScript developers. It provides a clean syntax for defining state machines, transitions, and verification properties.

## Design Principles

1. **Familiar Syntax**: Uses syntax similar to mainstream programming languages
2. **Strong Typing**: Supports both primitive and compound types with type inference
3. **Clear State Management**: Explicit state variable declarations and transitions
4. **Comprehensive Verification**: Supports both classic constraints (invariants) and temporal constraints
5. **Readable Error Messages**: Descriptions provide context in error traces and stack traces

## Descriptions and Error Messages

All language elements can include a `description` field to provide human-readable names and context. These descriptions are used in error messages, stack traces, and verification reports to make debugging easier and more intuitive.

### Syntax

Descriptions are added using the `description` keyword before the element definition:

```spectre
description "Human-readable description"
var counter: int

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Ensures counter never goes negative"
invariant nonNegative {
  counter >= 0
}
```

### Supported Elements

Descriptions can be added to:
- **State Variables**: `description "..." var name: Type`
- **Initial States**: `description "..." init { ... }`
- **Actions**: `description "..." action name { ... }`
- **Pure Functions**: `description "..." fun name(...): Type { ... }`
- **Invariants**: `description "..." invariant name { ... }`
- **Temporal Properties**: `description "..." temporal name { ... }`
- **Preconditions**: `description "..." require condition`
- **Postconditions**: `description "..." ensure condition`

### Error Messages with Descriptions

When a verification error occurs, descriptions provide context:

**Without descriptions:**
```
Error: Invariant violated
  Invariant: nonNegative
  State: counter = -1
```

**With descriptions:**
```
Error: Invariant violated
  Invariant: "Ensures counter never goes negative" (nonNegative)
  State: counter = -1
  Trace:
    1. Initial state: "System starts with counter at zero"
    2. Action: "Decrements counter when positive" (decrement)
    3. Action: "Decrements counter when positive" (decrement)
```

### Stack Traces

When an invariant or temporal property fails, the verifier reports a full execution trace showing the sequence of actions that led to the error state:

```
Error: Invariant violated
  Invariant: "Mutual exclusion must be maintained" (mutualExclusion)
  State: process1 = Critical, process2 = Critical, lock = true
  
  Execution Trace:
    Step 0: Initial state: "System starts with both processes idle"
      process1 = Idle, process2 = Idle, lock = false
    
    Step 1: Action: "Process 1 requests lock" (process1Request)
      process1 = Critical, process2 = Idle, lock = true
    
    Step 2: Action: "Process 2 requests lock" (process2Request)
      process1 = Critical, process2 = Critical, lock = true
      ❌ Invariant violated here
```

### Best Practices

1. **Be Descriptive**: Use clear, concise descriptions that explain the purpose
   ```spectre
   description "Counter tracks the number of active user sessions"
   var counter: int
   ```

2. **Include Context**: Mention why the constraint exists
   ```spectre
   description "Prevents account balance from going negative, ensuring financial integrity"
   invariant balanceNonNegative {
     account.balance >= 0
   }
   ```

3. **Action Descriptions**: Explain what the action does and when it's used
   ```spectre
   description "Transfers funds between accounts, ensuring sufficient balance and account validity"
   action transfer(fromId: int, toId: int, amount: int) {
     ...
   }
   ```

4. **Temporal Properties**: Explain what behavior is being verified
   ```spectre
   description "Guarantees that every request eventually receives a response"
   temporal requestResponse {
     always (requestSent → eventually responseReceived)
   }
   ```

---

## Table of Contents

1. [Types](#types)
2. [Modules](#modules)
3. [Constants](#constants)
4. [State Variables](#state-variables)
5. [Initial States](#initial-states)
6. [Transitions](#transitions)
7. [Pure Functions](#pure-functions)
8. [Constraints](#constraints)
9. [Temporal Properties](#temporal-properties)
10. [Examples](#examples)

---

## Types

### Primitive Types

- `int`: Integer numbers
- `bool`: Boolean values (`true`, `false`)
- `str`: String literals
- `float`: Floating-point numbers

### Compound Types

#### Records (Structs)
```spectre
type User = {
  id: int,
  name: str,
  active: bool
}
```

#### Tuples
```spectre
type Point = (int, int)
```

#### Sets

Sets represent unordered collections of unique elements:

```spectre
type UserSet = Set<User>
type IntSet = Set<int>
```

**Set Literals:**

You can create sets using literal syntax with curly braces:

```spectre
// Empty set
const emptySet: Set<int> = {}

// Set with elements
const numbers: Set<int> = { 1, 2, 3, 4, 5 }
const names: Set<str> = { "alice", "bob", "charlie" }

// Set of records
type User = { id: int, name: str }
const users: Set<User> = { 
  { id: 1, name: "alice" }, 
  { id: 2, name: "bob" } 
}

// Sets can also be created using static methods (see Set Operations)
const altEmpty: Set<int> = Set.empty()
const singleton: Set<int> = Set.of(42)
```

**Note**: Set literals use curly braces `{ }`. To distinguish them from record literals (which use `{ field: value }`), the parser checks if the content contains colons. If it does, it's a record literal; otherwise, it's a set literal.

#### Maps (Dictionaries)
```spectre
type UserMap = Map<int, User>  // Maps int keys to User values
```

#### Arrays/Lists

Lists represent ordered sequences of elements:

```spectre
type UserList = List<User>
type IntArray = List<int>
```

**List Literals:**

You can create lists using literal syntax with square brackets:

```spectre
// Empty list
const emptyList: List<int> = []

// List with elements
const numbers: List<int> = [ 1, 2, 3, 4, 5 ]
const names: List<str> = [ "alice", "bob", "charlie" ]

// List of records
type User = { id: int, name: str }
const users: List<User> = [ 
  { id: 1, name: "alice" }, 
  { id: 2, name: "bob" } 
]

// Lists can also be created using static methods (see List Operations)
const altEmpty: List<int> = List.empty()
```

**Note**: List literals use square brackets `[ ]` to distinguish them from sets (which use curly braces `{ }`).

#### Enums
```spectre
enum Status {
  Pending,
  Processing,
  Completed,
  Failed
}
```

#### Optionals
```spectre
type OptionalUser = Option<User>  // Can be Some(User) or None
```

---

## Modules

Modules provide code organization and reusability. A module can contain [types](#types), [constants](#constants), [state variables](#state-variables), [actions](#transitions), [invariants](#constraints), and [temporal properties](#temporal-properties).

### Module Definition

```spectre
module Counter {
  var counter: int
  
  init {
    counter = 0
  }
  
  action increment {
    counter' = counter + 1
  }
}
```

### Module Imports

Import modules to use their definitions:

```spectre
import Counter
import UserManagement

// Use imported module members
action useCounter {
  Counter.increment()  // Call action from Counter module
  let max = Counter.MAX_VALUE  // Access constant from Counter module
}
```

**Import Rules:**
- **Qualified access**: Use `ModuleName.memberName` to access imported members
- **Name conflicts**: If multiple modules have members with the same name, use qualified names
- **Public members only**: Only public members can be accessed from imported modules
- **Scope**: Imported modules are available throughout the current module/file

### Extending Modules

Extend a module to add new definitions or override existing ones:

```spectre
module BoundedCounter extends Counter {
  const MAX_VALUE: int = 100
  
  // Override parent action
  action increment {
    require counter < MAX_VALUE
    super.increment()  // Call parent's increment action
  }
  
  // Add new invariant
  invariant bounded {
    counter <= MAX_VALUE
  }
}
```

**Extension Rules:**
- **Inheritance**: Child module inherits all public members from parent
- **Override**: Child can override parent actions by defining a new action with the same name
- **Super calls**: Use `super.actionName()` to call the parent's version of an action
- **New members**: Child can add new constants, variables, actions, invariants, etc.
- **State variables**: Inherited state variables are shared (not copied)

### Module Instances

Create instances of modules with parameter substitution. This allows you to use a module with different variable names:

```spectre
module Counter {
  var counter: int
  action increment { counter' = counter + 1 }
}

// Create an instance with a different variable name
module MyCounter = Counter with {
  counter = myCounter
}

// Now you can use MyCounter.increment() which modifies myCounter
```

**Note:** Module instances are primarily useful for reusing module logic with different variable names. For most cases, importing and extending modules is preferred.

### Public and Private

Control visibility of module members:

```spectre
module Counter {
  public var counter: int  // Can be accessed from outside
  private var internal: int  // Only accessible within module
  
  public const MAX_VALUE: int = 100  // Can be accessed as Counter.MAX_VALUE
  private const INTERNAL_CONST: int = 5  // Only accessible within module
  
  public action increment { ... }
  private action helper { ... }
}
```

**Visibility Rules:**
- **Public members**: Can be accessed from outside the module using `ModuleName.memberName`
- **Private members**: Only accessible within the module
- **Default visibility**: If not specified, members are public
- **Constants**: Public constants can be accessed as `ModuleName.CONSTANT_NAME`

### Example: Module Organization

```spectre
// File: base/counter.spec
module Counter {
  var counter: int
  
  init {
    counter = 0
  }
  
  action increment {
    counter' = counter + 1
  }
}

// File: main.spec
import Counter

module BoundedCounter extends Counter {
  action increment {
    require counter < 100
    super.increment()  // Call parent action
  }
}
```

---

## Constants

Constants are values that do not change during execution. They are useful for parameterizing specifications and defining configuration values.

### Syntax

Constants are declared using the `const` keyword:

```spectre
const MAX_USERS: int = 100
const TIMEOUT: int = 30
const SERVER_NAME: str = "api.example.com"
```

### Constant Types

Constants can be of any type:

```spectre
const N: int = 10
const ENABLED: bool = true
const DEFAULT_ROLE: str = "user"
const MAX_RETRIES: int = 3
```

### Using Constants

Constants can be used in:
- Initial states
- Actions (in conditions and expressions)
- Invariants
- Temporal properties
- Pure functions

```spectre
const MAX_COUNTER: int = 100

var counter: int

init {
  counter = 0
}

action increment {
  require counter < MAX_COUNTER
  counter' = counter + 1
}

invariant bounded {
  counter <= MAX_COUNTER
}
```

### Constant Expressions

Constants can be computed from other constants:

```spectre
const BASE_TIMEOUT: int = 10
const MAX_TIMEOUT: int = BASE_TIMEOUT * 3
const RETRY_COUNT: int = 5
```

### Parameterized Specifications

Constants enable parameterized specifications (see also [Modules](#modules) for more advanced parameterization):

```spectre
const NUM_PROCESSES: int = 2
const MAX_ITERATIONS: int = 1000

var processes: Set<Process>

init {
  processes = {}  // Empty set literal
}

action addProcess {
  require processes.size() < NUM_PROCESSES
  // ...
}
```

### Constants vs Variables

- **Constants** (`const`): Fixed values, set at specification time, cannot change
- **Variables** (`var`): State values, can change through actions

```spectre
const MAX_SIZE: int = 100  // Never changes
var currentSize: int       // Can change via actions
```

---

## State Variables

State variables are declared using the `var` keyword:

```spectre
var counter: int
var users: Set<User>
var status: Status
var config: Map<str, int>
```

Multiple variables can be declared together:

```spectre
var x: int, y: int, z: int
```

---

## Initial States

The initial state is defined using the `init` action:

```spectre
init {
  counter = 0
  users = {}  // Empty set literal
  status = Status.Pending
}
```

Or using a single expression:

```spectre
init counter = 0 && users = {}
```

### Non-deterministic Initial States

When a system can start from multiple possible initial states, use the `oneOf` operator to specify a set of initial conditions. The verifier will explore all possible initial states.

**Single Variable Syntax:**
```spectre
init oneOf {
  counter = 0,
  counter = 5,
  counter = 10
}
```

**Multiple Variables - Tuple Syntax:**
For multiple variables, use tuple syntax where each option specifies values for all state variables:

```spectre
init oneOf {
  { counter = 0, status = Status.Pending },
  { counter = 5, status = Status.Processing },
  { counter = 10, status = Status.Completed }
}
```

**Multiple Variables - Block Syntax:**
For readability with many variables, use block syntax:

```spectre
init oneOf {
  {
    counter = 0
    status = Status.Pending
  },
  {
    counter = 5
    status = Status.Processing
  },
  {
    counter = 10
    status = Status.Completed
  }
}
```

**All three syntaxes are valid.** Use tuple syntax for simple cases and block syntax when you have many variables or want better readability.

The `oneOf` operator is particularly useful for:
- Testing systems with different starting configurations
- Modeling systems where the initial state is unknown or variable
- Exploring all possible initial conditions during verification

---

## Transitions

Transitions are defined using the `action` keyword. The prime notation (`'`) is used to denote the next state:

```spectre
action increment {
  counter' = counter + 1
}

action decrement {
  counter' = counter - 1
}
```

### Conditional Transitions

```spectre
action incrementIfPositive {
  if (counter > 0) {
    counter' = counter + 1
  } else {
    counter' = counter  // No change
  }
}
```

### Multiple State Updates

```spectre
action addUser(user: User) {
  users' = users.union({ user })  // Using set literal
  counter' = counter + 1
}
```

### Transition Guards

Guards can be specified inline:

```spectre
action increment when counter < 100 {
  counter' = counter + 1
}
```

---

## Pure Functions

Pure functions are computational helpers that perform calculations without modifying state. They are defined using the `fun` keyword and can only contain pure operations (no state changes).

### Syntax

```spectre
fun functionName(arg1: Type1, arg2: Type2): ReturnType {
  // Pure operations only - no state changes
  // Can read arguments and local variables
  // Can call other pure functions
  // Cannot access or modify state variables
  return expression
}
```

### Pure Operations

Pure functions can only use:
- Function arguments
- Local variables (declared with `let`)
- Constants (`const` declarations)
- Other pure functions
- Pure expressions (arithmetic, comparisons, etc.)
- Collection operations that don't modify state (e.g., `map`, `filter`, `reduce`)

Pure functions **cannot**:
- Access state variables (`var` declarations)
- Modify state variables
- Call actions
- Use prime notation (`'`)

### Examples

```spectre
// Simple arithmetic function
fun add(x: int, y: int): int {
  return x + y
}

// Function with conditional logic
fun max(a: int, b: int): int {
  if (a > b) {
    return a
  } else {
    return b
  }
}

// Function operating on collections
fun sum(numbers: List<int>): int {
  return numbers.reduce(0, (acc, n) => acc + n)
}

// Function operating on records
fun isActive(user: User): bool {
  return user.active && user.id > 0
}

// Function with multiple statements
fun calculateTotal(users: Set<User>): int {
  let activeUsers = users.filter(u => u.active)
  return activeUsers.size()
}
```

### Using Pure Functions

Pure functions can be called from:
- Actions (transitions)
- Invariants
- Temporal properties
- Other pure functions
- Preconditions and postconditions

```spectre
var counter: int
var users: Set<User>

fun getUserCount(userSet: Set<User>): int {
  return userSet.size()
}

fun isValidCount(count: int): bool {
  return count >= 0 && count <= 100
}

action increment {
  require isValidCount(counter)
  counter' = counter + 1
}

invariant userCountValid {
  isValidCount(getUserCount(users))
}
```

### Recursive Functions

Pure functions can be recursive:

```spectre
fun factorial(n: int): int {
  if (n <= 1) {
    return 1
  } else {
    return n * factorial(n - 1)
  }
}

fun fibonacci(n: int): int {
  if (n <= 1) {
    return n
  } else {
    return fibonacci(n - 1) + fibonacci(n - 2)
  }
}
```

### Type Inference

Return types can often be inferred, but explicit types are recommended for clarity:

```spectre
fun add(x: int, y: int): int {  // Explicit return type
  return x + y
}

fun multiply(x: int, y: int) {  // Inferred return type (int)
  return x * y
}
```

---

## Constraints

### Invariants (Classic Constraints)

Invariants are properties that must hold in all states:

```spectre
invariant nonNegative {
  counter >= 0
}

invariant usersNotEmpty {
  users.size() > 0
}
```

### Preconditions

Preconditions must hold before a transition can execute:

```spectre
action decrement {
  require counter > 0
  counter' = counter - 1
}
```

### Postconditions

Postconditions must hold after a transition executes:

```spectre
action increment {
  counter' = counter + 1
  ensure counter' > counter
}
```

---

## Temporal Properties

Temporal properties describe behavior over sequences of states:

### Always

```spectre
temporal alwaysPositive {
  always (counter >= 0)
}
```

### Eventually

```spectre
temporal eventuallyReachesTen {
  eventually (counter = 10)
}
```

### Until

```spectre
temporal counterUntilTen {
  counter < 10 until counter = 10
}
```

### Leads To (→)

The `→` operator (pronounced "leads to") states that if the left-hand condition becomes true, then the right-hand condition will eventually become true:

```spectre
temporal requestLeadsToResponse {
  requestSent → responseReceived
}
```

This means: whenever `requestSent` becomes true, `responseReceived` will eventually become true.

**Note:** The `→` operator can only be used within temporal properties. It is syntactic sugar for `always (P → eventually Q)`.

### Combination Operators

```spectre
temporal eventuallyAlways {
  eventually always (status = Status.Completed)
}

temporal alwaysEventually {
  always eventually (counter = 0)
}
```

### Fairness Conditions

Fairness conditions ensure that certain actions eventually execute when they are enabled. This is crucial for verifying liveness properties in concurrent systems.

#### Weak Fairness

Weak fairness states that if an action is continuously enabled, it will eventually execute:

```spectre
temporal weakFairness {
  WF(increment)
}
```

This means: if `increment` is **continuously enabled** (enabled at every step from some point onward), it will execute infinitely often.

#### Strong Fairness

Strong fairness states that if an action is enabled infinitely often, it will execute infinitely often:

```spectre
temporal strongFairness {
  SF(decrement)
}
```

This is stronger than weak fairness: even if the action is not continuously enabled, as long as it becomes enabled infinitely often, it will execute.

#### Fairness on Actions

Apply fairness to specific actions:

```spectre
description "Ensures increment action executes fairly"
temporal incrementFairness {
  WF(increment)
}

description "Ensures decrement action executes fairly"
temporal decrementFairness {
  SF(decrement)
}
```

#### Fairness on Variables

Apply fairness to all actions that modify specific variables. This applies fairness to every action that changes the variable's value:

```spectre
description "Weak fairness for all actions modifying counter"
temporal counterFairness {
  WF(counter)
}

description "Strong fairness for all actions modifying users"
temporal usersFairness {
  SF(users)
}
```

**How it works:**
- `WF(variable)` applies weak fairness to **all** actions that modify the variable
- `SF(variable)` applies strong fairness to **all** actions that modify the variable
- An action modifies a variable if it assigns a new value using prime notation (`variable' = ...`)
- If multiple actions modify the variable, fairness applies to all of them

#### Fairness Examples

**Example 1: Process Fairness**

```spectre
enum ProcessState {
  Idle,
  Critical
}

var process1: ProcessState
var process2: ProcessState
var lock: bool

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

description "Ensures process1 gets fair access to the lock"
temporal process1Fairness {
  WF(process1Request)
}

description "Ensures process2 gets fair access to the lock"
temporal process2Fairness {
  WF(process2Request)
}
```

**Example 2: Message Processing Fairness**

```spectre
var queue: List<Message>
var processed: Set<int>

action dequeue {
  require queue.size() > 0
  let msg = queue.head()
  queue' = queue.tail()
  processed' = processed.union({ msg.id })  // Using set literal
}

description "Ensures messages are processed fairly when queue is not empty"
temporal processingFairness {
  WF(dequeue)
}

description "Guarantees all messages eventually get processed"
temporal allMessagesProcessed {
  always (queue.size() > 0 → eventually processed.contains(queue.head().id))
}
```

**Example 3: Strong Fairness for Critical Actions**

```spectre
action criticalOperation {
  require canExecute()
  // Critical operation
}

description "Strong fairness ensures critical operation executes when possible"
temporal criticalFairness {
  SF(criticalOperation)
}
```

#### Fairness and Liveness

Fairness conditions are essential for proving [liveness properties](#temporal-properties) (properties that guarantee something eventually happens):

```spectre
description "Without fairness, this might not hold"
temporal eventuallyProcess1Critical {
  eventually (process1 = ProcessState.Critical)
}

description "With fairness, this is guaranteed"
temporal fairnessGuaranteesAccess {
  WF(process1Request) → eventually (process1 = ProcessState.Critical)
}
```

#### When to Use Weak vs Strong Fairness

- **Weak Fairness (WF)**: Use when an action should execute if it's continuously available
  - Example: A process waiting for a lock should eventually get it
  
- **Strong Fairness (SF)**: Use when an action should execute even if it's intermittently enabled
  - Example: A message that keeps getting added to a queue should eventually be processed

#### Fairness Syntax Summary

```spectre
// Weak fairness on action
WF(actionName)

// Strong fairness on action
SF(actionName)

// Weak fairness on variable (all actions modifying it)
WF(variableName)

// Strong fairness on variable
SF(variableName)

// In temporal properties
temporal propertyName {
  WF(actionName) && SF(otherAction)
}
```

---

## Examples

### Example 1: Simple Counter

```spectre
// Counter specification

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

description "Decrements the counter by one, only when positive"
action decrement {
  require counter > 0
  counter' = counter - 1
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

### Example 2: User Management System

```spectre
// User management specification

type User = {
  id: int,
  name: str,
  active: bool
}

description "Collection of all registered users"
var users: Set<User>

description "Next available user ID to assign"
var nextId: int

description "System initializes with no users and first ID set to 1"
init {
  users = {}  // Empty set literal
  nextId = 1
}

description "Adds a new user to the system with the next available ID"
action addUser(name: str) {
  users' = users.union({ { id: nextId, name: name, active: true } })
  nextId' = nextId + 1
}

description "Removes a user from the system by their ID"
action removeUser(id: int) {
  require users.exists(u => u.id = id)
  users' = users.filter(u => u.id != id)
}

description "Deactivates a user account, marking them as inactive"
action deactivateUser(id: int) {
  require users.exists(u => u.id = id && u.active)
  users' = users.map(u => 
    if (u.id = id) then { ...u, active: false } else u
  )
}

description "Validates that all user IDs are unique"
invariant uniqueIds {
  users.size() = users.map(u => u.id).toSet().size()
}

description "Ensures that nextId is always positive"
invariant nextIdPositive {
  nextId > 0
}

description "If users exist, they will eventually all be deactivated"
temporal eventuallyAllInactive {
  eventually users.forall(u => !u.active)
}
```

### Example 3: Mutex Lock

```spectre
// Mutex lock specification

enum ProcessState {
  Idle,
  Waiting,
  Critical
}

description "State of the first process"
var process1: ProcessState

description "State of the second process"
var process2: ProcessState

description "Lock flag indicating if a process is in critical section"
var lock: bool

description "System starts with both processes idle and lock free"
init {
  process1 = ProcessState.Idle
  process2 = ProcessState.Idle
  lock = false
}

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

description "Process 2 requests and acquires the lock"
action process2Request {
  require process2 = ProcessState.Idle && !lock
  process2' = ProcessState.Critical
  lock' = true
}

description "Process 2 releases the lock"
action process2Release {
  require process2 = ProcessState.Critical
  process2' = ProcessState.Idle
  lock' = false
}

description "CRITICAL: Ensures only one process can be in critical section at a time"
invariant mutualExclusion {
  !(process1 = ProcessState.Critical && process2 = ProcessState.Critical)
}

description "Ensures lock flag accurately reflects critical section state"
invariant lockConsistency {
  (process1 = ProcessState.Critical || process2 = ProcessState.Critical) = lock
}

description "Verifies that process 1 eventually enters critical section"
temporal eventuallyProcess1Critical {
  eventually (process1 = ProcessState.Critical)
}

description "Verifies that process 2 eventually enters critical section"
temporal eventuallyProcess2Critical {
  eventually (process2 = ProcessState.Critical)
}
```

### Example 4: Message Queue

```spectre
// Message queue specification

type Message = {
  id: int,
  content: str,
  priority: int
}

description "Queue of messages waiting to be processed"
var queue: List<Message>

description "Set of message IDs that have been processed"
var processed: Set<int>

description "Next available message ID to assign"
var nextMessageId: int

description "System starts with empty queue, no processed messages, and first ID set to 1"
init {
  queue = []      // Empty list literal
  processed = {}  // Empty set literal
  nextMessageId = 1
}

description "Adds a new message to the queue with given content and priority"
action enqueue(content: str, priority: int) {
  let msg = { id: nextMessageId, content: content, priority: priority }
  queue' = queue.append(msg).sortBy(m => m.priority)
  nextMessageId' = nextMessageId + 1
}

description "Removes and processes the highest priority message from the queue"
action dequeue {
  require queue.size() > 0
  let msg = queue.head()
  queue' = queue.tail()
  processed' = processed.union({ msg.id })  // Using set literal
}

description "Ensures no duplicate message IDs exist in the queue"
invariant noDuplicates {
  queue.map(m => m.id).toSet().size() = queue.size()
}

description "Ensures processed messages are not still in the queue"
invariant processedNotInQueue {
  queue.forall(m => !processed.contains(m.id))
}

description "Verifies that messages will eventually be processed"
temporal eventuallyProcessed {
  eventually processed.size() > 0
}

description "Fairness guarantee: if queue is not empty, messages will eventually be processed"
temporal fairness {
  always (queue.size() > 0 → eventually processed.size() > processed.size())
}
```

### Example 5: Bank Account

```spectre
// Bank account specification

type Account = {
  id: int,
  balance: int,
  frozen: bool
}

description "Map of account IDs to account records"
var accounts: Map<int, Account>

description "List of all transactions, each tuple contains (fromId, toId, amount)"
var transactions: List<(int, int, int)>

description "System starts with no accounts and no transactions"
init {
  accounts = Map.empty()
  transactions = []  // Empty list literal
}

description "Creates a new account with the given ID and initial balance"
action createAccount(id: int, initialBalance: int) {
  require !accounts.contains(id) && initialBalance >= 0
  accounts' = accounts.put(id, { id: id, balance: initialBalance, frozen: false })
}

description "Deposits money into an account, increasing its balance"
action deposit(accountId: int, amount: int) {
  require accounts.contains(accountId) && amount > 0
  let account = accounts.get(accountId)
  require !account.frozen
  accounts' = accounts.put(accountId, { 
    ...account, 
    balance: account.balance + amount 
  })
}

description "Withdraws money from an account, decreasing its balance"
action withdraw(accountId: int, amount: int) {
  require accounts.contains(accountId) && amount > 0
  let account = accounts.get(accountId)
  require !account.frozen && account.balance >= amount
  accounts' = accounts.put(accountId, { 
    ...account, 
    balance: account.balance - amount 
  })
}

description "Transfers money from one account to another, recording the transaction"
action transfer(fromId: int, toId: int, amount: int) {
  require accounts.contains(fromId) && accounts.contains(toId) && amount > 0
  let fromAccount = accounts.get(fromId)
  let toAccount = accounts.get(toId)
  require !fromAccount.frozen && !toAccount.frozen
  require fromAccount.balance >= amount
  
  accounts' = accounts
    .put(fromId, { ...fromAccount, balance: fromAccount.balance - amount })
    .put(toId, { ...toAccount, balance: toAccount.balance + amount })
  
  transactions' = transactions.append((fromId, toId, amount))
}

description "Freezes an account, preventing all transactions"
action freezeAccount(accountId: int) {
  require accounts.contains(accountId)
  let account = accounts.get(accountId)
  accounts' = accounts.put(accountId, { ...account, frozen: true })
}

description "CRITICAL: Ensures account balances never go negative"
invariant balanceNonNegative {
  accounts.values().forall(a => a.balance >= 0)
}

description "Ensures total balance is conserved across transfers (no money created or destroyed)"
invariant totalBalanceConserved {
  let totalBefore = accounts.values().map(a => a.balance).sum()
  // After transfer, total should remain the same
  accounts.values().map(a => a.balance).sum() = totalBefore
}

description "Verifies that accounts will eventually be created"
temporal eventuallyAccountCreated {
  eventually accounts.size() > 0
}

description "Verifies that transactions will eventually occur"
temporal eventuallyTransaction {
  eventually transactions.size() > 0
}
```

---

## Syntax Summary

### Keywords

- `module`: Module definition
- `import`: Import a module
- `extends`: Extend a module
- `public`: Public visibility modifier
- `private`: Private visibility modifier
- `const`: Constant declaration
- `var`: State variable declaration
- `description`: Human-readable description for error messages
- `init`: Initial state definition
- `oneOf`: Non-deterministic initial state selection
- `action`: Transition definition
- `fun`: Pure function definition
- `invariant`: Classic constraint (must hold in all states)
- `temporal`: Temporal property
- `require`: Precondition
- `ensure`: Postcondition
- `type`: Type definition
- `enum`: Enumeration definition
- `let`: Local variable binding
- `if`, `then`, `else`: Conditional expressions
- `return`: Return value from function
- `super`: Call parent module's action (in extended modules)
- `WF`: Weak fairness operator
- `SF`: Strong fairness operator

### Temporal Operators

- `always P`: P holds in all states
- `eventually P`: P holds in at least one future state
- `P until Q`: P holds until Q becomes true
- `P → Q`: P leads to Q (if P becomes true, Q will eventually become true). This is syntactic sugar for `always (P → eventually Q)` and can only be used within temporal properties.
- `next P`: P holds in the next state
- `WF(action)`: Weak fairness - action executes if continuously enabled
- `SF(action)`: Strong fairness - action executes if enabled infinitely often
- `WF(variable)`: Weak fairness on all actions modifying the variable
- `SF(variable)`: Strong fairness on all actions modifying the variable

### Set Operations

**Set Literals:**
- `{ value1, value2, ... }`: Create a set with elements
- `{}`: Empty set literal

**Static Constructors:**
- `Set.empty()`: Empty set (alternative to `{}`)
- `Set.of(x)`: Singleton set containing one element (alternative to `{ x }`)

**Set Methods:**
- `s.union(t)`: Union of sets
- `s.intersection(t)`: Intersection of sets
- `s.contains(x)`: Membership test
- `s.size()`: Set size
- `s.forall(predicate)`: Universal quantification
- `s.exists(predicate)`: Existential quantification
- `s.filter(predicate)`: Filter elements
- `s.map(fn)`: Map function over elements
- `s.toList()`: Convert set to list

**Examples:**
```spectre
// Using literals
const names = { "alice", "bob", "charlie" }
const empty = {}

// Using static methods (equivalent to literals, but verbose)
// Note: Literal syntax { "alice", "bob", "charlie" } is preferred
const namesAlt = { "alice", "bob", "charlie" }  // Literal syntax is cleaner
const emptyAlt = {}  // Or Set.empty() - both are equivalent

// Set operations
const all = names.union({ "dave", "eve" })
const filtered = names.filter(name => name.length() > 4)
```

### List Operations

**List Literals:**
- `[ value1, value2, ... ]`: Create a list with elements
- `[]`: Empty list literal

**Static Constructors:**
- `List.empty()`: Empty list (alternative to `[]`)
- `List.of(x)`: Singleton list containing one element (alternative to `[ x ]`)

**List Methods:**
- `l.append(x)`: Append element (returns new list)
- `l.head()`: First element
- `l.tail()`: List without first element (returns new list)
- `l.size()`: List size
- `l.filter(predicate)`: Filter elements
- `l.map(fn)`: Map function over elements
- `l.reduce(initial, fn)`: Reduce to single value
- `l.forall(predicate)`: Universal quantification
- `l.exists(predicate)`: Existential quantification
- `l.toSet()`: Convert list to set

**Examples:**
```spectre
// Using literals
const numbers = [ 1, 2, 3, 4, 5 ]
const empty = []

// Using static methods (equivalent to literals, but verbose)
// Note: Literal syntax [ 1, 2 ] is preferred
const numbersAlt = [ 1, 2 ]  // Literal syntax is cleaner
const emptyAlt = []  // Or List.empty() - both are equivalent

// List operations
const doubled = numbers.map(x => x * 2)
const sum = numbers.reduce(0, (acc, x) => acc + x)
const first = numbers.head()
const rest = numbers.tail()
```

### Map Operations

- `Map.empty()`: Empty map
- `m.put(k, v)`: Add/update key-value pair
- `m.get(k)`: Get value for key
- `m.contains(k)`: Check if key exists
- `m.keys()`: Set of keys
- `m.values()`: List of values

---

## Verification

The Spectre verifier checks:

1. **Type Safety**: All expressions are well-typed
2. **Invariants**: All invariants hold in all reachable states
3. **Temporal Properties**: All temporal properties hold for all execution traces
4. **Preconditions**: All preconditions are satisfied before transitions execute
5. **Postconditions**: All postconditions are satisfied after transitions execute
6. **Initial States**: When using `oneOf`, all specified initial states are explored

### Verification Modes

- **Bounded Model Checking**: Check properties up to a certain depth
- **Unbounded Model Checking**: Check properties for all possible executions
- **Liveness Checking**: Verify temporal properties (eventually, always, etc.)
- **Initial State Exploration**: When `oneOf` is used, verify properties starting from each initial state

### Error Reporting

When verification fails, the verifier provides:
- **Descriptive Error Messages**: Using `description` fields to explain what failed
- **Execution Traces**: Full stack traces showing the sequence of actions leading to the error
- **State Snapshots**: Complete state information at each step
- **Context Information**: Which invariants, temporal properties, or preconditions were violated

---

## Notes

- The prime notation (`'`) denotes the next state value
- State variables without primes refer to the current state
- Actions can be non-deterministic (multiple possible next states)
- The verifier explores all possible execution paths

