// Benchmark: resource allocator — N resources shared among 3 clients.
// ~hundreds of states; tests larger state space.
param N: int

var free: int
var held_a: int
var held_b: int
var held_c: int

init { free = N  held_a = 0  held_b = 0  held_c = 0 }

action alloc_a { require free > 0  free' = free - 1  held_a' = held_a + 1  held_b' = held_b  held_c' = held_c }
action alloc_b { require free > 0  free' = free - 1  held_a' = held_a  held_b' = held_b + 1  held_c' = held_c }
action alloc_c { require free > 0  free' = free - 1  held_a' = held_a  held_b' = held_b  held_c' = held_c + 1 }

action free_a  { require held_a > 0  free' = free + 1  held_a' = held_a - 1  held_b' = held_b  held_c' = held_c }
action free_b  { require held_b > 0  free' = free + 1  held_a' = held_a  held_b' = held_b - 1  held_c' = held_c }
action free_c  { require held_c > 0  free' = free + 1  held_a' = held_a  held_b' = held_b  held_c' = held_c - 1 }

invariant conservation { free + held_a + held_b + held_c == N }
invariant non_negative { free >= 0 && held_a >= 0 && held_b >= 0 && held_c >= 0 }
