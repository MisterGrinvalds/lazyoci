#!/bin/sh
# install.sh — Install lazyoci binary and shell completions.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/mistergrinvalds/lazyoci/main/install.sh | sh
#
# Options (via environment variables):
#   BINDIR    Install directory for the binary (default: /usr/local/bin)
#   VERSION   Specific version to install (default: latest release)
#
# Examples:
#   # Install latest to /usr/local/bin
#   curl -fsSL https://raw.githubusercontent.com/mistergrinvalds/lazyoci/main/install.sh | sh
#
#   # Install to ~/.local/bin (no sudo)
#   curl -fsSL https://raw.githubusercontent.com/mistergrinvalds/lazyoci/main/install.sh | BINDIR=~/.local/bin sh
#
#   # Install specific version
#   curl -fsSL https://raw.githubusercontent.com/mistergrinvalds/lazyoci/main/install.sh | VERSION=v0.1.0 sh

set -eu

REPO="mistergrinvalds/lazyoci"
BINARY="lazyoci"
BINDIR="${BINDIR:-/usr/local/bin}"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

info() {
    printf "\033[1;34m==>\033[0m %s\n" "$1"
}

warn() {
    printf "\033[1;33mwarning:\033[0m %s\n" "$1" >&2
}

error() {
    printf "\033[1;31merror:\033[0m %s\n" "$1" >&2
    exit 1
}

# ---------------------------------------------------------------------------
# Detect OS and architecture
# ---------------------------------------------------------------------------

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)  OS="linux" ;;
        darwin) OS="darwin" ;;
        *)      error "Unsupported OS: $OS. Use Windows? Try: scoop install lazyoci" ;;
    esac

    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        *)              error "Unsupported architecture: $ARCH" ;;
    esac

    info "Detected platform: ${OS}/${ARCH}"
}

# ---------------------------------------------------------------------------
# Resolve version
# ---------------------------------------------------------------------------

resolve_version() {
    if [ -n "${VERSION:-}" ]; then
        info "Using specified version: $VERSION"
        return
    fi

    info "Fetching latest release..."
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Set VERSION env var explicitly."
    fi

    info "Latest version: $VERSION"
}

# ---------------------------------------------------------------------------
# Download and install
# ---------------------------------------------------------------------------

download_and_install() {
    # Strip leading v for archive name (goreleaser uses version without v)
    VERSION_NUM="${VERSION#v}"
    ARCHIVE_NAME="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"

    TMPDIR=$(mktemp -d)
    trap 'rm -rf "$TMPDIR"' EXIT

    info "Downloading ${DOWNLOAD_URL}..."
    curl -fsSL "$DOWNLOAD_URL" -o "${TMPDIR}/archive.tar.gz" || \
        error "Download failed. Check that version ${VERSION} exists at https://github.com/${REPO}/releases"

    info "Extracting..."
    tar -xzf "${TMPDIR}/archive.tar.gz" -C "$TMPDIR"

    # Install binary
    install_binary

    # Install shell completions
    install_completions
}

install_binary() {
    info "Installing ${BINARY} to ${BINDIR}..."

    # Create directory if it doesn't exist
    if [ ! -d "$BINDIR" ]; then
        mkdir -p "$BINDIR" 2>/dev/null || sudo mkdir -p "$BINDIR"
    fi

    # Try without sudo first, then with sudo
    if [ -w "$BINDIR" ]; then
        install -m 755 "${TMPDIR}/${BINARY}" "${BINDIR}/${BINARY}"
    else
        info "Requesting sudo access to install to ${BINDIR}..."
        sudo install -m 755 "${TMPDIR}/${BINARY}" "${BINDIR}/${BINARY}"
    fi
}

install_completions() {
    # Only install if completion files exist in the archive
    if [ ! -d "${TMPDIR}/completions" ]; then
        return
    fi

    info "Installing shell completions..."

    # Bash completions
    if [ -f "${TMPDIR}/completions/lazyoci.bash" ]; then
        BASH_DIR=""
        if [ -d "/usr/local/share/bash-completion/completions" ]; then
            BASH_DIR="/usr/local/share/bash-completion/completions"
        elif [ -d "/etc/bash_completion.d" ]; then
            BASH_DIR="/etc/bash_completion.d"
        fi
        if [ -n "$BASH_DIR" ]; then
            if [ -w "$BASH_DIR" ]; then
                cp "${TMPDIR}/completions/lazyoci.bash" "${BASH_DIR}/${BINARY}"
            else
                sudo cp "${TMPDIR}/completions/lazyoci.bash" "${BASH_DIR}/${BINARY}" 2>/dev/null || true
            fi
        fi
    fi

    # Zsh completions
    if [ -f "${TMPDIR}/completions/_lazyoci" ]; then
        ZSH_DIR=""
        if [ -d "/usr/local/share/zsh/site-functions" ]; then
            ZSH_DIR="/usr/local/share/zsh/site-functions"
        fi
        if [ -n "$ZSH_DIR" ]; then
            if [ -w "$ZSH_DIR" ]; then
                cp "${TMPDIR}/completions/_lazyoci" "${ZSH_DIR}/_${BINARY}"
            else
                sudo cp "${TMPDIR}/completions/_lazyoci" "${ZSH_DIR}/_${BINARY}" 2>/dev/null || true
            fi
        fi
    fi

    # Fish completions
    if [ -f "${TMPDIR}/completions/lazyoci.fish" ]; then
        FISH_DIR="${HOME}/.config/fish/completions"
        if command -v fish >/dev/null 2>&1; then
            mkdir -p "$FISH_DIR" 2>/dev/null || true
            cp "${TMPDIR}/completions/lazyoci.fish" "${FISH_DIR}/${BINARY}.fish" 2>/dev/null || true
        fi
    fi
}

# ---------------------------------------------------------------------------
# Verify and print results
# ---------------------------------------------------------------------------

verify_install() {
    if command -v "$BINARY" >/dev/null 2>&1; then
        INSTALLED_VERSION=$("$BINARY" version 2>/dev/null | head -1 || echo "unknown")
        printf "\n"
        info "Successfully installed!"
        printf "  %s\n" "$INSTALLED_VERSION"
        printf "  Binary: %s\n" "${BINDIR}/${BINARY}"
    else
        printf "\n"
        warn "${BINARY} was installed to ${BINDIR}/${BINARY} but is not in your PATH."
        printf "\n"
        printf "  Add this to your shell profile:\n"
        printf "    export PATH=\"%s:\$PATH\"\n" "$BINDIR"
    fi

    printf "\n"
    printf "  Get started:\n"
    printf "    %s              # Launch TUI\n" "$BINARY"
    printf "    %s --help       # See all commands\n" "$BINARY"
    printf "\n"

    # Shell completion hints for go install users
    if [ ! -d "${TMPDIR:-}/completions" ] 2>/dev/null; then
        printf "  Shell completions (add to your shell profile):\n"
        printf "    %s completion bash  # Bash\n" "$BINARY"
        printf "    %s completion zsh   # Zsh\n" "$BINARY"
        printf "    %s completion fish  # Fish\n" "$BINARY"
        printf "\n"
    fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

main() {
    printf "\n"
    info "Installing lazyoci..."
    printf "\n"

    detect_platform
    resolve_version
    download_and_install
    verify_install
}

main
