// Fairness Example
// Demonstrates weak and strong fairness conditions for concurrent systems

enum ProcessState {
  Idle,
  Waiting,
  Critical
}

description "Number of processes in the system"
const NUM_PROCESSES: int = 2

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

// Weak Fairness: If process1Request is continuously enabled, it will execute
description "Weak fairness ensures process1 gets fair access when continuously waiting"
temporal process1WeakFairness {
  WF(process1Request)
}

description "Weak fairness ensures process2 gets fair access when continuously waiting"
temporal process2WeakFairness {
  WF(process2Request)
}

// Strong Fairness: If process1Request is enabled infinitely often, it will execute
description "Strong fairness ensures process1 executes even if intermittently enabled"
temporal process1StrongFairness {
  SF(process1Request)
}

description "Strong fairness ensures process2 executes even if intermittently enabled"
temporal process2StrongFairness {
  SF(process2Request)
}

// Fairness on variables: All actions modifying the variable get fairness
description "Weak fairness for all actions modifying process1 state"
temporal process1VariableFairness {
  WF(process1)
}

description "Weak fairness for all actions modifying process2 state"
temporal process2VariableFairness {
  WF(process2)
}

// Liveness properties that depend on fairness
description "With weak fairness, process1 will eventually enter critical section"
temporal eventuallyProcess1Critical {
  WF(process1Request) → eventually (process1 = ProcessState.Critical)
}

description "With weak fairness, process2 will eventually enter critical section"
temporal eventuallyProcess2Critical {
  WF(process2Request) → eventually (process2 = ProcessState.Critical)
}

description "With fairness, processes will eventually release the lock"
temporal eventuallyRelease {
  always (process1 = ProcessState.Critical → eventually process1 = ProcessState.Idle)
}

description "Combined fairness guarantees both processes get fair access"
temporal combinedFairness {
  WF(process1Request) && WF(process2Request)
}

