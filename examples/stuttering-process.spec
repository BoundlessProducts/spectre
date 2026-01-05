enum ProcessState {
  Idle,
  Running
}

module StutteringProcess {
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
  
  description "Process 1 does nothing - causes stuttering"
  action process1Noop {
    process1' = process1
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
  
  description "Process 2 does nothing - causes stuttering"
  action process2Noop {
    process2' = process2
  }
  
  description "Verifies process1 eventually runs"
  temporal process1EventuallyRuns {
    eventually (process1 = ProcessState.Running)
  }
}
