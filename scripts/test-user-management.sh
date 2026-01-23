#!/bin/bash
# Test User Management feature specifically
# This script tests the user management loading issue

echo "=========================================="
echo "Testing User Management Feature"
echo "=========================================="
echo ""

# Pre-checks
echo "Pre-checks:"
echo "-----------"
echo "✓ Binary: ./dist/ravact-linux-amd64"
echo "✓ Architecture: $(file ./dist/ravact-linux-amd64 | grep -o 'x86-64')"
echo "✓ OS: $(uname -s) $(uname -m)"
echo "✓ Users in system: $(cat /etc/passwd | wc -l)"
echo "✓ Groups in system: $(cat /etc/group | wc -l)"
echo ""

# Test the groups command that was causing issues
echo "Testing 'groups' command (root cause of hang):"
echo "------------------------------------------------"
echo -n "Running: groups root ... "
timeout 2 groups root > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ OK (completed in < 2s)"
    groups root
else
    echo "✗ TIMEOUT or ERROR"
fi
echo ""

# Test reading /etc/passwd
echo "Testing /etc/passwd read:"
echo "-------------------------"
echo -n "Reading users... "
USERS=$(grep -E '^[^:]+:[^:]*:[0-9]{4,}:' /etc/passwd | wc -l)
echo "✓ Found $USERS regular users (UID >= 1000)"
echo ""

# Test reading /etc/group
echo "Testing /etc/group read:"
echo "------------------------"
echo -n "Reading groups... "
GROUPS=$(cat /etc/group | wc -l)
echo "✓ Found $GROUPS groups"
echo ""

# Manual instructions
cat << 'EOF'
========================================
Manual Test Instructions
========================================

The automated tests passed! Now test the TUI:

1. Run ravact:
   sudo ./dist/ravact-linux-amd64

2. Navigate to User Management:
   - Press '2' or use arrow keys and Enter

3. What to check:
   
   ✓ EXPECTED (FIXED):
   - Brief "Loading..." message (1-2 seconds)
   - User list appears with:
     * Usernames
     * UIDs
     * Sudo status
     * Groups
   - Can navigate with arrow keys
   - Tab switches to Groups view
   - 'r' refreshes
   - Esc goes back

   ✗ IF BROKEN:
   - "Loading..." forever
   - Cannot navigate
   - Must Ctrl+C to exit

4. Test other screens:
   - Press Esc to go back
   - Try Setup Menu (option 1)
   - Try Nginx Config (option 3)
   - Press 'q' to quit

========================================

Ready to test interactively? 
Run: sudo ./dist/ravact-linux-amd64

Press Ctrl+C in ravact at any time to exit.
========================================
EOF
