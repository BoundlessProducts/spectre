// Concurrent System with Lock - Violation Version
// Demonstrates concurrent processes accessing a shared resource

enum ProcessState {
  Idle,
  Waiting,
  Critical
}

description "State of process 1"
var process1: ProcessState

description "State of process 2"
var process2: ProcessState

description "State of process 3"
var process3: ProcessState

description "Lock holder (0 = unlocked, 1-3 = held by process 1-3)"
var lock: int

description "Resource value being accessed"
var resource: int

description "Initial state: all processes idle, lock free, resource at 0"
init {
  process1 = ProcessState.Idle
  process2 = ProcessState.Idle
  process3 = ProcessState.Idle
  lock = 0
  resource = 0
}

description "Process 1 requests lock"
action requestLock1 {
  process1' = ProcessState.Waiting
}

description "Process 1 acquires lock and enters critical section"
description "PROBLEM: No check if lock is available"
action acquireLock1 {
  // Missing: require lock == 0
  lock' = 1
  process1' = ProcessState.Critical
}

description "Process 1 accesses resource in critical section"
description "PROBLEM: No check if process holds the lock"
action accessResource1 {
  // Missing: require lock == 1 && process1 == ProcessState.Critical
  resource' = resource + 1
}

description "Process 1 releases lock and exits critical section"
description "PROBLEM: No check if process holds the lock"
action releaseLock1 {
  // Missing: require lock == 1 && process1 == ProcessState.Critical
  lock' = 0
  process1' = ProcessState.Idle
}

description "Process 2 requests lock"
action requestLock2 {
  process2' = ProcessState.Waiting
}

description "Process 2 acquires lock and enters critical section"
description "PROBLEM: No check if lock is available"
action acquireLock2 {
  // Missing: require lock == 0
  lock' = 2
  process2' = ProcessState.Critical
}

description "Process 2 accesses resource in critical section"
description "PROBLEM: No check if process holds the lock"
action accessResource2 {
  // Missing: require lock == 2 && process2 == ProcessState.Critical
  resource' = resource + 1
}

description "Process 2 releases lock and exits critical section"
description "PROBLEM: No check if process holds the lock"
action releaseLock2 {
  // Missing: require lock == 2 && process2 == ProcessState.Critical
  lock' = 0
  process2' = ProcessState.Idle
}

description "Process 3 requests lock"
action requestLock3 {
  process3' = ProcessState.Waiting
}

description "Process 3 acquires lock and enters critical section"
description "PROBLEM: No check if lock is available"
action acquireLock3 {
  // Missing: require lock == 0
  lock' = 3
  process3' = ProcessState.Critical
}

description "Process 3 accesses resource in critical section"
description "PROBLEM: No check if process holds the lock"
action accessResource3 {
  // Missing: require lock == 3 && process3 == ProcessState.Critical
  resource' = resource + 1
}

description "Process 3 releases lock and exits critical section"
description "PROBLEM: No check if process holds the lock"
action releaseLock3 {
  // Missing: require lock == 3 && process3 == ProcessState.Critical
  lock' = 0
  process3' = ProcessState.Idle
}

description "Invariant: Only one process can hold the lock at a time"
invariant mutualExclusion {
  (lock == 0) || (lock == 1 && !(process2 == ProcessState.Critical) && !(process3 == ProcessState.Critical)) ||
  (lock == 2 && !(process1 == ProcessState.Critical) && !(process3 == ProcessState.Critical)) ||
  (lock == 3 && !(process1 == ProcessState.Critical) && !(process2 == ProcessState.Critical))
}

description "Invariant: Process in critical section must hold the lock"
invariant criticalSectionHoldsLock {
  (!(process1 == ProcessState.Critical) || lock == 1) &&
  (!(process2 == ProcessState.Critical) || lock == 2) &&
  (!(process3 == ProcessState.Critical) || lock == 3)
}

description "Temporal: If a process requests the lock, it eventually gets it"
temporal progress {
  always ((process1 == ProcessState.Waiting) → eventually (process1 == ProcessState.Critical))
}

description "Temporal: The resource value eventually increases"
temporal resourceProgress {
  always eventually resource > 0
}

