# Contract: Docker API Integration

**Feature**: 001-stack-manager-tui  
**Date**: 2025-12-28  
**Type**: External System Integration

## Overview

This contract defines the interface between the TUI application and the Docker Engine API via the Docker SDK for Go. It specifies required operations, error handling patterns, timeout policies, and data transformation rules.

---

## API Client Contract

### Initialization

**Method**: `NewClient(ctx context.Context) (*Client, error)`

**Inputs**:
- `ctx`: Context with cancellation for graceful shutdown

**Outputs**:
- `*Client`: Wrapper around Docker SDK client
- `error`: Connection error if Docker daemon unreachable

**Behavior**:
- Attempts to connect to Docker daemon via default socket (`unix:///var/run/docker.sock` on Unix, named pipe on Windows)
- Negotiates API version with daemon
- Returns error if daemon not running or insufficient permissions

**Error Conditions**:
| Error | Meaning | User Message |
|-------|---------|--------------|
| Connection refused | Docker daemon not running | "Docker daemon not running. Start Docker Desktop and press 'r' to retry." |
| Permission denied | User not in docker group | "Cannot connect to Docker. Add your user to the docker group or run with sudo." |
| Timeout | Daemon slow/unresponsive | "Docker daemon not responding. Check Docker Desktop status." |

---

## Container Operations

### List Containers

**Method**: `ListContainers(projectName string) ([]Container, error)`

**Inputs**:
- `projectName`: Docker Compose project name (filter label)

**Outputs**:
- `[]Container`: List of containers matching project
- `error`: API error if operation fails

**Behavior**:
- Filters by label: `com.docker.compose.project=<projectName>`
- Returns all containers (running and stopped)
- Maps Docker SDK types to internal Container type

**Timeout**: 5 seconds

**Example**:
```go
containers, err := client.ListContainers("myproject")
// Returns: [{ID: "a3f2d1b8", Name: "myproject-apache-1", Status: Running}, ...]
```

---

### Start Container

**Method**: `StartContainer(containerID string) error`

**Inputs**:
- `containerID`: Docker container ID (12-char hash)

**Outputs**:
- `error`: nil on success, error on failure

**Behavior**:
- Starts stopped container
- No-op if container already running (not an error)
- Returns immediately, start is async

**Timeout**: 5 seconds for API call (container startup may take longer)

**Error Conditions**:
| Error | User Message |
|-------|--------------|
| Container not found | "Container not found. It may have been removed." |
| Port conflict | "Failed to start: port {port} already in use" |
| Image missing | "Failed to start: image {image} not found. Run 'docker compose pull'." |

---

### Stop Container

**Method**: `StopContainer(containerID string, timeout int) error`

**Inputs**:
- `containerID`: Docker container ID
- `timeout`: Seconds to wait before force-kill (default: 10)

**Outputs**:
- `error`: nil on success

**Behavior**:
- Sends SIGTERM to container
- Waits `timeout` seconds for graceful shutdown
- Sends SIGKILL if timeout exceeded
- No-op if already stopped

**Timeout**: `timeout` + 5 seconds for API overhead

---

### Restart Container

**Method**: `RestartContainer(containerID string, timeout int) error`

**Inputs**:
- `containerID`: Docker container ID
- `timeout`: Seconds to wait during stop phase

**Outputs**:
- `error`: nil on success

**Behavior**:
- Equivalent to Stop + Start
- Uses same timeout semantics as StopContainer

**Timeout**: (timeout + 5) + 5 seconds (stop + start)

---

### Remove Container

**Method**: `RemoveContainer(containerID string, force bool) error`

**Inputs**:
- `containerID`: Docker container ID
- `force`: If true, stop running container before removing

**Outputs**:
- `error`: nil on success

**Behavior**:
- Removes container and its anonymous volumes
- If `force=true`, stops container first (SIGKILL after 2s)
- If `force=false`, returns error if container running

**Timeout**: 10 seconds

**Error Conditions**:
- Container running and force=false: "Cannot remove running container. Stop it first or use force."

---

## Stats Streaming

### Watch Stats

**Method**: `WatchStats(containerID string) (<-chan Stats, error)`

**Inputs**:
- `containerID`: Docker container ID

**Outputs**:
- `<-chan Stats`: Channel receiving stats updates
- `error`: Error if container doesn't exist or not running

**Behavior**:
- Opens streaming connection to Docker stats API
- Sends `Stats` object to channel every ~1 second
- Runs in background goroutine
- Closes channel when container stops or context canceled
- **Non-blocking**: Returns channel immediately, stats arrive async

**Stats Object**:
```go
type Stats struct {
    ContainerID     string
    CPUPercent      float64  // 0-400 (4 cores)
    MemoryUsed      uint64   // bytes
    MemoryLimit     uint64   // bytes (0 = unlimited)
    MemoryPercent   float64  // 0-100
    NetworkRxBytes  uint64   // cumulative
    NetworkTxBytes  uint64   // cumulative
    Timestamp       time.Time
}
```

**Example**:
```go
statsChan, err := client.WatchStats("a3f2d1b8")
for stats := range statsChan {
    // Update UI with stats.CPUPercent, stats.MemoryUsed
}
```

**Cleanup**: Close context to stop stats stream and free resources

---

## Log Streaming

### Stream Logs

**Method**: `StreamLogs(containerID string, since time.Time, follow bool) (<-chan string, error)`

**Inputs**:
- `containerID`: Docker container ID
- `since`: Only return logs after this timestamp (use `time.Time{}` for all logs)
- `follow`: If true, stream new logs as they arrive (tail -f mode)

**Outputs**:
- `<-chan string`: Channel receiving log lines (one line per message)
- `error`: Error if container doesn't exist

**Behavior**:
- Opens streaming connection to Docker logs API
- Sends one log line per channel message
- Interleaves stdout and stderr (no distinction in MVP)
- Runs in background goroutine
- Closes channel when container stops (if follow=true) or all logs sent (if follow=false)

**Timeout**: None (long-lived stream)

**Example**:
```go
// Load last 100 lines
logsChan, _ := client.StreamLogs("a3f2d1b8", time.Now().Add(-5*time.Minute), false)

// Follow mode
logsChan, _ := client.StreamLogs("a3f2d1b8", time.Now(), true)
for line := range logsChan {
    // Append to log buffer
}
```

---

## Project Detection

### Get Compose Project

**Method**: `GetComposeProject(composeFilePath string) (string, error)`

**Inputs**:
- `composeFilePath`: Absolute path to docker-compose.yml

**Outputs**:
- `string`: Project name (from COMPOSE_PROJECT_NAME or directory name)
- `error`: Error if file doesn't exist or is invalid

**Behavior**:
- Parses docker-compose.yml for `name:` field (Compose v2)
- Falls back to directory name, sanitized (lowercase, alphanumeric+hyphens)
- Matches Docker Compose's project name algorithm

**Example**:
```go
projectName, _ := client.GetComposeProject("/Users/me/myproject/docker-compose.yml")
// Returns: "myproject"
```

---

## Docker Compose Operations

### Down (Destroy)

**Method**: `ComposeDown(projectPath string, removeVolumes bool) error`

**Inputs**:
- `projectPath`: Directory containing docker-compose.yml
- `removeVolumes`: If true, remove named and anonymous volumes

**Outputs**:
- `error`: nil on success

**Behavior**:
- Equivalent to `docker compose down` (or `docker compose down -v` if removeVolumes=true)
- Stops all containers in project
- Removes containers and networks
- Optionally removes volumes (⚠️ data loss!)

**Timeout**: 60 seconds (allows graceful shutdown)

**Example**:
```go
// Destroy stack (for "D" key action)
err := client.ComposeDown("/Users/me/myproject", true)
```

---

### Restart All

**Method**: `ComposeRestart(projectPath string) error`

**Inputs**:
- `projectPath`: Directory containing docker-compose.yml

**Outputs**:
- `error`: nil on success

**Behavior**:
- Equivalent to `docker compose restart`
- Restarts all containers in project (stop + start each)

**Timeout**: 60 seconds

---

### Stop All

**Method**: `ComposeStop(projectPath string) error`

**Inputs**:
- `projectPath`: Directory containing docker-compose.yml

**Outputs**:
- `error`: nil on success

**Behavior**:
- Equivalent to `docker compose stop`
- Stops all running containers in project (SIGTERM, 10s timeout, then SIGKILL)

**Timeout**: 30 seconds

---

## Error Handling Contract

### Error Types

All methods return one of these error categories:

| Error Type | Condition | Retry Strategy |
|------------|-----------|----------------|
| `ErrDaemonUnreachable` | Docker daemon not running | Auto-retry every 5s, show error screen |
| `ErrPermissionDenied` | User lacks Docker permissions | Show error, require manual fix |
| `ErrTimeout` | Operation exceeded timeout | Allow manual retry (button) |
| `ErrNotFound` | Container/image doesn't exist | Show error, refresh list |
| `ErrConflict` | Port conflict, name conflict | Show actionable error message |
| `ErrUnknown` | Unexpected error | Log details, show generic error |

### Error Message Formatting

**Template**: `{Action} failed: {Reason}`

**Examples**:
- "Start container 'apache' failed: port 80 already in use"
- "Stop container 'mariadb' failed: timeout after 10s"
- "List containers failed: Docker daemon not running"

**User-Friendly Mapping**:
```go
func formatError(err error, action string, containerName string) string {
    if errors.Is(err, ErrDaemonUnreachable) {
        return "Docker daemon not running. Start Docker Desktop."
    }
    if strings.Contains(err.Error(), "port is already allocated") {
        port := extractPort(err.Error())
        return fmt.Sprintf("Port %s already in use. Stop conflicting service.", port)
    }
    return fmt.Sprintf("%s '%s' failed: %s", action, containerName, err)
}
```

---

## Performance Requirements

| Operation | Max Latency | Notes |
|-----------|-------------|-------|
| ListContainers | 500ms | Cached on client side, refresh every 2s |
| Start/Stop/Restart | 5s API call | Actual operation may take longer, but API returns quickly |
| WatchStats (first frame) | 1s | Subsequent frames arrive every ~1s |
| StreamLogs (initial load) | 2s for 100 lines | Follow mode is real-time (<100ms) |
| ComposeDown | 60s | Graceful shutdown can be slow |

---

## Security Considerations

1. **Socket Permissions**: TUI must have read/write access to Docker socket
2. **No Credential Storage**: TUI doesn't store Docker credentials (relies on Docker's auth)
3. **Command Injection**: Never shell out to `docker` CLI - use SDK only
4. **Volume Mounts**: When removing volumes, confirm with user (data loss risk)

---

## Testing Contract

### Mock Client for Unit Tests

```go
type MockDockerClient struct {
    Containers []Container
    StatsMap   map[string]chan Stats
    LogsMap    map[string]chan string
    Errors     map[string]error  // Method name -> error to return
}

func (m *MockDockerClient) ListContainers(project string) ([]Container, error) {
    if err, ok := m.Errors["ListContainers"]; ok {
        return nil, err
    }
    return m.Containers, nil
}
```

### Integration Tests

- Start real Docker container (nginx:alpine)
- Test start/stop/restart operations
- Verify stats streaming works
- Clean up container after test

---

## Version Compatibility

- **Minimum Docker API**: 1.41 (Docker 20.10+)
- **Recommended**: 1.43+ (Docker 24.0+)
- **Docker Compose**: v2.0+ (integrated `docker compose`, not standalone `docker-compose`)

**Version Detection**:
```go
info, err := client.ServerVersion(ctx)
if info.APIVersion < "1.41" {
    return fmt.Errorf("Docker API 1.41+ required, got %s", info.APIVersion)
}
```
