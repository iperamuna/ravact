#!/bin/bash
#
# Ravact Release Script
# Automates version bumping, changelog generation, building, and GitHub release creation
#
# Usage: ./scripts/release.sh
# Requirements: Clean git working directory, Go 1.24+
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    Ravact Release Script                          ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: Must be run from project root${NC}"
    exit 1
fi

# Check if git is clean
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}Error: Working directory is not clean${NC}"
    echo "Please commit or stash your changes first"
    exit 1
fi

# Get current version from git tags
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT_VERSION=${CURRENT_VERSION#v} # Remove 'v' prefix

echo -e "${GREEN}Current version: ${CURRENT_VERSION}${NC}"
echo ""

# Parse semantic version
IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR="${VERSION_PARTS[0]}"
MINOR="${VERSION_PARTS[1]}"
PATCH="${VERSION_PARTS[2]}"

# Calculate next versions
NEXT_MAJOR="$((MAJOR + 1)).0.0"
NEXT_MINOR="${MAJOR}.$((MINOR + 1)).0"
NEXT_PATCH="${MAJOR}.${MINOR}.$((PATCH + 1))"

echo -e "${YELLOW}Suggested version bumps:${NC}"
echo "  1) Patch release: v${NEXT_PATCH} (bug fixes)"
echo "  2) Minor release: v${NEXT_MINOR} (new features, backwards compatible)"
echo "  3) Major release: v${NEXT_MAJOR} (breaking changes)"
echo "  4) Custom version"
echo ""

read -p "Select version type [1-4]: " VERSION_TYPE

case $VERSION_TYPE in
    1)
        NEW_VERSION=$NEXT_PATCH
        ;;
    2)
        NEW_VERSION=$NEXT_MINOR
        ;;
    3)
        NEW_VERSION=$NEXT_MAJOR
        ;;
    4)
        read -p "Enter custom version (x.y.z): " NEW_VERSION
        ;;
    *)
        echo -e "${RED}Invalid selection${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}New version will be: v${NEW_VERSION}${NC}"
echo ""
read -p "Continue? [y/N]: " CONFIRM

if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    echo "Release cancelled"
    exit 0
fi

# Generate release notes from commits
echo ""
echo -e "${BLUE}Generating release notes from recent commits...${NC}"

LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || git rev-list --max-parents=0 HEAD)
COMMITS=$(git log ${LAST_TAG}..HEAD --pretty=format:"- %s" | head -30)

# Check if CHANGELOG.md exists
if [ -f "docs/project/CHANGELOG.md" ]; then
    CHANGELOG_CONTENT=$(head -50 docs/project/CHANGELOG.md)
else
    CHANGELOG_CONTENT=""
fi

# Create temporary file for release notes
RELEASE_NOTES_FILE=$(mktemp)

cat > $RELEASE_NOTES_FILE << RELEASE_EOF
# Release v${NEW_VERSION}

## What's New

${COMMITS}

## Changelog

${CHANGELOG_CONTENT}

## Installation

\`\`\`bash
# Linux AMD64
curl -L https://github.com/iperamuna/ravact/releases/download/v${NEW_VERSION}/ravact-linux-amd64 -o ravact
chmod +x ravact
sudo mv ravact /usr/local/bin/

# Linux ARM64
curl -L https://github.com/iperamuna/ravact/releases/download/v${NEW_VERSION}/ravact-linux-arm64 -o ravact
chmod +x ravact
sudo mv ravact /usr/local/bin/

# Or use the install script:
curl -sSL https://raw.githubusercontent.com/iperamuna/ravact/main/scripts/install.sh | sudo bash
\`\`\`

## Requirements

- Ubuntu 24.04 LTS or compatible Linux distribution
- systemd
- sudo/root access

## Features

- MySQL and PostgreSQL management
- PHP-FPM pool configuration
- Supervisor process management with XML-RPC
- Nginx site management
- Redis configuration
- User management
- Quick system commands
- And much more!

See the [full documentation](https://github.com/iperamuna/ravact#readme) for details.

---

**Full Changelog**: https://github.com/iperamuna/ravact/blob/main/docs/project/CHANGELOG.md
RELEASE_EOF

# Allow user to edit release notes
echo ""
echo -e "${YELLOW}Opening editor for release notes...${NC}"
echo "Edit the release notes, save and exit when done."
echo ""
read -p "Press Enter to continue..."

${EDITOR:-nano} $RELEASE_NOTES_FILE

echo ""
echo -e "${BLUE}Building release binaries...${NC}"

# Create dist directory
mkdir -p dist

# Build for all platforms
echo "Building for Linux AMD64..."
rm -rf cmd/ravact/assets
cp -r assets cmd/ravact/
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=${NEW_VERSION} -s -w" -o dist/ravact-linux-amd64 ./cmd/ravact

echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=${NEW_VERSION} -s -w" -o dist/ravact-linux-arm64 ./cmd/ravact

echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=${NEW_VERSION} -s -w" -o dist/ravact-darwin-amd64 ./cmd/ravact

echo "Building for macOS ARM64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=${NEW_VERSION} -s -w" -o dist/ravact-darwin-arm64 ./cmd/ravact

rm -rf cmd/ravact/assets

echo ""
echo -e "${GREEN}✓ All binaries built successfully!${NC}"
ls -lh dist/
echo ""

# Create checksums
echo -e "${BLUE}Generating checksums...${NC}"
cd dist
sha256sum ravact-* > checksums.txt
cd ..

echo -e "${GREEN}✓ Checksums generated${NC}"
echo ""

# Create git tag
echo -e "${BLUE}Creating git tag v${NEW_VERSION}...${NC}"
git tag -a "v${NEW_VERSION}" -m "Release v${NEW_VERSION}"

echo -e "${GREEN}✓ Tag created${NC}"
echo ""

# Display summary
echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                        Release Summary                            ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}Version:${NC} v${NEW_VERSION}"
echo -e "${GREEN}Binaries:${NC}"
echo "  • ravact-linux-amd64 ($(du -h dist/ravact-linux-amd64 | cut -f1))"
echo "  • ravact-linux-arm64 ($(du -h dist/ravact-linux-arm64 | cut -f1))"
echo "  • ravact-darwin-amd64 ($(du -h dist/ravact-darwin-amd64 | cut -f1))"
echo "  • ravact-darwin-arm64 ($(du -h dist/ravact-darwin-arm64 | cut -f1))"
echo "  • checksums.txt"
echo ""
echo -e "${GREEN}Release Notes:${NC} $RELEASE_NOTES_FILE"
echo ""

# Next steps
echo -e "${YELLOW}Next steps:${NC}"
echo ""
echo "1. Push tag to GitHub:"
echo -e "   ${BLUE}git push origin v${NEW_VERSION}${NC}"
echo ""
echo "2. Create GitHub release:"
echo "   a. Go to: https://github.com/iperamuna/ravact/releases/new"
echo "   b. Tag: v${NEW_VERSION}"
echo "   c. Title: Ravact v${NEW_VERSION}"
echo "   d. Copy release notes from: $RELEASE_NOTES_FILE"
echo "   e. Upload binaries from: dist/"
echo "   f. Publish release"
echo ""
echo "Or use GitHub CLI:"
echo -e "   ${BLUE}gh release create v${NEW_VERSION} dist/* --notes-file $RELEASE_NOTES_FILE --title \"Ravact v${NEW_VERSION}\"${NC}"
echo ""

# Ask if user wants to push
read -p "Push tag to GitHub now? [y/N]: " PUSH_TAG

if [ "$PUSH_TAG" = "y" ] || [ "$PUSH_TAG" = "Y" ]; then
    git push origin "v${NEW_VERSION}"
    echo -e "${GREEN}✓ Tag pushed to GitHub${NC}"
    echo ""
    echo "You can now create the release on GitHub or use:"
    echo "  gh release create v${NEW_VERSION} dist/* --notes-file $RELEASE_NOTES_FILE"
fi

echo ""
echo -e "${GREEN}✅ Release preparation complete!${NC}"
echo ""
