# Feature Specification: Stacklane Rebrand And Unified Command Surface

**Feature Branch**: `002-project-rebrand`  
**Created**: 2026-04-01  
**Status**: Draft  
**Input**: User description: "Rename and rebrand the project with a memorable future-facing name, propagate the new identity across the repository, update all documentation, and replace the current command set with one central command that uses modifiers such as --up and --down."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Adopt The Stacklane Identity (Priority: P1)

As a current or new operator, I want the project to use the name Stacklane consistently everywhere, so that I can understand what the stack is called without seeing competing names across the repo, UI, and help text.

**Why this priority**: Naming is the foundation for every other change. If the replacement identity is not chosen and applied first, the command redesign and documentation migration remain inconsistent.

**Independent Test**: Review the repository's primary user-facing surfaces and confirm they present Stacklane as the official project name, one short description, and one clear statement of scope, with historical naming only appearing in migration-specific contexts.

**Acceptance Scenarios**:

1. **Given** the project has adopted Stacklane as its official name, **When** a user reads the repository landing documentation, **Then** Stacklane appears as the primary identity and the previous name is not presented as the active brand.
2. **Given** the project includes GUI wrappers, scripts, and support docs, **When** a user opens those surfaces, **Then** the same approved name and description are used consistently.

---

### User Story 2 - Use One Memorable Command Entry Point (Priority: P1)

As an operator, I want to manage the stack through the `stacklane` command with clear modifiers, so that I do not need to remember a separate executable for each lifecycle action.

**Why this priority**: The command surface is the main day-to-day UX. Collapsing the verbs into one entry point directly addresses the requested usability and technical debt reduction.

**Independent Test**: From the repository's documented workflow, confirm that a user can perform the primary lifecycle actions through the `stacklane` command, with each action expressed as a modifier rather than a separate primary command name.

**Acceptance Scenarios**:

1. **Given** a user wants to start a project runtime, **When** they follow the documented CLI workflow, **Then** they use `stacklane` with an action modifier instead of a dedicated start executable.
2. **Given** a user wants status, logs, attach, detach, shutdown, or DNS setup behavior, **When** they consult command help or documentation, **Then** those actions are expressed through the same `stacklane` command family.

---

### User Story 3 - Migrate Existing Users Without Ambiguity (Priority: P2)

As an existing user of the current project, I want the rename and command migration to be explicit, so that I can update my habits, scripts, and repository references without guessing what changed.

**Why this priority**: Existing operators already rely on the current repo name and command names. Migration clarity reduces breakage and prevents support churn.

**Independent Test**: Follow the migration guidance from the current naming and command model to Stacklane, and confirm the user can identify the new repo name, the manual folder-rename boundary, and the old-to-new command mapping in one pass.

**Acceptance Scenarios**:

1. **Given** a user is familiar with the previous command set, **When** they review the migration documentation, **Then** they can see how each previous action maps to the new central-command syntax.
2. **Given** the repository itself is renamed, **When** a user reviews setup and installation instructions, **Then** they can distinguish the repository rename from the manually performed containing-folder rename.

### Edge Cases

- A historical project name remains in a help screen, GUI label, or secondary document and creates conflicting branding.
- A user invokes an old command name from habit or from automation and needs a clear migration outcome instead of silent failure or misleading behavior.
- The repository name changes before every documentation and UI surface is updated, creating mixed instructions during onboarding.
- The containing folder is not manually renamed immediately after the repository rename, so instructions must avoid assuming the folder name always matches the repo name.
- The deployed copy of the stack diverges from the repository copy, so rename guidance must make clear which maintained surfaces are updated by this feature and which operational copies still require manual sync.

## Operational Impact *(mandatory)*

### Ease Of Use & Workflow Impact

- Affected commands, wrappers, or entry points: all current `legacy wrapper commands` CLI scripts, shell integration examples, AppleScript app naming, workflow service naming, and user-facing help text.
- Backward compatibility or migration expectation: `stacklane` becomes the only primary documented workflow; the existing `legacy wrapper commands` commands remain temporarily as wrappers that direct users toward the new syntax during migration.
- Operator friction removed or introduced: memorizing multiple top-level commands is removed; the main temporary friction is learning the new command syntax and updated repository identity.

### Configuration & Precedence

- New or changed configuration inputs: command invocation syntax changes from multiple executables to one executable with action modifiers; existing project configuration inputs are expected to remain logically equivalent unless explicitly renamed for consistency with the new brand.
- Precedence order: existing runtime configuration precedence remains unchanged unless a renamed command help surface requires terminology updates only.

### State, Isolation & Recovery

- Affected runtime state: command entry points, recorded project identity in user-visible state summaries, registry- and status-facing terminology, installation and clone instructions, and GUI-launch labels. Runtime isolation rules, containers, networks, and data remain behaviorally unchanged unless required for command parity.
- Isolation risk and mitigation: the rename must not change which project runtime is targeted by a given action; command migration messaging must preserve existing project-selection safeguards.
- Reliability and recovery path: users must have a documented fallback path when invoking outdated commands or following outdated clone/setup instructions, including clear migration guidance and a defined manual step for renaming the containing folder.

### Documentation Surfaces

- Docs and interfaces requiring updates: `README.md`, `AUTOMATION-README.md`, `GUI-HELP.md`, migration guidance, runtime contract terminology, installation examples, shell integration examples, app/workflow labels, and any inline help or prompts surfaced by command wrappers.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The project MUST adopt Stacklane as the official replacement name and document it as the primary identity.
- **FR-002**: The project MUST document the rationale and evaluation criteria used to choose Stacklane so future contributors understand why it was selected.
- **FR-003**: All maintained repository-facing surfaces MUST adopt Stacklane as the primary identity, including documentation, setup guidance, GUI labels, and help text.
- **FR-004**: The repository rename MUST be reflected throughout the project materials, while explicitly stating that the containing local folder rename is a separate manual operator action.
- **FR-005**: The command experience MUST provide one central top-level command named `stacklane` that supports the current lifecycle actions through action modifiers rather than separate primary executable names.
- **FR-006**: Users MUST be able to perform the currently supported lifecycle actions through the `stacklane` command pattern, including startup, attachment, shutdown, detachment, status review, log access, and local DNS setup.
- **FR-007**: The project MUST provide clear help and documentation showing the central command syntax, including representative usage examples for the primary actions.
- **FR-008**: The existing `legacy wrapper commands` command names MUST remain temporarily as wrappers with deprecation guidance that redirects users to `stacklane`, and they MUST NOT remain the primary documented workflow.
- **FR-009**: Migration documentation MUST include an old-to-new command mapping and a concise explanation of what changes for current users versus what remains operationally the same.
- **FR-010**: User-facing text MUST use the new brand and command vocabulary consistently, including setup steps, shell integration, status descriptions, and GUI-support documentation.
- **FR-011**: Historical references to the previous project name MUST be limited to migration or historical-context sections where they prevent confusion.
- **FR-012**: The rename and command redesign MUST not require users to relearn the underlying runtime model, configuration precedence, or project isolation rules unless the documentation explicitly calls out the difference.

### Key Entities *(include if feature involves data)*

- **Brand Identity**: The Stacklane name, its short description, and the naming criteria that justify why it becomes the official public-facing identity.
- **Unified Command Surface**: The `stacklane` command, its supported action modifiers, examples, help wording, and migration expectations.
- **Migration Mapping**: The documented relationship between the previous repository/command names and the new naming model, including what is automatic versus manual.
- **Rename Surface Inventory**: The set of maintained user-facing places where branding and command vocabulary must be updated together.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new user can identify the official project name and its primary purpose from the main repository documentation within 2 minutes, without encountering competing active names.
- **SC-002**: All maintained user-facing surfaces in the repository use Stacklane and the `stacklane` command vocabulary consistently, except for explicitly labeled migration references.
- **SC-003**: An existing operator can translate any previously documented primary command into the new command pattern using one migration reference section or less.
- **SC-004**: The primary lifecycle actions can all be understood from one central help path and one central command pattern, rather than requiring separate top-level command discovery.
- **SC-005**: Setup and migration documentation clearly distinguish the repository rename from the manual containing-folder rename, so users can complete the transition without assuming an automated local-folder rename.

## Assumptions

- Stacklane is the approved replacement name for this feature and will not be reopened during implementation unless the scope is explicitly changed.
- The local containing folder rename will be performed manually by the operator and must be documented, not automated.
- Existing runtime behavior, project isolation, and configuration precedence remain materially the same unless a change is required to support the unified command UX.
- GUI assets and automation wrappers remain in scope as user-facing surfaces even if they still lag behind full CLI parity in other areas.
- Historical references to `20i` may remain only where needed to explain migration from the previous name and command set.
- Existing `legacy wrapper commands` commands remain available temporarily as migration wrappers and are expected to surface deprecation guidance toward `stacklane`.
