// Pure functions example: demonstrates function declarations in Spectre.
// Pure functions are side-effect-free helpers used inside actions and invariants.

description "Running total"
var total: int

description "Current maximum seen value"
var maximum: int

description "Count of values added"
var count: int

description "Start with empty accumulator"
init {
  total = 0
  maximum = 0
  count = 0
}

fun add(x: int, y: int): int {
  return x + y
}

fun max(a: int, b: int): int {
  if a >= b {
    return a
  } else {
    return b
  }
}

fun clamp(value: int, lo: int, hi: int): int {
  if value < lo {
    return lo
  } else {
    if value > hi {
      return hi
    } else {
      return value
    }
  }
}

fun isPositive(n: int): bool {
  return n > 0
}

description "Add a positive value to the accumulator"
action addValue(amount: int) {
  require amount > 0
  total' = add(total, amount)
  maximum' = max(maximum, amount)
  count' = count + 1
}

description "Total is always non-negative"
invariant totalNonNegative {
  total >= 0
}

description "Maximum is always non-negative"
invariant maxNonNegative {
  maximum >= 0
}

description "Total is at least as large as the maximum when count is positive"
invariant totalAtLeastMax {
  count == 0 || total >= maximum
}
