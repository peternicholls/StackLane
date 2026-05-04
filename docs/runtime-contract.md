# StageServe Runtime Contract

This document locks the StageServe command semantics and state model.

## Goals

- Preserve an easy shell-first workflow.
- Make repeated command runs reliable and predictable.
- Keep the stack robust under partial failure and stale state.
- Make `up`, `attach`, `down`, `detach`, and status behavior explicit.
- Keep the active command surface focused on `stage <subcommand>`.

## Command Behavior

### `stage up`

- Works from a project root.
- Resolves StageServe config from project-root `.env.stageserve`, shell environment, stack-wide `.env.stageserve`, and CLI flags while leaving project `.env` application-owned.
- Requires `STAGESERVE_STACK` to resolve to a supported stack kind. Today that means `20i`.
- If project-root `.env.stageserve` is missing, assumes defaults for the current run and writes a starter project file for later edits.
- Derives a project slug from the folder name by default.
- Plans a hostname of `<slug>.test` unless overridden.
- Ensures shared routing is available, reuses it when already healthy, and repairs it when missing.
- Starts the current Docker compose project.
- Captures the live runtime container identity.
- Writes a project state file, refreshes the stack registry, and marks the project as `attached`.
- Refreshes the shared hostname-aware route set from the registry.

### `stage attach`

- Uses the same runtime resolution as `stage up`.
- If project-root `.env.stageserve` is missing, assumes defaults for the current run and writes a starter project file for later edits.
- Is the explicit command for bringing an additional repo into the managed set.
- Bootstraps the shared routing layer if it is not already running.
- Starts a second isolated per-project runtime and refreshes the hostname-aware shared routing rules.

### `stage down`

- Stops only the current project runtime by default.
- Retains the project record and marks it `down`.
- Repoints the shared gateway to another attached project when one exists, otherwise leaves the gateway on a no-route response.
- Supports `stage down --all` for global teardown.

### `stage detach`

- Stops only the current project runtime.
- Removes the project record entirely.
- Repoints the shared gateway to another attached project when one exists.

### `stage status`

- Reports tracked projects from the state directory.
- Supports `--project <selector>` where selector may be a project slug, project name, hostname, or repo path.
- Shows shared routing health, local DNS health, planned hostname, hostname route URL, localhost probe URL, document root, container docroot, project path, registry file path, recorded live container identity, registry drift, and Docker status.

### `stage logs`

- Follows logs for the current project runtime by default.
- Supports `--project <selector>` to target another recorded project by slug, project name, hostname, or repo path.

### `stage dns-setup`

- Writes stack-managed `dnsmasq` config for the chosen suffix.
- Starts or restarts Homebrew `dnsmasq` on the configured local port.
- Installs or instructs the user to install the matching `/etc/resolver/<suffix>` file.
- Operator-facing examples should prefer `--site-suffix develop` unless a test or migration case specifically needs another allowed suffix.
- Browser access should use a full URL such as `http://my-project.develop/`; some browser address bars treat a bare hostname as a search query.
- When `SITE_SUFFIX=dev`, generates a local wildcard TLS certificate via `mkcert` and configures the shared gateway HTTPS port (default `8443`).
- Fails with a clear message when Homebrew is missing, `dnsmasq` is not installed, `mkcert` is required but absent, privileges are needed for `/etc/resolver`, or the resulting DNS health check is not ready.

## Onboarding Command Contract

### `stage setup`

- Runs an ordered set of machine-readiness steps: Docker CLI, Docker daemon, state directory, port 80, port 443, local DNS resolver, and mkcert local CA.
- Each step returns a `StepResult` with `id`, `label`, `status` (ready | needs_action | error), `message`, and optional `remediation`.
- Exit codes: 0 = all ready, 1 = needs_action, 2 = error, 3 = unsupported-os (highest precedence).
- `--suffix` accepts: `develop`, `dev`, `test`, or empty (stack default). Invalid values are rejected with an error.
- `--recheck` forces a full check re-run even if the machine is already healthy.
- `--json` emits a `CommandResult` JSON envelope and suppresses interactive output.
- `--no-tui` forces plain-text output. `--tui` forces TUI mode.
- `--non-interactive` suppresses prompts; implies `--no-tui`.
- Does not mutate machine state on its own — flags state that is not ready and provides `remediation` strings.

### `stage doctor`

- Read-only diagnostics reusing the same readiness steps as `setup`.
- Runs Docker binary, Docker daemon, state directory, port 80/443, DNS, and mkcert checks.
- Does not attempt repairs or prompt for elevated privileges.
- Exit codes follow the same 0-1-2-3 convention as `setup`.
- `--json`, `--no-tui`, `--non-interactive` flags behave the same as `setup`.

### `stage init`

- Writes a starter `.env.stageserve` in the project root (or `--project-dir`).
- Validates that the project root exists.
- Validates `--docroot` is inside the project root when supplied.
- Without `--force`, skips writing if `.env.stageserve` already exists.
- With `--force`, overwrites the existing file.
- `--site-name` sets `STAGESERVE_SITE_NAME` in the generated file.
- `--json`, `--no-tui`, `--non-interactive` flags behave the same as `setup`.
- Emits a `CommandResult` envelope (step `init.env_file`) with action in the message.

### Output modes

All three onboarding commands resolve output mode using the same precedence:

1. `--json` → JSON envelope (exit code still conveys overall status)
2. `--no-tui` → plain text
3. `--tui` → forced TUI
4. `--non-interactive` → plain text
5. TTY auto-detect → TUI on TTY, text otherwise

### CommandResult envelope (JSON)

```json
{
  "overall_status": "ready|needs_action|error",
  "exit_code": 0,
  "steps": [
    {
      "id": "docker.binary",
      "label": "Docker CLI",
      "status": "ready",
      "message": "docker found at /usr/local/bin/docker",
      "remediation": null,
      "code": "",
      "meta": null
    }
  ],
  "result": null,
  "next_steps": []
}
```

## Removed Wrappers

Root-level `20i-*` wrappers are not part of the active runtime. Use `stage <subcommand>`.

## Config Precedence

1. CLI flags
2. Project `.env.stageserve`
3. Shell environment
4. Stack `.env.stageserve`
5. Built-in defaults

Location defines ownership for the shared filename:

- `<project>/.env.stageserve` is the user-editable project override surface.
- `<stack-home>/.env.stageserve` is the stack-owned shared baseline.
- `<stack-home>/.stageserve-state/envfiles/*.env` is machine-generated runtime material and must not be edited.

Shared gateway settings are runtime-owned and no longer part of the supported env contract.

`STAGESERVE_STACK` is the explicit stack-kind selector in the env contract. The current runtime supports `20i` only; future values such as `laravel` or `node` must not be accepted until their stack implementations exist.

Selecting a different installed StageServe copy remains outside the project file for now. Use `STACK_HOME` or `--stack-home` until a dedicated project install/setup flow exists.

## Hostname Derivation

- Default source: project folder name
- Override source: `SITE_NAME`
- Full override: `SITE_HOSTNAME`
- Default suffix: `.test`
- `.dev` uses HTTPS on port `8443` by default and requires `mkcert` (`brew install mkcert && mkcert -install`); a local wildcard TLS cert is generated automatically by `stage dns-setup`

## Document Root Contract

- `DOCROOT` is the preferred project override.
- `CODE_DIR` remains an alias for `DOCROOT`.
- Default behavior is `public_html` when present, otherwise the project root.
- The current 20i-style container layout mounts the whole repo at `/home/sites/<project-slug>`.
- The effective container docroot becomes `/home/sites/<project-slug>/<docroot-relative-path>`.
- For the current implementation, `DOCROOT` must resolve inside the project directory.

## Project State Model

- `attached`: runtime is intended to be active and recorded in state
- `down`: runtime has been stopped but the record is retained
- `detached`: record removed from active state storage
- global teardown: all records removed and all known runtimes stopped

## Error Model

Every operator-facing failure from a lifecycle command is a typed `StepError` containing the failing step name, the affected project slug, the underlying cause, and a remediation hint. The CLI renders these as:

```
step <name> failed for project <slug>: <cause>
  next: <remediation>
```

The lifecycle layer rolls back partial progress before returning. A failed `stage up` never leaves a half-attached project: the state file is not written and any compose-up that completed is reversed.

## Runtime Identity Mapping

- Project slug defaults to the repo folder name, unless `SITE_NAME` overrides it.
- Compose project defaults to `stage-<slug>`.
- Runtime network is `<compose-project>-runtime`.
- Database volume is `<compose-project>-db-data`.
- Web alias on the shared network is `<compose-project>-web`.
- Planned hostname defaults to `<slug>.test` unless overridden.
- State records keep the repo path, slug, compose project, hostname, and docroot together so live Docker resources can be mapped back to the originating repo.

## State Storage

- Location: `<stack-home>/.stageserve-state/projects/<slug>.json`
- One JSON file per project, schema-versioned.
- Stores resolved runtime identity, container path layout, published per-project service ports, and recorded live container identity.
- Atomic writes (temp file + `os.Rename`) so a crash mid-write never leaves a half-formed file.
- The stack registry is computed in-memory from the JSON files; there is no positional `registry.tsv`.
- Obsolete `.20i-state` and Bash `.env` state files are ignored by default.

## Shared Infrastructure

- Shared routing compose file: `docker-compose.shared.yml`
- Active 20i project compose file: `docker-compose.20i.yml`
- StageServe keeps one shared routing layer available across attached projects and repairs it when the layer is missing or unhealthy.
- Shared gateway host ports: `80/443` by default; `.dev` runtime resolution moves HTTPS to `8443` when needed
- Per-project web containers no longer publish host ports directly for normal site access
- phpMyAdmin runs only when the `debug` compose profile is enabled (`stage up --profile debug`); it still publishes its own host port when active
- MariaDB remains per project and resolves database name, user, password, and root password from the project-specific runtime config
- All long-running services (nginx, apache/PHP-FPM, MariaDB) declare Docker `HEALTHCHECK` directives. `stage up` waits until they report healthy before returning success; the default deadline is 120 s, overrideable via `--wait-timeout` or `STAGESERVE_WAIT_TIMEOUT`. On timeout the error names the unhealthy services.
- Gateway config is generated from the in-memory registry as one server block per attached hostname (rendered by `infra/gateway` via `text/template`; output is golden-tested)
- Invalid registry rows are skipped during route generation so one bad registration does not invalidate the full gateway config
- When a hostname is registered but its runtime is unavailable, the gateway returns a clear `503` response for that hostname

Exact shared resource names remain part of the lower-level workflow contract and troubleshooting material; ordinary operator workflows should treat the shared routing layer as StageServe-managed.

## Local DNS

- First implementation target: macOS
- Provider: Homebrew `dnsmasq`
- Listen address: `127.0.0.1`
- Listen port: `53535`
- Resolver file: `/etc/resolver/test` by default
- Status reports DNS readiness separately from Docker/gateway health
- Missing Homebrew, missing `dnsmasq`, missing resolver privileges, resolver mismatch, and stopped service are surfaced as explicit named codes (`brew-missing`, `dnsmasq-missing`, `resolver-missing`, `resolver-mismatch`, `dnsmasq-stopped`, `config-missing`)
- On Linux, `stage dns-setup` returns the named `unsupported-os` code rather than silently no-op’ing

## Port Allocation

- The first project on a fresh stack gets the canonical pair `MYSQL_PORT=3306` / `PMA_PORT=8081`.
- Subsequent projects scan upward (3307+, 8082+) avoiding ports already reserved by another slug in the registry.
- Bind-checks are performed before any compose action runs.
- Concurrent `stage up` invocations are serialised via an exclusive `flock` on `<state-dir>/.port-allocation.lock`, so two parallel runs never claim the same port.
- An explicit per-project port (`MYSQL_PORT=`/`PMA_PORT=` or `--mysql-port`/`--pma-port`) that conflicts with the registry fails fast with `step allocate-ports failed ... next: free the conflicting port or pass --mysql-port / --pma-port`.

## Phase Boundary

This contract now includes registry-driven hostname-aware gateway rules plus the first macOS local DNS bootstrap path. The `.test` hostname is the routing target at the gateway layer, and `dnsmasq` plus `/etc/resolver` provide the resolution path for that suffix.
