# Implementation Plan: StageServe Rename And stage Command Cutover

**Branch**: `master` | **Date**: 2026-05-04 | **Workplan**: `specs/006-project-and-command-renaming/tickets.md`
**Input**: Gated rename workplan from `specs/006-project-and-command-renaming/tickets.md`

## Summary

Finalize the StageServe identity and canonical `stage` CLI across the remaining active internal and documentation surfaces so the spec closes with no active legacy references left in code or docs. Execute the rename as a gated hybrid plan: freeze the final end state first, inventory every active rename surface, rehearse local install and release mechanics, cut over the public command and docs in narrow slices, then finish with the internal rename sweep and a zero-active-reference verification gate.

This phase is intentionally dry-run-first. No public-facing cutover step should happen before local binary/install rehearsal, docs/spec rehearsal, CI/release rehearsal, and rollback rehearsal are all green.

## Technical Context

**Language/Version**: Go 1.26.2 for the active CLI, shell installer for binary distribution, Markdown docs/specs for operator contract
**Primary Dependencies**: `github.com/spf13/cobra`, current release/install flow in `install.sh`, GitHub Releases distribution, existing docs/spec artifacts under `README.md`, `docs/`, and `specs/`
**Storage**: local files under `.env.stageserve`, `.stageserve-state`, release assets and checksums, shell completion scripts
**Testing**: focused `go test` for `cmd/stage/commands`, `core/config`, and `core/lifecycle`, plus dry-run and manual operator validation for install, PATH, completion, CI, release, and rollback surfaces
**Target Platform**: macOS primary, with CI/release surfaces validated wherever the installer currently supports
**Project Type**: CLI/runtime tool with installer, documentation, and release pipeline
**Performance Goals**: preserve existing command behavior and output contracts; do not add hidden migration work or widen runtime behavior during rename
**Constraints**: phased execution is allowed, but no active legacy references may remain by closeout except clearly historical or archival references; preserve JSON/stdout purity during any transition step; avoid combining rename slices that hide rollback boundaries
**Scale/Scope**: one product rename, one command cutover, the remaining active internal naming surfaces, all maintained user-facing docs/specs, installer/release/CI surfaces, and explicit clean/dirty-machine verification

## Constitution Check

- [x] Ease-of-use impact is explicit: the rename shortens the CLI and changes the public product name, but the operator journey must remain copy-pasteable and predictable.
- [x] Reliability expectations are explicit: phased migration is allowed, but the end state is a repository with no active legacy references outside historical or archival material; output contracts remain intact, and no compatibility shim survives closeout.
- [x] Robustness boundaries are explicit: this is a naming migration, not a runtime behavior redesign.
- [x] Same-change documentation scope is explicit: active docs, active normative specs, install/release guidance, and operator-facing help text all change together.
- [x] Validation covers install, help, version, clean/dirty-machine command behavior, CI/release rehearsal, and rollback rehearsal.

## Project Structure

### Documentation (this feature)

```text
specs/006-project-and-command-renaming/
├── inventory.md
├── plan.md
├── research.md
├── runbook.md
├── tickets.md
└── contracts/
    └── stageserve-rename-contract.md
```

### Source Code (repository root)

```text
README.md
install.sh
Makefile
stage
stage-bin
cmd/
  stage/
docs/
specs/
scripts/
```

**Structure Decision**: keep this feature centered on the existing spec folder plus the active command/install/docs/release surfaces. Do not introduce a second competing planning authority outside `specs/006-project-and-command-renaming/`.

## Decision Record

### Planning Style

- Decision: use a gated hybrid plan.
- Rationale: a pure backlog is too easy to execute out of order, while a pure prose plan is too coarse for careful execution and evidence capture.

### Temporary Staging Boundary

- Decision: literal legacy-named internal surfaces may remain temporarily during early dry-run phases only, but they are not acceptable as the final state for this spec.
- Rationale: phasing reduces cutover risk, but leaving active internal legacy references behind would contradict the required end state.
- Exception: `stage-*` remains the accepted final runtime prefix.

### Dry-Run Requirement

- Decision: every irreversible surface must pass a rehearsal before cutover.
- Rationale: installer, PATH, completion, CI cache, asset naming, and rollback failures are easier to catch before public mutation than after release publication.

## Implementation Plan

### Phase 0 - Freeze The Contract

1. Write and review the rename contract in `contracts/stageserve-rename-contract.md`.
2. Record the owner matrix, abort criteria, and rollback ownership in `runbook.md`.
3. Refuse implementation work until the final no-active-legacy end state and any temporary staging exceptions are both explicit.

### Phase 1 - Inventory Before Edit

1. Record active `stage` surfaces in `research.md` and `inventory.md`.
2. Classify each surface as rename now, rename later in this spec, mark legacy, or archive only.
3. Record namespace and shadowing risks for `stage`.

### Phase 2 - Rehearse Local Command And Install Behavior

1. Rehearse local binary naming and install path behavior before changing release surfaces.
2. Rehearse shell completion generation and stale cache cleanup.
3. Do not add a compatibility shim; validate only the canonical `stage` path.

### Phase 3 - Cut Over The External Command Surface In Narrow Slices

1. Update binary/install naming.
2. Update root help and user-facing command examples.
3. Reject any forwarding shim; this spec closes only when `stage` is the sole active command path.

### Phase 4 - Migrate Remaining Active Internal Naming Surfaces

1. Rename the remaining active internal `stage`-named env, state, runtime, path, and code surfaces that are still repository-owned.
2. Keep the migration behavior-scoped and reversible by slice rather than mixing it into release publication.
3. Remove temporary staging exceptions once the replacement names are proven.

### Phase 5 - Align Docs And Normative Specs

1. Update active docs to use `stage` as canonical.
2. Update active normative spec examples to match the final renamed surfaces and explicitly label transitional or historical notes.
3. Keep archived material archived; do not revive legacy wrappers as current behavior.

### Phase 6 - Rehearse CI, Release, And Rollback

1. Rehearse CI with renamed command expectations and isolated cache assumptions.
2. Rehearse asset names, checksums, and installer retrieval without public publication.
3. Rehearse rollback against the prepared release surface.

### Phase 7 - Final Cutover And Verification

1. Apply the rehearsed CI and release changes.
2. Run clean-machine, dirty-machine, focused test, docs copy-paste, and zero-active-reference verification.
3. Publish migration guidance and hold a short post-release verification window.

## Validation Strategy

### Planning Validation

- Ensure `tickets.md`, `plan.md`, `research.md`, and the rename contract agree on the temporary staging rule, the final no-active-legacy end state, and what requires dry-run evidence.

### Focused Technical Validation

- `go test ./cmd/stage/commands`
- `go test ./core/config`
- `go test ./core/lifecycle`

### Operator Dry Runs

- Local build/install rehearsal for `stage`
- Completion and shell-cache rehearsal
- Clean-machine install rehearsal
- Dirty-machine upgrade rehearsal with old `stage` residue present
- CI/release rehearsal with rotated cache assumptions
- Rollback rehearsal against prepared release surfaces
- Zero-active-reference sweep across active code and docs before closeout

## Risks And Mitigations

| Risk | Why It Matters | Mitigation |
|---|---|---|
| Generic command name collision for `stage` | PATH shadowing or package ecosystem conflicts can make the rename look broken | Explicit namespace dry run, PATH guidance, and dirty-machine validation |
| Internal rename scope expands too early | Mixed risk classes make failures harder to isolate and harder to roll back | Phase internal migration after the public command cutover, but keep it inside this spec's final end state |
| CI or installer false green due to stale binary or cache | The cutover can appear healthy while still shipping the old command | Rehearse with isolated caches and validate actual installed command behavior |
| Release asset name mismatch | Installer and checksum failures surface too late if not rehearsed | Release-like artifact dry run before publication |
| Docs promise a contract not yet proven by rehearsal | Users hit broken copy-paste paths immediately after release | Keep docs cutover behind docs/spec dry-run and clean-machine verification |
| Rollback path is theoretical only | A failed cutover becomes operationally expensive | Rehearse rollback before public release |

## Complexity Tracking

No constitution violations require justification.