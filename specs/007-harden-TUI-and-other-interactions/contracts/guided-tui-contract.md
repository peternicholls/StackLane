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
- Renders the current context through the guided surfaces: status header, decision bar when a real user choice exists, tool work panel when StageServe is doing setup or recovery work, and persistent footer affordances.

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

- Decision item labels use plain user language.
- Direct command names appear in "show commands", direct-command help, or advanced/troubleshooting views.
- First-level copy avoids implementation terms such as attach, detach, daemon, gateway, compose, container, registry, and runtime unless there is no clearer StageServe-level wording.
- When an implementation term is unavoidable for recovery, the copy must pair it with the user action, for example "Start Docker Desktop, then run setup again."
- Direct command equivalents remain available for every action.
- `stage doctor` is not exposed as a peer first-level action. Equivalent checks run inline when StageServe detects a blocker, and the direct command appears through "show commands" or advanced troubleshooting.
- Examples use `.develop` when no active configuration is being demonstrated, but renderers must show the configured suffix, URL scheme, and port from StageServe's effective config/capabilities.

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
| `doctor` | Troubleshoot this problem | `stage doctor` |
| `diagnose` | Troubleshoot this problem | `stage doctor` |
| `init_here` | Set up this directory as a project | `stage init` |
| `setup_help` | Get setup help | `stage setup` |
| `recovery_help` | Show recovery help | none |
| `edit_config` | Edit project settings | `.env.stageserve` |
| `show_commands` | Show commands | none |
| `advanced` | Advanced troubleshooting | none |

## Required Guided Situations

The situations below are planner outputs, not screens the user navigates between. Each situation maps to surfaces. Setup, diagnostics, and recovery are tool-owned work panels; direct commands and advanced details live in the footer path.

| Situation | Status Header | Decision Bar | Tool Work Panel | Footer |
|---|---|---|---|---|
| `machine_not_ready` | "Your computer isn't ready yet." | Hidden while checklist is active | Ordered setup checklist; pauses only for approval or external blockers | Help, details, show direct commands, plain text output, quit |
| `project_missing_config` | "This folder doesn't have StageServe settings yet." | Use these settings / Edit before writing | Config preview with target path, values, URL, and validation notes | Help, show direct commands, plain text output, quit |
| `project_ready_to_run` | "This project is ready to run." | Run this project / Edit project settings | Hidden unless start fails, then progress or blocker panel | Help, show direct commands, advanced troubleshooting, quit |
| `project_running` | "This project is running at <URL>." | View project logs / Stop this project | Hidden unless logs or progress are active | Open URL, help, show direct commands, advanced troubleshooting, quit |
| `project_down` | "This project is stopped." | Run this project or Add this project to StageServe / Remove this project from StageServe | Hidden unless start/add/remove progress is active | Help, show direct commands, advanced troubleshooting, quit |
| `drift_detected` | "This project doesn't match what StageServe expects." | Use the safe next step / Try to start it again / Show what doesn't match | Plain-language comparison and safe-step preview | Help, show direct commands, advanced troubleshooting, quit |
| `not_project` | "This folder isn't a StageServe project yet." | Set up this folder as a project / Pick a different folder | Proposed defaults and path context when available | Help, show direct commands, plain text output, quit |
| `unknown_error` | "StageServe couldn't safely choose a next step." | Run next recovery step / Show what went wrong / Stop here | Ordered recovery path from least invasive to most invasive | Help, show direct commands, advanced troubleshooting, quit |

Situation semantics:

- `project_ready_to_run`: project config exists and there is no retained down record or active runtime requiring special handling.
- `project_down`: StageServe has a retained record for the project marked down. If the project runtime is already running, the primary action becomes `attach` with the easy-mode label "Add this project to StageServe".
- `unknown_error`: planning/context collection failed and normal action cannot be chosen safely. The recovery panel must list a concrete ordered next-step sequence (typically `stage doctor`, `stage status`, `stage logs`) rather than generic guidance.

## Visible Defaults And Local URL Rules

- Every value StageServe will use must be visible before the user commits: site name, web folder, suffix, URL scheme, port when non-default, target `.env.stageserve` path, and selected stack.
- First-level project setup edits are limited to project name, web folder, and local address/suffix. Existing PHP, MySQL, timeout, hostname, stack-home, or post-up custom settings are summarized and inspectable through details/advanced paths.
- The renderer must not hard-code the example suffix. `.develop` is the easy-mode example suffix, while existing `.test`, `.dev`, full-hostname, or custom-suffix projects must render exactly as configured.
- Local DNS copy must lead with the user outcome: "your computer can open `<project>.<suffix>`". Resolver files, `dnsmasq`, and service names are detail-view or advanced wording.
- HTTPS/certificate copy appears as required only when the selected local URL is HTTPS. For plain HTTP `.develop` examples, certificate setup is optional or advanced unless another contract makes it required.
- If `SITE_SUFFIX` and `LOCAL_DNS_SUFFIX` differ, the TUI must show the mismatch and recommend the lowest-risk correction before offering run/start actions.

## Action Execution Rules

- Mutating actions require confirmation.
- Config writes show preview before write.
- Lifecycle actions use existing lifecycle semantics.
- Setup/init/doctor actions use existing onboarding result semantics.
- Status/log actions use existing observability/logging semantics.
- Ctrl-C cancels the current session or action.
- Cancellation before confirmation leaves no changes.
- Ctrl-C during logs or long-running actions must leave the terminal usable and surface the safest next action.
- Pressing `enter` on a running-project screen must not stop, remove, or rewrite anything unless the user has explicitly selected and confirmed that mutating action.
- Out-of-sync recovery that changes StageServe records must confirm exactly what will and will not change.

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
