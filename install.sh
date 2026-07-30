#!/bin/sh
# PPHLX Universal Toolchain Installer for macOS and Linux
# Installs the native Go compiler binary to $HOME/.pphlx/bin/pphlx without requiring sudo.

set -e

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
CLEAR='\033[0m'

echo "${CYAN}------------------------------------------------${CLEAR}"
echo "           PPHLX Installer Boot Sequence       "
echo "${CYAN}------------------------------------------------${CLEAR}"

# 1. Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    darwin*) OS="darwin" ;;
    linux*) OS="linux" ;;
    *)
        echo "${RED}Error: Unsupported operating system: $OS${CLEAR}"
        exit 1
        ;;
esac

# 2. Detect CPU Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo "${RED}Error: Unsupported CPU architecture: $ARCH${CLEAR}"
        exit 1
        ;;
esac

# 3. Resolve Download URL (GitHub Releases CDN)
REPO="pphlx/pphlx"
VERSION="latest" # Dynamically fetch latest release version tag

if [ "$VERSION" = "latest" ]; then
    # Query GitHub API to get the latest tag
    TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -o '\"tag_name\": \"[^\"]*\"' | head -n 1 | cut -d '"' -f 4)
    if [ -z "$TAG" ]; then
        TAG="v1.1.6"
    fi
else
    TAG="$VERSION"
fi

BINARY_NAME="pphlx-$OS-$ARCH"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/$BINARY_NAME.tar.gz"

echo "Detected System: ${GREEN}$OS ($ARCH)${CLEAR}"
echo "Target Release:  ${GREEN}$TAG${CLEAR}"
echo "Downloading from: $DOWNLOAD_URL"

# 4. Download and Extract to Temp Directory
TMP_DIR=$(mktemp -d)
CLEANUP() {
    rm -rf "$TMP_DIR"
}
trap CLEANUP EXIT

TAR_PATH="$TMP_DIR/pphlx.tar.gz"

# Fetch binary
curl -fsSL "$DOWNLOAD_URL" -o "$TAR_PATH" || {
    echo "${RED}Error: Failed to download the PPHLX binary archive. It might not be released yet for this platform.${CLEAR}"
    exit 1
}

# Extract binary
tar -xzf "$TAR_PATH" -C "$TMP_DIR"

# 5. Move binary to user directory (~/.pphlx/bin)
INSTALL_DIR="$HOME/.pphlx/bin"
mkdir -p "$INSTALL_DIR"
DEST="$INSTALL_DIR/pphlx"

echo "Installing binary to: $DEST"
mv "$TMP_DIR/pphlx" "$DEST"
chmod +x "$DEST"

# 6. Add ~/.pphlx/bin to user PATH in shell profile files
SHELL_CONFIG=""
if [ -n "$BASH_VERSION" ] || [ -f "$HOME/.bashrc" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
elif [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
elif [ -f "$HOME/.profile" ]; then
    SHELL_CONFIG="$HOME/.profile"
fi

PATH_LINE='export PATH="$HOME/.pphlx/bin:$PATH"'
if [ -n "$SHELL_CONFIG" ]; then
    if ! grep -q "$INSTALL_DIR" "$SHELL_CONFIG" 2>/dev/null; then
        echo "" >> "$SHELL_CONFIG"
        echo "$PATH_LINE" >> "$SHELL_CONFIG"
        echo "Added $INSTALL_DIR to $SHELL_CONFIG"
    fi
fi

echo "${CYAN}------------------------------------------------${CLEAR}"
echo "  ${GREEN}Success! PPHLX has been installed successfully.${CLEAR}"
echo "  Restart your terminal or run: ${CYAN}export PATH=\"\$HOME/.pphlx/bin:\$PATH\"${CLEAR}"
echo "  Verify by running: ${CYAN}pphlx --version${CLEAR}"
echo "${CYAN}------------------------------------------------${CLEAR}"
