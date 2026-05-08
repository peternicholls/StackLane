# StageServe Terminal Style Guide TUI

This directory contains the runnable terminal edition of the StageServe design style guide.

It is a documentation artifact, not a production command and not a spec prototype. It demonstrates the terminal identity using Bubble Tea and Lip Gloss while keeping Docker, project state, and StageServe runtime behavior untouched.

## Run

```bash
go run ./docs/design/style-guide-tui
```

## Controls

- `↑` / `↓` or `tab`: move through guide sections
- `j` / `k`, `page up`, `page down`: scroll the current section
- `enter`: return the current section to the top
- `?`: show or hide help
- `q` / `esc`: quit

## Plain Text

```bash
go run ./docs/design/style-guide-tui --plain
go run ./docs/design/style-guide-tui --section colour --plain
go run ./docs/design/style-guide-tui --list-sections
```

Plain output is provided so the guide can be reviewed in non-interactive terminals and checked in scripts.