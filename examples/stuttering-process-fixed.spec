enum ProcessState {
  Idle,
  Running
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
    process1' = ProcessState.Idle
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
