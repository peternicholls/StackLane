# Stacklane Interaction Surface Notes

These documents are exploratory. They preserve useful ideas from an earlier TUI direction, but they are not a description of an implemented Stacklane interface.

The current product reality is simpler:

- `stacklane` is the implemented and authoritative interface.
- The runtime contract lives in the shell/Docker layer documented in the main README and runtime contract.
- A richer operator surface may still happen later, but it does not need to be a TUI.
- A GUI may prove to be a better fit for the project, and these notes should stay open to that possibility.

## Why These Notes Exist

The earlier TUI thinking contains good groundwork:

- it treats the stack as an operator-facing product rather than just a shell script bundle,
- it highlights the need for status, logs, and health visibility,
- it pushes toward clearer lifecycle operations and safer destructive actions,
- it encourages a better user experience around multi-project local development.

What needed correcting was the level of certainty. The copied docs assumed a specific Go/Bubble Tea implementation and presented it as already built. That is not the current state of Stacklane.

## Current Position

For now, the project should be read as:

- CLI first,
- runtime contract first,
- interaction layer second,
- toolkit and form factor still open.

That means any future TUI or GUI should wrap the existing Stacklane behavior rather than invent a separate product model.

## What A Future Operator Surface Must Respect

- The source of truth is the existing Stacklane runtime contract.
- Config precedence remains CLI flags, `.stacklane-local`, shell environment, stack `.env`, then defaults.
- The current state model is based on project slugs, hostname planning, `.stacklane-state`, registry rows, and live Docker identity.
- Multi-project attachment and shared-gateway behavior are core parts of the product, not optional embellishments.
- The interface must help with diagnosis and trust, not hide the runtime behind opaque abstractions.

## Candidate Directions

### 1. Stay CLI-only for now

This remains the default path. It has the least risk and matches the implemented product.

### 2. Build a thin TUI on top of the CLI/runtime contract

This keeps terminal-native workflows fast, but only makes sense if it genuinely improves operator clarity rather than duplicating shell commands in a harder-to-maintain layer.

### 3. Build a desktop GUI or local web GUI

This may be the better long-term direction if the project wants richer setup flows, clearer health views, friendlier multi-project management, and fewer keyboard-discovery problems.

## Reading Guide

- [architecture.md](architecture.md): recommended architecture for a future interaction layer, without locking the project into TUI.
- [configuration.md](configuration.md): how runtime configuration should relate to any future UI layer.
- [keyboard-shortcuts.md](keyboard-shortcuts.md): provisional interaction patterns if a TUI is explored further.
- [troubleshooting.md](troubleshooting.md): the main risks, false assumptions, and decision traps found in the copied material.

## Recommended Next Step

Do not treat these docs as a build brief for a TUI. Treat them as a cleaned-up design notebook for the broader question: what operator surface should sit on top of Stacklane after the CLI has proven itself?
