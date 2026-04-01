# 20i Stack - Docker Development Environment

## Overview

20i Stack is moving from a single-project localhost workflow toward a multi-project shared-front-door model. This repo now includes a real command/runtime layer plus a shared gateway split, so per-project runtimes are now fronted by one persistent gateway while the hostname and DNS work continues in later phases.

What is implemented now:

- `20i-up`, `20i-attach`, `20i-down`, `20i-detach`, `20i-status`, `20i-logs`, and `20i-dns-setup` are real repo scripts.
- Project config is resolved consistently from `.env`, `.20i-local`, and CLI flags.
- Project identity is standardized around a slug and a planned `.test` hostname.
- Project state is recorded under `.20i-state`, with a stack-level `registry.tsv` snapshot for status, detach, and global teardown semantics.
- One shared gateway now owns the host web ports and routes to one attached project at a time.
- Per-project web containers are isolated behind the shared Docker network instead of publishing host ports directly.
- Project code is mounted internally at `/home/sites/<project-slug>/...` to better mirror the 20i-style hosting layout.
- Per-project runtimes now get deterministic Docker names: compose project `20i-<slug>`, network `20i-<slug>-runtime`, and DB volume `20i-<slug>-db-data`.

What is not implemented yet:

- Full GUI parity

## Quick Start

From the stack repo itself or a deployed copy of it, add the scripts to your shell path and run them from a project root:

```bash
export STACK_HOME="$HOME/docker/20i-stack"

cd /path/to/project
"$STACK_HOME/20i-dns-setup"
"$STACK_HOME/20i-up"
"$STACK_HOME/20i-status"
"$STACK_HOME/20i-down"
```

Optional overrides:

```bash
"$STACK_HOME/20i-up" --php-version 8.4
"$STACK_HOME/20i-up" --docroot web --site-name marketing-site
"$STACK_HOME/20i-up" version=8.4
"$STACK_HOME/20i-status" --project marketing-site
```

## Command Semantics

- `20i-up`: Ensure the shared gateway exists, start the current project runtime, validate the live containers, register it in `.20i-state`, and mark it `attached`.
- `20i-attach`: Attach-or-bootstrap the current project runtime and regenerate hostname-aware gateway routes from the registry.
- `20i-down`: Stop only the current project runtime and retain its record with state `down`.
- `20i-detach`: Stop only the current project runtime and remove its attachment record.
- `20i-down --all`: Stop every known runtime and remove all recorded attachment state.
- `20i-status [--project SELECTOR]`: Show shared gateway health plus recorded projects, their planned hostnames, hostname route URLs, gateway probe URL, container docroots, registry file path, recorded live container identity, registry drift, and Docker state.
- `20i-logs [--project SELECTOR] [service]`: Follow logs for a selected project runtime.
- `20i-dns-setup`: Bootstrap local `.test` resolution on macOS using Homebrew `dnsmasq` on `127.0.0.1:53535` and an `/etc/resolver/<suffix>` file.

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
- Bootstrap command: `20i-dns-setup`
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
├── 20i-up
├── 20i-attach
├── 20i-down
├── 20i-dns-setup
├── 20i-detach
├── 20i-status
├── 20i-logs
├── lib/20i-common.sh
├── docker-compose.yml
├── docker-compose.shared.yml
├── .env.example
└── docs/plan.md
```

## Shell Integration

Add this to `.zshrc` if you want the commands globally:

```bash
export STACK_HOME="${STACK_HOME:-$HOME/docker/20i-stack}"
export PATH="$STACK_HOME:$PATH"

alias 20i='20i-status'
alias dcu='20i-up'
alias dcd='20i-down'
```

## Workflow Examples

Single project:

```bash
cd /path/to/project-a
20i-up
20i-status
20i-down
```

Concurrent shared-gateway attachment:

```bash
cd /path/to/project-a
20i-up

cd /path/to/project-b
20i-attach --site-name project-b

20i-status
20i-status --project project-b
```

Global teardown:

```bash
20i-down --all
```

## Troubleshooting

Check the resolved config without starting containers:

```bash
20i-up --dry-run
```

Follow logs:

```bash
20i-logs
20i-logs apache
```

Reset a specific project by removing its state and volumes only after stopping it:

```bash
20i-down
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
