#!/usr/bin/env bash
set -euo pipefail

REPO="eggnita/adapto_cms_cli"
BINARY="adapto"
INSTALL_DIR="/usr/local/bin"

# Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

SUFFIX=""
if [ "$OS" = "windows" ]; then
  SUFFIX=".exe"
fi

# Get latest version
if command -v curl &>/dev/null; then
  LATEST=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
elif command -v wget &>/dev/null; then
  LATEST=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
else
  echo "Error: curl or wget required"
  exit 1
fi

if [ -z "$LATEST" ]; then
  echo "Error: could not determine latest version"
  exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}-${OS}-${ARCH}${SUFFIX}"

echo "Installing ${BINARY} ${LATEST} for ${OS}/${ARCH}..."

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if command -v curl &>/dev/null; then
  curl -sL "$DOWNLOAD_URL" -o "${TMP_DIR}/${BINARY}${SUFFIX}"
else
  wget -q "$DOWNLOAD_URL" -O "${TMP_DIR}/${BINARY}${SUFFIX}"
fi

chmod +x "${TMP_DIR}/${BINARY}${SUFFIX}"

if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP_DIR}/${BINARY}${SUFFIX}" "${INSTALL_DIR}/${BINARY}${SUFFIX}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMP_DIR}/${BINARY}${SUFFIX}" "${INSTALL_DIR}/${BINARY}${SUFFIX}"
fi

echo "Installed ${BINARY} ${LATEST} to ${INSTALL_DIR}/${BINARY}${SUFFIX}"
echo "Run '${BINARY} --help' to get started."
