---
description: Create a new release with semantic versioning and automated changelog generation
---

# Release Command 🚀

This command helps you create a new release for the Otter project with proper semantic versioning, automated changelog generation from commit history, and git tag creation.

## Prerequisites

Before running this command, ensure:

1. You are on the `main` branch
2. Your working directory is clean (no uncommitted changes)
3. You have pulled the latest changes from origin
4. All tests pass (`make test`)
5. The build succeeds (`make build`)

## Release Process

### Step 1: Verify Current State

```bash
# Check current branch
git branch --show-current

# Verify clean working directory
git status

# Pull latest changes
git pull origin main

# Run tests
make test

# Build the project
make build
```

### Step 2: Determine Version Number

Follow [Semantic Versioning](https://semver.org/) (MAJOR.MINOR.PATCH):

- **MAJOR**: Incompatible API changes (breaking changes)
- **MINOR**: New functionality in a backwards-compatible manner
- **PATCH**: Backwards-compatible bug fixes

**Examples:**
- `v1.0.0` → `v2.0.0` (breaking changes)
- `v1.0.0` → `v1.1.0` (new features)
- `v1.0.0` → `v1.0.1` (bug fixes)

### Step 3: Generate Changelog

Generate changelog from commit history since the last tag:

```bash
# Get the last release tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

# If this is the first release
if [ -z "$LAST_TAG" ]; then
  LAST_TAG=$(git rev-list --max-parents=0 HEAD)
fi

# Generate changelog grouped by type
echo "# Changelog for v<NEW_VERSION>"
echo ""
echo "## 🎉 Features"
git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^feat" --no-merges
echo ""
echo ""
echo "## 🐛 Bug Fixes"
git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^fix" --no-merges
echo ""
echo ""
echo "## 📚 Documentation"
git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^docs" --no-merges
echo ""
echo ""
echo "## 🔧 Maintenance"
git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^chore\|^refactor\|^test\|^ci\|^perf\|^style" --no-merges
echo ""
```

### Step 4: Create Release Tag

```bash
# Set the new version
NEW_VERSION="v1.0.0"  # Update this with the actual version

# Create annotated tag with changelog
git tag -a $NEW_VERSION -m "Release $NEW_VERSION

## 🎉 Features
- feat: implement hooks and lifecycle events system (2e7ec71)
- feat: implement template variable processing in layer files (153d61a)

## 🐛 Bug Fixes
- fix: resolve edge case in variable substitution (abc1234)

## 📚 Documentation
- docs: add comprehensive task list (def5678)

## 🔧 Maintenance
- chore: update dependencies (ghi9012)
"

# Verify tag was created
git tag -n9 $NEW_VERSION
```

### Step 5: Push Release

```bash
# Push the tag to remote
git push origin $NEW_VERSION

# Or push all tags
git push origin --tags
```

### Step 6: Create GitHub Release (Optional)

If using GitHub, create a release from the tag:

```bash
# Using GitHub CLI (gh)
gh release create $NEW_VERSION \
  --title "Release $NEW_VERSION" \
  --notes "$(git tag -l --format='%(contents)' $NEW_VERSION)" \
  --verify-tag

# Or manually visit:
# https://github.com/geoffjay/otter/releases/new?tag=$NEW_VERSION
```

## Automated Release Script

You can create a helper script at `scripts/release.sh`:

```bash
#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_error() { echo -e "${RED}ERROR: $1${NC}"; }
print_success() { echo -e "${GREEN}SUCCESS: $1${NC}"; }
print_info() { echo -e "${YELLOW}INFO: $1${NC}"; }

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

print_info "Starting release process for $NEW_VERSION"

# Step 1: Verify current state
print_info "Checking current branch..."
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
  print_error "Must be on main branch. Currently on: $CURRENT_BRANCH"
  exit 1
fi

print_info "Checking working directory..."
if ! git diff-index --quiet HEAD --; then
  print_error "Working directory has uncommitted changes"
  git status
  exit 1
fi

print_info "Pulling latest changes..."
git pull origin main

# Step 2: Run tests and build
print_info "Running tests..."
if ! make test; then
  print_error "Tests failed"
  exit 1
fi

print_info "Building project..."
if ! make build; then
  print_error "Build failed"
  exit 1
fi

# Step 3: Generate changelog
print_info "Generating changelog..."
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || git rev-list --max-parents=0 HEAD)
print_info "Changes since $LAST_TAG"

CHANGELOG="Release $NEW_VERSION

## 🎉 Features
$(git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^feat" --no-merges)

## 🐛 Bug Fixes
$(git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^fix" --no-merges)

## 📚 Documentation
$(git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^docs" --no-merges)

## 🔧 Maintenance
$(git log $LAST_TAG..HEAD --pretty=format:"- %s (%h)" --grep="^chore\|^refactor\|^test\|^ci\|^perf\|^style" --no-merges)
"

# Display changelog
echo ""
echo "================================"
echo "$CHANGELOG"
echo "================================"
echo ""

# Step 4: Confirm release
read -p "Create release $NEW_VERSION with this changelog? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  print_info "Release cancelled"
  exit 0
fi

# Step 5: Create tag
print_info "Creating tag $NEW_VERSION..."
git tag -a "$NEW_VERSION" -m "$CHANGELOG"

# Step 6: Push tag
print_info "Pushing tag to origin..."
git push origin "$NEW_VERSION"

print_success "Release $NEW_VERSION created successfully!"
print_info "Next steps:"
echo "  1. GitHub Actions will automatically build and publish release artifacts"
echo "  2. Create GitHub release at: https://github.com/geoffjay/otter/releases/new?tag=$NEW_VERSION"
echo "  3. Update documentation if needed"
echo "  4. Announce the release"

# Step 7: Create GitHub release (if gh CLI is available)
if command -v gh &> /dev/null; then
  read -p "Create GitHub release now? (y/N) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_info "Creating GitHub release..."
    gh release create "$NEW_VERSION" \
      --title "Release $NEW_VERSION" \
      --notes "$CHANGELOG" \
      --verify-tag
    print_success "GitHub release created!"
  fi
else
  print_info "GitHub CLI (gh) not found. Create release manually at:"
  echo "  https://github.com/geoffjay/otter/releases/new?tag=$NEW_VERSION"
fi
```

Make the script executable:

```bash
chmod +x scripts/release.sh
```

## Usage Examples

### Manual Release

```bash
# Example: Releasing v1.2.3
NEW_VERSION="v1.2.3"

# 1. Verify prerequisites
git checkout main
git pull origin main
make test
make build

# 2. Generate changelog
LAST_TAG=$(git describe --tags --abbrev=0)
git log $LAST_TAG..HEAD --oneline

# 3. Create tag with changelog
git tag -a $NEW_VERSION -m "Release $NEW_VERSION

## Features
- feat: add new awesome feature

## Bug Fixes
- fix: resolve critical bug
"

# 4. Push tag
git push origin $NEW_VERSION

# 5. Create GitHub release
gh release create $NEW_VERSION \
  --title "Release $NEW_VERSION" \
  --notes-file CHANGELOG.md
```

### Automated Release

```bash
# Using the release script
./scripts/release.sh v1.2.3

# The script will:
# 1. Verify you're on main branch
# 2. Check for uncommitted changes
# 3. Pull latest changes
# 4. Run tests
# 5. Build the project
# 6. Generate changelog from commits
# 7. Show changelog and ask for confirmation
# 8. Create and push tag
# 9. Optionally create GitHub release
```

## Changelog Format

The changelog follows the [Keep a Changelog](https://keepachangelog.com/) format with sections:

- **🎉 Features** - New features (`feat:` commits)
- **🐛 Bug Fixes** - Bug fixes (`fix:` commits)
- **📚 Documentation** - Documentation changes (`docs:` commits)
- **🔧 Maintenance** - Chores, refactoring, tests, CI, performance, style (`chore:`, `refactor:`, `test:`, `ci:`, `perf:`, `style:` commits)
- **⚠️ Breaking Changes** - Breaking changes (commits with `BREAKING CHANGE:` in body or footer)

## Release Checklist

- [ ] All tests pass
- [ ] Build succeeds
- [ ] Documentation is up to date
- [ ] CHANGELOG.md is updated (if maintained separately)
- [ ] Version number follows semantic versioning
- [ ] Commit history is clean and follows conventional commits
- [ ] Tag annotation includes comprehensive changelog
- [ ] Tag is pushed to remote
- [ ] GitHub release is created (if applicable)
- [ ] Release artifacts are generated by CI/CD
- [ ] Release is announced (if applicable)

## Rollback a Release

If you need to rollback a release:

```bash
# Delete local tag
git tag -d v1.2.3

# Delete remote tag
git push origin :refs/tags/v1.2.3

# Delete GitHub release (if created)
gh release delete v1.2.3 --yes
```

## Tips

1. **Use Conventional Commits**: Ensure all commits follow the conventional commits format for accurate changelog generation
2. **Review Changelog**: Always review the generated changelog before creating the tag
3. **Test First**: Always run tests and build before creating a release
4. **Backup**: Consider creating a backup branch before releasing: `git branch backup-pre-v1.2.3`
5. **Automate**: Use the release script to automate and standardize the process
6. **GitHub Actions**: The project's CI/CD pipeline will automatically build artifacts when a tag is pushed

## Related Commands

- `/git-commit` - Git commit conventions
- `make test` - Run all tests
- `make build` - Build the project
- `make build-all` - Build for all platforms
- `gh release create` - Create GitHub release (requires GitHub CLI)

## References

- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github)
