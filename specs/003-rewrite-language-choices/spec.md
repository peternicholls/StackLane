# Feature Specification: Rewrite Stacklane Core In A Compiled, Modular Language

**Feature Branch**: `003-rewrite-language-choices`  
**Created**: 2026-04-23  
**Status**: Draft  
**Input**: User description: "Rewrite the Stacklane runtime in a more appropriate language than Bash so the project can grow without compounding fragility, distributing it as a single installable binary and decomposing the current Bash monolith into clear modules with enforced boundaries."

## Clarifications

### Session 2026-04-23

- Q: Output parity scope — what does "matches the previous implementation's contract" (SC-004) mean for parity tests? → A: Semantic parity for human-readable output (status tables, errors, logs); byte-for-byte parity for machine artifacts (per-project state JSON, generated nginx config, project registry). Rationale: locking human output byte-for-byte would contradict FR-013's improved error messages, while machine artifacts must remain stable for downstream tooling (`nginx -t`, operator scripts).
- Q: Disposition of legacy state files after first-contact migration (FR-004) → A: Backup-then-replace. The legacy file is renamed to a `.legacy` sibling (e.g., `<slug>.env.legacy`) and the new-format file is written alongside. Backups are not auto-deleted; operators clean them up manually once they trust the migration.
- Q: Concrete threshold for "no perceptible startup regression" (SC-007) → A: `stacklane --help` cold invocation MUST complete in ≤ 100 ms on the supported macOS reference machine, and MUST NOT exceed 2× the Bash entry-point baseline measured on the same machine. CI fails if either bound is breached.
- Q: Default health-wait timeout for `stacklane --up` (FR-009) → A: 120 seconds, configurable via the `--wait-timeout` flag (CLI) and `STACKLANE_WAIT_TIMEOUT` environment variable, honoring the standard precedence chain.
- Q: Stabilization criteria before the legacy Bash implementation may be removed (Phase 10) → A: At minimum 4 weeks of the binary deployed in the field with zero unresolved parity divergences logged in `docs/migration.md` "Known differences", and successful migration confirmed on every operator machine that reports back. Phase 10 may not execute earlier even if no issues are filed.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install And Run Stacklane Without A Runtime Toolchain (Priority: P1)

A new operator wants to start using Stacklane on a fresh macOS machine. They install the
distributable, run `stacklane --up` in a project directory, and the project comes online with
no separate language runtime to manage and no shell-version compatibility surprises.

**Why this priority**: Distribution simplicity is the single largest payoff of the rewrite. If
the new implementation does not install and run cleanly without an interpreter or toolchain,
the rewrite has not delivered its primary user-facing benefit. Every other improvement depends
on operators being able to adopt the new binary in the first place.

**Independent Test**: On a clean macOS environment with Docker Desktop installed and no
language runtime added, install the Stacklane distributable via the documented method, run
`stacklane --up` from a project directory, and confirm the project starts successfully and
the gateway routes traffic to it.

**Acceptance Scenarios**:

1. **Given** a clean macOS environment with Docker Desktop and no extra language runtime
   installed, **When** the operator follows the documented install path and runs
   `stacklane --up`, **Then** the project starts and is reachable through the shared
   gateway without the operator installing any additional interpreter or package manager
   beyond what Stacklane itself ships with.
2. **Given** an operator already on the previous Bash implementation, **When** they install
   the new Stacklane distributable and run `stacklane --status` inside an existing project,
   **Then** they see their previously recorded project state without losing data and without
   manual file conversion.
3. **Given** the operator runs any Stacklane command, **When** the command is invoked from
   a cold shell, **Then** startup time is fast enough that operators do not perceive a
   regression compared with the existing Bash entry point.

---

### User Story 2 - Get Predictable Behavior And Clear Errors From The Lifecycle Commands (Priority: P1)

An operator running multiple Stacklane projects on the same machine wants `--up`, `--down`,
`--attach`, `--detach`, `--status`, and `--logs` to behave the same way every time, with
errors that name what went wrong and what to do next, instead of failing silently or with
opaque shell stack traces.

**Why this priority**: The current Bash implementation has well-known fragility around global
state, partial failures, and string-parsing assumptions. Operators must trust the lifecycle
commands or they stop using them. Reliability and visible failure are non-negotiable for
infrastructure tooling.

**Independent Test**: Run a representative project lifecycle scenario (`--up`,
`--status`, induce a port collision with another project, then `--down`) and confirm each
command produces the expected outcome, that recoverable errors contain an actionable next
step, and that no unrelated project is affected by a failure in the project under test.

**Acceptance Scenarios**:

1. **Given** two projects exist with overlapping requested ports, **When** the operator runs
   `stacklane --up` on the second project, **Then** the command refuses to proceed, names
   the specific port that conflicts and the project already using it, and leaves the first
   project's runtime untouched.
2. **Given** a `stacklane --up` invocation fails partway through bringing up containers,
   **When** the operator runs `stacklane --status`, **Then** the reported state matches
   reality (no phantom "running" entries for containers that never started), and the
   operator can re-run `--up` or `--down` to recover without manual file editing.
3. **Given** the same project configuration on the same machine, **When** the operator runs
   `stacklane --up` repeatedly, **Then** the same ports, container identities, and gateway
   routes are produced each time within the documented precedence rules.
4. **Given** an operator runs `stacklane --logs`, **When** the underlying container is
   missing or unhealthy, **Then** the command reports the missing or unhealthy container
   by name rather than producing an empty stream or an unrelated shell error.

---

### User Story 3 - Preserve The Existing Operator Command Surface And State (Priority: P1)

An operator who has learned the `stacklane` command surface from the previous spec wants the
rewrite to be invisible at the command-line level: same flags, same precedence rules, same
state file locations, same shared gateway behavior. The change of language must not become a
change of contract.

**Why this priority**: Spec-002 established `stacklane` as the unified command surface and
documented the configuration precedence chain. Breaking that contract during a language
rewrite would silently redo migration work and confuse operators who already adopted the
new vocabulary. Backward compatibility is what makes the rewrite safe to ship.

**Independent Test**: Take an existing project that runs under the Bash implementation,
install the new binary alongside it, and confirm that every documented `stacklane`
invocation accepts the same flags, applies the same precedence chain, and produces the same
operator-visible result.

**Acceptance Scenarios**:

1. **Given** the documented configuration precedence chain (CLI flags → `.20i-local` →
   shell environment → `.env` → defaults), **When** the operator sets a value at any layer
   under the new binary, **Then** the resolved value matches what the previous Bash
   implementation produced for the same input.
2. **Given** existing projects with state recorded by the Bash implementation, **When** the
   new binary first reads that state, **Then** the projects continue to work without the
   operator manually rewriting any state, registry, or configuration file.
3. **Given** the deprecated wrapper command names from spec-002, **When** an operator
   invokes one of them, **Then** the wrapper still delegates to the unified command and
   surfaces the same deprecation guidance, regardless of whether the underlying engine is
   Bash or the new binary.

---

### User Story 4 - Onboard As A Contributor To A Modular, Testable Codebase (Priority: P2)

A contributor wants to add a new feature or fix a bug without having to read 2,000+ lines of
shell to understand which globals are touched. They want to find the responsible module, read
its interface, write a unit test for the change, and ship it with confidence.

**Why this priority**: Contributor-facing friction in the current Bash codebase is a real
constraint on the project's growth, but operators do not feel it directly. It is a
high-impact secondary outcome of the rewrite rather than a prerequisite for shipping.

**Independent Test**: Pick a representative change (for example, adjusting how port
collisions are reported) and confirm that a contributor can locate the responsible module
in the documented structure, read its public interface, write a unit test that covers the
new behavior without spinning up Docker, and run that test in isolation.

**Acceptance Scenarios**:

1. **Given** the new codebase structure, **When** a contributor wants to change how the
   gateway nginx config is generated, **Then** they can find the responsible module from
   the documented project layout in a single step and modify it without editing unrelated
   modules.
2. **Given** a module that depends on Docker, the filesystem, or the network, **When** a
   contributor writes a unit test for that module, **Then** they can substitute the
   external dependency with a test double through the module's documented interface.
3. **Given** the published contributor documentation, **When** a contributor reads it,
   **Then** they can identify which module owns each operator-visible behavior in the
   `stacklane` command surface.

---

### User Story 5 - Treat Docker As A Declarative Partner Instead Of A Polled Subprocess (Priority: P2)

An operator who runs `stacklane --up` wants the command to wait for the project to be
genuinely ready before returning, rather than returning early and leaving the operator to
poll the status themselves. They also want optional development-only services like
phpMyAdmin to start only when explicitly requested.

**Why this priority**: This is a quality-of-life upgrade that becomes possible during the
rewrite without significant extra cost, because the new implementation talks to Docker
through a typed interface rather than parsing CLI output. It is not a precondition for
shipping but is a meaningful operator-visible win once delivered.

**Independent Test**: Run `stacklane --up` on a project and confirm the command returns
only once the gateway and the project's primary services report healthy. Separately,
confirm that phpMyAdmin does not start unless the operator explicitly opts in.

**Acceptance Scenarios**:

1. **Given** a project with a defined health condition, **When** the operator runs
   `stacklane --up`, **Then** the command returns only after the project's primary
   services report healthy, and a timeout produces a clear failure with the names of the
   services that did not become healthy.
2. **Given** phpMyAdmin is configured as an opt-in development service, **When** the
   operator runs the default `stacklane --up`, **Then** phpMyAdmin does not start, and the
   operator sees in `--help` how to opt in when they want it.
3. **Given** an operator runs `stacklane --status`, **When** the command queries running
   containers, **Then** the project's containers are retrieved in a single labeled query
   rather than one subprocess per service.

---

### Edge Cases

- An operator on Linux runs `stacklane --dns-setup` (which only works on macOS today). The
  new binary must report a clear "not supported on this platform" message rather than
  silently failing or producing macOS-specific shell errors.
- An operator runs two `stacklane --up` invocations concurrently in different terminals.
  Port allocation must remain race-safe so two projects cannot claim the same port.
- An operator's machine contains state files written by the previous Bash implementation in
  the legacy format. The new binary must read them and either continue to honor that format
  or migrate them non-destructively on first read.
- An operator runs `stacklane --status` while one project's containers have been removed
  externally (for example, by `docker rm`). The status command must surface the drift
  rather than treating the recorded state as ground truth.
- The gateway nginx config is regenerated while a project is mid-startup. Atomic write
  semantics must guarantee the gateway never reloads against a half-written file.
- An operator upgrades from the old Bash implementation to the new binary while a project
  is currently running. The new binary must reconcile with the live containers without
  requiring a forced `--down` first.

## Operational Impact *(mandatory)*

### Ease Of Use & Workflow Impact

- Affected commands, wrappers, or entry points: the `stacklane` command and every action
  modifier defined in spec-002 (`--up`, `--down`, `--attach`, `--detach`, `--status`,
  `--logs`, `--dns-setup`). The deprecated wrapper command names from spec-002 continue to
  delegate to `stacklane` unchanged.
- Backward compatibility or migration expectation: the operator-visible command surface,
  flags, precedence chain, state file locations, and shared gateway behavior remain
  unchanged. Internal state file format may evolve, but the new binary is responsible for
  reading the legacy format on first contact and migrating non-destructively.
- Operator friction removed: no language runtime or interpreter to install or version-manage;
  fewer opaque shell errors; readiness is reported by the tool instead of polled by the
  operator; opt-in development services. Friction introduced: operators must install the new
  binary and complete the one-time silent state migration the first time they run it.

### Configuration & Precedence

- New or changed configuration inputs: none. The configuration model is unchanged.
- Precedence order: unchanged from the existing Stacklane contract — CLI flags override
  `.20i-local`, which overrides shell environment, which overrides `.env`, which overrides
  built-in defaults. The new binary MUST produce the same resolved values for the same
  inputs as the current Bash implementation.

### State, Isolation & Recovery

- Affected runtime state: per-project state files, the project registry, the shared
  gateway nginx configuration, and the shared Docker network. All locations and semantics
  remain unchanged from an operator's perspective; the on-disk format of state and registry
  files may change but must round-trip through a non-destructive migration.
- Isolation risk and mitigation: the rewrite preserves per-project isolation of containers,
  volumes, networks, and recorded state. Failures in one project's lifecycle MUST NOT
  corrupt another project's state, gateway routes, or recorded ports. Concurrent invocations
  MUST be serialized at the points where they touch shared state (port allocation, registry
  writes, gateway config writes).
- Reliability and recovery path: state writes MUST be atomic so a crashed or interrupted
  command leaves the previous good state intact. `stacklane --status` MUST detect and
  report drift between recorded state and live containers. `stacklane --down` MUST remain
  the documented recovery path when a project is in an inconsistent state, and it must
  succeed even if the project is partially up.

### Documentation Surfaces

- Docs and interfaces requiring updates: `README.md`, `docs/migration.md`,
  `docs/runtime-contract.md`, the contributor documentation describing the module layout
  and how to extend it, distribution and install instructions, and the `--help` text
  emitted by the binary itself. The deprecated wrapper command help text from spec-002
  remains unchanged in content but must continue to delegate correctly.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The Stacklane runtime MUST be reimplemented in a compiled language that
  produces a single installable executable with no external interpreter or runtime
  dependency on the operator's machine.
- **FR-002**: The new implementation MUST preserve the `stacklane` command surface
  established in spec-002, including every documented action modifier and flag, with no
  operator-visible behavior change in the supported common path.
- **FR-003**: The new implementation MUST honor the existing configuration precedence
  chain (CLI flags → `.20i-local` → shell environment → `.env` → defaults) and produce the
  same resolved values for the same inputs as the current Bash implementation.
- **FR-004**: The new implementation MUST read existing per-project state files and the
  existing project registry without requiring the operator to perform a manual format
  conversion. Any internal format change MUST be applied through a non-destructive,
  idempotent migration on first contact. The legacy file MUST be preserved alongside the
  new file with a `.legacy` suffix (for example, `<slug>.env` becomes `<slug>.env.legacy`
  while the new-format file is written next to it). Backups are retained until the
  operator removes them manually.
- **FR-005**: The new implementation MUST decompose the current Bash monolith into
  separately addressable modules with documented public interfaces and explicit boundaries
  between configuration, state, port allocation, Docker orchestration, gateway management,
  platform integration, and observability concerns.
- **FR-006**: Every cross-module dependency that touches Docker, the filesystem, the
  network, or another machine subsystem MUST be expressed through an interface that can be
  replaced with a test double, so unit tests can exercise module logic without spinning up
  Docker.
- **FR-007**: Port allocation MUST be race-safe across concurrent `stacklane --up`
  invocations on the same machine, and collision detection MUST operate over a typed view
  of the registry rather than a positional text format.
- **FR-008**: State, registry, and gateway configuration writes MUST be atomic, so a
  crashed or interrupted command never leaves another project running against a partially
  written file.
- **FR-009**: `stacklane --up` MUST return only after the project's primary services
  report healthy, with a default timeout of 120 seconds and a clear failure message naming
  the services that did not become healthy. The timeout MUST be configurable via the
  `--wait-timeout` CLI flag and the `STACKLANE_WAIT_TIMEOUT` environment variable, honoring
  the standard precedence chain (CLI > env > default).
- **FR-010**: `stacklane --status` MUST report drift between recorded project state and
  live containers, and MUST retrieve a project's containers through a single labeled query
  rather than one subprocess per service.
- **FR-011**: Optional development-only services (such as phpMyAdmin) MUST NOT start by
  default and MUST be reachable through an explicit opt-in flag or profile, with the opt-in
  documented in `--help`.
- **FR-012**: Platform-specific code paths (notably the macOS-only DNS bootstrap, AppleScript
  privilege escalation, and Homebrew-managed dependencies) MUST be isolated so the binary
  builds cleanly for non-macOS platforms and reports a clear "not supported on this
  platform" error when an unsupported platform invokes a macOS-only command.
- **FR-013**: Lifecycle errors MUST surface to the operator with the failing step named,
  the affected project named, and a stated next action — never as a raw shell stack trace
  or a silent non-zero exit.
- **FR-014**: The deprecated wrapper command names from spec-002 MUST continue to
  delegate to `stacklane` and emit the same deprecation guidance, regardless of the
  underlying engine.
- **FR-015**: Contributor documentation MUST publish the module layout, the public
  interface of each module, the testing conventions, and how to add a new module or
  command, so a new contributor can locate the right place to make a change without
  reading the entire codebase.
- **FR-016**: The build, test, and release process for the new implementation MUST be
  documented and reproducible from a clean checkout, including how to produce the
  distributable binary for the supported platforms.

### Key Entities *(include if feature involves data)*

- **Stacklane Binary**: The single distributable executable that replaces the Bash
  monolith. Owns argument parsing, dispatch to lifecycle actions, exit status, and the
  operator-facing error surface.
- **Module Boundary**: A named, documented unit of the new codebase (configuration, state,
  port allocation, Docker client, gateway, DNS, TLS, observability) with a public
  interface and an isolated set of responsibilities. Modules communicate only through
  their published interfaces.
- **Project Configuration**: The resolved view of a single project's settings after the
  precedence chain has been applied. Replaces the loose set of global shell variables
  currently produced by `twentyi_finalize_context`.
- **Project State Record**: The persisted, per-project view of allocated ports, hostname,
  recorded container identities, and routing assignment. Persists across invocations and
  is the source of truth that `--status` compares against live containers.
- **Project Registry**: The aggregated view of every recorded project on the machine,
  used for collision detection and for generating the shared gateway configuration.
- **Gateway Route**: A typed entry describing how the shared gateway forwards a hostname
  to a project's internal service. Replaces the positional text format currently shared
  between the registry and the nginx config generator.
- **Migration Footprint**: The set of legacy state and registry files written by the
  previous Bash implementation that the new binary must read and migrate non-destructively
  on first contact.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new operator on a clean macOS machine with Docker Desktop installed can go
  from "no Stacklane installed" to a successful `stacklane --up` in a project in under
  10 minutes by following the documented install path, with no separate language runtime
  or interpreter installation step.
- **SC-002**: An existing operator upgrading from the Bash implementation can run
  `stacklane --status` against their existing projects on first install of the new binary,
  with zero manual file edits or format conversions required, and the reported state matches
  the live containers.
- **SC-003**: For every documented combination of inputs to the configuration precedence
  chain, the new binary resolves the same values as the previous Bash implementation,
  verified against a published set of test cases.
- **SC-004**: For every documented `stacklane` command and flag, the new binary produces
  output that satisfies the parity contract: semantic equivalence for human-readable text
  (status tables, errors, logs) and byte-for-byte equivalence for machine artifacts
  (per-project state JSON, generated nginx config, project registry). Any operator-visible
  human-output deviation is called out in migration documentation.
- **SC-005**: A failure during `stacklane --up` for one project leaves every other project
  on the machine running, and the failed project's recorded state matches what is actually
  on disk, so the operator can re-run `--up` or `--down` to recover.
- **SC-006**: A contributor unfamiliar with the codebase can locate the module responsible
  for a given operator-visible behavior using the contributor documentation in under 5
  minutes, and can write a unit test for that module without starting Docker.
- **SC-007**: The new binary's `stacklane --help` cold invocation completes in ≤ 100 ms
  on the supported macOS reference machine and never exceeds 2× the Bash entry-point
  baseline measured on the same machine. CI fails if either bound is breached.
- **SC-008**: Concurrent `stacklane --up` invocations on the same machine never produce
  two projects with overlapping ports or two writers to the same state, registry, or
  gateway file, verified through a documented stress scenario.
- **SC-009**: `stacklane --up` returns success only when the project's primary services
  report healthy, and on timeout it names the services that failed to become healthy.

## Assumptions

- The replacement language is chosen during planning, not in the spec. The spec requires
  only that the chosen language produce a single distributable executable with no external
  runtime dependency. Planning has separately recommended Go and that recommendation may be
  ratified during the plan stage; nothing in this spec depends on that choice.
- Operator-visible scope is bounded by spec-002's command surface. New commands, new
  configuration knobs, and new operator-facing features are out of scope for this rewrite
  and belong in follow-up specs.
- The shared gateway's nginx semantics, the per-project Docker Compose topology, and the
  set of supported services remain unchanged. Internal implementation details such as
  whether the gateway adds healthcheck directives, whether phpMyAdmin moves behind a
  Compose profile, and how the binary talks to Docker are implementation choices made
  during planning.
- macOS with Docker Desktop remains the primary supported operator environment. Linux
  portability is preserved at the build level (the binary compiles cleanly and reports a
  clear unsupported-platform error for macOS-only commands), but full Linux support for
  DNS and TLS is out of scope and would be a separate spec.
- A stabilization period during which operators can run both implementations side-by-side
  is required. The legacy Bash implementation is removed only after the binary has been
  deployed in the field for at least 4 weeks with zero unresolved parity divergences
  recorded in `docs/migration.md` "Known differences" and with successful migration
  confirmed on every operator machine that reports back.
- Distribution mechanics (Homebrew tap, GitHub release artifacts, installer scripts) are
  scoped during planning. The spec requires only that the chosen mechanism produces a
  single executable installable on macOS without an interpreter.
- Existing operators are willing to install a new executable as part of this upgrade. The
  rewrite is not required to be hot-swappable into an already-running Bash invocation.
