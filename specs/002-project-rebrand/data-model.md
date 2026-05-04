# Data Model: StageServe Rebrand And `stage` Command Cutover

## Brand Identity

- Purpose: Represents the official user-facing identity of the project.
- Fields:
  - `display_name`: `StageServe`
  - `command_name`: `stage`
  - `short_description`: single-line product description used across docs and help text
  - `historical_name`: prior approved rename candidate
  - `historical_command_name`: prior root command name
- Relationships:
  - Drives naming alignment across documentation, spec text, and help output.
- Validation rules:
  - `display_name` must be the only active brand outside explicitly labeled migration references.
  - `command_name` must match the primary documented CLI entrypoint.

## Command Surface

- Purpose: Represents one operator-visible lifecycle action exposed through `stage`.
- Fields:
  - `subcommand_name`: one of `up`, `attach`, `down`, `detach`, `status`, `logs`, `dns-setup`, `doctor`, `init`, `setup`
  - `supports_project_selector`: whether the action accepts `--project`
  - `supports_all`: whether the action accepts `--all`
  - `supports_service_name`: whether the action accepts trailing service selection for logs
- Relationships:
  - Each subcommand is described in the CLI contract and quickstart.
- Validation rules:
  - Help text must describe current subcommands using StageServe terminology.
  - No legacy wrapper command may be represented as an active command surface.

## Config And State Contract

- Purpose: Represents the operator-visible naming boundary for config and runtime state.
- Fields:
  - `project_config_file`: `.env.stageserve`
  - `stack_defaults_file`: `.env.stageserve`
  - `state_dir`: `.stageserve-state`
  - `root_command_dir`: `cmd/stage`
- Relationships:
  - Shared by operator docs, the current CLI, and older spec text.
- Validation rules:
  - Maintained spec and doc text must use these names consistently.

## Surface Inventory

- Purpose: Represents every maintained user-facing place that must stay aligned during the rename.
- Fields:
  - `surface_path`: repository path or asset name
  - `surface_type`: documentation, spec, shell entrypoint, inline help, migration guide
  - `requires_brand_update`: boolean
  - `requires_command_update`: boolean
  - `requires_directory_update`: boolean
- Relationships:
  - Surface inventory is driven by `Brand Identity`, `Command Surface`, and `Config And State Contract`.
- Validation rules:
  - A maintained surface cannot keep prior naming unless explicitly marked as migration-only or archival.

## State Transition Notes

- `Brand Identity`: prior naming -> StageServe primary state
- `Command Surface`: prior root command naming -> `stage <subcommand>` primary state
- `Config And State Contract`: prior config/state names -> `.env.stageserve` and `.stageserve-state`
- `Surface Inventory`: not updated -> updated for StageServe vocabulary -> validated for no conflicting active names