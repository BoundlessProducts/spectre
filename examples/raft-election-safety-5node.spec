// Raft Leader Election + Log Replication Safety — 5-node cluster
// Generated from the 3-node spec template.
// BFS state space exceeds practical limits for N>3; use spectre simulate.

enum NodeState { Follower, Candidate, Leader }

description "Role of node 1"
var state1:     NodeState
description "Current Raft term for node 1 (bounded 0-3)"
var term1:      int
description "Node that node 1 voted for this term (0 = no vote, 1-5 = node ID)"
var voted_for1: int
description "Votes collected by node 1 as Candidate"
var votes1:     int
description "Log length of node 1 (bounded 0-3)"
var log_len1:   int

description "Role of node 2"
var state2:     NodeState
description "Current Raft term for node 2 (bounded 0-3)"
var term2:      int
description "Node that node 2 voted for this term (0 = no vote, 1-5 = node ID)"
var voted_for2: int
description "Votes collected by node 2 as Candidate"
var votes2:     int
description "Log length of node 2 (bounded 0-3)"
var log_len2:   int

description "Role of node 3"
var state3:     NodeState
description "Current Raft term for node 3 (bounded 0-3)"
var term3:      int
description "Node that node 3 voted for this term (0 = no vote, 1-5 = node ID)"
var voted_for3: int
description "Votes collected by node 3 as Candidate"
var votes3:     int
description "Log length of node 3 (bounded 0-3)"
var log_len3:   int

description "Role of node 4"
var state4:     NodeState
description "Current Raft term for node 4 (bounded 0-3)"
var term4:      int
description "Node that node 4 voted for this term (0 = no vote, 1-5 = node ID)"
var voted_for4: int
description "Votes collected by node 4 as Candidate"
var votes4:     int
description "Log length of node 4 (bounded 0-3)"
var log_len4:   int

description "Role of node 5"
var state5:     NodeState
description "Current Raft term for node 5 (bounded 0-3)"
var term5:      int
description "Node that node 5 voted for this term (0 = no vote, 1-5 = node ID)"
var voted_for5: int
description "Votes collected by node 5 as Candidate"
var votes5:     int
description "Log length of node 5 (bounded 0-3)"
var log_len5:   int

init {
  state1 = NodeState.Follower  state2 = NodeState.Follower  state3 = NodeState.Follower  state4 = NodeState.Follower  state5 = NodeState.Follower
  term1 = 0  term2 = 0  term3 = 0  term4 = 0  term5 = 0
  voted_for1 = 0  voted_for2 = 0  voted_for3 = 0  voted_for4 = 0  voted_for5 = 0
  votes1 = 0  votes2 = 0  votes3 = 0  votes4 = 0  votes5 = 0
  log_len1 = 0  log_len2 = 0  log_len3 = 0  log_len4 = 0  log_len5 = 0
}

// ── Election initiation ─────────────────────────────────────────────────
description "Node 1 starts a new election: bumps term, becomes Candidate, self-votes"
action startElection1 {
  require state1 != NodeState.Leader
  require term1 < 3
  state1'     = NodeState.Candidate
  term1'      = term1 + 1
  voted_for1' = 0
  votes1'     = 1
}

description "Node 2 starts a new election: bumps term, becomes Candidate, self-votes"
action startElection2 {
  require state2 != NodeState.Leader
  require term2 < 3
  state2'     = NodeState.Candidate
  term2'      = term2 + 1
  voted_for2' = 0
  votes2'     = 1
}

description "Node 3 starts a new election: bumps term, becomes Candidate, self-votes"
action startElection3 {
  require state3 != NodeState.Leader
  require term3 < 3
  state3'     = NodeState.Candidate
  term3'      = term3 + 1
  voted_for3' = 0
  votes3'     = 1
}

description "Node 4 starts a new election: bumps term, becomes Candidate, self-votes"
action startElection4 {
  require state4 != NodeState.Leader
  require term4 < 3
  state4'     = NodeState.Candidate
  term4'      = term4 + 1
  voted_for4' = 0
  votes4'     = 1
}

description "Node 5 starts a new election: bumps term, becomes Candidate, self-votes"
action startElection5 {
  require state5 != NodeState.Leader
  require term5 < 3
  state5'     = NodeState.Candidate
  term5'      = term5 + 1
  voted_for5' = 0
  votes5'     = 1
}

// ── Vote granting ──────────────────────────────────────────────────────
description "Node 2 (Follower) grants its vote to Candidate node 1"
action vote2for1 {
  require state1 == NodeState.Candidate
  require state2 == NodeState.Follower
  require votes1 < 5
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
  require votes1 < 5
  require term1 > term3 || (term1 == term3 && voted_for3 == 0)
  require voted_for3 == 0 || voted_for3 == 1
  voted_for3' = 1
  term3'      = term1
  votes1'     = votes1 + 1
}

description "Node 4 (Follower) grants its vote to Candidate node 1"
action vote4for1 {
  require state1 == NodeState.Candidate
  require state4 == NodeState.Follower
  require votes1 < 5
  require term1 > term4 || (term1 == term4 && voted_for4 == 0)
  require voted_for4 == 0 || voted_for4 == 1
  voted_for4' = 1
  term4'      = term1
  votes1'     = votes1 + 1
}

description "Node 5 (Follower) grants its vote to Candidate node 1"
action vote5for1 {
  require state1 == NodeState.Candidate
  require state5 == NodeState.Follower
  require votes1 < 5
  require term1 > term5 || (term1 == term5 && voted_for5 == 0)
  require voted_for5 == 0 || voted_for5 == 1
  voted_for5' = 1
  term5'      = term1
  votes1'     = votes1 + 1
}

description "Node 1 (Follower) grants its vote to Candidate node 2"
action vote1for2 {
  require state2 == NodeState.Candidate
  require state1 == NodeState.Follower
  require votes2 < 5
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
  require votes2 < 5
  require term2 > term3 || (term2 == term3 && voted_for3 == 0)
  require voted_for3 == 0 || voted_for3 == 2
  voted_for3' = 2
  term3'      = term2
  votes2'     = votes2 + 1
}

description "Node 4 (Follower) grants its vote to Candidate node 2"
action vote4for2 {
  require state2 == NodeState.Candidate
  require state4 == NodeState.Follower
  require votes2 < 5
  require term2 > term4 || (term2 == term4 && voted_for4 == 0)
  require voted_for4 == 0 || voted_for4 == 2
  voted_for4' = 2
  term4'      = term2
  votes2'     = votes2 + 1
}

description "Node 5 (Follower) grants its vote to Candidate node 2"
action vote5for2 {
  require state2 == NodeState.Candidate
  require state5 == NodeState.Follower
  require votes2 < 5
  require term2 > term5 || (term2 == term5 && voted_for5 == 0)
  require voted_for5 == 0 || voted_for5 == 2
  voted_for5' = 2
  term5'      = term2
  votes2'     = votes2 + 1
}

description "Node 1 (Follower) grants its vote to Candidate node 3"
action vote1for3 {
  require state3 == NodeState.Candidate
  require state1 == NodeState.Follower
  require votes3 < 5
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
  require votes3 < 5
  require term3 > term2 || (term3 == term2 && voted_for2 == 0)
  require voted_for2 == 0 || voted_for2 == 3
  voted_for2' = 3
  term2'      = term3
  votes3'     = votes3 + 1
}

description "Node 4 (Follower) grants its vote to Candidate node 3"
action vote4for3 {
  require state3 == NodeState.Candidate
  require state4 == NodeState.Follower
  require votes3 < 5
  require term3 > term4 || (term3 == term4 && voted_for4 == 0)
  require voted_for4 == 0 || voted_for4 == 3
  voted_for4' = 3
  term4'      = term3
  votes3'     = votes3 + 1
}

description "Node 5 (Follower) grants its vote to Candidate node 3"
action vote5for3 {
  require state3 == NodeState.Candidate
  require state5 == NodeState.Follower
  require votes3 < 5
  require term3 > term5 || (term3 == term5 && voted_for5 == 0)
  require voted_for5 == 0 || voted_for5 == 3
  voted_for5' = 3
  term5'      = term3
  votes3'     = votes3 + 1
}

description "Node 1 (Follower) grants its vote to Candidate node 4"
action vote1for4 {
  require state4 == NodeState.Candidate
  require state1 == NodeState.Follower
  require votes4 < 5
  require term4 > term1 || (term4 == term1 && voted_for1 == 0)
  require voted_for1 == 0 || voted_for1 == 4
  voted_for1' = 4
  term1'      = term4
  votes4'     = votes4 + 1
}

description "Node 2 (Follower) grants its vote to Candidate node 4"
action vote2for4 {
  require state4 == NodeState.Candidate
  require state2 == NodeState.Follower
  require votes4 < 5
  require term4 > term2 || (term4 == term2 && voted_for2 == 0)
  require voted_for2 == 0 || voted_for2 == 4
  voted_for2' = 4
  term2'      = term4
  votes4'     = votes4 + 1
}

description "Node 3 (Follower) grants its vote to Candidate node 4"
action vote3for4 {
  require state4 == NodeState.Candidate
  require state3 == NodeState.Follower
  require votes4 < 5
  require term4 > term3 || (term4 == term3 && voted_for3 == 0)
  require voted_for3 == 0 || voted_for3 == 4
  voted_for3' = 4
  term3'      = term4
  votes4'     = votes4 + 1
}

description "Node 5 (Follower) grants its vote to Candidate node 4"
action vote5for4 {
  require state4 == NodeState.Candidate
  require state5 == NodeState.Follower
  require votes4 < 5
  require term4 > term5 || (term4 == term5 && voted_for5 == 0)
  require voted_for5 == 0 || voted_for5 == 4
  voted_for5' = 4
  term5'      = term4
  votes4'     = votes4 + 1
}

description "Node 1 (Follower) grants its vote to Candidate node 5"
action vote1for5 {
  require state5 == NodeState.Candidate
  require state1 == NodeState.Follower
  require votes5 < 5
  require term5 > term1 || (term5 == term1 && voted_for1 == 0)
  require voted_for1 == 0 || voted_for1 == 5
  voted_for1' = 5
  term1'      = term5
  votes5'     = votes5 + 1
}

description "Node 2 (Follower) grants its vote to Candidate node 5"
action vote2for5 {
  require state5 == NodeState.Candidate
  require state2 == NodeState.Follower
  require votes5 < 5
  require term5 > term2 || (term5 == term2 && voted_for2 == 0)
  require voted_for2 == 0 || voted_for2 == 5
  voted_for2' = 5
  term2'      = term5
  votes5'     = votes5 + 1
}

description "Node 3 (Follower) grants its vote to Candidate node 5"
action vote3for5 {
  require state5 == NodeState.Candidate
  require state3 == NodeState.Follower
  require votes5 < 5
  require term5 > term3 || (term5 == term3 && voted_for3 == 0)
  require voted_for3 == 0 || voted_for3 == 5
  voted_for3' = 5
  term3'      = term5
  votes5'     = votes5 + 1
}

description "Node 4 (Follower) grants its vote to Candidate node 5"
action vote4for5 {
  require state5 == NodeState.Candidate
  require state4 == NodeState.Follower
  require votes5 < 5
  require term5 > term4 || (term5 == term4 && voted_for4 == 0)
  require voted_for4 == 0 || voted_for4 == 5
  voted_for4' = 5
  term4'      = term5
  votes5'     = votes5 + 1
}

// ── Leader promotion ──────────────────────────────────────────────────
description "Node 1 becomes Leader once it holds a strict majority"
action becomeLeader1 {
  require state1 == NodeState.Candidate
  require votes1 * 2 > 5
  state1' = NodeState.Leader
}

description "Node 2 becomes Leader once it holds a strict majority"
action becomeLeader2 {
  require state2 == NodeState.Candidate
  require votes2 * 2 > 5
  state2' = NodeState.Leader
}

description "Node 3 becomes Leader once it holds a strict majority"
action becomeLeader3 {
  require state3 == NodeState.Candidate
  require votes3 * 2 > 5
  state3' = NodeState.Leader
}

description "Node 4 becomes Leader once it holds a strict majority"
action becomeLeader4 {
  require state4 == NodeState.Candidate
  require votes4 * 2 > 5
  state4' = NodeState.Leader
}

description "Node 5 becomes Leader once it holds a strict majority"
action becomeLeader5 {
  require state5 == NodeState.Candidate
  require votes5 * 2 > 5
  state5' = NodeState.Leader
}

// ── Log replication: leader appends ──────────────────────────────────
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

description "Leader node 4 appends one new log entry"
action appendEntry4 {
  require state4 == NodeState.Leader
  require log_len4 < 3
  log_len4' = log_len4 + 1
}

description "Leader node 5 appends one new log entry"
action appendEntry5 {
  require state5 == NodeState.Leader
  require log_len5 < 3
  log_len5' = log_len5 + 1
}

// ── Log replication: follower accepts ────────────────────────────────
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

description "Leader node 1 replicates one entry to Follower node 4"
action replicate1to4 {
  require state1 == NodeState.Leader
  require state4 == NodeState.Follower
  require term4 == term1
  require log_len4 < log_len1
  log_len4' = log_len4 + 1
}

description "Leader node 1 replicates one entry to Follower node 5"
action replicate1to5 {
  require state1 == NodeState.Leader
  require state5 == NodeState.Follower
  require term5 == term1
  require log_len5 < log_len1
  log_len5' = log_len5 + 1
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

description "Leader node 2 replicates one entry to Follower node 4"
action replicate2to4 {
  require state2 == NodeState.Leader
  require state4 == NodeState.Follower
  require term4 == term2
  require log_len4 < log_len2
  log_len4' = log_len4 + 1
}

description "Leader node 2 replicates one entry to Follower node 5"
action replicate2to5 {
  require state2 == NodeState.Leader
  require state5 == NodeState.Follower
  require term5 == term2
  require log_len5 < log_len2
  log_len5' = log_len5 + 1
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

description "Leader node 3 replicates one entry to Follower node 4"
action replicate3to4 {
  require state3 == NodeState.Leader
  require state4 == NodeState.Follower
  require term4 == term3
  require log_len4 < log_len3
  log_len4' = log_len4 + 1
}

description "Leader node 3 replicates one entry to Follower node 5"
action replicate3to5 {
  require state3 == NodeState.Leader
  require state5 == NodeState.Follower
  require term5 == term3
  require log_len5 < log_len3
  log_len5' = log_len5 + 1
}

description "Leader node 4 replicates one entry to Follower node 1"
action replicate4to1 {
  require state4 == NodeState.Leader
  require state1 == NodeState.Follower
  require term1 == term4
  require log_len1 < log_len4
  log_len1' = log_len1 + 1
}

description "Leader node 4 replicates one entry to Follower node 2"
action replicate4to2 {
  require state4 == NodeState.Leader
  require state2 == NodeState.Follower
  require term2 == term4
  require log_len2 < log_len4
  log_len2' = log_len2 + 1
}

description "Leader node 4 replicates one entry to Follower node 3"
action replicate4to3 {
  require state4 == NodeState.Leader
  require state3 == NodeState.Follower
  require term3 == term4
  require log_len3 < log_len4
  log_len3' = log_len3 + 1
}

description "Leader node 4 replicates one entry to Follower node 5"
action replicate4to5 {
  require state4 == NodeState.Leader
  require state5 == NodeState.Follower
  require term5 == term4
  require log_len5 < log_len4
  log_len5' = log_len5 + 1
}

description "Leader node 5 replicates one entry to Follower node 1"
action replicate5to1 {
  require state5 == NodeState.Leader
  require state1 == NodeState.Follower
  require term1 == term5
  require log_len1 < log_len5
  log_len1' = log_len1 + 1
}

description "Leader node 5 replicates one entry to Follower node 2"
action replicate5to2 {
  require state5 == NodeState.Leader
  require state2 == NodeState.Follower
  require term2 == term5
  require log_len2 < log_len5
  log_len2' = log_len2 + 1
}

description "Leader node 5 replicates one entry to Follower node 3"
action replicate5to3 {
  require state5 == NodeState.Leader
  require state3 == NodeState.Follower
  require term3 == term5
  require log_len3 < log_len5
  log_len3' = log_len3 + 1
}

description "Leader node 5 replicates one entry to Follower node 4"
action replicate5to4 {
  require state5 == NodeState.Leader
  require state4 == NodeState.Follower
  require term4 == term5
  require log_len4 < log_len5
  log_len4' = log_len4 + 1
}

// ── Step-down ─────────────────────────────────────────────────────────
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

description "Node 1 steps down upon observing node 4 in a higher term"
action stepDown1_sees4 {
  require state1 != NodeState.Follower
  require term4 > term1
  state1'     = NodeState.Follower
  term1'      = term4
  voted_for1' = 0
  votes1'     = 0
}

description "Node 1 steps down upon observing node 5 in a higher term"
action stepDown1_sees5 {
  require state1 != NodeState.Follower
  require term5 > term1
  state1'     = NodeState.Follower
  term1'      = term5
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

description "Node 2 steps down upon observing node 4 in a higher term"
action stepDown2_sees4 {
  require state2 != NodeState.Follower
  require term4 > term2
  state2'     = NodeState.Follower
  term2'      = term4
  voted_for2' = 0
  votes2'     = 0
}

description "Node 2 steps down upon observing node 5 in a higher term"
action stepDown2_sees5 {
  require state2 != NodeState.Follower
  require term5 > term2
  state2'     = NodeState.Follower
  term2'      = term5
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

description "Node 3 steps down upon observing node 4 in a higher term"
action stepDown3_sees4 {
  require state3 != NodeState.Follower
  require term4 > term3
  state3'     = NodeState.Follower
  term3'      = term4
  voted_for3' = 0
  votes3'     = 0
}

description "Node 3 steps down upon observing node 5 in a higher term"
action stepDown3_sees5 {
  require state3 != NodeState.Follower
  require term5 > term3
  state3'     = NodeState.Follower
  term3'      = term5
  voted_for3' = 0
  votes3'     = 0
}

description "Node 4 steps down upon observing node 1 in a higher term"
action stepDown4_sees1 {
  require state4 != NodeState.Follower
  require term1 > term4
  state4'     = NodeState.Follower
  term4'      = term1
  voted_for4' = 0
  votes4'     = 0
}

description "Node 4 steps down upon observing node 2 in a higher term"
action stepDown4_sees2 {
  require state4 != NodeState.Follower
  require term2 > term4
  state4'     = NodeState.Follower
  term4'      = term2
  voted_for4' = 0
  votes4'     = 0
}

description "Node 4 steps down upon observing node 3 in a higher term"
action stepDown4_sees3 {
  require state4 != NodeState.Follower
  require term3 > term4
  state4'     = NodeState.Follower
  term4'      = term3
  voted_for4' = 0
  votes4'     = 0
}

description "Node 4 steps down upon observing node 5 in a higher term"
action stepDown4_sees5 {
  require state4 != NodeState.Follower
  require term5 > term4
  state4'     = NodeState.Follower
  term4'      = term5
  voted_for4' = 0
  votes4'     = 0
}

description "Node 5 steps down upon observing node 1 in a higher term"
action stepDown5_sees1 {
  require state5 != NodeState.Follower
  require term1 > term5
  state5'     = NodeState.Follower
  term5'      = term1
  voted_for5' = 0
  votes5'     = 0
}

description "Node 5 steps down upon observing node 2 in a higher term"
action stepDown5_sees2 {
  require state5 != NodeState.Follower
  require term2 > term5
  state5'     = NodeState.Follower
  term5'      = term2
  voted_for5' = 0
  votes5'     = 0
}

description "Node 5 steps down upon observing node 3 in a higher term"
action stepDown5_sees3 {
  require state5 != NodeState.Follower
  require term3 > term5
  state5'     = NodeState.Follower
  term5'      = term3
  voted_for5' = 0
  votes5'     = 0
}

description "Node 5 steps down upon observing node 4 in a higher term"
action stepDown5_sees4 {
  require state5 != NodeState.Follower
  require term4 > term5
  state5'     = NodeState.Follower
  term5'      = term4
  voted_for5' = 0
  votes5'     = 0
}

// ── Safety invariants ────────────────────────────────────────────────
description "Election safety: at most one Leader per Raft term"
invariant electionSafety {
  !(state1 == NodeState.Leader && state2 == NodeState.Leader && term1 == term2) &&
  !(state1 == NodeState.Leader && state3 == NodeState.Leader && term1 == term3) &&
  !(state1 == NodeState.Leader && state4 == NodeState.Leader && term1 == term4) &&
  !(state1 == NodeState.Leader && state5 == NodeState.Leader && term1 == term5) &&
  !(state2 == NodeState.Leader && state3 == NodeState.Leader && term2 == term3) &&
  !(state2 == NodeState.Leader && state4 == NodeState.Leader && term2 == term4) &&
  !(state2 == NodeState.Leader && state5 == NodeState.Leader && term2 == term5) &&
  !(state3 == NodeState.Leader && state4 == NodeState.Leader && term3 == term4) &&
  !(state3 == NodeState.Leader && state5 == NodeState.Leader && term3 == term5) &&
  !(state4 == NodeState.Leader && state5 == NodeState.Leader && term4 == term5)
}

description "Leader legitimacy: every Leader holds a strict majority of votes"
invariant leaderMajority {
  (!(state1 == NodeState.Leader) || votes1 * 2 > 5) &&
  (!(state2 == NodeState.Leader) || votes2 * 2 > 5) &&
  (!(state3 == NodeState.Leader) || votes3 * 2 > 5) &&
  (!(state4 == NodeState.Leader) || votes4 * 2 > 5) &&
  (!(state5 == NodeState.Leader) || votes5 * 2 > 5)
}

description "Candidate self-vote: every Candidate has cast at least its own vote"
invariant candidateSelfVote {
  (!(state1 == NodeState.Candidate) || votes1 >= 1) &&
  (!(state2 == NodeState.Candidate) || votes2 >= 1) &&
  (!(state3 == NodeState.Candidate) || votes3 >= 1) &&
  (!(state4 == NodeState.Candidate) || votes4 >= 1) &&
  (!(state5 == NodeState.Candidate) || votes5 >= 1)
}
