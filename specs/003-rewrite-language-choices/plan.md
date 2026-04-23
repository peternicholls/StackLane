# Implementation Plan: StackLane Go Rewrite — Modular Architecture

**Branch**: `003-rewrite-language-choices` | **Date**: 2026-04-03  
**Inputs**: `Language-Choices-Research-Report.md`, `StackLane-Modular-Architecture-Rewrite-Research-Report.md`

---

## Summary

Rewrite StackLane's 2,213-line Bash monolith (`lib/stacklane-common.sh`) in Go, decomposed into
enforced module boundaries that eliminate the three documented coupling hotspots: the 22-variable
global save/restore pattern, the hard-coded TSV column dependency between the registry and gateway
config generation, and the absence of typed, rollback-capable orchestration steps.

The deliverable is a single compiled binary (`stacklane`) that exposes the same operator-visible
command surface defined in spec-002, ships without a runtime dependency, and leverages Docker's
own declarative capabilities (healthchecks, `depends_on` conditions, label filtering) instead of
reimplementing them in application code.

---

## Decision Record

**Language chosen: Go**

| Criterion | Rationale |
|---|---|
| No runtime dependency | Single static binary; operators install once with no toolchain to manage |
| Startup speed | Sub-millisecond; indistinguishable from shell invocation |
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
**Operator-visible surface**: Unchanged from spec-002 (`stacklane --up`, `--down`, `--status`, `--attach`, `--detach`, `--logs`, `--dns-setup`)  
**Constraints**: Must preserve config precedence chain (CLI flags → `.20i-local` → shell env → `.env` → defaults), state file locations, and shared gateway semantics

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
type DockerClient interface {
    NetworkExists(name string) (bool, error)
    CreateNetwork(name string) error
    RemoveNetwork(name string) error
    ComposeUp(projectDir, composeFile, projectName string, env []string) error
    ComposeDown(projectDir, composeFile, projectName string) error
    ListContainersByLabel(labels map[string]string) ([]Container, error)
    WaitHealthy(projectName string, timeout time.Duration) error
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
6.  DockerClient.ComposeUp(...)               → per-project stack
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
| Per-service sequential label queries | `docker ps --filter label=io.20i.project.slug=<slug>` — single atomic query |
| Bash subprocess calls to Docker CLI | Docker Engine SDK typed client in `infra/docker` |
| Hard-coded TSV column order in gateway config generation | Typed `RegistryRow` struct; gateway reads fields by name |
| Sequential port checks with global variable save/restore | `StateStore.Registry()` returns typed slice; port check is a pure function over that slice |

---

## Phases

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

- Implement `ConfigLoader` with the full precedence chain: CLI flags → `.20i-local` → shell env → `.env` → defaults
- Implement slug derivation and hostname resolution as pure functions in `core/project`
- Unit-test precedence rules exhaustively — this is the most logic-dense module

**Exit criteria**: `ConfigLoader.Load` returns an identical `ProjectConfig` for every combination of input that the current Bash `twentyi_finalize_context` handles; verified by golden-file tests

---

### Phase 2 — State Storage (`core/state`)

Replaces: `twentyi_write_state`, `twentyi_load_state_file`, `twentyi_remove_state`,
`twentyi_refresh_registry`, registry TSV logic

- Implement `StateStore` backed by per-project JSON files (atomic write via temp-file + `os.Rename`)
- Implement `Registry()` returning `[]RegistryRow` (typed; no TSV column positions)
- Write a one-time migration function that reads existing `.env`-format state files and writes JSON equivalents
- Unit-test atomic write behavior and registry round-trip

**Exit criteria**: State round-trips without data loss; existing state files can be migrated non-destructively

---

### Phase 3 — Port Allocation (`platform/ports`)

Replaces: `twentyi_port_in_use`, `twentyi_port_reserved`, `twentyi_find_available_port`,
`twentyi_validate_requested_ports`, `twentyi_validate_collision`

- Implement port availability check using `net.Listen` (no `lsof`/`netstat` subprocess) with `ss`/`lsof` as a fallback on platforms where bind-check is insufficient
- Implement collision detection as a pure function over `[]RegistryRow` — eliminates the 22-variable save/restore pattern entirely
- Add file-based lock (`flock`-equivalent via `os.File` exclusive open) to prevent race between concurrent `stacklane --up` invocations

**Exit criteria**: Port allocation is deterministic and race-safe; no subprocess calls to `lsof` on the happy path

---

### Phase 4 — Docker Client (`infra/docker`)

Replaces: all `twentyi_compose`, `twentyi_shared_compose`, `twentyi_ensure_shared_infra` subprocess calls

- Implement `DockerClient` wrapping the Docker Engine SDK
- Implement `WaitHealthy` using the SDK's event stream rather than a bash polling loop
- Add `HEALTHCHECK` directives to `docker-compose.yml` for nginx, apache/PHP-FPM, and MariaDB
- Add `depends_on: condition: service_healthy` where applicable
- Add Compose `profiles` for phpMyAdmin (`--profile debug`)

**Exit criteria**: `stacklane --up` and `--down` work end-to-end against a real Docker daemon; no `os/exec` calls to `docker compose` remain in the orchestration path

---

### Phase 5 — Gateway Config Generation (`infra/gateway`)

Replaces: `twentyi_write_gateway_config`, `twentyi_gateway_route_lines`,
`twentyi_gateway_block_for_route`, `twentyi_update_gateway_route`

- Replace heredoc string interpolation with `text/template` and a typed `Route` struct
- Implement `GatewayManager` with `WriteConfig`, `AddRoute`, `RemoveRoute`, `Reload`
- Atomic config writes (temp-file + rename); preserve the existing nginx upstream/DNS resolver pattern (127.0.0.11)
- Unit-test template rendering against known-good nginx config fixtures

**Exit criteria**: Generated nginx configs are byte-for-byte equivalent to the current Bash output for all documented route shapes

---

### Phase 6 — DNS and TLS (`platform/dns`, `platform/tls`)

Replaces: `twentyi_dns_setup`, `twentyi_dnsmasq_*`, `twentyi_ensure_tls_cert`, `twentyi_tls_*`

- Implement `DNSProvider` for macOS (dnsmasq via Homebrew + `/etc/resolver/<suffix>` + `osascript` privilege escalation) behind build tag `darwin`
- Add a Linux stub (`platform/dns/linux.go`) that returns a clear "not supported on this platform" error rather than silently failing
- Implement TLS via `mkcert` subprocess wrapper in `platform/tls`
- Isolate all `osascript` and `brew` calls to the `darwin` build-tagged file — Linux builds compile clean without them

**Exit criteria**: `stacklane --dns-setup` works on macOS; Linux build compiles and returns a clear error; `tls` package returns correct cert/key paths and detects expiry

---

### Phase 7 — Observability (`observability/status`, `observability/logs`)

Replaces: `twentyi_status`, `twentyi_docker_status`, `twentyi_registry_drift_status`,
`twentyi_live_container_summary`, `twentyi_logs`

- Implement status rendering using `DockerClient.ListContainersByLabel` (single atomic label query per project)
- Implement drift detection by comparing `StateStore.Registry()` against live container labels
- Implement log streaming via Docker SDK `ContainerLogs` with `Follow: true`

**Exit criteria**: `stacklane --status` and `stacklane --logs` produce output equivalent to current Bash implementation

---

### Phase 8 — CLI Surface and Legacy Forwarding (`cmd/stacklane`)

Replaces: `stacklane` entry point, `legacy wrapper commands` wrapper scripts

- Wire all cobra subcommands to their lifecycle implementations
- Implement legacy forwarding: `deprecated --up wrapper` → `stacklane --up` (thin shell shim calling new binary; same contract as spec-002)
- Add `--help` output and version flag

**Exit criteria**: All operator-visible commands from spec-002 work with the compiled binary; `legacy wrapper commands` shims produce deprecation notice and delegate correctly

---

### Phase 9 — Integration Tests and State Migration

- Write integration tests that exercise the full `--up` / `--status` / `--down` lifecycle against a real Docker daemon (CI-gated, not run by default locally)
- Write and validate the state migration script: existing `.env`-format state + TSV registry → JSON
- Validate that a system mid-flight (some projects up in Bash state format) migrates cleanly on first `stacklane --status` run

**Exit criteria**: Integration test suite passes; migration is non-destructive and idempotent

---

### Phase 10 — Deprecation and Cleanup

- Remove `lib/stacklane-common.sh` (archive to `previous-version-archive/`)
- Remove Bash-era `legacy wrapper commands` shims from repository root (replace with compiled-binary shims)
- Update `README.md`, `docs/runtime-contract.md`, `docs/migration.md` to reflect Go binary

**Gate**: Only executed after a defined stabilisation period where both Bash and Go implementations run in parallel and produce identical state

---

## Open Questions

| # | Question | Impact | Owner |
|---|---|---|---|
| 1 | Minimum Go version: 1.22 (for `slices`/`maps` stdlib) or 1.21? | Module and test code style | Resolve before Phase 0 |
| 2 | Integration test Docker daemon: use a real local Docker Desktop or a CI service container? | CI config, test isolation | Resolve before Phase 9 |
| 3 | Incremental migration path: run Go binary and Bash monolith in parallel during a stabilisation period? | Risk management | Decide before Phase 8 |
| 4 | Homebrew formula: day-one alongside Phase 10, or separate follow-up spec? | Distribution | Defer unless ops demand it |
| 5 | Linux DNS support scope: stub that errors clearly, or implement systemd-resolved integration? | Portability | Stub is sufficient for initial rewrite; full Linux support is a separate spec |

---

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Golden-file tests for config/gateway diverge from Bash output on edge cases | Medium | Run both implementations side-by-side in Phase 9; capture all divergences before deprecating Bash |
| Docker Engine SDK version mismatch with Docker Desktop installation | Low | Pin SDK to the oldest Docker Desktop version in active use; test against it in CI |
| State migration corrupts registry on systems with non-standard state file encoding | Low | Migration is read-only until explicitly committed; dry-run mode required |
| Go toolchain version management for contributors | Low | `.go-version` file in repo root; `go.mod` minimum version constraint |
| macOS-only DNS code compiled into non-Darwin builds | None | Build tags (`//go:build darwin`) ensure clean cross-compilation |
