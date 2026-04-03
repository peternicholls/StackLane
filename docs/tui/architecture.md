# Stacklane Interaction Layer Architecture

## Purpose

This document describes the architecture a richer Stacklane operator surface should follow if one is built. It is intentionally broader than a TUI architecture because the project has not committed to TUI over GUI.

The important architectural question is not "how do we build a Bubble Tea app?" The important question is "how do we add an operator surface without creating a second, competing runtime model?"

## Current Ground Truth

Today, Stacklane already has a real core:

- a canonical CLI entrypoint: `stacklane`,
- a defined runtime contract,
- project-scoped state under `.20i-state`,
- registry-driven routing and attachment semantics,
- Docker Compose-based runtime orchestration,
- local DNS and shared gateway behavior.

Any future TUI or GUI must sit on top of that core.

## Architectural Principle

The runtime contract stays authoritative. The operator surface is an adapter and presentation layer, not a second implementation of stack behavior.

That yields one strong recommendation:

- avoid embedding fresh lifecycle rules in the UI layer,
- avoid duplicating config resolution logic in the UI layer,
- avoid building separate state files for UI concepts that already exist in `.20i-state`,
- prefer thin orchestration boundaries that call or wrap the existing Stacklane operations.

## Recommended Layering

### 1. Runtime Core

This already exists.

Responsibilities:

- resolve config precedence,
- derive project identity and hostname,
- manage Docker Compose projects,
- update `.20i-state` and `registry.tsv`,
- manage shared gateway and DNS integration,
- surface status and logs.

This is where truth lives.

### 2. Integration Adapter

This is the boundary a future TUI or GUI should talk to.

Initial recommendation:

- treat the existing `stacklane` command as the integration API,
- parse structured output only where necessary,
- add machine-readable output to the CLI if richer interfaces need it,
- move logic below the CLI only when there is a clear testing or maintainability benefit.

This is safer than rewriting the runtime and UI at the same time.

### 3. View Model Layer

This layer translates runtime truth into interface-friendly concepts.

Examples:

- current project,
- attached projects,
- gateway health,
- DNS readiness,
- route URL,
- runtime drift,
- per-project log target choices,
- destructive-action confirmation state.

This layer should not invent hidden rules. It should mainly reshape existing runtime information.

### 4. Surface Layer

This is the user-facing shell. It could be:

- a TUI,
- a desktop GUI,
- a local web GUI,
- or a hybrid approach.

The choice here should be led by user experience and maintenance cost, not by attachment to the earlier TUI concept.

## Why The Earlier TUI Docs Need Revision

The copied version assumed:

- Go as the chosen language,
- Bubble Tea as the chosen framework,
- a three-panel dashboard as the chosen interaction model,
- direct Docker SDK integration,
- packaging and installation details that are not part of this repo,
- feature coverage beyond the current Stacklane contract.

Those are implementation guesses, not project decisions.

## Stable Design Decisions Worth Keeping

These ideas are still strong and should survive regardless of TUI or GUI:

- show shared gateway, DNS, and per-project runtime state together,
- make destructive actions explicit and hard to trigger accidentally,
- make project selection first-class,
- distinguish recorded state from live Docker state,
- treat logs and diagnostics as core product features rather than secondary tools,
- make it obvious which hostname, docroot, and repo path a runtime actually maps to.

## Design Decisions That Should Stay Open

These should not be locked yet:

- TUI versus GUI,
- Go versus another implementation language,
- Bubble Tea versus another toolkit,
- fixed three-panel layouts,
- mouse-first interactions,
- aggressive background refresh intervals,
- deep per-container controls in the first iteration.

## Suggested Delivery Strategy

### Stage 1: Strengthen the CLI as the substrate

Before building a richer surface, make sure the CLI can expose everything the surface needs clearly and predictably.

That may include:

- consistent status fields,
- cleaner selector behavior,
- machine-readable status output,
- clearer error codes and failure messages.

### Stage 2: Build the thinnest useful operator surface

Start with a narrow problem:

- project overview,
- status visibility,
- attach/up/down/detach actions,
- logs entry points,
- DNS and gateway health.

Do not begin with an ambitious full control plane.

### Stage 3: Decide whether the surface should stay terminal-native

Once the interaction model is clearer, choose whether a TUI still makes sense or whether the project would benefit more from a GUI.

## Evaluation Criteria

If choosing between TUI and GUI, use these criteria:

- Does it expose the current runtime honestly?
- Does it reduce operator uncertainty?
- Does it improve setup and recovery flows?
- Does it avoid creating a second implementation burden?
- Does it fit the likely user environment, which is currently macOS-heavy?

## Provisional Recommendation

The best current architecture is:

- CLI/runtime contract remains the core,
- any future TUI or GUI is a thin adapter over that core,
- no commitment yet to Go or Bubble Tea,
- no commitment yet to terminal UI over desktop UI.
