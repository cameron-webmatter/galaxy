#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

if [ -z "$1" ]; then
    echo -e "${RED}Error: Version number required${NC}"
    echo "Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh 0.3.0"
    exit 1
fi

NEW_VERSION=$1

if ! [[ $NEW_VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format${NC}"
    echo "Version must be in format: X.Y.Z (e.g., 0.3.0)"
    exit 1
fi

CURRENT_VERSION=$(cat VERSION)

echo -e "${YELLOW}Current version: ${CURRENT_VERSION}${NC}"
echo -e "${YELLOW}New version:     ${NEW_VERSION}${NC}"
echo ""

if git diff-index --quiet HEAD --; then
    echo -e "${GREEN}✓ Working directory clean${NC}"
else
    echo -e "${RED}Error: Working directory has uncommitted changes${NC}"
    echo "Please commit or stash your changes first"
    exit 1
fi

if git rev-parse "v${NEW_VERSION}" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag v${NEW_VERSION} already exists${NC}"
    exit 1
fi

read -p "Create release v${NEW_VERSION}? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release cancelled"
    exit 0
fi

echo ""
echo "Updating VERSION file..."
echo "$NEW_VERSION" > VERSION

echo "Updating pkg/cli/root.go..."
sed -i '' "s/Version = \".*\"/Version = \"${NEW_VERSION}\"/" pkg/cli/root.go

echo "Updating pkg/lsp/server.go..."
sed -i '' "s/Version: \".*\"/Version: \"${NEW_VERSION}\"/" pkg/lsp/server.go

echo "Committing version bump..."
git add VERSION pkg/cli/root.go pkg/lsp/server.go
git commit -m "chore: bump version to v${NEW_VERSION}"

echo "Creating git tag v${NEW_VERSION}..."
git tag -a "v${NEW_VERSION}" -m "Release v${NEW_VERSION}"

echo ""
echo -e "${GREEN}✓ Release v${NEW_VERSION} created successfully!${NC}"
echo ""
echo "Next steps:"
echo "  1. Review the commit and tag"
echo "  2. Push changes:  git push origin main"
echo "  3. Push tag:      git push origin v${NEW_VERSION}"
echo ""
echo "To undo (if needed):"
echo "  git tag -d v${NEW_VERSION}"
echo "  git reset --hard HEAD~1"
