# Implementation Plan: Stacklane Go Rewrite — Modular Architecture

**Branch**: `003-rewrite-language-choices` | **Date**: 2026-04-23 | **Spec**: [spec.md](./spec.md)  
**Inputs**: [spec.md](./spec.md), [Language-Choices-Research-Report.md](./Language-Choices-Research-Report.md), [StackLane-Modular-Architecture-Rewrite-Research-Report.md](./StackLane-Modular-Architecture-Rewrite-Research-Report.md)

---

## Summary

Deliver the spec's primary requirement — replace the 2,213-line Bash monolith
(`lib/stacklane-common.sh`) with a single distributable executable that uses the current
`stacklane` command surface directly — by rewriting in Go with enforced module boundaries.

The rewrite eliminates the three documented coupling hotspots: the 22-variable global
save/restore pattern, the hard-coded TSV column dependency between the registry and the
gateway config generator, and the absence of typed, rollback-capable orchestration steps.
It also leverages Docker's own declarative capabilities (healthchecks, `depends_on`
conditions, label filtering, profiles) rather than reimplementing them in application code.

Operator-visible scope, command surface, configuration precedence, and state file semantics
are defined by [spec.md](./spec.md). The current contract is `stacklane <subcommand>`,
`.stacklane-local`, and `.stacklane-state`; legacy `.20i-*` paths and `20i-*` wrappers are
not live requirements.

---

## Clarifications Applied

The following decisions from [spec.md](./spec.md) Clarifications (Session 2026-04-23) shape
specific phases below; each is referenced again in-context where it applies:

| Spec decision | Where it lands in this plan |
|---|---|
| FR-004 — `.stacklane-state` is the sole default state location; no automatic legacy migration on startup | Phase 2 (state storage) |
| FR-009 — health-wait default 120 s; override via `--wait-timeout` flag and `STACKLANE_WAIT_TIMEOUT` env using the standard precedence chain | Phase 4 (Docker client / WaitHealthy) |
| SC-004 — current Stacklane contract validation: nginx config and state JSON are stable machine artifacts; human-readable status, logs, and error messages are semantic-equivalence only | Phase 5 (gateway), Phase 7 (status/logs), Phase 9 (integration validation) |
| SC-007 — cold-shell `stacklane --help` MUST complete in ≤100 ms on the macOS reference machine and within 2× the captured Bash entry-point baseline on the same machine | Phase 7 (release / CI benchmark) |

---

## Decision Record

**Language chosen: Go**

| Criterion | Rationale |
|---|---|
| No runtime dependency | Single static binary; operators install once with no toolchain to manage |
| Startup speed | Cold-shell `stacklane --help` ≤100 ms on the macOS reference machine, within 2× the Bash entry-point baseline (SC-007) |
| Type safety | Typed structs replace 22-variable global save/restore and positional TSV fields |
| Testability | Every module boundary is an interface; unit tests mock I/O without Docker |
| Distribution | `go install`, Homebrew tap, or GitHub Releases binary — all trivial |
| Ecosystem fit | Docker Engine SDK, `cobra`, `text/template` are idiomatic; contributors familiar with adjacent tooling |

Python rejected: distribution friction (no single binary without PyInstaller) and 100–300 ms
cold-start overhead conflict with Principle I.  
Rust rejected: compile-time complexity exceeds functional gain for a subprocess-heavy CLI tool.  
Node/Deno/Bun rejected: runtime dependency and `node_modules` model conflict with ease-of-use principle.

---

## Technical Context

**Language/Version**: Go 1.22+  
**Key dependencies**: `github.com/spf13/cobra` (CLI), Docker Engine SDK (`github.com/docker/docker`), `encoding/json` (state), `text/template` (gateway config)  
**Target platform**: macOS (arm64/amd64); Linux portability preserved by design  
**Testing strategy**: Unit tests mock all interface boundaries; integration tests shell out to the compiled binary against a real Docker daemon  
**Distribution**: Single binary; Homebrew formula deferred to a follow-up spec  
**Operator-visible surface**: Current Stacklane subcommands (`stacklane up`, `down`, `status`, `attach`, `detach`, `logs`, `dns-setup`)
**Constraints**: Must use the current config precedence chain (CLI flags → `.stacklane-local` → shell env → `.env` → defaults), `.stacklane-state`, and shared gateway semantics

---

## Constitution Check

*GATE: Must pass before Phase 0. Re-check after Phase 1.*

- [x] **Ease of use**: Single binary install replaces a multi-file Bash library; the
  shortest obvious operator path (`stacklane up`) is direct. No new prompts, no compatibility
  shims, and no old state migration are part of the common path.
- [x] **Reliability**: Configuration precedence chain is preserved (FR-003, SC-003), state
  writes are atomic (FR-008), and the rewrite is verified against current Stacklane
  contract tests.
- [x] **Robustness**: Per-project isolation preserved; partial-failure recovery is
  explicit in the Orchestration Flow (rollback at steps 6–9); status command reports
  drift between recorded state and live containers (FR-010, SC-005); Docker healthchecks
  replace bash polling (FR-009).
- [x] **Friction removal**: Eliminates the 22-variable save/restore pattern, the
  positional TSV format, and the bash polling loops. Operators get readiness reporting
  instead of polling, opt-in dev services instead of always-on phpMyAdmin, and clear
  error messages instead of shell stack traces (FR-011, FR-013).
- [x] **Documentation alignment**: Phase 10 explicitly updates `README.md`,
  `docs/runtime-contract.md`, `docs/migration.md`, plus contributor-facing module layout
  documentation (FR-015).
- [x] **Validation coverage**: Each phase has typed exit criteria; Phase 9 covers
  end-to-end startup, status, and teardown. Stress
  scenario for concurrent `stacklane up` invocations (SC-008) is part of Phase 3 exit criteria.

**Result**: Constitution Check passes. No violations to record in Complexity Tracking.

---

## Project Structure

### Documentation (this feature)

```text
specs/003-rewrite-language-choices/
├── spec.md                                          # Feature spec (this is the contract)
├── plan.md                                          # This file
├── tasks.md                                         # Generated from this plan
├── Language-Choices-Research-Report.md              # Phase 0 research
├── StackLane-Modular-Architecture-Rewrite-Research-Report.md  # Phase 0 research
└── checklists/
    └── requirements.md                              # Spec quality checklist
```

### Source Code (repository root)

See "Target Module Structure" below for the concrete Go module layout.
The repo-root `stacklane` launcher execs `stacklane-bin` directly. Bash-era wrappers and
the shell runtime are archived under `previous-version-archive/`, not kept at the repo root.

**Structure Decision**: Single Go module rooted at the repository root, with the binary
defined under `cmd/stacklane/`. Module boundaries are enforced by package layout under
`core/`, `infra/`, `platform/`, and `observability/`. No second project, no separate
backend/frontend split.

---

## Target Module Structure

```text
stacklane/
├── cmd/
│   └── stacklane/          # cobra root command + subcommand registration
├── core/
│   ├── config/             # ConfigLoader: precedence chain, validation
│   ├── project/            # ProjectConfig: slug derivation, path resolution (pure, no I/O)
│   ├── state/              # StateStore: read/write/registry (atomic rename, JSON)
│   └── lifecycle/          # Orchestrator: up/down/attach/detach transaction logic
├── infra/
│   ├── docker/             # DockerClient: Engine SDK wrapper
│   ├── compose/            # Compose file templating and subprocess invocation
│   └── gateway/            # GatewayManager: nginx config generation + reload
├── platform/
│   ├── dns/
│   │   ├── macos.go        # dnsmasq + /etc/resolver (build tag: darwin)
│   │   └── linux.go        # systemd-resolved stub (build tag: linux)
│   ├── tls/                # mkcert subprocess wrapper
│   └── ports/              # Port availability: ss/lsof platform-aware
└── observability/
    ├── status/             # Status rendering, drift reporting
    └── logs/               # Log streaming via Docker SDK
```

---

## Interface Definitions

These are the contracts between modules. All cross-boundary calls go through these interfaces;
no module imports another module's concrete types directly.

```go
// core/config
type ProjectConfig struct {
    Name     string
    Slug     string
    Dir      string
    Hostname string
    Suffix   string
    Ports    PortAllocation
    DocRoot  string
    PHPVersion string
    // ... full field list derived from twentyi_finalize_context globals
}

type ConfigLoader interface {
    Load(projectDir string, flags CLIFlags) (ProjectConfig, error)
}

// core/state
type RegistryRow struct {
    Slug     string
    Hostname string
    Ports    PortAllocation
    // ... typed; replaces positional TSV columns
}

type StateStore interface {
    Save(cfg ProjectConfig, state AttachmentState) error
    Load(slug string) (ProjectConfig, AttachmentState, error)
    Remove(slug string) error
    Registry() ([]RegistryRow, error)
}

// infra/docker
// SDK-native operations only. Compose orchestration lives in infra/compose.
type DockerClient interface {
    NetworkExists(name string) (bool, error)
    CreateNetwork(name string) error
    RemoveNetwork(name string) error
    ListContainersByLabel(labels map[string]string) ([]Container, error)
    WaitHealthy(projectName string, timeout time.Duration) error
}

// infra/compose
// Owns every `docker compose` subprocess invocation. Kept separate from
// DockerClient so that Compose CLI subprocessing is isolated from SDK calls
// and can be swapped for SDK-native compose support in a future iteration.
type Composer interface {
    Up(projectDir, composeFile, projectName string, profiles []string, env []string, waitTimeout time.Duration) error
    Down(projectDir, composeFile, projectName string) error
}

// infra/gateway
type GatewayManager interface {
    WriteConfig(routes []Route) error
    AddRoute(r Route) error
    RemoveRoute(slug string) error
    Reload() error
    Health() (HealthState, error)
}

// platform/dns
type DNSProvider interface {
    Bootstrap(suffix, ip string, port int) error
    Status() DNSStatus
}
```

---

## Orchestration Flow (Up)

The lifecycle module coordinates across interfaces. Each step returns a typed error; failures at
steps 6–9 trigger rollback without corrupting other projects.

```
1.  ConfigLoader.Load(projectDir, flags)      → ProjectConfig
2.  StateStore.Registry()                     → port collision check (replaces 22-var save/restore)
3.  DNSProvider.Status()                      → warn if DNS not bootstrapped
4.  DockerClient.NetworkExists / CreateNetwork → idempotent shared network
5.  GatewayManager.WriteConfig(existingRoutes) → placeholder (no new route yet)
6.  Composer.Up(...)                          → per-project stack (subprocess `docker compose --wait`)
7.  DockerClient.WaitHealthy(...)             → replaces bash polling loop
8.  DockerClient.ListContainersByLabel(...)   → capture runtime identity
9.  StateStore.Save(cfg, identity)            → atomic write
10. GatewayManager.AddRoute(newRoute)         → typed Route struct, not TSV string
11. GatewayManager.Reload()                  → nginx hot reload
```

---

## Docker Capability Upgrades

These gaps are addressed during the rewrite, not as separate work:

| Current gap | Replacement |
|---|---|
| Bash retry loop polling nginx readiness (L~1000–1020) | `HEALTHCHECK` in compose files + `docker compose up --wait` / `DockerClient.WaitHealthy` |
| phpMyAdmin always starts | Compose `--profile debug`; opt-in |
| Per-service sequential label queries | One Docker SDK label query per Stacklane project slug |
| Bash subprocess calls to Docker CLI | Docker Engine SDK typed client in `infra/docker` |
| Hard-coded TSV column order in gateway config generation | Typed `RegistryRow` struct; gateway reads fields by name |
| Sequential port checks with global variable save/restore | `StateStore.Registry()` returns typed slice; port check is a pure function over that slice |

---

## Phases

> **Note**: Phase numbers in `plan.md` and `tasks.md` are independent. `plan.md` phases
> are sequenced by implementation layer (config → state → docker → gateway → …).
> `tasks.md` is organised by user story (US1–US5). See the cross-reference at the top
> of `tasks.md` Phase 1.

### Phase 0 — Scaffolding and Interfaces (no working code)

- Set up Go module (`go mod init github.com/peternicholls/stacklane`)
- Create directory structure matching the target layout above
- Define all interface types in their respective packages (no implementations yet)
- Write `ProjectConfig` struct with every field currently tracked as a global variable
- Add `cobra` root command with subcommand stubs that return `ErrNotImplemented`
- Establish test harness conventions: table-driven unit tests, interface mocks via `testify/mock`

**Exit criteria**: `go build ./...` succeeds; `go test ./...` passes (all stubs); interfaces are reviewed and locked

---

### Phase 1 — Config and Project Identity (`core/config`, `core/project`)

Replaces: `twentyi_init_defaults`, `twentyi_load_env_file`, `twentyi_finalize_context`,
`twentyi_resolve_docroot`, `twentyi_resolve_hostname`, `twentyi_resolve_ports`,
`twentyi_load_stack_and_project_config`

- Implement `ConfigLoader` with the full precedence chain: CLI flags → `.stacklane-local` → shell env → `.env` → defaults
- Implement slug derivation and hostname resolution as pure functions in `core/project`
- Unit-test precedence rules exhaustively — this is the most logic-dense module

**Exit criteria**: `ConfigLoader.Load` returns the documented current Stacklane `ProjectConfig` for every supported input combination; verified by unit tests

---

### Phase 2 — State Storage (`core/state`)

Replaces: `twentyi_write_state`, `twentyi_load_state_file`, `twentyi_remove_state`,
`twentyi_refresh_registry`, registry TSV logic

- Implement `StateStore` backed by per-project JSON files (atomic write via temp-file + `os.Rename`)
- Implement `Registry()` returning `[]RegistryRow` (typed; no TSV column positions)
- Do not read `.20i-state` or run automatic state migration on startup
- Unit-test atomic write behavior and registry round-trip

**Exit criteria**: State round-trips without data loss under `.stacklane-state`; no active runtime path reads `.20i-state`.

---

### Phase 3 — Port Allocation (`platform/ports`)

Replaces: `twentyi_port_in_use`, `twentyi_port_reserved`, `twentyi_find_available_port`,
`twentyi_validate_requested_ports`, `twentyi_validate_collision`

- Implement port availability check using `net.Listen` (no `lsof`/`netstat` subprocess) with `ss`/`lsof` as a fallback on platforms where bind-check is insufficient
- Implement collision detection as a pure function over `[]RegistryRow` — eliminates the 22-variable save/restore pattern entirely
- Add file-based lock (`flock`-equivalent via `os.File` exclusive open) to prevent race between concurrent `stacklane up` invocations

**Exit criteria**: Port allocation is deterministic and race-safe; no subprocess calls to `lsof` on the happy path

---

### Phase 4 — Docker Client (`infra/docker`)

Replaces: all `twentyi_compose`, `twentyi_shared_compose`, `twentyi_ensure_shared_infra` subprocess calls

- Implement `DockerClient` wrapping the Docker Engine SDK (network and container query operations only — compose subprocessing lives in `infra/compose`)
- Implement `WaitHealthy` using the SDK's event stream / container inspect loop rather than a bash polling loop; default timeout 120 s; honor the `--wait-timeout` flag and `STACKLANE_WAIT_TIMEOUT` env via the standard precedence chain (FR-009)
- Implement `Composer.Up` / `Composer.Down` in `infra/compose` invoking `docker compose --wait` (subprocess); pass through profiles for opt-in services
- Add `HEALTHCHECK` directives to `docker-compose.yml` for nginx, apache/PHP-FPM, and MariaDB
- Add `depends_on: condition: service_healthy` where applicable
- Add Compose `profiles` for phpMyAdmin (`--profile debug`)

**Exit criteria**: `stacklane up` and `stacklane down` work end-to-end against a real Docker daemon; no per-service `docker ps` / `docker network` subprocess calls remain (those go through the SDK); compose orchestration may continue to subprocess `docker compose` until SDK-native compose support is available

---

### Phase 5 — Gateway Config Generation (`infra/gateway`)

Replaces: `twentyi_write_gateway_config`, `twentyi_gateway_route_lines`,
`twentyi_gateway_block_for_route`, `twentyi_update_gateway_route`

- Replace heredoc string interpolation with `text/template` and a typed `Route` struct
- Implement `GatewayManager` with `WriteConfig`, `AddRoute`, `RemoveRoute`, `Reload`
- Atomic config writes (temp-file + rename); preserve the existing nginx upstream/DNS resolver pattern (127.0.0.11)
- Unit-test template rendering against known-good nginx config fixtures

**Exit criteria**: Generated nginx configs match the current Stacklane gateway contract for all documented route shapes.

---

### Phase 6 — DNS and TLS (`platform/dns`, `platform/tls`)

Replaces: `twentyi_dns_setup`, `twentyi_dnsmasq_*`, `twentyi_ensure_tls_cert`, `twentyi_tls_*`

- Implement `DNSProvider` for macOS (dnsmasq via Homebrew + `/etc/resolver/<suffix>` + `osascript` privilege escalation) behind build tag `darwin`
- Add a Linux stub (`platform/dns/linux.go`) that returns a clear "not supported on this platform" error rather than silently failing
- Implement TLS via `mkcert` subprocess wrapper in `platform/tls`
- Isolate all `osascript` and `brew` calls to the `darwin` build-tagged file — Linux builds compile clean without them

**Exit criteria**: `stacklane dns-setup` works on macOS; Linux build compiles and returns a clear error; `tls` package returns correct cert/key paths and detects expiry

---

### Phase 7 — Observability (`observability/status`, `observability/logs`)

Replaces: `twentyi_status`, `twentyi_docker_status`, `twentyi_registry_drift_status`,
`twentyi_live_container_summary`, `twentyi_logs`

- Implement status rendering using `DockerClient.ListContainersByLabel` (single atomic label query per project)
- Implement drift detection by comparing `StateStore.Registry()` against live container labels
- Implement log streaming via Docker SDK `ContainerLogs` with `Follow: true`

**Exit criteria**: `stacklane status` and `stacklane logs` satisfy the current output contract for human-readable text.

---

### Phase 8 — CLI Surface (`cmd/stacklane`)

Replaces: `stacklane` entry point

- Wire all cobra subcommands to their lifecycle implementations
- Add `--help` output and version flag

**Exit criteria**: All current operator-visible Stacklane subcommands work with the compiled binary; the repo-root `stacklane` shim execs the binary directly.

---

### Phase 9 — Integration Tests

- Write integration tests that exercise the full `up` / `status` / `down` lifecycle against a real Docker daemon (CI-gated, not run by default locally)
- Validate that obsolete `.20i-*` files are ignored by default

**Exit criteria**: Integration test suite passes; the active runtime uses only current Stacklane paths.

---

### Phase 10 — Deprecation and Cleanup

- Remove `lib/stacklane-common.sh` from the active tree (archive to `previous-version-archive/`)
- Remove Bash-era `20i-*` wrapper scripts from the repository root
- Update `README.md`, `docs/runtime-contract.md`, `docs/migration.md` to reflect Go binary

**Gate**: Execute as part of the current rewrite once the Go command path is functional.

---

## Open Questions

| # | Question | Impact | Owner |
|---|---|---|---|
| 1 | Minimum Go version: 1.22 (for `slices`/`maps` stdlib) or 1.21? | Module and test code style | Resolve before Phase 0 |
| 2 | Integration test Docker daemon: use a real local Docker Desktop or a CI service container? | CI config, test isolation | Resolve before Phase 9 |
| 3 | ~~Incremental migration path: run Go binary and Bash monolith in parallel during a stabilisation period?~~ | Resolved | **Resolved** by current repository policy — no stabilization window is required for removing migration-era runtime paths. |
| 4 | Homebrew formula: day-one alongside Phase 10, or separate follow-up spec? | Distribution | Defer unless ops demand it |
| 5 | ~~Linux DNS support scope: stub that errors clearly, or implement systemd-resolved integration?~~ | Resolved | **Resolved** by spec.md FR-012 and Assumptions — Linux build emits a clear unsupported-platform error; full Linux DNS is a separate spec. |

---

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Golden-file tests for gateway output diverge from the current contract on edge cases | Low–Medium | Keep machine-artifact tests focused on the current nginx template and state schema |
| Docker Engine SDK version mismatch with Docker Desktop installation | Low | Pin SDK to the oldest Docker Desktop version in active use; test against it in CI |
| Operators still have obsolete `.20i-*` files | Medium | The binary ignores them by default; document current `.stacklane-*` paths clearly |
| Go toolchain version management for contributors | Low | `.go-version` file in repo root; `go.mod` minimum version constraint |
| macOS-only DNS code compiled into non-Darwin builds | None | Build tags (`//go:build darwin`) ensure clean cross-compilation |
