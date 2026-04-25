# Feature Specification: Automated Version Update Pipeline

**Feature Branch**: `003-auto-version-updates`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟡 High  
**Input**: User description: "GitHub Actions workflow to keep PHP, MariaDB, Nginx versions current"

## Definition of “Latest Stable” *(mandatory)*

Docker image tags vary widely across registries and repositories. This spec defines **stable** as a tag that:

- Represents a GA release (no `alpha`, `beta`, `rc`, `preview`, `dev`, `nightly` markers)
- Matches a predictable version pattern for the component
- Is compatible with how versions are stored in `config/stack-vars.yml`

### Tag selection and normalization rules

- **PHP**:
  - Consider tags that represent the PHP version family used by this stack (e.g. `8.5`, `8.5.x`, `8.6`, `8.6.x`), excluding any tag containing `alpha`, `beta`, `rc`, `preview`, `dev`, `nightly`.
  - Normalise discovered tags to the format stored in `stack-vars.yml` (e.g. if the stack stores `8.5`, then map any `8.5.x` to `8.5` unless the stack explicitly stores patch versions).

- **MariaDB**:
  - Consider stable tags matching the major.minor line used by the stack (e.g. `10.6`, `10.11`), excluding pre-release markers.
  - Normalise to the format stored in `stack-vars.yml`.

- **Nginx**:
  - Prefer the stable release line over mainline if both exist.
  - Exclude pre-release markers.
  - Normalise to the format stored in `stack-vars.yml`.

### Stability priority

If multiple valid stable tags exist, the workflow MUST select the **highest stable version** after normalisation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automatic Weekly Version Checks (Priority: P1)

As a maintainer, I want the system to automatically check for new stable versions of PHP, MariaDB, and Nginx weekly so that the stack stays current without manual intervention.

**Why this priority**: Keeping dependencies current is critical for security and compatibility; automation removes maintenance burden.

**Independent Test**: Trigger the workflow manually, verify it queries Docker Hub for latest versions, and compare against current `stack-vars.yml` values.

**Acceptance Scenarios**:

1. **Given** the scheduled time (Sunday 00:00 UTC), **When** the workflow runs, **Then** it queries the registry for candidate tags and selects the latest stable versions per this spec’s stable-tag rules
2. **Given** all versions are current, **When** the workflow completes, **Then** no PR is created and workflow exits successfully
3. **Given** a newer version is available, **When** the workflow detects it, **Then** it proceeds to create an update PR

---

### User Story 2 - Automated PR Creation for Updates (Priority: P2)

As a maintainer, I want the system to create a pull request with version updates so that I can review and approve changes before they're merged.

**Why this priority**: PRs provide a review checkpoint and audit trail for version changes.

**Independent Test**: Mock a newer PHP version available, run the workflow, and verify a PR is created with updated `stack-vars.yml` and CHANGELOG entry.

**Acceptance Scenarios**:

1. **Given** PHP 8.6 is available and current is 8.5, **When** the workflow runs, **Then** a PR is created updating `PHP_VERSION` to 8.6
2. **Given** the PR is created, **When** viewing the PR, **Then** it includes a CHANGELOG entry explaining the update rationale
3. **Given** multiple components have updates, **When** the workflow runs, **Then** a single PR is created with all updates bundled

#### PR safety rules

- The workflow MUST create a PR only if it produces a real diff in `config/stack-vars.yml`.
- The workflow MUST NOT open multiple concurrent update PRs for the same branch of work; if an update PR already exists, it MUST update that PR instead of creating a new one.
- The workflow MUST use a predictable branch name (e.g. `auto/version-bump-YYYY-MM-DD`).
- The workflow MUST include a concise PR description that lists:
  - Components updated
  - Old version → new version
  - Test summary
- The workflow MUST include a note in the PR description about pre-built image availability expectations:
  - If the update changes a version that affects pre-built images (e.g. `PHP_VERSION`), the PR MUST state whether pre-built images are expected to be available only after the next stack release, and that local-build fallback may be required on the branch.
- The workflow SHOULD, where feasible, link to the relevant image/tag strategy (release pin and PHP-line tags) used by the stack’s pre-built images.

---

### User Story 3 - Automated Test Suite Before PR (Priority: P3)

As a maintainer, I want the system to run automated tests before creating an update PR so that only working updates are proposed.

**Why this priority**: Prevents broken updates from being proposed, saving reviewer time and maintaining quality.

**Independent Test**: Introduce a breaking change in a test update, run the workflow, and verify no PR is created due to test failure.

**Acceptance Scenarios**:

1. **Given** a version update is detected, **When** the workflow runs tests, **Then** it boots the stack with the updated versions, verifies key services respond, and tears down cleanly
2. **Given** tests pass, **When** the workflow completes, **Then** a PR is created
3. **Given** tests fail, **When** the workflow completes, **Then** no PR is created and failure is logged

#### Minimum automated test contract

The workflow’s pre-PR tests MUST, at minimum:

- Bring the stack up using the proposed updated versions
- Tests MUST validate the local-build path (`USE_PREBUILT=false`) to ensure the stack remains buildable even when pre-built images are not yet available.
- Verify:
  - Nginx is responding on the configured host port
  - phpMyAdmin is reachable on the configured port
  - MariaDB is reachable and accepts a connection (credentialed)
- Bring the stack down cleanly

Tests MUST be designed to run reliably in CI (avoid port conflicts and ensure teardown always runs).

#### Pre-built images cadence note

- Pre-built images are guaranteed for official stack releases (see spec 006).
- Update PRs may reference versions (especially `PHP_VERSION`) that do not yet have pre-built images published; this is expected on branches, and local-build tests ensure the update is still safe to propose.

---

### User Story 4 - Manual Workflow Trigger (Priority: P4)

As a maintainer, I want to manually trigger the version check workflow so that I can check for updates on-demand.

**Why this priority**: Manual triggers provide flexibility for urgent security updates outside the weekly schedule.

**Independent Test**: Use GitHub Actions UI to manually trigger the workflow, verify it runs and checks for updates.

**Acceptance Scenarios**:

1. **Given** the workflow definition, **When** a maintainer triggers it manually, **Then** the workflow runs immediately
2. **Given** a manual trigger with a component selector, **When** the workflow runs, **Then** it checks only the selected component(s)

---

### Edge Cases

- What happens when Docker Hub API is unavailable or rate-limited?
- How does the system handle pre-release or release candidate versions (should skip)?
- What happens if the automated PR conflicts with existing changes?
- How does the system handle rollback if a merged update causes issues?
- Update PR bumps `PHP_VERSION` but no matching pre-built image exists yet (must be documented in PR; local-build path remains the source of truth until release)

## Non-goals *(mandatory)*

- This feature does NOT change Dockerfile build steps, installed extensions, or configuration defaults beyond version bumps.
- This feature does NOT migrate major versions automatically if doing so requires manual intervention or stack reconfiguration.
- This feature does NOT change runtime behaviour of the stack except where upstream version changes inherently alter behaviour.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST run automated version checks weekly on Sunday 00:00 UTC
- **FR-002**: System MUST query the relevant container registry for PHP, MariaDB, and Nginx candidate tags and select latest stable versions
- **FR-003**: System MUST compare discovered versions against current `config/stack-vars.yml` values
- **FR-004**: System MUST create a PR only when newer stable versions are available
- **FR-005**: System MUST run automated stack start/stop tests before creating PR
- **FR-006**: System MUST include CHANGELOG entry in update PR
- **FR-007**: System MUST ensure update PRs clearly communicate any expected lag between version bumps and availability of pre-built images, and indicate the fallback expectation on branches
- **FR-008**: System MUST support manual workflow trigger via GitHub Actions UI
- **FR-009**: System MUST skip pre-release, alpha, beta, and RC versions
- **FR-010**: System MUST handle API failures gracefully with appropriate logging

### Key Entities

- **Version Check Workflow**: Scheduled GitHub Actions workflow that runs weekly to check for updates
- **Version Manifest**: Current versions stored in `config/stack-vars.yml`
- **Update PR**: Pull request containing version updates, CHANGELOG entry, and test results

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Version checks run automatically every week without manual intervention
- **SC-002**: PRs are created within 15 minutes of workflow detecting updates
- **SC-003**: 100% of update PRs pass automated tests before creation
- **SC-004**: Zero broken updates merged (all PRs are tested before creation)
- **SC-005**: Maintainers spend less than 5 minutes reviewing and merging update PRs
- **SC-006**: Stack stays within 2 weeks of latest stable versions

## Assumptions

- Container registries provide tags that can be filtered deterministically into stable vs pre-release
- The workflow may require authenticated registry access (e.g. a token) to avoid rate limits
- GitHub Actions has sufficient runtime minutes for weekly execution
- Test suite accurately validates stack functionality
- Maintainers review PRs within reasonable timeframe (1 week)
