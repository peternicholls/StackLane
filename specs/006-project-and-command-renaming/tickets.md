---
spec phase: 006 Project and Command Renaming
description: "Gated workplan for StageServe project rename and stage CLI cutover"
source: /stageserve-rename-ticket-list.md
planning_style: gated-hybrid
---

# Workplan: StageServe Rename And stage Command Cutover

## Planning Choice

This phase should not be run as a pure ticket backlog.

The safer planning style here is a gated hybrid:

1. A short contract-first phase plan that defines decision gates, dry-run exits, and cutover boundaries.
2. Actionable task packets under each phase, each with one primary surface, one narrow validation target, and explicit abort and rollback notes.

Why this style is preferred for this phase:

- The rename crosses operator contract, install paths, shell behavior, docs, CI, release assets, and rollback.
- A story-style backlog encourages out-of-order execution and hides when a step is only partially reversible.
- A pure implementation plan is too coarse for careful handoff and evidence capture.
- The work needs explicit DRY RUN checkpoints before any public or operator-visible cutover.

## Objective

Execute the final StageServe cutover with CLI command `stage`, while preserving runtime safety, allowing phased execution where necessary, and finishing with no active legacy references left in code or maintained documentation.

## Scope Baseline

- Product identity: StageServe
- Primary command: `stage`
- Current state: pre-release
- Default posture: phase risky rename slices if needed, but do not leave any active legacy references behind by spec closeout

Temporary staging surfaces that may survive early gates but not final closeout:

- `.env.stage`
- `.stageserve-state`

Runtime prefixes `stln-*` must be renamed to `stage-*` by closeout. This covers Docker Compose project names, network names, and volume names. No live runtime migration is required; a `docker system prune` before first use under the new prefix is acceptable.

## Out Of Scope Unless Explicitly Reopened

- Runtime architecture changes unrelated to naming
- Compose topology changes
- Behavior changes to setup, init, doctor, or lifecycle flows beyond command and branding surfaces
- Archive cleanup beyond making sure archived material is not described as current

## Definition Of Done

- `stage` is the documented and installed canonical command everywhere active
- No active legacy references remain in code or maintained documentation by closeout
- Every change packet has dry-run evidence before the real change packet lands
- Validation matrix passes on both clean and dirty machines
- Rollback path is rehearsed and documented before cutover
- Only historical or archival material may still mention StageServe after closeout
- No compatibility shim remains in scope or implementation

## Contract Freeze Before Any Implementation

These decisions must be locked before any rename changes start:

1. External naming contract
	 - Product name: StageServe
	 - Canonical command: `stage`
	 - Active branding uses StageServe only
2. Internal naming end-state contract
	 - Temporary staging exceptions are allowed during early phases only
	 - `.env.stage` cannot remain in active code or maintained docs by final closeout
	 - `stln-*` must be renamed to `stage-*` by closeout
3. Compatibility contract
	 - No temporary compatibility shim exists
4. Distribution contract
	 - Decide which release, installer, checksum, and package surfaces change in this phase
5. Output contract
	 - Keep stdout clean for JSON and machine-readable modes
	 - Preserve exit-code parity for canonical `stage` command

## How Actionable Tasks Must Be Written

Each task in this phase must be a controlled change packet with the following fields:

- Goal: one sentence, one primary surface
- Preconditions: exact gate or task that must already be green
- Dry run: the rehearsal step that proves the slice is safe before real cutover
- Real change: the smallest operator-visible mutation that follows from the dry run
- Validation: the narrowest executable or inspectable check for that slice
- Evidence: where commands, outputs, or notes are recorded
- Abort if: the condition that blocks promotion
- Rollback: how to revert this slice only

Task writing rules:

- Separate decision tasks from implementation tasks
- Separate dry-run tasks from production-facing tasks
- Do not combine binary rename, docs sweep, CI migration, and release publication in one task
- Make clean-machine and dirty-machine validation separate tasks
- Prefer one owning surface per task: command help, installer, completions, docs, CI, release, rollback
- Every task must define the cheapest discriminating check, not just "tests pass"

## Required DRY RUN Gates

No public or irreversible cutover work starts until these gates are green.

### Gate A: Contract Dry Run

Purpose: freeze the rename contract, final end state, and non-goals.

Exit when:

- External naming and temporary staging boundaries are explicit
- No compatibility shim is permitted
- Abort criteria and rollback owner are named

### Gate B: Inventory Dry Run

Purpose: locate every active `stage` reference and classify it before editing.

Exit when:

- Each hit is classified as one of: rename now, rename later in this spec, mark legacy, archive only
- No unresolved namespace or ownership ambiguity remains

### Gate C: Local Binary And Install Dry Run

Purpose: prove that the renamed command can be built, installed, found on PATH, and invoked safely before release surfaces change.

Exit when:

- Local install produces `stage`
- `stage --help` and `stage version` behave as expected
- PATH shadowing and shell hash behavior are understood
- Completion generation is rehearsed
- No compatibility command path remains in rehearsal scope

### Gate D: Docs And Spec Dry Run

Purpose: prove that active docs and normative specs match the intended external contract exactly.

Exit when:

- Active docs use `stage` as canonical command
- Internal stable surfaces remain documented correctly where relevant
- Historical or transition notes are clearly marked

### Gate E: CI And Release Dry Run

Purpose: rehearse asset naming, installer retrieval, checksum publication, workflow invocations, and cache rotation without public cutover.

Exit when:

- CI paths work with renamed command and asset expectations
- Release-like artifacts are produced and consumed successfully
- No stale cache or old binary masks failures

### Gate F: Rollback Dry Run

Purpose: prove the team can restore the prior public contract quickly if cutover fails.

Exit when:

- Pre-cutover tag and rollback procedure exist
- Installer target, docs entry point, and release-note rollback steps are rehearsed once
- Time-to-rollback estimate is recorded

## Specialist Cross-Check

This workplan was checked against specialist planning patterns before being written.

### Agent Pattern Guidance Applied

- Senior Project Manager pattern: use a gated hybrid plan instead of a story backlog, and require controlled change packets with explicit dry run, abort, and rollback fields.
- Software Architect pattern: freeze the external contract and the final no-active-`stage` end state before implementation, and do not combine release mechanics with rename mechanics in the same change packet.

### Recommended Skill And Agent Fit By Phase

- Contract and wording review
	- Skills: `clarify`, `distill`
	- Agents: `Senior Project Manager`, `Technical Writer`
- Rename surface and risk review
	- Skills: `harden`
	- Agents: `Software Architect`, `Code Reviewer`
- CLI and runtime validation
	- Skills: none required beyond repo instructions; prefer focused runtime checks
	- Agents: `Backend Architect`, `API Tester`
- Docs and operator journey review
	- Skills: `clarify`, `normalize`
	- Agents: `Technical Writer`, `UX Researcher`
- Release and rollback rehearsal
	- Skills: none required; focus on operational evidence
	- Agents: `DevOps Automator`, `Reality Checker`

## Workplan Phases

### Phase 0: Freeze The Contract

#### RENAME-001 Contract ADR

- Goal: lock the external rename contract and the final no-active-`stage` end state in one ADR.
- Preconditions: none.
- Dry run: review proposed contract against current repo docs, install flow, and runtime naming surfaces without editing implementation.
- Real change: add the ADR that freezes naming, shim policy, non-goals, and sunset expectations.
- Validation: contract contains product name, command name, shim decision, non-goals, rollback ownership, and the final zero-active-reference rule.
- Evidence: ADR plus a brief decision summary in the spec folder.
- Abort if: any surface is still ambiguous about whether it is external, internal-stable, or legacy.
- Rollback: discard ADR and reopen contract review before implementation.

#### RENAME-002 Cutover Runbook Skeleton

- Goal: define owner, verifier, abort criteria, and release-day sequence before technical changes start.
- Preconditions: RENAME-001.
- Dry run: walk the release-day sequence as a tabletop exercise.
- Real change: write `runbook.md` with the runbook skeleton and owner matrix.
- Validation: every later task packet maps to an owner and verification role.
- Evidence: runbook section in the active spec or docs surface.
- Abort if: no named rollback owner or no explicit no-go criteria.
- Rollback: revise runbook before any rename implementation proceeds.

### Phase 1: Inventory Before Change

#### RENAME-003 Active Surface Inventory

- Goal: classify all active `stage` references before editing.
- Preconditions: RENAME-001.
- Dry run: search code, docs, specs, install, tests, CI, and release metadata and classify each hit.
- Real change: record a disposition table in `inventory.md` for active references and explicit exclusions for archived material.
- Validation: every active hit is marked rename now, rename later in this spec, mark legacy, or archive only.
- Evidence: inventory table stored in the spec folder.
- Abort if: any reference cannot be classified confidently.
- Rollback: keep the current plan frozen and resolve classification gaps first.

#### RENAME-004 Command Namespace Risk Check

- Goal: verify that `stage` is safe enough to use as the canonical command on target systems.
- Preconditions: RENAME-003.
- Dry run: inspect PATH conflicts, package ecosystem collisions, and local shell behavior.
- Real change: record the collision posture and operator guidance.
- Validation: documented handling exists for PATH shadowing, shell hashing, and any discovered conflicts.
- Evidence: risk note attached to the workplan or runbook.
- Abort if: a blocking namespace conflict is found with no practical operator guidance.
- Rollback: pause rename cutover and reopen command selection.

### Phase 2: Rehearse Local Command And Install Behavior

#### RENAME-005 Local Build And Install Dry Run

- Goal: prove local build and install paths can produce and expose `stage` before touching release surfaces.
- Preconditions: RENAME-002, RENAME-003, RENAME-004.
- Dry run: build and install locally using release-like naming without publishing anything.
- Real change: none beyond local rehearsal artifacts and recorded findings.
- Validation: `stage --help` and `stage version` work from a fresh shell.
- Evidence: command transcript and findings summary.
- Abort if: binary naming, PATH resolution, or invocation behavior is inconsistent.
- Rollback: delete rehearsal artifacts and resolve local path issues before implementation.

#### RENAME-006 Completion And Shell Cache Dry Run

- Goal: prove completion scripts and shell caches do not hide rename failures.
- Preconditions: RENAME-005.
- Dry run: regenerate zsh, bash, and fish completions and test with stale cache conditions.
- Real change: none beyond rehearsal artifacts and cleanup notes.
- Validation: completions load under `stage`, and old cache cleanup guidance is confirmed.
- Evidence: shell-specific notes plus exact refresh commands.
- Abort if: stale completion or cache behavior cannot be corrected predictably.
- Rollback: keep canonical cutover blocked until shell guidance is fixed.

#### RENAME-007 No-Compatibility Verification Dry Run

- Goal: prove the canonical `stage` path stands alone with no forwarding dependency.
- Preconditions: RENAME-001 and RENAME-005.
- Dry run: verify there is no active legacy command path in the repository-owned tree and no docs depend on one.
- Real change: none until the direct cutover path is verified.
- Validation: canonical command behavior is validated without any shim parity matrix.
- Evidence: zero-compatibility search log.
- Abort if: an active flow still depends on a forwarding command.
- Rollback: keep cutover blocked until the dependency is removed.

### Phase 3: Implement External Rename In Narrow Slices

#### RENAME-008 Binary And Installer Change Packet

- Goal: rename the installed binary target and installer-facing command references to `stage`.
- Preconditions: Gates A through C green.
- Dry run: use findings from RENAME-005 to confirm exact file and path changes.
- Real change: update build and install paths so fresh installs produce `stage`.
- Validation: local install smoke with `stage version`.
- Evidence: focused validation output plus updated runbook note.
- Abort if: installer retrieval, symlink behavior, or invocation path differs from rehearsal.
- Rollback: restore prior binary target and install references.

#### RENAME-009 Help And Command Surface Change Packet

- Goal: update root help and command examples so user-facing help is StageServe and `stage` only.
- Preconditions: RENAME-008.
- Dry run: inventory command help output and example strings before editing.
- Real change: update help banners, examples, and user-facing usage strings.
- Validation: targeted command help checks.
- Evidence: before and after help snapshots.
- Abort if: any user-facing legacy examples remain in active help.
- Rollback: restore prior help strings and reopen help sweep.

#### RENAME-010 Remove Residual Compatibility Paths

- Goal: remove any remaining repository-owned legacy command path or compatibility wording.
- Preconditions: RENAME-007 and RENAME-008.
- Dry run: confirm the zero-compatibility search log still matches the implementation slice.
- Real change: delete or rewrite any remaining compatibility residue.
- Validation: focused zero-active-reference checks for command-path and doc surfaces.
- Evidence: updated zero-compatibility search log.
- Abort if: removing the residue exposes a hidden runtime dependency.
- Rollback: restore only the affected slice long enough to replace it correctly, not as a permanent shim.

### Phase 4: Align Active Docs And Normative Specs

#### RENAME-011A Internal Naming Migration Packet

- Goal: rename the remaining active internal `stage`-named surfaces after the public command cutover has been proven locally.
- Preconditions: RENAME-009 and any dry-run evidence required for the affected internal slice.
- Dry run: identify exact repository-owned env, state, runtime, path, and code names to change in this slice.
- Real change: apply one narrow internal rename slice at a time.
- Validation: focused tests or behavior checks for the owning slice still pass after the rename.
- Evidence: per-slice migration notes and validation outputs.
- Abort if: the rename changes behavior beyond naming or creates a rollback boundary larger than one slice.
- Rollback: revert the affected internal slice only.

#### RENAME-011 Active Docs Change Packet

- Goal: update active docs so `stage` and the final renamed internal surfaces are canonical, and any transition notes are explicit and temporary.
- Preconditions: Gates A through D green, RENAME-009 complete, and relevant internal rename slices landed.
- Dry run: copy-paste audit of active command blocks before editing.
- Real change: update README and active docs with `stage` and the final no-compatibility contract.
- Validation: docs command-block audit using the updated command.
- Evidence: list of audited command blocks and any exceptions.
- Abort if: any active doc still depends on `stage` or another superseded internal name as canonical.
- Rollback: revert docs packet and reopen audit.

#### RENAME-012 Normative Spec Alignment Packet

- Goal: align active normative specs and quickstart artifacts to the frozen final rename contract.
- Preconditions: RENAME-011.
- Dry run: identify which command literals are normative versus historical.
- Real change: update only the normative surfaces that govern active behavior and final naming.
- Validation: every changed normative command example is executable as written.
- Evidence: spec alignment checklist.
- Abort if: a normative example conflicts with current runtime or docs.
- Rollback: restore the spec packet and resolve the contract mismatch first.

### Phase 5: CI, Release, And Distribution Rehearsal Before Cutover

#### RENAME-013 CI Invocation And Cache Dry Run

- Goal: rehearse CI with renamed command expectations and rotated caches before public asset changes.
- Preconditions: Gates A through E green for rehearsal.
- Dry run: run CI and smoke paths with the renamed command and isolated cache assumptions.
- Real change: none until the rehearsal is green.
- Validation: focused CI smoke steps succeed without depending on a stale `stage` binary.
- Evidence: CI rehearsal logs summary.
- Abort if: cache reuse or old binary paths hide failures.
- Rollback: keep CI and release changes blocked until the issue is isolated.

#### RENAME-014 Release Asset Dry Run

- Goal: prove release-like artifacts, checksums, and installer retrieval all line up before publication.
- Preconditions: RENAME-013 and RENAME-008.
- Dry run: produce release-like assets and run the installer against them without public publication.
- Real change: none until the rehearsal is green.
- Validation: install from rehearsal assets yields a working `stage` binary.
- Evidence: artifact manifest and install transcript.
- Abort if: artifact names, checksum names, or retrieval logic do not match exactly.
- Rollback: discard rehearsal assets and fix the mapping.

### Phase 6: Final Cutover Packets

#### RENAME-015 CI And Release Change Packet

- Goal: switch active CI, release metadata, asset naming, and installer retrieval to the rehearsed `stage` contract.
- Preconditions: RENAME-013 and RENAME-014.
- Dry run: confirm rehearsed artifact and CI outputs still match the intended implementation.
- Real change: update active CI and release surfaces.
- Validation: focused pipeline and install smoke checks.
- Evidence: post-change pipeline result summary.
- Abort if: any live CI or install path diverges from rehearsal evidence.
- Rollback: restore prior release metadata and installer targets.

#### RENAME-016 Package And Metadata Packet

- Goal: update repo and distribution discoverability surfaces that are in scope for this phase.
- Preconditions: RENAME-015.
- Dry run: verify exact metadata surfaces to change now versus later.
- Real change: update only approved metadata surfaces and transitional notes.
- Validation: metadata and docs point to the same canonical command and product name.
- Evidence: metadata checklist.
- Abort if: package ecosystem or repo metadata creates a conflicting identity story.
- Rollback: revert metadata packet and keep the runtime cutover intact.

### Phase 7: Verification Matrix

#### RENAME-017 Clean Machine Verification

- Goal: prove the canonical `stage` journey works on a machine with no prior StageServe install.
- Preconditions: RENAME-015.
- Dry run: verify the exact clean-machine script before running it.
- Real change: execute the clean-machine validation matrix.
- Validation: install plus `init`, `up`, `status`, `logs`, and `down` run as documented.
- Evidence: command transcript and output summary.
- Abort if: any documented happy-path step fails or requires undocumented cleanup.
- Rollback: pause release promotion until the clean path is fixed.

#### RENAME-018 Dirty Machine Verification

- Goal: prove upgrade and coexistence behavior when an old `stage` binary or shell residue exists.
- Preconditions: RENAME-015 and RENAME-017.
- Dry run: prepare the dirty-machine state intentionally with old binaries, completions, and PATH conflicts.
- Real change: execute the upgrade verification matrix.
- Validation: `stage` becomes canonical and any shim behavior matches contract.
- Evidence: dirty-machine transcript and residue cleanup notes.
- Abort if: stale binaries, shell caches, or completions can silently mask the renamed command.
- Rollback: hold release and reopen the cutover guidance.

#### RENAME-019 Focused Test Suites And Contract Checks

- Goal: run the narrow automated checks that cover the changed rename surface.
- Preconditions: RENAME-009, RENAME-015.
- Dry run: confirm the smallest relevant test commands based on touched slices.
- Real change: run focused tests and command-surface checks.
- Validation: relevant tests pass and help/version coverage reflects the new command literal.
- Evidence: test summary added to the validation notes.
- Abort if: tests reveal contract drift or stale command literals.
- Rollback: fix the slice before broader validation continues.

#### RENAME-020 Docs Copy-Paste Verification

- Goal: prove active docs are executable as written using `stage`.
- Preconditions: RENAME-011, RENAME-012, RENAME-017.
- Dry run: confirm the final audit list of active command blocks.
- Real change: run the copy-paste audit and log failures.
- Validation: all active command blocks execute or are explicitly marked as illustrative only.
- Evidence: docs audit log.
- Abort if: active docs still require `stage` or undocumented operator assumptions.
- Rollback: fix docs before release closeout.

#### RENAME-020A Zero Active Reference Sweep

- Goal: prove the spec closes with no active legacy references remaining outside historical or archival material.
- Preconditions: RENAME-011A, RENAME-011, RENAME-012, and the main cutover packets complete.
- Dry run: define the exact search scope for active code and maintained documentation.
- Real change: run the final sweep, classify any remaining hits, and remove or relabel them.
- Validation: all remaining `stage` hits are either historical, archival, or explicitly accepted migration history outside active operator guidance.
- Evidence: final search log and disposition table.
- Abort if: any active code path, maintained doc, or normative spec still carries `stage` as a current surface.
- Rollback: fix or relabel the remaining hits before closeout.

### Phase 8: Rollback, Release, And Closeout

#### RENAME-021 Rollback Drill

- Goal: rehearse a cutover rollback before public release is finalized.
- Preconditions: Gate F green and RENAME-015 complete.
- Dry run: tabletop the rollback steps against the recorded runbook.
- Real change: run one rollback rehearsal against the prepared release surfaces.
- Validation: rollback can restore the previous public contract within the recorded window.
- Evidence: rollback drill notes and time-to-rollback estimate.
- Abort if: rollback depends on undocumented manual recovery.
- Rollback: improve the rollback plan before release.

#### RENAME-022 Release Notes And Migration Guidance

- Goal: publish concise migration guidance that matches the actual cutover behavior.
- Preconditions: RENAME-021 and all verification tasks green.
- Dry run: verify release notes against the implemented contract and known issues.
- Real change: publish release notes and migration guidance.
- Validation: links, command table, and cleanup guidance are accurate.
- Evidence: final release note checklist.
- Abort if: release guidance overpromises compatibility or omits known cleanup steps.
- Rollback: correct the notes before announcing the release.

#### RENAME-023 Post-Release Verification Window

- Goal: keep the rename under observation long enough to catch real operator fallout.
- Preconditions: RENAME-022.
- Dry run: define the observation checklist and triage owner before release.
- Real change: run the post-release verification window and collect issues.
- Validation: any rename regressions are triaged and either fixed or documented before closeout.
- Evidence: closeout report in the spec or docs location chosen by the contract.
- Abort if: unresolved rename regressions remain in active operator paths.
- Rollback: invoke the rollback plan if regressions exceed the cutover threshold.

## Unsafe Task Combinations To Avoid

- Do not combine binary/install rename with public release asset publication.
- Do not combine CI invocation changes with cache-key rotation without a rehearsal.
- Do not combine docs sweep with a broad internal env/state/runtime rename; keep internal renames behavior-scoped and slice them first.
- Do not combine clean-machine and dirty-machine validation into one task.

## Recommended Execution Order

1. Freeze the contract.
2. Build the active-surface inventory.
3. Rehearse local build, install, PATH, and completion behavior.
4. Implement the external rename in narrow command and installer slices.
5. Rename the remaining active internal `stage` surfaces in narrow slices.
6. Align active docs and normative specs.
7. Rehearse CI, assets, installer retrieval, and rollback.
8. Apply public-facing CI and release changes.
9. Run clean-machine, dirty-machine, focused test, docs audits, and the zero-active-reference sweep.
10. Rehearse rollback, publish release guidance, and hold a post-release verification window.

## Approval Rule For This Phase

No phase advances unless the previous phase has both:

1. Passed its dry-run exit criteria.
2. Recorded evidence in the spec or linked validation notes.

If a dry run reveals a gap, add or split task packets before proceeding. Do not absorb the gap into a later implementation ticket.

## Risk Register Snapshot

- High risk: stale PATH entries causing users to run old binary unexpectedly.
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
