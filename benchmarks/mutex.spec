// Benchmark: two-process mutual exclusion via a ticket lock.
// ~9 states; tests multi-variable interaction.
enum ProcState { Idle, Waiting, Critical }

var p1: ProcState
var p2: ProcState
var lock_held: bool

init { p1 = ProcState.Idle  p2 = ProcState.Idle  lock_held = false }

action p1_request { require p1 == ProcState.Idle    p1' = ProcState.Waiting  p2' = p2  lock_held' = lock_held }
action p1_acquire { require p1 == ProcState.Waiting && !lock_held  p1' = ProcState.Critical  p2' = p2  lock_held' = true }
action p1_release { require p1 == ProcState.Critical  p1' = ProcState.Idle  p2' = p2  lock_held' = false }

action p2_request { require p2 == ProcState.Idle    p1' = p1  p2' = ProcState.Waiting  lock_held' = lock_held }
action p2_acquire { require p2 == ProcState.Waiting && !lock_held  p1' = p1  p2' = ProcState.Critical  lock_held' = true }
action p2_release { require p2 == ProcState.Critical  p1' = p1  p2' = ProcState.Idle  lock_held' = false }

invariant mutual_exclusion {
  !(p1 == ProcState.Critical && p2 == ProcState.Critical)
}
