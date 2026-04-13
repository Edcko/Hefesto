#!/bin/bash
# Hefesto Multi-Distro E2E Test Runner
# Runs basic validation across all supported Linux distributions
#
# Usage:
#   ./scripts/e2e-multi-distro.sh           # Test all distros
#   ./scripts/e2e-multi-distro.sh debian    # Test specific distro
#
# Supported: ubuntu, debian, fedora, alpine
set -e

# Resolve paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Distro-to-service mapping
declare -A DISTRO_SERVICE=(
    ["ubuntu"]="hefesto-test"
    ["debian"]="hefesto-debian"
    ["fedora"]="hefesto-fedora"
    ["alpine"]="hefesto-alpine"
)

# Distro-to-Dockerfile mapping
declare -A DISTRO_DOCKERFILE=(
    ["ubuntu"]="Dockerfile.test"
    ["debian"]="Dockerfile.debian"
    ["fedora"]="Dockerfile.fedora"
    ["alpine"]="Dockerfile.alpine"
)

# Results tracking
declare -A RESULTS
TOTAL_PASS=0
TOTAL_FAIL=0

# ─── Helpers ────────────────────────────────────────────────────────────

header() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

log_pass() {
    echo -e "  ${GREEN}✅ PASS${NC}: $1"
    TOTAL_PASS=$((TOTAL_PASS + 1))
}

log_fail() {
    echo -e "  ${RED}❌ FAIL${NC}: $1"
    echo -e "    ${YELLOW}Details${NC}: $2"
    TOTAL_FAIL=$((TOTAL_FAIL + 1))
}

log_info() {
    echo -e "  ${CYAN}ℹ️${NC} $1"
}

# ─── Test a single distro ──────────────────────────────────────────────

test_distro() {
    local DISTRO="$1"
    local SERVICE="${DISTRO_SERVICE[$DISTRO]}"
    local DOCKERFILE="${DISTRO_DOCKERFILE[$DISTRO]}"

    if [ -z "$SERVICE" ]; then
        echo -e "${RED}Unknown distro: $DISTRO${NC}"
        echo "Supported: ${!DISTRO_SERVICE[*]}"
        exit 1
    fi

    header "🧪 Testing: $DISTRO (service: $SERVICE)"

    # Step 1: Build container
    log_info "Building container from $DOCKERFILE..."
    if ! docker compose -f "$PROJECT_DIR/docker-compose.yml" build "$SERVICE" 2>&1 | tail -5; then
        log_fail "Build container" "docker compose build $SERVICE failed"
        RESULTS["$DISTRO"]="BUILD_FAILED"
        return 1
    fi
    log_pass "Container built successfully"

    # Step 2: Start container
    log_info "Starting container..."
    if ! docker compose -f "$PROJECT_DIR/docker-compose.yml" up -d "$SERVICE" 2>&1; then
        log_fail "Start container" "docker compose up $SERVICE failed"
        RESULTS["$DISTRO"]="START_FAILED"
        return 1
    fi

    # Wait for container to be ready
    sleep 2

    # Verify container is running
    local CONTAINER_STATUS
    CONTAINER_STATUS=$(docker inspect -f '{{.State.Status}}' "$SERVICE" 2>/dev/null || echo "not found")
    if [ "$CONTAINER_STATUS" != "running" ]; then
        log_fail "Container running" "Status: $CONTAINER_STATUS"
        RESULTS["$DISTRO"]="NOT_RUNNING"
        docker compose -f "$PROJECT_DIR/docker-compose.yml" down "$SERVICE" 2>/dev/null
        return 1
    fi
    log_pass "Container is running"

    # Step 3: Build and deploy Hefesto binary
    log_info "Building Hefesto binary..."
    if ! GOOS=linux GOARCH=amd64 go build -o /tmp/hefesto-e2e-multi "$PROJECT_DIR/cmd/hefesto" 2>&1; then
        log_fail "Build Hefesto binary" "Go build failed"
        RESULTS["$DISTRO"]="BINARY_BUILD_FAILED"
        docker compose -f "$PROJECT_DIR/docker-compose.yml" down "$SERVICE" 2>/dev/null
        return 1
    fi

    log_info "Deploying binary to container..."
    if ! docker cp /tmp/hefesto-e2e-multi "$SERVICE":/usr/local/bin/hefesto 2>&1; then
        log_fail "Deploy binary" "docker cp failed"
        RESULTS["$DISTRO"]="DEPLOY_FAILED"
        docker compose -f "$PROJECT_DIR/docker-compose.yml" down "$SERVICE" 2>/dev/null
        return 1
    fi
    docker exec "$SERVICE" chmod +x /usr/local/bin/hefesto
    log_pass "Binary deployed"

    # Step 4: Test --version
    local VERSION_OUT
    VERSION_OUT=$(docker exec "$SERVICE" hefesto version 2>&1)
    if [ $? -eq 0 ]; then
        log_pass "hefesto version: $(echo "$VERSION_OUT" | head -1)"
    else
        log_fail "hefesto version" "$VERSION_OUT"
    fi

    # Step 5: Test status on clean state
    docker exec "$SERVICE" sh -c "rm -rf /root/.config/opencode /root/.config/opencode-backup-*" 2>/dev/null
    local STATUS_OUT
    STATUS_OUT=$(docker exec "$SERVICE" hefesto status 2>&1)
    if echo "$STATUS_OUT" | grep -q "Not installed"; then
        log_pass "Status shows 'Not installed' on clean state"
    else
        log_fail "Status on clean state" "$(echo "$STATUS_OUT" | head -3)"
    fi

    # Step 6: Test doctor on clean state
    local DOCTOR_OUT
    DOCTOR_OUT=$(docker exec "$SERVICE" hefesto doctor 2>&1)
    if echo "$DOCTOR_OUT" | grep -qi "issue\|not installed\|Run.*install"; then
        log_pass "Doctor detects issues on clean state"
    else
        log_fail "Doctor on clean state" "$(echo "$DOCTOR_OUT" | head -3)"
    fi

    # Step 7: Test install
    local INSTALL_OUT
    INSTALL_OUT=$(docker exec "$SERVICE" hefesto install --yes 2>&1)
    if echo "$INSTALL_OUT" | grep -q "installed successfully"; then
        log_pass "Install succeeds"
    else
        log_fail "Install" "$(echo "$INSTALL_OUT" | head -3)"
    fi

    # Step 8: Test status after install
    local STATUS_AFTER
    STATUS_AFTER=$(docker exec "$SERVICE" hefesto status 2>&1)
    if echo "$STATUS_AFTER" | grep -q "Installed"; then
        log_pass "Status shows installed after install"
    else
        log_fail "Status after install" "$(echo "$STATUS_AFTER" | head -3)"
    fi

    # Step 9: Test doctor after install
    local DOCTOR_AFTER
    DOCTOR_AFTER=$(docker exec "$SERVICE" hefesto doctor 2>&1)
    if echo "$DOCTOR_AFTER" | grep -q "All checks passed"; then
        log_pass "Doctor passes all checks after install"
    else
        log_fail "Doctor after install" "$(echo "$DOCTOR_AFTER" | head -3)"
    fi

    # Step 10: Test uninstall
    local UNINSTALL_OUT
    UNINSTALL_OUT=$(docker exec "$SERVICE" hefesto uninstall --yes --purge 2>&1)
    if echo "$UNINSTALL_OUT" | grep -q "removed"; then
        log_pass "Uninstall with purge succeeds"
    else
        log_fail "Uninstall" "$(echo "$UNINSTALL_OUT" | head -3)"
    fi

    # Step 11: Verify opencode binary works
    local OPENCODE_OUT
    OPENCODE_OUT=$(docker exec "$SERVICE" opencode version 2>&1 || docker exec "$SERVICE" opencode --version 2>&1)
    if [ $? -eq 0 ]; then
        log_pass "opencode binary works"
    else
        log_fail "opencode binary" "$(echo "$OPENCODE_OUT" | head -2)"
    fi

    # Step 12: Verify engram binary works
    local ENGRAM_OUT
    ENGRAM_OUT=$(docker exec "$SERVICE" engram version 2>&1)
    if [ $? -eq 0 ]; then
        log_pass "engram binary works: $(echo "$ENGRAM_OUT" | head -1)"
    else
        log_fail "engram binary" "$(echo "$ENGRAM_OUT" | head -2)"
    fi

    # Cleanup
    log_info "Stopping container..."
    docker compose -f "$PROJECT_DIR/docker-compose.yml" down "$SERVICE" 2>/dev/null

    RESULTS["$DISTRO"]="DONE"
    log_info "$DISTRO testing complete"
}

# ─── Main ───────────────────────────────────────────────────────────────

header "🔥 Hefesto Multi-Distro E2E Tests"

# Determine which distros to test
if [ $# -gt 0 ]; then
    DISTROS=("$@")
else
    DISTROS=("ubuntu" "debian" "fedora" "alpine")
fi

log_info "Testing distros: ${DISTROS[*]}"
echo ""

# Run tests for each distro
for DISTRO in "${DISTROS[@]}"; do
    test_distro "$DISTRO"
    echo ""
done

# ─── Summary ────────────────────────────────────────────────────────────

header "📊 Multi-Distro Test Summary"

for DISTRO in "${DISTROS[@]}"; do
    STATUS="${RESULTS[$DISTRO]:-UNKNOWN}"
    if [ "$STATUS" = "DONE" ]; then
        echo -e "  ${GREEN}✅ $DISTRO${NC} — completed"
    else
        echo -e "  ${RED}❌ $DISTRO${NC} — $STATUS"
    fi
done

echo ""
echo -e "  Total: ${GREEN}$TOTAL_PASS passed${NC}, ${RED}$TOTAL_FAIL failed${NC}"
echo ""

if [ $TOTAL_FAIL -eq 0 ]; then
    echo -e "${GREEN}✅ All multi-distro tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed — review output above${NC}"
    exit 1
fi
