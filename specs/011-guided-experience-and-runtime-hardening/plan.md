# Implementation Plan: Guided Experience And Runtime Hardening

**Branch**: `011-guided-experience-and-runtime-hardening` | **Date**: 2026-05-30 | **Spec**: [spec.md](./spec.md)

## Summary

Turn the code-review and field-feedback follow-up into a repair-first hardening spec that closes runtime error gaps, fixes broken installed-binary paths, makes direct and guided output feel like one product, expands the running-project experience, and uses more of the Charm stack for confirmation and long-running feedback.

## Locked Planning Decisions

- 011 is a repair spec first. Code-review correctness and operator-safety issues must be fixed before broader UX expansion is treated as done.
- This work builds on spec 007. It does not reopen bare `stage` routing, easy-mode copy, or the existing opt-out contract.
- T001a uses a bundled release/install artifact that ships the binary and runtime stack assets together. A future binary-embedded asset model is deferred and tracked in [embedded-runtime-assets-plan.md](./embedded-runtime-assets-plan.md).
- Missing runtime assets are handled as product-level StageServe errors, not as installer-only assumptions.
- Shared colour and hierarchy tokens come from one guidance style module used by TUI and non-TUI paths.
- Running-project action expansion is scoped to high-value day-2 actions first: open browser, logs, stop, restart, and advanced command equivalents.
- Confirmation and progress improvements use Bubble Tea and Bubbles components where they strengthen the product without creating a second runtime layer.
- Security and seam cleanup work lands alongside UX changes when it directly supports correctness or operator safety.

## Technical Context

**Language/Version**: Go 1.26.2  
**Primary Dependencies**: `github.com/spf13/cobra`, `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/bubbles/spinner`, `github.com/charmbracelet/bubbles/viewport`, `github.com/charmbracelet/bubbles/help`, `github.com/charmbracelet/lipgloss`, existing StageServe config/lifecycle/onboarding/status packages  
**Primary Packages**: `core/guidance`, `cmd/stage/commands`, `core/lifecycle`, `infra/compose`, `core/config`  
**Validation Focus**: focused Go tests plus manual terminal checks for missing assets, failure recovery, running-project actions, confirmation prominence, and long-running feedback  
**Primary Constraint**: keep the solution contract-driven and incremental; do not create a second implementation of lifecycle behavior inside the TUI

## Decision Record

### Product-Level Failure Translation First

- Decision: fix runtime failures at the StageServe boundary before deepening visual polish.
- Rationale: if `stage up` still leaks raw missing-file or registry errors, more TUI chrome does not solve the product gap.
- Rejected: treating installer drift and lifecycle error translation as separate future cleanup.

### Shared Style Registry

- Decision: extract semantic style tokens into a reusable guidance module and consume them everywhere.
- Rationale: the current style language is correct but trapped in `shell.go`.
- Rejected: maintaining separate style choices for TUI and direct output.

### Async Guided Actions

- Decision: move long-running guided actions onto `tea.Cmd`-driven async execution with visible loading state.
- Rationale: this keeps Bubble Tea responsive and enables spinner/progress work without duplicating lifecycle logic.
- Rejected: continuing to block the event loop during lifecycle calls.

### Primary Actions First, Advanced Shelf Second

- Decision: add a small high-value day-2 action set before broader framework-specific utilities.
- Rationale: open URL, logs, restart, and stop cover the largest real gap while keeping the planner and runtime summary manageable.
- Rejected: trying to surface every possible project or framework tool in the first follow-up pass.

### Seam Cleanup As Enabler Work

- Decision: include the review's seam fixes in the same spec so later UX work lands on stable primitives.
- Rationale: duplicated env parsing, dead compose fields, bypassed state loading, and ignored registry errors are direct blockers for reliable follow-up.
- Rejected: postponing all internal cleanup until after more TUI work.

## Delivery Split

### Must Fix Before Merge

The current branch should not merge until the following outcome set is complete:

1. Supported installs provision the required runtime compose assets, and drift still fails early with StageServe-native remedies.
2. Bare `stage` can reach `machine_not_ready` from guided root routing.
3. `Attach()` no longer ignores registry read failures, and touched lifecycle errors carry remedies.
4. Sensitive password handling removes command-line password entry and uses env/config-based input only.
5. Docs/help for the merge-gate repair work land with that work.
6. Focused Go tests and essential terminal validation cover only the merge-gate repair set.

This maps to tasks `T001a` through `T001c`, `T001` through `T007`, `T011`, `T041`, `T042`, `T043`, and `T044` in [tasks.md](./tasks.md).

### Follow-up After Merge

The remaining work in this spec remains planned and valuable, but it does not block merging the current branch once the merge-gate work above is complete. That follow-up scope includes:

1. Remaining seam cleanup such as dead compose options, state-store seam cleanup, env helper deduplication, and partial-failure/concurrency hardening.
2. Shared style extraction and direct-output consistency.
3. Guided failure-recovery surfaces for direct commands.
4. Running-project action expansion.
5. Confirmation modal, spinner/progress work, and broader Bubble Tea polish.
6. Broader docs/help alignment for recovery and day-2 features after those features land.

## Planned Workstreams

### Workstream 1 - Runtime Hardening And Operator Safety

1. Repair install/release provisioning of runtime compose assets through a bundled artifact and add pre-flight runtime asset checks.
2. Route machine readiness into guided root context collection.
3. Fix attach registry error handling and empty remedy strings.
4. Remove dead compose options and unsafe secret-handling defaults.
5. Improve `DownAll()` and route-refresh behavior around partial and concurrent failures.

### Workstream 2 - Shared Presentation Layer

1. Extract shared style tokens from `shell.go`.
2. Add reusable text/direct error rendering helpers.
3. Apply the shared presentation layer to guidance text fallback and direct command error paths.
4. Validate `NO_COLOR` parity and severity consistency.

### Workstream 3 - Recovery Surfaces

1. Add typed failure situations or structured failure-plan inputs.
2. Route `stage up` failures into recovery surfaces.
3. Keep non-TTY recovery output equivalent in content even when not interactive.
4. Add focused tests for the new planner/recovery paths.

### Workstream 4 - Running-Project Action Expansion

1. Extend runtime summary with service metadata.
2. Add `open browser`, `view logs`, and `restart service` actions.
3. Add an advanced-actions path for direct commands and later extensibility.
4. Keep action ordering safe: inspect first, mutate later.

### Workstream 5 - Confirmation And Long-Running Feedback

1. Replace inline confirmation with an elevated modal treatment.
2. Add async spinner state for guided actions.
3. Define cancel/quit semantics for async actions so context cancellation and rollback expectations stay coherent.
4. Add progress reporting where the orchestrator already has meaningful step boundaries.
5. Add help/keymap affordances and narrow-width checks for the new surfaces.

### Workstream 6 - Final Cleanup, Docs, And Validation

1. Land small hygiene fixes while touching the affected files.
2. Update docs/help for env-first secret guidance and richer recovery/day-2 actions.
3. Run focused validation and terminal checks.
4. Record deferred items that should remain outside this follow-up spec.

## Proposed Sequence

### Phase 0 - Crosswalk And Scope Lock

1. Translate CR-01 through CR-14 and FR-01 through FR-06 into one consolidated backlog.
2. Lock what is in scope for this spec versus deferred follow-up.
3. Preserve spec 007 contract assumptions while allowing implementation cleanup.

### Phase 1 - Runtime Hardening

1. Implement bundled release/install provisioning of runtime compose assets.
2. Add compose asset pre-flight checks and remedies.
3. Inject machine readiness into guided context collection.
4. Fix attach registry handling and remedy completeness.
5. Remove dead compose detach behavior.
6. Deduplicate env-truthy parsing and clean state-store access seams.
7. Improve partial-failure and route-refresh behavior.

If only the merge-gate repair set is being delivered on the current branch, stop after Phase 1 and then run the merge-gate docs/help and validation tasks from [tasks.md](./tasks.md).

### Phase 2 - Shared Output Consistency

1. Extract guidance styles into a shared module.
2. Update the TUI to consume the extracted styles.
3. Apply the same styles and hierarchy to text fallback and direct command errors.
4. Add focused parity checks for `NO_COLOR` and severity rendering.

### Phase 3 - Failure Recovery Paths

1. Add recovery-plan inputs or new failure situations.
2. Implement `stage up` recovery entrypoints for TTY and non-TTY flows.
3. Ensure every surfaced error includes next-step guidance and direct command equivalents.
4. Add tests for machine-not-ready and start-failed paths.

### Phase 4 - Running-Project Expansion

1. Extend runtime summary/service inspection.
2. Add browser, logs, restart, and advanced-actions handlers.
3. Keep the default action non-destructive.
4. Validate these flows in a real terminal.

### Phase 5 - Confirmation And Progress Polish

1. Add bordered modal confirmations.
2. Add spinner-based async action handling.
3. Add progress reporting where feasible.
4. Add help/footer polish for the new interaction states.

### Phase 6 - Final Alignment

1. Update docs and help text tied to the new behavior.
2. Run focused Go tests.
3. Run terminal validation for the critical paths.
4. Record any deferred work such as broader framework-specific utility actions.

The deferred long-term runtime-asset direction is captured in [embedded-runtime-assets-plan.md](./embedded-runtime-assets-plan.md).

## Validation Strategy

### Focused Automated Validation

- `go test ./core/guidance ./cmd/stage/commands ./core/lifecycle ./infra/compose ./core/config`
- targeted tests for missing compose assets, attach registry failure, `machine_not_ready` reachability, compose detach option cleanup, recovery rendering, and `NO_COLOR` parity

### Terminal Validation

- installer or supported local-install validation that runtime assets are provisioned in the default stack-home layout
- manual `stage up` with missing assets in `STACK_HOME`
- bare `stage` on a machine-not-ready path
- running-project guided shell with open browser, logs, and stop actions
- destructive confirmation visibility and cancel path
- long-running guided action feedback in a real TTY
- `NO_COLOR=1 stage` and non-TTY fallback checks

### Safety Checks

- direct commands remain non-interactive when invoked directly in non-TTY contexts
- JSON-output commands stay free of styled guidance
- gateway/shared-route behavior is still correct after attach/down-all/up changes

## Risks And Mitigations

| Risk | Why It Matters | Mitigation |
|---|---|---|
| Scope balloons from hardening into a full shell redesign | Work slows and correctness fixes get delayed | Keep primary action set narrow and defer framework-specific extras |
| Async Bubble Tea integration destabilizes current guided flows | Regressions in core UX | Introduce async execution behind existing action handlers and add focused update-loop tests |
| Shared styles leak TUI assumptions into non-TTY output | Plain output becomes noisy or fragile | Keep the shared style registry semantic, with `NO_COLOR` and plain-text-safe fallbacks |
| Asset-path fixes solve only the local symptom | Future stack layout changes regress again | Check for required assets via config-resolved paths and fail early with remedies |
| Partial-failure fixes widen lifecycle churn | Risk of unrelated regressions | Keep lifecycle changes local, test-driven, and limited to the reviewed seams |

## Deliverables

- new follow-up spec package in this directory
- runtime-hardening plan mapped from the review findings
- implementation task list that combines review gaps and field feedback
- focused validation matrix for code and terminal behavior