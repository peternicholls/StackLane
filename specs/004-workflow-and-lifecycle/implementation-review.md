# Implementation Review: Spec 004 Workflow And Lifecycle

Date: 2026-04-25
Reviewer: GitHub Copilot (GPT-5.4)
Scope: branch delta against `master`, plus current workspace state after implementation and validation

## Verdict

Needs follow-up before I would call the implementation fully finished.

The runtime-facing code changes are mostly correct and well-covered by tests. The main gaps are in finish work: some operator docs still reference a redundant project-local example file that should remain deleted, operator docs are only partially aligned, and the spec task ledger was not updated to reflect the completed work.

## Validation Performed

- Compared the branch delta against `master` with `git diff master...HEAD`.
- Re-ran the existing validation already performed on the branch:
  - `go test ./...`
  - `go build -o stage-bin ./cmd/stage`
  - CLI help spot-check for `up`, `status`, `down`, `attach`, `logs`
  - `go vet ./...`
  - `gofmt`
- Re-checked the current workspace contents, including the current state of [README.md](../../../README.md), [docs/runtime-contract.md](../../../docs/runtime-contract.md), and [specs/004-workflow-and-lifecycle/tasks.md](./tasks.md).

## Findings

### 1. Stale references to a redundant project-local example file

Severity: high

The repository now contains [`.env.stageserve.example`](../../../.env.stageserve.example), and [README.md](../../../README.md#L194) still tells operators there is also a project-local `.env.example` to copy into a project as `.stage-local` at [README.md](../../../README.md#L195). Per the current constraint, `.env.example` is intentionally redundant and should remain deleted; the bug is that the docs and task wording still describe it as a live surface.

Impact:

- The documented operator path is broken.
- T011 and T027 in [tasks.md](./tasks.md#L62) and [tasks.md](./tasks.md#L127) are now misleading because they still treat `.env.example` as part of the intended deliverable.

Recommendation:

- Remove the stale `.env.example` references from the operator docs and update the task wording so the deliverable is explicit: keep `.env.stageserve.example`, delete `.stackenv.example`, and do not carry a redundant project-local compatibility template.

### 2. README does not clearly describe the central stack authority model

Severity: medium

[README.md](../../../README.md) correctly documents `.env.stageserve`, `stage-<slug>`, and the `.stage-local` restriction for `STAGESERVE_POST_UP_COMMAND`. But it still does not clearly explain the intended operating model: there is one central StageServe-managed 20i-style stack authority, and individual projects are expected to tap into that shared authority rather than define their own stack shape.

That missing explanation matters because StageServe has a real operator split between:

- the repository working copy where code and docs are edited
- the live StageServe installation on `PATH` that embodies the shared 20i-style stack contract projects plug into

Without that explanation, an operator can misread the product model in two ways:

- assume each project owns or defines the stack contract, when the intent is that the stack contract is centrally owned and projects only provide project-local inputs
- run validation from an edited repo checkout and treat the result as authoritative even when the live stack installation on `PATH` has not been synced and rebuilt yet

In this environment, the known live copy appears to be under `$HOME/docker/20i-stack`, but that path is an instance-specific example of the broader model rather than the model itself.

Impact:

- Operators can follow the branch docs and still validate from the wrong working copy.
- The README is no longer fully aligned with [quickstart.md](./quickstart.md#L9), which does carry the deployed-copy warning.
- This creates noisy false negatives during manual validation: behavior seen at runtime may reflect stale binaries, stale compose templates, or stale example/config files rather than the branch under review.
- The README also misses a core product framing point: projects should "tap into" the shared 20i-style stack authority, not recreate or reinterpret it locally.

Recommendation:

- Add one short note near the validation or command semantics section stating that StageServe is the central authority for the 20i-style local stack contract, while projects contribute only project-local configuration and content. Then add a follow-on note that if this repository checkout is not the live StageServe install on `PATH`, operators must sync the changed files and rebuild `stage-bin` in that live copy before treating manual runtime results as evidence for or against spec 004. `$HOME/docker/20i-stack` can be mentioned as the known local example, not as a universal product path.

### 3. Operator-facing docs expose too much gateway implementation detail

Severity: medium

[docs/runtime-contract.md](../../../docs/runtime-contract.md#L133) and [README.md](../../../README.md#L151) currently expose shared gateway naming details such as `stage-shared` and `stage-gateway` directly to operators. That may be correct as an internal implementation contract, but it does not match the intended product model: StageServe should detect whether the shared gateway layer is running, start or reconcile it when needed, and keep ordinary operators focused on projects and routes rather than on gateway internals.

Impact:

- Operator-facing docs are carrying implementation detail that most users should not need to reason about.
- The current wording risks teaching users to manage the gateway directly instead of treating it as a StageServe-owned subsystem.
- T030 parity is still incomplete, but the bigger issue is abstraction level, not just naming completeness.

Recommendation:

- Split the documentation surface by audience.
- In operator-facing docs such as [README.md](../../../README.md) and the main sections of [docs/runtime-contract.md](../../../docs/runtime-contract.md), describe behavior in user terms: StageServe ensures shared routing is available, reuses it when already running, and heals it when missing.
- Keep names like `stage-shared` and `stage-gateway` in lower-level contract or troubleshooting sections only, where advanced users may need them.

### 4. Task bookkeeping was not updated

Severity: low

The implementation appears to cover T008, T009, most of T011, T025, T030a, and T030b, but those items are still unchecked in [tasks.md](./tasks.md#L59), [tasks.md](./tasks.md#L60), [tasks.md](./tasks.md#L62), [tasks.md](./tasks.md#L125), [tasks.md](./tasks.md#L139), [tasks.md](./tasks.md#L140), and [tasks.md](./tasks.md#L141).

Impact:

- The spec execution trail is no longer trustworthy.
- Reviewers cannot tell which work is genuinely outstanding versus completed but unrecorded.

Recommendation:

- After fixing the stale `.env.example` references and the remaining doc parity gaps, mark the completed tasks `[X]` and leave only genuinely unfinished items open.

## File-By-File Notes

### Runtime and config

- [core/config/loader.go](../../../core/config/loader.go): pass. Root-cause fix is in the correct layer. Removing legacy stack-default fallbacks and excluding `STAGESERVE_POST_UP_COMMAND` from shell/stack merges matches the contract and avoids a shim.
- [core/config/types.go](../../../core/config/types.go): pass. The docstring now matches the actual precedence chain.
- [core/config/loader_test.go](../../../core/config/loader_test.go): pass. Strong coverage added for negative legacy-path behavior and for the project-local-only post-up hook rule.

### Lifecycle and rollback coverage

- [core/lifecycle/errors_test.go](../../../core/lifecycle/errors_test.go): pass. Useful focused coverage for step classification, especially `post-up-hook`.
- [core/lifecycle/orchestrator_internal_test.go](../../../core/lifecycle/orchestrator_internal_test.go): pass. Straight rename alignment to the new `stage-` defaults.
- [core/lifecycle/orchestrator_test.go](../../../core/lifecycle/orchestrator_test.go): pass with a note. The new rollback-isolation and attach tests are valuable and scoped correctly. I re-checked the current file state because it had been reformatted after the earlier edits; no behavioral issue stands out.

### State, status, and gateway

- [core/state/store_test.go](../../../core/state/store_test.go): pass. Naming updates are consistent with the loader default change.
- [observability/status/status_test.go](../../../observability/status/status_test.go): pass. The new tests cover the exact phantom-state risk that matters after rollback.
- [infra/gateway/manager_test.go](../../../infra/gateway/manager_test.go): pass. The route alias rename is correctly reflected in tests.
- [infra/gateway/testdata/multi-route-no-tls.conf](../../../infra/gateway/testdata/multi-route-no-tls.conf): pass. Golden data is aligned with the `stage-` rename.
- [infra/gateway/testdata/single-route-tls.conf](../../../infra/gateway/testdata/single-route-tls.conf): pass. Golden data is aligned with the `stage-` rename.

### Operator docs and examples

- [README.md](../../../README.md): partial. Good on naming and bootstrap restriction; still missing a clear statement that StageServe centrally defines the 20i-style stack contract that projects tap into, plus the deployed-copy sync note, and it still references a redundant project-local example file that should stay deleted.
- [docs/runtime-contract.md](../../../docs/runtime-contract.md): partial. Good on `stage-` and rollback semantics, but it should present the shared gateway as a StageServe-managed internal subsystem for normal operators and reserve explicit `stage-shared` / `stage-gateway` naming for contract or troubleshooting depth.
- [docs/migration.md](../../../docs/migration.md): pass. The rename note is explicit and historically framed.
- [`.env.stageserve.example`](../../../.env.stageserve.example): pass. Good explanation of scope and the FR-016 restriction.
- `.env.example`: intentionally absent. No action needed to restore it; action is needed to stop advertising it.
- `.stackenv.example`: pass. Removal is correct.

### Incidental or non-functional changes

- [go.mod](../../../go.mod): neutral. `github.com/containerd/errdefs` moved from indirect to direct. I did not find a regression tied to this change, but it is incidental to the spec 004 implementation rather than a core behavior change.
- [infra/compose/types.go](../../../infra/compose/types.go): neutral. Formatting-only diff from `gofmt`; no behavior change.

### Spec and process artifacts

- [specs/004-workflow-and-lifecycle/tasks.md](./tasks.md): needs follow-up. The task ledger no longer reflects actual completion state.
- [specs/004-workflow-and-lifecycle/quickstart.md](./quickstart.md): pass with a note. This file is the clearest operator-facing validation script in the spec set and already includes the deployed-copy sync note, but the top-level README still needs to carry the higher-level "central authority / projects tap in" framing.
- [specs/004-workflow-and-lifecycle/contracts/workflow-lifecycle-contract.md](./contracts/workflow-lifecycle-contract.md): pass. This remains the most precise written statement of the new contract.
- Other spec artifacts under [specs/004-workflow-and-lifecycle](./): reviewed only for parity and bookkeeping, not for implementation correctness.

## Recommendations

1. Remove all stale `.env.example` references from docs and task wording; do not restore the file if the intended contract is to avoid the redundant project-local template.
2. Update [README.md](../../../README.md) to state explicitly that StageServe is the central authority for the 20i-style stack contract and that projects tap into it via project-local config. Then add the deployed-copy sync note so the top-level operator guide matches [quickstart.md](./quickstart.md).
3. Refactor the shared-routing docs so operator-facing guidance talks about StageServe automatically ensuring routing is available, while explicit gateway resource names are moved to advanced contract or troubleshooting material.
4. Mark completed tasks in [tasks.md](./tasks.md) after the file/example/doc cleanup is finished.
5. Keep the current test additions. They cover the most failure-prone parts of the spec and are worth preserving even if the docs/example cleanup lands separately.
