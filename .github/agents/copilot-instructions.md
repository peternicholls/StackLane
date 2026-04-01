# Stacklane Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-01

## Active Technologies

- Bash on macOS with POSIX shell workflow, Docker Compose YAML + Docker Desktop, Docker Compose, Homebrew `dnsmasq`, Bash helper library in `lib/stacklane-common.sh`

## Project Structure

```text
stacklane          # canonical CLI entrypoint
20i-*              # deprecated compatibility wrappers (migration window only)
lib/
└── stacklane-common.sh   # shared runtime engine
docker/
docker-compose.yml
docker-compose.shared.yml
docs/
specs/
previous-version-archive/
```

## Commands

# Shell syntax validation: bash -n stacklane lib/stacklane-common.sh
# Entrypoint: ./stacklane --help | --up | --down | --attach | --detach | --status | --logs | --dns-setup

## Code Style

Bash on macOS with POSIX shell workflow, Docker Compose YAML: Follow standard conventions

## Recent Changes

- 002-project-rebrand: Rebranded CLI to Stacklane; `stacklane` is now the canonical entrypoint; `20i-*` scripts are deprecated wrappers; shared helper moved to `lib/stacklane-common.sh`

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
