// Two-Phase Commit Protocol
//
// Models a coordinator and 3 participant nodes running 2PC.
// Demonstrates Set<int> and Map collection variables, the `in`
// operator, and distributed safety invariants.
//
// Run: spectre verify examples/two_phase_commit.spec

const NUM_NODES: int = 3

enum CoordPhase { Preparing, Committed, Aborted }

// coordinator phase
var coord: CoordPhase

// per-node vote tracking
var voted_yes: Set<int>   // nodes that voted yes
var voted_no:  Set<int>   // nodes that voted no

// per-node decision tracking
var committed: Set<int>   // nodes that have applied commit
var aborted:   Set<int>   // nodes that have applied abort

init {
  coord      = CoordPhase.Preparing
  voted_yes  = {}
  voted_no   = {}
  committed  = {}
  aborted    = {}
}

// --- Voting phase ---

action node1_vote_yes {
  require coord == CoordPhase.Preparing
  require !(1 in voted_yes) && !(1 in voted_no)
  voted_yes' = voted_yes.union({1})
}

action node2_vote_yes {
  require coord == CoordPhase.Preparing
  require !(2 in voted_yes) && !(2 in voted_no)
  voted_yes' = voted_yes.union({2})
}

action node3_vote_yes {
  require coord == CoordPhase.Preparing
  require !(3 in voted_yes) && !(3 in voted_no)
  voted_yes' = voted_yes.union({3})
}

action node1_vote_no {
  require coord == CoordPhase.Preparing
  require !(1 in voted_yes) && !(1 in voted_no)
  voted_no' = voted_no.union({1})
}

action node2_vote_no {
  require coord == CoordPhase.Preparing
  require !(2 in voted_yes) && !(2 in voted_no)
  voted_no' = voted_no.union({2})
}

action node3_vote_no {
  require coord == CoordPhase.Preparing
  require !(3 in voted_yes) && !(3 in voted_no)
  voted_no' = voted_no.union({3})
}

// --- Coordinator decision ---

action coordinator_commit {
  // commit only when all NUM_NODES nodes voted yes
  require coord == CoordPhase.Preparing
  require voted_yes.size() == NUM_NODES
  require voted_no.size() == 0
  coord' = CoordPhase.Committed
}

action coordinator_abort {
  // abort as soon as any node votes no
  require coord == CoordPhase.Preparing
  require voted_no.size() > 0
  coord' = CoordPhase.Aborted
}

// --- Node apply phase ---

action node1_commit {
  require coord == CoordPhase.Committed
  require !(1 in committed)
  committed' = committed.union({1})
}

action node2_commit {
  require coord == CoordPhase.Committed
  require !(2 in committed)
  committed' = committed.union({2})
}

action node3_commit {
  require coord == CoordPhase.Committed
  require !(3 in committed)
  committed' = committed.union({3})
}

action node1_abort {
  require coord == CoordPhase.Aborted
  require !(1 in aborted)
  aborted' = aborted.union({1})
}

action node2_abort {
  require coord == CoordPhase.Aborted
  require !(2 in aborted)
  aborted' = aborted.union({2})
}

action node3_abort {
  require coord == CoordPhase.Aborted
  require !(3 in aborted)
  aborted' = aborted.union({3})
}

// --- Safety invariants ---

// A node never both commits and aborts (atomicity)
invariant atomicity {
  committed.intersection(aborted).size() == 0
}

// A node only commits if the coordinator committed
invariant commit_requires_coord_committed {
  committed.size() == 0 || coord == CoordPhase.Committed
}

// A node only aborts if the coordinator aborted
invariant abort_requires_coord_aborted {
  aborted.size() == 0 || coord == CoordPhase.Aborted
}

// The coordinator only commits if all nodes voted yes
invariant unanimous_commit {
  coord != CoordPhase.Committed || voted_yes.size() == NUM_NODES
}

// No node votes both yes and no
invariant votes_disjoint {
  voted_yes.intersection(voted_no).size() == 0
}
