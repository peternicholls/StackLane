---
description: "Tasks for guided experience and runtime hardening"
---

# Tasks: Guided Experience And Runtime Hardening

**Input**: Design documents from `/specs/011-guided-experience-and-runtime-hardening/`  
**Prerequisites**: [spec.md](./spec.md), [plan.md](./plan.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/guided-runtime-hardening-contract.md](./contracts/guided-runtime-hardening-contract.md), [quickstart.md](./quickstart.md)

**Verification**: Focused Go tests and manual terminal validation for startup, status/inspection, teardown, and failure/recovery paths.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependency on incomplete tasks)
- **[Story]**: User story label for story-phase tasks only (`[US1]`, `[US2]`, ...)

## Terminal Design Operating Contract

All remaining guided, report, fallback, confirmation, and recovery tasks in this plan must apply the current terminal design system, not just make output functional. Tasks already marked complete before this contract was added are verified by T041-T042 and validation tasks rather than treated as retroactively design-approved.

Design source order:

1. `docs/design/terminal-visual-style-guide.md` for visual identity, semantic colour, component vocabulary, and review rules.
2. `docs/design/terminal-experience-style-guide.md` for surface families, screen grammar, default behavior, and text/TUI parity.
3. `docs/design/terminal-copy-style-guide.md` for labels, verdicts, remediation, warnings, confirmations, and advanced-only vocabulary.
4. `docs/design/guided-flow-map.md` for durable bare-`stage` situations and default actions.
5. `.github/instructions/terminal-*.instructions.md` for scoped implementation rules where their `applyTo` patterns match; the design docs above govern all terminal surfaces, including `core/guidance/**`.

Non-negotiable design rules for implementation tasks:

- Colour is semantic only: ready/success, needs action, error, primary command/action, supporting accent, structure, muted text, and dim evidence. Do not add decorative colour.
- Every screen follows the screen grammar: surface header, human verdict, key facts/visible defaults, one dominant focus section, secondary surface, and contextual footer help.
- First-level labels use user-goal language. Docker, compose, gateway, registry, state record, runtime, `attach`, and `detach` stay in `More…`, details, advanced, troubleshooting, logs, or exact command equivalents.
- Styled TUI and plain text fallback carry the same truth. Styling may add emphasis and rhythm, but not extra information.
- `More…` or an equivalent secondary surface is the boundary for direct commands, plain-text fallback, hidden paths, and advanced troubleshooting.
- Confirmation surfaces must name what changes, what does not change, and the concrete path, URL, or project affected.
- Validation evidence for terminal UX work must include hierarchy, semantic colour/no-colour behavior, narrow-width behavior where relevant, copy boundary checks, and exact command copy-pasteability.

## Phase 1: Setup (Spec Artifacts And Workflow Inputs)

**Goal**: Preserve planning-artifact alignment as maintenance context; these tasks are satisfied before code implementation begins.

- [x] T001 Align merge-gate and follow-up boundaries in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/spec.md and /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/plan.md.
- [x] T002 Update verification command matrix and known-gap handling in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/quickstart.md.
- [x] T003 Record task-to-requirement traceability notes in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/contracts/guided-runtime-hardening-contract.md.

## Phase 2: Foundational (Blocking Runtime And Contract Prerequisites)

**Goal**: Land shared runtime and contract foundations that block every user story.

- [x] T004a Define bundled runtime artifact manifest and extraction contract (binary + `stacks/20i` assets, target paths) in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/contracts/guided-runtime-hardening-contract.md and /Users/peternicholls/Dev/stageserve/docs/installer-onboarding.md.
- [x] T004b Implement installer bundle consumption and runtime asset provisioning in /Users/peternicholls/Dev/stageserve/install.sh.
- [x] T004c Add local/test bundle assembly and fixture validation path in /Users/peternicholls/Dev/stageserve/scripts/tests/.
- [x] T005 [P] Add runtime asset existence helper(s) for shared/project compose files in /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator.go.
- [x] T006 [P] Add onboarding/runtime required-file readiness checks in /Users/peternicholls/Dev/stageserve/core/onboarding/machine_readiness.go and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/readiness.go.
- [x] T007 Canonicalize `STAGESERVE_NO_TUI` truthy parsing via one shared helper in /Users/peternicholls/Dev/stageserve/cmd/stage/commands/root.go and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/onboarding_mode.go.
- [x] T008 Remove root `--mysql-password` flag wiring in /Users/peternicholls/Dev/stageserve/cmd/stage/commands/root.go and align password-source handling in /Users/peternicholls/Dev/stageserve/core/config/loader.go.
- [x] T009 Add/adjust focused regression tests for bundled assets and foundational seams in /Users/peternicholls/Dev/stageserve/cmd/stage/commands/setup_test.go, /Users/peternicholls/Dev/stageserve/cmd/stage/commands/doctor_test.go, /Users/peternicholls/Dev/stageserve/core/config/loader_test.go, and /Users/peternicholls/Dev/stageserve/scripts/tests/.

**Checkpoint**: Runtime prerequisites and config safety boundaries are enforceable across command paths.

## Phase 3: User Story 1 - Runtime Startup Fails Early And Clearly (Priority: P1)

**Goal**: Ensure startup preflight catches missing runtime assets and routes users to StageServe-native recovery.

**Independent Test**: Remove `cfg.SharedFile` or `cfg.StackFile`, run `stage up`, and verify StageServe-native failure/remedy before compose startup.

- [x] T010 [US1] Add pre-compose runtime asset preflight in /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator.go.
- [x] T011 [US1] Wire StageServe-native user-facing recovery output for preflight failures in /Users/peternicholls/Dev/stageserve/cmd/stage/commands/ and /Users/peternicholls/Dev/stageserve/core/guidance/.
- [x] T012 [US1] Update readiness copy/status mapping for runtime-asset blockers in /Users/peternicholls/Dev/stageserve/core/onboarding/projection_shared.go and /Users/peternicholls/Dev/stageserve/core/onboarding/projection_text.go.
- [x] T012a [US1] Preserve deterministic bare `stage` routing across project-missing-config, ready-to-run, and recovery-needed states in /Users/peternicholls/Dev/stageserve/core/guidance/planner.go and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/guidance_context.go.
- [x] T012b [US1] Set missing-config default action to `Set up this directory as a project` and encode least-invasive ordered recovery affordances in /Users/peternicholls/Dev/stageserve/core/guidance/planner.go and /Users/peternicholls/Dev/stageserve/core/guidance/shell.go.
- [x] T013 [P] [US1] Add lifecycle regression tests for missing shared/project compose files in /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator_test.go.
- [x] T014 [P] [US1] Add root routing reachability tests for missing runtime assets, missing config, stopped-project defaults, and ordered recovery states in /Users/peternicholls/Dev/stageserve/cmd/stage/commands/root_guidance_test.go.

**Checkpoint**: Startup failures are early, explicit, and actionable.

## Phase 4: User Story 2 - Lifecycle Errors Preserve Safety And Explain Recovery (Priority: P1)

**Goal**: Close silent lifecycle failure modes and improve partial-failure clarity.

**Independent Test**: Induce attach registry-read failure and partial `DownAll` failure; verify explicit remedies and safe-state messaging.

- [x] T015 [US2] Fail fast on registry read errors in attach path in /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator.go.
- [x] T016 [US2] Fill missing/empty lifecycle remedies in touched wrap calls in /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator.go.
- [x] T017 [US2] Improve `DownAll` partial-failure reporting and cleanup behavior in /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator.go.
- [x] T018 [US2] Refresh registry state before route writes in `Up()` or enforce/document single-instance expectation in /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator.go and /Users/peternicholls/Dev/stageserve/docs/runtime-contract.md.
- [x] T019 [P] [US2] Add attach/downall/up safety regression tests in /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator_test.go.

**Checkpoint**: Lifecycle errors preserve safe behavior and communicate real recovery steps.

## Phase 5: User Story 3 - Guided And Direct Surfaces Speak One Language (Priority: P1)

**Goal**: Unify severity, hierarchy, and next-step language across guided, text fallback, and direct command output.

**Independent Test**: Compare equivalent failures in guided TUI, plain-text fallback, and direct command output with and without `NO_COLOR=1`.

- [x] T020 [US3] Extract shared semantic style tokens from /Users/peternicholls/Dev/stageserve/core/guidance/shell.go into a reusable styles module under /Users/peternicholls/Dev/stageserve/core/guidance/.
- [x] T021 [US3] Apply shared style/severity model to text renderer in /Users/peternicholls/Dev/stageserve/core/guidance/text.go.
- [x] T022 [US3] Add direct-command StepError renderer and wire it in command paths under /Users/peternicholls/Dev/stageserve/cmd/stage/commands/.
- [x] T023 [US3] Add `NO_COLOR` parity checks for guided/text/direct paths in /Users/peternicholls/Dev/stageserve/core/guidance/planner_test.go and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/root_guidance_test.go.
- [x] T023a [US3] Add non-TTY fallback and JSON purity regression coverage for guided recovery and direct command failures in /Users/peternicholls/Dev/stageserve/cmd/stage/commands/setup_test.go, /Users/peternicholls/Dev/stageserve/cmd/stage/commands/doctor_test.go, and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/root_guidance_test.go; assert parseable outputs are free of TUI-only copy, ANSI, hidden shortcut hints, and advanced-first vocabulary.
- [x] T024 [P] [US3] Add command-surface consistency tests in /Users/peternicholls/Dev/stageserve/cmd/stage/commands/status_test.go and /Users/peternicholls/Dev/stageserve/observability/status/status_test.go; assert report/guided/direct surfaces preserve the same human verdict, next action, semantic severity, and user-goal labels.

**Checkpoint**: Users see one coherent StageServe error/recovery language regardless of entry path.

## Phase 6: User Story 4 - Running-Project Screen Supports Day-2 Work (Priority: P1, Follow-up)

**Goal**: Expand guided running-project actions for normal day-2 workflows.

**Independent Test**: With a running project, verify open/status/logs/stop and advanced fallback flows from bare `stage`.

- [x] T025 [US4] Extend runtime summary with service metadata in /Users/peternicholls/Dev/stageserve/core/guidance/types.go and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/guidance_context.go; expose only normalized service facts needed for user-goal actions, not raw container/runtime dumps.
- [x] T026 [US4] Implement deterministic service selection rules for logs/restart in /Users/peternicholls/Dev/stageserve/core/guidance/planner.go and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/tui.go; keep ambiguous service detail in a secondary selector or advanced fallback.
- [x] T027 [US4] Add guided `open_browser` action and launcher handling in /Users/peternicholls/Dev/stageserve/cmd/stage/commands/tui.go; label it as the user goal `Open in browser` and surface unsupported launcher states as calm recovery copy.
- [x] T028 [US4] Add guided logs viewport flow in /Users/peternicholls/Dev/stageserve/core/guidance/shell.go and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/tui.go; preserve log stream text, add only project identity, stable exit hints, and return-to-screen consistency.
- [x] T029 [US4] Add guided restart-service action reusing lifecycle/compose seams in /Users/peternicholls/Dev/stageserve/cmd/stage/commands/tui.go and /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator.go; require explicit service identity and confirmation copy before mutation.
- [x] T030 [US4] Add `More…`/advanced-actions fallback surface and reorder actions to non-destructive defaults in /Users/peternicholls/Dev/stageserve/core/guidance/planner.go and /Users/peternicholls/Dev/stageserve/core/guidance/shell.go; keep direct commands, plain-text fallback, hidden paths, and advanced troubleshooting secondary.
- [x] T031 [P] [US4] Add running-project day-2 behavior tests in /Users/peternicholls/Dev/stageserve/core/guidance/planner_test.go and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/root_guidance_test.go; assert default action is non-destructive, visible defaults precede actions, and implementation vocabulary does not lead the screen.

**Checkpoint**: Running-project guided flow supports core day-2 needs without forcing raw command troubleshooting.

## Phase 7: User Story 5 - Consequential Actions Feel Deliberate (Priority: P1, Follow-up)

**Goal**: Make destructive actions and long-running flows visibly trustworthy.

**Independent Test**: Validate elevated confirmations, async progress feedback, and cancellation semantics in real TTY paths.

- [x] T032 [US5] Replace inline destructive confirmations with elevated modal-style confirmation surfaces in /Users/peternicholls/Dev/stageserve/core/guidance/shell.go; follow the confirmation-sheet grammar and default rare destructive choices toward the non-destructive option where applicable.
- [x] T033 [US5] Add explicit impact/non-impact confirmation copy for destructive actions in /Users/peternicholls/Dev/stageserve/core/guidance/shell.go and /Users/peternicholls/Dev/stageserve/core/guidance/text.go; name the exact path, URL, project, or service affected.
- [x] T034 [US5] Move long-running lifecycle guided actions to async `tea.Cmd` handling in /Users/peternicholls/Dev/stageserve/core/guidance/shell.go and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/tui.go; loading surfaces must name the current user-goal step instead of generic `Working…` copy.
- [x] T035 [US5] Implement cancel/quit semantics for async lifecycle actions in /Users/peternicholls/Dev/stageserve/core/guidance/shell.go and /Users/peternicholls/Dev/stageserve/core/lifecycle/orchestrator.go; explain what cancellation can and cannot stop in user language.
- [x] T036 [US5] Add spinner/progress messaging hooks for async lifecycle actions in /Users/peternicholls/Dev/stageserve/core/guidance/shell.go; use semantic status text and avoid decorative animation-only feedback.
- [x] T036a [US5] Add explicit latency verification for first visible progress feedback (<= 250 ms) in /Users/peternicholls/Dev/stageserve/core/guidance/ tests and record criteria in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/quickstart.md.
- [x] T037 [P] [US5] Add confirmation/progress narrow-width and keymap tests in /Users/peternicholls/Dev/stageserve/core/guidance/planner_test.go; assert target values remain visible, footer help is contextual, and text does not rely on colour to communicate risk.

**Checkpoint**: Destructive and long-running interactions are explicit, stable, and observable.

## Phase 8: TUI Design Polish (Follow-up)

**Goal**: Apply a dedicated visual polish pass after behavior is working.

- [ ] T038 Refine spacing, section rhythm, and hierarchy in guided shell views in /Users/peternicholls/Dev/stageserve/core/guidance/shell.go against `docs/design/terminal-visual-style-guide.md`; enforce surface header, verdict-first order, visible defaults, one dominant focus section, secondary `More…`, and contextual footer help.
- [ ] T039 Normalize colour semantics, component reuse, and fallback rendering across guided states in /Users/peternicholls/Dev/stageserve/core/guidance/shell.go and /Users/peternicholls/Dev/stageserve/core/guidance/text.go; map every colour to a named semantic role and remove any decorative or one-off styling.
- [ ] T040 Validate responsive behavior at narrow/normal/wide terminal widths and record evidence in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/quickstart.md; include screenshots or captured text for colour and `NO_COLOR=1`, plus notes on wrapping, alignment, and footer discoverability.

## Phase 9: Polish & Cross-Cutting Validation

**Goal**: Close docs/help parity and execute merge-gate/follow-up validation gates.

- [ ] T041 Audit and rewrite merge-gate docs/help surfaces so StageServe actions remain primary and Docker/gateway details move to advanced or troubleshooting sections in /Users/peternicholls/Dev/stageserve/README.md, /Users/peternicholls/Dev/stageserve/docs/installer-onboarding.md, and /Users/peternicholls/Dev/stageserve/docs/runtime-contract.md; preserve the design source hierarchy and avoid duplicating stale prototype language.
- [ ] T042 Audit guided labels and direct-command recovery rendering so friendly StageServe labels remain primary and command equivalents remain secondary in /Users/peternicholls/Dev/stageserve/core/guidance/ and /Users/peternicholls/Dev/stageserve/cmd/stage/commands/; verify advanced-only terms do not lead first-level copy and update pattern examples if a new reusable screen pattern is introduced.
- [ ] T043 Update sensitive-value guidance in /Users/peternicholls/Dev/stageserve/.env.stageserve.example.
- [ ] T044 Run merge-gate focused startup and root-routing validation command set and record outcomes in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/quickstart.md.
- [ ] T045 Run merge-gate focused status/inspection and teardown validation command set and record outcomes in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/quickstart.md.
- [ ] T046 Run merge-gate manual failure/recovery, non-TTY, and JSON-safety validation set and record outcomes in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/quickstart.md.
- [ ] T047 Update follow-up docs/help for recovery/day-2 UX expansion in /Users/peternicholls/Dev/stageserve/README.md and /Users/peternicholls/Dev/stageserve/docs/runtime-contract.md.
- [ ] T048 Run follow-up focused automated validation and record outcomes in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/quickstart.md.
- [ ] T049 Run follow-up manual TTY validation and record outcomes in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/quickstart.md.
- [ ] T050 Record intentionally deferred work and residual risks in /Users/peternicholls/Dev/stageserve/specs/011-guided-experience-and-runtime-hardening/quickstart.md.

## Dependencies & Execution Order

### Phase Dependencies

- Setup (Phase 1): start immediately.
- Foundational (Phase 2): blocks all story phases.
- User stories (Phases 3-7): depend on Foundational completion.
- TUI polish (Phase 8): depends on completion of behavior phases (especially Phases 6-7).
- Final polish/validation (Phase 9): depends on all desired implementation phases.

### User Story Dependencies

- US1 (Phase 3): first merge-gate runtime behavior slice.
- US2 (Phase 4): depends on US1 runtime seams for coherent lifecycle recovery.
- US3 (Phase 5): can proceed after foundational, but best done after US1/US2 error paths exist.
- US4 (Phase 6, follow-up): depends on runtime and guidance surfaces from US1-US3.
- US5 (Phase 7, follow-up): depends on US4 action wiring and shared shell surfaces.

### Within Each User Story

- Land seam/model changes first, then action/render wiring, then focused validations.
- Keep each story independently testable at its checkpoint.

## Parallel Execution Examples

### User Story 1

```bash
# Parallel checks after implementation:
Task: T013 core/lifecycle missing-asset tests
Task: T014 root routing reachability tests across missing assets, missing config, stopped-project defaults, and ordered recovery states
```

### User Story 3

```bash
# Parallel implementation slices:
Task: T020 extract shared style tokens
Task: T024 add command-surface consistency tests
```

### User Story 5

```bash
# Parallel hardening after async conversion starts:
Task: T036 add spinner/progress messaging
Task: T037 add narrow-width/keymap tests
```

## Implementation Strategy

### MVP First (Merge Gate)

1. Complete Phase 1 and Phase 2.
2. Complete US1, US2, and US3 (Phases 3-5).
3. Complete merge-gate docs/validation tasks (T041-T046 in Phase 9).
4. Stop and verify merge-gate readiness.

### Incremental Follow-up

1. Deliver US4 running-project day-2 expansion.
2. Deliver US5 confirmation/progress UX.
3. Run dedicated TUI polish phase.
4. Close final docs/validation/deferred-work tasks.

### Team Parallelization

1. One developer handles runtime/lifecycle hardening (US1/US2).
2. One developer handles presentation convergence (US3).
3. One developer handles guided day-2 and async UX follow-up (US4/US5) after merge-gate completion.
