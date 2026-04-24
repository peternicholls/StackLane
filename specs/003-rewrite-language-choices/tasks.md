---
description: "Tasks for the Stacklane Go rewrite (modular architecture)"
---

# Tasks: Rewrite Stacklane Core In A Compiled, Modular Language

**Input**: Design documents from `/specs/003-rewrite-language-choices/`  
**Prerequisites**: [plan.md](./plan.md) (required), [spec.md](./spec.md) (required for user stories), [Language-Choices-Research-Report.md](./Language-Choices-Research-Report.md), [StackLane-Modular-Architecture-Rewrite-Research-Report.md](./StackLane-Modular-Architecture-Rewrite-Research-Report.md)

**Tests**: Tests are explicitly required by this feature (see plan Phase 0 and Phase 9, FR-006, SC-003, SC-008). Unit tests for module logic and golden-file tests for gateway output are mandatory; full Docker-against-real-daemon integration tests are CI-gated.

**Operational Verification**: Each user story includes operator-facing validation against the current Stacklane contract, failure-path validation, and documentation alignment tasks per the constitution.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US5)
- File paths assume the new Go module rooted at the repository root (per [plan.md](./plan.md) "Target Module Structure").

## Path Conventions

Single Go module at the repository root:

- `cmd/stacklane/` — cobra root command + subcommand wiring
- `core/{config,project,state,lifecycle}/` — operator-facing semantics
- `infra/{docker,compose,gateway}/` — orchestration of Docker and the shared gateway
- `platform/{dns,tls,ports}/` — host-OS integrations (build-tagged where needed)
- `observability/{status,logs}/` — status, drift, log streaming
- Tests sit alongside the code they cover (`*_test.go`); integration tests live under `tests/integration/`.
- The Bash implementation is archived under `previous-version-archive/`; the repo-root `stacklane` shim execs `stacklane-bin`, and root-level `20i-*` wrappers are not part of the active runtime.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the Go module, project layout, and toolchain conventions.

- [X] T001 Initialize Go module at the repository root (`go mod init github.com/peternicholls/stacklane`); pin minimum Go version (1.22+) and add `.go-version` file.
- [X] T002 Create the directory structure from [plan.md](./plan.md) "Target Module Structure" (`cmd/stacklane/`, `core/{config,project,state,lifecycle}/`, `infra/{docker,compose,gateway}/`, `platform/{dns,tls,ports}/`, `observability/{status,logs}/`, `tests/integration/`) with placeholder `doc.go` files.
- [X] T003 [P] Configure linting and formatting: `gofmt`, `go vet`, `staticcheck`; add `Makefile` (or `task` file) targets for `build`, `test`, `lint`, `release`.
- [X] T004 [P] Add CI configuration that runs `go build ./...`, `go vet ./...`, `staticcheck ./...`, `go test ./...` on every push to the feature branch.
- [X] T005 [P] Add `testify` (`require`, `assert`, `mock`) to `go.mod` and document the test harness convention (table-driven unit tests, interface mocks at module boundaries) in `docs/contributing.md`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define every cross-module interface and the typed data model that replaces the existing global shell variables. No implementations yet — this phase locks contracts so user stories can proceed in parallel.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete. Phases 1 and 2 together correspond to plan Phase 0 (Scaffolding and Interfaces).

- [X] T006 Define the `ProjectConfig` struct in `core/config/types.go` with the current Stacklane fields (name, slug, dir, hostname, suffix, ports, docroot, PHP version, etc. per plan "Interface Definitions").
- [X] T007 [P] Define the `ConfigLoader` interface and `CLIFlags` struct in `core/config/loader.go`.
- [X] T008 [P] Define the `RegistryRow`, `AttachmentState`, and `StateStore` interface in `core/state/types.go` (typed; no positional TSV columns).
- [X] T009 [P] Define the `PortAllocation` type and the `PortAllocator` interface in `platform/ports/types.go` (collision check operates over `[]RegistryRow`, not globals).
- [X] T010 [P] Define the `DockerClient` interface (network and container query operations only — no compose methods), `Container`, and `WaitHealthy` helpers in `infra/docker/types.go`. Define the `Composer` interface (`Up`, `Down`) in `infra/compose/types.go` so compose subprocessing has a single owner separate from SDK calls.
- [X] T011 [P] Define the typed `Route` struct, `HealthState`, and `GatewayManager` interface in `infra/gateway/types.go`.
- [X] T012 [P] Define the `DNSProvider`, `DNSStatus`, and `TLSProvider` interfaces in `platform/dns/types.go` and `platform/tls/types.go`.
- [X] T013 [P] Define the `Orchestrator` interface in `core/lifecycle/types.go` covering Up, Down, Attach, Detach, Status, Logs, with typed step errors (named-step + project + remediation) per FR-013.
- [X] T014 Add cobra root command and subcommand stubs in `cmd/stacklane/main.go` and `cmd/stacklane/commands/*.go` that return `ErrNotImplemented`. Validate the current command surface: `up`, `down`, `attach`, `detach`, `status`, `logs`, `dns-setup`.
- [X] T015 Document the module layout, public interfaces, and "where to make a change" guide in `docs/architecture.md` (supports FR-015 and SC-006 for US4; written now because the interfaces are locked here).

**Checkpoint**: `go build ./...` succeeds; `go test ./...` passes (no real tests yet); interfaces are reviewed and locked. User stories can now proceed in parallel.

---

## Phase 3: User Story 3 — Current Stacklane Contract Layer (Priority: P1) 🎯 MVP

**Goal**: Make the new binary read current `.stacklane-local` config and `.stacklane-state` JSON correctly without falling back to `.20i-*` compatibility paths.

**Independent Test**: Take a current Stacklane project, load `.env` / `.stacklane-local` / shell env / CLI flag combinations through `ConfigLoader`, and confirm the resolved values follow the documented current precedence chain. Read and write JSON state under `.stacklane-state`.

This story is sequenced first because **US2 and US1 cannot ship without it**: lifecycle and CLI both consume `ProjectConfig` and `StateStore`.

### Tests for User Story 3

- [X] T016 [P] [US3] Unit-test fixtures in `core/config/testdata/` covering every documented combination of CLI flag / `.stacklane-local` / shell env / `.env` / default that the current Stacklane contract supports.
- [X] T017 [P] [US3] Round-trip tests for `core/state` covering write → read → registry projection.
- [X] T018 [P] [US3] State tests confirm JSON read/write and registry projection under `.stacklane-state`; no automatic legacy state migration is part of startup.

### Implementation for User Story 3

- [X] T019 [P] [US3] Implement slug derivation, hostname resolution, and path canonicalization as pure functions in `core/project/project.go`. (Replaces `twentyi_resolve_docroot`, `twentyi_resolve_hostname`.)
- [X] T020 [US3] Implement `ConfigLoader.Load` in `core/config/loader.go` honoring the precedence chain CLI flags → `.stacklane-local` → shell environment → `.env` → defaults (FR-003). Depends on T006, T019.
- [X] T021 [P] [US3] Implement `StateStore` in `core/state/store.go` with atomic writes (temp file + `os.Rename`) and per-project JSON files (FR-008). Depends on T008.
- [X] T022 [US3] Implement `Registry()` in `core/state/registry.go` returning `[]RegistryRow` aggregated from per-project state files (no TSV column positions; replaces `twentyi_refresh_registry`). Depends on T021.
- [X] T023 [US3] Remove startup migration from `core/state`; `.stacklane-state` JSON is the only active state format (FR-004). Depends on T021, T022.
- [X] T024 [US3] Remove root-level `20i-*` wrapper scripts from the active runtime; do not add compatibility-wrapper tests by default.

**Checkpoint**: Config and state layers follow the current Stacklane contract; `stacklane status` is not yet operator-facing — that arrives with US2.

---

## Phase 4: User Story 2 — Predictable Lifecycle With Clear Errors (Priority: P1)

**Goal**: Make `up`, `down`, `attach`, `detach`, `status`, and `logs` behave deterministically, fail with actionable errors, and never corrupt unrelated projects.

**Independent Test**: Run the lifecycle scenario from spec acceptance scenarios — start two projects with overlapping requested ports and confirm the second `stacklane up` refuses with a named conflict; force a partial failure mid-`up` and confirm `status` reflects reality and recovery is possible without manual file editing; run two concurrent `stacklane up` invocations and confirm port allocation is race-safe.

### Tests for User Story 2

- [X] T025 [P] [US2] Unit tests in `platform/ports/allocator_test.go` for collision detection as a pure function over `[]RegistryRow` (covers SC-008 case logic).
- [X] T026 [P] [US2] Concurrency stress test in `platform/ports/allocator_concurrent_test.go` driving two `Allocate` calls in parallel and asserting no overlap (SC-008).
- [X] T027 [P] [US2] Unit tests in `core/lifecycle/orchestrator_test.go` driving the Up flow with mocked `DockerClient`, `StateStore`, `GatewayManager`, asserting rollback at each failure point.
- [X] T028 [P] [US2] Unit tests in `core/lifecycle/errors_test.go` asserting that every typed step error names the failing step, the affected project, and a stated next action (FR-013).

### Implementation for User Story 2

- [X] T029 [P] [US2] Implement `PortAllocator` in `platform/ports/allocator.go`: bind-check via `net.Listen` with `lsof`/`ss` fallback; collision detection over `[]RegistryRow`; file-based lock (exclusive `os.OpenFile`) to serialize concurrent `stacklane up` invocations (FR-007, SC-008). Depends on T009, T022.
- [X] T030 [P] [US2] Implement typed step errors in `core/lifecycle/errors.go` that wrap underlying errors with step name + project name + remediation hint, and a top-level error renderer for the cobra surface (FR-013).
- [X] T031 [US2] Implement `Orchestrator.Up` in `core/lifecycle/up.go` matching the Orchestration Flow in [plan.md](./plan.md) steps 1–11, with rollback at steps 6–9. Depends on T020, T021, T022, T029, T030, plus stub `DockerClient` and `GatewayManager` (filled in US5 and parallel work).
- [X] T032 [P] [US2] Implement `Orchestrator.Down` in `core/lifecycle/down.go` (idempotent; succeeds even when the project is partially up).
- [X] T033 [P] [US2] Implement `Orchestrator.Attach` and `Orchestrator.Detach` in `core/lifecycle/attach.go` preserving the existing semantics from `twentyi_up_like`/`twentyi_down_like` for attach/detach modes.
- [X] T034 [US2] Implement `Status` in `observability/status/status.go` reading `StateStore.Registry()` and reconciling against `DockerClient.ListContainersByLabel`; report drift explicitly (FR-010, SC-005). Depends on T022 and the docker label query (T040).
- [X] T035 [US2] Implement `Logs` in `observability/logs/logs.go` streaming via `DockerClient.ContainerLogs` with `Follow: true`; on missing/unhealthy container, emit a named diagnostic instead of an empty stream (per spec edge case).
- [X] T036 [US2] Wire the cobra subcommand stubs from T014 to the orchestrator implementations; ensure exit code conventions match spec-002 (non-zero on failure, named errors).
- [ ] T037 [US2] Operator validation: run `stacklane up` / `stacklane status` / `stacklane down` for a representative project under the Go implementation and verify current Stacklane behavior. **DEFERRED** (Option A scope: requires live Docker daemon or release pipeline).

**Checkpoint**: Lifecycle commands work end-to-end against a real Docker daemon when invoked through the binary; concurrent `stacklane up` is race-safe; failures are operator-actionable.

---

## Phase 5: User Story 5 — Docker As A Declarative Partner (Priority: P2)

**Goal**: Replace shell polling with Docker healthchecks; use a single labeled query for project introspection; move phpMyAdmin behind an opt-in profile.

**Independent Test**: Run `stacklane up` and confirm it returns only after the project's primary services report healthy; on timeout, confirm the failure names the unhealthy services. Run default `stacklane up` and confirm phpMyAdmin does not start; opt in and confirm it does.

This story is P2 because the lifecycle works without it — but it's the natural home for the Docker SDK implementation that US2 needs in stub form, so it runs in parallel with US2 and the stubs get filled here.

### Tests for User Story 5

- [X] T038 [P] [US5] Unit tests in `infra/docker/client_test.go` with a mocked Docker SDK transport covering `ListContainersByLabel` and `WaitHealthy`. Compose subprocess invocation is covered separately by tests against the `Composer` interface in `infra/compose/compose_test.go`.
- [ ] T039 [P] [US5] Integration test in `tests/integration/healthcheck_wait_test.go` (Docker-gated) that asserts `WaitHealthy` blocks until services report healthy and times out with a named-services error. **DEFERRED** (Option A scope: requires live Docker daemon or release pipeline).

### Implementation for User Story 5

- [X] T040 [US5] Implement `DockerClient` in `infra/docker/client.go` wrapping the Docker Engine SDK (`github.com/docker/docker/client`). Implement `NetworkExists`, `CreateNetwork`, `RemoveNetwork`, `ListContainersByLabel` (single labeled query — replaces per-service `docker ps` subprocess calls). Compose operations are intentionally **not** on this interface — they live in `infra/compose` (T042).
- [X] T041 [P] [US5] Implement `WaitHealthy` in `infra/docker/health.go` via the Docker SDK event stream / container inspect loop. Default timeout 120 s; honor `--wait-timeout` flag and `STACKLANE_WAIT_TIMEOUT` env via the standard precedence chain.
- [X] T042 [P] [US5] Implement `Composer` (defined in T010) in `infra/compose/compose.go` — the single owner of `docker compose` subprocess invocation. `Up` invokes `docker compose --wait` where supported and accepts profiles + a wait timeout; `Down` is idempotent. Kept separate from `DockerClient` so SDK and CLI surfaces don't tangle (FR-011).
- [X] T043 [US5] Add `HEALTHCHECK` directives to `docker-compose.yml` for nginx, apache/PHP-FPM, and MariaDB; add `depends_on: condition: service_healthy` where applicable.
- [X] T044 [US5] Move phpMyAdmin into a Compose `profiles: [debug]` block (FR-011); add `--profile debug` plumbing to the Up command and document the opt-in in `--help`.

**Checkpoint**: `stacklane up` waits for health; phpMyAdmin is opt-in.

---

## Phase 6: Gateway, DNS, And TLS

**Goal**: Implement the remaining `infra/` and `platform/` modules so the lifecycle has real backing services. These tasks support US2 and US3 — they're listed as a separate phase only because they're mechanically distinct from lifecycle logic.

### Implementation

- [X] T045 [P] [US2] Implement `GatewayManager` in `infra/gateway/manager.go`: `WriteConfig`, `AddRoute`, `RemoveRoute`, `Reload`. Use `text/template` over a typed `Route` struct (replaces heredoc string interpolation in `twentyi_write_gateway_config`). Atomic writes (temp file + rename); preserves the existing nginx upstream / 127.0.0.11 DNS resolver pattern.
- [X] T046 [P] [US2] Golden-file tests in `infra/gateway/manager_test.go` asserting generated nginx configs match the current Stacklane gateway contract for every documented route shape.
- [X] T047 [P] [US3] Implement macOS `DNSProvider` in `platform/dns/macos.go` (build tag `darwin`): dnsmasq via Homebrew, `/etc/resolver/<suffix>`, `osascript` privilege escalation. Isolates every `osascript` and `brew` call behind the build tag.
- [X] T048 [P] [US3] Implement Linux `DNSProvider` stub in `platform/dns/linux.go` (build tag `linux`) returning a clear "not supported on this platform" error per FR-012.
- [X] T049 [P] [US3] Implement `TLSProvider` in `platform/tls/mkcert.go` wrapping `mkcert` subprocess; expose cert/key paths and expiry detection.
- [X] T050 [US3] Wire `dns-setup` cobra command to `DNSProvider.Bootstrap`; verify on macOS and verify the clear unsupported-platform error on Linux build.

---

## Phase 7: User Story 1 — Single Binary Distribution (Priority: P1) 🚀 MVP — Ship Gate

**Goal**: Make the rewrite installable on a clean macOS machine without an interpreter, and ensure startup time does not regress.

**Independent Test**: On a clean macOS environment with Docker Desktop and no extra language runtime, install via the documented method and run `stacklane up` in a project; confirm it starts and is reachable through the gateway. Measure `stacklane --help` cold-shell startup time.

This story sits at the end because it depends on US3 + US2 being functionally complete. The story itself is mostly distribution work, not new feature code.

### Tests for User Story 1

- [ ] T051 [P] [US1] Startup-time benchmark in `tests/perf/startup_test.go` (or a Makefile target) measuring `stacklane --help` cold invocation; fail if cold-shell time exceeds 100 ms on the macOS reference machine. **DEFERRED** (Option A scope: requires live Docker daemon or release pipeline).
- [ ] T052 [P] [US1] Clean-environment install test (documented manual procedure under `tests/install/README.md`): on a fresh macOS user account with only Docker Desktop, follow the install path and run `stacklane up`. Captures SC-001. **DEFERRED** (Option A scope: requires live Docker daemon or release pipeline).

### Implementation for User Story 1

- [ ] T053 [US1] Add release pipeline (GitHub Actions or equivalent) producing signed `darwin/arm64` and `darwin/amd64` binaries on tag push; publish as GitHub Release artifacts. **DEFERRED** (Option A scope: requires release pipeline).
- [ ] T054 [US1] Document the install path in `README.md` (download artifact + place on `PATH`, or `go install`); explicitly state no language-runtime dependency (FR-001, SC-001). **DEFERRED** (Option A scope: requires release pipeline).
- [X] T055 [US1] Update `stacklane` entry point to a thin shim that execs `stacklane-bin` directly.
- [X] T056 [US1] Update documentation to describe the current Go binary path and remove default silent-migration claims.

**Checkpoint**: 🎯 **MVP READY**. A new operator can install and run; an existing operator can upgrade and keep their projects.

---

## Phase 8: User Story 4 — Modular Contributor Experience (Priority: P2)

**Goal**: Make it easy for a new contributor to find the right module, mock its dependencies, and ship a change.

**Independent Test**: Pick a representative change (adjusting how port collisions are reported) and confirm a contributor can locate the responsible module from the published docs in under 5 minutes and write a unit test for it without starting Docker.

Most of the architectural foundation for this story already lands in T015 during Phase 2. The remaining tasks publish the developer-facing surface.

### Tests for User Story 4

- [X] T057 [P] [US4] Add a "no Docker required" tag/build constraint check in CI: the unit test suite (`go test ./... -short`) MUST pass without a Docker daemon present (validates the interface-mock discipline of FR-006).

### Implementation for User Story 4

- [X] T058 [P] [US4] Expand `docs/architecture.md` (started in T015) with: ownership table mapping each operator-visible behavior in `stacklane` to its module; "how to add a new command" walkthrough; "how to add a new module" walkthrough; testing conventions.
- [X] T059 [P] [US4] Add `CONTRIBUTING.md` at the repo root summarizing build/test/lint workflow, branch protections, and link into `docs/architecture.md`.
- [X] T060 [P] [US4] Generate interface mocks (e.g., via `mockery` or hand-rolled) for every interface defined in Phase 2; commit under `internal/mocks/` so contributors don't need to regenerate them to run tests.

---

## Phase 9: Integration, Parity, And Migration Validation

**Purpose**: Cross-story validation against a real Docker daemon and the current Stacklane contract. Maps to plan Phase 9.

- [ ] T061 [US2] Integration test suite in `tests/integration/lifecycle_test.go` (Docker-gated) exercising full `up` → `status` → `logs` → `down` for a representative project against a real Docker daemon. **DEFERRED** (Option A scope: requires live Docker daemon or release pipeline).
- [X] T062 [US3] Validate there is no automatic `.20i-state` migration path in the active runtime; obsolete state can only be handled by future explicit tooling.
- [ ] T063 [US3] Mid-flight reconciliation test: with current `.stacklane-state` records and live containers, run `stacklane status` from the new binary and assert it identifies drift without forcing `stacklane down`. **DEFERRED** (Option A scope: requires live Docker daemon or release pipeline).
- [ ] T064 [US2] Failure-path tests: simulate a port collision, a missing Docker daemon, and a mid-`up` container failure; assert each surfaces a named, actionable error per FR-013. **DEFERRED** (Option A scope: requires live Docker daemon or release pipeline).
- [X] T065 [US1] Remove side-by-side Bash parity stabilization as a default requirement; validate the Go binary against the current Stacklane contract instead.

---

## Phase 10: Polish & Cleanup

**Purpose**: Remove the Bash implementation from active runtime paths and finalize documentation alignment. Maps to plan Phase 10.

- [X] T066 Documentation alignment sweep across `README.md`, `docs/runtime-contract.md`, and `docs/migration.md` to reflect the binary-based runtime.
- [X] T067 [P] Update the `stacklane --help` text and per-subcommand help to match published documentation, including the `--profile debug` opt-in for phpMyAdmin.
- [X] T068 Move the Bash implementation (`lib/stacklane-common.sh` and Bash-era helpers) to `previous-version-archive/`; remove deprecated wrapper scripts from the repo root.
- [ ] T069 [P] Validate startup, status/inspection, teardown, and at least one failure path one final time after the cleanup move (constitution validation gate). **DEFERRED** (Option A scope: requires live Docker daemon or release pipeline).
- [ ] T070 Validate the claimed friction reductions: no language-runtime install, opt-in phpMyAdmin, named errors, race-safe concurrent `stacklane up`, drift reporting on `stacklane status` (constitution: friction-removal validation). **DEFERRED** (Option A scope: requires live Docker daemon or release pipeline).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)** → independent; can start immediately.
- **Foundational (Phase 2)** → depends on Setup; blocks all user stories.
- **US3 (Phase 3)** → blocks US2 and US1 (config + state are consumed by lifecycle and CLI).
- **US2 (Phase 4)** → depends on US3; can run in parallel with US5 (Phase 5) and Gateway/DNS/TLS (Phase 6) once their interface stubs exist.
- **US5 (Phase 5)** → can run in parallel with US2 once interfaces are locked; finishes the Docker stubs US2 leans on.
- **Gateway/DNS/TLS (Phase 6)** → can run in parallel with US2 and US5.
- **US1 (Phase 7)** → depends on US2 + US3 being functionally complete and on the Docker, gateway, DNS, TLS modules existing. Ships the MVP.
- **US4 (Phase 8)** → depends on Phase 2 (interfaces locked); the bulk of its docs land in T015. Can finish in parallel with later phases.
- **Integration & Parity (Phase 9)** → depends on US1, US2, US3, US5 being implemented.
- **Polish & Deprecation (Phase 10)** → depends on Phase 9 sign-off.

### User Story Dependencies

- **US3 (P1)** — no dependencies; first to start after Phase 2.
- **US2 (P1)** — depends on US3 (consumes `ConfigLoader` and `StateStore`).
- **US5 (P2)** — independent of US2/US3 once interfaces are locked; runs in parallel and supplies the Docker layer US2 needs in real form.
- **US1 (P1)** — depends on US2 + US3 + US5 + Phase 6.
- **US4 (P2)** — depends only on Phase 2; can land any time after.

### Within Each User Story

- Tests are written first (table-driven units) and asserted to fail before implementation lands, per Phase 0 convention.
- Pure functions (`core/project`) before stateful loaders (`ConfigLoader`).
- Stores (`StateStore`) before consumers (`Orchestrator`).
- Interface implementations before CLI wiring.

### Parallel Opportunities

- All Phase 2 interface-definition tasks marked [P] (T007–T013) can run in parallel; they touch different files.
- Within US3: T019, T021 can run in parallel after the interfaces land; T020 depends on T019.
- Within US2: T029, T030, T032, T033 can run in parallel; T031 (Up flow) depends on most of them.
- US5 (Phase 5) and Phase 6 can run entirely in parallel with US2 (different packages, no shared files).
- US4 documentation tasks (T058–T060) are wholly parallel to lifecycle work.

---

## Implementation Strategy

### MVP First (US3 → US2 → US1)

1. Complete Phase 1 (Setup) and Phase 2 (Foundational interfaces).
2. Complete Phase 3 (US3 — contract layer): config and state load existing operator data.
3. Complete Phase 4 (US2 — predictable lifecycle), Phase 5 (US5 — Docker SDK), and Phase 6 (gateway/DNS/TLS) in parallel.
4. Complete Phase 7 (US1 — distribution): cut a release artifact.
5. **STOP and VALIDATE**: Run Phase 9 integration tests against a real machine.
6. Ship as MVP once SC-001, SC-002, SC-003, SC-005, SC-008, SC-009 pass.

### Incremental Delivery

- Phase 1 + 2 → contributors can build and run tests (no operator value yet).
- + US3 → current config/state tooling works (no operator-facing command yet).
- + US2 → `up`, `down`, `status`, `logs` work end-to-end against a Docker daemon (developer-only at this point; binary not yet packaged).
- + US5 → `up` waits for health; phpMyAdmin opt-in.
- + US1 → **MVP ship**: operator can install and run.
- + US4 → contributor onboarding friction drops.
- + Phase 9 + 10 → Bash implementation archived and removed from active paths.

### Parallel Team Strategy

With 2–3 developers after Phase 2 lands:

1. Developer A: US3 (config + state).
2. Developer B: US5 (Docker SDK + healthchecks) — supplies real implementations of the interfaces US2 will consume.
3. Developer C: Phase 6 (gateway template + DNS + TLS) once US3's `RegistryRow` is stable.
4. After US3 lands: Developer A picks up US2 (lifecycle orchestration) using B's and C's modules.
5. US1 packaging is owned by whichever developer is freed up first after the lifecycle ships.

---

## Notes

- **Tests are required, not optional**, by the spec (FR-006, SC-003, SC-008) and by the constitution validation gate. Unit tests must run without a Docker daemon (T057).
- **Current Stacklane behavior is the hard constraint**: validation tasks must prove the Go runtime follows `.stacklane-local`, `.stacklane-state`, and `stacklane <subcommand>` behavior.
- **No operator-visible new commands**: this rewrite is intentionally scope-bounded to the spec-002 surface. Anything beyond is a follow-up spec.
- **Stop at any checkpoint** to validate the user story before proceeding. The MVP is shippable after Phase 7 + Phase 9 sign-off.
- **Avoid**: cross-module import shortcuts (use the published interfaces); same-file conflicts in parallel tasks; reintroducing `.20i-*` fallback paths or root-level wrapper shims without an explicit user request.
