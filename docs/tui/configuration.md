# Stacklane Interaction Surface Configuration Notes

## Main Point

Any future TUI or GUI should inherit the current Stacklane runtime configuration model rather than create a parallel one.

That means the first configuration question is not "how should the TUI be configured?" It is "what configuration already belongs to Stacklane itself?"

## Runtime Configuration Is Already Defined

The active Stacklane runtime currently resolves configuration in this order:

1. CLI flags
2. `.stacklane-local`
3. shell environment
4. stack `.env`
5. built-in defaults

That precedence should remain unchanged regardless of whether a future interaction layer is terminal-based or graphical.

## Runtime Inputs A Future Surface Must Understand

These are runtime settings, not UI settings:

- `SITE_NAME`
- `SITE_HOSTNAME`
- `SITE_SUFFIX`
- `DOCROOT`
- `CODE_DIR` as an alias
- `PHP_VERSION`
- `MYSQL_VERSION`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `MYSQL_PORT`
- `PMA_PORT`
- `SHARED_GATEWAY_HTTP_PORT`
- `SHARED_GATEWAY_HTTPS_PORT`
- `LOCAL_DNS_PROVIDER`
- `LOCAL_DNS_IP`
- `LOCAL_DNS_PORT`
- `LOCAL_DNS_SUFFIX`

Any richer interface should display and edit these carefully, but should not silently reinterpret them.

## State Inputs A Future Surface Must Read

The current runtime also has persistent state that matters to any operator surface:

- `.stacklane-state/projects/<slug>.json`
- `.stacklane-state/shared/gateway.conf`
- live Docker state
- shared gateway health
- local DNS health

That state should be treated as the basis for visibility screens, project lists, and health panels.

## What A Future UI Config Should Probably Cover

If a TUI or GUI is eventually built, its own configuration should stay small and separate from runtime config.

Likely UI-only preferences:

- refresh cadence,
- default view or landing screen,
- whether logs auto-follow,
- compact versus detailed layout,
- whether to show archived or down projects by default,
- theme or accessibility preferences.

These should not live in `.stacklane-local` because they are operator preferences, not project runtime settings.

## What Should Not Happen

Avoid these traps:

- inventing a second precedence system,
- making the UI write hidden runtime defaults that the CLI does not know about,
- storing project runtime truth in a UI-specific config file,
- requiring a richer UI just to access configuration the CLI already supports.

## Proposed Split

### Runtime config

Owned by Stacklane itself.

Examples:

- hostnames,
- docroot,
- PHP version,
- database credentials,
- gateway ports,
- DNS suffix and DNS service settings.

### Operator-surface config

Owned by the future TUI or GUI only if one exists.

Examples:

- panel widths,
- refresh preferences,
- layout density,
- remembered last-selected project,
- log presentation options.

## Recommendation For An Initial Interface

If a richer surface is built soon, it should begin in read-mostly mode:

- inspect current resolved config,
- inspect current state,
- launch existing actions through the CLI,
- avoid adding complex configuration editing until the surface proves useful.

This keeps the first iteration honest and low risk.
