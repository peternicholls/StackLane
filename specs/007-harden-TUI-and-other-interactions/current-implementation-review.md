# Current Implementation Review

## Scope

This review examines the current codebase to identify which parts of the spec 004 and spec 005 intentions have been carried out, and which parts remain incomplete against the original simple-guided-product intention.

Files reviewed include:

- `cmd/stage/commands/root.go`
- `cmd/stage/commands/setup.go`
- `cmd/stage/commands/init.go`
- `cmd/stage/commands/doctor.go`
- `cmd/stage/commands/onboarding_mode.go`
- `cmd/stage/commands/project_env.go`
- `core/config/loader.go`
- `core/lifecycle/orchestrator.go`
- `core/onboarding/*`
- `install.sh`
- `README.md`
- `docs/runtime-contract.md`
- `specs/004-workflow-and-lifecycle/*`
- `specs/005-installer-and-onboarding/*`

## Implemented From Spec 004

The lifecycle and configuration foundations are mostly present:

- `.env.stageserve` is the canonical stack and project StageServe config filename.
- `core/config/loader.go` no longer loads legacy stack defaults from `.stackenv` or stack-home `.env`.
- Project `.env` is treated as application-owned, with limited DB fallback behavior.
- `STAGESERVE_POST_UP_COMMAND` is only honored from project-root `.env.stageserve`.
- Project-scoped runtime names default to `stage-<slug>`.
- Shared gateway settings are runtime-owned in config.
- `core/lifecycle/orchestrator.go` runs a post-up hook after health checks.
- Hook failure rolls back the project.
- Lifecycle errors are wrapped with named steps such as `post-up-hook`, `shared-gateway`, `wait-healthy`, and `gateway-reload`.
- Tests exist for core lifecycle and config behavior, including rollback and attach slices.

## Implemented From Spec 005

The installer and onboarding command slices are partly present:

- `install.sh` detects OS and architecture.
- `install.sh` downloads release assets and verifies SHA-256 checksums.
- `install.sh` installs `stage` to a deterministic destination.
- `install.sh` prints `stage setup --tui` as the interactive next step.
- `stage setup` exists.
- `stage doctor` exists.
- `stage init` exists.
- Shared onboarding result types exist in `core/onboarding/types.go`.
- Text, JSON, and TUI projectors exist.
- A common `Projector` interface and `NewProjector` factory exist.
- Machine-readiness checks exist for Docker binary, Docker daemon, state directory, ports, DNS, and mkcert.
- Project env writing exists in `core/onboarding/project_env.go`.
- Project env values are now double-quoted safely.
- Setup, doctor, init, onboarding, and project-env tests exist.

## Root Command Gap

`cmd/stage/commands/root.go` defines a Cobra root command with subcommands, but no root `RunE`.

Current behavior:

- `stage up`, `stage setup`, `stage init`, `stage doctor`, and other subcommands work as separate command surfaces.
- Bare `stage` falls back to Cobra help behavior.

Gap:

- The original product intention requires bare `stage` to expose the simple guided process.
- This is the largest remaining interaction gap.

## TUI Gap

`core/onboarding/projection_tui.go` states:

> It does not run a full Bubble Tea event loop; it produces styled text suitable for non-interactive TUI rendering in the setup/doctor/init flows.

Current behavior:

- TUI mode is styled result rendering.
- There is no interactive menu, wizard, or stateful guided flow.
- Bubble Tea and Huh are in `go.mod`, but the current TUI implementation only uses Lip Gloss directly.

Gap:

- The intended TUI is a guided first-level process, not just a result projector.

## Onboarding Flow Gap

Current flow:

- Installer prints the next setup command.
- `stage setup` checks readiness and reports status.
- `stage init` writes a starter config and prints `stage up`.
- `stage up` starts the project and can create a starter `.env.stageserve` automatically.
- `stage doctor` is separate from setup and lifecycle recovery.

Gap:

- There is no guided handoff from setup to init to up.
- There is no simple decision tree that detects current context and offers the correct next action.
- The user still needs to know or read the command sequence.

## Config And Artifact Boundary

Implemented:

- User-editable StageServe config is centered on `.env.stageserve`.
- Runtime env files are generated under hidden state directories.
- Project env files are protected from overwrite unless forced.
- Hidden state is under `.stageserve-state`.

Incomplete:

- Normal docs and command descriptions still reveal Docker and gateway internals.
- The product abstraction is not yet consistently "StageServe manages runtime; user manages `.env.stageserve`".

## Documentation State

Useful alignment:

- README explains the central StageServe stack authority.
- README documents config precedence and `.env.stageserve`.
- `docs/runtime-contract.md` captures detailed command semantics.

Gaps:

- README and runtime contract still expose implementation details in primary sections.
- `docs/runtime-contract.md` documents `stage setup --recheck`, but the setup command no longer defines a `--recheck` flag.
- Spec 005 task T049 says `docs/installer-onboarding.md` was published, but that file is not present. The closest document is `specs/005-installer-and-onboarding/installer-onboarding.md`.
- Some validation evidence in spec 005 is summarized as expected outcomes rather than full captured command output.

## Code Quality Notes Relevant To Spec 007

The following are not blockers for writing this documentation, but they matter for the recovery plan:

- `setup.go`, `doctor.go`, and `init.go` still manually switch over output modes instead of using `onboarding.NewProjector`.
- `stage init` does not expose a `--tui` flag and cannot force TUI mode the same way `stage setup` can.
- `ValidateDocroot` checks containment but not existence.
- `stage setup` checks readiness but does not currently run the one-time DNS bootstrap itself; it reports readiness and remediation.
- `stage doctor` checks Docker/DNS/mkcert/state/ports, but the implementation does not yet include a distinct gateway-specific readiness check despite docs saying it does.

## Overall Assessment

The implementation has delivered a solid command and readiness foundation:

- Config ownership is much clearer.
- Lifecycle rollback and bootstrap boundaries are stronger.
- Setup, init, and doctor exist.
- Output can be rendered as text, JSON, or styled TUI output.

It has not yet delivered the original first-level experience:

- Bare `stage` is not guided.
- TUI is not interactive guidance.
- Setup/init/up/doctor remain separate steps rather than one guided path.
- Primary docs still teach too much of the Docker/gateway implementation.

Spec 007 should focus on completing that interaction layer without weakening the direct CLI surfaces that now exist.

