# Feature Specification: Release Versioning and Workflow

**Feature Branch**: `004-release-workflow`  
**Created**: 2025-12-28  
**Status**: ✅ Complete  
**Priority**: 🟡 High  
**Input**: User description: "Formal release process with semantic versioning, automated CHANGELOG, and GitHub Releases"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Versioned Release (Priority: P1)

As a maintainer, I want to create a versioned release by triggering a workflow so that users can access stable, documented versions of the stack.

**Why this priority**: Versioned releases are essential for users to track changes and roll back if needed.

**Independent Test**: Trigger the release workflow with version "v2.0.0", verify a git tag is created, and a GitHub Release is published with release notes.

**Required:** use of a GitHub Actions workflow to automate the release process. Release-please will manage versioning in `.release-please-manifest.json` and update CHANGELOG.md automatically. Workflow should create a git tag, generate a changelog from commit messages, and publish a GitHub Release. README and documentation should be updated to reflect the new versioning system, use badges and other indicators to show current version, status, and links to release notes.

**Acceptance Scenarios**:

1. **Given** a maintainer triggers release workflow with version "v2.0.0", **When** the workflow completes, **Then** a git tag `v2.0.0` is created
2. **Given** the release workflow runs, **When** it completes successfully, **Then** a GitHub Release is created with the tag
3. **Given** the release workflow runs, **When** it completes, **Then** the release page shows changelog entries since last release

---

### User Story 2 - Automated CHANGELOG Generation (Priority: P2)

As a maintainer, I want CHANGELOG entries to be generated from conventional commits so that release notes are accurate and consistent.

**Why this priority**: Automated changelog reduces manual documentation effort and ensures completeness.

**Independent Test**: Make commits with conventional prefixes (feat:, fix:, docs:), trigger release, and verify CHANGELOG includes all changes categorized correctly.

**Acceptance Scenarios**:

1. **Given** commits with `feat: add new feature`, **When** release is created, **Then** CHANGELOG shows feature under "Features" section
2. **Given** commits with `fix: resolve bug`, **When** release is created, **Then** CHANGELOG shows fix under "Bug Fixes" section
3. **Given** commits with `docs: update readme`, **When** release is created, **Then** CHANGELOG shows entry under "Documentation" section

---

### User Story 3 - Release Validation Before Publishing (Priority: P3)

As a maintainer, I want the release workflow to validate prerequisites before publishing so that incomplete releases are prevented.

**Why this priority**: Validation ensures release quality and prevents user-facing issues.

**Independent Test**: Attempt to create a release without updated CHANGELOG, verify the workflow fails with a clear error message.

**Acceptance Scenarios**:

1. **Given** CHANGELOG.md has not been updated since last release, **When** release workflow runs, **Then** it fails with message "CHANGELOG not updated"
2. **Given** all tests pass and CHANGELOG is updated, **When** release workflow runs, **Then** release is published successfully
3. **Given** tests fail, **When** release workflow runs, **Then** no release is created and failure is logged

---

### User Story 4 - Attach Release Artifacts (Priority: P4)

As a user, I want release artifacts attached to GitHub Releases so that I can download specific components for my setup.

**Why this priority**: Artifacts provide easy access to installable components without cloning the repo.

**Independent Test**: Create a release, verify artifacts (stack archive, install script) are attached to the GitHub Release page.

**Acceptance Scenarios**:

1. **Given** a release is published, **When** viewing the release page, **Then** stack archive (tar.gz) is attached
2. **Given** a release is published, **When** viewing the release page, **Then** installation script is attached
3. **Given** a release is published, **When** downloading artifacts, **Then** checksums are provided for verification

---

### Edge Cases

- What happens if two releases are triggered simultaneously? → Concurrency control prevents parallel releases
- How does the system handle releases from non-default branches? → Workflow only runs on default branch (main), other branches are ignored
- What happens if artifact build fails but release is already tagged? → Tag and release remain, artifacts can be uploaded manually or workflow re-run
- How are pre-release versions (alpha, beta, rc) handled differently? → Pre-release flag set in GitHub Release, version format includes suffix (e.g., v2.0.0-alpha.1)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST create git tags in format `vMAJOR.MINOR.PATCH` (e.g., `v2.0.0`)
- **FR-002**: System MUST create GitHub Releases linked to version tags
- **FR-003**: System MUST generate CHANGELOG entries from conventional commits
- **FR-004**: System MUST validate CHANGELOG is updated before publishing release
- **FR-005**: System MUST run all tests before finalizing release
- **FR-006**: System MUST attach release artifacts (stack archive, install script, checksums) to GitHub Release
- **FR-007**: System MUST support pre-release versions with appropriate tags (alpha, beta, rc)
- **FR-008**: System MUST prevent duplicate version numbers by validating tag doesn't already exist
- **FR-009**: Release workflow MUST be triggerable by maintainers via GitHub Actions UI

### Key Entities

- **Release Version**: Semantic version number following MAJOR.MINOR.PATCH format
- **Git Tag**: Immutable reference to specific commit marking a release
- **GitHub Release**: Published release with notes, artifacts, and download links
- **CHANGELOG**: Document tracking all notable changes between versions

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Releases can be created in under 10 minutes from trigger to published
- **SC-002**: 100% of releases have accurate, auto-generated CHANGELOG entries
- **SC-003**: All release artifacts are available within 15 minutes of release
- **SC-004**: Zero releases published with failing tests
- **SC-005**: Users can identify current version and upgrade path within 30 seconds of visiting releases page

## Assumptions

- Maintainers follow conventional commit format for meaningful changelog generation
- GitHub Actions has sufficient permissions to create tags and releases
- Artifact build process is reliable and reproducible
- Semantic versioning rules are understood by maintainers
