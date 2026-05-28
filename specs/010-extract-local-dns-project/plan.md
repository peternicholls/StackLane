# Implementation Plan: Extract Local DNS Into A Standalone Project

**Branch**: `010-extract-local-dns-project` | **Date**: 2026-05-28 | **Spec**: [spec.md](./spec.md)

## Summary

Split StageServe's host DNS subsystem into a neutral, standalone DNS project that ships both a Go module and a small CLI. StageServe will keep its current command and readiness surfaces, but those surfaces will call a versioned dependency through a narrow adapter instead of using in-repo provider code.

## Locked Planning Decisions

- StageServe keeps `stage dns-setup`.
- StageServe consumes the new DNS project as a Go module dependency in phase 1.
- A standalone DNS CLI is part of the new project so the split creates an independently useful tool.
- macOS feature parity comes first; unsupported-platform behavior stays explicit.

## Technical Context

**Language/Version**: Go 1.26.2  
**Current StageServe Surface**: `platform/dns`, `cmd/stage/commands/dnssetup.go`, `core/onboarding/readiness_dns_darwin.go`, `core/config.LocalDNS`  
**Target Shape**: two repos, one dependency contract  
**Primary Constraint**: do not make current StageServe users install a separate DNS binary  
**Primary Validation**: focused StageServe tests plus real macOS `stage dns-setup` parity checks

## Decision Record

### Library + CLI, Not CLI Only

- Decision: the new DNS repo ships both a library and a standalone CLI.
- Rationale: this satisfies the reuse goal without pushing an extra runtime dependency onto StageServe users.
- Rejected: shelling out from StageServe to an externally installed DNS binary.

### Adapter Boundary Inside StageServe

- Decision: StageServe should add one adapter seam instead of calling the dependency from multiple places in multiple shapes.
- Rationale: this keeps config translation and result mapping local and makes future upgrades smaller.
- Rejected: wiring the dependency directly in every caller.

### Keep StageServe UX Stable

- Decision: preserve `stage dns-setup` and existing readiness semantics during the split.
- Rationale: the point of the split is ownership and reuse, not user-visible churn.
- Rejected: renaming or removing StageServe DNS surfaces as part of extraction.

## Planned Workstreams

### Workstream 1 - Freeze The Boundary In StageServe

1. Identify all StageServe call sites that currently use `platform/dns`.
2. Add a small adapter layer or boundary package inside StageServe.
3. Move StageServe-specific naming concerns into the adapter.
4. Keep StageServe tests green while still backed by the in-repo implementation.

### Workstream 2 - Build The New DNS Repo

1. Create the new repo with public library and standalone CLI structure.
2. Port the provider types, result codes, preview generation, and platform logic.
3. Replace StageServe-specific names with generic request fields.
4. Add unit coverage and CLI smoke checks in the new repo.

### Workstream 3 - Switch StageServe To The Dependency

1. Add a local module replacement in StageServe for development.
2. Update the StageServe adapter to call the external dependency.
3. Re-run focused StageServe tests and command-level checks.
4. Remove the in-repo `platform/dns` implementation only after the dependency path is stable.

### Workstream 4 - Release And Document

1. Tag a prerelease from the DNS repo.
2. Update StageServe to the tagged version.
3. Document the split in both repos.
4. Decide whether to add distribution extras such as a Homebrew formula immediately or defer them.

## Proposed Sequence

### Phase 0 - Scope And Naming

1. Confirm repo name, module path, and CLI name.
2. Freeze the StageServe integration contract.
3. Decide whether the first public version is `v0.x` or `v1.0.0`.

### Phase 1 - StageServe Prep

1. Add a StageServe adapter package or helper for DNS dependency calls.
2. Route current command and readiness logic through that adapter, still using in-repo code.
3. Add or update focused tests around the adapter mapping.

### Phase 2 - New Repo Scaffold

1. Create the new repo and module.
2. Port types, code vocabulary, preview generation, and platform implementations.
3. Add standalone CLI commands for status/bootstrap/preview or their final equivalent.
4. Add CI and test harnesses.

### Phase 3 - Local Integration

1. Add a local `replace` directive in StageServe.
2. Point the StageServe adapter at the new module.
3. Run focused StageServe validations.
4. Fix contract rough edges before any tag is cut.

### Phase 4 - Release Cutover

1. Tag the DNS repo.
2. Update StageServe to the tagged module version.
3. Remove the in-repo DNS implementation.
4. Re-run StageServe validation and update docs.

### Phase 5 - Standalone Product Polish

1. Improve standalone docs.
2. Decide whether to add packaging channels such as Homebrew.
3. Plan next-platform work separately if Linux automation becomes a goal.

## Validation Strategy

### StageServe Validation

- `go test ./cmd/stage/commands ./core/onboarding ./core/config`
- targeted terminal validation for `stage dns-setup`
- readiness check validation on macOS

### New DNS Repo Validation

- `go test ./...`
- CLI smoke checks for status/bootstrap behavior
- explicit unsupported-platform tests for non-darwin builds

### Integration Validation

- StageServe local module replacement works before release
- tagged dependency upgrade works without local replacement
- StageServe no longer references the old in-repo DNS provider package after cutover

## Risks And Mitigations

| Risk | Why It Matters | Mitigation |
|---|---|---|
| API is still StageServe-shaped | The new project will not be genuinely reusable | Rename `StateDir` to `PreviewRoot`, parameterize ownership metadata |
| StageServe integration spreads across too many files | Upgrades become noisy and error-prone | Add one adapter seam and route all callers through it |
| Users need a second install step | Split becomes a UX regression | Keep StageServe on a Go module dependency in phase 1 |
| Status code drift breaks readiness | StageServe UI loses determinism | Freeze code vocabulary in the contract and version changes explicitly |
| Split balloons into broader onboarding/TLS refactor | Schedule slips and scope muddies | Keep TLS, onboarding redesign, and Linux automation out of scope |

## Deliverables

- new spec package in this directory
- a new standalone DNS repo plan
- a StageServe adapter migration plan
- an execution task list with rollback points