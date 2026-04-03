# Feature Specification: Optional Service Modules

**Feature Branch**: `007-optional-services`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟢 Medium  
**Input**: User description: "Easy addition of common development services like Redis, Mailhog, Elasticsearch via CLI enable/disable commands"

## Product Contract *(mandatory)*

Optional services are an **add-on layer** to the core stack. They MUST be:

- **Opt-in** (disabled by default)
- **Per-project** (stored in `.20i-config.yml`)
- **Deterministic** (same project config yields the same enabled services and Compose configuration)
- **Non-invasive** (no manual editing of compose files required by the user)

### Where configuration lives

- Enabled optional services MUST be stored in the project’s `.20i-config.yml`.
- User-level preferences (UI state, last-used options, caches) MUST live in `~/.20i/` and MUST NOT be stored inside project directories.

### Module-aware behaviour *(mandatory)*

Optional services MUST be **module-aware**.

- A selected stack module (e.g. `20i`) may define which optional services exist and how they map to Compose profiles.
- Availability of optional services MAY differ by module (e.g. a minimal module may not offer Elasticsearch).
- Enabling/disabling services MUST remain **module-agnostic** at the CLI/TUI level (same commands), but the implementation MUST use the selected module’s definitions.
- User-installed modules MAY live in `~/.20i/modules/`, but optional service enablement state MUST remain in the project `.20i-config.yml`.

### Service module mechanism

- Each optional service MUST be defined using Docker Compose profiles, composed into the base stack without duplicating the base stack definition.
- Compose profiles MUST be defined within the selected module’s Compose file(s) (module-local).
- The mechanism MUST work in both CLI and TUI using the shared stack engine.

#### Composition strategy

- Docker Compose profiles are the canonical mechanism for optional services.
- Enabling a service maps to activating one or more Compose profiles.
- The stack engine MUST translate enabled services into the appropriate `--profile` selections when invoking `docker compose`.

#### Resolution flow

When starting the stack:

1. Determine the selected module for the project (from `.20i-config.yml`).
2. Load the module’s service catalog (service → Compose profile mapping).
3. Resolve enabled services from `.20i-config.yml` (and any explicit CLI flags).
4. Invoke `docker compose` for the module with the required `--profile` selections.

If an enabled service is not available in the selected module, the system MUST refuse with an actionable message.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Enable Optional Service via CLI (Priority: P1)

As a developer, I want to enable a service like Redis with a single command so that I can add caching to my project without manually editing configuration files.

**Why this priority**: Simple enablement is the core value proposition - reducing complexity of adding services.

**Independent Test**: Run `20i enable redis`, restart the stack, verify Redis container is running and accessible on port 6379.

**Acceptance Scenarios**:

1. **Given** a running stack without Redis, **When** the user runs `20i enable redis`, **Then** Redis is recorded as enabled in `.20i-config.yml` (idempotent)
2. **Given** Redis is enabled, **When** the stack restarts, **Then** Redis container starts and is accessible on port 6379
3. **Given** `20i enable mailhog`, **When** the stack restarts, **Then** Mailhog SMTP is available on port 1025 and web UI on port 8025
4. **Given** Redis is already enabled, **When** the user runs `20i enable redis` again, **Then** no duplicate config is created and the command exits successfully

---

### User Story 2 - Disable Optional Service via CLI (Priority: P2)

As a developer, I want to disable a service I no longer need so that I can free up resources and reduce stack complexity.

**Why this priority**: Disabling services completes the lifecycle management of optional services.

**Independent Test**: Run `20i disable redis` on stack with Redis enabled, restart stack, verify Redis container is not running.

**Acceptance Scenarios**:

1. **Given** Redis is enabled in the stack, **When** the user runs `20i disable redis`, **Then** Redis is removed from configuration
2. **Given** Redis is disabled, **When** the stack restarts, **Then** no Redis container is started
3. **Given** another enabled service depends on Redis, **When** the user runs `20i disable redis`, **Then** the CLI refuses with an actionable message (or offers `--force` to disable dependents)
4. **Given** the `--remove-data` flag, **When** disabling a service, **Then** associated volumes are also removed

---

### User Story 3 - List Available and Enabled Services (Priority: P3)

As a developer, I want to see which optional services are available and which are currently enabled so that I can make informed decisions.

**Why this priority**: Visibility into service state helps users manage their stack effectively.

**Independent Test**: Run `20i services`, verify output shows all available services with their enabled/disabled status.

**Acceptance Scenarios**:

1. **Given** a stack with Redis enabled, **When** the user runs `20i services`, **Then** Redis shows as "enabled" and other services show as "available"
2. **Given** no optional services enabled, **When** the user runs `20i services`, **Then** the output lists services available for the selected module with descriptions
3. **Given** enabled services are running, **When** the user runs `20i services --status`, **Then** the output includes container status and exposed ports (health if available)

---

### User Story 4 - Persist Service Configuration (Priority: P4)

As a developer, I want my enabled services to be remembered across stack restarts so that I don't need to re-enable them each time.

**Why this priority**: Persistence is expected behavior but is dependent on basic enable/disable working first.

**Independent Test**: Enable Redis, stop stack, start stack again, verify Redis is still running without re-enabling.

**Acceptance Scenarios**:

1. **Given** Redis is enabled, **When** the user runs `20i stop` then `20i start`, **Then** Redis starts automatically
2. **Given** service configuration in `.20i-config.yml`, **When** viewing the file, **Then** enabled services are listed
3. **Given** project is cloned fresh, **When** `.20i-config.yml` exists, **Then** enabled services start with the stack

#### Persistence contract

- Enabled services are part of the project configuration and MUST travel with the repo when `.20i-config.yml` is committed.
- The system MUST NOT write user-specific state into the project.
- `20i start` MUST apply enabled services automatically without requiring re-enable commands.

---

### Edge Cases

- Enabling a service with a port conflict (must refuse with actionable guidance and a suggested override)
- Enabling a service whose image cannot be pulled (must fail with actionable guidance; if local build is supported, it may suggest it)
- Disabling a service that has dependents (must refuse or require explicit `--force`)
- Enabling a service with required dependencies (must automatically enable dependencies OR prompt with a clear plan)
- Service already enabled/disabled (commands must be idempotent)
- `--remove-data` used when volumes do not exist (must succeed safely)
- Running `20i services --status` when Docker is unavailable (must provide actionable guidance)
- Enabled service not supported by the selected module (must refuse and list module-supported services)
- User switches module for a project (must require explicit re-init/migration flow; services must be re-validated)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: CLI MUST support `20i enable <service>` command to enable optional services
- **FR-002**: CLI MUST support `20i disable <service>` command to disable optional services
- **FR-003**: CLI MUST support `20i services` command to list available and enabled services
- **FR-004**: System MUST persist service enablement in `.20i-config.yml`
- **FR-004a**: System MUST treat `20i enable/disable` as idempotent operations
- **FR-004b**: System MUST handle dependencies between optional services (auto-enable dependencies OR refuse with actionable guidance)
- **FR-004c**: System MUST handle port conflicts with clear errors and a supported override mechanism
- **FR-004d**: System MUST implement optional services using Docker Compose profiles (no generated override compose files in MVP)
- **FR-005**: System MUST support Redis service (caching, port 6379)
- **FR-006**: System MUST support Mailhog service (email testing, SMTP 1025, UI 8025)
- **FR-007**: System MUST support Elasticsearch service (search, port 9200)
- **FR-008**: System MUST support RabbitMQ service (messaging, port 5672, UI 15672)
- **FR-009**: System MUST support MinIO service (S3 storage, port 9000)
- **FR-010**: Services MUST start automatically with stack when enabled

### Key Entities

- **Optional Service**: A container service that can be enabled/disabled per-project, with attributes: name, description, ports, dependencies
- **Service Configuration**: Persistent record of enabled services in `.20i-config.yml`
- **Service Module**: Module-local definition mapping an optional service to one or more Compose profiles within the selected module

## Non-goals *(mandatory)*

- This feature does NOT introduce production-grade hardening of optional services; defaults are for development use.
- This feature does NOT enable services globally across all projects; enablement is per-project.
- This feature does NOT store user preferences inside projects; user preferences belong in `~/.20i/`.
- This feature does NOT generate or mutate compose override files for optional services in the MVP.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Enabling a service is a single command and results in an updated `.20i-config.yml` plus a successful stack restart (time varies by image pull)
- **SC-002**: Disabling a service takes under 10 seconds
- **SC-003**: Service configuration persists across 100% of stack restarts
- **SC-004**: Zero manual file editing required to enable/disable services
- **SC-005**: All services start with stack in under 60 seconds after enable
- **SC-006**: Service documentation accessible via `20i help services`

## Assumptions

- Service images are available from public Docker registries
- Default ports don't conflict with host services (can be customized)
- Services have reasonable default configurations for development
- Users understand service purposes from brief descriptions
