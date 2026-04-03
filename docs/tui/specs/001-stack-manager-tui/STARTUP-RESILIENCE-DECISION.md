# Startup Resilience Strategy Decision

**Date**: 2025-12-28  
**Feature**: 001-stack-manager-tui  
**Decision ID**: SR-001  
**Status**: Recommended

## Context

The TUI application needs to decide how to handle Docker daemon unavailability at startup. There are two primary approaches:

### Option 1: Hard-Fail (Current Implementation)

**Behavior**: Exit immediately with an error if Docker is unavailable at startup.

```go
rootModel, err := app.NewRootModel(ctx)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error initializing TUI: %v\n", err)
    os.Exit(1)
}
```

**Pros**:
- ✅ Simple implementation - fail-fast principle
- ✅ Clear error messaging to user
- ✅ No complex state management for degraded mode
- ✅ Prevents confusing UI states
- ✅ User immediately knows Docker is required

**Cons**:
- ❌ No graceful degradation
- ❌ Cannot show helpful troubleshooting UI
- ❌ Prevents displaying system status or help information

### Option 2: Soft-Fail (Resilient Startup)

**Behavior**: Start the application with degraded UI showing Docker unavailability message.

```go
rootModel, err := app.NewRootModel(ctx)
if err != nil {
    // Create degraded RootModel with nil Docker client
    rootModel = &app.RootModel{
        dockerClient: nil,
        activeView:   "dashboard",
        lastError:    err,
    }
}
```

**Pros**:
- ✅ Graceful degradation - app still runs
- ✅ Can show helpful troubleshooting UI
- ✅ Could display help, keyboard shortcuts, documentation
- ✅ Better user experience for transient Docker issues
- ✅ Could implement "retry connection" feature

**Cons**:
- ❌ More complex implementation
- ❌ Requires null-client pattern throughout codebase
- ❌ Risk of confusing partial functionality
- ❌ Increased test surface area
- ❌ May mask configuration problems

## Recommendation: **Hard-Fail (Option 1)**

**Rationale**:

1. **MVP Simplicity**: For the baseline 20i-gui replacement, the hard-fail approach is simpler and sufficient. Docker must be running for the application to have any meaningful functionality.

2. **Clear Requirements**: The application's entire purpose is Docker stack management. Without Docker, there is no value in showing a degraded UI.

3. **Fail-Fast Principle**: Users get immediate, clear feedback that Docker is required and unavailable, rather than a confusing partial UI.

4. **Implementation Cost**: Implementing soft-fail requires significant additional complexity (null object pattern, degraded UI states, retry logic) that doesn't align with MVP goals.

5. **Current Test Coverage**: Tests already verify the app can handle nil Docker client (T024b), proving the code is resilient. This supports future enhancement if needed.

## Implementation Status

**Current**: Hard-fail is implemented and tested ✅
- `main.go` exits with error code 1 and stderr message if Docker unavailable
- Tests verify behavior (T024b)
- Error messages go to stderr (T025a/T025d) ✅

**Future Enhancement** (Post-MVP):
If user feedback indicates a need, soft-fail could be implemented in Phase 8 (Polish) or a future release with:
- Retry connection UI
- Troubleshooting guide display
- Help modal accessible without Docker
- System status indicators

## Decision

**For v1.0 MVP**: Keep hard-fail approach ✅

**Future Consideration**: Soft-fail could be added in v1.1+ if user research shows value in degraded UI for troubleshooting scenarios.

## Test Coverage

- ✅ T024b: Tests verify RootModel can handle nil Docker client
- ✅ T025a/T025d: Tests verify errors go to stderr
- ✅ Current implementation: Hard-fail with clear error messages

## References

- Task: T025b (Decide startup resilience strategy)
- Related Tests: T024b, T025a, T025d
- Implementation: `/tui/main.go`, `/tui/internal/app/root.go`
