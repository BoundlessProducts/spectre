// Mutual exclusion example with two competing processes.
// Models a Peterson-style mutex: each process sets a "want" flag before
// entering its critical section and yields to the other if both want in.

description "Whether process 1 wants to enter the critical section"
var want1: bool

description "Whether process 2 wants to enter the critical section"
var want2: bool

description "Which process has priority when both want in (1 or 2)"
var turn: int

description "Whether process 1 is in the critical section"
var inCritical1: bool

description "Whether process 2 is in the critical section"
var inCritical2: bool

description "Both processes start outside the critical section"
init {
  want1 = false
  want2 = false
  turn = 1
  inCritical1 = false
  inCritical2 = false
}

description "Process 1 signals intent to enter the critical section"
action request1 {
  require want1 == false
  want1' = true
  turn' = 2
}

description "Process 1 enters the critical section when it is safe to do so"
action enter1 {
  require want1 == true
  require inCritical1 == false
  require want2 == false || turn == 1
  inCritical1' = true
}

description "Process 1 exits the critical section"
action exit1 {
  require inCritical1 == true
  inCritical1' = false
  want1' = false
}

description "Process 2 signals intent to enter the critical section"
action request2 {
  require want2 == false
  want2' = true
  turn' = 1
}

description "Process 2 enters the critical section when it is safe to do so"
action enter2 {
  require want2 == true
  require inCritical2 == false
  require want1 == false || turn == 2
  inCritical2' = true
}

description "Process 2 exits the critical section"
action exit2 {
  require inCritical2 == true
  inCritical2' = false
  want2' = false
}

description "At most one process is in the critical section at any time"
invariant mutualExclusion {
  inCritical1 == false || inCritical2 == false
}

description "A process is only in the critical section if it wanted to enter"
invariant lockConsistency {
  (inCritical1 == false || want1 == true) &&
  (inCritical2 == false || want2 == true)
}

description "Process 1 can eventually enter the critical section"
temporal eventuallyProcess1Critical {
  eventually (inCritical1 == true)
}

description "Process 2 can eventually enter the critical section"
temporal eventuallyProcess2Critical {
  eventually (inCritical2 == true)
}

description "If process 1 keeps wanting in, it will eventually enter"
temporal fairnessProcess1 {
  WF(enter1)
}

description "If process 2 keeps wanting in, it will eventually enter"
temporal fairnessProcess2 {
  WF(enter2)
}

description "Any process that enters the critical section will eventually release it"
temporal eventuallyRelease {
  always (inCritical1 == true -> eventually inCritical1 == false)
}
