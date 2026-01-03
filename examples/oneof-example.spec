// Example: System with Multiple Initial States using oneOf
// Demonstrates non-deterministic initialization and verification across all initial states

description "Tracks a numeric counter value"
var counter: int

description "Current operating mode of the system"
var mode: str

description "Whether the system has been initialized"
var initialized: bool

description "System can start from any of three initial configurations: fresh start, resume, or restart"
init oneOf {
  {
    counter = 0
    mode = "start"
    initialized = false
  },
  {
    counter = 10
    mode = "resume"
    initialized = true
  },
  {
    counter = 20
    mode = "restart"
    initialized = true
  }
}

description "Increments the counter by one, ensuring it stays within bounds"
action increment {
  require counter < 100
  counter' = counter + 1
  mode' = mode
  initialized' = true
}

description "Decrements the counter by one, ensuring it stays positive"
action decrement {
  require counter > 0
  counter' = counter - 1
  mode' = mode
  initialized' = initialized
}

description "Changes the system mode to a new value"
action setMode(newMode: str) {
  require newMode = "start" || newMode = "resume" || newMode = "restart"
  counter' = counter
  mode' = newMode
  initialized' = initialized
}

description "Initializes the system if not already initialized"
action initialize {
  require !initialized
  initialized' = true
  counter' = counter
  mode' = mode
}

description "Ensures counter is always within valid range"
invariant counterValid {
  counter >= 0 && counter <= 100
}

description "Ensures mode is always one of the valid values"
invariant modeValid {
  mode = "start" || mode = "resume" || mode = "restart"
}

description "Ensures initialized systems have non-negative counters"
invariant initializationConsistency {
  !initialized || counter >= 0
}

description "Verifies that counter will eventually reach a high value"
temporal eventuallyHighCounter {
  eventually (counter >= 50)
}

description "Verifies that system will eventually become initialized"
temporal eventuallyInitialized {
  eventually initialized
}

description "Guarantees progress: if starting from resume or restart mode, counter will eventually change"
temporal progressFromResume {
  always ((mode = "resume" || mode = "restart") → eventually counter != counter)
}

