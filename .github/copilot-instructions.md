# Stacklane Copilot Instructions

Stacklane is a Go CLI for local Docker development that emulates a 20i-style shared-hosting workflow. Treat the Go implementation as the active product; Bash wrappers and older workflow material are reference-only unless the task explicitly targets archive cleanup.

Prefer the current Stacklane contract over legacy compatibility behavior. Use `stacklane` as the canonical CLI entrypoint, keep `STACKLANE_STACK=20i` explicit, treat `docker-compose.20i.yml` as the active per-project runtime definition, and treat `docker-compose.shared.yml` as the shared routing layer.

Preserve the current config ownership model. Project-local overrides belong in `<project>/.env.stacklane`, stack-wide defaults belong in `<stack-home>/.env.stacklane`, and machine-generated runtime env files under `.stacklane-state` are not user-owned inputs.

When changing config, lifecycle, naming, or runtime behavior, keep implementation, operator docs, and active spec-004 artifacts aligned. Update the relevant files together rather than letting README, docs, and workflow-contract/spec text drift apart.

Treat `previous-version-archive/` as historical reference only. Do not restore legacy `20i-*` wrappers, archived TUI plans, or removed migration fallbacks unless the user explicitly asks for archival or compatibility work.

Prefer focused validation over broad test runs. Use the narrowest relevant checks first, especially `go test ./cmd/stacklane/commands`, `go test ./core/config`, and `go test ./core/lifecycle` when those areas change. Use `make test`, `make vet`, or `make lint` only when the change scope justifies it.

Keep changes minimal and contract-driven. Avoid inventing new config surfaces when an existing stack-home, project-local, or runtime-owned boundary already exists.