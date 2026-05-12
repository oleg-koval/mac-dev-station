#!/usr/bin/env bash
set -euo pipefail

REPO="oleg-koval/mac-dev-station"
BINARY="mac-dev-station"
INSTALL_DIR="${HOME}/.local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [[ "$OS" != "darwin" ]]; then
  echo "error: mac-dev-station supports macOS only." >&2
  exit 1
fi

case "$ARCH" in
  arm64)  ARCH="arm64" ;;
  x86_64) ARCH="amd64" ;;
  *)
    echo "error: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Resolve latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed -E 's/.*"([^"]+)".*/\1/')

if [[ -z "$LATEST" ]]; then
  echo "error: could not determine latest release." >&2
  exit 1
fi

# GoReleaser uses .Version (tag without leading v) in the archive name template
VERSION="${LATEST#v}"
ARCHIVE="${BINARY}-${VERSION}-${OS}-${ARCH}.zip"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "→ Downloading ${BINARY} ${LATEST} (${OS}/${ARCH})..."
curl -fsSL "$URL" -o "${TMP}/${ARCHIVE}"

echo "→ Extracting..."
unzip -q "${TMP}/${ARCHIVE}" -d "$TMP"

mkdir -p "$INSTALL_DIR"
mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo "✓ Installed to ${INSTALL_DIR}/${BINARY}"

# PATH hint if needed
if ! command -v "$BINARY" &>/dev/null; then
  echo ""
  echo "  Add ${INSTALL_DIR} to your PATH:"
  echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
  echo ""
fi

echo "Run: mac-dev-station"
