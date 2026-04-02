#!/bin/bash
# Hefesto OpenCode Testing Environment - Setup Script
# Builds and starts the Docker container for testing Hefesto configuration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HEFESTO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "═══════════════════════════════════════════════════════════════"
echo "  🔥 Hefesto OpenCode Testing Environment"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Navigate to Hefesto root
cd "$HEFESTO_ROOT"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running. Please start Docker first."
    exit 1
fi

echo "📦 Building Docker image..."
docker build -f Dockerfile.test -t hefesto-test .

if [ $? -ne 0 ]; then
    echo "❌ Build failed. Check the Dockerfile.test for errors."
    exit 1
fi

echo ""
echo "🚀 Starting container..."
docker compose up -d

if [ $? -ne 0 ]; then
    echo "❌ Failed to start container."
    exit 1
fi

# Wait for container to be ready
echo "⏳ Waiting for container to start..."
sleep 2

# Verify container is running
if ! docker ps | grep -q hefesto-test; then
    echo "❌ Container is not running. Check logs with: docker compose logs"
    exit 1
fi

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  ✅ Hefesto test environment is ready!"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "📖 HOW TO TEST:"
echo ""
echo "  1️⃣  Enter the container:"
echo "      docker exec -it hefesto-test bash"
echo ""
echo "  2️⃣  Run OpenCode (inside container):"
echo "      opencode"
echo ""
echo "  3️⃣  Test Hefesto personality:"
echo "      Type: 'hola, ¿quién eres?' or 'explain SOLID principles'"
echo "      Expected: Hefesto's warm, mentor-like personality"
echo ""
echo "  4️⃣  Test SDD workflow:"
echo "      Type: /sdd-init"
echo "      Expected: SDD initialization process"
echo ""
echo "  5️⃣  Exit OpenCode:"
echo "      Press Ctrl+C or type /exit"
echo ""
echo "  6️⃣  Exit container:"
echo "      exit"
echo ""
echo "  7️⃣  Stop and clean up:"
echo "      ./scripts/clean.sh"
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  🔧 Useful commands:"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "  View logs:     docker compose logs -f"
echo "  Restart:       docker compose restart"
echo "  Shell access:  docker exec -it hefesto-test bash"
echo "  Check status:  docker ps | grep hefesto-test"
echo ""
echo "═══════════════════════════════════════════════════════════════"
