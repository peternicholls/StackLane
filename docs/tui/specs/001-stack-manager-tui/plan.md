# Implementation Plan: Stacklane TUI

**Branch**: `001-stack-manager-tui` | **Date**: 2025-12-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-stack-manager-tui/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

A professional terminal UI (TUI) built with Bubble Tea framework to replace the earlier GUI script. Provides a modern, keyboard-driven interface with **project-aware stack management**. The TUI is run from a web project directory and manages the 20i stack for THAT project.

**Phase 3a MVP Focus**: Project detection → Pre-flight validation → Stack lifecycle → Status table with URLs

**Core Workflow** (matches the previous GUI script):
1. User navigates to web project directory: `cd ~/my-website/`
2. User launches TUI: `20i-stack-manager`
3. TUI detects project, validates `public_html/`, shows status
4. User presses `S` to start stack (sets `CODE_DIR`, `COMPOSE_PROJECT_NAME`)
5. Right panel shows compose output, then status table with URLs
6. User can stop (`T`), restart (`R`), or destroy (`D`) stack

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: Bubble Tea v1.3.10+, Bubbles v1.0.0+, Lipgloss v1.0.0+, Docker SDK v27.0.0+  
**Storage**: N/A (reads docker-compose.yml and .20i-local; no persistent state)  
**Testing**: Go testing package, table-driven tests for Docker client wrapper  
**Target Platform**: macOS (primary), Linux (secondary) - terminal-based, cross-platform  
**Project Type**: Single CLI application (TUI binary)  
**Performance Goals**: <2s startup, <50ms panel switching, <200ms stats refresh cycle  
**Constraints**: <30MB memory with 4 services + 40MB log buffer, 80x24 min terminal, no blocking I/O in UI thread  
**Scale/Scope**: 4-10 containers per project, 10k log lines buffered per container, ~1500 LOC MVP

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ I. Environment-Driven Configuration
**Status**: PASS  
**Analysis**: TUI reads from existing environment-driven config (docker-compose.yml, .20i-local, stack-vars.yml). No hard-coded credentials or paths. Respects `STACK_FILE` and `STACK_HOME` detection patterns from existing stack.

### ✅ II. Multi-Platform First
**Status**: PASS  
**Analysis**: Go compiles to native binaries for both Intel/AMD64 and ARM64. TUI is platform-agnostic (terminal-based). Docker SDK handles architecture detection automatically.

### ✅ III. Path Independence
**Status**: PASS  
**Analysis**: TUI uses Docker client to discover project directory via compose file path. No absolute paths hard-coded. Project name sanitization inherited from compose project detection.

### ✅ IV. Centralized Defaults with Override Hierarchy
**Status**: PASS  
**Analysis**: TUI respects existing hierarchy: ENV vars → .20i-local → .env → stack-vars.yml → compose defaults. No new config layer added.

### ✅ V. User Experience & Feedback
**Status**: PASS  
**Analysis**: Spec mandates clear feedback (✅/❌ emojis, inline messages, confirmation prompts for destructive ops). Footer always shows shortcuts. Error messages are actionable.

### ✅ VI. Documentation as First-Class Artifact
**Status**: PASS (pending completion)  
**Analysis**: Spec requires tui/README.md with install/usage. README.md update with TUI section. CHANGELOG.md entry planned. Inline comments mandated in code.

### ✅ VII. Version Consistency
**Status**: PASS  
**Analysis**: TUI does not introduce version variables. Reads existing PHP_VERSION, MYSQL_VERSION from environment (no sync issues).

### 🟢 All Gates Passed - Proceed to Phase 0

---

## Post-Phase 1 Re-evaluation

*Re-checked after Phase 1 design (data model, contracts, quickstart)*

### ✅ I. Environment-Driven Configuration
**Status**: PASS (unchanged)  
**Validation**: Data model shows Container, Project, and LogStream entities read from docker-compose.yml and .20i-local. No new configuration layer introduced. All settings remain environment-driven.

### ✅ II. Multi-Platform First
**Status**: PASS (unchanged)  
**Validation**: Quickstart confirms Go builds native binaries for both architectures. No platform-specific code in contracts or data model.

### ✅ III. Path Independence
**Status**: PASS (unchanged)  
**Validation**: Project entity uses absolute paths resolved at runtime. Docker client contract includes `GetComposeProject()` method that discovers project name from compose file location.

### ✅ IV. Centralized Defaults with Override Hierarchy
**Status**: PASS (unchanged)  
**Validation**: TUI remains a consumer of existing config hierarchy. No new defaults or overrides introduced.

### ✅ V. User Experience & Feedback
**Status**: PASS (validated)  
**Validation**: UI Events contract defines clear feedback messages (`containerActionResultMsg` with success/error states). Error handling contract specifies user-friendly messages ("port 80 already in use" not "bind error"). Quickstart confirms visual feedback patterns.

### ✅ VI. Documentation as First-Class Artifact
**Status**: PASS (in progress)  
**Validation**: Comprehensive documentation generated: plan.md, research.md, data-model.md, quickstart.md, contracts/. Inline code comments mandated in quickstart. README and CHANGELOG updates planned.

### ✅ VII. Version Consistency
**Status**: PASS (unchanged)  
**Validation**: TUI introduces no new version variables. All Docker/PHP/MySQL versions read from existing environment.

### 🟢 All Gates Still Pass - Ready for Phase 2 (Implementation)

## Project Structure

### Documentation (this feature)

```text
specs/001-stack-manager-tui/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification (input)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── docker-api.md    # Docker SDK integration contract
│   └── ui-events.md     # Bubble Tea message contracts
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
tui/
├── main.go                    # Entry point, creates RootModel
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── README.md                  # Build and usage instructions
├── Makefile                   # Build targets (build, install, clean)
└── internal/
    ├── app/
    │   ├── root.go            # RootModel (top-level app state)
    │   └── messages.go        # Custom tea.Msg types
    ├── project/               # Phase 3a - Project detection logic (singular)
    │   ├── detector.go        # Project detection ($PWD, public_html check)
    │   ├── template.go        # Template installation from demo-site-folder
    │   └── sanitize.go        # Project name sanitization (legacy GUI script compatible)
    ├── views/
    │   ├── dashboard/
    │   │   ├── dashboard.go   # DashboardModel (three-panel layout)
    │   │   ├── left_panel.go  # Project info panel
    │   │   ├── right_panel.go # Dynamic: pre-flight/output/status table
    │   │   ├── bottom_panel.go # Commands and status messages
    │   │   └── status_table.go # Stack status table with URLs
    │   ├── help/
    │   │   └── help.go        # Help modal
    │   └── projects/          # Phase 4+ - Multi-project browser (deferred)
    │       └── projects.go    # ProjectListModel (to be implemented)
    ├── stack/
    │   ├── compose.go         # Docker Compose operations (up/down/restart)
    │   ├── env.go             # Environment variable handling (CODE_DIR, STACK_FILE, etc)
    │   ├── platform.go        # Platform detection (ARM64 vs x86) - Phase 3a
    │   └── status.go          # Stack status detection
    ├── docker/
    │   ├── client.go          # Docker SDK wrapper
    │   └── stats.go           # Container stats (CPU%)
    └── ui/
        ├── styles.go          # Lipgloss styles (colors, borders)
        ├── components.go      # Reusable components (StatusIcon, ProgressBar)
        └── layout.go          # Panel sizing functions

tests/
└── integration/
    └── tui_test.go            # Integration tests (mock Docker client)
```

**Structure Decision**: Single project structure with project-aware modules. The `internal/project/` package handles the core legacy GUI script workflow (detect project, validate, sanitize name). The `internal/stack/` package wraps Docker Compose operations with proper environment variable handling.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
