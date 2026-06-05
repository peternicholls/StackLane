# Research: Guided Experience And Runtime Hardening

## Decision 1: Ship runtime assets via bundled installer artifact in 011

- Decision: Use a bundled release/install artifact containing StageServe binary plus `stacks/20i` runtime assets as the supported 011 repair path.
- Rationale: This resolves current missing-asset startup failures without requiring immediate embedded-asset architecture changes.
- Alternatives considered:
  - Embed runtime assets directly in binary now.
  - Keep current split delivery and document manual copy.

## Decision 2: Enforce runtime asset preflight before compose startup

- Decision: Add explicit preflight existence checks for `cfg.SharedFile` and `cfg.StackFile` in startup paths before compose invocation.
- Rationale: Failing early prevents opaque compose errors and allows StageServe-native remedies with reliable next actions.
- Alternatives considered:
  - Detect missing files only after compose returns an error.
  - Attempt auto-regeneration of missing runtime files at startup.

## Decision 3: Treat machine-not-ready as a planner reachability contract

- Decision: Root guidance context must include readiness checks that can classify `machine_not_ready` for bare `stage` when prerequisites are missing.
- Rationale: The easy-mode contract is broken if root flow defaults to ready-to-run while required assets/setup are absent.
- Alternatives considered:
  - Keep cheap heuristic as state-dir-only.
  - Force full setup command execution for every bare `stage` invocation.

## Decision 4: Standardize error/presentation semantics across guided, text, and direct paths

- Decision: Use shared semantic severity language and style tokens across TUI, text fallback, and direct command error rendering.
- Rationale: Mixed tone and hierarchy currently make troubleshooting inconsistent and harder to trust.
- Alternatives considered:
  - Keep current shell-only token ownership in guided path.
  - Copy style logic into direct command renderers ad hoc.

## Decision 5: Deliver failure recovery and day-2 actions incrementally

- Decision: Keep merge-gate focused on runtime safety; deliver richer recovery and running-project day-2 actions in the follow-up slice with explicit gating.
- Rationale: This preserves release safety while still committing to full spec completion paths.
- Alternatives considered:
  - Block merge on full UX expansion.
  - Defer recovery/day-2 entirely to a separate future spec.

## Decision 6: Restrict password input to existing env/config surfaces

- Decision: Remove root `--mysql-password` flag and preserve only existing `MYSQL_PASSWORD` sources (shell env, project `.env.stageserve`, stack `.env.stageserve`).
- Rationale: CLI password flags leak secrets and violate operator-safety expectations for this stack.
- Alternatives considered:
  - Keep password CLI flag for compatibility.
  - Add interactive password prompt/new secret storage surface.

## Decision 7: Scope async action handling to long-running lifecycle operations first

- Decision: Move `up`, `down`, `attach`, `detach`, and lifecycle retry/recovery actions onto async `tea.Cmd` handling first; keep short inspect/open actions synchronous.
- Rationale: This yields immediate UX value (progress/cancel visibility) with lower risk than broad async conversion.
- Alternatives considered:
  - Convert all guided actions to async immediately.
  - Keep all actions synchronous and add only cosmetic loading text.