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
- **(new file)**: Marker on tasks that create a file that does not yet exist
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

- [ ] T001 [P] Add config precedence tests for `.env.stacklane` as the canonical stack-owned defaults file and for project `.env` staying application-owned in `core/config/loader_test.go`. Include a `STACK_HOME` override case that points at a directory containing `.env.stacklane` and asserts it loads from there.
- [ ] T002 [P] Add runtime naming default tests in `core/config/loader_test.go` for every project-scoped derivation: `ComposeProjectName` = `stln-<slug>`, `WebNetworkAlias` = `stln-<slug>-web`, `RuntimeNetwork` = `<compose-project>-runtime`, `DatabaseVolume` = `<compose-project>-db-data`.
- [ ] T003 [P] Add bootstrap failure classification and rollback coherence tests in `core/lifecycle/orchestrator_test.go`. Cover the post-readiness, pre-state-persist failure window so a rolled-back project is never left as `attached` in the state store.
- [ ] T004 [P] **(new file)** Create `observability/status/status_test.go` and add status reporting tests proving rollback does not leave phantom running state.
- [ ] T005 Codify the shared-resource naming rule in `core/config/loader_test.go` and `docs/runtime-contract.md` before changing project-scoped defaults: shared compose project and shared network use `stln-shared`; the gateway service network alias uses `stln-gateway`.
- [ ] T006 Hand-write updated gateway golden fixtures and tests in `infra/gateway/manager_test.go` and `infra/gateway/testdata/*` to express the final `stln-<slug>` upstream naming. Goldens are written by hand in this phase (not regenerated) so they fail until implementation lands.

**Checkpoint**: The test suite names the final contract before implementation begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Land the shared config and naming substrate that every user story depends on.

**⚠️ CRITICAL**: No user story is complete until this phase is complete.

- [ ] T007 Update stack-default loading in `core/config/loader.go` so `.env.stacklane` is the only supported stack-owned defaults file. Update the package-level docstring (lines ~1–8) and the `loadStackEnv` comment to match.
- [X] T008 Remove both legacy stack-default behaviors from `core/config/loader.go`: (a) reading `<stackHome>/.stackenv`, and (b) falling back to `<stackHome>/.env`. Add a regression test asserting `<stackHome>/.env` is NOT loaded as stack defaults.
- [X] T009 Update default project-scoped runtime naming in `core/config/loader.go` and `core/config/types.go` from `stacklane-` to `stln-`. Enumerated defaults to change: `ComposeProjectName` (`stln-<slug>`), `WebNetworkAlias` (`stln-<slug>-web`); confirm derived `RuntimeNetwork` and `DatabaseVolume` follow automatically. Update the `ProjectConfig` doc comment in `types.go` line ~109 that still references `.stackenv`.
- [X] T010 Apply the explicit shared-resource naming rule in `core/config/loader.go`, `docker-compose.shared.yml`, and any related config comments: shared compose project and shared network use `stln-shared`; the gateway service network alias uses `stln-gateway`. Confirm rendered nginx upstreams in `infra/gateway/templates.go` still resolve after the project-scoped rename.
- [X] T011 Update example/default env surfaces so operators are pointed at the correct stack-owned file: keep `.env.stacklane.example` as the stack-defaults template, confirm `.env.example` is not restored as a supported compatibility surface, and delete `.stackenv.example` from the repository.

**Checkpoint**: Config loading and naming defaults reflect the 004 contract everywhere the runtime derives them.

---

## Phase 3: User Story 1 - Bootstrap A Project Predictably (Priority: P1) 🎯 MVP

**Goal**: Keep one post-up bootstrap phase, keep it project-local, and make the outcome explicit and recoverable.

**Independent Test**: Configure `STACKLANE_POST_UP_COMMAND` in project-root `.env.stacklane`, run `stacklane up`, and confirm the runtime either completes bootstrap successfully or reports a named bootstrap failure and rolls the project back.

### Tests for User Story 1

- [X] T012 [P] [US1] Extend bootstrap precedence and phase tests in `core/config/loader_test.go` and `core/lifecycle/orchestrator_test.go`. Include three negative-path tests proving `STACKLANE_POST_UP_COMMAND` is ignored when set via (a) shell environment, (b) stack-home `.env.stacklane`, (c) project `.env` — and only honored when set via project-root `.env.stacklane`. Also assert the hook only runs in the post-up phase, after readiness, inside the apache service container.

### Implementation for User Story 1

- [X] T013 [US1] Restrict bootstrap source to project-root `.env.stacklane` only in `core/config/loader.go`. Concretely: remove `STACKLANE_POST_UP_COMMAND` from `trackedEnvKeys`, exclude it from the `loadStackEnv` merge into `merged`, and resolve `cfg.PostUpCommand` from the project-local map only. Document the special-case in the package docstring.
- [ ] T014 [US1] Make the single post-up bootstrap step explicit in `core/lifecycle/orchestrator.go` (`runPostUpHook`), keep it bound to the `apache` service container, and keep its lifecycle step name `post-up-hook` stable in `core/lifecycle/errors.go`. If the contract requires an explicit working directory (see `data-model.md`), set it on the docker exec invocation here. Honor the operator's context cancellation (`Ctrl-C`) by aborting the hook and triggering rollback.
- [ ] T015 [US1] Update supporting mocks and touched tests in `internal/mocks/mocks.go` and `core/lifecycle/orchestrator_test.go` to match the final bootstrap contract, including any new working-directory or context-cancel behavior introduced by T014.
- [ ] T016 [US1] Run focused validation for the bootstrap slice with `go test ./core/config ./core/lifecycle` and fix any contract regressions before moving on.

**Checkpoint**: A project-local bootstrap command declared in project-root `.env.stacklane` behaves predictably and fails as a named, rollback-triggering lifecycle step.

---

## Phase 4: User Story 2 - Distinguish Stacklane Failures From App Failures (Priority: P1)

**Goal**: Make lifecycle reporting distinguish bootstrap failure from infrastructure failure and keep status coherent after rollback.

**Independent Test**: Trigger a bootstrap failure after readiness and confirm Stacklane reports a bootstrap-specific lifecycle failure, rolls the project back, and leaves `stacklane status` coherent.

### Tests for User Story 2

- [ ] T017 [P] [US2] **(new file)** Create `core/lifecycle/errors_test.go` with failure-classification assertions for bootstrap vs gateway/DNS/readiness failures, exercising `StepError` step names against the contract.
- [ ] T018 [P] [US2] In `observability/status/status_test.go` (created in T004), add rollback-state assertions proving bootstrap failure does not report the project as still running and no record is left as `attached` in the state store.
- [ ] T019 [P] [US2] Add rollback-isolation assertions in `core/lifecycle/orchestrator_test.go` and `observability/status/status_test.go` proving one project's bootstrap failure does not mutate another attached project's routes, registry entry, or reported state. Include the post-readiness/pre-persist failure window from T003.

### Implementation for User Story 2

- [ ] T020 [US2] Tighten bootstrap failure wrapping and remediation messaging in `core/lifecycle/errors.go` and `core/lifecycle/orchestrator.go`. Spot-check the rendered text via `go test` golden output and ensure `cmd/stacklane/commands/up.go` still surfaces the error legibly to operators.
- [ ] T021 [P] [US2] Update rollback handling in `core/lifecycle/orchestrator.go` so failed bootstrap attempts leave state and route outcomes coherent: state record is removed (or never persisted as `attached`), gateway route is not added, and unrelated projects' routes are untouched.
- [ ] T022 [P] [US2] Update `observability/status/status.go` so post-rollback status output reflects reality, keeps bootstrap failure separate from infrastructure readiness failures, and preserves unrelated attached project state.
- [ ] T023 [US2] Run focused validation for failure-path reporting with `go test ./core/lifecycle ./observability/status` and record any remaining real-daemon-only gap in `specs/004-workflow-and-lifecycle/quickstart.md`.

**Checkpoint**: Bootstrap failure is operator-visible as its own class of lifecycle error, and rollback no longer leaves ambiguous state behind.

---

## Phase 5: User Story 3 - Validate Multi-Project Workflow Against Real Projects (Priority: P2)

**Goal**: Align naming, docs, and real-project validation so multi-project operator workflows are easy to follow and verify.

**Independent Test**: Validate one representative bootstrap-sensitive app and one multi-project scenario, explicitly exercising `attach`, DNS routing, shared-gateway readiness, runtime env injection, DB provisioning alignment, bootstrap behavior, rollback isolation, teardown, `.env.stacklane`, and `stln-` naming.

### Tests for User Story 3

- [ ] T024 [P] [US3] Extend config and gateway tests in `core/config/loader_test.go` and `infra/gateway/manager_test.go` for final naming behavior after the shared-resource rule is applied (T005, T010).
- [ ] T024a [P] [US3] Add an attach-slice test in `core/lifecycle/orchestrator_test.go` covering `Orchestrator.Attach` against the new naming: route is added to the gateway, state moves to `attached`, and the rendered upstream uses the `stln-<slug>-web` alias. If a behavior cannot be exercised without a real daemon, record the explicit gap in `quickstart.md` instead.

### Implementation for User Story 3

- [X] T025 [US3] Update `README.md` to document `.env.stacklane`, `stln-` project-scoped runtime names, explicit `attach` validation, Stacklane-managed shared routing behavior for normal operators, the bootstrap source restriction, and the repo-to-deployed-copy sync point under `$HOME/docker/20i-stack` as a local example. Remove the `.stackenv.example` reference from the directory tree (line ~193).
- [ ] T026 [P] [US3] Update `docs/runtime-contract.md` to match the final config precedence, naming contract, failure classification, shared-resource naming, bootstrap timeout/cancel semantics (per `research.md`), and validation expectations.
- [X] T027 [P] [US3] Update operator-facing examples and example env guidance around `.env.stacklane.example` and project-local `.env.stacklane`. Confirm `.stackenv.example` was deleted in T011 and that no doc still advertises `.env.example` as a supported surface.
- [ ] T028 [US3] Update any project-scoped gateway/upstream naming assumptions that changed with `stln-` in `infra/gateway/testdata/*`, `infra/gateway/manager.go`, `infra/gateway/templates.go`, and related docs.
- [ ] T029 [US3] Execute the validation workflow in `specs/004-workflow-and-lifecycle/quickstart.md` against one representative app and one multi-project scenario, explicitly checking `attach`, DNS routing, shared-gateway readiness, runtime env injection, DB provisioning alignment, bootstrap behavior, rollback isolation, and teardown; if any check is unrun, record the exact gap in that same file.

**Checkpoint**: The multi-project workflow is documented, named clearly, and validated against real runtime behavior.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finish parity, run final checks, and keep documentation and runtime behavior aligned.

- [X] T030 [P] Documentation parity sweep across `README.md`, `docs/runtime-contract.md`, `docs/architecture.md`, `docs/migration.md`, `CONTRIBUTING.md`, `specs/004-workflow-and-lifecycle/quickstart.md`, and `specs/004-workflow-and-lifecycle/contracts/workflow-lifecycle-contract.md` so operator-facing surfaces agree on `.env.stacklane`, `stln-`, Stacklane-managed shared routing for normal operators, and the bootstrap source restriction, while lower-level contract material retains exact internal names where needed.
- [X] T030a [P] Code-comment sweep: grep `core/`, `infra/`, `cmd/`, and `internal/` for surviving `stackenv|stacklane-<slug>|stacklane-\$` references in docstrings, comments, help text, and error remediation strings. Update or remove every non-historical, non-regression-test hit. Any remaining matches must be intentional negative-path coverage or lower-level internal naming.
- [X] T030b [P] CLI surface check: rebuild `stacklane-bin` and confirm `stacklane up --help`, `stacklane status --help`, `stacklane down --help`, `stacklane attach --help`, `stacklane logs --help` contain no references to `.stackenv` or `stacklane-<slug>` paths.
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
- **US2 (P1)**: Depends on US1's final bootstrap contract.
- **US3 (P2)**: Depends on foundational naming/config work and should consume the final lifecycle behavior from US1/US2 before final validation.

### Within Each User Story

- Write or update tests first and ensure they fail against the old behavior.
- Change config derivation before changing docs that describe it.
- Change lifecycle error handling before changing status reporting that depends on it.
- Run focused validation immediately after each story slice is implemented.

### Parallel Opportunities

- T001–T004 can run in parallel.
- T006 depends on T005.
- T007–T011 can run in parallel in small groups once T005 resolves the shared-resource rule.
- T017–T019 can run in parallel.
- T021 and T022 can run in parallel.
- T024 and T024a can run in parallel.
- T025–T028 can run in parallel once the final naming behavior is settled.
- T030, T030a, T030b can run in parallel.

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
- Do not reintroduce `.stackenv` or `stacklane-<slug>` as supported names while implementing this feature. Per the workspace legacy policy, no backward-compatibility shim is added.
- Treat explicit validation notes as deliverables, not as optional commentary.
- `handoff.md` is a historical sprint-closure artifact and intentionally still references `.stackenv`. Do not edit it as part of doc-parity work; the banner at the top of that file marks it superseded.
