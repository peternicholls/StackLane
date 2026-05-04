---
spec phase: 007 Version Management, Releases, and Update Workflows
description: "Placeholder scope for versioning, release management, and update delivery"
status: placeholder
---

# 007: Version Management, Releases, and Update Workflows

## Purpose
Define the versioning and release operating model for StageServe so builds, releases, downloads, and updates are predictable, testable, and easy to operate.

## Seed Scope
- Semantic version management flow and tracking.
- Release workflow, including GitHub release and tagging, for downloadable artifacts.
- Update workflows, release management, and self-update behavior for the CLI.
- Effective use of CI workflows to support release confidence and repeatability.

## Expected Outputs
- A clear versioning policy.
- A release workflow that covers tagging, artifact publication, and release notes.
- An update strategy for installed CLI binaries.
- Validation steps for CI and release automation.

## Open Questions
- What is the canonical source of truth for the current version?
- Will the CLI support self-update on all supported platforms or only selected targets?
- How should pre-releases, rollback releases, and hotfixes be represented?
- Which release steps must be automated versus operator-driven?

## Out Of Scope For This Placeholder
- Unrelated CLI behavior changes.
- Large installer redesigns unless required for update delivery.
- Distribution channels not yet used by the project.

## Starting Notes
This is intentionally a placeholder rather than a full spec package. Expand it into a working spec only when phase 007 becomes active.
