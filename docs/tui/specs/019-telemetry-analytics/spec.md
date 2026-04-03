# Feature Specification: Telemetry & Analytics (Exploratory)

**Feature Branch**: `019-telemetry-analytics`  
**Created**: 2025-12-28  
**Status**: Exploratory / Parking Spec  
**Priority**: ⚪ Very Low (Optional, Future Consideration)  
**Input**: User description: "Optional, ethical telemetry to understand high-level usage patterns"

## Product Contract *(mandatory)*

Telemetry, if implemented, MUST be:

- **Fully opt-in** (disabled by default)
- **Anonymous by design** (no PII, no user identifiers, no project identifiers)
- **Non-essential** to core functionality
- **Transparent** (users can see exactly what would be collected)

The project MUST remain fully usable without telemetry.

## Scope *(exploratory)*

Telemetry is intended only to provide **high-level, aggregate signals** that help maintainers understand:

- Which commands are used most often
- Which optional services or modules are commonly enabled
- Which platform combinations are most common (OS, architecture)

Telemetry is **not** intended for behavioural tracking, profiling, or user identification.

## Non-goals *(mandatory)*

- Telemetry MUST NOT influence feature access or behaviour.
- Telemetry MUST NOT be required for support or bug fixes.
- Telemetry MUST NOT include file paths, project names, hostnames, IP addresses, or timestamps precise enough to identify individuals.
- Telemetry MUST NOT become a KPI or success metric for the project.
- Telemetry MUST NOT pressure users to opt in.

## Consent & Control *(mandatory)*

- Telemetry is **disabled by default**.
- Users MAY enable telemetry explicitly via a command (e.g. `20i telemetry enable`).
- Users MUST be able to disable telemetry at any time.
- The `DO_NOT_TRACK=1` environment variable MUST override all other settings.

Consent state MUST be stored in the user configuration directory (e.g. `~/.20i/`), never in project directories.

## Transparency *(mandatory)*

If telemetry exists, users MUST be able to:

- View exactly what data categories would be collected
- See example payloads (sanitized)
- Read a plain-language privacy policy

Commands such as `20i telemetry info` MAY be provided for this purpose.

## Data Categories *(illustrative only)*

If implemented, telemetry MAY include:

- CLI version
- OS type and architecture
- Invoked command name (e.g. `up`, `down`, `status`)
- Optional services or modules enabled (by identifier only)
- Anonymized error categories (not raw stack traces)

Exact payloads are intentionally unspecified and subject to review.

## Error Reporting *(optional)*

Anonymized error reporting MAY be considered separately from usage telemetry.

If implemented:

- Error reports MUST strip paths, usernames, and project identifiers
- Error reports MUST be aggregated by category
- Users MUST be able to opt in/out independently

## Governance & Review *(mandatory)*

- Telemetry code MUST be open source and auditable.
- Any change to telemetry behaviour MUST be documented clearly in release notes.
- Telemetry collection SHOULD be reviewed periodically for necessity and scope creep.

## Success Criteria *(exploratory)*

- **SC-001**: Users trust that telemetry is optional and respectful.
- **SC-002**: Telemetry (if enabled) does not impact correctness or stability.
- **SC-003**: Telemetry can be removed entirely without affecting core features.

## Assumptions

- Some users may choose to opt in to help guide development.
- Maintainers prefer qualitative feedback over aggressive metrics.
- The project prioritizes trust and simplicity over analytics depth.

---
