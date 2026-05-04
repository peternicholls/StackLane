# Data Model: Guided TUI And Next-Action Planning

## Guided Session

Represents one no-args `stage` invocation.

Fields:

- `cwd`: directory where the command started.
- `interactive`: whether stdin/stdout can support TUI interaction.
- `tui_disabled`: whether env/flag disables TUI.
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
- `no_tui`
- `no_color`
- `term`
- `reason`

Rules:

- TUI is allowed only when stdin and stdout are interactive and `no_tui` is false.
- Color is disabled when `NO_COLOR` is set or terminal capability is insufficient.
- Non-TTY fallback must not block for input.

## Guided Context

A cheap snapshot of StageServe's current situation.

Fields:

- `cwd`
- `project_root`
- `project_env_path`
- `project_env_exists`
- `project_env_valid`
- `stack_home`
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

Rules:

- Exactly one primary action should be present for known situations.
- Every action must include a direct command equivalent or an explicit reason why none exists.
- Warnings must be actionable.

## Guided Action

Represents one user-selectable operation.

Fields:

- `id`: stable action id such as `setup`, `init`, `up`, `status`, `logs`, `down`, `doctor`, `edit_config`, `advanced`.
- `label`
- `description`
- `mutates_state`
- `requires_confirmation`
- `direct_command`
- `expected_result`

Rules:

- Mutating actions require confirmation.
- Actions must route through existing command/domain behavior.
- Advanced actions may reveal implementation details, but primary actions should not.

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

