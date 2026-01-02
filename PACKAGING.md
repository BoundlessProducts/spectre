# Packaging and Distribution Guide

This guide explains how to create installers and packages for Spectre.

## Prerequisites

- **GoReleaser** (for automated releases): `brew install goreleaser`
- **Go 1.19+** for building
- **Git** for versioning

## Release Process

### 1. Tag a Release

```bash
git tag -a v0.1.0 -m "Release version 0.1.0"
git push origin v0.1.0
```

### 2. Build Releases with GoReleaser

```bash
# Dry run (test)
goreleaser release --snapshot

# Production release
goreleaser release
```

This will create:
- Binaries for Linux, macOS, Windows (amd64, arm64)
- Homebrew formula
- Debian (.deb) packages
- RPM (.rpm) packages
- Snap packages
- Release notes

## Homebrew Formula

The Homebrew formula is in `Formula/spectre.rb`.

### Testing Locally

```bash
brew install --build-from-source Formula/spectre.rb
```

### Setting Up Homebrew Tap

1. Create repository: `github.com/spectre-lang/homebrew-spectre`
2. Copy formula to that repository
3. Users can then install with:
   ```bash
   brew tap spectre-lang/spectre
   brew install spectre
   ```

## Linux Packages

### Debian (.deb)

Built automatically by GoReleaser. Manual build:

```bash
# Install fpm if needed
gem install fpm

# Build
make build-all
fpm -s dir -t deb -n spectre -v 0.1.0 \
  --prefix /usr/local/bin \
  dist/spectre-linux-amd64=/usr/local/bin/spectre
```

### RPM (.rpm)

Built automatically by GoReleaser. Manual build:

```bash
fpm -s dir -t rpm -n spectre -v 0.1.0 \
  --prefix /usr/local/bin \
  dist/spectre-linux-amd64=/usr/local/bin/spectre
```

### Snap

Built automatically by GoReleaser. Manual build:

```bash
snapcraft
snap install spectre_*.snap --dangerous
```

## Windows Installer

### Using WiX Toolset (Advanced)

1. Install WiX Toolset
2. Create `.wxs` file
3. Build: `candle spectre.wxs && light spectre.wixobj`

### Using Inno Setup (Simpler)

1. Install Inno Setup
2. Create `.iss` script
3. Build installer

### Simple ZIP Distribution

For now, Windows users can:
1. Download the `.zip` from releases
2. Extract `spectre.exe`
3. Add to PATH manually or use `install.ps1`

## Manual Build Scripts

### Build All Platforms

```bash
make build-all
```

Creates binaries in `dist/` directory:
- `spectre-linux-amd64`
- `spectre-linux-arm64`
- `spectre-darwin-amd64`
- `spectre-darwin-arm64`
- `spectre-windows-amd64.exe`
- `spectre-windows-arm64.exe`

### Build Single Platform

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o spectre-linux ./cmd/spectre

# macOS
GOOS=darwin GOARCH=amd64 go build -o spectre-darwin ./cmd/spectre

# Windows
GOOS=windows GOARCH=amd64 go build -o spectre-windows.exe ./cmd/spectre
```

## Installation Scripts

### Linux (`scripts/install.sh`)

Features:
- Checks for Go
- Clones repository
- Builds from source
- Installs to `/usr/local/bin`
- Verifies installation

Usage:
```bash
curl -fsSL https://raw.githubusercontent.com/spectre-lang/spectre/main/scripts/install.sh | bash
```

### Windows (`scripts/install.ps1`)

Features:
- Checks for Go
- Clones repository
- Builds from source
- Installs to `C:\Program Files\Spectre`
- Adds to PATH

Usage:
```powershell
powershell -ExecutionPolicy Bypass -File scripts/install.ps1
```

## CI/CD Integration

### GitHub Actions

Example workflow for automated releases:

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - uses: goreleaser/goreleaser-action@v3
        with:
          version: latest
          args: release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Version Management

Version is defined in `cmd/spectre/main.go`:

```go
const Version = "0.1.0"
```

Update this for each release.

## Distribution Checklist

- [ ] Update version in `main.go`
- [ ] Update `CHANGELOG.md`
- [ ] Tag release: `git tag v0.1.0`
- [ ] Push tag: `git push origin v0.1.0`
- [ ] Run GoReleaser: `goreleaser release`
- [ ] Verify all packages build correctly
- [ ] Test installation on each platform
- [ ] Update documentation with new version
- [ ] Announce release

## Testing Installers

### macOS
```bash
brew install --build-from-source Formula/spectre.rb
spectre --version
```

### Linux
```bash
./scripts/install.sh
spectre --version
```

### Windows
```powershell
.\scripts\install.ps1
spectre --version
```

## Troubleshooting

### GoReleaser Issues

- Ensure `GITHUB_TOKEN` is set
- Check `.goreleaser.yml` syntax
- Run with `--snapshot` first to test

### Homebrew Formula Issues

- Test locally first: `brew install --build-from-source`
- Check formula syntax: `brew audit Formula/spectre.rb`
- Ensure URL and SHA256 are correct

### Package Build Issues

- Ensure all dependencies are installed
- Check file permissions
- Verify binary works before packaging

