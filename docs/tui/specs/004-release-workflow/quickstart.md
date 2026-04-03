# Quickstart: Release Workflow for Maintainers

**Feature**: 004-release-workflow  
**Audience**: Project maintainers  
**Time**: 5 minutes to understand, 2 minutes per release

---

## How Releases Work

The 20i Stack uses **automated release management** powered by [release-please](https://github.com/googleapis/release-please). Here's what happens automatically:

1. **You merge PRs** with [conventional commit](https://www.conventionalcommits.org/) messages
2. **release-please creates a Release PR** that accumulates your changes
3. **You merge the Release PR** when ready to release
4. **GitHub Release is created** with changelog and artifacts

```
feat: add X  ──┐
fix: bug Y   ──┼──> Release PR created ──> Merge ──> v2.0.0 released!
docs: update ──┘
```

---

## For Contributors: Commit Message Format

All commits must follow **Conventional Commits** format:

```bash
# Features (triggers MINOR version bump)
git commit -m "feat: add new configuration option"

# Bug fixes (triggers PATCH version bump)
git commit -m "fix: resolve port detection issue"

# Documentation (triggers PATCH version bump)
git commit -m "docs: update README with ARM instructions"

# Breaking changes (triggers MAJOR version bump)
git commit -m "feat!: change default PHP version to 8.5

BREAKING CHANGE: Projects using PHP 7.x must update their configuration."
```

### Quick Reference

| Prefix | Description | Version Impact |
|--------|-------------|----------------|
| `feat:` | New feature | Minor (1.X.0) |
| `fix:` | Bug fix | Patch (1.0.X) |
| `docs:` | Documentation | Patch |
| `perf:` | Performance | Patch |
| `feat!:` or `BREAKING CHANGE:` | Breaking change | Major (X.0.0) |

---

## For Maintainers: Creating a Release

### Automatic Flow (Recommended)

1. **Merge feature PRs** to `main` with conventional commits
2. **Review the Release PR** that release-please creates automatically
   - Title: `chore(main): release X.Y.Z`
   - Contains auto-generated CHANGELOG entries
3. **Merge the Release PR** when ready to release
4. **Done!** Tag is created, GitHub Release is published, artifacts are attached

### Manual Override (Rare Cases)

Use the GitHub Actions UI for manual control:

1. Go to **Actions** → **Release** workflow
2. Click **Run workflow**
3. Enter version (e.g., `2.0.0` or `2.0.0-beta.1`)
4. Select pre-release type if applicable
5. Click **Run workflow**

---

## Pre-releases (Alpha/Beta/RC)

For testing before stable release:

```bash
# Alpha - early testing
git commit -m "feat: experimental feature"
# Then use workflow_dispatch with version "2.0.0-alpha.1"

# Beta - feature complete, testing stability
# Use version "2.0.0-beta.1"

# Release Candidate - final testing
# Use version "2.0.0-rc.1"
```

Pre-releases are marked separately on GitHub Releases and won't show as "Latest".

---

## Checking Release Status

### Current Version
```bash
cat VERSION
# Output: 1.0.0
```

### View Release History
- **GitHub**: [Releases page](../../releases)
- **CHANGELOG**: [CHANGELOG.md](../../CHANGELOG.md)

### Badge in README
The version badge automatically updates:
```markdown
![Version](https://img.shields.io/github/v/release/peternicholls/20i-stack)
```

---

## Release Artifacts

Each release includes:

| File | Description |
|------|-------------|
| `20i-stack-vX.Y.Z.tar.gz` | Complete distribution archive |
| `install.sh` | Standalone quick installer |
| `checksums.sha256` | SHA256 verification hashes |

### Verify Downloads
```bash
# Download the checksum file and archive
curl -LO https://github.com/peternicholls/20i-stack/releases/download/vX.Y.Z/checksums.sha256
curl -LO https://github.com/peternicholls/20i-stack/releases/download/vX.Y.Z/20i-stack-vX.Y.Z.tar.gz

# Verify
sha256sum -c checksums.sha256
```

---

## Troubleshooting

### "Release PR not appearing"
- Ensure commits use conventional format (`feat:`, `fix:`, etc.)
- Check Actions tab for release-please workflow runs
- Verify push was to `main` branch

### "Release failed after merging PR"
- Check Actions tab for error details
- Artifact build may have failed
- Manually re-run the workflow from Actions tab

### "Wrong version released"
- For pre-releases: Use workflow_dispatch with correct version
- For version correction: Delete tag and release, then re-run

### "Need to skip a release"
- Close the Release PR without merging
- Changes will accumulate in the next Release PR

---

## Configuration Files

These files control release behavior:

| File | Purpose |
|------|---------|
| `VERSION` | Current version number |
| `CHANGELOG.md` | Release history |
| `release-please-config.json` | release-please settings |
| `.release-please-manifest.json` | Version tracking |
| `.github/workflows/release.yml` | Release workflow |

---

## Summary Checklist

- [ ] Use conventional commit messages (`feat:`, `fix:`, `docs:`, etc.)
- [ ] Review and merge the auto-generated Release PR
- [ ] Verify release on GitHub Releases page
- [ ] Announce release if significant (blog, social media)
