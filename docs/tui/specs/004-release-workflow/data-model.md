# Data Model: Release Workflow

**Feature**: 004-release-workflow  
**Date**: 2025-12-28  
**Status**: Complete

## Overview

This document defines the data structures, file formats, and configuration schemas for the automated release workflow system.

---

## Entities

### 1. Version

**Description**: Represents the current version of the 20i Stack project following Semantic Versioning 2.0.

**Storage**: `/VERSION` file (plaintext, single line)

**Format**:
```
MAJOR.MINOR.PATCH[-PRERELEASE[.N]]
```

**Schema**:
```yaml
Version:
  major: integer (>= 0)        # Breaking changes
  minor: integer (>= 0)        # New features, backward compatible
  patch: integer (>= 0)        # Bug fixes, backward compatible
  prerelease: string | null    # alpha, beta, rc (optional)
  prerelease_number: integer   # Prerelease iteration (optional, >= 1)
```

**Examples**:
```
1.0.0           # Stable release
2.0.0-alpha.1   # Alpha pre-release
2.0.0-beta.3    # Beta pre-release
2.0.0-rc.1      # Release candidate
```

**Validation Rules**:
- Major, minor, patch must be non-negative integers
- Prerelease identifier must be one of: `alpha`, `beta`, `rc`
- Prerelease number must be >= 1 when present
- No leading zeros in numeric components
- No whitespace or trailing newlines in file

**State Transitions**:
```
[Development] ──feat:──> [Minor Bump]
[Development] ──fix:───> [Patch Bump]
[Development] ──BREAKING CHANGE:──> [Major Bump]
[Any Version] ──prerelease──> [Pre-release]
[Pre-release] ──promote──> [Stable Release]
```

---

### 2. Git Tag

**Description**: Immutable reference to a specific commit marking a release point.

**Format**: `v{VERSION}`

**Schema**:
```yaml
GitTag:
  name: string              # e.g., "v2.0.0"
  commit_sha: string        # 40-character hex SHA
  tagger: string            # GitHub username or bot
  message: string           # Tag annotation message
  created_at: datetime      # ISO 8601 timestamp
  signed: boolean           # GPG signature status
```

**Examples**:
```
v1.0.0
v2.0.0-alpha.1
v2.0.0-rc.1
v2.0.0
```

**Validation Rules**:
- Must start with lowercase `v`
- Version portion must be valid semver
- Must point to commit on main/master branch (for stable releases)
- Must be unique (no duplicate tags)
- Should be annotated (not lightweight) for releases

---

### 3. GitHub Release

**Description**: Published release with notes, artifacts, and download links on GitHub.

**Schema**:
```yaml
GitHubRelease:
  id: integer               # GitHub release ID
  tag_name: string          # Associated git tag (e.g., "v2.0.0")
  name: string              # Release title (e.g., "v2.0.0")
  body: string              # Release notes (Markdown)
  draft: boolean            # Draft status (false for published)
  prerelease: boolean       # Pre-release flag
  created_at: datetime      # Creation timestamp
  published_at: datetime    # Publication timestamp
  author: string            # GitHub username
  assets: ReleaseAsset[]    # Attached artifacts
  target_commitish: string  # Branch or SHA for tag target
```

**Release Types**:
| Type | `prerelease` | `draft` | Visibility |
|------|--------------|---------|------------|
| Stable | `false` | `false` | Listed as "Latest" |
| Pre-release | `true` | `false` | Listed with warning |
| Draft | any | `true` | Hidden from public |

---

### 4. Release Asset

**Description**: File artifact attached to a GitHub Release.

**Schema**:
```yaml
ReleaseAsset:
  id: integer               # GitHub asset ID
  name: string              # Filename (e.g., "20i-stack-v2.0.0.tar.gz")
  label: string | null      # Display label (optional)
  content_type: string      # MIME type
  size: integer             # Size in bytes
  download_count: integer   # Download statistics
  browser_download_url: string  # Direct download URL
  created_at: datetime      # Upload timestamp
  uploaded_by: string       # GitHub username
```

**Required Assets**:
| Asset Name | Content Type | Description |
|------------|--------------|-------------|
| `20i-stack-v{VERSION}.tar.gz` | application/gzip | Main distribution archive |
| `install.sh` | text/x-shellscript | Standalone installer script |
| `checksums.sha256` | text/plain | SHA256 checksums for verification |

**Naming Convention**:
```
{project}-v{version}.{extension}

Examples:
20i-stack-v2.0.0.tar.gz
20i-stack-v2.0.0-alpha.1.tar.gz
```

---

### 5. Changelog Entry

**Description**: A single version entry in CHANGELOG.md following Keep a Changelog format.

**Schema**:
```yaml
ChangelogEntry:
  version: string           # Semantic version (e.g., "2.0.0")
  release_date: date        # YYYY-MM-DD format
  sections:
    - type: string          # Added, Changed, Deprecated, Removed, Fixed, Security
      items: string[]       # List of change descriptions
  compare_url: string       # GitHub compare URL to previous version
```

**Section Types** (per Keep a Changelog):
| Section | Commit Types | Description |
|---------|--------------|-------------|
| Added | `feat:` | New features |
| Changed | `refactor:`, `perf:` | Changes to existing functionality |
| Deprecated | `deprecate:` | Soon-to-be removed features |
| Removed | `remove:` | Removed features |
| Fixed | `fix:` | Bug fixes |
| Security | `security:` | Security vulnerability fixes |

**Example**:
```markdown
## [2.0.0] - 2025-12-28

### Added
- New release workflow with automated changelog generation
- VERSION file for version tracking

### Changed
- Updated PHP default version to 8.5

### Fixed
- Container detection regex for hyphenated project names

[2.0.0]: https://github.com/peternicholls/20i-stack/compare/v1.0.0...v2.0.0
```

---

### 6. Release Configuration

**Description**: Configuration for release behavior, stored in project.

**Storage**: `/config/release/config.yml`

**Schema**:
```yaml
ReleaseConfig:
  # Changelog configuration
  changelog:
    file: string            # Path to CHANGELOG.md (default: "CHANGELOG.md")
    sections:               # Commit type to section mapping
      - type: string        # Conventional commit type
        section: string     # Changelog section name
        hidden: boolean     # Whether to hide from changelog

  # Release behavior
  release:
    draft: boolean          # Create as draft first (default: false)
    prerelease_pattern: string  # Regex for pre-release detection
    assets:                 # Assets to attach
      - name: string        # Asset filename pattern
        path: string        # Source path in repository

  # Version file locations
  versioning:
    file: string            # Primary version file (default: "VERSION")
    sync_files: string[]    # Additional files to update version in
```

**Default Configuration**:
```yaml
changelog:
  file: "CHANGELOG.md"
  sections:
    - type: feat
      section: Features
      hidden: false
    - type: fix
      section: Bug Fixes
      hidden: false
    - type: docs
      section: Documentation
      hidden: false
    - type: perf
      section: Performance
      hidden: false
    - type: refactor
      section: Code Refactoring
      hidden: true
    - type: chore
      section: Maintenance
      hidden: true
    - type: test
      section: Testing
      hidden: true

release:
  draft: false
  prerelease_pattern: ".*-(alpha|beta|rc)\\."
  assets:
    - name: "20i-stack-v*.tar.gz"
      path: "dist/"
    - name: "install.sh"
      path: "dist/"
    - name: "checksums.sha256"
      path: "dist/"

versioning:
  file: "VERSION"
  sync_files: []
```

---

### 7. Release-Please Configuration

**Description**: Configuration for release-please automation, stored at repository root.

**Storage**: `/release-please-config.json`

**Schema**:
```json
{
  "release-type": "simple",
  "bump-minor-pre-major": true,
  "bump-patch-for-minor-pre-major": true,
  "include-component-in-tag": false,
  "tag-separator": "",
  "packages": {
    ".": {
      "changelog-path": "CHANGELOG.md",
      "release-type": "simple"
    }
  },
  "changelog-sections": [
    {"type": "feat", "section": "Features"},
    {"type": "fix", "section": "Bug Fixes"},
    {"type": "docs", "section": "Documentation"},
    {"type": "perf", "section": "Performance"},
    {"type": "refactor", "section": "Code Refactoring", "hidden": true}
  ]
}
```

**Manifest File**: `/.release-please-manifest.json`
```json
{
  ".": "1.0.0"
}
```

---

## Relationships

```
┌─────────────┐     creates     ┌─────────────┐
│   VERSION   │ ───────────────>│   Git Tag   │
│    file     │                 │             │
└─────────────┘                 └──────┬──────┘
                                       │
                                       │ associated with
                                       ▼
┌─────────────┐     generates   ┌─────────────┐
│  Changelog  │ <───────────────│   GitHub    │
│   Entry     │                 │   Release   │
└─────────────┘                 └──────┬──────┘
                                       │
                                       │ contains
                                       ▼
                                ┌─────────────┐
                                │   Release   │
                                │   Assets    │
                                └─────────────┘
```

---

## File Formats

### VERSION File
```
1.0.0
```
- Single line, no trailing newline
- No `v` prefix
- Valid semantic version

### Checksums File (checksums.sha256)
```
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  20i-stack-v2.0.0.tar.gz
a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2  install.sh
```
- BSD-style format: `{hash}  {filename}`
- Two spaces between hash and filename
- One file per line
- SHA256 algorithm (64 hex characters)

---

## Validation Requirements

### Pre-release Validation (FR-004, FR-005)
1. VERSION file exists and contains valid semver
2. CHANGELOG.md has entry for version being released
3. All CI tests pass (lint, docker build test)
4. No duplicate git tags exist

### Release Artifact Validation
1. Archive contains all required files
2. Checksums match generated values
3. Archive extracts successfully
4. install.sh is executable

### Post-release Validation
1. GitHub Release is published (not draft)
2. All assets are attached
3. Release notes match CHANGELOG entry
4. Git tag points to correct commit
