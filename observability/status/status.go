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
	return ProjectStatus{}, fmt.Errorf("status: project %q not found: %w", slug, state.ErrNotFound)
}

// OneBySelector reports a single project matched by slug, project name,
// hostname, or recorded project path.
func (r *Reporter) OneBySelector(ctx context.Context, selector string) (ProjectStatus, error) {
	rec, _, err := r.State.StateFileForSelector(selector)
	if err != nil {
		return ProjectStatus{}, fmt.Errorf("status: project %q not found: %w", selector, err)
	}
	return r.One(ctx, rec.Project.Slug)
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
	fmt.Fprintf(&b, "StageServe Status\n%s\n\n%s\n", severityLabel(s), statusVerdict(s))
	fmt.Fprintf(&b, "\nKey facts:\n")
	fmt.Fprintf(&b, "  Project:     %s\n", displayProject(s))
	if s.Hostname != "" {
		fmt.Fprintf(&b, "  Local URL:   http://%s\n", s.Hostname)
	}
	fmt.Fprintf(&b, "  Status:      %s\n", statusLabel(s))

	fmt.Fprintf(&b, "\nProject services:\n")
	if len(s.Containers) == 0 {
		b.WriteString("  No project services are running.\n")
	}
	for _, c := range s.Containers {
		fmt.Fprintf(&b, "  %s  %-30s %s\n", c.Service, c.Name, c.Status)
	}
	if len(s.Drift) > 0 {
		fmt.Fprintf(&b, "\nNeeds attention:\n")
		for _, d := range s.Drift {
			fmt.Fprintf(&b, "  %s\n", d)
		}
	}
	fmt.Fprintf(&b, "\nNext: %s\n", nextAction(s))
	return b.String()
}

func severityLabel(s ProjectStatus) string {
	if len(s.Drift) > 0 {
		return "Needs attention"
	}
	return "Ready"
}

func statusVerdict(s ProjectStatus) string {
	switch {
	case hasDrift(s, "not added to StageServe yet"):
		return "This project is not added to StageServe yet."
	case len(s.Drift) > 0:
		return "This project doesn't match what StageServe expects."
	case s.AttachmentState == state.StateAttached:
		if s.Hostname != "" {
			return "This project is running at http://" + s.Hostname + "."
		}
		return "This project is running."
	default:
		return "This project is stopped."
	}
}

func statusLabel(s ProjectStatus) string {
	switch {
	case hasDrift(s, "not added to StageServe yet"):
		return "not added to StageServe"
	case len(s.Drift) > 0:
		return "needs attention"
	case s.AttachmentState == state.StateAttached:
		return "running"
	default:
		return "stopped"
	}
}

func nextAction(s ProjectStatus) string {
	switch {
	case hasDrift(s, "not added to StageServe yet"):
		return "stage up"
	case len(s.Drift) > 0:
		return "stage doctor"
	case s.AttachmentState == state.StateAttached:
		return "stage logs"
	default:
		return "stage up"
	}
}

func displayProject(s ProjectStatus) string {
	if s.Name != "" && s.Slug != "" && s.Name != s.Slug {
		return fmt.Sprintf("%s (%s)", s.Name, s.Slug)
	}
	if s.Name != "" {
		return s.Name
	}
	return s.Slug
}

func hasDrift(s ProjectStatus, value string) bool {
	for _, drift := range s.Drift {
		if drift == value {
			return true
		}
	}
	return false
}
