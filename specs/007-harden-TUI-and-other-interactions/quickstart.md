# Quickstart: Validating Spec 007

## Goal

Validate that StageServe now provides a simple guided first-level path while preserving direct command and automation behavior.

## Prerequisites

- Go 1.26 toolchain.
- A terminal capable of TTY interaction.
- A test project directory without `.env.stageserve`.
- A configured test project directory with `.env.stageserve`.
- Docker Desktop available for full lifecycle validation, or explicit notes when daemon validation is not run.

## Verification Approach

This spec run uses terminal verification as the primary development loop. The goal is to catch interaction problems through real `stage` usage before relying on narrower package checks.

For each implementation slice:

1. Build or run the current checkout's `stage` binary.
2. Run the relevant terminal scenario.
3. Capture command, exit code, and key output.
4. Fix the behavior.
5. Re-run the same terminal scenario.
6. Only then run focused package checks as supporting evidence.

Use the repository-local command during verification so results are tied to the code under review:

```bash
make build
./stage --version
```

If using an installed `stage` on `PATH`, record which binary is being exercised:

```bash
command -v stage
stage --version
```

## Terminal Verification - Primary

Use these scenarios during implementation and closeout.

### 1. Bare `stage` opens guided path

```bash
stage
```

Expected:

- TUI opens in an interactive terminal.
- It shows current context.
- It offers one primary next action.
- It shows help/quit.
- It does not show Docker implementation names on the first screen.
- It uses user-goal labels such as "run this project", "create project settings", "stop this project", or "find issues" rather than command jargon.

Evidence to record:

- command
- exit code after quit
- screenshot or concise output description
- primary action shown

### 2. Non-interactive no-args does not hang

```bash
stage > /tmp/stage-guidance.txt
printf 'exit=%s\n' "$?"
sed -n '1,80p' /tmp/stage-guidance.txt
```

Expected:

- Does not hang.
- Prints compact guidance.
- Exits 0 unless context collection fails fatally.

### 3. TUI disable path

```bash
stage --notui
stage --cli
STAGESERVE_NO_TUI=1 stage
```

Expected:

- Text fallback is shown.
- No interactive UI is opened.
- Text fallback follows the same plain-language rules as the TUI.

### 4. Missing project config

From a project without `.env.stageserve`:

```bash
stage
```

Expected:

- TUI proposes creating `.env.stageserve`.
- It previews path and values before writing.
- Cancel before confirmation leaves no file.
- Confirm writes `.env.stageserve`.
- Result screen offers `stage up` equivalent.
- First-level label is "create project settings"; `stage init` is shown as the command equivalent.

### 5. Configured stopped project

```bash
stage
```

Expected:

- TUI identifies project as configured and stopped.
- Primary action is to run the project.
- Direct command equivalent is visible: `stage up`.
- First-level label is "run this project".

### 6. Running project

```bash
stage
```

Expected:

- TUI offers status/logs/down/doctor.
- Stop action confirms before running.
- Stop preserves data and uses `stage down` semantics.
- Action labels use plain language: "check project status", "view logs", "stop this project", "find issues", and "show commands".

### 7. Logs terminal behavior

```bash
stage
```

Choose logs from the guided UI.

Expected:

- logs view has a visible exit path
- exiting logs leaves the terminal usable
- output does not smear over the shell prompt

### 8. Ctrl-C cancellation

Run the guided UI and press Ctrl-C:

- before confirming config write
- during a long-running action when feasible

Expected:

- cancellation before confirmation leaves no `.env.stageserve` or runtime state change
- cancellation during an action surfaces the safest next action
- terminal remains usable

### 9. Failure path

Simulate a missing Docker daemon, DNS drift, invalid `.env.stageserve`, or bootstrap failure.

Expected:

- TUI shows the problem.
- It provides a StageServe recovery path first.
- Advanced implementation details are available only behind an advanced/troubleshooting action.
- Terms such as attach, detach, daemon, gateway, compose, container, registry, runtime, and state do not appear in first-level recovery copy unless they are the only actionable recovery clue.

### 10. JSON remains pure

```bash
stage setup --json > /tmp/stage-setup.json
jq . /tmp/stage-setup.json >/dev/null

stage doctor --json > /tmp/stage-doctor.json
jq . /tmp/stage-doctor.json >/dev/null
```

Expected:

- stdout is valid JSON.
- no styled text or next-step prose is mixed into JSON.

### 11. Direct commands follow the spec 007 contract

```bash
stage up --help
stage attach --help
stage status --help
stage logs --help
stage down --help
stage detach --help
```

Expected:

- Direct commands bypass the root guided TUI.
- Help and flag output match the final spec 007 contract.
- Easy-mode screens do not require users to understand `attach` or `detach`; those words are acceptable in direct command help and show-commands output.

When a real Docker daemon and disposable configured project are available, also verify:

```bash
stage up
stage status
stage logs
stage down
stage attach
stage detach
```

Expected:

- direct up/status/logs/down behavior matches the final spec 007 command contract
- direct attach/detach behavior matches the final spec 007 command contract
- any unrun daemon-dependent check is recorded as a real-daemon gap

Easy-mode label expectation:

- `stage attach` is presented as "add this project to StageServe" outside direct command help.
- `stage detach` is presented as "remove this project from StageServe" outside direct command help.

### 12. `stage init` default TUI

```bash
stage init
```

Expected:

- opens the guided project-config form in an interactive terminal
- previews `.env.stageserve` before writing
- preserves `stage init --notui`, `stage init --cli`, and `stage init --json` behavior
- final spec 007 help does not advertise `stage init --tui`

### 13. First-screen render time

```bash
time stage
```

Expected:

- first useful screen renders within 500 ms excluding explicitly selected long-running checks
- if the threshold is missed, record the measured time and cause

### 14. Text fallback parity

Compare:

```bash
stage --notui
stage --cli
STAGESERVE_NO_TUI=1 stage
stage > /tmp/stage-guidance.txt
```

Expected:

- text fallback includes the same situation, primary action, and direct command equivalent shown in the TUI

### 15. Installer handoff

Run the installer in test mode or another safe local path after the guided entrypoint exists.

Expected:

- interactive install points to bare `stage`
- non-interactive install prints explicit commands such as `stage setup`, `stage init`, and `stage up`

### 16. Plain-language review

Capture the first screen, text fallback, first-run docs, and installer handoff copy.

Expected:

- primary actions describe user goals before command names
- command equivalents remain discoverable through "show commands" or direct help
- implementation terms appear only in advanced/troubleshooting copy unless needed for a concrete recovery step

### 17. Multi-project scope

When multiple projects are available through StageServe, run `stage` from one project root.

Expected:

- the guided planner remains scoped to the current directory
- the UI does not imply a cross-project switcher in the first implementation
- any future multi-project guided switching is treated as out of scope for spec 007

## Supporting Package Checks

After terminal scenarios are green for the current slice, run focused packages:

```bash
go test ./core/guidance ./core/onboarding ./cmd/stage/commands
go test ./core/config ./core/lifecycle ./observability/status ./infra/gateway
```

Expected:

- planner states pass
- root no-args routing tests pass
- setup/init/doctor JSON purity tests pass
- direct lifecycle tests remain green

## Documentation Validation

Check:

- README first-run path starts with bare `stage`.
- Docker/gateway names do not appear in the primary first-run path.
- attach/detach and runtime/state vocabulary do not appear as easy-mode labels.
- Advanced/troubleshooting sections still contain enough implementation detail for power users.
- `.env.stageserve` is the only normal user-editable StageServe config file.
