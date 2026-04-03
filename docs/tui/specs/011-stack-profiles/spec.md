# Feature Specification: Stack Profiles (User Presets) (Command: 20i)

**Feature Branch**: `011-stack-profiles`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟢 Medium  
**Input**: User description: "Save and switch between multiple dev presets (e.g. older PHP versions) as a convenience layer without creating a second config system"

## Product Contract *(mandatory)*

Stack Profiles are a **user convenience layer** (named presets) that apply a set of overrides to the project’s canonical configuration.

- The **single source of truth** for project configuration remains `.20i-config.yml` in the project root.
- Profiles MUST NOT introduce a parallel project configuration system.
- Profiles MUST be **per-user**, not per-repo.

### Storage rules

- Profile presets MUST be stored in the user state directory: `~/.20i/profiles/`.
- The CLI MUST NOT store profile presets inside project directories.
- The CLI MAY store user-level metadata (timestamps, last-used, etc.) in `~/.20i/`.

### Application semantics

- `20i profile use <name>` MUST apply the preset by updating `.20i-config.yml` (write-through).
- The project MUST remain fully functional even if the user has no profiles.
- Profiles are **not portable by default**; they are intentionally user-local.
- Presets MUST NOT modify project identity metadata such as the selected template (if present) in `.20i-config.yml`.

### Determinism and precedence

When applying a profile, precedence MUST be:

0. CLI flags  
1. Process environment  
2. Project-local overrides (e.g. `.env` / `.20i-local` where supported)  
3. Project config file (`.20i-config.yml`)  
4. Active profile preset overrides (applied only during `profile use` / `profile apply`)  
5. Central defaults (`config/stack-vars.yml`)  
6. Hardcoded defaults

Notes:

- Profile presets MUST NOT silently override settings at runtime; they are applied explicitly via commands.
- The resulting `.20i-config.yml` MUST be valid and deterministic.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Save Current Settings as a User Preset (Priority: P1)

As a developer, I want to save my current stack configuration as a named preset so that I can quickly apply it later as a convenience layer.

**Why this priority**: Saving presets is the foundation - without it, presets cannot be created or applied.

**Independent Test**: Configure stack with PHP 8.3 and Redis enabled, run `20i profile save dev-php83`, verify a preset file is created in `~/.20i/profiles/` capturing only the relevant overrides.

**Acceptance Scenarios**:

1. **Given** a configured project, **When** the user runs `20i profile save dev-php83`, **Then** a preset is written to `~/.20i/profiles/dev-php83.yml`
2. **Given** a preset name already exists, **When** saving with the same name, **Then** the system prompts for confirmation to overwrite
3. **Given** the `--force` flag, **When** saving an existing preset name, **Then** the preset is overwritten without prompt

---

### User Story 2 - Apply a Preset to the Project (Priority: P1)

As a developer testing compatibility, I want to apply presets so that I can quickly change PHP versions or service configurations during development.

**Why this priority**: Applying a preset is the primary workflow for quickly switching PHP versions or enabled services during development.

**Independent Test**: Create presets for PHP 8.3 and PHP 8.4, apply each one, verify `.20i-config.yml` updates and the stack restarts only when changes require it.

**Acceptance Scenarios**:

1. **Given** presets `dev-php83` and `dev-php84` exist, **When** the user runs `20i profile use dev-php84`, **Then** the preset overrides are applied by updating `.20i-config.yml`
2. **Given** the applied preset changes stack-relevant settings (e.g. PHP version), **When** `profile use` completes, **Then** the stack is restarted (or the user is instructed to run `20i restart`)
3. **Given** the applied preset does not change stack-relevant settings, **When** `profile use` completes, **Then** no restart is performed and the command exits successfully
4. **Given** the user runs `20i profile use unknown`, **When** the preset does not exist, **Then** the CLI lists available presets and exits non-zero

---

### User Story 3 - List Available Profiles (Priority: P2)

As a developer, I want to see all my saved presets so that I can choose which one to use.

**Why this priority**: Listing enables discovery of available presets for selection.

**Independent Test**: Create multiple presets, run `20i profile list`, verify all presets are shown with basic info.

**Acceptance Scenarios**:

1. **Given** multiple presets exist, **When** the user runs `20i profile list`, **Then** all preset names are displayed
2. **Given** preset listing, **When** viewing output, **Then** each preset shows key settings (PHP version, enabled services)
3. **Given** a preset was last applied recently, **When** listing presets, **Then** the most recently used preset is marked/highlighted

---

### User Story 4 - Delete Profiles (Priority: P3)

As a developer, I want to delete presets I no longer need so that I can keep my preset list clean.

**Why this priority**: Deletion is maintenance functionality, less critical than core preset operations.

**Independent Test**: Run `20i profile delete dev-old`, verify preset is removed from user storage.

**Acceptance Scenarios**:

1. **Given** a preset exists, **When** the user runs `20i profile delete dev-old`, **Then** the preset is removed
2. **Given** a preset was applied previously, **When** user deletes it, **Then** the preset file is removed from `~/.20i/profiles/` (deleting does not modify `.20i-config.yml`)
3. **Given** a non-existent preset name, **When** trying to delete, **Then** error message indicates preset not found

---

### User Story 5 - Profile Metadata and History (Priority: P4)

As a developer, I want presets to track when they were created and last used so that I can identify stale presets.

**Why this priority**: Metadata is helpful for management but not essential for core functionality.

**Independent Test**: Create a preset, use it, run `20i profile list --details`, verify creation and last-used timestamps are shown.

**Acceptance Scenarios**:

1. **Given** a preset is created, **When** viewing preset details, **Then** creation timestamp is shown
2. **Given** a preset is applied, **When** viewing preset details later, **Then** last-used timestamp is updated
3. **Given** `--details` flag on list command, **When** listing presets, **Then** extended metadata is displayed

---

### Edge Cases

- Applying a preset while the stack is running (must restart only if changes require it)
- Preset file corrupted or invalid YAML (must fail fast with actionable guidance)
- Preset references services/profiles that are no longer supported (must refuse and list valid options)
- Permission errors writing to `~/.20i/` (must fail fast with actionable guidance)
- Applying a preset results in an invalid `.20i-config.yml` (must refuse and leave existing config intact)
- Concurrent operations (two `profile use` commands) (must serialize deterministically)
- Preset attempts to include or apply a template field (must ignore or refuse; template identity remains unchanged)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: CLI MUST support `20i profile save <name>` to save current configuration as named preset
- **FR-002**: CLI MUST support `20i profile use <name>` to apply a saved preset
- **FR-003**: CLI MUST support `20i profile list` to display all saved presets
- **FR-004**: CLI MUST support `20i profile delete <name>` to remove a preset
- **FR-005**: Presets MUST capture relevant overrides for PHP version, enabled services/profiles, port mappings, and database settings (as needed)
- **FR-005a**: Presets MUST NOT change template identity/selection fields in `.20i-config.yml` (templates are set by `20i init` and remain stable)
- **FR-007**: Presets MUST store creation timestamp and last-used timestamp
- **FR-008**: Preset storage MUST be in the user state directory: `~/.20i/profiles/`
- **FR-009**: Applying a preset MUST update `.20i-config.yml` (write-through) and MUST NOT require storing an "active preset" pointer in the project
- **FR-010**: Applying a preset MUST trigger a stack restart only when the applied changes require it

### Key Entities

- **Preset (Profile)**: Named set of override values stored per-user (e.g. older PHP version) used to update project config on demand
- **Preset Storage**: User state directory `~/.20i/profiles/` containing preset YAML files and metadata
- **Project Config**: `.20i-config.yml` in the project root (single source of truth)

## Non-goals *(mandatory)*

- Presets are not intended to be committed to repos; they are user-local.
- This feature does NOT create a second project configuration system.
- This feature does NOT store user preferences or preset files inside project directories.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Applying a preset updates `.20i-config.yml` deterministically and completes quickly on a typical developer machine (restart time excluded)
- **SC-002**: Presets correctly restore all saved override settings 100% of the time
- **SC-003**: Developers can switch between PHP versions without editing files
- **SC-004**: Preset list command provides enough information to choose the right preset
- **SC-005**: Preset files remain small and human-readable; size is tracked per release and regressions are justified

## Assumptions

- Users have a limited number of presets per user (typically under 20)
- Profile names follow simple naming conventions (alphanumeric, hyphens)
- Configuration format is stable between minor versions
- File system storage is reliable for preset persistence

---

### Implementation Note

This feature is intentionally designed as a **convenience layer**, not a configuration system. A common use case is temporarily testing against an older or alternative PHP version:

- Applying a preset updates `.20i-config.yml`
- If the change affects stack-relevant settings (e.g. PHP version), a **simple stack restart** is sufficient
- Users should not need to re-run the setup wizard or recreate project configuration

This keeps experimentation fast while preserving a single, canonical project configuration.
