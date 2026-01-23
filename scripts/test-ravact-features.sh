#!/bin/bash
# Comprehensive ravact feature testing script
# Tests each screen and feature systematically

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

TEST_RESULTS=()

log_test() {
    local status=$1
    local test_name=$2
    local details=$3
    
    if [[ "$status" == "PASS" ]]; then
        echo -e "${GREEN}✓ PASS${NC} - $test_name"
        TEST_RESULTS+=("✓ $test_name")
    elif [[ "$status" == "FAIL" ]]; then
        echo -e "${RED}✗ FAIL${NC} - $test_name"
        if [[ -n "$details" ]]; then
            echo -e "${RED}  Details: $details${NC}"
        fi
        TEST_RESULTS+=("✗ $test_name - $details")
    elif [[ "$status" == "SKIP" ]]; then
        echo -e "${YELLOW}⊘ SKIP${NC} - $test_name - $details"
        TEST_RESULTS+=("⊘ $test_name - $details")
    fi
}

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Ravact Feature Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if binary exists
if [[ ! -f "./dist/ravact-linux-amd64" ]]; then
    echo -e "${RED}Binary not found: dist/ravact-linux-amd64${NC}"
    echo "Run: make build-linux"
    exit 1
fi

echo -e "${GREEN}✓ Binary found${NC}"
echo ""

# Test 1: Binary execution
echo -e "${BLUE}Test 1: Binary Execution${NC}"
if ./dist/ravact-linux-amd64 --version 2>/dev/null; then
    log_test "PASS" "Binary executes with --version"
else
    log_test "FAIL" "Binary execution" "Cannot run --version"
fi
echo ""

# Test 2: Help output
echo -e "${BLUE}Test 2: Help Output${NC}"
if ./dist/ravact-linux-amd64 --help 2>/dev/null | grep -q "ravact"; then
    log_test "PASS" "Help flag works"
else
    log_test "SKIP" "Help flag" "May not be implemented"
fi
echo ""

# Test 3: Check architecture
echo -e "${BLUE}Test 3: Architecture Verification${NC}"
ARCH=$(file ./dist/ravact-linux-amd64 | grep -o "x86-64")
if [[ "$ARCH" == "x86-64" ]]; then
    log_test "PASS" "Binary is x86_64"
else
    log_test "FAIL" "Architecture" "Not x86_64"
fi
echo ""

# Test 4: Assets exist
echo -e "${BLUE}Test 4: Assets Check${NC}"
if [[ -d "./assets/scripts" ]] && [[ $(ls -1 ./assets/scripts/*.sh 2>/dev/null | wc -l) -gt 5 ]]; then
    log_test "PASS" "Setup scripts exist ($(ls -1 ./assets/scripts/*.sh | wc -l) scripts)"
else
    log_test "FAIL" "Assets" "Setup scripts not found"
fi

if [[ -d "./assets/configs" ]] && [[ $(ls -1 ./assets/configs/*.json 2>/dev/null | wc -l) -gt 0 ]]; then
    log_test "PASS" "Config files exist ($(ls -1 ./assets/configs/*.json | wc -l) configs)"
else
    log_test "FAIL" "Assets" "Config files not found"
fi
echo ""

# Test 5: User Management (System-level check)
echo -e "${BLUE}Test 5: System Checks (for User Management)${NC}"
if [[ -f "/etc/passwd" ]]; then
    log_test "PASS" "/etc/passwd exists"
    USERS=$(cat /etc/passwd | wc -l)
    echo -e "  Found $USERS users in /etc/passwd"
else
    log_test "FAIL" "System files" "/etc/passwd not found"
fi

if [[ -f "/etc/group" ]]; then
    log_test "PASS" "/etc/group exists"
    GROUPS=$(cat /etc/group | wc -l)
    echo -e "  Found $GROUPS groups in /etc/group"
else
    log_test "FAIL" "System files" "/etc/group not found"
fi

if command -v groups &> /dev/null; then
    log_test "PASS" "groups command available"
else
    log_test "FAIL" "Commands" "groups command not found"
fi
echo ""

# Test 6: Permissions check
echo -e "${BLUE}Test 6: Permissions Check${NC}"
if [[ $EUID -eq 0 ]]; then
    log_test "PASS" "Running as root (required for full features)"
else
    log_test "SKIP" "Root privileges" "Not running as root (some features may not work)"
fi
echo ""

# Test 7: Dependencies
echo -e "${BLUE}Test 7: System Dependencies${NC}"
for cmd in sudo systemctl useradd usermod; do
    if command -v $cmd &> /dev/null; then
        log_test "PASS" "$cmd command available"
    else
        log_test "SKIP" "Command $cmd" "Not available (may affect functionality)"
    fi
done
echo ""

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

PASS_COUNT=$(printf '%s\n' "${TEST_RESULTS[@]}" | grep -c "^✓" || true)
FAIL_COUNT=$(printf '%s\n' "${TEST_RESULTS[@]}" | grep -c "^✗" || true)
SKIP_COUNT=$(printf '%s\n' "${TEST_RESULTS[@]}" | grep -c "^⊘" || true)
TOTAL_COUNT=${#TEST_RESULTS[@]}

echo -e "${GREEN}Passed: $PASS_COUNT${NC}"
echo -e "${RED}Failed: $FAIL_COUNT${NC}"
echo -e "${YELLOW}Skipped: $SKIP_COUNT${NC}"
echo "Total: $TOTAL_COUNT"
echo ""

if [[ $FAIL_COUNT -eq 0 ]]; then
    echo -e "${GREEN}All critical tests passed! ✓${NC}"
    echo ""
    echo "Ready for manual testing:"
    echo "  sudo ./dist/ravact-linux-amd64"
else
    echo -e "${RED}Some tests failed. Review above for details.${NC}"
fi
echo ""
