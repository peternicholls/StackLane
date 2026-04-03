# Troubleshooting The Earlier TUI Direction

## Purpose

This is not a runtime troubleshooting guide for a shipped TUI. It is a guide to the main problems in the copied TUI documentation and the design assumptions behind it.

## Problem 1: The Docs Describe An Implemented Product That Does Not Exist

Symptoms:

- fixed build instructions,
- package install instructions,
- hard-coded Go and Bubble Tea versions,
- specific panel layouts and shortcuts presented as done,
- direct references to code paths that do not exist in this repo.

Correction:

- treat the older material as exploratory design work,
- remove statements that imply the TUI is built,
- describe decisions as proposals unless they are reflected in the actual repository.

## Problem 2: The Docs Drift Away From The Current Stacklane Runtime

Symptoms:

- they talk about a generic stack manager instead of Stacklane,
- they assume a simpler stack lifecycle than the current attachable multi-project model,
- they do not center `.20i-state`, registry-driven routing, shared gateway health, or DNS readiness,
- they use an older product vocabulary.

Correction:

- anchor all future interface thinking to the current Stacklane runtime contract,
- make project attachment, hostname routing, and registry state central concerns,
- keep the CLI as the canonical behavior definition.

## Problem 3: The Docs Lock In Too Many Technology Choices Too Early

Symptoms:

- Go is treated as chosen,
- Bubble Tea is treated as chosen,
- a TUI is treated as chosen,
- a specific panel architecture is treated as chosen.

Correction:

- separate the product problem from the implementation toolkit,
- document the operator experience goals first,
- only commit to a toolkit after deciding whether the right surface is a TUI, a GUI, or continued CLI evolution.

## Problem 4: The Earlier Design Over-Reaches The First Useful Version

Symptoms:

- heavy emphasis on dashboards before the core status model is fully stabilized,
- suggestions for deep container management before the product proves the need,
- extra UI complexity that may duplicate CLI work rather than clarify it.

Correction:

- start with visibility, project selection, health, and lifecycle entry points,
- avoid turning the first richer interface into a full control plane,
- prefer thin wrappers over runtime reinvention.

## Problem 5: The Docs Confuse Runtime Config With UI Config

Symptoms:

- environment variables are discussed as if they belong to the TUI,
- runtime settings and interface preferences are mixed together,
- it becomes unclear which layer owns behavior.

Correction:

- keep Stacklane runtime config authoritative,
- keep future UI preferences separate and optional,
- do not let a future interface invent hidden runtime state.

## Decision Questions That Are Still Open

These are the real unresolved questions now:

- Do we actually want a TUI?
- Would a GUI better fit the user experience we want?
- Should the next step be a richer CLI and machine-readable output instead?
- If a richer surface exists, should it be a thin wrapper over the current CLI first?

Those are healthier questions than immediately choosing a framework.

## Recommended Reading Order

If revisiting this direction later:

1. Read the main Stacklane README and runtime contract first.
2. Read the rewritten interaction notes in this folder second.
3. Decide the product surface before deciding the toolkit.
4. Keep the historical TUI ideas as inspiration, not as binding architecture.

**For TUI Issues**:
- Check this troubleshooting guide
- Review logs: `docker logs <service>`
- Check Docker daemon logs: `journalctl -u docker` (Linux)

**For Docker Issues**:
- Docker documentation: https://docs.docker.com/
- Stack Overflow: tag `docker` and `docker-compose`

**For 20i-Specific Issues**:
- Check 20i Hosting Stack documentation
- Contact 20i support
