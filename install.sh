#!/usr/bin/env bash
# StageServe installer — downloads the correct release asset, verifies checksum,
# and places the binary in the install destination.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/peternicholls/StageServe/master/install.sh | bash
#   NONINTERACTIVE=1 bash install.sh
#
# Environment overrides (for testing):
#   STAGESERVE_INSTALL_DIR   — destination directory (default: $HOME/.local/bin)
#   STAGESERVE_TEST_ASSET_PATH — bypass download; copy this path as the binary
#   NONINTERACTIVE          — suppress prompts and TUI handoff (set to 1)
#
# Test-mode flags (internal — used by smoke tests):
#   --test-mode             — enable test-harness hooks
#   --no-tty                — simulate non-TTY environment
#   --verify-sha            — run only checksum verification, then exit
set -euo pipefail

# ──────────────────────────────────────────────────────────────────────────────
# Globals / defaults
# ──────────────────────────────────────────────────────────────────────────────
STAGESERVE_VERSION="${STAGESERVE_VERSION:-latest}"
STAGESERVE_REPO="peternicholls/StageServe"
STAGESERVE_INSTALL_DIR="${STAGESERVE_INSTALL_DIR:-$HOME/.local/bin}"
NONINTERACTIVE="${NONINTERACTIVE:-0}"

_test_mode=0
_no_tty=0
_verify_sha_only=0

# ──────────────────────────────────────────────────────────────────────────────
# Argument parsing
# ──────────────────────────────────────────────────────────────────────────────
for arg in "$@"; do
  case "$arg" in
    --test-mode)    _test_mode=1 ;;
    --no-tty)       _no_tty=1 ;;
    --verify-sha)   _verify_sha_only=1 ;;
  esac
done

# ──────────────────────────────────────────────────────────────────────────────
# Helpers
# ──────────────────────────────────────────────────────────────────────────────
info()  { printf '\033[0;34m==>\033[0m %s\n' "$*"; }
ok()    { printf '\033[0;32m ✓\033[0m %s\n' "$*"; }
warn()  { printf '\033[0;33m !\033[0m %s\n' "$*"; }
die()   { printf '\033[0;31mERROR:\033[0m %s\n' "$*" >&2; exit 1; }

# ──────────────────────────────────────────────────────────────────────────────
# T017: OS/arch detection and asset naming
# ──────────────────────────────────────────────────────────────────────────────
detect_os() {
  case "$(uname -s)" in
    Darwin) echo "Darwin" ;;
    Linux)  echo "Linux" ;;
    *)      die "Unsupported OS: $(uname -s). Only macOS and Linux are supported." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)   echo "x86_64" ;;
    arm64|aarch64)  echo "arm64" ;;
    *)              die "Unsupported architecture: $(uname -m)." ;;
  esac
}

asset_name() {
  local version="$1"
  echo "stage_${version}_$(detect_os)_$(detect_arch)"
}

# ──────────────────────────────────────────────────────────────────────────────
# Test-mode: expose asset_name and exit (used by smoke tests)
# ──────────────────────────────────────────────────────────────────────────────
if [[ "${STAGESERVE_TEST_ONLY_ASSET_NAME:-0}" == "1" ]]; then
  asset_name "${STAGESERVE_VERSION}"
  exit 0
fi

# ──────────────────────────────────────────────────────────────────────────────
# T018: Checksum fetch and verification
# ──────────────────────────────────────────────────────────────────────────────
verify_sha256() {
  local file="$1"
  local shafile="$2"

  if ! command -v shasum &>/dev/null && ! command -v sha256sum &>/dev/null; then
    warn "No SHA-256 tool found (shasum or sha256sum); skipping integrity check."
    return 0
  fi

  local expected actual
  expected=$(awk '{print $1}' "$shafile")
  if command -v shasum &>/dev/null; then
    actual=$(shasum -a 256 "$file" | awk '{print $1}')
  else
    actual=$(sha256sum "$file" | awk '{print $1}')
  fi

  if [[ "$actual" == "$expected" ]]; then
    ok "checksum ok ($actual)"
  else
    die "Checksum mismatch.\n  expected: $expected\n  actual:   $actual"
  fi
}

# ──────────────────────────────────────────────────────────────────────────────
# Test-mode: verify-sha only (used by T016 smoke tests)
# ──────────────────────────────────────────────────────────────────────────────
if [[ $_verify_sha_only -eq 1 ]]; then
  if [[ -z "${STAGESERVE_TEST_VERIFY_SHA:-}" ]] || [[ -z "${STAGESERVE_TEST_SHA_FILE:-}" ]]; then
    die "--verify-sha requires STAGESERVE_TEST_VERIFY_SHA and STAGESERVE_TEST_SHA_FILE"
  fi
  verify_sha256 "$STAGESERVE_TEST_VERIFY_SHA" "$STAGESERVE_TEST_SHA_FILE"
  exit 0
fi

# ──────────────────────────────────────────────────────────────────────────────
# T019: Deterministic install destination and PATH warning
# ──────────────────────────────────────────────────────────────────────────────
ensure_install_dir() {
  if [[ ! -d "$STAGESERVE_INSTALL_DIR" ]]; then
    info "Creating install directory: $STAGESERVE_INSTALL_DIR"
    mkdir -p "$STAGESERVE_INSTALL_DIR"
  fi
}

check_path_warning() {
  if [[ ":$PATH:" != *":$STAGESERVE_INSTALL_DIR:"* ]]; then
    warn "$STAGESERVE_INSTALL_DIR is not in your PATH."
    warn "Add the following to your shell profile:"
    warn "  export PATH=\"\$PATH:$STAGESERVE_INSTALL_DIR\""
  fi
}

# ──────────────────────────────────────────────────────────────────────────────
# Download or use test-supplied asset
# ──────────────────────────────────────────────────────────────────────────────
resolve_version() {
  if [[ "$STAGESERVE_VERSION" == "latest" ]]; then
    local resolved_version=""
    if [[ "${STAGESERVE_TEST_DISABLE_RELEASE_LOOKUP:-0}" != "1" ]] && command -v curl &>/dev/null; then
      resolved_version=$(curl -fsSL \
        "https://api.github.com/repos/$STAGESERVE_REPO/releases/latest" \
        | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    fi
    # Fallback if curl/API unavailable (e.g., offline test env)
    STAGESERVE_VERSION="${resolved_version:-dev}"
  fi
}

if [[ "${STAGESERVE_TEST_ONLY_RESOLVED_VERSION:-0}" == "1" ]]; then
  resolve_version
  printf '%s\n' "$STAGESERVE_VERSION"
  exit 0
fi

download_binary() {
  local dest="$1"

  # Test mode: copy provided asset instead of downloading.
  if [[ -n "${STAGESERVE_TEST_ASSET_PATH:-}" ]]; then
    cp "$STAGESERVE_TEST_ASSET_PATH" "$dest"
    chmod +x "$dest"
    return 0
  fi

  resolve_version

  local name
  name=$(asset_name "$STAGESERVE_VERSION")
  local url="https://github.com/$STAGESERVE_REPO/releases/download/$STAGESERVE_VERSION/$name"
  local sha_url="$url.sha256"

  info "Downloading $name ..."
  local tmpdir
  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' RETURN

  local tmpbin="$tmpdir/$name"
  local tmpsha="$tmpdir/$name.sha256"

  curl -fsSL "$url"     -o "$tmpbin"
  curl -fsSL "$sha_url" -o "$tmpsha"

  verify_sha256 "$tmpbin" "$tmpsha"

  cp "$tmpbin" "$dest"
  chmod +x "$dest"
}

# ──────────────────────────────────────────────────────────────────────────────
# T020: Interactive handoff / NONINTERACTIVE next-step behavior
# ──────────────────────────────────────────────────────────────────────────────
print_next_steps() {
  local is_tty
  if [[ $_no_tty -eq 1 ]]; then
    is_tty=0
  elif [[ -t 1 ]]; then
    is_tty=1
  else
    is_tty=0
  fi

  echo ""
  ok "StageServe installed to $STAGESERVE_INSTALL_DIR/stage"
  echo ""

  if [[ "$NONINTERACTIVE" == "1" ]]; then
    info "Next steps:"
    echo "  stage setup    — run machine-readiness checks and first-run setup"
    echo "  stage doctor   — diagnose machine drift at any time"
  elif [[ $is_tty -eq 1 ]]; then
    info "To complete setup, run:"
    echo "  stage setup --tui"
  else
    info "Next steps:"
    echo "  stage setup    — run machine-readiness checks and first-run setup"
    echo "  stage doctor   — diagnose machine drift at any time"
  fi
}

# ──────────────────────────────────────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────────────────────────────────────
main() {
  info "Installing StageServe ..."

  ensure_install_dir
  local dest="$STAGESERVE_INSTALL_DIR/stage"
  download_binary "$dest"
  check_path_warning
  print_next_steps
}

main "$@"
