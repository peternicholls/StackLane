#!/usr/bin/env bash
# Run all installer smoke test suites and report a combined result.
# Usage: bash scripts/tests/run_installer_tests.sh
# Exit 0 if all suites pass; 1 if any suite fails.

set -euo pipefail

SCRIPTS_DIR="$(cd "$(dirname "$0")" && pwd)"

SUITES=(
  "$SCRIPTS_DIR/install_handoff_smoke.sh"
  "$SCRIPTS_DIR/install_checksum_smoke.sh"
  "$SCRIPTS_DIR/install_install_smoke.sh"
)

TOTAL_PASS=0
TOTAL_FAIL=0
FAILED_SUITES=()

for suite in "${SUITES[@]}"; do
  name="$(basename "$suite")"
  printf '\n── %s ──\n' "$name"
  if bash "$suite"; then
    TOTAL_PASS=$((TOTAL_PASS+1))
  else
    TOTAL_FAIL=$((TOTAL_FAIL+1))
    FAILED_SUITES+=("$name")
  fi
done

printf '\n══════════════════════════════════════\n'
printf 'Installer test suites: %d passed, %d failed\n' "$TOTAL_PASS" "$TOTAL_FAIL"

if [[ ${#FAILED_SUITES[@]} -gt 0 ]]; then
  printf 'Failed suites:\n'
  for s in "${FAILED_SUITES[@]}"; do
    printf '  - %s\n' "$s"
  done
  exit 1
fi
