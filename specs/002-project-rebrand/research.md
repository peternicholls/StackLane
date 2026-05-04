# Phase 0 Research: StageServe Rebrand And `stage` Command Cutover

## Decision 1: Keep the current Go runtime and thin root launcher

- Decision: Use the current `stage` launcher and `stage-bin` binary as the active command path.
- Rationale: The active product is already the Go CLI. Reusing that surface keeps the rename work focused on naming alignment rather than runtime churn.
- Alternatives considered: Restoring archived shell dispatch as a migration layer was rejected because archived Bash behavior is not current functionality.

## Decision 2: Use subcommands on `stage`

- Decision: Express primary actions as `stage up`, `stage attach`, `stage status`, `stage down`, and related subcommands.
- Rationale: This is the implemented command surface and the supported operator contract.
- Alternatives considered: Flag-driven action syntax was rejected because it no longer matches the active CLI.

## Decision 3: Ship no compatibility forwarding shim

- Decision: Treat `stage` as the only supported executable path.
- Rationale: The current rename contract explicitly rejects forwarding shims so machine-readable output and help text stay clean.
- Alternatives considered: Temporary old-command wrappers were rejected because they prolong stale naming in active docs and runtime expectations.

## Decision 4: Use current StageServe config and state names

- Decision: Standardize on project `.env.stageserve`, stack-home `.env.stageserve`, and `.stageserve-state`.
- Rationale: These are the implemented human-owned and runtime-owned paths and match the active operator docs.
- Alternatives considered: Keeping prior config/state names in maintained specs was rejected because that would leave active contract drift behind after the rename.

## Decision 5: Treat docs and older specs as same-change surfaces

- Decision: Update repository docs, migration docs, runtime contract wording, and older spec text in the same pass as the command/name cleanup.
- Rationale: Mixed naming across active docs and specs makes the contract harder to trust.
- Alternatives considered: Deferring older spec cleanup was rejected because it leaves stale references searchable in maintained material.

## Decision 6: Make deployed-copy sync explicit in migration guidance

- Decision: Document that changes in this repository may still need to be synced to the stack copy an operator actually runs.
- Rationale: A rename sweep is only useful if operators can distinguish workspace edits from their live install.
- Alternatives considered: Ignoring deployed-copy drift was rejected because it leaves a known operator trap undocumented.