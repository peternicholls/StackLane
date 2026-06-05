# Quickstart: Guided Experience And Runtime Hardening

This document defines how to validate spec 011 in two slices.

## Prerequisites

- macOS shell environment with Docker Desktop available.
- StageServe repository checked out with spec 011 changes.
- Ability to run focused Go tests and manual terminal commands.

## A. Merge-Gate Validation

### A1. Automated Focused Checks

Run:

```bash
go test ./core/guidance ./cmd/stage/commands ./core/lifecycle ./infra/compose ./core/config
```

Expected:

- All suites pass.
- Tests cover missing runtime assets, root routing reachability, attach error handling, password-source behavior, and non-TTY or JSON safety.

### A2. Manual Runtime-Asset Preflight

1. Validate normal startup with provisioned assets:

```bash
stage up
```

2. Temporarily remove/rename one required runtime compose file and rerun `stage up`.

Expected:

- Failure occurs before compose startup attempt.
- Output is StageServe-native and includes a concrete remedy.

### A3. Manual Root Reachability

From a context with missing readiness/runtime prerequisites:

```bash
stage
```

Expected:

- Bare `stage` enters machine-not-ready style guidance.
- It does not default to ready-to-run actions.

### A4. Manual Root Routing Continuity

1. From a machine-ready project directory with no `.env.stageserve`, run:

```bash
stage
```

2. From a configured but stopped project, run:

```bash
stage
```

Expected:

- In a missing-config project directory, the default guided action is `Set up this directory as a project`.
- In a configured but stopped project, the default guided action is to run the project.

### A5. Manual Status, Teardown, And Failure Recovery

1. Exercise a status or inspection path from bare `stage` or direct status command.
2. Exercise a teardown path (`stage down` or equivalent guided stop).
3. Exercise at least one recovery path where a lifecycle or readiness blocker occurs.

Expected:

- Status or inspection remains available and understandable through StageServe-first language.
- Teardown remains explicit and safe.
- Recovery defaults to the least-invasive next action and surfaces `doctor`, `status`, and `logs` in that order.

### A6. Manual Attach/Password Safety

1. Exercise attach path with forced/induced registry read failure.
2. Verify root help/flags no longer include `--mysql-password` and existing env sources continue to work.

Expected:

- Attach fails explicitly with remedy.
- Password is not accepted as a CLI flag.

### A7. Non-TTY And JSON Safety

Run:

```bash
stage < /dev/null
stage setup --json
stage doctor --json
```

Expected:

- Non-interactive bare `stage` remains concise plain-text guidance.
- `setup --json` and `doctor --json` remain parseable and free of interactive recovery copy.

## B. Follow-up Validation

Run after day-2 and async UX phases are implemented.

### B1. Running-Project Day-2 Flows

With project running:

```bash
stage
```

Validate:

- non-destructive default action,
- open/status/logs flows,
- service-scoped logs/restart selection rules,
- advanced command fallback when selection is ambiguous.

### B2. Confirmation And Async Progress

Validate:

- elevated destructive confirmation surfaces,
- async spinner/progress for long-running lifecycle actions,
- first visible progress feedback appears within 250 ms of confirmation,
- cancellation behavior and safe return states,
- narrow-width terminal usability.

### B3. Follow-up Automated Checks

Run:

```bash
go test ./core/guidance ./cmd/stage/commands ./core/lifecycle ./infra/compose ./core/config
```

Add any package-targeted suites introduced by follow-up implementation.

## Known Validation Gaps

- If environment limitations prevent real Docker daemon end-to-end validation for any required scenario, record the exact gap and command attempted in spec closeout notes.