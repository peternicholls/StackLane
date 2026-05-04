# Research: Guided TUI And Simple-First StageServe Interaction

## Research Goal

Spec 007 needs to restore the original product intention: a simple first-level `stage` experience that guides normal users through setup, project initialization, run/stop/status, and recovery, while preserving direct commands for power users.

This research looks at comparable guided CLI/TUI products and extracts design rules for StageServe.

## Sources Reviewed

- [DDEV command usage and interactive dashboard](https://docs.ddev.com/en/stable/users/usage/cli/)
- [Vercel CLI deploy command](https://vercel.com/docs/cli/deploy)
- [Fly Launch overview](https://fly.io/docs/reference/fly-launch/)
- [GitHub CLI `gh auth login`](https://cli.github.com/manual/gh_auth_login)
- [Ollama CLI reference](https://docs.ollama.com/cli)
- [Bubble Tea framework](https://github.com/charmbracelet/bubbletea)
- [Huh terminal forms and prompts](https://github.com/charmbracelet/huh)

## Pattern 1: No-Args Command Can Be A Product Surface

### Evidence

DDEV documents that running `ddev` with no arguments launches an interactive terminal dashboard showing projects, status, and keyboard actions. It also documents a fallback: set `DDEV_NO_TUI=true` or `no_tui: true` to get classic help output instead.

Vercel documents `vercel` without a subcommand as a normal deploy entrypoint. It also offers explicit command forms such as `vercel deploy`, and documents `--yes` for skipping setup prompts with defaults.

### Decision For StageServe

Bare `stage` should become a product surface, not just help output.

- In a TTY: open the guided StageServe TUI.
- Outside a TTY: print compact next-step guidance and exit successfully.
- With `stage --help`: show standard Cobra help.
- With `stage <subcommand>`: use the direct CLI path instead of the root guided TUI.
- Provide a disable path: `--notui` or `--cli` for the current invocation and `STAGESERVE_NO_TUI=1` for shell-level fallback.

### What Works

- A no-args command works when it is context-aware and gives one obvious next action.
- Power users stay comfortable when explicit subcommands still exist.
- TUI disable paths avoid breaking CI, SSH, limited terminals, and users who prefer plain output.

### What Does Not Work

- A no-args TUI is harmful if it traps automation or replaces help with an unexpected interactive process in non-TTY contexts.
- A dashboard without an escape hatch frustrates power users.

## Pattern 2: Guided Defaults Should Be Inspectable Before Mutation

### Evidence

Fly Launch scans a project, proposes a configuration, and gives users a chance to tweak settings before proceeding. Its launch outcomes document how it chooses between image, Dockerfile, scanners, defaults, and generated config.

GitHub CLI `gh auth login` starts interactive setup by default, but also supports explicit non-interactive/token paths.

### Decision For StageServe

The guided TUI should show a StageServe plan before it writes config or starts/stops runtime state.

For example:

- "Machine readiness needs DNS setup. Run setup now?"
- "This project has no `.env.stageserve`. Create one with these values?"
- "Project is configured and stopped. Start it now?"
- "Project is running. View status, logs, stop it, or open advanced commands?"
- "Add this project to StageServe?" instead of "attach this project?"
- "Remove this project from StageServe?" instead of "detach this project?"

The user should see which `.env.stageserve` values will be written before writing, and which StageServe action will run before it changes runtime state.

### What Works

- Derived defaults reduce effort, but a preflight summary protects trust.
- A "tweak settings" path keeps the simple path from becoming a rigid path.

### What Does Not Work

- Auto-writing hidden or project files without showing intent makes the guided UI feel magical in the bad sense.
- Asking too many questions before presenting a useful default recreates the friction the TUI is meant to remove.

## Pattern 3: Keep Machine Output And Human Guidance Separate

### Evidence

Vercel documents that deployment `stdout` is always the deployment URL, enabling shell redirection and scripting. Vercel also offers optional `--guidance` to show suggested next steps after deployment.

GitHub CLI supports interactive auth, but also documents token and environment-variable paths for automation.

Spec 005 already introduced stable JSON envelopes for setup, doctor, and init.

### Decision For StageServe

The guided TUI must not break machine-readable paths.

- `stage setup --json`, `stage doctor --json`, and future JSON outputs remain pure JSON.
- Bare `stage` outside a TTY does not start an interactive UI.
- Any command with a stdout contract must keep that contract stable.
- Guidance belongs in TTY UI, stderr, or explicit guidance modes, not in JSON or parseable stdout.

### What Works

- Automation users tolerate guided UX when they can bypass it deterministically.
- JSON and non-interactive modes make tests and future GUI wrappers easier.

### What Does Not Work

- Printing styled guidance into stdout that scripts consume.
- Making TUI the only way to complete setup.

## Pattern 4: TUI Frameworks Need Terminal Fallbacks And Accessibility

### Evidence

DDEV documents terminal compatibility concerns and multiple fallbacks: `NO_COLOR`, simple formatting, terminal environment fixes, and disabling the dashboard entirely.

Bubble Tea is designed for stateful terminal applications and supports simple and complex terminal apps. Its docs also note that TUI apps take over stdin/stdout, so debugging/logging needs care.

Huh supports terminal forms, can integrate with Bubble Tea, and includes a screen-reader accessible mode.

### Decision For StageServe

The StageServe TUI should be progressive enhancement over the command layer.

- Detect TTY before launching.
- Respect `NO_COLOR`.
- Provide `STAGESERVE_NO_TUI=1`, `--notui`, and `--cli`.
- Keep text output as a first-class mode.
- Avoid relying on mouse-only interactions.
- Include accessible form mode where Huh supports it.
- Keep logs out of stdout while the TUI owns the screen.

### What Works

- Keyboard-first TUIs with visible help and predictable keys.
- Text fallback with the same semantic information.
- A clear "quit without changes" path.

### What Does Not Work

- Styling-only "TUI" that does not change the user journey.
- TUI screens that hide the actual operation or make failure output hard to copy.

## Pattern 5: Guided CLI Should Still Teach Power Commands

### Evidence

DDEV's dashboard exposes common actions through keys, but its docs still show explicit commands for start, stop, describe, logs, SSH, exec, and config.

Ollama documents interactive launch flows for integrations, but also gives direct forms such as launching a specific integration or configuring without launching.

### Decision For StageServe

The StageServe TUI should expose the direct command equivalent for every action:

- "Run project" maps to `stage up`.
- "Stop project" maps to `stage down`.
- "Inspect" maps to `stage status`.
- "Diagnose" maps to `stage doctor`.
- "Create config" maps to `stage init`.
- "Advanced" shows the command, config path, and relevant docs.

### What Works

- The TUI helps first-time users and quietly trains power use.
- Advanced users can leave the TUI and run the exact command themselves.

### What Does Not Work

- A TUI that invents hidden flows unavailable from direct commands.
- A TUI that exposes raw Docker operations as primary actions.

## Design Decisions For Spec 007

### Decision 1: Bare `stage` Opens A Guided TUI In Interactive Terminals

Rationale: This restores the original product intention and matches DDEV's proven no-args dashboard pattern.

Alternatives considered:

- Keep bare `stage` as help. Rejected because it preserves the current gap.
- Add only `stage tui`. Rejected as the primary path because users still have to discover the TUI command first.

### Decision 2: Add A Non-UI Next-Action Planner

Rationale: The TUI should not own business rules. A planner can be inspected through terminal verification output and reused by text fallback, docs, and future UI wrappers.

Alternatives considered:

- Put decision logic in Bubble Tea model. Rejected because it makes behavior harder to verify from the terminal and easier to duplicate.
- Dispatch directly to subcommands from root. Rejected because it would not provide a coherent guided state model.

### Decision 3: Keep Direct Commands Stable

Rationale: Power users and automation already rely on `stage setup`, `stage init`, `stage up`, `stage status`, `stage logs`, `stage down`, and JSON modes.

Alternatives considered:

- Route all command behavior through TUI. Rejected because it would break automation and violate the constitution's shell-first constraint.

### Decision 4: Treat Docker Detail As Advanced/Troubleshooting Material

Rationale: StageServe should serve its own API and config contract first. Docker remains the implementation, not the normal operator language.

Alternatives considered:

- Keep Docker names in primary docs for transparency. Rejected because it makes the simple path less approachable; transparency belongs in advanced material.

### Decision 5: Use Bubble Tea For The Guided Shell, Huh For Forms, Existing Projectors For Reports

Rationale: Bubble Tea is appropriate for stateful navigation; Huh is appropriate for bounded forms; existing text/JSON projectors remain appropriate for direct command reports.

Alternatives considered:

- Use only Lip Gloss. Rejected because styled text cannot implement the required guided journey.
- Replace all output modes with Bubble Tea. Rejected because JSON/text modes are necessary for automation and fallback.

### Decision 6: Use User-Goal Labels In Easy Mode

Rationale: Similar guided CLIs work best when prompts describe what the user is trying to accomplish before exposing command or implementation terms. For StageServe, `attach` and `detach` are useful direct command names, but easy-mode users are more likely to understand "add this project to StageServe" and "remove this project from StageServe".

Alternatives considered:

- Reuse command names as TUI labels. Rejected because it forces users to learn lifecycle vocabulary before choosing a safe next action.
- Hide command names entirely. Rejected because "show commands" is important for learning, automation, and power-user control.
