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
STAGESERVE_NO_TUI=1 stage
```

Expected:

- Text fallback is shown.
- No interactive UI is opened.

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

### 5. Configured stopped project

```bash
stage
```

Expected:

- TUI identifies project as configured and stopped.
- Primary action is to run the project.
- Direct command equivalent is visible: `stage up`.

### 6. Running project

```bash
stage
```

Expected:

- TUI offers status/logs/down/doctor.
- Stop action confirms before running.
- Stop preserves data and uses `stage down` semantics.

### 7. Failure path

Simulate a missing Docker daemon, DNS drift, invalid `.env.stageserve`, or bootstrap failure.

Expected:

- TUI shows the problem.
- It provides a StageServe recovery path first.
- Advanced implementation details are available only behind an advanced/troubleshooting action.

### 8. JSON remains pure

```bash
stage setup --json > /tmp/stage-setup.json
jq . /tmp/stage-setup.json >/dev/null

stage doctor --json > /tmp/stage-doctor.json
jq . /tmp/stage-doctor.json >/dev/null
```

Expected:

- stdout is valid JSON.
- no styled text or next-step prose is mixed into JSON.

### 9. Direct commands remain stable

```bash
stage up --help
stage status --help
stage logs --help
stage down --help
```

Expected:

- Existing behavior and help semantics remain stable.

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
- Advanced/troubleshooting sections still contain enough implementation detail for power users.
- `.env.stageserve` is the only normal user-editable StageServe config file.
