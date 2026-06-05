package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestStatus_CurrentProjectWithoutRecordRendersFallback(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()

	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--stack-home", stackHome, "--project-dir", projectDir, "status"})

	if err := root.Execute(); err != nil {
		t.Fatalf("status should not fail when current project is unrecorded: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "not added to StageServe yet") {
		t.Fatalf("fallback status output missing unrecorded hint:\n%s", out)
	}
}

func TestStatus_ExplicitSelectorStillErrorsWhenProjectMissing(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()

	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--stack-home", stackHome, "--project-dir", projectDir, "status", "--project", "missing"})

	if err := root.Execute(); err == nil {
		t.Fatal("expected status --project missing to fail")
	}
}
