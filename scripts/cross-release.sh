#!/usr/bin/env bash
# cross-release.sh — build and upload Linux and Windows binaries for a Spectre release.
#
# Usage:
#   scripts/cross-release.sh v0.4.0
#
# Reads GITHUB_TOKEN from .env in the repo root (never commit that file).
# Copy .env.example to .env and fill in your token before running.
#
# What it does:
#   1. Loads GITHUB_TOKEN from .env
#   2. Cross-compiles linux/amd64, linux/arm64, windows/amd64
#   3. Packages each into .tar.gz / .zip with README, LICENSE, examples, book PDF
#   4. Generates a checksums file
#   5. Uploads all assets to the GitHub release for the given tag

set -euo pipefail

# ── args ──────────────────────────────────────────────────────────────────────
VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version-tag>  (e.g. $0 v0.4.0)" >&2
  exit 1
fi

VER="${VERSION#v}"
REPO="BoundlessProducts/spectre"
DIST="dist/cross-release"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# ── load .env ─────────────────────────────────────────────────────────────────
ENV_FILE="$REPO_ROOT/.env"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "Error: $ENV_FILE not found." >&2
  echo "Copy .env.example to .env and set GITHUB_TOKEN." >&2
  exit 1
fi
# shellcheck source=/dev/null
set -o allexport
source "$ENV_FILE"
set +o allexport

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "Error: GITHUB_TOKEN is not set in $ENV_FILE." >&2
  exit 1
fi

# ── check gh ──────────────────────────────────────────────────────────────────
if ! command -v gh &>/dev/null; then
  echo "Error: gh CLI not found. Install with: brew install gh" >&2
  exit 1
fi

# ── verify release exists ─────────────────────────────────────────────────────
if ! gh release view "$VERSION" --repo "$REPO" &>/dev/null; then
  echo "No GitHub release found for $VERSION in $REPO." >&2
  echo "Create the tag and release first, then run this script." >&2
  exit 1
fi

# ── build targets ─────────────────────────────────────────────────────────────
# Format: "GOOS GOARCH archive-extension"
TARGETS=(
  "linux   amd64 tar.gz"
  "linux   arm64 tar.gz"
  "windows amd64 zip"
)

rm -rf "$DIST"
mkdir -p "$DIST"

cd "$REPO_ROOT"

for target in "${TARGETS[@]}"; do
  read -r GOOS GOARCH EXT <<< "$target"

  BINARY="spectre"
  [[ "$GOOS" == "windows" ]] && BINARY="spectre.exe"

  STEM="spectre_${VER}_${GOOS}_${GOARCH}"
  STAGEDIR="$DIST/$STEM"
  mkdir -p "$STAGEDIR"

  echo "==> Building $GOOS/$GOARCH..."
  GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 \
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

  echo "==> Packaging ${STEM}.${EXT}..."
  if [[ "$EXT" == "zip" ]]; then
    (cd "$DIST" && zip -qr "${STEM}.zip" "$STEM")
  else
    tar -czf "$DIST/${STEM}.tar.gz" -C "$DIST" "$STEM"
  fi
  rm -rf "$STAGEDIR"
done

# ── checksums ─────────────────────────────────────────────────────────────────
echo "==> Generating checksums..."
(cd "$DIST" && shasum -a 256 ./*.tar.gz ./*.zip 2>/dev/null > checksums-cross.txt)
cat "$DIST/checksums-cross.txt"

# ── upload ────────────────────────────────────────────────────────────────────
echo "==> Uploading to GitHub release $VERSION..."
gh release upload "$VERSION" \
  "$DIST"/spectre_"${VER}"_linux_amd64.tar.gz \
  "$DIST"/spectre_"${VER}"_linux_arm64.tar.gz \
  "$DIST"/spectre_"${VER}"_windows_amd64.zip \
  "$DIST/checksums-cross.txt" \
  --repo "$REPO" \
  --clobber

echo ""
echo "Done. Linux and Windows assets uploaded to:"
echo "  https://github.com/$REPO/releases/tag/$VERSION"
