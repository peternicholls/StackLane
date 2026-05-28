# StageServe Guided Flow Map

This file pulls the durable interaction contract for spec 007 into the main design system. It is the short reference for how bare `stage` and guided handoffs should route, what each situation means, and what the default action must be.

Use the more detailed source material in `specs/007-harden-TUI-and-other-interactions/flow-diagrams/` when revising a specific screen or transition.

## Core Principle

The planner decides what situation the user is in. The UI decides how to present that situation. Bubble Tea does not decide lifecycle truth, routing truth, or recovery truth on its own.

That separation is the main protection against design drift.

## Top-Level Routing

Bare `stage` should always resolve to one of these outcomes:

1. The machine is not ready.
2. The current folder is not a StageServe project.
3. The project needs its first `.env.stageserve` file.
4. The project is ready to run.
5. The project is already running.
6. The project is configured but stopped.
7. The project no longer matches what StageServe expects.
8. StageServe cannot safely choose a next step.

Those are product situations, not renderer states.

## Situation Table

| Internal situation | User-facing verdict | Default action | Notes |
| --- | --- | --- | --- |
| `machine_not_ready` | `Your computer isn't ready yet.` | Run the next setup step StageServe can safely own. | Setup is a checklist, not a menu of diagnostics. |
| `not_project` | `This folder isn't a StageServe project yet.` | Open project setup preview with visible defaults. | The user should not need README command order first. |
| `project_missing_config` | `This folder doesn't have StageServe settings yet.` | Preview and confirm `.env.stageserve` creation. | Show path, values, and resulting local URL before any write. |
| `project_ready_to_run` | `This project is ready to run.` | Run this project. | The local URL is visible even before the project is started. |
| `project_running` | `This project is running at <URL>.` | View project logs. | The default must stay non-destructive. |
| `project_down` | `This project is stopped.` | Run this project again, or add it back without restarting when it is already running. | Removing the project from StageServe is secondary and confirmed. |
| `drift_detected` | `This project doesn't match what StageServe expects.` | Preview the safe recovery step and confirm it. | The user never sees the word `drift`. |
| `unknown_error` | `StageServe couldn't safely choose a next step.` | Run the first safe recovery step. | Recovery is ordered from read-only to more invasive. |

## Screen Families

### Report surfaces

These are evidence-first screens such as `stage doctor` or readiness checks.

- Show context, verdict, blockers, and exact remediation first.
- Passing checks are reassurance, not the main event.
- Interactive terminals may offer guided help only after the passive report is already useful on its own.

### Guided surfaces

These are action-first screens such as bare `stage`, project setup, run/stop, or recovery.

- Show the human verdict first.
- Show the values StageServe will use before asking the user to commit.
- Highlight the safest likely next action.
- Keep direct commands, advanced detail, and hidden file paths behind a `More…` path or an equivalent secondary surface.

### Utility surfaces

These are confirmations, logs, details, advanced troubleshooting, and More panels.

- Confirmations must name the exact path, URL, or project they affect.
- Logs preserve the log stream and a stable exit hint.
- Advanced and troubleshooting views are opt-in and may include implementation vocabulary.

## The Defaults-Visible Rule

Every screen with a default value or a default action must show that value or action inline before the user commits anything.

That includes:

- site name
- web folder
- domain suffix
- local URL
- target file path
- what pressing `enter` will do
- what a confirmation will change and what it will leave alone

If the value exists, the user can see it. If `enter` does something, the user can read that contract before pressing it.

## Language Boundary

First-level user text stays in StageServe's product language. Docker, compose, gateway, registry, state record, runtime, `attach`, and `detach` are advanced-only vocabulary.

User-goal labels come first:

- `Run this project`
- `Stop this project`
- `Set up this folder as a project`
- `Remove this project from StageServe`
- `Plain text output`
- `Advanced and troubleshooting`

The implementation model can still exist behind those labels. It does not need to lead the conversation.

## Bubble Tea Boundary

The recommended split is:

1. Planner and domain layer: decides the situation and builds the screen plan.
2. Bubble Tea model: owns cursor state, modal state, edit drafts, focus, and in-flight actions.
3. Renderer and styling layer: owns component layout, Lip Gloss styles, glyphs, spacing, and TUI/text parity.
4. Action layer: executes lifecycle, setup, logs, and recovery work, then asks the planner to re-detect.

This keeps design work reusable and prevents functional behavior from being hard-coded into view logic.

## Review Checks

Before approving a new guided screen or report handoff, check:

- The planner situation is clear and maps to one obvious screen.
- The verdict appears before detail.
- The default action is visible and safe.
- The screen shows the values StageServe will actually use.
- First-level copy stays in user language.
- More advanced commands and hidden file paths are secondary, not primary.
- Bubble Tea is styling and guiding the interaction, not deciding the domain truth.