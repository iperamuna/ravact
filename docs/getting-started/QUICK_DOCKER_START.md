# Quick Docker Start

## ✅ Once Docker is Installed

### **Step 1: Start Docker Desktop**
```bash
open -a Docker
# Wait for whale icon in menu bar to be stable
```

### **Step 2: Test Docker Works**
```bash
docker --version
docker ps
```

### **Step 3: Choose Your Workflow**

#### **Option A: Quick Test** (One command)
```bash
cd ravact-go/scripts
./test-docker-amd64.sh
```

#### **Option B: Development Container** (Persistent)
```bash
cd ravact-go/scripts
./docker-amd64-dev.sh
# Container stays open, you can reconnect anytime
```

#### **Option C: Automated Build & Test** (Easiest!)
```bash
cd ravact-go/scripts
./docker-build-and-test.sh
# Builds on Mac, tests in container automatically!
```

---

## 🔄 Daily Workflow

```bash
# Morning: Start container
./scripts/docker-amd64-dev.sh

# Terminal 1 (Mac): Edit & Build
vim internal/ui/screens/main_menu.go
make build-linux

# Terminal 2 (Container): Test
sudo ./dist/ravact-linux-amd64

# Repeat: Edit on Mac → Build → Test in container
# No sync needed! Files update automatically!
```

---

## 🛠️ Container Management

```bash
./scripts/docker-manager.sh start    # Start
./scripts/docker-manager.sh shell    # Connect
./scripts/docker-manager.sh stop     # Stop
./scripts/docker-manager.sh status   # Check status
```

---

## ✨ Key Point: No Syncing!

**Docker uses volume mounts:**
- Your Mac folder is mounted live into container
- Changes on Mac → Instantly in container
- No copy, no sync, just works!

---

**Read full guide:** `DOCKER_WORKFLOW.md`

**When Docker finishes installing, come back here!** 🚀
