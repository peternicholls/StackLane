# Long-Term Plan: Embedded Runtime Assets

**Status**: Deferred long-term direction  
**Related Spec**: [spec.md](./spec.md)  
**Related Repair Choice**: 011 uses a bundled release/install artifact for runtime asset provisioning. This document captures the longer-term direction of embedding runtime stack assets directly into the StageServe binary.

## Purpose

Eliminate drift between the installed StageServe binary and the runtime compose assets by making the binary the source of truth for versioned stack assets.

This is not part of the 011 merge gate. The 011 repair work uses bundled install/release assets because that is a smaller and safer repair. This document exists so the better long-term architecture is explicit and does not need to be rediscovered later.

## Why Defer It

The embedded-assets model is attractive, but it is a broader runtime and packaging change than the 011 repair gate needs.

The short-term bundled-artifact repair is enough to:

- restore a working installed-binary path
- keep runtime asset paths compatible with the current loader and stack-home model
- avoid turning a repair spec into a packaging and runtime architecture rewrite

## Target Outcome

A future StageServe binary should carry the active stack assets internally and materialize them deterministically when needed, so that:

- the installed binary and runtime assets cannot version-drift silently
- operators do not need a second asset-copy step after installation
- supported install paths stay simple even as stack assets evolve
- StageServe can validate, refresh, or re-materialize assets predictably

## Proposed Architecture

### 1. Bundle Stack Assets In The Binary

Use Go embed for the active stack asset tree, initially:

- `stacks/20i/docker-compose.shared.yml`
- `stacks/20i/docker-compose.20i.yml`
- any additional stack-owned static assets required for the supported runtime contract

### 2. Add A Runtime Asset Manager

Introduce a small runtime-asset manager responsible for:

- exposing the bundled asset version/hash
- checking whether the materialized assets on disk are present and current
- writing missing or stale assets to disk safely
- preserving operator-owned config while managing only StageServe-owned generated assets

This should be a narrow seam, not a second config system.

### 3. Keep The On-Disk Runtime Contract Stable

Materialized assets should continue to appear in the current stack-home layout unless there is a deliberate contract change:

- `<stack-home>/stacks/20i/docker-compose.shared.yml`
- `<stack-home>/stacks/20i/docker-compose.20i.yml`

That keeps loader and orchestrator behavior compatible while changing where the files come from.

### 4. Materialize Deterministically

Materialization should happen only when needed:

- first run on a new machine
- missing asset detected
- StageServe version changes and bundled asset hash differs
- explicit repair or refresh command in a future enhancement

The write path should be idempotent and should not overwrite operator-owned files outside the managed runtime asset set.

### 5. Make Drift Visible

If the bundled hash and on-disk managed asset hash disagree, StageServe should be able to distinguish:

- expected upgrade refresh
- missing managed files
- user-modified managed files

That distinction matters because forced overwrite behavior should be explicit.

## Candidate Implementation Shape

### Option A - Materialize During Config Load

Pros:

- asset availability is resolved very early
- later runtime code can assume files exist

Cons:

- config loading takes on filesystem side effects
- harder to preserve a clean read-only loading seam

### Option B - Materialize Just Before Lifecycle Use

Pros:

- keeps config loading mostly read-only
- side effects occur only when runtime assets are actually needed

Cons:

- every lifecycle path must consistently call the materializer
- bare `stage` and diagnostic flows may need extra branching

### Option C - Explicit `ensure-runtime-assets` Seam Used By Loader And Lifecycle

Pros:

- keeps the side-effect boundary explicit
- easiest path to testing and future repair commands

Cons:

- one more internal seam to maintain

**Preferred long-term shape**: Option C.

## Key Design Constraints

- Do not create a new user-owned config surface.
- Keep stack-home layout compatible unless there is a deliberate contract update.
- Do not overwrite operator-owned files silently.
- Do not bury runtime asset repair behind raw Compose errors.
- Preserve non-TTY and direct-command safety guarantees.

## Migration Strategy

### Phase 1 - Land 011 Repair

Use the bundled release/install artifact from 011 to restore correctness.

### Phase 2 - Introduce Embedded Asset Manager Behind The Same Paths

Add the embedded asset manager while keeping the on-disk file paths stable.

### Phase 3 - Transition Installer Responsibilities

Reduce installer responsibility from copying stack assets to installing only the binary and any supporting metadata needed by the embedded asset manager.

### Phase 4 - Add Refresh/Repair Semantics If Needed

Optionally add an explicit repair path if operators need a supported way to force re-materialization.

## Validation Strategy

When this long-term plan is eventually implemented, validation should include:

- first install on a clean machine
- upgrade from one StageServe version to another with changed bundled assets
- recovery when managed asset files are deleted
- behavior when managed asset files are manually edited
- `stage up` and bare `stage` behavior when materialization is required
- non-TTY and JSON paths remaining clean

## Open Questions

- Should StageServe preserve a backup when replacing stale managed asset files?
- How should the tool behave if managed asset files were hand-edited?
- Is a user-visible repair command necessary, or is automatic repair enough?
- Should asset version/hash information be recorded in `.stageserve-state`, or derived solely from the binary and on-disk files?

## Non-Goals

- Reworking the stack-home contract as part of this long-term plan.
- Expanding into broader multi-stack packaging work before the current 20i path is stable.
- Turning runtime asset management into a user-facing feature unless a real operator need appears.
