---
applyTo: "cmd/**,core/**,infra/**,observability/**,platform/**"
---

Prefer the current Go runtime and lifecycle contract over archived Bash behavior.

Before changing behavior, identify the owning slice first:
- config precedence and env ownership live under `core/config`
- orchestration and rollback behavior live under `core/lifecycle`
- compose subprocess behavior lives under `infra/compose`
- gateway routing behavior lives under `infra/gateway`

Keep the active runtime contract intact unless the task explicitly changes it:
- `STAGESERVE_STACK=20i` is the only supported stack kind today
- the active project compose file is `docker-compose.20i.yml`
- the shared layer is `docker-compose.shared.yml`
- shared routing resources use the `stage-*` naming contract

When touching these areas, prefer focused validation with the smallest relevant test command first:
- `go test ./core/config`
- `go test ./core/lifecycle`
- `go test ./cmd/stage/commands`

Do not reintroduce legacy fallback behavior such as `.stackenv`, `<stack-home>/.env`, `.stageserve-local`, or deprecated `20i-*` wrapper semantics unless the task explicitly asks for compatibility restoration.