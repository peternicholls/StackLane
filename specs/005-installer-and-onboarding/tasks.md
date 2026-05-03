---

description: "Detailed implementation tasks for installer, onboarding, and readiness surfaces"

---

# Tasks: Installer, Onboarding, And Environment Readiness

**Input**: Design documents from `/specs/005-installer-and-onboarding/`
**Prerequisites**: `plan.md` (required), `spec.md` (required), `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Include focused automated tests because the feature spec defines mandatory user-story test criteria and stable machine contracts. Keep behavioral tests behind `core/onboarding` seams whenever possible; reserve command-package tests for adapter wiring, flag handling, and top-level contract compliance.

**TDD Rule**: For every implementation task in this file, first add or extend the narrowest failing test that proves the target behavior, then implement the minimum code to pass, then refactor locally while staying green. Do not front-load all tests for a story before writing code.

**Organization**: Tasks are grouped by user story so each story can be implemented and verified independently, but work inside each story should stay concept-local: deepen runtime, readiness, ownership, and projection modules before spreading behavior across command files.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish command scaffolding and shared onboarding artifacts.

- [x] T001 Add onboarding command registrations in `cmd/stacklane/commands/root.go`.
- [x] T002 Create setup command scaffold in `cmd/stacklane/commands/setup.go`.
- [x] T003 Create doctor command scaffold in `cmd/stacklane/commands/doctor.go`.
- [x] T004 Create init command scaffold in `cmd/stacklane/commands/init.go`.
- [x] T005 Add Bubble Tea and Huh dependencies in `go.mod`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build the deep modules that command adapters and tests will depend on.

**CRITICAL**: Complete this phase before user-story implementation.

- [x] T006 Implement shared onboarding types and result envelope in `core/onboarding/types.go`.
- [x] T007 Implement shared onboarding runtime and exit-code reduction in `core/onboarding/runtime.go`.
- [x] T008 [P] Implement text projection adapter in `core/onboarding/projection_text.go`.
- [x] T009 [P] Implement JSON projection adapter in `core/onboarding/projection_json.go`.
- [x] T010 [P] Implement TUI projection adapter in `core/onboarding/projection_tui.go`.
- [x] T011 Implement command-mode resolution for adapters in `cmd/stacklane/commands/onboarding_mode.go`.
- [x] T012 Implement shared machine-readiness module skeleton in `core/onboarding/machine_readiness.go`.
- [x] T013 Implement shared project-env ownership module skeleton in `core/onboarding/project_env.go`.
- [x] T014 Add shared runtime and projection tests in `core/onboarding/runtime_test.go`.

**Checkpoint**: Shared onboarding runtime is ready; user stories can proceed.

---

## Phase 3: User Story 1 - Install StackLane Through A Recommended Path (Priority: P1) 🎯 MVP

**Goal**: Deliver one deterministic install path with checksum verification and onboarding handoff behavior.

**Independent Test**: On a clean supported macOS machine, run the canonical install command and verify `stacklane --version` output and handoff behavior.

### Tests for User Story 1

- [x] T015 [P] [US1] Add installer behavior test script for interactive and non-interactive handoff in `scripts/tests/install_handoff_smoke.sh`.
- [x] T016 [P] [US1] Add installer checksum and asset-name validation test script in `scripts/tests/install_checksum_smoke.sh`.

### Implementation for User Story 1

Execute US1 in red-green-refactor slices: write the failing test for asset selection or handoff behavior first, make it pass, then move to the next install behavior.

- [x] T017 [US1] Implement release asset OS/arch detection and naming in `install.sh`.
- [x] T018 [US1] Implement checksum fetch and verification workflow in `install.sh`.
- [x] T019 [US1] Implement deterministic install destination and PATH warning behavior in `install.sh`.
- [x] T020 [US1] Implement interactive `stacklane setup --tui` handoff and `NONINTERACTIVE=1` next-step behavior in `install.sh`.
- [x] T021 [US1] Implement release artifact naming and checksum publication updates in `.github/workflows/release.yml`.
- [x] T022 [US1] Publish installer path, fallback verification, and next-step docs in `README.md`.

**Checkpoint**: Installer path is fully functional and independently verifiable.

---

## Phase 4: User Story 2 - Complete First-Run Machine Setup Reliably (Priority: P1)

**Goal**: Deliver `stacklane setup` with ordered readiness checks, explicit privilege semantics, and deterministic outputs.

**Independent Test**: Run `stacklane setup` with at least one missing prerequisite and verify `ready`/`needs_action`/`error` statuses with exact remediation.

### Tests for User Story 2

- [x] T023 [P] [US2] Add setup command-adapter flag and contract tests in `cmd/stacklane/commands/setup_test.go`.
- [x] T024 [P] [US2] Add machine-readiness and setup-policy tests in `core/onboarding/machine_readiness_test.go`.

### Implementation for User Story 2

Execute US2 in red-green-refactor slices: start with one failing readiness or policy test, pass it through the shared runtime or machine-readiness seam, then move to the next step behavior.

- [x] T025 [US2] Implement setup command adapter and mode wiring in `cmd/stacklane/commands/setup.go`.
- [x] T026 [US2] Implement suffix-resolution and setup policy on top of the runtime seam in `core/onboarding/runtime.go`.
- [x] T027 [US2] Implement shared Docker, state-dir, and port readiness rules in `core/onboarding/machine_readiness.go`.
- [x] T028 [US2] Implement shared DNS and mkcert readiness rules in `core/onboarding/machine_readiness.go` and `platform/dns/macos.go`.
- [x] T029 [US2] Implement setup-specific privilege prompts and manual remediation behavior in `cmd/stacklane/commands/setup.go`.
- [x] T030 [US2] Implement setup TUI projection adapter usage in `cmd/stacklane/commands/setup.go` and `core/onboarding/projection_tui.go`.
- [x] T031 [P] [US2] Add setup unsupported-platform and exit-code tests in `cmd/stacklane/commands/setup_platform_test.go`.
- [x] T032 [US2] Implement setup unsupported-platform policy on top of machine-readiness results in `cmd/stacklane/commands/setup.go`.

**Checkpoint**: Setup flow is independently usable, idempotent, and automation-safe.

---

## Phase 5: User Story 4 - Initialize A Project For StackLane (Priority: P1)

**Goal**: Deliver `stacklane init` with guided defaults, strict docroot validation, and safe write semantics.

**Independent Test**: In a project without `.env.stacklane`, run `stacklane init` and verify minimal file output, validation behavior, and next-step guidance.

### Tests for User Story 4

- [x] T033 [P] [US4] Add init command-adapter and contract tests in `cmd/stacklane/commands/init_test.go`.
- [x] T034 [P] [US4] Add project-env ownership tests in `core/onboarding/project_env_test.go`.

### Implementation for User Story 4

Execute US4 in red-green-refactor slices: start with one failing ownership or validation test, pass it in `core/onboarding/project_env.go`, then layer the adapter behavior afterward.

- [x] T035 [US4] Implement init command adapter and mode wiring in `cmd/stacklane/commands/init.go`.
- [x] T036 [US4] Implement project-root resolution, validation, and ownership rules in `core/onboarding/project_env.go`.
- [x] T037 [US4] Implement init interactive confirmation and adjustment flow on top of the ownership seam in `cmd/stacklane/commands/init.go` and `core/onboarding/projection_tui.go`.
- [x] T038 [US4] Implement minimal `.env.stacklane` write, overwrite protection, and preservation semantics in `core/onboarding/project_env.go`.
- [x] T039 [US4] Route the silent fallback helper through the shared project-env ownership module in `cmd/stacklane/commands/project_env.go`.
- [x] T040 [US4] Implement init success summary and next-step projection in `cmd/stacklane/commands/init.go`.

**Checkpoint**: Init flow is independently usable and safe for repeated runs.

---

## Phase 6: User Story 3 - Diagnose Drift And Recover Quickly (Priority: P2)

**Goal**: Deliver `stacklane doctor` read-only diagnostics with targeted recovery guidance and JSON parity.

**Independent Test**: Break one dependency (for example resolver file drift), run `stacklane doctor`, and verify precise detection plus recovery command.

### Tests for User Story 3

- [x] T041 [P] [US3] Add doctor command-adapter contract tests in `cmd/stacklane/commands/doctor_test.go`.
- [x] T042 [P] [US3] Add machine-readiness read-only policy tests in `core/onboarding/machine_readiness_test.go`.

### Implementation for User Story 3

Execute US3 in red-green-refactor slices: start with one failing read-only diagnostic behavior, pass it through the shared readiness seam, then add the next diagnostic projection.

- [x] T043 [US3] Implement doctor command adapter and mode wiring in `cmd/stacklane/commands/doctor.go`.
- [x] T044 [US3] Reuse shared machine-readiness rules in read-only mode for binary, Docker, DNS, state-dir, and port checks in `core/onboarding/machine_readiness.go`.
- [x] T045 [US3] Implement gateway-specific readiness adapter on top of the runtime seam in `cmd/stacklane/commands/doctor.go` and `infra/gateway/manager.go`.
- [x] T046 [US3] Implement doctor unsupported-platform policy on top of shared readiness results in `cmd/stacklane/commands/doctor.go`.
- [x] T047 [US3] Implement doctor plain-text and JSON projection usage in `cmd/stacklane/commands/doctor.go`.

**Checkpoint**: Doctor flow is independently usable and reports actionable drift diagnostics.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Complete parity, validation, and operator documentation across all stories.

- [x] T048 [P] Align onboarding command-contract ownership and module seams in `docs/runtime-contract.md`.
- [x] T049 [P] Publish compatibility matrix, integrity verification guidance, and repo-to-deployed-stack sync guidance in `docs/installer-onboarding.md`.
- [x] T050 [P] Align first-run sequence, command examples, and repo-to-deployed-stack sync notes in `README.md`.
- [x] T051 Execute startup and setup validation protocol, record commands, expected statuses, and evidence in `specs/005-installer-and-onboarding/quickstart.md`.
- [x] T052 Execute doctor/status inspection and failure-path validation protocol, record commands and evidence in `specs/005-installer-and-onboarding/quickstart.md`.
- [x] T053 Execute config precedence, isolation boundary, and recovery validation protocol, record commands and evidence in `specs/005-installer-and-onboarding/quickstart.md`.
- [x] T054 Execute performance validation protocol for healthy-path `setup`, healthy-path `doctor`, and TUI overhead, then record measured results in `specs/005-installer-and-onboarding/quickstart.md`.
- [x] T055 Run focused `core/onboarding`, `cmd/stacklane/commands`, `core/config`, and `core/lifecycle` test suites and record outcomes in `specs/005-installer-and-onboarding/quickstart.md`.

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup (Phase 1) has no dependencies.
- Foundational (Phase 2) depends on Setup and blocks all story work.
- User stories (Phases 3-6) depend on Foundational completion.
- Polish (Phase 7) depends on completion of all targeted user stories.

### User Story Dependencies

- US1 (Install): starts after Foundational; no dependency on other stories.
- US2 (Setup): starts after Foundational; depends on the shared runtime and machine-readiness seams, not on US1 implementation details.
- US4 (Init): starts after Foundational; depends on the shared project-env ownership seam and remains independently testable.
- US3 (Doctor): starts after Foundational; depends on the shared runtime and machine-readiness seams and remains independently testable.

### Recommended Delivery Order

1. US1 (MVP entrypoint)
2. US2 (first-run readiness)
3. US4 (project onboarding)
4. US3 (day-2 diagnostics)

---

## Parallel Execution Examples

### User Story 1

- T015 and T016 can run in parallel.
- T015/T016 should be used to open the first red cycle before T017-T020.
- T017-T020 are sequential in `install.sh` and should each be driven by the next failing behavior test.
- T021 can run in parallel with T022.

### User Story 2

- T023 and T024 can run in parallel.
- T023/T024 should open the first red cycle before T025-T032.
- T027 and T028 can run in parallel.
- T029 and T030 depend on the runtime and readiness seams being in place.

### User Story 4

- T033 and T034 can run in parallel.
- T033/T034 should open the first red cycle before T035-T040.
- T036 and T038 are the core ownership work and should finish before T039.
- T037 and T040 depend on the ownership seam.

### User Story 3

- T041 and T042 can run in parallel.
- T041/T042 should open the first red cycle before T043-T047.
- T044 and T045 can run in parallel after T043.
- T046 and T047 complete the story-specific policy and projection work.

---

## Implementation Strategy

## TDD Tracer Bullet Order

Before broad story work, use this concrete first-slice order:

1. T014 with `TestReduceExitCode_PrefersUnsupportedOS` in `core/onboarding/runtime_test.go`.
2. T014 with `TestOverallStatus_NeedsActionWithoutError` in `core/onboarding/runtime_test.go`.
3. T024 with `TestDockerBinaryCheck_MissingBinary` in `core/onboarding/machine_readiness_test.go`.
4. T023 with `TestSetup_NonInteractiveMissingSuffix` in `cmd/stacklane/commands/setup_test.go` or the equivalent runtime-level behavior test.
5. T034 with `TestValidateProjectEnv_RejectsDocrootOutsideProjectRoot` in `core/onboarding/project_env_test.go`.
6. T041 or T042 with `TestDoctor_UnhealthyGateway` in the deepest seam that can prove the behavior cleanly.

If a slice exposes a missing lower-level seam, add that seam and keep the test at the deepest public interface that still proves the same behavior.

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 (US1) only.
3. Validate canonical install success on clean macOS machine.

### Incremental Delivery

1. Deliver US1 for deterministic installation.
2. Deliver US2 for machine setup and readiness.
3. Deliver US4 for project bootstrap onboarding.
4. Deliver US3 for drift diagnostics and recovery.
5. Complete polish and run focused validation.

### Team Parallelization

1. One engineer completes Setup + Foundational phases.
2. Then split story work in parallel:
	- Engineer A: US2
	- Engineer B: US4
	- Engineer C: US3
3. Rejoin for Phase 7 parity and final validation.

---

## Notes

- `[P]` marks tasks that can execute concurrently without file conflicts.
- `[USx]` labels map each task to one user story for traceability.
- Keep each story independently runnable and verifiable at its checkpoint.
- Preserve shared JSON and status semantics across text, TUI, and JSON outputs.
