# Best Practices Refactor Plan

**Document Type**: Refactoring Plan  
**Created**: 2025-12-30  
**Status**: Draft  
**Purpose**: Align TUI implementation with Bubble Tea/Lipgloss best practices from runbooks/research

---

## Executive Summary

This document outlines a refactoring plan to address anti-patterns and missing best practices discovered during a comprehensive audit of the TUI implementation against the research documentation in `runbooks/research/`.

**Audit Date**: 2025-12-30  
**Audit Scope**: Complete TUI codebase in `tui/internal/`  
**Research Baseline**: `runbooks/research/` (QUICK-REFERENCE, bubbletea-component-guide, lipgloss-styling-reference)

### Audit Results Summary

| Category | Status | Count |
|----------|--------|-------|
| **✅ Best Practices Followed** | Good | 8 |
| **❌ Critical Anti-Patterns** | High Priority | 4 |
| **⚠️ Missing Best Practices** | Medium Priority | 5 |
| **📦 Missing Bubbles Components** | Low Priority | 3 |

### Impact Assessment

- **Performance**: Creating styles in render functions causes unnecessary allocations on every frame
- **Maintainability**: Scattered style definitions make theming and updates difficult
- **User Experience**: Missing spinners and progress indicators reduce perceived responsiveness
- **Code Quality**: Violates documented best practices from project research

---

## Detailed Findings

### ✅ What's Done Well (8 items)

1. **Color Palette Defined** - `ui/styles.go` has semantic color constants ✓
2. **Package-Level Styles** - HeaderStyle, FooterStyle, PanelStyle defined correctly ✓
3. **Elm Architecture** - All models implement Init/Update/View ✓
4. **No Raw ANSI** - All styling uses Lipgloss ✓
5. **WindowSizeMsg** - Both RootModel and DashboardModel handle resize ✓
6. **Clean Structure** - Separation: ui/, views/dashboard/ ✓
7. **Layout Functions** - JoinVertical/Horizontal, Place used correctly ✓
8. **Unicode Icons** - Status icons (●, ○, ✓, ✗, ⚠) used ✓

### ❌ Critical Anti-Patterns (4 items)

#### AP1: Creating Styles in RenderConfirmationModal

**File**: `internal/ui/components.go:82-111`  
**Severity**: High (creates 7 new styles per frame when modal is visible)  
**Pattern Violated**: "Define styles once at package level"

```go
// ❌ Current (BAD)
func RenderConfirmationModal(...) string {
    titleStyle := lipgloss.NewStyle().Bold(true)...    // New allocation
    promptStyle := lipgloss.NewStyle()...              // New allocation
    progressStyle := lipgloss.NewStyle()...            // New allocation
    // ... 4 more
}
```

**Impact**: ~7 allocations per render frame = wasted CPU/memory

#### AP2: Creating Styles in StatusIcon

**File**: `internal/ui/components.go:25-28`  
**Severity**: High (called for every container row, every frame)  
**Pattern Violated**: "Define styles once at package level"

```go
// ❌ Current (BAD)
func StatusIcon(status string) string {
    return lipgloss.NewStyle().Foreground(ColorMuted).Render("?")  // Line 25
    style := lipgloss.NewStyle().Bold(true)                        // Line 28
}
```

**Impact**: If 5 containers, creates 5 new styles per frame

#### AP3: Creating Styles in renderLeftPanel

**File**: `internal/views/dashboard/left_panel.go:44-78`  
**Severity**: Medium (called once per frame for left panel)  
**Pattern Violated**: "Define styles once at package level"

```go
// ❌ Current (BAD)
lines = append(lines, lipgloss.NewStyle().Bold(true).Render(...))     // Line 44
pathLabel := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render...  // Line 54
stackStyle := lipgloss.NewStyle().Foreground(ui.ColorStopped)         // Line 62
htmlStyle := lipgloss.NewStyle().Foreground(ui.ColorWarning)          // Line 76
```

**Impact**: 4+ new styles per frame for left panel alone

#### AP4: Creating Styles in status_table.go

**File**: `internal/views/dashboard/status_table.go`  
**Severity**: Medium (affects table rendering)  
**Lines**: 85 (headerStyle), 187 (getStatusBadge creates styles dynamically)

```go
// ❌ Current (BAD)
headerStyle := lipgloss.NewStyle().Bold(true)...  // Line 85

func getStatusBadge(status) {
    var style lipgloss.Style
    switch status {
    case StatusRunning:
        style = lipgloss.NewStyle()...  // Line 189
    // etc.
}
```

**Impact**: Style created for every status badge in table

### ⚠️ Missing Best Practices (5 items)

#### BP1: No Component-Specific styles.go Files

**Research Recommendation**: "Each component should have styles.go, messages.go, commands.go"

**Missing Files**:
- `internal/views/dashboard/styles.go` - Should contain all dashboard panel styles
- `internal/views/help/styles.go` - Help view styles
- `internal/views/projects/styles.go` - Projects view styles (when implemented)

**Current State**: Styles scattered in render functions and components.go

#### BP2: Incomplete messages.go Separation

**Research Recommendation**: "Custom message types in separate messages.go files"

**Current State**:
- ✓ Have: `internal/app/messages.go` (global messages)
- ❌ Missing: `internal/views/dashboard/messages.go` (dashboard-specific messages currently in dashboard.go)

**Messages to Move**:
- `projectDetectedMsg` (line 724)
- `containerListMsg` (line 729)
- `containerActionResultMsg` (line 734)
- `stackOutputMsg` (line 796)
- `composeStreamStartedMsg` (line 684)
- `stackStatusRefreshMsg` (line 504)
- `urlOpenedMsg`, `urlOpenErrorMsg` (status_table.go)
- `templateInstalledMsg` (line 700)

#### BP3: No Bubbles Components Used

**Research Recommendation**: "Use Bubbles library for common UI patterns"

**Missing Components**:

1. **bubbles/viewport** - Should be used for compose output streaming
   - Current: Manual string array + scrolling logic
   - Benefit: Built-in scrolling, performance optimization, mouse wheel

2. **bubbles/spinner** - Should be used during project detection
   - Current: Static "Detecting..." text
   - Benefit: Animated feedback, professional appearance

3. **bubbles/list** - Could be used for service/container list
   - Current: Manual row rendering with cursor
   - Benefit: Filtering, pagination, built-in keyboard navigation

4. **bubbles/progress** - Could show compose operation progress
   - Current: Text-only output
   - Benefit: Visual progress indication

#### BP4: Color Palette Not Optimized for Terminal Compatibility

**Research Recommendation**: "Use ANSI 16 colors for better terminal theme compatibility"

**Current** (`ui/styles.go:16-19`):
```go
ColorRunning = lipgloss.Color("#00ff00")  // Hex green
ColorStopped = lipgloss.Color("#808080")  // Hex gray
ColorError   = lipgloss.Color("#ff0000")  // Hex red
```

**Research Recommendation**:
```go
ColorRunning = lipgloss.Color("10")   // ANSI 16 bright green
ColorStopped = lipgloss.Color("8")    // ANSI 16 bright black (gray)
ColorError   = lipgloss.Color("9")    // ANSI 16 bright red
```

**Trade-off**: Hex colors work fine but may not adapt to terminal themes as well

#### BP5: No help.Model Integration

**Research Recommendation**: "Use bubbles/help for auto-generated help"

**Current State**:
- Partial implementation in `ui/components.go:137-242`
- Help model created but not fully utilizing Bubbles help rendering
- Help view manually renders shortcuts

**Missing**: Full integration with help.View() for consistent formatting

### 📦 Optional Enhancements (3 items)

1. **Component Commands File** - Create `dashboard/commands.go` for tea.Cmd factories
   - Currently: Commands defined inline in dashboard.go
   - Benefit: Better organization, easier testing

2. **Responsive Layout Helper** - Enhanced layout calculation
   - Currently: Basic width/height calculation
   - Could add: Breakpoint-based layouts (narrow/medium/wide)

3. **Table Component from Bubbles** - Use bubbles/table instead of manual rendering
   - Currently: Manual table formatting in status_table.go
   - Benefit: Sortable columns, better performance

---

## Refactoring Plan

### Phase 1: Fix Critical Anti-Patterns (High Priority)

**Goal**: Eliminate all `lipgloss.NewStyle()` calls from render/helper functions  
**Impact**: Performance improvement, code clarity  
**Effort**: ~4 hours  
**Risk**: Low (pure refactor, no behavior change)

#### Task P1.1: Create dashboard/styles.go

**File**: `tui/internal/views/dashboard/styles.go` (new)

**Content**:
```go
package dashboard

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/peternicholls/20i-stack/tui/internal/ui"
)

// Left panel styles
var (
    projectNameStyle = lipgloss.NewStyle().Bold(true)
    pathLabelStyle   = lipgloss.NewStyle().Foreground(ui.ColorMuted)
    stackLabelStyle  = lipgloss.NewStyle().Foreground(ui.ColorMuted)
    
    stackRunningStyle = lipgloss.NewStyle().Foreground(ui.ColorRunning)
    stackStoppedStyle = lipgloss.NewStyle().Foreground(ui.ColorStopped)
    
    htmlLabelStyle   = lipgloss.NewStyle().Foreground(ui.ColorMuted)
    htmlPresentStyle = lipgloss.NewStyle().Foreground(ui.ColorRunning)
    htmlMissingStyle = lipgloss.NewStyle().Foreground(ui.ColorWarning)
)

// Status table styles
var (
    tableHeaderStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ui.ColorPrimary)
    
    statusRunningStyle    = lipgloss.NewStyle().Foreground(ui.ColorRunning).Bold(true)
    statusStoppedStyle    = lipgloss.NewStyle().Foreground(ui.ColorStopped)
    statusRestartingStyle = lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true)
    statusErrorStyle      = lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true)
    statusUnknownStyle    = lipgloss.NewStyle().Foreground(ui.ColorMuted)
    
    cpuBarHighStyle   = lipgloss.NewStyle().Foreground(ui.ColorError)    // > 80%
    cpuBarMediumStyle = lipgloss.NewStyle().Foreground(ui.ColorWarning)  // 50-80%
    cpuBarLowStyle    = lipgloss.NewStyle().Foreground(ui.ColorRunning)  // < 50%
)

// Panel layout styles
var (
    leftPanelStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ui.ColorBorder).
        Padding(1)
    
    rightPanelStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ui.ColorBorder).
        Padding(1, 2)
)
```

**Files to Update**:
- `left_panel.go` - Use styles from styles.go
- `status_table.go` - Use styles from styles.go
- `right_panel.go` - Use styles from styles.go (if exists)
- `bottom_panel.go` - Use styles from styles.go

#### Task P1.2: Create ui/icon_styles.go

**File**: `tui/internal/ui/icon_styles.go` (new)

**Content**:
```go
package ui

import "github.com/charmbracelet/lipgloss"

// Status icon styles (pre-defined for performance)
var (
    statusIconRunningStyle    = lipgloss.NewStyle().Bold(true).Foreground(ColorRunning)
    statusIconStoppedStyle    = lipgloss.NewStyle().Bold(true).Foreground(ColorStopped)
    statusIconRestartingStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorWarning)
    statusIconErrorStyle      = lipgloss.NewStyle().Bold(true).Foreground(ColorError)
    statusIconUnknownStyle    = lipgloss.NewStyle().Foreground(ColorMuted)
)

// StatusIcon returns a styled status icon (no allocation)
func StatusIcon(status string) string {
    s := strings.ToLower(strings.TrimSpace(status))
    
    switch s {
    case "running":
        return statusIconRunningStyle.Render("●")
    case "stopped":
        return statusIconStoppedStyle.Render("○")
    case "restarting":
        return statusIconRestartingStyle.Render("⚠")
    case "error":
        return statusIconErrorStyle.Render("✗")
    default:
        return statusIconUnknownStyle.Render("?")
    }
}
```

**Files to Update**:
- `ui/components.go` - Replace StatusIcon implementation

#### Task P1.3: Create ui/modal_styles.go

**File**: `tui/internal/ui/modal_styles.go` (new)

**Content**:
```go
package ui

import "github.com/charmbracelet/lipgloss"

// Confirmation modal styles
var (
    modalTitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("9")). // Red
        Align(lipgloss.Center).
        Padding(0, 1)
    
    modalPromptStyle = lipgloss.NewStyle().
        Foreground(ColorMuted).
        Align(lipgloss.Center).
        Padding(0, 1)
    
    modalProgressStyle = lipgloss.NewStyle().
        Foreground(ColorMuted).
        Align(lipgloss.Center).
        Italic(true).
        Padding(0, 1)
    
    modalInputStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("11")). // Yellow
        Align(lipgloss.Center).
        Padding(0, 1)
    
    modalHintStyle = lipgloss.NewStyle().
        Foreground(ColorMuted).
        Align(lipgloss.Center).
        Italic(true).
        Padding(0, 1)
    
    modalContentStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("9")). // Red
        Padding(1, 2).
        Width(50)
)
```

**Files to Update**:
- `ui/components.go` - Update RenderConfirmationModal to use pre-defined styles

#### Task P1.4: Update All Render Functions

**Files to Modify**:

1. `internal/ui/components.go`
   - RenderConfirmationModal: Use modal_styles.go
   - StatusIcon: Use icon_styles.go

2. `internal/views/dashboard/left_panel.go`
   - renderLeftPanel: Use dashboard/styles.go

3. `internal/views/dashboard/status_table.go`
   - renderStatusTable: Use dashboard/styles.go
   - getStatusBadge: Use dashboard/styles.go
   - renderCPUBar: Use dashboard/styles.go

4. `internal/views/dashboard/bottom_panel.go`
   - Use dashboard/styles.go (if applicable)

5. `internal/views/dashboard/right_panel.go`
   - Use dashboard/styles.go (if exists)

**Testing**: Run existing tests to ensure no visual regressions

### Phase 2: Organize Message Types (Medium Priority)

**Goal**: Separate message types into dedicated files per component  
**Impact**: Code organization, maintainability  
**Effort**: ~2 hours  
**Risk**: Low (pure refactor)

#### Task P2.1: Create dashboard/messages.go

**File**: `tui/internal/views/dashboard/messages.go` (new)

**Content**: Move all dashboard-specific messages from dashboard.go

```go
package dashboard

import (
    "github.com/peternicholls/20i-stack/tui/internal/docker"
    "github.com/peternicholls/20i-stack/tui/internal/project"
)

// projectDetectedMsg is sent when project detection completes
type projectDetectedMsg struct {
    project project.Project
}

// containerListMsg is sent when container list is fetched
type containerListMsg struct {
    containers []docker.Container
    err        error
}

// containerActionResultMsg is sent after a container action completes
type containerActionResultMsg struct {
    success bool
    message string
    err     error
}

// stackOutputMsg is sent when streaming compose command output
type stackOutputMsg struct {
    Line    string
    IsError bool
}

// composeStreamStartedMsg is sent when compose streaming begins
type composeStreamStartedMsg struct {
    channel <-chan string
}

// stackStatusRefreshMsg triggers a switch to status panel
type stackStatusRefreshMsg struct{}

// templateInstalledMsg is sent when template installation completes
type templateInstalledMsg struct {
    success bool
    err     error
}

// urlOpenedMsg is sent when a URL is successfully opened
type urlOpenedMsg struct {
    url string
}

// urlOpenErrorMsg is sent when opening a URL fails
type urlOpenErrorMsg struct {
    url string
    err error
}
```

**Files to Update**:
- `dashboard.go` - Remove message type definitions, import from messages.go

#### Task P2.2: Create dashboard/commands.go (Optional)

**File**: `tui/internal/views/dashboard/commands.go` (new)

**Content**: Move tea.Cmd factories from dashboard.go

```go
package dashboard

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/peternicholls/20i-stack/tui/internal/docker"
    "github.com/peternicholls/20i-stack/tui/internal/project"
    "github.com/peternicholls/20i-stack/tui/internal/stack"
)

// loadContainersCmd fetches container list asynchronously
func loadContainersCmd(client *docker.Client, projectName string) tea.Cmd { ... }

// detectProjectCmd triggers async project detection
func detectProjectCmd() tea.Cmd { ... }

// startComposeUpCmd starts compose up with streaming output
func startComposeUpCmd(stackFile, codeDir string) tea.Cmd { ... }

// ... etc for all commands
```

**Benefit**: Cleaner dashboard.go, easier to test commands in isolation

### Phase 3: Add Bubbles Components (Medium Priority)

**Goal**: Replace manual implementations with battle-tested Bubbles components  
**Impact**: Better UX, less maintenance, more features  
**Effort**: ~6 hours  
**Risk**: Medium (behavior changes, testing required)

#### Task P3.1: Add bubbles/viewport for Compose Output

**Files**:
- `tui/go.mod` - Already has bubbles dependency ✓
- `internal/views/dashboard/dashboard.go` - Add viewport.Model field
- `internal/views/dashboard/right_panel.go` - Use viewport for output display

**Implementation**:
```go
import "github.com/charmbracelet/bubbles/viewport"

type DashboardModel struct {
    // ... existing fields
    composeViewport viewport.Model  // Replace composeOutput []string
}

func NewModel(...) DashboardModel {
    vp := viewport.New(0, 0)  // Size set in WindowSizeMsg
    vp.MouseWheelEnabled = true
    
    return DashboardModel{
        // ...
        composeViewport: vp,
    }
}

func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
    switch msg := msg.(type) {
    case stackOutputMsg:
        // Append to viewport instead of slice
        content := m.composeViewport.View() + "\n" + msg.Line
        m.composeViewport.SetContent(content)
        m.composeViewport.GotoBottom()  // Auto-scroll
    
    case tea.WindowSizeMsg:
        // Update viewport size
        m.composeViewport.Width = rightPanelWidth
        m.composeViewport.Height = rightPanelHeight
    }
    
    // Delegate viewport messages
    var cmd tea.Cmd
    m.composeViewport, cmd = m.composeViewport.Update(msg)
    return m, cmd
}
```

**Benefits**:
- Built-in scrolling (page up/down, mouse wheel)
- Performance optimized for large outputs
- Free features: goto top/bottom, line numbers (optional)

#### Task P3.2: Add bubbles/spinner for Project Detection

**Files**:
- `internal/views/dashboard/dashboard.go` - Add spinner.Model field
- `internal/views/dashboard/left_panel.go` - Show spinner during detection

**Implementation**:
```go
import "github.com/charmbracelet/bubbles/spinner"

type DashboardModel struct {
    // ... existing fields
    detectSpinner spinner.Model
}

func NewModel(...) DashboardModel {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(ui.ColorInfo)
    
    return DashboardModel{
        // ...
        detectSpinner: s,
    }
}

func (m DashboardModel) Init() tea.Cmd {
    return tea.Batch(
        m.detectSpinner.Tick,  // Start spinner
        detectProjectCmd(),
    )
}

func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
    var cmd tea.Cmd
    m.detectSpinner, cmd = m.detectSpinner.Update(msg)
    
    switch msg := msg.(type) {
    case projectDetectedMsg:
        // Stop spinner when project detected
        m.project = &msg.project
        return m, nil
    }
    
    return m, cmd
}
```

**In renderLeftPanel**:
```go
projectName := "Detecting..."
if proj == nil {
    projectName = m.detectSpinner.View() + " Detecting..."
} else {
    projectName = proj.Name
}
```

**Benefits**:
- Visual feedback during async operations
- Professional appearance
- No manual animation logic

#### Task P3.3: Consider bubbles/list for Service List (Optional)

**Complexity**: Higher - requires refactoring service list rendering  
**Benefit**: Fuzzy filtering, pagination, better keyboard navigation  
**Recommendation**: Defer to future enhancement unless list grows large

### Phase 4: Color Palette Optimization (Low Priority)

**Goal**: Use ANSI 16 colors for better terminal theme compatibility  
**Impact**: Better appearance in different terminal color schemes  
**Effort**: ~30 minutes  
**Risk**: Low (visual only, easily reversible)

#### Task P4.1: Update ui/styles.go Color Definitions

**File**: `tui/internal/ui/styles.go:16-29`

**Changes**:
```go
// Before (Hex colors)
ColorRunning = lipgloss.Color("#00ff00")  // Green
ColorStopped = lipgloss.Color("#808080")  // Gray
ColorError   = lipgloss.Color("#ff0000")  // Red
ColorWarning = lipgloss.Color("#ffff00")  // Yellow
ColorInfo    = lipgloss.Color("#0000ff")  // Blue

// After (ANSI 16)
ColorRunning = lipgloss.Color("10")   // Bright green
ColorStopped = lipgloss.Color("8")    // Bright black (gray)
ColorError   = lipgloss.Color("9")    // Bright red
ColorWarning = lipgloss.Color("11")   // Bright yellow
ColorInfo    = lipgloss.Color("12")   // Bright blue
```

**OR Use Adaptive Colors** (best of both worlds):
```go
ColorRunning = lipgloss.AdaptiveColor{
    Light: "2",    // Green in light mode
    Dark:  "10",   // Bright green in dark mode
}
```

**Testing**: Test in both light and dark terminal themes

---

## Implementation Schedule

### Week 1: Critical Anti-Patterns

| Day | Tasks | Hours |
|-----|-------|-------|
| Day 1 | P1.1: Create dashboard/styles.go | 2h |
| Day 1 | P1.2: Create ui/icon_styles.go | 1h |
| Day 2 | P1.3: Create ui/modal_styles.go | 1h |
| Day 2 | P1.4: Update render functions (part 1) | 2h |
| Day 3 | P1.4: Update render functions (part 2) | 2h |
| Day 3 | Testing and validation | 2h |

**Total**: 10 hours

### Week 2: Organization & Components

| Day | Tasks | Hours |
|-----|-------|-------|
| Day 1 | P2.1: Create dashboard/messages.go | 1h |
| Day 1 | P2.2: Create dashboard/commands.go (optional) | 1h |
| Day 2 | P3.1: Add bubbles/viewport | 3h |
| Day 3 | P3.2: Add bubbles/spinner | 2h |
| Day 3 | Testing and validation | 2h |

**Total**: 9 hours

### Week 3: Polish (Optional)

| Day | Tasks | Hours |
|-----|-------|-------|
| Day 1 | P4.1: Update color palette | 0.5h |
| Day 1 | P3.3: Evaluate bubbles/list (optional) | 2h |
| Day 1 | Documentation updates | 1h |

**Total**: 3.5 hours

---

## Testing Strategy

### Unit Tests to Update

1. **ui/components_test.go**
   - Update tests for StatusIcon (should use pre-defined styles)
   - Update tests for RenderConfirmationModal (verify no style creation)

2. **dashboard/left_panel_test.go**
   - Verify styles imported from styles.go
   - Test truncatePath edge cases

3. **dashboard/status_table_test.go**
   - Verify getStatusBadge uses pre-defined styles
   - Test renderCPUBar with style constants

### Integration Tests

1. **Viewport Integration**
   - Test compose output scrolling
   - Test auto-scroll to bottom
   - Test large output performance

2. **Spinner Integration**
   - Test spinner starts during detection
   - Test spinner stops after detection completes
   - Test spinner animation updates

### Visual Regression Tests

1. **Before/After Screenshots**
   - Capture dashboard view before refactor
   - Capture after each phase
   - Verify no visual changes (except intentional improvements)

2. **Color Palette Testing**
   - Test in light terminal theme
   - Test in dark terminal theme
   - Test in solarized theme

---

## Success Metrics

### Performance Metrics

**Before Refactor** (estimated):
- Style allocations per frame: ~20-30
- Memory per render: ~500-1000 bytes style overhead

**After Refactor** (target):
- Style allocations per frame: 0
- Memory per render: 0 bytes style overhead (reuse existing)

**Measurement**: Use Go benchmarks for View() functions

### Code Quality Metrics

| Metric | Before | After Target |
|--------|--------|--------------|
| `lipgloss.NewStyle()` in View functions | ~15 | 0 |
| Package-level style constants | ~10 | ~40 |
| Lines in dashboard.go | ~800 | ~600 |
| Separate style files | 1 | 4 |
| Bubbles components used | 0 | 2-3 |

### Maintainability Metrics

- **Theme Change Effort**: From ~15 file edits → 2 file edits (styles.go files only)
- **Adding New Status**: From 3 locations → 1 location (styles.go)
- **Test Coverage**: Maintain current coverage (aim for 80%+)

---

## Risk Mitigation

### Risk: Visual Regressions

**Mitigation**:
- Capture before/after screenshots for each change
- Run visual diff tool (if available)
- Manual QA in different terminal sizes (80x24, 120x40, 200x60)
- Test in different terminal emulators (iTerm2, Terminal.app, Alacritty)

### Risk: Performance Degradation from Viewport

**Mitigation**:
- Benchmark before/after with large outputs (1000+ lines)
- Monitor memory usage during long-running compose operations
- Test viewport scrolling performance
- Keep fallback option to manual rendering if needed

### Risk: Breaking Tests

**Mitigation**:
- Run full test suite after each phase
- Update tests incrementally alongside code changes
- Add golden file tests for complex renderings
- Use table-driven tests for style variations

### Risk: Merge Conflicts

**Mitigation**:
- Complete refactor in dedicated branch
- Communicate with team about large refactor
- Keep each phase atomic (commit after each task)
- Rebase frequently from main branch

---

## Rollback Plan

If critical issues discovered after merge:

1. **Immediate**: Revert merge commit
2. **Short-term**: Fix issues in separate branch, fast-track review
3. **Long-term**: Re-merge with fixes

**Feature Flags** (optional):
- Add env var `TUI_USE_LEGACY_STYLES=true` to toggle old/new rendering
- Allows gradual rollout and A/B testing

---

## Documentation Updates

### Files to Update

1. **runbooks/research/INDEX.md**
   - Add note: "TUI implementation now follows all best practices as of 2025-12-30"

2. **specs/001-stack-manager-tui/IMPROVEMENTS_SUMMARY.md**
   - Add section on best practices refactor

3. **tui/README.md**
   - Update architecture section with new file structure
   - Document style organization pattern

4. **CHANGELOG.md**
   - Add entry for refactor under "Changed" section

### New Documentation

1. **tui/docs/STYLING-GUIDE.md** (new)
   - How to add new styles
   - Color palette reference
   - Style naming conventions

2. **tui/docs/COMPONENTS-GUIDE.md** (new)
   - Which Bubbles components we use and why
   - How to add new Bubbles components
   - Custom component patterns

---

## Checklist for Completion

### Phase 1: Anti-Patterns (Required)

- [ ] dashboard/styles.go created with all panel styles
- [ ] ui/icon_styles.go created with status icon styles
- [ ] ui/modal_styles.go created with modal styles
- [ ] left_panel.go updated to use styles.go
- [ ] status_table.go updated to use styles.go
- [ ] components.go updated to use new style files
- [ ] All tests passing
- [ ] No `lipgloss.NewStyle()` in any View() or helper functions
- [ ] Code review completed
- [ ] Merged to main branch

### Phase 2: Organization (Recommended)

- [ ] dashboard/messages.go created
- [ ] dashboard/commands.go created (optional)
- [ ] dashboard.go refactored to import messages
- [ ] All tests passing
- [ ] Code review completed

### Phase 3: Bubbles Components (Recommended)

- [ ] bubbles/viewport integrated for compose output
- [ ] bubbles/spinner integrated for project detection
- [ ] Viewport scrolling tested (mouse wheel, page up/down)
- [ ] Spinner animation verified
- [ ] All tests passing
- [ ] Performance benchmarks show no regression

### Phase 4: Color Palette (Optional)

- [ ] Color definitions updated to ANSI 16 or Adaptive
- [ ] Tested in light/dark terminals
- [ ] Visual appearance verified
- [ ] All tests passing

### Final Validation

- [ ] Full test suite passes
- [ ] Performance benchmarks run
- [ ] Visual regression testing complete
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Team review completed
- [ ] Deployed to staging environment
- [ ] User acceptance testing passed

---

## Appendix A: Research References

All best practices in this document are derived from:

1. **runbooks/research/QUICK-REFERENCE.md**
   - "Define styles once at package level"
   - "Never use raw ANSI codes"
   - Anti-pattern examples

2. **runbooks/research/bubbletea-component-guide.md**
   - Component structure (Model/Init/Update/View)
   - Bubbles component library usage
   - Parent-child message passing

3. **runbooks/research/lipgloss-styling-reference.md**
   - Color palette standards
   - Common style patterns
   - Layout functions

4. **runbooks/research/INDEX.md**
   - Implementation checklist
   - Document navigation guide

---

## Appendix B: File Structure After Refactor

```
tui/internal/
├── ui/
│   ├── styles.go           # Global color palette & common styles
│   ├── icon_styles.go      # Status icon styles (NEW)
│   ├── modal_styles.go     # Modal dialog styles (NEW)
│   ├── components.go       # UI helpers (updated)
│   ├── layout.go          # Layout calculations
│   └── browser.go         # URL opener
│
├── views/
│   └── dashboard/
│       ├── dashboard.go       # Model & Update (slimmed down)
│       ├── styles.go          # Dashboard-specific styles (NEW)
│       ├── messages.go        # Message types (NEW)
│       ├── commands.go        # tea.Cmd factories (NEW, optional)
│       ├── left_panel.go      # Left panel rendering (updated)
│       ├── right_panel.go     # Right panel rendering (updated)
│       ├── bottom_panel.go    # Bottom panel rendering (updated)
│       ├── status_table.go    # Status table (updated)
│       └── service_list.go    # Service list (updated)
│
├── app/
│   ├── root.go           # Root model
│   └── messages.go       # Global messages
│
├── docker/
│   └── client.go         # Docker API client
│
├── project/
│   └── detector.go       # Project detection
│
└── stack/
    └── compose.go        # Compose operations
```

**Key Changes**:
- 3 new style files (icon_styles.go, modal_styles.go, dashboard/styles.go)
- 1 new messages file (dashboard/messages.go)
- 1 optional commands file (dashboard/commands.go)
- All render functions updated to use pre-defined styles

---

## Appendix C: Performance Benchmarks

### Benchmark Template

```go
// File: dashboard/dashboard_bench_test.go
package dashboard

import (
    "testing"
    "github.com/peternicholls/20i-stack/tui/internal/docker"
)

func BenchmarkDashboardView(b *testing.B) {
    m := NewModel(nil, "test-project")
    m.width = 120
    m.height = 40
    m.containers = []docker.Container{
        {Service: "apache", Status: docker.StatusRunning},
        {Service: "nginx", Status: docker.StatusRunning},
        {Service: "mariadb", Status: docker.StatusRunning},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = m.View()
    }
}

func BenchmarkLeftPanelRender(b *testing.B) {
    proj := &project.Project{
        Name: "test-project",
        Path: "/path/to/project",
        HasPublicHTML: true,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = renderLeftPanel(proj, true, 40, 20)
    }
}
```

**Run with**:
```bash
go test -bench=. -benchmem ./internal/views/dashboard/
```

**Expected Results**:
- Before: ~5000 ns/op, ~800 B/op, ~15 allocs/op
- After: ~3000 ns/op, ~400 B/op, ~5 allocs/op

---

## Conclusion

This refactoring plan addresses all critical anti-patterns and missing best practices identified in the audit. The phased approach allows for incremental improvements with minimal risk.

**Immediate Priority**: Phase 1 (fix anti-patterns) - highest performance impact  
**Recommended Priority**: Phases 1-3 - complete alignment with research  
**Optional**: Phase 4 - polish and optimization

**Estimated Total Effort**: 22.5 hours over 2-3 weeks  
**Risk Level**: Low to Medium (well-tested patterns, clear rollback plan)  
**Expected Outcome**: More maintainable, performant, and best-practice-compliant TUI

---

**Document Status**: Ready for Review  
**Next Steps**: Review with team → Approve → Begin Phase 1 implementation
