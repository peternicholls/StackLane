# Feature Specification: 20i Stack Manager TUI

**Feature Branch**: `001-stack-manager-tui`  
**Created**: 2025-12-28  
**Updated**: 2025-12-28  
**Status**: Draft  
**Priority**: 🔴 Critical  
**Input**: User description: "Full-featured terminal UI for managing 20i stack: containers, config, logs, monitoring, projects"

---

## Overview

A professional terminal user interface (TUI) built with Bubble Tea framework to replace and enhance the existing 20i-gui bash script. This MVP replicates all 20i-gui functionality (start/stop/restart/status/logs/destroy) with a modern, keyboard-driven interface following best practices from lazydocker, lazygit, k9s, and gh-dash.

**Core Concept**: The TUI is a **project-aware web stack manager**, not a generic Docker container viewer. Users run it from their web project directory, and it manages the 20i stack for THAT project.

**Phase 3a Scope** (MVP - this spec):
- **PROJECT-AWARE STACK MANAGEMENT** (replicates 20i-gui):
  - Detect current directory as web project root
  - Validate `public_html/` folder exists (pre-flight check)
  - Offer template installation from `demo-site-folder/` if missing
  - Start/stop/restart stack for current project (`docker compose up/down`)
  - Set environment variables (`CODE_DIR`, `COMPOSE_PROJECT_NAME`)
  - Display stack status table with URLs, ports, CPU%
  - Stack destruction with confirmation - **PRIORITY 2**
  
- **THREE-PANEL LAYOUT**:
  - Left panel: Project info (name, path, stack status)
  - Right panel: Dynamic view (pre-flight → compose output → status table)
  - Bottom panel: Commands and status messages

**Phase 3b** (Container Lifecycle - after 3a):
  - Individual container start/stop/restart (for debugging)
  - Real-time CPU/memory monitoring per container
  - Container detail panel

**Phase 4+** (future specs):
- Multi-project browser (list all projects, switch between them)
- Log viewer with follow mode
- Project selection and switching
- Configuration editor (replaces manual .20i-local editing)

**Design Principles** (from research):
- **Project-first** - TUI operates on web project directories, not generic Docker containers
- **Panel-based layout** (not tabs) - see multiple contexts simultaneously
- **Component composition** - each view is a standalone Bubble Tea model
- **Keyboard-first** - vim bindings + arrow keys, max 3 keystrokes to any action
- **Real-time updates** - background goroutines, never block UI
- **Progressive disclosure** - show essentials first, details on demand
- **Context-aware help** - footer shows current view's shortcuts

---

## User Scenarios & Testing *(mandatory)*

### User Story 0 - Project Detection & Pre-flight (Priority: P0 - Core) 🎯 MVP FIRST

**Replaces**: 20i-gui project directory detection + public_html validation

As a developer, I want the TUI to detect my current directory as a web project and validate that `public_html/` exists, so that I can ensure my project is ready to run before starting the stack.

**Why this priority**: This is the ENTRY POINT - the TUI must understand the project context before any stack operations. Without this, we're just a generic Docker viewer.

**Independent Test**: Run TUI from `~/my-website/`, verify left panel shows project name and path, right panel shows pre-flight status (public_html exists or missing).

**Acceptance Scenarios**:

1. **Given** a directory with `public_html/`, **When** I launch the TUI, **Then** left panel shows "✅ my-website" with path, right panel shows "Ready to start stack"
2. **Given** a directory WITHOUT `public_html/`, **When** I launch the TUI, **Then** left panel shows "⚠️ my-website", right panel shows "Missing public_html/" with option to create from template
3. **Given** missing `public_html/`, **When** I press `T` (create template), **Then** `public_html/` is created from `demo-site-folder/` template
4. **Given** the TUI, **When** viewing left panel, **Then** it shows: project name, directory path, stack status (Not Running / Running / Starting)
5. **Given** a non-project directory (no `public_html/` possible), **When** I launch TUI, **Then** show clear error with instructions

---

### User Story 1 - Stack Lifecycle (Priority: P0 - Core) 🎯 MVP

**Replaces**: 20i-gui "Start Stack", "Stop Stack", "Restart Stack" commands

As a developer, I want to start, stop, and restart the entire 20i stack for my current project so that I can control my development environment with proper `CODE_DIR` mounting.

**Why this priority**: Stack lifecycle is the PRIMARY use case - getting the web project running with proper environment variables. This is what 20i-gui does.

**Independent Test**: Press `S` to start stack, verify compose output streams in right panel, then status table appears showing all 4 services with URLs.

**Acceptance Scenarios**:

1. **Given** a project with `public_html/` and stack NOT running, **When** I press `S`, **Then** `docker compose up -d` runs with `CODE_DIR=$(pwd)` and `COMPOSE_PROJECT_NAME={project-name}`
2. **Given** stack starting, **When** compose runs, **Then** right panel shows live compose output (pulling images, starting containers)
3. **Given** stack running, **When** compose completes, **Then** right panel shows status table: container names, status, image, URLs, CPU%
4. **Given** stack running, **When** I press `T` (stop), **Then** `docker compose down` runs and stack status updates to "Not Running"
5. **Given** stack running, **When** I press `R`, **Then** `docker compose restart` runs with feedback

---

### User Story 2 - Stack Status Table (Priority: P0 - Core)

**Replaces**: 20i-gui status view + Docker Desktop container table

As a developer, I want to see a status table of all running stack services with URLs and resource usage so I can access my site and monitor health.

**Why this priority**: After starting the stack, users need to know the access URLs and verify services are healthy.

**Independent Test**: Start stack, verify table shows: nginx (http://localhost:80), apache, mariadb (localhost:3306), phpmyadmin (http://localhost:8081), with CPU% bars.

**Acceptance Scenarios**:

1. **Given** stack running, **When** viewing status table, **Then** shows: Service Name, Status (●/○), Image, URL/Port, CPU%
2. **Given** nginx service, **When** in table, **Then** URL shows `http://localhost:{HOST_PORT}` (clickable)
3. **Given** nginx URL, **When** I click it with mouse, **Then** URL opens in default web browser
4. **Given** phpmyadmin service, **When** in table, **Then** URL shows `http://localhost:{PMA_PORT}` (clickable)
5. **Given** mariadb service, **When** in table, **Then** Port shows `localhost:{MYSQL_PORT}` (not clickable)
6. **Given** any service, **When** viewing CPU%, **Then** shows block graph (▓░░░░░░░░░) with percentage
7. **Given** any URL in table, **When** I hover over it, **Then** text style indicates clickability (underlined/highlighted)

---

### User Story 3 - Destroy Stack (Priority: P0 - Core)

**Replaces**: 20i-gui "Destroy Stack" command (menu option 6)

As a developer, I want to destroy a stack (stop containers + remove volumes) so that I can clean up projects or reset database state.

**Why this priority**: One of the 6 core 20i-gui commands - needed for cleanup and fresh starts.

**Independent Test**: Press `D` (shift-d), verify confirmation modal shows "⚠️ This will REMOVE ALL VOLUMES and data", type "yes", verify stack destroyed.

**Acceptance Scenarios**:

1. **Given** stack running, **When** I press `D` (shift-d), **Then** first confirmation modal appears: "⚠️ Destroy stack? Type 'yes' to continue"
2. **Given** first confirmation, **When** I type "yes", **Then** second confirmation modal appears: "🔴 Are you SURE? Type 'destroy' to confirm (Step 2/2)"
3. **Given** second confirmation, **When** I type "destroy", **Then** `docker compose down -v` runs and all containers + volumes are removed
4. **Given** any confirmation stage, **When** I press `Esc`, **Then** operation cancels and all modals close
5. **Given** any confirmation stage, **When** I type incorrect text, **Then** error hint shows and modal remains open
6. **Given** stack destroyed, **When** operation completes, **Then** right panel shows "Stack destroyed" and status updates to "Not Running"

---

### Deferred Stories (Phase 3b and Phase 4+)

These features are intentionally excluded from Phase 3a MVP:

**Phase 3b - Container Lifecycle** (after project-aware MVP):
- **Individual Container Actions**: Start/stop/restart individual services (s/r keys)
- **Container Navigation**: Navigate service list with j/k keys
- **Container Detail Panel**: Show selected container's ports, image, uptime, volumes

**Phase 4+ - Multi-Project & Advanced Features**:
- **Project Switcher**: List all web projects, switch between them (p key)
- **Log Viewer**: View live container logs with follow mode (l key)
- **Configuration Editor**: Edit .20i-local, .env, stack-vars.yml files in TUI
- **phpMyAdmin Architecture Selection**: Choose ARM vs x86 image
- **Custom Port Selection**: Interactive port picker like 20i-gui

**Rationale**: Phase 3a focuses on the CORE workflow: detect project → validate → start stack → show status. Individual container management and multi-project features come after the foundation is solid.

---

### Edge Cases

- What happens when Docker daemon is not running? (show error screen with retry button)
- How does the TUI handle terminal resize events? (SIGWINCH - recalculate layout, minimum 80x24)
- What happens when a stack operation fails? (show inline error in bottom panel, maintain UI state)
- What happens if user starts TUI from non-project directory? (show error: "No public_html/ found. Press T to create from template.")
- What happens when stack is not running? (right panel shows "Stack not running. Press S to start.")
- What happens when `public_html/` is missing? (right panel shows warning with option to create from template)
- What happens if demo-site-folder template is missing? (show error with path to expected template location)
- How does project name sanitization work? (same as 20i-gui: lowercase, replace invalid chars with hyphens)

---

## Requirements *(mandatory)*

### Functional Requirements

**Project Detection & Pre-flight (Phase 3a)**
- **FR-001**: TUI MUST detect current working directory (`$PWD`) as project root on startup
- **FR-002**: TUI MUST derive project name from directory name, sanitized same as 20i-gui (lowercase, hyphens)
- **FR-003**: TUI MUST check for `public_html/` directory existence as pre-flight validation
- **FR-004**: TUI MUST display pre-flight status in right panel: "Ready" (✅) or "Missing public_html/" (⚠️)
- **FR-005**: TUI MUST support `T` key to create `public_html/` from `demo-site-folder/` template
- **FR-006**: TUI MUST NOT allow stack start (`S`) if `public_html/` is missing (prevent accidental starts)

**Three-Panel Layout (Phase 3a)**
- **FR-010**: TUI MUST use three-panel layout: left (25%) | right (75%) | bottom (3 lines)
- **FR-011**: Left panel MUST show: project name, directory path, stack status indicator
- **FR-012**: Right panel MUST be dynamic: pre-flight status → compose output → status table
- **FR-013**: Bottom panel MUST show: available commands + status messages
- **FR-014**: TUI MUST use Lipgloss for ALL styling (no raw ANSI codes)
- **FR-015**: TUI MUST run in alternate screen mode with clean restoration on exit

**Stack Lifecycle (Phase 3a)**
- **FR-020**: `S` key MUST start stack: run `docker compose up -d` with `CODE_DIR=$(pwd)` and `COMPOSE_PROJECT_NAME={sanitized-name}`
- **FR-021**: Stack start MUST use 20i-stack docker-compose.yml (from `STACK_FILE` env or default location)
- **FR-022**: During stack start, right panel MUST show live compose output (streaming)
- **FR-023**: `T` key (when stack running) MUST stop stack: run `docker compose down`
- **FR-024**: `R` key MUST restart stack: run `docker compose restart`
- **FR-025**: All operations MUST provide feedback in bottom panel: "Starting stack..." → "✅ Stack running"
- **FR-026**: Failed operations MUST show user-friendly error: "❌ Port 80 in use" (not raw Docker error)

**Stack Status Table (Phase 3a)**
- **FR-030**: Status table MUST show columns: Service, Status (●/○), Image, URL/Port, CPU%
- **FR-031**: nginx row MUST show URL: `http://localhost:{HOST_PORT}` (default 80)
- **FR-032**: phpmyadmin row MUST show URL: `http://localhost:{PMA_PORT}` (default 8081)
- **FR-033**: mariadb row MUST show Port: `localhost:{MYSQL_PORT}` (default 3306)
- **FR-034**: CPU% MUST display as block graph: `▓▓▓░░░░░░░` (10 chars) with percentage
- **FR-035**: Status table MUST refresh automatically every 5s while stack is running
- **FR-036**: apache row MUST show "internal" (not directly accessible, proxied via nginx)

**Destroy Stack (Phase 3a)**
- **FR-040**: `D` key MUST show first confirmation modal with data loss warning
- **FR-041**: First confirmation MUST require typing "yes" to proceed to second confirmation
- **FR-041a**: Second confirmation MUST show "Are you SURE? Type 'destroy' to confirm" with red warning
- **FR-041b**: Only typing "destroy" exactly proceeds with operation
- **FR-042**: Destroy MUST run `docker compose down -v` (removes volumes)
- **FR-043**: After destroy, status MUST update to "Not Running" and status table clears
- **FR-043a**: Both Esc and incorrect input at any confirmation stage cancels operation

**Defaults (Phase 3a - simplified)**
- **FR-050**: phpMyAdmin image MUST default to ARM-native on Apple Silicon: `arm64v8/phpmyadmin:latest`
- **FR-050a**: phpMyAdmin image MUST fall back to standard image on x86: `phpmyadmin:latest`
- **FR-050b**: phpMyAdmin image selection MUST be overridable via `PHPMYADMIN_IMAGE` environment variable
- **FR-051**: Port selection MUST use auto-selection (find free port) like 20i-gui `find_free_port()` (deferred to Phase 4)
- **FR-052**: TUI MUST read `STACK_FILE` and `STACK_HOME` from environment if set
- **FR-052a**: TUI MUST validate STACK_FILE exists before executing docker compose commands

**Core TUI Framework**
- **FR-060**: TUI MUST use Bubble Tea v1.3.10+ framework with Elm Architecture (Model-Update-View pattern)
- **FR-061**: TUI MUST use Lipgloss for ALL styling (no raw ANSI codes)
- **FR-062**: TUI MUST handle terminal resize events (SIGWINCH - recalculate layout, minimum 80x24)
- **FR-063**: TUI MUST support mouse input for clicking URLs, selecting rows, and button interactions
- **FR-064**: URLs in status table MUST be clickable - clicking opens URL in default web browser
- **FR-065**: TUI MUST enable Bubble Tea mouse support via tea.WithMouseCellMotion() program option

**Error Handling**
- **FR-070**: When Docker daemon unreachable, MUST show error screen with "Docker not running" + retry hint
- **FR-071**: Terminal < 80x24 MUST show error: "Terminal too small (need 80x24, got {w}x{h})"
- **FR-072**: All Docker errors MUST be user-friendly (e.g., "port 80 in use" not "bind: address already in use")

---

### Component Architecture

**Following Bubble Tea + Research Best Practices**

```go
// Root model composes all views (Elm Architecture)
type RootModel struct {
    activeView   string            // "dashboard" | "logs" | "help" | "projects"
    dashboard    DashboardModel    // Main view (default)
    help         HelpModel         // Modal overlay
    projects     ProjectListModel  // Modal overlay
    dockerClient *docker.Client    // Shared Docker API wrapper
    err          error             // Global error state
}

func (m RootModel) Init() tea.Cmd {
    // Start background stats updater
    return tea.Batch(
        m.dashboard.Init(),
        tickEvery(2 * time.Second),  // Stats refresh
    )
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Global shortcuts
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "?":
            m.activeView = "help"
            return m, nil
        case "p":
            m.activeView = "projects"
            return m, m.projects.Load()
        }
        
        // Delegate to active view
        if m.activeView == "help" {
            updated, cmd := m.help.Update(msg)
            m.help = updated.(HelpModel)
            if m.help.closed {
                m.activeView = "dashboard"
            }
            return m, cmd
        }
        // ... delegate to other views
        
    case statsMsg:  // From background goroutine
        m.dashboard.UpdateStats(msg.stats)
        return m, tickEvery(2 * time.Second)  // Schedule next update
    }
    
    // Delegate to dashboard by default
    updated, cmd := m.dashboard.Update(msg)
    m.dashboard = updated.(DashboardModel)
    return m, cmd
}

func (m RootModel) View() string {
    // Modals overlay dashboard
    base := m.dashboard.View()
    if m.activeView == "help" {
        return overlayModal(base, m.help.View())
    }
    if m.activeView == "projects" {
        return overlayModal(base, m.projects.View())
    }
    return base
}

// Dashboard model (3-panel layout)
type DashboardModel struct {
    serviceList    list.Model       // Bubbles list (left panel)
    detailPanel    DetailPanel      // Custom component (right top)
    logPanel       *viewport.Model  // Bubbles viewport (right bottom, optional)
    logVisible     bool
    stats          map[string]Stats // Container stats cache
    width, height  int
}

func (m DashboardModel) View() string {
    leftPanel := m.renderServiceList()
    
    var rightPanel string
    if m.logVisible {
        rightPanel = lipgloss.JoinVertical(
            lipgloss.Top,
            m.renderDetail(),     // 30% height
            m.renderLogs(),       // 70% height
        )
    } else {
        rightPanel = m.renderDetail()  // 100% height
    }
    
    main := lipgloss.JoinHorizontal(
        lipgloss.Left,
        leftPanel,   // 25% width
        rightPanel,  // 75% width
    )
    
    footer := m.renderFooter()
    
    return lipgloss.JoinVertical(
        lipgloss.Top,
        main,
        footer,
    )
}
```

**File Structure**:
```
tui/
  main.go                    # Entry point, creates RootModel
  go.mod, go.sum             # Dependencies
  internal/
    app/
      root.go                # RootModel + routing logic
      messages.go            # Custom tea.Msg types (statsMsg, errorMsg)
    views/
      dashboard/
        dashboard.go         # DashboardModel
        service_list.go      # Service list panel
        detail.go            # Detail panel
        logs.go              # Log panel
      help/
        help.go              # HelpModel (modal)
      projects/
        projects.go          # ProjectListModel (modal)
    docker/
      client.go              # Docker SDK wrapper
      stats.go               # Background stats collector
      filters.go             # Container filtering by project
    ui/
      styles.go              # Lipgloss styles (colors, borders, layouts)
      components.go          # Reusable components (StatusIcon, ProgressBar)
      layout.go              # Panel sizing calculations
```

---

### Visual Design Specification

**Layout Dimensions** (Phase 3a - Three-Panel Layout):

```
Minimum: 80x24 characters
Recommended: 120x40 characters

Phase 3a MVP Layout (Project-Aware):
┌─ 20i Stack Manager ─────────────────────────────────────────────┐
│                                                                  │
├───────────────────┬──────────────────────────────────────────────┤
│ ✅ my-website     │ Stack Status (4 containers running)         │
│                   │                                              │
│ 📁 ~/projects/    │ ┌─────────────────────────────────────────┐ │
│    my-website     │ │ Service   Status  Image         URL/Port│ │
│                   │ ├─────────────────────────────────────────┤ │
│ Stack: Running    │ │ ● nginx   Running nginx:1.25   :80 →    │ │
│                   │ │            http://localhost:80          │ │
│                   │ │ ● apache  Running dev-apache   internal │ │
│                   │ │ ● mariadb Running mariadb:10.6 :3306    │ │
│ (25% width)       │ │ ● pma     Running phpmyadmin   :8081 →  │ │
│                   │ │            http://localhost:8081        │ │
│                   │ │                                         │ │
│                   │ │ CPU: ▓▓▓░░░░░░░ 28%                    │ │
│                   │ └─────────────────────────────────────────┘ │
│                   │                           (75% width)       │
├───────────────────┴──────────────────────────────────────────────┤
│ S:Start  T:Stop  R:Restart  D:Destroy  ?:Help  q:Quit           │
│ ✅ Stack running • 4 containers • http://localhost:80           │
└──────────────────────────────────────────────────────────────────┘

Pre-flight State (missing public_html/):
┌─ 20i Stack Manager ─────────────────────────────────────────────┐
├───────────────────┬──────────────────────────────────────────────┤
│ ⚠️ my-website     │                                              │
│                   │   ⚠️  Missing public_html/ directory         │
│ 📁 ~/projects/    │                                              │
│    my-website     │   Your web project needs a public_html/      │
│                   │   folder to serve files from.                │
│ Stack: Not Ready  │                                              │
│                   │   Press T to create from template            │
│                   │   (copies demo-site-folder/ structure)       │
│                   │                                              │
├───────────────────┴──────────────────────────────────────────────┤
│ T:Create Template  ?:Help  q:Quit                                │
│ ⚠️ Cannot start stack without public_html/                       │
└──────────────────────────────────────────────────────────────────┘

Starting State (compose output streaming):
┌─ 20i Stack Manager ─────────────────────────────────────────────┐
├───────────────────┬──────────────────────────────────────────────┤
│ 🔄 my-website     │ Starting Stack...                            │
│                   │                                              │
│ 📁 ~/projects/    │ [+] Running 4/4                              │
│    my-website     │  ⠿ Container my-website-nginx-1    Started  │
│                   │  ⠿ Container my-website-apache-1   Started  │
│ Stack: Starting   │  ⠿ Container my-website-mariadb-1  Started  │
│                   │  ⠿ Container my-website-pma-1      Started  │
│                   │                                              │
│                   │ Waiting for containers to be healthy...      │
│                   │                                              │
├───────────────────┴──────────────────────────────────────────────┤
│ Starting stack...                                                │
└──────────────────────────────────────────────────────────────────┘
```

**Color Palette** (Lipgloss named colors):

```go
var (
    // Status colors
    ColorRunning   = lipgloss.Color("10")  // Bright Green
    ColorStopped   = lipgloss.Color("8")   // Gray
    ColorRestart   = lipgloss.Color("11")  // Yellow
    ColorError     = lipgloss.Color("9")   // Bright Red
    
    // UI colors
    ColorAccent    = lipgloss.Color("12")  // Bright Blue (selected items)
    ColorBorder    = lipgloss.Color("8")   // Gray (panel borders)
    ColorText      = lipgloss.Color("7")   // White/Default
    ColorDim       = lipgloss.Color("8")   // Gray (secondary text)
    ColorHighlight = lipgloss.Color("13")  // Magenta (search matches)
    
    // Semantic colors
    ColorSuccess   = lipgloss.Color("10")  // Green
    ColorWarning   = lipgloss.Color("11")  // Yellow
    ColorDanger    = lipgloss.Color("9")   // Red
    ColorInfo      = lipgloss.Color("12")  // Blue
)
```

**Typography & Spacing**:
- **Panel borders**: Use lipgloss.Border (thin lines `│─┌┐└┘`)
- **Service list item**: `{icon} {name}` (e.g., `🟢 apache`)
- **Status icon + text**: Icon first, one space, then text
- **Progress bars**: Use Unicode blocks `▓` (filled) and `░` (empty), 10 chars wide
- **Padding**: 1 space inside panels, 0 between panels (border provides separation)
- **Line height**: Single-spaced (no blank lines within panels)

**Status Icons**:
- 🟢 Running (green circle)
- ⚪ Stopped (white/gray circle)  
- 🟡 Restarting (yellow circle)
- 🔴 Error/Unhealthy (red circle)
- ✅ Success (checkmark for actions)
- ❌ Error (X for failures)
- ⚠️  Warning (triangle for confirmations)

---

### Anti-Patterns to Avoid

**Based on research findings** - these cause "messy TUIs that don't work":

1. **❌ God Object Model**
   - DON'T: Put all state in one massive struct with 50+ fields
   - DO: Each view (Dashboard, Help, Projects) is a separate Bubble Tea model
   
2. **❌ Blocking I/O in Update**
   - DON'T: Call `docker.ListContainers()` directly in Update() method
   - DO: Use background goroutines + channels, send results as messages
   
3. **❌ Tab Navigation for Context Views**
   - DON'T: Use tabs to switch between Services/Logs/Stats (requires clicking through)
   - DO: Use panels so user sees service list + detail simultaneously
   
4. **❌ Inconsistent Key Bindings**
   - DON'T: `d` = delete in one view, `d` = download in another
   - DO: Establish global conventions (`s`=start/stop everywhere, `d`=delete everywhere)
   
5. **❌ No Visual Hierarchy**
   - DON'T: Everything same color/weight, no spacing, walls of text
   - DO: Use color for meaning (green=running), bold for emphasis, spacing for grouping
   
6. **❌ Hidden Features**
   - DON'T: Shortcuts only in help modal, no hints in UI
   - DO: Footer always shows top shortcuts for current view
   
7. **❌ Vague Errors**
   - DON'T: "Error: 1" or "operation failed"
   - DO: "❌ Failed to start apache: port 80 already in use"
   
8. **❌ Too Many Features in v1**
   - DON'T: Try to build config editor + monitoring + image management all at once
   - DO: MVP = dashboard + lifecycle + logs (match 20i-gui baseline)

---

### Non-Functional Requirements

- **NFR-001**: TUI MUST start and display dashboard in <2s on modern hardware (M1+, i5+)
- **NFR-002**: Panel focus switching MUST complete in <50ms (imperceptible lag)
- **NFR-003**: Container actions MUST show feedback within <300ms ("Starting..." message)
- **NFR-004**: Memory usage MUST stay <30MB with 4 services + 10k log lines per container
- **NFR-005**: TUI MUST handle terminal sizes from 80x24 (minimum) to 300x100 (practical maximum)
- **NFR-006**: Dashboard stats MUST refresh every 2s (configurable 1-5s range)
- **NFR-007**: Log buffer MUST NOT exceed 40MB total (4 containers × 10k lines × 1KB avg)
- **NFR-008**: Background goroutines MUST NOT block main UI thread (all I/O async)
- **NFR-009**: User settings and TUI artifacts MUST live under `~/.20istackman` (create on first run)
- **NFR-010**: All exported Go types, functions, and packages MUST have godoc comments
- **NFR-011**: Complex logic MUST have inline comments explaining "why" (not just "what")
- **NFR-012**: Main README.md MUST include TUI installation, usage, and keyboard shortcuts
- **NFR-013**: User documentation MUST exist in `/docs/tui/` for detailed guides and troubleshooting

### Keyboard Shortcuts

**Phase 3a - Project Stack Management** (MVP):
- `S` - Start stack (`docker compose up -d` with CODE_DIR)
- `T` - Stop stack / Create template (context-dependent)
  - If `public_html/` missing: Create from template
  - If stack running: Stop stack (`docker compose down`)
- `R` - Restart stack (`docker compose restart`)
- `D` - Destroy stack (double-confirmation required, `docker compose down -v`)
- `?` - Show help modal
- `q` or `Ctrl-C` - Quit TUI

**Phase 3b - Container Navigation** (deferred):
- `↑` or `k` - Move selection up in service list
- `↓` or `j` - Move selection down
- `s` - Start/stop selected individual container
- `r` - Restart selected individual container

**Phase 4+ - Multi-Project & Logs** (deferred):
- `p` - Open project switcher modal
- `l` - Open logs for selected service
- `f` - Toggle follow mode in logs
- `/` - Search/filter

**Mouse Support** (Phase 3a):
- **Click URL** - Opens URL in default web browser (nginx, phpMyAdmin)
- **Click status table row** - Selects container (Phase 3b)
- **Scroll wheel** - Navigate lists and output panels
- **Click modal buttons** - Confirm/cancel actions

**Design Rationale**:
- **Uppercase for stack operations** - S/T/R/D operate on entire stack
- **Lowercase for container operations** - s/r operate on selected container (Phase 3b)
- **Single-key actions** - no Ctrl/Alt combos for common tasks
- **`Esc` always cancels/closes** - modal dialogs, confirmations
- **Mouse optional** - all actions remain keyboard-accessible

---

### Key Entities

- **Service**: A logical component of the 20i stack (apache, mariadb, nginx, phpmyadmin); maps 1:1 to a Docker Compose service definition
- **Container**: Runtime Docker container instance; attributes: ID, name, image, status ("running"|"stopped"|"restarting"|"error"), state, ports, stats (CPU%, memory%, network I/O)
- **Project**: Directory containing docker-compose.yml; identified by directory name; may have .20i-local file for project-specific config
- **Log Stream**: Container stdout/stderr output; tail mode (last N lines) or follow mode (real-time); max 10,000 lines buffered
- **Stats**: Resource metrics for a container: CPU% (0-400% on 4-core), memory (bytes used/limit), network RX/TX bytes; refreshed every 2s

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

**Performance** (quantitative):
- **SC-001**: TUI starts and shows dashboard in <2s from launch on modern hardware
- **SC-002**: Container start/stop operations complete in <5s (Docker API time, not TUI latency)
- **SC-003**: Log follow mode displays new entries within <500ms of generation
- **SC-004**: Stats refresh cycle completes in <200ms (does not block UI)
- **SC-005**: Panel switching (Tab key) feels instant (<50ms perceived lag)

**Usability** (qualitative but measurable via user testing):
- **SC-006**: New users can navigate to logs view in <10 seconds without reading docs (follow visual hints in footer)
- **SC-007**: Common workflow (check status → restart service → view logs) achievable in ≤3 keystrokes per step
- **SC-008**: Help system (`?`) shows all available shortcuts grouped by context (no hidden features)
- **SC-009**: Error messages are actionable ("port 80 in use" not "bind error")

**Reliability** (functional correctness):
- **SC-010**: 100% parity with 20i-gui features (start, stop, restart, status, logs, destroy)
- **SC-011**: Zero data loss from failed operations (atomic Docker API calls, clear error states)
- **SC-012**: TUI handles Docker daemon restart gracefully (shows error, auto-retries, reconnects)
- **SC-013**: Terminal resize never crashes or corrupts display (re-layout on SIGWINCH)

**Documentation** (completeness):
- **SC-021**: All exported Go types have godoc comments (verified by `go doc`)
- **SC-022**: Main README.md includes TUI section with install, usage, shortcuts
- **SC-023**: User guide exists in /docs/tui/ with troubleshooting section
- **SC-024**: CHANGELOG.md documents all Phase 3a features
- **SC-025**: Complex logic has inline "why" comments (e.g., sanitization regex, state transitions)

**Compatibility**:
- **SC-014**: Works with existing .20i-local files (no migration needed)
- **SC-015**: Runs on macOS (primary), Linux (secondary) - matches 20i-gui platform support
- **SC-016**: Requires only Go 1.21+ and Docker (no additional dependencies beyond 20i-gui)

**Comparison to 20i-gui** (upgrade value):
- **SC-017**: Faster workflow: TUI dashboard shows all services at once (vs 20i-gui multi-step menus)
- **SC-018**: Live updates: Stats refresh every 2s (vs 20i-gui static status snapshots)
- **SC-019**: Richer logs: Follow mode + search + scroll (vs 20i-gui `docker compose logs -f` passthrough)
- **SC-020**: Keyboard-first: All actions <3 keystrokes (vs 20i-gui requires mouse for dialog selection)

---

## Assumptions

- Go 1.21+ is available (or install instructions provided in tui/README.md)
- Docker daemon is running and accessible via default socket (unix:///var/run/docker.sock or similar)
- User has permissions to manage Docker containers (same as 20i-gui requirements)
- Terminal supports ANSI colors (8-color minimum, 256-color recommended)
- Terminal supports alternate screen mode (standard since 1980s - xterm, iTerm2, Terminal.app, etc.)
- Minimum terminal size: 80x24 characters (enforced with error message if smaller)
- Docker Compose v2+ installed (docker compose, not docker-compose)
- Projects follow standard structure: docker-compose.yml at root, services named apache/mariadb/nginx/phpmyadmin
- User runs TUI from project directory (cd into project, then run `20i-tui`)

---

## Dependencies

**Go Modules** (specific versions for reproducible builds):
- **Bubble Tea** v1.3.10+ - TUI framework ([github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea))
- **Bubbles** v1.0.0+ - TUI components: list, viewport, textinput ([github.com/charmbracelet/bubbles](https://github.com/charmbracelet/bubbles))
- **Lipgloss** v1.0.0+ - Terminal styling ([github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss))
- **Docker SDK for Go** v27.0.0+ - Docker API client ([github.com/docker/docker/client](https://github.com/docker/docker))

**Optional** (not in MVP):
- Cobra - CLI flags/subcommands (if we add `20i-tui --project /path` later)
- YAML v3 - Config parsing (Phase 2 when we add config editor)

**Build Requirements**:
- Go 1.21 or later
- No C dependencies (pure Go, cross-platform)

**Runtime Requirements**:
- Docker daemon accessible (same as 20i-gui)
- Docker Compose v2+

---

## Files Affected

**New Files** (all under `tui/` directory):
```
tui/
  main.go                           # Entry point
  go.mod                            # Go module definition
  go.sum                            # Dependency checksums
  README.md                         # Build and usage instructions
  Makefile                          # Build targets (build, install, clean)
  internal/
    app/
      root.go                       # RootModel (top-level app state)
      messages.go                   # Custom tea.Msg types
    views/
      dashboard/
        dashboard.go                # DashboardModel
        service_list.go             # Service list panel (Bubbles list)
        detail.go                   # Detail panel
        logs.go                     # Log panel (Bubbles viewport)
      help/
        help.go                     # Help modal
      projects/
        projects.go                 # Project switcher modal
    docker/
      client.go                     # Docker SDK wrapper
      stats.go                      # Background stats collector
      filters.go                    # Project/container filtering
    ui/
      styles.go                     # Lipgloss styles (colors, borders)
      components.go                 # Reusable components (StatusIcon, ProgressBar)
      layout.go                     # Panel sizing functions
```

**New Documentation Files**:
```
docs/
  tui/
    user-guide.md                   # End-user guide (installation, usage, shortcuts)
    troubleshooting.md              # Common issues and solutions
    architecture.md                 # Developer guide (Bubble Tea, components)
```

**Modified Files**:
- `README.md` - Add "Terminal UI" section with install/usage/shortcuts
- `CHANGELOG.md` - Document Phase 3a features in [Unreleased] section
- `.gitignore` - Add `tui/20i-stack-manager` (compiled binary)

**Unchanged Files** (important for compatibility):
- `20i-gui` - Existing GUI still works, can coexist
- `docker-compose.yml` - No changes to stack definition
- `config/stack-vars.yml` - No changes (Phase 2 will extend)
- `.20i-local` - No changes to format

**Installation**:
- Binary installed to `/usr/local/bin/20i-tui` (or user's $GOBIN)
- Symlink `tui` → `20i-tui` for shorter command
- No config files needed for MVP (all defaults hardcoded)
