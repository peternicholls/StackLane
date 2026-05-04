# Feature Specification: Guided TUI And Simple-First StageServe Interaction

**Feature Branch**: `007-harden-TUI-and-other-interactions`  
**Created**: 2026-05-04  
**Status**: Draft  
**Input**: User description: "Flesh out the recovery plan, specification, and planning documents for spec 007. Research similar guided TUI style interfaces, what has worked and what has not, and bring that into planning."

## User Scenarios & Testing

### User Story 1 - Start From Bare `stage` (Priority: P1)

A normal user runs `stage` from a terminal and gets a guided StageServe entrypoint that tells them the current situation and offers the next safe action.

**Why this priority**: This is the central gap. The original intention was that `stage` alone exposes the simple path; current behavior still requires the user to know subcommands first.

**Independent Test**: Run `stage` in an interactive terminal from a project directory and verify it opens a guided TUI that detects context and offers a primary action without requiring Docker knowledge.

**Acceptance Scenarios**:

1. **Given** stdout is a TTY and no subcommand is provided, **When** the operator runs `stage`, **Then** StageServe opens the guided TUI rather than only showing help.
2. **Given** stdout is not a TTY and no subcommand is provided, **When** the operator runs `stage`, **Then** StageServe prints concise next-step guidance and exits without blocking for input.
3. **Given** the operator runs `stage --help`, **When** help is requested explicitly, **Then** Cobra help is shown and the guided TUI is not launched.
4. **Given** the operator runs `stage up`, **When** a direct subcommand is provided, **Then** existing subcommand behavior remains unchanged.

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

**Independent Test**: With an attached/running project, run `stage` and verify the guided TUI offers status, logs, down, doctor, and advanced commands without exposing Docker as the primary model.

**Acceptance Scenarios**:

1. **Given** the current project is running, **When** the TUI starts, **Then** the primary screen shows project identity, route/status summary, and common actions.
2. **Given** the user chooses logs, **When** logs are followed, **Then** the UI provides a clear exit path and does not corrupt the terminal.
3. **Given** the user chooses stop, **When** the stop action is confirmed, **Then** `stage down` semantics are used and project data is preserved.
4. **Given** drift is detected, **When** the TUI presents recovery, **Then** the path starts with StageServe commands before advanced implementation detail.

---

### User Story 4 - Preserve Power-User And Automation Paths (Priority: P1)

A power user or script can bypass all TUI behavior and keep using direct commands, flags, and JSON output.

**Why this priority**: A guided first-level path must not get in the way of the existing shell-first contract.

**Independent Test**: Run direct commands and JSON modes under non-TTY conditions and verify behavior is unchanged.

**Acceptance Scenarios**:

1. **Given** `STAGESERVE_NO_TUI=1`, **When** the operator runs `stage`, **Then** StageServe prints plain guidance instead of opening the TUI.
2. **Given** `stage setup --json`, **When** the command runs, **Then** stdout remains a stable JSON envelope with no styled guidance.
3. **Given** `stage init --no-tui`, **When** the command runs, **Then** it uses plain text output.
4. **Given** an existing direct command such as `stage down --all`, **When** it runs, **Then** its behavior and exit semantics remain unchanged.

## Edge Cases

- stdout is a TTY but stdin is not interactive.
- terminal has limited color or reports an unknown `TERM`.
- user presses Ctrl-C before confirming any mutation.
- user presses Ctrl-C during a long-running action.
- `.env.stageserve` exists but is invalid.
- current directory is not a project root.
- machine setup requires privileged DNS work.
- Docker is missing or daemon is stopped.
- project is recorded as attached but runtime is missing.
- multiple projects are attached.
- `NO_COLOR`, `STAGESERVE_NO_TUI`, or `--no-tui` is set.
- command is run in CI or through redirected stdout.

## Operational Impact

### Ease Of Use & Workflow Impact

- Affected entry points: bare `stage`, `stage setup`, `stage init`, `stage doctor`, `stage up`, `stage status`, `stage logs`, `stage down`, help text, installer handoff, README first-run guidance.
- Backward compatibility: direct subcommands remain stable; bare `stage` changes from help-only to guided interactive behavior only in TTY contexts.
- Friction removed: users no longer need to know the exact first command or command sequence before StageServe can help them.
- Friction introduced: users who expect help from bare `stage` in a TTY need `stage --help`, `STAGESERVE_NO_TUI=1`, or `--no-tui`.

### Configuration & Precedence

- User-facing StageServe config remains `.env.stageserve`.
- Project `.env` remains application-owned.
- Config precedence remains: CLI flags, project `.env.stageserve`, shell environment, stack `.env.stageserve`, built-in defaults.
- New or changed controls:
  - `STAGESERVE_NO_TUI=1` disables guided TUI.
  - `NO_COLOR=1` disables color styling where applicable.
  - Optional global `--no-tui` for root/no-args path if it fits Cobra wiring cleanly.

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

## Requirements

### Functional Requirements

- **FR-001**: Bare `stage` MUST launch the guided TUI when run with no subcommand in an interactive terminal unless TUI is disabled.
- **FR-002**: Bare `stage` MUST NOT launch an interactive TUI when stdout or stdin is non-interactive.
- **FR-003**: `stage --help` MUST show standard CLI help and bypass the guided TUI.
- **FR-004**: Direct subcommands MUST retain existing behavior and exit semantics.
- **FR-005**: The guided TUI MUST use a shared non-UI next-action planner to decide the current situation and recommended actions.
- **FR-006**: The next-action planner MUST be terminal-verifiable through real `stage` invocations and may also have narrow package tests for deterministic decision rules.
- **FR-007**: The guided TUI MUST detect at least these situations: machine not ready, project missing config, project configured and stopped, project running, project drift/error, and non-project directory.
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

### Non-Functional Requirements

- **NFR-001**: Guided TUI startup SHOULD render its first screen within 500 ms excluding external Docker checks.
- **NFR-002**: The next-action planner SHOULD avoid long-running checks by default; expensive checks should be explicit or cached where safe.
- **NFR-003**: The TUI MUST remain keyboard-first and usable without mouse support.
- **NFR-004**: Text fallback MUST contain the same core semantic guidance as the TUI.
- **NFR-005**: Added abstractions MUST not duplicate lifecycle or config precedence logic.

### Out Of Scope

- Replacing direct CLI subcommands.
- New Docker runtime behavior beyond what existing commands already perform.
- Automatic Docker Desktop installation.
- Framework-specific app repair or migration presets.
- A full graphical desktop app.
- Remote/cloud synchronization.
- Changing `.env.stageserve` precedence.

## Key Entities

- **Guided Session**: One invocation of bare `stage`, including detected context, selected action, confirmations, and outcome.
- **Next Action Plan**: Non-UI decision output describing situation, primary action, secondary actions, advanced actions, warnings, and direct command equivalents.
- **Guided Action**: A user-selectable operation such as setup, init, up, status, logs, down, doctor, edit config, or advanced command guidance.
- **TUI Capability**: Runtime assessment of terminal suitability: interactive, color support, no-color, TUI disabled, text fallback.
- **Config Preview**: The `.env.stageserve` target path and values shown before writing.
- **Recovery Path**: A user-facing next step for a non-ready or failed state, with direct command equivalent.

## Success Criteria

- **SC-001**: A first-time user can run bare `stage` and reach a clear next action without consulting README command order.
- **SC-002**: A project without `.env.stageserve` can be initialized through the guided path with preview and confirmation.
- **SC-003**: A configured stopped project can be started from the guided path using existing `stage up` semantics.
- **SC-004**: A running project can be inspected and stopped from the guided path without exposing Docker as the primary model.
- **SC-005**: Direct commands and JSON modes remain compatible with their pre-007 behavior.
- **SC-006**: Primary docs no longer require Docker/gateway implementation vocabulary before advanced/troubleshooting sections.
- **SC-007**: Terminal verification evidence covers planner states, root no-args behavior, TUI-disable behavior, command compatibility, and JSON purity.

## Assumptions

- StageServe remains a terminal-first product.
- Go, Cobra, Bubble Tea, Lip Gloss, and Huh remain acceptable dependencies because spec 005 already introduced the Charm stack.
- The first guided TUI can be local-only and project-directory-based.
- The TUI may shell through existing command/domain seams for action execution as long as output and cancellation are handled coherently.
- The first version should optimize for clarity over visual richness.
- This spec run prioritizes terminal verification over TDD. Narrow automated tests are supporting evidence, not the primary development gate.
