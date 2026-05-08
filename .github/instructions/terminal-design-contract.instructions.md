---
applyTo: "core/onboarding/**,cmd/stage/commands/**"
---

# StageServe Terminal Design Contract

StageServe terminal output is a product interface, not incidental console text. It is the StageServe sub-brand most users will see most often, so every command should feel calm, intentional, useful, and recognisably part of the same product.

This contract defines design intent. It is platform-agnostic: humans, Codex, and other agents should use it before reasoning about ANSI codes, helper functions, or exact markup.

When the work asks for a design style guide, interpret that in the graphic-design sense. The human-facing style guide is `docs/design/terminal-visual-style-guide.md`: a viewable reference for visual identity, semantic colour, component usage, examples, and review rules. This contract is the scoped, agent-loadable policy layer that mirrors that guide for command and onboarding code.

Bubble Tea, Lip Gloss, renderer helpers, and command adapters implement the design system. They are not the design system themselves.

Use `.github/instructions/terminal-copy-style.instructions.md` for concrete voice, vocabulary, label, verdict, and remediation rules after this contract establishes the experience.

## Identity

StageServe terminal UX should feel:

- **Clear:** every line helps the user understand state, importance, or action.
- **Competent:** the command knows what happened, what matters, and what should happen next.
- **Calm:** failures are direct without being dramatic, apologetic, or noisy.
- **Practical:** remediation is concrete, copy-pasteable, and close to the problem.
- **Consistent:** repeated concepts use repeated structure, language, and colour semantics.
- **Evolvable:** the system supports new workflows without freezing early implementation choices.

Avoid generic CLI habits: status dumps, internal names, decorative banners, raw enum values, unexplained counters, and vague "try again" messages.

## Surface Families

StageServe has three related surface families:

- **Report surfaces:** diagnosis and evidence-first views such as `stage doctor`, readiness checks, and passive setup reports.
- **Guided surfaces:** action-first views such as bare `stage`, project setup, run/stop/logs, and recovery.
- **Utility surfaces:** confirmations, More panels, detail views, advanced and troubleshooting, and logs.

These surfaces should feel like one product system, not separate visual languages.

## Design Layers

Think about every terminal interaction in three complementary layers.

**UI design:** visual hierarchy, spacing, colour semantics, icons, grouping, scanability, and rhythm.

**UX design:** user journey, information order, default path, blocked path, next actions, follow-up commands, and whether the command should continue automatically.

**Language:** labels, explanations, problem statements, remediation copy, verdicts, and the level of detail needed for the user's current state.

Do not solve one layer while ignoring the others. A visually tidy output that leaves the user unsure what to do is not successful. A correct remediation buried in noisy prose is not successful.

## Interaction Model

Each command should answer four questions in order:

1. **Where am I?** Name the command surface or product area.
2. **What is the state?** Give a human verdict before details.
3. **What matters?** Surface blockers, warnings, and confirmations in priority order.
4. **What can I do next?** Provide the next useful action when there is one.

Prefer progressive disclosure. Start with the outcome, then the reason, then the action. Do not force users to infer status from a table of raw checks.

## Operating Rules

### Tool drives, user confirms

StageServe should own the workflow where there is one obvious safe next step. The user should choose only when there is a real difference in goal, not just a different implementation detail.

### Defaults stay visible

If StageServe has picked a value, the user can see it before committing. If pressing `enter` performs an action, the screen says what that action is.

This applies to URLs, file paths, site names, web folders, suffixes, and highlighted default actions.

### Safe defaults only

The default action should be the lowest-risk likely goal.

- Running-project screens default to inspect or logs, not stop.
- Destructive or mutating actions always confirm.
- Rare destructive confirmations may default to the non-destructive option.

### Plain language first

User-goal labels belong in the primary flow. Docker, compose, gateway, runtime, registry, state record, `attach`, and `detach` are advanced-only vocabulary.

### Text and TUI carry the same truth

Styled output may add emphasis and rhythm. It must not add or remove information compared with plain text.

## Information Hierarchy

Use whitespace, order, and weight to create hierarchy. Do not rely on boxes, repeated dividers, or decorative frames.

Detailed output should support reading top to bottom. Compact output should support quick scanning.

When there are problems, lead with problems. Passing checks can appear afterward as reassurance, not as a barrier before the user sees what needs attention.

When everything passes, the output should become quieter. Do not print ceremonial success text or a long proof list unless the command context needs it.

## Screen Grammar

StageServe screens should be composed from a small reusable grammar:

- **Surface header:** StageServe identity plus the current surface such as `Doctor`, `Setup`, `Project`, or `Recovery`.
- **Verdict line:** the first human sentence about the current state.
- **Key facts or visible defaults:** aligned rows for values StageServe will use.
- **Focus section:** one dominant section such as `Needs fixing`, `All clear`, `Setup steps`, `Recovery steps`, or `What you can do`.
- **Secondary surfaces:** `More…`, details, advanced and troubleshooting, logs, or confirmation.
- **Footer help:** only the keys relevant to the current screen.

Do not stack multiple equally-important sections if one of them is clearly the reason the screen exists.

## Action Design

Commands shown to users must be exact shell invocations. Do not wrap commands in explanatory prose, placeholders, or half-commands.

When a single next action is clearly best, provide that action directly.

When multiple next actions are valid, separate the decision from the commands. Explain the choice briefly, then show exact commands.

When a command can safely perform the next step automatically, prefer doing the work over asking the user to manually stitch steps together. When automation would be risky, explain the risk in human terms and stop at a clear command or confirmation.

## Copy Rules

Write like a capable product surface, not a logger.

- Use natural human language, not internal state names.
- Keep explanations factual and short.
- Prefer present-tense descriptions of state.
- Avoid "please", apologies, cheerleading, jokes, and theatrical failure language.
- Avoid "Note:" and "you must" in check descriptions.
- Use "To fix:" or "Next:" for remediation labels consistently.
- End explanatory sentences with punctuation.
- Do not print status tags such as `needs_action`, `StatusReady`, or `[docker.binary]`.

The user should never need to know the internal data model to understand the output.

## Bubble Tea Boundary

Keep design logic separate from functional implementation.

- The planner or domain layer decides the user situation and produces a normalized screen plan.
- Bubble Tea owns cursor movement, modal state, edit drafts, focus, and in-flight interaction state.
- Render helpers own section layout, spacing, glyphs, Lip Gloss styles, and TUI/text parity.
- Command execution runs lifecycle, setup, status, log, or recovery work and then asks the planner to re-detect.

Bubble Tea should not decide whether a project is running, out of sync, or needs recovery. It should present that truth coherently.

## Colour And Meaning

Colour carries semantic meaning only. It is not decoration, brand garnish, or a way to make output feel more designed.

Each colour must have one clear purpose. If a new state needs a new colour, update the markup spec and add an example before using it.

Plain text output must carry the same information as styled output. Styled output may add emphasis and colour; it must not add or remove content.

## Report-To-Assistance Rule

When an interactive report finds issues, the passive report still comes first.

- Power users can copy the remediation and leave.
- Interactive users can opt into guided help.
- Guided help should take one blocker at a time and explain what StageServe can do, what it cannot do, and what needs confirmation.

Assistance is a handoff, not a replacement for a useful report.

## Guardrails

Protect against these forms of drift:

- **Brand drift:** output becomes generic, noisy, decorative, or mechanically "CLI-like".
- **UX drift:** commands dump state instead of guiding users through a useful flow.
- **Copy drift:** wording becomes verbose, apologetic, robotic, cute, or over-explanatory.
- **Markup drift:** styling changes break semantic colour, hierarchy, parity, or copy-pasteability.
- **Pattern drift:** old examples continue shaping new work after the product has moved on.

When changing an established pattern, name the design reason in the change description and update the pattern catalog if the change should guide future work.

## Design Review Checklist

Before approving terminal output, check:

- The first meaningful line establishes context.
- The verdict is human-readable and appears before detailed diagnostics.
- The screen shows the values StageServe will actually use.
- The most important user action is obvious.
- The default action is visible and low-risk.
- Any command shown is exact and copy-pasteable.
- The same information exists in TUI and plain-text output.
- Colour has semantic purpose and follows the markup spec.
- Copy avoids internal status names and implementation details.
- Direct commands and advanced material are secondary, not primary.
- Passing and failing states are both intentionally designed.
- The output gets quieter when there is less to do.
- A new or changed pattern has an example when it should guide future work.
