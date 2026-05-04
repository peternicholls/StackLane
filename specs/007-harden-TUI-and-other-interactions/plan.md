# Implementation Plan: Guided TUI And Simple-First StageServe Interaction

**Branch**: `007-harden-TUI-and-other-interactions` | **Date**: 2026-05-04 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/007-harden-TUI-and-other-interactions/spec.md`

## Summary

Restore the intended simple-first StageServe experience by making bare `stage` open a guided TUI in interactive terminals. Keep the final direct CLI and JSON automation contracts explicit. Build the TUI on a terminal-verifiable next-action planner so the interaction layer guides existing setup, init, lifecycle, status, logs, and doctor behavior without duplicating runtime logic.

## Override Markers

These plan assumptions are locked for spec 007 and overrule earlier priors:

- Root routing is no longer a help-only surface in TTY mode; bare `stage` is the guided entrypoint.
- The plan targets default TUI plus `--notui` / `--cli` opt-outs; it does not preserve `--tui` as part of the final contract.
- `stage init` default interactive behavior is decided; this plan does not leave that question open for implementation-time debate.
- Plain-language easy-mode labels are mandatory; implementation should not fall back to command terminology as the first-level UI model.
- Validation is terminal-first for this spec run; package tests support the work but do not define done on their own.

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
**Constraints**: no new user-facing config surface beyond `.env.stageserve`; final direct-command behavior must be explicit and automation-safe; no TUI in non-TTY automation; no Docker concepts in primary user path unless needed for recovery; easy-mode copy must use plain user-goal language before command or runtime terminology
**Scale/Scope**: one CLI binary, one guided no-args entrypoint, current setup/init/doctor/lifecycle/status/logs surfaces, single-project context plus existing multi-project state awareness

## Constitution Check

- [x] Ease-of-use impact is documented: bare `stage` becomes the shortest obvious path while direct commands remain available.
- [x] Reliability expectations are explicit: direct subcommands bypass the root TUI, JSON output remains pure, config precedence stays deterministic, and lifecycle rollback stays governed by existing lifecycle semantics.
- [x] Robustness boundaries are defined: TUI can call existing commands/domains but must not introduce separate runtime state or bypass rollback semantics.
- [x] Documentation surfaces requiring same-change updates are identified: README, runtime contract, installer/onboarding docs, `.env.stageserve.example`, command help, and spec 007 validation.
- [x] Validation covers startup, status/inspection, teardown, failure/recovery, TTY and non-TTY behavior, and direct command behavior.

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

### Easy-Mode Language

- Decision: the guided TUI, text fallback, installer handoff, and first-run docs use plain goal labels such as "run this project", "stop this project", "add this project to StageServe", and "remove this project from StageServe".
- Rationale: command words such as attach and detach are precise for lifecycle/state work, but they do not match how a front-end developer or hobbyist normally describes what they want to do.
- Rejected: using direct command names as first-level labels, because that makes easy mode teach the CLI map instead of guiding the user's next action.

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
4. Keep direct command behavior and JSON purity as explicit acceptance criteria.

### Phase 1 - Planner Foundation

1. Add `core/guidance` package.
2. Implement `TUICapability`, `GuidedContext`, `NextActionPlan`, and `GuidedAction` types.
3. Add a terminal-facing planner inspection path or equivalent debug output that can verify machine-not-ready, missing config, configured stopped, running, drift/error, and non-project directory from real invocations.
4. Keep planner checks cheap by default. Long-running checks should be explicit or injected.
5. Add narrow package tests only where they protect pure decision rules that are hard to exercise reliably from the terminal.

### Phase 2 - Output And Mode Cleanup

1. Route setup, doctor, and init through `onboarding.NewProjector`.
2. Implement the spec decision that `stage init` opens a guided project-config form by default in interactive terminals, with `--notui` and `--cli` as equivalent opt-outs.
3. Remove or correct stale docs for non-existent flags such as `--recheck`.
4. Add terminal JSON parse checks proving JSON output remains pure.

### Phase 3 - Root No-Args Routing

1. Add root no-args detection in `cmd/stage/commands/root.go`.
2. In TTY mode, call the guided TUI adapter.
3. In non-TTY mode, print compact text guidance.
4. Respect `--notui`, `--cli`, `STAGESERVE_NO_TUI=1`, `NO_COLOR=1`, explicit help, and direct subcommands.
5. Add terminal verification commands for each routing path.

### Phase 4 - Guided TUI Shell

1. Add a minimal Bubble Tea model around the planner.
2. Render a context summary, primary action, secondary actions, advanced actions, and help.
3. Add keyboard-first navigation and visible quit/cancel.
4. Use Huh only for bounded forms and confirmations.
5. Keep mutations behind explicit confirmation.
6. Apply the plain-language label map from the contract before rendering either TUI or text fallback.

### Phase 5 - Action Execution

1. Wire setup action through existing onboarding runtime/checks.
2. Wire init action through existing project env module with preview.
3. Wire up/attach/down/detach/status/logs/doctor actions through existing command/domain seams.
4. Ensure Ctrl-C and cancel behavior remains coherent during long-running actions.
5. Show result and next recommended action after each action.

### Phase 6 - Documentation And Abstraction Cleanup

1. Update README first-run path to start with bare `stage`.
2. Move Docker/gateway names from primary docs into advanced/troubleshooting sections.
3. Add active `docs/installer-onboarding.md`.
4. Update `.env.stageserve.example` for guided path language.
5. Align command help text with StageServe-first terminology.
6. Update `install.sh` so interactive handoff points to bare `stage` after the guided entrypoint lands.
7. Review first-level copy for jargon and rename easy-mode labels while preserving direct command equivalents.

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
- Power commands: `stage setup --json`, `stage up`, `stage status`, `stage logs`, `stage down`, `stage attach`, `stage detach`.
- Direct command verification includes both help-path and real-behavior checks for `stage up`, `stage status`, `stage logs`, `stage down`, `stage attach`, and `stage detach`, with any daemon-only gap recorded explicitly.
- Plain-language verification confirms easy-mode labels do not expose attach/detach or runtime terms before "show commands" or advanced/troubleshooting views.
- Architectural verification confirms the planner and renderers reuse existing lifecycle and config-precedence seams rather than duplicating them.
- Parse JSON from `stage setup --json` and `stage doctor --json` with `jq` or an equivalent parser.
- Measure first-screen render time, keyboard-only operation, and text fallback parity.
- Verify installer handoff output.
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
| Easy-mode labels mirror internal command names | Non-specialist users must learn lifecycle jargon before they can act | Use user-goal labels first and expose command equivalents through show-commands |
| TUI gets too ambitious | Large UI delays recovery of original intent | MVP is one landing screen, action list, confirmations, results |
| Docs over-correct and hide implementation | Power users lose inspectability | Keep advanced/troubleshooting sections |
| Long checks slow first screen | No-args feels sluggish | Planner uses cheap checks first and labels deeper checks explicitly |

## Complexity Tracking

No constitution violations require justification.
