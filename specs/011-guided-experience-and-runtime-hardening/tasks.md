---
description: "Tasks for guided experience and runtime hardening follow-up work"
---

# Tasks: Guided Experience And Runtime Hardening

**Input**: Design documents from `/specs/011-guided-experience-and-runtime-hardening/`  
**Prerequisites**: [spec.md](./spec.md), [plan.md](./plan.md)

**Verification**: Focused Go tests for guidance, commands, lifecycle, compose, and config packages, plus real-terminal checks for missing assets, recovery, running-project actions, confirmation visibility, and long-running feedback.

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel with other tasks in different files.

## Delivery Split

### Must Fix Before Merge

These tasks are the release gate for the current branch because they address active broken paths, silent correctness failures, or operator-safety issues already identified in review and field testing.

- [ ] T001a Implement a bundled release/install artifact that ships the StageServe binary and required runtime compose assets together, provisioning those assets into the default stack-home layout.
- [ ] T001 Add pre-flight existence checks for `cfg.SharedFile` and `cfg.StackFile` before compose startup paths run.
- [ ] T002 Return StageServe-native `StepError` remedies for missing runtime assets, including a supported recovery command or install path.
- [ ] T003 Add or update focused tests covering missing compose assets.
- [ ] T004 Wire machine-readiness collection into guided root context gathering.
- [ ] T005 Add focused tests proving bare `stage` can reach `machine_not_ready`.
- [ ] T006 Make `Attach()` fail explicitly on registry read errors instead of ignoring them.
- [ ] T007 Fill empty remedy strings in `Attach()` and other touched lifecycle error wraps.
- [ ] T011 Remove command-line MySQL password entry and update config loading/help so MySQL password is accepted through environment-variable or supported config inputs only.
- [ ] T041 Update operator-facing docs and command help for merge-gate repair changes: runtime asset provisioning/remedies and env-first password guidance.
- [ ] T042 Update `.env.stageserve.example` or other active config docs for sensitive-value guidance where needed.
- [ ] T043 Run merge-gate focused validation: `go test ./core/guidance ./cmd/stage/commands ./core/lifecycle ./infra/compose ./core/config`.
- [ ] T044 Run merge-gate manual terminal validation for install/runtime asset provisioning, missing-assets remedies, machine-not-ready reachability, and attach/password handling changes.

### Follow-up After Merge

These tasks remain in scope for the follow-up spec, but they do not block merging once the must-fix set above is complete.

- [ ] T008 Remove the dead `UpOptions.Detach` branch and update compose callers/tests accordingly.
- [ ] T009 Replace private `readRecord()` path logic with a shared state-store seam.
- [ ] T010 Deduplicate `STAGESERVE_NO_TUI` truthy parsing and route both command and guidance paths through the same helper.
- [ ] T012 Improve `DownAll()` partial-failure reporting and keep best-effort gateway cleanup.
- [ ] T013 Refresh registry state before writing shared gateway routes in `Up()`, or explicitly lock/document the single-instance assumption.
- [ ] T014 Land small hardening hygiene in touched files: rename shadowing locals, remove dead footer branches, and restore missing step comments.
- [ ] T015 Extract semantic style tokens from `core/guidance/shell.go` into a shared guidance styles module.
- [ ] T016 Update the guided shell to consume the extracted styles module.
- [ ] T017 Update `core/guidance/text.go` to render severity and next-step output using the shared styles.
- [ ] T018 Add a reusable direct-command / `StepError` renderer that uses the same severity language and style tokens.
- [ ] T019 Apply the shared renderer to relevant command paths in `cmd/stage/commands`.
- [ ] T020 Add focused checks for `NO_COLOR` parity and plain-text-safe fallback behavior.
- [ ] T021 Add typed failure-plan inputs or new planner situations for start/gateway/setup failure states.
- [ ] T022 Implement `runFailureRecovery` or equivalent recovery entrypoint for `stage up` failures in TTY mode.
- [ ] T023 Ensure non-TTY failure output includes the same blocker explanation, next step, and direct command equivalent.
- [ ] T024 Add recovery decision items such as retry, run doctor, view logs, and show full error where appropriate.
- [ ] T025 Add focused tests for failure-plan rendering and `stage up` recovery routing.
- [ ] T026 Extend runtime summary types with service metadata needed for service-scoped actions.
- [ ] T027 Populate service metadata during guided runtime status collection.
- [ ] T028 Add an `open_browser` guided action and cross-platform launcher helper.
- [ ] T029 Add a guided logs surface using a viewport-style terminal component.
- [ ] T030 Add a guided restart-service action that reuses the compose/lifecycle seam.
- [ ] T031 Add an advanced-actions surface that shows direct command equivalents first.
- [ ] T032 Reorder running-project actions so inspect/open actions come before mutating actions, with a non-destructive default.
- [ ] T033 Run real-terminal validation for running-project day-2 flows.
- [ ] T034 Replace inline destructive confirmations with an elevated bordered confirmation modal or equivalent surface.
- [ ] T035 Add warning/destructive confirmation variants with explicit confirm/cancel hints inside the elevated surface.
- [ ] T036 Move guided action execution onto async `tea.Cmd` handling so the event loop stays responsive.
- [ ] T036a Define and implement cancel/quit semantics for async guided actions, preserving context cancellation and rollback expectations.
- [ ] T037 Add spinner-based loading state for long-running guided actions.
- [ ] T038 Add progress reporting where lifecycle step boundaries already exist and can be surfaced safely.
- [ ] T039 Add help/keymap affordances and narrow-width behavior checks for confirmation and loading states.
- [ ] T040 Run follow-up terminal validation for confirmation visibility, cancellation, running-project day-2 actions, and long-running feedback in a real TTY.
- [ ] T041a Update operator-facing docs and command help for richer recovery guidance and running-project/day-2 actions after those features land.
- [ ] T043a Run full focused validation after the follow-up implementation phases land.
- [ ] T044a Run full follow-up manual terminal validation for running-project actions, destructive confirmation, and long-running progress.
- [ ] T045 Record any intentionally deferred follow-up work such as broader framework-specific utility actions or multi-project controls.

## Phase 1: Runtime Hardening

**Purpose**: Fix the reviewed runtime and operator-safety gaps before expanding the shell further.

- [ ] T001a Implement a bundled release/install artifact that ships the StageServe binary and required runtime compose assets together, provisioning those assets into the default stack-home layout.
- [ ] T001 Add pre-flight existence checks for `cfg.SharedFile` and `cfg.StackFile` before compose startup paths run.
- [ ] T002 Return StageServe-native `StepError` remedies for missing runtime assets, including a supported recovery command or install path.
- [ ] T003 Add or update focused tests covering missing compose assets.
- [ ] T004 Wire machine-readiness collection into guided root context gathering.
- [ ] T005 Add focused tests proving bare `stage` can reach `machine_not_ready`.
- [ ] T006 Make `Attach()` fail explicitly on registry read errors instead of ignoring them.
- [ ] T007 Fill empty remedy strings in `Attach()` and other touched lifecycle error wraps.
- [ ] T008 Remove the dead `UpOptions.Detach` branch and update compose callers/tests accordingly.
- [ ] T009 Replace private `readRecord()` path logic with a shared state-store seam.
- [ ] T010 Deduplicate `STAGESERVE_NO_TUI` truthy parsing and route both command and guidance paths through the same helper.
- [ ] T011 Remove command-line MySQL password entry and update config loading/help so MySQL password is accepted through environment-variable or supported config inputs only.
- [ ] T012 Improve `DownAll()` partial-failure reporting and keep best-effort gateway cleanup.
- [ ] T013 Refresh registry state before writing shared gateway routes in `Up()`, or explicitly lock/document the single-instance assumption.
- [ ] T014 Land small hardening hygiene in touched files: rename shadowing locals, remove dead footer branches, and restore missing step comments.

**Checkpoint**: Core runtime failures surface as StageServe-native errors and the reviewed safety seams are closed.

## Phase 2: Shared Presentation Layer

**Purpose**: Make guided and direct output share one visual and severity language.

- [ ] T015 Extract semantic style tokens from `core/guidance/shell.go` into a shared guidance styles module.
- [ ] T016 Update the guided shell to consume the extracted styles module.
- [ ] T017 Update `core/guidance/text.go` to render severity and next-step output using the shared styles.
- [ ] T018 Add a reusable direct-command / `StepError` renderer that uses the same severity language and style tokens.
- [ ] T019 Apply the shared renderer to relevant command paths in `cmd/stage/commands`.
- [ ] T020 Add focused checks for `NO_COLOR` parity and plain-text-safe fallback behavior.

**Checkpoint**: TUI, text fallback, and direct command errors present a consistent StageServe voice.

## Phase 3: Recovery Surfaces

**Purpose**: Turn direct command failures into guided next steps instead of dead-end errors.

- [ ] T021 Add typed failure-plan inputs or new planner situations for start/gateway/setup failure states.
- [ ] T022 Implement `runFailureRecovery` or equivalent recovery entrypoint for `stage up` failures in TTY mode.
- [ ] T023 Ensure non-TTY failure output includes the same blocker explanation, next step, and direct command equivalent.
- [ ] T024 Add recovery decision items such as retry, run doctor, view logs, and show full error where appropriate.
- [ ] T025 Add focused tests for failure-plan rendering and `stage up` recovery routing.

**Checkpoint**: `stage up` and similar failure paths offer recovery instead of a dead end.

## Phase 4: Running-Project Action Expansion

**Purpose**: Expand the guided shell from minimal lifecycle controls to real day-2 usage.

- [ ] T026 Extend runtime summary types with service metadata needed for service-scoped actions.
- [ ] T027 Populate service metadata during guided runtime status collection.
- [ ] T028 Add an `open_browser` guided action and cross-platform launcher helper.
- [ ] T029 Add a guided logs surface using a viewport-style terminal component.
- [ ] T030 Add a guided restart-service action that reuses the compose/lifecycle seam.
- [ ] T031 Add an advanced-actions surface that shows direct command equivalents first.
- [ ] T032 Reorder running-project actions so inspect/open actions come before mutating actions, with a non-destructive default.
- [ ] T033 Run real-terminal validation for running-project day-2 flows.

**Checkpoint**: The running-project screen supports normal day-2 tasks without dropping users into raw Docker recovery.

## Phase 5: Confirmation And Progress UX

**Purpose**: Make consequential actions easier to notice and long operations easier to trust.

- [ ] T034 Replace inline destructive confirmations with an elevated bordered confirmation modal or equivalent surface.
- [ ] T035 Add warning/destructive confirmation variants with explicit confirm/cancel hints inside the elevated surface.
- [ ] T036 Move guided action execution onto async `tea.Cmd` handling so the event loop stays responsive.
- [ ] T036a Define and implement cancel/quit semantics for async guided actions, preserving context cancellation and rollback expectations.
- [ ] T037 Add spinner-based loading state for long-running guided actions.
- [ ] T038 Add progress reporting where lifecycle step boundaries already exist and can be surfaced safely.
- [ ] T039 Add help/keymap affordances and narrow-width behavior checks for confirmation and loading states.
- [ ] T040 Run follow-up terminal validation for confirmation visibility, cancellation, running-project day-2 actions, and long-running feedback in a real TTY.

**Checkpoint**: Confirmations stand out and long-running actions visibly progress.

## Phase 6: Docs, Help, And Final Validation

**Purpose**: Align the operator contract and lock in evidence for the follow-up work.

- [ ] T041 Update operator-facing docs and command help for merge-gate repair changes: runtime asset provisioning/remedies and env-first password guidance.
- [ ] T041a Update operator-facing docs and command help for richer recovery guidance and running-project/day-2 actions after those features land.
- [ ] T042 Update `.env.stageserve.example` or other active config docs for sensitive-value guidance where needed.
- [ ] T043 Run merge-gate focused validation: `go test ./core/guidance ./cmd/stage/commands ./core/lifecycle ./infra/compose ./core/config`.
- [ ] T043a Run full focused validation after the follow-up implementation phases land.
- [ ] T044 Run merge-gate manual terminal validation for install/runtime asset provisioning, missing-assets remedies, machine-not-ready reachability, and attach/password handling changes.
- [ ] T044a Run full follow-up manual terminal validation for running-project actions, destructive confirmation, and long-running progress.
- [ ] T045 Record any intentionally deferred follow-up work such as broader framework-specific utility actions or multi-project controls.

## Dependencies

- Phase 1 should land before later UX expansion so new surfaces rely on correct runtime and error semantics.
- Phase 2 should finish before broader recovery and day-2 work so new surfaces reuse shared presentation rather than creating more drift.
- Phase 3 depends on Phase 1 runtime fixes and benefits from Phase 2 shared rendering.
- Phase 4 depends on the planner/runtime-summary work from earlier phases.
- Phase 5 depends on stable action wiring from Phases 3 and 4.
- Final docs and validation depend on all implementation phases.

## Notes

- Keep the simple-first product model from spec 007 intact.
- Prefer existing StageServe seams over new ad hoc subprocess logic.
- Do not add new persistent user-facing config unless separately specified.
- Keep direct command behavior automation-safe while improving interactive recovery.
- Treat the review findings as implementation constraints, not optional polish.
- The merge gate validates only the essential repair subset; the broader UX follow-up remains planned in the same spec.
- The long-term embedded runtime-asset direction is deferred to [embedded-runtime-assets-plan.md](./embedded-runtime-assets-plan.md) after the bundled-artifact repair ships.