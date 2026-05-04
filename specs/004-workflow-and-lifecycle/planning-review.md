# Spec 004 Planning Review

**Branch**: `004-workflow-and-lifecycle`
**Reviewed**: 2026-04-25
**Artifacts in scope**: `spec.md`, `plan.md`, `research.md`, `data-model.md`, `quickstart.md`, `handoff.md`, `contracts/workflow-lifecycle-contract.md`, `tasks.md`
**Cross-checked against**: `core/config/loader.go`, `core/config/types.go`, `core/lifecycle/orchestrator.go`, `core/lifecycle/errors.go`, `observability/status/status.go`, `infra/gateway/`, `docker-compose.shared.yml`, `docker-compose.yml`, `README.md`, `docs/runtime-contract.md`, `.env.example`, `.stackenv.example`.

This review only flags planning gaps. It does not change code or specs.

---

## 1. Summary

The spec, plan, and contract are internally coherent. The decision record is tight, the priority ordering of user stories is sound, and the failure/rollback contract is well-named. The two highest-risk gaps are:

- **Test files referenced by tasks do not exist yet** (`observability/status/status_test.go`, `core/lifecycle/errors_test.go`). Several "Add … in <file>" tasks will silently become "create the file" without scaffolding tasks.
- **The bootstrap precedence change is real code work, not just docs.** Today `STAGESERVE_POST_UP_COMMAND` is read from the merged precedence map (which includes shell env and `.env.stageserve`). The task list captures this in T013 but the contract surface (data-model + quickstart) does not include a negative-path assertion that proves shell-env injection is ignored.

Everything else is nits, missing-target enumeration, or scope clarifications.

---

## 2. Gaps In The Tasks List

### 2.1 Tasks reference test files that do not exist

| Task | Referenced file | Current state |
|---|---|---|
| T004 | `observability/status/status_test.go` | Does not exist |
| T017 | `core/lifecycle/errors_test.go` | Does not exist |
| T018 | `observability/status/status_test.go` | Does not exist |
| T019 | `observability/status/status_test.go` | Does not exist |

**Recommendation**: Reword these tasks from "Add … in <file>" to "Create <file> and add …". Or insert a one-line scaffolding task at the top of Phase 1 to create the empty test files explicitly. Otherwise Phase 1's "test the contract first" gate is trivially passable by skipping creation.

### 2.2 `WebNetworkAlias` default is missed by the rename task

`core/config/loader.go` line 277 derives `cfg.WebNetworkAlias = "stage-" + cfg.Slug + "-web"`. T009 only mentions "default project-scoped runtime naming … from `stage-` to `stage-`" without enumerating fields. The reader can miss `WebNetworkAlias` because `data-model.md` only lists `compose_project_name`, `runtime_network`, `database_volume`, and a vaguely worded `web_network_alias`.

**Recommendation**: Update T009 to enumerate the fields that change: `ComposeProjectName`, `WebNetworkAlias`, plus the derived `RuntimeNetwork`/`DatabaseVolume` (which come for free). Mirror the same enumeration in `data-model.md` so review evidence matches code.

### 2.3 `stage-gateway` alias inside the shared compose file is unclassified

`docker-compose.shared.yml` line 13 defines a network alias `stage-gateway` on the gateway service. The contract splits the world into "project-scoped → `stage-`" and "shared → `stage-shared`", but does not say which bucket `stage-gateway` falls into. Today gateway upstream rendering may rely on this exact alias.

**Recommendation**: In `contracts/workflow-lifecycle-contract.md`, add one line under "Shared resources" stating that the gateway service network alias remains `stage-gateway` (or moves), and add an assertion to T010 / T028 to confirm the rendered nginx upstream still resolves after rename.

### 2.4 Legacy `.env` fallback inside `loadStackEnv` is not explicitly killed

`core/config/loader.go` `loadStackEnv` falls back to `<stackHome>/.env` when `.stackenv` is absent. The handoff documents this as an intentional "legacy fallback". Plan §"Phase 2 step 2" says "Remove old-name handling", and T008 says "Remove old stack-default naming behavior". Neither explicitly names the legacy `<stackHome>/.env` fallback path. Per the workspace's legacy policy (`/memories/repo/legacy-policy.md`), it should be deleted.

**Recommendation**: In T008, name the two specific behaviors to remove: `(a) load from .stackenv`, `(b) fall back to <stackHome>/.env`. Add a test asserting that `<stackHome>/.env` is NOT loaded.

### 2.5 `STAGESERVE_POST_UP_COMMAND` source restriction is not asserted negatively

T012 says the hook is "sourced only from `.stage-local`". The current code reads `merged["STAGESERVE_POST_UP_COMMAND"]` so shell env, `.env.stageserve`, and project `.env` all leak through today.

**Recommendation**: Make T012 explicit about three negative-path tests:
- Setting via shell env → ignored
- Setting via `.env.stageserve` → ignored
- Setting via project `.env` → ignored
And cite which line in `loader.go` will need a per-key short-circuit (around the merged-map population at line 337).

### 2.6 No task touches `attach` despite US3 elevating it

US3's independent test "explicitly exercising `attach`". `Orchestrator.Attach` exists but no task verifies it under the new naming. T029 covers manual real-daemon attach; nothing in code-tests covers it.

**Recommendation**: Either add an attach-slice test task (preferred) or add an explicit note that attach behavior is validated only through the manual quickstart run, with that gap recorded per the "Explicit Validation Gap Policy" in the plan.

### 2.7 Documentation enumeration is incomplete

T025 / T026 / T027 cover `README.md`, `docs/runtime-contract.md`, `.env.example`, `.env.stageserve.example`. But the following surfaces also reference the old contract and are not enumerated:

- `core/config/types.go` line 109 docstring: `"-> shell env -> .stackenv/.env -> defaults"`
- `core/config/loader.go` package-level docstring line 5–8 (precedence comment block)
- `core/config/loader.go` lines 115–116 (`loadStackEnv` comment + filename literal)
- `README.md` line 193 (`.stackenv.example`)
- `docs/architecture.md` (not yet inspected; should be grep-checked)
- `docs/migration.md` (likely references old names)
- `CONTRIBUTING.md` (likely references old names)
- `handoff.md` itself (historical record, but contains contradictions to the new contract — see §3.4)

**Recommendation**: Add a single Phase 6 task: "Sweep `core/`, `docs/`, `README.md`, `CONTRIBUTING.md` for surviving `.stackenv` / `stage-<slug>` references and update or remove them. Code comments and docstrings count." This costs nothing if zero hits remain.

### 2.8 `.stackenv.example` deletion is implied but not stated

T011 / T027 say "retire `.stackenv.example` from the supported path". Per the repo's legacy policy, the file should be deleted, not left as a parallel example.

**Recommendation**: Reword to "Delete `.stackenv.example`."

### 2.9 No task addresses CLI surface (`cmd/stage/commands/up.go`, etc.)

If failure classification messaging changes (T020), the cobra layer that prints `StepError.Error()` may render new strings. No task verifies that the operator-visible output is still readable, and `--help` text in `up.go` is not reviewed for stale references.

**Recommendation**: Add one short verification step under T020 or T031 to `grep` `cmd/stage/commands/` for `stackenv|stage-<slug>` and to spot-check `stage up --help` output.

### 2.10 No bootstrap-timeout / bootstrap-cancel contract

Spec, contract, and data-model are silent on:
- Maximum runtime for a bootstrap command
- Behavior on `Ctrl-C` / context cancellation mid-bootstrap (does rollback still run?)
- Whether `WaitTimeoutSecs` covers bootstrap or only readiness

This is an actual lifecycle semantics gap, not just docs.

**Recommendation**: Either add a one-line decision to `research.md` ("Bootstrap inherits the operator's `Ctrl-C`; on cancel, project is rolled back") and a corresponding line to the contract, or list it under "Out of scope" explicitly. Today the answer is implicit and operators will guess.

### 2.11 No coverage for "registered as stable attached runtime" boundary

Spec edge case: *"A bootstrap failure occurs after containers become healthy but before the project is registered as a stable attached runtime."*

This is the trickiest rollback case. T021 ("rollback handling") covers it implicitly but no test or contract line ensures the registry / state store does not record an attached entry that is then orphaned.

**Recommendation**: Add an assertion to T021 / T019 spelled out as: "After bootstrap failure, the state store must not contain the project as `attached`. The registry projection must agree." That is testable today against the existing `state.StateStore`.

---

## 3. Inconsistencies And Unclear Language

### 3.1 `data-model.md` "Failure Classification" includes a label with no producer

The `class` field lists `application-follow-up`. No code path emits this class — it is a docs/policy label only. Operators or implementers will look for it in `errors.go`.

**Recommendation**: Mark it explicitly as "documentation-only label, never emitted by StageServe lifecycle code".

### 3.2 `data-model.md` "execution_target" and "working_directory" overstate guarantees

`execution_target: fixed to the apache service container` matches `orchestrator.go`. But `working_directory: project site root inside the container` is not actually asserted in code today (the bootstrap exec runs with whatever working directory the apache image defaults to). Either the code needs to set it explicitly, or the contract needs to soften.

**Recommendation**: Decide one. If the working directory matters for typical bootstrap commands like `php artisan migrate`, add a code task to set it explicitly. Otherwise drop the field from `data-model.md`.

### 3.3 Plan precedence vs. spec precedence wording

- `spec.md` §"Configuration & Precedence": *"CLI flags > .stage-local > shell environment > .env.stageserve > built-in defaults"*
- `contracts/workflow-lifecycle-contract.md` §"Precedence order": same five-step ordering.
- `core/config/loader.go` package docstring line 1–8: still describes the **old** precedence (`.env in the stack home`, no `.env.stageserve`).

The code docstring contradicts the spec.

**Recommendation**: Add explicit task to update `loader.go` docstring as part of T007 (currently only mentions behavior, not docstring).

### 3.4 `handoff.md` contradicts the 004 contract

`handoff.md` lines 37, 121, 124, 167 still describe `.stackenv` as the canonical name. The handoff is a historical sprint-closure artifact, but it lives inside the spec 004 directory and a reader following the directory in order will read contradictory facts.

**Recommendation**: Either (a) prepend a one-line banner to `handoff.md` ("Historical: superseded by spec 004 contract — see `contracts/workflow-lifecycle-contract.md`"), or (b) move it out of `specs/004-workflow-and-lifecycle/` into a sprint-history folder. Do not silently rewrite historical content.

### 3.5 Quickstart is inconsistent about `attach` for the first project

`quickstart.md` §"Multi-Project Validation" says `stage attach` for the second project but uses `stage up` for the first. Behaviorally `stage up` already implies attach, so this is correct, but it reads as if attach is special to project number two. The contract explicitly elevates `attach` ("must execute `attach` explicitly"). A first-time reader will think attach was tested only once.

**Recommendation**: Add one sentence: "The first project's `stage up` is the implicit attach for project one; the explicit `stage attach` for project two is what proves multi-project routing."

### 3.6 Tasks "Tests" preamble says tests are required, plan §Phase 1 step 4 says re-check constitution

The tasks file's preamble: *"Focused Go tests are required for touched slices …"*. The plan's §"Implementation Plan" Phase 0–4 ordering implies tests-first; Phase 1 of tasks mirrors this. But the plan's §"Validation Strategy" lists "Add or update tests that cover" — which uses the same wording for new and pre-existing tests. There is no marker for which tests already exist vs. which are net-new.

**Recommendation**: In tasks.md, prefix net-new test files with `(new file)` so reviewers know not to look for them in `git log`.

### 3.7 Spec "Out of scope" coverage is implicit, not listed

Multiple sentences describe what is NOT in scope (multi-phase lifecycle, framework-specific bootstrap, app migration repair, release pipeline work) but they are scattered across `spec.md` Assumptions and `handoff.md`. There is no "Out Of Scope" section in `spec.md`.

**Recommendation**: Add a 4–6 line "Out Of Scope" section to `spec.md` listing: additional lifecycle phases, framework-specific bootstrap helpers, application migration repair, release/distribution work, TUI/UI surface changes, automation of real-daemon validation.

### 3.8 `tasks.md` calls dependency T006 → T005 but groups them in different phases

T005 is foundational decision codification; T006 is gateway golden test update. Both are listed in Phase 1 but Phase 1's "Checkpoint" says the suite "names the final contract before implementation begins" — golden tests usually need the implementation to regenerate. There is risk the gateway golden tests will simply diff against the not-yet-changed code.

**Recommendation**: Move T006 to Phase 2 immediately after T009/T010, or explicitly note in T006 that the goldens are written by hand (not regenerated) to express the target.

### 3.9 `STACK_HOME` env var is not addressed

`STACK_HOME` is referenced in `core/config/loader.go` and is part of the operator surface. No task confirms whether the resolution of `<stackHome>/.env.stageserve` is correct after rename. Likely fine because `loadStackEnv` already takes a `stackHome` param, but worth one assertion.

**Recommendation**: Add a single test case to T001 covering `STACK_HOME` override pointing at a directory containing `.env.stageserve`.

---

## 4. Scope Boundary Issues

### 4.1 In scope (clear)

- Bootstrap contract lock-down (one phase, project-local source, rollback on failure, named failure class)
- Stack-defaults rename `.stackenv` → `.env.stageserve` and removal of legacy fallbacks
- Project-scoped runtime rename `stage-<slug>` → `stage-<slug>`
- Documentation parity across `README.md`, `docs/runtime-contract.md`, contract, examples, quickstart
- Real-daemon validation across one single-project app and one multi-project scenario, including explicit `attach`

### 4.2 Out of scope but not stated as such

- New lifecycle phases beyond `post-up`
- Framework-specific bootstrap helpers
- Application migration / schema repair
- Release/distribution pipeline
- TUI / GUI surface changes
- Automating the manual real-daemon validation in CI

These are visible only by reading `handoff.md` and Assumptions. See §3.7.

### 4.3 Ambiguous boundary

- **`stage-gateway` alias** (§2.3) — not classified as shared or project-scoped.
- **Bootstrap timeout / cancel semantics** (§2.10) — implicit.
- **Bootstrap working directory** (§3.2) — claimed but not asserted.
- **Legacy `<stackHome>/.env` fallback** (§2.4) — implicitly killed.

---

## 5. Risks Plan Already Names (For Reference)

The plan's existing risks table is well-formed and covers:
- Shared vs. project-scoped naming drift
- `.env.stageserve` vs project `.env` ownership confusion
- Stale rollback state
- Informal real-project validation

These are accurate. No changes recommended to the risks table itself; the gaps in §2 above are *uncovered* risks.

---

## 6. Recommended Remediation (Prioritised)

### Must-do before implementation starts

1. **Reword T004 / T017 / T018 / T019** to "create new file" rather than "Add … in" (§2.1).
2. **Enumerate the rename surface in T009** to include `WebNetworkAlias` and any other `stage-<slug>-*` derivations (§2.2).
3. **Strengthen T012** with explicit negative-path tests for shell env, `.env.stageserve`, and project `.env` sources of `STAGESERVE_POST_UP_COMMAND` (§2.5).
4. **Strengthen T008** to name both the `.stackenv` reader and the `<stackHome>/.env` legacy fallback for removal, and add a test (§2.4).
5. **Resolve `stage-gateway` alias classification** in the contract before T010 lands (§2.3).
6. **Decide bootstrap timeout/cancel contract** in `research.md` and contract (§2.10), or list explicitly as out of scope.

### Should-do during implementation

7. **Add a `cmd/stage/commands/` sweep** to T031 / Phase 6 (§2.9).
8. **Add the documentation-surface sweep task** to Phase 6 covering `core/config/types.go` docstring, `loader.go` package docstring, `docs/architecture.md`, `docs/migration.md`, `CONTRIBUTING.md` (§2.7, §3.3).
9. **State `.stackenv.example` is deleted** in T011/T027 (§2.8).
10. **Add an attach-slice test or an explicit recorded gap** for US3 (§2.6).
11. **Add a state-store assertion** for "rolled-back project not attached" to T019/T021 (§2.11).

### Nice-to-have polish

12. **Add an Out Of Scope section to `spec.md`** (§3.7).
13. **Annotate `handoff.md`** as historical (§3.4).
14. **Resolve bootstrap working-directory claim** in `data-model.md` (§3.2).
15. **Mark `application-follow-up` failure class as docs-only** in `data-model.md` (§3.1).
16. **Tighten quickstart language** about implicit attach for project one (§3.5).
17. **Move T006 out of Phase 1** or annotate it as hand-written goldens (§3.8).
18. **Add `STACK_HOME` override test case** to T001 (§3.9).

---

## 7. What Looks Good

- The decision record in `research.md` is unusually clean: every decision has a rejected alternative attached.
- The user story priority order (US1 → US2 → US3) matches actual implementation dependency, and the MVP boundary (US1+US2) is the right stopping point if scope must be cut.
- The contract document is the single source of truth and is the right artifact to point implementers at.
- The "Repo-To-Deployed-Copy Sync Requirement" in the plan is a real operator hazard surfaced explicitly — that is the kind of thing usually missed.
- Per-project rollback isolation (FR-006, T019) is correctly elevated; that was the most expensive lesson from sprint 003.
- Failure classification is a StageServe-vs-app boundary, not a try-harder-on-app boundary — this is correctly stated in spec, contract, and data-model.

---

## 8. Suggested Single-Sentence Patches

If only the highest-impact remediations land, these three lines fix the structural gaps:

1. In `tasks.md` T009: append `"including WebNetworkAlias and any other derived stage-<slug>-* names"`.
2. In `tasks.md` T012: append `"and assert STAGESERVE_POST_UP_COMMAND set via shell env, .env.stageserve, or project .env is ignored"`.
3. In `contracts/workflow-lifecycle-contract.md` §"Shared resources": add `"The gateway service network alias remains stage-gateway."` (or whatever the resolved decision is).

The rest of §6 is cleanup and documentation parity, not contract correctness.
