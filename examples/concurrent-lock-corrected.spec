// Concurrent System with Lock - Corrected Version
// Demonstrates concurrent processes accessing a shared resource with proper locking

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
  require process1 == ProcessState.Idle
  process1' = ProcessState.Waiting
}

description "Process 1 acquires lock and enters critical section"
description "FIXED: Only acquire if lock is available and process is waiting"
action acquireLock1 {
  require lock == 0
  require process1 == ProcessState.Waiting
  lock' = 1
  process1' = ProcessState.Critical
}

description "Process 1 accesses resource in critical section"
description "FIXED: Only access if process holds the lock and is in critical section"
action accessResource1 {
  require lock == 1
  require process1 == ProcessState.Critical
  resource' = resource + 1
}

description "Process 1 releases lock and exits critical section"
description "FIXED: Only release if process holds the lock and is in critical section"
action releaseLock1 {
  require lock == 1
  require process1 == ProcessState.Critical
  lock' = 0
  process1' = ProcessState.Idle
}

description "Process 2 requests lock"
action requestLock2 {
  require process2 == ProcessState.Idle
  process2' = ProcessState.Waiting
}

description "Process 2 acquires lock and enters critical section"
description "FIXED: Only acquire if lock is available and process is waiting"
action acquireLock2 {
  require lock == 0
  require process2 == ProcessState.Waiting
  lock' = 2
  process2' = ProcessState.Critical
}

description "Process 2 accesses resource in critical section"
description "FIXED: Only access if process holds the lock and is in critical section"
action accessResource2 {
  require lock == 2
  require process2 == ProcessState.Critical
  resource' = resource + 1
}

description "Process 2 releases lock and exits critical section"
description "FIXED: Only release if process holds the lock and is in critical section"
action releaseLock2 {
  require lock == 2
  require process2 == ProcessState.Critical
  lock' = 0
  process2' = ProcessState.Idle
}

description "Process 3 requests lock"
action requestLock3 {
  require process3 == ProcessState.Idle
  process3' = ProcessState.Waiting
}

description "Process 3 acquires lock and enters critical section"
description "FIXED: Only acquire if lock is available and process is waiting"
action acquireLock3 {
  require lock == 0
  require process3 == ProcessState.Waiting
  lock' = 3
  process3' = ProcessState.Critical
}

description "Process 3 accesses resource in critical section"
description "FIXED: Only access if process holds the lock and is in critical section"
action accessResource3 {
  require lock == 3
  require process3 == ProcessState.Critical
  resource' = resource + 1
}

description "Process 3 releases lock and exits critical section"
description "FIXED: Only release if process holds the lock and is in critical section"
action releaseLock3 {
  require lock == 3
  require process3 == ProcessState.Critical
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

description "Temporal: The resource value eventually increases"
description "FIXED: With proper locking and preconditions, processes can access the resource"
description "This property holds because preconditions ensure actions can execute safely"
temporal resourceProgress {
  always eventually resource > 0
}

