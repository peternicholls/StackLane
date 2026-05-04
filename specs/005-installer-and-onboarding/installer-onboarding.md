# StageServe Installer and Onboarding

This document covers the one-line installer, binary integrity verification, and the canonical first-run command sequence.

## Platform Compatibility Matrix

| OS            | Architecture | Installer | `setup` DNS | `setup` mkcert | Notes                              |
|---------------|--------------|-----------|-------------|----------------|------------------------------------|
| macOS         | arm64        | ✓         | ✓ (dnsmasq) | ✓              | Primary target                     |
| macOS         | x86_64       | ✓         | ✓ (dnsmasq) | ✓              |                                    |
| Linux         | arm64        | ✓         | manual only | manual only    | DNS/mkcert steps print remediation |
| Linux         | x86_64       | ✓         | manual only | manual only    | DNS/mkcert steps print remediation |
| Windows       | —            | ✗         | ✗           | ✗              | Not supported                      |

## Installing the Binary

### One-line install (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/peternicholls/StageServe/master/install.sh | bash
```

The installer:
1. Detects OS and CPU architecture.
2. Constructs the release asset name: `stage_<version>_<OS>_<arch>`.
3. Downloads the binary from the GitHub Release.
4. Fetches and verifies the SHA-256 checksum (`<asset>.sha256`).
5. Installs to `$STAGESERVE_INSTALL_DIR` (default `$HOME/.local/bin`).
6. Warns if the install directory is not on `$PATH`.
7. Hands off to `stage setup --tui` when running in a TTY, or prints next steps in non-interactive mode.

### Non-interactive install

```bash
NONINTERACTIVE=1 curl -fsSL https://raw.githubusercontent.com/peternicholls/StageServe/master/install.sh | bash
```

### Custom install directory

```bash
STAGESERVE_INSTALL_DIR="$HOME/bin" curl -fsSL ... | bash
```

## Manual integrity verification

If you prefer to download and verify the binary yourself:

```bash
VERSION="v0.1.0"
ASSET="stage_${VERSION}_Darwin_arm64"

# Download
curl -fsSL "https://github.com/peternicholls/StageServe/releases/download/${VERSION}/${ASSET}" -o stage
curl -fsSL "https://github.com/peternicholls/StageServe/releases/download/${VERSION}/${ASSET}.sha256" -o stage.sha256

# Verify
shasum -a 256 -c stage.sha256   # should print: stage: OK

# Install
chmod +x stage
mv stage ~/.local/bin/
```

## First-run sequence

After install, run these commands once per machine:

```bash
stage setup           # Check machine readiness; follow remediation prompts
stage doctor          # Verify no drift after manual remediation steps
```

For each project you want to bring under StageServe management:

```bash
cd /path/to/your/project
stage init            # Writes .env.stageserve with documented defaults
stage up              # Starts the project stack
stage status          # Verify healthy
```

## Repo-to-deployed-stack sync

StageServe is not installed from a clone. The binary on `$PATH` is independent of any source checkout. After updating the binary (re-run the installer), build state is unaffected.

If you also maintain a local source checkout, do not mix the source tree with the installed binary path. Use `$STACK_HOME` (pointing to the source checkout) only if you intend to run from source.

## Installer exit codes

| Code | Meaning                                         |
|------|-------------------------------------------------|
| 0    | Binary installed and handoff completed          |
| 1    | Unsupported OS / architecture                   |
| 2    | Download or checksum verification failed        |
| 3    | Install directory not writable                  |
