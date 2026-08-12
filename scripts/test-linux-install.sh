#!/bin/bash
# Test script for Linux installation on macOS using Docker

set -e

echo "Testing Linux installation script using Docker..."
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed."
    echo ""
    echo "Install Docker Desktop for Mac from: https://www.docker.com/products/docker-desktop"
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo "Error: Docker is not running."
    echo "Please start Docker Desktop and try again."
    exit 1
fi

echo "✓ Docker is installed and running"
echo ""

# Choose Linux distribution (default: ubuntu)
DISTRO="${1:-ubuntu}"
echo "Testing on ${DISTRO}..."
echo ""

# Create a test script that will be run inside the container
cat > /tmp/test-install.sh << 'EOF'
#!/bin/bash
set -e

# Install prerequisites
apt-get update -qq
apt-get install -y -qq git curl

# Run the install script
curl -fsSL https://raw.githubusercontent.com/BoundlessProducts/spectre/main/scripts/install.sh | bash

# Verify installation
echo ""
echo "Verifying installation..."
spectre --version || echo "Version check failed"
EOF

chmod +x /tmp/test-install.sh

# Run the test in a Docker container
echo "Starting Docker container..."
docker run --rm -it \
    -v /tmp/test-install.sh:/test-install.sh:ro \
    "${DISTRO}:latest" \
    bash /test-install.sh

echo ""
echo "Test complete!"

