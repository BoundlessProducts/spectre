// Modules example: demonstrates module declarations, visibility, and composition.
// This file contains multiple module declarations showing the module system.

module Counter {
  description "Tracks a numeric counter value"
  var counter: int

  description "System starts with counter initialized to zero"
  init {
    counter = 0
  }

  description "Increments the counter by one"
  public action increment {
    counter' = counter + 1
  }

  description "Decrements the counter by one, only when counter is positive"
  public action decrement {
    require counter > 0
    counter' = counter - 1
  }

  description "Resets the counter back to zero"
  public action reset {
    counter' = 0
  }

  description "Ensures counter never becomes negative"
  public invariant nonNegative {
    counter >= 0
  }
}

module BoundedCounter extends Counter {
  description "Maximum allowed counter value"
  const MAX_VALUE: int = 100

  description "Increments counter but enforces maximum bound"
  public action increment {
    require counter < MAX_VALUE
    super.increment()
  }

  description "Ensures counter stays within bounds"
  public invariant bounded {
    counter <= MAX_VALUE
  }
}

module App {
  import BoundedCounter

  description "A bounded counter used by the application"
  var myCounter: int

  description "Application starts with counter at zero"
  init {
    myCounter = 0
  }

  description "Increment the application counter within bounds"
  action increment {
    require myCounter < BoundedCounter.MAX_VALUE
    myCounter' = myCounter + 1
  }

  description "Application counter is always non-negative"
  invariant appCounterNonNegative {
    myCounter >= 0
  }
}
