# Feature Specification: Health Checks and Auto-Restart (Command: 20i)

**Feature Branch**: `009-health-checks`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟡 High  
**Input**: User description: "Ensure services are truly ready before marking stack as started with container health checks and optional auto-restart"

## Product Contract *(mandatory)*

Health checks define when the stack is considered **ready**. The `20i start` command MUST NOT report success until the stack meets the readiness contract (unless `--no-wait` is used).

### Readiness contract

- **Ready** means:
  - Required containers are running (core services are mandatory and blocking)
  - Containers with health checks report `healthy`
  - Containers without health checks are either:
    - explicitly marked as **non-blocking** for readiness, OR
    - verified via an explicit lightweight readiness probe
- Readiness MUST be evaluated deterministically: given the same config and container states, readiness decisions MUST be consistent.

#### Core services *(non-negotiable)*

The following services are core to the 20i shared hosting environment and are **mandatory**:

- Nginx
- PHP-FPM
- MariaDB
- phpMyAdmin

Rules:

- These services MUST always be treated as **blocking** for readiness.
- `20i start` MUST NOT report "ready" unless all core services meet readiness conditions.
- Failure of any core service MUST cause `20i start` to time out or exit non-zero with actionable diagnostics.
- Optional services (see spec 007) MUST be non-blocking by default unless explicitly marked as required by a template.

### Scope and storage

- Health-check tuning is **per-project** and MUST be stored in `.20i-config.yml`.
- User-level preferences (UI state, caches) MUST live in `~/.20i/` and MUST NOT be stored inside project directories.

### Configuration precedence

Configuration MUST follow the global precedence rules:

0. CLI flags  
1. Process environment  
2. Project-local overrides (e.g. `.env` / `.20i-local` where supported)  
3. Project config file (`.20i-config.yml`)  
4. Central defaults (`config/stack-vars.yml`)  
5. Hardcoded defaults

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Wait for Services to Be Ready (Priority: P1)

As a developer, I want the `20i start` command to wait until all services are truly ready so that I can immediately start using the stack without connection errors.

**Why this priority**: This is the core value - preventing premature "ready" state that leads to failed connections.

**Independent Test**: Run `20i start`, measure time from command to "ready" message, then immediately connect to web server and database - both should succeed.

**Acceptance Scenarios**:

1. **Given** a fresh stack start, **When** the user runs `20i start`, **Then** the command waits until readiness conditions pass (or a timeout occurs) before reporting "ready"
2. **Given** MariaDB is still initializing, **When** PHP tries to connect, **Then** the connection succeeds because start waits for MariaDB health check
3. **Given** the `--no-wait` flag, **When** the user runs `20i start --no-wait`, **Then** the command starts the stack and returns immediately without waiting for readiness

#### Waiting contract

- `20i start` MUST support a configurable maximum wait duration (default: 120s).
- If the timeout is reached, `20i start` MUST exit non-zero and print:
  - which service(s) failed readiness
  - their last known state (starting/healthy/unhealthy)
  - a suggested next step (e.g. `20i status`, `20i logs <service>`)
- `--no-wait` MUST NOT change whether health checks exist; it only changes whether `start` waits.

---

### User Story 2 - Display Clear Health Status (Priority: P2)

As a developer, I want to see the health status of all services so that I can quickly identify which service is having issues.

**Why this priority**: Visibility into health status enables faster troubleshooting.

**Independent Test**: Run `20i status` while stack is starting, verify each service shows its current health state (starting, healthy, unhealthy).

**Acceptance Scenarios**:

1. **Given** a running stack, **When** the user runs `20i status`, **Then** health status (healthy/unhealthy) is shown for each service
2. **Given** a service is unhealthy, **When** viewing status, **Then** the status output includes the service name, state, and a suggested next action (e.g. `20i logs <service>`)
3. **Given** health check is in progress, **When** viewing status, **Then** "starting" or "waiting" state is displayed

#### Status contract

- `20i status` MUST display health states derived from Docker health status where available.
- For services without health checks, `status` MUST clearly label them as `no-healthcheck` (or equivalent) and indicate whether they are blocking readiness.
- Core services MUST always be shown prominently in status output and clearly distinguished from optional services.

---

### User Story 3 - Auto-Restart Unhealthy Containers (Priority: P3)

As a developer, I want unhealthy containers to automatically restart so that temporary issues resolve without manual intervention.

**Why this priority**: Auto-restart improves resilience but is less critical than basic health checking.

**Independent Test**: Configure auto-restart, simulate container failure, verify container restarts automatically and becomes healthy.

**Acceptance Scenarios**:

1. **Given** auto-restart is enabled, **When** a container becomes unhealthy, **Then** the system attempts recovery according to the configured policy (restart with limits)
2. **Given** a restart limit of 3 attempts, **When** a container fails recovery 3 times, **Then** the system stops attempting recovery and surfaces an alert/state to the user
3. **Given** auto-restart is disabled, **When** a container becomes unhealthy, **Then** it remains in unhealthy state until manual intervention

#### Auto-restart contract

- Auto-restart MUST be implemented using a simple, well-understood mechanism:
  - Prefer Docker Compose `restart:` policies for crash/restart behaviour, AND
  - Use an explicit health-driven recovery loop only if required (must remain bounded).
- Auto-restart MUST NEVER fight `20i stop`:
  - If the user runs `20i stop`, the stack MUST stop and remain stopped.
  - Recovery loops/timers MUST be cancelled when stopping.

---

### User Story 4 - Configurable Health Check Parameters (Priority: P4)

As an advanced user, I want to configure health check intervals and thresholds so that I can tune the behavior for my environment.

**Why this priority**: Configuration is for advanced users; defaults should work for most cases.

**Independent Test**: Set custom health check interval in config, restart stack, verify health checks run at the configured interval.

**Acceptance Scenarios**:

1. **Given** default configuration, **When** health checks run, **Then** they use sensible defaults (10s interval, 3 retries)
2. **Given** custom interval of 5s in config, **When** health checks run, **Then** they execute every 5 seconds
3. **Given** custom retry count of 5, **When** a service fails, **Then** 5 attempts are made before marking unhealthy

#### Configuration keys *(MVP)*

Health check tuning MUST be configurable per-project in `.20i-config.yml`, with at minimum:

- `health.wait_timeout_seconds` (int)
- `health.interval_seconds` (int)
- `health.timeout_seconds` (int)
- `health.retries` (int)
- `health.auto_restart_enabled` (bool)
- `health.auto_restart_max_attempts` (int)

---

### Edge Cases

- Health check command errors or times out (must mark service unhealthy and include actionable guidance)
- Service has no defined health check (must be labeled `no-healthcheck` and treated as blocking or non-blocking per policy)
- Docker daemon not running (must fail fast with actionable message)
- `docker compose` unavailable (must fail fast with actionable message)
- User runs `20i start` while a previous start is still waiting (must refuse or coalesce deterministically)
- All services unhealthy simultaneously (must surface summary and avoid restart storms)
- User runs `20i stop` while auto-restart is active (must cancel recovery and stop cleanly)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Core service containers SHOULD have health checks defined in the Compose stack where feasible; services without health checks MUST be handled explicitly by readiness/status logic
- **FR-001a**: Nginx, PHP-FPM, MariaDB, and phpMyAdmin MUST be treated as mandatory core services and MUST always block readiness
- **FR-002**: `20i start` MUST wait for all health checks to pass before reporting "ready"
- **FR-002a**: `20i start` MUST enforce a maximum wait timeout and exit non-zero with actionable diagnostics on timeout
- **FR-003**: `20i status` MUST display health status for each service
- **FR-004**: System MUST support optional recovery of unhealthy containers with bounded limits and must never override `20i stop`
- **FR-005**: System MUST support `--no-wait` flag to skip health check waiting
- **FR-006**: System MUST display clear error messages when services fail health checks
- **FR-007**: Health check parameters MUST be configurable (interval, timeout, retries)
- **FR-008**: Auto-restart MUST have configurable attempt limits to prevent infinite loops
- **FR-009**: System MUST log health check failures for troubleshooting
- **FR-010**: Health and recovery configuration MUST be stored per-project in `.20i-config.yml` and MUST NOT require manual compose edits

### Key Entities

- **Health Check**: Configuration defining how to verify service readiness, with attributes: check command, interval, timeout, retries
- **Service Health State**: Current status of a service (starting, healthy, unhealthy)
- **Auto-Restart Policy**: Rules for automatic container restart, with attributes: enabled flag, max attempts, cooldown period

## Non-goals *(mandatory)*

- This feature does NOT attempt to be a full orchestration system; it provides readiness and bounded recovery for local development.
- This feature does NOT guarantee perfect application-level readiness (e.g. database schema migrations); it guarantees service/container readiness as defined.
- This feature does NOT store user preferences inside projects; user preferences belong in `~/.20i/`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Stack reports "ready" only when all services are accessible (zero false positives)
- **SC-002**: Additional wait time due to readiness checks is tracked per release; regressions are identified and justified (target: minimal added delay on healthy services)
- **SC-003**: Unhealthy services are identified within 30 seconds of failure
- **SC-004**: Auto-restart resolves 80%+ of transient failures without user intervention
- **SC-005**: Health status is visible within 1 second of running `20i status`
- **SC-006**: Zero "connection refused" errors when using stack immediately after "ready" message

## Assumptions

- Services have predictable startup times within configured health check windows
- Health check commands are lightweight and don't impact service performance
- Container orchestration supports health check and restart policies
- Network connectivity between containers is reliable
