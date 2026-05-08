# StageServe Terminal Experience Style Guide

StageServe treats every CLI interaction as a designed terminal interface. In an interactive terminal that usually means Bubble Tea and Lip Gloss. In text fallback or JSON-adjacent flows it means the same hierarchy, semantics, and wording without the wrapper.

When the repo talks about the terminal design style guide in the graphic-design sense, start with [Terminal Visual Identity Style Guide](terminal-visual-style-guide.md). This experience guide is the companion interaction layer: it establishes product intent, surface families, and operating rules before specific copy or renderer details.

## Experience Intent

StageServe terminal UX should feel:

- **Clear:** every line helps the user understand state, importance, or action.
- **Competent:** the tool knows what happened, what matters now, and what should happen next.
- **Calm:** failures are direct without theatre, apology, or blame.
- **Practical:** remediation is exact, copy-pasteable, and close to the problem.
- **Consistent:** report views and guided views still feel like the same product.
- **Evolvable:** new flows can reuse the same grammar without freezing early implementation choices.

Avoid generic CLI habits: raw status dumps, decorative banners, internal names in first-level copy, mystery defaults, and vague "try again" guidance.

## Surface Families

StageServe has three related surface families.

### Report surfaces

Examples: `stage doctor`, readiness checks, passive setup reports.

- Report surfaces are evidence-first.
- They answer context, verdict, blockers, and exact remediation.
- Interactive terminals may offer assistance after the report is already useful as a standalone artifact.

### Guided surfaces

Examples: bare `stage`, project setup, run/stop/logs, recovery.

- Guided surfaces are action-first.
- They show the current truth, the values StageServe will use, and the safest likely next action.
- They should reduce the need to remember command names.

### Utility surfaces

Examples: confirmations, More panels, detail views, advanced and troubleshooting, logs.

- Utility surfaces support the main flow.
- They should never become the primary way a user discovers the normal path.

## Operating Rules

### Tool drives, user confirms

StageServe should own the workflow where there is one obvious safe next step. The user should choose only when there is a real difference in goal, not just a different implementation mechanism.

### Defaults stay visible

If StageServe has picked a value, the user can see it before committing. If pressing `enter` performs an action, the screen says what that action is.

This applies to:

- site name
- web folder
- domain suffix
- local URL
- target `.env.stageserve` path
- the highlighted default action

### Outcome first, details second

Every screen should answer four questions in order:

1. Where am I?
2. What is the state?
3. What matters?
4. What can I do next?

Prefer progressive disclosure. Start with the outcome, then the reason, then the action. Do not force users to infer meaning from a table of raw checks.

### Safe defaults only

The default action should be the lowest-risk likely goal.

- A running-project screen defaults to logs or inspect, not stop.
- A destructive or state-changing action confirms before it mutates anything.
- Rare destructive confirmations bias toward the non-destructive option as the default.

### Plain language first

User-goal labels belong in the main flow. Implementation vocabulary belongs behind `More…` or `Advanced and troubleshooting`.

### Text and TUI carry the same truth

Styled output may add emphasis, colour, and rhythm. It must not add or remove information compared with text fallback.

## Screen Grammar

StageServe screens should be composed from a small reusable grammar.

### Surface header

The header establishes StageServe identity and the current surface, such as `Doctor`, `Setup`, `Project`, or `Recovery`.

### Verdict line

The verdict is the first human sentence about the current state. It appears before diagnostics, values, or actions.

### Key facts or visible defaults

When the screen depends on values, show them in aligned rows with short source notes.

### Focus section

Each screen should have one dominant section. Examples:

- `Needs fixing`
- `All clear`
- `Setup steps`
- `Recovery steps`
- `What you can do`

Do not stack multiple equally-important sections if one of them is clearly the reason the screen exists.

### Secondary surfaces

Direct commands, advanced detail, hidden file paths, and power-user explanations belong in a `More…` path or equivalent secondary surface.

### Footer help

Footer hints should show only the keys that matter on that screen. Essential behaviour should not depend on undiscoverable shortcuts.

## Bubble Tea And Go Boundaries

StageServe should keep design concerns separate from functional implementation.

### Planner and domain layer

This layer decides the user situation and produces a normalized screen plan. It should not know about cursor movement, Lip Gloss styles, or key bindings.

### Bubble Tea view model

This layer owns transient UI state: focus, cursor position, modal state, edit drafts, pending confirmation, and in-flight actions.

### Renderer and styling layer

This layer owns layout, component helpers, section headers, spacing, glyphs, and semantic colour. It should be reusable across report views and guided views.

### Action layer

This layer executes lifecycle, setup, status, log, or recovery actions and then asks the planner to re-detect. Bubble Tea should not hard-code the business truth of whether a project is running or out of sync.

The current `stage doctor` report rendering is the best visual anchor for report anatomy. Future guided screens should converge on that language rather than inventing a separate one.

## Report-To-Assistance Rule

When an interactive report finds problems, the report still comes first.

- Power users can copy the remediation and leave.
- Interactive users can opt into guided help.
- Guided help should take one blocker at a time and explain what StageServe can do, what it cannot do, and what needs confirmation.

Assistance is a handoff, not a replacement for a useful report.

## Review Checklist

Before approving a terminal interaction, check:

- The first meaningful line establishes context.
- The verdict is human-readable and appears before detailed diagnostics.
- The screen shows the values StageServe will actually use.
- The default action is visible and low-risk.
- Any command shown is exact and copy-pasteable.
- First-level copy avoids internal vocabulary.
- Direct commands and advanced material are secondary, not primary.
- TUI and text fallback preserve the same information.
- Success gets quieter as there is less to do.
- A reusable pattern is reflected in the component catalog or flow map.
