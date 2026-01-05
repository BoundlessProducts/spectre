module StutteringCounterWithFairness {
  var counter: int
  
  init {
    counter = 0
  }
  
  description "Increments the counter by one"
  action increment {
    counter' = counter + 1
  }
  
  description "Idle step - doesn't change state"
  description "This action can execute but doesn't make progress"
  action noop {
    counter' = counter
  }
  
  description "Resets the counter to zero"
  action reset {
    counter' = 0
  }
  
  description "Verifies counter eventually reaches 10 with weak fairness on increment"
  description "Weak fairness ensures increment executes when continuously enabled"
  temporal eventuallyReachesTen {
    WF(increment) → eventually (counter = 10)
  }
}
