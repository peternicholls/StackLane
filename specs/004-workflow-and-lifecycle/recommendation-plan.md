# Recommendation Plan: Review Follow-Up For Spec 004

Date: 2026-04-25
Source: [implementation-review.md](./implementation-review.md)

## Objective

Turn the implementation review into a small, controlled follow-up pass that closes the remaining gaps without reopening the core runtime work.

This plan assumes the code-level behavior introduced for spec 004 is largely correct and that the remaining work is primarily:

- documentation framing
- operator-surface cleanup
- task bookkeeping

## Working Assumptions

- `.env.example` stays deleted. It is considered redundant and is not being restored for backward compatibility.
- StageServe is the central authority for the 20i-style local stack model.
- Projects are expected to tap into that shared model through project-local inputs rather than define their own stack shape.
- Shared gateway behavior is a StageServe-managed internal subsystem for normal operators.
- Explicit gateway resource names remain useful, but only in advanced contract or troubleshooting material.

## Scope

### In scope

- Remove stale references to `.env.example`
- Reframe README around the central-authority stack model
- Reduce gateway implementation detail in operator-facing docs
- Reconcile `tasks.md` with the work already completed

### Out of scope

- Reopening the loader or lifecycle design
- Renaming internal shared resources again
- Replacing the current test strategy
- Adding compatibility shims for old config filenames

## Workstreams

### Workstream 1: Eliminate stale `.env.example` references

Goal:

Make the docs and tasks consistent with the intended contract: `.env.stageserve.example` remains, `.stackenv.example` is deleted, `.env.example` is not part of the supported surface.

Target files:

- [README.md](../../../README.md)
- [specs/004-workflow-and-lifecycle/tasks.md](./tasks.md)
- Any remaining operator-facing references discovered during grep

Actions:

1. Remove wording that tells operators to copy or expect `.env.example`.
2. Reword task items T011 and T027 so they reflect the actual intended deliverable.
3. Run a final grep for `.env.example` references and review each remaining hit for whether it is historical, intentional, or stale.

Acceptance criteria:

- No operator-facing doc incorrectly advertises `.env.example`.
- `tasks.md` no longer describes `.env.example` as an active deliverable unless intentionally retained.

### Workstream 2: Reframe README around the central stack authority model

Goal:

Make the README communicate the product model clearly: StageServe defines the shared 20i-style stack contract, and projects plug into it.

Target files:

- [README.md](../../../README.md)
- Optional cross-check against [specs/004-workflow-and-lifecycle/quickstart.md](./quickstart.md)

Actions:

1. Add a short statement near the early conceptual sections or command semantics explaining that StageServe centrally owns the stack shape.
2. Make clear that project-local configuration customizes a project within that shared contract instead of redefining the stack.
3. Add the deployed-copy validation note in generic terms:
   the live StageServe installation on `PATH` is the authoritative runtime surface.
4. Mention `$HOME/docker/20i-stack` only as a known local example if that helps this repository's operators, not as a universal product rule.

Acceptance criteria:

- README explains the central-authority model in plain operator language.
- README no longer depends on readers inferring the difference between repo checkout and live installed copy.
- README and quickstart tell a consistent validation story.

### Workstream 3: Hide gateway internals from normal operator docs

Goal:

Keep the gateway implementation contract available for advanced use, while preventing ordinary operators from feeling responsible for managing gateway resources directly.

Target files:

- [README.md](../../../README.md)
- [docs/runtime-contract.md](../../../docs/runtime-contract.md)
- Optional cross-check against [specs/004-workflow-and-lifecycle/contracts/workflow-lifecycle-contract.md](./contracts/workflow-lifecycle-contract.md)

Actions:

1. Rewrite operator-facing wording so it says StageServe ensures shared routing is running, reuses it when present, and repairs it when missing.
2. Move or confine explicit names like `stage-shared` and `stage-gateway` to advanced contract or troubleshooting depth where needed.
3. Preserve correctness of the internal contract, but stop teaching it as a first-class operator workflow.

Acceptance criteria:

- Main operator docs describe behavior, not gateway internals.
- Advanced/internal names remain discoverable in lower-level material.
- The contract doc still remains technically accurate.

### Workstream 4: Repair spec bookkeeping

Goal:

Make the task ledger trustworthy again.

Target files:

- [specs/004-workflow-and-lifecycle/tasks.md](./tasks.md)

Actions:

1. Reconcile actual completed work against T008, T009, T011, T025, T027, T030, T030a, and T030b.
2. Mark tasks `[X]` only where the repository state now fully matches the intended deliverable.
3. Leave follow-up work open if the doc cleanup above has not yet landed.

Acceptance criteria:

- `tasks.md` reflects actual repository state.
- A reviewer can tell what remains without rereading the full implementation history.

## Recommended Execution Order

1. Workstream 1
2. Workstream 2
3. Workstream 3
4. Workstream 4

Reasoning:

- Remove stale file references first so later doc edits do not perpetuate them.
- Reframe the top-level README next because it sets operator expectations for the rest of the docs.
- Then reduce gateway implementation leakage while preserving contract accuracy.
- Update task bookkeeping last, once the repository state is stable.

## Validation Plan

After the follow-up edits:

1. Run a targeted grep for stale references:
   - `.env.example`
   - `.stackenv`
   - `stage-shared`
   - `stage-gateway`
2. Manually review each remaining hit to confirm it is either:
   - intentional advanced/internal documentation
   - historical artifact explicitly marked as such
   - actual operator-facing content that still needs revision
3. Rebuild and re-run the lightweight validation already used for this branch:
   - `go test ./...`
   - `go build -o stage-bin ./cmd/stage`
4. Spot-check operator-facing help and docs for wording drift.

## Done Definition

This recommendation plan is complete when:

- operator docs no longer advertise `.env.example`
- README explains the central-authority stack model clearly
- gateway internals are de-emphasized in operator-facing surfaces
- `tasks.md` accurately reflects completed and remaining work
- existing runtime tests continue to pass unchanged
