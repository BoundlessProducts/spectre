// Repair benchmark: enum assignment — transition to Running only from Idle.

enum Status { Idle, Running, Done }

var status: Status
var result: int

init {
  status = Status.Idle
  result = 0
}

description "Start: transition to Running (missing guard — should require Idle)"
action start {
  status' = Status.Running
}

description "Compute: produce result while Running"
action compute {
  require status == Status.Running
  require result < 5
  result' = result + 1
}

description "Finish: mark Done"
action finish {
  require status == Status.Running
  status' = Status.Done
}

description "A job transitions to Running only from Idle"
invariant idleBeforeRun {
  !(status == Status.Running && result == 0)
}
