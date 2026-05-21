#!/bin/bash
set -e

REPO="plutobe/cr-cli"
BINARY="cr-cli"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}$*${NC}"; }
warn()  { echo -e "${YELLOW}$*${NC}"; }
error() { echo -e "${RED}$*${NC}" >&2; exit 1; }

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux*)   OS="linux" ;;
        Darwin*)  OS="darwin" ;;
        MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
        *) error "Unsupported OS: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)   ARCH="arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Download binary
download() {
    local url="$1"
    local dest="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$dest"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$dest" "$url"
    else
        error "Neither curl nor wget found. Please install one."
    fi
}

# Get latest release version
get_latest_version() {
    local url="https://api.github.com/repos/${REPO}/releases/latest"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$url" | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
    fi
}

main() {
    detect_os
    detect_arch

    info "Detecting system: ${OS}/${ARCH}"

    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        error "Failed to get latest version. Check your network."
    fi
    info "Latest version: ${VERSION}"

    if [ "$OS" = "windows" ]; then
        FILENAME="${BINARY}-${OS}-${ARCH}.exe"
    else
        FILENAME="${BINARY}-${OS}-${ARCH}"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
    TMP_FILE="/tmp/${FILENAME}"

    info "Downloading ${DOWNLOAD_URL} ..."
    download "$DOWNLOAD_URL" "$TMP_FILE"

    if [ "$OS" = "windows" ]; then
        chmod +x "$TMP_FILE"
        info "Downloaded to: ${TMP_FILE}"
        info "Please move it to a directory in your PATH."
    else
        info "Installing to ${INSTALL_DIR}/${BINARY} ..."
        chmod +x "$TMP_FILE"
        mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY}"
        info "Installed successfully!"
    fi

    echo ""
    info "Run 'cr-cli review' in any git project to get started."
}

main "$@"
