# Contract: Guided TUI And Root Interaction

## Root Command

## Contract Override Notes

This contract is the final authority for spec 007 guided interaction behavior. Where earlier drafts, spec 005 onboarding behavior, or historical implementation notes differ, this contract overrules them.

Resolved overrides:

- Bare `stage` is the guided entrypoint in interactive terminals.
- The final contract does not include `--tui`.
- `--notui` and `--cli` are the opt-out controls for the current invocation.
- `stage init` participates in the guided default model in interactive terminals.
- Easy-mode labels are plain-language labels first; direct command names are secondary and appear through "show commands", direct help, or advanced/troubleshooting views.

### `stage`

Interactive terminal:

- Starts the guided TUI unless disabled.
- Detects current context.
- Shows one primary action, secondary actions, advanced actions, and quit/help.

Non-interactive terminal:

- Prints compact text guidance.
- Does not prompt.
- Exits 0 unless context collection itself fails fatally.

Disabled TUI:

- `stage --notui` uses text fallback for bare `stage`.
- `stage --cli` is an alias for `stage --notui`.
- `STAGESERVE_NO_TUI=1 stage` uses text fallback.
- `STAGESERVE_NO_TUI` is shell-env-only and is not read from `.env.stageserve`.
- `NO_COLOR=1` disables color, not the TUI itself.

Explicit help:

- `stage --help` and `stage -h` show Cobra help.

Direct command:

- `stage <subcommand>` bypasses root TUI routing.

## Plain-Language Contract

The guided TUI is the easy-mode surface. Its first-level labels must describe user goals, not implementation mechanics.

Rules:

- Primary and secondary action labels use plain user language.
- Direct command names appear in "show commands", direct-command help, or advanced/troubleshooting views.
- First-level copy avoids implementation terms such as attach, detach, daemon, gateway, compose, container, registry, and runtime unless there is no clearer StageServe-level wording.
- When an implementation term is unavoidable for recovery, the copy must pair it with the user action, for example "Start Docker Desktop, then run setup again."
- Direct command equivalents remain available for every action.

Recommended label mapping:

| Internal Action | Easy-Mode Label | Direct Command |
|---|---|---|
| `setup` | Set up this computer | `stage setup` |
| `init` | Create project settings | `stage init` |
| `up` | Run this project | `stage up` |
| `attach` | Add this project to StageServe | `stage attach` |
| `status` | Check project status | `stage status` |
| `logs` | View project logs | `stage logs` |
| `down` | Stop this project | `stage down` |
| `detach` | Remove this project from StageServe | `stage detach` |
| `doctor` | Find issues | `stage doctor` |
| `diagnose` | Find issues | `stage doctor` |
| `setup_help` | Get setup help | `stage setup` |
| `recovery_help` | Show recovery help | none |
| `edit_config` | Edit project settings | `.env.stageserve` |
| `show_commands` | Show commands | none |
| `advanced` | Advanced troubleshooting | none |

## Required Guided Situations

| Situation | Primary Action | Secondary Actions |
|---|---|---|
| `machine_not_ready` | Set up this computer | Find issues, show commands, quit |
| `project_missing_config` | Create project settings | Edit settings, show commands, quit |
| `project_ready_to_run` | Run this project | Check project status, edit settings, find issues, show commands, quit |
| `project_running` | Check project status | View logs, stop this project, find issues, show commands, quit |
| `project_down` | Run this project | Check project status, remove this project from StageServe, find issues, show commands, quit |
| `drift_detected` | Find issues | Check project status, view logs, show commands, quit |
| `not_project` | Get setup help | Show commands, quit |
| `unknown_error` | Show recovery help | Find issues when project context is available, show commands, quit |

Situation semantics:

- `project_ready_to_run`: project config exists and there is no retained down record or active runtime requiring special handling.
- `project_down`: StageServe has a retained record for the project marked down.
- `unknown_error`: planning/context collection failed and normal action cannot be chosen safely.

## Action Execution Rules

- Mutating actions require confirmation.
- Config writes show preview before write.
- Lifecycle actions use existing lifecycle semantics.
- Setup/init/doctor actions use existing onboarding result semantics.
- Status/log actions use existing observability/logging semantics.
- Ctrl-C cancels the current session or action.
- Cancellation before confirmation leaves no changes.
- Ctrl-C during logs or long-running actions must leave the terminal usable and surface the safest next action.

## Output Rules

- TUI owns terminal rendering only in TTY mode.
- JSON command modes remain pure JSON.
- Text fallback contains the same core guidance as the TUI.
- Advanced implementation detail must not appear on the first screen unless it is the only actionable recovery path.
- Text fallback follows the same plain-language rules as the TUI.

## Advanced Actions

- `show_commands` displays the direct StageServe commands for the current situation.
- `advanced` may show implementation details such as state paths or Docker-oriented troubleshooting only after the user chooses the advanced path.
- `edit_config` shows the project `.env.stageserve` path and StageServe-level guidance for editing it. Spec 007 does not launch an external editor from the guided TUI.

## Direct Command Contract

The following command forms define the direct CLI contract for spec 007:

- `stage setup`
- `stage setup --notui`
- `stage setup --cli`
- `stage setup --json`
- `stage init`
- `stage init --notui`
- `stage init --cli`
- `stage doctor`
- `stage doctor --notui`
- `stage doctor --cli`
- `stage doctor --json`
- `stage up`
- `stage attach`
- `stage status`
- `stage logs`
- `stage down`
- `stage detach`
- `stage down --all`

The final spec 007 contract does not include `--tui`; TUI is the default easy-mode behavior, and `--notui` / `--cli` are equivalent opt-outs.
