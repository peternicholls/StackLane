# Migration Guide: Older Workflow to Stacklane

This guide covers the transition from the original localhost-centric single-project workflow to Stacklane and the shared-gateway multi-project model.

## What Changed

The original stack was built around a simple model: one project at a time, accessed through `localhost`. The shared-gateway model introduces persistent infrastructure that lives outside any one project, enabling multiple projects to run concurrently at stable local hostnames without port juggling.

| Concept | Old model | New model |
|---|---|---|
| Access URL | `http://localhost` or `http://localhost:8080` | `http://project-name.test` |
| Multiple projects | Stop one, start another | Attach both concurrently |
| Gateway ownership | Per-project compose file published to host | Shared `docker-compose.shared.yml` |
| DNS | `/etc/hosts` edit or none | `dnsmasq` + `/etc/resolver/test` (one-time) |
| Project identity | `COMPOSE_PROJECT_NAME` only | Slug, hostname, state file, registry |
| State tracking | None (Docker state only) | `.stacklane-state/` JSON project records |

## Old Workflow

```bash
# Basic docker compose
cd /path/to/project
docker compose -p myproject up -d
# visit http://localhost
docker compose -p myproject down

# Or using Stacklane
cd /path/to/project
stacklane up
stacklane down
```

Switching to a second project meant stopping the first.

## New Workflow

### One-time per-machine setup

```bash
stacklane dns-setup
```

This installs Homebrew `dnsmasq`, configures it to resolve `*.test` to `127.0.0.1`, and writes `/etc/resolver/test`. The last step may require `sudo`; the command prints the exact copy-paste if so. Run once — it persists across reboots.

If you switch to `.dev`, the same command also generates a local wildcard TLS certificate. Local `.dev` uses HTTPS on port `8443` by default, which avoids common `443` conflicts while keeping the URL stable.

### Normal per-project usage

```bash
cd /path/to/project
stacklane up            # shared gateway starts if not running; project registers at project-name.test
# visit http://project-name.test
stacklane status        # shows hostname, container health, gateway and DNS state
stacklane down          # stops the project runtime; retains its state record
```

Stacklane's active command surface is `stacklane <subcommand>`. Root-level `20i-*` wrappers are no longer part of the runtime.

### Running two projects concurrently

```bash
cd /path/to/project-a
stacklane up            # http://project-a.test is live

cd /path/to/project-b
stacklane attach        # http://project-b.test starts alongside project-a.test

stacklane status        # both projects shown
```

`stacklane attach` is the explicit multi-project command. `stacklane up` in a new repo also attaches if the shared layer is already running.

### Teardown

```bash
# Stop one project and retain its state record
stacklane down

# Stop one project and remove its state record entirely
stacklane detach

# Stop all projects and clear all state
stacklane down --all
```

## Command Mapping

| Old command | Stacklane equivalent | Notes |
|---|---|---|
| `docker compose up -d` | `stacklane up` | Also starts shared gateway, registers hostname |
| `docker compose down` | `stacklane down` | Project-scoped; retains record by default |
| `docker compose logs -f` | `stacklane logs` | Project-aware; use `--project` to switch scope |
| `docker compose ps` | `stacklane status` | Now shows gateway, DNS, hostnames, and drift |
| Stop A, then start B | `stacklane attach` from project B | Both run concurrently |
| _(no equivalent)_ | `stacklane detach` | Removes project record and routing |
| _(no equivalent)_ | `stacklane down --all` | Global teardown |
| _(no equivalent)_ | `stacklane dns-setup` | One-time local DNS bootstrap |

Older `20i-*` wrapper names are intentionally not retained at the repository root.

## Config Migration

The canonical per-project config file is now `.env.stacklane`:

```bash
# .env.stacklane
export SITE_NAME=mysite
export DOCROOT=public_html
export PHP_VERSION=8.4
export MYSQL_DATABASE=mysite_db
```

`HOST_PORT` is still resolved and honoured, but the new model routes all web traffic through the shared gateway rather than publishing a direct host port. Setting a specific `HOST_PORT` is rarely needed unless you want the phpMyAdmin or database ports to land on a particular number.

## Hostname Strategy

The hostname defaults to the project folder name. For a repo at `/path/to/my-project` the hostname is `my-project.test`.

Override in project `.env.stacklane`:

```bash
export SITE_NAME=brand-name      # becomes brand-name.test
export SITE_HOSTNAME=exact.test  # full override, no suffix appended
```

## What Stays the Same

- Runtime behavior, config precedence, and state isolation are defined under `stacklane`.
- Config resolution order is CLI flags → project `.env.stacklane` → environment → stack-home `.env.stacklane` → defaults.
- PHP version, database credentials, and document root overrides all work as before.
- The `public_html` default document root fallback is unchanged.
- The GUI layer (`Stacklane Manager.app` and the Services menu workflow) still starts and stops a project. It does not yet expose attach, detach, or per-project hostname reporting. GUI assets and documentation have been moved to `previous-version-archive/GUI-HELP.md`.

## Repository Rename vs Local Folder Rename

- A GitHub repository rename and your local checkout folder name are separate changes.
- The GitHub repository is now named `StackLane`, but existing local clones do not rename themselves.
- If you want your local stack folder to be named `stacklane`, rename that directory manually and then update `STACK_HOME`, shell aliases, launchers, and any deployment copy that still points at an older `~/docker/20i-stack` path.

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
# Note: the GitHub repository is named StackLane. The recommended local folder is ~/docker/stacklane.
# Existing checkouts under ~/docker/20i-stack continue to work until you choose to rename them.
git clone https://github.com/peternicholls/StackLane.git ~/docker/stacklane

# 2. Add stack commands to your path — add to ~/.zshrc and reload
echo 'export STACK_HOME="$HOME/docker/stacklane"' >> ~/.zshrc
echo 'export PATH="$STACK_HOME:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 3. Bootstrap local DNS (once per machine)
stacklane dns-setup

# 4. Start your first project
cd /path/to/your-project
stacklane up
```

If `stacklane dns-setup` requires elevated privileges it prints the exact `sudo` command to run to finish the resolver file installation. After that single manual step everything resolves automatically.

---

## Bash Archive And Go Runtime (Spec 003)

The 003 rewrite replaces the active `lib/stacklane-common.sh` path with a single Go binary at
the same `stacklane` entrypoint. The archived Bash implementation lives under
`previous-version-archive/` for reference only.

### Output format

| Surface | Archived Bash | Go | Notes |
|---|---|---|---|
| `up` success | Styled `▶ up <slug>` line | Plain `up <slug> ok` | Easier to parse from scripts. |
| Status table | `printf "%-20s"` aligned columns | Same column ordering, alignment may shift on long names | Driven by `observability/status.Render`; not parsed by tests. |
| Errors | Free-form `echo "ERROR: ..."` | Single `step <name> failed for project <slug>: <cause>\n  next: <action>` block | All lifecycle errors flow through `lifecycle.StepError`. |

### Behaviour deltas

- **Wait timeout**: Go defaults to 120s and honours `--wait-timeout` / `STACKLANE_WAIT_TIMEOUT` (FR-009).
- **DNS bootstrap on Linux**: Go returns `unsupported-os` from `stacklane dns-setup` (FR-012).
- **State files**: active state lives in `.stacklane-state/projects/<slug>.json`; obsolete `.20i-state` and Bash `.env` state files are ignored by default.
- **Port allocation**: explicit per-project conflicts now fail BEFORE any docker action runs, with `step allocate-ports failed ... next: free the conflicting port or pass --mysql-port / --pma-port`.
- **phpMyAdmin**: planned move behind the `debug` compose profile (T044). Until that lands, behaviour is unchanged.

### Things that are exactly the same

- Project slug derivation (`stacklane_slugify`).
- Hostname resolution rules (explicit > `<slug>.<suffix>`).
- Document root canonicalisation (explicit `DOCROOT` > `CODE_DIR` alias > `public_html` > project root).
- Compose project name (`stln-<slug>`, renamed from `stacklane-<slug>` in spec 004), web alias (`stln-<slug>-web`), runtime network, database volume.
- nginx gateway block layout (golden-tested under `infra/gateway/testdata/`).

### Deferred for follow-up

The following items are carried forward in `specs/003-rewrite-language-choices/tasks.md` as deferred:

- Docker-gated integration tests (T039/T061), release pipeline + signing + install scripts (T051–T054), and final runtime validation (T069–T070).

These require live Docker or signed-release infrastructure.
