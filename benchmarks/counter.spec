// Benchmark: simple linear counter.
// ~10 states; baseline for BFS overhead.
var n: int
init { n = 0 }
action incr { require n < 10  n' = n + 1 }
action decr { require n > 0   n' = n - 1 }
invariant non_negative { n >= 0 }
invariant bounded      { n <= 10 }
