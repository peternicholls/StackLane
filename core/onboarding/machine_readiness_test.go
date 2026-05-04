package onboarding_test

import (
	"net"
	"os"
	"testing"

	"github.com/peternicholls/stacklane/core/onboarding"
)

// TestDockerBinaryCheck_MissingBinary verifies that when the Docker binary
// cannot be found, the readiness step reports needs_action with a remediation.
func TestDockerBinaryCheck_MissingBinary(t *testing.T) {
	r := onboarding.CheckDockerBinary("/nonexistent/docker/binary/path")
	if r.ID != "docker.binary" {
		t.Errorf("want id docker.binary, got %s", r.ID)
	}
	if r.Status != onboarding.StatusNeedsAction {
		t.Errorf("want needs_action, got %s", r.Status)
	}
	if r.Remediation == nil || *r.Remediation == "" {
		t.Error("want non-empty remediation for missing docker binary")
	}
}

// TestDockerBinaryCheck_PresentBinary verifies that when Docker binary exists
// and is executable, the step reports ready.
func TestDockerBinaryCheck_PresentBinary(t *testing.T) {
	// /usr/bin/true is a reliable always-present executable on macOS/Linux.
	r := onboarding.CheckDockerBinary("/usr/bin/true")
	if r.Status != onboarding.StatusReady {
		t.Errorf("want ready, got %s — %s", r.Status, r.Message)
	}
}

// TestStateDirCheck_MissingDir verifies a missing state dir is reported as
// needs_action with a creation remediation.
func TestStateDirCheck_MissingDir(t *testing.T) {
	r := onboarding.CheckStateDir("/tmp/stacklane_test_nonexistent_xyz123")
	if r.ID != "state.dir" {
		t.Errorf("want id state.dir, got %s", r.ID)
	}
	if r.Status != onboarding.StatusNeedsAction {
		t.Errorf("want needs_action, got %s", r.Status)
	}
	if r.Remediation == nil || *r.Remediation == "" {
		t.Error("want non-empty remediation for missing state dir")
	}
}

// TestStateDirCheck_ExistingDir verifies that an existing directory reports ready.
func TestStateDirCheck_ExistingDir(t *testing.T) {
	dir := t.TempDir()
	r := onboarding.CheckStateDir(dir)
	if r.Status != onboarding.StatusReady {
		t.Errorf("want ready for existing dir, got %s — %s", r.Status, r.Message)
	}
}

// TestStateDirCheck_FileNotDir verifies that a path pointing to a file (not a
// directory) is reported as error.
func TestStateDirCheck_FileNotDir(t *testing.T) {
	f, err := os.CreateTemp("", "stacklane_test_notdir_*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	r := onboarding.CheckStateDir(f.Name())
	if r.Status != onboarding.StatusError {
		t.Errorf("want error for file-not-dir, got %s — %s", r.Status, r.Message)
	}
}

// TestPortCheck_FreePort verifies that a port known to be free is reported ready.
func TestPortCheck_FreePort(t *testing.T) {
	// Bind an ephemeral port, note the port, release it, then check it.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot listen on ephemeral port; skipping test")
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	r := onboarding.CheckPort("port.test", port)
	if r.Status != onboarding.StatusReady {
		t.Errorf("want ready for free port %d, got %s — %s", port, r.Status, r.Message)
	}
}

// TestPortCheck_BusyPort verifies that a port that is actively listening is
// reported as needs_action.
func TestPortCheck_BusyPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot listen on ephemeral port; skipping test")
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	r := onboarding.CheckPort("port.test", port)
	if r.Status != onboarding.StatusNeedsAction {
		t.Errorf("want needs_action for busy port %d, got %s — %s", port, r.Status, r.Message)
	}
}
