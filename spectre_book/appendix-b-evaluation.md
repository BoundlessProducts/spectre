# Appendix B: Evaluation Order and Priming Rules

Two questions come up constantly when writing Spectre specs:

1. **Priming:** "If I write `x' = y` and `y' = x` in the same action, do I get a swap or an infinite loop?"
2. **Expressions:** "What order are sub-expressions evaluated in, and can they affect state?"

This appendix answers both definitively.

---

## B.1 Simultaneous assignment

Spectre actions use **simultaneous assignment** — the semantics of TLA+ actions.

The rule:

> Every right-hand side in an action is evaluated in the **pre-action state** (the state before the action fires). No assignment takes effect until all right-hand sides have been evaluated.

```spectre
action swap {
  x' = y    // reads y from pre-action state
  y' = x    // reads x from pre-action state
}
```

After `swap` fires, `x` has the old value of `y` and `y` has the old value of `x`. This is a correct swap — there is no "temporary variable" needed.

Compare with an imperative language:
```
// Imperative (wrong for swap)
x = y    // x is now y's old value
y = x    // y is now x's NEW value (= y's old value) — no swap!

// Spectre (correct swap)
x' = y   // x_next = y_old
y' = x   // y_next = x_old   <- reads x_old, not x_next
```

---

## B.2 Implicit UNCHANGED

Variables not assigned in an action keep their values automatically.

```spectre
var balance: int
var frozen:  bool

action deposit(amount: int) {
  require !frozen
  balance' = balance + amount
  // frozen is NOT assigned — implicitly: frozen' = frozen
}
```

You never need to write `frozen' = frozen` manually. Any variable with no `x'` assignment in an action is implicitly unchanged.

---

## B.3 Double assignment is an error

Each variable may be assigned at most once per action. Assigning the same variable twice is a parse error:

```spectre
// ILLEGAL — balance' assigned twice
action bad {
  balance' = balance + 10
  balance' = balance + 20   // parse error
}
```

If you need conditional assignment, use `if`:

```spectre
action conditional(amount: int) {
  balance' = if amount > 0 then balance + amount else balance
}
```

---

## B.4 Guards evaluate in the pre-action state

`require` and `ensure` are evaluated in the same pre-action state as all other right-hand sides.

```spectre
action withdraw(amount: int) {
  require balance >= amount   // reads balance BEFORE withdrawal
  balance' = balance - amount
}
```

If `require` evaluates to `false`, the action is **disabled** — it simply does not fire in that state. The verifier tries the next enabled action.

`ensure` is a postcondition: it is checked in the **post-action state** (after all assignments take effect).

```spectre
action increment {
  x' = x + 1
  ensure x' > 0   // checked in next state: x + 1 > 0
}
```

---

## B.5 Init is also simultaneous

`init` bodies follow the same simultaneous-assignment rule:

```spectre
init {
  x = 10
  y = x    // y = 10 (reads x from the init state being built, NOT from x=10 above)
}
```

In practice, write `init` assignments as independent constant expressions — this avoids confusion and is consistent with how the verifier builds the initial state.

---

## B.6 Actions are atomic

An action fires as a single indivisible step. There is no "intermediate state" between the first and last assignment. The transition relation maps the pre-action state directly to the post-action state.

```
pre-state s
    |
    | action fires (all RHS evaluated in s, all LHS applied simultaneously)
    |
    v
post-state s'
```

This matches TLA+'s action semantics exactly.

---

## B.7 Expression evaluation

### Purity

Expressions in Spectre are **pure**: they produce a value and have no side effects. No expression can modify state. Only `x' = e` assignments (in action and init bodies) change state.

### Arithmetic

`int` is **unbounded**: there is no overflow, no wrapping, no undefined behavior. `+`, `-`, `*` are exact.

```spectre
invariant no_overflow { x + y < 1000000 }  // always checkable, never wraps
```

`/` is integer division (truncates toward zero). `%` is the remainder (sign matches dividend).

### Short-circuit evaluation

`&&` and `||` short-circuit left-to-right:

- `false && e` → `false` without evaluating `e`
- `true || e` → `true` without evaluating `e`

This matters when `e` is a method call that might fail on an empty collection:

```spectre
// Safe: if nodes is empty, nodes.head() is never called
invariant safe { nodes.size() > 0 && nodes.head() > 0 }
```

### The `in` operator

`x in S` checks whether `x` is an element of set `S` (or a key in map `S`). It evaluates `S` first, then checks membership.

```spectre
invariant leader_valid { leader in active_nodes }
```

For maps, `k in m` returns `true` if key `k` has an entry in map `m`.

### Method calls

`obj.method(args)` evaluates `obj`, then each argument left-to-right, then applies the method. Methods return a new value — they never mutate `obj` in place.

```spectre
// nodes is unchanged; union returns a new set
nodes' = nodes.union({new_node})
```

### Collection method quick reference

| Collection | Methods |
|------------|---------|
| `Set<T>` | `.union(S)`, `.intersection(S)`, `.difference(S)`, `.contains(x)`, `.filter(f)`, `.size()`, `.toList()` |
| `Map<K,V>` | `.put(k,v)`, `.get(k)`, `.contains(k)`, `.size()` |
| `List<T>` | `.append(x)`, `.head()`, `.tail()`, `.filter(f)`, `.map(f)`, `.size()`, `.toSet()` |

---

## B.8 Priming summary

| Syntax | Meaning | Where allowed |
|--------|---------|---------------|
| `x` | Current (pre-action) value of variable `x` | Anywhere in expressions |
| `x'` on LHS | Next-state assignment: set `x` in post-state | Left-hand side of `AssignStmt` only |
| `x'` on RHS | Next-state read: read `x` from post-state | Allowed inside `ensure` only |
| Unassigned `x` | Implicit `x' = x` | Variables with no assignment in an action |

**The golden rule:** write `x' = f(x, y, z)` where `x`, `y`, `z` are all pre-action values. If you need to refer to a computed value multiple times, use `let`:

```spectre
action update(delta: int) {
  let new_val = balance + delta
  require new_val >= 0
  require new_val <= LIMIT
  balance' = new_val
}
```
