# Code Review — Spec 005: Installer, Onboarding & Environment Readiness

**Date**: 3 May 2026  
**Reviewer**: GitHub Copilot  
**Scope**: All implementation delivered for spec-005: `core/onboarding`, `cmd/stage/commands` (setup, doctor, init, onboarding_mode, project_env), `platform/dns`, and `install.sh`.  
**Build status**: ✅ All tests pass · `go vet` clean · race detector clean

---

## Executive Summary

The implementation is structurally sound. The domain model, projection adapters, and command wiring follow the spec contracts closely. The biggest risks are two correctness bugs that will produce confusing output or silently corrupt state in production, one security issue in the macOS privilege escalation path, and several design gaps that will require copy-paste maintenance as the codebase grows.

---

## Issues

### 🔴 Critical — Correctness

---

#### CR-001 · Linux DNS/mkcert stubs emit a self-contradictory envelope

**File**: [core/onboarding/readiness_dns_linux.go](../../core/onboarding/readiness_dns_linux.go)  
**Lines**: 5–23

**Problem**  
Both `checkDNS` and `checkMkcert` on Linux set `Status: StatusNeedsAction` alongside `Code: "unsupported-os"`. The two reduction functions then disagree:

- `ReduceExitCode` short-circuits on `Code == "unsupported-os"` and returns `ExitUnsupportedOS` (3).
- `DeriveOverallStatus` sees `StatusNeedsAction` and returns `OverallNeedsAction`.

The resulting JSON envelope says `overall_status: needs_action` with `exit_code: 3`, which violates the contract that defines exit 3 as the unsupported-OS sentinel. Consumers and smoke tests that branch on either field will see inconsistent signal.

The `!darwin && !linux` stub in `readiness_dns_other.go` gets this right by using `StatusError`.

**Fix**  
Change `Status` to `StatusError` in both Linux stubs, consistent with the "other" platform stub:

```go
// readiness_dns_linux.go — before
Status: StatusNeedsAction,
Code:   "unsupported-os",

// readiness_dns_linux.go — after
Status: StatusError,
Code:   "unsupported-os",
```

---

#### CR-002 · `renderEnv` writes unquoted, unsanitised values to `.env.stageserve`

**File**: [core/onboarding/project_env.go](../../core/onboarding/project_env.go)  
**Lines**: 82–91

**Problem**  
`siteName` and `docroot` are written by bare string concatenation:

```go
b.WriteString("SITE_NAME=" + siteName + "\n")
b.WriteString("DOCROOT=" + docroot + "\n")
```

If either value contains spaces, `$`, a backtick, or a quote character, the resulting file will be unparseable by `source` / `export` (the shell contract for `.env` files). An adversarially crafted value (e.g. from a flag passed to `stage init`) could inject arbitrary env-var assignments.

The parallel `renderEnvValue` + `shellDoubleQuote` pipeline in `cmd/stage/commands/project_env.go` already solves this problem correctly. It is not reused here.

**Fix**  
Mirror the quoting logic from the commands package, or move it to a shared utility and call it from both sites:

```go
b.WriteString("SITE_NAME=" + shellQuote(siteName) + "\n")
b.WriteString("DOCROOT=" + shellQuote(docroot) + "\n")
```

---

#### CR-003 · `os.UserHomeDir()` error silently discarded in setup and doctor

**Files**: [cmd/stage/commands/setup.go](../../cmd/stage/commands/setup.go) L48, [cmd/stage/commands/doctor.go](../../cmd/stage/commands/doctor.go) L31

**Problem**  
Both commands resolve the state directory with:

```go
home, _ := os.UserHomeDir()
stateDir = filepath.Join(home, ".stageserve-state")
```

`os.UserHomeDir` can fail in stripped or container environments where `$HOME` and the passwd database are both absent. When it does, `home` is `""` and `stateDir` silently becomes `/.stageserve-state`. The command then proceeds to stat or create a path at the filesystem root rather than surfacing a clear error.

**Fix**

```go
home, err := os.UserHomeDir()
if err != nil {
    return fmt.Errorf("cannot determine home directory: %w", err)
}
stateDir = filepath.Join(home, ".stageserve-state")
```

---

### 🟡 Security

---

#### CR-004 · Shell injection possible via `%q`-formatted paths in osascript privilege escalation

**File**: [platform/dns/macos.go](../../platform/dns/macos.go)  
**Lines**: 192–196

**Problem**  
`installResolver` constructs the admin shell command using Go's `%q` verb:

```go
cmd := fmt.Sprintf("mkdir -p %q && cp %q %q", filepath.Dir(resolverFile), previewResolver, resolverFile)
exec.Command("osascript", "-e", "do shell script \""+cmd+"\" with administrator privileges")
```

`%q` produces Go string-literal quoting (e.g. `"path"`, with `\"` for embedded quotes), not POSIX shell quoting. A `previewResolver` path that contains `\"`, `$(...)`, `` `...` ``, or `\` can break out of the double-quoted region inside the `do shell script` string and inject arbitrary commands executed with administrator privileges.

`previewResolver` is derived from `Settings.StateDir` and `Settings.Suffix`. `StateDir` ultimately comes from an environment variable (`STAGESERVE_STATE_DIR`). A user or process that controls that variable on a shared machine could escalate to root.

**Fix**  
Pass the paths as separate `argv` elements to `/bin/sh` rather than embedding them in a string:

```go
// Use osascript with a heredoc-style argument that never interpolates variables.
shScript := fmt.Sprintf("/bin/mkdir -p %s && /bin/cp %s %s",
    shEscape(filepath.Dir(resolverFile)),
    shEscape(previewResolver),
    shEscape(resolverFile),
)
err := exec.Command("osascript", "-e",
    `do shell script "`+shScript+`" with administrator privileges`).Run()
```

Where `shEscape` single-quote-wraps the path (replacing `'` with `'\''`), which is injection-safe for POSIX shell regardless of path content.

---

### 🟡 Design Gaps

---

#### CR-005 · `--recheck` flag accepted, stored, but never consulted

**File**: [cmd/stage/commands/setup.go](../../cmd/stage/commands/setup.go)  
**Lines**: 24, 85

**Problem**  
`setupFlags.Recheck` is declared and registered:

```go
cmd.Flags().BoolVar(&f.Recheck, "recheck", false, "Rerun full check inventory even if already healthy")
```

The `RunE` closure never reads `f.Recheck`. The test `TestSetup_RecheckFlagAccepted` passes because it only checks that the flag parses without error — it does not verify the flag changes behaviour. The flag is therefore a documented promise with no implementation.

**Options**: Either implement the intended caching/short-circuit logic that `--recheck` would bypass, or remove the flag and its test until the feature exists.

---

#### CR-006 · `setup_platform_test.go` marked complete in T031 but the file does not exist

**Task**: T031 `[x]` in tasks.md  
**Expected file**: `cmd/stage/commands/setup_platform_test.go`

**Problem**  
The task is checked off, but the file is absent from the filesystem. The unsupported-OS exit-code contract for `setup` is untested at the command adapter level. Any regression in `ReduceExitCode` or the platform stubs would be invisible.

**Fix**: Create the file with at least one test asserting that running `setup --json` on an unsupported platform produces `exit_code: 3` in the JSON output, or uncheck T031 and re-open it.

---

#### CR-007 · Projector interface inconsistency — three different `Project` signatures

**Files**: `core/onboarding/projection_json.go`, `projection_text.go`, `projection_tui.go`

**Problem**  
The three projectors have incompatible method signatures:

| Projector | Signature |
|-----------|-----------|
| `JSONProjector` | `Project(r CommandResult) error` |
| `TextProjector` | `Project(r CommandResult)` |
| `TUIProjector` | `Project(r CommandResult)` |

There is no common `Projector` interface. As a result, the switch-on-mode dispatch block is copy-pasted verbatim into `setup.go`, `doctor.go`, and `init.go`. The text and TUI projectors write to `io.Writer` but swallow write errors silently. Adding a fourth output mode (e.g. a machine-readable NDJSON stream) requires editing three files.

**Fix**  
Define a `Projector` interface in `types.go`:

```go
type Projector interface {
    Project(r CommandResult) error
}
```

Have all three concrete types implement it (add the `error` return to `TextProjector` and `TUIProjector`). Move the mode-to-projector switch to a single `NewProjector(mode OutputMode, w io.Writer) Projector` factory. Each command then calls:

```go
p := onboarding.NewProjector(mode, cmd.OutOrStdout())
return p.Project(result)
```

---

### 🟠 Code Quality

---

#### CR-008 · Projection dispatch triplicated across setup, doctor, and init

**Files**: `setup.go`, `doctor.go`, `init.go`

Directly related to CR-007. The switch block currently appears three times:

```go
switch mode {
case onboarding.OutputModeJSON:
    p := onboarding.JSONProjector{W: cmd.OutOrStdout()}
    return p.Project(result)
case onboarding.OutputModeTUI:
    p := onboarding.TUIProjector{W: cmd.OutOrStdout()}
    p.Project(result)
default:
    p := onboarding.TextProjector{W: cmd.OutOrStdout()}
    p.Project(result)
}
```

The JSON branch returns its error; the other two discard theirs. This is a maintenance trap: any future change to the dispatch logic must be applied in three places consistently.

---

#### CR-009 · `brewPrefix()` subprocesses run multiple times per `Bootstrap` call

**File**: [platform/dns/macos.go](../../platform/dns/macos.go)

`Bootstrap` calls `dnsmasqManagedFile` → `brewPrefix`, then `ensureDnsmasqInclude` → `dnsmasqMainConf` → `brewPrefix`, then `ensureDnsmasqInclude` reads `brewPrefix` again inline. Each call shells out to `brew --prefix`. On a loaded machine, spawning three redundant subprocesses adds latency and makes tracing errors harder.

**Fix**: Cache the result at the top of `Bootstrap` and pass it to helpers as an argument.

---

### 🟢 Minor / Cosmetic

---

#### CR-010 · `buildSetupCmd` helper in `setup_test.go` is dead code

**File**: [cmd/stage/commands/setup_test.go](../../cmd/stage/commands/setup_test.go)  
**Lines**: 12–15

`buildSetupCmd` is defined but never called. All tests in the file use `NewRoot("test")` directly. Remove it to avoid misleading future readers about an alternative construction pattern.

---

#### CR-011 · `install.sh` "Launching first-run setup …" message promises action it does not take

**File**: [install.sh](../../install.sh)  
**Lines**: 210–213

The TTY branch prints:

```
==> Launching first-run setup …
  Run: stage setup --tui

  Or start setup now:  stage setup --tui
```

The banner says "Launching" but immediately prints the command without executing it. This reads as a bug to users. Either exec `stage setup --tui` (subject to it being on PATH) or change the wording to "To complete setup, run:".

---

#### CR-012 · `ValidateDocroot` does not check that the path exists

**File**: [core/onboarding/project_env.go](../../core/onboarding/project_env.go)

`ValidateDocroot` verifies containment but not existence. `stage init --docroot nonexistent` will write a config that references a path that does not yet exist, with no warning. The current behaviour is technically documented (the function name says "validate", not "ensure"), but a note in the doc comment or an advisory `StepResult` in the init command would save operator confusion.

---

## Issue Index

| ID | File | Severity | Category | Actionable |
|----|------|----------|----------|------------|
| CR-001 | `readiness_dns_linux.go` | 🔴 Critical | Correctness | Fix now |
| CR-002 | `project_env.go` (onboarding) | 🔴 Critical | Correctness / Security | Fix now |
| CR-003 | `setup.go`, `doctor.go` | 🔴 Critical | Correctness | Fix now |
| CR-004 | `platform/dns/macos.go` | 🟡 High | Security | Fix before release |
| CR-005 | `setup.go` | 🟡 Medium | Design | Fix or remove |
| CR-006 | `setup_platform_test.go` | 🟡 Medium | Test coverage | Create file or re-open task |
| CR-007 | Projector types | 🟡 Medium | Design | Refactor |
| CR-008 | `setup/doctor/init.go` | 🟠 Low | Maintainability | Refactor with CR-007 |
| CR-009 | `platform/dns/macos.go` | 🟠 Low | Performance | Fix at convenience |
| CR-010 | `setup_test.go` | 🟢 Trivial | Dead code | Delete |
| CR-011 | `install.sh` | 🟢 Trivial | UX copy | Fix wording |
| CR-012 | `project_env.go` (onboarding) | 🟢 Trivial | Doc/UX | Add note |

---

## Recommended Fix Order

1. **CR-001, CR-002, CR-003** — correct the Linux exit-code mismatch, add env-value quoting, and propagate the `UserHomeDir` error. All are small, isolated changes.
2. **CR-004** — replace the `%q`-based osascript shell construction before the first public release.
3. **CR-005 / CR-006** — either implement `--recheck` behaviour or remove the flag; create the missing platform test.
4. **CR-007 / CR-008** — introduce the `Projector` interface and factory to eliminate the tripled dispatch. This is the right time to fix the silent write-error discard in TextProjector and TUIProjector.
5. **CR-009 through CR-012** — address at convenience during the next polish pass.
