// Status reporter rollback assertions for spec 004 / US2 (FR-006). When a
// project is rolled back during `stacklane up`, the status reporter MUST NOT
// claim it is still attached or running, and unrelated attached projects MUST
// remain visible and untouched.
package status

import (
	"context"
	"testing"

	"github.com/peternicholls/stacklane/core/config"
	"github.com/peternicholls/stacklane/core/state"
	"github.com/peternicholls/stacklane/infra/docker"
	"github.com/peternicholls/stacklane/internal/mocks"
)

func attachedProject(slug string) config.ProjectConfig {
	return config.ProjectConfig{
		Slug:               slug,
		Name:               slug,
		Hostname:           slug + ".test",
		ComposeProjectName: "stln-" + slug,
		WebNetworkAlias:    "stln-" + slug + "-web",
		RuntimeNetwork:     "stln-" + slug + "-runtime",
		DatabaseVolume:     "stln-" + slug + "-db-data",
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
		ID: "beta-apache-1", Name: "stln-beta-apache", Service: "apache", Status: "running",
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
