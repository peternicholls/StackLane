# Release Workflow Contract

**Type**: GitHub Actions Workflow  
**Trigger**: Push to main (via release-please PR merge)  
**File**: `.github/workflows/release.yml`

## Overview

This workflow handles the complete release lifecycle:
1. release-please creates/updates Release PRs from conventional commits
2. Merging a Release PR triggers tag creation and GitHub Release
3. Tag creation triggers artifact building and attachment

---

## Workflow: release-please.yml

### Trigger Events
```yaml
on:
  push:
    branches:
      - main
```

### Inputs
None - fully automated from conventional commits.

### Outputs
```yaml
outputs:
  release_created:
    description: "Whether a release was created"
    value: ${{ steps.release.outputs.release_created }}
  tag_name:
    description: "The release tag name (e.g., v2.0.0)"
    value: ${{ steps.release.outputs.tag_name }}
  version:
    description: "The release version (e.g., 2.0.0)"
    value: ${{ steps.release.outputs.version }}
  upload_url:
    description: "URL for uploading release assets"
    value: ${{ steps.release.outputs.upload_url }}
```

### Permissions Required
```yaml
permissions:
  contents: write       # Create tags, releases, push VERSION updates
  pull-requests: write  # Create and update Release PRs
```

### Jobs

#### Job: release-please
```yaml
release-please:
  runs-on: ubuntu-latest
  outputs:
    release_created: ${{ steps.release.outputs.release_created }}
    tag_name: ${{ steps.release.outputs.tag_name }}
    version: ${{ steps.release.outputs.version }}
  steps:
    - uses: googleapis/release-please-action@v4
      id: release
      with:
        release-type: simple
        config-file: release-please-config.json
        manifest-file: .release-please-manifest.json
```

#### Job: build-artifacts (conditional)
```yaml
build-artifacts:
  needs: release-please
  if: ${{ needs.release-please.outputs.release_created }}
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Build release archive
      run: scripts/release/artifacts.sh ${{ needs.release-please.outputs.version }}
    - name: Upload artifacts
      uses: softprops/action-gh-release@v2
      with:
        tag_name: ${{ needs.release-please.outputs.tag_name }}
        files: |
          dist/20i-stack-*.tar.gz
          dist/install.sh
          dist/checksums.sha256
```

---

## Workflow: validate-pr.yml

### Trigger Events
```yaml
on:
  pull_request:
    branches: [main]
    types: [opened, synchronize, reopened]
```

### Purpose
Validates PRs for:
- Conventional commit format in PR title
- Docker Compose syntax validity
- ShellCheck linting for scripts

### Steps
```yaml
steps:
  - uses: actions/checkout@v4

  - name: Validate PR title (conventional commit)
    uses: amannn/action-semantic-pull-request@v5
    env:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    with:
      types: |
        feat
        fix
        docs
        style
        refactor
        perf
        test
        build
        ci
        chore
        revert

  - name: Lint shell scripts
    uses: ludeeus/action-shellcheck@master
    with:
      scandir: './scripts'

  - name: Validate Docker Compose
    run: docker compose config --quiet
```

### Outputs
None - validation-only workflow.

---

## Workflow: changelog-preview.yml

### Trigger Events
```yaml
on:
  pull_request:
    branches: [main]
    types: [opened, synchronize]
```

### Purpose
Adds a comment to PRs showing what the changelog entry will look like when merged.

### Steps
```yaml
steps:
  - uses: actions/checkout@v4
    with:
      fetch-depth: 0

  - name: Generate changelog preview
    id: preview
    run: |
      # Generate preview from PR commits
      echo "preview<<EOF" >> $GITHUB_OUTPUT
      scripts/release/changelog-preview.sh >> $GITHUB_OUTPUT
      echo "EOF" >> $GITHUB_OUTPUT

  - name: Comment on PR
    uses: actions/github-script@v7
    with:
      script: |
        github.rest.issues.createComment({
          owner: context.repo.owner,
          repo: context.repo.repo,
          issue_number: context.issue.number,
          body: `## 📋 Changelog Preview\n\n${{ steps.preview.outputs.preview }}`
        })
```

---

## API: scripts/release/artifacts.sh

### Synopsis
```bash
scripts/release/artifacts.sh <version>
```

### Arguments
| Argument | Type | Required | Description |
|----------|------|----------|-------------|
| version | string | Yes | Version number without 'v' prefix (e.g., "2.0.0") |

### Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| DIST_DIR | `dist/` | Output directory for artifacts |

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Invalid arguments |
| 2 | Build failure |

### Output Files
```
dist/
├── 20i-stack-v{version}.tar.gz
├── install.sh
└── checksums.sha256
```

### Archive Contents
```
20i-stack-v{version}/
├── docker-compose.yml
├── docker/
│   ├── nginx.conf.tmpl
│   └── apache/
│       ├── Dockerfile
│       ├── httpd.conf
│       └── php.ini
├── config/
│   └── stack-vars.yml
├── scripts/
│   └── setup-local.sh
├── 20i-gui
├── zsh-example-script.zsh
├── .env.example
├── README.md
├── LICENSE
└── VERSION
```

---

## API: scripts/release/validate.sh

### Synopsis
```bash
scripts/release/validate.sh [--version <version>] [--changelog] [--all]
```

### Options
| Option | Description |
|--------|-------------|
| `--version <ver>` | Validate specific version format |
| `--changelog` | Validate CHANGELOG has entry for current version |
| `--all` | Run all validations |

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | All validations passed |
| 1 | Version format invalid |
| 2 | CHANGELOG validation failed |
| 3 | Missing required files |

### Example
```bash
$ scripts/release/validate.sh --all
✅ VERSION file valid: 2.0.0
✅ CHANGELOG has entry for 2.0.0
✅ All required files present
```

---

## API: scripts/release/version.sh

### Synopsis
```bash
scripts/release/version.sh [get|bump <type>|set <version>]
```

### Commands
| Command | Description |
|---------|-------------|
| `get` | Print current version from VERSION file |
| `bump major` | Increment major version (X.0.0) |
| `bump minor` | Increment minor version (x.X.0) |
| `bump patch` | Increment patch version (x.x.X) |
| `set <ver>` | Set version to specific value |

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Invalid command or version |

### Example
```bash
$ scripts/release/version.sh get
1.0.0

$ scripts/release/version.sh bump minor
1.1.0

$ scripts/release/version.sh set 2.0.0-alpha.1
2.0.0-alpha.1
```

---

## Error Handling

### Workflow Failures
| Failure | Response | Recovery |
|---------|----------|----------|
| Release PR merge conflict | Block merge | Manual conflict resolution |
| Artifact build failure | Fail workflow | Fix and re-trigger |
| Asset upload failure | Retry 3x | Manual upload via `gh release upload` |
| Duplicate tag | Fail workflow | Delete tag manually, re-run |

### Concurrency
```yaml
concurrency:
  group: release
  cancel-in-progress: false  # Never cancel in-progress releases
```

---

## Security Considerations

1. **GITHUB_TOKEN permissions**: Minimum required (contents: write, pull-requests: write)
2. **No secrets in artifacts**: Archive excludes .env, .20i-local, etc.
3. **Signed tags**: Consider GPG signing for high-security releases
4. **Checksum verification**: Always verify downloads against checksums.sha256
