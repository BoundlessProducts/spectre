# Chapter 5: Concurrent Systems and Locking

This chapter demonstrates how to specify and verify concurrent systems where multiple processes compete for access to a shared resource. We'll use a locking mechanism to ensure mutual exclusion and show how missing preconditions can lead to both invariant and temporal property violations.

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

