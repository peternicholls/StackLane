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

# Test 1: NONINTERACTIVE=1 path prints next-step guidance containing "stacklane setup".
tmpdir=$(mktemp -d)
out=$(NONINTERACTIVE=1 \
  STACKLANE_INSTALL_DIR="$tmpdir" \
  STACKLANE_TEST_ASSET_PATH="/usr/bin/true" \
  bash "$INSTALL_SH" --test-mode 2>&1) && true
rm -rf "$tmpdir"
if echo "$out" | grep -q "stacklane setup"; then
  pass "noninteractive_handoff"
else
  fail "noninteractive_handoff"
  echo "Output was: $out"
fi

# Test 2: Non-TTY path prints next-step guidance.
tmpdir=$(mktemp -d)
out=$(STACKLANE_INSTALL_DIR="$tmpdir" \
  STACKLANE_TEST_ASSET_PATH="/usr/bin/true" \
  bash "$INSTALL_SH" --test-mode --no-tty 2>&1) && true
rm -rf "$tmpdir"
if echo "$out" | grep -qE "stacklane setup|stacklane setup --tui"; then
  pass "nontty_handoff"
else
  fail "nontty_handoff"
  echo "Output was: $out"
fi

printf '\nResults: %d passed, %d failed\n' "$PASS" "$FAIL"
[[ $FAIL -eq 0 ]]
