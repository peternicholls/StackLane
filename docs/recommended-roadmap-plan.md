# Recommended Roadmap Implementation Plan

Date: 2026-05-22
Source: [Project Analysis Report](project-analysis-report.md)

## Purpose

This plan turns the recommended roadmap from the project analysis into an executable delivery sequence. The work is intentionally ordered to reduce contract drift before adding more guided UI behavior.

The target outcome is a StageServe product that has:

1. Direct CLI commands whose behavior, help text, docs, and specs agree.
2. A guided bare `stage` entrypoint built on reusable planning logic.
3. First-run and day-2 guided workflows that call existing lifecycle/onboarding domains instead of duplicating runtime behavior.
4. A dedicated guided-TUI design pass that tightens formatting, colour, and component reuse with the Charm stack once the core flow contract settles.
5. A validation and release process strong enough to prove the Docker, gateway, DNS, state, and TUI claims before release.
6. A cleaner command and onboarding implementation with known duplication either removed or explicitly deferred.

## Guiding Principles

- Contract reconciliation comes before new behavior.
- Direct CLI and JSON automation paths stay stable while guided behavior is added.
- Guided flows reuse existing config, lifecycle, state, status, logs, and onboarding seams.
- Documentation changes land with the implementation changes they describe.
- Real-daemon validation is treated as a deliverable, not as optional notes.
- Historical archive material remains reference-only.

## Scope

### In Scope

- CLI/docs/spec reconciliation for command flags and documented examples.
- Lifecycle rollback hardening around gateway routes and state persistence.
- Decision on `.dev` TLS support: either implement fully or document as partial/deferred.
- `--profile debug` contract: either wire through compose or remove from active surfaces.
- `core/guidance` planner and root no-args routing.
- Text fallback and later guided TUI shell for first-run and day-2 actions.
- `NO_COLOR`, no-TUI, non-TTY, direct-command, and JSON automation safety.
- Command-help, installer handoff, and `.env.stageserve.example` copy updates that follow the chosen contract.
- Cleanup of root flag noise, onboarding projection duplication, project-env rendering drift, and compose/lifecycle helper ambiguity where it blocks the guided work.
- Integration and release smoke validation.
- Documentation information architecture under spec 009.

### Out Of Scope

- Reintroducing legacy Bash wrappers or archived compatibility behavior.
- Adding new runtime stack kinds beyond `20i`.
- Framework-specific app repair, migrations, or presets beyond the existing project-declared post-up hook.
- Replacing direct CLI subcommands with TUI-only behavior.
- Building a desktop GUI.

## Phase 1: Contract Reconciliation

### Goal

Make the implemented CLI, active docs, runtime contract, and existing specs agree before expanding the product surface.

### Key Decisions

- Decide whether to implement or remove/document each mismatched command surface: `status --project`, `logs` positional service names, `setup --recheck`, doctor gateway checks, `setup` mutation/remediation behavior, and final TUI opt-out flags.
- Decide whether `.dev` TLS is complete enough to keep as active user-facing behavior.
- Decide whether `--profile debug` remains a supported runtime flag.

### Implementation Shape

Work in small slices. Each slice should start with the contract decision, then update code, command help, README/runtime docs, and focused tests together.

High-risk slices:

- Late `stage up` rollback cleanup after gateway route reload.
- `.dev` TLS wiring if support remains active.
- Project selector behavior for `status` and `logs` if the documented contract is kept.

### Exit Criteria

- Active docs do not advertise missing command flags or argument shapes.
- `stage <command> --help` agrees with README and runtime contract examples.
- Help output does not show irrelevant root flags on commands that cannot use them, or that cleanup is explicitly deferred with rationale.
- Late lifecycle failures do not leave stale gateway routes for a project that was not saved as attached.
- `.dev` TLS is either fully wired or explicitly marked partial/deferred.
- `--profile debug` is either verified or removed from active docs/help.

## Phase 2: Guided Foundation

### Goal

Build the non-UI decision layer and root command routing needed for a simple-first `stage` entrypoint.

### Implementation Shape

Add a `core/guidance` package that collects cheap context and returns a `NextActionPlan`. The planner should classify the canonical situations from spec 007 without mutating files or starting long-running Docker operations before the first render.

Root command behavior should then route as follows:

- Interactive bare `stage`: guided path.
- Non-TTY bare `stage`: concise text fallback.
- `STAGESERVE_NO_TUI=1`, `--notui`, or `--cli`: text fallback.
- `NO_COLOR=1`: no color styling in fallback or projected output where color would otherwise be emitted.
- `stage --help`: Cobra help.
- Direct subcommands: direct command behavior.

### Exit Criteria

- Planner scenarios are terminal-verifiable without mutation.
- Bare `stage` no longer falls through to help in an interactive terminal.
- Non-TTY and disabled-TUI paths do not hang or prompt.
- Text fallback exposes the same core plan as the TUI entrypoint will use.
- `NO_COLOR=1` is honored for captured output.
- Direct subcommands and JSON output remain automation-safe.

## Phase 3: Guided First-Run And Day-2 Actions

### Goal

Make guided mode useful for real first-run setup and ordinary project management.

### Implementation Shape

Implement guided surfaces in thin adapters over existing domains:

- Project config preview and confirmation uses the onboarding/project-env seam.
- Setup and doctor reporting use onboarding result semantics.
- Up, attach, down, status, and logs use lifecycle/status/logs seams.
- Recovery actions are ordered from least invasive to most invasive and rerun planning after each action.

The guided experience should use plain goal language first and direct command names only as command equivalents or advanced details.

### Exit Criteria

- A project without `.env.stageserve` can be initialized through guided preview and confirmation.
- `stage init` uses the guided project-config form by default in interactive terminals, while non-guided flags keep automation behavior.
- A configured project can be run through guided mode after shutdown cleanup removes runtime-owned state.
- A running project can be inspected without making destructive actions the default.
- Stop, overwrite, and state-changing recovery paths require explicit confirmation.
- Terminal verification evidence is recorded in spec 007 quickstart.

## Phase 4: Guided TUI Design Polish

### Goal

Use a dedicated visual pass to make the guided shell feel intentional and cohesive once the interaction model is stable.

### Implementation Shape

- Audit the guided TUI against the StageServe terminal design system and current Bubble Tea/Lip Gloss/Huh guidance.
- Extract or refine reusable Charm-based components for headers, facts, decision rows, confirmations, details, and footer hints.
- Tighten spacing, formatting, and semantic colour while preserving text fallback parity and `NO_COLOR` behavior.
- Record design review evidence and any explicitly deferred visual work alongside the guided-flow validation.

### Exit Criteria

- Guided screens use consistent hierarchy, spacing, and semantic colour in a real terminal.
- Visual polish remains layered on top of the existing planner and command/domain seams rather than reintroducing behavioral duplication.
- Text fallback still matches the same screen truth even after the TTY presentation improves.
- Design review notes identify any remaining intentional gaps instead of leaving them implicit.

## Phase 5: Process Hardening

### Goal

Make the project easier to validate, release, and navigate after the product surface is reconciled.

### Implementation Shape

- Add tagged Docker integration tests or scripted smoke tests for real runtime behavior.
- Add command/docs contract checks for common drift cases.
- Consolidate or explicitly defer cleanup for root shared flags, onboarding projection helpers, project-env rendering, compose detach semantics, and non-flow lifecycle helpers.
- Plan the future split between the shipped stack catalog and a user-owned custom stack registry before adding non-20i runtime definitions.
- Declutter docs under spec 009 into user, command-reference, contributor, and archive-oriented surfaces.
- Align `Makefile`, CI, installer, and release workflow so maintainers have one obvious release path.

### Exit Criteria

- Integration validation covers single-project lifecycle, multi-project attachment, teardown, rollback, profile behavior, and `.dev` TLS or its documented deferral.
- README is a user-facing first door, not a contributor-heavy index.
- Command reference is easy to regenerate or verify against Cobra help.
- Known code-cleanup findings from the analysis report are either resolved or tracked as explicit follow-up work with rationale.
- The roadmap names a concrete ownership and storage plan for any future user-defined stack registry instead of leaving SQLite-versus-files undecided.
- Release workflow is documented and executable without contradictory Makefile targets.

## Dependencies

| Dependency | Blocks |
|---|---|
| Contract decisions for flags/selectors | Docs reconciliation, guided command equivalents |
| Rollback route cleanup | Reliable guided `up` and recovery flows |
| `.dev` TLS decision | Guided URL previews, setup/mkcert copy, integration tests |
| `core/guidance` planner | Root no-args behavior, TUI shell, text fallback |
| Text fallback | Full TUI confidence and non-TTY automation safety |
| Command-help cleanup | User-facing docs, guided command equivalents, and spec 009 command reference |
| Project-env rendering model | Guided init preview and first-run `.env.stageserve` behavior |
| Integration smoke harness | Release readiness claims |

## Validation Strategy

Use the narrowest check that proves the changed slice, then broaden at phase gates.

### Focused Checks

- `go test ./cmd/stage/commands`
- `go test ./core/config`
- `go test ./core/lifecycle`
- `go test ./core/onboarding`
- `go test ./observability/status`
- `go test ./infra/gateway`

### Broad Checks

- `go test -short ./...`
- `go vet ./...`
- `go build ./...`
- `git diff --check`

### Manual Or Integration Checks

- Bare `stage` in a real TTY.
- Non-TTY bare `stage > out.txt`.
- `STAGESERVE_NO_TUI=1 stage`.
- `NO_COLOR=1 stage --notui` or equivalent captured output.
- `stage setup --json` and `stage doctor --json` parsed as JSON.
- Single-project `up/status/logs/down`.
- Two-project `up/attach/status/project-down/down --all`.
- Failed post-up hook rollback.
- `.dev` TLS route or documented skipped case.
- `stage up --profile debug` phpMyAdmin activation or documented removal.

## Risks And Mitigations

| Risk | Mitigation |
|---|---|
| Guided TUI duplicates lifecycle/config behavior | Keep planner side-effect free and call existing domains for effects. |
| Docs drift while code changes | Update docs, command help, and tests in the same slice. |
| Root TUI breaks automation | Keep non-TTY, JSON, direct subcommand, and no-TUI checks in every phase gate. |
| Real Docker validation becomes too slow for daily work | Use tagged integration tests and run them before release or on demand. |
| `.dev` TLS grows scope | Make an explicit support/defer decision in phase 1 before guided URL work. |

## Definition Of Done

The roadmap is complete when direct command contracts are reconciled, bare `stage` opens a guided path in interactive terminals, first-run/day-2 guided workflows are validated, docs are decluttered around the active product, and real-daemon validation evidence exists for the runtime claims that cannot be proven by unit tests alone.