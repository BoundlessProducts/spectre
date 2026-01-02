#!/bin/bash
# Test the Linux install script locally on macOS (dry-run mode)
# This tests the script logic without actually installing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_SCRIPT="${SCRIPT_DIR}/install.sh"

echo "Testing Linux install script (dry-run mode)..."
echo ""

# Check if script exists
if [ ! -f "${INSTALL_SCRIPT}" ]; then
    echo "Error: install.sh not found at ${INSTALL_SCRIPT}"
    exit 1
fi

# Test 1: Check script syntax
echo "1. Checking script syntax..."
bash -n "${INSTALL_SCRIPT}" && echo "   ✓ Syntax is valid" || exit 1

# Test 2: Check for required variables
echo "2. Checking required variables..."
grep -q "REPO_URL" "${INSTALL_SCRIPT}" && echo "   ✓ REPO_URL defined" || echo "   ✗ REPO_URL missing"
grep -q "MIN_GO_VERSION" "${INSTALL_SCRIPT}" && echo "   ✓ MIN_GO_VERSION defined" || echo "   ✗ MIN_GO_VERSION missing"
grep -q "check_go_version" "${INSTALL_SCRIPT}" && echo "   ✓ Go version check function exists" || echo "   ✗ Go version check missing"

# Test 3: Check repository URL is correct
echo "3. Checking repository URL..."
if grep -q "akkeshavan/spectre" "${INSTALL_SCRIPT}"; then
    echo "   ✓ Repository URL is correct (akkeshavan/spectre)"
else
    echo "   ✗ Repository URL may be incorrect"
fi

# Test 4: Check error handling
echo "4. Checking error handling..."
grep -q "set -e" "${INSTALL_SCRIPT}" && echo "   ✓ Error handling enabled (set -e)" || echo "   ✗ Error handling may be missing"
grep -q "trap.*EXIT" "${INSTALL_SCRIPT}" && echo "   ✓ Cleanup trap exists" || echo "   ✗ Cleanup trap missing"

# Test 5: Check Go version check logic
echo "5. Checking Go version validation..."
if grep -A 10 "check_go_version" "${INSTALL_SCRIPT}" | grep -q "MIN_GO_VERSION"; then
    echo "   ✓ Go version check uses MIN_GO_VERSION"
else
    echo "   ✗ Go version check may not be working correctly"
fi

echo ""
echo "Basic script validation complete!"
echo ""
echo "To fully test, use Docker:"
echo "  ./scripts/test-linux-install.sh"

