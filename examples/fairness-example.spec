// Fairness constraints example.
// Shows the difference between weak fairness (WF) and strong fairness (SF)
// on a simple process scheduler.

description "Current state of the process (0=idle, 1=ready, 2=running)"
var processState: int

description "Number of completed work units"
var completedWork: int

description "Total scheduler cycles elapsed"
var schedulerCycles: int

description "Process starts idle with no completed work"
init {
  processState = 0
  completedWork = 0
  schedulerCycles = 0
}

description "Scheduler ticks advance the cycle count"
action schedulerTick {
  schedulerCycles' = schedulerCycles + 1
}

description "The process transitions from idle to ready"
action becomeReady {
  require processState == 0
  processState' = 1
}

description "The scheduler selects the ready process to run"
action schedule {
  require processState == 1
  processState' = 2
}

description "The running process completes a work unit and returns to idle"
action completeWork {
  require processState == 2
  processState' = 0
  completedWork' = completedWork + 1
}

description "Process state is always a valid value"
invariant validState {
  processState >= 0 && processState <= 2
}

description "Completed work never decreases"
invariant workNeverDecreases {
  completedWork >= 0
}

description "If the process is continuously ready, it will eventually be scheduled"
temporal weakFairnessSchedule {
  WF(schedule)
}

description "If the scheduler tick is enabled infinitely often, it will fire infinitely often"
temporal weakFairnessSchedulerTick {
  WF(schedulerTick)
}

description "Strong fairness: if the process is ready even intermittently, it will eventually run"
temporal strongFairnessSchedule {
  SF(schedule)
}

description "Strong fairness on work completion: if completion is enabled even intermittently, it will happen"
temporal strongFairnessComplete {
  SF(completeWork)
}

description "The process can eventually complete work"
temporal eventuallyCompletes {
  eventually (completedWork >= 1)
}
