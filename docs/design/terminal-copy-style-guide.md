# StageServe Terminal Copy Style Guide

This guide governs the words inside StageServe terminal interfaces: labels, verdicts, explanations, remediation, next actions, warnings, and empty states.

Use it when designing output, changing command text, reviewing agent-written copy, or deciding whether a message belongs in terminal output at all.

## Voice

StageServe sounds like a capable local-development tool.

- **Plain:** use common words before implementation terms.
- **Specific:** name the state, blocker, or action.
- **Calm:** describe failure without drama, apology, or blame.
- **Operational:** write copy that helps the user continue.
- **Short:** remove lines that do not change user understanding or action.

Do not write like logs, docs, marketing, or a chat assistant. Terminal copy is product UI.

## First-Level Language Boundary

StageServe has two language zones:

- **First-level product language:** what the user sees in the normal flow.
- **Advanced and troubleshooting language:** implementation detail for power users.

First-level copy stays in user-goal language. Docker, compose, gateway, runtime, registry, state record, `attach`, and `detach` do not lead the conversation.

## Copy Hierarchy

Write terminal output in this order:

1. **Context:** the command surface or product area.
2. **Verdict:** the human state of the interaction.
3. **Reason:** the blocker, confirmation, or relevant detail.
4. **Action:** the exact next command or next decision.

If a line does not serve one of those roles, remove it.

## Vocabulary

Use stable user-facing terms.

| Concept | Preferred copy | Avoid |
|---|---|---|
| Product | `StageServe` | `stage runtime`, `gateway manager` |
| Success verdict | `Ready - all checks passed.` | `Success!`, `OK`, `OverallReady` |
| Blocked verdict | `Not ready - N of M checks need attention.` | `Failed`, `needs_action`, `Result: Needs action` |
| Fix label | `To fix:` | `Run:`, `Please run`, `Try` |
| Next step label | `Next:` | `You should`, `Continue by`, `Recommended` |
| No items | `No projects are registered yet.` | `0 projects`, `Empty list` |
| Planned change | `Will update` | `Mutation plan`, `Diff payload` |

Use internal IDs only in tests, logs, or developer-facing diagnostics. Do not print them as user copy.

## User-Goal Labels

Prefer labels that describe what the user is trying to achieve.

| Intent | Preferred label | Avoid |
| --- | --- | --- |
| Start a configured project | `Run this project` | `Attach`, `Bootstrap runtime` |
| Stop a running project | `Stop this project` | `Down`, `Shutdown runtime` |
| Create first project config | `Set up this folder as a project` | `Init`, `Create stack config` |
| Forget a stopped project | `Remove this project from StageServe` | `Detach`, `Remove record` |
| Show power-user paths | `More…` | `Extras`, `Utility menu` |
| Show command equivalents | `Show direct commands` | `Command list`, `CLI mode` |
| Explain implementation detail | `Advanced and troubleshooting` | `Internals`, `Diagnostics` |
| Non-interactive path | `Plain text output` | `CLI fallback`, `No TUI mode` |

## Advanced-Only Terms

These terms are valid, but they do not belong in first-level user text unless the screen is explicitly an advanced or troubleshooting surface.

| Advanced-only term | Say this first |
| --- | --- |
| `drift` | `doesn't match what StageServe expects` |
| `gateway` | `local URL routing` or no first-level term at all |
| `compose` | no first-level term |
| `container` | no first-level term |
| `daemon` | `Docker Desktop isn't running` |
| `runtime` | `project` or `local site` |
| `registry` / `state record` | `StageServe record` only in advanced detail |
| `docroot` | `web folder` |
| `attach` | `Add this project to StageServe` |
| `detach` | `Remove this project from StageServe` |

## Sentence Patterns

Verdicts:

```text
Ready - all checks passed.
Not ready - 2 of 7 checks need attention.
Local setup is not ready yet.
No projects are registered yet.
Previewing changes. Nothing has been changed.
```

Reasons:

```text
Docker is installed, but the daemon is not running.
Another process is already listening on port 443.
*.test domains are not resolving to localhost.
The project will stop routing through the local gateway.
```

Actions:

```text
To fix:  open -a Docker
Next:  stage init
Next:  stage project remove api.test --confirm
```

Guided verdicts:

```text
Your computer isn't ready yet.
This folder doesn't have StageServe settings yet.
This project is ready to run.
This project is running at http://pete-site.develop.
This project doesn't match what StageServe expects.
StageServe couldn't safely choose a next step.
```

## Remediation Copy

Remediation must be an exact shell command, not an instruction sentence.

Good:

```text
To fix:  sudo lsof -nP -iTCP:443 -sTCP:LISTEN
```

Bad:

```text
To fix:  Check what is using port 443 and close it.
Run: sudo lsof -nP -iTCP:443 -sTCP:LISTEN (shows listeners)
```

When there are multiple valid fixes, explain the choice in a sentence above the commands. Keep each command copy-pasteable.

## Labels

Labels should match user concepts and fit compact terminal layouts.

- Prefer `Docker daemon` over `docker.daemon`.
- Prefer `State directory` over `state.dir`.
- Prefer `DNS resolver` over `dnsmasq check`.
- Prefer `Port 443` over `https_port_available`.

If a label needs extra explanation, keep the label short and put the explanation in the description line.

## Defaults And Action Copy

When a screen has a highlighted default action, the description beneath that action is part of the contract.

Good:

```text
▶ Run this project
	Start the project and open it in your browser.
```

Bad:

```text
▶ Continue
	Proceed with the next step.
```

The user should know exactly what pressing `enter` will do.

## Descriptions

Descriptions explain why a check or step matters. They are not instructions.

Good:

```text
The Docker daemon must be running before any container can start.
Port 443 must be free for the local HTTPS gateway to bind to it.
```

Bad:

```text
You must start Docker.
Note: port 443 is required.
This is used by the runtime subsystem.
```

## State-Specific Guidance

Success should be quiet. Give the verdict, concise evidence when useful, and the next command if there is a natural handoff.

Failure should be useful. Lead with the blocker, keep passing details secondary, and show an exact next action.

Empty states should teach one next step. Do not print empty tables.

Warnings should state the consequence. If an action is destructive, say what changes and what does not change.

Long-running work should name the current step and set expectation only when waiting is normal.

## Confirmation Copy

Confirmation screens must show:

- what StageServe is about to act on
- what value or path it will use
- what will change
- what will not change

Good:

```text
StageServe will create /Users/pete/sites/pete-site/.env.stageserve.
StageServe will not change any other file.
```

Good:

```text
StageServe will stop pete-site.
Your files will not be touched.
http://pete-site.develop will no longer respond.
```

## Recovery Copy

Recovery copy should describe observable truth first, then the safe next step.

Good:

```text
StageServe expected this project to be running, but http://pete-site.develop is not responding.
StageServe will treat this project as stopped. Nothing in your folder will be deleted.
```

Bad:

```text
Drift detected.
The registry and runtime are inconsistent.
```

## Editing Checklist

Before shipping terminal copy, check:

- The verdict is human-readable.
- The first action is exact and copy-pasteable.
- Internal status names, IDs, counters, and enum values are absent.
- First-level copy uses user-goal language before implementation terms.
- The copy still works without colour or icons.
- Success is quieter than failure.
- Descriptions explain why, not what command to run.
- Each line changes the user's understanding or next action.
- New reusable wording is reflected in the prototype catalog.
