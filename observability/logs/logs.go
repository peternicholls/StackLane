// Logs streams Docker logs for a project, with named diagnostics when a
// requested service is missing or unhealthy.
package logs

import (
	"context"
	"fmt"
	"io"

	"github.com/peternicholls/stageserve/infra/docker"
)

// Streamer streams logs for one container by name (or compose service).
type Streamer struct {
	Docker docker.DockerClient
}

// Stream finds the container by service+project label and copies its logs to w.
func (s *Streamer) Stream(ctx context.Context, composeProject, service string, follow bool, w io.Writer) error {
	containers, err := s.Docker.ListContainersByLabel(ctx, map[string]string{
		"com.docker.compose.project": composeProject,
		"com.docker.compose.service": service,
	})
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return fmt.Errorf("logs: no container found for service %q in project %q", service, composeProject)
	}
	stream, err := s.Docker.ContainerLogs(ctx, containers[0].ID, follow)
	if err != nil {
		return err
	}
	defer stream.Close()
	_, err = io.Copy(w, stream)
	return err
}
