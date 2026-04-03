# Phase 3 Implementation Roadmap

**Quick Start Guide**: Step-by-step execution plan for implementing Container Lifecycle (MVP)

---

## Pre-Flight Checklist

Before starting implementation:

- [X] Phase 2 complete (Foundation ready)
- [X] All Phase 2 tests passing
- [ ] Read PHASE3-IMPLEMENTATION-NOTES.md
- [ ] Open QUICK-REFERENCE.md in browser/side panel
- [ ] Review docker-api.md contract
- [ ] Create feature branch: `git checkout -b feature/phase3-lifecycle`

---

## Implementation Blocks

### Block 1: Docker Client - Entity Layer (2-3 hours)

**Goal**: Define Container entity and Docker state mapping

**Tasks**:
1. [ ] T026: Add Container struct to internal/docker/client.go (minimal 6 fields)
2. [ ] T027: Add ContainerStatus enum (Running/Stopped/Restarting/Error)
3. [ ] T028: Create client_test.go with mapDockerState() tests
4. [ ] T029: Implement mapDockerState() helper function

**Checkpoint**: Run `go test ./internal/docker/...` → all tests pass

**Files Modified**:
- internal/docker/client.go (~30 lines added)
- internal/docker/client_test.go (~60 lines, NEW FILE)

---

### Block 2: Docker Client - List Containers (1-2 hours)

**Goal**: Fetch container list filtered by project

**Tasks**:
1. [ ] T030: Implement ListContainers(projectName) method
2. [ ] T031: Add unit tests to client_test.go

**Checkpoint**: Test with mock Docker data → returns correct containers

**Files Modified**:
- internal/docker/client.go (~50 lines added)
- internal/docker/client_test.go (~40 lines added)

**Test Command**:
```bash
go test ./internal/docker -run TestListContainers -v
```

---

### Block 3: Docker Client - Lifecycle Operations (2-3 hours)

**Goal**: Implement Start/Stop/Restart for individual containers

**Tasks**:
1. [ ] T032: Implement StartContainer(containerID) method
2. [ ] T033: Implement StopContainer(containerID, timeout) method
3. [ ] T034: Implement RestartContainer(containerID, timeout) method
4. [ ] T035: Add table-driven tests for all 3 methods

**Checkpoint**: Mock tests pass for success/error scenarios

**Files Modified**:
- internal/docker/client.go (~90 lines added)
- internal/docker/client_test.go (~80 lines added)

**Test Command**:
```bash
go test ./internal/docker -run TestStartContainer -v
go test ./internal/docker -run TestStopContainer -v
go test ./internal/docker -run TestRestartContainer -v
```

---

### Block 4: Docker Client - Compose Operations (2-3 hours)

**Goal**: Implement stack-wide operations (Stop All, Restart All, Destroy)

**Tasks**:
1. [ ] T036: Implement ComposeStop(projectPath) method
2. [ ] T037: Implement ComposeRestart(projectPath) method
3. [ ] T038: Implement ComposeDown(projectPath, removeVolumes) method
4. [ ] T039: Add unit tests for Compose operations

**Checkpoint**: Tests verify docker compose commands execute correctly

**Files Modified**:
- internal/docker/client.go (~100 lines added)
- internal/docker/client_test.go (~60 lines added)

**Note**: These use exec.Command to run `docker compose` CLI

---

### Block 5: Message Types Enhancement (30 min)

**Goal**: Document message type contracts

**Tasks**:
1. [ ] T040: Add ContainerAction enum comment in messages.go
2. [ ] T041: Add ComposeAction enum comment in messages.go
3. [ ] T042: Add composeActionResultMsg type (if not exists)
4. [ ] T043: Create messages_test.go with validation tests

**Checkpoint**: Message types documented and validated

**Files Modified**:
- internal/app/messages.go (~20 lines added)
- internal/app/messages_test.go (~40 lines, NEW FILE)

---

### Block 6: Dashboard Model - Foundation (2-3 hours)

**Goal**: Create DashboardModel struct and wire to RootModel

**Tasks**:
1. [ ] T044: Create internal/views/dashboard/dashboard.go with DashboardModel struct
2. [ ] T045: Implement Init() method (load container list)
3. [ ] T046: Implement containerListMsg handler in Update()
4. [ ] T050: Wire DashboardModel into RootModel in internal/app/root.go

**Checkpoint**: Can create DashboardModel, loads containers on init

**Files Created**:
- internal/views/dashboard/dashboard.go (~100 lines)

**Files Modified**:
- internal/app/root.go (~30 lines modified)

**Code Pattern**:
```go
type DashboardModel struct {
    containers    []Container
    selectedIndex int
    dockerClient  *docker.Client
    width, height int
    lastError     error
}

func (m DashboardModel) Init() tea.Cmd {
    return loadContainersCmd(m.dockerClient, "myproject")
}
```

---

### Block 7: Dashboard Rendering - Service List (1-2 hours)

**Goal**: Render simple service list with status icons

**Tasks**:
1. [ ] T047: Create internal/views/dashboard/service_list.go
2. [ ] T049: Implement DashboardModel.View() with 2-panel layout
3. [ ] T048: Create dashboard_test.go with rendering tests

**Checkpoint**: Run TUI → see service list with colored status icons

**Files Created**:
- internal/views/dashboard/service_list.go (~60 lines)
- internal/views/dashboard/dashboard_test.go (~50 lines)

**Files Modified**:
- internal/views/dashboard/dashboard.go (~40 lines for View() method)

**Visual Target**:
```
● apache
○ mariadb
● nginx
○ phpmyadmin
```

---

### Block 8: Navigation (1 hour)

**Goal**: Keyboard navigation (up/down/k/j)

**Tasks**:
1. [ ] T051: Add navigation key handlers to DashboardModel.Update()
2. [ ] T052: Add navigation tests to dashboard_test.go

**Checkpoint**: Arrow keys and vim keys change selected service

**Files Modified**:
- internal/views/dashboard/dashboard.go (~30 lines)
- internal/views/dashboard/dashboard_test.go (~40 lines)

**Key Bindings**:
- ↑ or k: Move selection up
- ↓ or j: Move selection down
- Wrap at top/bottom (circular navigation)

---

### Block 9: Container Actions - Commands (2-3 hours)

**Goal**: Wire up 's' and 'r' keys to Docker operations

**Tasks**:
1. [ ] T058-T060: Create containerActionCmd() function
2. [ ] T061: Implement containerActionResultMsg handler
3. [ ] T053: Add 's' key handler (toggle start/stop)
4. [ ] T054: Add 'r' key handler (restart)
5. [ ] T062: Add command function tests

**Checkpoint**: Pressing 's' starts/stops container, 'r' restarts it

**Files Modified**:
- internal/views/dashboard/dashboard.go (~100 lines)
- internal/views/dashboard/dashboard_test.go (~60 lines)

**Command Pattern**:
```go
func containerActionCmd(client *docker.Client, containerID string, action string) tea.Cmd {
    return func() tea.Msg {
        var err error
        switch action {
        case "start": err = client.StartContainer(containerID)
        case "stop": err = client.StopContainer(containerID, 10)
        case "restart": err = client.RestartContainer(containerID, 10)
        }
        return ContainerActionResultMsg{Success: err == nil, Error: err}
    }
}
```

---

### Block 10: Stack Actions (1-2 hours)

**Goal**: Wire up 'S' and 'R' keys for entire stack

**Tasks**:
1. [ ] T055: Add 'S' key handler (stop all)
2. [ ] T056: Add 'R' key handler (restart all)
3. [ ] T057: Add tests for compose action keys

**Checkpoint**: 'S' stops all containers, 'R' restarts all

**Files Modified**:
- internal/views/dashboard/dashboard.go (~60 lines)
- internal/views/dashboard/dashboard_test.go (~40 lines)

**Note**: 'D' (destroy) is Phase 4 (requires confirmation modal)

---

### Block 11: Status Messages & Error Handling (2-3 hours)

**Goal**: Show success/error feedback to user

**Tasks**:
1. [ ] T063: Add status message panel to dashboard layout
2. [ ] T064: Trigger containerListMsg refresh after successful action
3. [ ] T065: Implement formatDockerError() function
4. [ ] T066: Add error formatting tests
5. [ ] T067: Add footer with shortcuts

**Checkpoint**: See "✅ Container started" or "❌ Failed: port conflict" messages

**Files Modified**:
- internal/views/dashboard/dashboard.go (~80 lines)
- internal/views/dashboard/dashboard_test.go (~50 lines)

**Layout Update**:
```
┌──────────────────────────────────────┐
│ Services │ Status Messages           │
│──────────│──────────────────────────│
│ ● apache │ ✅ Container started      │
│          │                           │
├──────────┴──────────────────────────┤
│ s:start/stop r:restart S:stop-all   │
└──────────────────────────────────────┘
```

---

### Block 12: Polish & Detail Panel (2-3 hours)

**Goal**: Add detail panel and Tab key focus switching

**Tasks**:
1. [ ] T068: Implement Enter key → show detail panel for selected container
2. [ ] T069: Implement Tab key → cycle focus between panels

**Checkpoint**: Can view container details (image, status, uptime)

**Files Modified**:
- internal/views/dashboard/dashboard.go (~60 lines)

**Optional**: Can skip this block for MVP and add later

---

### Block 13: Integration & Acceptance Testing (2-3 hours)

**Goal**: Validate full workflow end-to-end

**Tasks**:
1. [ ] T070: Create tests/integration/lifecycle_test.go
2. [ ] T071: Manual acceptance testing (all 6 scenarios)
3. [ ] T072: Run `make test` → verify >85% coverage

**Checkpoint**: All tests pass, can manually verify all features work

**Test Scenarios**:
1. Start stopped container → verify status changes
2. Stop running container → verify status changes
3. Restart running container → verify brief restarting state
4. Stop all containers → verify all stop
5. Restart all containers → verify all restart
6. Error handling → verify user-friendly messages

---

## Time Estimates

**Sequential Implementation**: 25-29 hours total
- Block 1: 2-3 hours
- Block 2: 1-2 hours
- Block 3: 2-3 hours
- Block 4: 2-3 hours
- Block 5: 0.5 hours
- Block 6: 2-3 hours
- Block 7: 1-2 hours
- Block 8: 1 hour
- Block 9: 2-3 hours
- Block 10: 1-2 hours
- Block 11: 2-3 hours
- Block 12: 2-3 hours (optional)
- Block 13: 2-3 hours

**Parallel Implementation** (3 developers):
- Developer A: Blocks 1-4 (Docker client layer) → 8-12 hours
- Developer B: Blocks 6-8 (Dashboard foundation + rendering) → 4-7 hours
- Developer C: Blocks 5, 9-11 (Messages + actions + status) → 6-9 hours
- Everyone: Block 13 (Integration testing) → 2-3 hours
- **Total with parallelization**: 14-18 hours wall-clock time

---

## Daily Breakdown (Solo Developer)

### Day 1: Docker Client (6-8 hours)
- Morning: Blocks 1-2 (entities + list containers)
- Afternoon: Blocks 3-4 (lifecycle + compose operations)
- End of Day: All Docker client methods tested and working

### Day 2: Dashboard UI (6-8 hours)
- Morning: Blocks 5-7 (messages + dashboard foundation + rendering)
- Afternoon: Block 8 (navigation)
- End of Day: Can see service list and navigate

### Day 3: Actions & Polish (6-8 hours)
- Morning: Blocks 9-10 (container actions + stack actions)
- Afternoon: Block 11 (status messages + error handling)
- End of Day: Full lifecycle working

### Day 4: Testing & Polish (4-6 hours)
- Morning: Block 12 (detail panel - optional)
- Afternoon: Block 13 (integration tests + acceptance)
- End of Day: Phase 3 complete, ready for Phase 4

---

## Testing Strategy Per Block

**After Each Block**:
```bash
# Run unit tests for modified package
go test ./internal/docker -v
go test ./internal/views/dashboard -v

# Run all tests
make test

# Check coverage
make test-coverage
```

**Continuous Integration**:
- Commit after each completed block
- Push to feature branch
- Run CI pipeline (if configured)

---

## Troubleshooting Guide

### Docker Client Issues

**Problem**: "Cannot connect to Docker daemon"
- **Solution**: Start Docker Desktop, verify socket permissions

**Problem**: "API version negotiation failed"
- **Solution**: Update Docker to 20.10+, check client.NewClient() error handling

### Bubble Tea Issues

**Problem**: "UI not updating after action"
- **Solution**: Verify tea.Cmd returns message, check Update() handles message type

**Problem**: "Panic: index out of range"
- **Solution**: Clamp selectedIndex when container list changes

### Layout Issues

**Problem**: "Panels overlap or misaligned"
- **Solution**: Verify lipgloss.Width() calculations, handle tea.WindowSizeMsg

**Problem**: "Styles not applied"
- **Solution**: Check styles.go imported correctly, verify color palette

---

## Success Criteria

Phase 3 is **COMPLETE** when:

✅ All 47 tasks (T026-T072) checked off in tasks.md  
✅ `make test` passes with >85% coverage  
✅ All 6 manual acceptance scenarios verified  
✅ No blocking bugs or crashes  
✅ Error messages are user-friendly  
✅ Code follows Go best practices (gofmt, golint)

---

## Next Steps After Phase 3

Once Phase 3 complete:

1. **Commit & Push**: `git commit -am "feat: implement container lifecycle (Phase 3)"`
2. **Merge to Main**: Create PR, review, merge
3. **Tag Release**: `git tag v0.3.0-alpha`
4. **Update Tasks**: Mark all Phase 3 tasks as [X] in tasks.md
5. **Start Phase 4**: Implement Destroy Stack confirmation modal (T073-T089)

---

**Good luck! 🚀**

Refer back to PHASE3-IMPLEMENTATION-NOTES.md for detailed architectural decisions and patterns.
