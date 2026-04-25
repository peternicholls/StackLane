# Stacklane TUI Interaction Patterns

## Status

There is no committed TUI keymap today. The old shortcut list assumed a finished product and should not be treated as a contract.

This document keeps only the useful interaction ideas and recasts them as provisional patterns in case a TUI is explored again.

## Principles For A TUI If We Build One

- shortcuts should mirror real Stacklane actions,
- project selection should be first-class,
- destructive actions should require clear confirmation,
- keybindings should be easy to discover on screen,
- the TUI should not expose operations the CLI/runtime does not actually support.

## Good Candidate Actions

These are the actions most worth surfacing in a TUI because they map cleanly to the current product:

- project selection,
- `up`,
- `attach`,
- `down`,
- `detach`,
- `status` refresh,
- logs for a selected project or service,
- opening the routed hostname,
- viewing DNS and gateway health.

## Provisional Shortcut Shape

If a TUI is pursued, a sensible starting point would be:

| Key | Candidate action |
|-----|------------------|
| `q` | Quit |
| `?` | Toggle help |
| `p` | Pick project |
| `u` | Up current project |
| `a` | Attach current project |
| `d` | Down current project |
| `x` | Detach current project |
| `l` | Show logs |
| `r` | Refresh status |
| `o` | Open routed hostname |

This is better aligned with the current Stacklane vocabulary than the older stack-manager shortcuts.

## Navigation Suggestions

If the interface stays terminal-native, it should support:

- arrow keys,
- `j` and `k` style movement where natural,
- `Tab` only if there is a clear focus model,
- minimal modal complexity.

The earlier three-panel assumption may prove useful, but it should not be locked before the actual information hierarchy is tested.

## Confirmations

Destructive actions should remain explicit.

Reasonable candidates for confirmation:

- detach,
- down when data loss or service interruption is likely,
- global teardown,
- any future reset or purge operation.

The exact confirmation style can differ between TUI and GUI. What matters is that the user understands scope: current project, selected project, or all projects.

## Mouse And Click Support

This should stay optional.

Helpful if available:

- selecting a project,
- opening a hostname,
- expanding logs or diagnostics,
- copying paths or URLs.

But the interaction model should not depend on mouse support to be usable.

## If We Choose GUI Instead

Most of the same interaction ideas still apply. The difference is only the affordance:

- shortcuts become toolbar actions or menu items,
- project selection becomes a list or sidebar,
- health becomes cards, tables, or badges,
- logs become panes or drawers.

That is one reason not to over-invest in a terminal-specific keymap too early.
