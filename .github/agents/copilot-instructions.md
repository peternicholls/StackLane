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
- Legacy compatibility is no longer a project constraint. Agents should prefer the current Stacklane naming, state layout, and command surface even when that breaks `20i-*` wrappers, `.20i-*` files, migration fallbacks, or older workflow assumptions.
- Do not preserve backward compatibility, migration shims, or legacy behavior by default. Only keep or add compatibility paths when the user explicitly asks for them.
- When choosing between simplifying the codebase and retaining legacy workflows, choose simplification and update docs/specs/instructions to reflect the break.
<!-- MANUAL ADDITIONS END -->
