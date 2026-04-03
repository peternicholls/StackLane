# Feature Specification: Performance Metrics and Insights (Command: 20i)

**Feature Branch**: `012-performance-metrics`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟢 Medium  
**Input**: User description: "Show resource usage, startup time, and actionable recommendations for optimization"

## Product Contract *(mandatory)*

Performance metrics provide **observability**, not control. This feature MUST NOT alter stack behaviour.

### Scope and behaviour

- Metrics are **read-only** and advisory.
- Metrics MUST be collected on-demand via user commands (no background daemons in the MVP).
- The system MUST prefer existing Docker APIs over bespoke instrumentation.

### Storage rules

- Real-time metrics MUST NOT be persisted by default.
- Optional historical metrics, if enabled, MUST be stored in the user state directory: `~/.20i/`.
- Metrics data MUST NOT be stored inside project directories.

### Determinism and performance

- Running `20i stats` MUST NOT materially affect container performance.
- Metrics output MUST be derived deterministically from the observed Docker state at invocation time.

### UI integration *(optional)*

- `20i stats` MAY open a focused view of the existing TUI dashboard, pre-selected to the metrics/stats section.
- CLI flags (e.g. `--history`, `--service <name>`, `--startup`) SHOULD map to narrowing/filtering the metrics view rather than introducing a separate UI mode.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Resource Usage per Service (Priority: P1)

As a developer, I want to see CPU and memory usage for each service so that I can identify resource-hungry containers.

**Why this priority**: Resource visibility is the core value - understanding where resources are consumed.

**Independent Test**: Run `20i stats` with stack running, verify CPU and memory usage is displayed for each container using Docker stats API.

**Acceptance Scenarios**:

1. **Given** a running stack, **When** the user runs `20i stats`, **Then** CPU and memory usage for each container is displayed
2. **Given** stats output, **When** viewing each service, **Then** current usage and limits (if set) are shown
3. **Given** a service using high resources, **When** stats are displayed, **Then** the high-usage service is highlighted or flagged

#### Metrics source contract

- CPU and memory metrics MUST be sourced from Docker's stats API.
- Only containers belonging to the current project stack MUST be included.

---

### User Story 2 - Track Startup Time (Priority: P2)

As a developer, I want to see how long stack startup takes so that I can identify slow components and optimize.

**Why this priority**: Startup time affects daily developer experience; tracking enables optimization.

**Independent Test**: Run `20i start`, verify startup time is displayed upon completion and recorded for future reference.

**Acceptance Scenarios**:

1. **Given** a stack starting, **When** startup completes, **Then** total startup time is displayed
2. **Given** startup time tracking, **When** user runs `20i stats`, **Then** last startup time and average are shown
3. **Given** individual service startup, **When** viewing detailed stats, **Then** per-service startup time (container start to healthy/ready) is available where determinable

#### Startup timing contract

- Stack startup time MUST measure from invocation of `20i start` to readiness (per spec 009).
- Per-service startup time SHOULD measure from container start to healthy/ready state.
- If health checks are unavailable, per-service startup time MUST be marked as approximate.

---

### User Story 3 - Receive Actionable Recommendations (Priority: P3)

As a developer, I want to receive recommendations for improving performance so that I can optimize my stack without research.

**Why this priority**: Recommendations add value by translating data into action, but require metrics first.

**Independent Test**: Configure MariaDB with low memory limit, run `20i stats`, verify recommendation is shown to increase limit.

**Acceptance Scenarios**:

1. **Given** MariaDB using 80%+ of memory limit, **When** stats are shown, **Then** recommendation to increase limit is displayed
2. **Given** startup takes over a configurable threshold, **When** stats are shown, **Then** recommendation to use pre-built images or review enabled services is displayed
3. **Given** disk usage is high, **When** stats are shown, **Then** recommendation to prune unused images/volumes is displayed

#### Recommendation contract

- Recommendations MUST be rule-based and explain *why* they are shown.
- Recommendations MUST NOT apply changes automatically.
- Recommendations MUST reference an explicit next action (command or config key).

---

### User Story 4 - View Historical Metrics (Priority: P4)

As a developer, I want to see historical resource usage so that I can identify trends and intermittent issues.

**Why this priority**: History is valuable for troubleshooting but less critical than real-time stats.

**Independent Test**: Enable history collection, run stack for extended period, run `20i stats --history`, verify historical data is displayed from user-local storage.

**Acceptance Scenarios**:

1. **Given** stack has been running for 1 hour, **When** the user runs `20i stats --history`, **Then** historical metrics are shown
2. **Given** historical data available, **When** viewing history, **Then** data is shown in time-series format (table or simple graph)
3. **Given** no historical data (first run), **When** history is requested, **Then** message indicates not enough data yet

#### History contract

- Historical metrics MUST be opt-in.
- History storage MUST be bounded (e.g. rolling window or capped size).
- If history is disabled or empty, the system MUST explain why.

---

### Edge Cases

- Docker runtime does not support stats API (must fail gracefully with actionable message)
- Stack is stopped when stats are requested (must explain no live data available)
- Containers restart frequently (must label averages as unstable or recent)
- Metrics collection fails for one container (must still show others with warning)
- History enabled but storage unavailable or permission denied (must fail fast with guidance)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: CLI MUST support `20i stats` command to display real-time resource usage
- **FR-001a**: `20i stats` MAY be implemented as a focused TUI entrypoint that defaults to the metrics section and applies requested filters
- **FR-002**: System MUST show CPU usage (percentage) per service
- **FR-003**: System MUST show memory usage (current/limit) per service
- **FR-004**: System MUST show disk usage for volumes associated with the current project stack
- **FR-005**: System MUST track and display startup time for stack and individual services
- **FR-006**: System MUST provide rule-based, non-invasive recommendations based on observed metrics
- **FR-007**: Recommendations MUST be contextual and specific to observed issues
- **FR-008**: System SHOULD track historical metrics for trend analysis
- **FR-009**: Stats command MUST work without additional tools/dependencies
- **FR-010**: Metrics and history storage MUST reside in user state directory (`~/.20i/`) and MUST NOT modify project files

### Key Entities

- **Service Metrics**: Real-time resource usage data for a container (CPU, memory, network, disk)
- **Startup Metrics**: Time measurements for stack and service startup
- **Recommendation**: Actionable suggestion based on observed metrics, with context and action
- **Metrics History Store**: User-local storage in `~/.20i/` containing optional, bounded historical metrics

## Non-goals *(mandatory)*

- This feature does NOT control or tune resource limits automatically.
- This feature does NOT introduce background agents or daemons in the MVP.
- This feature does NOT store metrics inside project directories.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `20i stats` returns visible metrics quickly on a typical developer machine (target: responsive, non-blocking UX)
- **SC-002**: Recommendations are accurate and actionable 80%+ of the time
- **SC-003**: Developers can identify resource bottlenecks within 30 seconds
- **SC-004**: Startup time improvements are trackable after implementing recommendations
- **SC-005**: Metrics collection overhead is low and observable regressions are tracked across releases

## Assumptions

- Container runtime provides stats API (standard in Docker)
- Resource usage data is accurate and available in real-time
- Historical storage requirements are reasonable (hourly aggregates)
- Recommendations are based on common developer scenarios
