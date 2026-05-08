---
applyTo: "core/onboarding/**,cmd/stage/commands/**"
---

# StageServe Terminal Experience Index

StageServe treats every CLI interaction as a designed terminal interface. Users may only see terminal output, but that output is still one of the product's primary surfaces.

This index points humans and agents to the right source of truth before designing, reviewing, or changing terminal output.

Human-facing reference artifacts live in `docs/design/`. Keep those docs and these agent-facing instruction files aligned when the design system changes.

When a task says `style guide`, `design style guide`, `visual identity`, or `terminal brand guide`, interpret that in the graphic-design sense: the central, viewable reference is `docs/design/terminal-visual-style-guide.md`. Bubble Tea, Lip Gloss, and renderer helpers implement that reference; they are not the source of design truth.

## Reading order

1. **Visual identity style guide:** `docs/design/terminal-visual-style-guide.md`
   - Defines the graphic-design intent, visual foundations, semantic colour, component vocabulary, examples, and Bubble Tea/Lip Gloss separation.
   - Read this first when work is about TUI direction, style consolidation, visual identity, or component design.
2. **Design contract:** `.github/instructions/terminal-design-contract.instructions.md`
   - Defines the StageServe terminal identity, UX principles, language rules, and evolution guardrails.
   - Read this after the visual guide for any new interaction, UX change, or copy change.
3. **Markup spec:** `.github/instructions/terminal-markup-spec.instructions.md`
   - Maps the design contract onto concrete terminal rendering: ANSI colour, layout, glyphs, helpers, and text/TUI parity.
   - Read this before editing renderers, projectors, or command output helpers.
4. **Copy style guide:** `.github/instructions/terminal-copy-style.instructions.md`
   - Defines voice, vocabulary, verdicts, labels, remediation copy, and review checks for human- and agent-written terminal copy.
   - Read this before writing labels, descriptions, status messages, warnings, empty states, or next actions.
5. **Pattern catalog:** `.github/instructions/terminal-pattern-catalog.instructions.md`
   - Provides examples, anti-examples, and reusable interaction patterns for current and anticipated StageServe flows.
   - Read this when designing a new output shape or reviewing whether an interaction feels consistent.
6. **Goal prompt:** `.github/instructions/terminal-experience-goal.instructions.md`
   - Long-form operating prompt for agents asked to design, review, or revise terminal interactions.
   - Give this to Codex when the work itself is to improve a StageServe terminal experience.

## Guided Flow Reference

When changing bare `stage`, planner situations, project setup flow, run/stop flow, recovery flow, or report-to-assistance handoffs, also read:

- `docs/design/guided-flow-map.md`
- `specs/007-harden-TUI-and-other-interactions/flow-diagrams/README.md`

Those files define the durable interaction model that sits above any one Bubble Tea implementation.

## Authority model

The visual identity style guide is the top-level human design artifact. It governs the graphic-design meaning of the terminal identity: visual foundations, component roles, examples, and what is fixed versus flexible.

The design contract is the agent-facing normative layer. It mirrors the visual guide for scoped code work and governs product intent, user experience, tone, and brand.

The markup spec is subordinate. It can evolve as terminal libraries, helper functions, or rendering needs change, but it must still satisfy the visual identity style guide and design contract.

The copy style guide is subordinate to the visual identity style guide and design contract, and peer to the markup spec. It governs words while the markup spec governs rendering.

The pattern catalog is precedent. It should grow as StageServe grows, but examples are not frozen templates. Prefer the intent behind a pattern over copying its surface exactly.

The goal prompt is operational. It tells an agent how to apply the other files during long-form work; it does not override them.

## Evolution rule

StageServe's terminal identity is allowed to evolve. When it does, make the design reason explicit, update the visual identity guide first when the identity changes, then update the contract, markup spec, copy guide, or examples needed for future contributors to understand the change.

Do not preserve old output patterns only because they existed before. Keep the catalog current with the product StageServe is becoming.
