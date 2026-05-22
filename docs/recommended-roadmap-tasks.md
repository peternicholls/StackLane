# Recommended Roadmap Task List

Date: 2026-05-22
Plan: [Recommended Roadmap Implementation Plan](recommended-roadmap-plan.md)
Source: [Project Analysis Report](project-analysis-report.md)

## Format

- `[P]` means the task can usually run in parallel with nearby tasks if files do not overlap.
- Validation tasks are deliverables and should be updated with command output or notes when run.
- If a task is intentionally deferred, move it to a deferred section with the reason.

## Phase 1: Contract Reconciliation

**Goal**: Make active code, command help, docs, and specs agree before adding guided behavior.

### Contract Decisions

- [x] T001 Decide whether `status --project <selector>` remains active contract or docs are corrected to current-only/`--all` behavior. Decision: keep and implement `--project` for recorded project selectors.
- [x] T002 Decide whether `logs` should accept positional service names, keep only `--service`, or support both. Decision: support both, with a conflict error when they disagree.
- [x] T003 Decide whether `setup --recheck` should be implemented or removed from active docs. Decision: remove from active docs because setup has no cache to bypass.
- [x] T004 Decide whether `doctor` owns shared-gateway checks or active docs should stop claiming that behavior. Decision: active docs/help now describe the implemented read-only readiness scope.
- [x] T005 Decide whether `setup` is check-only or guided repair/mutation, then update remediation language accordingly. Decision: setup remains check-only; DNS remediation points to `stage dns-setup --site-suffix <suffix>`.
- [x] T006 Decide final output-mode flag shape for active code: spec 007 `--notui`/`--cli` versus current `--tui`/`--no-tui`. Decision: expose `--notui` and `--cli`; keep legacy `--no-tui` hidden for compatibility; remove exposed `--tui`.
- [x] T007 Decide whether `.dev` TLS is fully supported now or explicitly partial/deferred. Decision: keep active and wire lifecycle TLS generation/mounting.
- [x] T008 Decide whether `--profile debug` remains supported or is removed from active surfaces until wired. Decision: keep and wire `debug` into project compose up.
- [x] T008b Decide which cleanup items from the analysis report are release-blocking versus tracked follow-up work. Decision: Phase 1 release blockers were command/docs contract drift, late rollback route cleanup, debug profile wiring, and `.dev` TLS wiring; root flag architecture cleanup remains follow-up.

### Implementation Tasks

- [x] T009 Add or correct `--project <selector>` support for `status` in [cmd/stage/commands/status.go](../cmd/stage/commands/status.go), or remove it from [README.md](../README.md) and [docs/runtime-contract.md](runtime-contract.md).
- [x] T010 Add shared project selector resolution for `status` and `logs` if T001 keeps the selector contract.
- [x] T011 Add positional service support to [cmd/stage/commands/logs.go](../cmd/stage/commands/logs.go), or correct README examples to `stage logs --service apache`.
- [x] T012 Implement or remove `setup --recheck` across [cmd/stage/commands/setup.go](../cmd/stage/commands/setup.go), [docs/runtime-contract.md](runtime-contract.md), and tests.
- [x] T013 Implement gateway diagnostics in [cmd/stage/commands/doctor.go](../cmd/stage/commands/doctor.go), or correct active docs/specs to the implemented readiness scope.
- [x] T014 Fix DNS readiness remediation so a DNS bootstrap blocker points to `stage dns-setup --site-suffix <suffix>` or the decided setup repair flow.
- [x] T015 Update setup/help/README/runtime copy so `setup` mutation behavior is described consistently.
- [x] T016 Implement final no-TUI flag aliases or clean up docs/help to match the current output-mode contract.
- [x] T017 Wire `--profile debug` through lifecycle compose `UpOptions`, or remove the flag/docs until supported.
- [x] T017b Update command `Short` and `Long` text touched by reconciliation so help prefers StageServe user goals over Docker/gateway internals.
- [x] T018 Add route cleanup after late `stage up` failures in [core/lifecycle/orchestrator.go](../core/lifecycle/orchestrator.go).
- [x] T019 Add lifecycle tests proving failed state save or post-route failure does not leave a stale gateway route.
- [x] T020 If `.dev` TLS remains active, wire mkcert generation through [platform/tls](../platform/tls), pass `SHARED_GATEWAY_CERTS_DIR`, and preserve TLS settings when adding routes.
- [x] T021 If `.dev` TLS is deferred, update [README.md](../README.md), [docs/runtime-contract.md](runtime-contract.md), and `.env.stageserve.example` to mark it partial. Not applicable: `.dev` TLS remains active and T020 was implemented.
- [x] T022 Reconcile spec 004 and spec 007 task checkboxes with implemented reality; keep open gaps explicit.
- [x] T023 Add docs/code drift checks for known mismatch patterns: `--project`, `--recheck`, `stage logs apache`, `--tui`, and `--no-tui`.
- [x] T023a Add a command-help contract check that compares documented examples against `stage <command> --help` for non-mutating paths.

### Phase 1 Validation

- [x] T024 Run `go test ./cmd/stage/commands ./core/lifecycle ./observability/status ./infra/gateway`.
- [x] T025 Run `go test -short ./...`.
- [x] T026 Run `go vet ./...`.
- [x] T027 Run `go build ./...`.
- [x] T028 Run `git diff --check`.
- [x] T029 Capture `stage --help`, `stage setup --help`, `stage status --help`, and `stage logs --help` evidence after reconciliation.

### Phase 1 Deferred

- [ ] T008a/T017a Root-level shared flag cleanup is deferred to Phase 4 T078f because moving config, lifecycle, and selector flags to command-local ownership is broader command architecture work. Phase 1 help-copy expectations were updated for the active flags touched here.

### Phase 1 Validation Evidence

- 2026-05-22: `go test ./cmd/stage/commands ./core/lifecycle ./observability/status ./infra/gateway` passed.
- 2026-05-22: `go test -short ./...` passed.
- 2026-05-22: `go vet ./...` passed.
- 2026-05-22: `go build ./...` passed.
- 2026-05-22: `git diff --check` passed.
- 2026-05-22: `go run ./cmd/stage --help`, `go run ./cmd/stage setup --help`, `go run ./cmd/stage status --help`, and `go run ./cmd/stage logs --help` captured from source. Evidence showed `setup` exposes `--notui`/`--cli` without exposed `--tui`/`--no-tui`, `status` exposes `--project`, and `logs` exposes `[service]`, `--project`, and `--service`.

## Phase 2: Guided Foundation

**Goal**: Add the non-UI planner and root routing needed for bare `stage` guided behavior.

### Planner Foundation

- [x] T030 [P] Create `core/guidance/types.go` with TUI capability, guided context, next action plan, action, visible defaults, warnings, and direct command equivalents.
- [x] T031 [P] Create `core/guidance/context.go` with cheap context collection and injected seams for config, state, readiness, and status checks.
- [x] T032 Implement `core/guidance/planner.go` for `machine_not_ready`, `project_missing_config`, `project_ready_to_run`, `project_running`, `project_down`, `drift_detected`, `not_project`, and `unknown_error`.
- [x] T033 Add `core/guidance` tests for pure planner decisions and no-mutation context collection.
- [x] T034 Add a terminal-verifiable planner inspection path or debug command that does not mutate project config or state. Bare `stage` now renders the planner's text fallback without writing project config or state.

### Root Routing

- [x] T035 Add injectable terminal/mode detection for root interaction in [cmd/stage/commands](../cmd/stage/commands).
- [x] T036 Add root no-args `RunE` in [cmd/stage/commands/root.go](../cmd/stage/commands/root.go).
- [x] T037 Implement non-TTY text fallback rendering for `core/guidance.NextActionPlan`.
- [x] T038 Implement `STAGESERVE_NO_TUI=1`, `--notui`, and `--cli` handling for bare `stage`.
- [x] T039 Ensure `stage --help` and direct subcommands bypass guided routing.
- [x] T040 Keep JSON output modes free of guided copy.
- [x] T040a Honor `NO_COLOR=1` in text fallback and any projected output that would otherwise use color.

### Phase 2 Validation

- [x] T041 Run `go test ./core/guidance ./cmd/stage/commands`.
- [x] T042 Verify bare `stage` in a real TTY opens the guided path or initial guided shell.
- [x] T043 Verify `stage > /tmp/stage-guidance.txt` exits without prompting.
- [x] T044 Verify `STAGESERVE_NO_TUI=1 stage` uses text fallback.
- [x] T045 Verify `stage --notui` and `stage --cli` use text fallback.
- [x] T046 Verify `stage --help` and `stage up --help` remain direct help paths.
- [x] T046a Verify `NO_COLOR=1 stage --notui` or equivalent captured output contains no color styling.

### Phase 2 Validation Evidence

- 2026-05-22: `go test ./core/guidance ./cmd/stage/commands` passed.
- 2026-05-22: `go run ./cmd/stage` rendered the initial guided text shell.
- 2026-05-22: `go run ./cmd/stage > /tmp/stage-guidance.txt` exited and wrote plain guidance without prompting.
- 2026-05-22: `STAGESERVE_NO_TUI=1 go run ./cmd/stage` used text fallback.
- 2026-05-22: `go run ./cmd/stage --notui` and `go run ./cmd/stage --cli` used text fallback.
- 2026-05-22: `go run ./cmd/stage --help` and `go run ./cmd/stage up --help` stayed on direct help paths.
- 2026-05-22: `NO_COLOR=1 go run ./cmd/stage --notui` captured no ANSI styling.

## Phase 3: Guided First-Run And Day-2 Actions

**Goal**: Make guided mode useful for setup, project initialization, running, status, logs, stopping, and recovery.

### First-Run Flow

- [x] T047 Build the first guided TUI shell with status header, decision bar, tool work panel, details view, and persistent footer.
- [ ] T048 Route setup checks through existing onboarding result semantics.
- [ ] T049 Implement guided project config preview using the shared project-env model.
- [ ] T050 Add confirmation before writing `.env.stageserve`.
- [ ] T051 Add inline editing for project name, web folder, and local suffix/address preview.
- [ ] T052 Recompute the planner after config creation and offer the next safe action.
- [ ] T052a Make `stage init` open the guided project-config form by default in interactive terminals.
- [ ] T052b Keep `stage init --notui`, `stage init --cli`, `stage init --json`, and non-TTY `stage init` non-guided and automation-safe.

### Day-2 Flow

- [ ] T053 Route guided run/add actions through existing `up` and `attach` lifecycle semantics.
- [ ] T054 Route guided status through existing status reporting.
- [ ] T055 Route guided logs through existing logs behavior with a clear exit path.
- [ ] T056 Route guided stop/remove actions through `down` and `detach` with explicit confirmation.
- [ ] T057 Route diagnostics/recovery through doctor/readiness/status seams without exposing `doctor` as a first-level peer action.
- [ ] T058 Order recovery actions from least invasive to most invasive and rerun planning after each step.
- [ ] T059 Ensure primary guided labels use user-goal language before command terminology.
- [ ] T060 Ensure running-project default actions are non-destructive.

### Documentation And Evidence

- [ ] T061 Update [README.md](../README.md) first-run path to start with bare `stage` after implementation lands.
- [ ] T062 Update [docs/runtime-contract.md](runtime-contract.md) for guided root behavior, no-TUI controls, text fallback, and direct command behavior.
- [ ] T063 Add or restore an active installer/onboarding doc if still referenced by specs.
- [ ] T064 Record terminal validation evidence in [specs/007-harden-TUI-and-other-interactions/quickstart.md](../specs/007-harden-TUI-and-other-interactions/quickstart.md).
- [ ] T064a Update [install.sh](../install.sh) so interactive install handoff points to bare `stage` after guided routing lands, while non-interactive installs keep explicit commands.
- [ ] T064b Update [.env.stageserve.example](../.env.stageserve.example) comments for guided config creation and active TLS/setup decisions.
- [ ] T064c Update command `Short` and `Long` strings for guided surfaces so first-level help uses plain user-goal language.

### Phase 3 Validation

- [ ] T065 Validate bare `stage` from a project without `.env.stageserve`.
- [ ] T066 Validate cancel-before-write leaves no `.env.stageserve` file.
- [ ] T067 Validate bare `stage` from a configured stopped project.
- [ ] T068 Validate bare `stage` from a running project.
- [ ] T069 Validate logs action exits cleanly.
- [ ] T070 Validate stop/detach/overwrite/recovery confirmations.
- [ ] T071 Validate text fallback parity with the guided situation and default action.
- [ ] T072 Run `go test ./core/guidance ./core/onboarding ./cmd/stage/commands ./core/config ./core/lifecycle ./observability/status ./infra/gateway`.

## Phase 4: Process Hardening

**Goal**: Strengthen validation, documentation architecture, and release workflow.

### Integration And Smoke Tests

- [ ] T073 Add a tagged integration test or script for single-project `up/status/logs/down`.
- [ ] T074 Add a tagged integration test or script for two-project `up/attach/status/detach/down --all`.
- [ ] T075 Add integration coverage for failed `STAGESERVE_POST_UP_COMMAND` rollback.
- [ ] T076 Add integration coverage for `--profile debug` phpMyAdmin activation, or remove the feature from active contract.
- [ ] T077 Add integration coverage for `.dev` TLS, or record an explicit skipped/deferred case.
- [ ] T078 Add release smoke checks for installer asset naming, checksum verification, install handoff, and `stage --version`.

### Architecture Cleanup

- [ ] T078a Consolidate onboarding projection switching in `setup`, `doctor`, and `init` into one shared helper, or record why it is deferred.
- [ ] T078b Replace duplicate setup/doctor/init silent exit-code error types with one shared implementation.
- [ ] T078c Consolidate `.env.stageserve` rendering rules for `stage init`, first `stage up`, first `stage attach`, and guided previews into one shared model.
- [ ] T078d Address the dead `Detach` branch in [infra/compose/compose.go](../infra/compose/compose.go) by honoring the option or removing it from `UpOptions`.
- [ ] T078e Split non-flow helpers out of [core/lifecycle/orchestrator.go](../core/lifecycle/orchestrator.go) where doing so clarifies ownership without changing behavior, or create follow-up tasks for the split.
- [ ] T078f Review [cmd/stage/commands/root.go](../cmd/stage/commands/root.go) after guided routing lands and move any remaining command-specific flags off the root command.

### Docs And Command Reference

- [ ] T079 Expand spec 009 from placeholder into a working documentation plan.
- [ ] T080 Restructure README around install, first run, common workflows, and troubleshooting links.
- [ ] T081 Define a user manual information architecture for setup, project config, daily work, DNS/TLS, multi-project, and recovery.
- [ ] T082 Define a command reference strategy generated from or checked against Cobra help.
- [ ] T083 Move contributor-heavy material to contributor docs.
- [ ] T084 Mark archive-only material clearly and keep historical behavior out of active user docs.

### Release And CI Alignment

- [ ] T085 Align [Makefile](../Makefile) `release` target with [.github/workflows/release.yml](../.github/workflows/release.yml), or remove the placeholder target.
- [ ] T086 Add CI or manual workflow documentation for tagged integration tests.
- [ ] T087 Add docs/command contract checks to CI if they are cheap and stable.
- [ ] T088 Record final real-daemon validation evidence before release.

### Phase 4 Validation

- [ ] T089 Run `go test -short ./...`.
- [ ] T090 Run `go vet ./...`.
- [ ] T091 Run `go build ./...`.
- [ ] T092 Run tagged integration suite on a Docker-capable machine.
- [ ] T093 Run installer/release smoke checks.
- [ ] T094 Run `git diff --check`.
- [ ] T094a Run Markdown/link sanity checks for active docs touched by the roadmap.

## Final Roadmap Gate

- [ ] T095 Confirm active docs, command help, specs, and tests agree on the supported product surface.
- [ ] T096 Confirm bare `stage` guided behavior is validated in TTY, non-TTY, disabled-TUI, and direct-command modes.
- [ ] T097 Confirm single-project and multi-project runtime claims have real-daemon evidence.
- [ ] T098 Confirm release and install paths are documented and tested.
- [ ] T099 Move unresolved items into explicit deferred follow-up specs with rationale.
- [ ] T100 Confirm analysis-report cleanup findings are resolved, deferred, or represented by follow-up tasks.