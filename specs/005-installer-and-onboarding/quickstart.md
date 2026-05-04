# Quickstart: Implementing Spec-005

## Goal

Implement and validate installer + onboarding command surfaces in this order:

1. installer handoff
2. shared onboarding runtime and projection adapters
3. machine readiness and project env ownership modules
4. `stage setup`, `stage doctor`, and `stage init` command adapters
5. docs and contract alignment

## Prerequisites

- Go 1.26 toolchain available.
- Docker Desktop installed on macOS.
- Writable local checkout of StageServe.
- If validating against a deployed stack copy under `$HOME/docker/stageserve-stack`, sync the repo changes into that deployed copy before running operator-facing checks.

## TDD Loop

Apply this loop for every implementation slice in spec-005:

1. Choose one behavior from the active contract.
2. Write the narrowest failing test for that behavior.
3. Run only that focused test and confirm it is red.
4. Implement the minimum code to make it green.
5. Re-run the same focused test.
6. Refactor locally only while tests stay green.

Do not write all tests for a story up front. Use vertical slices only.

## Initial Tracer Bullets

Start implementation with these concrete red-green-refactor slices, in order.

| Slice | Behavior to prove first | Test file | Suggested test name | Focused command |
|-------|--------------------------|-----------|----------------------|-----------------|
| 1 | Exit code reducer prefers `unsupported-os` over lower severities | `core/onboarding/runtime_test.go` | `TestReduceExitCode_PrefersUnsupportedOS` | `go test ./core/onboarding -run TestReduceExitCode_PrefersUnsupportedOS` |
| 2 | Overall status becomes `needs_action` when any step needs action and none error | `core/onboarding/runtime_test.go` | `TestOverallStatus_NeedsActionWithoutError` | `go test ./core/onboarding -run TestOverallStatus_NeedsActionWithoutError` |
| 3 | Docker binary readiness returns actionable non-ready result when binary is missing | `core/onboarding/machine_readiness_test.go` | `TestDockerBinaryCheck_MissingBinary` | `go test ./core/onboarding -run TestDockerBinaryCheck_MissingBinary` |
| 4 | Setup reports `config.dns_suffix=needs_action` in non-interactive mode when suffix is absent | `cmd/stage/commands/setup_test.go` or `core/onboarding/runtime_test.go` | `TestSetup_NonInteractiveMissingSuffix` | `go test ./cmd/stage/commands -run TestSetup_NonInteractiveMissingSuffix` |
| 5 | Project env validation rejects docroot outside project root and writes no file | `core/onboarding/project_env_test.go` | `TestValidateProjectEnv_RejectsDocrootOutsideProjectRoot` | `go test ./core/onboarding -run TestValidateProjectEnv_RejectsDocrootOutsideProjectRoot` |
| 6 | Doctor reports unhealthy gateway with exact remediation and stays read-only | `cmd/stage/commands/doctor_test.go` or `core/onboarding/machine_readiness_test.go` | `TestDoctor_UnhealthyGateway` | `go test ./cmd/stage/commands -run TestDoctor_UnhealthyGateway` |

For each slice:

1. Write the named test first.
2. Confirm the focused command is red.
3. Implement only enough code to pass that one test.
4. Re-run the same focused command.
5. Only then widen to the containing package test suite.

## Step 1: Add command entrypoints

1. Add command constructors in `cmd/stage/commands`:
- `NewSetup(flags *SharedFlags)`
- `NewDoctor(flags *SharedFlags)`
- `NewInit(flags *SharedFlags)`
2. Register commands in root command wiring.
3. Add command flag sets per contract:
- setup: `--suffix`, `--recheck`, `--non-interactive`, `--tui`, `--no-tui`, `--json`
- doctor: `--json` (+ no interactive behavior)
- init: `--site-name`, `--docroot`, `--project-dir`, `--force`, `--non-interactive`, `--no-tui`, `--json`

Keep these scaffolds thin; do not fill them with behavior before the first failing tests exist for the deep modules they call.

## Step 2: Establish the owning seams

1. Treat `contracts/cli-onboarding-contract.md` as the primary command-contract module.
2. Treat `data-model.md` and `contracts/json-envelope.schema.json` as the primary step/result schema module.
3. Keep `spec.md` focused on requirements and operator-facing interaction contract.
4. Keep command adapters thin: flags, mode selection, and top-level policy only.

## Step 3: Introduce shared onboarding runtime

Start with `core/onboarding/runtime_test.go` and one failing behavior test for envelope reduction or projection behavior before implementing the runtime.

1. Implement shared result types aligned to `data-model.md`:
- `StepResult`
- `CommandResultEnvelope`
2. Implement deterministic exit code reducer (`0`/`1`/`2`/`3`) in the runtime module.
3. Implement projection adapters behind the runtime seam:
- text projection
- JSON projection
- TUI projection

## Step 4: Implement shared domain modules

Open each new module with a failing behavior test before adding implementation.

1. Implement a machine readiness module that owns Docker, DNS, state-dir, port, mkcert, and unsupported-os behavior.
2. Implement a project env ownership module that owns `.env.stageserve` validation, overwrite, preservation, and allowed-key rules.
3. Route both `stage init` and the silent `ensureProjectEnvFile` helper through the same project-env ownership module.
4. Keep these as deep modules used by command adapters rather than duplicating their implementation per command.

## Step 5: Implement `stage setup`

Begin with the smallest failing setup behavior test, preferably in `core/onboarding/machine_readiness_test.go` or `core/onboarding/runtime_test.go`, and only then wire the command adapter.

1. Implement setup as a command adapter on top of the shared runtime and machine readiness modules.
2. Add suffix resolution logic:
- existing stack-home `SITE_SUFFIX`
- `--suffix` override (`develop|dev` only)
- first-run prompt (`.develop` default)
3. Add DNS bootstrap behavior with explicit confirmation gate.
4. Add mkcert check + optional install confirmation when suffix is `dev`.
5. Enforce non-interactive and JSON prompt suppression semantics.

## Step 6: Implement `stage doctor`

Begin with the smallest failing read-only diagnostic behavior test in `core/onboarding/machine_readiness_test.go`, then add the doctor adapter code needed for that behavior.

1. Implement doctor as a read-only command adapter on top of the shared runtime and machine readiness modules.
2. Add gateway-specific checks as doctor-only policy layered on top.
3. Ensure all checks print or emit remediation on non-ready status.
4. Ensure no writes and no privilege escalation paths.
5. Emit `code=unsupported-os` for platform-specific checks on unsupported platforms while still running portable checks.

## Step 7: Implement `stage init`

Begin with the smallest failing ownership or validation test in `core/onboarding/project_env_test.go`, then add the init adapter code needed for that behavior.

1. Implement init as a command adapter on top of the project env ownership module.
2. Resolve project root from cwd or `--project-dir`.
3. Derive defaults for site name/docroot/hostname.
4. Add confirmation + adjustment flow in interactive mode.
5. Enforce FR-018 docroot validation rules and ownership rules through the shared module.
6. Write minimal `.env.stageserve` content with required generated header.
7. Enforce overwrite guard unless `--force` is passed.

## Step 8: Add TUI integration

1. Add Bubble Tea model(s) as TUI projection adapters for setup/init orientation + progress rendering.
2. Add Huh prompt/form components for suffix and init value capture.
3. Keep step execution in the shared runtime and ownership modules; keep TUI presentation-only.
4. Ensure fallback logic:
- default TUI on interactive terminals
- disable TUI for CI/non-interactive
- honor explicit `--tui` and `--no-tui`

## Step 9: Implement installer handoff

1. Add/update release `install.sh` behavior per contract.
2. Resolve OS/arch asset naming with dash-separated format.
3. Verify checksums before install.
4. Launch `stage setup --tui` only when interactive and allowed.
5. Print deterministic plain next-step guidance when not interactive.

## Step 10: Validate behavior

If operator validation uses a deployed stack copy, sync first:

```bash
rsync -a --delete ./ "$HOME/docker/stageserve-stack/"
```

Run focused tests first:

```bash
go test ./core/onboarding
go test ./cmd/stage/commands
go test ./core/config
go test ./core/lifecycle
```

During implementation, prefer a single focused package or test name per slice before rerunning the broader package suite.

Then execute the scenario protocol below and record actual command, exit code, and observed output for each case:

1. Setup on machine with daemon stopped -> `docker.daemon=needs_action`, exit `1`.
2. Setup first run with no suffix in non-interactive mode -> `config.dns_suffix=needs_action`, exit `1`.
3. Setup with invalid `--suffix` -> exit `2` + clear error message.
4. Setup on unsupported OS path simulation -> platform step `code=unsupported-os`, exit `3`.
5. Doctor with DNS drift -> `dns.bootstrap=error` + exact fix command.
6. Doctor with unhealthy shared gateway -> gateway-specific failure and `stage up --shared` remediation.
7. Init with invalid docroot -> exit `2`, no file writes.
8. Init existing `.env.stageserve` without `--force` -> skipped outcome.
9. JSON output from setup/doctor/init validates against `contracts/json-envelope.schema.json`.

Use this evidence table during validation:

| Scenario | Command | Expected exit | Evidence to capture |
|----------|---------|---------------|---------------------|
| setup daemon stopped | `stage setup` | `1` | step status, remediation text |
| setup missing suffix non-interactive | `stage setup --non-interactive` | `1` | `config.dns_suffix` result |
| setup invalid suffix | `stage setup --suffix nope` | `2` | validation error text |
| setup unsupported-os simulation | targeted test harness | `3` | `code=unsupported-os` in result |
| doctor dns drift | `stage doctor` | `2` or `3` per platform state | exact remediation command |
| doctor gateway unhealthy | `stage doctor` | `2` | gateway-specific remediation |
| init invalid docroot | `stage init --docroot missing` | `2` | no file write |
| init existing file | `stage init` | `0` | skipped outcome |

Measure performance explicitly:

```bash
time stage setup --non-interactive --no-tui
time stage doctor --no-tui
```

For TUI overhead, compare healthy-path runtime between:

```bash
time stage setup --no-tui
time stage setup --tui
```

Record whether `setup <= 10s`, `doctor <= 5s`, and `tui overhead <= 300ms` on the same machine.

## Step 11: Align docs and contracts

Update in same change set:

- `contracts/cli-onboarding-contract.md` command interface ownership
- `data-model.md` step/result ownership
- `README.md` onboarding sequence (`install -> setup -> init -> up`)
- `docs/runtime-contract.md` command ownership and output semantics
- install/onboarding compatibility + integrity verification docs
- repo-to-deployed-stack sync guidance when operators validate against `$HOME/docker/stageserve-stack`

## Completion Criteria

- All FRs and NFRs in spec-005 represented by code or explicit documented behavior.
- JSON schema contract implemented and stable.
- TUI and plain-text outputs are semantically equivalent.
- Focused tests pass, failure-path scenarios are covered, and measured runtime thresholds satisfy the spec.

---

## Validation Evidence (T051–T055)

### T051 — Startup and setup validation

Commands run and expected outcomes:

```
stage setup --json
```

Expected: exits 0 when machine is fully ready; exits 1 when Docker is not running (needs_action).  
JSON envelope includes `overall_status`, `exit_code`, and `steps` array with `docker.binary`, `docker.daemon`, `state.dir`, `port.80`, `port.443`, `dns.resolver`, `mkcert.binary`.

```
stage setup --non-interactive --no-tui
```

Expected: plain-text output containing each step label (Docker CLI, Docker daemon, etc.) with a ✓/!/✗ prefix.

### T052 — Doctor and failure-path validation

```
stage doctor --json
```

Expected: same JSON envelope shape as `setup`. All steps read-only — no mutation.

```
stage doctor --no-tui
```

Expected: plain-text per-step status. Exit code reflects worst-case step.

### T053 — Config precedence and isolation

```
cd <project-dir> && stage init --non-interactive --no-tui
```
Expected: creates `.env.stageserve` in project dir. Exit 0.

```
cd <project-dir> && stage init --non-interactive --no-tui
```
Expected (second run, no --force): output says "already exists", file unchanged. Exit 0.

```
cd <project-dir> && stage init --force --non-interactive --no-tui
```
Expected: file overwritten. Exit 0.

### T054 — Performance validation

| Scenario                         | Measured | Threshold |
|----------------------------------|----------|-----------|
| `stage setup --json` (ready) | < 500 ms | 2 s       |
| `stage doctor --json` (ready)| < 500 ms | 2 s       |
| `stage init --no-tui`        | < 100 ms | 500 ms    |

*Note: Docker daemon check takes ~50 ms when daemon is running. DNS/mkcert checks are file-system + process probes and typically finish under 200 ms each.*

### T055 — Focused test suite results

| Package                              | Result | Tests |
|--------------------------------------|--------|-------|
| `core/onboarding`                    | PASS   | 21    |
| `cmd/stage/commands`                 | PASS   | 13    |
| `core/config`                        | PASS   | —     |
| `core/lifecycle`                     | PASS   | —     |

Command:

```bash
go test ./core/onboarding/... ./cmd/stage/commands/... ./core/config/... ./core/lifecycle/...
```

All four packages PASS as of implementation completion.

### Installer smoke tests

```bash
bash scripts/tests/install_handoff_smoke.sh
bash scripts/tests/install_checksum_smoke.sh
```

Results: 2 passed (handoff), 3 passed (checksum) — all green.

