# Feature Specification: Dependency Scanning (CI / Maintainer Guardrails)

**Feature Branch**: `014-dependency-scanning`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟡 High  
**Input**: User description: "Detect vulnerable base images and dependencies with automated security scanning"

## Product Contract *(mandatory)*

Dependency scanning is a **maintainer-facing CI guardrail**, not a runtime feature.

### Scope

- Scanning applies to the **20i stack repository and images**, not to user projects.
- Scanning MUST run in CI (e.g. GitHub Actions) and MUST NOT run inside the local stack or TUI.
- The 20i runtime MUST NOT perform vulnerability scanning on user machines.

### Responsibility boundaries

- The 20i project is responsible for the security posture of its **base images and Dockerfiles**.
- User applications and dependencies remain the responsibility of the application and hosting environment.
- Scan results are informational and gating for maintainers only.

### Determinism and noise control

- Scan results MUST be reproducible for a given image digest and vulnerability database snapshot.
- The project MUST provide a controlled allowlist mechanism for known false positives.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automated Weekly Security Scans (Priority: P1)

As a maintainer, I want automated weekly CI scans of stack images so that vulnerabilities are detected before users are affected.

**Why this priority**: Proactive vulnerability detection is the core security value of this feature.

**Independent Test**: Trigger CI scanning workflow, verify it scans stack Dockerfiles and base images and publishes a vulnerability report artifact.

**Acceptance Scenarios**:

1. **Given** scheduled time (weekly), **When** scan workflow runs, **Then** all stack images are scanned for vulnerabilities
2. **Given** scan completes without vulnerabilities, **When** workflow ends, **Then** success status is reported
3. **Given** vulnerabilities are found, **When** workflow ends, **Then** vulnerability report is generated with severity levels

---

### User Story 2 - Block PRs with Critical Vulnerabilities (Priority: P2)

As a maintainer, I want PRs that introduce critical vulnerabilities to be blocked so that insecure changes don't get merged.

**Why this priority**: PR blocking prevents vulnerabilities from reaching the main branch.

**Independent Test**: Create a PR that updates to a base image with known critical CVE, verify PR check fails with vulnerability details.

**Acceptance Scenarios**:

1. **Given** a PR changes Dockerfile base image, **When** PR scan detects critical vulnerability, **Then** PR status check fails according to the configured severity threshold (default: critical only)
2. **Given** PR scan finds no critical vulnerabilities, **When** scan completes, **Then** PR status check passes
3. **Given** vulnerability is in allowed list (false positive), **When** PR scan runs, **Then** vulnerability is ignored and check passes

#### PR gating policy

- Only vulnerabilities at or above the configured severity threshold MUST block merges.
- Lower-severity findings MUST be reported but MUST NOT block by default.

---

### User Story 3 - Vulnerability Report in PR Comments (Priority: P3)

As a maintainer reviewing PRs, I want vulnerability summaries posted as PR comments so that security impact is visible without leaving GitHub.

**Why this priority**: Inline comments improve review workflow but aren't required for blocking.

**Independent Test**: Create PR with image containing vulnerabilities, verify bot comment summarizes findings with CVE links.

**Acceptance Scenarios**:

1. **Given** PR scan finds vulnerabilities, **When** scan completes, **Then** bot comments with vulnerability summary
2. **Given** vulnerability summary, **When** viewing comment, **Then** severity, CVE IDs, and fix recommendations are shown
3. **Given** multiple PRs in repo, **When** each is scanned, **Then** each gets its own summary comment

---

### User Story 4 - Documented Upgrade Path for CVEs (Priority: P4)

As a maintainer, I want clear upgrade paths for detected vulnerabilities so that I can quickly remediate issues.

**Why this priority**: Upgrade paths accelerate remediation but require scan detection first.

**Independent Test**: Review vulnerability report for known CVE, verify it includes recommended version or patch information.

**Acceptance Scenarios**:

1. **Given** a CVE with known fix, **When** vulnerability is reported, **Then** recommended version or patch is included
2. **Given** multiple vulnerabilities, **When** viewing report, **Then** they are prioritized by severity (critical first)
3. **Given** no direct fix available, **When** viewing report, **Then** mitigation suggestions are provided

---

### Edge Cases

- Vulnerability database unavailable (scan must fail gracefully or retry; PR must not be merged silently)
- Scan exceeds CI time limits (must fail with actionable diagnostics)
- False positives detected (must be suppressible via allowlist with justification)
- Vulnerabilities present only in dev/test layers (must be clearly labeled)
- Multi-arch images produce differing results (must report per-architecture findings)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST scan PHP-FPM Dockerfile and base images for vulnerabilities
- **FR-002**: System MUST run automated scans weekly on main branch
- **FR-003**: System MUST run scans on PRs that modify Dockerfiles or base images
- **FR-004**: System MUST fail PR checks when critical vulnerabilities are detected
- **FR-005**: System MUST generate vulnerability reports with CVE IDs and severity
- **FR-006**: System MUST post vulnerability summaries as PR comments
- **FR-007**: System MUST support vulnerability allowlist for false positives
- **FR-008**: System MUST include fix recommendations in vulnerability reports
- **FR-009**: System MUST use established scanning tools (Trivy, Grype, or similar)
- **FR-010**: Dependency scanning MUST run only in CI workflows and MUST NOT execute as part of the local stack runtime or TUI
- **FR-011**: Scan configuration, allowlists, and thresholds MUST live in the stack repository and MUST NOT affect user projects

### Key Entities

- **Vulnerability Scan**: CI-based security analysis of 20i stack Dockerfiles and built images
- **CVE Report**: Maintainer-facing report listing detected vulnerabilities with severity and remediation guidance
- **Allowlist**: Repository-managed configuration for suppressing known false positives with documented rationale

## Non-goals *(mandatory)*

- This feature does NOT scan user application code or dependencies.
- This feature does NOT run on end-user machines or inside the TUI.
- This feature does NOT guarantee zero vulnerabilities; it provides visibility and guardrails.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Weekly scans complete within 15 minutes
- **SC-002**: 100% of PRs modifying Dockerfiles are scanned before merge
- **SC-003**: No PRs introducing vulnerabilities above the configured severity threshold are merged to main
- **SC-004**: Vulnerability reports provide actionable fix information 90%+ of the time
- **SC-005**: False positive handling is documented and used sparingly, with clear justification

## Assumptions

- Vulnerability databases are regularly updated by scanning tool vendors
- Critical vulnerability classification follows CVSS scoring standards
- GitHub Actions has sufficient permissions to comment on PRs
- Scanning tools work with multi-architecture Docker images
