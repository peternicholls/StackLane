# Implementation Plan: Stacklane Rebrand And Unified Command Surface

**Branch**: `002-project-rebrand` | **Date**: 2026-04-01 | **Spec**: `specs/002-project-rebrand/spec.md`
**Input**: Feature specification from `specs/002-project-rebrand/spec.md`

## Summary

Rebrand the project to Stacklane, replace the current family of `legacy wrapper commands` entrypoints with a single `stacklane` command that uses action modifiers such as `--up` and `--down`, and update all maintained user-facing surfaces in the same delivery unit. Implementation will preserve the existing runtime model in `lib/stacklane-common.sh`, introduce one canonical command dispatcher, retain `legacy wrapper commands` scripts temporarily as deprecation wrappers, and update docs plus GUI-facing text so the rename and migration are explicit and low-friction.

## Technical Context

**Language/Version**: Bash on macOS with POSIX shell workflow, Docker Compose YAML, AppleScript/Automator assets for GUI wrappers  
**Primary Dependencies**: Docker Desktop, Docker Compose, Homebrew `dnsmasq`, Bash helper library in `lib/stacklane-common.sh`, macOS Automator/workflow assets  
**Storage**: File-based project and stack state under `.20i-state/`, documentation files, shell wrapper scripts, app/workflow metadata  
**Testing**: Manual CLI validation for startup/status/teardown and failure messaging, plus shell-level syntax validation where practical (`bash -n`; `shellcheck` if available)  
**Target Platform**: macOS developer machines using Docker Desktop
**Project Type**: Shell-first local development stack with CLI, Docker runtime templates, and macOS GUI wrappers  
**Performance Goals**: Preserve current operator responsiveness; command dispatch and help output should remain effectively immediate for local shell usage  
**Constraints**: Must preserve current runtime semantics, config precedence, and state isolation; must document repo rename separately from manual local-folder rename; deployed copy under `$HOME/docker/20i-stack` may require explicit sync messaging  
**Scale/Scope**: One repository-wide rename, one unified CLI surface, all maintained docs, and all current user-facing wrapper surfaces in this repo

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Ease-of-use impact is documented, including whether the change preserves
  the shortest obvious operator path and what friction it removes or adds.
- [x] Reliability expectations are explicit, including backward compatibility or
  migration behavior, canonical variable names, defaults, required values, and
  precedence order (CLI override -> `.20i-local` -> shell environment -> stack
  defaults).
- [x] Robustness boundaries are defined for containers, volumes, networks,
  shared services, runtime data, isolation, and recovery from partial failure
  or drift.
- [x] Documentation surfaces requiring same-change updates are identified
  (`README.md`, `AUTOMATION-README.md`, `GUI-HELP.md`, shell help text,
  automation messaging, and other affected docs).
- [x] Validation covers startup, inspection/status, teardown, and at least one
  relevant failure path; any untestable area is called out explicitly.

### Constitution Check Notes

- Ease of use is improved by moving to one memorable command, `stacklane`, while keeping temporary wrapper scripts to avoid an abrupt migration cliff.
- Reliability is preserved by reusing the current helper library and keeping config precedence unchanged; only the invocation surface and user-facing vocabulary change.
- Robustness work includes keeping current state files, shared gateway behavior, and project-selection safeguards intact while ensuring deprecated entrypoints fail clearly only after the migration period.
- Same-change documentation scope includes all top-level docs, runtime contract terminology, migration content, shell integration examples, wrapper messaging, and macOS GUI labels/help.
- Validation will cover `stacklane --up`, `stacklane --status`, `stacklane --down`, and one failure path for a legacy `legacy wrapper commands` wrapper or invalid action syntax. GUI parity remains partial and any unvalidated asset-packaging step must be called out explicitly.

## Project Structure

### Documentation (this feature)

```text
specs/002-project-rebrand/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli-contract.md
└── tasks.md
```

### Source Code (repository root)

```text
deprecated --up wrapper
deprecated --attach wrapper
deprecated --down wrapper
deprecated --detach wrapper
deprecated --dns-setup wrapper
deprecated --status wrapper
deprecated --logs wrapper
legacy GUI wrapper
README.md
AUTOMATION-README.md
GUI-HELP.md
docs/
├── migration.md
├── runtime-contract.md
└── plan.md
lib/
└── stacklane-common.sh
Stacklane Manager.app/ (archived → previous-version-archive/)
Stacklane Manager.workflow/ (archived → previous-version-archive/)
20i-stack-launcher.workflow (archived → previous-version-archive/)
20i-stack-manager.scpt (archived → previous-version-archive/)
docker-compose.yml
docker-compose.shared.yml
docker/
```

**Structure Decision**: This is a single-repository shell and documentation feature. Core implementation will center on `lib/stacklane-common.sh`, a new top-level `stacklane` dispatcher script at the repo root, the existing `legacy wrapper commands` wrapper scripts, and the maintained documentation and GUI-facing text surfaces listed above.

## Post-Design Constitution Check

- [x] Ease-of-use remains improved after design: one canonical `stacklane` entrypoint, temporary migration wrappers, and same-change help/documentation updates are defined.
- [x] Reliability remains explicit after design: the plan preserves the existing runtime engine, config precedence, and state model while changing only invocation and wording surfaces.
- [x] Robustness remains bounded after design: shared gateway behavior, project selection, `.20i-state` storage, and recovery messaging remain in scope and unchanged unless explicitly documented.
- [x] Documentation parity remains covered after design: top-level docs, migration docs, runtime contract text, shell examples, and GUI-facing labels are all named as required same-change surfaces.
- [x] Validation remains sufficient after design: quickstart covers startup, status, teardown, migration-wrapper behavior, and invalid-action failure paths, with GUI packaging caveats still called out if they cannot be fully exercised.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**
> Include the exact principle or gate being violated.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
