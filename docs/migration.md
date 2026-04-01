# Migration Guide: Single-Project to Multi-Project Stack

This guide covers the transition from the original localhost-centric single-project workflow to the shared-gateway multi-project model introduced across Phases 1–7.

## What Changed

The original stack was built around a simple model: one project at a time, accessed through `localhost`. The shared-gateway model introduces persistent infrastructure that lives outside any one project, enabling multiple projects to run concurrently at stable local hostnames without port juggling.

| Concept | Old model | New model |
|---|---|---|
| Access URL | `http://localhost` or `http://localhost:8080` | `http://project-name.test` |
| Multiple projects | Stop one, start another | Attach both concurrently |
| Gateway ownership | Per-project compose file published to host | Shared `docker-compose.shared.yml` |
| DNS | `/etc/hosts` edit or none | `dnsmasq` + `/etc/resolver/test` (one-time) |
| Project identity | `COMPOSE_PROJECT_NAME` only | Slug, hostname, state file, registry |
| State tracking | None (Docker state only) | `.20i-state/` per-project env files + registry snapshot |

## Old Workflow

```bash
# Basic docker compose
cd /path/to/project
docker compose -p myproject up -d
# visit http://localhost
docker compose -p myproject down

# Or using older 20i scripts
cd /path/to/project
20i-up   # started containers, published web port to localhost
20i-down # stopped containers
```

Switching to a second project meant stopping the first.

## New Workflow

### One-time per-machine setup

```bash
20i-dns-setup
```

This installs Homebrew `dnsmasq`, configures it to resolve `*.test` to `127.0.0.1`, and writes `/etc/resolver/test`. The last step may require `sudo`; the command prints the exact copy-paste if so. Run once — it persists across reboots.

If you switch to `.dev`, the same command also generates a local wildcard TLS certificate. Local `.dev` uses HTTPS on port `8443` by default, which avoids common `443` conflicts while keeping the URL stable.

### Normal per-project usage

```bash
cd /path/to/project
20i-up            # shared gateway starts if not running; project registers at project-name.test
# visit http://project-name.test
20i-status        # shows hostname, container health, gateway and DNS state
20i-down          # stops the project runtime; retains its state record
```

The muscle memory for `20i-up` and `20i-down` is unchanged. The difference is what you type in the browser.

### Running two projects concurrently

```bash
cd /path/to/project-a
20i-up            # http://project-a.test is live

cd /path/to/project-b
20i-attach        # http://project-b.test starts alongside project-a.test

20i-status        # both projects shown
```

`20i-attach` is the explicit multi-project command. `20i-up` in a new repo also attaches if the shared layer is already running.

### Teardown

```bash
# Stop one project and retain its state record
20i-down

# Stop one project and remove its state record entirely
20i-detach

# Stop all projects and clear all state
20i-down --all
```

## Command Mapping

| Old command | New equivalent | Notes |
|---|---|---|
| `docker compose up -d` | `20i-up` | Also starts shared gateway, registers hostname |
| `docker compose down` | `20i-down` | Project-scoped; retains record by default |
| `docker compose logs -f` | `20i-logs` | Project-aware; use `--project` to switch scope |
| `docker compose ps` | `20i-status` | Now shows gateway, DNS, hostnames, and drift |
| Stop A, then start B | `20i-attach` from project B | Both run concurrently |
| _(no equivalent)_ | `20i-detach` | Removes project record and routing |
| _(no equivalent)_ | `20i-down --all` | Global teardown |
| _(no equivalent)_ | `20i-dns-setup` | One-time local DNS bootstrap |

## Config Migration

Per-project `.env` files from the old workflow are still read. The canonical per-project config file is now `.20i-local`:

```bash
# Old: .env in project root (still resolved but not preferred)
COMPOSE_PROJECT_NAME=mysite
HOST_PORT=8080
MYSQL_DATABASE=mysite_db

# New: .20i-local (preferred)
export SITE_NAME=mysite
export DOCROOT=public_html
export PHP_VERSION=8.4
export MYSQL_DATABASE=mysite_db
```

`HOST_PORT` is still resolved and honoured, but the new model routes all web traffic through the shared gateway rather than publishing a direct host port. Setting a specific `HOST_PORT` is rarely needed unless you want the phpMyAdmin or database ports to land on a particular number.

## Hostname Strategy

The hostname defaults to the project folder name. For a repo at `/path/to/my-project` the hostname is `my-project.test`.

Override in `.20i-local`:

```bash
export SITE_NAME=brand-name      # becomes brand-name.test
export SITE_HOSTNAME=exact.test  # full override, no suffix appended
```

## What Stays the Same

- `20i-up` and `20i-down` work exactly as before for single-project use.
- Config resolution order (CLI flags → `.20i-local` → environment → `.env` → defaults) is unchanged.
- PHP version, database credentials, and document root overrides all work as before.
- The `public_html` default document root fallback is unchanged.
- The GUI layer (`20i Stack Manager.app` and the Services menu workflow) still starts and stops a project. It does not yet expose attach, detach, or per-project hostname reporting — see [GUI-HELP.md](../GUI-HELP.md).

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

- **`.dev` support**: requires local TLS; deferred until the stack gains HTTPS support.
- **Full GUI parity**: the GUI trails the CLI. See [GUI-HELP.md](../GUI-HELP.md).
- **Windows / Linux DNS**: the `dnsmasq` bootstrap is macOS-only for now.

## Clean-Machine Bootstrap

Requirements: macOS, Docker Desktop, Homebrew.

```bash
# 1. Clone or copy the stack
git clone https://github.com/peternicholls/20i-stack ~/docker/20i-stack

# 2. Add stack commands to your path — add to ~/.zshrc and reload
echo 'export STACK_HOME="$HOME/docker/20i-stack"' >> ~/.zshrc
echo 'export PATH="$STACK_HOME:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 3. Bootstrap local DNS (once per machine)
20i-dns-setup

# 4. Start your first project
cd /path/to/your-project
20i-up
```

If `20i-dns-setup` requires elevated privileges it prints the exact `sudo` command to run to finish the resolver file installation. After that single manual step everything resolves automatically.
