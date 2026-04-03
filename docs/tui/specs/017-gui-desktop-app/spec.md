# Feature Specification: Desktop GUI Wrapper (Exploratory)

**Feature Branch**: `017-gui-desktop-app`  
**Created**: 2025-12-28  
**Status**: Exploratory / Parking Spec  
**Priority**: ⚪ Very Low (Future Exploration)  
**Input**: User description: "Optional desktop GUI for visual stack management"

## Product Contract *(mandatory)*

If a desktop GUI is ever implemented, it MUST be a **thin wrapper** over the existing CLI/TUI behaviour.

- The CLI remains the **canonical interface** for stack operations and configuration.
- The GUI MUST NOT duplicate or re-implement stack orchestration, environment detection, config rules, or Docker/Compose logic.
- The GUI SHOULD invoke the same underlying commands/APIs used by the CLI/TUI (or call a shared Go library used by both).
- The GUI MUST use the same project configuration files and state rules (see specs 005–011).

This constraint prevents multi-version drift (CLI vs TUI vs GUI) and keeps maintenance manageable.

### Implementation preference

- The project SHOULD prefer a Go-based UI framework where feasible (to keep a single-language toolchain).
- Acceptable future options MAY include Go-native GUI frameworks or a minimal webview approach.

## Scope *(exploratory)*

- A GUI wrapper may provide a visual front-end for:
  - selecting a project directory
  - starting/stopping the stack
  - showing status and key URLs
  - viewing logs
  - toggling optional services (via existing config + restart)

## Non-goals *(mandatory)*

- The GUI is NOT a replacement for the CLI/TUI.
- The GUI does NOT introduce new behaviour unavailable via CLI/TUI.
- The GUI does NOT implement installers, templates, or deployment automation.
- The GUI does NOT create an alternative configuration system or store secrets.

## Capability Tiers *(not a delivery plan)*

These tiers describe potential scope if a GUI is ever pursued.

### Tier A: Core Controls

- Project picker (choose working directory)
- Start / Stop / Restart controls
- Status indicator (Running / Stopped / Starting)
- Display key URLs (Website, phpMyAdmin, DB port)

### Tier B: Observability

- Live logs viewer (tail selected service logs)
- Quick access to open URLs in browser
- Surface actionable errors (Docker not running, port conflicts) using the same messages as CLI/TUI

### Tier C: Convenience

- Optional service toggles (mapped to spec 007 profiles, applied via `.20i-config.yml` + restart)
- “Open terminal” shortcut (launches external terminal in project directory)

## User Scenarios *(illustrative only)*

1. **Start/Stop**: User opens GUI, selects a project, clicks Start, sees Running.
2. **View logs**: User clicks Logs, selects nginx, sees tail output.
3. **Enable service**: User toggles Redis, GUI writes config (same keys as CLI), restarts stack.

## Edge Cases *(exploratory)*

- Docker not installed or daemon not running
- Port conflicts on start
- Multiple projects open (should default to one active project)
- Very large log volumes

## Requirements *(lightweight, exploratory)*

### Functional Requirements

- **FR-001**: If implemented, GUI SHOULD support macOS and Windows.
- **FR-002**: GUI MUST delegate stack operations to the canonical CLI/TUI core.
- **FR-003**: GUI SHOULD present the same status and error information as CLI/TUI.
- **FR-004**: GUI MAY provide logs viewing by calling existing log streaming endpoints/commands.

### Key Entities

- **GUI Wrapper**: Desktop UI layer that invokes canonical CLI/TUI behaviour
- **Project Context**: Selected working directory that drives stack operations
- **Action Delegation**: Mechanism for calling CLI commands or shared Go library functions

## Success Criteria *(exploratory)*

- **SC-001**: No duplicated orchestration logic exists in the GUI layer.
- **SC-002**: GUI actions produce identical results to running the equivalent CLI commands.
- **SC-003**: The GUI can lag behind CLI/TUI feature growth without breaking correctness.

## Assumptions

- The CLI/TUI remains stable and fully functional as the primary interface.
- Any GUI work would be optional and not required for core project value.

---</file>