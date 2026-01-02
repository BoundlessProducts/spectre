// Modules Example
// Demonstrates module organization, imports, and extension

// Base counter module
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

  description "Decrements the counter by one"
  public action decrement {
    require counter > 0
    counter' = counter - 1
  }

  description "Ensures counter never becomes negative"
  public invariant nonNegative {
    counter >= 0
  }
}

// Bounded counter extends base counter
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

// Usage example
module App {
  import BoundedCounter

  description "Create an instance of bounded counter"
  var myCounter: int

  description "Initialize using bounded counter"
  init {
    myCounter = 0
  }

  description "Increment with bound checking"
  action increment {
    require myCounter < BoundedCounter.MAX_VALUE
    myCounter' = myCounter + 1
  }
}

