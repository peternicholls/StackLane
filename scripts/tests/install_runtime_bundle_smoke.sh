#!/usr/bin/env bash
# Smoke tests — installer provisions bundled runtime assets.

set -euo pipefail

INSTALL_SH="$(cd "$(dirname "$0")/../.." && pwd)/install.sh"
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

mkdir -p "$tmpdir/runtime/stacks/20i"
printf 'services: {}\n' > "$tmpdir/runtime/stacks/20i/docker-compose.shared.yml"
printf 'services: {}\n' > "$tmpdir/runtime/stacks/20i/docker-compose.20i.yml"
tar -czf "$tmpdir/runtime.tar.gz" -C "$tmpdir/runtime" stacks

STAGESERVE_INSTALL_DIR="$tmpdir/bin" \
STAGESERVE_STACK_HOME="$tmpdir/stack-home" \
STAGESERVE_TEST_ASSET_PATH="/usr/bin/true" \
STAGESERVE_TEST_BUNDLE_PATH="$tmpdir/runtime.tar.gz" \
NONINTERACTIVE=1 \
bash "$INSTALL_SH" --test-mode >/tmp/stageserve-install-runtime-bundle.out 2>&1

test -x "$tmpdir/bin/stage"
test -f "$tmpdir/stack-home/stacks/20i/docker-compose.shared.yml"
test -f "$tmpdir/stack-home/stacks/20i/docker-compose.20i.yml"

printf 'PASS: install_runtime_bundle_provisioned\n'