# Feature Specification: Installer, Onboarding, And Environment Readiness

**Feature Branch**: `005-installer-and-onboarding`  
**Created**: 2026-04-26  
**Status**: Draft  
**Input**: User description: "Research installer patterns, behaviours, and onboarding setup for similar software. Expand recommendations into a more detailed specification."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install StageServe Through A Recommended Path (Priority: P1)

A developer on a clean machine wants one clear install path that is easy to update and uninstall.

**Why this priority**: If install is ambiguous or fragile, all downstream onboarding work suffers.

**Independent Test**: On a clean supported macOS machine, follow only the recommended install path and verify the `stage` command is available and versioned.

**Acceptance Scenarios**:

1. **Given** a supported macOS machine, **When** the operator follows the documented primary install path, **Then** StageServe installs without requiring source builds.
2. **Given** the operator runs `stage --version`, **When** install is complete, **Then** the version output is shown and matches the installed release.
3. **Given** the recommended install path is unavailable, **When** the operator uses the fallback signed binary path, **Then** integrity verification steps are documented and executable.

---

### User Story 2 - Complete First-Run Machine Setup Reliably (Priority: P1)

A developer wants a single command that validates prerequisites and performs one-time setup such as DNS readiness.

**Why this priority**: Current onboarding requires several manual steps and troubleshooting heuristics.

**Independent Test**: Run `stage setup` on a machine with at least one missing prerequisite and confirm the output reports each step as `ready`, `needs_action`, or `error` with concrete remediation.

**Acceptance Scenarios**:

1. **Given** Docker is installed but daemon is stopped, **When** the operator runs `stage setup`, **Then** setup reports Docker as `needs_action` and provides exact restart/recheck instructions.
2. **Given** local DNS for the configured suffix is not bootstrapped, **When** the operator runs `stage setup`, **Then** setup invokes or guides `dns-setup` and reports final DNS readiness explicitly.
3. **Given** setup is rerun after successful completion, **When** the operator runs `stage setup` again, **Then** complete steps are reported as `ready` and are not redundantly re-applied.

---

### User Story 3 - Diagnose Drift And Recover Quickly (Priority: P2)

A developer with a previously working machine wants a fast diagnosis command that identifies drift and recommends targeted fixes.

**Why this priority**: Day-2 reliability matters as much as first-run success.

**Independent Test**: Break one readiness dependency (for example resolver file missing), run `stage doctor`, and confirm detection plus fix guidance without ambiguous error text.

**Acceptance Scenarios**:

1. **Given** resolver state drifts out of compliance, **When** `stage doctor` runs, **Then** the output identifies DNS drift and provides a specific recovery command.
2. **Given** shared gateway resources are unhealthy, **When** `stage doctor` runs, **Then** output distinguishes gateway issues from project runtime issues.
3. **Given** the operator uses automation, **When** `stage doctor --json` runs, **Then** the command emits stable machine-readable statuses and exit semantics.


---

### User Story 4 - Initialize A Project For StageServe (Priority: P1)

A developer in a project root wants an optional guided initializer that creates project-local StageServe config and links the app cleanly to the stack contract.

**Why this priority**: First machine setup and first project setup are separate concerns; project onboarding needs a safe, repeatable entrypoint.

**Independent Test**: In a repository without `.env.stageserve`, run `stage init` from project root and confirm it writes a starter config, validates docroot/site settings, and provides next commands.

**Acceptance Scenarios**:

1. **Given** a project root without `.env.stageserve`, **When** the operator runs `stage init`, **Then** StageServe creates a starter project config with documented defaults and ownership boundaries.
2. **Given** a project with custom app layout, **When** the operator runs `stage init --docroot <path> --site-name <name>`, **Then** the written config preserves those values and validates them against project paths.
3. **Given** a project already initialized, **When** the operator reruns `stage init`, **Then** StageServe does not overwrite existing settings unless an explicit overwrite flag is provided.

## Operational Impact *(mandatory)*

### Ease Of Use & Workflow Impact

- Affected commands: new `stage setup`, new `stage doctor`, new optional `stage init` (project-root), existing `stage dns-setup`, and first-run documentation.
- Backward compatibility: existing lifecycle commands (`up`, `attach`, `status`, `down`) remain unchanged; this feature adds onboarding/readiness surfaces.
- Friction removed: manual prerequisite discovery, unclear remediation steps, and inconsistent first-run command order.

### Configuration & Precedence

- New configuration inputs may include optional setup/doctor flags (for example `--json`, `--recheck`, `--non-interactive`).
- Existing config precedence for runtime behavior remains unchanged.
- DNS setup continues to respect existing local DNS config keys and platform support constraints.
- `stage init` writes project-local `.env.stageserve` and must preserve existing config precedence/ownership rules.

### State, Isolation & Recovery

- Setup/doctor must be machine-scoped and must not mutate per-project runtime state unexpectedly.
- Setup actions requiring elevated privileges must be explicit and bounded to DNS/bootstrap surfaces.
- Doctor checks must remain read-mostly and avoid destructive operations by default.

### Documentation Surfaces

- `README.md` first-run section.
- `docs/runtime-contract.md` for command ownership and setup/doctor contract.
- New install/onboarding reference docs for support matrix and troubleshooting.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: StageServe MUST define one documented recommended install path per supported OS, with a signed-binary fallback path.
- **FR-002**: StageServe MUST provide a `stage setup` command that runs prerequisite and one-time environment readiness checks.
- **FR-003**: `stage setup` MUST report each check using normalized statuses: `ready`, `needs_action`, or `error`.
- **FR-004**: `stage setup` MUST print concrete remediation instructions for all `needs_action` and `error` outcomes.
- **FR-005**: `stage setup` MUST be idempotent; rerunning setup after successful completion MUST not reapply completed actions unnecessarily.
- **FR-006**: Setup MUST include Docker binary presence and daemon reachability checks.
- **FR-007**: Setup MUST include local DNS readiness checks for supported platforms and integrate with existing `dns-setup` behavior.
- **FR-008**: Setup MUST NOT silently escalate privileges; any privileged operation MUST be explicit and operator-confirmed.
- **FR-009**: StageServe MUST provide a `stage doctor` command for post-install diagnostics and drift detection.
- **FR-010**: `stage doctor` MUST include checks for binary/version visibility, Docker readiness, DNS readiness, shared gateway health, and writable state directories.
- **FR-011**: Both setup and doctor MUST support machine-readable output via `--json` with stable step identifiers.
- **FR-012**: Setup and doctor MUST provide non-zero exit behavior when blocking readiness issues remain unresolved.
- **FR-013**: Documentation MUST publish a compatibility matrix covering supported OS versions, architectures, and supported Docker providers.
- **FR-014**: Release documentation MUST include artifact integrity verification instructions (checksums/signatures) for direct binary installs.
- **FR-015**: Installer and onboarding documentation MUST include concise post-install “next step” guidance (for example: `stage setup`, then `stage up`).
- **FR-016**: StageServe MUST provide an optional `stage init` command intended to run from a project root directory.
- **FR-017**: `stage init` MUST generate a starter project `.env.stageserve` when absent and MUST NOT overwrite an existing file unless explicitly requested.
- **FR-018**: `stage init` MUST support explicit project bootstrap inputs (at minimum site name and docroot) and validate them against the current repository.
- **FR-019**: `stage init` MUST output clear next-step guidance for connecting the initialized app to StageServe runtime flow (`setup` as needed, then `up`).
- **FR-020**: `stage init` SHOULD support machine-readable output via `--json` using stable result identifiers consistent with setup/doctor.

### Non-Functional Requirements

- **NFR-001**: Setup first-pass runtime on a healthy machine SHOULD complete within 10 seconds excluding any human-entered privileged command.
- **NFR-002**: Doctor checks SHOULD complete within 5 seconds on a healthy machine.
- **NFR-003**: Error messages MUST be actionable and include exact command examples for recovery.
- **NFR-004**: JSON output schema SHOULD remain backward compatible within a minor release line.

### Out Of Scope

- GUI/TUI onboarding experiences.
- Automatic installation of Docker Desktop or third-party providers by StageServe itself.
- Full certificate lifecycle automation for future `.dev` TLS workflows (defer to follow-up spec).
- Framework-specific deep app rewrites or auto-detection beyond basic project config scaffolding during `stage init`.
- CI/CD pipeline redesign beyond release artifact integrity documentation.

## Key Entities *(include if feature involves data)*

- **Setup Step Result**: Structured status object for one setup check/action (`id`, `label`, `status`, `message`, `remediation`, optional metadata).
- **Doctor Check Result**: Structured diagnostics object representing one readiness or drift check with severity and fix hints.
- **Support Matrix Record**: Versioned documentation entity describing supported platforms, architectures, Docker providers, and constraints.
- **Project Init Result**: Structured outcome for project initialization (created/updated/skipped), validated settings, and recommended next actions.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On clean supported machines, at least 90% of first-run users can reach a `setup complete` state without consulting external troubleshooting docs.
- **SC-002**: `stage setup` and `stage doctor` produce deterministic step statuses and actionable remediation for all blocking checks in validation scenarios.
- **SC-003**: Re-running `stage setup` after completion performs no destructive reconfiguration and reports previously-completed steps as `ready`.
- **SC-004**: Documentation consistently presents the onboarding sequence as `install -> setup -> init -> up` and includes integrity verification for direct binary installs.
- **SC-005**: In repos without StageServe project config, operators can run `stage init` and reach a valid `.env.stageserve` configuration with no manual file authoring.

## Assumptions

- StageServe remains a CLI-first product.
- Docker provider runtime and DNS bootstrapping continue as core prerequisites for local hostname workflows.
- Existing project runtime lifecycle semantics remain unchanged by this installer/onboarding feature, except for adding optional project initialization scaffolding via `stage init`.
- Initial implementation focus is macOS, with explicit unsupported/experimental guidance for additional platforms until validated.
