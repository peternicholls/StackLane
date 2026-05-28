# Run, Stop, And Inspect Flow

This flow covers day-2 use: the project is configured and the user wants to run it, stop it, see what's happening, or look at logs. It runs when the planner reports `project_ready_to_run`, `project_running`, or `project_down`.

## Source Of Truth

This flow does not run lifecycle commands itself. It calls the existing `stage up`, `stage down`, `stage status`, and `stage logs` paths and shows their results.

## Defaults That Must Be Visible

On every screen in this flow, the user must see:

- The project name.
- The local URL, including scheme, suffix, and port when needed (whether or not it is currently reachable).
- The web folder being served.
- The current status, in plain words.
- What enter will do, before they press it.

## Screen: `project_ready_to_run`

```
StageServe 0.7.0  pete-site

  http://pete-site.develop                               (not running yet)
  Serving ./public_html

▶ Run this project
    Start the project. StageServe will open it in your browser.

  Edit project settings
    Change site name, web folder, or domain suffix

  More…
    Show direct commands, plain text output, advanced and troubleshooting

  ↑/↓ navigate • enter run • → open settings • esc quit
```

Notes:

- The URL is shown even though it isn't reachable yet, so the user knows what they will see.
- The default action is the user's most likely goal: run the project.
- "Edit project settings" routes back to the preview/edit screens in [05-project-setup.md](./05-project-setup.md), but with the existing values pre-filled.

### What Happens When The User Presses Enter

1. StageServe shows a one-line "Starting your project…" status.
2. The existing `stage up` runs in the background.
3. As progress comes in, the same screen updates one row at a time. The user does not see Docker output.
4. When the project is running, the planner re-runs and the screen transitions to `project_running`.
5. If `stage up` fails, the planner reports `drift_detected` or `unknown_error` and the relevant screen takes over.

The transition is one continuous screen with status updates, not a separate "log dump". The user is never shown raw container output unless they explicitly chose to view logs.

## Screen: `project_running`

```
StageServe 0.7.0  pete-site is running

  http://pete-site.develop                               ↗ open in browser
  Serving ./public_html
  Started 4 minutes ago • healthy

▶ View project logs
    Watch what your project is doing right now

  Stop this project
    Free up the local URL and shut down the project

  More…
    Show direct commands, restart, plain text output,
    advanced and troubleshooting

  ↑/↓ navigate • enter open logs • → open in browser • esc quit
```

Notes:

- The default action is non-destructive: `enter` opens logs, not stop. Pressing enter on a running project must never stop it.
- "Open in browser" is a quick keypress (`→`) so the most common day-2 action takes one keystroke.
- "Restart" lives under `More…` because it is rarely the right first instinct; the user usually wants logs first if something is wrong.
- "healthy" is the plain-language status. Internal health-check names do not appear here.

### Sub-Flow: View Logs

Logs are shown in a follow-mode screen. The user always knows how to leave.

```
StageServe 0.7.0  pete-site logs

  10:42:13  GET /  200  12ms
  10:42:14  GET /api/users  200  8ms
  10:42:21  GET /favicon.ico  404  2ms
  ...

  q exit logs • / search • esc exit logs
```

Notes:

- The exit key is shown at the bottom and never moves.
- Leaving logs returns to the `project_running` screen. State is preserved.
- The exact log format is whatever `stage logs` already produces; this flow does not reformat it.

### Sub-Flow: Stop This Project (Confirm Before Stopping)

```
StageServe 0.7.0  Stop pete-site?

  StageServe will stop this project. Your files won't be touched.

  After stopping:
    http://pete-site.develop will no longer respond.
    You can run it again any time.

  ▶ Yes, stop it    No, keep it running

  ←/→ choose • enter confirm • esc cancel
```

Notes:

- Stop always confirms.
- The confirmation tells the user what will happen to their work ("Your files won't be touched").
- `No` is one keypress away, and `esc` also cancels.

After confirming, StageServe runs `stage down`, the planner re-runs, and the user lands on `project_down`.

## Screen: `project_down`

```
StageServe 0.7.0  pete-site is stopped

  http://pete-site.develop                               (stopped)
  Serving ./public_html
  Last stopped: just now

▶ Run this project
    Start the project again

  Remove this project from StageServe
    StageServe will stop tracking this project. Your files won't be touched.

  More…
    Show direct commands, plain text output, advanced and troubleshooting

  ↑/↓ navigate • enter run • esc quit
```

Notes:

- The default is to run the project again, because that is the most common reason to revisit a stopped project.
- If StageServe detects that the project runtime is already running, the first action changes to "Add this project to StageServe" and routes through `stage attach` instead of restarting the project.
- "Remove this project from StageServe" is the easy-mode label for `detach`. It always confirms with a screen explaining that disk files are not deleted.

### Sub-Flow: Remove This Project From StageServe (Confirm Before Removing)

```
StageServe 0.7.0  Remove pete-site from StageServe?

  StageServe will forget about this project.
    .env.stageserve in this folder is left as it is.
    All your project files are left as they are.
    StageServe will no longer route http://pete-site.develop.

  ▶ No, keep it    Yes, remove it

  ←/→ choose • enter confirm • esc cancel
```

Notes:

- This is the only confirmation in the entire TUI where the **non-destructive** option is the default. Removing is rare; accidentally removing is annoying. So `No` is highlighted, and the user has to deliberately move to `Yes`.
- The screen is explicit about what stays untouched on disk.

## State Transitions Out Of This Flow

| From | Trigger | To |
|---|---|---|
| `project_ready_to_run` | `Run this project` | `project_running` (after `stage up` succeeds) |
| `project_ready_to_run` | `Edit project settings` | project setup edit form |
| `project_running` | `View project logs` | logs screen, then back |
| `project_running` | `Stop this project` confirmed | `project_down` |
| `project_running` | open in browser | stays on `project_running` |
| `project_down` | `Run this project` | `project_running` |
| `project_down` | `Add this project to StageServe` | `project_running` |
| `project_down` | `Remove this project` confirmed | `not_project` |
| any | `esc` | exit TUI; the project's actual state is unchanged |

## What This Flow Does Not Do

- It does not show Docker output, container names, image names, network names, or volume names.
- It does not show "find issues" as a peer action. If something goes wrong while running, the planner moves the user to `drift_detected` or `unknown_error` automatically.
- It does not delete project files. Stopping or removing a project never deletes anything the user authored.
- It does not show advanced-only actions in the primary list. They live under `More…`.
- It does not require the user to remember any command name to do any of the standard day-2 actions.
