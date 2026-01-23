# Quick Test - User Management

## 🎯 One Command Test

```bash
docker exec -it ravact-amd64-dev bash -c 'cd /workspace && sudo ./dist/ravact-linux-amd64'
```

## 📋 What to Do

1. **Wait for main menu** (should appear immediately)
2. **Press `2`** for User Management
3. **Wait 1-2 seconds**
4. **Check result:**

   ✅ **PASS** - User list appears with usernames, UIDs, groups
   
   ✗ **FAIL** - Shows "Loading..." forever, cannot exit

5. **If PASS:**
   - Try arrow keys to navigate
   - Press Tab to switch to Groups
   - Press 'r' to refresh
   - Press Esc to go back
   - Press 'q' to quit

6. **If FAIL:**
   - Press Ctrl+C to exit
   - Report the issue

## 📊 Report Result

Just tell me:
- ✅ "User Management works!" 
- ✗ "Still broken - hangs on loading"

That's it! 🚀
