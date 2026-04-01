# Quickstart: Stacklane Rebrand Validation

## Goal

Validate the Stacklane rename and unified command surface without changing the underlying runtime contract.

## Preconditions

- macOS with Docker Desktop available
- Repository changes synced to the location you actually run as the stack, if that differs from the dev workspace
- A sample project directory containing either `public_html/` or another valid docroot

## Happy-Path Validation

1. Add the repo root to `PATH` or invoke the command by absolute path.
2. Run `stacklane --help`.
3. Confirm help text uses Stacklane branding and shows the primary action flags.
4. From a sample project directory, run `stacklane --up`.
5. Confirm the project starts and reports the expected hostname route and gateway probe.
6. Run `stacklane --status` and confirm state, route, and runtime details appear.
7. Run `stacklane --down` and confirm the project stops cleanly.

## Migration Validation

1. Invoke one retained wrapper such as `20i-up` from a sample project.
2. Confirm the command still works or forwards correctly during the migration window.
3. Confirm the wrapper surfaces deprecation guidance toward `stacklane --up`.
4. Review the top-level docs and migration guide to confirm they describe `stacklane` as the primary interface and `20i-*` as migration-only wrappers.

## Failure-Path Validation

1. Run `stacklane` with no primary action flag.
2. Confirm the command exits non-zero and prints concise usage guidance.
3. Run `stacklane --up --down`.
4. Confirm the command exits non-zero and reports that primary actions are mutually exclusive.

## Documentation Validation

1. Review `README.md`, `AUTOMATION-README.md`, `GUI-HELP.md`, `docs/migration.md`, and `docs/runtime-contract.md`.
2. Confirm Stacklane is the active project name across those surfaces.
3. Confirm the repo rename and manual containing-folder rename are described separately.
4. Confirm examples use `stacklane --up`, `stacklane --down`, and related action-flag syntax.

## Validation Notes

- 2026-04-01: Validated `stacklane --help`, `stacklane --status`, `stacklane --status --project 20i-stack`, `stacklane --up --dry-run`, and `stacklane --down --dry-run` from the repository root.
- 2026-04-01: Validated failure handling for `stacklane` with no primary action and with conflicting primary actions (`--up --down`).
- 2026-04-01: Validated `20i-up --dry-run` forwards to `stacklane --up` and prints deprecation guidance.
- 2026-04-01: Recompiled `20i Stack Manager.app/Contents/Resources/Scripts/main.scpt` from the updated AppleScript source.
- 2026-04-01: Untested caveat: the macOS app and workflow packaging were not exercised through Finder or Services UI during this implementation pass.
- 2026-04-01: Untested caveat: `shellcheck` was not available in the current environment.
- 2026-04-01: GitHub repository rename completed to `StackLane`; clone URL guidance was updated while keeping the local containing-folder rename manual.