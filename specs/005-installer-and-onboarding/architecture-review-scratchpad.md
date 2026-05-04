# Architecture Review Scratch Pad: Spec-005

## Purpose

Working document for the spec-005 architecture review.

Use this to preserve:
- candidate deepening opportunities
- evidence anchors across spec, plan, tasks, and companion artifacts
- open questions
- sequencing for any follow-up fixes

This is a planning scratch pad, not a normative source of truth.

## Source Of Truth

Primary artifacts under review:
- `spec.md`
- `plan.md`
- `tasks.md`

Supporting artifacts:
- `data-model.md`
- `quickstart.md`
- `contracts/cli-onboarding-contract.md`
- `contracts/json-envelope.schema.json`

## Review Vocabulary

Use the architecture skill vocabulary consistently:
- module
- interface
- implementation
- depth
- seam
- adapter
- leverage
- locality

## Candidate Queue

### C1. Command contract spread across too many artifacts

Status: implemented in source-of-truth artifacts

Files:
- `spec.md`
- `plan.md`
- `tasks.md`
- `data-model.md`
- `contracts/cli-onboarding-contract.md`
- `quickstart.md`

Problem summary:
- understanding one command requires bouncing across too many artifacts
- the interface is spread across multiple parallel normative descriptions
- locality is low and review burden is high

Potential direction:
- choose one primary command-contract module
- demote other artifacts into planning, validation, or delivery adapters

Plan of action:
1. Compare `spec.md`, `data-model.md`, and `contracts/cli-onboarding-contract.md` to decide which module should own the command interface.
2. Classify each remaining artifact as either normative interface, planning adapter, or validation adapter.
3. Identify duplicated command-behavior prose that should move behind the chosen seam.
4. Update the scratch pad with the target artifact roles before making any source-of-truth edits.
5. If the choice is clear, then propose a minimal redistribution of responsibility across artifacts without changing locked behavior.

Review outcome:
- `contracts/cli-onboarding-contract.md` is now the primary command-contract module.
- `data-model.md` is now the primary step/result schema module.
- `spec.md` remains the operator-facing interaction contract and requirements module.
- `quickstart.md` remains the implementation and validation adapter.

Follow-up note:
- Re-check whether any remaining command behavior is still redundantly normative after later candidate work.

### C2. Planned onboarding runtime looks like a shallow module swarm

Status: implemented in source-of-truth artifacts

Files:
- `plan.md`
- `tasks.md`

Problem summary:
- runtime is planned as many small peer modules
- deletion test suggests some planned seams mostly redistribute coordination cost
- depth is likely too low around execution and projection

Potential direction:
- deepen around one execution-and-projection module
- keep text, JSON, and TUI as adapters behind it

Plan of action:
1. Inspect `plan.md` and `tasks.md` for the planned runtime modules and group them into execution, projection, and command-entry concerns.
2. Apply the deletion test to each planned module to find which seams are only coordination seams.
3. Identify the smallest deep module that could hide result reduction and output projection behind one interface.
4. Note which planned files should become adapters rather than peer modules.
5. Translate that into task-level restructuring guidance if the deep module shape is stronger than the current swarm.

Review outcome:
- The plan now centers one shared onboarding runtime module in `core/onboarding/runtime.go`.
- Text, JSON, and TUI are now explicitly framed as projection adapters behind that runtime seam.
- Command files are framed as thin adapters for flags, mode selection, and top-level policy.

Follow-up note:
- Preserve runtime depth during code implementation; do not move behavior back into command-local helper swarms.

### C3. Machine readiness behavior split by command instead of domain

Status: implemented in source-of-truth artifacts

Files:
- `spec.md`
- `data-model.md`
- `tasks.md`

Problem summary:
- setup and doctor share readiness rules but tasks encourage command-shaped implementations
- the real seam appears to be machine readiness, not command identity

Potential direction:
- concentrate readiness implementation by environment domain
- treat setup and doctor as adapters with different policy

Plan of action:
1. Compare setup and doctor step inventories in `spec.md` and `data-model.md` to isolate shared machine-readiness rules.
2. Separate shared readiness behavior from command-specific policy such as prompting, write capability, and remediation style.
3. Identify the module seam where machine readiness should live and the two adapters that should sit on top of it.
4. Check whether the current tasks preserve or undermine that seam.
5. Prepare a recommendation for regrouping implementation work by readiness domain instead of by command identity.

Review outcome:
- `setup` and `doctor` are now described as projections over one shared machine-readiness domain.
- The shared domains are Docker, local state, local DNS, port availability, TLS helper presence, and shared gateway health.
- `config.dns_suffix` remains a setup-only policy step ahead of readiness evaluation.
- Command differences are now framed as policy differences: write capability, prompting, privilege, and summary style.

Follow-up note:
- Re-check later candidate work to keep gateway-only and mkcert-only behavior from drifting back into command-shaped implementation tasks.

### C4. Project env ownership crosses a real seam without enough depth

Status: implemented in source-of-truth artifacts

Files:
- `spec.md`
- `tasks.md`

Problem summary:
- explicit init and silent fallback are two adapters touching the same rules
- ownership, overwrite, and allowed-key rules risk leaking across callers

Potential direction:
- deepen one project env ownership module
- route both adapters through the same implementation

Plan of action:
1. Re-read the `stage init` and fallback helper rules in `spec.md` and `tasks.md` as one ownership problem.
2. Identify the shared invariants: allowed keys, overwrite policy, validation, preservation semantics, and file ownership.
3. Distinguish the real seam from the two adapters: explicit init and silent fallback.
4. Define what the deepened ownership module would need to hide from both adapters.
5. Prepare task and plan adjustments so the adapters depend on one implementation instead of each carrying part of the rules.

Review outcome:
- Project-local `.env.stageserve` ownership is now explicit as one shared seam.
- `stage init` is the explicit operator-facing adapter.
- `ensureProjectEnvFile` is the silent fallback adapter.
- Shared invariants are now called out directly: allowed keys, overwrite policy, validation boundary, and preservation semantics.

Follow-up note:
- Keep later runtime and test reshaping focused on this seam so fallback behavior does not regain independent ownership rules.

### C5. Task breakdown encourages low-locality implementation work

Status: implemented in source-of-truth artifacts

Files:
- `plan.md`
- `tasks.md`
- `quickstart.md`

Problem summary:
- several concepts are sliced by file before the concept is whole
- suffix resolution and output projection look especially fragmented

Potential direction:
- regroup work around concept-local modules
- let file structure follow those modules, not drive them

Plan of action:
1. Review the task list for concepts that are currently split across too many files before the behavior is whole.
2. Mark high-friction clusters, especially suffix resolution, output projection, readiness evaluation, and project env ownership.
3. Re-cut those clusters into concept-local work packages that improve locality for implementation and verification.
4. Ensure any regrouping still preserves independent user-story delivery and checkpoints.
5. Reflect the revised work package shape back into the scratch pad before proposing any task edits.

Review outcome:
- Foundational and story tasks now group work around runtime, machine-readiness, project-env ownership, and projection seams.
- Command tasks now read as adapter work layered on top of those deeper modules.
- Independent user-story checkpoints remain intact while implementation locality is improved.

Follow-up note:
- Re-check later code work against the task shape so file-level convenience does not erode concept locality.

### C6. Planned test surface sits too close to command wiring

Status: implemented in source-of-truth artifacts

Files:
- `plan.md`
- `tasks.md`
- `data-model.md`

Problem summary:
- too much behavior appears test-targeted at the command package seam
- command wiring risks becoming the effective interface

Potential direction:
- move most behavioral verification behind deeper modules
- keep command tests focused on flags, wiring, and top-level contract compliance

Plan of action:
1. Inventory the current planned tests in `tasks.md` and classify them as command-wiring tests or behavior tests.
2. Identify behavior tests that currently sit at the command seam only because no deeper module has been named.
3. Map those tests to the deeper modules surfaced in C2, C3, and C4.
4. Keep only top-level contract, flag, and integration checks at the command adapter seam.
5. Use that classification to recommend a more coherent test surface once the deeper modules are chosen.

Review outcome:
- The plan and task guide now state that behavioral verification should live behind `core/onboarding` seams.
- Command-package tests are now framed as adapter, flag, wiring, and contract checks.
- Focused validation now starts with `go test ./core/onboarding` before widening to command/config/lifecycle suites.

Follow-up note:
- Keep future task additions from reintroducing broad behavior tests at the command seam unless they are true integration checks.

## Working Order

Initial order for systematic review:
1. C1 command contract concentration
2. C3 machine readiness concentration
3. C4 project env ownership
4. C2 execution-and-projection runtime shape
5. C6 test surface depth
6. C5 task locality reshaping

## Non-Negotiables

Preserve decisions already locked by the artifacts:
- Bubble Tea + Huh remain the onboarding TUI framework for spec-005
- `install -> setup -> init -> up` remains the operator sequence
- `setup`, `doctor`, and `init` keep shared status and exit semantics
- unsupported platform behavior stays explicit, not silent
- config ownership boundaries remain stack-home vs project-local vs runtime-owned

## Open Questions To Resolve During Review

- Which artifact should become the primary command-contract module?
- Which planned seams are real, and which are only hypothetical?
- Which task groups should be re-cut by concept instead of file?
- Which tests should move behind deeper module interfaces?

## TDD Hardening Outcome

The planning set now carries a concrete tracer-bullet start, not just a general TDD preference.

Initial implementation slices should open with:

1. runtime exit-code precedence
2. runtime overall-status derivation
3. machine-readiness Docker binary failure
4. setup non-interactive missing suffix behavior
5. project-env docroot boundary rejection
6. doctor unhealthy gateway reporting

Reason:
- this sequence opens the deepest shared seams first
- it delays command-adapter expansion until lower-level behavior exists
- it gives implementation a direct red-green-refactor starting line instead of a generic testing reminder

## Ready State

Use this section to mark when the review has moved from exploration to concrete fix planning.

Current state: scratch pad created, candidate set captured, ready for systematic deepening review.

Next milestone: complete candidate-by-candidate review notes for C1-C6, then decide which findings justify source-of-truth edits versus simple recording.
