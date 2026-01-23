# ARM64 VM Testing Instructions

## ✅ Setup Complete

- VM: **ravact-dev** (Running)
- IP: **192.168.64.10**
- Architecture: **ARM64** (native speed)
- Binary: **Deployed and ready**

---

## 🧪 How to Test

### **Option 1: Direct SSH Test** (Recommended)

```bash
# SSH into VM
ssh ravact-dev

# Go to project folder
cd ravact-go

# Run ravact with sudo (required for user management)
sudo ./ravact
```

### **Option 2: Quick Run from Mac**

```bash
# Run directly from your Mac
export PATH="/usr/local/bin:$PATH"
multipass exec ravact-dev -- sudo /home/ubuntu/ravact-go/ravact
```

**Note:** This might not have proper TTY, so Option 1 is better for testing!

---

## 📋 Test Plan

### **Priority: User Management** (The reported issue)

1. Run ravact: `sudo ./ravact`
2. Navigate to "User Management" (press `2`)
3. **Check what happens:**

   ✅ **EXPECTED (Fixed):**
   - Shows "Loading users and groups..." for 1-2 seconds
   - Then displays user list with usernames, UIDs, sudo status
   - Can navigate with arrow keys
   - Press Tab to switch to Groups view
   - Press 'r' to refresh
   - Press Esc to go back

   ❌ **IF BROKEN:**
   - "Loading..." shows forever
   - Cannot navigate
   - Need Ctrl+C to exit

---

## 🎯 Step-by-Step Testing

### **Test 1: Main Menu**
```bash
sudo ./ravact
```
- [ ] Main menu appears
- [ ] All options visible
- [ ] Can navigate with arrows
- [ ] Press 'q' to quit works

### **Test 2: User Management** ⭐ (CRITICAL)
```bash
# From main menu
Press: 2
```
- [ ] "Loading..." message shows (brief)
- [ ] User list appears
- [ ] Shows root, ubuntu users
- [ ] Arrow keys work
- [ ] Tab switches to Groups
- [ ] 'r' refreshes
- [ ] Esc goes back

### **Test 3: Setup Menu**
```bash
# From main menu
Press: 1
```
- [ ] Setup scripts listed
- [ ] Can navigate
- [ ] Esc goes back

### **Test 4: Nginx Config**
```bash
# From main menu  
Press: 3
```
- [ ] Shows nginx status
- [ ] Esc goes back

### **Test 5: Quick Commands**
```bash
# From main menu
Press: 4
```
- [ ] Commands listed
- [ ] Esc goes back

### **Test 6: Installed Apps**
```bash
# From main menu
Press: 5
```
- [ ] Apps detected
- [ ] Esc goes back

---

## 📊 What to Report

For each test, note:
- ✅ **PASS** - Works as expected
- ❌ **FAIL** - Broken (describe issue)
- ⚠️ **PARTIAL** - Works but has issues

### **Example Report:**

```
TEST: User Management
STATUS: ✅ PASS
LOADING TIME: 1.5 seconds
USERS SHOWN: 3 (root, ubuntu, _apt)
NAVIGATION: Works perfectly
TAB SWITCH: Works
REFRESH: Works  
ISSUES: None
```

---

## 🔧 Quick Commands Reference

```bash
# Connect to VM
ssh ravact-dev

# Navigate to project
cd ravact-go

# Run ravact
sudo ./ravact

# In ravact:
# - Arrow keys: Navigate
# - Enter: Select
# - Tab: Switch views (in User Management)
# - r: Refresh (in User Management)
# - Esc: Go back
# - q: Quit (from main menu)
# - Ctrl+C: Force quit
```

---

## 🚀 Ready to Test!

**Just run:**

```bash
ssh ravact-dev
cd ravact-go
sudo ./ravact
```

**Then press `2` for User Management and see if it loads!**

---

## 💡 Tips

1. **User Management is the priority** - This was the reported bug
2. **Loading should be fast** - 1-3 seconds max
3. **Navigation should be smooth** - Arrow keys work everywhere
4. **Can always exit** - Esc or Ctrl+C
5. **Test on real Linux** - VM has proper users/groups/sudo

---

## 📝 After Testing

Let me know the results! Especially:

1. **Does User Management load properly?** (Yes/No)
2. **How long did it take?** (seconds)
3. **Can you navigate users?** (Yes/No)
4. **Any errors or issues?** (describe)

This will tell us if the fix worked! 🎉
