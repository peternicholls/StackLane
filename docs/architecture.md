# StageServe Architecture (Go rewrite)

Spec: [specs/003-rewrite-language-choices](../specs/003-rewrite-language-choices/spec.md).

This document describes the post-rewrite layout. It is the authoritative
reference for "where should code X live?" questions.

## Module ownership

| Module | Lives in | Owns |
|---|---|---|
| CLI surface | `cmd/stage`, `cmd/stage/commands` | Cobra root + every subcommand. The CLI may only call `core/lifecycle`, `core/state`, `core/config`, `observability/*`, and `platform/dns`. It must NOT touch Docker or compose directly. |
| Configuration precedence | `core/config` | `Loader.Load` resolves CLI flags → project `.env.stageserve` → shell env → stack `.env.stageserve` → defaults. Tests own the precedence chain and the location-based ownership split. |
| Project identity | `core/project` | Pure helpers for slug/hostname/docroot. No I/O beyond stat. |
| State persistence | `core/state` | JSON-per-project store, atomic writes, registry projection. |
| Lifecycle orchestration | `core/lifecycle` | Sequences the 11 steps of `up`, plus down/attach/detach. Wraps every operator-visible failure in `StepError`. |
| Docker SDK | `infra/docker` | Engine SDK only: networks, container queries, healthcheck waiting, log streams. NOT `docker compose`. |
| Compose subprocess | `infra/compose` | The single owner of `docker compose ...` invocation against the resolved stack compose file (currently `docker-compose.20i.yml`). |
| Gateway template | `infra/gateway` | text/template-rendered nginx config + atomic write + add/remove route helpers. Golden-tested. |
| Port allocation | `platform/ports` | flock-protected allocator with bind-check + registry awareness. |
| DNS bootstrap | `platform/dns` | macOS Homebrew + dnsmasq + `/etc/resolver` flow. Linux build returns the unsupported-platform code. |
| TLS provider | `platform/tls` | mkcert wrapper. |
| Status reporting | `observability/status` | Reconciles registry against live `docker ps`; surfaces drift (FR-010). |
| Log streaming | `observability/logs` | Surfaces named diagnostics for missing services. |
| Test mocks | `internal/mocks` | Hand-rolled mocks for every interface so `go test -short ./...` runs without Docker. |

## Adding a new subcommand

1. Create `cmd/stage/commands/<verb>.go`.
2. Build a `*cobra.Command` and have its `RunE` call `loadConfig(flags)` then drive the orchestrator (or a more specific reporter for read-only verbs).
3. Register it in `NewRoot` in `root.go`.
4. Add an integration test under `cmd/stage/commands/<verb>_test.go` (use the mocks package).

Subcommands MUST NOT call docker, compose, or the gateway directly.

## Adding a new module

1. Decide whether it is _core_ (business logic), _infra_ (external system adapter), _platform_ (host OS), or _observability_ (read-only reporting). Place the package accordingly.
2. Define an interface in `<package>/types.go`.
3. Provide a default implementation in a sibling file.
4. Provide a hand-rolled mock in `internal/mocks/`.
5. Add unit tests against the interface using the mock for collaborators.

## Testing conventions

- Unit tests live next to the code they test (`*_test.go`).
- Use `t.TempDir()` for any filesystem state.
- For collaborators across modules, depend on the interface and inject a mock from `internal/mocks`.
- Golden files live under `<package>/testdata/`. Regenerate via the helper in `infra/gateway/manager_test.go` (writes `*.actual` for diff review).
- `go test -short ./...` MUST pass without Docker. Tests that need Docker should be tagged `//go:build integration`.

## Error handling

Every operator-facing error from the lifecycle layer is `*lifecycle.StepError`.
The CLI can call `lifecycle.AsStepError` to surface a structured "next step"
hint (FR-013). Code below the lifecycle returns plain errors; lifecycle wraps
them.
