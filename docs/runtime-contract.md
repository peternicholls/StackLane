# Runtime Contract

This document locks the implemented Phase 1 through Phase 6 command semantics and state model.

## Goals

- Preserve the existing shell-first workflow.
- Standardize project identity and future hostname derivation.
- Make `up`, `attach`, `down`, `detach`, and status behavior explicit.
- Provide a stable state model before the shared gateway and DNS work begins.

## Command Behavior

### `20i-up`

- Works from a project root.
- Resolves config from `.env`, `.20i-local`, environment variables, and CLI flags.
- Derives a project slug from the folder name by default.
- Plans a hostname of `<slug>.test` unless overridden.
- Ensures the shared gateway and shared Docker network exist.
- Starts the current Docker compose project.
- Captures the live runtime container identity.
- Writes a project state file, refreshes the stack registry, and marks the project as `attached`.
- Refreshes the shared gateway's hostname-aware route set from the registry.

### `20i-attach`

- Uses the same runtime resolution as `20i-up`.
- Is the explicit command for bringing an additional repo into the managed set.
- Bootstraps the shared layer if it is not already running.
- Starts a second isolated per-project runtime and refreshes the hostname-aware shared gateway rules.

### `20i-down`

- Stops only the current project runtime by default.
- Retains the project record and marks it `down`.
- Repoints the shared gateway to another attached project when one exists, otherwise leaves the gateway on a no-route response.
- Supports `20i-down --all` for global teardown.

### `20i-detach`

- Stops only the current project runtime.
- Removes the project record entirely.
- Repoints the shared gateway to another attached project when one exists.

### `20i-status`

- Reports tracked projects from the state directory.
- Supports `--project <selector>` where selector may be a project slug, project name, hostname, or repo path.
- Shows shared gateway health, local DNS health, planned hostname, hostname route URL, localhost probe URL, document root, container docroot, project path, registry file path, recorded live container identity, registry drift, and Docker status.

### `20i-logs`

- Follows logs for the current project runtime by default.
- Supports `--project <selector>` to target another recorded project by slug, project name, hostname, or repo path.

### `20i-dns-setup`

- Writes stack-managed `dnsmasq` config for the chosen suffix.
- Starts or restarts Homebrew `dnsmasq` on the configured local port.
- Installs or instructs the user to install the matching `/etc/resolver/<suffix>` file.
- Fails with a clear message when Homebrew is missing, `dnsmasq` is not installed, privileges are needed for `/etc/resolver`, or the resulting DNS health check is not ready.

## Config Precedence

1. CLI flags
2. `.20i-local`
3. Shell environment
4. Stack `.env`
5. Built-in defaults

## Hostname Derivation

- Default source: project folder name
- Override source: `SITE_NAME`
- Full override: `SITE_HOSTNAME`
- Default suffix: `.test`
- `.dev` remains deferred until the stack gains local HTTPS support

## Document Root Contract

- `DOCROOT` is the preferred project override.
- `CODE_DIR` remains as a legacy alias.
- Default behavior is `public_html` when present, otherwise the project root.
- The current 20i-style container layout mounts the whole repo at `/home/sites/<project-slug>`.
- The effective container docroot becomes `/home/sites/<project-slug>/<docroot-relative-path>`.
- For the current implementation, `DOCROOT` must resolve inside the project directory.

## Project State Model

- `attached`: runtime is intended to be active and recorded in state
- `down`: runtime has been stopped but the record is retained
- `detached`: record removed from active state storage
- global teardown: all records removed and all known runtimes stopped

## Runtime Identity Mapping

- Project slug defaults to the repo folder name, unless `SITE_NAME` overrides it.
- Compose project defaults to `20i-<slug>`.
- Runtime network is `<compose-project>-runtime`.
- Database volume is `<compose-project>-db-data`.
- Planned hostname defaults to `<slug>.test` unless overridden.
- State records keep the repo path, slug, compose project, hostname, and docroot together so live Docker resources can be mapped back to the originating repo.

## State Storage

- Location: `<stack-home>/.20i-state/projects/<slug>.env`
- One file per project
- Stores resolved runtime identity, container path layout, published per-project service ports, and recorded live container identity
- Stack registry snapshot: `<stack-home>/.20i-state/registry.tsv`
- Registry columns include repo path, project name, hostname, docroot, runtime settings, and recorded container summary for each project

## Shared Infrastructure

- Shared gateway compose file: `docker-compose.shared.yml`
- Shared network: `twentyi-shared` by default
- Shared gateway host ports: `80/443` by default, overrideable via `SHARED_GATEWAY_HTTP_PORT` and `SHARED_GATEWAY_HTTPS_PORT`
- Per-project web containers no longer publish host ports directly for normal site access
- phpMyAdmin remains per project in the current milestone and still publishes its own host port
- MariaDB remains per project and now resolves database name, user, password, and root password from the project-specific runtime config
- Gateway config is generated from the stack registry as one server block per attached hostname
- Invalid registry rows are skipped during route generation so one bad registration does not invalidate the full gateway config
- When a hostname is registered but its runtime is unavailable, the gateway returns a clear `503` response for that hostname

## Local DNS

- First implementation target: macOS
- Provider: Homebrew `dnsmasq`
- Listen address: `127.0.0.1`
- Listen port: `53535`
- Resolver file: `/etc/resolver/test` by default
- Status reports DNS readiness separately from Docker/gateway health
- Missing Homebrew, missing `dnsmasq`, missing resolver privileges, resolver mismatch, and stopped service are surfaced as explicit states

## Phase Boundary

This contract now includes registry-driven hostname-aware gateway rules plus the first macOS local DNS bootstrap path. The `.test` hostname is the routing target at the gateway layer, and `dnsmasq` plus `/etc/resolver` provide the resolution path for that suffix.