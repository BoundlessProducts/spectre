#!/usr/bin/env bash
# mac-release.sh — build and upload macOS binaries for a Spectre release.
#
# Usage:
#   scripts/mac-release.sh v0.4.0
#   GITHUB_TOKEN=ghp_xxx scripts/mac-release.sh v0.4.0
#
# What it does:
#   1. Runs local CI checks (build + test + vet)
#   2. Builds darwin/arm64 and darwin/amd64 binaries natively
#   3. Packages each into a .tar.gz with README, LICENSE, examples, and the book PDF
#   4. Generates a checksums file
#   5. Uploads the archives to the GitHub release for the given tag (requires gh + GITHUB_TOKEN)

set -euo pipefail

# ── args ──────────────────────────────────────────────────────────────────────
VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version-tag>  (e.g. $0 v0.4.0)" >&2
  exit 1
fi

# Strip leading 'v' for filenames
VER="${VERSION#v}"
REPO="BoundlessProducts/spectre"
DIST="dist/mac-release"

# ── ci checks ─────────────────────────────────────────────────────────────────
echo "==> Running local CI checks..."
go build -v ./...
go test -race ./...
go vet ./...
echo "    CI checks passed."

# ── build ─────────────────────────────────────────────────────────────────────
rm -rf "$DIST"
mkdir -p "$DIST"

for ARCH in arm64 amd64; do
  BINARY="spectre"
  ARCHIVE="spectre_${VER}_darwin_${ARCH}.tar.gz"
  STAGEDIR="$DIST/spectre_${VER}_darwin_${ARCH}"

  echo "==> Building darwin/$ARCH..."
  GOOS=darwin GOARCH="$ARCH" CGO_ENABLED=0 \
    go build -ldflags "-s -w -X main.version=${VER}" \
    -o "$STAGEDIR/$BINARY" ./cmd/spectre

  # Bundle extras
  cp README.md LICENSE "$STAGEDIR/"
  mkdir -p "$STAGEDIR/examples"
  cp examples/*.spec "$STAGEDIR/examples/" 2>/dev/null || true
  if [[ -f docs/spectre-book.pdf ]]; then
    mkdir -p "$STAGEDIR/docs"
    cp docs/spectre-book.pdf "$STAGEDIR/docs/"
  fi

  echo "==> Packaging $ARCHIVE..."
  tar -czf "$DIST/$ARCHIVE" -C "$DIST" "spectre_${VER}_darwin_${ARCH}"
  rm -rf "$STAGEDIR"
done

# ── checksums ─────────────────────────────────────────────────────────────────
echo "==> Generating checksums..."
(cd "$DIST" && shasum -a 256 ./*.tar.gz > checksums-darwin.txt)
cat "$DIST/checksums-darwin.txt"

# ── upload ────────────────────────────────────────────────────────────────────
if ! command -v gh &>/dev/null; then
  echo ""
  echo "gh CLI not found — archives are in $DIST. Upload manually or install gh:"
  echo "  brew install gh"
  exit 0
fi

if ! gh release view "$VERSION" --repo "$REPO" &>/dev/null; then
  echo "No GitHub release found for $VERSION in $REPO." >&2
  echo "Create the release first (push the tag and run goreleaser, or create it via gh release create)." >&2
  echo "Archives are ready in $DIST." >&2
  exit 1
fi

echo "==> Uploading to GitHub release $VERSION..."
gh release upload "$VERSION" \
  "$DIST"/spectre_"${VER}"_darwin_arm64.tar.gz \
  "$DIST"/spectre_"${VER}"_darwin_amd64.tar.gz \
  "$DIST/checksums-darwin.txt" \
  --repo "$REPO" \
  --clobber

echo ""
echo "Done. macOS assets uploaded to:"
echo "  https://github.com/$REPO/releases/tag/$VERSION"
