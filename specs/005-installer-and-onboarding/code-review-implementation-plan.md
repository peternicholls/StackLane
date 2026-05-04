# Implementation Plan — Code Review Fixes

**Status**: Ready to begin  
**Reviewed**: 3 May 2026  
**Target completion**: By end of spec-005 wrap-up

---

## Plan Overview

This plan addresses all 12 issues from the code review in a logical sequence that respects dependencies and prioritizes correctness over polish.

**Four phases:**
1. **Phase 1 (Critical)** — Fix three correctness bugs that can corrupt state or throw confusing errors
2. **Phase 2 (Security)** — Fix shell injection vulnerability before public release
3. **Phase 3 (Design)** — Resolve design gaps and interface inconsistencies
4. **Phase 4 (Polish)** — Dead code removal, UX copy, performance micro-optimizations

---

## Phase 1: Critical Correctness (Est. 30 min)

All three issues are isolated, small changes with clear acceptance criteria. Each can be implemented and tested independently, but should be batched together for efficiency.

### T1.1 — CR-001: Fix Linux DNS/mkcert status envelope mismatch

**Issue**: Both Linux stubs emit `Status: StatusNeedsAction` + `Code: "unsupported-os"`, which violates the contract that `exit_code: 3` means unsupported OS.

**File**: `core/onboarding/readiness_dns_linux.go`

**Changes**:
- Change `Status: StatusNeedsAction` to `Status: StatusError` in both `checkDNS()` and `checkMkcert()`
- Verify consistency with `readiness_dns_other.go` (already uses `StatusError`)

**Test**: Existing tests should pass; verify JSON output envelope matches `readiness_dns_other.go` pattern

**Acceptance**: JSON output has `overall_status: error` with `exit_code: 3` (not `needs_action`)

---

### T1.2 — CR-002: Add shell quoting to env value rendering

**Issue**: `renderEnv` writes unquoted `siteName` and `docroot` to `.env.stageserve`, allowing injection if values contain special characters or quotes.

**Files**: 
- `core/onboarding/project_env.go` (renderEnv)
- `cmd/stage/commands/project_env.go` (reference implementation for shellQuote)

**Changes**:
- Extract or copy the `shellDoubleQuote` logic from `cmd/stage/commands/project_env.go` into a shared utility (e.g., `core/onboarding/shell_quote.go` or add to `project_env.go`)
- Update lines 82–91 in `core/onboarding/project_env.go` to quote both values:
  ```go
  b.WriteString("SITE_NAME=" + shellQuote(siteName) + "\n")
  b.WriteString("DOCROOT=" + shellQuote(docroot) + "\n")
  ```

**Test**: 
- Add test cases with values containing spaces, `$`, backticks, and quotes
- Verify the resulting `.env` file is parseable by `source` / `export`

**Acceptance**: `ValidateProjectEnv` or integration test confirms env file with special chars round-trips correctly

---

### T1.3 — CR-003: Propagate UserHomeDir errors

**Issue**: `os.UserHomeDir()` errors are silently discarded in `setup.go` and `doctor.go`, leading to state paths under `/` if HOME is unavailable.

**Files**:
- `cmd/stage/commands/setup.go` (line 48)
- `cmd/stage/commands/doctor.go` (line 31)

**Changes**:
- Replace `home, _ := os.UserHomeDir()` with proper error handling in both files
- Return `fmt.Errorf("cannot determine home directory: %w", err)` if the call fails

**Test**: 
- Existing tests should continue to pass
- Add a unit test that mocks `os.UserHomeDir` to return an error and verify the command exits with the correct error message

**Acceptance**: Running `setup` / `doctor` in an environment without `$HOME` produces a clear error (not a silent root-path state directory)

---

## Phase 2: Security (Est. 45 min)

### T2.1 — CR-004: Fix osascript shell injection vulnerability

**Issue**: `platform/dns/macos.go` uses Go `%q` formatting (not POSIX shell quoting) to build paths for `osascript ... with administrator privileges`. A crafted `STAGESERVE_STATE_DIR` can inject arbitrary commands executed as root.

**File**: `platform/dns/macos.go` lines 192–196 (`installResolver`)

**Changes**:
- Replace `fmt.Sprintf` with `%q` with a dedicated POSIX shell-escaping function
- Define a `shEscape(s string) string` helper that single-quote-wraps and handles embedded quotes:
  ```go
  func shEscape(s string) string {
      return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
  }
  ```
- Rewrite the cmd construction:
  ```go
  shScript := fmt.Sprintf("/bin/mkdir -p %s && /bin/cp %s %s",
      shEscape(filepath.Dir(resolverFile)),
      shEscape(previewResolver),
      shEscape(resolverFile),
  )
  err := exec.Command("osascript", "-e",
      `do shell script "`+shScript+`" with administrator privileges`).Run()
  ```

**Test**:
- Add a unit test that passes a path containing `\"`, `$()`, and backslashes to `installResolver` and verifies they are safely escaped
- Verify existing smoke tests continue to pass

**Acceptance**: Unit test with malicious path confirms no injection; smoke test confirms normal paths still work

---

## Phase 3: Design Gaps (Est. 1.5 hours)

### T3.1 — CR-005: Resolve --recheck flag

**Issue**: `setup.go` accepts `--recheck` flag but never reads it; test only checks parsing, not behaviour.

**Files**: `cmd/stage/commands/setup.go`

**Options**:
- **Option A (Defer)**: Remove flag, test, and any documentation references. Re-add when feature is planned.
- **Option B (Implement)**: Add caching logic to `ReadinessState` so subsequent calls return cached results, and `--recheck` bypasses the cache.

**Recommendation**: **Option A** (defer). The caching strategy requires broader changes to state management and is out of scope for spec-005 wrap-up.

**Changes** (Option A):
- Remove `Recheck` field from `setupFlags` struct
- Remove `cmd.Flags().BoolVar(...)` call
- Remove `TestSetup_RecheckFlagAccepted` test
- Search codebase for any other references and remove

**Acceptance**: No `--recheck` flag exists; `stage setup --recheck` produces "unknown flag" error

---

### T3.2 — CR-006: Create missing platform test for setup

**Issue**: Task T031 is checked off, but `setup_platform_test.go` does not exist. Unsupported-OS exit-code contract is untested at command level.

**File**: `cmd/stage/commands/setup_platform_test.go` (new)

**Changes**:
- Create test file with platform-specific build tags (e.g., `//go:build !darwin && !linux`)
- Implement `TestSetup_UnsupportedOSExitCode` that runs `setup --json` on an unsupported platform and verifies:
  - `exit_code: 3` in JSON output
  - Exit status is 3
  - `overall_status: error`

**Test approach**: Can skip on Darwin/Linux using build tags, or mock the platform detection if needed.

**Acceptance**: Test passes on unsupported platform stub; marked as passing in T031 checklist

---

### T3.3 — CR-007: Create Projector interface (blocks T3.4)

**Issue**: Three projectors (`JSONProjector`, `TextProjector`, `TUIProjector`) have incompatible signatures and are dispatched via copy-pasted switch blocks in three files.

**Files**: 
- `core/onboarding/types.go` (add interface)
- `core/onboarding/projection_json.go`
- `core/onboarding/projection_text.go`
- `core/onboarding/projection_tui.go`

**Changes**:
- Add to `types.go`:
  ```go
  type Projector interface {
      Project(r CommandResult) error
  }
  ```
- Update `TextProjector.Project` signature to return `error` (currently returns nothing)
- Update `TUIProjector.Project` signature to return `error` (currently returns nothing)
- Implement error handling in both (write errors must propagate, not be silent)
- Add factory function:
  ```go
  func NewProjector(mode OutputMode, w io.Writer) Projector {
      switch mode {
      case OutputModeJSON:
          return JSONProjector{W: w}
      case OutputModeTUI:
          return TUIProjector{W: w}
      default:
          return TextProjector{W: w}
      }
  }
  ```

**Test**: All existing projection tests should pass with the new error signatures

**Acceptance**: Interface is defined; all three projectors conform; factory compiles

---

### T3.4 — CR-008: Eliminate tripled dispatch (depends on T3.3)

**Issue**: Switch-on-mode dispatch block is copy-pasted verbatim into `setup.go`, `doctor.go`, `init.go`.

**Files**:
- `cmd/stage/commands/setup.go`
- `cmd/stage/commands/doctor.go`
- `cmd/stage/commands/init.go`

**Changes**:
- Replace all three dispatch blocks with a single call to the factory:
  ```go
  p := onboarding.NewProjector(mode, cmd.OutOrStdout())
  return p.Project(result)
  ```

**Test**: All command tests should pass; output should be identical to before

**Acceptance**: No switch blocks remain in the three commands; test coverage unchanged

---

## Phase 4: Polish (Est. 30 min)

### T4.1 — CR-009: Cache brewPrefix() calls

**Issue**: `platform/dns/macos.go` shells out to `brew --prefix` multiple times per `Bootstrap` call.

**File**: `platform/dns/macos.go`

**Changes**:
- Cache the result at the top of `Bootstrap`:
  ```go
  prefix, err := brewPrefix()
  if err != nil {
      return err
  }
  ```
- Pass `prefix` to `dnsmasqManagedFile`, `ensureDnsmasqInclude`, and any other helpers that need it
- Remove inline `brewPrefix()` calls within those helpers

**Test**: Existing tests should pass; verify subprocess count is reduced

**Acceptance**: Only one `brew --prefix` subprocess call per `Bootstrap`, not three

---

### T4.2 — CR-010: Remove dead code

**Issue**: `buildSetupCmd` helper in `setup_test.go` is defined but never called.

**File**: `cmd/stage/commands/setup_test.go`

**Changes**:
- Delete the `buildSetupCmd` function (lines 12–15)

**Test**: All tests should pass

**Acceptance**: Function is removed; tests unchanged

---

### T4.3 — CR-011: Fix install.sh messaging

**Issue**: `install.sh` prints "Launching first-run setup …" but does not actually launch it, just prints the command.

**File**: `install.sh` lines 210–213

**Changes**:
- Change wording from "Launching first-run setup …" to "To complete setup, run:"
- Or optionally: exec `stage setup --tui` (if it's guaranteed to be on PATH)

**Recommendation**: Change wording (safer, no PATH dependency).

**Test**: Manual inspection; run installer and verify message is clear

**Acceptance**: Message says "To complete setup" (not "Launching"); no misleading promises

---

### T4.4 — CR-012: Add doc note to ValidateDocroot

**Issue**: `ValidateDocroot` checks containment but not existence; config can reference nonexistent paths.

**File**: `core/onboarding/project_env.go`

**Changes**:
- Add comment to `ValidateDocroot` doc string:
  ```go
  // ValidateDocroot validates that docroot is within the project root.
  // Note: Existence is not validated. Use EnsureDocroot or similar if creation is needed.
  ```
- Optionally add an advisory in the `init` command output or step result when a docroot is created that doesn't yet exist

**Test**: No test changes needed; doc-only improvement

**Acceptance**: Doc comment is clear about the validation scope

---

## Implementation Checklist

- [ ] **Phase 1** (Critical): T1.1, T1.2, T1.3
- [ ] **Phase 2** (Security): T2.1
- [ ] **Phase 3** (Design): T3.1, T3.2, T3.3, T3.4
- [ ] **Phase 4** (Polish): T4.1, T4.2, T4.3, T4.4
- [ ] All tests pass: `go test ./...`
- [ ] Vet clean: `go vet ./...`
- [ ] Race detector clean: `go test -race ./...`
- [ ] Code review items closed

---

## Estimated Effort

| Phase | Issues | Est. Time | Blocker |
|-------|--------|-----------|---------|
| 1 | CR-001, CR-002, CR-003 | 30 min | No |
| 2 | CR-004 | 45 min | No |
| 3 | CR-005, CR-006, CR-007, CR-008 | 1.5 hrs | T3.3 → T3.4 |
| 4 | CR-009–CR-012 | 30 min | No |
| **Total** | **12** | **~2.75 hrs** | — |

---

## Success Criteria

- All 12 code review issues are either fixed or deferred with clear reasoning
- Test suite passes: `make test`
- No linting violations: `go vet ./...`, `make lint`
- No race detector failures: `go test -race ./...`
- Commit history is clean and grouped by issue (one commit per issue or phase)
- PR references the code review and closes all related issues
