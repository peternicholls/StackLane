---

description: "Tasks for workflow and lifecycle hardening"

---

# Tasks: Workflow And Lifecycle Hardening

**Input**: Design documents from `/specs/004-workflow-and-lifecycle/`  
**Prerequisites**: [plan.md](./plan.md) (required), [spec.md](./spec.md) (required for user stories), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/workflow-lifecycle-contract.md](./contracts/workflow-lifecycle-contract.md)

**Tests**: Focused Go tests are required for touched slices in `core/config`, `core/lifecycle`, `observability/status`, and `infra/gateway`. Real-daemon validation is also required by the feature, but it remains a manual workflow unless explicit automation is added during implementation.

**Operational Verification**: This task list includes validation for startup, bootstrap execution, failure classification, rollback clarity, config precedence, naming clarity, isolation boundaries, teardown behavior, and documentation parity.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (`US1`, `US2`, `US3`)
- Include exact file paths in descriptions

## Path Conventions

Single Go module at repository root:

- `cmd/stacklane/` — CLI wiring
- `core/config/` — precedence, stack defaults, runtime naming
- `core/lifecycle/` — bootstrap execution, rollback, failure classification
- `core/state/` — project state persistence and registry projection
- `infra/gateway/` — route rendering and gateway config behavior
- `observability/status/` — operator-visible runtime state and drift reporting
- `docs/` and `README.md` — runtime contract and operator guidance

---

## Phase 1: Setup (Shared Contract Codification)

**Purpose**: Codify the new naming and lifecycle contract in tests before changing runtime behavior.

- [ ] T001 [P] Add config precedence tests for `.env.stacklane` as the canonical stack-owned defaults file and for project `.env` staying application-owned in `core/config/loader_test.go`.
- [ ] T002 [P] Add runtime naming default tests for `stln-<slug>`, `<compose-project>-runtime`, and `<compose-project>-db-data` in `core/config/loader_test.go`.
- [ ] T003 [P] Add bootstrap failure classification and rollback coherence tests in `core/lifecycle/orchestrator_test.go`.
- [ ] T004 [P] Add status reporting tests that prove rollback does not leave phantom running state in `observability/status/status_test.go`.
- [ ] T005 Decide and codify the shared-gateway naming rule as `stacklane-shared` in `core/config/loader_test.go` and `docs/runtime-contract.md` before changing project-scoped defaults.
- [ ] T006 Update gateway golden tests and fixtures for the final project-scoped and shared naming contract in `infra/gateway/manager_test.go` and `infra/gateway/testdata/*`.

**Checkpoint**: The test suite names the final contract before implementation begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Land the shared config and naming substrate that every user story depends on.

**⚠️ CRITICAL**: No user story is complete until this phase is complete.

- [ ] T007 Update stack-default loading in `core/config/loader.go` so `.env.stacklane` is the only supported stack-owned defaults file.
- [ ] T008 Remove old stack-default naming behavior from `core/config/loader.go` and any related tests so `.stackenv` is not part of the active contract.
- [ ] T009 Update default project-scoped runtime naming in `core/config/loader.go` and `core/config/types.go` from `stacklane-` to `stln-`.
- [ ] T010 Apply the explicit shared-gateway naming rule in `core/config/loader.go`, `docker-compose.shared.yml`, and any related config comments so shared infrastructure remains `stacklane-shared`.
- [ ] T011 Update example/default env surfaces so operators are pointed at the correct stack-owned file in `.env.example` and a new `.env.stacklane.example`; retire `.stackenv.example` from the supported path.

**Checkpoint**: Config loading and naming defaults reflect the 004 contract everywhere the runtime derives them.

---

## Phase 3: User Story 1 - Bootstrap A Project Predictably (Priority: P1) 🎯 MVP

**Goal**: Keep one post-up bootstrap phase, keep it project-local, and make the outcome explicit and recoverable.

**Independent Test**: Configure `STACKLANE_POST_UP_COMMAND` in `.stacklane-local`, run `stacklane up`, and confirm the runtime either completes bootstrap successfully or reports a named bootstrap failure and rolls the project back.

### Tests for User Story 1

- [ ] T012 [P] [US1] Extend bootstrap precedence and phase tests in `core/config/loader_test.go` and `core/lifecycle/orchestrator_test.go` so the hook is sourced only from `.stacklane-local` and only runs in the post-up phase.

### Implementation for User Story 1

- [ ] T013 [US1] Keep bootstrap configuration project-local in `core/config/loader.go` by resolving `STACKLANE_POST_UP_COMMAND` only from `.stacklane-local`.
- [ ] T014 [US1] Make the single post-up bootstrap step explicit in `core/lifecycle/orchestrator.go`, keep it bound to the `apache` service container, and keep its lifecycle step naming stable in `core/lifecycle/errors.go`.
- [ ] T015 [US1] Update supporting mocks and touched tests in `internal/mocks/mocks.go` and `core/lifecycle/orchestrator_test.go` to match the final bootstrap contract.
- [ ] T016 [US1] Run focused validation for the bootstrap slice with `go test ./core/config ./core/lifecycle` and fix any contract regressions before moving on.

**Checkpoint**: A project-local bootstrap command behaves predictably and fails as a named, rollback-triggering lifecycle step.

---

## Phase 4: User Story 2 - Distinguish Stacklane Failures From App Failures (Priority: P1)

**Goal**: Make lifecycle reporting distinguish bootstrap failure from infrastructure failure and keep status coherent after rollback.

**Independent Test**: Trigger a bootstrap failure after readiness and confirm Stacklane reports a bootstrap-specific lifecycle failure, rolls the project back, and leaves `stacklane status` coherent.

### Tests for User Story 2

- [ ] T017 [P] [US2] Add failure-classification assertions in `core/lifecycle/orchestrator_test.go` and `core/lifecycle/errors_test.go` for bootstrap vs gateway/DNS/readiness failures.
- [ ] T018 [P] [US2] Add rollback-state assertions in `observability/status/status_test.go` to prove bootstrap failure does not report the project as still running.
- [ ] T019 [P] [US2] Add rollback-isolation assertions proving one project's bootstrap failure does not mutate another attached project's routes, registry entry, or reported state in `core/lifecycle/orchestrator_test.go` and `observability/status/status_test.go`.

### Implementation for User Story 2

- [ ] T020 [US2] Tighten bootstrap failure wrapping and remediation messaging in `core/lifecycle/errors.go` and `core/lifecycle/orchestrator.go`.
- [ ] T021 [P] [US2] Update rollback handling in `core/lifecycle/orchestrator.go` so failed bootstrap attempts leave state and route outcomes coherent.
- [ ] T022 [P] [US2] Update `observability/status/status.go` so post-rollback status output reflects reality, keeps bootstrap failure separate from infrastructure readiness failures, and preserves unrelated attached project state.
- [ ] T023 [US2] Run focused validation for failure-path reporting with `go test ./core/lifecycle ./observability/status` and record any remaining real-daemon-only gap in `specs/004-workflow-and-lifecycle/quickstart.md`.

**Checkpoint**: Bootstrap failure is operator-visible as its own class of lifecycle error, and rollback no longer leaves ambiguous state behind.

---

## Phase 5: User Story 3 - Validate Multi-Project Workflow Against Real Projects (Priority: P2)

**Goal**: Align naming, docs, and real-project validation so multi-project operator workflows are easy to follow and verify.

**Independent Test**: Validate one representative bootstrap-sensitive app and one multi-project scenario, explicitly exercising `attach`, DNS routing, shared-gateway readiness, runtime env injection, DB provisioning alignment, bootstrap behavior, rollback isolation, teardown, `.env.stacklane`, and `stln-` naming.

### Tests for User Story 3

- [ ] T024 [P] [US3] Extend config and gateway tests in `core/config/loader_test.go` and `infra/gateway/manager_test.go` for final naming behavior after the shared-gateway rule is applied.

### Implementation for User Story 3

- [ ] T025 [US3] Update `README.md` to document `.env.stacklane`, `stln-` project-scoped runtime names, explicit `attach` validation, the shared `stacklane-shared` contract, and the repo-to-deployed-copy sync point under `$HOME/docker/20i-stack`.
- [ ] T026 [P] [US3] Update `docs/runtime-contract.md` to match the final config precedence, naming contract, failure classification, shared-gateway naming, and validation expectations.
- [ ] T027 [P] [US3] Update operator-facing examples and example env files in `.env.example` and `.env.stacklane.example`, and remove `.stackenv.example` from the supported documentation path.
- [ ] T028 [US3] Update any project-scoped gateway/upstream naming assumptions that changed with `stln-` in `infra/gateway/testdata/*`, `infra/gateway/manager.go`, and related docs.
- [ ] T029 [US3] Execute the validation workflow in `specs/004-workflow-and-lifecycle/quickstart.md` against one representative app and one multi-project scenario, explicitly checking `attach`, DNS routing, shared-gateway readiness, runtime env injection, DB provisioning alignment, bootstrap behavior, rollback isolation, and teardown; if any check is unrun, record the exact gap in that same file.

**Checkpoint**: The multi-project workflow is documented, named clearly, and validated against real runtime behavior.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finish parity, run final checks, and keep documentation and runtime behavior aligned.

- [ ] T030 [P] Documentation parity sweep across `README.md`, `docs/runtime-contract.md`, `specs/004-workflow-and-lifecycle/quickstart.md`, and `specs/004-workflow-and-lifecycle/contracts/workflow-lifecycle-contract.md` so every operator-facing surface says the same thing.
- [ ] T031 Run the focused implementation test suite with `go test ./core/config ./core/lifecycle ./observability/status ./infra/gateway`.
- [ ] T032 Validate startup, `attach`, status/inspection, teardown, and one failure path from the final operator workflow; if any part remains manual-only or unrun, record the gap explicitly in `specs/004-workflow-and-lifecycle/quickstart.md`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: Start immediately.
- **Phase 2 (Foundational)**: Depends on Phase 1. Blocks all user stories.
- **Phase 3 (US1)**: Depends on Phase 2.
- **Phase 4 (US2)**: Depends on US1 because it builds on the final bootstrap step behavior.
- **Phase 5 (US3)**: Depends on Phase 2 and can overlap late US1/US2 work once naming defaults are stable.
- **Phase 6 (Polish)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: Starts after foundational config and naming work lands.
- **US2 (P1)**: Depends on US1’s final bootstrap contract.
- **US3 (P2)**: Depends on foundational naming/config work and should consume the final lifecycle behavior from US1/US2 before final validation.

### Within Each User Story

- Write or update tests first and ensure they fail against the old behavior.
- Change config derivation before changing docs that describe it.
- Change lifecycle error handling before changing status reporting that depends on it.
- Run focused validation immediately after each story slice is implemented.

### Parallel Opportunities

- T001-T004 can run in parallel.
- T006 depends on T005.
- T007-T011 can run in parallel in small groups once T005 resolves the shared-gateway rule.
- T017-T019 can run in parallel.
- T021 and T022 can run in parallel.
- T025-T028 can run in parallel once the final naming behavior is settled.

---

## Implementation Strategy

### MVP First (US1 Then US2)

1. Codify the contract in tests.
2. Land the foundational naming/config work.
3. Finish US1 so bootstrap behavior is explicit and predictable.
4. Finish US2 so failure classification and rollback reporting are trustworthy.
5. Stop and validate the lifecycle slice before expanding into broader docs and multi-project verification.

### Incremental Delivery

1. Finish Phases 1 and 2.
2. Deliver US1 and validate it independently.
3. Deliver US2 and validate it independently.
4. Deliver US3 and run the real-project workflow.
5. Finish polish and parity checks.

## Notes

- Keep language imperative in code, docs, and task execution where ambiguity would weaken the contract.
- Do not reintroduce `.stackenv` or `stacklane-` as supported names while implementing this feature.
- Treat explicit validation notes as deliverables, not as optional commentary.