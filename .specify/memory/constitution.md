<!--
Sync Impact Report
Version change: 2.0.0 -> 2.0.1
Modified principles:
- Formatting and line wrapping only; no semantic changes
Added sections:
- None
Removed sections:
- None
Templates requiring updates:
- ✅ .specify/memory/constitution.md
- ✅ No dependent template or documentation changes required for this patch-only amendment
Follow-up TODOs:
- None
-->

# 20i Stack Constitution

## Core Principles

### I. Ease Of Use Is A Product Requirement
The primary operator experience MUST stay simple enough to use from memory.
Changes to commands, GUI wrappers, and future attach or detach commands MUST
preserve the shortest obvious path for the common workflow or provide a clearly
documented migration in the same change.
Any added friction MUST be justified by the pain it removes, and any deferred
complexity MUST be called out explicitly in the plan and docs.
Routine operations MUST prefer sensible defaults, low setup overhead, and
minimal manual coordination between shell, GUI, and automation entry points.

Rationale: this stack only creates value when local project startup and
inspection remain easier than hand-rolled Docker workflows.

### II. Reliability Must Be Boring And Predictable
The same input MUST produce the same operational result across shell commands,
automation wrappers, and repeated runs. Configuration precedence MUST remain
deterministic, documented, and operator-visible. New variables MUST use a single
canonical name, declare a default or required state, and avoid hidden coupling
to undocumented environment state. User-visible behavior changes MUST update all
affected operator surfaces in the same delivery unit.

Rationale: the stack is infrastructure, not a novelty feature; confidence comes
from repeatable behavior and documentation that matches reality.

### III. Robustness Must Hold Under Real Failure
Automation MUST handle partial startup, stale state, shared-service drift,
missing dependencies, and project-level breakage without corrupting unrelated
projects. Isolation of code mounts, naming, persistent state, and database data
is mandatory unless a resource is explicitly defined as shared infrastructure.
Changes affecting startup, routing, DNS, ports, attach or detach flows, or
teardown MUST expose actionable diagnostics through status, logs, health checks,
or explicit recovery instructions. Ambiguous state MUST be reported, not hidden.

Rationale: local infrastructure fails in messy ways, so resilience and recovery
clarity matter more than optimistic happy-path behavior.

### IV. Remove Pinch Points And User Friction
Features and process changes MUST actively reduce recurring operator pain rather
than shift it elsewhere. Any new step, prompt, manual sync, or stateful
exception MUST be justified in the plan, along with the friction it removes and
the simpler alternative considered. Where the repository and the deployed stack
copy under `$HOME/docker/20i-stack` can diverge, the workflow and docs MUST make
that sync point explicit so operators do not discover it by failure.

Rationale: this project exists to remove the repeated annoyances that make local
stack management slower, harder to remember, or easier to get wrong.

## Operational Constraints

- The primary supported operator environment is macOS with Docker Desktop and a
  POSIX shell workflow.
- Compose-based launches MUST continue to support invocation from an arbitrary
  project directory through the repo's documented environment contract, or the
  replacement contract MUST be documented and migration-tested.
- Changes that affect the working copy in this repository and the deployed stack
  copy under `$HOME/docker/20i-stack` MUST call out that sync requirement in the
  implementation plan and user-facing docs.
- Common-path operations MUST continue to fit within a shell-first workflow even
  when GUI wrappers or shared services are added.
- Development defaults such as local credentials, open ports, and phpMyAdmin
  exposure MUST remain clearly labeled as development-only behavior.
- Shared infrastructure additions MUST define bootstrap, steady-state, detach,
  teardown, and recovery expectations before implementation begins.

## Delivery Workflow & Quality Gates

- Every feature specification MUST identify the operator friction being removed
  or introduced, affected commands and interfaces, configuration precedence,
  state or isolation impact, and the documentation surfaces that need updating.
- Every implementation plan MUST pass a Constitution Check covering ease of use,
  reliability, robustness, friction removal, and operational validation.
- Every task list MUST include the work needed to keep docs and alternate entry
  points aligned when behavior changes, plus validation for any claimed
  reduction in operator friction.
- Changes to Compose files, runtime images, routing, or automation MUST be
  validated against startup, status/inspection, teardown, and at least one
  failure path relevant to the change. If validation cannot be run, the gap MUST
  be recorded explicitly.
- Complexity that violates this constitution MAY be approved only when the plan
  records the violation, the simpler rejected option, and the reason the extra
  complexity is necessary now.

## Governance

This constitution supersedes conflicting workflow guidance in repository docs and
Speckit templates. Amendments MUST update this file and any affected templates or
operator docs in the same change.

Versioning policy for this constitution follows semantic versioning:

- MAJOR: remove a principle, redefine a principle incompatibly, or weaken a
  governance requirement in a materially different way.
- MINOR: add a new principle or materially expand project-wide obligations.
- PATCH: clarify wording, tighten examples, or make non-semantic editorial fixes.

Compliance review expectations:

- Specs MUST show operator-facing impact, especially any friction removed,
  added, or deferred.
- Plans MUST document how the work satisfies the Constitution Check and why any
  added complexity is justified.
- Tasks and implementation reviews MUST confirm documentation parity, the
  required validation scope, and the operator path for recovery from failure.
- Unresolved non-compliance MUST be treated as a blocker until explicitly
  justified and accepted in the plan.

**Version**: 2.0.1 | **Ratified**: 2026-04-01 | **Last Amended**: 2026-04-01