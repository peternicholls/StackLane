# Workflow Events Specification

**Feature**: 004-release-workflow  
**Date**: 2025-12-28

## Overview

This document defines the GitHub Actions events that trigger release workflows and the data payloads they carry.

---

## Event: Push to Main

### Trigger
```yaml
on:
  push:
    branches:
      - main
```

### Relevant Payload Fields
```json
{
  "ref": "refs/heads/main",
  "before": "<previous-commit-sha>",
  "after": "<new-commit-sha>",
  "commits": [
    {
      "id": "<commit-sha>",
      "message": "feat: add new feature\n\nDescription of feature",
      "author": {
        "name": "Author Name",
        "email": "author@example.com"
      }
    }
  ],
  "head_commit": {
    "id": "<commit-sha>",
    "message": "feat: add new feature"
  }
}
```

### Workflow Response
1. release-please analyzes commits since last release
2. If releasable commits found:
   - Creates/updates Release PR with version bump and CHANGELOG
3. If Release PR is merged:
   - Creates git tag
   - Creates GitHub Release
   - Triggers artifact build job

---

## Event: Tag Created (via release-please)

### Trigger
```yaml
on:
  push:
    tags:
      - 'v*'
```

### Relevant Payload Fields
```json
{
  "ref": "refs/tags/v2.0.0",
  "ref_type": "tag",
  "base_ref": "refs/heads/main"
}
```

### Workflow Response
1. Checkout code at tagged commit
2. Build release artifacts
3. Generate checksums
4. Upload assets to GitHub Release

---

## Event: Pull Request (Validation)

### Trigger
```yaml
on:
  pull_request:
    branches: [main]
    types: [opened, synchronize, reopened]
```

### Relevant Payload Fields
```json
{
  "action": "opened|synchronize|reopened",
  "number": 123,
  "pull_request": {
    "title": "feat: add new feature",
    "head": {
      "ref": "feature-branch",
      "sha": "<commit-sha>"
    },
    "base": {
      "ref": "main"
    }
  }
}
```

### Workflow Response
1. Validate PR title follows conventional commit format
2. Run ShellCheck on scripts
3. Validate Docker Compose syntax
4. Post changelog preview comment

---

## Event: Workflow Dispatch (Manual Release)

### Trigger
```yaml
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to release (e.g., 2.0.0)'
        required: true
        type: string
      prerelease:
        description: 'Pre-release type'
        required: false
        type: choice
        options:
          - stable
          - alpha
          - beta
          - rc
        default: stable
```

### Use Cases
- Force release with specific version
- Create pre-release versions
- Re-release after failed artifact build

### Workflow Response
1. Validate version input
2. Create/update VERSION file
3. Create git tag
4. Build and upload artifacts
5. Create GitHub Release

---

## Event Flow Diagram

```
┌─────────────────┐
│   Developer     │
│   pushes to     │
│   main branch   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     No releasable     ┌─────────────┐
│  release-please │ ───────commits──────> │   No action │
│   analyzes      │                       └─────────────┘
│   commits       │
└────────┬────────┘
         │ Releasable commits found
         ▼
┌─────────────────┐
│  Release PR     │
│  created/       │
│  updated        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     PR not merged     ┌─────────────┐
│  Maintainer     │ ────────────────────> │   Pending   │
│  reviews PR     │                       └─────────────┘
└────────┬────────┘
         │ PR merged
         ▼
┌─────────────────┐
│  release-please │
│  creates tag    │
│  & release      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Artifact job   │
│  builds &       │
│  uploads        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Release        │
│  published      │
│  with assets    │
└─────────────────┘
```

---

## Commit Message Format

### Conventional Commits Specification

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types and Version Impact

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat` | New feature | Minor |
| `fix` | Bug fix | Patch |
| `docs` | Documentation only | Patch |
| `style` | Code style (no logic change) | None |
| `refactor` | Code refactoring | Patch |
| `perf` | Performance improvement | Patch |
| `test` | Adding/updating tests | None |
| `build` | Build system changes | None |
| `ci` | CI configuration changes | None |
| `chore` | Maintenance tasks | None |
| `revert` | Revert previous commit | Patch |

### Breaking Changes

```
feat!: remove deprecated API

BREAKING CHANGE: The old API has been removed.
```

Breaking changes trigger a **major** version bump regardless of commit type.

---

## Release PR Format

### Title
```
chore(main): release 2.0.0
```

### Body
```markdown
## [2.0.0](https://github.com/owner/repo/compare/v1.0.0...v2.0.0) (2025-12-28)

### Features

* add new feature ([abc1234](https://github.com/owner/repo/commit/abc1234))

### Bug Fixes

* fix critical bug ([def5678](https://github.com/owner/repo/commit/def5678))

---
This PR was generated by release-please.
```

### Labels
- `autorelease: pending` - PR is awaiting merge
- `autorelease: tagged` - PR was merged and tag created
