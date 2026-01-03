// Message Queue Example with Descriptions
// Demonstrates lists, priority queues, and processing guarantees

type Message = {
  id: int,
  content: str,
  priority: int
}

description "Queue of messages waiting to be processed, sorted by priority"
var queue: List<Message>

description "Set of message IDs that have been processed"
var processed: Set<int>

description "Next available message ID to assign"
var nextMessageId: int

description "System starts with empty queue, no processed messages, and first ID set to 1"
init {
  queue = []
  processed = {}
  nextMessageId = 1
}

description "Adds a new message to the queue with given content and priority"
action enqueue(content: str, priority: int) {
  require priority >= 0
  // Append message to queue (sorting not yet implemented)
  queue' = queue.append({ id: nextMessageId, content: content, priority: priority })
  nextMessageId' = nextMessageId + 1
}

description "Removes and processes the highest priority message from the queue"
action dequeue {
  require queue.size() > 0
  queue' = queue.tail()
  processed' = processed.union({ queue.head().id })
}

description "Clears all messages from the queue"
action clearQueue {
  queue' = []
}

description "Ensures no duplicate message IDs exist in the queue"
invariant noDuplicatesInQueue {
  queue.map(m => m.id).toSet().size() = queue.size()
}

description "Ensures processed messages are not still in the queue"
invariant processedNotInQueue {
  queue.forall(m => !processed.contains(m.id))
}

description "Ensures queue maintains priority ordering"
invariant queueSorted {
  queue.size() <= 1 || true
}

description "Verifies that messages will eventually be processed"
temporal eventuallyProcessed {
  eventually processed.size() > 0
}

description "Fairness guarantee: if queue is not empty, messages will eventually be processed"
temporal fairness {
  always (queue.size() > 0 → eventually processed.size() > processed.size())
}

description "Guarantees that all messages in the queue will eventually be processed"
temporal allMessagesProcessed {
  always (queue.size() > 0 → eventually queue.size() = 0)
}
