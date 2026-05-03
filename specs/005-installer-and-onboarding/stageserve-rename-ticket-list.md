---
description: "Execution ticket list for StageServe project rename and stage CLI cutover"
---

# Tickets: StageServe Rename And stage Command Cutover

**Objective**: Execute a controlled rename from Stacklane to StageServe with CLI command `stage`, while preserving runtime safety and minimizing operator surprise.

**Scope baseline**:
- Project identity: StageServe
- Primary command: `stage`
- Current state: pre-release
- Default recommendation: keep runtime/state internals (`STACKLANE_*`, `.env.stacklane`, `.stacklane-state`, `stln-*`) stable unless explicitly changed by a dedicated migration ticket

**Out of scope unless explicitly added**:
- Runtime architecture changes unrelated to naming
- Compose topology changes
- Behavior changes to setup/init/doctor flows

**Definition of done**:
- `stage` is the documented and installed command everywhere
- Validation matrix passes on clean and dirty machines
- Rollback path is tested and documented
- No stale command references remain in active docs/specs

## Phase 0: Governance And Contract Lock

- [ ] RENAME-001 Lock naming contract decisions in an ADR at specs/005-installer-and-onboarding/contracts/stageserve-rename-contract.md
Acceptance criteria:
1. Contract states which internal surfaces stay unchanged: `STACKLANE_*`, `.env.stacklane`, `.stacklane-state`, `stln-*`.
2. Contract states whether a temporary `stacklane` compatibility shim is included.
3. Contract defines sunset milestone for shim removal if enabled.
Dependencies: none.

- [ ] RENAME-002 Create cutover checklist owner matrix in specs/005-installer-and-onboarding/quickstart.md
Acceptance criteria:
1. Every ticket has a named owner role and verification role.
2. Release-day runbook sequence is documented.
3. Abort criteria are explicit.
Dependencies: RENAME-001.

## Phase 1: Binary, Install, And Command Surface

- [ ] RENAME-003 Rename installed binary target from `stacklane` to `stage` in install/build paths
Primary files:
1. install.sh
2. Makefile
3. release workflow metadata
Acceptance criteria:
1. Fresh install produces executable named `stage` on PATH.
2. `stage version` works and returns expected version metadata.
3. Installer output text references `stage` only, except explicit compatibility notes.
Dependencies: RENAME-001.

- [ ] RENAME-004 Update root command usage/help strings for StageServe and `stage`
Primary files:
1. cmd/stacklane/commands/*
2. any shared help/usage constants
Acceptance criteria:
1. `stage --help` top banner uses StageServe branding.
2. All command examples in runtime help use `stage`.
3. No user-facing `stacklane` examples remain in command help.
Dependencies: RENAME-003.

- [ ] RENAME-005 Regenerate and validate shell completions for `stage`
Acceptance criteria:
1. zsh, bash, and fish completions install/load under `stage`.
2. Completion scripts do not require old command name.
3. Cached completion behavior is tested in a new shell session.
Dependencies: RENAME-003.

## Phase 2: Compatibility Shim (Recommended Even Pre-Release)

- [ ] RENAME-006 Add optional `stacklane` shim forwarding to `stage`
Acceptance criteria:
1. `stacklane <args>` forwards to `stage <args>` preserving exit code.
2. Shim prints concise deprecation notice to stderr once per invocation.
3. Shim does not alter stdout payload contract for JSON modes.
Dependencies: RENAME-003.

- [ ] RENAME-007 Define shim removal milestone and telemetry proxy signal
Acceptance criteria:
1. Removal version/milestone documented in README and contract.
2. If telemetry is unavailable, a manual removal criterion is documented.
3. A release note template exists for final removal.
Dependencies: RENAME-006.

## Phase 3: Documentation, Specs, And Operator Contract

- [ ] RENAME-008 Update active docs command references to `stage`
Primary files:
1. README.md
2. docs/runtime-contract.md
3. docs/installer-onboarding.md
4. docs/migration.md
Acceptance criteria:
1. Active docs use `stage` for all canonical commands.
2. If shim exists, transitional note is present and time-bound.
3. No archived behavior is presented as active.
Dependencies: RENAME-004.

- [ ] RENAME-009 Add discoverability alias note: "StageServe (formerly Stacklane)"
Acceptance criteria:
1. README opening section includes searchable rename note.
2. docs/migration.md includes old-to-new command table.
3. Repo short description/metadata plan is documented.
Dependencies: RENAME-008.

- [ ] RENAME-010 Update spec artifacts impacted by command naming
Primary files:
1. specs/004-workflow-and-lifecycle/* where command literal is normative
2. specs/005-installer-and-onboarding/quickstart.md
Acceptance criteria:
1. Normative command examples reflect `stage`.
2. Historical notes clearly marked as legacy where required.
3. Quickstart validation commands are executable as written.
Dependencies: RENAME-008.

## Phase 4: CI/CD, Release, And Distribution

- [ ] RENAME-011 Update release asset naming and checksum publication
Acceptance criteria:
1. Artifact names and checksum manifest refer to `stage` binaries.
2. install.sh retrieval logic matches published asset names exactly.
3. Release notes include rename callout and migration guidance.
Dependencies: RENAME-003.

- [ ] RENAME-012 Update CI command invocations and smoke tests
Primary files:
1. scripts/tests/*
2. .github/workflows/*
Acceptance criteria:
1. CI calls `stage` as canonical command.
2. Any shim test is explicit and isolated.
3. Pipeline passes without depending on stale `stacklane` binary in runner cache.
Dependencies: RENAME-003.

- [ ] RENAME-013 Audit cache keys, artifact paths, and restore behavior
Acceptance criteria:
1. Cache keys containing old command/repo names are rotated.
2. No restore path silently masks missing `stage` binary.
3. Build logs show expected new paths.
Dependencies: RENAME-011.

## Phase 5: Niche And Weird Edge-Case Hardening

- [ ] RENAME-014 Shell hash and PATH shadowing validation on macOS
Acceptance criteria:
1. Validation run includes old binary present in a competing PATH directory.
2. Test procedure includes `hash -r` or clean shell restart.
3. Operator guidance includes how to detect/resolve shadowed binaries.
Dependencies: RENAME-003.

- [ ] RENAME-015 Completion cache residue and stale plugin checks
Acceptance criteria:
1. Old completion files are identified and cleanup guidance documented.
2. zsh compinit cache refresh procedure is validated.
3. No stale completion references old command after cleanup.
Dependencies: RENAME-005.

- [ ] RENAME-016 Case-sensitive filesystem and symlink trap test
Acceptance criteria:
1. Rename flow is tested on case-sensitive path assumptions.
2. Symlink loop or broken symlink behavior is explicitly guarded.
3. Installer idempotency verified when both names exist.
Dependencies: RENAME-003, RENAME-006.

- [ ] RENAME-017 Noninteractive and JSON output parity under shim
Acceptance criteria:
1. `NONINTERACTIVE=1` paths behave identically via `stage` and shim.
2. JSON output has no deprecation text contamination on stdout.
3. Exit codes remain equivalent for success, needs_action, and error cases.
Dependencies: RENAME-006.

- [ ] RENAME-018 Namespace recheck before release candidate cut
Acceptance criteria:
1. `which stage` run on target validation machines.
2. Formula and package index spot-check repeated (Homebrew plus at least one secondary ecosystem).
3. Results recorded with date in quickstart evidence section.
Dependencies: RENAME-002.

## Phase 6: Verification Matrix (Must Pass)

- [ ] RENAME-019 Clean machine install and first-run smoke
Acceptance criteria:
1. Install on machine with no prior Stacklane binary succeeds.
2. `stage init`, `stage up`, `stage status`, `stage logs`, `stage down` execute as documented.
3. Recorded evidence includes exact commands and outputs summary.
Dependencies: RENAME-011, RENAME-012.

- [ ] RENAME-020 Dirty machine upgrade with old binary present
Acceptance criteria:
1. Existing `stacklane` binary does not prevent `stage` from becoming canonical.
2. Expected precedence behavior documented and validated.
3. Shim behavior (if enabled) verified end to end.
Dependencies: RENAME-006, RENAME-019.

- [ ] RENAME-021 Focused test suites and contract checks
Acceptance criteria:
1. Run focused tests for cmd/stacklane/commands, core/config, core/lifecycle.
2. Command help/version tests include new command literal coverage.
3. Test evidence captured in quickstart.md.
Dependencies: RENAME-004, RENAME-012.

- [ ] RENAME-022 Docs copy/paste audit
Acceptance criteria:
1. Every command block in active docs executes with `stage`.
2. Stale `stacklane` references are either removed or marked legacy.
3. Failures are logged and fixed before cut.
Dependencies: RENAME-008, RENAME-010.

## Phase 7: Release, Rollback, And Closure

- [ ] RENAME-023 Create rollback tag and rollback drill
Acceptance criteria:
1. Pre-cutover git tag is created and documented.
2. Rollback procedure (installer target, docs quickstart, release note) is rehearsed once.
3. Time-to-rollback estimate is recorded.
Dependencies: RENAME-011.

- [ ] RENAME-024 Publish rename release notes and migration guidance
Acceptance criteria:
1. Release notes include command migration table and shim timeline.
2. Known issues section includes PATH shadowing and completion cache cleanup.
3. Links to updated docs are validated.
Dependencies: RENAME-009, RENAME-023.

- [ ] RENAME-025 Post-release verification window and closeout report
Acceptance criteria:
1. Post-release checks run for at least one full day-cycle.
2. Any reported rename regressions are triaged and resolved.
3. Closeout report stored at specs/005-installer-and-onboarding/code-review.md or equivalent.
Dependencies: RENAME-024.

---

## Execution Order

1. Phase 0: Governance and contract lock.
2. Phase 1: Binary and command surface.
3. Phase 2: Compatibility shim (if enabled).
4. Phase 3: Docs/spec alignment.
5. Phase 4: CI/CD and release assets.
6. Phase 5: Weird edge-case hardening.
7. Phase 6: Verification matrix.
8. Phase 7: Release and closure.

## Parallelization Guidance

- Can run in parallel:
1. RENAME-005 and RENAME-008 after RENAME-003/004.
2. RENAME-011 and RENAME-012 after RENAME-003.
3. RENAME-014/015/016 once core rename is in place.

- Keep sequential:
1. RENAME-001 before all implementation tickets.
2. RENAME-019 before RENAME-020.
3. RENAME-023 before RENAME-024.

## Risk Register Snapshot

- High risk: stale PATH entries causing users to run old binary unexpectedly.
- High risk: deprecation text leaking into JSON stdout during shim forwarding.
- Medium risk: stale shell completion caches producing misleading command hints.
- Medium risk: CI cache key reuse masking artifact/path mistakes.
- Medium risk: docs drift where old command appears in less-visible pages.

## Ticket Template For Tracking System

Use this structure for each imported ticket:

- Title: RENAME-XXX Short action phrase
- Problem statement: One paragraph
- Scope: Explicit include and exclude bullets
- Acceptance criteria: Numbered, testable, binary pass/fail
- Dependencies: Ticket IDs
- Validation evidence: command output paths, test logs, and doc links
- Rollback note: specific reversal steps if ticket breaks release path
