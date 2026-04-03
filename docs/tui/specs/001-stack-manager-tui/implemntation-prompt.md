# GitHub Copilot Cloud Implementation Prompt: Phase 3a TUI

## Project Context

You are implementing **Phase 3a** of the 20i Stack Manager Terminal UI (TUI) - a project-aware web stack management tool built with Go and Bubble Tea framework.

**Repository**: `peternicholls/20i-Hosting-Stack-for-Docker`  
**Branch**: `001-stack-manager-tui`  
**Working Directory**: `/Users/peternicholls/docker/20i-stack`  
**TUI Source**: `tui/` (at repository root)

### What This Replaces

The TUI replaces the existing `20i-gui` bash script with a modern, keyboard-driven interface. The bash script workflow is:

```bash
cd ~/my-website/              # Navigate to web project
./20i-gui                     # Launch GUI
# Menu: 1=Start, 2=Stop, 3=Restart, 4=Status, 5=Logs, 6=Destroy
```

The bash script:
- Detects current directory as project root
- Validates `public_html/` exists
- Sets `CODE_DIR=$(pwd)` and `COMPOSE_PROJECT_NAME={sanitized-name}`
- Runs `docker compose -f $STACK_FILE up -d` to start stack
- Shows container status and URLs

**Your task**: Replicate this workflow in a TUI with real-time updates and better UX.

---

## Phase 3a Objectives

**Goal**: Project-aware stack management MVP - detect project, validate, start/stop/restart/destroy stack

**Core Workflow** (what the user experiences):
1. User navigates to web project: `cd ~/my-website/`
2. User launches TUI: `20i-stack-manager`
3. TUI detects project name from directory, validates `public_html/` exists
4. Right panel shows pre-flight status (✅ Ready or ⚠️ Missing public_html)
5. User presses `S` to start stack (or `T` to create template if missing)
6. TUI runs `docker compose up -d` with proper environment variables
7. Right panel streams compose output, then shows status table with URLs
8. User can click URLs to open in browser, or use keyboard shortcuts

**What's In Scope** (Phase 3a - 61 tasks):
- ✅ Project detection from `$PWD`
- ✅ Project name sanitization (same as 20i-gui)
- ✅ `public_html/` validation (pre-flight check)
- ✅ Template installation from `demo-site-folder/`
- ✅ Three-panel layout (left: project info, right: dynamic content, bottom: commands)
- ✅ Stack lifecycle: Start, Stop, Restart, Destroy
- ✅ Status table with URLs, ports, CPU%
- ✅ Clickable URLs (mouse support)
- ✅ Double-confirmation for destroy
- ✅ Auto-refresh status every 5 seconds
- ✅ Streaming compose output
- ✅ User-friendly error messages
- ✅ Comprehensive documentation (godoc, user guides, README, CHANGELOG)

**What's Deferred** (Phase 3b+):
- ❌ Individual container start/stop/restart
- ❌ Multi-project browser
- ❌ Log viewer with follow mode
- ❌ Configuration editor

---

## Technical Stack

**Language**: Go 1.21+  
**Framework**: Bubble Tea v1.3.10+ (Elm Architecture: Model-Update-View)  
**Components**: Bubbles v1.0.0+ (list, viewport, textinput)  
**Styling**: Lipgloss v1.0.0+ (NO raw ANSI codes)  
**Docker**: Docker SDK for Go v27.0.0+

**Architecture Pattern** (Bubble Tea Elm Architecture):
```go
// Every component implements tea.Model interface
type Model struct { /* state */ }

func (m Model) Init() tea.Cmd {
    // Return initial commands (async operations)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle messages, update state, return commands
    // NEVER block here - use tea.Cmd for async work
}

func (m Model) View() string {
    // Pure function - render state to string
    // Use Lipgloss for all styling
}
```

**Key Constraints**:
- **NEVER block in Update()** - use `tea.Cmd` for all I/O
- **NEVER use raw ANSI codes** - always use Lipgloss
- **NEVER hard-code paths** - use `STACK_FILE`/`STACK_HOME` detection
- **ALWAYS use godoc comments** for exported types
- **ALWAYS handle errors** with user-friendly messages

---

## Project Structure

```
tui/
├── main.go                    # Entry point (create RootModel, enable mouse)
├── go.mod                     # Dependencies (Bubble Tea, Lipgloss, Docker SDK)
├── README.md                  # Build and usage instructions
├── Makefile                   # Build targets
└── internal/
    ├── app/
    │   ├── root.go            # RootModel - routes global shortcuts
    │   └── messages.go        # Custom tea.Msg types
    ├── project/
    │   ├── detector.go        # DetectProject() - read $PWD, check public_html
    │   ├── sanitize.go        # SanitizeProjectName() - 20i-gui compatible
    │   └── template.go        # InstallTemplate() - copy demo-site-folder
    ├── stack/
    │   ├── compose.go         # ComposeUp/Down/Restart/Destroy
    │   ├── env.go             # STACK_FILE/STACK_HOME detection + validation
    │   ├── platform.go        # ARM64 vs x86 phpMyAdmin image selection
    │   └── status.go          # GetStackStatus() - list containers, URLs, CPU%
    ├── docker/
    │   ├── client.go          # Docker SDK wrapper
    │   └── stats.go           # CPU% collection
    ├── views/
    │   ├── dashboard/
    │   │   ├── dashboard.go   # DashboardModel (three-panel layout)
    │   │   ├── left_panel.go  # Project info (name, path, status)
    │   │   ├── right_panel.go # Dynamic: pre-flight → output → status table
    │   │   ├── bottom_panel.go # Commands and status messages
    │   │   └── status_table.go # Table with clickable URLs
    │   └── help/
    │       └── help.go        # Help modal
    └── ui/
        ├── styles.go          # Lipgloss color palette
        ├── components.go      # StatusIcon, confirmation modal
        └── layout.go          # Panel sizing functions

docs/tui/                      # User documentation
├── user-guide.md              # Installation, usage, shortcuts
├── troubleshooting.md         # Common issues
└── architecture.md            # Developer guide
```

---

## Constitution Compliance (NON-NEGOTIABLE)

These principles from `.specify/memory/constitution.md` **MUST** be followed:

### I. Environment-Driven Configuration
- ✅ Read `STACK_FILE` and `STACK_HOME` from environment
- ✅ Fall back to executable-relative path detection
- ✅ Validate `STACK_FILE` exists before compose operations
- ❌ NO hard-coded paths to docker-compose.yml

### II. Multi-Platform First
- ✅ Detect ARM64 vs x86 using `runtime.GOARCH`
- ✅ Use `arm64v8/phpmyadmin:latest` on ARM, `phpmyadmin:latest` on x86
- ✅ Allow override via `PHPMYADMIN_IMAGE` env var

### III. Path Independence
- ✅ Use `$PWD` for project root detection
- ✅ Sanitize project names same as 20i-gui: lowercase, hyphens, no leading numbers

### V. User Experience & Feedback
- ✅ Show pre-flight summary before stack start
- ✅ Stream compose output in real-time
- ✅ Display clear error messages (not raw Docker errors)
- ✅ Confirm destructive operations (double-confirmation for destroy)

### VI. Documentation as First-Class Artifact
- ✅ Godoc comments on all exported types
- ✅ Inline comments explaining "why" for complex logic
- ✅ Update main README.md with TUI section
- ✅ Create user guides in /docs/tui/
- ✅ Update CHANGELOG.md

---

## Implementation Tasks (61 Total)

**Prerequisites** (already complete):
- ✅ Phase 1 (Setup): 13 tasks - Go project initialized, dependencies added
- ✅ Phase 2 (Foundational): 14 tasks - Docker client, RootModel, styles, components

**Your Work** (Phase 3a):

### Block 1: Project Detection (T100-T107a) - 9 tasks
**Purpose**: Detect current directory as web project, validate `public_html/`

1. **T100**: Create `internal/project/detector.go` with `DetectProject()` function
   - Read `$PWD` as project root (use `os.Getwd()`)
   - Derive project name from `filepath.Base()`
   - Check if `public_html/` exists using `os.Stat()`
   - Return `Project{Name, Path, HasPublicHTML, StackStatus}`

2. **T101**: Create `internal/project/sanitize.go` with `SanitizeProjectName()`
   - Port logic from `20i-gui` bash script (see `/20i-gui` line 45-60)
   - Lowercase, replace invalid chars with hyphens, ensure starts with letter/number
   - Use regex: `[^a-z0-9-]+` to strip invalid chars

3. **T102**: Create `internal/project/detector_test.go` (table-driven tests)
   - Test project detection with various directory structures
   - Test sanitization: "My Website!" → "my-website", "123-test" → "test-123"

4. **T103**: Create `internal/project/template.go` with `InstallTemplate()`
   - Find `demo-site-folder/public_html/` relative to STACK_HOME
   - Copy contents to current directory using `io.Copy()` + `filepath.Walk()`

5. **T104**: Create `internal/project/template_test.go`
   - Test template installation to temp directory
   - Test error when template not found

6. **T105**: Create `internal/project/types.go` with `Project` struct
7. **T106**: Add `projectDetectedMsg` to `internal/app/messages.go`
8. **T107**: Add `templateInstalledMsg` to `internal/app/messages.go`
9. **T107a**: Add godoc comments to project/ package files

**Reference**: See `/runbooks/research/QUICK-REFERENCE.md` for Go patterns

---

### Block 2: Stack Lifecycle (T108-T118a) - 12 tasks
**Purpose**: Start/stop/restart/destroy stack with proper environment variables

10. **T108**: Create `internal/stack/env.go` with `STACK_FILE`/`STACK_HOME` detection
    - Check `os.Getenv("STACK_FILE")` first
    - Fall back to executable-relative: `filepath.Join(execDir, "../../docker-compose.yml")`
    - Match 20i-gui logic from `/20i-gui` line 100-120

11. **T109**: Add `ValidateStackFile()` function in `env.go`
    - Return error if STACK_FILE not set and cannot be detected
    - Use `os.Stat()` to verify file exists and is readable

12. **T110**: Create `internal/stack/compose.go` with `ComposeUp()` function
    - Call `ValidateStackFile()` first
    - Build environment: `CODE_DIR=$(pwd)`, `COMPOSE_PROJECT_NAME={sanitized-name}`
    - Execute: `docker compose -f $STACK_FILE up -d`
    - Return channel for streaming output (use `cmd.StdoutPipe()`)

13. **T111**: Update `ComposeDown()` to validate STACK_FILE before execution
14. **T112**: Create `ComposeDown()` function (`docker compose down`)
15. **T113**: Create `ComposeRestart()` function (`docker compose restart`)
16. **T114**: Create `ComposeDestroy()` function (`docker compose down -v`)
17. **T115**: Create `internal/stack/compose_test.go` (mock exec)
18. **T116-T118**: Add message types: `stackStartMsg`, `stackStopMsg`, `stackOutputMsg`, `stackStatusMsg`
19. **T118a**: Add godoc comments to stack/ package

**Important**: Use `os/exec` package, NOT shell commands directly. Build command: `exec.Command("docker", "compose", "-f", stackFile, "up", "-d")`

---

### Block 3: Status Table (T120-T125) - 9 tasks
**Purpose**: Display running containers with URLs and CPU%

20. **T120**: Create `internal/stack/status.go` with `GetStackStatus()`
    - Use Docker client to list containers by project label
    - Filter: `label=com.docker.compose.project={projectName}`

21. **T121**: Create `ContainerInfo` struct:
    ```go
    type ContainerInfo struct {
        Name       string  // "my-website-nginx-1"
        Service    string  // "nginx"
        Status     string  // "running" | "stopped" | "starting"
        Image      string  // "nginx:1.25-alpine"
        Port       string  // "80" (host port)
        URL        string  // "http://localhost:80" (for web services)
        CPUPercent float64 // CPU usage percentage
    }
    ```

22. **T122**: Implement URL generation logic:
    - nginx: `http://localhost:{HOST_PORT}` (default 80)
    - phpmyadmin: `http://localhost:{PMA_PORT}` (default 8081)
    - mariadb: `localhost:{MYSQL_PORT}` (no http)
    - apache: "internal" (proxied via nginx)

23. **T123**: Implement CPU% collection using Docker stats API
    - Use `client.ContainerStats()` with `stream=false` (one-shot)

24. **T123a**: Create `internal/stack/platform.go` with architecture detection
    - Use `runtime.GOARCH` to detect ARM64 vs x86_64
    - Return appropriate phpMyAdmin image

25. **T124**: Create `internal/stack/status_test.go`
26. **T125**: Add `stackContainersMsg` to messages.go

---

### Block 4: Dashboard View (T130-T137a) - 9 tasks
**Purpose**: Implement three-panel TUI layout

27. **T130**: Create `internal/views/dashboard/dashboard.go` with `DashboardModel`
    - Fields: `project Project`, `containers []ContainerInfo`, `rightPanelState string`
    - rightPanelState: "preflight" | "output" | "status"

28. **T131**: Implement `DashboardModel.Init()` - trigger project detection
29. **T132**: Create `internal/views/dashboard/left_panel.go`
    - Render project name with status icon (✅/⚠️/🔄)
    - Use Lipgloss for styling (see `internal/ui/styles.go`)

30. **T133**: Create `internal/views/dashboard/bottom_panel.go`
    - Show available commands based on state
    - Example: "S:Start  T:Stop  R:Restart  D:Destroy  ?:Help  q:Quit"

31. **T134**: Create `internal/views/dashboard/right_panel.go`
    - Switch rendering based on `rightPanelState`
    - Use `switch` statement to render different content

32. **T135**: Create `internal/views/dashboard/status_table.go`
    - Render table with Lipgloss `JoinVertical()`
    - CPU graph: Use Unicode blocks `▓` (filled) `░` (empty)
    - Track URL positions for mouse clicks

33. **T135a**: Implement clickable URL support
    - Handle `tea.MouseMsg` events
    - Detect clicks on URL regions
    - Open browser: `exec.Command("open", url)` (macOS) or `exec.Command("xdg-open", url)` (Linux)

34. **T135b**: Add tests for URL click detection
35. **T136**: Implement `DashboardModel.View()` with three-panel layout
    - Left: 25% width, Right: 75% width, Bottom: 3 lines
    - Use `lipgloss.JoinHorizontal()` and `JoinVertical()`

36. **T137**: Create `internal/views/dashboard/dashboard_test.go`
37. **T137a**: Add godoc comments to views/dashboard/ package

**Reference**: See `/runbooks/research/lipgloss-styling-reference.md` for layout patterns

---

### Block 5: Keyboard & Mouse Handling (T140-T146a) - 9 tasks
**Purpose**: Implement S/T/R/D keys for stack operations + mouse support

38. **T140**: Implement 'S' key handler in `DashboardModel.Update()`
    - Only works if `public_html` exists
    - Return command: `composeUpCmd()` (creates tea.Cmd)
    - Switch right panel to "output" state

39. **T141**: Implement 'T' key handler (dual purpose)
    - If `public_html` missing: trigger template installation
    - If stack running: trigger ComposeDown

40. **T142**: Implement 'R' key handler - trigger ComposeRestart
41. **T143**: Implement 'D' key handler - show first destroy confirmation

42. **T144**: Create double-confirmation modal in `internal/ui/components.go`
    - First modal: "⚠️ Destroy stack? Type 'yes' to continue"
    - Second modal: "🔴 Are you SURE? Type 'destroy' to confirm"
    - Use Bubbles `textinput.Model` for text entry

43. **T144a**: Add modal state tracking to DashboardModel
    - `confirmationStage: 0 | 1 | 2`
    - `firstInput`, `secondInput` fields

44. **T145**: Implement double-confirmation flow
    - Stage 1: "yes" → advance to stage 2
    - Stage 2: "destroy" → execute ComposeDestroy
    - Esc at any stage: cancel

45. **T146**: Add key handler tests
46. **T146a**: Add mouse handler tests

**Reference**: See `/runbooks/research/bubbletea-component-guide.md` for textinput examples

---

### Block 6: Output Streaming (T150-T155) - 6 tasks
**Purpose**: Stream compose output to right panel

47. **T150**: Implement compose output streaming in `stack/compose.go`
    - Execute command with `cmd.StdoutPipe()`
    - Read lines in goroutine, send via channel
    - Close channel on completion

48. **T151**: Create `composeOutputCmd` in `dashboard.go`
    - Subscribe to output channel
    - Send `stackOutputMsg` for each line

49. **T152**: Implement `stackOutputMsg` handler in `DashboardModel.Update()`
    - Append line to output buffer
    - Scroll to bottom

50. **T153**: Implement output viewport in `right_panel.go`
    - Use Bubbles `viewport.Model` for scrollable output

51. **T154**: Transition to status table when compose completes
    - Detect completion, refresh container list, switch to "status" state

52. **T155**: Add tests for output streaming

---

### Block 7: Status Refresh (T160-T164) - 5 tasks
**Purpose**: Auto-refresh status table every 5 seconds

53. **T160**: Implement 5-second auto-refresh timer
    - Use `time.NewTicker(5 * time.Second)`
    - Return `tea.Tick` command

54. **T161**: Implement `stackContainersMsg` handler
55. **T162**: Implement CPU% collection in refresh cycle
56. **T163**: Implement timer cleanup (cancel on view switch)
57. **T164**: Add timer tests

---

### Block 8: Error Handling (T170-T172) - 3 tasks
**Purpose**: User-friendly error messages

58. **T170**: Create `internal/stack/errors.go` with `formatDockerError()`
    - Port conflict: "Port 80 is already in use. Stop the conflicting service."
    - Docker not running: "Docker daemon is not running. Start Docker Desktop."
    - Use regex to extract port numbers from Docker error strings

59. **T171**: Implement error display in bottom panel
    - Show ❌ prefix with red color
    - Clear after 5 seconds

60. **T172**: Add error formatting tests

---

### Block 9: Integration & Documentation (T180-T188f) - 16 tasks
**Purpose**: Wire everything together, comprehensive docs

61. **T180**: Wire DashboardModel into RootModel
62. **T180a**: Enable mouse support in main.go
    - Add `tea.WithMouseCellMotion()` to `tea.NewProgram()` options

63. **T181**: Implement '?' help key
64. **T182**: Create help modal content
65. **T183**: Update footer with commands
66. **T184**: Create `tests/integration/phase3a_test.go`
67. **T185-T187**: Manual testing + `make test`

68. **T188**: Update `tui/README.md`
69. **T188a**: Create `/docs/tui/user-guide.md`
    - Installation, quick start, shortcuts, workflows

70. **T188b**: Create `/docs/tui/troubleshooting.md`
    - Common issues: Docker not running, port conflicts, STACK_FILE not found

71. **T188c**: Create `/docs/tui/architecture.md`
    - Bubble Tea overview, component hierarchy, message flow

72. **T188d**: Update main `README.md` with TUI section
    - Add after GUI section
    - Include installation: `go install github.com/peternicholls/20i-stack/tui@latest`
    - Quick start: `cd ~/my-project && 20i-stack-manager`
    - Keyboard shortcuts table

73. **T188e**: Update `CHANGELOG.md`
    - Add [Unreleased] section
    - Under ### Added: Terminal UI, clickable URLs, double-confirmation, etc.

74. **T188f**: Documentation audit
    - Verify all exported types have godoc
    - Check complex logic has inline "why" comments

---

## Testing Requirements

**Unit Tests** (table-driven):
```go
func TestSanitizeProjectName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"spaces", "My Website", "my-website"},
        {"uppercase", "MYSITE", "mysite"},
        {"leading number", "123-test", "test-123"},
        {"special chars", "my_site!", "my-site"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := SanitizeProjectName(tt.input)
            if result != tt.expected {
                t.Errorf("got %s, want %s", result, tt.expected)
            }
        })
    }
}
```

**Integration Test Pattern**:
```go
func TestPhase3aWorkflow(t *testing.T) {
    // Setup mock Docker client
    mockClient := &MockDockerClient{
        containers: []Container{
            {Name: "test-nginx-1", Status: "running"},
        },
    }
    
    // Create model with mock
    model := NewDashboardModel(mockClient)
    
    // Send project detected message
    model, _ = model.Update(projectDetectedMsg{project: testProject})
    
    // Verify right panel shows pre-flight
    view := model.View()
    assert.Contains(t, view, "Ready to start stack")
    
    // Press S to start
    model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
    
    // Verify command returned
    assert.NotNil(t, cmd)
}
```

**Coverage Target**: >80% for all packages

---

## Documentation Requirements

### Godoc Comments (NFR-010)
```go
// Package project provides project detection and validation for the 20i Stack Manager TUI.
// It handles detecting the current directory as a web project root, validating the
// presence of public_html/, and installing templates when needed.
package project

// DetectProject analyzes the current working directory to determine if it contains
// a valid 20i web project. It checks for the presence of public_html/ and derives
// the project name from the directory basename.
//
// Returns a Project struct with Name, Path, HasPublicHTML, and StackStatus fields.
// The Name field is sanitized to be Docker Compose compatible (lowercase, hyphens only).
func DetectProject() (*Project, error) {
    // Implementation...
}
```

### Inline Comments (NFR-011)
```go
// Sanitize project name to match Docker Compose requirements.
// Why: Docker Compose project names must be lowercase alphanumeric with hyphens,
// and cannot start with a number. This matches the logic from 20i-gui bash script.
func SanitizeProjectName(name string) string {
    // Convert to lowercase first to simplify regex
    name = strings.ToLower(name)
    
    // Replace invalid characters with hyphens
    // Why: Docker Compose accepts [a-z0-9-] only
    re := regexp.MustCompile(`[^a-z0-9-]+`)
    name = re.ReplaceAllString(name, "-")
    
    // Strip leading/trailing hyphens
    // Why: Prevents names like "-myproject-" which look unprofessional
    name = strings.Trim(name, "-")
    
    return name
}
```

---

## Success Criteria

**Functional** (must all pass):
- [ ] Run TUI from directory with `public_html/`, press `S`, stack starts
- [ ] Run TUI from directory WITHOUT `public_html/`, press `T`, template created
- [ ] Press `R` while stack running, containers restart
- [ ] Press `D`, type "yes", type "destroy", stack destroyed (volumes removed)
- [ ] Click nginx URL in status table, browser opens to http://localhost:80
- [ ] CPU% graphs update every 5 seconds
- [ ] Esc cancels destroy confirmation at any stage
- [ ] Terminal resize doesn't crash or corrupt display
- [ ] `make test` passes with >80% coverage

**Documentation** (must all exist):
- [ ] All exported types have godoc comments (verify with `go doc`)
- [ ] Complex logic has inline "why" comments
- [ ] `/docs/tui/user-guide.md` exists with installation and shortcuts
- [ ] `/docs/tui/troubleshooting.md` exists with common issues
- [ ] `/docs/tui/architecture.md` exists with Bubble Tea patterns
- [ ] Main `README.md` includes TUI section
- [ ] `CHANGELOG.md` documents Phase 3a features

**Performance** (verify with benchmarks):
- [ ] TUI starts in <2s on M1 Mac
- [ ] Panel switching <50ms
- [ ] Container action feedback <300ms
- [ ] Memory usage <30MB with 4 services

---

## Critical Resources

**Must Read Before Starting**:
1. `/runbooks/research/QUICK-REFERENCE.md` - Bubble Tea cheat sheet (keep open)
2. `/runbooks/research/bubbletea-component-guide.md` - Component patterns
3. `/runbooks/research/lipgloss-styling-reference.md` - Styling and colors
4. `/20i-gui` - Original bash script (lines 45-60 for sanitization, 100-120 for env detection)
5. `/specs/001-stack-manager-tui/spec.md` - Full specification
6. `/specs/001-stack-manager-tui/tasks.md` - Detailed task breakdown
7. `.specify/memory/constitution.md` - Non-negotiable principles

**Code References**:
- Bubble Tea examples: https://github.com/charmbracelet/bubbletea/tree/master/examples
- Lipgloss gallery: https://github.com/charmbracelet/lipgloss/tree/master/examples
- Docker SDK docs: https://pkg.go.dev/github.com/docker/docker/client

---

## Common Pitfalls to Avoid

❌ **DON'T**:
- Block in `Update()` method (use `tea.Cmd` for async work)
- Use raw ANSI codes (always use Lipgloss)
- Hard-code paths to docker-compose.yml (use STACK_FILE detection)
- Call Docker API directly in Update (wrap in tea.Cmd)
- Forget to handle terminal resize (tea.WindowSizeMsg)
- Skip error handling (wrap all Docker errors with formatDockerError)
- Forget godoc comments on exported types
- Use `fmt.Println()` for debugging (use Bubble Tea's debugging tools)

✅ **DO**:
- Use `tea.Cmd` for all I/O operations
- Style everything with Lipgloss
- Validate STACK_FILE exists before compose operations
- Test with table-driven tests
- Add godoc comments as you write code
- Follow Elm Architecture strictly (Model-Update-View)
- Handle mouse events alongside keyboard shortcuts
- Clean up timers when switching views

---

## Development Workflow

**Initial Setup**:
```bash
cd /Users/peternicholls/docker/20i-stack/tui
go mod download  # Download dependencies (already in go.mod)
```

**Build & Run**:
```bash
make build       # Compile binary to bin/20i-stack-manager
./bin/20i-stack-manager  # Run TUI

# Or during development:
go run main.go
```

**Testing**:
```bash
make test        # Run all tests
make test-coverage  # Generate coverage report (target >80%)

# Run specific package tests:
go test ./internal/project/... -v
go test ./internal/stack/... -v
```

**Manual Testing Checklist**:
1. Test from demo-site-folder: `cd ../demo-site-folder && ../tui/bin/20i-stack-manager`
2. Test from empty directory: `mkdir /tmp/test-project && cd /tmp/test-project && .../20i-stack-manager`
3. Test template creation (press T)
4. Test stack start (press S)
5. Test clicking URLs in status table
6. Test destroy with double-confirmation (press D, type yes, type destroy)
7. Test Esc canceling at each confirmation stage
8. Test terminal resize (resize window while TUI running)

---

## Implementation Order (Critical Path)

**Week 1** (Foundation):
- Days 1-2: Project detection (T100-T107a) + Stack lifecycle (T108-T118a)
- Days 3-4: Status table (T120-T125) + Dashboard view (T130-T137a)

**Week 2** (Interaction):
- Days 1-2: Keyboard handling (T140-T146a) + Output streaming (T150-T155)
- Days 3-4: Status refresh (T160-T164) + Error handling (T170-T172)

**Week 3** (Polish):
- Days 1-2: Integration (T180-T187)
- Days 3-5: Documentation (T188-T188f) + Final testing

**Parallel Opportunities**:
- T102/T104/T115/T124 (tests) can run alongside implementation
- T107a/T118a/T137a (godoc) can happen as you write code
- T188a-T188f (docs) can be written by separate team member

---

## Deliverables

When complete, you should have:

1. **Working TUI Binary**: `tui/bin/20i-stack-manager`
2. **Test Coverage**: >80% across all packages
3. **User Documentation**:
   - `/docs/tui/user-guide.md`
   - `/docs/tui/troubleshooting.md`
   - `/docs/tui/architecture.md`
4. **Updated Repository Docs**:
   - Main `README.md` with TUI section
   - `CHANGELOG.md` with Phase 3a features
5. **Code Documentation**:
   - Godoc comments on all exported types
   - Inline comments on complex logic
6. **Passing Tests**: `make test` succeeds

---

## Questions & Support

**Reference Order**:
1. Check `/runbooks/research/QUICK-REFERENCE.md` first (fastest)
2. Check `/runbooks/research/INDEX.md` for which detailed guide to read
3. Refer to spec.md or tasks.md for requirements
4. Look at Bubble Tea examples for patterns

**Key Contacts**:
- Spec questions: See `/specs/001-stack-manager-tui/spec.md`
- Task questions: See `/specs/001-stack-manager-tui/tasks.md`
- Constitution questions: See `.specify/memory/constitution.md`

---

## Final Notes

This is **Phase 3a MVP** - focus on replicating the 20i-gui workflow first. Don't get distracted by Phase 3b features (individual container management) or Phase 4+ features (multi-project, logs, config editor). Those come later.

**Your mission**: Make it easy for web developers to start their 20i stack by simply running `20i-stack-manager` from their project directory. Everything else is secondary.

Good luck! 🚀
