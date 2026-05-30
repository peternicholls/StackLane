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
