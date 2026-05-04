---
description: "Tasks for guided TUI and simple-first StageServe interaction"
---

# Tasks: Guided TUI And Simple-First StageServe Interaction

**Input**: Design documents from `/specs/007-harden-TUI-and-other-interactions/`  
**Prerequisites**: [spec.md](./spec.md), [plan.md](./plan.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/guided-tui-contract.md](./contracts/guided-tui-contract.md)

**Verification**: Terminal verification is primary for this spec run. Focused automated tests are supporting checks for pure planner logic, JSON parsing, and regression safety, but work should be driven by real `stage` invocations.

**Operational Verification**: Validate TTY, non-TTY, TUI-disabled, first-run, running-project, teardown, failure/recovery, and direct-command paths.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel with other tasks in different files.
- **[Story]**: User story label from [spec.md](./spec.md).

## Phase 1: Setup And Contract

**Purpose**: Lock the guided interaction contract before implementation.

- [ ] T001 [P] Add or update spec 007 contract references in `specs/007-harden-TUI-and-other-interactions/contracts/guided-tui-contract.md`.
- [ ] T002 [P] Add terminal verification checklist scaffolding to `specs/007-harden-TUI-and-other-interactions/quickstart.md`.
- [ ] T003 [P] Confirm `go.mod` has only the Charm dependencies already introduced by spec 005; do not add new UI dependencies.
- [ ] T004 Add a short implementation note in `docs/runtime-contract.md` that direct command semantics remain authoritative under the guided TUI.

**Checkpoint**: Contract and validation surfaces agree on root no-args behavior.

## Phase 2: Foundational Planner

**Purpose**: Build the non-UI decision layer that the TUI and text fallback will use.

- [ ] T005 [P] Create `core/guidance/types.go` with `TUICapability`, `GuidedContext`, `NextActionPlan`, and `GuidedAction`.
- [ ] T006 [P] Add terminal planner-inspection command or debug output so `machine_not_ready`, `project_missing_config`, `project_ready_to_run`, `project_running`, `project_down`, `drift_detected`, and `not_project` can be verified from real `stage` invocations.
- [ ] T007 Implement `core/guidance/planner.go` to produce one primary action and direct command equivalents for each situation.
- [ ] T008 Add cheap context collection seams in `core/guidance/context.go` without running long Docker checks before first render.
- [ ] T009 Add terminal verification steps proving planner collection does not write `.env.stageserve` or mutate `.stageserve-state`.
- [ ] T010 Run focused supporting checks for `core/guidance` after terminal planner scenarios are green.

**Checkpoint**: Planner can decide the guided path without a terminal UI.

## Phase 3: Projection And Mode Cleanup

**Purpose**: Remove output-mode drift before adding a larger TUI.

- [ ] T011 [P] Refactor `cmd/stage/commands/setup.go` to use `onboarding.NewProjector`.
- [ ] T012 [P] Refactor `cmd/stage/commands/doctor.go` to use `onboarding.NewProjector`.
- [ ] T013 [P] Refactor `cmd/stage/commands/init.go` to use `onboarding.NewProjector`.
- [ ] T014 Decide and implement the `stage init --tui` contract, or explicitly document why init remains auto/text/JSON only.
- [ ] T015 Remove or correct stale `--recheck` documentation in `docs/runtime-contract.md` and any other active docs.
- [ ] T016 Add terminal JSON purity checks for setup, doctor, and init using real command output parsed with `jq` or equivalent.
- [ ] T017 Run focused supporting checks for `core/onboarding` and `cmd/stage/commands` after terminal JSON checks are green.

**Checkpoint**: Existing onboarding commands are cleaner and compatible before root TUI work begins.

## Phase 4: User Story 1 - Start From Bare `stage` (Priority: P1)

**Goal**: Bare `stage` opens guided TUI in TTY and text fallback outside TTY.

**Independent Verification**: Terminal scenarios prove TTY, non-TTY, help, disabled TUI, and direct subcommand routing.

### Terminal Verification

- [ ] T018 [P] [US1] Verify bare `stage` in a real TTY opens the guided path and record evidence in `quickstart.md`.
- [ ] T019 [P] [US1] Verify `stage > /tmp/stage-guidance.txt` does not hang and records plain guidance.
- [ ] T020 [P] [US1] Verify `stage --help` bypasses the guided path.
- [ ] T021 [P] [US1] Verify `STAGESERVE_NO_TUI=1 stage` uses text fallback.
- [ ] T022 [P] [US1] Verify direct subcommand help paths such as `stage up --help` remain direct.

### Implementation

- [ ] T023 [US1] Add injectable TTY/mode detection seam in `cmd/stage/commands/onboarding_mode.go` or a new root interaction helper.
- [ ] T024 [US1] Add root no-args `RunE` wiring in `cmd/stage/commands/root.go`.
- [ ] T025 [US1] Add text fallback renderer for `core/guidance.NextActionPlan`.
- [ ] T026 [US1] Add thin TUI adapter entrypoint in `cmd/stage/commands/tui.go`.
- [ ] T027 [US1] Run focused supporting checks for `cmd/stage/commands` and `core/guidance`.

**Checkpoint**: Bare `stage` routes correctly without implementing every guided action yet.

## Phase 5: User Story 2 - Complete First-Run Setup (Priority: P1)

**Goal**: Guided TUI can walk from not-ready or uninitialized state toward setup/init/up.

### Terminal Verification

- [ ] T028 [P] [US2] Verify a machine-not-ready scenario shows `setup` as the primary action.
- [ ] T029 [P] [US2] Verify a project without `.env.stageserve` shows `init` as the primary action.
- [ ] T030 [P] [US2] Verify config preview shows `.env.stageserve` path and values before write.
- [ ] T031 [P] [US2] Verify cancel-before-write leaves no `.env.stageserve` file.

### Implementation

- [ ] T032 [US2] Implement first TUI landing screen with context summary, primary action, secondary actions, advanced actions, and quit/help.
- [ ] T033 [US2] Wire setup action through existing onboarding setup checks and result projection.
- [ ] T034 [US2] Wire init action through `core/onboarding.WriteProjectEnv` with preview and confirmation.
- [ ] T035 [US2] Add Huh form or equivalent prompt for site name/docroot/suffix edits.
- [ ] T036 [US2] After init, recompute planner state and offer run action.
- [ ] T037 [US2] Validate first-run TUI manually and record evidence in `quickstart.md`.

**Checkpoint**: New users can use guided path for setup and config creation.

## Phase 6: User Story 3 - Manage A Running Project (Priority: P2)

**Goal**: Guided TUI supports day-to-day run/status/logs/down/doctor actions.

### Terminal Verification

- [ ] T038 [P] [US3] Verify a configured stopped project shows `up` as the primary action.
- [ ] T039 [P] [US3] Verify a running project shows status/logs/down/doctor actions.
- [ ] T040 [P] [US3] Verify stop action requires confirmation before running.
- [ ] T041 [P] [US3] Verify down/status/doctor results lead to a clear next action.

### Implementation

- [ ] T042 [US3] Wire `up` action through existing lifecycle command/domain seam.
- [ ] T043 [US3] Wire `status` action through existing status behavior.
- [ ] T044 [US3] Wire `logs` action with clear exit/cancel behavior.
- [ ] T045 [US3] Wire `down` action with confirmation and existing data-preserving semantics.
- [ ] T046 [US3] Wire `doctor` action through existing diagnostics.
- [ ] T047 [US3] Validate running-project TUI manually and record evidence in `quickstart.md`.

**Checkpoint**: Daily project management works from bare `stage`.

## Phase 7: User Story 4 - Preserve Power-User And Automation Paths (Priority: P1)

**Goal**: TUI additions do not break direct commands or automation.

- [ ] T048 [P] [US4] Verify non-TTY bare `stage` does not prompt.
- [ ] T049 [P] [US4] Verify `NO_COLOR=1` removes color styling where output is captured.
- [ ] T050 [P] [US4] Verify `stage setup --json` and `stage doctor --json` parse as JSON from terminal output.
- [ ] T051 [P] [US4] Verify direct command smoke paths for `stage up --help`, `stage status --help`, `stage logs --help`, and `stage down --help`.
- [ ] T052 [US4] Run focused supporting command compatibility checks and fix regressions.

**Checkpoint**: Power-user and automation paths remain stable.

## Phase 8: Documentation And Abstraction Cleanup

**Purpose**: Make docs reflect the simple-first product model.

- [ ] T053 [P] Update `README.md` so the first-run path starts with bare `stage`.
- [ ] T054 [P] Move Docker/gateway/network/volume details in `README.md` to advanced/troubleshooting sections.
- [ ] T055 [P] Update `docs/runtime-contract.md` for root guided behavior, no-TUI controls, and direct command compatibility.
- [ ] T056 [P] Add active `docs/installer-onboarding.md` using the researched install/setup/init/up/doctor flow.
- [ ] T057 [P] Update `.env.stageserve.example` comments to align with guided config creation.
- [ ] T058 [P] Update command `Short` and `Long` strings in `cmd/stage/commands/*.go` to prefer StageServe concepts over Docker/gateway internals.
- [ ] T059 Run docs grep for primary-path leaks: `docker compose|network|volume|gateway alias|nginx` in README first-run sections and command help.

**Checkpoint**: Primary docs teach StageServe usage before implementation internals.

## Phase 9: Final Verification

- [ ] T060 Build the binary with `make build` or the repository's canonical build target and record which binary is being exercised.
- [ ] T061 Run terminal verification matrix from `quickstart.md` and record command, exit code, and observed output.
- [ ] T062 Manually validate bare `stage` in a TTY from a project without `.env.stageserve`.
- [ ] T063 Manually validate bare `stage` in a TTY from a configured stopped project.
- [ ] T064 Manually validate bare `stage` in a TTY from a running project.
- [ ] T065 Manually validate non-TTY `stage > /tmp/stage-guidance.txt`.
- [ ] T066 Manually validate `STAGESERVE_NO_TUI=1 stage`.
- [ ] T067 Manually validate JSON purity for `stage setup --json` and `stage doctor --json`.
- [ ] T068 Run supporting `go test ./core/guidance ./core/onboarding ./cmd/stage/commands ./core/config ./core/lifecycle ./observability/status ./infra/gateway`.
- [ ] T069 Record final evidence and any unrun real-daemon gaps in `quickstart.md`.

## Dependencies

- Phase 2 blocks all user stories.
- Phase 3 should finish before Phase 4 to avoid projection drift.
- US1 root routing blocks US2 and US3 guided workflows.
- US4 compatibility checks run throughout but final verification depends on all implementation phases.
- Documentation cleanup can start after root behavior is stable, but final wording depends on implemented behavior.

## Notes

- Keep implementation slices small.
- Do not add new dependencies without explicit approval.
- Do not route normal users to Docker commands unless no StageServe-level recovery exists.
- Keep every mutating TUI action confirmable and cancellable.
- Preserve direct command examples for power users.
