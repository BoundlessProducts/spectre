# TLA+ Feature Analysis for Spectre

This document analyzes which TLA+ features are present in Spectre and which are missing.

## ✅ Features Already Present

### Core Language Features
- ✅ **State Variables** (`var` vs TLA+ `VARIABLES`)
- ✅ **Initial States** (`init` vs TLA+ `Init`)
- ✅ **Actions/Transitions** (`action` vs TLA+ actions)
- ✅ **Invariants** (`invariant` vs TLA+ `Invariant`)
- ✅ **Temporal Properties** (`temporal` vs TLA+ `Temporal`)
- ✅ **Prime Notation** (`'` for next state)
- ✅ **Sets, Maps, Lists** (with operations)
- ✅ **Records/Tuples**
- ✅ **Pure Functions** (`fun` vs TLA+ operators)
- ✅ **Preconditions** (`require`)
- ✅ **Postconditions** (`ensure`)
- ✅ **Descriptions** (Spectre-specific enhancement)

### Temporal Operators
- ✅ `always` (□ in TLA+)
- ✅ `eventually` (◇ in TLA+)
- ✅ `until`
- ✅ `→` (leads to)
- ✅ `next` (mentioned in syntax summary)

### Logical Operators
- ✅ Boolean operators (`&&`, `||`, `!`)
- ✅ Quantifiers (`forall`, `exists`)
- ✅ Comparisons (`=`, `!=`, `<`, `>`, `<=`, `>=`)

### Collection Operations
- ✅ Set operations (union, intersection, filter, map, etc.)
- ✅ List operations (append, head, tail, etc.)
- ✅ Map operations (put, get, contains, etc.)

## ❌ Missing Features from TLA+

### 1. **Modules and Imports**
**TLA+ Feature:** Modules can extend and import other modules
```tla
EXTENDS Naturals, Sequences
LOCAL INSTANCE ModuleName
```

**Status:** ❌ Not present
**Impact:** High - Limits code reuse and modularity
**Recommendation:** Add module system:
```spectre
module Counter {
  // specification
}

import Counter
extend Counter
```

### 2. **Constants vs Variables**
**TLA+ Feature:** Distinguishes between `CONSTANTS` (parameters) and `VARIABLES` (state)
```tla
CONSTANTS N, MaxValue
VARIABLES counter, users
```

**Status:** ❌ Not present (only `var` exists)
**Impact:** Medium - Constants are useful for parameterized specifications
**Recommendation:** Add `const` keyword:
```spectre
const N: int = 10
const MaxValue: int = 100
var counter: int
```

### 3. **CHOOSE Operator**
**TLA+ Feature:** Non-deterministic choice from a set
```tla
CHOOSE x \in S : P(x)
```

**Status:** ❌ Not present
**Impact:** Medium - Useful for non-deterministic specifications
**Recommendation:** Add `choose` operator:
```spectre
let x = choose s in Set where predicate(s)
```

### 4. **CASE Expressions**
**TLA+ Feature:** Pattern matching/switch-like expressions
```tla
CASE x = 1 -> value1
    x = 2 -> value2
    OTHER -> default
```

**Status:** ❌ Not present (only `if/else` exists)
**Impact:** Low-Medium - Can be approximated with `if/else`, but CASE is cleaner
**Recommendation:** Add `case` expression:
```spectre
let result = case x {
  1 => value1,
  2 => value2,
  _ => default
}
```

### 5. **Fairness Conditions**
**TLA+ Feature:** Weak Fairness (WF) and Strong Fairness (SF)
```tla
WF_vars(Action)
SF_vars(Action)
```

**Status:** ❌ Not present
**Impact:** High - Essential for liveness properties in concurrent systems
**Recommendation:** Add fairness operators:
```spectre
temporal weakFairness {
  WF(process1Request)
}

temporal strongFairness {
  SF(process2Request)
}
```

### 6. **Action Composition**
**TLA+ Feature:** Actions can be composed with `/\` (conjunction) and `\/` (disjunction)
```tla
Action1 /\ Action2  // Both happen simultaneously
Action1 \/ Action2  // One or the other happens
```

**Status:** ⚠️ Partially present (actions can have multiple statements, but no explicit composition)
**Impact:** Medium - Useful for complex transitions
**Recommendation:** Add action composition:
```spectre
action combined {
  action1 /\ action2  // Both execute
}

action alternative {
  action1 \/ action2  // One executes
}
```

### 7. **Stuttering Steps**
**TLA+ Feature:** Actions can leave state unchanged (stuttering)
```tla
Action \/ UNCHANGED <<vars>>
```

**Status:** ⚠️ Implicitly supported (actions can have no state changes)
**Impact:** Low - Already possible but not explicit
**Recommendation:** Add explicit `unchanged` keyword:
```spectre
action noop {
  unchanged counter, users
}
```

### 8. **THEOREM and ASSUME**
**TLA+ Feature:** For proofs and assumptions
```tla
ASSUME N > 0
THEOREM Spec => Invariant
```

**Status:** ❌ Not present
**Impact:** Low-Medium - Useful for proof systems, less critical for model checking
**Recommendation:** Add for future proof system:
```spectre
assume N > 0
theorem specImpliesInvariant {
  spec → invariant
}
```

### 9. **Sequences vs Lists**
**TLA+ Feature:** Sequences are functions from natural numbers to values
```tla
Seq = [i \in 1..Len | value]
```

**Status:** ⚠️ Lists exist but may not have all sequence operations
**Impact:** Low-Medium - Lists may be sufficient, but sequences have specific operations
**Recommendation:** Verify list operations match sequence needs, or add explicit `Sequence` type

### 10. **Model Parameters**
**TLA+ Feature:** Constants can be instantiated in models
```tla
CONSTANTS N
```

**Status:** ❌ Not present (no constants)
**Impact:** Medium - Important for parameterized verification
**Recommendation:** Add constants (see #2)

### 11. **Instance Substitution**
**TLA+ Feature:** `INSTANCE` keyword for module instantiation
```tla
INSTANCE Counter WITH counter <- myCounter
```

**Status:** ❌ Not present (no modules)
**Impact:** Medium - Depends on modules (#1)
**Recommendation:** Add with module system

### 12. **Set Comprehensions**
**TLA+ Feature:** Set builder notation
```tla
{x \in S : P(x)}
{x : x \in S /\ P(x)}
```

**Status:** ⚠️ Partially present (via `filter` and `map`)
**Impact:** Low - Functional style covers this
**Recommendation:** Consider adding set comprehensions for familiarity:
```spectre
let result = {x in S where P(x)}
let result = {f(x) for x in S where P(x)}
```

### 13. **Function Updates**
**TLA+ Feature:** `[f EXCEPT !.key = value]`
```tla
[f EXCEPT !.x = 5]
```

**Status:** ⚠️ Present via spread operator (`...`)
**Impact:** Low - Spread operator covers this
**Recommendation:** Current approach is fine

### 14. **Temporal Formula Operators**
**TLA+ Feature:** Additional temporal operators
- `ENABLED` - Action is enabled
- `UNCHANGED` - Variables unchanged
- `\cdot` - Action composition in temporal formulas

**Status:** ⚠️ Partially present
**Impact:** Medium - `ENABLED` is useful
**Recommendation:** Add `enabled` operator:
```spectre
temporal actionEnabled {
  always enabled(increment)
}
```

### 15. **Recursive Definitions**
**TLA+ Feature:** Recursive operators
```tla
RECURSIVE Factorial(_)
Factorial(n) == IF n = 0 THEN 1 ELSE n * Factorial(n-1)
```

**Status:** ✅ Present (pure functions can be recursive)
**Impact:** None - Already supported

### 16. **LET-IN Expressions**
**TLA+ Feature:** Local definitions
```tla
LET x == 5
    y == 10
IN x + y
```

**Status:** ✅ Present (`let` keyword)
**Impact:** None - Already supported

## Priority Recommendations

### High Priority (Essential Features)
1. **Modules and Imports** - Critical for code organization and reuse
2. **Fairness Conditions** - Essential for concurrent system verification
3. **Constants** - Important for parameterized specifications

### Medium Priority (Useful Features)
4. **CHOOSE Operator** - Useful for non-deterministic specifications
5. **Action Composition** - Helps express complex transitions
6. **ENABLED Operator** - Useful for temporal properties
7. **CASE Expressions** - Improves readability

### Low Priority (Nice to Have)
8. **THEOREM/ASSUME** - For future proof system integration
9. **Set Comprehensions** - Syntactic sugar (already possible)
10. **Explicit UNCHANGED** - Clarifies intent

## Summary

Spectre covers most core TLA+ features but is missing:
- **Modularity** (modules, imports, instances)
- **Fairness** (weak/strong fairness)
- **Constants** (vs variables)
- **Some operators** (CHOOSE, CASE, ENABLED)

The language is well-designed for basic specifications but would benefit from these additions for more complex, real-world specifications.

