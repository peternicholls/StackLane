# Recommendation Tasks: Review Follow-Up For Spec 004

Date: 2026-04-25
Source: [recommendation-plan.md](./recommendation-plan.md)

## Execution Rules

- Execute tasks in order unless a task is explicitly marked `[P]`.
- Do not reopen loader, lifecycle, or gateway behavior unless a validation step proves the docs no longer match reality.
- Treat `.env.example` as intentionally deleted; the task is to remove stale references, not restore the file.
- Treat gateway resource names as advanced/internal detail; operator-facing wording should describe behavior first.

## Tasks

- [X] RT001 Remove stale `.env.example` references from operator-facing docs.
  Deliverables: [README.md](../../../README.md) and any other operator-facing doc no longer instruct operators to use or expect `.env.example`.
  Verification: `rg -n "\.env\.example" README.md docs/ specs/004-workflow-and-lifecycle` returns only intentional historical or planning references.

- [X] RT002 Reword T011 and T027 in [tasks.md](./tasks.md) so they match the intended contract.
  Deliverables: T011 and T027 no longer describe `.env.example` as an active deliverable unless explicitly justified.
  Verification: read [tasks.md](./tasks.md) and confirm the wording says `.env.stacklane.example` remains, `.stackenv.example` is deleted, and `.env.example` is not being restored as compatibility surface.

- [X] RT003 Add central-authority product framing to [README.md](../../../README.md).
  Deliverables: README states that Stacklane defines the 20i-style stack contract and projects tap into it through project-local config.
  Verification: a reviewer can open [README.md](../../../README.md) and find one concise statement that distinguishes shared stack authority from project-local customization.

- [X] RT004 Add the deployed-copy validation note to [README.md](../../../README.md).
  Deliverables: README explains that the live Stacklane installation on `PATH` is the authoritative runtime surface for manual validation.
  Verification: [README.md](../../../README.md) contains a note that tells operators to sync/rebuild the live install before using runtime behavior as validation evidence; if `$HOME/docker/20i-stack` is mentioned, it is clearly described as a local example rather than a universal path.

- [X] RT005 Rewrite operator-facing routing language in [README.md](../../../README.md) to describe behavior rather than gateway internals.
  Deliverables: README says Stacklane ensures shared routing is available, reuses it when running, and repairs it when missing.
  Verification: main operator sections of [README.md](../../../README.md) do not teach gateway management as a normal workflow.

- [X] RT006 Rewrite operator-facing routing language in [docs/runtime-contract.md](../../../docs/runtime-contract.md) to describe behavior rather than gateway internals.
  Deliverables: runtime-contract describes routing in user terms for the main operator flow.
  Verification: the main sections of [docs/runtime-contract.md](../../../docs/runtime-contract.md) describe Stacklane-managed routing behavior and do not require ordinary operators to reason about `stln-shared` or `stln-gateway`.

- [X] RT007 [P] Preserve advanced/internal routing details in lower-level material.
  Deliverables: advanced names such as `stln-shared` and `stln-gateway` remain discoverable in contract or troubleshooting depth where appropriate.
  Verification: [contracts/workflow-lifecycle-contract.md](./contracts/workflow-lifecycle-contract.md) and any troubleshooting-focused sections still contain the necessary internal naming details for advanced users.

- [X] RT008 Reconcile completion state in [tasks.md](./tasks.md).
  Deliverables: T008, T009, T011, T025, T027, T030, T030a, and T030b are marked to match actual repository state after the doc cleanup lands.
  Verification: a reviewer can compare the repository state to [tasks.md](./tasks.md) and see no obvious mismatch between completed work and unchecked tasks.

- [X] RT009 Run stale-reference grep checks after doc edits.
  Deliverables: grep evidence for `.env.example`, `.stackenv`, `stln-shared`, and `stln-gateway` is reviewed and classified.
  Verification: run:
  `rg -n "\\.env\\.example|\\.stackenv|stln-shared|stln-gateway" README.md docs/ specs/004-workflow-and-lifecycle core/ infra/ cmd/ internal/`
  Then confirm each remaining hit is either intentional advanced/internal documentation or an explicitly historical artifact.

- [X] RT010 Re-run lightweight repo validation.
  Deliverables: follow-up doc cleanup is confirmed not to have introduced unrelated breakage or drift.
  Verification:
  `go test ./...`
  `go build -o stacklane-bin ./cmd/stacklane`

- [X] RT011 Perform a final operator-surface spot check.
  Deliverables: README, runtime-contract, quickstart, and tasks tell a coherent story.
  Verification: manually read [README.md](../../../README.md), [docs/runtime-contract.md](../../../docs/runtime-contract.md), [quickstart.md](./quickstart.md), and [tasks.md](./tasks.md) and confirm they agree on all of the following:
  `.env.stacklane` is the stack-owned defaults file.
  `.env.example` is not part of the supported surface.
  Stacklane centrally defines the 20i-style stack model.
  Projects tap into that shared model with project-local config.
  Shared routing is a Stacklane-managed internal subsystem for normal operators.

## Exit Criteria

The follow-up is complete when all tasks above are checked and the repository satisfies all of the following:

- no stale operator-facing `.env.example` references remain
- README explains the central-authority stack model clearly
- operator docs describe routing behavior without forcing gateway internals into the normal user path
- advanced/internal gateway names remain available where needed
- [tasks.md](./tasks.md) is trustworthy again
- `go test ./...` and `go build -o stacklane-bin ./cmd/stacklane` succeed
- visual checking of launching§ and routing behavior matches the docs, and the README, runtime-contract, quickstart, and tasks all tell a consistent story. 
- Multiple sites can be attached and routed through the shared gateway without conflict, and a bootstrap failure in one project does not affect the other's attachment or routing.

§ Launching is defined as `stacklane up` with a project attached.