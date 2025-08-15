#!/bin/bash

set -e

# Check if CHANGELOG.md file exists
if [ ! -f CHANGELOG.md ]; then
  echo "CHANGELOG.md not found."
  exit 1
fi

# Extract the latest version number from CHANGELOG.md (compatible with macOS/Linux)
CURRENT_VERSION=$(grep -E '^##[[:space:]]+\[?v?([0-9]+\.[0-9]+\.[0-9]+)\]?' CHANGELOG.md | head -1 | sed -E 's/[^0-9]*([0-9]+\.[0-9]+\.[0-9]+).*/\1/')

if [ -z "$CURRENT_VERSION" ]; then
  echo "Failed to parse version from CHANGELOG.md"
  exit 1
fi

# Confirm the current branch is main or master
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "master" ]]; then
  echo "Current branch is '$CURRENT_BRANCH'. Are you sure? [y/N]"
  read -r confirm
  if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    echo "Release aborted."
    exit 1
  fi
fi

# Add CHANGELOG.md and commit the changes
git add CHANGELOG.md
git commit -m "chore: release v$CURRENT_VERSION"

# Create Git tag
git tag -a "v$CURRENT_VERSION" -m "Release v$CURRENT_VERSION"

# Push code and tag to two repositories
echo "Pushing to origin..."
git push origin "$CURRENT_BRANCH" --tags

echo "Pushing to github..."
git push github "$CURRENT_BRANCH" --tags

echo "Release v$CURRENT_VERSION completed!"
