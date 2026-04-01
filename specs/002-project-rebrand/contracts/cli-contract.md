# CLI Contract: Stacklane

## Primary Entry Point

- Command name: `stacklane`
- Responsibility: provide one canonical operator-facing command surface for all current lifecycle actions.

## Action Selection Rules

- Exactly one primary action flag must be supplied per invocation.
- Supported primary action flags:
  - `--up`
  - `--attach`
  - `--down`
  - `--detach`
  - `--status`
  - `--logs`
  - `--dns-setup`
- If zero primary action flags are provided, the command must exit non-zero and print concise usage guidance.
- If more than one primary action flag is provided, the command must exit non-zero and explain that actions are mutually exclusive.

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
- Compatibility alias support such as `version=8.4` remains available during migration unless explicitly removed in implementation.

## Action Mapping

| Primary action | Internal runtime action | Legacy wrapper | Notes |
|---|---|---|---|
| `--up` | `up` | `20i-up` | Starts current project and ensures shared infrastructure |
| `--attach` | `attach` | `20i-attach` | Attaches an additional project |
| `--down` | `down` | `20i-down` | Stops current project; `--all` keeps global teardown behavior |
| `--detach` | `detach` | `20i-detach` | Stops current project and removes its record |
| `--status` | `status` | `20i-status` | Reports gateway, DNS, registry, and project state |
| `--logs` | `logs` | `20i-logs` | Supports optional service selection |
| `--dns-setup` | `dns-setup` | `20i-dns-setup` | Bootstraps local DNS on macOS |

## Legacy Wrapper Contract

- Existing `20i-*` scripts remain temporarily available only as deprecated migration wrappers.
- Each wrapper must forward to the equivalent `stacklane` action.
- Each wrapper must emit deprecation guidance that shows the preferred `stacklane` syntax and warns that the wrapper will be removed in a future update.
- Wrappers must not be the primary path in help text or top-level docs.

## Help Contract

- `stacklane --help` must present:
  - the Stacklane brand name
  - the one-command mental model
  - supported primary action flags
  - shared options
  - at least one representative usage example
  - legacy migration note pointing from `20i-*` commands to `stacklane`

## Error Contract

- Invalid action combinations must fail clearly.
- Unsupported arguments must identify the offending token.
- Migration guidance must remain concise and actionable.
- Failure output must not obscure existing runtime diagnostics from the helper engine.