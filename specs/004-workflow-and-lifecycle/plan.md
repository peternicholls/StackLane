# Implementation Plan: Workflow And Lifecycle Hardening

**Branch**: `004-workflow-and-lifecycle` | **Date**: 2026-04-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-workflow-and-lifecycle/spec.md`

## Summary

Lock the operator lifecycle contract that now exists in partial form. Keep one post-up bootstrap phase, keep bootstrap configuration project-local, classify bootstrap failure separately from StageServe infrastructure failure, and keep rollback mandatory on bootstrap failure.

Implement the spec by tightening four concrete surfaces rather than widening the runtime: refactor stack-wide naming from `.stackenv` to `.env.stageserve`, shorten project-scoped runtime naming from `stage-` to `stage-`, align documentation and operator guidance to that contract, and formalize real-project validation around the existing orchestrator, config loader, gateway manager, and state store.

## Technical Context

**Language/Version**: Go 1.26.2  
**Primary Dependencies**: `github.com/spf13/cobra`, Docker Engine SDK `github.com/docker/docker`, Go standard library packages for files/templates/JSON, existing compose subprocess wrapper under `infra/compose`  
**Storage**: local files under `.stageserve-state`, stack-owned env defaults file, generated gateway config, Docker runtime state  
**Testing**: `go test` for unit and slice integration coverage in `core/...`, `infra/...`, `platform/...`; manual real-daemon validation for representative projects  
**Target Platform**: macOS with Docker Desktop as the primary operator environment  
**Project Type**: CLI infrastructure tool  
**Performance Goals**: preserve current operator path speed; do not add extra lifecycle phases, prompts, or repeated manual steps to `stage up`  
**Constraints**: preserve deterministic precedence; keep project `.env` application-owned; keep rollback project-scoped; do not add backward-compatibility behavior for old naming; keep shared gateway naming decisions explicit because they affect compose labels and routing  
**Scale/Scope**: one CLI, one shared gateway, multiple attached local projects, one documented representative single-project app and one multi-project validation scenario

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Ease-of-use impact is documented. The plan preserves the shortest operator path (`stage up`) and removes undocumented bootstrap follow-up work plus ambiguous stack-owned naming.
- [x] Reliability expectations are explicit. Canonical names, precedence order, required config scope, rollback semantics, and failure classification are all fixed in the spec with no compatibility window.
- [x] Robustness boundaries are defined. Bootstrap execution, rollback, gateway routing, state persistence, and per-project isolation stay project-scoped and must not affect unrelated attached runtimes.
- [x] Documentation surfaces requiring same-change updates are identified: `README.md`, `docs/runtime-contract.md`, `docs/architecture.md`, `docs/migration.md`, `CONTRIBUTING.md`, `.env.example`, `.env.stageserve.example`, and any operator guidance or in-code docstring that still references `.stackenv` or `stage-<slug>` runtime names.
- [x] Validation covers startup, inspection/status, teardown, and a relevant failure path. Manual real-daemon validation remains required for representative projects because that workflow is not yet formalized in CI.

**Post-Design Re-Check**: Pass. The design keeps one lifecycle phase, no compatibility mode, explicit naming ownership, and real-daemon validation requirements without violating the constitution.

## Decision Record

### Lifecycle Contract

- Keep one bootstrap phase only.
- Run it after StageServe-owned readiness succeeds.
- Run it inside the `apache` service container.
- Source it only from project-root `.env.stageserve`. Enforce that restriction in the config loader (not the orchestrator) by removing `STAGESERVE_POST_UP_COMMAND` from `trackedEnvKeys` and excluding it from the stack-home defaults merge.
- Roll the project back if bootstrap fails, including on operator `Ctrl-C` cancellation. Bootstrap has no separate timeout setting; it inherits the foreground process.

Rationale: this locks the already-implemented behavior in `core/lifecycle/orchestrator.go` instead of widening scope into additional phases or optional degraded states.

Alternatives considered:

- Add post-attach or multi-phase lifecycle hooks now. Rejected because it expands scope before the first phase is properly documented and validated.
- Preserve a failed-but-inspectable runtime. Rejected because the current runtime already rolls back and the spec now prioritizes predictability over inspection mode.

### Naming Contract

- Use `.env.stageserve` as the only stack-owned defaults file. Remove both legacy paths: `<stackHome>/.stackenv` and the `<stackHome>/.env` fallback.
- Use project-root `.env.stageserve` as the canonical project-local StageServe config surface as well; location now defines ownership.
- Use `stage-` as the project-scoped runtime prefix. Enumerated defaults to change: `ComposeProjectName` (`stage-<slug>`) and `WebNetworkAlias` (`stage-<slug>-web`); `RuntimeNetwork` and `DatabaseVolume` derive from `ComposeProjectName`.
- Keep shared resources aligned to the same `stage-` family: `stage-shared` for the shared compose project and shared network, `stage-gateway` for the gateway service network alias.
- Keep project `.env` application-owned.

Rationale: the stack-owned file should be visually obvious and editor-friendly, while project-scoped Docker names should leave more room to identify the attached project in operator views.

Alternatives considered:

- Keep `.stackenv`. Rejected because it is less clear in editor workflows and no longer matches the intended contract.
- Use uppercase `STLN-`. Rejected because current compose project naming and downstream runtime naming are lowercase-oriented.

## Project Structure

### Documentation (this feature)

```text
specs/004-workflow-and-lifecycle/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── workflow-lifecycle-contract.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
└── stage/

core/
├── config/
├── lifecycle/
├── project/
└── state/

infra/
├── compose/
├── docker/
└── gateway/

platform/
├── dns/
├── ports/
└── tls/

observability/
├── logs/
└── status/

docs/
├── runtime-contract.md
└── architecture.md

README.md
.env.example
.env.stageserve.example
docker-compose.shared.yml
docker-compose.yml
```

**Structure Decision**: Keep the existing single-module CLI structure. Implement the feature by updating the existing config loader, lifecycle orchestrator, gateway and state semantics, then align top-level docs and examples. Do not introduce new packages for this feature.

## Implementation Plan

### Phase 0 - Lock The Contract

1. Confirm the current bootstrap execution path in `core/lifecycle/orchestrator.go` is the only lifecycle phase in scope.
2. Confirm the current precedence path in `core/config/loader.go` and remove any ambiguity between stack-owned defaults and application-owned `.env`.
3. Confirm the current runtime naming derivation in `core/config/loader.go` and document which names are project-scoped versus shared.
4. Record the decisions in `research.md` so implementation does not reopen them later.

### Phase 1 - Design The Change Surface

1. Model the bootstrap contract, validation scenario, and naming contract in `data-model.md`.
2. Write the operator-facing workflow contract in `contracts/workflow-lifecycle-contract.md`.
3. Write a validation-first operator runbook in `quickstart.md` that covers happy-path and failure-path checks.
4. Re-check the constitution after those artifacts are written.

### Phase 2 - Implement Runtime Naming And Config Ownership

1. Rename the stack-wide defaults surface from `.stackenv` to `.env.stageserve` in the config loader, docs, and examples. Update the loader package docstring and `loadStackEnv` comment so the in-code description matches the spec.
2. Remove old-name handling rather than keeping compatibility behavior in the common path. Specifically: (a) drop the `<stackHome>/.stackenv` reader, (b) drop the `<stackHome>/.env` fallback, (c) add a regression test asserting neither is loaded.
3. Change project-scoped runtime naming defaults from `stage-<slug>` to `stage-<slug>` where those defaults are derived (`ComposeProjectName`, `WebNetworkAlias`).
4. Keep the shared-gateway compose project and shared network explicit as `stage-shared`, and keep the gateway service network alias as `stage-gateway`, while still distinguishing those fixed shared names from per-project `stage-<slug>` runtime resources.
5. Restrict `STAGESERVE_POST_UP_COMMAND` to project-root `.env.stageserve` only by removing it from `trackedEnvKeys` and excluding it from the stack-defaults merge in `loader.go`.
6. Delete `.stackenv.example` and add `.env.stageserve.example`.

### Phase 3 - Tighten Lifecycle Diagnostics And Boundaries

1. Keep the single post-up hook contract explicit in the lifecycle path, including operator-cancellation behavior (`Ctrl-C` aborts the hook and triggers rollback under the same step name).
2. Improve operator-visible diagnostics so bootstrap failures remain distinct from gateway, DNS, and container health failures.
3. Ensure rollback leaves status and recorded state coherent after bootstrap failure: no record left as `attached`, no gateway route added.
4. Preserve project isolation across rollback, route generation, and state persistence, including the post-readiness/pre-state-persist failure window.

### Phase 4 - Align Documentation And Validation

1. Update `README.md`, `docs/runtime-contract.md`, `docs/architecture.md`, `docs/migration.md`, and `CONTRIBUTING.md` to describe the final naming and lifecycle contract.
2. Update example env files and any operator guidance that still references the old stack-owned file name. Sweep code comments, docstrings, and CLI help text for surviving `.stackenv` / `stage-<slug>` mentions.
3. Validate the documented workflow against one representative bootstrap-sensitive app and one multi-project scenario.
4. Record any remaining manual-only validation gap explicitly instead of implying automation exists.

## Validation Strategy

### Automated Validation

- Run focused tests for `core/config`, `core/lifecycle`, and any touched naming or gateway slices.
- Add or update tests that cover:
  - `.env.stageserve` loading as the canonical stack-wide defaults surface, including `STACK_HOME` override
  - removal of `<stackHome>/.stackenv` and `<stackHome>/.env` as stack-defaults sources (negative regression test)
  - project `.env` remaining application-owned fallback only
  - `stage-<slug>` and `stage-<slug>-web` project-scoped runtime naming defaults
  - shared-resource naming staying `stage-shared` / `stage-gateway`
  - `STAGESERVE_POST_UP_COMMAND` source restriction (negative tests for shell env, `.env.stageserve`, project `.env`)
  - bootstrap failure classification, operator-cancellation rollback, and post-rollback state coherence
  - rollback isolation between concurrently attached projects
  - `Orchestrator.Attach` against the new naming

### Manual Real-Daemon Validation

- Validate one representative application with a post-up bootstrap command.
- Validate one multi-project attach/up/status/down scenario through the shared gateway.
- Check DNS routing, shared-gateway readiness, runtime env injection, DB provisioning alignment, bootstrap execution, rollback after hook failure, isolation of unrelated attached projects, and status output after rollback.

### Repo-To-Deployed-Copy Sync Requirement

- If the repository working copy is not the same copy you run from, sync the relevant changes into the deployed stack copy under `$HOME/docker/20i-stack` before manual validation.
- Validation notes and operator docs must call out that sync point explicitly rather than implying the repository and deployed copy are always the same location.

### Explicit Validation Gap Policy

- If any real-daemon path cannot be rerun during implementation, record the exact gap in docs or validation notes.
- Do not claim end-to-end automation where only manual validation exists.

## Risks And Mitigations

| Risk | Why It Matters | Mitigation |
|---|---|---|
| Shared and project-scoped naming drift apart | Operators will not know which Docker resources belong to which contract | Keep `stage-shared` and `stage-gateway` explicit for shared infrastructure and distinguish them from per-project `stage-<slug>` runtime resources |
| `.env.stageserve` rename leaks into application-owned `.env` behavior | StageServe would blur the ownership boundary the spec is trying to enforce | Keep stack defaults loading and app `.env` fallback tests separate in `core/config` |
| `STAGESERVE_POST_UP_COMMAND` leaks through shell env or stack defaults | The bootstrap source restriction would be a doc-only claim | Enforce the restriction in the config loader (`trackedEnvKeys`, stack-defaults merge) and assert it with three negative-path tests |
| Rollback leaves stale recorded state or gateway routes | Failure handling becomes harder to trust than the bootstrap hook it added | Validate rollback through `stage status`, state-store assertions, and gateway route checks; assert no `attached` record after rollback |
| Bootstrap timeout is implicit and operator-controlled | Operators may expect a baked-in timeout | Document the foreground/`Ctrl-C` model explicitly in the contract and quickstart |
| Real-project validation remains informal | The feature would look complete on paper but remain unproven in practice | Treat the quickstart validation workflow as a required deliverable, not an optional note |

## Complexity Tracking

No constitution violations require justification.
