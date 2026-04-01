# Phase 0 Research: Stacklane Rebrand And Unified Command Surface

## Decision 1: Keep the current runtime engine and add a single dispatcher command

- Decision: Implement `stacklane` as the new canonical entrypoint while continuing to route lifecycle behavior through the existing helper engine in `lib/stacklane-common.sh`.
- Rationale: The current command family is already thin wrappers around `twentyi_main`. Reusing that runtime engine minimizes behavior drift and isolates the change to command dispatch, help text, migration messaging, and documentation.
- Alternatives considered: Rewriting the runtime API around a new helper library was rejected because it adds avoidable risk to startup, state handling, gateway routing, and DNS behavior during what is primarily a rename and UX-surface change.

## Decision 2: Use action flags on `stacklane` rather than subcommands

- Decision: Express primary actions as flags such as `stacklane --up`, `stacklane --attach`, `stacklane --status`, and `stacklane --down`.
- Rationale: The feature request explicitly calls for one central command with modifiers like `--up` and `--down`. This preserves the user's stated UX direction and makes the command model easy to scan from help output.
- Alternatives considered: Subcommands such as `stacklane up` were rejected because they do not match the requested modifier-based interaction model.

## Decision 3: Keep legacy `20i-*` commands temporarily as wrappers with deprecation guidance

- Decision: Retain the existing command files as compatibility wrappers that forward to `stacklane` and tell users what the new syntax is.
- Rationale: Existing operators, shell aliases, and local habits already depend on the current command names. Temporary wrappers reduce migration shock while still making `stacklane` the only primary documented interface.
- Alternatives considered: Removing the legacy commands immediately was rejected because it would create unnecessary breakage in operator workflows and support docs. Keeping both command families as equal first-class interfaces was rejected because it would fail the simplicity goal.

## Decision 4: Do not rename configuration keys unless strictly necessary

- Decision: Keep `.20i-local`, `.env`, and current runtime variable names intact unless a rename is necessary to make the new CLI understandable.
- Rationale: The spec requires the underlying runtime model and precedence order to remain familiar. Renaming environment keys in the same feature would increase migration cost without materially improving the central command UX.
- Alternatives considered: Renaming variables and state paths to match Stacklane was rejected for this feature because it mixes identity cleanup with operational contract changes.

## Decision 5: Treat documentation and GUI-facing text as same-change surfaces

- Decision: Update repository docs, migration docs, runtime contract wording, shell integration examples, AppleScript/workflow labels, and GUI help in the same implementation phase as the CLI rename.
- Rationale: The constitution requires affected operator surfaces to stay aligned. Mixed branding or mixed command vocabulary would directly violate the spec's clarity and migration goals.
- Alternatives considered: Deferring GUI/help wording until later was rejected because it would leave conflicting names and instructions in operator-visible surfaces.

## Decision 6: Make deployed-copy sync explicit in migration guidance

- Decision: Document that changes in this repository may still need to be synced to the deployed stack copy under `$HOME/docker/20i-stack` for live local usage.
- Rationale: The repository already has an explicit divergence risk between the dev workspace and the deployed copy. The rebrand changes command names and docs, so the sync boundary must be made explicit to avoid silent mismatch.
- Alternatives considered: Ignoring the deployed-copy divergence was rejected because it would leave a known friction point undocumented.