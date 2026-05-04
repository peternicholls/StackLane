#!/usr/bin/env bash
# T016: Smoke test — installer asset naming and checksum verification.
# Usage: bash scripts/tests/install_checksum_smoke.sh
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

# Test 1: Asset name format matches expected pattern.
name=$(STAGESERVE_TEST_ONLY_ASSET_NAME=1 bash "$INSTALL_SH" --test-mode 2>&1) && true
printf 'Asset name: %s\n' "$name"
if [[ "$name" =~ ^stage_[^_]+_(Darwin|Linux)_(x86_64|arm64)$ ]]; then
  pass "asset_name_format"
else
  fail "asset_name_format"
fi

# Test 2: SHA-256 verification passes on a known-good file.
tmpdir=$(mktemp -d)
testfile="$tmpdir/payload"
echo "hello stage" > "$testfile"
expected_sha=$(shasum -a 256 "$testfile" | awk '{print $1}')
shafile="$tmpdir/payload.sha256"
echo "$expected_sha  payload" > "$shafile"
out=$(STAGESERVE_TEST_VERIFY_SHA="$testfile" \
  STAGESERVE_TEST_SHA_FILE="$shafile" \
  bash "$INSTALL_SH" --test-mode --verify-sha 2>&1) && true
rm -rf "$tmpdir"
if echo "$out" | grep -q "checksum ok"; then
  pass "sha256_verify_pass"
else
  fail "sha256_verify_pass"
  echo "Output was: $out"
fi

# Test 3: SHA-256 verification fails on a tampered file.
tmpdir=$(mktemp -d)
testfile="$tmpdir/payload"
echo "hello stage" > "$testfile"
shafile="$tmpdir/payload.sha256"
echo "0000000000000000000000000000000000000000000000000000000000000000  payload" > "$shafile"
out=$(STAGESERVE_TEST_VERIFY_SHA="$testfile" \
  STAGESERVE_TEST_SHA_FILE="$shafile" \
  bash "$INSTALL_SH" --test-mode --verify-sha 2>&1) && rc=$? || rc=$?
rm -rf "$tmpdir"
if [[ $rc -ne 0 ]] || ! echo "$out" | grep -q "checksum ok"; then
  pass "sha256_verify_fail"
else
  fail "sha256_verify_fail"
fi

# Test 4: latest version falls back to dev when release lookup yields no tag.
name=$(STAGESERVE_TEST_DISABLE_RELEASE_LOOKUP=1 STAGESERVE_VERSION="latest" STAGESERVE_TEST_ONLY_RESOLVED_VERSION=1 /bin/bash "$INSTALL_SH" --test-mode 2>&1) && true
if [[ "$name" == "dev" ]]; then
  pass "latest_version_falls_back_to_dev"
else
  fail "latest_version_falls_back_to_dev"
  echo "Output was: $name"
fi

printf '\nResults: %d passed, %d failed\n' "$PASS" "$FAIL"
[[ $FAIL -eq 0 ]]
