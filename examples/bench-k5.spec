// Benchmark spec: k=5 invariants (for monitor overhead scaling measurement)
var aliceBalance: int
var bobBalance: int

init {
  aliceBalance = 0
  bobBalance = 0
}

action depositAlice50 {
  aliceBalance' = aliceBalance + 50
}

action withdrawAlice30 {
  require aliceBalance >= 30
  aliceBalance' = aliceBalance - 30
}

action transfer50ToBob {
  require aliceBalance >= 50
  aliceBalance' = aliceBalance - 50
  bobBalance' = bobBalance + 50
}

invariant inv1 { aliceBalance >= 0 }
invariant inv2 { bobBalance >= 0 }
invariant inv3 { aliceBalance <= 1000000 }
invariant inv4 { bobBalance <= 1000000 }
invariant inv5 { aliceBalance + bobBalance >= 0 }
