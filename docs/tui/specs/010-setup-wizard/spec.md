# Feature Specification: Interactive Setup Wizard (Command: 20i)

**Feature Branch**: `010-setup-wizard`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟢 Medium  
**Input**: User description: "Guided setup for first-time users with project type detection, preferences, and automatic configuration"

## Product Contract *(mandatory)*

The setup wizard is a guided UX layer over existing deterministic configuration rules. It MUST NOT introduce new stack behaviours.

### Configuration and storage

- All project configuration MUST be persisted in `.20i-config.yml` in the project root.
- The wizard MUST NOT store user preferences, UI state, or caches inside project directories.
- Any user-level preferences or cached state MUST live in `~/.20i/` (macOS/Linux).

### Determinism and precedence

Wizard choices MUST map to the same configuration outcomes as non-interactive usage.

Configuration precedence MUST follow the global rules:

0. CLI flags  
1. Process environment  
2. Project-local overrides (e.g. `.env` / `.20i-local` where supported)  
3. Project config file (`.20i-config.yml`)  
4. Central defaults (`config/stack-vars.yml`)  
5. Hardcoded defaults

### Non-interactive environments

- If no TTY is available, the wizard MUST NOT prompt.
- In non-interactive mode, `20i init` MUST either:
  - use defaults, OR
  - fail fast with an actionable message explaining which flags are required.

### Core services *(non-negotiable)*

The wizard MUST always configure the core 20i services as mandatory:

- Nginx
- PHP-FPM
- MariaDB
- phpMyAdmin

Optional services are additive and MUST be non-blocking by default.

### Optional services mechanism

- Optional services MUST be implemented via Docker Compose profiles (see spec 007).
- Wizard service selection MUST persist enabled service profiles into `.20i-config.yml`.

### Templates

- Project type detection MUST suggest templates (spec 008).
- Applying a template MUST NOT require network access in the MVP (embedded assets or `STACK_HOME`).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Guided First-Time Setup (Priority: P1)

As a first-time user, I want to be guided through setup with interactive prompts so that I can configure the stack without reading documentation.

**Why this priority**: First-time experience is critical for adoption; a smooth wizard reduces friction for new users.

**Independent Test**: Run `20i init` in an empty directory without any flags, verify interactive prompts guide the user and `.20i-config.yml` is written deterministically from the choices.

**Acceptance Scenarios**:

1. **Given** a first-time user runs `20i init`, **When** the command starts, **Then** an interactive wizard begins with welcome message  
2. **Given** the wizard is running, **When** user completes all prompts, **Then** `.20i-config.yml` is written and no other project files are modified unexpectedly  
3. **Given** the wizard completes, **When** user runs `20i start`, **Then** stack starts with the configured settings

#### Init safety rules

- `20i init` MUST be idempotent.
- The wizard MUST NOT overwrite an existing `.20i-config.yml` without explicit confirmation (or `--force`).
- If the directory is non-empty, the wizard MUST refuse to initialize unless `--force` is provided.

---

### User Story 2 - Automatic Project Type Detection (Priority: P2)

As a developer with an existing project, I want the wizard to detect my project type so that it suggests appropriate configuration.

**Why this priority**: Auto-detection reduces manual choices and shows intelligence that builds trust.

**Independent Test**: Run `20i init` in a directory with `composer.json` containing Laravel, verify wizard suggests Laravel template.

**Acceptance Scenarios**:

1. **Given** a directory with Laravel `composer.json`, **When** wizard runs, **Then** it suggests the Laravel template (spec 008) and explains why (e.g. composer.json dependencies)  
2. **Given** a directory with WordPress files, **When** wizard runs, **Then** it suggests WordPress template  
3. **Given** an empty directory, **When** wizard runs, **Then** it asks user to select project type manually

#### Detection contract

- Detection MUST be best-effort and MUST NOT change stack behaviour automatically.
- The wizard MUST present a suggested template with an explicit option to override.

---

### User Story 3 - Port Conflict Detection (Priority: P3)

As a developer with multiple projects, I want the wizard to detect port conflicts so that I don't have to troubleshoot startup failures.

**Why this priority**: Port conflicts are a common frustration; proactive detection prevents issues.

**Independent Test**: Run `20i init` while another service uses port 80, verify wizard warns about conflict and suggests alternative.

**Acceptance Scenarios**:

1. **Given** port 80 is in use, **When** wizard runs, **Then** it warns about the conflict  
2. **Given** port conflict detected, **When** user is prompted, **Then** an alternative port is suggested and persisted to `.20i-config.yml`  
3. **Given** no port conflicts, **When** wizard runs, **Then** default ports are suggested without warnings

#### Port contract

- Port checks MUST be advisory; the user can still choose a conflicting port (with a clear warning).
- Invalid ports MUST be rejected (range 1–65535) with an actionable message.

---

### User Story 4 - Optional Services Selection (Priority: P4)

As a developer, I want to select optional services during setup so that my stack includes everything I need from the start.

**Why this priority**: Service selection during init is convenient but not essential for basic setup.

**Independent Test**: Run wizard, select Redis and Mailhog, verify both services are enabled as Compose profiles in `.20i-config.yml`.

**Acceptance Scenarios**:

1. **Given** wizard reaches services step, **When** user selects Redis, **Then** the Redis profile is recorded as enabled in `.20i-config.yml`  
2. **Given** wizard reaches services step, **When** user selects multiple services, **Then** all selected service profiles are recorded as enabled  
3. **Given** user skips optional services, **When** wizard completes, **Then** only core services are configured

---

### User Story 5 - Skip Wizard with Defaults (Priority: P5)

As an experienced user, I want to skip the wizard and use defaults so that I can set up quickly without prompts.

**Why this priority**: Experienced users shouldn't be slowed down by interactive prompts.

**Independent Test**: Run `20i init --no-wizard`, verify `.20i-config.yml` is created using defaults without any prompts.

**Acceptance Scenarios**:

1. **Given** the `--no-wizard` flag, **When** the user runs `20i init --no-wizard`, **Then** no prompts are shown  
2. **Given** `--no-wizard` is used, **When** initialization completes, **Then** default configuration is applied  
3. **Given** both `--no-wizard` and `--template laravel`, **When** init runs, **Then** Laravel template is used without prompts

#### Non-interactive contract

- `--no-wizard` MUST be safe for CI (no prompts).
- If `--no-wizard` is not provided and no TTY is available, `20i init` MUST fail fast with an actionable message recommending `--no-wizard` and/or required flags.

---

### Edge Cases

- No TTY available (must not prompt; must use `--no-wizard` or fail with actionable guidance)
- User cancels wizard (Ctrl+C) (must exit cleanly with no partial/ambiguous config; if `.20i-config.yml` was written, it must be valid)
- Invalid input (e.g. non-numeric port) (must re-prompt interactively; must fail fast non-interactively)
- Port conflict detected (must warn; must allow override)
- Existing `.20i-config.yml` present (must confirm before overwrite; support `--force`)
- Running wizard in a non-empty directory (must refuse unless `--force`)
- Detection is wrong (must allow override easily)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Wizard MUST guide users through creation of a valid `.20i-config.yml` using deterministic prompts  
- **FR-002**: Wizard MUST guide users through database configuration (name, credentials) and persist it in `.20i-config.yml`  
- **FR-003**: Wizard MUST guide users through port selection with conflict detection and validation  
- **FR-004**: Wizard MUST guide users through optional services selection and persist enabled services as Compose profiles  
- **FR-005**: Wizard MUST detect project type from existing files when possible  
- **FR-005a**: Wizard MUST suggest templates based on detection but MUST allow easy override and MUST not require network access in the MVP  
- **FR-006**: Wizard MUST validate all user inputs before proceeding  
- **FR-006a**: Wizard MUST enforce init safety rules (idempotent; no overwrite without confirmation; refuse non-empty dirs unless `--force`)  
- **FR-007**: Wizard MUST support `--no-wizard` flag to use defaults without prompts  
- **FR-008**: Wizard MUST provide clear feedback and progress indication  
- **FR-009**: Wizard MUST behave safely in non-interactive environments (no prompts; defaults or actionable failure)  

### Key Entities

- **Setup Wizard**: Interactive command-line interface that guides users through configuration choices  
- **Project Detection**: Logic that analyzes directory contents to identify framework/project type  
- **Port Conflict Checker**: Utility that scans system for ports in use and reports conflicts  
- **Project Config**: `.20i-config.yml` stored in the project root, containing project-scoped settings  
- **User State Directory**: `~/.20i/` used for per-user preferences and caches that must not live in projects  

## Non-goals *(mandatory)*

- This feature does NOT store user preferences inside project directories; user preferences belong in `~/.20i/`.
- This feature does NOT require network access for templates in the MVP.
- This feature does NOT change the underlying stack behaviour; it only guides configuration.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Non-technical users can complete setup without reading documentation  
- **SC-002**: Wizard completion time is tracked in user testing; the flow is designed to be quick with sensible defaults (target: typically < 2 minutes)  
- **SC-003**: Project type detection is accurate for 90%+ of Laravel, WordPress, and Symfony projects  
- **SC-004**: Port conflicts are detected and reported before they cause startup failures  
- **SC-005**: 95%+ of first-time users successfully complete wizard and start stack  
- **SC-006**: `20i init --no-wizard` produces a valid `.20i-config.yml` deterministically and completes quickly on a typical developer machine  

## Assumptions

- Users have basic terminal familiarity (can type responses and press enter)  
- Terminal supports standard input/output for interactive prompts  
- Common frameworks have identifiable project files (composer.json, wp-config.php, etc.)  
- Port scanning is available and permitted on the host system
