#!/usr/bin/env bash
# Hefesto installer — https://github.com/Edcko/Hefesto
# Usage: curl -fsSL https://raw.githubusercontent.com/Edcko/Hefesto/main/install.sh | bash
#
# Install the latest version of Hefesto CLI.
# Designed for macOS, Linux, and Android/Termux.

set -euo pipefail

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------
REPO="Edcko/Hefesto"
BINARY_NAME="hefesto"
GITHUB_RELEASE_BASE="https://github.com/${REPO}/releases/latest/download"
DEFAULT_INSTALL_DIR="${HOME}/.local/bin"
FALLBACK_INSTALL_DIR="/usr/local/bin"

# ---------------------------------------------------------------------------
# Colors
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
info()    { printf "${BLUE}[INFO]${NC}    %s\n" "$*"; }
success() { printf "${GREEN}[OK]${NC}      %s\n" "$*"; }
warn()    { printf "${YELLOW}[WARN]${NC}    %s\n" "$*"; }
error()   { printf "${RED}[ERROR]${NC}   %s\n" "$*" >&2; }
banner()  { printf "\n${BOLD}${BLUE}%s${NC}\n" "$*"; }

cleanup() {
  if [[ -n "${TMP_FILE:-}" && -f "${TMP_FILE}" ]]; then
    rm -f "${TMP_FILE}"
  fi
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Pre-flight checks
# ---------------------------------------------------------------------------
check_dependencies() {
  if ! command -v curl &>/dev/null; then
    error "curl is required but not found in PATH."
    error "Install it with your package manager and try again."
    exit 1
  fi
}

check_connectivity() {
  if ! curl -fsSL --max-time 10 "https://github.com" >/dev/null 2>&1; then
    error "Cannot reach GitHub. Check your internet connection and try again."
    exit 1
  fi
}

# ---------------------------------------------------------------------------
# OS & Architecture detection
# ---------------------------------------------------------------------------
detect_os() {
  local kernel
  kernel="$(uname -s | tr '[:upper:]' '[:lower:]')"

  case "${kernel}" in
    darwin)
      echo "darwin"
      ;;
    linux)
      # Detect Android/Termux
      if [[ -n "${TERMUX_VERSION:-}" ]] || [[ "$(uname -o 2>/dev/null)" == "Android" ]]; then
        echo "android"
      else
        echo "linux"
      fi
      ;;
    *)
      error "Unsupported operating system: ${kernel}"
      error "Hefesto is available for macOS, Linux, and Android/Termux."
      exit 1
      ;;
  esac
}

detect_arch() {
  local machine
  machine="$(uname -m)"

  case "${machine}" in
    x86_64|amd64)
      echo "amd64"
      ;;
    aarch64|arm64)
      echo "arm64"
      ;;
    *)
      error "Unsupported architecture: ${machine}"
      error "Hefesto is available for amd64 and arm64."
      exit 1
      ;;
  esac
}

# ---------------------------------------------------------------------------
# Installation
# ---------------------------------------------------------------------------
determine_install_dir() {
  # Prefer user-local directory (no sudo needed)
  if [[ -d "${DEFAULT_INSTALL_DIR}" ]] || mkdir -p "${DEFAULT_INSTALL_DIR}" 2>/dev/null; then
    echo "${DEFAULT_INSTALL_DIR}"
    return
  fi

  # Fallback to system directory if user has write access or sudo
  if [[ -w "${FALLBACK_INSTALL_DIR}" ]]; then
    echo "${FALLBACK_INSTALL_DIR}"
    return
  fi

  if command -v sudo &>/dev/null; then
    warn "Cannot write to ${DEFAULT_INSTALL_DIR} or ${FALLBACK_INSTALL_DIR}."
    warn "Will attempt to use sudo for ${FALLBACK_INSTALL_DIR}."
    echo "${FALLBACK_INSTALL_DIR}"
    return
  fi

  error "No suitable installation directory found."
  error "Ensure ${DEFAULT_INSTALL_DIR} is writable or run with sudo."
  exit 1
}

needs_sudo() {
  local dir="$1"
  [[ ! -w "${dir}" ]]
}

download_binary() {
  local os="$1"
  local arch="$2"
  local output="$3"

  local remote_binary="${BINARY_NAME}-${os}-${arch}"
  local url="${GITHUB_RELEASE_BASE}/${remote_binary}"

  info "Downloading ${remote_binary}..."

  # Use a temporary file for the download
  TMP_FILE="$(mktemp "${output}.XXXXXX")"

  local http_code
  http_code=$(curl -fsSL -w "%{http_code}" -o "${TMP_FILE}" "${url}" 2>/dev/null) || {
    error "Download failed. Could not fetch ${url}"
    exit 1
  }

  if [[ "${http_code}" != "200" ]]; then
    error "Download failed with HTTP ${http_code}."
    error "Binary '${remote_binary}' may not be available for your platform."
    error "Check available releases at: https://github.com/${REPO}/releases"
    exit 1
  fi

  # Move temp file to final destination
  mv -f "${TMP_FILE}" "${output}"
  TMP_FILE=""
}

make_executable() {
  local target="$1"
  if needs_sudo "$(dirname "${target}")"; then
    sudo chmod +x "${target}"
  else
    chmod +x "${target}"
  fi
}

# ---------------------------------------------------------------------------
# PATH configuration
# ---------------------------------------------------------------------------
configure_path() {
  local install_dir="$1"
  local shell_rc=""
  local shell_name=""

  # Detect current shell
  if [[ -n "${ZSH_VERSION:-}" ]] || [[ "${SHELL:-}" == */zsh ]]; then
    shell_rc="${HOME}/.zshrc"
    shell_name="zsh"
  elif [[ -n "${BASH_VERSION:-}" ]] || [[ "${SHELL:-}" == */bash ]]; then
    shell_rc="${HOME}/.bashrc"
    shell_name="bash"
  fi

  # Also check for Android/Termux
  if [[ -n "${TERMUX_VERSION:-}" ]] && [[ -f "${HOME}/.bashrc" ]]; then
    shell_rc="${HOME}/.bashrc"
    shell_name="bash (Termux)"
  fi

  if [[ -z "${shell_rc}" ]]; then
    warn "Could not detect shell config file."
    info "Add this line to your shell profile manually:"
    printf "  export PATH=\"${install_dir}:\$PATH\"\n"
    return
  fi

  # Check if PATH is already configured
  local path_entry="export PATH=\"${install_dir}:\$PATH\""
  if grep -qF "${install_dir}" "${shell_rc}" 2>/dev/null; then
    info "PATH already configured in ${shell_rc}"
    return
  fi

  # Add to shell config
  printf "\n# Added by Hefesto installer\n%s\n" "${path_entry}" >> "${shell_rc}"
  success "Added ${install_dir} to PATH in ${shell_rc}"
  info "Run 'source ${shell_rc}' or open a new terminal to update your PATH."
}

# ---------------------------------------------------------------------------
# Verification
# ---------------------------------------------------------------------------
verify_installation() {
  local target="$1"

  if "${target}" version &>/dev/null; then
    local version
    version="$("${target}" version 2>/dev/null | head -1)"
    success "Hefesto installed successfully: ${version}"
  else
    # Binary exists but 'version' might not be the right subcommand
    # Try just running the binary
    if [[ -x "${target}" ]]; then
      success "Hefesto binary installed at ${target}"
      info "Run '${BINARY_NAME} --help' to get started."
    else
      warn "Binary installed at ${target} but may not be executable."
    fi
  fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
  banner "Hefesto Installer"
  printf "${BOLD}https://github.com/${REPO}${NC}\n\n"

  # Step 1: Pre-flight
  info "Checking dependencies..."
  check_dependencies

  info "Checking connectivity..."
  check_connectivity
  success "Pre-flight checks passed"

  # Step 2: Detect platform
  info "Detecting platform..."
  OS="$(detect_os)"
  ARCH="$(detect_arch)"
  success "Platform: ${OS}/${ARCH}"

  # Step 3: Determine install location
  INSTALL_DIR="$(determine_install_dir)"
  TARGET="${INSTALL_DIR}/${BINARY_NAME}"
  info "Install target: ${TARGET}"

  # Step 4: Download
  download_binary "${OS}" "${ARCH}" "${TARGET}"
  success "Download complete"

  # Step 5: Make executable
  make_executable "${TARGET}"
  success "Binary made executable"

  # Step 6: Verify
  verify_installation "${TARGET}"

  # Step 7: Configure PATH
  info "Configuring PATH..."
  configure_path "${INSTALL_DIR}"

  # Done
  printf "\n${GREEN}${BOLD}✓ Installation complete!${NC}\n"
  info "Run '${BINARY_NAME} --help' to get started."
  printf "\n"

  # Warn if install dir is not in current PATH
  if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* ]]; then
    warn "${INSTALL_DIR} is not in your current PATH."
    info "Open a new terminal or run: source ~/.${shell_name:-bash}rc"
  fi
}

main "$@"
