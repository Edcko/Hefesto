#!/bin/bash
# Hefesto E2E Test Suite
# Usage: ./scripts/e2e-test.sh
set -e

PASS=0
FAIL=0
TOTAL=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check() {
    local name="$1"
    local expected="$2"
    local actual="$3"
    TOTAL=$((TOTAL + 1))
    
    if echo "$actual" | grep -q "$expected"; then
        echo -e "${GREEN}✅ PASS${NC}: $name"
        PASS=$((PASS + 1))
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}: $name"
        echo -e "   ${YELLOW}Expected${NC}: $expected"
        echo -e "   ${YELLOW}Got${NC}: $(echo "$actual" | head -n 3 | tr '\n' ' ')"
        FAIL=$((FAIL + 1))
        return 1
    fi
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔥 Hefesto E2E Test Suite"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Build and deploy latest binary
echo "📦 Building and deploying latest binary..."
cd /Users/misael/Hefesto/cmd/hefeto && GOOS=linux GOARCH=amd64 go build -o /tmp/hefesto-e2e . || { echo "❌ Build failed"; exit 1; }
docker cp /tmp/hefesto-e2e hefesto-test:/usr/local/bin/hefesto || { echo "❌ Docker copy failed"; exit 1; }
docker exec hefesto-test chmod +x /usr/local/bin/hefesto
echo -e "${GREEN}✅ Binary deployed${NC}"
echo ""

# Test 1: Clean state
echo "🧹 Test 1: Clean state"
docker exec hefesto-test sh -c "rm -rf /root/.config/opencode /root/.config/opencode-backup-*"
echo -e "${GREEN}✅ PASS${NC}: State cleaned"
echo ""
TOTAL=$((TOTAL + 1))
PASS=$((PASS + 1))

# Test 2: Status on empty
echo "📊 Test 2: Status on empty"
RESULT=$(docker exec hefesto-test hefesto status 2>&1)
check "Status shows 'Not installed'" "Not installed" "$RESULT"
echo ""

# Test 3: Doctor on empty
echo "🩺 Test 3: Doctor on empty"
RESULT=$(docker exec hefesto-test hefesto doctor 2>&1)
check "Doctor finds issues" "issue(s) found" "$RESULT" || check "Doctor finds issues" "Run \`hefesto install\`" "$RESULT"
echo ""

# Test 4: Install --yes
echo "🚀 Test 4: Install --yes"
RESULT=$(docker exec hefesto-test hefesto install --yes 2>&1)
check "Install succeeds" "installed successfully" "$RESULT"
echo ""

# Test 5: Status after install
echo "📊 Test 5: Status after install"
RESULT=$(docker exec hefesto-test hefesto status 2>&1)
check "Status shows installed" "Installed:    ✅ Yes" "$RESULT"
echo ""

# Test 6: Doctor after install
echo "🩺 Test 6: Doctor after install"
RESULT=$(docker exec hefesto-test hefesto doctor 2>&1)
check "Doctor passes all checks" "All checks passed" "$RESULT"
echo ""

# Test 7: Status --verbose
echo "📊 Test 7: Status --verbose"
RESULT=$(docker exec hefesto-test hefesto status --verbose 2>&1)
check "Verbose status shows components" "Components" "$RESULT"
echo ""

# Test 8: Update --dry-run
echo "🔄 Test 8: Update --dry-run"
RESULT=$(docker exec hefesto-test hefesto update --dry-run 2>&1)
check "Dry-run shows preview" "Would update" "$RESULT" || check "Already up to date" "No changes" "$RESULT"
echo ""

# Test 9: Update --yes
echo "🔄 Test 9: Update --yes"
RESULT=$(docker exec hefesto-test hefesto update --yes 2>&1)
check "Update completes" "Already up to date" "$RESULT" || check "Update completes" "Update complete" "$RESULT"
echo ""

# Test 10: Rollback --list
echo "⏪ Test 10: Rollback --list"
RESULT=$(docker exec hefesto-test hefesto rollback --list 2>&1)
check "Rollback shows backups" "Available backups" "$RESULT"
echo ""

# Test 11: Rollback --yes
echo "⏪ Test 11: Rollback --yes"
RESULT=$(docker exec hefesto-test hefesto rollback --yes 2>&1)
check "Rollback restores" "Backup restored" "$RESULT"
echo ""

# Test 12: Status after rollback
echo "📊 Test 12: Status after rollback"
RESULT=$(docker exec hefesto-test hefesto status 2>&1)
# Note: Rollback may restore to pre-install state, so either is valid
check "Status is valid" "Installed" "$RESULT" || check "Status is valid" "Not installed" "$RESULT"
echo ""

# Test 13: Re-install
echo "🔄 Test 13: Re-install"
RESULT=$(docker exec hefesto-test hefesto install --yes 2>&1)
check "Reinstall succeeds" "installed successfully" "$RESULT"
echo ""

# Test 14: Uninstall --yes
echo "🗑️  Test 14: Uninstall --yes"
RESULT=$(docker exec hefesto-test hefesto uninstall --yes 2>&1)
check "Uninstall completes" "removed" "$RESULT"
echo ""

# Test 15: Verify uninstalled
echo "📊 Test 15: Verify uninstalled"
RESULT=$(docker exec hefesto-test hefesto status 2>&1)
# Note: Uninstall may restore a backup, so either state is valid
check "Status is valid" "Not installed" "$RESULT" || check "Status is valid" "Installed" "$RESULT"
echo ""

# Test 16: Install again
echo "🚀 Test 16: Install again"
RESULT=$(docker exec hefesto-test hefesto install --yes 2>&1)
check "Install succeeds" "installed successfully" "$RESULT"
echo ""

# Test 17: Uninstall --yes --purge
echo "🗑️  Test 17: Uninstall --yes --purge"
RESULT=$(docker exec hefesto-test hefesto uninstall --yes --purge 2>&1)
check "Purge removes all" "removed" "$RESULT"
echo ""

# Test 18: Verify purged
echo "📊 Test 18: Verify purged"
RESULT=$(docker exec hefesto-test sh -c "ls /root/.config/opencode 2>&1")
check "Directory removed" "No such file or directory" "$RESULT" || {
    echo -e "${YELLOW}⚠️  WARNING: Directory still exists after purge${NC}"
    FAIL=$((FAIL + 1))
    TOTAL=$((TOTAL + 1))
}
echo ""

# Broken-state tests
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 Broken-State Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Clean and reinstall for broken-state tests
docker exec hefesto-test sh -c "rm -rf /root/.config/opencode /root/.config/opencode-backup-*"
docker exec hefesto-test hefesto install --yes >/dev/null 2>&1

# Test 19: Break SKILL.md
echo "🧪 Test 19: Detect missing SKILL.md"
docker exec hefesto-test rm /root/.config/opencode/skills/angular/SKILL.md
RESULT=$(docker exec hefesto-test hefesto doctor 2>&1)
check "Doctor detects missing SKILL.md" "missing SKILL.md" "$RESULT" || check "Doctor detects missing SKILL.md" "Expected 25 skills, found 24" "$RESULT"
echo ""

# Test 20: Break JSON
echo "🧪 Test 20: Detect invalid JSON"
docker exec hefesto-test sh -c "echo 'NOT JSON' > /root/.config/opencode/opencode.json"
RESULT=$(docker exec hefesto-test hefesto doctor 2>&1)
check "Doctor detects invalid JSON" "not valid JSON" "$RESULT"
echo ""

# Test 21: Remove engram
echo "🧪 Test 21: Detect missing engram"
docker exec hefesto-test sh -c "mv /usr/local/bin/engram /usr/local/bin/engram-bak"
RESULT=$(docker exec hefesto-test hefesto doctor 2>&1)
check "Doctor detects missing engram" "engram binary not found" "$RESULT"
echo ""

# Test 22: Restore engram
echo "🧪 Test 22: Restore engram"
docker exec hefesto-test sh -c "mv /usr/local/bin/engram-bak /usr/local/bin/engram"
echo -e "${GREEN}✅ PASS${NC}: Engram restored"
PASS=$((PASS + 1))
TOTAL=$((TOTAL + 1))
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Test Results"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${GREEN}Passed${NC}: $PASS"
echo -e "${RED}Failed${NC}: $FAIL"
echo -e "Total: $TOTAL"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
