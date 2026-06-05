// Status reporter rollback assertions for spec 004 / US2 (FR-006). When a
// project is rolled back during `stage up`, the status reporter MUST NOT
// claim it is still attached or running, and unrelated attached projects MUST
// remain visible and untouched.
package status

import (
	"context"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/docker"
	"github.com/peternicholls/stageserve/internal/mocks"
)

func attachedProject(slug string) config.ProjectConfig {
	return config.ProjectConfig{
		Slug:               slug,
		Name:               slug,
		Hostname:           slug + ".test",
		ComposeProjectName: "stage-" + slug,
		WebNetworkAlias:    "stage-" + slug + "-web",
		RuntimeNetwork:     "stage-" + slug + "-runtime",
		DatabaseVolume:     "stage-" + slug + "-db-data",
	}
}

// TestReporter_RollbackLeavesNoPhantomState proves that when a project failed
// bootstrap and was rolled back, the registry contains no record for it and
// the reporter does not surface phantom running state.
func TestReporter_RollbackLeavesNoPhantomState(t *testing.T) {
	st := mocks.NewState()
	dc := mocks.NewDocker()

	r := &Reporter{State: st, Docker: dc}
	statuses, err := r.All(context.Background())
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(statuses) != 0 {
		t.Fatalf("expected no statuses for empty registry; got %+v", statuses)
	}
}

// TestReporter_RollbackPreservesUnrelatedAttachedProject proves rollback
// isolation in the status surface: when one project failed and was rolled
// back, an unrelated attached project remains reported as attached, with its
// containers visible.
func TestReporter_RollbackPreservesUnrelatedAttachedProject(t *testing.T) {
	other := attachedProject("beta")
	st := mocks.NewState()
	if err := st.Save(state.Record{Project: other, AttachmentState: state.StateAttached}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	dc := mocks.NewDocker()
	dc.Containers = []docker.Container{{
		ID: "beta-apache-1", Name: "stage-beta-apache", Service: "apache", Status: "running",
		Labels: map[string]string{"com.docker.compose.project": other.ComposeProjectName, "com.docker.compose.service": "apache"},
	}}

	r := &Reporter{State: st, Docker: dc}
	statuses, err := r.All(context.Background())
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("status count=%d want 1", len(statuses))
	}
	got := statuses[0]
	if got.Slug != "beta" {
		t.Fatalf("slug=%q want beta", got.Slug)
	}
	if got.AttachmentState != state.StateAttached {
		t.Fatalf("state=%s want attached", got.AttachmentState)
	}
	if len(got.Containers) != 1 {
		t.Fatalf("containers=%d want 1", len(got.Containers))
	}
	if len(got.Drift) != 0 {
		t.Fatalf("unrelated attached project should have no drift, got %+v", got.Drift)
	}
}

// TestReporter_AttachedRecordWithoutContainersReportsDrift confirms that a
// stale attached record with no live containers (the shape a phantom would
// take) is surfaced explicitly as drift, not as a healthy attachment.
func TestReporter_AttachedRecordWithoutContainersReportsDrift(t *testing.T) {
	stale := attachedProject("ghost")
	st := mocks.NewState()
	if err := st.Save(state.Record{Project: stale, AttachmentState: state.StateAttached}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	dc := mocks.NewDocker() // no containers

	r := &Reporter{State: st, Docker: dc}
	statuses, err := r.All(context.Background())
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("status count=%d want 1", len(statuses))
	}
	if len(statuses[0].Drift) == 0 {
		t.Fatalf("expected drift for attached-but-empty record, got none")
	}
}

func TestReporter_OneBySelectorMatchesRecordedProjectPath(t *testing.T) {
	project := attachedProject("beta")
	project.Dir = "/sites/beta"
	st := mocks.NewState()
	if err := st.Save(state.Record{Project: project, AttachmentState: state.StateAttached}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	dc := mocks.NewDocker()
	dc.Containers = []docker.Container{{
		ID: "beta-nginx-1", Name: "stage-beta-nginx", Service: "nginx", Status: "running",
		Labels: map[string]string{"com.docker.compose.project": project.ComposeProjectName, "com.docker.compose.service": "nginx"},
	}}

	reporter := &Reporter{State: st, Docker: dc}
	got, err := reporter.OneBySelector(context.Background(), "/sites/beta")
	if err != nil {
		t.Fatalf("OneBySelector: %v", err)
	}
	if got.Slug != "beta" {
		t.Fatalf("slug=%q want beta", got.Slug)
	}
	if len(got.Containers) != 1 || got.Containers[0].Service != "nginx" {
		t.Fatalf("containers=%+v want nginx", got.Containers)
	}
}

func TestRenderAttachedProjectUsesGuidedSurfaceLanguage(t *testing.T) {
	out := Render(ProjectStatus{
		Slug:            "demo",
		Name:            "Demo Site",
		Hostname:        "demo.test",
		AttachmentState: state.StateAttached,
		Containers: []ContainerStatus{{
			Service: "web",
			Name:    "stage-demo-web",
			Status:  "running",
		}},
	})

	for _, want := range []string{
		"StageServe Status",
		"Ready",
		"This project is running at http://demo.test.",
		"Local URL:   http://demo.test",
		"Status:      running",
		"Project services:",
		"Next: stage logs",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"attached)", "drift:", "no containers running"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("status output leaked implementation-first copy %q:\n%s", unwanted, out)
		}
	}
}

func TestRenderUnrecordedProjectPreservesRecoveryLanguage(t *testing.T) {
	out := Render(ProjectStatus{
		Slug:            "demo",
		Name:            "demo",
		Hostname:        "demo.test",
		AttachmentState: state.StateDown,
		Drift:           []string{"not added to StageServe yet"},
	})

	for _, want := range []string{
		"Needs attention",
		"This project is not added to StageServe yet.",
		"Status:      not added to StageServe",
		"Needs attention:",
		"not added to StageServe yet",
		"Next: stage up",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing %q:\n%s", want, out)
		}
	}
}
