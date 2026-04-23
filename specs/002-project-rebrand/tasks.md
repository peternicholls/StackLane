# Tasks: Stacklane Rebrand And Unified Command Surface

**Input**: Design documents from `/specs/002-project-rebrand/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/cli-contract.md`, `quickstart.md`

**Tests**: Formal automated tests are not explicitly requested in the specification. This task list therefore emphasizes shell validation, operational verification, and documentation/interface parity checks.

**Operational Verification**: This feature changes command workflows, migration guidance, and user-facing branding. Tasks therefore include validation for ease of use, command behavior, configuration precedence, isolation boundaries, failure visibility, recovery clarity, friction reduction, and documentation/interface parity.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., `US1`, `US2`, `US3`)
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the canonical Stacklane entrypoint and shared implementation surfaces that later story work will build on.

- [X] T001 Create the canonical CLI entrypoint file at `stacklane`
- [X] T002 [P] Prepare shared Stacklane branding and command-mapping support in `lib/stacklane-common.sh`
- [X] T003 [P] Align the shebang, executable behavior, and root-script launch pattern for `stacklane` with the existing repo entrypoints

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build the shared command-dispatch and migration infrastructure that all user stories depend on.

**⚠️ CRITICAL**: No user story work should begin until this phase is complete.

- [X] T004 Extend action parsing for the canonical CLI in `lib/stacklane-common.sh` so `stacklane` accepts exactly one primary action flag
- [X] T005 Implement central dispatch from `stacklane` into the existing runtime actions in `lib/stacklane-common.sh`
- [X] T006 Add canonical Stacklane help and invalid-action error handling in `lib/stacklane-common.sh`
- [X] T007 Preserve current shared options, `--all`, `--project`, `--dry-run`, and `version=` compatibility behavior in `lib/stacklane-common.sh`
- [X] T008 Implement a shared deprecation-forwarding path for legacy wrappers in `lib/stacklane-common.sh`

**Checkpoint**: Foundation ready; user story work can proceed in priority order.

---

## Phase 3: User Story 1 - Adopt The Stacklane Identity (Priority: P1) 🎯 MVP

**Goal**: Make Stacklane the sole active brand across maintained docs and user-facing macOS wrapper surfaces.

**Independent Test**: Review the main docs and GUI-facing surfaces and confirm Stacklane is the active identity everywhere except explicitly labeled migration references.

### Implementation for User Story 1

- [X] T009 [US1] Update the primary brand name, project summary, and top-level usage narrative in `README.md`
- [X] T010 [P] [US1] Update Stacklane branding and CLI references in `AUTOMATION-README.md`
- [X] T011 [P] [US1] Update Stacklane branding and GUI narrative in `GUI-HELP.md`
- [X] T012 [P] [US1] Update active branding language in `docs/runtime-contract.md`
- [X] T013 [P] [US1] Update active branding language in `docs/migration.md`
- [X] T014 [P] [US1] Update macOS app and workflow display names in `Stacklane Manager.app/Contents/Info.plist` and `Stacklane Manager.workflow/Contents/Info.plist`
- [X] T015 [P] [US1] Update AppleScript and workflow-facing brand text in `20i-stack-manager.scpt`, `20i-stack-launcher.workflow`, and `Stacklane Manager.app/Contents/Resources/Scripts/main.scpt`
- [X] T016 [US1] Validate that `README.md`, `AUTOMATION-README.md`, `GUI-HELP.md`, `docs/runtime-contract.md`, and the macOS wrapper metadata present Stacklane consistently with legacy naming only in migration contexts

**Checkpoint**: Stacklane is the visible identity across maintained primary surfaces.

---

## Phase 4: User Story 2 - Use One Memorable Command Entry Point (Priority: P1)

**Goal**: Make `stacklane` the canonical flag-driven command without changing the underlying runtime behavior.

**Independent Test**: From a sample project directory, confirm that `stacklane --help`, `stacklane --up`, `stacklane --status`, and `stacklane --down` work as the primary documented workflow.

### Implementation for User Story 2

- [X] T017 [US2] Implement the canonical `stacklane` dispatcher in `stacklane`
- [X] T018 [US2] Refine command usage text, action validation, and dispatch behavior for `stacklane` in `lib/stacklane-common.sh`
- [X] T019 [US2] Preserve runtime-state, config-precedence, and selector behavior under the `stacklane` invocation path in `lib/stacklane-common.sh`
- [X] T020 [US2] Update primary command examples and shell integration to use `stacklane` in `README.md`
- [X] T021 [P] [US2] Update primary command examples to use `stacklane` in `AUTOMATION-README.md` and `GUI-HELP.md`
- [X] T022 [P] [US2] Update command syntax examples to use `stacklane` in `docs/migration.md` and `docs/runtime-contract.md`
- [X] T023 [US2] Validate the happy-path command flow from `specs/002-project-rebrand/quickstart.md` using `stacklane --help`, `stacklane --up`, `stacklane --status`, and `stacklane --down`
- [X] T024 [US2] Validate failure handling for `stacklane` with no primary action and conflicting primary actions in `stacklane` and `lib/stacklane-common.sh`

**Checkpoint**: `stacklane` is the canonical usable interface for primary lifecycle actions.

---

## Phase 5: User Story 3 - Migrate Existing Users Without Ambiguity (Priority: P2)

**Goal**: Keep current operators moving by preserving compatibility wrappers and explicit migration guidance.

**Independent Test**: Invoke a retained legacy command and review the migration docs to confirm users can translate the old workflow to the new Stacklane command and distinguish the repo rename from the manual local-folder rename.

### Implementation for User Story 3

- [X] T025 [P] [US3] Convert `deprecated --up wrapper`, `deprecated --attach wrapper`, and `deprecated --down wrapper` into temporary forwarding wrappers with deprecation guidance toward `stacklane`
- [X] T026 [P] [US3] Convert `deprecated --detach wrapper`, `deprecated --status wrapper`, `deprecated --logs wrapper`, and `deprecated --dns-setup wrapper` into temporary forwarding wrappers with deprecation guidance toward `stacklane`
- [X] T027 [US3] Update old-to-new command mapping, wrapper expectations, and migration wording in `docs/migration.md`
- [X] T028 [US3] Update install, clone, and shell-path guidance to distinguish the repository rename from the manual containing-folder rename in `README.md`
- [X] T029 [P] [US3] Update repo-rename, wrapper, and sync guidance in `AUTOMATION-README.md` and `GUI-HELP.md`
- [X] T030 [US3] Validate that a retained wrapper invocation still succeeds and points users to the equivalent `stacklane` syntax using `deprecated --up wrapper` and the migration flow in `specs/002-project-rebrand/quickstart.md`

**Checkpoint**: Existing users can migrate from `legacy wrapper commands` to `stacklane` without guesswork.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finalize parity, validation, and external rename follow-through across the full feature.

- [X] T031 [P] Run shell syntax validation on `stacklane`, `deprecated --up wrapper`, `deprecated --attach wrapper`, `deprecated --down wrapper`, `deprecated --detach wrapper`, `deprecated --status wrapper`, `deprecated --logs wrapper`, `deprecated --dns-setup wrapper`, and `lib/stacklane-common.sh`
- [X] T032 [P] Run `shellcheck` on changed shell scripts if available and resolve relevant findings in `stacklane`, `legacy wrapper commands`, and `lib/stacklane-common.sh` (tool unavailable in the current environment; no local findings remained after syntax validation)
- [X] T033 Validate documentation and interface parity across `README.md`, `AUTOMATION-README.md`, `GUI-HELP.md`, `docs/migration.md`, and `docs/runtime-contract.md`
- [X] T034 Validate configuration precedence and runtime isolation remain unchanged under `stacklane` for stack `.env`, project `.20i-local`, shell environment overrides, and CLI flags via `lib/stacklane-common.sh`
- [X] T035 Validate claimed friction reduction by comparing the old multi-command flow against the new `stacklane` flow using `specs/002-project-rebrand/quickstart.md`
- [X] T036 Run the complete validation flow in `specs/002-project-rebrand/quickstart.md` and record any remaining untested macOS packaging caveats in `specs/002-project-rebrand/quickstart.md`
- [X] T037 Complete the external GitHub repository rename to the Stacklane name and reconcile any remaining clone URLs or repository-name references in `README.md` and `docs/migration.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; can start immediately.
- **Foundational (Phase 2)**: Depends on Phase 1; blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Phase 2; can deliver the branded MVP surfaces independently.
- **User Story 2 (Phase 4)**: Depends on Phase 2 and should follow User Story 1 for consistent naming in help output and docs.
- **User Story 3 (Phase 5)**: Depends on Phase 2 and should follow User Story 2 because wrapper forwarding targets the canonical `stacklane` command.
- **Polish (Phase 6)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other stories after Foundational is complete.
- **User Story 2 (P1)**: Reuses foundational dispatch work and benefits from User Story 1 branding updates, but can be validated independently once `stacklane` exists.
- **User Story 3 (P2)**: Depends on User Story 2 because legacy wrappers must point to the finished canonical command and migration docs must reference the final syntax.

### Within Each User Story

- Shared helper changes must land before wrapper or help-surface updates that depend on them.
- CLI behavior must be implemented before migration wrappers are switched over.
- Documentation updates should follow the final behavior for each story, then be validated immediately.
- Validation tasks complete the story before moving on.

### Parallel Opportunities

- `T002` and `T003` can run in parallel once `T001` is established.
- In User Story 1, `T010` through `T015` can run in parallel because they touch different files and surfaces.
- In User Story 2, `T021` and `T022` can run in parallel after the core command behavior is implemented.
- In User Story 3, `T025` and `T026` can run in parallel because they split the wrapper files into two non-overlapping sets.
- In the Polish phase, `T031` and `T032` can run in parallel.

---

## Parallel Example: User Story 1

```bash
Task: "Update Stacklane branding and CLI references in AUTOMATION-README.md"
Task: "Update Stacklane branding and GUI narrative in GUI-HELP.md"
Task: "Update active branding language in docs/runtime-contract.md"
Task: "Update active branding language in docs/migration.md"
Task: "Update macOS app and workflow display names in Stacklane Manager.app/Contents/Info.plist and Stacklane Manager.workflow/Contents/Info.plist"
```

---

## Parallel Example: User Story 3

```bash
Task: "Convert deprecated --up wrapper, deprecated --attach wrapper, and deprecated --down wrapper into temporary forwarding wrappers with deprecation guidance toward stacklane"
Task: "Convert deprecated --detach wrapper, deprecated --status wrapper, deprecated --logs wrapper, and deprecated --dns-setup wrapper into temporary forwarding wrappers with deprecation guidance toward stacklane"
Task: "Update repo-rename, wrapper, and sync guidance in AUTOMATION-README.md and GUI-HELP.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. Stop and validate that Stacklane is the consistent active brand across primary surfaces.

### Incremental Delivery

1. Complete Setup + Foundational.
2. Deliver User Story 1 for branding consistency.
3. Deliver User Story 2 for the canonical `stacklane` workflow.
4. Deliver User Story 3 for compatibility wrappers and migration clarity.
5. Finish with Polish and full quickstart validation.

### Parallel Team Strategy

1. One developer completes the foundational command-dispatch work.
2. After Phase 2, branding/documentation work for User Story 1 can be split from CLI examples for User Story 2.
3. Once `stacklane` is stable, wrapper conversion and migration docs for User Story 3 can proceed in parallel.

---

## Notes

- `[P]` tasks touch different files and can be worked in parallel.
- Story labels map each task back to the specification for traceability.
- Every story ends with an independent validation task so it can be demonstrated on its own.
- The external repository rename is complete; the local containing-folder rename remains a separate manual operator step.