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
	for _, want := range []string{"StageServe Status", "Needs attention", "This project is not added to StageServe yet.", "Next: stage up"} {
		if !strings.Contains(out, want) {
			t.Fatalf("fallback status output missing %q:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"attached)", "drift:", "no containers running"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("fallback status output leaked implementation-first copy %q:\n%s", unwanted, out)
		}
	}
}

func TestStatus_CurrentProjectFallbackMatchesGuidedRecoveryLanguage(t *testing.T) {
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
	for _, want := range []string{"Needs attention", "This project is not added to StageServe yet.", "Status:      not added to StageServe", "Next: stage up"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status output missing guided recovery language %q:\n%s", want, out)
		}
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
