# Phase 3 Architecture Decision Record (ADR)

**Feature**: 001-stack-manager-tui  
**Phase**: 3 - Container Lifecycle (MVP)  
**Date**: 2025-12-28  
**Status**: Prepared for Implementation

---

## ADR-001: Minimal Container Schema in Phase 3

**Decision**: Implement Container entity with only 6 fields in Phase 3, extend to 9 fields in Phase 5

**Context**:
- Phase 3 scope: Container lifecycle (start/stop/restart)
- Phase 5 scope: Dashboard monitoring enhancement (stats, ports, uptime)
- data-model.md defines full 9-field Container schema

**Options Considered**:

1. **Implement Full Schema Now** (9 fields)
   - ✅ Pro: Complete from start, no breaking changes later
   - ❌ Con: Implements unused fields (Ports, CreatedAt, StartedAt not used in Phase 3)
   - ❌ Con: Violates YAGNI (You Aren't Gonna Need It) principle

2. **Minimal Schema Now, Extend Later** (6 fields → 9 fields)
   - ✅ Pro: Focused on current requirements (lifecycle operations)
   - ✅ Pro: Faster initial implementation
   - ✅ Pro: Clear separation between lifecycle (Phase 3) and monitoring (Phase 5)
   - ❌ Con: Schema migration needed in Phase 5
   - ❌ Con: Tests might need updates

**Decision**: **Option 2** - Minimal schema now

**Phase 3 Schema**:
```go
type Container struct {
    ID      string          // Required for all Docker operations
    Name    string          // Display in service list
    Service string          // Display in service list (short name)
    Image   string          // Show in detail panel
    Status  ContainerStatus // Render status icon
    State   string          // Docker state detail (e.g., "Up 2 hours")
}
```

**Phase 5 Extensions**:
```go
// Add these fields in T090 (Phase 5):
Ports      []PortMapping // Show in detail panel
CreatedAt  time.Time     // Calculate uptime
StartedAt  time.Time     // Calculate uptime
```

**Migration Strategy**:
1. Phase 3: Implement with 6 fields
2. Document in code: "// Extended in Phase 5 with Ports, CreatedAt, StartedAt"
3. Phase 5: Add new fields to struct, update ListContainers() to populate them
4. No breaking changes: New fields are additive, existing code unaffected

**Consequences**:
- ✅ Faster Phase 3 implementation (focus on lifecycle, not monitoring)
- ✅ Clear feature boundaries (lifecycle vs. monitoring)
- ⚠️ Need to update ListContainers() in Phase 5 to populate new fields

**Status**: Approved - Implement T026 with 6 fields only

---

## ADR-002: Simplified 2-Panel Dashboard Layout in Phase 3

**Decision**: Implement 2-panel layout (service list + status messages) in Phase 3, expand to 3-panel in Phase 5

**Context**:
- Phase 3 focus: Container operations (start/stop/restart)
- Phase 5 focus: Real-time monitoring (CPU, memory, network stats)
- 3-panel layout (list | detail | status) described in plan.md for final version

**Options Considered**:

1. **Implement 3-Panel Layout Now**
   - ✅ Pro: Final layout structure from start
   - ❌ Con: Detail panel is empty in Phase 3 (stats not implemented until Phase 5)
   - ❌ Con: More complex layout code for no immediate benefit

2. **Start with 2-Panel, Expand to 3-Panel Later**
   - ✅ Pro: Simpler initial implementation
   - ✅ Pro: Detail panel only added when stats available (Phase 5)
   - ✅ Pro: Matches progressive enhancement philosophy
   - ❌ Con: Layout code needs refactoring in Phase 5

**Decision**: **Option 2** - 2-panel layout now

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
├────────────────────┴────────────────────────────────────────┤
│ s:start/stop r:restart S:stop-all R:restart-all           │
└─────────────────────────────────────────────────────────────┘
```

**Phase 5 Enhancement**:
```
┌───────────────────────────────────────────────────────────────┐
│ 20i Stack Manager - myproject                                │
├──────────┬──────────────────────┬──────────────────────────────┤
│ Services │ Container Details    │ Status Messages              │
│ (20%)    │ (40%)                │ (40%)                        │
│          │                      │                              │
│ ● apache │ Image: php:8.2       │ ✅ Container started         │
│          │ CPU: 12.3%  ░░▓▓▓░░  │                              │
│          │ Mem: 45.2%  ░░░▓▓▓▓  │                              │
│          │ Ports: 8080:80       │                              │
│          │ Uptime: 2h 34m       │                              │
├──────────┴──────────────────────┴──────────────────────────────┤
│ s:start/stop r:restart l:logs Tab:focus S/R/D:stack           │
└───────────────────────────────────────────────────────────────┘
```

**Migration Strategy**:
1. Phase 3: Implement 2-panel layout in T049
2. Phase 5: Refactor View() to split status panel into detail + status (T100)
3. Use same lipgloss.JoinHorizontal pattern, just add one more panel

**Consequences**:
- ✅ Simpler Phase 3 implementation (fewer layout calculations)
- ✅ Detail panel appears when it has content to show
- ⚠️ View() method needs refactoring in Phase 5 (low risk - pure rendering change)

**Status**: Approved - Implement T049 with 2-panel layout

---

## ADR-003: String-Based Action Enums vs. Typed Enums

**Decision**: Use string-based action values, NOT Go typed enums

**Context**:
- Existing messages.go (Phase 2) uses strings: `ContainerActionMsg.Action = "start"`
- Tasks T040-T041 suggest adding "action enums"
- Go doesn't have native enum support (uses const + iota pattern)

**Options Considered**:

1. **Typed Enums** (Go const pattern)
   ```go
   type ContainerAction int
   const (
       ActionStart ContainerAction = iota
       ActionStop
       ActionRestart
   )
   ```
   - ✅ Pro: Type safety, autocomplete in IDE
   - ❌ Con: Breaking change to existing message types
   - ❌ Con: Breaks Phase 2 tests
   - ❌ Con: More verbose code

2. **String-Based Actions** (current approach)
   ```go
   type ContainerActionMsg struct {
       Action      string // "start" | "stop" | "restart"
       ContainerID string
   }
   ```
   - ✅ Pro: Matches existing Phase 2 implementation
   - ✅ Pro: No breaking changes
   - ✅ Pro: Simpler, more idiomatic for message passing
   - ✅ Pro: Easier to serialize/log
   - ❌ Con: No compile-time validation (typos possible)

**Decision**: **Option 2** - String-based actions

**Mitigation for Lack of Type Safety**:
1. **Document valid values in comments**:
   ```go
   // ContainerActionMsg requests an operation on a container.
   // Valid actions: "start", "stop", "restart"
   type ContainerActionMsg struct {
       Action      string
       ContainerID string
   }
   ```

2. **Validate in handlers**:
   ```go
   func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       switch msg := msg.(type) {
       case ContainerActionMsg:
           switch msg.Action {
           case "start", "stop", "restart":
               return m, containerActionCmd(m.client, msg.ContainerID, msg.Action)
           default:
               // Invalid action - log error, show warning
               return m, nil
           }
       }
   }
   ```

3. **Add validation tests** (T043):
   ```go
   func TestInvalidAction(t *testing.T) {
       msg := ContainerActionMsg{Action: "invalid"}
       // Verify Update() handles gracefully (no panic)
   }
   ```

**Consequences**:
- ✅ No breaking changes to Phase 2
- ✅ Consistent with Bubble Tea conventions (messages often use strings)
- ⚠️ Developers must refer to comments for valid values
- ⚠️ Runtime validation required (covered by tests)

**Status**: Approved - Use strings, document valid values in T040-T041

---

## ADR-004: Generic containerActionCmd vs. Separate Functions

**Decision**: Implement ONE generic command function for all container actions

**Context**:
- Tasks T058-T060 list separate functions: startContainerCmd, stopContainerCmd, restartContainerCmd
- All three have identical patterns (call Docker API, return result message)

**Options Considered**:

1. **Separate Functions** (as described in tasks)
   ```go
   func startContainerCmd(...) tea.Cmd { ... }
   func stopContainerCmd(...) tea.Cmd { ... }
   func restartContainerCmd(...) tea.Cmd { ... }
   ```
   - ✅ Pro: Explicit naming
   - ❌ Con: Code duplication (3 nearly identical functions)
   - ❌ Con: More functions to test

2. **Generic Command Function**
   ```go
   func containerActionCmd(client *docker.Client, containerID, action string) tea.Cmd {
       return func() tea.Msg {
           var err error
           switch action {
           case "start": err = client.StartContainer(containerID)
           case "stop": err = client.StopContainer(containerID, 10)
           case "restart": err = client.RestartContainer(containerID, 10)
           }
           return ContainerActionResultMsg{...}
       }
   }
   ```
   - ✅ Pro: DRY (Don't Repeat Yourself)
   - ✅ Pro: Single function to test (table-driven)
   - ✅ Pro: Easier to maintain
   - ❌ Con: Slightly less explicit naming

**Decision**: **Option 2** - Generic function

**Rationale**:
- Start, Stop, Restart are semantically the same operation (action on container)
- Error handling is identical for all three
- Result message structure is identical
- Action type is already parameterized in ContainerActionMsg

**Implementation**:
```go
// Single generic command factory
func containerActionCmd(client *docker.Client, containerID, action string) tea.Cmd {
    return func() tea.Msg {
        var err error
        
        switch action {
        case "start":
            err = client.StartContainer(containerID)
        case "stop":
            err = client.StopContainer(containerID, 10)
        case "restart":
            err = client.RestartContainer(containerID, 10)
        default:
            return ContainerActionResultMsg{
                Success: false,
                Error:   fmt.Errorf("invalid action: %s", action),
            }
        }
        
        return ContainerActionResultMsg{
            Success: err == nil,
            Error:   err,
        }
    }
}
```

**Task Mapping**:
- T058-T060: Collapse into single implementation
- T062: Test with table-driven approach (one test, multiple actions)

**Consequences**:
- ✅ Less code to write and maintain
- ✅ Easier to extend (add new actions by adding case to switch)
- ⚠️ Tasks list slightly misaligned (acceptable - tasks are guidance, not law)

**Status**: Approved - Implement generic containerActionCmd in T058

---

## ADR-005: No ComposeUp/ComposeStart Implementation

**Decision**: Do NOT implement ComposeUp or ComposeStart methods

**Context**:
- Phase 3 implements: ComposeStop, ComposeRestart, ComposeDown
- User Story 2 doesn't mention starting entire stack
- 20i-gui doesn't provide "start all" functionality

**Options Considered**:

1. **Implement ComposeUp** (start entire stack)
   - ✅ Pro: Completes lifecycle symmetry (start/stop/restart)
   - ❌ Con: Not part of user story
   - ❌ Con: Not in 20i-gui baseline
   - ❌ Con: Complex (needs pull, build, depends_on ordering)

2. **Omit ComposeUp** (current plan)
   - ✅ Pro: Stays focused on management, not initialization
   - ✅ Pro: Users already know `docker compose up -d` for starting
   - ✅ Pro: Simpler implementation (no build/pull logic needed)
   - ❌ Con: Asymmetry (can stop all, but not start all)

**Decision**: **Option 2** - Omit ComposeUp

**Rationale**:
- Stack initialization is typically done ONCE (setup phase)
- TUI is for ONGOING MANAGEMENT of running stacks
- Starting individual containers (already implemented) is sufficient
- `docker compose up` is well-known and reliable

**User Workflow**:
```bash
# Initial setup (outside TUI)
docker compose up -d

# Ongoing management (inside TUI)
- Press 's' to start/stop individual services
- Press 'S' to stop entire stack
- Press 'R' to restart entire stack
- Press 'D' to destroy stack

# Restart stack after destroy (outside TUI)
docker compose up -d
```

**Consequences**:
- ✅ Simpler implementation (3 Compose methods, not 4)
- ✅ Clear separation: Docker Compose CLI for setup, TUI for management
- ⚠️ "Start all" not available (acceptable - individual start works fine)

**Status**: Approved - Implement only Stop/Restart/Down in T036-T038

---

## ADR-006: Error Message Formatting Strategy

**Decision**: Implement centralized formatDockerError() function with regex patterns

**Context**:
- Docker SDK returns low-level errors ("bind: address already in use")
- Need user-friendly messages ("Port 8080 already in use. Stop conflicting service.")
- Contract specifies error message templates

**Options Considered**:

1. **Inline Error Formatting** (in each handler)
   - ✅ Pro: Simple to implement
   - ❌ Con: Duplicated logic across handlers
   - ❌ Con: Inconsistent error messages

2. **Centralized Error Formatter**
   ```go
   func formatDockerError(err error, action, containerName string) string {
       if strings.Contains(err.Error(), "port is already allocated") {
           port := extractPort(err.Error())
           return fmt.Sprintf("Port %s already in use", port)
       }
       // ... more patterns
   }
   ```
   - ✅ Pro: Consistent error messages
   - ✅ Pro: Single place to update patterns
   - ✅ Pro: Easy to test
   - ❌ Con: Slightly more complex

**Decision**: **Option 2** - Centralized formatter

**Implementation Location**:
- Create internal/views/dashboard/errors.go (or add to dashboard.go)
- Called from containerActionResultMsg handler

**Pattern Matching Rules** (from docker-api.md):
```go
func formatDockerError(err error, action, containerName string) string {
    if err == nil {
        return ""
    }
    
    errStr := err.Error()
    
    // Port conflict
    if strings.Contains(errStr, "port is already allocated") {
        port := extractPort(errStr)
        return fmt.Sprintf("❌ Port %s already in use. Stop conflicting service.", port)
    }
    
    // Timeout
    if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
        return fmt.Sprintf("❌ Timeout: Container took too long to %s. Try again.", action)
    }
    
    // Not found
    if strings.Contains(errStr, "No such container") {
        return "❌ Container not found. It may have been removed. Press 'r' to refresh."
    }
    
    // Permission denied
    if strings.Contains(errStr, "permission denied") {
        return "❌ Permission denied. Add your user to the docker group."
    }
    
    // Generic fallback
    return fmt.Sprintf("❌ Failed to %s '%s': %s", action, containerName, err)
}

func extractPort(errStr string) string {
    // Regex to extract port number from Docker error
    re := regexp.MustCompile(`port (\d+)`)
    matches := re.FindStringSubmatch(errStr)
    if len(matches) > 1 {
        return matches[1]
    }
    return "unknown"
}
```

**Test Coverage** (T066):
```go
func TestFormatDockerError(t *testing.T) {
    tests := []struct {
        name     string
        err      error
        action   string
        expected string
    }{
        {
            name:     "port conflict",
            err:      errors.New("bind: address already in use, port 8080 is already allocated"),
            action:   "start",
            expected: "❌ Port 8080 already in use",
        },
        // ... more test cases
    }
}
```

**Consequences**:
- ✅ Consistent, user-friendly error messages
- ✅ Easy to add new patterns as issues are discovered
- ✅ Good test coverage
- ⚠️ Regex patterns may need tuning for different Docker versions

**Status**: Approved - Implement formatDockerError() in T065

---

## Summary of Decisions

| ADR | Decision | Impact | Status |
|-----|----------|--------|--------|
| ADR-001 | Minimal Container schema (6 fields) | Faster Phase 3, extend in Phase 5 | ✅ Approved |
| ADR-002 | 2-panel layout in Phase 3 | Simpler UI, expand in Phase 5 | ✅ Approved |
| ADR-003 | String-based actions (not enums) | No breaking changes, validate at runtime | ✅ Approved |
| ADR-004 | Generic containerActionCmd | Less code duplication, DRY | ✅ Approved |
| ADR-005 | No ComposeUp implementation | Focus on management, not setup | ✅ Approved |
| ADR-006 | Centralized error formatter | Consistent UX, testable | ✅ Approved |

---

## Implementation Implications

### Simplified Task List

Based on these decisions, some tasks consolidate:

**Original Tasks**:
- T058: startContainerCmd
- T059: stopContainerCmd
- T060: restartContainerCmd

**Actual Implementation**:
- T058-T060 (combined): Implement generic containerActionCmd

**Testing Simplification**:
- T062: Table-driven test for containerActionCmd (not 3 separate tests)

### Code Structure

```
internal/
├── docker/
│   └── client.go
│       ├── Container (6 fields) ← ADR-001
│       ├── ListContainers()
│       ├── Start/Stop/Restart()
│       └── ComposeStop/Restart/Down() ← ADR-005 (no ComposeUp)
├── views/
│   └── dashboard/
│       ├── dashboard.go
│       │   ├── DashboardModel
│       │   ├── View() (2-panel) ← ADR-002
│       │   └── containerActionCmd() ← ADR-004
│       ├── service_list.go
│       └── errors.go ← ADR-006
│           └── formatDockerError()
└── app/
    └── messages.go
        └── ContainerActionMsg.Action (string) ← ADR-003
```

---

## References

- **PHASE3-IMPLEMENTATION-NOTES.md**: Detailed architectural decisions and patterns
- **PHASE3-ROADMAP.md**: Step-by-step execution plan
- **tasks.md**: Full task breakdown (T026-T072)
- **contracts/docker-api.md**: Docker client contract
- **contracts/ui-events.md**: Message type contract
- **data-model.md**: Entity schemas and relationships

---

**Last Updated**: 2025-12-28  
**Next Review**: After Phase 3 completion (validate decisions against actual implementation)
