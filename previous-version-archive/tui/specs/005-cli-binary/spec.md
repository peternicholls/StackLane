<file name=0 path=/Users/peternicholls/docker/20i-stack/specs/005-cli-binary/spec.md># Feature Specification: Self-Contained CLI and TUI Binary (Command: 20i)

**Feature Branch**: `005-cli-binary`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🔴 Critical  
**Input**: User description: "Single-file executable for global or per-project installation with commands for init, start, stop, status, logs"

## Product Contract *(mandatory)*

The CLI and TUI are two interfaces to the **same underlying stack engine**. The CLI MUST NOT re-implement stack behaviour that already exists in the stack engine packages.

### Project identity and determinism

- The stack instance identity MUST be derived deterministically from the project directory.
- The CLI MUST use the same project name sanitisation rules as the TUI/stack (so container labels and filtering remain consistent).

### Configuration precedence

When resolving configuration values, the CLI MUST apply precedence in this order:

0. CLI flags (e.g. `--port`, `--profile`)  
1. Process environment (explicit `export VAR=...`)  
2. Project-local overrides (e.g. `.env` and/or `.20i-local` where supported)  
3. Project config file (`.20i-config.yml`)  
4. Central defaults (`config/stack-vars.yml`)  
5. Hardcoded defaults (only if nothing else specifies a value)

### User-level state and preferences

- The CLI MUST NOT store user-level settings, preferences, or cached state inside project directories.
- All user-level state (e.g. preferences, UI state, caches, last-used options) MUST be stored in a dedicated dot directory in the user’s home directory.
- The default location MUST be:
  - `~/.20i/` on macOS and Linux
- Project directories MUST remain portable and free of hidden user-specific state.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Initialize Stack in Project (Priority: P1)

As a developer, I want to run a single command to initialize the 20i stack in my project directory so that I can start developing immediately without manual setup.

**Why this priority**: Initialization is the entry point to using the stack; without it, no other commands work.

**Independent Test**: Run `20i init` in an empty directory, verify `.20i-config.yml` is created and the CLI can resolve the stack definition (embedded assets or STACK_HOME reference) without manual steps.

**Acceptance Scenarios**:

1. **Given** an empty project directory, **When** the user runs `20i init`, **Then** `.20i-config.yml` is created with default settings  
2. **Given** `20i init` is run, **When** initialization completes, **Then** the project is linked to a stack definition (default: embedded assets; fallback: resolved STACK_HOME/STACK_FILE) and the CLI can run `20i status` without extra setup  
3. **Given** a directory already initialized, **When** the user runs `20i init`, **Then** the system prompts for confirmation before overwriting

#### Init contract

- `20i init` MUST be idempotent.  
- `20i init` MUST NOT overwrite an existing `.20i-config.yml` without explicit confirmation.  
- `20i init` MUST validate that the stack definition can be resolved (default: embedded assets; fallback: STACK_HOME/STACK_FILE).

#### `.20i-config.yml` minimal schema *(MVP)*

The project config file lives in the project root. For Phase 1 of the CLI, the schema MUST support at least:

- `schema_version` (integer)
- `profile` (string, optional)
- `ports` (object, optional)
  - `host` (int)
  - `pma` (int)
  - `mysql` (int)
- `overrides` (object, optional)
  - `php_version` (string)
  - `mysql_version` (string)
  - `pma_image` (string)

Example:

```yaml
schema_version: 1
profile: default
ports:
  host: 80
  pma: 8081
  mysql: 3306
overrides:
  php_version: "8.5"
  mysql_version: "10.6"
  pma_image: "phpmyadmin/phpmyadmin:latest"
```

Notes:

- Unknown keys MUST be ignored safely (but SHOULD warn in logs).
- Invalid values MUST fail fast with an actionable error.

---

### User Story 2 - Start and Stop Stack (Priority: P1)

As a developer, I want to start and stop the stack with simple commands so that I can manage my development environment efficiently.

**Why this priority**: Start/stop are the core daily operations; essential for basic functionality.

**Independent Test**: Run `20i start` in initialized project, verify all containers are running; run `20i stop`, verify all containers are stopped.

**Acceptance Scenarios**:

1. **Given** an initialized project, **When** the user runs `20i start`, **Then** all stack services start within 60 seconds  
2. **Given** a running stack, **When** the user runs `20i stop`, **Then** all containers stop gracefully  
3. **Given** a custom port specified, **When** the user runs `20i start --port 8080`, **Then** the web server is accessible on port 8080

#### Start/stop contract

- `20i start` MUST fail fast with an actionable error if Docker is not available.  
- `20i start` MUST set `CODE_DIR` to the project root and `COMPOSE_PROJECT_NAME` to the sanitized project name.  
- `20i stop` MUST stop only the stack for the current project.  
- Commands MUST be executed without a shell (no string-based `sh -c` execution).

---

### User Story 3 - View Stack Status (Priority: P2)

As a developer, I want to see the status of all stack services so that I can diagnose issues quickly.

**Why this priority**: Status visibility is essential for troubleshooting but not required for basic operation.

**Independent Test**: Run `20i status` with stack running, verify output shows container states, ports, and uptime.

**Acceptance Scenarios**:

1. **Given** a running stack, **When** the user runs `20i status`, **Then** all service names, states, and ports are displayed  
2. **Given** no stack running, **When** the user runs `20i status`, **Then** the system shows "No stack running in this directory"  
3. **Given** multiple stacks running in different directories, **When** the user runs `20i status --all`, **Then** all running stacks are listed

#### Status contract

- `20i status` MUST use container labels (e.g. compose project label) to identify the current project’s stack.  
- If run outside an initialized project directory, it MUST print a clear message and a next-step suggestion (e.g. "Run `20i init`.").

---

### User Story 4 - View Service Logs (Priority: P2)

As a developer, I want to view logs for specific services so that I can debug application issues.

**Why this priority**: Log access is important for debugging but not required for basic stack operation.

**Independent Test**: Run `20i logs apache` with stack running, verify Apache access/error logs are streamed to terminal.

**Acceptance Scenarios**:

1. **Given** a running stack, **When** the user runs `20i logs apache`, **Then** Apache logs are streamed to the terminal  
2. **Given** a running stack, **When** the user runs `20i logs` (no service specified), **Then** logs from all services are interleaved  
3. **Given** the `--follow` flag, **When** the user runs `20i logs --follow`, **Then** new log entries are streamed in real-time  
4. **Given** a service name is unknown, **When** the user runs `20i logs unknown`, **Then** the CLI prints valid service names and exits non-zero

---

### User Story 5 - Destroy Stack and Volumes (Priority: P3)

As a developer, I want to completely remove a stack including volumes so that I can start fresh or clean up after a project.

**Why this priority**: Cleanup is important for maintenance but less frequently used than start/stop.

**Independent Test**: Run `20i destroy` in project with running stack, verify all containers are removed and volumes are deleted.

**Acceptance Scenarios**:

1. **Given** a running stack, **When** the user runs `20i destroy`, **Then** the system prompts for confirmation  
2. **Given** confirmation provided, **When** destroy completes, **Then** all containers and volumes are removed  
3. **Given** the `--force` flag, **When** the user runs `20i destroy --force`, **Then** no confirmation prompt is shown

#### Destroy safety rules

- `20i destroy` MUST target only the current project stack.  
- The confirmation prompt MUST clearly state that volumes will be removed.  
- `--force` MUST still be scoped to the current project and MUST NOT affect other projects.

---

### User Story 6 - Install CLI Globally (Priority: P3)

As a developer, I want to install the CLI globally using a single command so that I can use it across all my projects.

**Why this priority**: Easy installation improves adoption but users can also use per-project installation.

**Independent Test**: Run the curl installer command, verify `20i` is available in PATH and responds to `20i version`.

**Acceptance Scenarios**:

1. **Given** a macOS system, **When** the user runs the curl installer, **Then** `20i` binary is installed to a user-writable PATH location (prefer `~/.local/bin`, otherwise `/usr/local/bin` if writable). The canonical command name MUST be `20i`.  
2. **Given** Homebrew is installed, **When** the user runs `brew install peternicholls/tap/20i-stack`, **Then** the CLI is installed and available  
3. **Given** the CLI is installed, **When** the user runs `20i version`, **Then** CLI version and stack version are displayed

#### Installation contract

- Install scripts MUST NOT require sudo unless the target path requires it.  
- The CLI MUST provide `20i version` showing:  
  - CLI version  
  - Git commit SHA (if available)  
  - Stack definition version/source (embedded assets version or STACK_HOME reference)

---

### User Story 7 - Launch TUI Dashboard (Priority: P2)

As a developer, I want to launch the interactive TUI dashboard from the CLI so that I can manage the stack with a guided interface.

**Why this priority**: The TUI is the default interactive experience; the CLI must be able to launch it reliably.

**Independent Test**: Run `20i tui` in an initialized project directory and verify the dashboard opens.

**Acceptance Scenarios**:

1. **Given** an interactive terminal, **When** the user runs `20i tui`, **Then** the dashboard TUI launches
2. **Given** a non-interactive environment (no TTY), **When** the user runs `20i tui`, **Then** the CLI exits non-zero with a clear message explaining that TUI requires an interactive terminal

---

### Edge Cases

- Docker not installed, or Docker daemon not running (must produce actionable guidance)  
- `docker compose` not available (must produce actionable guidance)  
- Permission denied (Docker socket or privileged ports) (must produce actionable guidance)  
- `.20i-config.yml` invalid or corrupted (must refuse to run and suggest re-init or repair)  
- Run outside a project directory (must suggest `20i init`)  
- Multiple stacks running across directories (commands must remain scoped unless `--all` is explicitly used)  
- Port conflicts (must name the port if detectable, and suggest override)


## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: CLI MUST be distributed as a single self-contained executable  
- **FR-001a**: The canonical installed binary/command name MUST be `20i`  
- **FR-002**: CLI MUST support `init`, `start`, `stop`, `status`, `logs`, `destroy`, `version`, and `tui` entrypoints  
- **FR-003**: CLI MUST work from any project directory with an initialized stack  
- **FR-004**: CLI MUST be able to resolve the stack definition either from embedded assets OR from a resolved STACK_HOME/STACK_FILE reference  
- **FR-005**: CLI MUST support macOS (Intel and ARM) and Linux (x64 and ARM64)  
- **FR-006**: CLI MUST be installable via curl one-liner, Homebrew, and per-project  
- **FR-007**: CLI MUST validate Docker availability on first run  
- **FR-008**: CLI MUST provide clear error messages with actionable suggestions  
- **FR-009**: CLI MUST support `--help` flag for all commands with usage examples  
- **FR-010**: CLI MUST behave deterministically: given the same project path and config, it MUST generate the same stack identity and environment variables

## Key Entities

- **Project**: A directory initialized via `20i init`, identified deterministically by its absolute path and sanitized project name.
- **Stack Definition**: The compose-based stack configuration resolved either from embedded assets or from a resolved `STACK_HOME/STACK_FILE` reference.
- **Stack Engine**: The shared core logic responsible for project detection, environment resolution, compose execution, and status inspection; used by both CLI and TUI.
- **CLI**: The command-line interface that orchestrates the stack engine; must not re-implement stack behaviour.
- **TUI**: The terminal UI interface that orchestrates the same stack engine and presents interactive views.
- **Container Identity**: Containers and resources labeled by the sanitized project name (e.g. compose project label) to ensure scoping and isolation.
- **User State Directory**: A per-user directory located at `~/.20i/` used to store preferences, UI state, and caches that must not live inside project directories.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: New users can install CLI and start first stack within 5 minutes  
- **SC-002**: `20i init` completes in under 10 seconds on a typical developer machine  
- **SC-003**: `20i start` brings up full stack in under 60 seconds (first run may be longer due to image pulls)  
- **SC-004**: CLI binary size is kept reasonably small; size regressions are tracked per release and justified (target < 25MB per platform unless embedding requires more)  
- **SC-005**: No Docker Compose knowledge required for basic usage  
- **SC-006**: CLI works on 95%+ of developer machines (macOS 10.15+, Ubuntu 20.04+, Debian 10+)

## Non-goals *(mandatory)*

- This feature does NOT replace the TUI; it complements it.  
- This feature does NOT attempt to perfectly emulate every aspect of 20i shared hosting; it focuses on the stack behaviours this repo provides.  
- This feature does NOT manage external secrets; users are responsible for secure credential handling.
- This feature does NOT store user preferences or machine-specific state inside project repositories.

## Assumptions

- Docker Desktop or Docker Engine is installed and running  
- Users have basic terminal familiarity  
- Internet connectivity available for initial image downloads  
- File system permissions allow writing to project directory
</file>
