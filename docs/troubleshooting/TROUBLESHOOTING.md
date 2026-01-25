# Troubleshooting Guide

[← Back to Documentation](../README.md)

## Common Issues and Solutions

### TUI shows "Loading..." after navigation

**Symptom**: After pressing a key on the splash screen, the app shows "Loading..." instead of the main menu.

**Cause**: Child screens check for `width > 0` before rendering. They need to receive a `WindowSizeMsg` to know the terminal dimensions.

**Solution**: ✅ Fixed in version 0.1.0 - Navigation now sends WindowSizeMsg to newly displayed screens.

**Technical Details**: 
- The main Update() handler captures WindowSizeMsg and stores width/height
- When NavigateMsg is received, it sends a new WindowSizeMsg to the newly active screen
- This ensures all screens have proper dimensions before rendering

### Screen appears garbled or improperly sized

**Solution**: Try resizing your terminal window. The TUI should automatically adjust.

### Navigation not working

**Check**:
- Arrow keys (↑/↓) or vim keys (j/k) for navigation
- Enter or Space to select
- Esc or Backspace to go back
- q to quit

### "Permission denied" when running setup scripts

**Solution**: Run with sudo on Linux:
```bash
sudo ./ravact
```

Setup scripts often require root privileges to install system packages.

### Scripts not appearing in Setup menu

**Check**:
1. Scripts exist in `assets/scripts/` directory
2. Scripts have `.sh` extension
3. Scripts are executable: `chmod +x assets/scripts/*.sh`
4. Scripts start with `#!/bin/bash` or `#!/bin/sh`

### Configuration templates not loading

**Check**:
1. Templates exist in `assets/configs/` directory
2. Templates have `.json` extension
3. JSON is valid (use `jq` or JSON validator)
4. Required fields are present (id, service_id, name, fields)

### Tests fail on macOS

**Expected**: Some tests are Linux-specific and will be skipped on macOS.

**Solution**: Use Docker for full testing:
```bash
make docker-test
```

### Application crashes on startup

**Debug**:
```bash
# Run with error output
./ravact 2> error.log

# Check Go version
go version  # Should be 1.21+

# Rebuild from scratch
make clean
make build
```

### Slow startup time

**Normal**: First run may take 50-100ms for system detection.

**If slower**:
- Check disk I/O
- Verify no hanging processes
- Check system resources

### Cannot build for Linux

**Check**:
```bash
# Ensure cross-compilation tools are available
GOOS=linux GOARCH=amd64 go build ./cmd/ravact

# Try explicit build
make build-linux
```

### Docker tests fail

**Check**:
```bash
# Is Docker running?
docker info

# Try rebuilding image
docker build --no-cache -t ravact-test -f Dockerfile.test .

# Check Docker Desktop is running (on Mac/Windows)
```

### Binary too large

**Normal**: ~3MB per binary is expected with TUI libraries.

**To reduce**:
```bash
# Already using -s -w flags for stripping
# Build flags in Makefile optimize for size
make build-linux  # Produces minimal binary
```

## Debug Mode

To enable verbose logging:

```go
// In your code, add:
import "log"
import "os"

f, _ := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
defer f.Close()
log.SetOutput(f)
log.Println("Debug message here")
```

Or set environment variable (if implemented):
```bash
export RAVACT_DEBUG=1
./ravact
```

## Getting Help

1. Check the documentation:
   - `README.md` - Overview
   - `QUICKSTART.md` - Quick start
   - `DEVELOPMENT.md` - Development guide

2. Check test files for usage examples

3. Review the code:
   - `internal/` - Core logic
   - `cmd/ravact/main.go` - Entry point

## Reporting Issues

When reporting issues, include:

1. **Environment**:
   - OS and version
   - Go version (`go version`)
   - Terminal emulator
   - Running as root/sudo?

2. **Steps to reproduce**:
   - Exact commands run
   - What you expected
   - What actually happened

3. **Logs/Output**:
   - Error messages
   - Screenshots if relevant
   - Debug logs

4. **Version**:
   ```bash
   ./ravact --version
   ```

## Performance Tips

1. **System Detection**: Cached after first run per session
2. **Script Execution**: Can be slow for large packages - this is normal
3. **TUI Rendering**: Should be smooth - if laggy, check terminal performance

## Known Limitations

1. **Linux-specific features**: Many features only work on Linux (by design)
2. **Root required**: Installation and some commands need sudo
3. **Terminal size**: Very small terminals may not display correctly
4. **No mouse support yet**: Keyboard navigation only
5. **Single server**: No remote or multi-server support yet

## Quick Fixes

### Reset to clean state
```bash
make clean
go mod tidy
make build
```

### Test in isolated environment
```bash
make docker-test
```

### Verify installation
```bash
./ravact --version
go test ./...
```

---

**Still having issues?** Check the code or create an issue with details above.
