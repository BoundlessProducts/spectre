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

**[📖 The Spectre Language Book](./spectre_book/)**

The book is organized into chapters:
- **[Chapter 1: Getting Started](./spectre_book/01-getting-started.md)** - Installation instructions and examples
- **[Chapter 2: Language Overview](./spectre_book/02-language-overview.md)** - All language elements and concepts
- **[Chapter 3: Invariants and Violations](./spectre_book/03-invariants-and-violations.md)** - Understanding and fixing invariant violations
- **[Chapter 4: Temporal and Fairness Properties](./spectre_book/04-temporal-and-fairness-properties.md)** - Temporal logic and fairness constraints
- **[Chapter 5: Concurrent Systems and Locking](./spectre_book/05-concurrent-systems-and-locking.md)** - Modeling concurrent systems
- **[Chapter 6: Distributed Message Queue](./spectre_book/06-distributed-message-queue.md)** - Message queue system example

## Building from Source

To build Spectre from source, you'll need:

- **Go 1.21 or later** ([download](https://go.dev/dl/))

### Build Steps

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd spectre
   ```

2. **Build the CLI tool**:
   ```bash
   go build -o spectre ./cmd/spectre
   ```

   This creates a `spectre` executable in the current directory.

3. **Test the build**:
   ```bash
   ./spectre --help
   ```

4. **Run tests** (optional):
   ```bash
   go test ./...
   ```

5. **Install globally** (optional):
   ```bash
   go install ./cmd/spectre
   ```

   This installs `spectre` to your `$GOPATH/bin` or `$HOME/go/bin` directory.

### Development Setup

For development setup and workflow, see **[README_DEV.md](./README_DEV.md)**.

## Additional Resources

- **[SPEC.md](./SPEC.md)** - Complete language specification
- **[USAGE.md](./USAGE.md)** - CLI usage guide
- **[README_DEV.md](./README_DEV.md)** - Development setup and workflow
- **[STATUS.md](./STATUS.md)** - Implementation status and progress

## License

(To be determined)
