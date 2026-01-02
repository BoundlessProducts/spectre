#!/bin/bash
# Installation script for Spectre on Linux

set -e

VERSION="0.1.0"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
REPO_URL="https://github.com/spectre-lang/spectre"

echo "Installing Spectre ${VERSION}..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    echo "Visit: https://golang.org/dl/"
    exit 1
fi

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

# Clone repository
echo "Downloading Spectre..."
cd "${TMP_DIR}"
git clone "${REPO_URL}" spectre
cd spectre

# Build
echo "Building Spectre..."
go build -o spectre ./cmd/spectre

# Install
echo "Installing to ${INSTALL_DIR}..."
sudo mkdir -p "${INSTALL_DIR}"
sudo cp spectre "${INSTALL_DIR}/spectre"
sudo chmod +x "${INSTALL_DIR}/spectre"

# Verify installation
if command -v spectre &> /dev/null; then
    echo "✓ Spectre installed successfully!"
    spectre --version || echo "Installation complete. Run 'spectre --help' for usage."
else
    echo "Warning: spectre command not found in PATH"
    echo "Make sure ${INSTALL_DIR} is in your PATH"
fi

