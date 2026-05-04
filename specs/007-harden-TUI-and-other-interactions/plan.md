# Implementation Plan: Guided TUI And Simple-First StageServe Interaction

**Branch**: `007-harden-TUI-and-other-interactions` | **Date**: 2026-05-04 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/007-harden-TUI-and-other-interactions/spec.md`

## Summary

Restore the intended simple-first StageServe experience by making bare `stage` open a guided TUI in interactive terminals. Keep all direct CLI subcommands and JSON automation paths stable. Build the TUI on a terminal-verifiable next-action planner so the interaction layer guides existing setup, init, lifecycle, status, logs, and doctor behavior without duplicating runtime logic.

The implementation should land in thin, reversible slices:

1. planner and contracts
2. root no-args routing
3. projection cleanup
4. guided TUI shell
5. first-run and day-2 actions
6. documentation abstraction cleanup
7. validation

## Technical Context

**Language/Version**: Go 1.26.2  
**Primary Dependencies**: `github.com/spf13/cobra`, `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/huh`, `github.com/charmbracelet/lipgloss`, existing StageServe config/lifecycle/onboarding/status packages  
**Storage**: project `.env.stageserve`, stack `.env.stageserve`, `.stageserve-state` JSON records and generated runtime files  
**Testing**: terminal verification is the primary loop for this spec; focused `go test` runs are supporting checks for deterministic package behavior, JSON purity, and regression safety  
**Target Platform**: macOS primary with text fallback for unsupported or limited terminals  
**Project Type**: CLI/runtime tool with guided terminal UI  
**Performance Goals**: first TUI screen renders within 500 ms excluding explicitly selected long-running checks; direct commands keep existing runtime performance  
**Constraints**: no new user-facing config surface beyond `.env.stageserve`; no direct-command breaking changes; no TUI in non-TTY automation; no Docker concepts in primary user path unless needed for recovery  
**Scale/Scope**: one CLI binary, one guided no-args entrypoint, current setup/init/doctor/lifecycle/status/logs surfaces, single-project context plus existing multi-project state awareness

## Constitution Check

- [x] Ease-of-use impact is documented: bare `stage` becomes the shortest obvious path while direct commands remain available.
- [x] Reliability expectations are explicit: direct subcommands, JSON output, config precedence, and lifecycle rollback stay stable.
- [x] Robustness boundaries are defined: TUI can call existing commands/domains but must not introduce separate runtime state or bypass rollback semantics.
- [x] Documentation surfaces requiring same-change updates are identified: README, runtime contract, installer/onboarding docs, `.env.stageserve.example`, command help, and spec 007 validation.
- [x] Validation covers startup, status/inspection, teardown, failure/recovery, TTY and non-TTY behavior, and direct command compatibility.

## Decision Record

### Guided Root Entry

- Decision: bare `stage` opens the guided TUI only in interactive terminals.
- Rationale: matches the original intention and proven DDEV no-args dashboard pattern while avoiding automation breakage.
- Rejected: keeping bare `stage` as help only, because it preserves the main gap.

### Next-Action Planner

- Decision: create a non-UI planner that determines situation and actions before any Bubble Tea screen renders.
- Rationale: keeps interaction policy terminal-verifiable and reusable across TUI and text fallback.
- Rejected: embedding context logic in the Bubble Tea model, because it would be harder to test and easier to duplicate.

### Verification Style

- Decision: use terminal verification as the primary development loop for spec 007.
- Rationale: this feature is an interaction change. Real `stage` invocations in TTY, non-TTY, disabled-TUI, JSON, and lifecycle contexts catch the most important failures faster than abstract tests alone.
- Supporting checks: narrow package tests remain useful for pure decision tables, JSON parsing, and regression safety, but they do not replace terminal evidence.
- Rejected: strict TDD-first execution, because previous spec runs over-emphasized abstract tests while missing the lived interaction gap.

### TUI Role

- Decision: TUI coordinates existing StageServe actions and shows results; it does not reimplement config, lifecycle, state, or readiness logic.
- Rationale: specs 004 and 005 already hardened those seams.
- Rejected: a separate TUI runtime layer, because it would create divergence.

### Documentation Abstraction

- Decision: primary docs describe StageServe API and user concepts; Docker/gateway details move to advanced/troubleshooting sections.
- Rationale: keeps the simple user path clear while preserving power-user transparency.
- Rejected: removing implementation details entirely, because power users still need them.

## Project Structure

### Documentation

```text
specs/007-harden-TUI-and-other-interactions/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── recovery-plan.md
├── original-intentions-and-decisions.md
├── spec-planning-departures.md
├── current-implementation-review.md
├── contracts/
│   └── guided-tui-contract.md
└── tasks.md
```

### Source Code

```text
cmd/stage/commands/
├── root.go
├── onboarding_mode.go
├── setup.go
├── init.go
├── doctor.go
└── tui.go                 # new or equivalent root TUI adapter

core/
├── guidance/              # new next-action planner package
├── onboarding/
├── config/
├── lifecycle/
└── state/

observability/status/

docs/
├── runtime-contract.md
└── installer-onboarding.md

README.md
.env.stageserve.example
```

**Structure Decision**: add one deep planner package and one thin command/TUI adapter. Do not place planning rules directly inside Cobra command wiring or Bubble Tea screen models.

## Implementation Plan

### Phase 0 - Research And Contract

1. Record guided CLI/TUI patterns and anti-patterns in `research.md`.
2. Define the root interaction and fallback contract in `contracts/guided-tui-contract.md`.
3. Lock the entities in `data-model.md`.
4. Keep direct command and JSON compatibility as explicit acceptance criteria.

### Phase 1 - Planner Foundation

1. Add `core/guidance` package.
2. Implement `TUICapability`, `GuidedContext`, `NextActionPlan`, and `GuidedAction` types.
3. Add a terminal-facing planner inspection path or equivalent debug output that can verify machine-not-ready, missing config, configured stopped, running, drift/error, and non-project directory from real invocations.
4. Keep planner checks cheap by default. Long-running checks should be explicit or injected.
5. Add narrow package tests only where they protect pure decision rules that are hard to exercise reliably from the terminal.

### Phase 2 - Output And Mode Cleanup

1. Route setup, doctor, and init through `onboarding.NewProjector`.
2. Align `stage init` output flags with the spec, including forced TUI if accepted.
3. Remove or correct stale docs for non-existent flags such as `--recheck`.
4. Add terminal JSON parse checks proving JSON output remains pure.

### Phase 3 - Root No-Args Routing

1. Add root no-args detection in `cmd/stage/commands/root.go`.
2. In TTY mode, call the guided TUI adapter.
3. In non-TTY mode, print compact text guidance.
4. Respect `STAGESERVE_NO_TUI=1`, `NO_COLOR=1`, explicit help, and direct subcommands.
5. Add terminal verification commands for each routing path.

### Phase 4 - Guided TUI Shell

1. Add a minimal Bubble Tea model around the planner.
2. Render a context summary, primary action, secondary actions, advanced actions, and help.
3. Add keyboard-first navigation and visible quit/cancel.
4. Use Huh only for bounded forms and confirmations.
5. Keep mutations behind explicit confirmation.

### Phase 5 - Action Execution

1. Wire setup action through existing onboarding runtime/checks.
2. Wire init action through existing project env module with preview.
3. Wire up/down/status/logs/doctor actions through existing command/domain seams.
4. Ensure Ctrl-C and cancel behavior remains coherent during long-running actions.
5. Show result and next recommended action after each action.

### Phase 6 - Documentation And Abstraction Cleanup

1. Update README first-run path to start with bare `stage`.
2. Move Docker/gateway names from primary docs into advanced/troubleshooting sections.
3. Add active `docs/installer-onboarding.md`.
4. Update `.env.stageserve.example` for guided path language.
5. Align command help text with StageServe-first terminology.

### Phase 7 - Validation

1. Run terminal verification scenarios first.
2. Run manual TUI validation in a real TTY.
3. Run non-TTY and JSON validation.
4. Validate startup, status, logs, down, doctor, setup, init, and failure recovery through the guided path.
5. Run focused automated checks after terminal behavior is proven.
6. Record any real-daemon-only gaps in `quickstart.md`.

## Validation Strategy

### Terminal Verification - Primary

- TTY: `stage` from a clean project without `.env.stageserve`.
- TTY: `stage` from a configured stopped project.
- TTY: `stage` from a running project.
- TTY: cancel before init write.
- TTY: cancel during a long-running action where feasible.
- Non-TTY: `stage > out.txt`.
- Disabled TUI: `STAGESERVE_NO_TUI=1 stage`.
- Power commands: `stage setup --json`, `stage up`, `stage status`, `stage down`.
- Parse JSON from `stage setup --json` and `stage doctor --json` with `jq` or an equivalent parser.
- Capture output and exit codes for every scenario in `quickstart.md`.

### Automated Checks - Supporting

- `go test ./core/guidance ./core/onboarding ./cmd/stage/commands`
- `go test ./core/config ./core/lifecycle ./observability/status ./infra/gateway`
- Use automated tests to protect pure planner decisions, JSON schemas, and direct command regressions after terminal behavior has been exercised.

## Risks And Mitigations

| Risk | Why It Matters | Mitigation |
|---|---|---|
| TUI duplicates runtime logic | Behavior diverges from tested commands | Planner owns decisions; existing domains own effects |
| TUI traps automation | CI or scripts hang | TTY detection, no-TUI env/flag, JSON purity tests |
| TUI hides useful failure detail | Operators cannot recover | Show StageServe remediation first, advanced details second |
| TUI gets too ambitious | Large UI delays recovery of original intent | MVP is one landing screen, action list, confirmations, results |
| Docs over-correct and hide implementation | Power users lose inspectability | Keep advanced/troubleshooting sections |
| Long checks slow first screen | No-args feels sluggish | Planner uses cheap checks first and labels deeper checks explicitly |

## Complexity Tracking

No constitution violations require justification.
