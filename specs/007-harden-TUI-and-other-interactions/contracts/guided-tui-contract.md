# Contract: Guided TUI And Root Interaction

## Root Command

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

- `STAGESERVE_NO_TUI=1 stage` uses text fallback.
- `NO_COLOR=1` disables color, not the TUI itself.

Explicit help:

- `stage --help` and `stage -h` show Cobra help.

Direct command:

- `stage <subcommand>` bypasses root TUI routing.

## Required Guided Situations

| Situation | Primary Action | Secondary Actions |
|---|---|---|
| Machine not ready | Run setup | Doctor, show commands, quit |
| Project missing config | Create project config | Edit values, show commands, quit |
| Project ready to run | Run project | Status, edit config, doctor, quit |
| Project running | View status | Logs, stop, doctor, quit |
| Project down | Run project | Status, detach, doctor, quit |
| Drift detected | Diagnose | Status, logs, show commands, quit |
| Not a project | Show setup/help | Choose project dir, advanced commands, quit |

## Action Execution Rules

- Mutating actions require confirmation.
- Config writes show preview before write.
- Lifecycle actions use existing lifecycle semantics.
- Setup/init/doctor actions use existing onboarding result semantics.
- Status/log actions use existing observability/logging semantics.
- Ctrl-C cancels the current session or action.
- Cancellation before confirmation leaves no changes.

## Output Rules

- TUI owns terminal rendering only in TTY mode.
- JSON command modes remain pure JSON.
- Text fallback contains the same core guidance as the TUI.
- Advanced implementation detail must not appear on the first screen unless it is the only actionable recovery path.

## Direct Command Compatibility

The following command forms must keep their existing behavior:

- `stage setup`
- `stage setup --json`
- `stage init`
- `stage doctor`
- `stage doctor --json`
- `stage up`
- `stage attach`
- `stage status`
- `stage logs`
- `stage down`
- `stage down --all`

