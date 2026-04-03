# Quickstart: 20i Stack Manager TUI

**Feature**: 001-stack-manager-tui  
**Date**: 2025-12-28  
**Audience**: Developers implementing this feature

## 30-Second Overview

Building a terminal UI with Go + Bubble Tea to replace the 20i-gui bash script. Users get a modern, keyboard-driven dashboard for managing Docker containers, viewing logs, and monitoring resources - all without leaving the terminal.

**Tech Stack**: Go 1.21+ | Bubble Tea (TUI framework) | Docker SDK | Lipgloss (styling)

---

## Prerequisites

Before starting implementation:

1. **Go 1.21+** installed (`go version`)
2. **Docker Desktop** running (tested against Docker 24.0+)
3. **Existing 20i stack** to test against (clone repo, run `./20i-gui` to start stack)
4. **Terminal** with 256-color support (iTerm2, Terminal.app, or modern Linux terminal)
5. **Read the research**: `/runbooks/research/01-tui-excellence/findings.md` (critical patterns)

**Recommended Reading Order**:
1. Research findings (30 min) - understand TUI best practices
2. Feature spec `/specs/001-stack-manager-tui/spec.md` (20 min) - know what to build
3. Data model (10 min) - understand entities
4. This quickstart (5 min) - get started

---

## Development Workflow

### Step 1: Initialize Go Module (5 minutes)

```bash
# From repo root
cd /Users/peternicholls/docker/20i-stack
mkdir -p tui
cd tui

# Initialize Go module
go mod init github.com/peternicholls/20i-stack/tui

# Add dependencies
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/docker/docker@latest

# Create basic structure
mkdir -p internal/{app,views/{dashboard,help,projects},docker,ui}
touch main.go
touch internal/app/root.go
touch internal/views/dashboard/dashboard.go
```

**Verify**:
```bash
go mod tidy
# Should show bubbletea v1.3+, bubbles v1.0+, lipgloss v1.0+, docker v27.0+
```

---

### Step 2: Hello World TUI (10 minutes)

Create `main.go`:

```go
package main

import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    message string
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() string {
    return "20i Stack Manager TUI\n\nPress 'q' to quit.\n"
}

func main() {
    p := tea.NewProgram(model{message: "Hello"})
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Run**:
```bash
go run main.go
# Should show message, press 'q' to exit
```

✅ **Milestone**: Bubble Tea works, alternate screen mode activates

---

### Step 3: Docker Client Wrapper (30 minutes)

Create `internal/docker/client.go`:

```go
package docker

import (
    "context"
    "time"
    
    "github.com/docker/docker/client"
)

type Client struct {
    cli *client.Client
    ctx context.Context
}

func NewClient(ctx context.Context) (*Client, error) {
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return nil, err
    }
    
    // Test connection
    _, err = cli.Ping(ctx)
    if err != nil {
        return nil, err
    }
    
    return &Client{cli: cli, ctx: ctx}, nil
}

// Add methods per docker-api.md contract
```

**Test Docker connection**:
```go
// In main.go
client, err := docker.NewClient(context.Background())
if err != nil {
    fmt.Println("Docker not running:", err)
    os.Exit(1)
}
fmt.Println("Docker connected ✓")
```

✅ **Milestone**: Can connect to Docker daemon

---

### Step 4: Dashboard View (2-3 hours)

Implement the 3-panel layout using research findings:

**internal/views/dashboard/dashboard.go**:
```go
package dashboard

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/lipgloss"
)

type Model struct {
    serviceList   list.Model
    width, height int
}

func New() Model {
    // Create service list with Bubbles list component
    items := []list.Item{
        // Add services: apache, mariadb, nginx, phpmyadmin
    }
    
    l := list.New(items, list.NewDefaultDelegate(), 0, 0)
    l.Title = "Services"
    
    return Model{
        serviceList: l,
    }
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.serviceList.SetSize(m.width/4, m.height-2)  // 25% width, minus header/footer
    }
    
    m.serviceList, cmd = m.serviceList.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    // Implement 3-panel layout per spec
    listPanel := m.serviceList.View()
    detailPanel := "Detail panel (placeholder)"
    footer := "Tab:panels  s:start/stop  r:restart  q:quit"
    
    main := lipgloss.JoinHorizontal(lipgloss.Left, listPanel, detailPanel)
    return lipgloss.JoinVertical(lipgloss.Top, main, footer)
}
```

✅ **Milestone**: See service list in left panel

---

### Step 5: Real Container Data (1 hour)

Wire up Docker client to populate service list:

```go
// Add to dashboard.go
func LoadContainers(client *docker.Client, projectName string) tea.Cmd {
    return func() tea.Msg {
        containers, err := client.ListContainers(projectName)
        return containerListMsg{containers: containers, err: err}
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case containerListMsg:
        // Update service list items from msg.containers
        items := containersToListItems(msg.containers)
        m.serviceList.SetItems(items)
    }
    // ...
}
```

✅ **Milestone**: See real containers (apache, mariadb, nginx, phpmyadmin)

---

### Step 6: Container Actions (2 hours)

Implement start/stop/restart:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "s":
            // Toggle start/stop selected container
            container := m.getSelectedContainer()
            if container.Status == Running {
                return m, stopContainerCmd(container.ID)
            } else {
                return m, startContainerCmd(container.ID)
            }
        case "r":
            // Restart selected container
            container := m.getSelectedContainer()
            return m, restartContainerCmd(container.ID)
        }
    
    case containerActionResultMsg:
        // Show success/error message
        if msg.success {
            m.statusText = "✅ Container " + msg.action.String()
        } else {
            m.statusText = "❌ Failed: " + msg.err.Error()
        }
        // Refresh container list
        return m, LoadContainers(m.client, m.projectName)
    }
    // ...
}
```

✅ **Milestone**: Can start/stop containers from TUI

---

### Step 7: Real-Time Stats (1-2 hours)

Add background stats updates:

```go
// Ticker message
type tickMsg time.Time

func tickEvery(d time.Duration) tea.Cmd {
    return tea.Tick(d, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func (m Model) Init() tea.Cmd {
    return tickEvery(2 * time.Second)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tickMsg:
        // Refresh stats for all containers
        return m, tea.Batch(
            refreshStatsCmd(m.containers),
            tickEvery(2 * time.Second),  // Schedule next tick
        )
    
    case statsMsg:
        // Update stats map
        m.stats[msg.containerID] = msg.stats
    }
    // ...
}
```

✅ **Milestone**: See CPU/memory update every 2 seconds

---

### Step 8: Log Viewer (2-3 hours)

Add log panel using Bubbles viewport:

```go
import "github.com/charmbracelet/bubbles/viewport"

type Model struct {
    // ... existing fields
    logPanel  viewport.Model
    logVisible bool
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "l":
            // Toggle log panel
            if m.logVisible {
                m.logVisible = false
            } else {
                m.logVisible = true
                m.logPanel = viewport.New(m.width*3/4, m.height/2)
                container := m.getSelectedContainer()
                return m, streamLogsCmd(container.ID)
            }
        }
    
    case logLineMsg:
        // Append to viewport
        m.logPanel.SetContent(m.logPanel.Content() + msg.line + "\n")
        m.logPanel.GotoBottom()
    }
    // ...
}
```

✅ **Milestone**: Press `l` to view logs, see live updates

---

## Testing Strategy

### Manual Testing Checklist

```bash
# 1. Start TUI with running stack
cd /path/to/project
/path/to/20i-tui

# 2. Verify dashboard shows 4 services
# 3. Press 's' on running service → should stop
# 4. Press 's' on stopped service → should start
# 5. Press 'r' on running service → should restart
# 6. Press 'l' → logs panel opens
# 7. In logs: press 'f' → follow mode, make web request, see new log
# 8. Press '?' → help modal appears
# 9. Resize terminal → layout adjusts
# 10. Press 'q' → exits cleanly
```

### Unit Tests

```bash
# Test Docker client wrapper
go test ./internal/docker -v

# Test UI components (with mock Docker client)
go test ./internal/views/dashboard -v
```

### Integration Test

```bash
# Start test stack
docker compose up -d

# Run TUI in test mode
go run main.go --test-mode

# Verify all operations work
# Clean up
docker compose down -v
```

---

## Build & Install

```bash
# Build binary
go build -o 20i-tui

# Install to system
sudo cp 20i-tui /usr/local/bin/

# Create symlink
sudo ln -s /usr/local/bin/20i-tui /usr/local/bin/tui

# Test
cd /path/to/20i-project
tui
```

---

## Common Issues & Solutions

| Issue | Cause | Solution |
|-------|-------|----------|
| "Docker daemon not running" | Docker Desktop not started | Start Docker Desktop, press 'r' in TUI |
| Terminal too small error | Terminal < 80x24 | Resize terminal window |
| Stats not updating | Background goroutine blocked | Check `tickEvery` is being called in Init |
| Logs not appearing | StreamLogs not called | Verify `streamLogsCmd` returns valid tea.Cmd |
| UI flickers | Re-rendering too often | Add dirty flag, only render on state change |
| Key presses ignored | Wrong panel focused | Implement panel focus switching (Tab key) |

---

## Development Tips

**From Research Findings**:

1. **Start small**: Get dashboard working before logs, logs before stats, etc.
2. **Use Bubbles components**: Don't reinvent `list` or `viewport` - they handle edge cases
3. **Test with real data**: Use actual running containers, not mocks initially
4. **Watch the research examples**: lazydocker source for Docker patterns, gh-dash for Bubble Tea patterns
5. **Lipgloss everything**: Never write raw ANSI codes - use Lipgloss styles

**Debugging**:
```go
// Log to file (TUI owns stdout)
f, _ := os.OpenFile("debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
fmt.Fprintf(f, "Debug: %+v\n", msg)
```

**Performance**:
- Use `go build -race` to detect race conditions (background goroutines)
- Profile with `go tool pprof` if stats updates slow
- Test with 10+ containers to ensure scales

---

## Next Steps After MVP

Once core dashboard + logs + stats working:

**Phase 2** (v1.1):
- Project switcher modal (detect multiple projects)
- Configuration editor (edit .20i-local in TUI)
- Help system improvements (searchable, categorized)

**Phase 3** (v1.2):
- Image management (list, pull, remove images)
- Extended YAML config support
- Custom commands/shortcuts

**Phase 4** (v2.0):
- Resource monitoring graphs (CPU/memory over time)
- Plugin system
- Multi-stack management

---

## Resources

**Documentation**:
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Bubbles: https://github.com/charmbracelet/bubbles
- Lipgloss: https://github.com/charmbracelet/lipgloss
- Docker SDK: https://pkg.go.dev/github.com/docker/docker/client

**Example Code**:
- lazydocker: https://github.com/jesseduffield/lazydocker
- gh-dash: https://github.com/dlvhdr/gh-dash
- Bubble Tea examples: https://github.com/charmbracelet/bubbletea/tree/master/examples

**This Project**:
- Research: `/runbooks/research/01-tui-excellence/findings.md`
- Spec: `/specs/001-stack-manager-tui/spec.md`
- Data Model: `/specs/001-stack-manager-tui/data-model.md`
- Contracts: `/specs/001-stack-manager-tui/contracts/`

---

## Estimated Time to MVP

| Phase | Hours | Description |
|-------|-------|-------------|
| Setup + Hello World | 1 | Go module, basic TUI |
| Docker client | 2 | Wrapper, list containers |
| Dashboard layout | 3 | 3-panel Lipgloss layout |
| Service list | 2 | Bubbles list with real data |
| Container actions | 3 | Start/stop/restart |
| Stats updates | 2 | Background goroutine, ticker |
| Log viewer | 3 | Viewport, streaming |
| Polish | 2 | Error handling, help, styling |
| **Total** | **18-20 hours** | Full MVP |

**Accelerators**:
- Copy patterns from gh-dash (Bubble Tea structure)
- Copy Docker patterns from lazydocker
- Use research findings as checklist

**Blockers**:
- Docker SDK learning curve (2-3 hours)
- Bubble Tea message flow understanding (1-2 hours)
- Layout calculations for responsive design (1 hour)

---

## Success Criteria

MVP is complete when:

✅ Dashboard shows 4 services with status (green/gray/red icons)  
✅ Can start/stop/restart individual services  
✅ Stats (CPU%, memory) update every 2s  
✅ Log viewer opens with 'l', shows last 100 lines  
✅ Follow mode ('f') streams new logs in real-time  
✅ Help modal ('?') shows all shortcuts  
✅ Terminal resize works (min 80x24)  
✅ Errors show user-friendly messages  
✅ `Ctrl-C` or 'q' exits cleanly  
✅ All 20i-gui features replicated (start/stop/restart/status/logs/destroy)

**Definition of Done**:
- All acceptance scenarios from spec pass
- Manual testing checklist complete
- No crashes or hangs in normal operation
- README.md updated with installation instructions
- CHANGELOG.md entry added
