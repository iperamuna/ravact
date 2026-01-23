#!/bin/bash
# Manual Testing Guide for Ravact
# Step-by-step testing instructions

cat << 'EOF'
========================================
Ravact Manual Testing Guide
========================================

SETUP:
------
1. Connect to container: docker exec -it ravact-amd64-dev bash
2. Go to workspace: cd /workspace
3. Run as root: sudo ./dist/ravact-linux-amd64

TESTS TO PERFORM:
-----------------

TEST 1: Main Menu
-----------------
✓ Check: Main menu appears
✓ Check: All menu options visible
✓ Check: Navigation works (up/down arrows)
✓ Check: Press 'q' to quit works

TEST 2: User Management (ISSUE REPORTED)
-----------------------------------------
Steps:
1. Press '2' or navigate to "User Management"
2. Wait 2-3 seconds for loading
3. Check what happens:
   
   Expected (FIXED):
   ✓ Shows "Loading users and groups..." briefly
   ✓ Then shows user list OR shows macOS warning
   ✓ Can navigate back with Esc
   
   If BROKEN:
   ✗ Shows "Loading..." forever
   ✗ Cannot exit
   ✗ Need to Ctrl+C

4. If users load:
   - Check: Can see usernames
   - Check: Can see UIDs
   - Check: Can see sudo status
   - Press Tab: Switch to Groups view
   - Press 'r': Refresh data
   - Press Esc: Go back to main menu

TEST 3: Setup Menu
------------------
1. Navigate to "Setup" (option 1)
2. Check: Setup scripts listed
3. Check: Can navigate with arrows
4. Press Esc: Go back

TEST 4: Nginx Configuration
---------------------------
1. Navigate to "Nginx Configuration"
2. Check if nginx installed message appears
3. Press Esc: Go back

TEST 5: Quick Commands
----------------------
1. Navigate to "Quick Commands"
2. Check: Commands listed
3. Press Esc: Go back

TEST 6: Installed Apps
----------------------
1. Navigate to "Installed Apps"
2. Check: Detection works
3. Press Esc: Go back

REPORTING RESULTS:
------------------
For each test, note:
- ✓ PASS - Works as expected
- ✗ FAIL - Broken, describe issue
- ⊘ SKIP - Cannot test (missing deps, etc.)

========================================
Press Enter to continue...
EOF

read

# Now run basic checks
echo ""
echo "Pre-flight checks:"
echo ""

# Check binary
if [[ -f "./dist/ravact-linux-amd64" ]]; then
    echo "✓ Binary exists"
else
    echo "✗ Binary not found"
    exit 1
fi

# Check version
echo ""
echo "Binary version:"
./dist/ravact-linux-amd64 --version 2>&1 || echo "No version flag"

# Check architecture
echo ""
echo "Binary architecture:"
file ./dist/ravact-linux-amd64 | grep -o "x86-64\|aarch64"

# Check system
echo ""
echo "System info:"
echo "  OS: $(cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d '"')"
echo "  Arch: $(uname -m)"
echo "  Users in /etc/passwd: $(cat /etc/passwd | wc -l)"
echo "  Root: $(if [[ $EUID -eq 0 ]]; then echo 'Yes'; else echo 'No (use sudo)'; fi)"

echo ""
echo "========================================"
echo "Ready to test!"
echo "========================================"
echo ""
echo "Run: sudo ./dist/ravact-linux-amd64"
echo ""
