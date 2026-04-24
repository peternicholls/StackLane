// Package docker defines the typed Docker Engine SDK contract.
//
// The SDK surface is intentionally narrow: only network and container query
// operations live here. Compose-file orchestration (`docker compose up/down`)
// is a separate concern and lives in infra/compose; this keeps the SDK
// transport isolated from CLI subprocess invocation.
package docker

import (
	"context"
	"time"
)

// Container is the typed projection of a docker container we care about.
type Container struct {
	ID      string
	Name    string
	Status  string
	Service string // io.stacklane.service / com.docker.compose.service label
	Project string // com.docker.compose.project label
	Labels  map[string]string
}

// HealthCheck reports the running health state of a single service.
type HealthCheck struct {
	Service string
	Status  string // "healthy", "unhealthy", "starting", "none"
	Reason  string // free-form, populated on failure / timeout
}

// DockerClient wraps Docker Engine SDK operations. Implementations must
// satisfy this contract; compose orchestration is NOT here on purpose.
type DockerClient interface {
	NetworkExists(ctx context.Context, name string) (bool, error)
	CreateNetwork(ctx context.Context, name string) error
	RemoveNetwork(ctx context.Context, name string) error
	ListContainersByLabel(ctx context.Context, labels map[string]string) ([]Container, error)
	// WaitHealthy blocks until every container belonging to the named compose
	// project either reports healthy or exhausts the timeout. On timeout, the
	// returned error names the still-unhealthy services (FR-009 / SC-009).
	WaitHealthy(ctx context.Context, composeProject string, timeout time.Duration) error
	// ContainerLogs streams logs for a single container. Caller closes the
	// returned reader; implementations MAY return a never-ending stream when
	// follow is true.
	ContainerLogs(ctx context.Context, containerID string, follow bool) (LogStream, error)
	// Available returns nil if the daemon is reachable, an error otherwise.
	Available(ctx context.Context) error
}

// LogStream is the small Reader-plus-Close surface used for streaming logs.
type LogStream interface {
	Read(p []byte) (int, error)
	Close() error
}
