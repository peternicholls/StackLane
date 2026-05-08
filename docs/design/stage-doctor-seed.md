# StageServe Doctor Seed

This document snapshots the current production `stage doctor` report as implemented today. It is the seed example for StageServe report surfaces.

Use it when deciding what should stay stable, what can evolve, and how future guided or Bubble Tea work should stay grounded in the product's real shipped output.

## Owning Code Path

The current doctor surface is owned by these files:

- [cmd/stage/commands/doctor.go](../../cmd/stage/commands/doctor.go)
  The command adapter. It resolves output mode, assembles the ordered readiness checks, and always renders doctor in detailed mode.
- [core/onboarding/projection_tui.go](../../core/onboarding/projection_tui.go)
  The Lip Gloss styled report projector used when doctor runs in TTY/TUI mode.
- [core/onboarding/projection_text.go](../../core/onboarding/projection_text.go)
  The plain-text projector used for `--no-tui`, non-interactive mode, or non-TTY output.
- [core/onboarding/projection_shared.go](../../core/onboarding/projection_shared.go)
  Shared formatting rules: section headers, check descriptions, compact status words, remediation cleanup, and footer follow-up command.

## What Doctor Actually Is Today

The current doctor UI is not a full Bubble Tea screen. It is a static detailed report projected from a shared `CommandResult`.

That matters because this seed is defined by:

- ordered checks
- deterministic verdict calculation
- TUI/text parity
- exact remediation rendering
- semantic colour and weight

It is not currently defined by cursor movement, menus, modals, or live interaction.

## Ordered Check Set

`doctor.go` renders these checks in this order:

1. Docker CLI
2. Docker daemon
3. State directory
4. Port 80
5. Port 443
6. DNS resolver
7. mkcert local CA

This ordered list is part of the seed. Problems appear in that order, then passing items are summarized afterward.

## Current TUI Anatomy

Representative non-ready output:

```text
  ◆  StageServe Doctor
  ──────────────────────────────────────

  ✗  Not ready — 2 of 7 checks need attention.

── Needs fixing ────────────────────────

  1  Docker daemon
     The Docker daemon must be running before any container can start.

      Docker daemon is not reachable
      To fix:  Start Docker Desktop or run: sudo systemctl start docker

  2  Port 443
     Port 443 must be free for the local HTTPS gateway to bind to it.

      port 443 is already in use
     To fix:  sudo lsof -nP -iTCP:443 -sTCP:LISTEN

── All clear ───────────────────────────

  ✓  State directory    exists
  ✓  mkcert local CA    installed

  ──────────────────────────────────────
  Fix the issues above, then run: stage doctor
```

Representative ready output:

```text
  ◆  StageServe Doctor
  ──────────────────────────────────────

  ✓  All 7 checks passed — your machine is ready.

── Checks passed ───────────────────────

  ✓  Docker CLI         docker found at /usr/local/bin/docker
  ✓  Docker daemon      Docker daemon running (server 27.5.1)
  ✓  State directory    exists
  ✓  Port 80            available
  ✓  Port 443           available
  ✓  DNS resolver       configured
  ✓  mkcert local CA    installed

  ──────────────────────────────────────
  Your machine is ready. Run: stage up
```

These sketches reflect the current renderer structure and wording patterns. Individual messages vary with machine state and platform, especially DNS and port-owner detection.

## Current Text Fallback Anatomy

Plain text keeps the same ordering, wording, and sections, but removes colour and the header glyph:

```text
  StageServe Doctor
  ──────────────────────────────────────

  ✗  Not ready — 2 of 7 checks need attention.

── Needs fixing ────────────────────────
...
── All clear ───────────────────────────
...
  ──────────────────────────────────────
  Fix the issues above, then run: stage doctor
```

That parity is part of the seed. TUI adds styling, not extra information.

## Formatting Rules Taken From Code

### Header

- One blank line before the report.
- Two-space left inset.
- `◆` header glyph in cyan.
- `StageServe Doctor` title in bright white bold.
- One full-width dim divider directly under the title.

### Verdict

- The verdict always appears before detail sections.
- Ready verdict is green and bold.
- Non-ready verdict is red and bold, even when the underlying blockers are `needs_action` rather than `error`.
- Current exact verdict strings are:
  - `All N checks passed — your machine is ready.`
  - `Not ready — N of M checks need attention.`

### Needs Fixing Section

- Section title is `Needs fixing`.
- The section header uses a dim rule with a coloured bold title.
- Each issue is numbered starting at `1`.
- Number colour is yellow for `needs_action` and red for `error`.
- The check label is bright white bold.
- The check description is light grey and italic in TUI mode.
- The observed problem message is dark grey.
- The remediation label is bold `To fix:` followed by a bright cyan bold remediation string.
- Current code often renders an exact command here, but not always. Some checks still emit instruction-led remediation text.
- Blank lines separate title, description, problem, and remediation so each issue scans as a compact card without borders.

### All Clear Or Checks Passed Section

- When there are blockers, the passing section title is `All clear`.
- When everything passes, the title changes to `Checks passed`.
- Each passing row uses:
  - a green `✓`
  - a padded bold label column
  - a dim compact status word or evidence string

This is intentionally quieter than the failing section.

### Footer

- The footer repeats the full dim divider.
- When not ready, the footer gives one next action: `Fix the issues above, then run: stage doctor`.
- When ready, the footer hands off to the next lifecycle step: `Your machine is ready. Run: stage up`.

## Colour And Style Choices In Current Code

These are the actual Lip Gloss style variables used by the current doctor TUI projector.

| Style | ANSI | Current use |
| --- | --- | --- |
| `styleCyan` | `6` | Header glyph `◆` |
| `styleWhite` | `15` + bold | Report title and issue labels |
| `styleDim` | `8` | Divider lines, problem messages, compact passed-state text |
| `styleMuted` | `7` | Why-this-matters descriptions beneath issue labels |
| `styleReady` | `2` | Ready verdict, green checks, ready section title |
| `styleNeedsAction` | `3` | Warning semantics, warning issue numbers, `Needs fixing` title |
| `styleError` | `1` | Blocking verdict, error issue numbers |
| `styleBrightCyan` | `14` + bold | Copy-pasteable remediation commands |
| `styleBold` | bold only | Labels such as `To fix:` and padded passed labels |

The current seed uses colour semantically, not decoratively. Every colour has a specific job.

## Language Choices In Current Doctor

### What the copy does well

- Leads with a clear context label: `StageServe Doctor`.
- Gives a machine verdict before details.
- Orders blockers before passing checks.
- Keeps remediation adjacent to the blocker.
- Uses concise section names: `Needs fixing`, `All clear`, `Checks passed`.

### What the copy still exposes

The current seed is useful, but it still carries some implementation-heavy language in the descriptive layer:

- `Docker daemon`
- `local HTTP gateway`
- `project registry`
- `*.test`
- `dnsmasq`
- `mkcert`

The current seed also mixes remediation styles:

- command-only fixes such as `sudo lsof -nP -iTCP:443 -sTCP:LISTEN`
- command handoffs such as `stage setup`
- instruction-led fixes such as `Start Docker Desktop or run: sudo systemctl start docker`
- install guidance such as `Install mkcert: brew install mkcert && mkcert -install`

That is part of today's production reality. It should be treated as the baseline to evolve from, not as the final ceiling for user-language quality.

## Seed Rules To Preserve

When evolving report surfaces from this seed, preserve these traits unless there is a deliberate design reason to change them:

- Report surfaces stay report-first, not menu-first.
- The verdict comes before detail.
- Problems appear before confirmations.
- Remediation stays on a single line immediately below the problem.
- TUI and text fallback carry the same information.
- Passing states are quieter than failing states.
- The footer gives one obvious next action.

## Safe Evolution From This Seed

It is safe to improve language, simplify implementation-heavy descriptions, or refine section wording, as long as the following remain stable:

- the report anatomy
- the problem-first ordering
- the semantic colour model
- the one-line remediation slot, even if the remediation wording improves later
- TUI/text parity

This seed should be the starting point for future report surfaces, not a frozen template.