# StageServe Terminal Visual Identity Style Guide

This is the StageServe terminal style guide in the graphic-design sense: a central, viewable reference that defines the terminal identity, visual foundations, component vocabulary, usage rules, and example applications.

Use this guide when someone asks for the StageServe TUI style guide, terminal brand guide, visual identity guide, component guide, or design system reference.

The guide also has a runnable terminal edition:

```bash
go run ./docs/design/style-guide-tui
```

That app is a documentation surface. It demonstrates the identity and components in Bubble Tea/Lip Gloss without using the spec prototype or production command paths.

## What This Guide Is

This guide is the design source for how StageServe terminal surfaces should look, scan, and feel before a developer decides which Bubble Tea model, Lip Gloss style, renderer helper, or command path will produce them.

It plays the same role that a brand style guide or web design style guide would play in a visual design workflow:

- It establishes product intent and personality.
- It names the visual primitives: colour, weight, spacing, glyphs, and rhythm.
- It defines reusable components and how they are used.
- It shows examples and anti-patterns.
- It gives reviewers a checklist for deciding whether a new screen belongs to the same identity.

It is not a task list, prototype, implementation plan, or screenshot freeze. Examples show intent and relationships; they are not permission to preserve old output after the design has moved on.

## Design Intent

StageServe should feel like a capable local-development tool that is quietly in control.

The terminal identity should be:

- **Clear:** the first scan tells the user where they are, what state they are in, and what matters.
- **Competent:** StageServe appears to understand the workflow rather than dumping implementation state.
- **Calm:** failures are direct, useful, and untheatrical.
- **Practical:** every blocker has exact remediation or a safe guided next step.
- **Recognisable:** reports, guided screens, confirmations, and fallbacks feel like one product.
- **Evolvable:** the system can absorb better patterns without being trapped by earlier prototypes.

The best current visual anchor is `stage doctor`. It is the seed for report anatomy, semantic colour, section rhythm, and quiet confidence. It is not the ceiling for the final guided TUI.

## Source Hierarchy

Use the design artifacts in this order:

1. **This visual identity style guide** defines the graphic-design intent and identity rules.
2. [Terminal Experience Style Guide](terminal-experience-style-guide.md) defines product interaction principles, surface families, and operating rules.
3. [Terminal Copy Style Guide](terminal-copy-style-guide.md) defines labels, voice, remediation, warnings, confirmations, and vocabulary.
4. [StageServe Doctor Seed](stage-doctor-seed.md) documents the current production report surface.
5. [Terminal Components And Prototypes](terminal-interface-prototypes.md) shows reusable component sketches and application examples.
6. [Guided Flow Map](guided-flow-map.md) defines durable guided-routing situations and default actions.
7. `.github/instructions/terminal-*.instructions.md` mirrors these ideas for agents working in scoped code paths.

Spec files and prototypes can supply context, but they do not outrank the current design guide. If an older spec or prototype conflicts with this guide, update the older artifact or treat it as historical context.

## Visual Foundations

### Medium

StageServe is a terminal product surface. The identity must work in:

- Bubble Tea interactive screens.
- Lip Gloss styled reports.
- plain text fallback.
- non-colour terminals.
- copied output in issues, docs, and support conversations.

The plain-text version must carry the same information as the styled version. Colour and weight clarify meaning; they do not create meaning that disappears without styling.

### Composition

Use whitespace, order, and alignment before adding decoration.

- Start with a surface header.
- Put the human verdict before diagnostics or choices.
- Show values before actions that commit to those values.
- Give each screen one dominant focus section.
- Keep advanced material in `More…`, details, or troubleshooting surfaces.
- Use dividers to create rhythm, not to build decorative frames.

The current report rhythm uses a small left inset, a compact divider, clear section titles, and blank lines around the verdict and focus sections. Guided screens should inherit that rhythm while staying action-first.

### Colour System

StageServe colour is semantic, not decorative. New colours require a design reason, a token name, and an example.

| Role | Lip Gloss / ANSI | Current token | Use |
| --- | --- | --- | --- |
| Success / ready | `"2"` green | `styleReady` | ready verdicts, check marks, `All clear`, healthy state |
| Warning / needs action | `"3"` yellow | `styleNeedsAction` | needs-action issue numbers, setup attention, warning section titles |
| Error | `"1"` red | `styleError` | blocking verdicts, error issue numbers, unsafe states |
| Primary action / command | `"14"` bright cyan, bold | `styleBrightCyan` | exact commands and primary command-like remediation |
| Supporting accent | `"6"` cyan | `styleCyan` | StageServe header glyph, compact arrows, secondary command references |
| Primary structure | `"15"` bright white, bold | `styleWhite` | titles and key issue labels |
| Supporting text | `"7"` light grey | `styleMuted` | descriptions and secondary explanatory copy |
| De-emphasised text | `"8"` dark grey | `styleDim` | dividers, quiet evidence, compact passed-state text |
| Neutral emphasis | bold only | `styleBold` | labels such as `To fix:` and aligned row labels |

Do not use colour to make a screen feel more designed. Use colour only to help the user read state, priority, and action.

### Weight And Type Hierarchy

Terminal typography is limited, so hierarchy comes from role, order, weight, and spacing.

| Element | Treatment |
| --- | --- |
| Product/surface title | bright white, bold |
| Verdict | status colour, bold, early |
| Section title | semantic colour title with dim rule |
| Primary issue label | bright white, bold |
| Description | muted, short, sentence case |
| Observed problem | dim, close to remediation |
| Exact command | bright cyan, bold, copy-pasteable |
| Passing evidence | dim and compact |

Avoid visual shouting. StageServe should not use banners, ornamental borders, celebratory success copy, or decorative colour blocks.

### Glyphs

Glyphs are part of the visual identity when they communicate state or focus.

| Glyph | Meaning |
| --- | --- |
| `◆` | StageServe surface identity |
| `✓` | ready, passed, complete |
| `✗` | not ready or blocked |
| `!` | warning or needs action in compact contexts |
| `▶` | current focus or default action |
| `•` | pending, inactive, or secondary step |
| `▸` | compact next-step indicator |

Plain text fallback may replace glyphs with words or simpler markers when needed, but it must preserve the same hierarchy and meaning.

## Component Vocabulary

Use these components as the reusable design language. The implementation may render them with Bubble Tea, Lip Gloss, or plain text, but the component intent stays stable.

| Component | Job | Design notes |
| --- | --- | --- |
| Surface header | Establish product and current surface | `◆ StageServe` plus `Doctor`, `Setup`, `Project`, or `Recovery` |
| Verdict line | State the human outcome | First sentence about state; appears before detail |
| Key facts | Show values StageServe will use | Aligned labels, values, and short source notes |
| Report section | Present evidence | Blockers before passing checks; remediation adjacent to blocker |
| Work checklist | Show tool-owned setup or recovery sequence | One active step; visible next action; no menu of diagnostics |
| Decision list | Offer real user choices | Default first, low-risk, plain-language labels |
| Assistance handoff | Move from report to guided help | Optional, after the report remains useful on its own |
| Confirmation sheet | Bound mutation and risk | Names what changes, what does not, and the target path or URL |
| More panel | Hold power-user material | Direct commands, plain text output, advanced troubleshooting |
| Footer help | Show local key hints | Only keys that matter on that screen |

Every new renderer helper should map to one of these components or deliberately extend the component vocabulary in this guide.

## Example: Report Seed

```text
  ◆  StageServe Doctor
  ──────────────────────────────────────

  ✗  Not ready - 2 of 7 checks need attention.

── Needs fixing ────────────────────────

  1  Port 443
     Port 443 must be free for the local HTTPS gateway to bind to it.

     Something else on your computer is using port 443.
     To fix:  sudo lsof -nP -iTCP:443 -sTCP:LISTEN

── All clear ───────────────────────────

  ✓  State directory    exists

  ──────────────────────────────────────
  Fix the issues above, then run: stage doctor
```

This works because it is evidence-first. The report answers context, verdict, blocker, reason, and exact remediation before offering anything interactive.

## Example: Guided Surface

```text
  ◆  StageServe                         Project
  ──────────────────────────────────────

  This project is ready to run.

── Key facts ───────────────────────────

  Local URL       http://pete-site.develop
  Web folder      ./public_html
  Status          not running yet

── What you can do ─────────────────────

▶ Run this project
    Start the project and open it in your browser.

  Edit project settings
    Change site name, web folder, or domain suffix first.

  More…
    Show direct commands, plain text output, and advanced detail.
```

This works because it is action-first. It shows visible defaults before action, gives one low-risk default, and keeps implementation detail secondary.

## Bubble Tea And Lip Gloss Separation

Bubble Tea and Lip Gloss are implementation tools for this style guide. They do not own product truth.

Keep the layers separate:

| Layer | Owns | Must not own |
| --- | --- | --- |
| Planner/domain | situation, effective values, blockers, action availability | cursor state, colours, key bindings |
| Screen plan | normalized facts, sections, actions, severity, copy slots | lifecycle execution or terminal styling mechanics |
| Bubble Tea model | focus, selected action, modal state, drafts, in-flight UI state | whether a project is running or what setup requires |
| Renderer/style layer | components, spacing, glyphs, Lip Gloss tokens, text/TUI parity | command behavior or business decisions |
| Action layer | setup, lifecycle, logs, status, recovery execution | visual hierarchy or copy composition |

Good design coding turns this guide into reusable tokens and components. Functional implementation consumes the planner output and invokes actions. The two should meet through a screen-plan contract, not through ad hoc string assembly inside command logic.

## What Is Fixed And What Can Evolve

Fixed until explicitly redesigned:

- semantic colour roles.
- report-first anatomy for `stage doctor` and similar diagnostics.
- visible defaults before mutating actions.
- plain-language first-level labels.
- `More…` as the boundary for direct commands and advanced detail.
- text/TUI information parity.
- Bubble Tea as interaction state, not domain truth.

Allowed to evolve with a documented design reason:

- exact divider width.
- spacing details for narrow terminals.
- component helper names.
- glyph substitutions for accessibility or platform support.
- section titles when a clearer product term emerges.
- guided-flow examples as implementation teaches us more.

## Agent Guardrails

When an agent is asked to work on the TUI design, it should treat this guide as the viewable style guide and then apply the scoped instruction files.

Do:

- Start from product situation and visual identity before code.
- Use `stage doctor` as the production report seed, not as a limitation.
- Preserve separation between planner truth, Bubble Tea state, renderer components, and actions.
- Update this guide, examples, and agent instructions together when the identity changes.
- Replace stale prototype decisions when they conflict with the current guide.

Do not:

- Treat a prototype, spec draft, or existing renderer helper as the style guide.
- Add decorative colours, boxes, banners, or novelty glyphs.
- Let Bubble Tea decide lifecycle or config truth.
- Keep Docker, compose, gateway, registry, runtime, `attach`, or `detach` in first-level copy unless no clearer user wording exists.
- Add a new visual primitive without documenting its role here and in the markup spec.

## Style Review Checklist

Before approving a terminal design or implementation, check:

- The screen looks like the same StageServe identity as `stage doctor`.
- The first scan reveals context, verdict, priority, and next action.
- Colour, glyph, weight, and spacing all have jobs.
- The default action and default values are visible before commitment.
- Report surfaces remain useful before any assisted flow begins.
- Guided surfaces are action-first and safe-by-default.
- Advanced implementation detail is secondary.
- Plain text output preserves the same truth.
- Renderer helpers express components from this guide rather than one-off layouts.
- Any departure from this guide has an explicit design reason and updated examples.