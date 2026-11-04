# Collections and distributed specifications

Scalar variables (`int`, `bool`, enum) are enough for counters and state machines. Real distributed protocols need something more: sets of nodes, vote tallies, message queues. Spectre has first-class `Set<T>`, `Map<K,V>`, and `List<T>` types for exactly this.

---

## The three collection types

| Type | Use for | Example init |
|------|---------|-------------|
| `Set<T>` | Unordered membership: nodes, voters, committed replicas | `nodes = {1, 2, 3}` |
| `Map<K,V>` | Key→value associations: votes, roles, config | `votes = {}` |
| `List<T>` | Ordered sequences: message queues, logs, histories | `queue = []` |

Declare them exactly like any variable:

```spectre
var members:  Set<int>
var role:     Map<int, bool>
var log:      List<int>
```

---

## Set operations

```spectre
var S: Set<int>
var T: Set<int>

init { S = {1, 2, 3}  T = {3, 4} }

action grow   { S' = S.union({4}) }
action shrink { S' = S.difference({1}) }
```

The full method set:

| Method | Returns | Meaning |
|--------|---------|---------|
| `S.union(T)` | `Set<T>` | All elements of S or T |
| `S.intersection(T)` | `Set<T>` | Elements in both S and T |
| `S.difference(T)` | `Set<T>` | Elements in S but not T |
| `S.contains(x)` | `bool` | True if x is in S |
| `x in S` | `bool` | Same — infix form, reads naturally in guards |
| `S.size()` | `int` | Number of elements |
| `S.filter(fn)` | `Set<T>` | Elements matching predicate |
| `S.toList()` | `List<T>` | Convert to list (order unspecified) |

The `in` operator is the most natural way to write membership guards:

```spectre
action vote {
  require !(node_id in voted)   // has not voted yet
  voted' = voted.union({node_id})
}
```

---

## Map operations

```spectre
var scores: Map<int, int>
init { scores = {} }

action record { scores' = scores.put(42, 100) }
```

| Method | Returns | Meaning |
|--------|---------|---------|
| `M.put(k, v)` | `Map<K,V>` | Returns map with key k set to v |
| `M.get(k)` | `V` | Value for key k (error if missing) |
| `k in M` | `bool` | True if k has an entry in M |
| `M.contains(k)` | `bool` | Same — method form |
| `M.size()` | `int` | Number of entries |

---

## List operations

```spectre
var queue: List<int>
init { queue = [] }

action enqueue { queue' = queue.append(42) }
```

| Method | Returns | Meaning |
|--------|---------|---------|
| `L.append(x)` | `List<T>` | List with x added at end |
| `L.head()` | `T` | First element (error if empty) |
| `L.tail()` | `List<T>` | List without first element |
| `L.size()` | `int` | Number of elements |
| `L.filter(fn)` | `List<T>` | Elements matching predicate |
| `L.toSet()` | `Set<T>` | Convert to set (drops order/duplicates) |

---

## Example: Two-Phase Commit

Two-Phase Commit (2PC) is a classic distributed transaction protocol. It requires sets to reason about: which nodes have voted, which have committed, and whether the results are consistent.

```
Coordinator                   Nodes
     |                           |
     |---- Prepare ------------->|
     |                           |
     |<--- Vote (yes/no) --------|  each node votes independently
     |                           |
  all yes?                       |
     |---- Commit/Abort -------->|
     |                           |
     |<--- Ack ------------------|
```

**The spec** (`examples/two_phase_commit.spec`):

```spectre
const NUM_NODES: int = 3

enum CoordPhase { Preparing, Committed, Aborted }

var coord:     CoordPhase
var voted_yes: Set<int>   // nodes that voted yes
var voted_no:  Set<int>   // nodes that voted no
var committed: Set<int>   // nodes that applied commit
var aborted:   Set<int>   // nodes that applied abort

init {
  coord     = CoordPhase.Preparing
  voted_yes = {}
  voted_no  = {}
  committed = {}
  aborted   = {}
}

action node1_vote_yes {
  require coord == CoordPhase.Preparing
  require !(1 in voted_yes) && !(1 in voted_no)
  voted_yes' = voted_yes.union({1})
}

// ... node2_vote_yes, node3_vote_yes, node1_vote_no, ... (same pattern)

action coordinator_commit {
  require coord == CoordPhase.Preparing
  require voted_yes.size() == NUM_NODES
  require voted_no.size() == 0
  coord' = CoordPhase.Committed
}

action coordinator_abort {
  require coord == CoordPhase.Preparing
  require voted_no.size() > 0
  coord' = CoordPhase.Aborted
}

action node1_commit {
  require coord == CoordPhase.Committed
  require !(1 in committed)
  committed' = committed.union({1})
}

// ... node2_commit, node3_commit, node1_abort, ... (same pattern)
```

**The safety invariants:**

```spectre
// A node never both commits and aborts (atomicity)
invariant atomicity {
  committed.intersection(aborted).size() == 0
}

// A node only commits if the coordinator committed
invariant commit_requires_coord_committed {
  committed.size() == 0 || coord == CoordPhase.Committed
}

// No node votes both yes and no
invariant votes_disjoint {
  voted_yes.intersection(voted_no).size() == 0
}

// The coordinator only commits on unanimous yes vote
invariant unanimous_commit {
  coord != CoordPhase.Committed || voted_yes.size() == NUM_NODES
}
```

**Run it:**

```bash
spectre verify examples/two_phase_commit.spec
# Exhaustive verification: traversed 673 states (full reachable space)
# Found no violations.
```

673 states covers every possible interleaving of votes and decisions across all 3 nodes. BFS finds this exhaustively in milliseconds — a great use of the learning backend before scaling up.

---

## Why collection specs find bugs scalars can't

Consider a buggy 2PC coordinator that commits as soon as *any* node votes yes (not all):

```spectre
action coordinator_commit_buggy {
  require coord == CoordPhase.Preparing
  require voted_yes.size() > 0    // BUG: should be == NUM_NODES
  coord' = CoordPhase.Committed
}
```

The `unanimous_commit` invariant immediately catches this:

```
Violation 1: invariant "unanimous_commit"
  Step 0: [coord=Preparing, voted_yes={}, voted_no={}, ...]
  Step 1: [coord=Preparing, voted_yes={1}, voted_no={},  ...]  <- node 1 votes yes
  Step 2: [coord=Committed, voted_yes={1}, voted_no={},  ...]  <- coordinator commits too early
  violated: coord == Committed but voted_yes.size() == 1 != 3
```

This class of bug is invisible in scalar-variable specs. The set tracks *which* nodes voted, not just *how many* — making it possible to state the unanimous condition precisely.

---

## State-space note: collections multiply states

Each unique set value is a distinct state. `Set<int>` over `{1,2,3}` has 8 possible values (every subset). With 5 variables like the 2PC spec, the product is large — but BFS still exhausts it at 673 states because the invariants and guards constrain which subsets are reachable.

For larger protocols (Raft, Paxos, 5+ nodes), state spaces quickly exceed BFS capacity. That's when you escalate:

```
Small protocol (2-3 nodes)   -->  --backend bfs        (learning, exhaustive)
Medium protocol (4-5 nodes)  -->  --backend tlc        (exhaustive, validated)
Large protocol (6+ nodes)    -->  --backend apalache   (symbolic, exhaustive)
Liveness / leader election   -->  --backend tlc        (validated fairness)
```

**A note on Z3 and collections**: The Z3 backend (bounded model checking) translates `Set<T>` to SMT-LIB2 sets and can catch shallow set bugs. However, `Cardinality` checks over `Set<int>` are unreliable in Z3 because integer sets are infinite — Z3 may report spurious violations. Use Z3 for set membership and boolean checks; use TLC or Apalache for specs with `.size()` comparisons on integer sets.

---

## Collection syntax quick-reference

```spectre
// Set literal
var S: Set<int>
init { S = {1, 2, 3} }

// Empty set / empty map / empty list
var empty_set:  Set<int>    -- init to {}
var empty_map:  Map<int,bool>  -- init to {}
var empty_list: List<int>   -- init to []

// Membership test (in a require or invariant)
require x in S
invariant inv { !(x in S) || x > 0 }

// Set update (all assignments return new values — no mutation)
S' = S.union({x})
S' = S.difference({x})

// Map update
M' = M.put(key, value)

// List update
L' = L.append(x)
let head = L.head()   // read first element (use in expressions/requires)
let rest = L.tail()   // read tail (use in expressions/requires)
```
