#!/bin/bash
# Hefesto OpenCode Testing Environment - Cleanup Script
# Stops and removes the Docker container and image

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HEFESTO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "═══════════════════════════════════════════════════════════════"
echo "  🧹 Hefesto OpenCode - Cleanup"
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Navigate to Hefesto root
cd "$HEFESTO_ROOT"

echo "🛑 Stopping container..."
docker compose down 2>/dev/null || true

echo "🗑️  Removing image..."
docker rmi hefesto-test 2>/dev/null || true

echo "🧹 Cleaning test workspace..."
# Keep the directory but clean contents (except .gitkeep if exists)
if [ -d "test-workspace" ]; then
    # Don't delete .gitkeep
    find test-workspace -mindepth 1 ! -name '.gitkeep' -delete 2>/dev/null || true
fi

# Clean up any dangling images/containers
echo "🗑️  Cleaning up Docker resources..."
docker container prune -f 2>/dev/null || true

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  ✅ Cleanup complete!"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "To start again, run: ./scripts/test.sh"
echo ""
