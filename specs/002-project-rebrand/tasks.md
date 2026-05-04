# Tasks: StageServe Rebrand And `stage` Command Cutover

**Input**: Design documents from `/specs/002-project-rebrand/`  
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/cli-contract.md`, `quickstart.md`

**Tests**: This task list emphasizes focused command validation, documentation/spec sweeps, and targeted stale-reference checks.

**Operational Verification**: This feature changes naming, root-command expectations, and operator guidance. Tasks therefore include validation for active branding, command behavior, config/state naming, migration clarity, and documentation/spec parity.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g. `US1`, `US2`, `US3`)
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the canonical StageServe naming vocabulary and active command surfaces that later story work will build on.

- [X] T001 Confirm the canonical CLI entrypoint is `stage` at the repository root.
- [X] T002 [P] Confirm the active compiled command surface is rooted in `cmd/stage/` and `stage-bin`.
- [X] T003 [P] Record archived shell/runtime material as historical-only under `previous-version-archive/`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Align the primary naming contract for docs, config, state, and command examples.

**⚠️ CRITICAL**: No user story work should begin until this phase is complete.

- [X] T004 Align the canonical product name to `StageServe` in the maintained spec and doc surfaces.
- [X] T005 Align the canonical root command to `stage <subcommand>` in maintained examples.
- [X] T006 Align active config/state directory references to `.env.stageserve` and `.stageserve-state`.
- [X] T007 Remove compatibility-wrapper assumptions from maintained operator guidance.

**Checkpoint**: Foundation ready; user story work can proceed in priority order.

---

## Phase 3: User Story 1 - Adopt The StageServe Identity (Priority: P1) 🎯 MVP

**Goal**: Make StageServe the sole active brand across maintained docs and specs.

**Independent Test**: Review the main docs and maintained specs and confirm StageServe is the active identity everywhere except explicitly labeled migration references.

### Implementation for User Story 1

- [X] T008 [US1] Update the primary brand name, project summary, and top-level usage narrative in `README.md`.
- [X] T009 [P] [US1] Update active branding language in `docs/runtime-contract.md` and `docs/migration.md`.
- [X] T010 [P] [US1] Update older maintained specs that still describe prior naming as the active brand.
- [X] T011 [US1] Validate that maintained docs/specs present StageServe consistently with legacy naming only in migration or archival contexts.

**Checkpoint**: StageServe is the visible identity across maintained primary surfaces.

---

## Phase 4: User Story 2 - Use One Canonical Root Command (Priority: P1)

**Goal**: Make `stage` the only active documented root command.

**Independent Test**: From the documented workflow, confirm that `stage --help`, `stage up`, `stage status`, and `stage down` are the primary examples.

### Implementation for User Story 2

- [X] T012 [US2] Update primary command examples and shell integration to use `stage` in maintained docs.
- [X] T013 [US2] Update older maintained specs that still describe prior root-command syntax as current.
- [X] T014 [US2] Remove compatibility-shim wording from maintained planning and contract artifacts.
- [X] T015 [US2] Validate the happy-path command flow from `specs/002-project-rebrand/quickstart.md` using `stage --help`, `stage up`, `stage status`, and `stage down`.

**Checkpoint**: `stage` is the canonical usable interface for primary lifecycle actions.

---

## Phase 5: User Story 3 - Migrate Existing Users Without Ambiguity (Priority: P2)

**Goal**: Keep migration guidance explicit while ensuring active docs no longer describe prior names as supported behavior.

**Independent Test**: Review migration guidance and maintained specs to confirm users can translate old naming to the current StageServe contract in one pass.

### Implementation for User Story 3

- [X] T016 [US3] Update migration wording so prior names are historical only.
- [X] T017 [US3] Distinguish repository rename, local folder naming, and deployed-copy sync expectations in maintained guidance.
- [X] T018 [US3] Validate that maintained specs/docs do not describe compatibility wrappers or prior root commands as active runtime behavior.

**Checkpoint**: Existing users can migrate to StageServe and `stage` without guesswork.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finalize parity and stale-reference cleanup across the full feature.

- [X] T019 [P] Run targeted stale-reference sweeps for prior project names, prior root-command paths, prior config/state names, and removed wrapper assumptions across maintained specs.
- [X] T020 [P] Validate docs and spec parity across `README.md`, `docs/`, and older maintained specs.
- [X] T021 Run the complete validation flow in `specs/002-project-rebrand/quickstart.md` and record any remaining untested caveats there.

## Notes

- `[P]` tasks touch different files and can be worked in parallel.
- Story labels map each task back to the specification for traceability.
- Every story ends with an independent validation task so it can be demonstrated on its own.