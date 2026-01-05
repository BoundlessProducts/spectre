# Chapter 1: Getting Started

This chapter will help you install Spectre and run your first specification.

---

## Installation

Install Spectre using your preferred method:

**macOS (Homebrew):**

**Prerequisites**: Go 1.19 or later must be installed (Homebrew will build Spectre from source).

If Go is not installed, install it first:
```bash
brew install go
```

Since the Homebrew formula is in the main repository (not a separate `homebrew-spectre` repo), use the full GitHub URL:

```bash
brew tap akkeshavan/spectre https://github.com/akkeshavan/spectre.git
brew install spectre
```

**Note**: The installation builds from source, so Go is required. Homebrew will automatically install Go as a dependency if it's not already installed, but it's recommended to install Go first to ensure you have the correct version.

**Note**: If you later create a separate `homebrew-spectre` repository, you can simplify to:
```bash
brew tap akkeshavan/spectre
brew install spectre
```

**Linux:**

**Prerequisites**: Go 1.19 or later must be installed (the script builds Spectre from source).

The install script downloads the source code and builds Spectre locally:

```bash
curl -fsSL https://raw.githubusercontent.com/akkeshavan/spectre/main/scripts/install.sh | bash
```

The script will:
- Check for Go 1.19+ installation (with helpful error messages if not found)
- Check for Git installation
- Clone the repository from GitHub
- Build Spectre from source using Go
- Install to `/usr/local/bin` (or custom `INSTALL_DIR`)

**If Go is not installed**, the script will provide installation instructions. You can install Go with:

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install golang-go

# Fedora/RHEL/CentOS
sudo dnf install golang
```

Or download from: https://golang.org/dl/

**Note**: The script automatically detects if it's running as root (e.g., in Docker containers) and skips `sudo` commands. On regular Linux systems, it will use `sudo` for installation.

**Windows:**  
```powershell
powershell -ExecutionPolicy Bypass -File scripts/install.ps1
```

---

## Building from Source

If you prefer to build Spectre from source (useful for development or if you want the latest code), follow these steps:

### Prerequisites

- **Go 1.21 or later** ([download](https://go.dev/dl/))
- **Git** (for cloning the repository)

### Build Steps

1. **Clone the repository**:
   ```bash
   git clone https://github.com/akkeshavan/spectre.git
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
   
   You should see usage information for the Spectre CLI.

4. **Run tests** (optional):
   ```bash
   go test ./...
   ```
   
   This runs the full test suite to ensure everything works correctly.

5. **Install globally** (optional):
   ```bash
   go install ./cmd/spectre
   ```
   
   This installs `spectre` to your `$GOPATH/bin` or `$HOME/go/bin` directory. Make sure this directory is in your `PATH` environment variable.

### Using the Built Executable

After building, you can use the `spectre` executable directly:

```bash
# From the project root directory
./spectre parse examples/counter.spec
./spectre typecheck examples/counter.spec
./spectre verify examples/counter.spec
```

If you installed globally with `go install`, you can use `spectre` from anywhere:

```bash
spectre parse examples/counter.spec
spectre typecheck examples/counter.spec
spectre verify examples/counter.spec
```

### Accessing Examples

When you clone the repository, all example files are included in the `examples/` directory:

```bash
# List all examples
ls examples/

# Use an example directly
./spectre verify examples/counter.spec

# Or copy examples to a working directory
mkdir -p ~/my-spectre-examples
cp examples/*.spec ~/my-spectre-examples/
cd ~/my-spectre-examples
./spectre verify counter.spec
```

### Development Setup

If you're planning to contribute or modify the code:

1. **Fork the repository** on GitHub
2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/spectre.git
   cd spectre
   ```
3. **Set up the upstream remote**:
   ```bash
   git remote add upstream https://github.com/akkeshavan/spectre.git
   ```
4. **Run tests** to ensure everything works:
   ```bash
   go test ./...
   ```

For more details on development workflow, see [README_DEV.md](../README_DEV.md).

---

## Finding Example Files After Installation

When you install Spectre using Homebrew (macOS) or the installation script (Linux), **example files are automatically included** with the installation and placed in standard system directories.

**macOS (Homebrew Installation):**

The examples are installed to the Homebrew share directory:
- **Location**: `/opt/homebrew/share/spectre/examples/` (Apple Silicon) or `/usr/local/share/spectre/examples/` (Intel)
- **Using Homebrew prefix**: `$(brew --prefix)/share/spectre/examples/`

```bash
# List all available examples:
ls $(brew --prefix)/share/spectre/examples/

# Copy examples to a working directory (recommended):
mkdir -p ~/my-spectre-examples
cp $(brew --prefix)/share/spectre/examples/*.spec ~/my-spectre-examples/
cd ~/my-spectre-examples

# Now test the examples:
spectre parse counter.spec
spectre typecheck counter.spec
spectre verify counter.spec
```


**Linux (Installation Script):**

The examples are installed to the system share directory:
- **Location**: `/usr/local/share/spectre/examples/`

```bash
# List all available examples:
ls /usr/local/share/spectre/examples/

# Copy examples to a working directory (recommended):
mkdir -p ~/my-spectre-examples
cp /usr/local/share/spectre/examples/*.spec ~/my-spectre-examples/
cd ~/my-spectre-examples

# Now test the examples:
spectre parse counter.spec
spectre typecheck counter.spec
spectre verify counter.spec
```

> **⚠️ Important**: The examples are installed in system directories. **Before testing or modifying examples, copy them to a new directory** (e.g., `~/my-spectre-examples/` or `./test-examples/`). This prevents accidental modification of system files and allows you to experiment freely.
>
> **Example:**
> ```bash
> # macOS
> mkdir -p ~/my-spectre-examples
> cp $(brew --prefix)/share/spectre/examples/*.spec ~/my-spectre-examples/
> cd ~/my-spectre-examples
> spectre parse counter.spec
>
> # Linux
> mkdir -p ~/my-spectre-examples
> cp /usr/local/share/spectre/examples/*.spec ~/my-spectre-examples/
> cd ~/my-spectre-examples
> spectre parse counter.spec
> ```

**Note**: 
- If you installed to a custom location using `INSTALL_DIR` or `SHARE_DIR` environment variables, adjust the paths accordingly.
- The examples directory contains all the example `.spec` files from the repository, so you don't need to clone the repository just to access examples.

---

## Your First Specification

Let's start with a simple counter example:

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

Save this as `counter.spec` and verify it:

```bash
spectre parse counter.spec
spectre typecheck counter.spec
spectre verify counter.spec
```

**Or use the installed example files:**

The examples are automatically installed with Spectre and available in the appropriate directories:

```bash
# macOS (Homebrew installation)
spectre parse $(brew --prefix)/share/spectre/examples/counter.spec

# Linux (installation script)
spectre parse /usr/local/share/spectre/examples/counter.spec
```

> **⚠️ Important**: Before testing or modifying examples, **copy them to a new directory** first. The examples are in system directories and should not be modified directly. See the [Finding Example Files After Installation](#finding-example-files-after-installation) section above for instructions on copying examples to a working directory.

### Understanding the Example

- **`var counter: int`**: Declares a state variable named `counter` of type `int`
- **`init { ... }`**: Defines the initial state of the system
- **`action increment { ... }`**: Defines a transition that can change the state
- **`counter' = counter + 1`**: The prime (`'`) denotes the next state value
- **`invariant nonNegative { ... }`**: A property that must always hold
- **`description "..."`**: Human-readable text that appears in error messages

---

## Next Steps

Now that you have Spectre installed and have run your first specification, you're ready to learn about the language. Continue to [Chapter 2: Language Overview](02-language-overview.md) to understand all the elements of the Spectre language.


