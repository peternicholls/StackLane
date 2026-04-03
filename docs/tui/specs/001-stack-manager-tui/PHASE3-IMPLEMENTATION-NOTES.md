# Phase 3 Implementation Notes: Container Lifecycle (MVP)

**Feature**: 001-stack-manager-tui  
**Date**: 2025-12-28  
**Phase**: 3 - User Story 2 (Container Lifecycle)  
**Status**: Ready to Start  
**Developer**: Implementation prep complete

## Executive Summary

Phase 3 implements **User Story 2: Container Lifecycle** - the core MVP functionality that replicates all 20i-gui baseline capabilities (start/stop/restart containers, view status). This is the FIRST user-facing feature and the foundation for all subsequent enhancements.

**Scope**: Tasks T026-T072 (47 tasks total, including comprehensive testing)  
**Goal**: Replace 20i-gui bash script with professional TUI for container management  
**Success Criteria**: Can start/stop/restart individual containers and entire stack, verify status changes visually

---

## Architecture Decisions

### 1. Entity Design Strategy

**Decision**: Start with MINIMAL Container entity schema in Phase 3, extend in Phase 5

**Rationale**:
- Phase 3 only needs: ID, Name, Service, Image, Status, State (6 fields)
- Ports, CreatedAt, StartedAt are used ONLY in Phase 5 (dashboard monitoring enhancement)
- This avoids implementing unused fields and keeps Phase 3 focused on lifecycle operations

**Implementation**:
```go
// T026: Phase 3 minimal Container struct
type Container struct {
    ID      string          // Docker container ID (12-char)
    Name    string          // Full container name (e.g., "myproject-apache-1")
    Service string          // Compose service name (e.g., "apache")
    Image   string          // Image with tag (e.g., "php:8.2-apache")
    Status  ContainerStatus // Enum: Running, Stopped, Restarting, Error
    State   string          // Docker state detail (e.g., "Up 2 hours")
}

// T090: Phase 5 will ADD these fields via schema extension
// Ports      []PortMapping
// CreatedAt  time.Time
// StartedAt  time.Time
```

**Action Items**:
- [ ] T026: Create minimal Container struct (6 fields only)
- [ ] Document in code comments: "Extended in Phase 5 with Ports, CreatedAt, StartedAt"
- [ ] T090: When implementing Phase 5, add new fields WITHOUT breaking Phase 3 code

---

### 2. Dashboard Layout Strategy

**Decision**: Implement SIMPLIFIED 2-panel layout in Phase 3, expand to 3-panel in Phase 5

**Rationale**:
- Phase 3 needs: Service list (30%) + Status messages (70%) + Footer
- Detail panel with CPU/memory stats is Phase 5 feature (monitoring enhancement)
- Starting simple reduces complexity and gets MVP working faster

**Phase 3 Layout**:
```
┌─────────────────────────────────────────────────────────────┐
│ 20i Stack Manager - myproject                              │
├────────────────────┬────────────────────────────────────────┤
│ Services (30%)     │ Status Messages (70%)                  │
│                    │                                        │
│ ● apache           │ ✅ Container 'apache' started          │
│ ○ mariadb          │                                        │
│ ● nginx            │                                        │
│ ○ phpmyadmin       │                                        │
│                    │                                        │
│                    │                                        │
├────────────────────┴────────────────────────────────────────┤
│ s:start/stop r:restart S:stop-all R:restart-all D:destroy  │
└─────────────────────────────────────────────────────────────┘
```

**Phase 5 Enhancement** (not Phase 3):
- Add detail panel between service list and status messages
- Show CPU%, memory usage, ports, uptime in detail panel
- Requires Stats entity and WatchStats() implementation

**Action Items**:
- [ ] T044: Create DashboardModel with 2-panel layout (NOT 3-panel)
- [ ] T049: Implement View() with lipgloss.JoinHorizontal for service list + status panel
- [ ] Document: "Detail panel with stats added in Phase 5 (US1)"

---

### 3. Service List Rendering

**Decision**: Render SIMPLE list with status icons and names only (no stats)

**Phase 3 Rendering**:
```go
// T047: service_list.go - Simple rendering
func renderServiceList(containers []Container, selectedIndex int) string {
    var rows []string
    for i, c := range containers {
        icon := StatusIcon(c.Status)  // From T013: ●/○/✗/⚠
        style := RowStyle
        if i == selectedIndex {
            style = SelectedRowStyle
        }
        rows = append(rows, style.Render(fmt.Sprintf("%s %s", icon, c.Service)))
    }
    return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
```

**Phase 5 Enhancement** (add CPU/memory inline):
```go
// T100: Extended rendering with stats
func renderServiceList(containers []Container, stats map[string]Stats, selectedIndex int) string {
    // Add: fmt.Sprintf("%s %s  CPU:%0.1f%%  Mem:%s", icon, c.Service, stats[c.ID].CPUPercent, ...)
}
```

**Action Items**:
- [ ] T047: Implement simple rendering (icon + service name only)
- [ ] Use StatusIcon() from T013 (already implemented in Phase 2)
- [ ] Apply RowStyle/SelectedRowStyle from internal/ui/styles.go
- [ ] NO stats rendering in Phase 3 (save for T100)

---

### 4. Message Type Design

**Decision**: Use string-based action enums, NOT typed enums

**Rationale**:
- Existing messages.go already uses string actions (see ContainerActionMsg.Action = "start")
- Changing to typed enums would break Phase 2 tests
- String approach is simpler and matches Bubble Tea conventions

**Implementation**:
```go
// T040-T041: Keep existing string-based approach
type ContainerActionMsg struct {
    Action      string // "start" | "stop" | "restart"
    ContainerID string
}

type ComposeActionMsg struct {
    Action string // "stop" | "restart" | "down"
}

// Validate actions in handlers, not at type level
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ContainerActionMsg:
        switch msg.Action {
        case "start", "stop", "restart":
            return m, startContainerCmd(m.dockerClient, msg.ContainerID, msg.Action)
        default:
            return m, nil // Invalid action, ignore
        }
    }
}
```

**Action Items**:
- [ ] T040-T041: Document valid action values in code comments
- [ ] Add validation in Update() handlers
- [ ] NO new enum types needed (use existing message structs from Phase 2)

---

### 5. Command Function Pattern

**Decision**: One generic containerActionCmd, NOT separate functions per action

**Rationale**:
- Start, Stop, Restart have identical patterns (call Docker API, return result message)
- Generic function reduces code duplication (DRY principle)
- Action type passed as parameter

**Implementation**:
```go
// T058-T060: Single generic command function
func containerActionCmd(client *docker.Client, containerID string, action string) tea.Cmd {
    return func() tea.Msg {
        var err error
        
        switch action {
        case "start":
            err = client.StartContainer(containerID)
        case "stop":
            err = client.StopContainer(containerID, 10) // 10s timeout
        case "restart":
            err = client.RestartContainer(containerID, 10)
        default:
            return ContainerActionResultMsg{
                Success: false,
                Message: fmt.Sprintf("Invalid action: %s", action),
                Error:   fmt.Errorf("invalid action"),
            }
        }
        
        if err != nil {
            return ContainerActionResultMsg{
                Success: false,
                Message: fmt.Sprintf("Failed to %s container: %s", action, err),
                Error:   err,
            }
        }
        
        return ContainerActionResultMsg{
            Success: true,
            Message: fmt.Sprintf("✅ Container %sed successfully", action),
        }
    }
}
```

**Action Items**:
- [ ] T058-T060: Implement single containerActionCmd function
- [ ] T061: Handle ContainerActionResultMsg (show success/error in status panel)
- [ ] T062: Test with mock client (verify all 3 actions work)

---

### 6. Compose Operations Strategy

**Decision**: Implement ComposeStop, ComposeRestart, ComposeDown (NOT ComposeStart/Up)

**Rationale**:
- Phase 3 focus: STOPPING and DESTROYING stacks (matches 20i-gui)
- Starting entire stack is NOT a TUI feature (users run `docker compose up` manually)
- Keeps scope focused on management, not initialization

**Docker Client Methods** (see docker-api.md contract):
```go
// T036: Implement (replaces 'S' key action)
func (c *Client) ComposeStop(projectPath string) error

// T037: Implement (replaces 'R' key action)
func (c *Client) ComposeRestart(projectPath string) error

// T038: Implement (replaces 'D' key action with confirmation)
func (c *Client) ComposeDown(projectPath string, removeVolumes bool) error
```

**User Flow**:
```
Press 'S' → ComposeStop (graceful shutdown, containers remain)
Press 'R' → ComposeRestart (stop + start all)
Press 'D' → Show confirmation → ComposeDown(removeVolumes=true) → DESTROYS EVERYTHING
```

**Action Items**:
- [ ] T036-T038: Implement Compose operations in client.go
- [ ] T055: Wire 'S' key → ComposeStop
- [ ] T056: Wire 'R' key → ComposeRestart
- [ ] T073-T080: Wire 'D' key → Confirmation modal → ComposeDown (Phase 4)

---

### 7. Error Message Formatting

**Decision**: Implement user-friendly error mapping per docker-api.md contract

**Pattern**:
```go
// T065: Error message formatter
func formatDockerError(err error, action string, containerName string) string {
    if err == nil {
        return ""
    }
    
    // Check for known Docker error patterns
    errStr := err.Error()
    
    if strings.Contains(errStr, "port is already allocated") {
        port := extractPort(errStr) // Parse port number from error
        return fmt.Sprintf("❌ Port %s already in use. Stop conflicting service.", port)
    }
    
    if strings.Contains(errStr, "timeout") {
        return fmt.Sprintf("❌ Timeout: Container took too long to %s. Try again.", action)
    }
    
    if strings.Contains(errStr, "No such container") {
        return "❌ Container not found. It may have been removed. Press 'r' to refresh."
    }
    
    if strings.Contains(errStr, "permission denied") {
        return "❌ Permission denied. Add your user to the docker group."
    }
    
    // Generic fallback
    return fmt.Sprintf("❌ Failed to %s '%s': %s", action, containerName, err)
}
```

**Action Items**:
- [ ] T065: Implement formatDockerError in dashboard.go or new errors.go file
- [ ] T066: Add unit tests for error formatting (table-driven tests)
- [ ] Use in ContainerActionResultMsg handler (T061)

---

### 8. Testing Strategy

**3-Layer Testing Approach**:

**Layer 1: Unit Tests** (mock Docker client)
```go
// T031: Test ListContainers
func TestListContainers(t *testing.T) {
    mockClient := &MockDockerClient{
        Containers: []Container{
            {ID: "abc123", Service: "apache", Status: Running},
        },
    }
    
    containers, err := mockClient.ListContainers("myproject")
    assert.NoError(t, err)
    assert.Len(t, containers, 1)
}

// T035: Test Start/Stop/Restart
func TestStartContainer(t *testing.T) {
    // Table-driven test for all actions
}
```

**Layer 2: Component Tests** (Bubble Tea test program)
```go
// T048: Test DashboardModel
func TestDashboardNavigation(t *testing.T) {
    model := NewDashboardModel(mockClient)
    
    // Test up/down keys
    updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
    assert.Equal(t, 1, updated.(DashboardModel).selectedIndex)
}
```

**Layer 3: Integration Tests** (full workflow)
```go
// T070: Test lifecycle workflow
func TestContainerLifecycle(t *testing.T) {
    // Start container, verify status changes, stop container, verify status
}
```

**Action Items**:
- [ ] T028, T031, T035, T039: Unit tests for Docker client methods
- [ ] T043: Unit tests for message type creation
- [ ] T048, T052, T057, T062, T066: Component tests for dashboard
- [ ] T070: Integration test for full lifecycle
- [ ] T072: Run `make test` to verify >85% coverage

---

## Implementation Order

### Critical Path (must be sequential)

1. **T026-T027**: Container entity + ContainerStatus enum (foundation for everything)
2. **T028-T029**: mapDockerState() helper + tests (required by ListContainers)
3. **T030-T031**: ListContainers() + tests (required by dashboard)
4. **T044-T046**: DashboardModel struct + Init + containerListMsg handler (wire up list)
5. **T047-T049**: Service list rendering + View() method (make it visible)
6. **T050**: Wire into RootModel (make it runnable)
7. **T051-T052**: Navigation (up/down/k/j keys) + tests

**CHECKPOINT 1**: At this point, you can run TUI and see service list with navigation ✅

8. **T032-T035**: Docker lifecycle methods (Start/Stop/Restart) + tests
9. **T040-T043**: Enhance message types + tests
10. **T058-T062**: Command functions + result handlers + tests
11. **T053-T054**: Wire 's' and 'r' keys to actions

**CHECKPOINT 2**: Can start/stop/restart individual containers ✅

12. **T036-T039**: Compose operations + tests
13. **T055-T056**: Wire 'S' and 'R' keys for stack operations
14. **T063-T067**: Status messages, error formatting, footer

**CHECKPOINT 3**: Full lifecycle working, all 20i-gui baseline replicated ✅

15. **T068-T069**: Detail panel + Tab focus (nice-to-have)
16. **T070-T072**: Integration tests + final validation

---

### Parallel Opportunities

**After T026-T027 complete** (entities ready):
- **Track A**: T030-T031 (ListContainers)
- **Track B**: T032-T035 (Start/Stop/Restart) - can run in parallel with Track A

**After T044-T050 complete** (dashboard wired):
- **Track A**: T051-T052 (navigation)
- **Track B**: T040-T043 (message enhancements) - independent
- **Track C**: T047 service_list.go rendering - independent

**After T030-T035 complete** (all Docker methods ready):
- **Track A**: T058-T062 (container commands)
- **Track B**: T036-T039 (compose operations) - different file
- **Track C**: T063-T067 (UI polish) - different file

---

## Key Files to Create/Modify

### Files to CREATE (new in Phase 3)

1. **internal/docker/client.go extensions**
   - T026: Container struct
   - T027: ContainerStatus enum
   - T029: mapDockerState() helper
   - T030: ListContainers() method
   - T032-T034: Start/Stop/Restart methods
   - T036-T038: Compose operations

2. **internal/views/dashboard/dashboard.go**
   - T044: DashboardModel struct
   - T045: Init() method
   - T046: containerListMsg handler
   - T049: View() method
   - T051: Navigation key handlers
   - T053-T056: Action key handlers
   - T058-T060: Command functions
   - T061: Result message handlers

3. **internal/views/dashboard/service_list.go**
   - T047: renderServiceList() function

4. **internal/ui/components.go extensions**
   - T063: Status message rendering helpers

5. **Test files**
   - internal/docker/client_test.go (T028, T031, T035, T039)
   - internal/app/messages_test.go (T043)
   - internal/views/dashboard/dashboard_test.go (T048, T052, T057, T062, T066)
   - tests/integration/lifecycle_test.go (T070)

### Files to MODIFY (existing from Phase 2)

1. **internal/app/messages.go**
   - Already has ContainerActionMsg, ComposeActionMsg (from Phase 2)
   - T040-T041: Document valid action values in comments
   - T042: Add composeActionResultMsg type (NOT in existing file!)

2. **internal/app/root.go**
   - T050: Wire DashboardModel into RootModel
   - Update Update() to delegate to dashboard
   - Update View() to render dashboard

3. **tui/main.go**
   - NO CHANGES needed (already creates RootModel from Phase 2)

---

## Testing Checklist

### Unit Tests (15 tests total)

- [ ] T028: mapDockerState() - all Docker states → ContainerStatus
- [ ] T031: ListContainers() - success, empty results, errors
- [ ] T035: Start/Stop/Restart - success, container not found, timeout
- [ ] T039: Compose operations - success, invalid path, permission errors
- [ ] T043: Message type creation and field validation

### Component Tests (8 tests total)

- [ ] T048: DashboardModel Init/Update/View rendering
- [ ] T052: Navigation keys (up/down/k/j) update selectedIndex
- [ ] T057: Action keys ('s', 'r') send correct messages
- [ ] T062: Command functions send correct result messages
- [ ] T066: Error message formatting

### Integration Tests (1 comprehensive test)

- [ ] T070: Full lifecycle workflow (start → verify status → stop → verify status)

### Manual Acceptance Tests

Per User Story 2 acceptance criteria in spec.md:

- [ ] T071: Open TUI, see 4 services (apache, mariadb, nginx, phpmyadmin)
- [ ] Press 's' on stopped apache → verify it starts (icon changes ○ → ●)
- [ ] Press 's' on running apache → verify it stops (icon changes ● → ○)
- [ ] Press 'r' on running service → verify it restarts (icon briefly shows ⚠ then ●)
- [ ] Press 'S' → verify all services stop (all icons become ○)
- [ ] Press 'R' → verify all services restart

### Coverage Target

- [ ] T072: Run `make test-coverage` → verify >85% coverage for Phase 3 code

---

## Risk Assessment

### High Risk Items

1. **Docker Client Integration** (T030-T038)
   - **Risk**: Docker SDK API changes, permission errors, platform differences
   - **Mitigation**: Comprehensive error handling per docker-api.md contract, table-driven tests

2. **Message Handling Complexity** (T058-T061)
   - **Risk**: Race conditions, missed messages, state desync
   - **Mitigation**: Follow Bubble Tea patterns strictly (no blocking in Update), use mock client for testing

3. **Layout Rendering** (T049)
   - **Risk**: Terminal size issues, overflow, style conflicts
   - **Mitigation**: Test with various terminal sizes (80x24 minimum), use lipgloss measurement functions

### Medium Risk Items

1. **Navigation State** (T051-T052)
   - **Risk**: Selected index out of bounds when list changes
   - **Mitigation**: Clamp selectedIndex in containerListMsg handler

2. **Error Message UX** (T065-T066)
   - **Risk**: Cryptic Docker errors confuse users
   - **Mitigation**: Comprehensive error mapping with actionable suggestions

### Low Risk Items

1. **Footer Shortcuts** (T067)
   - Simple static text rendering
2. **Status Messages** (T063)
   - Simple panel with success/error display

---

## Performance Targets (Phase 3)

- **Container List Refresh**: <500ms (T030)
- **Start/Stop/Restart API Call**: <5s (T032-T034)
- **Compose Operations**: <60s (T036-T038)
- **Panel Switching**: <50ms (already met in Phase 2)
- **Navigation**: <16ms (instant response to key presses)

---

## Success Metrics

**Phase 3 Complete When**:

✅ All 47 tasks checked off in tasks.md  
✅ Can run `make test` with all tests passing  
✅ >85% code coverage for Phase 3 modules  
✅ Manual acceptance tests pass (all 6 scenarios)  
✅ Can run TUI, navigate services, start/stop/restart containers  
✅ Error messages are user-friendly and actionable  
✅ Footer shows all available shortcuts  

**Ready for Phase 4** (Destroy Stack):
- Dashboard layout established
- Message handling patterns proven
- Docker client tested and reliable

---

## Reference Materials

Keep these open while implementing:

1. **Primary References**:
   - `/runbooks/research/QUICK-REFERENCE.md` - Copy-paste patterns for common tasks
   - `/runbooks/research/bubbletea-component-guide.md` - Component architecture
   - `/runbooks/research/lipgloss-styling-reference.md` - Styling patterns

2. **Contracts**:
   - `contracts/docker-api.md` - Docker client method signatures and error handling
   - `contracts/ui-events.md` - Message type definitions and flow patterns

3. **Data Model**:
   - `data-model.md` - Entity schemas, relationships, state transitions

4. **Tasks**:
   - `tasks.md` - Full task breakdown with file paths and references

---

## Notes to Future Self

### What Worked Well in Phase 2

- Table-driven tests for Docker client (use same pattern in T028, T031, T035)
- Lipgloss color palette centralized in styles.go (reuse in Phase 3)
- Mock Docker client for testing (extend for Phase 3 methods)

### What to Avoid

- **DON'T** hardcode panel widths → use tea.WindowSizeMsg and calculate dynamically
- **DON'T** block in Update() → all Docker calls must be in tea.Cmd goroutines
- **DON'T** use raw ANSI codes → always use Lipgloss
- **DON'T** implement stats/detail panel in Phase 3 → that's Phase 5

### Implementation Tips

1. **Start Small**: Get ListContainers + basic rendering working first (T026-T050)
2. **Test Early**: Write tests alongside implementation (not after)
3. **Use Quick Reference**: Copy-paste patterns from QUICK-REFERENCE.md
4. **Check Tasks**: Mark off each task in tasks.md as completed
5. **Run Tests Often**: `make test` after each logical group (every 5-10 tasks)

---

## Final Checklist Before Starting

- [ ] Read this entire document
- [ ] Review tasks.md Phase 3 section (T026-T072)
- [ ] Open QUICK-REFERENCE.md in separate window
- [ ] Review docker-api.md contract
- [ ] Review ui-events.md contract
- [ ] Check Phase 2 tests pass (`make test`)
- [ ] Create feature branch: `git checkout -b feature/phase3-lifecycle`
- [ ] Ready to implement T026! 🚀

---

**Last Updated**: 2025-12-28  
**Status**: Ready for implementation  
**Next Step**: Implement T026 (Container entity struct)
