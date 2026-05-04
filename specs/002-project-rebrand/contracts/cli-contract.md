# CLI Contract: StageServe

## Primary Entry Point

- Command name: `stage`
- Responsibility: provide one canonical operator-facing command surface for current lifecycle actions.

## Subcommand Rules

- Supported primary subcommands include:
  - `up`
  - `attach`
  - `down`
  - `detach`
  - `status`
  - `logs`
  - `dns-setup`
  - `doctor`
  - `init`
  - `setup`
- Unknown subcommands must fail non-zero with concise usage guidance.
- The root command help must make the subcommand model obvious.

## Shared Options

- `--project-dir PATH`
- `--project SELECTOR`
- `--site-name NAME`
- `--site-hostname HOST`
- `--site-suffix SUFFIX`
- `--docroot PATH`
- `--php-version VERSION`
- `--mysql-database NAME`
- `--mysql-user USER`
- `--mysql-password PASS`
- `--mysql-port PORT`
- `--pma-port PORT`
- `--all`
- `--dry-run`
- `--help`

## Command Mapping

| Subcommand | Responsibility | Notes |
|---|---|---|
| `up` | Starts the current project and ensures shared infrastructure | Canonical runtime bring-up path |
| `attach` | Attaches an additional project | Reuses the shared gateway when healthy |
| `down` | Stops the current project | `--all` preserves global teardown behavior |
| `detach` | Stops the current project and removes its attachment record | |
| `status` | Reports gateway, DNS, registry, and project state | |
| `logs` | Follows logs for a selected project runtime | Supports optional service selection |
| `dns-setup` | Bootstraps local DNS on macOS | |
| `doctor` | Diagnoses drift and readiness issues | |
| `init` | Writes project-local starter config | |
| `setup` | Performs machine-readiness checks and one-time setup | |

## Compatibility Contract

- `stage` is the only supported active executable path.
- Compatibility forwarding shims are not part of the supported contract.
- Historical command names may appear only in migration or archival notes.

## Help Contract

- `stage --help` must present:
  - the StageServe brand name
  - the one-command mental model
  - supported subcommands
  - shared options
  - at least one representative usage example

## Error Contract

- Invalid subcommands must fail clearly.
- Unsupported arguments must identify the offending token.
- Failure output must not obscure current runtime diagnostics.