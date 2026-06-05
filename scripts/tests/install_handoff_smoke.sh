#!/usr/bin/env bash
# T015: Smoke test — installer handoff behavior
# Usage: bash scripts/tests/install_handoff_smoke.sh
# Returns exit 0 on pass, 1 on failure.

INSTALL_SH="$(cd "$(dirname "$0")/../.." && pwd)/install.sh"

if [[ ! -f "$INSTALL_SH" ]]; then
  printf 'FAIL: install.sh not found at %s\n' "$INSTALL_SH" >&2
  exit 1
fi

PASS=0
FAIL=0

pass() { printf 'PASS: %s\n' "$1"; PASS=$((PASS+1)); }
fail() { printf 'FAIL: %s\n' "$1"; FAIL=$((FAIL+1)); }

make_runtime_bundle() {
  local tmpdir="$1"
  local root="$tmpdir/runtime"
  mkdir -p "$root/stacks/20i"
  printf 'services: {}\n' > "$root/stacks/20i/docker-compose.shared.yml"
  printf 'services: {}\n' > "$root/stacks/20i/docker-compose.20i.yml"
  tar -czf "$tmpdir/runtime.tar.gz" -C "$root" stacks
  printf '%s\n' "$tmpdir/runtime.tar.gz"
}

contains_all() {
  local haystack="$1"
  shift
  local needle
  for needle in "$@"; do
    if ! echo "$haystack" | grep -q "$needle"; then
      return 1
    fi
  done
}

# Test 1: Interactive TTY path hands off to bare `stage`.
if command -v script >/dev/null 2>&1; then
  tmpdir=$(mktemp -d)
  out=$(env \
    STAGESERVE_INSTALL_DIR="$tmpdir" \
    STAGESERVE_STACK_HOME="$tmpdir/stack-home" \
    STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
    STAGESERVE_TEST_BUNDLE_PATH="$(make_runtime_bundle "$tmpdir")" \
    script -q /dev/null bash "$INSTALL_SH" --test-mode 2>&1) && true
  rm -rf "$tmpdir"
  clean_out=$(printf '%s' "$out" | tr -d '\r')
  if contains_all "$clean_out" "To continue with guided setup, run:" "stage" && \
    ! echo "$clean_out" | grep -q "stage setup --tui"; then
    pass "tty_handoff"
  else
    fail "tty_handoff"
    echo "Output was: $out"
  fi
fi

# Test 2: NONINTERACTIVE=1 path prints the explicit direct-command path.
tmpdir=$(mktemp -d)
out=$(NONINTERACTIVE=1 \
  STAGESERVE_INSTALL_DIR="$tmpdir" \
  STAGESERVE_STACK_HOME="$tmpdir/stack-home" \
  STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
  STAGESERVE_TEST_BUNDLE_PATH="$(make_runtime_bundle "$tmpdir")" \
  bash "$INSTALL_SH" --test-mode 2>&1) && true
rm -rf "$tmpdir"
if contains_all "$out" "stage setup" "stage init" "stage up" "stage doctor"; then
  pass "noninteractive_handoff"
else
  fail "noninteractive_handoff"
  echo "Output was: $out"
fi

# Test 3: Non-TTY path prints the same explicit direct-command path.
tmpdir=$(mktemp -d)
out=$(STAGESERVE_INSTALL_DIR="$tmpdir" \
  STAGESERVE_STACK_HOME="$tmpdir/stack-home" \
  STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
  STAGESERVE_TEST_BUNDLE_PATH="$(make_runtime_bundle "$tmpdir")" \
  bash "$INSTALL_SH" --test-mode --no-tty 2>&1) && true
rm -rf "$tmpdir"
if contains_all "$out" "stage setup" "stage init" "stage up" "stage doctor"; then
  pass "nontty_handoff"
else
  fail "nontty_handoff"
  echo "Output was: $out"
fi

printf '\nResults: %d passed, %d failed\n' "$PASS" "$FAIL"
[[ $FAIL -eq 0 ]]
