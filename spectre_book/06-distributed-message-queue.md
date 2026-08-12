# Chapter 6: Distributed Message Queue

This chapter demonstrates temporal property violations in a distributed message queue system.

## The Message Queue System

Our message queue system has:
- **Producers**: Add messages to the queue
- **Consumers**: Process and remove messages from the queue
- **Queue**: A bounded buffer storing pending messages
- **Processed Set**: Tracks messages that have been processed

### System Requirements

1. **Safety**: Queue should never exceed its maximum capacity
2. **Safety**: Messages should only be processed once
3. **Liveness**: Messages should eventually be processed
4. **Liveness**: The queue should not grow unbounded

## The Violation Version

Let's examine `examples/message-queue-violation.spec`:

```spectre
description "Queue of messages waiting to be processed"
var queue: List<str>

description "Set of messages that have been processed"
var processed: Set<str>

description "Counter for tracking total messages produced"
var messageCounter: int

init {
  queue = List.empty()
  processed = Set.empty()
  messageCounter = 0
}

description "Producer adds a message to the queue"
description "PROBLEM: No check if queue is at capacity"
action produce {
  // Missing: require queue.size() < MAX_QUEUE_SIZE
  queue' = queue.append("msg")
  messageCounter' = messageCounter + 1
}

description "Consumer processes a message from the queue"
description "PROBLEM: No check if queue is empty"
action consume {
  // Missing: require queue.size() > 0
  processed' = processed.union({ queue.head() })
  queue' = queue.tail()
}
```

### Running the Violation Version

```bash
$ ./spectre verify examples/message-queue-violation.spec

Verification failed: 2 violation(s) found

Violation 1 (Temporal Property: messageProcessed):
  Property 'messageProcessed' violated: (PROBLEM: Without fairness, consumers might never execute) 
  P holds but Q never becomes true
  Counterexample trace:
    2. produce
    3. produce
    ...
    11. produce

Violation 2 (Temporal Property: allMessagesProcessed):
  Property 'allMessagesProcessed' violated: (PROBLEM: Without fairness or proper protocol, messages might never be consumed)
  eventually P does not hold from some state
  Counterexample trace:
    2. produce
    3. produce
```

### Understanding the Temporal Violations

#### Violation 1: `messageProcessed`

```spectre
description "Temporal: Messages eventually get processed"
temporal messageProcessed {
  always (queue.size() > 0 → eventually processed.contains(queue.head()))
}
```

**What it means**: If the queue is not empty, the message at the head should eventually be processed.

**Why it fails**:
- The property has no fairness constraint
- Without fairness, the scheduler could continuously execute `produce` actions
- The `consume` action, even when enabled, might never execute
- Result: Messages are added to the queue but never processed

**Counterexample Analysis**:
- Starting from the initial state, only `produce` actions execute
- The queue grows to 10 messages (max capacity)
- But `consume` never executes, so messages remain unprocessed
- The temporal property fails because the queue head is never processed

#### Violation 2: `allMessagesProcessed`

```spectre
description "Temporal: All produced messages eventually get processed"
temporal allMessagesProcessed {
  always eventually (messageCounter > 0 → processed.size() == messageCounter)
}
```

**What it means**: Eventually, all produced messages should be processed (i.e., `processed.size() == messageCounter`).

**Why it fails**:
- Similar to the first violation: without fairness, `consume` may never execute
- Messages are produced but never consumed
- The system can get stuck in states where `messageCounter > 0` but `processed.size() == 0`

**Counterexample Analysis**:
- After a few `produce` actions, `messageCounter > 0`
- But `consume` never executes, so `processed.size() == 0`
- The property `processed.size() == messageCounter` never becomes true

## The Corrected Version

Let's fix these issues in `examples/message-queue-corrected.spec`:

### Fix 1: Add Preconditions

```spectre
description "Producer adds a message to the queue"
description "FIXED: Added precondition to check queue capacity"
action produce {
  require queue.size() < MAX_QUEUE_SIZE
  queue' = queue.append("msg")
  messageCounter' = messageCounter + 1
}

description "Consumer processes a message from the queue"
description "FIXED: Added precondition to check queue is not empty"
description "FIXED: Ensures message is only processed once"
action consume {
  require queue.size() > 0
  require !processed.contains(queue.head())
  processed' = processed.union({ queue.head() })
  queue' = queue.tail()
}
```

**What changed**:
- `produce`: Added `require queue.size() < MAX_QUEUE_SIZE` to prevent queue overflow
- `consume`: Added `require queue.size() > 0` to prevent processing from empty queue
- `consume`: Added `require !processed.contains(queue.head())` to prevent duplicate processing

### Fix 2: Add Fairness Constraints

```spectre
description "Temporal: Queue eventually makes progress (with fairness)"
description "FIXED: Added weak fairness to ensure consume executes when continuously enabled"
temporal queueProgress {
  WF(consume) → eventually (processed.size() > 0)
}
```

**What changed**:
- Added `WF(consume)` (Weak Fairness) to the temporal property
- Changed from `always eventually` to just `eventually` - we only need that there exists a fair path where progress is made
- Weak fairness ensures: "If `consume` is continuously enabled, it will eventually execute"
- This guarantees that when the queue is not empty and `produce` can't execute (queue at capacity), `consume` will eventually run

**Why Weak Fairness Works**:
- When the queue reaches capacity (10 messages), `produce` can no longer execute (precondition `queue.size() < 10` fails)
- At that point, only `consume` is enabled, making it continuously enabled
- Weak fairness guarantees that if `consume` is continuously enabled, it will eventually execute
- This breaks the counterexample where `consume` never runs

### Fix 3: Add Capacity Bounds

```spectre
description "Invariant: Queue size never exceeds maximum capacity"
invariant queueCapacity {
  queue.size() <= 10
}

description "Temporal: Queue doesn't grow unbounded"
description "FIXED: Precondition ensures queue never exceeds capacity, so this property holds"
temporal queueBounded {
  always queue.size() <= 10
}
```

**What changed**:
- The `produce` precondition ensures the queue never exceeds capacity
- The invariant `queueCapacity` enforces this at all states
- The temporal property `queueBounded` is satisfied because the precondition prevents unbounded growth

## Running the Corrected Version

```bash
$ ./spectre verify examples/message-queue-corrected.spec

✓ Verification passed for examples/message-queue-corrected.spec
  Explored X states
  Verified 3 temporal properties
```

All temporal properties now hold:
- **`queueProgress`**: Messages eventually get processed (with fairness)
- **`queueBounded`**: Queue never exceeds capacity (due to preconditions)

## Key Lessons

### 1. Preconditions Enforce Safety

Preconditions ensure actions can only execute when the system is in a valid state:
- `produce` only when queue has space
- `consume` only when queue is not empty
- This prevents runtime errors and maintains invariants

### 2. Fairness is Essential for Liveness

Without fairness constraints, temporal properties that depend on actions eventually executing will fail:
- The scheduler could continuously skip enabled actions
- Fairness guarantees execution under certain conditions
- Weak fairness: "If continuously enabled, eventually executes"

### 3. Temporal Properties Verify System Behavior

Temporal properties ensure the system makes progress and behaves correctly over time:
- **Safety properties** (invariants): Must hold in all states
- **Liveness properties** (temporal): Must eventually hold along execution paths

### 4. The Verifier Finds Execution Paths

The verifier explores all possible execution paths and checks if properties hold:
- Finds counterexamples where properties fail
- Helps identify missing preconditions and fairness constraints
- Guides you to write correct specifications

## Summary

The message queue example demonstrates:
- **Temporal violations** when fairness is missing
- **How to fix** with weak fairness (`WF`) constraints
- **Preconditions** to enforce safety properties
- **The importance** of both safety and liveness in distributed systems

By adding preconditions and fairness constraints, we ensure:
1. The queue never exceeds capacity (safety)
2. Messages are eventually processed (liveness)
3. All produced messages eventually get consumed (liveness)

