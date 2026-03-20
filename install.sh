#!/usr/bin/env bash
set -euo pipefail

REPO="pozii/RepoSearcher"
BINARY="repo-searcher"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and ARCH
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"

echo "Downloading ${BINARY} for ${OS}/${ARCH}..."

# Get latest release tag
LATEST=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4)

if [ -z "$LATEST" ]; then
    echo "Error: Could not fetch latest release"
    exit 1
fi

echo "Latest version: ${LATEST}"

# Download binary
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"
TMPFILE=$(mktemp)

curl -sL "$DOWNLOAD_URL" -o "$TMPFILE"
chmod +x "$TMPFILE"

# Install
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMPFILE" "${INSTALL_DIR}/${BINARY}"
else
    sudo mv "$TMPFILE" "${INSTALL_DIR}/${BINARY}"
fi

echo ""
echo "✅ ${BINARY} ${LATEST} installed to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Run '${BINARY} version' to verify installation."
