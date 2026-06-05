# Contract: Guided Runtime Hardening And Follow-up Expansion

## Purpose

Define the behavioral contract for spec 011 across merge-gate hardening and follow-up guided UX expansion without changing the spec 007 simple-first product model.

## 1. Entry And Routing Contract

- Bare `stage` remains guided entry in interactive terminals.
- Bare `stage` in non-interactive contexts remains plain-text guidance.
- `--notui` and `--cli` remain opt-out controls.
- Direct subcommands remain authoritative automation/power-user path.
- Bare `stage` routing remains deterministic across machine-not-ready, project-missing-config, ready-to-run, running-project, and recovery-needed states.
- In a plausible project directory without `.env.stageserve`, the highlighted default is `Set up this directory as a project`; machine setup is not the default unless machine prerequisites are missing.

## 2. Runtime Asset Contract (Merge Gate)

- Required runtime files must be validated before compose startup.
- At minimum:
  - `cfg.SharedFile`
  - `cfg.StackFile`
- Missing files must produce StageServe-native remedies (not raw compose path errors).
- Installer path for 011 must provision required runtime assets via bundled artifact.

## 3. Lifecycle Error Contract (Merge Gate)

- Attach must fail explicitly on registry read failures.
- Touched lifecycle error wrappers must include non-empty remedies.
- DownAll partial failures must preserve explicit operator context (what succeeded, what failed, what to do next).
- Up shared-route write behavior must either:
  - refresh registry before writing routes, or
  - explicitly document and enforce single-instance assumptions.

## 4. Config Safety Contract (Merge Gate)

- Root CLI password flag `--mysql-password` is removed.
- `MYSQL_PASSWORD` remains sourced only from:
  - shell environment,
  - project `.env.stageserve`,
  - stack `.env.stageserve`.
- No replacement password prompt, secret file, or new persistent config surface is introduced in 011.

## 5. Presentation Contract

- Guided TUI, text fallback, and direct command errors share one semantic severity model.
- `NO_COLOR=1` disables ANSI styling while preserving readable hierarchy and actionability.
- Failure output must include an explicit next action in StageServe language.

## 6. Recovery Contract (Follow-up)

- Interactive direct-command lifecycle failures should route users into guided recovery surfaces where applicable.
- Non-interactive failures must provide equivalent text guidance and direct command equivalents.
- For unknown or ambiguous failure states, the highlighted recovery path is least-invasive first: `stage doctor`, then `stage status`, then `stage logs`, with direct command equivalents visible in the footer or advanced fallback.
- Failure and recovery surfaces use StageServe task language as the primary label; raw command names remain secondary equivalents.

## 7. Running-Project Day-2 Contract (Follow-up)

- Running-project guided surface exposes at least:
  - open browser,
  - status inspection,
  - logs,
  - stop with confirmation,
  - advanced command fallback.
- Service-scoped logs/restart actions follow deterministic selection:
  - exactly one eligible service: auto-select,
  - multiple eligible services: explicit selector,
  - missing/ambiguous metadata: advanced fallback.

## 8. Confirmation And Async Progress Contract (Follow-up)

- Destructive actions use elevated confirmation surfaces.
- Long-running lifecycle actions execute asynchronously and show visible progress/spinner feedback.
- Cancel/quit semantics must preserve current cancellation/rollback safety.

## 9. Validation Contract

Merge-gate completion requires:

- runtime asset provisioning and missing-asset remedies validated,
- machine-not-ready reachability validated,
- root missing-config and ready-to-run routing validated,
- status/inspection path validated,
- teardown path validated,
- lifecycle attach/down-all safety paths validated,
- at least one failure/recovery path validated,
- password source contract validated,
- non-TTY fallback and JSON output safety validated,
- focused automated and manual terminal checks recorded.

Follow-up completion additionally requires:

- running-project day-2 action validation,
- confirmation elevation validation,
- async progress/cancel validation,
- docs/help parity for expanded behaviors.