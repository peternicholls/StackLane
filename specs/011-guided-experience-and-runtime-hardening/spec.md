# Feature Specification: Guided Experience And Runtime Hardening

**Feature Branch**: `011-guided-experience-and-runtime-hardening`  
**Created**: 2026-05-30  
**Last Updated**: 2026-06-05  
**Status**: Draft  
**Input**: Follow-up work from spec 007 recovery intent, code review findings (CR-01..CR-14), and field reports (FR-01..FR-06).

## Intent

Spec 011 is the execution-focused follow-up that closes the most important gaps left after spec 007.

This spec does not redefine the StageServe model. It hardens and completes it:

- `stage` remains the simple guided entrypoint.
- direct commands remain stable for power users and automation.
- StageServe-first language remains the default user-facing contract.

## Decision Record

The following decisions are fixed for spec 011:

- Merge-gate work is correctness and operator-safety first.
- Runtime asset provisioning is solved with a bundled installer artifact in 011.
- Embedded runtime assets in the binary are explicitly deferred (tracked separately).
- Missing runtime assets and lifecycle blockers must surface as StageServe-native guidance before raw Docker/Compose failure output.
- Guided TUI, text fallback, and direct command errors must converge on one semantic presentation language.
- Command-line password entry for MySQL password is removed; existing env-based config ownership remains.

## Problem Statement

The current codebase has a strong guided shell foundation but still fails users in four places:

1. Runtime prerequisites can fail late and leak implementation-shaped errors.
2. Some lifecycle paths remain fragile under partial failure or stale state.
3. Guided and direct surfaces still feel inconsistent in hierarchy and severity language.
4. Running-project and long-running interaction quality is still below the intended day-2 standard.

## Scope Model

Spec 011 has two delivery slices:

- Merge-gate slice: hardening and safety repairs required before merge.
- Follow-up slice: deeper guided UX expansion that can land after merge.

The authoritative task split is maintained in [tasks.md](./tasks.md).

## User Scenarios And Testing

### User Story 1 - Runtime Startup Fails Early And Clearly (P1)

As an operator, I want StageServe to detect missing runtime assets before trying to run Compose, so startup failures are actionable and not confusing.

**Independent test**: remove required compose assets and run `stage up`; StageServe must fail with StageServe-native guidance and next action.

**Acceptance scenarios**:

1. **Given** runtime assets are provisioned by the supported installer path, **When** I run `stage up`, **Then** startup does not require manual asset copying.
2. **Given** a required runtime compose file is missing, **When** I run `stage up`, **Then** StageServe fails before Compose execution and prints a StageServe remediation path.
3. **Given** the machine cannot satisfy readiness prerequisites, **When** I run bare `stage`, **Then** the planner enters machine-not-ready instead of ready-to-run.
4. **Given** machine readiness is satisfied but this directory has no StageServe project config, **When** bare `stage` opens, **Then** the default guided action is to set up this directory as a project and preview init before writing `.env.stageserve`.
5. **Given** project config exists and the project is stopped, **When** bare `stage` opens, **Then** the default guided action is to run this project rather than send the user to setup or recovery first.
6. **Given** project state indicates failure, drift, or an unknown blocker, **When** bare `stage` opens, **Then** the planner defaults to the least-invasive recovery action and surfaces `stage doctor`, `stage status`, and `stage logs` as ordered recovery affordances.

### User Story 2 - Lifecycle Errors Preserve Safety And Explain Recovery (P1)

As an operator, I want lifecycle failures to preserve safe state and explain what happened, so I can recover without guessing.

**Independent test**: induce attach/down/down-all/up failure variants and validate explicit remedies, partial-failure messaging, and safe gateway cleanup behavior.

**Acceptance scenarios**:

1. **Given** attach cannot read state registry, **When** attach is requested, **Then** the command fails explicitly and does not silently continue.
2. **Given** down-all partially fails, **When** StageServe reports the result, **Then** the output identifies partial effects and best-effort cleanup actions.
3. **Given** route/state races are possible during up, **When** routes are written, **Then** StageServe either refreshes registry first or documents/enforces a single-instance assumption.

### User Story 3 - Guided And Direct Surfaces Speak One Language (P1)

As a user, I want errors and next steps to look and read consistently across guided TUI, text fallback, and direct commands.

**Independent test**: compare equivalent failures across TUI/text/direct paths with and without color.

**Acceptance scenarios**:

1. **Given** semantic states (success/warn/error), **When** output is rendered, **Then** all surfaced modes use the same severity model.
2. **Given** `NO_COLOR=1`, **When** output is rendered, **Then** hierarchy remains intact without ANSI styling.
3. **Given** interactive direct-command failure in TTY, **When** the error is shown, **Then** the user is offered actionable recovery guidance.
4. **Given** a guided or direct failure surface, **When** recovery is rendered, **Then** primary labels use StageServe task language and direct command names appear only as secondary equivalents or advanced fallback.
5. **Given** non-interactive invocation or JSON mode, **When** guidance or failure output is produced, **Then** interactive copy does not pollute parseable output and documented JSON surfaces remain machine-safe.

### User Story 4 - Running-Project Screen Supports Day-2 Work (P1, follow-up)

As a user with a running project, I want to perform common day-2 actions from guided mode without dropping immediately into raw command troubleshooting.

**Independent test**: run a project and verify open/status/logs/stop flow plus advanced fallback commands.

**Acceptance scenarios**:

1. **Given** project is running, **When** bare `stage` opens, **Then** the default action is non-destructive and inspect/open actions are available first.
2. **Given** service metadata is sufficient, **When** I choose logs or restart, **Then** service selection is deterministic and explicit.
3. **Given** metadata is ambiguous, **When** I open advanced actions, **Then** direct command equivalents are shown as fallback.

### User Story 5 - Consequential Actions Feel Deliberate (P1, follow-up)

As a user, I want destructive actions to stand out and long-running actions to show progress, so I can trust what StageServe is doing.

**Independent test**: perform destructive confirmation and long-running lifecycle actions in TTY and narrow terminal widths.

**Acceptance scenarios**:

1. **Given** action is destructive, **When** confirmation is shown, **Then** it is visually elevated and explicitly states impact and non-impact.
2. **Given** lifecycle action takes time, **When** it runs in guided mode, **Then** spinner/progress appears promptly and cancel semantics remain clear.
3. **Given** narrow terminal dimensions, **When** these surfaces render, **Then** they remain usable and readable.

## Edge Cases

- Stack-home exists but one compose file is missing.
- State registry is unreadable or stale during attach/up.
- `STAGESERVE_NO_TUI`, non-TTY, and `NO_COLOR` combinations.
- `DownAll` succeeds for some projects and fails for others.
- Browser-launch helper unsupported in current OS/session.
- Service metadata missing for logs/restart selection.
- Users/scripts still attempt removed password CLI flag.

## Requirements

### Functional Requirements

- **FR-001**: Bare `stage` must reach machine-not-ready when prerequisites are missing, including runtime asset prerequisites.
- **FR-001a**: Bare `stage` must route deterministically among machine-not-ready, project-missing-config, ready-to-run, running-project, and recovery-needed states.
- **FR-001b**: In a project directory without `.env.stageserve`, bare `stage` must prefer `Set up this directory as a project` over machine setup unless machine prerequisites are missing.
- **FR-001c**: In recovery-needed states, bare `stage` must default to the next least-invasive recovery action and surface `stage doctor`, `stage status`, and `stage logs` as ordered recovery affordances.
- **FR-002**: StageServe must preflight required runtime compose files before compose startup paths.
- **FR-003**: Missing runtime assets must return StageServe-native remedies with clear next actions.
- **FR-004**: Installer/release path for 011 must ship and provision binary plus required runtime assets as a bundled artifact.
- **FR-005**: Attach must fail explicitly on state/registry read failure.
- **FR-006**: Lifecycle StepError wrapping in touched attach/up/down paths must include non-empty remedies.
- **FR-007**: DownAll must report partial failures and preserve best-effort gateway cleanup behavior.
- **FR-008**: Up route-write path must refresh state before route generation or explicitly enforce/document single-instance assumptions.
- **FR-009**: Guided, text fallback, and direct command error output must share one semantic severity model.
- **FR-010**: `NO_COLOR=1` behavior must preserve readable hierarchy in all touched output paths.
- **FR-011**: `STAGESERVE_NO_TUI` truthy parsing must be canonicalized and reused across command/guidance entry paths.
- **FR-012**: Root `--mysql-password` flag must be removed; `MYSQL_PASSWORD` remains supported only through existing shell/project/stack env surfaces.
- **FR-013**: Running-project guided surface must support open/status/logs/stop plus advanced command fallback (follow-up slice).
- **FR-014**: Service-scoped logs/restart actions must follow deterministic selection rules and ambiguity fallback behavior (follow-up slice).
- **FR-015**: Destructive confirmations must be visually elevated and include explicit impact language (follow-up slice).
- **FR-016**: Long-running guided lifecycle actions must run asynchronously with visible progress and stable cancel/quit semantics (follow-up slice).
- **FR-017**: Touched docs/help/config examples must remain aligned with implemented contract in each delivery slice.
- **FR-018**: Primary guided surfaces, README first-path sections, and installer onboarding docs must describe StageServe actions first; Docker, Compose, gateway, container, and attach/detach terminology may appear only in advanced, troubleshooting, or command-equivalent contexts.

### Non-Functional Requirements

- **NFR-001**: For long-running guided lifecycle actions, visible activity feedback should appear within 250 ms of confirmation.
- **NFR-002**: Hardening changes should prefer seam-level fixes over broad architectural rewrites.
- **NFR-003**: Non-TTY and JSON output safety must not regress.
- **NFR-003a**: Interactive recovery rendering must not alter non-TTY fallback behavior or `stage setup --json` and `stage doctor --json` output contracts.
- **NFR-004**: New guided actions should reuse existing lifecycle/status seams rather than ad hoc subprocess logic.

## Out Of Scope

- Multi-project dashboard redesign.
- New persistent config surfaces.
- Broad framework-specific action catalogs.
- Embedded-in-binary runtime assets (tracked separately).
- Replacing direct commands with TUI-only flows.

## Key Entities

- **Runtime Asset**: Required stack files needed before compose startup.
- **Recovery Surface**: Guided/text failure response with blocker + next safe actions.
- **Semantic Style Tokens**: Shared severity/presentation language across guided/text/direct output.
- **Service Summary**: Runtime metadata supporting service-scoped day-2 actions.
- **Elevated Confirmation Surface**: Prominent destructive-action confirmation UI.

## Validation Gates

Merge-gate completion requires:

- runtime asset preflight/remedy behavior validated,
- machine-not-ready reachability validated,
- root missing-config and ready-to-run routing validated,
- status/inspection path validated,
- teardown path validated,
- attach/down-all safety behavior validated,
- at least one failure/recovery path validated,
- password flag removal and env-based guidance validated,
- focused automated checks passing,
- merge-gate manual terminal checks recorded with explicit gaps when daemon-dependent scenarios cannot be exercised.

Follow-up completion requires:

- running-project day-2 action validation,
- destructive confirmation validation,
- long-running progress/cancel validation,
- final docs/help parity checks,
- deferred-work record updated.

## Success Criteria

- **SC-001**: Users on supported installs do not need manual runtime asset copying to run first startup.
- **SC-002**: Missing-asset failures are StageServe-native and actionable.
- **SC-003**: Bare `stage` no longer misclassifies missing prerequisite states as ready-to-run.
- **SC-004**: Lifecycle hardening paths (attach/down-all/up race assumptions) are explicit and test-covered.
- **SC-005**: Guided, text, and direct outputs use consistent severity language and no-color behavior.
- **SC-006**: Running-project/day-2 and long-running UX improvements are validated in real TTY follow-up checks.