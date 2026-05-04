# Rename Contract: StageServe And `stage`

## Status

Proposed contract for Gate A review.

This file exists to freeze the rename boundaries before implementation. No public cutover work should proceed until this contract is accepted.

## External Identity Contract

- Product name: `StageServe`
- Canonical CLI command: `stage`
- Searchability note: active docs should use StageServe-only branding.

## Internal Naming End-State Contract

The final state for this spec is:

- no active legacy references remain in code or maintained documentation
- legacy naming survives only in clearly historical notes, archival material, or explicitly labeled migration history

This includes repository-owned internal surfaces such as:

- lingering legacy branding
- any env-file references that do not use `.env.stageserve`
- any state-dir references that do not use `.stageserve-state`
- package, module, path, and layout names that still carry the legacy identity

Temporary staging rule:

- these surfaces may remain temporarily during early dry-run and cutover phases if needed to reduce risk
- they are not acceptable as the final closeout state for this spec

Runtime prefix rename:

- `stln-*` must be renamed to `stage-*` by spec closeout
- this includes Docker Compose project names, network names, and volume names (e.g. `stln-shared` → `stage-shared`, `stln-<slug>` → `stage-<slug>`)
- no live runtime migration is required; this is a local-development-only product and a clean `docker system prune` is acceptable before first use under the new prefix

## Output Contract

- The canonical command must preserve existing exit-code behavior for equivalent success, needs-action, and failure paths.
- Machine-readable and JSON output must remain clean on stdout.
- Help text and examples must present `stage` as canonical once cutover is complete.

## Compatibility Contract

- No compatibility forwarding shim is shipped.
- `stage` is the only supported executable path.
- Machine-readable and JSON modes stay clean because there is no compatibility layer injecting messaging.

## Distribution Contract

The following external surfaces are in scope for this phase:

- installed binary name
- installer output and retrieval targets
- release asset names and checksums
- shell completions
- active docs and normative specs
- CI smoke paths and release automation that invoke the canonical command
- active internal naming surfaces that still carry the legacy identity

The following surfaces require explicit classification before change:

- repo metadata and short description
- package-manager specific identities
- any external references that are not repository-owned but are linked from active docs

## Recommended Gate A Defaults

These defaults are the current recommended path unless dry runs uncover a blocker:

- use a hard switch: ship `stage` as canonical and do not keep any compatibility shim
- use `.env.stageserve` as the env-file contract
- use `.stageserve-state` as the state-dir contract
- use `STAGESERVE_*` as the environment-variable contract
- rename `stln-*` to `stage-*` (Docker Compose project names, network names, volume names)
- use `StageServe` / `stageserve` for runtime-visible gateway headers and sentinels
- rename the Go module path to `github.com/peternicholls/stageserve` only when repository identity is ready to support it

Rationale:

- these changes remove literal legacy references from active code and docs
- they keep the highest-risk scope focused on names rather than runtime behavior changes
- they avoid widening the migration to every internal abbreviation unless there is independent value

## Non-Goals

- widen lifecycle behavior or change compose topology
- restore archived Bash wrappers as current behavior
- keep temporary compatibility surfaces longer than needed for the cutover rehearsals

## Dry-Run Gates Required By This Contract

Before public cutover, the following must be green:

1. Active-surface inventory and classification
2. Local binary/install rehearsal
3. Completion and shell-cache rehearsal
4. Docs/spec rehearsal
5. CI/release rehearsal
6. Rollback rehearsal
7. Final zero-active-reference rehearsal for active code and docs

## Approval Checklist

- [x] External name and canonical command accepted
- [x] Final no-active-legacy end state accepted
- [x] Shim policy accepted (hard switch; no shim)
- [x] `stln-*` → `stage-*` runtime prefix rename accepted (local dev only; `docker system prune` acceptable)
- [x] Gate A approved — implementation may begin
- [ ] Distribution surfaces for this phase accepted
- [ ] Abort criteria and rollback owner named
- [ ] Dry-run evidence locations defined