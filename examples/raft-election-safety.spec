// Raft Leader Election + Log Replication Safety — 3-node cluster
//
// This specification models a 3-node Raft cluster at the leader-election
// and log-replication level.  It verifies the core Raft safety properties:
//
//   1. electionSafety    — at most one leader per Raft term
//   2. leaderMajority    — every leader holds a strict majority of votes
//   3. candidateSelfVote — every Candidate has at least one vote (its own)
//
// Compared to a single-round election model, this spec adds:
//   • Per-node term numbers (bounded 0–3) enabling per-term election safety
//   • Log-length tracking and leader append / follower replication actions
//   • Step-down actions: a leader or candidate steps down when it observes
//     a higher term on another node
//   • Per-term vote discipline: a node may vote for at most one candidate
//     per term, and only for a candidate whose term is at least as recent
//
// Variable layout: five variables per node (state, term, voted_for, votes, log_len)
// for a total of 15 state variables.  Actions total 27.
//
// Per-node spec skeleton was mined from examples/rust/raft.rs via
// `spectre mine --lang rust examples/rust/raft.rs`; cluster composition and
// invariants were added manually.

enum NodeState { Follower, Candidate, Leader }

// ── Node 1 ───────────────────────────────────────────────────────────────────
description "Role of node 1"
var state1:      NodeState
description "Current Raft term for node 1 (bounded 0–3)"
var term1:       int
description "Node that node 1 voted for this term (0 = no vote, 1–3 = node ID)"
var voted_for1:  int
description "Votes collected by node 1 as Candidate"
var votes1:      int
description "Log length of node 1 (bounded 0–3)"
var log_len1:    int

// ── Node 2 ───────────────────────────────────────────────────────────────────
description "Role of node 2"
var state2:      NodeState
description "Current Raft term for node 2 (bounded 0–3)"
var term2:       int
description "Node that node 2 voted for this term (0 = no vote, 1–3 = node ID)"
var voted_for2:  int
description "Votes collected by node 2 as Candidate"
var votes2:      int
description "Log length of node 2 (bounded 0–3)"
var log_len2:    int

// ── Node 3 ───────────────────────────────────────────────────────────────────
description "Role of node 3"
var state3:      NodeState
description "Current Raft term for node 3 (bounded 0–3)"
var term3:       int
description "Node that node 3 voted for this term (0 = no vote, 1–3 = node ID)"
var voted_for3:  int
description "Votes collected by node 3 as Candidate"
var votes3:      int
description "Log length of node 3 (bounded 0–3)"
var log_len3:    int

init {
  state1 = NodeState.Follower  state2 = NodeState.Follower  state3 = NodeState.Follower
  term1 = 0   term2 = 0   term3 = 0
  voted_for1 = 0  voted_for2 = 0  voted_for3 = 0
  votes1 = 0   votes2 = 0   votes3 = 0
  log_len1 = 0  log_len2 = 0  log_len3 = 0
}

// ── Election initiation ───────────────────────────────────────────────────────
// A Follower or Candidate (not a Leader) begins a new election: bumps its term,
// transitions to Candidate, and casts a self-vote.

description "Node 1 starts a new election: bumps term, becomes Candidate, self-votes"
action startElection1 {
  require state1 != NodeState.Leader
  require term1 < 3
  state1'     = NodeState.Candidate
  term1'      = term1 + 1
  voted_for1' = 0
  votes1'     = 1
}

description "Node 2 starts a new election"
action startElection2 {
  require state2 != NodeState.Leader
  require term2 < 3
  state2'     = NodeState.Candidate
  term2'      = term2 + 1
  voted_for2' = 0
  votes2'     = 1
}

description "Node 3 starts a new election"
action startElection3 {
  require state3 != NodeState.Leader
  require term3 < 3
  state3'     = NodeState.Candidate
  term3'      = term3 + 1
  voted_for3' = 0
  votes3'     = 1
}

// ── Vote granting ─────────────────────────────────────────────────────────────
// A Follower grants a vote to a Candidate when:
//   (a) the candidate's term is strictly higher, OR
//       the candidate's term equals the voter's and the voter has not yet voted
//   (b) the voter has not already cast a vote for a different candidate

description "Node 2 (Follower) grants its vote to Candidate node 1"
action vote2for1 {
  require state1 == NodeState.Candidate
  require state2 == NodeState.Follower
  require votes1 < 3
  require term1 > term2 || (term1 == term2 && voted_for2 == 0)
  require voted_for2 == 0 || voted_for2 == 1
  voted_for2' = 1
  term2'      = term1
  votes1'     = votes1 + 1
}

description "Node 3 (Follower) grants its vote to Candidate node 1"
action vote3for1 {
  require state1 == NodeState.Candidate
  require state3 == NodeState.Follower
  require votes1 < 3
  require term1 > term3 || (term1 == term3 && voted_for3 == 0)
  require voted_for3 == 0 || voted_for3 == 1
  voted_for3' = 1
  term3'      = term1
  votes1'     = votes1 + 1
}

description "Node 1 (Follower) grants its vote to Candidate node 2"
action vote1for2 {
  require state2 == NodeState.Candidate
  require state1 == NodeState.Follower
  require votes2 < 3
  require term2 > term1 || (term2 == term1 && voted_for1 == 0)
  require voted_for1 == 0 || voted_for1 == 2
  voted_for1' = 2
  term1'      = term2
  votes2'     = votes2 + 1
}

description "Node 3 (Follower) grants its vote to Candidate node 2"
action vote3for2 {
  require state2 == NodeState.Candidate
  require state3 == NodeState.Follower
  require votes2 < 3
  require term2 > term3 || (term2 == term3 && voted_for3 == 0)
  require voted_for3 == 0 || voted_for3 == 2
  voted_for3' = 2
  term3'      = term2
  votes2'     = votes2 + 1
}

description "Node 1 (Follower) grants its vote to Candidate node 3"
action vote1for3 {
  require state3 == NodeState.Candidate
  require state1 == NodeState.Follower
  require votes3 < 3
  require term3 > term1 || (term3 == term1 && voted_for1 == 0)
  require voted_for1 == 0 || voted_for1 == 3
  voted_for1' = 3
  term1'      = term3
  votes3'     = votes3 + 1
}

description "Node 2 (Follower) grants its vote to Candidate node 3"
action vote2for3 {
  require state3 == NodeState.Candidate
  require state2 == NodeState.Follower
  require votes3 < 3
  require term3 > term2 || (term3 == term2 && voted_for2 == 0)
  require voted_for2 == 0 || voted_for2 == 3
  voted_for2' = 3
  term2'      = term3
  votes3'     = votes3 + 1
}

// ── Leader promotion ──────────────────────────────────────────────────────────
// A Candidate with a strict majority of votes (> cluster_size / 2 = 1.5, i.e. >= 2)
// transitions to Leader.

description "Node 1 becomes Leader once it holds a strict majority (≥ 2 of 3 votes)"
action becomeLeader1 {
  require state1 == NodeState.Candidate
  require votes1 * 2 > 3
  state1' = NodeState.Leader
}

description "Node 2 becomes Leader"
action becomeLeader2 {
  require state2 == NodeState.Candidate
  require votes2 * 2 > 3
  state2' = NodeState.Leader
}

description "Node 3 becomes Leader"
action becomeLeader3 {
  require state3 == NodeState.Candidate
  require votes3 * 2 > 3
  state3' = NodeState.Leader
}

// ── Log replication: leader appends ──────────────────────────────────────────
// Only the Leader may append entries to its log.

description "Leader node 1 appends one new log entry"
action appendEntry1 {
  require state1 == NodeState.Leader
  require log_len1 < 3
  log_len1' = log_len1 + 1
}

description "Leader node 2 appends one new log entry"
action appendEntry2 {
  require state2 == NodeState.Leader
  require log_len2 < 3
  log_len2' = log_len2 + 1
}

description "Leader node 3 appends one new log entry"
action appendEntry3 {
  require state3 == NodeState.Leader
  require log_len3 < 3
  log_len3' = log_len3 + 1
}

// ── Log replication: follower accepts one entry from leader ───────────────────
// A Follower in the same term as the Leader accepts one replicated entry when
// its log is shorter than the leader's.

description "Leader node 1 replicates one entry to Follower node 2"
action replicate1to2 {
  require state1 == NodeState.Leader
  require state2 == NodeState.Follower
  require term2 == term1
  require log_len2 < log_len1
  log_len2' = log_len2 + 1
}

description "Leader node 1 replicates one entry to Follower node 3"
action replicate1to3 {
  require state1 == NodeState.Leader
  require state3 == NodeState.Follower
  require term3 == term1
  require log_len3 < log_len1
  log_len3' = log_len3 + 1
}

description "Leader node 2 replicates one entry to Follower node 1"
action replicate2to1 {
  require state2 == NodeState.Leader
  require state1 == NodeState.Follower
  require term1 == term2
  require log_len1 < log_len2
  log_len1' = log_len1 + 1
}

description "Leader node 2 replicates one entry to Follower node 3"
action replicate2to3 {
  require state2 == NodeState.Leader
  require state3 == NodeState.Follower
  require term3 == term2
  require log_len3 < log_len2
  log_len3' = log_len3 + 1
}

description "Leader node 3 replicates one entry to Follower node 1"
action replicate3to1 {
  require state3 == NodeState.Leader
  require state1 == NodeState.Follower
  require term1 == term3
  require log_len1 < log_len3
  log_len1' = log_len1 + 1
}

description "Leader node 3 replicates one entry to Follower node 2"
action replicate3to2 {
  require state3 == NodeState.Leader
  require state2 == NodeState.Follower
  require term2 == term3
  require log_len2 < log_len3
  log_len2' = log_len2 + 1
}

// ── Step-down ─────────────────────────────────────────────────────────────────
// A Leader or Candidate steps down to Follower when it observes that another
// node is in a higher term (simulating receipt of a message with a higher term).
// The stepping-down node adopts the higher term and resets its vote.

description "Node 1 steps down upon observing node 2 in a higher term"
action stepDown1_sees2 {
  require state1 != NodeState.Follower
  require term2 > term1
  state1'     = NodeState.Follower
  term1'      = term2
  voted_for1' = 0
  votes1'     = 0
}

description "Node 1 steps down upon observing node 3 in a higher term"
action stepDown1_sees3 {
  require state1 != NodeState.Follower
  require term3 > term1
  state1'     = NodeState.Follower
  term1'      = term3
  voted_for1' = 0
  votes1'     = 0
}

description "Node 2 steps down upon observing node 1 in a higher term"
action stepDown2_sees1 {
  require state2 != NodeState.Follower
  require term1 > term2
  state2'     = NodeState.Follower
  term2'      = term1
  voted_for2' = 0
  votes2'     = 0
}

description "Node 2 steps down upon observing node 3 in a higher term"
action stepDown2_sees3 {
  require state2 != NodeState.Follower
  require term3 > term2
  state2'     = NodeState.Follower
  term2'      = term3
  voted_for2' = 0
  votes2'     = 0
}

description "Node 3 steps down upon observing node 1 in a higher term"
action stepDown3_sees1 {
  require state3 != NodeState.Follower
  require term1 > term3
  state3'     = NodeState.Follower
  term3'      = term1
  voted_for3' = 0
  votes3'     = 0
}

description "Node 3 steps down upon observing node 2 in a higher term"
action stepDown3_sees2 {
  require state3 != NodeState.Follower
  require term2 > term3
  state3'     = NodeState.Follower
  term3'      = term2
  voted_for3' = 0
  votes3'     = 0
}

// ── Safety invariants ─────────────────────────────────────────────────────────

description "Election safety: at most one Leader per Raft term (the core Raft guarantee)"
invariant electionSafety {
  !(state1 == NodeState.Leader && state2 == NodeState.Leader && term1 == term2) &&
  !(state1 == NodeState.Leader && state3 == NodeState.Leader && term1 == term3) &&
  !(state2 == NodeState.Leader && state3 == NodeState.Leader && term2 == term3)
}

description "Leader legitimacy: every Leader holds a strict majority of votes"
invariant leaderMajority {
  (!(state1 == NodeState.Leader) || votes1 * 2 > 3) &&
  (!(state2 == NodeState.Leader) || votes2 * 2 > 3) &&
  (!(state3 == NodeState.Leader) || votes3 * 2 > 3)
}

description "Candidate self-vote: every Candidate has cast at least its own vote"
invariant candidateSelfVote {
  (!(state1 == NodeState.Candidate) || votes1 >= 1) &&
  (!(state2 == NodeState.Candidate) || votes2 >= 1) &&
  (!(state3 == NodeState.Candidate) || votes3 >= 1)
}
