#!/bin/sh
# PPHLX Universal Toolchain Installer for macOS and Linux
# Installs the native Go compiler binary to /usr/local/bin/PPHLX

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
VERSION="v1.0.0" # Hardcoded specific release version for dev phase

if [ "$VERSION" = "latest" ]; then
    # Query GitHub API to get the latest tag
    TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -o '\"tag_name\": \"[^\"]*\"' | head -n 1 | cut -d '"' -f 4)
    if [ -z "$TAG" ]; then
        # Fallback if API rate limited
        TAG="v1.0.0"
    fi
else
    TAG="$VERSION"
fi

# Map target filename based on OS and ARCH
# Release files structure: PPHLX-[os]-[arch].tar.gz or zip
BINARY_NAME="PPHLX-$OS-$ARCH"
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

TAR_PATH="$TMP_DIR/PPHLX.tar.gz"

# Fetch binary
curl -fsSL "$DOWNLOAD_URL" -o "$TAR_PATH" || {
    echo "${RED}Error: Failed to download the PPHLX binary archive. It might not be released yet for this platform.${CLEAR}"
    exit 1
}

# Extract binary
tar -xzf "$TAR_PATH" -C "$TMP_DIR"

# 5. Move binary to system path
DEST="/usr/local/bin/PPHLX"
echo "Installing binary to: $DEST"

if [ -w "/usr/local/bin" ]; then
    mv "$TMP_DIR/PPHLX" "$DEST"
    chmod +x "$DEST"
else
    echo "${CYAN}Note: Write access to /usr/local/bin requires administrator privileges. Running with sudo...${CLEAR}"
    sudo mv "$TMP_DIR/PPHLX" "$DEST"
    sudo chmod +x "$DEST"
fi

echo "${CYAN}------------------------------------------------${CLEAR}"
echo "  ${GREEN}Success! PPHLX has been installed successfully.${CLEAR}"
echo "  Verify by running: ${CYAN}PPHLX --version${CLEAR}"
echo "${CYAN}------------------------------------------------${CLEAR}"

