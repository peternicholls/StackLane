# Code Review: Phases 1–4

**Date:** 2026-05-30  
**Branch:** `codex/007-harden-TUI-and-other-interactions`  
**Scope:** Roadmap phases 1–4 (T001–T072f). Phase 5 work (T073–T100) is explicitly out of scope.  
**Build/test state at review:** `go vet ./...` passes, `go test ./core/guidance ./cmd/stage/commands ./core/lifecycle` passes.

---

## Summary

The implementation is structurally sound and well-aligned with the roadmap contract. The guided shell, planner, text fallback, lifecycle orchestrator, and command contract are all coherent and the key test assertions are meaningful. The findings below are ordered from highest to lowest severity within each section.

---

## Critical

### CR-01 — `SituationMachineNotReady` is unreachable from bare `stage`

**File:** `cmd/stage/commands/guidance_context.go`  
**Roadmap ref:** T048 (route setup checks through guided shell)

`collectGuidedContext` passes only `RuntimeStatus` to `CollectOptions`. It never injects `MachineReadiness`. The planner's `SituationMachineNotReady` branch requires `ctx.MachineReadiness.Checked == true && ctx.MachineReadiness.Blocked == true`, but that struct is always zero-valued when arrived at from bare `stage`.

A user whose Docker daemon is not running will hit `SituationProjectReadyToRun` or `SituationProjectMissingConf` instead of a clear "Your computer isn't ready yet." screen. The `setup` and `doctor` commands correctly populate readiness, but the guided root path silently skips the check.

**Fix:** Wire `buildMachineReadinessResult` (already in `cmd/stage/commands/readiness.go`) into a `CollectOptions.MachineReadiness` adapter in `guidance_context.go`, similar to how `guidedRuntimeStatus` is wired.

---

### CR-02 — `--mysql-password` is a persistent root flag (cleartext in process list)

**File:** `cmd/stage/commands/root.go` line 74  
**Roadmap ref:** T008b (security cleanup)

`MySQLPassword` is registered as a `PersistentFlag` on the root command with `--mysql-password`. The value is plaintext and will appear in `ps aux` output and shell history on any operating system. The other MySQL credentials (`--mysql-user`, `--mysql-database`) share the same exposure surface but the password flag is the most sensitive.

**Fix:** Accept the password via environment variable (`STAGESERVE_MYSQL_PASSWORD`) and document that pattern in `.env.stageserve.example`. The flag may remain for convenience but should not be marked `Required` and should carry a help note about the env-var alternative.

---

## Major

### CR-03 — `UpOptions.Detach` field is never honoured

**File:** `infra/compose/compose.go`  
**Roadmap ref:** T078d

Both branches of the `if opts.Detach` check append `"-d"`:

```go
if opts.Detach {
    args = append(args, "-d")
} else {
    args = append(args, "-d") // we always want detach for orchestration
}
```

`Detach bool` in `infra/compose/types.go` is dead. All callers set `Detach: true` anyway, but this creates a misleading API surface and will confuse anyone who tries to set `Detach: false` expecting the flag to be dropped.

**Fix:** Remove the `else` branch and the `Detach` field. If a non-detached path is ever needed, add it explicitly at that point.

---

### CR-04 — `Attach()` silently ignores registry read error

**File:** `core/lifecycle/orchestrator.go`

```go
registry, _ := o.D.State.Registry()
```

A registry read failure here produces an empty route slice, which causes the gateway config to be written with no routes. All other projects' local URLs will stop responding. The error is discarded, so there is no operator message and the attach call returns success.

**Fix:** Return a `Wrap("registry", ...)` error here, consistent with every other registry read in the orchestrator.

---

### CR-05 — `Attach()` `Wrap()` calls have empty remedy strings

**File:** `core/lifecycle/orchestrator.go`

Three `Wrap()` calls in `Attach()` pass `""` for the remedy argument:

```go
return Wrap("save-state", cfg.Slug, err, "")
return Wrap("tls-cert", cfg.Slug, err, "")
return Wrap("gateway-config", cfg.Slug, err, "")
```

The `StepError` contract requires a stated operator next action per the comment at the top of `errors.go`. These three steps already have remedy text in `Up()` (where the same operations appear) — they can be copied across.

---

### CR-06 — `readRecord()` in `guidance/context.go` bypasses the state package

**File:** `core/guidance/context.go`

`readRecord()` reads `stateDir/projects/slug.json` directly with `os.ReadFile` + `json.Unmarshal`. The `state` package already provides `Store.Load(slug)` which owns that path convention. If the file layout ever changes (compression, schema migration), `readRecord()` will silently drift.

The function exists because `guidance` would otherwise depend on the `state.Store` interface, but `state.ErrNotFound` is already imported and `state.Record` is referenced through the `GuidedContext.ProjectState` field, so the package boundary is already crossed.

**Fix:** Accept a `state.StateStore` read-seam in `CollectOptions` (or accept the existing store interface directly) and remove the private `readRecord` reimplementation.

---

## Minor

### CR-07 — `tuiDisabledByEnv()` is duplicated

**Files:** `cmd/stage/commands/onboarding_mode.go` and `core/guidance/capability.go`

Both files parse `STAGESERVE_NO_TUI` with identical truthy/falsy logic. `capability.go` uses `envTruthy()`, `onboarding_mode.go` has an inline version. If the env variable name or falsy values ever change, both need updating.

**Fix:** Export `guidance.envTruthy` or move the canonical implementation to a shared location and have `onboarding_mode.go` call it. Alternatively, call `guidance.DetectCapability` from `resolveOutputMode` and use the result directly.

---

### CR-08 — `resolveOutputMode` hardwires `os.Stdout.Fd()` for TTY detection

**File:** `cmd/stage/commands/onboarding_mode.go`

```go
case isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()):
```

Unlike `guidance.DetectCapability()` which accepts injected `*os.File` parameters, `resolveOutputMode` reads the real `os.Stdout` directly. This makes the function hard to test without redirecting the real stdout. The tests in `init_test.go` and `setup_test.go` work around this by relying on the non-TTY fallback path in CI.

**Fix:** Accept a `tty bool` parameter (pre-computed by the caller with injected fds), or call `guidance.DetectCapability` once at command entry and thread the result through.

---

### CR-09 — Local variable `context` shadows the `context` package in `guidance/context.go`

**File:** `core/guidance/context.go`

```go
context := GuidedContext{
    CWD: cfg.Dir,
    ...
}
```

The local variable `context` shadows the imported `"context"` package. The code works because all uses of the package-level `context` occur before this line, but it is a code smell that will confuse grep, linters, and reviewers. The variable is returned at end of function.

**Fix:** Rename to `gctx` or `guided`.

---

### CR-10 — `DownAll()` has a partial-failure gap

**File:** `core/lifecycle/orchestrator.go`

`DownAll()` iterates projects and stops each one. If `stopProject()` fails for project N, the function returns immediately, leaving projects 0..N-1 stopped (state saved as `StateDown`) and projects N+1..end untouched. The caller receives a single error for project N with no information about which other projects were affected. There is no partial-success report.

This is a pre-existing concern tracked as a follow-up in the roadmap but worth noting explicitly here because the `--all` flag in `stage down` is now a documented active command (T001/T009 decision).

**Fix:** Collect all stop errors, return a joined multi-error, and ensure gateway sync runs even when individual stops fail (best-effort cleanup).

---

### CR-11 — Step 8 in `Up()` uses a stale registry snapshot

**File:** `core/lifecycle/orchestrator.go`

At step 8, `currentRoutes := routesFromRegistry(registry)` reuses the `registry` captured at step 2. If two `stage up` calls execute concurrently (two terminal windows), the second caller's step 8 will not see the first caller's new route in the registry, and may overwrite the gateway config without it.

StageServe is a single-developer CLI so true concurrency is rare, but the comment at the top of `orchestrator.go` does not say the tool is single-instance. It is worth either documenting the single-instance assumption or refreshing the registry at step 8.

---

### CR-12 — Footer render has a dead code branch

**File:** `core/guidance/shell.go`

```go
footer := footerHint(lineWidth, utility, confirming, editing)
if confirming {
    footer = footerHint(lineWidth, utility, confirming, editing) // identical call
}
```

The `if confirming` block reassigns `footer` to the exact same call. The `footerHint` function already accounts for the `confirming` parameter internally, so the outer `if` is a no-op.

**Fix:** Remove the redundant `if confirming` block.

---

### CR-13 — `shell.go` is 850 lines with interleaved model and render logic

**File:** `core/guidance/shell.go`  
**Roadmap ref:** T078c (partially deferred, but relevant to T072b)

`projectSettingsDraft` and all its methods, the `shellModel` Bubble Tea implementation, all render functions, and all style logic are in one file. The T072b audit work added good shared render helpers, but the file is still hard to navigate. The `projectSettingsDraft` edit logic in particular (lines ~300–500) is repeated almost verbatim twice in the file due to the read tool pagination exposing the same functions multiple times from different offsets — on inspection this appears to be a tool artefact, not actual duplication in the file itself. However, extracting `projectSettingsDraft` into its own file would improve clarity.

**This is already tracked as T078c.** Noting here for completeness.

---

### CR-14 — `Up()` step comment numbering gap

**File:** `core/lifecycle/orchestrator.go`

The code comments reference steps 1, 2, 4, 5, 6, 7, 8, 9, 10/11. Step 3 (TLS cert generation, now moved before step 4 gateway start) is not explicitly labelled. This makes the 11-step flow described in the package comment harder to trace.

**Fix:** Add `// Step 3: ensure TLS certificates.` before the `ensureSharedGatewayTLS` call at line ~68.

---

## Tests

The test coverage for the in-scope work is good:

- `planner_test.go` covers all 8 situations, the runtime-already-running attach branch, recovery step ordering, and the `MachineReadinessFromResult` adapter.
- `orchestrator_test.go` covers happy path, rollback on health failure, and post-up hook rollback.
- `root_guidance_test.go` covers no-args text output, `--cli` fallback, `--notui`, guided init action, overwrite guard, and stop confirmation copy.
- `help_contract_test.go` validates help text against documented examples.

**Gaps:**

- No test for `SituationMachineNotReady` being reached from the guided root path (directly related to CR-01).
- No test for `Attach()` registry error propagation (CR-04).
- `DownAll()` partial failure is not covered (CR-10).
- The `UpOptions.Detach` dead branch in `compose.go` has no test asserting that `Detach: false` drops the `-d` flag, which would have caught CR-03.

---

## Phase 5 / Follow-up Candidates

The following roadmap items are already tracked in Phase 5 but are now validated as genuine gaps by this review:

| Roadmap task | Finding |
|---|---|
| T078c | `shell.go` draft/render tangling |
| T078d | `UpOptions.Detach` dead field (CR-03) |
| T078b | Duplicate exit-code error types |
| T078f | Root-level flag ownership (incl. `--mysql-password` OPSEC) |

---

## Verdict

The codebase is in a healthy state for the guided UX work completed in phases 1–4. No finding here blocks current functionality, but CR-01 (machine-not-ready situation unreachable from guided root) and CR-04 (silent registry error in `Attach`) should be fixed before this branch is merged to `master`. The remaining findings can be addressed in Phase 5 or as a targeted cleanup PR.

---

## Field-Reported Issues — Planning

The following six issues were reported after phases 1–4 were marked complete. They are not regressions — they represent gaps in breadth of coverage and design ambition that the earlier phases did not address. Each is broken down into root cause and a concrete plan.

---

### FR-01 — `stage up` fails with "no such file or directory" for compose files

**Symptom:**
```
open /Users/peternicholls/docker/stageserve/docker-compose.shared.yml: no such file or directory
step shared-gateway failed: ...
```

**Root cause:** The compose files moved from the repo root to `stacks/20i/` during an earlier refactor. The `Makefile` `install-dev` target already copies them to `STACK_HOME/stacks/20i/`, but that target was not run after the move. The runtime STACK_HOME (`~/docker/stageserve`) contains only `.stageserve-state` — no compose files at all.

**Immediate fix:** Run `make install-dev`. This copies the correct files to `~/docker/stageserve/stacks/20i/`.

**Permanent fix (two tasks):**

1. **Pre-flight compose file check.** In `core/lifecycle/orchestrator.go`, before launching `infra/compose` commands, add an explicit existence check on `cfg.SharedFile` and `cfg.StackFile`. If either is missing, return a `StepError` with a remedy that names the install command:
   ```
   remedy: "Run `make install-dev` from the StageServe repo, or re-run the installer."
   ```
   This turns a cryptic `open … no such file or directory` from deep inside compose into a clear operator message at the top of the error chain.

2. **`install-dev` should verify the copy succeeded.** After the `cp` commands, add a `diff` or `cmp` check and print a warning if the installed files differ from the repo source (i.e., if a future move is missed again).

---

### FR-02 — `stage up` text output does not follow the established color scheme

**Symptom:** The semantic color palette established in `shell.go` (terminal indices 0–15, `verdictReady`/`verdictWarn`/`verdictError`) is applied only inside the guided TUI. Direct `stage up`, `stage down`, `stage status`, and `stage logs` output uses plain text or ad-hoc formatting that does not match.

**Root cause:** `guidance/text.go` (`RenderText`) was written as a minimal non-Charm fallback. The StepError display in `cmd/stage/commands/root.go` uses plain `fmt.Fprintf`. Neither pulls from the shared style registry in `shell.go`.

**Plan:**

1. Extract the color/style definitions from `shell.go` into a new file `core/guidance/styles.go`. Export `Styles` (the struct) and a `NewStyles(noColor bool)` constructor so both the TUI and text paths can use the same tokens.

2. Update `guidance/text.go` `RenderText` to accept a `Styles` and use `verdictReady`, `verdictWarn`, `verdictError` for output line prefixes (✓ / ⚠ / ✗), matching the visual language of the guided shell.

3. Update the `StepError` display in `root.go` (and `up.go`, `down.go` error paths) to call a shared `RenderStepError(w, err, styles)` helper rather than inline `fmt.Fprintf`.

4. Respect `NO_COLOR` and `--notui`/`--cli` in all three paths — the `Styles` constructor already handles this via the `noColor bool` parameter.

---

### FR-03 — `stage up` failures present no recovery options

**Symptom:** When `stage up` fails (e.g. port conflict, compose error, TLS failure), the user sees a plain error string and must independently decide what to do. There is no guided recovery path.

**Root cause:** The guided shell surfaces `SituationDriftDetected`, `SituationUnknownError`, and recovery `DecisionItems` only for the bare `stage` situation (i.e. persistent state from a previous run). A freshly-failed `stage up` returns an error to the shell and exits. The `Up()` function does roll back, but the UX after rollback is a plain error line.

**Plan:**

1. **Post-failure recovery TUI.** In `cmd/stage/commands/up.go`, when `Up()` returns a non-nil `StepError`, call a new function `runFailureRecovery(ctx, stepErr, shared)`. This function:
   - Builds a `NextActionPlan` with `Situation = SituationUnknownError` and populates `WorkItems` from the `StepError`'s step name and remedy.
   - Appends a `DecisionItems` slice with contextually appropriate recovery actions (e.g. "Try again", "Run doctor", "Open logs", "Show full error").
   - Runs this plan through `guidance.RunShell` (TUI-capable terminal) or `guidance.RenderText` (non-TTY / `--notui`).

2. **Extend planner for new recovery situations.** Add two new `Situation` constants:
   - `SituationStartFailed` — compose started but container(s) never became healthy.
   - `SituationGatewayFailed` — shared gateway could not be reached or configured.
   These allow the planner to provide more specific recovery copy rather than the generic `SituationUnknownError` path.

3. **Remedy text completeness.** As part of FR-03, audit all `Wrap()` calls in `Up()`, `Down()`, and `Attach()` and ensure every `remedy` string is non-empty (see also CR-05 above).

---

### FR-04 — Bare `stage` guided shell exposes only a small subset of project actions

**Symptom:** When a project is running, the guided shell shows only "Stop project" and "Detach from project". Key day-2 operations — opening the project URL in a browser, tailing logs, restarting a specific service, running `wp-cli` / `artisan`, attaching a shell to a container — are missing.

**Root cause:** The `DecisionItems` in `planner.go`'s `SituationProjectRunning` branch were designed as a minimal first pass. The planner only knows the `RuntimeSummary` (URL, port, state) and cannot yet derive which per-service actions are appropriate.

**Plan:**

1. **Extend `RuntimeSummary`** to carry `Services []ServiceSummary` (name, health, image). Populate this in `guidedRuntimeStatus()` from Docker label inspection (already done for gateway config — reuse the pattern).

2. **Add primary actions to the running-project plan:**
   - `open_browser` — opens `cfg.LocalURL` in the default browser using `open` (macOS) / `xdg-open` (Linux).
   - `view_logs` — launches an in-TUI log stream (a `viewport` bubbles component) for the selected service.
   - `restart_service` — runs `docker compose restart <service>` via `infra/compose`.

3. **Add utility-surface actions:**
   - `run_wp_cli` / `run_artisan` — available when the stack's `Capabilities` indicate a PHP project; opens a sub-shell in the app container.
   - `container_shell` — `docker exec -it <container> sh`.

4. **Handle `open_browser` in `tui.go`** `handleGuidedAction` using `exec.Command("open", url)` on Darwin and `exec.Command("xdg-open", url)` on Linux.

5. **Planner ordering:** Primary actions (Open, Logs) first; secondary (Restart); destructive (Stop) last. This follows the established least-to-most-invasive ordering from T069.

---

### FR-05 — Confirmation prompts are not visually prominent enough

**Symptom:** When the guided shell asks "Stop this project? [y/N]", the confirmation text appears inline in the normal flow. For a destructive action like `down`, this is easy to miss or accidentally confirm. There is no visual separation from the surrounding content.

**Root cause:** `renderConfirmView` in `shell.go` renders the confirmation as a padded block using `sectionTitle` and `bodyText` — the same visual weight as a normal surface. There is no border, modal overlay, or colour distinction to signal "this is a consequential decision".

**Plan:**

1. **Bordered confirmation panel.** Replace the inline `renderConfirmView` with a Lip Gloss bordered box. Use `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())` with `verdictWarn` (yellow/3) as the border foreground. The panel width should be capped at 60 columns and centred in the terminal.

2. **Destructive action colour.** For actions where `ActionKind` is `down`, `detach`, or any future delete/purge action, use `verdictError` (red/1) as the border foreground and prefix the title with a `⚠` glyph, distinct from the `!` used for warnings.

3. **Explicit keybind hint.** The confirmation footer hint should be rendered inside the panel, not in the global footer bar, so the user's eye does not need to travel. Something like:  
   `  [y] Confirm   [n / Esc] Cancel  ` rendered in `dimText` style.

4. **Terminal-width guard.** If `lineWidth < 44`, fall back to the existing inline rendering (narrow-width safe path, consistent with the <58-column layout already in the shell).

5. **Huh `Confirm` field as alternative.** Evaluate using `huh.NewConfirm()` as an embedded form step (the `charm-huh-forms` skill covers Huh-in-BubbleTea embedding). This gives accessibility benefits (screen reader friendly) at the cost of introducing a `huh` model step in the `shellModel` update loop. Decide based on whether the `huh` dependency is acceptable.

---

### FR-06 — Long-running operations have no visual feedback; Bubble Tea potential is underused

**Symptom:** `stage up` and `stage down` can take 30–120 seconds (container build, health checks, DNS propagation). During this time the terminal shows nothing. The broader TUI offers no spinners, progress bars, animated transitions, log streaming, or any of the richer Bubble Tea/Bubbles component features available in the Charm ecosystem.

**Root cause:** The guided shell was designed around instantaneous state display. `handleGuidedAction` in `tui.go` calls the lifecycle orchestrator synchronously — the Bubble Tea event loop is blocked for the duration. There are no `tea.Cmd` async dispatches for long-running operations.

**Plan (ordered by impact, each independently shippable):**

1. **Async action execution with spinner.**
   - Add a new `shellModel` state flag: `running bool` and `runningLabel string`.
   - Extract the `executeGuidedAction` call into a `tea.Cmd` (returns a `tea.Msg` when done). Use `spinner.New()` from `github.com/charmbracelet/bubbles/spinner` with `spinner.Dot` style, coloured with `accentText`.
   - While `running == true`, render a spinner line replacing the decision items: `  ⠿  Starting project…` (or whatever `runningLabel` is set to).
   - On completion, receive the result `tea.Msg` and transition back to normal state, triggering a context re-collect.

2. **Real-time log streaming in TUI.**
   - Add a `viewport.Model` from `github.com/charmbracelet/bubbles/viewport` as an optional overlay surface in `shellModel`.
   - When the user selects `view_logs`, launch `docker compose logs --follow <service>` as an `exec.Cmd` with its stdout piped. Send each line as a `tea.Msg` to the viewport. Use `viewport.GotoBottom()` on each new line.
   - The `charm-bubbletea-components` skill covers the viewport component and the exec-pipe pattern.

3. **Progress indicator for multi-step `stage up`.**
   - Add a `progress.Model` from `github.com/charmbracelet/bubbles/progress` to the running state.
   - Expose a progress channel from `orchestrator.Up()` (one message per completed step, total 11 steps). Advance the progress bar as messages arrive.
   - Use `harmonica` spring physics (from the `charm-tui-motion-observability` skill) to smooth the bar animation.

4. **Animated state transitions.**
   - Use `lipgloss.NewStyle().Faint(true)` fade-out on the previous surface content while `running == true`.
   - On transition from running → done (success), briefly flash the status line in `verdictReady` before settling.
   - Keep animations subtle — no more than 300ms total. Follow the motion budget in the `charm-tui-motion-observability` skill.

5. **Key-binding help bar.**
   - Add a `help.Model` from `github.com/charmbracelet/bubbles/help` at the bottom of every surface. Map to a `key.Map` struct per surface so `?` shows full keybindings.
   - The `charm-bubbletea-components` skill documents the help/key-binding pattern.

**Skills to load before implementation:** `charm-tui-builder`, `charm-bubbletea-components`, `charm-tui-motion-observability`, `charm-huh-forms` (for FR-05 Huh option).
