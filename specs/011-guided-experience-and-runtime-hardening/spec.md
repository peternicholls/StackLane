# Feature Specification: Guided Experience And Runtime Hardening

**Feature Branch**: `011-guided-experience-and-runtime-hardening`  
**Created**: 2026-05-30  
**Status**: Draft  
**Input**: Follow-up planning derived from `/docs/code-review-phases-1-4.md`, including review findings CR-01 through CR-14 and field-reported issues FR-01 through FR-06.

## Decision Overrides

The following decisions are resolved for this follow-up spec and constrain implementation:

- This spec is a repair pass driven by code-review findings and field-reported breakage. Essential correctness and operator-safety fixes take priority over broader UX expansion.
- This spec extends spec 007 rather than reopening its core contract. Bare `stage`, plain-language guidance, and direct-command opt-outs remain the product model.
- Runtime asset repair in 011 uses a bundled release/install artifact that ships the binary and required stack assets together. Fully embedded binary-owned runtime assets are deferred as a long-term direction and tracked in [embedded-runtime-assets-plan.md](./embedded-runtime-assets-plan.md).
- Missing runtime assets and lifecycle blockers MUST surface as StageServe-native operator guidance before raw Compose or Docker errors.
- Guided TUI, text fallback, and direct command error output MUST share one semantic style language.
- The running-project guided surface MUST expose richer day-2 actions than the current minimal stop/detach set.
- Destructive confirmations MUST be visually elevated above normal page content.
- Long-running actions MUST provide visible progress or activity feedback inside the TUI.
- Secrets and root-level flag ownership cleanup are in scope when they materially affect operator safety.

## Problem Statement

Phases 1 through 4 of spec 007 delivered the core guided shell, planner, text fallback, and visual polish pass. The implementation now works, but the review and field feedback show three remaining gaps.

First, some runtime failures still break through as implementation-shaped errors instead of StageServe-shaped recovery. Missing Compose assets, unreachable machine-readiness states, ignored registry errors, and partial-failure paths all weaken confidence in the operator contract.

Second, the guided experience is still narrower than the product promise. The running-project screen exposes only a small subset of StageServe's day-2 capabilities, direct `stage up` failures do not open recovery paths, and confirmations are too easy to miss.

Third, the TUI and non-TUI surfaces do not yet feel like one product. Semantic colour choices are concentrated in `shell.go`, long-running actions block without feedback, and direct command output does not consistently reuse the same hierarchy or next-step language.

This follow-up spec turns those gaps into an implementation-ready hardening and expansion plan.

## User Scenarios And Testing

### User Story 1 - Runtime Failures Explain Themselves In StageServe Terms (Priority: P1)

As an operator, when StageServe cannot find required runtime assets or hits a lifecycle blocker, I get a clear StageServe error with an actionable next step instead of a raw Compose failure.

**Why this priority**: The current compose-path failure is a real broken user path.

**Independent Test**: Install StageServe through the supported bundled release/install path and verify the runtime compose assets are provisioned for the default stack-home layout; then remove or relocate those assets and verify `stage up` reports the missing asset with a StageServe-native remedy.

**Acceptance Scenarios**:

1. **Given** StageServe is installed through the supported bundled release/install path, **When** the operator runs `stage up` with the default stack-home layout, **Then** the required runtime compose assets are already present without a manual copy step.
2. **Given** `cfg.SharedFile` or `cfg.StackFile` is missing, **When** `stage up` starts, **Then** StageServe fails before invoking Compose and reports the missing file with a concrete remedy.
3. **Given** the machine is not ready, **When** the operator runs bare `stage`, **Then** the planner reaches `machine_not_ready` and shows the blocker instead of falling through to project actions.
4. **Given** `Attach()` cannot read the registry, **When** attach is requested, **Then** the command fails explicitly and does not silently write an empty gateway route set.

### User Story 2 - Direct Commands And Guided Output Feel Like One Product (Priority: P1)

As an operator, whether I use bare `stage`, text fallback, or direct commands, I see the same semantic colour language, status hierarchy, and next-step guidance.

**Why this priority**: The current product feels split between the guided shell and the direct command path.

**Independent Test**: Compare a guided error, a text fallback error, and a direct `stage up` failure and verify they use the same severity language, colour semantics when enabled, and equivalent next-step guidance.

**Acceptance Scenarios**:

1. **Given** colour output is enabled, **When** a warning, success, or error is rendered in either TUI or text/direct-command output, **Then** the same semantic colour tokens are used.
2. **Given** `NO_COLOR=1`, **When** any guided or direct output is rendered, **Then** styling falls back cleanly without breaking hierarchy.
3. **Given** `stage up` fails in a TTY, **When** the failure is shown, **Then** the user gets next-step options rather than a dead-end error line.

### User Story 3 - The Running-Project Screen Covers Real Day-2 Work (Priority: P1)

As a user working on a running project, I can open the site, inspect logs, stop safely, and reach advanced actions from the guided shell without first dropping to raw Docker commands.

**Why this priority**: The current running-project screen is too narrow for normal day-2 use.

**Independent Test**: Start a project, open bare `stage`, and verify the running-project screen offers open URL, logs, stop, and an advanced-actions path while keeping a non-destructive default.

**Acceptance Scenarios**:

1. **Given** a project is running, **When** the guided shell starts, **Then** the default action is non-destructive and the screen exposes `open browser`, `view logs`, and `stop this project`.
2. **Given** service metadata is available, **When** the user chooses logs or restart actions, **Then** StageServe scopes the action to a named service.
3. **Given** a user wants command-level control, **When** they open advanced actions, **Then** StageServe shows direct command equivalents before raw implementation details.

### User Story 4 - Destructive Actions And Long Operations Feel Deliberate (Priority: P1)

As a user, I can clearly see when I am being asked to confirm a consequential action, and I can tell when StageServe is working during long-running tasks.

**Why this priority**: The current confirmation layout and blocking execution model underuse the terminal UI.

**Independent Test**: Trigger `down` from the guided shell and run `up` on a slow path; verify confirmation is visually elevated and the long-running action shows spinner/progress feedback without freezing the interface.

**Acceptance Scenarios**:

1. **Given** the user selects a destructive action, **When** confirmation is requested, **Then** the confirmation is rendered as a bordered modal or equivalent elevated surface with explicit confirm/cancel hints.
2. **Given** a long-running action is in progress, **When** the TUI is active, **Then** the user sees a spinner or progress indicator and a stable cancel/quit story.
3. **Given** the terminal is narrow, **When** confirmation or loading state is shown, **Then** the screen remains usable and falls back safely.

### User Story 5 - Internal Seams Are Safer And Easier To Extend (Priority: P2)

As a maintainer, I can extend the guided shell and lifecycle behavior without fighting duplicated helpers, dead API branches, or bypassed abstractions.

**Why this priority**: Several review findings are not immediately user-facing but directly limit reliable follow-up work.

**Independent Test**: Focused tests cover the cleaned seams and the affected packages no longer carry dead fields, duplicated env parsing, or private state-file path logic.

**Acceptance Scenarios**:

1. **Given** compose startup options are reviewed, **When** the code is read or tested, **Then** there is no dead `Detach` branch that claims to support a false case it ignores.
2. **Given** guidance context needs project state, **When** it reads state, **Then** it uses the state store seam instead of reconstructing file paths privately.
3. **Given** `DownAll()` or concurrent `Up()` paths fail partially, **When** the operator gets the result, **Then** the error explains partial effects or the implementation refreshes shared state before writing gateway routes.

## Edge Cases

- `STACK_HOME` exists but is missing one or both compose files.
- `make install-dev` or installer copy drift leaves outdated runtime assets in place.
- `NO_COLOR=1`, `STAGESERVE_NO_TUI=1`, or non-TTY execution disables rich presentation.
- Terminal width is too narrow for a centered confirmation modal.
- Browser launch is unsupported or `open` / `xdg-open` is unavailable.
- Runtime service inspection succeeds for some services but not all.
- A registry read fails during attach or while refreshing routes in `Up()`.
- `DownAll()` stops some projects before one project fails.
- Two `stage up` invocations race on registry/gateway updates.
- Existing scripts or users still attempt `--mysql-password` after command-line password entry is removed.

## Operational Impact

### Operator Surface Impact

- Affected commands and flows: bare `stage`, `stage up`, `stage down`, `stage status`, `stage logs`, `stage doctor`, running-project guided actions, confirmation prompts, and direct command error output.
- Primary change: more StageServe-native recovery and richer day-2 actions, not a new product model.
- Direct commands remain automation-safe and must not open TUI in non-interactive contexts.

### Internal Impact

- Affected packages: `core/guidance`, `cmd/stage/commands`, `core/lifecycle`, `infra/compose`, `core/config`, and selected operator docs/help text.
- Main implementation pattern: reuse current domain seams rather than reimplementing runtime logic inside Bubble Tea models.

## Requirements

### Functional Requirements

- **FR-001**: Bare `stage` MUST be able to reach `machine_not_ready` by running or injecting the existing machine-readiness path into guided context collection.
- **FR-002**: StageServe MUST check for required compose asset files before invoking Compose and MUST return a StageServe-native remedy when they are missing.
- **FR-002a**: The supported install and release path MUST use a bundled artifact that delivers the StageServe binary and required runtime compose assets together, installing those assets into the default stack-home layout so a normal installed StageServe binary can run without a manual asset-copy step.
- **FR-003**: Direct lifecycle failures in interactive terminals MUST offer a recovery surface or equivalent next-step guidance rather than only printing a raw error line.
- **FR-004**: Guided TUI, text fallback, and direct command error rendering MUST share one semantic style token set for success, warning, and error states.
- **FR-005**: `NO_COLOR=1` MUST disable those shared style tokens cleanly in all relevant output paths.
- **FR-006**: The running-project guided surface MUST expose at least `open browser`, `view logs`, and `stop this project`, with a non-destructive default action.
- **FR-007**: The guided runtime summary MUST carry enough service metadata to support service-scoped logs and restart actions where supported.
- **FR-008**: Advanced actions in the guided shell MUST show StageServe command equivalents before implementation-level runtime details.
- **FR-009**: Destructive confirmations MUST state what will change and what will not change, and MUST be rendered with elevated visual prominence over normal surface content.
- **FR-010**: Long-running guided actions MUST execute without freezing the interface, MUST show spinner or progress feedback, and MUST define a stable cancel/quit path that preserves existing cancellation and rollback semantics.
- **FR-011**: `Attach()` MUST not ignore registry read failures.
- **FR-012**: Lifecycle `Wrap()` calls for attach/up/down error paths MUST provide non-empty remedy strings.
- **FR-013**: Compose orchestration code MUST not expose dead behavior such as a `Detach` option that is ignored.
- **FR-014**: Guidance context state loading MUST use a shared state-store seam rather than re-implementing file-path-based record loading privately.
- **FR-015**: `STAGESERVE_NO_TUI` truthy parsing MUST have one canonical implementation reused across command and guidance paths.
- **FR-016**: The `DownAll()` path MUST report partial failures clearly and SHOULD still attempt best-effort shared gateway cleanup.
- **FR-017**: The `Up()` gateway route write path MUST either refresh registry state before writing routes or explicitly enforce and document a single-instance assumption.
- **FR-018**: Operator-facing password handling MUST remove command-line password entry for sensitive values such as MySQL password. StageServe MUST accept MySQL password through environment-variable or supported config inputs only.
- **FR-019**: Docs and command help changed by the merge-gate repair work MUST land with that repair work. Broader docs/help for richer recovery and day-2 actions MUST land with the related follow-up features.
- **FR-020**: The implementation MUST preserve existing non-TTY, JSON, and direct-command safety guarantees from spec 007.

### Non-Functional Requirements

- **NFR-001**: The first visible guided response for a long-running action SHOULD appear within 250 ms of user confirmation.
- **NFR-002**: Runtime hardening changes SHOULD prefer focused seam cleanup over broad refactors.
- **NFR-003**: Added TUI feedback components SHOULD preserve plain-text fallback parity for primary outcomes and next steps.
- **NFR-004**: Any new action added to the running-project screen SHOULD reuse existing lifecycle or status seams rather than shelling out ad hoc.

## Out Of Scope

- A multi-project switcher or dashboard redesign.
- New persistent config surfaces beyond the current StageServe config model.
- Full framework-specific action catalogs for every stack in the first pass.
- Embedding runtime stack assets directly into the binary as part of the 011 repair; that long-term direction is tracked separately in [embedded-runtime-assets-plan.md](./embedded-runtime-assets-plan.md).
- Reworking spec 010 DNS extraction or unrelated release-pipeline work.
- Replacing direct CLI commands with TUI-only flows.

## Key Entities

- **Runtime Asset**: A file StageServe expects in stack home before lifecycle orchestration begins, including shared and project compose files.
- **Recovery Surface**: A guided or text-rendered post-failure screen that explains the blocker and offers the next safe action.
- **Shared Style Tokens**: The reusable semantic style definitions for success, warning, error, accent, muted text, confirmation, and footer hints.
- **Service Summary**: Service-level runtime metadata used by the running-project screen for logs, restart, and other scoped actions.
- **Confirmation Modal**: The elevated destructive-action confirmation surface rendered separately from normal decision content.

## Success Criteria

- **SC-001**: Supported StageServe installs use the bundled release/install artifact to provision the required runtime compose assets, and if those assets later drift missing StageServe fails with a StageServe-native remedy instead of a raw compose path error.
- **SC-002**: Bare `stage` reaches the machine-readiness blocker path in interactive terminals.
- **SC-003**: Guided and direct output share visibly consistent severity language and `NO_COLOR` behavior.
- **SC-004**: The running-project screen supports core day-2 tasks beyond stop/detach.
- **SC-005**: Destructive confirmations and long-running actions feel intentionally designed rather than incidental.
- **SC-006**: Focused tests cover the hardening seams surfaced by the review.