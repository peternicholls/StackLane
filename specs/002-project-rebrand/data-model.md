# Data Model: Stacklane Rebrand And Unified Command Surface

## Brand Identity

- Purpose: Represents the official user-facing identity of the project.
- Fields:
  - `display_name`: `Stacklane`
  - `command_name`: `stacklane`
  - `short_description`: single-line product description used across docs and help text
  - `legacy_name`: `Stacklane`
  - `legacy_prefix`: `20i-`
- Relationships:
  - Drives text replacement across documentation, app labels, workflow labels, and shell help.
- Validation rules:
  - `display_name` must be the only active brand outside explicitly labeled migration references.
  - `command_name` must match the primary documented CLI entrypoint.

## Command Action

- Purpose: Represents one operator-visible lifecycle action exposed through `stacklane`.
- Fields:
  - `flag_name`: one of `--up`, `--attach`, `--down`, `--detach`, `--status`, `--logs`, `--dns-setup`
  - `legacy_command`: matching current wrapper name such as `deprecated --up wrapper`
  - `runtime_action`: internal action passed to the existing helper engine
  - `supports_project_selector`: whether the action accepts `--project`
  - `supports_all`: whether the action accepts `--all`
  - `supports_service_name`: whether the action accepts trailing service selection for logs
- Relationships:
  - Each command action maps to zero or one legacy wrappers.
  - Each command action is described in the CLI contract and quickstart.
- Validation rules:
  - Exactly one primary action flag must be accepted per invocation.
  - Help text must describe the action in both primary and migration contexts.

## Legacy Wrapper

- Purpose: Represents a temporary compatibility script retained during migration.
- Fields:
  - `wrapper_name`: existing command file such as `deprecated --up wrapper`
  - `forwarded_action`: target `stacklane` action flag
  - `deprecation_message`: concise guidance that shows the preferred new syntax
  - `retention_state`: temporary migration support only
- Relationships:
  - Each legacy wrapper forwards to one `Command Action`.
  - Each wrapper is referenced from migration documentation.
- Validation rules:
  - Wrapper behavior must not become the primary documented path.
  - Wrapper guidance must point users to `stacklane` using the equivalent action syntax.

## Surface Inventory

- Purpose: Represents every maintained user-facing place that must stay aligned during the rebrand.
- Fields:
  - `surface_path`: repository path or asset name
  - `surface_type`: documentation, shell entrypoint, app label, workflow label, inline help, migration guide
  - `requires_brand_update`: boolean
  - `requires_command_update`: boolean
  - `notes`: migration or packaging caveats
- Relationships:
  - Surface inventory is driven by `Brand Identity` and `Command Action` changes.
- Validation rules:
  - A surface cannot keep the legacy brand or legacy command vocabulary unless explicitly marked as migration-only.

## State Transition Notes

- `Brand Identity`: legacy -> dual-reference migration state -> Stacklane primary state
- `Legacy Wrapper`: absent from docs -> retained with deprecation guidance -> removable in a later cleanup feature
- `Surface Inventory`: not updated -> updated for Stacklane vocabulary -> validated for no conflicting active names