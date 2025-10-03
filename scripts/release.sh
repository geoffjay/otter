#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_error() { echo -e "${RED}ERROR: $1${NC}"; }
print_success() { echo -e "${GREEN}SUCCESS: $1${NC}"; }
print_info() { echo -e "${YELLOW}INFO: $1${NC}"; }
print_step() { echo -e "${BLUE}>>> $1${NC}"; }

# Check if version argument is provided
if [ -z "$1" ]; then
  print_error "Version number required"
  echo "Usage: ./scripts/release.sh <version>"
  echo "Example: ./scripts/release.sh v1.0.0"
  exit 1
fi

NEW_VERSION="$1"

# Validate version format (vX.Y.Z)
if ! [[ $NEW_VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  print_error "Invalid version format. Use vX.Y.Z (e.g., v1.0.0)"
  exit 1
fi

echo ""
print_step "Starting release process for $NEW_VERSION"
echo ""

# Step 1: Verify current state
print_step "Step 1: Verifying current state"

print_info "Checking current branch..."
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
  print_error "Must be on main branch. Currently on: $CURRENT_BRANCH"
  exit 1
fi
print_success "On main branch"

print_info "Checking working directory..."
if ! git diff-index --quiet HEAD --; then
  print_error "Working directory has uncommitted changes"
  git status
  exit 1
fi
print_success "Working directory is clean"

print_info "Checking if tag already exists..."
if git rev-parse "$NEW_VERSION" >/dev/null 2>&1; then
  print_error "Tag $NEW_VERSION already exists"
  exit 1
fi
print_success "Tag does not exist yet"

print_info "Pulling latest changes..."
git pull origin main
print_success "Up to date with origin/main"

# Step 2: Run tests and build
print_step "Step 2: Running tests and build"

print_info "Running tests..."
if ! make test; then
  print_error "Tests failed"
  exit 1
fi
print_success "All tests passed"

print_info "Building project..."
if ! make build; then
  print_error "Build failed"
  exit 1
fi
print_success "Build succeeded"

# Step 3: Generate changelog
print_step "Step 3: Generating changelog"

LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || git rev-list --max-parents=0 HEAD)
print_info "Generating changelog from $LAST_TAG to HEAD"

# Get commit counts by type
FEAT_COUNT=$(git log $LAST_TAG..HEAD --oneline --grep="^feat" --no-merges | wc -l | tr -d ' ')
FIX_COUNT=$(git log $LAST_TAG..HEAD --oneline --grep="^fix" --no-merges | wc -l | tr -d ' ')
DOCS_COUNT=$(git log $LAST_TAG..HEAD --oneline --grep="^docs" --no-merges | wc -l | tr -d ' ')
OTHER_COUNT=$(git log $LAST_TAG..HEAD --oneline --grep="^chore\|^refactor\|^test\|^ci\|^perf\|^style" --no-merges | wc -l | tr -d ' ')

# Generate sections
FEATURES=$(git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^feat" --no-merges)
FIXES=$(git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^fix" --no-merges)
DOCS=$(git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^docs" --no-merges)
MAINTENANCE=$(git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^chore\|^refactor\|^test\|^ci\|^perf\|^style" --no-merges)

# Build changelog
CHANGELOG="Release $NEW_VERSION"

if [ ! -z "$FEATURES" ]; then
  CHANGELOG="$CHANGELOG

## 🎉 Features
$FEATURES"
fi

if [ ! -z "$FIXES" ]; then
  CHANGELOG="$CHANGELOG

## 🐛 Bug Fixes
$FIXES"
fi

if [ ! -z "$DOCS" ]; then
  CHANGELOG="$CHANGELOG

## 📚 Documentation
$DOCS"
fi

if [ ! -z "$MAINTENANCE" ]; then
  CHANGELOG="$CHANGELOG

## 🔧 Maintenance
$MAINTENANCE"
fi

# Display summary
echo ""
print_info "Changelog summary:"
echo "  - Features: $FEAT_COUNT"
echo "  - Bug Fixes: $FIX_COUNT"
echo "  - Documentation: $DOCS_COUNT"
echo "  - Maintenance: $OTHER_COUNT"
echo ""

# Display full changelog
echo "================================"
echo "$CHANGELOG"
echo "================================"
echo ""

# Step 4: Confirm release
print_step "Step 4: Confirming release"
read -p "Create release $NEW_VERSION with this changelog? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  print_info "Release cancelled"
  exit 0
fi

# Step 5: Create tag
print_step "Step 5: Creating tag"
print_info "Creating annotated tag $NEW_VERSION..."
git tag -a "$NEW_VERSION" -m "$CHANGELOG"
print_success "Tag created"

# Step 6: Push tag
print_step "Step 6: Pushing tag to origin"
print_info "Pushing tag $NEW_VERSION..."
git push origin "$NEW_VERSION"
print_success "Tag pushed to origin"

echo ""
print_success "Release $NEW_VERSION created successfully!"
echo ""
print_info "Next steps:"
echo "  1. GitHub Actions will automatically build and publish release artifacts"
echo "  2. View the release at: https://github.com/geoffjay/otter/releases/tag/$NEW_VERSION"
echo "  3. Create GitHub release notes at: https://github.com/geoffjay/otter/releases/new?tag=$NEW_VERSION"
echo ""

# Step 7: Create GitHub release (if gh CLI is available)
if command -v gh &> /dev/null; then
  read -p "Create GitHub release now? (y/N) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_step "Step 7: Creating GitHub release"
    print_info "Creating release on GitHub..."
    gh release create "$NEW_VERSION" \
      --title "Release $NEW_VERSION" \
      --notes "$CHANGELOG" \
      --verify-tag
    print_success "GitHub release created!"
    echo ""
    print_info "View release at: https://github.com/geoffjay/otter/releases/tag/$NEW_VERSION"
  fi
else
  print_info "GitHub CLI (gh) not installed. To create release manually:"
  echo "  gh release create $NEW_VERSION --title \"Release $NEW_VERSION\" --notes-file <(echo \"$CHANGELOG\")"
  echo "  Or visit: https://github.com/geoffjay/otter/releases/new?tag=$NEW_VERSION"
fi

echo ""
print_success "🎉 Release process complete!"
