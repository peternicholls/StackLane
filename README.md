# StageServe - Shared Hosting Stack Emulation With Docker

## Overview

StageServe is a workflow for local Docker development that aims to mirror the shared hosting environment of 20i webhosting services. To achieve this, it introduces a command/runtime layer plus a shared gateway split, so per-project runtimes are fronted by one persistent gateway while hostname and DNS setup continue to mature.

StageServe centrally defines the 20i-style local stack contract. Each project taps into that shared model through project-local config such as hostname, docroot, versions, and optional bootstrap behavior rather than redefining the stack shape.

The command surface is implemented as a single Go binary (`stage-bin`, exposed as `stage`). The Bash implementation is archived in `previous-version-archive/` for reference only.

### What is implemented now:

- `stage` is the canonical CLI entrypoint, with subcommands such as `up`, `attach`, `status`, and `down`.
- The runtime is a single statically-linked Go binary; no language runtime is required to run it.
- Root-level `20i-*` wrapper entrypoints are not part of the active runtime.
- Project config is resolved consistently with this precedence: CLI flags, then project-root `.env.stageserve`, then shell environment, then stack-home `.env.stageserve`, then built-in defaults.
- Project identity is standardized around a slug and a `.test` (or configured) hostname.
- Project state is recorded as one JSON file per project under `.stageserve-state/projects/<slug>.json`.
- StageServe keeps shared hostname-aware routing available for attached projects, reuses it when already healthy, and repairs it when missing.
- Per-project web containers are isolated behind the shared Docker network instead of publishing host ports directly.
- Project code is mounted internally at `/home/sites/<project-slug>/...` to mirror a 20i-style hosting layout.
- Per-project runtimes get deterministic Docker names: compose project `stage-<slug>`, network `stage-<slug>-runtime`, DB volume `stage-<slug>-db-data`. Shared routing resources stay distinct and StageServe-managed.
- Healthcheck-driven readiness: `stage up` blocks until nginx, apache/PHP-FPM, and MariaDB report healthy (default 120 s, override via `--wait-timeout` or `STAGESERVE_WAIT_TIMEOUT`).
- phpMyAdmin is opt-in via the `debug` compose profile.

## Install

The recommended install path — no source build required:

```bash
curl -fsSL https://raw.githubusercontent.com/peternicholls/StageServe/master/install.sh | bash
```

The installer detects your OS and architecture, downloads the matching signed binary from GitHub Releases, verifies the SHA-256 checksum, and places `stage` in `~/.local/bin`. After install it prints the exact next-step command to run.

**Verify manually (fallback path)**:

```bash
# Download binary and checksum
curl -fsSL https://github.com/peternicholls/StageServe/releases/latest/download/stage_<VERSION>_Darwin_arm64 -o stage
curl -fsSL https://github.com/peternicholls/StageServe/releases/latest/download/stage_<VERSION>_Darwin_arm64.sha256 -o stage.sha256
# Verify
shasum -a 256 -c stage.sha256
chmod +x stage && mv stage ~/.local/bin/
```

**Canonical first-run sequence** (after install):

```bash
stage setup     # machine-readiness checks + one-time DNS/mkcert setup
stage init      # initialize project config in your repo root (optional)
stage up        # bring the project stack online
stage doctor    # diagnose drift at any time
```

## Quick Start

From the stack repo itself or a deployed copy of it, add the scripts to your shell path and run StageServe from a project root:

```bash
export STACK_HOME="$HOME/docker/stage"

cd /path/to/project
"$STACK_HOME/stage" dns-setup --site-suffix develop
"$STACK_HOME/stage" up --site-suffix develop
"$STACK_HOME/stage" status
"$STACK_HOME/stage" down
```

Optional overrides:

```bash
"$STACK_HOME/stage" up --php-version 8.4
"$STACK_HOME/stage" up --docroot web --site-name marketing-site
"$STACK_HOME/stage" status --project marketing-site
```

## First-time Setup

Requirements: macOS, Docker Desktop, and Homebrew. Installing the binary requires no language runtime; building from source requires Go 1.26.2+.

```bash
# 1. Clone the stack
git clone https://github.com/peternicholls/StageServe.git ~/docker/stage
cd ~/docker/stage

# 2. Build the binary (or download a release artifact)
make build           # produces ./stage-bin

# 3. Add StageServe to your PATH (in ~/.zshrc, then reload)
export STACK_HOME="$HOME/docker/stage"
export PATH="$STACK_HOME:$PATH"

# 4. Bootstrap local DNS (once per machine, macOS only)
stage dns-setup
```

The `stage` shim at the repo root execs `stage-bin`. Invoke commands as `stage <subcommand>`.

The GitHub repository and the local folder that contains it are separate concerns. The remote repository is now named `StageServe`, but existing local checkout directories do not rename themselves. Keep `STACK_HOME` pointed at the folder you actually run, whether that folder is still named `stage` or you rename it manually.

For manual runtime validation, the live StageServe installation on your `PATH` is the authoritative surface. If you edit a different checkout than the one you actually run, rebuild or sync that live install first before treating observed runtime behavior as validation evidence. `$HOME/docker/20i-stack` is one local example deployment path, not a universal product rule.

If `stage dns-setup` requires elevated privileges it prints the exact `sudo` command to finish the resolver file installation. Run it once — it persists across reboots.

The project examples in this README use `.develop` as the preferred local hostname suffix. When opening a site in Safari or VS Code Simple Browser, type the full URL with the scheme, for example `http://my-project.develop/`. Entering only `my-project.develop` may be treated as a search query instead of a hostname.

If you use `.dev`, the local HTTPS URL defaults to port `8443`. This avoids collisions with other local services that commonly use `443`, such as Tailscale Serve, while keeping the route stable and predictable.

For a migration walk-through if you are coming from the old single-project localhost workflow, see [docs/migration.md](docs/migration.md).

## Command Semantics

- `stage up`: Ensure shared routing is available, start the current project runtime, validate the live containers, register it in `.stageserve-state`, and mark it `attached`.
- `stage attach`: Attach-or-bootstrap the current project runtime, reuse the running shared routing layer when healthy, and repair route generation when it is missing.
- `stage down`: Stop only the current project runtime and retain its record with state `down`.
- `stage detach`: Stop only the current project runtime and remove its attachment record.
- `stage down --all`: Stop every known runtime and remove all recorded attachment state.
- `stage status [--project SELECTOR]`: Show shared routing health plus recorded projects, their planned hostnames, hostname route URLs, gateway probe URL, container docroots, registry file path, recorded live container identity, registry drift, and Docker state.
- `stage logs [--project SELECTOR] [service]`: Follow logs for a selected project runtime.
- `stage dns-setup`: Bootstrap local suffix resolution on macOS using Homebrew `dnsmasq` on `127.0.0.1:53535` and an `/etc/resolver/<suffix>` file. The documented examples prefer `.develop`.

When `.dev` TLS is enabled, `stage up` and `stage status` surface the route as `https://<hostname>:8443`.

## Config Precedence

Config is resolved in this order:

1. CLI flags such as `--php-version`, `--docroot`, or `--site-name`
2. Project-root `.env.stageserve`
3. Current shell environment
4. Stack-wide `<stack-home>/.env.stageserve`
5. Built-in defaults

The same filename now serves both human-owned config scopes, and location is the contract:

- `<stack-home>/.env.stageserve`: stack-owned shared defaults for one installed StageServe copy
- `<project>/.env.stageserve`: project-local user overrides for that repo
- `<stack-home>/.stageserve-state/envfiles/*.env`: machine-generated runtime files, not for manual editing

The stack-wide `.env.stageserve` is still the only stack-defaults source StageServe reads — there is no `<stack-home>/.env` fallback (FR-014).
`STAGESERVE_POST_UP_COMMAND` is the one project-local escape hatch intended for app bootstrap, such as migrations, after StageServe has already declared the containers healthy. It is honored **only** when set in the project's `.env.stageserve` (FR-016) — it is intentionally ignored if present in the stack-home `.env.stageserve` or the shell so that one project cannot smuggle a hook into another.

`STAGESERVE_STACK` makes the intended stack explicit. The current runtime only implements `20i`, but the key is reserved so future stacks such as a lighter `laravel` or `node` runtime can be introduced without inventing another config surface.

If `stage up` or `stage attach` runs in a repo that does not have a project `.env.stageserve` yet, StageServe proceeds with defaults and writes a starter file for later edits instead of blocking first-run setup.

Choosing which installed stack a project points at remains a machine-level concern for now. Use `STACK_HOME` or `--stack-home` for that selection until StageServe grows a dedicated multi-step install/setup flow for projects.

## Project `.env.stageserve` Contract

Create `.env.stageserve` in your project root using simple `KEY=value` or `export KEY=value` syntax. If the file is missing, `stage up` and `stage attach` create a starter project file automatically on first run:

```bash
export STAGESERVE_STACK=20i
export SITE_NAME=my-site
export DOCROOT=public_html
export PHP_VERSION=8.4
export MYSQL_DATABASE=my_site
export MYSQL_USER=my_site
export MYSQL_PASSWORD=devpass
```

Supported keys:

- `STAGESERVE_STACK`: Explicit stack kind. Current runtime support is `20i` only; other values are rejected until those stacks exist.
- `SITE_NAME`: Base value used to derive the project slug and planned hostname
- `SITE_HOSTNAME`: Full hostname override when you do not want `<slug>.test`
- `SITE_SUFFIX`: Hostname suffix override. Stage one defaults to `.test`
- `DOCROOT`: Document root relative to the project root or an absolute path
- `CODE_DIR`: Alias for `DOCROOT`
- `PHP_VERSION`
- `MYSQL_VERSION`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `MYSQL_PORT`, `PMA_PORT`: Optional per-project published port overrides
- `LOCAL_DNS_PROVIDER`, `LOCAL_DNS_IP`, `LOCAL_DNS_PORT`, `LOCAL_DNS_SUFFIX`: Local DNS bootstrap defaults
- `STAGESERVE_POST_UP_COMMAND`: Optional command run inside the `apache` container after healthchecks pass. Example: `php artisan migrate --force --no-interaction`

Ownership line:

- Project `.env.stageserve` is the user-editable override surface.
- Stack-home `.env.stageserve` is the shared baseline for one installed StageServe copy.
- `.stageserve-state/envfiles/*.env` is machine-made runtime material and should not be edited.
- Shared gateway settings are runtime-owned and no longer part of the user env contract.

Default document root behavior:

- If `DOCROOT` or `CODE_DIR` is set, that value is used.
- Otherwise, `public_html` is used when present.
- Otherwise, the project root is mounted.

Current container path model:

- Project root mounts at `/home/sites/<project-slug>`
- `public_html` becomes `/home/sites/<project-slug>/public_html`
- A custom `DOCROOT` becomes `/home/sites/<project-slug>/<docroot-relative-path>`

Current runtime naming model:

- Compose project: `stage-<slug>` by default
- Runtime network: `<compose-project>-runtime`
- Database volume: `<compose-project>-db-data`
- Web alias on the shared network: `<compose-project>-web`
- Shared routing resources use StageServe-managed internal names that remain separate from the per-project `stage-` runtime names.
- State file: `.stageserve-state/projects/<slug>.json`
- Stack registry: derived from the JSON state directory (no positional `registry.tsv` file)

That mapping is what ties live Docker resources back to the repo path and planned hostname recorded in state.

## Current Access Model

The current implementation now generates hostname-aware gateway rules from the stack registry and bootstraps local suffix resolution on macOS through Homebrew `dnsmasq`. The documented examples below prefer `.develop`.

- Planned hostname and routed hostname: `my-project.develop`
- Manual gateway probe URL: `http://localhost` or another configured shared gateway port
- DNS implementation: `dnsmasq` on `127.0.0.1:53535`
- Resolver file: `/etc/resolver/<suffix>`
- Bootstrap command: `stage dns-setup --site-suffix develop`
- If resolver installation still needs elevated privileges, the command prints the exact `sudo` copy step to finish setup
- Browser note: enter the full URL, for example `http://my-project.develop/`, rather than only the hostname
- Project databases and phpMyAdmin still publish per-project host ports
- MariaDB credentials, database name, and data volume are resolved per project, so project `.env.stageserve` overrides stay isolated to that runtime

This keeps the shell-first workflow intact while removing direct per-project web port publishing from normal site access.

## Default Credentials

- MySQL root: `root` / `root`
- Project database user: defaults to the project slug
- Project database name: defaults to the project slug

## Files of Interest

```text
stage/
├── stage                 # shim that execs stage-bin
├── stage-bin             # compiled Go binary (built by `make build`)
├── cmd/stage/            # cobra root + subcommand wiring
├── core/                     # config, project, state, lifecycle (operator semantics)
├── infra/                    # docker SDK, compose subprocess, gateway template
├── platform/                 # ports, dns, tls (host integrations)
├── observability/            # status, logs (read-only reporting)
├── internal/mocks/           # interface mocks for unit tests
├── docker-compose.20i.yml    # 20i per-project runtime template (with healthchecks; phpMyAdmin under `debug` profile)
├── docker-compose.shared.yml # shared gateway and network
├── docker/
│   └── nginx.conf.tmpl       # reference nginx template (Go renderer is authoritative)
├── .env.stageserve.example    # stack-wide defaults reference (copy to <stack-home>/.env.stageserve)
├── .stageserve-state/         # runtime state (git-ignored)
│   ├── projects/<slug>.json  # per-project state file
│   └── shared/               # generated gateway config
├── docs/
│   ├── architecture.md       # Go module ownership + contribution map
│   ├── contributing.md       # Go workflow, mocks, golden tests
│   ├── migration.md          # older workflow → StageServe guide
│   ├── runtime-contract.md   # command semantics and state model
│   └── plan.md               # historical implementation plan
├── previous-version-archive/ # archived Bash implementation, kept for reference
└── README.md
```

Each attached project creates its own project-root `.env.stageserve`. StageServe still keeps machine-generated envfiles under `.stageserve-state/envfiles/` rather than mixing them into the user-edited config surface.

## Shell Integration

Add this to `.zshrc` if you want the commands globally:

```bash
export STACK_HOME="${STACK_HOME:-$HOME/docker/stage}"
export PATH="$STACK_HOME:$PATH"

alias sl='stage'
alias sstatus='stage status'
alias sup='stage up'
alias sdown='stage down'
```

## Workflow Examples

Single project:

```bash
cd /path/to/project-a
stage up --site-suffix develop
stage status
stage down
```

Concurrent shared-gateway attachment:

```bash
cd /path/to/project-a
stage up --site-suffix develop

cd /path/to/project-b
stage attach --site-name project-b --site-suffix develop

stage status
stage status --project project-b
```

Global teardown:

```bash
stage down --all
```

## Troubleshooting

Check the resolved config without starting containers:

```bash
stage up --dry-run
```

Follow logs:

```bash
stage logs
stage logs apache
```

Reset a specific project by removing its state and volumes only after stopping it:

```bash
stage down
rm -f "$STACK_HOME/.stageserve-state/projects/<slug>.json"
docker volume ls
```

## Requirements

- macOS (Linux DNS bootstrap is a documented "unsupported platform" surface; lifecycle commands work on Linux but `dns-setup` does not)
- Docker Desktop (or Docker Engine ≥ Compose v2)
- Homebrew (only required for `dns-setup`)
- Go 1.26.2+ (only required to build from source; not required to run a downloaded binary)

## Project Status

The Bash implementation has been rewritten as a Go binary (spec [`003-rewrite-language-choices`](specs/003-rewrite-language-choices/spec.md)). The active runtime uses the current StageServe contract: `stage <subcommand>`, location-based `.env.stageserve`, and `.stageserve-state`.
