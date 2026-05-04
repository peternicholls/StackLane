# Original Intentions, Research, And Decisions

## Purpose

This document pulls together the product intention, specification research, and decisions found in spec 004, spec 005, the project constitution, and the current developer guidance for spec 007.

The central product intention is not "a Docker CLI with some helper commands". The intended first-level experience is a simple guided StageServe surface that helps a normal user install, set up the machine, set up a project, run it, stop it, inspect it, and recover from problems without needing to understand the Docker implementation.

Power users should still be able to use finer-grained command calls and inspect hidden artifacts, but those controls should sit behind the simple first-level path.

## Primary Product Intention

The intended operator model is:

- `stage` on its own opens the simple guided path.
- The guided path covers first install handoff, machine setup, per-project setup, running, stopping, status, logs, and recovery.
- The user configures StageServe through `.env.stageserve` files only.
- Project `.env` remains application-owned.
- Docker and compose details are implementation internals, not normal project-directory or command-line concepts.
- Runtime artifacts remain hidden in StageServe-owned directories, mainly `.stageserve-state`, but remain inspectable and editable by power users.
- A primary "simple" user is never left at a dead end; every non-ready state should have a specific next action.
- A power user can bypass the guided path with direct commands such as `stage up`, `stage setup --json`, `stage init --force`, `stage status`, and `stage down --all`.

This matches the constitution:

- "Ease Of Use Is A Product Requirement": the primary operator experience must stay simple enough to use from memory.
- "Reliability Must Be Boring And Predictable": repeated runs and config precedence must be deterministic and visible.
- "Robustness Must Hold Under Real Failure": failures must be actionable through status, logs, health checks, or recovery instructions.
- "Remove Pinch Points And User Friction": new steps, prompts, and stateful exceptions must remove recurring pain rather than move it elsewhere.

## Spec 004 Decisions

Spec 004, "Workflow And Lifecycle Hardening", focused on making the existing lifecycle trustworthy. Its main decisions were:

- Keep one lifecycle bootstrap phase only.
- Run the bootstrap phase after StageServe-owned readiness succeeds.
- Run the bootstrap command inside the `apache` service container.
- Source `STAGESERVE_POST_UP_COMMAND` only from the project-root `.env.stageserve`.
- Treat bootstrap failure as a named lifecycle failure, not as a gateway, DNS, Docker, or generic app failure.
- Roll back the current project runtime on bootstrap failure, including operator cancellation.
- Use `.env.stageserve` as the only supported stack-owned defaults filename.
- Keep project `.env` application-owned.
- Use deterministic StageServe-owned runtime names such as `stage-<slug>` for project-scoped resources.
- Keep shared routing StageServe-managed and distinguish it from project-scoped runtime resources.
- Require real-daemon validation of representative single-project and multi-project workflows.

Spec 004 explicitly put "TUI / GUI surface changes" out of scope. That was a planning boundary, not a rejection of the original product intention. It meant the lifecycle had to be hardened before the guided surface could be completed.

## Spec 005 Research And Decisions

Spec 005, "Installer, Onboarding, And Environment Readiness", researched comparable installer and onboarding patterns. The research favored:

- One recommended install path per supported OS.
- Binary integrity verification for direct installs.
- A guided setup sequence with deterministic checkpoints.
- Idempotent setup that can be safely rerun.
- Machine-readable output for automation.
- Explicit privilege boundaries.
- Clear support matrix and compatibility guidance.
- A project initializer to write `.env.stageserve` safely.
- A diagnostics command for day-2 drift.

The concrete decisions in spec 005 were:

- Add `stage setup` for machine readiness and first-run checks.
- Add `stage init` for project-local `.env.stageserve` creation.
- Add `stage doctor` for read-only diagnostics.
- Add text, JSON, and TUI output modes for onboarding commands.
- Use normalized step statuses: `ready`, `needs_action`, and `error`.
- Use stable exit semantics: 0 ready, 1 needs action, 2 error, 3 unsupported OS.
- Keep privileged setup actions explicit and bounded.
- Handoff from the installer to `stage setup --tui` when interactive, or print next steps otherwise.

Spec 005 reintroduced the TUI idea, but in a narrower guise than the original intention: the TUI became an output projection for `setup`, `doctor`, and `init`, not the primary first-level product surface exposed by plain `stage`.

## Reconciled Intent For Spec 007

Spec 007 should treat specs 004 and 005 as useful foundations, not as the final product shape.

The target product model should be:

- Bare `stage` opens a guided TUI command center.
- The guided TUI is the normal first-level path for non-power users.
- The TUI routes users through setup, init, up, status, logs, down, doctor, and recovery.
- The TUI uses the same deep modules and command contracts as direct CLI commands.
- Direct subcommands remain stable for power users and automation.
- All user-editable config is expressed through `.env.stageserve`.
- Docker concepts are hidden from normal project-facing language unless the user opens an advanced/troubleshooting path.
- Hidden runtime artifacts stay available for inspection, but the simple path never requires manual editing of those artifacts.

