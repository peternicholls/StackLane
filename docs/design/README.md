# StageServe Terminal Design System

StageServe has one terminal design system for reports, guided flows, text fallback, and future Bubble Tea screens. This folder is the human-facing source of truth for that system.

In design terms, the terminal style guide is a graphic-design artifact: a viewable reference for visual identity, component usage, examples, and review rules. Bubble Tea, Lip Gloss, renderer helpers, and command adapters implement that reference; they do not replace it.

Use these docs when changing `stage doctor`, `stage setup`, onboarding, guided bare `stage`, or any other user-facing CLI interaction.

## Reading Order

1. [Terminal Visual Identity Style Guide](terminal-visual-style-guide.md)
   - Graphic-design intent, visual foundations, semantic colour, component vocabulary, examples, and Bubble Tea/Lip Gloss separation.
   - Runnable terminal edition: `go run ./docs/design/style-guide-tui`.
2. [Terminal Experience Style Guide](terminal-experience-style-guide.md)
   - Product intent, screen grammar, surface families, implementation seams, and review rules.
3. [Terminal Copy Style Guide](terminal-copy-style-guide.md)
   - Voice, vocabulary, banned first-level terms, confirmations, recovery copy, and action phrasing.
4. [StageServe Doctor Seed](stage-doctor-seed.md)
   - The current production report seed, taken from the real `stage doctor` command and projector code.
5. [Terminal Components And Prototypes](terminal-interface-prototypes.md)
   - Reusable components, representative screen sketches, and TUI/plain-text twins.
6. [Guided Flow Map](guided-flow-map.md)
   - Durable routing, situations, default actions, and planner boundaries for the guided shell.

## Design Scope

This design system covers three related surface families:

- Report surfaces such as `stage doctor`, `stage setup`, and readiness summaries.
- Guided interactive surfaces such as bare `stage`, project setup, run/stop/logs, and recovery flows.
- Utility and modal surfaces such as confirmations, More panels, detail views, advanced troubleshooting, and logs.

The design goal is one recognizable StageServe language across all three, not separate styles for reports and guided screens.

## Runnable Style Guide

The style guide is also available as a standalone Bubble Tea/Lip Gloss terminal app:

```bash
go run ./docs/design/style-guide-tui
```

This app is documentation. It demonstrates the identity, components, report patterns, guided patterns, assistance handoff, and design/code boundaries without touching Docker, project state, or production command paths.

For non-interactive review:

```bash
go run ./docs/design/style-guide-tui --plain
go run ./docs/design/style-guide-tui --list-sections
```

## Source Material

The design system is distilled from current implementation, active spec work, and visual-identity consolidation:

- The current `stage doctor` rendering is the best production visual anchor.
- The current `stage doctor` rendering is documented in [StageServe Doctor Seed](stage-doctor-seed.md).
- The visual identity rules are consolidated in [Terminal Visual Identity Style Guide](terminal-visual-style-guide.md).
- The prototype under `specs/007-harden-TUI-and-other-interactions/prototype/` is the design sandbox for guided flows.
- The detailed scenario notes and vocabulary live under `specs/007-harden-TUI-and-other-interactions/flow-diagrams/`.

The [Guided Flow Map](guided-flow-map.md) is the durable summary. The spec flow-diagram files remain the deeper working material when a screen or transition needs revision.

## Authority Model

The human-facing docs in this folder are the authored reference surface. The visual identity style guide is the top-level design artifact; the experience, copy, component, and flow documents elaborate it for specific decision layers.

The agent-facing instruction files under `.github/instructions/` mirror the same system for tools that can automatically load scoped guidance:

- `.github/instructions/terminal-experience-index.instructions.md`
- `.github/instructions/terminal-design-contract.instructions.md`
- `.github/instructions/terminal-markup-spec.instructions.md`
- `.github/instructions/terminal-copy-style.instructions.md`
- `.github/instructions/terminal-pattern-catalog.instructions.md`
- `.github/instructions/terminal-experience-goal.instructions.md`

When the design system changes, update the matching human docs, agent instructions, and reusable prototype examples in the same change so future work does not fork into parallel design languages.
