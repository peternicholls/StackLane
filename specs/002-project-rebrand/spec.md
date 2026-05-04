# Feature Specification: StageServe Rebrand And `stage` Command Cutover

**Feature Branch**: `002-project-rebrand`  
**Created**: 2026-04-01  
**Status**: Draft  
**Input**: User description: "Rename and rebrand the project, propagate the new identity across the repository, update all documentation, and replace legacy entrypoints with one canonical root command."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Adopt The StageServe Identity (Priority: P1)

As a current or new operator, I want the project to use the name StageServe consistently everywhere, so that I can understand what the stack is called without seeing competing names across the repo, UI, and help text.

**Why this priority**: Naming is the foundation for every other change. If the approved identity is not applied first, command examples, operator docs, and migration guidance stay inconsistent.

**Independent Test**: Review the repository's primary user-facing surfaces and confirm they present StageServe as the official project name, with historical naming only appearing in explicitly labeled migration or archival contexts.

**Acceptance Scenarios**:

1. **Given** the project has adopted StageServe as its official name, **When** a user reads the repository landing documentation, **Then** StageServe appears as the primary identity and prior names are not presented as active branding.
2. **Given** the project includes support docs and archived wrapper metadata, **When** a user opens those surfaces, **Then** the approved product name and short description are used consistently for active behavior.

---

### User Story 2 - Use One Canonical Root Command (Priority: P1)

As an operator, I want to manage the stack through the `stage` command with subcommands, so that I have one memorable entrypoint for the primary lifecycle workflow.

**Why this priority**: The command surface is the main day-to-day UX. Moving to one canonical root command directly addresses the rename while keeping the operator workflow copy-pasteable.

**Independent Test**: From the repository's documented workflow, confirm that a user can perform the primary lifecycle actions through `stage <subcommand>` and does not need any legacy wrapper command.

**Acceptance Scenarios**:

1. **Given** a user wants to start a project runtime, **When** they follow the documented CLI workflow, **Then** they use `stage up`.
2. **Given** a user wants status, logs, attach, detach, shutdown, DNS setup, or doctor behavior, **When** they consult command help or documentation, **Then** those actions are expressed through the `stage` command family.

---

### User Story 3 - Migrate Existing Users Without Ambiguity (Priority: P2)

As an existing user, I want the rename and command cutover to be explicit, so that I can update habits, scripts, and repository references without guessing what changed.

**Why this priority**: Existing operators already rely on older names. Migration clarity reduces breakage and avoids stale docs surviving as active guidance.

**Independent Test**: Follow the migration guidance from the prior naming model to StageServe and confirm the user can identify the new repo name, the manual local-folder boundary, the new `stage` root command, and the current config/state directory names in one pass.

**Acceptance Scenarios**:

1. **Given** a user is familiar with prior naming, **When** they review the migration documentation, **Then** they can see that `stage` is the only supported root command and that old names are historical only.
2. **Given** the repository itself is renamed, **When** a user reviews setup and installation instructions, **Then** they can distinguish the repository rename from any manual local folder rename or deployment-path cleanup.

### Edge Cases

- A historical project name remains in a help screen, spec, or secondary document and creates conflicting branding.
- A user follows an older example or older directory name and needs clear correction toward `stage`, `.env.stageserve`, and `.stageserve-state`.
- The repository name changes before every documentation and support surface is updated, creating mixed instructions during onboarding.
- The containing folder is not manually renamed immediately after the repository rename, so instructions must avoid assuming the folder always matches the repo name.
- A deployed copy of the stack diverges from the workspace copy, so the docs must make clear which maintained surfaces are updated here and which live installs still require manual sync.

## Operational Impact *(mandatory)*

### Ease Of Use & Workflow Impact

- Affected commands, wrappers, or entry points: the canonical `stage` CLI, shell integration examples, archived launcher metadata, and user-facing help text.
- Backward compatibility or migration expectation: `stage` is the only supported executable path; legacy names survive only in migration notes or archived material.
- Operator friction removed or introduced: memorizing multiple entrypoints is removed; the main transition cost is updating local habits and scripts to the new command and naming vocabulary.

### Configuration & Precedence

- New or changed configuration inputs: active operator docs and specs must use project-root `.env.stageserve`, stack-home `.env.stageserve`, and `.stageserve-state` terminology.
- Precedence order: CLI flags, then project-root `.env.stageserve`, then shell environment, then stack-home `.env.stageserve`, then built-in defaults.

### State, Isolation & Recovery

- Affected runtime state: user-visible state summaries, install and clone instructions, root-command help, and documentation references to state/config directories.
- Isolation risk and mitigation: the rename must not change which project runtime is targeted by a given command; docs must keep project-selection safeguards and stack-home boundaries explicit.
- Reliability and recovery path: migration docs must explain the supported current names and make clear that archived or historical names are not active runtime behavior.

### Documentation Surfaces

- Docs and interfaces requiring updates: `README.md`, migration guidance, runtime-contract wording, installation examples, shell integration examples, and any spec text that still describes prior command names, prior config/state names, or removed wrapper entrypoints as current.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The project MUST adopt StageServe as the official replacement name and document it as the primary identity.
- **FR-002**: Maintained repository-facing surfaces MUST use StageServe as the primary identity, with historical naming limited to explicitly labeled migration or archival contexts.
- **FR-003**: The repository rename MUST be reflected throughout project materials, while explicitly stating that local folder or deployment-path cleanup is a separate manual operator action.
- **FR-004**: The command experience MUST provide one central top-level command named `stage` with the current lifecycle subcommands.
- **FR-005**: Users MUST be able to perform the currently supported lifecycle actions through the `stage` command pattern, including startup, attachment, shutdown, detachment, status review, log access, DNS setup, and doctor behavior.
- **FR-006**: The project MUST provide clear help and documentation showing the canonical `stage <subcommand>` syntax, including representative usage examples for the primary actions.
- **FR-007**: No compatibility forwarding shim may be documented as supported current behavior; old command names are migration history only.
- **FR-008**: Active documentation and specs MUST use `.env.stageserve` for human-owned config and `.stageserve-state` for runtime-owned state.
- **FR-009**: Migration documentation MUST include a concise explanation of what changed for current users versus what remains operationally the same.
- **FR-010**: User-facing text MUST use the new brand and command vocabulary consistently, including setup steps, shell integration, status descriptions, and operator docs.
- **FR-011**: Historical references to previous project names, commands, or directory names MUST be limited to migration or archival sections where they prevent confusion.
- **FR-012**: The rename and command cutover MUST not require users to relearn the underlying runtime model or isolation rules beyond the explicitly documented naming changes.

### Key Entities *(include if feature involves data)*

- **Brand Identity**: The StageServe name and its short description.
- **Canonical Command Surface**: The `stage` command, its supported subcommands, examples, and help wording.
- **Migration Mapping**: The documented relationship between prior project/command names and the current naming model.
- **Rename Surface Inventory**: The maintained user-facing places where branding and command vocabulary must be updated together.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new user can identify the official project name and its primary purpose from the main repository documentation within 2 minutes, without encountering competing active names.
- **SC-002**: All maintained user-facing surfaces in the repository use StageServe, `stage`, `.env.stageserve`, and `.stageserve-state` consistently, except for explicitly labeled migration references.
- **SC-003**: An existing operator can translate any previously documented root-command workflow into the new naming model using one migration reference section or less.
- **SC-004**: The primary lifecycle actions can all be understood from one central help path and one central command family.
- **SC-005**: Setup and migration documentation clearly distinguish the repository rename from manual local-folder cleanup so users can complete the transition without assuming an automated rename.

## Assumptions

- StageServe is the approved replacement name for this feature and will not be reopened unless scope changes.
- The local containing folder rename remains a manual operator action and must be documented, not automated.
- Existing runtime behavior, project isolation, and configuration precedence remain materially the same unless a change is explicitly called out elsewhere.
- Historical references to `20i` may remain only where needed to explain hosting semantics or migration context.
- `stage` is the only supported active root command.
