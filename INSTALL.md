# Installation Guide

This guide provides detailed installation instructions for Spectre on macOS, Linux, and Windows.

## Prerequisites

- **Go 1.19 or later** (for building from source)
- **Git** (for cloning the repository)

## macOS

### Option 1: Homebrew (Recommended)

```bash
# Add the tap
brew tap spectre-lang/spectre

# Install
brew install spectre
```

### Option 2: Build from Source

```bash
git clone https://github.com/spectre-lang/spectre.git
cd spectre
go build -o spectre ./cmd/spectre
sudo mv spectre /usr/local/bin/
```

### Uninstall

```bash
brew uninstall spectre
```

## Linux

### Option 1: Install Script (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/spectre-lang/spectre/main/scripts/install.sh | bash
```

The script will:
- Check for Go installation
- Clone the repository
- Build Spectre
- Install to `/usr/local/bin` (or custom `INSTALL_DIR`)

### Option 2: Package Managers

#### Debian/Ubuntu (.deb)

```bash
# Download the .deb package from releases
wget https://github.com/spectre-lang/spectre/releases/download/v0.1.0/spectre_0.1.0_linux_amd64.deb
sudo dpkg -i spectre_0.1.0_linux_amd64.deb
```

#### RHEL/CentOS/Fedora (.rpm)

```bash
# Download the .rpm package from releases
wget https://github.com/spectre-lang/spectre/releases/download/v0.1.0/spectre_0.1.0_linux_amd64.rpm
sudo rpm -i spectre_0.1.0_linux_amd64.rpm
```

#### Snap

```bash
sudo snap install spectre
```

### Option 3: Build from Source

```bash
git clone https://github.com/spectre-lang/spectre.git
cd spectre
go build -o spectre ./cmd/spectre
sudo mv spectre /usr/local/bin/
```

### Uninstall

**Using install script:**
```bash
curl -fsSL https://raw.githubusercontent.com/spectre-lang/spectre/main/scripts/uninstall.sh | bash
```

**Manual:**
```bash
sudo rm /usr/local/bin/spectre
```

## Windows

### Option 1: PowerShell Script (Recommended)

Open PowerShell as Administrator and run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/install.ps1
```

The script will:
- Check for Go installation
- Clone the repository
- Build Spectre
- Install to `C:\Program Files\Spectre`
- Add to PATH

### Option 2: Build from Source

```powershell
git clone https://github.com/spectre-lang/spectre.git
cd spectre
go build -o spectre.exe ./cmd/spectre
# Manually add to PATH or use install.ps1
```

### Option 3: Chocolatey (Coming Soon)

```powershell
choco install spectre
```

### Uninstall

**Using PowerShell script:**
```powershell
powershell -ExecutionPolicy Bypass -File scripts/uninstall.ps1
```

**Manual:**
1. Delete `C:\Program Files\Spectre`
2. Remove from PATH environment variable

## Custom Installation Directory

### Linux

```bash
INSTALL_DIR=/opt/spectre scripts/install.sh
```

### Windows

Edit `scripts/install.ps1` and change `$InstallDir` variable.

## Verify Installation

After installation, verify it works:

```bash
spectre --version
```

Expected output:
```
spectre version 0.1.0
```

Test with a command:
```bash
spectre parse examples/counter.spec
```

## Troubleshooting

### Command Not Found

**Linux/macOS:**
- Ensure `/usr/local/bin` (or your `INSTALL_DIR`) is in your PATH
- Check: `echo $PATH`
- Add to `~/.bashrc` or `~/.zshrc`: `export PATH=$PATH:/usr/local/bin`

**Windows:**
- Restart terminal after installation
- Check PATH: `$env:PATH`
- Manually add `C:\Program Files\Spectre` to PATH if needed

### Go Not Found

Install Go from https://golang.org/dl/

**macOS:**
```bash
brew install go
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install golang-go

# Fedora
sudo dnf install golang
```

**Windows:**
Download installer from https://golang.org/dl/

### Permission Denied

**Linux/macOS:**
- Use `sudo` for installation scripts
- Ensure install directory is writable

**Windows:**
- Run PowerShell as Administrator

## Building from Source

For all platforms:

```bash
git clone https://github.com/spectre-lang/spectre.git
cd spectre
go build -o spectre ./cmd/spectre
```

The binary will be created in the current directory.

## Development Installation

For development, use Go's install command:

```bash
go install ./cmd/spectre
```

This installs to `$GOPATH/bin` or `$HOME/go/bin`.

## Updating

### Homebrew (macOS)
```bash
brew upgrade spectre
```

### Linux (from source)
```bash
cd spectre
git pull
go build -o spectre ./cmd/spectre
sudo mv spectre /usr/local/bin/
```

### Windows
Re-run the install script or rebuild from source.

## Next Steps

After installation, see:
- [USAGE.md](./USAGE.md) - CLI usage guide
- [README.md](./README.md) - Overview and examples
- [SPEC.md](./SPEC.md) - Language specification

