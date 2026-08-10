// Extended Raft node — election + log replication perspective.
//
// Models the per-node state of a node in a Raft cluster, covering:
//   • Leader election with per-term voting discipline
//   • Log replication (append + replicate to follower)
//   • Leader step-down upon observing a higher term
//
// Used with `spectre mine --lang rust` to extract the per-node skeleton.
// The cluster-level specification (raft-election-safety.spec) composes three
// instances and adds the cluster-level safety invariants.

#[derive(Debug, Clone, PartialEq)]
pub enum NodeState {
    Follower,
    Candidate,
    Leader,
}

/// Per-node state for a node in a Raft cluster.
pub struct RaftNode {
    pub state:           NodeState,
    pub current_term:    u64,   // Raft term; bounded in spec to keep state space finite
    pub voted_for:       u64,   // 0 = no vote; 1-3 = candidate node ID
    pub votes_received:  u64,   // ballots collected while Candidate
    pub log_length:      u64,   // length of this node's log; bounded in spec
    pub cluster_size:    u64,   // fixed; 3 for the cluster model
}

impl RaftNode {
    pub fn new(cluster_size: u64) -> Self {
        assert!(cluster_size >= 3);
        Self {
            state:          NodeState::Follower,
            current_term:   0,
            voted_for:      0,
            votes_received: 0,
            log_length:     0,
            cluster_size,
        }
    }

    /// Begin a new election: bump term, become Candidate, cast a self-vote.
    pub fn start_election(&mut self) {
        assert!(self.state != NodeState::Leader);
        assert!(self.current_term < 3); // term bound; prevents unbounded state space
        self.current_term += 1;
        self.state = NodeState::Candidate;
        self.voted_for = 0;
        self.votes_received = 1;
    }

    /// Grant a vote to a candidate running in `candidate_term`.
    /// Requires that the candidate's term is at least as recent as ours,
    /// and that we have not already voted for a different candidate this term.
    pub fn grant_vote(&mut self, candidate_term: u64, candidate_id: u64) {
        assert!(candidate_term >= self.current_term);
        assert!(self.voted_for == 0 || self.voted_for == candidate_id);
        self.state        = NodeState::Follower;
        self.current_term = candidate_term;
        self.voted_for    = candidate_id;
    }

    /// Record one incoming vote from a peer.
    pub fn receive_vote(&mut self) {
        assert!(self.state == NodeState::Candidate);
        self.votes_received += 1;
    }

    /// Promote to Leader once a strict majority of votes is collected.
    pub fn become_leader(&mut self) {
        assert!(self.state == NodeState::Candidate);
        assert!(self.votes_received * 2 > self.cluster_size);
        self.state = NodeState::Leader;
    }

    /// Leader appends one new entry to its own log.
    pub fn append_entry(&mut self) {
        assert!(self.state == NodeState::Leader);
        assert!(self.log_length < 3); // log length bound
        self.log_length += 1;
    }

    /// Follower accepts one replicated entry from the current leader.
    /// `leader_log_length` is the length of the leader's log (passed in from the cluster model).
    pub fn replicate_entry(&mut self, leader_log_length: u64) {
        assert!(self.state == NodeState::Follower);
        assert!(self.log_length < leader_log_length);
        self.log_length += 1;
    }

    /// Step down to Follower upon observing a message from a higher-term node.
    pub fn step_down(&mut self, new_term: u64) {
        assert!(new_term > self.current_term);
        assert!(new_term <= 3); // term bound
        self.current_term   = new_term;
        self.state          = NodeState::Follower;
        self.voted_for      = 0;
        self.votes_received = 0;
    }
}
