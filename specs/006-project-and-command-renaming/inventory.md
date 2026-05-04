# Rename Inventory: Active Legacy Surfaces

## Purpose

This file is the concrete Gate B disposition table for spec 006.

It records repository-owned active surfaces that still carry the old name and classifies each one as:

- rename now
- rename later in this spec
- historical only
- archive only

`rename later in this spec` is a staging classification, not a final-state exception.

## Disposition Table

| Surface | Examples observed | Disposition | Why | First implementation anchor |
|---|---|---|---|---|
| CLI root command and help text | `cmd/stage/commands/root.go`, `version.go`, `setup.go`, `doctor.go`, `logs.go`, `up.go`, `down.go`, `attach.go`, `status.go`, `dnssetup.go`, `init.go` | rename now | These files define the active operator-facing command surface and help text | `cmd/stage/commands/root.go` |
| Installer command/output surface | `install.sh` asset names, output messages, next-step commands, install destination | rename now | The installer is a direct public contract and must cut over early | `install.sh` |
| Build output names | `Makefile` `BINARY := stage-bin` and `go build -o $(BINARY) ./cmd/stage` | rename now | Build outputs and installer expectations need one canonical binary story | `Makefile` |
| README and install docs | `README.md`, `docs/installer-onboarding.md` | rename now | These are the primary entry points for new operators | `README.md` |
| Runtime and migration docs | `docs/runtime-contract.md`, `docs/migration.md`, `docs/architecture.md`, `docs/contributing.md` | rename now | These are maintained docs and must not keep the old active contract | `docs/runtime-contract.md` |
| Active normative specs and quickstarts | maintained examples under `specs/004-*`, `specs/005-*`, and active planning text under `specs/006-*` | rename now | They still teach the current command and naming model | `specs/005-installer-and-onboarding/quickstart.md` |
| Compatibility shim | any temporary compatibility forwarding path | rename now and forbid by closeout | This spec is a hard switch; no compatibility command path may survive | none |
| Go module path | `go.mod` `module github.com/peternicholls/stageserve` | rename now | Active code identity must match the final product name | `go.mod` |
| Go import paths | imports under `cmd/`, `core/`, `infra/`, `observability/`, `platform/` that still carry the legacy identity | rename now | Stale imports are active code identity defects, not deferred compatibility work | `cmd/stage/main.go` |
| Package and directory names | any repository-owned command/package paths still carrying the legacy identity | rename now | Path/layout rename is part of the complete cutover, not a deferred compatibility slice | `cmd/stage` |
| Internal env and state filenames | `.env.stageserve`, `.stageserve-state`, `.stageserve-local` references, generated text that still emits them | rename later in this spec | They are active surfaces today but must not remain by closeout | `core/config/loader.go` |
| Internal env variable names | `STAGESERVE_*`, `STAGESERVE_PROJECT_ENV_*`, `STACK_HOME` comments or active emitted config text where rename is required | rename later in this spec | Active code identity and config contract must eventually drop the old brand | `core/config/types.go` |
| Runtime Docker prefix | `stage-<slug>` compose project, `stage-<slug>-runtime` network, `stage-<slug>-db-data` volume, `stage-shared` gateway project/network — set in `core/config/loader.go` lines ~297-334 | keep as final state | `stage-*` is the accepted canonical runtime prefix for StageServe | `core/config/loader.go` |
| Internal generated config text | `cmd/stage/commands/project_env.go`, `core/onboarding/project_env.go` comments and emitted file content | rename now | These are user-visible generated artifacts and must match the final naming contract | `cmd/stage/commands/project_env.go` |
| Lifecycle remediation messages | guidance strings like `Run \`stage up\` first` or `Check \`stage logs\`` in `core/lifecycle/orchestrator.go` | rename later in this spec | Active behavior messaging must align after the command cutover | `core/lifecycle/orchestrator.go` |
| Gateway and observability headers | `X-StageServe-*`, `stage-no-route`, `__stage_gateway_health`, gateway error strings in `infra/gateway/**` | rename later in this spec | These are active runtime-visible surfaces and must be renamed deliberately with testdata updates | `infra/gateway/templates.go` |
| Docker label namespace | `io.stage.service` comments or label contracts in `infra/docker/types.go` | rename later in this spec | Active runtime metadata should not retain the old brand by closeout | `infra/docker/types.go` |
| Platform-specific filenames and remediation text | `stage-*.conf`, `stage setup`, `StageServe managed include` under `platform/dns/**` | rename later in this spec | Active system integration artifacts and messages must be renamed carefully | `platform/dns/macos.go` |
| Test fixtures and assertions that model active behavior | `cmd/**_test.go`, `core/**_test.go`, `infra/gateway/testdata/**`, `observability/**_test.go`, `platform/**_test.go` | rename later in this spec | Test expectations must follow the active contract once implementation slices land | owning file for each touched slice |
| Historical CLI analysis | `docs/cli-naming-analysis.md` | historical only | It may retain the old name as analysis, but must be clearly historical | `docs/cli-naming-analysis.md` |
| Archived wrappers and legacy UI | `previous-version-archive/**` | archive only | Archive content must not drive the active rename | `previous-version-archive/` |

## Open Questions Still Requiring Gate A Decisions

1. None. Gate A is closed: `.env.stageserve`, `.stageserve-state`, and `STAGESERVE_*` are final.
2. None. The active module path is `github.com/peternicholls/stageserve`.
3. None. No temporary compatibility shim is allowed.
4. None. Runtime-visible headers and health endpoints use StageServe naming with no compatibility aliases.
5. None. `stage-*` remains the final runtime prefix.

## First Execution Slices Suggested By Inventory

1. Public command and help cutover: `cmd/stage/commands/root.go`
2. Installer and asset naming cutover: `install.sh` and `Makefile`
3. Top-level docs cutover: `README.md` and `docs/installer-onboarding.md`
4. Internal naming contract slice: `core/config/loader.go` and `core/config/types.go`
5. Module/import path slice: `go.mod` and `cmd/stage/main.go`
6. Gateway header and testdata slice: `infra/gateway/templates.go` and `infra/gateway/testdata/**`