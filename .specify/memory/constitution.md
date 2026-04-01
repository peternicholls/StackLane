<!--
Sync Impact Report
Version change: none -> 1.0.0
Modified principles:
- Initial ratification; no prior constitution existed
Added sections:
- Core Principles
- Operational Constraints
- Delivery Workflow & Quality Gates
- Governance
Removed sections:
- None
Templates requiring updates:
- ✅ .specify/memory/constitution.md
- ✅ .specify/templates/plan-template.md
- ✅ .specify/templates/spec-template.md
- ✅ .specify/templates/tasks-template.md
- ✅ No .specify/templates/commands/*.md files exist in this repository
Follow-up TODOs:
- None
-->

# 20i Stack Constitution

## Core Principles

### I. Shell Workflow Is The Primary Contract
The shell commands are the authoritative operator interface for this project.
Changes to `20i-up`, `20i-down`, `20i-status`, `20i-logs`, `20i-gui`, or any
future attach/detach commands MUST preserve documented behavior or ship with an
explicit migration note in the same change. GUI and AppleScript automation MUST
remain wrappers around the same runtime contract and MUST NOT invent divergent
semantics, defaults, or state models.

Rationale: this stack is adopted for low-friction local operations; once the
shell contract drifts, every wrapper, alias, and habit breaks at once.

### II. Project Isolation Is Non-Negotiable
Every project runtime MUST isolate code mounts, container naming, persistent
state, and database data from other projects unless a resource is explicitly
designed as shared infrastructure. Any change that introduces shared routing,
DNS, or control-plane services MUST document the isolation boundary, teardown
behavior, and drift recovery path. No feature may silently reuse another
project's writable state.

Rationale: this repository exists to run multiple local PHP projects safely; a
faster workflow is not acceptable if it risks cross-project contamination.

### III. Configuration Must Be Explicit And Precedence-Driven
All operator-facing behavior MUST be configurable through a documented and
deterministic precedence chain. The default precedence for this project is CLI
override, then `.20i-local`, then global stack defaults. New variables MUST use
a single canonical name, include a default or required-state declaration, and be
documented wherever operators are expected to set them. Hidden environment
coupling and implicit path assumptions are prohibited.

Rationale: this stack is launched from arbitrary project directories and is
expected to remain predictable under per-project overrides.

### IV. Documentation And Interface Parity Are Required
Any change to user-visible behavior MUST update every affected operator surface
in the same delivery unit. This includes `README.md`, `AUTOMATION-README.md`,
`GUI-HELP.md`, shell help text, and GUI/AppleScript messaging when they describe
the changed workflow. Examples, defaults, hostnames, ports, and setup steps MUST
match shipped behavior.

Rationale: operational drift is one of the fastest ways to make local tooling
feel unreliable, especially when this repository already exposes multiple entry
points.

### V. Failure Visibility Before Convenience
Automation MUST fail loudly with actionable diagnostics. Changes affecting
startup, status, routing, DNS, ports, or teardown MUST provide a clear operator
inspection path through logs, status output, or health checks. Destructive or
wide-scope actions MUST be explicitly scoped, and ambiguous state MUST be
reported instead of guessed away.

Rationale: local infrastructure fails in messy ways; the project must optimize
for recovery speed and operator clarity rather than silent magic.

## Operational Constraints

- The primary supported operator environment is macOS with Docker Desktop and a
  POSIX shell workflow.
- Compose-based launches MUST continue to support invocation from an arbitrary
  project directory through the repo's documented environment contract, or the
  replacement contract MUST be documented and migration-tested.
- Changes that affect the working copy in this repository and the deployed stack
  copy under `$HOME/docker/20i-stack` MUST call out that sync requirement in the
  implementation plan and user-facing docs.
- Development defaults such as local credentials, open ports, and phpMyAdmin
  exposure MUST remain clearly labeled as development-only behavior.
- Shared infrastructure additions MUST define bootstrap, steady-state, detach,
  teardown, and recovery expectations before implementation begins.

## Delivery Workflow & Quality Gates

- Every feature specification MUST identify affected commands/interfaces,
  configuration precedence, state or isolation impact, and the documentation
  surfaces that need updating.
- Every implementation plan MUST pass a Constitution Check covering command
  compatibility, isolation boundaries, configuration precedence, documentation
  parity, and operational validation.
- Every task list MUST include the work needed to keep docs and alternate entry
  points aligned when behavior changes.
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

- Specs MUST show the operator-facing impact of the change.
- Plans MUST document how the work satisfies the Constitution Check.
- Tasks and implementation reviews MUST confirm documentation parity and the
  required validation scope.
- Unresolved non-compliance MUST be treated as a blocker until explicitly
  justified and accepted in the plan.

**Version**: 1.0.0 | **Ratified**: 2026-04-01 | **Last Amended**: 2026-04-01