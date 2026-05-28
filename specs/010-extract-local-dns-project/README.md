# 010: Extract Local DNS Into Its Own Project

## Goal

Turn StageServe's local DNS bootstrap and inspection logic into a standalone project that StageServe consumes as a dependency, while keeping StageServe's existing operator workflow stable.

## Why This Exists

- The current DNS code is small, isolated, and already behaves like a subsystem.
- Local development DNS bootstrap is useful outside StageServe.
- Independent versioning and release cadence will make the DNS work easier to reuse and maintain.
- StageServe should own project config, onboarding, and lifecycle UX, not the underlying host DNS implementation.

## Recommended Direction

- Create a new neutral-branded DNS project with a versioned Go module and a small standalone CLI.
- Keep `stage dns-setup` and StageServe readiness surfaces, but delegate status/bootstrap work to the dependency.
- Keep StageServe as the owner of `.env.stageserve`, stack defaults, `.stageserve-state`, and user-facing workflow decisions.

## Current Seam Snapshot

- Candidate extraction package: `platform/dns/*`
- Current StageServe command entrypoint: `cmd/stage/commands/dnssetup.go`
- Current readiness integration: `core/onboarding/readiness_dns_darwin.go`
- Current config boundary: `core/config.LocalDNS`

## Document Map

- [spec.md](./spec.md): feature contract and success criteria
- [research.md](./research.md): option analysis and recommendation
- [architecture.md](./architecture.md): target repo split and API shape
- [contracts/stageserve-integration-contract.md](./contracts/stageserve-integration-contract.md): StageServe-to-dependency contract
- [plan.md](./plan.md): implementation sequence across both repos
- [migration.md](./migration.md): staged move and rollback points
- [tasks.md](./tasks.md): dependency-ordered execution checklist

## Working Assumptions

- Phase 1 should use a Go module dependency, not require StageServe users to install a second binary.
- The new DNS project should still ship its own CLI so it is useful without StageServe.
- macOS behavior must reach feature parity first; Linux and other platforms can keep the current explicit unsupported surface until later work is planned.
- Public naming in the new project should be generic rather than StageServe-branded.