#!/bin/bash
# Installation script for Spectre on Linux
# This script downloads the source code and builds Spectre locally

set -e

VERSION="0.1.0"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
REPO_URL="https://github.com/akkeshavan/spectre.git"
BRANCH="${BRANCH:-main}"
MIN_GO_VERSION="1.19"

echo "Installing Spectre ${VERSION}..."
echo "This will download the source code and build Spectre locally."
echo ""

# Function to check Go version
check_go_version() {
    if ! command -v go &> /dev/null; then
        return 1
    fi
    
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    local major=$(echo "$go_version" | cut -d. -f1)
    local minor=$(echo "$go_version" | cut -d. -f2)
    local min_major=$(echo "$MIN_GO_VERSION" | cut -d. -f1)
    local min_minor=$(echo "$MIN_GO_VERSION" | cut -d. -f2)
    
    if [ "$major" -lt "$min_major" ] || ([ "$major" -eq "$min_major" ] && [ "$minor" -lt "$min_minor" ]); then
        return 1
    fi
    return 0
}

# Check if Go is installed and meets version requirement
if ! check_go_version; then
    echo "Error: Go ${MIN_GO_VERSION} or later is required but not found."
    echo ""
    echo "Please install Go first:"
    echo ""
    echo "  Ubuntu/Debian:"
    echo "    sudo apt-get update"
    echo "    sudo apt-get install golang-go"
    echo ""
    echo "  Fedora/RHEL/CentOS:"
    echo "    sudo dnf install golang"
    echo ""
    echo "  Or download from: https://golang.org/dl/"
    echo ""
    echo "After installing Go, run this script again."
    exit 1
fi

# Display Go version
GO_VERSION=$(go version)
echo "✓ Found Go: ${GO_VERSION}"
echo ""

# Check if git is installed
if ! command -v git &> /dev/null; then
    echo "Error: Git is not installed. Please install Git first."
    echo ""
    echo "  Ubuntu/Debian: sudo apt-get install git"
    echo "  Fedora/RHEL/CentOS: sudo dnf install git"
    exit 1
fi

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

# Clone repository
echo "Downloading Spectre source code from ${REPO_URL}..."
cd "${TMP_DIR}"
if ! git clone -b "${BRANCH}" "${REPO_URL}" spectre 2>/dev/null; then
    echo "Error: Failed to clone repository. Check your internet connection."
    exit 1
fi
cd spectre

# Build
echo "Building Spectre from source..."
echo "This may take a few moments..."
if ! go build -o spectre ./cmd/spectre; then
    echo "Error: Build failed. Please check the error messages above."
    exit 1
fi

# Verify binary was created
if [ ! -f "./spectre" ]; then
    echo "Error: Build completed but binary not found."
    exit 1
fi

# Install
echo ""
echo "Installing to ${INSTALL_DIR}..."
sudo mkdir -p "${INSTALL_DIR}"
sudo cp spectre "${INSTALL_DIR}/spectre"
sudo chmod +x "${INSTALL_DIR}/spectre"

# Verify installation
echo ""
if command -v spectre &> /dev/null; then
    echo "✓ Spectre installed successfully!"
    echo ""
    spectre --version || echo "Installation complete. Run 'spectre --help' for usage."
else
    echo "✓ Binary installed to ${INSTALL_DIR}/spectre"
    echo ""
    echo "Note: Make sure ${INSTALL_DIR} is in your PATH."
    echo "Add this to your ~/.bashrc or ~/.zshrc if needed:"
    echo "  export PATH=\"\${PATH}:${INSTALL_DIR}\""
    echo ""
    echo "Then run: source ~/.bashrc  (or source ~/.zshrc)"
fi

echo ""
echo "Installation complete!"

