# Ravact Release Guide

This guide explains how to use the release scripts to create and distribute new versions of Ravact.

## Overview

Ravact has two main scripts for release management:

1. **release.sh** - Creates releases with automatic version bumping and multi-platform builds
2. **install.sh** - Installs Ravact from GitHub releases

---

## Release Process (release.sh)

### Prerequisites

- Git repository with commits ready to release
- Working directory must be clean
- GitHub CLI (optional, for automated release creation)

### Usage

```bash
./scripts/release.sh
```

### What release.sh Does

1. **Version Management**
   - Shows current version from git tags
   - Suggests next versions based on semantic versioning:
     - Patch: Bug fixes (1.0.1 → 1.0.2)
     - Minor: New features, backwards compatible (1.0.1 → 1.1.0)
     - Major: Breaking changes (1.0.1 → 2.0.0)
   - Allows custom version input

2. **Release Notes Generation**
   - Extracts recent commits since last tag
   - Loads changelog from `docs/project/CHANGELOG.md`
   - Opens editor for manual customization
   - Includes installation instructions

3. **Multi-Platform Builds**
   - Builds for Linux AMD64
   - Builds for Linux ARM64
   - Builds for macOS AMD64
   - Builds for macOS ARM64 (Apple Silicon)
   - Generates SHA256 checksums

4. **Git Management**
   - Creates annotated git tag
   - Optionally pushes tag to GitHub

### Step-by-Step Example

```bash
# 1. Run the release script
./scripts/release.sh

# 2. Choose version type (or enter custom)
# Current version: 1.0.0
# Select version type [1-4]: 2  # Choose Minor release
# New version will be: v1.1.0

# 3. Confirm release
# Continue? [y/N]: y

# 4. Edit release notes
# (Opens in editor - customize as needed)

# 5. Binaries are built automatically
# ✓ All binaries built successfully!
# ✓ Checksums generated

# 6. Choose to push tag
# Push tag to GitHub now? [y/N]: y
# ✓ Tag pushed to GitHub

# 7. Create release on GitHub
# Use GitHub CLI or web interface to create the release
```

### Release Notes Template

The release notes automatically include:
- Recent commits (up to 30)
- Changelog content
- Installation instructions for all platforms
- Feature list
- Requirements

Users can customize before publishing.

---

## Installation (install.sh)

### Usage

#### Option 1: Direct Script Execution

```bash
curl -sSL https://raw.githubusercontent.com/iperamuna/ravact/main/scripts/install.sh | sudo bash
```

#### Option 1b: Local Binary Install (Offline Testing)

```bash
sudo bash ./install.sh --local /path/to/ravact-linux-arm64
```

#### Option 2: Manual Download and Execute

```bash
# Download the script
curl -L https://raw.githubusercontent.com/iperamuna/ravact/main/scripts/install.sh -o install.sh

# Make it executable
chmod +x install.sh

# Run with sudo
sudo ./install.sh
```

### What install.sh Does

1. **System Detection**
   - Detects OS (Linux, macOS)
   - Detects architecture (AMD64, ARM64)
   - Determines correct binary to download

2. **Version Selection**
   - Fetches latest version from GitHub releases
   - Allows user to specify custom version
   - Default is latest version

3. **Download and Verification**
   - Downloads binary from GitHub releases
   - Verifies file integrity
   - Makes binary executable

4. **Installation**
   - Backs up existing installation if present
   - Copies binary to `/usr/local/bin/ravact`
   - Sets proper permissions

5. **Verification**
   - Confirms successful installation
   - Shows version information
   - Provides usage instructions

### Supported Platforms

| OS    | Architecture | Binary Name          |
|-------|--------------|----------------------|
| Linux | AMD64        | ravact-linux-amd64   |
| Linux | ARM64        | ravact-linux-arm64   |
| macOS | AMD64        | ravact-darwin-amd64  |
| macOS | ARM64        | ravact-darwin-arm64  |

### Post-Installation

After installation, run Ravact with:

```bash
sudo ravact
```

To update to a newer version, simply run the install script again.

---

## Release Checklist

Before releasing a new version:

- [ ] All features implemented and tested
- [ ] Documentation updated
- [ ] CHANGELOG.md updated with new features
- [ ] All tests passing
- [ ] Code review completed
- [ ] Working tree is clean
- [ ] Latest changes committed

## Release Steps

1. Ensure working tree is clean:
   ```bash
   git status
   ```

2. Run release script:
   ```bash
   ./scripts/release.sh
   ```

3. Follow prompts for version selection and release notes

4. Verify binaries in `dist/` directory:
   ```bash
   ls -lh dist/
   ```

5. Push tag to GitHub:
   ```bash
   git push origin v1.1.0
   ```

6. Create GitHub release:

   **Option A: Using GitHub CLI**
   ```bash
   gh release create v1.1.0 dist/* --notes-file release_notes.txt
   ```

   **Option B: Using GitHub Web Interface**
   - Go to: https://github.com/iperamuna/ravact/releases/new
   - Select tag: v1.1.0
   - Add title: Ravact v1.1.0
   - Add description from release notes
   - Upload binaries from `dist/`
   - Publish

---

## Troubleshooting

### release.sh Issues

**Error: "Working directory is not clean"**
- Commit or stash all changes
- Run `git status` to see what needs committing

**Error: "Must be run from project root"**
- Navigate to project root directory
- Ensure `go.mod` exists in current directory

**Build fails**
- Ensure Go is installed: `go version`
- Check go.mod is valid: `go mod tidy`

### install.sh Issues

**Error: "Unsupported architecture"**
- Check your system: `uname -m`
- Currently supported: x86_64, aarch64, arm64

**Error: "Unsupported OS"**
- Check your OS: `uname -s`
- Currently supported: Linux, Darwin (macOS)

**Error: "Failed to download binary"**
- Check internet connection: `curl -I https://github.com`
- Verify release exists: `curl https://api.github.com/repos/iperamuna/ravact/releases/latest`
- Verify version is published on GitHub

**Error: "Installation verification failed"**
- Check permissions: `ls -l /usr/local/bin/ravact`
- Try manual installation steps

---

## Version History

Ravact follows semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, no new features

Each release is tagged in git as `v{version}`.

View all releases:
```bash
git tag -l
```

View specific release:
```bash
git show v1.1.0
```

---

## Automating with GitHub Actions

Future enhancement: Automate release creation with GitHub Actions workflow that:
- Automatically builds on tag push
- Creates GitHub release automatically
- Uploads all binaries
- Generates release notes from commits

See `.github/workflows/release.yml` for current setup.

---

**Last Updated:** January 2026
