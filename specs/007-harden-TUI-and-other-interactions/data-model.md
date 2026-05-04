# Data Model: Guided TUI And Next-Action Planning

## Guided Session

Represents one no-args `stage` invocation.

Fields:

- `cwd`: directory where the command started.
- `interactive`: whether stdin/stdout can support TUI interaction.
- `tui_disabled`: whether shell environment disables TUI.
- `color_disabled`: whether color should be disabled.
- `plan`: the current `NextActionPlan`.
- `selected_action`: the current action, if any.
- `confirmed`: whether the user confirmed a mutating action.
- `result`: action result, when an action has run.

Rules:

- A session must not mutate state before confirmation.
- A session may be represented in TUI or text fallback.
- A session must preserve direct command equivalents for every action.

## TUI Capability

Represents terminal suitability.

Fields:

- `stdin_tty`
- `stdout_tty`
- `stderr_tty`
- `notui_flag`
- `cli_flag`
- `no_tui_shell_env`
- `no_color`
- `term`
- `reason`

Rules:

- TUI is allowed only when stdin and stdout are interactive, `notui_flag` is false, `cli_flag` is false, and `no_tui_shell_env` is false.
- Color is disabled when `NO_COLOR` is set or terminal capability is insufficient.
- Non-TTY fallback must not block for input.
- `STAGESERVE_NO_TUI` is shell-env-only and must not be read from project or stack `.env.stageserve`.

## Guided Context

A cheap snapshot of StageServe's current situation.

Fields:

- `cwd`
- `project_root`
- `project_env_path`
- `project_env_exists`
- `project_env_valid`
- `stack_home`
- `stack_id`: normalized `STAGESERVE_STACK` value, currently `20i`.
- `state_dir`
- `machine_readiness_summary`
- `project_state`
- `runtime_summary`
- `warnings`

Rules:

- Context collection should avoid expensive checks before first render.
- Expensive checks can be run when the user selects setup, doctor, status, or refresh.
- Context collection must not create `.env.stageserve`.

## Next Action Plan

Planner output consumed by TUI and text fallback.

Fields:

- `situation`: one of `machine_not_ready`, `project_missing_config`, `project_ready_to_run`, `project_running`, `project_down`, `drift_detected`, `not_project`, `unknown_error`.
- `title`
- `summary`
- `primary_action`
- `secondary_actions`
- `advanced_actions`
- `warnings`
- `direct_commands`
- `plain_language_terms`: copy decisions used by TUI and text fallback for the current plan.

Rules:

- Exactly one primary action should be present for known situations.
- Every action must include a direct command equivalent or an explicit reason why none exists.
- Warnings must be actionable.
- Titles, summaries, warnings, and action labels must describe user goals before implementation mechanics.
- Implementation words such as attach, detach, daemon, gateway, compose, container, registry, and runtime must stay out of first-level copy unless no StageServe-level phrase can explain the required action.

Canonical situation semantics:

| Situation | Meaning |
|---|---|
| `machine_not_ready` | Setup-level prerequisites block normal operation. |
| `project_missing_config` | Current directory can be a project, but `.env.stageserve` is absent. |
| `project_ready_to_run` | Project config exists and no retained down state or active runtime needs special handling. |
| `project_running` | Project is recorded and runtime/status checks indicate it is active. |
| `project_down` | Project has retained StageServe state marked down. |
| `drift_detected` | State, runtime, DNS, gateway, or config disagree and diagnosis is safer than normal action. |
| `not_project` | Current directory is not usable as a StageServe project root for guided project actions. |
| `unknown_error` | Context collection or planning failed without a safe classification. |

## Guided Action

Represents one user-selectable operation.

Fields:

- `id`: stable action id such as `setup`, `init`, `up`, `attach`, `status`, `logs`, `down`, `detach`, `doctor`, `diagnose`, `setup_help`, `recovery_help`, `edit_config`, `show_commands`, `advanced`.
- `label`: easy-mode label shown in the TUI, such as "Run this project" or "Remove this project from StageServe".
- `description`: plain-language explanation of what will happen and why it helps.
- `internal_name`: command or domain name used by implementation and "show commands", when different from the label.
- `mutates_state`
- `requires_confirmation`
- `direct_command`
- `expected_result`

Rules:

- Mutating actions require confirmation.
- Actions must route through existing command/domain behavior.
- Advanced actions may reveal implementation details, but primary actions should not.
- Direct command names are command equivalents, not first-level labels.
- The TUI and text fallback must use the same `label` and `description` for the same action.
- Planner-only ids such as `diagnose`, `setup_help`, and `recovery_help` may alias existing command/domain behavior through `internal_name` instead of introducing new command surfaces.

Recommended easy-mode labels:

| Internal Action | Easy-Mode Label |
|---|---|
| `setup` | Set up this computer |
| `init` | Create project settings |
| `up` | Run this project |
| `attach` | Add this project to StageServe |
| `status` | Check project status |
| `logs` | View project logs |
| `down` | Stop this project |
| `detach` | Remove this project from StageServe |
| `doctor` | Find issues |
| `diagnose` | Find issues |
| `setup_help` | Get setup help |
| `recovery_help` | Show recovery help |
| `edit_config` | Edit project settings |
| `show_commands` | Show commands |
| `advanced` | Advanced troubleshooting |

## Config Preview

Represents a pending `.env.stageserve` write.

Fields:

- `path`
- `values`
- `comments`
- `overwrite`
- `source`

Rules:

- Show before write.
- Preserve existing overwrite protection unless user explicitly confirms force behavior.
- Write only `.env.stageserve`.

## Recovery Path

Represents guidance for a non-ready or failed state.

Fields:

- `problem`
- `why_it_matters`
- `primary_fix`
- `direct_command`
- `advanced_detail`

Rules:

- The primary fix should be a StageServe command or config edit.
- Advanced detail can include Docker only when needed.
