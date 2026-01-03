// Error Trace Example with Descriptions
// Demonstrates how descriptions create readable error messages and stack traces

enum ProcessState {
  Idle,
  Critical
}

description "State of the first process (Idle or Critical)"
var process1: ProcessState

description "State of the second process (Idle or Critical)"
var process2: ProcessState

description "Lock flag indicating if a process is in critical section"
var lock: bool

description "System starts with both processes idle and lock free"
init {
  process1 = ProcessState.Idle
  process2 = ProcessState.Idle
  lock = false
}

description "Process 1 requests and acquires the lock, entering critical section"
action process1Request {
  require process1 == ProcessState.Idle && !lock
  process1' = ProcessState.Critical
  lock' = true
}

description "Process 1 releases the lock and returns to idle state"
action process1Release {
  require process1 == ProcessState.Critical
  process1' = ProcessState.Idle
  lock' = false
}

description "Process 2 requests and acquires the lock, entering critical section"
action process2Request {
  require process2 == ProcessState.Idle && !lock
  process2' = ProcessState.Critical
  lock' = true
}

description "Process 2 releases the lock and returns to idle state"
action process2Release {
  require process2 == ProcessState.Critical
  process2' = ProcessState.Idle
  lock' = false
}

description "CRITICAL: Ensures only one process can be in critical section at a time"
invariant mutualExclusion {
  !(process1 == ProcessState.Critical && process2 == ProcessState.Critical)
}

description "Ensures lock flag accurately reflects whether any process is critical"
invariant lockConsistency {
  (process1 == ProcessState.Critical || process2 == ProcessState.Critical) == lock
}

description "Verifies that process 1 eventually enters critical section"
temporal eventuallyProcess1Critical {
  eventually (process1 == ProcessState.Critical)
}

description "Verifies that process 2 eventually enters critical section"
temporal eventuallyProcess2Critical {
  eventually (process2 == ProcessState.Critical)
}

