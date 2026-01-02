# Spectre Language - Development Guide

This is the Go implementation of the Spectre specification language.

## Project Structure

```
spectre/
├── cmd/spectre/          # CLI tool entry point
│   ├── main.go          # Main entry point
│   ├── commands.go      # CLI command implementations
│   ├── file_processor.go # File processing utilities
│   └── *_test.go        # CLI tests
├── internal/             # Internal packages (not exported)
│   ├── lexer/           # Tokenizer/lexer
│   ├── parser/          # Parser
│   ├── types/           # Type system
│   ├── semantic/        # Semantic analysis
│   ├── state/           # State machine model
│   ├── exec/            # Execution engine
│   ├── explore/         # State space exploration
│   ├── temporal/        # Temporal property evaluation
│   ├── eval/            # Expression evaluation
│   └── errors/          # Error reporting
├── pkg/                  # Public packages
│   └── ast/             # AST definitions
├── examples/            # Example .spec files for testing
├── scripts/             # Installation scripts
│   ├── install.sh       # Linux installation script
│   ├── install.ps1      # Windows installation script
│   ├── uninstall.sh     # Linux uninstall script
│   ├── uninstall.ps1    # Windows uninstall script
│   ├── test-linux-install.sh        # Test Linux install with Docker
│   └── test-linux-install-local.sh   # Test Linux install script locally
├── Formula/             # Homebrew formula
│   └── spectre.rb       # Homebrew formula for macOS
├── Makefile             # Build automation
├── .goreleaser.yml      # GoReleaser configuration
└── DEVELOPMENT_PLAN.md  # Development phases and progress
```

## Building

### Local Development Build

```bash
go build -o spectre ./cmd/spectre
```

### Build for All Platforms

```bash
make build-all
```

This creates binaries in `dist/` for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

### Build Single Platform

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o spectre-linux ./cmd/spectre

# macOS
GOOS=darwin GOARCH=amd64 go build -o spectre-darwin ./cmd/spectre

# Windows
GOOS=windows GOARCH=amd64 go build -o spectre-windows.exe ./cmd/spectre
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
make test-coverage
```

This generates `coverage.html` showing test coverage.

### Run Specific Package Tests

```bash
go test ./internal/parser -v
go test ./cmd/spectre -v
```

## Development Workflow

1. Work on one phase at a time
2. All tests must pass before moving to next phase
3. Test with example files in examples/ directory

## Packaging and Distribution

### Creating Releases

1. **Update version** in `cmd/spectre/main.go`:
   ```go
   const Version = "0.1.0"
   ```

2. **Tag the release**:
   ```bash
   git tag -a v0.1.0 -m "Release version 0.1.0"
   git push origin v0.1.0
   ```

3. **Build releases** with GoReleaser:
   ```bash
   # Install GoReleaser
   brew install goreleaser
   
   # Test release (dry run)
   goreleaser release --snapshot
   
   # Production release
   goreleaser release
   ```

### Installation Scripts

- **Linux**: `scripts/install.sh` - Automated installation script
- **Windows**: `scripts/install.ps1` - PowerShell installation script
- **Homebrew**: `Formula/spectre.rb` - Homebrew formula for macOS

See [PACKAGING.md](./PACKAGING.md) for detailed packaging instructions.

### Testing Installers

**macOS (Homebrew)**:
```bash
brew install --build-from-source Formula/spectre.rb
```

**Linux**:
```bash
./scripts/install.sh
```

**Windows**:
```powershell
powershell -ExecutionPolicy Bypass -File scripts/install.ps1
```

## Testing Installation Scripts

### Testing Linux Installation Script on macOS

#### Method 1: Using Docker (Recommended)

Test the Linux installation script in a real Linux environment:

```bash
# Test on Ubuntu (default)
./scripts/test-linux-install.sh

# Test on different distributions
./scripts/test-linux-install.sh ubuntu
./scripts/test-linux-install.sh fedora
./scripts/test-linux-install.sh debian
```

**Prerequisites**: Docker Desktop must be installed and running.

#### Method 2: Local Script Validation

Test the script syntax and logic without running it:

```bash
./scripts/test-linux-install-local.sh
```

This checks:
- Script syntax
- Required variables
- Error handling
- Go version check logic

#### Method 3: Manual Docker Test

You can also manually test in a Docker container:

```bash
# Start an Ubuntu container
docker run --rm -it ubuntu:latest bash

# Inside the container:
apt-get update
apt-get install -y git curl
curl -fsSL https://raw.githubusercontent.com/akkeshavan/spectre/main/scripts/install.sh | bash
spectre --version
```

#### Method 4: Test Specific Scenarios

Test error cases (missing Go, wrong version, etc.):

```bash
# Test without Go installed (should show helpful error)
docker run --rm -it ubuntu:latest bash -c "
  apt-get update -qq && apt-get install -y -qq git curl
  curl -fsSL https://raw.githubusercontent.com/akkeshavan/spectre/main/scripts/install.sh | bash
"
```

---

## Makefile Targets

- `make build` - Build binary locally
- `make build-all` - Build for all platforms
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make clean` - Remove build artifacts
- `make release` - Create snapshot release (test)
- `make release-production` - Create production release

