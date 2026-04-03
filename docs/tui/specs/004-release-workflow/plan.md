# Implementation Plan: Release Versioning and Workflow

**Branch**: `004-release-workflow` | **Date**: 2025-12-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-release-workflow/spec.md`

## Summary

Implement a fully automated release workflow using GitHub Actions that creates semantic versioned releases with auto-generated changelogs, release artifacts, and comprehensive validation gates. The system will enforce conventional commits, prevent invalid releases, and provide maintainers with a streamlined release process via GitHub Actions UI.

## Technical Context

**Language/Version**: YAML (GitHub Actions), Bash 5.x  
**Primary Dependencies**: GitHub Actions, conventional-changelog-cli, gh CLI, git  
**Storage**: N/A (Git tags, GitHub Releases as storage)  
**Testing**: GitHub Actions workflow validation, ShellCheck for scripts  
**Target Platform**: GitHub-hosted runners (ubuntu-latest)  
**Project Type**: Single project - DevOps/Infrastructure automation  
**Performance Goals**: Release completion in <10 minutes (per SC-001)  
**Constraints**: GitHub Actions rate limits, artifact size limits (2GB per file), 6-hour workflow timeout  
**Scale/Scope**: Single repository, maintainer-triggered releases, ~4 artifacts per release

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Compliance Notes |
|-----------|--------|------------------|
| I. Environment-Driven Configuration | ✅ PASS | Version in `VERSION` file, workflow inputs for overrides, no hard-coded values |
| II. Multi-Platform First | ✅ PASS | Artifacts support AMD64/ARM64, workflow runs on ubuntu-latest (standard) |
| III. Path Independence | ✅ PASS | All scripts use `$GITHUB_WORKSPACE`, relative paths within repo |
| IV. Centralized Defaults | ✅ PASS | Release-please manifest as single source, workflow inputs for release-time overrides |
| V. User Experience & Feedback | ✅ PASS | Preflight summary, emoji status indicators, clear error messages |
| VI. Documentation as First-Class | ✅ PASS | Auto-generated CHANGELOG, README badges, release notes |
| VII. Version Consistency | ✅ PASS | Release-please manifest synced to git tags, CHANGELOG, README badge |
| Commit Hygiene (Dev Workflow) | ✅ PASS | Conventional commits enforced via PR validation workflow |

**Constitution Gate**: ✅ PASSED - All principles satisfied

## Project Structure

### Documentation (this feature)

```text
specs/004-release-workflow/
├── plan.md              # This file
├── research.md          # Phase 0: Tooling decisions and best practices
├── data-model.md        # Phase 1: VERSION format, config schema
├── quickstart.md        # Phase 1: Maintainer release guide
├── contracts/           # Phase 1: Workflow API specifications
│   ├── release-workflow.md     # Main release workflow interface
│   └── events.md               # Workflow events specification
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
.github/
├── workflows/
│   ├── release.yml              # Main release workflow (manual trigger)
│   ├── validate-pr.yml          # PR validation (commit format, tests)
│   ├── changelog-preview.yml    # Generate changelog preview on PR
│   └── auto-release.yml         # Optional: auto-release on main push
├── PULL_REQUEST_TEMPLATE.md     # PR template with conventional commit guide
├── ISSUE_TEMPLATE/
│   └── release-request.yml      # Release request issue template
└── release.yml                  # Release drafter config (optional)

config/
├── stack-vars.yml               # Existing - unchanged
└── release/
    ├── config.yml               # Release categories, labels, templates
    └── changelog-template.hbs   # Handlebars template for CHANGELOG

scripts/
├── setup-local.sh               # Existing - unchanged
└── release/
    ├── validate.sh              # Pre-release validation checks
    ├── changelog.sh             # Generate/update CHANGELOG
    ├── version.sh               # Read version from release-please manifest
    ├── artifacts.sh             # Package release artifacts
    └── publish.sh               # Publish to GitHub Releases

CHANGELOG.md                     # Existing - auto-updated by workflow
README.md                        # Existing - add version badge
```

**Structure Decision**: Single project with `.github/workflows/` for CI/CD automation and `scripts/release/` for reusable release utilities. This follows the existing `scripts/` convention and keeps release logic testable outside GitHub Actions.

## Complexity Tracking

> No constitution violations - section not required.

---

## Post-Design Constitution Re-Check

*Re-evaluated after Phase 1 design completion.*

| Principle | Pre-Design | Post-Design | Notes |
|-----------|------------|-------------|-------|
| I. Environment-Driven Configuration | ✅ PASS | ✅ PASS | Confirmed: VERSION file, workflow inputs, release-please config |
| II. Multi-Platform First | ✅ PASS | ✅ PASS | Confirmed: ubuntu-latest runners, platform-agnostic artifacts |
| III. Path Independence | ✅ PASS | ✅ PASS | Confirmed: All scripts use relative paths, GITHUB_WORKSPACE |
| IV. Centralized Defaults | ✅ PASS | ✅ PASS | Confirmed: release-please-config.json centralizes settings |
| V. User Experience & Feedback | ✅ PASS | ✅ PASS | Confirmed: PR preview comments, release notes, status badges |
| VI. Documentation as First-Class | ✅ PASS | ✅ PASS | Confirmed: Auto CHANGELOG, quickstart.md, inline comments |
| VII. Version Consistency | ✅ PASS | ✅ PASS | Confirmed: Manifest → tag → CHANGELOG → README badge |
| Commit Hygiene | ✅ PASS | ✅ PASS | Confirmed: validate-pr.yml enforces conventional commits |

**Post-Design Gate**: ✅ PASSED - Design maintains full compliance

### Design Decisions Aligned with Constitution

1. **release-please over semantic-release**: Simpler, no Node.js dependency (Principle I, III)
2. **Manifest-based versioning**: release-please manifest as single source of truth (Principle IV)
3. **Separate scripts/release/**: Testable outside CI, reusable (Principle VI)
4. **SHA256 checksums**: Security verification for artifacts (best practice)
5. **PR-based releases**: Maintainer control, clear audit trail (Principle V)

---

## Phase 2 Planning (via /speckit.tasks)

The following areas are ready for task breakdown:

1. **Workflow Implementation**: Create GitHub Actions YAML files
2. **Script Development**: Build release utility scripts
3. **Configuration Files**: Set up release-please config
4. **Documentation Updates**: README badges, CONTRIBUTING guide
5. **Testing**: Workflow validation, script testing
