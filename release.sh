#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

# --- Helpers ---
die() { echo "ERROR: $*" >&2; exit 1; }

# --- Usage ---
usage() {
    cat <<EOF
Usage: ./release.sh <version>

Creates a release tag and pushes it to trigger the GitHub Actions release workflow.

Examples:
  ./release.sh v1.0.0
  ./release.sh v1.2.3

The version must follow semver format: vMAJOR.MINOR.PATCH
EOF
    exit 1
}

# --- Validate arguments ---
[[ $# -eq 1 ]] || usage
VERSION="$1"

# Validate semver format (vX.Y.Z)
[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "Version '$VERSION' does not match semver format (vX.Y.Z)"

# --- Pre-flight checks ---
command -v git >/dev/null 2>&1 || die "git is not installed"

# Ensure we're in a git repo
git rev-parse --git-dir >/dev/null 2>&1 || die "Not a git repository"

# Ensure working tree is clean
if [[ -n "$(git status --porcelain)" ]]; then
    echo "Dirty working tree:"
    git status --short
    die "Commit or stash your changes before releasing"
fi

# Ensure we're on main branch
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$BRANCH" != "main" && "$BRANCH" != "master" ]]; then
    echo "WARNING: You are on branch '$BRANCH', not main/master."
    read -rp "Continue anyway? [y/N] " confirm
    [[ "$confirm" =~ ^[Yy]$ ]] || exit 0
fi

# Ensure tag doesn't already exist
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    die "Tag '$VERSION' already exists"
fi

# Ensure remote is reachable
REMOTE="${REMOTE:-origin}"
git ls-remote --exit-code "$REMOTE" >/dev/null 2>&1 || die "Cannot reach remote '$REMOTE'"

# Ensure local branch is up to date with remote
git fetch "$REMOTE" --tags --quiet
LOCAL_SHA="$(git rev-parse HEAD)"
REMOTE_SHA="$(git rev-parse "$REMOTE/$BRANCH" 2>/dev/null || echo "")"
if [[ -n "$REMOTE_SHA" && "$LOCAL_SHA" != "$REMOTE_SHA" ]]; then
    die "Local branch is not up to date with $REMOTE/$BRANCH. Pull or push first."
fi

# --- Confirmation ---
echo ""
echo "  Release: $VERSION"
echo "  Branch:  $BRANCH"
echo "  Commit:  $(git rev-parse --short HEAD)"
echo "  Remote:  $REMOTE"
echo ""
read -rp "Create and push tag '$VERSION'? [y/N] " confirm
[[ "$confirm" =~ ^[Yy]$ ]] || { echo "Aborted."; exit 0; }

# --- Create and push tag ---
echo ""
echo "==> Creating annotated tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

echo "==> Pushing tag to $REMOTE..."
git push "$REMOTE" "$VERSION"

echo ""
echo "==> Done! Tag '$VERSION' pushed to $REMOTE."
echo "    GitHub Actions will build and create the release."
echo "    Watch progress at: https://github.com/tafli/calling-parents/actions"
