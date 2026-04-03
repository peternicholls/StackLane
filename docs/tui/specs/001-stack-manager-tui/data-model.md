# Data Model: 20i Stack Manager TUI

**Feature**: 001-stack-manager-tui  
**Date**: 2025-12-28  
**Phase**: 1 - Design & Contracts

## Overview

This document defines the core entities, their attributes, relationships, and state transitions for the TUI application. The data model follows Domain-Driven Design principles with clear boundaries between UI models (Bubble Tea), domain models (business logic), and external models (Docker SDK).

---

## Entity Definitions

### 1. Container

**Description**: Represents a running or stopped Docker container from the current project.

**Attributes**:
- `ID` (string, required): Docker container ID (12-char hash, e.g., "a3f2d1b8c9e4")
- `Name` (string, required): Human-readable container name (e.g., "myproject-apache-1")
- `Service` (string, required): Docker Compose service name (e.g., "apache", "mariadb")
- `Image` (string, required): Docker image with tag (e.g., "php:8.2-apache")
- `Status` (ContainerStatus, required): Current state (see State Transitions)
- `State` (string, optional): Docker state detail (e.g., "Up 2 hours", "Exited (0)")
- `Ports` ([]PortMapping, optional): Published ports (e.g., [{Host: "8080", Container: "80"}])
- `CreatedAt` (time.Time, required): Container creation timestamp
- `StartedAt` (time.Time, optional): Last start time (nil if never started)

**Relationships**:
- Belongs to one `Project`
- Has zero or one `Stats` (current metrics)
- Has zero or one `LogStream` (when logs viewer is open)

**Validation Rules**:
- `ID` must match regex `^[a-f0-9]{12}$`
- `Service` must be one of: "apache", "mariadb", "nginx", "phpmyadmin"
- `Status` transitions must follow state machine (see below)

---

### 2. ContainerStatus (Enum)

**Description**: Normalized container state for UI rendering.

**Values**:
- `Running` - Container is actively running
- `Stopped` - Container exists but is not running
- `Restarting` - Container is in the process of restarting
- `Error` - Container is in an unhealthy or crashed state

**Mapping from Docker SDK**:
```go
func mapDockerState(state string) ContainerStatus {
    switch strings.ToLower(state) {
    case "running":
        return Running
    case "restarting":
        return Restarting
    case "exited", "created", "paused", "dead":
        return Stopped
    default:
        return Error
    }
}
```

**UI Representation**:
| Status | Icon | Color | Example Use |
|--------|------|-------|-------------|
| Running | 🟢 | Green (#10) | Container is healthy |
| Stopped | ⚪ | Gray (#8) | User stopped or never started |
| Restarting | 🟡 | Yellow (#11) | Temporary state during restart |
| Error | 🔴 | Red (#9) | Failed to start or crashed |

---

### 3. Stats

**Description**: Real-time resource usage metrics for a container.

**Attributes**:
- `ContainerID` (string, required): Container ID this stats object belongs to
- `CPUPercent` (float64, required): CPU usage percentage (0-400 on 4-core, 0-100 per core)
- `MemoryUsed` (uint64, required): Memory used in bytes
- `MemoryLimit` (uint64, required): Memory limit in bytes (0 = unlimited)
- `MemoryPercent` (float64, required): Memory usage percentage (0-100)
- `NetworkRxBytes` (uint64, optional): Network received bytes (cumulative)
- `NetworkTxBytes` (uint64, optional): Network transmitted bytes (cumulative)
- `Timestamp` (time.Time, required): When stats were collected

**Relationships**:
- Belongs to one `Container`

**Validation Rules**:
- `CPUPercent` must be >= 0 (no upper limit - can exceed 100% on multi-core)
- `MemoryPercent` must be 0-100
- `MemoryUsed` <= `MemoryLimit` (unless limit is 0)
- `Timestamp` must not be in the future

**Calculation Details**:
```go
// CPU percentage calculation (from Docker SDK stats)
func calculateCPUPercent(stats types.StatsJSON) float64 {
    cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
    systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
    
    if systemDelta > 0 && cpuDelta > 0 {
        return (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
    }
    return 0
}

// Memory percentage
func calculateMemoryPercent(used, limit uint64) float64 {
    if limit == 0 {
        return 0  // No limit = can't calculate percentage
    }
    return (float64(used) / float64(limit)) * 100.0
}
```

---

### 4. PortMapping

**Description**: Published port mapping (container → host).

**Attributes**:
- `ContainerPort` (string, required): Port inside container (e.g., "80")
- `HostPort` (string, required): Port on host machine (e.g., "8080")
- `Protocol` (string, optional): "tcp" or "udp" (defaults to "tcp")

**Example**:
```go
// Docker Compose: ports: ["8080:80", "8443:443"]
PortMapping{ContainerPort: "80", HostPort: "8080", Protocol: "tcp"}
PortMapping{ContainerPort: "443", HostPort: "8443", Protocol: "tcp"}
```

---

### 5. Project

**Description**: A Docker Compose project containing one or more 20i stack services.

**Attributes**:
- `Name` (string, required): Project name (directory name, sanitized)
- `Path` (string, required): Absolute path to project directory
- `ComposeFile` (string, required): Path to docker-compose.yml (usually `Path/docker-compose.yml`)
- `IsActive` (bool, required): True if this is the currently displayed project
- `ContainerCount` (int, required): Number of running containers (for project switcher display)
- `Is20iStack` (bool, required): True if project has 20i stack services (apache+mariadb+nginx OR .20i-local exists)

**Relationships**:
- Has many `Container` (0..N)

**Validation Rules**:
- `Path` must be absolute path to existing directory
- `ComposeFile` must exist and be readable
- `Name` must match Docker Compose project name rules (lowercase, alphanumeric+hyphens, starts with letter/number)

**20i Stack Detection**:
```go
func is20iStack(projectPath string) bool {
    // Check for .20i-local file
    if _, err := os.Stat(filepath.Join(projectPath, ".20i-local")); err == nil {
        return true
    }
    
    // Parse docker-compose.yml and check for expected services
    compose := parseComposeFile(filepath.Join(projectPath, "docker-compose.yml"))
    hasApache := compose.Services["apache"] != nil
    hasMariaDB := compose.Services["mariadb"] != nil
    hasNginx := compose.Services["nginx"] != nil
    
    return hasApache && hasMariaDB && hasNginx
}
```

---

### 6. LogStream

**Description**: Buffered log output from a container's stdout/stderr.

**Attributes**:
- `ContainerID` (string, required): Container ID this stream belongs to
- `Buffer` ([]string, required): Ring buffer of log lines (max 10,000 lines)
- `Following` (bool, required): True if in follow mode (auto-scroll on new lines)
- `FilterText` (string, optional): Current search/filter text (empty = no filter)
- `Head` (int, required): Next write position in ring buffer (for circular buffer)
- `Size` (int, required): Current number of lines in buffer (0..10000)

**Relationships**:
- Belongs to one `Container`

**State Transitions**:
1. **Created** → `Following=false`, `Size=0`, load last 100 lines from Docker
2. **Following Enabled** → `Following=true`, append new lines as they arrive, auto-scroll
3. **Following Disabled** → `Following=false`, stop auto-scroll (user can manually scroll)
4. **Filter Applied** → `FilterText` set, viewport shows only matching lines
5. **Closed** → LogStream destroyed, buffer freed

**Memory Management**:
```go
const MaxLogLines = 10000

type LogStream struct {
    ContainerID string
    buffer      []string  // Fixed size array
    head        int
    size        int
    Following   bool
    FilterText  string
}

func (ls *LogStream) Append(line string) {
    ls.buffer[ls.head] = line
    ls.head = (ls.head + 1) % MaxLogLines
    if ls.size < MaxLogLines {
        ls.size++
    }
}

func (ls *LogStream) GetFilteredLines() []string {
    lines := ls.getAllLines()
    
    if ls.FilterText == "" {
        return lines
    }
    
    filtered := make([]string, 0, len(lines))
    for _, line := range lines {
        if strings.Contains(strings.ToLower(line), strings.ToLower(ls.FilterText)) {
            filtered = append(filtered, line)
        }
    }
    return filtered
}
```

---

## State Transitions

### Container Lifecycle

```
[Created/Stopped] ──(start)──> [Running] ──(stop)──> [Stopped]
        │                         │
        │                         │
        └──(restart)──> [Restarting] ──> [Running]
        
[Running] ──(crash)──> [Error] ──(manual start)──> [Running]
        
[Any State] ──(docker compose down -v)──> [Removed]
```

**Transition Rules**:
- `Stopped` → `Running`: User presses `s` on stopped container, Docker starts it
- `Running` → `Stopped`: User presses `s` on running container, Docker stops it
- `Running` → `Restarting` → `Running`: User presses `r`, Docker restarts container (stop + start)
- `Running` → `Error`: Container crashes, health check fails, or fails to start
- `Any` → `Removed`: User confirms destroy (`D`), runs `docker compose down -v`

**UI Feedback**:
- During transition: Show temporary status (e.g., "Starting..." for 300ms)
- On success: Show ✅ checkmark + message for 3s
- On failure: Show ❌ X + error message (persist until user dismisses with Esc)

---

### LogStream Lifecycle

```
[Closed] ──(press 'l')──> [Loading] ──(loaded 100 lines)──> [Open, Not Following]
                                                                    │
                                                                    │
        ┌───────────────────────────────────────────────────────────┘
        │
        ├──(press 'f')──> [Open, Following] ──(press 'f')──> [Open, Not Following]
        │                       │
        │                       └──(new log line)──> [Append + Auto-scroll]
        │
        ├──(press '/')──> [Open, Filter Mode] ──(type text)──> [Filtered View]
        │                       │
        │                       └──(press Esc)──> [Open, Not Following]
        │
        └──(press Esc/q)──> [Closed]
```

---

## Domain Model Relationships

```
Project (1) ──────< (N) Container
                         │
                         ├──< (0..1) Stats
                         └──< (0..1) LogStream

User ───interacts with──> DashboardModel ───displays──> Container[]
                               │
                               ├─────> DetailPanel ───shows──> Container (selected)
                               └─────> LogPanel ───renders──> LogStream
```

---

## UI Models (Bubble Tea)

### RootModel

**Purpose**: Top-level application state and view routing.

**Fields**:
```go
type RootModel struct {
    activeView   string            // "dashboard" | "help" | "projects"
    dashboard    DashboardModel
    help         HelpModel
    projects     ProjectListModel
    dockerClient *docker.Client
    currentProject *Project
    err          error             // Global error state
    width, height int
}
```

**Responsibilities**:
- Route key events to active view
- Handle global shortcuts (`?`, `q`, `p`)
- Coordinate view transitions (e.g., projects modal → dashboard)

---

### DashboardModel

**Purpose**: Main view - displays service list, detail panel, and optional log panel.

**Fields**:
```go
type DashboardModel struct {
    serviceList   list.Model           // Bubbles list component
    containers    []Container          // Data source for list
    selectedIndex int
    
    detailPanel   DetailPanel          // Custom component
    
    logPanel      *viewport.Model      // Bubbles viewport (nil when closed)
    logStream     *LogStream           // Current log buffer (nil when closed)
    logVisible    bool
    
    stats         map[string]Stats     // ContainerID → Stats cache
    
    width, height int
    listWidth, detailWidth int
    detailHeight, logHeight int
}
```

**Responsibilities**:
- Render 3-panel layout (list | detail | footer)
- Handle container operations (start/stop/restart)
- Manage log panel visibility and state
- Coordinate stats updates from background goroutine

---

### HelpModel

**Purpose**: Modal overlay showing context-aware keyboard shortcuts.

**Fields**:
```go
type HelpModel struct {
    context string   // "dashboard" | "logs" | "projects"
    closed  bool
}
```

**Shortcuts by Context**:
- **Dashboard**: `s`=start/stop, `r`=restart, `l`=logs, `S`=stop all, `R`=restart all, `D`=destroy
- **Logs**: `f`=follow, `/`=search, `↑↓`=scroll, `g/G`=top/bottom, `Esc`=close
- **Projects**: `↑↓`=navigate, `Enter`=switch, `Esc`=cancel

---

### ProjectListModel

**Purpose**: Modal overlay for switching between detected 20i projects.

**Fields**:
```go
type ProjectListModel struct {
    projects      []Project        // Detected projects
    selectedIndex int
    closed        bool
}
```

**Responsibilities**:
- Scan for docker-compose.yml files (current dir + up 2 levels)
- Filter for 20i stacks (apache+mariadb+nginx OR .20i-local)
- Display with current project marked, running container counts
- Switch project context on selection

---

## Validation Summary

| Entity | Required Fields | Validation Rules | Default Values |
|--------|----------------|------------------|----------------|
| Container | ID, Name, Service, Image, Status | ID regex, Service enum | CreatedAt=now |
| Stats | ContainerID, CPUPercent, Memory* | CPU>=0, MemPercent 0-100 | Network=0 |
| LogStream | ContainerID, Buffer | MaxLogLines=10000 | Following=false |
| Project | Name, Path, ComposeFile | Path exists, Name sanitized | ContainerCount=0 |
| PortMapping | ContainerPort, HostPort | - | Protocol="tcp" |

---

## Open Questions / Decisions Deferred

None - all entities align with MVP scope. Phase 2 features (config editor, image management) will introduce new entities in their own specs.
