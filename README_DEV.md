# Spectre Language - Development

This is the Go implementation of the Spectre specification language.

## Project Structure

```
spectre/
├── cmd/spectre/          # CLI tool entry point
├── internal/             # Internal packages (not exported)
│   ├── lexer/           # Tokenizer/lexer
│   ├── parser/          # Parser
│   ├── types/           # Type system
│   ├── semantic/        # Semantic analysis
│   └── verifier/        # Verification engine
├── pkg/                  # Public packages
│   ├── ast/             # AST definitions
│   └── errors/          # Error types
└── examples/            # Example .spec files for testing
```

## Building

```bash
go build -o spectre ./cmd/spectre
```

## Testing

```bash
go test ./...
```

## Development Workflow

1. Work on one phase at a time
2. All tests must pass before moving to next phase
3. Test with example files in examples/ directory

