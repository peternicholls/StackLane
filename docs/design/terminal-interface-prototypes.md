# StageServe Terminal Components And Prototypes

These sketches define the reusable component language for StageServe reports, guided flows, and utility surfaces. They are not byte-perfect snapshots. They are the design contract for how a screen should feel and what it must contain.

Use this file when shaping a new screen, reviewing a prototype, or checking whether the TUI and text fallback still speak the same design language.

For the current production report seed, start with [StageServe Doctor Seed](stage-doctor-seed.md). That document is grounded in the real `stage doctor` command and projector code rather than a forward-looking sketch.

## Component Vocabulary

### Surface header

```text
  ◆  StageServe                         Doctor
  ──────────────────────────────────────
```

Must contain the StageServe identity and current surface. It establishes context, not verdict.

### Verdict line

```text
  ✗  Not ready - 2 of 7 checks need attention.
```

The verdict is the first human sentence about the state.

### Key facts or visible defaults

```text
  Local URL         http://pete-site.develop    (what you'll visit)
  Web folder        ./public_html               (found here)
  Target file       /Users/pete/sites/pete-site/.env.stageserve
```

The user sees the values StageServe will actually use before committing.

### Attention block

```text
  1  Port 443
     Port 443 must be free for the local HTTPS gateway to bind to it.

     Something else on your computer is using port 443.
     To fix:  sudo lsof -nP -iTCP:443 -sTCP:LISTEN
```

Each item carries a label, why-it-matters line, observed problem, and exact remediation.

### Ready summary row

```text
  ✓  State directory    exists
```

Ready rows stay terse and secondary.

### Work checklist

```text
── Setup steps ─────────────────────────

  ✓  Docker Desktop                     ready
  ▶  Local DNS for .develop             needs your approval
     StageServe will add a small file so your browser can open local project URLs.
     Enter: ask for permission and preview the change.
```

The active step explains what happens next. The tool owns the sequence.

### Decision list

```text
── What you can do ─────────────────────

▶ Run this project
    Start the project and open it in your browser.

  Edit project settings
    Change site name, web folder, or domain suffix first.

  More…
    Show direct commands, plain text output, and advanced detail.
```

The default action is first and low-risk.

### Confirmation sheet

```text
Stop pete-site?

  StageServe will stop this project.
  Your files will not be touched.
  http://pete-site.develop will no longer respond.

▶ Yes, stop it    No, keep it running
```

Confirmations explain what changes and what does not.

### More panel

```text
More

  Show direct commands
    stage up
    stage status
    stage logs

  Plain text output
    stage --notui

  Advanced and troubleshooting
    Hidden working folder, exact checks, and implementation detail.
```

Power-user material is grouped here, not mixed into the main action list.

### Logs view

```text
pete-site logs

10:42:13  GET /                  200  12ms
10:42:14  GET /admin             200  21ms

q/esc exit logs
```

Logs preserve the stream and the exit hint.

## Reusable Patterns

### Current Production Seed: StageServe Doctor

**Context:** The current shipped report surface for machine readiness and diagnostics.

```text
  ◆  StageServe Doctor
  ──────────────────────────────────────

  ✗  Not ready — 2 of 7 checks need attention.

── Needs fixing ────────────────────────

  1  Docker daemon
     The Docker daemon must be running before any container can start.

      Docker daemon is not reachable
      To fix:  Start Docker Desktop or run: sudo systemctl start docker

── All clear ───────────────────────────

  ✓  State directory    exists

  ──────────────────────────────────────
  Fix the issues above, then run: stage doctor
```

**Why it matters:** This is the real production seed for report surfaces. It establishes the current section grammar, verdict placement, semantic colour model, and the current remediation slot, even where the wording is still more operational than polished.

For the full code-anchored write-up, see [StageServe Doctor Seed](stage-doctor-seed.md).

### Detailed report with assistance handoff

**Context:** `stage doctor` in an interactive terminal finds blockers.

```text
  ◆  StageServe                         Doctor
  ──────────────────────────────────────

  ✗  Not ready - 2 of 7 checks need attention.

── Needs fixing ────────────────────────

  1  Port 443
     Port 443 must be free for the local HTTPS gateway to bind to it.

     Something else on your computer is using port 443.
     To fix:  sudo lsof -nP -iTCP:443 -sTCP:LISTEN

  2  Local DNS resolver
     Routes local project addresses to this computer.

     Your computer cannot yet open local project URLs.
     To fix:  stage setup

── All clear ───────────────────────────

  ✓  Docker Desktop    running
  ✓  State directory   exists
  ✓  mkcert local CA   installed

── Assistance ──────────────────────────

  StageServe can help with the issues above.

▶ Help me fix these
    Walk through each issue one at a time.

  Leave it here
    Exit without changing anything.
```

**Why it works:** the passive report is already useful, then the interactive handoff stays explicit and optional.

### Machine setup checklist

**Context:** the machine is not ready yet.

```text
  ◆  StageServe                         Setup
  ──────────────────────────────────────

  Your computer isn't ready yet.

  StageServe is checking the computer before it looks at this project.

── Setup steps ─────────────────────────

  ✓  Docker Desktop                     ready
  ✓  StageServe working folder          ready
  ▶  Local DNS for .develop             needs your approval
     StageServe will add a small file so your browser can open local project URLs.
     Enter: preview the change and confirm it.

  •  Local HTTPS certificates           optional for this URL
  •  Network ports 80 and 443           pending

  ? details • q quit
```

**Why it works:** setup is presented as tool-owned work, not a diagnostic menu.

### Project setup preview

**Context:** this folder needs its first `.env.stageserve` file.

```text
  ◆  StageServe                         Project setup
  ──────────────────────────────────────

  This folder doesn't have StageServe settings yet.

  StageServe will create one file only: .env.stageserve.

── Key facts ───────────────────────────

  Site name         pete-site                          (from folder name)
  Web folder        ./public_html                      (found here)
  Domain suffix     .develop                           (machine setting)
  Local URL         http://pete-site.develop           (what you'll visit)
  Target file       /Users/pete/sites/pete-site/.env.stageserve

── What you can do ─────────────────────

▶ Use these settings
    Write .env.stageserve and continue.

  Edit before writing
    Change site name, web folder, or suffix first.

  More…
    Show direct commands, plain text output, and advanced detail.
```

**Why it works:** the values are visible before the write, and the edit path is one step away.

### Running project screen

**Context:** a configured project is already running.

```text
  ◆  StageServe                         Project
  ──────────────────────────────────────

  pete-site is running.

── Key facts ───────────────────────────

  Local URL         http://pete-site.develop           (open in browser)
  Web folder        ./public_html                      (serving now)
  Status            healthy                            (started 4 minutes ago)

── What you can do ─────────────────────

▶ View project logs
    Watch what your project is doing right now.

  Stop this project
    Free up the local URL and shut down the project.

  More…
    Show direct commands, plain text output, and advanced detail.

  → open in browser • q quit
```

**Why it works:** the default action is informative, not destructive.

### Recovery flow

**Context:** StageServe cannot safely choose the next step.

```text
  ◆  StageServe                         Recovery
  ──────────────────────────────────────

  StageServe couldn't safely choose a next step.

  Something went wrong while StageServe was checking this project.
  StageServe doesn't want to guess. Here is what it can try, in order.

── Recovery steps ──────────────────────

  ▶  Step 1: look at this project's current state   read-only
     Nothing on your computer will be changed.

  •  Step 2: look at the running log                read-only
  •  Step 3: stop and forget the running record     confirmed change
  •  Step 4: run this project from scratch          uses current settings

── What you can do ─────────────────────

▶ Run step 1
    Start with the least invasive step.

  Show what went wrong in detail
    Read a longer plain-language explanation.

  More…
    Show direct commands, plain text output, and advanced detail.
```

**Why it works:** the recovery path is ordered, visible, and safe-by-default.

### Plain-text twin

**Context:** `--notui`, `--cli`, or no TTY.

```text
StageServe Project setup

This folder doesn't have StageServe settings yet.
StageServe will create one file only: .env.stageserve.

Key facts
  Site name: pete-site (from folder name)
  Web folder: ./public_html (found here)
  Domain suffix: .develop (machine setting)
  Local URL: http://pete-site.develop (what you'll visit)
  Target file: /Users/pete/sites/pete-site/.env.stageserve

What you can do
> Use these settings
  Write .env.stageserve and continue.

- Edit before writing
  Change site name, web folder, or suffix first.

- More…
  Show direct commands, plain text output, and advanced detail.
```

**Why it works:** text mode keeps the same truth and ordering, even without colour or richer layout.
```

**Why it works:** The output explains the current step without streaming noisy implementation logs. It names why waiting is normal.

**Rules demonstrated:** progress as product state, concise expectation setting.

## Destructive Action Warning

**Context:** A command may delete, overwrite, or unregister local state.

**User state:** The user needs the consequence and a clear confirmation path.

**Interface sketch:**

```text
  *  StageServe
  --------------------------------------

  !  This will unregister api.test from StageServe.

-- Impact ------------------------------

  The project will stop routing through the local gateway.
  Certificates and source files will not be deleted.

  Next:  stage project remove api.test --confirm
```

**Why it works:** It states the concrete impact and what is not affected. Confirmation is explicit in the command.

**Rules demonstrated:** risk clarity, bounded consequence, explicit confirmation.

## Dry-Run Preview

**Context:** A command can show planned changes before applying them.

**User state:** The user needs to compare intent with effect.

**Interface sketch:**

```text
  *  StageServe Plan
  --------------------------------------

  !  Previewing changes. Nothing has been changed.

-- Will update -------------------------

  Project domain     api.test
  HTTPS port         8443
  Gateway route      /Users/example/api

  Next:  stage apply
```

**Why it works:** The dry-run state is explicit. Planned changes use labels that map to user concepts, not internal keys.

**Rules demonstrated:** state clarity, user-facing labels, safe handoff.

## Empty State

**Context:** A list command finds no registered projects.

**User state:** The user needs to know whether this is normal and how to create the first item.

**Interface sketch:**

```text
  *  StageServe Projects
  --------------------------------------

  !  No projects are registered yet.

  Next:  stage init
```

**Why it works:** It does not print an empty table. It gives the next action.

**Rules demonstrated:** empty state as UX, no empty scaffolding.

## Post-Success Handoff

**Context:** A command succeeds and there is one natural next workflow step.

**User state:** The user is ready to continue.

**Interface sketch:**

```text
  OK api.test is registered.

  Next:  stage up
```

**Why it works:** It avoids ceremony. The state is clear and the next command is exact.

**Rules demonstrated:** quiet success, direct continuation.

## Plain-Text Parity

**Context:** `--no-tui`, redirected output, or non-interactive mode.

**User state:** The user or automation needs the same information without styling.

**Interface sketch:**

```text
StageServe Doctor

Not ready - 1 of 7 checks needs attention.

Needs fixing

1. Docker daemon
   The Docker daemon must be running before any container can start.

   Docker is installed, but the daemon is not running.
   To fix: open -a Docker

All clear

- State directory: exists
- mkcert: installed

Fix the issues above, then run: stage setup
```

**Why it works:** It preserves information and order without relying on colour, glyphs, or weight.

**Rules demonstrated:** text/TUI parity, semantic content independence.

## Anti-Patterns

Do not use internal state dumps:

```text
Result: Needs action (exit 1)
Summary: 5 ready, 2 needs_action, 0 errors
[docker.daemon] StatusNeedsAction
Run: open -a Docker (starts docker)
```

Do not use decorative terminal frames:

```text
########################################
#          STAGESERVE CHECKS!!!        #
########################################
```

Do not use vague remediation:

```text
Something went wrong with DNS.
Try checking your settings and run again.
```
