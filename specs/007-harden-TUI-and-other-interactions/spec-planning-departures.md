# Where Planning And Tasks Departed From The Original Intentions

## Summary

The specs did not abandon the original intention all at once. The departure happened in stages:

1. Spec 004 correctly hardened lifecycle behavior, but explicitly deferred TUI and GUI surface work.
2. Spec 005 brought "TUI" back, but treated it as styled output for selected onboarding commands rather than as the primary guided `stage` surface.
3. The task plans completed useful CLI capabilities, but they did not return to the original top-level interaction model.
4. Documentation still teaches several implementation details that should be hidden from ordinary users.

The result is a stronger CLI, but not yet the intended "simple first-level guided StageServe" product.

## Departure 1: `stage` Remained A Subcommand Dispatcher

Original intention:

- Calling `stage` on its own should expose the simple guided process.
- The first screen should help a user decide what to do next: first machine setup, project setup, run, stop, inspect, repair, or advanced controls.

Planning result:

- Spec 004 assumes operators use `stage <subcommand>`.
- Spec 005 adds `stage setup`, `stage init`, and `stage doctor`.
- No spec 005 task defines bare `stage` behavior.

Implementation result:

- `cmd/stage/commands/root.go` defines no root `RunE`.
- With no subcommand, Cobra shows command help rather than starting a guided flow.

Why this matters:

- The normal user still has to know which subcommand to run first.
- The product still exposes a CLI map before it offers a path.

## Departure 2: TUI Became Projection, Not Guidance

Original intention:

- The TUI should guide decisions and actions.
- The user should not need to choose the right command before the guided experience begins.

Planning result:

- Spec 005 tasks add a TUI projection adapter and output-mode resolution.
- The quickstart says to add Bubble Tea models and Huh prompt/form components, but keeps step execution in shared runtime modules.

Implementation result:

- `core/onboarding/projection_tui.go` explicitly says it does not run a full Bubble Tea event loop.
- It renders styled text with Lip Gloss.
- `stage setup --tui`, `stage doctor`, and `stage init` display step results; they do not provide a navigable command center.
- `stage init` does not use the `TUI` flag and passes `forceTUI=false` into output mode resolution.

Why this matters:

- A styled report is useful, but it is not the intended guided first-level interaction.
- The simple user still has to understand command order and recovery paths outside the TUI.

## Departure 3: The First-Run Flow Split Into Commands Without A Primary Path

Original intention:

- StageServe should help the user do everything at the first instance.
- The flow should be continuous: install, machine readiness, project setup, run, status, and recovery.

Planning result:

- Spec 005 documents `install -> setup -> init -> up`.
- Installer handoff prints `stage setup --tui` for TTY users.
- `setup`, `init`, and `doctor` are separate commands, with `up` still outside the onboarding flow.

Implementation result:

- `install.sh` prints the next command instead of launching setup.
- `stage setup` checks readiness but does not transition into project initialization or running.
- `stage init` writes config and prints `stage up` as a next step.
- `stage up` can silently create a starter `.env.stageserve`, but that is not the same as guided project setup.

Why this matters:

- The user can still be bounced between commands.
- The product has improved stepping stones, but not a guided path.

## Departure 4: Docker Implementation Details Still Leak Into User Surfaces

Original intention:

- Usage is via `.env.stageserve` files only.
- Do not reveal Docker information in the project directory or command line beyond what the StageServe API allows.
- Docker details should be available only for advanced inspection or troubleshooting.

Planning result:

- Spec 004 tried to lower gateway detail in normal operator docs.
- The implementation review explicitly warned that docs exposed too much gateway implementation detail.

Implementation result:

- README and `docs/runtime-contract.md` still discuss compose project names, Docker networks, DB volumes, web aliases, gateway config, shared gateway ports, Docker labels, and compose files in normal documentation sections.
- CLI help still describes `stage` as orchestrating containers behind a shared nginx gateway.
- Some remediation text points users toward Docker-level concepts such as Docker daemon, Docker Desktop, `docker ps`, and compose logs.

Why this matters:

- Power users need this information, but primary users should not have to parse it.
- The docs still teach implementation concepts before the StageServe API abstraction is complete.
- The same risk applies to lifecycle command words such as attach and detach when they are used as first-level guided labels; those should be command equivalents, not the plain-language goal.

## Departure 5: Spec 004 Validation Remained Partly Incomplete

Original intention:

- Lifecycle changes should be validated against real startup, status, teardown, and failure paths.

Planning result:

- Spec 004 tasks included real-daemon validation for representative single-project and multi-project scenarios.

Implementation trail:

- `specs/004-workflow-and-lifecycle/tasks.md` still leaves several validation and documentation tasks unchecked, including T029, T031, and T032.
- The implementation review says runtime-facing tests were mostly correct, but follow-up work was still needed.

Why this matters:

- The code may now be improved, but the spec execution trail does not prove full completion of the original lifecycle validation.

## Departure 6: Spec 005 Task Completion Overstated Product Completion

Original intention:

- The onboarding flow should be simple and guided end to end.

Planning result:

- Spec 005 tasks are all checked as complete.

Implementation reality:

- Some code review concerns have been fixed, such as quoted `.env.stageserve` rendering and a common projector interface.
- Some gaps remain:
  - Commands still duplicate output projection dispatch instead of using `onboarding.NewProjector`.
  - `docs/runtime-contract.md` still mentions `--recheck`, but `stage setup` no longer defines that flag.
  - `docs/installer-onboarding.md` is referenced by tasks but is not present in the repository.
  - The TUI is not a full guided event loop.
  - The root command still has no guided behavior.

Why this matters:

- Task completion captured the CLI slices, not the original product-level interaction.

## Corrective Interpretation

Spec 007 should not discard the work from 004 and 005. It should correct the abstraction boundary:

- Treat 004 as the lifecycle reliability foundation.
- Treat 005 as the readiness and projection foundation.
- Build the missing first-level guided experience on top of those foundations.
- Move implementation details into advanced/troubleshooting material.
- Keep direct commands as stable power-user and automation surfaces.
