// Status reports the runtime view of a project as a typed value plus a
// human-rendered string. Combines what the Bash code did inline in
// twentyi_status / twentyi_status_summary.
package status

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/docker"
)

// ProjectStatus is the typed status of a single project.
type ProjectStatus struct {
	Slug            string
	Name            string
	Hostname        string
	AttachmentState state.AttachmentState
	Containers      []ContainerStatus
	Drift           []string // human-readable drift descriptions (FR-010)
}

// ContainerStatus is the typed status of one container.
type ContainerStatus struct {
	Service string
	Name    string
	Status  string
}

// Reporter materialises ProjectStatus values from registry + live docker.
type Reporter struct {
	State  state.StateStore
	Docker docker.DockerClient
}

// All reports status for every recorded project.
func (r *Reporter) All(ctx context.Context) ([]ProjectStatus, error) {
	rows, err := r.State.Registry()
	if err != nil {
		return nil, err
	}
	out := make([]ProjectStatus, 0, len(rows))
	for _, row := range rows {
		ps, err := r.byRow(ctx, row)
		if err != nil {
			ps.Drift = append(ps.Drift, "registry: "+err.Error())
		}
		out = append(out, ps)
	}
	return out, nil
}

// One reports the status of a single project by slug.
func (r *Reporter) One(ctx context.Context, slug string) (ProjectStatus, error) {
	rows, err := r.State.Registry()
	if err != nil {
		return ProjectStatus{}, err
	}
	for _, row := range rows {
		if row.Slug == slug {
			return r.byRow(ctx, row)
		}
	}
	return ProjectStatus{}, fmt.Errorf("status: project %q not found", slug)
}

func (r *Reporter) byRow(ctx context.Context, row state.RegistryRow) (ProjectStatus, error) {
	ps := ProjectStatus{
		Slug:            row.Slug,
		Name:            row.Name,
		Hostname:        row.Hostname,
		AttachmentState: row.AttachmentState,
	}
	containers, err := r.Docker.ListContainersByLabel(ctx, map[string]string{"com.docker.compose.project": row.ComposeProject})
	if err != nil {
		return ps, err
	}
	for _, c := range containers {
		ps.Containers = append(ps.Containers, ContainerStatus{Service: c.Service, Name: c.Name, Status: c.Status})
	}
	sort.SliceStable(ps.Containers, func(i, j int) bool { return ps.Containers[i].Service < ps.Containers[j].Service })
	if row.AttachmentState == state.StateAttached && len(ps.Containers) == 0 {
		ps.Drift = append(ps.Drift, "marked attached but no containers found")
	}
	return ps, nil
}

// Render returns a human-readable status block matching the spirit of the
// bash output. Semantic equivalence; not byte-for-byte (FR-014).
func Render(s ProjectStatus) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (%s) — %s\n", s.Slug, s.AttachmentState, s.Hostname)
	if len(s.Containers) == 0 {
		b.WriteString("  no containers running\n")
	}
	for _, c := range s.Containers {
		fmt.Fprintf(&b, "  %s  %-30s %s\n", c.Service, c.Name, c.Status)
	}
	for _, d := range s.Drift {
		fmt.Fprintf(&b, "  ! drift: %s\n", d)
	}
	return b.String()
}
