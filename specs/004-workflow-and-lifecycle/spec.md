# Feature Specification: Workflow And Lifecycle Hardening

**Feature Branch**: `004-workflow-and-lifecycle`  
**Created**: 2026-04-24  
**Status**: Draft  
**Input**: User description: "Close sprint 003 with a handoff that captures the workflow and lifecycle gaps discovered during live validation, carry those into a focused follow-up spec, rename `.stackenv` to `.env.stageserve`, and shorten runtime-owned Docker resource prefixes from `stage-` to `stage-`."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Bootstrap A Project Predictably (Priority: P1)

An operator starts a real application with `stage up` and expects StageServe to either finish the documented bootstrap work or fail in a named, recoverable way. The operator should not have to remember undocumented follow-up commands after containers become healthy.

**Why this priority**: This is the highest-value workflow gap left after spec 003. If startup still depends on ad hoc manual migration or setup steps, the runtime is not yet predictable enough for daily use.

**Independent Test**: Configure a project-local bootstrap command, run `stage up` from a stopped state, and confirm StageServe either completes the bootstrap path successfully or reports a bootstrap-specific failure and rolls the project runtime back.

**Acceptance Scenarios**:

1. **Given** a project defines a bootstrap command in project-root `.env.stageserve`, **When** the operator runs `stage up`, **Then** StageServe runs that command after StageServe-owned readiness succeeds and surfaces the result as part of the lifecycle outcome.
2. **Given** a project defines a bootstrap command that fails, **When** the operator runs `stage up`, **Then** StageServe reports a named bootstrap failure, rolls the project runtime back, and provides a documented recovery path.
3. **Given** a project does not define a bootstrap command, **When** the operator runs `stage up`, **Then** StageServe completes the normal runtime lifecycle without inventing an implicit framework-specific bootstrap step.

---

### User Story 2 - Distinguish StageServe Failures From App Failures (Priority: P1)

An operator validating a project needs to know whether a failure belongs to StageServe infrastructure or to the application running inside it. StageServe should own its gateway, routing, runtime env injection, and lifecycle reporting without pretending that every application-level error is a StageServe defect.

**Why this priority**: Lifecycle hardening is not only about starting containers. It is also about making failure boundaries legible, so validation work does not collapse into unbounded app repair.

**Independent Test**: Run one project that has healthy StageServe infrastructure but an application-level bootstrap or migration problem, then confirm the runtime reports StageServe readiness separately from the application defect and preserves the documented ownership boundary.

**Acceptance Scenarios**:

1. **Given** StageServe-owned health checks succeed but an application bootstrap command fails, **When** the operator inspects the lifecycle result, **Then** the failure is reported as a bootstrap lifecycle failure rather than as a gateway, DNS, or generic health error.
2. **Given** a project route responds with an application-specific error after StageServe readiness has passed, **When** the operator validates the project, **Then** the documented workflow distinguishes StageServe runtime success from application-level follow-up work.
3. **Given** the operator runs `stage status` after a failed bootstrap attempt, **When** StageServe has already rolled the project back, **Then** the reported state reflects the rollback outcome rather than leaving phantom running state behind.

---

### User Story 3 - Validate Multi-Project Workflow Against Real Projects (Priority: P2)

An operator using several local sites wants the attach, up, status, and down workflow to be validated against realistic projects, with naming and configuration surfaces that are easy to scan in the repository and in Docker listings.

**Why this priority**: Real-project validation and naming clarity are the parts that make the runtime sustainable in daily use, but they depend on the bootstrap contract and failure model being explicit first.

**Independent Test**: Validate one representative application and one multi-project scenario, explicitly exercising `attach`, DNS routing, shared-gateway readiness, runtime env injection, database provisioning alignment, bootstrap behavior, rollback isolation, teardown, stack-owned config discovery via `.env.stageserve`, and runtime-owned Docker naming under the `stage-` prefix.

**Acceptance Scenarios**:

1. **Given** one representative real application and one multi-project local scenario, **When** the operator follows the documented validation workflow, **Then** the workflow covers attach, up, status, and down along with DNS, gateway, runtime env, database, and bootstrap checks.
2. **Given** stack-wide defaults are required, **When** the operator looks for the StageServe-owned env file, **Then** the documentation and workflow consistently point to `.env.stageserve` rather than an ambiguous or editor-hostile alternative.
3. **Given** the operator inspects runtime-owned Docker resources during validation, **When** they compare multiple attached projects, **Then** the documented naming scheme uses the shorter `stage-` prefix for project-scoped StageServe resources.

### Edge Cases

- A project has no bootstrap command configured but still requires application-owned setup that StageServe does not own.
- A bootstrap command succeeds, but the application remains broken for reasons outside StageServe infrastructure.
- A bootstrap failure occurs after containers become healthy but before the project is registered as a stable attached runtime.
- Multiple projects are attached, and one project fails bootstrap; rollback must not affect other attached runtimes.

## Operational Impact *(mandatory)*

### Ease Of Use & Workflow Impact

- Affected commands, wrappers, or entry points: `stage up`, `stage status`, `stage down`, and the real-project validation workflow that exercises attach/up/status/down behavior.
- Backward compatibility or migration expectation: none. `stage <subcommand>`, location-based `.env.stageserve`, and `stage-` are the supported contract after this feature lands.
- Operator friction removed or introduced: the feature removes undocumented post-start commands, makes bootstrap failure handling explicit, makes stack-owned config easier to identify in editors and repos, and shortens runtime-owned Docker names so attached projects are easier to distinguish.

### Configuration & Precedence

- New or changed configuration inputs: `.env.stageserve` becomes the canonical StageServe config filename for both stack-home defaults and project-local overrides; location defines ownership. `STAGESERVE_POST_UP_COMMAND` remains the bootstrap setting but is valid only from project-root `.env.stageserve` in this feature scope, and is ignored if set via shell environment, stack-home `.env.stageserve`, or project `.env`.
- Precedence order: CLI flags override project-root `.env.stageserve`, which overrides shell environment, which overrides stack-home `.env.stageserve`, which overrides built-in defaults. Project `.env` remains application-owned and is not a generic StageServe config surface, aside from documented runtime fallbacks that already exist.

### State, Isolation & Recovery

- Affected runtime state: per-project containers, networks, volumes, recorded project state, generated gateway state, stack-wide defaults, and runtime-owned Docker identifiers that expose StageServe project ownership.
- Isolation risk and mitigation: bootstrap execution and rollback must remain project-scoped. A failed bootstrap for one project must not mutate the state, naming, routes, or recorded attachment of another project.
- Reliability and recovery path: when bootstrap fails after readiness, StageServe rolls the project runtime back, records the lifecycle as a bootstrap failure, and leaves `stage down` and the documented rerun path as the recovery mechanism.

### Documentation Surfaces

- Docs and interfaces requiring updates: `README.md`, `docs/runtime-contract.md`, any operator guidance that currently references `.stackenv`, any docs that describe runtime-owned Docker names, and the handoff or validation artifacts that explain failure classification and real-project verification.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: StageServe MUST define a single documented bootstrap phase that runs inside the `apache` service container after StageServe-owned readiness succeeds during `stage up`.
- **FR-002**: StageServe MUST source bootstrap behavior only from project-root `.env.stageserve` for this feature scope.
- **FR-003**: StageServe MUST treat bootstrap execution as explicit operator-declared behavior and MUST NOT add implicit framework-specific bootstrap actions when no bootstrap command is configured.
- **FR-004**: StageServe MUST report bootstrap failure as a named lifecycle failure distinct from gateway, DNS, container health, or generic application-route errors.
- **FR-005**: StageServe MUST roll the current project runtime back when the bootstrap phase fails after readiness.
- **FR-006**: StageServe MUST preserve project isolation so a bootstrap failure or rollback in one project does not alter another attached project's runtime state, routes, or recorded attachment.
- **FR-007**: StageServe MUST document the operator workflow that distinguishes StageServe-owned infrastructure success from application-owned follow-up defects discovered during validation.
- **FR-008**: StageServe MUST define a real-project validation workflow that covers at least one representative application and one multi-project scenario across attach, up, status, and down behavior.
- **FR-009**: The validation workflow MUST explicitly cover DNS routing, shared gateway readiness, runtime env injection, database provisioning alignment, bootstrap execution, rollback behavior, and teardown outcomes.
- **FR-010**: StageServe MUST use `.env.stageserve` as the canonical stack-owned defaults file.
- **FR-011**: StageServe MUST keep the ownership boundary explicit by documenting `.env.stageserve` as stack-owned configuration and project `.env` as application-owned configuration.
- **FR-012**: StageServe MUST shorten the documented prefix for project-scoped runtime-owned Docker resources from `stage-` to `stage-`.
- **FR-013**: StageServe MUST document which runtime-owned resource names adopt the `stage-` prefix.
- **FR-014**: StageServe MUST update operator-facing documentation so the canonical naming and lifecycle contract are consistent across runtime guidance and validation guidance.
- **FR-015**: StageServe MUST honor operator-initiated cancellation (`Ctrl-C` / context cancel) during the bootstrap phase by aborting the bootstrap command and triggering project rollback under the same `post-up-hook` step name.
- **FR-016**: StageServe MUST resolve `STAGESERVE_POST_UP_COMMAND` only from project-root `.env.stageserve`. Setting the variable via shell environment, stack-home `.env.stageserve`, or project `.env` MUST NOT cause it to be honored.

### Out Of Scope

The following are explicitly out of scope for spec 004 and MUST NOT be added to this feature without a follow-up spec:

- Additional lifecycle phases beyond the single `post-up` bootstrap (e.g. pre-up, post-attach, post-down).
- Framework-specific bootstrap helpers (Laravel/Symfony/etc. presets baked into the runtime).
- Application migration, seeding, or schema repair beyond what the operator's own bootstrap command performs.
- Bootstrap-execution timeout configuration beyond the existing `STAGESERVE_WAIT_TIMEOUT` (which covers readiness, not bootstrap). Bootstrap inherits the operator's foreground process and is bounded by `Ctrl-C` only.
- Release/distribution pipeline work and CI automation of the manual real-daemon validation workflow.
- TUI / GUI surface changes.
- Backward-compatibility shims for `.stackenv` or `stage-<slug>` runtime names.

### Key Entities *(include if feature involves data)*

- **Bootstrap Contract**: The project-scoped declaration of whether StageServe runs a post-up bootstrap command, when it runs, how failure is classified, and what recovery path applies.
- **Validation Scenario**: A documented real-project workflow used to prove StageServe behavior across single-project and multi-project lifecycle operations.
- **Naming Contract**: The set of stack-owned and runtime-owned names that operators rely on to identify StageServe configuration and Docker resources, including `.env.stageserve` and the `stage-` runtime prefix.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator with a documented bootstrap command can bring a stopped project to either a usable post-bootstrap state or a named bootstrap failure using a single `stage up` run with no undocumented follow-up step.
- **SC-002**: In the documented validation workflow, at least one representative application and one multi-project scenario complete the required attach, up, status, and down checks with all required lifecycle checkpoints recorded.
- **SC-003**: When bootstrap fails for one project, the operator receives a bootstrap-specific failure outcome and recovery guidance, and no unrelated attached project changes state as a result.
- **SC-004**: Operator documentation consistently names `.env.stageserve` as the stack-owned defaults file and `stage-` as the project-scoped runtime-owned Docker prefix.

## Assumptions

- Operators continue to use the current `stage <subcommand>` CLI rather than a new UI or wrapper surface.
- The current StageServe readiness model remains the trigger for bootstrap execution; this feature does not redefine the health model itself.
- Project `.env` files remain application-owned and are not repurposed as the generic StageServe configuration surface.
- `.env.stageserve` and `stage-` are the intended contract for stack-owned config and project-scoped runtime naming in this feature scope.
- Application migration, seeding, or schema bugs discovered during validation remain out of scope unless they reveal a StageServe infrastructure defect.
