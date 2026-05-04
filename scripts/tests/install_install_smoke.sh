#!/usr/bin/env bash
# Smoke tests — installer binary placement, permissions, dir creation, PATH warning.
# Usage: bash scripts/tests/install_install_smoke.sh
# Returns exit 0 on pass, 1 on any failure.

INSTALL_SH="$(cd "$(dirname "$0")/../.." && pwd)/install.sh"

if [[ ! -f "$INSTALL_SH" ]]; then
  printf 'FAIL: install.sh not found at %s\n' "$INSTALL_SH" >&2
  exit 1
fi

PASS=0
FAIL=0

pass() { printf 'PASS: %s\n' "$1"; PASS=$((PASS+1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL+1)); }

# ── Test helpers ──────────────────────────────────────────────────────────────

# Run installer with a throw-away dir and a dummy binary.
# Usage: run_installer <extra-env-pairs...>
run_installer() {
  local tmpdir
  tmpdir=$(mktemp -d)
  out=$(env STAGESERVE_INSTALL_DIR="$tmpdir" \
             STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
             NONINTERACTIVE=1 \
             "$@" \
             bash "$INSTALL_SH" --test-mode 2>&1) && rc=0 || rc=$?
  printf '%s' "$tmpdir"          # caller captures tmpdir via substitution
}

# ── Test 1: Install dir is created when it does not exist ─────────────────────
tmpdir=$(mktemp -d)
install_target="$tmpdir/nested/install"
out=$(STAGESERVE_INSTALL_DIR="$install_target" \
      STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
      NONINTERACTIVE=1 \
      bash "$INSTALL_SH" --test-mode 2>&1) && rc=0 || rc=$?
if [[ $rc -eq 0 && -d "$install_target" ]]; then
  pass "install_creates_missing_dir"
else
  fail "install_creates_missing_dir" "exit=$rc, dir exists=$(test -d "$install_target" && echo yes || echo no)"
fi
rm -rf "$tmpdir"

# ── Test 2: Binary is placed at STAGESERVE_INSTALL_DIR/stage ───────────────
tmpdir=$(mktemp -d)
out=$(STAGESERVE_INSTALL_DIR="$tmpdir" \
      STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
      NONINTERACTIVE=1 \
      bash "$INSTALL_SH" --test-mode 2>&1) && rc=0 || rc=$?
if [[ $rc -eq 0 && -f "$tmpdir/stage" ]]; then
  pass "install_places_binary"
else
  fail "install_places_binary" "exit=$rc, file exists=$(test -f "$tmpdir/stage" && echo yes || echo no)"
fi
rm -rf "$tmpdir"

# ── Test 3: Installed binary has execute permission ───────────────────────────
tmpdir=$(mktemp -d)
out=$(STAGESERVE_INSTALL_DIR="$tmpdir" \
      STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
      NONINTERACTIVE=1 \
      bash "$INSTALL_SH" --test-mode 2>&1) && rc=0 || rc=$?
if [[ $rc -eq 0 && -x "$tmpdir/stage" ]]; then
  pass "install_binary_executable"
else
  fail "install_binary_executable" "exit=$rc, executable=$(test -x "$tmpdir/stage" && echo yes || echo no)"
fi
rm -rf "$tmpdir"

# ── Test 4: STAGESERVE_INSTALL_DIR override is respected ───────────────────────
tmpdir=$(mktemp -d)
custom="$tmpdir/custom-bin"
mkdir -p "$custom"
out=$(STAGESERVE_INSTALL_DIR="$custom" \
      STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
      NONINTERACTIVE=1 \
      bash "$INSTALL_SH" --test-mode 2>&1) && rc=0 || rc=$?
if [[ $rc -eq 0 && -f "$custom/stage" ]]; then
  pass "install_custom_dir_respected"
else
  fail "install_custom_dir_respected" "exit=$rc, file=$(ls "$custom" 2>/dev/null || echo empty)"
fi
rm -rf "$tmpdir"

# ── Test 5: Installer exits 0 on successful install ──────────────────────────
tmpdir=$(mktemp -d)
out=$(STAGESERVE_INSTALL_DIR="$tmpdir" \
      STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
      NONINTERACTIVE=1 \
      bash "$INSTALL_SH" --test-mode 2>&1) && rc=0 || rc=$?
if [[ $rc -eq 0 ]]; then
  pass "install_exit_zero_on_success"
else
  fail "install_exit_zero_on_success" "exit=$rc; output: $out"
fi
rm -rf "$tmpdir"

# ── Test 6: Installer is idempotent — running twice succeeds ─────────────────
tmpdir=$(mktemp -d)
out1=$(STAGESERVE_INSTALL_DIR="$tmpdir" \
       STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
       NONINTERACTIVE=1 \
       bash "$INSTALL_SH" --test-mode 2>&1) && rc1=0 || rc1=$?
out2=$(STAGESERVE_INSTALL_DIR="$tmpdir" \
       STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
       NONINTERACTIVE=1 \
       bash "$INSTALL_SH" --test-mode 2>&1) && rc2=0 || rc2=$?
if [[ $rc1 -eq 0 && $rc2 -eq 0 && -x "$tmpdir/stage" ]]; then
  pass "install_idempotent"
else
  fail "install_idempotent" "first=$rc1, second=$rc2"
fi
rm -rf "$tmpdir"

# ── Test 7: PATH warning emitted when install dir not in PATH ─────────────────
tmpdir=$(mktemp -d)
# Use a dir that is definitely not in PATH
unique_dir="$tmpdir/zz_not_in_path_$$"
mkdir -p "$unique_dir"
out=$(STAGESERVE_INSTALL_DIR="$unique_dir" \
      STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
      NONINTERACTIVE=1 \
      PATH="/usr/bin:/bin" \
      bash "$INSTALL_SH" --test-mode 2>&1) && true
if echo "$out" | grep -q "not in your PATH"; then
  pass "install_path_warning_emitted"
else
  fail "install_path_warning_emitted" "Output was: $out"
fi
rm -rf "$tmpdir"

# ── Test 8: No PATH warning when install dir is already in PATH ───────────────
tmpdir=$(mktemp -d)
out=$(STAGESERVE_INSTALL_DIR="$tmpdir" \
      STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
      NONINTERACTIVE=1 \
      PATH="$tmpdir:/usr/bin:/bin" \
      bash "$INSTALL_SH" --test-mode 2>&1) && true
if echo "$out" | grep -q "not in your PATH"; then
  fail "install_no_path_warning_when_in_path" "PATH warning appeared even though dir is in PATH"
else
  pass "install_no_path_warning_when_in_path"
fi
rm -rf "$tmpdir"

# ── Test 9: NONINTERACTIVE=1 suppresses TUI launch prompt ────────────────────
tmpdir=$(mktemp -d)
out=$(STAGESERVE_INSTALL_DIR="$tmpdir" \
      STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
      NONINTERACTIVE=1 \
      bash "$INSTALL_SH" --test-mode 2>&1) && true
rm -rf "$tmpdir"
if echo "$out" | grep -qi "launching first-run"; then
  fail "install_noninteractive_no_tui_launch" "Should not show TUI launch in NONINTERACTIVE mode"
else
  pass "install_noninteractive_no_tui_launch"
fi

# ── Test 10: STAGESERVE_VERSION env var overrides default version in asset name ─
orig=$(STAGESERVE_TEST_ONLY_ASSET_NAME=1 STAGESERVE_VERSION="v1.2.3" bash "$INSTALL_SH" 2>&1) && true
if [[ "$orig" == *"v1.2.3"* ]]; then
  pass "install_version_env_respected_in_asset_name"
else
  fail "install_version_env_respected_in_asset_name" "Got: $orig"
fi

# ── Results ───────────────────────────────────────────────────────────────────
printf '\nResults: %d passed, %d failed\n' "$PASS" "$FAIL"
[[ $FAIL -eq 0 ]]
