# Feature Specification: Guided TUI And Simple-First StageServe Interaction

**Feature Branch**: `007-harden-TUI-and-other-interactions`  
**Created**: 2026-05-04  
**Status**: Draft  
**Input**: User description: "Flesh out the recovery plan, specification, and planning documents for spec 007. Research similar guided TUI style interfaces, what has worked and what has not, and bring that into planning."

## Decision Overrides

The following decisions are resolved for spec 007 and overrule earlier draft assumptions, soft guidance, or spec 005-era onboarding behavior:

- Bare `stage` in an interactive terminal is the default easy-mode entrypoint. This overrules help-first root behavior and any assumption that users should begin by learning subcommands.
- TUI is the default behavior. This overrules any prior expectation that TUI must be explicitly requested.
- `--notui` and `--cli` are the supported opt-outs for the current invocation. This overrules earlier `--tui`, `--no-tui`, or unresolved flag-shape discussion for the final spec 007 contract.
- `stage init` opens the guided project-config form by default in interactive terminals. This overrules any earlier assumption that init remains text-only, JSON-only, or undecided.
- `STAGESERVE_NO_TUI=1` is shell-environment-only. This overrules any interpretation that `.env.stageserve` should disable the guided TUI.
- Easy-mode labels use plain user-goal language. This overrules any prior use of command words such as `attach` and `detach` as first-level guided labels.
- Spec 007 is still developer-only, so unreleased flag names may be cleaned up. This overrules any pressure to preserve pre-release TUI flag experiments for backward compatibility.
- Terminal verification is the primary development gate for this spec run. This overrules the prior TDD-first default used in earlier spec runs.

## User Scenarios & Testing

### User Story 1 - Start From Bare `stage` (Priority: P1)

A normal user runs `stage` from a terminal and gets a guided StageServe entrypoint that tells them the current situation and offers the next safe action.

**Why this priority**: This is the central gap. The original intention was that `stage` alone exposes the simple path; current behavior still requires the user to know subcommands first.

**Independent Test**: Run `stage` in an interactive terminal from a project directory and verify it opens a guided TUI that detects context and offers a primary action without requiring Docker knowledge.

**Acceptance Scenarios**:

1. **Given** stdout is a TTY and no subcommand is provided, **When** the operator runs `stage`, **Then** StageServe opens the guided TUI rather than only showing help.
2. **Given** stdout is not a TTY and no subcommand is provided, **When** the operator runs `stage`, **Then** StageServe prints concise next-step guidance and exits without blocking for input.
3. **Given** the operator runs `stage --help`, **When** help is requested explicitly, **Then** Cobra help is shown and the guided TUI is not launched.
4. **Given** the operator runs `stage up`, **When** a direct subcommand is provided, **Then** StageServe runs the direct CLI path instead of opening the root guided TUI.

---

### User Story 2 - Complete First-Run Machine And Project Setup (Priority: P1)

A first-time user opens the guided TUI and is walked through machine readiness, project config creation, and the next run action without needing to know command order.

**Why this priority**: Setup and init exist, but they are still separate stepping stones. The TUI must turn them into one guided path.

**Independent Test**: In a project without `.env.stageserve`, run `stage` in a clean or simulated not-ready environment and confirm the TUI offers setup, config creation, and run in sequence.

**Acceptance Scenarios**:

1. **Given** machine readiness is incomplete, **When** the guided TUI starts, **Then** it shows the blocking readiness step and offers `stage setup` as the primary action.
2. **Given** the project lacks `.env.stageserve`, **When** readiness is acceptable, **Then** the TUI proposes a starter config and lets the user confirm or edit values before writing.
3. **Given** project config exists and the project is stopped, **When** the TUI starts, **Then** it offers to run the project.
4. **Given** setup or init cannot complete automatically, **When** the TUI displays the issue, **Then** it gives a specific recovery path and the equivalent direct command.

---

### User Story 3 - Manage A Running Project Simply (Priority: P2)

A user working on a configured project opens `stage` and can inspect status, follow logs, stop the project, or diagnose drift from one simple interface.

**Why this priority**: The simple path must cover day-to-day use, not just first run.

**Independent Test**: With a running project, run `stage` and verify the guided TUI offers user-goal actions such as check status, view logs, stop project, find issues, and show commands without exposing Docker, gateway, attach, or detach terminology as the primary model.

**Acceptance Scenarios**:

1. **Given** the current project is running, **When** the TUI starts, **Then** the primary screen shows project identity, route/status summary, and common actions.
2. **Given** the user chooses logs, **When** logs are followed, **Then** the UI provides a clear exit path and does not corrupt the terminal.
3. **Given** the user chooses stop, **When** the stop action is confirmed, **Then** `stage down` semantics are used and project data is preserved.
4. **Given** drift is detected, **When** the TUI presents recovery, **Then** the path starts with StageServe commands before advanced implementation detail.

---

### User Story 4 - Define Power-User And Automation Paths (Priority: P1)

A power user or script can bypass all TUI behavior and use direct commands, flags, and JSON output. This spec is still developer-only, so existing unreleased command flags may be renamed or removed cleanly to match the final contract.

**Why this priority**: A guided first-level path must not block shell-first usage or automation, but the command contract can still be cleaned up before release.

**Independent Test**: Run direct commands and JSON modes under non-TTY conditions and verify they follow the final spec 007 CLI contract.

**Acceptance Scenarios**:

1. **Given** `STAGESERVE_NO_TUI=1`, **When** the operator runs `stage`, **Then** StageServe prints plain guidance instead of opening the TUI.
2. **Given** `stage setup --json`, **When** the command runs, **Then** stdout remains a valid JSON envelope with no styled guidance.
3. **Given** `stage init --notui` or `stage init --cli`, **When** the command runs, **Then** it uses plain text output.
4. **Given** a direct command such as `stage down --all`, **When** it runs, **Then** StageServe runs the direct CLI path and does not open the guided TUI.

## Edge Cases

- stdout is a TTY but stdin is not interactive.
- terminal has limited color or reports an unknown `TERM`.
- user presses Ctrl-C before confirming any mutation.
- user presses Ctrl-C during a long-running action.
- `.env.stageserve` exists but is invalid.
- current directory is not a project root.
- machine setup requires privileged DNS work.
- Docker is missing or daemon is stopped.
- project is recorded as available through StageServe but runtime is missing.
- multiple projects are available through StageServe.
- `NO_COLOR`, `STAGESERVE_NO_TUI`, `--notui`, or `--cli` is set.
- command is run in CI or through redirected stdout.
- an implementation command name is precise for power users but confusing as an easy-mode label.
- when multiple projects are available through StageServe, the first implementation remains scoped to the current directory rather than adding a project switcher.

## Operational Impact

### Ease Of Use & Workflow Impact

- Affected entry points: bare `stage`, `stage setup`, `stage init`, `stage doctor`, `stage up`, `stage status`, `stage logs`, `stage down`, help text, installer handoff, README first-run guidance.
- Compatibility expectation: none for unreleased developer-only command flags. Spec 007 may rename or remove pre-release TUI flags, but the final direct CLI contract must be internally consistent and terminal-verified.
- Friction removed: users no longer need to know the exact first command or command sequence before StageServe can help them.
- Friction introduced: users who expect help from bare `stage` in a TTY need `stage --help`, `STAGESERVE_NO_TUI=1`, `--notui`, or `--cli`.

### Configuration & Precedence

- User-facing StageServe config remains `.env.stageserve`.
- Project `.env` remains application-owned.
- Config precedence remains: CLI flags, project `.env.stageserve`, shell environment, stack `.env.stageserve`, built-in defaults.
- New or changed controls:
  - `STAGESERVE_NO_TUI=1` disables guided TUI from the shell environment only. It is not read from project or stack `.env.stageserve`.
  - `NO_COLOR=1` disables color styling where applicable.
  - `--notui` and `--cli` are equivalent opt-out flags for the current invocation. They disable TUI behavior for bare `stage` and for subcommands that would otherwise use TUI by default.

### State, Isolation & Recovery

- The TUI must not introduce new runtime state beyond existing state and config files unless separately specified.
- The TUI may write project `.env.stageserve` only after preview and confirmation.
- The TUI may trigger existing lifecycle commands, but must use current rollback semantics.
- Cancellation before confirmation must leave no changes.
- Cancellation during actions must rely on existing context cancellation and rollback behavior.
- One project's guided failure must not alter unrelated project state.

### Documentation Surfaces

- `README.md`
- `docs/runtime-contract.md`
- new `docs/installer-onboarding.md` or equivalent active onboarding doc
- `.env.stageserve.example`
- command help strings in `cmd/stage/commands`
- spec 007 quickstart and validation artifacts

### Plain-Language Interaction Rules

- Easy-mode screens and text fallback describe goals and outcomes before commands.
- First-level labels use words a front-end developer or hobbyist can understand without StageServe internals.
- Direct command names remain available through "show commands" and direct CLI help.
- `attach` means "add this project to StageServe" in easy mode.
- `detach` means "remove this project from StageServe" in easy mode.
- `down` means "stop this project" in easy mode.
- Runtime details such as Docker, daemon, gateway, compose, container, registry, and state are advanced/troubleshooting terms unless they are the only actionable recovery clue.

## Requirements

### Functional Requirements

- **FR-001**: Bare `stage` MUST launch the guided TUI when run with no subcommand in an interactive terminal unless TUI is disabled.
- **FR-002**: Bare `stage` MUST NOT launch an interactive TUI when stdout or stdin is non-interactive.
- **FR-003**: `stage --help` MUST show standard CLI help and bypass the guided TUI.
- **FR-004**: Direct subcommands MUST bypass the root guided TUI and follow the final spec 007 CLI contract.
- **FR-005**: The guided TUI MUST use a shared non-UI next-action planner to decide the current situation and recommended actions.
- **FR-006**: The next-action planner MUST be terminal-verifiable through real `stage` invocations and may also have narrow package tests for deterministic decision rules.
- **FR-007**: The guided TUI MUST detect exactly these canonical situations for the first implementation: `machine_not_ready`, `project_missing_config`, `project_ready_to_run`, `project_running`, `project_down`, `drift_detected`, `not_project`, and `unknown_error`.
- **FR-008**: The guided TUI MUST present one primary recommended action and secondary actions for every supported situation.
- **FR-009**: Any file write from the guided TUI MUST preview the target path and relevant values before confirmation.
- **FR-010**: The guided TUI MUST write or update only `.env.stageserve` for user-editable StageServe config.
- **FR-011**: The guided TUI MUST expose the equivalent direct command for each action it offers.
- **FR-012**: JSON output modes MUST remain free of styled or human guidance text.
- **FR-013**: `NO_COLOR=1` MUST disable color styling in guided or projected output where color is otherwise used.
- **FR-014**: `STAGESERVE_NO_TUI=1` MUST disable the guided TUI.
- **FR-015**: The TUI MUST provide a visible quit/cancel path that does not mutate state before confirmation.
- **FR-016**: The TUI MUST route setup/init/doctor reporting through existing onboarding result semantics.
- **FR-017**: The TUI MUST route run/stop/status/logs through existing lifecycle/status/log command semantics rather than reimplementing them.
- **FR-018**: Operator docs MUST move Docker, compose, network, volume, and gateway implementation names out of the primary user path and into advanced/troubleshooting material.
- **FR-019**: The installer handoff MUST point to the bare guided `stage` path after this feature lands, while preserving explicit command guidance for non-interactive installs.
- **FR-020**: The spec MUST include validation for startup, status/inspection, teardown, and at least one failure/recovery path through both guided and direct command surfaces.
- **FR-021**: `stage init` MUST use the guided project-config form by default in interactive terminals, while `stage init --notui`, `stage init --cli`, and `stage init --json` keep their non-guided output contracts.
- **FR-022**: Direct command verification MUST include `stage attach` and `stage detach` as well as setup, init, doctor, up, status, logs, and down.
- **FR-023**: Advanced actions such as "show commands" MUST display StageServe command equivalents first, with implementation details only in advanced/troubleshooting copy.
- **FR-024**: StageServe MUST NOT expose a `--tui` flag in the final spec 007 command contract; TUI is the default easy-mode behavior and `--notui` / `--cli` are equivalent opt-outs.
- **FR-025**: Guided TUI labels MUST use easy-mode user language for first-level actions. Internal/direct command names such as `attach` and `detach` may appear in "show commands" or advanced views, but the primary labels MUST describe the user outcome, such as "add this project to StageServe", "remove this project from StageServe", "run this project", or "stop this project".
- **FR-026**: Text fallback, command help updated by this spec, README first-run copy, and installer handoff copy MUST follow the same plain-language rule as the guided TUI.

### Canonical Situation Semantics

- `machine_not_ready`: One or more setup-level prerequisites blocks normal operation.
- `project_missing_config`: The current directory can be treated as a project, but project `.env.stageserve` is absent.
- `project_ready_to_run`: Project config exists and there is no recorded active/down state for this project, or the recorded state does not require a specific recovery path.
- `project_running`: Project is recorded and runtime/status checks indicate it is active.
- `project_down`: Project has a retained StageServe state record marked down; the runtime is intentionally stopped but known.
- `drift_detected`: Recorded state, runtime state, DNS, gateway, or config disagree in a way that should be diagnosed before normal action.
- `not_project`: Current directory cannot be treated as a StageServe project root for guided project actions.
- `unknown_error`: Context collection or planning failed in a way that cannot be classified safely.

### Non-Functional Requirements

- **NFR-001**: Guided TUI startup SHOULD render its first screen within 500 ms excluding external Docker checks.
- **NFR-002**: The next-action planner SHOULD avoid long-running checks by default; expensive checks should be explicit or cached where safe.
- **NFR-003**: The TUI MUST remain keyboard-first and usable without mouse support.
- **NFR-004**: Text fallback MUST contain the same core semantic guidance as the TUI.
- **NFR-005**: Added abstractions MUST not duplicate lifecycle or config precedence logic.
- **NFR-006**: First-level TUI copy SHOULD be understandable without prior Docker, Compose, gateway, attach/detach, or StageServe state-model knowledge.

### Out Of Scope

- Replacing direct CLI subcommands.
- Adding a multi-project guided project switcher in spec 007. The first implementation stays scoped to the current directory while preserving existing multi-project runtime awareness.
- New Docker runtime behavior beyond what existing commands already perform.
- Automatic Docker Desktop installation.
- Framework-specific app repair or migration presets.
- A full graphical desktop app.
- Remote/cloud synchronization.
- Changing `.env.stageserve` precedence.

## Key Entities

- **Guided Session**: One invocation of bare `stage`, including detected context, selected action, confirmations, and outcome.
- **Next Action Plan**: Non-UI decision output describing situation, primary action, secondary actions, advanced actions, warnings, user-facing labels, and direct command equivalents.
- **Guided Action**: A user-selectable operation such as set up this computer, create project settings, run this project, add this project to StageServe, check status, view logs, stop this project, remove this project from StageServe, find issues, edit project settings, show commands, or advanced command guidance.
- **TUI Capability**: Runtime assessment of terminal suitability: stdin/stdout interactivity, color support, `NO_COLOR`, shell-only `STAGESERVE_NO_TUI`, and text fallback reason.
- **Config Preview**: The `.env.stageserve` target path and values shown before writing.
- **Recovery Path**: A user-facing next step for a non-ready or failed state, with direct command equivalent.

## Success Criteria

- **SC-001**: A first-time user can run bare `stage` and reach a clear next action without consulting README command order.
- **SC-002**: A project without `.env.stageserve` can be initialized through the guided path with preview and confirmation.
- **SC-003**: A configured stopped project can be started from the guided path using existing `stage up` semantics.
- **SC-004**: A running project can be inspected and stopped from the guided path without exposing Docker as the primary model.
- **SC-005**: Direct commands and JSON modes follow the final spec 007 CLI contract without opening the guided TUI unexpectedly.
- **SC-006**: Primary docs no longer require Docker/gateway implementation vocabulary before advanced/troubleshooting sections.
- **SC-007**: Terminal verification evidence covers planner states, root no-args behavior, TUI-disable behavior, direct command behavior, and JSON purity.
- **SC-008**: Easy-mode TUI and text fallback labels pass a plain-language review: primary actions describe user goals, and implementation terms appear only in show-commands or advanced/troubleshooting paths.

## Assumptions

- StageServe remains a terminal-first product.
- Go, Cobra, Bubble Tea, Lip Gloss, and Huh remain acceptable dependencies because spec 005 already introduced the Charm stack.
- The first guided TUI can be local-only and project-directory-based.
- The TUI may shell through existing command/domain seams for action execution as long as output and cancellation are handled coherently.
- The first version should optimize for clarity over visual richness.
- This spec run prioritizes terminal verification over TDD. Narrow automated tests are supporting evidence, not the primary development gate.
