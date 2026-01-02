# IMPORTANT: This is still under construction. It has been made public only to test installations etc. No PRs WILL BE ENTERTAINED.

# Spectre Language

A programmer-friendly specification language inspired by TLA+ and Quint, designed specifically for Java and TypeScript developers.

## What is Spectre?

Spectre is a formal specification language that makes it easy to model systems as state machines and verify their correctness. Unlike traditional formal methods that can be intimidating, Spectre provides:

- **Familiar syntax** similar to Java and TypeScript
- **Strong typing** with type inference
- **Clear semantics** for state machines and transitions
- **Comprehensive verification** tools for invariants and temporal properties
- **Excellent error messages** with descriptions and execution traces

## Key Features

- ✅ **Type System**: Primitive and compound types (records, sets, maps, lists, enums, options)
- ✅ **State Management**: Explicit state variables with prime notation (`'`) for next-state
- ✅ **Non-determinism**: `oneOf` operator for multiple initial states
- ✅ **Pure Functions**: Reusable computational helpers without side effects
- ✅ **Temporal Logic**: `always`, `eventually`, `until`, `leads-to` operators
- ✅ **Fairness**: Weak Fairness (WF) and Strong Fairness (SF) conditions for concurrent systems
- ✅ **Modules**: Code organization with imports and inheritance
- ✅ **Descriptions**: Human-readable context in error messages

## Quick Example

```spectre
description "Tracks a numeric counter value"
var counter: int

description "System starts with counter initialized to zero"
init {
  counter = 0
}

description "Increments the counter by one"
action increment {
  counter' = counter + 1
}

description "Ensures counter never becomes negative"
invariant nonNegative {
  counter >= 0
}
```

## Documentation

For complete installation instructions, usage guide, language reference, and examples, please see:

**[📖 The Spectre Language Book](./SPECTRE_BOOK.md)**

The book includes:
- Installation instructions for macOS, Linux, and Windows
- Getting started guide with examples
- Complete language reference
- Tutorials and advanced topics
- Debugging guide

## Additional Resources

- **[SPEC.md](./SPEC.md)** - Complete language specification
- **[USAGE.md](./USAGE.md)** - CLI usage guide
- **[README_DEV.md](./README_DEV.md)** - Development setup and workflow

## License

(To be determined)
