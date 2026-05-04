# Migration Guide: Older Workflow to StageServe

This guide covers the transition from the original localhost-centric single-project workflow to StageServe and the shared-gateway multi-project model.

## What Changed

The original stack was built around a simple model: one project at a time, accessed through `localhost`. The shared-gateway model introduces persistent infrastructure that lives outside any one project, enabling multiple projects to run concurrently at stable local hostnames without port juggling.

| Concept | Old model | New model |
|---|---|---|
| Access URL | `http://localhost` or `http://localhost:8080` | `http://project-name.develop` |
| Multiple projects | Stop one, start another | Attach both concurrently |
| Gateway ownership | Per-project compose file published to host | Shared `docker-compose.shared.yml` |
| DNS | `/etc/hosts` edit or none | `dnsmasq` + `/etc/resolver/develop` (one-time) |
| Project identity | `COMPOSE_PROJECT_NAME` only | Slug, hostname, state file, registry |
| State tracking | None (Docker state only) | `.stageserve-state/` JSON project records |

## Old Workflow

```bash
# Basic docker compose
cd /path/to/project
docker compose -p myproject up -d
# visit http://localhost
docker compose -p myproject down

# Or using StageServe
cd /path/to/project
stage up
stage down
```

Switching to a second project meant stopping the first.

## New Workflow

### One-time per-machine setup

```bash
stage dns-setup --site-suffix develop
```

This installs Homebrew `dnsmasq`, configures it to resolve `*.develop` to `127.0.0.1`, and writes `/etc/resolver/develop`. The last step may require `sudo`; the command prints the exact copy-paste if so. Run once — it persists across reboots.

When typing the route into Safari or VS Code Simple Browser, include the scheme: `http://project-name.develop/`. A bare hostname may be treated as a search string rather than a URL.

If you switch to `.dev`, the same command also generates a local wildcard TLS certificate. Local `.dev` uses HTTPS on port `8443` by default, which avoids common `443` conflicts while keeping the URL stable.

### Normal per-project usage

```bash
cd /path/to/project
stage up --site-suffix develop   # shared gateway starts if not running; project registers at project-name.develop
# visit http://project-name.develop
stage status        # shows hostname, container health, gateway and DNS state
stage down          # stops the project runtime; retains its state record
```

StageServe's active command surface is `stage <subcommand>`. Root-level `20i-*` wrappers are no longer part of the runtime.

### Running two projects concurrently

```bash
cd /path/to/project-a
stage up --site-suffix develop      # http://project-a.develop is live

cd /path/to/project-b
stage attach --site-suffix develop  # http://project-b.develop starts alongside project-a.develop

stage status        # both projects shown
```

`stage attach` is the explicit multi-project command. `stage up` in a new repo also attaches if the shared layer is already running.

### Teardown

```bash
# Stop one project and retain its state record
stage down

# Stop one project and remove its state record entirely
stage detach

# Stop all projects and clear all state
stage down --all
```

## Command Mapping

| Old command | StageServe equivalent | Notes |
|---|---|---|
| `docker compose up -d` | `stage up` | Also starts shared gateway, registers hostname |
| `docker compose down` | `stage down` | Project-scoped; retains record by default |
| `docker compose logs -f` | `stage logs` | Project-aware; use `--project` to switch scope |
| `docker compose ps` | `stage status` | Now shows gateway, DNS, hostnames, and drift |
| Stop A, then start B | `stage attach` from project B | Both run concurrently |
| _(no equivalent)_ | `stage detach` | Removes project record and routing |
| _(no equivalent)_ | `stage down --all` | Global teardown |
| _(no equivalent)_ | `stage dns-setup` | One-time local DNS bootstrap |

Older `20i-*` wrapper names are intentionally not retained at the repository root.

## Config Migration

The canonical per-project config file is now `.env.stageserve`:

```bash
# .env.stageserve
export SITE_NAME=mysite
export DOCROOT=public_html
export PHP_VERSION=8.4
export MYSQL_DATABASE=mysite_db
```

`HOST_PORT` is still resolved and honoured, but the new model routes all web traffic through the shared gateway rather than publishing a direct host port. Setting a specific `HOST_PORT` is rarely needed unless you want the phpMyAdmin or database ports to land on a particular number.

## Hostname Strategy

The hostname defaults to the project folder name. For a repo at `/path/to/my-project` the hostname is `my-project.test`.

Override in project `.env.stageserve`:

```bash
export SITE_NAME=brand-name      # becomes brand-name.test
export SITE_HOSTNAME=exact.test  # full override, no suffix appended
```

## What Stays the Same

- Runtime behavior, config precedence, and state isolation are defined under `stage`.
- Config resolution order is CLI flags → project `.env.stageserve` → environment → stack-home `.env.stageserve` → defaults.
- PHP version, database credentials, and document root overrides all work as before.
- The `public_html` default document root fallback is unchanged.
- The GUI layer (`StageServe Manager.app` and the Services menu workflow) still starts and stops a project. It does not yet expose attach, detach, or per-project hostname reporting. GUI assets and documentation have been moved to `previous-version-archive/GUI-HELP.md`.

## Repository Rename vs Local Folder Rename

- A GitHub repository rename and your local checkout folder name are separate changes.
- The GitHub repository is now named `StageServe`, but existing local clones do not rename themselves.
- If you want your local stack folder to be named `stage`, rename that directory manually and then update `STACK_HOME`, shell aliases, launchers, and any deployment copy that still points at an older `~/docker/20i-stack` path.

## Wordfence / `.user.ini` Note

WordPress sites cloned from live often have a hardcoded host path in `public_html/.user.ini`:

```
auto_prepend_file = '/absolute/host/path/wordfence-waf.php'
```

Change this to the container path before starting:

```
auto_prepend_file = '/var/www/site/wordfence-waf.php'
```

Then restart: `docker restart <project>-apache-1`

This was true in the old model and remains true in the new one.

## Known Deferred Items

- **Full GUI parity**: the GUI trails the CLI.
- **Windows / Linux DNS**: the `dnsmasq` bootstrap is macOS-only for now.

## Clean-Machine Bootstrap

Requirements: macOS, Docker Desktop, Homebrew.

```bash
# 1. Clone or copy the stack
# Note: the GitHub repository is named StageServe. The recommended local folder is ~/docker/stage.
# Existing checkouts under ~/docker/20i-stack continue to work until you choose to rename them.
git clone https://github.com/peternicholls/StageServe.git ~/docker/stage

# 2. Add stack commands to your path — add to ~/.zshrc and reload
echo 'export STACK_HOME="$HOME/docker/stage"' >> ~/.zshrc
echo 'export PATH="$STACK_HOME:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 3. Bootstrap local DNS (once per machine)
stage dns-setup

# 4. Start your first project
cd /path/to/your-project
stage up
```

If `stage dns-setup` requires elevated privileges it prints the exact `sudo` command to run to finish the resolver file installation. After that single manual step everything resolves automatically.

---

## Bash Archive And Go Runtime (Spec 003)

The 003 rewrite replaces the active `lib/stage-common.sh` path with a single Go binary at
the same `stage` entrypoint. The archived Bash implementation lives under
`previous-version-archive/` for reference only.

### Output format

| Surface | Archived Bash | Go | Notes |
|---|---|---|---|
| `up` success | Styled `▶ up <slug>` line | Plain `up <slug> ok` | Easier to parse from scripts. |
| Status table | `printf "%-20s"` aligned columns | Same column ordering, alignment may shift on long names | Driven by `observability/status.Render`; not parsed by tests. |
| Errors | Free-form `echo "ERROR: ..."` | Single `step <name> failed for project <slug>: <cause>\n  next: <action>` block | All lifecycle errors flow through `lifecycle.StepError`. |

### Behaviour deltas

- **Wait timeout**: Go defaults to 120s and honours `--wait-timeout` / `STAGESERVE_WAIT_TIMEOUT` (FR-009).
- **DNS bootstrap on Linux**: Go returns `unsupported-os` from `stage dns-setup` (FR-012).
- **State files**: active state lives in `.stageserve-state/projects/<slug>.json`; obsolete `.20i-state` and Bash `.env` state files are ignored by default.
- **Port allocation**: explicit per-project conflicts now fail BEFORE any docker action runs, with `step allocate-ports failed ... next: free the conflicting port or pass --mysql-port / --pma-port`.
- **phpMyAdmin**: planned move behind the `debug` compose profile (T044). Until that lands, behaviour is unchanged.

### Things that are exactly the same

- Project slug derivation (`stage_slugify`).
- Hostname resolution rules (explicit > `<slug>.<suffix>`).
- Document root canonicalisation (explicit `DOCROOT` > `CODE_DIR` alias > `public_html` > project root).
- Compose project name (`stage-<slug>`, renamed from `stage-<slug>` in spec 004), web alias (`stage-<slug>-web`), runtime network, database volume.
- nginx gateway block layout (golden-tested under `infra/gateway/testdata/`).

### Deferred for follow-up

The following items are carried forward in `specs/003-rewrite-language-choices/tasks.md` as deferred:

- Docker-gated integration tests (T039/T061), release pipeline + signing + install scripts (T051–T054), and final runtime validation (T069–T070).

These require live Docker or signed-release infrastructure.
