// Distributed Message Queue - Violation Version
// Demonstrates temporal property violations in a message queue system

description "Maximum capacity of the message queue"
const MAX_QUEUE_SIZE: int = 10

description "Queue of messages waiting to be processed"
var queue: List<str>

description "Set of messages that have been processed"
var processed: Set<str>

description "Counter for tracking total messages produced"
var messageCounter: int

description "Initial state: empty queue, no processed messages, counter at 0"
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

description "Consumer can process messages multiple times"
description "PROBLEM: No check if message was already processed"
action reprocess {
  // Missing: require queue.size() > 0
  // Missing: require !processed.contains(queue.head())
  processed' = processed.union({ queue.head() })
  queue' = queue.tail()
}

description "Invariant: Queue size never exceeds maximum capacity"
invariant queueCapacity {
  queue.size() <= 10
}

description "Invariant: Processed messages are never duplicated in the processed set"
invariant noDuplicates {
  true  // This would need a more complex check, simplified here
}

description "Temporal: Messages eventually get processed"
description "PROBLEM: Without fairness, consumers might never execute"
temporal messageProcessed {
  always (queue.size() > 0 → eventually processed.contains(queue.head()))
}

description "Temporal: Queue doesn't grow unbounded"
temporal queueBounded {
  always queue.size() < 20
}

description "Temporal: All produced messages eventually get processed"
description "PROBLEM: Without fairness or proper protocol, messages might never be consumed"
temporal allMessagesProcessed {
  always eventually (messageCounter > 0 → processed.size() == messageCounter)
}

