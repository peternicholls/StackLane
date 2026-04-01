# Stacklane - Docker Development Environment

## Overview

Stacklane is a workflow for local Docker development that aims to mirror the shared hosting environment of 20i webhosting services. To achieve this, it introduces a command/runtime layer plus a shared gateway split, so per-project runtimes are fronted by one persistent gateway while hostname and DNS setup continue to mature.

This is a work in progress. The CLI is the primary interface for the implemented runtime contract. 

### What is implemented now:

- `stacklane` is the canonical CLI entrypoint, with action flags such as `--up`, `--attach`, `--status`, and `--down`.
- Legacy wrappers such as `20i-up` and `20i-status` are deprecated, now forward to `stacklane`, and are intended to be removed in a future update.
- Project config is resolved consistently from `.env`, `.20i-local`, and CLI flags.
- Project identity is standardized around a slug and a planned `.test` hostname.
- Project state is recorded under `.20i-state`, with a stack-level `registry.tsv` snapshot for status, detach, and global teardown semantics.
- One shared gateway now owns the host web ports and routes to one attached project at a time.
- Per-project web containers are isolated behind the shared Docker network instead of publishing host ports directly.
- Project code is mounted internally at `/home/sites/<project-slug>/...` to better mirror the 20i-style hosting layout.
- Per-project runtimes now get deterministic Docker names: compose project `20i-<slug>`, network `20i-<slug>-runtime`, and DB volume `20i-<slug>-db-data`.

## Quick Start

From the stack repo itself or a deployed copy of it, add the scripts to your shell path and run Stacklane from a project root:

```bash
export STACK_HOME="$HOME/docker/20i-stack"

cd /path/to/project
"$STACK_HOME/stacklane" --dns-setup
"$STACK_HOME/stacklane" --up
"$STACK_HOME/stacklane" --status
"$STACK_HOME/stacklane" --down
```

Optional overrides:

```bash
"$STACK_HOME/stacklane" --up --php-version 8.4
"$STACK_HOME/stacklane" --up --docroot web --site-name marketing-site
"$STACK_HOME/stacklane" --up version=8.4
"$STACK_HOME/stacklane" --status --project marketing-site
```

## First-time Setup

Requirements: macOS, Docker Desktop, and Homebrew.

```bash
# 1. Clone or copy the stack
git clone https://github.com/peternicholls/20i-stack ~/docker/20i-stack

# 2. Add Stacklane to your path — add to ~/.zshrc and reload
export STACK_HOME="$HOME/docker/20i-stack"
export PATH="$STACK_HOME:$PATH"

# 3. Bootstrap local DNS (once per machine)
stacklane --dns-setup
```

The GitHub repository and the local folder that contains it are separate concerns. If the remote repository is renamed later, your local checkout directory does not rename itself. Keep `STACK_HOME` pointed at the folder you actually run, whether that folder is still named `20i-stack` or you rename it manually.

If `stacklane --dns-setup` requires elevated privileges it prints the exact `sudo` command to finish the resolver file installation. Run it once — it persists across reboots.

If you use `.dev`, the local HTTPS URL defaults to port `8443`. This avoids collisions with other local services that commonly use `443`, such as Tailscale Serve, while keeping the route stable and predictable.

For a migration walk-through if you are coming from the old single-project localhost workflow, see [docs/migration.md](docs/migration.md).

## Command Semantics

- `stacklane --up`: Ensure the shared gateway exists, start the current project runtime, validate the live containers, register it in `.20i-state`, and mark it `attached`.
- `stacklane --attach`: Attach-or-bootstrap the current project runtime and regenerate hostname-aware gateway routes from the registry.
- `stacklane --down`: Stop only the current project runtime and retain its record with state `down`.
- `stacklane --detach`: Stop only the current project runtime and remove its attachment record.
- `stacklane --down --all`: Stop every known runtime and remove all recorded attachment state.
- `stacklane --status [--project SELECTOR]`: Show shared gateway health plus recorded projects, their planned hostnames, hostname route URLs, gateway probe URL, container docroots, registry file path, recorded live container identity, registry drift, and Docker state.
- `stacklane --logs [--project SELECTOR] [service]`: Follow logs for a selected project runtime.
- `stacklane --dns-setup`: Bootstrap local `.test` resolution on macOS using Homebrew `dnsmasq` on `127.0.0.1:53535` and an `/etc/resolver/<suffix>` file.

When `.dev` TLS is enabled, `stacklane --up` and `stacklane --status` surface the route as `https://<hostname>:8443` unless you explicitly override `SHARED_GATEWAY_HTTPS_PORT`.

## Config Precedence

Config is resolved in this order:

1. CLI flags such as `--php-version`, `--docroot`, or `--site-name`
2. Project-local `.20i-local`
3. Current shell environment
4. Stack-wide `.env`
5. Built-in defaults

The stack-wide `.env` is for defaults. `.20i-local` is the project contract.

## `.20i-local` Contract

Create `.20i-local` in your project root using simple `KEY=value` or `export KEY=value` syntax:

```bash
export SITE_NAME=my-site
export DOCROOT=public_html
export PHP_VERSION=8.4
export MYSQL_DATABASE=my_site
export MYSQL_USER=my_site
export MYSQL_PASSWORD=devpass
```

Supported keys:

- `SITE_NAME`: Base value used to derive the project slug and planned hostname
- `SITE_HOSTNAME`: Full hostname override when you do not want `<slug>.test`
- `SITE_SUFFIX`: Hostname suffix override. Stage one defaults to `.test`
- `DOCROOT`: Document root relative to the project root or an absolute path
- `CODE_DIR`: Legacy alias for `DOCROOT`
- `PHP_VERSION`
- `MYSQL_VERSION`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `MYSQL_PORT`, `PMA_PORT`: Optional per-project published port overrides
- `SHARED_GATEWAY_HTTP_PORT`, `SHARED_GATEWAY_HTTPS_PORT`: Shared gateway host port overrides
- `LOCAL_DNS_PROVIDER`, `LOCAL_DNS_IP`, `LOCAL_DNS_PORT`, `LOCAL_DNS_SUFFIX`: Local DNS bootstrap defaults

Default document root behavior:

- If `DOCROOT` or `CODE_DIR` is set, that value is used.
- Otherwise, `public_html` is used when present.
- Otherwise, the project root is mounted.

Current container path model:

- Project root mounts at `/home/sites/<project-slug>`
- `public_html` becomes `/home/sites/<project-slug>/public_html`
- A custom `DOCROOT` becomes `/home/sites/<project-slug>/<docroot-relative-path>`

Current runtime naming model:

- Compose project: `20i-<slug>` by default
- Runtime network: `<compose-project>-runtime`
- Database volume: `<compose-project>-db-data`
- State file: `.20i-state/projects/<slug>.env`
- Registry snapshot: `.20i-state/registry.tsv`

That mapping is what ties live Docker resources back to the repo path and planned hostname recorded in state.

## Current Access Model

The current implementation now generates hostname-aware gateway rules from the stack registry and bootstraps local `.test` resolution on macOS through Homebrew `dnsmasq`.

- Planned hostname and routed hostname: `my-project.test`
- Manual gateway probe URL: `http://localhost` or another configured shared gateway port
- DNS implementation: `dnsmasq` on `127.0.0.1:53535`
- Resolver file: `/etc/resolver/test` by default
- Bootstrap command: `stacklane --dns-setup`
- If resolver installation still needs elevated privileges, the command prints the exact `sudo` copy step to finish setup
- Project databases and phpMyAdmin still publish per-project host ports
- MariaDB credentials, database name, and data volume are resolved per project, so `.20i-local` overrides stay isolated to that runtime

This keeps the shell-first workflow intact while removing direct per-project web port publishing from normal site access.

## Default Credentials

- MySQL root: `root` / `root`
- Project database user: defaults to the project slug
- Project database name: defaults to the project slug

## Files of Interest

```text
20i-stack/
├── stacklane
├── 20i-up
├── 20i-attach
├── 20i-down
├── 20i-detach
├── 20i-dns-setup
├── 20i-status
├── 20i-logs
├── lib/
│   └── 20i-common.sh        # shared config resolution, state helpers
├── docker-compose.yml        # per-project runtime template
├── docker-compose.shared.yml # shared gateway and network
├── docker/
│   └── nginx.conf.tmpl      # gateway route template
├── .env.example              # stack-wide defaults reference
├── .20i-state/               # runtime state (git-ignored)
│   ├── projects/<slug>.env   # per-project state file
│   ├── registry.tsv          # registry snapshot
│   └── shared/               # generated gateway config
├── docs/
│   ├── migration.md          # old-to-new workflow guide
│   ├── runtime-contract.md   # command semantics and state model
│   └── plan.md               # implementation plan and progress
├── AUTOMATION-README.md      # macOS GUI automation docs
└── GUI-HELP.md               # GUI wrapper notes (CLI leads)
```

The `20i-gui-depricated` script in the repo root is the original pre-Phase-1 GUI wrapper. It is kept for reference but does not integrate with the shared gateway or project registry; use `stacklane` instead.

## Shell Integration

Add this to `.zshrc` if you want the commands globally:

```bash
export STACK_HOME="${STACK_HOME:-$HOME/docker/20i-stack}"
export PATH="$STACK_HOME:$PATH"

alias sl='stacklane'
alias sstatus='stacklane --status'
alias sup='stacklane --up'
alias sdown='stacklane --down'
```

## Workflow Examples

Single project:

```bash
cd /path/to/project-a
stacklane --up
stacklane --status
stacklane --down
```

Concurrent shared-gateway attachment:

```bash
cd /path/to/project-a
stacklane --up

cd /path/to/project-b
stacklane --attach --site-name project-b

stacklane --status
stacklane --status --project project-b
```

Global teardown:

```bash
stacklane --down --all
```

## Troubleshooting

Check the resolved config without starting containers:

```bash
stacklane --up --dry-run
```

Follow logs:

```bash
stacklane --logs
stacklane --logs apache
```

Reset a specific project by removing its state and volumes only after stopping it:

```bash
stacklane --down
rm -f "$STACK_HOME/.20i-state/projects/<slug>.env"
docker volume ls
```

## Requirements

- Docker Desktop for Mac
- Bash or Zsh
- Optional: `dialog` for the experimental GUI wrapper

## Phase Notes

Stage one fixes the contract first and keeps `.test` as the canonical future suffix. `.dev` is intentionally deferred until the stack has a proper HTTPS-capable local gateway.

Phase 2 landed the shared gateway and hid per-project web ports behind it. Phase 3 made runtime naming, docroot mapping, PHP selection, and database config explicitly project-specific. Phase 4 added a stack-level registry snapshot plus post-start validation of live container identity. Phase 5 renders hostname-aware gateway rules from that registry. Phase 6 adds macOS `.test` DNS bootstrap around Homebrew `dnsmasq` plus resolver health checks.
