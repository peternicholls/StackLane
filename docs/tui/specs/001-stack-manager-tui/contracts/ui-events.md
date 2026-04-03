# Contract: UI Events & Messages

**Feature**: 001-stack-manager-tui  
**Date**: 2025-12-28  
**Type**: Internal Message Protocol (Bubble Tea)

## Overview

This contract defines all custom message types (`tea.Msg`) used in the TUI for inter-component communication. Bubble Tea's Elm Architecture uses messages to communicate state changes, async operation results, and user actions.

---

## Message Type Categories

### 1. Docker State Messages

Messages sent from Docker client goroutines to UI components.

#### statsMsg

**Purpose**: Update container resource usage stats

**Fields**:
```go
type statsMsg struct {
    containerID   string
    stats         Stats
    err           error  // If stats fetch failed
}
```

**Sent By**: Background stats collector goroutine  
**Received By**: DashboardModel  
**Frequency**: Every 1-2 seconds per container

**Behavior**:
- Dashboard updates `stats` map with new values
- If `err != nil`, show warning icon next to container
- Triggers re-render of detail panel and service list

---

#### containerListMsg

**Purpose**: Update list of containers in current project

**Fields**:
```go
type containerListMsg struct {
    containers []Container
    err        error
}
```

**Sent By**: Background container list refresh goroutine  
**Received By**: DashboardModel  
**Frequency**: Every 2 seconds, or after user action (start/stop)

**Behavior**:
- Replace `DashboardModel.containers` with new list
- Preserve selected index if possible (match by container ID)
- If selected container no longer exists, select first item
- Update service list component

---

#### logLineMsg

**Purpose**: Append new log line to buffer

**Fields**:
```go
type logLineMsg struct {
    containerID string
    line        string
    timestamp   time.Time
}
```

**Sent By**: Log streaming goroutine  
**Received By**: LogPanel (within DashboardModel)  
**Frequency**: As logs arrive (0-1000/sec depending on container activity)

**Behavior**:
- Append `line` to ring buffer
- Update viewport content
- If following mode, auto-scroll to bottom
- Apply filter if active

---

### 2. User Action Messages

Messages triggered by user keyboard input.

#### containerActionMsg

**Purpose**: User initiated container action

**Fields**:
```go
type containerActionMsg struct {
    action      ContainerAction  // Start | Stop | Restart | Remove
    containerID string
}
```

**Sent By**: DashboardModel (in response to key press)  
**Received By**: Self (processed in Update method)  
**Trigger**: User presses `s`, `r`, or `d` key

**Behavior**:
- Show "Starting..." / "Stopping..." status
- Launch goroutine to call Docker API
- Return `containerActionResultMsg` when complete

---

#### containerActionResultMsg

**Purpose**: Result of container action

**Fields**:
```go
type containerActionResultMsg struct {
    action      ContainerAction
    containerID string
    success     bool
    err         error
}
```

**Sent By**: Container action goroutine  
**Received By**: DashboardModel  

**Behavior**:
- If `success=true`: Show "✅ Container started" for 3 seconds
- If `success=false`: Show "❌ Failed to start: {err}" (persist until dismissed)
- Trigger container list refresh

---

#### composeActionMsg

**Purpose**: Whole-stack action (stop all, restart all, destroy)

**Fields**:
```go
type composeActionMsg struct {
    action ComposeAction  // StopAll | RestartAll | Destroy
    confirmed bool         // True if user confirmed (for destructive actions)
}
```

**Sent By**: DashboardModel (Shift-S, Shift-R, Shift-D keys)  
**Received By**: Self  

**Behavior**:
- If `action=Destroy` and `confirmed=false`: Show confirmation modal
- If confirmed: Launch goroutine to call Docker Compose API
- Return `composeActionResultMsg` when complete

---

### 3. View Navigation Messages

Messages for switching views and panels.

#### switchViewMsg

**Purpose**: Change active view (dashboard → help → projects)

**Fields**:
```go
type switchViewMsg struct {
    view ViewType  // Dashboard | Help | Projects
}
```

**Sent By**: RootModel (in response to `?`, `p`, `Esc` keys)  
**Received By**: RootModel  

**Behavior**:
- Update `RootModel.activeView`
- Initialize new view if needed (load project list, etc.)
- Re-render

---

#### focusPanelMsg

**Purpose**: Change focus between panels (list, detail, logs)

**Fields**:
```go
type focusPanelMsg struct {
    panel PanelType  // List | Detail | Logs
}
```

**Sent By**: DashboardModel (Tab/Shift-Tab keys)  
**Received By**: DashboardModel  

**Behavior**:
- Update `DashboardModel.focusedPanel`
- Change border color (focused=bright, unfocused=dim)
- Route subsequent keyboard input to focused panel

---

#### toggleLogPanelMsg

**Purpose**: Open/close log viewer

**Fields**:
```go
type toggleLogPanelMsg struct {
    containerID string  // Container to show logs for (empty = close)
}
```

**Sent By**: DashboardModel (`l` key to open, `Esc` to close)  
**Received By**: DashboardModel  

**Behavior**:
- If `containerID != ""`: Create LogPanel, start log stream, resize layout
- If `containerID == ""`: Close LogPanel, stop log stream, resize layout

---

### 4. Error Messages

#### errorMsg

**Purpose**: Display error to user

**Fields**:
```go
type errorMsg struct {
    title   string      // "Docker Connection Error", "Container Action Failed"
    message string      // Detailed error message
    level   ErrorLevel  // Warning | Error | Fatal
    action  string      // Optional: "Press 'r' to retry"
}
```

**Sent By**: Any component encountering error  
**Received By**: RootModel (displays error overlay)  

**Behavior**:
- If `level=Fatal`: Show error screen, block all input except quit
- If `level=Error`: Show inline error banner, allow dismissal
- If `level=Warning`: Show temporary status (3s), auto-dismiss

---

### 5. Lifecycle Messages

#### tickMsg

**Purpose**: Periodic timer for stats/container refresh

**Fields**:
```go
type tickMsg struct {
    time time.Time
}
```

**Sent By**: `tea.Tick()` command  
**Received By**: DashboardModel  
**Frequency**: Every 2 seconds

**Behavior**:
- Trigger background goroutine to refresh stats and container list
- Schedule next tick with `tea.Tick(2 * time.Second)`

---

#### shutdownMsg

**Purpose**: Clean shutdown of all goroutines

**Fields**: None

**Sent By**: RootModel (when user presses `q`)  
**Received By**: All components  

**Behavior**:
- Cancel all context (stops stats/log streams)
- Close all channels
- Return `tea.Quit` command

---

## Message Flow Examples

### Starting a Container

```
User presses 's' on selected container
  ↓
DashboardModel.Update receives tea.KeyMsg("s")
  ↓
Send containerActionMsg{action: Start, containerID: "a3f2d1b8"}
  ↓
Launch goroutine: docker.StartContainer("a3f2d1b8")
  ↓ (async)
Return containerActionResultMsg{success: true}
  ↓
DashboardModel shows "✅ Container started"
  ↓
Send containerListMsg to refresh list
  ↓
Dashboard updates with new status
```

### Opening Logs

```
User presses 'l'
  ↓
DashboardModel.Update receives tea.KeyMsg("l")
  ↓
Send toggleLogPanelMsg{containerID: "a3f2d1b8"}
  ↓
Create LogPanel, resize layout
  ↓
Launch goroutine: docker.StreamLogs("a3f2d1b8", follow=true)
  ↓ (continuous)
Receive logLineMsg every time new log appears
  ↓
Append to buffer, update viewport
```

### Stats Refresh Loop

```
App starts
  ↓
RootModel.Init returns tea.Tick(2 * time.Second)
  ↓ (2s later)
DashboardModel.Update receives tickMsg
  ↓
For each container: launch WatchStats goroutine
  ↓ (async, every 1s)
Receive statsMsg for each container
  ↓
Update stats map, re-render detail panel
  ↓
Schedule next tick
```

---

## Command Factories

Bubble Tea commands (`tea.Cmd`) are functions that return messages. Use these patterns:

### Async Docker Operation

```go
func startContainerCmd(client *docker.Client, containerID string) tea.Cmd {
    return func() tea.Msg {
        err := client.StartContainer(containerID)
        return containerActionResultMsg{
            action: Start,
            containerID: containerID,
            success: err == nil,
            err: err,
        }
    }
}
```

### Periodic Ticker

```go
func tickEvery(d time.Duration) tea.Cmd {
    return tea.Tick(d, func(t time.Time) tea.Msg {
        return tickMsg{time: t}
    })
}
```

### Channel Listener

```go
func listenToLogStream(ch <-chan string, containerID string) tea.Cmd {
    return func() tea.Msg {
        line, ok := <-ch
        if !ok {
            // Channel closed, stream ended
            return nil
        }
        return logLineMsg{
            containerID: containerID,
            line: line,
            timestamp: time.Now(),
        }
    }
}
```

---

## Message Routing

**RootModel** routes messages based on active view:

```go
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Global messages (handled by root)
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
        if msg.String() == "?" {
            m.activeView = "help"
            return m, nil
        }
    case errorMsg:
        // Show error overlay
        m.err = msg
        return m, nil
    }
    
    // Delegate to active view
    switch m.activeView {
    case "dashboard":
        updated, cmd := m.dashboard.Update(msg)
        m.dashboard = updated.(DashboardModel)
        return m, cmd
    case "help":
        updated, cmd := m.help.Update(msg)
        m.help = updated.(HelpModel)
        return m, cmd
    // ...
    }
}
```

---

## Testing Messages

### Unit Tests

```go
func TestContainerActionMsg(t *testing.T) {
    model := DashboardModel{...}
    
    msg := containerActionMsg{
        action: Start,
        containerID: "abc123",
    }
    
    updated, cmd := model.Update(msg)
    
    // Assert command was returned (async Docker call)
    assert.NotNil(t, cmd)
    
    // Execute command, get result message
    resultMsg := cmd().(containerActionResultMsg)
    
    // Assert result
    assert.True(t, resultMsg.success)
}
```

### Integration Tests

Use `tea.NewProgram()` with test mode:

```go
func TestLogStreaming(t *testing.T) {
    model := newDashboardModel(mockDockerClient)
    
    program := tea.NewProgram(model, tea.WithOutput(os.Stderr))
    
    // Send log panel toggle
    program.Send(toggleLogPanelMsg{containerID: "abc123"})
    
    // Send mock log lines
    program.Send(logLineMsg{line: "test log 1"})
    program.Send(logLineMsg{line: "test log 2"})
    
    // Verify model state
    finalModel := program.Wait().(DashboardModel)
    assert.Equal(t, 2, finalModel.logPanel.buffer.Size())
}
```

---

## Performance Considerations

**Message Rate Limits**:
- Stats updates: Max 1/sec per container (throttle if Docker sends faster)
- Log lines: No limit (Docker controls rate, ring buffer prevents memory growth)
- Container list refresh: Max 1 every 2s (debounce rapid refreshes)

**Batch Messages**:
```go
// DON'T: Send individual stats messages
for _, container := range containers {
    send(statsMsg{containerID: container.ID, ...})
}

// DO: Send batch message
send(statsBatchMsg{stats: map[string]Stats{...}})
```

**Debouncing**:
```go
// User types in search box rapidly
// Debounce filter updates to avoid excessive re-renders
var filterDebounceTimer *time.Timer

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.KeyMsg:
        m.filterText += msg.String()
        
        // Cancel previous timer
        if filterDebounceTimer != nil {
            filterDebounceTimer.Stop()
        }
        
        // Set new timer - only apply filter after 300ms of no typing
        filterDebounceTimer = time.AfterFunc(300*time.Millisecond, func() {
            send(applyFilterMsg{text: m.filterText})
        })
}
```
