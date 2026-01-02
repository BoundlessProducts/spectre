# Spectre CLI Usage Guide

## Commands

### `parse`

Parses a Spectre specification file and reports syntax errors.

**Usage:**
```bash
spectre parse <file.spec>
```

**Examples:**
```bash
# Parse a single file
spectre parse examples/counter.spec

# Parse multiple files
spectre parse examples/counter.spec examples/mutex.spec

# Parse all files in a directory
spectre parse examples/
```

**Output:**
- On success: `✓ Successfully parsed <filename>`
- On error: Lists all parse errors with line and column numbers

### `typecheck`

Type-checks a Spectre specification file and reports type errors.

**Usage:**
```bash
spectre typecheck <file.spec>
```

**Examples:**
```bash
# Type-check a single file
spectre typecheck examples/counter.spec

# Type-check multiple files
spectre typecheck examples/*.spec
```

**Output:**
- On success: `✓ Successfully type-checked <filename>`
- On error: Lists all type errors with positions and descriptions

**Note:** Temporal expressions are not fully type-checked yet and may produce warnings.

### `verify`

Verifies a Spectre specification by checking invariants and temporal properties.

**Usage:**
```bash
spectre verify <file.spec>
```

**Examples:**
```bash
# Verify a single file
spectre verify examples/counter.spec

# Verify multiple files
spectre verify examples/counter.spec examples/mutex.spec
```

**Output:**
- On success: Shows number of states explored
- On violation: Shows violation details with execution path

**Verification Process:**
1. Parses the specification
2. Builds symbol table and resolves modules
3. Type-checks all declarations
4. Creates state machine model
5. Explores state space (BFS)
6. Checks invariants at each state
7. Evaluates temporal properties over traces

## Error Messages

Spectre provides detailed error messages with:
- **Source positions**: Line and column numbers
- **Descriptions**: Context from `description` fields
- **Stack traces**: Execution paths leading to violations

**Example error message:**
```
21:2: Invariant 'counterNonNegative' violated: (Counter must be non-negative) condition evaluated to false
```

## File Processing

The CLI supports:
- **Single files**: `spectre parse file.spec`
- **Multiple files**: `spectre parse file1.spec file2.spec`
- **Directories**: `spectre parse examples/` (processes all `.spec` files)

When processing multiple files:
- Each file is processed independently
- Errors in one file don't stop processing of others
- Summary shows total succeeded/failed counts

## Exit Codes

- `0`: Success
- `1`: Error (parse error, type error, or verification failure)

## Performance

Typical performance for example files:
- **Parse**: < 1ms per file
- **Typecheck**: < 10ms per file
- **Verify**: Depends on state space size (typically < 1s for small examples)

For larger specifications, verification time scales with:
- Number of state variables
- Number of actions
- State space size
- Depth of exploration

## Tips

1. **Start with parse**: Always parse first to catch syntax errors
2. **Then typecheck**: Fix type errors before verification
3. **Use descriptions**: Add `description` fields for better error messages
4. **Test incrementally**: Verify small parts before combining

## Troubleshooting

**"Parse failed"**: Check syntax, ensure file has `.spec` extension

**"Type checking failed"**: Verify variable types match, check function signatures

**"Verification failed"**: Review invariants and temporal properties, check for violations in execution traces

**"File not found"**: Ensure file path is correct, check working directory

