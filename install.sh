#!/bin/sh
set -e

REPO="Aaronmacaron/dict"
BINARY="dict"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  ;;
  darwin) ;;
  *) echo "error: unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) echo "error: unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Fetch latest release tag from GitHub API
VERSION=$(curl -sf "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "error: could not determine latest version" >&2
  exit 1
fi

ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

echo "Installing ${BINARY} ${VERSION} (${OS}/${ARCH})..."

# Download to a temp directory, cleaned up on exit
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -sfL "${BASE_URL}/${ARCHIVE}"      -o "${TMP}/${ARCHIVE}"
curl -sfL "${BASE_URL}/checksums.txt"   -o "${TMP}/checksums.txt"

# Verify checksum
cd "$TMP"
if command -v sha256sum > /dev/null 2>&1; then
  grep "${ARCHIVE}" checksums.txt | sha256sum -c -
elif command -v shasum > /dev/null 2>&1; then
  grep "${ARCHIVE}" checksums.txt | shasum -a 256 -c -
else
  echo "warning: no sha256 tool found, skipping checksum verification" >&2
fi

# Extract binary
tar -xzf "${ARCHIVE}" -C "$TMP"

# Install — use sudo only if the target directory isn't writable
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} requires sudo..."
  sudo mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo "Installed: $(command -v ${BINARY})"
echo "Run '${BINARY} --help' to get started."
