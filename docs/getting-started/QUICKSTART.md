# Ravact Quick Start Guide

[← Back to Documentation](../README.md)

Get Ravact up and running in 5 minutes!

## Prerequisites

- macOS (for development) or Linux (for production use)
- Go 1.21+ (for building from source)
- Docker Desktop (optional, for testing)

## Installation

### Option 1: Build from Source

```bash
# Clone or navigate to the project
cd ravact-go

# Build for your current platform
make build

# Run it
./ravact
```

### Option 2: Use Pre-built Binary

```bash
# Build all platforms
make build-all

# On Linux
./dist/ravact-linux-amd64

# On macOS (for testing only)
./dist/ravact-darwin-arm64  # Apple Silicon
./dist/ravact-darwin-amd64  # Intel
```

## First Run

1. **Launch the application**:
   ```bash
   ./ravact
   ```

2. **You'll see the splash screen** - press any key to continue

3. **Main Menu appears** with three options:
   - Setup: Install software
   - Configuration: Manage configs (coming soon)
   - Quick Commands: Run common tasks

## Quick Tasks

### Install Nginx

1. Select **Setup** from main menu
2. Navigate to **Nginx Web Server**
3. Press Enter to install
4. Wait for installation to complete

**Note**: This requires root access. Use `sudo ./ravact` on Linux.

### Run Quick Commands

1. Select **Quick Commands** from main menu
2. Choose a command (e.g., "View Nginx Status")
3. Press Enter to execute

### Navigate the Interface

- **↑/↓** or **j/k**: Move up/down
- **Enter/Space**: Select item
- **Esc/Backspace**: Go back
- **q**: Quit application
- **Ctrl+C**: Force quit

## Testing

### Run Unit Tests
```bash
make test
```

### Run Integration Tests
```bash
go test -tags=integration ./tests/... -v
```

### Test in Docker (Ubuntu 24.04)
```bash
make docker-test
```

### Interactive Docker Shell
```bash
make docker-shell
# Inside container:
go test ./...
./dist/ravact-linux-amd64 --version
```

## Development

### Build for All Platforms
```bash
make build-all
```

### Run Tests with Coverage
```bash
make test-coverage
# Open coverage.html in browser
```

### Clean Build Artifacts
```bash
make clean
```

## Adding Content

### Add a Setup Script

1. Create `assets/scripts/myservice.sh`:
   ```bash
   #!/bin/bash
   echo "Installing MyService..."
   # Your installation commands
   ```

2. Make executable:
   ```bash
   chmod +x assets/scripts/myservice.sh
   ```

3. It automatically appears in Setup menu!

### Add a Configuration Template

1. Create `assets/configs/myservice.json`:
   ```json
   {
     "id": "myservice-config",
     "service_id": "myservice",
     "name": "MyService Configuration",
     "fields": [
       {
         "key": "port",
         "label": "Port",
         "type": "int",
         "required": true,
         "default": 8080
       }
     ]
   }
   ```

## Troubleshooting

### "Permission denied" errors
```bash
# Run with sudo on Linux
sudo ./ravact
```

### Can't see scripts in Setup menu
```bash
# Make sure scripts are executable
chmod +x assets/scripts/*.sh
```

### Tests fail on macOS
```bash
# Some tests are Linux-specific, use Docker
make docker-test
```

### Docker issues
```bash
# Make sure Docker Desktop is running
docker info

# Rebuild image
docker build --no-cache -t ravact-test -f Dockerfile.test .
```

## Next Steps

- Read [README.md](README.md) for detailed information
- Check [DEVELOPMENT.md](DEVELOPMENT.md) for contributing
- Review [PROJECT_STATUS.md](PROJECT_STATUS.md) for current state
- Explore the code in `internal/` directory

## Getting Help

- Check the documentation files
- Review test files for usage examples
- Examine existing scripts in `assets/scripts/`

## Tips

1. **Use Docker for testing Linux features** while developing on macOS
2. **All setup scripts must start with** `#!/bin/bash` or `#!/bin/sh`
3. **Test scripts thoroughly** before running on production servers
4. **Create backups** - config changes create `.backup` files automatically
5. **Review script output** to ensure installations succeed

---

**Ready to go? Start with:** `./ravact`

**Happy server managing! 🚀**
