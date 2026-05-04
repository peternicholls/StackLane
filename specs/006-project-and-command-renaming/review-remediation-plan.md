# PR Review Remediation Plan

PR: #8

## Purpose

Track the concrete follow-up work from the PR review on the StageServe rename branch.

This plan separates:

- real product defects
- real docs/spec contract mismatches
- lower-signal duplicate planning cleanup

The goal is to address the valid review findings without widening scope beyond the rename and runtime-validation work already in flight.

## Review Findings Addressed

### A. Product defect

1. `cmd/stage/commands/init.go`
   - `next_steps` contained `stage up\n` rather than `stage up`.
   - This leaked a newline into text output and JSON output.
   - Status: fixed with text and JSON regression coverage.
   - Fix priority: high.

### B. Contract mismatches between code and docs

1. `core/config/types.go` versus loader behavior
   - `ProjectConfig.StateDir` is documented as `STAGESERVE_STATE_DIR`.
   - The loader honored `STACK_STATE_DIR`.
   - Status: fixed by honoring and tracking `STAGESERVE_STATE_DIR`; `STACK_STATE_DIR` is covered as ignored legacy input.
   - Fix priority: high.

2. `docs/installer-research-report.md`
   - The document referenced `stage setup --recheck docker`.
   - `setup` does not implement `--recheck`.
   - Status: fixed by replacing the example with `stage setup`.
   - Fix priority: medium.

3. Source-layout references to `cmd/stage`
   - Some docs/spec files describe the CLI source as `cmd/stage`.
   - The current source layout builds from `cmd/stage`.
   - Status: the earlier review was based on an older tree; `cmd/stage` is now the active source layout.
   - Fix priority: medium.

### C. Rename-plan wording cleanup

The remaining review comments mostly describe the same underlying issue: some planning artifacts describe no-op renames or self-referential shim behavior.

The affected files include:

- `specs/006-project-and-command-renaming/inventory.md`
- `specs/006-project-and-command-renaming/plan.md`
- `specs/006-project-and-command-renaming/contracts/stageserve-rename-contract.md`
- `specs/005-installer-and-onboarding/stageserve-rename-ticket-list.md`
- `docs/cli-naming-analysis.md`
- `specs/004-workflow-and-lifecycle/research.md`

These should be normalized so they clearly describe:

- old surface -> new surface
- whether a shim refers to `stage` forwarding to `stage`. **No forwarding should be accepted. This is a complete renaming spec, not a compatibility spec.**
- whether a runtime prefix is actually changing or intentionally staying `stage-*`

Fix priority: medium.

## Remediation Phases

### Phase 1: Narrow product fix

Scope:

- remove the trailing newline from `stage init` next-step output

Validation:

- focused test in `cmd/stage/commands`
- confirm text and JSON output both emit `stage up` without an embedded newline

Exit criteria:

- reviewer comment on `init.go` can be resolved directly

### Phase 2: Contract alignment

Scope:

- align the `StateDir` env-var contract across code comments, loader behavior, and operator docs
- remove or rewrite the unsupported `--recheck` example
- correct stale legacy command-root references so the repo consistently uses `cmd/stage`

Validation:

- `go test ./core/config ./cmd/stage/commands`
- targeted grep over active docs/spec files for `--recheck docker` and stale `cmd/stacklane`

Exit criteria:

- no active maintained doc teaches unsupported flags or nonexistent source paths

### Phase 3: Planning artifact cleanup

Scope:

- rewrite no-op rename wording in inventory, plan, contract, and ticket files
- Remove all shim forwarding: **No forwarding should be accepted. This is a complete renaming spec, not a compatibility spec.**
- remove or clarify any `stage-* -> stage-*` statements that read as meaningless renames

Validation:

- manual review of the touched planning/spec files
- targeted grep for self-referential shim wording and no-op rename expressions

Exit criteria:

- spec text is internally consistent and actionable for future work

## Proposed Execution Order

1. Fix the `stage init` newline defect.
2. Resolve the `StateDir` contract mismatch.
3. Clean up unsupported `--recheck` documentation.
4. Correct stale legacy source-layout references.
5. Sweep the planning/spec wording issues in one documentation-only pass.
6. Reply to the PR comments in groups rather than one by one where multiple comments reduce to the same root cause.

## Validation Plan

Focused checks after remediation:

- `go test ./cmd/stage/commands`
- `go test ./core/config`
- targeted grep for:
  - `stage up\\n`
  - `--recheck`
  - stale `cmd/stacklane` references in active rename docs
  - self-referential shim wording
  - no-op rename wording around `stage-*`

PR follow-up:

- push remediation commits on the same branch
- resolve review threads that are fixed directly
- leave explicit responses on any comments that are acknowledged but intentionally not acted on

## Non-Goals

- do not redesign `stage setup`
- do not expand runtime behavior beyond the reviewed issues
- do not reopen archive-only material
- do not rename source directories again unless that work is intentionally added as a scoped follow-up change

## Done Criteria

- the real product defect is fixed
- the code/docs contract mismatches are aligned
- the planning docs no longer contain misleading no-op rename statements
- the valid review comments can be resolved with evidence
