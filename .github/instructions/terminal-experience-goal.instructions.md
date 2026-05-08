---
applyTo: "core/onboarding/**,cmd/stage/commands/**"
---

# Goal: StageServe Terminal Experience

Give this goal to Codex or another agent when asking for long-form work on any StageServe CLI interaction. StageServe treats terminal output as a TUI surface even when users only see plain terminal text.

This goal is for humans, Codex, and other agent platforms. It is an operating prompt for doing the work, not the documentation entrypoint and not the full design system.

## Source Documents

Read these in order:

1. `.github/instructions/terminal-experience-index.instructions.md`
2. `docs/design/terminal-visual-style-guide.md`
3. `.github/instructions/terminal-design-contract.instructions.md`
4. `.github/instructions/terminal-markup-spec.instructions.md`
5. `.github/instructions/terminal-copy-style.instructions.md`
6. `.github/instructions/terminal-pattern-catalog.instructions.md`

When the work changes bare `stage`, planner situations, project setup routing, run/stop flow, recovery, or report-to-assistance handoffs, also read:

- `docs/design/guided-flow-map.md`
- `specs/007-harden-TUI-and-other-interactions/flow-diagrams/README.md`

Use the visual identity style guide to understand the graphic-design intent, visual foundations, semantic colour, components, and separation of concerns. Use the design contract to decide what the experience should be. Use the markup spec only after the experience is clear. Use the copy style guide for labels, verdicts, explanations, remediation, and next actions. Use the pattern catalog for precedent and examples.

## Goal

Produce terminal interactions that are clear, consistent, useful, copy-pasteable, and recognisably StageServe.

The output must consider:

- UI design: hierarchy, rhythm, colour semantics, grouping, and scanability.
- UX design: flow, information order, default path, blockers, next actions, and automation boundaries.
- Language: labels, explanations, verdicts, remediation, and clarity.

## Workflow

1. Identify the interaction: command, user journey, current state, target outcome, and failure modes.
2. Separate planner truth from UI state: define the situation and screen-plan contract before deciding Bubble Tea interactions.
3. Decide the task type: design new output, review existing output, or revise an established pattern.
4. Read the visual identity style guide and design contract before proposing output.
5. State whether the change is visual identity, interaction pattern, copy, renderer implementation, or functional behavior.
6. Check the copy style guide before writing or revising labels, verdicts, descriptions, warnings, or remediation.
7. Check the pattern catalog for a matching or adjacent precedent.
8. Draft the interaction in plain text first, with the information hierarchy visible without colour.
9. Map the design to the markup spec: colour, weight, glyphs, spacing, helper functions, and text/TUI parity.
10. Implement or recommend changes using existing renderer helpers and local patterns.
11. Add or update visual guide, catalog examples, or agent instructions when a new pattern should guide future work.
12. Verify with tests, snapshots, or targeted command output checks where possible.
13. Report changed files, design choices, verification performed, and remaining risks.

## Review Mode

When reviewing existing terminal output, lead with findings:

- Does the verdict appear early and use human language?
- Does the screen match the visual identity guide's hierarchy, component roles, and semantic colour rules?
- Are blockers ordered before passing details?
- Is each remediation exact and copy-pasteable?
- Does the output avoid internal status names and implementation details?
- Does colour carry semantic meaning only?
- Do TUI and plain-text modes preserve the same information?
- Is the success state quieter than the failure state?
- Would a user know the next action without reading source code or docs?
- Does the wording follow the copy style guide's voice, vocabulary, and remediation rules?
- Does the pattern catalog need a new or revised example?

## Revision Mode

When revising output:

- Preserve established StageServe semantics unless there is a clear design reason to change them.
- Prefer deleting noisy lines over adding explanatory prose.
- Prefer existing helpers over one-off rendering.
- Keep TUI and plain-text paths aligned.
- Update docs and examples when the change alters a reusable pattern.
- Run `go test ./core/onboarding ./cmd/stage/commands` when changing onboarding or command output behavior.

## Guardrails

Do not:

- Print internal enum names, raw check IDs, or status tags as user-facing copy.
- Add decorative colour, borders, banners, or boxes.
- Add a command that is not an exact shell invocation.
- Add TUI-only information that plain text does not contain.
- Keep old output patterns alive after the design contract has moved past them.
- Treat examples as rigid templates when the user journey needs a better pattern.

Do:

- Start from user state and next action.
- Make the output readable without colour.
- Use styling to reinforce meaning.
- Keep copy calm, short, and specific.
- Make evolution explicit through design reasons and catalog updates.
