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
- It shows the current context in a status header.
- It shows a decision bar only when the user has a real choice.
- It shows setup, diagnostics, and recovery as tool-owned work panels rather than peer menu choices.
- It shows help/quit.
- It does not show Docker implementation names on the first screen.
- It uses user-goal labels such as "run this project", "create project settings", "view project logs", or "stop this project" rather than command jargon.
- It shows the active suffix, scheme, port when needed, and local URL before any run or write.

Evidence to record:

- command
- exit code after quit
- screenshot or concise output description
- status header, highlighted default, visible defaults, and footer affordances shown

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
- Highlighted default is to run the project.
- Direct command equivalent is visible: `stage up`.
- First-level label is "run this project".

### 6. Running project

```bash
stage
```

Expected:

- TUI shows URL/status and defaults to a non-destructive action such as viewing logs.
- Stop action confirms before running.
- Stop preserves data and uses `stage down` semantics.
- Action labels use plain language: "view project logs" and "stop this project".
- Direct commands and troubleshooting are discoverable through the footer rather than shown as peer actions.

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

- text fallback includes the same situation, highlighted default, visible defaults, and direct command equivalent shown in the TUI

### 15. Installer handoff

Run the installer in test mode or another safe local path after the guided entrypoint exists.

Expected:

- interactive install points to bare `stage`
- non-interactive install prints explicit commands such as `stage setup`, `stage init`, and `stage up`

### 16. Plain-language review

Capture the first screen, text fallback, first-run docs, and installer handoff copy.

Expected:

- decision-bar actions describe user goals before command names
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

## Recorded Evidence

### Design Audit Notes

- 2026-06-05: Guided shell audit against `docs/design/terminal-visual-style-guide.md`, `.github/instructions/terminal-design-contract.instructions.md`, and `.github/instructions/terminal-markup-spec.instructions.md` found the current shell structurally aligned on surface header, verdict-first layout, key facts, work checklist, decision list, confirmation, and contextual footer hints.
- 2026-06-05: Intended visual deltas for the dedicated polish phase are now explicit: move secondary command/detail material toward a clearer `More…` or advanced surface instead of relying only on footer shortcuts, refine work-checklist emphasis so one active setup/recovery step reads more clearly than the surrounding rows, tighten spacing and scan rhythm around utility surfaces and confirmations, and keep long-value handling usable at narrow widths without relying only on terminal wrapping.
- 2026-06-05: the guided shell polish pass extracted reusable row render helpers for key facts, decision rows, and work rows; added a focused accent style for the selected decision, active setup/recovery row, and active editor field; and kept `go test ./core/guidance` green after the visual refactor.
- 2026-06-05: the current guidance test coverage still exercises the 48-column narrow layout plus the details, confirmation, editor, utility, and loading transitions, so the design pass verification remains anchored in executable checks while the deferred visual work stays the same: a clearer `More…` surface and stronger long-value handling than terminal wrapping alone.

- 2026-05-28: `go run ./cmd/stage --help` kept direct help output and described bare `stage` as the guided next-step entrypoint.
- 2026-05-28: `go run ./cmd/stage init --help`, `go run ./cmd/stage setup --help`, and `go run ./cmd/stage doctor --help` used plain-language goal copy aligned with the guided surface.
- 2026-05-28: `STAGESERVE_INSTALL_DIR=$(mktemp -d) STAGESERVE_TEST_ASSET_PATH=/bin/echo ./install.sh` printed the interactive installer handoff `stage`.
- 2026-05-28: `NONINTERACTIVE=1 STAGESERVE_INSTALL_DIR=$(mktemp -d) STAGESERVE_TEST_ASSET_PATH=/bin/echo ./install.sh` printed explicit next steps: `stage setup`, `stage init`, `stage up`, and `stage doctor`.
- 2026-05-28: `go run ./cmd/stage --stack-home /tmp/stageserve-guided-stack --project-dir /tmp/stageserve-guided-missing` in a TTY opened the guided missing-config screen with `Create project settings`, a local URL preview, and no pre-write mutation.
- 2026-06-05: `./stage-bin --stack-home <temp-stack-home> --project-dir <temp-project>` in a disposable TTY case with no existing state directory opened `Your computer isn't ready yet.` with a tool-owned `Setup steps` checklist and the first blocker (`State directory`) highlighted, rather than skipping straight to project settings.- 2026-06-05: NFR-001 exceeded with the initial full-readiness approach (1.327s due to `docker info` and DNS probes). Replaced with a cheap state-directory heuristic in `collectRootGuidedContext` so the first render on a set-up machine takes 0.431s (within 500ms target). The full diagnostics run only when the user explicitly requests them through the setup action or `d` doctor shortcut.
- 2026-06-05: architectural verification confirms `core/guidance` (planner, context, shell, types) has zero direct imports of `core/lifecycle`, `infra/compose`, `infra/gateway`, or Docker client; config precedence remains fully owned by `core/config.NewLoader`; lifecycle and state mutations happen only through the command-layer action handler seam in `cmd/stage/commands/tui.go`.
- 2026-06-05: Ctrl-C cancellation analysis: `shell.go` returns `tea.Quit` in all five interactive states (loading, editing, utility, confirming, main). In loading state, key events are absorbed but `Ctrl-C` specifically quits — the async goroutine completes independently but the program exits cleanly. No guided session writes `.env.stageserve` or state files unless the user confirms a mutating action. Cancel before confirmation leaves no files.- 2026-05-28: quitting disposable guided sessions before any confirmed write left `/tmp/stageserve-guided-missing/.env.stageserve` and `/tmp/stageserve-guided-cancel/.env.stageserve` absent.
- 2026-05-28: `go run ./cmd/stage --stack-home /tmp/stageserve-guided-stack --project-dir /tmp/stageserve-guided-missing --notui` produced the same situation and default action in the plain-text fallback.
- 2026-05-28: with `STACK_HOME` set to the repo checkout and `STAGESERVE_STATE_DIR=/tmp/stageserve-live-state`, a disposable live project at `/tmp/stageserve-live-project` reached the guided running screen (`This project is running at http://stageserve-live.test.`) and the retained stopped screen (`This project is stopped.`) after `stage down`.
- 2026-06-05: `./stage-bin --project-dir <temp-project-with-.env.stageserve>` opened `This project is ready to run.` with `▶ Run this project` highlighted first, confirming the configured ready-to-run screen on the current binary.
- 2026-05-28: a scripted live logs run on the disposable project entered the logs utility, matched the visible exit hint `q exit logs`, returned to the main guided screen with `Esc`, and exited cleanly with `expect_exit=0`.
- 2026-05-28: a scripted live stop confirmation on the disposable project matched `StageServe will stop this project.` and `Your files will not be touched.`, then cancelled back to `No changes made.` with `expect_exit=0`.
- 2026-05-28: a scripted interactive `stage init` run on the same project matched the guided overwrite flow (`This folder already has StageServe settings.` and `Update project settings`) and its confirmation copy, then cancelled back to `No changes made.` with `expect_exit=0`.
- 2026-05-28: `go test ./core/guidance ./cmd/stage/commands` passed after adding focused coverage for the drift recovery stop confirmation path and the guided overwrite path in `stage init`.
- 2026-05-28: the Phase 4 guided-shell polish pass added named surface headers (`Project`, `Project setup`, `Setup`, `Recovery`), moved setup and recovery checklists ahead of the choice list, and stacked key facts on 48-column renders while keeping the `NO_COLOR` view ANSI-free.
- 2026-05-28: `go test ./core/guidance ./cmd/stage/commands` passed after adding deterministic render coverage for the project screen, the recovery screen, and the 48-column narrow layout; the deliberate deferral for this slice is that very long path/value rows still rely on terminal wrapping rather than dedicated truncation.
- 2026-06-05: `make build && ./stage setup --json > /tmp/stageserve-setup.json && jq . /tmp/stageserve-setup.json >/dev/null` kept stdout parseable JSON with `setup_exit=0`.
- 2026-06-05: `./stage doctor --json > /tmp/stageserve-doctor.json && jq . /tmp/stageserve-doctor.json >/dev/null` kept stdout parseable JSON with `doctor_exit=0`.
- 2026-06-05: `./stage init --project-dir $(mktemp -d) --json > /tmp/stageserve-init.json && jq . /tmp/stageserve-init.json >/dev/null` kept stdout parseable JSON with `init_exit=0`.
- 2026-06-05: `go test ./core/onboarding ./cmd/stage/commands` passed after the real JSON-purity terminal checks and the guided footer command/doctor additions.
- 2026-06-05: `make build` built the repository-local binary as `./stage-bin`, and subsequent terminal verification in this slice used that artifact rather than the legacy `./stage` wrapper.
- 2026-06-05: `./stage-bin guidance-plan --project-dir $(mktemp -d) --skip-readiness | jq -r '.situation, .status_header'` returned `project_missing_config` and `This folder doesn't have StageServe settings yet.`, proving the hidden planner-inspection command works from a real binary invocation.
- 2026-06-05: `NO_COLOR=1 ./stage-bin --project-dir $(mktemp -d) --notui` and `NO_COLOR=1 ./stage-bin --project-dir $(mktemp -d) --cli` produced ANSI-free plain-text output (`notui_ansi=no`, `cli_ansi=no`), confirming both invocation-scoped opt-outs stay on the text fallback path.
- 2026-06-05: `./stage-bin --project-dir $(mktemp -d) > /tmp/stageserve-guidance.txt` exited `0` and wrote plain guidance starting with `StageServe` and `This folder doesn't have StageServe settings yet.`, confirming the non-TTY no-args path stays non-interactive.
- 2026-06-05: `STAGESERVE_NO_TUI=1 ./stage-bin --project-dir $(mktemp -d) > /tmp/stageserve-no-tui.txt` wrote the same plain-text missing-config guidance, confirming the shell-env no-TUI path stays on the text fallback.
- 2026-06-05: `./stage-bin up --help`, `attach --help`, `status --help`, `logs --help`, `down --help`, and `detach --help` all exited `0` and rendered direct help text without entering the guided root path.
- 2026-06-05: after the direct-command copy pass, rebuilt help text led with StageServe user goals rather than Docker/gateway mechanics, for example `Run this project`, `Add this project to StageServe`, and `Removes the current project from StageServe and clears its local project URL without touching your files or project settings.`
- 2026-06-05: `go test ./core/guidance` passed after adding label-split coverage proving the running-project planner no longer exposes `detach` as a primary easy-mode choice and the text fallback keeps `Add this project to StageServe` while still surfacing direct commands such as `stage attach`.
- 2026-06-05: `sed -n '1,120p' README.md | rg -n 'docker compose|network|volume|gateway alias|nginx|attach|detach|daemon|gateway|compose|container|registry|runtime|state'` no longer found the old Docker/gateway implementation terms in the first-run path beyond unavoidable references such as `Docker Desktop` as a prerequisite and later runtime-reference handoffs.
- 2026-06-05: after the final plain-language cleanup pass, `go test ./core/guidance` still passed and the top README first-run slice only retained the intentional handoff to the deeper runtime contract rather than exposing `attach`, `detach`, `daemon`, `compose`, `container`, `registry`, or `state` as first-level guidance.
- 2026-06-05: `go test ./core/guidance ./core/onboarding ./cmd/stage/commands ./core/config ./core/lifecycle ./observability/status ./infra/gateway` passed, giving the focused final package sweep required by the spec's closeout checklist.
- 2026-06-05: direct daemon-backed `stage attach`, `stage detach`, `stage up`, `stage status`, `stage logs`, and `stage down` were not rerun end-to-end in a disposable live project during this slice. Treat those as explicit real-daemon follow-up gaps; the current closeout evidence for this slice covers direct help, JSON purity, planner inspection, and focused package validation.
- 2026-06-05: easy-mode language review with the current planner, text fallback, help copy, and README first-run path confirmed that primary labels stay goal-first (`Create project settings`, `Run this project`, `Add this project to StageServe`, `Remove this project from StageServe`) while direct command names remain discoverable through help and show-commands output.

## Final Evidence Summary (2026-06-05)

**Binary:** `make build` → `./stage-bin` (repository checkout, `codex/007-harden-TUI-and-other-interactions` branch)

**Verified in this spec run:**

| Scenario | Result |
|---|---|
| Bare `stage` in TTY — missing-config | ✓ guided missing-config screen, `Create project settings` highlighted |
| Bare `stage` in TTY — configured stopped | ✓ `This project is ready to run.`, `▶ Run this project` highlighted |
| Bare `stage` in TTY — machine not ready (no state dir) | ✓ `Your computer isn't ready yet.`, setup checklist, `State directory` item highlighted |
| Non-TTY bare `stage` | ✓ exits 0, plain guidance, no prompts |
| `--notui`, `--cli`, `STAGESERVE_NO_TUI=1` | ✓ all use text fallback, ANSI-free output |
| `stage setup --json` | ✓ parseable JSON, exit 0 |
| `stage doctor --json` | ✓ parseable JSON, exit 0 |
| `stage init --json` | ✓ parseable JSON, exit 0 |
| Direct `--help` for all subcommands | ✓ all bypass guided root |
| `guidance-plan --skip-readiness` (hidden) | ✓ correct JSON situation/header |
| NFR-001 first-render time | ✓ 0.431s (< 500ms) after switching to cheap state-dir heuristic |
| Installer handoff (test mode) | ✓ interactive → `stage`, non-interactive → explicit steps |
| Focused package sweep | ✓ all 7 packages pass |

**Real-daemon gaps (follow-up required):**

- Live end-to-end `stage attach`, `stage detach`, `stage up`, `stage status`, `stage logs`, `stage down` not rerun in a disposable configured project during this spec run.
- Running-project TTY screen (`This project is running at …`) was previously verified (2026-05-28 via scripted live run); not rerun with the current binary.
- Keyboard-only navigation was verified through shell.go code inspection and existing test coverage; live TTY keyboard validation is a follow-up item.

**Intentionally deferred visual work:**

- A clearer `More…` or advanced surface to surface direct commands without relying only on footer key shortcuts.
- Long-value handling at narrow widths beyond terminal wrapping.
