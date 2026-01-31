#!/bin/bash

# Check if version argument is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <new_version>"
    echo "Example: $0 0.4.2"
    exit 1
fi

NEW_VERSION=$1
# Remove 'v' prefix if present (e.g. v0.4.2 -> 0.4.2)
NEW_VERSION=${NEW_VERSION#v}

echo "Bumping version to ${NEW_VERSION}..."

# 1. Update Makefile
if [ -f "Makefile" ]; then
    # Check for macOS (darwin) which requires empty string for sed -i
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/^VERSION?=.*/VERSION?=${NEW_VERSION}/" Makefile
    else
        sed -i "s/^VERSION?=.*/VERSION?=${NEW_VERSION}/" Makefile
    fi
    echo "✓ Updated Makefile"
else
    echo "✗ Makefile not found"
fi

# 2. Update cmd/ravact/main.go
if [ -f "cmd/ravact/main.go" ]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/var Version = \".*\"/var Version = \"${NEW_VERSION}\"/" cmd/ravact/main.go
    else
        sed -i "s/var Version = \".*\"/var Version = \"${NEW_VERSION}\"/" cmd/ravact/main.go
    fi
    echo "✓ Updated cmd/ravact/main.go"
fi

# 3. Update README.md
if [ -f "README.md" ]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/\*\*Version\*\*: .*/\*\*Version\*\*: ${NEW_VERSION}/" README.md
    else
        sed -i "s/\*\*Version\*\*: .*/\*\*Version\*\*: ${NEW_VERSION}/" README.md
    fi
    echo "✓ Updated README.md"
fi

echo ""
echo "Version updated to ${NEW_VERSION} successfully! 🚀"
