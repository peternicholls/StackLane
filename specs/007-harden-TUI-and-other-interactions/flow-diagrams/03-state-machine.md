# State Machine: Planner Situations And Transitions

The planner runs whenever the guided TUI needs to know "what should I show now?". It runs once on startup, and again after every action that could change state. The planner produces exactly one situation; that situation chooses one screen.

This is the canonical list of planner situations and what causes them to transition.

## States

| Internal Name | What The User Sees | Owner |
|---|---|---|
| `machine_not_ready` | "Your computer isn't ready yet." | Tool |
| `not_project` | "This folder isn't a StageServe project yet." | Tool |
| `project_missing_config` | "This folder doesn't have StageServe settings yet." | Tool |
| `project_ready_to_run` | "This project is ready to run." | User chooses |
| `project_running` | "This project is running at \<URL\>." | User chooses |
| `project_down` | "This project is stopped." | User chooses |
| `drift_detected` | "This project doesn't match what StageServe expects." | Tool, then user confirms |
| `unknown_error` | "StageServe couldn't safely choose a next step." | Tool, then user confirms |

The internal names stay as they are in the spec. The user never sees `drift_detected` or `unknown_error`; they see the plain-language labels above.

## Initial Detection (One Run From Bare `stage`)

```
Start
  |
  v
Check: is the machine ready?
  |
  +-- no  --> machine_not_ready
  |
  +-- yes --> Check: is the current folder a StageServe project?
                |
                +-- no  --> not_project
                |
                +-- yes --> Check: does .env.stageserve exist?
                              |
                              +-- no  --> project_missing_config
                              |
                              +-- yes --> Check: do recorded state, runtime, DNS, gateway, and config agree?
                                            |
                                            +-- no  --> drift_detected
                                            |
                                            +-- yes --> Check: is the project recorded as running?
                                                          |
                                                          +-- yes --> project_running
                                                          |
                                                          +-- no, recorded down --> project_down
                                                          |
                                                          +-- not recorded --> project_ready_to_run

If any check itself fails (planner can't reach Docker, can't read state, etc.):
  --> unknown_error
```

Each "Check" line maps to an existing function in `core/onboarding`, `core/lifecycle`, or `observability/status`. The planner does not invent any new check.

## Transition Triggers

These are the events that cause a re-detection.

| From | Event | To (typical) |
|---|---|---|
| `machine_not_ready` | A setup step succeeded | re-detect (often `project_missing_config` or `project_ready_to_run`) |
| `machine_not_ready` | User chose to skip a step | stay, but mark the step skipped |
| `not_project` | User chose "Set up this folder" | `project_missing_config` (immediately, since config was just absent) |
| `not_project` | User picked a different folder | re-detect at the new path |
| `project_missing_config` | User chose "Use these settings" or finished editing | `project_ready_to_run` |
| `project_missing_config` | User chose cancel | exit (no file written) |
| `project_ready_to_run` | User chose "Run this project" | `project_running` (after `stage up` succeeds) |
| `project_ready_to_run` | `stage up` failed | `drift_detected` or `unknown_error` depending on classification |
| `project_running` | User chose "Stop this project" and confirmed | `project_down` (after `stage down`) |
| `project_running` | Project becomes unreachable mid-session | `drift_detected` |
| `project_down` | User chose "Run this project" | `project_running` |
| `project_down` | User chose "Add this project to StageServe" | `project_running` |
| `project_down` | User chose "Remove this project from StageServe" | `not_project` (after detach) |
| `drift_detected` | User chose "Use the safe next step" and confirmed | re-detect (typically `project_down` or `project_ready_to_run`) |
| `drift_detected` | User chose "Try to run it again" | re-detect (`project_running` or `unknown_error`) |
| `unknown_error` | User followed a recovery step | re-detect |
| any | User pressed esc at top of TUI | exit |

## Default Action Per State

This is the action that runs if the user simply presses enter at the first screen of that state. Every screen must show this default value to the user before they decide.

| State | Default Action On Enter |
|---|---|
| `machine_not_ready` | Run the next outstanding setup step that StageServe can run itself, or wait for the user on the next external blocker. |
| `not_project` | Open the project setup preview with a name proposed from the folder name. |
| `project_missing_config` | Use the proposed defaults and write `.env.stageserve` after one final confirmation. |
| `project_ready_to_run` | Run this project. |
| `project_running` | View project logs (no destructive default). |
| `project_down` | Run this project, or add it back to StageServe if it is already running. |
| `drift_detected` | Preview the safe recovery, then ask for confirmation before changing StageServe records. |
| `unknown_error` | Run the first recovery step in the recovery list (typically a read-only check). |

The `project_running` default is intentionally non-destructive: pressing enter on a running project must not stop it. Stop is one keypress away but never the default.

## What The Planner Does Not Do

- It does not run lifecycle commands (`stage up`, `stage down`, etc.). The screen layer does that and then asks the planner to re-detect.
- It does not write files.
- It does not change Docker state.
- It does not pick the wording. It returns the internal situation name; the screen layer applies the plain-language label from [00-vocabulary.md](./00-vocabulary.md).
- It does not throttle re-detection. The screen layer is responsible for not asking too often (typically only after a user action or at most once per second on a status screen).

## What Triggers `drift_detected` Specifically

The user-facing name for this state is "this project doesn't match what StageServe expects". It fires when any of the following are inconsistent:

- The recorded project state says "running", but the project URL is not reachable.
- The recorded state says "running", but no underlying runtime is listed.
- The recorded state says "stopped", but the project URL is reachable.
- The DNS for the project's URL is missing.
- The shared routing layer is not handling the project's URL even though the project is recorded as attached.
- `.env.stageserve` was edited in a way that conflicts with the recorded state (for example the URL changed, but the recorded URL still routes).

The user is never told the names of these checks. They are told what is observably true ("StageServe expected this to be reachable, but it isn't") and offered the safe default. See [07-out-of-sync.md](./07-out-of-sync.md) for the screen.
