module StutteringCounter {
  var counter: int
  
  init {
    counter = 0
  }
  
  description "Increments the counter by one"
  action increment {
    counter' = counter + 1
  }
  
  description "Does nothing - causes stuttering"
  action noop {
    counter' = counter
  }
  
  description "Resets the counter to zero"
  action reset {
    counter' = 0
  }
  
  description "Verifies counter eventually reaches 10"
  temporal eventuallyReachesTen {
    eventually (counter = 10)
  }
}
