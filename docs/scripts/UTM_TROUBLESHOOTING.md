# UTM Synchronous Exception Fix for Ubuntu ARM64

## Problem
When booting Ubuntu 24.04 ARM64 ISO in UTM on M1 Mac, you get:
```
Synchronous Exception at [address]
```

## Solutions (Try in Order)

### Solution 1: Use Different Ubuntu Image (RECOMMENDED) ✅

**Problem:** The live-server ISO sometimes has boot issues with UTM.

**Fix:** Use Ubuntu **Desktop** ARM64 or **Cloud Image** instead:

#### Option A: Ubuntu Desktop ARM64 (Easiest)
```bash
cd ~/Downloads
curl -LO https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04-desktop-arm64.iso
```

**In UTM:**
- Use this ISO instead
- Everything else same
- Desktop has better compatibility

#### Option B: Ubuntu Cloud Image (Lighter, Faster)
```bash
cd ~/Downloads

# Download cloud image
curl -LO https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-arm64.img

# This is pre-configured, no installation needed!
```

**In UTM:**
1. Create New VM → Virtualize
2. Linux → Skip ISO (click Continue)
3. Set RAM/CPU as normal
4. **IMPORTANT:** Delete the default disk
5. Import the downloaded .img file as disk
6. Boot and configure cloud-init

---

### Solution 2: Fix UTM VM Settings (If Using Server ISO)

**Change these UTM settings:**

1. **Select your VM → Edit → System**
   - Architecture: `ARM64 (aarch64)` ✅
   - System: Change from `virt-4.2` to `virt-6.2` or `virt-7.0`
   - Boot: UEFI

2. **Display Tab**
   - Change from `virtio-gpu-gl-pci` to `virtio-ramfb` or `virtio-gpu-pci`
   - Disable "Retina Mode" if enabled

3. **Save and try booting again**

---

### Solution 3: Use Older Ubuntu Version

If Ubuntu 24.04 keeps failing:

```bash
cd ~/Downloads
curl -LO https://cdimage.ubuntu.com/releases/22.04/release/ubuntu-22.04.3-live-server-arm64.iso
```

Ubuntu 22.04 LTS is more stable with UTM and works perfectly for ravact.

---

### Solution 4: Use Multipass (Alternative to UTM) ⚡

**Multipass** is even easier than UTM for Ubuntu VMs on M1!

```bash
# Install
brew install multipass

# Create VM (one command!)
multipass launch -n ravact-dev -c 4 -m 4G -d 20G 24.04

# Get IP
multipass info ravact-dev

# SSH into VM
multipass shell ravact-dev

# Or use SSH
multipass exec ravact-dev -- ip addr show
ssh ubuntu@<VM-IP>
```

**Advantages:**
- ✅ No ISO download needed
- ✅ No boot issues
- ✅ Faster setup
- ✅ Built-in file sharing
- ✅ Native ARM64 performance

---

### Solution 5: Try Alternative Boot Parameters

**In UTM, before booting:**

1. When you see GRUB menu (Try or Install Ubuntu)
2. Press `e` to edit boot parameters
3. Find line starting with `linux`
4. Add these parameters at the end:
   ```
   nomodeset console=ttyAMA0
   ```
5. Press `Ctrl+X` or `F10` to boot

---

## Recommended Approach for You

### **BEST: Use Multipass** 🥇

It's simpler and more reliable:

```bash
# Install Multipass
brew install multipass

# Create VM
multipass launch -n ravact-dev -c 4 -m 4G -d 20G 24.04

# Get VM details
multipass info ravact-dev

# This shows IP address - use it with our scripts!
```

Then use our `setup-vm-only.sh` script:
```bash
# Get VM IP from multipass info
VM_IP=$(multipass info ravact-dev | grep IPv4 | awk '{print $2}')

# Copy setup script
multipass transfer scripts/setup-vm-only.sh ravact-dev:/home/ubuntu/

# Run setup
multipass exec ravact-dev -- bash /home/ubuntu/setup-vm-only.sh

# Or SSH and run
ssh ubuntu@$VM_IP
bash setup-vm-only.sh
```

---

### **GOOD: Use Ubuntu Desktop ISO in UTM** 🥈

```bash
cd ~/Downloads
curl -LO https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04-desktop-arm64.iso
```

Use this in UTM - much better compatibility than server ISO.

---

### **OK: Fix UTM Settings** 🥉

Change System type to `virt-6.2` or newer and Display to `virtio-ramfb`.

---

## Updated Setup Script for Multipass

I can create a `setup-multipass.sh` script that:
1. Installs Multipass
2. Creates VM automatically
3. Runs setup
4. Deploys ravact

Would you like that?

---

## Which Should You Use?

| Method | Difficulty | Speed | Reliability |
|--------|-----------|-------|-------------|
| **Multipass** | ⭐ Easy | ⚡⚡⚡ Fast | ✅✅✅ Very High |
| **UTM Desktop ISO** | ⭐⭐ Medium | ⚡⚡ Medium | ✅✅ High |
| **UTM Server ISO** | ⭐⭐⭐ Hard | ⚡ Slow | ✅ Medium |

**My recommendation:** Use **Multipass** - it's designed exactly for this use case!

---

## Need Help?

Let me know which approach you want to try:
1. Switch to Multipass (I can create a script)
2. Try Ubuntu Desktop ISO in UTM
3. Fix UTM settings for server ISO
4. Use Ubuntu 22.04 instead

I can guide you through any of these!
