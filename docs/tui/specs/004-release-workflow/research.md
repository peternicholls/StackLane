# Research: Release Workflow Tooling and Best Practices

**Feature**: 004-release-workflow  
**Date**: 2025-12-28  
**Status**: Complete

## Overview

This document captures research findings for implementing automated release versioning and workflow for 20i Stack. All NEEDS CLARIFICATION items from the Technical Context have been resolved.

---

## 1. Changelog Generation Tooling

### Decision: **release-please** (by Google)

### Rationale
- **Native GitHub Actions integration**: Designed as a GitHub Action first, eliminating complex setup
- **"Simple" release type support**: Perfect for non-npm projects like 20i Stack - uses `VERSION` file without needing package.json
- **PR-based workflow**: Creates "Release PRs" that accumulate changes, giving maintainers control over when to release
- **Automatic version calculation**: Parses conventional commits to determine semver bumps automatically
- **Actively maintained**: 6.2k+ stars, 345+ releases, maintained by Google

### Alternatives Considered

| Tool | Status | Reason for Rejection |
|------|--------|---------------------|
| semantic-release | ❌ Rejected | Requires Node.js ecosystem, complex plugin configuration, designed for npm packages |
| conventional-changelog-cli | ❌ Rejected | Manual CLI tool, doesn't handle GitHub Releases or tagging, requires additional automation |
| github-changelog-generator | ❌ Rejected | Ruby dependency, less integrated with GitHub Actions, primarily retrospective |
| commit-and-tag-version | ❌ Rejected | Good for npm projects, but lacks GitHub Release creation and PR workflow |

### Implementation Notes
```yaml
# release-please-config.json
{
  "release-type": "simple",
  "bump-minor-pre-major": true,
  "bump-patch-for-minor-pre-major": true,
  "include-component-in-tag": false,
  "changelog-sections": [
    {"type": "feat", "section": "Features"},
    {"type": "fix", "section": "Bug Fixes"},
    {"type": "docs", "section": "Documentation"},
    {"type": "perf", "section": "Performance"},
    {"type": "refactor", "section": "Code Refactoring"}
  ]
}
```

---

## 2. GitHub Actions Release Patterns

### Decision: **Hybrid Approach** - release-please + workflow_dispatch

### Rationale
- release-please handles automatic PR-based releases from conventional commits
- workflow_dispatch provides manual trigger capability for exceptional cases
- Supports both automated and controlled release scenarios

### Required Permissions
```yaml
permissions:
  contents: write       # Create tags, releases, push changes
  pull-requests: write  # Create/update release PRs
```

### Concurrency Control
```yaml
concurrency:
  group: release-${{ github.ref }}
  cancel-in-progress: false  # Never cancel in-progress releases
```

### Alternatives Considered

| Pattern | Status | Reason |
|---------|--------|--------|
| Manual-only workflow_dispatch | ❌ Rejected | Requires maintainer to remember version numbers, error-prone |
| Tag-push triggered | ⚠️ Partial | Used as secondary trigger after release-please creates tag |
| Cron-based releases | ❌ Rejected | Inappropriate for infrastructure project, unpredictable timing |

### Implementation Notes
- Primary: `release-please.yml` runs on push to main
- Secondary: Tag-triggered workflow for artifact building
- Use `softprops/action-gh-release@v2` for artifact uploads
- Artifact limits: 5GB per file, 500 files per release (sufficient)

---

## 3. Pre-release Version Handling

### Decision: **SemVer 2.0 Pre-release Identifiers**

### Rationale
- Standard format recognized by all package managers and tools
- release-please supports `prerelease-type` configuration
- Clear ordering for users and tooling

### Standard Format
```
vMAJOR.MINOR.PATCH-PRERELEASE.N

Examples:
v2.0.0-alpha.1   # Early testing, unstable
v2.0.0-beta.1    # Feature complete, testing stability  
v2.0.0-rc.1      # Release candidate, final testing
v2.0.0           # Stable release
```

### Version Ordering (per SemVer spec)
```
v1.0.0-alpha.1 < v1.0.0-alpha.2 < v1.0.0-beta.1 < v1.0.0-rc.1 < v1.0.0
```

### Alternatives Considered

| Format | Status | Reason |
|--------|--------|--------|
| Date-based (v2024.12.28) | ❌ Rejected | Doesn't convey compatibility info, awkward for pre-releases |
| Build metadata only | ❌ Rejected | Build metadata ignored in precedence, not suitable for pre-releases |
| Custom suffixes (dev, nightly) | ❌ Rejected | Non-standard, tooling compatibility issues |

### Implementation Notes
```yaml
# For pre-release in release-please
release-type: simple
prerelease: true
prerelease-type: alpha  # or beta, rc
```

---

## 4. Artifact Packaging Strategy

### Decision: **softprops/action-gh-release@v2** with SHA256 checksums

### Rationale
- Modern, actively maintained action (436+ commits)
- Simple glob-based file upload
- Supports external release notes
- Handles existing releases gracefully

### Artifact Structure
```
20i-stack-v2.0.0/
├── 20i-stack-v2.0.0.tar.gz   # Main distribution archive
├── install.sh                 # Quick install script (standalone)
├── checksums.sha256           # SHA256 verification file
└── RELEASE_NOTES.md           # Version-specific notes
```

### Checksum Generation
```bash
cd dist/
sha256sum *.tar.gz *.sh > checksums.sha256
```

### Alternatives Considered

| Action | Status | Reason |
|--------|--------|--------|
| actions/upload-release-asset | ❌ Rejected | Deprecated, limited functionality |
| gh release upload (CLI) | ⚠️ Fallback | Good fallback, more verbose in workflow |
| ncipollo/release-action | ⚠️ Alternative | Viable but less feature-rich than softprops |

### Implementation Notes
```yaml
- name: Upload release artifacts
  uses: softprops/action-gh-release@v2
  with:
    files: |
      dist/20i-stack-*.tar.gz
      dist/install.sh
      dist/checksums.sha256
```

---

## 5. Version File Management

### Decision: **VERSION file** as primary source + git tags as authoritative record

### Rationale
- `VERSION` file is simple, language-agnostic, human-readable
- Git tags are immutable and integrated with GitHub Releases
- Avoids package.json overhead for non-npm project
- Easy to reference in shell scripts: `VERSION=$(cat VERSION)`

### Source of Truth Hierarchy
1. **Git tags** - Authoritative, immutable record of releases
2. **VERSION file** - Current development version, updated by release-please
3. **CHANGELOG.md** - Human-readable history with dates

### Alternatives Considered

| Approach | Status | Reason |
|----------|--------|--------|
| package.json only | ❌ Rejected | Requires Node.js ecosystem, adds unnecessary dependency |
| Git tags only | ❌ Rejected | Harder to reference in scripts/docs without parsing git |
| config/stack-vars.yml | ❌ Rejected | Mixes configuration with versioning, complicates tooling |

### Implementation Notes
- Create `/VERSION` file with initial content: `1.0.0`
- release-please config: `"version-file": "VERSION"`
- Reference in scripts: `VERSION=$(cat VERSION)`
- README badge: `![Version](https://img.shields.io/github/v/release/peternicholls/20i-stack)`

---

## 6. Changelog Format

### Decision: **Keep a Changelog** format with conventional commits automation

### Rationale
- Keep a Changelog is the industry standard (keepachangelog.com)
- release-please generates compliant format automatically
- Human-readable sections match conventional commit types
- Existing CHANGELOG.md already follows this format

### Conventional Commit Mapping

| Commit Type | Changelog Section | Version Bump |
|-------------|------------------|--------------|
| `feat:` | Added | Minor |
| `fix:` | Fixed | Patch |
| `docs:` | Documentation | Patch |
| `refactor:` | Changed | Patch |
| `perf:` | Performance | Patch |
| `BREAKING CHANGE:` | ⚠ Breaking Changes | Major |
| `deprecate:` | Deprecated | Minor |
| `security:` | Security | Patch |

### Alternatives Considered

| Format | Status | Reason |
|--------|--------|--------|
| GitHub auto-generated notes | ❌ Rejected | Less structured, not Keep a Changelog compliant |
| Raw commit list | ❌ Rejected | Not user-friendly, lacks categorization |
| Custom format | ❌ Rejected | Increases maintenance burden, non-standard |

### Implementation Notes
- release-please auto-generates sections from commits
- Add footer with compare links for version diffs
- Include [Unreleased] section for accumulating changes

---

## Summary: Technology Stack

| Component | Selected Tool | Confidence |
|-----------|---------------|------------|
| Changelog Generation | release-please (simple type) | ✅ High |
| Release Automation | google-github-actions/release-please-action@v4 | ✅ High |
| Artifact Upload | softprops/action-gh-release@v2 | ✅ High |
| Version Storage | VERSION file + git tags | ✅ High |
| Pre-release Format | SemVer 2.0 (-alpha.N, -beta.N, -rc.N) | ✅ High |
| Changelog Format | Keep a Changelog | ✅ High |
| Checksum Format | SHA256 (checksums.sha256) | ✅ High |

---

## Open Questions Resolved

| Question | Resolution |
|----------|------------|
| Which changelog tool? | release-please - native GitHub Actions, simple type for non-npm |
| How to handle pre-releases? | SemVer 2.0 pre-release identifiers with release-please config |
| Where to store version? | VERSION file (primary) + git tags (authoritative) |
| How to attach artifacts? | softprops/action-gh-release@v2 with glob patterns |
| Checksum format? | SHA256 in single checksums.sha256 file |
| Manual vs automatic releases? | Hybrid: release-please PRs (auto) + workflow_dispatch (manual override) |

---

## References

- [release-please](https://github.com/googleapis/release-please) - Google's release automation
- [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) - Changelog format standard
- [Semantic Versioning 2.0.0](https://semver.org/) - Version numbering standard
- [softprops/action-gh-release](https://github.com/softprops/action-gh-release) - GitHub Release action
- [Conventional Commits](https://www.conventionalcommits.org/) - Commit message standard
