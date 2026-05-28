# StageServe Project Analysis Report

Date: 2026-05-22
Branch reviewed: `codex/007-harden-TUI-and-other-interactions`

## Executive Summary

StageServe is aiming to be a single-command local development runtime that emulates a 20i-style shared-hosting workflow on Docker. Its core promise is strong: a developer should be able to open a project folder, run StageServe, get a stable local hostname, isolated PHP/MariaDB runtime, shared gateway routing, and enough diagnostics to recover without learning Docker internals first.

The project has a solid Go foundation. The direct CLI, configuration loader, state store, gateway renderer, lifecycle orchestration, and onboarding readiness model are all real implementation, not just plans. The short Go test suite, `go vet`, and full package build all pass.

The project is not complete as a product experience yet. The direct-command runtime is broadly usable, but the stated "easy mode" goal is still mostly planned work. The largest gaps are guided bare `stage` behavior, docs/code drift around command flags and selectors, incomplete `.dev` TLS wiring, real-Docker validation coverage, and a few lifecycle edge cases where rollback/state/routing can drift after a late failure.

## Goals And Aims

StageServe's active docs and specs point to five product aims:

1. Make local Docker development feel like shared hosting instead of container management.
2. Keep one shared gateway and local DNS layer in front of isolated per-project runtimes.
3. Use `stage` as the canonical CLI entrypoint and avoid restoring legacy wrapper behavior.
4. Keep configuration ownership clear: project-local `.env.stageserve`, stack-home `.env.stageserve`, and generated `.stageserve-state` runtime files.
5. Provide a simple-first terminal experience where non-expert users can be guided through setup, project config, running, stopping, logs, and recovery.

The first four aims are substantially represented in code. The fifth aim is documented and prototyped, but not implemented in the active CLI.

## Completeness Snapshot

| Area | Status | Notes |
|---|---:|---|
| Go CLI direct commands | Mostly implemented | `up`, `attach`, `detach`, `down`, `status`, `logs`, `dns-setup`, `setup`, `doctor`, `init`, and `version` exist. Some documented flags/argument shapes are missing. |
| Config precedence and ownership | Strong | The loader follows CLI > project `.env.stageserve` > shell > stack `.env.stageserve` > defaults, with the project-only post-up hook restriction. |
| Multi-project runtime model | Mostly implemented | Per-project compose names, networks, DB volumes, generated env files, state records, and gateway routes exist. Real-daemon validation remains under-documented/incomplete. |
| Shared gateway rendering | Strong core, some wiring gaps | Template rendering is typed and golden-tested. `.dev` TLS and route reload behavior need attention. |
| Local DNS | macOS implemented, Linux explicit stub | `dns-setup` has a macOS Homebrew/dnsmasq implementation and non-darwin unsupported behavior. `setup`/`doctor` checks do not fully align with effective suffixes. |
| Onboarding commands | Partial | `setup`, `doctor`, and `init` exist with text/JSON/TUI projection. The TUI is a projection style, not the guided easy-mode flow from spec 007. |
| Bare `stage` easy mode | Not implemented | Running `go run ./cmd/stage` currently prints Cobra help, not a guided TUI or text fallback. |
| Documentation | Useful but drifting | The docs are rich, but some active docs describe behavior not present in code. Spec 009 correctly recognizes the need for decluttering. |
| Tests | Good unit coverage, weak integration coverage | `go test -short ./...` passes. There is no tagged Docker integration suite for the full lifecycle. |
| Release/install process | Partial | Installer and GitHub release workflow exist. The Makefile `release` target is still a placeholder. |

## What Is Working Well

### Layering

The architecture in [docs/architecture.md](architecture.md) is mostly reflected in the code:

- CLI command wiring lives in [cmd/stage/commands](../cmd/stage/commands).
- Configuration lives in [core/config](../core/config).
- Lifecycle sequencing lives in [core/lifecycle](../core/lifecycle).
- State storage lives in [core/state](../core/state).
- Compose subprocess behavior lives in [infra/compose](../infra/compose).
- Docker SDK behavior lives in [infra/docker](../infra/docker).
- Gateway rendering lives in [infra/gateway](../infra/gateway).
- Read-only reporting lives in [observability](../observability).

That separation gives the project a maintainable shape and makes focused tests practical.

### Configuration Contract

[core/config/loader.go](../core/config/loader.go) is one of the strongest parts of the implementation. It centralizes precedence, keeps stack defaults and project overrides location-based, keeps `.env` application-owned except for narrow DB fallbacks, and rejects unsupported `STAGESERVE_STACK` values.

### State And Gateway Safety Patterns

[core/state/store.go](../core/state/store.go) and [infra/gateway/manager.go](../infra/gateway/manager.go) use atomic temp-file-plus-rename writes. That is the right pattern for a CLI that may be interrupted and for generated files that the shared gateway depends on.

### Operator-Facing Errors

The lifecycle `StepError` model is a good product choice. It lets the runtime say which step failed, which project was affected, and what the operator can do next. That fits StageServe's goal of making failures recoverable without raw Docker spelunking.

### Test Baseline

The following checks passed during this review:

| Command | Result |
|---|---|
| `go test -short ./...` | Passed |
| `go vet ./...` | Passed |
| `go build ./...` | Passed |
| `go run ./cmd/stage` | Printed Cobra help, confirming current no-args behavior |
| `go run ./cmd/stage setup --help` | Showed older `--tui` / `--no-tui` flags |
| `go run ./cmd/stage init --help` | Showed older `--no-tui` flag and no guided init form contract |

No real-Docker lifecycle validation was run as part of this report.

## Major Gaps And Risks

### 1. The Easy-Mode Product Goal Is Still Planned, Not Implemented

[docs/concept.md](concept.md) and [specs/007-harden-TUI-and-other-interactions/spec.md](../specs/007-harden-TUI-and-other-interactions/spec.md) define bare `stage` as the guided entrypoint. The active root command in [cmd/stage/commands/root.go](../cmd/stage/commands/root.go) has no no-args `RunE`, so Cobra prints help.

This is the largest product completeness gap. StageServe's current product is a competent direct-command CLI; it is not yet the simple-first guided tool described by spec 007.

Recommended next step: implement the spec 007 planner first, then wire root no-args TTY/non-TTY behavior. Avoid putting guidance logic directly in Cobra or Bubble Tea models.

### 2. Active Docs Describe Commands That Do Not Exist Or Behave Differently

The docs currently overstate several command surfaces:

- [README.md](../README.md) shows `stage status --project project-b`, but [cmd/stage/commands/status.go](../cmd/stage/commands/status.go) only supports the current resolved project or `--all`.
- [README.md](../README.md) shows `stage logs apache`, but [cmd/stage/commands/logs.go](../cmd/stage/commands/logs.go) ignores positional args and expects `--service apache`.
- [docs/runtime-contract.md](runtime-contract.md) documents `setup --recheck`, but [cmd/stage/commands/setup.go](../cmd/stage/commands/setup.go) has no `--recheck` flag.
- [docs/runtime-contract.md](runtime-contract.md) says `status` and `logs` support `--project <selector>`, but the code does not implement that selector.
- [docs/runtime-contract.md](runtime-contract.md) says `doctor` includes shared gateway diagnostics; [cmd/stage/commands/doctor.go](../cmd/stage/commands/doctor.go) checks Docker, state dir, ports, DNS, and mkcert, but not gateway health.
- Spec 007 says final opt-outs are `--notui` and `--cli`, while current commands expose `--tui` and `--no-tui`.

These are not cosmetic differences. They affect onboarding trust: users will copy commands that fail or silently do the wrong thing.

Recommended next step: make a small contract reconciliation pass before more UI work. Either implement the documented flags/selectors or correct the docs to the implemented surface.

### 3. Setup And DNS Remediation Are Inconsistent

`stage setup` currently runs readiness checks. It does not perform DNS bootstrap. However, the DNS readiness check's remediation can point back to `stage setup`, creating a loop for users who need `stage dns-setup`.

There is also a product-language conflict:

- [README.md](../README.md) describes `stage setup` as machine-readiness checks plus one-time DNS/mkcert setup.
- [docs/runtime-contract.md](runtime-contract.md) says `setup` does not mutate machine state on its own.
- [cmd/stage/commands/setup.go](../cmd/stage/commands/setup.go) only checks and reports.

Recommended next step: decide whether `setup` is check-only or guided repair. If it remains check-only, update README/help/remediation to point users to `stage dns-setup --site-suffix <suffix>` and external mkcert steps.

### 4. `.dev` HTTPS Is Documented More Completely Than It Is Wired

The runtime contract says `.dev` generates local wildcard TLS certs and configures the shared gateway HTTPS port. The code has pieces, but the full path is not connected:

- [platform/tls/mkcert.go](../platform/tls/mkcert.go) exists, but no active lifecycle or `dns-setup` command calls it.
- [stacks/20i/docker-compose.shared.yml](../stacks/20i/docker-compose.shared.yml) can mount `SHARED_GATEWAY_CERTS_DIR`, but [core/lifecycle/orchestrator.go](../core/lifecycle/orchestrator.go) does not pass that env var.
- [infra/gateway/manager.go](../infra/gateway/manager.go) `AddRoute` calls `WriteConfig` without preserving `TLSEnabled` or `HTTPSPort`, so route additions can render non-TLS config even when `cfg.SiteSuffix == "dev"`.

Recommended next step: either downgrade `.dev` docs to partial support or complete TLS generation, cert-dir env wiring, and route rendering for `.dev` before presenting it as complete.

### 5. Late `stage up` Failures Can Leave Gateway State Behind

In [core/lifecycle/orchestrator.go](../core/lifecycle/orchestrator.go), `Up` adds and reloads the gateway route before saving project state. If the state save fails after the gateway reload, `rollbackProject` stops the compose project but does not remove the route or resync gateway config.

That breaks the documented promise that a failed `stage up` never leaves a half-attached project. It also matches unfinished lifecycle tasks in [specs/004-workflow-and-lifecycle/tasks.md](../specs/004-workflow-and-lifecycle/tasks.md).

Recommended next step: treat route add/reload plus state save as one rollback-aware section. On any failure after route add, remove the route and reload/sync the gateway before returning.

### 6. `--profile debug` Appears Exposed But Not Applied

The root command exposes `--profile`, and docs describe phpMyAdmin as opt-in via `stage up --profile debug`. [infra/compose/types.go](../infra/compose/types.go) supports profiles, but [core/lifecycle/orchestrator.go](../core/lifecycle/orchestrator.go) does not pass `flags.Profile` into `compose.UpOptions` for the project runtime.

Recommended next step: either wire profiles into `UpOptions` or remove the flag/docs until it is supported.

### 7. Spec Tracking Is Inconsistent

[docs/plan.md](plan.md) presents the multi-project runtime plan as mostly complete, with only a few validation checkboxes open. [specs/004-workflow-and-lifecycle/tasks.md](../specs/004-workflow-and-lifecycle/tasks.md) still has many unchecked implementation and validation tasks, some of which appear partially implemented and some of which remain real gaps. [specs/007-harden-TUI-and-other-interactions/tasks.md](../specs/007-harden-TUI-and-other-interactions/tasks.md) is entirely unchecked. [specs/009-documentation-update-and-declutter/spec.md](../specs/009-documentation-update-and-declutter/spec.md) is a placeholder.

Spec 007 and older onboarding task material also refer to an active `docs/installer-onboarding.md` style document, but that file is not present in the repository.

Recommended next step: reconcile task files with the current code before using them as delivery truth. Mark truly done work, preserve open gaps, and move superseded tasks to explicit "deferred" or "no longer applicable" sections.

## Architectural Issues

### SharedFlags Pollutes The CLI Surface

[cmd/stage/commands/root.go](../cmd/stage/commands/root.go) puts most runtime flags on the root command. As a result, commands like `stage setup --help` and `stage init --help` display MySQL, phpMyAdmin, host-port, profile, wait-timeout, and other flags that are irrelevant or ignored for those commands.

Impact: this directly works against the simple-first goal. The CLI teaches users implementation breadth before task intent.

Recommendation: split flags by command family. Keep only truly global concerns such as `--stack-home` and maybe `--project-dir` at root. Move lifecycle, DB, profile, dry-run, and selector flags to the commands that consume them.

### Orchestrator Owns Too Many Helpers

[core/lifecycle/orchestrator.go](../core/lifecycle/orchestrator.go) is doing correct high-level work, but it also owns generated env-file rendering, gateway port fallback scanning, container lookup, observed runtime summarization, post-up exec wiring, and route derivation.

Impact: the lifecycle flow is harder to audit, and helper ownership is blurrier than the architecture doc intends.

Recommendation: keep orchestration order in `core/lifecycle`, but move helpers to narrower owners:

- Runtime env materialization: `core/config` or a small lifecycle env writer.
- Gateway TLS/port materialization: `infra/gateway` or a gateway runtime helper.
- Container/service lookup helpers: `infra/docker` or a lifecycle collaborator interface.
- Registry-to-route projection: `core/state` or `infra/gateway`, depending on ownership preference.

### Onboarding Projection And Exit-Code Code Is Repeated

[cmd/stage/commands/setup.go](../cmd/stage/commands/setup.go), [cmd/stage/commands/doctor.go](../cmd/stage/commands/doctor.go), and [cmd/stage/commands/init.go](../cmd/stage/commands/init.go) repeat projection switching and command-specific silent exit error types.

Recommendation: add a shared `projectOnboardingResult` helper and one generic `exitCodeError`. This will also make spec 007's output-mode cleanup easier.

### Project Env Rendering Is Split Across Two Paths

There are two user-facing `.env.stageserve` render paths:

- `stage init` uses [core/onboarding/project_env.go](../core/onboarding/project_env.go).
- First `stage up` / `stage attach` uses [cmd/stage/commands/project_env.go](../cmd/stage/commands/project_env.go).

They intentionally differ today, but they encode overlapping rules and quoting helpers. As guided init grows, this split will become a drift risk.

Recommendation: create one project-env planning/rendering package that can produce both the minimal guided init body and the richer first-run body from a shared model.

### Compose Detach Option Is Dead Code

[infra/compose/compose.go](../infra/compose/compose.go) appends `-d` in both branches of the `Detach` conditional. This is small, but it signals unclear API intent.

Recommendation: either remove `Detach` from `UpOptions` and document that orchestration always detaches, or honor the option.

## Cleaner Code Opportunities

1. Add a `core/guidance` package before writing any root TUI behavior, as spec 007 proposes.
2. Consolidate onboarding output rendering and silent exit-code errors.
3. Move route add/remove into lifecycle-owned rollback sections rather than gateway convenience methods that lose TLS context.
4. Replace root-level shared flags with command-local flags.
5. Give `status` and `logs` shared project selector resolution if the documented `--project` contract remains desired.
6. Wire `--profile` through lifecycle or remove it until supported.
7. Use one `.env.stageserve` rendering model for `init`, first `up`, and future guided previews.
8. Consider a read/write mutex in [core/state/store.go](../core/state/store.go) only if status operations become frequent; it is low priority for a CLI.

## More Efficient Process Opportunities

### Add A Tagged Docker Integration Suite

The current unit suite is healthy, but StageServe's highest-risk behavior involves Docker, compose, ports, DNS, gateway config, and filesystem state. Add `//go:build integration` tests or scripted smoke tests for:

- one project `up` / `status` / `down`
- two projects concurrently attached
- `detach` preserving the other project
- `down --all`
- failing `STAGESERVE_POST_UP_COMMAND` rollback
- `.dev` / TLS behavior, or an explicit skipped test until supported
- profile activation for phpMyAdmin

The integration suite does not need to run on every unit-test loop. It should run manually before release and in CI where Docker is available.

### Reconcile Specs Before Implementation

Before continuing spec 007, do a short tracking pass:

- Mark implemented spec 004/005 work as complete.
- Move stale or contradicted tasks into a deferred/obsolete section.
- Convert currently discovered drift into concrete tasks.
- Keep validation gaps as first-class deliverables rather than prose notes.

This will make future agent work much less likely to implement against an old promise.

### Make Docs Contract Tests Cheap

Several docs/code mismatches could be caught with small checks:

- grep README/runtime docs for `--project`, `--recheck`, `stage logs apache`, `--tui`, and `--no-tui`
- compare documented command examples against `stage <command> --help`
- run a non-mutating command-example smoke suite for help/version/status dry paths

### Finish Documentation Information Architecture

Spec 009 is the right idea. The current docs contain product docs, architecture docs, plans, specs, visual design, prototypes, migration notes, and historical analysis in nearby locations. Users need a smaller front door:

- README: install, first run, common workflows, troubleshooting links
- User manual: setup, project config, daily work, DNS/TLS, multi-project, recovery
- Command reference: generated or maintained from Cobra help
- Contributor docs: architecture, tests, release, specs
- Archive: historical-only material clearly outside the active contract

### Align Release Surfaces

The GitHub release workflow exists in [.github/workflows/release.yml](../.github/workflows/release.yml), while [Makefile](../Makefile) still has a placeholder `release` target. Align them so maintainers have one obvious release path.

## Recommended Roadmap

This roadmap has been expanded into [Recommended Roadmap Implementation Plan](recommended-roadmap-plan.md) and [Recommended Roadmap Task List](recommended-roadmap-tasks.md).

### Phase 1: Contract Reconciliation

1. Fix or document the mismatches for `status --project`, `logs` positional service names, `setup --recheck`, `doctor` gateway checks, `setup` mutation/remediation, and output-mode flags.
2. Wire or remove `--profile`.
3. Fix late `stage up` rollback route cleanup.
4. Decide whether `.dev` TLS is supported now or explicitly partial.

### Phase 2: Guided Foundation

1. Add `core/guidance` planner and cheap context collection.
2. Add root no-args routing for TTY, non-TTY, `STAGESERVE_NO_TUI=1`, `--help`, and direct subcommands.
3. Add text fallback before full Bubble Tea screens.
4. Keep all mutations behind existing lifecycle/onboarding seams.

### Phase 3: Guided First-Run And Day-2 Actions

1. Implement guided project config preview and confirmation.
2. Route setup/init/up/status/logs/down through existing domains.
3. Add safe confirmations for stop, detach, overwrite, and recovery changes.
4. Validate in real terminals and record evidence in spec 007 quickstart.

### Phase 4: Process Hardening

1. Add Docker integration tests and release smoke checks.
2. Declutter docs under spec 009.
3. Align Makefile, CI, and release workflow.
4. Run final real-daemon validation before calling the guided experience complete.

## Bottom Line

StageServe has a genuinely strong runtime core and a clear product direction. The direct-command Go rewrite is far enough along to be useful, and the architecture gives future work good seams to build on.

The next risk is not "can this be built?" The next risk is contract drift. The project has several places where docs, specs, help text, and implementation disagree. Tightening those now will make the guided TUI work much smoother and keep the product from feeling larger than it actually is.