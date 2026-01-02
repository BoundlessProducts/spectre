#!/bin/bash
# Uninstallation script for Spectre on Linux

set -e

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

echo "Uninstalling Spectre..."

if [ -f "${INSTALL_DIR}/spectre" ]; then
    sudo rm -f "${INSTALL_DIR}/spectre"
    echo "✓ Spectre uninstalled successfully!"
else
    echo "Spectre not found at ${INSTALL_DIR}/spectre"
    exit 1
fi

