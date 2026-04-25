# Research: Workflow And Lifecycle Hardening

## Decision 1: Keep One Bootstrap Phase

- Decision: Keep a single documented post-up bootstrap phase.
- Rationale: The current runtime already runs `STACKLANE_POST_UP_COMMAND` after readiness in `core/lifecycle/orchestrator.go`. Locking that one phase avoids reopening lifecycle design while the first phase is still being documented and validated.
- Alternatives considered:
  - Add post-attach or additional named phases now. Rejected because it broadens scope and multiplies failure semantics before the current behavior is fully hardened.
  - Add framework-specific lifecycle phases. Rejected because the spec explicitly keeps the core runtime framework-agnostic.

## Decision 2: Keep Bootstrap Project-Local

- Decision: Source bootstrap behavior only from `.stacklane-local`.
- Rationale: Bootstrap commands are project behavior, not stack-wide defaults. Keeping the hook project-local makes the operator intent explicit and prevents one stack-level setting from silently affecting unrelated repos.
- Alternatives considered:
  - Allow shell environment overrides. Rejected because it hides a project-specific lifecycle behavior in ambient operator state.
  - Allow stack-wide defaults in the stack-owned env file. Rejected because it weakens the ownership boundary between stack defaults and app/project behavior.

## Decision 3: Roll Back On Bootstrap Failure

- Decision: Keep rollback mandatory when the post-up bootstrap command fails.
- Rationale: The current orchestrator already rolls the project back on hook failure. Keeping that behavior makes lifecycle outcomes boring and repeatable: either the project comes up cleanly or the runtime returns to a known stopped state.
- Alternatives considered:
  - Preserve a failed-but-inspectable state. Rejected because it would add a second recovery mode and complicate status semantics.
  - Make rollback configurable. Rejected because it introduces more state combinations than the current workflow needs.

## Decision 4: Rename Stack-Owned Defaults To `.env.stacklane`

- Decision: Use `.env.stacklane` as the only supported stack-owned defaults file.
- Rationale: The name is visually explicit, reads as an env file in editor workflows, and keeps stack-owned settings distinct from both project `.env` and `.stacklane-local`.
- Alternatives considered:
  - Keep `.stackenv`. Rejected because it is less clear in the repo and weaker in editor tooling.
  - Reuse project `.env`. Rejected because project `.env` is application-owned, not generic Stacklane configuration.

## Decision 5: Shorten Project-Scoped Runtime Names To `stln-`

- Decision: Change project-scoped runtime naming defaults from `stacklane-` to `stln-`.
- Rationale: Docker resource lists are easier to scan when the runtime prefix is shorter and leaves more room for the actual project slug. This matters most in multi-project workflows, where the current prefix consumes too much of the operator-visible resource name.
- Alternatives considered:
  - Keep `stacklane-`. Rejected because it adds no functional value and reduces scanability.
  - Use uppercase `STLN-`. Rejected because current compose project naming and downstream runtime naming are lowercase-oriented.

## Decision 6: Keep Shared Resources Explicit

- Decision: Keep the shared-gateway compose project name and shared network explicit as `stacklane-shared`.
- Rationale: The `stln-` shortening applies only to project-scoped runtime resources. Shared infrastructure already represents a distinct cross-project surface, so keeping `stacklane-shared` preserves that boundary and avoids conflating shared services with per-project runtimes.
- Alternatives considered:
  - Rename only project-scoped resources and leave shared names implicit. Rejected because it would leave the contract underspecified.
  - Rename every shared and project-scoped resource mechanically. Rejected because the shared boundary is meaningful and should remain explicit.

## Decision 7: Keep Real-Daemon Validation As A Required Deliverable

- Decision: Require one representative app workflow and one multi-project scenario as part of completion criteria.
- Rationale: Spec 003 already proved that mocked tests alone were not enough to flush out workflow gaps. This feature is specifically about operator workflow and lifecycle behavior, so it must be checked against real daemon behavior.
- Alternatives considered:
  - Rely only on unit tests. Rejected because the feature’s main risks live at the integration boundary.
  - Defer validation to a later docs-only pass. Rejected because that would separate the workflow contract from the evidence that it actually works.