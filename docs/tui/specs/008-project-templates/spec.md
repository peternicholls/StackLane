# Feature Specification: Project Templates (Command: 20i)

**Feature Branch**: `008-project-templates`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟢 Medium  
**Input**: User description: "Quick-start templates for common frameworks like Laravel, WordPress, Symfony, and plain PHP"

## Product Contract *(mandatory)*

Project templates provide a **repeatable starting point** for common PHP frameworks while keeping the core stack behaviour unchanged.

- Templates MUST be **opt-in** (except the default plain PHP template).
- Template application MUST be **deterministic**: given the same template name and the same project directory, the resulting project configuration MUST be the same.
- Templates MUST NOT require users to manually edit Compose files.

### Where configuration lives

- The selected template MUST be persisted in the project’s `.20i-config.yml`.
- User-level preferences, caches, and UI state MUST live in `~/.20i/` and MUST NOT be stored inside project directories.
- The selected template is **project identity metadata** and MUST remain stable after initialization.
- Post-init convenience layers (e.g. user presets in spec 011) MUST NOT modify the template selection field in `.20i-config.yml`.

### Composition strategy

- Templates MUST use Docker Compose **profiles** (see spec 007) to enable framework-related auxiliary services (e.g. Laravel queue worker).
- Templates MUST NOT generate or mutate compose override files in the MVP.

### Template delivery

- The default mechanism MUST be embedded template assets within the `20i` binary (or within `STACK_HOME` when running from source).
- Template initialization MUST NOT require network access.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Initialize with Framework Template (Priority: P1)

As a developer starting a new project, I want to initialize the stack with a framework-specific template so that I have optimal configuration and boilerplate for that framework.

**Why this priority**: Template initialization is the entry point - without it, templates provide no value.

**Independent Test**: Run `20i init --template laravel`, verify Laravel-specific project configuration is written to `.20i-config.yml`, required directories are created, and framework profiles/services are enabled when the stack starts.

**Acceptance Scenarios**:

1. **Given** an empty directory, **When** the user runs `20i init --template laravel`, **Then** Laravel template is selected, persisted in `.20i-config.yml`, and the project is initialized  
2. **Given** Laravel template is selected, **When** the stack is started, **Then** Laravel-related profiles (queue worker and scheduler) are activated  
3. **Given** `20i init --template wordpress`, **When** initialization completes, **Then** WordPress template is selected and persisted in `.20i-config.yml` and WP tooling is enabled via profiles where applicable

#### Init safety rules

- `20i init` MUST refuse to apply a template into a non-empty directory unless `--force` is provided.  
- `20i init` MUST be idempotent.  
- `20i init` MUST NOT overwrite an existing `.20i-config.yml` without explicit confirmation (or `--force`).
- Changing the selected template after init MUST require an explicit re-initialization workflow (e.g. `20i init --template <name> --force`) and MUST NOT happen implicitly.

---

### User Story 2 - List Available Templates (Priority: P2)

As a developer, I want to see what templates are available so that I can choose the right one for my project.

**Why this priority**: Template discovery helps users make informed choices before initialization.

**Independent Test**: Run `20i templates`, verify list shows all available templates with descriptions.

**Acceptance Scenarios**:

1. **Given** the CLI is installed, **When** the user runs `20i templates`, **Then** a list of available templates is displayed  
2. **Given** template listing, **When** viewing each template, **Then** a brief description and included features are shown  
3. **Given** a specific template, **When** the user runs `20i templates laravel --details`, **Then** detailed information about the template is shown

#### Listing contract

- `20i templates` MUST list templates available from embedded assets (or `STACK_HOME` when running from source).  
- Listing MUST NOT require network access.

---

### User Story 3 - Use Plain PHP Template (Priority: P3)

As a developer working on a custom PHP project, I want a plain PHP template so that I get a clean setup without framework-specific overhead.

**Why this priority**: Plain PHP is the default/fallback template for users not using a framework.

**Independent Test**: Run `20i init` without template flag, verify a minimal PHP configuration is created.

**Acceptance Scenarios**:

1. **Given** no template specified, **When** the user runs `20i init`, **Then** plain PHP template is used by default  
2. **Given** plain PHP template, **When** viewing configuration, **Then** only core services (Nginx, PHP-FPM, MariaDB, phpMyAdmin) are configured  
3. **Given** `20i init --template php`, **When** initialization completes, **Then** same result as no template flag

---

### User Story 4 - Template Includes Framework Conveniences (Priority: P4)

As a Laravel developer, I want artisan command shortcuts so that I can run common commands without entering the container.

**Why this priority**: Conveniences enhance developer experience but aren't required for basic functionality.

**Independent Test**: Initialize with Laravel template, run `20i artisan migrate`, verify command executes inside container and returns output.

**Acceptance Scenarios**:

1. **Given** Laravel template initialized, **When** the user runs `20i artisan migrate`, **Then** artisan command runs inside the PHP container  
2. **Given** WordPress template initialized, **When** the user runs `20i wp plugin list`, **Then** WP-CLI command runs inside the PHP container  
3. **Given** Symfony template initialized, **When** the user runs `20i console cache:clear`, **Then** Symfony console command executes

#### Convenience command contract

- Convenience commands MUST be thin wrappers over `docker compose exec` (no bespoke container logic).  
- Commands MUST be scoped to the current project stack identity.  
- If the required service/container is not running, the CLI MUST exit non-zero with an actionable message (e.g. "Run `20i start` first").

---

### Edge Cases

- Template requested does not exist (must list available templates and exit non-zero)  
- Applying a template in a non-empty directory (must refuse unless `--force`)  
- `.20i-config.yml` invalid or corrupted (must refuse and suggest re-init/repair)  
- Template requires a service profile not supported by the current stack version (must fail with actionable guidance)  
- Convenience command invoked when stack is stopped (must fail with actionable guidance)  
- WordPress or Laravel files already present (must refuse unless `--force` or offer a safe plan)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: CLI MUST support `20i init --template <name>` to initialize with specific template  
- **FR-001a**: CLI MUST support `20i init --template <name> --force` to allow initialization in non-empty directories  
- **FR-002**: CLI MUST support `20i templates` to list available templates  
- **FR-003**: System MUST provide a Laravel template that enables Laravel-related service profiles (e.g. queue worker and scheduler)  
- **FR-004**: System MUST provide a WordPress template that enables WordPress tooling (e.g. WP-CLI) via profiles where applicable  
- **FR-005**: System MUST provide a Symfony template that enables Symfony console conveniences and any required profiles  
- **FR-006**: System MUST provide plain PHP template as default  
- **FR-007**: Templates MUST include framework-specific environment variables  
- **FR-008**: Templates MUST register convenience commands for common framework operations (implemented as compose-exec wrappers)  
- **FR-009**: Template selection MUST be persisted in `.20i-config.yml`  
- **FR-009a**: The template selection field in `.20i-config.yml` MUST be treated as stable project identity metadata and MUST NOT be changed by non-init commands  
- **FR-010**: Templates MUST be delivered from embedded assets (or `STACK_HOME` when running from source) and MUST NOT require network access  
- **FR-011**: Templates MUST NOT generate or mutate Compose override files in the MVP

### Key Entities

- **Project Template**: Predefined configuration set for a specific framework, includes project configuration, environment variables, enabled Compose profiles, and convenience commands  
- **Template Metadata**: Information about a template including name, description, included services, and framework version compatibility  
- **Convenience Command**: CLI shortcut that executes framework-specific commands inside containers  

## Non-goals *(mandatory)*

- This feature does NOT download templates from the internet in the MVP.  
- This feature does NOT attempt to support every framework version; templates target current stable versions.  
- This feature does NOT generate or mutate Compose override files in the MVP.  
- This feature does NOT store user preferences or caches in project directories.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Template initialization is a single command and results in a ready-to-start project (time varies by disk and template size)  
- **SC-002**: Developers can start coding in their framework within 5 minutes of init  
- **SC-003**: Framework-specific services start correctly on 100% of template initializations  
- **SC-004**: Convenience commands reduce common operations to single CLI commands  
- **SC-005**: Template documentation is accessible via `20i templates <name> --help`

## Assumptions

- Users know which framework they want to use before initialization  
- Templates target current stable versions of each framework  
- Framework-specific services use standard configurations  
- Templates are updated when major framework versions are released
