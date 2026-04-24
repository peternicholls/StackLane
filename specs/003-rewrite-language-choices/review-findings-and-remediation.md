# Spec 003 Review: Findings And Remediation Plan

## Purpose

This document captures the current review of spec `003-rewrite-language-choices` after the repository policy changed.

New governing rule:

- Legacy compatibility is no longer a default project constraint.
- Agents should prefer current Stacklane naming, state layout, and command surface even when that breaks `20i-*` wrappers, `.20i-*` files, migration fallbacks, or older workflow assumptions.
- Backward compatibility, migration shims, and legacy behavior should only be kept when explicitly requested by the user.

This review is intended as a handoff artifact for the next agent session.

## Executive Summary

Spec 003, its implementation plan, its task list, parts of the current code, and parts of the current documentation still assume that preserving legacy behavior is required. That is now out of date.

The main problem is not just stray wording. The current spec suite still tells agents to preserve compatibility as a hard constraint, and some active runtime paths still implement that policy:

- the repo-root `stacklane` entrypoint still runs the Bash engine
- the Go loader still falls back to `.20i-local` and `.20i-state`
- the CLI still runs legacy state migration on startup
- the deprecated `20i-*` wrappers still exist and still forward through Bash
- user-facing docs still describe migration-window behavior as current policy

## Findings

### 1. Spec 003 still treats backward compatibility as a hard requirement

This is the primary conflict.

Evidence:

- `specs/003-rewrite-language-choices/spec.md`
  - User Story 3 says the rewrite should be invisible at the command-line level and explicitly frames backward compatibility as what makes the rewrite safe to ship.
  - Operational Impact says command surface, flags, precedence chain, state locations, and gateway behavior remain unchanged.
  - Configuration section still defines precedence in terms of `.20i-local`.
  - State section still requires reading the legacy format and migrating non-destructively.
- `specs/003-rewrite-language-choices/plan.md`
  - Constraints still require preserving `.20i-local`, state locations, and migration behavior.
  - Phase descriptions still assume a stabilization window and parity/migration gates.
- `specs/003-rewrite-language-choices/tasks.md`
  - Explicitly states that backward compatibility is a hard constraint.
  - Includes preserved migration, wrapper delegation, parity-against-Bash, and Bash-removal-only-after-stabilization as required work.

Impact:

- Any future agent using spec 003 as the implementation contract is likely to preserve or reintroduce legacy behavior.
- This now conflicts with the repo-level instruction file and the repository memory note.

### 2. The documented primary entrypoint is not the active one

Evidence:

- `stacklane` at repo root still sources `lib/stacklane-common.sh` and runs `stacklane_main`.
- `lib/stacklane-common.sh` still contains the active command dispatcher for `up`, `down`, `attach`, `detach`, `status`, `dns-setup`, and `logs`.
- `README.md` says the command surface is implemented as a single Go binary and that the `stacklane` shim execs `stacklane-bin`.

Impact:

- The current operator entrypoint is still the Bash runtime, not the Go binary.
- This conflicts with both the docs and the desired direction of spec 003.
- Under the new repo rule, this should no longer be treated as an acceptable migration stage.

### 3. Legacy `.20i-*` fallbacks are still active runtime behavior

Evidence:

- `core/config/loader.go`
  - `defaultStateDir()` falls back from `.stacklane-state` to `.20i-state`.
  - `loadProjectEnv()` reads `.20i-local` and `.stacklane-local`.
- `cmd/stacklane/commands/root.go`
  - `buildOrchestrator()` calls `state.Migrate(cfg.StateDir)` before running the orchestrator.
- `lib/stacklane-common.sh`
  - still resolves `.20i-state` as a fallback state dir
  - still resolves `.20i-local` as a fallback project config

Impact:

- Legacy behavior is still present in the active control path, not just in archive files or historical specs.
- The new policy implies these fallbacks should be removed or made opt-in only if explicitly requested.

### 4. Deprecated wrapper scripts are still treated as supported migration-era behavior

Evidence:

- Root-level `20i-up`, `20i-down`, `20i-attach`, `20i-detach`, `20i-status`, `20i-logs`, `20i-dns-setup` still exist.
- They still source `lib/stacklane-common.sh` and call the legacy forwarding function.
- `lib/stacklane-common.sh` still contains compatibility help text and deprecation notices for these wrappers.
- `README.md`, `docs/runtime-contract.md`, and `docs/migration.md` still describe these wrappers as retained for a migration window.

Impact:

- This keeps the old interface alive in both runtime and docs.
- Under the new rule, the wrappers should not be preserved simply because they existed before.

### 5. Spec-003 migration preservation still shapes the state model

Evidence:

- `core/state/migration.go` preserves legacy `.env` state files as `.legacy` siblings and never deletes them.
- `spec.md`, `plan.md`, and `tasks.md` all require non-destructive migration and explicit retention of legacy files.

Impact:

- The code still assumes migration safety is a primary goal.
- That may no longer match the current project direction if breaking old workflows is acceptable.

### 6. Some docs now overstate the Go rewrite status

Evidence:

- `README.md` describes the runtime as a single Go binary and says the repo-root launcher execs `stacklane-bin`.
- In practice, the repo-root launcher still runs Bash.

Impact:

- There is a documentation/runtime mismatch.
- The next agent should treat docs and runtime together, not separately.

## What Should Be Treated As Historical Only

The following should be treated as historical background, not as current requirements:

- any claim that backward compatibility is a hard constraint
- any requirement to preserve `.20i-local` or `.20i-state`
- any requirement to preserve `20i-*` wrapper commands
- any multi-week stabilization gate before removing Bash
- any requirement to keep migration shims by default
- any parity requirement whose sole purpose is preserving old workflows rather than validating current Stacklane behavior

## Remediation Plan

### Phase 1: Rewrite The Spec-003 Contract First

Goal: stop future agents from following the wrong contract.

Update these files first:

- `specs/003-rewrite-language-choices/spec.md`
- `specs/003-rewrite-language-choices/plan.md`
- `specs/003-rewrite-language-choices/tasks.md`

Required changes:

- Remove or rewrite User Story 3 so it no longer treats invisible migration and backward compatibility as the success condition.
- Replace `.20i-local` precedence references with `.stacklane-local` only.
- Replace `.20i-state` migration/fallback requirements with `.stacklane-state` as the sole canonical state location.
- Remove wrapper-delegation requirements.
- Remove stabilization-window language that blocks Bash removal.
- Reframe parity work so it validates Stacklane behavior where useful, not Bash compatibility as a hard contract.
- Rewrite deferred tasks T024, T037, T056, T062, T065, T068 and related notes so they no longer preserve migration-era behaviors by default.
- Remove the note in `tasks.md` that says backward compatibility is a hard constraint.

Why first:

- If code changes happen before the spec suite is corrected, the next agent may be pulled back toward legacy-preserving decisions by the existing spec text.

### Phase 2: Switch The Live Entrypoint To The Go Binary

Goal: make the actual command path match the intended architecture.

Target files:

- `stacklane`
- root-level `20i-*` scripts
- possibly `Makefile`, release/build wiring, and help docs

Required changes:

- Change `stacklane` so it execs `stacklane-bin` directly instead of sourcing `lib/stacklane-common.sh`.
- Remove root-level dependency on the Bash runtime for normal operation.
- Decide whether root-level `20i-*` scripts should be deleted outright or converted to hard-fail notices. Under the current rule, deletion is acceptable.

Validation:

- `stacklane --help`
- targeted command wiring checks for `up`, `down`, `status`, `logs`, `dns-setup`
- focused Go tests on `cmd/stacklane/commands`

### Phase 3: Remove Active Legacy Fallbacks From Go Code

Goal: stop the runtime from honoring `.20i-*` compatibility paths.

Target files:

- `core/config/loader.go`
- `core/config/types.go`
- `cmd/stacklane/commands/root.go`
- `core/state/migration.go`
- related tests in `core/config/*_test.go` and `core/state/*_test.go`

Required changes:

- Remove `.20i-local` fallback from config loading.
- Remove `.20i-state` fallback from state-dir resolution.
- Stop automatically migrating legacy `.env` state on startup unless there is an explicit user requirement to keep that path.
- Rework or delete migration tests whose only purpose is preserving old formats.
- Update comments and docstrings so they no longer describe compatibility behavior as live behavior.

Validation:

- targeted Go tests for `core/config`, `core/state`, and `cmd/stacklane/commands`
- confirm state directory resolution only points to `.stacklane-state`

### Phase 4: Remove Or Archive The Remaining Bash Runtime

Goal: make Bash historical, not operational.

Target files:

- `lib/stacklane-common.sh`
- root-level `20i-*`
- root-level `stacklane` if any shim remains
- `previous-version-archive/`

Required changes:

- Move any still-needed historical Bash sources into `previous-version-archive/`.
- Ensure nothing in the active runtime imports or sources `lib/stacklane-common.sh`.
- If the archive keeps wrapper samples, keep them only under `previous-version-archive/`, not at the repo root.

Validation:

- search for `stacklane-common.sh` references in active runtime surfaces
- search for `twentyi_legacy_forward` references outside archive/spec history

### Phase 5: Rewrite Docs To Match The New Policy And Runtime

Goal: align docs with what the product now is.

Target files:

- `README.md`
- `docs/runtime-contract.md`
- `docs/migration.md`
- `docs/architecture.md`
- `docs/tui/*.md`
- any still-maintained plan or onboarding docs

Required changes:

- Remove migration-window language that is no longer policy.
- Remove references to `.20i-local` and `.20i-state` as supported fallbacks.
- Remove or rewrite wrapper-command guidance.
- Make the docs accurately describe the actual entrypoint and runtime path.
- Keep `20i` only where it describes the emulation target, not the product or operations.

Validation:

- grep for `legacy|migration|compatib|\.20i-|20i-\*` across maintained docs
- manually confirm remaining `20i` mentions are emulation-only or archival

## Recommended Execution Order For The Next Agent

1. Rewrite `spec.md`, `plan.md`, and `tasks.md` for spec 003.
2. Switch the repo-root `stacklane` entrypoint to `stacklane-bin`.
3. Remove `.20i-local` and `.20i-state` fallbacks from active Go code.
4. Remove or archive root-level `20i-*` scripts and live Bash runtime dependencies.
5. Sweep docs so the documented contract matches the new code.
6. Run focused validation after each phase, not only at the end.

## Focused Validation Checklist

Use the narrowest available checks after each change set:

- `go test ./cmd/stacklane/commands`
- `go test ./core/config ./core/state ./core/project ./core/lifecycle`
- `go test ./platform/dns ./infra/gateway`
- if any Bash shims remain temporarily, `bash -n stacklane lib/stacklane-common.sh`
- grep for remaining active references:
  - `.20i-local`
  - `.20i-state`
  - `twentyi_legacy_forward`
  - `stacklane-common.sh`
  - `20i-*`

## Notes For The Next Agent

- Do not treat spec 003 as authoritative until its contract files are rewritten.
- Prefer deleting legacy-preserving behavior over wrapping it in one more compatibility layer.
- If a change forces a small break in local developer workflow, that is acceptable under the current repo policy.
- Historical files under `previous-version-archive/` and older planning material can be left for a later archival pass unless they directly mislead active implementation work.
