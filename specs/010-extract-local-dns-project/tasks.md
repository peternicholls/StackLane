---
description: "Tasks for extracting StageServe local DNS into a standalone project"
---

# Tasks: Extract Local DNS Into A Standalone Project

**Input**: Design documents from `/specs/010-extract-local-dns-project/`  
**Prerequisites**: [spec.md](./spec.md), [research.md](./research.md), [architecture.md](./architecture.md), [contracts/stageserve-integration-contract.md](./contracts/stageserve-integration-contract.md), [plan.md](./plan.md), [migration.md](./migration.md)

**Verification**: Focused StageServe tests and macOS smoke checks for `stage dns-setup`, plus isolated tests in the new DNS repo.

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel with other tasks in different files or repos.

## Phase 1: Lock Scope And Contract

**Purpose**: make the split concrete before code moves.

- [ ] T001 [P] Confirm the new repo name, module path, and standalone CLI name.
- [ ] T002 [P] Freeze the StageServe integration contract in `contracts/stageserve-integration-contract.md`.
- [ ] T003 [P] Confirm the initial versioning strategy for the DNS repo (`v0.x` or `v1.0.0`).
- [ ] T004 Decide whether Homebrew packaging is part of the first release or a follow-up.

**Checkpoint**: both repos can target one agreed dependency contract.

## Phase 2: Prepare StageServe For The Split

**Purpose**: add one internal seam before introducing the external repo.

- [ ] T005 Create a StageServe DNS adapter/helper that owns request/result mapping.
- [ ] T006 Update `cmd/stage/commands/dnssetup.go` to call the adapter instead of the provider directly.
- [ ] T007 Update `core/onboarding/readiness_dns_darwin.go` to call the adapter instead of the provider directly.
- [ ] T008 Add or update focused tests for the StageServe adapter mapping.
- [ ] T009 Run `go test ./cmd/stage/commands ./core/onboarding ./core/config` with the in-repo provider still in place.

**Checkpoint**: StageServe has one clear DNS boundary and no behavior change.

## Phase 3: Scaffold The New DNS Repo

**Purpose**: create the standalone project with both library and CLI structure.

- [ ] T010 [P] Create the new repo, `go.mod`, README, and CI workflow.
- [ ] T011 [P] Create the public library package structure and request/result types.
- [ ] T012 [P] Create the standalone CLI entrypoint and command structure.
- [ ] T013 Add basic repo documentation: install, supported platforms, and privilege expectations.

**Checkpoint**: the new repo builds as an independent project.

## Phase 4: Port DNS Behavior Into The New Repo

**Purpose**: move the reusable implementation out of StageServe.

- [ ] T014 Port result-code definitions and status-message behavior from StageServe.
- [ ] T015 Port preview artifact generation while replacing `StateDir` with a generic preview-root concept.
- [ ] T016 Port the macOS provider implementation and parameterize managed-file naming.
- [ ] T017 Port the explicit unsupported-platform implementation for non-darwin builds.
- [ ] T018 Add unit coverage for request validation, preview generation, and status-code behavior.
- [ ] T019 Add CLI smoke checks for status/bootstrap on supported and unsupported platforms.

**Checkpoint**: the new repo reaches functional parity with the current StageServe DNS subsystem.

## Phase 5: Integrate StageServe Through A Local Dependency

**Purpose**: validate the new repo without cutting a public release too early.

- [ ] T020 Add a local `replace` directive in StageServe for the new DNS module.
- [ ] T021 Update the StageServe adapter to use the new dependency.
- [ ] T022 Re-run focused StageServe tests: `go test ./cmd/stage/commands ./core/onboarding ./core/config`.
- [ ] T023 Run a macOS smoke check for `stage dns-setup`.
- [ ] T024 Run a macOS readiness smoke check covering the DNS blocker path.
- [ ] T025 Fix any contract mismatches before tagging the dependency.

**Checkpoint**: StageServe works against the local external module.

## Phase 6: Release And Adopt The Dependency

**Purpose**: make the split official and remove the in-repo implementation.

- [ ] T026 Tag the DNS repo with its initial release.
- [ ] T027 Replace the local `replace` directive in StageServe with the tagged module version.
- [ ] T028 Remove `platform/dns` from StageServe.
- [ ] T029 Re-run focused StageServe tests after the removal.
- [ ] T030 Re-run `stage dns-setup` smoke validation after the removal.

**Checkpoint**: StageServe consumes a tagged external DNS dependency and no longer carries the provider code.

## Phase 7: Documentation And Cleanup

**Purpose**: make the new split understandable and maintainable.

- [ ] T031 Update StageServe docs that describe local DNS bootstrap and readiness behavior.
- [ ] T032 Add or polish standalone DNS project docs for direct users.
- [ ] T033 Document the local development workflow for working across both repos.
- [ ] T034 Decide whether to add packaging channels such as Homebrew now or later and document the decision.

**Checkpoint**: both repos explain the dependency relationship clearly.

## Phase 8: Final Validation

- [ ] T035 Run focused StageServe validation one final time: `go test ./cmd/stage/commands ./core/onboarding ./core/config`.
- [ ] T036 Run new repo validation one final time: `go test ./...`.
- [ ] T037 Record the final macOS smoke-test evidence for `stage dns-setup` and the standalone CLI.
- [ ] T038 Record any intentionally deferred follow-up work such as Linux automation or extra packaging.

## Notes

- Keep TLS and mkcert outside this split.
- Preserve the current StageServe command contract during the cutover.
- Prefer one adapter seam in StageServe rather than multiple direct dependency call sites.
- Do not publish the DNS repo until StageServe has validated it through a local replacement.