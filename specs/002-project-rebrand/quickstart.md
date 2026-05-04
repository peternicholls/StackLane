# Quickstart: StageServe Rename Validation

## Goal

Validate the StageServe rename and canonical `stage` command surface without widening runtime behavior.

## Preconditions

- macOS with Docker Desktop available
- Repository changes synced to the location you actually run as the stack, if that differs from the dev workspace
- A sample project directory containing either `public_html/` or another valid docroot

## Happy-Path Validation

1. Add the repo root to `PATH` or invoke the command by absolute path.
2. Run `stage --help`.
3. Confirm help text uses StageServe branding and shows the current subcommands.
4. From a sample project directory, run `stage up`.
5. Confirm the project starts and reports the expected hostname route and gateway probe.
6. Run `stage status` and confirm state, route, and runtime details appear.
7. Run `stage down` and confirm the project stops cleanly.

## Migration Validation

1. Review the top-level docs and migration guide.
2. Confirm they describe `stage` as the only supported root command.
3. Confirm they use `.env.stageserve` and `.stageserve-state` as the current directory names.
4. Confirm any prior names are explicitly labeled as historical or archival.

## Failure-Path Validation

1. Run `stage` with an invalid subcommand such as `stage not-a-command`.
2. Confirm the command exits non-zero and prints concise usage guidance.
3. Run `stage up --help`.
4. Confirm the subcommand help renders current terminology and examples.

## Documentation Validation

1. Review `README.md`, `docs/migration.md`, `docs/runtime-contract.md`, and the maintained older specs.
2. Confirm StageServe is the active project name across those surfaces.
3. Confirm the repo rename and manual local-folder handling are described separately.
4. Confirm examples use `stage up`, `stage down`, and related subcommand syntax.

## Validation Notes

- 2026-04-01: Validated `stage --help`, `stage status`, `stage up --dry-run`, and `stage down --dry-run` from the repository root.
- 2026-04-01: Validated invalid-subcommand handling from the root command.
- 2026-04-01: Untested caveat: macOS app and workflow packaging were not exercised through Finder or Services UI during this pass.
- 2026-04-01: GitHub repository rename completed to `StageServe`; local folder naming remains an operator-managed concern.