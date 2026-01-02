# Spectre vs Quint: Key Differences

## Overview

Both **Spectre** and **Quint** are programmer-friendly specification languages inspired by TLA+, designed to make formal specification accessible to developers familiar with mainstream programming languages. However, there are several key differences in their design philosophy, syntax, and features.

## Key Differences

### 1. **Language Status**

- **Quint**: Mature, actively developed language with production-ready tooling
  - Developed by Informal Systems
  - Has a working compiler, IDE integrations, and verification tools
  - Used in real-world projects (blockchain protocols, distributed systems)
  
- **Spectre**: Specification/design document (as of this writing)
  - Designed as a conceptual language specification
  - Tooling to be implemented
  - Focus on clarity and programmer-friendliness

### 2. **Syntax Philosophy**

#### Quint
- More functional programming style
- Uses `val` for immutable values and `var` for state variables
- Actions are defined with `action` keyword
- More mathematical notation influence from TLA+

#### Spectre
- More imperative/OOP style familiar to Java/TypeScript developers
- Uses `var` for state variables (similar)
- Actions use `action` keyword (similar)
- More familiar control flow (`if/else` blocks)
- Explicit `require` and `ensure` keywords for pre/postconditions

### 3. **Type System**

#### Quint
- Strong type system with type inference
- Supports records, sets, maps, lists
- Uses `type` keyword for type aliases
- More functional type system

#### Spectre
- Similar type system but with more OOP-like syntax
- Explicit support for enums: `enum Status { Pending, Completed }`
- Option types: `Option<T>` (Some/None pattern)
- More explicit type annotations encouraged

### 4. **State and Transitions**

#### Quint
```quint
var counter: int

action init = counter' = 0

action increment = counter' = counter + 1
```

#### Spectre
```spectre
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}
```

**Difference**: Spectre uses block syntax (`{}`) for multi-line definitions, making it more familiar to Java/TypeScript developers.

### 5. **Constraints and Properties**

#### Quint
- Uses `invariant` for safety properties
- Temporal properties defined with `temporal` keyword
- Properties are expressions, not blocks

#### Spectre
- Uses `invariant` for classic constraints (same)
- Uses `temporal` for temporal properties (same)
- Adds explicit `require` (precondition) and `ensure` (postcondition) keywords
- Block syntax for complex constraints

**Example Spectre addition**:
```spectre
action withdraw(amount: int) {
  require account.balance >= amount
  account' = { ...account, balance: account.balance - amount }
  ensure account'.balance >= 0
}
```

### 6. **Control Flow**

#### Quint
- More functional/expression-based
- Uses ternary operators and functional combinators
- Less explicit control flow constructs

#### Spectre
- More imperative style
- Explicit `if/else` blocks
- Familiar to Java/TypeScript developers
- More verbose but potentially clearer

**Example Spectre**:
```spectre
action process {
  if (counter > 0) {
    counter' = counter - 1
  } else {
    counter' = counter
  }
}
```

### 7. **Temporal Operators**

Both languages support similar temporal operators:
- `always P`: P holds in all states
- `eventually P`: P holds in at least one future state
- `P until Q`: P holds until Q becomes true
- `P → Q`: P leads to Q

**Difference**: Spectre explicitly documents these in the spec with examples, making them more accessible.

### 8. **Collections and Data Structures**

#### Quint
- Sets: `Set<T>`
- Maps: `Map<K, V>`
- Lists: `List<T>`
- Records: `{ field: type }`

#### Spectre
- Same collection types
- Adds explicit `enum` support
- Adds `Option<T>` type
- More explicit about tuple syntax: `(int, int)`

### 9. **Tooling and Ecosystem**

#### Quint
- **Mature tooling**:
  - Quint compiler
  - IDE support (VS Code extension)
  - Integration with Apalache for model checking
  - REPL for interactive development
  - Test framework
  - Documentation generator

#### Spectre
- **Planned tooling** (to be implemented):
  - Compiler/parser
  - Verifier (bounded and unbounded model checking)
  - IDE support
  - Integration with existing TLA+ tools

### 10. **Use Cases**

#### Quint
- Production use in blockchain protocols
- Distributed systems verification
- Real-world projects requiring formal verification
- Active development and community

#### Spectre
- Designed for similar use cases
- Focus on maximum programmer-friendliness
- Emphasis on clear, readable specifications
- May be easier for teams new to formal methods

## Similarities

Both languages share many common goals and features:

1. **Programmer-friendly**: Both aim to be accessible to Java/TypeScript developers
2. **Type system**: Both have strong type systems with similar type constructs
3. **State machines**: Both use similar concepts (state variables, actions, transitions)
4. **Verification**: Both support invariants and temporal properties
5. **TLA+ foundation**: Both are inspired by TLA+ but with more familiar syntax

## Syntax Comparison Example

### Counter Example

#### Quint
```quint
module Counter {
  var counter: int

  action init = counter' = 0

  action increment = counter' = counter + 1

  action decrement = counter' = counter - 1

  val nonNegative = counter >= 0

  temporal eventuallyTen = eventually (counter = 10)
}
```

#### Spectre
```spectre
var counter: int

init {
  counter = 0
}

action increment {
  counter' = counter + 1
}

action decrement {
  require counter > 0
  counter' = counter - 1
}

invariant nonNegative {
  counter >= 0
}

temporal eventuallyTen {
  eventually (counter = 10)
}
```

## Summary

| Feature | Quint | Spectre |
|---------|-------|---------|
| **Status** | Production-ready | Specification |
| **Syntax Style** | Functional | Imperative/OOP |
| **Control Flow** | Expression-based | Block-based |
| **Preconditions** | Inline guards | Explicit `require` |
| **Postconditions** | Not explicit | Explicit `ensure` |
| **Enums** | Via types | Native `enum` |
| **Option Types** | Via types | Native `Option<T>` |
| **Tooling** | Mature | Planned |
| **Block Syntax** | Minimal | Extensive |

## When to Choose Which?

### Choose Quint if:
- You need production-ready tooling now
- You prefer functional programming style
- You want active community support
- You're working on real-world projects requiring immediate verification

### Choose Spectre if:
- You prefer imperative/OOP syntax
- You want explicit pre/postconditions
- You value block-based syntax for clarity
- You're designing a new language or tooling
- You want native enum and Option type support

## Conclusion

Spectre and Quint are both excellent approaches to making TLA+ more accessible. Quint is the mature, production-ready option, while Spectre represents a design that emphasizes even more programmer-friendly syntax with explicit control flow and constraint keywords. The choice between them depends on your team's preferences, tooling needs, and syntax style preferences.

