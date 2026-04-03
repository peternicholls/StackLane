# Feature Specification: Extended Central YAML Configuration

**Feature Branch**: `002-extended-yaml-config`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟢 Medium  
**Input**: User description: "Add more stack variables to config/stack-vars.yml for network, ports, database, and optional services"

## Configuration Precedence *(mandatory)*

When the system resolves configuration values, it MUST apply precedence in this order:

1. **Process environment** (explicit `export VAR=...` in the user’s shell)
2. **Project-local overrides** (`.env` and/or `.20i-local` if supported by the stack)
3. **Central defaults** (`config/stack-vars.yml`)
4. **Hardcoded defaults** (only when the above do not define a value)

If multiple sources define the same key, the higher-precedence source MUST win.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Network Ports Centrally (Priority: P1)

As a developer, I want to set default port mappings in a central configuration file so that I can avoid port conflicts and maintain consistent defaults across projects.

**Why this priority**: Port conflicts are a common pain point; centralized defaults reduce friction for new setups.

**Independent Test**: Set `HOST_PORT: "8080"` in `config/stack-vars.yml`, start a new project without local overrides, and verify the web server is accessible on port 8080.

**Acceptance Scenarios**:

1. **Given** `HOST_PORT: "8080"` in `stack-vars.yml`, **When** the stack starts without local override, **Then** the web server listens on port 8080
2. **Given** `MYSQL_PORT: "3307"` in `stack-vars.yml`, **When** the stack starts, **Then** MariaDB is accessible on port 3307
3. **Given** `PMA_PORT: "8082"` in `stack-vars.yml`, **When** the stack starts, **Then** phpMyAdmin is accessible on port 8082

#### Port resolution contract

- Ports MUST be treated as integers in the range 1–65535.
- If a port is configured as a string in YAML (e.g. `"8080"`), the system MUST still parse it as a port number.
- If an invalid port is configured, the system MUST:
  - Refuse to start the stack
  - Return an actionable error message that identifies the invalid key and value

---

### User Story 2 - Set Default Database Credentials (Priority: P2)

As a developer, I want to configure default database credentials centrally so that new projects have sensible defaults without manual configuration.

**Why this priority**: Database setup is required for most projects; defaults reduce initial setup time.

**Independent Test**: Start a new project without `.env` file, connect to the database using credentials from `stack-vars.yml`, and verify connection succeeds.

**Acceptance Scenarios**:

1. **Given** `MYSQL_DATABASE: "myapp"` in `stack-vars.yml`, **When** the stack starts, **Then** a database named "myapp" is created
2. **Given** `MYSQL_USER: "appuser"` and `MYSQL_PASSWORD: "secret"` in `stack-vars.yml`, **When** the stack starts, **Then** the user "appuser" can authenticate with password "secret"
3. **Given** a project `.env` overrides `MYSQL_DATABASE`, **When** the stack starts, **Then** the local override takes precedence

#### Credentials and security notes

- The central YAML file is intended for local development defaults.
- The system MUST support local overrides for credentials per project.
- Documentation MUST warn users not to commit real production secrets into `stack-vars.yml`.

---

### User Story 3 - Toggle Optional Services (Priority: P3)

As a developer, I want to enable or disable optional services like Redis and Mailhog via configuration so that I can include only the services I need.

**Why this priority**: Optional services add flexibility without bloating the default stack.

**Independent Test**: Set `ENABLE_REDIS: true` in `stack-vars.yml`, start the stack, and verify Redis container is running and accessible.

**Acceptance Scenarios**:

1. **Given** `ENABLE_REDIS: false` (default), **When** the stack starts, **Then** no Redis container is started
2. **Given** `ENABLE_REDIS: true`, **When** the stack starts, **Then** Redis container starts and is accessible on port 6379
3. **Given** `ENABLE_MAILHOG: true`, **When** the stack starts, **Then** Mailhog container starts with SMTP on 1025 and UI on 8025

#### Optional services contract

- Optional services MUST be disabled by default.
- Enabling an optional service MUST:
  - Start the service container
  - Expose documented ports
  - Integrate cleanly with `docker compose down` teardown
- Disabling an optional service MUST ensure it is not started, and does not leave orphaned containers.

---

### Edge Cases

- Invalid YAML syntax in `stack-vars.yml` (must produce a clear parse error)
- `stack-vars.yml` missing entirely (must fall back to hardcoded defaults)
- Missing keys in `stack-vars.yml` (must fall back to hardcoded defaults for those keys)
- Port number outside valid range (1–65535) (must refuse to start with actionable error)
- Key present but empty string (must treat as missing and fall back)
- Boolean toggles specified as strings (e.g. `"true"`) (must parse correctly)
- Unknown keys present (must ignore safely, but SHOULD warn in logs)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST read network port configuration (`HOST_PORT`, `MYSQL_PORT`, `PMA_PORT`) from `config/stack-vars.yml`
- **FR-002**: System MUST read database credentials (`MYSQL_ROOT_PASSWORD`, `MYSQL_DATABASE`, `MYSQL_USER`, `MYSQL_PASSWORD`) from `config/stack-vars.yml`
- **FR-003**: System MUST read optional service toggles (`ENABLE_REDIS`, `ENABLE_MAILHOG`) from `config/stack-vars.yml`
- **FR-004**: Local `.env` file values MUST override central YAML defaults
- **FR-005**: System MUST provide sensible hardcoded defaults when YAML variables are missing
- **FR-005a**: System MUST validate YAML types and coerce common representations safely (e.g. numeric strings for ports, string booleans for toggles)
- **FR-005b**: On invalid YAML or invalid values, system MUST fail fast with an actionable error including key name and value
- **FR-006**: System MUST validate port numbers are within valid range (1-65535)
- **FR-007**: Documentation MUST be updated to list all available configuration variables
- **FR-008**: System MUST maintain backward compatibility with existing `.env` files

### Key Entities

- **Stack Variables**: Central configuration stored in `config/stack-vars.yml`, with categories: network/ports, database, optional services
- **Local Overrides**: Project-specific settings in `.env` that take precedence over central defaults
- **Optional Services**: Container services that can be enabled/disabled via configuration flags

## Non-goals *(mandatory)*

- This feature does NOT replace `.env` workflows; it only adds a richer central defaults layer.
- This feature does NOT introduce complex templating or conditionals inside YAML.
- This feature does NOT require optional services to be installed or enabled unless explicitly toggled.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: New projects require zero manual configuration to start with sensible defaults
- **SC-002**: Developers can customize any port in under 30 seconds by editing one file
- **SC-003**: Optional services can be enabled with a single configuration change
- **SC-004**: 100% backward compatibility with existing projects using `.env` files
- **SC-005**: README documentation covers all new variables within same release

## Assumptions

- YAML format is familiar to target developers
- Default values represent common development scenarios
- Optional services (Redis, Mailhog) use standard ports
- Users benefit from explicit documentation of precedence rules and common override patterns
