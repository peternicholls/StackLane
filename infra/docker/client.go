// Docker Engine SDK client. Wraps github.com/docker/docker/client with the
// narrow surface defined in types.go: networks, container queries, healthcheck
// waits, and a logs reader. No `docker compose` invocation here — that lives
// in infra/compose.
package docker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	dfilters "github.com/docker/docker/api/types/filters"
	dnetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// SDKClient is the default DockerClient backed by the Docker Engine SDK.
type SDKClient struct {
	once sync.Once
	cli  *client.Client
	err  error
}

// NewSDKClient returns a lazily-initialised SDK client.
func NewSDKClient() *SDKClient { return &SDKClient{} }

func (s *SDKClient) ensure() (*client.Client, error) {
	s.once.Do(func() {
		c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		s.cli = c
		s.err = err
	})
	return s.cli, s.err
}

// Available checks daemon connectivity via /_ping equivalent (Info).
func (s *SDKClient) Available(ctx context.Context) error {
	c, err := s.ensure()
	if err != nil {
		return err
	}
	_, err = c.Ping(ctx)
	return err
}

func (s *SDKClient) NetworkExists(ctx context.Context, name string) (bool, error) {
	c, err := s.ensure()
	if err != nil {
		return false, err
	}
	if _, err := c.NetworkInspect(ctx, name, dnetwork.InspectOptions{}); err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *SDKClient) CreateNetwork(ctx context.Context, name string) error {
	c, err := s.ensure()
	if err != nil {
		return err
	}
	_, err = c.NetworkCreate(ctx, name, dnetwork.CreateOptions{Driver: "bridge"})
	return err
}

func (s *SDKClient) RemoveNetwork(ctx context.Context, name string) error {
	c, err := s.ensure()
	if err != nil {
		return err
	}
	if err := c.NetworkRemove(ctx, name); err != nil {
		if client.IsErrNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func (s *SDKClient) ListContainersByLabel(ctx context.Context, labels map[string]string) ([]Container, error) {
	c, err := s.ensure()
	if err != nil {
		return nil, err
	}
	args := dfilters.NewArgs()
	for k, v := range labels {
		if v == "" {
			args.Add("label", k)
		} else {
			args.Add("label", fmt.Sprintf("%s=%s", k, v))
		}
	}
	list, err := c.ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil {
		return nil, err
	}
	out := make([]Container, 0, len(list))
	for _, ci := range list {
		name := ""
		if len(ci.Names) > 0 {
			name = strings.TrimPrefix(ci.Names[0], "/")
		}
		out = append(out, Container{
			ID:      ci.ID,
			Name:    name,
			Status:  ci.Status,
			Service: ci.Labels["com.docker.compose.service"],
			Project: ci.Labels["com.docker.compose.project"],
			Labels:  ci.Labels,
		})
	}
	return out, nil
}

// ContainerLogs streams logs for one container.
func (s *SDKClient) ContainerLogs(ctx context.Context, containerID string, follow bool) (LogStream, error) {
	c, err := s.ensure()
	if err != nil {
		return nil, err
	}
	rc, err := c.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       "all",
	})
	if err != nil {
		return nil, err
	}
	return rc, nil
}

// ErrTimeout is the sentinel WaitHealthy returns when timeout elapses.
var ErrTimeout = errors.New("wait-healthy: timeout")

// WaitHealthy polls every container belonging to composeProject until each
// one is healthy or starting→running with no healthcheck. Returns a typed
// error naming still-unhealthy services on timeout.
func (s *SDKClient) WaitHealthy(ctx context.Context, composeProject string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	deadline := time.Now().Add(timeout)
	pollEvery := 500 * time.Millisecond

	for {
		checks, err := s.healthSnapshot(ctx, composeProject)
		if err != nil {
			return err
		}
		if len(checks) == 0 {
			// No containers yet; keep polling until we see them or timeout.
		}
		stillBad := []string{}
		for _, ch := range checks {
			switch ch.Status {
			case "healthy", "running-no-check":
				continue
			default:
				stillBad = append(stillBad, ch.Service+"="+ch.Status)
			}
		}
		if len(checks) > 0 && len(stillBad) == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w (services not healthy: %s)", ErrTimeout, strings.Join(stillBad, ", "))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollEvery):
		}
	}
}

func (s *SDKClient) healthSnapshot(ctx context.Context, composeProject string) ([]HealthCheck, error) {
	c, err := s.ensure()
	if err != nil {
		return nil, err
	}
	args := dfilters.NewArgs()
	args.Add("label", "com.docker.compose.project="+composeProject)
	list, err := c.ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil {
		return nil, err
	}
	out := make([]HealthCheck, 0, len(list))
	for _, ci := range list {
		insp, err := c.ContainerInspect(ctx, ci.ID)
		if err != nil {
			out = append(out, HealthCheck{Service: ci.Labels["com.docker.compose.service"], Status: "inspect-error", Reason: err.Error()})
			continue
		}
		svc := ci.Labels["com.docker.compose.service"]
		switch {
		case insp.State == nil:
			out = append(out, HealthCheck{Service: svc, Status: "unknown"})
		case insp.State.Health != nil:
			out = append(out, HealthCheck{Service: svc, Status: insp.State.Health.Status})
		case insp.State.Running:
			out = append(out, HealthCheck{Service: svc, Status: "running-no-check"})
		default:
			out = append(out, HealthCheck{Service: svc, Status: insp.State.Status})
		}
	}
	return out, nil
}
