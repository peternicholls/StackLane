# Feature Specification: Rewrite Stacklane Core In A Compiled, Modular Language

**Feature Branch**: `003-rewrite-language-choices`  
**Created**: 2026-04-23  
**Status**: Draft  
**Input**: User description: "Rewrite the Stacklane runtime in a more appropriate language than Bash so the project can grow without compounding fragility, distributing it as a single installable binary and decomposing the current Bash monolith into clear modules with enforced boundaries."

## Clarifications

### Session 2026-04-23

- Q: Output validation scope — what does SC-004 require for tests? → A: Human-readable output (status tables, errors, logs) is validated semantically; machine artifacts (per-project state JSON and generated nginx config) are validated byte-for-byte where the current Stacklane contract defines stable output.
- Q: Concrete threshold for "no perceptible startup regression" (SC-007) → A: `stacklane --help` cold invocation MUST complete in ≤ 100 ms on the supported macOS reference machine. CI fails if the bound is breached.
- Q: Default health-wait timeout for `stacklane up` (FR-009) → A: 120 seconds, configurable via the `--wait-timeout` flag (CLI) and `STACKLANE_WAIT_TIMEOUT` environment variable, honoring the standard precedence chain.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install And Run Stacklane Without A Runtime Toolchain (Priority: P1)

A new operator wants to start using Stacklane on a fresh macOS machine. They install the
distributable, run `stacklane up` in a project directory, and the project comes online with
no separate language runtime to manage and no shell-version compatibility surprises.

**Why this priority**: Distribution simplicity is the single largest payoff of the rewrite. If
the new implementation does not install and run cleanly without an interpreter or toolchain,
the rewrite has not delivered its primary user-facing benefit. Every other improvement depends
on operators being able to adopt the new binary in the first place.

**Independent Test**: On a clean macOS environment with Docker Desktop installed and no
language runtime added, install the Stacklane distributable via the documented method, run
`stacklane up` from a project directory, and confirm the project starts successfully and
the gateway routes traffic to it.

**Acceptance Scenarios**:

1. **Given** a clean macOS environment with Docker Desktop and no extra language runtime
   installed, **When** the operator follows the documented install path and runs
   `stacklane up`, **Then** the project starts and is reachable through the shared
   gateway without the operator installing any additional interpreter or package manager
   beyond what Stacklane itself ships with.
2. **Given** an operator installs the new Stacklane distributable, **When** they run
   `stacklane status` inside a project that uses the current `.stacklane-local` and
   `.stacklane-state` layout, **Then** the binary reads the canonical state without
   requiring a shell runtime or compatibility wrapper.
3. **Given** the operator runs any Stacklane command, **When** the command is invoked from
   a cold shell, **Then** startup time is fast enough that operators do not perceive a
   regression compared with the existing Bash entry point.

---

### User Story 2 - Get Predictable Behavior And Clear Errors From The Lifecycle Commands (Priority: P1)

An operator running multiple Stacklane projects on the same machine wants `up`, `down`,
`attach`, `detach`, `status`, and `logs` to behave the same way every time, with
errors that name what went wrong and what to do next, instead of failing silently or with
opaque shell stack traces.

**Why this priority**: The current Bash implementation has well-known fragility around global
state, partial failures, and string-parsing assumptions. Operators must trust the lifecycle
commands or they stop using them. Reliability and visible failure are non-negotiable for
infrastructure tooling.

**Independent Test**: Run a representative project lifecycle scenario (`up`,
`status`, induce a port collision with another project, then `down`) and confirm each
command produces the expected outcome, that recoverable errors contain an actionable next
step, and that no unrelated project is affected by a failure in the project under test.

**Acceptance Scenarios**:

1. **Given** two projects exist with overlapping requested ports, **When** the operator runs
   `stacklane up` on the second project, **Then** the command refuses to proceed, names
   the specific port that conflicts and the project already using it, and leaves the first
   project's runtime untouched.
2. **Given** a `stacklane up` invocation fails partway through bringing up containers,
   **When** the operator runs `stacklane status`, **Then** the reported state matches
   reality (no phantom "running" entries for containers that never started), and the
   operator can re-run `up` or `down` to recover without manual file editing.
3. **Given** the same project configuration on the same machine, **When** the operator runs
   `stacklane up` repeatedly, **Then** the same ports, container identities, and gateway
   routes are produced each time within the documented precedence rules.
4. **Given** an operator runs `stacklane logs`, **When** the underlying container is
   missing or unhealthy, **Then** the command reports the missing or unhealthy container
   by name rather than producing an empty stream or an unrelated shell error.

---

### User Story 3 - Use The Current Stacklane Contract Directly (Priority: P1)

An operator wants the rewrite to use the current Stacklane command surface and state model
directly: cobra subcommands, `.stacklane-local`, `.stacklane-state`, and shared gateway
behavior. Old `20i-*` wrappers and `.20i-*` files are not part of the default runtime
contract.

**Why this priority**: The rewrite is the point where Stacklane stops carrying migration-era
interfaces as live behavior. A clean current contract keeps the CLI, docs, and tests from
preserving old workflow assumptions by accident.

**Independent Test**: In a project using `.stacklane-local`, run representative commands
through the Go binary and confirm config resolution, state reads/writes, and gateway output
use only `.stacklane-*` paths and `stacklane <subcommand>` invocations.

**Acceptance Scenarios**:

1. **Given** the documented configuration precedence chain (CLI flags → `.stacklane-local` →
   shell environment → `.env` → defaults), **When** the operator sets a value at any layer
   under the new binary, **Then** the resolved value follows that current Stacklane order.
2. **Given** existing projects with state recorded in `.stacklane-state`, **When** the
   binary reads state, **Then** it loads the JSON records without consulting `.20i-state`.
3. **Given** a repo root install, **When** an operator invokes `stacklane`, **Then** the
   repo shim executes the Go binary directly rather than sourcing the old Bash runtime.

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

An operator who runs `stacklane up` wants the command to wait for the project to be
genuinely ready before returning, rather than returning early and leaving the operator to
poll the status themselves. They also want optional development-only services like
phpMyAdmin to start only when explicitly requested.

**Why this priority**: This is a quality-of-life upgrade that becomes possible during the
rewrite without significant extra cost, because the new implementation talks to Docker
through a typed interface rather than parsing CLI output. It is not a precondition for
shipping but is a meaningful operator-visible win once delivered.

**Independent Test**: Run `stacklane up` on a project and confirm the command returns
only once the gateway and the project's primary services report healthy. Separately,
confirm that phpMyAdmin does not start unless the operator explicitly opts in.

**Acceptance Scenarios**:

1. **Given** a project with a defined health condition, **When** the operator runs
   `stacklane up`, **Then** the command returns only after the project's primary
   services report healthy, and a timeout produces a clear failure with the names of the
   services that did not become healthy.
2. **Given** phpMyAdmin is configured as an opt-in development service, **When** the
   operator runs the default `stacklane up`, **Then** phpMyAdmin does not start, and the
   operator sees in `--help` how to opt in when they want it.
3. **Given** an operator runs `stacklane status`, **When** the command queries running
   containers, **Then** the project's containers are retrieved in a single labeled query
   rather than one subprocess per service.

---

### Edge Cases

- An operator on Linux runs `stacklane dns-setup` (which only works on macOS today). The
  new binary must report a clear "not supported on this platform" message rather than
  silently failing or producing macOS-specific shell errors.
- An operator runs two `stacklane up` invocations concurrently in different terminals.
  Port allocation must remain race-safe so two projects cannot claim the same port.
- An operator's machine contains obsolete `.20i-*` files. The new binary must ignore them
  by default and use only the current `.stacklane-*` locations unless a future explicit
  migration command is specified.
- An operator runs `stacklane status` while one project's containers have been removed
  externally (for example, by `docker rm`). The status command must surface the drift
  rather than treating the recorded state as ground truth.
- The gateway nginx config is regenerated while a project is mid-startup. Atomic write
  semantics must guarantee the gateway never reloads against a half-written file.
- An operator switches from an archived Bash install while a project is currently running.
  The new binary reports drift against the canonical `.stacklane-state` records and does
  not consult archived `.20i-*` state automatically.

## Operational Impact *(mandatory)*

### Ease Of Use & Workflow Impact

- Affected commands, wrappers, or entry points: the canonical `stacklane` command and its
  subcommands (`up`, `down`, `attach`, `detach`, `status`, `logs`, `dns-setup`). Deprecated
  root-level `20i-*` wrappers are removed from the active runtime.
- Backward compatibility or migration expectation: the operator-visible current Stacklane
  command surface, `.stacklane-local` precedence, `.stacklane-state` location, and shared
  gateway behavior are the default contract. Legacy compatibility exists only when a future
  user request explicitly asks for it.
- Operator friction removed: no language runtime or interpreter to install or version-manage;
  fewer opaque shell errors; readiness is reported by the tool instead of polled by the
  operator; opt-in development services. Friction introduced: operators must install the new
  binary and use the current `.stacklane-*` files.

### Configuration & Precedence

- New or changed configuration inputs: `.stacklane-local` is the sole project-local
  configuration file.
- Precedence order: CLI flags override `.stacklane-local`, which overrides shell
  environment, which overrides `.env`, which overrides built-in defaults.

### State, Isolation & Recovery

- Affected runtime state: per-project JSON state files under `.stacklane-state`, the
  computed project registry, the shared gateway nginx configuration, and the shared Docker
  network.
- Isolation risk and mitigation: the rewrite preserves per-project isolation of containers,
  volumes, networks, and recorded state. Failures in one project's lifecycle MUST NOT
  corrupt another project's state, gateway routes, or recorded ports. Concurrent invocations
  MUST be serialized at the points where they touch shared state (port allocation, registry
  writes, gateway config writes).
- Reliability and recovery path: state writes MUST be atomic so a crashed or interrupted
  command leaves the previous good state intact. `stacklane status` MUST detect and
  report drift between recorded state and live containers. `stacklane down` MUST remain
  the documented recovery path when a project is in an inconsistent state, and it must
  succeed even if the project is partially up.

### Documentation Surfaces

- Docs and interfaces requiring updates: `README.md`, `docs/migration.md`,
  `docs/runtime-contract.md`, the contributor documentation describing the module layout
  and how to extend it, distribution and install instructions, and the `--help` text
  emitted by the binary itself.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The Stacklane runtime MUST be reimplemented in a compiled language that
  produces a single installable executable with no external interpreter or runtime
  dependency on the operator's machine.
- **FR-002**: The new implementation MUST provide the current `stacklane` cobra command
  surface (`stacklane up`, `down`, `attach`, `detach`, `status`, `logs`, `dns-setup`) with
  no dependency on old `20i-*` root wrappers.
- **FR-003**: The new implementation MUST honor the current configuration precedence chain
  (CLI flags → `.stacklane-local` → shell environment → `.env` → defaults).
- **FR-004**: The new implementation MUST use `.stacklane-state` as the sole default state
  location and MUST NOT automatically read `.20i-state` or migrate legacy state on startup.
- **FR-005**: The new implementation MUST decompose the current Bash monolith into
  separately addressable modules with documented public interfaces and explicit boundaries
  between configuration, state, port allocation, Docker orchestration, gateway management,
  platform integration, and observability concerns.
- **FR-006**: Every cross-module dependency that touches Docker, the filesystem, the
  network, or another machine subsystem MUST be expressed through an interface that can be
  replaced with a test double, so unit tests can exercise module logic without spinning up
  Docker.
- **FR-007**: Port allocation MUST be race-safe across concurrent `stacklane up`
  invocations on the same machine, and collision detection MUST operate over a typed view
  of the registry rather than a positional text format.
- **FR-008**: State, registry, and gateway configuration writes MUST be atomic, so a
  crashed or interrupted command never leaves another project running against a partially
  written file.
- **FR-009**: `stacklane up` MUST return only after the project's primary services
  report healthy, with a default timeout of 120 seconds and a clear failure message naming
  the services that did not become healthy. The timeout MUST be configurable via the
  `--wait-timeout` CLI flag and the `STACKLANE_WAIT_TIMEOUT` environment variable, honoring
  the standard precedence chain (CLI > env > default).
- **FR-010**: `stacklane status` MUST report drift between recorded project state and
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
- **FR-014**: The repo-root `stacklane` launcher MUST exec the Go binary directly. Legacy
  `20i-*` root wrappers MUST NOT be part of the active runtime by default.
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
  is the source of truth that `status` compares against live containers.
- **Project Registry**: The aggregated view of every recorded project on the machine,
  used for collision detection and for generating the shared gateway configuration.
- **Gateway Route**: A typed entry describing how the shared gateway forwards a hostname
  to a project's internal service. Replaces the positional text format currently shared
  between the registry and the nginx config generator.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new operator on a clean macOS machine with Docker Desktop installed can go
  from "no Stacklane installed" to a successful `stacklane up` in a project in under
  10 minutes by following the documented install path, with no separate language runtime
  or interpreter installation step.
- **SC-002**: An operator using the current `.stacklane-state` layout can run
  `stacklane status` on first install of the new binary, and the reported state matches
  the live containers.
- **SC-003**: For every documented combination of inputs to the current configuration
  precedence chain, the new binary resolves values according to CLI flags,
  `.stacklane-local`, shell environment, `.env`, and defaults.
- **SC-004**: For every documented `stacklane` subcommand and flag, the new binary produces
  current Stacklane behavior; machine artifacts such as generated nginx config and
  per-project state JSON are verified by tests.
- **SC-005**: A failure during `stacklane up` for one project leaves every other project
  on the machine running, and the failed project's recorded state matches what is actually
  on disk, so the operator can re-run `up` or `down` to recover.
- **SC-006**: A contributor unfamiliar with the codebase can locate the module responsible
  for a given operator-visible behavior using the contributor documentation in under 5
  minutes, and can write a unit test for that module without starting Docker.
- **SC-007**: The new binary's `stacklane --help` cold invocation completes in ≤ 100 ms
  on the supported macOS reference machine. CI fails if the bound is breached.
- **SC-008**: Concurrent `stacklane up` invocations on the same machine never produce
  two projects with overlapping ports or two writers to the same state, registry, or
  gateway file, verified through a documented stress scenario.
- **SC-009**: `stacklane up` returns success only when the project's primary services
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
- Distribution mechanics (Homebrew tap, GitHub release artifacts, installer scripts) are
  scoped during planning. The spec requires only that the chosen mechanism produces a
  single executable installable on macOS without an interpreter.
- Existing operators are willing to install a new executable as part of this upgrade. The
  rewrite is not required to be hot-swappable into an already-running Bash invocation.
