#!/bin/bash
set -e

# Wade installer script
# Usage: curl -fsSL https://github.com/wadefengx/wade/releases/latest/download/install.sh | bash

REPO="wadefengx/wade"
DEFAULT_INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BOLD}🏄 Wade Installer${NC}\n"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  darwin)
    case "$ARCH" in
      arm64) PLATFORM="darwin-arm64" ;;
      x86_64|amd64) PLATFORM="darwin-amd64" ;;
      *) echo -e "${RED}Error: unsupported architecture $ARCH on macOS${NC}"; exit 1 ;;
    esac
    ;;
  linux)
    case "$ARCH" in
      x86_64|amd64) PLATFORM="linux-amd64" ;;
      aarch64|arm64) PLATFORM="linux-arm64" ;;
      *) echo -e "${RED}Error: unsupported architecture $ARCH on Linux${NC}"; exit 1 ;;
    esac
    ;;
  *)
    echo -e "${RED}Error: unsupported OS $OS${NC}"
    echo "Windows users: please use Scoop or download the binary manually."
    exit 1
    ;;
esac

# Get latest version — via HTTP redirect (no API rate limits)
# https://github.com/REPO/releases/latest → 302 → .../releases/tag/vX.Y.Z
echo "Fetching latest release..."
LATEST=$(curl -sfIL "https://github.com/$REPO/releases/latest" | grep -i '^location:' | sed -E 's|.*/tag/([^/]+).*|\1|' | tr -d '\r')

if [ -z "$LATEST" ]; then
  echo -e "${RED}Error: could not determine latest version${NC}"
  exit 1
fi
echo "Latest version: $LATEST"

# Download URL
FILENAME="wade-$PLATFORM.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST/$FILENAME"

echo -e "Downloading ${CYAN}$FILENAME${NC}..."
TMP_DIR=$(mktemp -d)
curl -fsSL "$URL" -o "$TMP_DIR/$FILENAME"

# Extract
echo "Extracting..."
tar xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

# Install
INSTALL_DIR="${WADE_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
if [ ! -w "$INSTALL_DIR" ]; then
  echo -e "${RED}Need sudo to install to $INSTALL_DIR${NC}"
  sudo cp "$TMP_DIR/wade" "$INSTALL_DIR/wade"
  sudo chmod +x "$INSTALL_DIR/wade"
else
  cp "$TMP_DIR/wade" "$INSTALL_DIR/wade"
  chmod +x "$INSTALL_DIR/wade"
fi

# Cleanup
rm -rf "$TMP_DIR"

echo -e "${GREEN}wade $LATEST installed to $INSTALL_DIR/wade${NC}"

# Setup
echo -e "\nRunning ${BOLD}wade setup${NC}..."
"$INSTALL_DIR/wade" setup

# PATH reminder
SHELL_RC=""
case "$SHELL" in
  */zsh) SHELL_RC="$HOME/.zshrc" ;;
  */bash) SHELL_RC="$HOME/.bashrc" ;;
esac

if [ -n "$SHELL_RC" ]; then
  if ! grep -q "\.wade/shims" "$SHELL_RC" 2>/dev/null; then
    echo ""
    echo -e "${CYAN}Add this to $SHELL_RC:${NC}"
    echo -e "  export PATH=\"\$HOME/.wade/shims:\$PATH\""
  fi
fi

echo -e "\n${GREEN}Done!${NC} Try: ${BOLD}wade status${NC}"
