# Stacklane - Shared Hosting Stack Emulation With Docker

## Overview

Stacklane is a workflow for local Docker development that aims to mirror the shared hosting environment of 20i webhosting services. To achieve this, it introduces a command/runtime layer plus a shared gateway split, so per-project runtimes are fronted by one persistent gateway while hostname and DNS setup continue to mature.

The command surface is implemented as a single Go binary (`stacklane-bin`, exposed as `stacklane`). The Bash implementation is archived in `previous-version-archive/` for reference only.

### What is implemented now:

- `stacklane` is the canonical CLI entrypoint, with subcommands such as `up`, `attach`, `status`, and `down`.
- The runtime is a single statically-linked Go binary; no language runtime is required to run it.
- Root-level `20i-*` wrapper entrypoints are not part of the active runtime.
- Project config is resolved consistently from stack `.env.stacklane`, project `.stacklane-local`, shell environment, and CLI flags.
- Project identity is standardized around a slug and a `.test` (or configured) hostname.
- Project state is recorded as one JSON file per project under `.stacklane-state/projects/<slug>.json`.
- One shared gateway owns the host web ports and routes to one or more attached projects via hostname-aware nginx rules.
- Per-project web containers are isolated behind the shared Docker network instead of publishing host ports directly.
- Project code is mounted internally at `/home/sites/<project-slug>/...` to mirror a 20i-style hosting layout.
- Per-project runtimes get deterministic Docker names: compose project `stln-<slug>`, network `stln-<slug>-runtime`, DB volume `stln-<slug>-db-data`. Shared infrastructure stays under `stacklane-shared` (compose project + network) and `stacklane-gateway` (container).
- Healthcheck-driven readiness: `stacklane up` blocks until nginx, apache/PHP-FPM, and MariaDB report healthy (default 120 s, override via `--wait-timeout` or `STACKLANE_WAIT_TIMEOUT`).
- phpMyAdmin is opt-in via the `debug` compose profile.

## Quick Start

From the stack repo itself or a deployed copy of it, add the scripts to your shell path and run Stacklane from a project root:

```bash
export STACK_HOME="$HOME/docker/stacklane"

cd /path/to/project
"$STACK_HOME/stacklane" dns-setup
"$STACK_HOME/stacklane" up
"$STACK_HOME/stacklane" status
"$STACK_HOME/stacklane" down
```

Optional overrides:

```bash
"$STACK_HOME/stacklane" up --php-version 8.4
"$STACK_HOME/stacklane" up --docroot web --site-name marketing-site
"$STACK_HOME/stacklane" status --project marketing-site
```

## First-time Setup

Requirements: macOS, Docker Desktop, and Homebrew. Installing the binary requires no language runtime; building from source requires Go 1.26.2+.

```bash
# 1. Clone the stack
git clone https://github.com/peternicholls/StackLane.git ~/docker/stacklane
cd ~/docker/stacklane

# 2. Build the binary (or download a release artifact)
make build           # produces ./stacklane-bin

# 3. Add Stacklane to your PATH (in ~/.zshrc, then reload)
export STACK_HOME="$HOME/docker/stacklane"
export PATH="$STACK_HOME:$PATH"

# 4. Bootstrap local DNS (once per machine, macOS only)
stacklane dns-setup
```

The `stacklane` shim at the repo root execs `stacklane-bin`. Invoke commands as `stacklane <subcommand>`.

The GitHub repository and the local folder that contains it are separate concerns. The remote repository is now named `StackLane`, but existing local checkout directories do not rename themselves. Keep `STACK_HOME` pointed at the folder you actually run, whether that folder is still named `stacklane` or you rename it manually.

If `stacklane dns-setup` requires elevated privileges it prints the exact `sudo` command to finish the resolver file installation. Run it once — it persists across reboots.

If you use `.dev`, the local HTTPS URL defaults to port `8443`. This avoids collisions with other local services that commonly use `443`, such as Tailscale Serve, while keeping the route stable and predictable.

For a migration walk-through if you are coming from the old single-project localhost workflow, see [docs/migration.md](docs/migration.md).

## Command Semantics

- `stacklane up`: Ensure the shared gateway exists, start the current project runtime, validate the live containers, register it in `.stacklane-state`, and mark it `attached`.
- `stacklane attach`: Attach-or-bootstrap the current project runtime and regenerate hostname-aware gateway routes from the registry.
- `stacklane down`: Stop only the current project runtime and retain its record with state `down`.
- `stacklane detach`: Stop only the current project runtime and remove its attachment record.
- `stacklane down --all`: Stop every known runtime and remove all recorded attachment state.
- `stacklane status [--project SELECTOR]`: Show shared gateway health plus recorded projects, their planned hostnames, hostname route URLs, gateway probe URL, container docroots, registry file path, recorded live container identity, registry drift, and Docker state.
- `stacklane logs [--project SELECTOR] [service]`: Follow logs for a selected project runtime.
- `stacklane dns-setup`: Bootstrap local `.test` resolution on macOS using Homebrew `dnsmasq` on `127.0.0.1:53535` and an `/etc/resolver/<suffix>` file.

When `.dev` TLS is enabled, `stacklane up` and `stacklane status` surface the route as `https://<hostname>:8443` unless you explicitly override `SHARED_GATEWAY_HTTPS_PORT`.

## Config Precedence

Config is resolved in this order:

1. CLI flags such as `--php-version`, `--docroot`, or `--site-name`
2. Project-local `.stacklane-local`
3. Current shell environment
4. Stack-wide `<stack-home>/.env.stacklane`
5. Built-in defaults

The stack-wide `.env.stacklane` is the only stack-defaults source Stacklane reads — there is no `<stack-home>/.env` fallback (FR-014). `.stacklane-local` is the Stacklane project contract.
`STACKLANE_POST_UP_COMMAND` is the one project-local escape hatch intended for app bootstrap, such as migrations, after Stacklane has already declared the containers healthy. It is honored **only** when set in the project's `.stacklane-local` (FR-016) — it is intentionally ignored if present in `.env.stacklane` or the shell so that one project cannot smuggle a hook into another.

## `.stacklane-local` Contract

Create `.stacklane-local` in your project root using simple `KEY=value` or `export KEY=value` syntax:

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
- `CODE_DIR`: Alias for `DOCROOT`
- `PHP_VERSION`
- `MYSQL_VERSION`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `MYSQL_PORT`, `PMA_PORT`: Optional per-project published port overrides
- `SHARED_GATEWAY_HTTP_PORT`, `SHARED_GATEWAY_HTTPS_PORT`: Shared gateway host port overrides
- `LOCAL_DNS_PROVIDER`, `LOCAL_DNS_IP`, `LOCAL_DNS_PORT`, `LOCAL_DNS_SUFFIX`: Local DNS bootstrap defaults
- `STACKLANE_POST_UP_COMMAND`: Optional command run inside the `apache` container after healthchecks pass. Example: `php artisan migrate --force --no-interaction`

Default document root behavior:

- If `DOCROOT` or `CODE_DIR` is set, that value is used.
- Otherwise, `public_html` is used when present.
- Otherwise, the project root is mounted.

Current container path model:

- Project root mounts at `/home/sites/<project-slug>`
- `public_html` becomes `/home/sites/<project-slug>/public_html`
- A custom `DOCROOT` becomes `/home/sites/<project-slug>/<docroot-relative-path>`

Current runtime naming model:

- Compose project: `stln-<slug>` by default
- Runtime network: `<compose-project>-runtime`
- Database volume: `<compose-project>-db-data`
- Web alias on the shared network: `<compose-project>-web`
- Shared gateway resources keep the `stacklane-` prefix: compose project `stacklane-shared`, network `stacklane-shared`, container `stacklane-gateway` (FR-013)
- State file: `.stacklane-state/projects/<slug>.json`
- Stack registry: derived from the JSON state directory (no positional `registry.tsv` file)

That mapping is what ties live Docker resources back to the repo path and planned hostname recorded in state.

## Current Access Model

The current implementation now generates hostname-aware gateway rules from the stack registry and bootstraps local `.test` resolution on macOS through Homebrew `dnsmasq`.

- Planned hostname and routed hostname: `my-project.test`
- Manual gateway probe URL: `http://localhost` or another configured shared gateway port
- DNS implementation: `dnsmasq` on `127.0.0.1:53535`
- Resolver file: `/etc/resolver/test` by default
- Bootstrap command: `stacklane dns-setup`
- If resolver installation still needs elevated privileges, the command prints the exact `sudo` copy step to finish setup
- Project databases and phpMyAdmin still publish per-project host ports
- MariaDB credentials, database name, and data volume are resolved per project, so `.stacklane-local` overrides stay isolated to that runtime

This keeps the shell-first workflow intact while removing direct per-project web port publishing from normal site access.

## Default Credentials

- MySQL root: `root` / `root`
- Project database user: defaults to the project slug
- Project database name: defaults to the project slug

## Files of Interest

```text
stacklane/
├── stacklane                 # shim that execs stacklane-bin
├── stacklane-bin             # compiled Go binary (built by `make build`)
├── cmd/stacklane/            # cobra root + subcommand wiring
├── core/                     # config, project, state, lifecycle (operator semantics)
├── infra/                    # docker SDK, compose subprocess, gateway template
├── platform/                 # ports, dns, tls (host integrations)
├── observability/            # status, logs (read-only reporting)
├── internal/mocks/           # interface mocks for unit tests
├── docker-compose.yml        # per-project runtime template (with healthchecks; phpMyAdmin under `debug` profile)
├── docker-compose.shared.yml # shared gateway and network
├── docker/
│   └── nginx.conf.tmpl       # reference nginx template (Go renderer is authoritative)
├── .env.stacklane.example    # stack-wide defaults reference (copy to <stack-home>/.env.stacklane)
├── .env.example              # project-local override reference (copy into a project as .stacklane-local)
├── .stacklane-state/         # runtime state (git-ignored)
│   ├── projects/<slug>.json  # per-project state file
│   └── shared/               # generated gateway config
├── docs/
│   ├── architecture.md       # Go module ownership + contribution map
│   ├── contributing.md       # Go workflow, mocks, golden tests
│   ├── migration.md          # older workflow → Stacklane guide
│   ├── runtime-contract.md   # command semantics and state model
│   └── plan.md               # historical implementation plan
├── previous-version-archive/ # archived Bash implementation, kept for reference
└── README.md
```

## Shell Integration

Add this to `.zshrc` if you want the commands globally:

```bash
export STACK_HOME="${STACK_HOME:-$HOME/docker/stacklane}"
export PATH="$STACK_HOME:$PATH"

alias sl='stacklane'
alias sstatus='stacklane status'
alias sup='stacklane up'
alias sdown='stacklane down'
```

## Workflow Examples

Single project:

```bash
cd /path/to/project-a
stacklane up
stacklane status
stacklane down
```

Concurrent shared-gateway attachment:

```bash
cd /path/to/project-a
stacklane up

cd /path/to/project-b
stacklane attach --site-name project-b

stacklane status
stacklane status --project project-b
```

Global teardown:

```bash
stacklane down --all
```

## Troubleshooting

Check the resolved config without starting containers:

```bash
stacklane up --dry-run
```

Follow logs:

```bash
stacklane logs
stacklane logs apache
```

Reset a specific project by removing its state and volumes only after stopping it:

```bash
stacklane down
rm -f "$STACK_HOME/.stacklane-state/projects/<slug>.json"
docker volume ls
```

## Requirements

- macOS (Linux DNS bootstrap is a documented "unsupported platform" surface; lifecycle commands work on Linux but `dns-setup` does not)
- Docker Desktop (or Docker Engine ≥ Compose v2)
- Homebrew (only required for `dns-setup`)
- Go 1.26.2+ (only required to build from source; not required to run a downloaded binary)

## Project Status

The Bash implementation has been rewritten as a Go binary (spec [`003-rewrite-language-choices`](specs/003-rewrite-language-choices/spec.md)). The active runtime uses the current Stacklane contract: `stacklane <subcommand>`, `.stacklane-local`, and `.stacklane-state`.
