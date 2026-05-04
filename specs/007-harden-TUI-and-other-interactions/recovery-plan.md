# Recovery Plan: Returning To The Original Guided StageServe Intention

## Purpose

The current codebase has a stronger CLI and onboarding foundation than before specs 004 and 005, but it has not yet fulfilled the original interaction intention:

- simple users start with `stage`
- StageServe guides them through the right next action
- Docker stays behind the StageServe abstraction
- power users can still use direct commands and inspect hidden artifacts

This plan turns that conclusion into executable recovery work for spec 007.

## Research-Backed Direction

The research in [research.md](./research.md) supports four planning choices:

1. No-args command as product surface: DDEV and Vercel show that a bare command can be useful when it is context-aware and has non-interactive fallbacks.
2. Guided defaults with preview: Fly Launch shows the value of scanning context, proposing defaults, and allowing tweaks before mutation.
3. Automation-safe output: Vercel and GitHub CLI show that interactive guidance must not pollute parseable output or headless flows.
4. TUI fallback and accessibility: DDEV, Bubble Tea, and Huh show the need for no-TUI controls, text fallback, keyboard-first operation, and accessible prompts.

## Recovery Target

After spec 007:

- `stage` in an interactive terminal opens a guided TUI.
- `stage` in a non-interactive context prints concise guidance and exits.
- `stage --help` remains standard help.
- `stage <subcommand>` remains the power-user and automation path.
- `stage setup --json` and `stage doctor --json` remain pure JSON.
- The TUI uses the existing config, onboarding, lifecycle, state, status, and logs seams.
- The TUI writes only `.env.stageserve` after preview and confirmation.
- Primary docs explain StageServe actions and config, not Docker resource management.

## What Must Be Recovered

### 1. Bare `stage` Guided Entry

Current state:

- Root command is only a Cobra dispatcher.
- No subcommand means help behavior.

Recovery:

- Add root no-args routing.
- TTY launches guided TUI.
- Non-TTY prints plain next-step guidance.
- Direct subcommands remain unchanged.

### 2. Real TUI, Not Styled Projection

Current state:

- `TUIProjector` renders styled output.
- It explicitly does not run a full Bubble Tea event loop.

Recovery:

- Keep projectors for reports.
- Add a Bubble Tea guided shell for root `stage`.
- Use Huh for bounded forms and confirmations.
- Preserve text/JSON paths.

### 3. Continuous First-Run Path

Current state:

- Installer prints `stage setup --tui`.
- Setup, init, and up remain separate.
- Init prints `stage up` as a next step.

Recovery:

- Guided TUI detects state and offers the next action.
- Missing readiness leads to setup.
- Missing config leads to init with preview.
- Configured stopped project leads to up.
- Failed state leads to doctor/recovery.

### 4. StageServe-First Language

Current state:

- Primary docs and help expose Docker/gateway internals early.

Recovery:

- Primary path uses StageServe concepts.
- Docker/gateway names move to advanced/troubleshooting.
- Remediation starts with StageServe commands.

### 5. Spec Trail Integrity

Current state:

- Spec 004 still shows incomplete validation tasks.
- Spec 005 tasks are checked complete, but some claims and docs do not match current code.

Recovery:

- Spec 007 must include precise tasks and terminal verification evidence.
- Any unrun real-daemon path must be recorded in `quickstart.md`.

## Phased Recovery

### Phase A: Contract And Planner

Deliver:

- `contracts/guided-tui-contract.md`
- `core/guidance` planner package
- terminal-verifiable planner situations for every required situation
- narrow package tests only where terminal verification cannot isolate the logic cleanly

Exit criteria:

- planner returns one primary action and direct command equivalents for all required situations
- planner has no terminal or Bubble Tea dependency

### Phase B: Root Routing

Deliver:

- root no-args TTY path
- root no-args non-TTY fallback
- TUI disable behavior
- help/direct command bypass

Exit criteria:

- `stage` behavior is predictable in TTY, non-TTY, disabled, help, and direct-command modes

### Phase C: TUI Shell

Deliver:

- first Bubble Tea screen
- action list
- keyboard help
- quit/cancel
- result screen
- next-action refresh

Exit criteria:

- a user can start from bare `stage`, understand context, and choose an action

### Phase D: First-Run Actions

Deliver:

- setup action
- init action with `.env.stageserve` preview
- run action after init
- recovery guidance when setup/init cannot finish

Exit criteria:

- a project without `.env.stageserve` can be guided to a confirmed starter config

### Phase E: Day-2 Actions

Deliver:

- status action
- logs action
- down action with confirmation
- doctor action
- drift/error recovery path

Exit criteria:

- a running or down project can be managed from bare `stage`

### Phase F: Power-User Compatibility

Deliver:

- terminal JSON purity checks
- direct command smoke checks
- no-TUI fallback checks
- docs for advanced commands

Exit criteria:

- automation and direct subcommands behave as before

### Phase G: Documentation Cleanup

Deliver:

- README first-run rewrite
- active `docs/installer-onboarding.md`
- runtime contract update
- command help wording cleanup
- `.env.stageserve.example` alignment

Exit criteria:

- primary docs no longer require Docker/gateway implementation vocabulary before advanced sections

## Validation Gate

Do not close spec 007 until all of the following are true:

- terminal planner/root-routing scenarios are recorded
- setup/init/doctor JSON output remains parseable
- direct subcommand help and behavior remain stable
- manual TTY validation is recorded
- non-TTY validation is recorded
- at least one failure/recovery path is recorded
- docs and command help agree with implemented behavior
- focused automated checks pass after terminal verification

## Remaining Risks

- TUI action execution may be awkward if command adapters cannot be reused cleanly.
  - Mitigation: keep deep logic in packages and make command adapters thinner first.
- Long-running lifecycle output may not fit neatly inside Bubble Tea.
  - Mitigation: first version can run action, show spinner/status, then render summarized result with a logs/status escape.
- Hidden implementation detail may still leak through remediation text.
  - Mitigation: add docs/help grep checks and rewrite primary messages.
- Terminal compatibility issues may appear late.
  - Mitigation: provide no-TUI env/flag and text fallback from the start.

## Closeout Definition

The recovery is complete when a normal user can type `stage`, follow the guided path to setup/init/run/status/down/recovery, and never need Docker concepts unless they choose advanced troubleshooting, while a power user can continue to use direct commands without friction.
